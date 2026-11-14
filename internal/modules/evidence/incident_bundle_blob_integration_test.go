package evidence_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"reflect"
	"slices"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	"github.com/JochiRaider/cartulary/internal/testutil/portabilitysupport"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
)

func portableBlobRewriteFixture(t testing.TB, rows []map[string]any) portabilitysupport.MapBundle {
	t.Helper()
	b := evidencePortableBundle(t)
	var out bytes.Buffer
	for _, source := range rows {
		row := maps.Clone(source)
		row["storage_key"] = "incidents/" + row["incident_id"].(string) + "/object-blobs/" + row["object_blob_id"].(string) + "/objects/30000000-0000-4000-8000-000000000001"
		if err := json.NewEncoder(&out).Encode(row); err != nil {
			t.Fatal(err)
		}
	}
	b[portableBlobTestPath] = out.Bytes()
	return b
}
func TestEvidencePortableBlobTransaction_Integration(t *testing.T) {
	ctx := t.Context()
	db := pgtest.Start(t).BeginRollbackDBT(t, "evidence-blob-port")
	actor, other, incident, otherIncident := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	for _, id := range []uuid.UUID{actor, other} {
		if _, err := db.Exec(ctx, `INSERT INTO users(id,email,display_name,password_hash,mfa_required,is_active,is_deployment_admin) VALUES($1,$2,'Blob importer','test-only',false,true,true)`, id, id.String()+"@example.test"); err != nil {
			t.Fatal(err)
		}
	}
	for i, id := range []uuid.UUID{incident, otherIncident} {
		if _, err := db.Exec(ctx, `INSERT INTO incidents(id,incident_key,incident_key_canonical,title,status,created_by_user_id,updated_by_user_id) VALUES($1,$2,$2,'Blob portability','active',$3,$3)`, id, fmt.Sprintf("BLOB-PORT-%d", i), actor); err != nil {
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
	prepare := func(t testing.TB, rows ...map[string]any) (sourceport.Prepared, sourceport.ImportContext, *evidencePortableAttributions) {
		t.Helper()
		input := evidencePortableInput(t, incident, actor)
		p, err := port.PrepareImport(ctx, portableBlobTestBundle(t, rows...), input)
		if err != nil {
			t.Fatal(err)
		}
		a := &evidencePortableAttributions{local: actor}
		input.Attributions = a
		input.RewrittenObjectBlobs = portableBlobRewriteFixture(t, rows)
		return p, input, a
	}
	for _, variant := range []string{"pending", "failed", "available", "quarantined", "null", "129 rows", "empty"} {
		t.Run(variant, func(t *testing.T) {
			tx := begin(t)
			count := 1
			state := variant
			if variant == "null" || variant == "129 rows" || variant == "empty" {
				state = "pending"
			}
			if variant == "129 rows" {
				count = 129
			}
			if variant == "empty" {
				count = 0
			}
			rows := []map[string]any{}
			for i := 1; i <= count; i++ {
				row := portableBlobTestRow(incident, i, state)
				if variant == "null" {
					row["filename_hint"], row["content_type_hint"] = nil, nil
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
			var storedCount int
			if err := tx.QueryRow(ctx, `SELECT count(*) FROM object_blobs WHERE incident_id=$1`, incident).Scan(&storedCount); err != nil || storedCount != count {
				t.Fatalf("stored=%d want=%d err=%v", storedCount, count, err)
			}
			expectedAttrs := []incidentportability.ImportedAttribution{}
			for _, source := range rows {
				expected := maps.Clone(source)
				expected["created_by_user_id"] = actor.String()
				expected["storage_key"] = "incidents/" + incident.String() + "/object-blobs/" + source["object_blob_id"].(string) + "/objects/30000000-0000-4000-8000-000000000001"
				var raw []byte
				if err := tx.QueryRow(ctx, `SELECT to_jsonb(b) FROM object_blobs b WHERE object_blob_id=$1`, source["object_blob_id"]).Scan(&raw); err != nil {
					t.Fatal(err)
				}
				var actual map[string]any
				decoder := json.NewDecoder(bytes.NewReader(raw))
				decoder.UseNumber()
				if err := decoder.Decode(&actual); err != nil {
					t.Fatal(err)
				}
				if !reflect.DeepEqual(actual, expected) {
					t.Fatalf("stored blob=%#v want=%#v", actual, expected)
				}
				expectedAttrs = append(expectedAttrs, incidentportability.ImportedAttribution{SourceTable: "object_blobs", SourceRowID: source["object_blob_id"].(string), SourceColumn: "created_by_user_id", SourceActorID: evidencePortableActor.String(), LocalUserID: actor})
			}
			if !reflect.DeepEqual(attrs.rows, expectedAttrs) && !(len(attrs.rows) == 0 && len(expectedAttrs) == 0) {
				t.Fatalf("attributions=%#v want=%#v", attrs.rows, expectedAttrs)
			}
			var historicalUser bool
			if err := tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE id=$1)`, evidencePortableActor).Scan(&historicalUser); err != nil || historicalUser {
				t.Fatalf("historical user created=%v err=%v", historicalUser, err)
			}
			exported, err := portabilitysupport.Export(port, ctx, sourceport.ExportContext{Query: tx, IncidentID: incident, PortableAttributions: attrs})
			if err != nil {
				t.Fatal(err)
			}
			exportedBundle := portabilitysupport.MapBundle{}
			for _, file := range exported {
				exportedBundle[file.Path] = file.Payload
			}
			actualRows := []map[string]any{}
			if err := incidentportability.WalkPath(ctx, exportedBundle, portableBlobTestPath, func(row map[string]any) error { actualRows = append(actualRows, row); return nil }); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(actualRows, rows) {
				t.Fatalf("re-export=%#v want=%#v", actualRows, rows)
			}
			if count > 0 {
				before := slices.Clone(attrs.rows)
				requireEvidencePortableFailure(t, port.ApplyImportTx(ctx, tx, p, input), "source_identity_admitted")
				if !reflect.DeepEqual(before, attrs.rows) {
					t.Fatal("duplicate changed attribution")
				}
				if err := port.ValidateImportTx(ctx, tx, p, input); err != nil {
					t.Fatal(err)
				}
			}
		})
	}
	t.Run("terminal reason alone", func(t *testing.T) {
		tx := begin(t)
		source := portableBlobTestRow(incident, 1, "failed")
		p, input, _ := prepare(t, source)
		if err := port.ApplyImportTx(ctx, tx, p, input); err != nil {
			t.Fatal(err)
		}
		if _, err := tx.Exec(ctx, `UPDATE object_blobs SET terminal_reason='declared_size_mismatch' WHERE object_blob_id=$1`, source["object_blob_id"]); err != nil {
			t.Fatal(err)
		}
		requireEvidencePortableFailure(t, port.ValidateImportTx(ctx, tx, p, input), "source_state_exact")
	})
	t.Run("complete stored relation and sidecars", func(t *testing.T) {
		tx := begin(t)
		source := portableBlobTestRow(incident, 1, "available")
		p, input, attrs := prepare(t, source)
		if err := port.ApplyImportTx(ctx, tx, p, input); err != nil {
			t.Fatal(err)
		}
		changes := []struct {
			field string
			value any
		}{
			{"object_blob_id", uuid.New()}, {"incident_id", otherIncident}, {"created_by_user_id", other}, {"storage_key", "incidents/" + incident.String() + "/object-blobs/" + source["object_blob_id"].(string) + "/objects/40000000-0000-4000-8000-000000000001"},
			{"upload_state", "quarantined"}, {"byte_size", 15}, {"filename_hint", "changed.txt"}, {"content_type_hint", "changed/type"}, {"expected_sha256_hex", strings.Repeat("0", 64)},
			{"observed_size", 15}, {"observed_content_type", "changed/observed"}, {"observed_sha256_hex", strings.Repeat("0", 64)},
			{"target_expires_at", "2026-01-01T02:00:00Z"}, {"pending_expires_at", "2026-01-03T00:00:00Z"}, {"finalized_at", "2026-01-01T00:30:00.000001Z"},
			{"failed_at", "2026-01-02T00:00:00Z"}, {"cleanup_due_at", "2026-01-02T00:45:00Z"}, {"cleaned_up_at", "2026-01-02T01:00:00Z"}, {"created_at", "2025-01-01T00:00:00Z"}, {"updated_at", "2027-01-01T00:00:00Z"},
		}
		for _, change := range changes {
			t.Run(change.field, func(t *testing.T) {
				nested, err := tx.Begin(ctx)
				if err != nil {
					t.Fatal(err)
				}
				defer nested.Rollback(context.Background())
				if _, err := nested.Exec(ctx, "UPDATE object_blobs SET "+change.field+"=$2 WHERE object_blob_id=$1", source["object_blob_id"], change.value); err != nil {
					t.Fatal(err)
				}
				requireEvidencePortableFailure(t, port.ValidateImportTx(ctx, nested, p, input), "source_state_exact")
			})
		}
		t.Run("missing", func(t *testing.T) {
			nested, err := tx.Begin(ctx)
			if err != nil {
				t.Fatal(err)
			}
			defer nested.Rollback(context.Background())
			if _, err := nested.Exec(ctx, `DELETE FROM object_blobs WHERE object_blob_id=$1`, source["object_blob_id"]); err != nil {
				t.Fatal(err)
			}
			requireEvidencePortableFailure(t, port.ValidateImportTx(ctx, nested, p, input), "source_state_exact")
		})
		t.Run("extra", func(t *testing.T) {
			empty, emptyInput, _ := prepare(t)
			if err := port.ApplyImportTx(ctx, tx, empty, emptyInput); err != nil {
				t.Fatal(err)
			}
			requireEvidencePortableFailure(t, port.ValidateImportTx(ctx, tx, empty, emptyInput), "source_state_exact")
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
					attrs.rows[0].SourceTable = "evidence"
				case "row":
					attrs.rows[0].SourceRowID = uuid.NewString()
				case "column":
					attrs.rows[0].SourceColumn = "owner_user_id"
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
					a.SourceColumn = "unknown_user_id"
					attrs.rows = append(attrs.rows, a)
				}
				requireEvidencePortableFailure(t, port.ValidateImportTx(ctx, tx, p, input), "attribution_exact")
			})
		}
		t.Run("rewritten generation cannot be replaced", func(t *testing.T) {
			changed := input
			replacement := portableBlobRewriteFixture(t, []map[string]any{source})
			replacement[portableBlobTestPath] = bytes.ReplaceAll(replacement[portableBlobTestPath], []byte("30000000-0000-4000-8000-000000000001"), []byte("40000000-0000-4000-8000-000000000001"))
			changed.RewrittenObjectBlobs = replacement
			if err := port.ValidateImportTx(ctx, tx, p, changed); !errors.Is(err, sourceport.ErrPreparedBinding) {
				t.Fatalf("rebound=%v", err)
			}
		})
	})
	t.Run("rewritten input checked before writes", func(t *testing.T) {
		source := portableBlobTestRow(incident, 1, "pending")
		for _, tc := range []struct {
			name, invariant string
			mutate          func(map[string]any)
		}{
			{"absent key", "storage_reference_legal", func(r map[string]any) { delete(r, "storage_key") }},
			{"bad key", "storage_reference_legal", func(r map[string]any) { r["storage_key"] = "../outside" }},
			{"foreign incident key", "storage_reference_legal", func(r map[string]any) {
				r["storage_key"] = "incidents/" + otherIncident.String() + "/object-blobs/" + source["object_blob_id"].(string) + "/objects/30000000-0000-4000-8000-000000000001"
			}},
			{"nil generation", "storage_reference_legal", func(r map[string]any) {
				r["storage_key"] = "incidents/" + incident.String() + "/object-blobs/" + source["object_blob_id"].(string) + "/objects/" + uuid.Nil.String()
			}},
			{"changed creator", "source_state_exact", func(r map[string]any) { r["created_by_user_id"] = actor.String() }},
			{"missing source field", "source_state_exact", func(r map[string]any) { delete(r, "filename_hint") }},
			{"unknown source field", "source_state_exact", func(r map[string]any) { r["extra"] = true }},
			{"changed source value", "source_state_exact", func(r map[string]any) { r["filename_hint"] = "other.txt" }},
		} {
			t.Run(tc.name, func(t *testing.T) {
				p, input, attrs := prepare(t, source)
				b := input.RewrittenObjectBlobs.(portabilitysupport.MapBundle)
				var row map[string]any
				dec := json.NewDecoder(bytes.NewReader(b[portableBlobTestPath]))
				dec.UseNumber()
				if err := dec.Decode(&row); err != nil {
					t.Fatal(err)
				}
				tc.mutate(row)
				raw, err := json.Marshal(row)
				if err != nil {
					t.Fatal(err)
				}
				b[portableBlobTestPath] = append(raw, '\n')
				requireEvidencePortableFailure(t, port.ApplyImportTx(ctx, forbiddenEvidencePortableTx{}, p, input), tc.invariant)
				if len(attrs.rows) != 0 {
					t.Fatal("rejected rewrite recorded attribution")
				}
			})
		}
		for _, variant := range []string{"missing row", "extra row", "duplicate row", "malformed"} {
			t.Run(variant, func(t *testing.T) {
				p, input, _ := prepare(t, source)
				b := input.RewrittenObjectBlobs.(portabilitysupport.MapBundle)
				switch variant {
				case "missing row":
					b[portableBlobTestPath] = nil
				case "extra row":
					extra := portableBlobRewriteFixture(t, []map[string]any{portableBlobTestRow(incident, 2, "pending")})
					b[portableBlobTestPath] = append(bytes.Clone(b[portableBlobTestPath]), extra[portableBlobTestPath]...)
				case "duplicate row":
					b[portableBlobTestPath] = append(bytes.Clone(b[portableBlobTestPath]), b[portableBlobTestPath]...)
				case "malformed":
					b[portableBlobTestPath] = []byte("{\n")
				}
				requireEvidencePortableFailure(t, port.ApplyImportTx(ctx, forbiddenEvidencePortableTx{}, p, input), "source_state_exact")
			})
		}
	})
}
