package incidents_test

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
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

func TestConcurrentCrossAdminDeletionPreservesAnIncidentAdmin_Integration(t *testing.T) {
	ctx := context.Background()
	testDB := pgtest.Start(t).PrepareIsolatedDatabaseT(t, "incident-admin-concurrency")
	pool, err := pgxpool.New(ctx, testDB.DSN)
	if err != nil {
		t.Fatalf("open incident admin concurrency pool: %v", err)
	}
	t.Cleanup(pool.Close)

	firstAdmin := authstoretest.SeedLocalUserRecord(
		t, pool, "incident-admin-concurrency-first@example.test", "First concurrent admin",
		"IncidentAdminConcurrencyFirst1!", false, false, true,
	)
	secondAdmin := authstoretest.SeedLocalUserRecord(
		t, pool, "incident-admin-concurrency-second@example.test", "Second concurrent admin",
		"IncidentAdminConcurrencySecond1!", false, false, true,
	)
	application, err := incidents.NewApplication(incidents.ApplicationDependencies{
		Postgres:            pool,
		PreferenceBootstrap: workbookstartuppostgres.NewWriter(),
		Now:                 time.Now,
	})
	if err != nil {
		t.Fatalf("construct Incidents application: %v", err)
	}
	incident := storetest.CreateIncidentInStore(
		t, application, firstAdmin,
		"txn-incident-admin-concurrency-create", "IR-ADMIN-CONCURRENCY", "Concurrent incident administration",
	).Incident
	secondMembership := storetest.CreateMembershipInStore(
		t, pool, firstAdmin, incident.ID, secondAdmin,
		"txn-incident-admin-concurrency-second", "admin",
	).Membership
	firstMembership := storetest.GetMembership(t, pool, incident.ID, firstAdmin.ID)

	start := make(chan struct{})
	results := make(chan error, 2)
	var workers sync.WaitGroup
	deleteMembership := func(actor authn.UserRecord, targetID uuid.UUID, version int64, requestID string) {
		defer workers.Done()
		<-start
		_, err := application.DeleteMembership(
			ctx,
			actor,
			incident.ID,
			targetID,
			admissiontest.MembershipDelete(t, version),
			requestID,
		)
		results <- err
	}
	workers.Add(2)
	go deleteMembership(firstAdmin, secondAdmin.ID, secondMembership.MembershipVersion, "req-delete-second-admin")
	go deleteMembership(secondAdmin, firstAdmin.ID, firstMembership.MembershipVersion, "req-delete-first-admin")
	close(start)
	workers.Wait()
	close(results)

	successes := 0
	for result := range results {
		if result == nil {
			successes++
			continue
		}
		if !admission.IsDenied(result, admission.DenialNotVisible) && !errors.Is(result, incidents.ErrLastIncidentAdmin) {
			t.Fatalf("unexpected serialized cross-admin deletion result: %v", result)
		}
	}
	if successes != 1 {
		t.Fatalf("serialized cross-admin deletions succeeded %d times, want exactly one", successes)
	}

	var adminCount int
	if err := pool.QueryRow(ctx, `SELECT COUNT(*) FROM incident_memberships WHERE incident_id = $1 AND role = 'admin'`, incident.ID).Scan(&adminCount); err != nil {
		t.Fatalf("count surviving incident admins: %v", err)
	}
	if adminCount < 1 {
		t.Fatal("concurrent cross-admin deletion removed every incident admin")
	}
}
