package timeline

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	sqlc "example.com/todo/cartulary/internal/gen/sql"
	"example.com/todo/cartulary/internal/modules/links"
	"example.com/todo/cartulary/internal/modules/projections"
	"example.com/todo/cartulary/internal/modules/revisions"
	"example.com/todo/cartulary/internal/platform/authn"
)

var (
	ErrRecordNotFound     = errors.New("timeline: record not found")
	ErrRowVersionConflict = errors.New("timeline: row version conflict")
	ErrIllegalTransition  = errors.New("timeline: illegal transition")
	ErrNoEffectiveChange  = errors.New("timeline: no effective change")
)

type Store struct {
	pool            *pgxpool.Pool
	authStore       *authn.Store
	revisionsStore  *revisions.Store
	projectionStore *projections.Store
	linkStore       *links.Store
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
	RowVersion       int64
	ChangedFieldKeys []string
	Row              projectedRecord
}

func NewStore(pool *pgxpool.Pool) *Store {
	return &Store{
		pool:            pool,
		authStore:       authn.NewStore(pool),
		revisionsStore:  revisions.NewStore(),
		projectionStore: projections.NewStore(pool),
		linkStore:       links.NewStore(),
	}
}

func (s *Store) GetRecordIncident(ctx context.Context, recordID uuid.UUID) (uuid.UUID, error) {
	var incidentID uuid.UUID
	if err := s.pool.QueryRow(ctx, `SELECT incident_id FROM timeline_events WHERE record_id = $1`, recordID).Scan(&incidentID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return uuid.UUID{}, ErrRecordNotFound
		}
		return uuid.UUID{}, fmt.Errorf("get timeline record incident: %w", err)
	}
	return incidentID, nil
}

func (s *Store) QueryRows(ctx context.Context, incidentID uuid.UUID) ([]map[string]any, error) {
	rows, err := sqlc.New(s.pool).ListTimelineProjectionRows(ctx, pgUUID(incidentID))
	if err != nil {
		return nil, fmt.Errorf("list timeline projection rows: %w", err)
	}
	result := make([]map[string]any, 0, len(rows))
	for _, row := range rows {
		projected, err := projectedRecordFromSQL(row)
		if err != nil {
			return nil, err
		}
		result = append(result, BuildRow(projected))
	}
	return result, nil
}

