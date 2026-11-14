package revisions_test

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/google/uuid"
)

func requireDecisionRestoreSupersession(t *testing.T, h *appsupport.ServerHarness, login authcookietest.AuthCookies, incident, actor uuid.UUID) {
	cases := []struct {
		name                                               string
		targetStatus                                       string
		restoreTarget, deleteTarget, deleteSource, blocked bool
	}{
		{"target with live replacement", "proposed", true, true, false, false},
		{"selected source lock contention", "proposed", true, true, false, false},
		{"peer source lock contention", "proposed", true, true, false, false},
		{"target without retained relation", "proposed", true, true, false, true},
		{"target with deleted replacement", "proposed", true, true, true, true},
		{"replacement with deleted target", "proposed", false, true, true, false},
		{"executed target with deleted replacement", "executed", true, true, true, false},
		{"replacement with live executed target", "executed", false, false, true, false},
		{"replacement would invalidate live peer", "executed", false, false, true, true},
		{"replacement would activate multiple outgoing relations", "executed", false, false, true, true},
		{"tombstoned relation stays inactive", "executed", false, false, true, false},
		{"missing typed peer", "executed", false, false, true, true},
		{"wrong peer envelope type", "executed", false, false, true, true},
	}
	for _, tc := range cases {
		t.Run("Decision restore supersession/"+tc.name, func(t *testing.T) {
			options := []func(*http.Request){httptestx.WithCookies(login.Session, login.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, login.CSRF.Value)}
			create := func(summary, status string) uuid.UUID {
				t.Helper()
				body := map[string]any{"client_txn_id": uuid.NewString(), "decision.summary": summary, "decision.status": status, "decision.owner_user_id": actor.String(), "decision.decision_type": "containment", "decision.decided_at": "2026-11-14T00:00:00Z", "decision.rationale": "Retained rationale"}
				data := httptestx.RequireSuccessData(t, httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, h.Server.HTTP.URL+"/api/v1/incidents/"+incident.String()+"/views/cartulary.view.decisions.v1/rows", body, options...), http.StatusCreated)
				return uuid.MustParse(data["row"].(map[string]any)["record_id"].(string))
			}
			supersede := func(target, source uuid.UUID, version int) {
				t.Helper()
				httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, h.Server.HTTP.URL+"/api/v1/records/"+target.String()+"/supersede", map[string]any{"base_row_version": version, "client_txn_id": uuid.NewString(), "reason": "Supersession fixture", "replacement_record_id": source.String()}, options...), http.StatusOK)
			}
			target, source := create("Original restore decision", tc.targetStatus), create("Replacement restore decision", "approved")
			supersede(target, source, 1)
			if tc.deleteTarget {
				httptestx.RequireSuccessEnvelope(t, deleteRecord(t, h, login, target, map[string]any{"base_row_version": 2, "client_txn_id": uuid.NewString()}), http.StatusOK)
			}
			if tc.deleteSource {
				httptestx.RequireSuccessEnvelope(t, deleteRecord(t, h, login, source, map[string]any{"base_row_version": 2, "client_txn_id": uuid.NewString()}), http.StatusOK)
			}
			switch tc.name {
			case "replacement would invalidate live peer":
				mustExec(t, h.DB, `UPDATE decisions SET status='approved' WHERE record_id=$1`, target)
			case "replacement would activate multiple outgoing relations":
				otherTarget := create("Another retained target", "executed")
				mustExec(t, h.DB, `INSERT INTO record_links(incident_id,src_record_id,dst_record_id,link_type,provenance,owner_user_id,created_by_user_id,decided_at) VALUES($1,$2,$3,'supersedes','manual',$4,$4,now())`, incident, source, otherTarget, actor)
			case "tombstoned relation stays inactive", "target without retained relation":
				mustExec(t, h.DB, `UPDATE record_links SET deleted_at=now(),deleted_by_user_id=$2 WHERE src_record_id=$1 AND link_type='supersedes'`, source, actor)
			case "missing typed peer":
				mustExec(t, h.DB, `DELETE FROM decisions WHERE record_id=$1`, target)
			case "wrong peer envelope type":
				mustExec(t, h.DB, `UPDATE records SET record_type='party' WHERE record_id=$1`, target)
			}
			selected := source
			if tc.restoreTarget {
				selected = target
			}
			body := map[string]any{"base_row_version": 3, "client_txn_id": uuid.NewString()}
			linksBefore := decisionSupersessionLinks(t, h, incident)
			before := snapshotDecisionReferenceState(t, h, incident, h.Pool.QueryRow)
			if tc.name == "selected source lock contention" || tc.name == "peer source lock contention" {
				lock, err := h.Pool.Begin(t.Context())
				if err != nil {
					t.Fatal(err)
				}
				defer lock.Rollback(context.Background())
				lockedID := selected
				if tc.name == "peer source lock contention" {
					lockedID = source
				}
				if _, err := lock.Exec(t.Context(), `SELECT record_id FROM decisions WHERE record_id=$1 FOR UPDATE`, lockedID); err != nil {
					t.Fatal(err)
				}
				response := httptestx.DoJSON(t, &http.Client{Timeout: 5 * time.Second}, http.MethodPost, h.Server.HTTP.URL+"/api/v1/records/"+selected.String()+"/restore", body, options...)
				failure := httptestx.RequireErrorEnvelope(t, response, http.StatusConflict, "record_locked")["error"].(map[string]any)
				details := failure["details"].(map[string]any)
				if failure["retryable"] != true || len(details) != 1 || details["record_id"] != selected.String() {
					t.Fatalf("Decision source contention=%#v", failure)
				}
				if err := lock.Rollback(t.Context()); err != nil {
					t.Fatal(err)
				}
				if snapshotDecisionReferenceState(t, h, incident, h.Pool.QueryRow) != before {
					t.Fatal("source contention changed retained state")
				}
			}
			response := restoreRecord(t, h, login, selected, body)
			if tc.blocked {
				failure := httptestx.RequireErrorEnvelope(t, response, http.StatusConflict, "record_restore_blocked")["error"].(map[string]any)
				details := failure["details"].(map[string]any)
				if failure["retryable"] != false || len(details) != 2 || details["record_id"] != selected.String() || details["reason_code"] != "decision_supersession_dependency" {
					t.Fatalf("Decision restore blocker=%#v", failure)
				}
				if snapshotDecisionReferenceState(t, h, incident, h.Pool.QueryRow) != before {
					t.Fatal("blocked Decision restore changed retained state")
				}
				if tc.name != "target with deleted replacement" {
					return
				}
				httptestx.RequireSuccessEnvelope(t, restoreRecord(t, h, login, source, map[string]any{"base_row_version": 3, "client_txn_id": uuid.NewString()}), http.StatusOK)
				response = restoreRecord(t, h, login, selected, body)
			}
			result := httptestx.RequireSuccessEnvelope(t, response, http.StatusOK)["data"].(map[string]any)
			if result["deleted"] != false || result["row_version"] != float64(4) || result["deleted_at"] != nil || result["deleted_by_user_id"] != nil {
				t.Fatalf("Decision restore=%#v", result)
			}
			expectedTarget := tc.targetStatus
			if expectedTarget == "proposed" {
				expectedTarget = "superseded"
			}
			var exact bool
			if err := h.DB.QueryRow(`SELECT d.status=$2 AND r.deleted_at IS NULL AND r.row_version=4 AND p.status=$2 AND p.row_version=4 FROM decisions d JOIN records r USING(record_id) JOIN decision_grid_projection p USING(record_id) WHERE d.record_id=$1`, selected, map[bool]string{true: expectedTarget, false: "approved"}[tc.restoreTarget]).Scan(&exact); err != nil || !exact {
				t.Fatalf("restored Decision source/envelope/projection exact=%v error=%v", exact, err)
			}
			var sourceStatus, targetStatus string
			if err := h.DB.QueryRow(`SELECT a.status,b.status FROM decisions a JOIN decisions b ON b.record_id=$2 WHERE a.record_id=$1`, source, target).Scan(&sourceStatus, &targetStatus); err != nil {
				t.Fatal(err)
			}
			var sourceDeleted, targetDeleted bool
			if err := h.DB.QueryRow(`SELECT a.deleted_at IS NOT NULL,b.deleted_at IS NOT NULL FROM records a JOIN records b ON b.record_id=$2 WHERE a.record_id=$1`, source, target).Scan(&sourceDeleted, &targetDeleted); err != nil {
				t.Fatal(err)
			}
			expectedSourceDeleted := tc.restoreTarget && tc.deleteSource && tc.name != "target with deleted replacement"
			expectedTargetDeleted := !tc.restoreTarget && tc.deleteTarget
			if sourceDeleted != expectedSourceDeleted || targetDeleted != expectedTargetDeleted {
				t.Fatalf("restored endpoint visibility=%v/%v want=%v/%v", sourceDeleted, targetDeleted, expectedSourceDeleted, expectedTargetDeleted)
			}
			if sourceStatus != "approved" || targetStatus != expectedTarget || decisionSupersessionLinks(t, h, incident) != linksBefore {
				t.Fatal("restore changed source status or retained relations")
			}
			if countRecordMutations(t, h.DB, selected, "restore") != 1 || countRecordRevisions(t, h.DB, selected) != 4 {
				t.Fatal("Decision restore history did not advance exactly once")
			}
			requireSingleRecordChangeIntent(t, h.DB, result["change_set_id"].(string), selected, "invalidate")
			after := snapshotDecisionReferenceState(t, h, incident, h.Pool.QueryRow)
			replay := httptestx.RequireSuccessEnvelope(t, restoreRecord(t, h, login, selected, body), http.StatusOK)["data"].(map[string]any)
			if replay["change_set_id"] != result["change_set_id"] || replay["row_version"] != float64(4) || snapshotDecisionReferenceState(t, h, incident, h.Pool.QueryRow) != after {
				t.Fatal("Decision restore replay changed result or retained state")
			}
		})
	}
}
