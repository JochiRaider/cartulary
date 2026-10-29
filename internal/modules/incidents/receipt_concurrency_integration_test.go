package incidents_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/admissiontest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/storetest"
	workbookstartuppostgres "github.com/JochiRaider/cartulary/internal/modules/workbook/startup/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestConcurrentIncidentReceiptsAndMembershipReplay_Integration(t *testing.T) {
	ctx := t.Context()
	database := pgtest.Start(t).PrepareIsolatedDatabaseT(t, "incident-receipts")
	pool, err := pgxpool.New(ctx, database.DSN)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(pool.Close)
	application, err := incidents.NewApplication(incidents.ApplicationDependencies{Postgres: pool, PreferenceBootstrap: workbookstartuppostgres.NewWriter(), Now: time.Now})
	if err != nil {
		t.Fatal(err)
	}
	actor := authstoretest.SeedActor(t, pool, authstoretest.ActorAttributes{Email: "receipt-admin@example.test", DisplayName: "Admin"})
	target := authstoretest.SeedActor(t, pool, authstoretest.ActorAttributes{Email: "receipt-target@example.test", DisplayName: "Target"})
	create := admissiontest.IncidentCreate(t, admissiontest.IncidentCreateInput{ClientTxnID: "receipt-create", IncidentKey: "RECEIPT-CONCURRENT", Title: "Concurrent receipts"})
	created := concurrentIncidentReceipt(t, pool, "incidents", func(ctx context.Context) (map[string]any, error) {
		result, err := application.CreateIncident(ctx, actor, create, "receipt-create")
		return result.Payload, err
	})
	var identity struct {
		ID uuid.UUID `json:"incident_id"`
	}
	if err := json.Unmarshal(created, &identity); err != nil {
		t.Fatal(err)
	}
	id := identity.ID
	if count := storetest.QueryCount(t, pool, `SELECT count(*) FROM incidents WHERE id=$1`, id); count != 1 {
		t.Fatalf("incident count=%d", count)
	}
	member := admissiontest.MembershipCreate(t, admissiontest.MembershipCreateInput{ClientTxnID: "receipt-member", Email: &target.Email, Role: "viewer"})
	original := concurrentIncidentReceipt(t, pool, "incident_memberships", func(ctx context.Context) (map[string]any, error) {
		result, err := application.CreateMembership(ctx, actor, id, member, "receipt-member")
		return result.Payload, err
	})
	closeRequest := admissiontest.IncidentLifecycle(t, incidents.LifecycleActionClose, admissiontest.IncidentLifecycleInput{BaseIncidentVersion: 1, ClientTxnID: "receipt-close", Reason: "resolved"})
	concurrentIncidentReceipt(t, pool, "incidents", func(ctx context.Context) (map[string]any, error) {
		result, err := application.TransitionIncidentLifecycle(ctx, actor, id, closeRequest, "receipt-close")
		return result.Payload, err
	})
	if count := storetest.QueryCount(t, pool, `SELECT count(*) FROM route_idempotency WHERE actor_user_id=$1`, actor.ID); count != 3 {
		t.Fatalf("receipt count=%d", count)
	}
	conflict := admissiontest.IncidentLifecycle(t, incidents.LifecycleActionClose, admissiontest.IncidentLifecycleInput{BaseIncidentVersion: 1, ClientTxnID: "receipt-close", Reason: "changed"})
	if _, err := application.TransitionIncidentLifecycle(ctx, actor, id, conflict, "conflict"); !errors.Is(err, authn.ErrClientTxnConflict) {
		t.Fatalf("divergent stale replay=%v", err)
	}
	if _, err := application.DeleteMembership(ctx, actor, id, target.ID, admissiontest.MembershipDelete(t, 1), "remove-target"); err != nil {
		t.Fatal(err)
	}
	if _, err := pool.Exec(ctx, `UPDATE users SET email='renamed@example.test', is_active=false WHERE id=$1`, target.ID); err != nil {
		t.Fatal(err)
	}
	caseVariant := "RECEIPT-TARGET@EXAMPLE.TEST"
	replayRequest := admissiontest.MembershipCreate(t, admissiontest.MembershipCreateInput{ClientTxnID: "receipt-member", Email: &caseVariant, Role: "viewer"})
	replay, err := application.CreateMembership(ctx, actor, id, replayRequest, "replay")
	if err != nil {
		t.Fatal(err)
	}
	encoded, err := json.Marshal(replay.Payload)
	if err != nil || !bytes.Equal(encoded, original) || replay.Created {
		t.Fatalf("changed-target replay mismatch: %s err=%v", encoded, err)
	}
	if count := storetest.QueryCount(t, pool, `SELECT count(*) FROM incident_memberships WHERE incident_id=$1 AND user_id=$2`, id, target.ID); count != 0 {
		t.Fatalf("replay recreated membership: %d", count)
	}
	fresh := admissiontest.MembershipCreate(t, admissiontest.MembershipCreateInput{ClientTxnID: "fresh-inactive", UserID: &target.ID, Role: "viewer"})
	if _, err := application.CreateMembership(ctx, actor, id, fresh, "fresh-inactive"); !errors.Is(err, incidents.ErrMembershipTargetInactive) {
		t.Fatalf("fresh inactive target=%v", err)
	}
	if _, err := pool.Exec(ctx, `UPDATE incident_memberships SET role='viewer' WHERE incident_id=$1 AND user_id=$2`, id, actor.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := application.CreateMembership(ctx, actor, id, replayRequest, "unauthorized-replay"); !admission.IsDenied(err, admission.DenialInsufficientRole) {
		t.Fatalf("membership replay bypassed current role: %v", err)
	}
	if _, err := application.TransitionIncidentLifecycle(ctx, actor, id, closeRequest, "unauthorized-close-replay"); !admission.IsDenied(err, admission.DenialInsufficientRole) {
		t.Fatalf("lifecycle replay bypassed current role: %v", err)
	}
}

