package collaborationsupport

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"sort"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

const TestJobKind = "test_platform.generic_v1"

var testJobDefinitions = []jobs.Definition{
	{
		JobKind:        TestJobKind,
		ProgressUnitID: "test_platform.generic.operation.v1",
		HandlerName:    "test_platform.worker_v1",
	},
	{
		JobKind:        "test_profile.run_v1",
		ProgressUnitID: "test_profile.run.attempt.v1",
		HandlerName:    "test_profile.worker_v1",
		Extension: &jobs.ExtensionPolicy{
			OwnerProfileID: "test_profile",
			OperationKind:  "test_profile.run",
			ContractSHA256: strings.Repeat("a", 64),
			ProofRequired:  true,
			MaxProofBytes:  4096,
		},
	},
}

var testHandlerNames = []string{
	"test.claim", "test.complete", "test.concurrent", "test.duplicate",
	"test.error", "test.exhausted", "test.exhaustion", "test.nil",
	"test.panic", "test.recover",
}

func TestJobKindForHandler(handlerName string) string {
	return "test_platform." + strings.ReplaceAll(handlerName, ".", "_") + "_v1"
}

func TestJobDefinitions() []jobs.Definition {
	definitions := make([]jobs.Definition, 0, len(testJobDefinitions)+len(testHandlerNames))
	for _, definition := range testJobDefinitions {
		clone := definition
		if definition.Extension != nil {
			policy := *definition.Extension
			policy.ResourceRefs = append([]jobs.ExtensionResourceRefContract(nil), definition.Extension.ResourceRefs...)
			clone.Extension = &policy
		}
		definitions = append(definitions, clone)
	}
	for _, handlerName := range testHandlerNames {
		token := strings.ReplaceAll(handlerName, ".", "_")
		definitions = append(definitions, jobs.Definition{
			JobKind:        TestJobKindForHandler(handlerName),
			ProgressUnitID: "test_platform." + token + ".operation.v1",
			HandlerName:    handlerName,
		})
	}
	return definitions
}

func TestWorkerRuntimeContracts(definitions []jobs.Definition) []jobs.WorkerRuntimeContract {
	byWorker := map[string]*jobs.WorkerRuntimeContract{}
	for _, definition := range definitions {
		profileID := "base"
		if definition.Extension != nil {
			profileID = definition.Extension.OwnerProfileID
		}
		contract := byWorker[definition.HandlerName]
		if contract == nil {
			contract = &jobs.WorkerRuntimeContract{
				ProfileID: profileID, WorkerKind: definition.HandlerName,
				MaxActiveAttemptsPerProcess: 8,
			}
			byWorker[definition.HandlerName] = contract
		}
		contract.JobKinds = append(contract.JobKinds, definition.JobKind)
	}
	workerKinds := make([]string, 0, len(byWorker))
	for workerKind := range byWorker {
		workerKinds = append(workerKinds, workerKind)
	}
	sort.Strings(workerKinds)
	result := make([]jobs.WorkerRuntimeContract, 0, len(workerKinds))
	for _, workerKind := range workerKinds {
		contract := *byWorker[workerKind]
		sort.Strings(contract.JobKinds)
		result = append(result, contract)
	}
	return result
}

// IntentAdapters supplies the same narrow source-to-Collaboration translation
// used by application composition to service-backed tests.
type IntentAdapters struct {
	appender collaboration.IntentAppender
}

func NewIntentAdapters() IntentAdapters {
	return IntentAdapters{appender: collaboration.NewIntentAppender()}
}

func NewJobTransactions() *jobs.TransactionService {
	return NewJobTransactionsWithDefinitions(TestJobDefinitions()...)
}

func NewJobCatalog() *jobs.Catalog {
	catalog, err := jobs.NewCatalog(TestJobDefinitions())
	if err != nil {
		panic(err)
	}
	return catalog
}

