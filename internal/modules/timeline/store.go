package timeline

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"sort"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"

	sqlc "github.com/JochiRaider/cartulary/internal/gen/sql"
	"github.com/JochiRaider/cartulary/internal/modules/entities"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/projections"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

var (
	ErrRecordNotFound     = errors.New("timeline: record not found")
	ErrRowVersionConflict = errors.New("timeline: row version conflict")
	ErrIllegalTransition  = errors.New("timeline: illegal transition")
	ErrNoEffectiveChange  = errors.New("timeline: no effective change")
)

type RowVersionConflictError struct {
	RecordID          uuid.UUID
	BaseRowVersion    int64
	CurrentRowVersion int64
}

func (e *RowVersionConflictError) Error() string {
	return ErrRowVersionConflict.Error()
}

func (e *RowVersionConflictError) Unwrap() error {
	return ErrRowVersionConflict
}

func (e *RowVersionConflictError) Details() map[string]any {
	if e == nil {
		return map[string]any{}
	}
	return map[string]any{
		"record_id":           e.RecordID.String(),
		"base_row_version":    e.BaseRowVersion,
		"current_row_version": e.CurrentRowVersion,
	}
}

type SameFieldConflictError struct {
	Conflict map[string]any
}

func (e *SameFieldConflictError) Error() string {
	return "timeline: same field conflict"
}

type TimelineConflictTokenClaims struct {
	Version                 int64  `json:"cartulary_conflict_token_v"`
	RecordID                string `json:"record_id"`
	ViewSchemaID            string `json:"view_schema_id"`
	FieldKey                string `json:"field_key"`
	ConflictResolutionClass string `json:"conflict_resolution_class"`
	BaseRowVersion          int64  `json:"base_row_version"`
	CurrentRowVersion       int64  `json:"current_row_version"`
	RequestHash             string `json:"request_hash"`
	Signature               string `json:"sig"`
}

type patchConflictWindow struct {
	BaseRow       map[string]any
	ChangedFields map[string]patchChangedField
}

type patchChangedField struct {
	FieldKey        string
	ServerUpdatedBy uuid.UUID
	ServerUpdatedAt time.Time
}

type Store struct {
	pool            postgres.DB
	authStore       *authn.Store
	recordStore     *records.Store
	revisionsStore  *revisions.Store
	projectionStore *projections.Store
	linkStore       *links.Store
	hooks           StoreHooks
}

type projectedRecord struct {
	RecordID              uuid.UUID
	IncidentID            uuid.UUID
	RowVersion            int64
	OccurredAt            *time.Time
	Summary               *string
	Details               *string
	SourceText            *string
	RecordedAt            time.Time
	EditedAt              time.Time
	SortTs                time.Time
	CaptureState          string
	ReplacementRecordID   *uuid.UUID
	OccurredDay           *time.Time
	RecordedDay           time.Time
	EvidenceCount         int
	HasEvidence           bool
	HasUnresolvedMentions bool
	HostRefs              []map[string]any
	IdentityRefs          []map[string]any
	AttachedEvidence      []map[string]any
	Tags                  []map[string]any
}

type sourceRecord struct {
	RecordID           uuid.UUID
	IncidentID         uuid.UUID
	OccurredAt         *time.Time
	Summary            *string
	Details            *string
	SourceText         *string
	CaptureState       string
	RowVersion         int64
	RecordedAt         time.Time
	EditedAt           time.Time
	CreatedByUserID    uuid.UUID
	UpdatedByUserID    uuid.UUID
	ReviewedByUserID   *uuid.UUID
	ReviewedAt         *time.Time
	SupersededByUserID *uuid.UUID
	SupersededAt       *time.Time
}

type MutationResult struct {
	Payload          map[string]any
	StatusCode       int
	Replayed         bool
	IncidentID       uuid.UUID
	RecordID         uuid.UUID
	ChangeSetID      uuid.UUID
	ClientTxnID      string
	RowVersion       int64
	ChangedFieldKeys []string
	Row              projectedRecord
}

type RecordSubstrateSnapshot struct {
	RecordID            uuid.UUID
	RowVersion          int64
	CaptureState        string
	ReplacementRecordID *uuid.UUID
	RecordRevisionCount int
}

func NewStore(pool postgres.DB) *Store {
	return NewStoreWithHooks(pool, currentStoreHooks())
}

func NewStoreWithHooks(pool postgres.DB, hooks StoreHooks) *Store {
	return &Store{
		pool:            pool,
		authStore:       authn.NewStore(pool),
		recordStore:     records.NewStore(),
		revisionsStore:  revisions.NewStore(),
		projectionStore: projections.NewStore(pool),
		linkStore:       links.NewStore(),
		hooks:           hooks,
	}
}

func (s *Store) GetRecordIncident(ctx context.Context, recordID uuid.UUID) (uuid.UUID, error) {
	var incidentID uuid.UUID
	if err := s.pool.QueryRow(ctx, `SELECT incident_id FROM records WHERE record_id = $1`, recordID).Scan(&incidentID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.UUID{}, ErrRecordNotFound
		}
		return uuid.UUID{}, fmt.Errorf("get record incident: %w", err)
	}
	return incidentID, nil
}

func (s *Store) QueryRows(ctx context.Context, incidentID uuid.UUID, query viewschema.QueryMeta) ([]map[string]any, error) {
	sqlText, args, err := buildTimelineQuerySQL(incidentID, query)
	if err != nil {
		return nil, err
	}
	rows, err := s.pool.Query(ctx, sqlText, args...)
	if err != nil {
		return nil, fmt.Errorf("list timeline projection rows: %w", err)
	}
	defer rows.Close()

	projectedRows := make([]projectedRecord, 0)
	for rows.Next() {
		projected, err := scanProjectedRecord(rows)
		if err != nil {
			return nil, err
		}
		projectedRows = append(projectedRows, projected)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate timeline projection rows: %w", err)
	}
	rows.Close()

	result := make([]map[string]any, 0, len(projectedRows))
	for index := range projectedRows {
		if err := hydrateProjectedCollections(ctx, s.pool, &projectedRows[index]); err != nil {
			return nil, err
		}
		result = append(result, BuildRow(projectedRows[index]))
	}
	return result, nil
}

func (s *Store) SnapshotRecordSubstrate(ctx context.Context, recordID uuid.UUID) (RecordSubstrateSnapshot, error) {
	row, err := sqlc.New(s.pool).GetTimelineProjectionRow(ctx, pgUUID(recordID))
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return RecordSubstrateSnapshot{}, ErrRecordNotFound
		}
		return RecordSubstrateSnapshot{}, fmt.Errorf("get timeline projection row: %w", err)
	}
	projected, err := projectedRecordFromSQL(row)
	if err != nil {
		return RecordSubstrateSnapshot{}, err
	}

	var revisionCount int
	if err := s.pool.QueryRow(ctx, `SELECT COUNT(*) FROM record_revisions WHERE record_id = $1`, recordID).Scan(&revisionCount); err != nil {
		return RecordSubstrateSnapshot{}, fmt.Errorf("count record revisions: %w", err)
	}

	return RecordSubstrateSnapshot{
		RecordID:            projected.RecordID,
		RowVersion:          projected.RowVersion,
		CaptureState:        projected.CaptureState,
		ReplacementRecordID: projected.ReplacementRecordID,
		RecordRevisionCount: revisionCount,
	}, nil
}

func (s *Store) CreateRow(ctx context.Context, actor authn.UserRecord, incidentID uuid.UUID, request CreateRequest, requestHash []byte, requestID string, now time.Time) (MutationResult, error) {
	scopeKey := incidentID.String() + ":" + TimelineViewSchemaID
	idempotencyKey := authn.RouteIdempotencyKey{
		RouteKey:    createRouteKey,
		ActorUserID: actor.ID,
		ScopeKey:    scopeKey,
		ClientTxnID: request.ClientTxnID,
	}
	if existing, err := s.authStore.GetRouteIdempotency(ctx, idempotencyKey); err == nil {
		markCreateTiming(ctx, "store_idempotency")
		if !hashesEqual(existing.RequestHash, requestHash) {
			return MutationResult{}, authn.ErrClientTxnConflict
		}
		payload, err := decodeStoredResponse(existing.ResponseJSON)
		if err != nil {
			return MutationResult{}, fmt.Errorf("decode replayed timeline create payload: %w", err)
		}
		recordID, err := extractUUIDFromPayload(payload, "row", "record_id")
		if err != nil {
			return MutationResult{}, err
		}
		return MutationResult{
			Payload:    payload,
			StatusCode: http.StatusOK,
			Replayed:   true,
			IncidentID: incidentID,
			RecordID:   recordID,
		}, nil
	} else if !errors.Is(err, authn.ErrNotFound) {
		return MutationResult{}, fmt.Errorf("query timeline create idempotency: %w", err)
	}
	markCreateTiming(ctx, "store_idempotency")

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return MutationResult{}, fmt.Errorf("begin timeline create transaction: %w", err)
	}
	markCreateTiming(ctx, "store_tx_begin")
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	recordID := uuid.New()
	changeSetID := uuid.New()
	current := sourceRecord{
		RecordID:        recordID,
		IncidentID:      incidentID,
		OccurredAt:      request.OccurredAt,
		Summary:         request.Summary,
		Details:         request.Details,
		SourceText:      request.SourceText,
		CaptureState:    InitialCaptureState(),
		RowVersion:      1,
		RecordedAt:      now.UTC(),
		EditedAt:        now.UTC(),
		CreatedByUserID: actor.ID,
		UpdatedByUserID: actor.ID,
	}
	if _, err := s.recordStore.InsertTx(ctx, tx, records.InsertParams{
		RecordID:        &recordID,
		IncidentID:      incidentID,
		RecordType:      "timeline_event",
		CreatedByUserID: actor.ID,
		CreatedAt:       now.UTC(),
		UpdatedByUserID: actor.ID,
		UpdatedAt:       now.UTC(),
		RowVersion:      1,
	}); err != nil {
		return MutationResult{}, err
	}
	if _, err := tx.Exec(ctx, `
INSERT INTO timeline_events (
    record_id,
    incident_id,
    occurred_at,
    summary,
    details,
    source_text,
    capture_state,
    row_version,
    recorded_at,
    edited_at,
    created_by_user_id,
    updated_by_user_id
)
VALUES ($1, $2, $3, $4, $5, $6, 'rough', 1, $7, $7, $8, $8)
`, current.RecordID, incidentID, request.OccurredAt, request.Summary, request.Details, request.SourceText, now.UTC(), actor.ID); err != nil {
		return MutationResult{}, fmt.Errorf("insert timeline record: %w", err)
	}
	markCreateTiming(ctx, "store_base_insert")

	if err := applyCreateMentionActionsTx(ctx, tx, actor.ID, current.IncidentID, current.RecordID, request.HostRefs, request.IdentityRefs, now.UTC()); err != nil {
		return MutationResult{}, err
	}
	if err := applyCreateTagActionsTx(ctx, tx, actor.ID, current.IncidentID, current.RecordID, request.Tags, now.UTC()); err != nil {
		return MutationResult{}, err
	}
	if err := applyAttachedEvidenceActionsTx(ctx, tx, actor.ID, current.IncidentID, current.RecordID, request.AttachedEvidence, now.UTC()); err != nil {
		return MutationResult{}, err
	}
	markCreateTiming(ctx, "store_collection_actions")

	projected := projectRecord(current, nil)
	if createRequestHasCollectionActions(request) {
		if err := hydrateProjectedCollections(ctx, tx, &projected); err != nil {
			return MutationResult{}, err
		}
	}
	markCreateTiming(ctx, "store_project_row")
	if _, err := s.revisionsStore.InsertChangeSetTx(ctx, tx, revisions.ChangeSetParams{
		ChangeSetID: &changeSetID,
		IncidentID:  incidentID,
		ActorUserID: actor.ID,
		Source:      createRouteKey,
		ClientTxnID: &request.ClientTxnID,
		RequestID:   &requestID,
		CreatedAt:   now.UTC(),
	}); err != nil {
		return MutationResult{}, err
	}

	afterRow := BuildRow(projected)
	afterVersion := versionID(current.RecordID, projected.RowVersion)
	if err := s.revisionsStore.InsertMutationTx(ctx, tx, revisions.MutationParams{
		ChangeSetID:    changeSetID,
		SequenceNo:     1,
		TargetKind:     "timeline_record",
		TargetID:       current.RecordID.String(),
		OperationKind:  "create",
		AfterVersionID: &afterVersion,
		AfterValue:     afterRow,
	}); err != nil {
		return MutationResult{}, err
	}
	if err := s.revisionsStore.InsertRecordRevisionTx(ctx, tx, revisions.RecordRevisionParams{
		ChangeSetID: changeSetID,
		RecordID:    current.RecordID,
		RowVersion:  projected.RowVersion,
		AfterValue:  afterRow,
	}); err != nil {
		return MutationResult{}, err
	}
	if err := s.projectionStore.UpsertTimelineRowTx(ctx, tx, projectionInput(projected)); err != nil {
		return MutationResult{}, err
	}
	markCreateTiming(ctx, "store_revision_projection")

	payload := BuildMutationPayload(projected, changeSetID)
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, idempotencyKey, nil, requestHash, http.StatusCreated, payload); err != nil {
		if authn.IsUniqueViolation(err) {
			return MutationResult{}, authn.ErrClientTxnConflict
		}
		return MutationResult{}, err
	}
	markCreateTiming(ctx, "store_route_idempotency")
	if err := s.beforeCommit(createRouteKey, current.RecordID); err != nil {
		return MutationResult{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return MutationResult{}, fmt.Errorf("commit timeline create transaction: %w", err)
	}
	markCreateTiming(ctx, "store_commit")
	return MutationResult{
		Payload:          payload,
		StatusCode:       http.StatusCreated,
		IncidentID:       incidentID,
		RecordID:         current.RecordID,
		ChangeSetID:      changeSetID,
		ClientTxnID:      request.ClientTxnID,
		RowVersion:       projected.RowVersion,
		ChangedFieldKeys: ComputeChangedFieldKeys(nil, projected),
		Row:              projected,
	}, nil
}

