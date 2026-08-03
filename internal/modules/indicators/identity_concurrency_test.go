package indicators_test

import (
	"context"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/revisionsupport"
)

func TestConcurrentIndicatorFindOrCreateConvergesOnOneActiveIdentity_Integration(t *testing.T) {
	ctx := context.Background()
	harness := appsupport.StartStore(t, "indicators-concurrent-identity")
	store := newIndicatorTestStore(t, harness.DB, revisionsupport.MustAppender(t))
	actor := authstoretest.SeedLocalUserRecord(t, harness.DB, "indicator-concurrency@example.test", "Indicator Concurrency", "IndicatorConcurrencyPass1!", false, false, true)
	incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-indicator-concurrency-incident", "IR-IND-CONCURRENCY", "Indicator identity convergence")

	firstTx, err := harness.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin first transaction: %v", err)
	}
	defer func() { _ = firstTx.Rollback(ctx) }()
	secondTx, err := harness.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin second transaction: %v", err)
	}
	defer func() { _ = secondTx.Rollback(ctx) }()

	first, err := store.FindOrCreateIndicatorParticipantTx(ctx, firstTx, indicators.IndicatorFindOrCreateParticipantCommand{
		IncidentID:        incident.ID,
		Actor:             actor,
		IndicatorType:     "ipv6_addr",
		ValueKind:         "atomic",
		DisplayValue:      "2001:db8::9",
		OperationContext:  "identity_convergence_first",
		OperationOccurred: time.Date(2026, 8, 3, 16, 0, 0, 0, time.UTC),
	})
	if err != nil {
		t.Fatalf("first find or create: %v", err)
	}
	if first.Status != "created" {
		t.Fatalf("first status = %q, want created", first.Status)
	}

	type outcome struct {
		result indicators.IndicatorFindOrCreateParticipantResult
		err    error
	}
	secondOutcome := make(chan outcome, 1)
	go func() {
		result, callErr := store.FindOrCreateIndicatorParticipantTx(ctx, secondTx, indicators.IndicatorFindOrCreateParticipantCommand{
			IncidentID:        incident.ID,
			Actor:             actor,
			IndicatorType:     "ipv6_addr",
			ValueKind:         "atomic",
			DisplayValue:      "2001:0db8:0:0:0:0:0:9",
			OperationContext:  "identity_convergence_second",
			OperationOccurred: time.Date(2026, 8, 3, 16, 0, 1, 0, time.UTC),
		})
		secondOutcome <- outcome{result: result, err: callErr}
	}()

	waitForTransactionLock(t, harness, secondTx.Conn().PgConn().PID())
	if err := firstTx.Commit(ctx); err != nil {
		t.Fatalf("commit first transaction: %v", err)
	}

	var second outcome
	select {
	case second = <-secondOutcome:
	case <-time.After(10 * time.Second):
		t.Fatal("second find or create did not resume after identity owner committed")
	}
	if second.err != nil {
		t.Fatalf("second find or create: %v", second.err)
	}
	if second.result.Status != "reused" || second.result.Indicator.RecordID != first.Indicator.RecordID {
		t.Fatalf("concurrent identity did not converge: first=%#v second=%#v", first, second.result)
	}
	if err := secondTx.Commit(ctx); err != nil {
		t.Fatalf("commit second transaction: %v", err)
	}

	requireEntityCount(t, harness, `
SELECT count(*)
  FROM indicators AS indicator
  JOIN records AS envelope
    ON envelope.incident_id = indicator.incident_id
   AND envelope.record_id = indicator.record_id
 WHERE indicator.incident_id = $1
   AND indicator.indicator_type = 'ipv6_addr'
   AND indicator.dedupe_key = $2
   AND envelope.deleted_at IS NULL
`, incident.ID, first.Indicator.DedupeKey, 1)
	requireEntityCount(t, harness, `
SELECT count(*)
  FROM indicator_active_identities
 WHERE incident_id = $1
   AND indicator_type = 'ipv6_addr'
   AND dedupe_key = $2
   AND indicator_record_id = $3
`, incident.ID, first.Indicator.DedupeKey, first.Indicator.RecordID, 1)
	requireEntityCount(t, harness, `
SELECT count(*)
  FROM records
 WHERE incident_id = $1
   AND record_type = 'indicator'
   AND deleted_at IS NULL
`, incident.ID, 1)
}

func waitForTransactionLock(t testing.TB, harness *appsupport.StoreHarness, processID uint32) {
	t.Helper()
	deadline := time.Now().Add(10 * time.Second)
	for time.Now().Before(deadline) {
		var waiting bool
		if err := harness.DB.QueryRow(context.Background(), `
SELECT COALESCE(wait_event_type = 'Lock', false)
  FROM pg_stat_activity
 WHERE pid = $1
`, processID).Scan(&waiting); err != nil {
			t.Fatalf("inspect concurrent identity waiter: %v", err)
		}
		if waiting {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatal("second find or create did not wait for the first owner transaction")
}
