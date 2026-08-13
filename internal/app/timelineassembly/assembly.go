package timelineassembly

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/app/assessmentassembly"
	assessmentprojection "github.com/JochiRaider/cartulary/internal/modules/assessments/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity"
	"github.com/JochiRaider/cartulary/internal/modules/entities/mentions"
	"github.com/JochiRaider/cartulary/internal/modules/entities/merge"
	entityprojection "github.com/JochiRaider/cartulary/internal/modules/entities/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/mentioneffects"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/sourcerepository"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type Dependencies struct {
	Postgres            postgres.DB
	ConflictTokens      conflicttokens.ConflictTokenCodec
	Revisions           *revisions.Appender
	Collaboration       collaboration.IntentAppender
	EvidenceAttachments evidence.TimelineAttachmentContribution
	TimelineProjection  workbookprojection.Writer
	EntityProjection    entityprojection.Writer
	AssessmentRows      assessmentprojection.Rows
}

type Bundle struct {
	Facade             *timeline.Facade
	MentionEffects     *mentioneffects.Provider
	EntityMentionStore *mentions.Store
	EntityMergeStore   *merge.Store
	Collaborators      timeline.Collaborators
}

type composition struct {
	mentionEffects     *mentioneffects.Provider
	entityMentionStore *mentions.Store
	entityMergeStore   *merge.Store
	collaborators      timeline.Collaborators
}

func NewBundle(dependencies Dependencies) (*Bundle, error) {
	if dependencies.Postgres == nil {
		return nil, fmt.Errorf("compose Timeline bundle: Postgres is required")
	}
	if dependencies.Revisions == nil {
		return nil, fmt.Errorf("compose Timeline bundle: Revisions appender is required")
	}
	if dependencies.Collaboration == nil {
		return nil, fmt.Errorf("compose Timeline bundle: Collaboration intent appender is required")
	}
	if dependencies.EvidenceAttachments == nil {
		return nil, fmt.Errorf("compose Timeline bundle: Evidence attachment contribution is required")
	}
	if dependencies.TimelineProjection == nil {
		return nil, fmt.Errorf("compose Timeline bundle: Timeline projection writer is required")
	}
	if dependencies.EntityProjection == nil {
		return nil, fmt.Errorf("compose Timeline bundle: Entities projection writer is required")
	}
	if dependencies.AssessmentRows == nil {
		return nil, fmt.Errorf("compose Timeline bundle: Assessment projection rows are required")
	}
	components := compose(dependencies)
	return &Bundle{
		Facade:             timeline.NewFacade(dependencies.Postgres, components.collaborators, dependencies.ConflictTokens),
		MentionEffects:     components.mentionEffects,
		EntityMentionStore: components.entityMentionStore,
		EntityMergeStore:   components.entityMergeStore,
		Collaborators:      components.collaborators,
	}, nil
}

// NewCollaborators composes Timeline's typed application boundary for focused
// facade tests that replace one collaborator without starting a server.
func NewCollaborators(dependencies Dependencies) (timeline.Collaborators, error) {
	bundle, err := NewBundle(dependencies)
	if err != nil {
		return timeline.Collaborators{}, err
	}
	return bundle.Collaborators, nil
}

