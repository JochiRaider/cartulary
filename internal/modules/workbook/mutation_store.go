package workbook

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

func (s *Store) CreateWorkbookRow(ctx context.Context, actor authn.UserRecord, incidentID uuid.UUID, request CreateRequest, requestHash []byte, requestID string, now time.Time) (MutationResult, error) {
	scopeKey := incidentID.String() + ":" + request.ViewSchemaID
	idempotencyKey := authn.RouteIdempotencyKey{
		RouteKey:    workbookCreateRouteKey,
		ActorUserID: actor.ID,
		ScopeKey:    scopeKey,
		ClientTxnID: request.ClientTxnID,
	}
	if existing, err := s.authStore.GetRouteIdempotency(ctx, idempotencyKey); err == nil {
		if !bytes.Equal(existing.RequestHash, requestHash) {
			return MutationResult{}, authn.ErrClientTxnConflict
		}
		payload, err := decodeStoredResponse(existing.ResponseJSON)
		if err != nil {
			return MutationResult{}, fmt.Errorf("decode replayed workbook create payload: %w", err)
		}
		recordID, err := extractPayloadUUID(payload, "row", "record_id")
		if err != nil {
			return MutationResult{}, err
		}
		return MutationResult{Payload: payload, StatusCode: http.StatusOK, Replayed: true, IncidentID: incidentID, RecordID: recordID, ViewSchemaID: request.ViewSchemaID, ClientTxnID: request.ClientTxnID}, nil
	} else if !errors.Is(err, authn.ErrNotFound) {
		return MutationResult{}, fmt.Errorf("query workbook create idempotency: %w", err)
	}
	if err := validateCreateRequest(request); err != nil {
		return MutationResult{}, err
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return MutationResult{}, fmt.Errorf("begin workbook create transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if err := validateCreateReferencesTx(ctx, tx, incidentID, request); err != nil {
		return MutationResult{}, err
	}

	recordType := recordTypeForView(request.ViewSchemaID)
	if recordType == "" {
		return MutationResult{}, mutationValidationError("view_schema_id", "unknown_view_schema")
	}
	recordID, err := s.recordStore.InsertTx(ctx, tx, recordsInsertParams(incidentID, recordType, actor.ID, now.UTC()))
	if err != nil {
		return MutationResult{}, err
	}
	switch request.ViewSchemaID {
	case EvidenceViewSchemaID:
		if err := insertEvidenceTx(ctx, tx, recordID, incidentID, request, now.UTC()); err != nil {
			return MutationResult{}, err
		}
	case PartiesViewSchemaID:
		if err := insertPartyTx(ctx, tx, recordID, incidentID, request, now.UTC()); err != nil {
			return MutationResult{}, err
		}
	case TaskRequestsViewSchemaID:
		if err := insertTaskRequestTx(ctx, tx, recordID, incidentID, actor.ID, request, now.UTC()); err != nil {
			return MutationResult{}, err
		}
	case DecisionsViewSchemaID:
		if err := insertDecisionTx(ctx, tx, recordID, incidentID, actor.ID, request, now.UTC()); err != nil {
			return MutationResult{}, err
		}
	default:
		if err := insertArtifactTx(ctx, tx, recordID, incidentID, actor.ID, request, now.UTC()); err != nil {
			return MutationResult{}, err
		}
	}
	if err := applyCollectionPayloadsTx(ctx, tx, incidentID, recordID, actor.ID, request.Collections, now.UTC()); err != nil {
		return MutationResult{}, err
	}

	row, err := s.loadGenericRowTx(ctx, tx, request.ViewSchemaID, recordID)
	if err != nil {
		return MutationResult{}, err
	}
	changeSetID, err := s.revisionStore.InsertChangeSetTx(ctx, tx, revisions.ChangeSetParams{
		IncidentID:  incidentID,
		ActorUserID: actor.ID,
		Source:      workbookCreateRouteKey,
		ClientTxnID: &request.ClientTxnID,
		RequestID:   &requestID,
		CreatedAt:   now.UTC(),
	})
	if err != nil {
		return MutationResult{}, err
	}
	afterVersionID := workbookVersionID(recordID, 1)
	if err := s.revisionStore.InsertMutationTx(ctx, tx, revisions.MutationParams{
		ChangeSetID:    changeSetID,
		SequenceNo:     1,
		TargetKind:     "record",
		TargetID:       recordID.String(),
		OperationKind:  "create",
		AfterVersionID: &afterVersionID,
		AfterValue:     row,
	}); err != nil {
		return MutationResult{}, err
	}
	if err := s.revisionStore.InsertRecordRevisionTx(ctx, tx, revisions.RecordRevisionParams{
		ChangeSetID: changeSetID,
		RecordID:    recordID,
		RowVersion:  1,
		AfterValue:  row,
	}); err != nil {
		return MutationResult{}, err
	}
	payload := BuildMutationPayload(request.ViewSchemaID, changeSetID, row)
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, idempotencyKey, nil, requestHash, http.StatusCreated, payload); err != nil {
		if authn.IsUniqueViolation(err) {
			return MutationResult{}, authn.ErrClientTxnConflict
		}
		return MutationResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return MutationResult{}, fmt.Errorf("commit workbook create transaction: %w", err)
	}
	return MutationResult{
		Payload:          payload,
		StatusCode:       http.StatusCreated,
		IncidentID:       incidentID,
		RecordID:         recordID,
		ChangeSetID:      changeSetID,
		ClientTxnID:      request.ClientTxnID,
		RowVersion:       1,
		ViewSchemaID:     request.ViewSchemaID,
		ChangedFieldKeys: changedFieldKeys(nil, row),
	}, nil
}

