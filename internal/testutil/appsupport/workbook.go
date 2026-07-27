package appsupport

import (
	"github.com/JochiRaider/cartulary/internal/app/timelineassembly"
	"github.com/JochiRaider/cartulary/internal/app/workbookassembly"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/conflicttokens"
	"github.com/JochiRaider/cartulary/internal/modules/workbook"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

// NewWorkbookStore composes the same code-backed projection catalog used by
// the server for focused module tests that do not need an HTTP runtime.
func NewWorkbookStore(pool postgres.DB, conflictTokens conflicttokens.ConflictTokenCodec) *workbook.Store {
	timelineBundle := timelineassembly.NewBundle(pool, conflictTokens)
	catalog, err := workbookassembly.NewContributionCatalog(
		pool,
		timelineBundle.ProjectionCatalog.Catalog,
		timelineBundle.ProjectionCatalog.Query,
		timelineBundle.Facade,
		conflictTokens,
	)
	if err != nil {
		panic(err)
	}
	return workbookassembly.NewMutationStore(pool, catalog)
}
