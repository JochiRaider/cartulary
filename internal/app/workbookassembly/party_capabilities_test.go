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
	"github.com/JochiRaider/cartulary/internal/modules/parties"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestPartyIdempotencyAdapterPreservesSemanticResultsAndReplayOrder_Integration(t *testing.T) {
	ctx := context.Background()
	db := pgtest.Start(t).BeginRollbackDBT(t, "parties-idempotency-adapter")
	actor := authstoretest.SeedLocalUserRecord(
		t, db, "parties-idempotency@example.test",
		"Parties Idempotency", "PartiesIdempotency1!", false, false, true,
	)
	adapter := partyIdempotency{store: authn.NewStore(db)}
	key := parties.IdempotencyKey{
		RouteKey: workbookCreateOperation, ActorUserID: actor.ID,
		ScopeKey:    uuid.NewString() + ":" + parties.ViewSchemaID,
		ClientTxnID: "txn-parties-idempotency-adapter",
	}
	recordID := uuid.New()
	changeSetID := uuid.New()
	result := parties.NewStoredCreateResult(parties.StoredRowMutationResult{
		Outcome: parties.MutationCreated, ViewSchemaID: parties.ViewSchemaID,
		RecordID: recordID, ChangeSetID: changeSetID,
		Row: map[string]any{"record_id": recordID.String()},
	})

	t.Run("operation mismatch", func(t *testing.T) {
		tx, err := db.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			t.Fatalf("begin mismatch transaction: %v", err)
		}
		defer func() { _ = tx.Rollback(ctx) }()
		patchResult := parties.NewStoredPatchResult(parties.StoredRowMutationResult{
			Outcome: parties.MutationUpdated, ViewSchemaID: parties.ViewSchemaID,
			RecordID: recordID, ChangeSetID: changeSetID,
			Row: map[string]any{"record_id": recordID.String()},
		})
		if err := adapter.PutTx(ctx, tx, key, []byte("hash"), patchResult); !errors.Is(err, parties.ErrStoredMutationKindMismatch) {
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
	var status int
	if err := db.QueryRow(ctx, `SELECT status_code FROM route_idempotency WHERE route_key = $1 AND actor_user_id = $2 AND scope_key = $3 AND client_txn_id = $4`, key.RouteKey, key.ActorUserID, key.ScopeKey, key.ClientTxnID).Scan(&status); err != nil {
		t.Fatalf("load stored status: %v", err)
	}
	if status != 201 {
		t.Fatalf("created Party replay status = %d, want 201", status)
	}

	stored, err := adapter.Get(ctx, key, []byte("hash"))
	if err != nil {
		t.Fatalf("load exact replay: %v", err)
	}
	row, ok := stored.Result.RowMutationResult()
	if !ok || stored.Result.Kind() != parties.StoredMutationCreate || row.RecordID != recordID || row.ChangeSetID != changeSetID {
		t.Fatalf("stored semantic result = %#v", stored.Result)
	}

	t.Run("platform uniqueness conflict", func(t *testing.T) {
		tx, err := db.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			t.Fatalf("begin duplicate transaction: %v", err)
		}
		defer func() { _ = tx.Rollback(ctx) }()
		if err := adapter.PutTx(ctx, tx, key, []byte("hash"), result); !errors.Is(err, parties.ErrClientTxnConflict) {
			t.Fatalf("duplicate write error = %v", err)
		}
	})

	t.Run("request hash conflict precedes stored payload decoding", func(t *testing.T) {
		malformedKey := key
		malformedKey.ClientTxnID = "txn-parties-malformed-replay"
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

		conflictRecord, err := adapter.Get(ctx, malformedKey, []byte("changed-request-hash"))
		if err != nil {
			t.Fatalf("changed-hash lookup decoded malformed payload: %v", err)
		}
		if !bytes.Equal(conflictRecord.RequestHash, storedHash) || conflictRecord.Result.Kind() != "" {
			t.Fatalf("changed-hash lookup = %#v", conflictRecord)
		}
		_, err = adapter.Get(ctx, malformedKey, storedHash)
		if err == nil || !strings.Contains(err.Error(), "decode Parties stored mutation result") {
			t.Fatalf("exact replay malformed-payload error = %v", err)
		}
	})
}
