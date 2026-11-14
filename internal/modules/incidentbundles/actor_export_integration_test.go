package incidentbundles

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestActorExportCombinesNativeAndImportedDescriptorsAndRejectsCollisions_Integration(t *testing.T) {
	ctx := context.Background()
	db := pgtest.Start(t).BeginRollbackDBT(t, "incident-bundle-actor-export")
	localActorID := uuid.MustParse("00000000-0000-4000-8000-000000110501")
	importedActorID := uuid.MustParse("00000000-0000-4000-8000-000000110502")
	incidentID := uuid.MustParse("00000000-0000-4000-8000-000000110503")
	t.Run("closed actor roster admission", func(t *testing.T) {
		const actorID = "aaaaaaaa-0000-4000-8000-000000000001"
		const row = `{"actor_id":"aaaaaaaa-0000-4000-8000-000000000001","display_name":"Actor"}`
		for _, tc := range []struct{ name, body, invariant string }{
			{"minimal", row, ""},
			{"empty name", strings.Replace(row, `"Actor"`, `""`, 1), ""},
			{"missing identity", `{"display_name":"Actor"}`, "actors.reference_complete"},
			{"null identity", strings.Replace(row, `"`+actorID+`"`, `null`, 1), "actors.reference_complete"},
			{"numeric identity", strings.Replace(row, `"`+actorID+`"`, `1`, 1), "actors.reference_complete"},
			{"padded identity", strings.Replace(row, actorID, " "+actorID, 1), "actors.reference_complete"},
			{"uppercase identity", strings.Replace(row, actorID, strings.ToUpper(actorID), 1), "actors.reference_complete"},
			{"compact identity", strings.Replace(row, actorID, strings.ReplaceAll(actorID, "-", ""), 1), "actors.reference_complete"},
			{"nil identity", strings.Replace(row, actorID, "00000000-0000-0000-0000-000000000000", 1), "actors.reference_complete"},
			{"missing name", `{"actor_id":"` + actorID + `"}`, "actors.reference_complete"},
			{"null name", strings.Replace(row, `"Actor"`, `null`, 1), "actors.reference_complete"},
			{"numeric name", strings.Replace(row, `"Actor"`, `17`, 1), "actors.reference_complete"},
			{"legacy alias", strings.Replace(row, "actor_id", "source_actor_id", 1), "actors.inert"},
			{"differently cased member", strings.Replace(row, "display_name", "Display_Name", 1), "actors.inert"},
			{"login material", strings.TrimSuffix(row, "}") + `,"password_hash":"forbidden"}`, "actors.inert"},
			{"duplicate identity", row + "\n" + row, "actors.reference_complete"},
			{"decreasing identity", strings.Replace(row, actorID, "bbbbbbbb-0000-4000-8000-000000000001", 1) + "\n" + row, "actors.reference_complete"},
		} {
			t.Run(tc.name, func(t *testing.T) {
				members := fixtureMembers(t, map[string][]byte{"data/actors.ndjson": []byte(tc.body + "\n")})
				catalog, err := validateImportedActors(ctx, members, incidentID)
				if tc.invariant != "" {
					var failure *verificationError
					if !errors.As(err, &failure) || failure.ReasonCode != "source_family_invalid" || failure.SourceFamilyID != "actors" || failure.InvariantID != tc.invariant {
						t.Fatalf("actor admission = %v, want %s", err, tc.invariant)
					}
					return
				}
				name := "Actor"
				if tc.name == "empty name" {
					name = ""
				}
				descriptor, exists := catalog.Lookup(actorID)
				if err != nil || catalog.Len() != 1 || !exists || descriptor.SourceActorID != actorID || descriptor.DisplayName != name || descriptor.EmailHint != "" || descriptor.ProviderSubjectHint != "" {
					t.Fatalf("valid actor descriptor = %#v, exists=%v, error=%v", descriptor, exists, err)
				}
			})
		}
		for _, field := range []string{"email_hint", "provider_subject_hint"} {
			for _, value := range []string{"null", `""`, "17", "true", "{}", "[]"} {
				t.Run(field+"/"+value, func(t *testing.T) {
					body := strings.TrimSuffix(row, "}") + `,"` + field + `":` + value + "}\n"
					_, err := validateImportedActors(ctx, fixtureMembers(t, map[string][]byte{"data/actors.ndjson": []byte(body)}), incidentID)
					var failure *verificationError
					if !errors.As(err, &failure) || failure.ReasonCode != "source_family_invalid" || failure.SourceFamilyID != "actors" || failure.InvariantID != "actors.reference_complete" {
						t.Fatalf("invalid hint admission = %v", err)
					}
				})
			}
		}
	})
	if _, err := db.Exec(ctx, `
INSERT INTO users (
    id, email, display_name, password_hash, mfa_required,
    is_active, is_deployment_admin, created_at, updated_at
) VALUES (
    $1, 'native-actor@example.test', 'Native Actor', 'test-only',
    false, true, true, '2026-07-02T09:00:00Z', '2026-07-02T09:00:00Z'
)
`, localActorID); err != nil {
		t.Fatalf("seed native actor: %v", err)
	}
	if _, err := db.Exec(ctx, `
INSERT INTO incidents (
    id, incident_key, incident_key_canonical, title, status,
    created_by_user_id, updated_by_user_id, created_at, updated_at
) VALUES (
    $1, 'ACTOR-EXPORT', 'ACTOR-EXPORT', 'Actor export', 'active',
    $2, $2, '2026-07-02T09:30:00Z', '2026-07-02T09:30:00Z'
)
`, incidentID, localActorID); err != nil {
		t.Fatalf("seed actor export incident: %v", err)
	}
	members := fixtureMembers(t, map[string][]byte{
		"data/actors.ndjson": []byte(`{"actor_id":"00000000-0000-4000-8000-000000110502","display_name":"Imported Actor","email_hint":"imported-actor@example.test","provider_subject_hint":"issuer.example.test/subject-17"}` + "\n"),
	})
	if _, err := validateImportedActors(ctx, members, incidentID); err != nil {
		t.Fatalf("validate actor descriptor: %v", err)
	}
	importTx, err := db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatal(err)
	}
	defer importTx.Rollback(ctx)
	if err := (importer{}).importActors(ctx, importTx, members, incidentID); err != nil {
		t.Fatalf("import actor descriptor: %v", err)
	}
	if err := importTx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	var inert bool
	if err := db.QueryRow(ctx, `SELECT local_user_id IS NULL FROM incident_bundle_imported_actors WHERE incident_id=$1 AND source_actor_id=$2`, incidentID, importedActorID.String()).Scan(&inert); err != nil || !inert {
		t.Fatalf("imported actor inert=%v: %v", inert, err)
	}

	references := newActorReferenceSet()
	for _, id := range []uuid.UUID{localActorID, importedActorID} {
		if err := references.RecordActorReference(id); err != nil {
			t.Fatal(err)
		}
	}
	payload, err := (bundleBuilder{}).exportActors(ctx, db, incidentID, references)
	if err != nil {
		t.Fatalf("export native and imported actors: %v", err)
	}
	rows, err := incidentportability.DecodeNDJSON(payload)
	if err != nil {
		t.Fatalf("decode actor export: %v", err)
	}
	if len(rows) != 2 ||
		rows[0]["actor_id"] != localActorID.String() ||
		rows[1]["actor_id"] != importedActorID.String() {
		t.Fatalf("actor export rows = %#v", rows)
	}
	for index, want := range []struct{ displayName, emailHint, providerHint string }{
		{"Native Actor", "native-actor@example.test", ""},
		{"Imported Actor", "imported-actor@example.test", "issuer.example.test/subject-17"},
	} {
		wantFields := 3
		if want.providerHint != "" {
			wantFields++
			if rows[index]["provider_subject_hint"] != want.providerHint {
				t.Fatalf("actor descriptor %d provider hint = %#v, want %q", index, rows[index]["provider_subject_hint"], want.providerHint)
			}
		}
		if len(rows[index]) != wantFields || rows[index]["display_name"] != want.displayName || rows[index]["email_hint"] != want.emailHint {
			t.Fatalf("actor descriptor %d = %#v, want %d fields, display_name=%q, email_hint=%q", index, rows[index], wantFields, want.displayName, want.emailHint)
		}
	}
	t.Run("empty display name and opaque hints remain explicit", func(t *testing.T) {
		tx, err := db.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			t.Fatal(err)
		}
		defer tx.Rollback(ctx)
		if _, err := tx.Exec(ctx, `DELETE FROM incident_bundle_imported_actors WHERE incident_id=$1 AND source_actor_id=$2`, incidentID, importedActorID.String()); err != nil {
			t.Fatal(err)
		}
		members := fixtureMembers(t, map[string][]byte{
			"data/actors.ndjson": []byte(`{"actor_id":"00000000-0000-4000-8000-000000110502","display_name":"","email_hint":" ","provider_subject_hint":"\t"}` + "\n"),
		})
		if _, err := validateImportedActors(ctx, members, incidentID); err != nil {
			t.Fatal(err)
		}
		if err := (importer{}).importActors(ctx, tx, members, incidentID); err != nil {
			t.Fatal(err)
		}
		payload, err := (bundleBuilder{}).exportActors(ctx, tx, incidentID, references)
		if err != nil {
			t.Fatal(err)
		}
		rows, err := incidentportability.DecodeNDJSON(payload)
		if err != nil || len(rows) != 2 || len(rows[1]) != 4 || rows[1]["actor_id"] != importedActorID.String() || rows[1]["display_name"] != "" || rows[1]["email_hint"] != " " || rows[1]["provider_subject_hint"] != "\t" {
			t.Fatalf("empty-name export = %#v, %v", rows, err)
		}
	})
	t.Run("native mutations preserve immutable attribution and rollback", func(t *testing.T) {
		viewID := uuid.New()
		exec := func(statement string, args ...any) {
			t.Helper()
			if _, err := db.Exec(ctx, statement, args...); err != nil {
				t.Fatal(err)
			}
		}
		exec(`INSERT INTO saved_views(saved_view_id, incident_id, view_schema_id, scope, display_name, query_json, layout_json, owner_user_id)
VALUES ($1,$2,'cartulary.view.timeline.v2','shared','Historical owner','{}','{}',$3)`, viewID, incidentID, localActorID)
		for _, binding := range []struct{ table, row, column string }{
			{"incidents", incidentID.String(), "created_by_user_id"},
			{"incidents", incidentID.String(), "updated_by_user_id"},
			{"saved_views", viewID.String(), "owner_user_id"},
		} {
			exec(`INSERT INTO incident_bundle_imported_attributions(incident_id,source_table,source_row_id,source_column,source_actor_id,local_user_id)
VALUES ($1,$2,$3,$4,$5,$6)`, incidentID, binding.table, binding.row, binding.column, importedActorID.String(), localActorID)
		}
		type attribution [5]string
		creation := attribution{"incidents", incidentID.String(), "created_by_user_id", importedActorID.String(), localActorID.String()}
		updater := attribution{"incidents", incidentID.String(), "updated_by_user_id", importedActorID.String(), localActorID.String()}
		viewOwner := attribution{"saved_views", viewID.String(), "owner_user_id", importedActorID.String(), localActorID.String()}
		assertAttributions := func(query incidentportability.Queryer, expected ...attribution) {
			t.Helper()
			rows, err := query.Query(ctx, `SELECT source_table,source_row_id,source_column,source_actor_id,local_user_id::text
FROM incident_bundle_imported_attributions WHERE incident_id=$1
ORDER BY source_table,source_row_id,source_column`, incidentID)
			if err != nil {
				t.Fatal(err)
			}
			defer rows.Close()
			var actual []attribution
			for rows.Next() {
				var binding attribution
				if err := rows.Scan(&binding[0], &binding[1], &binding[2], &binding[3], &binding[4]); err != nil {
					t.Fatal(err)
				}
				actual = append(actual, binding)
			}
			if err := rows.Err(); err != nil {
				t.Fatal(err)
			}
			if !slices.Equal(actual, expected) {
				t.Fatalf("historical attributions = %#v, want %#v", actual, expected)
			}
		}
		// No-op assignments and non-attribution edits retain the historical owner.
		exec(`UPDATE incidents SET updated_by_user_id=updated_by_user_id WHERE id=$1`, incidentID)
		exec(`UPDATE saved_views SET owner_user_id=owner_user_id, display_name='Renamed' WHERE saved_view_id=$1`, viewID)
		assertAttributions(db, creation, updater, viewOwner)
		tx, err := db.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			t.Fatal(err)
		}
		defer tx.Rollback(ctx)
		if _, err := tx.Exec(ctx, `UPDATE incidents SET incident_version=incident_version+1 WHERE id=$1`, incidentID); err != nil {
			t.Fatal(err)
		}
		assertAttributions(tx, creation, viewOwner)
		if err := tx.Rollback(ctx); err != nil {
			t.Fatal(err)
		}
		assertAttributions(db, creation, updater, viewOwner)
		// A real edit by the physical importer retires only the mutable updater.
		exec(`UPDATE incidents SET incident_version=incident_version+1 WHERE id=$1`, incidentID)
		assertAttributions(db, creation, viewOwner)
		// Assignment away and back must never resurrect the source owner.
		exec(`UPDATE saved_views SET scope='system', owner_user_id=NULL WHERE saved_view_id=$1`, viewID)
		exec(`UPDATE saved_views SET scope='shared', owner_user_id=$2 WHERE saved_view_id=$1`, viewID, localActorID)
		assertAttributions(db, creation)

	})

	if _, err := db.Exec(ctx, `
INSERT INTO incident_bundle_imported_actors (
    incident_id, source_actor_id, display_name, email_hint, local_user_id
) VALUES ($1, $2, 'Conflicting Actor', 'conflict@example.test', NULL)
`, incidentID, localActorID.String()); err != nil {
		t.Fatalf("seed colliding actor descriptor: %v", err)
	}
	_, err = (bundleBuilder{}).exportActors(ctx, db, incidentID, references)
	var failure *verificationError
	if !errors.As(err, &failure) ||
		failure.SourceFamilyID != "actors" ||
		failure.InvariantID != "actors.reference_complete" {
		t.Fatalf("actor descriptor collision error = %T %v", err, err)
	}
	if _, err := db.Exec(ctx, `UPDATE incident_bundle_imported_actors
SET display_name='Native Actor',email_hint='native-actor@example.test',provider_subject_hint='different-provider-subject'
WHERE incident_id=$1 AND source_actor_id=$2`, incidentID, localActorID.String()); err != nil {
		t.Fatal(err)
	}
	_, err = (bundleBuilder{}).exportActors(ctx, db, incidentID, references)
	if !errors.As(err, &failure) || failure.SourceFamilyID != "actors" || failure.InvariantID != "actors.reference_complete" {
		t.Fatalf("provider-hint-only collision error = %T %v", err, err)
	}
}
