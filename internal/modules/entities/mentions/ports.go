package mentions

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/links"
	projectionadapters "github.com/JochiRaider/cartulary/internal/modules/projections/adapters"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	mentioneffects "github.com/JochiRaider/cartulary/internal/modules/timeline/mentioneffects"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type Store struct {
	pool      postgres.DB
	authStore *authn.Store
	ports     storePorts
}

func NewStore(pool postgres.DB) *Store {
	return &Store{
		pool:      pool,
		authStore: authn.NewStore(pool),
		ports:     newStorePorts(pool),
	}
}

type storePorts struct {
	records     recordPort
	revisions   revisionPort
	links       linkPort
	projections projectionPort
	timeline    timelinePort
}

type recordPort interface {
	AdvanceVersionTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, time.Time) (int64, error)
	LoadRowVersionTx(context.Context, pgx.Tx, uuid.UUID) (int64, error)
}

type revisionPort interface {
	InsertChangeSetTx(context.Context, pgx.Tx, changeSetParams) (uuid.UUID, error)
	InsertMutationTx(context.Context, pgx.Tx, mutationParams) error
	InsertRecordRevisionTx(context.Context, pgx.Tx, recordRevisionParams) error
}

type linkPort interface {
	GetActiveLinkTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, uuid.UUID, string) (recordLink, error)
	UpsertLinkTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, uuid.UUID, string, string, *int, uuid.UUID, time.Time) (recordLink, bool, error)
	TombstoneLinkTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, time.Time) (recordLink, error)
}

type projectionPort interface {
	RefreshEntityRowTx(context.Context, pgx.Tx, uuid.UUID, string) error
}

type timelinePort interface {
	LoadSourceRecordTx(context.Context, pgx.Tx, uuid.UUID) (timelineSourceRecord, error)
	UpdateSourceRecordTx(context.Context, pgx.Tx, timelineSourceRecord) error
	BuildRecordRowTx(context.Context, pgx.Tx, uuid.UUID) (map[string]any, error)
	RebuildTimelineProjectionTx(context.Context, pgx.Tx, uuid.UUID) error
	VersionID(uuid.UUID, int64) string
}

type changeSetParams struct {
	ChangeSetID *uuid.UUID
	IncidentID  uuid.UUID
	ActorUserID uuid.UUID
	Source      string
	Reason      *string
	ClientTxnID *string
	RequestID   *string
	CreatedAt   time.Time
}

type mutationParams struct {
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

type recordRevisionParams struct {
	ChangeSetID uuid.UUID
	RecordID    uuid.UUID
	RowVersion  int64
	BeforeValue any
	AfterValue  any
}

type recordLink struct {
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

type timelineSourceRecord struct {
	RecordID        uuid.UUID
	RowVersion      int64
	EditedAt        time.Time
	UpdatedByUserID uuid.UUID
}

var errRecordLinkNotFound = links.ErrRecordLinkNotFound

func newStorePorts(pool postgres.DB) storePorts {
	return storePorts{
		records:     recordAdapter{store: records.NewStore()},
		revisions:   revisionAdapter{store: revisions.NewStore()},
		links:       linkAdapter{store: links.NewStore()},
		projections: projectionAdapter{projector: projectionadapters.NewRowProjector(pool)},
		timeline:    timelineAdapter{projector: projectionadapters.NewRowProjector(pool)},
	}
}

type recordAdapter struct {
	store *records.Store
}

func (a recordAdapter) AdvanceVersionTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, actorUserID uuid.UUID, now time.Time) (int64, error) {
	return a.store.AdvanceVersionTx(ctx, tx, recordID, actorUserID, now)
}

func (a recordAdapter) LoadRowVersionTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (int64, error) {
	var rowVersion int64
	if err := tx.QueryRow(ctx, `SELECT row_version FROM records WHERE record_id = $1`, recordID).Scan(&rowVersion); err != nil {
		return 0, err
	}
	return rowVersion, nil
}

type revisionAdapter struct {
	store *revisions.Store
}

func (a revisionAdapter) InsertChangeSetTx(ctx context.Context, tx pgx.Tx, params changeSetParams) (uuid.UUID, error) {
	return a.store.InsertChangeSetTx(ctx, tx, revisions.ChangeSetParams(params))
}

func (a revisionAdapter) InsertMutationTx(ctx context.Context, tx pgx.Tx, params mutationParams) error {
	return a.store.InsertMutationTx(ctx, tx, revisions.MutationParams(params))
}

func (a revisionAdapter) InsertRecordRevisionTx(ctx context.Context, tx pgx.Tx, params recordRevisionParams) error {
	return a.store.InsertRecordRevisionTx(ctx, tx, revisions.RecordRevisionParams(params))
}

type linkAdapter struct {
	store *links.Store
}

func (a linkAdapter) GetActiveLinkTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, srcRecordID uuid.UUID, dstRecordID uuid.UUID, linkType string) (recordLink, error) {
	link, err := a.store.GetActiveLinkTx(ctx, tx, incidentID, srcRecordID, dstRecordID, linkType)
	if err != nil {
		return recordLink{}, err
	}
	return recordLinkFromLinks(link), nil
}

