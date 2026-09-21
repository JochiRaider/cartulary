package artifacts

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

var (
	ErrClientTxnConflict          = fmt.Errorf("artifacts: client transaction conflict")
	ErrStoredMutationKindMismatch = fmt.Errorf("artifacts: stored mutation kind mismatch")
)

// OperationID is the closed set of Workbook-coordinated operations implemented
// by the Artifacts source owner. Its values remain the durable route identities.
type OperationID string

const (
	OperationCreate           OperationID = "workbook.rows.create"
	OperationPatch            OperationID = "workbook.records.patch"
	OperationConflictResolve  OperationID = "workbook.records.conflicts.resolve"
	OperationLinkedNoteCreate OperationID = "workbook.records.linked_notes.create"
	OperationNoteAssociations OperationID = "workbook.records.note_associations.mutate"
)

type IdempotencyKey struct {
	OperationID OperationID
	ActorUserID uuid.UUID
	ScopeKey    string
	ClientTxnID string
}

type StoredMutationKind string

const (
	StoredMutationCreate           StoredMutationKind = "create"
	StoredMutationPatch            StoredMutationKind = "patch"
	StoredMutationLinkedNote       StoredMutationKind = "linked_note"
	StoredMutationNoteAssociations StoredMutationKind = "note_associations"
)

type StoredMutationPayload struct {
	ViewSchemaID   string
	IncidentID     uuid.UUID
	RecordID       uuid.UUID
	RowVersion     int64
	ChangeSetID    *uuid.UUID
	Row            map[string]any
	ContextualLink *ContextualLink
}

// StoredMutationResult is a closed operation-tagged union used by the
// application idempotency adapter. It carries source facts, never HTTP status
// codes or transport response envelopes.
type StoredMutationResult struct {
	kind     StoredMutationKind
	workbook StoredMutationPayload
}

func NewStoredCreateResult(result StoredMutationPayload) StoredMutationResult {
	return StoredMutationResult{kind: StoredMutationCreate, workbook: cloneStoredMutationPayload(result)}
}

func NewStoredPatchResult(result StoredMutationPayload) StoredMutationResult {
	return StoredMutationResult{kind: StoredMutationPatch, workbook: cloneStoredMutationPayload(result)}
}

func NewStoredLinkedNoteResult(result StoredMutationPayload) StoredMutationResult {
	return StoredMutationResult{kind: StoredMutationLinkedNote, workbook: cloneStoredMutationPayload(result)}
}

func (r StoredMutationResult) Kind() StoredMutationKind { return r.kind }

func NewStoredNoteAssociationResult(result StoredMutationPayload) StoredMutationResult {
	return StoredMutationResult{kind: StoredMutationNoteAssociations, workbook: cloneStoredMutationPayload(result)}
}

func (r StoredMutationResult) Payload() (StoredMutationPayload, bool) {
	switch r.kind {
	case StoredMutationCreate, StoredMutationPatch, StoredMutationLinkedNote, StoredMutationNoteAssociations:
		return cloneStoredMutationPayload(r.workbook), validStoredMutationPayload(r.kind, r.workbook)
	default:
		return StoredMutationPayload{}, false
	}
}

type IdempotencyCapability interface {
	Get(context.Context, IdempotencyKey, []byte) (StoredMutationResult, bool, error)
	GetTx(context.Context, pgx.Tx, IdempotencyKey, []byte) (StoredMutationResult, bool, error)
	PutTx(context.Context, pgx.Tx, IdempotencyKey, []byte, StoredMutationResult) error
}

type storedMutationExpectation struct {
	sourceRecordID *uuid.UUID
	kind           StoredMutationKind
	viewSchemaID   string
	recordID       *uuid.UUID
}

func (f *MutationFacade) replayStoredMutation(
	ctx context.Context,
	key IdempotencyKey,
	requestHash []byte,
	operation string,
	expectation storedMutationExpectation,
) (StoredMutationPayload, bool, error) {
	existing, found, err := f.idempotency.Get(ctx, key, requestHash)
	return validateStoredMutation(existing, found, err, operation, expectation)
}

func validateStoredMutation(existing StoredMutationResult, found bool, err error, operation string, expectation storedMutationExpectation) (StoredMutationPayload, bool, error) {
	if err != nil {
		return StoredMutationPayload{}, false, fmt.Errorf("query artifact %s idempotency: %w", operation, err)
	}
	if !found {
		return StoredMutationPayload{}, false, nil
	}
	if existing.Kind() != expectation.kind {
		return StoredMutationPayload{}, false, ErrStoredMutationKindMismatch
	}
	stored, ok := existing.Payload()
	if !ok || !validStoredMutationPayload(existing.Kind(), stored) || stored.ViewSchemaID != expectation.viewSchemaID ||
		(expectation.recordID != nil && stored.RecordID != *expectation.recordID) ||
		(expectation.kind != StoredMutationPatch && !sameContextualSource(stored.ContextualLink, expectation.sourceRecordID)) {
		return StoredMutationPayload{}, false, ErrStoredMutationKindMismatch
	}
	return stored, true, nil
}

func validStoredMutationPayload(kind StoredMutationKind, stored StoredMutationPayload) bool {
	if stored.ViewSchemaID == "" || stored.IncidentID == uuid.Nil || stored.RecordID == uuid.Nil ||
		stored.RowVersion < 1 || stored.Row == nil {
		return false
	}
	if stored.ChangeSetID == nil && kind != StoredMutationNoteAssociations {
		return false
	}
	if stored.ChangeSetID != nil && *stored.ChangeSetID == uuid.Nil {
		return false
	}
	switch kind {
	case StoredMutationNoteAssociations:
		return stored.ContextualLink == nil && stored.ViewSchemaID == NotesViewSchemaID
	case StoredMutationCreate:
		return stored.ContextualLink == nil || (isCoordinationView(stored.ViewSchemaID) && stored.ContextualLink.SourceRecordID != uuid.Nil && stored.ContextualLink.LinkType == "references_artifact")
	case StoredMutationPatch:
		return stored.ContextualLink == nil
	case StoredMutationLinkedNote:
		return stored.ContextualLink != nil && stored.ContextualLink.SourceRecordID != uuid.Nil &&
			stored.ContextualLink.LinkType == "references_artifact"
	default:
		return false
	}
}

func cloneStoredMutationPayload(stored StoredMutationPayload) StoredMutationPayload {
	stored.ChangeSetID = cloneUUIDPointer(stored.ChangeSetID)
	stored.ContextualLink = cloneContextualLink(stored.ContextualLink)
	if stored.Row != nil {
		stored.Row = cloneMap(stored.Row)
	}
	return stored
}

func sameContextualSource(link *ContextualLink, source *uuid.UUID) bool {
	if source == nil {
		return link == nil
	}
	return link != nil && link.SourceRecordID == *source && link.LinkType == "references_artifact"
}