func createRequestHasCollectionActions(request CreateRequest) bool {
	return (request.HostRefs != nil && len(request.HostRefs.Actions) > 0) ||
		(request.IdentityRefs != nil && len(request.IdentityRefs.Actions) > 0) ||
		(request.Tags != nil && len(request.Tags.Actions) > 0) ||
		(request.AttachedEvidence != nil && len(request.AttachedEvidence.Actions) > 0)
}

func (s *Store) PatchRow(ctx context.Context, actor authn.UserRecord, recordID uuid.UUID, request PatchRequest, requestHash []byte, requestID string, now time.Time) (MutationResult, error) {
	return s.applyPatch(ctx, actor, recordID, request, requestHash, requestID, now, patchRouteKey)
}

func (s *Store) ResolveConflict(ctx context.Context, actor authn.UserRecord, recordID uuid.UUID, claims TimelineConflictTokenClaims, request ConflictResolveRequest, requestHash []byte, requestID string, now time.Time) (MutationResult, error) {
	if request.ResolutionKind == "keep_saved" {
		return s.clearConflict(ctx, actor, recordID, claims, request, requestHash)
	}
	if request.ResolvedChange == nil {
		return MutationResult{}, ErrNoEffectiveChange
	}
	patch := PatchRequest{
		ViewSchemaID:    TimelineViewSchemaID,
		BaseRowVersion:  claims.CurrentRowVersion,
		ClientTxnID:     request.ClientTxnID,
		CanonicalChange: []PatchChange{*request.ResolvedChange},
	}
	return s.applyPatch(ctx, actor, recordID, patch, requestHash, requestID, now, conflictResolveRouteKey)
}

func (s *Store) clearConflict(ctx context.Context, actor authn.UserRecord, recordID uuid.UUID, claims TimelineConflictTokenClaims, request ConflictResolveRequest, requestHash []byte) (MutationResult, error) {
	if claims.ViewSchemaID != TimelineViewSchemaID {
		return MutationResult{}, ErrRecordNotFound
	}
	idempotencyKey := authn.RouteIdempotencyKey{
		RouteKey:    conflictResolveRouteKey,
		ActorUserID: actor.ID,
		ScopeKey:    recordID.String(),
		ClientTxnID: request.ClientTxnID,
	}
	if existing, err := s.authStore.GetRouteIdempotency(ctx, idempotencyKey); err == nil {
		if !hashesEqual(existing.RequestHash, requestHash) {
			return MutationResult{}, authn.ErrClientTxnConflict
		}
		payload, err := decodeStoredResponse(existing.ResponseJSON)
		if err != nil {
			return MutationResult{}, fmt.Errorf("decode replayed timeline conflict clear payload: %w", err)
		}
		return MutationResult{
			Payload:    payload,
			StatusCode: http.StatusOK,
			Replayed:   true,
			RecordID:   recordID,
		}, nil
	} else if !errors.Is(err, authn.ErrNotFound) {
		return MutationResult{}, fmt.Errorf("query timeline conflict clear idempotency: %w", err)
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return MutationResult{}, fmt.Errorf("begin timeline conflict clear transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	current, err := loadSourceRecordTx(ctx, tx, recordID)
	if err != nil {
		return MutationResult{}, err
	}
	projected := projectRecord(current, nil)
	if err := hydrateProjectedCollections(ctx, tx, &projected); err != nil {
		return MutationResult{}, err
	}
	payload := map[string]any{
		"view_schema_id": TimelineViewSchemaID,
		"row":            BuildRow(projected),
	}
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, idempotencyKey, nil, requestHash, http.StatusOK, payload); err != nil {
		if authn.IsUniqueViolation(err) {
			return MutationResult{}, authn.ErrClientTxnConflict
		}
		return MutationResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return MutationResult{}, fmt.Errorf("commit timeline conflict clear transaction: %w", err)
	}
	return MutationResult{
		Payload:     payload,
		StatusCode:  http.StatusOK,
		IncidentID:  current.IncidentID,
		RecordID:    recordID,
		ClientTxnID: request.ClientTxnID,
		RowVersion:  current.RowVersion,
		Row:         projected,
	}, nil
}

func (s *Store) applyPatch(ctx context.Context, actor authn.UserRecord, recordID uuid.UUID, request PatchRequest, requestHash []byte, requestID string, now time.Time, routeKey string) (MutationResult, error) {
	idempotencyKey := authn.RouteIdempotencyKey{
		RouteKey:    routeKey,
		ActorUserID: actor.ID,
		ScopeKey:    recordID.String(),
		ClientTxnID: request.ClientTxnID,
	}
	if existing, err := s.authStore.GetRouteIdempotency(ctx, idempotencyKey); err == nil {
		if !hashesEqual(existing.RequestHash, requestHash) {
			return MutationResult{}, authn.ErrClientTxnConflict
		}
		payload, err := decodeStoredResponse(existing.ResponseJSON)
		if err != nil {
			return MutationResult{}, fmt.Errorf("decode replayed timeline patch payload: %w", err)
		}
		return MutationResult{
			Payload:    payload,
			StatusCode: http.StatusOK,
			Replayed:   true,
			RecordID:   recordID,
		}, nil
	} else if !errors.Is(err, authn.ErrNotFound) {
		return MutationResult{}, fmt.Errorf("query timeline patch idempotency: %w", err)
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return MutationResult{}, fmt.Errorf("begin timeline patch transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	current, err := loadSourceRecordTx(ctx, tx, recordID)
	if err != nil {
		return MutationResult{}, err
	}
	if current.RowVersion < request.BaseRowVersion {
		return MutationResult{}, &RowVersionConflictError{
			RecordID:          recordID,
			BaseRowVersion:    request.BaseRowVersion,
			CurrentRowVersion: current.RowVersion,
		}
	}
	if current.RowVersion > request.BaseRowVersion {
		window, err := loadPatchConflictWindowTx(ctx, tx, recordID, request.BaseRowVersion, current.RowVersion)
		if err != nil {
			return MutationResult{}, err
		}
		if change, changed, ok := overlappingPatchChange(request.CanonicalChange, window.ChangedFields); ok {
			currentProjected := projectRecord(current, nil)
			if err := hydrateProjectedCollections(ctx, tx, &currentProjected); err != nil {
				return MutationResult{}, err
			}
			conflict, err := buildSameFieldConflict(recordID, currentProjected, request.BaseRowVersion, requestHash, window, change, changed)
			if err != nil {
				return MutationResult{}, err
			}
			return MutationResult{}, conflict
		}
	}
	if current.CaptureState == "superseded" {
		return MutationResult{}, newIllegalTransitionError("superseded_terminal", current.CaptureState, captureStateEnriched)
	}

	next := current
	mentionChanged := false
	tagChanged := false
	evidenceChanged := false
	for _, change := range request.CanonicalChange {
		switch change.FieldKey {
		case "timeline.occurred_at":
			next.OccurredAt = change.OccurredAt
		case "timeline.summary":
			next.Summary = change.TextValue
		case "timeline.details":
			next.Details = change.TextValue
		case "timeline.source_text":
			next.SourceText = change.TextValue
		case "timeline.host_refs", "timeline.identity_refs":
			if change.ActionPayload != nil && len(change.ActionPayload.Actions) > 0 {
				mentionChanged = true
			}
		case "timeline.tags":
			if change.ActionPayload != nil && len(change.ActionPayload.Actions) > 0 {
				tagChanged = true
			}
		case "timeline.attached_evidence_ids":
			if change.ActionPayload != nil && len(change.ActionPayload.Actions) > 0 {
				evidenceChanged = true
			}
		}
	}

	beforeProjected := projectRecord(current, nil)
	if err := hydrateProjectedCollections(ctx, tx, &beforeProjected); err != nil {
		return MutationResult{}, err
	}
	materialChanged := hasMaterialChange(current, next)
	if mentionChanged {
		if err := applyPatchMentionActionsTx(ctx, tx, actor, current.IncidentID, recordID, request.CanonicalChange, now.UTC()); err != nil {
			return MutationResult{}, err
		}
	}
	if tagChanged {
		if err := applyPatchTagActionsTx(ctx, tx, actor.ID, current.IncidentID, recordID, request.CanonicalChange, now.UTC()); err != nil {
			return MutationResult{}, err
		}
	}
	if evidenceChanged {
		if err := applyPatchAttachedEvidenceActionsTx(ctx, tx, actor.ID, current.IncidentID, recordID, request.CanonicalChange, now.UTC()); err != nil {
			return MutationResult{}, err
		}
	}
	if !materialChanged && !mentionChanged && !tagChanged && !evidenceChanged {
		return MutationResult{}, ErrNoEffectiveChange
	}
	nextState, err := CaptureStateAfterMaterialPatch(current.CaptureState)
	if err != nil {
		return MutationResult{}, err
	}
	next.CaptureState = nextState
	next.RowVersion, err = s.recordStore.AdvanceVersionTx(ctx, tx, current.RecordID, actor.ID, now.UTC())
	if err != nil {
		return MutationResult{}, err
	}
	next.EditedAt = now.UTC()
	next.UpdatedByUserID = actor.ID
	if current.CaptureState == captureStateReviewed {
		next.ReviewedAt = nil
		next.ReviewedByUserID = nil
	}

	if err := tx.QueryRow(ctx, `
UPDATE timeline_events
   SET occurred_at = $2,
       summary = $3,
       details = $4,
       source_text = $5,
       capture_state = $6,
       row_version = $7,
       edited_at = $8,
       updated_by_user_id = $9,
       reviewed_at = $10,
       reviewed_by_user_id = $11
 WHERE record_id = $1
RETURNING recorded_at
`, recordID, next.OccurredAt, next.Summary, next.Details, next.SourceText, next.CaptureState, next.RowVersion, next.EditedAt, actor.ID, next.ReviewedAt, next.ReviewedByUserID).Scan(&next.RecordedAt); err != nil {
		return MutationResult{}, fmt.Errorf("update timeline record: %w", err)
	}

	afterProjected := projectRecord(next, nil)
	if err := hydrateProjectedCollections(ctx, tx, &afterProjected); err != nil {
		return MutationResult{}, err
	}
	changeSetID, err := s.revisionsStore.InsertChangeSetTx(ctx, tx, revisions.ChangeSetParams{
		IncidentID:  current.IncidentID,
		ActorUserID: actor.ID,
		Source:      routeKey,
		ClientTxnID: &request.ClientTxnID,
		RequestID:   &requestID,
		CreatedAt:   now.UTC(),
	})
	if err != nil {
		return MutationResult{}, err
	}

	beforeRow := BuildRow(beforeProjected)
	afterRow := BuildRow(afterProjected)
	beforeVersion := versionID(current.RecordID, current.RowVersion)
	afterVersion := versionID(next.RecordID, next.RowVersion)
	if err := s.revisionsStore.InsertMutationTx(ctx, tx, revisions.MutationParams{
		ChangeSetID:     changeSetID,
		SequenceNo:      1,
		TargetKind:      "timeline_record",
		TargetID:        current.RecordID.String(),
		OperationKind:   "patch",
		BeforeVersionID: &beforeVersion,
		AfterVersionID:  &afterVersion,
		BeforeValue:     beforeRow,
		AfterValue:      afterRow,
	}); err != nil {
		return MutationResult{}, err
	}
	if err := s.revisionsStore.InsertRecordRevisionTx(ctx, tx, revisions.RecordRevisionParams{
		ChangeSetID: changeSetID,
		RecordID:    current.RecordID,
		RowVersion:  next.RowVersion,
		BeforeValue: beforeRow,
		AfterValue:  afterRow,
	}); err != nil {
		return MutationResult{}, err
	}
	if err := s.projectionStore.UpsertTimelineRowTx(ctx, tx, projectionInput(afterProjected)); err != nil {
		return MutationResult{}, err
	}

	payload := BuildMutationPayload(afterProjected, changeSetID)
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, idempotencyKey, nil, requestHash, http.StatusOK, payload); err != nil {
		if authn.IsUniqueViolation(err) {
			return MutationResult{}, authn.ErrClientTxnConflict
		}
		return MutationResult{}, err
	}
	if err := s.beforeCommit(routeKey, recordID); err != nil {
		return MutationResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return MutationResult{}, fmt.Errorf("commit timeline patch transaction: %w", err)
	}
	return MutationResult{
		Payload:          payload,
		StatusCode:       http.StatusOK,
		IncidentID:       current.IncidentID,
		RecordID:         recordID,
		ChangeSetID:      changeSetID,
		ClientTxnID:      request.ClientTxnID,
		RowVersion:       afterProjected.RowVersion,
		ChangedFieldKeys: ComputeChangedFieldKeys(&beforeProjected, afterProjected),
		Row:              afterProjected,
	}, nil
}

func loadPatchConflictWindowTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, baseRowVersion int64, currentRowVersion int64) (patchConflictWindow, error) {
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
		return patchConflictWindow{}, fmt.Errorf("query timeline patch conflict window: %w", err)
	}
	defer rows.Close()

	window := patchConflictWindow{
		ChangedFields: make(map[string]patchChangedField),
	}
	for rows.Next() {
		var (
			rowVersion int64
			beforeJSON []byte
			afterJSON  []byte
			actorID    uuid.UUID
			createdAt  time.Time
		)
		if err := rows.Scan(&rowVersion, &beforeJSON, &afterJSON, &actorID, &createdAt); err != nil {
			return patchConflictWindow{}, fmt.Errorf("scan timeline patch conflict window: %w", err)
		}

		if rowVersion == baseRowVersion {
			baseRow, ok := decodeRevisionRow(afterJSON)
			if !ok {
				return patchConflictWindow{}, newRowVersionConflict(recordID, baseRowVersion, currentRowVersion)
			}
			window.BaseRow = baseRow
			continue
		}

		beforeRow, beforeOK := decodeRevisionRow(beforeJSON)
		afterRow, afterOK := decodeRevisionRow(afterJSON)
		if !beforeOK || !afterOK {
			return patchConflictWindow{}, newRowVersionConflict(recordID, baseRowVersion, currentRowVersion)
		}
		for _, fieldKey := range changedRevisionWritableFieldKeys(beforeRow, afterRow) {
			window.ChangedFields[fieldKey] = patchChangedField{
				FieldKey:        fieldKey,
				ServerUpdatedBy: actorID,
				ServerUpdatedAt: createdAt.UTC(),
			}
		}
	}
	if err := rows.Err(); err != nil {
		return patchConflictWindow{}, fmt.Errorf("iterate timeline patch conflict window: %w", err)
	}
	if window.BaseRow == nil {
		return patchConflictWindow{}, newRowVersionConflict(recordID, baseRowVersion, currentRowVersion)
	}
	return window, nil
}

