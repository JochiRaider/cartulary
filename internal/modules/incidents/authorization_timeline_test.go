package incidents_test

import (
	"net/http"
	"testing"

	"github.com/coder/websocket"
	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	timelineroutetest "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport/routetest"
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport/incidentwstest"
)

func TestIncidentTimelineMutationsAuthorization_Integration(t *testing.T) {
	t.Run("create", func(t *testing.T) {
		f := newIncidentAuthorization(t, "authz-timeline-create")
		for index, tc := range []struct {
			name, role string
			actor      authcookietest.AuthCookies
			status     int
			code       string
		}{
			{"no-membership", "", f.subject, 404, "incident_not_found"},
			{"viewer", "viewer", f.subject, 403, "authorization_denied"},
			{"editor", "editor", f.subject, 201, ""},
			{"reviewer", "reviewer", f.subject, 201, ""},
			{"admin", "admin", f.subject, 201, ""},
			{"downgraded-viewer", "viewer", f.subject, 403, "authorization_denied"},
			{"removed", "", f.subject, 404, "incident_not_found"},
			{"deployment-admin-without-membership", "", f.deploymentAdmin, 404, "incident_not_found"},
		} {
			if !t.Run(tc.name, func(t *testing.T) {
				f.setSubjectRole(t, tc.role)
				path := "/api/v1/incidents/" + f.incidentID + "/views/" + timeline.TimelineViewSchemaID + "/rows"
				body := map[string]any{"client_txn_id": uuid.NewString(), "timeline.activity_synopsis_text": tc.name}
				before := snapshotTimelineAuthorization(t, f)
				if index == 0 {
					requireWriteAuthentication(t, f, http.MethodPost, path, body)
				}
				resp := authorizationWrite(t, f, tc.actor, http.MethodPost, path, body)
				data := requireAuthorizationResponse(t, resp, tc.status, tc.code)
				if tc.code == "" {
					result := data["row"].(map[string]any)
					if result["record_id"] == nil {
						t.Fatalf("missing record: %#v", data)
					}
					if result["row_version"] != float64(1) {
						t.Fatalf("initial row version = %#v", result)
					}
				}
				if tc.code != "" {
					if after := snapshotTimelineAuthorization(t, f); after != before {
						t.Fatal("denied Timeline mutation changed protected source, projection or effects")
					}
				} else {
					recordID := data["row"].(map[string]any)["record_id"].(string)
					wantVersion := int64(1)
					persisted := readAuthorizationTimelineRecord(t, f, recordID)
					want := authorizationTimelineRecord{wantVersion, wantVersion, "rough", tc.name, f.subjectID}
					if persisted != want {
						t.Fatalf("persisted Timeline row = %+v, want %+v", persisted, want)
					}
					requireTimelineAuthorizationHistory(t, f, recordID, body["client_txn_id"].(string), "timeline.rows.create", resp.Header.Get("X-Request-ID"), wantVersion)
					requireAuthorizationReceipt(t, f, body["client_txn_id"].(string), f.incidentID+":"+timeline.TimelineViewSchemaID, 201)
				}

			}) {
				return
			}
		}
	})
	t.Run("patch", func(t *testing.T) {
		f := newIncidentAuthorization(t, "authz-timeline-patch")
		for index, tc := range []struct {
			name, role string
			actor      authcookietest.AuthCookies
			status     int
			code       string
		}{
			{"no-membership", "", f.subject, 404, "incident_not_found"},
			{"viewer", "viewer", f.subject, 403, "authorization_denied"},
			{"editor", "editor", f.subject, 200, ""},
			{"reviewer", "reviewer", f.subject, 200, ""},
			{"admin", "admin", f.subject, 200, ""},
			{"downgraded-viewer", "viewer", f.subject, 403, "authorization_denied"},
			{"removed", "", f.subject, 404, "incident_not_found"},
			{"deployment-admin-without-membership", "", f.deploymentAdmin, 404, "incident_not_found"},
		} {
			if !t.Run(tc.name, func(t *testing.T) {
				f.setSubjectRole(t, tc.role)
				created := timelineroutetest.CreateRow(t, f.harness.Server, f.controller, f.incidentID, map[string]any{"client_txn_id": uuid.NewString(), "timeline.activity_synopsis_text": "Authorization target"})
				row := created["row"].(map[string]any)
				recordID := row["record_id"].(string)
				version := row["row_version"]
				path := "/api/v1/records/" + recordID
				body := map[string]any{"client_txn_id": uuid.NewString(), "base_row_version": version}
				body["view_schema_id"] = timeline.TimelineViewSchemaID
				body["changes"] = []map[string]any{{"field_key": "timeline.activity_synopsis_text", "value": tc.name}}
				before := snapshotTimelineAuthorization(t, f)
				if index == 0 {
					requireWriteAuthentication(t, f, http.MethodPatch, path, body)
				}
				resp := authorizationWrite(t, f, tc.actor, http.MethodPatch, path, body)
				data := requireAuthorizationResponse(t, resp, tc.status, tc.code)
				if tc.code == "" {
					result := data["row"].(map[string]any)
					if result["record_id"] == nil {
						t.Fatalf("missing record: %#v", data)
					}
					if authorizationVersion(t, result["row_version"]) != authorizationVersion(t, version)+1 {
						t.Fatalf("patch version = %#v", result)
					}
				}
				if tc.code != "" {
					if after := snapshotTimelineAuthorization(t, f); after != before {
						t.Fatal("denied Timeline mutation changed protected source, projection or effects")
					}
				} else {
					wantVersion := authorizationVersion(t, version) + 1
					persisted := readAuthorizationTimelineRecord(t, f, recordID)
					want := authorizationTimelineRecord{wantVersion, wantVersion, "enriched", tc.name, f.subjectID}
					if persisted != want {
						t.Fatalf("persisted Timeline row = %+v, want %+v", persisted, want)
					}
					requireTimelineAuthorizationHistory(t, f, recordID, body["client_txn_id"].(string), "timeline.records.patch", resp.Header.Get("X-Request-ID"), wantVersion)
					requireAuthorizationReceipt(t, f, body["client_txn_id"].(string), recordID, 200)
				}

			}) {
				return
			}
		}
	})
	t.Run("review", func(t *testing.T) {
		f := newIncidentAuthorization(t, "authz-timeline-review")
		for index, tc := range []struct {
			name, role string
			actor      authcookietest.AuthCookies
			status     int
			code       string
		}{
			{"no-membership", "", f.subject, 404, "incident_not_found"},
			{"viewer", "viewer", f.subject, 403, "authorization_denied"},
			{"editor", "editor", f.subject, 403, "authorization_denied"},
			{"reviewer", "reviewer", f.subject, 200, ""},
			{"admin", "admin", f.subject, 200, ""},
			{"downgraded-viewer", "viewer", f.subject, 403, "authorization_denied"},
			{"removed", "", f.subject, 404, "incident_not_found"},
			{"deployment-admin-without-membership", "", f.deploymentAdmin, 404, "incident_not_found"},
		} {
			if !t.Run(tc.name, func(t *testing.T) {
				f.setSubjectRole(t, tc.role)
				created := timelineroutetest.CreateRow(t, f.harness.Server, f.controller, f.incidentID, map[string]any{"client_txn_id": uuid.NewString(), "timeline.activity_synopsis_text": "Authorization target"})
				row := created["row"].(map[string]any)
				recordID := row["record_id"].(string)
				version := row["row_version"]
				path := "/api/v1/records/" + recordID
				body := map[string]any{"client_txn_id": uuid.NewString(), "base_row_version": version}
				path += "/mark-reviewed"
				before := snapshotTimelineAuthorization(t, f)
				if index == 0 {
					requireWriteAuthentication(t, f, http.MethodPost, path, body)
				}
				resp := authorizationWrite(t, f, tc.actor, http.MethodPost, path, body)
				data := requireAuthorizationResponse(t, resp, tc.status, tc.code)
				if tc.code == "" && (data["capture_state"] != "reviewed" || authorizationVersion(t, data["row_version"]) != authorizationVersion(t, version)+1) {
					t.Fatalf("transition = %#v", data)
				}
				if tc.code != "" {
					if after := snapshotTimelineAuthorization(t, f); after != before {
						t.Fatal("denied Timeline mutation changed protected source, projection or effects")
					}
				} else {
					wantVersion := authorizationVersion(t, version) + 1
					persisted := readAuthorizationTimelineRecord(t, f, recordID)
					want := authorizationTimelineRecord{wantVersion, wantVersion, "reviewed", "Authorization target", f.subjectID}
					if persisted != want {
						t.Fatalf("persisted Timeline row = %+v, want %+v", persisted, want)
					}
					requireTimelineAuthorizationHistory(t, f, recordID, body["client_txn_id"].(string), "timeline.records.mark_reviewed", resp.Header.Get("X-Request-ID"), wantVersion)
					requireAuthorizationReceipt(t, f, body["client_txn_id"].(string), recordID, 200)
				}

			}) {
				return
			}
		}
	})
	t.Run("supersede-rough", func(t *testing.T) {
		f := newIncidentAuthorization(t, "authz-timeline-supersede-rough")
		for index, tc := range []struct {
			name, role string
			actor      authcookietest.AuthCookies
			status     int
			code       string
		}{
			{"no-membership", "", f.subject, 404, "incident_not_found"},
			{"viewer", "viewer", f.subject, 403, "authorization_denied"},
			{"editor", "editor", f.subject, 403, "authorization_denied"},
			{"reviewer", "reviewer", f.subject, 200, ""},
			{"admin", "admin", f.subject, 200, ""},
			{"downgraded-viewer", "viewer", f.subject, 403, "authorization_denied"},
			{"removed", "", f.subject, 404, "incident_not_found"},
			{"deployment-admin-without-membership", "", f.deploymentAdmin, 404, "incident_not_found"},
		} {
			if !t.Run(tc.name, func(t *testing.T) {
				f.setSubjectRole(t, tc.role)
				created := timelineroutetest.CreateRow(t, f.harness.Server, f.controller, f.incidentID, map[string]any{"client_txn_id": uuid.NewString(), "timeline.activity_synopsis_text": "Authorization target"})
				row := created["row"].(map[string]any)
				recordID := row["record_id"].(string)
				version := row["row_version"]
				path := "/api/v1/records/" + recordID
				body := map[string]any{"client_txn_id": uuid.NewString(), "base_row_version": version}
				replacement := timelineroutetest.CreateRow(t, f.harness.Server, f.controller, f.incidentID, map[string]any{"client_txn_id": uuid.NewString(), "timeline.activity_synopsis_text": "Replacement target"})
				path += "/supersede"
				body["replacement_record_id"] = replacement["row"].(map[string]any)["record_id"]
				body["reason"] = "Authorization supersede " + tc.name
				before := snapshotTimelineAuthorization(t, f)
				if index == 0 {
					requireWriteAuthentication(t, f, http.MethodPost, path, body)
				}
				resp := authorizationWrite(t, f, tc.actor, http.MethodPost, path, body)
				data := requireAuthorizationResponse(t, resp, tc.status, tc.code)
				if tc.code == "" && (data["capture_state"] != "superseded" || authorizationVersion(t, data["row_version"]) != authorizationVersion(t, version)+1) {
					t.Fatalf("transition = %#v", data)
				}
				if tc.code != "" {
					if after := snapshotTimelineAuthorization(t, f); after != before {
						t.Fatal("denied Timeline mutation changed protected source, projection or effects")
					}
				} else {
					wantVersion := authorizationVersion(t, version) + 1
					persisted := readAuthorizationTimelineRecord(t, f, recordID)
					want := authorizationTimelineRecord{wantVersion, wantVersion, "superseded", "Authorization target", f.subjectID}
					if persisted != want {
						t.Fatalf("persisted Timeline row = %+v, want %+v", persisted, want)
					}
					changeSetID := requireTimelineAuthorizationHistory(t, f, recordID, body["client_txn_id"].(string), "timeline.records.supersede", resp.Header.Get("X-Request-ID"), wantVersion)
					requireAuthorizationReceipt(t, f, body["client_txn_id"].(string), recordID, 200)
					var links int
					if err := f.harness.DB.QueryRow(`SELECT count(*) FROM record_links WHERE incident_id::text=$1 AND src_record_id::text=$2 AND dst_record_id::text=$3 AND link_type='supersedes'`, f.incidentID, body["replacement_record_id"], recordID).Scan(&links); err != nil {
						t.Fatal(err)
					}
					if links != 1 {
						t.Fatalf("persisted replacement links = %d, want 1", links)
					}
					var reason string
					if err := f.harness.DB.QueryRow(`SELECT reason FROM change_sets WHERE change_set_id::text=$1`, changeSetID).Scan(&reason); err != nil {
						t.Fatal(err)
					}
					if reason != body["reason"] {
						t.Fatalf("supersede history reason = %q", reason)
					}
				}

			}) {
				return
			}
		}
	})
	t.Run("supersede-reviewed", func(t *testing.T) {
		f := newIncidentAuthorization(t, "authz-timeline-supersede-reviewed")
		for index, tc := range []struct {
			name, role string
			actor      authcookietest.AuthCookies
			status     int
			code       string
		}{
			{"no-membership", "", f.subject, 404, "incident_not_found"},
			{"viewer", "viewer", f.subject, 403, "authorization_denied"},
			{"editor", "editor", f.subject, 403, "authorization_denied"},
			{"reviewer", "reviewer", f.subject, 200, ""},
			{"admin", "admin", f.subject, 200, ""},
			{"downgraded-viewer", "viewer", f.subject, 403, "authorization_denied"},
			{"removed", "", f.subject, 404, "incident_not_found"},
			{"deployment-admin-without-membership", "", f.deploymentAdmin, 404, "incident_not_found"},
		} {
			if !t.Run(tc.name, func(t *testing.T) {
				f.setSubjectRole(t, tc.role)
				created := timelineroutetest.CreateRow(t, f.harness.Server, f.controller, f.incidentID, map[string]any{"client_txn_id": uuid.NewString(), "timeline.activity_synopsis_text": "Authorization target"})
				row := created["row"].(map[string]any)
				recordID := row["record_id"].(string)
				version := row["row_version"]
				reviewed := requireAuthorizationResponse(t, authorizationWrite(t, f, f.controller, http.MethodPost, "/api/v1/records/"+recordID+"/mark-reviewed", map[string]any{"client_txn_id": uuid.NewString(), "base_row_version": version}), 200, "")
				if reviewed["capture_state"] != "reviewed" {
					t.Fatalf("review prerequisite = %#v", reviewed)
				}
				version = reviewed["row_version"]
				path := "/api/v1/records/" + recordID
				body := map[string]any{"client_txn_id": uuid.NewString(), "base_row_version": version}
				replacement := timelineroutetest.CreateRow(t, f.harness.Server, f.controller, f.incidentID, map[string]any{"client_txn_id": uuid.NewString(), "timeline.activity_synopsis_text": "Replacement target"})
				path += "/supersede"
				body["replacement_record_id"] = replacement["row"].(map[string]any)["record_id"]
				body["reason"] = "Authorization supersede " + tc.name
				before := snapshotTimelineAuthorization(t, f)
				if index == 0 {
					requireWriteAuthentication(t, f, http.MethodPost, path, body)
				}
				resp := authorizationWrite(t, f, tc.actor, http.MethodPost, path, body)
				data := requireAuthorizationResponse(t, resp, tc.status, tc.code)
				if tc.code == "" && (data["capture_state"] != "superseded" || authorizationVersion(t, data["row_version"]) != authorizationVersion(t, version)+1) {
					t.Fatalf("transition = %#v", data)
				}
				if tc.code != "" {
					if after := snapshotTimelineAuthorization(t, f); after != before {
						t.Fatal("denied Timeline mutation changed protected source, projection or effects")
					}
				} else {
					wantVersion := authorizationVersion(t, version) + 1
					persisted := readAuthorizationTimelineRecord(t, f, recordID)
					want := authorizationTimelineRecord{wantVersion, wantVersion, "superseded", "Authorization target", f.subjectID}
					if persisted != want {
						t.Fatalf("persisted Timeline row = %+v, want %+v", persisted, want)
					}
					changeSetID := requireTimelineAuthorizationHistory(t, f, recordID, body["client_txn_id"].(string), "timeline.records.supersede", resp.Header.Get("X-Request-ID"), wantVersion)
					requireAuthorizationReceipt(t, f, body["client_txn_id"].(string), recordID, 200)
					var links int
					if err := f.harness.DB.QueryRow(`SELECT count(*) FROM record_links WHERE incident_id::text=$1 AND src_record_id::text=$2 AND dst_record_id::text=$3 AND link_type='supersedes'`, f.incidentID, body["replacement_record_id"], recordID).Scan(&links); err != nil {
						t.Fatal(err)
					}
					if links != 1 {
						t.Fatalf("persisted replacement links = %d, want 1", links)
					}
					var reason string
					if err := f.harness.DB.QueryRow(`SELECT reason FROM change_sets WHERE change_set_id::text=$1`, changeSetID).Scan(&reason); err != nil {
						t.Fatal(err)
					}
					if reason != body["reason"] {
						t.Fatalf("supersede history reason = %q", reason)
					}
				}

			}) {
				return
			}
		}
	})
}