func (a linkAdapter) UpsertLinkTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, srcRecordID uuid.UUID, dstRecordID uuid.UUID, linkType string, provenance string, confidence *int, ownerUserID uuid.UUID, now time.Time) (recordLink, bool, error) {
	link, inserted, err := a.store.UpsertLinkTx(ctx, tx, incidentID, srcRecordID, dstRecordID, linkType, provenance, confidence, ownerUserID, now)
	if err != nil {
		return recordLink{}, false, err
	}
	return recordLinkFromLinks(link), inserted, nil
}

func (a linkAdapter) TombstoneLinkTx(ctx context.Context, tx pgx.Tx, recordLinkID uuid.UUID, actorUserID uuid.UUID, now time.Time) (recordLink, error) {
	link, err := a.store.TombstoneLinkTx(ctx, tx, recordLinkID, actorUserID, now)
	if err != nil {
		return recordLink{}, err
	}
	return recordLinkFromLinks(link), nil
}

func recordLinkFromLinks(link links.RecordLink) recordLink {
	return recordLink{
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

type projectionAdapter struct {
	projector *projectionadapters.RowProjector
}

func (a projectionAdapter) RefreshEntityRowTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, entityType string) error {
	switch entityType {
	case "host":
		return a.projector.RefreshRowTx(ctx, tx, projectionadapters.HostsViewSchemaID, recordID)
	case "identity":
		return a.projector.RefreshRowTx(ctx, tx, projectionadapters.IdentitiesViewSchemaID, recordID)
	default:
		return nil
	}
}

type timelineAdapter struct {
	projector *projectionadapters.RowProjector
}

func (a timelineAdapter) LoadSourceRecordTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (timelineSourceRecord, error) {
	record, err := mentioneffects.LoadSourceRecordTx(ctx, tx, recordID)
	if errors.Is(err, mentioneffects.ErrSourceRecordNotFound) {
		return timelineSourceRecord{}, ErrSourceRecordNotFound
	}
	if err != nil {
		return timelineSourceRecord{}, err
	}
	return timelineSourceRecord{
		RecordID:        record.RecordID,
		RowVersion:      record.RowVersion,
		EditedAt:        record.EditedAt,
		UpdatedByUserID: record.UpdatedByUserID,
	}, nil
}

func (a timelineAdapter) UpdateSourceRecordTx(ctx context.Context, tx pgx.Tx, record timelineSourceRecord) error {
	return mentioneffects.UpdateSourceRecordTx(ctx, tx, mentioneffects.SourceRecord{
		RecordID:        record.RecordID,
		RowVersion:      record.RowVersion,
		EditedAt:        record.EditedAt,
		UpdatedByUserID: record.UpdatedByUserID,
	})
}

func (a timelineAdapter) BuildRecordRowTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (map[string]any, error) {
	row, err := mentioneffects.BuildRecordRowTx(ctx, tx, recordID)
	if errors.Is(err, mentioneffects.ErrSourceRecordNotFound) {
		return nil, ErrSourceRecordNotFound
	}
	return row, err
}

func (a timelineAdapter) RebuildTimelineProjectionTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID) error {
	return mentioneffects.RebuildTimelineProjectionTx(ctx, tx, a.projector, incidentID)
}