func newRowVersionConflict(recordID uuid.UUID, baseRowVersion int64, currentRowVersion int64) *RowVersionConflictError {
	return &RowVersionConflictError{
		RecordID:          recordID,
		BaseRowVersion:    baseRowVersion,
		CurrentRowVersion: currentRowVersion,
	}
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

func changedRevisionWritableFieldKeys(beforeRow map[string]any, afterRow map[string]any) []string {
	beforeCells, _ := beforeRow["cells"].(map[string]any)
	afterCells, _ := afterRow["cells"].(map[string]any)
	changed := make([]string, 0)
	for fieldKey, afterCell := range afterCells {
		field, ok := viewschema.LookupField(TimelineViewSchemaID, fieldKey)
		if !ok || !field.Writable {
			continue
		}
		if !reflect.DeepEqual(beforeCells[fieldKey], afterCell) {
			changed = append(changed, fieldKey)
		}
	}
	sort.Strings(changed)
	return changed
}

func overlappingPatchChange(changes []PatchChange, changedFields map[string]patchChangedField) (PatchChange, patchChangedField, bool) {
	for _, change := range changes {
		changed, ok := changedFields[change.FieldKey]
		if ok {
			return change, changed, true
		}
	}
	return PatchChange{}, patchChangedField{}, false
}

func buildSameFieldConflict(recordID uuid.UUID, current projectedRecord, baseRowVersion int64, requestHash []byte, window patchConflictWindow, change PatchChange, changed patchChangedField) (*SameFieldConflictError, error) {
	baseValue, ok := rowCellValue(window.BaseRow, change.FieldKey)
	if !ok {
		return nil, newRowVersionConflict(recordID, baseRowVersion, current.RowVersion)
	}
	serverValue, ok := rowCellValue(BuildRow(current), change.FieldKey)
	if !ok {
		return nil, newRowVersionConflict(recordID, baseRowVersion, current.RowVersion)
	}
	clientValue, err := patchClientConflictValue(change, baseValue, requestHash)
	if err != nil {
		return nil, newRowVersionConflict(recordID, baseRowVersion, current.RowVersion)
	}

	field, _ := viewschema.LookupField(TimelineViewSchemaID, change.FieldKey)
	conflictClass := field.ConflictResolutionClass
	if conflictClass == "" {
		conflictClass = "atomic_replace"
	}
	token := conflictToken(recordID, change.FieldKey, baseRowVersion, current.RowVersion, requestHash)
	return &SameFieldConflictError{
		Conflict: map[string]any{
			"conflict_token":            token,
			"record_id":                 recordID.String(),
			"field_key":                 change.FieldKey,
			"conflict_resolution_class": conflictClass,
			"base_row_version":          baseRowVersion,
			"current_row_version":       current.RowVersion,
			"client_value":              clientValue,
			"server_value":              serverValue,
			"server_updated_by":         changed.ServerUpdatedBy.String(),
			"server_updated_at":         formatTimestamp(changed.ServerUpdatedAt),
			"base_value":                baseValue,
		},
	}, nil
}

func conflictToken(recordID uuid.UUID, fieldKey string, baseRowVersion int64, currentRowVersion int64, requestHash []byte) string {
	field, _ := viewschema.LookupField(TimelineViewSchemaID, fieldKey)
	conflictClass := field.ConflictResolutionClass
	if conflictClass == "" {
		conflictClass = "atomic_replace"
	}
	claims := TimelineConflictTokenClaims{
		Version:                 1,
		RecordID:                recordID.String(),
		ViewSchemaID:            TimelineViewSchemaID,
		FieldKey:                fieldKey,
		ConflictResolutionClass: conflictClass,
		BaseRowVersion:          baseRowVersion,
		CurrentRowVersion:       currentRowVersion,
		RequestHash:             base64.RawURLEncoding.EncodeToString(requestHash),
	}
	claims.Signature = timelineConflictTokenSignature(claims)
	data, _ := json.Marshal(claims)
	return base64.RawURLEncoding.EncodeToString(data)
}

func ParseConflictToken(token string) (TimelineConflictTokenClaims, bool) {
	data, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return TimelineConflictTokenClaims{}, false
	}
	var claims TimelineConflictTokenClaims
	if err := json.Unmarshal(data, &claims); err != nil {
		return TimelineConflictTokenClaims{}, false
	}
	if claims.Version != 1 || claims.Signature == "" || claims.Signature != timelineConflictTokenSignature(claims) {
		return TimelineConflictTokenClaims{}, false
	}
	recordID, err := uuid.Parse(claims.RecordID)
	if err != nil || recordID == uuid.Nil {
		return TimelineConflictTokenClaims{}, false
	}
	if claims.ViewSchemaID != TimelineViewSchemaID ||
		claims.FieldKey == "" ||
		claims.BaseRowVersion < 1 ||
		claims.CurrentRowVersion < claims.BaseRowVersion ||
		claims.RequestHash == "" {
		return TimelineConflictTokenClaims{}, false
	}
	return claims, true
}

