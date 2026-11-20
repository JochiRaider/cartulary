package evidence

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/postgres/dbport"
)

// NewBlobQuarantinerForTest composes the private Blob quarantine workflow over
// a test database for the external lock-order interleaving tests. Quarantine
// has no public route; the returned function runs the production
// QuarantineBlob locking and state logic.
func NewBlobQuarantinerForTest(t testing.TB, database dbport.DB) func(ctx context.Context, actorUserID uuid.UUID, objectBlobID uuid.UUID) error {
	t.Helper()
	service := newTestBlobLifecycleService(t, database)
	return func(ctx context.Context, actorUserID uuid.UUID, objectBlobID uuid.UUID) error {
		_, err := service.QuarantineBlob(ctx, actorUserID, objectBlobID, "admin_quarantine", "req-lock-order-quarantine", time.Now().UTC())
		return err
	}
}

// ErrQuarantineAssociationUnstableForTest is quarantine's fail-closed outcome
// when the Blob association changed before every bounded attempt.
var ErrQuarantineAssociationUnstableForTest = errQuarantineAssociationUnstable

// QuarantineAssociationAttemptsForTest is quarantine's attempt bound
// (REQ-02-190a).
const QuarantineAssociationAttemptsForTest = quarantineAssociationAttempts
