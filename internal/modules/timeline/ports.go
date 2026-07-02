package timeline

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/entities/mentions"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	projectionadapters "github.com/JochiRaider/cartulary/internal/modules/projections/adapters"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type timelineStorePorts struct {
	idempotency timelineIdempotencyPort
	records     timelineRecordPort
	revisions   timelineRevisionPort
	projections timelineProjectionPort
	links       timelineLinkPort
	mentions    timelineMentionPort
}

type timelineIdempotencyPort interface {
	GetRouteIdempotency(context.Context, authn.RouteIdempotencyKey) (authn.RouteIdempotencyRecord, error)
	InsertRouteIdempotencyPayload(context.Context, pgx.Tx, authn.RouteIdempotencyKey, *uuid.UUID, []byte, int, any) error
}

type timelineRecordPort interface {
	AdvanceVersionTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, time.Time) (int64, error)
}

type timelineRevisionPort interface {
	InsertChangeSetTx(context.Context, pgx.Tx, timelineChangeSetParams) (uuid.UUID, error)
	InsertMutationTx(context.Context, pgx.Tx, timelineMutationParams) error
	InsertRecordRevisionTx(context.Context, pgx.Tx, timelineRecordRevisionParams) error
}

type timelineProjectionPort interface {
	UpsertTimelineRowTx(context.Context, pgx.Tx, timelineProjectionInput) error
	RebuildIncidentHostsTx(context.Context, pgx.Tx, uuid.UUID) error
	RebuildIncidentIdentitiesTx(context.Context, pgx.Tx, uuid.UUID) error
}

type timelineLinkPort interface {
	InsertSupersedesTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, uuid.UUID, uuid.UUID, time.Time) (supersedesLink, error)
	UpsertLinkTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, uuid.UUID, string, string, *int, uuid.UUID, time.Time) error
}

type timelineMentionPort interface {
	ResolveExistingFromMentionTx(context.Context, pgx.Tx, authn.UserRecord, uuid.UUID, string, uuid.UUID, *uuid.UUID, time.Time) error
	ApplyMentionLifecycleTx(context.Context, pgx.Tx, authn.UserRecord, uuid.UUID, string, uuid.UUID, string, *uuid.UUID, time.Time) error
}

type timelineChangeSetParams struct {
	ChangeSetID *uuid.UUID
	IncidentID  uuid.UUID
	ActorUserID uuid.UUID
	Source      string
	Reason      *string
	ClientTxnID *string
	RequestID   *string
	CreatedAt   time.Time
}

type timelineMutationParams struct {
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

type timelineRecordRevisionParams struct {
	ChangeSetID uuid.UUID
	RecordID    uuid.UUID
	RowVersion  int64
	BeforeValue any
	AfterValue  any
}

type timelineProjectionInput struct {
	RecordID              uuid.UUID
	IncidentID            uuid.UUID
	RowVersion            int64
	DateEnteredText       *string
	AnalystText           *string
	MitreStageText        *string
	DeviceObjectText      *string
	IPAddressText         *string
	ActivityUTCText       *string
	ActivityLocalText     *string
	RawActivityText       *string
	ActivitySynopsisText  *string
	DataSourceText        *string
	RecordedAt            time.Time
	EditedAt              time.Time
	ActivitySortTS        *time.Time
	DateEnteredSortDay    *time.Time
	ActivityTimePairState string
	CaptureState          string
	ReplacementRecordID   *uuid.UUID
	EvidenceCount         int
	HasEvidence           bool
	HasUnresolvedMentions bool
}

type supersedesLink struct {
	RecordLinkID uuid.UUID
	IncidentID   uuid.UUID
	SrcRecordID  uuid.UUID
	DstRecordID  uuid.UUID
}

func newTimelineStorePorts(pool postgres.DB) timelineStorePorts {
	return timelineStorePorts{
		idempotency: timelineIdempotencyAdapter{store: authn.NewStore(pool)},
		records:     timelineRecordAdapter{store: records.NewStore()},
		revisions:   timelineRevisionAdapter{store: revisions.NewStore()},
		projections: timelineProjectionAdapter{
			timeline: projectionadapters.NewTimelineProjector(pool),
			rows:     projectionadapters.NewRowProjector(pool),
		},
		links:    timelineLinkAdapter{store: links.NewStore()},
		mentions: timelineMentionAdapter{store: mentions.NewStore(nil)},
	}
}

type timelineIdempotencyAdapter struct {
	store *authn.Store
}

func (a timelineIdempotencyAdapter) GetRouteIdempotency(ctx context.Context, key authn.RouteIdempotencyKey) (authn.RouteIdempotencyRecord, error) {
	return a.store.GetRouteIdempotency(ctx, key)
}

func (a timelineIdempotencyAdapter) InsertRouteIdempotencyPayload(ctx context.Context, tx pgx.Tx, key authn.RouteIdempotencyKey, targetUserID *uuid.UUID, requestHash []byte, statusCode int, payload any) error {
	return authn.InsertRouteIdempotencyPayload(ctx, tx, key, targetUserID, requestHash, statusCode, payload)
}

type timelineRecordAdapter struct {
	store *records.Store
}

func (a timelineRecordAdapter) AdvanceVersionTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, actorUserID uuid.UUID, now time.Time) (int64, error) {
	return a.store.AdvanceVersionTx(ctx, tx, recordID, actorUserID, now)
}

