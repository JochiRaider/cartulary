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
	if err := module.prepareActiveApplications(); err != nil {
		return nil, err
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

type graphCleanupSchedule struct {
	cursor               *graphprojection.ResultCleanupCandidateV2
	restartFromBeginning bool
	delay                time.Duration
}

func (dispatcher *GraphResultCleanupDispatcher) runOnce(ctx context.Context, schedule *graphCleanupSchedule) (graphResultCleanupSweepResult, error) {
	dispatcher.runMu.Lock()
	defer dispatcher.runMu.Unlock()
	started := time.Now()
	inputCursor := cloneCleanupCandidate(schedule.cursor)
	result, err := dispatcher.sweeper.SweepGraphResults(ctx, dispatcher.now().UTC(), inputCursor)
	if result.DeletedLeases > 0 && inputCursor != nil {
		schedule.restartFromBeginning = true
	}
	if result.NextCursor != nil {
		schedule.cursor = cloneCleanupCandidate(result.NextCursor)
	}
	schedule.delay = dispatcher.baseCadence
	continuation := false
	switch {
	case err != nil:
		schedule.delay = dispatcher.retryDelay
	case result.HasMore:
		continuation = true
	case result.Exhausted && schedule.restartFromBeginning:
		schedule.cursor = nil
		schedule.restartFromBeginning = false
		continuation = true
	default:
		schedule.cursor = nil
		schedule.restartFromBeginning = false
	}
	if continuation {
		schedule.delay = dispatcher.continuationDelay
	}
	telemetryResult, errorClass := graphTelemetryOutcomeForError(err)
	observeGraphCleanupSafely(context.WithoutCancel(ctx), dispatcher.telemetry, GraphCleanupTelemetryObservation{
		Operation: graphTelemetryOperationCleanup, Result: telemetryResult, ErrorClass: errorClass,
		Duration:      nonnegativeGraphTelemetryDuration(started),
		DeletedLeases: result.DeletedLeases, DeletedResults: result.DeletedResults,
		Examined: result.Examined, Continuation: continuation,
	})
	return result, err
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
	schedule := &graphCleanupSchedule{}
	for {
		select {
		case <-ctx.Done():
			expectedStop = true
			return
		case <-timer.C:
		}

		_, _ = dispatcher.runOnce(ctx, schedule)
		if ctx.Err() != nil {
			expectedStop = true
			return
		}
		timer.Reset(schedule.delay)
	}
}
