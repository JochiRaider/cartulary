package crossownertransaction

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync/atomic"
	"testing"
	"time"
)

func TestCoordinator_Unit_ExecutesDeterministicBoundedProtocol(t *testing.T) {
	assertOwnerPreconditionBoundary(t)
	assertRollbackConflictOutcomes(t)
	assertRollbackInterruptionOutcomes(t)
	events := []string{}
	tx := &transactionFake{events: &events, commitOutcome: CommitProven}
	coordinator := newTestCoordinator(t, []Descriptor{descriptor("a"), descriptor("b")}, &backendFake{tx: tx}, nil, nil)
	finalizer := finalizerFake{events: &events}
	result, err := coordinator.Execute(context.Background(), Operation{
		OperationID: "operation-1", NormalizedRequestSHA256: strings.Repeat("a", 64),
		Participants: []Participant{
			&participantFake{id: "b", events: &events, keys: []SerializationKey{{KeyKind: "b.kind", Key: "2"}}},
			&participantFake{id: "a", events: &events, keys: []SerializationKey{{KeyKind: "a.kind", Key: "1"}}},
		},
		Finalizer: finalizer,
	})
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"input:a", "input:b", "prepare:a", "prepare:b",
		"begin", "lock:a:a.kind:1", "lock:b:b.kind:2",
		"read:a", "validate:a", "read:b", "validate:b",
		"write-cap:a", "write:a", "write-cap:b", "write:b",
		"final-cap", "finalize", "commit",
	}
	if !reflect.DeepEqual(events, want) {
		t.Fatalf("events = %#v, want %#v", events, want)
	}
	if result.ParticipantValues["a"] != "value:a" || result.ParticipantValues["b"] != "value:b" {
		t.Fatalf("participant results = %#v", result.ParticipantValues)
	}
}

func TestCoordinator_Unit_RejectsLimitsAndMalformedKeysBeforeTransaction(t *testing.T) {
	assertProtocolByteLimits(t)
	assertProtocolContractAdmission(t)
	backend := &backendFake{tx: &transactionFake{}}
	coordinator := newTestCoordinator(t, []Descriptor{descriptor("a")}, backend, nil, nil)
	cases := []struct {
		name        string
		participant Participant
		want        error
	}{
		{name: "empty input", participant: &participantFake{id: "a", input: []byte{}}, want: ErrInput},
		{name: "participant input overflow", participant: &participantFake{id: "a", input: make([]byte, int(coordinator.limits.ParticipantInputBytes)+1)}, want: ErrInput},
		{name: "undeclared key", participant: &participantFake{id: "a", keys: []SerializationKey{{KeyKind: "wrong", Key: "x"}}}, want: ErrSerializationKeys},
		{name: "duplicate key", participant: &participantFake{id: "a", keys: []SerializationKey{{KeyKind: "a.kind", Key: "x"}, {KeyKind: "a.kind", Key: "x"}}}, want: ErrSerializationKeys},
	}
	for _, test := range cases {
		t.Run(test.name, func(t *testing.T) {
			backend.beginCalls = 0
			_, err := coordinator.Execute(context.Background(), Operation{
				OperationID: "operation", NormalizedRequestSHA256: "digest",
				Participants: []Participant{test.participant},
			})
			if !errors.Is(err, test.want) || backend.beginCalls != 0 {
				t.Fatalf("Execute() = %v, begins=%d", err, backend.beginCalls)
			}
		})
	}
}