func NewJobTransactionsForCatalog(catalog *jobs.Catalog, workerContractSets ...[]jobs.WorkerRuntimeContract) *jobs.TransactionService {
	ownerPorts := NewJobOwnerTransactionAdapters()
	workerContracts := TestWorkerRuntimeContracts(TestJobDefinitions())
	if len(workerContractSets) > 0 {
		workerContracts = workerContractSets[0]
	}
	selection, err := jobs.FullRuntimeSelection(catalog, workerContracts)
	if err != nil {
		panic(err)
	}
	service, err := jobs.NewTransactionService(NewIntentAdapters(), jobs.OwnerTransactionPorts{
		RouteIdempotency:      ownerPorts,
		ExtensionCancellation: ownerPorts,
	}, catalog, selection)
	if err != nil {
		panic(err)
	}
	return service
}

func NewJobTransactionsWithDefinitions(definitions ...jobs.Definition) *jobs.TransactionService {
	catalog, err := jobs.NewCatalog(definitions)
	if err != nil {
		panic(err)
	}
	return NewJobTransactionsForCatalog(catalog)
}

// JobOwnerTransactionAdapters provides service-backed tests with the same
// narrow owner boundaries used by application composition.
type JobOwnerTransactionAdapters struct{}

func NewJobOwnerTransactionAdapters() JobOwnerTransactionAdapters {
	return JobOwnerTransactionAdapters{}
}

func (JobOwnerTransactionAdapters) LookupRouteIdempotencyTx(ctx context.Context, tx pgx.Tx, key jobs.RouteIdempotencyKey) (jobs.RouteIdempotencyRecord, bool, error) {
	record, err := authn.GetRouteIdempotencyTx(ctx, tx, testAuthRouteKey(key))
	if errors.Is(err, authn.ErrNotFound) {
		return jobs.RouteIdempotencyRecord{}, false, nil
	}
	if err != nil {
		return jobs.RouteIdempotencyRecord{}, false, err
	}
	return jobs.RouteIdempotencyRecord{
		RequestHash:  append([]byte(nil), record.RequestHash...),
		ResponseJSON: append([]byte(nil), record.ResponseJSON...),
	}, true, nil
}

func (JobOwnerTransactionAdapters) CommitRouteIdempotencyTx(ctx context.Context, tx pgx.Tx, key jobs.RouteIdempotencyKey, requestHash []byte, statusCode int, payload any) error {
	return authn.InsertRouteIdempotencyPayload(ctx, tx, testAuthRouteKey(key), nil, requestHash, statusCode, payload)
}

func (JobOwnerTransactionAdapters) UpdateFinalIdempotencyOutcomeTx(ctx context.Context, tx pgx.Tx, key jobs.RouteIdempotencyKey, requestHash []byte, resource jobs.Resource) (bool, error) {
	return authn.UpdateRouteIdempotencyPayload(ctx, tx, testAuthRouteKey(key), requestHash, resource)
}

func (JobOwnerTransactionAdapters) AppendExtensionCancellationObservationTx(ctx context.Context, tx pgx.Tx, observation jobs.ExtensionCancellationObservation) error {
	key := observation.Key
	identity := key.RouteKey + "\x00" + key.ActorUserID.String() + "\x00" + key.ScopeKey + "\x00" + key.ClientTxnID
	digest := sha256.Sum256([]byte(identity))
	_, err := tx.Exec(ctx, `
INSERT INTO extension_job_cancellation_observations (
    cancellation_request_id, job_id, observed_at, observed_before_final_commit
) VALUES ($1, $2, $3, TRUE)
`, fmt.Sprintf("cancel:%x", digest[:]), observation.JobID, observation.ObservedAt.UTC())
	return err
}

func testAuthRouteKey(key jobs.RouteIdempotencyKey) authn.RouteIdempotencyKey {
	return authn.RouteIdempotencyKey{
		RouteKey: key.RouteKey, ActorUserID: key.ActorUserID,
		ScopeKey: key.ScopeKey, ClientTxnID: key.ClientTxnID,
	}
}

func (a IntentAdapters) AppendProgressIntentTx(ctx context.Context, tx pgx.Tx, source jobs.ProgressIntent) error {
	intent, err := collaboration.NewEventIntent(
		source.IntentKey,
		source.IncidentID,
		collaboration.EventFamilyJobProgress,
		source.CanonicalPayload,
		source.SourceIdentity,
		0,
		source.CreatedAt,
	)
	if err != nil {
		return err
	}
	return a.appender.AppendIntentTx(ctx, tx, intent)
}