func timelineConflictTokenSignature(claims TimelineConflictTokenClaims) string {
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

func rowCellValue(row map[string]any, fieldKey string) (any, bool) {
	cells, ok := row["cells"].(map[string]any)
	if !ok {
		return nil, false
	}
	cell, ok := cells[fieldKey].(map[string]any)
	if !ok {
		return nil, false
	}
	value, ok := cell["value"]
	return value, ok
}

func patchClientConflictValue(change PatchChange, baseValue any, requestHash []byte) (any, error) {
	if change.ActionPayload == nil {
		return canonicalChangeValue(change), nil
	}
	return applyCollectionConflictActions(change.FieldKey, baseValue, change.ActionPayload, requestHash)
}

func applyCollectionConflictActions(fieldKey string, baseValue any, payload *CollectionActionPayload, requestHash []byte) (map[string]any, error) {
	ordered, items, ok := cloneCollectionConflictValue(baseValue)
	if !ok {
		return nil, fmt.Errorf("invalid base collection value for %s", fieldKey)
	}
	for index, action := range payload.Actions {
		switch action.Op {
		case "add_token":
			items = append(items, newClientCollectionItem(fieldKey, action, requestHash, index, false))
		case "add_resolved_ref":
			items = append(items, newClientCollectionItem(fieldKey, action, requestHash, index, true))
		case "add_record_ref":
			items = append(items, newClientCollectionItem(fieldKey, action, requestHash, index, true))
		case "resolve_item":
			if item := findCollectionItem(items, action.ItemRef); item != nil {
				item["item_kind"] = "resolved_ref"
				if action.ResolvedRecord != nil {
					item["resolved_record_id"] = action.ResolvedRecord.String()
				}
				removeResolutionMetadata(item, false)
			}
		case "dismiss_item":
			items = removeCollectionItem(items, action.ItemRef)
		case "revert_to_unresolved":
			if item := findCollectionItem(items, action.ItemRef); item != nil {
				item["item_kind"] = "unresolved_mention"
				removeResolutionMetadata(item, true)
			}
		case "remove_record_ref":
			items = removeCollectionItem(items, action.ItemRef)
		default:
			return nil, fmt.Errorf("unsupported collection action: %s", action.Op)
		}
	}
	if !ordered {
		sort.SliceStable(items, func(left int, right int) bool {
			return collectionSortKey(items[left]) < collectionSortKey(items[right])
		})
	}
	return collectionValue(ordered, items), nil
}

func cloneCollectionConflictValue(value any) (bool, []map[string]any, bool) {
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
			items = append(items, cloneMap(item))
		}
	case []map[string]any:
		for _, item := range rawItems {
			items = append(items, cloneMap(item))
		}
	default:
		return false, nil, false
	}
	return ordered, items, true
}

func cloneMap(source map[string]any) map[string]any {
	cloned := make(map[string]any, len(source))
	for key, value := range source {
		cloned[key] = value
	}
	return cloned
}

func newClientCollectionItem(fieldKey string, action CollectionAction, requestHash []byte, actionIndex int, resolved bool) map[string]any {
	rawText := action.RawText
	displayText := action.RawText
	if fieldKey == "timeline.tags" {
		rawText = action.NormalizedText
		displayText = action.NormalizedText
	}
	item := map[string]any{
		"item_ref":     clientCollectionItemRef(fieldKey, action, requestHash, actionIndex),
		"display_text": displayText,
		"raw_text":     rawText,
	}
	if fieldKey == "timeline.tags" {
		item["item_kind"] = "tag"
		return item
	}
	if fieldKey == "timeline.attached_evidence_ids" {
		item["item_kind"] = "record_ref"
		if action.LinkedRecordID != nil {
			item["item_ref"] = "record_ref:" + action.LinkedRecordID.String()
			item["linked_record_id"] = action.LinkedRecordID.String()
			item["display_text"] = action.LinkedRecordID.String()
		}
		return item
	}

	item["entity_type"] = collectionEntityType(fieldKey)
	if resolved {
		item["item_kind"] = "resolved_ref"
		if action.ResolvedRecord != nil {
			item["resolved_record_id"] = action.ResolvedRecord.String()
		}
		return item
	}
	item["item_kind"] = "unresolved_mention"
	return item
}

func clientCollectionItemRef(fieldKey string, action CollectionAction, requestHash []byte, actionIndex int) string {
	sum := hashRequestPayload(map[string]any{
		"request_hash":     base64.RawURLEncoding.EncodeToString(requestHash),
		"field_key":        fieldKey,
		"action_index":     actionIndex,
		"op":               action.Op,
		"raw_text":         action.NormalizedText,
		"item_ref":         action.ItemRef,
		"linked_record_id": formatUUIDPointer(action.LinkedRecordID),
	})
	token := base64.RawURLEncoding.EncodeToString(sum)
	if len(token) > 18 {
		token = token[:18]
	}
	return "client:" + token
}

func collectionEntityType(fieldKey string) string {
	if fieldKey == "timeline.identity_refs" {
		return "identity"
	}
	return "host"
}

func findCollectionItem(items []map[string]any, itemRef string) map[string]any {
	for _, item := range items {
		if item["item_ref"] == itemRef {
			return item
		}
	}
	return nil
}

func removeCollectionItem(items []map[string]any, itemRef string) []map[string]any {
	for index, item := range items {
		if item["item_ref"] == itemRef {
			return append(items[:index], items[index+1:]...)
		}
	}
	return items
}

func removeResolutionMetadata(item map[string]any, removeResolvedID bool) {
	if removeResolvedID {
		delete(item, "resolved_record_id")
	}
	delete(item, "resolution_method")
	delete(item, "auto_resolved")
	delete(item, "provenance")
	delete(item, "confidence")
	delete(item, "matched_alias_text")
}

func collectionSortKey(item map[string]any) string {
	for _, key := range []string{"display_text", "raw_text", "item_ref"} {
		if value, ok := item[key].(string); ok {
			return value
		}
	}
	return ""
}

func (s *Store) MarkReviewed(ctx context.Context, actor authn.UserRecord, recordID uuid.UUID, request ActionRequest, requestHash []byte, requestID string, now time.Time) (MutationResult, error) {
	return s.applyAction(ctx, actor, reviewRouteKey, recordID, request.BaseRowVersion, request.ClientTxnID, requestHash, requestID, now, request.Reason, nil, func(current sourceRecord) (sourceRecord, *links.SupersedesLink, *string, error) {
		if !CaptureStateAllowsMarkReviewed(current.CaptureState) {
			return sourceRecord{}, nil, nil, newIllegalTransitionError("mark_reviewed_not_allowed", current.CaptureState, captureStateReviewed)
		}
		next := current
		next.CaptureState = captureStateReviewed
		next.RowVersion = current.RowVersion + 1
		next.EditedAt = now.UTC()
		next.UpdatedByUserID = actor.ID
		next.ReviewedAt = &next.EditedAt
		reviewerID := actor.ID
		next.ReviewedByUserID = &reviewerID
		return next, nil, request.Reason, nil
	})
}

func (s *Store) Supersede(ctx context.Context, actor authn.UserRecord, recordID uuid.UUID, request SupersedeRequest, requestHash []byte, requestID string, now time.Time) (MutationResult, error) {
	return s.applyAction(ctx, actor, supersedeRouteKey, recordID, request.BaseRowVersion, request.ClientTxnID, requestHash, requestID, now, &request.Reason, request.ReplacementRecordID, func(current sourceRecord) (sourceRecord, *links.SupersedesLink, *string, error) {
		if !CaptureStateAllowsSupersede(current.CaptureState) {
			return sourceRecord{}, nil, nil, newIllegalTransitionError("supersede_not_allowed", current.CaptureState, captureStateSuperseded)
		}

		next := current
		next.CaptureState = captureStateSuperseded
		next.RowVersion = current.RowVersion + 1
		next.EditedAt = now.UTC()
		next.UpdatedByUserID = actor.ID
		next.SupersededAt = &next.EditedAt
		supersededBy := actor.ID
		next.SupersededByUserID = &supersededBy
		return next, nil, &request.Reason, nil
	})
}

