package pgtest

import (
	"context"
	"database/sql"
	database_migrations "github.com/JochiRaider/cartulary/internal/modules/database_migrations"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
	"github.com/jackc/pgx/v5"
	"strings"
	"testing"
)

func TestPrepareIsolatedDatabaseTReturnsMigratedDatabase(t *testing.T) {
	harness := Start(t)
	testDB := harness.PrepareIsolatedDatabaseT(t, "bootstrap_t")

	db, err := sql.Open("pgx", testDB.DSN)
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() {
		_ = db.Close()
	})

	var exists bool
	if err := db.QueryRowContext(context.Background(), `
SELECT EXISTS (
    SELECT 1
    FROM information_schema.tables
    WHERE table_schema = 'public' AND table_name = 'users'
)`).Scan(&exists); err != nil {
		t.Fatalf("query users table: %v", err)
	}
	if !exists {
		t.Fatal("expected PrepareIsolatedDatabaseT to return a migrated database")
	}
	var format string
	if err := db.QueryRowContext(context.Background(), `SELECT format_id FROM public.application_crypto_format WHERE singleton`).Scan(&format); err != nil || format != database_migrations.ApplicationCryptoFormat {
		t.Fatalf("fresh fixture format admission = %q, %v", format, err)
	}
}

func TestBeginRollbackDBTIsolatesRowsPerTransaction(t *testing.T) {
	harness := Start(t)
	var originalRole, originalReplicationRole string

	t.Run("insert rows inside rollback fixture", func(t *testing.T) {
		db := harness.BeginRollbackDBT(t, "rollback-db")
		if err := db.QueryRow(context.Background(), `SELECT current_user, current_setting('session_replication_role')`).Scan(&originalRole, &originalReplicationRole); err != nil {
			t.Fatalf("read initial transaction roles: %v", err)
		}
		if _, err := db.Exec(context.Background(), `INSERT INTO users (email, display_name, password_hash, mfa_required) VALUES ($1, $2, $3, false)`, "rollback@example.test", "Rollback", "hash"); err != nil {
			t.Fatalf("seed rollback transaction: %v", err)
		}
		// A committed nested transaction is only a released savepoint. Its writes
		// and transaction-local roles must not survive the outer fixture rollback.
		nested, err := db.BeginTx(context.Background(), pgx.TxOptions{})
		if err != nil {
			t.Fatalf("begin nested transaction: %v", err)
		}
		if _, err := nested.Exec(context.Background(), `INSERT INTO users (email, display_name, password_hash, mfa_required) VALUES ($1, $2, $3, false)`, "nested-rollback@example.test", "Nested rollback", "hash"); err != nil {
			t.Fatalf("seed nested transaction: %v", err)
		}
		if err := nested.Commit(context.Background()); err != nil {
			t.Fatalf("commit nested savepoint: %v", err)
		}
		for _, statement := range []string{`SET LOCAL ROLE cartulary_recovery`, `SET LOCAL session_replication_role = replica`} {
			if _, err := db.Exec(context.Background(), statement); err != nil {
				t.Fatalf("set transaction-local state: %v", err)
			}
		}
	})

	t.Run("next rollback fixture starts clean", func(t *testing.T) {
		db := harness.BeginRollbackDBT(t, "rollback-db")
		var count int
		if err := db.QueryRow(context.Background(), `SELECT COUNT(*) FROM users WHERE email IN ($1, $2)`, "rollback@example.test", "nested-rollback@example.test").Scan(&count); err != nil {
			t.Fatalf("count rollback rows: %v", err)
		}
		if count != 0 {
			t.Fatalf("expected rollback fixture to discard rows, got %d", count)
		}
		var role, replicationRole string
		if err := db.QueryRow(context.Background(), `SELECT current_user, current_setting('session_replication_role')`).Scan(&role, &replicationRole); err != nil {
			t.Fatalf("read next borrower's transaction roles: %v", err)
		}
		if role != originalRole || replicationRole != originalReplicationRole {
			t.Fatalf("transaction-local state leaked: role=%q replication=%q", role, replicationRole)
		}
	})

	scope, ok, err := suiteservices.Summarize(nil)
	if err != nil {
		t.Fatalf("summarize suite service events: %v", err)
	}
	if !ok {
		t.Fatal("expected suite service summary")
	}
	transactionCount := 0
	for _, activity := range scope.Fixture.ByTest {
		if !strings.HasPrefix(activity.TestName, t.Name()+"/") {
			continue
		}
		if activity.Operation == "database-reset" {
			t.Fatalf("rollback fixture must not emit database-reset activity: %#v", activity)
		}
		if activity.Operation == "transaction" && activity.FixturePolicy == postgresFixturePolicyTransaction {
			transactionCount += activity.Count
		}
	}
	if transactionCount != 2 {
		t.Fatalf("expected two attributed rollback transaction events, got %d", transactionCount)
	}
}
