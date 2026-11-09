package packformat

import (
	"bytes"
	"context"
	"encoding/json"
	"regexp"
	"slices"
	"strings"
	"unicode/utf8"

	"github.com/JochiRaider/cartulary/referencepacks/format/canonicaljson"
	"github.com/JochiRaider/cartulary/referencepacks/format/envelope"
	"golang.org/x/text/unicode/norm"
)

const ContractID = "cartulary.reference_pack_contract.v3"

// SafeManifestIdentity attributes a candidate without trusting its content or
// materializing producer defaults. Every other field remains unverified.
func SafeManifestIdentity(data []byte) (string, string) {
	value, err := canonicaljson.DecodeStrict(data)
	if err != nil {
		return "", ""
	}
	m, ok := value.(map[string]any)
	if !ok {
		return "", ""
	}
	key, ok := m["pack_key"].(string)
	version, versionOK := m["pack_version"].(string)
	if !ok || !versionOK || len(key) > 128 || !keyPattern.MatchString(key) || !versionPattern.MatchString(version) {
		return "", ""
	}
	return key, version
}

type Profile struct {
	Key          string            `json:"pack_key"`
	Kind         string            `json:"pack_kind"`
	ID           string            `json:"content_profile_id"`
	Version      string            `json:"content_profile_version"`
	Shape        string            `json:"payload_shape"`
	Authority    string            `json:"authority_class"`
	Required     bool              `json:"required_base"`
	Schemas      map[string]string `json:"schemas"`
	MaxLineBytes int               `json:"max_line_bytes"`
}
type Manifest = envelope.Manifest
type SourceArtifact = envelope.SourceArtifact
type Dependency = envelope.Dependency
type Conflict = envelope.Conflict
type File = envelope.File
type License = envelope.License
type LicenseBinding = envelope.LicenseBinding

var manifestShape = compileProjection("manifest.v2.schema.json")
var hintShape = compileProjection("bundle_hint.v1.schema.json")
var profiles = loadProfiles()
var keyPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,31}(\.[a-z][a-z0-9_]{0,31}){1,7}$`)
var versionPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._+-]{0,127}$`)

func loadProfiles() map[string]Profile {
	var catalog struct {
		Profiles []Profile `json:"profiles"`
	}
	if err := json.Unmarshal(projection("profiles.v3.json"), &catalog); err != nil {
		panic("invalid reference pack profile projection")
	}
	result := map[string]Profile{}
	for _, profile := range catalog.Profiles {
		result[profile.Key] = profile
	}
	return result
}

// ProfileForKey reserves every standardized key and the registry namespace.
// Unregistered advisory keys share the generic model; content cannot invent
// registry authority or substitute a different profile for a reserved key.
func ProfileForKey(key string) (Profile, bool) {
	if p, ok := profiles[key]; ok {
		return p, true
	}
	if len(key) > 128 || !keyPattern.MatchString(key) || strings.HasPrefix(key, "type_registry.") {
		return Profile{}, false
	}
	return Profile{Key: key, ID: "cartulary.reference_pack.dataset.v1", Version: "1", Shape: "dataset", Schemas: map[string]string{"descriptor": "profiles/dataset.descriptor.v1.schema.json", "object": "profiles/dataset.object.v1.schema.json", "relationship": "profiles/dataset.relationship.v1.schema.json"}, MaxLineBytes: 4194304}, true
}

