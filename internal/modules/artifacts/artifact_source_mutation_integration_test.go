package artifacts_test

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"slices"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/conflicttest"
)

func TestArtifactWorkbookMutationContractMatrix(t *testing.T) {
	ctx := context.Background()
	harness := appsupport.StartStore(t, "artifacts-workbook-mutation-contract")
	actor := authstoretest.SeedLocalUserRecord(
		t,
		harness.DB,
		"artifact-contract-owner@example.test",
		"Artifact Contract Owner",
		"ArtifactContractOwner1!",
		false,
		false,
		true,
	)
	incident := appsupport.CreateIncidentInStore(
		t,
		harness.DB,
		actor,
		"txn-artifacts-contract-incident",
		"IR-ARTIFACTS-CONTRACT",
		"Artifact mutation contract",
	)
	facade := mustArtifactMutationFacade(
		t,
		harness.DB,
		conflicttest.NewCodec("artifacts-contract"),
	)
	now := time.Date(2026, 7, 30, 14, 0, 0, 0, time.UTC)

	cases := []struct {
		name         string
		viewSchemaID string
		artifactType string
		values       func() map[string]any
		patchField   string
		patchValue   string
	}{
		{"note", artifacts.NotesViewSchemaID, "note", func() map[string]any {
			return artifactTextValues("note.title", "Contract note")
		}, "note.body", "Patched note body"},
		{"comm_log", artifacts.CommLogViewSchemaID, "comm_log", func() map[string]any {
			return artifactTextValues(
				"comm_log.comm_type", "briefing",
				"comm_log.audience", "responders",
				"comm_log.channel_or_meeting", "bridge",
				"comm_log.summary", "Initial briefing",
			)
		}, "comm_log.summary", "Updated briefing"},
		{"handoff", artifacts.HandoffViewSchemaID, "handoff", func() map[string]any {
			return map[string]any{
				"handoff.incoming_owner_user_id": actor.ID.String(),
				"handoff.current_state_summary":  "Containment is stable",
			}
		}, "handoff.next_checks", "Review telemetry"},
		{"status_review", artifacts.StatusReviewViewSchemaID, "status_review", func() map[string]any {
			return artifactTextValues("status_review.current_state_summary", "Recovery in progress")
		}, "status_review.active_risks_summary", "One residual risk"},
		{"lesson", artifacts.LessonViewSchemaID, "lesson", func() map[string]any {
			return artifactTextValues("lesson.summary", "Capture volatile data early")
		}, "lesson.summary", "Capture volatile data immediately"},
		{"finding", artifacts.FindingsViewSchemaID, "finding", func() map[string]any {
			return artifactTextValues("finding.statement", "Credential misuse observed")
		}, "finding.statement", "Credential misuse confirmed"},
		{"investigative_query", artifacts.InvestigativeQueriesViewSchemaID, "investigative_query", func() map[string]any {
			return artifactTextValues(
				"investigative_query.platform", "edr",
				"investigative_query.purpose", "scope",
				"investigative_query.query_text", "synthetic query",
			)
		}, "investigative_query.purpose", "containment"},
		{"forensic_keyword", artifacts.ForensicKeywordsViewSchemaID, "forensic_keyword", func() map[string]any {
			return artifactTextValues(
				"forensic_keyword.pattern", "synthetic-pattern",
				"forensic_keyword.reason", "Synthetic fixture",
			)
		}, "forensic_keyword.reason", "Updated synthetic fixture"},
	}

	for index, tc := range cases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			clientTxnID := fmt.Sprintf("txn-artifacts-contract-create-%02d", index)
			values := tc.values()
			command := artifacts.CreateCommand{
				ActorUserID: actor.ID,
				IncidentID:  incident.ID,
				Admission:   mustArtifactCreateAdmission(t, tc.viewSchemaID, clientTxnID, values, nil),
				RequestID:   "req-" + clientTxnID,
				Now:         now.Add(time.Duration(index) * time.Minute),
			}
			created, err := facade.Create(ctx, command)
			if err != nil {
				t.Fatalf("create %s: %v", tc.viewSchemaID, err)
			}
			if created.RecordID == uuid.Nil || created.ChangeSetID == nil || created.RowVersion != 1 {
				t.Fatalf("create result is incomplete: %#v", created)
			}
			requireArtifactRecordChangedIntent(t, harness, created, actor.ID)
			requireCount(t, harness, `
SELECT count(*)
  FROM records r
  JOIN artifacts a USING (incident_id, record_id)
 WHERE r.record_id = $1
   AND r.record_type = 'artifact'
   AND r.row_version = 1
   AND a.artifact_type = $2
   AND a.created_by_user_id = $3
`, created.RecordID, tc.artifactType, actor.ID, 1)
			requireCount(t, harness, `SELECT count(*) FROM artifact_grid_projection WHERE record_id = $1 AND artifact_type = $2`, created.RecordID, tc.artifactType, 1)
			requireCount(t, harness, `SELECT count(*) FROM change_sets WHERE change_set_id = $1 AND source = $2`, created.ChangeSetID, string(artifacts.OperationCreate), 1)
			requireCount(t, harness, `SELECT count(*) FROM change_set_mutations WHERE change_set_id = $1`, created.ChangeSetID, 1)
			requireCount(t, harness, `SELECT count(*) FROM record_revisions WHERE record_id = $1 AND row_version = 1`, created.RecordID, 1)

			replayed, err := facade.Create(ctx, command)
			if err != nil {
				t.Fatalf("replay %s create: %v", tc.viewSchemaID, err)
			}
			if replayed.Outcome != artifacts.MutationOutcomeReplayed || replayed.RecordID != created.RecordID {
				t.Fatalf("create replay = %#v, want original record %s", replayed, created.RecordID)
			}
			requireCount(t, harness, `SELECT count(*) FROM records WHERE record_id = $1`, created.RecordID, 1)
			requireCount(t, harness, `SELECT count(*) FROM record_revisions WHERE record_id = $1`, created.RecordID, 1)
			collaborationsupport.RequireIntentCount(t, harness.DB, collaborationsupport.IntentSelector{
				EventFamily: "record_changed", SourceChangeSetID: created.ChangeSetID.String(), SourceRecordID: created.RecordID.String(),
			}, 1)

			conflicting := command
			divergentValues := maps.Clone(values)
			divergentValues[tc.patchField] = tc.patchValue
			conflicting.Admission = mustArtifactCreateAdmission(t, tc.viewSchemaID, clientTxnID, divergentValues, nil)
			if _, err := facade.Create(ctx, conflicting); !errors.Is(err, artifacts.ErrClientTxnConflict) {
				t.Fatalf("changed create replay error = %v, want client transaction conflict", err)
			}

			patchTxnID := fmt.Sprintf("txn-artifacts-contract-patch-%02d", index)
			patchCommand := artifacts.PatchCommand{
				ActorUserID: actor.ID,
				RecordID:    created.RecordID,
				Admission:   mustArtifactPatchAdmission(t, tc.viewSchemaID, 1, patchTxnID, []map[string]any{artifactValueChange(tc.patchField, tc.patchValue)}),
				RequestID:   "req-" + patchTxnID,
				Now:         now.Add(2*time.Hour + time.Duration(index)*time.Minute),
			}
			patched, err := facade.Patch(ctx, patchCommand)
			if err != nil {
				t.Fatalf("patch %s: %v", tc.viewSchemaID, err)
			}
			if patched.RowVersion != 2 || !slices.Contains(patched.ChangedFieldKeys, tc.patchField) {
				t.Fatalf("patch result = %#v, want row version 2 and changed field %s", patched, tc.patchField)
			}
			requireArtifactRecordChangedIntent(t, harness, patched, actor.ID)
			requireCount(t, harness, `SELECT count(*) FROM record_revisions WHERE record_id = $1`, created.RecordID, 2)
			requireCount(t, harness, `SELECT count(*) FROM artifact_grid_projection WHERE record_id = $1 AND row_version = 2`, created.RecordID, 1)

			patchReplay, err := facade.Patch(ctx, patchCommand)
			if err != nil || patchReplay.Outcome != artifacts.MutationOutcomeReplayed || patchReplay.RecordID != created.RecordID {
				t.Fatalf("patch replay = %#v, %v; want original record", patchReplay, err)
			}
			requireCount(t, harness, `SELECT count(*) FROM record_revisions WHERE record_id = $1`, created.RecordID, 2)
			collaborationsupport.RequireIntentCount(t, harness.DB, collaborationsupport.IntentSelector{
				EventFamily: "record_changed", SourceChangeSetID: patched.ChangeSetID.String(), SourceRecordID: created.RecordID.String(),
			}, 1)
		})
	}
}
