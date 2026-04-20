package incidents

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/gen/contracts"
	"github.com/JochiRaider/cartulary/internal/modules/auth"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

func TestPhase2_U_2_01_IncidentCreateAcceptsDeclaredMembersAndNormalizesIncidentKey(t *testing.T) {
	request, apiErr := DecodeIncidentCreateRequest(strings.NewReader(`{
		"client_txn_id":"txn-u-2-01",
		"incident_key":"  IR-E\u0301-2026-001  ",
		"title":"  Incident E\u0301xample  ",
		"description":"  First line\r\nSecond line  ",
		"severity":"  high  ",
		"tlp":"  amber  ",
		"current_phase":"  triage  ",
		"primary_external_case_ref":"  CASE-42  "
	}`))
	if apiErr != nil {
		t.Fatalf("decode valid incident create request: %v", apiErr)
	}
	requireWritableStringNormalization(t, request.IncidentKey, "IR-\u00C9-2026-001")
	requireWritableStringNormalization(t, request.Title, "Incident \u00C9xample")
	if request.Description == nil {
		t.Fatal("expected normalized description")
	}
	requireWritableStringNormalization(t, *request.Description, "First line\nSecond line")

	_, apiErr = DecodeIncidentCreateRequest(strings.NewReader(`{
		"client_txn_id":"txn-u-2-01-invalid-memberships",
		"incident_key":"IR-2026-002",
		"title":"Example",
		"initial_memberships":[]
	}`))
	requireAPIError(t, apiErr, http.StatusBadRequest, "invalid_incident_create", "initial_memberships", "initial_memberships_not_supported")

	_, apiErr = DecodeIncidentCreateRequest(strings.NewReader(`{
		"client_txn_id":"txn-u-2-01-invalid-top-level",
		"incident_key":"IR-2026-003",
		"title":"Example",
		"unexpected":true
	}`))
	requireAPIError(t, apiErr, http.StatusBadRequest, "invalid_incident_create", "unexpected", "unknown_top_level_member")
}

func TestPhase2_U_2_05_IncidentPatchAllowsPromotedFieldsAndKeepsNoOpVersionStable(t *testing.T) {
	request, apiErr := DecodeIncidentPatchRequest(strings.NewReader(`{
		"base_incident_version":7,
		"tlp":null,
		"current_phase":"  containment  "
	}`))
	if apiErr != nil {
		t.Fatalf("decode valid incident patch request: %v", apiErr)
	}
	if !request.TLP.Present || request.TLP.Value != nil {
		t.Fatalf("expected tlp explicit null clear, got %#v", request.TLP)
	}
	if !request.CurrentPhase.Present || request.CurrentPhase.Value == nil || *request.CurrentPhase.Value != "containment" {
		t.Fatalf("unexpected current_phase patch field: %#v", request.CurrentPhase)
	}
	if request.PrimaryExternalCaseRef.Present {
		t.Fatalf("unexpected omitted field to be marked present: %#v", request.PrimaryExternalCaseRef)
	}

	current := IncidentRecord{
		ID:                     uuid.MustParse("00000000-0000-0000-0000-000000000905"),
		TLP:                    stringRef("amber"),
		CurrentPhase:           stringRef("containment"),
		PrimaryExternalCaseRef: stringRef("CASE-1"),
		UpdatedAt:              timeRef(2026, 4, 17, 12, 0),
		IncidentVersion:        7,
	}
	noOp, changed := ApplyIncidentPatch(current, IncidentPatchRequest{BaseIncidentVersion: 7}, uuid.MustParse("00000000-0000-0000-0000-000000000777"), timeRef(2026, 4, 17, 13, 0))
	if changed {
		t.Fatalf("expected structurally valid no-op patch to remain version-stable: %#v", noOp)
	}
	if noOp.IncidentVersion != current.IncidentVersion || !noOp.UpdatedAt.Equal(current.UpdatedAt) {
		t.Fatalf("no-op patch must keep version and updated_at stable: before=%#v after=%#v", current, noOp)
	}

	material, changed := ApplyIncidentPatch(current, IncidentPatchRequest{
		BaseIncidentVersion:    7,
		TLP:                    OptionalNullableString{Present: true, Value: stringRef("green")},
		PrimaryExternalCaseRef: OptionalNullableString{Present: true, Value: nil},
	}, uuid.MustParse("00000000-0000-0000-0000-000000000777"), timeRef(2026, 4, 17, 13, 0))
	if !changed {
		t.Fatal("expected material patch to change promoted fields")
	}
	if material.IncidentVersion != 8 || material.TLP == nil || *material.TLP != "green" || material.PrimaryExternalCaseRef != nil {
		t.Fatalf("unexpected material patch projection: %#v", material)
	}

	_, apiErr = DecodeIncidentPatchRequest(strings.NewReader(`{
		"base_incident_version":7,
		"title":"forbidden"
	}`))
	requireAPIError(t, apiErr, http.StatusBadRequest, "invalid_incident_patch", "title", "forbidden_field")

	_, apiErr = DecodeIncidentPatchRequest(strings.NewReader(`{
		"tlp":"amber"
	}`))
	requireAPIError(t, apiErr, http.StatusBadRequest, "invalid_incident_patch", "base_incident_version", "missing_required_field")

	_, apiErr = DecodeIncidentPatchRequest(strings.NewReader(`{
		"unknown":"field"
	}`))
	requireAPIError(t, apiErr, http.StatusBadRequest, "invalid_incident_patch", "unknown", "unknown_top_level_member")
}

