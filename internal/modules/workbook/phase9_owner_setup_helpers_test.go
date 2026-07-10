package workbook_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/workbook"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

func mustCreatePartyForU911(t testing.TB, store *workbook.Store, actor authn.UserRecord, incidentID uuid.UUID, clientTxnID string, displayName string) uuid.UUID {
	t.Helper()
	result, err := store.CreateWorkbookRow(context.Background(), actor, incidentID, workbook.CreateRequest{
		ViewSchemaID: workbook.PartiesViewSchemaID,
		ClientTxnID:  clientTxnID,
		Values: map[string]workbook.ValueChange{
			"party.display_name": {Kind: "text", Text: &displayName},
			"party.party_kind":   {Kind: "text", Text: stringPtrU911("organization")},
		},
	}, []byte(clientTxnID), "req-"+clientTxnID, time.Date(2026, 5, 18, 12, 0, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("create party %s: %v", clientTxnID, err)
	}
	return result.RecordID
}

func mustCreateEvidenceForU911(t testing.TB, store *workbook.Store, actor authn.UserRecord, incidentID uuid.UUID, clientTxnID string, title string) uuid.UUID {
	t.Helper()
	result, err := store.CreateWorkbookRow(context.Background(), actor, incidentID, workbook.CreateRequest{
		ViewSchemaID: workbook.EvidenceViewSchemaID,
		ClientTxnID:  clientTxnID,
		Values: map[string]workbook.ValueChange{
			"evidence.title": {Kind: "text", Text: &title},
		},
	}, []byte(clientTxnID), "req-"+clientTxnID, time.Date(2026, 5, 18, 12, 1, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("create evidence %s: %v", clientTxnID, err)
	}
	return result.RecordID
}

func mustCreateTaskForU911(t testing.TB, store *workbook.Store, actor authn.UserRecord, incidentID uuid.UUID, clientTxnID string, title string) uuid.UUID {
	t.Helper()
	result, err := store.CreateWorkbookRow(context.Background(), actor, incidentID, workbook.CreateRequest{
		ViewSchemaID: workbook.TaskRequestsViewSchemaID,
		ClientTxnID:  clientTxnID,
		Values: map[string]workbook.ValueChange{
			"task.title":     {Kind: "text", Text: &title},
			"task.task_kind": {Kind: "text", Text: stringPtrU911("request")},
		},
	}, []byte(clientTxnID), "req-"+clientTxnID, time.Date(2026, 5, 18, 12, 2, 0, 0, time.UTC))
	if err != nil {
		t.Fatalf("create task %s: %v", clientTxnID, err)
	}
	return result.RecordID
}

func mustCreateSprint6Decision(t testing.TB, store *workbook.Store, actor authn.UserRecord, incidentID uuid.UUID, clientTxnID string, status string, summary string) uuid.UUID {
	t.Helper()
	values := map[string]workbook.ValueChange{
		"decision.summary":       {Kind: "text", Text: &summary},
		"decision.decision_type": {Kind: "text", Text: stringPtrU911("containment")},
		"decision.rationale":     {Kind: "text", Text: stringPtrU911("The decision is needed for coordinated response.")},
	}
	if status != "" {
		values["decision.status"] = workbook.ValueChange{Kind: "text", Text: &status}
	}
	result, err := store.CreateWorkbookRow(context.Background(), actor, incidentID, workbook.CreateRequest{
		ViewSchemaID: workbook.DecisionsViewSchemaID,
		ClientTxnID:  clientTxnID,
		Values:       values,
	}, []byte(clientTxnID), "req-"+clientTxnID, sprint7Time(0))
	if err != nil {
		t.Fatalf("create decision %s: %v", clientTxnID, err)
	}
	return result.RecordID
}

func sprint6RecordVersion(t testing.TB, db interface {
	QueryRow(context.Context, string, ...any) pgx.Row
}, recordID uuid.UUID) int64 {
	t.Helper()
	var version int64
	if err := db.QueryRow(context.Background(), `SELECT row_version FROM records WHERE record_id = $1`, recordID).Scan(&version); err != nil {
		t.Fatalf("query record version: %v", err)
	}
	return version
}

func stringPtrU911(value string) *string {
	return &value
}
