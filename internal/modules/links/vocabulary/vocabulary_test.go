package vocabulary

import "testing"

func TestLinkVocabularyExactRoundTrips(t *testing.T) {
	linkTypes := []struct {
		token string
		value LinkType
	}{
		{"observed_on_host", LinkTypeObservedOnHost},
		{"observed_as_identity", LinkTypeObservedAsIdentity},
		{"references_indicator", LinkTypeReferencesIndicator},
		{"attached_evidence", LinkTypeAttachedEvidence},
		{"references_artifact", LinkTypeReferencesArtifact},
		{"derived_from", LinkTypeDerivedFrom},
		{"merged_into", LinkTypeMergedInto},
		{"supported_by", LinkTypeSupportedBy},
		{"references_record", LinkTypeReferencesRecord},
		{"supersedes", LinkTypeSupersedes},
	}
	for _, test := range linkTypes {
		parsed, err := ParseLinkType(test.token)
		if err != nil || parsed != test.value || parsed.String() != test.token {
			t.Fatalf("link type round trip %q = (%v, %v), want %v", test.token, parsed, err, test.value)
		}
	}

	provenances := []struct {
		token string
		value LinkProvenance
	}{
		{"manual", LinkProvenanceManual},
		{"auto_match", LinkProvenanceAutoMatch},
		{"import", LinkProvenanceImport},
		{"rollback", LinkProvenanceRollback},
		{"system", LinkProvenanceSystem},
	}
	for _, test := range provenances {
		parsed, err := ParseLinkProvenance(test.token)
		if err != nil || parsed != test.value || parsed.String() != test.token {
			t.Fatalf("provenance round trip %q = (%v, %v), want %v", test.token, parsed, err, test.value)
		}
	}
}

func TestLinkVocabularyRejectsNonExactTokens(t *testing.T) {
	invalid := []string{"", " manual", "manual ", "Manual", "MANUAL", "auto-match", "references-record", " references_record", "references_record ", "References_Record", "unknown"}
	for _, token := range invalid {
		if parsed, err := ParseLinkType(token); err == nil || parsed != LinkTypeInvalid {
			t.Fatalf("ParseLinkType(%q) = (%v, %v), want invalid error", token, parsed, err)
		}
		if parsed, err := ParseLinkProvenance(token); err == nil || parsed != LinkProvenanceInvalid {
			t.Fatalf("ParseLinkProvenance(%q) = (%v, %v), want invalid error", token, parsed, err)
		}
	}
	if LinkTypeInvalid.String() != "" || LinkType(255).String() != "" || LinkProvenanceInvalid.String() != "" || LinkProvenance(255).String() != "" {
		t.Fatal("invalid vocabulary values must serialize to empty strings")
	}
}