func TestCoordinator_Unit_CancellationValidationAndRollbackBoundaries(t *testing.T) {
	assertParticipantLifetime(t)
	assertParticipantPrecedence(t)
	t.Run("canceled before begin", func(t *testing.T) {
		backend := &backendFake{tx: &transactionFake{}}
		coordinator := newTestCoordinator(t, []Descriptor{descriptor("a")}, backend, nil, nil)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		_, err := coordinator.Execute(ctx, operation(&participantFake{id: "a"}))
		if !errors.Is(err, ErrCanceled) || backend.beginCalls != 0 {
			t.Fatalf("Execute() = %v, begins=%d", err, backend.beginCalls)
		}
	})
	t.Run("first invalid participant stops validation and rolls back", func(t *testing.T) {
		events := []string{}
		tx := &transactionFake{events: &events}
		coordinator := newTestCoordinator(t, []Descriptor{descriptor("a"), descriptor("b")}, &backendFake{tx: tx}, nil, nil)
		_, err := coordinator.Execute(context.Background(), Operation{
			OperationID: "operation", NormalizedRequestSHA256: "digest",
			Participants: []Participant{
				&participantFake{id: "a", events: &events, validation: ValidationResult{
					Status: "invalid", Findings: []Finding{{Path: "/x", ReasonCode: "invalid", Message: "invalid", Details: []byte("{}")}},
				}},
				&participantFake{id: "b", events: &events},
			},
		})
		if !errors.Is(err, ErrValidation) || !tx.rolledBack || contains(events, "validate:b") || contains(events, "write:a") {
			t.Fatalf("Execute() = %v, events=%v rollback=%t", err, events, tx.rolledBack)
		}
	})
	t.Run("cancellation after write proves absence", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		tx := &transactionFake{}
		coordinator := newTestCoordinator(t, []Descriptor{descriptor("a")}, &backendFake{tx: tx}, nil, nil)
		_, err := coordinator.Execute(ctx, operation(&participantFake{id: "a", cancelOnWrite: cancel}))
		if !errors.Is(err, ErrCanceled) || !tx.rolledBack || tx.commitCalls != 0 {
			t.Fatalf("Execute() = %v rollback=%t commits=%d", err, tx.rolledBack, tx.commitCalls)
		}
	})
}

func TestCoordinator_Unit_ClosedCommitOutcomesAndNoRetry(t *testing.T) {
	assertTransactionCompletionLifetime(t)
	t.Run("proven serialization absence is conflict", func(t *testing.T) {
		tx := &transactionFake{commitOutcome: CommitAbsent, commitErr: sqlStateError("40001")}
		backend := &backendFake{tx: tx}
		coordinator := newTestCoordinator(t, []Descriptor{descriptor("a")}, backend, nil, nil)
		_, err := coordinator.Execute(context.Background(), operation(&participantFake{id: "a"}))
		var conflict *ConflictError
		if !errors.As(err, &conflict) || conflict.ReasonCode != "serialization_failure" || backend.beginCalls != 1 || tx.commitCalls != 1 {
			t.Fatalf("Execute() = %v, begins=%d commits=%d", err, backend.beginCalls, tx.commitCalls)
		}
	})
	t.Run("indeterminate commit is fatal", func(t *testing.T) {
		fatalCalls := 0
		tx := &transactionFake{commitOutcome: CommitUnknown, commitErr: errors.New("connection lost")}
		coordinator := newTestCoordinator(t, []Descriptor{descriptor("a")}, &backendFake{tx: tx}, nil, func(error) { fatalCalls++ })
		_, err := coordinator.Execute(context.Background(), operation(&participantFake{id: "a"}))
		if !IsFatalIntegrity(err) || fatalCalls != 1 || tx.rolledBack {
			t.Fatalf("Execute() = %v, fatal=%d rollback=%t", err, fatalCalls, tx.rolledBack)
		}
	})
}

func TestCoordinator_Unit_DeadlineOverflowAndBoundaryTie(t *testing.T) {
	calls := 0
	clock := func() int64 {
		calls++
		if calls == 1 {
			return 100
		}
		return 100 + int64(time.Second)
	}
	backend := &backendFake{tx: &transactionFake{}}
	coordinator := newTestCoordinator(t, []Descriptor{descriptor("a")}, backend, clock, nil)
	_, err := coordinator.Execute(context.Background(), Operation{
		OperationID: "operation", NormalizedRequestSHA256: "digest",
		Participants: []Participant{&participantFake{id: "a"}}, Timeout: time.Second,
	})
	if !errors.Is(err, ErrTimeout) || backend.beginCalls != 0 {
		t.Fatalf("deadline tie = %v, begins=%d", err, backend.beginCalls)
	}
}

func descriptor(id string) Descriptor {
	return Descriptor{
		ParticipantID: id, OwnerProfileID: "profile", ContractSHA256: "digest",
		InputSchemaID: id + ".input", PrepareAlgorithmID: id + ".prepare",
		ValidationAlgorithmID: id + ".validate", WriteAlgorithmID: id + ".write",
		SerializationKeyKinds: []string{id + ".kind"}, OwnedStateFamilyIDs: []string{id + ".state"},
	}
}

func operation(participant Participant) Operation {
	return Operation{OperationID: "operation", NormalizedRequestSHA256: "digest", Participants: []Participant{participant}}
}

