package jobs

import (
	"context"
	"errors"
	"fmt"
	"sync"
	"time"

	"github.com/google/uuid"
)

var ErrRunnerClosed = errors.New("jobs: runner closed")
var ErrRunnerActivated = errors.New("jobs: runner already activated")
var ErrHandlerNotRegistered = errors.New("jobs: handler not registered")
var ErrHandlerAlreadyRegistered = errors.New("jobs: handler already registered")
var ErrHandlerNotInCatalog = errors.New("jobs: handler not in catalog")

type HandlerFunc func(context.Context, Execution) error

type DequeueGate interface {
	AdmissionOpen() bool
}

type RunnerOptions struct {
	Manager         *Manager
	Catalog         *Catalog
	Policy          RuntimePolicy
	DequeueGate     DequeueGate
	OnComponentLoss func()
}

type Runner struct {
	ctx             context.Context
	cancel          context.CancelFunc
	wg              sync.WaitGroup
	mu              sync.Mutex
	closed          bool
	activated       bool
	manager         *Manager
	catalog         *Catalog
	policy          RuntimePolicy
	handlers        map[string]HandlerFunc
	gate            DequeueGate
	notifications   chan uuid.UUID
	notified        map[uuid.UUID]struct{}
	inFlight        map[uuid.UUID]struct{}
	attemptSlots    chan struct{}
	onComponentLoss func()
	lossOnce        sync.Once
}

func NewRunner(options RunnerOptions) (*Runner, error) {
	if options.Manager == nil || options.Catalog == nil || options.DequeueGate == nil ||
		options.Manager.catalog != options.Catalog {
		return nil, ErrNotConfigured
	}
	if err := options.Policy.validate(); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	return &Runner{
		ctx:             ctx,
		cancel:          cancel,
		manager:         options.Manager,
		catalog:         options.Catalog,
		policy:          cloneRuntimePolicy(options.Policy),
		handlers:        map[string]HandlerFunc{},
		gate:            options.DequeueGate,
		notifications:   make(chan uuid.UUID, options.Policy.RecoveryBatch),
		notified:        map[uuid.UUID]struct{}{},
		inFlight:        map[uuid.UUID]struct{}{},
		attemptSlots:    make(chan struct{}, options.Policy.HandlerConcurrency),
		onComponentLoss: options.OnComponentLoss,
	}, nil
}

func (r *Runner) RegisterHandler(name string, handler HandlerFunc) error {
	if r == nil || name == "" || handler == nil {
		return ErrInvalidJobDefinition
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return ErrRunnerClosed
	}
	if r.activated {
		return ErrRunnerActivated
	}
	if !r.catalog.hasHandlerName(name) {
		return fmt.Errorf("%w: %s", ErrHandlerNotInCatalog, name)
	}
	if _, exists := r.handlers[name]; exists {
		return ErrHandlerAlreadyRegistered
	}
	r.handlers[name] = handler
	return nil
}

// Notify submits a best-effort, non-blocking acceleration hint. The bounded
// buffer is deduplicated, and periodic durable scans remain authoritative.
func (r *Runner) Notify(jobID uuid.UUID) {
	if r == nil || jobID == uuid.Nil {
		return
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed {
		return
	}
	if _, present := r.notified[jobID]; present {
		return
	}
	if _, present := r.inFlight[jobID]; present {
		return
	}
	select {
	case r.notifications <- jobID:
		r.notified[jobID] = struct{}{}
	default:
	}
}

func (r *Runner) Activate(ctx context.Context) error {
	if r == nil {
		return ErrNotConfigured
	}
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		return ErrRunnerClosed
	}
	if r.activated {
		r.mu.Unlock()
		return ErrRunnerActivated
	}
	for _, handlerName := range r.catalog.handlerNames() {
		if r.handlers[handlerName] == nil {
			r.mu.Unlock()
			return fmt.Errorf("%w: %s", ErrHandlerNotRegistered, handlerName)
		}
	}
	if len(r.handlers) != len(r.catalog.handlerNames()) {
		r.mu.Unlock()
		return ErrHandlerNotInCatalog
	}
	r.activated = true
	// This reservation prevents Close from observing a zero wait-group while
	// activation performs the synchronous initial scan.
	r.wg.Add(1)
	r.mu.Unlock()

	if err := r.scan(ctx); err != nil {
		r.wg.Done()
		return err
	}
	go r.supervise()
	return nil
}

