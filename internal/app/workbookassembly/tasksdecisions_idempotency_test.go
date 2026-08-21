package workbookassembly

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestTaskDecisionIdempotencyAdapterBoundary_Integration(t *testing.T) {
	ctx := context.Background()
	db := pgtest.Start(t).BeginRollbackDBT(t, "tasks-decisions-idempotency-adapter")
	actor := authstoretest.SeedLocalUserRecord(
		t, db, "tasks-decisions-idempotency@example.test",
		"Tasks Decisions Idempotency", "TasksDecisionsIdempotency1!", false, false, true,
	)
	adapter := taskDecisionIdempotency{store: authn.NewStore(db)}
	key := tasksdecisions.IdempotencyKey{
		RouteKey:    "workbook.rows.create",
		ActorUserID: actor.ID,
		ScopeKey:    uuid.NewString() + ":" + tasksdecisions.TaskRequestsViewSchemaID,
		ClientTxnID: "txn-idempotency-adapter-boundary",
	}
	recordID := uuid.New()
	result := tasksdecisions.NewStoredCreateResult(tasksdecisions.StoredRowMutationResult{
		ViewSchemaID: tasksdecisions.TaskRequestsViewSchemaID,
		RecordID:     recordID,
		ChangeSetID:  uuid.New(),
		Row:          map[string]any{"record_id": recordID.String()},
	})

	t.Run("operation mismatch", func(t *testing.T) {
		tx, err := db.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			t.Fatalf("begin mismatch transaction: %v", err)
		}
		defer func() { _ = tx.Rollback(ctx) }()
		patchResult := tasksdecisions.NewStoredPatchResult(tasksdecisions.StoredRowMutationResult{
			ViewSchemaID: tasksdecisions.TaskRequestsViewSchemaID,
			RecordID:     recordID,
			ChangeSetID:  uuid.New(),
			Row:          map[string]any{"record_id": recordID.String()},
		})
		if err := adapter.PutTx(ctx, tx, key, []byte("hash"), patchResult); !errors.Is(err, tasksdecisions.ErrStoredMutationKindMismatch) {
			t.Fatalf("operation mismatch error = %v", err)
		}
	})

	tx, err := db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin first write transaction: %v", err)
	}
	if err := adapter.PutTx(ctx, tx, key, []byte("hash"), result); err != nil {
		t.Fatalf("write first idempotency result: %v", err)
	}
	if err := tx.Commit(ctx); err != nil {
		t.Fatalf("commit first idempotency result: %v", err)
	}

	t.Run("platform uniqueness conflict", func(t *testing.T) {
		tx, err := db.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			t.Fatalf("begin duplicate transaction: %v", err)
		}
		defer func() { _ = tx.Rollback(ctx) }()
		if err := adapter.PutTx(ctx, tx, key, []byte("hash"), result); !errors.Is(err, tasksdecisions.ErrClientTxnConflict) {
			t.Fatalf("duplicate write error = %v; want client transaction conflict", err)
		}
	})

	t.Run("request hash conflict precedes stored payload decoding", func(t *testing.T) {
		malformedKey := key
		malformedKey.ClientTxnID = "txn-idempotency-malformed-replay"
		storedHash := []byte("stored-request-hash")
		tx, err := db.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			t.Fatalf("begin malformed replay transaction: %v", err)
		}
		if err := authn.InsertRouteIdempotency(ctx, tx, authn.RouteIdempotencyKey{
			RouteKey: malformedKey.RouteKey, ActorUserID: malformedKey.ActorUserID,
			ScopeKey: malformedKey.ScopeKey, ClientTxnID: malformedKey.ClientTxnID,
		}, nil, storedHash, 201, []byte(`{"view_schema_id":42}`)); err != nil {
			_ = tx.Rollback(ctx)
			t.Fatalf("seed malformed replay: %v", err)
		}
		if err := tx.Commit(ctx); err != nil {
			t.Fatalf("commit malformed replay: %v", err)
		}

		countRows := func(table string) int {
			t.Helper()
			var count int
			if err := db.QueryRow(ctx, "SELECT COUNT(*) FROM "+table).Scan(&count); err != nil {
				t.Fatalf("count %s: %v", table, err)
			}
			return count
		}
		trackedTables := []string{
			"task_requests",
			"decisions",
			"record_revisions",
			"task_request_grid_projection",
			"decision_grid_projection",
		}
		before := make(map[string]int, len(trackedTables))
		for _, table := range trackedTables {
			before[table] = countRows(table)
		}

		conflictRecord, err := adapter.Get(ctx, malformedKey, []byte("changed-request-hash"))
		if err != nil {
			t.Fatalf("changed-hash lookup decoded malformed payload: %v", err)
		}
		if !bytes.Equal(conflictRecord.RequestHash, storedHash) {
			t.Fatalf("changed-hash lookup request hash = %q; want %q", conflictRecord.RequestHash, storedHash)
		}
		if conflictRecord.Result.Kind() != "" {
			t.Fatalf("changed-hash lookup returned a stored result: %#v", conflictRecord.Result)
		}

		_, err = adapter.Get(ctx, malformedKey, storedHash)
		if err == nil || !strings.Contains(err.Error(), "decode Tasks/Decisions stored mutation result") {
			t.Fatalf("exact replay malformed-payload error = %v", err)
		}
		for _, table := range trackedTables {
			if after := countRows(table); after != before[table] {
				t.Fatalf("idempotency lookup changed %s rows: before=%d after=%d", table, before[table], after)
			}
		}
	})
}
