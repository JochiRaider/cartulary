package artifacts_test

import (
	"context"
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/gen/importtargetregistry"
	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/modules/imports/ownerfacade"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/revisionsupport"
)

func TestArtifactImportCreateFacadeContract(t *testing.T) {
	ctx := context.Background()
	harness := appsupport.StartStore(t, "artifacts-import-create-contract")
	actor := authstoretest.SeedLocalUserRecord(
		t,
		harness.DB,
		"artifact-import-owner@example.test",
		"Artifact Import Owner",
		"ArtifactImportOwner1!",
		false,
		false,
		true,
	)
	incident := appsupport.CreateIncidentInStore(
		t,
		harness.DB,
		actor,
		"txn-artifacts-import-incident",
		"IR-ARTIFACTS-IMPORT",
		"Artifact import contract",
	)
	appender := revisionsupport.MustAppender(t)
	now := time.Date(2026, 7, 30, 17, 0, 0, 0, time.UTC)

	artifactTargets := make(map[string]importtargetregistry.Target)
	for _, target := range importtargetregistry.Targets {
		if target.TargetViewSchemaID == nil || target.OwnerContractRef != "module.artifacts@1" {
			continue
		}
		artifactTargets[*target.TargetViewSchemaID] = target
	}
	if len(artifactTargets) != 8 {
		t.Fatalf("generated artifact import targets = %d, want 8", len(artifactTargets))
	}

	cases := []struct {
		viewSchemaID string
		fields       []ownerfacade.ImportFieldValue
	}{
		{artifacts.NotesViewSchemaID, importTextFields("note.title", "Imported note")},
		{artifacts.CommLogViewSchemaID, importTextFields(
			"comm_log.comm_type", "briefing",
			"comm_log.audience", "responders",
			"comm_log.channel_or_meeting", "bridge",
			"comm_log.summary", "Imported briefing",
		)},
		{artifacts.HandoffViewSchemaID, append(
			importTextFields("handoff.current_state_summary", "Imported handoff"),
			ownerfacade.ImportFieldValue{
				FieldKey:        "handoff.incoming_owner_user_id",
				NormalizedValue: ownerfacade.NewUUIDImportScalar(actor.ID),
			},
		)},
		{artifacts.StatusReviewViewSchemaID, importTextFields("status_review.current_state_summary", "Imported status")},
		{artifacts.LessonViewSchemaID, importTextFields("lesson.summary", "Imported lesson")},
	}

	for index, tc := range cases {
		target := artifactTargets[tc.viewSchemaID]
		if target.AvailabilityKind != "enabled" || target.FacadeID == nil {
			t.Fatalf("%s target = %#v, want enabled owner-create facade", tc.viewSchemaID, target)
		}
		facade, err := artifacts.NewImportContribution(
			tc.viewSchemaID,
			*target.FacadeID,
			artifactImportDependencies(harness.DB, appender),
		)
		if err != nil {
			t.Fatalf("construct %s import facade: %v", tc.viewSchemaID, err)
		}
		tx, err := harness.DB.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			t.Fatalf("begin %s import transaction: %v", tc.viewSchemaID, err)
		}
		changeSetID, err := appender.AppendChangeSetTx(ctx, tx, revisions.AppendChangeSetParams{
			IncidentID: incident.ID, ActorUserID: actor.ID,
			Source:    "imports.unit.apply",
			CreatedAt: now.Add(time.Duration(index) * time.Minute),
		})
		if err != nil {
			_ = tx.Rollback(ctx)
			t.Fatalf("append %s import change set: %v", tc.viewSchemaID, err)
		}
		response, err := facade.CreateImportRowTx(ctx, tx, ownerfacade.ImportOwnerCreateCommand{
			Request: ownerfacade.ImportOwnerCreateRequest{
				IncidentID: incident.ID, ActorUserID: actor.ID,
				TargetViewSchemaID: tc.viewSchemaID,
				ImportSessionID:    uuid.New(),
				ImportUnitID:       uuid.New(),
				MappingFingerprint: fmt.Sprintf("synthetic-map-%02d", index),
				SourceFileKind:     "csv",
				ParserProfileID:    "synthetic",
				ParserVersion:      "1",
				LocatorKind:        "row",
				Locator:            fmt.Sprintf("%d", index+1),
				ClientTxnID:        fmt.Sprintf("txn-artifacts-import-%02d", index),
				FieldValues:        tc.fields,
			},
			ChangeSetID: changeSetID,
			SequenceNo:  1,
			Now:         now.Add(time.Duration(index) * time.Minute),
		})
		if err != nil {
			_ = tx.Rollback(ctx)
			t.Fatalf("create %s import row: %v", tc.viewSchemaID, err)
		}
		if response.RecordID == uuid.Nil || response.RowVersion != 1 ||
			response.CreatedOrReused != "created" || response.OwnerResultCode != "created" ||
			response.ChangeSetMutationRef == "" || response.RowRefresh["record_id"] == nil {
			_ = tx.Rollback(ctx)
			t.Fatalf("%s import response is incomplete: %#v", tc.viewSchemaID, response)
		}
		if err := tx.Commit(ctx); err != nil {
			t.Fatalf("commit %s import transaction: %v", tc.viewSchemaID, err)
		}
		requireCount(t, harness, `SELECT count(*) FROM artifacts WHERE record_id = $1 AND artifact_type = $2`, response.RecordID, artifactTypeForTestView(tc.viewSchemaID), 1)
		requireCount(t, harness, `SELECT count(*) FROM artifact_grid_projection WHERE record_id = $1`, response.RecordID, 1)
		requireCount(t, harness, `SELECT count(*) FROM record_revisions WHERE record_id = $1 AND change_set_id = $2`, response.RecordID, changeSetID, 1)
	}

	for _, viewSchemaID := range []string{
		artifacts.FindingsViewSchemaID,
		artifacts.InvestigativeQueriesViewSchemaID,
		artifacts.ForensicKeywordsViewSchemaID,
	} {
		target := artifactTargets[viewSchemaID]
		if target.AvailabilityKind != "reserved" || target.FacadeID != nil || target.FacadeBindingID != nil {
			t.Fatalf("%s target = %#v, want reserved with no dispatch facade", viewSchemaID, target)
		}
	}

	t.Run("inactive user is rejected before source insertion", func(t *testing.T) {
		inactive := authstoretest.SeedLocalUserRecord(
			t,
			harness.DB,
			"artifact-import-inactive@example.test",
			"Artifact Import Inactive",
			"ArtifactImportInactive1!",
			false,
			false,
			false,
		)
		target := artifactTargets[artifacts.HandoffViewSchemaID]
		facade, err := artifacts.NewImportContribution(
			artifacts.HandoffViewSchemaID,
			*target.FacadeID,
			artifactImportDependencies(harness.DB, appender),
		)
		if err != nil {
			t.Fatalf("construct inactive-user import facade: %v", err)
		}
		before := artifactContractCount(t, harness, `SELECT count(*) FROM records WHERE incident_id = $1`, incident.ID)
		tx, err := harness.DB.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			t.Fatalf("begin inactive-user import transaction: %v", err)
		}
		_, err = facade.CreateImportRowTx(ctx, tx, ownerfacade.ImportOwnerCreateCommand{
			Request: ownerfacade.ImportOwnerCreateRequest{
				IncidentID: incident.ID, ActorUserID: actor.ID,
				TargetViewSchemaID: artifacts.HandoffViewSchemaID,
				ClientTxnID:        "txn-artifacts-import-inactive-user",
				FieldValues: append(
					importTextFields("handoff.current_state_summary", "Rejected handoff"),
					ownerfacade.ImportFieldValue{
						FieldKey:        "handoff.incoming_owner_user_id",
						NormalizedValue: ownerfacade.NewUUIDImportScalar(inactive.ID),
					},
				),
			},
			ChangeSetID: uuid.New(), SequenceNo: 1, Now: now.Add(45 * time.Minute),
		})
		var validation *artifacts.ValidationError
		if !errors.As(err, &validation) || validation.Field != "handoff.incoming_owner_user_id" || validation.ReasonCode != "invalid_value" {
			_ = tx.Rollback(ctx)
			t.Fatalf("inactive-user import error = %#v, %v", validation, err)
		}
		if err := tx.Rollback(ctx); err != nil {
			t.Fatalf("rollback inactive-user import transaction: %v", err)
		}
		requireCount(t, harness, `SELECT count(*) FROM records WHERE incident_id = $1`, incident.ID, before)
	})

	t.Run("injected finalization failure rolls back all source effects", func(t *testing.T) {
		target := artifactTargets[artifacts.NotesViewSchemaID]
		dependencies := artifactImportDependencies(harness.DB, appender)
		dependencies.Revisions = failingArtifactImportAppender{Appender: appender}
		facade, err := artifacts.NewImportContribution(artifacts.NotesViewSchemaID, *target.FacadeID, dependencies)
		if err != nil {
			t.Fatalf("construct failing import facade: %v", err)
		}
		beforeRecords := artifactContractCount(t, harness, `SELECT count(*) FROM records WHERE incident_id = $1`, incident.ID)
		beforeArtifacts := artifactContractCount(t, harness, `SELECT count(*) FROM artifacts WHERE incident_id = $1`, incident.ID)
		tx, err := harness.DB.BeginTx(ctx, pgx.TxOptions{})
		if err != nil {
			t.Fatalf("begin failing import transaction: %v", err)
		}
		changeSetID, err := appender.AppendChangeSetTx(ctx, tx, revisions.AppendChangeSetParams{
			IncidentID: incident.ID, ActorUserID: actor.ID, Source: "imports.unit.apply", CreatedAt: now.Add(50 * time.Minute),
		})
		if err != nil {
			_ = tx.Rollback(ctx)
			t.Fatalf("append failing import change set: %v", err)
		}
		_, err = facade.CreateImportRowTx(ctx, tx, ownerfacade.ImportOwnerCreateCommand{
			Request: ownerfacade.ImportOwnerCreateRequest{
				IncidentID: incident.ID, ActorUserID: actor.ID,
				TargetViewSchemaID: artifacts.NotesViewSchemaID,
				ClientTxnID:        "txn-artifacts-import-injected-failure",
				FieldValues:        importTextFields("note.title", "Rejected after source insertion"),
			},
			ChangeSetID: changeSetID, SequenceNo: 1, Now: now.Add(50 * time.Minute),
		})
		if !errors.Is(err, errInjectedArtifactImportFinalization) {
			_ = tx.Rollback(ctx)
			t.Fatalf("injected finalization error = %v", err)
		}
		var transactionRecords int
		if err := tx.QueryRow(ctx, `SELECT count(*) FROM records WHERE incident_id = $1`, incident.ID).Scan(&transactionRecords); err != nil {
			_ = tx.Rollback(ctx)
			t.Fatalf("count transaction-visible import records: %v", err)
		}
		if transactionRecords != beforeRecords+1 {
			_ = tx.Rollback(ctx)
			t.Fatalf("transaction-visible records = %d, want %d", transactionRecords, beforeRecords+1)
		}
		if err := tx.Rollback(ctx); err != nil {
			t.Fatalf("rollback failing import transaction: %v", err)
		}
		requireCount(t, harness, `SELECT count(*) FROM records WHERE incident_id = $1`, incident.ID, beforeRecords)
		requireCount(t, harness, `SELECT count(*) FROM artifacts WHERE incident_id = $1`, incident.ID, beforeArtifacts)
		requireCount(t, harness, `SELECT count(*) FROM change_sets WHERE change_set_id = $1`, changeSetID, 0)
	})

	target := artifactTargets[artifacts.NotesViewSchemaID]
	facade, err := artifacts.NewImportContribution(
		artifacts.NotesViewSchemaID,
		*target.FacadeID,
		artifactImportDependencies(harness.DB, appender),
	)
	if err != nil {
		t.Fatalf("construct rollback import facade: %v", err)
	}
	tx, err := harness.DB.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		t.Fatalf("begin rollback import transaction: %v", err)
	}
	changeSetID, err := appender.AppendChangeSetTx(ctx, tx, revisions.AppendChangeSetParams{
		IncidentID: incident.ID, ActorUserID: actor.ID, Source: "imports.unit.apply", CreatedAt: now.Add(time.Hour),
	})
	if err != nil {
		_ = tx.Rollback(ctx)
		t.Fatalf("append rollback import change set: %v", err)
	}
	rolledBack, err := facade.CreateImportRowTx(ctx, tx, ownerfacade.ImportOwnerCreateCommand{
		Request: ownerfacade.ImportOwnerCreateRequest{
			IncidentID: incident.ID, ActorUserID: actor.ID,
			TargetViewSchemaID: artifacts.NotesViewSchemaID,
			ClientTxnID:        "txn-artifacts-import-rollback",
			FieldValues:        importTextFields("note.title", "Rolled back note"),
		},
		ChangeSetID: changeSetID, SequenceNo: 1, Now: now.Add(time.Hour),
	})
	if err != nil {
		_ = tx.Rollback(ctx)
		t.Fatalf("create rollback import row: %v", err)
	}
	if err := tx.Rollback(ctx); err != nil {
		t.Fatalf("rollback caller-owned import transaction: %v", err)
	}
	requireCount(t, harness, `SELECT count(*) FROM records WHERE record_id = $1`, rolledBack.RecordID, 0)
	requireCount(t, harness, `SELECT count(*) FROM change_sets WHERE change_set_id = $1`, changeSetID, 0)
	requireCount(t, harness, `SELECT count(*) FROM artifact_grid_projection WHERE record_id = $1`, rolledBack.RecordID, 0)
}
