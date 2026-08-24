package workbookassembly

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"slices"
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
	incidentID := uuid.New()
	recordID := uuid.New()
	changeSetID := uuid.New()
	rowVersion := int64(1)
	key := parties.IdempotencyKey{
		RouteKey: workbookCreateOperation, ActorUserID: actor.ID,
		ScopeKey:    incidentID.String() + ":" + parties.ViewSchemaID,
		ClientTxnID: "txn-parties-idempotency-adapter",
	}
	adapter := partyIdempotency{
		store: authn.NewStore(db),
		metadata: partyReplayMetadataLoaderFunc(func(_ context.Context, gotChangeSetID uuid.UUID, gotRecordID uuid.UUID) (partyReplayMetadata, error) {
			if gotChangeSetID != changeSetID || gotRecordID != recordID {
				return partyReplayMetadata{}, fmt.Errorf("unexpected replay identity %s/%s", gotChangeSetID, gotRecordID)
			}
			return partyReplayMetadata{
				IncidentID: incidentID, ActorUserID: actor.ID, Source: workbookCreateOperation,
				ClientTxnID: key.ClientTxnID, OperationKind: "create", RowVersion: &rowVersion,
				ChangedFieldKeys: []string{"party.display_name"},
			}, nil
		}),
	}
	result := parties.NewStoredCreateResult(parties.StoredRowMutationResult{
		Outcome: parties.MutationCreated, ViewSchemaID: parties.ViewSchemaID,
		IncidentID: incidentID, RecordID: recordID, ChangeSetID: changeSetID, RowVersion: rowVersion,
		ChangedFieldKeys: []string{"party.display_name"},
		Row:              map[string]any{"record_id": recordID.String(), "row_version": rowVersion},
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
	var storedResponseBefore []byte
	if err := db.QueryRow(ctx, `SELECT status_code, response_json FROM route_idempotency WHERE route_key = $1 AND actor_user_id = $2 AND scope_key = $3 AND client_txn_id = $4`, key.RouteKey, key.ActorUserID, key.ScopeKey, key.ClientTxnID).Scan(&status, &storedResponseBefore); err != nil {
		t.Fatalf("load stored status: %v", err)
	}
	if status != 201 {
		t.Fatalf("created Party replay status = %d, want 201", status)
	}

	stored, found, err := adapter.Get(ctx, key, []byte("hash"))
	if err != nil {
		t.Fatalf("load exact replay: %v", err)
	}
	row, ok := stored.RowMutationResult()
	if !found || !ok || stored.Kind() != parties.StoredMutationCreate || row.RecordID != recordID ||
		row.ChangeSetID != changeSetID || row.IncidentID != incidentID || row.RowVersion != rowVersion ||
		!slices.Equal(row.ChangedFieldKeys, []string{"party.display_name"}) {
		t.Fatalf("stored semantic result = %#v", stored)
	}
	var storedStatusAfter int
	var storedResponseAfter []byte
	if err := db.QueryRow(ctx, `SELECT status_code, response_json FROM route_idempotency WHERE route_key = $1 AND actor_user_id = $2 AND scope_key = $3 AND client_txn_id = $4`, key.RouteKey, key.ActorUserID, key.ScopeKey, key.ClientTxnID).Scan(&storedStatusAfter, &storedResponseAfter); err != nil {
		t.Fatalf("reload stored response: %v", err)
	}
	if storedStatusAfter != status || !bytes.Equal(storedResponseAfter, storedResponseBefore) {
		t.Fatalf("replay changed stored payload/status: before=(%d,%s) after=(%d,%s)", status, storedResponseBefore, storedStatusAfter, storedResponseAfter)
	}

	t.Run("miss is distinct", func(t *testing.T) {
		missKey := key
		missKey.ClientTxnID = "txn-parties-missing"
		_, found, err := adapter.Get(ctx, missKey, []byte("hash"))
		if err != nil || found {
			t.Fatalf("missing lookup = found %t, err %v", found, err)
		}
	})

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

		_, found, err := adapter.Get(ctx, malformedKey, []byte("changed-request-hash"))
		if !errors.Is(err, parties.ErrClientTxnConflict) || found {
			t.Fatalf("changed-hash lookup = found %t, err %v", found, err)
		}
		_, found, err = adapter.Get(ctx, malformedKey, storedHash)
		if err == nil || !strings.Contains(err.Error(), "decode Parties stored mutation result") {
			t.Fatalf("exact replay malformed-payload = found %t, err %v", found, err)
		}
	})
}

type partyReplayMetadataLoaderFunc func(context.Context, uuid.UUID, uuid.UUID) (partyReplayMetadata, error)

func (load partyReplayMetadataLoaderFunc) Load(
	ctx context.Context,
	changeSetID uuid.UUID,
	recordID uuid.UUID,
) (partyReplayMetadata, error) {
	return load(ctx, changeSetID, recordID)
}
