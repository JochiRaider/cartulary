package incidents

import (
	"net/http"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

func TestIncidentAccessDecisionUsesNotFoundForMissingMembershipAndDeniedForInsufficientRole_Unit(t *testing.T) {
	requireErrorContract(t, "incident_not_found", http.StatusNotFound)
	readDenied := IncidentAccessError(nil, false)
	if readDenied == nil || readDenied.Status != http.StatusNotFound || readDenied.Code != "incident_not_found" {
		t.Fatalf("missing incident membership must keep incident_not_found for read access: %#v", readDenied)
	}

	writeDenied := IncidentAccessError(nil, false, "reviewer", "admin")
	if writeDenied == nil || writeDenied.Status != http.StatusNotFound || writeDenied.Code != "incident_not_found" {
		t.Fatalf("missing incident membership must keep incident_not_found for write access: %#v", writeDenied)
	}

	membership := &MembershipRecord{Role: "viewer"}
	guard := IncidentAccessError(membership, false, "reviewer", "admin")
	if guard == nil || guard.Status != http.StatusForbidden || guard.Code != "authorization_denied" {
		t.Fatalf("insufficient incident role must keep authorization_denied: %#v", guard)
	}
}

func TestRegisterRoutesRequiresIncidentPorts(t *testing.T) {
	err := RegisterRoutes(RouteOptions{})(http.NewServeMux(), httpapi.DependencySet{})
	if err == nil || !strings.Contains(err.Error(), "collaboration session port is required") {
		t.Fatalf("missing collaboration session port must fail route registration, got %v", err)
	}
}
