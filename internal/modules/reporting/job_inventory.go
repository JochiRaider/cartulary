package reporting

import (
	"context"
	"errors"

	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// ValidateRetainedJobInventory reconciles independently owned Jobs with private
// input and terminal evidence. Scanning private rows alone cannot see a missing
// input, and a truthful integrity diagnostic does not repair that missing state.
func ValidateRetainedJobInventory(ctx context.Context, jobReader jobs.RetainedJobReader, pool *pgxpool.Pool) error {
	if jobReader == nil || pool == nil {
		return errors.New("reporting: job inventory requires storage")
	}
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return err
	}
	defer rollbackSnapshotTx(tx)
	var after uuid.UUID
	var inputs, failures int64
	for {
		ids, err := jobs.ReadRetainedExtensionJobPage(ctx, tx, ProfileID, SnapshotCreateJobKind, JobWorkerKind, after)
		if err != nil {
			return err
		}
		if len(ids) == 0 {
			break
		}
		for _, id := range ids {
			retained, err := jobReader.ReadRetainedExtensionJob(ctx, tx, id)
			if err != nil {
				return err
			}
			if retained.OwnerProfileID != ProfileID || retained.HandlerName != JobWorkerKind {
				return errors.New("reporting: retained Job registration differs from its owner")
			}
			var want string
			switch retained.JobKind {
			case SnapshotCreateJobKind:
				want = reportingJobKindSnapshotCreate
			case ReleaseCreateJobKind:
				want = reportingJobKindReleaseCreate
			case CompositionPreviewJobKind:
				want = reportingJobKindCompositionPreview
			default:
				return errors.New("reporting: unsupported retained Job kind")
			}
			var kind string
			if err := tx.QueryRow(ctx, `SELECT job_kind FROM reporting_job_payloads WHERE job_id=$1`, id).Scan(&kind); err != nil {
				return err
			}
			if kind != want {
				return errors.New("reporting: private input differs from Jobs dispatch")
			}
			inputs++
			terminalFailure := retained.Resource.Status == jobs.StatusFailed || retained.Resource.Status == jobs.StatusCanceled
			if want != reportingJobKindSnapshotCreate && terminalFailure {
				if _, _, err := readRenderFailureRecordTx(ctx, jobReader, tx, id); err != nil {
					return err
				}
				failures++
			} else {
				if err := validateRetainedRenderFailureTx(ctx, jobReader, tx, id, renderInput{}, nil); err != nil {
					return err
				}
			}
			after = id
		}
	}
	var storedInputs, storedFailures int64
	if err := tx.QueryRow(ctx, `SELECT (SELECT count(*) FROM reporting_job_payloads),(SELECT count(*) FROM reporting_render_failure_records)`).Scan(&storedInputs, &storedFailures); err != nil {
		return err
	}
	if inputs != storedInputs || failures != storedFailures {
		return errors.New("reporting: orphaned private input or failure evidence")
	}
	return tx.Commit(ctx)
}
