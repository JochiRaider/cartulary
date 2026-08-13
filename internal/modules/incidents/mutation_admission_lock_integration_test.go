package incidents_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestIncidentMutationAdmissionUsesSharedLifecycleGuard_Integration(t *testing.T) {
	ctx := context.Background()
	testDB := pgtest.Start(t).PrepareGroupDatabaseT(t, "incident-mutation-admission", "incident-mutation-admission-lock")
	firstConnection, err := pgx.Connect(ctx, testDB.DSN)
	if err != nil {
		t.Fatalf("connect first mutation: %v", err)
	}
	t.Cleanup(func() { _ = firstConnection.Close(context.Background()) })
	secondConnection, err := pgx.Connect(ctx, testDB.DSN)
	if err != nil {
		t.Fatalf("connect second mutation: %v", err)
	}
	t.Cleanup(func() { _ = secondConnection.Close(context.Background()) })
	lifecycleConnection, err := pgx.Connect(ctx, testDB.DSN)
	if err != nil {
		t.Fatalf("connect incident lifecycle writer: %v", err)
	}
	t.Cleanup(func() { _ = lifecycleConnection.Close(context.Background()) })

	actorID := uuid.New()
	incidentID := uuid.New()
	if _, err := firstConnection.Exec(ctx, `
INSERT INTO users (id, email, display_name, password_hash, mfa_required, is_active, is_deployment_admin)
VALUES ($1, $2, 'Mutation admission actor', 'test-only', false, true, true)
`, actorID, "mutation-admission-"+actorID.String()+"@example.test"); err != nil {
		t.Fatalf("seed mutation admission actor: %v", err)
	}
	if _, err := firstConnection.Exec(ctx, `
INSERT INTO incidents (
    id, incident_key, incident_key_canonical, title, status,
    created_by_user_id, updated_by_user_id
) VALUES ($1, $2, $2, 'Mutation admission lock contract', 'active', $3, $3)
`, incidentID, "IR-MUTATION-ADMISSION-"+incidentID.String(), actorID); err != nil {
		t.Fatalf("seed mutation admission incident: %v", err)
	}

	firstMutation, err := firstConnection.Begin(ctx)
	if err != nil {
		t.Fatalf("begin first mutation admission: %v", err)
	}
	defer func() { _ = firstMutation.Rollback(context.Background()) }()
	secondMutation, err := secondConnection.Begin(ctx)
	if err != nil {
		t.Fatalf("begin second mutation admission: %v", err)
	}
	defer func() { _ = secondMutation.Rollback(context.Background()) }()
	access := incidents.NewAccess(firstConnection)
	if err := access.EnsureOpenTx(ctx, firstMutation, incidentID); err != nil {
		t.Fatalf("admit first concurrent mutation: %v", err)
	}
	concurrentContext, cancelConcurrent := context.WithTimeout(ctx, 2*time.Second)
	defer cancelConcurrent()
	if err := access.EnsureOpenTx(concurrentContext, secondMutation, incidentID); err != nil {
		t.Fatalf("shared incident admission serialized concurrent mutations: %v", err)
	}

	lifecycleContext, cancelLifecycle := context.WithTimeout(ctx, 100*time.Millisecond)
	defer cancelLifecycle()
	if _, err := lifecycleConnection.Exec(lifecycleContext, `
UPDATE incidents
   SET status = 'closed', closed_at = transaction_timestamp()
 WHERE id = $1
`, incidentID); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("incident lifecycle update bypassed active mutation admission: %v", err)
	}
	if err := secondMutation.Rollback(ctx); err != nil {
		t.Fatalf("release second mutation admission: %v", err)
	}
	if err := firstMutation.Rollback(ctx); err != nil {
		t.Fatalf("release first mutation admission: %v", err)
	}
	tag, err := firstConnection.Exec(ctx, `
UPDATE incidents
   SET status = 'closed', closed_at = transaction_timestamp()
 WHERE id = $1
`, incidentID)
	if err != nil {
		t.Fatalf("lifecycle update after mutation admissions settled: %v", err)
	}
	if tag.RowsAffected() != 1 {
		t.Fatalf("lifecycle update affected %d incidents, want 1", tag.RowsAffected())
	}
}
