package workbook_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

// TestWorkbookRowVersionConflictDetails_Integration verifies, through the
// public single-record routes, that every row_version_conflict names the
// record and both versions (REQ-01-075) and commits nothing. A Timeline
// supersede or mark-reviewed with a stale base version returns it (AC-614),
// and so does a stale patch whose retained revision history cannot rebuild the
// stale window, for every owner of a mutable record type: unusable history
// fails closed (REQ-01-070) and is never an internal error.
func TestWorkbookRowVersionConflictDetails_Integration(t *testing.T) {
	harness, login, _, incidentID := ConflictFixture(t, "workbook-row-version-conflict-details", "IR-ROW-VERSION-CONFLICT-DETAILS")
	recordsURL := harness.Server.HTTP.URL + "/api/v1/records/"
	authOptions := []func(*http.Request){httptestx.WithCookies(login.Session, login.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, login.CSRF.Value)}
	type effects struct{ changeSets, receipts, version int }
	snapshot := func(recordID uuid.UUID) effects {
		t.Helper()
		return effects{
			changeSets: appsupport.QueryCount(t, harness.DB, `SELECT count(*) FROM change_sets WHERE incident_id = $1`, incidentID),
			receipts:   appsupport.QueryCount(t, harness.DB, `SELECT count(*) FROM route_idempotency WHERE scope_key = $1`, recordID.String()),
			version:    appsupport.QueryCount(t, harness.DB, `SELECT COALESCE(max(row_version), 0) FROM records WHERE record_id = $1`, recordID),
		}
	}
	requireRowVersionConflict := func(t *testing.T, recordID uuid.UUID, base, current int64, send func() *http.Response) {
		t.Helper()
		before := snapshot(recordID)
		response := send()
		if response.StatusCode != http.StatusConflict {
			t.Fatalf("status %d, want 409 row_version_conflict; body=%#v", response.StatusCode, httptestx.ReadJSONBody(t, response))
		}
		details, _ := httptestx.RequireErrorEnvelope(t, response, http.StatusConflict, "row_version_conflict")["error"].(map[string]any)["details"].(map[string]any)
		for key, want := range map[string]any{"record_id": recordID.String(), "base_row_version": float64(base), "current_row_version": float64(current)} {
			if details[key] != want {
				t.Fatalf("row_version_conflict details = %#v, want %s=%v", details, key, want)
			}
		}
		if after := snapshot(recordID); after != before {
			t.Fatalf("row_version_conflict committed effects: %+v -> %+v", before, after)
		}
	}
	createAndAdvance := func(t *testing.T, view, field, prefix string, create map[string]any) uuid.UUID {
		t.Helper()
		body := map[string]any{"client_txn_id": prefix + "-create"}
		for key, value := range create {
			body[key] = value
		}
		created := requireWorkbookCreate(t, harness, login, incidentID, view, body)["row"].(map[string]any)
		recordID := appsupport.MustUUID(t, created["record_id"].(string))
		requireWorkbookPatch(t, harness, login, recordID, map[string]any{
			"view_schema_id": view, "base_row_version": 1, "client_txn_id": prefix + "-saved",
			"changes": []map[string]any{{"field_key": field, "value": "Saved " + prefix}},
		})
		return recordID
	}

	t.Run("timeline lifecycle actions with a stale base version", func(t *testing.T) {
		recordID := createAndAdvance(t, TimelineViewSchemaID, "timeline.activity_synopsis_text", "rvd-lifecycle", map[string]any{"timeline.activity_synopsis_text": "Lifecycle base"})
		for _, action := range []string{"mark-reviewed", "supersede"} {
			t.Run(action, func(t *testing.T) {
				requireRowVersionConflict(t, recordID, 1, 2, func() *http.Response {
					return httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, recordsURL+recordID.String()+"/"+action, map[string]any{
						"base_row_version": 1, "client_txn_id": "rvd-stale-" + action, "reason": "Stale lifecycle base",
					}, authOptions...)
				})
			})
		}
	})

	owners := []struct {
		name, view, field string
		create            map[string]any
	}{
		{"timeline", TimelineViewSchemaID, "timeline.activity_synopsis_text", map[string]any{"timeline.activity_synopsis_text": "Timeline base"}},
		{"note", "cartulary.view.notes.v1", "note.title", map[string]any{"note.title": "Note base", "note.body": "Note body"}},
		{"evidence", "cartulary.view.evidence.v1", "evidence.title", map[string]any{"evidence.title": "Evidence base"}},
		{"host", "cartulary.view.hosts.v1", "host.display_name", map[string]any{"host.display_name": "Host base", "host.hostname": "details-host"}},
		{"identity", "cartulary.view.identities.v1", "identity.display_name", map[string]any{"identity.display_name": "Identity base", "identity.email": "details@example.test"}},
		{"party", "cartulary.view.parties.v1", "party.display_name", map[string]any{"party.display_name": "Party base", "party.party_kind": "person"}},
		{"task", "cartulary.view.task_requests.v1", "task.title", map[string]any{"task.title": "Task base", "task.task_kind": "question"}},
		{"decision", "cartulary.view.decisions.v1", "decision.summary", map[string]any{"decision.summary": "Decision base", "decision.decision_type": "scope", "decision.rationale": "Details fixture"}},
	}
	for _, owner := range owners {
		t.Run(owner.name+"/unusable revision history", func(t *testing.T) {
			prefix := "rvd-history-" + owner.name
			recordID := createAndAdvance(t, owner.view, owner.field, prefix, owner.create)
			removed, err := harness.DB.ExecContext(context.Background(), `DELETE FROM record_revisions WHERE record_id = $1 AND row_version = 1`, recordID)
			if err != nil {
				t.Fatalf("remove base revision: %v", err)
			}
			if count, err := removed.RowsAffected(); err != nil || count != 1 {
				t.Fatalf("removed base revisions = %d, %v; want 1", count, err)
			}
			requireRowVersionConflict(t, recordID, 1, 2, func() *http.Response {
				return doWorkbookJSON(t, harness, login, http.MethodPatch, uuid.Nil, "", recordID, map[string]any{
					"view_schema_id": owner.view, "base_row_version": 1, "client_txn_id": prefix + "-stale",
					"changes": []map[string]any{{"field_key": owner.field, "value": "Stale " + prefix}},
				})
			})
		})
	}
}
