package graphassembly

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/JochiRaider/cartulary/internal/modules/graphprojection/postgresresult"
	graphrestore "github.com/JochiRaider/cartulary/internal/modules/graphprojection/restore"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/postgres/dbport"
	"github.com/JochiRaider/cartulary/internal/testutil/graphfixture"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestClaimedGraphReconstructionUsesOneTransaction_Integration(t *testing.T) {
	postgresHarness := pgtest.Start(t)
	for _, scenario := range []string{"success", "source enumeration failure", "source cancellation", "reverse result lifecycle"} {
		t.Run(scenario, func(t *testing.T) {
			database := postgresHarness.PrepareIsolatedDatabaseT(t, "claimed-graph-assembly")
			dsn, err := database.DSNForPurpose(postgres.PurposeRuntime)
			if err != nil {
				t.Fatal(err)
			}
			handle, err := postgres.Setup(t.Context(), postgres.Settings{BindingKind: "managed_service", DSN: dsn, Purpose: postgres.PurposeRuntime, ExpectedRole: "cartulary_runtime"})
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(handle.Close)
			db := handle.Pool()
			if scenario == "reverse result lifecycle" {
				checkReverseReconstruction(t, db)
				return
			}
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			seedClaimedGraphResult(t, db)
			before := claimedGraphState(t, db)
			observed := &graphSourceTransactionWitness{DB: db}
			switch scenario {
			case "source enumeration failure":
				observed.sourceErr = errors.New("SECRET source enumeration failure")
			case "source cancellation":
				observed.cancelSource = cancel
			}
			result, err := RebuildClaimedGraph(ctx, observed, claimedGraphRequest(ctx))
			if observed.sourceOutside != 0 || observed.sourceTransaction == 0 || observed.sourceTransaction != observed.clearTransaction ||
				observed.transactions != 1 || observed.sourceBeforeClear || !observed.clearConfirmed ||
				observed.isolation != "repeatable read" || observed.role != "cartulary_runtime" {
				t.Fatalf("reconstruction transaction witness: %#v", observed)
			}
			if scenario == "success" {
				if err != nil || !result.ReadinessSatisfied() || len(result.RebuiltViews) != 0 || observed.commitAttempts != 1 {
					t.Fatalf("assembled Graph restore result: %#v, %v; commits=%d", result, err, observed.commitAttempts)
				}
				if after := claimedGraphState(t, db); after == before {
					t.Fatal("successful empty-source rebuild did not commit the real clear")
				}
				return
			}
			if code, ok := graphrestore.RestoreErrorCodeOf(err); !ok || code != graphrestore.RestoreErrorSourceEnumeration {
				t.Fatalf("prepublication source failure: code=%q err=%v", code, err)
			}
			if result.Validate() != nil || result.Status != graphrestore.RestoreStatusFailed || result.ReadinessOutcome != graphrestore.RestoreReadinessIncomplete ||
				len(result.ClearedTableIDs) != 0 || len(result.RebuiltViews) != 0 || result.ReconciledNonterminalJobCount != 0 || result.PostconditionSHA256 != nil {
				t.Fatalf("failed reconstruction claimed committed facts: %#v", result)
			}
			body, marshalErr := json.Marshal(result)
			if marshalErr != nil || strings.Contains(string(body), "SECRET") || strings.Contains(err.Error(), "SECRET") {
				t.Fatal("source failure did not retain its closed safe result")
			}
			if observed.commitAttempts != 0 || observed.rollbackCalls != 1 || !observed.rollbackSucceeded {
				t.Fatalf("prepublication failure did not roll back: commits=%d rollbacks=%d succeeded=%v", observed.commitAttempts, observed.rollbackCalls, observed.rollbackSucceeded)
			}
			// Read through the pool with an independent context before any fixture
			// cleanup. This compares full persisted rows, not only object counts.
			if after := claimedGraphState(t, db); after != before {
				t.Fatal("failed reconstruction changed committed Graph or owner state")
			}
			retry, err := db.BeginTx(t.Context(), pgx.TxOptions{})
			if err != nil {
				t.Fatal(err)
			}
			if _, err := retry.Exec(t.Context(), "SELECT 1"); err != nil {
				t.Fatal(err)
			}
			if err := retry.Rollback(t.Context()); err != nil {
				t.Fatal(err)
			}
		})
	}
}

func claimedGraphRequest(ctx context.Context) graphrestore.RestoreRebuildRequest {
	registry := graphrestore.CurrentRestoreSourceRegistry()
	binding := graphrestore.CurrentRestoreImplementationBinding()
	return graphrestore.RestoreRebuildRequest{
		Context: ctx, RestoreOperationID: uuid.MustParse("00000000-0000-0000-0000-000000009001"),
		PublishedAt: time.Date(2026, 5, 30, 0, 0, 0, 0, time.UTC), TargetGenerationID: uuid.MustParse("00000000-0000-0000-0000-000000009003"),
		RecoveryStateCatalog: graphrestore.RestoreRecoveryCatalogRef{
			DigestSHA256: binding.Binding.RecoveryStateCatalogSHA256, AlgorithmID: graphrestore.RestoreAlgorithmID, GraphTableIDs: graphrestore.RestoreGraphTableIDs(),
		},
		SourceRegistry: graphrestore.RestoreSourceRegistryRef{Registry: registry, SHA256: registry.DigestSHA256()}, ImplementationBinding: binding,
	}
}

