package workbookassembly

import (
	"github.com/JochiRaider/cartulary/internal/modules/artifacts/linkednotes"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions"
	"github.com/JochiRaider/cartulary/internal/modules/workbook"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

// NewMutationStore binds Workbook route coordination to dedicated owner
// facades. Workbook receives only the operations it coordinates and never the
// database capability used to construct them.
func NewMutationStore(
	pool postgres.DB,
	contributionCatalog *workbook.WorkbookContributionCatalog,
	appender *revisions.Appender,
) *workbook.Store {
	return workbook.NewStore(workbook.StoreDependencies{
		RecordTargets:       records.NewRouteTargetResolver(pool),
		LinkedNoteOwner:     linkednotes.NewFacade(pool, appender),
		SupersedeOwner:      tasksdecisions.NewSupersedeFacade(pool, appender),
		ContributionCatalog: contributionCatalog,
	})
}
