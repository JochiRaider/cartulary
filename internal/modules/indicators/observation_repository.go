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
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/fieldnorm"
)

func (repository observationRepository) insertTx(ctx context.Context, tx pgx.Tx, actorUserID uuid.UUID, params IndicatorObservationCreateParams, createdAt time.Time) (IndicatorObservationRecord, error) {
	if params.IncidentID == uuid.Nil || params.SourceRecordID == uuid.Nil {
		return IndicatorObservationRecord{}, ErrInvalidCreateRequest
	}
	if strings.TrimSpace(params.SourceFieldKey) == "" || strings.TrimSpace(params.OriginLocator) == "" {
		return IndicatorObservationRecord{}, ErrInvalidCreateRequest
	}
	originKind, err := params.Producer.originForWrite()
	if err != nil || originKind != params.originKind {
		return IndicatorObservationRecord{}, ErrInvalidObservationOrigin
	}
	observedText, ok := fieldnorm.NormalizeLine(params.ObservedText)
	if !ok {
		return IndicatorObservationRecord{}, ErrInvalidCreateRequest
	}
	if err := repository.validateSourceIncidentTx(ctx, tx, params.IncidentID, params.SourceRecordID); err != nil {
		return IndicatorObservationRecord{}, err
	}
	parsedIndicatorType, normalizedCandidate, err := identity.NormalizeObservationCandidate(params.ParsedIndicatorType, params.NormalizedCandidate, observedText)
	if err != nil {
		return IndicatorObservationRecord{}, err
	}
	resolutionStatus := "unresolved"
	var resolvedByUserID *uuid.UUID
	var resolvedAt *time.Time
	resolutionMethod := params.ResolutionMethod
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
		OriginKind:                originKind,
		OriginLocator:             params.OriginLocator,
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
`, record.IncidentID, record.SourceRecordID, record.SourceFieldKey, record.OriginKind.String(), record.OriginLocator, record.ObservedText, record.ParsedIndicatorType, record.NormalizedCandidate, record.ResolutionStatus, record.ResolvedIndicatorRecordID, record.CreatedByUserID, record.CreatedAt.UTC(), record.ResolvedByUserID, record.ResolvedAt, record.ResolutionMethod).Scan(&record.ObservationID); err != nil {
		return IndicatorObservationRecord{}, fmt.Errorf("insert indicator observation: %w", err)
	}
	return record, nil
}

func (observationRepository) loadTx(ctx context.Context, tx pgx.Tx, observationID uuid.UUID, forUpdate bool) (IndicatorObservationRecord, error) {
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
	record, err := scanIndicatorObservationRecord(tx.QueryRow(ctx, query, observationID))
	if errors.Is(err, pgx.ErrNoRows) {
		return IndicatorObservationRecord{}, ErrIndicatorObservationNotFound
	}
	if err != nil {
		return IndicatorObservationRecord{}, fmt.Errorf("load indicator observation: %w", err)
	}
	return record, nil
}

func (observationRepository) resolveTx(ctx context.Context, tx pgx.Tx, next IndicatorObservationRecord) error {
	tag, err := tx.Exec(ctx, `
UPDATE indicator_observations
   SET resolution_status = 'resolved',
       resolved_indicator_record_id = $2,
       resolved_by_user_id = $3,
       resolved_at = $4,
       resolution_method = $5,
       row_version = $6
 WHERE indicator_observation_id = $1
   AND deleted_at IS NULL
`, next.ObservationID, *next.ResolvedIndicatorRecordID, *next.ResolvedByUserID, *next.ResolvedAt, *next.ResolutionMethod, next.RowVersion)
	if err != nil {
		return fmt.Errorf("resolve indicator observation: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return ErrIndicatorObservationNotFound
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
