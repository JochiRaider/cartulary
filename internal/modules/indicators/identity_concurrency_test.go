package indicators_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/jackc/pgx/v5"

	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/revisionsupport"
)

func TestConcurrentIndicatorFindOrCreateConvergesOnOneActiveIdentity_Integration(t *testing.T) {
	ctx := context.Background()
	harness := appsupport.StartStore(t, "indicators-concurrent-identity")
	application := newIndicatorTestApplication(t, harness.DB, revisionsupport.MustAppender(t))
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

	first, err := application.FindOrCreateIndicatorParticipantTx(ctx, firstTx, indicators.IndicatorFindOrCreateParticipantCommand{
		IncidentID:        incident.ID,
		ActorUserID:       actor.ID,
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
		result, callErr := application.FindOrCreateIndicatorParticipantTx(ctx, secondTx, indicators.IndicatorFindOrCreateParticipantCommand{
			IncidentID:        incident.ID,
			ActorUserID:       actor.ID,
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

func TestConcurrentIndicatorCreateReceipts_Integration(t *testing.T) {
	for _, scenario := range []string{"same key", "distinct keys", "divergent key"} {
		t.Run(scenario, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			harness := appsupport.StartStore(t, "indicator-create-receipts")
			application := newIndicatorTestApplication(t, harness.DB, revisionsupport.MustAppender(t))
			actor := authstoretest.SeedLocalUserRecord(t, harness.DB, "create-race@example.test", "Create race", "CreateRacePass1!", false, false, true)
			incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-create-race-incident", "IR-CREATE-RACE", "Create receipts")
			// Hold identity ownership with an uncommitted participant. Rolling it back
			// releases both direct requests onto an identity that does not yet exist.
			gate, err := harness.DB.BeginTx(ctx, pgx.TxOptions{})
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = gate.Rollback(context.Background()) }()
			_, err = application.FindOrCreateIndicatorParticipantTx(ctx, gate, indicators.IndicatorFindOrCreateParticipantCommand{
				IncidentID: incident.ID, ActorUserID: actor.ID, IndicatorType: "ipv6_addr", ValueKind: "atomic", DisplayValue: "2001:db8::9",
				OperationContext: "test_identity_gate", OperationOccurred: time.Now().UTC(),
			})
			if err != nil {
				t.Fatal(err)
			}
			first := indicators.CreateCommand{ClientTxnID: "txn-race", IndicatorType: "ipv6_addr", ValueKind: "atomic", DisplayValue: "2001:db8::9"}
			second := first
			second.DisplayValue = "2001:0db8:0:0:0:0:0:9"
			if scenario == "distinct keys" {
				second.ClientTxnID = "txn-race-second"
			}
			if scenario == "divergent key" {
				value := "supplied metadata"
				second.DefangedValue = &value
			}
			type outcome struct {
				result indicators.CreateResult
				err    error
			}
			outcomes := make(chan outcome, 2)
			for _, command := range []indicators.CreateCommand{first, second} {
				go func() {
					result, callErr := application.CreateIndicatorRow(ctx, actor.ID, incident.ID, command, "req-race")
					outcomes <- outcome{result, callErr}
				}()
			}
			deadline := time.Now().Add(10 * time.Second)
			waiting := 0
			for time.Now().Before(deadline) {
				if err := harness.DB.QueryRow(ctx, `SELECT count(*) FROM pg_stat_activity WHERE datname = current_database() AND wait_event_type = 'Lock' AND query LIKE '%pg_advisory_xact_lock%'`).Scan(&waiting); err != nil {
					t.Fatal(err)
				}
				if waiting >= 2 {
					break
				}
				time.Sleep(10 * time.Millisecond)
			}
			if waiting < 2 {
				t.Fatal("direct creates did not overlap at transaction locks")
			}
			if err := gate.Rollback(ctx); err != nil {
				t.Fatal(err)
			}
			a, b := <-outcomes, <-outcomes
			commits := 1
			if scenario == "divergent key" {
				if a.err != nil {
					a, b = b, a
				}
				if a.err != nil || !errors.Is(b.err, authn.ErrClientTxnConflict) {
					t.Fatalf("divergent outcomes: %#v / %#v", a, b)
				}
			} else {
				if a.err != nil || b.err != nil {
					t.Fatalf("concurrent create errors: %v / %v", a.err, b.err)
				}
				if a.result.RecordID != b.result.RecordID || string(mustCreateRowJSON(t, a.result)) != string(mustCreateRowJSON(t, b.result)) {
					t.Fatalf("canonical results diverged: %#v / %#v", a.result, b.result)
				}
				if scenario == "same key" {
					if a.result.Replayed == b.result.Replayed || a.result.ChangeSetID != b.result.ChangeSetID {
						t.Fatalf("same key did not commit once and replay: %#v / %#v", a.result, b.result)
					}
				} else {
					commits = 2
					if a.result.Replayed || b.result.Replayed || a.result.ChangeSetID == b.result.ChangeSetID {
						t.Fatal("distinct keys lost independent receipts")
					}
					if a.result.Created {
						a, b = b, a
					}
					requireEntityCount(t, harness, `SELECT count(*) FROM change_set_mutations WHERE change_set_id = $1 AND before_value = after_value AND before_version_id = after_version_id`, a.result.ChangeSetID, 1)
				}
			}
			requireEntityCount(t, harness, `SELECT count(*) FROM indicators WHERE incident_id = $1`, incident.ID, 1)
			requireEntityCount(t, harness, `SELECT count(*) FROM change_sets WHERE incident_id = $1 AND source = 'indicators.rows.create'`, incident.ID, commits)
			requireEntityCount(t, harness, `SELECT count(*) FROM route_idempotency WHERE scope_key = $1 AND route_key = 'indicators.rows.create'`, incident.ID.String()+":"+indicators.ViewSchemaID, commits)
			collaborationsupport.RequireIntentCount(t, harness.DB, collaborationsupport.IntentSelector{SourceRecordID: a.result.RecordID.String()}, 1)
			requireEntityCount(t, harness, `SELECT count(*) FROM record_revisions WHERE record_id = $1`, a.result.RecordID, 1)
			requireEntityCount(t, harness, `SELECT count(*) FROM records WHERE record_id = $1 AND row_version = 1`, a.result.RecordID, 1)
		})
	}
}

func mustCreateRowJSON(t testing.TB, result indicators.CreateResult) []byte {
	t.Helper()
	body, err := json.Marshal(result.CanonicalRow)
	if err != nil {
		t.Fatal(err)
	}
	return body
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
