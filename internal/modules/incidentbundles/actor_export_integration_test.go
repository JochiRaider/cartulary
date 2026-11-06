package incidentbundles

import (
	"context"
	"errors"
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
	if _, err := db.Exec(ctx, `
INSERT INTO incident_bundle_imported_actors (
    incident_id, source_actor_id, display_name, email_hint, local_user_id
) VALUES ($1, $2, 'Imported Actor', 'imported-actor@example.test', NULL)
`, incidentID, importedActorID.String()); err != nil {
		t.Fatalf("seed imported actor descriptor: %v", err)
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
		assertCount := func(query incidentportability.Queryer, expected int) {
			t.Helper()
			var count int
			if err := query.QueryRow(ctx, `SELECT count(*) FROM incident_bundle_imported_attributions WHERE incident_id=$1`, incidentID).Scan(&count); err != nil || count != expected {
				t.Fatalf("historical attribution count=%d, want %d: %v", count, expected, err)
			}
		}
		// No-op assignments and non-attribution edits retain the historical owner.
		exec(`UPDATE incidents SET updated_by_user_id=updated_by_user_id WHERE id=$1`, incidentID)
		exec(`UPDATE saved_views SET owner_user_id=owner_user_id, display_name='Renamed' WHERE saved_view_id=$1`, viewID)
		assertCount(db, 3)
		tx, err := db.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			t.Fatal(err)
		}
		defer tx.Rollback(ctx)
		if _, err := tx.Exec(ctx, `UPDATE incidents SET incident_version=incident_version+1 WHERE id=$1`, incidentID); err != nil {
			t.Fatal(err)
		}
		assertCount(tx, 2)
		if err := tx.Rollback(ctx); err != nil {
			t.Fatal(err)
		}
		assertCount(db, 3)
		// A real edit by the physical importer retires only the mutable updater.
		exec(`UPDATE incidents SET incident_version=incident_version+1 WHERE id=$1`, incidentID)
		assertCount(db, 2)
		// Assignment away and back must never resurrect the source owner.
		exec(`UPDATE saved_views SET scope='system', owner_user_id=NULL WHERE saved_view_id=$1`, viewID)
		exec(`UPDATE saved_views SET scope='shared', owner_user_id=$2 WHERE saved_view_id=$1`, viewID, localActorID)
		assertCount(db, 1)
		var column string
		if err := db.QueryRow(ctx, `SELECT source_column FROM incident_bundle_imported_attributions WHERE incident_id=$1`, incidentID).Scan(&column); err != nil || column != "created_by_user_id" {
			t.Fatalf("immutable creation attribution lost: %s %v", column, err)
		}
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
}
