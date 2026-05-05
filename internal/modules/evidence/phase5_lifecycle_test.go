package evidence_test

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/workbook"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/phase4storetest"
)

func TestPhase5_EvidenceLifecycleSeparateFromBlob_U_5_04(t *testing.T) {
	harness := phase4storetest.StartStore(t, "phase5-evidence-lifecycle")
	workbookStore := workbook.NewStore(harness.DB)
	evidenceStore := evidence.NewStore(harness.DB)
	actor := phase4storetest.SeedLocalUserFlags(t, harness.DB, "phase5-lifecycle@example.test", "Phase5 Lifecycle", "Phase5Lifecycle1!", false, false, true)
	incident := phase4storetest.CreateIncidentInStore(t, harness.DB, actor, "txn-phase5-lifecycle-incident", "IR-P5-LIFECYCLE", "Phase 5 lifecycle")

	requested := createPhase5EvidenceViaWorkbook(t, workbookStore, actor, incident.ID, `{
		"client_txn_id":"txn-phase5-lifecycle-requested",
		"evidence.title":" Requested endpoint package ",
		"evidence.storage_ref":" ticket://collect-1 ",
		"evidence.collector_party_text":" IR collector ",
		"evidence.source_party_text":" Endpoint owner "
	}`)
	requestedRow := requested.Payload["row"].(map[string]any)
	recordID := mustPhase5RowID(t, requestedRow)
	requirePhase5RowCellValue(t, requestedRow, "evidence.lifecycle_state", "requested")
	requirePhase5RowCellValue(t, requestedRow, "evidence.upload_state", "pending")
	requirePhase5RowCellValue(t, requestedRow, "evidence.storage_ref", "ticket://collect-1")
	requirePhase5RowCellValue(t, requestedRow, "evidence.collector_party_text", "IR collector")
	requirePhase5RowCellValue(t, requestedRow, "evidence.source_party_text", "Endpoint owner")
	requirePhase5RowCellNonEmpty(t, requestedRow, "evidence.requested_at")

	pendingReceipt := patchPhase5EvidenceViaWorkbook(t, workbookStore, actor, recordID, `{
		"view_schema_id":"cartulary.view.evidence.v1",
		"base_row_version":1,
		"client_txn_id":"txn-phase5-lifecycle-pending",
		"changes":[{"field_key":"evidence.lifecycle_state","value":"pending_receipt"}]
	}`)
	requirePhase5RowCellValue(t, pendingReceipt.Payload["row"].(map[string]any), "evidence.lifecycle_state", "pending_receipt")

	received := patchPhase5EvidenceViaWorkbook(t, workbookStore, actor, recordID, `{
		"view_schema_id":"cartulary.view.evidence.v1",
		"base_row_version":2,
		"client_txn_id":"txn-phase5-lifecycle-received",
		"changes":[{"field_key":"evidence.lifecycle_state","value":"received"}]
	}`)
	requirePhase5RowCellValue(t, received.Payload["row"].(map[string]any), "evidence.lifecycle_state", "received")

	requirePhase5IllegalLifecyclePatch(t, workbookStore, actor, recordID, `{
		"view_schema_id":"cartulary.view.evidence.v1",
		"base_row_version":3,
		"client_txn_id":"txn-phase5-lifecycle-available-without-blob",
		"changes":[{"field_key":"evidence.lifecycle_state","value":"available"}]
	}`)
	requirePhase5IllegalLifecyclePatch(t, workbookStore, actor, recordID, `{
		"view_schema_id":"cartulary.view.evidence.v1",
		"base_row_version":3,
		"client_txn_id":"txn-phase5-lifecycle-released-direct",
		"changes":[{"field_key":"evidence.lifecycle_state","value":"released"}]
	}`)

	attachTarget := createPhase5EvidenceViaWorkbook(t, workbookStore, actor, incident.ID, `{
		"client_txn_id":"txn-phase5-lifecycle-attach-target",
		"evidence.title":"Blob arrives later",
		"evidence.storage_ref":"ticket://collect-2",
		"evidence.collector_party_text":"IR collector",
		"evidence.source_party_text":"Endpoint owner"
	}`)
	attachRecordID := mustPhase5RowID(t, attachTarget.Payload["row"].(map[string]any))
	blobID := seedPhase5Blob(t, harness.DB, incident.ID, actor.ID, "available", phase5BlobOptions{
		ByteSize: 4, ObservedSize: ptrInt64(4), ObservedSHA: ptrString(strings.Repeat("b", 64)), ObservedContentType: ptrString("text/plain"),
	})
	attachRequest := evidence.AttachBlobRequest{ObjectBlobID: blobID, BaseRowVersion: 1, ClientTxnID: "txn-phase5-lifecycle-attach"}
	attached, err := evidenceStore.AttachBlob(context.Background(), actor, attachRecordID, attachRequest, evidence.AttachBlobRequestHash(attachRequest), nil, "req-lifecycle-attach", time.Now().UTC())
	if err != nil {
		t.Fatalf("attach available blob to requested evidence: %v", err)
	}
	attachedRow := attached.Payload["row"].(map[string]any)
	requirePhase5RowCellValue(t, attachedRow, "evidence.lifecycle_state", "available")
	requirePhase5RowCellValue(t, attachedRow, "evidence.upload_state", "available")
	requirePhase5RowCellValue(t, attachedRow, "evidence.collector_party_text", "IR collector")
	requirePhase5RowCellValue(t, attachedRow, "evidence.source_party_text", "Endpoint owner")
	requirePhase5RowCellNonEmpty(t, attachedRow, "evidence.received_at")

	quarantined := createPhase5EvidenceViaWorkbook(t, workbookStore, actor, incident.ID, `{
		"client_txn_id":"txn-phase5-lifecycle-quarantined",
		"evidence.title":"Quarantined evidence"
	}`)
	quarantinedRecordID := mustPhase5RowID(t, quarantined.Payload["row"].(map[string]any))
	if _, err := harness.DB.Exec(context.Background(), `UPDATE evidence SET lifecycle_state = 'quarantined' WHERE record_id = $1`, quarantinedRecordID); err != nil {
		t.Fatalf("seed quarantined lifecycle: %v", err)
	}
	quarantinedBlobID := seedPhase5Blob(t, harness.DB, incident.ID, actor.ID, "available", phase5BlobOptions{ByteSize: 1, ObservedSize: ptrInt64(1)})
	quarantinedAttach := evidence.AttachBlobRequest{ObjectBlobID: quarantinedBlobID, BaseRowVersion: 1, ClientTxnID: "txn-phase5-quarantined-attach"}
	if _, err := evidenceStore.AttachBlob(context.Background(), actor, quarantinedRecordID, quarantinedAttach, evidence.AttachBlobRequestHash(quarantinedAttach), nil, "req-quarantined-attach", time.Now().UTC()); !errors.Is(err, evidence.ErrBlobNotAttachable) {
		t.Fatalf("quarantined evidence attach got %v want ErrBlobNotAttachable", err)
	}

	requirePhase5EvidenceState(t, harness.DB, attachRecordID, "available", "available", blobID)
}

