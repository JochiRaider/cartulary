package workbookassembly

import (
	"fmt"

	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
	"github.com/JochiRaider/cartulary/internal/modules/records"
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
	artifactOwner *artifacts.MutationFacade,
	taskDecisionOwner *tasksdecisions.MutationFacade,
) (*workbook.Store, error) {
	if contributionCatalog == nil {
		return nil, fmt.Errorf("compose Workbook mutation store: contribution catalog is required")
	}
	if artifactOwner == nil {
		return nil, fmt.Errorf("compose Workbook mutation store: Artifacts mutation contribution is required")
	}
	if taskDecisionOwner == nil {
		return nil, fmt.Errorf("compose Workbook mutation store: Tasks/Decisions mutation contribution is required")
	}
	return workbook.NewStore(workbook.StoreDependencies{
		RecordTargets:       records.NewRouteTargetResolver(pool),
		ContextualNoteOwner: artifactOwner,
		SupersedeOwner:      taskDecisionOwner,
		ContributionCatalog: contributionCatalog,
	}), nil
}
