package reportcomposition

import (
	"context"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

type PreviewJobAdmission struct {
	IdempotencyKey authn.RouteIdempotencyKey
	Scope          jobs.Scope
	ActorUserID    uuid.UUID
	Normalized     []byte
	Now            time.Time
}

// PreviewJobPort is Report Composition's narrow transactional handoff to the
// Reporting owner. Report Composition persists the selected preview source;
// Reporting owns job identity, execution, rendering, and terminal proof.
type PreviewJobPort interface {
	AdmitPreviewJob(context.Context, pgx.Tx, PreviewJobAdmission) (jobs.Resource, error)
	DispatchPreviewJob(string) error
}

// LockPreviewAttemptForRenderTx revalidates the Report Composition-owned
// source/job binding in the caller's finalization transaction.
func LockPreviewAttemptForRenderTx(
	ctx context.Context,
	tx pgx.Tx,
	previewAttemptID uuid.UUID,
	renderAttemptID uuid.UUID,
) error {
	var persistedJobID uuid.UUID
	if err := tx.QueryRow(ctx, `
SELECT render_attempt_id
  FROM report_composition_preview_attempts
 WHERE preview_attempt_id = $1
 FOR UPDATE
`, previewAttemptID).Scan(&persistedJobID); err != nil {
		return err
	}
	if persistedJobID != renderAttemptID {
		return fmt.Errorf("reportcomposition: preview attempt has conflicting render job identity")
	}
	return nil
}