func (s *Store) applyAction(
	ctx context.Context,
	actor authn.UserRecord,
	routeKey string,
	recordID uuid.UUID,
	baseRowVersion int64,
	clientTxnID string,
	requestHash []byte,
	requestID string,
	now time.Time,
	reason *string,
	replacementRecordID *uuid.UUID,
	prepare func(sourceRecord) (sourceRecord, *links.SupersedesLink, *string, error),
) (MutationResult, error) {
	idempotencyKey := authn.RouteIdempotencyKey{
		RouteKey:    routeKey,
		ActorUserID: actor.ID,
		ScopeKey:    recordID.String(),
		ClientTxnID: clientTxnID,
	}
	if existing, err := s.authStore.GetRouteIdempotency(ctx, idempotencyKey); err == nil {
		if !hashesEqual(existing.RequestHash, requestHash) {
			return MutationResult{}, authn.ErrClientTxnConflict
		}
		payload, err := decodeStoredResponse(existing.ResponseJSON)
		if err != nil {
			return MutationResult{}, fmt.Errorf("decode replayed timeline action payload: %w", err)
		}
		return MutationResult{
			Payload:    payload,
			StatusCode: http.StatusOK,
			Replayed:   true,
			RecordID:   recordID,
		}, nil
	} else if !errors.Is(err, authn.ErrNotFound) {
		return MutationResult{}, fmt.Errorf("query timeline action idempotency: %w", err)
	}

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return MutationResult{}, fmt.Errorf("begin timeline action transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	current, err := loadSourceRecordTx(ctx, tx, recordID)
	if err != nil {
		return MutationResult{}, err
	}
	if current.RowVersion != baseRowVersion {
		return MutationResult{}, ErrRowVersionConflict
	}

	next, _, effectiveReason, err := prepare(current)
	if err != nil {
		return MutationResult{}, err
	}

	var validatedReplacementID *uuid.UUID
	if routeKey == supersedeRouteKey {
		if err := validateSupersedeReplacementTx(ctx, tx, current, replacementRecordID); err != nil {
			return MutationResult{}, err
		}
		if replacementRecordID != nil {
			replacementID := *replacementRecordID
			validatedReplacementID = &replacementID
		}
	} else if replacementRecordID != nil {
		replacementID := *replacementRecordID
		validatedReplacementID = &replacementID
	}
	next.RowVersion, err = s.recordStore.AdvanceVersionTx(ctx, tx, current.RecordID, actor.ID, now.UTC())
	if err != nil {
		return MutationResult{}, err
	}

	beforeProjected := projectRecord(current, nil)
	if err := hydrateProjectedCollections(ctx, tx, &beforeProjected); err != nil {
		return MutationResult{}, err
	}
	if err := tx.QueryRow(ctx, `
UPDATE timeline_events
   SET capture_state = $2,
       row_version = $3,
       edited_at = $4,
       updated_by_user_id = $5,
       reviewed_at = $6,
       reviewed_by_user_id = $7,
       superseded_at = $8,
       superseded_by_user_id = $9
 WHERE record_id = $1
RETURNING recorded_at
`, recordID, next.CaptureState, next.RowVersion, next.EditedAt, actor.ID, next.ReviewedAt, next.ReviewedByUserID, next.SupersededAt, next.SupersededByUserID).Scan(&next.RecordedAt); err != nil {
		return MutationResult{}, fmt.Errorf("update timeline action state: %w", err)
	}

	var insertedLink *links.SupersedesLink
	if routeKey == supersedeRouteKey && validatedReplacementID != nil {
		link, err := s.linkStore.InsertSupersedesTx(ctx, tx, current.IncidentID, *validatedReplacementID, current.RecordID, actor.ID, now.UTC())
		if err != nil {
			if isRecordLinkConflict(err) {
				return MutationResult{}, newIllegalTransitionError("supersede_not_allowed", current.CaptureState, captureStateSuperseded, supersedeGuardTargetMustNotHaveActiveReplacement)
			}
			return MutationResult{}, err
		}
		insertedLink = &link
	}

	afterProjected := projectRecord(next, validatedReplacementID)
	if err := hydrateProjectedCollections(ctx, tx, &afterProjected); err != nil {
		return MutationResult{}, err
	}
	changeSetID, err := s.revisionsStore.InsertChangeSetTx(ctx, tx, revisions.ChangeSetParams{
		IncidentID:  current.IncidentID,
		ActorUserID: actor.ID,
		Source:      routeKey,
		Reason:      effectiveReason,
		ClientTxnID: &clientTxnID,
		RequestID:   &requestID,
		CreatedAt:   now.UTC(),
	})
	if err != nil {
		return MutationResult{}, err
	}

	beforeRow := BuildRow(beforeProjected)
	afterRow := BuildRow(afterProjected)
	beforeVersion := versionID(current.RecordID, current.RowVersion)
	afterVersion := versionID(next.RecordID, next.RowVersion)
	if err := s.revisionsStore.InsertMutationTx(ctx, tx, revisions.MutationParams{
		ChangeSetID:     changeSetID,
		SequenceNo:      1,
		TargetKind:      "timeline_record",
		TargetID:        current.RecordID.String(),
		OperationKind:   "patch",
		BeforeVersionID: &beforeVersion,
		AfterVersionID:  &afterVersion,
		BeforeValue:     beforeRow,
		AfterValue:      afterRow,
	}); err != nil {
		return MutationResult{}, err
	}
	if insertedLink != nil {
		linkAfter := map[string]any{
			"record_link_id": insertedLink.RecordLinkID.String(),
			"incident_id":    insertedLink.IncidentID.String(),
			"src_record_id":  insertedLink.SrcRecordID.String(),
			"dst_record_id":  insertedLink.DstRecordID.String(),
			"link_type":      "supersedes",
		}
		if err := s.revisionsStore.InsertMutationTx(ctx, tx, revisions.MutationParams{
			ChangeSetID:    changeSetID,
			SequenceNo:     2,
			TargetKind:     "record_link",
			TargetID:       insertedLink.RecordLinkID.String(),
			OperationKind:  "create",
			AfterVersionID: nil,
			AfterValue:     linkAfter,
		}); err != nil {
			return MutationResult{}, err
		}
	}
	if err := s.revisionsStore.InsertRecordRevisionTx(ctx, tx, revisions.RecordRevisionParams{
		ChangeSetID: changeSetID,
		RecordID:    current.RecordID,
		RowVersion:  next.RowVersion,
		BeforeValue: beforeRow,
		AfterValue:  afterRow,
	}); err != nil {
		return MutationResult{}, err
	}
	if err := s.projectionStore.UpsertTimelineRowTx(ctx, tx, projectionInput(afterProjected)); err != nil {
		return MutationResult{}, err
	}

	payload := BuildActionPayload(afterProjected, changeSetID, effectiveReason)
	if err := authn.InsertRouteIdempotencyPayload(ctx, tx, idempotencyKey, nil, requestHash, http.StatusOK, payload); err != nil {
		if authn.IsUniqueViolation(err) {
			return MutationResult{}, authn.ErrClientTxnConflict
		}
		return MutationResult{}, err
	}
	if err := s.beforeCommit(routeKey, recordID); err != nil {
		return MutationResult{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return MutationResult{}, fmt.Errorf("commit timeline action transaction: %w", err)
	}
	return MutationResult{
		Payload:          payload,
		StatusCode:       http.StatusOK,
		IncidentID:       current.IncidentID,
		RecordID:         recordID,
		ChangeSetID:      changeSetID,
		ClientTxnID:      clientTxnID,
		RowVersion:       afterProjected.RowVersion,
		ChangedFieldKeys: ComputeChangedFieldKeys(&beforeProjected, afterProjected),
		Row:              afterProjected,
	}, nil
}

func validateSupersedeReplacementTx(ctx context.Context, tx pgx.Tx, current sourceRecord, replacementRecordID *uuid.UUID) error {
	if replacementRecordID == nil {
		return nil
	}

	guards := make([]string, 0, 3)
	if *replacementRecordID == current.RecordID {
		guards = append(guards, supersedeGuardReplacementDifferent)
	}

	targetHasActiveReplacement, err := hasActiveIncomingSupersedesLinkTx(ctx, tx, current.IncidentID, current.RecordID)
	if err != nil {
		return err
	}
	if targetHasActiveReplacement {
		guards = append(guards, supersedeGuardTargetMustNotHaveActiveReplacement)
	}

	if *replacementRecordID != current.RecordID {
		replacement, err := loadSourceRecordTx(ctx, tx, *replacementRecordID)
		if errors.Is(err, ErrRecordNotFound) {
			guards = append(guards, supersedeGuardReplacementVisibleActiveSameIncident)
		} else if err != nil {
			return err
		} else {
			if replacement.IncidentID != current.IncidentID {
				guards = append(guards, supersedeGuardReplacementVisibleActiveSameIncident)
			}
			if replacement.CaptureState == captureStateSuperseded {
				guards = append(guards, supersedeGuardReplacementNotSuperseded)
			}
		}
	}

	if len(guards) > 0 {
		return newIllegalTransitionError("supersede_not_allowed", current.CaptureState, captureStateSuperseded, guards...)
	}
	return nil
}

func hasActiveIncomingSupersedesLinkTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID) (bool, error) {
	var linkID uuid.UUID
	if err := tx.QueryRow(ctx, `
SELECT record_link_id
  FROM record_links
 WHERE incident_id = $1
   AND dst_record_id = $2
   AND link_type = 'supersedes'
   AND deleted_at IS NULL
 ORDER BY created_at DESC, record_link_id DESC
 LIMIT 1
 FOR UPDATE
`, incidentID, recordID).Scan(&linkID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, fmt.Errorf("query active incoming supersedes link: %w", err)
	}
	return true, nil
}

func loadSourceRecordTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (sourceRecord, error) {
	row := tx.QueryRow(ctx, `
SELECT
    e.record_id,
    e.incident_id,
    e.occurred_at,
    e.summary,
    e.details,
    e.source_text,
    e.capture_state,
    r.row_version,
    e.recorded_at,
    e.edited_at,
    r.created_by_user_id,
    r.updated_by_user_id,
    e.reviewed_by_user_id,
    e.reviewed_at,
    e.superseded_by_user_id,
    e.superseded_at
FROM timeline_events e
JOIN records r
  ON r.record_id = e.record_id
 AND r.record_type = 'timeline_event'
 AND r.deleted_at IS NULL
WHERE e.record_id = $1
FOR UPDATE OF e, r
`, recordID)

	var record sourceRecord
	if err := row.Scan(
		&record.RecordID,
		&record.IncidentID,
		&record.OccurredAt,
		&record.Summary,
		&record.Details,
		&record.SourceText,
		&record.CaptureState,
		&record.RowVersion,
		&record.RecordedAt,
		&record.EditedAt,
		&record.CreatedByUserID,
		&record.UpdatedByUserID,
		&record.ReviewedByUserID,
		&record.ReviewedAt,
		&record.SupersededByUserID,
		&record.SupersededAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return sourceRecord{}, ErrRecordNotFound
		}
		return sourceRecord{}, fmt.Errorf("load timeline record: %w", err)
	}
	return record, nil
}

func projectRecord(record sourceRecord, replacementRecordID *uuid.UUID) projectedRecord {
	sortTS := record.RecordedAt.UTC()
	if record.OccurredAt != nil {
		sortTS = record.OccurredAt.UTC()
	}

	var occurredDay *time.Time
	if record.OccurredAt != nil {
		day := time.Date(record.OccurredAt.UTC().Year(), record.OccurredAt.UTC().Month(), record.OccurredAt.UTC().Day(), 0, 0, 0, 0, time.UTC)
		occurredDay = &day
	}
	recordedDay := time.Date(record.RecordedAt.UTC().Year(), record.RecordedAt.UTC().Month(), record.RecordedAt.UTC().Day(), 0, 0, 0, 0, time.UTC)

	return projectedRecord{
		RecordID:              record.RecordID,
		IncidentID:            record.IncidentID,
		RowVersion:            record.RowVersion,
		OccurredAt:            normalizeTimePointer(record.OccurredAt),
		Summary:               cloneStringPointer(record.Summary),
		Details:               cloneStringPointer(record.Details),
		SourceText:            cloneStringPointer(record.SourceText),
		RecordedAt:            record.RecordedAt.UTC(),
		EditedAt:              record.EditedAt.UTC(),
		SortTs:                sortTS,
		CaptureState:          record.CaptureState,
		ReplacementRecordID:   replacementRecordID,
		OccurredDay:           occurredDay,
		RecordedDay:           recordedDay,
		EvidenceCount:         0,
		HasEvidence:           false,
		HasUnresolvedMentions: false,
	}
}

func projectionInput(record projectedRecord) projections.TimelineProjectionInput {
	return projections.TimelineProjectionInput{
		RecordID:              record.RecordID,
		IncidentID:            record.IncidentID,
		RowVersion:            record.RowVersion,
		OccurredAt:            record.OccurredAt,
		Summary:               record.Summary,
		Details:               record.Details,
		SourceText:            record.SourceText,
		RecordedAt:            record.RecordedAt,
		EditedAt:              record.EditedAt,
		CaptureState:          record.CaptureState,
		ReplacementRecordID:   record.ReplacementRecordID,
		EvidenceCount:         record.EvidenceCount,
		HasEvidence:           record.HasEvidence,
		HasUnresolvedMentions: record.HasUnresolvedMentions,
	}
}

func projectedRecordFromSQL(row sqlc.GetTimelineProjectionRowRow) (projectedRecord, error) {
	recordID, err := uuidFromPG(row.RecordID)
	if err != nil {
		return projectedRecord{}, err
	}
	incidentID, err := uuidFromPG(row.IncidentID)
	if err != nil {
		return projectedRecord{}, err
	}
	recordedAt, err := timeFromPG(row.RecordedAt)
	if err != nil {
		return projectedRecord{}, err
	}
	editedAt, err := timeFromPG(row.EditedAt)
	if err != nil {
		return projectedRecord{}, err
	}
	sortTS, err := timeFromPG(row.SortTs)
	if err != nil {
		return projectedRecord{}, err
	}
	recordedDay, err := dateFromPG(row.RecordedDay)
	if err != nil {
		return projectedRecord{}, err
	}

	return projectedRecord{
		RecordID:              recordID,
		IncidentID:            incidentID,
		RowVersion:            row.RowVersion,
		OccurredAt:            optionalTimeFromPG(row.OccurredAt),
		Summary:               optionalTextFromPG(row.Summary),
		Details:               optionalTextFromPG(row.Details),
		SourceText:            optionalTextFromPG(row.SourceText),
		RecordedAt:            recordedAt,
		EditedAt:              editedAt,
		SortTs:                sortTS,
		CaptureState:          row.CaptureState,
		ReplacementRecordID:   optionalUUIDFromPG(row.ReplacementRecordID),
		OccurredDay:           optionalDateFromPG(row.OccurredDay),
		RecordedDay:           recordedDay,
		EvidenceCount:         int(row.EvidenceCount),
		HasEvidence:           row.HasEvidence,
		HasUnresolvedMentions: row.HasUnresolvedMentions,
	}, nil
}

