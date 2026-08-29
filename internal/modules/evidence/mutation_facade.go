package evidence

// Evidence-native mutation contract and composition.
import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	evidenceprojection "github.com/JochiRaider/cartulary/internal/modules/evidence/projectionports"
	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type mutationFacade struct {
	pool              postgres.DB
	idempotency       IdempotencyCapability
	incidentAccess    IncidentStateCapability
	recordEnvelopes   RecordEnvelopeCapability
	projectionRows    evidenceprojection.MutationRows
	supportEffects    evidenceprojection.AssociationEffects
	revisions         RevisionCapability
	sourceMutations   *sourceMutationService
	blobs             blobRepository
	blobLifecycle     blobLifecycleRepository
	conflictTokens    conflicttokens.ConflictTokenCodec
	conflictFields    conflicttokens.FieldResolver
	conflictSnapshots conflicttokens.RevisionSnapshotProjector
	keepSaved         conflicttokens.IdempotencyPort
	collaboration     collaboration.RecordChangedAppender
	mutations         evidenceSourceMutationKernel
	objects           objectstore.TypedStore
}

type createRequest struct {
	ViewSchemaID        string
	ClientTxnID         string
	Values              map[string]FieldValue
	InitialObjectBlobID *uuid.UUID
}

type patchRequest struct {
	ViewSchemaID   string
	BaseRowVersion int64
	ClientTxnID    string
	Changes        []patchChange
}

type patchChange struct {
	FieldKey       string
	Value          *FieldValue
	CanonicalValue any
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
	operation   OperationID
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
}

type RowVersionConflictError struct {
	RecordID          uuid.UUID
	BaseRowVersion    int64
	CurrentRowVersion int64
}

func (e *RowVersionConflictError) Error() string {
	return "evidence: row version conflict"
}

type SameFieldConflictError struct {
	Conflict SameFieldConflict
}

func (e *SameFieldConflictError) Error() string {
	return "evidence: same field conflict"
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

func newMutationFacade(
	pool postgres.DB,
	conflictTokens conflicttokens.ConflictTokenCodec,
	sourceMutations *sourceMutationService,
	objects objectstore.TypedStore,
	dependencies MutationDependencies,
) (*mutationFacade, error) {
	switch {
	case isNilMutationCapability(pool):
		return nil, fmt.Errorf("compose Evidence workbook facade: Postgres is required")
	case sourceMutations == nil:
		return nil, fmt.Errorf("compose Evidence workbook facade: source mutations are required")
	case isNilMutationCapability(objects):
		return nil, fmt.Errorf("compose Evidence workbook facade: object store is required")
	}
	if err := dependencies.validate(); err != nil {
		return nil, err
	}
	return &mutationFacade{
		pool:              pool,
		idempotency:       dependencies.Idempotency,
		incidentAccess:    dependencies.IncidentState,
		recordEnvelopes:   dependencies.RecordEnvelopes,
		projectionRows:    dependencies.ProjectionRows,
		supportEffects:    dependencies.AssociationEffects,
		revisions:         dependencies.Revisions,
		sourceMutations:   sourceMutations,
		blobs:             blobRepository{db: pool},
		blobLifecycle:     blobLifecycleRepository{db: pool},
		conflictTokens:    conflictTokens,
		conflictFields:    dependencies.ConflictFields,
		conflictSnapshots: newEvidenceConflictSnapshotProjector(),
		keepSaved:         dependencies.KeepSavedIdempotency,
		collaboration:     dependencies.Collaboration,
		objects:           objects,
		mutations:         sourceMutations.mutations,
	}, nil
}
