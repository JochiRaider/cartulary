package merge

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/assessments"
	"github.com/JochiRaider/cartulary/internal/modules/entities/mentions"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	projectionadapters "github.com/JochiRaider/cartulary/internal/modules/projections/adapters"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/linkeffects"
	mentioneffects "github.com/JochiRaider/cartulary/internal/modules/timeline/mentioneffects"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type entityStorePorts struct {
	assessments entityAssessmentPort
	mentions    entityMentionPort
	records     entityRecordPort
	revisions   entityRevisionPort
	links       entityLinkPort
	projections entityProjectionPort
	timeline    entityTimelinePort
}

type entityRecordPort interface {
	InsertTx(context.Context, pgx.Tx, entityRecordInsertParams) (uuid.UUID, error)
	AdvanceVersionTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, time.Time) (int64, error)
	LoadRowVersionTx(context.Context, pgx.Tx, uuid.UUID) (int64, error)
}

type entityRevisionPort interface {
	LockRecordEnvelopesNowaitTx(context.Context, pgx.Tx, []uuid.UUID) error
	InsertChangeSetTx(context.Context, pgx.Tx, entityChangeSetParams) (uuid.UUID, error)
	InsertMutationTx(context.Context, pgx.Tx, entityMutationParams) error
	InsertRecordRevisionTx(context.Context, pgx.Tx, entityRecordRevisionParams) error
}

type entityLinkPort interface {
	GetActiveLinkTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, uuid.UUID, string) (entityRecordLink, error)
	UpsertLinkTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, uuid.UUID, string, string, *int, uuid.UUID, time.Time) (entityRecordLink, bool, error)
	TombstoneLinkTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, time.Time) (entityRecordLink, error)
	RepointMergedLinksTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID, time.Time) ([]mergeMutation, int, int, map[uuid.UUID][]string, error)
	RepointMergedTagsTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID, time.Time) ([]mergeMutation, int, int, error)
}

type entityProjectionPort interface {
	RefreshEntityRowTx(context.Context, pgx.Tx, uuid.UUID, string) error
	DeleteEntityRowTx(context.Context, pgx.Tx, uuid.UUID, string) error
	RebuildEntityProjectionTx(context.Context, pgx.Tx, uuid.UUID, string) error
}

type entityAssessmentPort interface {
	LoadMergeProtectedRecordIDsTx(context.Context, pgx.Tx, uuid.UUID, string, uuid.UUID) ([]uuid.UUID, error)
	RepointMergedAssessmentsTx(context.Context, pgx.Tx, uuid.UUID, string, uuid.UUID, uuid.UUID, map[uuid.UUID]struct{}, time.Time) ([]mergeMutation, int, error)
}

type entityMentionPort interface {
	RepointMergedMentionsTx(context.Context, pgx.Tx, uuid.UUID, string, uuid.UUID, uuid.UUID, uuid.UUID, time.Time) ([]mergeMutation, int, map[uuid.UUID][]string, error)
}

type entityTimelinePort interface {
	LoadTimelineInvalidationsTx(context.Context, pgx.Tx, map[uuid.UUID][]string) ([]MergeTimelineInvalidation, error)
	LoadRelationshipInvalidationsTx(context.Context, pgx.Tx, map[uuid.UUID][]string) ([]MergeTimelineInvalidation, error)
}

type entityChangeSetParams struct {
	ChangeSetID *uuid.UUID
	IncidentID  uuid.UUID
	ActorUserID uuid.UUID
	Source      string
	Reason      *string
	ClientTxnID *string
	RequestID   *string
	CreatedAt   time.Time
}

type entityRecordInsertParams struct {
	RecordID        *uuid.UUID
	IncidentID      uuid.UUID
	RecordType      string
	CreatedByUserID uuid.UUID
	CreatedAt       time.Time
	UpdatedByUserID uuid.UUID
	UpdatedAt       time.Time
	RowVersion      int64
}

type entityMutationParams struct {
	ChangeSetID     uuid.UUID
	SequenceNo      int
	TargetKind      string
	TargetID        string
	OperationKind   string
	BeforeVersionID *string
	AfterVersionID  *string
	BeforeValue     any
	AfterValue      any
}

type entityRecordRevisionParams struct {
	ChangeSetID uuid.UUID
	RecordID    uuid.UUID
	RowVersion  int64
	BeforeValue any
	AfterValue  any
}

type entityRecordLink struct {
	RecordLinkID uuid.UUID
	IncidentID   uuid.UUID
	SrcRecordID  uuid.UUID
	DstRecordID  uuid.UUID
	LinkType     string
	Provenance   string
	Confidence   *int
	OwnerUserID  uuid.UUID
	DecidedAt    time.Time
	CreatedAt    time.Time
	DeletedAt    *time.Time
}

