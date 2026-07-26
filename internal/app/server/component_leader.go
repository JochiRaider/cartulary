package server

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/processlease"
)

type componentLeader struct {
	backend       processlease.Backend
	retryInterval time.Duration
	proofTimeout  time.Duration
	run           func(context.Context) error
	lost          func()

	mu     sync.Mutex
	cancel context.CancelFunc
	done   chan struct{}
	held   atomic.Bool
}

func newComponentLeader(
	backend processlease.Backend,
	retryInterval time.Duration,
	proofTimeout time.Duration,
	run func(context.Context) error,
	lost func(),
) *componentLeader {
	return &componentLeader{
		backend:       backend,
		retryInterval: retryInterval,
		proofTimeout:  proofTimeout,
		run:           run,
		lost:          lost,
	}
}

func (leader *componentLeader) Start(parent context.Context) {
	if leader == nil {
		return
	}
	leader.mu.Lock()
	defer leader.mu.Unlock()
	if leader.cancel != nil {
		return
	}
	ctx, cancel := context.WithCancel(parent)
	leader.cancel = cancel
	leader.done = make(chan struct{})
	go leader.contend(ctx, leader.done)
}

func (leader *componentLeader) Close() {
	if leader == nil {
		return
	}
	leader.mu.Lock()
	cancel := leader.cancel
	done := leader.done
	leader.cancel = nil
	leader.done = nil
	leader.mu.Unlock()
	if cancel == nil {
		return
	}
	cancel()
	<-done
}

func (leader *componentLeader) IsLeader() bool {
	return leader != nil && leader.held.Load()
}

func (leader *componentLeader) contend(ctx context.Context, done chan<- struct{}) {
	defer close(done)
	for {
		if ctx.Err() != nil {
			return
		}
		session, err := leader.backend.Open(ctx)
		if err != nil {
			if !waitForComponentRetry(ctx, leader.retryInterval) {
				return
			}
			continue
		}
		acquired, err := session.TryAcquire(ctx)
		if err != nil || !acquired {
			session.Close()
			if !waitForComponentRetry(ctx, leader.retryInterval) {
				return
			}
			continue
		}
		leader.held.Store(true)
		if leader.serveAsLeader(ctx, session) {
			leader.held.Store(false)
			return
		}
		leader.held.Store(false)
	}
}

func (leader *componentLeader) serveAsLeader(ctx context.Context, session processlease.Session) bool {
	componentCtx, cancelComponent := context.WithCancel(ctx)
	componentDone := make(chan error, 1)
	go func() {
		componentDone <- leader.run(componentCtx)
	}()
	proofInterval := leader.proofTimeout / 2
	if proofInterval <= 0 {
		proofInterval = time.Second
	}
	ticker := time.NewTicker(proofInterval)
	defer ticker.Stop()
	defer cancelComponent()
	defer session.Close()
	for {
		select {
		case <-ctx.Done():
			cancelComponent()
			<-componentDone
			releaseCtx, cancelRelease := context.WithTimeout(context.Background(), leader.proofTimeout)
			_ = session.Release(releaseCtx)
			cancelRelease()
			return true
		case <-ticker.C:
			proofCtx, cancelProof := context.WithTimeout(ctx, leader.proofTimeout)
			proof := session.Prove(proofCtx)
			cancelProof()
			if proof == processlease.ProofContinuous {
				continue
			}
			cancelComponent()
			<-componentDone
			if leader.lost != nil {
				leader.lost()
			}
			return true
		case <-componentDone:
			if ctx.Err() == nil && leader.lost != nil {
				leader.lost()
			}
			return true
		}
	}
}

func waitForComponentRetry(ctx context.Context, interval time.Duration) bool {
	if interval <= 0 {
		interval = time.Second
	}
	timer := time.NewTimer(interval)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return false
	case <-timer.C:
		return true
	}
}
