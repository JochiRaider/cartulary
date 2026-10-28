package incidents_test

import (
	"net/http"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestIncidentCollectionAuthorization_Integration(t *testing.T) {
	t.Run("list", func(t *testing.T) {
		f := newAuthorizationAccounts(t, "authz-collection-list")
		scenariotest.CreateIncident(t, f.harness.Server, f.controller, map[string]any{
			"client_txn_id": uuid.NewString(), "incident_key": "IR-" + uuid.NewString(), "title": "Hidden incident",
		})
		requireReadAuthentication(t, f, http.MethodGet, "/api/v1/incidents", nil)
		resp := httptestx.DoJSON(t, http.MethodGet, f.harness.Server.HTTP.URL+"/api/v1/incidents", nil, httptestx.WithCookies(f.subject.SessionCookie))
		data := requireAuthorizationResponse(t, resp, 200, "")
		if items, ok := data["incidents"].([]any); !ok || len(items) != 0 {
			t.Fatalf("non-member incident list = %#v", data)
		}
	})
	t.Run("create", func(t *testing.T) {
		f := newAuthorizationAccounts(t, "authz-collection-create")
		body := map[string]any{"client_txn_id": uuid.NewString(), "incident_key": "IR-" + uuid.NewString(), "title": "Ordinary user incident"}
		before := snapshotIncidentCreationAuthorization(t, f, body["client_txn_id"].(string))
		requireWriteAuthentication(t, f, http.MethodPost, "/api/v1/incidents", body)
		if after := snapshotIncidentCreationAuthorization(t, f, body["client_txn_id"].(string)); after != before {
			t.Fatal("authentication/CSRF rejection changed incident bootstrap state")
		}
		resp := authorizationWrite(t, f, f.subject, http.MethodPost, "/api/v1/incidents", body)
		data := requireAuthorizationResponse(t, resp, 201, "")
		if got := resp.Header.Get("Location"); got != "/api/v1/incidents/"+data["incident_id"].(string) {
			t.Fatalf("Location = %q", got)
		}
		f.incidentID = data["incident_id"].(string)
		var incidentVersion int64
		var creator, role string
		if err := f.harness.DB.QueryRow(`SELECT i.incident_version,i.created_by_user_id::text,m.role FROM incidents i JOIN incident_memberships m ON m.incident_id=i.id WHERE i.id::text=$1 AND m.user_id::text=$2`, f.incidentID, f.subjectID).Scan(&incidentVersion, &creator, &role); err != nil {
			t.Fatal(err)
		}
		if incidentVersion != 1 || creator != f.subjectID || role != "admin" {
			t.Fatalf("bootstrap version/creator/role = %d/%s/%s", incidentVersion, creator, role)
		}
		var preferences int
		if err := f.harness.DB.QueryRow(`SELECT (SELECT count(*) FROM incident_workbook_preferences WHERE incident_id::text=$1)+(SELECT count(*) FROM user_workbook_preferences WHERE incident_id::text=$1 AND user_id::text=$2 AND home_sheet_ref IS NULL)`, f.incidentID, f.subjectID).Scan(&preferences); err != nil {
			t.Fatal(err)
		}
		if preferences != 2 {
			t.Fatalf("bootstrap preference resources = %d, want 2", preferences)
		}
		requireIncidentAuthorizationAudit(t, f, resp.Header.Get("X-Request-ID"), "incident_created", "")
		requireIncidentAuthorizationAudit(t, f, resp.Header.Get("X-Request-ID"), "incident_membership_created", f.subjectID)
		requireAuthorizationReceipt(t, f, body["client_txn_id"].(string), "actor", 201)
		if data["incident_version"] != float64(1) {
			t.Fatalf("initial version = %v", data["incident_version"])
		}
	})
}

func TestIncidentResourceAuthorization_Integration(t *testing.T) {
	t.Run("read", func(t *testing.T) {
		f := newIncidentAuthorization(t, "authz-resource-read")
		path := "/api/v1/incidents/" + f.incidentID
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
				resp := httptestx.DoJSON(t, http.MethodGet, f.harness.Server.HTTP.URL+path, nil, httptestx.WithCookies(tc.actor.SessionCookie))
				data := requireAuthorizationResponse(t, resp, tc.status, tc.code)
				if tc.code == "" && data["incident_id"] != f.incidentID {
					t.Fatalf("incident = %#v", data)
				}

			}) {
				return
			}
		}
	})
	t.Run("patch", func(t *testing.T) {
		f := newIncidentAuthorization(t, "authz-resource-patch")
		path := "/api/v1/incidents/" + f.incidentID
		current := f.incident
		for index, tc := range []struct {
			name, role string
			actor      flowtest.LoginResult
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

				body := map[string]any{"base_incident_version": current["incident_version"], "current_phase": tc.name}
				before := snapshotIncidentResourceAuthorization(t, f)
				if index == 0 {
					requireWriteAuthentication(t, f, http.MethodPatch, path, body)
				}
				resp := authorizationWrite(t, f, tc.actor, http.MethodPatch, path, body)
				data := requireAuthorizationResponse(t, resp, tc.status, tc.code)
				if tc.code == "" {
					if data["current_phase"] != tc.name || authorizationVersion(t, data["incident_version"]) != authorizationVersion(t, current["incident_version"])+1 {
						t.Fatalf("material patch = %#v", data)
					}
					var phase, actor string
					var version int64
					if err := f.harness.DB.QueryRow(`SELECT current_phase,incident_version,updated_by_user_id::text FROM incidents WHERE id::text=$1`, f.incidentID).Scan(&phase, &version, &actor); err != nil {
						t.Fatal(err)
					}
					if phase != tc.name || version != authorizationVersion(t, data["incident_version"]) || actor != f.subjectID {
						t.Fatalf("persisted metadata = %s/%d/%s", phase, version, actor)
					}
					requireIncidentAuthorizationAudit(t, f, resp.Header.Get("X-Request-ID"), "incident_updated", "")
					current = data
				} else if after := snapshotIncidentResourceAuthorization(t, f); after != before {
					t.Fatal("denied incident patch changed protected state")
				}

			}) {
				return
			}
		}
	})
}
