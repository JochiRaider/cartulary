package recovery_test

import (
	"context"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

type emptyWorkbookProjectionQuery struct{}

func (emptyWorkbookProjectionQuery) QueryRows(context.Context, uuid.UUID, string, viewschema.QueryMeta) ([]map[string]any, error) {
	return []map[string]any{}, nil
}

func TestRestoreVerificationWorkbookProbe(t *testing.T) {
	ctx := context.Background()
	t.Run("nil postgres fails", func(t *testing.T) {
		err := recovery.RestoreVerificationWorkbookProbe{}.ProbeRestoredBackup(ctx, recovery.RestoreResult{})
		if err == nil || !strings.Contains(err.Error(), "requires postgres") {
			t.Fatalf("nil postgres error got %v", err)
		}
	})

	t.Run("zero incidents skips", func(t *testing.T) {
		db := pgtest.Start(t).BeginRollbackDBT(t, "restore-workbook-probe-zero-incidents")
		if err := (recovery.RestoreVerificationWorkbookProbe{Postgres: db, Query: emptyWorkbookProjectionQuery{}}).ProbeRestoredBackup(ctx, recovery.RestoreResult{}); err != nil {
			t.Fatalf("zero incidents should skip probe: %v", err)
		}
	})

	t.Run("selected incident with empty result succeeds", func(t *testing.T) {
		db := pgtest.Start(t).BeginRollbackDBT(t, "restore-workbook-probe-empty-result")
		userID := uuid.MustParse("00000000-0000-0000-0000-000000009001")
		incidentID := uuid.MustParse("00000000-0000-0000-0000-000000009101")
		if _, err := db.Exec(ctx, `
INSERT INTO users (id, email, display_name, password_hash, mfa_required, is_active, is_deployment_admin)
VALUES ($1, 'restore-probe@example.test', 'Restore Probe', 'hash', true, true, false)
`, userID); err != nil {
			t.Fatalf("seed user: %v", err)
		}
		if _, err := db.Exec(ctx, `
INSERT INTO incidents (
    id,
    incident_key,
    incident_key_canonical,
    title,
    status,
    created_by_user_id,
    updated_by_user_id
)
VALUES ($1, 'RESTORE-PROBE', 'restore-probe', 'Restore Probe', 'active', $2, $2)
`, incidentID, userID); err != nil {
			t.Fatalf("seed incident: %v", err)
		}
		if err := (recovery.RestoreVerificationWorkbookProbe{Postgres: db, Query: emptyWorkbookProjectionQuery{}}).ProbeRestoredBackup(ctx, recovery.RestoreResult{}); err != nil {
			t.Fatalf("empty timeline query should not fail probe: %v", err)
		}
	})
}
