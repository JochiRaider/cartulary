package indicators

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/indicators/internal/identity"
	indicatororigin "github.com/JochiRaider/cartulary/internal/modules/indicators/internal/origin"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
)

func (repository observationRepository) insertTx(ctx context.Context, tx pgx.Tx, actorUserID uuid.UUID, params IndicatorObservationCreateParams, createdAt time.Time) (IndicatorObservationRecord, error) {
	if params.IncidentID == uuid.Nil || params.SourceRecordID == uuid.Nil {
		return IndicatorObservationRecord{}, ErrInvalidCreateRequest
	}
	if strings.TrimSpace(params.SourceFieldKey) == "" || strings.TrimSpace(params.originLocator) == "" {
		return IndicatorObservationRecord{}, ErrInvalidCreateRequest
	}
	if params.originKind != indicatororigin.ManualEntry {
		return IndicatorObservationRecord{}, ErrInvalidCreateRequest
	}
	originKind := params.originKind
	observedText := params.observedText
	if observedText == "" || strings.ContainsRune(observedText, 0) {
		return IndicatorObservationRecord{}, ErrInvalidCreateRequest
	}
	if err := repository.validateSourceIncidentTx(ctx, tx, params.IncidentID, params.SourceRecordID); err != nil {
		return IndicatorObservationRecord{}, err
	}
	parsedIndicatorType, normalizedCandidate, err := identity.NormalizeObservationCandidate(params.ParsedIndicatorType, params.normalizedCandidate, observedText)
	if err != nil {
		return IndicatorObservationRecord{}, err
	}
	resolutionStatus := "unresolved"
	var resolvedByUserID *uuid.UUID
	var resolvedAt *time.Time
	var resolutionMethod *string
	if params.ResolvedIndicatorRecordID != nil {
		if err := (sourceRepository{}).validateIncidentTx(ctx, tx, params.IncidentID, *params.ResolvedIndicatorRecordID); err != nil {
			return IndicatorObservationRecord{}, err
		}
		resolutionStatus = "resolved"
		resolvedByUserID = &actorUserID
		value := createdAt.UTC()
		resolvedAt = &value
		if resolutionMethod == nil {
			value := observationCreateSource
			resolutionMethod = &value
		}
	}
	record := IndicatorObservationRecord{
		IncidentID:                params.IncidentID,
		SourceRecordID:            params.SourceRecordID,
		SourceFieldKey:            params.SourceFieldKey,
		OriginKind:                originKind.String(),
		OriginLocator:             params.originLocator,
		ObservedText:              observedText,
		ParsedIndicatorType:       parsedIndicatorType,
		NormalizedCandidate:       normalizedCandidate,
		ResolutionStatus:          resolutionStatus,
		ResolvedIndicatorRecordID: params.ResolvedIndicatorRecordID,
		RowVersion:                1,
		CreatedByUserID:           actorUserID,
		CreatedAt:                 createdAt,
		ResolvedByUserID:          resolvedByUserID,
		ResolvedAt:                resolvedAt,
		ResolutionMethod:          resolutionMethod,
	}
	if err := tx.QueryRow(ctx, `
INSERT INTO indicator_observations (
    incident_id,
    source_record_id,
    source_field_key,
    origin_kind,
    origin_locator,
    observed_text,
    parsed_indicator_type,
    normalized_candidate,
    resolution_status,
    resolved_indicator_record_id,
    row_version,
    created_by_user_id,
    created_at,
    resolved_by_user_id,
    resolved_at,
    resolution_method
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, 1, $11, $12, $13, $14, $15)
RETURNING indicator_observation_id
`, record.IncidentID, record.SourceRecordID, record.SourceFieldKey, record.OriginKind, record.OriginLocator, record.ObservedText, record.ParsedIndicatorType, record.NormalizedCandidate, record.ResolutionStatus, record.ResolvedIndicatorRecordID, record.CreatedByUserID, record.CreatedAt.UTC(), record.ResolvedByUserID, record.ResolvedAt, record.ResolutionMethod).Scan(&record.ObservationID); err != nil {
		return IndicatorObservationRecord{}, fmt.Errorf("insert indicator observation: %w", err)
	}
	return record, nil
}

func (observationRepository) loadTx(ctx context.Context, tx pgx.Tx, observationID uuid.UUID, forUpdate bool) (IndicatorObservationRecord, error) {
	return loadIndicatorObservation(ctx, tx, observationID, forUpdate)
}

