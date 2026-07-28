package mentions

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/modules/links"
	"github.com/JochiRaider/cartulary/internal/modules/projections"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	mentioneffects "github.com/JochiRaider/cartulary/internal/modules/timeline/mentioneffects"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type Store struct {
	pool           postgres.DB
	authStore      *authn.Store
	incidentAccess incidents.Access
	ports          storePorts
}

type StoreOption func(*storePorts)

func WithTimelineEffects(effects TimelineEffectsPort) StoreOption {
	return func(ports *storePorts) {
		ports.timeline = effects
	}
}

func WithCollaborationIntents(appender collaboration.IntentAppender) StoreOption {
	return func(ports *storePorts) {
		ports.collaboration = appender
	}
}

func NewStore(pool postgres.DB, appender *revisions.Appender, options ...StoreOption) *Store {
	ports := newStorePorts(pool)
	ports.revisions = revisionAdapter{appender: appender}
	for _, option := range options {
		if option != nil {
			option(&ports)
		}
	}
	return &Store{
		pool:           pool,
		authStore:      authn.NewStore(pool),
		incidentAccess: incidents.NewAccess(pool),
		ports:          ports,
	}
}

type storePorts struct {
	records       recordPort
	revisions     revisionPort
	links         linkPort
	projections   projectionPort
	timeline      TimelineEffectsPort
	collaboration collaboration.IntentAppender
}

type recordPort interface {
	LoadRowVersionTx(context.Context, pgx.Tx, uuid.UUID) (int64, error)
}

type revisionPort interface {
	AppendChangeSetTx(context.Context, pgx.Tx, changeSetParams) (uuid.UUID, error)
	AppendMutationTx(context.Context, pgx.Tx, mutationParams) error
	AppendRecordRevisionTx(context.Context, pgx.Tx, recordRevisionParams) error
}

type linkPort interface {
	GetActiveLinkTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, uuid.UUID, string) (recordLink, error)
	UpsertLinkCommandTx(context.Context, pgx.Tx, links.UpsertLinkCommand) (recordLink, bool, error)
	TombstoneLinkTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, time.Time) (recordLink, error)
}

type projectionPort interface {
	RefreshEntityRowTx(context.Context, pgx.Tx, uuid.UUID, string) error
}

type TimelineEffectsPort interface {
	PrepareMentionActionTx(context.Context, pgx.Tx, uuid.UUID) (mentioneffects.ActionState, error)
	ApplyMentionActionEffectsTx(context.Context, pgx.Tx, mentioneffects.ActionState, mentioneffects.ActionCommand) (mentioneffects.ActionResult, error)
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

var errRecordLinkNotFound = links.ErrRecordLinkNotFound

func newStorePorts(pool postgres.DB) storePorts {
	return storePorts{
		records:     recordAdapter{store: records.NewStore()},
		links:       linkAdapter{store: links.NewStore()},
		projections: projectionAdapter{rows: projections.NewEntityRows(pool)},
	}
}

type recordAdapter struct {
	store *records.Store
}

func (a recordAdapter) LoadRowVersionTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (int64, error) {
	return a.store.LoadRowVersionTx(ctx, tx, recordID)
}

type revisionAdapter struct {
	appender *revisions.Appender
}

func (a revisionAdapter) AppendChangeSetTx(ctx context.Context, tx pgx.Tx, params changeSetParams) (uuid.UUID, error) {
	return a.appender.AppendChangeSetTx(ctx, tx, revisions.AppendChangeSetParams(params))
}

func (a revisionAdapter) AppendMutationTx(ctx context.Context, tx pgx.Tx, params mutationParams) error {
	return a.appender.AppendMutationTx(ctx, tx, revisions.AppendMutationParams(params))
}

func (a revisionAdapter) AppendRecordRevisionTx(ctx context.Context, tx pgx.Tx, params recordRevisionParams) error {
	return a.appender.AppendRecordRevisionOnlyTx(ctx, tx, revisions.AppendRecordRevisionParams(params))
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

func (a linkAdapter) UpsertLinkCommandTx(ctx context.Context, tx pgx.Tx, command links.UpsertLinkCommand) (recordLink, bool, error) {
	link, inserted, err := a.store.UpsertLinkCommandTx(ctx, tx, command)
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
	rows *projections.EntityRows
}

func (a projectionAdapter) RefreshEntityRowTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID, entityType string) error {
	switch entityType {
	case "host":
		return a.rows.RefreshHostTx(ctx, tx, recordID)
	case "identity":
		return a.rows.RefreshIdentityTx(ctx, tx, recordID)
	default:
		return nil
	}
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

func uuidPointerFromPG(value pgtype.UUID) *uuid.UUID {
	if !value.Valid {
		return nil
	}
	parsed := uuid.UUID(value.Bytes)
	return &parsed
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
