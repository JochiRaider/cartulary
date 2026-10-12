package harnessreview

import (
	"context"
	"errors"
	"sync"
	"time"
)

// Driver supplies physical operations and Go-owned dependency cleanup. It never
// decides public states or publishes retained operation/terminal outcomes.
type Driver interface {
	Prepare(context.Context) (*int, error)
	Validate(context.Context, string, jsonRequest) error
	Execute(context.Context, string, jsonRequest, *string, int) (Operation, error)
	Interrupt()
	Close(context.Context) []error
}
type jsonRequest map[string]any
type Operation struct {
	Epoch    *int
	Counts   Counts
	Bundle   *string
	Private  []PrivateReference
	Failures []Failure
	Exit     int
	Stop     bool
}
type Publisher interface {
	Locator(Locator) error
	Operation(Receipt) (*Reference, error)
	Terminal(Locator, Receipt, bool) (*Reference, error)
}
type Session struct {
	mu                                                    sync.Mutex
	id, runID, mode, state                                string
	started                                               time.Time
	epoch                                                 *int
	operation                                             int
	counts                                                Counts
	failures                                              []Failure
	exit                                                  int
	receipt                                               *Reference
	driver                                                Driver
	publisher                                             Publisher
	now                                                   func() time.Time
	lockTimeout, operationTimeout, drainTimeout, lifetime time.Duration
	ctx                                                   context.Context
	cancel                                                context.CancelFunc
	prepareDone                                           chan struct{}
	done                                                  chan struct{}
	slot                                                  chan struct{}
	stopOnce                                              sync.Once
}

