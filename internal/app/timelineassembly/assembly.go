package timelineassembly

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity"
	"github.com/JochiRaider/cartulary/internal/modules/entities/mentions"
	"github.com/JochiRaider/cartulary/internal/modules/entities/merge"
	"github.com/JochiRaider/cartulary/internal/modules/evidence"
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	projectionadapters "github.com/JochiRaider/cartulary/internal/modules/projections/adapters"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/recovery/restorecontract"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/timeline"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/mentioneffects"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/sourcerepository"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

func NewFacade(pool postgres.DB) *timeline.Facade {
	return timeline.NewFacade(pool, Dependencies(pool))
}

func NewProjectionSource(pool postgres.DB) *timeline.ProjectionSource {
	recordsAdapter := recordAdapter{
		store:   records.NewStore(),
		targets: records.NewRouteTargetResolver(pool),
	}
	return timeline.NewProjectionSource(recordsAdapter, collectionReadAdapter{})
}

func NewMentionEffects(pool postgres.DB) *mentioneffects.Provider {
	recordsAdapter := recordAdapter{
		store:   records.NewStore(),
		targets: records.NewRouteTargetResolver(pool),
	}
	return mentioneffects.NewProvider(
		recordsAdapter,
		collectionReadAdapter{},
		projectionadapters.NewTimelineWriter(pool),
	)
}

func NewEntityMentionStore(pool postgres.DB) *mentions.Store {
	return mentions.NewStore(
		pool,
		mentions.WithTimelineEffects(NewMentionEffects(pool)),
		mentions.WithCollaborationIntents(collaboration.NewStore(pool, nil)),
	)
}

func NewEntityMergeStore(pool postgres.DB) *merge.Store {
	return merge.NewStore(
		pool,
		merge.WithTimelineEffects(NewMentionEffects(pool)),
		merge.WithCollaborationIntents(collaboration.NewStore(pool, nil)),
	)
}

func NewEvidenceStore(pool postgres.DB) *evidence.Store {
	return evidence.NewStore(
		pool,
		evidence.WithProjectionPort(NewRowProjector(pool)),
		evidence.WithCollaborationIntents(collaboration.NewStore(pool, nil)),
	)
}

func NewRowProjector(pool postgres.DB) *projectionadapters.RowProjector {
	return projectionadapters.NewRowProjector(pool, NewProjectionSource(pool))
}

func NewRestoreRebuilder(pool postgres.DB) restorecontract.ProjectionRebuilder {
	return projectionadapters.NewRestoreRebuilder(pool, NewProjectionSource(pool))
}

func RegisterTestRoutes() httpapi.RouteRegistrar {
	return func(mux *http.ServeMux, deps httpapi.DependencySet) error {
		return registerTestControlRoutes(mux, deps)
	}
}

func Dependencies(pool postgres.DB) timeline.Dependencies {
	return timeline.Dependencies{
		Idempotency: idempotencyAdapter{store: authn.NewStore(pool)},
		Incidents:   incidentAdapter{access: incidents.NewAccess(pool)},
		Records: recordAdapter{
			store:   records.NewStore(),
			targets: records.NewRouteTargetResolver(pool),
		},
		Revisions:     revisionAdapter{appender: revisions.NewAppender(), reader: revisions.NewReader()},
		Projections:   projectionadapters.NewTimelineWriter(pool),
		Links:         linkAdapter{store: links.NewStore()},
		Mentions:      mentionAdapter{store: mentions.NewStore(nil)},
		Entities:      entityAdapter{store: hostidentity.NewStore(pool)},
		Evidence:      evidenceAdapter{store: NewEvidenceStore(pool)},
		Collections:   collectionReadAdapter{},
		Collaboration: collaborationAdapter{store: collaboration.NewStore(pool, nil)},
	}
}

type collaborationAdapter struct {
	store *collaboration.Store
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
	return a.store.AppendIntentTx(ctx, tx, intent)
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

func (a recordAdapter) LoadEnvelopeTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, lock bool) (timeline.RecordEnvelope, error) {
	envelope, err := a.store.LoadEnvelopeTx(ctx, tx, recordID, lock)
	if errors.Is(err, records.ErrEnvelopeNotFound) {
		return timeline.RecordEnvelope{}, sourcerepository.ErrEnvelopeNotFound
	}
	return timeline.RecordEnvelope(envelope), err
}

func (a recordAdapter) LoadEnvelopesTx(ctx context.Context, tx pgx.Tx, recordIDs []uuid.UUID, lock bool) (map[uuid.UUID]timeline.RecordEnvelope, error) {
	envelopes, err := a.store.LoadEnvelopesTx(ctx, tx, recordIDs, lock)
	if err != nil {
		return nil, err
	}
	result := make(map[uuid.UUID]timeline.RecordEnvelope, len(envelopes))
	for recordID, envelope := range envelopes {
		result[recordID] = timeline.RecordEnvelope(envelope)
	}
	return result, nil
}

func (a recordAdapter) ResolveIncident(ctx context.Context, recordID uuid.UUID) (uuid.UUID, error) {
	return a.targets.ResolveIncident(ctx, recordID)
}

type revisionAdapter struct {
	appender revisions.Appender
	reader   revisions.Reader
}

func (a revisionAdapter) AppendChangeSetTx(ctx context.Context, tx pgx.Tx, params timeline.ChangeSetParams) (uuid.UUID, error) {
	return a.appender.AppendChangeSetTx(ctx, tx, revisions.AppendChangeSetParams(params))
}

func (a revisionAdapter) AppendMutationTx(ctx context.Context, tx pgx.Tx, params timeline.MutationParams) error {
	return a.appender.AppendMutationTx(ctx, tx, revisions.AppendMutationParams(params))
}

func (a revisionAdapter) AppendRecordRevisionTx(ctx context.Context, tx pgx.Tx, params timeline.RecordRevisionParams) error {
	return a.appender.AppendRecordRevisionTx(ctx, tx, revisions.AppendRecordRevisionParams(params))
}

func (a revisionAdapter) ListRecordRevisionWindowTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, firstVersion int64, lastVersion int64) ([]timeline.RecordRevisionWindowEntry, error) {
	entries, err := a.reader.ListRecordRevisionWindowTx(ctx, tx, recordID, firstVersion, lastVersion)
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
	store *evidence.Store
}

func (a evidenceAdapter) ValidateTimelineAttachmentsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordIDs []uuid.UUID) error {
	err := a.store.ValidateTimelineAttachmentsTx(ctx, tx, incidentID, recordIDs)
	if errors.Is(err, evidence.ErrEvidenceNotFound) {
		return &links.CollectionValidationError{
			Field:      "timeline.attached_evidence_ids",
			ReasonCode: "invalid_value",
		}
	}
	return err
}

type collectionReadAdapter struct{}

func (collectionReadAdapter) HydrateTimelineCollectionsTx(ctx context.Context, tx pgx.Tx, record *workbookprojection.DerivedRecord) error {
	return hydrateTimelineCollections(ctx, tx, record)
}
