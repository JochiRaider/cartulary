package httpapi

import (
	"net/http"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
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

func TestRegisterRoutesRequiresIncidentPorts(t *testing.T) {
	t.Run("application", func(t *testing.T) {
		err := RegisterRoutes(RouteOptions{})(http.NewServeMux(), httpapi.DependencySet{})
		if err == nil || !strings.Contains(err.Error(), "application is required") {
			t.Fatalf("missing application must fail route registration, got %v", err)
		}
	})
	t.Run("terminal mutation coordinator", func(t *testing.T) {
		err := RegisterRoutes(RouteOptions{Application: &incidents.Application{}})(http.NewServeMux(), httpapi.DependencySet{})
		if err == nil || !strings.Contains(err.Error(), "terminal mutation coordinator is required") {
			t.Fatalf("missing terminal mutation coordinator must fail route registration, got %v", err)
		}
	})
}
