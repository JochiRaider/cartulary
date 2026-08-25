package mentions

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"reflect"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/modules/entities/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
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
	incidentAccess *admission.Checker
	ports          storePorts
}

type StoreDependencies struct {
	Postgres      postgres.DB
	Revisions     *revisions.Appender
	Links         LinkOperationsPort
	Projections   workbookprojection.Writer
	Timeline      TimelineEffectsPort
	Collaboration collaboration.RecordChangedAppender
}

func NewStore(dependencies StoreDependencies) (*Store, error) {
	for _, dependency := range []struct {
		name  string
		value any
	}{
		{name: "Postgres", value: dependencies.Postgres},
		{name: "Revisions", value: dependencies.Revisions},
		{name: "Links", value: dependencies.Links},
		{name: "Projections", value: dependencies.Projections},
		{name: "Timeline", value: dependencies.Timeline},
		{name: "Collaboration", value: dependencies.Collaboration},
	} {
		if isNilStoreDependency(dependency.value) {
			return nil, fmt.Errorf("compose Mention store: %s is required", dependency.name)
		}
	}
	ports := newStorePorts()
	ports.revisions = revisionAdapter{appender: dependencies.Revisions}
	ports.links = dependencies.Links
	ports.projections = dependencies.Projections
	ports.timeline = dependencies.Timeline
	ports.collaboration = dependencies.Collaboration
	return &Store{
		pool:           dependencies.Postgres,
		authStore:      authn.NewStore(dependencies.Postgres),
		incidentAccess: admission.NewChecker(dependencies.Postgres),
		ports:          ports,
	}, nil
}

func isNilStoreDependency(dependency any) bool {
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

type storePorts struct {
	records       recordPort
	revisions     revisionPort
	links         LinkOperationsPort
	projections   workbookprojection.Writer
	timeline      TimelineEffectsPort
	collaboration collaboration.RecordChangedAppender
}

type recordPort interface {
	LoadRowVersionTx(context.Context, pgx.Tx, uuid.UUID) (int64, error)
}

type revisionPort interface {
	CaptureRecordSnapshotTx(context.Context, pgx.Tx, uuid.UUID) (revisions.RecordSnapshot, error)
	AppendChangeSetTx(context.Context, pgx.Tx, changeSetParams) (uuid.UUID, error)
	AppendMutationTx(context.Context, pgx.Tx, mutationParams) error
	AppendRecordMutationTx(context.Context, pgx.Tx, revisions.AppendRecordMutationParams) error
	AppendLiveRevisionTx(context.Context, pgx.Tx, revisions.LiveRevisionInput) error
}

type LinkOperationsPort interface {
	UpsertMentionLinkTx(context.Context, pgx.Tx, LinkCommand) (LinkCommandResult, error)
	TombstoneActiveMentionLinkTx(context.Context, pgx.Tx, TombstoneLinkCommand) (LinkCommandResult, bool, error)
}

type LinkType string

const (
	LinkTypeObservedOnHost     LinkType = "observed_on_host"
	LinkTypeObservedAsIdentity LinkType = "observed_as_identity"
)

type LinkCommand struct {
	IncidentID  uuid.UUID
	SrcRecordID uuid.UUID
	DstRecordID uuid.UUID
	LinkType    LinkType
	ActorUserID uuid.UUID
	Now         time.Time
}

type TombstoneLinkCommand = LinkCommand

type LinkCommandResult struct {
	RecordLinkID uuid.UUID
	SrcRecordID  uuid.UUID
	DstRecordID  uuid.UUID
	LinkType     LinkType
	Mutation     *LinkMutation
}

type LinkMutation struct {
	RecordLinkID uuid.UUID
	Operation    string
	BeforeValue  map[string]any
	AfterValue   map[string]any
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

func newStorePorts() storePorts {
	return storePorts{
		records: recordAdapter{store: records.NewStore()},
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

func (a revisionAdapter) CaptureRecordSnapshotTx(ctx context.Context, tx pgx.Tx, recordID uuid.UUID) (revisions.RecordSnapshot, error) {
	return a.appender.CaptureRecordSnapshotTx(ctx, tx, recordID)
}

func (a revisionAdapter) AppendMutationTx(ctx context.Context, tx pgx.Tx, params mutationParams) error {
	return a.appender.AppendNonRowMutationTx(ctx, tx, revisions.AppendNonRowMutationParams(params))
}

func (a revisionAdapter) AppendRecordMutationTx(ctx context.Context, tx pgx.Tx, params revisions.AppendRecordMutationParams) error {
	return a.appender.AppendRecordMutationTx(ctx, tx, params)
}

func (a revisionAdapter) AppendLiveRevisionTx(ctx context.Context, tx pgx.Tx, input revisions.LiveRevisionInput) error {
	return a.appender.AppendLiveRevisionTx(ctx, tx, input)
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
