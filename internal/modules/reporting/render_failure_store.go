package reporting

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// CompleteRenderFailureJobTx records owner state under the finalizer's Jobs
// lock. It reads frozen admission without acquiring new graph leases or running
// a participant, so cancellation before execution can retain truthful evidence.
func (s *Store) CompleteRenderFailureJobTx(ctx context.Context, tx pgx.Tx, jobID uuid.UUID, failure RenderFailure, now time.Time) (*uuid.UUID, error) {
	var kind string
	if err := tx.QueryRow(ctx, `SELECT job_kind FROM reporting_job_payloads WHERE job_id=$1`, jobID).Scan(&kind); err != nil {
		return nil, err
	}
	if kind == reportingJobKindSnapshotCreate {
		return nil, nil
	}
	a, digest, err := readRenderAdmissionTx(ctx, tx, jobID)
	if err != nil {
		return nil, err
	}
	seal, _, err := readRenderPreparationTx(ctx, tx, jobID, a, digest)
	if err != nil {
		return nil, err
	}
	input, err := renderAdmissionMetadata(a, seal)
	if err != nil {
		return nil, err
	}
	switch kind {
	case reportingJobKindCompositionPreview:
		var hasOutput bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM reporting_composition_preview_outputs WHERE render_attempt_id=$1)`, jobID).Scan(&hasOutput); err != nil {
			return nil, err
		}
		if hasOutput {
			return nil, errors.New("reporting: failed preview already has output")
		}
		return nil, retainRenderFailureTx(ctx, tx, jobID, input, failure)
	case reportingJobKindReleaseCreate:
		if seal == nil {
			if _, err := getReleaseRecordByCreateJobIDTx(ctx, tx, jobID); !errors.Is(err, ErrNotFound) {
				return nil, errors.New("reporting: unprepared render has a release outcome")
			}
			return nil, retainRenderFailureTx(ctx, tx, jobID, input, failure)
		}
		payload, err := releasePayloadFromPrepared(a, input)
		if err != nil {
			return nil, err
		}
		retained, err := jobs.ReadRetainedExtensionJob(ctx, tx, jobID)
		if err != nil {
			return nil, err
		}
		if retained.Resource.Status == jobs.StatusFailed || retained.Resource.Status == jobs.StatusCanceled {
			if _, err := getReleaseRecordByCreateJobIDTx(ctx, tx, jobID); err != nil {
				return nil, errors.New("reporting: terminal prepared render has no release outcome")
			}
		}
		if err := retainRenderFailureTx(ctx, tx, jobID, input, failure); err != nil {
			return nil, err
		}
		id, err := s.completeReleaseRenderFailedJobTx(ctx, tx, jobID, payload, failure.ReasonCode, now)
		if err != nil {
			return nil, err
		}
		return &id, nil
	default:
		return nil, errors.New("reporting: unsupported failed attempt kind")
	}
}

func retainRenderFailureTx(ctx context.Context, tx pgx.Tx, jobID uuid.UUID, input renderInput, failure RenderFailure) error {
	retained, err := jobs.ReadRetainedExtensionJob(ctx, tx, jobID)
	if err != nil {
		return err
	}
	var reason *string
	switch retained.Resource.Status {
	case jobs.StatusFailed, jobs.StatusCanceled:
		// A terminal replay validates retained bytes before writing. Missing
		// diagnostics are corruption, never permission to reconstruct history.
		reason = &failure.ReasonCode
	case jobs.StatusRunning, jobs.StatusCancelRequested:
	case jobs.StatusSucceeded:
		return errors.New("reporting: successful render cannot acquire failure diagnostics")
	default:
		return errors.New("reporting: failure requires a claimed render attempt")
	}
	if err := validateRetainedRenderFailureTx(ctx, tx, jobID, input, reason); err != nil {
		return err
	}
	raw, err := failedRenderSummary(input, failure)
	if err != nil {
		return err
	}
	digest := reportingObjectDigest("cartulary.reporting_render_validation_summary.v1", raw)
	tag, err := tx.Exec(ctx, `UPDATE reporting_job_payloads SET render_failure_summary=$2,render_failure_sha256=$3 WHERE job_id=$1 AND (render_failure_summary IS NULL OR (render_failure_summary=$2 AND render_failure_sha256=$3))`, jobID, raw, digest)
	if err != nil {
		return err
	}
	if tag.RowsAffected() != 1 {
		return errors.New("reporting: failure replay changed retained diagnostics")
	}
	return nil
}

func validateRetainedRenderFailureTx(ctx context.Context, tx pgx.Tx, jobID uuid.UUID, input renderInput, reason *string) error {
	var raw []byte
	var digest *string
	if err := tx.QueryRow(ctx, `SELECT render_failure_summary,render_failure_sha256 FROM reporting_job_payloads WHERE job_id=$1`, jobID).Scan(&raw, &digest); err != nil {
		return err
	}
	if reason == nil {
		if len(raw) != 0 || digest != nil {
			return errors.New("reporting: unfinished or successful render has failure diagnostics")
		}
		return nil
	}
	if digest == nil || *digest != reportingObjectDigest("cartulary.reporting_render_validation_summary.v1", raw) {
		return errors.New("reporting: invalid retained render failure digest")
	}
	return validateFailedRenderSummary(raw, input, *reason)
}

func validateRetainedPreviewFailureTx(ctx context.Context, tx pgx.Tx, jobID uuid.UUID, input renderInput, retained jobs.RetainedExtensionJob) error {
	resource := retained.Resource
	if resource.Status != jobs.StatusFailed && resource.Status != jobs.StatusCanceled {
		return validateRetainedRenderFailureTx(ctx, tx, jobID, input, nil)
	}
	var raw []byte
	if err := tx.QueryRow(ctx, `SELECT render_failure_summary FROM reporting_job_payloads WHERE job_id=$1`, jobID).Scan(&raw); err != nil {
		return err
	}
	var summary reportingRenderValidationSummary
	if json.Unmarshal(raw, &summary) != nil || summary.ReasonCode == nil {
		return errors.New("reporting: aborted preview has no failure diagnostic")
	}
	if err := validateRetainedRenderFailureTx(ctx, tx, jobID, input, summary.ReasonCode); err != nil {
		return err
	}
	if resource.Status == jobs.StatusCanceled {
		if *summary.ReasonCode != "render_canceled" && *summary.ReasonCode != "partial_output_cleanup_failed" {
			return errors.New("reporting: canceled preview has another failure outcome")
		}
		if retained.ExpiredAt == nil && (resource.ResultSummary == nil || resource.ResultSummary.Code != "job_canceled") {
			return errors.New("reporting: preview cancellation differs from Jobs outcome")
		}
	} else if retained.ExpiredAt == nil && (resource.ErrorSummary == nil || resource.ErrorSummary.Code != renderFailureJobCode(*summary.ReasonCode) || resource.ErrorSummary.Details["reason_code"] != *summary.ReasonCode) {
		return errors.New("reporting: preview failure differs from Jobs outcome")
	}
	return nil
}