func TestPhase2_MembershipCreateDecodeRejectsInvalidSelectorsAndInvitationFields_U_2_06(t *testing.T) {
	request, apiErr := DecodeMembershipCreateRequest(strings.NewReader(`{
		"client_txn_id":"txn-u-2-06-email",
		"email":"  Analyst@Example.Test  ",
		"role":"reviewer"
	}`))
	if apiErr != nil {
		t.Fatalf("decode valid membership create request: %v", apiErr)
	}
	if request.Email == nil {
		t.Fatal("expected normalized email target")
	}
	requireWritableStringNormalization(t, *request.Email, "Analyst@Example.Test")

	userOnly, apiErr := DecodeMembershipCreateRequest(strings.NewReader(`{
		"client_txn_id":"txn-u-2-06-user",
		"user_id":"00000000-0000-0000-0000-000000000601",
		"role":"viewer"
	}`))
	if apiErr != nil {
		t.Fatalf("decode user-target membership create request: %v", apiErr)
	}
	if userOnly.UserID == nil || userOnly.Email != nil {
		t.Fatalf("expected user_id-only membership target, got %#v", userOnly)
	}

	_, apiErr = DecodeMembershipCreateRequest(strings.NewReader(`{
		"client_txn_id":"txn-u-2-06-missing-target",
		"role":"viewer"
	}`))
	requireAPIError(t, apiErr, http.StatusBadRequest, "invalid_mutation_payload", "user_id", "exactly_one_target_selector")

	_, apiErr = DecodeMembershipCreateRequest(strings.NewReader(`{
		"client_txn_id":"txn-u-2-06-dual-target",
		"user_id":"00000000-0000-0000-0000-000000000602",
		"email":"dual@example.test",
		"role":"viewer"
	}`))
	requireAPIError(t, apiErr, http.StatusBadRequest, "invalid_mutation_payload", "user_id", "exactly_one_target_selector")

	_, apiErr = DecodeMembershipCreateRequest(strings.NewReader(`{
		"client_txn_id":"txn-u-2-06-invalid-role",
		"email":"solo@example.test",
		"role":"owner"
	}`))
	requireClosedVocabularyRejected(t, apiErr, "role", "invalid_role")

	_, apiErr = DecodeMembershipCreateRequest(strings.NewReader(`{
		"client_txn_id":"txn-u-2-06-no-invite",
		"email":"solo@example.test",
		"role":"viewer",
		"invitation_email":"new@example.test"
	}`))
	requireAPIError(t, apiErr, http.StatusBadRequest, "invalid_mutation_payload", "invitation_email", "unknown_field")
}

