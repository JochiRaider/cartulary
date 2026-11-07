package deadline

import (
	"context"
	"errors"
	"sync"
	"time"
)

var errInvocationPanic = errors.New("extension invocation panicked")

type graceExhaustedError struct{}

func (graceExhaustedError) Error() string { return "extension cancellation grace exhausted" }
func (graceExhaustedError) FatalReasonCode() string {
	return "operation_cancellation_grace_exhausted"
}

var ErrGraceExhausted error = graceExhaustedError{}

type interruptionError struct{ cause error }

func (e *interruptionError) Error() string { return e.cause.Error() }
func (e *interruptionError) Unwrap() error { return e.cause }

// Interrupted distinguishes a sampled operation boundary from an error merely
// returned by an owner callback (which may itself use a private child context).
func Interrupted(err error) bool {
	var interruption *interruptionError
	return errors.As(err, &interruption)
}

// InvocationPolicy is supplied by the owner of the operation. Exhausted must
// signal the imported process lifecycle immediately; it must not drain workers
// or wait for the canceled callback.
type InvocationPolicy struct {
	Grace          time.Duration
	Exhausted      func(error)
	Deadline       *Deadline
	MonotonicNowNS func() int64
}

// WithCancellationTracking shares the first interruption and its grace deadline
// across an operation's callbacks and transaction completion. Child-context
// cleanup does not record cancellation: only an executing boundary can do so.
func WithCancellationTracking(ctx context.Context) context.Context {
	return context.WithValue(ctx, cancellationKey{}, &cancellationObservation{})
}

// ReuseCancellationTracking preserves an enclosing operation's first
// interruption. A new operation starts tracking only when none was inherited.
func ReuseCancellationTracking(ctx context.Context) context.Context {
	if _, ok := ctx.Value(cancellationKey{}).(*cancellationObservation); ok {
		return ctx
	}
	return WithCancellationTracking(ctx)
}

// GraceExhausted reads the inherited operation's terminal lifetime selection.
// It neither starts observation nor grants a new cancellation grace.
func GraceExhausted(ctx context.Context) bool {
	observed, ok := ctx.Value(cancellationKey{}).(*cancellationObservation)
	return ok && observed.exhausted()
}

type cancellationKey struct{}
type cancellationObservation struct {
	mu           sync.Mutex
	cause        error
	at           time.Time
	fatal        bool
	publications map[*context.CancelCauseFunc]struct{}
}

func observation(ctx context.Context) *cancellationObservation {
	if value, ok := ctx.Value(cancellationKey{}).(*cancellationObservation); ok {
		return value
	}
	return &cancellationObservation{}
}
func (o *cancellationObservation) sample(ctx context.Context, policy InvocationPolicy) (error, time.Time) {
	o.mu.Lock()
	defer o.mu.Unlock()
	if o.cause == nil {
		if cause := invocationContextError(ctx, policy); cause != nil {
			o.cause = cause
			o.at = time.Now()
			if limit, ok := ctx.Deadline(); ok && errors.Is(cause, context.DeadlineExceeded) && limit.Before(o.at) {
				o.at = limit
			}
		}
	}
	return o.cause, o.at
}
func (o *cancellationObservation) exhaust(policy InvocationPolicy) {
	o.mu.Lock()
	first := !o.fatal
	o.fatal = true
	// Revoke publication before exposing the fatal selection or calling the
	// asynchronous process sink. Physical cleanup still remains joined.
	for cancel := range o.publications {
		(*cancel)(ErrGraceExhausted)
	}
	o.mu.Unlock()
	if first {
		policy.Exhausted(ErrGraceExhausted)
	}
}
func (o *cancellationObservation) exhausted() bool {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.fatal
}

// Sample records an operation boundary before the next callback or rollback.
// Reusing the tracked context keeps an earlier cancellation from turning into a
// later timeout and prevents cleanup from receiving a fresh grace interval.
func Sample(ctx context.Context, policy InvocationPolicy) error {
	observed := observation(ctx)
	cause, _ := observed.sample(ctx, policy)
	if cause == nil {
		return nil
	}
	return &interruptionError{cause: cause}
}

// Completion preserves a transaction owner's outcome, including a successful
// commit racing cancellation. Interruption is the first sampled boundary; it
// cannot change while a proof is being resolved.
type Completion[T any] struct {
	Value        T
	Err          error
	Interruption error
}

// Complete allows a transaction owner to finish even on a canceled context.
// A rollback callback may use WithoutCancel for its driver call; the surrounding
// lifetime still observes the original cancellation and the unextended grace.
// A non-nil lifetime error is fatal and must precede any late transaction proof.
func Complete[T any](ctx context.Context, policy InvocationPolicy, finish func(context.Context) (T, error)) (Completion[T], error) {
	return invoke(ctx, policy, finish, true)
}

