package timeline

import (
	"context"
	"crypto/sha256"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestTimelineProvenanceCapacityConcurrency_Integration(t *testing.T) {
	database := pgtest.Start(t).PrepareIsolatedDatabaseT(t, "timeline-provenance-capacity")
	connect := func(t testing.TB) *pgx.Conn {
		t.Helper()
		conn, err := pgx.Connect(t.Context(), database.DSN)
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
	observer := connect(t)
	for _, variant := range []string{"commit fills capacity", "rollback frees capacity", "same identity replay", "same identity collision"} {
		t.Run(variant, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(t.Context(), 15*time.Second)
			defer cancel()
			incident, actor, record := uuid.New(), uuid.New(), uuid.New()
			exec := func(sql string, args ...any) {
				t.Helper()
				if _, err := observer.Exec(ctx, sql, args...); err != nil {
					t.Fatal(err)
				}
			}
			exec(`INSERT INTO users(id,email,display_name,password_hash,mfa_required,is_active,is_deployment_admin) VALUES($1,$2,'Provenance concurrency','test-only',false,true,true)`, actor, actor.String()+"@example.test")
			exec(`INSERT INTO incidents(id,incident_key,incident_key_canonical,title,status,created_by_user_id,updated_by_user_id) VALUES($1,$2,$2,'Provenance concurrency','active',$3,$3)`, incident, incident.String(), actor)
			exec(`INSERT INTO records(record_id,incident_id,record_type,created_by_user_id,updated_by_user_id) VALUES($1,$2,'timeline_event',$3,$3)`, record, incident, actor)
			exec(`INSERT INTO timeline_events(record_id,incident_id,capture_state,created_by_user_id,updated_by_user_id) VALUES($1,$2,'rough',$3,$3)`, record, incident, actor)
			hash := sha256.Sum256([]byte(`{"source_kind":"capture"}`))
			exec(`INSERT INTO timeline_source_provenance(record_id,source_identity_hash,source_row_ordinal,source_column_ordinal,source_kind,source_metadata,source_header_json,raw_value,created_at) SELECT $1,$2,n,1,'capture','{"source_kind":"capture"}','null','raw','2026-01-02T03:04:05.123456Z'::timestamptz FROM generate_series(1,4095) n`, record, hash[:])
			holderConn, waiterConn := connect(t), connect(t)
			holder, err := holderConn.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = holder.Rollback(context.Background()) })
			waiter, err := waiterConn.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() { _ = waiter.Rollback(context.Background()) })
			first := ClipboardRawImportColumn{SourceKind: "capture", SourceRowOrdinal: 4096, SourceColumnOrdinal: 1, RawValue: "first"}
			if err := insertSourceProvenanceTx(ctx, holder, record, []ClipboardRawImportColumn{first}); err != nil {
				t.Fatal(err)
			}
			var firstCreated time.Time
			if err := holder.QueryRow(ctx, `SELECT created_at FROM timeline_source_provenance WHERE record_id=$1 AND source_row_ordinal=4096`, record).Scan(&firstCreated); err != nil {
				t.Fatal(err)
			}
			second := first
			if variant == "commit fills capacity" || variant == "rollback frees capacity" {
				second.SourceRowOrdinal = 4097
				second.RawValue = "second"
			}
			if variant == "same identity collision" {
				second.RawValue = "collision"
			}
			done := make(chan error, 1)
			go func() { done <- insertSourceProvenanceTx(ctx, waiter, record, []ClipboardRawImportColumn{second}) }()
			consumed := false
			t.Cleanup(func() {
				cancel()
				_ = holder.Rollback(context.Background())
				if !consumed {
					select {
					case <-done:
					case <-time.After(5 * time.Second):
						t.Error("provenance writer did not drain")
					}
				}
			})
			for {
				var blocked bool
				if err := observer.QueryRow(ctx, `SELECT $1::integer=ANY(pg_blocking_pids($2::integer))`, int32(holderConn.PgConn().PID()), int32(waiterConn.PgConn().PID())).Scan(&blocked); err != nil {
					t.Fatal(err)
				}
				if blocked {
					break
				}
				select {
				case err := <-done:
					consumed = true
					t.Fatalf("second writer bypassed target lock: %v", err)
				case <-ctx.Done():
					t.Fatal("writer did not reach lock", ctx.Err())
				case <-time.After(10 * time.Millisecond):
				}
			}
			if variant == "rollback frees capacity" {
				err = holder.Rollback(ctx)
			} else {
				err = holder.Commit(ctx)
			}
			if err != nil {
				t.Fatal(err)
			}
			select {
			case err = <-done:
				consumed = true
			case <-ctx.Done():
				t.Fatal("writer did not finish after lock release", ctx.Err())
			}
			switch variant {
			case "commit fills capacity":
				if !errors.Is(err, ErrInvalidProvenance) || errors.Is(err, ErrProvenanceIdentityCollision) {
					t.Fatalf("capacity failure=%v", err)
				}
			case "same identity collision":
				if !errors.Is(err, ErrProvenanceIdentityCollision) {
					t.Fatalf("collision failure=%v", err)
				}
			default:
				if err != nil {
					t.Fatal(err)
				}
			}
			if err == nil {
				if err := waiter.Commit(ctx); err != nil {
					t.Fatal(err)
				}
			} else {
				if err := waiter.Rollback(ctx); err != nil {
					t.Fatal(err)
				}
			}
			var count, last int
			var raw string
			var created time.Time
			if err := observer.QueryRow(ctx, `SELECT count(*) FROM timeline_source_provenance WHERE record_id=$1`, record).Scan(&count); err != nil {
				t.Fatal(err)
			}
			if err := observer.QueryRow(ctx, `SELECT source_row_ordinal,raw_value,created_at FROM timeline_source_provenance WHERE record_id=$1 ORDER BY source_row_ordinal DESC LIMIT 1`, record).Scan(&last, &raw, &created); err != nil {
				t.Fatal(err)
			}
			if count != 4096 {
				t.Fatalf("concurrent capacity=%d", count)
			}
			if variant == "rollback frees capacity" {
				if last != 4097 || raw != "second" {
					t.Fatalf("rollback did not free slot: %d %q", last, raw)
				}
			} else if last != 4096 || raw != "first" || !created.Equal(firstCreated) {
				t.Fatalf("competing append changed first identity: %d %q %s", last, raw, created)
			}
		})
	}
}