func newTestCoordinator(t *testing.T, catalog []Descriptor, backend Backend, clock func() int64, fatal func(error)) *Coordinator {
	t.Helper()
	if fatal == nil {
		fatal = func(err error) { t.Fatalf("unexpected fatal sink: %v", err) }
	}
	coordinator, err := New(Options{
		Limits:           protocolLimitsFixture(),
		ValidationPolicy: validationPolicyFixture{},
		Backend:          backend, Catalog: catalog, Timeout: time.Minute, Clock: clock, FatalSink: fatal, CancellationGrace: 100 * time.Millisecond,
	})
	if err != nil {
		t.Fatal(err)
	}
	return coordinator
}

type participantFake struct {
	id              string
	input           []byte
	keys            []SerializationKey
	validation      ValidationResult
	validationErr   error
	writeErr        error
	validationPanic bool
	events          *[]string
	cancelOnWrite   context.CancelFunc
	block           func(context.Context, string)
}

func (p *participantFake) ID() string { return p.id }

func (p *participantFake) BuildInput(ctx context.Context, _ OperationContext) (Input, error) {
	if p.block != nil {
		p.block(ctx, "input")
	}
	p.event("input:" + p.id)
	input := p.input
	if input == nil {
		input = []byte("{\"ok\":true}\n")
	}
	return Input{SchemaID: p.id + ".input", CanonicalBytes: input}, nil
}

func (p *participantFake) Prepare(ctx context.Context, _ Invocation) (PrepareResult, error) {
	if p.block != nil {
		p.block(ctx, "prepare")
	}
	p.event("prepare:" + p.id)
	return PrepareResult{SerializationKeys: append([]SerializationKey{}, p.keys...)}, nil
}

func (p *participantFake) Validate(ctx context.Context, _ Invocation) (ValidationResult, error) {
	if p.block != nil {
		p.block(ctx, "validate")
	}
	p.event("validate:" + p.id)
	if p.validationPanic {
		panic("private participant bytes")
	}
	if p.validationErr != nil {
		return p.validation, p.validationErr
	}
	if p.validation.Status == "" {
		return Valid(), nil
	}
	return p.validation, nil
}

func (p *participantFake) Write(ctx context.Context, _ Invocation) (WriteResult, error) {
	if p.block != nil {
		p.block(ctx, "write")
	}
	p.event("write:" + p.id)
	if p.cancelOnWrite != nil {
		p.cancelOnWrite()
	}
	return Written("value:" + p.id), p.writeErr
}

func (p *participantFake) event(value string) {
	if p.events != nil {
		*p.events = append(*p.events, value)
	}
}

type backendFake struct {
	tx         *transactionFake
	beginCalls int
}

func (b *backendFake) Begin(context.Context, []Descriptor) (Transaction, error) {
	b.beginCalls++
	if b.tx.events != nil {
		*b.tx.events = append(*b.tx.events, "begin")
	}
	return b.tx, nil
}

type transactionFake struct {
	events          *[]string
	commitOutcome   CommitOutcome
	commitErr       error
	commitCalls     int
	rolledBack      bool
	rollbackOutcome CommitOutcome
	rollbackErr     error
	rollbackCalls   atomic.Int64
	block           func(context.Context, string)
}

func (t *transactionFake) AcquireSerializationLock(_ context.Context, key OrderedSerializationKey) error {
	t.event("lock:" + key.ParticipantID + ":" + key.KeyKind + ":" + key.Key)
	return nil
}
func (t *transactionFake) ReadCapability(id string) (ReadCapability, error) {
	t.event("read:" + id)
	return capabilityFake{id: id}, nil
}
func (t *transactionFake) WriteCapability(id string) (WriteCapability, error) {
	t.event("write-cap:" + id)
	return capabilityFake{id: id}, nil
}
func (t *transactionFake) FinalizationCapability() (FinalizationCapability, error) {
	t.event("final-cap")
	return finalizationCapabilityFake{}, nil
}
func (t *transactionFake) PreconditionCapability() (PreconditionCapability, error) {
	t.event("precondition-cap")
	return preconditionCapabilityFake{}, nil
}

type preconditionCapabilityFake struct{}

func (preconditionCapabilityFake) PreconditionScope() string { return "shared.preconditions" }

type preconditionsFake struct {
	events *[]string
	err    error
}

