package artifacts

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/artifacts/internal/sourcecatalog"
	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type MutationFacade struct {
	pool              postgres.DB
	idempotency       IdempotencyCapability
	incidentAccess    IncidentStateCapability
	memberReferences  MemberReferenceCapability
	linkStore         LinkCapability
	revisions         RevisionCapability
	recordEnvelopes   RecordEnvelopeCapability
	source            artifactSourceKernel
	conflictTokens    conflicttokens.ConflictTokenCodec
	conflictFields    conflicttokens.FieldResolver
	conflictSnapshots conflicttokens.RevisionSnapshotProjector
	keepSaved         conflicttokens.IdempotencyPort
}

type CreateRequest struct {
	ViewSchemaID string
	ClientTxnID  string
	Values       map[string]FieldValue
	Collections  map[string]CollectionActionPayload
}

type PatchRequest struct {
	ViewSchemaID   string
	BaseRowVersion int64
	ClientTxnID    string
	Changes        []PatchChange
}

type PatchChange struct {
	FieldKey       string
	Value          *FieldValue
	Collection     *CollectionActionPayload
	CanonicalValue any
}

type CollectionActionPayload struct {
	Actions []CollectionAction
}

type CollectionAction struct {
	Op             string
	RawText        string
	LinkedRecordID *uuid.UUID
	PartyID        *uuid.UUID
	ItemRef        string
	RiskRefText    string
	NormalizedText string
}

type CreateCommand struct {
	ActorUserID uuid.UUID
	IncidentID  uuid.UUID
	Request     CreateRequest
	RequestHash []byte
	RequestID   string
	OperationID OperationID
	Now         time.Time
}

type PatchCommand struct {
	ActorUserID         uuid.UUID
	RecordID            uuid.UUID
	Request             PatchRequest
	RequestHash         []byte
	RequestID           string
	OperationID         OperationID
	ConflictOperationID OperationID
	Now                 time.Time
}

type MutationResult struct {
	Row              map[string]any
	Created          bool
	Replayed         bool
	IncidentID       uuid.UUID
	RecordID         uuid.UUID
	ChangeSetID      uuid.UUID
	ClientTxnID      string
	RowVersion       int64
	ViewSchemaID     string
	ChangedFieldKeys []string
	ContextualLink   *ContextualLink
}

type ContextualLink struct {
	SourceRecordID uuid.UUID
	LinkType       string
}

type RowVersionConflictError struct {
	RecordID          uuid.UUID
	BaseRowVersion    int64
	CurrentRowVersion int64
}

func (e *RowVersionConflictError) Error() string {
	return "artifacts: row version conflict"
}

type SameFieldConflictError struct {
	Conflict SameFieldConflict
}

func (e *SameFieldConflictError) Error() string {
	return "artifacts: same field conflict"
}

type OptionalConflictValue struct {
	Present bool
	Value   any
}

type SameFieldConflict struct {
	ConflictToken           string
	RecordID                uuid.UUID
	FieldKey                string
	ConflictResolutionClass string
	BaseRowVersion          int64
	CurrentRowVersion       int64
	ClientValue             any
	ServerValue             any
	BaseValue               OptionalConflictValue
	ServerUpdatedBy         uuid.UUID
	ServerUpdatedAt         time.Time
	SuggestedMergedValue    OptionalConflictValue
}

func NewMutationContribution(
	pool postgres.DB,
	conflictTokens conflicttokens.ConflictTokenCodec,
	dependencies MutationDependencies,
) (*MutationFacade, error) {
	if pool == nil {
		return nil, fmt.Errorf("artifacts mutation composition: Postgres is required")
	}
	if err := dependencies.validate(); err != nil {
		return nil, err
	}
	catalog, err := sourcecatalog.Load()
	if err != nil {
		return nil, fmt.Errorf("compose Artifacts source catalog: %w", err)
	}
	conflictSnapshots, err := conflicttokens.NewRevisionSnapshotProjector(
		"cartulary.revisions.snapshot.artifact.v1",
		catalog.ConflictFieldSourceKeys(),
	)
	if err != nil {
		return nil, fmt.Errorf("compose Artifacts conflict snapshot projector: %w", err)
	}
	return &MutationFacade{
		pool:             pool,
		idempotency:      dependencies.Idempotency,
		incidentAccess:   dependencies.IncidentState,
		memberReferences: dependencies.MemberReferences,
		linkStore:        dependencies.Links,
		revisions:        dependencies.Revisions,
		recordEnvelopes:  dependencies.RecordEnvelopes,
		source: artifactSourceKernel{
			records:     dependencies.RecordEnvelopes,
			rows:        newSourceStore(catalog),
			projections: dependencies.Projections,
		},
		conflictTokens:    conflictTokens,
		conflictFields:    dependencies.ConflictFields,
		conflictSnapshots: conflictSnapshots,
		keepSaved:         dependencies.KeepSavedIdempotency,
	}, nil
}