var errEntityRecordEnvelopeNotFound = revisions.ErrRecordNotFound

type entityRecordLockedError struct {
	RecordID uuid.UUID
}

func (e *entityRecordLockedError) Error() string {
	return "entities: record envelope locked"
}

func newEntityStorePorts(pool postgres.DB) entityStorePorts {
	return entityStorePorts{
		assessments: entityAssessmentAdapter{store: assessments.NewStore(pool)},
		mentions:    entityMentionAdapter{store: mentions.NewStore(pool)},
		records:     entityRecordAdapter{store: records.NewStore()},
		revisions:   entityRevisionAdapter{store: revisions.NewStore()},
		links:       entityLinkAdapter{store: links.NewStore()},
		projections: entityProjectionAdapter{projector: projectionadapters.NewRowProjector(pool)},
		timeline:    entityTimelineAdapter{},
	}
}

type entityRecordAdapter struct {
	store *records.Store
}

func (a entityRecordAdapter) InsertTx(ctx context.Context, tx pgx.Tx, params entityRecordInsertParams) (uuid.UUID, error) {
	return a.store.InsertTx(ctx, tx, records.InsertParams(params))
}

func (a entityRecordAdapter) AdvanceVersionTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, actorUserID uuid.UUID, now time.Time) (int64, error) {
	return a.store.AdvanceVersionTx(ctx, tx, recordID, actorUserID, now)
}

func (a entityRecordAdapter) LoadRowVersionTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (int64, error) {
	return a.store.LoadRowVersionTx(ctx, tx, recordID)
}

type entityRevisionAdapter struct {
	store *revisions.Store
}

func (a entityRevisionAdapter) LockRecordEnvelopesNowaitTx(ctx context.Context, tx pgx.Tx, recordIDs []uuid.UUID) error {
	if err := revisions.LockRecordEnvelopesNowaitTx(ctx, tx, recordIDs); err != nil {
		var locked *revisions.RecordLockedError
		if errors.As(err, &locked) {
			return &entityRecordLockedError{RecordID: locked.RecordID}
		}
		return err
	}
	return nil
}

func (a entityRevisionAdapter) InsertChangeSetTx(ctx context.Context, tx pgx.Tx, params entityChangeSetParams) (uuid.UUID, error) {
	return a.store.InsertChangeSetTx(ctx, tx, revisions.ChangeSetParams(params))
}

func (a entityRevisionAdapter) InsertMutationTx(ctx context.Context, tx pgx.Tx, params entityMutationParams) error {
	return a.store.InsertMutationTx(ctx, tx, revisions.MutationParams(params))
}

func (a entityRevisionAdapter) InsertRecordRevisionTx(ctx context.Context, tx pgx.Tx, params entityRecordRevisionParams) error {
	return a.store.InsertRecordRevisionTx(ctx, tx, revisions.RecordRevisionParams(params))
}

type entityLinkAdapter struct {
	store *links.Store
}

func (a entityLinkAdapter) GetActiveLinkTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, srcRecordID uuid.UUID, dstRecordID uuid.UUID, linkType string) (entityRecordLink, error) {
	link, err := a.store.GetActiveLinkTx(ctx, tx, incidentID, srcRecordID, dstRecordID, linkType)
	if err != nil {
		return entityRecordLink{}, err
	}
	return entityRecordLinkFromLinks(link), nil
}

func (a entityLinkAdapter) UpsertLinkTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, srcRecordID uuid.UUID, dstRecordID uuid.UUID, linkType string, provenance string, confidence *int, ownerUserID uuid.UUID, now time.Time) (entityRecordLink, bool, error) {
	link, inserted, err := a.store.UpsertLinkTx(ctx, tx, incidentID, srcRecordID, dstRecordID, linkType, provenance, confidence, ownerUserID, now)
	if err != nil {
		return entityRecordLink{}, false, err
	}
	return entityRecordLinkFromLinks(link), inserted, nil
}

func (a entityLinkAdapter) TombstoneLinkTx(ctx context.Context, tx pgx.Tx, recordLinkID uuid.UUID, actorUserID uuid.UUID, now time.Time) (entityRecordLink, error) {
	link, err := a.store.TombstoneLinkTx(ctx, tx, recordLinkID, actorUserID, now)
	if err != nil {
		return entityRecordLink{}, err
	}
	return entityRecordLinkFromLinks(link), nil
}

