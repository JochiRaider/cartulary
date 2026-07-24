package extensionassembly

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/extensions"
	"github.com/JochiRaider/cartulary/internal/platform/extensionstore"
)

type InactiveJobStore struct {
	store *extensionstore.Store
	now   func() time.Time
}

func NewInactiveJobStore(store *extensionstore.Store, now func() time.Time) (*InactiveJobStore, error) {
	if store == nil {
		return nil, errors.New("inactive extension job store is required")
	}
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	return &InactiveJobStore{store: store, now: now}, nil
}

func (s *InactiveJobStore) LoadInactiveJobs(ctx context.Context, profileID string, limit int) ([]extensions.InactiveJob, error) {
	records, err := s.store.LoadInactiveJobRecords(ctx, profileID, limit)
	if err != nil {
		return nil, err
	}
	result := make([]extensions.InactiveJob, 0, len(records))
	for _, record := range records {
		item := extensions.InactiveJob{
			JobID:                   record.JobID.String(),
			OwnerProfileID:          record.OwnerProfileID,
			JobKind:                 record.JobKind,
			SubmittedAt:             record.SubmittedAt,
			IdempotencyIdentity:     append([]byte(nil), record.IdempotencyIdentity...),
			NormalizedRequestSHA256: record.NormalizedRequestSHA256,
		}
		if record.Proof != nil {
			item.Proof = &extensions.InactiveJobProof{
				JobID:                   record.Proof.JobID.String(),
				OwnerProfileID:          record.Proof.OwnerProfileID,
				OperationKind:           record.Proof.OperationKind,
				FinalCommitID:           record.Proof.FinalCommitID,
				IdempotencyIdentity:     append([]byte(nil), record.Proof.IdempotencyIdentity...),
				NormalizedRequestSHA256: record.Proof.NormalizedRequestSHA256,
				TerminalResult:          append([]byte(nil), record.Proof.TerminalResult...),
				TerminalResultSHA256:    record.Proof.TerminalResultSHA256,
				ResourceRefs:            append([]byte(nil), record.Proof.ResourceRefs...),
				AuditCorrelationID:      record.Proof.AuditCorrelationID,
				CommittedAt:             record.Proof.CommittedAt,
			}
		}
		if record.Cancellation != nil {
			item.Cancellation = &extensions.InactiveJobCancellation{
				CancellationRequestID:     record.Cancellation.CancellationRequestID,
				JobID:                     record.Cancellation.JobID.String(),
				ObservedAt:                record.Cancellation.ObservedAt,
				ObservedBeforeFinalCommit: record.Cancellation.ObservedBeforeFinalCommit,
			}
		}
		result = append(result, item)
	}
	return result, nil
}

func (s *InactiveJobStore) ApplyInactiveJobOutcomes(ctx context.Context, profileID string, outcomes []extensions.InactiveJobTerminalOutcome) (extensions.ReconciliationCommitOutcome, error) {
	records := make([]extensionstore.InactiveJobOutcomeRecord, 0, len(outcomes))
	for _, outcome := range outcomes {
		jobID, err := uuid.Parse(outcome.JobID)
		if err != nil {
			return extensions.ReconciliationCommitAbsent, err
		}
		records = append(records, extensionstore.InactiveJobOutcomeRecord{
			JobID:                     jobID,
			SubmittedAt:               outcome.SubmittedAt,
			Status:                    outcome.Status,
			TerminalResult:            append([]byte(nil), outcome.TerminalResult...),
			EvidenceKind:              outcome.EvidenceKind,
			ProofTerminalResultSHA256: outcome.ProofTerminalResultSHA256,
			CancellationRequestID:     outcome.CancellationRequestID,
		})
	}
	commitOutcome, err := s.store.ApplyInactiveJobOutcomeRecords(ctx, profileID, records, s.now().UTC())
	switch commitOutcome {
	case extensionstore.CommitProven:
		return extensions.ReconciliationCommitted, err
	case extensionstore.CommitUnknown:
		return extensions.ReconciliationIndeterminate, err
	default:
		return extensions.ReconciliationCommitAbsent, err
	}
}
