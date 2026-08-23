package rollback

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
)

func TestIndicatorRollbackSourcePatchIsClosedAndPresenceAware(t *testing.T) {
	t.Parallel()
	recordID := uuid.MustParse("00000000-0000-4000-8000-000000000001")
	incidentID := uuid.MustParse("00000000-0000-4000-8000-000000000002")
	value := validIndicatorRollbackSnapshot(recordID, incidentID)
	patch, err := parseIndicatorSourcePatch(value)
	if err != nil {
		t.Fatalf("parse canonical Indicator snapshot: %v", err)
	}
	state := rowState{
		recordID: recordID, incidentID: incidentID,
		indicatorType: "domain_name", valueKind: "atomic", displayValue: "current.example",
		normalized: stringPointer("current.example"), dedupeKey: "current-dedupe",
		defanged: stringPointer("current[.]example"), hashAlgorithm: stringPointer("sha256"),
		hashValue:   stringPointer("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"),
		stixPattern: stringPointer("[domain-name:value = 'current.example']"),
	}
	if !patch.overlay(&state) {
		t.Fatal("canonical Indicator snapshot did not overlay its matching row")
	}
	if state.displayValue != "before.example" || state.normalized == nil || *state.normalized != "before.example" ||
		state.defanged == nil || *state.defanged != "before[.]example" || state.hashAlgorithm != nil ||
		state.hashValue != nil || state.stixPattern == nil || *state.stixPattern != "[domain-name:value = 'before.example']" {
		t.Fatalf("canonical Indicator snapshot overlay = %#v", state)
	}

	partial, err := parseIndicatorSourcePatch(map[string]any{"source": map[string]any{
		"display_value": "partial.example", "normalized_value": "partial.example", "defanged_value": nil,
	}})
	if err != nil {
		t.Fatalf("parse valid partial Indicator patch: %v", err)
	}
	if !partial.overlay(&state) || state.displayValue != "partial.example" || state.normalized == nil || *state.normalized != "partial.example" || state.defanged != nil {
		t.Fatalf("partial Indicator patch overlay = %#v", state)
	}
	assertMalformedIndicatorRollbackSourcePatches(t, recordID, incidentID)
}

func assertMalformedIndicatorRollbackSourcePatches(t *testing.T, recordID uuid.UUID, incidentID uuid.UUID) {
	t.Helper()
	canonical := validIndicatorRollbackSnapshot(recordID, incidentID)

	cases := map[string]map[string]any{
		"missing source":        {},
		"null source":           {"source": nil},
		"wrong source type":     {"source": "indicator"},
		"empty source":          {"source": map[string]any{}},
		"unknown source member": indicatorRollbackSnapshotWith(canonical, "row_version", float64(1)),
	}
	for _, field := range []string{"record_id", "incident_id", "indicator_type", "value_kind", "display_value", "dedupe_key"} {
		cases[field+" null"] = indicatorRollbackSnapshotWith(canonical, field, nil)
		cases[field+" wrong type"] = indicatorRollbackSnapshotWith(canonical, field, float64(1))
		cases[field+" empty"] = indicatorRollbackSnapshotWith(canonical, field, "")
		cases[field+" blank"] = indicatorRollbackSnapshotWith(canonical, field, " \t")
		cases[field+" NUL"] = indicatorRollbackSnapshotWith(canonical, field, "bad\x00value")
	}
	for _, field := range []string{"normalized_value", "defanged_value", "hash_algorithm", "hash_value", "stix_pattern"} {
		cases[field+" wrong type"] = indicatorRollbackSnapshotWith(canonical, field, float64(1))
		cases[field+" empty"] = indicatorRollbackSnapshotWith(canonical, field, "")
		cases[field+" blank"] = indicatorRollbackSnapshotWith(canonical, field, " \t")
		cases[field+" NUL"] = indicatorRollbackSnapshotWith(canonical, field, "bad\x00value")
	}
	cases["invalid Indicator type"] = indicatorRollbackSnapshotWith(canonical, "indicator_type", "domain")
	cases["invalid value kind"] = indicatorRollbackSnapshotWith(canonical, "value_kind", "literal")
	cases["invalid record UUID"] = indicatorRollbackSnapshotWith(canonical, "record_id", "not-a-uuid")
	cases["nil record UUID"] = indicatorRollbackSnapshotWith(canonical, "record_id", uuid.Nil.String())
	cases["noncanonical record UUID"] = indicatorRollbackSnapshotWith(canonical, "record_id", "00000000-0000-4000-8000-0000000000AA")
	cases["invalid incident UUID"] = indicatorRollbackSnapshotWith(canonical, "incident_id", "not-a-uuid")
	cases["nil incident UUID"] = indicatorRollbackSnapshotWith(canonical, "incident_id", uuid.Nil.String())
	cases["noncanonical incident UUID"] = indicatorRollbackSnapshotWith(canonical, "incident_id", "00000000-0000-4000-8000-0000000000AA")

	for name, value := range cases {
		value := value
		t.Run(name, func(t *testing.T) {
			err := (Provider{}).RestoreTx(context.Background(), nil, rollbackcontract.RestoreRequest{RecordID: recordID, RetainedValue: value})
			if !errors.Is(err, rollbackcontract.ErrTargetNotReversible) {
				t.Fatalf("RestoreTx malformed retained history = %v", err)
			}
		})
	}
}

func validIndicatorRollbackSnapshot(recordID uuid.UUID, incidentID uuid.UUID) map[string]any {
	return map[string]any{"source": map[string]any{
		"record_id": recordID.String(), "incident_id": incidentID.String(),
		"indicator_type": "domain_name", "value_kind": "atomic",
		"display_value": "before.example", "normalized_value": "before.example",
		"dedupe_key": "before-dedupe", "defanged_value": "before[.]example",
		"hash_algorithm": nil, "hash_value": nil,
		"stix_pattern": "[domain-name:value = 'before.example']",
	}}
}

func indicatorRollbackSnapshotWith(base map[string]any, key string, value any) map[string]any {
	source := base["source"].(map[string]any)
	cloned := make(map[string]any, len(source))
	for sourceKey, sourceValue := range source {
		cloned[sourceKey] = sourceValue
	}
	cloned[key] = value
	return map[string]any{"source": cloned}
}

func stringPointer(value string) *string { return &value }

func TestParseChildValueRejectsNonRegistryObservationOrigin(t *testing.T) {
	t.Parallel()
	value := map[string]any{
		"incident_id":                  "00000000-0000-4000-8000-000000000001",
		"row_version":                  float64(1),
		"deleted_at":                   nil,
		"deleted_by_user_id":           nil,
		"indicator_observation_id":     "00000000-0000-4000-8000-000000000002",
		"source_record_id":             "00000000-0000-4000-8000-000000000003",
		"source_field_key":             "timeline.raw_activity_text",
		"resolution_status":            "unresolved",
		"origin_kind":                  "interactive_cell",
		"origin_locator":               "rollback-origin-test",
		"observed_text":                "192.0.2.10",
		"created_by_user_id":           "00000000-0000-4000-8000-000000000004",
		"created_at":                   "2026-08-03T16:00:00Z",
		"resolved_indicator_record_id": nil,
		"resolved_by_user_id":          nil,
		"resolved_at":                  nil,
		"resolution_method":            nil,
	}
	if _, err := parseChildValue("indicator_observation", value); !errors.Is(err, rollbackcontract.ErrTargetNotReversible) {
		t.Fatalf("parse invalid historical origin = %v", err)
	}
}