func scanProjectedRecord(scanner interface {
	Scan(dest ...any) error
}) (projectedRecord, error) {
	var row sqlc.GetTimelineProjectionRowRow
	if err := scanner.Scan(
		&row.RecordID,
		&row.IncidentID,
		&row.RowVersion,
		&row.OccurredAt,
		&row.Summary,
		&row.Details,
		&row.SourceText,
		&row.RecordedAt,
		&row.EditedAt,
		&row.SortTs,
		&row.CaptureState,
		&row.ReplacementRecordID,
		&row.OccurredDay,
		&row.RecordedDay,
		&row.EvidenceCount,
		&row.HasEvidence,
		&row.HasUnresolvedMentions,
	); err != nil {
		return projectedRecord{}, fmt.Errorf("scan timeline projection row: %w", err)
	}
	return projectedRecordFromSQL(row)
}

var timelineSortExpressions = map[string]string{
	"record_id":                        "t.record_id",
	"timeline.sort_ts":                 "t.sort_ts",
	"timeline.summary":                 "t.summary",
	"timeline.evidence_count":          "t.evidence_count",
	"timeline.edited_at":               "t.edited_at",
	"timeline.capture_state":           "t.capture_state",
	"timeline.occurred_day":            "t.occurred_day",
	"timeline.recorded_day":            "t.recorded_day",
	"timeline.has_evidence":            "t.has_evidence",
	"timeline.has_unresolved_mentions": "t.has_unresolved_mentions",
}

func buildTimelineQuerySQL(incidentID uuid.UUID, query viewschema.QueryMeta) (string, []any, error) {
	var builder strings.Builder
	builder.WriteString(`
SELECT
    t.record_id,
    t.incident_id,
    r.row_version,
    t.occurred_at,
    t.summary,
    t.details,
    t.source_text,
    t.recorded_at,
    t.edited_at,
    t.sort_ts,
    t.capture_state,
    t.replacement_record_id,
    t.occurred_day,
    t.recorded_day,
    t.evidence_count,
    t.has_evidence,
    t.has_unresolved_mentions
  FROM timeline_grid_projection t
  JOIN records r
    ON r.record_id = t.record_id
 WHERE t.incident_id = $1`)

	args := []any{incidentID}
	for _, filter := range query.Filters {
		if err := appendTimelineFilter(&builder, &args, filter); err != nil {
			return "", nil, err
		}
	}

	builder.WriteString(" ORDER BY ")
	for index, sortEntry := range query.Sort {
		if index > 0 {
			builder.WriteString(", ")
		}
		expr, ok := timelineSortExpressions[sortEntry.FieldKey]
		if !ok {
			return "", nil, fmt.Errorf("timeline query sort field %q not mapped", sortEntry.FieldKey)
		}
		builder.WriteString(expr)
		if sortEntry.Direction == "desc" {
			builder.WriteString(" DESC")
		} else {
			builder.WriteString(" ASC")
		}
	}

	return builder.String(), args, nil
}

func appendTimelineFilter(builder *strings.Builder, args *[]any, filter viewschema.Filter) error {
	switch filter.FieldKey {
	case "timeline.occurred_day":
		return appendDateFilterClause(builder, args, "t.occurred_day", filter)
	case "timeline.recorded_day":
		return appendDateFilterClause(builder, args, "t.recorded_day", filter)
	case "timeline.capture_state":
		return appendStringFilterClause(builder, args, "t.capture_state", filter)
	case "timeline.has_evidence":
		return appendBoolFilterClause(builder, args, "t.has_evidence", filter)
	case "timeline.has_unresolved_mentions":
		return appendBoolFilterClause(builder, args, "t.has_unresolved_mentions", filter)
	case "timeline.tags":
		return appendTimelineTagFilterClause(builder, args, filter)
	default:
		return fmt.Errorf("timeline query filter field %q not mapped", filter.FieldKey)
	}
}

func appendTimelineTagFilterClause(builder *strings.Builder, args *[]any, filter viewschema.Filter) error {
	values := stringValues(filter.Arg["values"])
	switch filter.Op {
	case "contains_any":
		builder.WriteString(`
   AND EXISTS (
        SELECT 1
          FROM record_tags rt
         WHERE rt.incident_id = t.incident_id
           AND rt.record_id = t.record_id
           AND rt.deleted_at IS NULL
           AND rt.normalized_tag_name = ANY(`)
		builder.WriteString(bindWithCast(args, values, "text[]"))
		builder.WriteString(`)
   )`)
		return nil
	case "contains_all":
		for _, value := range values {
			builder.WriteString(`
   AND EXISTS (
        SELECT 1
          FROM record_tags rt
         WHERE rt.incident_id = t.incident_id
           AND rt.record_id = t.record_id
           AND rt.deleted_at IS NULL
           AND rt.normalized_tag_name = `)
			builder.WriteString(bind(args, value))
			builder.WriteString(`
   )`)
		}
		return nil
	default:
		return fmt.Errorf("timeline tag operator %q not mapped", filter.Op)
	}
}

func appendStringFilterClause(builder *strings.Builder, args *[]any, expr string, filter viewschema.Filter) error {
	switch filter.Op {
	case "eq":
		return appendEqualityClause(builder, args, expr, filter.Arg)
	case "prefix":
		value, _ := filter.Arg["value"].(string)
		builder.WriteString("\n   AND lower(coalesce((")
		builder.WriteString(expr)
		builder.WriteString(")::text, '')) LIKE ")
		builder.WriteString(bindWithCast(args, value+"%", ""))
		return nil
	default:
		return fmt.Errorf("string filter operator %q not mapped", filter.Op)
	}
}

func appendBoolFilterClause(builder *strings.Builder, args *[]any, expr string, filter viewschema.Filter) error {
	switch filter.Op {
	case "eq":
		return appendEqualityClause(builder, args, expr, filter.Arg)
	default:
		return fmt.Errorf("bool filter operator %q not mapped", filter.Op)
	}
}

func appendDateFilterClause(builder *strings.Builder, args *[]any, expr string, filter viewschema.Filter) error {
	switch filter.Op {
	case "eq":
		return appendEqualityClauseWithCast(builder, args, expr, filter.Arg, "date")
	case "range":
		return appendRangeClause(builder, args, expr, filter.Arg, "date")
	default:
		return fmt.Errorf("date filter operator %q not mapped", filter.Op)
	}
}

func appendEqualityClause(builder *strings.Builder, args *[]any, expr string, arg map[string]any) error {
	return appendEqualityClauseWithCast(builder, args, expr, arg, "")
}

func appendEqualityClauseWithCast(builder *strings.Builder, args *[]any, expr string, arg map[string]any, cast string) error {
	if value, ok := arg["value"]; ok {
		if value == nil {
			builder.WriteString("\n   AND ")
			builder.WriteString(expr)
			builder.WriteString(" IS NULL")
			return nil
		}
		builder.WriteString("\n   AND ")
		builder.WriteString(expr)
		builder.WriteString(" = ")
		builder.WriteString(bindWithCast(args, value, cast))
		return nil
	}

	values, ok := arg["values"].([]any)
	if !ok || len(values) == 0 {
		return fmt.Errorf("missing equality values for %s", expr)
	}
	builder.WriteString("\n   AND (")
	for index, value := range values {
		if index > 0 {
			builder.WriteString(" OR ")
		}
		builder.WriteString(expr)
		builder.WriteString(" = ")
		builder.WriteString(bindWithCast(args, value, cast))
	}
	builder.WriteString(")")
	return nil
}

func appendRangeClause(builder *strings.Builder, args *[]any, expr string, arg map[string]any, cast string) error {
	for _, bound := range []struct {
		Key string
		Op  string
	}{
		{Key: "gt", Op: ">"},
		{Key: "gte", Op: ">="},
		{Key: "lt", Op: "<"},
		{Key: "lte", Op: "<="},
	} {
		value, ok := arg[bound.Key]
		if !ok {
			continue
		}
		builder.WriteString("\n   AND ")
		builder.WriteString(expr)
		builder.WriteByte(' ')
		builder.WriteString(bound.Op)
		builder.WriteByte(' ')
		builder.WriteString(bindWithCast(args, value, cast))
	}
	return nil
}

func bind(args *[]any, value any) string {
	return bindWithCast(args, value, "")
}

func bindWithCast(args *[]any, value any, cast string) string {
	*args = append(*args, value)
	placeholder := fmt.Sprintf("$%d", len(*args))
	if cast == "" {
		return placeholder
	}
	return placeholder + "::" + cast
}

func stringValues(raw any) []string {
	items, ok := raw.([]any)
	if !ok {
		return nil
	}
	values := make([]string, 0, len(items))
	for _, item := range items {
		value, ok := item.(string)
		if !ok {
			continue
		}
		values = append(values, value)
	}
	return values
}

type mentionQueryer interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
}

type projectedMentionItem struct {
	MentionID        uuid.UUID
	EntityType       string
	SourceFieldKey   string
	RawText          string
	ResolutionStatus string
	ResolvedRecordID *uuid.UUID
	ResolutionMethod *string
}

