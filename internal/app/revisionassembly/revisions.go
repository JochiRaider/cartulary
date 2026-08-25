package revisionassembly

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
	"github.com/JochiRaider/cartulary/internal/modules/assessments"
	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/entities"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/parties"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

// Runtime is the composition-scoped Revisions boundary. Build completes all
// immutable catalog validation before mutable source facades are constructed.
type Runtime struct {
	appender        *revisions.Appender
	targetSemantics *revisions.TargetSemanticsCatalog
	deleteRestore   *revisions.DeleteRestoreSourceCatalog
	fieldResolver   *conflicts.FieldResolverCatalog
}

func CurrentProviderContributions() ([]revisions.ProviderContribution, error) {
	artifactContribution, err := artifacts.NewRevisionContribution()
	if err != nil {
		return nil, fmt.Errorf("revision assembly: compose Artifacts contribution: %w", err)
	}
	return []revisions.ProviderContribution{
		artifactContribution,
		assessments.RevisionProviderContribution(),
		entities.RevisionProviderContribution(),
		evidence.RevisionProviderContribution(),
		indicators.NewRevisionContribution(),
		links.RevisionProviderContribution(),
		parties.NewRevisionContribution(),
		tasksdecisions.NewRevisionContribution(),
		timeline.RevisionProviderContribution(),
	}, nil
}

func CurrentConflictFieldResolver() (conflicts.FieldResolver, error) {
	contributions, err := CurrentProviderContributions()
	if err != nil {
		return nil, err
	}
	return buildConflictFieldResolver(contributions)
}

func CurrentTargetSemanticsCatalog() (*revisions.TargetSemanticsCatalog, error) {
	contributions, err := CurrentProviderContributions()
	if err != nil {
		return nil, err
	}
	return buildTargetSemanticsCatalog(contributions)
}

func NewRecordEnvelopeReader(db postgres.DB) revisions.RecordEnvelopeReader {
	return recordEnvelopeAdapter{store: records.NewStore(db)}
}

func Build(contributions ...revisions.ProviderContribution) (*Runtime, error) {
	copied := cloneProviderContributions(contributions)
	if err := revisions.ValidateProviderContributions(copied); err != nil {
		return nil, fmt.Errorf("revision assembly: validate provider contributions: %w", err)
	}
	if _, err := buildRecordViewCatalog(copied); err != nil {
		return nil, fmt.Errorf("revision assembly: build record/view catalog: %w", err)
	}
	fieldResolver, err := buildConflictFieldResolver(copied)
	if err != nil {
		return nil, fmt.Errorf("revision assembly: build conflict field resolver catalog: %w", err)
	}
	snapshotCaptures, err := revisions.NewRecordSnapshotCaptureCatalog(copied)
	if err != nil {
		return nil, fmt.Errorf("revision assembly: build snapshot capture catalog: %w", err)
	}
	targetSemantics, err := buildTargetSemanticsCatalog(copied)
	if err != nil {
		return nil, fmt.Errorf("revision assembly: build target semantics catalog: %w", err)
	}
	deleteRestore, err := revisions.NewDeleteRestoreSourceCatalogFromContributions(copied)
	if err != nil {
		return nil, fmt.Errorf("revision assembly: build delete/restore source catalog: %w", err)
	}
	appender, err := revisions.NewAppender(
		recordEnvelopeAdapter{store: records.NewStore()},
		snapshotCaptures,
		targetSemantics,
	)
	if err != nil {
		return nil, fmt.Errorf("revision assembly: build appender: %w", err)
	}
	return &Runtime{
		appender:        appender,
		targetSemantics: targetSemantics,
		deleteRestore:   deleteRestore,
		fieldResolver:   fieldResolver,
	}, nil
}

func (r *Runtime) ConflictFieldResolver() conflicts.FieldResolver {
	if r == nil {
		return nil
	}
	return r.fieldResolver
}

func (r *Runtime) Appender() *revisions.Appender {
	if r == nil {
		return nil
	}
	return r.appender
}

func (r *Runtime) NewCommandService(
	db postgres.DB,
	attributionResolver revisions.ImportedAttributionResolver,
	projections revisions.ProjectionRebuilder,
	liveRecords revisions.LiveRecordReader,
	clock func() time.Time,
	publications collaboration.RecordChangedAppender,
) (*revisions.CommandService, error) {
	if r == nil || r.appender == nil || r.deleteRestore == nil || r.targetSemantics == nil {
		return nil, errors.New("revision assembly: runtime is required")
	}
	return revisions.NewCommandService(revisions.CommandServiceDependencies{
		Transactions:                transactionRunnerAdapter{database: db},
		Authorization:               commandAuthorizerAdapter{access: admission.NewChecker(db)},
		Idempotency:                 commandIdempotencyAdapter{store: authn.NewStore(db)},
		ImportedAttributionResolver: attributionResolver,
		Projections:                 projections,
		LiveRecords:                 liveRecords,
		DeleteRestoreSources:        r.deleteRestore,
		TargetSemantics:             r.targetSemantics,
		Appender:                    r.appender,
		RecordEnvelopes:             recordEnvelopeAdapter{store: records.NewStore(db)},
		RecordPublications:          recordPublicationAdapter{appender: publications},
		Clock:                       clock,
	})
}

