package reporting

import (
	"bytes"
	"context"
	"encoding/json"
	"path"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/artifactport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"github.com/google/uuid"
	"golang.org/x/text/unicode/norm"
)

const (
	portableArtifactSchema      = "cartulary.reporting_portable_artifacts.v3"
	portableArtifactCatalogPath = "ext/snapshots/catalog.json"
	portableSnapshotLimit       = 1024
	portableReleaseLimit        = 4096
	portableCatalogBytes        = 16 << 20
)

type portableSnapshot struct {
	SnapshotID    string `json:"snapshot_id"`
	CapturePath   string `json:"source_capture_path"`
	CaptureSHA256 string `json:"source_capture_sha256"`
}
type portableRelease struct {
	ReleaseID      string `json:"release_id"`
	SnapshotID     string `json:"snapshot_id"`
	ManifestPath   string `json:"manifest_path"`
	ManifestSHA256 string `json:"manifest_sha256"`
}
type portableArtifactCatalog struct {
	SchemaID   string             `json:"schema_id"`
	IncidentID string             `json:"incident_id"`
	Snapshots  []portableSnapshot `json:"snapshots"`
	Releases   []portableRelease  `json:"releases"`
}

type preparedArtifactImport struct {
	incident  uuid.UUID
	operation uuid.UUID
	members   *incidentportability.BoundMembers
}

var portableDigestPattern = regexp.MustCompile(`^[0-9a-f]{64}$`)

func emptyPortableCatalog(incident uuid.UUID) portableArtifactCatalog {
	return portableArtifactCatalog{SchemaID: portableArtifactSchema, IncidentID: incident.String(), Snapshots: []portableSnapshot{}, Releases: []portableRelease{}}
}
func portableCapturePath(id string) string  { return "ext/snapshots/" + id + "/source-capture.json" }
func portableManifestPath(id string) string { return "ext/snapshots/releases/" + id + "/manifest.json" }
func portableReleasePath(id, member string) string {
	return "ext/snapshots/releases/" + id + "/files/" + member
}
func canonicalUUID(value string) bool {
	id, err := uuid.Parse(value)
	return err == nil && id != uuid.Nil && id.String() == value
}
func portableMemberPath(value string) bool {
	if value == "" || len(value) > 512 || !utf8.ValidString(value) || !norm.NFC.IsNormalString(value) || strings.ContainsAny(value, "\\:\x00") || strings.HasPrefix(value, "/") || path.Clean(value) != value {
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
	decoder := json.NewDecoder(bytes.NewReader(admitted))
	decoder.UseNumber()
	if err = decoder.Decode(&result); err != nil {
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

func admittedPortableCatalogSize(size int) bool {
	return size > 0 && size <= portableCatalogBytes
}

func decodePortableCatalog(raw []byte, incident uuid.UUID) (portableArtifactCatalog, error) {
	var c portableArtifactCatalog
	if !admittedPortableCatalogSize(len(raw)) {
		return c, artifactport.ErrInvalid
	}
	admitted, err := canonicaljson.Canonicalize(raw)
	if err != nil || !bytes.Equal(admitted, raw) || json.Unmarshal(raw, &c) != nil {
		return c, artifactport.ErrInvalid
	}
	encoded, err := canonicaljson.Marshal(c)
	if err != nil || !bytes.Equal(encoded, raw) || c.SchemaID != portableArtifactSchema || c.IncidentID != incident.String() || c.Snapshots == nil || c.Releases == nil || len(c.Snapshots) > portableSnapshotLimit || len(c.Releases) > portableReleaseLimit {
		return c, artifactport.ErrInvalid
	}
	previous := ""
	snapshots := map[string]bool{}
	for _, s := range c.Snapshots {
		if !canonicalUUID(s.SnapshotID) || s.SnapshotID <= previous || s.CapturePath != portableCapturePath(s.SnapshotID) || !portableDigestPattern.MatchString(s.CaptureSHA256) {
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
	paths := []string{}
	for _, member := range request.Bundle.Members() {
		if strings.HasPrefix(member.Path, "ext/snapshots/") {
			paths = append(paths, member.Path)
		}
	}
	members, err := incidentportability.BindMembers(request.Bundle, paths)
	if err != nil {
		return nil, artifactport.ErrInvalid
	}
	p := &preparedArtifactImport{incident: request.IncidentID, operation: request.OperationID, members: members}
	if len(paths) == 0 {
		return p, nil
	}
	catalogBytes, err := incidentportability.ReadControl(ctx, members, portableArtifactCatalogPath, portableCatalogBytes)
	if err != nil {
		return nil, artifactport.ErrInvalid
	}
	c, err := decodePortableCatalog(catalogBytes, request.IncidentID)
	if err != nil {
		return nil, err
	}
	used := map[string]bool{portableArtifactCatalogPath: true}
	bindings := map[string]reference_data.SetBinding{}
	for _, s := range c.Snapshots {
		raw, err := incidentportability.ReadControl(ctx, members, s.CapturePath, incidentportability.MaxControlMemberBytes)
		if err != nil {
			return nil, artifactport.ErrInvalid
		}
		model, encoded, err := decodePortableObject[snapshotCapture](raw, s.CaptureSHA256)
		if err != nil || !bytes.Equal(encoded, raw) || validateSnapshotCapture(model) != nil || model.DerivationVersion != DerivationVersion || model.SnapshotID != s.SnapshotID || model.IncidentID != c.IncidentID {
			return nil, artifactport.ErrInvalid
		}
		if err := reference_data.ValidatePortableBinding(refs, model.ReferencePacks); err != nil {
			return nil, artifactport.ErrInvalid
		}
		used[s.CapturePath] = true
		bindings[s.SnapshotID] = model.ReferencePacks
	}
	for _, r := range c.Releases {
		raw, err := incidentportability.ReadControl(ctx, members, r.ManifestPath, incidentportability.MaxControlMemberBytes)
		if err != nil {
			return nil, artifactport.ErrInvalid
		}
		manifest, encoded, err := decodePortableObject[RenderBundleManifest](raw, r.ManifestSHA256)
		if err != nil || !bytes.Equal(encoded, raw) || validateBundleIdentity(manifest) != nil || manifest.ReleaseID == nil || *manifest.ReleaseID != r.ReleaseID || manifest.SnapshotID != r.SnapshotID || manifest.BundleCreatedAt != manifest.RenderAdmittedAt || manifest.SchemaID != RenderBundleManifestSchemaID || manifest.ManifestVersion != RenderBundleManifestVersion || manifest.Files == nil || len(manifest.Files) == 0 || len(manifest.Files) > 10000 {
			return nil, artifactport.ErrInvalid
		}
		expected, _ := canonicaljson.Marshal(bindings[r.SnapshotID])
		actual, _ := canonicaljson.Marshal(manifest.ReferencePacks)
		if !bytes.Equal(expected, actual) {
			return nil, artifactport.ErrInvalid
		}
		used[r.ManifestPath] = true
		if err := validatePortableReleaseMembers(ctx, members, manifest, r.ReleaseID, used); err != nil {
			return nil, err
		}
	}

	if len(used) != len(paths) {
		return nil, artifactport.ErrInvalid
	}
	return p, nil
}

var _ artifactport.Prepared = (*preparedArtifactImport)(nil)
