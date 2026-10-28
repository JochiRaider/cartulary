package incidents_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestIncidentWorkbookPreferencesAuthorization_Integration(t *testing.T) {
	t.Run("me-read", func(t *testing.T) {
		f := newIncidentAuthorization(t, "authz-preferences-me-read")
		path := "/api/v1/incidents/" + f.incidentID + "/workbook-preferences/me"
		for index, tc := range []struct {
			name, role string
			actor      flowtest.LoginResult
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
				if index == 1 {
					requireAuthorizationResponse(t, authorizationWrite(t, f, f.subject, http.MethodPut, path, map[string]any{"home_sheet_ref": nil}), 200, "")
				}
				if index == 0 {
					requireReadAuthentication(t, f, http.MethodGet, path, nil)
				}
				resp := httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, f.harness.Server.HTTP.URL+path, nil, httptestx.WithCookies(tc.actor.SessionCookie))
				requireAuthorizationResponse(t, resp, tc.status, tc.code)

			}) {
				return
			}
		}
	})
	t.Run("me-write", func(t *testing.T) {
		f := newIncidentAuthorization(t, "authz-preferences-me-write")
		path := "/api/v1/incidents/" + f.incidentID + "/workbook-preferences/me"
		for index, tc := range []struct {
			name, role string
			actor      flowtest.LoginResult
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
				var ref any
				if index%2 == 1 {
					ref = map[string]any{"kind": "view_schema", "id": timeline.TimelineViewSchemaID}
				}
				body := map[string]any{"home_sheet_ref": ref}
				before := snapshotPreferenceAuthorization(t, f)
				if index == 0 {
					requireWriteAuthentication(t, f, http.MethodPut, path, body)
				}
				resp := authorizationWrite(t, f, tc.actor, http.MethodPut, path, body)
				requireAuthorizationResponse(t, resp, tc.status, tc.code)
				if tc.code != "" {
					if after := snapshotPreferenceAuthorization(t, f); after != before {
						t.Fatal("denied preference update changed protected state")
					}
				} else {
					encoded, err := json.Marshal(ref)
					if err != nil {
						t.Fatal(err)
					}
					var matches int
					if err := f.harness.DB.QueryRow(`SELECT count(*) FROM user_workbook_preferences WHERE incident_id::text=$1 AND user_id::text=$2 AND coalesce(home_sheet_ref,'null'::jsonb)=$3::jsonb`, f.incidentID, f.subjectID, string(encoded)).Scan(&matches); err != nil {
						t.Fatal(err)
					}
					if matches != 1 {
						t.Fatalf("persisted preference matches = %d, want 1", matches)
					}
					after := snapshotPreferenceAuthorization(t, f)
					if after.Source == before.Source {
						t.Fatal("authorized material preference update did not change stored preference")
					}
					if after.Effects != before.Effects {
						t.Fatal("personal/default preference update changed source history or publication state")
					}
				}

			}) {
				return
			}
		}
	})
	t.Run("default-read", func(t *testing.T) {
		f := newIncidentAuthorization(t, "authz-preferences-default-read")
		path := "/api/v1/incidents/" + f.incidentID + "/workbook-preferences/default"
		for index, tc := range []struct {
			name, role string
			actor      flowtest.LoginResult
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
					requireReadAuthentication(t, f, http.MethodGet, path, nil)
				}
				resp := httptestx.DoJSON(t, http.DefaultClient, http.MethodGet, f.harness.Server.HTTP.URL+path, nil, httptestx.WithCookies(tc.actor.SessionCookie))
				requireAuthorizationResponse(t, resp, tc.status, tc.code)

			}) {
				return
			}
		}
	})
	t.Run("default-write", func(t *testing.T) {
		f := newIncidentAuthorization(t, "authz-preferences-default-write")
		path := "/api/v1/incidents/" + f.incidentID + "/workbook-preferences/default"
		requireAuthorizationResponse(t, authorizationWrite(t, f, f.controller, http.MethodPut, path, map[string]any{"default_sheet_ref": map[string]any{"kind": "view_schema", "id": timeline.TimelineViewSchemaID}}), 200, "")
		for index, tc := range []struct {
			name, role string
			actor      flowtest.LoginResult
			status     int
			code       string
		}{
			{"no-membership", "", f.subject, 404, "incident_not_found"},
			{"viewer", "viewer", f.subject, 403, "authorization_denied"},
			{"editor", "editor", f.subject, 403, "authorization_denied"},
			{"reviewer", "reviewer", f.subject, 403, "authorization_denied"},
			{"admin", "admin", f.subject, 200, ""},
			{"downgraded-viewer", "viewer", f.subject, 403, "authorization_denied"},
			{"removed", "", f.subject, 404, "incident_not_found"},
			{"deployment-admin-without-membership", "", f.deploymentAdmin, 404, "incident_not_found"},
		} {
			if !t.Run(tc.name, func(t *testing.T) {
				f.setSubjectRole(t, tc.role)
				var ref any
				body := map[string]any{"default_sheet_ref": ref}
				before := snapshotPreferenceAuthorization(t, f)
				if index == 0 {
					requireWriteAuthentication(t, f, http.MethodPut, path, body)
				}
				resp := authorizationWrite(t, f, tc.actor, http.MethodPut, path, body)
				requireAuthorizationResponse(t, resp, tc.status, tc.code)
				if tc.code != "" {
					if after := snapshotPreferenceAuthorization(t, f); after != before {
						t.Fatal("denied preference update changed protected state")
					}
				} else {
					encoded, err := json.Marshal(ref)
					if err != nil {
						t.Fatal(err)
					}
					var matches int
					if err := f.harness.DB.QueryRow(`SELECT count(*) FROM incident_workbook_preferences WHERE incident_id::text=$1 AND updated_by_user_id::text=$2 AND coalesce(default_sheet_ref,'null'::jsonb)=$3::jsonb`, f.incidentID, f.subjectID, string(encoded)).Scan(&matches); err != nil {
						t.Fatal(err)
					}
					if matches != 1 {
						t.Fatalf("persisted preference matches = %d, want 1", matches)
					}
					after := snapshotPreferenceAuthorization(t, f)
					if after.Source == before.Source {
						t.Fatal("authorized material preference update did not change stored preference")
					}
					if after.Effects != before.Effects {
						t.Fatal("personal/default preference update changed source history or publication state")
					}
				}

			}) {
				return
			}
		}
	})
}

