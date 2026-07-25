package reference_data

import "testing"

func TestReferencePackStorageReferencesAreStrictAndRootFree_Unit(t *testing.T) {
	valid := []string{
		"reference-packs/imports/a.bundle",
		"reference-packs/bundles/0123456789abcdef.bundle",
	}
	for _, raw := range valid {
		if reference, err := ParseStorageRef(raw); err != nil || reference.String() != raw {
			t.Fatalf("ParseStorageRef(%q) = %#v, %v", raw, reference, err)
		}
		if reference, err := ParseStagingRef(raw); err != nil || reference.String() != raw {
			t.Fatalf("ParseStagingRef(%q) = %#v, %v", raw, reference, err)
		}
	}

	invalid := []string{
		"",
		".",
		"..",
		"/absolute.bundle",
		"reference-packs/../escape.bundle",
		"reference-packs/./pack.bundle",
		"reference-packs//pack.bundle",
		`reference-packs\pack.bundle`,
		"reference-packs/pack.bundle/",
		"reference-packs/\x00pack.bundle",
		"reference-packs/e\u0301.bundle",
	}
	for _, raw := range invalid {
		if _, err := ParseStorageRef(raw); err == nil {
			t.Errorf("ParseStorageRef(%q) succeeded", raw)
		}
		if _, err := ParseStagingRef(raw); err == nil {
			t.Errorf("ParseStagingRef(%q) succeeded", raw)
		}
	}
}
