package collaboration_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	privatestream "github.com/JochiRaider/cartulary/internal/modules/collaboration/internal/stream"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
)

func TestDurableIncidentStream_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness, admin, adminID, atomicIncidentID := setupSocketIncidentWithAdminID(
		t,
		runtime,
		"collaboration-durable-stream",
	)
	ctx := context.Background()
	closeCtx, cancelClose := context.WithTimeout(ctx, 5*time.Second)
	defer cancelClose()
	if err := harness.Collaboration.CloseDispatcher(closeCtx); err != nil {
		t.Fatalf("stop runtime collaboration dispatcher: %v", err)
	}

	pool := harness.Pool
	replay := newPostgresStreamForTest(t, pool)
	intents := privatestream.IntentWriter{}
	recovery := collaboration.NewRecoveryCapability(pool)
	atomicIncidentUUID := uuid.MustParse(atomicIncidentID)

	runIntentWriterScenarios(t, ctx, pool, harness, admin, atomicIncidentID, atomicIncidentUUID, intents)
	dispatchIncidentUUID, clockNow := runTailingScenarios(t, ctx, pool, harness, admin, intents)
	runSequencingRecoveryScenarios(t, ctx, pool, harness, admin, intents, recovery, clockNow)
	runReplayRetentionScenarios(t, ctx, pool, harness, admin, adminID, dispatchIncidentUUID, replay, intents)
}
