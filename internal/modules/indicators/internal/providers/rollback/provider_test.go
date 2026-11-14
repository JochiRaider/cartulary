package rollback

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
)

func TestIndicatorRollbackRequiresCompleteSource(t *testing.T) {
	t.Parallel()
	t.Run("complete source successor", requireCompleteIndicatorSource)
	recordID := uuid.MustParse("00000000-0000-4000-8000-000000000001")
	incidentID := uuid.MustParse("00000000-0000-4000-8000-000000000002")
	assertMalformedIndicatorRollbackSources(t, recordID, incidentID)
}

func assertMalformedIndicatorRollbackSources(t *testing.T, recordID uuid.UUID, incidentID uuid.UUID) {
	t.Helper()
	canonical := validIndicatorRollbackSnapshot(recordID, incidentID)
	if err := (Provider{}).ValidateRollbackValue(canonical); err != nil {
		t.Fatalf("valid baseline: %v", err)
	}

	if err := (Provider{}).RestoreTx(context.Background(), nil, rollbackcontract.RestoreRequest{RecordID: uuid.New(), RetainedValue: canonical}); !errors.Is(err, rollbackcontract.ErrTargetNotReversible) {
		t.Fatalf("request identity mismatch reached DB: %v", err)
	}

	cases := map[string]map[string]any{
		"missing source":        {"record": canonical["record"]},
		"null source":           {"record": canonical["record"], "source": nil},
		"wrong source type":     {"record": canonical["record"], "source": "indicator"},
		"empty source":          {"record": canonical["record"], "source": map[string]any{}},
		"unknown source member": indicatorRollbackSnapshotWith(canonical, "row_version", float64(1)),
	}
	for _, field := range []string{"record_id", "incident_id", "indicator_type", "value_kind", "display_value", "dedupe_key", "normalized_value"} {
		cases[field+" null"] = indicatorRollbackSnapshotWith(canonical, field, nil)
		cases[field+" wrong type"] = indicatorRollbackSnapshotWith(canonical, field, float64(1))
		cases[field+" empty"] = indicatorRollbackSnapshotWith(canonical, field, "")
		cases[field+" blank"] = indicatorRollbackSnapshotWith(canonical, field, " \t")
		cases[field+" NUL"] = indicatorRollbackSnapshotWith(canonical, field, "bad\x00value")
	}
	for _, field := range []string{"defanged_value", "hash_algorithm", "hash_value", "stix_pattern"} {
		cases[field+" wrong type"] = indicatorRollbackSnapshotWith(canonical, field, float64(1))
		if field == "hash_algorithm" || field == "hash_value" {
			cases[field+" empty"] = indicatorRollbackSnapshotWith(canonical, field, "")
			cases[field+" blank"] = indicatorRollbackSnapshotWith(canonical, field, " \t")
		}
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
	return map[string]any{"record": map[string]any{"record_id": recordID.String(), "incident_id": incidentID.String()}, "source": map[string]any{
		"record_id": recordID.String(), "incident_id": incidentID.String(),
		"indicator_type": "domain_name", "value_kind": "atomic",
		"display_value": "before.example", "normalized_value": "before.example",
		"dedupe_key": "domain_name:before.example", "defanged_value": "before[.]example",
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
	return map[string]any{"record": base["record"], "source": cloned}
}

func stringPointer(value string) *string { return &value }

func TestParseChildValueRejectsNonRegistryObservationOrigin(t *testing.T) {
	t.Parallel()
	value := validRollbackObservationValue()
	if _, err := parseChildValue("indicator_observation", value); err != nil {
		t.Fatalf("valid Observation origin control: %v", err)
	}
	value["origin_kind"] = "interactive_cell"

	if _, err := parseChildValue("indicator_observation", value); !errors.Is(err, rollbackcontract.ErrTargetNotReversible) {
		t.Fatalf("parse invalid historical origin = %v", err)
	}
}
