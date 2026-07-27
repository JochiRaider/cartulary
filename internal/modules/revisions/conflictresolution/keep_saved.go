package conflictresolution

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/revisions/conflicttokens"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type Command struct {
	Actor       authn.UserRecord
	RecordID    uuid.UUID
	Claims      conflicttokens.ConflictTokenClaims
	ClientTxnID string
	RequestHash []byte
	RequestID   string
	RouteKey    string
}

type Target struct {
	IncidentID uuid.UUID
	RecordID   uuid.UUID
	RowVersion int64
	Row        map[string]any
}

type TargetLoader func(context.Context, pgx.Tx, Command) (Target, error)

type Result struct {
	Payload      map[string]any
	StatusCode   int
	Replayed     bool
	IncidentID   uuid.UUID
	RecordID     uuid.UUID
	ClientTxnID  string
	RowVersion   int64
	ViewSchemaID string
}

// KeepSaved records only the route-idempotency success for a source-validated
// canonical row. Source owners supply the locking loader and semantic
// revalidation; this coordinator owns only generic transaction/replay mechanics.
func KeepSaved(
	ctx context.Context,
	pool postgres.DB,
	authStore *authn.Store,
	command Command,
	load TargetLoader,
) (Result, error) {
	idempotencyKey := authn.RouteIdempotencyKey{
		RouteKey:    command.RouteKey,
		ActorUserID: command.Actor.ID,
		ScopeKey:    command.RecordID.String(),
		ClientTxnID: command.ClientTxnID,
	}
	if existing, err := authStore.GetRouteIdempotency(ctx, idempotencyKey); err == nil {
		if !bytes.Equal(existing.RequestHash, command.RequestHash) {
			return Result{}, authn.ErrClientTxnConflict
		}
		payload, err := decodeStoredResponse(existing.ResponseJSON)
		if err != nil {
			return Result{}, fmt.Errorf("decode replayed conflict keep-saved payload: %w", err)
		}
		return Result{
			Payload:      payload,
			StatusCode:   http.StatusOK,
			Replayed:     true,
			RecordID:     command.RecordID,
			ClientTxnID:  command.ClientTxnID,
			ViewSchemaID: command.Claims.ViewSchemaID,
		}, nil
	} else if !errors.Is(err, authn.ErrNotFound) {
		return Result{}, fmt.Errorf("query conflict keep-saved idempotency: %w", err)
	}

	tx, err := pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return Result{}, fmt.Errorf("begin conflict keep-saved transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	target, err := load(ctx, tx, command)
	if err != nil {
		return Result{}, err
	}
	payload := map[string]any{
		"view_schema_id": command.Claims.ViewSchemaID,
		"row":            target.Row,
	}
	if err := authn.InsertRouteIdempotencyPayload(
		ctx,
		tx,
		idempotencyKey,
		nil,
		command.RequestHash,
		http.StatusOK,
		payload,
	); err != nil {
		if authn.IsUniqueViolation(err) {
			return Result{}, authn.ErrClientTxnConflict
		}
		return Result{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Result{}, fmt.Errorf("commit conflict keep-saved transaction: %w", err)
	}
	return Result{
		Payload:      payload,
		StatusCode:   http.StatusOK,
		IncidentID:   target.IncidentID,
		RecordID:     target.RecordID,
		ClientTxnID:  command.ClientTxnID,
		RowVersion:   target.RowVersion,
		ViewSchemaID: command.Claims.ViewSchemaID,
	}, nil
}

func decodeStoredResponse(data []byte) (map[string]any, error) {
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	return payload, nil
}
