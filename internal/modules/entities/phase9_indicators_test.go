package entities_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	. "github.com/JochiRaider/cartulary/internal/modules/entities"
	phase4storetest "github.com/JochiRaider/cartulary/internal/testutil/phase4storetest"
)

func TestPhase9_IndicatorsCanonicalObservationLifecycle_U_9_04(t *testing.T) {
	harness := phase4storetest.StartStore(t, "phase9-u-9-04-indicators")
	store := NewStore(harness.DB)
	actor := phase4storetest.SeedLocalUserFlags(t, harness.DB, "u904@example.test", "U904 Indicators", "U904IndicatorsPass1!", false, false, true)
	incident := phase4storetest.CreateIncidentInStore(t, harness.DB, actor, "txn-phase9-u-9-04-incident", "IR-U904", "Phase 9 U-9-04")

	created, err := store.CreateIndicatorRow(context.Background(), actor, incident.ID, CreateRequest{
		ClientTxnID: "txn-phase9-u-9-04-indicator-create",
		Values: map[string]string{
			"indicator.indicator_type":   "ipv4_addr",
			"indicator.value_kind":       "atomic",
			"indicator.display_value":    "203.0.113.88",
			"indicator.normalized_value": "203.0.113.88",
			"indicator.defanged_value":   "203(.)0(.)113(.)88",
		},
	}, []byte("txn-phase9-u-9-04-indicator-create"), "req-phase9-u-9-04-indicator-create", time.Date(2026, 5, 17, 17, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("create canonical indicator: %v", err)
	}
	replayed, err := store.CreateIndicatorRow(context.Background(), actor, incident.ID, CreateRequest{
		ClientTxnID: "txn-phase9-u-9-04-indicator-dedupe",
		Values: map[string]string{
			"indicator.indicator_type": "ipv4_addr",
			"indicator.value_kind":     "atomic",
			"indicator.display_value":  "203.0.113.88",
			"indicator.stix_pattern":   "[ipv4-addr:value = '203.0.113.88']",
		},
	}, []byte("txn-phase9-u-9-04-indicator-dedupe"), "req-phase9-u-9-04-indicator-dedupe", time.Date(2026, 5, 17, 17, 1, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("dedupe canonical indicator: %v", err)
	}
	if replayed.RecordID != created.RecordID || replayed.StatusCode != 200 {
		t.Fatalf("expected same canonical indicator identity on duplicate create, got first=%#v replay=%#v", created, replayed)
	}
	requireEntityCount(t, harness, `
SELECT count(*)
  FROM records r
  JOIN indicators i ON i.incident_id = r.incident_id AND i.record_id = r.record_id
 WHERE r.record_id = $1
   AND r.record_type = 'indicator'
   AND i.display_value = '203.0.113.88'
`, created.RecordID, 1)
	requireEntityCount(t, harness, `SELECT count(*) FROM indicator_observations WHERE incident_id = $1`, incident.ID, 0)

	sourceOne := uuid.New()
	sourceTwo := uuid.New()
	phase4storetest.SeedTimelineRecord(t, harness.DB, incident.ID, actor.ID, sourceOne)
	phase4storetest.SeedTimelineRecord(t, harness.DB, incident.ID, actor.ID, sourceTwo)
	firstObserved := time.Date(2026, 5, 17, 16, 30, 0, 0, time.UTC)
	lastObserved := time.Date(2026, 5, 17, 16, 45, 0, 0, time.UTC)
	observationOne, _, err := store.CreateIndicatorObservation(context.Background(), actor, IndicatorObservationCreateParams{
		IncidentID:                incident.ID,
		SourceRecordID:            sourceOne,
		SourceFieldKey:            "timeline.activity_synopsis_text",
		OriginKind:                "interactive_cell",
		OriginLocator:             "timeline:one:summary:0-12",
		ObservedText:              "203[.]0[.]113[.]88",
		ResolvedIndicatorRecordID: &created.RecordID,
		CreatedAt:                 firstObserved,
	})
	if err != nil {
		t.Fatalf("create first observation: %v", err)
	}
	observationTwo, _, err := store.CreateIndicatorObservation(context.Background(), actor, IndicatorObservationCreateParams{
		IncidentID:                incident.ID,
		SourceRecordID:            sourceTwo,
		SourceFieldKey:            "timeline.raw_activity_text",
		OriginKind:                "interactive_cell",
		OriginLocator:             "timeline:two:source:0-12",
		ObservedText:              "203[.]0[.]113[.]88",
		ResolvedIndicatorRecordID: &created.RecordID,
		CreatedAt:                 lastObserved,
	})
	if err != nil {
		t.Fatalf("create second observation: %v", err)
	}
	if observationOne.ObservationID == observationTwo.ObservationID {
		t.Fatalf("observations collapsed into one occurrence: %#v %#v", observationOne, observationTwo)
	}
	lifecycleTime := time.Date(2026, 5, 17, 15, 0, 0, 0, time.UTC)
	interval, _, err := store.AppendIndicatorLifecycleInterval(context.Background(), actor, IndicatorLifecycleAppendParams{
		IncidentID:        incident.ID,
		IndicatorRecordID: created.RecordID,
		LifecycleState:    "active",
		ValidFrom:         lifecycleTime,
		CreatedAt:         lifecycleTime,
	})
	if err != nil {
		t.Fatalf("append lifecycle interval: %v", err)
	}
	if interval.IndicatorRecordID != created.RecordID || interval.ValidFrom.Equal(firstObserved) {
		t.Fatalf("lifecycle interval is not distinct from observation timestamps: %#v", interval)
	}

	projected := phase4storetest.LookupIndicatorProjection(t, harness.DB, created.RecordID)
	if projected.ObservationCount != 2 {
		t.Fatalf("expected observation_count=2, got %#v", projected)
	}
	if projected.FirstObservedAt == nil || !projected.FirstObservedAt.UTC().Equal(firstObserved) {
		t.Fatalf("expected first_observed_at from observations, got %#v", projected)
	}
	if projected.LastObservedAt == nil || !projected.LastObservedAt.UTC().Equal(lastObserved) {
		t.Fatalf("expected last_observed_at from observations, got %#v", projected)
	}
	if projected.LifecycleSummary == nil || *projected.LifecycleSummary != "active" {
		t.Fatalf("expected lifecycle_summary active, got %#v", projected)
	}
}

func requireEntityCount(t testing.TB, harness *phase4storetest.StoreHarness, query string, args ...any) {
	t.Helper()
	want := args[len(args)-1].(int)
	args = args[:len(args)-1]
	var got int
	if err := harness.DB.QueryRow(context.Background(), query, args...).Scan(&got); err != nil {
		t.Fatalf("query scalar count: %v", err)
	}
	if got != want {
		t.Fatalf("unexpected count for %q: got %d want %d", query, got, want)
	}
}
