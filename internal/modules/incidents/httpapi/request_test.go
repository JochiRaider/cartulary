package httpapi

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/platform/contracttest"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
)

func TestIncidentCreateAcceptsDeclaredMembersAndNormalizesIncidentKey_Unit(t *testing.T) {
	_, apiErr := admitIncidentCreateJSON(strings.NewReader(`{
		"client_txn_id":"txn-u-2-01",
		"incident_key":"  IR-E\u0301-2026-001  ",
		"title":"  Incident E\u0301xample  ",
		"description":"  First line\r\nSecond line  ",
		"severity":"  high  ",
		"tlp":"TLP:AMBER",
		"current_phase":"  triage  ",
		"primary_external_case_ref":"  CASE-42  "
	}`))
	if apiErr != nil {
		t.Fatalf("decode valid incident create request: %v", apiErr)
	}
	_, apiErr = admitIncidentCreateJSON(strings.NewReader(`{
		"client_txn_id":"txn-u-2-01-invalid-memberships",
		"incident_key":"IR-2026-002",
		"title":"Example",
		"initial_memberships":[]
	}`))
	requireAPIError(t, apiErr, http.StatusBadRequest, "invalid_incident_create", "initial_memberships", "collaborator_seeding_not_supported")

	_, apiErr = admitIncidentCreateJSON(strings.NewReader(`{
		"client_txn_id":"txn-u-2-01-invalid-top-level",
		"incident_key":"IR-2026-003",
		"title":"Example",
		"unexpected":true
	}`))
	requireAPIError(t, apiErr, http.StatusBadRequest, "invalid_incident_create", "unexpected", "unknown_field")

	_, apiErr = admitIncidentCreateJSON(strings.NewReader(`{
		"client_txn_id":"txn-u-2-01-server-managed",
		"incident_key":"IR-2026-004",
		"title":"Example",
		"incident_version":1
	}`))
	requireAPIError(t, apiErr, http.StatusBadRequest, "invalid_incident_create", "incident_version", "server_managed_field")

	_, apiErr = admitIncidentCreateJSON(strings.NewReader(`{
		"client_txn_id":"txn-u-2-01-invalid-tlp",
		"incident_key":"IR-2026-005",
		"title":"Example",
		"tlp":"amber"
	}`))
	requireAPIError(t, apiErr, http.StatusBadRequest, "invalid_incident_create", "tlp", "invalid_value")
}

func TestIncidentRequestStringContracts(t *testing.T) {
	t.Run("membership query and cursor contract", testMembershipQueryAndCursorContract)
	_, tooLarge := admitIncidentCreateJSON(strings.NewReader(strings.Repeat(" ", 262145)))
	if tooLarge == nil || tooLarge.Status != 413 || tooLarge.Code != "request_body_too_large" || tooLarge.Details["limit_bytes"] != 262144 {
		t.Fatalf("oversized body response=%#v", tooLarge)
	}

	createPayload := func(overrides map[string]any) string {
		payload := map[string]any{
			"client_txn_id": "txn-string-contract",
			"incident_key":  "IR-STRING-CONTRACT",
			"title":         "String Contract",
		}
		for key, value := range overrides {
			payload[key] = value
		}
		encoded, err := json.Marshal(payload)
		if err != nil {
			t.Fatalf("marshal create payload: %v", err)
		}
		return string(encoded)
	}

	createCases := []struct {
		name  string
		field string
		value any
	}{
		{name: "incident key max bytes", field: "incident_key", value: strings.Repeat("\u00e9", 65)},
		{name: "title max scalar values", field: "title", value: strings.Repeat("a", 513)},
		{name: "description max scalar values", field: "description", value: strings.Repeat("a", 16385)},
		{name: "metadata max scalar values", field: "severity", value: strings.Repeat("a", 129)},
		{name: "tlp exact token", field: "tlp", value: " TLP:AMBER "},
	}
	for _, tc := range createCases {
		t.Run(tc.name, func(t *testing.T) {
			_, apiErr := admitIncidentCreateJSON(strings.NewReader(createPayload(map[string]any{tc.field: tc.value})))
			requireAPIError(t, apiErr, http.StatusBadRequest, "invalid_incident_create", tc.field, "invalid_value")
		})
	}

	patchPayload, err := json.Marshal(map[string]any{
		"base_incident_version": 3,
		"current_phase":         " \t ",
	})
	if err != nil {
		t.Fatalf("marshal patch payload: %v", err)
	}
	_, apiErr := admitIncidentPatchJSON(strings.NewReader(string(patchPayload)))
	if apiErr != nil {
		t.Fatalf("decode whitespace metadata clear patch: %v", apiErr)
	}
	_, apiErr = admitIncidentPatchJSON(strings.NewReader(`{
		"base_incident_version":3,
		"tlp":" TLP:AMBER "
	}`))
	requireAPIError(t, apiErr, http.StatusBadRequest, "invalid_incident_patch", "tlp", "invalid_value")
}

