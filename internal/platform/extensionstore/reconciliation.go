package extensionstore

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

type InactiveJobProofRecord struct {
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

type InactiveJobCancellationRecord struct {
	CancellationRequestID     string
	JobID                     uuid.UUID
	ObservedAt                time.Time
	ObservedBeforeFinalCommit bool
}

type InactiveJobRecord struct {
	JobID                   uuid.UUID
	OwnerProfileID          string
	JobKind                 string
	SubmittedAt             time.Time
	IdempotencyIdentity     json.RawMessage
	NormalizedRequestSHA256 string
	Proof                   *InactiveJobProofRecord
	Cancellation            *InactiveJobCancellationRecord
}

type InactiveJobOutcomeRecord struct {
	JobID                     uuid.UUID
	SubmittedAt               time.Time
	Status                    string
	TerminalResult            json.RawMessage
	EvidenceKind              string
	ProofTerminalResultSHA256 string
	CancellationRequestID     string
}

type InactiveJobTerminalWriter interface {
	ValidateInactiveJobTx(context.Context, pgx.Tx, uuid.UUID, string, time.Time) (jobs.InactiveJobGrant, error)
	CompleteInactiveJobTx(context.Context, pgx.Tx, jobs.InactiveJobGrant, jobs.InactiveTerminalOutcome, time.Time) error
}

func (s *Store) LoadInactiveJobRecords(ctx context.Context, profileID string, limit int) ([]InactiveJobRecord, error) {
	if s == nil || s.pool == nil || profileID == "" || limit < 1 {
		return nil, ErrInvalidTransition
	}
	rows, err := s.pool.Query(ctx, `
SELECT j.job_id, j.extension_owner_profile_id, j.job_kind,
       j.submitted_at, j.extension_idempotency_identity,
       j.extension_normalized_request_sha256,
       p.job_id, p.owner_profile_id, p.operation_kind, p.final_commit_id,
       p.idempotency_identity, p.normalized_request_sha256,
       p.terminal_result, p.terminal_result_sha256, p.resource_refs,
       p.audit_correlation_id, p.committed_at,
       c.cancellation_request_id, c.job_id, c.observed_at,
       c.observed_before_final_commit
  FROM jobs j
  LEFT JOIN extension_job_commit_proofs p ON p.job_id = j.job_id
  LEFT JOIN extension_job_cancellation_observations c ON c.job_id = j.job_id
 WHERE j.extension_owner_profile_id = $1
   AND j.status IN ('queued', 'running', 'cancel_requested')
 ORDER BY j.submitted_at ASC, j.job_id ASC
 LIMIT $2
`, profileID, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	result := make([]InactiveJobRecord, 0)
	for rows.Next() {
		var record InactiveJobRecord
		var proofJobID *uuid.UUID
		var proofOwner, proofOperation, finalCommitID *string
		var proofIdentity, proofTerminal, proofRefs []byte
		var proofRequestDigest, proofTerminalDigest *string
		var proofAudit *string
		var proofCommittedAt *time.Time
		var cancellationRequestID *string
		var cancellationJobID *uuid.UUID
		var cancellationObservedAt *time.Time
		var cancellationPrecommit *bool
		if err := rows.Scan(
			&record.JobID, &record.OwnerProfileID, &record.JobKind,
			&record.SubmittedAt, &record.IdempotencyIdentity,
			&record.NormalizedRequestSHA256,
			&proofJobID, &proofOwner, &proofOperation, &finalCommitID,
			&proofIdentity, &proofRequestDigest, &proofTerminal,
			&proofTerminalDigest, &proofRefs, &proofAudit, &proofCommittedAt,
			&cancellationRequestID, &cancellationJobID,
			&cancellationObservedAt, &cancellationPrecommit,
		); err != nil {
			return nil, err
		}
		if proofJobID != nil {
			record.Proof = &InactiveJobProofRecord{
				JobID:                   *proofJobID,
				OwnerProfileID:          stringValuePointer(proofOwner),
				OperationKind:           stringValuePointer(proofOperation),
				FinalCommitID:           stringValuePointer(finalCommitID),
				IdempotencyIdentity:     append(json.RawMessage(nil), proofIdentity...),
				NormalizedRequestSHA256: stringValuePointer(proofRequestDigest),
				TerminalResult:          append(json.RawMessage(nil), proofTerminal...),
				TerminalResultSHA256:    stringValuePointer(proofTerminalDigest),
				ResourceRefs:            append(json.RawMessage(nil), proofRefs...),
				AuditCorrelationID:      proofAudit,
				CommittedAt:             timeValuePointer(proofCommittedAt),
			}
		}
		if cancellationJobID != nil {
			record.Cancellation = &InactiveJobCancellationRecord{
				CancellationRequestID:     stringValuePointer(cancellationRequestID),
				JobID:                     *cancellationJobID,
				ObservedAt:                timeValuePointer(cancellationObservedAt),
				ObservedBeforeFinalCommit: boolValuePointer(cancellationPrecommit),
			}
		}
		result = append(result, record)
	}
	return result, rows.Err()
}

func (s *Store) ApplyInactiveJobOutcomeRecords(ctx context.Context, terminalWriter InactiveJobTerminalWriter, profileID string, outcomes []InactiveJobOutcomeRecord, now time.Time) (CommitOutcome, error) {
	if s == nil || s.pool == nil || terminalWriter == nil || profileID == "" || now.IsZero() {
		return CommitAbsent, ErrInvalidTransition
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return CommitAbsent, err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	var previousSubmittedAt time.Time
	var previousJobID uuid.UUID
	for _, outcome := range outcomes {
		if outcome.JobID == uuid.Nil || outcome.SubmittedAt.IsZero() ||
			(!previousSubmittedAt.IsZero() &&
				(outcome.SubmittedAt.Before(previousSubmittedAt) ||
					(outcome.SubmittedAt.Equal(previousSubmittedAt) && outcome.JobID.String() <= previousJobID.String()))) {
			return CommitAbsent, ErrInvalidTransition
		}
		var proofDigest *string
		var cancellationRequestID *string
		grant, err := terminalWriter.ValidateInactiveJobTx(
			ctx, tx, outcome.JobID, profileID, outcome.SubmittedAt,
		)
		if err != nil {
			return CommitAbsent, err
		}
		if err := tx.QueryRow(ctx, `
SELECT (SELECT p.terminal_result_sha256
          FROM extension_job_commit_proofs p
         WHERE p.job_id = $1),
       (SELECT c.cancellation_request_id
          FROM extension_job_cancellation_observations c
         WHERE c.job_id = $1)
`, outcome.JobID).Scan(&proofDigest, &cancellationRequestID); err != nil {
			return CommitAbsent, err
		}
		if !inactiveEvidenceMatches(outcome, proofDigest, cancellationRequestID) {
			return CommitAbsent, ErrIntegrity
		}
		terminalOutcome, err := inactiveTerminalOutcome(outcome)
		if err != nil {
			return CommitAbsent, err
		}
		if err := terminalWriter.CompleteInactiveJobTx(ctx, tx, grant, terminalOutcome, now.UTC()); err != nil {
			return CommitAbsent, err
		}
		previousSubmittedAt = outcome.SubmittedAt
		previousJobID = outcome.JobID
	}
	if err := tx.Commit(ctx); err != nil {
		return CommitUnknown, err
	}
	return CommitProven, nil
}

func inactiveTerminalOutcome(outcome InactiveJobOutcomeRecord) (jobs.InactiveTerminalOutcome, error) {
	switch outcome.Status {
	case jobs.StatusSucceeded:
		var summary jobs.ResultSummary
		if err := json.Unmarshal(outcome.TerminalResult, &summary); err != nil {
			return jobs.InactiveTerminalOutcome{}, ErrInvalidTransition
		}
		return jobs.NewInactiveSuccessOutcome(summary), nil
	case jobs.StatusFailed:
		var summary jobs.ErrorSummary
		if err := json.Unmarshal(outcome.TerminalResult, &summary); err != nil {
			return jobs.InactiveTerminalOutcome{}, ErrInvalidTransition
		}
		return jobs.NewInactiveFailureOutcome(summary), nil
	case jobs.StatusCanceled:
		return jobs.NewInactiveCancellationOutcome(), nil
	default:
		return jobs.InactiveTerminalOutcome{}, ErrInvalidTransition
	}
}

func inactiveEvidenceMatches(outcome InactiveJobOutcomeRecord, proofDigest *string, cancellationRequestID *string) bool {
	switch outcome.EvidenceKind {
	case "proof":
		return outcome.Status == "succeeded" &&
			proofDigest != nil &&
			*proofDigest == outcome.ProofTerminalResultSHA256
	case "cancellation":
		return outcome.Status == "canceled" &&
			proofDigest == nil &&
			cancellationRequestID != nil &&
			*cancellationRequestID == outcome.CancellationRequestID
	case "absence":
		return outcome.Status == "failed" &&
			proofDigest == nil &&
			cancellationRequestID == nil
	default:
		return false
	}
}

func stringValuePointer(value *string) string {
	if value == nil {
		return ""
	}
	return *value
}

func timeValuePointer(value *time.Time) time.Time {
	if value == nil {
		return time.Time{}
	}
	return *value
}

func boolValuePointer(value *bool) bool {
	return value != nil && *value
}
