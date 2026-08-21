package tasksdecisions

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func TestStoredMutationOperationMismatchRejectedBeforeSourceMutation_Unit(t *testing.T) {
	recordID := uuid.New()
	idempotency := fixedReplayIdempotency{
		record: IdempotencyRecord{
			RequestHash: []byte("same-request"),
			Result: NewStoredPatchResult(StoredRowMutationResult{
				ViewSchemaID: TaskRequestsViewSchemaID,
				RecordID:     recordID,
				ChangeSetID:  uuid.New(),
				Row:          map[string]any{"record_id": recordID.String()},
			}),
		},
	}
	facade := &MutationFacade{idempotency: idempotency}
	_, err := facade.Create(context.Background(), CreateCommand{
		ActorUserID: uuid.New(),
		IncidentID:  uuid.New(),
		Request: CreateRequest{
			ViewSchemaID: TaskRequestsViewSchemaID,
			ClientTxnID:  "txn-operation-mismatch",
		},
		RequestHash: []byte("same-request"),
		RouteKey:    "workbook.rows.create",
	})
	if !errors.Is(err, ErrStoredMutationKindMismatch) {
		t.Fatalf("Create error = %v; want stored operation mismatch", err)
	}
}

type fixedReplayIdempotency struct {
	record IdempotencyRecord
}

func (f fixedReplayIdempotency) Get(context.Context, IdempotencyKey, []byte) (IdempotencyRecord, error) {
	return f.record, nil
}

func (fixedReplayIdempotency) PutTx(context.Context, pgx.Tx, IdempotencyKey, []byte, StoredMutationResult) error {
	return nil
}
