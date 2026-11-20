package workbook_test

import (
	"context"
	"net/http"
	"reflect"
	"testing"

	"github.com/google/uuid"

	authflowtest "github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/workbookscenariotest"
)

// TestLinkedNoteCreationClassifiesItsSourceRecord_Integration verifies
// REQ-01-705 for POST /api/v1/records/{record_id}/linked-notes, which creates
// under the record its path addresses. After incident visibility and role, an
// exact committed replay returns its original receipt, even after the source
// is deleted and the incident closed. A soft-deleted source is 409
// record_deleted_use_restore before its type and the incident lifecycle, and a
// missing source is incident_not_found. A rejected request commits no Note,
// link, change set, receipt or publication, and a non-member cannot tell a
// deleted source from a missing one (REQ-01-704).
func TestLinkedNoteCreationClassifiesItsSourceRecord_Integration(t *testing.T) {
	harness, login, _, incidentID := ConflictFixture(t, "linked-note-target-classification", "IR-LINKED-NOTE-TARGET")
	createLinked := func(caller authcookietest.AuthCookies, sourceID uuid.UUID, txn string) *http.Response {
		t.Helper()
		return httptestx.DoJSON(t, http.DefaultClient, http.MethodPost,
			harness.Server.HTTP.URL+"/api/v1/records/"+sourceID.String()+"/linked-notes",
			map[string]any{"client_txn_id": txn, "note.title": "Linked " + txn, "note.body": "Linked body"},
			httptestx.WithCookies(caller.Session, caller.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, caller.CSRF.Value),
		)
	}
	type effects struct {
		history workbookscenariotest.MutationEffects
		links   int
	}
	snapshot := func() effects {
		t.Helper()
		return effects{
			history: workbookscenariotest.SnapshotIncidentMutationEffects(t, harness.DB, incidentID.String()),
			links:   appsupport.QueryCount(t, harness.DB, `SELECT count(*) FROM record_links WHERE incident_id = $1`, incidentID),
		}
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
	requireReplay := func(name string, sourceID uuid.UUID, txn string, want map[string]any) {
		t.Helper()
		before := snapshot()
		got := httptestx.RequireSuccessEnvelope(t, createLinked(login, sourceID, txn), http.StatusOK)["data"].(map[string]any)
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("%s: replay = %#v, want the original receipt %#v", name, got, want)
		}
		if after := snapshot(); after != before {
			t.Fatalf("%s replay committed effects: %+v -> %+v", name, before, after)
		}
	}
	createHost := func(txn, hostname string) uuid.UUID {
		t.Helper()
		row := requireWorkbookCreate(t, harness, login, incidentID, "cartulary.view.hosts.v1", map[string]any{
			"client_txn_id": txn, "host.display_name": "Linked source " + hostname, "host.hostname": hostname,
		})["row"].(map[string]any)
		return appsupport.MustUUID(t, row["record_id"].(string))
	}
	deleteRecord := func(recordID uuid.UUID, txn string) {
		t.Helper()
		version := appsupport.QueryCount(t, harness.DB, `SELECT row_version FROM records WHERE record_id = $1`, recordID)
		httptestx.RequireSuccessEnvelope(t, deleteRecordViaWorkbookRoute(t, harness, login, recordID, map[string]any{
			"base_row_version": version, "client_txn_id": txn,
		}), http.StatusOK)
	}

	sourceID := createHost("ln-source-create", "linked-note-source")
	liveID := createHost("ln-live-create", "linked-note-live")
	committed := httptestx.RequireSuccessEnvelope(t, createLinked(login, sourceID, "ln-committed"), http.StatusCreated)["data"].(map[string]any)
	if committed["source_record_id"] != sourceID.String() || committed["change_set_id"] == nil {
		t.Fatalf("committed linked Note = %#v", committed)
	}
	deleteRecord(sourceID, "ln-source-delete")

	requireReplay("replay after source deletion", sourceID, "ln-committed", committed)
	requireRejected("deleted source", http.StatusConflict, "record_deleted_use_restore", func() *http.Response {
		return createLinked(login, sourceID, "ln-deleted-fresh")
	})
	missingID := uuid.New()
	missing := requireRejected("missing source", http.StatusNotFound, "incident_not_found", func() *http.Response {
		return createLinked(login, missingID, "ln-missing")
	})

	// The deleted state precedes the record type: a Note cannot hold linked
	// Notes, yet a deleted Note must still be restored first.
	note := CreateNote(t, harness, login, incidentID, "ln-note-create", "Unlinkable Note", "Unlinkable body")
	noteID := appsupport.MustUUID(t, note["record_id"].(string))
	requireRejected("live unsupported type", http.StatusNotFound, "incident_not_found", func() *http.Response {
		return createLinked(login, noteID, "ln-note-live")
	})
	deleteRecord(noteID, "ln-note-delete")
	requireRejected("deleted unsupported type", http.StatusConflict, "record_deleted_use_restore", func() *http.Response {
		return createLinked(login, noteID, "ln-note-deleted")
	})

	// REQ-01-704: visibility precedes the deleted state.
	outsider := authflowtest.SeedLocalUserRecord(t, harness.DB, "linked-note-outsider@example.test", "Outsider", "OutsiderPass1!", false, false, true)
	outsiderLogin := authflowtest.LoginLocalUser(t, http.DefaultClient, harness.Server.HTTP.URL, outsider.Email, "OutsiderPass1!", nil)
	concealed := requireRejected("non-member deleted source", http.StatusNotFound, "incident_not_found", func() *http.Response {
		return createLinked(outsiderLogin, sourceID, "ln-outsider")
	})
	if !reflect.DeepEqual(concealed, missing) {
		t.Fatalf("non-member response %#v differs from the missing-source response %#v", concealed, missing)
	}

	if _, err := harness.DB.ExecContext(context.Background(), `UPDATE incidents SET status = 'closed', closed_at = now() WHERE id = $1`, incidentID); err != nil {
		t.Fatalf("close incident: %v", err)
	}
	requireRejected("live source after close", http.StatusConflict, "incident_closed", func() *http.Response {
		return createLinked(login, liveID, "ln-closed-live")
	})
	requireRejected("deleted source after close", http.StatusConflict, "record_deleted_use_restore", func() *http.Response {
		return createLinked(login, sourceID, "ln-closed-deleted")
	})
	requireReplay("replay after deletion and closure", sourceID, "ln-committed", committed)
}
