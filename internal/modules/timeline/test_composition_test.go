package timeline_test

import (
	"testing"

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
	revisionComposition := revisionsupport.MustComposition(t)
	return timelineassembly.NewBundle(
		pool,
		conflictTokens,
		revisionComposition.Runtime.Appender(),
		revisionComposition.Intents,
		evidence.NewTimelineAttachmentContribution(pool),
	)
}
