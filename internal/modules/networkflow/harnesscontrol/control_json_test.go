package harnesscontrol

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func assertStrictNetworkFlowControlRequests(t *testing.T) {
	t.Helper()
	for _, test := range []struct {
		name, valid string
		decode      func(*http.Request) error
	}{
		{"fault", `{"boundary":"network_flow.import.before_owner_apply","fault_kind":"panic","consume_once":true}`, func(r *http.Request) error { _, err := decodeNetworkFlowFaultRequest(r); return err }},
		{"randomness", `{"stream":"network_flow.cursor_nonce","value_kind":"hex_bytes","values":["000102030405060708090a0b"],"consume_once":true,"exhaustion":"fail_closed"}`, func(r *http.Request) error { _, err := decodeNetworkFlowRandomnessRequest(r); return err }},
		{"authorization", `{"boundary":"network_flow.route.before_authorization","transition_kind":"session_revoked","actor_ref":"actor","incident_ref":"incident","resource_kind":"network_flow_table","resource_ref":"table","consume_once":true}`, func(r *http.Request) error { _, err := decodeNetworkFlowAuthTransitionRequest(r); return err }},
		{"audit", `{"assertion_kind":"zero_occurrences","event_code":"network_flow_table_created","operation_ref":"operation","actor_ref":"actor","incident_ref":"incident","resource_kind":"network_flow_table","resource_ref":"table","baseline_count":0,"expected_final_count":0,"consume_once":true}`, func(r *http.Request) error { _, err := decodeNetworkFlowAuditAssertionRequest(r); return err }},
	} {
		t.Run("strict "+test.name, func(t *testing.T) {
			if err := test.decode(httptest.NewRequest(http.MethodPost, "/", strings.NewReader(test.valid))); err != nil {
				t.Fatal(err)
			}
			invalid := []string{"null", "[]", "{}", test.valid + " {}", `{"unknown":true,` + test.valid[1:], `{"consume_once":true,` + test.valid[1:], strings.Replace(test.valid, `"consume_once":true`, `"consume_once":null`, 1), strings.Replace(test.valid, `,"consume_once":true`, "", 1)}
			for _, raw := range invalid {
				if err := test.decode(httptest.NewRequest(http.MethodPost, "/", strings.NewReader(raw))); err == nil {
					t.Fatalf("invalid control JSON accepted: %s", raw)
				}
			}
		})
	}
}
