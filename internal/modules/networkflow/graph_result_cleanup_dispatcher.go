package networkflow

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/graphprojection"
)

const (
	graphResultCleanupDispatcherIdentity = "network_flow_activity.graph_result_cleanup.v1"
	graphResultCleanupBaseCadence        = 5 * time.Minute
	graphResultCleanupContinuationDelay  = 5 * time.Second
	graphResultCleanupRetryDelay         = 30 * time.Second
)

// GraphResultCleanupDispatcher is Network Flow's private lifecycle component.
// It exposes no route or operator command; application assembly only starts and
// stops it with the serving epoch.
type GraphResultCleanupDispatcher struct {
	sweeper           graphResultCleanupSweeper
	telemetry         GraphTelemetryObserver
	now               func() time.Time
	onUnexpectedLoss  func()
	baseCadence       time.Duration
	continuationDelay time.Duration
	retryDelay        time.Duration

	mu     sync.Mutex
	runMu  sync.Mutex
	cancel context.CancelFunc
	done   chan struct{}
}

func (module *Module) NewGraphResultCleanupDispatcher(onUnexpectedLoss func()) (*GraphResultCleanupDispatcher, error) {
	if module == nil || module.store == nil || module.store.pool == nil {
		return nil, errors.New("compose Network Flow graph-result cleanup dispatcher: module persistence is required")
	}
	sweeper, err := newGraphResultCleanupService(module.store.pool, module.store)
	if err != nil {
		return nil, err
	}
	dispatcher, err := newGraphResultCleanupDispatcher(sweeper, module.now, onUnexpectedLoss)
	if err != nil {
		return nil, err
	}
	dispatcher.telemetry = module.graphTelemetry
	return dispatcher, nil
}

func newGraphResultCleanupDispatcher(
	sweeper graphResultCleanupSweeper,
	now func() time.Time,
	onUnexpectedLoss func(),
) (*GraphResultCleanupDispatcher, error) {
	if sweeper == nil || onUnexpectedLoss == nil {
		return nil, errors.New("compose Network Flow graph-result cleanup dispatcher: dependencies are required")
	}
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	return &GraphResultCleanupDispatcher{
		sweeper:           sweeper,
		now:               now,
		onUnexpectedLoss:  onUnexpectedLoss,
		baseCadence:       graphResultCleanupBaseCadence,
		continuationDelay: graphResultCleanupContinuationDelay,
		retryDelay:        graphResultCleanupRetryDelay,
	}, nil
}

func (*GraphResultCleanupDispatcher) identity() string {
	return graphResultCleanupDispatcherIdentity
}

func (dispatcher *GraphResultCleanupDispatcher) Start(parent context.Context) error {
	if dispatcher == nil || dispatcher.sweeper == nil || dispatcher.now == nil || dispatcher.onUnexpectedLoss == nil ||
		dispatcher.baseCadence <= 0 || dispatcher.continuationDelay <= 0 || dispatcher.retryDelay <= 0 {
		return errors.New("network flow graph-result cleanup dispatcher is not configured")
	}
	if parent == nil {
		parent = context.Background()
	}
	dispatcher.mu.Lock()
	defer dispatcher.mu.Unlock()
	if dispatcher.cancel != nil {
		return nil
	}
	ctx, cancel := context.WithCancel(parent)
	dispatcher.cancel = cancel
	dispatcher.done = make(chan struct{})
	go dispatcher.run(ctx, dispatcher.done)
	return nil
}

func (dispatcher *GraphResultCleanupDispatcher) Close(ctx context.Context) error {
	if dispatcher == nil {
		return nil
	}
	dispatcher.mu.Lock()
	cancel := dispatcher.cancel
	done := dispatcher.done
	dispatcher.cancel = nil
	dispatcher.done = nil
	dispatcher.mu.Unlock()
	if cancel == nil {
		return nil
	}
	cancel()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (dispatcher *GraphResultCleanupDispatcher) runOnce(
	ctx context.Context,
	cursor *graphprojection.ResultCleanupCandidateV2,
) (graphResultCleanupSweepResult, error) {
	dispatcher.runMu.Lock()
	defer dispatcher.runMu.Unlock()
	started := time.Now()
	result, err := dispatcher.sweeper.SweepGraphResults(ctx, dispatcher.now().UTC(), cursor)
	telemetryResult, errorClass := graphTelemetryOutcomeForError(err)
	observeGraphCleanupSafely(context.WithoutCancel(ctx), dispatcher.telemetry, GraphCleanupTelemetryObservation{
		Operation: graphTelemetryOperationCleanup, Result: telemetryResult, ErrorClass: errorClass,
		Duration:      nonnegativeGraphTelemetryDuration(started),
		DeletedLeases: result.DeletedLeases, DeletedResults: result.DeletedResults,
		HealthSnapshotValid: result.HealthSnapshotValid, EligibleResultBacklog: result.EligibleResultBacklog,
		OldestEligibleResultAge: cloneDuration(result.OldestEligibleResultAge),
	})
	return result, err
}

func cloneDuration(value *time.Duration) *time.Duration {
	if value == nil {
		return nil
	}
	cloned := *value
	return &cloned
}

func (dispatcher *GraphResultCleanupDispatcher) run(ctx context.Context, done chan<- struct{}) {
	expectedStop := false
	defer close(done)
	defer func() {
		if recovered := recover(); recovered != nil || !expectedStop {
			dispatcher.onUnexpectedLoss()
		}
	}()

	timer := time.NewTimer(0)
	defer timer.Stop()
	var cursor *graphprojection.ResultCleanupCandidateV2
	restartFromBeginning := false
	for {
		select {
		case <-ctx.Done():
			expectedStop = true
			return
		case <-timer.C:
		}

		inputCursor := cloneCleanupCandidate(cursor)
		result, err := dispatcher.runOnce(ctx, inputCursor)
		if ctx.Err() != nil {
			expectedStop = true
			return
		}
		if result.DeletedLeases > 0 && inputCursor != nil {
			restartFromBeginning = true
		}
		if result.NextCursor != nil {
			cursor = cloneCleanupCandidate(result.NextCursor)
		}

		delay := dispatcher.baseCadence
		switch {
		case err != nil:
			delay = dispatcher.retryDelay
		case result.HasMore:
			delay = dispatcher.continuationDelay
		case result.Exhausted && restartFromBeginning:
			cursor = nil
			restartFromBeginning = false
			delay = dispatcher.continuationDelay
		default:
			cursor = nil
			restartFromBeginning = false
		}
		timer.Reset(delay)
	}
}