func seedClaimedGraphResult(t *testing.T, db dbport.DB) {
	t.Helper()
	result := graphfixture.Result(t, graphfixture.ResultOptions{
		GraphViewID: "nfgv_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", SourceOwnerID: "network_flow_activity",
		SourceSnapshotID: "snapshot-before-rebuild", ProjectionVersion: "assembly-fixture.v1",
		VertexKind: "endpoint", EdgeKind: "flow", VertexCount: 2, PublishedAt: time.Date(2026, 5, 29, 0, 0, 0, 0, time.UTC),
	})
	if len(result.Vertices) == 0 || len(result.Edges) == 0 {
		t.Fatal("rollback fixture must populate every derived table")
	}
	tx, err := db.BeginTx(t.Context(), pgx.TxOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = tx.Rollback(context.WithoutCancel(t.Context())) }()
	publisher, err := postgresresult.NewPublisher(tx)
	if err != nil {
		t.Fatal(err)
	}
	if err := publisher.PublishResult(t.Context(), result); err != nil {
		t.Fatal(err)
	}
	if err := tx.Commit(t.Context()); err != nil {
		t.Fatal(err)
	}
}

func claimedGraphState(t *testing.T, db dbport.DB) string {
	t.Helper()
	var state string
	if err := db.QueryRow(t.Context(), `SELECT jsonb_build_array(
 (SELECT coalesce(jsonb_agg(to_jsonb(r) ORDER BY projection_result_id), '[]') FROM graph_projection_results r),
 (SELECT coalesce(jsonb_agg(to_jsonb(v) ORDER BY projection_result_id, vertex_id), '[]') FROM graph_projection_result_vertices v),
 (SELECT coalesce(jsonb_agg(to_jsonb(e) ORDER BY projection_result_id, edge_id), '[]') FROM graph_projection_result_edges e),
 (SELECT coalesce(jsonb_agg(to_jsonb(d) ORDER BY graph_view_id), '[]') FROM network_flow_graph_views d),
 (SELECT coalesce(jsonb_agg(to_jsonb(j) ORDER BY job_id), '[]') FROM jobs j),
 (SELECT coalesce(jsonb_agg(to_jsonb(p) ORDER BY job_id), '[]') FROM reporting_job_payloads p)
)::text`).Scan(&state); err != nil {
		t.Fatal(err)
	}
	return state
}

// This witness observes actual managed PostgreSQL calls across the application
// boundary; a pool read cannot masquerade as the reconstruction transaction.
type graphSourceTransactionWitness struct {
	dbport.DB
	transactions, sourceOutside, sourceTransaction, clearTransaction int
	commitAttempts, rollbackCalls                                    int
	sourceBeforeClear, clearConfirmed, rollbackSucceeded             bool
	isolation, role                                                  string
	sourceErr                                                        error
	cancelSource                                                     context.CancelFunc
}

func (db *graphSourceTransactionWitness) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	if strings.Contains(sql, "FROM network_flow_graph_views") {
		db.sourceOutside++
	}
	return db.DB.Query(ctx, sql, args...)
}

func (db *graphSourceTransactionWitness) BeginTx(ctx context.Context, options pgx.TxOptions) (pgx.Tx, error) {
	tx, err := db.DB.BeginTx(ctx, options)
	if err != nil {
		return nil, err
	}
	if err := tx.QueryRow(ctx, "SELECT current_setting('transaction_isolation'), current_user").Scan(&db.isolation, &db.role); err != nil {
		_ = tx.Rollback(ctx)
		return nil, err
	}
	db.transactions++
	return &graphSourceTransaction{Tx: tx, owner: db, id: db.transactions}, nil
}

type graphSourceTransaction struct {
	pgx.Tx
	owner *graphSourceTransactionWitness
	id    int
}

func (tx *graphSourceTransaction) Query(ctx context.Context, sql string, args ...any) (pgx.Rows, error) {
	if strings.Contains(sql, "FROM network_flow_graph_views") {
		tx.owner.sourceTransaction = tx.id
		tx.owner.sourceBeforeClear = !tx.owner.clearConfirmed || tx.owner.clearTransaction != tx.id
		if tx.owner.sourceErr != nil {
			return nil, tx.owner.sourceErr
		}
		if tx.owner.cancelSource != nil {
			tx.owner.cancelSource()
		}
	}
	return tx.Tx.Query(ctx, sql, args...)
}

func (tx *graphSourceTransaction) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	tag, err := tx.Tx.Exec(ctx, sql, args...)
	if err == nil && strings.Contains(sql, "DELETE FROM graph_projection_results") {
		tx.owner.clearTransaction = tx.id
		var remaining int
		if err := tx.Tx.QueryRow(ctx, `SELECT
 (SELECT count(*) FROM graph_projection_results) +
 (SELECT count(*) FROM graph_projection_result_vertices) +
 (SELECT count(*) FROM graph_projection_result_edges)`).Scan(&remaining); err != nil {
			return tag, err
		}
		if remaining != 0 {
			return tag, fmt.Errorf("clear left %d derived rows in the transaction", remaining)
		}
		tx.owner.clearConfirmed = true
	}
	return tag, err
}

func (tx *graphSourceTransaction) Commit(ctx context.Context) error {
	tx.owner.commitAttempts++
	return tx.Tx.Commit(ctx)
}

func (tx *graphSourceTransaction) Rollback(ctx context.Context) error {
	tx.owner.rollbackCalls++
	err := tx.Tx.Rollback(ctx)
	tx.owner.rollbackSucceeded = err == nil
	return err
}