func (r *Runner) scan(ctx context.Context) error {
	jobIDs, err := r.manager.RecoverableJobs(ctx, r.policy.RecoveryBatch)
	if err != nil {
		return err
	}
	for _, jobID := range jobIDs {
		r.schedule(jobID)
	}
	return nil
}

func (r *Runner) supervise() {
	defer func() {
		if recover() != nil || r.ctx.Err() == nil {
			r.signalComponentLoss()
		}
		r.wg.Done()
	}()
	ticker := time.NewTicker(r.policy.RecoveryScan)
	defer ticker.Stop()
	expiryTicker := time.NewTicker(r.policy.ExpirySweep)
	defer expiryTicker.Stop()
	for {
		select {
		case <-r.ctx.Done():
			return
		case jobID := <-r.notifications:
			r.mu.Lock()
			delete(r.notified, jobID)
			r.mu.Unlock()
			r.schedule(jobID)
		case <-ticker.C:
			// A transient scan failure leaves the supervisor alive. The next
			// fixed-cadence scan retries from durable state.
			_ = r.scan(r.ctx)
		case <-expiryTicker.C:
			// Expiry is private maintenance: it publishes no lifecycle event and
			// retries from durable candidates on the next full interval.
			_, _ = r.manager.compactExpiredJobs(r.ctx, r.policy.ExpiryBatch)
		}
	}
}

func (r *Runner) schedule(jobID uuid.UUID) {
	if jobID == uuid.Nil {
		return
	}
	r.mu.Lock()
	if r.closed || !r.activated || !r.gate.AdmissionOpen() {
		r.mu.Unlock()
		return
	}
	if _, present := r.inFlight[jobID]; present {
		r.mu.Unlock()
		return
	}
	select {
	case r.attemptSlots <- struct{}{}:
	default:
		r.mu.Unlock()
		return
	}
	r.inFlight[jobID] = struct{}{}
	r.wg.Add(1)
	r.mu.Unlock()
	go r.execute(jobID)
}

func (r *Runner) execute(jobID uuid.UUID) {
	defer func() {
		<-r.attemptSlots
		r.mu.Lock()
		delete(r.inFlight, jobID)
		r.mu.Unlock()
		r.wg.Done()
	}()
	execution, handlerName, jobKind, claimed, err := r.manager.claimForRunner(r.ctx, jobID)
	if err != nil || !claimed {
		return
	}
	r.mu.Lock()
	handler := r.handlers[handlerName]
	r.mu.Unlock()
	if handler == nil {
		r.signalComponentLoss()
		_ = r.manager.ReleaseExecution(context.Background(), execution)
		return
	}
	r.runExecution(execution, handler, jobKind)
}

type handlerResult struct {
	err      error
	panicked bool
}

type attemptTelemetryOutcome struct {
	result         string
	terminalStatus string
	failed         bool
}

var errAttemptTelemetryFailure = errors.New("jobs: closed attempt failure")

func (outcome attemptTelemetryOutcome) telemetryError() error {
	if outcome.failed {
		return errAttemptTelemetryFailure
	}
	return nil
}