func NewSession(id, runID, mode string, driver Driver, publisher Publisher) *Session {
	ctx, cancel := context.WithCancel(context.Background())
	return &Session{id: id, runID: runID, mode: mode, state: "preparing", started: time.Now(), driver: driver, publisher: publisher, now: time.Now, lockTimeout: time.Duration(LimitLock) * time.Millisecond, operationTimeout: time.Duration(LimitOperation) * time.Millisecond, drainTimeout: time.Duration(LimitDrain) * time.Millisecond, lifetime: time.Duration(LimitLifetime) * time.Millisecond, ctx: ctx, cancel: cancel, prepareDone: make(chan struct{}), done: make(chan struct{}), slot: make(chan struct{}, 1), failures: []Failure{}}
}
func (s *Session) locatorLocked() Locator {
	return Locator{Schema: "cartulary.ui_review_session.v1", Session: s.id, RunID: s.runID, Mode: s.mode, State: s.state, Created: timestamp(s.started), Updated: timestamp(s.now()), Receipt: s.receipt}
}
func (s *Session) publishLocked() error { return s.publisher.Locator(s.locatorLocked()) }
func (s *Session) Prepare() error {
	s.mu.Lock()
	err := s.publishLocked()
	s.mu.Unlock()
	var epoch *int
	if err == nil {
		epoch, err = s.driver.Prepare(s.ctx)
	}
	s.mu.Lock()
	if err == nil && s.ctx.Err() == nil {
		s.epoch = epoch
		s.state = "ready"
		err = s.publishLocked()
	}
	if err != nil && !errors.Is(err, context.Canceled) {
		e := normalized(err, "preparation_failed")
		if e.Record.Code != "interrupted" || s.ctx.Err() == nil {
			s.addPrimaryLocked(e)
		}
	}
	s.mu.Unlock()
	close(s.prepareDone)
	if err != nil {
		s.Stop(normalized(err, "preparation_failed"))
		return err
	}
	if s.ctx.Err() != nil {
		return s.ctx.Err()
	}
	go func() {
		timer := time.NewTimer(s.lifetime)
		defer timer.Stop()
		select {
		case <-timer.C:
			s.Stop(nil)
		case <-s.done:
		}
	}()
	return nil
}
func (s *Session) valueLocked(command string) Result {
	v := result(command)
	v.Session = pointer(s.id)
	v.State = pointer(s.state)
	v.Epoch = s.epoch
	if s.state == "closed" || s.state == "failed" {
		v.Epoch = nil
		v.Receipt = s.receipt
		v.Failures = append([]Failure{}, s.failures...)
		v.Exit = s.exit
		if len(v.Failures) > 0 {
			v.Status = "error"
		}
	}
	return v
}
func (s *Session) Value(command string) Result {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.valueLocked(command)
}
func (s *Session) fail(command string, err error) Result {
	v := s.Value(command)
	e := normalized(err, "unsafe_artifact")
	v.Status = "error"
	v.Exit = e.Exit
	v.Failures = []Failure{e.Record}
	return v
}
func (s *Session) Handle(ctx context.Context, command string, request jsonRequest, bundle *string) Result {
	if command == "ui-review-status" {
		return s.Value(command)
	}
	if command == "ui-review-stop" {
		s.Stop(nil)
		select {
		case <-s.done:
			return s.Value(command)
		case <-ctx.Done():
			return s.fail(command, failure("cleanup_failed"))
		}
	}
	timer := time.NewTimer(s.lockTimeout)
	defer timer.Stop()
	select {
	case s.slot <- struct{}{}:
	case <-timer.C:
		return s.fail(command, failure("capacity_exceeded"))
	case <-ctx.Done():
		return s.fail(command, failure("interrupted"))
	case <-s.ctx.Done():
		return s.fail(command, failure("session_mismatch"))
	}
	defer func() { <-s.slot }()
	s.mu.Lock()
	ready := s.state == "ready"
	s.mu.Unlock()
	if !ready {
		return s.fail(command, failure("session_mismatch"))
	}
	validation, done := context.WithTimeout(s.ctx, s.operationTimeout)
	err := s.driver.Validate(validation, command, request)
	done()
	if err != nil {
		return s.fail(command, err)
	}
	s.mu.Lock()
	if s.ctx.Err() != nil {
		s.mu.Unlock()
		return s.fail(command, failure("session_mismatch"))
	}
	s.operation++
	id := s.operation
	s.state = "busy"
	err = s.publishLocked()
	s.mu.Unlock()
	if err != nil {
		s.Stop(failure("unsafe_artifact"))
		return s.fail(command, failure("unsafe_artifact"))
	}
	watchdog := time.AfterFunc(s.operationTimeout, func() { s.Stop(failure("operation_expired")) })
	defer watchdog.Stop()
	started := s.now()
	opCtx, cancel := context.WithTimeout(s.ctx, s.operationTimeout)
	defer cancel()
	outcome, err := s.driver.Execute(opCtx, command, request, bundle, id)
	if errors.Is(opCtx.Err(), context.DeadlineExceeded) {
		err = failure("operation_expired")
		s.Stop(err)
	}
	if ctx.Err() != nil {
		s.Stop(failure("interrupted"))
	}
	if errors.Is(err, context.Canceled) {
		s.mu.Lock()
		if len(s.failures) > 0 {
			err = &ReviewError{Record: s.failures[0], Exit: s.exit}
		} else {
			err = failure("interrupted")
		}
		s.mu.Unlock()
	}
	if err != nil {
		e := normalized(err, "unsafe_artifact")
		outcome.Failures = []Failure{e.Record}
		outcome.Exit = e.Exit
	}
	s.mu.Lock()
	if s.ctx.Err() != nil && len(outcome.Failures) == 0 {
		e := failure("interrupted")
		outcome.Failures = []Failure{e.Record}
		outcome.Exit = e.Exit
	}
	s.epoch = outcome.Epoch
	value := s.valueLocked(command)
	value.Operation = &id
	value.Bundle = outcome.Bundle
	value.Private = outcome.Private
	if value.Private == nil {
		value.Private = []PrivateReference{}
	}
	value.Failures = outcome.Failures
	if value.Failures == nil {
		value.Failures = []Failure{}
	}
	value.Exit = outcome.Exit
	if len(value.Failures) > 0 {
		value.Status = "error"
		value.Private = []PrivateReference{}
	} else {
		s.counts.Add(outcome.Counts)
	}
	if s.state == "stopping" {
		value.Private = []PrivateReference{}
	} else {
		value.State = pointer("ready")
	}
	receipt := Receipt{Schema: "cartulary.ui_review_receipt.v2", Command: commandID(command), Session: s.id, Operation: &id, Mode: s.mode, State: "ready", Status: value.Status, Exit: value.Exit, Started: timestamp(started), Finished: timestamp(s.now()), Duration: s.now().Sub(started).Milliseconds(), Failures: value.Failures, Counts: outcome.Counts, Bundle: value.Bundle, Cleanup: "not_terminal"}
	// The operation producer settles its one receipt before stop can finalize.
	value.Receipt, err = s.publisher.Operation(receipt)
	if s.state == "busy" {
		s.state = "ready"
		err = errors.Join(err, s.publishLocked())
	}
	s.mu.Unlock()
	if err != nil {
		s.Stop(failure("unsafe_artifact"))
		value = s.fail(command, failure("unsafe_artifact"))
		value.Operation = &id
	}
	if outcome.Stop {
		cause := error(failure("cleanup_failed"))
		if len(outcome.Failures) > 0 {
			cause = &ReviewError{Record: outcome.Failures[0], Exit: outcome.Exit, Secondary: outcome.Failures[1:]}
		}
		s.Stop(cause)
	}
	return value
}
func (s *Session) Stop(cause error) {
	s.stopOnce.Do(func() {
		s.mu.Lock()
		s.state = "stopping"
		if cause != nil {
			e := normalized(cause, "unsafe_artifact")
			s.addPrimaryLocked(e)
		}
		publishErr := s.publishLocked()
		s.mu.Unlock()
		s.cancel()
		s.driver.Interrupt()
		go s.finish(publishErr)
	})
}
func (s *Session) finish(publishErr error) {
	defer close(s.done)
	failures := []error{}
	if publishErr != nil {
		failures = append(failures, publishErr)
	}
	// Preparation owns bounded cancellation and any late-acquisition cleanup.
	// Do not start a competing aggregate clock while that owner relinquishes.
	<-s.prepareDone
	// Close first drains physical browser/work children so an active operation can
	// settle. Its Go driver attempts independent cleanup even after a failed step.
	failures = append(failures, s.driver.Close(context.Background())...)
	drain, cancel := context.WithTimeout(context.Background(), s.drainTimeout)
	defer cancel()
	select {
	case s.slot <- struct{}{}:
		defer func() { <-s.slot }()
	case <-drain.Done():
		s.failUndrained()
		return
	}
	s.mu.Lock()
	defer s.mu.Unlock()
	for _, err := range failures {
		e := normalized(err, "cleanup_failed")
		s.failures = appendFailure(s.failures, e.Record)
		if s.exit == 0 {
			s.exit = e.Exit
		}
	}
	s.state = "closed"
	status := "ok"
	if len(s.failures) > 0 {
		s.state = "failed"
		status = "error"
	}
	receipt := Receipt{Schema: "cartulary.ui_review_receipt.v2", Command: commandID("ui-review"), Session: s.id, Mode: s.mode, State: s.state, Status: status, Exit: s.exit, Started: timestamp(s.started), Finished: timestamp(s.now()), Duration: s.now().Sub(s.started).Milliseconds(), Failures: s.failures, Counts: s.counts, Cleanup: "complete"}
	if len(failures) > 0 {
		receipt.Cleanup = "failed"
	}
	ref, err := s.publisher.Terminal(s.locatorLocked(), receipt, len(failures) > 0)
	if err != nil {
		s.state = "failed"
		cause := normalized(err, "unsafe_artifact")
		s.failures = appendFailure(s.failures, cause.Record)
		if s.exit == 0 {
			s.exit = cause.Exit
		}
		s.receipt = nil
		_ = s.publishLocked()
	} else {
		s.receipt = ref
	}
}
func (s *Session) Done() <-chan struct{} { return s.done }

func (s *Session) failUndrained() {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state = "failed"
	s.failures = appendFailure(s.failures, failure("cleanup_failed").Record)
	if s.exit == 0 {
		s.exit = 12
	}
	_ = s.publishLocked()
}

func (s *Session) Ready() bool { s.mu.Lock(); defer s.mu.Unlock(); return s.state == "ready" }

func (s *Session) addPrimaryLocked(e *ReviewError) {
	if len(s.failures) == 0 {
		s.failures = appendFailure(s.failures, e.Record)
		s.exit = e.Exit
	} else if s.failures[0].Code == "interrupted" && e.Record.Code != "interrupted" {
		s.failures = append([]Failure{e.Record}, s.failures...)
		s.exit = e.Exit
	} else {
		s.failures = appendFailure(s.failures, e.Record)
	}
	for _, secondary := range e.Secondary {
		s.failures = appendFailure(s.failures, secondary)
	}
}
