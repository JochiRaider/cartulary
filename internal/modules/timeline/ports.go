package timeline

import (
	"context"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/sourcerepository"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
)

// Collaborators are source-owner ports supplied by application assembly.
// Every mutation method receives Timeline's initiating transaction and must
// not commit, roll back, publish, or start a nested transaction.
type Collaborators struct {
	Core        CoreCollaborators
	Collections CollectionCollaborators
	Commit      CommitCollaborators
}

type CoreCollaborators struct {
	Idempotency IdempotencyPort
	Incidents   IncidentPort
	Records     RecordPort
	Revisions   RevisionPort
}

type CollectionCollaborators struct {
	Links    LinkPort
	Mentions MentionPort
	Entities EntityPort
	Evidence EvidencePort
	Facts    CollectionFactPort
}

type CommitCollaborators struct {
	Projection       workbookprojection.Writer
	EntityProjection EntityProjectionPort
	Collaboration    CollaborationPort
}

type IdempotencyPort interface {
	GetRouteIdempotency(context.Context, authn.RouteIdempotencyKey) (authn.RouteIdempotencyRecord, error)
	InsertRouteIdempotencyPayload(context.Context, pgx.Tx, authn.RouteIdempotencyKey, *uuid.UUID, []byte, int, any) error
}

type IncidentPort interface {
	RequireOpenTx(context.Context, pgx.Tx, uuid.UUID) error
}

type RecordCreateParams struct {
	RecordID        *uuid.UUID
	IncidentID      uuid.UUID
	RecordType      string
	CreatedByUserID uuid.UUID
	CreatedAt       time.Time
	UpdatedByUserID uuid.UUID
	UpdatedAt       time.Time
	RowVersion      int64
}

type RecordPort interface {
	InsertTx(context.Context, pgx.Tx, RecordCreateParams) (uuid.UUID, error)
	AdvanceVersionTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, time.Time) (int64, error)
	LoadRowVersionTx(context.Context, pgx.Tx, uuid.UUID) (int64, error)
	LoadEnvelopeTx(context.Context, pgx.Tx, uuid.UUID, bool) (sourcerepository.Envelope, error)
	LoadEnvelopesTx(context.Context, pgx.Tx, []uuid.UUID, bool) (map[uuid.UUID]sourcerepository.Envelope, error)
	ResolveIncident(context.Context, uuid.UUID) (uuid.UUID, error)
}

type RevisionPort interface {
	CaptureRecordSnapshotTx(context.Context, pgx.Tx, uuid.UUID) (revisions.RecordSnapshot, error)
	AppendChangeSetTx(context.Context, pgx.Tx, ChangeSetParams) (uuid.UUID, error)
	AppendMutationTx(context.Context, pgx.Tx, MutationParams) error
	AppendRecordMutationTx(context.Context, pgx.Tx, revisions.AppendRecordMutationParams) error
	AppendLiveRevisionTx(context.Context, pgx.Tx, revisions.LiveRevisionInput) error
	ListRecordRevisionWindowTx(context.Context, pgx.Tx, uuid.UUID, int64, int64) ([]RecordRevisionWindowEntry, error)
}

type ChangeSetParams struct {
	ChangeSetID *uuid.UUID
	IncidentID  uuid.UUID
	ActorUserID uuid.UUID
	Source      string
	Reason      *string
	ClientTxnID *string
	RequestID   *string
	CreatedAt   time.Time
}

type MutationParams struct {
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

type RecordRevisionWindowEntry struct {
	ChangeSetID uuid.UUID
	RowVersion  int64
	BeforeJSON  []byte
	AfterJSON   []byte
	ActorUserID uuid.UUID
	CreatedAt   time.Time
}

type EntityProjectionPort interface {
	RefreshHostTx(context.Context, pgx.Tx, uuid.UUID) error
	RefreshIdentityTx(context.Context, pgx.Tx, uuid.UUID) error
}

type LinkPort interface {
	InsertSupersedesCommandTx(context.Context, pgx.Tx, InsertSupersedesCommand) (RecordLinkCommandResult, error)
	UpsertLinkCommandTx(context.Context, pgx.Tx, UpsertLinkCommand) (RecordLinkCommandResult, error)
	HasActiveIncomingSupersedesLinkForUpdateTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID) (bool, error)
	ApplyRecordRefCollectionWithMutationValuesTx(context.Context, pgx.Tx, RecordRefCollectionCommand) (CollectionMutationResult, error)
	ApplyTagCollectionWithMutationValuesTx(context.Context, pgx.Tx, TagCollectionCommand) (CollectionMutationResult, error)
	LoadCollectionFieldsChangedTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID, time.Time) ([]string, error)
}

type InsertSupersedesCommand struct {
	IncidentID          uuid.UUID
	ReplacementRecordID uuid.UUID
	SupersededRecordID  uuid.UUID
	OwnerUserID         uuid.UUID
	Now                 time.Time
}

type UpsertLinkCommand struct {
	IncidentID  uuid.UUID
	SrcRecordID uuid.UUID
	DstRecordID uuid.UUID
	LinkType    string
	Provenance  string
	Confidence  *int
	OwnerUserID uuid.UUID
	Now         time.Time
}

type RecordLinkCommandResult struct {
	RecordLinkID uuid.UUID
	SrcRecordID  uuid.UUID
	DstRecordID  uuid.UUID
	LinkType     string
	Mutation     *RecordLinkMutation
}

