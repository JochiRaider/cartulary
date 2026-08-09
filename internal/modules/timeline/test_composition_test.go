package timeline_test

import (
	"testing"

	"github.com/JochiRaider/cartulary/internal/app/projectionassembly"
	"github.com/JochiRaider/cartulary/internal/app/timelineassembly"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/revisionsupport"
)

func newTestTimelineBundle(
	t testing.TB,
	pool postgres.DB,
	conflictTokens conflicttokens.ConflictTokenCodec,
) *timelineassembly.Bundle {
	t.Helper()
	bundle, _ := newTestTimelineComposition(t, pool, conflictTokens)
	return bundle
}

func newTestTimelineComposition(
	t testing.TB,
	pool postgres.DB,
	conflictTokens conflicttokens.ConflictTokenCodec,
) (*timelineassembly.Bundle, *projectionassembly.Runtime) {
	t.Helper()
	revisionComposition := revisionsupport.MustComposition(t)
	projections := mustBuildProjectionRuntime(t, pool)
	bundle, err := timelineassembly.NewBundle(timelineassembly.Dependencies{
		Postgres:            pool,
		ConflictTokens:      conflictTokens,
		Revisions:           revisionComposition.Runtime.Appender(),
		Collaboration:       revisionComposition.Intents,
		EvidenceAttachments: evidence.NewTimelineAttachmentContribution(pool),
		TimelineProjection:  projections.TimelinePorts().Writer,
		EntityProjection:    projections.EntityPorts().Writer,
		AssessmentRows:      projections.AssessmentPorts().Rows,
	})
	if err != nil {
		t.Fatalf("compose Timeline test bundle: %v", err)
	}
	return bundle, projections
}

func mustBuildProjectionRuntime(t testing.TB, pool postgres.DB) *projectionassembly.Runtime {
	t.Helper()
	runtime, err := projectionassembly.Build(pool)
	if err != nil {
		t.Fatalf("compose projection runtime: %v", err)
	}
	return runtime
}
