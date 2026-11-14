package artifacts_test

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/conflicttest"
)

func TestArtifactPatchEnvelopeOutcomes(t *testing.T) {
	ctx := context.Background()
	harness := appsupport.StartStore(t, "artifacts-patch-envelope-outcomes")
	actor := authstoretest.SeedActor(t, harness.DB, authstoretest.ActorAttributes{Email: "artifact-envelope-owner@example.test", DisplayName: "Artifact Envelope Owner"})
	incident := appsupport.CreateIncidentInStore(
		t,
		harness.DB,
		actor,
		"txn-artifacts-envelope-incident",
		"IR-ARTIFACTS-ENVELOPE",
		"Artifact envelope outcomes",
	)
	now := time.Date(2026, 7, 30, 18, 0, 0, 0, time.UTC)
	facade := mustArtifactMutationFacade(
		t,
		harness.DB,
		conflicttest.NewCodec("artifacts-envelope-outcomes"),
	)

	create := func(t testing.TB, suffix string) artifacts.MutationResult {
		t.Helper()
		result, err := facade.Create(ctx, artifacts.CreateCommand{
			ActorUserID: actor.ID, IncidentID: incident.ID,
			Admission: mustArtifactCreateAdmission(
				t, artifacts.NotesViewSchemaID, "txn-artifacts-envelope-create-"+suffix,
				artifactTextValues("note.title", "Envelope "+suffix), nil,
			),
			RequestID: "req-artifacts-envelope-create-" + suffix,
			Now:       now,
		})
		if err != nil {
			t.Fatalf("create %s envelope fixture: %v", suffix, err)
		}
		requireCount(t, harness, `SELECT count(*) FROM change_sets WHERE incident_id = $1 AND request_id = $2`, incident.ID, "req-artifacts-envelope-create-"+suffix, 1)
		return result
	}
	patch := func(recordID uuid.UUID, baseVersion int64, suffix string) (artifacts.MutationResult, error) {
		return facade.Patch(ctx, artifacts.PatchCommand{
			ActorUserID: actor.ID, RecordID: recordID,
			Admission: mustArtifactPatchAdmission(
				t, artifacts.NotesViewSchemaID, baseVersion, "txn-artifacts-envelope-patch-"+suffix,
				[]map[string]any{artifactValueChange("note.title", "Patched "+suffix)},
			),
			RequestID: "req-artifacts-envelope-patch-" + suffix,
			Now:       now.Add(time.Minute),
		})
	}

	t.Run("optional Note text clears through ordinary history", func(t *testing.T) {
		created, err := facade.Create(ctx, artifacts.CreateCommand{
			ActorUserID: actor.ID, IncidentID: incident.ID,
			Admission: mustArtifactCreateAdmission(t, artifacts.NotesViewSchemaID, "note-clear-create", artifactTextValues("note.title", "Title", "note.body", "Body"), nil),
			RequestID: "note-clear-create", Now: now,
		})
		if err != nil {
			t.Fatal(err)
		}
		clear := func(field, txn string, version int64, value any) (artifacts.MutationResult, error) {
			return facade.Patch(ctx, artifacts.PatchCommand{
				ActorUserID: actor.ID, RecordID: created.RecordID,
				Admission: mustArtifactPatchAdmission(t, artifacts.NotesViewSchemaID, version, txn, []map[string]any{artifactValueChange(field, value)}),
				RequestID: txn, Now: now.Add(time.Duration(version) * time.Minute),
			})
		}
		var first artifacts.MutationResult
		for index, field := range []string{"note.title", "note.body"} {
			version := int64(index + 2)
			var raw any
			var body any = "Body"
			if index == 1 {
				raw, body = " \u2003 ", nil
			}
			result, err := clear(field, "clear-"+field, version-1, raw)
			if err != nil || result.Outcome != artifacts.MutationOutcomeUpdated || result.RowVersion != version || result.ChangeSetID == nil || !slices.Equal(result.ChangedFieldKeys, []string{field, "note.updated_at"}) {
				t.Fatalf("clear result=%#v error=%v", result, err)
			}
			if index == 0 {
				first = result
			}
			var exact bool
			if err := harness.DB.QueryRow(ctx, `SELECT r.row_version=$2 AND p.row_version=$2
AND a.title IS NULL AND p.title IS NULL AND a.body IS NOT DISTINCT FROM $3::text AND p.body IS NOT DISTINCT FROM $3::text
FROM records r JOIN artifacts a USING(record_id) JOIN artifact_grid_projection p USING(record_id) WHERE r.record_id=$1`, created.RecordID, version, body).Scan(&exact); err != nil || !exact {
				t.Fatalf("source/projection clear exact=%v error=%v", exact, err)
			}
			var oldTitle any = "Title"
			if index == 1 {
				oldTitle = nil
			}
			requireCount(t, harness, `SELECT count(*) FROM record_revisions r JOIN change_set_mutations m USING(change_set_id)
WHERE r.change_set_id=$1 AND r.record_id=$2 AND r.row_version=$3 AND m.target_kind='record' AND m.target_id=$2::text AND m.operation_kind='patch'
AND r.before_json->'source'->'title'=coalesce(to_jsonb($4::text),'null'::jsonb) AND r.before_json->'source'->'body'='"Body"'::jsonb
AND r.after_json->'source'->'title'='null'::jsonb AND r.after_json->'source'->'body'=coalesce(to_jsonb($5::text),'null'::jsonb)
AND m.before_value=r.before_json AND m.after_value=r.after_json`, result.ChangeSetID, created.RecordID, version, oldTitle, body, 1)
			requireArtifactRecordChangedIntent(t, harness, result, actor.ID)
			intent := collaborationsupport.LoadLatestIntent(t, harness.DB, collaborationsupport.IntentSelector{
				EventFamily: "record_changed", SourceChangeSetID: result.ChangeSetID.String(), SourceRecordID: created.RecordID.String(),
			})
			var published struct {
				Views []struct {
					Patch struct {
						Cells map[string]map[string]json.RawMessage `json:"cells"`
					} `json:"patch_cells"`
				} `json:"affected_views"`
			}
			if err := json.Unmarshal(intent.CanonicalPayload, &published); err != nil || len(published.Views) != 1 {
				t.Fatalf("clear publication shape=%#v error=%v", published, err)
			}
			cells := published.Views[0].Patch.Cells
			expectedUpdatedAt, err := json.Marshal(now.Add(time.Duration(version-1) * time.Minute).Format(time.RFC3339Nano))
			if err != nil {
				t.Fatal(err)
			}
			if len(cells) != 2 || string(cells[field]["value"]) != "null" || string(cells["note.updated_at"]["value"]) != string(expectedUpdatedAt) {
				t.Fatalf("clear publication must include explicit null for %s: %#v", field, cells)
			}
		}
		publication := collaborationsupport.SnapshotPublicationState(t, harness.DB, incident.ID)
		replay, err := clear("note.title", "clear-note.title", 1, "")
		if err != nil || replay.Outcome != artifacts.MutationOutcomeReplayed || replay.ChangeSetID == nil || *replay.ChangeSetID != *first.ChangeSetID || replay.RowVersion != 2 {
			t.Fatalf("normalized clear replay=%#v error=%v", replay, err)
		}
		_, err = clear("note.body", "clear-note-no-effect", 3, nil)
		var invalid *artifacts.ValidationError
		if !errors.As(err, &invalid) || invalid.Field != "changes" || invalid.ReasonCode != "no_effective_change" {
			t.Fatalf("already-null patch error=%v", err)
		}
		requireCount(t, harness, `SELECT count(*) FROM records r JOIN artifacts a USING(record_id) WHERE r.record_id=$1 AND r.row_version=3 AND a.title IS NULL AND a.body IS NULL`, created.RecordID, 1)
		requireCount(t, harness, `SELECT count(*) FROM record_revisions WHERE record_id=$1`, created.RecordID, 3)
		requireCount(t, harness, `SELECT count(*) FROM change_sets WHERE incident_id=$1 AND client_txn_id='clear-note-no-effect'`, incident.ID, 0)
		if !reflect.DeepEqual(publication, collaborationsupport.SnapshotPublicationState(t, harness.DB, incident.ID)) {
			t.Fatal("replay/no-effect clear changed publication state")
		}
	})

	t.Run("future base version fails with authoritative current version", func(t *testing.T) {
		created := create(t, "stale")
		_, err := patch(created.RecordID, 2, "stale")
		var conflict *artifacts.RowVersionConflictError
		if !errors.As(err, &conflict) || conflict.CurrentRowVersion != 1 || conflict.BaseRowVersion != 2 {
			t.Fatalf("future base-version error = %#v, %v", conflict, err)
		}
		requireCount(t, harness, `SELECT count(*) FROM record_revisions WHERE record_id = $1`, created.RecordID, 1)
		requireCount(t, harness, `SELECT count(*) FROM change_sets WHERE incident_id = $1 AND request_id = $2`, incident.ID, "req-artifacts-envelope-patch-stale", 0)
	})

	t.Run("deleted record requires restore", func(t *testing.T) {
		created := create(t, "deleted")
		if _, err := harness.DB.Exec(ctx, `
UPDATE records
   SET deleted_at = $2,
       deleted_by_user_id = $3
 WHERE record_id = $1
`, created.RecordID, now.Add(30*time.Second), actor.ID); err != nil {
			t.Fatalf("mark artifact envelope deleted: %v", err)
		}
		if _, err := patch(created.RecordID, 1, "deleted"); !errors.Is(err, revisions.ErrRecordDeletedUseRestore) {
			t.Fatalf("deleted artifact patch error = %v, want ErrRecordDeletedUseRestore", err)
		}
		requireCount(t, harness, `SELECT count(*) FROM record_revisions WHERE record_id = $1`, created.RecordID, 1)
		requireCount(t, harness, `SELECT count(*) FROM change_sets WHERE incident_id = $1 AND request_id = $2`, incident.ID, "req-artifacts-envelope-patch-deleted", 0)
	})

	t.Run("wrong record type remains concealed", func(t *testing.T) {
		recordID := seedArtifactContractRecord(t, harness, incident.ID, actor.ID, "host", now)
		if _, err := patch(recordID, 1, "wrong-type"); !errors.Is(err, pgx.ErrNoRows) {
			t.Fatalf("wrong-type artifact patch error = %v, want pgx.ErrNoRows", err)
		}
		requireCount(t, harness, `SELECT count(*) FROM change_sets WHERE incident_id = $1 AND request_id = $2`, incident.ID, "req-artifacts-envelope-patch-wrong-type", 0)
	})
}
