package revisionassembly

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type transactionRunnerAdapter struct {
	database postgres.DB
}

func (adapter transactionRunnerAdapter) BeginTx(ctx context.Context, options pgx.TxOptions) (pgx.Tx, error) {
	return adapter.database.BeginTx(ctx, options)
}

type commandAuthorizerAdapter struct {
	access *admission.Checker
}

func (adapter commandAuthorizerAdapter) AuthorizeCommandTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	actor revisions.ActorID,
	kind revisions.CommandKind,
) error {
	roles := admission.RolesReviewerAdmin
	if kind == revisions.CommandSoftDelete {
		roles = admission.RolesEditorReviewerAdmin
	}
	_, err := adapter.access.CheckTx(ctx, tx, incidentID, actor.UUID(), admission.Requirement{
		AllowedRoles: roles,
		Lifecycle:    admission.LifecycleOpen,
	})
	switch {
	case admission.IsDenied(err, admission.DenialIncidentClosed):
		return revisions.ErrCommandIncidentClosed
	case admission.IsDenied(err, admission.DenialNotVisible):
		return revisions.ErrCommandTargetNotFound
	case admission.IsDenied(err, admission.DenialInsufficientRole):
		return revisions.ErrCommandRoleDenied
	default:
		return err
	}
}

type commandIdempotencyAdapter struct {
	store *authn.Store
}

func (adapter commandIdempotencyAdapter) Get(ctx context.Context, key revisions.IdempotencyKey) (revisions.IdempotencyRecord, error) {
	record, err := adapter.store.GetRouteIdempotency(ctx, authn.RouteIdempotencyKey{
		RouteKey:    key.RouteKey,
		ActorUserID: key.ActorID.UUID(),
		ScopeKey:    key.ScopeKey,
		ClientTxnID: key.ClientTxnID,
	})
	if errors.Is(err, authn.ErrNotFound) {
		return revisions.IdempotencyRecord{}, revisions.ErrIdempotencyNotFound
	}
	if err != nil {
		return revisions.IdempotencyRecord{}, err
	}
	return revisions.IdempotencyRecord{RequestHash: record.RequestHash, ResponseJSON: record.ResponseJSON}, nil
}

func (adapter commandIdempotencyAdapter) PutSuccessTx(
	ctx context.Context,
	tx pgx.Tx,
	key revisions.IdempotencyKey,
	requestHash []byte,
	payload map[string]any,
) error {
	err := authn.InsertRouteIdempotencyPayload(ctx, tx, authn.RouteIdempotencyKey{
		RouteKey:    key.RouteKey,
		ActorUserID: key.ActorID.UUID(),
		ScopeKey:    key.ScopeKey,
		ClientTxnID: key.ClientTxnID,
	}, nil, requestHash, http.StatusOK, payload)
	if authn.IsUniqueViolation(err) {
		return revisions.ErrClientTxnConflict
	}
	return err
}

type recordEnvelopeAdapter struct {
	store *records.Store
}

func (adapter recordEnvelopeAdapter) RecordTypeTx(
	ctx context.Context,
	tx pgx.Tx,
	incidentID uuid.UUID,
	recordID uuid.UUID,
) (string, error) {
	if adapter.store == nil {
		return "", pgx.ErrNoRows
	}
	envelope, err := adapter.store.LoadEnvelopeTx(ctx, tx, recordID, false)
	if err != nil {
		if errors.Is(err, records.ErrEnvelopeNotFound) {
			return "", pgx.ErrNoRows
		}
		return "", err
	}
	if envelope.IncidentID != incidentID {
		return "", pgx.ErrNoRows
	}
	return envelope.RecordType, nil
}

func (adapter recordEnvelopeAdapter) LoadEnvelope(ctx context.Context, recordID uuid.UUID) (revisions.RecordEnvelope, error) {
	envelope, err := adapter.store.LoadEnvelope(ctx, recordID)
	return adaptRecordEnvelope(envelope, err)
}

func (adapter recordEnvelopeAdapter) LoadEnvelopeTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, lock bool) (revisions.RecordEnvelope, error) {
	envelope, err := adapter.store.LoadEnvelopeTx(ctx, tx, recordID, lock)
	return adaptRecordEnvelope(envelope, err)
}

func (adapter recordEnvelopeAdapter) AdvanceVersionTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, actorID uuid.UUID, now time.Time) (int64, error) {
	value, err := adapter.store.AdvanceVersionTx(ctx, tx, recordID, actorID, now)
	return value, adaptRecordEnvelopeError(err)
}

func (adapter recordEnvelopeAdapter) SetDeleteStateTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, actorID uuid.UUID, now time.Time, deleting bool) (int64, error) {
	value, err := adapter.store.SetDeleteStateTx(ctx, tx, recordID, actorID, now, deleting)
	return value, adaptRecordEnvelopeError(err)
}

func (adapter recordEnvelopeAdapter) LockDestructiveRecordsNowaitTx(ctx context.Context, tx pgx.Tx, recordIDs []uuid.UUID) error {
	err := records.LockDestructiveOperationRecordsNowaitTx(ctx, tx, recordIDs)
	var locked *records.DestructiveOperationRecordLockedError
	if errors.As(err, &locked) {
		return &revisions.EnvelopeLockError{RecordID: locked.RecordID}
	}
	return adaptRecordEnvelopeError(err)
}

func adaptRecordEnvelope(envelope records.Envelope, err error) (revisions.RecordEnvelope, error) {
	if err != nil {
		return revisions.RecordEnvelope{}, adaptRecordEnvelopeError(err)
	}
	return revisions.RecordEnvelope{
		RecordID:        envelope.RecordID,
		IncidentID:      envelope.IncidentID,
		RecordType:      envelope.RecordType,
		RowVersion:      envelope.RowVersion,
		CreatedByUserID: envelope.CreatedByUserID,
		CreatedAt:       envelope.CreatedAt,
		UpdatedByUserID: envelope.UpdatedByUserID,
		UpdatedAt:       envelope.UpdatedAt,
		DeletedAt:       envelope.DeletedAt,
		DeletedByUserID: envelope.DeletedByUserID,
	}, nil
}

func adaptRecordEnvelopeError(err error) error {
	if errors.Is(err, records.ErrEnvelopeNotFound) {
		return revisions.ErrEnvelopeNotFound
	}
	return err
}
