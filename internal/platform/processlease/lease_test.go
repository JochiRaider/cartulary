package processlease

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

type drainingProofSession struct {
	started  chan context.Context
	complete chan struct{}
	failed   atomic.Bool
}

func (s *drainingProofSession) Identity() string                         { return "owned-session" }
func (s *drainingProofSession) TryAcquire(context.Context) (bool, error) { return true, nil }
func (s *drainingProofSession) Prove(ctx context.Context) Proof {
	s.started <- ctx
	select {
	case <-s.complete:
	case <-ctx.Done():
	}
	if ctx.Err() != nil {
		s.failed.Store(true)
		return ProofLost
	}
	return ProofContinuous
}
func (s *drainingProofSession) Release(context.Context) error {
	if s.failed.Load() {
		return ErrLeaseLost
	}
	return nil
}
func (s *drainingProofSession) Close() {}

type drainingProofBackend struct{ session *drainingProofSession }

func (b drainingProofBackend) Open(context.Context) (Session, error) { return b.session, nil }

func monitorWithPendingProof(t *testing.T, lossDetection time.Duration) (*Lease, context.Context, context.CancelFunc, func()) {
	t.Helper()
	session := &drainingProofSession{started: make(chan context.Context, 1), complete: make(chan struct{})}
	lease, err := Acquire(context.Background(), drainingProofBackend{session}, time.Second, lossDetection)
	if err != nil {
		t.Fatal(err)
	}
	monitorCtx, cancel := context.WithCancel(context.Background())
	var once sync.Once
	complete := func() { once.Do(func() { close(session.complete) }) }
	t.Cleanup(func() { cancel(); complete(); lease.Close() })
	lease.StartMonitor(monitorCtx)
	select {
	case proofCtx := <-session.started:
		return lease, proofCtx, cancel, complete
	case <-time.After(5 * time.Second):
		t.Fatal("ownership proof did not start")
		return nil, nil, nil, nil
	}
}

func TestLeaseMonitorShutdownDrainsOwnershipProof(t *testing.T) {
	lease, proofCtx, cancel, complete := monitorWithPendingProof(t, time.Second)
	cancel()
	if err := proofCtx.Err(); err != nil {
		t.Errorf("monitor shutdown canceled the owned session query: %v", err)
	}
	complete()
	if err := lease.Release(context.Background()); err != nil {
		t.Fatalf("release after draining proof: %v", err)
	}
	if lease.State() != StateReleased {
		t.Fatalf("state = %s; want released", lease.State())
	}
}

func TestLeaseMonitorShutdownPreservesProofDeadline(t *testing.T) {
	lease, proofCtx, cancel, _ := monitorWithPendingProof(t, 100*time.Millisecond)
	cancel()
	select {
	case <-proofCtx.Done():
		if !errors.Is(proofCtx.Err(), context.DeadlineExceeded) {
			t.Errorf("proof must retain its own bounded deadline: %v", proofCtx.Err())
		}
	case <-time.After(5 * time.Second):
		t.Fatal("ownership proof outlived its deadline")
	}
	if err := lease.Release(context.Background()); err == nil {
		t.Fatal("failed ownership proof was reported as successful release")
	}
}
