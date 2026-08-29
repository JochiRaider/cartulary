package evidence

import (
	"context"
	"fmt"
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	evidenceprojection "github.com/JochiRaider/cartulary/internal/modules/evidence/projectionports"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
)

type IncidentStateCapability interface {
	RequireOpenTx(context.Context, pgx.Tx, uuid.UUID) error
}

type RecordEnvelopeCapability interface {
	InsertTx(context.Context, pgx.Tx, records.InsertParams) (uuid.UUID, error)
	AdvanceVersionTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, time.Time) (int64, error)
	LoadEnvelopeTx(context.Context, pgx.Tx, uuid.UUID, bool) (records.Envelope, error)
}

type RevisionCapability interface {
	CaptureRecordSnapshotTx(context.Context, pgx.Tx, uuid.UUID) (revisions.RecordSnapshot, error)
	AppendChangeSetTx(context.Context, pgx.Tx, revisions.AppendChangeSetParams) (uuid.UUID, error)
	AppendRecordMutationTx(context.Context, pgx.Tx, revisions.AppendRecordMutationParams) error
	AppendLiveRevisionTx(context.Context, pgx.Tx, revisions.LiveRevisionInput) error
	LoadRevisionWindowTx(context.Context, pgx.Tx, uuid.UUID, int64, int64) ([]conflicttokens.RevisionWindowRow, error)
}

// MutationDependencies is assembled at the application composition root. The
// Evidence mutation facade never constructs peer-owner stores.
type MutationDependencies struct {
	IncidentState        IncidentStateCapability
	Idempotency          IdempotencyCapability
	LifecycleIdempotency LifecycleIdempotencyCapability
	RecordEnvelopes      RecordEnvelopeCapability
	Revisions            RevisionCapability
	ProjectionRows       evidenceprojection.MutationRows
	AssociationEffects   evidenceprojection.AssociationEffects
	ConflictFields       conflicttokens.FieldResolver
	KeepSavedIdempotency conflicttokens.IdempotencyPort
	Collaboration        collaboration.RecordChangedAppender
}

func (dependencies MutationDependencies) validate() error {
	required := []struct {
		name  string
		value any
	}{
		{name: "Incident admission", value: dependencies.IncidentState},
		{name: "Route idempotency", value: dependencies.Idempotency},
		{name: "Lifecycle idempotency", value: dependencies.LifecycleIdempotency},
		{name: "Record envelopes", value: dependencies.RecordEnvelopes},
		{name: "Revisions/history", value: dependencies.Revisions},
		{name: "Projection rows", value: dependencies.ProjectionRows},
		{name: "Association effects", value: dependencies.AssociationEffects},
		{name: "Conflict fields", value: dependencies.ConflictFields},
		{name: "Keep-saved idempotency", value: dependencies.KeepSavedIdempotency},
		{name: "Collaboration publication", value: dependencies.Collaboration},
	}
	for _, dependency := range required {
		if isNilMutationCapability(dependency.value) {
			return fmt.Errorf("evidence mutation dependencies: %s is required", dependency.name)
		}
	}
	return nil
}

func isNilMutationCapability(value any) bool {
	if value == nil {
		return true
	}
	candidate := reflect.ValueOf(value)
	switch candidate.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return candidate.IsNil()
	default:
		return false
	}
}
