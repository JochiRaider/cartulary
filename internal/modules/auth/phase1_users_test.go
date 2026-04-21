package auth

import (
	"strings"
	"testing"
)

func TestPhase1_UserPatchAndLastAdminGuard_U_1_08(t *testing.T) {
	request, apiErr := DecodeUserPatchRequest(strings.NewReader(`{
		"base_user_version":2,
		"display_name":" Analyst Patched ",
		"is_deployment_admin":false
	}`))
	if apiErr != nil {
		t.Fatalf("decode valid user patch request: %v", apiErr)
	}
	if request.BaseUserVersion != 2 {
		t.Fatalf("unexpected base_user_version: got %d want 2", request.BaseUserVersion)
	}
	if request.DisplayName == nil || *request.DisplayName != "Analyst Patched" {
		t.Fatalf("unexpected normalized display_name: %#v", request.DisplayName)
	}

	_, apiErr = DecodeUserPatchRequest(strings.NewReader(`{"display_name":"Missing version"}`))
	requireAPIError(t, apiErr, 400, "invalid_mutation_payload", "base_user_version")

	if !WouldLeaveNoActiveDeploymentAdmins(
		true,
		true,
		1,
		false,
		true,
	) {
		t.Fatal("demoting the last active deployment admin must be rejected")
	}
	if !WouldLeaveNoActiveDeploymentAdmins(
		true,
		true,
		1,
		true,
		false,
	) {
		t.Fatal("deactivating the last active deployment admin must be rejected")
	}
	if WouldLeaveNoActiveDeploymentAdmins(
		true,
		true,
		2,
		false,
		true,
	) {
		t.Fatal("another active deployment admin should satisfy the guard")
	}
}
