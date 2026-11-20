package evidence_test

import (
	"context"
	"crypto/sha256"
	"fmt"
	"net/http"
	"reflect"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/records/testsupport/envelopetest"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/workbookscenariotest"
)

// TestEvidenceAttachClassifiesItsTargetRecord_Integration verifies REQ-01-705
// for POST /api/v1/evidence-records/{record_id}/attach-blob. After incident
// visibility and role, an exact committed attach returns its original receipt
// before every other gate: after its Blob is quarantined, after the Evidence
// record is deleted, and after the incident closes, whether the record is live
// or deleted (REQ-01-245, REQ-01-591). A fresh request on the quarantined Blob
// is still evidence_attach_rejected, and a divergent request under the same key
// is client_txn_conflict. A soft-deleted record is 409
// record_deleted_use_restore before its type and the incident lifecycle, and a
// missing record is evidence_record_not_found. A replay or rejected request
// attaches no Blob and commits no change set, receipt, custody event or
// publication, and a non-member cannot tell a deleted record from a missing
// one (REQ-01-704).
func TestEvidenceAttachClassifiesItsTargetRecord_Integration(t *testing.T) {
	harness := appsupport.StartServer(t, appsupport.ServerOptions{Prefix: "evidence-attach-target", TestRouteMode: httptestx.TestRouteModeDisabled})
	login, adminID := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, harness.Server.HTTP.URL)
	incident := appsupport.CreateIncident(t, harness.Server, login, map[string]any{
		"client_txn_id": "txn-evidence-attach-target-incident",
		"incident_key":  "IR-EVIDENCE-ATTACH-TARGET",
		"title":         "Evidence attach target classification",
	})
	incidentID := appsupport.MustUUID(t, incident["incident_id"].(string))
	attach := func(caller authcookietest.AuthCookies, recordID uuid.UUID, blobID any, base int, txn string) *http.Response {
		t.Helper()
		return httptestx.DoJSON(t, http.DefaultClient, http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/evidence-records/"+recordID.String()+"/attach-blob",
			map[string]any{"object_blob_id": blobID, "base_row_version": base, "client_txn_id": txn}, authOptions(caller)...)
	}
	verifiedBlob := func(txn string) any {
		t.Helper()
		payload := []byte("attach target " + txn)
		sum := sha256.Sum256(payload)
		created := httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/object-blobs", map[string]any{
			"incident_id": incidentID.String(), "client_txn_id": txn, "byte_size": len(payload),
			"filename_hint": "target.txt", "content_type_hint": "text/plain", "sha256_hex": fmt.Sprintf("%x", sum[:]),
		}, authOptions(login)...), http.StatusCreated)["data"].(map[string]any)
		putObject(t, harness.Server.HTTP.URL, created["upload_target"].(map[string]any)["href"].(string), payload, "text/plain", login)
		verifyUploadedBlob(t, harness, login, created["object_blob_id"], txn+"-verify", "succeeded")
		return created["object_blob_id"]
	}
	type effects struct {
		history  workbookscenariotest.MutationEffects
		evidence string
	}
	snapshot := func() effects {
		t.Helper()
		var evidence string
		if err := harness.DB.QueryRowContext(context.Background(), `SELECT jsonb_build_object(
 'evidence',(SELECT coalesce(jsonb_agg(to_jsonb(e) ORDER BY record_id),'[]') FROM evidence e WHERE incident_id=$1),
 'blobs',(SELECT coalesce(jsonb_agg(to_jsonb(b) ORDER BY object_blob_id),'[]') FROM object_blobs b WHERE incident_id=$1),
 'custody',(SELECT count(*) FROM evidence_custody_events WHERE incident_id=$1)
)::text`, incidentID).Scan(&evidence); err != nil {
			t.Fatalf("snapshot Evidence attachment state: %v", err)
		}
		return effects{history: workbookscenariotest.SnapshotIncidentMutationEffects(t, harness.DB, incidentID.String()), evidence: evidence}
	}
	requireRejected := func(name string, status int, code string, send func() *http.Response) map[string]any {
		t.Helper()
		before := snapshot()
		response := send()
		if response.StatusCode != status {
			t.Fatalf("%s: status %d, want %d %s; body=%#v", name, response.StatusCode, status, code, httptestx.ReadJSONBody(t, response))
		}
		body := httptestx.RequireErrorEnvelope(t, response, status, code)
		if after := snapshot(); after != before {
			t.Fatalf("%s committed effects: %+v -> %+v", name, before, after)
		}
		// Everything but the per-request identifier must match between
		// responses that conceal a record.
		errorObject := body["error"].(map[string]any)
		delete(errorObject, "request_id")
		return errorObject
	}
	requireReplay := func(name string, recordID uuid.UUID, blobID any, txn string, want map[string]any) {
		t.Helper()
		before := snapshot()
		got := httptestx.RequireSuccessEnvelope(t, attach(login, recordID, blobID, 1, txn), http.StatusOK)["data"].(map[string]any)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("%s: replay = %#v, want the original receipt %#v", name, got, want)
		}
		if after := snapshot(); after != before {
			t.Fatalf("%s replay committed effects: %+v -> %+v", name, before, after)
		}
	}

	committedID, deletedID, liveID, liveCommittedID := uuid.New(), uuid.New(), uuid.New(), uuid.New()
	for _, recordID := range []uuid.UUID{committedID, deletedID, liveID, liveCommittedID} {
		seedEvidenceRecord(t, harness, incidentID, adminID, recordID)
	}
	committedBlob := verifiedBlob("attach-target-committed")
	committed := httptestx.RequireSuccessEnvelope(t, attach(login, committedID, committedBlob, 1, "attach-target-committed"), http.StatusOK)["data"].(map[string]any)
	if committed["object_blob_id"] != committedBlob || committed["change_set_id"] == nil {
		t.Fatalf("committed attach = %#v", committed)
	}
	liveBlob := verifiedBlob("attach-target-live-committed")
	liveCommitted := httptestx.RequireSuccessEnvelope(t, attach(login, liveCommittedID, liveBlob, 1, "attach-target-live-committed"), http.StatusOK)["data"].(map[string]any)
	if liveCommitted["object_blob_id"] != liveBlob || liveCommitted["change_set_id"] == nil {
		t.Fatalf("live committed attach = %#v", liveCommitted)
	}
	envelopetest.SoftDeleteRecord(t, harness.DB, envelopetest.Tombstone{RecordID: committedID, DeletedBy: adminID})
	envelopetest.SoftDeleteRecord(t, harness.DB, envelopetest.Tombstone{RecordID: deletedID, DeletedBy: adminID})
	freshBlob := verifiedBlob("attach-target-fresh")

	// REQ-01-245: Blob eligibility binds fresh requests only. An exact replay
	// returns its receipt over a quarantined Blob, while a fresh request and a
	// divergent reuse of the key are still rejected.
	quarantine := evidence.NewBlobQuarantinerForTest(t, harness.Pool)
	if err := quarantine(context.Background(), adminID, appsupport.MustUUID(t, liveBlob.(string))); err != nil {
		t.Fatalf("quarantine live committed Blob: %v", err)
	}
	requireReplay("replay after quarantine", liveCommittedID, liveBlob, "attach-target-live-committed", liveCommitted)
	quarantined := requireRejected("fresh attach of a quarantined Blob", http.StatusConflict, "evidence_attach_rejected", func() *http.Response {
		return attach(login, liveCommittedID, liveBlob, 2, "attach-target-live-fresh")
	})
	if reason := quarantined["details"].(map[string]any)["reason_code"]; reason != "blob_quarantined" {
		t.Fatalf("fresh attach of a quarantined Blob reason = %#v, want blob_quarantined", reason)
	}
	requireRejected("divergent replay after quarantine", http.StatusConflict, "client_txn_conflict", func() *http.Response {
		return attach(login, liveCommittedID, liveBlob, 2, "attach-target-live-committed")
	})

	requireReplay("replay after deletion", committedID, committedBlob, "attach-target-committed", committed)
	requireRejected("deleted Evidence record", http.StatusConflict, "record_deleted_use_restore", func() *http.Response {
		return attach(login, deletedID, freshBlob, 1, "attach-target-deleted")
	})
	missingID := uuid.New()
	missing := requireRejected("missing record", http.StatusNotFound, "evidence_record_not_found", func() *http.Response {
		return attach(login, missingID, freshBlob, 1, "attach-target-missing")
	})

	// The deleted state precedes the record type: a Note is not Evidence, yet
	// a deleted Note must still be restored first.
	note := httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(t, http.DefaultClient, http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID.String()+"/views/cartulary.view.notes.v1/rows",
		map[string]any{"client_txn_id": "attach-target-note", "note.title": "Not Evidence"}, authOptions(login)...,
	), http.StatusCreated)["data"].(map[string]any)["row"].(map[string]any)
	noteID := appsupport.MustUUID(t, note["record_id"].(string))
	requireRejected("live non-Evidence record", http.StatusNotFound, "evidence_record_not_found", func() *http.Response {
		return attach(login, noteID, freshBlob, 1, "attach-target-note-live")
	})
	httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(t, http.DefaultClient, http.MethodDelete,
		harness.Server.HTTP.URL+"/api/v1/records/"+noteID.String(),
		map[string]any{"base_row_version": 1, "client_txn_id": "attach-target-note-delete"}, authOptions(login)...,
	), http.StatusOK)
	requireRejected("deleted non-Evidence record", http.StatusConflict, "record_deleted_use_restore", func() *http.Response {
		return attach(login, noteID, freshBlob, 2, "attach-target-note-deleted")
	})

	// REQ-01-704: visibility precedes the deleted state.
	outsider := flowtest.SeedLocalUserRecord(t, harness.DB, "evidence-attach-outsider@example.test", "Outsider", "OutsiderPass1!", false, false, true)
	outsiderLogin := flowtest.LoginLocalUser(t, http.DefaultClient, harness.Server.HTTP.URL, outsider.Email, "OutsiderPass1!", nil)
	concealed := requireRejected("non-member deleted record", http.StatusNotFound, "evidence_record_not_found", func() *http.Response {
		return attach(outsiderLogin, deletedID, freshBlob, 1, "attach-target-outsider")
	})
	if !reflect.DeepEqual(concealed, missing) {
		t.Fatalf("non-member response %#v differs from the missing-record response %#v", concealed, missing)
	}

	if _, err := harness.DB.ExecContext(context.Background(), `UPDATE incidents SET status = 'closed', closed_at = now() WHERE id = $1`, incidentID); err != nil {
		t.Fatalf("close incident: %v", err)
	}
	requireRejected("live Evidence record after close", http.StatusConflict, "incident_closed", func() *http.Response {
		return attach(login, liveID, freshBlob, 1, "attach-target-closed-live")
	})
	requireRejected("deleted Evidence record after close", http.StatusConflict, "record_deleted_use_restore", func() *http.Response {
		return attach(login, deletedID, freshBlob, 1, "attach-target-closed-deleted")
	})
	requireReplay("replay after deletion and closure", committedID, committedBlob, "attach-target-committed", committed)
	requireReplay("replay after closure of a live record", liveCommittedID, liveBlob, "attach-target-live-committed", liveCommitted)
}
