package timelineassembly

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/app/assessmentassembly"
	"github.com/JochiRaider/cartulary/internal/app/projectionassembly"
	artifactprovider "github.com/JochiRaider/cartulary/internal/modules/artifacts/projectionprovider"
	artifactprojection "github.com/JochiRaider/cartulary/internal/modules/artifacts/workbookprojection"
	assessmentprojection "github.com/JochiRaider/cartulary/internal/modules/assessments/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity"
	entityprovider "github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity/projectionprovider"
	"github.com/JochiRaider/cartulary/internal/modules/entities/mentions"
	"github.com/JochiRaider/cartulary/internal/modules/entities/merge"
	entityprojection "github.com/JochiRaider/cartulary/internal/modules/entities/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	evidenceprojection "github.com/JochiRaider/cartulary/internal/modules/evidence/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	indicatorprovider "github.com/JochiRaider/cartulary/internal/modules/indicators/projectionprovider"
	indicatorprojection "github.com/JochiRaider/cartulary/internal/modules/indicators/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	partyprovider "github.com/JochiRaider/cartulary/internal/modules/parties/projectionprovider"
	partyprojection "github.com/JochiRaider/cartulary/internal/modules/parties/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/recovery/restorecontract"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	taskdecisionprovider "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/projectionprovider"
	taskdecisionprojection "github.com/JochiRaider/cartulary/internal/modules/tasksdecisions/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/mentioneffects"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/sourcerepository"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/workbookprojection"
	workbookrestoreprobe "github.com/JochiRaider/cartulary/internal/modules/workbook/restoreprobe"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
	"github.com/JochiRaider/cartulary/internal/platform/workbookprobe"
)

type projectionSourceTextRows interface {
	RefreshRowTx(context.Context, pgx.Tx, string, uuid.UUID) error
	LoadRowTx(context.Context, pgx.Tx, string, uuid.UUID) (map[string]any, error)
}

type Bundle struct {
	Facade                   *timeline.Facade
	ProjectionSource         *workbookprojection.Source
	MentionEffects           *mentioneffects.Provider
	EntityMentionStore       *mentions.Store
	EntityMergeStore         *merge.Store
	Projections              *projectionassembly.Bundle
	ProjectionSourceTextRows projectionSourceTextRows
	TimelineProjections      workbookprojection.Ports
	EntityProjections        entityprojection.Ports
	IndicatorProjections     indicatorprojection.Ports
	AssessmentProjections    assessmentprojection.Ports
	ArtifactProjections      artifactprojection.Ports
	EvidenceProjections      evidenceprojection.Ports
	PartyProjections         partyprojection.Ports
	TaskDecisionProjections  taskdecisionprojection.Ports
	RestoreRebuilder         restorecontract.ProjectionRebuilder
	Collaborators            timeline.Collaborators
}

type composition struct {
	projectionSource         *workbookprojection.Source
	mentionEffects           *mentioneffects.Provider
	entityMentionStore       *mentions.Store
	entityMergeStore         *merge.Store
	projections              *projectionassembly.Bundle
	projectionSourceTextRows projectionSourceTextRows
	timelineProjections      workbookprojection.Ports
	entityProjections        entityprojection.Ports
	indicatorProjections     indicatorprojection.Ports
	assessmentProjections    assessmentprojection.Ports
	artifactProjections      artifactprojection.Ports
	evidenceProjections      evidenceprojection.Ports
	partyProjections         partyprojection.Ports
	taskDecisionProjections  taskdecisionprojection.Ports
	restoreRebuilder         restorecontract.ProjectionRebuilder
	collaborators            timeline.Collaborators
}

type ProjectionBundle struct {
	Source           *workbookprojection.Source
	Projections      *projectionassembly.Bundle
	SourceTextRows   projectionSourceTextRows
	Timeline         workbookprojection.Ports
	Entities         entityprojection.Ports
	Indicators       indicatorprojection.Ports
	Assessments      assessmentprojection.Ports
	Artifacts        artifactprojection.Ports
	Evidence         evidenceprojection.Ports
	Parties          partyprojection.Ports
	TasksDecisions   taskdecisionprojection.Ports
	AssessmentSource assessmentprojection.SourceReader
	Recovery         restorecontract.ProjectionPorts
	Revisions        revisions.ProjectionServices
	RestoreQuery     workbookrestoreprobe.ProjectionQuery
}

