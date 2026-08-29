package runtime

import (
	"fmt"
	"reflect"

	artifactprojection "github.com/JochiRaider/cartulary/internal/modules/artifacts/workbookprojection"
	assessmentprojection "github.com/JochiRaider/cartulary/internal/modules/assessments/workbookprojection"
	entitycontract "github.com/JochiRaider/cartulary/internal/modules/entities/projectioncontract"
	evidenceprojection "github.com/JochiRaider/cartulary/internal/modules/evidence/projectioncontract"
	indicatorprojection "github.com/JochiRaider/cartulary/internal/modules/indicators/workbookprojection"
	partyprojection "github.com/JochiRaider/cartulary/internal/modules/parties/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/projections/providercontract"
	taskdecisionprojection "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/projectioncontract"
)

type ProviderSources struct {
	Timeline     TimelineSource
	Entities     entitycontract.SourceReader
	Indicators   indicatorprojection.SourceReader
	Assessments  assessmentprojection.SourceReader
	Artifacts    artifactprojection.SourceReader
	Evidence     evidenceprojection.SourceReader
	Parties      partyprojection.SourceReader
	TaskRequests taskdecisionprojection.TaskRequestSourceReader
	Decisions    taskdecisionprojection.DecisionSourceReader
}

type requiredProviderBinding struct {
	providerID string
	owner      string
	factory    func(providercontract.ProviderDescriptor, ProviderSources) Provider
}

var requiredProviderCatalog = []requiredProviderBinding{
	{providerID: "timeline", owner: "timeline", factory: func(descriptor providercontract.ProviderDescriptor, sources ProviderSources) Provider {
		return newTimelineProvider(descriptor, sources.Timeline)
	}},
	{providerID: "host", owner: "entities", factory: func(descriptor providercontract.ProviderDescriptor, sources ProviderSources) Provider {
		return newHostProvider(descriptor, sources.Entities)
	}},
	{providerID: "identity", owner: "entities", factory: func(descriptor providercontract.ProviderDescriptor, sources ProviderSources) Provider {
		return newIdentityProvider(descriptor, sources.Entities)
	}},
	{providerID: "indicator", owner: "indicators", factory: func(descriptor providercontract.ProviderDescriptor, sources ProviderSources) Provider {
		return newIndicatorProvider(descriptor, sources.Indicators)
	}},
	{providerID: "assessment", owner: "assessments", factory: func(descriptor providercontract.ProviderDescriptor, sources ProviderSources) Provider {
		return newAssessmentProvider(descriptor, sources.Assessments)
	}},
	{providerID: "artifact", owner: "artifacts", factory: func(descriptor providercontract.ProviderDescriptor, sources ProviderSources) Provider {
		return newArtifactProvider(descriptor, sources.Artifacts)
	}},
	{providerID: "evidence", owner: "evidence", factory: func(descriptor providercontract.ProviderDescriptor, sources ProviderSources) Provider {
		return newEvidenceProvider(descriptor, sources.Evidence)
	}},
	{providerID: "party", owner: "parties", factory: func(descriptor providercontract.ProviderDescriptor, sources ProviderSources) Provider {
		return newPartyProvider(descriptor, sources.Parties)
	}},
	{providerID: "task_request", owner: "tasksdecisions", factory: func(descriptor providercontract.ProviderDescriptor, sources ProviderSources) Provider {
		return newTaskRequestProvider(descriptor, sources.TaskRequests)
	}},
	{providerID: "decision", owner: "tasksdecisions", factory: func(descriptor providercontract.ProviderDescriptor, sources ProviderSources) Provider {
		return newDecisionProvider(descriptor, sources.Decisions)
	}},
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
		if isNilSource(required.source) {
			return nil, fmt.Errorf("projection %s source is required", required.name)
		}
	}

	providers := make([]Provider, 0, len(requiredProviderCatalog))
	for _, required := range requiredProviderCatalog {
		descriptor, ok := facts.descriptors.Lookup(required.providerID)
		if !ok {
			return nil, fmt.Errorf("missing active projection provider %q", required.providerID)
		}
		providers = append(providers, required.factory(descriptor, sources))
	}
	catalog, err := newCatalog(facts.descriptors, providers)
	if err != nil {
		return nil, err
	}
	if err := validateSemanticIntents(catalog.registry, facts.intents, facts.intentOwners); err != nil {
		return nil, err
	}
	return catalog, nil
}

func isNilSource(source any) bool {
	if source == nil {
		return true
	}
	value := reflect.ValueOf(source)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}