func createPhase5EvidenceViaWorkbook(t testing.TB, store *workbook.Store, actor authn.UserRecord, incidentID uuid.UUID, body string) workbook.MutationResult {
	t.Helper()
	request, apiErr := workbook.DecodeCreateRequest(workbook.EvidenceViewSchemaID, strings.NewReader(body))
	if apiErr != nil {
		t.Fatalf("decode evidence create: %#v", apiErr)
	}
	result, err := store.CreateWorkbookRow(context.Background(), actor, incidentID, request, workbook.CreateRequestHash(request), "req-"+request.ClientTxnID, time.Now().UTC())
	if err != nil {
		t.Fatalf("create evidence via workbook store: %v", err)
	}
	return result
}

func patchPhase5EvidenceViaWorkbook(t testing.TB, store *workbook.Store, actor authn.UserRecord, recordID uuid.UUID, body string) workbook.MutationResult {
	t.Helper()
	request, apiErr := workbook.DecodePatchRequest(strings.NewReader(body))
	if apiErr != nil {
		t.Fatalf("decode evidence patch: %#v", apiErr)
	}
	result, err := store.PatchWorkbookRow(context.Background(), actor, recordID, request, workbook.PatchRequestHash(request), "req-"+request.ClientTxnID, time.Now().UTC())
	if err != nil {
		t.Fatalf("patch evidence via workbook store: %v", err)
	}
	return result
}

func requirePhase5IllegalLifecyclePatch(t testing.TB, store *workbook.Store, actor authn.UserRecord, recordID uuid.UUID, body string) {
	t.Helper()
	request, apiErr := workbook.DecodePatchRequest(strings.NewReader(body))
	if apiErr != nil {
		t.Fatalf("decode illegal evidence patch: %#v", apiErr)
	}
	_, err := store.PatchWorkbookRow(context.Background(), actor, recordID, request, workbook.PatchRequestHash(request), "req-"+request.ClientTxnID, time.Now().UTC())
	var lifecycleErr *workbook.LifecycleValidationError
	if !errors.As(err, &lifecycleErr) {
		t.Fatalf("illegal evidence lifecycle patch got %v want LifecycleValidationError", err)
	}
}

func mustPhase5RowID(t testing.TB, row map[string]any) uuid.UUID {
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

func requirePhase5RowCellValue(t testing.TB, row map[string]any, fieldKey string, want any) {
	t.Helper()
	cells := row["cells"].(map[string]any)
	got := cells[fieldKey].(map[string]any)["value"]
	if got != want {
		t.Fatalf("%s got %#v want %#v", fieldKey, got, want)
	}
}

func requirePhase5RowCellNonEmpty(t testing.TB, row map[string]any, fieldKey string) {
	t.Helper()
	cells := row["cells"].(map[string]any)
	if got := cells[fieldKey].(map[string]any)["value"]; got == nil || got == "" {
		t.Fatalf("%s got empty value %#v", fieldKey, got)
	}
}