func NewBundle(
	pool postgres.DB,
	conflictTokens conflicttokens.ConflictTokenCodec,
	appender *revisions.Appender,
	intents collaboration.IntentAppender,
	evidenceAttachments evidence.TimelineAttachmentContribution,
) *Bundle {
	return NewBundleWithProjection(
		pool,
		conflictTokens,
		appender,
		intents,
		NewProjectionBundle(pool),
		evidenceAttachments,
	)
}

func NewBundleWithProjection(
	pool postgres.DB,
	conflictTokens conflicttokens.ConflictTokenCodec,
	appender *revisions.Appender,
	intents collaboration.IntentAppender,
	projection *ProjectionBundle,
	evidenceAttachments evidence.TimelineAttachmentContribution,
) *Bundle {
	if appender == nil {
		panic("compose Timeline bundle: Revisions appender is required")
	}
	if intents == nil {
		panic("compose Timeline bundle: Collaboration intent appender is required")
	}
	if projection == nil {
		panic("compose Timeline bundle: validated projection bundle is required")
	}
	if evidenceAttachments == nil {
		panic("compose Timeline bundle: Evidence attachment contribution is required")
	}
	components := compose(pool, appender, intents, projection, evidenceAttachments)
	return &Bundle{
		Facade:                   timeline.NewFacade(pool, components.collaborators, conflictTokens),
		ProjectionSource:         components.projectionSource,
		MentionEffects:           components.mentionEffects,
		EntityMentionStore:       components.entityMentionStore,
		EntityMergeStore:         components.entityMergeStore,
		Projections:              components.projections,
		ProjectionSourceTextRows: components.projectionSourceTextRows,
		TimelineProjections:      components.timelineProjections,
		EntityProjections:        components.entityProjections,
		IndicatorProjections:     components.indicatorProjections,
		AssessmentProjections:    components.assessmentProjections,
		ArtifactProjections:      components.artifactProjections,
		EvidenceProjections:      components.evidenceProjections,
		PartyProjections:         components.partyProjections,
		TaskDecisionProjections:  components.taskDecisionProjections,
		RestoreRebuilder:         components.restoreRebuilder,
		Collaborators:            components.collaborators,
	}
}

// NewCollaborators composes Timeline's typed application boundary for focused
// facade tests that replace one collaborator without starting a server.
func NewCollaborators(
	pool postgres.DB,
	appender *revisions.Appender,
	intents collaboration.IntentAppender,
	evidenceAttachments evidence.TimelineAttachmentContribution,
) timeline.Collaborators {
	if appender == nil {
		panic("compose Timeline collaborators: Revisions appender is required")
	}
	if intents == nil {
		panic("compose Timeline collaborators: Collaboration intent appender is required")
	}
	if evidenceAttachments == nil {
		panic("compose Timeline collaborators: Evidence attachment contribution is required")
	}
	return compose(pool, appender, intents, NewProjectionBundle(pool), evidenceAttachments).collaborators
}

func NewRestoreRebuilder(pool postgres.DB) restorecontract.ProjectionRebuilder {
	return NewProjectionBundle(pool).Recovery.Rebuilder
}

func NewRecoveryProjectionServices(pool postgres.DB) (restorecontract.ProjectionRebuilder, workbookprobe.Executor, error) {
	components := NewProjectionBundle(pool)
	registry, err := workbookrestoreprobe.NewRegistry(
		components.RestoreQuery,
		timeline.RestoreWorkbookProbeRegistration(),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("compose restore workbook probe registry: %w", err)
	}
	return components.Recovery.Rebuilder, registry, nil
}

