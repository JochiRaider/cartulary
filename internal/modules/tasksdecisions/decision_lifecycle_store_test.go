package tasksdecisions_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/conflicttest"
)

func TestDecisionTerminalTransitionMatrix_Unit(t *testing.T) {
	harness := appsupport.StartStore(t, "workbook_interaction-task-decision-decision-terminal-matrix")
	codec := conflicttest.NewCodec("workbook")
	owner := appsupport.NewTaskDecisionOwner(harness.DB, codec)
	actor := authstoretest.SeedLocalUserRecord(t, harness.DB, "task-decision-decision-terminal@example.test", "TaskDecision Decision Terminal", "TaskDecisionDecisionTerminal1!", false, false, true)
	incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-workbook_interaction-task-decision-decision-terminal-incident", "IR-TASK-DECISION-DECISION-TERMINAL", "Workbook inspector task and decision workflow decision terminal matrix")

	for _, from := range []string{"rejected", "executed", "superseded"} {
		for _, to := range []string{"proposed", "approved", "rejected", "executed", "superseded"} {
			name := from + "-to-" + to
			decisionID := mustCreateDecisionInTerminalState(t, owner, actor, incident.ID, "txn-workbook_interaction-task-decision-decision-terminal-base-"+name, from)
			before := decisionSnapshot(t, harness.DB, decisionID)
			changes := []tasksdecisions.PatchChange{
				valueChange("decision.status", tasksdecisions.FieldValue{Text: stringPtr(to)}),
			}
			if from == to {
				changes = append(changes, valueChange("decision.rationale", tasksdecisions.FieldValue{Text: stringPtr("Idempotent in-state terminal write remains ordinary scalar work.")}))
			}
			_, err := patchRecord(owner, actor, decisionID, tasksdecisions.DecisionsViewSchemaID, before.RowVersion, "txn-workbook_interaction-task-decision-decision-terminal-"+name, changes...)
			if from == to && from != "superseded" {
				if err != nil {
					t.Fatalf("%s should allow in-state terminal write, got %v", name, err)
				}
				after := decisionSnapshot(t, harness.DB, decisionID)
				if after.Status != from || after.Rationale != "Idempotent in-state terminal write remains ordinary scalar work." || after.RowVersion <= before.RowVersion {
					t.Fatalf("%s unexpected in-state result: before=%#v after=%#v", name, before, after)
				}
				continue
			}
			requireLifecycle(t, err)
			requireDecisionSnapshot(t, decisionSnapshot(t, harness.DB, decisionID), before, name)
		}
	}
}

func mustCreateDecisionInTerminalState(t testing.TB, owner *tasksdecisions.MutationFacade, actor authn.UserRecord, incidentID uuid.UUID, clientTxnID string, status string) uuid.UUID {
	t.Helper()
	if status != "superseded" {
		return mustCreateDecision(t, owner, actor, incidentID, clientTxnID, status, "Terminal "+status)
	}
	target := mustCreateDecision(t, owner, actor, incidentID, clientTxnID+"-target", "proposed", "Superseded target")
	source := mustCreateDecision(t, owner, actor, incidentID, clientTxnID+"-source", "approved", "Superseding source")
	request := tasksdecisions.SupersedeRequest{
		BaseRowVersion:      1,
		ClientTxnID:         clientTxnID + "-supersede",
		Reason:              "Create explicit superseded terminal state.",
		ReplacementRecordID: &source,
	}
	if _, err := supersedeDecision(context.Background(), owner, actor, target, request, "req-"+clientTxnID+"-supersede", testTime(time.Hour)); err != nil {
		t.Fatalf("create superseded decision %s: %v", clientTxnID, err)
	}
	return target
}