func (p preconditionsFake) Check(_ context.Context, access PreconditionCapability) error {
	if access.PreconditionScope() != "shared.preconditions" {
		return ErrUnavailable
	}
	*p.events = append(*p.events, "preconditions")
	return p.err
}

func assertOwnerPreconditionBoundary(t *testing.T) {
	t.Helper()
	t.Run("owner preconditions precede participant validation and mutation", func(t *testing.T) {
		events := []string{}
		tx := &transactionFake{events: &events, commitOutcome: CommitProven}
		coordinator := newTestCoordinator(t, []Descriptor{descriptor("a")}, &backendFake{tx: tx}, nil, nil)
		rejected := errors.New("current owner precondition rejected")
		_, err := coordinator.Execute(context.Background(), Operation{
			OperationID: "unit-1", NormalizedRequestSHA256: strings.Repeat("a", 64),
			Participants:  []Participant{&participantFake{id: "a", events: &events, keys: []SerializationKey{{KeyKind: "a.kind", Key: "1"}}}},
			Preconditions: preconditionsFake{events: &events, err: rejected},
		})
		want := []string{"input:a", "prepare:a", "begin", "lock:a:a.kind:1", "precondition-cap", "preconditions", "rollback"}
		if !errors.Is(err, rejected) || !reflect.DeepEqual(events, want) || tx.commitCalls != 0 {
			t.Fatalf("precondition rejection: err=%v events=%v commits=%d; want %v", err, events, tx.commitCalls, want)
		}
	})
}
func (t *transactionFake) Commit(ctx context.Context) (CommitOutcome, error) {
	t.commitCalls++
	if t.block != nil {
		t.block(ctx, "commit")
	}
	t.event("commit")
	return t.commitOutcome, t.commitErr
}
func (t *transactionFake) Rollback(ctx context.Context) (CommitOutcome, error) {
	t.rollbackCalls.Add(1)
	if t.block != nil {
		t.block(ctx, "rollback")
	}
	t.rolledBack = true
	t.event("rollback")
	if t.rollbackOutcome != "" {
		return t.rollbackOutcome, t.rollbackErr
	}
	return CommitAbsent, t.rollbackErr
}
func (t *transactionFake) event(value string) {
	if t.events != nil {
		*t.events = append(*t.events, value)
	}
}

type capabilityFake struct{ id string }

func (c capabilityFake) ParticipantScope() string { return c.id }

type finalizationCapabilityFake struct{}

func (finalizationCapabilityFake) FinalizationScope() string { return "shared.finalization" }

type finalizerFake struct {
	err    error
	events *[]string
	block  func(context.Context, string)
}

func (f finalizerFake) Publish(ctx context.Context, _ FinalizationCapability, _ map[string]any) error {
	if f.block != nil {
		f.block(ctx, "finalize")
	}
	if f.events != nil {
		*f.events = append(*f.events, "finalize")
	}
	return f.err
}

type sqlStateError string

