package collaboration_test

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	collabscenariotest "github.com/JochiRaider/cartulary/internal/modules/collaboration/testsupport/scenariotest"
)

func TestHistoricalIntentSuppressionIsTransactionLocal_Integration(t *testing.T) {
	runtime := collabscenariotest.StartRuntime(t)
	harness := runtime.StartServer(t, "historical-intent-suppression")
	ctx := context.Background()
	conn, err := harness.Pool.Acquire(ctx)
	if err != nil {
		t.Fatalf("acquire postgres connection: %v", err)
	}
	defer conn.Release()
	policy := collaboration.NewHistoricalIntentPolicy()

	for _, finish := range []struct {
		name string
		run  func(context.Context, pgx.Tx) error
	}{
		{name: "rollback", run: func(ctx context.Context, tx pgx.Tx) error { return tx.Rollback(ctx) }},
		{name: "commit", run: func(ctx context.Context, tx pgx.Tx) error { return tx.Commit(ctx) }},
	} {
		t.Run(finish.name, func(t *testing.T) {
			tx, err := conn.BeginTx(ctx, pgx.TxOptions{})
			if err != nil {
				t.Fatalf("begin suppression transaction: %v", err)
			}
			if historicalIntentsSuppressed(t, ctx, policy, tx) {
				_ = tx.Rollback(ctx)
				t.Fatal("new transaction unexpectedly began with historical intents suppressed")
			}
			if err := policy.SuppressTx(ctx, tx); err != nil {
				_ = tx.Rollback(ctx)
				t.Fatalf("suppress historical intents: %v", err)
			}
			if !historicalIntentsSuppressed(t, ctx, policy, tx) {
				_ = tx.Rollback(ctx)
				t.Fatal("suppression was not visible inside its transaction")
			}
			if err := finish.run(ctx, tx); err != nil {
				t.Fatalf("%s suppression transaction: %v", finish.name, err)
			}

			next, err := conn.BeginTx(ctx, pgx.TxOptions{})
			if err != nil {
				t.Fatalf("begin transaction after %s: %v", finish.name, err)
			}
			defer next.Rollback(ctx)
			if historicalIntentsSuppressed(t, ctx, policy, next) {
				t.Fatalf("suppression leaked into transaction after %s", finish.name)
			}
			if err := next.Rollback(ctx); err != nil {
				t.Fatalf("roll back transaction after %s: %v", finish.name, err)
			}
		})
	}
}

func historicalIntentsSuppressed(t testing.TB, ctx context.Context, policy *collaboration.HistoricalIntentPolicy, tx pgx.Tx) bool {
	t.Helper()
	suppressed, err := policy.IsSuppressedTx(ctx, tx)
	if err != nil {
		t.Fatalf("read historical intent suppression: %v", err)
	}
	return suppressed
}