func hydrateProjectedCollections(ctx context.Context, querier mentionQueryer, record *projectedRecord) error {
	if record == nil {
		return nil
	}
	rows, err := querier.Query(ctx, `
SELECT entity_mention_id, entity_type, source_field_key, raw_text, resolution_status, resolved_record_id, resolution_method, ordinal
  FROM entity_mentions
 WHERE source_record_id = $1
   AND resolution_status IN ('unresolved', 'resolved')
 ORDER BY ordinal ASC, entity_mention_id ASC
`, record.RecordID)
	if err != nil {
		return fmt.Errorf("query timeline mention collections: %w", err)
	}

	mentions := make([]projectedMentionItem, 0)
	for rows.Next() {
		var (
			mentionID        uuid.UUID
			entityType       string
			sourceFieldKey   string
			rawText          string
			resolutionStatus string
			resolvedRecordID pgtype.UUID
			resolutionMethod pgtype.Text
			ordinal          int
		)
		if err := rows.Scan(&mentionID, &entityType, &sourceFieldKey, &rawText, &resolutionStatus, &resolvedRecordID, &resolutionMethod, &ordinal); err != nil {
			rows.Close()
			return fmt.Errorf("scan timeline mention collection row: %w", err)
		}
		mention := projectedMentionItem{
			MentionID:        mentionID,
			EntityType:       entityType,
			SourceFieldKey:   sourceFieldKey,
			RawText:          rawText,
			ResolutionStatus: resolutionStatus,
		}
		if resolvedRecordID.Valid {
			resolved := uuid.Must(uuid.FromBytes(resolvedRecordID.Bytes[:]))
			mention.ResolvedRecordID = &resolved
		}
		if resolutionMethod.Valid {
			value := resolutionMethod.String
			mention.ResolutionMethod = &value
		}
		mentions = append(mentions, mention)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return fmt.Errorf("iterate timeline mention collection rows: %w", err)
	}
	rows.Close()

	hostRefs := make([]map[string]any, 0)
	identityRefs := make([]map[string]any, 0)
	hasUnresolved := false
	for _, mention := range mentions {
		item := map[string]any{
			"item_ref":     "entity_mention:" + mention.MentionID.String(),
			"entity_type":  mention.EntityType,
			"display_text": mention.RawText,
			"raw_text":     mention.RawText,
		}
		if mention.ResolutionStatus == "resolved" && mention.ResolvedRecordID != nil {
			item["item_kind"] = "resolved_ref"
			item["resolved_record_id"] = mention.ResolvedRecordID.String()
			if mention.ResolutionMethod != nil && *mention.ResolutionMethod != "" {
				item["resolution_method"] = *mention.ResolutionMethod
				if *mention.ResolutionMethod == autoResolutionMethod {
					item["auto_resolved"] = true
				}
			}
			if linkType, ok := timelineRelationshipLinkType(mention.SourceFieldKey); ok {
				linkMetadata, err := loadActiveCollectionLinkMetadata(ctx, querier, record.IncidentID, record.RecordID, *mention.ResolvedRecordID, linkType)
				if err != nil {
					return err
				}
				if linkMetadata != nil {
					item["provenance"] = linkMetadata.Provenance
					item["confidence"] = linkMetadata.Confidence
				}
			}
			if mention.ResolutionMethod != nil && *mention.ResolutionMethod == autoResolutionMethod {
				matchedAliasText, err := lookupMatchedAliasText(ctx, querier, *mention.ResolvedRecordID, mention.EntityType, mention.RawText)
				if err != nil {
					return err
				}
				if matchedAliasText != nil {
					item["matched_alias_text"] = *matchedAliasText
				}
			}
		} else {
			item["item_kind"] = "unresolved_mention"
			hasUnresolved = true
		}

		switch mention.EntityType {
		case "host":
			hostRefs = append(hostRefs, item)
		case "identity":
			identityRefs = append(identityRefs, item)
		}
	}
	record.HostRefs = hostRefs
	record.IdentityRefs = identityRefs
	record.HasUnresolvedMentions = hasUnresolved
	tagRows, err := querier.Query(ctx, `
SELECT record_tag_id, tag_name
  FROM record_tags
 WHERE incident_id = $1
   AND record_id = $2
   AND deleted_at IS NULL
 ORDER BY normalized_tag_name ASC, record_tag_id ASC
`, record.IncidentID, record.RecordID)
	if err != nil {
		return fmt.Errorf("query timeline tags: %w", err)
	}

	tags := make([]map[string]any, 0)
	for tagRows.Next() {
		var (
			recordTagID uuid.UUID
			tagName     string
		)
		if err := tagRows.Scan(&recordTagID, &tagName); err != nil {
			tagRows.Close()
			return fmt.Errorf("scan timeline tag row: %w", err)
		}
		tags = append(tags, map[string]any{
			"item_ref":     "record_tag:" + recordTagID.String(),
			"item_kind":    "tag",
			"display_text": tagName,
			"raw_text":     tagName,
		})
	}
	if err := tagRows.Err(); err != nil {
		tagRows.Close()
		return fmt.Errorf("iterate timeline tags: %w", err)
	}
	tagRows.Close()
	record.Tags = tags

	evidenceRows, err := querier.Query(ctx, `
SELECT
    rl.dst_record_id,
    COALESCE(ev.title, rl.dst_record_id::text) AS title,
    ev.lifecycle_state,
    COALESCE(b.upload_state, ev.upload_state, 'pending') AS upload_state
  FROM record_links rl
  JOIN evidence ev
    ON ev.incident_id = rl.incident_id
   AND ev.record_id = rl.dst_record_id
  LEFT JOIN object_blobs b
    ON b.object_blob_id = ev.object_blob_id
 WHERE rl.incident_id = $1
   AND rl.src_record_id = $2
   AND rl.link_type = 'attached_evidence'
   AND rl.deleted_at IS NULL
 ORDER BY COALESCE(ev.title, rl.dst_record_id::text) ASC, rl.dst_record_id ASC
`, record.IncidentID, record.RecordID)
	if err != nil {
		return fmt.Errorf("query timeline attached evidence: %w", err)
	}
	attachedEvidence := make([]map[string]any, 0)
	availableEvidenceCount := 0
	for evidenceRows.Next() {
		var (
			evidenceRecordID uuid.UUID
			title            string
			lifecycleState   string
			uploadState      string
		)
		if err := evidenceRows.Scan(&evidenceRecordID, &title, &lifecycleState, &uploadState); err != nil {
			evidenceRows.Close()
			return fmt.Errorf("scan timeline attached evidence row: %w", err)
		}
		attachedEvidence = append(attachedEvidence, map[string]any{
			"item_ref":         "record_ref:" + evidenceRecordID.String(),
			"item_kind":        "record_ref",
			"display_text":     title,
			"linked_record_id": evidenceRecordID.String(),
		})
		if uploadState == "available" && (lifecycleState == "available" || lifecycleState == "released") {
			availableEvidenceCount += 1
		}
	}
	if err := evidenceRows.Err(); err != nil {
		evidenceRows.Close()
		return fmt.Errorf("iterate timeline attached evidence rows: %w", err)
	}
	evidenceRows.Close()
	record.AttachedEvidence = attachedEvidence
	record.EvidenceCount = availableEvidenceCount
	record.HasEvidence = availableEvidenceCount > 0
	return nil
}

type mentionInsertOptions struct {
	allowInteractiveAutoResolution bool
}

func applyCreateMentionActionsTx(ctx context.Context, tx pgx.Tx, actorUserID uuid.UUID, incidentID uuid.UUID, recordID uuid.UUID, hostRefs *CollectionActionPayload, identityRefs *CollectionActionPayload, now time.Time) error {
	if err := insertMentionActionsTx(ctx, tx, actorUserID, incidentID, recordID, "timeline.host_refs", "host", hostRefs, mentionInsertOptions{}, now); err != nil {
		return err
	}
	if err := insertMentionActionsTx(ctx, tx, actorUserID, incidentID, recordID, "timeline.identity_refs", "identity", identityRefs, mentionInsertOptions{}, now); err != nil {
		return err
	}
	return nil
}

func applyCreateTagActionsTx(ctx context.Context, tx pgx.Tx, actorUserID uuid.UUID, incidentID uuid.UUID, recordID uuid.UUID, tags *CollectionActionPayload, now time.Time) error {
	return insertTagActionsTx(ctx, tx, actorUserID, incidentID, recordID, tags, now)
}

func applyPatchMentionActionsTx(ctx context.Context, tx pgx.Tx, actor authn.UserRecord, incidentID uuid.UUID, recordID uuid.UUID, changes []PatchChange, now time.Time) error {
	entityStore := entities.NewStore(nil)
	for _, change := range changes {
		if change.ActionPayload == nil {
			continue
		}
		entityType := "host"
		if change.FieldKey == "timeline.identity_refs" {
			entityType = "identity"
		}
		for _, action := range change.ActionPayload.Actions {
			switch action.Op {
			case "add_token", "add_resolved_ref":
				if err := insertMentionActionsTx(ctx, tx, actor.ID, incidentID, recordID, change.FieldKey, entityType, &CollectionActionPayload{Actions: []CollectionAction{action}}, mentionInsertOptions{
					allowInteractiveAutoResolution: true,
				}, now); err != nil {
					return err
				}
			case "resolve_item":
				mentionID, err := mentionIDFromItemRef(action.ItemRef)
				if err != nil {
					return err
				}
				if _, err := entityStore.ResolveOrCreateFromMentionTx(ctx, tx, actor, recordID, change.FieldKey, mentionID, action.ResolvedRecord, now); err != nil {
					return err
				}
			case "dismiss_item", "revert_to_unresolved":
				mentionID, err := mentionIDFromItemRef(action.ItemRef)
				if err != nil {
					return err
				}
				if err := entityStore.ApplyMentionLifecycleTx(ctx, tx, actor, recordID, change.FieldKey, mentionID, action.Op, nil, now); err != nil {
					return err
				}
			default:
				return fmt.Errorf("unsupported mention action: %s", action.Op)
			}
		}
	}
	return nil
}

func applyPatchTagActionsTx(ctx context.Context, tx pgx.Tx, actorUserID uuid.UUID, incidentID uuid.UUID, recordID uuid.UUID, changes []PatchChange, now time.Time) error {
	for _, change := range changes {
		if change.FieldKey != "timeline.tags" || change.ActionPayload == nil {
			continue
		}
		if err := insertTagActionsTx(ctx, tx, actorUserID, incidentID, recordID, change.ActionPayload, now); err != nil {
			return err
		}
	}
	return nil
}

func applyPatchAttachedEvidenceActionsTx(ctx context.Context, tx pgx.Tx, actorUserID uuid.UUID, incidentID uuid.UUID, recordID uuid.UUID, changes []PatchChange, now time.Time) error {
	for _, change := range changes {
		if change.FieldKey != "timeline.attached_evidence_ids" || change.ActionPayload == nil {
			continue
		}
		if err := applyAttachedEvidenceActionsTx(ctx, tx, actorUserID, incidentID, recordID, change.ActionPayload, now); err != nil {
			return err
		}
	}
	return nil
}

func applyAttachedEvidenceActionsTx(ctx context.Context, tx pgx.Tx, actorUserID uuid.UUID, incidentID uuid.UUID, recordID uuid.UUID, payload *CollectionActionPayload, now time.Time) error {
	if payload == nil || len(payload.Actions) == 0 {
		return nil
	}
	for _, action := range payload.Actions {
		switch action.Op {
		case "add_record_ref":
			if action.LinkedRecordID == nil {
				return fmt.Errorf("missing linked evidence record")
			}
			if err := validateAttachedEvidenceTargetTx(ctx, tx, incidentID, *action.LinkedRecordID); err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, `
INSERT INTO record_links (
    incident_id,
    src_record_id,
    dst_record_id,
    link_type,
    field_key,
    provenance,
    confidence,
    owner_user_id,
    created_by_user_id,
    decided_at,
    created_at
)
VALUES ($1, $2, $3, 'attached_evidence', 'timeline.attached_evidence_ids', 'manual', NULL, $4, $4, $5, $5)
ON CONFLICT DO NOTHING
`, incidentID, recordID, *action.LinkedRecordID, actorUserID, now.UTC()); err != nil {
				return fmt.Errorf("insert attached evidence link: %w", err)
			}
		case "remove_record_ref":
			evidenceRecordID, err := recordIDFromRecordRefItem(action.ItemRef)
			if err != nil {
				return err
			}
			if _, err := tx.Exec(ctx, `
UPDATE record_links
   SET deleted_at = COALESCE(deleted_at, $5),
       deleted_by_user_id = COALESCE(deleted_by_user_id, $4)
 WHERE incident_id = $1
   AND src_record_id = $2
   AND dst_record_id = $3
   AND link_type = 'attached_evidence'
   AND field_key = 'timeline.attached_evidence_ids'
   AND deleted_at IS NULL
`, incidentID, recordID, evidenceRecordID, actorUserID, now.UTC()); err != nil {
				return fmt.Errorf("remove attached evidence link: %w", err)
			}
		default:
			return fmt.Errorf("unsupported attached evidence action: %s", action.Op)
		}
	}
	return nil
}

