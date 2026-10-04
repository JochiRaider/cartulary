package packformat

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"regexp"
	"slices"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"golang.org/x/text/unicode/norm"
)

const ContractID = "cartulary.reference_pack_contract.v1"

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
	Key       string `json:"pack_key"`
	Kind      string `json:"pack_kind"`
	ID        string `json:"content_profile_id"`
	Version   string `json:"content_profile_version"`
	Shape     string `json:"payload_shape"`
	Authority string `json:"authority_class"`
	Required  bool   `json:"required_base"`
}
type Manifest struct {
	SchemaID            string  `json:"schema_id"`
	Contract            string  `json:"pack_contract_version"`
	Key                 string  `json:"pack_key"`
	Kind                string  `json:"pack_kind"`
	Version             string  `json:"pack_version"`
	Sequence            int64   `json:"pack_release_sequence"`
	ProfileID           string  `json:"content_profile_id"`
	ProfileVersion      string  `json:"content_profile_version"`
	SourceProfileID     string  `json:"source_profile_id"`
	SourceProfileSHA256 string  `json:"source_profile_sha256"`
	Repository          *string `json:"trust_repository_id"`
	SourceIdentifier    string  `json:"source_identifier"`
	SourceVersion       string  `json:"source_version"`
	SourceAsOf          *string `json:"source_as_of"`
	BuiltAt             string  `json:"built_at"`
	Builder             struct {
		ID      string `json:"builder_id"`
		Version string `json:"builder_version"`
		SHA256  string `json:"builder_source_sha256"`
	} `json:"builder"`
	Artifacts     []SourceArtifact `json:"source_artifacts"`
	Compatibility struct {
		Majors       []int    `json:"cartulary_contract_majors"`
		Capabilities []string `json:"required_capabilities"`
	} `json:"compatibility"`
	Dependencies []Dependency `json:"dependencies"`
	Conflicts    []Conflict   `json:"conflicts"`
	Files        []File       `json:"files"`
	Summary      struct {
		Kind          string `json:"kind"`
		Entries       *int64 `json:"entry_count,omitempty"`
		Objects       *int64 `json:"object_count,omitempty"`
		Relationships *int64 `json:"relationship_count,omitempty"`
	} `json:"content_summary"`
	License    License        `json:"license"`
	Extensions map[string]any `json:"extensions"`
}
type SourceArtifact struct {
	Ref     string `json:"source_ref"`
	Version string `json:"source_version"`
	SHA256  string `json:"sha256"`
}
type Dependency struct {
	Key     string `json:"pack_key"`
	Version string `json:"pack_version"`
	SHA256  string `json:"payload_sha256"`
}
type Conflict struct {
	Key     string  `json:"pack_key"`
	Version *string `json:"pack_version"`
}
type File struct {
	Path      string `json:"path"`
	Role      string `json:"role"`
	MediaType string `json:"media_type"`
	Size      int64  `json:"size_bytes"`
	SHA256    string `json:"sha256"`
}
type License struct {
	Expression     string           `json:"expression"`
	ListVersion    string           `json:"license_list_version"`
	Redistribution string           `json:"redistribution"`
	Notices        []string         `json:"notice_paths"`
	Bindings       []LicenseBinding `json:"license_ref_bindings"`
}
type LicenseBinding struct {
	Ref  string `json:"license_ref"`
	Path string `json:"notice_path"`
}

var manifestShape = compileProjection("manifest.v1.schema.json")
var hintShape = compileProjection("bundle_hint.v1.schema.json")
var profiles = loadProfiles()
var keyPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,31}(\.[a-z][a-z0-9_]{0,31}){1,7}$`)
var versionPattern = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9._+-]{0,127}$`)

func loadProfiles() map[string]Profile {
	var catalog struct {
		Profiles []Profile `json:"profiles"`
	}
	if err := json.Unmarshal(projection("profiles.v1.json"), &catalog); err != nil {
		panic("invalid reference pack profile projection")
	}
	result := map[string]Profile{}
	for _, profile := range catalog.Profiles {
		result[profile.Key] = profile
	}
	return result
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
func packTime(value string) (time.Time, bool) {
	t, err := time.Parse("2006-01-02T15:04:05Z", value)
	return t, err == nil && t.Year() >= 1 && t.Format("2006-01-02T15:04:05Z") == value
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

// PayloadDigest hashes the already validated manifest order, including notices.
// It never sorts or repairs untrusted declarations.
func PayloadDigest(files []File) (string, error) {
	if len(files) < 1 || len(files) > 66 {
		return "", fail("manifest_schema_invalid")
	}
	h := sha256.New()
	h.Write([]byte("cartulary.reference_pack.payload.v1\x00"))
	previous := ""
	for _, file := range files {
		digest, ok := hexBytes(file.SHA256, 32)
		if !ok || file.Size < 0 || file.Size > 268435456 || ValidatePath(file.Path, false) != nil || file.Path <= previous {
			return "", fail("manifest_schema_invalid")
		}
		previous = file.Path
		var pathLength [4]byte
		binary.BigEndian.PutUint32(pathLength[:], uint32(len(file.Path)))
		h.Write(pathLength[:])
		h.Write([]byte(file.Path))
		var size [8]byte
		binary.BigEndian.PutUint64(size[:], uint64(file.Size))
		h.Write(size[:])
		h.Write(digest)
	}
	return hex.EncodeToString(h.Sum(nil)), nil
}

func ValidateNotice(data []byte) error {
	if len(data) == 0 || !utf8.Valid(data) || bytes.HasPrefix(data, []byte{0xef, 0xbb, 0xbf}) || !norm.NFC.IsNormal(data) || data[len(data)-1] != '\n' || (len(data) > 1 && data[len(data)-2] == '\n') {
		return fail("content_semantic_invalid")
	}
	for _, r := range string(data) {
		if unicode.IsControl(r) && r != '\n' && r != '\t' {
			return fail("content_semantic_invalid")
		}
	}
	return nil
}
