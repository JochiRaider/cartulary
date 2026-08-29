package merge

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/entities/mentions"
	"github.com/JochiRaider/cartulary/internal/modules/entities/projectionports"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type entityStorePorts struct {
	assessments   AssessmentEffectsPort
	mentions      entityMentionPort
	records       entityRecordPort
	revisions     entityRevisionPort
	links         LinkEffectsPort
	projections   projectionports.MutationRows
	timeline      TimelineEffectsPort
	collaboration collaboration.RecordChangedAppender
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
	AppendLiveRevisionTx(context.Context, pgx.Tx, revisions.LiveRevisionInput) error
}

type AssessmentProtectedSetCommand struct {
	IncidentID       uuid.UUID
	RecordType       string
	SurvivorRecordID uuid.UUID
	LoserRecordID    uuid.UUID
	Now              time.Time
}

type AssessmentRepointCommand struct {
	IncidentID         uuid.UUID
	RecordType         string
	SurvivorRecordID   uuid.UUID
	LoserRecordID      uuid.UUID
	ProtectedRecordIDs []uuid.UUID
	Now                time.Time
}

type AssessmentMutation struct {
	TargetKind     string
	TargetID       string
	OperationKind  string
	BeforeValue    any
	AfterValue     any
	BeforeSnapshot *revisions.RecordSnapshot
	AfterSnapshot  *revisions.RecordSnapshot
}

type AssessmentRepointResult struct {
	Mutations      []AssessmentMutation
	RepointedCount int
}

type AssessmentProtectedSetChangedError struct {
	RecordID uuid.UUID
}

func (e *AssessmentProtectedSetChangedError) Error() string {
	return "entities: assessment merge protected set changed"
}

type AssessmentEffectsPort interface {
	LoadProtectedRecordIDsTx(context.Context, pgx.Tx, AssessmentProtectedSetCommand) ([]uuid.UUID, error)
	RepointTx(context.Context, pgx.Tx, AssessmentRepointCommand) (AssessmentRepointResult, error)
}

type entityMentionPort interface {
	RepointMergedMentionsTx(context.Context, pgx.Tx, uuid.UUID, string, uuid.UUID, uuid.UUID, uuid.UUID, time.Time) ([]mergeMutation, int, map[uuid.UUID][]string, error)
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
	projectionRows projectionports.MutationRows,
) entityStorePorts {
	return entityStorePorts{
		records:     entityRecordAdapter{store: records.NewStore()},
		revisions:   entityRevisionAdapter{appender: appender},
		projections: projectionRows,
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

func (a entityRevisionAdapter) AppendLiveRevisionTx(ctx context.Context, tx pgx.Tx, input revisions.LiveRevisionInput) error {
	return a.appender.AppendLiveRevisionTx(ctx, tx, input)
}

func mergeMutationsFromAssessmentMutations(mutations []AssessmentMutation) []mergeMutation {
	result := make([]mergeMutation, 0, len(mutations))
	for _, mutation := range mutations {
		result = append(result, mergeMutation{
			TargetKind:     mutation.TargetKind,
			TargetID:       mutation.TargetID,
			OperationKind:  mutation.OperationKind,
			BeforeValue:    mutation.BeforeValue,
			AfterValue:     mutation.AfterValue,
			BeforeSnapshot: mutation.BeforeSnapshot,
			AfterSnapshot:  mutation.AfterSnapshot,
		})
	}
	return result
}

type entityMentionAdapter struct {
	store MentionEffectsPort
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