func compose(
	pool postgres.DB,
	appender *revisions.Appender,
	intents collaboration.IntentAppender,
	projection *ProjectionBundle,
	evidenceAttachments evidence.TimelineAttachmentContribution,
) composition {
	recordsPort := recordAdapter{
		store:   records.NewStore(),
		targets: records.NewRouteTargetResolver(pool),
	}
	collectionFacts := newCollectionReadAdapter()
	timelineWriter := projection.Timeline.Writer
	entityProjectionWriter := projection.Entities.Writer
	mentionEffects := mentioneffects.NewProvider(recordsPort, collectionFacts, timelineWriter)
	collaborators := timeline.Collaborators{
		Core: timeline.CoreCollaborators{
			Idempotency: idempotencyAdapter{store: authn.NewStore(pool)},
			Incidents:   incidentAdapter{access: incidents.NewAccess(pool)},
			Records:     recordsPort,
			Revisions:   revisionAdapter{appender: appender, reader: conflicttokens.NewRevisionWindowReader()},
		},
		Collections: timeline.CollectionCollaborators{
			Links: linkAdapter{store: links.NewStore()},
			Mentions: mentionAdapter{store: mentions.NewStore(
				nil,
				appender,
				mentions.WithWorkbookProjection(entityProjectionWriter),
			)},
			Entities: entityAdapter{store: hostidentity.NewStore(
				pool,
				appender,
				nil,
				entityProjectionWriter,
			)},
			Evidence: evidenceAdapter{attachments: evidenceAttachments},
			Facts:    collectionFacts,
		},
		Commit: timeline.CommitCollaborators{
			Projection:       timelineWriter,
			EntityProjection: entityProjectionWriter,
			Collaboration:    collaborationAdapter{appender: intents},
		},
	}
	return composition{
		projectionSource: projection.Source,
		mentionEffects:   mentionEffects,
		entityMentionStore: mentions.NewStore(
			pool,
			appender,
			mentions.WithTimelineEffects(mentionEffects),
			mentions.WithCollaborationIntents(intents),
			mentions.WithWorkbookProjection(entityProjectionWriter),
		),
		entityMergeStore: merge.NewStore(
			pool,
			appender,
			merge.WithAssessmentEffects(assessmentassembly.NewMergeEffects(
				projection.Assessments.Rows,
			)),
			merge.WithTimelineEffects(mentionEffects),
			merge.WithCollaborationIntents(intents),
			merge.WithWorkbookProjection(entityProjectionWriter),
		),
		projections:              projection.Projections,
		projectionSourceTextRows: projection.SourceTextRows,
		timelineProjections:      projection.Timeline,
		entityProjections:        projection.Entities,
		indicatorProjections:     projection.Indicators,
		assessmentProjections:    projection.Assessments,
		artifactProjections:      projection.Artifacts,
		evidenceProjections:      projection.Evidence,
		partyProjections:         projection.Parties,
		taskDecisionProjections:  projection.TasksDecisions,
		restoreRebuilder:         projection.Recovery.Rebuilder,
		collaborators:            collaborators,
	}
}