func (a timelineAdapter) VersionID(recordID uuid.UUID, rowVersion int64) string {
	return mentioneffects.VersionID(recordID, rowVersion)
}

func decodeStoredResponse(data []byte) (map[string]any, error) {
	var payload map[string]any
	if err := json.Unmarshal(data, &payload); err != nil {
		return nil, err
	}
	return payload, nil
}

func invalidMutationPayload(field string, reasonCode string) *httpapi.APIError {
	details := map[string]any{}
	if field != "" {
		details["field"] = field
	}
	if reasonCode != "" {
		details["reason_code"] = reasonCode
	}
	return &httpapi.APIError{
		Status:  http.StatusBadRequest,
		Code:    "invalid_mutation_payload",
		Message: "invalid mutation payload",
		Details: details,
	}
}

func decodeObject(reader io.Reader) (map[string]json.RawMessage, *httpapi.APIError) {
	var raw map[string]json.RawMessage
	decoder := json.NewDecoder(reader)
	if err := decoder.Decode(&raw); err != nil {
		return nil, invalidMutationPayload("", "request_not_object")
	}
	return raw, nil
}

func derefString(value *string) any {
	if value == nil {
		return nil
	}
	return *value
}

func formatUUIDPointer(value *uuid.UUID) any {
	if value == nil {
		return nil
	}
	return value.String()
}

func cloneStringPointer(value *string) *string {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func formatTimestampPointer(value *time.Time) any {
	if value == nil {
		return nil
	}
	return value.UTC().Format(time.RFC3339Nano)
}

func EntityMentionNotFoundError() *httpapi.APIError {
	return &httpapi.APIError{
		Status:  http.StatusNotFound,
		Code:    "entity_mention_not_found",
		Message: "entity mention not found",
		Details: map[string]any{},
	}
}

func ResolvedRecordNotFoundError() *httpapi.APIError {
	return &httpapi.APIError{
		Status:  http.StatusNotFound,
		Code:    "resolved_record_not_found",
		Message: "resolved record not found",
		Details: map[string]any{},
	}
}

func RowVersionConflictAPIError(conflict *MentionRowVersionConflictError) *httpapi.APIError {
	details := map[string]any{}
	if conflict != nil {
		details["entity_mention_id"] = conflict.EntityMentionID.String()
		details["base_mention_row_version"] = conflict.BaseMentionRowVersion
		details["current_mention_row_version"] = conflict.CurrentMentionRowVersion
		details["source_record_id"] = conflict.SourceRecordID.String()
	}
	return &httpapi.APIError{
		Status:  http.StatusConflict,
		Code:    "row_version_conflict",
		Message: "row version conflict",
		Details: details,
	}
}

func IllegalTransitionAPIError(err *MentionTransitionError) *httpapi.APIError {
	details := map[string]any{}
	if err != nil {
		details["from_status"] = err.FromStatus
		details["to_status"] = err.ToStatus
		details["violated_guards"] = append([]string(nil), err.ViolatedGuards...)
	}
	return &httpapi.APIError{
		Status:  http.StatusConflict,
		Code:    "illegal_transition",
		Message: "illegal transition",
		Details: details,
	}
}

func RecordDeletedUseRestoreError() *httpapi.APIError {
	return &httpapi.APIError{
		Status:  http.StatusConflict,
		Code:    "record_deleted_use_restore",
		Message: "record deleted use restore",
		Details: map[string]any{},
	}
}
