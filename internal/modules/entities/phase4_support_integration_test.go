package entities_test

import (
	"net/http"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/golden"
	"github.com/JochiRaider/cartulary/internal/testutil/phase4test"
)

func TestSupportPhase4Integration_RouteSurfaceInventory(t *testing.T) {
	harness := phase4test.StartServer(t, "phase4-support-route-surface")
	adminLogin, adminUserID := phase4test.ProvisionBootstrapAdmin(t, harness.Server)
	incident := phase4test.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-phase4-support-route-incident",
		"incident_key":  "IR-P4-SUPPORT",
		"title":         "Phase 4 support route surface inventory",
	})
	incidentID := phase4test.MustUUID(t, incident["incident_id"].(string))

	phase4test.SeedTimelineRecord(t, harness.DB, incidentID, adminUserID, golden.Phase4TimelineRecordID)
	phase4test.SeedHostRecord(t, harness.DB, incidentID, adminUserID, golden.Phase4CanonicalHostRecordID, "Gateway Support", "GATEWAY-SUPPORT", "", "")
	phase4test.SeedHostRecord(t, harness.DB, incidentID, adminUserID, golden.Phase4DuplicateHostRecordID, "Gateway Support Duplicate", "GATEWAY-SUPPORT-DUP", "", "")
	phase4test.SeedMention(t, harness.DB, adminUserID, golden.Phase4HostMentionID, golden.Phase4TimelineRecordID, golden.Phase4FieldTimelineHostRefs, "host", "GATEWAY-SUPPORT", "unresolved", nil, nil)

	requestOptions := []func(*http.Request){
		phase4test.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		phase4test.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	}
	for _, route := range phase4test.OwnedRouteSurfaceInventory(
		incidentID.String(),
		golden.Phase4TimelineRecordID.String(),
		golden.Phase4HostMentionID.String(),
		golden.Phase4CanonicalHostRecordID.String(),
	) {
		t.Run(route.Name, func(t *testing.T) {
			resp := phase4test.RequireRouteSurface(
				t,
				"support-route-surface",
				harness.Server,
				route.Method,
				route.Path,
				route.Body,
				requestOptions...,
			)
			_ = resp.Body.Close()
		})
	}
}