func assertRollbackConflictOutcomes(t *testing.T) {
	t.Helper()
	for _, state := range []string{"40P01", "40001"} {
		for _, wrapped := range []bool{false, true} {
			for _, phase := range []string{"preconditions", "write", "finalize"} {
				t.Run(fmt.Sprintf("rollback conflict %s %t %s", state, wrapped, phase), func(t *testing.T) {
					events := []string{}
					tx := &transactionFake{events: &events}
					backend := &backendFake{tx: tx}
					c := newTestCoordinator(t, []Descriptor{descriptor("a")}, backend, nil, nil)
					cause := error(sqlStateError(state))
					if wrapped {
						cause = fmt.Errorf("owner database operation: %w", cause)
					}
					p := &participantFake{id: "a", events: &events}
					op := operation(p)
					switch phase {
					case "preconditions":
						op.Preconditions = preconditionsFake{events: &events, err: cause}
					case "write":
						p.writeErr = cause
					case "finalize":
						op.Finalizer = finalizerFake{events: &events, err: cause}
					}
					result, err := c.Execute(t.Context(), op)
					want := "deadlock_detected"
					if state == "40001" {
						want = "serialization_failure"
					}
					var conflict *ConflictError
					if !errors.As(err, &conflict) || conflict.ReasonCode != want || !errors.Is(err, cause) || backend.beginCalls != 1 || tx.commitCalls != 0 || tx.rollbackCalls.Load() != 1 || len(result.ParticipantValues) != 0 {
						t.Fatalf("conflict outcome: err=%v events=%v begins=%d commits=%d rollbacks=%d result=%v", err, events, backend.beginCalls, tx.commitCalls, tx.rollbackCalls.Load(), result)
					}
					if phase == "preconditions" && (contains(events, "validate:a") || contains(events, "write:a")) {
						t.Fatal("failed precondition reached a participant")
					}
				})
			}
		}
	}
	for _, terminal := range []error{ErrCanceled, ErrTimeout, &FatalIntegrityError{Cause: errors.New("already fatal")}} {
		t.Run("terminal precedence "+terminal.Error(), func(t *testing.T) {
			events := []string{}
			tx := &transactionFake{}
			c := newTestCoordinator(t, []Descriptor{descriptor("a")}, &backendFake{tx: tx}, nil, nil)
			cause := errors.Join(terminal, sqlStateError("40001"))
			op := operation(&participantFake{id: "a"})
			op.Preconditions = preconditionsFake{events: &events, err: cause}
			_, err := c.Execute(t.Context(), op)
			var conflict *ConflictError
			if !errors.Is(err, terminal) || errors.As(err, &conflict) {
				t.Fatalf("terminal outcome replaced: %v", err)
			}
		})
	}
	for _, state := range []string{"40P01", "40001"} {
		t.Run("uncertain rollback "+state, func(t *testing.T) {
			events := []string{}
			tx := &transactionFake{rollbackOutcome: CommitUnknown, rollbackErr: errors.New("rollback acknowledgement lost")}
			fatal := 0
			c := newTestCoordinator(t, []Descriptor{descriptor("a")}, &backendFake{tx: tx}, nil, func(error) { fatal++ })
			op := operation(&participantFake{id: "a"})
			op.Preconditions = preconditionsFake{events: &events, err: sqlStateError(state)}
			_, err := c.Execute(t.Context(), op)
			var conflict *ConflictError
			if !IsFatalIntegrity(err) || errors.As(err, &conflict) || fatal != 1 || tx.commitCalls != 0 {
				t.Fatalf("uncertain rollback downgraded: %v fatal=%d", err, fatal)
			}
		})
		t.Run("validation invocation classification "+state, func(t *testing.T) {
			c := newTestCoordinator(t, []Descriptor{descriptor("a")}, &backendFake{tx: &transactionFake{}}, nil, nil)
			_, err := c.Execute(t.Context(), operation(&participantFake{id: "a", validationErr: sqlStateError(state)}))
			var validation *ValidationError
			var conflict *ConflictError
			if !errors.As(err, &validation) || validation.ReasonCode != "extension_admission_validation_failed" || errors.As(err, &conflict) {
				t.Fatalf("validation classifier bypassed: %v", err)
			}
		})
	}
}

