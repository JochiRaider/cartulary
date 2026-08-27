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
	request := evidence.CreateRequest{
		ViewSchemaID: evidence.ViewSchemaID, ClientTxnID: clientTxnID,
		Values: map[string]evidence.FieldValue{
			"evidence.title": {Text: &title},
		},
	}
	result, err := appsupport.NewEvidenceMutationOwner(pool, workbookTestConflictTokens()).Create(
		context.Background(),
		evidence.CreateCommand{
			Actor: actor, IncidentID: incidentID, Request: request,
			RequestHash: evidence.CreateRequestHash(request), RequestID: "req-" + clientTxnID,
			RouteKey: "workbook.rows.create", Now: time.Date(2026, 5, 18, 12, 1, 0, 0, time.UTC),
		},
	)
	if err != nil {
		t.Fatalf("create evidence %s: %v", clientTxnID, err)
	}
	return result.RecordID
}

func mustCreateTaskFor(t testing.TB, store *workbook.WorkbookContributionCatalog, actor authn.UserRecord, incidentID uuid.UUID, clientTxnID string, title string) uuid.UUID {
	t.Helper()
	result, err := workbookroutetest.CreateWorkbookRow(store, context.Background(), actor, incidentID, workbookroutetest.CreateRequest{
		ViewSchemaID: tasksdecisions.TaskRequestsViewSchemaID,
		ClientTxnID:  clientTxnID,
		Values: map[string]workbookroutetest.ValueChange{
			"task.title":     {Kind: "text", Text: &title},
			"task.task_kind": {Kind: "text", Text: stringPtr("request")},
		},
	}, "req-"+clientTxnID, time.Date(2026, 5, 18, 12, 2, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("create task %s: %v", clientTxnID, err)
	}
	return result.RecordID
}

func mustCreateDecision(t testing.TB, store *workbook.WorkbookContributionCatalog, actor authn.UserRecord, incidentID uuid.UUID, clientTxnID string, status string, summary string) uuid.UUID {
	t.Helper()
	values := map[string]workbookroutetest.ValueChange{
		"decision.summary":       {Kind: "text", Text: &summary},
		"decision.decision_type": {Kind: "text", Text: stringPtr("containment")},
		"decision.rationale":     {Kind: "text", Text: stringPtr("The decision is needed for coordinated response.")},
	}
	if status != "" {
		values["decision.status"] = workbookroutetest.ValueChange{Kind: "text", Text: &status}
	}
	result, err := workbookroutetest.CreateWorkbookRow(store, context.Background(), actor, incidentID, workbookroutetest.CreateRequest{
		ViewSchemaID: tasksdecisions.DecisionsViewSchemaID,
		ClientTxnID:  clientTxnID,
		Values:       values,
	}, "req-"+clientTxnID, Time(0))
	if err != nil {
		t.Fatalf("create decision %s: %v", clientTxnID, err)
	}
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

func stringPtr(value string) *string {
	return &value
}
