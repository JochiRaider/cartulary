package reference_data_test

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

type packHTTPCase struct{ method, path string }

var packHTTPOperations = []packHTTPCase{
	{"GET", "/api/v1/reference-packs"},
	{"GET", "/api/v1/reference-packs/type_registry.host/absent"},
	{"GET", "/api/v1/reference-packs/validation-summaries/absent"},
	{"POST", "/api/v1/reference-packs/import"},
	{"POST", "/api/v1/reference-packs/refresh"},
	{"POST", "/api/v1/reference-packs/type_registry.host/absent/activate"},
	{"POST", "/api/v1/reference-packs/type_registry.host/absent/disable"},
	{"POST", "/api/v1/reference-packs/type_registry.host/absent/remove"},
	{"POST", "/api/v1/reference-packs/type_registry.host/absent/reverify"},
}

var packNoncanonicalPaths = []string{
	"/api//v1/reference-packs",
	"/api/./v1/reference-packs",
	"/api/v1/other/../reference-packs",
	"/../api/v1/reference-packs",
	"/api/v1/reference-packs//type_registry.host/%61ctivate",
	"/api/v1/reference-packs/../users",
	"/api/v1/reference-packs/unknown/..",
}

func packHTTPRequest(t testing.TB, h *appsupport.ServerHarness, tc packHTTPCase, login authcookietest.AuthCookies, csrf, bearer bool, body string) *http.Response {
	t.Helper()
	r, err := http.NewRequest(tc.method, h.Server.HTTP.URL+tc.path, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	r.Header.Set("Content-Type", "application/json")
	if login.Session != nil {
		if bearer {
			r.Header.Set("Authorization", "Bearer "+login.Session.Value)
		} else {
			r.AddCookie(login.Session)
			if csrf {
				r.AddCookie(login.CSRF)
				r.Header.Set(authn.CSRFHeaderName, login.CSRF.Value)
			}
		}
	}
	client := &http.Client{CheckRedirect: func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }}
	return httptestx.Do(t, client, r)
}

func requirePackHTTPError(t testing.TB, response *http.Response, status int, code, reason string) {
	t.Helper()
	defer response.Body.Close()
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != status {
		t.Fatalf("status=%d want=%d body=%s", response.StatusCode, status, body)
	}
	if response.Request.Method == http.MethodHead || code == "" {
		return
	}
	var envelope struct {
		Error struct {
			Code    string `json:"code"`
			Details struct {
				Reason string `json:"reason_code"`
				Field  string `json:"field"`
			} `json:"details"`
		} `json:"error"`
	}
	if err := json.Unmarshal(body, &envelope); err != nil {
		t.Fatal(err)
	}
	if envelope.Error.Code != code || reason != "" && envelope.Error.Details.Reason != reason {
		t.Fatalf("error=%s want=%s/%s", body, code, reason)
	}
	if reason == "malformed_query" || reason == "unknown_query_member" {
		if envelope.Error.Details.Field != "query" || strings.Contains(string(body), "private-query-value") {
			t.Fatalf("unsafe query diagnostic: %s", body)
		}
	}
}

// Snapshot authoritative pack rows and shared admission outputs, including
// changes that preserve row counts. Session bookkeeping is a separate owner.
func packAdmissionState(t testing.TB, h *appsupport.ServerHarness) string {
	t.Helper()
	var result strings.Builder
	for _, table := range []string{"reference_pack_candidates", "reference_pack_current_set", "reference_pack_key_state", "reference_pack_operations", "reference_pack_attempts", "reference_pack_events", "reference_pack_objects", "reference_pack_object_refs", "jobs", "route_idempotency"} {
		var digest string
		if err := h.DB.QueryRow("SELECT COALESCE(md5(string_agg(row_to_json(r)::text, '' ORDER BY row_to_json(r)::text)), '') FROM " + table + " r").Scan(&digest); err != nil {
			t.Fatal(err)
		}
		result.WriteString(table + ":" + digest + ";")
	}
	return result.String()
}

