package extensionstore

import (
	"context"
	"encoding/json"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

type JobCommitProof struct {
	JobID                   uuid.UUID
	OwnerProfileID          string
	OperationKind           string
	FinalCommitID           string
	IdempotencyIdentity     json.RawMessage
	NormalizedRequestSHA256 string
	TerminalResult          json.RawMessage
	TerminalResultSHA256    string
	ResourceRefs            json.RawMessage
	AuditCorrelationID      *string
	CommittedAt             time.Time
}

// InsertJobCommitProof is deliberately transaction-scoped: the proof, terminal
// job state, authoritative mutations, audit row, and idempotency result share the
// caller's single final commit boundary.
func InsertJobCommitProof(ctx context.Context, tx pgx.Tx, proof JobCommitProof) error {
	if tx == nil || proof.JobID == uuid.Nil || proof.OwnerProfileID == "" || proof.OperationKind == "" || proof.FinalCommitID == "" || proof.NormalizedRequestSHA256 == "" || proof.TerminalResultSHA256 == "" || len(proof.TerminalResult) == 0 || len(proof.ResourceRefs) == 0 {
		return ErrInvalidTransition
	}
	_, err := tx.Exec(ctx, `
INSERT INTO extension_job_commit_proofs (
    job_id, owner_profile_id, operation_kind, final_commit_id,
    idempotency_identity, normalized_request_sha256, terminal_result,
    terminal_result_sha256, resource_refs, audit_correlation_id, committed_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
`, proof.JobID, proof.OwnerProfileID, proof.OperationKind, proof.FinalCommitID,
		nullableJSON(proof.IdempotencyIdentity), proof.NormalizedRequestSHA256,
		proof.TerminalResult, proof.TerminalResultSHA256, proof.ResourceRefs,
		proof.AuditCorrelationID, proof.CommittedAt.UTC())
	return err
}

func nullableJSON(value json.RawMessage) any {
	if len(value) == 0 || string(value) == "null" {
		return nil
	}
	return value
}

type JobCancellationObservation struct {
	CancellationRequestID     string
	JobID                     uuid.UUID
	ObservedAt                time.Time
	ObservedBeforeFinalCommit bool
}

func (s *Store) RecordJobCancellationObservation(ctx context.Context, observation JobCancellationObservation) error {
	if s == nil || s.pool == nil || observation.CancellationRequestID == "" || observation.JobID == uuid.Nil || !observation.ObservedBeforeFinalCommit {
		return ErrInvalidTransition
	}
	_, err := s.pool.Exec(ctx, `
INSERT INTO extension_job_cancellation_observations (
    cancellation_request_id, job_id, observed_at, observed_before_final_commit
) VALUES ($1, $2, $3, TRUE)
ON CONFLICT (cancellation_request_id) DO NOTHING
`, observation.CancellationRequestID, observation.JobID, observation.ObservedAt.UTC())
	return err
}

func (s *Store) JobCommitProof(ctx context.Context, jobID uuid.UUID) (*JobCommitProof, error) {
	if s == nil || s.pool == nil || jobID == uuid.Nil {
		return nil, ErrInvalidTransition
	}
	var proof JobCommitProof
	err := s.pool.QueryRow(ctx, `
SELECT job_id, owner_profile_id, operation_kind, final_commit_id,
       idempotency_identity, normalized_request_sha256, terminal_result,
       terminal_result_sha256, resource_refs, audit_correlation_id, committed_at
  FROM extension_job_commit_proofs
 WHERE job_id = $1
`, jobID).Scan(&proof.JobID, &proof.OwnerProfileID, &proof.OperationKind,
		&proof.FinalCommitID, &proof.IdempotencyIdentity, &proof.NormalizedRequestSHA256,
		&proof.TerminalResult, &proof.TerminalResultSHA256, &proof.ResourceRefs,
		&proof.AuditCorrelationID, &proof.CommittedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return &proof, nil
}
