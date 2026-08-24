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
	Timeline                ownerfacade.ImportOwnerCreateTx
	EntityProjections       entityprojection.Writer
	AssessmentProjections   assessmentprojection.Rows
	ArtifactProjections     artifactprojection.Rows
	Evidence                ownerfacade.ImportOwnerCreateFacade
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
	if dependencies.Evidence == nil {
		return nil, fmt.Errorf(
			"compose import owner-create registry: Evidence owner facade is required",
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
	artifactImportDependencies := artifacts.ImportDependencies{
		RecordEnvelopes: artifactRecordInserter{store: records.NewStore()},
		ActiveUsers:     artifactActiveUserLookup{},
		Projections:     artifactProjectionAdapter{rows: dependencies.ArtifactProjections},
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
			artifactImportDependencies,
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
	artifactImportDependencies artifacts.ImportDependencies,
	taskDecisionImportDependencies tasksdecisions.ImportDependencies,
) (ownerfacade.ImportOwnerCreateFacade, error) {
	switch ownerContractRef {
	case "module.artifacts@1":
		return artifacts.NewImportContribution(
			targetViewSchemaID,
			facadeID,
			artifactImportDependencies,
		)
	case "module.assessments@1":
		return newAssessmentImportCreateFacade(
			targetViewSchemaID,
			facadeID,
			dependencies.Postgres,
			dependencies.AssessmentProjections,
			dependencies.RevisionAppender,
		)
	case "module.entities@1":
		return hostidentity.NewImportCreateFacade(
			targetViewSchemaID,
			facadeID,
			hostidentity.ImportDependencies{
				Revisions:        dependencies.RevisionAppender,
				ProjectionWriter: dependencies.EntityProjections,
			},
		)
	case "module.evidence@1":
		binding := dependencies.Evidence.ImportOwnerCreateBinding()
		if binding.TargetViewSchemaID != targetViewSchemaID || binding.FacadeID != facadeID {
			return nil, fmt.Errorf(
				"evidence owner facade binding = %s/%s, want %s/%s",
				binding.TargetViewSchemaID,
				binding.FacadeID,
				targetViewSchemaID,
				facadeID,
			)
		}
		return dependencies.Evidence, nil
	case "module.indicators@1":
		return indicators.NewImportCreateFacade(
			targetViewSchemaID,
			facadeID,
			dependencies.Indicators,
		)
	case "module.parties@1":
		return parties.NewImportContribution(
			targetViewSchemaID,
			facadeID,
			parties.ImportDependencies{
				RecordEnvelopes: records.NewStore(),
				Projections:     dependencies.PartyProjections,
				Revisions:       dependencies.RevisionAppender,
			},
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
