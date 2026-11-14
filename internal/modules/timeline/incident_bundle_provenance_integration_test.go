package timeline_test

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestTimelinePortableProvenanceTransaction_Integration(t *testing.T) {
	ctx := t.Context()
	db := pgtest.Start(t).BeginRollbackDBT(t, "timeline-provenance-port")
	incident, actor := uuid.New(), uuid.New()
	if _, err := db.Exec(ctx, `INSERT INTO users(id,email,display_name,password_hash,mfa_required,is_active,is_deployment_admin) VALUES($1,$2,'Provenance importer','test-only',false,true,true)`, actor, actor.String()+"@example.test"); err != nil {
		t.Fatal(err)
	}
	if _, err := db.Exec(ctx, `INSERT INTO incidents(id,incident_key,incident_key_canonical,title,status,created_by_user_id,updated_by_user_id) VALUES($1,'TL-PROVENANCE','TL-PROVENANCE','Provenance portability','active',$2,$2)`, incident, actor); err != nil {
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
	seed := func(t testing.TB, tx pgx.Tx, index int) {
		t.Helper()
		row := timelineEventRow(incident, index)
		if _, err := tx.Exec(ctx, `INSERT INTO records(record_id,incident_id,record_type,created_by_user_id,updated_by_user_id) VALUES($1,$2,'timeline_event',$3,$3)`, row["record_id"], incident, actor); err != nil {
			t.Fatal(err)
		}
	}
	port := timeline.NewIncidentBundleSourcePort()
	prepare := func(t testing.TB, rows ...map[string]any) (sourceport.Prepared, sourceport.ImportContext) {
		t.Helper()
		input := timelineProfileInput(t, incident, actor)
		bundle := timelineProvenanceBundle(t, incident, rows...)
		bundle[timelineEventPath] = timelineEventBundle(t, timelineEventRow(incident, 1), timelineEventRow(incident, 2))[timelineEventPath]
		prepared, err := port.PrepareImport(ctx, bundle, input)
		if err != nil {
			t.Fatal(err)
		}
		input.Attributions = &timelineProfileAttributions{local: actor}
		return prepared, input
	}
	for _, variant := range []string{"populated", "nulls", "129 rows", "empty"} {
		t.Run(variant, func(t *testing.T) {
			tx := begin(t)
			seed(t, tx, 1)
			seed(t, tx, 2)
			count := 1
			if variant == "129 rows" {
				count = 129
			}
			if variant == "empty" {
				count = 0
			}
			rows := []map[string]any{}
			for i := 1; i <= count; i++ {
				row := timelineProvenanceRow(incident, i)
				if variant == "nulls" {
					row["source_header"], row["cell_kind"] = nil, nil
				}
				rows = append(rows, row)
			}
			prepared, input := prepare(t, rows...)
			if err := port.ApplyImportTx(ctx, tx, prepared, input); err != nil {
				t.Fatal(err)
			}
			if err := port.ValidateImportTx(ctx, tx, prepared, input); err != nil {
				t.Fatal(err)
			}
			var actualCount int
			if err := tx.QueryRow(ctx, `SELECT count(*) FROM timeline_source_provenance p JOIN timeline_events t ON t.record_id=p.record_id WHERE t.incident_id=$1`, incident).Scan(&actualCount); err != nil || actualCount != count {
				t.Fatalf("provenance count=%d want=%d err=%v", actualCount, count, err)
			}
			for _, expected := range rows {
				var id, digest, kind, raw string
				var ordinal, column int
				var metadataJSON, headerJSON []byte
				var cell *string
				var stamp time.Time
				if err := tx.QueryRow(ctx, `SELECT record_id::text,encode(source_identity_hash,'hex'),source_row_ordinal,source_column_ordinal,source_kind,source_metadata,source_header_json,raw_value,cell_kind,created_at FROM timeline_source_provenance WHERE record_id=$1 AND source_row_ordinal=$2`, expected["record_id"], expected["source_row_ordinal"]).Scan(&id, &digest, &ordinal, &column, &kind, &metadataJSON, &headerJSON, &raw, &cell, &stamp); err != nil {
					t.Fatal(err)
				}
				var metadata, header any
				if err := json.Unmarshal(metadataJSON, &metadata); err != nil {
					t.Fatal(err)
				}
				if err := json.Unmarshal(headerJSON, &header); err != nil {
					t.Fatal(err)
				}
				if id != expected["record_id"] || digest != timelineProvenanceDigest(`{"source_kind":"capture"}`) || ordinal != expected["source_row_ordinal"] || column != 1 || kind != "capture" || !reflect.DeepEqual(metadata, map[string]any{"source_kind": "capture"}) || !reflect.DeepEqual(header, expected["source_header"]) || raw != "raw\r\n\t<>&\u200d" || !stamp.Equal(time.Date(2026, 1, 2, 3, 4, 5, 123456000, time.UTC)) {
					t.Fatalf("retained provenance changed: id=%s hash=%s row=%d column=%d kind=%s metadata=%v header=%v raw=%q clock=%s", id, digest, ordinal, column, kind, metadata, header, raw, stamp)
				}
				if variant == "nulls" {
					if cell != nil {
						t.Fatal("null cell kind changed")
					}
				} else if cell == nil || *cell != "string" {
					t.Fatal("cell kind changed")
				}
			}
		})
	}
	for name, sql := range map[string]string{
		"record":         "UPDATE timeline_source_provenance SET record_id=$2 WHERE record_id=$1",
		"digest":         "UPDATE timeline_source_provenance SET source_identity_hash=decode(repeat('0',64),'hex') WHERE record_id=$1",
		"row ordinal":    "UPDATE timeline_source_provenance SET source_row_ordinal=2 WHERE record_id=$1",
		"column ordinal": "UPDATE timeline_source_provenance SET source_column_ordinal=2 WHERE record_id=$1",
		"kind":           "UPDATE timeline_source_provenance SET source_kind='other' WHERE record_id=$1",
		"metadata":       `UPDATE timeline_source_provenance SET source_metadata='{"source_kind":"capture","locator":"drift"}'::jsonb WHERE record_id=$1`,
		"header":         "UPDATE timeline_source_provenance SET source_header_json='true'::jsonb WHERE record_id=$1",
		"raw":            "UPDATE timeline_source_provenance SET raw_value='drift' WHERE record_id=$1",
		"cell":           "UPDATE timeline_source_provenance SET cell_kind='other' WHERE record_id=$1",
		"clock":          "UPDATE timeline_source_provenance SET created_at=created_at+interval '1 microsecond' WHERE record_id=$1",
		"missing":        "DELETE FROM timeline_source_provenance WHERE record_id=$1",
	} {
		t.Run("drift/"+name, func(t *testing.T) {
			tx := begin(t)
			seed(t, tx, 1)
			seed(t, tx, 2)
			row := timelineProvenanceRow(incident, 1)
			prepared, input := prepare(t, row)
			if err := port.ApplyImportTx(ctx, tx, prepared, input); err != nil {
				t.Fatal(err)
			}
			args := []any{row["record_id"]}
			if name == "record" {
				args = append(args, timelineEventRow(incident, 2)["record_id"])
			}
			if _, err := tx.Exec(ctx, sql, args...); err != nil {
				t.Fatal(err)
			}
			requireTimelineProfileFailure(t, port.ValidateImportTx(ctx, tx, prepared, input), "provenance_state_exact")
		})
	}
	for _, empty := range []bool{false, true} {
		t.Run(map[bool]string{false: "extra row", true: "unexpected row"}[empty], func(t *testing.T) {
			tx := begin(t)
			seed(t, tx, 1)
			seed(t, tx, 2)
			row := timelineProvenanceRow(incident, 1)
			rows := []map[string]any{row}
			if empty {
				rows = nil
			}
			prepared, input := prepare(t, rows...)
			if err := port.ApplyImportTx(ctx, tx, prepared, input); err != nil {
				t.Fatal(err)
			}
			if _, err := tx.Exec(ctx, `INSERT INTO timeline_source_provenance(record_id,source_identity_hash,source_row_ordinal,source_column_ordinal,source_kind,source_metadata,source_header_json,raw_value,cell_kind,created_at) VALUES($1,decode($2,'hex'),2,1,'capture','{"source_kind":"capture"}'::jsonb,'null'::jsonb,'extra',NULL,now())`, row["record_id"], row["source_identity_sha256"]); err != nil {
				t.Fatal(err)
			}
			requireTimelineProfileFailure(t, port.ValidateImportTx(ctx, tx, prepared, input), "provenance_state_exact")
		})
	}
	t.Run("changed sealed provenance rolls back earlier writes", func(t *testing.T) {
		tx := begin(t)
		seed(t, tx, 1)
		input := timelineProfileInput(t, incident, actor)
		bundle := timelineProvenanceBundle(t, incident, timelineProvenanceRow(incident, 1))
		prepared, err := port.PrepareImport(ctx, bundle, input)
		if err != nil {
			t.Fatal(err)
		}
		input.Attributions = &timelineProfileAttributions{local: actor}
		bundle[timelineProvenancePath] = append(bundle[timelineProvenancePath], '\n')
		if err := port.ApplyImportTx(ctx, tx, prepared, input); err == nil {
			t.Fatal("changed provenance admitted")
		}
		if err := tx.Rollback(ctx); err != nil {
			t.Fatal(err)
		}
	})
	var events, provenance, envelopes int
	if err := db.QueryRow(ctx, `SELECT (SELECT count(*) FROM timeline_events WHERE incident_id=$1),(SELECT count(*) FROM timeline_source_provenance p JOIN timeline_events t ON t.record_id=p.record_id WHERE t.incident_id=$1),(SELECT count(*) FROM records WHERE incident_id=$1)`, incident).Scan(&events, &provenance, &envelopes); err != nil || events != 0 || provenance != 0 || envelopes != 0 {
		t.Fatalf("state survived rollback: events=%d provenance=%d envelopes=%d err=%v", events, provenance, envelopes, err)
	}
}
