package assessments_test

import (
	"testing"

	assessmentpolicy "github.com/JochiRaider/cartulary/internal/modules/assessments/internal/policy"
)

func TestAssessmentPolicyCanonicalVocabulary_Unit(t *testing.T) {
	t.Parallel()
	for _, subjectType := range []string{"host", "identity"} {
		if !assessmentpolicy.ValidSubjectType(subjectType) {
			t.Fatalf("canonical subject type %q rejected", subjectType)
		}
	}
	for _, subjectType := range []string{"", "artifact", "Host"} {
		if assessmentpolicy.ValidSubjectType(subjectType) {
			t.Fatalf("invalid subject type %q accepted", subjectType)
		}
	}
	for _, state := range []string{"unknown", "suspected", "confirmed", "disproven", "cleared"} {
		if !assessmentpolicy.ValidState(state) {
			t.Fatalf("canonical state %q rejected", state)
		}
	}
	for _, state := range []string{"", "contained", "Confirmed"} {
		if assessmentpolicy.ValidState(state) {
			t.Fatalf("invalid state %q accepted", state)
		}
	}
	for _, test := range []struct {
		score *int
		band  string
		valid bool
	}{
		{score: nil, band: "unset", valid: true},
		{score: policyInt(-1), valid: false},
		{score: policyInt(0), band: "low", valid: true},
		{score: policyInt(39), band: "low", valid: true},
		{score: policyInt(40), band: "medium", valid: true},
		{score: policyInt(69), band: "medium", valid: true},
		{score: policyInt(70), band: "high", valid: true},
		{score: policyInt(100), band: "high", valid: true},
		{score: policyInt(101), valid: false},
	} {
		band, valid := assessmentpolicy.ConfidenceBand(test.score)
		if band != test.band || valid != test.valid {
			t.Fatalf("ConfidenceBand(%v) = %q, %v; want %q, %v", test.score, band, valid, test.band, test.valid)
		}
	}
}

func policyInt(value int) *int { return &value }