func NewProjectionBundle(pool postgres.DB) *ProjectionBundle {
	recordsPort := recordAdapter{
		store:   records.NewStore(),
		targets: records.NewRouteTargetResolver(pool),
	}
	collectionFacts := newCollectionReadAdapter()
	source := workbookprojection.NewSource(recordsPort, collectionFacts)
	timelineContribution, err := workbookprojection.NewRuntimeContribution(source)
	if err != nil {
		panic(fmt.Sprintf("compose Timeline projection contribution: %v", err))
	}
	entitiesContribution, err := entityprojection.NewRuntimeContribution(entityprovider.NewSource())
	if err != nil {
		panic(fmt.Sprintf("compose Entities projection contribution: %v", err))
	}
	indicatorsContribution, err := indicatorprovider.NewContribution()
	if err != nil {
		panic(fmt.Sprintf("compose Indicators projection contribution: %v", err))
	}
	assessmentsContribution, err := assessmentassembly.NewProjectionContribution()
	if err != nil {
		panic(fmt.Sprintf("compose Assessments projection contribution: %v", err))
	}
	artifactsContribution, err := artifactprovider.NewContribution()
	if err != nil {
		panic(fmt.Sprintf("compose Artifacts projection contribution: %v", err))
	}
	evidenceContribution, err := projectionassembly.NewEvidenceContribution()
	if err != nil {
		panic(fmt.Sprintf("compose Evidence projection contribution: %v", err))
	}
	partiesContribution, err := partyprovider.NewContribution()
	if err != nil {
		panic(fmt.Sprintf("compose Parties projection contribution: %v", err))
	}
	taskDecisionContribution, err := taskdecisionprovider.NewContribution()
	if err != nil {
		panic(fmt.Sprintf("compose Tasks/Decisions projection contribution: %v", err))
	}
	catalog, err := projectionassembly.NewBundle(
		pool,
		timelineContribution,
		entitiesContribution,
		indicatorsContribution,
		assessmentsContribution,
		artifactsContribution,
		evidenceContribution,
		partiesContribution,
		taskDecisionContribution,
	)
	if err != nil {
		panic(fmt.Sprintf("compose projection catalog: %v", err))
	}
	recoveryPorts := catalog.RecoveryPorts()
	if !recoveryPorts.Ready() {
		panic("compose projection catalog: Recovery ports are incomplete")
	}
	timelinePorts := catalog.TimelinePorts()
	if !timelinePorts.Ready() {
		panic("compose projection catalog: Timeline ports are incomplete")
	}
	entityPorts := catalog.EntityPorts()
	if !entityPorts.Ready() {
		panic("compose projection catalog: Entities ports are incomplete")
	}
	indicatorPorts := catalog.IndicatorPorts()
	if !indicatorPorts.Ready() {
		panic("compose projection catalog: Indicators ports are incomplete")
	}
	assessmentPorts := catalog.AssessmentPorts()
	if !assessmentPorts.Ready() {
		panic("compose projection catalog: Assessments ports are incomplete")
	}
	artifactPorts := catalog.ArtifactPorts()
	if !artifactPorts.Ready() {
		panic("compose projection catalog: Artifacts ports are incomplete")
	}
	evidencePorts := catalog.EvidencePorts()
	if !evidencePorts.Ready() {
		panic("compose projection catalog: Evidence ports are incomplete")
	}
	partyPorts := catalog.PartyPorts()
	if !partyPorts.Ready() {
		panic("compose projection catalog: Parties ports are incomplete")
	}
	taskDecisionPorts := catalog.TaskDecisionPorts()
	if !taskDecisionPorts.Ready() {
		panic("compose projection catalog: Tasks/Decisions ports are incomplete")
	}
	return &ProjectionBundle{
		Source:           source,
		Projections:      catalog,
		SourceTextRows:   catalog.SourceTextRows(),
		Timeline:         timelinePorts,
		Entities:         entityPorts,
		Indicators:       indicatorPorts,
		Assessments:      assessmentPorts,
		Artifacts:        artifactPorts,
		Evidence:         evidencePorts,
		Parties:          partyPorts,
		TasksDecisions:   taskDecisionPorts,
		AssessmentSource: assessmentsContribution.Source(),
		Recovery:         recoveryPorts,
		Revisions:        catalog.RevisionServices(),
		RestoreQuery:     catalog.RestoreProbeQuery(),
	}
}

func (bundle *ProjectionBundle) EvidenceProjectionPort() evidenceprojection.Rows {
	if bundle == nil {
		return nil
	}
	return bundle.Evidence.Rows
}

type collaborationAdapter struct {
	appender collaboration.IntentAppender
}

func (a collaborationAdapter) AppendRecordChangeIntentTx(ctx context.Context, tx pgx.Tx, params timeline.RecordChangeIntentParams) error {
	intent, err := collaboration.NewRecordChangeIntent(collaboration.RecordChange{
		IncidentID:       params.IncidentID,
		RecordID:         params.RecordID,
		RowVersion:       params.RowVersion,
		ChangeSetID:      params.ChangeSetID,
		ClientTxnID:      params.ClientTxnID,
		ActorUserID:      params.ActorUserID,
		ChangedFieldKeys: params.ChangedFieldKeys,
		ViewSchemaID:     params.ViewSchemaID,
		ChangeKind:       params.ChangeKind,
		Row:              params.Row,
		PatchCells:       params.PatchCells,
	}, params.MutationOrdinal, params.CreatedAt)
	if err != nil {
		return err
	}
	return a.appender.AppendIntentTx(ctx, tx, intent)
}