func (a entityLinkAdapter) RepointMergedLinksTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, survivorRecordID uuid.UUID, loserRecordID uuid.UUID, actorUserID uuid.UUID, now time.Time) ([]mergeMutation, int, int, map[uuid.UUID][]string, error) {
	result, err := a.store.RepointMergedLinksTx(ctx, tx, links.RepointMergedLinksCommand{
		IncidentID:       incidentID,
		SurvivorRecordID: survivorRecordID,
		LoserRecordID:    loserRecordID,
		ActorUserID:      actorUserID,
		Now:              now,
	})
	if err != nil {
		return nil, 0, 0, nil, err
	}
	return mergeMutationsFromLinkMutations(result.Mutations), result.RepointedCount, result.DedupedCount, result.LinkTypesBySourceRecordID, nil
}

func (a entityLinkAdapter) RepointMergedTagsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, survivorRecordID uuid.UUID, loserRecordID uuid.UUID, actorUserID uuid.UUID, now time.Time) ([]mergeMutation, int, int, error) {
	result, err := a.store.RepointMergedTagsTx(ctx, tx, links.RepointMergedTagsCommand{
		IncidentID:       incidentID,
		SurvivorRecordID: survivorRecordID,
		LoserRecordID:    loserRecordID,
		ActorUserID:      actorUserID,
		Now:              now,
	})
	if err != nil {
		return nil, 0, 0, err
	}
	return mergeMutationsFromLinkMutations(result.Mutations), result.RepointedCount, result.DedupedCount, nil
}

func entityRecordLinkFromLinks(link links.RecordLink) entityRecordLink {
	return entityRecordLink{
		RecordLinkID: link.RecordLinkID,
		IncidentID:   link.IncidentID,
		SrcRecordID:  link.SrcRecordID,
		DstRecordID:  link.DstRecordID,
		LinkType:     link.LinkType,
		Provenance:   link.Provenance,
		Confidence:   link.Confidence,
		OwnerUserID:  link.OwnerUserID,
		DecidedAt:    link.DecidedAt,
		CreatedAt:    link.CreatedAt,
		DeletedAt:    link.DeletedAt,
	}
}

func mergeMutationsFromLinkMutations(mutations []links.MergeMutation) []mergeMutation {
	result := make([]mergeMutation, 0, len(mutations))
	for _, mutation := range mutations {
		result = append(result, mergeMutation{
			TargetKind:      mutation.TargetKind,
			TargetID:        mutation.TargetID,
			OperationKind:   mutation.OperationKind,
			BeforeVersionID: mutation.BeforeVersionID,
			AfterVersionID:  mutation.AfterVersionID,
			BeforeValue:     mutation.BeforeValue,
			AfterValue:      mutation.AfterValue,
		})
	}
	return result
}

type entityAssessmentAdapter struct {
	store *assessments.Store
}

func (a entityAssessmentAdapter) LoadMergeProtectedRecordIDsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordType string, loserRecordID uuid.UUID) ([]uuid.UUID, error) {
	return a.store.LoadMergeProtectedRecordIDsTx(ctx, tx, incidentID, recordType, loserRecordID)
}

func (a entityAssessmentAdapter) RepointMergedAssessmentsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordType string, survivorRecordID uuid.UUID, loserRecordID uuid.UUID, protectedRecordSet map[uuid.UUID]struct{}, now time.Time) ([]mergeMutation, int, error) {
	result, err := a.store.RepointMergedAssessmentsTx(ctx, tx, assessments.RepointMergedAssessmentsCommand{
		IncidentID:         incidentID,
		SubjectType:        recordType,
		SurvivorRecordID:   survivorRecordID,
		LoserRecordID:      loserRecordID,
		ProtectedRecordIDs: protectedRecordSet,
		Now:                now,
	})
	if err != nil {
		var protectedSetChanged *assessments.MergeProtectedSetChangedError
		if errors.As(err, &protectedSetChanged) {
			return nil, 0, &MergePreconditionError{
				ReasonCode: "protected_set_changed",
				Details: map[string]any{
					"record_id": protectedSetChanged.RecordID.String(),
				},
			}
		}
		return nil, 0, err
	}
	return mergeMutationsFromAssessmentMutations(result.Mutations), result.RepointedCount, nil
}

func mergeMutationsFromAssessmentMutations(mutations []assessments.MergeMutation) []mergeMutation {
	result := make([]mergeMutation, 0, len(mutations))
	for _, mutation := range mutations {
		result = append(result, mergeMutation{
			TargetKind:      mutation.TargetKind,
			TargetID:        mutation.TargetID,
			OperationKind:   mutation.OperationKind,
			BeforeVersionID: mutation.BeforeVersionID,
			AfterVersionID:  mutation.AfterVersionID,
			BeforeValue:     mutation.BeforeValue,
			AfterValue:      mutation.AfterValue,
		})
	}
	return result
}

