package evidence

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestEvidenceCleanupDispatcherStartsAfterActivationRunsEveryFifteenMinutesAndStops(t *testing.T) {
	sweeper := &dispatcherTestSweeper{calls: make(chan struct{}, 8)}
	observer := &dispatcherTestObserver{observations: make(chan CleanupSweepObservation, 8)}
	dispatcher, err := newCleanupDispatcher(sweeper, dispatcherTestDeleter{}, observer, func() time.Time {
		return time.Date(2026, time.August, 12, 18, 0, 0, 0, time.UTC)
	})
	if err != nil {
		t.Fatalf("compose cleanup dispatcher: %v", err)
	}
	registerDispatcherCleanup(t, dispatcher)
	if dispatcher.identity() != cleanupDispatcherIdentity || dispatcher.interval != 15*time.Minute {
		t.Fatalf("cleanup dispatcher identity/interval = %q/%s", dispatcher.identity(), dispatcher.interval)
	}
	select {
	case <-sweeper.calls:
		t.Fatal("cleanup dispatcher ran before readiness activation")
	default:
	}

	dispatcher.interval = 10 * time.Millisecond
	serving := make(chan struct{})
	if err := dispatcher.Prepare(context.Background(), serving); err != nil {
		t.Fatal(err)
	}
	select {
	case <-sweeper.calls:
		t.Fatal("prepared cleanup dispatched before serving")
	case <-time.After(20 * time.Millisecond):
	}
	close(serving)
	requireDispatcherCall(t, sweeper.calls, "initial readiness-gated sweep")
	requireDispatcherCall(t, sweeper.calls, "scheduled sweep")
	closeCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := dispatcher.Close(closeCtx); err != nil {
		t.Fatalf("close cleanup dispatcher: %v", err)
	}
	callsAfterClose := sweeper.callCount()
	time.Sleep(30 * time.Millisecond)
	if got := sweeper.callCount(); got != callsAfterClose {
		t.Fatalf("cleanup dispatcher ran after shutdown: calls=%d want=%d", got, callsAfterClose)
	}
	if len(observer.observations) != callsAfterClose {
		t.Fatalf("cleanup observer count=%d want=%d", len(observer.observations), callsAfterClose)
	}
}

func TestEvidenceCleanupDispatcherShutdownCancelsInFlightSweep(t *testing.T) {
	started := make(chan struct{})
	sweeper := &dispatcherTestSweeper{
		calls: make(chan struct{}, 1),
		run: func(ctx context.Context) (cleanupSweepResult, error) {
			close(started)
			<-ctx.Done()
			return cleanupSweepResult{}, ctx.Err()
		},
	}
	observer := &dispatcherTestObserver{observations: make(chan CleanupSweepObservation, 1)}
	dispatcher, err := newCleanupDispatcher(sweeper, dispatcherTestDeleter{}, observer, time.Now)
	if err != nil {
		t.Fatal(err)
	}
	registerDispatcherCleanup(t, dispatcher)
	if err := dispatcher.Prepare(context.Background(), servingCleanupGate()); err != nil {
		t.Fatal(err)
	}
	select {
	case <-started:
	case <-time.After(time.Second):
		t.Fatal("cleanup dispatcher did not start initial sweep")
	}
	closeCtx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()
	if err := dispatcher.Close(closeCtx); err != nil {
		t.Fatalf("shutdown did not cancel in-flight cleanup: %v", err)
	}
	select {
	case observation := <-observer.observations:
		if observation.Result != "canceled" || observation.ErrorClass != "dependency_unavailable" {
			t.Fatalf("shutdown observation = %#v", observation)
		}
	case <-time.After(time.Second):
		t.Fatal("cleanup dispatcher omitted shutdown observation")
	}
}

func TestEvidenceCleanupDispatcherRejectsIncompleteDependencies(t *testing.T) {
	observer := &dispatcherTestObserver{observations: make(chan CleanupSweepObservation, 1)}
	sweeper := &dispatcherTestSweeper{calls: make(chan struct{}, 1)}
	for name, build := range map[string]func() (*CleanupDispatcher, error){
		"sweeper": func() (*CleanupDispatcher, error) {
			return newCleanupDispatcher(nil, dispatcherTestDeleter{}, observer, time.Now)
		},
		"deleter": func() (*CleanupDispatcher, error) { return newCleanupDispatcher(sweeper, nil, observer, time.Now) },
		"observer": func() (*CleanupDispatcher, error) {
			return newCleanupDispatcher(sweeper, dispatcherTestDeleter{}, nil, time.Now)
		},
		"typed-nil sweeper": func() (*CleanupDispatcher, error) {
			return newCleanupDispatcher((*dispatcherTestSweeper)(nil), dispatcherTestDeleter{}, observer, time.Now)
		},
		"typed-nil deleter": func() (*CleanupDispatcher, error) {
			return newCleanupDispatcher(sweeper, (*dispatcherTestDeleter)(nil), observer, time.Now)
		},
		"typed-nil observer": func() (*CleanupDispatcher, error) {
			return newCleanupDispatcher(sweeper, dispatcherTestDeleter{}, (*dispatcherTestObserver)(nil), time.Now)
		},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := build(); err == nil {
				t.Fatal("incomplete cleanup dispatcher dependencies were accepted")
			}
		})
	}
	if value, err := newCleanupDispatcher(sweeper, dispatcherTestDeleter{}, observer, nil); err != nil || value == nil || value.now == nil {
		t.Fatal("valid cleanup dependencies with the default clock were rejected", err)
	}
}

type dispatcherTestSweeper struct {
	mu    sync.Mutex
	count int
	calls chan struct{}
	run   func(context.Context) (cleanupSweepResult, error)
}

func (sweeper *dispatcherTestSweeper) SweepFailedUnattachedBlobs(
	ctx context.Context,
	_ cleanupObjectDeleter,
	_ time.Time,
) (cleanupSweepResult, error) {
	sweeper.mu.Lock()
	sweeper.count++
	sweeper.mu.Unlock()
	select {
	case sweeper.calls <- struct{}{}:
	case <-ctx.Done():
		return cleanupSweepResult{}, ctx.Err()
	}
	if sweeper.run != nil {
		return sweeper.run(ctx)
	}
	return cleanupSweepResult{HealthSnapshotValid: true}, nil
}

func (sweeper *dispatcherTestSweeper) callCount() int {
	sweeper.mu.Lock()
	defer sweeper.mu.Unlock()
	return sweeper.count
}

type dispatcherTestObserver struct {
	observations chan CleanupSweepObservation
}

func (observer *dispatcherTestObserver) ObserveCleanupSweep(_ context.Context, observation CleanupSweepObservation) {
	// A failed assertion may stop draining notifications. Successful paths still
	// require every observation, but teardown must not block on this test channel.
	select {
	case observer.observations <- observation:
	default:
	}
}

type dispatcherTestDeleter struct{}

func (dispatcherTestDeleter) DeleteObject(context.Context, string) error { return nil }

func requireDispatcherCall(t testing.TB, calls <-chan struct{}, label string) {
	t.Helper()
	select {
	case <-calls:
	case <-time.After(time.Second):
		t.Fatalf("cleanup dispatcher omitted %s", label)
	}
}

func servingCleanupGate() <-chan struct{} { gate := make(chan struct{}); close(gate); return gate }

func registerDispatcherCleanup(t testing.TB, dispatcher *CleanupDispatcher) {
	t.Helper()
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := dispatcher.Close(ctx); err != nil {
			t.Errorf("settle cleanup dispatcher: %v", err)
		}
	})
}
