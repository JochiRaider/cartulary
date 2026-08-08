package collaborationsupport

import (
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/collaboration"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

const TestJobKind = "test_platform.generic_v1"

var testJobDefinitions = []jobs.ExtensionJobContract{
	{
		OwnerProfileID: "test_platform",
		JobKind:        TestJobKind,
		ProgressUnitID: "test_platform.generic.operation.v1",
		OperationKind:  "test_platform.generic",
		WorkerKind:     "test_platform.worker_v1",
		ContractSHA256: strings.Repeat("b", 64),
		ProofRequired:  true,
		MaxProofBytes:  4096,
	},
	{
		OwnerProfileID: "test_profile",
		JobKind:        "test_profile.run_v1",
		ProgressUnitID: "test_profile.run.attempt.v1",
		OperationKind:  "test_profile.run",
		WorkerKind:     "test_profile.worker_v1",
		ContractSHA256: strings.Repeat("a", 64),
		ProofRequired:  true,
		MaxProofBytes:  4096,
	},
}

func TestJobDefinitions() []jobs.ExtensionJobContract {
	return append([]jobs.ExtensionJobContract(nil), testJobDefinitions...)
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
	return NewJobTransactionsWithDefinitions(testJobDefinitions...)
}

func NewJobTransactionsWithDefinitions(definitions ...jobs.ExtensionJobContract) *jobs.TransactionService {
	ownerPorts := NewJobOwnerTransactionAdapters()
	service, err := jobs.NewTransactionService(NewIntentAdapters(), jobs.OwnerTransactionPorts{
		RouteIdempotency:      ownerPorts,
		ExtensionCancellation: ownerPorts,
	}, definitions...)
	if err != nil {
		panic(err)
	}
	return service
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