func insertTagActionsTx(ctx context.Context, tx pgx.Tx, actorUserID uuid.UUID, incidentID uuid.UUID, recordID uuid.UUID, payload *CollectionActionPayload, now time.Time) error {
	if payload == nil || len(payload.Actions) == 0 {
		return nil
	}
	for _, action := range payload.Actions {
		if action.Op != "add_token" {
			return fmt.Errorf("unsupported tag action: %s", action.Op)
		}

		var exists bool
		if err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1
      FROM record_tags
     WHERE incident_id = $1
       AND record_id = $2
       AND normalized_tag_name = $3
       AND deleted_at IS NULL
)
`, incidentID, recordID, action.NormalizedText).Scan(&exists); err != nil {
			return fmt.Errorf("query active record tag: %w", err)
		}
		if exists {
			continue
		}

		if _, err := tx.Exec(ctx, `
INSERT INTO record_tags (
    incident_id,
    record_id,
    tag_name,
    normalized_tag_name,
    created_by_user_id,
    created_at,
    updated_at
)
VALUES ($1, $2, $3, $4, $5, $6, $6)
`, incidentID, recordID, action.NormalizedText, action.NormalizedText, actorUserID, now.UTC()); err != nil {
			return fmt.Errorf("insert record tag: %w", err)
		}
	}
	return nil
}

func insertMentionActionsTx(ctx context.Context, tx pgx.Tx, actorUserID uuid.UUID, incidentID uuid.UUID, recordID uuid.UUID, fieldKey string, entityType string, payload *CollectionActionPayload, options mentionInsertOptions, now time.Time) error {
	if payload == nil || len(payload.Actions) == 0 {
		return nil
	}
	nextOrdinal, err := nextMentionOrdinalTx(ctx, tx, recordID, fieldKey)
	if err != nil {
		return err
	}
	linkStore := links.NewStore()
	for _, action := range payload.Actions {
		if action.Op != "add_token" && action.Op != "add_resolved_ref" {
			return fmt.Errorf("unsupported mention action: %s", action.Op)
		}
		resolutionStatus := "unresolved"
		var resolvedRecordID *uuid.UUID
		var resolvedByUserID any
		var resolvedAt any
		var resolutionMethod any
		linkProvenance := "manual"
		var linkConfidence *int
		if action.Op == "add_resolved_ref" {
			if action.ResolvedRecord == nil {
				return fmt.Errorf("missing resolved record for action: %s", action.Op)
			}
			if err := validateTimelineResolvedTargetTx(ctx, tx, incidentID, entityType, *action.ResolvedRecord); err != nil {
				return err
			}
			resolutionStatus = "resolved"
			resolvedRecordID = action.ResolvedRecord
			resolvedByUserID = actorUserID
			resolvedAt = now.UTC()
			resolutionMethod = action.Op
		} else if options.allowInteractiveAutoResolution {
			match, err := lookupInteractiveAutoResolutionMatchTx(ctx, tx, incidentID, fieldKey, action.RawText)
			if err != nil {
				return err
			}
			if match != nil {
				resolutionStatus = "resolved"
				resolvedRecordID = &match.RecordID
				resolvedByUserID = actorUserID
				resolvedAt = now.UTC()
				resolutionMethod = autoResolutionMethod
				linkProvenance = autoResolutionMethod
				confidence := 100
				linkConfidence = &confidence
			}
		}
		if _, err := tx.Exec(ctx, `
INSERT INTO entity_mentions (
    source_record_id,
    entity_type,
    source_field_key,
    origin_kind,
    origin_locator,
    raw_text,
    normalized_text,
    resolution_status,
    row_version,
    ordinal,
    created_by_user_id,
    created_at,
    resolved_record_id,
    resolved_by_user_id,
    resolved_at,
    resolution_method
)
VALUES ($1, $2, $3, 'interactive_cell', $4, $5, $6, $7, 1, $8, $9, $10, $11, $12, $13, $14)
`, recordID, entityType, fieldKey, mentionOriginLocator(recordID, fieldKey, nextOrdinal), action.RawText, action.NormalizedText, resolutionStatus, nextOrdinal, actorUserID, now.UTC(), resolvedRecordID, resolvedByUserID, resolvedAt, resolutionMethod); err != nil {
			return fmt.Errorf("insert entity mention: %w", err)
		}
		if resolvedRecordID != nil {
			linkType, ok := timelineRelationshipLinkType(fieldKey)
			if !ok {
				return fmt.Errorf("unsupported link field: %s", fieldKey)
			}
			if _, _, err := linkStore.UpsertLinkTx(ctx, tx, incidentID, recordID, *resolvedRecordID, linkType, linkProvenance, linkConfidence, actorUserID, now.UTC()); err != nil {
				return fmt.Errorf("upsert record link: %w", err)
			}
		}
		nextOrdinal++
	}
	return nil
}

func timelineRelationshipLinkType(fieldKey string) (string, bool) {
	switch fieldKey {
	case "timeline.host_refs":
		return "observed_on_host", true
	case "timeline.identity_refs":
		return "observed_as_identity", true
	default:
		return "", false
	}
}

func nextMentionOrdinalTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, fieldKey string) (int, error) {
	var nextOrdinal int
	if err := tx.QueryRow(ctx, `
SELECT COALESCE(MAX(ordinal), 0) + 1
  FROM entity_mentions
 WHERE source_record_id = $1
   AND source_field_key = $2
`, recordID, fieldKey).Scan(&nextOrdinal); err != nil {
		return 0, fmt.Errorf("query next mention ordinal: %w", err)
	}
	return nextOrdinal, nil
}

func validateTimelineResolvedTargetTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, entityType string, resolvedRecordID uuid.UUID) error {
	var exists bool
	switch entityType {
	case "host":
		if err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1
      FROM hosts
     WHERE record_id = $1
       AND incident_id = $2
       AND host_state IN ('stub', 'canonical')
)
`, resolvedRecordID, incidentID).Scan(&exists); err != nil {
			return fmt.Errorf("validate host target: %w", err)
		}
	case "identity":
		if err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1
      FROM identities
     WHERE record_id = $1
       AND incident_id = $2
       AND identity_state IN ('stub', 'canonical')
)
`, resolvedRecordID, incidentID).Scan(&exists); err != nil {
			return fmt.Errorf("validate identity target: %w", err)
		}
	default:
		return entities.ErrResolvedRecordNotFound
	}
	if !exists {
		return entities.ErrResolvedRecordNotFound
	}
	return nil
}

func validateAttachedEvidenceTargetTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, evidenceRecordID uuid.UUID) error {
	var exists bool
	if err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1
      FROM records r
      JOIN evidence ev
        ON ev.incident_id = r.incident_id
       AND ev.record_id = r.record_id
     WHERE r.record_id = $1
       AND r.incident_id = $2
       AND r.record_type = 'evidence'
       AND r.deleted_at IS NULL
)
`, evidenceRecordID, incidentID).Scan(&exists); err != nil {
		return fmt.Errorf("validate attached evidence target: %w", err)
	}
	if !exists {
		return entities.ErrResolvedRecordNotFound
	}
	return nil
}

func recordIDFromRecordRefItem(itemRef string) (uuid.UUID, error) {
	const prefix = "record_ref:"
	if !strings.HasPrefix(itemRef, prefix) {
		return uuid.UUID{}, fmt.Errorf("invalid record ref item_ref: %s", itemRef)
	}
	recordID, err := uuid.Parse(strings.TrimPrefix(itemRef, prefix))
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("parse record ref item_ref: %w", err)
	}
	return recordID, nil
}

func mentionOriginLocator(recordID uuid.UUID, fieldKey string, ordinal int) string {
	return fmt.Sprintf("view:%s/record:%s/cell:%s/item:%d", TimelineViewSchemaID, recordID.String(), fieldKey, ordinal)
}

func mentionIDFromItemRef(itemRef string) (uuid.UUID, error) {
	const prefix = "entity_mention:"
	if !strings.HasPrefix(itemRef, prefix) {
		return uuid.UUID{}, fmt.Errorf("invalid mention item_ref: %s", itemRef)
	}
	mentionID, err := uuid.Parse(strings.TrimPrefix(itemRef, prefix))
	if err != nil {
		return uuid.UUID{}, fmt.Errorf("parse mention item_ref: %w", err)
	}
	return mentionID, nil
}

func (s *Store) beforeCommit(routeKey string, recordID uuid.UUID) error {
	if s == nil || s.hooks.BeforeCommit == nil {
		return nil
	}
	return s.hooks.BeforeCommit(routeKey, recordID)
}

func versionID(recordID uuid.UUID, rowVersion int64) string {
	return fmt.Sprintf("timeline:%s:%d", recordID.String(), rowVersion)
}

func hasMaterialChange(current sourceRecord, next sourceRecord) bool {
	return !timePointersEqual(current.OccurredAt, next.OccurredAt) ||
		!stringPointersEqual(current.Summary, next.Summary) ||
		!stringPointersEqual(current.Details, next.Details) ||
		!stringPointersEqual(current.SourceText, next.SourceText)
}

func isRecordLinkConflict(err error) bool {
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) {
		return false
	}
	return pgErr.Code == "23505"
}

func extractUUIDFromPayload(payload map[string]any, path ...string) (uuid.UUID, error) {
	current := any(payload)
	for _, segment := range path {
		object, ok := current.(map[string]any)
		if !ok {
			return uuid.UUID{}, fmt.Errorf("decode payload path %q", strings.Join(path, "."))
		}
		current = object[segment]
	}
	text, ok := current.(string)
	if !ok || text == "" {
		return uuid.UUID{}, fmt.Errorf("decode payload uuid path %q", strings.Join(path, "."))
	}
	parsed, err := uuid.Parse(text)
	if err != nil {
		return uuid.UUID{}, err
	}
	return parsed, nil
}

func cloneStringPointer(value *string) *string {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func normalizeTimePointer(value *time.Time) *time.Time {
	if value == nil {
		return nil
	}
	utc := value.UTC()
	return &utc
}

func stringPointersEqual(left *string, right *string) bool {
	switch {
	case left == nil && right == nil:
		return true
	case left == nil || right == nil:
		return false
	default:
		return *left == *right
	}
}

func timePointersEqual(left *time.Time, right *time.Time) bool {
	switch {
	case left == nil && right == nil:
		return true
	case left == nil || right == nil:
		return false
	default:
		return left.UTC().Equal(right.UTC())
	}
}

func pgUUID(value uuid.UUID) pgtype.UUID {
	return pgtype.UUID{Bytes: [16]byte(value), Valid: true}
}

func uuidFromPG(value pgtype.UUID) (uuid.UUID, error) {
	if !value.Valid {
		return uuid.UUID{}, errors.New("missing uuid")
	}
	return uuid.FromBytes(value.Bytes[:])
}

func optionalUUIDFromPG(value pgtype.UUID) *uuid.UUID {
	if !value.Valid {
		return nil
	}
	parsed := uuid.Must(uuid.FromBytes(value.Bytes[:]))
	return &parsed
}

func timeFromPG(value pgtype.Timestamptz) (time.Time, error) {
	if !value.Valid {
		return time.Time{}, errors.New("missing timestamp")
	}
	return value.Time.UTC(), nil
}

func optionalTimeFromPG(value pgtype.Timestamptz) *time.Time {
	if !value.Valid {
		return nil
	}
	utc := value.Time.UTC()
	return &utc
}

func optionalTextFromPG(value pgtype.Text) *string {
	if !value.Valid {
		return nil
	}
	text := value.String
	return &text
}

func dateFromPG(value pgtype.Date) (time.Time, error) {
	if !value.Valid {
		return time.Time{}, errors.New("missing date")
	}
	return value.Time.UTC(), nil
}

func optionalDateFromPG(value pgtype.Date) *time.Time {
	if !value.Valid {
		return nil
	}
	utc := value.Time.UTC()
	return &utc
}
