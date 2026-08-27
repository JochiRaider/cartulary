package workbookassembly

import (
	"context"
	"fmt"
	"reflect"

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

type ContributionDependencies struct {
	Postgres              postgres.DB
	ProjectionDescriptors providercontract.DescriptorSet
	ProjectionQueries     projectionQueryCatalog
	EntityProjections     entityprojection.Ports
	AssessmentProjections assessmentprojection.Rows
	PartyProjections      partyprojection.Rows
	IndicatorOwner        *indicators.Application
	TimelineOwner         TimelineOperations
	EvidenceOwner         evidence.MutationContribution
	ArtifactOwner         *artifacts.MutationFacade
	TaskDecisionOwner     *tasksdecisions.MutationFacade
	ConflictTokens        conflicttokens.ConflictTokenCodec
	ConflictFields        conflicttokens.FieldResolver
	Revisions             *revisions.Appender
	CollaborationIntents  collaboration.RecordChangedAppender
}

func NewContributionCatalog(input ContributionDependencies) (*workbook.WorkbookContributionCatalog, error) {
	if err := validateContributionDependencies(input); err != nil {
		return nil, err
	}
	pool := input.Postgres
	projectionDescriptors := input.ProjectionDescriptors
	projectionQueries := input.ProjectionQueries
	entityProjections := input.EntityProjections
	assessmentProjections := input.AssessmentProjections
	partyProjections := input.PartyProjections
	indicatorOwner := input.IndicatorOwner
	timelineOwner := input.TimelineOwner
	evidenceOwner := input.EvidenceOwner
	artifactOwner := input.ArtifactOwner
	taskDecisionOwner := input.TaskDecisionOwner
	conflictTokens := input.ConflictTokens
	conflictFields := input.ConflictFields
	appender := input.Revisions

	keepSaved := NewConflictIdempotencyPort(pool)

	entityStore, err := hostidentity.NewStore(hostidentity.StoreDependencies{
		Postgres:             pool,
		Revisions:            appender,
		ProjectionWriter:     entityProjections.Writer,
		ProjectionReader:     entityProjections.Reader,
		KeepSavedIdempotency: keepSaved,
		Collaboration:        input.CollaborationIntents,
	})
	if err != nil {
		return nil, fmt.Errorf("compose workbook contribution catalog: %w", err)
	}
	return buildContributionCatalog(contributionAssemblyInput{
		projectionDescriptors: projectionDescriptors,
		projectionQueries:     projectionQueries,
		entityStore:           entityStore,
		assessmentProjections: assessmentProjections,
		partyProjections:      partyProjections,
		indicatorOwner:        indicatorOwner,
		timelineOwner:         timelineOwner,
		evidenceOwner:         evidenceOwner,
		artifactOwner:         artifactOwner,
		taskDecisionOwner:     taskDecisionOwner,
		conflictTokens:        conflictTokens,
		conflictFields:        conflictFields,
		appender:              appender,
		collaborationIntents:  input.CollaborationIntents,
		pool:                  pool,
	})
}

func validateContributionDependencies(input ContributionDependencies) error {
	if isNilDependency(input.Postgres) {
		return fmt.Errorf("compose workbook contribution catalog: Postgres is required")
	}
	if input.ProjectionDescriptors.Len() == 0 {
		return fmt.Errorf("compose workbook contribution catalog: projection descriptors are required")
	}
	if isNilDependency(input.ProjectionQueries) {
		return fmt.Errorf("compose workbook contribution catalog: projection queries are required")
	}
	if isNilDependency(input.EntityProjections.Writer) ||
		isNilDependency(input.EntityProjections.Reader) {
		return fmt.Errorf("compose workbook contribution catalog: Entities projection ports are required")
	}
	if isNilDependency(input.AssessmentProjections) {
		return fmt.Errorf("compose workbook contribution catalog: Assessments projection rows are required")
	}
	if isNilDependency(input.PartyProjections) {
		return fmt.Errorf("compose workbook contribution catalog: Parties projection rows are required")
	}
	if input.IndicatorOwner == nil {
		return fmt.Errorf("compose workbook contribution catalog: Indicators owner is required")
	}
	if isNilDependency(input.TimelineOwner) {
		return fmt.Errorf("compose workbook contribution catalog: Timeline owner is required")
	}
	if isNilDependency(input.EvidenceOwner) {
		return fmt.Errorf("compose workbook contribution catalog: Evidence contribution is required")
	}
	if input.ArtifactOwner == nil {
		return fmt.Errorf("compose workbook contribution catalog: Artifacts mutation contribution is required")
	}
	if input.TaskDecisionOwner == nil {
		return fmt.Errorf("compose workbook contribution catalog: Tasks/Decisions mutation contribution is required")
	}
	if reflect.ValueOf(input.ConflictTokens).IsZero() {
		return fmt.Errorf("compose workbook contribution catalog: conflict token codec is required")
	}
	if isNilDependency(input.ConflictFields) {
		return fmt.Errorf("compose workbook contribution catalog: Revisions conflict field resolver is required")
	}
	if input.Revisions == nil {
		return fmt.Errorf("compose workbook contribution catalog: Revisions appender is required")
	}
	if isNilDependency(input.CollaborationIntents) {
		return fmt.Errorf("compose workbook contribution catalog: Collaboration intent appender is required")
	}
	return nil
}

func isNilDependency(dependency any) bool {
	if dependency == nil {
		return true
	}
	value := reflect.ValueOf(dependency)
	switch value.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return value.IsNil()
	default:
		return false
	}
}