func (s *Store) CreateLinkedNote(ctx context.Context, actor authn.UserRecord, sourceRecordID uuid.UUID, request LinkedNoteCreateRequest, requestHash []byte, requestID string, now time.Time) (MutationResult, error) {
	idempotencyKey := authn.RouteIdempotencyKey{
		RouteKey:    workbookLinkedNoteRouteKey,
		ActorUserID: actor.ID,
		ScopeKey:    sourceRecordID.String(),
		ClientTxnID: request.ClientTxnID,
	}
	if existing, err := s.authStore.GetRouteIdempotency(ctx, idempotencyKey); err == nil {
		if !bytes.Equal(existing.RequestHash, requestHash) {
			return MutationResult{}, authn.ErrClientTxnConflict
		}
		payload, err := decodeStoredResponse(existing.ResponseJSON)
		if err != nil {
			return MutationResult{}, fmt.Errorf("decode replayed linked note payload: %w", err)
		}
		recordID, err := extractPayloadUUID(payload, "row", "record_id")
		if err != nil {
			return MutationResult{}, err
		}
		return MutationResult{Payload: payload, StatusCode: http.StatusOK, Replayed: true, RecordID: recordID, ViewSchemaID: NotesViewSchemaID, ClientTxnID: request.ClientTxnID}, nil
	} else if !errors.Is(err, authn.ErrNotFound) {
		return MutationResult{}, fmt.Errorf("query linked note idempotency: %w", err)
	}
	create := CreateRequest{
		ViewSchemaID: NotesViewSchemaID,
		ClientTxnID:  request.ClientTxnID,
		Values:       request.Values,
		Collections:  request.Collections,
	}
	if err := validateCreateRequest(create); err != nil {
		return MutationResult{}, err
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return MutationResult{}, fmt.Errorf("begin linked note transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	incidentID, err := loadLinkedNoteSourceIncidentTx(ctx, tx, sourceRecordID)
	if err != nil {
		return MutationResult{}, err
	}
	recordID, err := s.recordStore.InsertTx(ctx, tx, recordsInsertParams(incidentID, "artifact", actor.ID, now.UTC()))
	if err != nil {
		return MutationResult{}, err
	}
	if err := insertArtifactTx(ctx, tx, recordID, incidentID, actor.ID, create, now.UTC()); err != nil {
		return MutationResult{}, err
	}
	if err := applyCollectionPayloadsTx(ctx, tx, incidentID, recordID, actor.ID, create.Collections, now.UTC()); err != nil {
		return MutationResult{}, err
	}
	if err := insertLinkedNoteReferenceTx(ctx, tx, incidentID, sourceRecordID, recordID, actor.ID, now.UTC()); err != nil {
		return MutationResult{}, err
	}
	row, err := s.loadGenericRowTx(ctx, tx, NotesViewSchemaID, recordID)
	if err != nil {
		return MutationResult{}, err
	}
	changeSetID, err := s.revisionStore.InsertChangeSetTx(ctx, tx, revisions.ChangeSetParams{
		IncidentID:  incidentID,
		ActorUserID: actor.ID,
		Source:      workbookLinkedNoteRouteKey,
		ClientTxnID: &request.ClientTxnID,
		RequestID:   &requestID,
		CreatedAt:   now.UTC(),
	})
	if err != nil {
		return MutationResult{}, err
	}
	afterVersionID := workbookVersionID(recordID, 1)
	if err := s.revisionStore.InsertMutationTx(ctx, tx, revisions.MutationParams{
		ChangeSetID:    changeSetID,
		SequenceNo:     1,
		TargetKind:     "record",
		TargetID:       recordID.String(),
		OperationKind:  "create",
		AfterVersionID: &afterVersionID,
		AfterValue:     row,
	}); err != nil {
		return MutationResult{}, err
	}
	linkAfter := map[string]any{
		"src_record_id": sourceRecordID.String(),
		"dst_record_id": recordID.String(),
		"link_type":     "references_artifact",
	}
	if err := s.revisionStore.InsertMutationTx(ctx, tx, revisions.MutationParams{
		ChangeSetID:   changeSetID,
		SequenceNo:    2,
		TargetKind:    "record_link",
		TargetID:      sourceRecordID.String() + ":references_artifact:" + recordID.String(),
		OperationKind: "create",
		AfterValue:    linkAfter,
	}); err != nil {
		return MutationResult{}, err
	}
	if err := s.revisionStore.InsertRecordRevisionTx(ctx, tx, revisions.RecordRevisionParams{
		ChangeSetID: changeSetID,
		RecordID:    recordID,
		RowVersion:  1,
		AfterValue:  row,
	}); err != nil {
		return MutationResult{}, err
	}
	payload := BuildMutationPayload(NotesViewSchemaID, changeSetID, row)
	payload["source_record_id"] = sourceRecordID.String()
	payload["link_type"] = "references_artifact"
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, idempotencyKey, nil, requestHash, http.StatusCreated, payload); err != nil {
		if authn.IsUniqueViolation(err) {
			return MutationResult{}, authn.ErrClientTxnConflict
		}
		return MutationResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return MutationResult{}, fmt.Errorf("commit linked note transaction: %w", err)
	}
	return MutationResult{
		Payload:          payload,
		StatusCode:       http.StatusCreated,
		IncidentID:       incidentID,
		RecordID:         recordID,
		ChangeSetID:      changeSetID,
		ClientTxnID:      request.ClientTxnID,
		RowVersion:       1,
		ViewSchemaID:     NotesViewSchemaID,
		ChangedFieldKeys: changedFieldKeys(nil, row),
	}, nil
}

func (s *Store) LinkedNoteSourceIncident(ctx context.Context, sourceRecordID uuid.UUID) (uuid.UUID, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return uuid.UUID{}, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	incidentID, err := loadLinkedNoteSourceIncidentTx(ctx, tx, sourceRecordID)
	if err != nil {
		return uuid.UUID{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return uuid.UUID{}, err
	}
	return incidentID, nil
}

func (s *Store) PatchWorkbookRow(ctx context.Context, actor authn.UserRecord, recordID uuid.UUID, request PatchRequest, requestHash []byte, requestID string, now time.Time) (MutationResult, error) {
	if len(request.Changes) == 0 {
		return MutationResult{}, mutationValidationError("changes", "empty_changes")
	}
	return s.applyWorkbookPatch(ctx, actor, recordID, request, requestHash, requestID, now, workbookPatchRouteKey)
}

func (s *Store) applyWorkbookPatch(ctx context.Context, actor authn.UserRecord, recordID uuid.UUID, request PatchRequest, requestHash []byte, requestID string, now time.Time, routeKey string) (MutationResult, error) {
	idempotencyKey := authn.RouteIdempotencyKey{
		RouteKey:    routeKey,
		ActorUserID: actor.ID,
		ScopeKey:    recordID.String(),
		ClientTxnID: request.ClientTxnID,
	}
	if existing, err := s.authStore.GetRouteIdempotency(ctx, idempotencyKey); err == nil {
		if !bytes.Equal(existing.RequestHash, requestHash) {
			return MutationResult{}, authn.ErrClientTxnConflict
		}
		payload, err := decodeStoredResponse(existing.ResponseJSON)
		if err != nil {
			return MutationResult{}, fmt.Errorf("decode replayed workbook patch payload: %w", err)
		}
		return MutationResult{Payload: payload, StatusCode: http.StatusOK, Replayed: true, RecordID: recordID, ViewSchemaID: request.ViewSchemaID, ClientTxnID: request.ClientTxnID}, nil
	} else if !errors.Is(err, authn.ErrNotFound) {
		return MutationResult{}, fmt.Errorf("query workbook patch idempotency: %w", err)
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return MutationResult{}, fmt.Errorf("begin workbook patch transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	meta, err := loadRecordMetaForUpdateTx(ctx, tx, recordID)
	if err != nil {
		return MutationResult{}, err
	}
	if !recordTypeMatchesView(meta.RecordType, request.ViewSchemaID) {
		return MutationResult{}, pgx.ErrNoRows
	}
	effectiveBeforeVersion := request.BaseRowVersion
	if meta.RowVersion != request.BaseRowVersion {
		if meta.RowVersion < request.BaseRowVersion {
			return MutationResult{}, &RowVersionConflictError{RecordID: recordID, BaseRowVersion: request.BaseRowVersion, CurrentRowVersion: meta.RowVersion}
		}
		window, err := loadWorkbookPatchConflictWindowTx(ctx, tx, recordID, request.ViewSchemaID, request.BaseRowVersion, meta.RowVersion)
		if err != nil {
			return MutationResult{}, err
		}
		if change, changed, ok := overlappingWorkbookPatchChange(request.Changes, window.ChangedFields); ok {
			current, err := s.loadGenericRowTx(ctx, tx, request.ViewSchemaID, recordID)
			if err != nil {
				return MutationResult{}, err
			}
			conflict, err := buildWorkbookSameFieldConflict(recordID, request.ViewSchemaID, request.BaseRowVersion, meta.RowVersion, requestHash, window, change, changed, current)
			if err != nil {
				return MutationResult{}, err
			}
			return MutationResult{}, conflict
		}
		effectiveBeforeVersion = meta.RowVersion
	}
	beforeRow, err := s.loadGenericRowTx(ctx, tx, request.ViewSchemaID, recordID)
	if err != nil {
		return MutationResult{}, err
	}
	if err := validatePatchReferencesTx(ctx, tx, meta.IncidentID, request); err != nil {
		return MutationResult{}, err
	}
	if err := validatePatchLifecycleTx(ctx, tx, recordID, request); err != nil {
		return MutationResult{}, err
	}
	changed, err := applyPatchTx(ctx, tx, meta.IncidentID, recordID, actor.ID, request, now.UTC())
	if err != nil {
		return MutationResult{}, err
	}
	if !changed {
		return MutationResult{}, mutationValidationError("changes", "no_effective_change")
	}
	rowVersion, err := s.recordStore.AdvanceVersionTx(ctx, tx, recordID, actor.ID, now.UTC())
	if err != nil {
		return MutationResult{}, err
	}
	if err := touchSourceRowTx(ctx, tx, request.ViewSchemaID, recordID, now.UTC()); err != nil {
		return MutationResult{}, err
	}
	afterRow, err := s.loadGenericRowTx(ctx, tx, request.ViewSchemaID, recordID)
	if err != nil {
		return MutationResult{}, err
	}
	changeSetID, err := s.revisionStore.InsertChangeSetTx(ctx, tx, revisions.ChangeSetParams{
		IncidentID:  meta.IncidentID,
		ActorUserID: actor.ID,
		Source:      routeKey,
		ClientTxnID: &request.ClientTxnID,
		RequestID:   &requestID,
		CreatedAt:   now.UTC(),
	})
	if err != nil {
		return MutationResult{}, err
	}
	beforeVersionID := workbookVersionID(recordID, request.BaseRowVersion)
	if effectiveBeforeVersion != request.BaseRowVersion {
		beforeVersionID = workbookVersionID(recordID, effectiveBeforeVersion)
	}
	afterVersionID := workbookVersionID(recordID, rowVersion)
	if err := s.revisionStore.InsertMutationTx(ctx, tx, revisions.MutationParams{
		ChangeSetID:     changeSetID,
		SequenceNo:      1,
		TargetKind:      "record",
		TargetID:        recordID.String(),
		OperationKind:   "patch",
		BeforeVersionID: &beforeVersionID,
		AfterVersionID:  &afterVersionID,
		BeforeValue:     beforeRow,
		AfterValue:      afterRow,
	}); err != nil {
		return MutationResult{}, err
	}
	if err := s.revisionStore.InsertRecordRevisionTx(ctx, tx, revisions.RecordRevisionParams{
		ChangeSetID: changeSetID,
		RecordID:    recordID,
		RowVersion:  rowVersion,
		BeforeValue: beforeRow,
		AfterValue:  afterRow,
	}); err != nil {
		return MutationResult{}, err
	}
	payload := BuildMutationPayload(request.ViewSchemaID, changeSetID, afterRow)
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, idempotencyKey, nil, requestHash, http.StatusOK, payload); err != nil {
		if authn.IsUniqueViolation(err) {
			return MutationResult{}, authn.ErrClientTxnConflict
		}
		return MutationResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return MutationResult{}, fmt.Errorf("commit workbook patch transaction: %w", err)
	}
	return MutationResult{
		Payload:          payload,
		StatusCode:       http.StatusOK,
		IncidentID:       meta.IncidentID,
		RecordID:         recordID,
		ChangeSetID:      changeSetID,
		ClientTxnID:      request.ClientTxnID,
		RowVersion:       rowVersion,
		ViewSchemaID:     request.ViewSchemaID,
		ChangedFieldKeys: changedFieldKeys(beforeRow, afterRow),
	}, nil
}

func (s *Store) ResolveWorkbookConflict(ctx context.Context, actor authn.UserRecord, recordID uuid.UUID, claims workbookConflictTokenClaims, request ConflictResolveRequest, requestHash []byte, requestID string, now time.Time) (MutationResult, error) {
	if request.ResolutionKind == "keep_saved" {
		return s.clearWorkbookConflict(ctx, actor, recordID, claims, request, requestHash)
	}
	if request.ResolvedChange == nil {
		return MutationResult{}, mutationValidationError("resolved_value", "missing_required_field")
	}
	patch := PatchRequest{
		ViewSchemaID:   claims.ViewSchemaID,
		BaseRowVersion: claims.CurrentRowVersion,
		ClientTxnID:    request.ClientTxnID,
		Changes:        []PatchChange{*request.ResolvedChange},
	}
	return s.applyWorkbookPatch(ctx, actor, recordID, patch, requestHash, requestID, now, workbookConflictResolveRouteKey)
}

func (s *Store) clearWorkbookConflict(ctx context.Context, actor authn.UserRecord, recordID uuid.UUID, claims workbookConflictTokenClaims, request ConflictResolveRequest, requestHash []byte) (MutationResult, error) {
	idempotencyKey := authn.RouteIdempotencyKey{
		RouteKey:    workbookConflictResolveRouteKey,
		ActorUserID: actor.ID,
		ScopeKey:    recordID.String(),
		ClientTxnID: request.ClientTxnID,
	}
	if existing, err := s.authStore.GetRouteIdempotency(ctx, idempotencyKey); err == nil {
		if !bytes.Equal(existing.RequestHash, requestHash) {
			return MutationResult{}, authn.ErrClientTxnConflict
		}
		payload, err := decodeStoredResponse(existing.ResponseJSON)
		if err != nil {
			return MutationResult{}, fmt.Errorf("decode replayed workbook conflict clear payload: %w", err)
		}
		return MutationResult{Payload: payload, StatusCode: http.StatusOK, Replayed: true, RecordID: recordID, ViewSchemaID: claims.ViewSchemaID, ClientTxnID: request.ClientTxnID}, nil
	} else if !errors.Is(err, authn.ErrNotFound) {
		return MutationResult{}, fmt.Errorf("query workbook conflict clear idempotency: %w", err)
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return MutationResult{}, fmt.Errorf("begin workbook conflict clear transaction: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	meta, err := loadRecordMetaForUpdateTx(ctx, tx, recordID)
	if err != nil {
		return MutationResult{}, err
	}
	if !recordTypeMatchesView(meta.RecordType, claims.ViewSchemaID) {
		return MutationResult{}, pgx.ErrNoRows
	}
	row, err := s.loadGenericRowTx(ctx, tx, claims.ViewSchemaID, recordID)
	if err != nil {
		return MutationResult{}, err
	}
	payload := map[string]any{
		"view_schema_id": claims.ViewSchemaID,
		"row":            row,
	}
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, idempotencyKey, nil, requestHash, http.StatusOK, payload); err != nil {
		if authn.IsUniqueViolation(err) {
			return MutationResult{}, authn.ErrClientTxnConflict
		}
		return MutationResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return MutationResult{}, fmt.Errorf("commit workbook conflict clear transaction: %w", err)
	}
	return MutationResult{
		Payload:      payload,
		StatusCode:   http.StatusOK,
		IncidentID:   meta.IncidentID,
		RecordID:     recordID,
		ClientTxnID:  request.ClientTxnID,
		RowVersion:   meta.RowVersion,
		ViewSchemaID: claims.ViewSchemaID,
	}, nil
}

func (s *Store) RecordIncident(ctx context.Context, recordID uuid.UUID, viewSchemaID string) (uuid.UUID, error) {
	var incidentID uuid.UUID
	if viewSchemaID == "cartulary.view.timeline.v1" {
		err := s.pool.QueryRow(ctx, `
SELECT incident_id
  FROM records
 WHERE record_id = $1
   AND record_type = 'timeline_event'
`, recordID).Scan(&incidentID)
		return incidentID, err
	}
	recordType := recordTypeForView(viewSchemaID)
	switch recordType {
	case "party":
		err := s.pool.QueryRow(ctx, `
SELECT r.incident_id
  FROM records r
  JOIN parties p
    ON p.incident_id = r.incident_id
   AND p.record_id = r.record_id
 WHERE r.record_id = $1
   AND r.record_type = 'party'
`, recordID).Scan(&incidentID)
		return incidentID, err
	case "artifact":
		err := s.pool.QueryRow(ctx, `
SELECT r.incident_id
  FROM records r
  JOIN artifacts a
    ON a.incident_id = r.incident_id
   AND a.record_id = r.record_id
 WHERE r.record_id = $1
   AND r.record_type = 'artifact'
   AND a.artifact_type = $2
`, recordID, artifactTypeForView(viewSchemaID)).Scan(&incidentID)
		return incidentID, err
	case "task_request", "decision":
		err := s.pool.QueryRow(ctx, `
SELECT incident_id
  FROM records
 WHERE record_id = $1
   AND record_type = $2
`, recordID, recordType).Scan(&incidentID)
		return incidentID, err
	default:
		err := s.pool.QueryRow(ctx, `
SELECT incident_id
  FROM records
 WHERE record_id = $1
`, recordID).Scan(&incidentID)
		return incidentID, err
	}
}

type recordMeta struct {
	IncidentID uuid.UUID
	RecordType string
	RowVersion int64
}

func recordsInsertParams(incidentID uuid.UUID, recordType string, actorID uuid.UUID, now time.Time) records.InsertParams {
	return records.InsertParams{
		IncidentID:      incidentID,
		RecordType:      recordType,
		CreatedByUserID: actorID,
		CreatedAt:       now,
		UpdatedByUserID: actorID,
		UpdatedAt:       now,
		RowVersion:      1,
	}
}

func loadRecordMetaForUpdateTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (recordMeta, error) {
	var meta recordMeta
	var deletedAt sql.NullTime
	err := tx.QueryRow(ctx, `
SELECT incident_id, record_type, row_version, deleted_at
  FROM records
 WHERE record_id = $1
 FOR UPDATE
`, recordID).Scan(&meta.IncidentID, &meta.RecordType, &meta.RowVersion, &deletedAt)
	if err != nil {
		return recordMeta{}, err
	}
	if deletedAt.Valid {
		return recordMeta{}, revisions.ErrRecordDeletedUseRestore
	}
	return meta, nil
}

func (s *Store) loadGenericRowTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, recordID uuid.UUID) (map[string]any, error) {
	definition, ok := genericSurfaces[viewSchemaID]
	if !ok {
		return nil, fmt.Errorf("workbook mutation surface %q not mapped", viewSchemaID)
	}
	var builder strings.Builder
	builder.WriteString("SELECT ")
	builder.WriteString(definition.recordExpr)
	builder.WriteString(", r.row_version")
	for _, field := range definition.fields {
		builder.WriteString(", ")
		builder.WriteString(field.expr)
	}
	builder.WriteString(" ")
	builder.WriteString(definition.fromSQL)
	builder.WriteString(" WHERE ")
	builder.WriteString(definition.recordExpr)
	builder.WriteString(" = $1 AND r.deleted_at IS NULL")
	if definition.whereSQL != "" {
		builder.WriteString(" AND ")
		builder.WriteString(definition.whereSQL)
	}
	row := tx.QueryRow(ctx, builder.String(), recordID)
	values := make([]any, len(definition.fields)+2)
	scanTargets := make([]any, len(values))
	for index := range values {
		scanTargets[index] = &values[index]
	}
	if err := row.Scan(scanTargets...); err != nil {
		return nil, err
	}
	return buildGenericRow(definition, nil, values)
}

func validateCreateRequest(request CreateRequest) error {
	switch request.ViewSchemaID {
	case NotesViewSchemaID:
		if !hasTextValue(request.Values, "note.title") && !hasTextValue(request.Values, "note.body") {
			return mutationValidationError("payload", "missing_minimum_create_signal")
		}
	case EvidenceViewSchemaID:
		if !hasTextValue(request.Values, "evidence.title") &&
			!hasTextValue(request.Values, "evidence.storage_ref") &&
			!hasTextValue(request.Values, "evidence.collector_party_text") &&
			!hasTextValue(request.Values, "evidence.source_party_text") {
			return mutationValidationError("payload", "missing_minimum_create_signal")
		}
		if value, ok := request.Values["evidence.lifecycle_state"]; ok && !validEvidenceLifecycleState(derefText(value.Text)) {
			return mutationValidationError("evidence.lifecycle_state", "invalid_value")
		}
	case PartiesViewSchemaID:
		if !hasTextValue(request.Values, "party.display_name") {
			return mutationValidationError("party.display_name", "missing_required_field")
		}
		if !validValueText(request.Values, "party.party_kind", validPartyKind) {
			return mutationValidationError("party.party_kind", "missing_required_field")
		}
	case CommLogViewSchemaID:
		for _, field := range []string{"comm_log.comm_type", "comm_log.audience", "comm_log.channel_or_meeting", "comm_log.summary"} {
			if !hasTextValue(request.Values, field) {
				return mutationValidationError(field, "missing_required_field")
			}
		}
		if !validValueText(request.Values, "comm_log.comm_type", validCommType) {
			return mutationValidationError("comm_log.comm_type", "invalid_value")
		}
	case HandoffViewSchemaID:
		if !hasUUIDValue(request.Values, "handoff.incoming_owner_user_id") {
			return mutationValidationError("handoff.incoming_owner_user_id", "missing_required_field")
		}
		if !hasTextValue(request.Values, "handoff.current_state_summary") {
			return mutationValidationError("handoff.current_state_summary", "missing_required_field")
		}
	case StatusReviewViewSchemaID:
		if !hasTextValue(request.Values, "status_review.current_state_summary") {
			return mutationValidationError("status_review.current_state_summary", "missing_required_field")
		}
	case LessonViewSchemaID:
		if !hasTextValue(request.Values, "lesson.summary") {
			return mutationValidationError("lesson.summary", "missing_required_field")
		}
		if value, ok := request.Values["lesson.closure_state"]; ok && !validClosureState(derefText(value.Text)) {
			return mutationValidationError("lesson.closure_state", "invalid_value")
		}
	case TaskRequestsViewSchemaID:
		if !hasTextValue(request.Values, "task.title") {
			return mutationValidationError("task.title", "missing_required_field")
		}
		if !validValueText(request.Values, "task.task_kind", validTaskKind) {
			return mutationValidationError("task.task_kind", "missing_required_field")
		}
		if value, ok := request.Values["task.status"]; ok && !validTaskStatus(derefText(value.Text)) {
			return mutationValidationError("task.status", "invalid_value")
		}
		if value, ok := request.Values["task.priority"]; ok && !validTaskPriority(derefText(value.Text)) {
			return mutationValidationError("task.priority", "invalid_value")
		}
	case DecisionsViewSchemaID:
		if !hasTextValue(request.Values, "decision.summary") {
			return mutationValidationError("decision.summary", "missing_required_field")
		}
		if !validValueText(request.Values, "decision.decision_type", validDecisionType) {
			return mutationValidationError("decision.decision_type", "missing_required_field")
		}
		if !hasTextValue(request.Values, "decision.rationale") {
			return mutationValidationError("decision.rationale", "missing_required_field")
		}
		if value, ok := request.Values["decision.status"]; ok {
			status := derefText(value.Text)
			if !validDecisionStatus(status) {
				return mutationValidationError("decision.status", "invalid_value")
			}
			if status == "superseded" {
				return &LifecycleValidationError{ToStatus: status, ReasonCode: "superseded_direct_write", ViolatedGuards: []string{"decision.status"}}
			}
		}
	}
	return nil
}

func validatePatchLifecycleTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, request PatchRequest) error {
	if request.ViewSchemaID != EvidenceViewSchemaID {
		return nil
	}
	var from string
	var objectBlobID sql.NullString
	var uploadState sql.NullString
	if err := tx.QueryRow(ctx, `
SELECT e.lifecycle_state, e.object_blob_id::text, b.upload_state
  FROM evidence e
  LEFT JOIN object_blobs b ON b.object_blob_id = e.object_blob_id
 WHERE e.record_id = $1
`, recordID).Scan(&from, &objectBlobID, &uploadState); err != nil {
		return err
	}
	to := from
	for _, change := range request.Changes {
		if change.FieldKey == "evidence.lifecycle_state" && change.Value != nil && change.Value.Text != nil {
			to = *change.Value.Text
		}
	}
	if to != from && !legalEvidenceLifecycleTransition(from, to) {
		return &LifecycleValidationError{FromStatus: from, ToStatus: to, ReasonCode: "illegal_status_transition", ViolatedGuards: []string{"evidence.lifecycle_state"}}
	}
	linkedBlobState := ""
	if uploadState.Valid {
		linkedBlobState = uploadState.String
	}
	if violatesEvidenceBlobBridge(to, objectBlobID.Valid, linkedBlobState) {
		return &LifecycleValidationError{FromStatus: from, ToStatus: to, ReasonCode: "violated_lifecycle_guards", ViolatedGuards: []string{"evidence.lifecycle_state", "object_blobs.upload_state"}}
	}
	return nil
}

func legalEvidenceLifecycleTransition(from string, to string) bool {
	if from == to {
		return true
	}
	switch from {
	case "requested":
		return to == "pending_receipt" || to == "received" || to == "available"
	case "pending_receipt":
		return to == "requested" || to == "received" || to == "available"
	case "received":
		return to == "pending_receipt" || to == "available" || to == "quarantined"
	case "available":
		return to == "received" || to == "quarantined" || to == "released"
	case "quarantined":
		return to == "received" || to == "available"
	case "released":
		return to == "available" || to == "quarantined"
	default:
		return false
	}
}

func violatesEvidenceBlobBridge(lifecycleState string, hasBlob bool, uploadState string) bool {
	switch lifecycleState {
	case "available", "released":
		return !hasBlob || uploadState != "available"
	case "quarantined":
		return hasBlob && uploadState != "quarantined"
	default:
		return false
	}
}

func insertEvidenceTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, incidentID uuid.UUID, request CreateRequest, now time.Time) error {
	lifecycleState := nullableTextValue(request.Values, "evidence.lifecycle_state")
	if lifecycleState == nil {
		lifecycleState = "requested"
	}
	requestedAt := nullableTimestampValue(request.Values, "evidence.requested_at")
	if requestedAt == nil && lifecycleState == "requested" {
		requestedAt = now
	}
	_, err := tx.Exec(ctx, `
INSERT INTO evidence (
    record_id, incident_id, title, lifecycle_state, requested_at, received_at,
    storage_ref, collector_party_text, collector_party_id, source_party_text,
    source_party_id, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6,
    $7, $8, $9, $10,
    $11, $12, $12
)
`, recordID, incidentID,
		nullableTextValue(request.Values, "evidence.title"),
		lifecycleState,
		requestedAt,
		nullableTimestampValue(request.Values, "evidence.received_at"),
		nullableTextValue(request.Values, "evidence.storage_ref"),
		nullableTextValue(request.Values, "evidence.collector_party_text"),
		nullableUUIDValue(request.Values, "evidence.collector_party_id"),
		nullableTextValue(request.Values, "evidence.source_party_text"),
		nullableUUIDValue(request.Values, "evidence.source_party_id"),
		now)
	if err != nil {
		return fmt.Errorf("insert evidence: %w", err)
	}
	return nil
}

func insertPartyTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, incidentID uuid.UUID, request CreateRequest, now time.Time) error {
	_, err := tx.Exec(ctx, `
INSERT INTO parties (
    record_id, incident_id, display_name, party_kind, organization_name, role_title,
    primary_email, timezone_name, external_ref, notes, created_at, updated_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $11)
`, recordID, incidentID,
		textValue(request.Values, "party.display_name"),
		textValue(request.Values, "party.party_kind"),
		nullableTextValue(request.Values, "party.organization_name"),
		nullableTextValue(request.Values, "party.role_title"),
		nullableTextValue(request.Values, "party.primary_email"),
		nullableTextValue(request.Values, "party.timezone_name"),
		nullableTextValue(request.Values, "party.external_ref"),
		nullableTextValue(request.Values, "party.notes"),
		now)
	if err != nil {
		return fmt.Errorf("insert party: %w", err)
	}
	return nil
}

func insertArtifactTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, incidentID uuid.UUID, actorID uuid.UUID, request CreateRequest, now time.Time) error {
	artifactType := artifactTypeForView(request.ViewSchemaID)
	timestamp := now
	if value, ok := request.Values[artifactType+".timestamp_utc"]; ok && value.Timestamp != nil {
		timestamp = value.Timestamp.UTC()
	}
	commID, handoffID, statusReviewID, lessonID := any(nil), any(nil), any(nil), any(nil)
	switch request.ViewSchemaID {
	case CommLogViewSchemaID:
		commID = uuid.NewString()
	case HandoffViewSchemaID:
		handoffID = uuid.NewString()
	case StatusReviewViewSchemaID:
		statusReviewID = uuid.NewString()
	case LessonViewSchemaID:
		lessonID = uuid.NewString()
	}
	outgoingOwner := nullableUUIDValue(request.Values, "handoff.outgoing_owner_user_id")
	if request.ViewSchemaID == HandoffViewSchemaID && outgoingOwner == nil {
		outgoingOwner = actorID
	}
	currentStateSummary := nullableTextValue(request.Values, "handoff.current_state_summary")
	if request.ViewSchemaID == StatusReviewViewSchemaID {
		currentStateSummary = nullableTextValue(request.Values, "status_review.current_state_summary")
	}
	summary := nullableTextValue(request.Values, "comm_log.summary")
	if request.ViewSchemaID == LessonViewSchemaID {
		summary = nullableTextValue(request.Values, "lesson.summary")
	}
	nextReportAt := nullableTimestampValue(request.Values, "comm_log.next_report_at")
	if request.ViewSchemaID == StatusReviewViewSchemaID {
		nextReportAt = nullableTimestampValue(request.Values, "status_review.next_report_at")
	}
	reviewOwner := nullableUUIDValue(request.Values, "status_review.review_owner_user_id")
	if request.ViewSchemaID == StatusReviewViewSchemaID && reviewOwner == nil {
		reviewOwner = actorID
	}
	lessonOwner := nullableUUIDValue(request.Values, "lesson.owner_user_id")
	if request.ViewSchemaID == LessonViewSchemaID && lessonOwner == nil {
		lessonOwner = actorID
	}
	closureState := nullableTextValue(request.Values, "lesson.closure_state")
	if request.ViewSchemaID == LessonViewSchemaID && closureState == nil {
		closureState = "open"
	}
	_, err := tx.Exec(ctx, `
INSERT INTO artifacts (
    record_id, incident_id, artifact_type, timestamp_utc, updated_at, created_at,
    title, body,
    comm_id, comm_type, audience, channel_or_meeting, summary, next_report_at, privilege_tag,
    handoff_id, outgoing_owner_user_id, incoming_owner_user_id, current_state_summary, next_checks, acknowledged_at,
    status_review_id, review_owner_user_id, active_risks_summary,
    lesson_id, owner_user_id, closure_state, created_by_user_id
) VALUES (
    $1, $2, $3, $4, $5, $5,
    $6, $7,
    $8, $9, $10, $11, $12, $13, $14,
    $15, $16, $17, $18, $19, $20,
    $21, $22, $23,
    $24, $25, $26, $27
)
	`, recordID, incidentID, artifactType, timestamp, now,
		nullableTextValue(request.Values, "note.title"), nullableTextValue(request.Values, "note.body"),
		commID, nullableTextValue(request.Values, "comm_log.comm_type"), nullableTextValue(request.Values, "comm_log.audience"), nullableTextValue(request.Values, "comm_log.channel_or_meeting"), summary, nextReportAt, nullableTextValue(request.Values, "comm_log.privilege_tag"),
		handoffID, outgoingOwner, nullableUUIDValue(request.Values, "handoff.incoming_owner_user_id"), currentStateSummary, nullableTextValue(request.Values, "handoff.next_checks"), nullableTimestampValue(request.Values, "handoff.acknowledged_at"),
		statusReviewID, reviewOwner, nullableTextValue(request.Values, "status_review.active_risks_summary"),
		lessonID, lessonOwner, closureState, actorID)
	if err != nil {
		return fmt.Errorf("insert artifact: %w", err)
	}
	return nil
}

func insertTaskRequestTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, incidentID uuid.UUID, actorID uuid.UUID, request CreateRequest, now time.Time) error {
	status := nullableTextValue(request.Values, "task.status")
	if status == nil {
		status = "open"
	}
	owner := nullableUUIDValue(request.Values, "task.owner_user_id")
	if owner == nil {
		owner = actorID
	}
	priority := nullableTextValue(request.Values, "task.priority")
	if priority == nil {
		priority = "normal"
	}
	completedAt := nullableTimestampValue(request.Values, "task.completed_at")
	if status == "done" && completedAt == nil {
		completedAt = now
	}
	if err := validateTaskCreateState(taskLifecycleState{
		Status:        status.(string),
		BlockedReason: nullableStringFromAny(nullableTextValue(request.Values, "task.blocked_reason")),
		CompletedAt:   nullableTimeFromAny(completedAt),
		OwnerUserID:   nullableUUIDStringFromAny(owner),
		CreatedAt:     now,
	}); err != nil {
		return err
	}
	_, err := tx.Exec(ctx, `
INSERT INTO task_requests (
    record_id, incident_id, title, status, owner_user_id, priority, task_kind,
    workstream, due_at, requester_party_text, requester_party_id, blocked_reason,
    completed_at, external_ticket_ref, closure_summary, decision_record_id,
    created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6, $7,
    $8, $9, $10, $11, $12,
    $13, $14, $15, $16,
    $17, $17
)
`, recordID, incidentID,
		textValue(request.Values, "task.title"),
		status,
		owner,
		priority,
		textValue(request.Values, "task.task_kind"),
		nullableTextValue(request.Values, "task.workstream"),
		nullableTimestampValue(request.Values, "task.due_at"),
		nullableTextValue(request.Values, "task.requester_party_text"),
		nullableUUIDValue(request.Values, "task.requester_party_id"),
		nullableTextValue(request.Values, "task.blocked_reason"),
		completedAt,
		nullableTextValue(request.Values, "task.external_ticket_ref"),
		nullableTextValue(request.Values, "task.closure_summary"),
		nullableUUIDValue(request.Values, "task.decision_record_id"),
		now)
	if err != nil {
		return fmt.Errorf("insert task request: %w", err)
	}
	return nil
}

func insertDecisionTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, incidentID uuid.UUID, actorID uuid.UUID, request CreateRequest, now time.Time) error {
	status := nullableTextValue(request.Values, "decision.status")
	if status == nil {
		status = "proposed"
	}
	owner := nullableUUIDValue(request.Values, "decision.owner_user_id")
	if owner == nil {
		owner = actorID
	}
	decidedAt := nullableTimestampValue(request.Values, "decision.decided_at")
	if decidedAt == nil {
		decidedAt = now
	}
	_, err := tx.Exec(ctx, `
INSERT INTO decisions (
    record_id, incident_id, summary, status, owner_user_id, decision_type,
    decided_at, rationale, created_at, updated_at
) VALUES (
    $1, $2, $3, $4, $5, $6,
    $7, $8, $9, $9
)
`, recordID, incidentID,
		textValue(request.Values, "decision.summary"),
		status,
		owner,
		textValue(request.Values, "decision.decision_type"),
		decidedAt,
		textValue(request.Values, "decision.rationale"),
		now)
	if err != nil {
		return fmt.Errorf("insert decision: %w", err)
	}
	return nil
}

func validateCreateReferencesTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, request CreateRequest) error {
	for fieldKey, value := range request.Values {
		if value.UUID != nil && strings.HasSuffix(fieldKey, "_user_id") {
			if err := validateActiveUserTx(ctx, tx, *value.UUID, fieldKey); err != nil {
				return err
			}
		}
		if value.UUID != nil {
			if err := validateDirectReferenceTx(ctx, tx, incidentID, fieldKey, *value.UUID); err != nil {
				return err
			}
		}
	}
	for fieldKey, payload := range request.Collections {
		if err := validateCollectionPayloadTx(ctx, tx, incidentID, fieldKey, payload); err != nil {
			return err
		}
	}
	return nil
}

func validatePatchReferencesTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, request PatchRequest) error {
	for _, change := range request.Changes {
		if change.Value != nil && change.Value.UUID != nil && strings.HasSuffix(change.FieldKey, "_user_id") {
			if err := validateActiveUserTx(ctx, tx, *change.Value.UUID, change.FieldKey); err != nil {
				return err
			}
		}
		if change.Value != nil && change.Value.UUID != nil {
			if err := validateDirectReferenceTx(ctx, tx, incidentID, change.FieldKey, *change.Value.UUID); err != nil {
				return err
			}
		}
		if change.Collection != nil {
			if err := validateCollectionPayloadTx(ctx, tx, incidentID, change.FieldKey, *change.Collection); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateCollectionPayloadTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, fieldKey string, payload CollectionActionPayload) error {
	for _, action := range payload.Actions {
		switch {
		case action.LinkedRecordID != nil:
			if err := validateTargetRecordTx(ctx, tx, incidentID, *action.LinkedRecordID, expectedTargetType(fieldKey), fieldKey); err != nil {
				return err
			}
		case action.PartyID != nil:
			if err := validateTargetRecordTx(ctx, tx, incidentID, *action.PartyID, "party", fieldKey); err != nil {
				return err
			}
		}
	}
	return nil
}

func validateDirectReferenceTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, fieldKey string, recordID uuid.UUID) error {
	switch fieldKey {
	case "task.requester_party_id":
		return validateTargetRecordTx(ctx, tx, incidentID, recordID, "party", fieldKey)
	case "task.decision_record_id":
		return validateTargetRecordTx(ctx, tx, incidentID, recordID, "decision", fieldKey)
	case "evidence.collector_party_id", "evidence.source_party_id":
		return validateTargetRecordTx(ctx, tx, incidentID, recordID, "party", fieldKey)
	default:
		return nil
	}
}

func validateActiveUserTx(ctx context.Context, tx pgx.Tx, userID uuid.UUID, field string) error {
	var exists bool
	if err := tx.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM users WHERE id = $1 AND is_active = true)`, userID).Scan(&exists); err != nil {
		return fmt.Errorf("validate user: %w", err)
	}
	if !exists {
		return mutationValidationError(field, "invalid_value")
	}
	return nil
}

func validateTargetRecordTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, expectedType string, field string) error {
	var exists bool
	if expectedType == "" {
		if err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1
      FROM records
     WHERE incident_id = $1
       AND record_id = $2
       AND deleted_at IS NULL
)
`, incidentID, recordID).Scan(&exists); err != nil {
			return fmt.Errorf("validate collection target: %w", err)
		}
		if !exists {
			return mutationValidationError(field, "invalid_value")
		}
		return nil
	}
	if err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1
      FROM records
     WHERE incident_id = $1
       AND record_id = $2
       AND record_type = $3
       AND deleted_at IS NULL
)
`, incidentID, recordID, expectedType).Scan(&exists); err != nil {
		return fmt.Errorf("validate collection target: %w", err)
	}
	if !exists {
		return mutationValidationError(field, "invalid_value")
	}
	return nil
}

func applyPatchTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, actorID uuid.UUID, request PatchRequest, now time.Time) (bool, error) {
	changed := false
	var beforeTask taskLifecycleState
	var beforeDecisionStatus string
	var err error
	if request.ViewSchemaID == TaskRequestsViewSchemaID {
		beforeTask, err = loadTaskLifecycleStateTx(ctx, tx, recordID)
		if err != nil {
			return false, err
		}
	}
	if request.ViewSchemaID == DecisionsViewSchemaID && touchesField(request.Changes, "decision.status") {
		beforeDecisionStatus, err = loadDecisionStatusTx(ctx, tx, recordID)
		if err != nil {
			return false, err
		}
	}
	for _, change := range request.Changes {
		if change.Value != nil {
			applied, err := applyDirectChangeTx(ctx, tx, recordID, request.ViewSchemaID, change, now)
			if err != nil {
				return false, err
			}
			changed = changed || applied
			continue
		}
		if change.Collection != nil {
			applied, err := applyCollectionPayloadTx(ctx, tx, incidentID, recordID, actorID, change.FieldKey, *change.Collection, now)
			if err != nil {
				return false, err
			}
			changed = changed || applied
		}
	}
	if request.ViewSchemaID == TaskRequestsViewSchemaID && touchesAnyField(request.Changes, "task.status", "task.blocked_reason", "task.completed_at", "task.owner_user_id") {
		applied, err := normalizeTaskLifecycleTx(ctx, tx, recordID, beforeTask, touchesField(request.Changes, "task.completed_at"), now)
		if err != nil {
			return false, err
		}
		changed = changed || applied
	}
	if request.ViewSchemaID == DecisionsViewSchemaID && touchesField(request.Changes, "decision.status") {
		afterDecisionStatus, err := loadDecisionStatusTx(ctx, tx, recordID)
		if err != nil {
			return false, err
		}
		if err := validateDecisionStatusTransition(beforeDecisionStatus, afterDecisionStatus); err != nil {
			return false, err
		}
	}
	return changed, nil
}

func applyDirectChangeTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, viewSchemaID string, change PatchChange, now time.Time) (bool, error) {
	table, column := tableColumnForField(change.FieldKey)
	if table == "" || column == "" {
		return false, mutationValidationError(change.FieldKey, "unsupported_field_key")
	}
	if err := validateDirectFieldValue(change); err != nil {
		return false, err
	}
	value := directDBValue(*change.Value)
	tag, err := tx.Exec(ctx, fmt.Sprintf(`UPDATE %s SET %s = $2, updated_at = $3 WHERE record_id = $1 AND %s IS DISTINCT FROM $2`, table, column, column), recordID, value, now)
	if err != nil {
		return false, fmt.Errorf("apply direct change: %w", err)
	}
	return tag.RowsAffected() > 0, nil
}

func validateDirectFieldValue(change PatchChange) error {
	if change.Value == nil || change.Value.Text == nil {
		return nil
	}
	value := *change.Value.Text
	switch change.FieldKey {
	case "task.status":
		if !validTaskStatus(value) {
			return mutationValidationError(change.FieldKey, "invalid_value")
		}
	case "task.task_kind":
		if !validTaskKind(value) {
			return mutationValidationError(change.FieldKey, "invalid_value")
		}
	case "task.priority":
		if !validTaskPriority(value) {
			return mutationValidationError(change.FieldKey, "invalid_value")
		}
	case "decision.status":
		if !validDecisionStatus(value) {
			return mutationValidationError(change.FieldKey, "invalid_value")
		}
	case "decision.decision_type":
		if !validDecisionType(value) {
			return mutationValidationError(change.FieldKey, "invalid_value")
		}
	case "evidence.lifecycle_state":
		if !validEvidenceLifecycleState(value) {
			return mutationValidationError(change.FieldKey, "invalid_value")
		}
	}
	return nil
}

func applyCollectionPayloadsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, actorID uuid.UUID, collections map[string]CollectionActionPayload, now time.Time) error {
	keys := make([]string, 0, len(collections))
	for key := range collections {
		keys = append(keys, key)
	}
	slices.Sort(keys)
	for _, key := range keys {
		if _, err := applyCollectionPayloadTx(ctx, tx, incidentID, recordID, actorID, key, collections[key], now); err != nil {
			return err
		}
	}
	return nil
}

func applyCollectionPayloadTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, actorID uuid.UUID, fieldKey string, payload CollectionActionPayload, now time.Time) (bool, error) {
	changed := false
	for _, action := range payload.Actions {
		switch action.Op {
		case "add_record_ref":
			applied, err := upsertReferenceLinkTx(ctx, tx, incidentID, recordID, *action.LinkedRecordID, fieldKey, actorID, now)
			if err != nil {
				return false, err
			}
			changed = changed || applied
		case "remove_record_ref":
			dst, err := uuidFromItemRef(action.ItemRef, "record_ref:")
			if err != nil {
				return false, mutationValidationError(fieldKey, "invalid_value")
			}
			applied, err := tombstoneReferenceLinkTx(ctx, tx, incidentID, recordID, dst, fieldKey, expectedTargetType(fieldKey), actorID, now)
			if err != nil {
				return false, err
			}
			changed = changed || applied
		case "add_tag":
			applied, err := upsertTagTx(ctx, tx, incidentID, recordID, action.RawText, action.NormalizedText, actorID, now)
			if err != nil {
				return false, err
			}
			changed = changed || applied
		case "remove_tag":
			_, tagID, err := recordTagItemRefParts(action.ItemRef)
			if err != nil {
				return false, mutationValidationError(fieldKey, "invalid_value")
			}
			applied, err := tombstoneTagTx(ctx, tx, incidentID, recordID, tagID, actorID, now)
			if err != nil {
				return false, err
			}
			changed = changed || applied
		case "add_party_ref":
			applied, err := upsertReferenceLinkTx(ctx, tx, incidentID, recordID, *action.PartyID, fieldKey, actorID, now)
			if err != nil {
				return false, err
			}
			changed = changed || applied
		case "remove_party_ref":
			dst, err := uuidFromItemRef(action.ItemRef, "party_ref:")
			if err != nil {
				return false, mutationValidationError(fieldKey, "invalid_value")
			}
			applied, err := tombstoneReferenceLinkTx(ctx, tx, incidentID, recordID, dst, fieldKey, "party", actorID, now)
			if err != nil {
				return false, err
			}
			changed = changed || applied
		case "add_risk_ref":
			applied, err := upsertRiskRefTx(ctx, tx, incidentID, recordID, action.RiskRefText, action.NormalizedText, actorID, now)
			if err != nil {
				return false, err
			}
			changed = changed || applied
		case "remove_risk_ref":
			riskRefID, err := uuidFromItemRef(action.ItemRef, "risk_ref:")
			if err != nil {
				return false, mutationValidationError(fieldKey, "invalid_value")
			}
			applied, err := tombstoneRiskRefTx(ctx, tx, incidentID, recordID, riskRefID, actorID, now)
			if err != nil {
				return false, err
			}
			changed = changed || applied
		default:
			return false, mutationValidationError(fieldKey, "invalid_value")
		}
	}
	return changed, nil
}

func upsertReferenceLinkTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, src uuid.UUID, dst uuid.UUID, fieldKey string, actorID uuid.UUID, now time.Time) (bool, error) {
	tag, err := tx.Exec(ctx, `
INSERT INTO record_links (
    incident_id, src_record_id, dst_record_id, link_type, field_key,
    provenance, confidence, owner_user_id, created_by_user_id, decided_at, created_at
) VALUES ($1, $2, $3, $7, $4, 'manual', NULL, $5, $5, $6, $6)
ON CONFLICT (incident_id, src_record_id, dst_record_id, link_type, field_key)
WHERE deleted_at IS NULL AND field_key IS NOT NULL
DO NOTHING
`, incidentID, src, dst, fieldKey, actorID, now, linkTypeForField(fieldKey))
	if err != nil {
		return false, fmt.Errorf("upsert reference link: %w", err)
	}
	return tag.RowsAffected() > 0, nil
}

func insertLinkedNoteReferenceTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, src uuid.UUID, dst uuid.UUID, actorID uuid.UUID, now time.Time) error {
	_, err := tx.Exec(ctx, `
INSERT INTO record_links (
    incident_id, src_record_id, dst_record_id, link_type, field_key,
    provenance, confidence, owner_user_id, created_by_user_id, decided_at, created_at
) VALUES ($1, $2, $3, 'references_artifact', NULL, 'manual', NULL, $4, $4, $5, $5)
ON CONFLICT (incident_id, src_record_id, dst_record_id, link_type)
WHERE deleted_at IS NULL AND field_key IS NULL
DO NOTHING
`, incidentID, src, dst, actorID, now)
	if err != nil {
		return fmt.Errorf("insert linked note reference: %w", err)
	}
	return nil
}

func tombstoneReferenceLinkTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, src uuid.UUID, dst uuid.UUID, fieldKey string, expectedTargetType string, actorID uuid.UUID, now time.Time) (bool, error) {
	targetTypePredicate := ""
	args := []any{incidentID, src, dst, fieldKey, actorID, now, linkTypeForField(fieldKey)}
	if expectedTargetType != "" {
		targetTypePredicate = "AND dst.record_type = $8"
		args = append(args, expectedTargetType)
	}
	tag, err := tx.Exec(ctx, `
UPDATE record_links
   SET deleted_at = $6,
       deleted_by_user_id = $5
 WHERE incident_id = $1
   AND src_record_id = $2
   AND dst_record_id = $3
   AND link_type = $7
   AND field_key = $4
   AND deleted_at IS NULL
   AND EXISTS (
       SELECT 1
         FROM records dst
        WHERE dst.incident_id = record_links.incident_id
          AND dst.record_id = record_links.dst_record_id
          AND dst.deleted_at IS NULL
          `+targetTypePredicate+`
   )
`, args...)
	if err != nil {
		return false, fmt.Errorf("remove reference link: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return false, mutationValidationError(fieldKey, "invalid_value")
	}
	return true, nil
}

func loadLinkedNoteSourceIncidentTx(ctx context.Context, tx pgx.Tx, sourceRecordID uuid.UUID) (uuid.UUID, error) {
	var incidentID uuid.UUID
	err := tx.QueryRow(ctx, `
