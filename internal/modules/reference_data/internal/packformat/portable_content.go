package packformat

import (
	"encoding/json"
	"slices"
	"strings"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

const PortableContentPath = "ext/reference_packs/content.json"
const PortableContentSchemaID = "reference_pack_content.v1"

type PortableContainer struct {
	ManifestSHA256  string `json:"manifest_sha256"`
	ContainerSHA256 string `json:"container_sha256"`
	SizeBytes       int64  `json:"size_bytes"`
}

type PortableRequiredMember struct {
	SetID string `json:"pack_set_id"`
	Key   string `json:"pack_key"`
}

type PortableContent struct {
	SchemaID        string                   `json:"schema_id"`
	Containers      []PortableContainer      `json:"containers"`
	RequiredMembers []PortableRequiredMember `json:"required_members"`
}

var portableContentShape = compileProjection("reference_pack_content.v1.schema.json")

func EmptyPortableContent() PortableContent {
	return PortableContent{SchemaID: PortableContentSchemaID, Containers: []PortableContainer{}, RequiredMembers: []PortableRequiredMember{}}
}

// DecodePortableContent admits selection metadata, never the referenced pack
// bytes or source trust. References must already have passed their closed
// catalog admission. Absence is handled by the enclosing transport; nil is not
// a valid present document.
func DecodePortableContent(data []byte, refs PortableReferences) (PortableContent, error) {
	bad := func(code string) (PortableContent, error) { return PortableContent{}, fail(code) }
	if len(data) == 0 || len(data) > 16777216 {
		return bad("reference_pack_refs.exact_shape")
	}
	value, err := canonicaljson.DecodeStrict(data)
	if err != nil || !portableContentShape.matches(value) {
		return bad("reference_pack_refs.exact_shape")
	}
	var content PortableContent
	if err := json.Unmarshal(data, &content); err != nil {
		return bad("reference_pack_refs.exact_shape")
	}
	versions := map[string]PortableVersion{}
	ambiguous := map[string]bool{}
	for _, version := range refs.Versions {
		if _, exists := versions[version.ManifestSHA256]; exists {
			ambiguous[version.ManifestSHA256] = true
		}
		versions[version.ManifestSHA256] = version
	}
	previous := ""
	for _, container := range content.Containers {
		v, ok := versions[container.ManifestSHA256]
		if container.ManifestSHA256 <= previous || !ok || ambiguous[container.ManifestSHA256] || v.DistributionKind != "operator_imported" {
			return bad("reference_pack_refs.identity_exact")
		}
		previous = container.ManifestSHA256
	}
	sets := map[string]Set{}
	for _, set := range refs.Sets {
		sets[set.ID] = set
	}
	previous = ""
	for _, required := range content.RequiredMembers {
		id := required.SetID + "\x00" + required.Key
		set, ok := sets[required.SetID]
		if id <= previous || !ok || !slices.ContainsFunc(set.Members, func(member SetMember) bool { return member.Key == required.Key }) {
			return bad("reference_pack_refs.identity_exact")
		}
		previous = id
	}
	return content, nil
}

func PortableContainerPath(manifestSHA256 string) (string, error) {
	if _, ok := hexBytes(manifestSHA256, 32); !ok {
		return "", fail("reference_pack_refs.identity_exact")
	}
	return "ext/reference_packs/containers/" + manifestSHA256 + ".pack", nil
}

// EncodePortableContent orders producer-owned facts but cannot add defaults,
// invent required membership or promote an embedded built-in to operator trust.
func EncodePortableContent(content PortableContent, refs PortableReferences) ([]byte, error) {
	content.Containers = slices.Clone(content.Containers)
	content.RequiredMembers = slices.Clone(content.RequiredMembers)
	slices.SortFunc(content.Containers, func(a, b PortableContainer) int { return strings.Compare(a.ManifestSHA256, b.ManifestSHA256) })
	slices.SortFunc(content.RequiredMembers, func(a, b PortableRequiredMember) int {
		return strings.Compare(a.SetID+"\x00"+a.Key, b.SetID+"\x00"+b.Key)
	})
	data, err := canonicaljson.Marshal(content)
	if err != nil {
		return nil, err
	}
	if _, err := DecodePortableContent(data, refs); err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}

// RequiredPortableVersions projects exact source-set requirements onto the
// already admitted version catalog. It never consults a current active set.
func RequiredPortableVersions(content PortableContent, refs PortableReferences) map[string]bool {
	required := map[string]bool{}
	sets := map[string]Set{}
	for _, set := range refs.Sets {
		sets[set.ID] = set
	}
	for _, item := range content.RequiredMembers {
		for _, member := range sets[item.SetID].Members {
			if member.Key == item.Key {
				required[member.Key+"\x00"+member.Version] = true
			}
		}
	}
	return required
}

// ValidatePortableContainerInventory admits only exact, declared namespace
// members. Missing optional bytes are resolved later; arbitrary aliases and
// unreferenced containers do not receive an implicit selection.
func ValidatePortableContainerInventory(content PortableContent, paths []string) error {
	allowed := map[string]bool{}
	for _, container := range content.Containers {
		path, err := PortableContainerPath(container.ManifestSHA256)
		if err != nil {
			return err
		}
		allowed[path] = true
	}
	seen := map[string]bool{}
	for _, path := range paths {
		if !allowed[path] || seen[path] {
			return fail("reference_pack_refs.identity_exact")
		}
		seen[path] = true
	}
	return nil
}
