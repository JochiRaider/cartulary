package artifacts

import (
	"bytes"
	"context"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var (
	ErrIdempotencyNotFound        = errors.New("artifacts: idempotency result not found")
	ErrClientTxnConflict          = errors.New("artifacts: client transaction conflict")
	ErrStoredMutationKindMismatch = errors.New("artifacts: stored mutation kind mismatch")
)

// OperationID is the closed set of Workbook-coordinated operations implemented
// by the Artifacts source owner. Its values remain the durable route identities.
type OperationID string

const (
	OperationCreate           OperationID = "workbook.rows.create"
	OperationPatch            OperationID = "workbook.records.patch"
	OperationConflictResolve  OperationID = "workbook.records.conflicts.resolve"
	OperationLinkedNoteCreate OperationID = "workbook.records.linked_notes.create"
)

type IdempotencyKey struct {
	OperationID OperationID
	ActorUserID uuid.UUID
	ScopeKey    string
	ClientTxnID string
}

type IdempotencyRecord struct {
	RequestHash []byte
	Result      StoredMutationResult
}

type StoredMutationKind string

const (
	StoredMutationCreate     StoredMutationKind = "create"
	StoredMutationPatch      StoredMutationKind = "patch"
	StoredMutationLinkedNote StoredMutationKind = "linked_note"
)

type StoredMutationPayload struct {
	ViewSchemaID   string
	RecordID       uuid.UUID
	ChangeSetID    uuid.UUID
	Row            map[string]any
	SourceRecordID *uuid.UUID
	LinkType       string
}

// StoredMutationResult is a closed operation-tagged union used by the
// application idempotency adapter. It carries source facts, never HTTP status
// codes or transport response envelopes.
type StoredMutationResult struct {
	kind     StoredMutationKind
	workbook StoredMutationPayload
}

func NewStoredCreateResult(result StoredMutationPayload) StoredMutationResult {
	return StoredMutationResult{kind: StoredMutationCreate, workbook: result}
}

func NewStoredPatchResult(result StoredMutationPayload) StoredMutationResult {
	return StoredMutationResult{kind: StoredMutationPatch, workbook: result}
}

func NewStoredLinkedNoteResult(result StoredMutationPayload) StoredMutationResult {
	return StoredMutationResult{kind: StoredMutationLinkedNote, workbook: result}
}

func (r StoredMutationResult) Kind() StoredMutationKind { return r.kind }

func (r StoredMutationResult) Payload() (StoredMutationPayload, bool) {
	switch r.kind {
	case StoredMutationCreate, StoredMutationPatch, StoredMutationLinkedNote:
		return r.workbook, true
	default:
		return StoredMutationPayload{}, false
	}
}

type IdempotencyCapability interface {
	Get(context.Context, IdempotencyKey, []byte) (IdempotencyRecord, error)
	PutTx(context.Context, pgx.Tx, IdempotencyKey, []byte, StoredMutationResult) error
}

type storedMutationExpectation struct {
	kind         StoredMutationKind
	viewSchemaID string
	recordID     *uuid.UUID
}

func (f *MutationFacade) replayStoredMutation(
	ctx context.Context,
	key IdempotencyKey,
	requestHash []byte,
	operation string,
	expectation storedMutationExpectation,
) (StoredMutationPayload, bool, error) {
	existing, err := f.idempotency.Get(ctx, key, requestHash)
	if errors.Is(err, ErrIdempotencyNotFound) {
		return StoredMutationPayload{}, false, nil
	}
	if err != nil {
		return StoredMutationPayload{}, false, fmt.Errorf("query artifact %s idempotency: %w", operation, err)
	}
	if !bytes.Equal(existing.RequestHash, requestHash) {
		return StoredMutationPayload{}, false, ErrClientTxnConflict
	}
	if existing.Result.Kind() != expectation.kind {
		return StoredMutationPayload{}, false, ErrStoredMutationKindMismatch
	}
	stored, ok := existing.Result.Payload()
	if !ok || stored.ViewSchemaID != expectation.viewSchemaID ||
		(expectation.recordID != nil && stored.RecordID != *expectation.recordID) {
		return StoredMutationPayload{}, false, ErrStoredMutationKindMismatch
	}
	return stored, true, nil
}
