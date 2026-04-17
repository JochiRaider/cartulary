package incidents

import (
	"net/http"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"example.com/todo/cartulary/internal/modules/auth"
	"example.com/todo/cartulary/internal/platform/httpapi"
)

func TestPhase2_IncidentCreateContracts_U_2_01(t *testing.T) {
	request, apiErr := DecodeIncidentCreateRequest(strings.NewReader(`{
		"client_txn_id":"txn-incident-1",
		"incident_key":"  IR-2026-\u00a0001  ",
		"title":"  Incident \u00c9xample  ",
		"description":"  First line\r\nSecond line  ",
		"severity":"  high  ",
		"tlp":"  amber  ",
		"current_phase":"  triage  ",
		"primary_external_case_ref":"  CASE-42  "
	}`))
	if apiErr != nil {
		t.Fatalf("decode valid incident create request: %v", apiErr)
	}
	if request.IncidentKey != "IR-2026-\u00a0001" {
		t.Fatalf("unexpected normalized incident_key: %q", request.IncidentKey)
	}
	if request.Title != "Incident Éxample" {
		t.Fatalf("unexpected normalized title: %q", request.Title)
	}
	if request.Description == nil || *request.Description != "First line\nSecond line" {
		t.Fatalf("unexpected normalized description: %#v", request.Description)
	}

	_, apiErr = DecodeIncidentCreateRequest(strings.NewReader(`{
		"client_txn_id":"txn-incident-2",
		"incident_key":"IR-2",
		"title":"Example",
		"initial_memberships":[]
	}`))
	requireAPIError(t, apiErr, http.StatusBadRequest, "invalid_incident_create", "initial_memberships")

	_, apiErr = DecodeIncidentCreateRequest(strings.NewReader(`{
		"client_txn_id":"txn-incident-3",
		"incident_key":"IR-3",
		"title":"Example",
		"unexpected":true
	}`))
	requireAPIError(t, apiErr, http.StatusBadRequest, "invalid_incident_create", "unexpected")
}

func TestPhase2_IncidentPatchContracts_U_2_02(t *testing.T) {
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

	_, apiErr = DecodeIncidentPatchRequest(strings.NewReader(`{
		"base_incident_version":7,
		"title":"forbidden"
	}`))
	requireAPIError(t, apiErr, http.StatusBadRequest, "invalid_incident_patch", "title")

	_, apiErr = DecodeIncidentPatchRequest(strings.NewReader(`{
		"base_incident_version":7,
		"unknown":"field"
	}`))
	requireAPIError(t, apiErr, http.StatusBadRequest, "invalid_incident_patch", "unknown")
}

func TestPhase2_MembershipCreateContracts_U_2_03(t *testing.T) {
	request, apiErr := DecodeMembershipCreateRequest(strings.NewReader(`{
		"client_txn_id":"txn-membership-1",
		"email":"  Analyst@Example.Test  ",
		"role":"reviewer"
	}`))
	if apiErr != nil {
		t.Fatalf("decode valid membership create request: %v", apiErr)
	}
	if request.Email == nil || *request.Email != "Analyst@Example.Test" {
		t.Fatalf("unexpected normalized email: %#v", request.Email)
	}

	_, apiErr = DecodeMembershipCreateRequest(strings.NewReader(`{
		"client_txn_id":"txn-membership-2",
		"user_id":"00000000-0000-0000-0000-000000000001",
		"email":"dual@example.test",
		"role":"viewer"
	}`))
	requireAPIError(t, apiErr, http.StatusBadRequest, "invalid_mutation_payload", "user_id")

	_, apiErr = DecodeMembershipCreateRequest(strings.NewReader(`{
		"client_txn_id":"txn-membership-3",
		"email":"solo@example.test",
		"role":"owner"
	}`))
	requireAPIError(t, apiErr, http.StatusBadRequest, "invalid_mutation_payload", "role")
}

