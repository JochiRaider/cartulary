package hostidentity

import (
	"bytes"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/entities/entitycontract"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

var ErrNoEffectivePatchChange = errors.New("entities: no effective patch change")

type PatchRequest struct {
	ViewSchemaID   string
	BaseRowVersion int64
	ClientTxnID    string
	Changes        []PatchChange
}

type PatchChange struct {
	FieldKey          string
	Value             *string
	CollectionActions []CollectionAction
}

type PatchMutationResult struct {
	Payload          map[string]any
	StatusCode       int
	Replayed         bool
	IncidentID       uuid.UUID
	RecordID         uuid.UUID
	ChangeSetID      uuid.UUID
	ClientTxnID      string
	RowVersion       int64
	ViewSchemaID     string
	ChangedFieldKeys []string
}

type RowVersionConflictError struct {
	RecordID          uuid.UUID
	BaseRowVersion    int64
	CurrentRowVersion int64
}

func (e *RowVersionConflictError) Error() string {
	return "entities: row version conflict"
}

func (s *Store) PatchEntityRow(ctx context.Context, actor authn.UserRecord, recordID uuid.UUID, request PatchRequest, requestHash []byte, requestID string, now time.Time, routeKey string) (PatchMutationResult, error) {
	idempotencyKey := authn.RouteIdempotencyKey{
		RouteKey:    routeKey,
		ActorUserID: actor.ID,
		ScopeKey:    recordID.String(),
		ClientTxnID: request.ClientTxnID,
	}
	if existing, err := s.authStore.GetRouteIdempotency(ctx, idempotencyKey); err == nil {
		if !bytes.Equal(existing.RequestHash, requestHash) {
			return PatchMutationResult{}, authn.ErrClientTxnConflict
		}
		payload, err := decodeStoredResponse(existing.ResponseJSON)
		if err != nil {
			return PatchMutationResult{}, fmt.Errorf("decode replayed entity patch payload: %w", err)
		}
		return PatchMutationResult{
			Payload:      payload,
			StatusCode:   http.StatusOK,
			Replayed:     true,
			RecordID:     recordID,
			ViewSchemaID: request.ViewSchemaID,
			ClientTxnID:  request.ClientTxnID,
		}, nil
	} else if !errors.Is(err, authn.ErrNotFound) {
		return PatchMutationResult{}, fmt.Errorf("query entity patch idempotency: %w", err)
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return PatchMutationResult{}, fmt.Errorf("begin entity patch transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	meta, err := loadEntityRecordMetaTx(ctx, tx, recordID, false)
	if err != nil {
		return PatchMutationResult{}, err
	}
	if !entityRecordTypeMatchesView(meta.RecordType, request.ViewSchemaID) {
		return PatchMutationResult{}, pgx.ErrNoRows
	}
	if err := s.incidentAccess.RequireOpenTx(ctx, tx, meta.IncidentID); err != nil {
		return PatchMutationResult{}, err
	}
	if meta.RowVersion != request.BaseRowVersion {
		return PatchMutationResult{}, &RowVersionConflictError{
			RecordID:          recordID,
			BaseRowVersion:    request.BaseRowVersion,
			CurrentRowVersion: meta.RowVersion,
		}
	}
	identifierTuples, err := recordIdentifierTuplesTx(ctx, tx, meta.IncidentID, meta.RecordType, recordID)
	if err != nil {
		return PatchMutationResult{}, err
	}
	switch request.ViewSchemaID {
	case entitycontract.HostsViewSchemaID:
		identifierTuples = mergeNormalizedIdentifierTuples(identifierTuples, hostPatchIdentifierTuples(request.Changes))
	case entitycontract.IdentitiesViewSchemaID:
		identifierTuples = mergeNormalizedIdentifierTuples(identifierTuples, identityPatchIdentifierTuples(request.Changes))
	}
	if err := prepareIdentifierMutationTx(ctx, tx, meta.IncidentID, meta.RecordType, recordID, identifierTuples); err != nil {
		return PatchMutationResult{}, err
	}
	lockedMeta, err := loadEntityRecordMetaTx(ctx, tx, recordID, true)
	if err != nil {
		return PatchMutationResult{}, err
	}
	if lockedMeta.IncidentID != meta.IncidentID || lockedMeta.RecordType != meta.RecordType {
		return PatchMutationResult{}, pgx.ErrNoRows
	}
	if lockedMeta.RowVersion != request.BaseRowVersion {
		return PatchMutationResult{}, &RowVersionConflictError{
			RecordID:          recordID,
			BaseRowVersion:    request.BaseRowVersion,
			CurrentRowVersion: lockedMeta.RowVersion,
		}
	}
	meta = lockedMeta

	switch request.ViewSchemaID {
	case entitycontract.HostsViewSchemaID:
		return s.patchHostRowTx(ctx, tx, actor, meta, recordID, request, idempotencyKey, requestHash, requestID, now.UTC())
	case entitycontract.IdentitiesViewSchemaID:
		return s.patchIdentityRowTx(ctx, tx, actor, meta, recordID, request, idempotencyKey, requestHash, requestID, now.UTC())
	default:
		return PatchMutationResult{}, pgx.ErrNoRows
	}
}

func (s *Store) patchHostRowTx(ctx context.Context, tx pgx.Tx, actor authn.UserRecord, meta entityRecordMeta, recordID uuid.UUID, request PatchRequest, idempotencyKey authn.RouteIdempotencyKey, requestHash []byte, requestID string, now time.Time) (PatchMutationResult, error) {
	beforeRecord, err := loadHostByRecordIDTx(ctx, tx, recordID)
	if err != nil {
		return PatchMutationResult{}, err
	}
	if err := hydrateHostRecordTx(ctx, tx, &beforeRecord); err != nil {
		return PatchMutationResult{}, err
	}
	beforeRow := buildHostRow(beforeRecord)
	beforeSnapshot, err := s.ports.revisions.CaptureRecordSnapshotTx(ctx, tx, recordID)
	if err != nil {
		return PatchMutationResult{}, err
	}
	next := beforeRecord
	changedFields := make([]string, 0, len(request.Changes))
	aliasMutations := make([]AliasAppliedMutation, 0)
	for _, change := range request.Changes {
		changed, applied, err := entityFields.applyHostPatch(
			ctx, tx, meta.IncidentID, recordID, actor.ID, now, &next, change,
		)
		if err != nil {
			return PatchMutationResult{}, err
		}
		if changed {
			changedFields = append(changedFields, change.FieldKey)
			aliasMutations = append(aliasMutations, applied...)
		}
	}
	if len(changedFields) == 0 {
		return PatchMutationResult{}, ErrNoEffectivePatchChange
	}

	rowVersion, err := s.ports.records.AdvanceVersionTx(ctx, tx, recordID, actor.ID, now)
	if err != nil {
		return PatchMutationResult{}, err
	}
	next.RowVersion = rowVersion
	next.UpdatedAt = now
	next.UpdatedByUser = actor.ID
	if err := updateHostTx(ctx, tx, next); err != nil {
		return PatchMutationResult{}, err
	}
	if err := s.ports.projections.RefreshHostTx(ctx, tx, next.RecordID); err != nil {
		return PatchMutationResult{}, err
	}
	if err := hydrateHostRecordTx(ctx, tx, &next); err != nil {
		return PatchMutationResult{}, err
	}
	afterRow := buildHostRow(next)

	return s.finishEntityPatchTx(ctx, tx, actor, meta.IncidentID, recordID, "host", request, idempotencyKey, requestHash, requestID, now, &beforeSnapshot, beforeRow, afterRow, rowVersion, changedFields, aliasMutations)
}

func (s *Store) patchIdentityRowTx(ctx context.Context, tx pgx.Tx, actor authn.UserRecord, meta entityRecordMeta, recordID uuid.UUID, request PatchRequest, idempotencyKey authn.RouteIdempotencyKey, requestHash []byte, requestID string, now time.Time) (PatchMutationResult, error) {
	beforeRecord, err := loadIdentityByRecordIDTx(ctx, tx, recordID)
	if err != nil {
		return PatchMutationResult{}, err
	}
	if err := hydrateIdentityRecordTx(ctx, tx, &beforeRecord); err != nil {
		return PatchMutationResult{}, err
	}
	beforeRow := buildIdentityRow(beforeRecord)
	beforeSnapshot, err := s.ports.revisions.CaptureRecordSnapshotTx(ctx, tx, recordID)
	if err != nil {
		return PatchMutationResult{}, err
	}
	next := beforeRecord
	changedFields := make([]string, 0, len(request.Changes))
	aliasMutations := make([]AliasAppliedMutation, 0)
	for _, change := range request.Changes {
		changed, applied, err := entityFields.applyIdentityPatch(
			ctx, tx, meta.IncidentID, recordID, actor.ID, now, &next, change,
		)
		if err != nil {
			return PatchMutationResult{}, err
		}
		if changed {
			changedFields = append(changedFields, change.FieldKey)
			aliasMutations = append(aliasMutations, applied...)
		}
	}
	if len(changedFields) == 0 {
		return PatchMutationResult{}, ErrNoEffectivePatchChange
	}

	rowVersion, err := s.ports.records.AdvanceVersionTx(ctx, tx, recordID, actor.ID, now)
	if err != nil {
		return PatchMutationResult{}, err
	}
	next.RowVersion = rowVersion
	next.UpdatedAt = now
	next.UpdatedByUser = actor.ID
	if err := updateIdentityTx(ctx, tx, next); err != nil {
		return PatchMutationResult{}, err
	}
	if err := s.ports.projections.RefreshIdentityTx(ctx, tx, next.RecordID); err != nil {
		return PatchMutationResult{}, err
	}
	if err := hydrateIdentityRecordTx(ctx, tx, &next); err != nil {
		return PatchMutationResult{}, err
	}
	afterRow := buildIdentityRow(next)

	return s.finishEntityPatchTx(ctx, tx, actor, meta.IncidentID, recordID, "identity", request, idempotencyKey, requestHash, requestID, now, &beforeSnapshot, beforeRow, afterRow, rowVersion, changedFields, aliasMutations)
}

func (s *Store) finishEntityPatchTx(ctx context.Context, tx pgx.Tx, actor authn.UserRecord, incidentID uuid.UUID, recordID uuid.UUID, targetKind string, request PatchRequest, idempotencyKey authn.RouteIdempotencyKey, requestHash []byte, requestID string, now time.Time, beforeSnapshot *revisions.RecordSnapshot, beforeRow map[string]any, afterRow map[string]any, rowVersion int64, changedFields []string, aliasMutations []AliasAppliedMutation) (PatchMutationResult, error) {
	afterSnapshot, err := s.ports.revisions.CaptureRecordSnapshotTx(ctx, tx, recordID)
	if err != nil {
		return PatchMutationResult{}, err
	}
	changeSetID, err := s.ports.revisions.AppendChangeSetTx(ctx, tx, entityChangeSetParams{
		IncidentID:  incidentID,
		ActorUserID: actor.ID,
		Source:      idempotencyKey.RouteKey,
		ClientTxnID: &request.ClientTxnID,
		RequestID:   &requestID,
		CreatedAt:   now,
	})
	if err != nil {
		return PatchMutationResult{}, err
	}
	beforeVersionID := entityVersionID(targetKind, recordID, request.BaseRowVersion)
	afterVersionID := entityVersionID(targetKind, recordID, rowVersion)
	if err := s.ports.revisions.AppendRecordMutationTx(ctx, tx, revisions.AppendRecordMutationParams{
		ChangeSetID:     changeSetID,
		SequenceNo:      1,
		TargetKind:      targetKind,
		RecordID:        recordID,
		OperationKind:   "patch",
		BeforeVersionID: &beforeVersionID,
		AfterVersionID:  &afterVersionID,
		BeforeSnapshot:  beforeSnapshot,
		AfterSnapshot:   &afterSnapshot,
	}); err != nil {
		return PatchMutationResult{}, err
	}
	for index, mutation := range aliasMutations {
		if err := s.ports.revisions.AppendMutationTx(ctx, tx, entityMutationParams{
			ChangeSetID:   changeSetID,
			SequenceNo:    index + 2,
			TargetKind:    "entity_alias",
			TargetID:      mutation.TargetID,
			OperationKind: mutation.OperationKind,
			BeforeValue:   mutation.BeforeValue,
			AfterValue:    mutation.AfterValue,
		}); err != nil {
			return PatchMutationResult{}, err
		}
	}
	if err := s.ports.revisions.AppendLiveRevisionTx(ctx, tx, revisions.LiveRevisionInput{
		ChangeSetID:    changeSetID,
		RecordID:       recordID,
		RowVersion:     rowVersion,
		BeforeSnapshot: beforeSnapshot,
		AfterSnapshot:  &afterSnapshot,
		ConflictFacts:  entityRevisionFacts(beforeRow, afterRow, changedFields),
	}); err != nil {
		return PatchMutationResult{}, err
	}
	if err := s.appendRecordChangedTx(ctx, tx, incidentID, actor.ID, request.ClientTxnID, changeSetID, recordID, rowVersion, 0, now, request.ViewSchemaID, afterRow, changedFields); err != nil {
		return PatchMutationResult{}, err
	}

	payload := buildMutationPayload(request.ViewSchemaID, changeSetID, afterRow)
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, idempotencyKey, nil, requestHash, http.StatusOK, payload); err != nil {
		if authn.IsUniqueViolation(err) {
			return PatchMutationResult{}, authn.ErrClientTxnConflict
		}
		return PatchMutationResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return PatchMutationResult{}, fmt.Errorf("commit entity patch transaction: %w", err)
	}
	return PatchMutationResult{
		Payload:          payload,
		StatusCode:       http.StatusOK,
		IncidentID:       incidentID,
		RecordID:         recordID,
		ChangeSetID:      changeSetID,
		ClientTxnID:      request.ClientTxnID,
		RowVersion:       rowVersion,
		ViewSchemaID:     request.ViewSchemaID,
		ChangedFieldKeys: changedFields,
	}, nil
}

type entityRecordMeta struct {
	IncidentID uuid.UUID
	RecordType string
	RowVersion int64
}

func loadEntityRecordMetaTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, lock bool) (entityRecordMeta, error) {
	var meta entityRecordMeta
	var deletedAt sql.NullTime
	query := `
SELECT incident_id, record_type, row_version, deleted_at
  FROM records
 WHERE record_id = $1
`
	if lock {
		query += " FOR UPDATE"
	}
	err := tx.QueryRow(ctx, query, recordID).Scan(&meta.IncidentID, &meta.RecordType, &meta.RowVersion, &deletedAt)
	if err != nil {
		return entityRecordMeta{}, err
	}
	if deletedAt.Valid {
		return entityRecordMeta{}, revisions.ErrRecordDeletedUseRestore
	}
	return meta, nil
}

func entityRecordTypeMatchesView(recordType string, viewSchemaID string) bool {
	switch viewSchemaID {
	case entitycontract.HostsViewSchemaID:
		return recordType == "host"
	case entitycontract.IdentitiesViewSchemaID:
		return recordType == "identity"
	default:
		return false
	}
}

func stringPointersEqual(left *string, right *string) bool {
	if left == nil || right == nil {
		return left == nil && right == nil
	}
	return *left == *right
}
