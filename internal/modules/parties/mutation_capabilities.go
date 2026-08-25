package parties

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	partyprojection "github.com/JochiRaider/cartulary/internal/modules/parties/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
)

const ViewSchemaID = "cartulary.view.parties.v1"

var (
	ErrClientTxnConflict          = errors.New("parties: client transaction conflict")
	ErrStoredMutationKindMismatch = errors.New("parties: stored mutation kind mismatch")
)

type IdempotencyKey struct {
	RouteKey    string
	ActorUserID uuid.UUID
	ScopeKey    string
	ClientTxnID string
}

type StoredMutationKind string

const (
	StoredMutationCreate StoredMutationKind = "create"
	StoredMutationPatch  StoredMutationKind = "patch"
)

type StoredRowMutationResult struct {
	Outcome          MutationOutcome
	ViewSchemaID     string
	IncidentID       uuid.UUID
	RecordID         uuid.UUID
	ChangeSetID      uuid.UUID
	RowVersion       int64
	ChangedFieldKeys []string
	Row              map[string]any
}

// StoredMutationResult is a closed operation-tagged union. Application
// adapters can construct only create and patch replay results.
type StoredMutationResult struct {
	kind StoredMutationKind
	row  StoredRowMutationResult
}

func NewStoredCreateResult(result StoredRowMutationResult) StoredMutationResult {
	return StoredMutationResult{kind: StoredMutationCreate, row: result}
}

func NewStoredPatchResult(result StoredRowMutationResult) StoredMutationResult {
	return StoredMutationResult{kind: StoredMutationPatch, row: result}
}

func (r StoredMutationResult) Kind() StoredMutationKind { return r.kind }

func (r StoredMutationResult) RowMutationResult() (StoredRowMutationResult, bool) {
	if r.kind != StoredMutationCreate && r.kind != StoredMutationPatch {
		return StoredRowMutationResult{}, false
	}
	return r.row, true
}

type IncidentStateCapability interface {
	RequireOpenTx(context.Context, pgx.Tx, uuid.UUID) error
}

type IdempotencyCapability interface {
	Get(context.Context, IdempotencyKey, []byte) (StoredMutationResult, bool, error)
	PutTx(context.Context, pgx.Tx, IdempotencyKey, []byte, StoredMutationResult) error
}

type RecordEnvelopeCapability interface {
	InsertTx(context.Context, pgx.Tx, records.InsertParams) (uuid.UUID, error)
	AdvanceVersionTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, time.Time) (int64, error)
	LoadEnvelope(context.Context, uuid.UUID) (records.Envelope, error)
}

type RevisionCapability interface {
	CaptureRecordSnapshotTx(context.Context, pgx.Tx, uuid.UUID) (revisions.RecordSnapshot, error)
	AppendChangeSetTx(context.Context, pgx.Tx, revisions.AppendChangeSetParams) (uuid.UUID, error)
	AppendRecordMutationTx(context.Context, pgx.Tx, revisions.AppendRecordMutationParams) error
	AppendLiveRevisionTx(context.Context, pgx.Tx, revisions.LiveRevisionInput) error
	LoadRevisionWindowTx(context.Context, pgx.Tx, uuid.UUID, int64, int64) ([]conflicttokens.RevisionWindowRow, error)
}

type KeepSavedResult struct {
	Replayed     bool
	IncidentID   uuid.UUID
	RecordID     uuid.UUID
	ClientTxnID  string
	RowVersion   int64
	ViewSchemaID string
	Row          map[string]any
}

// KeepSavedCapability keeps generic Revisions replay mechanics behind an
// application boundary and returns only Party-semantic state.
type KeepSavedCapability interface {
	KeepSaved(
		context.Context,
		conflicttokens.TransactionRunner,
		conflicttokens.Command,
		conflicttokens.TargetLoader,
	) (KeepSavedResult, error)
}

type MutationDependencies struct {
	IncidentState   IncidentStateCapability
	Idempotency     IdempotencyCapability
	RecordEnvelopes RecordEnvelopeCapability
	Projections     partyprojection.Rows
	Revisions       RevisionCapability
	ConflictFields  conflicttokens.FieldResolver
	KeepSaved       KeepSavedCapability
	Collaboration   collaboration.RecordChangedAppender
}

func (d MutationDependencies) validate() error {
	required := []struct {
		name  string
		value any
	}{
		{name: "Incident admission", value: d.IncidentState},
		{name: "Route idempotency", value: d.Idempotency},
		{name: "Record envelopes", value: d.RecordEnvelopes},
		{name: "Projections", value: d.Projections},
		{name: "Revisions/history", value: d.Revisions},
		{name: "Conflict fields", value: d.ConflictFields},
		{name: "Keep-saved resolution", value: d.KeepSaved},
		{name: "Collaboration publication", value: d.Collaboration},
	}
	for _, dependency := range required {
		if nilDependency(dependency.value) {
			return fmt.Errorf("parties mutation dependencies: %s is required", dependency.name)
		}
	}
	return nil
}

func nilDependency(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}

type ValidationError struct {
	Field      string
	ReasonCode string
}

func (e *ValidationError) Error() string { return "parties: invalid mutation request" }