func TestIncidentListQueryUsesCoreListQueryErrors(t *testing.T) {
	valid, apiErr := parseIncidentListScope("search=CASE+High&status=active&limit=1")
	if apiErr != nil {
		t.Fatalf("parse valid incident list query: %v", apiErr)
	}
	if valid.Scope["search"] != "case high" || valid.Scope["status"] != "active" {
		t.Fatalf("unexpected canonical incident list scope: %#v", valid.Scope)
	}

	cases := []struct {
		name       string
		query      string
		code       string
		reasonCode string
	}{
		{
			name:       "unknown query member",
			query:      "group_by=status",
			code:       "invalid_list_query",
			reasonCode: "unknown_query_member",
		},
		{
			name:       "duplicate member",
			query:      "status=active&status=closed",
			code:       "invalid_list_query",
			reasonCode: "duplicate_query_member",
		},
		{
			name:       "invalid status",
			query:      "status=archived",
			code:       "invalid_list_query",
			reasonCode: "invalid_filter_value",
		},
		{
			name:       "pagination alias",
			query:      "page=2",
			code:       "invalid_pagination_request",
			reasonCode: "invalid_limit",
		},
		{
			name:       "invalid search",
			query:      "search=---",
			code:       "invalid_list_query",
			reasonCode: "invalid_search",
		},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, apiErr := parseIncidentListScope(tc.query)
			requireAPIError(t, apiErr, http.StatusBadRequest, tc.code, "", tc.reasonCode)
		})
	}
}

func TestIncidentPatchAllowsPromotedFieldsAndKeepsNoOpVersionStable_Unit(t *testing.T) {
	_, apiErr := admitIncidentPatchJSON(strings.NewReader(`{
		"base_incident_version":7,
		"tlp":null,
		"current_phase":"  containment  "
	}`))
	if apiErr != nil {
		t.Fatalf("decode valid incident patch request: %v", apiErr)
	}
	_, apiErr = admitIncidentPatchJSON(strings.NewReader(`{
		"base_incident_version":7,
		"title":"forbidden"
	}`))
	requireAPIError(t, apiErr, http.StatusBadRequest, "invalid_incident_patch", "title", "forbidden_field")

	_, apiErr = admitIncidentPatchJSON(strings.NewReader(`{
		"tlp":"TLP:AMBER"
	}`))
	requireAPIError(t, apiErr, http.StatusBadRequest, "invalid_incident_patch", "base_incident_version", "missing_required_field")

	_, apiErr = admitIncidentPatchJSON(strings.NewReader(`{
		"base_incident_version":7,
		"tlp":"amber"
	}`))
	requireAPIError(t, apiErr, http.StatusBadRequest, "invalid_incident_patch", "tlp", "invalid_value")

	_, apiErr = admitIncidentPatchJSON(strings.NewReader(`{
		"unknown":"field"
	}`))
	requireAPIError(t, apiErr, http.StatusBadRequest, "invalid_incident_patch", "unknown", "unknown_field")
}

