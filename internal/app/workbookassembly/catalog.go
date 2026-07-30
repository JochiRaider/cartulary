package workbookassembly

import (
	"context"
	"fmt"

	"github.com/JochiRaider/cartulary/internal/modules/artifacts"
	"github.com/JochiRaider/cartulary/internal/modules/assessments"
	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/indicators"
	"github.com/JochiRaider/cartulary/internal/modules/parties"
	"github.com/JochiRaider/cartulary/internal/modules/projections"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/revisions/conflicttokens"
	"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/modules/workbook"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/querypage"
	"github.com/JochiRaider/cartulary/internal/platform/viewschema"
)

func NewContributionCatalog(
	pool postgres.DB,
	projectionCatalog *projections.Catalog,
	projectionQuery *projections.QueryService,
	timelineOwner *timeline.Facade,
	conflictTokens conflicttokens.ConflictTokenCodec,
	appender *revisions.Appender,
	intents collaboration.IntentAppender,
) (*workbook.WorkbookContributionCatalog, error) {
	if projectionCatalog == nil {
		return nil, fmt.Errorf("compose workbook contribution catalog: projection catalog is required")
	}
	if projectionQuery == nil {
		return nil, fmt.Errorf("compose workbook contribution catalog: projection query service is required")
	}
	if timelineOwner == nil {
		return nil, fmt.Errorf("compose workbook contribution catalog: Timeline owner is required")
	}
	if intents == nil {
		return nil, fmt.Errorf("compose workbook contribution catalog: Collaboration intent appender is required")
	}

	entityStore := hostidentity.NewStore(pool, appender)
	indicatorStore := indicators.NewStore(pool, appender)
	assessmentFacade, err := newAssessmentFacade(pool, projectionCatalog, entityStore, appender)
	if err != nil {
		return nil, fmt.Errorf("compose workbook contribution catalog: %w", err)
	}
	artifactOwner := artifacts.NewWorkbookFacade(pool, conflictTokens, appender)
	evidenceOwner := evidence.NewWorkbookFacade(pool, conflictTokens, appender, intents)
	partyOwner := parties.NewWorkbookFacade(pool, conflictTokens, appender)
	taskDecisionOwner := tasksdecisions.NewWorkbookFacade(pool, conflictTokens, appender)
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
		indicators.ViewSchemaID:             workbook.NewIndicatorCreateProvider(indicatorStore),
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

	descriptors := projectionCatalog.Descriptors()
	queryContributions := make([]workbook.QueryContribution, 0, len(viewschema.ListPublicResources()))
	createContributions := make([]workbook.CreateContribution, 0, len(viewschema.ListPublicResources()))
	for _, descriptor := range descriptors {
		if descriptor.Status != projections.ProviderStatusActive {
			continue
		}
		for _, viewSchemaID := range descriptor.ViewSchemaIDs {
			provider := sourceQueries[viewSchemaID]
			backendKind := workbook.QueryBackendSourceOwner
			if descriptor.Capabilities.Query {
				if !projectionQuery.Supports(viewSchemaID) {
					return nil, fmt.Errorf(
						"compose workbook contribution catalog: projection descriptor %q declares unsupported query surface %q",
						descriptor.ProviderKey,
						viewSchemaID,
					)
				}
				boundViewSchemaID := viewSchemaID
				backendKind = workbook.QueryBackendProjection
				provider = workbook.QueryProviderFunc(
					func(
						ctx context.Context,
						command workbook.QueryCommand,
					) (querypage.Result, error) {
						if command.ViewSchemaID != boundViewSchemaID {
							return querypage.Result{}, fmt.Errorf(
								"projection query contribution %q received view schema %q",
								boundViewSchemaID,
								command.ViewSchemaID,
							)
						}
						return projectionQuery.QueryRowsPage(
							ctx,
							command.IncidentID,
							boundViewSchemaID,
							command.Query,
							command.Window,
						)
					},
				)
			}
			if provider == nil {
				continue
			}
			queryContributions = append(queryContributions, workbook.QueryContribution{
				ViewSchemaID:      viewSchemaID,
				SourceOwnerKey:    descriptor.SourceOwnerKey,
				SourceRecordTypes: append([]string(nil), descriptor.SourceRecordTypes...),
				BackendKind:       backendKind,
				Provider:          provider,
			})
			createProvider := createProviders[viewSchemaID]
			if createProvider != nil {
				createContributions = append(createContributions, workbook.CreateContribution{
					ViewSchemaID:      viewSchemaID,
					SourceOwnerKey:    descriptor.SourceOwnerKey,
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
		descriptors,
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
