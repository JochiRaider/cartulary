package packformat

import (
	"encoding/json"
	"slices"
	"strings"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

const PortableReferencesSchemaID = "reference_pack_refs.v1"

type PortableVersion struct {
	SetMember
	DistributionKind    string `json:"distribution_kind"`
	VerificationMethod  string `json:"verification_method"`
	SourceProfileID     string `json:"source_profile_id"`
	SourceProfileSHA256 string `json:"source_profile_sha256"`
}

type PortableReferences struct {
	SchemaID string            `json:"schema_id"`
	Sets     []Set             `json:"sets"`
	Versions []PortableVersion `json:"versions"`
}

var portableReferencesShape = compileProjection("reference_pack_refs.v1.schema.json")

// DecodePortableReferences authenticates no content. It admits the complete
// reference graph before any destination lookup or trust decision is made.
func DecodePortableReferences(data []byte) (PortableReferences, error) {
	if !admittedPortableReferenceSize(len(data)) {
		return PortableReferences{}, fail("reference_pack_refs.exact_shape")
	}
	value, err := canonicaljson.DecodeStrict(data)
	if err != nil || !portableReferencesShape.matches(value) {
		return PortableReferences{}, fail("reference_pack_refs.exact_shape")
	}
	var refs PortableReferences
	if err := json.Unmarshal(data, &refs); err != nil {
		return PortableReferences{}, fail("reference_pack_refs.exact_shape")
	}
	bad := func() (PortableReferences, error) {
		return PortableReferences{}, fail("reference_pack_refs.identity_exact")
	}
	versions := make(map[string]PortableVersion, len(refs.Versions))
	previous := ""
	for _, v := range refs.Versions {
		tuple := v.Key + "\x00" + v.Version
		order := tuple + "\x00" + v.ManifestSHA256 + "\x00" + v.PayloadSHA256
		if order <= previous {
			return bad()
		}
		previous = order
		if _, exists := versions[tuple]; exists {
			return bad()
		}
		profile, ok := profiles[v.Key]
		if !ok || v.ProfileID != profile.ID || v.ProfileVersion != profile.Version ||
			(v.DistributionKind == "packaged_builtin" && v.VerificationMethod != "packaged_release_manifest_v1") ||
			(v.DistributionKind == "operator_imported" && v.VerificationMethod != "tuf_1_0_35_offline_bundle_v1") {
			return bad()
		}
		versions[tuple] = v
	}
	used := make(map[string]bool, len(versions))
	previous = ""
	for _, set := range refs.Sets {
		if set.ID <= previous {
			return bad()
		}
		previous = set.ID
		expected, err := BuildSet(set.Members)
		if err != nil || !equalJSON(set, expected) {
			return bad()
		}
		for _, member := range set.Members {
			tuple := member.Key + "\x00" + member.Version
			version, ok := versions[tuple]
			if !ok || version.SetMember != member {
				return bad()
			}
			used[tuple] = true
		}
	}
	if len(used) != len(versions) {
		return bad()
	}
	return refs, nil
}

func admittedPortableReferenceSize(size int) bool { return size > 0 && size <= 67108864 }

// EncodePortableReferences is a producer operation. It orders supplied facts
// but never fills in missing content identity, provenance, or trust facts.
func EncodePortableReferences(sets []Set, versions []PortableVersion) ([]byte, error) {
	refs := PortableReferences{SchemaID: PortableReferencesSchemaID, Sets: append([]Set{}, sets...), Versions: append([]PortableVersion{}, versions...)}
	slices.SortFunc(refs.Sets, func(a, b Set) int { return strings.Compare(a.ID, b.ID) })
	slices.SortFunc(refs.Versions, func(a, b PortableVersion) int {
		return strings.Compare(a.Key+"\x00"+a.Version+"\x00"+a.ManifestSHA256+"\x00"+a.PayloadSHA256, b.Key+"\x00"+b.Version+"\x00"+b.ManifestSHA256+"\x00"+b.PayloadSHA256)
	})
	data, err := canonicaljson.Marshal(refs)
	if err != nil {
		return nil, err
	}
	if _, err := DecodePortableReferences(data); err != nil {
		return nil, err
	}
	return append(data, '\n'), nil
}
