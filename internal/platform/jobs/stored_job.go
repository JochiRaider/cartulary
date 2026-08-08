package jobs

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

// storedJob is the private durable representation. Definition identity remains
// unavailable to public Resource projections.
type storedJob struct {
	resource       Resource
	jobKind        string
	progressUnitID string
}

func (job storedJob) publicResource() Resource {
	return job.resource
}

func getJobTx(ctx context.Context, tx pgx.Tx, jobID uuid.UUID) (Resource, error) {
	stored, err := scanStoredJob(tx.QueryRow(ctx, `
SELECT job_id, scope_kind, incident_id, status, cancelable, submitted_by_user_id,
       auth_policy,
       submitted_at, updated_at, progress_completed, progress_total, started_at,
       finished_at, retained_until, result_summary_json, error_summary_json, message,
       job_kind, progress_unit_id
  FROM jobs
 WHERE job_id = $1
 FOR UPDATE
`, jobID))
	if errors.Is(err, pgx.ErrNoRows) {
		return Resource{}, ErrNotFound
	}
	return stored.publicResource(), err
}

// requireVisibleJobTx establishes the public retention boundary before any
// replay or mutation lookup. Taking the row lock makes the visibility decision
// stable for the remainder of the caller's transaction.
func requireVisibleJobTx(ctx context.Context, tx pgx.Tx, jobID uuid.UUID, now time.Time) error {
	var present bool
	err := tx.QueryRow(ctx, `
SELECT true
  FROM jobs
 WHERE job_id = $1
   AND (retained_until IS NULL OR retained_until > $2)
 FOR UPDATE
`, jobID, now.UTC()).Scan(&present)
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	return err
}

func scanStoredJob(row pgx.Row) (storedJob, error) {
	var stored storedJob
	var jobID uuid.UUID
	var submittedBy uuid.UUID
	var incidentID *uuid.UUID
	var progressTotal *int
	var resultJSON []byte
	var errorJSON []byte
	var jobKind *string
	var progressUnitID *string
	record := &stored.resource
	if err := row.Scan(
		&jobID,
		&record.Scope.Kind,
		&incidentID,
		&record.Status,
		&record.Cancelable,
		&submittedBy,
		&record.AuthPolicy,
		&record.SubmittedAt,
		&record.UpdatedAt,
		&record.Progress.Completed,
		&progressTotal,
		&record.StartedAt,
		&record.FinishedAt,
		&record.RetainedUntil,
		&resultJSON,
		&errorJSON,
		&record.Message,
		&jobKind,
		&progressUnitID,
	); err != nil {
		return storedJob{}, err
	}
	if jobKind != nil {
		stored.jobKind = *jobKind
	}
	if progressUnitID != nil {
		stored.progressUnitID = *progressUnitID
	}
	if err := hydrateJobResource(record, jobID, submittedBy, incidentID, progressTotal, resultJSON, errorJSON); err != nil {
		return storedJob{}, err
	}
	return stored, nil
}

func scanJob(row pgx.Row) (Resource, error) {
	var record Resource
	var jobID uuid.UUID
	var submittedBy uuid.UUID
	var incidentID *uuid.UUID
	var progressTotal *int
	var resultJSON []byte
	var errorJSON []byte
	if err := row.Scan(
		&jobID,
		&record.Scope.Kind,
		&incidentID,
		&record.Status,
		&record.Cancelable,
		&submittedBy,
		&record.AuthPolicy,
		&record.SubmittedAt,
		&record.UpdatedAt,
		&record.Progress.Completed,
		&progressTotal,
		&record.StartedAt,
		&record.FinishedAt,
		&record.RetainedUntil,
		&resultJSON,
		&errorJSON,
		&record.Message,
	); err != nil {
		return Resource{}, err
	}
	if err := hydrateJobResource(&record, jobID, submittedBy, incidentID, progressTotal, resultJSON, errorJSON); err != nil {
		return Resource{}, err
	}
	return (storedJob{resource: record}).publicResource(), nil
}

func hydrateJobResource(record *Resource, jobID uuid.UUID, submittedBy uuid.UUID, incidentID *uuid.UUID, progressTotal *int, resultJSON []byte, errorJSON []byte) error {
	record.JobID = jobID.String()
	record.Scope.IncidentID = incidentID
	record.StatusRoute = "/api/v1/jobs/" + record.JobID
	record.SubmittedByUserID = submittedBy.String()
	record.Progress.Total = progressTotal
	if len(resultJSON) > 0 {
		var summary ResultSummary
		if err := json.Unmarshal(resultJSON, &summary); err != nil {
			return err
		}
		record.ResultSummary = &summary
	}
	if len(errorJSON) > 0 {
		var summary ErrorSummary
		if err := json.Unmarshal(errorJSON, &summary); err != nil {
			return err
		}
		record.ErrorSummary = &summary
	}
	return nil
}
