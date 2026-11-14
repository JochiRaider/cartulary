package revisions_test

import (
	"net/http"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/google/uuid"
)

func requireDecisionDeleteSupersession(t *testing.T, h *appsupport.ServerHarness, login authcookietest.AuthCookies, incident, actor uuid.UUID) {
	for _, variant := range []string{"superseded peer", "executed peer", "delete target", "unrelated"} {
		t.Run("Decision delete supersession/"+variant, func(t *testing.T) {
			options := []func(*http.Request){httptestx.WithCookies(login.Session, login.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, login.CSRF.Value)}
			create := func(summary, status string) uuid.UUID {
				t.Helper()
				body := map[string]any{"client_txn_id": uuid.NewString(), "decision.summary": summary, "decision.status": status, "decision.owner_user_id": actor.String(), "decision.decision_type": "containment", "decision.decided_at": "2026-11-13T12:00:00Z", "decision.rationale": "Retained rationale"}
				data := httptestx.RequireSuccessData(t, httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, h.Server.HTTP.URL+"/api/v1/incidents/"+incident.String()+"/views/cartulary.view.decisions.v1/rows", body, options...), http.StatusCreated)
				return uuid.MustParse(data["row"].(map[string]any)["record_id"].(string))
			}
			status := "proposed"
			if variant == "executed peer" {
				status = "executed"
			}
			target, source := create("Original decision", status), create("Replacement decision", "approved")
			base := 1
			if variant != "unrelated" {
				httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, h.Server.HTTP.URL+"/api/v1/records/"+target.String()+"/supersede", map[string]any{"base_row_version": 1, "client_txn_id": uuid.NewString(), "reason": "Replace incident decision", "replacement_record_id": source.String()}, options...), http.StatusOK)
				base = 2
			}
			selected := source
			if variant == "delete target" {
				selected = target
			}
			linksBefore := decisionSupersessionLinks(t, h, incident)
			before := snapshotDecisionReferenceState(t, h, incident, h.Pool.QueryRow)
			body := map[string]any{"base_row_version": base, "client_txn_id": uuid.NewString()}
			response := deleteRecord(t, h, login, selected, body)
			if variant == "superseded peer" {
				failure := httptestx.RequireErrorEnvelope(t, response, http.StatusConflict, "record_delete_blocked")["error"].(map[string]any)
				details := failure["details"].(map[string]any)
				if failure["retryable"] != false || len(details) != 2 || details["record_id"] != source.String() || details["reason_code"] != "decision_supersession_dependency" {
					t.Fatalf("Decision lifecycle blocker=%#v", failure)
				}
				if snapshotDecisionReferenceState(t, h, incident, h.Pool.QueryRow) != before {
					t.Fatal("blocked superseder deletion changed retained state")
				}
				httptestx.RequireSuccessEnvelope(t, deleteRecord(t, h, login, target, map[string]any{"base_row_version": 2, "client_txn_id": uuid.NewString()}), http.StatusOK)
				response = deleteRecord(t, h, login, selected, body)
			}
			result := httptestx.RequireSuccessEnvelope(t, response, http.StatusOK)["data"].(map[string]any)
			if result["deleted"] != true || result["row_version"] != float64(base+1) {
				t.Fatalf("Decision deletion=%#v", result)
			}
			if decisionSupersessionLinks(t, h, incident) != linksBefore {
				t.Fatal("Decision deletion changed authoritative Links")
			}
			expectedTarget := status
			if variant != "unrelated" && status != "executed" {
				expectedTarget = "superseded"
			}
			var actualSource, actualTarget string
			if err := h.DB.QueryRow(`SELECT status FROM decisions WHERE record_id=$1`, source).Scan(&actualSource); err != nil {
				t.Fatal(err)
			}
			if err := h.DB.QueryRow(`SELECT status FROM decisions WHERE record_id=$1`, target).Scan(&actualTarget); err != nil {
				t.Fatal(err)
			}
			if actualSource != "approved" || actualTarget != expectedTarget {
				t.Fatalf("deletion rewrote Decision lifecycle: %s/%s", actualSource, actualTarget)
			}
			if countRecordMutations(t, h.DB, selected, "soft_delete") != 1 || countRecordRevisions(t, h.DB, selected) != base+1 {
				t.Fatal("Decision lifecycle deletion appended unexpected history")
			}
			requireSingleRecordChangeIntent(t, h.DB, result["change_set_id"].(string), selected, "remove")
		})
	}
}

func decisionSupersessionLinks(t testing.TB, h *appsupport.ServerHarness, incident uuid.UUID) string {
	t.Helper()
	var state string
	if err := h.DB.QueryRow(`SELECT coalesce(jsonb_agg(to_jsonb(l) ORDER BY record_link_id),'[]'::jsonb)::text FROM record_links l WHERE incident_id=$1`, incident).Scan(&state); err != nil {
		t.Fatal(err)
	}
	return state
}
