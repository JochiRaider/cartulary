package httpapi

import (
	"net/http"
	"strings"
	"testing"
)

func TestUnit_MembershipPatchAndDeleteDecodeStayStable(t *testing.T) {
	_, apiErr := admitMembershipPatchJSON(strings.NewReader(`{
		"base_membership_version":5,
		"role":"admin"
	}`))
	if apiErr != nil {
		t.Fatalf("decode valid membership patch request: %v", apiErr)
	}
	_, apiErr = admitMembershipDeleteJSON(strings.NewReader(`{
		"base_membership_version":5
	}`))
	if apiErr != nil {
		t.Fatalf("decode valid membership delete request: %v", apiErr)
	}
	_, apiErr = admitMembershipDeleteJSON(strings.NewReader(`{}`))
	requireAPIError(t, apiErr, http.StatusBadRequest, "invalid_mutation_payload", "base_membership_version", "missing_required_field")
}
