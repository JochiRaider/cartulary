package reference_data

import (
	"context"
	"errors"
	"slices"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// JobTerminalEffects closes owner preparation when Jobs itself terminates an
// execution, including exhaustion and inactive-profile cancellation. Normal
// owner finalization already closed the operation and is an exact no-op here.
// No verification, storage I/O, activation, or trust advancement occurs here.
type JobTerminalEffects struct{}

func (JobTerminalEffects) ApplyJobTerminalEffectsTx(ctx context.Context, tx pgx.Tx, disposition jobs.TerminalDisposition) error {
	if tx == nil || disposition.JobID == uuid.Nil || disposition.FinishedAt.IsZero() || !slices.Contains([]string{jobs.StatusSucceeded, jobs.StatusFailed, jobs.StatusCanceled}, disposition.Status) {
		return errors.New("reference pack: invalid terminal Job disposition")
	}
	var operation uuid.UUID
	var kind string
	err := tx.QueryRow(ctx, `SELECT operation_id,kind FROM reference_pack_operations WHERE job_id=$1 AND terminal_at IS NULL FOR UPDATE`, disposition.JobID).Scan(&operation, &kind)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil
	}
	if err != nil {
		return err
	}
	if disposition.Status == jobs.StatusSucceeded {
		// A successful Job is not evidence that its private preparation was ever
		// published. Reject the transition instead of promoting that preparation.
		return errors.New("reference pack: successful Job lacks terminal owner publication")
	}
	if !slices.Contains([]string{"import", "reverify", "refresh", "portable_retention"}, kind) {
		return errors.New("reference pack: unexpected Job-owned operation kind")
	}
	outcome := "execution_failed"
	if disposition.Status == jobs.StatusCanceled {
		outcome = "canceled"
	}
	return abortReferenceOperationTx(ctx, tx, operation, kind, outcome, disposition.FinishedAt)
}

func abortReferenceOperationTx(ctx context.Context, tx pgx.Tx, operation uuid.UUID, kind, outcome string, completed time.Time) error {
	var count int64
	if err := tx.QueryRow(ctx, `SELECT count(*) FROM reference_pack_operation_members WHERE operation_id=$1`, operation).Scan(&count); err != nil {
		return err
	}
	if kind == "import" && count == 0 {
		count = 1
	}
	result, err := packformat.EncodeAttemptResult(outcome, count, 0)
	if err != nil {
		return err
	}
	if kind == "import" || kind == "portable_retention" {
		if err := closeUnverifiedCandidatesTx(ctx, tx, operation); err != nil {
			return err
		}
	}
	// Preserve prior interrupted attempts and any private findings; only the
	// existing unfinished attempt receives this operational outcome. A queued
	// cancellation creates no synthetic start instant or attempt row.
	if _, err := tx.Exec(ctx, `UPDATE reference_pack_attempts SET completed_at=$2,outcome=$3,canonical_result=$4 WHERE operation_id=$1 AND completed_at IS NULL`, operation, completed, outcome, result); err != nil {
		return err
	}
	if _, err := tx.Exec(ctx, `UPDATE reference_pack_operations SET terminal_at=$2,final_outcome=$3 WHERE operation_id=$1 AND terminal_at IS NULL`, operation, completed, result); err != nil {
		return err
	}
	return appendPackAuditTx(ctx, tx, operation, "verification_completed", outcome, completed, "", "", nil)
}

// Failed candidate health is separate from successful retained health. Clear
// a prior initial-import verdict when the latest explicit import aborts; keep
// all attempt history and never change a candidate with a successful envelope.
func closeUnverifiedCandidatesTx(ctx context.Context, tx pgx.Tx, operation uuid.UUID) error {
	rows, err := tx.Query(ctx, `SELECT DISTINCT pack_key COLLATE "C" FROM reference_pack_operation_members WHERE operation_id=$1 ORDER BY 1`, operation)
	if err != nil {
		return err
	}
	keys := []string{}
	for rows.Next() {
		var key string
		if err := rows.Scan(&key); err != nil {
			rows.Close()
			return err
		}
		keys = append(keys, key)
	}
	err = rows.Err()
	rows.Close()
	if err != nil {
		return err
	}
	for _, key := range keys {
		if _, err := tx.Exec(ctx, `SELECT 1 FROM reference_pack_key_state WHERE pack_key=$1 FOR UPDATE`, key); err != nil {
			return err
		}
		changed, err := tx.Exec(ctx, `UPDATE reference_pack_candidates c SET health='failed',last_failure_code=NULL,missing_reason=NULL
WHERE c.pack_key=$2 AND c.current_envelope_id IS NULL AND NOT c.removed
AND ROW(c.health,c.last_failure_code,c.missing_reason) IS DISTINCT FROM ROW('failed'::text,NULL::text,NULL::text)
AND EXISTS(SELECT 1 FROM reference_pack_operation_members m WHERE m.operation_id=$1 AND m.pack_key=c.pack_key AND m.pack_version=c.pack_version)`, operation, key)
		if err != nil {
			return err
		}
		if changed.RowsAffected() > 0 {
			if _, err := tx.Exec(ctx, `UPDATE reference_pack_key_state SET revision=revision+1 WHERE pack_key=$1`, key); err != nil {
				return err
			}
		}
	}

	return nil
}
