package runtime

import (
	"fmt"

	artifactprojection "github.com/JochiRaider/cartulary/internal/modules/artifacts/workbookprojection"
	assessmentprojection "github.com/JochiRaider/cartulary/internal/modules/assessments/workbookprojection"
	entityprojection "github.com/JochiRaider/cartulary/internal/modules/entities/workbookprojection"
	evidenceprojection "github.com/JochiRaider/cartulary/internal/modules/evidence/workbookprojection"
	indicatorprojection "github.com/JochiRaider/cartulary/internal/modules/indicators/workbookprojection"
	partyprojection "github.com/JochiRaider/cartulary/internal/modules/parties/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/projections/providercontract"
	taskdecisionprojection "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/workbookprojection"
)

type OperationalDependencies struct {
	Timeline     TimelineSource
	Entities     entityprojection.SourceReader
	Indicators   indicatorprojection.SourceReader
	Assessments  assessmentprojection.SourceReader
	Artifacts    artifactprojection.SourceReader
	Evidence     evidenceprojection.SourceReader
	Parties      partyprojection.SourceReader
	TaskRequests taskdecisionprojection.TaskRequestSourceReader
	Decisions    taskdecisionprojection.DecisionSourceReader
}

func NewOperationalCatalog(
	descriptors providercontract.DescriptorSet,
	dependencies OperationalDependencies,
) (*Catalog, error) {
	if descriptors.Len() != len(requiredProviderOwners) {
		return nil, fmt.Errorf("operational projection descriptor set has %d providers, want %d", descriptors.Len(), len(requiredProviderOwners))
	}
	for name, source := range map[string]any{
		"Timeline":      dependencies.Timeline,
		"Entities":      dependencies.Entities,
		"Indicators":    dependencies.Indicators,
		"Assessments":   dependencies.Assessments,
		"Artifacts":     dependencies.Artifacts,
		"Evidence":      dependencies.Evidence,
		"Parties":       dependencies.Parties,
		"Task requests": dependencies.TaskRequests,
		"Decisions":     dependencies.Decisions,
	} {
		if source == nil {
			return nil, fmt.Errorf("operational projection %s source is required", name)
		}
	}

	descriptor := func(providerID string) (ProviderDescriptor, error) {
		canonical, ok := descriptors.Lookup(providerID)
		if !ok {
			return ProviderDescriptor{}, fmt.Errorf("operational projection provider %q is missing", providerID)
		}
		return operationalDescriptor(canonical), nil
	}
	timeline, err := descriptor("timeline")
	if err != nil {
		return nil, err
	}
	host, err := descriptor("host")
	if err != nil {
		return nil, err
	}
	identity, err := descriptor("identity")
	if err != nil {
		return nil, err
	}
	indicator, err := descriptor("indicator")
	if err != nil {
		return nil, err
	}
	assessment, err := descriptor("assessment")
	if err != nil {
		return nil, err
	}
	artifact, err := descriptor("artifact")
	if err != nil {
		return nil, err
	}
	evidence, err := descriptor("evidence")
	if err != nil {
		return nil, err
	}
	party, err := descriptor("party")
	if err != nil {
		return nil, err
	}
	taskRequest, err := descriptor("task_request")
	if err != nil {
		return nil, err
	}
	decision, err := descriptor("decision")
	if err != nil {
		return nil, err
	}

	return NewCatalog([]Provider{
		NewTimelineProvider(timeline, dependencies.Timeline),
		NewHostProvider(host, dependencies.Entities),
		NewIdentityProvider(identity, dependencies.Entities),
		NewIndicatorProvider(indicator, dependencies.Indicators),
		NewAssessmentProvider(assessment, dependencies.Assessments),
		NewArtifactProvider(artifact, dependencies.Artifacts),
		NewEvidenceProvider(evidence, dependencies.Evidence),
		NewPartyProvider(party, dependencies.Parties),
		NewTaskRequestProvider(taskRequest, dependencies.TaskRequests),
		NewDecisionProvider(decision, dependencies.Decisions),
	})
}

func operationalDescriptor(canonical providercontract.ProviderDescriptor) ProviderDescriptor {
	return ProviderDescriptor{
		SchemaVersion:             canonical.SchemaVersion,
		Status:                    ProviderStatus(canonical.Status),
		ProviderKey:               canonical.ProviderID,
		SourceOwnerKey:            canonical.SourceOwnerModule,
		ViewSchemaIDs:             append([]string(nil), canonical.ViewSchemaIDs...),
		SourceRecordTypes:         append([]string(nil), canonical.SourceRecordTypes...),
		SourceAuthorityModules:    append([]string(nil), canonical.SourceAuthorityModules...),
		ProjectionTableFamilies:   append([]string(nil), canonical.ProjectionTableIDs...),
		ProjectionStorageOwnerKey: canonical.ProjectionStorageOwnerModule,
		Capabilities: ProviderCapabilities{
			Query:           canonical.Capabilities.Query,
			RefreshRow:      canonical.Capabilities.RefreshRow,
			RestoreRebuild:  canonical.Capabilities.RestoreRebuild,
			IncidentRebuild: canonical.Capabilities.IncidentRebuild,
		},
		RestoreRebuild:       RestoreRebuildParticipation(canonical.RestoreRebuild),
		FacadePackages:       append([]string(nil), canonical.FacadePackages...),
		RebuildAfter:         append([]string(nil), canonical.RebuildAfter...),
		CharacterizationRefs: append([]string(nil), canonical.CharacterizationRefs...),
	}
}