func (r *Runner) runExecution(execution Execution, handler HandlerFunc, jobKind string) {
	attemptCtx, span := r.manager.startJobSpan(r.ctx, "cartulary.jobs.run", jobKind, "run")
	attemptOutcome := attemptTelemetryOutcome{result: "failed", failed: true}
	defer func() {
		r.manager.recordAttempt(attemptCtx, jobKind, attemptOutcome.result)
		r.manager.finishJobSpan(
			span,
			"run",
			jobKind,
			attemptOutcome.terminalStatus,
			attemptOutcome.result,
			attemptOutcome.telemetryError(),
		)
	}()
	handlerCtx, cancelHandler := context.WithCancel(attemptCtx)
	defer cancelHandler()
	result := make(chan handlerResult, 1)
	go func() {
		outcome := handlerResult{}
		defer func() {
			if recover() != nil {
				outcome.panicked = true
			}
			result <- outcome
		}()
		outcome.err = handler(handlerCtx, execution)
	}()
	renewal := time.NewTicker(r.policy.LeaseRenewal)
	defer renewal.Stop()
	for {
		select {
		case <-r.ctx.Done():
			cancelHandler()
			r.withAttemptTimeout(func(ctx context.Context) {
				_ = r.manager.ReleaseExecution(ctx, execution)
			})
			attemptOutcome = r.classifyAttempt(execution.JobID(), "canceled")
			return
		case <-renewal.C:
			if err := r.manager.RenewExecution(r.ctx, execution); err != nil {
				result := "failed"
				if errors.Is(err, ErrExecutionLost) {
					result = "conflict"
				}
				r.manager.recordLeaseRenewalFailure(attemptCtx, jobKind, result)
				cancelHandler()
				if !errors.Is(err, ErrExecutionLost) {
					r.withAttemptTimeout(func(ctx context.Context) {
						_ = r.manager.RecordExecutionFailure(ctx, execution, false)
					})
				}
				attemptOutcome = r.classifyAttempt(execution.JobID(), result)
				return
			}
		case handlerOutcome := <-result:
			cancelHandler()
			fallback := "failed"
			switch {
			case handlerOutcome.panicked:
				r.withAttemptTimeout(func(ctx context.Context) {
					_ = r.manager.RecordExecutionFailure(ctx, execution, false)
				})
			case handlerOutcome.err == nil:
				r.withAttemptTimeout(func(ctx context.Context) {
					_ = r.manager.RecordExecutionFailure(ctx, execution, true)
				})
			case r.ctx.Err() != nil && errors.Is(handlerOutcome.err, context.Canceled):
				fallback = "canceled"
				r.withAttemptTimeout(func(ctx context.Context) {
					_ = r.manager.ReleaseExecution(ctx, execution)
				})
			default:
				r.withAttemptTimeout(func(ctx context.Context) {
					_ = r.manager.RecordExecutionFailure(ctx, execution, false)
				})
			}
			attemptOutcome = r.classifyAttempt(execution.JobID(), fallback)
			return
		}
	}
}

func (r *Runner) classifyAttempt(jobID uuid.UUID, fallback string) attemptTelemetryOutcome {
	outcome := attemptTelemetryOutcome{result: fallback, failed: fallback == "failed" || fallback == "conflict"}
	timeout := r.policy.LeaseRenewal
	if timeout <= 0 {
		timeout = time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	resource, err := r.manager.Get(ctx, jobID)
	if err != nil {
		return outcome
	}
	switch resource.Status {
	case StatusSucceeded, StatusFailed, StatusCanceled:
		outcome.result = resultForTerminalStatus(resource.Status)
		outcome.terminalStatus = resource.Status
		outcome.failed = resource.Status == StatusFailed
	}
	return outcome
}

func (r *Runner) withAttemptTimeout(operation func(context.Context)) {
	timeout := r.policy.LeaseRenewal
	if timeout <= 0 {
		timeout = time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	operation(ctx)
}

func (r *Runner) signalComponentLoss() {
	if r == nil || r.onComponentLoss == nil {
		return
	}
	r.lossOnce.Do(func() {
		defer func() { _ = recover() }()
		r.onComponentLoss()
	})
}

func (r *Runner) Close(ctx context.Context) error {
	if r == nil {
		return nil
	}
	r.mu.Lock()
	if !r.closed {
		r.closed = true
		r.cancel()
	}
	r.mu.Unlock()

	done := make(chan struct{})
	go func() {
		r.wg.Wait()
		close(done)
	}()
	select {
	case <-done:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