type idempotencyAdapter struct {
	store *authn.Store
}

func (a idempotencyAdapter) GetRouteIdempotency(ctx context.Context, key authn.RouteIdempotencyKey) (authn.RouteIdempotencyRecord, error) {
	return a.store.GetRouteIdempotency(ctx, key)
}

func (a idempotencyAdapter) InsertRouteIdempotencyPayload(ctx context.Context, tx pgx.Tx, key authn.RouteIdempotencyKey, targetUserID *uuid.UUID, requestHash []byte, statusCode int, payload any) error {
	return authn.InsertRouteIdempotencyPayload(ctx, tx, key, targetUserID, requestHash, statusCode, payload)
}

type incidentAdapter struct {
	access incidents.Access
}

func (a incidentAdapter) EnsureOpenTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	err := a.access.EnsureOpenTx(ctx, tx, incidentID)
	if errors.Is(err, incidents.ErrIncidentClosed) {
		return timeline.ErrIncidentClosed
	}
	return err
}

type recordAdapter struct {
	store   *records.Store
	targets *records.RouteTargetResolver
}

func (a recordAdapter) InsertTx(ctx context.Context, tx pgx.Tx, params timeline.RecordCreateParams) (uuid.UUID, error) {
	return a.store.InsertTx(ctx, tx, records.InsertParams(params))
}

func (a recordAdapter) AdvanceVersionTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, actorUserID uuid.UUID, now time.Time) (int64, error) {
	return a.store.AdvanceVersionTx(ctx, tx, recordID, actorUserID, now)
}

func (a recordAdapter) LoadRowVersionTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (int64, error) {
	return a.store.LoadRowVersionTx(ctx, tx, recordID)
}

func (a recordAdapter) LoadEnvelopeTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, lock bool) (sourcerepository.Envelope, error) {
	envelope, err := a.store.LoadEnvelopeTx(ctx, tx, recordID, lock)
	if errors.Is(err, records.ErrEnvelopeNotFound) {
		return sourcerepository.Envelope{}, sourcerepository.ErrEnvelopeNotFound
	}
	return timelineEnvelope(envelope), err
}

func (a recordAdapter) LoadEnvelopesTx(ctx context.Context, tx pgx.Tx, recordIDs []uuid.UUID, lock bool) (map[uuid.UUID]sourcerepository.Envelope, error) {
	envelopes, err := a.store.LoadEnvelopesTx(ctx, tx, recordIDs, lock)
	if err != nil {
		return nil, err
	}
	result := make(map[uuid.UUID]sourcerepository.Envelope, len(envelopes))
	for recordID, envelope := range envelopes {
		result[recordID] = timelineEnvelope(envelope)
	}
	return result, nil
}

func timelineEnvelope(envelope records.Envelope) sourcerepository.Envelope {
	return sourcerepository.Envelope{
		RecordID:        envelope.RecordID,
		IncidentID:      envelope.IncidentID,
		RecordType:      envelope.RecordType,
		RowVersion:      envelope.RowVersion,
		CreatedByUserID: envelope.CreatedByUserID,
		CreatedAt:       envelope.CreatedAt,
		UpdatedByUserID: envelope.UpdatedByUserID,
		UpdatedAt:       envelope.UpdatedAt,
		DeletedAt:       envelope.DeletedAt,
	}
}

func (a recordAdapter) ResolveIncident(ctx context.Context, recordID uuid.UUID) (uuid.UUID, error) {
	return a.targets.ResolveIncident(ctx, recordID)
}

type revisionAdapter struct {
	appender *revisions.Appender
	reader   conflicttokens.RevisionWindowReader
}

func (a revisionAdapter) AppendChangeSetTx(ctx context.Context, tx pgx.Tx, params timeline.ChangeSetParams) (uuid.UUID, error) {
	return a.appender.AppendChangeSetTx(ctx, tx, revisions.AppendChangeSetParams(params))
}

