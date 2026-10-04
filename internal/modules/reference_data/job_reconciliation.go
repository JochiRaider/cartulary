package reference_data

import (
	"context"
	"errors"

	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type TerminalJobReader interface {
	ReadTerminalDispositionTx(context.Context, pgx.Tx, uuid.UUID) (jobs.TerminalDisposition, bool, error)
}

// ReconcileTerminalJobs runs before readiness, including with the profile
// unclaimed. It repairs only terminal lifecycle facts proved by Jobs; queued or
// recoverable executions retain their exact frozen input and preparation.
func ReconcileTerminalJobs(ctx context.Context, pool *pgxpool.Pool, reader TerminalJobReader, finalizer ActionFinalizer) error {
	if pool == nil || reader == nil || finalizer == nil {
		return errors.New("reference pack: incomplete terminal reconciliation dependencies")
	}
	cursor := uuid.Nil
	for {
		rows, err := pool.Query(ctx, `SELECT operation_id,job_id FROM reference_pack_operations WHERE terminal_at IS NULL AND job_id IS NOT NULL AND operation_id>$1 ORDER BY operation_id LIMIT 64`, cursor)
		if err != nil {
			return err
		}
		type item struct{ operation, job uuid.UUID }
		batch := []item{}
		for rows.Next() {
			var v item
			if err := rows.Scan(&v.operation, &v.job); err != nil {
				rows.Close()
				return err
			}
			batch = append(batch, v)
		}
		err = rows.Err()
		rows.Close()
		if err != nil {
			return err
		}
		if len(batch) == 0 {
			return nil
		}
		for _, v := range batch {
			if err := reconcileTerminalJob(ctx, pool, reader, finalizer, v.operation, v.job); err != nil {
				return err
			}
			cursor = v.operation
		}
	}
}
func reconcileTerminalJob(ctx context.Context, pool *pgxpool.Pool, reader TerminalJobReader, finalizer ActionFinalizer, operation, job uuid.UUID) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	disposition, terminal, err := reader.ReadTerminalDispositionTx(ctx, tx, job)
	if err != nil {
		return err
	}
	if !terminal {
		return nil
	}
	if err := (JobTerminalEffects{}).ApplyJobTerminalEffectsTx(ctx, tx, disposition); err != nil {
		return err
	}
	// The exact immutable owner receipt, read while the terminal Job lock is
	// held, is sufficient to recognize this transaction after lost acknowledgement.
	var result []byte
	if err := tx.QueryRow(ctx, `SELECT final_outcome FROM reference_pack_operations WHERE operation_id=$1 AND job_id=$2 AND terminal_at IS NOT NULL`, operation, job).Scan(&result); err != nil {
		return err
	}
	return finalizer.FinalizeReferencePackAction(ctx, tx, func(proofCtx context.Context) (bool, error) {
		var proven bool
		err := pool.QueryRow(proofCtx, `SELECT EXISTS(SELECT 1 FROM reference_pack_operations WHERE operation_id=$1 AND job_id=$2 AND terminal_at IS NOT NULL AND final_outcome=$3)`, operation, job, result).Scan(&proven)
		return proven, err
	})
}