func TestIncidentSocketAdmissionAuthorization_Integration(t *testing.T) {
	f := newIncidentAuthorization(t, "authz-socket")
	for index, tc := range []struct {
		name, role string
		actor      authcookietest.AuthCookies
		status     int
		code       string
	}{
		{"no-membership", "", f.subject, 404, "incident_not_found"},
		{"viewer", "viewer", f.subject, 200, ""},
		{"editor", "editor", f.subject, 200, ""},
		{"reviewer", "reviewer", f.subject, 200, ""},
		{"admin", "admin", f.subject, 200, ""},
		{"downgraded-viewer", "viewer", f.subject, 200, ""},
		{"removed", "", f.subject, 404, "incident_not_found"},
		{"deployment-admin-without-membership", "", f.deploymentAdmin, 404, "incident_not_found"},
	} {
		if !t.Run(tc.name, func(t *testing.T) {
			f.setSubjectRole(t, tc.role)

			if index == 0 {
				incidentwstest.RequireDialErrorEnvelope(t, f.harness.Server.HTTP.URL, f.incidentID, incidentwstest.ConnectOptions{}, 401, "session_required")
			}
			if tc.code != "" {
				incidentwstest.RequireDialErrorEnvelope(t, f.harness.Server.HTTP.URL, f.incidentID, incidentwstest.ConnectOptions{SessionToken: tc.actor.Session.Value}, tc.status, tc.code)
				return
			}
			client := incidentwstest.ConnectAndHello(t, f.harness.Server.HTTP.URL, f.incidentID, incidentwstest.ConnectOptions{SessionToken: tc.actor.Session.Value})
			defer client.Close(websocket.StatusNormalClosure, "authorization_complete")

		}) {
			return
		}
	}
}