func compose(dependencies Dependencies) composition {
	recordsPort := recordAdapter{
		store:   records.NewStore(),
		targets: records.NewRouteTargetResolver(dependencies.Postgres),
	}
	collectionFacts := newCollectionReadAdapter()
	timelineWriter := dependencies.TimelineProjection
	entityProjectionWriter := dependencies.EntityProjection
	mentionEffects := mentioneffects.NewProvider(recordsPort, collectionFacts, timelineWriter)
	collaborators := timeline.Collaborators{
		Core: timeline.CoreCollaborators{
			Idempotency: idempotencyAdapter{store: authn.NewStore(dependencies.Postgres)},
			Incidents:   incidentAdapter{access: incidents.NewAccess(dependencies.Postgres)},
			Records:     recordsPort,
			Revisions:   revisionAdapter{appender: dependencies.Revisions, reader: conflicttokens.NewRevisionWindowReader()},
		},
		Collections: timeline.CollectionCollaborators{
			Links: linkAdapter{store: links.NewStore()},
			Mentions: mentionAdapter{store: mentions.NewStore(
				nil,
				dependencies.Revisions,
				mentions.WithWorkbookProjection(entityProjectionWriter),
			)},
			Entities: entityAdapter{store: hostidentity.NewStore(
				dependencies.Postgres,
				dependencies.Revisions,
				nil,
				entityProjectionWriter,
			)},
			Evidence: evidenceAdapter{attachments: dependencies.EvidenceAttachments},
			Facts:    collectionFacts,
		},
		Commit: timeline.CommitCollaborators{
			Projection:       timelineWriter,
			EntityProjection: entityProjectionWriter,
			Collaboration:    collaborationAdapter{appender: dependencies.Collaboration},
		},
	}
	return composition{
		mentionEffects: mentionEffects,
		entityMentionStore: mentions.NewStore(
			dependencies.Postgres,
			dependencies.Revisions,
			mentions.WithTimelineEffects(mentionEffects),
			mentions.WithCollaborationIntents(dependencies.Collaboration),
			mentions.WithWorkbookProjection(entityProjectionWriter),
		),
		entityMergeStore: merge.NewStore(
			dependencies.Postgres,
			dependencies.Revisions,
			merge.WithAssessmentEffects(assessmentassembly.NewMergeEffects(
				dependencies.AssessmentRows,
				dependencies.Revisions,
			)),
			merge.WithTimelineEffects(mentionEffects),
			merge.WithCollaborationIntents(dependencies.Collaboration),
			merge.WithWorkbookProjection(entityProjectionWriter),
		),
		collaborators: collaborators,
	}
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

func (a revisionAdapter) CaptureRecordSnapshotTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (revisions.RecordSnapshot, error) {
	return a.appender.CaptureRecordSnapshotTx(ctx, tx, recordID)
}

func (a revisionAdapter) AppendChangeSetTx(ctx context.Context, tx pgx.Tx, params timeline.ChangeSetParams) (uuid.UUID, error) {
	return a.appender.AppendChangeSetTx(ctx, tx, revisions.AppendChangeSetParams(params))
}

func (a revisionAdapter) AppendMutationTx(ctx context.Context, tx pgx.Tx, params timeline.MutationParams) error {
	return a.appender.AppendNonRowMutationTx(ctx, tx, revisions.AppendNonRowMutationParams(params))
}

func (a revisionAdapter) AppendRecordMutationTx(ctx context.Context, tx pgx.Tx, params revisions.AppendRecordMutationParams) error {
	return a.appender.AppendRecordMutationTx(ctx, tx, params)
}

func (a revisionAdapter) AppendRecordRevisionTx(ctx context.Context, tx pgx.Tx, params revisions.AppendRecordRevisionParams) error {
	return a.appender.AppendRecordRevisionTx(ctx, tx, params)
}

func (a revisionAdapter) ListRecordRevisionWindowTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, firstVersion int64, lastVersion int64) ([]timeline.RecordRevisionWindowEntry, error) {
	entries, err := a.reader.LoadRevisionWindowTx(ctx, tx, recordID, firstVersion, lastVersion)
	if err != nil {
		return nil, err
	}
	result := make([]timeline.RecordRevisionWindowEntry, 0, len(entries))
	for _, entry := range entries {
		result = append(result, timeline.RecordRevisionWindowEntry{
			ChangeSetID: entry.ChangeSetID,
			RowVersion:  entry.RowVersion,
			BeforeJSON:  entry.BeforeJSON,
			AfterJSON:   entry.AfterJSON,
			ActorUserID: entry.ActorUserID,
			CreatedAt:   entry.CreatedAt,
		})
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

func (a linkAdapter) LoadTimelineCollectionFieldsChangedTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, changedAt time.Time) ([]string, error) {
	return a.store.LoadTimelineCollectionFieldsChangedTx(ctx, tx, recordID, changedAt)
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

func (a mentionAdapter) LoadTimelineCollectionFieldsChangedTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, changedAt time.Time) ([]string, error) {
	return a.store.LoadTimelineCollectionFieldsChangedTx(ctx, tx, recordID, changedAt)
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

func (a evidenceAdapter) RefreshTimelineAttachmentProjectionsTx(ctx context.Context, tx pgx.Tx, recordIDs []uuid.UUID) error {
	return a.attachments.RefreshTimelineAttachmentProjectionsTx(ctx, tx, recordIDs)
}