func assertRollbackInterruptionOutcomes(t *testing.T) {
	t.Helper()
	type scenario struct {
		name, phase, interruption string
		cause                     error
		want                      error
		earlyCancel, unknown      bool
	}
	var cases []scenario
	for _, state := range []string{"40P01", "40001"} {
		for _, wrapped := range []bool{false, true} {
			cause := error(sqlStateError(state))
			if wrapped {
				cause = fmt.Errorf("owner database operation: %w", cause)
			}
			for _, phase := range []string{"preconditions", "write", "finalize"} {
				for _, interruption := range []string{"cancel", "deadline", "both"} {
					want := ErrTimeout
					if interruption == "cancel" {
						want = ErrCanceled
					}
					cases = append(cases, scenario{
						name:  fmt.Sprintf("%s wrapped=%t %s %s", state, wrapped, phase, interruption),
						phase: phase, cause: cause, interruption: interruption, want: want,
					})
				}
			}
		}
		cases = append(cases,
			scenario{name: state + " first cancellation survives expiry", phase: "write", cause: sqlStateError(state), earlyCancel: true, interruption: "deadline", want: ErrCanceled},
			scenario{name: state + " uncertain rollback stays fatal", phase: "preconditions", cause: sqlStateError(state), interruption: "cancel", unknown: true},
		)
		for _, interruption := range []string{"cancel", "deadline"} {
			want := ErrCanceled
			if interruption == "deadline" {
				want = ErrTimeout
			}
			cases = append(cases, scenario{name: state + " typed conflict " + interruption, phase: "preconditions", cause: classifyDatabaseError(sqlStateError(state)), interruption: interruption, want: want})
		}
	}
	ownerErr := errors.New("ordinary owner denial")
	for _, interruption := range []string{"none", "cancel", "deadline"} {
		want := ownerErr
		if interruption == "cancel" {
			want = ErrCanceled
		} else if interruption == "deadline" {
			want = ErrTimeout
		}
		cases = append(cases, scenario{name: "owner failure " + interruption, phase: "preconditions", cause: ownerErr, interruption: interruption, want: want})
		if interruption == "none" {
			want = ErrValidation
		}
		cases = append(cases, scenario{name: "registered validation failure " + interruption, phase: "validate", cause: sqlStateError("40001"), interruption: interruption, want: want})
		if interruption != "none" {
			for _, terminal := range []error{ErrCanceled, ErrTimeout, &FatalIntegrityError{Cause: errors.New("already fatal")}} {
				cases = append(cases, scenario{name: "selected " + terminal.Error() + " " + interruption, phase: "preconditions", cause: terminal, interruption: interruption, want: terminal})
			}
		}
	}
	for _, tc := range cases {
		t.Run("rollback interruption "+tc.name, func(t *testing.T) {
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			var clock atomic.Int64
			events := []string{}
			tx := &transactionFake{events: &events, block: func(_ context.Context, phase string) {
				if phase != "rollback" {
					return
				}
				// The interruption originates inside completion, after the failure.
				// For a tie, make expiry observable before cancellation is signaled.
				if tc.interruption == "deadline" || tc.interruption == "both" {
					clock.Store(int64(time.Minute))
				}
				if tc.interruption == "cancel" || tc.interruption == "both" {
					cancel()
				}
			}}
			if tc.unknown {
				tx.rollbackOutcome = CommitUnknown
				tx.rollbackErr = errors.New("rollback acknowledgement lost")
			}
			backend := &backendFake{tx: tx}
			fatalCalls := 0
			c := newTestCoordinator(t, []Descriptor{descriptor("a")}, backend, clock.Load, func(error) { fatalCalls++ })
			p := &participantFake{id: "a", events: &events}
			op := operation(p)
			switch tc.phase {
			case "preconditions":
				op.Preconditions = preconditionsFake{events: &events, err: tc.cause}
			case "validate":
				p.validationErr = tc.cause
			case "write":
				p.writeErr = tc.cause
				if tc.earlyCancel {
					p.cancelOnWrite = cancel
				}
			case "finalize":
				op.Finalizer = finalizerFake{events: &events, err: tc.cause}
			}
			result, err := c.Execute(ctx, op)
			if tc.unknown {
				if !IsFatalIntegrity(err) || fatalCalls != 1 {
					t.Errorf("uncertain rollback = %v, fatal calls=%d", err, fatalCalls)
				}
			} else if !errors.Is(err, tc.want) || fatalCalls != 0 {
				t.Errorf("rollback outcome = %v, want %v; fatal calls=%d", err, tc.want, fatalCalls)
			}
			if tc.phase == "validate" && tc.interruption == "none" {
				var validation *ValidationError
				if !errors.As(err, &validation) || validation.ReasonCode != "extension_admission_validation_failed" {
					t.Errorf("registered validation outcome changed: %v", err)
				}
			}
			if backend.beginCalls != 1 || tx.rollbackCalls.Load() != 1 || tx.commitCalls != 0 || !reflect.DeepEqual(result, Result{}) {
				t.Errorf("transaction effects: begins=%d rollbacks=%d commits=%d result=%v", backend.beginCalls, tx.rollbackCalls.Load(), tx.commitCalls, result)
			}
			if tc.phase == "preconditions" && (contains(events, "validate:a") || contains(events, "write:a")) {
				t.Error("failed precondition reached a participant")
			}
		})
	}
}

func (e sqlStateError) Error() string    { return string(e) }
func (e sqlStateError) SQLState() string { return string(e) }

func contains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}

// A small port fixture for coordinator transaction tests; real registry policy
// admission is exercised by the application composition integration tests.
type validationPolicyFixture struct{}