func TestIncidentWorkbookQueriesAuthorization_Integration(t *testing.T) {
	t.Run("timeline", func(t *testing.T) {
		f := newIncidentAuthorization(t, "authz-query-timeline")
		path := "/api/v1/incidents/" + f.incidentID + "/views/cartulary.view.timeline.v2/query"
		for index, tc := range []struct {
			name, role string
			actor      flowtest.LoginResult
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
					requireReadAuthentication(t, f, http.MethodPost, path, map[string]any{})
				}
				// Query POST is a read: intentionally send no CSRF cookie or header.
				resp := httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, f.harness.Server.HTTP.URL+path, map[string]any{}, httptestx.WithCookies(tc.actor.SessionCookie))
				requireAuthorizationResponse(t, resp, tc.status, tc.code)

			}) {
				return
			}
		}
	})
	t.Run("hosts", func(t *testing.T) {
		f := newIncidentAuthorization(t, "authz-query-hosts")
		path := "/api/v1/incidents/" + f.incidentID + "/views/cartulary.view.hosts.v1/query"
		for index, tc := range []struct {
			name, role string
			actor      flowtest.LoginResult
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
					requireReadAuthentication(t, f, http.MethodPost, path, map[string]any{})
				}
				// Query POST is a read: intentionally send no CSRF cookie or header.
				resp := httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, f.harness.Server.HTTP.URL+path, map[string]any{}, httptestx.WithCookies(tc.actor.SessionCookie))
				requireAuthorizationResponse(t, resp, tc.status, tc.code)

			}) {
				return
			}
		}
	})
	t.Run("identities", func(t *testing.T) {
		f := newIncidentAuthorization(t, "authz-query-identities")
		path := "/api/v1/incidents/" + f.incidentID + "/views/cartulary.view.identities.v1/query"
		for index, tc := range []struct {
			name, role string
			actor      flowtest.LoginResult
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
					requireReadAuthentication(t, f, http.MethodPost, path, map[string]any{})
				}
				// Query POST is a read: intentionally send no CSRF cookie or header.
				resp := httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, f.harness.Server.HTTP.URL+path, map[string]any{}, httptestx.WithCookies(tc.actor.SessionCookie))
				requireAuthorizationResponse(t, resp, tc.status, tc.code)

			}) {
				return
			}
		}
	})
	t.Run("indicators", func(t *testing.T) {
		f := newIncidentAuthorization(t, "authz-query-indicators")
		path := "/api/v1/incidents/" + f.incidentID + "/views/cartulary.view.indicators.v1/query"
		for index, tc := range []struct {
			name, role string
			actor      flowtest.LoginResult
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
					requireReadAuthentication(t, f, http.MethodPost, path, map[string]any{})
				}
				// Query POST is a read: intentionally send no CSRF cookie or header.
				resp := httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, f.harness.Server.HTTP.URL+path, map[string]any{}, httptestx.WithCookies(tc.actor.SessionCookie))
				requireAuthorizationResponse(t, resp, tc.status, tc.code)

			}) {
				return
			}
		}
	})
}
