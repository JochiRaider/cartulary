package importassembly

import (
	"fmt"

	"github.com/JochiRaider/cartulary/internal/gen/importtargetregistry"
	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
	artifactprojection "github.com/JochiRaider/cartulary/internal/modules/artifacts/workbookprojection"
	assessmentprojection "github.com/JochiRaider/cartulary/internal/modules/assessments/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity"
	entityprojection "github.com/JochiRaider/cartulary/internal/modules/entities/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	evidenceprojection "github.com/JochiRaider/cartulary/internal/modules/evidence/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/imports"
	"github.com/JochiRaider/cartulary/internal/modules/imports/ownerfacade"
	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/parties"
	partyprojection "github.com/JochiRaider/cartulary/internal/modules/parties/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions"
	taskdecisionprojection "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type OwnerRegistryDependencies struct {
	Postgres                postgres.DB
	RevisionAppender        *revisions.Appender
	Intents                 collaboration.IntentAppender
	Timeline                *timeline.Facade
	EntityProjections       entityprojection.Writer
	AssessmentProjections   assessmentprojection.Rows
	ArtifactProjections     artifactprojection.Rows
	EvidenceProjections     evidenceprojection.Rows
	PartyProjections        partyprojection.Rows
	TaskDecisionProjections taskdecisionprojection.Rows
	Indicators              *indicators.Store
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
	if dependencies.EntityProjections == nil {
		return nil, fmt.Errorf(
			"compose import owner-create registry: Entities projection writer is required",
		)
	}
	if dependencies.AssessmentProjections == nil {
		return nil, fmt.Errorf(
			"compose import owner-create registry: Assessments projection rows are required",
		)
	}
	if dependencies.ArtifactProjections == nil {
		return nil, fmt.Errorf(
			"compose import owner-create registry: Artifacts projection rows are required",
		)
	}
	if dependencies.EvidenceProjections == nil {
		return nil, fmt.Errorf(
			"compose import owner-create registry: Evidence projection rows are required",
		)
	}
	if dependencies.PartyProjections == nil {
		return nil, fmt.Errorf(
			"compose import owner-create registry: Parties projection rows are required",
		)
	}
	if dependencies.TaskDecisionProjections == nil {
		return nil, fmt.Errorf(
			"compose import owner-create registry: Tasks/Decisions projection rows are required",
		)
	}
	if dependencies.Indicators == nil {
		return nil, fmt.Errorf(
			"compose import owner-create registry: Indicators owner is required",
		)
	}

	taskDecisionImportDependencies := tasksdecisions.ImportDependencies{
		RecordEnvelopes: records.NewStore(),
		Links:           links.NewStore(),
		Projections:     dependencies.TaskDecisionProjections,
		Revisions:       dependencies.RevisionAppender,
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
			taskDecisionImportDependencies,
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
	taskDecisionImportDependencies tasksdecisions.ImportDependencies,
) (ownerfacade.ImportOwnerCreateFacade, error) {
	switch ownerContractRef {
	case "module.artifacts@1":
		return artifacts.NewImportCreateFacade(
			targetViewSchemaID,
			facadeID,
			dependencies.RevisionAppender,
			dependencies.ArtifactProjections,
		)
	case "module.assessments@1":
		return newAssessmentImportCreateFacade(
			targetViewSchemaID,
			facadeID,
			dependencies.Postgres,
			dependencies.AssessmentProjections,
			dependencies.RevisionAppender,
			dependencies.EntityProjections,
		)
	case "module.entities@1":
		return hostidentity.NewImportCreateFacade(
			targetViewSchemaID,
			facadeID,
			dependencies.Postgres,
			dependencies.RevisionAppender,
			dependencies.EntityProjections,
		)
	case "module.evidence@1":
		return evidence.NewImportCreateFacade(
			targetViewSchemaID,
			facadeID,
			dependencies.Postgres,
			dependencies.RevisionAppender,
			dependencies.Intents,
			dependencies.EvidenceProjections,
		)
	case "module.indicators@1":
		return indicators.NewImportCreateFacade(
			targetViewSchemaID,
			facadeID,
			dependencies.Indicators,
		)
	case "module.parties@1":
		return parties.NewImportCreateFacade(
			targetViewSchemaID,
			facadeID,
			dependencies.Postgres,
			dependencies.RevisionAppender,
			dependencies.PartyProjections,
		)
	case "module.tasksdecisions@1":
		return tasksdecisions.NewImportContribution(
			targetViewSchemaID,
			facadeID,
			taskDecisionImportDependencies,
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
