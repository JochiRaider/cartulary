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
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/modules/workbook"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/conflicttest"
	"github.com/JochiRaider/cartulary/internal/testutil/revisionsupport"
)

func TestEvidenceLifecycleSeparateFromBlob_Unit(t *testing.T) {
	harness := appsupport.StartStore(t, "evidence_lifecycle-evidence-lifecycle")
	workbookStore := appsupport.NewWorkbookStore(harness.DB, conflicttest.NewCodec("workbook"))
	revisionComposition := revisionsupport.MustComposition(t)
	evidenceStore := newTestBlobLifecycleService(harness.DB, revisionComposition.Runtime.Appender(), revisionComposition.Intents)
	actor := authstoretest.SeedLocalUserRecord(t, harness.DB, "evidence_lifecycle-lifecycle@example.test", "EvidenceLifecycle Lifecycle", "EvidenceLifecycleLifecycle1!", false, false, true)
	incident := appsupport.CreateIncidentInStore(t, harness.DB, actor, "txn-evidence_lifecycle-lifecycle-incident", "IR-P5-LIFECYCLE", "Evidence lifecycle")

	requested := createEvidenceViaWorkbook(t, workbookStore, actor, incident.ID, `{
		"client_txn_id":"txn-evidence_lifecycle-lifecycle-requested",
		"evidence.title":" Requested endpoint package ",
		"evidence.storage_ref":" ticket://collect-1 ",
		"evidence.collector_party_text":" IR collector ",
		"evidence.source_party_text":" Endpoint owner "
	}`)
	requestedRow := requested.Payload["row"].(map[string]any)
	recordID := mustRowID(t, requestedRow)
	requireRowCellValue(t, requestedRow, "evidence.lifecycle_state", "requested")
	requireRowCellValue(t, requestedRow, "evidence.upload_state", "pending")
	requireRowCellValue(t, requestedRow, "evidence.storage_ref", "ticket://collect-1")
	requireRowCellValue(t, requestedRow, "evidence.collector_party_text", "IR collector")
	requireRowCellValue(t, requestedRow, "evidence.source_party_text", "Endpoint owner")
	requireRowCellNonEmpty(t, requestedRow, "evidence.requested_at")

	pendingReceipt := patchEvidenceViaWorkbook(t, workbookStore, actor, recordID, `{
		"view_schema_id":"cartulary.view.evidence.v1",
		"base_row_version":1,
		"client_txn_id":"txn-evidence_lifecycle-lifecycle-pending",
		"changes":[{"field_key":"evidence.lifecycle_state","value":"pending_receipt"}]
	}`)
	requireRowCellValue(t, pendingReceipt.Payload["row"].(map[string]any), "evidence.lifecycle_state", "pending_receipt")

	received := patchEvidenceViaWorkbook(t, workbookStore, actor, recordID, `{
		"view_schema_id":"cartulary.view.evidence.v1",
		"base_row_version":2,
		"client_txn_id":"txn-evidence_lifecycle-lifecycle-received",
		"changes":[{"field_key":"evidence.lifecycle_state","value":"received"}]
	}`)
	requireRowCellValue(t, received.Payload["row"].(map[string]any), "evidence.lifecycle_state", "received")

	requireIllegalLifecyclePatch(t, workbookStore, actor, recordID, `{
		"view_schema_id":"cartulary.view.evidence.v1",
		"base_row_version":3,
		"client_txn_id":"txn-evidence_lifecycle-lifecycle-available-without-blob",
		"changes":[{"field_key":"evidence.lifecycle_state","value":"available"}]
	}`)
	requireIllegalLifecyclePatch(t, workbookStore, actor, recordID, `{
		"view_schema_id":"cartulary.view.evidence.v1",
		"base_row_version":3,
		"client_txn_id":"txn-evidence_lifecycle-lifecycle-released-direct",
		"changes":[{"field_key":"evidence.lifecycle_state","value":"released"}]
	}`)

	attachTarget := createEvidenceViaWorkbook(t, workbookStore, actor, incident.ID, `{
		"client_txn_id":"txn-evidence_lifecycle-lifecycle-attach-target",
		"evidence.title":"Blob arrives later",
		"evidence.storage_ref":"ticket://collect-2",
		"evidence.collector_party_text":"IR collector",
		"evidence.source_party_text":"Endpoint owner"
	}`)
	attachRecordID := mustRowID(t, attachTarget.Payload["row"].(map[string]any))
	unrelatedHistoryBefore := DurableAttachCounts(t, harness.DB, recordID)
	attachHistoryBefore := DurableAttachCounts(t, harness.DB, attachRecordID)
	blobID := seedBlob(t, harness.DB, incident.ID, actor.ID, "available", BlobOptions{
		ByteSize: 4, ObservedSize: ptrInt64(4), ObservedSHA: ptrString(strings.Repeat("b", 64)), ObservedContentType: ptrString("text/plain"),
	})
	attachRequest := evidence.AttachBlobRequest{ObjectBlobID: blobID, BaseRowVersion: 1, ClientTxnID: "txn-evidence_lifecycle-lifecycle-attach"}
	attached, err := evidenceStore.AttachBlob(context.Background(), actor, attachRecordID, attachRequest, evidence.AttachBlobRequestHash(attachRequest), nil, "req-lifecycle-attach", time.Now().UTC())
	if err != nil {
		t.Fatalf("attach available blob to requested evidence: %v", err)
	}
	attachedRow := attached.Payload["row"].(map[string]any)
	requireRowCellValue(t, attachedRow, "evidence.lifecycle_state", "requested")
	requireRowCellValue(t, attachedRow, "evidence.upload_state", "available")
	requireRowCellValue(t, attachedRow, "evidence.collector_party_text", "IR collector")
	requireRowCellValue(t, attachedRow, "evidence.source_party_text", "Endpoint owner")
	requireRowCellNonEmpty(t, attachedRow, "evidence.received_at")
	if after := DurableAttachCounts(t, harness.DB, recordID); after != unrelatedHistoryBefore {
		t.Fatalf("later attach mutated unrelated custody history: before=%+v after=%+v", unrelatedHistoryBefore, after)
	}
	attachHistoryAfter := DurableAttachCounts(t, harness.DB, attachRecordID)
	if attachHistoryAfter.ChangeSets != attachHistoryBefore.ChangeSets+1 ||
		attachHistoryAfter.Mutations != attachHistoryBefore.Mutations+1 ||
		attachHistoryAfter.Revisions != attachHistoryBefore.Revisions+1 ||
		attachHistoryAfter.BlobLinks != 1 {
		t.Fatalf("attach history got before=%+v after=%+v, want exactly one attach mutation and one blob link", attachHistoryBefore, attachHistoryAfter)
	}
	requireChangeSetAttribution(t, harness.DB, attached.Payload["change_set_id"].(string), actor.ID, "evidence.attach_blob", attachRequest.ClientTxnID)

	available := patchEvidenceViaWorkbook(t, workbookStore, actor, attachRecordID, `{
		"view_schema_id":"cartulary.view.evidence.v1",
		"base_row_version":2,
		"client_txn_id":"txn-evidence_lifecycle-lifecycle-available-after-attach",
		"changes":[{"field_key":"evidence.lifecycle_state","value":"available"}]
	}`)
	requireRowCellValue(t, available.Payload["row"].(map[string]any), "evidence.lifecycle_state", "available")

	quarantined := createEvidenceViaWorkbook(t, workbookStore, actor, incident.ID, `{
		"client_txn_id":"txn-evidence_lifecycle-lifecycle-quarantined",
		"evidence.title":"Quarantined evidence"
	}`)
	quarantinedRecordID := mustRowID(t, quarantined.Payload["row"].(map[string]any))
	if _, err := harness.DB.Exec(context.Background(), `UPDATE evidence SET lifecycle_state = 'quarantined' WHERE record_id = $1`, quarantinedRecordID); err != nil {
		t.Fatalf("seed quarantined lifecycle: %v", err)
	}
	quarantinedBlobID := seedBlob(t, harness.DB, incident.ID, actor.ID, "available", BlobOptions{ByteSize: 1, ObservedSize: ptrInt64(1)})
	quarantinedAttach := evidence.AttachBlobRequest{ObjectBlobID: quarantinedBlobID, BaseRowVersion: 1, ClientTxnID: "txn-evidence_lifecycle-quarantined-attach"}
	if _, err := evidenceStore.AttachBlob(context.Background(), actor, quarantinedRecordID, quarantinedAttach, evidence.AttachBlobRequestHash(quarantinedAttach), nil, "req-quarantined-attach", time.Now().UTC()); !errors.Is(err, evidence.ErrEvidenceQuarantined) {
		t.Fatalf("quarantined evidence attach got %v want ErrEvidenceQuarantined", err)
	} else {
		requireAttachRejectedReason(t, err, evidence.AttachReasonEvidenceQuarantined)
	}

	closedBlobID := seedBlob(t, harness.DB, incident.ID, actor.ID, "available", BlobOptions{ByteSize: 1, ObservedSize: ptrInt64(1)})
	if _, err := harness.DB.Exec(context.Background(), `UPDATE incidents SET status = 'closed', closed_at = now() WHERE id = $1`, incident.ID); err != nil {
		t.Fatalf("close incident before quarantine: %v", err)
	}
	if _, err := evidenceStore.QuarantineBlob(context.Background(), actor.ID, closedBlobID, "admin_quarantine", "req-closed-quarantine", time.Now().UTC()); !admission.IsDenied(err, admission.DenialIncidentClosed) {
		t.Fatalf("closed-incident quarantine got %v, want typed incident-closed denial", err)
	}

	requireEvidenceState(t, harness.DB, attachRecordID, "available", "available", blobID)
}

