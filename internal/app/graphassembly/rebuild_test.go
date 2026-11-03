package graphassembly

import (
	"context"
	graphrestore "github.com/JochiRaider/cartulary/internal/modules/graphprojection/restore"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"strings"
	"testing"
	"time"
)

func TestClaimedGraphReconstructionUsesOneTransaction_Integration(t *testing.T) {
	database := pgtest.Start(t).PrepareIsolatedDatabaseT(t, "claimed-graph-assembly")
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
	observed := &graphSourceTransactionWitness{DB: db}
	currentRegistry := graphrestore.CurrentRestoreSourceRegistry()
	registry := graphrestore.RestoreSourceRegistryRef{Registry: currentRegistry, SHA256: currentRegistry.DigestSHA256()}
	binding := graphrestore.CurrentRestoreImplementationBinding()
	ctx := context.Background()
	result, err := RebuildClaimedGraph(ctx, observed, graphrestore.RestoreRebuildRequest{
		Context:            ctx,
		RestoreOperationID: uuid.MustParse("00000000-0000-0000-0000-000000009001"),
		PublishedAt:        time.Date(2026, 5, 30, 0, 0, 0, 0, time.UTC),
		TargetGenerationID: uuid.MustParse("00000000-0000-0000-0000-000000009003"),
		RecoveryStateCatalog: graphrestore.RestoreRecoveryCatalogRef{
			DigestSHA256: binding.Binding.RecoveryStateCatalogSHA256, AlgorithmID: graphrestore.RestoreAlgorithmID,
			GraphTableIDs: graphrestore.RestoreGraphTableIDs(),
		},
		SourceRegistry: registry, ImplementationBinding: binding,
	})
	if err != nil {
		t.Fatalf("execute assembled Graph restore participant: %v result=%#v catalog=%s binding=%s", err, result, binding.Binding.RecoveryStateCatalogSHA256, binding.Binding.RecoveryStateCatalogSHA256)
	}
	if !result.ReadinessSatisfied() || len(result.RebuiltViews) != 0 {
		t.Fatalf("assembled Graph restore result mismatch: %#v", result)
	}
	if observed.sourceOutside != 0 || observed.sourceTransaction == 0 || observed.sourceTransaction != observed.publicationTransaction || observed.transactions != 1 || observed.sourceBeforeClear || observed.isolation != "repeatable read" || observed.role != "cartulary_runtime" {
		t.Fatalf("source enumeration and publication must share one transaction: outside=%d source=%d publication=%d transactions=%d", observed.sourceOutside, observed.sourceTransaction, observed.publicationTransaction, observed.transactions)
	}
}

// This witness observes actual managed PostgreSQL calls across the application
// boundary; a pool read cannot masquerade as the publication transaction.
type graphSourceTransactionWitness struct {
	postgres.DB
	transactions, sourceOutside, sourceTransaction, publicationTransaction int
	sourceBeforeClear                                                      bool
	isolation, role                                                        string
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
		tx.owner.sourceBeforeClear = tx.owner.publicationTransaction != tx.id
	}
	return tx.Tx.Query(ctx, sql, args...)
}
func (tx *graphSourceTransaction) Exec(ctx context.Context, sql string, args ...any) (pgconn.CommandTag, error) {
	if strings.Contains(sql, "DELETE FROM graph_projection_results") {
		tx.owner.publicationTransaction = tx.id
	}
	return tx.Tx.Exec(ctx, sql, args...)
}
