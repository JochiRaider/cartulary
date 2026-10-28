package incidents_test

import (
	"net/http"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestIncidentMembershipAuthorization_Integration(t *testing.T) {
	t.Run("list", func(t *testing.T) {
		f := newIncidentAuthorization(t, "authz-membership-list")
		path := "/api/v1/incidents/" + f.incidentID + "/memberships"
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
	t.Run("audit", func(t *testing.T) {
		f := newIncidentAuthorization(t, "authz-membership-audit")
		path := "/api/v1/incidents/" + f.incidentID + "/membership-audit-events"
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
	t.Run("create", func(t *testing.T) {
		f := newIncidentAuthorization(t, "authz-membership-create")
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
			{"admin", "admin", f.subject, 201, ""},
			{"downgraded-viewer", "viewer", f.subject, 403, "authorization_denied"},
			{"removed", "", f.subject, 404, "incident_not_found"},
			{"deployment-admin-without-membership", "", f.deploymentAdmin, 404, "incident_not_found"},
		} {
			if !t.Run(tc.name, func(t *testing.T) {
				f.setSubjectRole(t, tc.role)

				target := flowtest.SeedActor(t, f.harness.DB, flowtest.ActorAttributes{Email: uuid.NewString() + "@example.test", DisplayName: "Membership target"}).ID.String()
				path := "/api/v1/incidents/" + f.incidentID + "/memberships"
				body := map[string]any{"client_txn_id": uuid.NewString(), "user_id": target, "role": "viewer"}
				before := snapshotMembershipAuthorization(t, f, target)
				if index == 0 {
					requireWriteAuthentication(t, f, http.MethodPost, path, body)
				}
				resp := authorizationWrite(t, f, tc.actor, http.MethodPost, path, body)
				data := requireAuthorizationResponse(t, resp, tc.status, tc.code)
				if tc.code == "" && (data["user_id"] != target || data["role"] != "viewer" || data["membership_version"] != float64(1)) {
					t.Fatalf("created membership = %#v", data)
				}
				if tc.code != "" {
					if after := snapshotMembershipAuthorization(t, f, target); after != before {
						t.Fatal("denied membership mutation changed protected state")
					}
				} else {
					var role, actor string
					var version int64
					if err := f.harness.DB.QueryRow(`SELECT role,membership_version,updated_by_user_id::text FROM incident_memberships WHERE incident_id::text=$1 AND user_id::text=$2`, f.incidentID, target).Scan(&role, &version, &actor); err != nil {
						t.Fatal(err)
					}
					if role != data["role"] || version != authorizationVersion(t, data["membership_version"]) || actor != f.subjectID {
						t.Fatalf("persisted membership = %s/%d/%s", role, version, actor)
					}
					requireIncidentAuthorizationAudit(t, f, resp.Header.Get("X-Request-ID"), "incident_membership_created", target)
					requireAuthorizationReceipt(t, f, body["client_txn_id"].(string), f.incidentID, 201)
				}

			}) {
				return
			}
		}
	})
	t.Run("patch", func(t *testing.T) {
		f := newIncidentAuthorization(t, "authz-membership-patch")
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

				target := flowtest.SeedActor(t, f.harness.DB, flowtest.ActorAttributes{Email: uuid.NewString() + "@example.test", DisplayName: "Membership target"}).ID.String()
				path := "/api/v1/incidents/" + f.incidentID + "/memberships"
				membership := scenariotest.CreateMembership(t, f.harness.Server, f.controller, f.incidentID, map[string]any{"client_txn_id": uuid.NewString(), "user_id": target, "role": "viewer"})
				path += "/" + target
				body := map[string]any{"base_membership_version": membership["membership_version"]}
				body["role"] = "reviewer"
				before := snapshotMembershipAuthorization(t, f, target)
				if index == 0 {
					requireWriteAuthentication(t, f, http.MethodPatch, path, body)
				}
				resp := authorizationWrite(t, f, tc.actor, http.MethodPatch, path, body)
				data := requireAuthorizationResponse(t, resp, tc.status, tc.code)
				if tc.code == "" && (data["user_id"] != target || data["role"] != "reviewer" || authorizationVersion(t, data["membership_version"]) != authorizationVersion(t, membership["membership_version"])+1) {
					t.Fatalf("patched membership = %#v", data)
				}
				if tc.code != "" {
					if after := snapshotMembershipAuthorization(t, f, target); after != before {
						t.Fatal("denied membership mutation changed protected state")
					}
				} else {
					var role, actor string
					var version int64
					if err := f.harness.DB.QueryRow(`SELECT role,membership_version,updated_by_user_id::text FROM incident_memberships WHERE incident_id::text=$1 AND user_id::text=$2`, f.incidentID, target).Scan(&role, &version, &actor); err != nil {
						t.Fatal(err)
					}
					if role != data["role"] || version != authorizationVersion(t, data["membership_version"]) || actor != f.subjectID {
						t.Fatalf("persisted membership = %s/%d/%s", role, version, actor)
					}
					requireIncidentAuthorizationAudit(t, f, resp.Header.Get("X-Request-ID"), "incident_membership_updated", target)
				}

			}) {
				return
			}
		}
	})
	t.Run("delete", func(t *testing.T) {
		f := newIncidentAuthorization(t, "authz-membership-delete")
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
			{"admin", "admin", f.subject, 204, ""},
			{"downgraded-viewer", "viewer", f.subject, 403, "authorization_denied"},
			{"removed", "", f.subject, 404, "incident_not_found"},
			{"deployment-admin-without-membership", "", f.deploymentAdmin, 404, "incident_not_found"},
		} {
			if !t.Run(tc.name, func(t *testing.T) {
				f.setSubjectRole(t, tc.role)

				target := flowtest.SeedActor(t, f.harness.DB, flowtest.ActorAttributes{Email: uuid.NewString() + "@example.test", DisplayName: "Membership target"}).ID.String()
				path := "/api/v1/incidents/" + f.incidentID + "/memberships"
				membership := scenariotest.CreateMembership(t, f.harness.Server, f.controller, f.incidentID, map[string]any{"client_txn_id": uuid.NewString(), "user_id": target, "role": "viewer"})
				path += "/" + target
				body := map[string]any{"base_membership_version": membership["membership_version"]}
				before := snapshotMembershipAuthorization(t, f, target)
				if index == 0 {
					requireWriteAuthentication(t, f, http.MethodDelete, path, body)
				}
				resp := authorizationWrite(t, f, tc.actor, http.MethodDelete, path, body)
				requireAuthorizationResponse(t, resp, tc.status, tc.code)
				if tc.code != "" {
					if after := snapshotMembershipAuthorization(t, f, target); after != before {
						t.Fatal("denied membership mutation changed protected state")
					}
				} else {
					var remaining int
					if err := f.harness.DB.QueryRow(`SELECT count(*) FROM incident_memberships WHERE incident_id::text=$1 AND user_id::text=$2`, f.incidentID, target).Scan(&remaining); err != nil {
						t.Fatal(err)
					}
					if remaining != 0 {
						t.Fatalf("deleted membership still has %d rows", remaining)
					}
					requireIncidentAuthorizationAudit(t, f, resp.Header.Get("X-Request-ID"), "incident_membership_deleted", target)
				}

			}) {
				return
			}
		}
	})
}
