package admission

import (
	"context"
	"errors"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestRoleSetsAreExactAndNonOrdinal(t *testing.T) {
	testCases := []struct {
		name    string
		roles   RoleSet
		allowed []Role
	}{
		{name: "member", roles: RolesMember, allowed: []Role{RoleViewer, RoleEditor, RoleReviewer, RoleAdmin}},
		{name: "editor reviewer admin", roles: RolesEditorReviewerAdmin, allowed: []Role{RoleEditor, RoleReviewer, RoleAdmin}},
		{name: "reviewer admin", roles: RolesReviewerAdmin, allowed: []Role{RoleReviewer, RoleAdmin}},
		{name: "editor admin", roles: RolesEditorAdmin, allowed: []Role{RoleEditor, RoleAdmin}},
		{name: "admin", roles: RolesAdmin, allowed: []Role{RoleAdmin}},
	}
	for _, testCase := range testCases {
		t.Run(testCase.name, func(t *testing.T) {
			for _, role := range []Role{RoleViewer, RoleEditor, RoleReviewer, RoleAdmin} {
				want := false
				for _, allowed := range testCase.allowed {
					want = want || role == allowed
				}
				if got := testCase.roles.includes(role); got != want {
					t.Fatalf("role %d included = %t, want %t", role, got, want)
				}
			}
		})
	}
}

func TestInvalidRequirementsAndStoredEnumsFailAsInternalErrors(t *testing.T) {
	for _, requirement := range []Requirement{
		{},
		{AllowedRoles: RoleSet(1 << 7), Lifecycle: LifecycleAny},
		{AllowedRoles: RolesMember},
		{AllowedRoles: RolesMember, Lifecycle: Lifecycle(99)},
	} {
		if err := validateRequirement(requirement); err == nil {
			t.Fatalf("invalid requirement accepted: %#v", requirement)
		} else if IsDenied(err, DenialNotVisible) || IsDenied(err, DenialInsufficientRole) || IsDenied(err, DenialIncidentClosed) {
			t.Fatalf("invalid requirement became policy denial: %v", err)
		}
	}
	if _, err := decide("owner", "active", Requirement{AllowedRoles: RolesMember, Lifecycle: LifecycleAny}); err == nil {
		t.Fatal("malformed stored role was accepted")
	}
	if _, err := decide("admin", "archived", Requirement{AllowedRoles: RolesMember, Lifecycle: LifecycleAny}); err == nil {
		t.Fatal("malformed stored status was accepted")
	}
}

func TestNilDependenciesFailAsInternalErrors(t *testing.T) {
	var nilPool *pgxpool.Pool
	var typedNil postgres.DB = nilPool
	requirement := Requirement{AllowedRoles: RolesMember, Lifecycle: LifecycleAny}
	for _, checker := range []*Checker{nil, NewChecker(nil), NewChecker(typedNil)} {
		if _, err := checker.Check(context.Background(), uuid.New(), uuid.New(), requirement); err == nil {
			t.Fatal("nil admission dependency was accepted")
		} else if IsDenied(err, DenialNotVisible) || IsDenied(err, DenialInsufficientRole) || IsDenied(err, DenialIncidentClosed) {
			t.Fatalf("nil admission dependency became a policy denial: %v", err)
		}
	}
}

func TestDecisionPrecedenceIsRoleBeforeLifecycleAfterVisibility(t *testing.T) {
	requirement := Requirement{AllowedRoles: RolesAdmin, Lifecycle: LifecycleOpen}
	_, err := decide("viewer", "closed", requirement)
	if !IsDenied(err, DenialInsufficientRole) {
		t.Fatalf("closed viewer denial = %v, want insufficient role", err)
	}
	_, err = decide("admin", "closed", requirement)
	if !IsDenied(err, DenialIncidentClosed) {
		t.Fatalf("closed admin denial = %v, want incident closed", err)
	}
	if !errors.As(&Denied{Code: DenialNotVisible}, new(*Denied)) {
		t.Fatal("typed denial must remain discoverable")
	}
}
