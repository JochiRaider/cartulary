package incidentbundles

import (
	"context"
	"errors"
	"fmt"
	"math"
	"time"

	extensiondeadline "github.com/JochiRaider/cartulary/internal/modules/extensions/deadline"
)

// ErrPortabilityTimeout classifies an Incident Portability participant whose
// Extensions Table 9-E deadline expired. The operation publishes nothing and
// the Job fails with service_unavailable/extension_participant_timeout.
var ErrPortabilityTimeout = errors.New("extension_participant_timeout")

// PortabilityTimeoutError carries only the declared diagnostic facts of a
// participant deadline expiry: the profile and the effective Table 9-C value.
type PortabilityTimeoutError struct {
	ProfileID      string
	Operation      string
	TimeoutSeconds int64
}

func (e *PortabilityTimeoutError) Error() string {
	return fmt.Sprintf("%s: %s %s", ErrPortabilityTimeout, e.ProfileID, e.Operation)
}

func (e *PortabilityTimeoutError) Unwrap() error { return ErrPortabilityTimeout }

var portabilityMonotonicOrigin = time.Now()

// participantDeadline is one Table 9-E Incident Portability participant
// boundary. It starts immediately before the claimed-profile invocation and
// ends after the result and its referenced payload pass schema, digest and
// limit validation (finish).
type participantDeadline struct {
	orchestrator *PortabilityOrchestrator
	parent       context.Context
	ctx          context.Context
	cancel       context.CancelFunc
	bound        extensiondeadline.Deadline
	profileID    string
	operation    string
}

func (o *PortabilityOrchestrator) startParticipantDeadline(ctx context.Context, profileID, operation string) (*participantDeadline, error) {
	if o == nil || o.timeout <= 0 || o.grace < 0 || o.fatalSink == nil || o.clock == nil {
		return nil, newPortabilityFailure(ErrPortabilityUnavailable, profileID)
	}
	now := o.clock()
	bound := extensiondeadline.Deadline{MonotonicNS: saturatingDeadline(now, o.timeout), Source: extensiondeadline.SourceLocal}
	if parent, ok := ctx.Deadline(); ok {
		// Equal deadlines select the inherited parent (EXT-REQ-215).
		if inherited := saturatingDeadline(now, time.Until(parent)); inherited <= bound.MonotonicNS {
			bound = extensiondeadline.Deadline{MonotonicNS: inherited, Source: extensiondeadline.SourceInherited}
		}
	}
	remaining := time.Duration(0)
	if bound.MonotonicNS > now {
		remaining = time.Duration(bound.MonotonicNS - now)
	}
	callCtx, cancel := context.WithTimeout(ctx, remaining)
	return &participantDeadline{
		orchestrator: o, parent: ctx, ctx: extensiondeadline.WithCancellationTracking(callCtx), cancel: cancel,
		bound: bound, profileID: profileID, operation: operation,
	}, nil
}

func (d *participantDeadline) policy() extensiondeadline.InvocationPolicy {
	sink := d.orchestrator.fatalSink
	return extensiondeadline.InvocationPolicy{
		Grace:          d.orchestrator.grace,
		Deadline:       &d.bound,
		MonotonicNowNS: d.orchestrator.clock,
		Exhausted:      func(cause error) { sink(cause) },
	}
}

// invokeParticipant runs one participant callback under the deadline. A result
// returned after expiry is discarded; a callback still active after the
// cancellation grace selects the fatal process outcome.
func invokeParticipant[T any](d *participantDeadline, call func(context.Context) (T, error)) (T, error) {
	result, err := extensiondeadline.Invoke(d.ctx, d.policy(), call)
	if interruption := d.interruption(err); interruption != nil {
		var zero T
		return zero, interruption
	}
	return result, err
}

// finish samples the end boundary after result validation. Expiry there still
// discards the validated result.
func (d *participantDeadline) finish() error {
	defer d.cancel()
	return d.interruption(extensiondeadline.Sample(d.ctx, d.policy()))
}

func (d *participantDeadline) close() { d.cancel() }

// interruption maps an invocation lifetime outcome. Grace exhaustion stays
// fatal. The participant's own expiry is the portability timeout; a parent
// cancellation or an inherited parent deadline keeps the parent's outcome.
func (d *participantDeadline) interruption(err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, extensiondeadline.ErrGraceExhausted) {
		return err
	}
	if !extensiondeadline.Interrupted(err) {
		return nil
	}
	if errors.Is(err, context.DeadlineExceeded) && d.bound.Source == extensiondeadline.SourceLocal && d.bound.Expired(d.orchestrator.clock()) {
		return &PortabilityTimeoutError{ProfileID: d.profileID, Operation: d.operation, TimeoutSeconds: d.orchestrator.timeoutSeconds()}
	}
	if cause := context.Cause(d.parent); cause != nil {
		return cause
	}
	return err
}

// participantLifetimeError reports outcomes that a participant's own result
// must not replace: its deadline expiry, grace exhaustion, and interruption of
// the enclosing operation.
func participantLifetimeError(parent context.Context, err error) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, ErrPortabilityTimeout) || errors.Is(err, extensiondeadline.ErrGraceExhausted) {
		return err
	}
	if parent.Err() != nil {
		if cause := context.Cause(parent); cause != nil {
			return cause
		}
		return parent.Err()
	}
	return nil
}

func saturatingDeadline(now int64, timeout time.Duration) int64 {
	if timeout <= 0 {
		return now
	}
	if now > math.MaxInt64-int64(timeout) {
		return math.MaxInt64
	}
	return now + int64(timeout)
}

func defaultPortabilityClock() int64 { return time.Since(portabilityMonotonicOrigin).Nanoseconds() }

// Timeout is the effective timeouts.extensions.portability_participant_seconds
// value (Extensions Table 9-C) used by diagnostics.
func (o *PortabilityOrchestrator) Timeout() time.Duration {
	if o == nil {
		return 0
	}
	return o.timeout
}

// timeoutSeconds is Timeout as the declared integer timeout_seconds detail.
func (o *PortabilityOrchestrator) timeoutSeconds() int64 {
	return int64(o.Timeout() / time.Second)
}
