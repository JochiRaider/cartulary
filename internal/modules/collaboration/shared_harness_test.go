package collaboration_test

import (
	"slices"
	"testing"

	"github.com/JochiRaider/cartulary/internal/testutil/collaborationsupport/incidentwstest"
)

func TestSocketEventInventoryCoverage(t *testing.T) {
	inventory := incidentwstest.SocketEventInventory()
	incidentwstest.RequireHarnessInventory(t, inventory)

	required := incidentwstest.RequiredHarnessIDs(inventory)
	for _, harness := range []incidentwstest.HarnessID{
		incidentwstest.HarnessEnvelopeConsistency,
		incidentwstest.HarnessAuthorizationRederived,
		incidentwstest.HarnessFieldKeyConformance,
		incidentwstest.HarnessProjectionRebuild,
		incidentwstest.HarnessWebSocketLifecycle,
		incidentwstest.HarnessGridIdentity,
		incidentwstest.HarnessTopologyAuditSource,
	} {
		if !slices.Contains(required, harness) {
			t.Fatalf("socket event inventory must require %s, got %v", harness, required)
		}
	}
}
