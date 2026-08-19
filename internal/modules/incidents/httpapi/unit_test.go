package httpapi

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

func TestIncidentAccessDecisionUsesNotFoundForMissingMembershipAndDeniedForInsufficientRole_Unit(t *testing.T) {
	requireErrorContract(t, "incident_not_found", http.StatusNotFound)
	_, readDenied := incidentAdmissionResult(admission.Grant{}, &admission.Denied{Code: admission.DenialNotVisible}, "")
	if readDenied == nil || readDenied.Status != http.StatusNotFound || readDenied.Code != "incident_not_found" {
		t.Fatalf("missing incident membership must keep incident_not_found for read access: %#v", readDenied)
	}

	_, writeDenied := incidentAdmissionResult(admission.Grant{}, &admission.Denied{Code: admission.DenialNotVisible}, "reviewer|admin")
	if writeDenied == nil || writeDenied.Status != http.StatusNotFound || writeDenied.Code != "incident_not_found" {
		t.Fatalf("missing incident membership must keep incident_not_found for write access: %#v", writeDenied)
	}

	_, guard := incidentAdmissionResult(admission.Grant{}, &admission.Denied{Code: admission.DenialInsufficientRole}, "reviewer|admin")
	if guard == nil || guard.Status != http.StatusForbidden || guard.Code != "authorization_denied" {
		t.Fatalf("insufficient incident role must keep authorization_denied: %#v", guard)
	}
}

func TestRegisterRoutesRequiresExactDependencies(t *testing.T) {
	validApplication := &incidents.Application{}
	validAdmission := &admission.Checker{}
	var typedNilApplication *incidents.Application
	var typedNilAdmission *admission.Checker
	var typedNilCoordinator *terminalMutationCoordinatorStub

	tests := []struct {
		name         string
		dependencies Dependencies
		wantError    string
	}{
		{name: "missing application", dependencies: Dependencies{}, wantError: "application is required"},
		{name: "typed nil application", dependencies: Dependencies{Application: typedNilApplication}, wantError: "application is required"},
		{name: "missing admission", dependencies: Dependencies{Application: validApplication}, wantError: "admission checker is required"},
		{name: "typed nil admission", dependencies: Dependencies{Application: validApplication, AdmissionChecker: typedNilAdmission}, wantError: "admission checker is required"},
		{name: "missing coordinator", dependencies: Dependencies{Application: validApplication, AdmissionChecker: validAdmission}, wantError: "terminal mutation coordinator is required"},
		{name: "typed nil coordinator", dependencies: Dependencies{Application: validApplication, AdmissionChecker: validAdmission, TerminalMutationCoordinator: typedNilCoordinator}, wantError: "terminal mutation coordinator is required"},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			err := RegisterRoutes(test.dependencies)(http.NewServeMux(), httpapi.DependencySet{})
			if err == nil || !strings.Contains(err.Error(), test.wantError) {
				t.Fatalf("route registration error = %v, want %q", err, test.wantError)
			}
		})
	}
}

type terminalMutationCoordinatorStub struct{}

func (*terminalMutationCoordinatorStub) CoordinateIncidentLifecycle(
	context.Context,
	authn.UserRecord,
	uuid.UUID,
	string,
	incidents.IncidentLifecycleRequest,
	string,
	time.Time,
) (incidents.IncidentLifecycleResult, error) {
	return incidents.IncidentLifecycleResult{}, nil
}

func (*terminalMutationCoordinatorStub) CoordinateMembershipDeletion(
	context.Context,
	authn.UserRecord,
	uuid.UUID,
	uuid.UUID,
	incidents.MembershipDeleteRequest,
	string,
) (incidents.MembershipDeleteResult, error) {
	return incidents.MembershipDeleteResult{}, nil
}
