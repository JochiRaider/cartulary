package incidents

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

func TestPhase2_U_2_08_IncidentAccessDecisionUsesNotFoundForMissingMembershipAndDeniedForInsufficientRole(t *testing.T) {
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

func TestSupportPhase2_RegisterRoutesRequiresIncidentPorts(t *testing.T) {
	err := RegisterRoutes(RouteOptions{
		CollaborationSession: noopCollaborationSessionPort{},
	})(http.NewServeMux(), httpapi.DependencySet{})
	if err == nil || !strings.Contains(err.Error(), "workbook bootstrap port is required") {
		t.Fatalf("missing workbook bootstrap port must fail route registration, got %v", err)
	}

	err = RegisterRoutes(RouteOptions{
		WorkbookBootstrap: noopWorkbookBootstrapPort{},
	})(http.NewServeMux(), httpapi.DependencySet{})
	if err == nil || !strings.Contains(err.Error(), "collaboration session port is required") {
		t.Fatalf("missing collaboration session port must fail route registration, got %v", err)
	}
}

type noopWorkbookBootstrapPort struct{}

func (noopWorkbookBootstrapPort) BootstrapIncidentCreatePreferencesTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, time.Time) error {
	return nil
}

type noopCollaborationSessionPort struct{}

func (noopCollaborationSessionPort) NotifyIncidentClosed(context.Context, uuid.UUID) {}

func (noopCollaborationSessionPort) NotifyIncidentMembershipRevoked(context.Context, uuid.UUID, uuid.UUID) {
}
