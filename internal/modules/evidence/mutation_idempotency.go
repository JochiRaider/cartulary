package evidence

import (
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var (
	ErrClientTxnConflict          = errors.New("evidence: client transaction conflict")
	ErrStoredMutationKindMismatch = errors.New("evidence: stored mutation kind mismatch")
)

// OperationID is the closed set of Workbook-coordinated operations implemented
// by Evidence. Values are the existing durable route identities.
type OperationID string

const (
	OperationCreate          OperationID = "workbook.rows.create"
	OperationPatch           OperationID = "workbook.records.patch"
	OperationConflictResolve OperationID = "workbook.records.conflicts.resolve"
)

type IdempotencyKey struct {
	OperationID OperationID
	ActorUserID uuid.UUID
	ScopeKey    string
	ClientTxnID string
}

type StoredMutationKind string

const (
	StoredMutationCreate StoredMutationKind = "create"
	StoredMutationPatch  StoredMutationKind = "patch"
)

type StoredMutationPayload struct {
	ViewSchemaID string
	IncidentID   uuid.UUID
	RecordID     uuid.UUID
	RowVersion   int64
	ChangeSetID  *uuid.UUID
	Row          map[string]any
}

// StoredMutationResult is a closed operation-tagged union carrying source
// facts rather than transport status or response envelopes.
type StoredMutationResult struct {
	kind    StoredMutationKind
	payload StoredMutationPayload
}

func NewStoredCreateResult(payload StoredMutationPayload) StoredMutationResult {
	return StoredMutationResult{kind: StoredMutationCreate, payload: cloneStoredMutationPayload(payload)}
}

func NewStoredPatchResult(payload StoredMutationPayload) StoredMutationResult {
	return StoredMutationResult{kind: StoredMutationPatch, payload: cloneStoredMutationPayload(payload)}
}

func (result StoredMutationResult) Kind() StoredMutationKind { return result.kind }

func (result StoredMutationResult) Payload() (StoredMutationPayload, bool) {
	switch result.kind {
	case StoredMutationCreate, StoredMutationPatch:
		return cloneStoredMutationPayload(result.payload), true
	default:
		return StoredMutationPayload{}, false
	}
}

type IdempotencyCapability interface {
	Get(context.Context, IdempotencyKey, []byte) (StoredMutationResult, bool, error)
	PutTx(context.Context, pgx.Tx, IdempotencyKey, []byte, StoredMutationResult) error
}

func (f *mutationFacade) replayStoredMutation(
	ctx context.Context,
	key IdempotencyKey,
	requestHash []byte,
	want StoredMutationKind,
) (StoredMutationPayload, bool, error) {
	existing, found, err := f.idempotency.Get(ctx, key, requestHash)
	if err != nil {
		return StoredMutationPayload{}, false, fmt.Errorf("query Evidence idempotency: %w", err)
	}
	if !found {
		return StoredMutationPayload{}, false, nil
	}
	stored, ok := existing.Payload()
	if !ok || existing.Kind() != want || !validStoredMutationPayload(stored) {
		return StoredMutationPayload{}, false, ErrStoredMutationKindMismatch
	}
	return stored, true, nil
}

func validStoredMutationPayload(stored StoredMutationPayload) bool {
	return stored.ViewSchemaID == ViewSchemaID && stored.IncidentID != uuid.Nil && stored.RecordID != uuid.Nil &&
		stored.RowVersion > 0 && stored.ChangeSetID != nil && *stored.ChangeSetID != uuid.Nil && stored.Row != nil
}

func cloneStoredMutationPayload(stored StoredMutationPayload) StoredMutationPayload {
	if stored.ChangeSetID != nil {
		value := *stored.ChangeSetID
		stored.ChangeSetID = &value
	}
	stored.Row = cloneStringAnyMap(stored.Row)
	return stored
}

func cloneStringAnyMap(source map[string]any) map[string]any {
	if source == nil {
		return nil
	}
	cloned := make(map[string]any, len(source))
	for key, value := range source {
		cloned[key] = cloneAdmissionValue(value)
	}
	return cloned
}
