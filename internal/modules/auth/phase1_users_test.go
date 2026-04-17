package auth

import (
	"testing"
	"time"

	"github.com/google/uuid"

	"example.com/todo/cartulary/internal/platform/authn"
)

func TestPhase1_UserCreateDefaultsAndSafeShape_U_1_07(t *testing.T) {
	defaults := ApplyUserCreateDefaults(nil, nil)
	if !defaults.MFARequired {
		t.Fatal("omitted mfa_required must default to true")
	}
	if defaults.IsDeploymentAdmin {
		t.Fatal("omitted is_deployment_admin must default to false")
	}

	userID := uuid.MustParse("30000000-0000-0000-0000-000000000001")
	createdAt := time.Date(2026, time.April, 17, 15, 0, 0, 0, time.UTC)
	resource := BuildSafeUserResource(authn.UserRecord{
		ID:                userID,
		Email:             "analyst@example.test",
		DisplayName:       "Analyst",
		IsActive:          true,
		MFARequired:       true,
		IsDeploymentAdmin: false,
		CreatedAt:         createdAt,
		UpdatedAt:         createdAt,
		UserVersion:       1,
	})

	if got := resource["email"]; got != "analyst@example.test" {
		t.Fatalf("unexpected email in safe user resource: got %v", got)
	}
	if got := resource["is_active"]; got != true {
		t.Fatalf("unexpected is_active in safe user resource: got %v", got)
	}
	authBindings, ok := resource["auth_bindings"].([]map[string]any)
	if !ok || len(authBindings) != 1 {
		t.Fatalf("expected one local auth binding summary, got %#v", resource["auth_bindings"])
	}
	if got := authBindings[0]["provider_type"]; got != "local" {
		t.Fatalf("unexpected local binding provider_type: got %v", got)
	}
	if _, ok := resource["password_hash"]; ok {
		t.Fatal("safe user resource must not expose password_hash")
	}
}

func TestPhase1_UserPatchAndLastAdminGuard_U_1_08(t *testing.T) {
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

func TestPhase1_AdminCredentialActionsRequireDeploymentAdmin_U_1_13(t *testing.T) {
	if apiErr := RequireDeploymentAdmin(authn.UserRecord{IsDeploymentAdmin: true}); apiErr != nil {
		t.Fatalf("deployment admin should pass admin-action guard: %v", apiErr)
	}
	apiErr := RequireDeploymentAdmin(authn.UserRecord{IsDeploymentAdmin: false})
	if apiErr == nil {
		t.Fatal("non-deployment-admin must fail admin-action guard")
	}
	if apiErr.Status != 401 || apiErr.Code != unauthorizedCode {
		t.Fatalf("unexpected admin-action guard error: %#v", apiErr)
	}
}