type timelineRevisionAdapter struct {
	store *revisions.Store
}

func (a timelineRevisionAdapter) InsertChangeSetTx(ctx context.Context, tx pgx.Tx, params timelineChangeSetParams) (uuid.UUID, error) {
	return a.store.InsertChangeSetTx(ctx, tx, revisions.ChangeSetParams(params))
}

func (a timelineRevisionAdapter) InsertMutationTx(ctx context.Context, tx pgx.Tx, params timelineMutationParams) error {
	return a.store.InsertMutationTx(ctx, tx, revisions.MutationParams(params))
}

func (a timelineRevisionAdapter) InsertRecordRevisionTx(ctx context.Context, tx pgx.Tx, params timelineRecordRevisionParams) error {
	return a.store.InsertRecordRevisionTx(ctx, tx, revisions.RecordRevisionParams(params))
}

type timelineProjectionAdapter struct {
	timeline *projectionadapters.TimelineProjector
	rows     *projectionadapters.RowProjector
}

func (a timelineProjectionAdapter) UpsertTimelineRowTx(ctx context.Context, tx pgx.Tx, input timelineProjectionInput) error {
	return a.timeline.UpsertTimelineRowTx(ctx, tx, projectionadapters.TimelineProjectionInput(input))
}

func (a timelineProjectionAdapter) RebuildIncidentHostsTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	return a.rows.RebuildIncidentViewTx(ctx, tx, projectionadapters.HostsViewSchemaID, incidentID)
}

func (a timelineProjectionAdapter) RebuildIncidentIdentitiesTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	return a.rows.RebuildIncidentViewTx(ctx, tx, projectionadapters.IdentitiesViewSchemaID, incidentID)
}

type timelineLinkAdapter struct {
	store *links.Store
}

func (a timelineLinkAdapter) InsertSupersedesTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, replacementRecordID uuid.UUID, supersededRecordID uuid.UUID, ownerUserID uuid.UUID, now time.Time) (supersedesLink, error) {
	link, err := a.store.InsertSupersedesTx(ctx, tx, incidentID, replacementRecordID, supersededRecordID, ownerUserID, now)
	if err != nil {
		return supersedesLink{}, err
	}
	return supersedesLink{
		RecordLinkID: link.RecordLinkID,
		IncidentID:   link.IncidentID,
		SrcRecordID:  link.SrcRecordID,
		DstRecordID:  link.DstRecordID,
	}, nil
}

func (a timelineLinkAdapter) UpsertLinkTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, srcRecordID uuid.UUID, dstRecordID uuid.UUID, linkType string, provenance string, confidence *int, ownerUserID uuid.UUID, now time.Time) error {
	_, _, err := a.store.UpsertLinkTx(ctx, tx, incidentID, srcRecordID, dstRecordID, linkType, provenance, confidence, ownerUserID, now)
	return err
}

type timelineMentionAdapter struct {
	store *mentions.Store
}

func (a timelineMentionAdapter) ResolveExistingFromMentionTx(ctx context.Context, tx pgx.Tx, actor authn.UserRecord, sourceRecordID uuid.UUID, fieldKey string, mentionID uuid.UUID, resolvedRecordID *uuid.UUID, now time.Time) error {
	_, err := a.store.ResolveExistingFromMentionTx(ctx, tx, actor, sourceRecordID, fieldKey, mentionID, resolvedRecordID, now)
	return err
}

func (a timelineMentionAdapter) ApplyMentionLifecycleTx(ctx context.Context, tx pgx.Tx, actor authn.UserRecord, sourceRecordID uuid.UUID, sourceFieldKey string, mentionID uuid.UUID, action string, resolvedRecordID *uuid.UUID, now time.Time) error {
	return a.store.ApplyMentionLifecycleTx(ctx, tx, actor, sourceRecordID, sourceFieldKey, mentionID, action, resolvedRecordID, now)
}
