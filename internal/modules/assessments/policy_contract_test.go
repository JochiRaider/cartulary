package assessments_test

import (
	"encoding/json"
	"maps"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/assessments"

	assessmentpolicy "github.com/JochiRaider/cartulary/internal/modules/assessments/internal/policy"
)

func TestAssessmentPolicyCanonicalVocabulary_Unit(t *testing.T) {
	t.Parallel()
	t.Run("historical source admission", func(t *testing.T) {
		validate := assessments.RevisionProviderContribution().Records[0].ValidateSnapshotSource
		record := map[string]any{"record_id": "11111111-1111-4111-8111-111111111111", "incident_id": "22222222-2222-4222-8222-222222222222"}
		base := maps.Clone(record)
		base["subject_record_id"], base["subject_type"] = "33333333-3333-4333-8333-333333333333", "host"
		base["assessment_state"], base["confidence_score"], base["rationale"] = "suspected", nil, "Evidence\n\tReason"
		base["assessor_user_id"] = "44444444-4444-4444-8444-444444444444"
		base["assessed_at"], base["created_at"], base["updated_at"] = "2000-01-01T00:00:00Z", "2026-01-01T00:00:00Z", "2026-01-03T00:00:00+00:00"
		base["deleted_at"], base["deleted_by_user_id"] = nil, nil
		if validate == nil {
			t.Fatal("snapshot validator absent")
		}
		for _, subject := range []string{"host", "identity"} {
			for _, state := range []string{"unknown", "suspected", "confirmed", "disproven", "cleared"} {
				for _, score := range []any{nil, json.Number("0"), json.Number("55"), json.Number("100")} {
					value := maps.Clone(base)
					value["subject_type"], value["assessment_state"], value["confidence_score"] = subject, state, score
					before, err := json.Marshal(value)
					if err != nil {
						t.Fatal(err)
					}
					if err := validate(record, value); err != nil {
						t.Fatalf("valid %s/%s/%v rejected: %v", subject, state, score, err)
					}
					after, err := json.Marshal(value)
					if err != nil || string(after) != string(before) {
						t.Fatal("valid retained source changed")
					}
				}
			}
		}
		deleted := maps.Clone(base)
		deleted["deleted_at"], deleted["deleted_by_user_id"] = "2026-01-02T00:00:00Z", base["assessor_user_id"]
		if err := validate(record, deleted); err != nil {
			t.Fatalf("valid deletion tuple rejected: %v", err)
		}
		future := maps.Clone(base)
		future["assessed_at"] = "2030-01-01T00:00:00Z"
		if err := validate(record, future); err != nil {
			t.Fatalf("explicit assessment instant rejected: %v", err)
		}
		for _, test := range []struct {
			name   string
			mutate func(map[string]any)
		}{
			{"subject type", func(v map[string]any) { v["subject_type"] = "artifact" }},
			{"self subject", func(v map[string]any) { v["subject_record_id"] = v["record_id"] }},
			{"state", func(v map[string]any) { v["assessment_state"] = "contained" }},
			{"negative confidence", func(v map[string]any) { v["confidence_score"] = json.Number("-1") }},
			{"large confidence", func(v map[string]any) { v["confidence_score"] = json.Number("9223372036854775807") }},
			{"fractional confidence", func(v map[string]any) { v["confidence_score"] = json.Number("0.5") }},
			{"empty rationale", func(v map[string]any) { v["rationale"] = "" }},
			{"rationale trim", func(v map[string]any) { v["rationale"] = " Evidence " }},
			{"rationale NFC", func(v map[string]any) { v["rationale"] = "Cafe\u0301" }},
			{"rationale newline", func(v map[string]any) { v["rationale"] = "Evidence\r\nReason" }},
			{"rationale NUL", func(v map[string]any) { v["rationale"] = "bad\x00" }},
			{"reversed clocks", func(v map[string]any) { v["updated_at"] = "2000-01-01T00:00:00Z" }},
			{"deletion without actor", func(v map[string]any) { v["deleted_at"] = "2026-01-02T00:00:00Z" }},
			{"actor without deletion", func(v map[string]any) { v["deleted_by_user_id"] = v["assessor_user_id"] }},
			{"deletion before creation", func(v map[string]any) {
				v["deleted_at"] = "2000-01-01T00:00:00Z"
				v["deleted_by_user_id"] = v["assessor_user_id"]
			}},
			{"deletion after update", func(v map[string]any) {
				v["deleted_at"] = "2030-01-01T00:00:00Z"
				v["deleted_by_user_id"] = v["assessor_user_id"]
			}},
		} {
			t.Run(test.name, func(t *testing.T) {
				value := maps.Clone(base)
				test.mutate(value)
				before, err := json.Marshal(value)
				if err != nil {
					t.Fatal(err)
				}
				if err := validate(record, value); err == nil {
					t.Fatal("invalid historical assessment accepted")
				}
				after, err := json.Marshal(value)
				if err != nil || string(after) != string(before) {
					t.Fatal("invalid history repaired")
				}
			})
		}
	})
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
