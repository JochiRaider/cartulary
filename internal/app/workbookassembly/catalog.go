package workbookassembly

import (
	"context"
	"fmt"

	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
	"github.com/JochiRaider/cartulary/internal/modules/assessments"
	assessmentprojection "github.com/JochiRaider/cartulary/internal/modules/assessments/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity"
	entityprojection "github.com/JochiRaider/cartulary/internal/modules/entities/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	"github.com/JochiRaider/cartulary/internal/modules/parties"
	partyprojection "github.com/JochiRaider/cartulary/internal/modules/parties/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/projections/providercontract"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/modules/workbook"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/querypage"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

type projectionQueryCatalog interface {
	WorkbookQueryProvider(string) (workbook.QueryProvider, bool)
}

func NewContributionCatalog(
	pool postgres.DB,
	projectionDescriptors providercontract.DescriptorSet,
	projectionQueries projectionQueryCatalog,
	entityProjections entityprojection.Ports,
	assessmentProjections assessmentprojection.Rows,
	partyProjections partyprojection.Rows,
	indicatorOwner *indicators.Store,
	timelineOwner *timeline.Facade,
	evidenceOwner evidence.MutationContribution,
	artifactOwner *artifacts.MutationFacade,
	taskDecisionOwner *tasksdecisions.MutationFacade,
	conflictTokens conflicttokens.ConflictTokenCodec,
	conflictFields conflicttokens.FieldResolver,
	appender *revisions.Appender,
	intents collaboration.IntentAppender,
) (*workbook.WorkbookContributionCatalog, error) {
	if projectionDescriptors.Len() == 0 {
		return nil, fmt.Errorf("compose workbook contribution catalog: projection descriptors are required")
	}
	if projectionQueries == nil {
		return nil, fmt.Errorf("compose workbook contribution catalog: projection queries are required")
	}
	if entityProjections.Writer == nil || entityProjections.Reader == nil {
		return nil, fmt.Errorf("compose workbook contribution catalog: Entities projection ports are required")
	}
	if assessmentProjections == nil {
		return nil, fmt.Errorf("compose workbook contribution catalog: Assessments projection rows are required")
	}
	if partyProjections == nil {
		return nil, fmt.Errorf("compose workbook contribution catalog: Parties projection rows are required")
	}
	if indicatorOwner == nil {
		return nil, fmt.Errorf("compose workbook contribution catalog: Indicators owner is required")
	}
	if timelineOwner == nil {
		return nil, fmt.Errorf("compose workbook contribution catalog: Timeline owner is required")
	}
	if evidenceOwner == nil {
		return nil, fmt.Errorf("compose workbook contribution catalog: Evidence contribution is required")
	}
	if artifactOwner == nil {
		return nil, fmt.Errorf("compose workbook contribution catalog: Artifacts mutation contribution is required")
	}
	if taskDecisionOwner == nil {
		return nil, fmt.Errorf("compose workbook contribution catalog: Tasks/Decisions mutation contribution is required")
	}
	if intents == nil {
		return nil, fmt.Errorf("compose workbook contribution catalog: Collaboration intent appender is required")
	}
	if conflictFields == nil {
		return nil, fmt.Errorf("compose workbook contribution catalog: Revisions conflict field resolver is required")
	}
	keepSaved := NewConflictIdempotencyPort(pool)

	entityStore := hostidentity.NewStore(
		pool,
		appender,
		keepSaved,
		entityProjections.Writer,
		hostidentity.WithProjectionReader(entityProjections.Reader),
	)
	assessmentFacade, err := newAssessmentFacade(
		pool,
		assessmentProjections,
		entityStore,
		appender,
	)
	if err != nil {
		return nil, fmt.Errorf("compose workbook contribution catalog: %w", err)
	}
	partyOwner := parties.NewWorkbookFacade(
		pool,
		conflictTokens,
		appender,
		conflictFields,
		keepSaved,
		partyProjections,
	)
	sourceQueries := map[string]workbook.QueryProvider{
		hostidentity.HostsViewSchemaID: workbook.QueryProviderFunc(
			func(
				ctx context.Context,
				command workbook.QueryCommand,
			) (querypage.Result, error) {
				if command.ViewSchemaID != hostidentity.HostsViewSchemaID {
					return querypage.Result{}, fmt.Errorf(
						"host query contribution received view schema %q",
						command.ViewSchemaID,
					)
				}
				return entityStore.QueryHostRowsPage(ctx, command.IncidentID, command.Query, command.Window)
			},
		),
		hostidentity.IdentitiesViewSchemaID: workbook.QueryProviderFunc(
			func(
				ctx context.Context,
				command workbook.QueryCommand,
			) (querypage.Result, error) {
				if command.ViewSchemaID != hostidentity.IdentitiesViewSchemaID {
					return querypage.Result{}, fmt.Errorf(
						"identity query contribution received view schema %q",
						command.ViewSchemaID,
					)
				}
				return entityStore.QueryIdentityRowsPage(ctx, command.IncidentID, command.Query, command.Window)
			},
		),
	}
	createProviders := map[string]workbook.CreateProvider{
		timeline.TimelineViewSchemaID:       workbook.NewTimelineCreateProvider(timelineOwner),
		hostidentity.HostsViewSchemaID:      workbook.NewHostCreateProvider(entityStore),
		hostidentity.IdentitiesViewSchemaID: workbook.NewIdentityCreateProvider(entityStore),
		indicators.ViewSchemaID:             workbook.NewIndicatorCreateProvider(indicatorOwner),
		assessments.AssessmentsViewSchemaID: workbook.NewAssessmentCreateProvider(assessmentFacade),
		workbook.NotesViewSchemaID:          workbook.NewArtifactCreateProvider(workbook.NotesViewSchemaID, artifactOwner),
		workbook.CommLogViewSchemaID:        workbook.NewArtifactCreateProvider(workbook.CommLogViewSchemaID, artifactOwner),
		workbook.HandoffViewSchemaID:        workbook.NewArtifactCreateProvider(workbook.HandoffViewSchemaID, artifactOwner),
		workbook.StatusReviewViewSchemaID:   workbook.NewArtifactCreateProvider(workbook.StatusReviewViewSchemaID, artifactOwner),
		workbook.LessonViewSchemaID:         workbook.NewArtifactCreateProvider(workbook.LessonViewSchemaID, artifactOwner),
		workbook.FindingsViewSchemaID:       workbook.NewArtifactCreateProvider(workbook.FindingsViewSchemaID, artifactOwner),
		workbook.InvestigativeQueriesViewSchemaID: workbook.NewArtifactCreateProvider(
			workbook.InvestigativeQueriesViewSchemaID,
			artifactOwner,
		),
		workbook.ForensicKeywordsViewSchemaID: workbook.NewArtifactCreateProvider(
			workbook.ForensicKeywordsViewSchemaID,
			artifactOwner,
		),
		workbook.EvidenceViewSchemaID:     workbook.NewEvidenceCreateProvider(evidenceOwner),
		workbook.PartiesViewSchemaID:      workbook.NewPartyCreateProvider(partyOwner),
		workbook.TaskRequestsViewSchemaID: workbook.NewTaskDecisionCreateProvider(workbook.TaskRequestsViewSchemaID, taskDecisionOwner),
		workbook.DecisionsViewSchemaID:    workbook.NewTaskDecisionCreateProvider(workbook.DecisionsViewSchemaID, taskDecisionOwner),
	}

	descriptors := projectionDescriptors.All()
	queryContributions := make([]workbook.QueryContribution, 0, len(viewschema.ListPublicResources()))
	createContributions := make([]workbook.CreateContribution, 0, len(viewschema.ListPublicResources()))
	for _, descriptor := range descriptors {
		if descriptor.Status != providercontract.ProviderStatusActive {
			continue
		}
		for _, viewSchemaID := range descriptor.ViewSchemaIDs {
			provider := sourceQueries[viewSchemaID]
			backendKind := workbook.QueryBackendSourceOwner
			if descriptor.Capabilities.Query {
				backendKind = workbook.QueryBackendProjection
				if provider != nil {
					// Typed projection readers may select bounded derived rows and
					// delegate exact-ID authoritative hydration to the source owner.
					// They intentionally do not register a generic raw-SQL surface.
				} else if projectionProvider, ok := projectionQueries.WorkbookQueryProvider(viewSchemaID); !ok {
					return nil, fmt.Errorf(
						"compose workbook contribution catalog: projection descriptor %q declares unsupported query surface %q",
						descriptor.ProviderID,
						viewSchemaID,
					)
				} else {
					provider = projectionProvider
				}
			}
			if provider == nil {
				continue
			}
			queryContributions = append(queryContributions, workbook.QueryContribution{
				ViewSchemaID:      viewSchemaID,
				SourceOwnerKey:    descriptor.SourceOwnerModule,
				SourceRecordTypes: append([]string(nil), descriptor.SourceRecordTypes...),
				BackendKind:       backendKind,
				Provider:          provider,
			})
			createProvider := createProviders[viewSchemaID]
			if createProvider != nil {
				createContributions = append(createContributions, workbook.CreateContribution{
					ViewSchemaID:      viewSchemaID,
					SourceOwnerKey:    descriptor.SourceOwnerModule,
					SourceRecordTypes: append([]string(nil), descriptor.SourceRecordTypes...),
					Provider:          createProvider,
				})
			}
		}
	}
	patchContributions := []workbook.PatchContribution{
		{
			RecordType:    "timeline_event",
			ViewSchemaIDs: []string{timeline.TimelineViewSchemaID},
			Provider:      workbook.NewTimelinePatchProvider(timelineOwner),
		},
		{
			RecordType:    "host",
			ViewSchemaIDs: []string{hostidentity.HostsViewSchemaID},
			Provider:      workbook.NewHostPatchProvider(entityStore),
		},
		{
			RecordType:    "identity",
			ViewSchemaIDs: []string{hostidentity.IdentitiesViewSchemaID},
			Provider:      workbook.NewIdentityPatchProvider(entityStore),
		},
		{
			RecordType: "artifact",
			ViewSchemaIDs: []string{
				workbook.CommLogViewSchemaID,
				workbook.FindingsViewSchemaID,
				workbook.ForensicKeywordsViewSchemaID,
				workbook.HandoffViewSchemaID,
				workbook.InvestigativeQueriesViewSchemaID,
				workbook.LessonViewSchemaID,
				workbook.NotesViewSchemaID,
				workbook.StatusReviewViewSchemaID,
			},
			Provider: workbook.NewArtifactPatchProvider(artifactOwner),
		},
		{
			RecordType:    "evidence",
			ViewSchemaIDs: []string{workbook.EvidenceViewSchemaID},
			Provider:      workbook.NewEvidencePatchProvider(evidenceOwner),
		},
		{
			RecordType:    "party",
			ViewSchemaIDs: []string{workbook.PartiesViewSchemaID},
			Provider:      workbook.NewPartyPatchProvider(partyOwner),
		},
		{
			RecordType:    "task_request",
			ViewSchemaIDs: []string{workbook.TaskRequestsViewSchemaID},
			Provider: workbook.NewTaskDecisionPatchProvider(
				"task_request",
				workbook.TaskRequestsViewSchemaID,
				taskDecisionOwner,
			),
		},
		{
			RecordType:    "decision",
			ViewSchemaIDs: []string{workbook.DecisionsViewSchemaID},
			Provider: workbook.NewTaskDecisionPatchProvider(
				"decision",
				workbook.DecisionsViewSchemaID,
				taskDecisionOwner,
			),
		},
	}
	conflictContributions := []workbook.ConflictContribution{
		{
			RecordType:    "timeline_event",
			ViewSchemaIDs: []string{timeline.TimelineViewSchemaID},
			Provider:      workbook.NewTimelineConflictProvider(timelineOwner),
		},
		{
			RecordType:    "host",
			ViewSchemaIDs: []string{hostidentity.HostsViewSchemaID},
			Provider:      workbook.NewHostConflictProvider(entityStore),
		},
		{
			RecordType:    "identity",
			ViewSchemaIDs: []string{hostidentity.IdentitiesViewSchemaID},
			Provider:      workbook.NewIdentityConflictProvider(entityStore),
		},
		{
			RecordType: "artifact",
			ViewSchemaIDs: []string{
				workbook.CommLogViewSchemaID,
				workbook.FindingsViewSchemaID,
				workbook.ForensicKeywordsViewSchemaID,
				workbook.HandoffViewSchemaID,
				workbook.InvestigativeQueriesViewSchemaID,
				workbook.LessonViewSchemaID,
				workbook.NotesViewSchemaID,
				workbook.StatusReviewViewSchemaID,
			},
			Provider: workbook.NewArtifactConflictProvider(artifactOwner),
		},
		{
			RecordType:    "evidence",
			ViewSchemaIDs: []string{workbook.EvidenceViewSchemaID},
			Provider:      workbook.NewEvidenceConflictProvider(evidenceOwner),
		},
		{
			RecordType:    "party",
			ViewSchemaIDs: []string{workbook.PartiesViewSchemaID},
			Provider:      workbook.NewPartyConflictProvider(partyOwner),
		},
		{
			RecordType:    "task_request",
			ViewSchemaIDs: []string{workbook.TaskRequestsViewSchemaID},
			Provider: workbook.NewTaskDecisionConflictProvider(
				"task_request",
				workbook.TaskRequestsViewSchemaID,
				taskDecisionOwner,
			),
		},
		{
			RecordType:    "decision",
			ViewSchemaIDs: []string{workbook.DecisionsViewSchemaID},
			Provider: workbook.NewTaskDecisionConflictProvider(
				"decision",
				workbook.DecisionsViewSchemaID,
				taskDecisionOwner,
			),
		},
	}
	catalog, err := workbook.NewWorkbookContributionCatalog(
		projectionDescriptors,
		queryContributions,
		createContributions,
		patchContributions,
		conflictContributions,
	)
	if err != nil {
		return nil, fmt.Errorf("compose workbook contribution catalog: %w", err)
	}
	return catalog, nil
}
