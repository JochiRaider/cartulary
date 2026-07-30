package revisionassembly

import (
	"errors"
	"fmt"

	"github.com/JochiRaider/cartulary/internal/app/projectionassembly"
	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
	"github.com/JochiRaider/cartulary/internal/modules/assessments"
	"github.com/JochiRaider/cartulary/internal/modules/entities"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/parties"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

type projectionServices interface {
	revisions.ProjectionServices
}

type Dependencies struct {
	HistoricalIntentPolicy revisions.HistoricalIntentPolicy
	IntentAppender         revisions.IntentAppender
}

// Runtime is the composition-scoped Revisions boundary. Build completes all
// immutable catalog validation before mutable source facades are constructed.
type Runtime struct {
	appender      *revisions.Appender
	recordViews   *revisions.RecordViewCatalog
	contributions []revisions.ProviderContribution
}

func CurrentProviderContributions() []revisions.ProviderContribution {
	return []revisions.ProviderContribution{
		artifacts.RevisionProviderContribution(),
		assessments.RevisionProviderContribution(),
		entities.RevisionProviderContribution(),
		evidence.RevisionProviderContribution(),
		indicators.RevisionProviderContribution(),
		links.RevisionProviderContribution(),
		parties.RevisionProviderContribution(),
		tasksdecisions.RevisionProviderContribution(),
		timeline.RevisionProviderContribution(),
	}
}

func Build(dependencies Dependencies, contributions ...revisions.ProviderContribution) (*Runtime, error) {
	if dependencies.HistoricalIntentPolicy == nil {
		return nil, errors.New("revision assembly: historical intent policy is required")
	}
	copied := cloneProviderContributions(contributions)
	if err := revisions.ValidateProviderContributions(copied); err != nil {
		return nil, fmt.Errorf("revision assembly: validate provider contributions: %w", err)
	}
	projectionCatalog, err := projectionassembly.NewCatalog(nil)
	if err != nil {
		return nil, fmt.Errorf("revision assembly: build projection descriptor catalog: %w", err)
	}
	publicResources := viewschema.ListPublicResources()
	viewSchemaIDs := make([]string, 0, len(publicResources))
	for _, resource := range publicResources {
		viewSchemaIDs = append(viewSchemaIDs, resource.ViewSchemaID)
	}
	projectionDescriptors := projectionCatalog.Descriptors()
	recordViewProjectionDescriptors := make(
		[]revisions.RecordViewProjectionDescriptor,
		0,
		len(projectionDescriptors),
	)
	for _, descriptor := range projectionDescriptors {
		recordViewProjectionDescriptors = append(
			recordViewProjectionDescriptors,
			revisions.RecordViewProjectionDescriptor{
				Active:            descriptor.Status == "active",
				SourceRecordTypes: append([]string(nil), descriptor.SourceRecordTypes...),
				ViewSchemaIDs:     append([]string(nil), descriptor.ViewSchemaIDs...),
			},
		)
	}
	recordViews, err := revisions.NewRecordViewCatalog(
		copied,
		recordViewProjectionDescriptors,
		viewSchemaIDs,
	)
	if err != nil {
		return nil, fmt.Errorf("revision assembly: build record/view catalog: %w", err)
	}
	appender, err := revisions.NewAppender(recordViews, dependencies.HistoricalIntentPolicy, dependencies.IntentAppender)
	if err != nil {
		return nil, fmt.Errorf("revision assembly: build appender: %w", err)
	}
	return &Runtime{
		appender:      appender,
		recordViews:   recordViews,
		contributions: copied,
	}, nil
}

func (r *Runtime) Appender() *revisions.Appender {
	if r == nil {
		return nil
	}
	return r.appender
}

func (r *Runtime) RecordViewCatalog() *revisions.RecordViewCatalog {
	if r == nil {
		return nil
	}
	return r.recordViews
}

func (r *Runtime) NewCommandService(
	db postgres.DB,
	attributionResolver revisions.ImportedAttributionResolver,
	projections projectionServices,
) (*revisions.CommandService, error) {
	if r == nil || r.appender == nil || r.recordViews == nil {
		return nil, errors.New("revision assembly: runtime is required")
	}
	return revisions.NewCommandService(revisions.CommandServiceDependencies{
		Database:                    db,
		ImportedAttributionResolver: attributionResolver,
		Projections:                 projections,
		ProviderContributions:       cloneProviderContributions(r.contributions),
		Appender:                    r.appender,
		EnvelopeStore:               records.NewStore(db),
	})
}

func cloneProviderContributions(values []revisions.ProviderContribution) []revisions.ProviderContribution {
	cloned := make([]revisions.ProviderContribution, len(values))
	for index, contribution := range values {
		cloned[index] = contribution
		cloned[index].Records = make([]revisions.RecordProviderContribution, len(contribution.Records))
		for recordIndex, record := range contribution.Records {
			cloned[index].Records[recordIndex] = record
			cloned[index].Records[recordIndex].RecordViewRoutes = make(
				[]revisions.RecordViewRouteContribution,
				len(record.RecordViewRoutes),
			)
			for routeIndex, route := range record.RecordViewRoutes {
				clonedRoute := route
				clonedRoute.ViewSchemaIDs = append([]string(nil), route.ViewSchemaIDs...)
				if route.Variant != nil {
					variant := *route.Variant
					clonedRoute.Variant = &variant
				}
				cloned[index].Records[recordIndex].RecordViewRoutes[routeIndex] = clonedRoute
			}
		}
		cloned[index].NonRowTargets = append(
			[]revisions.NonRowProviderContribution(nil),
			contribution.NonRowTargets...,
		)
	}
	return cloned
}
