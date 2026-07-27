package reference_data

import "testing"

func TestIncidentBundleReferenceCatalogValidation_Unit(t *testing.T) {
	valid := []byte(`[{
		"pack_key":"baseline",
		"pack_version":"2026.07",
		"manifest_sha256":"aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		"payload_sha256":"bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb",
		"pack_contract_version":"1",
		"content_profile_id":"baseline",
		"content_profile_version":"1",
		"distribution_kind":"builtin",
		"verification_method":"sha256",
		"source_profile_id":"reference_pack",
		"source_profile_sha256":"cccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccccc"
	}]`)
	if err := ValidateIncidentBundleReferences(valid); err != nil {
		t.Fatalf("valid reference catalog: %v", err)
	}

	for _, test := range []struct {
		name          string
		payload       string
		wantInvariant string
	}{
		{name: "not an array", payload: `{}`, wantInvariant: IncidentBundleReferenceExactShapeInvariant},
		{name: "unknown member", payload: `[{"pack_key":"baseline","unknown":"value"}]`, wantInvariant: IncidentBundleReferenceExactShapeInvariant},
		{name: "empty identity", payload: `[{"pack_key":"","pack_version":"1","manifest_sha256":"a","payload_sha256":"b","pack_contract_version":"1","content_profile_id":"p","content_profile_version":"1","distribution_kind":"builtin","verification_method":"sha256","source_profile_id":"reference_pack","source_profile_sha256":"c"}]`, wantInvariant: IncidentBundleReferenceExactShapeInvariant},
		{name: "duplicate identity", payload: `[
			{"pack_key":"baseline","pack_version":"1","manifest_sha256":"a","payload_sha256":"b","pack_contract_version":"1","content_profile_id":"p","content_profile_version":"1","distribution_kind":"builtin","verification_method":"sha256","source_profile_id":"reference_pack","source_profile_sha256":"c"},
			{"pack_key":"baseline","pack_version":"1","manifest_sha256":"d","payload_sha256":"e","pack_contract_version":"1","content_profile_id":"p","content_profile_version":"1","distribution_kind":"builtin","verification_method":"sha256","source_profile_id":"reference_pack","source_profile_sha256":"f"}
		]`, wantInvariant: IncidentBundleReferenceIdentityInvariant},
		{name: "trailing document", payload: `[] []`, wantInvariant: IncidentBundleReferenceExactShapeInvariant},
	} {
		t.Run(test.name, func(t *testing.T) {
			err := ValidateIncidentBundleReferences([]byte(test.payload))
			invariantID, ok := IncidentBundleReferenceInvariant(err)
			if !ok || invariantID != test.wantInvariant {
				t.Fatalf("validation error = %v, %q, %t; want %q", err, invariantID, ok, test.wantInvariant)
			}
		})
	}
}
