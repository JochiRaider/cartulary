package reporting

import (
	"context"
	"errors"
	"time"

	sqlc "github.com/JochiRaider/cartulary/internal/gen/sql"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// SnapshotCandidate is a short-lived proof of a fully validated retained
// capture. Only bounded metadata crosses the final Jobs transition lock.
// The physical tuple identity is an optimistic concurrency fence, never a
// durable artifact identity or provenance claim.
type SnapshotCandidate struct {
	jobID      uuid.UUID
	xmin, ctid string
	metadata   snapshotPublicationMetadata
}

// Publication metadata belongs to Reporting. Generated query types are confined
// to the persistence call and cannot become part of the candidate's API shape.
type snapshotPublicationMetadata struct {
	snapshotID, incidentID, actorID uuid.UUID
	clientTxnID                     string
	snapshotAt                      time.Time
	highWatermark                   string
	sourceBoundary                  []byte
	captureSHA256                   string
}

func (s *Store) SnapshotCandidateForJob(ctx context.Context, execution jobs.Execution) (SnapshotCandidate, error) {
	return ValidateSnapshotCandidate(ctx, s.pool, execution.JobID())
}

// ValidateSnapshotCandidate validates case-content bytes in a consistent,
// read-only transaction before publication. It neither holds the Jobs lock
// nor mutates retained history. A subsequent payload update invalidates it,
// even when that update restores the same bytes or keeps embedded hashes.
func ValidateSnapshotCandidate(ctx context.Context, pool *pgxpool.Pool, jobID uuid.UUID) (SnapshotCandidate, error) {
	var candidate SnapshotCandidate
	if pool == nil || jobID == uuid.Nil {
		return candidate, errors.New("reporting: snapshot validation requires storage and identity")
	}
	tx, err := pool.BeginTx(ctx, pgx.TxOptions{IsoLevel: pgx.RepeatableRead, AccessMode: pgx.ReadOnly})
	if err != nil {
		return candidate, err
	}
	defer rollbackSnapshotTx(tx)
	payload, err := getSnapshotPayloadTx(ctx, tx, jobID)
	if err != nil {
		return candidate, err
	}
	if err := tx.QueryRow(ctx, `SELECT xmin::text,ctid::text FROM reporting_job_payloads WHERE job_id=$1`, jobID).Scan(&candidate.xmin, &candidate.ctid); err != nil {
		return candidate, err
	}
	candidate.jobID = jobID
	candidate.metadata = snapshotPublicationMetadata{
		snapshotID: reportingResourceID(jobID, "snapshot"), incidentID: uuid.MustParse(payload.IncidentID), actorID: uuid.MustParse(payload.ActorUserID),
		clientTxnID: payload.ClientTxnID, snapshotAt: payload.SnapshotAt, highWatermark: payload.SourceChangeSetHighWatermark,
		sourceBoundary: append([]byte{}, payload.SourceBoundaryJSON...), captureSHA256: payload.SourceCaptureSHA256,
	}

	if err := ctx.Err(); err != nil {
		return SnapshotCandidate{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return SnapshotCandidate{}, err
	}
	return candidate, nil
}

func (s *Store) CompleteSnapshotCreateJobTx(ctx context.Context, tx pgx.Tx, jobID, snapshotID uuid.UUID, candidate SnapshotCandidate) error {
	if jobID == uuid.Nil || candidate.jobID != jobID || snapshotID != reportingResourceID(jobID, "snapshot") || candidate.xmin == "" || candidate.ctid == "" {
		return errors.New("reporting: invalid snapshot publication candidate")
	}
	var xmin, ctid string
	if err := tx.QueryRow(ctx, `SELECT xmin::text,ctid::text FROM reporting_job_payloads WHERE job_id=$1 FOR UPDATE`, jobID).Scan(&xmin, &ctid); err != nil {
		return err
	}
	if xmin != candidate.xmin || ctid != candidate.ctid {
		return errors.New("reporting: snapshot capture changed after validation")
	}
	// Recheck the small independent admission and Jobs identity at publication.
	if _, err := readSnapshotAdmissionTx(ctx, tx, jobID); err != nil {
		return err
	}
	metadata := candidate.metadata
	m := sqlc.CreateReportingSnapshotParams{
		SnapshotID: pgUUID(metadata.snapshotID), IncidentID: pgUUID(metadata.incidentID), CreatedByUserID: pgUUID(metadata.actorID),
		ClientTxnID: metadata.clientTxnID, SnapshotAt: pgTimestamptz(metadata.snapshotAt), SourceChangeSetHighWatermark: metadata.highWatermark,
		SourceBoundaryJson: metadata.sourceBoundary, DerivationVersion: DerivationVersion, SourceCaptureSha256: metadata.captureSHA256, CreateJobID: pgUUID(jobID),
	}
	var existing sqlc.CreateReportingSnapshotParams
	err := tx.QueryRow(ctx, `SELECT snapshot_id,incident_id,created_by_user_id,client_txn_id,snapshot_at,source_change_set_high_watermark,source_boundary_json,derivation_version,source_capture_sha256,create_job_id FROM reporting_snapshots WHERE create_job_id=$1`, jobID).Scan(&existing.SnapshotID, &existing.IncidentID, &existing.CreatedByUserID, &existing.ClientTxnID, &existing.SnapshotAt, &existing.SourceChangeSetHighWatermark, &existing.SourceBoundaryJson, &existing.DerivationVersion, &existing.SourceCaptureSha256, &existing.CreateJobID)
	if err == nil {
		if existing.SnapshotID != m.SnapshotID || existing.IncidentID != m.IncidentID || existing.CreatedByUserID != m.CreatedByUserID || existing.ClientTxnID != m.ClientTxnID || !existing.SnapshotAt.Time.Equal(m.SnapshotAt.Time) || existing.SourceChangeSetHighWatermark != m.SourceChangeSetHighWatermark || !equalFrozenJSON(existing.SourceBoundaryJson, m.SourceBoundaryJson) || existing.DerivationVersion != m.DerivationVersion || existing.SourceCaptureSha256 != m.SourceCaptureSha256 || existing.CreateJobID != m.CreateJobID {
			return errors.New("reporting: retained snapshot differs from validated capture")
		}
		return nil
	}
	if !errors.Is(err, pgx.ErrNoRows) {
		return err
	}
	row, err := sqlc.New(tx).CreateReportingSnapshot(ctx, m)
	if err != nil {
		return err
	}
	if row.Bytes != snapshotID {
		return errors.New("reporting: snapshot publication identity mismatch")
	}
	return nil
}
