package processlease

import (
	"context"
	"errors"
	"sync"
	"time"
)

type State string

const (
	StateUnacquired State = "unacquired"
	StateAcquiring  State = "acquiring"
	StateHeld       State = "held"
	StateUncertain  State = "uncertain"
	StateLost       State = "lost"
	StateReleased   State = "released"
)

type Proof string

const (
	ProofContinuous Proof = "continuous"
	ProofUncertain  Proof = "uncertain"
	ProofLost       Proof = "lost"
)

var ErrApplicationProcessActive = errors.New("extension_application_process_active")
var ErrRecoveryServingLeaseActive = errors.New("recovery_serving_lease_active")
var ErrLeaseLost = errors.New("lease_lost")
var ErrApplicationProcessLeaseLost = fatalLeaseError("application_process_lease_lost")
var ErrRecoveryServingLeaseLost = fatalLeaseError("recovery_serving_lease_lost")
var ErrInvalidTransition = errors.New("application_process_lease_invalid_transition")

type fatalLeaseError string

func (err fatalLeaseError) Error() string {
	return string(err)
}

func (err fatalLeaseError) FatalReasonCode() string {
	return string(err)
}

type Session interface {
	Identity() string
	TryAcquire(context.Context) (bool, error)
	Prove(context.Context) Proof
	Release(context.Context) error
	Close()
}

type Backend interface {
	Open(context.Context) (Session, error)
}

type Event struct {
	Previous State
	State    State
	Reason   string
}

type Lease struct {
	mu            sync.RWMutex
	state         State
	session       Session
	sessionID     string
	lossDetection time.Duration
	events        chan Event
	monitorOnce   sync.Once
	monitorCancel context.CancelFunc
	monitorDone   chan struct{}
	closeOnce     sync.Once
}

func Acquire(ctx context.Context, backend Backend, timeout time.Duration, lossDetection time.Duration) (*Lease, error) {
	lease := &Lease{state: StateUnacquired, lossDetection: lossDetection, events: make(chan Event, 8)}
	if backend == nil || timeout <= 0 || lossDetection <= 0 {
		return nil, ErrInvalidTransition
	}
	lease.transition(StateUnacquired, StateAcquiring, "acquisition_started")
	acquireCtx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	session, err := backend.Open(acquireCtx)
	if err != nil {
		if errors.Is(acquireCtx.Err(), context.DeadlineExceeded) ||
			errors.Is(err, context.DeadlineExceeded) {
			return nil, ErrApplicationProcessActive
		}
		return nil, err
	}
	for {
		acquired, acquireErr := session.TryAcquire(acquireCtx)
		if acquireErr != nil {
			session.Close()
			if errors.Is(acquireCtx.Err(), context.DeadlineExceeded) {
				return nil, ErrApplicationProcessActive
			}
			return nil, acquireErr
		}
		if acquired {
			lease.mu.Lock()
			lease.session = session
			lease.sessionID = session.Identity()
			lease.mu.Unlock()
			lease.transition(StateAcquiring, StateHeld, "acquired")
			return lease, nil
		}
		select {
		case <-acquireCtx.Done():
			session.Close()
			return nil, ErrApplicationProcessActive
		case <-time.After(minDuration(50*time.Millisecond, timeout)):
		}
	}
}

func (l *Lease) State() State {
	if l == nil {
		return StateUnacquired
	}
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.state
}

func (l *Lease) SessionIdentity() string {
	if l == nil {
		return ""
	}
	l.mu.RLock()
	defer l.mu.RUnlock()
	return l.sessionID
}

func (l *Lease) Events() <-chan Event {
	if l == nil {
		return nil
	}
	return l.events
}

func (l *Lease) StartMonitor(ctx context.Context) {
	if l == nil {
		return
	}
	l.monitorOnce.Do(func() {
		monitorCtx, cancel := context.WithCancel(ctx)
		done := make(chan struct{})
		l.mu.Lock()
		l.monitorCancel = cancel
		l.monitorDone = done
		l.mu.Unlock()
		go func() {
			defer close(done)
			l.monitor(monitorCtx)
		}()
	})
}

func (l *Lease) monitor(ctx context.Context) {
	interval := minDuration(l.lossDetection/2, 500*time.Millisecond)
	if interval <= 0 {
		interval = time.Millisecond
	}
	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	var uncertainSince time.Time
	for {
		select {
		case <-ctx.Done():
			return
		case now := <-ticker.C:
			state := l.State()
			if state != StateHeld && state != StateUncertain {
				return
			}
			proofCtx, cancelProof := context.WithTimeout(ctx, l.lossDetection)
			proof := l.prove(proofCtx)
			cancelProof()
			switch proof {
			case ProofContinuous:
				if state == StateUncertain {
					l.transition(StateUncertain, StateHeld, "continuous_owner_proved")
					uncertainSince = time.Time{}
				}
			case ProofLost:
				if state == StateHeld {
					l.transition(StateHeld, StateUncertain, "ownership_proof_lost")
				}
				l.transition(StateUncertain, StateLost, "session_loss_confirmed")
				return
			case ProofUncertain:
				if state == StateHeld {
					l.transition(StateHeld, StateUncertain, "ownership_proof_uncertain")
					uncertainSince = now
				}
				if uncertainSince.IsZero() {
					uncertainSince = now
				}
				if !now.Before(uncertainSince.Add(l.lossDetection)) {
					l.transition(StateUncertain, StateLost, "loss_detection_deadline_expired")
					return
				}
			}
		}
	}
}

func (l *Lease) prove(ctx context.Context) Proof {
	l.mu.RLock()
	session := l.session
	sessionID := l.sessionID
	l.mu.RUnlock()
	if session == nil || session.Identity() != sessionID {
		return ProofLost
	}
	return session.Prove(ctx)
}

func (l *Lease) Release(ctx context.Context) error {
	if l == nil {
		return nil
	}
	var releaseErr error
	l.closeOnce.Do(func() {
		l.stopMonitor()
		if l.State() != StateHeld {
			releaseErr = ErrInvalidTransition
			l.closeSession()
			return
		}
		l.mu.RLock()
		session := l.session
		l.mu.RUnlock()
		releaseErr = session.Release(ctx)
		if releaseErr == nil {
			l.transition(StateHeld, StateReleased, "released")
		}
		l.closeSession()
	})
	return releaseErr
}

func (l *Lease) Close() {
	if l == nil {
		return
	}
	l.closeOnce.Do(func() {
		l.stopMonitor()
		l.closeSession()
	})
}

func (l *Lease) stopMonitor() {
	l.mu.RLock()
	cancel := l.monitorCancel
	done := l.monitorDone
	l.mu.RUnlock()
	if cancel != nil {
		cancel()
	}
	if done != nil {
		<-done
	}
}

func (l *Lease) closeSession() {
	l.mu.RLock()
	session := l.session
	l.mu.RUnlock()
	if session != nil {
		session.Close()
	}
}

func (l *Lease) transition(from State, to State, reason string) bool {
	l.mu.Lock()
	if l.state != from {
		l.mu.Unlock()
		return false
	}
	l.state = to
	l.mu.Unlock()
	select {
	case l.events <- Event{Previous: from, State: to, Reason: reason}:
	default:
	}
	return true
}

func minDuration(left time.Duration, right time.Duration) time.Duration {
	if left < right {
		return left
	}
	return right
}
