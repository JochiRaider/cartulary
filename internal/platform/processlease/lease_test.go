package processlease

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestLeaseAcquisitionDeadlineClassification(t *testing.T) {
	deadline := time.Now().Add(time.Hour)
	ctx, cancel := context.WithDeadline(context.Background(), deadline)
	defer cancel()
	transportTimeout := &net.OpError{Op: "read", Net: "tcp", Err: os.ErrDeadlineExceeded}
	transportFailure := errors.New("connection failed")
	explicitDeadline := fmt.Errorf("driver: %w", context.DeadlineExceeded)
	for _, err := range []error{transportTimeout, transportFailure, explicitDeadline} {
		if got := acquisitionFailure(ctx, err, deadline.Add(-time.Nanosecond)); got != err {
			t.Fatalf("pre-deadline transport failure was hidden: %v", got)
		}
		for _, observedAt := range []time.Time{deadline, deadline.Add(time.Nanosecond)} {
			if ctx.Err() != nil {
				t.Fatal("fixture must precede context cancellation publication")
			}
			if got := acquisitionFailure(ctx, err, observedAt); !errors.Is(got, ErrApplicationProcessActive) {
				t.Fatalf("elapsed acquisition deadline lost its required classification: %v", got)
			}
		}
	}
	if got := acquisitionFailure(context.Background(), transportTimeout, deadline); got != transportTimeout {
		t.Fatalf("unrelated transport timeout became acquisition expiry: %v", got)
	}
	cancel()
	if got := acquisitionFailure(ctx, context.Canceled, deadline.Add(-time.Second)); !errors.Is(got, context.Canceled) {
		t.Fatalf("pre-deadline cancellation became acquisition expiry: %v", got)
	}
}

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

// scriptedProofSession answers each ownership proof with the next scripted
// result, so a monitor walks a fixed sequence of transitions whatever its
// timing.
type scriptedProofSession struct {
	mu     sync.Mutex
	proofs []Proof
}

func (s *scriptedProofSession) Identity() string                         { return "owned-session" }
func (s *scriptedProofSession) TryAcquire(context.Context) (bool, error) { return true, nil }
func (s *scriptedProofSession) Prove(context.Context) Proof {
	s.mu.Lock()
	defer s.mu.Unlock()
	if len(s.proofs) == 0 {
		return ProofLost
	}
	proof := s.proofs[0]
	s.proofs = s.proofs[1:]
	return proof
}
func (s *scriptedProofSession) Release(context.Context) error { return nil }
func (s *scriptedProofSession) Close()                        {}

type scriptedProofBackend struct{ session *scriptedProofSession }

func (b scriptedProofBackend) Open(context.Context) (Session, error) { return b.session, nil }

// TestLeaseEventsDeliverTheLatestTransition covers Core 04 REQ-04-145 and
// EXT-REQ-213: confirmed loss must reach the fatal lifecycle, so a consumer
// that fell behind a burst of ownership transitions still receives the
// lease's latest transition, lost in particular, after earlier transitions
// overflowed the event buffer. Delivered events keep their order.
func TestLeaseEventsDeliverTheLatestTransition(t *testing.T) {
	var script []Proof
	for range 8 {
		script = append(script, ProofUncertain, ProofContinuous)
	}
	script = append(script, ProofLost)
	// Each uncertain proof follows a held state and is answered by the next
	// tick, so the detection deadline never expires; only the scripted proof
	// confirms loss.
	lease, err := Acquire(context.Background(), scriptedProofBackend{&scriptedProofSession{proofs: script}}, time.Second, 20*time.Millisecond)
	if err != nil {
		t.Fatal(err)
	}
	monitorCtx, cancel := context.WithCancel(context.Background())
	t.Cleanup(func() { cancel(); lease.Close() })
	lease.StartMonitor(monitorCtx)
	lease.mu.RLock()
	monitorDone := lease.monitorDone
	lease.mu.RUnlock()
	select {
	case <-monitorDone:
	case <-time.After(10 * time.Second):
		t.Fatalf("the monitor did not reach loss: state %s", lease.State())
	}
	if lease.State() != StateLost {
		t.Fatalf("state = %s; want lost", lease.State())
	}
	var events []Event
	for drained := false; !drained; {
		select {
		case event := <-lease.Events():
			events = append(events, event)
		default:
			drained = true
		}
	}
	if len(events) == 0 || events[len(events)-1].State != StateLost || events[len(events)-1].Previous != StateUncertain {
		t.Fatalf("delivered events %v; the latest transition must be uncertain -> lost", events)
	}
	for index := 1; index < len(events); index++ {
		if events[index].Previous != events[index-1].State {
			t.Fatalf("delivered events %v are out of order at %d", events, index)
		}
	}
}
