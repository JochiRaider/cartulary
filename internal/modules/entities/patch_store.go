package entities

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

	"github.com/JochiRaider/cartulary/internal/modules/incidents"
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
	FieldKey string
	Value    *string
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

	meta, err := loadEntityRecordMetaForUpdateTx(ctx, tx, recordID)
	if err != nil {
		return PatchMutationResult{}, err
	}
	if !entityRecordTypeMatchesView(meta.RecordType, request.ViewSchemaID) {
		return PatchMutationResult{}, pgx.ErrNoRows
	}
	if err := incidents.EnsureIncidentOpenTx(ctx, tx, meta.IncidentID); err != nil {
		return PatchMutationResult{}, err
	}
	if meta.RowVersion != request.BaseRowVersion {
		return PatchMutationResult{}, &RowVersionConflictError{
			RecordID:          recordID,
			BaseRowVersion:    request.BaseRowVersion,
			CurrentRowVersion: meta.RowVersion,
		}
	}

	switch request.ViewSchemaID {
	case HostsViewSchemaID:
		return s.patchHostRowTx(ctx, tx, actor, meta, recordID, request, idempotencyKey, requestHash, requestID, now.UTC())
	case IdentitiesViewSchemaID:
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
	beforeRow := BuildHostRow(beforeRecord)

	next := beforeRecord
	changedFields := make([]string, 0, len(request.Changes))
	for _, change := range request.Changes {
		switch change.FieldKey {
		case "host.display_name":
			if change.Value == nil {
				return PatchMutationResult{}, ErrNoEffectivePatchChange
			}
			if next.DisplayName != *change.Value {
				next.DisplayName = *change.Value
				changedFields = append(changedFields, change.FieldKey)
			}
		case "host.hostname":
			if !stringPointersEqual(next.Hostname, change.Value) {
				next.Hostname = cloneStringPointer(change.Value)
				changedFields = append(changedFields, change.FieldKey)
			}
		case "host.aad_device_id":
			if !stringPointersEqual(next.AADDeviceID, change.Value) {
				next.AADDeviceID = cloneStringPointer(change.Value)
				changedFields = append(changedFields, change.FieldKey)
			}
		case "host.fqdn":
			if !stringPointersEqual(next.FQDN, change.Value) {
				next.FQDN = cloneStringPointer(change.Value)
				changedFields = append(changedFields, change.FieldKey)
			}
		case "host.location":
			if !stringPointersEqual(next.Location, change.Value) {
				next.Location = cloneStringPointer(change.Value)
				changedFields = append(changedFields, change.FieldKey)
			}
		case "host.os_platform":
			if !stringPointersEqual(next.OSPlatform, change.Value) {
				next.OSPlatform = cloneStringPointer(change.Value)
				changedFields = append(changedFields, change.FieldKey)
			}
		case "host.business_owner":
			if !stringPointersEqual(next.BusinessOwner, change.Value) {
				next.BusinessOwner = cloneStringPointer(change.Value)
				changedFields = append(changedFields, change.FieldKey)
			}
		case "host.criticality":
			if !stringPointersEqual(next.Criticality, change.Value) {
				next.Criticality = cloneStringPointer(change.Value)
				changedFields = append(changedFields, change.FieldKey)
			}
		case "host.containment_status":
			if !stringPointersEqual(next.ContainmentStatus, change.Value) {
				next.ContainmentStatus = cloneStringPointer(change.Value)
				changedFields = append(changedFields, change.FieldKey)
			}
		default:
			return PatchMutationResult{}, ErrNoEffectivePatchChange
		}
	}
	if len(changedFields) == 0 {
		return PatchMutationResult{}, ErrNoEffectivePatchChange
	}

	rowVersion, err := s.recordStore.AdvanceVersionTx(ctx, tx, recordID, actor.ID, now)
	if err != nil {
		return PatchMutationResult{}, err
	}
	next.RowVersion = rowVersion
	next.UpdatedAt = now
	next.UpdatedByUser = actor.ID
	if err := updateHostTx(ctx, tx, next); err != nil {
		return PatchMutationResult{}, err
	}
	if err := upsertHostProjectionTx(ctx, tx, next); err != nil {
		return PatchMutationResult{}, err
	}
	if err := hydrateHostRecordTx(ctx, tx, &next); err != nil {
		return PatchMutationResult{}, err
	}
	afterRow := BuildHostRow(next)

	return s.finishEntityPatchTx(ctx, tx, actor, meta.IncidentID, recordID, "host", request, idempotencyKey, requestHash, requestID, now, beforeRow, afterRow, rowVersion, changedFields)
}

