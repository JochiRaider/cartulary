package evidence_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"maps"
	"reflect"
	"slices"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/JochiRaider/cartulary/internal/testutil/portabilitysupport"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func TestEvidencePortableCustodyTransaction_Integration(t *testing.T) {
	ctx := t.Context()
	db := pgtest.Start(t).BeginRollbackDBT(t, "evidence-custody-port")
	actor, other, incident, otherIncident := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	for _, id := range []uuid.UUID{actor, other} {
		if _, err := db.Exec(ctx, `INSERT INTO users(id,email,display_name,password_hash,mfa_required,is_active,is_deployment_admin) VALUES($1,$2,'Custody importer','test-only',false,true,true)`, id, id.String()+"@example.test"); err != nil {
			t.Fatal(err)
		}
	}
	for i, id := range []uuid.UUID{incident, otherIncident} {
		if _, err := db.Exec(ctx, `INSERT INTO incidents(id,incident_key,incident_key_canonical,title,status,created_by_user_id,updated_by_user_id) VALUES($1,$2,$2,'Custody portability','active',$3,$3)`, id, fmt.Sprintf("CUSTODY-PORT-%d", i), actor); err != nil {
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
	port := evidence.NewIncidentBundleSourcePort()
	record := evidencePortableRow(incident, 1)
	record["owner_user_id"] = nil
	seed := func(t testing.TB, tx pgx.Tx, deleted bool) {
		t.Helper()
		if _, err := tx.Exec(ctx, `INSERT INTO records(record_id,incident_id,record_type,created_by_user_id,updated_by_user_id) VALUES($1,$2,'evidence',$3,$3)`, record["record_id"], incident, actor); err != nil {
			t.Fatal(err)
		}
		if deleted {
			if _, err := tx.Exec(ctx, `UPDATE records SET deleted_at='2026-01-03T00:00:00Z',deleted_by_user_id=$2 WHERE record_id=$1`, record["record_id"], actor); err != nil {
				t.Fatal(err)
			}
		}
	}
	prepare := func(t testing.TB, rows ...map[string]any) (sourceport.Prepared, sourceport.ImportContext, *evidencePortableAttributions) {
		t.Helper()
		b := portableCustodyTestBundle(t, rows...)
		b[evidencePortablePath] = evidencePortableBundle(t, record)[evidencePortablePath]
		input := evidencePortableInput(t, incident, actor)
		p, err := port.PrepareImport(ctx, b, input)
		if err != nil {
			t.Fatal(err)
		}
		attrs := &evidencePortableAttributions{local: actor}
		input.Attributions = attrs
		input.RewrittenObjectBlobs = b
		return p, input, attrs
	}
	for _, variant := range []string{"populated", "null", "empty text", "129 rows", "empty", "deleted evidence", "chronology"} {
		t.Run(variant, func(t *testing.T) {
			tx := begin(t)
			seed(t, tx, variant == "deleted evidence")
			count := 1
			if variant == "129 rows" {
				count = 129
			}
			if variant == "empty" {
				count = 0
			}
			if variant == "chronology" {
				count = 3
			}
			rows := []map[string]any{}
			for i := 1; i <= count; i++ {
				row := portableCustodyTestRow(incident, i)
				if variant == "null" {
					row["actor_user_id"], row["location_text"], row["note"] = nil, nil, nil
					row["metadata"] = map[string]any{}
				}
				if variant == "empty text" {
					row["location_text"], row["note"] = "", ""
				}
				if variant == "chronology" {
					row["occurred_at"] = "2023-01-01T00:00:00+00:00"
					if i == 1 {
						row["occurred_at"] = "2024-01-01T00:00:00+00:00"
					}
				}
				rows = append(rows, row)
			}
			p, input, attrs := prepare(t, rows...)
			if err := port.ApplyImportTx(ctx, tx, p, input); err != nil {
				t.Fatal(err)
			}
			if err := port.ValidateImportTx(ctx, tx, p, input); err != nil {
				t.Fatal(err)
			}
			var actualCount int
			if err := tx.QueryRow(ctx, `SELECT count(*) FROM evidence_custody_events WHERE incident_id=$1`, incident).Scan(&actualCount); err != nil || actualCount != count {
				t.Fatalf("rows=%d want=%d err=%v", actualCount, count, err)
			}
			expectedAttrs := []incidentportability.ImportedAttribution{}
			for _, source := range rows {
				expected := maps.Clone(source)
				if source["actor_user_id"] != nil {
					expected["actor_user_id"] = actor.String()
					expectedAttrs = append(expectedAttrs, incidentportability.ImportedAttribution{SourceTable: "evidence_custody_events", SourceRowID: source["custody_event_id"].(string), SourceColumn: "actor_user_id", SourceActorID: evidencePortableActor.String(), LocalUserID: actor})
				}
				var raw []byte
				if err := tx.QueryRow(ctx, `SELECT to_jsonb(c) FROM evidence_custody_events c WHERE custody_event_id=$1`, source["custody_event_id"]).Scan(&raw); err != nil {
					t.Fatal(err)
				}
				var actual map[string]any
				dec := json.NewDecoder(bytes.NewReader(raw))
				dec.UseNumber()
				if err := dec.Decode(&actual); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(actual, expected) {
					t.Fatalf("stored custody=%#v want=%#v", actual, expected)
				}
			}
			if !reflect.DeepEqual(attrs.rows, expectedAttrs) && !(len(attrs.rows) == 0 && len(expectedAttrs) == 0) {
				t.Fatalf("sidecars=%#v want=%#v", attrs.rows, expectedAttrs)
			}
			exported, err := portabilitysupport.Export(port, ctx, sourceport.ExportContext{Query: tx, IncidentID: incident, PortableAttributions: attrs})
			if err != nil {
				t.Fatal(err)
			}
			bundle := portabilitysupport.MapBundle{}
			for _, file := range exported {
				bundle[file.Path] = file.Payload
			}
			exportedRows := []map[string]any{}
			if err := incidentportability.WalkPath(ctx, bundle, portableCustodyTestPath, func(row map[string]any) error { exportedRows = append(exportedRows, row); return nil }); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(exportedRows, rows) {
				t.Fatalf("re-export=%#v want=%#v", exportedRows, rows)
			}
			if variant == "deleted evidence" {
				var deleted bool
				if err := tx.QueryRow(ctx, `SELECT deleted_at IS NOT NULL AND deleted_by_user_id=$2 FROM records WHERE record_id=$1`, record["record_id"], actor).Scan(&deleted); err != nil || !deleted {
					t.Fatalf("retained event revived Evidence: %v %v", deleted, err)
				}
			}
			if variant == "chronology" {
				actual := []string{}
				result, err := tx.Query(ctx, `SELECT custody_event_id FROM evidence_custody_events WHERE incident_id=$1 ORDER BY occurred_at,custody_event_id`, incident)
				if err != nil {
					t.Fatal(err)
				}
				for result.Next() {
					var id string
					if err := result.Scan(&id); err != nil {
						t.Fatal(err)
					}
					actual = append(actual, id)
				}
				result.Close()
				if err := result.Err(); err != nil {
					t.Fatal(err)
				}
				want := []string{"40000000-0000-4000-8000-000000000002", "40000000-0000-4000-8000-000000000003", "40000000-0000-4000-8000-000000000001"}
				if !reflect.DeepEqual(actual, want) {
					t.Fatalf("chronology=%v want=%v", actual, want)
				}
			}
			var users int
			if err := tx.QueryRow(ctx, `SELECT count(*) FROM users WHERE id=$1`, evidencePortableActor).Scan(&users); err != nil || users != 0 {
				t.Fatalf("historical actor became user: %d %v", users, err)
			}
		})
	}
	t.Run("exact decimal values", func(t *testing.T) {
		tx := begin(t)
		seed(t, tx, false)
		row := portableCustodyTestRow(incident, 1)
		row["metadata"] = map[string]any{"integer": json.Number("9007199254740993"), "fraction": json.Number("1.0000000000000000000000000000000001"), "exponent": json.Number("1.2300e2"), "zero": json.Number("-0"), "nested": []any{json.Number("0.0100")}}
		p, input, _ := prepare(t, row)
		if err := port.ApplyImportTx(ctx, tx, p, input); err != nil {
			t.Fatal(err)
		}
		if err := port.ValidateImportTx(ctx, tx, p, input); err != nil {
			t.Fatal(err)
		}
		var raw []byte
		if err := tx.QueryRow(ctx, `SELECT metadata FROM evidence_custody_events WHERE custody_event_id=$1`, row["custody_event_id"]).Scan(&raw); err != nil {
			t.Fatal(err)
		}
		var actual map[string]any
		dec := json.NewDecoder(bytes.NewReader(raw))
		dec.UseNumber()
		if err := dec.Decode(&actual); err != nil {
			t.Fatal(err)
		}
		want := map[string]any{"integer": json.Number("9007199254740993"), "fraction": json.Number("1.0000000000000000000000000000000001"), "exponent": json.Number("123"), "zero": json.Number("0"), "nested": []any{json.Number("0.01")}}
		if !reflect.DeepEqual(actual, want) {
			t.Fatalf("numeric source changed=%#v want=%#v", actual, want)
		}
	})
	t.Run("complete stored source and attributions", func(t *testing.T) {
		tx := begin(t)
		seed(t, tx, false)
		source := portableCustodyTestRow(incident, 1)
		p, input, attrs := prepare(t, source)
		if err := port.ApplyImportTx(ctx, tx, p, input); err != nil {
			t.Fatal(err)
		}
		for _, change := range []struct {
			field string
			value any
		}{{"custody_event_id", uuid.New()}, {"custody_event_type", "received"}, {"actor_user_id", other}, {"occurred_at", "2024-01-01T00:00:00.123457Z"}, {"location_text", "changed"}, {"note", "changed"}} {
			t.Run(change.field, func(t *testing.T) {
				nested, err := tx.Begin(ctx)
				if err != nil {
					t.Fatal(err)
				}
				defer nested.Rollback(context.Background())
				if _, err := nested.Exec(ctx, "UPDATE evidence_custody_events SET "+change.field+"=$2 WHERE custody_event_id=$1", source["custody_event_id"], change.value); err != nil {
					t.Fatal(err)
				}
				requireEvidencePortableFailure(t, port.ValidateImportTx(ctx, nested, p, input), "source_state_exact")
			})
		}
		t.Run("exact numeric drift", func(t *testing.T) {
			nested, err := tx.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer nested.Rollback(context.Background())
			if _, err := nested.Exec(ctx, `UPDATE evidence_custody_events SET metadata=jsonb_set(metadata,'{counter}','9007199254740992'::jsonb) WHERE custody_event_id=$1`, source["custody_event_id"]); err != nil {
				t.Fatal(err)
			}
			requireEvidencePortableFailure(t, port.ValidateImportTx(ctx, nested, p, input), "source_state_exact")
		})
		t.Run("exact fraction drift", func(t *testing.T) {
			nested, err := tx.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer nested.Rollback(context.Background())
			if _, err := nested.Exec(ctx, `UPDATE evidence_custody_events SET metadata=jsonb_set(metadata,'{fraction}','1.0000000000000000000000000000000002'::jsonb) WHERE custody_event_id=$1`, source["custody_event_id"]); err != nil {
				t.Fatal(err)
			}
			requireEvidencePortableFailure(t, port.ValidateImportTx(ctx, nested, p, input), "source_state_exact")
		})
		t.Run("missing", func(t *testing.T) {
			nested, err := tx.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer nested.Rollback(context.Background())
			if _, err := nested.Exec(ctx, `DELETE FROM evidence_custody_events WHERE custody_event_id=$1`, source["custody_event_id"]); err != nil {
				t.Fatal(err)
			}
			requireEvidencePortableFailure(t, port.ValidateImportTx(ctx, nested, p, input), "source_state_exact")
		})
		t.Run("extra", func(t *testing.T) {
			nested, err := tx.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer nested.Rollback(context.Background())
			if _, err := nested.Exec(ctx, `INSERT INTO evidence_custody_events(custody_event_id,incident_id,evidence_record_id,custody_event_type,actor_user_id,occurred_at,location_text,note,metadata) SELECT $2,incident_id,evidence_record_id,custody_event_type,actor_user_id,occurred_at,location_text,note,metadata FROM evidence_custody_events WHERE custody_event_id=$1`, source["custody_event_id"], uuid.New()); err != nil {
				t.Fatal(err)
			}
			requireEvidencePortableFailure(t, port.ValidateImportTx(ctx, nested, p, input), "source_state_exact")
		})
		for _, variant := range []string{"missing", "duplicate", "table", "row", "column", "source actor", "local actor", "extra row", "extra column"} {
			t.Run("attribution "+variant, func(t *testing.T) {
				original := slices.Clone(attrs.rows)
				defer func() { attrs.rows = original }()
				switch variant {
				case "missing":
					attrs.rows = nil
				case "duplicate":
					attrs.rows = append(attrs.rows, attrs.rows[0])
				case "table":
					attrs.rows[0].SourceTable = "object_blobs"
				case "row":
					attrs.rows[0].SourceRowID = uuid.NewString()
				case "column":
					attrs.rows[0].SourceColumn = "created_by_user_id"
				case "source actor":
					attrs.rows[0].SourceActorID = uuid.NewString()
				case "local actor":
					attrs.rows[0].LocalUserID = other
				case "extra row":
					a := attrs.rows[0]
					a.SourceRowID = uuid.NewString()
					attrs.rows = append(attrs.rows, a)
				case "extra column":
					a := attrs.rows[0]
					a.SourceColumn = "other_user_id"
					attrs.rows = append(attrs.rows, a)
				}
				requireEvidencePortableFailure(t, port.ValidateImportTx(ctx, tx, p, input), "attribution_exact")
			})
		}
	})
	for _, variant := range []string{"missing", "foreign", "wrong type", "late missing"} {
		t.Run("references "+variant, func(t *testing.T) {
			tx := begin(t)
			seed(t, tx, false)
			target := uuid.New()
			if variant == "foreign" || variant == "wrong type" {
				targetIncident := incident
				kind := "timeline_event"
				if variant == "foreign" {
					targetIncident = otherIncident
					kind = "evidence"
				}
				if _, err := tx.Exec(ctx, `INSERT INTO records(record_id,incident_id,record_type,created_by_user_id,updated_by_user_id) VALUES($1,$2,$3,$4,$4)`, target, targetIncident, kind, actor); err != nil {
					t.Fatal(err)
				}
				if _, err := tx.Exec(ctx, `INSERT INTO evidence(record_id,incident_id) VALUES($1,$2)`, target, targetIncident); err != nil {
					t.Fatal(err)
				}
			}
			rows := []map[string]any{}
			if variant == "late missing" {
				rows = append(rows, portableCustodyTestRow(incident, 1))
			}
			bad := portableCustodyTestRow(incident, 2)
			bad["evidence_record_id"] = target.String()
			rows = append(rows, bad)
			p, input, attrs := prepare(t, rows...)
			requireEvidencePortableFailure(t, port.ApplyImportTx(ctx, tx, p, input), "custody_same_incident")
			var count int
			if err := tx.QueryRow(ctx, `SELECT count(*) FROM evidence_custody_events WHERE incident_id=$1`, incident).Scan(&count); err != nil || count != 0 {
				t.Fatalf("rejected reference wrote custody=%d err=%v", count, err)
			}
			if len(attrs.rows) != 0 {
				t.Fatal("rejected reference wrote custody attribution")
			}
		})
	}
	t.Run("stored Evidence reference", func(t *testing.T) {
		tx := begin(t)
		seed(t, tx, false)
		second := evidencePortableRow(incident, 2)
		second["owner_user_id"] = nil
		if _, err := tx.Exec(ctx, `INSERT INTO records(record_id,incident_id,record_type,created_by_user_id,updated_by_user_id) VALUES($1,$2,'evidence',$3,$3)`, second["record_id"], incident, actor); err != nil {
			t.Fatal(err)
		}
		source := portableCustodyTestRow(incident, 1)
		b := portableCustodyTestBundle(t, source)
		b[evidencePortablePath] = evidencePortableBundle(t, record, second)[evidencePortablePath]
		input := evidencePortableInput(t, incident, actor)
		p, err := port.PrepareImport(ctx, b, input)
		if err != nil {
			t.Fatal(err)
		}
		input.Attributions = &evidencePortableAttributions{local: actor}
		input.RewrittenObjectBlobs = b
		if err := port.ApplyImportTx(ctx, tx, p, input); err != nil {
			t.Fatal(err)
		}
		if err := port.ValidateImportTx(ctx, tx, p, input); err != nil {
			t.Fatal(err)
		}
		if _, err := tx.Exec(ctx, `UPDATE evidence_custody_events SET evidence_record_id=$2 WHERE custody_event_id=$1`, source["custody_event_id"], second["record_id"]); err != nil {
			t.Fatal(err)
		}
		requireEvidencePortableFailure(t, port.ValidateImportTx(ctx, tx, p, input), "source_state_exact")
	})
	t.Run("stored incident", func(t *testing.T) {
		tx := begin(t)
		seed(t, tx, false)
		source := portableCustodyTestRow(incident, 1)
		p, input, _ := prepare(t, source)
		if err := port.ApplyImportTx(ctx, tx, p, input); err != nil {
			t.Fatal(err)
		}
		target := uuid.New()
		if _, err := tx.Exec(ctx, `INSERT INTO records(record_id,incident_id,record_type,created_by_user_id,updated_by_user_id) VALUES($1,$2,'evidence',$3,$3)`, target, otherIncident, actor); err != nil {
			t.Fatal(err)
		}
		if _, err := tx.Exec(ctx, `INSERT INTO evidence(record_id,incident_id) VALUES($1,$2)`, target, otherIncident); err != nil {
			t.Fatal(err)
		}
		if _, err := tx.Exec(ctx, `UPDATE evidence_custody_events SET incident_id=$2,evidence_record_id=$3 WHERE custody_event_id=$1`, source["custody_event_id"], otherIncident, target); err != nil {
			t.Fatal(err)
		}
		requireEvidencePortableFailure(t, port.ValidateImportTx(ctx, tx, p, input), "source_state_exact")
	})
	t.Run("null actor rejects sidecar", func(t *testing.T) {
		tx := begin(t)
		seed(t, tx, false)
		source := portableCustodyTestRow(incident, 1)
		source["actor_user_id"] = nil
		p, input, attrs := prepare(t, source)
		if err := port.ApplyImportTx(ctx, tx, p, input); err != nil {
			t.Fatal(err)
		}
		attrs.rows = append(attrs.rows, incidentportability.ImportedAttribution{SourceTable: "evidence_custody_events", SourceRowID: source["custody_event_id"].(string), SourceColumn: "actor_user_id", SourceActorID: evidencePortableActor.String(), LocalUserID: actor})
		requireEvidencePortableFailure(t, port.ValidateImportTx(ctx, tx, p, input), "attribution_exact")
	})

	t.Run("existing custody identity preflight", func(t *testing.T) {
		tx := begin(t)
		seed(t, tx, false)
		source := portableCustodyTestRow(incident, 1)
		foreign := uuid.New()
		if _, err := tx.Exec(ctx, `INSERT INTO records(record_id,incident_id,record_type,created_by_user_id,updated_by_user_id) VALUES($1,$2,'evidence',$3,$3)`, foreign, otherIncident, actor); err != nil {
			t.Fatal(err)
		}
		if _, err := tx.Exec(ctx, `INSERT INTO evidence(record_id,incident_id) VALUES($1,$2)`, foreign, otherIncident); err != nil {
			t.Fatal(err)
		}
		if _, err := tx.Exec(ctx, `INSERT INTO evidence_custody_events(custody_event_id,incident_id,evidence_record_id,custody_event_type) VALUES($1,$2,$3,'requested')`, source["custody_event_id"], otherIncident, foreign); err != nil {
			t.Fatal(err)
		}
		p, input, attrs := prepare(t, source)
		requireEvidencePortableFailure(t, port.ApplyImportTx(ctx, tx, p, input), "source_identity_admitted")
		var count int
		if err := tx.QueryRow(ctx, `SELECT count(*) FROM evidence_custody_events WHERE incident_id=$1`, incident).Scan(&count); err != nil || count != 0 {
			t.Fatalf("identity rejection wrote custody=%d err=%v", count, err)
		}
		if len(attrs.rows) != 0 {
			t.Fatal("identity rejection wrote custody attribution")
		}
	})

}
