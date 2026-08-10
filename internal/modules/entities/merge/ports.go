package merge

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/assessments"
	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/entities/mentions"
	"github.com/JochiRaider/cartulary/internal/modules/entities/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type entityStorePorts struct {
	assessments   entityAssessmentPort
	mentions      entityMentionPort
	records       entityRecordPort
	revisions     entityRevisionPort
	links         entityLinkPort
	projections   workbookprojection.Writer
	timeline      entityTimelinePort
	collaboration collaboration.IntentAppender
}

type entityRecordPort interface {
	InsertTx(context.Context, pgx.Tx, entityRecordInsertParams) (uuid.UUID, error)
	AdvanceVersionTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, time.Time) (int64, error)
	LoadRowVersionTx(context.Context, pgx.Tx, uuid.UUID) (int64, error)
}

type entityRevisionPort interface {
	LockDestructiveOperationRecordsNowaitTx(context.Context, pgx.Tx, []uuid.UUID) error
	CaptureRecordSnapshotTx(context.Context, pgx.Tx, uuid.UUID) (revisions.RecordSnapshot, error)
	AppendChangeSetTx(context.Context, pgx.Tx, entityChangeSetParams) (uuid.UUID, error)
	AppendMutationTx(context.Context, pgx.Tx, entityMutationParams) error
	AppendRecordMutationTx(context.Context, pgx.Tx, revisions.AppendRecordMutationParams) error
	AppendRecordRevisionTx(context.Context, pgx.Tx, revisions.AppendRecordRevisionParams) error
}

type entityLinkPort interface {
	GetActiveLinkTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, uuid.UUID, string) (entityRecordLink, error)
	UpsertLinkCommandTx(context.Context, pgx.Tx, links.UpsertLinkCommand) (entityRecordLink, bool, error)
	TombstoneLinkTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, time.Time) (entityRecordLink, error)
	RepointMergedLinksTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID, time.Time) ([]mergeMutation, int, int, map[uuid.UUID][]string, error)
	RepointMergedTagsTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID, time.Time) ([]mergeMutation, int, int, error)
}

type entityAssessmentPort interface {
	LoadMergeProtectedRecordIDsTx(context.Context, pgx.Tx, uuid.UUID, string, uuid.UUID) ([]uuid.UUID, error)
	RepointMergedAssessmentsTx(context.Context, pgx.Tx, uuid.UUID, string, uuid.UUID, uuid.UUID, map[uuid.UUID]struct{}, time.Time) ([]mergeMutation, int, error)
}

type entityMentionPort interface {
	RepointMergedMentionsTx(context.Context, pgx.Tx, uuid.UUID, string, uuid.UUID, uuid.UUID, uuid.UUID, time.Time) ([]mergeMutation, int, map[uuid.UUID][]string, error)
}

type entityTimelinePort = TimelineEffectsPort

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

var errEntityRecordEnvelopeNotFound = records.ErrEnvelopeNotFound

type entityRecordLockedError struct {
	RecordID uuid.UUID
}

func (e *entityRecordLockedError) Error() string {
	return "entities: record envelope locked"
}

func newEntityStorePorts(
	pool postgres.DB,
	appender *revisions.Appender,
	projectionWriter workbookprojection.Writer,
) entityStorePorts {
	return entityStorePorts{
		records:     entityRecordAdapter{store: records.NewStore()},
		revisions:   entityRevisionAdapter{appender: appender},
		links:       entityLinkAdapter{store: links.NewStore()},
		projections: projectionWriter,
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
	appender *revisions.Appender
}

func (a entityRevisionAdapter) LockDestructiveOperationRecordsNowaitTx(ctx context.Context, tx pgx.Tx, recordIDs []uuid.UUID) error {
	if err := records.LockDestructiveOperationRecordsNowaitTx(ctx, tx, recordIDs); err != nil {
		var locked *records.DestructiveOperationRecordLockedError
		if errors.As(err, &locked) {
			return &entityRecordLockedError{RecordID: locked.RecordID}
		}
		return err
	}
	return nil
}

func (a entityRevisionAdapter) AppendChangeSetTx(ctx context.Context, tx pgx.Tx, params entityChangeSetParams) (uuid.UUID, error) {
	return a.appender.AppendChangeSetTx(ctx, tx, revisions.AppendChangeSetParams(params))
}

func (a entityRevisionAdapter) CaptureRecordSnapshotTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (revisions.RecordSnapshot, error) {
	return a.appender.CaptureRecordSnapshotTx(ctx, tx, recordID)
}

func (a entityRevisionAdapter) AppendMutationTx(ctx context.Context, tx pgx.Tx, params entityMutationParams) error {
	return a.appender.AppendNonRowMutationTx(ctx, tx, revisions.AppendNonRowMutationParams(params))
}

func (a entityRevisionAdapter) AppendRecordMutationTx(ctx context.Context, tx pgx.Tx, params revisions.AppendRecordMutationParams) error {
	return a.appender.AppendRecordMutationTx(ctx, tx, params)
}

func (a entityRevisionAdapter) AppendRecordRevisionTx(ctx context.Context, tx pgx.Tx, params revisions.AppendRecordRevisionParams) error {
	return a.appender.AppendRecordRevisionTx(ctx, tx, params)
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

func (a entityLinkAdapter) UpsertLinkCommandTx(ctx context.Context, tx pgx.Tx, command links.UpsertLinkCommand) (entityRecordLink, bool, error) {
	link, inserted, err := a.store.UpsertLinkCommandTx(ctx, tx, command)
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
	effects *assessments.MergeEffects
}

func (a entityAssessmentAdapter) LoadMergeProtectedRecordIDsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordType string, loserRecordID uuid.UUID) ([]uuid.UUID, error) {
	return a.effects.LoadProtectedRecordIDsTx(ctx, tx, incidentID, recordType, loserRecordID)
}

func (a entityAssessmentAdapter) RepointMergedAssessmentsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, recordType string, survivorRecordID uuid.UUID, loserRecordID uuid.UUID, protectedRecordSet map[uuid.UUID]struct{}, now time.Time) ([]mergeMutation, int, error) {
	mutations, repointedCount, err := a.effects.RepointTx(
		ctx,
		tx,
		incidentID,
		recordType,
		survivorRecordID,
		loserRecordID,
		protectedRecordSet,
		now,
	)
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
	return mergeMutationsFromAssessmentMutations(mutations), repointedCount, nil
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
			BeforeSnapshot:  mutation.BeforeSnapshot,
			AfterSnapshot:   mutation.AfterSnapshot,
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