func loadIndicatorObservation(ctx context.Context, db interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}, observationID uuid.UUID, forUpdate bool) (IndicatorObservationRecord, error) {
	query := `
SELECT
    indicator_observation_id,
    incident_id,
    source_record_id,
    source_field_key,
    origin_kind,
    origin_locator,
    observed_text,
    parsed_indicator_type,
    normalized_candidate,
    resolution_status,
    resolved_indicator_record_id,
    row_version,
    created_by_user_id,
    created_at,
    resolved_by_user_id,
    resolved_at,
    resolution_method,
    deleted_at,
    deleted_by_user_id
  FROM indicator_observations
 WHERE indicator_observation_id = $1
   AND deleted_at IS NULL`
	if forUpdate {
		query += ` FOR UPDATE`
	}
	record, err := scanIndicatorObservationRecord(db.QueryRow(ctx, query, observationID))
	if errors.Is(err, pgx.ErrNoRows) {
		return IndicatorObservationRecord{}, ErrIndicatorObservationNotFound
	}
	if err != nil {
		return IndicatorObservationRecord{}, fmt.Errorf("load indicator observation: %w", err)
	}
	return record, nil
}

func (observationRepository) updateTransitionTx(ctx context.Context, tx pgx.Tx, next IndicatorObservationRecord, expectedVersion int64) error {
	tag, err := tx.Exec(ctx, `
UPDATE indicator_observations
   SET resolution_status = $2,
       resolved_indicator_record_id = $3,
       resolved_by_user_id = $4,
       resolved_at = $5,
       resolution_method = $6,
       row_version = $7
 WHERE indicator_observation_id = $1
   AND row_version = $8
   AND deleted_at IS NULL
`, next.ObservationID, next.ResolutionStatus, next.ResolvedIndicatorRecordID, next.ResolvedByUserID, next.ResolvedAt, next.ResolutionMethod, next.RowVersion, expectedVersion)
	if err != nil {
		return fmt.Errorf("update indicator observation transition: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return ErrRowVersionConflict
	}
	return nil
}

func (observationRepository) validateSourceIncidentTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, sourceRecordID uuid.UUID) error {
	var exists bool
	if err := tx.QueryRow(ctx, `
SELECT EXISTS (
    SELECT 1
      FROM records
     WHERE record_id = $1
       AND incident_id = $2
)
`, sourceRecordID, incidentID).Scan(&exists); err != nil {
		return fmt.Errorf("validate source record incident: %w", err)
	}
	if !exists {
		return revisions.ErrRecordDeletedUseRestore
	}
	return nil
}

func (observationRepository) listBySource(ctx context.Context, db interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
}, sourceRecordID uuid.UUID, afterCreatedAt *time.Time, afterID *uuid.UUID, limit int) ([]IndicatorObservationRecord, error) {
	return listObservationRows(ctx, db, `o.source_record_id = $1`, sourceRecordID, afterCreatedAt, afterID, limit)
}

func (observationRepository) listByIndicator(ctx context.Context, db interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
}, indicatorID uuid.UUID, afterCreatedAt *time.Time, afterID *uuid.UUID, limit int) ([]IndicatorObservationRecord, error) {
	return listObservationRows(ctx, db, `o.resolved_indicator_record_id = $1`, indicatorID, afterCreatedAt, afterID, limit)
}

func listObservationRows(ctx context.Context, db interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
}, predicate string, recordID uuid.UUID, afterCreatedAt *time.Time, afterID *uuid.UUID, limit int) ([]IndicatorObservationRecord, error) {
	if limit < 1 {
		return nil, ErrInvalidCreateRequest
	}
	rows, err := db.Query(ctx, `
SELECT
    o.indicator_observation_id, o.incident_id, o.source_record_id, o.source_field_key,
    o.origin_kind, o.origin_locator, o.observed_text, o.parsed_indicator_type,
    o.normalized_candidate, o.resolution_status, o.resolved_indicator_record_id,
    o.row_version, o.created_by_user_id, o.created_at, o.resolved_by_user_id,
    o.resolved_at, o.resolution_method, o.deleted_at, o.deleted_by_user_id
  FROM indicator_observations o
 WHERE `+predicate+`
   AND o.deleted_at IS NULL
   AND ($2::timestamptz IS NULL OR (o.created_at, o.indicator_observation_id) < ($2, $3))
 ORDER BY o.created_at DESC, o.indicator_observation_id DESC
 LIMIT $4
`, recordID, afterCreatedAt, afterID, limit)
	if err != nil {
		return nil, fmt.Errorf("list indicator observations: %w", err)
	}
	defer rows.Close()
	result := make([]IndicatorObservationRecord, 0, limit)
	for rows.Next() {
		record, err := scanIndicatorObservationRecord(rows)
		if err != nil {
			return nil, fmt.Errorf("scan indicator observation list: %w", err)
		}
		result = append(result, record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate indicator observations: %w", err)
	}
	return result, nil
}