func TestPhase2_MembershipPatchAndDeleteDecodeEnforceBaseVersionAndLastAdminGuard_U_2_07(t *testing.T) {
	patchRequest, apiErr := DecodeMembershipPatchRequest(strings.NewReader(`{
		"base_membership_version":5,
		"role":"admin"
	}`))
	if apiErr != nil {
		t.Fatalf("decode valid membership patch request: %v", apiErr)
	}
	if patchRequest.BaseMembershipVersion != 5 || patchRequest.Role != "admin" {
		t.Fatalf("unexpected membership patch request: %#v", patchRequest)
	}

	deleteRequest, apiErr := DecodeMembershipDeleteRequest(strings.NewReader(`{
		"base_membership_version":5
	}`))
	if apiErr != nil {
		t.Fatalf("decode valid membership delete request: %v", apiErr)
	}
	if deleteRequest.BaseMembershipVersion != 5 {
		t.Fatalf("unexpected membership delete request: %#v", deleteRequest)
	}

	_, apiErr = DecodeMembershipDeleteRequest(strings.NewReader(`{}`))
	requireAPIError(t, apiErr, http.StatusBadRequest, "invalid_mutation_payload", "base_membership_version", "missing_required_field")

	nextRole := "reviewer"
	if !WouldLeaveNoIncidentAdmins("admin", 1, &nextRole, false) {
		t.Fatal("demoting the last admin must be rejected")
	}
	if WouldLeaveNoIncidentAdmins("admin", 2, &nextRole, false) {
		t.Fatal("demoting one of two admins must be allowed")
	}
	if !WouldLeaveNoIncidentAdmins("admin", 1, nil, true) {
		t.Fatal("deleting the last admin must be rejected")
	}
}

func TestPhase2_U_2_09_ExtensionDiscoveryReturnsExactSingletonProfileShape(t *testing.T) {
	query := url.Values{"cursor_token": []string{"opaque"}}
	apiErr := auth.ValidateSingletonReadQuery(query)
	requireAPIError(t, apiErr, http.StatusBadRequest, "invalid_pagination_request", "", "pagination_not_supported")

	data := BuildExtensionsResponseData(httpapi.CurrentExtensionProfiles())
	extensions, ok := data["extensions"].([]map[string]any)
	if !ok {
		t.Fatalf("unexpected extensions response data: %#v", data)
	}
	wantProfiles := currentProfileExtensions(t)
	if len(extensions) != len(wantProfiles) {
		t.Fatalf("unexpected extension profile count: got %d want %d", len(extensions), len(wantProfiles))
	}

	for index, want := range wantProfiles {
		got := extensions[index]
		if len(got) != 3 {
			t.Fatalf("extension resource must expose only profile_id, claimed, and route_families: %#v", got)
		}
		if got["profile_id"] != want.ProfileID || got["claimed"] != false {
			t.Fatalf("unexpected extension resource at index %d: %#v", index, got)
		}
		if families := toStrings(t, got["route_families"]); !equalStringSlices(families, want.RouteFamilies) {
			t.Fatalf("unexpected route_families at index %d: got %v want %v", index, families, want.RouteFamilies)
		}
	}
}

