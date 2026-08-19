package httpapi

import (
	"net/http"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/incidents"
)

func TestUnit_MembershipPatchAndDeleteDecodeAndLastAdminGuardStayStable(t *testing.T) {
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
	if !incidents.WouldLeaveNoIncidentAdmins("admin", 1, &nextRole, false) {
		t.Fatal("demoting the last admin must be rejected")
	}
	if incidents.WouldLeaveNoIncidentAdmins("admin", 2, &nextRole, false) {
		t.Fatal("demoting one of two admins must be allowed")
	}
	if !incidents.WouldLeaveNoIncidentAdmins("admin", 1, nil, true) {
		t.Fatal("deleting the last admin must be rejected")
	}
}