func TestReferencePackFamilyAuthorization_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	for _, claimed := range []bool{true, false} {
		t.Run(fmt.Sprintf("claimed=%t", claimed), func(t *testing.T) {
			var h *appsupport.ServerHarness
			if claimed {
				h = startReferencePackServerWithEnv(t, runtime, "reference-family-auth", map[string]string{"CARTULARY__REFERENCE_PACK__CLAIMED": "true", "CARTULARY__IMPORT__CLAIMED": "false"})
			} else {
				h = runtime.StartServer(t, appsupport.ServerOptions{Prefix: "reference-family-unclaimed", TestRouteMode: httptestx.TestRouteModeDisabled, Env: map[string]string{"CARTULARY__REFERENCE_PACK__CLAIMED": "false", "CARTULARY__IMPORT__CLAIMED": "true"}})
			}
			admin, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, h.Server.HTTP.URL)
			created := httptestx.DoJSON(t, http.DefaultClient, "POST", h.Server.HTTP.URL+"/api/v1/users", map[string]any{
				"client_txn_id": "family-create-user", "auth_kind": "local", "email": "family@example.test", "display_name": "Family caller", "initial_password": "ReferencePackPass123!", "mfa_required": false,
			}, csrfOptions(admin)...)
			httptestx.RequireSuccessEnvelope(t, created, http.StatusCreated)
			ordinary := flowtest.LoginLocalUser(t, http.DefaultClient, h.Server.HTTP.URL, "family@example.test", "ReferencePackPass123!", nil)
			before := packAdmissionState(t, h)
			cases := append([]packHTTPCase{}, packHTTPOperations...)
			cases = append(cases,
				packHTTPCase{"PATCH", "/api/v1/reference-packs"},
				packHTTPCase{"HEAD", "/api/v1/reference-packs"},
				packHTTPCase{"OPTIONS", "/api/v1/reference-packs/import"},
				packHTTPCase{"GET", "/api/v1/reference-packs/unknown/child/extra"},
				packHTTPCase{"GET", "/api/v1/reference-packs/type_registry.host/a%2Fb"},
				packHTTPCase{"GET", "/api/v1/reference-packs//unknown"},
				packHTTPCase{"GET", "/api/v1/reference-packs/type_registry.host/activate"})
			for _, path := range packNoncanonicalPaths {
				for _, method := range []string{"GET", "POST", "HEAD", "OPTIONS", "PATCH"} {
					cases = append(cases, packHTTPCase{method, path})
				}
			}
			for _, actor := range []struct {
				name   string
				login  authcookietest.AuthCookies
				status int
				code   string
			}{
				{"anonymous", authcookietest.AuthCookies{}, 401, "session_required"},
				{"ordinary", ordinary, 403, "authorization_denied"},
			} {
				for _, tc := range cases {
					for _, bearer := range []bool{false, true} {
						t.Run(fmt.Sprintf("%s/%s/%s/bearer=%t", actor.name, tc.method, tc.path, bearer), func(t *testing.T) {
							status, code := actor.status, actor.code
							if !claimed {
								status, code = 404, "extension_profile_not_claimed"
							}
							requestCase := packHTTPCase{tc.method, tc.path + "?limit=1&private-query-value=secret"}
							requirePackHTTPError(t, packHTTPRequest(t, h, requestCase, actor.login, false, bearer, "{"), status, code, "")
						})
					}
				}
			}
			for _, tc := range packHTTPOperations {
				t.Run("admin/"+tc.path, func(t *testing.T) {
					if !claimed {
						requirePackHTTPError(t, packHTTPRequest(t, h, tc, admin, false, false, "{"), 404, "extension_profile_not_claimed", "")
					} else if tc.method == "POST" {
						requirePackHTTPError(t, packHTTPRequest(t, h, tc, admin, false, false, "{"), 403, "csrf_verification_failed", "")
						for _, bearer := range []bool{false, true} {
							requirePackHTTPError(t, packHTTPRequest(t, h, tc, admin, true, bearer, "{"), 400, "invalid_reference_pack_request", "")
						}
					} else {
						for _, bearer := range []bool{false, true} {
							response := packHTTPRequest(t, h, tc, admin, false, bearer, "")
							if tc.path == "/api/v1/reference-packs" {
								httptestx.RequireSuccessEnvelope(t, response, http.StatusOK)
							} else {
								requirePackHTTPError(t, response, 404, "reference_pack_not_found", "")
							}
						}
					}
				})
			}
			for _, path := range packNoncanonicalPaths {
				for _, method := range []string{"GET", "POST", "HEAD", "OPTIONS", "PATCH"} {
					t.Run("admin/noncanonical/"+method+path, func(t *testing.T) {
						for _, bearer := range []bool{false, true} {
							response := packHTTPRequest(t, h, packHTTPCase{method, path + "?invalid=%zz"}, admin, false, bearer, "{")
							if response.Header.Get("Location") != "" {
								t.Fatal("noncanonical family path redirected")
							}
							code := ""
							if !claimed {
								code = "extension_profile_not_claimed"
							}
							requirePackHTTPError(t, response, 404, code, "")
						}
					})
				}
			}
			inactiveProfile, inactiveRoot := "reference_pack", "/api/v1/reference-packs"
			if claimed {
				inactiveProfile, inactiveRoot = "import", "/api/v1/import-sessions"
			}
			for _, path := range []string{
				"/api/v1/reference-packs/../import-sessions",
				"/api/v1/import-sessions/../reference-packs",
			} {
				for _, actor := range []struct {
					name  string
					login authcookietest.AuthCookies
				}{{"anonymous", authcookietest.AuthCookies{}}, {"ordinary", ordinary}, {"admin", admin}} {
					for _, method := range []string{"GET", "POST", "HEAD", "OPTIONS", "PATCH"} {
						for _, bearer := range []bool{false, true} {
							t.Run(fmt.Sprintf("cross-family/%s/%s/%s/bearer=%t", actor.name, method, path, bearer), func(t *testing.T) {
								response := packHTTPRequest(t, h, packHTTPCase{method, path + "?invalid=%zz"}, actor.login, false, bearer, "{")
								if response.Header.Get("Location") != "" {
									t.Fatal("cross-family traversal redirected")
								}
								if method == http.MethodHead {
									requirePackHTTPError(t, response, 404, "", "")
									return
								}
								errorBody := httptestx.RequireErrorEnvelope(t, response, 404, "extension_profile_not_claimed")["error"].(map[string]any)
								details := errorBody["details"].(map[string]any)
								if errorBody["retryable"] != false || details["profile_id"] != inactiveProfile || details["route_family"] != inactiveRoot {
									t.Fatalf("wrong inactive family: %#v", errorBody)
								}
							})
						}
					}
				}
			}
			if after := packAdmissionState(t, h); after != before {
				t.Error("rejected requests changed pack or admission state")
			}
			if claimed {
				if _, err := h.DB.Exec("UPDATE users SET is_deployment_admin=true WHERE email='family@example.test'"); err != nil {
					t.Fatal(err)
				}
				tc := packHTTPCase{"GET", "/api/v1/reference-packs/type_registry.host/absent?limit=1"}
				requirePackHTTPError(t, packHTTPRequest(t, h, tc, ordinary, false, false, ""), 400, "invalid_pagination_request", "pagination_not_supported")
				if _, err := h.DB.Exec("UPDATE users SET is_deployment_admin=false WHERE email='family@example.test'"); err != nil {
					t.Fatal(err)
				}
				requirePackHTTPError(t, packHTTPRequest(t, h, tc, ordinary, false, false, ""), 403, "authorization_denied", "")
			}
			requirePackHTTPError(t, packHTTPRequest(t, h, packHTTPCase{"GET", "/api/v1/reference-packs-other"}, ordinary, false, false, ""), 404, "", "")
		})
	}
}

