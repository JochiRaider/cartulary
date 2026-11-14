package evidence_test

import (
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

func TestEvidencePortableRecordTransaction_Integration(t *testing.T) {
	ctx := t.Context()
	db := pgtest.Start(t).BeginRollbackDBT(t, "evidence-record-port")
	actor, other, incident := uuid.New(), uuid.New(), uuid.New()
	for _, id := range []uuid.UUID{actor, other} {
		if _, err := db.Exec(ctx, `INSERT INTO users(id,email,display_name,password_hash,mfa_required,is_active,is_deployment_admin) VALUES($1,$2,'Evidence importer','test-only',false,true,true)`, id, id.String()+"@example.test"); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := db.Exec(ctx, `INSERT INTO incidents(id,incident_key,incident_key_canonical,title,status,created_by_user_id,updated_by_user_id) VALUES($1,'EV-PORT','EV-PORT','Evidence portability','active',$2,$2)`, incident, actor); err != nil {
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
	seed := func(t testing.TB, tx pgx.Tx, row map[string]any, kind string) {
		t.Helper()
		if _, err := tx.Exec(ctx, `INSERT INTO records(record_id,incident_id,record_type,created_by_user_id,updated_by_user_id) VALUES($1,$2,$3,$4,$4)`, row["record_id"], incident, kind, actor); err != nil {
			t.Fatal(err)
		}
	}
	port := evidence.NewIncidentBundleSourcePort()
	prepare := func(t testing.TB, rows ...map[string]any) (sourceport.Prepared, sourceport.ImportContext, *evidencePortableAttributions) {
		t.Helper()
		input := evidencePortableInput(t, incident, actor)
		bundle := evidencePortableBundle(t, rows...)
		prepared, err := port.PrepareImport(ctx, bundle, input)
		if err != nil {
			t.Fatal(err)
		}
		attribution := &evidencePortableAttributions{local: actor}
		input.Attributions = attribution
		input.RewrittenObjectBlobs = bundle
		return prepared, input, attribution
	}
	prepareAssociation := func(t testing.TB, rows []map[string]any, blobs ...map[string]any) (sourceport.Prepared, sourceport.ImportContext, *evidencePortableAttributions) {
		t.Helper()
		bundle := portableBlobTestBundle(t, blobs...)
		bundle[evidencePortablePath] = evidencePortableBundle(t, rows...)[evidencePortablePath]
		input := evidencePortableInput(t, incident, actor)
		prepared, err := port.PrepareImport(ctx, bundle, input)
		if err != nil {
			t.Fatal(err)
		}
		attributions := &evidencePortableAttributions{local: actor}
		input.Attributions = attributions
		input.RewrittenObjectBlobs = portableBlobRewriteFixture(t, blobs)
		return prepared, input, attributions
	}
	for _, lifecycle := range []string{"requested", "pending_receipt", "received", "available", "quarantined", "released"} {
		for _, blobState := range []string{"available", "quarantined", "pending", "failed"} {
			for _, deleted := range []bool{false, true} {
				t.Run(fmt.Sprintf("blob bridge/%s/%s/deleted=%t", lifecycle, blobState, deleted), func(t *testing.T) {
					tx := begin(t)
					blob := portableBlobTestRow(incident, 1, blobState)
					row := evidencePortableRow(incident, 1)
					row["object_blob_id"], row["lifecycle_state"], row["upload_state"] = blob["object_blob_id"], lifecycle, "available"
					if lifecycle == "quarantined" {
						row["upload_state"] = "quarantined"
					}
					seed(t, tx, row, "evidence")
					if deleted {
						if _, err := tx.Exec(ctx, `UPDATE records SET deleted_at=now(),deleted_by_user_id=$2 WHERE record_id=$1`, row["record_id"], actor); err != nil {
							t.Fatal(err)
						}
					}
					prepared, input, _ := prepareAssociation(t, []map[string]any{row}, blob)
					bridgeLegal := lifecycle == "requested" || lifecycle == "pending_receipt" || lifecycle == "received" || (lifecycle == "quarantined" && blobState == "quarantined") || ((lifecycle == "available" || lifecycle == "released") && blobState == "available")
					valid := (blobState == "available" || blobState == "quarantined") && (deleted || bridgeLegal)
					err := port.ApplyImportTx(ctx, tx, prepared, input)
					if !valid {
						requireEvidencePortableFailure(t, err, "lifecycle_legal")
						var count int
						if err := tx.QueryRow(ctx, `SELECT count(*) FROM evidence WHERE incident_id=$1`, incident).Scan(&count); err != nil || count != 0 {
							t.Fatalf("invalid bridge inserted Evidence: count=%d err=%v", count, err)
						}
						return
					}
					if err != nil {
						t.Fatal(err)
					}
					if err := port.ValidateImportTx(ctx, tx, prepared, input); err != nil {
						t.Fatal(err)
					}
					var storedLifecycle, storedCopy, storedBlob string
					if err := tx.QueryRow(ctx, `SELECT lifecycle_state,upload_state,object_blob_id::text FROM evidence WHERE record_id=$1`, row["record_id"]).Scan(&storedLifecycle, &storedCopy, &storedBlob); err != nil || storedLifecycle != lifecycle || storedCopy != row["upload_state"] || storedBlob != blob["object_blob_id"] {
						t.Fatalf("retained association changed: lifecycle=%s copy=%s blob=%s err=%v", storedLifecycle, storedCopy, storedBlob, err)
					}
					if deleted && !bridgeLegal {
						if _, err := tx.Exec(ctx, `UPDATE records SET deleted_at=NULL,deleted_by_user_id=NULL WHERE record_id=$1`, row["record_id"]); err != nil {
							t.Fatal(err)
						}
						requireEvidencePortableFailure(t, port.ValidateImportTx(ctx, tx, prepared, input), "lifecycle_legal")
					}
				})
			}
		}
	}
	for _, target := range []string{"missing", "foreign", "duplicate"} {
		t.Run("blob association "+target, func(t *testing.T) {
			tx := begin(t)
			blob := portableBlobTestRow(incident, 1, "available")
			row := evidencePortableRow(incident, 1)
			row["object_blob_id"], row["upload_state"] = blob["object_blob_id"], "available"
			seed(t, tx, row, "evidence")
			rows := []map[string]any{row}
			blobs := []map[string]any{}
			if target == "duplicate" {
				second := evidencePortableRow(incident, 2)
				second["object_blob_id"], second["upload_state"] = blob["object_blob_id"], "available"
				seed(t, tx, second, "evidence")
				rows, blobs = append(rows, second), append(blobs, blob)
			}
			if target == "foreign" {
				foreign := uuid.New()
				if _, err := tx.Exec(ctx, `INSERT INTO incidents(id,incident_key,incident_key_canonical,title,status,created_by_user_id,updated_by_user_id) VALUES($1,'EV-BLOB-FOREIGN','EV-BLOB-FOREIGN','Foreign Blob','active',$2,$2)`, foreign, actor); err != nil {
					t.Fatal(err)
				}
				if _, err := tx.Exec(ctx, `INSERT INTO object_blobs(object_blob_id,incident_id,created_by_user_id,storage_key,byte_size,target_expires_at,pending_expires_at) VALUES($1,$2,$3,'foreign-fixture',0,now()+interval '1 hour',now()+interval '24 hours')`, blob["object_blob_id"], foreign, actor); err != nil {
					t.Fatal(err)
				}
			}
			prepared, input, _ := prepareAssociation(t, rows, blobs...)
			requireEvidencePortableFailure(t, port.ApplyImportTx(ctx, tx, prepared, input), "object_metadata_agree")
			if err := tx.Rollback(ctx); err != nil {
				t.Fatal(err)
			}
			var records, blobsStored int
			if err := db.QueryRow(ctx, `SELECT (SELECT count(*) FROM evidence WHERE incident_id=$1),(SELECT count(*) FROM object_blobs WHERE incident_id=$1)`, incident).Scan(&records, &blobsStored); err != nil || records != 0 || blobsStored != 0 {
				t.Fatalf("failed association survived rollback: evidence=%d blobs=%d err=%v", records, blobsStored, err)
			}
		})
	}
	for _, field := range []string{"collector_party_id", "source_party_id"} {
		for _, deleted := range []bool{false, true} {
			for _, target := range []string{"active", "deleted", "missing", "wrong type", "missing subtype", "foreign"} {
				t.Run(fmt.Sprintf("%s/deleted=%t/%s", field, deleted, target), func(t *testing.T) {
					tx := begin(t)
					row := evidencePortableRow(incident, 1)
					seed(t, tx, row, "evidence")
					if deleted {
						if _, err := tx.Exec(ctx, `UPDATE records SET deleted_at=now(),deleted_by_user_id=$2 WHERE record_id=$1`, row["record_id"], actor); err != nil {
							t.Fatal(err)
						}
					}
					party, partyIncident := uuid.New(), incident
					row[field] = party.String()
					if target == "foreign" {
						partyIncident = uuid.New()
						if _, err := tx.Exec(ctx, `INSERT INTO incidents(id,incident_key,incident_key_canonical,title,status,created_by_user_id,updated_by_user_id) VALUES($1,'EV-FOREIGN','EV-FOREIGN','Foreign','active',$2,$2)`, partyIncident, actor); err != nil {
							t.Fatal(err)
						}
					}
					if target != "missing" {
						kind := "party"
						if target == "wrong type" {
							kind = "host"
						}
						if _, err := tx.Exec(ctx, `INSERT INTO records(record_id,incident_id,record_type,created_by_user_id,updated_by_user_id) VALUES($1,$2,$3,$4,$4)`, party, partyIncident, kind, actor); err != nil {
							t.Fatal(err)
						}
						if target != "missing subtype" {
							if _, err := tx.Exec(ctx, `INSERT INTO parties(record_id,incident_id,display_name,party_kind) VALUES($1,$2,'Retained Party','organization')`, party, partyIncident); err != nil {
								t.Fatal(err)
							}
						}
						if target == "deleted" {
							if _, err := tx.Exec(ctx, `UPDATE records SET deleted_at=now(),deleted_by_user_id=$2 WHERE record_id=$1`, party, actor); err != nil {
								t.Fatal(err)
							}
						}
					}
					prepared, input, attributions := prepare(t, row)
					err := port.ApplyImportTx(ctx, tx, prepared, input)
					valid := target == "active" || (deleted && target == "deleted")
					if !valid {
						requireEvidencePortableFailure(t, err, "references_same_incident")
						requireEvidenceOwnerEmpty(t, ctx, tx, incident, attributions)
						return
					}
					if err != nil {
						t.Fatal(err)
					}
					if err := port.ValidateImportTx(ctx, tx, prepared, input); err != nil {
						t.Fatal(err)
					}
					var collector, source *string
					if err := tx.QueryRow(ctx, `SELECT collector_party_id::text,source_party_id::text FROM evidence WHERE record_id=$1`, row["record_id"]).Scan(&collector, &source); err != nil {
						t.Fatal(err)
					}
					want := party.String()
					var wantCollector, wantSource *string
					if field == "collector_party_id" {
						wantCollector = &want
					} else {
						wantSource = &want
					}
					if !reflect.DeepEqual(collector, wantCollector) || !reflect.DeepEqual(source, wantSource) {
						t.Fatalf("Party references changed: collector=%v source=%v", collector, source)
					}
					for _, drift := range []string{"missing subtype", "wrong type", "deleted"} {
						if drift == "deleted" && deleted {
							continue
						}
						t.Run("target drift "+drift, func(t *testing.T) {
							nested, err := tx.Begin(ctx)
							if err != nil {
								t.Fatal(err)
							}
							defer nested.Rollback(context.Background())
							switch drift {
							case "missing subtype":
								_, err = nested.Exec(ctx, `DELETE FROM parties WHERE record_id=$1`, party)
							case "wrong type":
								_, err = nested.Exec(ctx, `UPDATE records SET record_type='host' WHERE record_id=$1`, party)
							case "deleted":
								_, err = nested.Exec(ctx, `UPDATE records SET deleted_at=now(),deleted_by_user_id=$2 WHERE record_id=$1`, party, actor)
							}
							if err != nil {
								t.Fatal(err)
							}
							requireEvidencePortableFailure(t, port.ValidateImportTx(ctx, nested, prepared, input), "references_same_incident")
						})
					}
				})
			}
		}
	}
	t.Run("later Party reference rejects before owner writes", func(t *testing.T) {
		tx := begin(t)
		first, second := evidencePortableRow(incident, 1), evidencePortableRow(incident, 2)
		second["source_party_id"] = uuid.NewString()
		seed(t, tx, first, "evidence")
		seed(t, tx, second, "evidence")
		blob := portableBlobTestRow(incident, 1, "pending")
		bundle := portableBlobTestBundle(t, blob)
		bundle[evidencePortablePath] = evidencePortableBundle(t, first, second)[evidencePortablePath]
		input := evidencePortableInput(t, incident, actor)
		prepared, err := port.PrepareImport(ctx, bundle, input)
		if err != nil {
			t.Fatal(err)
		}
		attributions := &evidencePortableAttributions{local: actor}
		input.Attributions = attributions
		input.RewrittenObjectBlobs = portableBlobRewriteFixture(t, []map[string]any{blob})
		requireEvidencePortableFailure(t, port.ApplyImportTx(ctx, tx, prepared, input), "references_same_incident")
		requireEvidenceOwnerEmpty(t, ctx, tx, incident, attributions)
		second["source_party_id"] = nil
		bundle[evidencePortablePath] = evidencePortableBundle(t, first, second)[evidencePortablePath]
		prepared, err = port.PrepareImport(ctx, bundle, input)
		if err != nil {
			t.Fatal(err)
		}
		if err := port.ApplyImportTx(ctx, tx, prepared, input); err != nil {
			t.Fatal(err)
		}
		if err := port.ValidateImportTx(ctx, tx, prepared, input); err != nil {
			t.Fatal(err)
		}
		var records, blobs int
		if err := tx.QueryRow(ctx, `SELECT (SELECT count(*) FROM evidence WHERE incident_id=$1),(SELECT count(*) FROM object_blobs WHERE incident_id=$1)`, incident).Scan(&records, &blobs); err != nil || records != 2 || blobs != 1 || len(attributions.rows) != 3 {
			t.Fatalf("valid control: evidence=%d blobs=%d attributions=%d err=%v", records, blobs, len(attributions.rows), err)
		}
	})
	for _, variant := range []string{"populated", "null", "129 rows", "empty"} {
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
				row := evidencePortableRow(incident, i)
				if variant == "null" {
					for _, field := range []string{"title", "requested_at", "storage_ref", "collector_party_text", "source_party_text", "owner_user_id"} {
						row[field] = nil
					}
				}
				seed(t, tx, row, "evidence")
				rows = append(rows, row)
			}
			prepared, input, attributions := prepare(t, rows...)
			if err := port.ApplyImportTx(ctx, tx, prepared, input); err != nil {
				t.Fatal(err)
			}
			if err := port.ValidateImportTx(ctx, tx, prepared, input); err != nil {
				t.Fatal(err)
			}
			var countStored int
			if err := tx.QueryRow(ctx, `SELECT count(*) FROM evidence WHERE incident_id=$1`, incident).Scan(&countStored); err != nil || countStored != count {
				t.Fatalf("rows=%d want=%d err=%v", countStored, count, err)
			}
			expectedAttributions := []incidentportability.ImportedAttribution{}
			for _, source := range rows {
				expected := maps.Clone(source)
				if source["owner_user_id"] != nil {
					expected["owner_user_id"] = actor.String()
					expectedAttributions = append(expectedAttributions, incidentportability.ImportedAttribution{SourceTable: "evidence", SourceRowID: source["record_id"].(string), SourceColumn: "owner_user_id", SourceActorID: evidencePortableActor.String(), LocalUserID: actor})
				}
				var raw []byte
				if err := tx.QueryRow(ctx, `SELECT to_jsonb(e) FROM evidence e WHERE record_id=$1`, source["record_id"]).Scan(&raw); err != nil {
					t.Fatal(err)
				}
				var actual map[string]any
				if err := json.Unmarshal(raw, &actual); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(actual, expected) {
					t.Fatalf("stored Evidence=%#v want=%#v", actual, expected)
				}
			}
			if !reflect.DeepEqual(attributions.rows, expectedAttributions) && !(len(attributions.rows) == 0 && len(expectedAttributions) == 0) {
				t.Fatalf("attributions=%#v want=%#v", attributions.rows, expectedAttributions)
			}
			exported, err := portabilitysupport.Export(port, ctx, sourceport.ExportContext{Query: tx, IncidentID: incident, PortableAttributions: attributions})
			if err != nil {
				t.Fatal(err)
			}
			exportedBundle := portabilitysupport.MapBundle{}
			for _, file := range exported {
				exportedBundle[file.Path] = file.Payload
			}
			exportedRows := []map[string]any{}
			if err := incidentportability.WalkPath(ctx, exportedBundle, evidencePortablePath, func(row map[string]any) error { exportedRows = append(exportedRows, row); return nil }); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(exportedRows, rows) {
				t.Fatalf("portable Evidence changed on re-export: got=%#v want=%#v", exportedRows, rows)
			}
			if variant != "populated" {
				return
			}
			// A second apply cannot silently skip an existing source identity.
			before := slices.Clone(attributions.rows)
			requireEvidencePortableFailure(t, port.ApplyImportTx(ctx, tx, prepared, input), "source_identity_admitted")
			if !reflect.DeepEqual(before, attributions.rows) {
				t.Fatal("rejected duplicate changed attribution")
			}
			if err := port.ValidateImportTx(ctx, tx, prepared, input); err != nil {
				t.Fatal("duplicate changed source", err)
			}
			id := rows[0]["record_id"]
			blob := uuid.New()
			for _, change := range []struct {
				field string
				value any
			}{
				{"title", "Changed"}, {"lifecycle_state", "pending_receipt"}, {"requested_at", "2026-01-01T00:00:00.123457Z"}, {"received_at", "2026-01-04T00:00:00Z"},
				{"storage_ref", "ticket://changed"}, {"blob_hash", "changed"}, {"collector_party_text", "Changed collector"}, {"collector_party_id", uuid.New()},
				{"source_party_text", "Changed source"}, {"source_party_id", uuid.New()}, {"upload_state", "failed"}, {"object_blob_id", blob}, {"owner_user_id", other},
				{"created_at", "2025-01-01T00:00:00Z"}, {"updated_at", "2027-01-01T00:00:00Z"},
			} {
				t.Run("drift_"+change.field, func(t *testing.T) {
					nested, err := tx.Begin(ctx)
					if err != nil {
						t.Fatal(err)
					}
					defer nested.Rollback(context.Background())
					// Only the association drift needs this foreign-key fixture. It must
					// not create an unrelated extra Blob for every other field/sidecar case.
					if change.field == "object_blob_id" {
						if _, err := nested.Exec(ctx, `INSERT INTO object_blobs(object_blob_id,incident_id,created_by_user_id,storage_key,byte_size,target_expires_at,pending_expires_at) VALUES($1,$2,$3,'portable-fixture',0,now()+interval '1 hour',now()+interval '24 hours')`, blob, incident, actor); err != nil {
							t.Fatal(err)
						}
					}
					if _, err := nested.Exec(ctx, "UPDATE evidence SET "+change.field+"=$2 WHERE record_id=$1", id, change.value); err != nil {
						t.Fatal(err)
					}
					requireEvidencePortableFailure(t, port.ValidateImportTx(ctx, nested, prepared, input), "source_state_exact")
				})
			}
			t.Run("missing row", func(t *testing.T) {
				nested, err := tx.Begin(ctx)
				if err != nil {
					t.Fatal(err)
				}
				defer nested.Rollback(context.Background())
				if _, err := nested.Exec(ctx, `DELETE FROM evidence WHERE record_id=$1`, id); err != nil {
					t.Fatal(err)
				}
				requireEvidencePortableFailure(t, port.ValidateImportTx(ctx, nested, prepared, input), "source_state_exact")
			})
			t.Run("extra row", func(t *testing.T) {
				empty, emptyInput, _ := prepare(t)
				requireEvidencePortableFailure(t, port.ValidateImportTx(ctx, tx, empty, emptyInput), "source_state_exact")
			})
			for _, variant := range []string{"missing", "duplicate", "table", "row", "column", "source actor", "local actor", "extra row", "extra column"} {
				t.Run("attribution_"+variant, func(t *testing.T) {
					original := slices.Clone(attributions.rows)
					defer func() { attributions.rows = original }()
					switch variant {
					case "missing":
						attributions.rows = nil
					case "duplicate":
						attributions.rows = append(attributions.rows, attributions.rows[0])
					case "table":
						attributions.rows[0].SourceTable = "records"
					case "row":
						attributions.rows[0].SourceRowID = uuid.NewString()
					case "column":
						attributions.rows[0].SourceColumn = "created_by_user_id"
					case "source actor":
						attributions.rows[0].SourceActorID = uuid.NewString()
					case "local actor":
						attributions.rows[0].LocalUserID = other
					case "extra row":
						extra := attributions.rows[0]
						extra.SourceRowID = uuid.NewString()
						attributions.rows = append(attributions.rows, extra)
					case "extra column":
						extra := attributions.rows[0]
						extra.SourceColumn = "unexpected_user_id"
						attributions.rows = append(attributions.rows, extra)
					}
					requireEvidencePortableFailure(t, port.ValidateImportTx(ctx, tx, prepared, input), "attribution_exact")
				})
			}
		})
	}
	for _, kind := range []string{"missing", "host"} {
		t.Run("envelope_"+kind, func(t *testing.T) {
			tx := begin(t)
			row := evidencePortableRow(incident, 1)
			if kind != "missing" {
				seed(t, tx, row, kind)
			}
			prepared, input, attribution := prepare(t, row)
			requireEvidencePortableFailure(t, port.ApplyImportTx(ctx, tx, prepared, input), "envelope_type_scope")
			var count int
			if err := tx.QueryRow(ctx, `SELECT count(*) FROM evidence WHERE incident_id=$1`, incident).Scan(&count); err != nil || count != 0 || len(attribution.rows) != 0 {
				t.Fatalf("failed preflight wrote rows=%d attribution=%v err=%v", count, attribution.rows, err)
			}
		})
	}
}

func requireEvidenceOwnerEmpty(t testing.TB, ctx context.Context, tx pgx.Tx, incident uuid.UUID, attributions *evidencePortableAttributions) {
	t.Helper()
	var records, blobs, custody int
	if err := tx.QueryRow(ctx, `SELECT
 (SELECT count(*) FROM evidence WHERE incident_id=$1),
 (SELECT count(*) FROM object_blobs WHERE incident_id=$1),
 (SELECT count(*) FROM evidence_custody_events WHERE incident_id=$1)`, incident).Scan(&records, &blobs, &custody); err != nil {
		t.Fatal(err)
	}
	if records != 0 || blobs != 0 || custody != 0 || len(attributions.rows) != 0 {
		t.Fatalf("rejected preflight changed owner: evidence=%d blobs=%d custody=%d attributions=%#v", records, blobs, custody, attributions.rows)
	}
}
