package revisions

import (
	"regexp"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
)

const (
	changeSetsBundlePath      = "data/change_sets.ndjson"
	changeSetMutationsPath    = "data/change_set_mutations.ndjson"
	recordRevisionsBundlePath = "data/record_revisions.ndjson"

	revisionsReferencesInvariant     = "revisions.references_complete"
	revisionsSourceStateInvariant    = "revisions.source_state_exact"
	revisionsActorsInvariant         = "revisions.actor_references_complete"
	revisionsSequenceInvariant       = "revisions.mutation_sequence_contiguous"
	revisionsRecordVersionInvariant  = "revisions.record_version_unique"
	revisionsHistoryInvariant        = "revisions.history_reconstruction"
	revisionsSequenceRepairInvariant = "revisions.sequence_repair_after_validation"
)

var revisionsPositiveInteger = regexp.MustCompile(`^[1-9][0-9]*$`)

type portableChangeSet struct {
	HistoryOrder        int64
	ActorImportContexts []revisionActorImportContext
	ChangeSetID         uuid.UUID
	IncidentID          uuid.UUID
	PortableActorID     uuid.UUID
	RuntimeActorID      uuid.UUID
	Source              string
	Reason              *string
	ClientTxnID         *string
	RequestID           *string
	CreatedAt           time.Time
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

// preparedRevisionsImport is the only state the final transaction receives.
// It holds sealed preparation streams and recorded verdicts, never a bundle
// member reader (Core 01 REQ-01-641).
type preparedRevisionsImport struct {
	actorContext                             revisionActorImportContext
	incidentID                               uuid.UUID
	runtimeActor                             uuid.UUID
	actors                                   sourceport.ActorCatalog
	sequenceInvalid, revisionIdentityInvalid bool
	verdicts                                 *incidentportability.PreparedVerdicts
	changeSets, mutations, revisions         incidentportability.PreparedRowStream
	references                               incidentportability.SortedKeyRuns
	attribution                              incidentportability.AttributionStream
}

type revisionsParseFailure struct {
	invariant string
	path      string
	identity  string
}
