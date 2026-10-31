package networkflow_test

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	. "github.com/JochiRaider/cartulary/internal/modules/networkflow"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/cryptography"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/networkflowfixtures"
	"github.com/google/uuid"
)

func (r *publicationFixtureRuntime) errorRequest(method, path string, body any, status int, code string) map[string]any {
	response := httptestx.RequireErrorEnvelope(r.t, httptestx.DoJSON(r.t, http.DefaultClient, method, r.h.Server.HTTP.URL+path, body, r.requestOptions...), status, code)
	raw, err := json.Marshal(response)
	if err != nil {
		r.t.Fatal(err)
	}
	r.transport = append(r.transport, map[string]any{"method": method, "path": path, "body": body, "status": status, "response": fixtureResponseEvidence(raw)})
	result := response["error"].(map[string]any)
	result["request_id"] = "$request-id"
	return result
}

// Ciphertext randomness remains production-owned. Compare every authenticated
// payload member against independent vectors and retain the exact opaque token.
func (r *publicationFixtureRuntime) cursorPayload(token string) map[string]any {
	const prefix = "nfc3.test-cursor."
	if !strings.HasPrefix(token, prefix) || len(token) > 4096 {
		r.t.Fatal("invalid cursor wire format")
	}
	sealed, err := base64.RawURLEncoding.Strict().DecodeString(strings.TrimPrefix(token, prefix))
	if err != nil {
		r.t.Fatal(err)
	}
	key, err := cryptography.AdmitKey(bytes.Repeat([]byte{1}, 32))
	if err != nil {
		r.t.Fatal(err)
	}
	plaintext, err := key.Open("cartulary.network_flow.cursor.v3", []string{"test-cursor"}, sealed, []byte("nfc3."))
	if err != nil {
		r.t.Fatal(err)
	}
	var payload map[string]any
	if err := json.Unmarshal(plaintext, &payload); err != nil {
		r.t.Fatal(err)
	}
	if payload["session_id"] != r.authSession.String() {
		r.t.Fatal("cursor is not bound to the authenticated session")
	}
	payload["session_id"] = "$authenticated-session"
	return payload
}
func (r *publicationFixtureRuntime) normalizedPage(page map[string]any, label string) (map[string]any, string, any) {
	paging := page["meta"].(map[string]any)["paging"].(map[string]any)
	if paging["next_cursor_token"] == nil {
		return page, "", nil
	}
	token, ok := paging["next_cursor_token"].(string)
	if !ok {
		r.t.Fatal("cursor is not a string")
	}
	payload := r.cursorPayload(token)
	paging["next_cursor_token"] = "<redacted>:cursor:" + label
	return page, token, payload
}
func (r *publicationFixtureRuntime) alternateActor(incident string) func() {
	previous := r.requestOptions
	actor := uuid.MustParse("00000000-0000-4000-8000-000000000099")
	flowtest.SeedActor(r.t, r.h.DB, flowtest.ActorAttributes{ID: actor, Email: "other-fixture@example.test", DisplayName: "Other fixture analyst"})
	appsupport.SeedMembershipSQL(r.t, r.h.DB, uuid.MustParse(incident), actor, "viewer", uuid.MustParse("00000000-0000-4000-8000-000000000001"))
	keys, err := authn.LoadMasterKeys(map[string]string{authn.AuthMasterKeyEnv: strings.TrimSpace(string(fixtures.MustRead("auth", "master-key.base64")))})
	if err != nil {
		r.t.Fatal(err)
	}
	const token = "network-flow-alternate-fixture-session"
	flowtest.SeedSession(r.t, r.h.DB, keys, actor, token, time.Now().UTC(), time.Now().UTC())
	csrf := authn.CSRFTokenForSessionToken(keys, token)
	r.requestOptions = []func(*http.Request){httptestx.WithCookies(&http.Cookie{Name: authn.SessionCookieName, Value: token}, &http.Cookie{Name: authn.CSRFCookieName, Value: csrf}), httptestx.WithHeader(authn.CSRFHeaderName, csrf)}
	return func() { r.requestOptions = previous }
}
func runPublicationFamily(t *testing.T, bundle *networkflowfixtures.Bundle, after func(*publicationFixtureRuntime, publicationFixtureCase, map[string]any) error) ([]map[string]any, error) {
	raw, err := bundle.Read("source/cases.json")
	if err != nil {
		return nil, err
	}
	var inputs []publicationFixtureCase
	if err := json.Unmarshal(raw, &inputs); err != nil {
		return nil, err
	}
	results := []map[string]any{}
	for _, input := range inputs {
		var result map[string]any
		var failure error
		passed := t.Run(input.Name, func(t *testing.T) {
			result, failure = runPublicationFixture(t, bundle, input, nil, func(r *publicationFixtureRuntime, result map[string]any) error { return after(r, input, result) })
		})
		if failure != nil {
			return nil, failure
		}
		if !passed {
			return nil, fmt.Errorf("%s: family adapter failed", input.Name)
		}
		results = append(results, result)
	}
	return results, nil
}
func TestNetworkFlowPagingFixtureComparison(t *testing.T) {
	RunFixtureComparison(t, "NF-FIX-014-cursor-pagination", 5, func(bundle *networkflowfixtures.Bundle) ([]map[string]any, error) {
		return runPublicationFamily(t, bundle, func(r *publicationFixtureRuntime, input publicationFixtureCase, result map[string]any) error {
			path := "/api/v1/incidents/" + input.IncidentID + "/network-flow/tables/nft_" + strings.ReplaceAll(input.TableUUID, "-", "")
			initial := map[string]any{"schema_id": "cartulary.network_flow.table_query_request.v1", "limit": 1, "sort": []map[string]any{{"field_key": "network_flow.bytes_count", "direction": "desc"}}}
			first, token, payload := r.normalizedPage(r.request(http.MethodPost, path+"/query", initial, http.StatusOK), "first")
			if token == "" {
				return fmt.Errorf("first page must have a continuation")
			}
			continuation := func(token string) map[string]any {
				return map[string]any{"schema_id": "cartulary.network_flow.table_query_continuation.v1", "cursor_token": token}
			}
			restore := r.alternateActor(input.IncidentID)
			mismatch := r.errorRequest(http.MethodPost, path+"/query", continuation(token), 400, "network_flow_cursor_invalid")
			restore()
			renamed := r.request(http.MethodPatch, path, map[string]any{"client_txn_id": "fixture-rename", "base_table_version": 1, "display_name": "Renamed flows"}, 200)
			second, next, payload2 := r.normalizedPage(r.request(http.MethodPost, path+"/query", continuation(token), 200), "second")
			if next == "" {
				return fmt.Errorf("second page must have a continuation")
			}
			third, last, payload3 := r.normalizedPage(r.request(http.MethodPost, path+"/query", continuation(next), 200), "third")
			if last != "" || payload3 != nil {
				return fmt.Errorf("last page must terminate")
			}
			result["paging"] = map[string]any{"initial_request": initial, "first": first, "first_cursor": payload, "actor_mismatch": mismatch, "rename": renamed, "second": second, "second_cursor": payload2, "third": third, "durable_after": r.comparableState(r.durableState())}
			// Read-only paging and rejected cross-actor continuation do not create receipts.
			result["paging"].(map[string]any)["mutation_receipts"] = r.jsonQuery(`SELECT to_jsonb(count(*)) FROM route_idempotency WHERE route_key LIKE 'nf.%'`)
			return nil
		})
	})
}

