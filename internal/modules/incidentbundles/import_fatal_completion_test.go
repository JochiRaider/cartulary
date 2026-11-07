package incidentbundles

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

// CompleteFailedImportForTest enters the production failure path with a forbidden
// finalizer. Report a test failure directly: a panic could be absorbed by the
// enclosing lifetime's sticky fatal result and would not prove exclusion.
func CompleteFailedImportForTest(t testing.TB, ctx context.Context, scope *reference_data.IncidentReferenceExecution, execution jobs.Execution, cause error) {
	t.Helper()
	worker := &incidentBundleWorker{jobFinalizer: forbiddenImportFinalizer{t: t}}
	worker.finishImportFailure(ctx, scope, execution, cause)
}

// AdmitPortableFailureForTest uses production admission, payload and receipt
// writes so the external composition test can use the real common finalizer.
func AdmitPortableFailureForTest(t testing.TB, pool *pgxpool.Pool, transactions JobTransactions, actor uuid.UUID, digest string) jobs.Resource {
	t.Helper()
	s := &store{pool: pool, jobTransactions: transactions}
	normalized := []byte(`{"client_txn_id":"portable-fatal"}`)
	accepted, err := s.acceptImport(t.Context(), importAcceptedParams{ActorUserID: actor, Request: importMetadataRequest{ClientTxnID: "portable-fatal", Normalized: normalized}, UploadedSHA256: digest, BundleStagingRef: BundleStagingRef{value: "portable/input"}, NormalizedRequest: normalized, Now: time.Now()})
	if err != nil {
		t.Fatal(err)
	}
	return accepted.Job
}

func FinalizePortableFailureForTest(ctx context.Context, scope *reference_data.IncidentReferenceExecution, execution jobs.Execution, pool *pgxpool.Pool, finalizer JobSuccessFinalizer, cause error) error {
	worker := &incidentBundleWorker{store: &store{pool: pool}, jobFinalizer: finalizer, now: time.Now}
	return worker.finishImportFailure(ctx, scope, execution, cause)
}

type forbiddenImportFinalizer struct {
	JobSuccessFinalizer
	t testing.TB
}

func (f forbiddenImportFinalizer) reject() (jobs.Resource, error) {
	f.t.Helper()
	f.t.Error("fatal import invoked an ordinary finalizer")
	return jobs.Resource{}, errors.New("forbidden ordinary finalization")
}

func (f forbiddenImportFinalizer) FinalizeIncidentBundleJobFailure(context.Context, JobFailureFinalization) (jobs.Resource, error) {
	return f.reject()
}
func (f forbiddenImportFinalizer) FinalizeIncidentBundleJobTimeout(context.Context, JobFailureFinalization) (jobs.Resource, error) {
	return f.reject()
}
func (f forbiddenImportFinalizer) FinalizeIncidentBundleJobCancellation(context.Context, JobCancellationFinalization) (jobs.Resource, error) {
	return f.reject()
}
