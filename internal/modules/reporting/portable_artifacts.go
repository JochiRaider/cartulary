package reporting

import (
	"bytes"
	"context"
	"encoding/json"
	"path"
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/artifactport"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

const (
	portableArtifactSchema      = "cartulary.reporting_portable_artifacts.v1"
	portableArtifactCatalogPath = "ext/snapshots/catalog.json"
	portableSnapshotLimit       = 1024
	portableReleaseLimit        = 4096
	portableCatalogBytes        = 16 << 20
)

type portableSnapshot struct {
	SnapshotID  string `json:"snapshot_id"`
	ModelPath   string `json:"export_model_path"`
	ModelSHA256 string `json:"export_model_sha256"`
}
type portableRelease struct {
	ReleaseID      string `json:"release_id"`
	SnapshotID     string `json:"snapshot_id"`
	ManifestPath   string `json:"manifest_path"`
	ManifestSHA256 string `json:"manifest_sha256"`
}
type portableOmission struct {
	ReleaseID string `json:"release_id"`
	Path      string `json:"path"`
	Reason    string `json:"reason"`
}
type portableArtifactCatalog struct {
	SchemaID     string             `json:"schema_id"`
	IncidentID   string             `json:"incident_id"`
	Snapshots    []portableSnapshot `json:"snapshots"`
	Releases     []portableRelease  `json:"releases"`
	OmittedFiles []portableOmission `json:"omitted_files"`
}

type preparedArtifactImport struct {
	incident  uuid.UUID
	operation uuid.UUID
	files     map[string][]byte
}

var portableDigestPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

func emptyPortableCatalog(incident uuid.UUID) portableArtifactCatalog {
	return portableArtifactCatalog{SchemaID: portableArtifactSchema, IncidentID: incident.String(), Snapshots: []portableSnapshot{}, Releases: []portableRelease{}, OmittedFiles: []portableOmission{}}
}
func portableModelPath(id string) string    { return "ext/snapshots/" + id + "/export-model.json" }
func portableManifestPath(id string) string { return "ext/snapshots/releases/" + id + "/manifest.json" }
func portableReleasePath(id, member string) string {
	return "ext/snapshots/releases/" + id + "/files/" + member
}
func canonicalUUID(value string) bool {
	id, err := uuid.Parse(value)
	return err == nil && id != uuid.Nil && id.String() == value
}
func portableMemberPath(value string) bool {
	if value == "" || len(value) > 512 || !utf8.ValidString(value) || strings.ContainsAny(value, "\\:\x00") || strings.HasPrefix(value, "/") || path.Clean(value) != value {
		return false
	}
	for _, part := range strings.Split(value, "/") {
		if part == "." || part == ".." || part == "" {
			return false
		}
	}
	for _, r := range value {
		if r < 32 || r == 127 {
			return false
		}
	}
	return true
}

// Database JSONB loses whitespace and key order. Admit strict JSON, project the
// complete owner shape, then compare logical content before recovering the
// exact owner-produced bytes and their stored digest. This never repairs a
// missing required member or removes an unknown member from signed history.
func decodePortableObject[T any](raw []byte, digest string) (T, []byte, error) {
	var result T
	admitted, err := canonicalReportingBytes(raw)
	if err != nil {
		return result, nil, artifactport.ErrInvalid
	}
	if err = json.Unmarshal(admitted, &result); err != nil {
		return result, nil, artifactport.ErrInvalid
	}
	encoded, err := canonicalJSON(result)
	var identity struct {
		SchemaID string `json:"schema_id"`
	}
	if err != nil || json.Unmarshal(encoded, &identity) != nil || identity.SchemaID == "" || reportingObjectDigest(identity.SchemaID, encoded) != digest {
		return result, nil, artifactport.ErrInvalid
	}
	projected, err := canonicalReportingBytes(encoded)
	if err != nil || !bytes.Equal(admitted, projected) {
		return result, nil, artifactport.ErrInvalid
	}
	return result, encoded, nil
}

func decodePortableCatalog(raw []byte, incident uuid.UUID) (portableArtifactCatalog, error) {
	var c portableArtifactCatalog
	if len(raw) == 0 || len(raw) > portableCatalogBytes {
		return c, artifactport.ErrInvalid
	}
	admitted, err := canonicaljson.Canonicalize(raw)
	if err != nil || !bytes.Equal(admitted, raw) || json.Unmarshal(raw, &c) != nil {
		return c, artifactport.ErrInvalid
	}
	encoded, err := canonicaljson.Marshal(c)
	if err != nil || !bytes.Equal(encoded, raw) || c.SchemaID != portableArtifactSchema || c.IncidentID != incident.String() || c.Snapshots == nil || c.Releases == nil || c.OmittedFiles == nil || len(c.Snapshots) > portableSnapshotLimit || len(c.Releases) > portableReleaseLimit || len(c.OmittedFiles) > portableReleaseLimit {
		return c, artifactport.ErrInvalid
	}
	previous := ""
	snapshots := map[string]bool{}
	for _, s := range c.Snapshots {
		if !canonicalUUID(s.SnapshotID) || s.SnapshotID <= previous || s.ModelPath != portableModelPath(s.SnapshotID) || !portableDigestPattern.MatchString(s.ModelSHA256) {
			return c, artifactport.ErrInvalid
		}
		previous = s.SnapshotID
		snapshots[s.SnapshotID] = true
	}
	previous = ""
	releases := map[string]bool{}
	for _, r := range c.Releases {
		if !canonicalUUID(r.ReleaseID) || r.ReleaseID <= previous || !snapshots[r.SnapshotID] || r.ManifestPath != portableManifestPath(r.ReleaseID) || !portableDigestPattern.MatchString(r.ManifestSHA256) {
			return c, artifactport.ErrInvalid
		}
		previous = r.ReleaseID
		releases[r.ReleaseID] = true
	}
	previous = ""
	for _, o := range c.OmittedFiles {
		order := o.ReleaseID + "\x00" + o.Path
		if !releases[o.ReleaseID] || !portableMemberPath(o.Path) || o.Reason != "sensitive_reveal_map" || order <= previous {
			return c, artifactport.ErrInvalid
		}
		previous = order
	}
	return c, nil
}

// Source attestations remain source evidence. Only Reference Data can admit
// destination content, and every snapshot binding must occur in its exact
// portable catalog before this preparation can be published.
func (PortableReferenceSource) PrepareArtifactImport(ctx context.Context, request artifactport.ImportRequest) (artifactport.Prepared, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if request.IncidentID == uuid.Nil || request.OperationID == uuid.Nil || request.Bundle == nil {
		return nil, artifactport.ErrInvalid
	}
	refs, err := reference_data.DecodeIncidentBundleReferences(request.References)
	if err != nil {
		return nil, artifactport.ErrInvalid
	}
	files := map[string][]byte{}
	for _, name := range request.Bundle.Paths() {
		if strings.HasPrefix(name, "ext/snapshots/") {
			data, ok := request.Bundle.File(name)
			if !ok {
				return nil, artifactport.ErrInvalid
			}
			files[name] = data
		}
	}
	p := &preparedArtifactImport{incident: request.IncidentID, operation: request.OperationID, files: map[string][]byte{}}
	if len(files) == 0 {
		return p, nil
	}
	c, err := decodePortableCatalog(files[portableArtifactCatalogPath], request.IncidentID)
	if err != nil {
		return nil, err
	}
	used := map[string]bool{portableArtifactCatalogPath: true}
	bindings := map[string]reference_data.SetBinding{}
	for _, s := range c.Snapshots {
		model, encoded, err := decodePortableObject[SnapshotModel](files[s.ModelPath], s.ModelSHA256)
		if err != nil || !bytes.Equal(encoded, files[s.ModelPath]) || validateSnapshotModelIdentity(model) != nil || model.DerivationVersion != DerivationVersion || model.SnapshotID != s.SnapshotID || model.IncidentID != c.IncidentID {
			return nil, artifactport.ErrInvalid
		}
		if err := reference_data.ValidatePortableBinding(refs, model.ReferencePacks); err != nil {
			return nil, artifactport.ErrInvalid
		}
		used[s.ModelPath] = true
		bindings[s.SnapshotID] = model.ReferencePacks
	}
	omissions := map[string]bool{}
	for _, o := range c.OmittedFiles {
		omissions[o.ReleaseID+"\x00"+o.Path] = true
	}
	for _, r := range c.Releases {
		manifest, encoded, err := decodePortableObject[RenderBundleManifest](files[r.ManifestPath], r.ManifestSHA256)
		if err != nil || !bytes.Equal(encoded, files[r.ManifestPath]) || validateBundleIdentity(manifest) != nil || manifest.ReleaseID == nil || *manifest.ReleaseID != r.ReleaseID || manifest.SnapshotID != r.SnapshotID || manifest.BundleCreatedAt != manifest.RenderAdmittedAt || manifest.SchemaID != RenderBundleManifestSchemaID || manifest.ManifestVersion != RenderBundleManifestVersion || manifest.Files == nil || len(manifest.Files) == 0 || len(manifest.Files) > 10000 {
			return nil, artifactport.ErrInvalid
		}
		expected, _ := canonicaljson.Marshal(bindings[r.SnapshotID])
		actual, _ := canonicaljson.Marshal(manifest.ReferencePacks)
		if !bytes.Equal(expected, actual) {
			return nil, artifactport.ErrInvalid
		}
		used[r.ManifestPath] = true
		previous := ""
		primary := false
		for _, f := range manifest.Files {
			if !portableMemberPath(f.Path) || f.Path <= previous || f.SizeBytes < 0 || !portableDigestPattern.MatchString(f.SHA256) || f.Role == "" || f.MediaType == "" {
				return nil, artifactport.ErrInvalid
			}
			previous = f.Path
			name := portableReleasePath(r.ReleaseID, f.Path)
			omission := r.ReleaseID + "\x00" + f.Path
			if f.Role == renderBundleRoleSensitiveRevealMap {
				if !omissions[omission] {
					return nil, artifactport.ErrInvalid
				}
				delete(omissions, omission)
				if _, present := files[name]; present {
					return nil, artifactport.ErrInvalid
				}
				continue
			}
			data, ok := files[name]
			if !ok || int64(len(data)) != f.SizeBytes || hashHex(data) != f.SHA256 {
				return nil, artifactport.ErrInvalid
			}
			used[name] = true
			if f.Path == manifest.PrimaryPath && f.MediaType == manifest.PrimaryMediaType {
				primary = true
			}
		}
		if !primary {
			return nil, artifactport.ErrInvalid
		}
	}
	if len(omissions) != 0 || len(used) != len(files) {
		return nil, artifactport.ErrInvalid
	}
	for name, data := range files {
		p.files[name] = bytes.Clone(data)
	}
	return p, nil
}

func (p *preparedArtifactImport) ApplyTx(ctx context.Context, tx pgx.Tx) error {
	if tx == nil || p == nil {
		return artifactport.ErrInvalid
	}
	names := make([]string, 0, len(p.files))
	for name := range p.files {
		names = append(names, name)
	}
	slices.Sort(names)
	for _, name := range names {
		data := p.files[name]
		// Exact replay is idempotent; an existing identity with different bytes or
		// operation attribution is never overwritten.
		tag, err := tx.Exec(ctx, `INSERT INTO reporting_imported_artifact_files(incident_id,bundle_path,operation_id,size_bytes,sha256,content) VALUES($1,$2,$3,$4,$5,$6) ON CONFLICT DO NOTHING`, p.incident, name, p.operation, len(data), hashHex(data), data)
		if err != nil {
			return err
		}
		if tag.RowsAffected() == 0 {
			var same bool
			err = tx.QueryRow(ctx, `SELECT operation_id=$3 AND content=$4 FROM reporting_imported_artifact_files WHERE incident_id=$1 AND bundle_path=$2`, p.incident, name, p.operation, data).Scan(&same)
			if err != nil {
				return err
			}
			if !same {
				return artifactport.ErrInvalid
			}
		}
	}
	return nil
}

func loadImportedArtifactFiles(ctx context.Context, tx pgx.Tx, incident uuid.UUID) (map[string][]byte, error) {
	rows, err := tx.Query(ctx, `SELECT bundle_path,content,sha256,size_bytes FROM reporting_imported_artifact_files WHERE incident_id=$1 ORDER BY bundle_path`, incident)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	files := map[string][]byte{}
	for rows.Next() {
		var name, digest string
		var data []byte
		var size int64
		if err := rows.Scan(&name, &data, &digest, &size); err != nil {
			return nil, err
		}
		if int64(len(data)) != size || hashHex(data) != digest || len(files) >= 10000 {
			return nil, artifactport.ErrInvalid
		}
		files[name] = data
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return files, nil
}

var _ artifactport.Prepared = (*preparedArtifactImport)(nil)