func (a revisionAdapter) AppendMutationTx(ctx context.Context, tx pgx.Tx, params timeline.MutationParams) error {
	return a.appender.AppendMutationTx(ctx, tx, revisions.AppendMutationParams(params))
}

func (a revisionAdapter) AppendRecordRevisionTx(ctx context.Context, tx pgx.Tx, params timeline.RecordRevisionParams) error {
	return a.appender.AppendRecordRevisionOnlyTx(ctx, tx, revisions.AppendRecordRevisionParams(params))
}

func (a revisionAdapter) ListRecordRevisionWindowTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, firstVersion int64, lastVersion int64) ([]timeline.RecordRevisionWindowEntry, error) {
	entries, err := a.reader.LoadRevisionWindowTx(ctx, tx, recordID, firstVersion, lastVersion)
	if err != nil {
		return nil, err
	}
	result := make([]timeline.RecordRevisionWindowEntry, 0, len(entries))
	for _, entry := range entries {
		result = append(result, timeline.RecordRevisionWindowEntry(entry))
	}
	return result, nil
}

type linkAdapter struct {
	store *links.Store
}

func (a linkAdapter) InsertSupersedesCommandTx(ctx context.Context, tx pgx.Tx, command timeline.InsertSupersedesCommand) (timeline.SupersedesLink, error) {
	link, err := a.store.InsertSupersedesCommandTx(ctx, tx, links.InsertSupersedesCommand(command))
	return timeline.SupersedesLink(link), err
}

func (a linkAdapter) UpsertLinkCommandTx(ctx context.Context, tx pgx.Tx, command timeline.UpsertLinkCommand) error {
	_, _, err := a.store.UpsertLinkCommandTx(ctx, tx, links.UpsertLinkCommand{
		IncidentID:  command.IncidentID,
		SrcRecordID: command.SrcRecordID,
		DstRecordID: command.DstRecordID,
		LinkType:    links.LinkType(command.LinkType),
		Provenance:  links.LinkProvenance(command.Provenance),
		Confidence:  command.Confidence,
		OwnerUserID: command.OwnerUserID,
		Now:         command.Now,
	})
	return err
}

func (a linkAdapter) HasActiveIncomingSupersedesLinkForUpdateTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordID uuid.UUID) (bool, error) {
	return a.store.HasActiveIncomingSupersedesLinkForUpdateTx(ctx, tx, incidentID, recordID)
}

func (a linkAdapter) LoadRecordLinkValueTx(ctx context.Context, tx pgx.Tx, recordLinkID uuid.UUID) (map[string]any, error) {
	return a.store.LoadRecordLinkValueTx(ctx, tx, recordLinkID)
}

func (a linkAdapter) ApplyRecordRefCollectionWithMutationValuesTx(ctx context.Context, tx pgx.Tx, command timeline.RecordRefCollectionCommand) (timeline.CollectionMutationResult, error) {
	result, err := a.store.ApplyRecordRefCollectionWithMutationValuesTx(ctx, tx, links.RecordRefCollectionCommand{
		IncidentID:         command.IncidentID,
		SourceRecordID:     command.SourceRecordID,
		ActorUserID:        command.ActorUserID,
		FieldKey:           command.FieldKey,
		LinkType:           links.LinkType(command.LinkType),
		ExpectedTargetType: command.ExpectedTargetType,
		AddRecordIDs:       command.AddRecordIDs,
		RemoveRecordIDs:    command.RemoveRecordIDs,
		Now:                command.Now,
	})
	return collectionResult(result), err
}

func (a linkAdapter) ApplyTagCollectionWithMutationValuesTx(ctx context.Context, tx pgx.Tx, command timeline.TagCollectionCommand) (timeline.CollectionMutationResult, error) {
	adds := make([]links.TagCollectionAdd, 0, len(command.AddTags))
	for _, add := range command.AddTags {
		adds = append(adds, links.TagCollectionAdd(add))
	}
	removes := make([]links.RecordTagRef, 0, len(command.RemoveTags))
	for _, remove := range command.RemoveTags {
		removes = append(removes, links.RecordTagRef(remove))
	}
	result, err := a.store.ApplyTagCollectionWithMutationValuesTx(ctx, tx, links.TagCollectionCommand{
		IncidentID:  command.IncidentID,
		RecordID:    command.RecordID,
		ActorUserID: command.ActorUserID,
		FieldKey:    command.FieldKey,
		AddTags:     adds,
		RemoveTags:  removes,
		Now:         command.Now,
	})
	return collectionResult(result), err
}

