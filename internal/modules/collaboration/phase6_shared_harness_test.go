package collaboration_test

import (
	"slices"
	"testing"

	"github.com/JochiRaider/cartulary/internal/testutil/phase6test"
)

func TestSupportPhase6SharedHarness_SocketEventInventoryCoverage(t *testing.T) {
	inventory := phase6test.Phase6SocketEventInventory()
	phase6test.RequireSharedHarnessInventory(t, inventory)

	required := phase6test.RequiredHarnessIDs(inventory)
	for _, harness := range []phase6test.SharedHarnessID{
		phase6test.HarnessEnvelopeConsistency,
		phase6test.HarnessAuthorizationRederived,
		phase6test.HarnessFieldKeyConformance,
		phase6test.HarnessProjectionRebuild,
		phase6test.HarnessWebSocketLifecycle,
		phase6test.HarnessGridIdentity,
		phase6test.HarnessTopologyAuditSource,
	} {
		if !slices.Contains(required, harness) {
			t.Fatalf("socket event inventory must require %s, got %v", harness, required)
		}
	}
}
