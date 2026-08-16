package jobs

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"strings"
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
	ctx              context.Context
	cancel           context.CancelFunc
	wg               sync.WaitGroup
	mu               sync.Mutex
	closed           bool
	activated        bool
	manager          *Manager
	selection        *RuntimeSelection
	policy           RuntimePolicy
	handlers         map[string]HandlerFunc
	gate             DequeueGate
	notifications    chan uuid.UUID
	notified         map[uuid.UUID]struct{}
	inFlight         map[uuid.UUID]struct{}
	attemptSlots     chan int
	workerActive     map[string]int
	onComponentLoss  func()
	lossOnce         sync.Once
	shutdownCtx      context.Context
	shutdownFailures []shutdownReleaseFailure
	renewalTicks     func(time.Duration) (<-chan time.Time, func())
	renewExecution   func(context.Context, Execution) error
	releaseExecution func(context.Context, Execution) error
	beforeWait       func()
}

type shutdownReleaseFailure struct {
	stage   string
	jobKind string
	slot    int
	reason  string
}

func NewRunner(options RunnerOptions) (*Runner, error) {
	if options.Manager == nil || options.Catalog == nil || options.DequeueGate == nil ||
		options.Manager.catalog != options.Catalog || options.Manager.transactions.selection == nil ||
		options.Manager.transactions.selection.catalog != options.Catalog {
		return nil, ErrNotConfigured
	}
	if err := options.Policy.validate(); err != nil {
		return nil, err
	}
	ctx, cancel := context.WithCancel(context.Background())
	runner := &Runner{
		ctx:             ctx,
		cancel:          cancel,
		manager:         options.Manager,
		selection:       options.Manager.transactions.selection,
		policy:          cloneRuntimePolicy(options.Policy),
		handlers:        map[string]HandlerFunc{},
		gate:            options.DequeueGate,
		notifications:   make(chan uuid.UUID, options.Policy.RecoveryBatch),
		notified:        map[uuid.UUID]struct{}{},
		inFlight:        map[uuid.UUID]struct{}{},
		attemptSlots:    make(chan int, options.Policy.HandlerConcurrency),
		workerActive:    map[string]int{},
		onComponentLoss: options.OnComponentLoss,
		renewalTicks: func(interval time.Duration) (<-chan time.Time, func()) {
			ticker := time.NewTicker(interval)
			return ticker.C, ticker.Stop
		},
		renewExecution:   options.Manager.RenewExecution,
		releaseExecution: options.Manager.ReleaseExecution,
	}
	for slot := 1; slot <= options.Policy.HandlerConcurrency; slot++ {
		runner.attemptSlots <- slot
	}
	return runner, nil
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
	if !r.selection.hasHandlerName(name) {
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
	for _, handlerName := range r.selection.handlerNames() {
		if r.handlers[handlerName] == nil {
			r.mu.Unlock()
			return fmt.Errorf("%w: %s", ErrHandlerNotRegistered, handlerName)
		}
	}
	if len(r.handlers) != len(r.selection.handlerNames()) {
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
	previousSelection := ""
	for {
		jobKinds := r.unsaturatedJobKinds()
		selectionKey := strings.Join(jobKinds, "\x00")
		if len(jobKinds) == 0 || selectionKey == previousSelection {
			return nil
		}
		candidates, err := r.manager.recoverableCandidatesForSelection(ctx, r.policy.RecoveryBatch, r.selection, jobKinds)
		if err != nil {
			return err
		}
		if len(candidates) == 0 {
			return nil
		}
		for _, candidate := range candidates {
			r.schedule(candidate)
		}
		previousSelection = selectionKey
	}
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
			candidate, present, err := r.manager.runnerCandidateForSelection(r.ctx, jobID, r.selection)
			if err == nil && present {
				r.schedule(candidate)
			}
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

func (r *Runner) unsaturatedJobKinds() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.closed || !r.activated || len(r.attemptSlots) == 0 {
		return nil
	}
	result := make([]string, 0, len(r.selection.byKind))
	for _, jobKind := range r.selection.jobKinds() {
		workerKind, present := r.selection.workerKindForJob(jobKind)
		contract, valid := r.selection.workerContract(workerKind)
		if present && valid && r.workerActive[workerKind] < contract.MaxActiveAttemptsPerProcess {
			result = append(result, jobKind)
		}
	}
	return result
}

func (r *Runner) schedule(candidate runnerCandidate) {
	if candidate.JobID == uuid.Nil || candidate.JobKind == "" || candidate.HandlerName == "" {
		return
	}
	r.mu.Lock()
	if r.closed || !r.activated || !r.gate.AdmissionOpen() {
		r.mu.Unlock()
		return
	}
	if _, present := r.inFlight[candidate.JobID]; present {
		r.mu.Unlock()
		return
	}
	workerKind, assigned := r.selection.workerKindForJob(candidate.JobKind)
	workerContract, valid := r.selection.workerContract(workerKind)
	if !assigned || !valid || workerKind != candidate.HandlerName ||
		r.workerActive[workerKind] >= workerContract.MaxActiveAttemptsPerProcess {
		r.mu.Unlock()
		return
	}
	var attemptSlot int
	select {
	case attemptSlot = <-r.attemptSlots:
	default:
		r.mu.Unlock()
		return
	}
	r.inFlight[candidate.JobID] = struct{}{}
	r.workerActive[workerKind]++
	r.wg.Add(1)
	r.mu.Unlock()
	go r.execute(candidate, attemptSlot)
}

func (r *Runner) execute(candidate runnerCandidate, attemptSlot int) {
	defer func() {
		r.mu.Lock()
		delete(r.inFlight, candidate.JobID)
		r.workerActive[candidate.HandlerName]--
		r.attemptSlots <- attemptSlot
		r.mu.Unlock()
		r.wg.Done()
	}()
	execution, handlerName, jobKind, claimed, err := r.manager.claimForRunnerSelection(r.ctx, candidate.JobID, r.selection)
	if err != nil || !claimed {
		return
	}
	if handlerName != candidate.HandlerName || jobKind != candidate.JobKind {
		r.signalComponentLoss()
		r.withAttemptTimeout(func(ctx context.Context) {
			_ = r.releaseExecution(ctx, execution)
		})
		return
	}
	r.mu.Lock()
	handler := r.handlers[handlerName]
	r.mu.Unlock()
	if handler == nil {
		r.signalComponentLoss()
		r.withAttemptTimeout(func(ctx context.Context) {
			_ = r.releaseExecution(ctx, execution)
		})
		return
	}
	r.runExecution(execution, handler, jobKind, attemptSlot)
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

func (r *Runner) runExecution(execution Execution, handler HandlerFunc, jobKind string, attemptSlot int) {
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
	renewal, stopRenewal := r.renewalTicks(r.policy.LeaseRenewal)
	defer stopRenewal()
	for {
		if r.beforeWait != nil {
			r.beforeWait()
		}
		select {
		case <-r.ctx.Done():
			attemptOutcome = r.releaseForShutdown(execution, jobKind, attemptSlot, cancelHandler, result, false)
			return
		case <-renewal:
			if r.ctx.Err() != nil {
				attemptOutcome = r.releaseForShutdown(execution, jobKind, attemptSlot, cancelHandler, result, false)
				return
			}
			if err := r.renewExecution(r.ctx, execution); err != nil {
				if r.ctx.Err() != nil {
					attemptOutcome = r.releaseForShutdown(execution, jobKind, attemptSlot, cancelHandler, result, false)
					return
				}
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
			if r.ctx.Err() != nil {
				attemptOutcome = r.releaseForShutdown(execution, jobKind, attemptSlot, cancelHandler, result, true)
				return
			}
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
					_ = r.releaseExecution(ctx, execution)
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

func (r *Runner) releaseForShutdown(
	execution Execution,
	jobKind string,
	attemptSlot int,
	cancelHandler context.CancelFunc,
	handlerResult <-chan handlerResult,
	handlerDrained bool,
) attemptTelemetryOutcome {
	cancelHandler()
	releaseCtx, cancelRelease := r.shutdownAttemptContext()
	releaseErr := r.releaseExecution(releaseCtx, execution)
	releaseContextErr := releaseCtx.Err()
	cancelRelease()
	if releaseErr != nil && !errors.Is(releaseErr, ErrExecutionLost) {
		reason := "operation_failed"
		if errors.Is(releaseErr, context.DeadlineExceeded) || errors.Is(releaseContextErr, context.DeadlineExceeded) {
			reason = "deadline_exceeded"
		} else if errors.Is(releaseErr, context.Canceled) || errors.Is(releaseContextErr, context.Canceled) {
			reason = "caller_canceled"
		}
		r.recordShutdownFailure(shutdownReleaseFailure{
			stage:   "release",
			jobKind: jobKind,
			slot:    attemptSlot,
			reason:  reason,
		})
	}
	if !handlerDrained {
		<-handlerResult
	}
	return r.classifyAttempt(execution.JobID(), "canceled")
}

func (r *Runner) recordShutdownFailure(failure shutdownReleaseFailure) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.shutdownFailures = append(r.shutdownFailures, failure)
}

func (r *Runner) classifyAttempt(jobID uuid.UUID, fallback string) attemptTelemetryOutcome {
	outcome := attemptTelemetryOutcome{result: fallback, failed: fallback == "failed" || fallback == "conflict"}
	timeout := r.policy.AttemptOperationTimeout
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
	timeout := r.policy.AttemptOperationTimeout
	if timeout <= 0 {
		timeout = time.Second
	}
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	operation(ctx)
}

func (r *Runner) shutdownAttemptContext() (context.Context, context.CancelFunc) {
	r.mu.Lock()
	base := r.shutdownCtx
	r.mu.Unlock()
	if base == nil {
		base = context.Background()
	}
	return context.WithTimeout(base, r.policy.AttemptOperationTimeout)
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
		r.shutdownCtx = ctx
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
		r.mu.Lock()
		defer r.mu.Unlock()
		return aggregateShutdownFailures(r.shutdownFailures)
	case <-ctx.Done():
		return ctx.Err()
	}
}

func aggregateShutdownFailures(failures []shutdownReleaseFailure) error {
	if len(failures) == 0 {
		return nil
	}
	ordered := append([]shutdownReleaseFailure(nil), failures...)
	sort.Slice(ordered, func(i int, j int) bool {
		if ordered[i].slot != ordered[j].slot {
			return ordered[i].slot < ordered[j].slot
		}
		if ordered[i].jobKind != ordered[j].jobKind {
			return ordered[i].jobKind < ordered[j].jobKind
		}
		return ordered[i].reason < ordered[j].reason
	})
	parts := make([]string, 0, len(ordered))
	for _, failure := range ordered {
		parts = append(parts, fmt.Sprintf(
			"stage=%s job_kind=%s attempt_slot=%d reason=%s",
			failure.stage,
			failure.jobKind,
			failure.slot,
			failure.reason,
		))
	}
	return fmt.Errorf("jobs: runner shutdown unsuccessful: %s", strings.Join(parts, "; "))
}