type entityMentionAdapter struct {
	store *mentions.Store
}

func (a entityMentionAdapter) RepointMergedMentionsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordType string, survivorRecordID uuid.UUID, loserRecordID uuid.UUID, actorUserID uuid.UUID, now time.Time) ([]mergeMutation, int, map[uuid.UUID][]string, error) {
	result, err := a.store.RepointMergedMentionsTx(ctx, tx, mentions.RepointMergedMentionsCommand{
		IncidentID:       incidentID,
		EntityType:       recordType,
		SurvivorRecordID: survivorRecordID,
		LoserRecordID:    loserRecordID,
		ActorUserID:      actorUserID,
		Now:              now,
	})
	if err != nil {
		return nil, 0, nil, err
	}
	return mergeMutationsFromMentionMutations(result.Mutations), result.RepointedCount, result.TimelineInvalidations, nil
}

func mergeMutationsFromMentionMutations(mutations []mentions.MergeMutation) []mergeMutation {
	result := make([]mergeMutation, 0, len(mutations))
	for _, mutation := range mutations {
		result = append(result, mergeMutation{
			TargetKind:      mutation.TargetKind,
			TargetID:        mutation.TargetID,
			OperationKind:   mutation.OperationKind,
			BeforeVersionID: mutation.BeforeVersionID,
			AfterVersionID:  mutation.AfterVersionID,
			BeforeValue:     mutation.BeforeValue,
			AfterValue:      mutation.AfterValue,
		})
	}
	return result
}

type entityProjectionAdapter struct {
	projector *projectionadapters.RowProjector
}

func (a entityProjectionAdapter) RefreshEntityRowTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, entityType string) error {
	switch entityType {
	case "host":
		return a.projector.RefreshRowTx(ctx, tx, projectionadapters.HostsViewSchemaID, recordID)
	case "identity":
		return a.projector.RefreshRowTx(ctx, tx, projectionadapters.IdentitiesViewSchemaID, recordID)
	default:
		return nil
	}
}

func (a entityProjectionAdapter) DeleteEntityRowTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, entityType string) error {
	switch entityType {
	case "host":
		return a.projector.DeleteRowTx(ctx, tx, projectionadapters.HostsViewSchemaID, recordID)
	case "identity":
		return a.projector.DeleteRowTx(ctx, tx, projectionadapters.IdentitiesViewSchemaID, recordID)
	default:
		return nil
	}
}

func (a entityProjectionAdapter) RebuildEntityProjectionTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, entityType string) error {
	switch entityType {
	case "host":
		return a.projector.RebuildIncidentViewTx(ctx, tx, projectionadapters.HostsViewSchemaID, incidentID)
	case "identity":
		return a.projector.RebuildIncidentViewTx(ctx, tx, projectionadapters.IdentitiesViewSchemaID, incidentID)
	default:
		return nil
	}
}

type entityTimelineAdapter struct{}

func (a entityTimelineAdapter) LoadTimelineInvalidationsTx(ctx context.Context, tx pgx.Tx, fieldKeysByRecord map[uuid.UUID][]string) ([]MergeTimelineInvalidation, error) {
	invalidations, err := mentioneffects.LoadTimelineInvalidationsTx(ctx, tx, fieldKeysByRecord)
	if err != nil {
		return nil, err
	}
	result := make([]MergeTimelineInvalidation, 0, len(invalidations))
	for _, invalidation := range invalidations {
		result = append(result, MergeTimelineInvalidation{
			RecordID:         invalidation.RecordID,
			RowVersion:       invalidation.RowVersion,
			ChangedFieldKeys: invalidation.ChangedFieldKeys,
		})
	}
	return result, nil
}

func (a entityTimelineAdapter) LoadRelationshipInvalidationsTx(ctx context.Context, tx pgx.Tx, linkTypesByRecord map[uuid.UUID][]string) ([]MergeTimelineInvalidation, error) {
	invalidations, err := linkeffects.LoadTimelineInvalidationsTx(ctx, tx, linkTypesByRecord)
	if err != nil {
		return nil, err
	}
	result := make([]MergeTimelineInvalidation, 0, len(invalidations))
	for _, invalidation := range invalidations {
		result = append(result, MergeTimelineInvalidation{
			RecordID:         invalidation.RecordID,
			RowVersion:       invalidation.RowVersion,
			ChangedFieldKeys: invalidation.ChangedFieldKeys,
		})
	}
	return result, nil
}
