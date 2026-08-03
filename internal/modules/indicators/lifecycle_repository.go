package indicators

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func (lifecycleRepository) insertTx(ctx context.Context, tx pgx.Tx, actorUserID uuid.UUID, params IndicatorLifecycleAppendParams, createdAt time.Time) (IndicatorLifecycleIntervalRecord, error) {
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
		SupportRefs:       append([]string(nil), params.SupportRefs...),
		Assessor:          cloneStringPointer(params.Assessor),
		AssessedAt:        createdAt,
		RowVersion:        1,
		CreatedByUserID:   actorUserID,
		CreatedAt:         createdAt,
	}
	if record.ValidFrom.IsZero() {
		record.ValidFrom = createdAt
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
`, record.IncidentID, record.IndicatorRecordID, record.LifecycleState, record.ValidFrom.UTC(), record.ValidTo, record.Confidence, record.Rationale, mustJSON(record.SupportRefs), record.Assessor, record.AssessedAt.UTC(), record.CreatedByUserID, record.CreatedAt.UTC()).Scan(&record.IntervalID); err != nil {
		return IndicatorLifecycleIntervalRecord{}, fmt.Errorf("insert indicator lifecycle interval: %w", err)
	}
	return record, nil
}
