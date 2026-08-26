package recoveryassembly

import (
	"fmt"

	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
	"github.com/JochiRaider/cartulary/internal/modules/assessments"
	"github.com/JochiRaider/cartulary/internal/modules/auth"
	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	database_migrations "github.com/JochiRaider/cartulary/internal/modules/database_migrations"
	"github.com/JochiRaider/cartulary/internal/modules/entities"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	extensionsrecovery "github.com/JochiRaider/cartulary/internal/modules/extensions/recoverycontribution"
	"github.com/JochiRaider/cartulary/internal/modules/graphprojection"
	"github.com/JochiRaider/cartulary/internal/modules/imports"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles"
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/networkflow"
	"github.com/JochiRaider/cartulary/internal/modules/parties"
	"github.com/JochiRaider/cartulary/internal/modules/projections/providercontract"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	referencedatarecovery "github.com/JochiRaider/cartulary/internal/modules/reference_data/recoverycontribution"
	"github.com/JochiRaider/cartulary/internal/modules/reportcomposition"
	"github.com/JochiRaider/cartulary/internal/modules/reporting"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/savedviews"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/modules/workbook"
	"github.com/JochiRaider/cartulary/internal/platform/administrativeaudit"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	recoverystate "github.com/JochiRaider/cartulary/internal/platform/recoverystate"
)

func CurrentRecoveryStateContributions() ([]recoverystate.Contribution, error) {
	artifactsContribution, err := artifacts.RecoveryStateContribution()
	if err != nil {
		return nil, fmt.Errorf("recovery assembly: Artifacts state contribution: %w", err)
	}
	indicatorsContribution, err := indicators.RecoveryStateContribution()
	if err != nil {
		return nil, fmt.Errorf("recovery assembly: Indicators state contribution: %w", err)
	}
	return []recoverystate.Contribution{
		artifactsContribution,
		assessments.RecoveryStateContribution(),
		administrativeaudit.RecoveryStateContribution(),
		auth.RecoveryStateContribution(),
		collaboration.RecoveryStateContribution(),
		database_migrations.RecoveryStateContribution(),
		recovery.DeploymentAdminRecoveryStateContribution(),
		entities.RecoveryStateContribution(),
		evidence.RecoveryStateContribution(),
		extensionsrecovery.RecoveryStateContribution(),
		graphprojection.RecoveryStateContribution(),
		imports.RecoveryStateContribution(),
		incidentbundles.RecoveryStateContribution(),
		incidents.RecoveryStateContribution(),
		indicatorsContribution,
		links.RecoveryStateContribution(),
		networkflow.RecoveryStateContribution(),
		parties.RecoveryStateContribution(),
		jobs.RecoveryStateContribution(),
		providercontract.RecoveryStateContribution(),
		records.RecoveryStateContribution(),
		recovery.RecoveryStateContribution(),
		referencedatarecovery.RecoveryStateContribution(),
		reportcomposition.RecoveryStateContribution(),
		reporting.RecoveryStateContribution(),
		revisions.RecoveryStateContribution(),
		savedviews.RecoveryStateContribution(),
		tasksdecisions.RecoveryStateContribution(),
		timeline.RecoveryStateContribution(),
		workbook.RecoveryStateContribution(),
	}, nil
}

func CurrentRecoveryStateCatalog() (*recoverystate.Catalog, error) {
	contributions, err := CurrentRecoveryStateContributions()
	if err != nil {
		return nil, err
	}
	return recoverystate.Build(contributions...)
}

func CurrentVNextObjectInventoryCatalog(
	source recovery.VNextObjectSource,
) (*recovery.VNextObjectInventoryCatalog, error) {
	stateCatalog, err := CurrentRecoveryStateCatalog()
	if err != nil {
		return nil, err
	}
	return recovery.NewVNextObjectInventoryCatalog(
		stateCatalog,
		evidence.VNextRecoveryObjectInventory(source),
		imports.VNextRecoveryObjectInventory(),
		extensionsrecovery.VNextRecoveryObjectInventory(source),
		incidentbundles.VNextRecoveryObjectInventory(source),
		referencedatarecovery.VNextRecoveryObjectInventory(source),
		reporting.VNextRecoveryObjectInventory(source),
	)
}
