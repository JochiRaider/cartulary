package incidentbundles

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"io"
	"sort"
	"strings"

	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"github.com/google/uuid"
)

// Catalogs are bounded control data. Every bulk byte remains in a separately
// sealed member governed by the archive owner's cumulative resource limits.
type extensionMemberCatalog struct {
	SchemaID        string                       `json:"schema_id"`
	ProfileID       string                       `json:"profile_id"`
	ContractMajor   int                          `json:"contract_major"`
	StateVersion    int                          `json:"state_version"`
	ContentSchemaID string                       `json:"content_schema_id"`
	Members         []incidentportability.Member `json:"members"`
}

type admittedExtensionCatalog struct {
	extensionMemberCatalog
	bound    *incidentportability.BoundMembers
	byteSize int64
	sha256   string
}

func extensionCatalogPath(profileID string) string {
	return "ext/extensions/" + profileID + "/catalog.json"
}
func extensionMemberPrefix(profileID string) string {
	return "ext/extensions/" + profileID + "/members/"
}

func encodeExtensionCatalog(catalog extensionMemberCatalog) ([]byte, error) {
	if !validExtensionCatalog(catalog) {
		return nil, ErrPortabilityPayload
	}
	return canonicaljson.MarshalExtension(catalog)
}

func validExtensionCatalog(c extensionMemberCatalog) bool {
	if c.SchemaID != ExtensionMemberCatalogSchema || !validExtensionProfileID(c.ProfileID) || c.ContractMajor < 1 || c.ContractMajor > 2147483647 || c.StateVersion < 1 || c.StateVersion > 2147483647 || c.ContentSchemaID == "" || len(c.ContentSchemaID) > 160 || len(c.Members) == 0 || len(c.Members) > int(defaultArchiveMaxMembers) {
		return false
	}
	previous := ""
	for _, member := range c.Members {
		if !safeBundlePath(member.Path) || !strings.HasPrefix(member.Path, extensionMemberPrefix(c.ProfileID)) || member.Path <= previous || member.ByteSize < 0 || !validMemberSHA(member.SHA256) {
			return false
		}
		previous = member.Path
	}
	return true
}

func validExtensionProfileID(id string) bool {
	if len(id) < 1 || len(id) > 64 || id[0] < 'a' || id[0] > 'z' {
		return false
	}
	for _, c := range id {
		if !(c >= 'a' && c <= 'z' || c >= '0' && c <= '9' || c == '_') {
			return false
		}
	}
	return true
}
func validMemberSHA(value string) bool {
	if len(value) != 64 || strings.ToLower(value) != value {
		return false
	}
	_, err := hex.DecodeString(value)
	return err == nil
}

// The shared owner owns the manifest and records what the participant actually
// emitted. A participant cannot overwrite Core data or another profile's files.
type extensionMemberWriter struct {
	destination incidentportability.MemberWriter
	profileID   string
	paths       map[string]incidentportability.Member
	closed      bool
}

func (w *extensionMemberWriter) WriteMember(ctx context.Context, path string, source io.Reader) (incidentportability.Member, error) {
	if w.closed || !safeBundlePath(path) || !strings.HasPrefix(path, extensionMemberPrefix(w.profileID)) || len(w.paths) >= int(defaultArchiveMaxMembers) {
		return incidentportability.Member{}, ErrPortabilityPayload
	}
	if _, duplicate := w.paths[path]; duplicate {
		return incidentportability.Member{}, ErrPortabilityPayload
	}
	member, err := w.destination.WriteMember(ctx, path, source)
	if err != nil {
		return incidentportability.Member{}, err
	}
	if member.Path != path || member.ByteSize < 0 || !validMemberSHA(member.SHA256) {
		return incidentportability.Member{}, ErrPortabilityResult
	}
	w.paths[path] = member
	return member, nil
}
func (w *extensionMemberWriter) members() []incidentportability.Member {
	result := make([]incidentportability.Member, 0, len(w.paths))
	for _, member := range w.paths {
		result = append(result, member)
	}
	sort.Slice(result, func(i, j int) bool { return result[i].Path < result[j].Path })
	return result
}

func decodeExtensionCatalogs(ctx context.Context, source incidentportability.MemberReader) ([]admittedExtensionCatalog, error) {
	if source == nil {
		return nil, ErrPortabilityPayload
	}
	members := map[string]incidentportability.Member{}
	paths := []string{}
	for _, member := range source.Members() {
		if !strings.HasPrefix(member.Path, "ext/extensions/") {
			continue
		}
		if _, duplicate := members[member.Path]; duplicate {
			return nil, ErrPortabilityPayload
		}
		members[member.Path] = member
		parts := strings.Split(member.Path, "/")
		if len(parts) == 4 && parts[3] == "catalog.json" {
			paths = append(paths, member.Path)
		}
	}
	sort.Strings(paths)
	result := make([]admittedExtensionCatalog, 0, len(paths))
	var total int64
	claimed := map[string]bool{}
	for _, path := range paths {
		descriptor := members[path]
		if descriptor.ByteSize < 0 || descriptor.ByteSize > PortabilityParticipantByteLimit-total {
			return nil, ErrPortabilityLimit
		}
		total += descriptor.ByteSize
		raw, err := incidentportability.ReadControl(ctx, source, path, PortabilityParticipantByteLimit)
		if err != nil {
			return nil, ErrPortabilityPayload
		}
		if _, err := canonicaljson.DecodeExtension(raw); err != nil {
			return nil, ErrPortabilityPayload
		}
		var catalog extensionMemberCatalog
		decoder := json.NewDecoder(bytes.NewReader(raw))
		decoder.DisallowUnknownFields()
		if err := decoder.Decode(&catalog); err != nil || !validExtensionCatalog(catalog) || path != extensionCatalogPath(catalog.ProfileID) {
			return nil, ErrPortabilityPayload
		}
		claimed[path] = true
		ownedPaths := make([]string, 0, len(catalog.Members))
		for _, member := range catalog.Members {
			actual, exists := members[member.Path]
			if !exists || actual != member || claimed[member.Path] {
				return nil, ErrPortabilityPayload
			}
			claimed[member.Path] = true
			ownedPaths = append(ownedPaths, member.Path)
		}
		bound, err := incidentportability.BindMembers(source, ownedPaths)
		if err != nil {
			return nil, ErrPortabilityPayload
		}
		result = append(result, admittedExtensionCatalog{extensionMemberCatalog: catalog, bound: bound, byteSize: descriptor.ByteSize, sha256: digestBytes(raw)})
	}
	if len(claimed) != len(members) {
		return nil, ErrPortabilityPayload
	}
	return result, nil
}

type actorReferenceSet struct{ ids map[string]struct{} }

func newActorReferenceSet() *actorReferenceSet { return &actorReferenceSet{ids: map[string]struct{}{}} }
func (s *actorReferenceSet) RecordActorReference(id uuid.UUID) error {
	if s == nil || id == uuid.Nil {
		return actorReferenceFailure()
	}
	if _, exists := s.ids[id.String()]; exists {
		return nil
	}
	if len(s.ids) >= PortabilityParticipantByteLimit/36 {
		return ErrPortabilityLimit
	}
	s.ids[id.String()] = struct{}{}
	return nil
}