// ManifestAuthority resolves policy from an admitted manifest, not the name of
// an unregistered key. A content model alone cannot supply source meaning.
func ManifestAuthority(m Manifest) (string, error) {
	if err := ValidateManifestContract(m); err != nil {
		return "", err
	}
	if p, standard := profiles[m.Key]; standard {
		return p.Authority, nil
	}
	switch m.Kind {
	case "framework":
		return "framework_reference", nil
	case "enrichment":
		return "advisory_enrichment", nil
	default:
		return "", fail("contract_incompatible")
	}
}
func Profiles() []Profile {
	result := make([]Profile, 0, len(profiles))
	for _, p := range profiles {
		result = append(result, p)
	}
	slices.SortFunc(result, func(a, b Profile) int { return strings.Compare(a.Key, b.Key) })
	return result
}
func strictDocument(data []byte, shape *shape, invalid, noncanonical string) (map[string]any, error) {
	value, err := canonicaljson.DecodeStrict(data)
	if err != nil {
		return nil, fail(invalid)
	}
	m, ok := object(value)
	if !ok || !validPackNumbers(value) || !shape.matches(value) {
		return nil, fail(invalid)
	}
	encoded, err := canonicaljson.Marshal(value)
	if err != nil {
		return nil, fail(invalid)
	}
	if !bytes.Equal(encoded, data) {
		return nil, fail(noncanonical)
	}
	return m, nil
}
func admittedHintSize(size int) bool { return size > 0 && size <= 16384 }

func DecodeHint(data []byte) (string, error) {
	if !admittedHintSize(len(data)) {
		return "", fail("bundle_hint_invalid")
	}
	m, err := strictDocument(data, hintShape, "bundle_hint_invalid", "bundle_hint_noncanonical")
	if err != nil {
		return "", err
	}
	return m["trust_repository_id"].(string), nil
}
func singleLine(value string, max int) bool {
	if value == "" || !utf8.ValidString(value) || !norm.NFC.IsNormalString(value) || strings.TrimSpace(value) != value || utf8.RuneCountInString(value) > max {
		return false
	}
	for _, r := range value {
		if r < 32 || (r >= 127 && r <= 159) || r == 0x2028 || r == 0x2029 {
			return false
		}
	}
	return true
}
func DecodeManifest(data []byte, operator bool) (Manifest, error) {
	m, err := DecodeManifestSchema(data, operator)
	if err != nil {
		return m, err
	}
	for _, check := range []func(Manifest) error{ValidateManifestContract, ValidateManifestLicense, ValidateRuntimeCompatibility, ValidateManifestDependencies, ValidateManifestConflicts} {
		if err := check(m); err != nil {
			return m, err
		}
	}
	return m, nil
}

// DecodeManifestSchema admits syntax and the closed manifest shape only.
// Cross-phase checks are separate so the live verifier cannot let a later
// compatibility, license or dependency failure hide an earlier content check.
func DecodeManifestSchema(data []byte, operator bool) (Manifest, error) {
	program, result := ManifestChecks(data, int64(len(data)), operator)
	if err := runCheckRange(context.Background(), "manifest_encoding", "manifest_canonical", nil, program, false); err != nil {
		return Manifest{}, err
	}
	return result(), nil
}

func ValidateManifestContract(m Manifest) error {
	c := &semanticChecks{}
	manifestCompatibility(m, false, c)
	return c.finish("contract_incompatible")
}

func ValidateManifestLicense(m Manifest) error {
	if !validLicense(m.License) {
		return fail("content_semantic_invalid")
	}
	return nil
}

func ValidateRuntimeCompatibility(m Manifest) error {
	c := &semanticChecks{}
	manifestCompatibility(m, true, c)
	return c.finish("contract_incompatible")
}

func ValidateManifestDependencies(m Manifest) error {
	if profiles[m.Key].Required && (len(m.Dependencies) != 0 || len(m.Conflicts) != 0) {
		return fail("dependency_unsatisfied")
	}
	return nil
}

func ValidateManifestConflicts(m Manifest) error {
	for _, conflict := range m.Conflicts {
		if target, ok := profiles[conflict.Key]; ok && target.Required {
			return fail("pack_conflict")
		}
	}
	return nil
}

func PayloadDigest(files []File) (string, error) {
	value, err := envelope.PayloadDigest(files)
	return value, formatError(err)
}
func ValidateNotice(data []byte) error { return formatError(envelope.ValidateNotice(data)) }
