package timeline

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math"
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
	pool             postgres.DB
	idempotencyStore timelineIdempotencyPort
	recordStore      timelineRecordPort
	revisionsStore   timelineRevisionPort
	projectionStore  timelineProjectionPort
	linkStore        timelineLinkPort
	mentionStore     timelineMentionPort
	hooks            StoreHooks
}

type projectedRecord struct {
	RecordID              uuid.UUID
	IncidentID            uuid.UUID
	RowVersion            int64
	DateEnteredText       *string
	AnalystText           *string
	MitreStageText        *string
	DeviceObjectText      *string
	IPAddressText         *string
	ActivityUTCText       *string
	ActivityLocalText     *string
	RawActivityText       *string
	ActivitySynopsisText  *string
	DataSourceText        *string
	RecordedAt            time.Time
	EditedAt              time.Time
	ActivitySortTS        *time.Time
	DateEnteredSortDay    *time.Time
	ActivityTimePairState string
	CaptureState          string
	ReplacementRecordID   *uuid.UUID
	EvidenceCount         int
	HasEvidence           bool
	HasUnresolvedMentions bool
	HostRefs              []map[string]any
	IdentityRefs          []map[string]any
	AttachedEvidence      []map[string]any
	Tags                  []map[string]any
}

type sourceRecord struct {
	RecordID               uuid.UUID
	IncidentID             uuid.UUID
	DateEnteredText        *string
	AnalystText            *string
	MitreStageText         *string
	DeviceObjectText       *string
	IPAddressText          *string
	ActivityUTCText        *string
	ActivityLocalText      *string
	RawActivityText        *string
	ActivitySynopsisText   *string
	DataSourceText         *string
	ActivityUTCGenerated   bool
	ActivityLocalGenerated bool
	ActivityTimePairState  string
	RawCapture             map[string]any
	CaptureState           string
	RowVersion             int64
	RecordedAt             time.Time
	EditedAt               time.Time
	CreatedByUserID        uuid.UUID
	UpdatedByUserID        uuid.UUID
	ReviewedByUserID       *uuid.UUID
	ReviewedAt             *time.Time
	SupersededByUserID     *uuid.UUID
	SupersededAt           *time.Time
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

type attachedEvidenceMutation struct {
	RecordLinkID uuid.UUID
	Operation    string
	BeforeValue  map[string]any
	AfterValue   map[string]any
}

type recordTagMutation struct {
	RecordTagID uuid.UUID
	RecordID    uuid.UUID
	Operation   string
	BeforeValue map[string]any
	AfterValue  map[string]any
}

type RecordSubstrateSnapshot struct {
	RecordID            uuid.UUID
	RowVersion          int64
	CaptureState        string
	ReplacementRecordID *uuid.UUID
	RecordRevisionCount int
}

type TimeConversionProfile struct {
	IncidentID         uuid.UUID
	Enabled            bool
	LocalOffsetMinutes *int
	LocalLabel         *string
	ProfileVersion     int64
	UpdatedAt          time.Time
	UpdatedByUserID    *uuid.UUID
}

type createRowOptions struct {
	allowInteractiveAutoResolution bool
}

func NewStore(pool postgres.DB) *Store {
	return newStoreWithHooks(pool, currentStoreHooks())
}

func newStoreWithHooks(pool postgres.DB, hooks StoreHooks) *Store {
	return newStoreWithPorts(pool, newTimelineStorePorts(pool), hooks)
}

func newStoreWithPorts(pool postgres.DB, ports timelineStorePorts, hooks StoreHooks) *Store {
	return &Store{
		pool:             pool,
		idempotencyStore: ports.idempotency,
		recordStore:      ports.records,
		revisionsStore:   ports.revisions,
		projectionStore:  ports.projections,
		linkStore:        ports.links,
		mentionStore:     ports.mentions,
		hooks:            hooks,
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
	return s.createRow(ctx, actor, incidentID, request, requestHash, requestID, now, createRowOptions{
		allowInteractiveAutoResolution: true,
	})
}

func (s *Store) CreateImportedRow(ctx context.Context, actor authn.UserRecord, incidentID uuid.UUID, request CreateRequest, requestHash []byte, requestID string, now time.Time) (MutationResult, error) {
	return s.createRow(ctx, actor, incidentID, request, requestHash, requestID, now, createRowOptions{})
}

func (s *Store) createRow(ctx context.Context, actor authn.UserRecord, incidentID uuid.UUID, request CreateRequest, requestHash []byte, requestID string, now time.Time, options createRowOptions) (MutationResult, error) {
	scopeKey := incidentID.String() + ":" + TimelineViewSchemaID
	idempotencyKey := authn.RouteIdempotencyKey{
		RouteKey:    createRouteKey,
		ActorUserID: actor.ID,
		ScopeKey:    scopeKey,
		ClientTxnID: request.ClientTxnID,
	}
	if existing, err := s.idempotencyStore.GetRouteIdempotency(ctx, idempotencyKey); err == nil {
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
		RecordID:              recordID,
		IncidentID:            incidentID,
		DateEnteredText:       request.DateEnteredText,
		AnalystText:           request.AnalystText,
		MitreStageText:        request.MitreStageText,
		DeviceObjectText:      request.DeviceObjectText,
		IPAddressText:         request.IPAddressText,
		ActivityUTCText:       request.ActivityUTCText,
		ActivityLocalText:     request.ActivityLocalText,
		RawActivityText:       request.RawActivityText,
		ActivitySynopsisText:  request.ActivitySynopsisText,
		DataSourceText:        request.DataSourceText,
		ActivityTimePairState: "disabled",
		CaptureState:          InitialCaptureState(),
		RowVersion:            1,
		RecordedAt:            now.UTC(),
		EditedAt:              now.UTC(),
		CreatedByUserID:       actor.ID,
		UpdatedByUserID:       actor.ID,
	}
	profile, err := getTimeConversionProfileTx(ctx, tx, incidentID, now.UTC())
	if err != nil {
		return MutationResult{}, err
	}
	applyTimelineTimeConversion(&current, profile)
	var recordEnvelopeInsertMs float64
	var timelineEventInsertMs float64
	if err := tx.QueryRow(ctx, `
WITH timing_start AS (
    SELECT clock_timestamp() AS started_at
),
inserted_record AS (
    INSERT INTO records (
        record_id,
        incident_id,
        record_type,
        created_by_user_id,
        created_at,
        updated_by_user_id,
        updated_at,
        row_version
    )
    VALUES ($1, $2, 'timeline_event', $3, $4, $3, $4, 1)
    RETURNING
        record_id,
        (SELECT started_at FROM timing_start) AS started_at,
        clock_timestamp() AS inserted_at
),
inserted_timeline_event AS (
    INSERT INTO timeline_events (
        record_id,
        incident_id,
        date_entered_text,
        analyst_text,
        mitre_stage_text,
        device_object_text,
        ip_address_text,
        activity_utc_text,
        activity_local_text,
        raw_activity_text,
        activity_synopsis_text,
        data_source_text,
        activity_utc_generated,
        activity_local_generated,
        activity_time_pair_state,
        capture_state,
        row_version,
        recorded_at,
        edited_at,
        created_by_user_id,
        updated_by_user_id
    )
    SELECT
        inserted_record.record_id,
        $2,
        $5,
        $6,
        $7,
        $8,
        $9,
        $10,
        $11,
        $12,
        $13,
        $14,
        $15,
        $16,
        $17,
        'rough',
        1,
        $4,
        $4,
        $3,
        $3
    FROM inserted_record
    RETURNING clock_timestamp() AS inserted_at
)
SELECT
    EXTRACT(EPOCH FROM (inserted_record.inserted_at - inserted_record.started_at)) * 1000,
    EXTRACT(EPOCH FROM (inserted_timeline_event.inserted_at - inserted_record.inserted_at)) * 1000
FROM inserted_record, inserted_timeline_event
`, current.RecordID, incidentID, actor.ID, now.UTC(), current.DateEnteredText, current.AnalystText, current.MitreStageText, current.DeviceObjectText, current.IPAddressText, current.ActivityUTCText, current.ActivityLocalText, current.RawActivityText, current.ActivitySynopsisText, current.DataSourceText, current.ActivityUTCGenerated, current.ActivityLocalGenerated, current.ActivityTimePairState).Scan(&recordEnvelopeInsertMs, &timelineEventInsertMs); err != nil {
		return MutationResult{}, fmt.Errorf("insert timeline base rows: %w", err)
	}
	markCreateTimingDuration(ctx, "store_record_envelope_insert", durationFromMilliseconds(recordEnvelopeInsertMs))
	markCreateTimingDuration(ctx, "store_timeline_event_insert", durationFromMilliseconds(timelineEventInsertMs))

	mentionProjectionRefresh, err := s.applyCreateMentionActionsTx(ctx, tx, actor.ID, current.IncidentID, current.RecordID, request.HostRefs, request.IdentityRefs, options, now.UTC())
	if err != nil {
		return MutationResult{}, err
	}
	if err := s.rebuildMentionEntityProjectionsTx(ctx, tx, current.IncidentID, mentionProjectionRefresh); err != nil {
		return MutationResult{}, err
	}
	tagMutations, err := applyCreateTagActionsTx(ctx, tx, actor.ID, current.IncidentID, current.RecordID, request.Tags, now.UTC())
	if err != nil {
		return MutationResult{}, err
	}
	attachedEvidenceMutations, err := applyAttachedEvidenceActionsTx(ctx, tx, actor.ID, current.IncidentID, current.RecordID, request.AttachedEvidence, now.UTC())
	if err != nil {
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
	if _, err := s.revisionsStore.InsertChangeSetTx(ctx, tx, timelineChangeSetParams{
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

	afterRow := buildRow(projected)
	afterVersion := versionID(current.RecordID, projected.RowVersion)
	if err := s.revisionsStore.InsertMutationTx(ctx, tx, timelineMutationParams{
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
	if err := s.insertAttachedEvidenceMutationEntriesTx(ctx, tx, changeSetID, 2, attachedEvidenceMutations); err != nil {
		return MutationResult{}, err
	}
	if err := s.insertRecordTagMutationEntriesTx(ctx, tx, changeSetID, 2+len(attachedEvidenceMutations), tagMutations); err != nil {
		return MutationResult{}, err
	}
	if err := s.revisionsStore.InsertRecordRevisionTx(ctx, tx, timelineRecordRevisionParams{
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
	if err := s.idempotencyStore.InsertRouteIdempotencyPayload(ctx, tx, idempotencyKey, nil, requestHash, http.StatusCreated, payload); err != nil {
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

func durationFromMilliseconds(milliseconds float64) time.Duration {
	if milliseconds <= 0 || math.IsNaN(milliseconds) || math.IsInf(milliseconds, 0) {
		return 0
	}
	return time.Duration(milliseconds * float64(time.Millisecond))
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
	if existing, err := s.idempotencyStore.GetRouteIdempotency(ctx, idempotencyKey); err == nil {
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
		"row":            buildRow(projected),
	}
	if err := s.idempotencyStore.InsertRouteIdempotencyPayload(ctx, tx, idempotencyKey, nil, requestHash, http.StatusOK, payload); err != nil {
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
	if existing, err := s.idempotencyStore.GetRouteIdempotency(ctx, idempotencyKey); err == nil {
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
		case "timeline.date_entered_text":
			next.DateEnteredText = change.TextValue
		case "timeline.analyst_text":
			next.AnalystText = change.TextValue
		case "timeline.mitre_stage_text":
			next.MitreStageText = change.TextValue
		case "timeline.device_object_text":
			next.DeviceObjectText = change.TextValue
		case "timeline.ip_address_text":
			next.IPAddressText = change.TextValue
		case "timeline.activity_utc_text":
			next.ActivityUTCText = change.TextValue
			next.ActivityUTCGenerated = false
		case "timeline.activity_local_text":
			next.ActivityLocalText = change.TextValue
			next.ActivityLocalGenerated = false
		case "timeline.raw_activity_text":
			next.RawActivityText = change.TextValue
		case "timeline.activity_synopsis_text":
			next.ActivitySynopsisText = change.TextValue
		case "timeline.data_source_text":
			next.DataSourceText = change.TextValue
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
	profile, err := getTimeConversionProfileTx(ctx, tx, current.IncidentID, now.UTC())
	if err != nil {
		return MutationResult{}, err
	}
	applyTimelineTimeConversion(&next, profile)

	beforeProjected := projectRecord(current, nil)
	if err := hydrateProjectedCollections(ctx, tx, &beforeProjected); err != nil {
		return MutationResult{}, err
	}
	materialChanged := hasMaterialChange(current, next)
	if mentionChanged {
		mentionProjectionRefresh, err := s.applyPatchMentionActionsTx(ctx, tx, actor, current.IncidentID, recordID, request.CanonicalChange, now.UTC())
		if err != nil {
			return MutationResult{}, err
		}
		if err := s.rebuildMentionEntityProjectionsTx(ctx, tx, current.IncidentID, mentionProjectionRefresh); err != nil {
			return MutationResult{}, err
		}
	}
	var tagMutations []recordTagMutation
	if tagChanged {
		var err error
		tagMutations, err = applyPatchTagActionsTx(ctx, tx, actor.ID, current.IncidentID, recordID, request.CanonicalChange, now.UTC())
		if err != nil {
			return MutationResult{}, err
		}
		tagChanged = len(tagMutations) > 0
	}
	var attachedEvidenceMutations []attachedEvidenceMutation
	if evidenceChanged {
		var err error
		attachedEvidenceMutations, err = applyPatchAttachedEvidenceActionsTx(ctx, tx, actor.ID, current.IncidentID, recordID, request.CanonicalChange, now.UTC())
		if err != nil {
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
   SET date_entered_text = $2,
       analyst_text = $3,
       mitre_stage_text = $4,
       device_object_text = $5,
       ip_address_text = $6,
       activity_utc_text = $7,
       activity_local_text = $8,
       raw_activity_text = $9,
       activity_synopsis_text = $10,
       data_source_text = $11,
       activity_utc_generated = $12,
       activity_local_generated = $13,
       activity_time_pair_state = $14,
       capture_state = $15,
       row_version = $16,
       edited_at = $17,
       updated_by_user_id = $18,
       reviewed_at = $19,
       reviewed_by_user_id = $20
 WHERE record_id = $1
RETURNING recorded_at
`, recordID, next.DateEnteredText, next.AnalystText, next.MitreStageText, next.DeviceObjectText, next.IPAddressText, next.ActivityUTCText, next.ActivityLocalText, next.RawActivityText, next.ActivitySynopsisText, next.DataSourceText, next.ActivityUTCGenerated, next.ActivityLocalGenerated, next.ActivityTimePairState, next.CaptureState, next.RowVersion, next.EditedAt, actor.ID, next.ReviewedAt, next.ReviewedByUserID).Scan(&next.RecordedAt); err != nil {
		return MutationResult{}, fmt.Errorf("update timeline record: %w", err)
	}

	afterProjected := projectRecord(next, nil)
	if err := hydrateProjectedCollections(ctx, tx, &afterProjected); err != nil {
		return MutationResult{}, err
	}
	changeSetID, err := s.revisionsStore.InsertChangeSetTx(ctx, tx, timelineChangeSetParams{
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

	beforeRow := buildRow(beforeProjected)
	afterRow := buildRow(afterProjected)
	beforeVersion := versionID(current.RecordID, current.RowVersion)
	afterVersion := versionID(next.RecordID, next.RowVersion)
	if err := s.revisionsStore.InsertMutationTx(ctx, tx, timelineMutationParams{
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
	if err := s.insertAttachedEvidenceMutationEntriesTx(ctx, tx, changeSetID, 2, attachedEvidenceMutations); err != nil {
		return MutationResult{}, err
	}
	if err := s.insertRecordTagMutationEntriesTx(ctx, tx, changeSetID, 2+len(attachedEvidenceMutations), tagMutations); err != nil {
		return MutationResult{}, err
	}
	if err := s.revisionsStore.InsertRecordRevisionTx(ctx, tx, timelineRecordRevisionParams{
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
	if err := s.idempotencyStore.InsertRouteIdempotencyPayload(ctx, tx, idempotencyKey, nil, requestHash, http.StatusOK, payload); err != nil {
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
	serverValue, ok := rowCellValue(buildRow(current), change.FieldKey)
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
		case "add_token", "add_tag":
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
		case "remove_record_ref", "remove_tag":
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
		rawText = ""
		displayText = action.RawText
	}
	item := map[string]any{
		"item_ref":     clientCollectionItemRef(fieldKey, action, requestHash, actionIndex),
		"display_text": displayText,
		"raw_text":     rawText,
	}
	if fieldKey == "timeline.tags" {
		item["item_kind"] = "tag"
		tagID := clientCollectionLocalUUID(fieldKey, action, requestHash, actionIndex)
		item["item_ref"] = "record_tag:client:" + tagID.String()
		item["tag_id"] = tagID.String()
		delete(item, "raw_text")
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

func clientCollectionLocalUUID(fieldKey string, action CollectionAction, requestHash []byte, actionIndex int) uuid.UUID {
	sum := hashRequestPayload(map[string]any{
		"request_hash": base64.RawURLEncoding.EncodeToString(requestHash),
		"field_key":    fieldKey,
		"action_index": actionIndex,
		"op":           action.Op,
		"text":         action.NormalizedText,
	})
	return uuid.NewSHA1(uuid.NameSpaceOID, sum)
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

func loadSourceRecordTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (sourceRecord, error) {
	row := tx.QueryRow(ctx, `
SELECT
    e.record_id,
    e.incident_id,
    e.date_entered_text,
    e.analyst_text,
    e.mitre_stage_text,
    e.device_object_text,
    e.ip_address_text,
    e.activity_utc_text,
    e.activity_local_text,
    e.raw_activity_text,
    e.activity_synopsis_text,
    e.data_source_text,
    e.activity_utc_generated,
    e.activity_local_generated,
    e.activity_time_pair_state,
    e.raw_capture,
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
	var rawCapture []byte
	if err := row.Scan(
		&record.RecordID,
		&record.IncidentID,
		&record.DateEnteredText,
		&record.AnalystText,
		&record.MitreStageText,
		&record.DeviceObjectText,
		&record.IPAddressText,
		&record.ActivityUTCText,
		&record.ActivityLocalText,
		&record.RawActivityText,
		&record.ActivitySynopsisText,
		&record.DataSourceText,
		&record.ActivityUTCGenerated,
		&record.ActivityLocalGenerated,
		&record.ActivityTimePairState,
		&rawCapture,
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
	if len(rawCapture) == 0 {
		record.RawCapture = map[string]any{}
	} else if err := json.Unmarshal(rawCapture, &record.RawCapture); err != nil {
		return sourceRecord{}, fmt.Errorf("decode timeline raw capture: %w", err)
	}
	if record.RawCapture == nil {
		record.RawCapture = map[string]any{}
	}
	return record, nil
}

func loadSourceRecordForIncidentTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID) (sourceRecord, error) {
	row := tx.QueryRow(ctx, `
SELECT
    e.record_id,
    e.incident_id,
    e.date_entered_text,
    e.analyst_text,
    e.mitre_stage_text,
    e.device_object_text,
    e.ip_address_text,
    e.activity_utc_text,
    e.activity_local_text,
    e.raw_activity_text,
    e.activity_synopsis_text,
    e.data_source_text,
    e.activity_utc_generated,
    e.activity_local_generated,
    e.activity_time_pair_state,
    e.raw_capture,
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
 AND r.incident_id = e.incident_id
 AND r.record_type = 'timeline_event'
 AND r.deleted_at IS NULL
WHERE e.record_id = $1
  AND e.incident_id = $2
FOR UPDATE OF e, r
`, recordID, incidentID)

	var record sourceRecord
	var rawCapture []byte
	if err := row.Scan(
		&record.RecordID,
		&record.IncidentID,
		&record.DateEnteredText,
		&record.AnalystText,
		&record.MitreStageText,
		&record.DeviceObjectText,
		&record.IPAddressText,
		&record.ActivityUTCText,
		&record.ActivityLocalText,
		&record.RawActivityText,
		&record.ActivitySynopsisText,
		&record.DataSourceText,
		&record.ActivityUTCGenerated,
		&record.ActivityLocalGenerated,
		&record.ActivityTimePairState,
		&rawCapture,
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
		return sourceRecord{}, fmt.Errorf("load timeline record for incident: %w", err)
	}
	if len(rawCapture) == 0 {
		record.RawCapture = map[string]any{}
	} else if err := json.Unmarshal(rawCapture, &record.RawCapture); err != nil {
		return sourceRecord{}, fmt.Errorf("decode timeline raw capture: %w", err)
	}
	if record.RawCapture == nil {
		record.RawCapture = map[string]any{}
	}
	return record, nil
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
	return !stringPointersEqual(current.DateEnteredText, next.DateEnteredText) ||
		!stringPointersEqual(current.AnalystText, next.AnalystText) ||
		!stringPointersEqual(current.MitreStageText, next.MitreStageText) ||
		!stringPointersEqual(current.DeviceObjectText, next.DeviceObjectText) ||
		!stringPointersEqual(current.IPAddressText, next.IPAddressText) ||
		!stringPointersEqual(current.ActivityUTCText, next.ActivityUTCText) ||
		!stringPointersEqual(current.ActivityLocalText, next.ActivityLocalText) ||
		!stringPointersEqual(current.RawActivityText, next.RawActivityText) ||
		!stringPointersEqual(current.ActivitySynopsisText, next.ActivitySynopsisText) ||
		!stringPointersEqual(current.DataSourceText, next.DataSourceText) ||
		current.ActivityUTCGenerated != next.ActivityUTCGenerated ||
		current.ActivityLocalGenerated != next.ActivityLocalGenerated ||
		current.ActivityTimePairState != next.ActivityTimePairState ||
		!reflect.DeepEqual(current.RawCapture, next.RawCapture)
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

func deriveActivitySortTS(utcText *string, localText *string) *time.Time {
	if parsed := parseTimelineUTCText(utcText); parsed != nil {
		return parsed
	}
	if parsed := parseTimelineLocalText(localText); parsed != nil {
		return parsed
	}
	return nil
}

func deriveDateEnteredSortDay(text *string) *time.Time {
	if text == nil || *text == "" {
		return nil
	}
	if parsed, err := time.Parse("2006-01-02", *text); err == nil {
		day := time.Date(parsed.Year(), parsed.Month(), parsed.Day(), 0, 0, 0, 0, time.UTC)
		return &day
	}
	if parsed := parseTimelineUTCText(text); parsed != nil {
		day := time.Date(parsed.UTC().Year(), parsed.UTC().Month(), parsed.UTC().Day(), 0, 0, 0, 0, time.UTC)
		return &day
	}
	if parsed := parseTimelineLocalText(text); parsed != nil {
		day := time.Date(parsed.UTC().Year(), parsed.UTC().Month(), parsed.UTC().Day(), 0, 0, 0, 0, time.UTC)
		return &day
	}
	return nil
}

func parseTimelineUTCText(text *string) *time.Time {
	if text == nil || *text == "" {
		return nil
	}
	parsed, err := time.Parse("2006-01-02T15:04:05Z", *text)
	if err != nil {
		return nil
	}
	utc := parsed.UTC()
	return &utc
}

func parseTimelineLocalText(text *string) *time.Time {
	if text == nil || *text == "" {
		return nil
	}
	parsed, err := time.Parse("2006-01-02T15:04:05-07:00", *text)
	if err != nil {
		return nil
	}
	utc := parsed.UTC()
	return &utc
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

func optionalIntFromPG(value pgtype.Int4) *int {
	if !value.Valid {
		return nil
	}
	parsed := int(value.Int32)
	return &parsed
}

func optionalDateFromPG(value pgtype.Date) *time.Time {
	if !value.Valid {
		return nil
	}
	utc := value.Time.UTC()
	return &utc
}
