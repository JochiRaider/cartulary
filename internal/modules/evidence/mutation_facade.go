package evidence

// Evidence-native mutation contract and composition.
import (
	"fmt"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	evidenceprojection "github.com/JochiRaider/cartulary/internal/modules/evidence/workbookprojection"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/modules/records"
	"github.com/JochiRaider/cartulary/internal/modules/revisions"
	conflicttokens "github.com/JochiRaider/cartulary/internal/modules/revisions/conflicts"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type mutationFacade struct {
	pool              postgres.DB
	authStore         *authn.Store
	incidentAccess    *admission.Checker
	recordStore       *records.Store
	projectionRows    evidenceprojection.Rows
	supportEffects    evidenceprojection.SupportProjectionEffectsTx
	revisionHistory   conflicttokens.RevisionWindowReader
	revisionAppender  *revisions.Appender
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

type CreateRequest struct {
	ViewSchemaID        string
	ClientTxnID         string
	Values              map[string]FieldValue
	InitialObjectBlobID *uuid.UUID
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
	CanonicalValue any
}

type CreateCommand struct {
	Actor       authn.UserRecord
	IncidentID  uuid.UUID
	Request     CreateRequest
	RequestHash []byte
	RequestID   string
	RouteKey    string
	Now         time.Time
}

type PatchCommand struct {
	Actor            authn.UserRecord
	RecordID         uuid.UUID
	Request          PatchRequest
	RequestHash      []byte
	RequestID        string
	RouteKey         string
	ConflictRouteKey string
	Now              time.Time
}

type MutationResult struct {
	Payload          map[string]any
	StatusCode       int
	Replayed         bool
	IncidentID       uuid.UUID
	RecordID         uuid.UUID
	ChangeSetID      uuid.UUID
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
	appender *revisions.Appender,
	intents collaboration.RecordChangedAppender,
	sourceMutations *sourceMutationService,
	objects objectstore.TypedStore,
	conflictFields conflicttokens.FieldResolver,
	keepSaved conflicttokens.IdempotencyPort,
	projectionRows evidenceprojection.Rows,
	supportEffects evidenceprojection.SupportProjectionEffectsTx,
) (*mutationFacade, error) {
	switch {
	case pool == nil:
		return nil, fmt.Errorf("compose Evidence workbook facade: Postgres is required")
	case appender == nil:
		return nil, fmt.Errorf("compose Evidence workbook facade: Revisions is required")
	case intents == nil:
		return nil, fmt.Errorf("compose Evidence workbook facade: Collaboration is required")
	case sourceMutations == nil:
		return nil, fmt.Errorf("compose Evidence workbook facade: source mutations are required")
	case objects == nil:
		return nil, fmt.Errorf("compose Evidence workbook facade: object store is required")
	case conflictFields == nil:
		return nil, fmt.Errorf("compose Evidence workbook facade: conflict fields are required")
	case keepSaved == nil:
		return nil, fmt.Errorf("compose Evidence workbook facade: conflict idempotency is required")
	case projectionRows == nil:
		return nil, fmt.Errorf("compose Evidence workbook facade: projection rows are required")
	case supportEffects == nil:
		return nil, fmt.Errorf("compose Evidence workbook facade: support projection effects are required")
	}
	recordStore := records.NewStore()
	incidentAccess := admission.NewChecker(pool)
	return &mutationFacade{
		pool:              pool,
		authStore:         authn.NewStore(pool),
		incidentAccess:    incidentAccess,
		recordStore:       recordStore,
		projectionRows:    projectionRows,
		supportEffects:    supportEffects,
		revisionHistory:   conflicttokens.NewRevisionWindowReader(),
		revisionAppender:  appender,
		sourceMutations:   sourceMutations,
		blobs:             blobRepository{db: pool},
		blobLifecycle:     blobLifecycleRepository{db: pool},
		conflictTokens:    conflictTokens,
		conflictFields:    conflictFields,
		conflictSnapshots: newEvidenceConflictSnapshotProjector(),
		keepSaved:         keepSaved,
		collaboration:     intents,
		objects:           objects,
		mutations:         sourceMutations.mutations,
	}, nil
}
