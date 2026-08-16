package networkflow

import (
	"context"
	"errors"
	"sync"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/graphprojection"
)

func TestGraphResultCleanupDispatcherLifecycleSchedulingAndFatalLoss_Unit(t *testing.T) {
	t.Run("immediate continuation restart cadence and shutdown", func(t *testing.T) {
		candidate := &graphprojection.ResultCleanupCandidateV2{
			ProjectionResultID: "gpres_aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
			PublishedAt:        time.Date(2026, 8, 16, 17, 0, 0, 0, time.UTC),
		}
		sweeper := &graphResultCleanupDispatcherTestSweeper{
			calls: make(chan *graphprojection.ResultCleanupCandidateV2, 8),
			responses: []graphResultCleanupDispatcherTestResponse{
				{result: graphResultCleanupSweepResult{HasMore: true, NextCursor: candidate}},
				{result: graphResultCleanupSweepResult{DeletedLeases: 1, Exhausted: true, NextCursor: candidate}},
				{result: graphResultCleanupSweepResult{Exhausted: true}},
				{result: graphResultCleanupSweepResult{Exhausted: true}},
			},
		}
		fatal := make(chan struct{}, 1)
		dispatcher, err := newGraphResultCleanupDispatcher(sweeper, func() time.Time { return candidate.PublishedAt }, func() { fatal <- struct{}{} })
		if err != nil {
			t.Fatal(err)
		}
		if dispatcher.identity() != graphResultCleanupDispatcherIdentity || dispatcher.baseCadence != 5*time.Minute ||
			dispatcher.continuationDelay != 5*time.Second {
			t.Fatalf("dispatcher contract drifted: identity=%q cadence=%s continuation=%s", dispatcher.identity(), dispatcher.baseCadence, dispatcher.continuationDelay)
		}
		dispatcher.baseCadence = 25 * time.Millisecond
		dispatcher.continuationDelay = 5 * time.Millisecond
		dispatcher.retryDelay = 5 * time.Millisecond
		if err := dispatcher.Start(context.Background()); err != nil {
			t.Fatal(err)
		}
		if cursor := requireGraphCleanupDispatcherCall(t, sweeper.calls); cursor != nil {
			t.Fatalf("immediate sweep cursor = %#v", cursor)
		}
		if cursor := requireGraphCleanupDispatcherCall(t, sweeper.calls); cursor == nil || cursor.ProjectionResultID != candidate.ProjectionResultID {
			t.Fatalf("paced continuation cursor = %#v", cursor)
		}
		if cursor := requireGraphCleanupDispatcherCall(t, sweeper.calls); cursor != nil {
			t.Fatalf("lease-drain restart cursor = %#v", cursor)
		}
		started := time.Now()
		if cursor := requireGraphCleanupDispatcherCall(t, sweeper.calls); cursor != nil {
			t.Fatalf("base cadence cursor = %#v", cursor)
		}
		if elapsed := time.Since(started); elapsed < 15*time.Millisecond {
			t.Fatalf("base cadence busy-looped after %s", elapsed)
		}
		closeCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := dispatcher.Close(closeCtx); err != nil {
			t.Fatal(err)
		}
		if err := dispatcher.Start(context.Background()); err != nil {
			t.Fatalf("restart cleanup dispatcher: %v", err)
		}
		requireGraphCleanupDispatcherCall(t, sweeper.calls)
		if err := dispatcher.Close(closeCtx); err != nil {
			t.Fatalf("close restarted cleanup dispatcher: %v", err)
		}
		if sweeper.maximumActive() != 1 {
			t.Fatalf("cleanup sweeps overlapped: maximum_active=%d", sweeper.maximumActive())
		}
		select {
		case <-fatal:
			t.Fatal("normal shutdown reported unexpected component loss")
		default:
		}
	})

	t.Run("transient failure uses bounded retry", func(t *testing.T) {
		sweeper := &graphResultCleanupDispatcherTestSweeper{
			calls: make(chan *graphprojection.ResultCleanupCandidateV2, 4),
			responses: []graphResultCleanupDispatcherTestResponse{
				{err: errors.New("transient database failure")},
				{result: graphResultCleanupSweepResult{Exhausted: true}},
			},
		}
		dispatcher, err := newGraphResultCleanupDispatcher(sweeper, time.Now, func() {})
		if err != nil {
			t.Fatal(err)
		}
		dispatcher.retryDelay = 10 * time.Millisecond
		dispatcher.baseCadence = time.Hour
		if err := dispatcher.Start(context.Background()); err != nil {
			t.Fatal(err)
		}
		requireGraphCleanupDispatcherCall(t, sweeper.calls)
		started := time.Now()
		requireGraphCleanupDispatcherCall(t, sweeper.calls)
		if elapsed := time.Since(started); elapsed < 5*time.Millisecond || elapsed > 500*time.Millisecond {
			t.Fatalf("bounded retry delay = %s", elapsed)
		}
		closeCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := dispatcher.Close(closeCtx); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("shutdown cancels in-flight sweep", func(t *testing.T) {
		started := make(chan struct{})
		sweeper := &graphResultCleanupDispatcherTestSweeper{
			calls: make(chan *graphprojection.ResultCleanupCandidateV2, 1),
			run: func(ctx context.Context) (graphResultCleanupSweepResult, error) {
				close(started)
				<-ctx.Done()
				return graphResultCleanupSweepResult{}, ctx.Err()
			},
		}
		dispatcher, err := newGraphResultCleanupDispatcher(sweeper, time.Now, func() {})
		if err != nil {
			t.Fatal(err)
		}
		if err := dispatcher.Start(context.Background()); err != nil {
			t.Fatal(err)
		}
		select {
		case <-started:
		case <-time.After(time.Second):
			t.Fatal("immediate cleanup sweep did not start")
		}
		closeCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := dispatcher.Close(closeCtx); err != nil {
			t.Fatalf("in-flight cleanup did not stop: %v", err)
		}
	})

	t.Run("panic is required-component loss", func(t *testing.T) {
		sweeper := &graphResultCleanupDispatcherTestSweeper{
			calls: make(chan *graphprojection.ResultCleanupCandidateV2, 1),
			run: func(context.Context) (graphResultCleanupSweepResult, error) {
				panic("unexpected cleanup loop loss")
			},
		}
		fatal := make(chan struct{}, 1)
		dispatcher, err := newGraphResultCleanupDispatcher(sweeper, time.Now, func() { fatal <- struct{}{} })
		if err != nil {
			t.Fatal(err)
		}
		if err := dispatcher.Start(context.Background()); err != nil {
			t.Fatal(err)
		}
		select {
		case <-fatal:
		case <-time.After(time.Second):
			t.Fatal("unexpected cleanup loop loss was not fatal")
		}
		closeCtx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := dispatcher.Close(closeCtx); err != nil {
			t.Fatal(err)
		}
	})
}

type graphResultCleanupDispatcherTestResponse struct {
	result graphResultCleanupSweepResult
	err    error
}

type graphResultCleanupDispatcherTestSweeper struct {
	mu        sync.Mutex
	responses []graphResultCleanupDispatcherTestResponse
	run       func(context.Context) (graphResultCleanupSweepResult, error)
	calls     chan *graphprojection.ResultCleanupCandidateV2
	active    int
	maxActive int
}

func (sweeper *graphResultCleanupDispatcherTestSweeper) SweepGraphResults(
	ctx context.Context,
	_ time.Time,
	cursor *graphprojection.ResultCleanupCandidateV2,
) (graphResultCleanupSweepResult, error) {
	sweeper.mu.Lock()
	sweeper.active++
	if sweeper.active > sweeper.maxActive {
		sweeper.maxActive = sweeper.active
	}
	sweeper.mu.Unlock()
	defer func() {
		sweeper.mu.Lock()
		sweeper.active--
		sweeper.mu.Unlock()
	}()
	sweeper.calls <- cloneCleanupCandidate(cursor)
	if sweeper.run != nil {
		return sweeper.run(ctx)
	}
	sweeper.mu.Lock()
	defer sweeper.mu.Unlock()
	if len(sweeper.responses) == 0 {
		return graphResultCleanupSweepResult{Exhausted: true}, nil
	}
	response := sweeper.responses[0]
	sweeper.responses = sweeper.responses[1:]
	return response.result, response.err
}

func (sweeper *graphResultCleanupDispatcherTestSweeper) maximumActive() int {
	sweeper.mu.Lock()
	defer sweeper.mu.Unlock()
	return sweeper.maxActive
}

func requireGraphCleanupDispatcherCall(
	t testing.TB,
	calls <-chan *graphprojection.ResultCleanupCandidateV2,
) *graphprojection.ResultCleanupCandidateV2 {
	t.Helper()
	select {
	case cursor := <-calls:
		return cursor
	case <-time.After(time.Second):
		t.Fatal("cleanup dispatcher omitted expected sweep")
		return nil
	}
}