func collectionResult(result links.CollectionMutationResult) timeline.CollectionMutationResult {
	converted := timeline.CollectionMutationResult{
		RecordLinks: make([]timeline.RecordLinkMutation, 0, len(result.RecordLinks)),
		RecordTags:  make([]timeline.RecordTagMutation, 0, len(result.RecordTags)),
	}
	for _, mutation := range result.RecordLinks {
		converted.RecordLinks = append(converted.RecordLinks, timeline.RecordLinkMutation(mutation))
	}
	for _, mutation := range result.RecordTags {
		converted.RecordTags = append(converted.RecordTags, timeline.RecordTagMutation(mutation))
	}
	return converted
}

type mentionAdapter struct {
	store *mentions.Store
}

func (a mentionAdapter) ResolveExistingFromMentionTx(ctx context.Context, tx pgx.Tx, actor authn.UserRecord, sourceRecordID uuid.UUID, fieldKey string, mentionID uuid.UUID, resolvedRecordID *uuid.UUID, now time.Time) error {
	_, err := a.store.ResolveExistingFromMentionTx(ctx, tx, actor, sourceRecordID, fieldKey, mentionID, resolvedRecordID, now)
	if errors.Is(err, mentions.ErrResolvedRecordNotFound) {
		return timeline.ErrResolvedRecordNotFound
	}
	return err
}

func (a mentionAdapter) ApplyMentionLifecycleTx(ctx context.Context, tx pgx.Tx, actor authn.UserRecord, sourceRecordID uuid.UUID, sourceFieldKey string, mentionID uuid.UUID, action string, resolvedRecordID *uuid.UUID, now time.Time) error {
	err := a.store.ApplyMentionLifecycleTx(ctx, tx, actor, sourceRecordID, sourceFieldKey, mentionID, action, resolvedRecordID, now)
	if errors.Is(err, mentions.ErrResolvedRecordNotFound) {
		return timeline.ErrResolvedRecordNotFound
	}
	return err
}

func (a mentionAdapter) NextOrdinalTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, fieldKey string) (int, error) {
	return a.store.NextOrdinalTx(ctx, tx, recordID, fieldKey)
}

func (a mentionAdapter) InsertTx(ctx context.Context, tx pgx.Tx, params timeline.MentionCreateParams) error {
	return a.store.InsertTx(ctx, tx, mentions.CreateParams(params))
}

type entityAdapter struct {
	store *hostidentity.Store
}

func (a entityAdapter) ListEligibleAliasesTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, entityType string) ([]timeline.EntityAlias, error) {
	aliases, err := a.store.ListEligibleAliasesTx(ctx, tx, incidentID, entityType)
	if err != nil {
		return nil, err
	}
	result := make([]timeline.EntityAlias, 0, len(aliases))
	for _, alias := range aliases {
		result = append(result, timeline.EntityAlias(alias))
	}
	return result, nil
}

func (a entityAdapter) ValidateResolvedTargetTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, entityType string, recordID uuid.UUID) error {
	err := a.store.ValidateResolvedTargetTx(ctx, tx, incidentID, entityType, recordID)
	if errors.Is(err, hostidentity.ErrHostIdentityRecordNotFound) {
		return timeline.ErrResolvedRecordNotFound
	}
	return err
}

type evidenceAdapter struct {
	attachments evidence.TimelineAttachmentContribution
}

func (a evidenceAdapter) ValidateTimelineAttachmentsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordIDs []uuid.UUID) error {
	err := a.attachments.ValidateTimelineAttachmentsTx(ctx, tx, incidentID, recordIDs)
	if errors.Is(err, evidence.ErrEvidenceNotFound) {
		return &links.CollectionValidationError{
			Field:      "timeline.attached_evidence_ids",
			ReasonCode: "invalid_value",
		}
	}
	return err
}