SELECT incident_id
  FROM records
 WHERE record_id = $1
   AND record_type IN ('timeline_event', 'host', 'identity', 'evidence')
   AND deleted_at IS NULL
`, sourceRecordID).Scan(&incidentID)
	return incidentID, err
}

func upsertRiskRefTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, text string, normalized string, actorID uuid.UUID, now time.Time) (bool, error) {
	tag, err := tx.Exec(ctx, `
INSERT INTO handoff_risk_refs (
    incident_id, handoff_record_id, risk_ref_text, normalized_risk_ref_text,
    created_by_user_id, created_at
) VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (handoff_record_id, normalized_risk_ref_text)
WHERE deleted_at IS NULL
DO NOTHING
`, incidentID, recordID, text, normalized, actorID, now)
	if err != nil {
		return false, fmt.Errorf("upsert risk ref: %w", err)
	}
	return tag.RowsAffected() > 0, nil
}

func tombstoneRiskRefTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, riskRefID uuid.UUID, actorID uuid.UUID, now time.Time) (bool, error) {
	tag, err := tx.Exec(ctx, `
UPDATE handoff_risk_refs
   SET deleted_at = $5,
       deleted_by_user_id = $4
 WHERE incident_id = $1
   AND handoff_record_id = $2
   AND risk_ref_id = $3
   AND deleted_at IS NULL
`, incidentID, recordID, riskRefID, actorID, now)
	if err != nil {
		return false, fmt.Errorf("remove risk ref: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return false, mutationValidationError("handoff.open_risk_refs", "invalid_value")
	}
	return true, nil
}

func upsertTagTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, tagName string, normalizedTagName string, actorID uuid.UUID, now time.Time) (bool, error) {
	if tagName == "" || normalizedTagName == "" {
		return false, mutationValidationError("note.tags", "invalid_value")
	}
	tag, err := tx.Exec(ctx, `
INSERT INTO record_tags (
    incident_id, record_id, tag_name, normalized_tag_name,
    created_by_user_id, created_at, updated_at
) VALUES ($1, $2, $3, $4, $5, $6, $6)
ON CONFLICT (incident_id, record_id, normalized_tag_name)
WHERE deleted_at IS NULL
DO NOTHING
`, incidentID, recordID, tagName, normalizedTagName, actorID, now)
	if err != nil {
		return false, fmt.Errorf("upsert record tag: %w", err)
	}
	return tag.RowsAffected() > 0, nil
}

func tombstoneTagTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID, tagID uuid.UUID, actorID uuid.UUID, now time.Time) (bool, error) {
	tag, err := tx.Exec(ctx, `
UPDATE record_tags
   SET deleted_at = $5,
       deleted_by_user_id = $4,
       updated_at = $5
 WHERE incident_id = $1
   AND record_id = $2
   AND record_tag_id = $3
   AND deleted_at IS NULL
`, incidentID, recordID, tagID, actorID, now)
	if err != nil {
		return false, fmt.Errorf("remove record tag: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return false, mutationValidationError("note.tags", "invalid_value")
	}
	return true, nil
}

func recordTagItemRefParts(itemRef string) (uuid.UUID, uuid.UUID, error) {
	parts := strings.Split(itemRef, ":")
	if len(parts) != 3 || parts[0] != "record_tag" {
		return uuid.UUID{}, uuid.UUID{}, fmt.Errorf("invalid record tag item ref")
	}
	recordID, err := uuid.Parse(parts[1])
	if err != nil {
		return uuid.UUID{}, uuid.UUID{}, err
	}
	tagID, err := uuid.Parse(parts[2])
	if err != nil {
		return uuid.UUID{}, uuid.UUID{}, err
	}
	return recordID, tagID, nil
}

func touchSourceRowTx(ctx context.Context, tx pgx.Tx, viewSchemaID string, recordID uuid.UUID, now time.Time) error {
	table := sourceTableForView(viewSchemaID)
	if table == "" {
		return mutationValidationError("view_schema_id", "unknown_view_schema")
	}
	if _, err := tx.Exec(ctx, fmt.Sprintf(`UPDATE %s SET updated_at = $2 WHERE record_id = $1`, table), recordID, now); err != nil {
		return fmt.Errorf("touch source row: %w", err)
	}
	return nil
}

func changedFieldKeys(before map[string]any, after map[string]any) []string {
	afterCells, _ := after["cells"].(map[string]any)
	beforeCells := map[string]any{}
	if before != nil {
		beforeCells, _ = before["cells"].(map[string]any)
	}
	keys := make([]string, 0)
	for fieldKey, afterValue := range afterCells {
		if beforeValue, ok := beforeCells[fieldKey]; !ok || !reflect.DeepEqual(beforeValue, afterValue) {
			keys = append(keys, fieldKey)
		}
	}
	slices.Sort(keys)
	return keys
}

type workbookPatchConflictWindow struct {
	BaseRow       map[string]any
	ChangedFields map[string]workbookPatchChangedField
}

type workbookPatchChangedField struct {
	ServerUpdatedBy uuid.UUID
	ServerUpdatedAt time.Time
}

func loadWorkbookPatchConflictWindowTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, viewSchemaID string, baseRowVersion int64, currentRowVersion int64) (workbookPatchConflictWindow, error) {
	rows, err := tx.Query(ctx, `
SELECT rr.row_version, rr.before_json, rr.after_json, cs.actor_user_id, cs.created_at
  FROM record_revisions rr
  JOIN change_sets cs
    ON cs.change_set_id = rr.change_set_id
 WHERE rr.record_id = $1
   AND rr.row_version >= $2
   AND rr.row_version <= $3
 ORDER BY rr.row_version ASC
