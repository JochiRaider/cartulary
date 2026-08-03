package indicators_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	timelinetest "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/revisionsupport"
)

func TestIndicatorsCanonicalObservationLifecycle_Integration(t *testing.T) {
	harness := appsupport.StartStore(t, "workbook_interaction-u-9-04-indicators")
	store := newIndicatorTestStore(t, harness.DB, revisionsupport.MustAppender(t))
	actor := authstoretest.SeedLocalUserRecord(t, harness.DB, "u904@example.test", "U904 Indicators", "U904IndicatorsPass1!", false, false, true)
	incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-workbook_interaction-u-9-04-incident", "IR-U904", "Workbook inspector indicator-storage")
	defangedValue := "203(.)0(.)113(.)88"
	stixPattern := "[ipv4-addr:value = '203.0.113.88']"

	created, err := store.CreateIndicatorRow(context.Background(), actor, incident.ID, indicators.CreateCommand{
		ClientTxnID:   "txn-workbook_interaction-u-9-04-indicator-create",
		IndicatorType: "ipv4_addr",
		ValueKind:     "atomic",
		DisplayValue:  "203.0.113.88",
		DefangedValue: &defangedValue,
	}, []byte("txn-workbook_interaction-u-9-04-indicator-create"), "req-workbook_interaction-u-9-04-indicator-create", time.Date(2026, 5, 17, 17, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("create canonical indicator: %v", err)
	}
	replayed, err := store.CreateIndicatorRow(context.Background(), actor, incident.ID, indicators.CreateCommand{
		ClientTxnID:   "txn-workbook_interaction-u-9-04-indicator-dedupe",
		IndicatorType: "ipv4_addr",
		ValueKind:     "atomic",
		DisplayValue:  "203.0.113.88",
		STIXPattern:   &stixPattern,
	}, []byte("txn-workbook_interaction-u-9-04-indicator-dedupe"), "req-workbook_interaction-u-9-04-indicator-dedupe", time.Date(2026, 5, 17, 17, 1, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("dedupe canonical indicator: %v", err)
	}
	if replayed.RecordID != created.RecordID || replayed.Outcome != indicators.CreateOutcomeUpdated {
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
	timelinetest.SeedTimelineRecord(t, harness.DB, incident.ID, actor.ID, sourceOne)
	timelinetest.SeedTimelineRecord(t, harness.DB, incident.ID, actor.ID, sourceTwo)
	observationOneResult, err := store.CreateIndicatorObservation(context.Background(), actor, manualObservationParams(
		incident.ID, sourceOne, "timeline.activity_synopsis_text", &created.RecordID, "txn-indicator-observation-one",
	))
	if err != nil {
		t.Fatalf("create first observation: %v", err)
	}
	observationOne := observationOneResult.Observation
	observationTwoResult, err := store.CreateIndicatorObservation(context.Background(), actor, manualObservationParams(
		incident.ID, sourceTwo, "timeline.raw_activity_text", &created.RecordID, "txn-indicator-observation-two",
	))
	if err != nil {
		t.Fatalf("create second observation: %v", err)
	}
	observationTwo := observationTwoResult.Observation
	if observationOne.ObservationID == observationTwo.ObservationID {
		t.Fatalf("observations collapsed into one occurrence: %#v %#v", observationOne, observationTwo)
	}
	if observationOne.OriginKind != "manual_entry" || observationTwo.OriginKind != "manual_entry" {
		t.Fatalf("repeated observation origins = %q, %q; want manual_entry", observationOne.OriginKind, observationTwo.OriginKind)
	}
	lifecycleTime := time.Date(2026, 5, 17, 15, 0, 0, 0, time.UTC)
	intervalResult, err := store.AppendIndicatorLifecycleInterval(context.Background(), actor, lifecycleAppendParams(
		incident.ID, created.RecordID, 4, lifecycleTime, "txn-indicator-lifecycle",
	))
	if err != nil {
		t.Fatalf("append lifecycle interval: %v", err)
	}
	interval := intervalResult.Interval
	if interval.IndicatorRecordID != created.RecordID || interval.ValidFrom.Equal(observationOne.CreatedAt) {
		t.Fatalf("lifecycle interval is not distinct from observation timestamps: %#v", interval)
	}

	projected := lookupIndicatorProjection(t, harness.DB, created.RecordID)
	if projected.ObservationCount != 2 {
		t.Fatalf("expected observation_count=2, got %#v", projected)
	}
	if projected.FirstObservedAt == nil || !projected.FirstObservedAt.UTC().Equal(observationOne.CreatedAt.Truncate(time.Microsecond)) {
		t.Fatalf("expected first_observed_at from observations, got %#v", projected)
	}
	if projected.LastObservedAt == nil || !projected.LastObservedAt.UTC().Equal(observationTwo.CreatedAt.Truncate(time.Microsecond)) {
		t.Fatalf("expected last_observed_at from observations, got %#v", projected)
	}
	if projected.LifecycleSummary == nil || *projected.LifecycleSummary != "active" {
		t.Fatalf("expected lifecycle_summary active, got %#v", projected)
	}
}

func TestNetworkFlowCore02_IndicatorFindOrCreateParticipantRollback(t *testing.T) {
	ctx := context.Background()
	harness := appsupport.StartStore(t, "network-flow-core02-indicator-participant")
	store := newIndicatorTestStore(t, harness.DB, revisionsupport.MustAppender(t))
	actor := authstoretest.SeedLocalUserRecord(t, harness.DB, "nfc02@example.test", "Network Flow Core 02", "NFCore02Pass1!", false, false, true)
	incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-network-flow-core02-incident", "IR-NFC02", "Network Flow Core 02")

	tx, err := harness.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin participant transaction: %v", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	normalized := "2001:0db8:0:0:0:0:0:1"
	first, err := store.FindOrCreateIndicatorParticipantTx(ctx, tx, indicators.IndicatorFindOrCreateParticipantCommand{
		IncidentID:        incident.ID,
		Actor:             actor,
		IndicatorType:     "ipv6_addr",
		ValueKind:         "atomic",
		DisplayValue:      "2001:db8::1",
		NormalizedValue:   &normalized,
		OperationContext:  "network_flow_indicator_binding",
		OperationOccurred: time.Date(2026, 5, 17, 18, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("first participant create: %v", err)
	}
	if first.SchemaID != indicators.IndicatorFindOrCreateParticipantV1 || first.Status != "created" {
		t.Fatalf("unexpected first participant result: %#v", first)
	}
	if first.Indicator.DisplayValue != "2001:db8::1" || first.Indicator.NormalizedValue == nil || *first.Indicator.NormalizedValue != "2001:db8::1" {
		t.Fatalf("participant did not return canonical IPv6 identity: %#v", first.Indicator)
	}

	second, err := store.FindOrCreateIndicatorParticipantTx(ctx, tx, indicators.IndicatorFindOrCreateParticipantCommand{
		IncidentID:        incident.ID,
		Actor:             actor,
		IndicatorType:     "ipv6_addr",
		ValueKind:         "atomic",
		DisplayValue:      "2001:0db8:0000:0000:0000:0000:0000:0001",
		OperationContext:  "network_flow_indicator_binding",
		OperationOccurred: time.Date(2026, 5, 17, 18, 1, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("second participant reuse: %v", err)
	}
	if second.Status != "reused" || second.Indicator.RecordID != first.Indicator.RecordID {
		t.Fatalf("participant did not reuse canonical indicator: first=%#v second=%#v", first, second)
	}

	if err := tx.Rollback(ctx); err != nil {
		t.Fatalf("rollback participant transaction: %v", err)
	}
	requireEntityCount(t, harness, `SELECT count(*) FROM indicators WHERE incident_id = $1 AND indicator_type = 'ipv6_addr'`, incident.ID, 0)
}

func requireEntityCount(t testing.TB, harness *appsupport.StoreHarness, query string, args ...any) {
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