func (s *Store) CreateRow(ctx context.Context, actor authn.UserRecord, incidentID uuid.UUID, request CreateRequest, requestHash []byte, requestID string, now time.Time) (MutationResult, error) {
	scopeKey := incidentID.String() + ":" + TimelineViewSchemaID
	if existing, err := s.authStore.GetRouteIdempotency(ctx, createRouteKey, scopeKey, request.ClientTxnID); err == nil {
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

	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return MutationResult{}, fmt.Errorf("begin timeline create transaction: %w", err)
	}
	defer func() {
		_ = tx.Rollback(ctx)
	}()

	current := sourceRecord{
		IncidentID:      incidentID,
		OccurredAt:      request.OccurredAt,
		Summary:         request.Summary,
		Details:         request.Details,
		SourceText:      request.SourceText,
		CaptureState:    "rough",
		RowVersion:      1,
		RecordedAt:      now.UTC(),
		EditedAt:        now.UTC(),
		CreatedByUserID: actor.ID,
		UpdatedByUserID: actor.ID,
	}
	if err := tx.QueryRow(ctx, `
INSERT INTO timeline_events (
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
VALUES ($1, $2, $3, $4, $5, 'rough', 1, $6, $6, $7, $7)
RETURNING record_id
`, incidentID, request.OccurredAt, request.Summary, request.Details, request.SourceText, now.UTC(), actor.ID).Scan(&current.RecordID); err != nil {
		return MutationResult{}, fmt.Errorf("insert timeline record: %w", err)
	}

	projected := projectRecord(current, nil)
	changeSetID, err := s.revisionsStore.InsertChangeSetTx(ctx, tx, revisions.ChangeSetParams{
		IncidentID:  incidentID,
		ActorUserID: actor.ID,
		Source:      createRouteKey,
		ClientTxnID: &request.ClientTxnID,
		RequestID:   &requestID,
		CreatedAt:   now.UTC(),
	})
	if err != nil {
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

	payload := BuildMutationPayload(projected, changeSetID)
	if err := insertRouteIdempotency(ctx, tx, createRouteKey, scopeKey, request.ClientTxnID, actor.ID, requestHash, http.StatusCreated, payload); err != nil {
		if authn.IsUniqueViolation(err) {
			return MutationResult{}, authn.ErrClientTxnConflict
		}
		return MutationResult{}, err
	}

	if err := tx.Commit(ctx); err != nil {
		return MutationResult{}, fmt.Errorf("commit timeline create transaction: %w", err)
	}
	return MutationResult{
		Payload:          payload,
		StatusCode:       http.StatusCreated,
		IncidentID:       incidentID,
		RecordID:         current.RecordID,
		ChangeSetID:      changeSetID,
		RowVersion:       projected.RowVersion,
		ChangedFieldKeys: ComputeChangedFieldKeys(nil, projected),
		Row:              projected,
	}, nil
}

func (s *Store) PatchRow(ctx context.Context, actor authn.UserRecord, recordID uuid.UUID, request PatchRequest, requestHash []byte, requestID string, now time.Time) (MutationResult, error) {
	if existing, err := s.authStore.GetRouteIdempotency(ctx, patchRouteKey, recordID.String(), request.ClientTxnID); err == nil {
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
	if current.RowVersion != request.BaseRowVersion {
		return MutationResult{}, ErrRowVersionConflict
	}
	if current.CaptureState == "superseded" {
		return MutationResult{}, ErrIllegalTransition
	}

	next := current
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
		}
	}

	beforeProjected := projectRecord(current, nil)
	materialChanged := hasMaterialChange(current, next)
	if !materialChanged {
		return MutationResult{}, ErrNoEffectiveChange
	}

	if current.CaptureState == "rough" || current.CaptureState == "reviewed" {
		next.CaptureState = "enriched"
	}
	next.RowVersion = current.RowVersion + 1
	next.EditedAt = now.UTC()
	next.UpdatedByUserID = actor.ID
	if current.CaptureState == "reviewed" {
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
	changeSetID, err := s.revisionsStore.InsertChangeSetTx(ctx, tx, revisions.ChangeSetParams{
		IncidentID:  current.IncidentID,
		ActorUserID: actor.ID,
		Source:      patchRouteKey,
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
	if err := insertRouteIdempotency(ctx, tx, patchRouteKey, recordID.String(), request.ClientTxnID, actor.ID, requestHash, http.StatusOK, payload); err != nil {
		if authn.IsUniqueViolation(err) {
			return MutationResult{}, authn.ErrClientTxnConflict
		}
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
		RowVersion:       afterProjected.RowVersion,
		ChangedFieldKeys: ComputeChangedFieldKeys(&beforeProjected, afterProjected),
		Row:              afterProjected,
	}, nil
}

func (s *Store) MarkReviewed(ctx context.Context, actor authn.UserRecord, recordID uuid.UUID, request ActionRequest, requestHash []byte, requestID string, now time.Time) (MutationResult, error) {
	return s.applyAction(ctx, actor, reviewRouteKey, recordID, request.BaseRowVersion, request.ClientTxnID, requestHash, requestID, now, request.Reason, nil, func(current sourceRecord) (sourceRecord, *links.SupersedesLink, *string, error) {
		if current.CaptureState != "rough" && current.CaptureState != "enriched" {
			return sourceRecord{}, nil, nil, ErrIllegalTransition
		}
		next := current
		next.CaptureState = "reviewed"
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
		if current.CaptureState != "rough" && current.CaptureState != "enriched" && current.CaptureState != "reviewed" {
			return sourceRecord{}, nil, nil, ErrIllegalTransition
		}
		if request.ReplacementRecordID != nil {
			if *request.ReplacementRecordID == current.RecordID {
				return sourceRecord{}, nil, nil, ErrIllegalTransition
			}
		}

		next := current
		next.CaptureState = "superseded"
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
	if existing, err := s.authStore.GetRouteIdempotency(ctx, routeKey, recordID.String(), clientTxnID); err == nil {
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

	var validatedReplacementID *uuid.UUID
	if replacementRecordID != nil {
		replacementRecord, err := loadSourceRecordTx(ctx, tx, *replacementRecordID)
		if err != nil {
			return MutationResult{}, err
		}
		if replacementRecord.IncidentID != current.IncidentID {
			return MutationResult{}, ErrIllegalTransition
		}
		replacementID := replacementRecord.RecordID
		validatedReplacementID = &replacementID
	}

	next, _, effectiveReason, err := prepare(current)
	if err != nil {
		return MutationResult{}, err
	}

	beforeProjected := projectRecord(current, nil)
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
				return MutationResult{}, ErrIllegalTransition
			}
			return MutationResult{}, err
		}
		insertedLink = &link
	}

	afterProjected := projectRecord(next, validatedReplacementID)
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
	if err := insertRouteIdempotency(ctx, tx, routeKey, recordID.String(), clientTxnID, actor.ID, requestHash, http.StatusOK, payload); err != nil {
		if authn.IsUniqueViolation(err) {
			return MutationResult{}, authn.ErrClientTxnConflict
		}
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
		RowVersion:       afterProjected.RowVersion,
		ChangedFieldKeys: ComputeChangedFieldKeys(&beforeProjected, afterProjected),
		Row:              afterProjected,
	}, nil
}

func loadSourceRecordTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (sourceRecord, error) {
	row := tx.QueryRow(ctx, `
SELECT
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
    updated_by_user_id,
    reviewed_by_user_id,
    reviewed_at,
    superseded_by_user_id,
    superseded_at
FROM timeline_events
WHERE record_id = $1
FOR UPDATE
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

func projectedRecordFromSQL(row sqlc.TimelineGridProjection) (projectedRecord, error) {
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

func insertRouteIdempotency(ctx context.Context, tx pgx.Tx, routeKey string, scopeKey string, clientTxnID string, actorUserID uuid.UUID, requestHash []byte, statusCode int, payload map[string]any) error {
	responseJSON, err := jsonMarshal(payload)
	if err != nil {
		return fmt.Errorf("marshal idempotency payload: %w", err)
	}
	if _, err := tx.Exec(ctx, `
INSERT INTO route_idempotency (
    route_key,
    scope_key,
    client_txn_id,
    actor_user_id,
    target_user_id,
    request_hash,
    status_code,
    response_json
)
VALUES ($1, $2, $3, $4, NULL, $5, $6, $7)
`, routeKey, scopeKey, clientTxnID, actorUserID, requestHash, statusCode, responseJSON); err != nil {
		return fmt.Errorf("insert route idempotency: %w", err)
	}
	return nil
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

func jsonMarshal(value any) ([]byte, error) {
	return json.Marshal(value)
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
