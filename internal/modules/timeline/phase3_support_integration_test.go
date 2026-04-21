package timeline_test

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/phase3test"
)

func TestSupportPhase3_AuthorizationMatrix(t *testing.T) {
	runtime := phase3test.StartRuntime(t)
	harness := runtime.StartServer(t, "phase3-support-auth")

	adminLogin, _ := phase3test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase3test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-support-phase3-auth-incident",
		"incident_key":  "IR-SUPPORT-AUTH",
		"title":         "Phase 3 support auth matrix",
	})
	incidentID := incident["incident_id"].(string)

	editorUser := phase3test.SeedLocalUserFlags(t, harness.DB, "phase3-editor@example.test", "Phase 3 Editor", "Phase3EditorPass1!", false, false, true)
	reviewerUser := phase3test.SeedLocalUserFlags(t, harness.DB, "phase3-reviewer@example.test", "Phase 3 Reviewer", "Phase3ReviewerPass1!", false, false, true)
	outsiderUser := phase3test.SeedLocalUserFlags(t, harness.DB, "phase3-outsider@example.test", "Phase 3 Outsider", "Phase3OutsiderPass1!", false, false, true)
	_ = outsiderUser

	phase3test.CreateMembership(t, harness.Server, incidentID, editorUser.ID.String(), editorUser.Email, "editor", adminLogin)
	phase3test.CreateMembership(t, harness.Server, incidentID, reviewerUser.ID.String(), reviewerUser.Email, "reviewer", adminLogin)

	editorSession, editorCSRF := phase3test.LoginLocalUser(t, harness.Server, editorUser.Email, "Phase3EditorPass1!")
	reviewerSession, reviewerCSRF := phase3test.LoginLocalUser(t, harness.Server, reviewerUser.Email, "Phase3ReviewerPass1!")
	outsiderSession, outsiderCSRF := phase3test.LoginLocalUser(t, harness.Server, "phase3-outsider@example.test", "Phase3OutsiderPass1!")

	editorLogin := phase3test.LoginResult{SessionCookie: editorSession, CSRFCookie: editorCSRF}
	reviewerLogin := phase3test.LoginResult{SessionCookie: reviewerSession, CSRFCookie: reviewerCSRF}
	outsiderLogin := phase3test.LoginResult{SessionCookie: outsiderSession, CSRFCookie: outsiderCSRF}

	t.Run("query route matrix", func(t *testing.T) {
		cases := []struct {
			name       string
			login      phase3test.LoginResult
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
				resp := phase3test.DoJSON(
					t,
					http.MethodPost,
					harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/views/"+timeline.TimelineViewSchemaID+"/query",
					map[string]any{},
					phase3test.WithCookies(tc.login.SessionCookie),
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
			login      phase3test.LoginResult
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
				resp := phase3test.DoJSON(
					t,
					http.MethodPost,
					harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/views/"+timeline.TimelineViewSchemaID+"/rows",
					map[string]any{
						"client_txn_id":    fmt.Sprintf("txn-support-phase3-auth-create-%s", tc.name),
						"timeline.summary": fmt.Sprintf("created by %s", tc.name),
					},
					phase3test.WithCookies(tc.login.SessionCookie, tc.login.CSRFCookie),
					phase3test.WithHeader(authn.CSRFHeaderName, tc.login.CSRFCookie.Value),
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
		editorTarget := phase3test.CreateTimelineRow(t, harness.Server, incidentID, adminLogin, map[string]any{
			"client_txn_id":    "txn-support-phase3-auth-patch-editor",
			"timeline.summary": "patch editor target",
		})
		reviewerTarget := phase3test.CreateTimelineRow(t, harness.Server, incidentID, adminLogin, map[string]any{
			"client_txn_id":    "txn-support-phase3-auth-patch-reviewer",
			"timeline.summary": "patch reviewer target",
		})
		adminTarget := phase3test.CreateTimelineRow(t, harness.Server, incidentID, adminLogin, map[string]any{
			"client_txn_id":    "txn-support-phase3-auth-patch-admin",
			"timeline.summary": "patch admin target",
		})
		deniedTarget := phase3test.CreateTimelineRow(t, harness.Server, incidentID, adminLogin, map[string]any{
			"client_txn_id":    "txn-support-phase3-auth-patch-denied",
			"timeline.summary": "patch denied target",
		})

		cases := []struct {
			name       string
			login      phase3test.LoginResult
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
				resp := phase3test.DoJSON(
					t,
					http.MethodPatch,
					harness.Server.HTTP.URL+"/api/v1/records/"+tc.recordID,
					map[string]any{
						"view_schema_id":   timeline.TimelineViewSchemaID,
						"base_row_version": 1,
						"client_txn_id":    fmt.Sprintf("txn-support-phase3-auth-patch-%s", tc.name),
						"changes": []map[string]any{
							{"field_key": "timeline.summary", "value": fmt.Sprintf("patched by %s", tc.name)},
						},
					},
					phase3test.WithCookies(tc.login.SessionCookie, tc.login.CSRFCookie),
					phase3test.WithHeader(authn.CSRFHeaderName, tc.login.CSRFCookie.Value),
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
		reviewerTarget := phase3test.CreateTimelineRow(t, harness.Server, incidentID, adminLogin, map[string]any{
			"client_txn_id":    "txn-support-phase3-auth-review-reviewer",
			"timeline.summary": "review reviewer target",
		})
		adminTarget := phase3test.CreateTimelineRow(t, harness.Server, incidentID, adminLogin, map[string]any{
			"client_txn_id":    "txn-support-phase3-auth-review-admin",
			"timeline.summary": "review admin target",
		})
		deniedTarget := phase3test.CreateTimelineRow(t, harness.Server, incidentID, adminLogin, map[string]any{
			"client_txn_id":    "txn-support-phase3-auth-review-denied",
			"timeline.summary": "review denied target",
		})

		cases := []struct {
			name       string
			login      phase3test.LoginResult
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
				resp := phase3test.DoJSON(
					t,
					http.MethodPost,
					harness.Server.HTTP.URL+"/api/v1/records/"+tc.recordID+"/mark-reviewed",
					map[string]any{
						"base_row_version": 1,
						"client_txn_id":    fmt.Sprintf("txn-support-phase3-auth-review-%s", tc.name),
					},
					phase3test.WithCookies(tc.login.SessionCookie, tc.login.CSRFCookie),
					phase3test.WithHeader(authn.CSRFHeaderName, tc.login.CSRFCookie.Value),
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
		reviewerTarget := phase3test.CreateTimelineRow(t, harness.Server, incidentID, adminLogin, map[string]any{
			"client_txn_id":    "txn-support-phase3-auth-supersede-reviewer",
			"timeline.summary": "supersede reviewer target",
		})
		adminTarget := phase3test.CreateTimelineRow(t, harness.Server, incidentID, adminLogin, map[string]any{
			"client_txn_id":    "txn-support-phase3-auth-supersede-admin",
			"timeline.summary": "supersede admin target",
		})
		deniedTarget := phase3test.CreateTimelineRow(t, harness.Server, incidentID, adminLogin, map[string]any{
			"client_txn_id":    "txn-support-phase3-auth-supersede-denied",
			"timeline.summary": "supersede denied target",
		})
		reviewerReplacement := phase3test.CreateTimelineRow(t, harness.Server, incidentID, adminLogin, map[string]any{
			"client_txn_id":    "txn-support-phase3-auth-supersede-reviewer-replacement",
			"timeline.summary": "supersede reviewer replacement",
		})
		adminReplacement := phase3test.CreateTimelineRow(t, harness.Server, incidentID, adminLogin, map[string]any{
			"client_txn_id":    "txn-support-phase3-auth-supersede-admin-replacement",
			"timeline.summary": "supersede admin replacement",
		})
		deniedReplacement := phase3test.CreateTimelineRow(t, harness.Server, incidentID, adminLogin, map[string]any{
			"client_txn_id":    "txn-support-phase3-auth-supersede-denied-replacement",
			"timeline.summary": "supersede denied replacement",
		})

		markReviewed := func(recordID string, txn string) {
			resp := phase3test.DoJSON(
				t,
				http.MethodPost,
				harness.Server.HTTP.URL+"/api/v1/records/"+recordID+"/mark-reviewed",
				map[string]any{
					"base_row_version": 1,
					"client_txn_id":    txn,
				},
				phase3test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
				phase3test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
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
			login               phase3test.LoginResult
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
				resp := phase3test.DoJSON(
					t,
					http.MethodPost,
					harness.Server.HTTP.URL+"/api/v1/records/"+tc.recordID+"/supersede",
					map[string]any{
						"base_row_version":      2,
						"client_txn_id":         fmt.Sprintf("txn-support-phase3-auth-supersede-%s", tc.name),
						"reason":                fmt.Sprintf("superseded by %s", tc.name),
						"replacement_record_id": tc.replacementRecordID,
					},
					phase3test.WithCookies(tc.login.SessionCookie, tc.login.CSRFCookie),
					phase3test.WithHeader(authn.CSRFHeaderName, tc.login.CSRFCookie.Value),
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