func TestPhase2_U_2_10_ReservedExtensionDispatchHonorsBaseRoutesClaimedFamiliesAndOutsideFallback(t *testing.T) {
	importProfile := extensionContract(t, "import")
	enterpriseProfile := extensionContract(t, "enterprise_authentication")

	baseHandler, err := httpapi.NewHandler(httpapi.Options{})
	if err != nil {
		t.Fatalf("build base handler: %v", err)
	}
	ready := performRequest(t, baseHandler, http.MethodGet, "/readyz", nil)
	if ready.Code != http.StatusOK || !strings.Contains(ready.Body.String(), "ready") {
		t.Fatalf("expected base route to retain precedence, got status=%d body=%q", ready.Code, ready.Body.String())
	}

	restoreClaimed := httpapi.SetCurrentExtensionProfilesForTesting([]httpapi.ExtensionProfile{
		{
			ProfileID:     importProfile.ProfileID,
			Claimed:       true,
			RouteFamilies: append([]string(nil), importProfile.RouteFamilies...),
		},
	})
	claimedHandler, err := httpapi.NewHandler(httpapi.Options{
		AdditionalRoutes: []httpapi.RouteRegistrar{
			func(mux *http.ServeMux, deps httpapi.DependencySet) error {
				mux.HandleFunc(importProfile.RouteFamilies[0], func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusNoContent)
				})
				mux.HandleFunc(importProfile.RouteFamilies[0]+"/children", func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusAccepted)
				})
				return nil
			},
		},
	})
	restoreClaimed()
	if err != nil {
		t.Fatalf("build claimed handler: %v", err)
	}
	if got := performRequest(t, claimedHandler, http.MethodGet, importProfile.RouteFamilies[0], nil); got.Code != http.StatusNoContent {
		t.Fatalf("claimed family root must dispatch to the registered route, got %d", got.Code)
	}
	if got := performRequest(t, claimedHandler, http.MethodGet, importProfile.RouteFamilies[0]+"/children", nil); got.Code != http.StatusAccepted {
		t.Fatalf("claimed family descendant must dispatch to the registered route, got %d", got.Code)
	}

	restoreUnclaimed := httpapi.SetCurrentExtensionProfilesForTesting([]httpapi.ExtensionProfile{
		{
			ProfileID:     importProfile.ProfileID,
			Claimed:       false,
			RouteFamilies: append([]string(nil), importProfile.RouteFamilies...),
		},
		{
			ProfileID:     enterpriseProfile.ProfileID,
			Claimed:       false,
			RouteFamilies: append([]string(nil), enterpriseProfile.RouteFamilies...),
		},
	})
	defer restoreUnclaimed()
	unclaimedHandler, err := httpapi.NewHandler(httpapi.Options{})
	if err != nil {
		t.Fatalf("build unclaimed handler: %v", err)
	}

	rootReserved := performRequest(t, unclaimedHandler, http.MethodGet, importProfile.RouteFamilies[0], nil)
	rootBody := readJSONMap(t, rootReserved.Body.String())
	errorBody := rootBody["error"].(map[string]any)
	requireErrorContract(t, "extension_profile_not_claimed", http.StatusNotFound)
	if rootReserved.Code != http.StatusNotFound || errorBody["code"] != "extension_profile_not_claimed" {
		t.Fatalf("unexpected reserved root response: status=%d body=%#v", rootReserved.Code, rootBody)
	}
	rootDetails := errorBody["details"].(map[string]any)
	if rootDetails["profile_id"] != importProfile.ProfileID || rootDetails["route_family"] != importProfile.RouteFamilies[0] {
		t.Fatalf("unexpected reserved root details: %#v", rootDetails)
	}

	descendantReserved := performRequest(t, unclaimedHandler, http.MethodGet, strings.Replace(enterpriseProfile.RouteFamilies[0], "{user_id}", uuid.NewString(), 1)+"/provider", nil)
	descendantBody := readJSONMap(t, descendantReserved.Body.String())
	descendantError := descendantBody["error"].(map[string]any)
	if descendantReserved.Code != http.StatusNotFound || descendantError["code"] != "extension_profile_not_claimed" {
		t.Fatalf("unexpected reserved descendant response: status=%d body=%#v", descendantReserved.Code, descendantBody)
	}
	descendantDetails := descendantError["details"].(map[string]any)
	if descendantDetails["profile_id"] != enterpriseProfile.ProfileID || descendantDetails["route_family"] != enterpriseProfile.RouteFamilies[0] {
		t.Fatalf("unexpected reserved descendant details: %#v", descendantDetails)
	}

	outsideReserved := performRequest(t, unclaimedHandler, http.MethodGet, "/api/v1/outside-reserved-families", nil)
	if outsideReserved.Code != http.StatusNotFound || strings.Contains(outsideReserved.Body.String(), "extension_profile_not_claimed") {
		t.Fatalf("ordinary unknown paths must keep ordinary not-found handling: status=%d body=%q", outsideReserved.Code, outsideReserved.Body.String())
	}
}

