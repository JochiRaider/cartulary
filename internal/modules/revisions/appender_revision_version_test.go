package revisions_test

import (
	"context"
	"errors"
	"testing"
	"time"

	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	entitytest "github.com/JochiRaider/cartulary/internal/modules/entities/testsupport"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/revisionsupport"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func exerciseAppenderRevisionVersions(t *testing.T) {
	h := appsupport.StartStore(t, "append-revision-version")
	actor := authstoretest.SeedActor(t, h.DB, authstoretest.ActorAttributes{Email: "append-version@example.test", DisplayName: "Append version"})
	incident := appsupport.CreateIncidentInStore(t, h.DB, actor, "append-version-incident", "IR-VERSION", "Append version")
	record := uuid.New()
	entitytest.SeedHostRecord(t, h.DB, incident.ID, actor.ID, record, "Version one", "version-host", "", "")
	appender := revisionsupport.MustComposition(t).Runtime.Appender()
	for _, method := range []string{"live", "historical"} {
		for _, tc := range []struct {
			name          string
			version       int64
			before, after int
			valid         bool
		}{
			{"wrong after version", 2, 0, 0, false},
			{"unexpected creation before", 1, 0, 0, false},
			{"missing later before", 2, -1, 1, false},
			{"skipped version", 3, 0, 1, false},
			{"unchanged version", 2, 1, 1, false},
			{"zero declared version", 0, -1, 0, false},
			{"creation", 1, -1, 0, true},
			{"adjacent advance", 2, 0, 1, true},
		} {
			t.Run(method+"/"+tc.name, func(t *testing.T) {
				tx, err := h.DB.BeginTx(t.Context(), pgx.TxOptions{})
				if err != nil {
					t.Fatal(err)
				}
				defer tx.Rollback(context.Background())
				first, err := appender.CaptureRecordSnapshotTx(t.Context(), tx, record)
				if err != nil {
					t.Fatal(err)
				}
				now := time.Now().UTC().Truncate(time.Microsecond)
				if _, err := tx.Exec(t.Context(), "UPDATE records SET row_version=2,updated_at=$2 WHERE record_id=$1", record, now); err != nil {
					t.Fatal(err)
				}
				if _, err := tx.Exec(t.Context(), "UPDATE hosts SET row_version=2,updated_at=$2,display_name='Version two' WHERE record_id=$1", record, now); err != nil {
					t.Fatal(err)
				}
				second, err := appender.CaptureRecordSnapshotTx(t.Context(), tx, record)
				if err != nil {
					t.Fatal(err)
				}
				snapshots := []*revisions.RecordSnapshot{&first, &second}
				var before *revisions.RecordSnapshot
				if tc.before >= 0 {
					before = snapshots[tc.before]
				}
				changed, err := appender.AppendChangeSetTx(t.Context(), tx, revisions.AppendChangeSetParams{IncidentID: incident.ID, ActorUserID: actor.ID, Source: "append.version.test", CreatedAt: now})
				if err != nil {
					t.Fatal(err)
				}
				if method == "live" {
					names := []string{"Version one", "Version two"}
					fact := revisions.RevisionConflictFact{FieldKey: "host.display_name", AfterPresent: true, AfterValue: names[tc.after]}
					if tc.before >= 0 {
						fact.BeforePresent, fact.BeforeValue = true, names[tc.before]
					}
					err = appender.AppendLiveRevisionTx(t.Context(), tx, revisions.LiveRevisionInput{ChangeSetID: changed, RecordID: record, RowVersion: tc.version, BeforeSnapshot: before, AfterSnapshot: snapshots[tc.after], ConflictFacts: []revisions.RevisionConflictFact{fact}})
				} else {
					err = appender.AppendHistoricalRevisionTx(t.Context(), tx, revisions.HistoricalRevisionInput{ChangeSetID: changed, RecordID: record, RowVersion: tc.version, BeforeSnapshot: before, AfterSnapshot: snapshots[tc.after]})
				}
				if tc.valid {
					if err != nil {
						t.Fatal(err)
					}
				} else if !errors.Is(err, revisions.ErrInvalidRecordSnapshot) {
					t.Errorf("version mismatch error=%v; want invalid snapshot", err)
				}
				var count, factCount int
				if err := tx.QueryRow(t.Context(), `SELECT count(*),(SELECT count(*) FROM record_revision_conflict_facts f JOIN record_revisions r USING(revision_id) WHERE r.change_set_id=$1) FROM record_revisions WHERE change_set_id=$1`, changed).Scan(&count, &factCount); err != nil {
					t.Fatal(err)
				}
				wantCount, wantFacts := 0, 0
				if tc.valid {
					wantCount = 1
					if method == "live" {
						wantFacts = 1
					}
				}
				if count != wantCount || factCount != wantFacts {
					t.Fatalf("persisted revisions=%d facts=%d; want %d/%d", count, factCount, wantCount, wantFacts)
				}
				if tc.valid {
					var exact bool
					if err := tx.QueryRow(t.Context(), `SELECT record_id=$2 AND row_version=$3 AND (after_json->'record'->>'row_version')::bigint=$3
AND CASE WHEN $3=1 THEN coalesce(before_json,'null'::jsonb)='null'::jsonb ELSE (before_json->'record'->>'row_version')::bigint=$3-1 END
FROM record_revisions WHERE change_set_id=$1`, changed, record, tc.version).Scan(&exact); err != nil || !exact {
						t.Fatalf("persisted revision version pair exact=%v error=%v", exact, err)
					}
				}
			})
		}
	}
}