func TestPhase2_MembershipPatchAndAdminGuard_U_2_04(t *testing.T) {
	request, apiErr := DecodeMembershipPatchRequest(strings.NewReader(`{
		"base_membership_version":5,
		"role":"admin"
	}`))
	if apiErr != nil {
		t.Fatalf("decode valid membership patch request: %v", apiErr)
	}
	if request.BaseMembershipVersion != 5 || request.Role != "admin" {
		t.Fatalf("unexpected membership patch request: %#v", request)
	}

	_, apiErr = DecodeMembershipPatchRequest(strings.NewReader(`{
		"base_membership_version":5,
		"user_id":"00000000-0000-0000-0000-000000000001",
		"role":"admin"
	}`))
	requireAPIError(t, apiErr, http.StatusBadRequest, "invalid_mutation_payload", "user_id")

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

func TestPhase2_ExtensionDiscoveryAndDispatch_U_2_05(t *testing.T) {
	query := url.Values{"cursor_token": []string{"opaque"}}
	apiErr := auth.ValidateSingletonReadQuery(query)
	requireAPIError(t, apiErr, http.StatusBadRequest, "invalid_pagination_request", "")

	profiles := httpapi.CurrentExtensionProfiles()
	if len(profiles) != 5 {
		t.Fatalf("unexpected extension profile count: got %d", len(profiles))
	}
	if profiles[0].ProfileID != "enterprise_authentication" || profiles[len(profiles)-1].ProfileID != "snapshot_reporting" {
		t.Fatalf("unexpected extension profile ordering: %#v", profiles)
	}

	match, ok := httpapi.MatchReservedExtensionFamily("/api/v1/import-sessions")
	if !ok || match.ProfileID != "import" || match.RouteFamily != "/api/v1/import-sessions" {
		t.Fatalf("unexpected reserved import dispatch match: %#v ok=%v", match, ok)
	}
	match, ok = httpapi.MatchReservedExtensionFamily("/api/v1/users/" + uuid.NewString() + "/auth-bindings/provider")
	if !ok || match.ProfileID != "enterprise_authentication" || match.RouteFamily != "/api/v1/users/{user_id}/auth-bindings" {
		t.Fatalf("unexpected nested auth-bindings dispatch match: %#v ok=%v", match, ok)
	}
}

func TestPhase2_ListCursorBinding_U_2_06(t *testing.T) {
	actorID := uuid.MustParse("00000000-0000-0000-0000-000000000111")
	incidentID := uuid.MustParse("00000000-0000-0000-0000-000000000222")
	snapshotAt := timeRef(2026, 4, 17, 12, 0)
	updatedAt := timeRef(2026, 4, 17, 11, 59)

	token, err := EncodeIncidentCursor(incidentCursor{
		Route:       "incidents.list",
		ActorUserID: actorID.String(),
		Limit:       50,
		SnapshotAt:  snapshotAt,
		UpdatedAt:   updatedAt,
		IncidentID:  incidentID.String(),
	})
	if err != nil {
		t.Fatalf("encode incident cursor: %v", err)
	}
	if _, apiErr := DecodeIncidentCursor(token, actorID, 50); apiErr != nil {
		t.Fatalf("decode valid incident cursor: %v", apiErr)
	}
	if _, apiErr := DecodeIncidentCursor(token, actorID, 25); apiErr == nil {
		t.Fatal("cursor limit binding mismatch must fail")
	}
}

func requireAPIError(t testing.TB, apiErr *auth.APIError, wantStatus int, wantCode string, wantField string) {
	t.Helper()
	if apiErr == nil {
		t.Fatal("expected api error")
	}
	if apiErr.Status != wantStatus {
		t.Fatalf("unexpected status: got %d want %d", apiErr.Status, wantStatus)
	}
	if apiErr.Code != wantCode {
		t.Fatalf("unexpected code: got %q want %q", apiErr.Code, wantCode)
	}
	if wantField == "" {
		return
	}
	if got := apiErr.Details["field"]; got != wantField {
		t.Fatalf("unexpected field detail: got %v want %s", got, wantField)
	}
}

func timeRef(year int, month int, day int, hour int, minute int) (value time.Time) {
	return time.Date(year, time.Month(month), day, hour, minute, 0, 0, time.UTC)
}