// A real owner-table barrier proves both requests have entered the database
// before either can commit. The polling interval is not the synchronization proof.
func concurrentIncidentReceipt(t *testing.T, pool *pgxpool.Pool, table string, mutation func(context.Context) (map[string]any, error)) []byte {
	t.Helper()
	ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
	defer cancel()
	barrier, err := pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = barrier.Rollback(context.Background()) }()
	if _, err := barrier.Exec(ctx, "LOCK TABLE "+pgx.Identifier{table}.Sanitize()+" IN SHARE MODE"); err != nil {
		t.Fatal(err)
	}
	pid := int32(barrier.Conn().PgConn().PID())
	type outcome struct {
		data map[string]any
		err  error
	}
	results := make(chan outcome, 2)
	for range 2 {
		go func() { data, err := mutation(ctx); results <- outcome{data, err} }()
	}
	for {
		var waiting int
		if err := pool.QueryRow(ctx, `WITH RECURSIVE blocked(pid) AS (
SELECT pid FROM pg_stat_activity WHERE $1=ANY(pg_blocking_pids(pid))
UNION SELECT a.pid FROM pg_stat_activity a JOIN blocked b ON b.pid=ANY(pg_blocking_pids(a.pid))
) SELECT count(DISTINCT pid) FROM blocked`, pid).Scan(&waiting); err != nil {
			t.Fatal(err)
		}
		if waiting >= 2 {
			break
		}
		select {
		case early := <-results:
			t.Fatalf("mutation settled before barrier: %v", early.err)
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		case <-time.After(10 * time.Millisecond):
		}
	}
	if err := barrier.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	var first []byte
	for range 2 {
		select {
		case result := <-results:
			if result.err != nil {
				t.Fatal(result.err)
			}
			encoded, err := json.Marshal(result.data)
			if err != nil {
				t.Fatal(err)
			}
			if first == nil {
				first = encoded
			} else if !bytes.Equal(first, encoded) {
				t.Fatalf("unequal receipts: %s / %s", first, encoded)
			}
		case <-ctx.Done():
			t.Fatal(ctx.Err())
		}
	}
	return first
}
