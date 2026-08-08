package extensionstore

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
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

type RouteCancellationIdentity struct {
	RouteKey    string
	ActorUserID uuid.UUID
	ScopeKey    string
	ClientTxnID string
}

// AppendRouteJobCancellationObservationTx derives the Extensions-owned stable
// request identity and appends the observation in the caller transaction.
func AppendRouteJobCancellationObservationTx(ctx context.Context, executor JobCancellationObservationExecutor, identity RouteCancellationIdentity, jobID uuid.UUID, observedAt time.Time) error {
	if identity.RouteKey == "" || identity.ActorUserID == uuid.Nil || identity.ScopeKey == "" || identity.ClientTxnID == "" {
		return ErrInvalidTransition
	}
	canonical := identity.RouteKey + "\x00" + identity.ActorUserID.String() + "\x00" +
		identity.ScopeKey + "\x00" + identity.ClientTxnID
	digest := sha256.Sum256([]byte(canonical))
	return AppendJobCancellationObservationTx(ctx, executor, JobCancellationObservation{
		CancellationRequestID:     fmt.Sprintf("cancel:%x", digest[:]),
		JobID:                     jobID,
		ObservedAt:                observedAt,
		ObservedBeforeFinalCommit: true,
	})
}

func (s *Store) RecordJobCancellationObservation(ctx context.Context, observation JobCancellationObservation) error {
	if s == nil || s.pool == nil || observation.CancellationRequestID == "" || observation.JobID == uuid.Nil || !observation.ObservedBeforeFinalCommit {
		return ErrInvalidTransition
	}
	return AppendJobCancellationObservationTx(ctx, s.pool, observation)
}

// JobCancellationObservationExecutor is satisfied by pgx.Tx and pgxpool.Pool
// so the Extensions-owned append can join a caller transaction.
type JobCancellationObservationExecutor interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
}

func AppendJobCancellationObservationTx(ctx context.Context, executor JobCancellationObservationExecutor, observation JobCancellationObservation) error {
	if executor == nil || observation.CancellationRequestID == "" || observation.JobID == uuid.Nil || !observation.ObservedBeforeFinalCommit {
		return ErrInvalidTransition
	}
	_, err := executor.Exec(ctx, `
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
