package reporting

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"slices"
	"strings"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/artifactport"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type artifactFiles map[string][]byte

func (f artifactFiles) File(name string) ([]byte, bool) { data, ok := f[name]; return data, ok }
func (f artifactFiles) Paths() []string {
	names := make([]string, 0, len(f))
	for name := range f {
		names = append(names, name)
	}
	slices.Sort(names)
	return names
}

func (source PortableReferenceSource) ExportArtifactsTx(ctx context.Context, tx pgx.Tx, incident uuid.UUID, references []byte, write artifactport.WriteFile) error {
	if tx == nil || incident == uuid.Nil || write == nil {
		return artifactport.ErrInvalid
	}
	refs, err := reference_data.DecodeIncidentBundleReferences(references)
	if err != nil {
		return artifactport.ErrInvalid
	}
	catalog := emptyPortableCatalog(incident)
	retained, err := loadImportedArtifactFiles(ctx, tx, incident)
	if err != nil {
		return err
	}
	if len(retained) > 0 {
		// Validation does not adopt source attestations as destination trust.
		if _, err := source.PrepareArtifactImport(ctx, artifactport.ImportRequest{IncidentID: incident, OperationID: incident, References: references, Bundle: artifactFiles(retained)}); err != nil {
			return err
		}
		catalog, err = decodePortableCatalog(retained[portableArtifactCatalogPath], incident)
		if err != nil {
			return err
		}
		for _, name := range artifactFiles(retained).Paths() {
			if name == portableArtifactCatalogPath {
				continue
			}
			if err := write(ctx, name, int64(len(retained[name])), bytes.NewReader(retained[name])); err != nil {
				return err
			}
		}
	}
	snapshots := map[string]reference_data.SetBinding{}
	for _, s := range catalog.Snapshots {
		model, _, err := decodePortableObject[snapshotCapture](retained[s.CapturePath], s.CaptureSHA256)
		if err != nil {
			return err
		}
		snapshots[s.SnapshotID] = model.ReferencePacks
	}
	rows, err := tx.Query(ctx, `SELECT s.snapshot_id,s.source_capture_sha256,p.request_json->'source_capture' FROM reporting_snapshots s JOIN reporting_job_payloads p ON p.job_id=s.create_job_id WHERE s.incident_id=$1 ORDER BY s.snapshot_id`, incident)
	if err != nil {
		return err
	}
	for rows.Next() {
		var id uuid.UUID
		var digest string
		var raw []byte
		if err := rows.Scan(&id, &digest, &raw); err != nil {
			rows.Close()
			return err
		}
		model, data, err := decodePortableObject[snapshotCapture](raw, digest)
		if err != nil || validateSnapshotCapture(model) != nil || model.DerivationVersion != DerivationVersion || model.SnapshotID != id.String() || model.IncidentID != incident.String() {
			rows.Close()
			return artifactport.ErrInvalid
		}
		if _, duplicate := snapshots[id.String()]; duplicate || len(catalog.Snapshots) >= portableSnapshotLimit {
			rows.Close()
			return artifactport.ErrInvalid
		}
		if err := reference_data.ValidatePortableBinding(refs, model.ReferencePacks); err != nil {
			rows.Close()
			return artifactport.ErrInvalid
		}
		name := portableCapturePath(id.String())
		if err := write(ctx, name, int64(len(data)), bytes.NewReader(data)); err != nil {
			rows.Close()
			return err
		}
		snapshots[id.String()] = model.ReferencePacks
		catalog.Snapshots = append(catalog.Snapshots, portableSnapshot{SnapshotID: id.String(), CapturePath: name, CaptureSHA256: digest})
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	// Close row streams before reading each release's member inventory.
	type nativeRelease struct {
		descriptor portableRelease
		manifest   RenderBundleManifest
		data       []byte
	}
	releases := []nativeRelease{}
	seen := map[string]bool{}
	for _, r := range catalog.Releases {
		seen[r.ReleaseID] = true
	}
	rows, err = tx.Query(ctx, `SELECT r.release_id,r.snapshot_id,b.bundle_manifest_sha256,b.bundle_manifest_json FROM reporting_releases r JOIN reporting_render_bundles b USING(release_id) WHERE r.incident_id=$1 ORDER BY r.release_id`, incident)
	if err != nil {
		return err
	}
	for rows.Next() {
		var id, snapshot uuid.UUID
		var digest string
		var raw []byte
		if err := rows.Scan(&id, &snapshot, &digest, &raw); err != nil {
			rows.Close()
			return err
		}
		manifest, data, err := decodePortableObject[RenderBundleManifest](raw, digest)
		if err != nil || validateBundleIdentity(manifest) != nil || manifest.ReleaseID == nil || *manifest.ReleaseID != id.String() || manifest.SnapshotID != snapshot.String() || manifest.SchemaID != RenderBundleManifestSchemaID || manifest.ManifestVersion != RenderBundleManifestVersion || len(manifest.Files) == 0 || len(manifest.Files) > 10000 || seen[id.String()] || len(catalog.Releases)+len(releases) >= portableReleaseLimit {
			rows.Close()
			return artifactport.ErrInvalid
		}
		binding, exists := snapshots[snapshot.String()]
		expected, _ := canonicaljson.Marshal(binding)
		actual, _ := canonicaljson.Marshal(manifest.ReferencePacks)
		if !exists || !bytes.Equal(expected, actual) {
			rows.Close()
			return artifactport.ErrInvalid
		}
		descriptor := portableRelease{ReleaseID: id.String(), SnapshotID: snapshot.String(), ManifestPath: portableManifestPath(id.String()), ManifestSHA256: digest}
		releases = append(releases, nativeRelease{descriptor, manifest, data})
		seen[id.String()] = true
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	for _, release := range releases {
		descriptor, manifest := release.descriptor, release.manifest
		previous := ""
		primary := false
		publicArtifacts := []RenderBundleFile{}
		for _, f := range manifest.Files {
			if !portableMemberPath(f.Path) || f.Path <= previous || f.SizeBytes < 0 || !portableDigestPattern.MatchString(f.SHA256) || !shareableBundleMember(f) {
				return artifactport.ErrInvalid
			}
			previous = f.Path
			var kind, digest, role, media string
			var data []byte
			var object *string
			var size int64
			err := tx.QueryRow(ctx, `SELECT storage_kind,inline_bytes,object_ref,file_sha256,size_bytes,role,media_type FROM reporting_render_bundle_files WHERE release_id=$1 AND bundle_path=$2`, descriptor.ReleaseID, f.Path).Scan(&kind, &data, &object, &digest, &size, &role, &media)
			if err != nil {
				return err
			}
			if digest != f.SHA256 || size != f.SizeBytes || role != f.Role || media != f.MediaType {
				return artifactport.ErrInvalid
			}
			var reader io.ReadCloser
			switch kind {
			case renderBundleStorageInline:
				if object != nil || int64(len(data)) != size {
					return artifactport.ErrInvalid
				}
				reader = io.NopCloser(bytes.NewReader(data))
			case "object_store":
				if source.Outputs == nil || object == nil || data != nil {
					return artifactport.ErrInvalid
				}
				var info objectstore.ObjectInfo
				reader, info, err = source.Outputs.Get(ctx, objectstore.GetObjectRequest{Key: *object, Purpose: objectstore.PurposeProductRead})
				if err != nil {
					return err
				}
				if info.Size != size {
					_ = reader.Close()
					return artifactport.ErrInvalid
				}
			default:
				return artifactport.ErrInvalid
			}
			h := sha256.New()
			var evidence bytes.Buffer
			stream := io.TeeReader(io.LimitReader(reader, size+1), h)
			capture := f.Role == renderBundleRoleRedactionManifest || f.Role == renderBundleRoleRedactionProfileView || f.Role == renderBundleRoleTokenManifest || f.Role == renderBundleRoleLocalAsset
			if f.Role == renderBundleRoleLocalAsset && size > 50<<20 {
				_ = reader.Close()
				return artifactport.ErrInvalid
			}
			if capture {
				stream = io.TeeReader(stream, &evidence)
			}
			err = write(ctx, portableReleasePath(descriptor.ReleaseID, f.Path), size, stream)
			closeErr := reader.Close()
			if err != nil {
				return err
			}
			if closeErr != nil {
				return closeErr
			}
			if hex.EncodeToString(h.Sum(nil)) != digest {
				return artifactport.ErrInvalid
			}
			if capture {
				publicArtifacts = append(publicArtifacts, RenderBundleFile{RenderBundleManifestFile: f, Bytes: evidence.Bytes()})
			}
			if f.Path == manifest.PrimaryPath && f.MediaType == manifest.PrimaryMediaType {
				primary = true
			}
		}
		if !primary || validatePublicRedactionArtifacts(manifest, publicArtifacts) != nil || validateBundleFilesSelfContained(publicArtifacts) != nil {
			return artifactport.ErrInvalid
		}
		var count int
		if err := tx.QueryRow(ctx, `SELECT count(*) FROM reporting_render_bundle_files WHERE release_id=$1`, descriptor.ReleaseID).Scan(&count); err != nil {
			return err
		}
		if count != len(manifest.Files) {
			return artifactport.ErrInvalid
		}
		if err := write(ctx, descriptor.ManifestPath, int64(len(release.data)), bytes.NewReader(release.data)); err != nil {
			return err
		}
		catalog.Releases = append(catalog.Releases, descriptor)
	}
	slices.SortFunc(catalog.Snapshots, func(a, b portableSnapshot) int { return strings.Compare(a.SnapshotID, b.SnapshotID) })
	slices.SortFunc(catalog.Releases, func(a, b portableRelease) int { return strings.Compare(a.ReleaseID, b.ReleaseID) })
	data, err := canonicaljson.Marshal(catalog)
	if err != nil {
		return err
	}
	if _, err := decodePortableCatalog(data, incident); err != nil {
		return err
	}
	return write(ctx, portableArtifactCatalogPath, int64(len(data)), bytes.NewReader(data))
}
