package incidentbundles

import (
	"context"
	"errors"
	"testing"

	"github.com/google/uuid"

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

	files := map[string][]byte{
		"data/saved_views.ndjson": []byte(
			`{"owner_user_id":"` + localActorID.String() + `"}` + "\n" +
				`{"owner_user_id":"` + importedActorID.String() + `"}` + "\n" +
				`{"owner_user_id":null}` + "\n",
		),
	}
	payload, err := (bundleBuilder{}).exportActors(ctx, db, incidentID, files)
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

	if _, err := db.Exec(ctx, `
INSERT INTO incident_bundle_imported_actors (
    incident_id, source_actor_id, display_name, email_hint, local_user_id
) VALUES ($1, $2, 'Conflicting Actor', 'conflict@example.test', NULL)
`, incidentID, localActorID.String()); err != nil {
		t.Fatalf("seed colliding actor descriptor: %v", err)
	}
	_, err = (bundleBuilder{}).exportActors(ctx, db, incidentID, files)
	var failure *verificationError
	if !errors.As(err, &failure) ||
		failure.SourceFamilyID != "actors" ||
		failure.InvariantID != "actors.reference_complete" {
		t.Fatalf("actor descriptor collision error = %T %v", err, err)
	}
}
