package recovery

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration/protocol"
	"github.com/JochiRaider/cartulary/internal/platform/administrativeaudit"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type FailureKind string

const (
	FailureIncidentNotQuarantined FailureKind = "incident_not_quarantined"
	FailureRepairNotVerified      FailureKind = "repair_not_verified"
	FailureTransaction            FailureKind = "transaction_failed"
	FailureCommitOutcomeUnknown   FailureKind = "commit_outcome_unknown"
	FailureCancelled              FailureKind = "cancelled"
	FailureTimedOut               FailureKind = "timed_out"
)

type Failure struct {
	Kind FailureKind
	err  error
}

func (failure *Failure) Error() string {
	if failure == nil {
		return "collaboration recovery failed"
	}
	return fmt.Sprintf("collaboration recovery failed: %s", failure.Kind)
}

func (failure *Failure) Unwrap() error {
	if failure == nil {
		return nil
	}
	return failure.err
}

type Request struct {
	OperationID uuid.UUID
	IncidentID  uuid.UUID
	MutatedAt   time.Time
}

type Result struct {
	RequeuedIntentCount int
}

type Capability interface {
	RequeueIncident(context.Context, Request) (Result, error)
}

type Adapter struct {
	db postgres.DB
}

func NewAdapter(db postgres.DB) Capability {
	return &Adapter{db: db}
}

func (adapter *Adapter) RequeueIncident(ctx context.Context, request Request) (Result, error) {
	if adapter == nil || adapter.db == nil || request.OperationID == uuid.Nil || request.IncidentID == uuid.Nil || request.MutatedAt.IsZero() {
		return Result{}, newFailure(FailureTransaction, errors.New("collaboration recovery request is incomplete"))
	}
	mutatedAt := request.MutatedAt.UTC()
	tx, err := adapter.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return Result{}, contextualFailure(ctx, err)
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
		return Result{}, newFailure(FailureIncidentNotQuarantined, err)
	} else if err != nil {
		return Result{}, contextualFailure(ctx, err)
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
		return Result{}, contextualFailure(ctx, err)
	}
	type pendingIntent struct {
		IntentID   uuid.UUID
		IntentKey  string
		IncidentID uuid.UUID
		Family     string
		Payload    []byte
	}
	pending := make([]pendingIntent, 0)
	for rows.Next() {
		var intent pendingIntent
		if err := rows.Scan(&intent.IntentID, &intent.IntentKey, &intent.IncidentID, &intent.Family, &intent.Payload); err != nil {
			rows.Close()
			return Result{}, contextualFailure(ctx, err)
		}
		pending = append(pending, intent)
	}
	if err := rows.Err(); err != nil {
		rows.Close()
		return Result{}, contextualFailure(ctx, err)
	}
	rows.Close()
	for _, intent := range pending {
		if err := protocol.ValidateReplayablePayload(intent.IncidentID, intent.Family, intent.Payload); err != nil {
			return Result{}, newFailure(FailureRepairNotVerified, err)
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
		return Result{}, contextualFailure(ctx, err)
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
		return Result{}, contextualFailure(ctx, err)
	}
	if intentTag.RowsAffected() != int64(len(pending)) {
		return Result{}, newFailure(FailureTransaction, errors.New("locked collaboration intent set changed"))
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
		return Result{}, contextualFailure(ctx, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return Result{}, newFailure(FailureCommitOutcomeUnknown, err)
	}
	return Result{RequeuedIntentCount: len(pending)}, nil
}

func contextualFailure(ctx context.Context, err error) error {
	switch {
	case errors.Is(ctx.Err(), context.DeadlineExceeded), errors.Is(err, context.DeadlineExceeded):
		return newFailure(FailureTimedOut, err)
	case errors.Is(ctx.Err(), context.Canceled), errors.Is(err, context.Canceled):
		return newFailure(FailureCancelled, err)
	default:
		return newFailure(FailureTransaction, err)
	}
}

func newFailure(kind FailureKind, err error) error {
	return &Failure{Kind: kind, err: err}
}