func TestReferencePackExactRoutes_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	h := startReferencePackServer(t, runtime, "reference-exact-routes")
	admin, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, h.Server.HTTP.URL)
	importReferencePack(t, h, admin, "type_registry.host", "activate", "exact-version-import")
	for _, version := range []string{"activate", "%61ctivate"} {
		t.Run(version, func(t *testing.T) {
			response := packHTTPRequest(t, h, packHTTPCase{"GET", "/api/v1/reference-packs/type_registry.host/" + version}, admin, false, false, "")
			body := httptestx.RequireSuccessEnvelope(t, response, 200)
			if body["data"].(map[string]any)["pack_version"] != "activate" {
				t.Fatal("version changed during routing")
			}
		})
	}
	for _, version := range []string{"disable", "remove", "reverify", "%2561ctivate"} {
		t.Run(version, func(t *testing.T) {
			requirePackHTTPError(t, packHTTPRequest(t, h, packHTTPCase{"GET", "/api/v1/reference-packs/type_registry.host/" + version}, admin, false, false, ""), 404, "reference_pack_not_found", "")
		})
	}
	for _, operation := range packHTTPOperations {
		for _, method := range []string{"HEAD", "OPTIONS", "PATCH"} {
			t.Run(method+operation.path, func(t *testing.T) {
				response := packHTTPRequest(t, h, packHTTPCase{method, operation.path}, admin, false, false, "{")
				if response.Header.Get("Allow") != operation.method {
					t.Errorf("Allow=%q want=%q", response.Header.Get("Allow"), operation.method)
				}
				requirePackHTTPError(t, response, 405, "", "")
			})
		}
	}
	requirePackHTTPError(t, packHTTPRequest(t, h, packHTTPCase{"POST", "/api/v1/reference-packs/type_registry.host/remove"}, admin, false, false, "{"), 405, "", "")
	for index, version := range []string{"activate", "disable", "reverify", "remove"} {
		t.Run("mutations/"+version, func(t *testing.T) {
			if version != "activate" {
				bundle := referencePackBundle(t, bundleOptions{
					PackKey: "type_registry.host", PackKind: "type_registry", PackVersion: version,
					ManifestTransform: func(manifest map[string]any) { manifest["pack_release_sequence"] = 1000 + index },
				})
				response := postReferencePackUpload(t, h.Server.HTTP.URL, admin, `{"client_txn_id":"exact-import-`+version+`"}`, bundle, version+".zip", reference_data.MediaTypeZip)
				accepted := httptestx.RequireSuccessEnvelope(t, response, http.StatusAccepted)["data"].(map[string]any)
				if job := requireJob(t, h, admin, accepted["job_id"].(string)); job["status"] != "succeeded" {
					t.Fatalf("fixture import Job: %#v", job)
				}
			}
			literal := "/api/v1/reference-packs/type_registry.host/" + version
			encoded := fmt.Sprintf("/api/v1/reference-packs/type_registry.host/%%%02x%s", version[0], version[1:])
			for _, action := range []string{"activate", "disable", "reverify", "remove"} {
				t.Run(action, func(t *testing.T) {
					if action == "remove" {
						before := packAdmissionState(t, h)
						for _, invalid := range []struct{ member, reason string }{
							{"", "missing_required_field"},
							{`,"reason":null`, "field_not_nullable"},
							{`,"reason":"  "`, "invalid_reason"},
						} {
							body := `{"client_txn_id":"exact-invalid-removal"` + invalid.member + `}`
							for _, path := range []string{literal, encoded} {
								requirePackHTTPError(t, packHTTPRequest(t, h, packHTTPCase{"POST", path + "/remove"}, admin, true, false, body), 400, "invalid_reference_pack_request", invalid.reason)
							}
						}
						if packAdmissionState(t, h) != before {
							t.Fatal("invalid removal reason changed admission state")
						}
					}
					txn := "exact-" + version + "-" + action
					reason := ""
					if action == "remove" {
						reason = "fixture retention complete"
					}
					status := http.StatusOK
					if action == "reverify" {
						status = http.StatusAccepted
					}
					data := httptestx.RequireSuccessEnvelope(t, postAction(t, h, admin, encoded+"/"+action, txn, reason), status)["data"].(map[string]any)
					// The literal spelling must replay the same operation identity.
					replay := httptestx.RequireSuccessEnvelope(t, postAction(t, h, admin, literal+"/"+action, txn, reason), status)["data"].(map[string]any)
					condition, active := reference_data.ConditionDisabled, false
					switch action {
					case "activate":
						condition, active = reference_data.ConditionVerifiedAvailable, true
					case "remove":
						condition = reference_data.ConditionMissing
					case "reverify":
						jobID := data["job_id"].(string)
						if replay["job_id"] != jobID {
							t.Fatal("literal reverify replay admitted a different Job")
						}
						job := requireJob(t, h, admin, jobID)
						if job["status"] != "succeeded" {
							t.Fatalf("reverify Job: %#v", job)
						}
						var key, target string
						if err := h.DB.QueryRow(`SELECT m.pack_key,m.pack_version FROM reference_pack_operation_members m JOIN reference_pack_operations o USING(operation_id) WHERE o.job_id=$1`, jobID).Scan(&key, &target); err != nil {
							t.Fatal(err)
						}
						if key != "type_registry.host" || target != version {
							t.Fatalf("reverify targeted %s/%s", key, target)
						}
					}
					if action != "reverify" {
						for _, result := range []map[string]any{data, replay} {
							requireReferencePackResource(t, result["pack_version"].(map[string]any), "type_registry.host", version, condition, active)
						}
					}
					resource := readReferencePack(t, h, admin, "type_registry.host", version)
					requireReferencePackResource(t, resource, "type_registry.host", version, condition, active)
					if action == "remove" && resource["removed"] != true {
						t.Fatal("exact version was not removed")
					}
				})
			}
		})
	}
}

