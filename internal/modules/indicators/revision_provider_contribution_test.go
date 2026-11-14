package indicators_test

import (
	"encoding/json"
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	referencefixture "github.com/JochiRaider/cartulary/internal/modules/reference_data/testsupport"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
)

func TestRevisionProviderContributionOwnsIndicatorHistoryFacet_Unit(t *testing.T) {
	t.Run("complete Observation history successor", requireCompleteObservationHistory)
	contribution := indicators.NewRevisionContribution(referencefixture.IndicatorRegistry{})
	if contribution.SourceOwnerModule != revisions.SourceOwnerIndicators || len(contribution.Records) != 1 || len(contribution.NonRowTargets) != 2 {
		t.Fatalf("unexpected Indicators revision contribution: %#v", contribution)
	}
	record := contribution.Records[0]
	if record.RecordType != "indicator" || record.SnapshotSchemaID != "cartulary.revisions.snapshot.indicator.v1" ||
		!slices.Equal(record.HistoryTargetKinds, []string{"indicator"}) {
		t.Fatalf("unexpected Indicator record semantics: %#v", record)
	}

	sourceID := uuid.MustParse("11111111-1111-4111-8111-111111111111")
	indicatorID := uuid.MustParse("22222222-2222-4222-8222-222222222222")
	for _, target := range contribution.NonRowTargets {
		t.Run(target.TargetKind+" references", func(t *testing.T) { requireRetainedHistoryReferences(t, target) })
		if target.RollbackProvider == nil {
			t.Fatalf("target %q has incomplete semantics", target.TargetKind)
		}
		var mutation revisions.StoredMutation
		var want []uuid.UUID
		switch target.TargetKind {
		case "indicator_observation":
			mutation = revisions.StoredMutation{
				TargetKind: "indicator_observation",
				TargetID:   uuid.NewString(),
				BeforeValue: map[string]any{
					"source_record_id":             sourceID.String(),
					"resolved_indicator_record_id": indicatorID.String(),
				},
				AfterValue: map[string]any{
					"source_record_id": sourceID.String(),
				},
			}
			want = []uuid.UUID{sourceID, indicatorID}
		case "indicator_state_interval":
			t.Run("complete interval history", func(t *testing.T) { requireCompleteIntervalHistory(t, target) })
			mutation = revisions.StoredMutation{
				TargetKind: "indicator_state_interval",
				TargetID:   uuid.NewString(),
				AfterValue: map[string]any{"indicator_record_id": indicatorID.String()},
			}
			want = []uuid.UUID{indicatorID}
		default:
			t.Fatalf("unexpected Indicators target %q", target.TargetKind)
		}
		description, err := target.HistoryFacet.DescribeMutation(mutation)
		if err != nil {
			t.Fatalf("describe %s: %v", target.TargetKind, err)
		}
		if !slices.Equal(description.HistoryRecordIDs, want) || !slices.Equal(description.HistoryEntryRecordIDs, want) {
			t.Fatalf("unexpected %s associations: records=%v entries=%v want=%v", target.TargetKind, description.HistoryRecordIDs, description.HistoryEntryRecordIDs, want)
		}
	}
}

