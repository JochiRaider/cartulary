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
	taskdecisionprojection "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/projectioncontract"
)

type ProviderSources struct {
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

func NewCatalog(
	contributions []providercontract.Contribution,
	sources ProviderSources,
) (*Catalog, error) {
	facts, err := collectContractFacts(contributions)
	if err != nil {
		return nil, err
	}
	requiredSources := []struct {
		name   string
		source any
	}{
		{name: "Timeline", source: sources.Timeline},
		{name: "Entities", source: sources.Entities},
		{name: "Indicators", source: sources.Indicators},
		{name: "Assessments", source: sources.Assessments},
		{name: "Artifacts", source: sources.Artifacts},
		{name: "Evidence", source: sources.Evidence},
		{name: "Parties", source: sources.Parties},
		{name: "Task requests", source: sources.TaskRequests},
		{name: "Decisions", source: sources.Decisions},
	}
	for _, required := range requiredSources {
		if required.source == nil {
			return nil, fmt.Errorf("projection %s source is required", required.name)
		}
	}

	providerIDs := []string{
		"timeline", "host", "identity", "indicator", "assessment",
		"artifact", "evidence", "party", "task_request", "decision",
	}
	descriptors := make(map[string]providercontract.ProviderDescriptor, len(providerIDs))
	for _, providerID := range providerIDs {
		descriptor, ok := facts.descriptors.Lookup(providerID)
		if !ok {
			return nil, fmt.Errorf("missing active projection provider %q", providerID)
		}
		descriptors[providerID] = descriptor
	}
	catalog, err := newCatalog(facts.descriptors, []Provider{
		NewTimelineProvider(descriptors["timeline"], sources.Timeline),
		NewHostProvider(descriptors["host"], sources.Entities),
		NewIdentityProvider(descriptors["identity"], sources.Entities),
		NewIndicatorProvider(descriptors["indicator"], sources.Indicators),
		NewAssessmentProvider(descriptors["assessment"], sources.Assessments),
		NewArtifactProvider(descriptors["artifact"], sources.Artifacts),
		NewEvidenceProvider(descriptors["evidence"], sources.Evidence),
		NewPartyProvider(descriptors["party"], sources.Parties),
		NewTaskRequestProvider(descriptors["task_request"], sources.TaskRequests),
		NewDecisionProvider(descriptors["decision"], sources.Decisions),
	})
	if err != nil {
		return nil, err
	}
	if err := validateSemanticIntents(catalog.registry, facts.intents, facts.intentOwners); err != nil {
		return nil, err
	}
	return catalog, nil
}