type RecordRefCollectionCommand struct {
	IncidentID         uuid.UUID
	SourceRecordID     uuid.UUID
	ActorUserID        uuid.UUID
	FieldKey           string
	LinkType           string
	ExpectedTargetType string
	AddRecordIDs       []uuid.UUID
	RemoveRecordIDs    []uuid.UUID
	Now                time.Time
}

type TagCollectionCommand struct {
	IncidentID  uuid.UUID
	RecordID    uuid.UUID
	ActorUserID uuid.UUID
	FieldKey    string
	AddTags     []TagCollectionAdd
	RemoveTags  []RecordTagRef
	Now         time.Time
}

type TagCollectionAdd struct {
	RawText        string
	NormalizedText string
}

type RecordTagRef struct {
	RecordID    uuid.UUID
	RecordTagID uuid.UUID
}

type CollectionMutationResult struct {
	RecordLinks []RecordLinkMutation
	RecordTags  []RecordTagMutation
}

type RecordLinkMutation struct {
	RecordLinkID uuid.UUID
	Operation    string
	BeforeValue  map[string]any
	AfterValue   map[string]any
}

type RecordTagMutation struct {
	RecordTagID uuid.UUID
	RecordID    uuid.UUID
	Operation   string
	BeforeValue map[string]any
	AfterValue  map[string]any
}

type MentionPort interface {
	ResolveExistingFromMentionTx(context.Context, pgx.Tx, authn.UserRecord, uuid.UUID, string, uuid.UUID, *uuid.UUID, time.Time) ([]RecordLinkMutation, error)
	ApplyMentionLifecycleTx(context.Context, pgx.Tx, authn.UserRecord, uuid.UUID, string, uuid.UUID, string, *uuid.UUID, time.Time) ([]RecordLinkMutation, error)
	NextOrdinalTx(context.Context, pgx.Tx, uuid.UUID, string) (int, error)
	InsertTx(context.Context, pgx.Tx, MentionCreateParams) error
	LoadTimelineCollectionFieldsChangedTx(context.Context, pgx.Tx, uuid.UUID, time.Time) ([]string, error)
}

type MentionCreateParams struct {
	SourceRecordID   uuid.UUID
	EntityType       string
	SourceFieldKey   string
	OriginKind       string
	OriginLocator    string
	RawText          string
	NormalizedText   string
	ResolutionStatus string
	Ordinal          int
	CreatedByUserID  uuid.UUID
	CreatedAt        time.Time
	ResolvedRecordID *uuid.UUID
	ResolvedByUserID *uuid.UUID
	ResolvedAt       *time.Time
	ResolutionMethod *string
}

type EntityAlias struct {
	RecordID uuid.UUID
	RawText  string
}

type EntityPort interface {
	ListEligibleAliasesTx(context.Context, pgx.Tx, uuid.UUID, string) ([]EntityAlias, error)
	ValidateResolvedTargetTx(context.Context, pgx.Tx, uuid.UUID, string, uuid.UUID) error
}

type EvidencePort interface {
	ValidateTimelineAttachmentsTx(context.Context, pgx.Tx, uuid.UUID, []uuid.UUID) error
	RefreshTimelineAttachmentProjectionsTx(context.Context, pgx.Tx, []uuid.UUID) error
}

type CollectionFactPort interface {
	LoadTimelineCollectionFactsTx(context.Context, pgx.Tx, uuid.UUID, uuid.UUID) (workbookprojection.CollectionFacts, error)
}

type RecordChangeIntentParams struct {
	IncidentID       uuid.UUID
	RecordID         uuid.UUID
	RowVersion       int64
	ChangeSetID      uuid.UUID
	ClientTxnID      string
	ActorUserID      uuid.UUID
	ChangedFieldKeys []string
	ViewSchemaID     string
	ChangeKind       string
	Row              map[string]any
	PatchCells       map[string]any
	MutationOrdinal  int
	CreatedAt        time.Time
}

type CollaborationPort interface {
	AppendRecordChangeIntentTx(context.Context, pgx.Tx, RecordChangeIntentParams) error
}

func linkRecordRefItemRef(recordID uuid.UUID) string {
	return "record_ref:" + recordID.String()
}

func linkRecordTagItemRef(recordID uuid.UUID, recordTagID uuid.UUID) string {
	return "record_tag:" + recordID.String() + ":" + recordTagID.String()
}

func parseRecordRefItemRef(itemRef string) (uuid.UUID, error) {
	const prefix = "record_ref:"
	if !strings.HasPrefix(itemRef, prefix) {
		return uuid.UUID{}, fmt.Errorf("invalid item ref")
	}
	value := strings.TrimPrefix(itemRef, prefix)
	parsed, err := uuid.Parse(value)
	if err != nil || parsed.String() != value {
		return uuid.UUID{}, fmt.Errorf("invalid item ref")
	}
	return parsed, nil
}

func parseRecordTagItemRef(itemRef string) (uuid.UUID, uuid.UUID, error) {
	parts := strings.Split(itemRef, ":")
	if len(parts) != 3 || parts[0] != "record_tag" {
		return uuid.UUID{}, uuid.UUID{}, fmt.Errorf("invalid record tag item ref")
	}
	recordID, err := uuid.Parse(parts[1])
	if err != nil || recordID.String() != parts[1] {
		return uuid.UUID{}, uuid.UUID{}, fmt.Errorf("invalid record tag item ref")
	}
	tagID, err := uuid.Parse(parts[2])
	if err != nil || tagID.String() != parts[2] {
		return uuid.UUID{}, uuid.UUID{}, fmt.Errorf("invalid record tag item ref")
	}
	return recordID, tagID, nil
}
