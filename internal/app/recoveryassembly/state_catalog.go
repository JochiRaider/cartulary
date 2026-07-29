package recoveryassembly

import (
	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
	"github.com/JochiRaider/cartulary/internal/modules/assessments"
	"github.com/JochiRaider/cartulary/internal/modules/auth"
	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/entities"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	extensionsrecovery "github.com/JochiRaider/cartulary/internal/modules/extensions/recoverycontribution"
	"github.com/JochiRaider/cartulary/internal/modules/graphprojection"
	"github.com/JochiRaider/cartulary/internal/modules/imports"
	incidentbundlesource "github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/networkflow"
	"github.com/JochiRaider/cartulary/internal/modules/parties"
	"github.com/JochiRaider/cartulary/internal/modules/projections"
	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	referencedatarecovery "github.com/JochiRaider/cartulary/internal/modules/reference_data/recoverycontribution"
	"github.com/JochiRaider/cartulary/internal/modules/reportcomposition"
	"github.com/JochiRaider/cartulary/internal/modules/reporting"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/savedviews"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/platform/administrativeaudit"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	recoverystate "github.com/JochiRaider/cartulary/internal/platform/recoverystate"
)

func CurrentRecoveryStateContributions() []recoverystate.Contribution {
	return []recoverystate.Contribution{
		artifacts.RecoveryStateContribution(),
		assessments.RecoveryStateContribution(),
		administrativeaudit.RecoveryStateContribution(),
		auth.RecoveryStateContribution(),
		collaboration.RecoveryStateContribution(),
		postgres.RecoveryStateContribution(),
		recovery.DeploymentAdminRecoveryStateContribution(),
		entities.RecoveryStateContribution(),
		evidence.RecoveryStateContribution(),
		extensionsrecovery.RecoveryStateContribution(),
		graphprojection.RecoveryStateContribution(),
		imports.RecoveryStateContribution(),
		incidentbundlesource.RecoveryStateContribution(),
		incidents.RecoveryStateContribution(),
		indicators.RecoveryStateContribution(),
		links.RecoveryStateContribution(),
		networkflow.RecoveryStateContribution(),
		parties.RecoveryStateContribution(),
		jobs.RecoveryStateContribution(),
		projections.RecoveryStateContribution(),
		recovery.RecoveryStateContribution(),
		referencedatarecovery.RecoveryStateContribution(),
		reportcomposition.RecoveryStateContribution(),
		reporting.RecoveryStateContribution(),
		revisions.RecoveryStateContribution(),
		savedviews.RecoveryStateContribution(),
		tasksdecisions.RecoveryStateContribution(),
		timeline.RecoveryStateContribution(),
	}
}

func CurrentRecoveryStateCatalog() (*recoverystate.Catalog, error) {
	return recoverystate.Build(CurrentRecoveryStateContributions()...)
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
		incidentbundlesource.VNextRecoveryObjectInventory(source),
		referencedatarecovery.VNextRecoveryObjectInventory(source),
		reporting.VNextRecoveryObjectInventory(source),
	)
}
