package timeline_test

import (
	"context"
	"maps"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestTimelinePortableProfileTransaction_Integration(t *testing.T) {
	ctx := t.Context()
	db := pgtest.Start(t).BeginRollbackDBT(t, "timeline-profile-port")
	actor, other, incident := uuid.New(), uuid.New(), uuid.New()
	for _, id := range []uuid.UUID{actor, other} {
		if _, err := db.Exec(ctx, `INSERT INTO users(id,email,display_name,password_hash,mfa_required,is_active,is_deployment_admin) VALUES($1,$2,'Profile importer','test-only',false,true,true)`, id, id.String()+"@example.test"); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.Exec(ctx, `INSERT INTO incidents(id,incident_key,incident_key_canonical,title,status,created_by_user_id,updated_by_user_id) VALUES($1,'TL-PROFILE','TL-PROFILE','Profile portability','active',$2,$2)`, incident, actor); err != nil {
		t.Fatal(err)
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
	port := timeline.NewIncidentBundleSourcePort()
	prepare := func(t testing.TB, rows ...map[string]any) (sourceport.Prepared, sourceport.ImportContext, *timelineProfileAttributions) {
		t.Helper()
		input := timelineProfileInput(t, incident, actor)
		prepared, err := port.PrepareImport(ctx, timelineProfileBundle(t, rows...), input)
		if err != nil {
			t.Fatal(err)
		}
		attrs := &timelineProfileAttributions{local: actor}
		input.Attributions = attrs
		return prepared, input, attrs
	}
	for _, variant := range []string{"populated", "nulls", "absent"} {
		t.Run(variant, func(t *testing.T) {
			tx := begin(t)
			row := timelineProfileRow(incident)
			if variant == "nulls" {
				row["enabled"], row["local_offset_minutes"], row["local_label"], row["updated_by_user_id"] = false, nil, nil, nil
			}
			rows := []map[string]any{row}
			if variant == "absent" {
				rows = nil
			}
			prepared, input, attrs := prepare(t, rows...)
			if err := port.ApplyImportTx(ctx, tx, prepared, input); err != nil {
				t.Fatal(err)
			}
			if err := port.ValidateImportTx(ctx, tx, prepared, input); err != nil {
				t.Fatal(err)
			}
			var count int
			if err := tx.QueryRow(ctx, `SELECT count(*) FROM timeline_time_conversion_profiles WHERE incident_id=$1`, incident).Scan(&count); err != nil {
				t.Fatal(err)
			}
			if variant == "absent" {
				if count != 0 || len(attrs.rows) != 0 {
					t.Fatalf("absence materialized: rows=%d attributions=%v", count, attrs.rows)
				}
			} else {
				var id uuid.UUID
				var enabled bool
				var offset *int
				var label *string
				var version int64
				var stamp time.Time
				var local *uuid.UUID
				if err := tx.QueryRow(ctx, `SELECT incident_id,enabled,local_offset_minutes,local_label,profile_version,updated_at,updated_by_user_id FROM timeline_time_conversion_profiles WHERE incident_id=$1`, incident).Scan(&id, &enabled, &offset, &label, &version, &stamp, &local); err != nil {
					t.Fatal(err)
				}
				wantStamp := time.Date(2026, 1, 2, 3, 4, 5, 123456000, time.UTC)
				if count != 1 || id != incident || version != 9007199254740993 || !stamp.Equal(wantStamp) {
					t.Fatalf("profile identity/version/time changed: count=%d id=%s version=%d stamp=%s", count, id, version, stamp)
				}
				if variant == "nulls" {
					if enabled || offset != nil || label != nil || local != nil || len(attrs.rows) != 0 {
						t.Fatal("null profile fields or absent attribution changed")
					}
				} else {
					if !enabled || offset == nil || *offset != -300 || label == nil || *label != "UTC−05" || local == nil || *local != actor {
						t.Fatal("populated profile fields changed")
					}
					want := []incidentportability.ImportedAttribution{{SourceTable: "timeline_time_conversion_profiles", SourceRowID: incident.String(), SourceColumn: "updated_by_user_id", SourceActorID: timelineSourceActor.String(), LocalUserID: actor}}
					if !reflect.DeepEqual(attrs.rows, want) {
						t.Fatalf("sidecars=%v want=%v", attrs.rows, want)
					}
				}
			}
			if variant != "populated" {
				attrs.rows = append(attrs.rows, incidentportability.ImportedAttribution{SourceTable: "timeline_time_conversion_profiles", SourceRowID: incident.String(), SourceColumn: "updated_by_user_id", SourceActorID: timelineSourceActor.String(), LocalUserID: actor})
				requireTimelineProfileFailure(t, port.ValidateImportTx(ctx, tx, prepared, input), "profile_attribution_exact")
			}
		})
	}
	for _, variant := range []string{"enabled", "offset", "label", "version", "time", "actor", "missing", "unexpected"} {
		t.Run("stored drift/"+variant, func(t *testing.T) {
			tx := begin(t)
			rows := []map[string]any{timelineProfileRow(incident)}
			if variant == "unexpected" {
				rows = nil
			}
			prepared, input, _ := prepare(t, rows...)
			if err := port.ApplyImportTx(ctx, tx, prepared, input); err != nil {
				t.Fatal(err)
			}
			var err error
			switch variant {
			case "enabled":
				_, err = tx.Exec(ctx, `UPDATE timeline_time_conversion_profiles SET enabled=false WHERE incident_id=$1`, incident)
			case "offset":
				_, err = tx.Exec(ctx, `UPDATE timeline_time_conversion_profiles SET local_offset_minutes=60 WHERE incident_id=$1`, incident)
			case "label":
				_, err = tx.Exec(ctx, `UPDATE timeline_time_conversion_profiles SET local_label='Changed' WHERE incident_id=$1`, incident)
			case "version":
				_, err = tx.Exec(ctx, `UPDATE timeline_time_conversion_profiles SET profile_version=profile_version+1 WHERE incident_id=$1`, incident)
			case "time":
				_, err = tx.Exec(ctx, `UPDATE timeline_time_conversion_profiles SET updated_at=updated_at+interval '1 microsecond' WHERE incident_id=$1`, incident)
			case "actor":
				_, err = tx.Exec(ctx, `UPDATE timeline_time_conversion_profiles SET updated_by_user_id=$2 WHERE incident_id=$1`, incident, other)
			case "missing":
				_, err = tx.Exec(ctx, `DELETE FROM timeline_time_conversion_profiles WHERE incident_id=$1`, incident)
			case "unexpected":
				_, err = tx.Exec(ctx, `INSERT INTO timeline_time_conversion_profiles(incident_id,enabled,profile_version,updated_at) VALUES($1,false,1,now())`, incident)
			}
			if err != nil {
				t.Fatal(err)
			}
			requireTimelineProfileFailure(t, port.ValidateImportTx(ctx, tx, prepared, input), "profile_state_exact")
		})
	}
	for _, variant := range []string{"missing", "duplicate", "row", "column", "source actor", "local actor", "table", "extra"} {
		t.Run("sidecar drift/"+variant, func(t *testing.T) {
			tx := begin(t)
			prepared, input, attrs := prepare(t, timelineProfileRow(incident))
			if err := port.ApplyImportTx(ctx, tx, prepared, input); err != nil {
				t.Fatal(err)
			}
			switch variant {
			case "missing":
				attrs.rows = nil
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
			case "table":
				attrs.rows[0].SourceTable = "timeline_events"
			case "extra":
				extra := attrs.rows[0]
				extra.SourceRowID = uuid.NewString()
				attrs.rows = append(attrs.rows, extra)
			}
			requireTimelineProfileFailure(t, port.ValidateImportTx(ctx, tx, prepared, input), "profile_attribution_exact")
		})
	}
	for _, absent := range []bool{false, true} {
		t.Run("occupied identity/absent="+map[bool]string{false: "false", true: "true"}[absent], func(t *testing.T) {
			tx := begin(t)
			rows := []map[string]any{timelineProfileRow(incident)}
			if absent {
				rows = nil
			}
			prepared, input, attrs := prepare(t, rows...)
			if _, err := tx.Exec(ctx, `INSERT INTO timeline_time_conversion_profiles(incident_id,enabled,profile_version,updated_at) VALUES($1,false,17,now())`, incident); err != nil {
				t.Fatal(err)
			}
			requireTimelineProfileFailure(t, port.ApplyImportTx(ctx, tx, prepared, input), "source_identity_admitted")
			var version int64
			var events int
			if err := tx.QueryRow(ctx, `SELECT profile_version,(SELECT count(*) FROM timeline_events WHERE incident_id=$1) FROM timeline_time_conversion_profiles WHERE incident_id=$1`, incident).Scan(&version, &events); err != nil || version != 17 || events != 0 || len(attrs.rows) != 0 {
				t.Fatalf("collision mutated owner: version=%d events=%d attrs=%v err=%v", version, events, attrs.rows, err)
			}
		})
	}
	// Both successful writes and rejected validation cases above roll back their savepoint.
	var profiles, events, provenance int
	if err := db.QueryRow(ctx, `SELECT (SELECT count(*) FROM timeline_time_conversion_profiles WHERE incident_id=$1),(SELECT count(*) FROM timeline_events WHERE incident_id=$1),(SELECT count(*) FROM timeline_source_provenance p JOIN records r ON r.record_id=p.record_id WHERE r.incident_id=$1)`, incident).Scan(&profiles, &events, &provenance); err != nil || profiles != 0 || events != 0 || provenance != 0 {
		t.Fatalf("owner state survived rollback: profiles=%d events=%d provenance=%d err=%v", profiles, events, provenance, err)
	}
	// Preparation rejection remains a transaction-free operation for noncanonical input.
	bad := maps.Clone(timelineProfileRow(incident))
	bad["local_label"] = " not canonical "
	_, err := port.PrepareImport(ctx, timelineProfileBundle(t, bad), timelineProfileInput(t, incident, actor))
	requireTimelineProfileFailure(t, err, "version_shape_exact")
}
