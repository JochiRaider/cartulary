package runtime_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	entitytest "github.com/JochiRaider/cartulary/internal/modules/entities/testsupport"
	projectiontestsupport "github.com/JochiRaider/cartulary/internal/modules/projections/testsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
)

func TestTypedEntityProjectionDeletionUsesCallerTransaction(t *testing.T) {
	ctx := context.Background()
	harness := appsupport.StartStore(t, "projection-typed-entity-deletion")
	actor := authstoretest.SeedLocalUserRecord(
		t,
		harness.DB,
		"projection-delete@example.test",
		"Projection Delete",
		"ProjectionDelete1!",
		false,
		false,
		true,
	)
	incident := appsupport.CreateIncidentInStore(
		t,
		harness.DB,
		actor,
		"txn-projection-delete-incident",
		"IR-PROJECTION-DELETE",
		"Projection delete transaction ownership",
	)
	rows := projectiontestsupport.MustBuild(t, harness.DB).EntityPorts().Writer

	tests := []struct {
		name        string
		table       string
		sourceTable string
		seed        func(uuid.UUID)
		insert      string
		deleteTx    func(context.Context, pgx.Tx, uuid.UUID) error
	}{
		{
			name:        "host",
			table:       "host_grid_projection",
			sourceTable: "hosts",
			seed: func(recordID uuid.UUID) {
				entitytest.SeedHostRecord(t, harness.DB, incident.ID, actor.ID, recordID, "Delete host", "delete-host", "", "")
			},
			insert: `INSERT INTO host_grid_projection (
    record_id, incident_id, row_version, display_name, hostname, host_state, edited_at
) VALUES ($1, $2, 1, 'Delete host', 'delete-host', 'canonical', now())`,
			deleteTx: rows.DeleteHostTx,
		},
		{
			name:        "identity",
			table:       "identity_grid_projection",
			sourceTable: "identities",
			seed: func(recordID uuid.UUID) {
				entitytest.SeedIdentityRecord(t, harness.DB, incident.ID, actor.ID, recordID, "Delete identity", "delete@example.test", "delete@example.test", "DELETE")
			},
			insert: `INSERT INTO identity_grid_projection (
    record_id, incident_id, row_version, display_name, upn, email, sam_account_name, identity_state, edited_at
) VALUES ($1, $2, 1, 'Delete identity', 'delete@example.test', 'delete@example.test', 'DELETE', 'canonical', now())`,
			deleteTx: rows.DeleteIdentityTx,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			recordID := uuid.New()
			test.seed(recordID)
			if _, err := harness.DB.Exec(ctx, test.insert, recordID, incident.ID); err != nil {
				t.Fatalf("seed %s projection: %v", test.name, err)
			}
			beforeHistory := projectionCharacterizationCount(t, harness, "record_revisions", recordID)

			tx, err := harness.DB.BeginTx(ctx, pgx.TxOptions{})
			if err != nil {
				t.Fatalf("begin rollback transaction: %v", err)
			}
			if err := test.deleteTx(ctx, tx, recordID); err != nil {
				t.Fatalf("delete %s projection for rollback: %v", test.name, err)
			}
			if got := projectionCharacterizationCountTx(t, tx, test.table, recordID); got != 0 {
				t.Fatalf("%s projection count inside delete transaction = %d", test.name, got)
			}
			if err := tx.Rollback(ctx); err != nil {
				t.Fatalf("rollback %s projection deletion: %v", test.name, err)
			}
			if got := projectionCharacterizationCount(t, harness, test.table, recordID); got != 1 {
				t.Fatalf("rollback retained %d %s projection rows, want 1", got, test.name)
			}

			tx, err = harness.DB.BeginTx(ctx, pgx.TxOptions{})
			if err != nil {
				t.Fatalf("begin commit transaction: %v", err)
			}
			if err := test.deleteTx(ctx, tx, recordID); err != nil {
				t.Fatalf("delete %s projection for commit: %v", test.name, err)
			}
			if err := tx.Commit(ctx); err != nil {
				t.Fatalf("commit %s projection deletion: %v", test.name, err)
			}
			if got := projectionCharacterizationCount(t, harness, test.table, recordID); got != 0 {
				t.Fatalf("committed deletion retained %d %s projection rows", got, test.name)
			}
			if got := projectionCharacterizationCount(t, harness, test.sourceTable, recordID); got != 1 {
				t.Fatalf("projection deletion changed %s authoritative rows: %d", test.name, got)
			}
			if got := projectionCharacterizationCount(t, harness, "record_revisions", recordID); got != beforeHistory {
				t.Fatalf("projection deletion changed %s history rows: got %d want %d", test.name, got, beforeHistory)
			}

			tx, err = harness.DB.BeginTx(ctx, pgx.TxOptions{})
			if err != nil {
				t.Fatalf("begin already-absent transaction: %v", err)
			}
			if err := test.deleteTx(ctx, tx, recordID); err != nil {
				t.Fatalf("already-absent %s projection delete: %v", test.name, err)
			}
			if err := tx.Commit(ctx); err != nil {
				t.Fatalf("commit already-absent %s projection deletion: %v", test.name, err)
			}
		})
	}
}

func projectionCharacterizationCount(
	t testing.TB,
	harness *appsupport.StoreHarness,
	table string,
	recordID uuid.UUID,
) int {
	t.Helper()
	query := fmt.Sprintf("SELECT count(*) FROM %s WHERE record_id = $1", pgx.Identifier{table}.Sanitize())
	var count int
	if err := harness.DB.QueryRow(context.Background(), query, recordID).Scan(&count); err != nil {
		t.Fatalf("count %s: %v", table, err)
	}
	return count
}

func projectionCharacterizationCountTx(t testing.TB, tx pgx.Tx, table string, recordID uuid.UUID) int {
	t.Helper()
	query := fmt.Sprintf("SELECT count(*) FROM %s WHERE record_id = $1", pgx.Identifier{table}.Sanitize())
	var count int
	if err := tx.QueryRow(context.Background(), query, recordID).Scan(&count); err != nil {
		t.Fatalf("count %s in transaction: %v", table, err)
	}
	return count
}
