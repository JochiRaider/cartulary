package projectionassembly

import (
	"fmt"

	"github.com/JochiRaider/cartulary/internal/app/assessmentassembly"
	"github.com/JochiRaider/cartulary/internal/app/timelinefactassembly"
	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
	entityprovider "github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity/projectionprovider"
	entityprojection "github.com/JochiRaider/cartulary/internal/modules/entities/workbookprojection"
	indicatorowner "github.com/JochiRaider/cartulary/internal/modules/indicators"
	"github.com/JochiRaider/cartulary/internal/modules/parties"
	"github.com/JochiRaider/cartulary/internal/modules/recovery/restorecontract"
	taskdecisionprovider "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/projectionprovider"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	timelineprovider "github.com/JochiRaider/cartulary/internal/modules/timeline/projectionprovider"
	timelineprojection "github.com/JochiRaider/cartulary/internal/modules/timeline/workbookprojection"
	workbookrestoreprobe "github.com/JochiRaider/cartulary/internal/modules/workbook/restoreprobe"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/workbookprobe"
)

// Build is the sole global application constructor for the Projections
// subsystem. It constructs source-owner contributions and delegates the
// executable module graph to adapters.New.
func Build(db postgres.DB) (*Runtime, error) {
	if db == nil {
		return nil, fmt.Errorf("assemble Projections: Postgres is required")
	}
	timelineContribution, err := timelineprojection.NewContribution(timelineprovider.NewSource(timelinefactassembly.NewLinkReader()))
	if err != nil {
		return nil, fmt.Errorf("assemble Timeline projection contribution: %w", err)
	}
	entitiesContribution, err := entityprojection.NewContribution(entityprovider.NewSource())
	if err != nil {
		return nil, fmt.Errorf("assemble Entities projection contribution: %w", err)
	}
	indicatorsContribution, err := indicatorowner.NewProjectionContribution()
	if err != nil {
		return nil, fmt.Errorf("assemble Indicators projection contribution: %w", err)
	}
	assessmentsContribution, err := assessmentassembly.NewProjectionContribution()
	if err != nil {
		return nil, fmt.Errorf("assemble Assessments projection contribution: %w", err)
	}
	artifactsContribution, err := artifacts.NewProjectionContribution()
	if err != nil {
		return nil, fmt.Errorf("assemble Artifacts projection contribution: %w", err)
	}
	evidenceContribution, err := newEvidenceContribution()
	if err != nil {
		return nil, fmt.Errorf("assemble Evidence projection contribution: %w", err)
	}
	partiesContribution, err := parties.NewProjectionContribution()
	if err != nil {
		return nil, fmt.Errorf("assemble Parties projection contribution: %w", err)
	}
	taskDecisionContribution, err := taskdecisionprovider.NewContribution()
	if err != nil {
		return nil, fmt.Errorf("assemble Tasks/Decisions projection contribution: %w", err)
	}
	runtime, err := buildRuntime(
		db,
		timelineContribution,
		entitiesContribution,
		indicatorsContribution,
		assessmentsContribution,
		artifactsContribution,
		evidenceContribution,
		partiesContribution,
		taskDecisionContribution,
	)
	if err != nil {
		return nil, fmt.Errorf("assemble Projections runtime: %w", err)
	}
	return runtime, nil
}

func NewRecoveryServices(db postgres.DB) (
	restorecontract.ProjectionRebuilder,
	workbookprobe.Executor,
	error,
) {
	runtime, err := Build(db)
	if err != nil {
		return nil, nil, err
	}
	recovery := runtime.RecoveryPorts()
	if !recovery.Ready() {
		return nil, nil, fmt.Errorf("assemble Projections recovery services: incomplete recovery ports")
	}
	registry, err := workbookrestoreprobe.NewRegistry(
		runtime.RestoreProbeQuery(),
		timeline.RestoreWorkbookProbeRegistration(),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("assemble restore workbook probe registry: %w", err)
	}
	return recovery.Rebuilder, registry, nil
}
