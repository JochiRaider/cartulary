package timeline_test

import (
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestTimelinePortableEventTransaction_Integration(t *testing.T) {
	ctx := t.Context()
	db := pgtest.Start(t).BeginRollbackDBT(t, "timeline-event-port")
	incident, actor, other, foreign := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	for _, id := range []uuid.UUID{actor, other} {
		if _, err := db.Exec(ctx, `INSERT INTO users(id,email,display_name,password_hash,mfa_required,is_active,is_deployment_admin) VALUES($1,$2,'Event importer','test-only',false,true,true)`, id, id.String()+"@example.test"); err != nil {
			t.Fatal(err)
		}
	}
	for i, id := range []uuid.UUID{incident, foreign} {
		if _, err := db.Exec(ctx, `INSERT INTO incidents(id,incident_key,incident_key_canonical,title,status,created_by_user_id,updated_by_user_id) VALUES($1,$2,$2,'Event portability','active',$3,$3)`, id, fmt.Sprintf("TL-EVENT-%d", i), actor); err != nil {
			t.Fatal(err)
		}
	}
	begin := func(t testing.TB) pgx.Tx {
		t.Helper()
		tx, err := db.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = tx.Rollback(context.Background()) })
		return tx
	}
	seed := func(t testing.TB, tx pgx.Tx, row map[string]any, scope uuid.UUID, kind string) {
		t.Helper()
		if _, err := tx.Exec(ctx, `INSERT INTO records(record_id,incident_id,record_type,row_version,created_at,updated_at,created_by_user_id,updated_by_user_id) VALUES($1,$2,$3,7,'2026-01-01T01:02:03.123456Z','2026-01-02T01:02:03.654321Z',$4,$5)`, row["record_id"], scope, kind, actor, other); err != nil {
			t.Fatal(err)
		}
	}
	port := timeline.NewIncidentBundleSourcePort()
	prepare := func(t testing.TB, rows ...map[string]any) (sourceport.Prepared, sourceport.ImportContext, *timelineProfileAttributions) {
		t.Helper()
		input := timelineProfileInput(t, incident, actor)
		prepared, err := port.PrepareImport(ctx, timelineEventBundle(t, rows...), input)
		if err != nil {
			t.Fatal(err)
		}
		attrs := &timelineProfileAttributions{local: actor}
		input.Attributions = attrs
		return prepared, input, attrs
	}
	populated := func(index int) map[string]any {
		row := timelineEventRow(incident, index)
		row["capture_state"], row["reviewed_by_user_id"], row["reviewed_at"], row["superseded_by_user_id"], row["superseded_at"] = "superseded", timelineSourceActor.String(), "2026-01-01T04:00:00.123456+00:00", timelineSourceActor.String(), "2026-01-02T04:00:00.654321+00:00"
		row["activity_utc_text"], row["activity_local_text"], row["activity_utc_generated"], row["activity_time_pair_state"] = "2026-01-01T00:00:00Z", "retained local", true, "paired_generated"
		return row
	}
	for _, variant := range []string{"populated", "nulls", "129 rows", "empty"} {
		t.Run(variant, func(t *testing.T) {
			tx := begin(t)
			count := 1
			if variant == "129 rows" {
				count = 129
			}
			if variant == "empty" {
				count = 0
			}
			rows := []map[string]any{}
			for i := 1; i <= count; i++ {
				row := populated(i)
				if variant == "nulls" {
					row = timelineEventRow(incident, i)
					for _, field := range strings.Fields(timelineTextFields) {
						row[field] = nil
					}
				}
				rows = append(rows, row)
				seed(t, tx, row, incident, "timeline_event")
			}
			prepared, input, attrs := prepare(t, rows...)
			if err := port.ApplyImportTx(ctx, tx, prepared, input); err != nil {
				t.Fatal(err)
			}
			if err := port.ValidateImportTx(ctx, tx, prepared, input); err != nil {
				t.Fatal(err)
			}
			var actualCount int
			if err := tx.QueryRow(ctx, `SELECT count(*) FROM timeline_events WHERE incident_id=$1`, incident).Scan(&actualCount); err != nil || actualCount != count {
				t.Fatalf("row count=%d want=%d err=%v", actualCount, count, err)
			}
			for _, source := range rows {
				expected := maps.Clone(source)
				if expected["reviewed_by_user_id"] != nil {
					expected["reviewed_by_user_id"] = actor.String()
				}
				if expected["superseded_by_user_id"] != nil {
					expected["superseded_by_user_id"] = actor.String()
				}
				var raw []byte
				if err := tx.QueryRow(ctx, `SELECT to_jsonb(t)-ARRAY['row_version','recorded_at','edited_at','created_by_user_id','updated_by_user_id'] FROM timeline_events t WHERE record_id=$1`, source["record_id"]).Scan(&raw); err != nil {
					t.Fatal(err)
				}
				var actual map[string]any
				if err := json.Unmarshal(raw, &actual); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(actual, expected) {
					t.Fatalf("source changed: got=%v want=%v", actual, expected)
				}
				var version int64
				var created, updated time.Time
				var creator, updater uuid.UUID
				if err := tx.QueryRow(ctx, `SELECT row_version,recorded_at,edited_at,created_by_user_id,updated_by_user_id FROM timeline_events WHERE record_id=$1`, source["record_id"]).Scan(&version, &created, &updated, &creator, &updater); err != nil || version != 7 || creator != actor || updater != other || !created.Equal(time.Date(2026, 1, 1, 1, 2, 3, 123456000, time.UTC)) || !updated.Equal(time.Date(2026, 1, 2, 1, 2, 3, 654321000, time.UTC)) {
					t.Fatalf("envelope mirrors changed: version=%d creator=%s updater=%s created=%s updated=%s err=%v", version, creator, updater, created, updated, err)
				}
				if variant != "nulls" {
					seen := map[string]bool{}
					for _, a := range attrs.rows {
						if a.SourceRowID == source["record_id"] {
							if a.SourceTable != "timeline_events" || (a.SourceColumn != "reviewed_by_user_id" && a.SourceColumn != "superseded_by_user_id") || a.SourceActorID != timelineSourceActor.String() || a.LocalUserID != actor || seen[a.SourceColumn] {
								t.Fatalf("bad attribution=%v", a)
							}
							seen[a.SourceColumn] = true
						}
					}
					if len(seen) != 2 {
						t.Fatalf("missing actor sidecar: %v", seen)
					}
				}
			}
			wantAttrs := count * 2
			if variant == "nulls" {
				wantAttrs = 0
			}
			if len(attrs.rows) != wantAttrs {
				t.Fatalf("attributions=%d want=%d", len(attrs.rows), wantAttrs)
			}
			if variant == "nulls" || variant == "empty" {
				attrs.rows = append(attrs.rows, incidentportability.ImportedAttribution{SourceTable: "timeline_events", SourceRowID: "10000000-0000-4000-8000-000000000001", SourceColumn: "reviewed_by_user_id", SourceActorID: timelineSourceActor.String(), LocalUserID: actor})
				requireTimelineProfileFailure(t, port.ValidateImportTx(ctx, tx, prepared, input), "event_attribution_exact")
			}
		})
	}
	for name, sql := range map[string]string{
		"capture":              "UPDATE timeline_events SET capture_state='enriched' WHERE record_id=$1",
		"review actor":         "UPDATE timeline_events SET reviewed_by_user_id=$2 WHERE record_id=$1",
		"review clock":         "UPDATE timeline_events SET reviewed_at=reviewed_at+interval '1 microsecond' WHERE record_id=$1",
		"superseding actor":    "UPDATE timeline_events SET superseded_by_user_id=$2 WHERE record_id=$1",
		"superseding clock":    "UPDATE timeline_events SET superseded_at=superseded_at+interval '1 microsecond' WHERE record_id=$1",
		"date":                 "UPDATE timeline_events SET date_entered_text='drift' WHERE record_id=$1",
		"analyst":              "UPDATE timeline_events SET analyst_text='drift' WHERE record_id=$1",
		"mitre":                "UPDATE timeline_events SET mitre_stage_text='drift' WHERE record_id=$1",
		"device":               "UPDATE timeline_events SET device_object_text='drift' WHERE record_id=$1",
		"ip":                   "UPDATE timeline_events SET ip_address_text='drift' WHERE record_id=$1",
		"utc":                  "UPDATE timeline_events SET activity_utc_text='drift' WHERE record_id=$1",
		"local":                "UPDATE timeline_events SET activity_local_text='drift' WHERE record_id=$1",
		"raw":                  "UPDATE timeline_events SET raw_activity_text='drift' WHERE record_id=$1",
		"synopsis":             "UPDATE timeline_events SET activity_synopsis_text='drift' WHERE record_id=$1",
		"data":                 "UPDATE timeline_events SET data_source_text='drift' WHERE record_id=$1",
		"utc generated":        "UPDATE timeline_events SET activity_utc_generated=false WHERE record_id=$1",
		"local generated":      "UPDATE timeline_events SET activity_local_generated=true WHERE record_id=$1",
		"pair":                 "UPDATE timeline_events SET activity_time_pair_state='disabled' WHERE record_id=$1",
		"version mirror":       "UPDATE timeline_events SET row_version=8 WHERE record_id=$1",
		"created clock mirror": "UPDATE timeline_events SET recorded_at=recorded_at+interval '1 microsecond' WHERE record_id=$1",
		"updated clock mirror": "UPDATE timeline_events SET edited_at=edited_at+interval '1 microsecond' WHERE record_id=$1",
		"creator mirror":       "UPDATE timeline_events SET created_by_user_id=$2 WHERE record_id=$1",
		"updater mirror":       "UPDATE timeline_events SET updated_by_user_id=created_by_user_id WHERE record_id=$1",
		"missing":              "DELETE FROM timeline_events WHERE record_id=$1",
	} {
		t.Run("state drift/"+name, func(t *testing.T) {
			tx := begin(t)
			row := populated(1)
			seed(t, tx, row, incident, "timeline_event")
			prepared, input, _ := prepare(t, row)
			if err := port.ApplyImportTx(ctx, tx, prepared, input); err != nil {
				t.Fatal(err)
			}
			args := []any{row["record_id"]}
			if strings.Contains(sql, "$2") {
				args = append(args, other)
			}
			if _, err := tx.Exec(ctx, sql, args...); err != nil {
				t.Fatal(err)
			}
			requireTimelineProfileFailure(t, port.ValidateImportTx(ctx, tx, prepared, input), "event_state_exact")
		})
	}
	t.Run("extra row", func(t *testing.T) {
		tx := begin(t)
		first, second := populated(1), populated(2)
		seed(t, tx, first, incident, "timeline_event")
		seed(t, tx, second, incident, "timeline_event")
		prepared, input, _ := prepare(t, first)
		if err := port.ApplyImportTx(ctx, tx, prepared, input); err != nil {
			t.Fatal(err)
		}
		if _, err := tx.Exec(ctx, `INSERT INTO timeline_events(record_id,incident_id,capture_state,row_version,recorded_at,edited_at,created_by_user_id,updated_by_user_id) SELECT record_id,incident_id,'rough',row_version,created_at,updated_at,created_by_user_id,updated_by_user_id FROM records WHERE record_id=$1`, second["record_id"]); err != nil {
			t.Fatal(err)
		}
		requireTimelineProfileFailure(t, port.ValidateImportTx(ctx, tx, prepared, input), "event_state_exact")
	})
	for _, variant := range []string{"missing", "duplicate", "row", "column", "source actor", "local actor", "extra"} {
		t.Run("sidecar drift/"+variant, func(t *testing.T) {
			tx := begin(t)
			row := populated(1)
			seed(t, tx, row, incident, "timeline_event")
			prepared, input, attrs := prepare(t, row)
			if err := port.ApplyImportTx(ctx, tx, prepared, input); err != nil {
				t.Fatal(err)
			}
			switch variant {
			case "missing":
				attrs.rows = attrs.rows[:1]
			case "duplicate":
				attrs.rows = append(attrs.rows, attrs.rows[0])
			case "row":
				attrs.rows[0].SourceRowID = uuid.NewString()
			case "column":
				attrs.rows[0].SourceColumn = "created_by_user_id"
			case "source actor":
				attrs.rows[0].SourceActorID = uuid.NewString()
			case "local actor":
				attrs.rows[0].LocalUserID = other
			case "extra":
				extra := attrs.rows[0]
				extra.SourceRowID = uuid.NewString()
				attrs.rows = append(attrs.rows, extra)
			}
			requireTimelineProfileFailure(t, port.ValidateImportTx(ctx, tx, prepared, input), "event_attribution_exact")
		})
	}
	for _, variant := range []string{"missing", "foreign", "wrong type", "occupied"} {
		t.Run("all rows preflight/"+variant, func(t *testing.T) {
			tx := begin(t)
			first, second := populated(1), populated(2)
			seed(t, tx, first, incident, "timeline_event")
			if variant != "missing" {
				scope, kind := incident, "timeline_event"
				if variant == "foreign" {
					scope = foreign
				}
				if variant == "wrong type" {
					kind = "host"
				}
				seed(t, tx, second, scope, kind)
			}
			if variant == "occupied" {
				if _, err := tx.Exec(ctx, `INSERT INTO timeline_events(record_id,incident_id,capture_state,created_by_user_id,updated_by_user_id,row_version) VALUES($1,$2,'rough',$3,$3,99)`, second["record_id"], incident, actor); err != nil {
					t.Fatal(err)
				}
			}
			bundle := timelineEventBundle(t, first, second)
			bundle[timelineProfilePath] = timelineProfileBundle(t, timelineProfileRow(incident))[timelineProfilePath]
			input := timelineProfileInput(t, incident, actor)
			attrs := &timelineProfileAttributions{local: actor}
			input.Attributions = attrs
			prepared, err := port.PrepareImport(ctx, bundle, input)
			if err != nil {
				t.Fatal(err)
			}
			invariant := "envelope_type_scope"
			if variant == "occupied" {
				invariant = "source_identity_admitted"
			}
			requireTimelineProfileFailure(t, port.ApplyImportTx(ctx, tx, prepared, input), invariant)
			var events, profiles int
			var version int64
			if err := tx.QueryRow(ctx, `SELECT (SELECT count(*) FROM timeline_events WHERE incident_id=$1),(SELECT count(*) FROM timeline_time_conversion_profiles WHERE incident_id=$1),COALESCE((SELECT row_version FROM timeline_events WHERE record_id=$2),0)`, incident, second["record_id"]).Scan(&events, &profiles, &version); err != nil {
				t.Fatal(err)
			}
			want := 0
			if variant == "occupied" {
				want = 1
			}
			if events != want || profiles != 0 || len(attrs.rows) != 0 || (variant == "occupied" && version != 99) {
				t.Fatalf("preflight wrote owner state: events=%d profiles=%d version=%d attrs=%v", events, profiles, version, attrs.rows)
			}
		})
	}
	var events, profiles, envelopes int
	if err := db.QueryRow(ctx, `SELECT (SELECT count(*) FROM timeline_events WHERE incident_id=$1),(SELECT count(*) FROM timeline_time_conversion_profiles WHERE incident_id=$1),(SELECT count(*) FROM records WHERE incident_id IN ($1,$2))`, incident, foreign).Scan(&events, &profiles, &envelopes); err != nil || events != 0 || profiles != 0 || envelopes != 0 {
		t.Fatalf("savepoint state survived rollback: events=%d profiles=%d envelopes=%d err=%v", events, profiles, envelopes, err)
	}
}
