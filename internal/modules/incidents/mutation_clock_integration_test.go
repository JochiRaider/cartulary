package incidents_test

import (
	"context"
	"fmt"
	"sync/atomic"
	"testing"
	"time"

	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/admissiontest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/storetest"
	workbookstartuppostgres "github.com/JochiRaider/cartulary/internal/modules/workbook/startup/postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func testIncidentMutationClocksAfterAdmission(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()
	for _, operation := range []string{"metadata patch", "membership patch", "membership delete"} {
		t.Run(operation, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 20*time.Second)
			defer cancel()
			identity := uuid.NewString()
			actor := authstoretest.SeedActor(t, pool, authstoretest.ActorAttributes{Email: identity + "-admin@example.test", DisplayName: "Clock admin"})
			target := authstoretest.SeedActor(t, pool, authstoretest.ActorAttributes{Email: identity + "-target@example.test", DisplayName: "Clock target"})
			initialTime := time.Now().UTC().Truncate(time.Microsecond)
			var clockValue, clockCalls atomic.Int64
			clockValue.Store(initialTime.UnixNano())
			application, err := incidents.NewApplication(incidents.ApplicationDependencies{
				Postgres: pool, PreferenceBootstrap: workbookstartuppostgres.NewWriter(),
				Now: func() time.Time {
					clockCalls.Add(1)
					return time.Unix(0, clockValue.Load())
				},
			})
			if err != nil {
				t.Fatal(err)
			}
			created, err := application.CreateIncident(ctx, actor, admissiontest.IncidentCreate(t, admissiontest.IncidentCreateInput{
				ClientTxnID: identity, IncidentKey: identity, Title: "Clock admission",
			}), identity+"-create")
			if err != nil {
				t.Fatal(err)
			}
			id := created.Incident.ID
			if _, err := application.CreateMembership(ctx, actor, id, admissiontest.MembershipCreate(t, admissiontest.MembershipCreateInput{
				ClientTxnID: identity, UserID: &target.ID, Role: "viewer",
			}), identity+"-member"); err != nil {
				t.Fatal(err)
			}
			clockCalls.Store(0)
			barrier, err := pool.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer func() { _ = barrier.Rollback(context.Background()) }()
			if _, err := barrier.Exec(ctx, `SELECT id FROM incidents WHERE id=$1 FOR UPDATE`, id); err != nil {
				t.Fatal(err)
			}
			pid := int32(barrier.Conn().PgConn().PID())
			requestID := identity + "-waiting"
			severity := "high"
			patch := admissiontest.IncidentPatch(t, admissiontest.IncidentPatchInput{
				BaseIncidentVersion: 1, Severity: admissiontest.OptionalStringInput{Present: true, Value: &severity},
			})
			memberPatch := admissiontest.MembershipPatch(t, admissiontest.MembershipPatchInput{BaseMembershipVersion: 1, Role: "reviewer"})
			memberDelete := admissiontest.MembershipDelete(t, 1)
			type outcome struct{ err error }
			result := make(chan outcome, 1)
			done := make(chan struct{})
			go func() {
				defer close(done)
				var err error
				switch operation {
				case "metadata patch":
					var changed bool
					_, changed, err = application.UpdateIncident(ctx, actor, id, patch, requestID)
					if err == nil && !changed {
						err = fmt.Errorf("metadata patch was not material")
					}
				case "membership patch":
					var changed bool
					_, changed, err = application.UpdateMembership(ctx, actor, id, target.ID, memberPatch, requestID)
					if err == nil && !changed {
						err = fmt.Errorf("membership patch was not material")
					}
				case "membership delete":
					_, err = application.DeleteMembership(ctx, actor, id, target.ID, memberDelete, requestID)
				}
				result <- outcome{err: err}
			}()
			defer func() {
				cancel()
				select {
				case <-done:
				case <-time.After(5 * time.Second):
					t.Error("mutation did not join after cancellation")
				}
			}()
			for {
				var blocked bool
				if err := pool.QueryRow(ctx, `SELECT EXISTS (SELECT 1 FROM pg_stat_activity WHERE $1=ANY(pg_blocking_pids(pid)))`, pid).Scan(&blocked); err != nil {
					t.Fatal(err)
				}
				if blocked {
					break
				}
				select {
				case early := <-result:
					t.Fatalf("mutation settled before lock release: %v", early.err)
				case <-ctx.Done():
					t.Fatal(ctx.Err())
				case <-time.After(10 * time.Millisecond):
				}
			}
			if calls := clockCalls.Load(); calls != 0 {
				t.Errorf("clock sampled before locked admission: %d calls", calls)
			}
			admittedTime := initialTime.Add(time.Hour + 1234*time.Nanosecond)
			wantTime := admittedTime.Truncate(time.Microsecond)
			clockValue.Store(admittedTime.UnixNano())
			if err := barrier.Commit(ctx); err != nil {
				t.Fatal(err)
			}
			select {
			case actual := <-result:
				if actual.err != nil {
					t.Fatal(actual.err)
				}
			case <-ctx.Done():
				t.Fatal(ctx.Err())
			}
			if calls := clockCalls.Load(); calls != 1 {
				t.Fatalf("admitted mutation clock calls = %d, want 1", calls)
			}
			projected := 1
			switch operation {
			case "metadata patch":
				projected = 0
				if count := storetest.QueryCount(t, pool, `SELECT count(*) FROM incidents WHERE id=$1 AND updated_at=$2 AND severity='high' AND incident_version=2 AND updated_by_user_id=$3`, id, wantTime, actor.ID); count != 1 {
					t.Error("incident did not retain admitted clock and exact changed state")
				}
			case "membership patch":
				if count := storetest.QueryCount(t, pool, `SELECT count(*) FROM incident_memberships WHERE incident_id=$1 AND user_id=$2 AND updated_at=$3 AND role='reviewer' AND membership_version=2 AND updated_by_user_id=$4`, id, target.ID, wantTime, actor.ID); count != 1 {
					t.Error("membership did not retain admitted clock and exact changed state")
				}
			case "membership delete":
				if count := storetest.QueryCount(t, pool, `SELECT count(*) FROM incident_memberships WHERE incident_id=$1 AND user_id=$2`, id, target.ID); count != 0 {
					t.Error("membership was not deleted")
				}
			}
			requireMutationAuditTime(t, &storetest.StoreHarness{DB: pool, Incidents: application}, requestID, wantTime, 1, projected)
		})
	}
}