// CompleteWithinGrace permits ordinary terminal publication after semantic
// cancellation, but revokes its context at the original grace's exhaustion.
// It preserves values and joins the callback; it grants neither a fresh grace
// nor authority to replace a commit already proved by the transaction owner.
// Use Complete for physical teardown that must remain joined after exhaustion.
func CompleteWithinGrace[T any](ctx context.Context, policy InvocationPolicy, finish func(context.Context) (T, error)) (Completion[T], error) {
	if finish == nil {
		return Completion[T]{}, errors.New("extension invocation requires a completion callback")
	}
	ctx = ReuseCancellationTracking(ctx)
	observed := observation(ctx)
	return Complete(ctx, policy, func(context.Context) (T, error) {
		if cause, at := observed.sample(ctx, policy); cause != nil && !time.Now().Before(at.Add(policy.Grace)) {
			observed.exhaust(policy)
		}
		publication, cancel := context.WithCancelCause(context.WithoutCancel(ctx))
		observed.mu.Lock()
		if observed.fatal {
			cancel(ErrGraceExhausted)
		} else {
			if observed.publications == nil {
				observed.publications = make(map[*context.CancelCauseFunc]struct{})
			}
			observed.publications[&cancel] = struct{}{}
		}
		observed.mu.Unlock()
		defer func() {
			observed.mu.Lock()
			delete(observed.publications, &cancel)
			observed.mu.Unlock()
			cancel(nil)
		}()
		if err := context.Cause(publication); err != nil {
			var zero T
			return zero, err
		}
		return finish(publication)
	})
}

// Invoke retains the caller's capabilities until the callback has returned or
// the process terminates. Normal callback results after interruption are discarded.
func Invoke[T any](ctx context.Context, policy InvocationPolicy, callback func(context.Context) (T, error)) (T, error) {
	var zero T
	result, lifetimeErr := invoke(ctx, policy, callback, false)
	if lifetimeErr != nil {
		return zero, lifetimeErr
	}
	if result.Interruption != nil {
		return zero, result.Interruption
	}
	return result.Value, result.Err
}

func invoke[T any](ctx context.Context, policy InvocationPolicy, callback func(context.Context) (T, error), complete bool) (Completion[T], error) {
	if policy.Grace < 0 || policy.Exhausted == nil || callback == nil || (policy.Deadline != nil && policy.MonotonicNowNS == nil) {
		return Completion[T]{}, errors.New("extension invocation requires a lifetime policy")
	}
	observed := observation(ctx)
	initial, _ := observed.sample(ctx, policy)
	if initial != nil && !complete {
		if observed.exhausted() {
			return Completion[T]{}, ErrGraceExhausted
		}
		return Completion[T]{Interruption: &interruptionError{cause: initial}}, nil
	}
	type outcome struct {
		value     T
		err       error
		completed time.Time
	}
	done := make(chan outcome, 1)
	go func() {
		var result outcome
		returned := false
		defer func() {
			if recover() != nil || !returned {
				result = outcome{err: errInvocationPanic}
			}
			result.completed = time.Now()
			done <- result
		}()
		result.value, result.err = callback(ctx)
		returned = true
	}()
	classify := func(result outcome) (Completion[T], error) {
		cause, at := observed.sample(ctx, policy)
		completion := Completion[T]{Value: result.value, Err: result.err}
		if cause != nil {
			completion.Interruption = &interruptionError{cause: cause}
			if !result.completed.Before(at.Add(policy.Grace)) {
				observed.exhaust(policy)
			}
		}
		if observed.exhausted() {
			return completion, ErrGraceExhausted
		}
		return completion, nil
	}
	if initial == nil {
		select {
		case result := <-done:
			return classify(result)
		case <-ctx.Done():
		}
	}
	_, at := observed.sample(ctx, policy)
	timer := time.NewTimer(time.Until(at.Add(policy.Grace)))
	defer timer.Stop()
	select {
	case result := <-done:
		return classify(result)
	case <-timer.C:
		// A ready completion is judged by its timestamp, never select ordering.
		select {
		case result := <-done:
			return classify(result)
		default:
		}
	}
	observed.exhaust(policy)
	// The process lifecycle terminates independently. Keep transaction/lock
	// ownership here while the callback is still active. A substituted test
	// lifecycle may permit a late proof, but cannot turn fatal into success.
	return classify(<-done)
}

func invocationContextError(ctx context.Context, policy InvocationPolicy) error {
	if policy.Deadline != nil && policy.Deadline.Expired(policy.MonotonicNowNS()) {
		return context.DeadlineExceeded
	}
	// Deadline observation wins an equal cancellation sample, including the
	// interval before the context timer's goroutine has published cancellation.
	if deadline, ok := ctx.Deadline(); ok && !time.Now().Before(deadline) {
		return context.DeadlineExceeded
	}
	return ctx.Err()
}
