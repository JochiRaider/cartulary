package packformat

import "encoding/json"

// BuiltinReleaseBinding is the exact release-authenticated identity. The
// packaging wrapper and logical member bytes are not members of this object.
type BuiltinReleaseBinding struct {
	SchemaID       string `json:"schema_id"`
	Key            string `json:"pack_key"`
	Version        string `json:"pack_version"`
	Sequence       int64  `json:"pack_release_sequence"`
	ManifestSHA256 string `json:"manifest_sha256"`
	PayloadSHA256  string `json:"payload_sha256"`
	ProfileID      string `json:"content_profile_id"`
	ProfileVersion string `json:"content_profile_version"`
}

var builtinBindingShape = compileProjection("builtin_release_binding.v1.schema.json")

func DecodeBuiltinReleaseBinding(data []byte) (BuiltinReleaseBinding, error) {
	var binding BuiltinReleaseBinding
	if _, err := strictDocument(data, builtinBindingShape, "builtin_release_binding_invalid", "builtin_release_binding_invalid"); err != nil {
		return binding, err
	}
	if err := json.Unmarshal(data, &binding); err != nil {
		return binding, err
	}
	profile, ok := profiles[binding.Key]
	if !ok || !profile.Required || binding.ProfileID != profile.ID || binding.ProfileVersion != profile.Version {
		return binding, fail("builtin_release_binding_invalid")
	}
	return binding, nil
}

func (b BuiltinReleaseBinding) Matches(m Manifest, manifestSHA, payloadSHA string) bool {
	return m.Repository == nil && b.Key == m.Key && b.Version == m.Version && b.Sequence == m.Sequence && b.ManifestSHA256 == manifestSHA && b.PayloadSHA256 == payloadSHA && b.ProfileID == m.ProfileID && b.ProfileVersion == m.ProfileVersion
}
