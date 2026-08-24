package conflicts

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type Command struct {
	ActorUserID uuid.UUID
	RecordID    uuid.UUID
	Claims      ConflictTokenClaims
	ClientTxnID string
	RequestHash []byte
	RequestID   string
	RouteKey    string
}

var (
	ErrIdempotencyNotFound = errors.New("conflict resolution: idempotency result not found")
	ErrClientTxnConflict   = errors.New("conflict resolution: client transaction conflict")
)

type IdempotencyKey struct {
	RouteKey    string
	ActorUserID uuid.UUID
	ScopeKey    string
	ClientTxnID string
}

type StoredTarget struct {
	ViewSchemaID string
	Row          map[string]any
}

type IdempotencyRecord struct {
	RequestHash []byte
	Target      StoredTarget
}

type IdempotencyPort interface {
	Get(context.Context, IdempotencyKey, []byte) (IdempotencyRecord, error)
	PutTx(context.Context, pgx.Tx, IdempotencyKey, []byte, StoredTarget) error
}

type TransactionRunner interface {
	BeginTx(context.Context, pgx.TxOptions) (pgx.Tx, error)
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
	transactions TransactionRunner,
	idempotency IdempotencyPort,
	command Command,
	load TargetLoader,
) (Result, error) {
	idempotencyKey := IdempotencyKey{
		RouteKey:    command.RouteKey,
		ActorUserID: command.ActorUserID,
		ScopeKey:    command.RecordID.String(),
		ClientTxnID: command.ClientTxnID,
	}
	if existing, err := idempotency.Get(ctx, idempotencyKey, command.RequestHash); err == nil {
		if !bytes.Equal(existing.RequestHash, command.RequestHash) {
			return Result{}, ErrClientTxnConflict
		}
		if existing.Target.ViewSchemaID != command.Claims.ViewSchemaID {
			return Result{}, fmt.Errorf("replayed conflict keep-saved target does not match operation")
		}
		payload := storedTargetPayload(existing.Target)
		return Result{
			Payload:      payload,
			Replayed:     true,
			RecordID:     command.RecordID,
			ClientTxnID:  command.ClientTxnID,
			ViewSchemaID: command.Claims.ViewSchemaID,
		}, nil
	} else if !errors.Is(err, ErrIdempotencyNotFound) {
		return Result{}, fmt.Errorf("query conflict keep-saved idempotency: %w", err)
	}

	tx, err := transactions.BeginTx(ctx, pgx.TxOptions{})
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
	if err := idempotency.PutTx(ctx, tx, idempotencyKey, command.RequestHash, StoredTarget{
		ViewSchemaID: command.Claims.ViewSchemaID,
		Row:          target.Row,
	}); err != nil {
		return Result{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return Result{}, fmt.Errorf("commit conflict keep-saved transaction: %w", err)
	}
	return Result{
		Payload:      payload,
		IncidentID:   target.IncidentID,
		RecordID:     target.RecordID,
		ClientTxnID:  command.ClientTxnID,
		RowVersion:   target.RowVersion,
		ViewSchemaID: command.Claims.ViewSchemaID,
	}, nil
}

func DecodeStoredTarget(data []byte) (StoredTarget, error) {
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	var payload map[string]any
	if err := decoder.Decode(&payload); err != nil {
		return StoredTarget{}, err
	}
	var trailing any
	if err := decoder.Decode(&trailing); !errors.Is(err, io.EOF) {
		return StoredTarget{}, fmt.Errorf("decode replayed conflict keep-saved target")
	}
	viewSchemaID, viewOK := payload["view_schema_id"].(string)
	row, rowOK := payload["row"].(map[string]any)
	if !viewOK || !rowOK {
		return StoredTarget{}, fmt.Errorf("decode replayed conflict keep-saved target")
	}
	rawRowVersion, ok := row["row_version"].(json.Number)
	if !ok {
		return StoredTarget{}, fmt.Errorf("decode replayed conflict keep-saved row version")
	}
	rowVersion, err := rawRowVersion.Int64()
	if err != nil || rowVersion < 1 {
		return StoredTarget{}, fmt.Errorf("decode replayed conflict keep-saved row version")
	}
	row["row_version"] = rowVersion
	return StoredTarget{ViewSchemaID: viewSchemaID, Row: row}, nil
}

func storedTargetPayload(target StoredTarget) map[string]any {
	return map[string]any{
		"view_schema_id": target.ViewSchemaID,
		"row":            target.Row,
	}
}
