package server

import (
	"context"
	"errors"
	"sync/atomic"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/extensions"
	extensiondeadline "github.com/JochiRaider/cartulary/internal/modules/extensions/deadline"
)

type publicationTiming struct {
	seconds int64
	grace   time.Duration
	now     func() int64
}

// One guarded operation spans component preparation, the listener-ready
// handshake and activation. Its timer ends at the serving transition; its
// context remains the prepared components' lifetime until shutdown.
type publicationOperation struct {
	controller  *publicationController
	ctx         context.Context
	cancel      context.CancelFunc
	timer       *time.Timer
	policy      extensiondeadline.InvocationPolicy
	coordinator *extensions.Coordinator
	seconds     int64
	prepared    chan struct{}
	activation  chan struct{}
	done        chan struct{}
	failures    chan error
	requested   atomic.Bool
	err         error // published by closing done
}

func (r *Runtime) beginPublication(parent context.Context, coordinator *extensions.Coordinator, timing publicationTiming, prepare func(context.Context) error) error {
	if r == nil || r.lifecycle == nil || r.publication == nil || coordinator == nil || timing.seconds <= 0 || timing.grace < 0 || prepare == nil || r.publicationOperation != nil {
		return errors.New("invalid publication operation binding")
	}
	if timing.now == nil {
		origin := time.Now()
		timing.now = func() int64 { return time.Since(origin).Nanoseconds() + 1 }
	}
	bound := extensiondeadline.New(timing.now(), timing.seconds, nil)
	ctx, cancel := context.WithCancel(parent)
	operation := &publicationOperation{
		controller: r.publication,
		ctx:        extensiondeadline.WithCancellationTracking(ctx), cancel: cancel,
		coordinator: coordinator, seconds: timing.seconds,
		prepared: make(chan struct{}), activation: make(chan struct{}), done: make(chan struct{}), failures: make(chan error, 1),
		policy: extensiondeadline.InvocationPolicy{Grace: timing.grace, Exhausted: r.fatalIntegrity, Deadline: &bound, MonotonicNowNS: timing.now},
	}
	r.publicationOperation = operation
	r.publication.mu.Lock()
	r.publication.operation = operation
	operation.timer = time.AfterFunc(time.Duration(bound.MonotonicNS-timing.now()), func() {
		r.publication.mu.Lock()
		defer r.publication.mu.Unlock()
		if r.publication.state != publicationServing {
			operation.cancel()
		}
	})
	r.publication.mu.Unlock()
	go func() {
		select {
		case <-r.lifecycle.FatalDone():
			operation.cancel()
		case <-operation.done:
		}
	}()
	go operation.run(r, prepare)
	select {
	case <-operation.prepared:
		return nil
	case <-operation.done:
		return operation.err
	}
}

func (o *publicationOperation) boundary() error { return extensiondeadline.Sample(o.ctx, o.policy) }

func (o *publicationOperation) run(r *Runtime, prepare func(context.Context) error) {
	completion, lifetimeErr := extensiondeadline.Complete(o.ctx, o.policy, func(ctx context.Context) (bool, error) {
		if err := o.boundary(); err != nil {
			return false, err
		}
		if err := prepare(ctx); err != nil {
			return false, err
		}
		if err := o.boundary(); err != nil {
			return false, err
		}
		close(o.prepared)
		select {
		case <-o.activation:
		case <-ctx.Done():
			return false, o.boundary()
		}
		if err := o.boundary(); err != nil {
			return false, err
		}
		if err := r.activatePreparedPublication(ctx); err != nil {
			return false, err
		}
		return true, nil
	})
	o.timer.Stop()
	if reason := r.lifecycle.FatalReason(); reason != "" {
		o.err = publicationStartupFatal{reason: reason}
	} else if lifetimeErr != nil {
		o.err = publicationStartupFatal{reason: "operation_cancellation_grace_exhausted"}
	} else if !completion.Value {
		o.err = o.coordinator.PublicationFailure(o.seconds, errors.Join(completion.Err, completion.Interruption))
	}
	if o.err != nil {
		o.cancel()
		r.publication.abortStartup()
		o.failures <- o.err
	}
	close(o.done)
}

func (o *publicationOperation) activate() error {
	if !o.requested.CompareAndSwap(false, true) {
		o.controller.mu.Lock()
		cause := o.controller.failLocked("duplicate_activation")
		o.controller.mu.Unlock()
		o.cancel()
		<-o.done
		if reason := o.controller.lifecycle.FatalReason(); reason != "" {
			return publicationStartupFatal{reason: reason}
		}
		return o.coordinator.PublicationFailure(o.seconds, cause)
	}
	close(o.activation)
	<-o.done
	return o.err
}

func (o *publicationOperation) stop() {
	o.cancel()
	<-o.done
}

// A transport still active after publication fails shares the original
// interruption and grace. Joining never grants a fresh cancellation interval.
func (o *publicationOperation) joinTransport(done <-chan error) error {
	_, err := extensiondeadline.Complete(o.ctx, o.policy, func(context.Context) (struct{}, error) {
		<-done
		return struct{}{}, nil
	})
	return err
}