`, recordID, baseRowVersion, currentRowVersion)
	if err != nil {
		return workbookPatchConflictWindow{}, fmt.Errorf("query workbook patch conflict window: %w", err)
	}
	defer rows.Close()

	window := workbookPatchConflictWindow{ChangedFields: make(map[string]workbookPatchChangedField)}
	for rows.Next() {
		var (
			rowVersion int64
			beforeJSON []byte
			afterJSON  []byte
			actorID    uuid.UUID
			createdAt  time.Time
		)
		if err := rows.Scan(&rowVersion, &beforeJSON, &afterJSON, &actorID, &createdAt); err != nil {
			return workbookPatchConflictWindow{}, fmt.Errorf("scan workbook patch conflict window: %w", err)
		}
		if rowVersion == baseRowVersion {
			baseRow, ok := decodeRevisionRow(afterJSON)
			if !ok {
				return workbookPatchConflictWindow{}, &RowVersionConflictError{RecordID: recordID, BaseRowVersion: baseRowVersion, CurrentRowVersion: currentRowVersion}
			}
			window.BaseRow = baseRow
			continue
		}
		beforeRow, beforeOK := decodeRevisionRow(beforeJSON)
		afterRow, afterOK := decodeRevisionRow(afterJSON)
		if !beforeOK || !afterOK {
			return workbookPatchConflictWindow{}, &RowVersionConflictError{RecordID: recordID, BaseRowVersion: baseRowVersion, CurrentRowVersion: currentRowVersion}
		}
		for _, fieldKey := range changedRevisionWritableFieldKeys(viewSchemaID, beforeRow, afterRow) {
			window.ChangedFields[fieldKey] = workbookPatchChangedField{
				ServerUpdatedBy: actorID,
				ServerUpdatedAt: createdAt.UTC(),
			}
		}
	}
	if err := rows.Err(); err != nil {
		return workbookPatchConflictWindow{}, fmt.Errorf("iterate workbook patch conflict window: %w", err)
	}
	if window.BaseRow == nil {
		return workbookPatchConflictWindow{}, &RowVersionConflictError{RecordID: recordID, BaseRowVersion: baseRowVersion, CurrentRowVersion: currentRowVersion}
	}
	return window, nil
}

func decodeRevisionRow(data []byte) (map[string]any, bool) {
	if len(data) == 0 {
		return nil, false
	}
	var row map[string]any
	if err := json.Unmarshal(data, &row); err != nil {
		return nil, false
	}
	if _, ok := row["cells"].(map[string]any); !ok {
		return nil, false
	}
	return row, true
}

func changedRevisionWritableFieldKeys(viewSchemaID string, beforeRow map[string]any, afterRow map[string]any) []string {
	beforeCells, _ := beforeRow["cells"].(map[string]any)
	afterCells, _ := afterRow["cells"].(map[string]any)
	changed := make([]string, 0)
	for fieldKey, afterCell := range afterCells {
		field, ok := viewschema.LookupField(viewSchemaID, fieldKey)
		if !ok || !field.Writable || isReadOnlySystemField(fieldKey) {
			continue
		}
		if !reflect.DeepEqual(beforeCells[fieldKey], afterCell) {
			changed = append(changed, fieldKey)
		}
	}
	slices.Sort(changed)
	return changed
}

func overlappingWorkbookPatchChange(changes []PatchChange, changedFields map[string]workbookPatchChangedField) (PatchChange, workbookPatchChangedField, bool) {
	for _, change := range changes {
		changed, ok := changedFields[change.FieldKey]
		if ok {
			return change, changed, true
		}
	}
	return PatchChange{}, workbookPatchChangedField{}, false
}

func buildWorkbookSameFieldConflict(recordID uuid.UUID, viewSchemaID string, baseRowVersion int64, currentRowVersion int64, requestHash []byte, window workbookPatchConflictWindow, change PatchChange, changed workbookPatchChangedField, currentRow map[string]any) (*SameFieldConflictError, error) {
	baseValue, ok := rowCellValue(window.BaseRow, change.FieldKey)
	if !ok {
		return nil, &RowVersionConflictError{RecordID: recordID, BaseRowVersion: baseRowVersion, CurrentRowVersion: currentRowVersion}
	}
	serverValue, ok := rowCellValue(currentRow, change.FieldKey)
	if !ok {
		return nil, &RowVersionConflictError{RecordID: recordID, BaseRowVersion: baseRowVersion, CurrentRowVersion: currentRowVersion}
	}
	clientValue, err := workbookPatchClientConflictValue(recordID, change, baseValue, requestHash)
	if err != nil {
		return nil, &RowVersionConflictError{RecordID: recordID, BaseRowVersion: baseRowVersion, CurrentRowVersion: currentRowVersion}
	}
	field, _ := viewschema.LookupField(viewSchemaID, change.FieldKey)
	conflictClass := field.ConflictResolutionClass
	if conflictClass == "" {
		conflictClass = "atomic_replace"
	}
	conflict := map[string]any{
		"conflict_token":            workbookConflictToken(recordID, viewSchemaID, change.FieldKey, conflictClass, baseRowVersion, currentRowVersion, requestHash),
		"record_id":                 recordID.String(),
		"field_key":                 change.FieldKey,
		"conflict_resolution_class": conflictClass,
		"base_row_version":          baseRowVersion,
		"current_row_version":       currentRowVersion,
		"client_value":              clientValue,
		"server_value":              serverValue,
		"server_updated_by":         changed.ServerUpdatedBy.String(),
		"server_updated_at":         changed.ServerUpdatedAt.UTC().Format(time.RFC3339Nano),
		"base_value":                baseValue,
	}
	if conflictClass == "text_compare_merge" {
		if suggested, ok := suggestedTextMergeValue(baseValue, serverValue, clientValue); ok {
			conflict["suggested_merged_value"] = suggested
		}
	}
	return &SameFieldConflictError{Conflict: conflict}, nil
}

func rowCellValue(row map[string]any, fieldKey string) (any, bool) {
	cells, _ := row["cells"].(map[string]any)
	cell, ok := cells[fieldKey].(map[string]any)
	if !ok {
		return nil, false
	}
	value, ok := cell["value"]
	return value, ok
}

func workbookPatchClientConflictValue(recordID uuid.UUID, change PatchChange, baseValue any, requestHash []byte) (any, error) {
	if change.Collection == nil {
		return canonicalValue(*change.Value), nil
	}
	return applyWorkbookCollectionConflictActions(recordID, change.FieldKey, baseValue, *change.Collection, requestHash)
}

func applyWorkbookCollectionConflictActions(recordID uuid.UUID, fieldKey string, baseValue any, payload CollectionActionPayload, requestHash []byte) (map[string]any, error) {
	ordered, items, ok := cloneWorkbookCollectionConflictValue(baseValue)
	if !ok {
		return nil, fmt.Errorf("invalid base collection value for %s", fieldKey)
	}
	for index, action := range payload.Actions {
		switch action.Op {
		case "add_record_ref", "add_party_ref", "add_tag", "add_risk_ref":
			items = upsertWorkbookCollectionConflictItem(items, newWorkbookClientCollectionItem(recordID, fieldKey, action, requestHash, index))
		case "remove_record_ref", "remove_party_ref", "remove_tag", "remove_risk_ref":
			items = removeWorkbookCollectionConflictItem(items, action.ItemRef)
		default:
			return nil, fmt.Errorf("unsupported collection action: %s", action.Op)
		}
	}
	if !ordered {
		slices.SortFunc(items, func(left map[string]any, right map[string]any) int {
			return strings.Compare(workbookCollectionSortKey(left), workbookCollectionSortKey(right))
		})
	}
	return map[string]any{"kind": "collection_value_v1", "ordered": ordered, "items": items}, nil
}

func cloneWorkbookCollectionConflictValue(value any) (bool, []map[string]any, bool) {
	object, ok := value.(map[string]any)
	if !ok || object["kind"] != "collection_value_v1" {
		return false, nil, false
	}
	ordered, ok := object["ordered"].(bool)
	if !ok {
		return false, nil, false
	}
	items := make([]map[string]any, 0)
	switch rawItems := object["items"].(type) {
	case []any:
		for _, rawItem := range rawItems {
			item, ok := rawItem.(map[string]any)
			if !ok {
				return false, nil, false
			}
			items = append(items, cloneWorkbookMap(item))
		}
	case []map[string]any:
		for _, item := range rawItems {
			items = append(items, cloneWorkbookMap(item))
		}
	default:
		return false, nil, false
	}
	return ordered, items, true
}

func cloneWorkbookMap(source map[string]any) map[string]any {
	cloned := make(map[string]any, len(source))
	for key, value := range source {
		cloned[key] = value
	}
	return cloned
}

func newWorkbookClientCollectionItem(recordID uuid.UUID, fieldKey string, action CollectionAction, requestHash []byte, actionIndex int) map[string]any {
	switch action.Op {
	case "add_record_ref":
		linkedID := action.LinkedRecordID.String()
		targetType := expectedTargetType(fieldKey)
		if targetType == "" {
			targetType = "record"
		}
		return map[string]any{
			"item_ref":         "record_ref:" + linkedID,
			"item_kind":        "record_ref",
			"display_text":     targetType + ":" + linkedID,
			"linked_record_id": linkedID,
		}
	case "add_party_ref":
		partyID := action.PartyID.String()
		return map[string]any{
			"item_ref":     "party_ref:" + partyID,
			"item_kind":    "party_ref",
			"display_text": "party:" + partyID,
			"party_id":     partyID,
		}
	case "add_tag":
		tagID := workbookConflictLocalUUID(recordID, fieldKey, action, requestHash, actionIndex)
		return map[string]any{
			"item_ref":     "record_tag:" + recordID.String() + ":" + tagID.String(),
			"item_kind":    "tag",
			"display_text": action.RawText,
			"tag_id":       tagID.String(),
		}
	case "add_risk_ref":
		riskRefID := workbookConflictLocalUUID(recordID, fieldKey, action, requestHash, actionIndex)
		return map[string]any{
			"item_ref":      "risk_ref:" + riskRefID.String(),
			"item_kind":     "risk_ref",
			"display_text":  action.RiskRefText,
			"risk_ref_id":   riskRefID.String(),
			"risk_ref_text": action.RiskRefText,
		}
	default:
		return map[string]any{}
	}
}

func upsertWorkbookCollectionConflictItem(items []map[string]any, item map[string]any) []map[string]any {
	itemRef, _ := item["item_ref"].(string)
	if itemRef == "" {
		return items
	}
	for index, existing := range items {
		if existing["item_ref"] == itemRef {
			items[index] = item
			return items
		}
		if item["item_kind"] == "risk_ref" && existing["item_kind"] == "risk_ref" && existing["risk_ref_text"] == item["risk_ref_text"] {
			items[index] = item
			return items
		}
	}
	return append(items, item)
}

func removeWorkbookCollectionConflictItem(items []map[string]any, itemRef string) []map[string]any {
	filtered := items[:0]
	for _, item := range items {
		if item["item_ref"] != itemRef {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func workbookCollectionSortKey(item map[string]any) string {
	for _, key := range []string{"item_kind", "display_text", "item_ref"} {
		if value, ok := item[key].(string); ok && value != "" {
			return value
		}
	}
	return ""
}

type workbookConflictTokenClaims struct {
	RecordID                string `json:"record_id"`
	ViewSchemaID            string `json:"view_schema_id"`
	FieldKey                string `json:"field_key"`
	ConflictResolutionClass string `json:"conflict_resolution_class"`
	BaseRowVersion          int64  `json:"base_row_version"`
	CurrentRowVersion       int64  `json:"current_row_version"`
	RequestHash             string `json:"request_hash"`
	Signature               string `json:"sig"`
}

func workbookConflictToken(recordID uuid.UUID, viewSchemaID string, fieldKey string, conflictClass string, baseRowVersion int64, currentRowVersion int64, requestHash []byte) string {
	claims := workbookConflictTokenClaims{
		RecordID:                recordID.String(),
		ViewSchemaID:            viewSchemaID,
		FieldKey:                fieldKey,
		ConflictResolutionClass: conflictClass,
		BaseRowVersion:          baseRowVersion,
		CurrentRowVersion:       currentRowVersion,
		RequestHash:             base64.RawURLEncoding.EncodeToString(requestHash),
	}
	claims.Signature = workbookConflictTokenSignature(claims)
	payload, _ := json.Marshal(claims)
	return base64.RawURLEncoding.EncodeToString(payload)
}

func parseWorkbookConflictToken(token string) (workbookConflictTokenClaims, bool) {
	data, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return workbookConflictTokenClaims{}, false
	}
	var claims workbookConflictTokenClaims
	if err := json.Unmarshal(data, &claims); err != nil {
		return workbookConflictTokenClaims{}, false
	}
	if claims.Signature == "" || claims.Signature != workbookConflictTokenSignature(claims) {
		return workbookConflictTokenClaims{}, false
	}
	if _, err := uuid.Parse(claims.RecordID); err != nil {
		return workbookConflictTokenClaims{}, false
	}
	if claims.ViewSchemaID == "" || claims.FieldKey == "" || claims.BaseRowVersion < 1 || claims.CurrentRowVersion < claims.BaseRowVersion {
		return workbookConflictTokenClaims{}, false
	}
	return claims, true
}

func workbookConflictTokenSignature(claims workbookConflictTokenClaims) string {
	payload := map[string]any{
		"record_id":                  claims.RecordID,
		"view_schema_id":             claims.ViewSchemaID,
		"field_key":                  claims.FieldKey,
		"conflict_resolution_class":  claims.ConflictResolutionClass,
		"base_row_version":           claims.BaseRowVersion,
		"current_row_version":        claims.CurrentRowVersion,
		"request_hash":               claims.RequestHash,
		"cartulary_conflict_token_v": 1,
	}
	return base64.RawURLEncoding.EncodeToString(hashRequestPayload(payload))
}

func workbookConflictLocalUUID(recordID uuid.UUID, fieldKey string, action CollectionAction, requestHash []byte, actionIndex int) uuid.UUID {
	seed, _ := json.Marshal(map[string]any{
		"record_id":     recordID.String(),
		"field_key":     fieldKey,
		"request_hash":  base64.RawURLEncoding.EncodeToString(requestHash),
		"action_index":  actionIndex,
		"op":            action.Op,
		"risk_ref_text": action.NormalizedText,
	})
	return uuid.NewSHA1(uuid.NameSpaceOID, seed)
}

func decodeStoredResponse(data []byte) (map[string]any, error) {
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func extractPayloadUUID(payload map[string]any, path ...string) (uuid.UUID, error) {
	current := any(payload)
	for _, segment := range path {
		object, ok := current.(map[string]any)
		if !ok {
			return uuid.UUID{}, fmt.Errorf("decode payload path %q", strings.Join(path, "."))
		}
		current = object[segment]
	}
	text, ok := current.(string)
	if !ok {
		return uuid.UUID{}, fmt.Errorf("decode payload uuid path %q", strings.Join(path, "."))
	}
	return uuid.Parse(text)
}

func workbookVersionID(recordID uuid.UUID, rowVersion int64) string {
	return fmt.Sprintf("record:%s:%d", recordID.String(), rowVersion)
}

func recordTypeMatchesView(recordType string, viewSchemaID string) bool {
	return recordType == recordTypeForView(viewSchemaID)
}

func recordTypeForView(viewSchemaID string) string {
	switch viewSchemaID {
	case EvidenceViewSchemaID:
		return "evidence"
	case PartiesViewSchemaID:
		return "party"
	case TaskRequestsViewSchemaID:
		return "task_request"
	case DecisionsViewSchemaID:
		return "decision"
	case NotesViewSchemaID, CommLogViewSchemaID, HandoffViewSchemaID, StatusReviewViewSchemaID, LessonViewSchemaID:
		return "artifact"
	default:
		return ""
	}
}

func sourceTableForView(viewSchemaID string) string {
	switch recordTypeForView(viewSchemaID) {
	case "evidence":
		return "evidence"
	case "party":
		return "parties"
	case "task_request":
		return "task_requests"
	case "decision":
		return "decisions"
	case "artifact":
		return "artifacts"
	default:
		return ""
	}
}

func artifactTypeForView(viewSchemaID string) string {
	switch viewSchemaID {
	case NotesViewSchemaID:
		return "note"
	case CommLogViewSchemaID:
		return "comm_log"
	case HandoffViewSchemaID:
		return "handoff"
	case StatusReviewViewSchemaID:
		return "status_review"
	case LessonViewSchemaID:
		return "lesson"
	default:
		return ""
	}
}

func tableColumnForField(fieldKey string) (string, string) {
	switch {
	case strings.HasPrefix(fieldKey, "evidence."):
		return "evidence", strings.TrimPrefix(fieldKey, "evidence.")
	case strings.HasPrefix(fieldKey, "party."):
		return "parties", strings.TrimPrefix(fieldKey, "party.")
	case fieldKey == "note.title":
		return "artifacts", "title"
	case fieldKey == "note.body":
		return "artifacts", "body"
	case strings.HasPrefix(fieldKey, "comm_log."):
		return "artifacts", strings.TrimPrefix(fieldKey, "comm_log.")
	case strings.HasPrefix(fieldKey, "handoff."):
		return "artifacts", strings.TrimPrefix(fieldKey, "handoff.")
	case strings.HasPrefix(fieldKey, "status_review."):
		return "artifacts", strings.TrimPrefix(fieldKey, "status_review.")
	case strings.HasPrefix(fieldKey, "lesson."):
		return "artifacts", strings.TrimPrefix(fieldKey, "lesson.")
	case strings.HasPrefix(fieldKey, "task."):
		return "task_requests", strings.TrimPrefix(fieldKey, "task.")
	case strings.HasPrefix(fieldKey, "decision."):
		return "decisions", strings.TrimPrefix(fieldKey, "decision.")
	default:
		return "", ""
	}
}

func directDBValue(value ValueChange) any {
	switch value.Kind {
	case "timestamp":
		if value.Timestamp == nil {
			return nil
		}
		return value.Timestamp.UTC()
	case "uuid":
		if value.UUID == nil {
			return nil
		}
		return *value.UUID
	case "text":
		if value.Text == nil {
			return nil
		}
		return *value.Text
	default:
		return nil
	}
}

func uuidFromItemRef(itemRef string, prefix string) (uuid.UUID, error) {
	if !strings.HasPrefix(itemRef, prefix) {
		return uuid.UUID{}, fmt.Errorf("invalid item ref")
	}
	return uuid.Parse(strings.TrimPrefix(itemRef, prefix))
}

func hasTextValue(values map[string]ValueChange, field string) bool {
	value, ok := values[field]
	return ok && value.Text != nil && *value.Text != ""
}

func hasUUIDValue(values map[string]ValueChange, field string) bool {
	value, ok := values[field]
	return ok && value.UUID != nil
}

func validValueText(values map[string]ValueChange, field string, valid func(string) bool) bool {
	value, ok := values[field]
	return ok && value.Text != nil && valid(*value.Text)
}

func textValue(values map[string]ValueChange, field string) string {
	value := values[field]
	return derefText(value.Text)
}

func nullableTextValue(values map[string]ValueChange, field string) any {
	value, ok := values[field]
	if !ok || value.Text == nil {
		return nil
	}
	return *value.Text
}

func nullableUUIDValue(values map[string]ValueChange, field string) any {
	value, ok := values[field]
	if !ok || value.UUID == nil {
		return nil
	}
	return *value.UUID
}

func nullableTimestampValue(values map[string]ValueChange, field string) any {
	value, ok := values[field]
	if !ok || value.Timestamp == nil {
		return nil
	}
	return value.Timestamp.UTC()
}

func derefText(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func expectedTargetType(fieldKey string) string {
	switch fieldKey {
	case "comm_log.decision_ids", "handoff.open_decision_ids", "status_review.open_decision_ids":
		return "decision"
	case "comm_log.action_task_ids", "handoff.open_task_ids", "status_review.blocked_task_ids", "lesson.follow_up_task_ids":
		return "task_request"
	case "status_review.pending_evidence_ids", "lesson.evidence_refs":
		return "evidence"
	case "task.linked_record_ids", "decision.support_refs":
		return ""
	default:
		return ""
	}
}

func linkTypeForField(fieldKey string) string {
	switch fieldKey {
	case "decision.support_refs":
		return "supported_by"
	default:
		return "references_record"
	}
}

func validPartyKind(value string) bool {
	switch value {
	case "person", "team", "organization", "distribution_list", "other":
		return true
	default:
		return false
	}
}

func validEvidenceLifecycleState(value string) bool {
	switch value {
	case "requested", "pending_receipt", "received", "available", "quarantined", "released":
		return true
	default:
		return false
	}
}

func validCommType(value string) bool {
	switch value {
	case "meeting", "notification", "approval", "briefing", "handoff":
		return true
	default:
		return false
	}
}

func validClosureState(value string) bool {
	return value == "open" || value == "closed"
}

func validTaskKind(value string) bool {
	switch value {
	case "question", "request", "collection", "containment", "follow_up":
		return true
	default:
		return false
	}
}

func validTaskStatus(value string) bool {
	switch value {
	case "open", "in_progress", "blocked", "done", "canceled":
		return true
	default:
		return false
	}
}

func validTaskPriority(value string) bool {
	switch value {
	case "low", "normal", "high", "urgent":
		return true
	default:
		return false
	}
}

func validDecisionType(value string) bool {
	switch value {
	case "scope", "containment", "communication", "evidence", "reporting":
		return true
	default:
		return false
	}
}

func validDecisionStatus(value string) bool {
	switch value {
	case "proposed", "approved", "rejected", "superseded", "executed":
		return true
	default:
		return false
	}
}

type taskLifecycleState struct {
	Status        string
	BlockedReason sql.NullString
	CompletedAt   sql.NullTime
	OwnerUserID   sql.NullString
	CreatedAt     time.Time
}

func loadTaskLifecycleStateTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (taskLifecycleState, error) {
	var state taskLifecycleState
	err := tx.QueryRow(ctx, `