func TestReferencePackClosedQueries_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	h := startReferencePackServer(t, runtime, "reference-closed-queries")
	admin, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, h.Server.HTTP.URL)
	before := packAdmissionState(t, h)
	for _, operation := range packHTTPOperations[1:] {
		queries := []struct{ query, code, reason string }{
			{"private-query-value=secret", "invalid_reference_pack_request", "unknown_query_member"},
			{"private-query-value=1&private-query-value=2", "invalid_reference_pack_request", "unknown_query_member"},
			{"%zz=secret&limit=1", "invalid_reference_pack_request", "malformed_query"},
			{"private-query-value=%ff", "invalid_reference_pack_request", "malformed_query"},
			{"%ff=secret", "invalid_reference_pack_request", "malformed_query"},
			{"a=1;b=2", "invalid_reference_pack_request", "malformed_query"},
		}
		for _, key := range []string{"limit", "cursor_token", "page", "offset", "page_size", "block_size"} {
			queries = append(queries, struct{ query, code, reason string }{key + "=1&" + key + "=2&private-query-value=secret", "invalid_pagination_request", "pagination_not_supported"})
		}
		for _, q := range queries {
			t.Run(operation.path+"/"+q.query, func(t *testing.T) {
				tc := packHTTPCase{operation.method, operation.path + "?" + q.query}
				requirePackHTTPError(t, packHTTPRequest(t, h, tc, admin, true, false, "{"), 400, q.code, q.reason)
			})
		}
	}
	if after := packAdmissionState(t, h); after != before {
		t.Error("query rejection changed pack or admission state")
	}
	// A committed receipt does not bypass current query admission.
	tc := packHTTPCase{"POST", "/api/v1/reference-packs/refresh"}
	body := `{"client_txn_id":"closed-query-replay"}`
	accepted := httptestx.RequireSuccessEnvelope(t, packHTTPRequest(t, h, tc, admin, true, false, body), 202)
	requireJob(t, h, admin, accepted["data"].(map[string]any)["job_id"].(string))
	before = packAdmissionState(t, h)
	tc.path += "?private-query-value=secret"
	requirePackHTTPError(t, packHTTPRequest(t, h, tc, admin, true, false, body), 400, "invalid_reference_pack_request", "unknown_query_member")
	if after := packAdmissionState(t, h); after != before {
		t.Error("rejected replay changed admission state")
	}
}
