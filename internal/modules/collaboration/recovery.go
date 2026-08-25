package collaboration

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"

	privaterecovery "github.com/JochiRaider/cartulary/internal/modules/collaboration/internal/recovery"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type RequeueFailureKind string

const (
	RequeueFailureIncidentNotQuarantined RequeueFailureKind = "incident_not_quarantined"
	RequeueFailureRepairNotVerified      RequeueFailureKind = "repair_not_verified"
	RequeueFailureTransaction            RequeueFailureKind = "transaction_failed"
	RequeueFailureCommitOutcomeUnknown   RequeueFailureKind = "commit_outcome_unknown"
	RequeueFailureCancelled              RequeueFailureKind = "cancelled"
	RequeueFailureTimedOut               RequeueFailureKind = "timed_out"
)

type RequeueFailure struct {
	Kind RequeueFailureKind
	err  error
}

func (failure *RequeueFailure) Error() string {
	if failure == nil {
		return "collaboration requeue failed"
	}
	return fmt.Sprintf("collaboration requeue failed: %s", failure.Kind)
}

func (failure *RequeueFailure) Unwrap() error {
	if failure == nil {
		return nil
	}
	return failure.err
}

type RequeueRequest struct {
	OperationID uuid.UUID
	IncidentID  uuid.UUID
	MutatedAt   time.Time
}

type RequeueResult struct {
	RequeuedIntentCount int
}

// RecoveryCapability is the deployment-local operation consumed by Operator.
// Persistence, locking, proof, journal, and transaction mechanics stay private
// to Collaboration.
type RecoveryCapability interface {
	RequeueIncident(context.Context, RequeueRequest) (RequeueResult, error)
}

type recoveryCapability struct {
	delegate privaterecovery.Capability
}

func NewRecoveryCapability(db postgres.DB) RecoveryCapability {
	return &recoveryCapability{delegate: privaterecovery.NewAdapter(db)}
}

func (capability *recoveryCapability) RequeueIncident(ctx context.Context, request RequeueRequest) (RequeueResult, error) {
	if capability == nil || capability.delegate == nil {
		return RequeueResult{}, &RequeueFailure{Kind: RequeueFailureTransaction, err: errors.New("collaboration recovery capability is not configured")}
	}
	result, err := capability.delegate.RequeueIncident(ctx, privaterecovery.Request{
		OperationID: request.OperationID,
		IncidentID:  request.IncidentID,
		MutatedAt:   request.MutatedAt,
	})
	if err != nil {
		var failure *privaterecovery.Failure
		if errors.As(err, &failure) {
			return RequeueResult{}, &RequeueFailure{Kind: publicRequeueFailureKind(failure.Kind), err: err}
		}
		return RequeueResult{}, &RequeueFailure{Kind: RequeueFailureTransaction, err: err}
	}
	return RequeueResult{RequeuedIntentCount: result.RequeuedIntentCount}, nil
}

func publicRequeueFailureKind(kind privaterecovery.FailureKind) RequeueFailureKind {
	switch kind {
	case privaterecovery.FailureIncidentNotQuarantined:
		return RequeueFailureIncidentNotQuarantined
	case privaterecovery.FailureRepairNotVerified:
		return RequeueFailureRepairNotVerified
	case privaterecovery.FailureCommitOutcomeUnknown:
		return RequeueFailureCommitOutcomeUnknown
	case privaterecovery.FailureCancelled:
		return RequeueFailureCancelled
	case privaterecovery.FailureTimedOut:
		return RequeueFailureTimedOut
	case privaterecovery.FailureTransaction:
		return RequeueFailureTransaction
	default:
		return RequeueFailureTransaction
	}
}
