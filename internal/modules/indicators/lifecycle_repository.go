package indicators

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func insertIndicatorLifecycleIntervalTx(ctx context.Context, tx pgx.Tx, actorUserID uuid.UUID, params IndicatorLifecycleAppendParams, createdAt time.Time) (IndicatorLifecycleIntervalRecord, error) {
	if strings.TrimSpace(params.LifecycleState) == "" {
		return IndicatorLifecycleIntervalRecord{}, ErrInvalidCreateRequest
	}
	record := IndicatorLifecycleIntervalRecord{
		IncidentID:        params.IncidentID,
		IndicatorRecordID: params.IndicatorRecordID,
		LifecycleState:    strings.TrimSpace(params.LifecycleState),
		ValidFrom:         params.ValidFrom.UTC(),
		ValidTo:           normalizeTimePointer(params.ValidTo),
		Confidence:        cloneIntPointer(params.Confidence),
		Rationale:         cloneStringPointer(params.Rationale),
		SupportRefs:       append(make([]uuid.UUID, 0, len(params.SupportRefs)), params.SupportRefs...),
		Assessor:          cloneStringPointer(params.Assessor),
		AssessedAt:        createdAt,
		RowVersion:        1,
		CreatedByUserID:   actorUserID,
		CreatedAt:         createdAt,
	}
	if record.ValidFrom.IsZero() {
		record.ValidFrom = createdAt
	}
	supportRefsJSON, err := json.Marshal(record.SupportRefs)
	if err != nil {
		return IndicatorLifecycleIntervalRecord{}, fmt.Errorf("encode indicator lifecycle support refs: %w", err)
	}
	if err := tx.QueryRow(ctx, `
INSERT INTO indicator_state_intervals (
    incident_id,
    indicator_record_id,
    lifecycle_state,
    valid_from,
    valid_to,
    confidence,
    rationale,
    support_refs,
    assessor,
    assessed_at,
    row_version,
    created_by_user_id,
    created_at
)
VALUES ($1, $2, $3, $4, $5, $6, $7, $8::jsonb, $9, $10, 1, $11, $12)
RETURNING indicator_state_interval_id
`, record.IncidentID, record.IndicatorRecordID, record.LifecycleState, record.ValidFrom.UTC(), record.ValidTo, record.Confidence, record.Rationale, supportRefsJSON, record.Assessor, record.AssessedAt.UTC(), record.CreatedByUserID, record.CreatedAt.UTC()).Scan(&record.IntervalID); err != nil {
		return IndicatorLifecycleIntervalRecord{}, fmt.Errorf("insert indicator lifecycle interval: %w", err)
	}
	return record, nil
}

func listIndicatorLifecycleIntervals(ctx context.Context, db interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
}, indicatorID uuid.UUID, afterValidFrom *time.Time, afterID *uuid.UUID, limit int) ([]IndicatorLifecycleIntervalRecord, error) {
	if limit < 1 {
		return nil, ErrInvalidCreateRequest
	}
	rows, err := db.Query(ctx, `
SELECT
    indicator_state_interval_id, incident_id, indicator_record_id, lifecycle_state,
    valid_from, valid_to, confidence, rationale, support_refs, assessor, assessed_at,
    row_version, created_by_user_id, created_at, deleted_at, deleted_by_user_id
  FROM indicator_state_intervals
 WHERE indicator_record_id = $1
   AND deleted_at IS NULL
   AND ($2::timestamptz IS NULL OR (valid_from, indicator_state_interval_id) < ($2, $3))
 ORDER BY valid_from DESC, indicator_state_interval_id DESC
 LIMIT $4
`, indicatorID, afterValidFrom, afterID, limit)
	if err != nil {
		return nil, fmt.Errorf("list indicator lifecycle intervals: %w", err)
	}
	defer rows.Close()
	result := make([]IndicatorLifecycleIntervalRecord, 0, limit)
	for rows.Next() {
		var record IndicatorLifecycleIntervalRecord
		var supportRefsJSON []byte
		if err := rows.Scan(
			&record.IntervalID, &record.IncidentID, &record.IndicatorRecordID, &record.LifecycleState,
			&record.ValidFrom, &record.ValidTo, &record.Confidence, &record.Rationale, &supportRefsJSON,
			&record.Assessor, &record.AssessedAt, &record.RowVersion, &record.CreatedByUserID,
			&record.CreatedAt, &record.DeletedAt, &record.DeletedByUserID,
		); err != nil {
			return nil, fmt.Errorf("scan indicator lifecycle interval: %w", err)
		}
		if err := json.Unmarshal(supportRefsJSON, &record.SupportRefs); err != nil {
			return nil, fmt.Errorf("decode indicator lifecycle support refs: %w", err)
		}
		result = append(result, record)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("iterate indicator lifecycle intervals: %w", err)
	}
	return result, nil
}
