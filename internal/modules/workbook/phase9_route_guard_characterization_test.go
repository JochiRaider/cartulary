package workbook_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/phase4test"
)

func TestSupportPhase9_WorkbookRouteGuardsFailBeforeMutation(t *testing.T) {
	harness := phase4test.StartServer(t, "phase9-workbook-route-guards")
	adminLogin, _ := phase4test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase4test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-phase9-workbook-route-guards-incident",
		"incident_key":  "IR-PHASE9-ROUTE-GUARDS",
		"title":         "Phase 9 workbook route guards",
	})
	incidentID := phase4test.MustUUID(t, incident["incident_id"].(string))

	createURL := harness.Server.HTTP.URL + "/api/v1/incidents/" + incidentID.String() + "/views/cartulary.view.notes.v1/rows"
	socket := phase4test.ConnectViewSocket(t, harness.Server, incidentID.String(), "cartulary.view.notes.v1", adminLogin.SessionCookie.Value)
	defer socket.Close(1000, "test_complete")

	beforeAuth := snapshotWorkbookRouteGuardState(t, harness, incidentID)
	body := map[string]any{
		"client_txn_id": "txn-phase9-workbook-route-guards-auth",
		"note.title":    "guarded note",
	}

	unauthenticated := phase4test.DoJSON(t, http.MethodPost, createURL, body)
	httptestx.RequireErrorEnvelope(t, unauthenticated, http.StatusUnauthorized, "session_required")

	missingCSRF := phase4test.DoJSON(
		t,
		http.MethodPost,
		createURL,
		body,
		phase4test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
	)
	httptestx.RequireErrorEnvelope(t, missingCSRF, http.StatusForbidden, "csrf_verification_failed")

	invalidCSRF := phase4test.DoJSON(
		t,
		http.MethodPost,
		createURL,
		body,
		phase4test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase4test.WithHeader(authn.CSRFHeaderName, "wrong-csrf-token"),
	)
	httptestx.RequireErrorEnvelope(t, invalidCSRF, http.StatusForbidden, "csrf_verification_failed")

	afterAuth := snapshotWorkbookRouteGuardState(t, harness, incidentID)
	if afterAuth != beforeAuth {
		t.Fatalf("auth/csrf failures mutated workbook state: before=%+v after=%+v", beforeAuth, afterAuth)
	}
	phase4test.ExpectNoSocketMessage(t, socket)

	noteData := requireWorkbookCreate(t, harness, adminLogin, incidentID, "cartulary.view.notes.v1", map[string]any{
		"client_txn_id": "txn-phase9-workbook-route-guards-note",
		"note.title":    "closed incident guard note",
	})
	recordID := phase4test.MustUUID(t, noteData["row"].(map[string]any)["record_id"].(string))
	closedSocket := phase4test.ConnectViewSocket(t, harness.Server, incidentID.String(), "cartulary.view.notes.v1", adminLogin.SessionCookie.Value)
	defer closedSocket.Close(1000, "test_complete")

	if _, err := harness.DB.ExecContext(context.Background(), `
UPDATE incidents
   SET status = 'closed',
       closed_at = $1
 WHERE id = $2
`, time.Date(2026, 6, 30, 12, 0, 0, 0, time.UTC), incidentID); err != nil {
		t.Fatalf("close incident: %v", err)
	}

	beforeClosed := snapshotWorkbookRouteGuardState(t, harness, incidentID)
	closedCreate := doWorkbookJSON(t, harness, adminLogin, http.MethodPost, incidentID, "cartulary.view.notes.v1", uuid.Nil, map[string]any{
		"client_txn_id": "txn-phase9-workbook-route-guards-closed-create",
		"note.title":    "blocked after close",
	})
	httptestx.RequireErrorEnvelope(t, closedCreate, http.StatusConflict, "incident_closed")

	closedPatch := doWorkbookJSON(t, harness, adminLogin, http.MethodPatch, uuid.Nil, "", recordID, map[string]any{
		"view_schema_id":   "cartulary.view.notes.v1",
		"base_row_version": 1,
		"client_txn_id":    "txn-phase9-workbook-route-guards-closed-patch",
		"changes": []map[string]any{{
			"field_key": "note.title",
			"value":     "blocked patch after close",
		}},
	})
	httptestx.RequireErrorEnvelope(t, closedPatch, http.StatusConflict, "incident_closed")

	afterClosed := snapshotWorkbookRouteGuardState(t, harness, incidentID)
	if afterClosed != beforeClosed {
		t.Fatalf("closed incident failures mutated workbook state: before=%+v after=%+v", beforeClosed, afterClosed)
	}
	phase4test.ExpectNoSocketMessage(t, closedSocket)
}

type workbookRouteGuardState struct {
	Records         int
	Artifacts       int
	ChangeSets      int
	Mutations       int
	RecordRevisions int
}

func snapshotWorkbookRouteGuardState(t testing.TB, harness *phase4test.ServerHarness, incidentID uuid.UUID) workbookRouteGuardState {
	t.Helper()
	return workbookRouteGuardState{
		Records: phase4test.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM records
 WHERE incident_id = $1
`, incidentID),
		Artifacts: phase4test.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM artifacts
 WHERE incident_id = $1
`, incidentID),
		ChangeSets: phase4test.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM change_sets
 WHERE incident_id = $1
`, incidentID),
		Mutations: phase4test.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM change_set_mutations m
  JOIN change_sets c ON c.change_set_id = m.change_set_id
 WHERE c.incident_id = $1
`, incidentID),
		RecordRevisions: phase4test.QueryCount(t, harness.DB, `
SELECT COUNT(*)
  FROM record_revisions rr
  JOIN records r ON r.record_id = rr.record_id
 WHERE r.incident_id = $1
`, incidentID),
	}
}