func requireAPIError(t testing.TB, apiErr *auth.APIError, wantStatus int, wantCode string, wantField string, wantReasonCode string) {
	t.Helper()
	if apiErr == nil {
		t.Fatal("expected api error")
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

func requireClosedVocabularyRejected(t testing.TB, apiErr *auth.APIError, wantField string, wantReasonCode string) {
	t.Helper()
	if apiErr == nil {
		t.Fatal("expected api error")
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

func extensionContract(t testing.TB, profileID string) extensionProfileContract {
	t.Helper()

	for _, profile := range currentProfileExtensions(t) {
		if profile.ProfileID == profileID {
			return profile
		}
	}
	t.Fatalf("missing extension contract for %q", profileID)
	return extensionProfileContract{}
}

func stringRef(value string) *string {
	return &value
}

func performRequest(t testing.TB, handler http.Handler, method string, target string, body io.Reader) *httptest.ResponseRecorder {
	t.Helper()
	req := httptest.NewRequest(method, target, body)
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)
	return recorder
}

func readJSONMap(t testing.TB, body string) map[string]any {
	t.Helper()
	var payload map[string]any
	if err := json.Unmarshal([]byte(body), &payload); err != nil {
		t.Fatalf("decode JSON body: %v", err)
	}
	return payload
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

func timeRef(year int, month int, day int, hour int, minute int) time.Time {
	return time.Date(year, time.Month(month), day, hour, minute, 0, 0, time.UTC)
}

type errorContract struct {
	Code       string `json:"code"`
	HTTPStatus int    `json:"http_status"`
}

type errorRegistry struct {
	Errors []errorContract `json:"errors"`
}

type extensionProfileContract struct {
	ProfileID     string   `json:"profile_id"`
	RouteFamilies []string `json:"route_families"`
}

type extensionRegistry struct {
	Profiles []extensionProfileContract `json:"profiles"`
}

func requireErrorContract(t testing.TB, code string, httpStatus int) {
	t.Helper()

	artifact, ok := contracts.ContractArtifactIndex["contracts/errors/index.json"]
	if !ok {
		t.Fatal("missing generated error contract registry")
	}

	var registry errorRegistry
	if err := json.Unmarshal([]byte(artifact.JSON), &registry); err != nil {
		t.Fatalf("decode generated error contract registry: %v", err)
	}

	for _, candidate := range registry.Errors {
		if candidate.Code == code {
			if candidate.HTTPStatus != httpStatus {
				t.Fatalf("unexpected contract status for %q: got %d want %d", code, candidate.HTTPStatus, httpStatus)
			}
			return
		}
	}
	t.Fatalf("missing generated error contract for %q", code)
}

func currentProfileExtensions(t testing.TB) []extensionProfileContract {
	t.Helper()

	artifact, ok := contracts.ContractArtifactIndex["contracts/extensions/index.json"]
	if !ok {
		t.Fatal("missing generated extension contract registry")
	}

	var registry extensionRegistry
	if err := json.Unmarshal([]byte(artifact.JSON), &registry); err != nil {
		t.Fatalf("decode generated extension contract registry: %v", err)
	}
	return append([]extensionProfileContract(nil), registry.Profiles...)
}
