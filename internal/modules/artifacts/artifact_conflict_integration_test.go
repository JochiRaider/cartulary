package artifacts_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/conflicttest"
)

func TestArtifactConflictSourceRevalidation(t *testing.T) {
	ctx := context.Background()
	harness := appsupport.StartStore(t, "artifacts-conflict-source-revalidation")
	actor := authstoretest.SeedLocalUserRecord(
		t,
		harness.DB,
		"artifact-conflict-owner@example.test",
		"Artifact Conflict Owner",
		"ArtifactConflictOwner1!",
		false,
		false,
		true,
	)
	incident := appsupport.CreateIncidentInStore(
		t,
		harness.DB,
		actor,
		"txn-artifacts-conflict-incident",
		"IR-ARTIFACTS-CONFLICT",
		"Artifact conflict contract",
	)
	facade := mustArtifactMutationFacade(
		t,
		harness.DB,
		conflicttest.NewCodec("artifacts-conflict"),
	)
	now := time.Date(2026, 7, 30, 18, 0, 0, 0, time.UTC)
	created, err := facade.Create(ctx, artifacts.CreateCommand{
		ActorUserID: actor.ID, IncidentID: incident.ID,
		Request: artifacts.CreateRequest{
			ViewSchemaID: artifacts.NotesViewSchemaID,
			ClientTxnID:  "txn-artifacts-conflict-create",
			Values: artifactTextValues(
				"note.title", "Conflict title",
				"note.body", "Base body",
			),
		},
		RequestHash: []byte("hash-artifacts-conflict-create"),
		RequestID:   "req-artifacts-conflict-create",
		OperationID: artifacts.OperationCreate,
		Now:         now,
	})
	if err != nil {
		t.Fatalf("create conflict source: %v", err)
	}

	serverBody := artifactTextValue("Server body")
	serverPatch := artifacts.PatchCommand{
		ActorUserID: actor.ID, RecordID: created.RecordID,
		Request: artifacts.PatchRequest{
			ViewSchemaID:   artifacts.NotesViewSchemaID,
			BaseRowVersion: 1,
			ClientTxnID:    "txn-artifacts-conflict-server",
			Changes: []artifacts.PatchChange{{
				FieldKey: "note.body", Value: &serverBody,
			}},
		},
		RequestHash:         []byte("hash-artifacts-conflict-server"),
		RequestID:           "req-artifacts-conflict-server",
		OperationID:         artifacts.OperationPatch,
		ConflictOperationID: artifacts.OperationConflictResolve,
		Now:                 now.Add(time.Minute),
	}
	if result, err := facade.Patch(ctx, serverPatch); err != nil || result.RowVersion != 2 {
		t.Fatalf("server-side patch = %#v, %v; want row version 2", result, err)
	}

	clientBody := artifactTextValue("Client body")
	stalePatch := artifacts.PatchCommand{
		ActorUserID: actor.ID, RecordID: created.RecordID,
		Request: artifacts.PatchRequest{
			ViewSchemaID:   artifacts.NotesViewSchemaID,
			BaseRowVersion: 1,
			ClientTxnID:    "txn-artifacts-conflict-stale",
			Changes: []artifacts.PatchChange{{
				FieldKey: "note.body", Value: &clientBody,
			}},
		},
		RequestHash:         []byte("hash-artifacts-conflict-stale"),
		RequestID:           "req-artifacts-conflict-stale",
		OperationID:         artifacts.OperationPatch,
		ConflictOperationID: artifacts.OperationConflictResolve,
		Now:                 now.Add(2 * time.Minute),
	}
	beforeChangeSets := artifactContractCount(t, harness, `SELECT count(*) FROM change_sets WHERE incident_id = $1`, incident.ID)
	_, err = facade.Patch(ctx, stalePatch)
	var conflict *artifacts.SameFieldConflictError
	if !errors.As(err, &conflict) {
		t.Fatalf("stale same-field patch error = %v, want SameFieldConflictError", err)
	}
	if conflict.Conflict.ConflictToken == "" ||
		conflict.Conflict.FieldKey != "note.body" ||
		conflict.Conflict.CurrentRowVersion < 1 {
		t.Fatalf("same-field conflict payload is incomplete: %#v", conflict.Conflict)
	}
	requireCount(t, harness, `SELECT count(*) FROM records WHERE record_id = $1 AND row_version = 2`, created.RecordID, 1)
	requireCount(t, harness, `SELECT count(*) FROM record_revisions WHERE record_id = $1`, created.RecordID, 2)
	if got := artifactContractCount(t, harness, `SELECT count(*) FROM change_sets WHERE incident_id = $1`, incident.ID); got != beforeChangeSets {
		t.Fatalf("stale conflict changed change-set count: got %d want %d", got, beforeChangeSets)
	}

	title := artifactTextValue("Rebased title")
	rebased, err := facade.Patch(ctx, artifacts.PatchCommand{
		ActorUserID: actor.ID, RecordID: created.RecordID,
		Request: artifacts.PatchRequest{
			ViewSchemaID:   artifacts.NotesViewSchemaID,
			BaseRowVersion: 1,
			ClientTxnID:    "txn-artifacts-conflict-rebase",
			Changes: []artifacts.PatchChange{{
				FieldKey: "note.title", Value: &title,
			}},
		},
		RequestHash:         []byte("hash-artifacts-conflict-rebase"),
		RequestID:           "req-artifacts-conflict-rebase",
		OperationID:         artifacts.OperationPatch,
		ConflictOperationID: artifacts.OperationConflictResolve,
		Now:                 now.Add(3 * time.Minute),
	})
	if err != nil || rebased.RowVersion != 3 {
		t.Fatalf("different-field stale patch = %#v, %v; want row version 3", rebased, err)
	}

	wrongView := stalePatch
	wrongView.Request.ViewSchemaID = artifacts.FindingsViewSchemaID
	wrongView.Request.ClientTxnID = "txn-artifacts-conflict-wrong-view"
	wrongView.RequestHash = []byte("hash-artifacts-conflict-wrong-view")
	if _, err := facade.Patch(ctx, wrongView); err == nil {
		t.Fatal("conflict source revalidation admitted a mismatched artifact view")
	}
	requireCount(t, harness, `SELECT count(*) FROM records WHERE record_id = $1 AND row_version = 3`, created.RecordID, 1)
}
