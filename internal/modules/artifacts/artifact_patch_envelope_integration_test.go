package artifacts_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/conflicttest"
)

func TestArtifactPatchEnvelopeOutcomes(t *testing.T) {
	ctx := context.Background()
	harness := appsupport.StartStore(t, "artifacts-patch-envelope-outcomes")
	actor := authstoretest.SeedLocalUserRecord(
		t,
		harness.DB,
		"artifact-envelope-owner@example.test",
		"Artifact Envelope Owner",
		"ArtifactEnvelopeOwner1!",
		false,
		false,
		true,
	)
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
			Request: artifacts.CreateRequest{
				ViewSchemaID: artifacts.NotesViewSchemaID,
				ClientTxnID:  "txn-artifacts-envelope-create-" + suffix,
				Values:       artifactTextValues("note.title", "Envelope "+suffix),
			},
			RequestHash: []byte("hash-artifacts-envelope-create-" + suffix),
			RequestID:   "req-artifacts-envelope-create-" + suffix,
			OperationID: artifacts.OperationCreate,
			Now:         now,
		})
		if err != nil {
			t.Fatalf("create %s envelope fixture: %v", suffix, err)
		}
		return result
	}
	patch := func(recordID uuid.UUID, baseVersion int64, suffix string) (artifacts.MutationResult, error) {
		value := artifactTextValue("Patched " + suffix)
		return facade.Patch(ctx, artifacts.PatchCommand{
			ActorUserID: actor.ID, RecordID: recordID,
			Request: artifacts.PatchRequest{
				ViewSchemaID:   artifacts.NotesViewSchemaID,
				BaseRowVersion: baseVersion,
				ClientTxnID:    "txn-artifacts-envelope-patch-" + suffix,
				Changes: []artifacts.PatchChange{{
					FieldKey: "note.title", Value: &value,
				}},
			},
			RequestHash:         []byte("hash-artifacts-envelope-patch-" + suffix),
			RequestID:           "req-artifacts-envelope-patch-" + suffix,
			OperationID:         artifacts.OperationPatch,
			ConflictOperationID: artifacts.OperationConflictResolve,
			Now:                 now.Add(time.Minute),
		})
	}

	t.Run("future base version fails with authoritative current version", func(t *testing.T) {
		created := create(t, "stale")
		_, err := patch(created.RecordID, 2, "stale")
		var conflict *artifacts.RowVersionConflictError
		if !errors.As(err, &conflict) || conflict.CurrentRowVersion != 1 || conflict.BaseRowVersion != 2 {
			t.Fatalf("future base-version error = %#v, %v", conflict, err)
		}
		requireCount(t, harness, `SELECT count(*) FROM record_revisions WHERE record_id = $1`, created.RecordID, 1)
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
	})

	t.Run("wrong record type remains concealed", func(t *testing.T) {
		recordID := seedArtifactContractRecord(t, harness, incident.ID, actor.ID, "host", now)
		if _, err := patch(recordID, 1, "wrong-type"); !errors.Is(err, pgx.ErrNoRows) {
			t.Fatalf("wrong-type artifact patch error = %v, want pgx.ErrNoRows", err)
		}
		requireCount(t, harness, `SELECT count(*) FROM change_sets WHERE incident_id = $1 AND source = 'workbook.rows.patch'`, incident.ID, 0)
	})
}
