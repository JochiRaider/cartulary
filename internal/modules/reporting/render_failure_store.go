package reporting

import (
	"context"
	"errors"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// The shared finalizer already holds the Jobs execution/terminal lock. Only
// candidate metadata and exact input version/presence are read at this boundary.
func (s *Store) completeRenderFailureTx(ctx context.Context, tx pgx.Tx, id uuid.UUID, candidate renderFailureCandidate, now time.Time) (*uuid.UUID, error) {
	invalid := errors.New("reporting: invalid failure publication candidate")
	if candidate.owner != s || candidate.jobID != id || id == uuid.Nil || len(candidate.raw) == 0 || len(candidate.raw) > maxRenderFailureRecordBytes {
		return nil, invalid
	}
	retained, err := s.jobTransactions.ReadRetainedExtensionJob(ctx, tx, id)
	if err != nil {
		return nil, err
	}
	if retained.Resource.Status != jobs.StatusRunning && retained.Resource.Status != jobs.StatusCancelRequested {
		return nil, jobs.ErrExecutionLost
	}
	if err := validateRenderFailureRecordBinding(candidate.record, retained); err != nil {
		return nil, err
	}
	var version string
	var admissionSHA, preparationSHA *string
	err = tx.QueryRow(ctx, `SELECT xmin::text,render_admission_sha256,render_preparation_sha256 FROM reporting_job_payloads WHERE job_id=$1 FOR UPDATE`, id).Scan(&version, &admissionSHA, &preparationSHA)
	if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return nil, err
	}
	if (err == nil) != candidate.inputPresent || candidate.inputPresent && version != candidate.inputVersion {
		return nil, invalid
	}
	binding := candidate.record.InputBinding
	if binding != nil && (admissionSHA == nil || *admissionSHA != binding.AdmissionSHA256 || !equalOptionalString(preparationSHA, binding.PreparationSHA256)) {
		return nil, invalid
	}
	var outputs int
	if err := tx.QueryRow(ctx, `SELECT (SELECT count(*) FROM reporting_releases WHERE create_job_id=$1)+(SELECT count(*) FROM reporting_composition_preview_outputs WHERE render_attempt_id=$1)`, id).Scan(&outputs); err != nil {
		return nil, err
	}
	if outputs != 0 {
		return nil, errors.New("reporting: unfinished attempt already has owner output")
	}
	tag, err := tx.Exec(ctx, `INSERT INTO reporting_render_failure_records(job_id,record_json,record_sha256) VALUES($1,$2,$3) ON CONFLICT(job_id) DO NOTHING`, id, candidate.raw, candidate.digest)
	if err != nil {
		return nil, err
	}
	if tag.RowsAffected() != 1 {
		return nil, errors.New("reporting: unfinished attempt already has failure evidence")
	}
	if candidate.release == nil {
		return nil, nil
	}
	if binding == nil || binding.PreparationSHA256 == nil || retained.JobKind != ReleaseCreateJobKind {
		return nil, invalid
	}
	releaseID, err := s.completeReleaseRenderFailedJobTx(ctx, tx, id, *candidate.release, candidate.failure.ReasonCode, now)
	if err != nil {
		return nil, err
	}
	return &releaseID, nil
}

func readRenderFailureRecordTx(ctx context.Context, jobReader jobs.RetainedJobReader, tx pgx.Tx, id uuid.UUID) (renderFailureRecord, reportingRenderValidationSummary, error) {
	var raw []byte
	var digest string
	err := tx.QueryRow(ctx, `SELECT record_json,record_sha256 FROM reporting_render_failure_records WHERE job_id=$1`, id).Scan(&raw, &digest)
	if errors.Is(err, pgx.ErrNoRows) {
		return renderFailureRecord{}, reportingRenderValidationSummary{}, ErrNotFound
	}
	if err != nil {
		return renderFailureRecord{}, reportingRenderValidationSummary{}, err
	}
	retained, err := jobReader.ReadRetainedExtensionJob(ctx, tx, id)
	if err != nil {
		return renderFailureRecord{}, reportingRenderValidationSummary{}, err
	}
	return decodeRenderFailureRecord(raw, digest, retained)
}

func validateRetainedRenderFailureTx(ctx context.Context, jobReader jobs.RetainedJobReader, tx pgx.Tx, jobID uuid.UUID, input renderInput, reason *string) error {
	if reason == nil {
		var exists bool
		if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM reporting_render_failure_records WHERE job_id=$1)`, jobID).Scan(&exists); err != nil {
			return err
		}
		if exists {
			return errors.New("reporting: unfinished or successful render has failure diagnostics")
		}
		return nil
	}
	record, summary, err := readRenderFailureRecordTx(ctx, jobReader, tx, jobID)
	if err != nil {
		return err
	}
	if summary.ReasonCode == nil || *summary.ReasonCode != *reason {
		return errors.New("reporting: retained failure reason differs from its outcome")
	}
	if record.InputBinding == nil {
		return nil
	}
	var admissionSHA string
	var preparationSHA *string
	if err := tx.QueryRow(ctx, `SELECT render_admission_sha256,render_preparation_sha256 FROM reporting_job_payloads WHERE job_id=$1`, jobID).Scan(&admissionSHA, &preparationSHA); err != nil {
		return err
	}
	if admissionSHA != record.InputBinding.AdmissionSHA256 || !equalOptionalString(preparationSHA, record.InputBinding.PreparationSHA256) {
		return errors.New("reporting: failure record input proof changed")
	}
	return validateFailedRenderSummary(record.Summary, input, *reason)
}

func validateRetainedPreviewFailureTx(ctx context.Context, jobReader jobs.RetainedJobReader, tx pgx.Tx, jobID uuid.UUID, input renderInput, retained jobs.RetainedExtensionJob) error {
	resource := retained.Resource
	if resource.Status != jobs.StatusFailed && resource.Status != jobs.StatusCanceled {
		return validateRetainedRenderFailureTx(ctx, jobReader, tx, jobID, input, nil)
	}
	_, summary, err := readRenderFailureRecordTx(ctx, jobReader, tx, jobID)
	if err != nil || summary.ReasonCode == nil {
		return errors.New("reporting: aborted preview has no failure diagnostic")
	}
	if err := validateRetainedRenderFailureTx(ctx, jobReader, tx, jobID, input, summary.ReasonCode); err != nil {
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
