package rollbackprovider

import (
	"errors"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/rollbackcontract"
)

func TestParseObservationTargetTreatsMissingTombstonesAsActive(t *testing.T) {
	t.Parallel()
	observationID, incidentID, sourceID := uuid.New(), uuid.New(), uuid.New()
	value := map[string]any{
		"indicator_observation_id": observationID.String(), "incident_id": incidentID.String(),
		"source_record_id": sourceID.String(), "source_field_key": "timeline.raw_activity_text",
		"origin_kind": "manual_entry", "origin_locator": "cell", "observed_text": "192.0.2.10",
		"resolution_status": "unresolved", "row_version": float64(1),
		"created_by_user_id": uuid.New().String(), "created_at": "2026-07-09T12:00:00Z",
	}
	identity, _, _, err := parseChildTarget(rollbackcontract.NonRowTarget{
		TargetKind: "indicator_observation", TargetID: observationID.String(), OperationKind: "create", AfterValue: value,
	})
	if err != nil || identity.deletedAt != nil || identity.deletedBy != nil {
		t.Fatalf("parse old observation value = %#v, %v", identity, err)
	}
}

func TestParseObservationResolveRejectsIdentityDrift(t *testing.T) {
	t.Parallel()
	observationID, incidentID, sourceID := uuid.New(), uuid.New(), uuid.New()
	before := map[string]any{
		"indicator_observation_id": observationID.String(), "incident_id": incidentID.String(),
		"source_record_id": sourceID.String(), "source_field_key": "timeline.raw_activity_text",
		"origin_kind": "manual_entry", "origin_locator": "cell", "observed_text": "192.0.2.10",
		"resolution_status": "unresolved", "row_version": float64(1),
		"created_by_user_id": uuid.New().String(), "created_at": "2026-07-09T12:00:00Z",
	}
	after := make(map[string]any, len(before))
	for key, value := range before {
		after[key] = value
	}
	after["source_record_id"] = uuid.New().String()
	after["row_version"] = float64(2)
	if _, _, _, err := parseChildTarget(rollbackcontract.NonRowTarget{
		TargetKind: "indicator_observation", TargetID: observationID.String(), OperationKind: "resolve", BeforeValue: before, AfterValue: after,
	}); !errors.Is(err, rollbackcontract.ErrTargetNotReversible) {
		t.Fatalf("identity drift error = %v", err)
	}
}

func TestChildChangedFieldsCoverSourceAndCanonicalIndicators(t *testing.T) {
	t.Parallel()
	sourceID, oldID, newID := uuid.New(), uuid.New(), uuid.New()
	identity := childIdentity{targetKind: "indicator_observation", sourceRecordID: sourceID, sourceFieldKey: "timeline.raw_activity_text"}
	changed := childChangedFieldKeys(identity, childIdentity{resolvedIDs: []uuid.UUID{oldID}}, childIdentity{resolvedIDs: []uuid.UUID{newID}})
	if len(changed[sourceID]) != 1 || changed[sourceID][0] != "timeline.raw_activity_text" || len(changed[oldID]) != 3 || len(changed[newID]) != 3 {
		t.Fatalf("changed fields = %#v", changed)
	}
}
