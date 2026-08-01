package workbookassembly

import (
	"context"
	"errors"
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
	result := tasksdecisions.NewStoredCreateResult(tasksdecisions.StoredWorkbookResult{
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
		patchResult := tasksdecisions.NewStoredPatchResult(tasksdecisions.StoredWorkbookResult{
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
}