func (validationPolicyFixture) AcceptDescriptor(d Descriptor) bool {
	return d.ParticipantID != "" && d.InputSchemaID == d.ParticipantID+".input"
}
func (validationPolicyFixture) AcceptInput(_ Descriptor, value any) bool {
	if _, ok := value.(string); ok {
		return true
	}
	object, ok := value.(map[string]any)
	if !ok || len(object) != 1 {
		return false
	}
	_, ok = object["ok"].(bool)
	return ok
}
func (validationPolicyFixture) AcceptFinding(_ Descriptor, finding Finding) bool {
	return finding.Path == "$.value" && finding.ReasonCode == "fixture_invalid" && finding.Message == "Invalid value." && string(finding.Details) == "{}"
}
func (validationPolicyFixture) ConditionReason(id string) (string, bool) {
	switch id {
	case "extension_validation_result_structural_invalid_transaction", "extension_validation_remaining_schema_defect_transaction":
		return "extension_validation_result_invalid", true
	case "extension_validation_result_overflow_transaction":
		return "extension_diagnostic_overflow", true
	case "extension_participant_validation_invocation_failure":
		return "extension_admission_validation_failed", true
	}
	return "", false
}

func (p validationPolicyFixture) FailureFinding(id string, owner Descriptor, timeoutSeconds int64, timedOut bool) (Finding, bool) {
	reason, ok := p.ConditionReason(id)
	if !ok {
		return Finding{}, false
	}
	message := "An extension validation algorithm returned an invalid result."
	details := map[string]any{"profile_id": owner.OwnerProfileID, "phase": "transaction", "algorithm_id": owner.ValidationAlgorithmID, "actual": "redacted"}
	switch reason {
	case "extension_diagnostic_overflow":
		message = "Extension validation produced too many findings."
		details = map[string]any{"phase": "transaction", "limit": 4096, "actual": 4097}
	case "extension_admission_validation_failed":
		message = "Extension admission validation failed."
		details = map[string]any{"profile_id": owner.OwnerProfileID, "phase": "transaction", "algorithm_id": owner.ValidationAlgorithmID, "timed_out": timedOut, "timeout_seconds": timeoutSeconds}
	}
	encoded, _ := json.Marshal(details)
	return Finding{Path: "$", ReasonCode: reason, Message: message, Details: encoded}, true
}

func assertParticipantLifetime(t *testing.T) {
	t.Helper()
	for _, phase := range []string{"input", "prepare", "validate", "write", "finalize"} {
		for _, grace := range []time.Duration{0, 5 * time.Millisecond} {
			t.Run("active_"+phase+"_"+grace.String(), func(t *testing.T) {
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				entered, release := make(chan struct{}), make(chan struct{})
				fatal := make(chan error, 1)
				done := make(chan error, 1)
				tx := &transactionFake{}
				coordinator := newTestCoordinator(t, []Descriptor{descriptor("a")}, &backendFake{tx: tx}, nil, func(err error) { fatal <- err })
				coordinator.cancellationGrace = grace
				participant := &participantFake{id: "a", block: func(_ context.Context, current string) {
					if current == phase {
						close(entered)
						<-release
					}
				}}
				op := operation(participant)
				if phase == "finalize" {
					op.Finalizer = finalizerFake{block: participant.block}
				}
				go func() { _, err := coordinator.Execute(ctx, op); done <- err }()
				select {
				case <-entered:
				case err := <-done:
					close(release)
					t.Fatalf("operation ended before callback: %v", err)
				case <-time.After(time.Second):
					close(release)
					cancel()
					t.Fatal("callback did not start")
				}
				cancel()
				select {
				case err := <-fatal:
					var failure *FatalIntegrityError
					if !errors.As(err, &failure) || failure.FatalReasonCode() != "operation_cancellation_grace_exhausted" {
						t.Errorf("fatal outcome = %v", err)
					}
				case <-time.After(time.Second):
					close(release)
					<-done
					t.Fatal("active participant did not exhaust grace")
				}
				if tx.rollbackCalls.Load() != 0 {
					t.Error("transaction rollback raced an active participant")
				}
				select {
				case err := <-done:
					close(release)
					t.Fatalf("active participant escaped: %v", err)
				default:
				}
				close(release)
				if err := <-done; !IsFatalIntegrity(err) {
					t.Fatalf("late participant result replaced fatal outcome: %v", err)
				}
				wantRollback := int64(0)
				if phase == "validate" || phase == "write" || phase == "finalize" {
					wantRollback = 1
				}
				if tx.rollbackCalls.Load() != wantRollback || tx.commitCalls != 0 {
					t.Fatalf("rollback=%d commit=%d", tx.rollbackCalls.Load(), tx.commitCalls)
				}
				select {
				case err := <-fatal:
					t.Fatalf("duplicate fatal signal: %v", err)
				default:
				}
			})
		}
	}
}