// This checks only reference projection, independently of full value admission.
func requireRetainedHistoryReferences(t *testing.T, target revisions.NonRowProviderContribution) {
	t.Helper()
	if target.HistoryReferences == nil {
		t.Fatal("reference projector missing")
	}
	firstIncident := uuid.MustParse("aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa")
	secondIncident := uuid.MustParse("bbbbbbbb-bbbb-4bbb-8bbb-bbbbbbbbbbbb")
	firstRecord := uuid.MustParse("cccccccc-cccc-4ccc-8ccc-cccccccccccc")
	secondRecord := uuid.MustParse("dddddddd-dddd-4ddd-8ddd-dddddddddddd")
	thirdRecord := uuid.MustParse("eeeeeeee-eeee-4eee-8eee-eeeeeeeeeeee")
	required, nullable, explicitIncident := "", "", true
	switch target.TargetKind {
	case "indicator_observation":
		required, nullable = "source_record_id", "resolved_indicator_record_id"
	case "indicator_state_interval":
		required = "indicator_record_id"
	default:
		t.Fatalf("unexpected target %s", target.TargetKind)
	}
	makeMutation := func() revisions.StoredMutation {
		before := map[string]any{required: firstRecord.String()}
		after := map[string]any{required: secondRecord.String()}
		if explicitIncident {
			before["incident_id"], after["incident_id"] = firstIncident.String(), secondIncident.String()
		}
		if nullable != "" {
			before[nullable], after[nullable] = secondRecord.String(), firstRecord.String()
		}
		if target.TargetKind == "indicator_state_interval" {
			before["support_refs"] = []any{secondRecord.String(), thirdRecord.String()}
			after["support_refs"] = []string{thirdRecord.String(), firstRecord.String()}
		}
		return revisions.StoredMutation{TargetKind: target.TargetKind, BeforeValue: before, AfterValue: after}
	}
	scopeAfter := secondIncident
	if !explicitIncident {
		scopeAfter = firstIncident
	}
	want := []revisions.HistoryRecordReference{{IncidentID: firstIncident, RecordID: firstRecord}}
	if nullable != "" {
		want = append(want, revisions.HistoryRecordReference{IncidentID: firstIncident, RecordID: secondRecord})
	}
	if target.TargetKind == "indicator_state_interval" {
		want = append(want, revisions.HistoryRecordReference{IncidentID: firstIncident, RecordID: secondRecord}, revisions.HistoryRecordReference{IncidentID: firstIncident, RecordID: thirdRecord})
	}
	want = append(want, revisions.HistoryRecordReference{IncidentID: scopeAfter, RecordID: secondRecord})
	if nullable != "" {
		want = append(want, revisions.HistoryRecordReference{IncidentID: scopeAfter, RecordID: firstRecord})
	}
	if target.TargetKind == "indicator_state_interval" {
		want = append(want, revisions.HistoryRecordReference{IncidentID: scopeAfter, RecordID: thirdRecord}, revisions.HistoryRecordReference{IncidentID: scopeAfter, RecordID: firstRecord})
	}
	input := makeMutation()
	unchanged := func(t *testing.T, before []byte, input revisions.StoredMutation) {
		t.Helper()
		after, err := json.Marshal(input)
		if err != nil || string(before) != string(after) {
			t.Fatal("reference projection changed retained values")
		}
	}
	bytes, err := json.Marshal(input)
	if err != nil {
		t.Fatal(err)
	}
	got, err := target.HistoryReferences(input, firstIncident)
	if err != nil || !slices.Equal(got, want) {
		t.Fatalf("references=%v want=%v error=%v", got, want, err)
	}
	unchanged(t, bytes, input)
	if nullable != "" {
		input.BeforeValue[nullable], input.AfterValue[nullable] = nil, nil
		got, err := target.HistoryReferences(input, firstIncident)
		expected := []revisions.HistoryRecordReference{{IncidentID: firstIncident, RecordID: firstRecord}, {IncidentID: scopeAfter, RecordID: secondRecord}}
		if err != nil || !slices.Equal(got, expected) {
			t.Fatalf("nullable references=%v error=%v", got, err)
		}
	}
	if target.TargetKind == "indicator_state_interval" {
		for _, empty := range []any{[]any{}, []string{}} {
			input := makeMutation()
			input.BeforeValue["support_refs"], input.AfterValue["support_refs"] = empty, empty
			got, err := target.HistoryReferences(input, firstIncident)
			expected := []revisions.HistoryRecordReference{{IncidentID: firstIncident, RecordID: firstRecord}, {IncidentID: scopeAfter, RecordID: secondRecord}}
			if err != nil || !slices.Equal(got, expected) {
				t.Fatalf("empty support references=%v error=%v", got, err)
			}
		}
		for _, side := range []string{"before", "after"} {
			for name, defect := range map[string]any{
				"missing": nil, "null": nil, "string": firstRecord.String(), "object": map[string]any{},
				"nil_any": []any(nil), "nil_strings": []string(nil), "null_element": []any{nil}, "wrong_element": []any{7},
				"zero": []string{uuid.Nil.String()}, "uppercase": []string{strings.ToUpper(firstRecord.String())},
				"compact": []string{strings.ReplaceAll(firstRecord.String(), "-", "")}, "whitespace": []string{" " + firstRecord.String()},
				"empty": []string{""}, "duplicate": []string{firstRecord.String(), firstRecord.String()},
			} {
				t.Run(side+"/support_refs/"+name, func(t *testing.T) {
					input := makeMutation()
					value := input.BeforeValue
					if side == "after" {
						value = input.AfterValue
					}
					value["support_refs"] = defect
					if name == "missing" {
						delete(value, "support_refs")
					}
					before, err := json.Marshal(input)
					if err != nil {
						t.Fatal(err)
					}
					if _, err := target.HistoryReferences(input, firstIncident); err == nil {
						t.Fatal("invalid support references accepted")
					}
					unchanged(t, before, input)
				})
			}
		}
	}
	fields := []string{required}
	if explicitIncident {
		fields = append(fields, "incident_id")
	}
	if nullable != "" {
		fields = append(fields, nullable)
	}
	for _, side := range []string{"before", "after"} {
		for _, field := range fields {
			for _, defect := range []string{"missing", "null", "type", "zero", "upper", "compact", "whitespace", "empty"} {
				if field == nullable && defect == "null" {
					continue
				}
				t.Run(side+"/"+field+"/"+defect, func(t *testing.T) {
					input := makeMutation()
					value := input.BeforeValue
					if side == "after" {
						value = input.AfterValue
					}
					switch defect {
					case "missing":
						delete(value, field)
					case "null":
						value[field] = nil
					case "type":
						value[field] = 7
					case "zero":
						value[field] = uuid.Nil.String()
					case "upper":
						value[field] = strings.ToUpper(firstRecord.String())
					case "compact":
						value[field] = strings.ReplaceAll(firstRecord.String(), "-", "")
					case "whitespace":
						value[field] = " " + firstRecord.String()
					case "empty":
						value[field] = ""
					}
					bytes, err := json.Marshal(input)
					if err != nil {
						t.Fatal(err)
					}
					if _, err := target.HistoryReferences(input, firstIncident); err == nil {
						t.Fatal("invalid reference accepted")
					}
					unchanged(t, bytes, input)
				})
			}
		}
	}
	if got, err := target.HistoryReferences(revisions.StoredMutation{TargetKind: target.TargetKind}, firstIncident); err != nil || len(got) != 0 {
		t.Fatalf("null sides=%v error=%v", got, err)
	}
}
