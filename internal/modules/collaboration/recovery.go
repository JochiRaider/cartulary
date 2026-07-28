package collaboration

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

// RecoveryService is the operator-facing quarantine recovery capability.
type RecoveryService struct {
	db postgres.DB
}

func NewRecoveryService(db postgres.DB) *RecoveryService {
	return &RecoveryService{db: db}
}

func (s *RecoveryService) RequeueIncident(ctx context.Context, incidentID uuid.UUID, now time.Time) error {
	if s == nil || s.db == nil || incidentID == uuid.Nil {
		return errors.New("collaboration recovery service is not configured")
	}
	now = now.UTC()
	tx, err := s.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return fmt.Errorf("begin collaboration incident requeue: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	tag, err := tx.Exec(ctx, `
UPDATE collaboration_incident_stream_cursors
   SET failure_count = 0,
       quarantined_at = NULL,
       quarantine_reason = NULL,
       updated_at = $2
 WHERE incident_id = $1
   AND quarantined_at IS NOT NULL
`, incidentID, now)
	if err != nil {
		return fmt.Errorf("release collaboration incident quarantine: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return errors.New("collaboration incident is not quarantined")
	}
	if _, err := tx.Exec(ctx, `
UPDATE collaboration_event_intents AS intent
   SET attempt_count = 0,
       next_attempt_at = $2,
       last_error_code = NULL,
       updated_at = $2
 WHERE intent.incident_id = $1
   AND intent.dispatch_state = 'pending'
`, incidentID, now); err != nil {
		return fmt.Errorf("requeue collaboration incident: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return fmt.Errorf("commit collaboration incident requeue: %w", err)
	}
	return nil
}
