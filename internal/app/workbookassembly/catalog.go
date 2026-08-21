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
	entityProviders, err := newEntityProviderSet(entityStore)
	if err != nil {
		return nil, fmt.Errorf("compose workbook contribution catalog: %w", err)
	}
	assessmentFacade, err := NewAssessmentMutationContribution(
		pool,
		assessmentProjections,
		entityStore,
		appender,
	)
	if err != nil {
		return nil, fmt.Errorf("compose workbook contribution catalog: %w", err)
	}
	assessmentCreateProvider, err := newAssessmentCreateProvider(assessmentFacade)
	if err != nil {
		return nil, fmt.Errorf("compose workbook contribution catalog: %w", err)
	}
	partyOwner := parties.NewMutationFacade(
		pool,
		conflictTokens,
		appender,
		conflictFields,
		keepSaved,
		partyProjections,
	)
	partyProviders, err := newPartyProviderSet(partyOwner)
	if err != nil {
		return nil, fmt.Errorf("compose workbook contribution catalog: %w", err)
	}
	evidenceProviders, err := newEvidenceProviderSet(evidenceOwner)
	if err != nil {
		return nil, fmt.Errorf("compose workbook contribution catalog: %w", err)
	}
	artifactProviders, err := newArtifactProviderSet(artifactOwner)
	if err != nil {
		return nil, fmt.Errorf("compose workbook contribution catalog: %w", err)
	}
	taskDecisionProviders, err := newTaskDecisionProviderSet(taskDecisionOwner)
	if err != nil {
		return nil, fmt.Errorf("compose workbook contribution catalog: %w", err)
	}
	timelineProviders, err := newTimelineProviderSet(timelineOwner)
	if err != nil {
		return nil, fmt.Errorf("compose workbook contribution catalog: %w", err)
	}
	indicatorCreateProvider, err := newIndicatorCreateProvider(indicatorOwner)
	if err != nil {
		return nil, fmt.Errorf("compose workbook contribution catalog: %w", err)
	}
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
		timeline.TimelineViewSchemaID:              timelineProviders.create,
		hostidentity.HostsViewSchemaID:             entityProviders.hostCreate,
		hostidentity.IdentitiesViewSchemaID:        entityProviders.identityCreate,
		indicators.ViewSchemaID:                    indicatorCreateProvider,
		assessments.AssessmentsViewSchemaID:        assessmentCreateProvider,
		artifacts.NotesViewSchemaID:                artifactProviders.creates[artifacts.NotesViewSchemaID],
		artifacts.CommLogViewSchemaID:              artifactProviders.creates[artifacts.CommLogViewSchemaID],
		artifacts.HandoffViewSchemaID:              artifactProviders.creates[artifacts.HandoffViewSchemaID],
		artifacts.StatusReviewViewSchemaID:         artifactProviders.creates[artifacts.StatusReviewViewSchemaID],
		artifacts.LessonViewSchemaID:               artifactProviders.creates[artifacts.LessonViewSchemaID],
		artifacts.FindingsViewSchemaID:             artifactProviders.creates[artifacts.FindingsViewSchemaID],
		artifacts.InvestigativeQueriesViewSchemaID: artifactProviders.creates[artifacts.InvestigativeQueriesViewSchemaID],
		artifacts.ForensicKeywordsViewSchemaID:     artifactProviders.creates[artifacts.ForensicKeywordsViewSchemaID],
		evidence.ViewSchemaID:                      evidenceProviders.create,
		parties.ViewSchemaID:                       partyProviders.create,
		tasksdecisions.TaskRequestsViewSchemaID:    taskDecisionProviders.creates[tasksdecisions.TaskRequestsViewSchemaID],
		tasksdecisions.DecisionsViewSchemaID:       taskDecisionProviders.creates[tasksdecisions.DecisionsViewSchemaID],
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
			Provider:      timelineProviders.patch,
		},
		{
			RecordType:    "host",
			ViewSchemaIDs: []string{hostidentity.HostsViewSchemaID},
			Provider:      entityProviders.hostPatch,
		},
		{
			RecordType:    "identity",
			ViewSchemaIDs: []string{hostidentity.IdentitiesViewSchemaID},
			Provider:      entityProviders.identityPatch,
		},
		{
			RecordType:    "artifact",
			ViewSchemaIDs: append([]string(nil), artifactViewSchemaIDs...),
			Provider:      artifactProviders.patch,
		},
		{
			RecordType:    "evidence",
			ViewSchemaIDs: []string{evidence.ViewSchemaID},
			Provider:      evidenceProviders.patch,
		},
		{
			RecordType:    "party",
			ViewSchemaIDs: []string{parties.ViewSchemaID},
			Provider:      partyProviders.patch,
		},
		{
			RecordType:    "task_request",
			ViewSchemaIDs: []string{tasksdecisions.TaskRequestsViewSchemaID},
			Provider:      taskDecisionProviders.patches["task_request"],
		},
		{
			RecordType:    "decision",
			ViewSchemaIDs: []string{tasksdecisions.DecisionsViewSchemaID},
			Provider:      taskDecisionProviders.patches["decision"],
		},
	}
	conflictContributions := []workbook.ConflictContribution{
		{
			RecordType:    "timeline_event",
			ViewSchemaIDs: []string{timeline.TimelineViewSchemaID},
			Provider:      timelineProviders.conflict,
		},
		{
			RecordType:    "host",
			ViewSchemaIDs: []string{hostidentity.HostsViewSchemaID},
			Provider:      entityProviders.hostConflict,
		},
		{
			RecordType:    "identity",
			ViewSchemaIDs: []string{hostidentity.IdentitiesViewSchemaID},
			Provider:      entityProviders.identityConflict,
		},
		{
			RecordType:    "artifact",
			ViewSchemaIDs: append([]string(nil), artifactViewSchemaIDs...),
			Provider:      artifactProviders.conflict,
		},
		{
			RecordType:    "evidence",
			ViewSchemaIDs: []string{evidence.ViewSchemaID},
			Provider:      evidenceProviders.conflict,
		},
		{
			RecordType:    "party",
			ViewSchemaIDs: []string{parties.ViewSchemaID},
			Provider:      partyProviders.conflict,
		},
		{
			RecordType:    "task_request",
			ViewSchemaIDs: []string{tasksdecisions.TaskRequestsViewSchemaID},
			Provider:      taskDecisionProviders.conflicts["task_request"],
		},
		{
			RecordType:    "decision",
			ViewSchemaIDs: []string{tasksdecisions.DecisionsViewSchemaID},
			Provider:      taskDecisionProviders.conflicts["decision"],
		},
	}
	timelineClipboardProvider, err := newTimelineClipboardProvider(timelineOwner)
	if err != nil {
		return nil, fmt.Errorf("compose workbook Timeline clipboard provider: %w", err)
	}
	hostClipboardProvider, err := newEntityClipboardProvider(hostidentity.HostsViewSchemaID, entityStore)
	if err != nil {
		return nil, fmt.Errorf("compose workbook Hosts clipboard provider: %w", err)
	}
	identityClipboardProvider, err := newEntityClipboardProvider(hostidentity.IdentitiesViewSchemaID, entityStore)
	if err != nil {
		return nil, fmt.Errorf("compose workbook Identities clipboard provider: %w", err)
	}
	timelineBulkProvider, err := newTimelineBulkProvider(timelineOwner)
	if err != nil {
		return nil, fmt.Errorf("compose workbook Timeline bulk provider: %w", err)
	}
	linkedNoteProvider, err := newLinkedNoteProvider(artifactOwner)
	if err != nil {
		return nil, fmt.Errorf("compose workbook linked-note provider: %w", err)
	}
	decisionSupersedeProvider, err := newDecisionSupersedeProvider(taskDecisionOwner)
	if err != nil {
		return nil, fmt.Errorf("compose workbook Decision supersede provider: %w", err)
	}
	timelineSupersedeProvider, err := newTimelineSupersedeProvider(timelineOwner)
	if err != nil {
		return nil, fmt.Errorf("compose workbook Timeline supersede provider: %w", err)
	}
	actionContributions := workbook.MutationActionContributions{
		Clipboard: []workbook.ClipboardContribution{
			{
				ViewSchemaID: timeline.TimelineViewSchemaID,
				Provider:     timelineClipboardProvider,
			},
			{
				ViewSchemaID: hostidentity.HostsViewSchemaID,
				Provider:     hostClipboardProvider,
			},
			{
				ViewSchemaID: hostidentity.IdentitiesViewSchemaID,
				Provider:     identityClipboardProvider,
			},
		},
		Bulk: []workbook.BulkContribution{
			{
				ViewSchemaID: timeline.TimelineViewSchemaID,
				Provider:     timelineBulkProvider,
			},
		},
		LinkedNote: []workbook.LinkedNoteContribution{
			{RecordType: "evidence", Provider: linkedNoteProvider},
			{RecordType: "host", Provider: linkedNoteProvider},
			{RecordType: "identity", Provider: linkedNoteProvider},
			{RecordType: "timeline_event", Provider: linkedNoteProvider},
		},
		Supersede: []workbook.SupersedeContribution{
			{
				RecordType: "decision",
				Provider:   decisionSupersedeProvider,
			},
			{
				RecordType: "timeline_event",
				Provider:   timelineSupersedeProvider,
			},
		},
	}
	catalog, err := workbook.NewWorkbookContributionCatalog(
		projectionDescriptors,
		queryContributions,
		createContributions,
		patchContributions,
		conflictContributions,
		workbook.ActionCapabilityRequirements{
			ClipboardViewSchemaIDs: []string{
				hostidentity.HostsViewSchemaID,
				hostidentity.IdentitiesViewSchemaID,
				timeline.TimelineViewSchemaID,
			},
			BulkViewSchemaIDs: []string{timeline.TimelineViewSchemaID},
			LinkedNoteRecordTypes: []string{
				"evidence", "host", "identity", "timeline_event",
			},
			SupersedeRecordTypes: []string{"decision", "timeline_event"},
		},
		actionContributions,
	)
	if err != nil {
		return nil, fmt.Errorf("compose workbook contribution catalog: %w", err)
	}
	return catalog, nil
}
