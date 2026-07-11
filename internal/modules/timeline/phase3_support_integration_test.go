package timeline_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	incidentscenariotest "github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestSupportPhase3Integration_AuthorizationMatrix(t *testing.T) {
	runtime := scenariotest.StartRuntime(t)
	harness := runtime.StartServer(t, "phase3-support-auth")

	adminLogin, _ := flowtest.ProvisionBootstrapAdminUUID(t, harness.Server.HTTP.URL)
	incident := incidentscenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-support-phase3-auth-incident",
		"incident_key":  "IR-SUPPORT-AUTH",
		"title":         "Phase 3 support auth matrix",
	})
	incidentID := incident["incident_id"].(string)

	editorUser := flowtest.SeedLocalUserRecord(t, harness.DB, "phase3-editor@example.test", "Phase 3 Editor", "Phase3EditorPass1!", false, false, true)
	reviewerUser := flowtest.SeedLocalUserRecord(t, harness.DB, "phase3-reviewer@example.test", "Phase 3 Reviewer", "Phase3ReviewerPass1!", false, false, true)
	outsiderUser := flowtest.SeedLocalUserRecord(t, harness.DB, "phase3-outsider@example.test", "Phase 3 Outsider", "Phase3OutsiderPass1!", false, false, true)
	_ = outsiderUser

	incidentscenariotest.CreateMembershipForUser(t, harness.Server, adminLogin, incidentID, editorUser.ID.String(), editorUser.Email, "editor")
	incidentscenariotest.CreateMembershipForUser(t, harness.Server, adminLogin, incidentID, reviewerUser.ID.String(), reviewerUser.Email, "reviewer")

	editorSession, editorCSRF := flowtest.LoginLocalUser(t, harness.Server.HTTP.URL, editorUser.Email, "Phase3EditorPass1!", nil)
	reviewerSession, reviewerCSRF := flowtest.LoginLocalUser(t, harness.Server.HTTP.URL, reviewerUser.Email, "Phase3ReviewerPass1!", nil)
	outsiderSession, outsiderCSRF := flowtest.LoginLocalUser(t, harness.Server.HTTP.URL, "phase3-outsider@example.test", "Phase3OutsiderPass1!", nil)

	editorLogin := flowtest.LoginResult{SessionCookie: editorSession, CSRFCookie: editorCSRF}
	reviewerLogin := flowtest.LoginResult{SessionCookie: reviewerSession, CSRFCookie: reviewerCSRF}
	outsiderLogin := flowtest.LoginResult{SessionCookie: outsiderSession, CSRFCookie: outsiderCSRF}

	t.Run("query route matrix", func(t *testing.T) {
		cases := []struct {
			name       string
			login      flowtest.LoginResult
			wantStatus int
			wantCode   string
		}{
			{name: "no-membership", login: outsiderLogin, wantStatus: http.StatusNotFound, wantCode: "incident_not_found"},
			{name: "editor", login: editorLogin, wantStatus: http.StatusOK},
			{name: "reviewer", login: reviewerLogin, wantStatus: http.StatusOK},
			{name: "admin", login: adminLogin, wantStatus: http.StatusOK},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				resp := httptestx.DoJSON(
					t,
					http.MethodPost,
					harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/views/"+timeline.TimelineViewSchemaID+"/query",
					map[string]any{},
					httptestx.WithCookies(tc.login.SessionCookie),
				)
				if tc.wantStatus == http.StatusOK {
					httptestx.RequireSuccessEnvelope(t, resp, tc.wantStatus)
					return
				}
				httptestx.RequireErrorEnvelope(t, resp, tc.wantStatus, tc.wantCode)
			})
		}
	})

	t.Run("create route matrix", func(t *testing.T) {
		cases := []struct {
			name       string
			login      flowtest.LoginResult
			wantStatus int
			wantCode   string
		}{
			{name: "no-membership", login: outsiderLogin, wantStatus: http.StatusNotFound, wantCode: "incident_not_found"},
			{name: "editor", login: editorLogin, wantStatus: http.StatusCreated},
			{name: "reviewer", login: reviewerLogin, wantStatus: http.StatusCreated},
			{name: "admin", login: adminLogin, wantStatus: http.StatusCreated},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				resp := httptestx.DoJSON(
					t,
					http.MethodPost,
					harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/views/"+timeline.TimelineViewSchemaID+"/rows",
					map[string]any{
						"client_txn_id":                   fmt.Sprintf("txn-support-phase3-auth-create-%s", tc.name),
						"timeline.activity_synopsis_text": fmt.Sprintf("created by %s", tc.name),
					},
					httptestx.WithCookies(tc.login.SessionCookie, tc.login.CSRFCookie),
					httptestx.WithHeader(authn.CSRFHeaderName, tc.login.CSRFCookie.Value),
				)
				if tc.wantStatus == http.StatusCreated {
					httptestx.RequireSuccessEnvelope(t, resp, tc.wantStatus)
					return
				}
				httptestx.RequireErrorEnvelope(t, resp, tc.wantStatus, tc.wantCode)
			})
		}
	})

	t.Run("patch route matrix", func(t *testing.T) {
		editorTarget := scenariotest.CreateTimelineRow(t, harness.Server, incidentID, adminLogin, map[string]any{
			"client_txn_id":                   "txn-support-phase3-auth-patch-editor",
			"timeline.activity_synopsis_text": "patch editor target",
		})
		reviewerTarget := scenariotest.CreateTimelineRow(t, harness.Server, incidentID, adminLogin, map[string]any{
			"client_txn_id":                   "txn-support-phase3-auth-patch-reviewer",
			"timeline.activity_synopsis_text": "patch reviewer target",
		})
		adminTarget := scenariotest.CreateTimelineRow(t, harness.Server, incidentID, adminLogin, map[string]any{
			"client_txn_id":                   "txn-support-phase3-auth-patch-admin",
			"timeline.activity_synopsis_text": "patch admin target",
		})
		deniedTarget := scenariotest.CreateTimelineRow(t, harness.Server, incidentID, adminLogin, map[string]any{
			"client_txn_id":                   "txn-support-phase3-auth-patch-denied",
			"timeline.activity_synopsis_text": "patch denied target",
		})

		cases := []struct {
			name       string
			login      flowtest.LoginResult
			recordID   string
			wantStatus int
			wantCode   string
		}{
			{name: "no-membership", login: outsiderLogin, recordID: deniedTarget["row"].(map[string]any)["record_id"].(string), wantStatus: http.StatusNotFound, wantCode: "incident_not_found"},
			{name: "editor", login: editorLogin, recordID: editorTarget["row"].(map[string]any)["record_id"].(string), wantStatus: http.StatusOK},
			{name: "reviewer", login: reviewerLogin, recordID: reviewerTarget["row"].(map[string]any)["record_id"].(string), wantStatus: http.StatusOK},
			{name: "admin", login: adminLogin, recordID: adminTarget["row"].(map[string]any)["record_id"].(string), wantStatus: http.StatusOK},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				resp := httptestx.DoJSON(
					t,
					http.MethodPatch,
					harness.Server.HTTP.URL+"/api/v1/records/"+tc.recordID,
					map[string]any{
						"view_schema_id":   timeline.TimelineViewSchemaID,
						"base_row_version": 1,
						"client_txn_id":    fmt.Sprintf("txn-support-phase3-auth-patch-%s", tc.name),
						"changes": []map[string]any{
							{"field_key": "timeline.activity_synopsis_text", "value": fmt.Sprintf("patched by %s", tc.name)},
						},
					},
					httptestx.WithCookies(tc.login.SessionCookie, tc.login.CSRFCookie),
					httptestx.WithHeader(authn.CSRFHeaderName, tc.login.CSRFCookie.Value),
				)
				if tc.wantStatus == http.StatusOK {
					httptestx.RequireSuccessEnvelope(t, resp, tc.wantStatus)
					return
				}
				httptestx.RequireErrorEnvelope(t, resp, tc.wantStatus, tc.wantCode)
			})
		}
	})

	t.Run("mark reviewed route matrix", func(t *testing.T) {
		reviewerTarget := scenariotest.CreateTimelineRow(t, harness.Server, incidentID, adminLogin, map[string]any{
			"client_txn_id":                   "txn-support-phase3-auth-review-reviewer",
			"timeline.activity_synopsis_text": "review reviewer target",
		})
		adminTarget := scenariotest.CreateTimelineRow(t, harness.Server, incidentID, adminLogin, map[string]any{
			"client_txn_id":                   "txn-support-phase3-auth-review-admin",
			"timeline.activity_synopsis_text": "review admin target",
		})
		deniedTarget := scenariotest.CreateTimelineRow(t, harness.Server, incidentID, adminLogin, map[string]any{
			"client_txn_id":                   "txn-support-phase3-auth-review-denied",
			"timeline.activity_synopsis_text": "review denied target",
		})

		cases := []struct {
			name       string
			login      flowtest.LoginResult
			recordID   string
			wantStatus int
			wantCode   string
		}{
			{name: "no-membership", login: outsiderLogin, recordID: deniedTarget["row"].(map[string]any)["record_id"].(string), wantStatus: http.StatusNotFound, wantCode: "incident_not_found"},
			{name: "editor", login: editorLogin, recordID: deniedTarget["row"].(map[string]any)["record_id"].(string), wantStatus: http.StatusForbidden, wantCode: "authorization_denied"},
			{name: "reviewer", login: reviewerLogin, recordID: reviewerTarget["row"].(map[string]any)["record_id"].(string), wantStatus: http.StatusOK},
			{name: "admin", login: adminLogin, recordID: adminTarget["row"].(map[string]any)["record_id"].(string), wantStatus: http.StatusOK},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				resp := httptestx.DoJSON(
					t,
					http.MethodPost,
					harness.Server.HTTP.URL+"/api/v1/records/"+tc.recordID+"/mark-reviewed",
					map[string]any{
						"base_row_version": 1,
						"client_txn_id":    fmt.Sprintf("txn-support-phase3-auth-review-%s", tc.name),
					},
					httptestx.WithCookies(tc.login.SessionCookie, tc.login.CSRFCookie),
					httptestx.WithHeader(authn.CSRFHeaderName, tc.login.CSRFCookie.Value),
				)
				if tc.wantStatus == http.StatusOK {
					httptestx.RequireSuccessEnvelope(t, resp, tc.wantStatus)
					return
				}
				httptestx.RequireErrorEnvelope(t, resp, tc.wantStatus, tc.wantCode)
			})
		}
	})

	t.Run("supersede route matrix", func(t *testing.T) {
		reviewerTarget := scenariotest.CreateTimelineRow(t, harness.Server, incidentID, adminLogin, map[string]any{
			"client_txn_id":                   "txn-support-phase3-auth-supersede-reviewer",
			"timeline.activity_synopsis_text": "supersede reviewer target",
		})
		adminTarget := scenariotest.CreateTimelineRow(t, harness.Server, incidentID, adminLogin, map[string]any{
			"client_txn_id":                   "txn-support-phase3-auth-supersede-admin",
			"timeline.activity_synopsis_text": "supersede admin target",
		})
		deniedTarget := scenariotest.CreateTimelineRow(t, harness.Server, incidentID, adminLogin, map[string]any{
			"client_txn_id":                   "txn-support-phase3-auth-supersede-denied",
			"timeline.activity_synopsis_text": "supersede denied target",
		})
		reviewerReplacement := scenariotest.CreateTimelineRow(t, harness.Server, incidentID, adminLogin, map[string]any{
			"client_txn_id":                   "txn-support-phase3-auth-supersede-reviewer-replacement",
			"timeline.activity_synopsis_text": "supersede reviewer replacement",
		})
		adminReplacement := scenariotest.CreateTimelineRow(t, harness.Server, incidentID, adminLogin, map[string]any{
			"client_txn_id":                   "txn-support-phase3-auth-supersede-admin-replacement",
			"timeline.activity_synopsis_text": "supersede admin replacement",
		})
		deniedReplacement := scenariotest.CreateTimelineRow(t, harness.Server, incidentID, adminLogin, map[string]any{
			"client_txn_id":                   "txn-support-phase3-auth-supersede-denied-replacement",
			"timeline.activity_synopsis_text": "supersede denied replacement",
		})

		markReviewed := func(recordID string, txn string) {
			resp := httptestx.DoJSON(
				t,
				http.MethodPost,
				harness.Server.HTTP.URL+"/api/v1/records/"+recordID+"/mark-reviewed",
				map[string]any{
					"base_row_version": 1,
					"client_txn_id":    txn,
				},
				httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
				httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
			)
			httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)
		}

		reviewerTargetID := reviewerTarget["row"].(map[string]any)["record_id"].(string)
		adminTargetID := adminTarget["row"].(map[string]any)["record_id"].(string)
		deniedTargetID := deniedTarget["row"].(map[string]any)["record_id"].(string)
		markReviewed(reviewerTargetID, "txn-support-phase3-auth-supersede-reviewer-mark")
		markReviewed(adminTargetID, "txn-support-phase3-auth-supersede-admin-mark")
		markReviewed(deniedTargetID, "txn-support-phase3-auth-supersede-denied-mark")

		cases := []struct {
			name                string
			login               flowtest.LoginResult
			recordID            string
			replacementRecordID string
			wantStatus          int
			wantCode            string
		}{
			{name: "no-membership", login: outsiderLogin, recordID: deniedTargetID, replacementRecordID: deniedReplacement["row"].(map[string]any)["record_id"].(string), wantStatus: http.StatusNotFound, wantCode: "incident_not_found"},
			{name: "editor", login: editorLogin, recordID: deniedTargetID, replacementRecordID: deniedReplacement["row"].(map[string]any)["record_id"].(string), wantStatus: http.StatusForbidden, wantCode: "authorization_denied"},
			{name: "reviewer", login: reviewerLogin, recordID: reviewerTargetID, replacementRecordID: reviewerReplacement["row"].(map[string]any)["record_id"].(string), wantStatus: http.StatusOK},
			{name: "admin", login: adminLogin, recordID: adminTargetID, replacementRecordID: adminReplacement["row"].(map[string]any)["record_id"].(string), wantStatus: http.StatusOK},
		}
		for _, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				resp := httptestx.DoJSON(
					t,
					http.MethodPost,
					harness.Server.HTTP.URL+"/api/v1/records/"+tc.recordID+"/supersede",
					map[string]any{
						"base_row_version":      2,
						"client_txn_id":         fmt.Sprintf("txn-support-phase3-auth-supersede-%s", tc.name),
						"reason":                fmt.Sprintf("superseded by %s", tc.name),
						"replacement_record_id": tc.replacementRecordID,
					},
					httptestx.WithCookies(tc.login.SessionCookie, tc.login.CSRFCookie),
					httptestx.WithHeader(authn.CSRFHeaderName, tc.login.CSRFCookie.Value),
				)
				if tc.wantStatus == http.StatusOK {
					httptestx.RequireSuccessEnvelope(t, resp, tc.wantStatus)
					return
				}
				httptestx.RequireErrorEnvelope(t, resp, tc.wantStatus, tc.wantCode)
			})
		}
	})
}
