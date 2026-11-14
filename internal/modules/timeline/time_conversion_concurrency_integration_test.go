package timeline

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	incidentadmission "github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestTimelineTimeProfileFirstWriteConcurrency_Integration(t *testing.T) {
	database := pgtest.Start(t).PrepareIsolatedDatabaseT(t, "timeline-profile-concurrency")
	ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
	defer cancel()
	connect := func() *pgx.Conn {
		t.Helper()
		conn, err := pgx.Connect(ctx, database.DSN)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if err := conn.Close(context.Background()); err != nil {
				t.Error(err)
			}
		})
		return conn
	}
	observer, holderConn := connect(), connect()
	writers := []*pgx.Conn{connect(), connect()}
	incident := uuid.New()
	actors := []uuid.UUID{uuid.New(), uuid.New()}
	for _, actor := range actors {
		if _, err := observer.Exec(ctx, `INSERT INTO users(id,email,display_name,password_hash,mfa_required,is_active,is_deployment_admin) VALUES($1,$2,'Profile concurrency','test-only',false,true,true)`, actor, actor.String()+"@example.test"); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := observer.Exec(ctx, `INSERT INTO incidents(id,incident_key,incident_key_canonical,title,status,created_by_user_id,updated_by_user_id) VALUES($1,$2,$2,'Profile concurrency','active',$3,$3)`, incident, incident.String(), actors[0]); err != nil {
		t.Fatal(err)
	}
	holder, err := holderConn.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = holder.Rollback(context.Background()) })
	// SHARE permits the initial SELECT FOR UPDATE on an absent row, but both
	// INSERTs must wait for this lock. Observe those waits before releasing it.
	if _, err := holder.Exec(ctx, `LOCK TABLE timeline_time_conversion_profiles IN SHARE MODE`); err != nil {
		t.Fatal(err)
	}
	offsets := []int{60, -120}
	labels := []string{"First candidate", "Second candidate"}
	clocks := []time.Time{
		time.Date(2026, 1, 2, 3, 4, 5, 123456000, time.UTC),
		time.Date(2026, 1, 2, 3, 4, 6, 654321000, time.UTC),
	}
	type outcome struct {
		index   int
		profile TimeConversionProfile
		err     error
	}
	done := make(chan outcome, 2)
	remaining := 2
	t.Cleanup(func() {
		cancel()
		_ = holder.Rollback(context.Background())
		for remaining > 0 {
			select {
			case <-done:
				remaining--
			case <-time.After(5 * time.Second):
				t.Error("profile writers did not drain")
				return
			}
		}
	})
	for i, conn := range writers {
		go func() {
			owner := &store{pool: conn, incidentAccess: incidentadmission.NewChecker(conn)}
			profile, err := owner.PutTimeConversionProfile(ctx, authn.UserRecord{ID: actors[i]}, incident, TimeConversionProfilePutRequest{
				BaseProfileVersion: 1, Enabled: true, LocalOffsetMinutes: &offsets[i], LocalLabel: &labels[i],
			}, clocks[i])
			done <- outcome{index: i, profile: profile, err: err}
		}()
	}
	for {
		var blocked int
		if err := observer.QueryRow(ctx, `SELECT count(*) FROM pg_locks WHERE pid=ANY($1::integer[]) AND relation='timeline_time_conversion_profiles'::regclass AND mode='RowExclusiveLock' AND NOT granted AND $2::integer=ANY(pg_blocking_pids(pid))`, []int32{int32(writers[0].PgConn().PID()), int32(writers[1].PgConn().PID())}, int32(holderConn.PgConn().PID())).Scan(&blocked); err != nil {
			t.Fatal(err)
		}
		if blocked == 2 {
			break
		}
		select {
		case result := <-done:
			remaining--
			t.Fatalf("writer %d completed before both insert waits: %v", result.index, result.err)
		case <-ctx.Done():
			t.Fatal("writers did not reach insert boundary", ctx.Err())
		case <-time.After(10 * time.Millisecond):
		}
	}
	if err := holder.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	winner, conflicts := -1, 0
	for remaining > 0 {
		select {
		case result := <-done:
			remaining--
			if result.err == nil {
				if winner != -1 {
					t.Fatal("both first profile writes committed")
				}
				winner = result.index
				profile := result.profile
				if profile.IncidentID != incident || profile.ProfileVersion != 2 || !profile.Enabled ||
					profile.LocalOffsetMinutes == nil || *profile.LocalOffsetMinutes != offsets[winner] ||
					profile.LocalLabel == nil || *profile.LocalLabel != labels[winner] ||
					profile.UpdatedByUserID == nil || *profile.UpdatedByUserID != actors[winner] || !profile.UpdatedAt.Equal(clocks[winner]) {
					t.Fatalf("winning response differs from its request: %#v", profile)
				}
			} else {
				var conflict *RowVersionConflictError
				if !errors.As(result.err, &conflict) || conflict.BaseRowVersion != 1 || conflict.CurrentRowVersion != 2 || result.profile != (TimeConversionProfile{}) {
					t.Fatalf("unexpected competing result: %#v / %v", result.profile, result.err)
				}
				conflicts++
			}
		case <-ctx.Done():
			t.Fatal("profile writers did not finish", ctx.Err())
		}
	}
	if winner == -1 || conflicts != 1 {
		t.Fatalf("winner=%d conflicts=%d", winner, conflicts)
	}
	var storedIncident, storedActor uuid.UUID
	var version int64
	var enabled bool
	var offset int
	var label string
	var updated time.Time
	if err := observer.QueryRow(ctx, `SELECT incident_id,enabled,local_offset_minutes,local_label,profile_version,updated_at,updated_by_user_id FROM timeline_time_conversion_profiles WHERE incident_id=$1`, incident).Scan(&storedIncident, &enabled, &offset, &label, &version, &updated, &storedActor); err != nil {
		t.Fatal(err)
	}
	if storedIncident != incident || !enabled || offset != offsets[winner] || label != labels[winner] || version != 2 || !updated.Equal(clocks[winner]) || storedActor != actors[winner] {
		t.Fatalf("loser overwrote committed profile: incident=%s enabled=%t offset=%d label=%q version=%d updated=%s actor=%s", storedIncident, enabled, offset, label, version, updated, storedActor)
	}
}
