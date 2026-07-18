package indicators_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	"github.com/JochiRaider/cartulary/internal/modules/records/testsupport/golden"
	"github.com/JochiRaider/cartulary/internal/modules/records/testsupport/storetest"
)

// U-4-07 / REQ-02-027, REQ-02-056..REQ-02-057, REQ-02-072..REQ-02-082 / AC-017, AC-077..AC-079.
func TestIndicatorObservationSeparation_Unit(t *testing.T) {
	harness := storetest.StartStore(t, "phase4-u-4-07-indicators")
	store := indicators.NewStore(harness.DB)
	actor := storetest.SeedLocalUserFlags(t, harness.DB, "u407@example.test", "U407", "U407Phase4Pass1!", false, false, true)
	incident := storetest.CreateIncidentInStore(t, harness.DB, actor, "txn-phase4-u-4-07-incident", "IR-U407", "Record relationships indicators")

	create := func(clientTxnID string) indicators.MutationResult {
		t.Helper()
		result, err := store.CreateIndicatorRow(context.Background(), actor, incident.ID, indicators.CreateRequest{
			ClientTxnID: clientTxnID,
			Values: map[string]string{
				"indicator.indicator_type":   golden.RecordIndicatorExamples[0].IndicatorType,
				"indicator.value_kind":       golden.RecordIndicatorExamples[0].ValueKind,
				"indicator.display_value":    golden.RecordIndicatorExamples[0].DisplayValue,
				"indicator.normalized_value": golden.RecordIndicatorExamples[0].NormalizedValue,
			},
		}, []byte(clientTxnID), "req-"+clientTxnID, golden.RecordBaseTime)
		if err != nil {
			t.Fatalf("create indicator: %v", err)
		}
		return result
	}
	first := create("txn-phase4-u-4-07-first")
	second := create("txn-phase4-u-4-07-second")
	if first.RecordID != second.RecordID || second.StatusCode != 200 {
		t.Fatalf("canonical indicator dedupe failed: first=%#v second=%#v", first, second)
	}

	storetest.SeedTimelineRecord(t, harness.DB, incident.ID, actor.ID, golden.RecordTimelineRecordID)
	storetest.SeedTimelineRecord(t, harness.DB, incident.ID, actor.ID, golden.RecordTimelineSiblingRecordID)
	for index, sourceRecordID := range []struct {
		id      uuid.UUID
		field   string
		created time.Time
	}{
		{id: golden.RecordTimelineRecordID, field: golden.RecordFieldTimelineSourceText, created: golden.RecordPastTime},
		{id: golden.RecordTimelineSiblingRecordID, field: golden.RecordFieldTimelineSummary, created: golden.RecordBaseTime},
	} {
		observation, _, err := store.CreateIndicatorObservation(context.Background(), actor, indicators.IndicatorObservationCreateParams{
			IncidentID:                incident.ID,
			SourceRecordID:            sourceRecordID.id,
			SourceFieldKey:            sourceRecordID.field,
			OriginKind:                "interactive_cell",
			OriginLocator:             "phase4-u-4-07-observation-" + string(rune('1'+index)),
			ObservedText:              golden.RecordIndicatorExamples[0].DefangedValue,
			ResolvedIndicatorRecordID: &first.RecordID,
			CreatedAt:                 sourceRecordID.created,
		})
		if err != nil || observation.ObservationID == first.RecordID {
			t.Fatalf("create source-bound observation %d: %#v %v", index, observation, err)
		}
	}
	if _, _, err := store.AppendIndicatorLifecycleInterval(context.Background(), actor, indicators.IndicatorLifecycleAppendParams{
		IncidentID:        incident.ID,
		IndicatorRecordID: first.RecordID,
		LifecycleState:    "active",
		ValidFrom:         golden.RecordPastTime,
		CreatedAt:         golden.RecordPastTime,
	}); err != nil {
		t.Fatalf("append lifecycle interval: %v", err)
	}
	projection := storetest.LookupIndicatorProjection(t, harness.DB, first.RecordID)
	if projection.ObservationCount != 2 || projection.FirstObservedAt == nil || projection.LastObservedAt == nil || projection.LifecycleSummary == nil || *projection.LifecycleSummary != "active" {
		t.Fatalf("observations did not remain distinct in projection: %#v", projection)
	}
}