func assertTransactionCompletionLifetime(t *testing.T) {
	t.Helper()
	for _, outcome := range []CommitOutcome{CommitProven, CommitAbsent, CommitUnknown} {
		t.Run("commit_proof_after_cancellation_"+string(outcome), func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			fatalCalls := 0
			tx := &transactionFake{commitOutcome: outcome, block: func(_ context.Context, phase string) {
				if phase == "commit" {
					cancel()
				}
			}}
			c := newTestCoordinator(t, []Descriptor{descriptor("a")}, &backendFake{tx: tx}, nil, func(error) { fatalCalls++ })
			_, err := c.Execute(ctx, operation(&participantFake{id: "a"}))
			switch outcome {
			case CommitProven:
				if err != nil {
					t.Fatalf("proven commit lost to cancellation: %v", err)
				}
			case CommitAbsent:
				if !errors.Is(err, ErrCanceled) {
					t.Fatalf("absent commit lost cancellation: %v", err)
				}
			case CommitUnknown:
				if !IsFatalIntegrity(err) || fatalCalls != 1 {
					t.Fatalf("unknown commit = %v, fatal=%d", err, fatalCalls)
				}
			}
			if tx.rollbackCalls.Load() != 0 || tx.commitCalls != 1 {
				t.Fatalf("commit/rollback retried: %d/%d", tx.commitCalls, tx.rollbackCalls.Load())
			}
		})
	}
	for _, tc := range []struct {
		phase      string
		cause      error
		validation bool
	}{
		{phase: "commit"},
		{phase: "rollback", cause: errors.New("invalid validation"), validation: true},
		{phase: "rollback", cause: sqlStateError("40P01")},
		{phase: "rollback", cause: sqlStateError("40001")},
	} {
		t.Run(fmt.Sprintf("blocked_transaction_%s_%v", tc.phase, tc.cause), func(t *testing.T) {
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			entered, release := make(chan struct{}), make(chan struct{})
			fatal, done := make(chan error, 2), make(chan error, 1)
			tx := &transactionFake{commitOutcome: CommitProven, block: func(_ context.Context, current string) {
				if current == tc.phase {
					close(entered)
					<-release
				}
			}}
			backend := &backendFake{tx: tx}
			c := newTestCoordinator(t, []Descriptor{descriptor("a")}, backend, nil, func(err error) { fatal <- err })
			c.cancellationGrace = 5 * time.Millisecond
			participant := &participantFake{id: "a"}
			if tc.validation {
				participant.validationErr = tc.cause
			} else if tc.phase == "rollback" {
				participant.writeErr = tc.cause
			}
			var result Result
			go func() {
				var err error
				result, err = c.Execute(ctx, operation(participant))
				done <- err
			}()
			select {
			case <-entered:
			case err := <-done:
				close(release)
				t.Fatalf("completion not entered: %v", err)
			case <-time.After(time.Second):
				close(release)
				t.Fatal("completion not entered")
			}
			cancel()
			select {
			case err := <-fatal:
				var failure *FatalIntegrityError
				if !errors.As(err, &failure) || failure.FatalReasonCode() != "operation_cancellation_grace_exhausted" {
					t.Errorf("fatal = %v", err)
				}
			case <-time.After(time.Second):
				close(release)
				<-done
				t.Fatal("blocked completion bypassed grace")
			}
			select {
			case err := <-done:
				close(release)
				t.Fatalf("completion escaped while active: %v", err)
			default:
			}
			close(release)
			if err := <-done; !IsFatalIntegrity(err) {
				t.Fatalf("late proof became success: %v", err)
			}
			wantCommit, wantRollback := 1, int64(0)
			if tc.phase == "rollback" {
				wantCommit, wantRollback = 0, 1
			}
			if backend.beginCalls != 1 || tx.commitCalls != wantCommit || tx.rollbackCalls.Load() != wantRollback || !reflect.DeepEqual(result, Result{}) {
				t.Fatalf("completion retried or returned values: begins=%d commits=%d rollbacks=%d result=%v", backend.beginCalls, tx.commitCalls, tx.rollbackCalls.Load(), result)
			}
			select {
			case err := <-fatal:
				t.Fatalf("duplicate fatal: %v", err)
			default:
			}
		})
	}
}
