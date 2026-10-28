package workbook_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/parties"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions"
	"github.com/JochiRaider/cartulary/internal/modules/workbook"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/workbookroutetest"
)

func mustCreatePartyFor(t testing.TB, pool postgres.DB, actor authn.UserRecord, incidentID uuid.UUID, clientTxnID string, displayName string) uuid.UUID {
	t.Helper()
	admission, apiErr := parties.AdmitCreateJSON(strings.NewReader(fmt.Sprintf(
		`{"client_txn_id":%q,"party.display_name":%q,"party.party_kind":"organization"}`,
		clientTxnID,
		displayName,
	)))
	if apiErr != nil {
		t.Fatalf("admit party %s: %#v", clientTxnID, apiErr)
	}
	result, err := appsupport.NewPartyOwner(pool, workbookTestConflictTokens()).Create(
		context.Background(),
		parties.CreateCommand{
			ActorUserID: actor.ID, IncidentID: incidentID, Admission: admission, RequestID: "req-" + clientTxnID,
			Now: time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC),
		},
	)
	if err != nil {
		t.Fatalf("create party %s: %v", clientTxnID, err)
	}
	return result.RecordID
}

func mustCreateEvidenceFor(t testing.TB, pool postgres.DB, actor authn.UserRecord, incidentID uuid.UUID, clientTxnID string, title string) uuid.UUID {
	t.Helper()
	request, admissionFailure := evidence.AdmitCreateJSON(strings.NewReader(fmt.Sprintf(
		`{"client_txn_id":%q,"evidence.title":%q}`,
		clientTxnID,
		title,
	)))
	if admissionFailure != nil {
		t.Fatalf("admit evidence %s: %#v", clientTxnID, admissionFailure)
	}
	result, err := appsupport.NewEvidenceMutationOwner(pool, workbookTestConflictTokens()).Create(
		context.Background(),
		evidence.CreateCommand{
			ActorUserID: actor.ID, IncidentID: incidentID, Admission: request,
			RequestID: "req-" + clientTxnID, Now: time.Date(2026, 5, 18, 12, 1, 0, 0, time.UTC),
		},
	)
	if err != nil {
		t.Fatalf("create evidence %s: %v", clientTxnID, err)
	}
	return result.RecordID
}

func mustCreateTaskFor(t testing.TB, store *workbook.WorkbookContributionCatalog, actor authn.UserRecord, incidentID uuid.UUID, clientTxnID string, title string) uuid.UUID {
	t.Helper()
	createOutcome1, err := workbookroutetest.CreateWorkbookRow(store, context.Background(), workbook.CreateCommand{Actor: actor, IncidentID: incidentID, ViewSchemaID: tasksdecisions.TaskRequestsViewSchemaID, RequestID: "req-" + clientTxnID, Now: time.Date(2026, 5, 18, 12, 2, 0, 0, time.UTC)}, httptestx.JSONReader(t, map[string]any{"client_txn_id": clientTxnID,
		"task.title":     title,
		"task.task_kind": "request"}))
	result := workbookroutetest.RequireMutationResult(t, createOutcome1, err)
	return result.RecordID
}

func mustCreateDecision(t testing.TB, store *workbook.WorkbookContributionCatalog, actor authn.UserRecord, incidentID uuid.UUID, clientTxnID string, status string, summary string) uuid.UUID {
	t.Helper()
	values := map[string]any{"decision.summary": summary,
		"decision.decision_type": "containment",
		"decision.rationale":     "The decision is needed for coordinated response."}
	if status != "" {
		values["decision.status"] = status
	}
	createBody2 := map[string]any{"client_txn_id": clientTxnID}
	for key, value := range values {
		createBody2[key] = value
	}
	createOutcome2, err := workbookroutetest.CreateWorkbookRow(store, context.Background(), workbook.CreateCommand{Actor: actor, IncidentID: incidentID, ViewSchemaID: tasksdecisions.DecisionsViewSchemaID, RequestID: "req-" + clientTxnID, Now: Time(0)}, httptestx.JSONReader(t, createBody2))
	result := workbookroutetest.RequireMutationResult(t, createOutcome2, err)
	return result.RecordID
}

func RecordVersion(t testing.TB, db interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}, recordID uuid.UUID) int64 {
	t.Helper()
	var version int64
	if err := db.QueryRow(context.Background(), `SELECT row_version FROM records WHERE record_id = $1`, recordID).Scan(&version); err != nil {
		t.Fatalf("query record version: %v", err)
	}
	return version
}