func (s *Store) patchIdentityRowTx(ctx context.Context, tx pgx.Tx, actor authn.UserRecord, meta entityRecordMeta, recordID uuid.UUID, request PatchRequest, idempotencyKey authn.RouteIdempotencyKey, requestHash []byte, requestID string, now time.Time) (PatchMutationResult, error) {
	beforeRecord, err := loadIdentityByRecordIDTx(ctx, tx, recordID)
	if err != nil {
		return PatchMutationResult{}, err
	}
	if err := hydrateIdentityRecordTx(ctx, tx, &beforeRecord); err != nil {
		return PatchMutationResult{}, err
	}
	beforeRow := BuildIdentityRow(beforeRecord)

	next := beforeRecord
	changedFields := make([]string, 0, len(request.Changes))
	for _, change := range request.Changes {
		switch change.FieldKey {
		case "identity.display_name":
			if change.Value == nil {
				return PatchMutationResult{}, ErrNoEffectivePatchChange
			}
			if next.DisplayName != *change.Value {
				next.DisplayName = *change.Value
				changedFields = append(changedFields, change.FieldKey)
			}
		case "identity.aad_object_id":
			if !stringPointersEqual(next.AADObjectID, change.Value) {
				next.AADObjectID = cloneStringPointer(change.Value)
				changedFields = append(changedFields, change.FieldKey)
			}
		case "identity.sid":
			if !stringPointersEqual(next.SID, change.Value) {
				next.SID = cloneStringPointer(change.Value)
				changedFields = append(changedFields, change.FieldKey)
			}
		case "identity.upn":
			if !stringPointersEqual(next.UPN, change.Value) {
				next.UPN = cloneStringPointer(change.Value)
				changedFields = append(changedFields, change.FieldKey)
			}
		case "identity.email":
			if !stringPointersEqual(next.Email, change.Value) {
				next.Email = cloneStringPointer(change.Value)
				changedFields = append(changedFields, change.FieldKey)
			}
		case "identity.sam_account_name":
			if !stringPointersEqual(next.SamAccountName, change.Value) {
				next.SamAccountName = cloneStringPointer(change.Value)
				changedFields = append(changedFields, change.FieldKey)
			}
		case "identity.privilege_level":
			if !stringPointersEqual(next.PrivilegeLevel, change.Value) {
				next.PrivilegeLevel = cloneStringPointer(change.Value)
				changedFields = append(changedFields, change.FieldKey)
			}
		case "identity.mfa_state":
			if !stringPointersEqual(next.MFAState, change.Value) {
				next.MFAState = cloneStringPointer(change.Value)
				changedFields = append(changedFields, change.FieldKey)
			}
		case "identity.reset_status":
			if !stringPointersEqual(next.ResetStatus, change.Value) {
				next.ResetStatus = cloneStringPointer(change.Value)
				changedFields = append(changedFields, change.FieldKey)
			}
		default:
			return PatchMutationResult{}, ErrNoEffectivePatchChange
		}
	}
	if len(changedFields) == 0 {
		return PatchMutationResult{}, ErrNoEffectivePatchChange
	}

	rowVersion, err := s.recordStore.AdvanceVersionTx(ctx, tx, recordID, actor.ID, now)
	if err != nil {
		return PatchMutationResult{}, err
	}
	next.RowVersion = rowVersion
	next.UpdatedAt = now
	next.UpdatedByUser = actor.ID
	if err := updateIdentityTx(ctx, tx, next); err != nil {
		return PatchMutationResult{}, err
	}
	if err := upsertIdentityProjectionTx(ctx, tx, next); err != nil {
		return PatchMutationResult{}, err
	}
	if err := hydrateIdentityRecordTx(ctx, tx, &next); err != nil {
		return PatchMutationResult{}, err
	}
	afterRow := BuildIdentityRow(next)

	return s.finishEntityPatchTx(ctx, tx, actor, meta.IncidentID, recordID, "identity", request, idempotencyKey, requestHash, requestID, now, beforeRow, afterRow, rowVersion, changedFields)
}

func (s *Store) finishEntityPatchTx(ctx context.Context, tx pgx.Tx, actor authn.UserRecord, incidentID uuid.UUID, recordID uuid.UUID, targetKind string, request PatchRequest, idempotencyKey authn.RouteIdempotencyKey, requestHash []byte, requestID string, now time.Time, beforeRow map[string]any, afterRow map[string]any, rowVersion int64, changedFields []string) (PatchMutationResult, error) {
	changeSetID, err := s.revisionsStore.InsertChangeSetTx(ctx, tx, revisions.ChangeSetParams{
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
	if err := s.revisionsStore.InsertMutationTx(ctx, tx, revisions.MutationParams{
		ChangeSetID:     changeSetID,
		SequenceNo:      1,
		TargetKind:      targetKind,
		TargetID:        recordID.String(),
		OperationKind:   "patch",
		BeforeVersionID: &beforeVersionID,
		AfterVersionID:  &afterVersionID,
		BeforeValue:     beforeRow,
		AfterValue:      afterRow,
	}); err != nil {
		return PatchMutationResult{}, err
	}
	if err := s.revisionsStore.InsertRecordRevisionTx(ctx, tx, revisions.RecordRevisionParams{
		ChangeSetID: changeSetID,
		RecordID:    recordID,
		RowVersion:  rowVersion,
		BeforeValue: beforeRow,
		AfterValue:  afterRow,
	}); err != nil {
		return PatchMutationResult{}, err
	}

	payload := BuildMutationPayload(request.ViewSchemaID, changeSetID, afterRow)
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

func loadEntityRecordMetaForUpdateTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (entityRecordMeta, error) {
	var meta entityRecordMeta
	var deletedAt sql.NullTime
	err := tx.QueryRow(ctx, `
SELECT incident_id, record_type, row_version, deleted_at
  FROM records
 WHERE record_id = $1
 FOR UPDATE
`, recordID).Scan(&meta.IncidentID, &meta.RecordType, &meta.RowVersion, &deletedAt)
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
	case HostsViewSchemaID:
		return recordType == "host"
	case IdentitiesViewSchemaID:
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
