package networkflow

import (
	"context"
	"encoding/json"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/admission"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/google/uuid"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

// NewGraphViewAuthorityTestHandler exercises the production handler with an
// explicitly controlled finalization boundary, without starting a worker.
func NewGraphViewAuthorityTestHandler(store *store, manager GraphViewJobManager, finalizer GraphViewJobFinalizer) jobs.HandlerFunc {
	module := &Module{store: store, limits: store.limits, now: time.Now,
		graphProjection: newGraphProjectionAdapter(), jobManager: manager, jobFinalizer: finalizer}
	module.graphComposer = &graphSourceComposer{store: store, limits: store.limits, graphProjection: module.graphProjection}
	return module.handleGraphViewMaterialization
}

// NewSavedGraphCreateCommandForTest exposes only a command invocation to the
// external transaction fixture, without exporting the application in builds.
func NewSavedGraphCreateCommandForTest(store *store, transactions GraphViewJobTransactions, runner GraphViewJobRunner, incidentID, actorID uuid.UUID, raw json.RawMessage) func(context.Context, string, string) error {
	application := &savedGraphApplication{store: store, incidentAccess: admission.NewChecker(store.pool), receipts: savedGraphReceiptAdapter{reader: authn.NewStore(store.pool)}, graphViewJobs: transactions, jobRunner: runner, now: time.Now}
	return func(ctx context.Context, txn, name string) error {
		semantic, failure := decodeGraphSemanticRequest(raw, store.limits)
		if failure != nil {
			return failure
		}
		_, err := application.commitGraphViewCreate(ctx, incidentID, actorID, graphViewCreateRequest{ClientTxnID: txn, DisplayName: name, Semantic: semantic}, "notification-test")
		return err
	}
}
