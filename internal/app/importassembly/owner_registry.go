package importassembly

import (
	"fmt"

	"github.com/JochiRaider/cartulary/internal/gen/importtargetregistry"
	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/imports"
	"github.com/JochiRaider/cartulary/internal/modules/imports/ownerfacade"
	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	"github.com/JochiRaider/cartulary/internal/modules/parties"
	"github.com/JochiRaider/cartulary/internal/modules/projections"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type OwnerRegistryDependencies struct {
	Postgres          postgres.DB
	RevisionAppender  *revisions.Appender
	Intents           collaboration.IntentAppender
	Timeline          *timeline.Facade
	ProjectionCatalog *projections.Catalog
}

func NewOwnerCreateRegistry(
	dependencies OwnerRegistryDependencies,
) (*ownerfacade.ImportOwnerCreateRegistry, error) {
	if dependencies.Postgres == nil {
		return nil, fmt.Errorf("compose import owner-create registry: Postgres is required")
	}
	if dependencies.RevisionAppender == nil {
		return nil, fmt.Errorf(
			"compose import owner-create registry: Revisions appender is required",
		)
	}
	if dependencies.Intents == nil {
		return nil, fmt.Errorf(
			"compose import owner-create registry: Collaboration intents are required",
		)
	}
	if dependencies.Timeline == nil {
		return nil, fmt.Errorf(
			"compose import owner-create registry: Timeline owner is required",
		)
	}
	if dependencies.ProjectionCatalog == nil {
		return nil, fmt.Errorf(
			"compose import owner-create registry: projection catalog is required",
		)
	}

	facades := make([]ownerfacade.ImportOwnerCreateFacade, 0)
	for _, target := range importtargetregistry.Targets {
		if target.TargetKind != imports.ImportTargetKindViewSchema ||
			target.AvailabilityKind != "enabled" {
			continue
		}
		if target.TargetViewSchemaID == nil || target.FacadeID == nil {
			return nil, fmt.Errorf(
				"compose import owner-create registry: generated target %s has no binding",
				target.TargetID,
			)
		}
		facade, err := newOwnerCreateFacade(
			target.OwnerContractRef,
			*target.TargetViewSchemaID,
			*target.FacadeID,
			dependencies,
		)
		if err != nil {
			return nil, fmt.Errorf(
				"compose import owner-create registry target %s: %w",
				target.TargetID,
				err,
			)
		}
		facades = append(facades, facade)
	}
	registry, err := imports.NewOwnerCreateRegistry(facades...)
	if err != nil {
		return nil, fmt.Errorf("compose import owner-create registry: %w", err)
	}
	return registry, nil
}

func newOwnerCreateFacade(
	ownerContractRef string,
	targetViewSchemaID string,
	facadeID string,
	dependencies OwnerRegistryDependencies,
) (ownerfacade.ImportOwnerCreateFacade, error) {
	switch ownerContractRef {
	case "module.artifacts@1":
		return artifacts.NewImportCreateFacade(
			targetViewSchemaID,
			facadeID,
			dependencies.RevisionAppender,
		)
	case "module.assessments@1":
		return newAssessmentImportCreateFacade(
			targetViewSchemaID,
			facadeID,
			dependencies.Postgres,
			dependencies.ProjectionCatalog,
			dependencies.RevisionAppender,
		)
	case "module.entities@1":
		return hostidentity.NewImportCreateFacade(
			targetViewSchemaID,
			facadeID,
			dependencies.Postgres,
			dependencies.RevisionAppender,
		)
	case "module.evidence@1":
		return evidence.NewImportCreateFacade(
			targetViewSchemaID,
			facadeID,
			dependencies.Postgres,
			dependencies.RevisionAppender,
			dependencies.Intents,
		)
	case "module.indicators@1":
		return indicators.NewImportCreateFacade(
			targetViewSchemaID,
			facadeID,
			dependencies.Postgres,
			dependencies.RevisionAppender,
		)
	case "module.parties@1":
		return parties.NewImportCreateFacade(
			targetViewSchemaID,
			facadeID,
			dependencies.Postgres,
			dependencies.RevisionAppender,
		)
	case "module.tasksdecisions@1":
		return tasksdecisions.NewImportContribution(
			targetViewSchemaID,
			facadeID,
			dependencies.RevisionAppender,
		)
	case "module.timeline@2":
		return timeline.NewImportCreateFacade(
			targetViewSchemaID,
			facadeID,
			dependencies.Timeline,
		)
	default:
		return nil, fmt.Errorf("unsupported source owner %s", ownerContractRef)
	}
}