func TestMembershipCreateUsesLookupOnlyForUserOrEmailTargets_Unit(t *testing.T) {
	request, apiErr := admitMembershipCreateJSON(strings.NewReader(`{
		"client_txn_id":"txn-u-2-06-email",
		"email":"  Analyst@Example.Test  ",
		"role":"reviewer"
	}`))
	if apiErr != nil {
		t.Fatalf("decode valid membership create request: %v", apiErr)
	}
	email, ok := request.TargetEmail()
	if !ok {
		t.Fatal("expected normalized email target")
	}
	requireWritableStringNormalization(t, email, "Analyst@Example.Test")
	userOnly, apiErr := admitMembershipCreateJSON(strings.NewReader(`{
		"client_txn_id":"txn-u-2-06-user",
		"user_id":"00000000-0000-0000-0000-000000000601",
		"role":"viewer"
	}`))
	if apiErr != nil {
		t.Fatalf("decode user-target membership create request: %v", apiErr)
	}
	_, hasUserID := userOnly.TargetUserID()
	_, hasEmail := userOnly.TargetEmail()
	if !hasUserID || hasEmail {
		t.Fatalf("expected user_id-only membership target, got %#v", userOnly)
	}
	_, apiErr = admitMembershipCreateJSON(strings.NewReader(`{
		"client_txn_id":"txn-u-2-06-missing-target",
		"role":"viewer"
	}`))
	requireAPIError(t, apiErr, http.StatusBadRequest, "invalid_mutation_payload", "user_id", "exactly_one_target_selector")

	_, apiErr = admitMembershipCreateJSON(strings.NewReader(`{
		"client_txn_id":"txn-u-2-06-dual-target",
		"user_id":"00000000-0000-0000-0000-000000000602",
		"email":"dual@example.test",
		"role":"viewer"
	}`))
	requireAPIError(t, apiErr, http.StatusBadRequest, "invalid_mutation_payload", "user_id", "exactly_one_target_selector")

	_, apiErr = admitMembershipCreateJSON(strings.NewReader(`{
		"client_txn_id":"txn-u-2-06-invalid-role",
		"email":"solo@example.test",
		"role":"owner"
	}`))
	requireClosedVocabularyRejected(t, apiErr, "role", "invalid_role")

	_, apiErr = admitMembershipCreateJSON(strings.NewReader(`{
		"client_txn_id":"txn-u-2-06-no-invite",
		"email":"solo@example.test",
		"role":"viewer",
		"invitation_email":"new@example.test"
	}`))
	requireAPIError(t, apiErr, http.StatusBadRequest, "invalid_mutation_payload", "invitation_email", "unknown_field")

}

func requireAPIError(t testing.TB, apiErr *httpapi.APIError, wantStatus int, wantCode string, wantField string, wantReasonCode string) {
	t.Helper()
	if apiErr == nil {
		t.Fatal("expected api error")
		return
	}
	requireErrorContract(t, wantCode, wantStatus)
	if apiErr.Status != wantStatus {
		t.Fatalf("unexpected status: got %d want %d", apiErr.Status, wantStatus)
	}
	if apiErr.Code != wantCode {
		t.Fatalf("unexpected code: got %q want %q", apiErr.Code, wantCode)
	}
	if wantField != "" {
		if got := apiErr.Details["field"]; got != wantField {
			t.Fatalf("unexpected field detail: got %v want %s", got, wantField)
		}
	}
	if wantReasonCode != "" {
		if got := apiErr.Details["reason_code"]; got != wantReasonCode {
			t.Fatalf("unexpected reason_code detail: got %v want %s", got, wantReasonCode)
		}
	}
}

func requireClosedVocabularyRejected(t testing.TB, apiErr *httpapi.APIError, wantField string, wantReasonCode string) {
	t.Helper()
	if apiErr == nil {
		t.Fatal("expected api error")
		return
	}
	if apiErr.Code != "invalid_mutation_payload" && apiErr.Code != "invalid_view_query" {
		t.Fatalf("unexpected closed-vocabulary rejection code: %q", apiErr.Code)
	}
	if apiErr.Details == nil {
		t.Fatal("expected closed-vocabulary rejection details")
	}
	if apiErr.Details["field"] != wantField {
		t.Fatalf("unexpected closed-vocabulary field: got %v want %q", apiErr.Details["field"], wantField)
	}
	if apiErr.Details["reason_code"] != wantReasonCode {
		t.Fatalf("unexpected closed-vocabulary reason_code: got %v want %q", apiErr.Details["reason_code"], wantReasonCode)
	}
}

