package httpapi

import (
	"net/http"
	"strings"
	"testing"
)

func TestUnit_MembershipPatchAndDeleteDecodeStayStable(t *testing.T) {
	patchRequest, apiErr := decodeMembershipPatchRequest(strings.NewReader(`{
		"base_membership_version":5,
		"role":"admin"
	}`))
	if apiErr != nil {
		t.Fatalf("decode valid membership patch request: %v", apiErr)
	}
	if patchRequest.BaseMembershipVersion != 5 || patchRequest.Role != "admin" {
		t.Fatalf("unexpected membership patch request: %#v", patchRequest)
	}

	deleteRequest, apiErr := decodeMembershipDeleteRequest(strings.NewReader(`{
		"base_membership_version":5
	}`))
	if apiErr != nil {
		t.Fatalf("decode valid membership delete request: %v", apiErr)
	}
	if deleteRequest.BaseMembershipVersion != 5 {
		t.Fatalf("unexpected membership delete request: %#v", deleteRequest)
	}

	_, apiErr = decodeMembershipDeleteRequest(strings.NewReader(`{}`))
	requireAPIError(t, apiErr, http.StatusBadRequest, "invalid_mutation_payload", "base_membership_version", "missing_required_field")
}