type queryNormalizationProbe struct {
	Name         string          `json:"name"`
	Route        string          `json:"route"`
	Body         json.RawMessage `json:"body"`
	CursorLength int             `json:"cursor_length,omitempty"`
	Code         string          `json:"code,omitempty"`
}
type queryNormalizationFixtureCase struct {
	publicationFixtureCase
	Probes []queryNormalizationProbe `json:"probes"`
}

func (r *publicationFixtureRuntime) cursorAtExpiry(token string, expires time.Time) string {
	payload := r.cursorPayload(token)
	payload["session_id"] = r.authSession.String()
	payload["issued_at"] = expires.Add(-15 * time.Minute).Format(time.RFC3339Nano)
	payload["expires_at"] = expires.Format(time.RFC3339Nano)
	raw, err := json.Marshal(payload)
	if err != nil {
		r.t.Fatal(err)
	}
	key, err := cryptography.AdmitKey(bytes.Repeat([]byte{1}, 32))
	if err != nil {
		r.t.Fatal(err)
	}
	sealed, err := key.Seal("cartulary.network_flow.cursor.v3", []string{"test-cursor"}, raw, []byte("nfc3."))
	if err != nil {
		r.t.Fatal(err)
	}
	return "nfc3.test-cursor." + base64.RawURLEncoding.EncodeToString(sealed)
}
func TestNetworkFlowQueryNormalizationFixtureComparison(t *testing.T) {
	RunFixtureComparison(t, "NF-FIX-024-query-normalization-cursors", 5, func(bundle *networkflowfixtures.Bundle) ([]map[string]any, error) {
		raw, err := bundle.Read("source/cases.json")
		if err != nil {
			return nil, err
		}
		var inputs []queryNormalizationFixtureCase
		if err := json.Unmarshal(raw, &inputs); err != nil {
			return nil, err
		}
		results := []map[string]any{}
		for _, input := range inputs {
			result, err := runPublicationFixture(t, bundle, input.publicationFixtureCase, nil, func(r *publicationFixtureRuntime, result map[string]any) error {
				root := "/api/v1/incidents/" + input.IncidentID + "/network-flow"
				path := root + "/tables/nft_" + strings.ReplaceAll(input.TableUUID, "-", "")
				probes := []map[string]any{}
				for _, probe := range input.Probes {
					endpoint := path + "/query"
					switch probe.Route {
					case "table":
					case "rows":
						endpoint = root + "/rows/query"
					case "graph":
						endpoint = root + "/graphs/query"
					case "saved_graph":
						endpoint = root + "/graph-views"
					default:
						return fmt.Errorf("unknown query fixture route %q", probe.Route)
					}
					var body map[string]any
					if err := json.Unmarshal(probe.Body, &body); err != nil {
						return err
					}
					if probe.CursorLength > 0 {
						body["cursor_token"] = strings.Repeat("x", probe.CursorLength)
					}
					var response map[string]any
					if probe.Code == "" {
						response = r.request(http.MethodPost, endpoint, body, 200)
					} else {
						before := r.queryAdmissionState()
						response = r.errorRequest(http.MethodPost, endpoint, body, 400, probe.Code)
						if !reflect.DeepEqual(before, r.queryAdmissionState()) {
							return fmt.Errorf("rejected query %s mutated state", probe.Name)
						}
					}
					probes = append(probes, map[string]any{"name": probe.Name, "response": response})
				}
				rows, token, payload := r.normalizedPage(r.request(http.MethodPost, path+"/query", map[string]any{"schema_id": "cartulary.network_flow.table_query_request.v1", "limit": 1}, 200), "row")
				if token == "" {
					return fmt.Errorf("row cursor is required")
				}
				diagPages := []map[string]any{}
				diagPayloads := []any{}
				next := ""
				firstDiagToken := ""
				for index := 0; index < 4; index++ {
					body := map[string]any{"schema_id": "cartulary.network_flow.rejected_rows_query_request.v1", "limit": 1}
					if index > 0 {
						body = map[string]any{"schema_id": "cartulary.network_flow.rejected_rows_query_continuation.v1", "cursor_token": next}
					}
					page, current, position := r.normalizedPage(r.request(http.MethodPost, path+"/rejected-rows/query", body, 200), fmt.Sprintf("diagnostic-%d", index+1))
					if index == 0 {
						firstDiagToken = current
					}
					diagPages = append(diagPages, page)
					diagPayloads = append(diagPayloads, position)
					next = current
					if index < 3 && current == "" {
						return fmt.Errorf("diagnostic traversal ended early")
					}
				}
				if next != "" {
					return fmt.Errorf("diagnostic traversal did not terminate")
				}
				rowContinuation := func(value string) map[string]any {
					return map[string]any{"schema_id": "cartulary.network_flow.table_query_continuation.v1", "cursor_token": value}
				}
				crossRow := r.errorRequest(http.MethodPost, path+"/query", rowContinuation(firstDiagToken), 400, "network_flow_cursor_invalid")
				crossDiag := r.errorRequest(http.MethodPost, path+"/rejected-rows/query", map[string]any{"schema_id": "cartulary.network_flow.rejected_rows_query_continuation.v1", "cursor_token": token}, 400, "network_flow_cursor_invalid")
				now := time.Date(2026, 7, 10, 12, 0, 0, 0, time.UTC)
				before, _, beforePayload := r.normalizedPage(r.request(http.MethodPost, path+"/query", rowContinuation(r.cursorAtExpiry(token, now.Add(time.Nanosecond))), 200), "expiry-before")
				expired := r.errorRequest(http.MethodPost, path+"/query", rowContinuation(r.cursorAtExpiry(token, now)), 400, "network_flow_cursor_invalid")
				result["normalization"] = map[string]any{"probes": probes, "row_page": rows, "row_cursor": payload, "diagnostic_pages": diagPages, "diagnostic_cursors": diagPayloads, "row_cursor_on_diagnostics": crossDiag, "diagnostic_cursor_on_rows": crossRow, "before_expiry": before, "before_expiry_cursor": beforePayload, "at_expiry": expired, "durable_after": r.comparableState(r.durableState())}
				return nil
			})
			if err != nil {
				return nil, err
			}
			results = append(results, result)
		}
		return results, nil
	})
}

// Query rejection is read-only across owner publication and graph/job effects.
func (r *publicationFixtureRuntime) queryAdmissionState() map[string]any {
	state := r.durableState()
	for _, table := range []string{"network_flow_graph_views", "graph_projection_results", "route_idempotency", "jobs", "extension_job_commit_proofs"} {
		state[table] = r.jsonQuery("SELECT coalesce(jsonb_agg(to_jsonb(v) ORDER BY to_jsonb(v)::text),'[]') FROM " + table + " v")
	}
	return state
}
