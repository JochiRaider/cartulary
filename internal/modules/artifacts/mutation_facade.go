package artifacts

import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/artifacts/internal/sourcecatalog"
	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
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
	publications      collaboration.RecordChangedAppender
}

type createRequest struct {
	ViewSchemaID string
	ClientTxnID  string
	Values       map[string]fieldValue
	Collections  map[string]collectionActionPayload
}

type patchRequest struct {
	ViewSchemaID   string
	BaseRowVersion int64
	ClientTxnID    string
	Changes        []patchChange
}

type patchChange struct {
	FieldKey       string
	Value          *fieldValue
	Collection     *collectionActionPayload
	CanonicalValue any
}

type collectionActionPayload struct {
	Actions []collectionAction
}

type collectionAction struct {
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
	Admission   CreateAdmission
	RequestID   string
	Now         time.Time
}

type PatchCommand struct {
	ActorUserID uuid.UUID
	RecordID    uuid.UUID
	Admission   PatchAdmission
	RequestID   string
	Now         time.Time
}

type MutationOutcome string

const (
	MutationOutcomeCreated   MutationOutcome = "created"
	MutationOutcomeUpdated   MutationOutcome = "updated"
	MutationOutcomeKeptSaved MutationOutcome = "kept_saved"
	MutationOutcomeReplayed  MutationOutcome = "replayed"
)

type MutationResult struct {
	Row              map[string]any
	Outcome          MutationOutcome
	IncidentID       uuid.UUID
	RecordID         uuid.UUID
	ChangeSetID      *uuid.UUID
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
	if isNilCapability(pool) {
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
		publications:      dependencies.Collaboration,
	}, nil
}
