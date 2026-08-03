package indicators_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"

	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	indicatortest "github.com/JochiRaider/cartulary/internal/modules/indicators/testsupport"
	timelinetest "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/revisionsupport"
)

// indicator-storage / REQ-02-027, REQ-02-056..REQ-02-057, REQ-02-072..REQ-02-082 / AC-017, AC-077..AC-079.
func TestIndicatorObservationSeparation_Unit(t *testing.T) {
	harness := appsupport.StartStore(t, "entity_linking-u-4-07-indicators")
	store := newIndicatorTestStore(t, harness.DB, revisionsupport.MustAppender(t))
	actor := authstoretest.SeedLocalUserRecord(t, harness.DB, "u407@example.test", "U407", "U407EntityLinkingPass1!", false, false, true)
	incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-entity_linking-u-4-07-incident", "IR-U407", "Record relationships indicators")

	create := func(clientTxnID string) indicators.CreateResult {
		t.Helper()
		result, err := store.CreateIndicatorRow(context.Background(), actor, incident.ID, indicators.CreateCommand{
			ClientTxnID:   clientTxnID,
			IndicatorType: indicatortest.Examples[0].IndicatorType,
			ValueKind:     indicatortest.Examples[0].ValueKind,
			DisplayValue:  indicatortest.Examples[0].DisplayValue,
		}, []byte(clientTxnID), "req-"+clientTxnID, indicatortest.BaseTime)
		if err != nil {
			t.Fatalf("create indicator: %v", err)
		}
		return result
	}
	first := create("txn-entity_linking-u-4-07-first")
	second := create("txn-entity_linking-u-4-07-second")
	if first.RecordID != second.RecordID || second.Outcome != indicators.CreateOutcomeReused {
		t.Fatalf("canonical indicator dedupe failed: first=%#v second=%#v", first, second)
	}

	timelinetest.SeedTimelineRecord(t, harness.DB, incident.ID, actor.ID, timelinetest.RecordID)
	timelinetest.SeedTimelineRecord(t, harness.DB, incident.ID, actor.ID, timelinetest.SiblingRecordID)
	for index, sourceRecordID := range []struct {
		id    uuid.UUID
		field string
	}{
		{id: timelinetest.RecordID, field: timelinetest.FieldSourceText},
		{id: timelinetest.SiblingRecordID, field: timelinetest.FieldSummary},
	} {
		result, err := store.CreateIndicatorObservation(context.Background(), actor, manualObservationParams(
			incident.ID, sourceRecordID.id, sourceRecordID.field, &first.RecordID,
			"txn-entity-linking-observation-"+string(rune('1'+index)),
		))
		if err != nil || result.Observation.ObservationID == first.RecordID {
			t.Fatalf("create source-bound observation %d: %#v %v", index, result, err)
		}
	}
	if _, err := store.AppendIndicatorLifecycleInterval(context.Background(), actor, lifecycleAppendParams(
		incident.ID, first.RecordID, 3, indicatortest.PastTime, "txn-entity-linking-lifecycle",
	)); err != nil {
		t.Fatalf("append lifecycle interval: %v", err)
	}
	projection := lookupIndicatorProjection(t, harness.DB, first.RecordID)
	if projection.ObservationCount != 2 || projection.FirstObservedAt == nil || projection.LastObservedAt == nil || projection.LifecycleSummary == nil || *projection.LifecycleSummary != "active" {
		t.Fatalf("observations did not remain distinct in projection: %#v", projection)
	}
}
