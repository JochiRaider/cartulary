package source

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/policy"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/internal/sourcecatalog"
)

func TouchSourceRowTx(ctx context.Context, tx pgx.Tx, catalog *sourcecatalog.Catalog, viewSchemaID string, recordID uuid.UUID, now time.Time) error {
	surface, ok := catalog.SurfaceByViewID(viewSchemaID)
	if !ok {
		return &policy.ValidationError{Field: "view_schema_id", ReasonCode: "unknown_view_schema"}
	}
	table := pgx.Identifier{surface.SourceTable}.Sanitize()
	if _, err := tx.Exec(ctx, fmt.Sprintf("UPDATE %s SET updated_at = $2 WHERE record_id = $1", table), recordID, now); err != nil {
		return fmt.Errorf("touch Tasks/Decisions source row: %w", err)
	}
	return nil
}

func NormalizeTaskLifecycleTx(
	ctx context.Context,
	tx pgx.Tx,
	recordID uuid.UUID,
	before policy.TaskLifecycleState,
	explicitCompletedAt bool,
	now time.Time,
) (bool, error) {
	changed := false
	after, err := LoadTaskLifecycleStateTx(ctx, tx, recordID)
	if err != nil {
		return false, err
	}
	if !policy.ValidTaskStatus(after.Status) {
		return false, &policy.ValidationError{Field: "task.status", ReasonCode: "invalid_value"}
	}
	if before.Status != after.Status && !policy.ValidTaskTransition(before.Status, after.Status) {
		return false, &policy.LifecycleValidationError{
			FromStatus: before.Status, ToStatus: after.Status,
			ReasonCode: "illegal_status_transition", ViolatedGuards: []string{"task.status"},
		}
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
	after, err = LoadTaskLifecycleStateTx(ctx, tx, recordID)
	if err != nil {
		return false, err
	}
	if err := policy.ValidateTaskState(after); err != nil {
		return false, err
	}
	return changed, nil
}

func LoadDecisionStatusTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (string, error) {
	var status string
	err := tx.QueryRow(ctx, `SELECT status FROM decisions WHERE record_id = $1`, recordID).Scan(&status)
	return status, err
}

func TouchSupersedingDecisionTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, now time.Time) error {
	if _, err := tx.Exec(ctx, `
UPDATE decisions
   SET updated_at = $2
 WHERE record_id = $1
`, recordID, now); err != nil {
		return fmt.Errorf("update superseding decision: %w", err)
	}
	return nil
}

func MarkSupersededDecisionTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, now time.Time) error {
	if _, err := tx.Exec(ctx, `
UPDATE decisions
   SET status = CASE WHEN status IN ('proposed', 'approved') THEN 'superseded' ELSE status END,
       updated_at = $2
 WHERE record_id = $1
`, recordID, now); err != nil {
		return fmt.Errorf("update superseded decision: %w", err)
	}
	return nil
}