type recordPublicationAdapter struct {
	appender collaboration.RecordChangedAppender
}

func (adapter recordPublicationAdapter) AppendRecordChangedTx(ctx context.Context, tx pgx.Tx, effect revisions.RecordPublicationEffect) error {
	return adapter.appender.AppendRecordChangedTx(ctx, tx, collaboration.RecordChangeIntentInput{
		IncidentID: effect.IncidentID, RecordID: effect.RecordID, ChangeSetID: effect.ChangeSetID,
		ActorUserID: effect.ActorUserID, RowVersion: effect.RowVersion, ClientTxnID: effect.ClientTxnID,
		MutationOrdinal: effect.MutationOrdinal, CreatedAt: effect.CreatedAt,
		PublicFieldKeys: effect.PublicFieldKeys,
		AffectedViews: []collaboration.AffectedViewChange{{
			ViewSchemaID: effect.ViewSchemaID, RecordID: effect.RecordID, RowVersion: effect.RowVersion,
			ChangeKind: effect.ChangeKind,
		}},
	})
}

func cloneProviderContributions(values []revisions.ProviderContribution) []revisions.ProviderContribution {
	cloned := make([]revisions.ProviderContribution, len(values))
	for index, contribution := range values {
		cloned[index] = contribution
		cloned[index].Records = make([]revisions.RecordProviderContribution, len(contribution.Records))
		for recordIndex, record := range contribution.Records {
			cloned[index].Records[recordIndex] = record
			cloned[index].Records[recordIndex].HistoryTargetKinds = append(
				[]string(nil),
				record.HistoryTargetKinds...,
			)
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

func buildTargetSemanticsCatalog(contributions []revisions.ProviderContribution) (*revisions.TargetSemanticsCatalog, error) {
	return revisions.NewTargetSemanticsCatalog(contributions)
}

func buildRecordViewCatalog(contributions []revisions.ProviderContribution) (*revisions.RecordViewCatalog, error) {
	publicResources := viewschema.ListPublicResources()
	viewSchemaIDs := make([]string, 0, len(publicResources))
	recordViewSurfaces := make([]revisions.RecordViewSurface, 0, len(publicResources))
	for _, resource := range publicResources {
		viewSchemaIDs = append(viewSchemaIDs, resource.ViewSchemaID)
		recordViewSurfaces = append(recordViewSurfaces, revisions.RecordViewSurface{
			SourceRecordTypes: append([]string(nil), resource.SourceRecordTypes...),
			ViewSchemaID:      resource.ViewSchemaID,
		})
	}
	return revisions.NewRecordViewCatalog(contributions, recordViewSurfaces, viewSchemaIDs)
}

func buildConflictFieldResolver(contributions []revisions.ProviderContribution) (*conflicts.FieldResolverCatalog, error) {
	required := make([]string, 0)
	providers := make([]conflicts.FieldResolverContribution, 0)
	for _, contribution := range contributions {
		for _, record := range contribution.Records {
			for _, route := range record.RecordViewRoutes {
				for _, viewSchemaID := range route.ViewSchemaIDs {
					schema, ok := viewschema.Lookup(viewSchemaID)
					if !ok {
						return nil, fmt.Errorf("%w: unknown view schema %q", conflicts.ErrMissingFieldResolver, viewSchemaID)
					}
					fields := schema.Fields()
					descriptors := make([]conflicts.FieldDescriptor, 0, len(fields))
					for _, field := range fields {
						descriptors = append(descriptors, conflicts.FieldDescriptor{
							FieldKey:                field.FieldKey,
							ValueKind:               field.WriteKind,
							Writable:                field.Writable,
							ConflictResolutionClass: field.ConflictResolutionClass,
						})
					}
					required = append(required, viewSchemaID)
					providers = append(providers, conflicts.FieldResolverContribution{
						ProviderID:    route.ContributionID + "#" + viewSchemaID,
						SourceOwnerID: string(record.SourceOwnerModule),
						ViewSchemaID:  viewSchemaID,
						Fields:        descriptors,
					})
				}
			}
		}
	}
	return conflicts.NewFieldResolverCatalog(required, providers...)
}
