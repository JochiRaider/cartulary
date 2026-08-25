package timeline_test

import (
	"testing"

	"github.com/JochiRaider/cartulary/internal/app/projectionassembly"
	"github.com/JochiRaider/cartulary/internal/app/timelineassembly"
	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
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
	evidenceOwner := appsupport.NewEvidenceOwnerRuntimeForTimeline(
		pool,
		conflictTokens,
		revisionComposition.Runtime.Appender(),
		revisionComposition.Publications,
		projections,
	)
	bundle, err := timelineassembly.NewBundle(timelineassembly.Dependencies{
		Postgres:            pool,
		ConflictTokens:      conflictTokens,
		Revisions:           revisionComposition.Runtime.Appender(),
		Collaboration:       revisionComposition.Publications,
		EvidenceAttachments: evidenceOwner.TimelineAttachmentContribution(),
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
