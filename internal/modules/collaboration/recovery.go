package collaboration

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/administrativeaudit"
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

// RecoveryService is the operator-facing quarantine recovery capability.
type RecoveryService struct {
	db postgres.DB
}

func NewRecoveryService(db postgres.DB) *RecoveryService {
	return &RecoveryService{db: db}
}

func (s *RecoveryService) RequeueIncident(ctx context.Context, request RequeueRequest) (RequeueResult, error) {
	if s == nil || s.db == nil || request.OperationID == uuid.Nil || request.IncidentID == uuid.Nil || request.MutatedAt.IsZero() {
		return RequeueResult{}, requeueFailure(RequeueFailureTransaction, errors.New("collaboration recovery request is incomplete"))
	}
	mutatedAt := request.MutatedAt.UTC()
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return RequeueResult{}, contextualRequeueFailure(ctx, err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	var (
		priorFailureCount     int
		priorQuarantinedAt    time.Time
		priorQuarantineReason string
	)
	if err := tx.QueryRow(ctx, `
SELECT failure_count, quarantined_at, quarantine_reason
  FROM collaboration_incident_stream_cursors
 WHERE incident_id = $1
   AND quarantined_at IS NOT NULL
 FOR UPDATE
`, request.IncidentID).Scan(&priorFailureCount, &priorQuarantinedAt, &priorQuarantineReason); errors.Is(err, pgx.ErrNoRows) {
		return RequeueResult{}, requeueFailure(RequeueFailureIncidentNotQuarantined, err)
	} else if err != nil {
		return RequeueResult{}, contextualRequeueFailure(ctx, err)
	}

	rows, err := tx.Query(ctx, `
SELECT intent_id, intent_key, incident_id, event_family, canonical_payload
  FROM collaboration_event_intents
 WHERE incident_id = $1
   AND dispatch_state = 'pending'
 ORDER BY intent_key
 FOR UPDATE
`, request.IncidentID)
	if err != nil {
		return RequeueResult{}, contextualRequeueFailure(ctx, err)
	}
	pending := make([]pendingIntent, 0)
	for rows.Next() {
		var intent pendingIntent
		if err := rows.Scan(&intent.IntentID, &intent.IntentKey, &intent.IncidentID, &intent.Family, &intent.Payload); err != nil {
			rows.Close()
			return RequeueResult{}, contextualRequeueFailure(ctx, err)
		}
		pending = append(pending, intent)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return RequeueResult{}, contextualRequeueFailure(ctx, err)
	}
	rows.Close()
	for _, intent := range pending {
		if err := validateReplayPayload(intent); err != nil {
			return RequeueResult{}, requeueFailure(RequeueFailureRepairNotVerified, err)
		}
	}

	if _, err := tx.Exec(ctx, `
UPDATE collaboration_incident_stream_cursors
   SET failure_count = 0,
       quarantined_at = NULL,
       quarantine_reason = NULL,
       updated_at = $2
 WHERE incident_id = $1
`, request.IncidentID, mutatedAt); err != nil {
		return RequeueResult{}, contextualRequeueFailure(ctx, err)
	}
	intentTag, err := tx.Exec(ctx, `
UPDATE collaboration_event_intents AS intent
   SET attempt_count = 0,
       next_attempt_at = $2,
       last_error_code = NULL,
       updated_at = $2
 WHERE intent.incident_id = $1
   AND intent.dispatch_state = 'pending'
`, request.IncidentID, mutatedAt)
	if err != nil {
		return RequeueResult{}, contextualRequeueFailure(ctx, err)
	}
	if intentTag.RowsAffected() != int64(len(pending)) {
		return RequeueResult{}, requeueFailure(RequeueFailureTransaction, errors.New("locked collaboration intent set changed"))
	}

	operationID := request.OperationID.String()
	incidentID := request.IncidentID
	if _, err := administrativeaudit.AppendRawTx(ctx, tx, administrativeaudit.RawEvent{
		IncidentID:  &incidentID,
		EventSource: administrativeaudit.SourceOperator,
		EventKind:   "collaboration_incident_requeued",
		ClientTxnID: &operationID,
		Before: map[string]any{
			"failure_count":     priorFailureCount,
			"quarantined_at":    priorQuarantinedAt.UTC(),
			"quarantine_reason": priorQuarantineReason,
		},
		After: map[string]any{
			"quarantine_released":   true,
			"retry_state_reset":     true,
			"requeued_intent_count": len(pending),
		},
		OccurredAt: mutatedAt,
	}); err != nil {
		return RequeueResult{}, contextualRequeueFailure(ctx, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return RequeueResult{}, requeueFailure(RequeueFailureCommitOutcomeUnknown, err)
	}
	return RequeueResult{RequeuedIntentCount: len(pending)}, nil
}

func contextualRequeueFailure(ctx context.Context, err error) error {
	switch {
	case errors.Is(ctx.Err(), context.DeadlineExceeded), errors.Is(err, context.DeadlineExceeded):
		return requeueFailure(RequeueFailureTimedOut, err)
	case errors.Is(ctx.Err(), context.Canceled), errors.Is(err, context.Canceled):
		return requeueFailure(RequeueFailureCancelled, err)
	default:
		return requeueFailure(RequeueFailureTransaction, err)
	}
}

func requeueFailure(kind RequeueFailureKind, err error) error {
	return &RequeueFailure{Kind: kind, err: err}
}