func requireWritableStringNormalization(t testing.TB, got string, want string) {
	t.Helper()
	if got != want {
		t.Fatalf("unexpected normalized string: got %q want %q", got, want)
	}
}

func toStrings(t testing.TB, raw any) []string {
	t.Helper()
	items, ok := raw.([]string)
	if ok {
		return append([]string(nil), items...)
	}
	interfaces, ok := raw.([]any)
	if !ok {
		t.Fatalf("unexpected string slice payload: %T", raw)
	}
	values := make([]string, 0, len(interfaces))
	for _, item := range interfaces {
		value, ok := item.(string)
		if !ok {
			t.Fatalf("unexpected string item payload: %T", item)
		}
		values = append(values, value)
	}
	return values
}

func equalStringSlices(left []string, right []string) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index] != right[index] {
			return false
		}
	}
	return true
}

func openAPIObjectAt(t testing.TB, root any, path ...string) map[string]any {
	t.Helper()

	current := root
	for _, segment := range path {
		object, ok := current.(map[string]any)
		if !ok {
			t.Fatalf("expected OpenAPI object before %q, got %T", segment, current)
		}
		var found bool
		current, found = object[segment]
		if !found {
			t.Fatalf("missing OpenAPI path segment %q in %#v", segment, object)
		}
	}
	object, ok := current.(map[string]any)
	if !ok {
		t.Fatalf("expected OpenAPI object at %v, got %T", path, current)
	}
	return object
}

func requireErrorContract(t testing.TB, code string, httpStatus int) {
	t.Helper()

	contracttest.RequireErrorContract(t, code, httpStatus)
}

func testMembershipQueryAndCursorContract(t *testing.T) {
	for query, want := range map[string]string{"%xx=x": "malformed_query", "%ff=value": "malformed_query", "limit=%ff": "malformed_query", "limit=1;extra=2": "malformed_query", "limit=1&limit=2": "duplicate_query_member", "extra=x": "unknown_query_member", "search=x": "unknown_query_member", "offset=1": "invalid_limit", "page=1": "invalid_limit", "limit=": "invalid_limit", "cursor_token=": "invalid_cursor_token"} {
		_, err := parseMembershipListQuery(query)
		if err == nil || err.Details["reason_code"] != want {
			t.Fatalf("query=%q error=%#v want=%s", query, err, want)
		}
	}
	binding := pagination.Binding{Route: "incident.memberships.list", ActorUserID: "actor", Limit: 2, Scope: map[string]string{"incident_id": "incident"}}
	cursor := pagination.Cursor{Version: pagination.CursorVersion, Mode: pagination.ModeOffset, Route: binding.Route, ActorUserID: binding.ActorUserID, Limit: 2, Scope: binding.Scope, Position: map[string]string{"offset": "2"}}
	if _, reason := membershipListPageRequest(binding, &cursor); reason != "invalid_cursor_token" {
		t.Fatal("legacy offset accepted")
	}
	cursor.Mode = pagination.ModeKeyset
	cursor.Position = map[string]string{"last_joined_at": time.Now().UTC().Truncate(time.Microsecond).Format(time.RFC3339Nano), "last_user_id": uuid.NewString()}
	if page, reason := membershipListPageRequest(binding, &cursor); reason != "" || page.Limit != 3 || page.After == nil {
		t.Fatalf("keyset rejected: %#v %s", page, reason)
	}
	codec := pagination.NewCodec([32]byte{1})
	token, err := codec.Encode(cursor)
	if err != nil {
		t.Fatal(err)
	}
	for _, scope := range []struct{ actor, incident, limit string }{{"other", "incident", "2"}, {"actor", "other", "2"}, {"actor", "incident", "3"}} {
		query, _ := parseMembershipListQuery("cursor_token=" + token + "&limit=" + scope.limit)
		if _, _, reason := codec.ResolveListRequest(query, binding.Route, scope.actor, map[string]string{"incident_id": scope.incident}); reason != "cursor_query_mismatch" {
			t.Fatalf("binding mismatch=%s", reason)
		}
	}
}
