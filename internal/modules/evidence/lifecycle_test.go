package evidence_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	authstoretest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/storetest"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/conflicttest"
)

func TestEvidenceLifecycleSeparateFromBlob_Unit(t *testing.T) {
	harness := appsupport.StartStore(t, "evidence_lifecycle-evidence-lifecycle")
	evidenceOwner := appsupport.NewEvidenceMutationOwner(harness.DB, conflicttest.NewCodec("workbook"))
	actor := authstoretest.SeedLocalUserRecord(t, harness.DB, "evidence_lifecycle-lifecycle@example.test", "EvidenceLifecycle Lifecycle", "EvidenceLifecycleLifecycle1!", false, false, true)
	incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-evidence_lifecycle-lifecycle-incident", "IR-P5-LIFECYCLE", "Evidence lifecycle")

	requested := createEvidenceViaOwner(t, evidenceOwner, actor, incident.ID, `{
		"client_txn_id":"txn-evidence_lifecycle-lifecycle-requested",
		"evidence.title":" Requested endpoint package ",
		"evidence.storage_ref":" ticket://collect-1 ",
		"evidence.collector_party_text":" IR collector ",
		"evidence.source_party_text":" Endpoint owner "
	}`)
	requestedRow := requested.Row
	recordID := mustRowID(t, requestedRow)
	requireRowCellValue(t, requestedRow, "evidence.lifecycle_state", "requested")
	requireRowCellValue(t, requestedRow, "evidence.upload_state", "pending")
	requireRowCellValue(t, requestedRow, "evidence.storage_ref", "ticket://collect-1")
	requireRowCellValue(t, requestedRow, "evidence.collector_party_text", "IR collector")
	requireRowCellValue(t, requestedRow, "evidence.source_party_text", "Endpoint owner")
	requireRowCellNonEmpty(t, requestedRow, "evidence.requested_at")

	pendingReceipt := patchEvidenceViaOwner(t, evidenceOwner, actor, recordID, `{
		"view_schema_id":"cartulary.view.evidence.v1",
		"base_row_version":1,
		"client_txn_id":"txn-evidence_lifecycle-lifecycle-pending",
		"changes":[{"field_key":"evidence.lifecycle_state","value":"pending_receipt"}]
	}`)
	requireRowCellValue(t, pendingReceipt.Row, "evidence.lifecycle_state", "pending_receipt")

	received := patchEvidenceViaOwner(t, evidenceOwner, actor, recordID, `{
		"view_schema_id":"cartulary.view.evidence.v1",
		"base_row_version":2,
		"client_txn_id":"txn-evidence_lifecycle-lifecycle-received",
		"changes":[{"field_key":"evidence.lifecycle_state","value":"received"}]
	}`)
	requireRowCellValue(t, received.Row, "evidence.lifecycle_state", "received")

	requireIllegalLifecyclePatch(t, evidenceOwner, actor, recordID, `{
		"view_schema_id":"cartulary.view.evidence.v1",
		"base_row_version":3,
		"client_txn_id":"txn-evidence_lifecycle-lifecycle-available-without-blob",
		"changes":[{"field_key":"evidence.lifecycle_state","value":"available"}]
	}`)
	requireIllegalLifecyclePatch(t, evidenceOwner, actor, recordID, `{
		"view_schema_id":"cartulary.view.evidence.v1",
		"base_row_version":3,
		"client_txn_id":"txn-evidence_lifecycle-lifecycle-released-direct",
		"changes":[{"field_key":"evidence.lifecycle_state","value":"released"}]
	}`)
}

func createEvidenceViaOwner(t testing.TB, owner evidence.MutationContribution, actor authn.UserRecord, incidentID uuid.UUID, body string) evidence.MutationResult {
	t.Helper()
	request, admissionFailure := evidence.AdmitCreateJSON(strings.NewReader(body))
	if admissionFailure != nil {
		t.Fatalf("admit evidence create: %#v", admissionFailure)
	}
	result, err := owner.Create(context.Background(), evidence.CreateCommand{
		ActorUserID: actor.ID, IncidentID: incidentID, Admission: request,
		RequestID: "req-" + request.ClientTxnID(), Now: time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("create evidence via owner: %v", err)
	}
	return result
}

func patchEvidenceViaOwner(t testing.TB, owner evidence.MutationContribution, actor authn.UserRecord, recordID uuid.UUID, body string) evidence.MutationResult {
	t.Helper()
	request, admissionFailure := evidence.AdmitPatchJSON(strings.NewReader(body))
	if admissionFailure != nil {
		t.Fatalf("admit evidence patch: %#v", admissionFailure)
	}
	result, err := owner.Patch(context.Background(), evidence.PatchCommand{
		ActorUserID: actor.ID, RecordID: recordID, Admission: request,
		RequestID: "req-" + request.ClientTxnID(), Now: time.Now().UTC(),
	})
	if err != nil {
		t.Fatalf("patch evidence via owner: %v", err)
	}
	return result
}

func requireIllegalLifecyclePatch(t testing.TB, owner evidence.MutationContribution, actor authn.UserRecord, recordID uuid.UUID, body string) {
	t.Helper()
	request, admissionFailure := evidence.AdmitPatchJSON(strings.NewReader(body))
	if admissionFailure != nil {
		t.Fatalf("admit illegal evidence patch: %#v", admissionFailure)
	}
	_, err := owner.Patch(context.Background(), evidence.PatchCommand{
		ActorUserID: actor.ID, RecordID: recordID, Admission: request,
		RequestID: "req-" + request.ClientTxnID(), Now: time.Now().UTC(),
	})
	var lifecycleErr *evidence.LifecycleValidationError
	if !errors.As(err, &lifecycleErr) {
		t.Fatalf("illegal evidence lifecycle patch got %v want LifecycleValidationError", err)
	}
}

func mustRowID(t testing.TB, row map[string]any) uuid.UUID {
	t.Helper()
	value, ok := row["record_id"].(string)
	if !ok {
		t.Fatalf("row record_id got %#v", row["record_id"])
	}
	parsed, err := uuid.Parse(value)
	if err != nil {
		t.Fatalf("parse row record_id: %v", err)
	}
	return parsed
}

func requireRowCellValue(t testing.TB, row map[string]any, fieldKey string, want any) {
	t.Helper()
	cells := row["cells"].(map[string]any)
	got := cells[fieldKey].(map[string]any)["value"]
	if got != want {
		t.Fatalf("%s got %#v want %#v", fieldKey, got, want)
	}
}

func requireRowCellNonEmpty(t testing.TB, row map[string]any, fieldKey string) {
	t.Helper()
	cells := row["cells"].(map[string]any)
	if got := cells[fieldKey].(map[string]any)["value"]; got == nil || got == "" {
		t.Fatalf("%s got empty value %#v", fieldKey, got)
	}
}