func createEvidenceViaWorkbook(t testing.TB, store *workbook.Store, actor authn.UserRecord, incidentID uuid.UUID, body string) workbook.MutationResult {
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

func patchEvidenceViaWorkbook(t testing.TB, store *workbook.Store, actor authn.UserRecord, recordID uuid.UUID, body string) workbook.MutationResult {
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

func requireIllegalLifecyclePatch(t testing.TB, store *workbook.Store, actor authn.UserRecord, recordID uuid.UUID, body string) {
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

func requireChangeSetAttribution(t testing.TB, db Queryer, changeSetID string, actorID uuid.UUID, wantSource string, wantClientTxnID string) {
	t.Helper()
	var gotActor uuid.UUID
	var source string
	var clientTxnID string
	if err := db.QueryRow(context.Background(), `
SELECT actor_user_id, source, client_txn_id
  FROM change_sets
 WHERE change_set_id = $1
`, changeSetID).Scan(&gotActor, &source, &clientTxnID); err != nil {
		t.Fatalf("load change set attribution: %v", err)
	}
	if gotActor != actorID || source != wantSource || clientTxnID != wantClientTxnID {
		t.Fatalf("change set attribution got actor=%s source=%s client_txn_id=%s want actor=%s source=%s client_txn_id=%s", gotActor, source, clientTxnID, actorID, wantSource, wantClientTxnID)
	}
}

type Queryer interface {
	QueryRow(ctx context.Context, sql string, args ...any) pgx.Row
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