type contributionAssemblyInput struct {
	pool                  postgres.DB
	projectionDescriptors providercontract.DescriptorSet
	projectionQueries     projectionQueryCatalog
	entityStore           *hostidentity.Store
	assessmentProjections assessmentprojection.Rows
	partyProjections      partyprojection.Rows
	indicatorOwner        *indicators.Application
	timelineOwner         TimelineOperations
	evidenceOwner         evidence.MutationContribution
	artifactOwner         *artifacts.MutationFacade
	taskDecisionOwner     *tasksdecisions.MutationFacade
	conflictTokens        conflicttokens.ConflictTokenCodec
	conflictFields        conflicttokens.FieldResolver
	appender              *revisions.Appender
	collaborationIntents  collaboration.RecordChangedAppender
}

func buildContributionCatalog(input contributionAssemblyInput) (*workbook.WorkbookContributionCatalog, error) {
	pool := input.pool
	projectionDescriptors := input.projectionDescriptors
	projectionQueries := input.projectionQueries
	entityStore := input.entityStore
	assessmentProjections := input.assessmentProjections
	partyProjections := input.partyProjections
	indicatorOwner := input.indicatorOwner
	timelineOwner := input.timelineOwner
	evidenceOwner := input.evidenceOwner
	artifactOwner := input.artifactOwner
	taskDecisionOwner := input.taskDecisionOwner
	conflictTokens := input.conflictTokens
	conflictFields := input.conflictFields
	appender := input.appender

	entityProviders, err := newEntityProviderSet(entityStore)
	if err != nil {
		return nil, fmt.Errorf("compose workbook contribution catalog: %w", err)
	}
	assessmentFacade, err := NewAssessmentMutationContribution(
		pool,
		assessmentProjections,
		hostidentity.NewSourceFacts(),
		appender,
		input.collaborationIntents,
	)
	if err != nil {
		return nil, fmt.Errorf("compose workbook contribution catalog: %w", err)
	}
	assessmentCreateProvider, err := newAssessmentCreateProvider(assessmentFacade)
	if err != nil {
		return nil, fmt.Errorf("compose workbook contribution catalog: %w", err)
	}
	partyOwner, err := NewPartyMutationContribution(
		pool,
		conflictTokens,
		appender,
		conflictFields,
		partyProjections,
		input.collaborationIntents,
	)
	if err != nil {
		return nil, fmt.Errorf("compose workbook contribution catalog: %w", err)
	}
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
	hostQueryProvider, err := workbook.NewQueryProvider(
		func(ctx context.Context, command workbook.QueryCommand) (querypage.Result, error) {
			if command.ViewSchemaID != hostidentity.HostsViewSchemaID {
				return querypage.Result{}, fmt.Errorf(
					"host query contribution received view schema %q",
					command.ViewSchemaID,
				)
			}
			return entityStore.QueryHostRowsPage(ctx, command.IncidentID, command.Query, command.Window)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("compose workbook Hosts query provider: %w", err)
	}
	identityQueryProvider, err := workbook.NewQueryProvider(
		func(ctx context.Context, command workbook.QueryCommand) (querypage.Result, error) {
			if command.ViewSchemaID != hostidentity.IdentitiesViewSchemaID {
				return querypage.Result{}, fmt.Errorf(
					"identity query contribution received view schema %q",
					command.ViewSchemaID,
				)
			}
			return entityStore.QueryIdentityRowsPage(ctx, command.IncidentID, command.Query, command.Window)
		},
	)
	if err != nil {
		return nil, fmt.Errorf("compose workbook Identities query provider: %w", err)
	}
	sourceQueries := map[string]workbook.QueryProvider{
		hostidentity.HostsViewSchemaID:      hostQueryProvider,
		hostidentity.IdentitiesViewSchemaID: identityQueryProvider,
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
	catalog, err := workbook.NewWorkbookContributionCatalog(workbook.ContributionCatalogInput{
		ProjectionDescriptors: projectionDescriptors,
		Queries:               queryContributions,
		Creates:               createContributions,
		Patches:               patchContributions,
		Conflicts:             conflictContributions,
		ActionRequirements: workbook.ActionCapabilityRequirements{
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
		Actions: actionContributions,
	})
	if err != nil {
		return nil, fmt.Errorf("compose workbook contribution catalog: %w", err)
	}
	return catalog, nil
}
