package revisions

import (
	"regexp"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
)

const (
	changeSetsBundlePath      = "data/change_sets.ndjson"
	changeSetMutationsPath    = "data/change_set_mutations.ndjson"
	recordRevisionsBundlePath = "data/record_revisions.ndjson"

	revisionsReferencesInvariant     = "revisions.references_complete"
	revisionsActorsInvariant         = "revisions.actor_references_complete"
	revisionsSequenceInvariant       = "revisions.mutation_sequence_contiguous"
	revisionsRecordVersionInvariant  = "revisions.record_version_unique"
	revisionsHistoryInvariant        = "revisions.history_reconstruction"
	revisionsSequenceRepairInvariant = "revisions.sequence_repair_after_validation"
)

var revisionsPositiveInteger = regexp.MustCompile(`^[1-9][0-9]*$`)

type portableChangeSet struct {
	ChangeSetID     uuid.UUID
	IncidentID      uuid.UUID
	PortableActorID uuid.UUID
	RuntimeActorID  uuid.UUID
	Source          string
	Reason          *string
	ClientTxnID     *string
	RequestID       *string
	CreatedAt       time.Time
}

type portableChangeSetMutation struct {
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

type portableRecordRevision struct {
	RevisionID  int64
	ChangeSetID uuid.UUID
	RecordID    uuid.UUID
	RowVersion  int64
	BeforeJSON  map[string]any
	AfterJSON   map[string]any
	CreatedAt   time.Time
}

type preparedRevisionsImport struct {
	incidentID   uuid.UUID
	runtimeActor uuid.UUID
	actors       sourceport.ActorCatalog
	changeSets   []portableChangeSet
	mutations    []portableChangeSetMutation
	revisions    []portableRecordRevision
}

type revisionsParseFailure struct {
	invariant string
	path      string
	identity  string
}
