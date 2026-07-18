package storetest

import (
	"context"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

func TestStartStoreUsesRollbackDBForSeedsAssertionsAndNestedTransactions(t *testing.T) {
	email := "authenticationstore-rollback@example.test"

	t.Run("seed visible inside rollback fixture", func(t *testing.T) {
		harness := StartStore(t, "authenticationstore-rollback")
		user := SeedLocalUserRecord(t, harness.DB, email, "Rollback User", "RollbackPass1!", false, false, true)

		store := authn.NewStore(harness.DB)
		lookedUp, err := store.GetUserByNormalizedEmail(context.Background(), email)
		if err != nil {
			t.Fatalf("lookup seeded user through store: %v", err)
		}
		if lookedUp.ID != user.ID {
			t.Fatalf("unexpected seeded user lookup: got %s want %s", lookedUp.ID, user.ID)
		}
		if got := QueryCount(t, harness.DB, `SELECT COUNT(*) FROM users WHERE email = $1`, email); got != 1 {
			t.Fatalf("expected assertion helper to see seeded user, got %d", got)
		}

		tx, err := harness.DB.BeginTx(context.Background(), pgx.TxOptions{})
		if err != nil {
			t.Fatalf("begin nested transaction: %v", err)
		}
		if _, err := tx.Exec(context.Background(), `
INSERT INTO users (email, display_name, password_hash, mfa_required, is_active, is_deployment_admin)
VALUES ($1, 'Nested User', 'hash', false, true, false)
`, "authenticationstore-nested@example.test"); err != nil {
			t.Fatalf("insert inside nested transaction: %v", err)
		}
		if err := tx.Rollback(context.Background()); err != nil {
			t.Fatalf("rollback nested transaction: %v", err)
		}
		if got := QueryCount(t, harness.DB, `SELECT COUNT(*) FROM users WHERE email = $1`, "authenticationstore-nested@example.test"); got != 0 {
			t.Fatalf("expected nested rollback to hide row, got %d", got)
		}
	})

	t.Run("next fixture does not see prior rollback fixture rows", func(t *testing.T) {
		harness := StartStore(t, "authenticationstore-rollback")
		if got := QueryCount(t, harness.DB, `SELECT COUNT(*) FROM users WHERE email = $1`, email); got != 0 {
			t.Fatalf("expected second rollback fixture to start isolated, got %d", got)
		}
	})
}
