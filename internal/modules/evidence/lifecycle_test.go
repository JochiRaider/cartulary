package evidence_test

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
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
	actor := authstoretest.SeedActor(t, harness.DB, authstoretest.ActorAttributes{Email: "evidence_lifecycle-lifecycle@example.test", DisplayName: "EvidenceLifecycle Lifecycle"})
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

	t.Run("bound text persists and clears canonically", func(t *testing.T) {
		result := createEvidenceViaOwner(t, evidenceOwner, actor, incident.ID, `{
   "client_txn_id":"evidence-normalized-text",
   "evidence.title":" Cafe\u0301  Title\u200d ",
   "evidence.storage_ref":" ticket://Cafe\u0301\u200d ",
   "evidence.collector_party_text":" Cafe\u0301  Collector\u200d ",
   "evidence.source_party_text":" Cafe\u0301  Source\u200d "
  }`)
		id := mustRowID(t, result.Row)
		check := func(expected map[string]any, version int64) {
			t.Helper()
			var raw []byte
			var gotVersion int64
			if err := harness.DB.QueryRow(context.Background(), `SELECT jsonb_build_object('title',e.title,'storage_ref',e.storage_ref,'collector_party_text',e.collector_party_text,'source_party_text',e.source_party_text), r.row_version FROM evidence e JOIN records r USING(record_id) WHERE e.record_id=$1`, id).Scan(&raw, &gotVersion); err != nil {
				t.Fatal(err)
			}
			var got map[string]any
			if err := json.Unmarshal(raw, &got); err != nil {
				t.Fatal(err)
			}
			if !reflect.DeepEqual(got, expected) || gotVersion != version {
				t.Fatalf("stored text=%#v version=%d want=%#v/%d", got, gotVersion, expected, version)
			}
		}
		check(map[string]any{"title": "Café  Title\u200d", "storage_ref": "ticket://Café\u200d", "collector_party_text": "Café  Collector\u200d", "source_party_text": "Café  Source\u200d"}, 1)
		cleared := patchEvidenceViaOwner(t, evidenceOwner, actor, id, `{
   "view_schema_id":"cartulary.view.evidence.v1","base_row_version":1,"client_txn_id":"evidence-clear-text",
   "changes":[{"field_key":"evidence.title","value":" \u00a0 "},{"field_key":"evidence.storage_ref","value":""},{"field_key":"evidence.collector_party_text","value":"\u2003"},{"field_key":"evidence.source_party_text","value":null}]
  }`)
		expected := map[string]any{"title": nil, "storage_ref": nil, "collector_party_text": nil, "source_party_text": nil}
		check(expected, 2)
		replay := patchEvidenceViaOwner(t, evidenceOwner, actor, id, `{
   "view_schema_id":"cartulary.view.evidence.v1","base_row_version":1,"client_txn_id":"evidence-clear-text",
   "changes":[{"field_key":"evidence.title","value":null},{"field_key":"evidence.storage_ref","value":null},{"field_key":"evidence.collector_party_text","value":null},{"field_key":"evidence.source_party_text","value":null}]
  }`)
		if replay.Outcome != evidence.MutationOutcomeReplayed || cleared.ChangeSetID == nil || replay.ChangeSetID == nil || *cleared.ChangeSetID != *replay.ChangeSetID {
			t.Fatalf("normalized clear replay identity/outcome changed: first=%v replay=%v outcome=%s", cleared.ChangeSetID, replay.ChangeSetID, replay.Outcome)
		}
		firstJSON, err := json.Marshal(cleared.Row)
		if err != nil {
			t.Fatal(err)
		}
		replayJSON, err := json.Marshal(replay.Row)
		if err != nil {
			t.Fatal(err)
		}
		if string(firstJSON) != string(replayJSON) {
			t.Fatalf("normalized clear replay row changed: first=%s replay=%s", firstJSON, replayJSON)
		}
		check(expected, 2)
	})
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