SELECT status, NULLIF(blocked_reason, ''), completed_at, owner_user_id::text, created_at
  FROM task_requests
 WHERE record_id = $1
`, recordID).Scan(&state.Status, &state.BlockedReason, &state.CompletedAt, &state.OwnerUserID, &state.CreatedAt)
	if err != nil {
		return taskLifecycleState{}, err
	}
	return state, nil
}

func normalizeTaskLifecycleTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, before taskLifecycleState, explicitCompletedAt bool, now time.Time) (bool, error) {
	changed := false
	after, err := loadTaskLifecycleStateTx(ctx, tx, recordID)
	if err != nil {
		return false, err
	}
	if !validTaskStatus(after.Status) {
		return false, mutationValidationError("task.status", "invalid_value")
	}
	if before.Status != after.Status && !validTaskTransition(before.Status, after.Status) {
		return false, &LifecycleValidationError{FromStatus: before.Status, ToStatus: after.Status, ReasonCode: "illegal_status_transition", ViolatedGuards: []string{"task.status"}}
	}
	if before.Status != after.Status && before.Status == "blocked" && after.Status != "blocked" && after.BlockedReason.Valid {
		if _, err := tx.Exec(ctx, `UPDATE task_requests SET blocked_reason = NULL WHERE record_id = $1`, recordID); err != nil {
			return false, fmt.Errorf("clear blocked reason: %w", err)
		}
		changed = true
	}
	if before.Status != after.Status && before.Status == "done" && after.Status != "done" && after.CompletedAt.Valid {
		if _, err := tx.Exec(ctx, `UPDATE task_requests SET completed_at = NULL WHERE record_id = $1`, recordID); err != nil {
			return false, fmt.Errorf("clear completed at: %w", err)
		}
		changed = true
	}
	if after.Status == "done" && !after.CompletedAt.Valid && !explicitCompletedAt {
		if _, err := tx.Exec(ctx, `UPDATE task_requests SET completed_at = $2 WHERE record_id = $1`, recordID, now); err != nil {
			return false, fmt.Errorf("fill completed at: %w", err)
		}
		changed = true
	}
	after, err = loadTaskLifecycleStateTx(ctx, tx, recordID)
	if err != nil {
		return false, err
	}
	if err := validateTaskCreateState(after); err != nil {
		return false, err
	}
	return changed, nil
}

func validateTaskCreateState(state taskLifecycleState) error {
	violated := taskLifecycleViolations(state)
	if len(violated) > 0 {
		return &LifecycleValidationError{ToStatus: state.Status, ReasonCode: "violated_lifecycle_guards", ViolatedGuards: violated}
	}
	return nil
}

func taskLifecycleViolations(state taskLifecycleState) []string {
	violated := []string{}
	if state.Status == "blocked" && !state.BlockedReason.Valid {
		violated = append(violated, "blocked_requires_reason")
	}
	if state.Status != "blocked" && state.BlockedReason.Valid {
		violated = append(violated, "non_blocked_clears_reason")
	}
	if state.Status == "done" {
		if !state.CompletedAt.Valid {
			violated = append(violated, "done_requires_completed_at")
		} else if state.CompletedAt.Time.Before(state.CreatedAt) {
			violated = append(violated, "completed_at_before_created_at")
		}
	}
	if state.Status != "done" && state.CompletedAt.Valid {
		violated = append(violated, "non_done_clears_completed_at")
	}
	if (state.Status == "open" || state.Status == "in_progress" || state.Status == "blocked") && !state.OwnerUserID.Valid {
		violated = append(violated, "active_task_requires_owner")
	}
	return violated
}

func validTaskTransition(from string, to string) bool {
	if from == to {
		return true
	}
	switch from {
	case "open", "in_progress", "blocked":
		return to == "open" || to == "in_progress" || to == "blocked" || to == "done" || to == "canceled"
	case "done":
		return to == "open" || to == "in_progress" || to == "blocked"
	case "canceled":
		return to == "open" || to == "in_progress" || to == "blocked"
	default:
		return false
	}
}

func loadDecisionStatusTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (string, error) {
	var status string
	err := tx.QueryRow(ctx, `SELECT status FROM decisions WHERE record_id = $1`, recordID).Scan(&status)
	return status, err
}

func validateDecisionStatusTransition(from string, to string) error {
	if !validDecisionStatus(to) {
		return mutationValidationError("decision.status", "invalid_value")
	}
	if to == "superseded" {
		return &LifecycleValidationError{FromStatus: from, ToStatus: to, ReasonCode: "superseded_direct_write", ViolatedGuards: []string{"decision.status"}}
	}
	if from == to {
		return nil
	}
	if (from == "proposed" && (to == "approved" || to == "rejected" || to == "executed")) ||
		(from == "approved" && to == "executed") {
		return nil
	}
	return &LifecycleValidationError{FromStatus: from, ToStatus: to, ReasonCode: "illegal_status_transition", ViolatedGuards: []string{"decision.status"}}
}

func nullableStringFromAny(value any) sql.NullString {
	if text, ok := value.(string); ok && text != "" {
		return sql.NullString{String: text, Valid: true}
	}
	return sql.NullString{}
}

func nullableTimeFromAny(value any) sql.NullTime {
	if timestamp, ok := value.(time.Time); ok {
		return sql.NullTime{Time: timestamp, Valid: true}
	}
	return sql.NullTime{}
}

func nullableUUIDStringFromAny(value any) sql.NullString {
	switch typed := value.(type) {
	case uuid.UUID:
		return sql.NullString{String: typed.String(), Valid: true}
	default:
		return sql.NullString{}
	}
}

func touchesField(changes []PatchChange, fieldKey string) bool {
	for _, change := range changes {
		if change.FieldKey == fieldKey {
			return true
		}
	}
	return false
}

func touchesAnyField(changes []PatchChange, fieldKeys ...string) bool {
	for _, fieldKey := range fieldKeys {
		if touchesField(changes, fieldKey) {
			return true
		}
	}
	return false
}
