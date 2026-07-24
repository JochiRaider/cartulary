package crossownertransaction

import (
	"context"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"
)

func TestCoordinator_Unit_ExecutesDeterministicBoundedProtocol(t *testing.T) {
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
	backend := &backendFake{tx: &transactionFake{}}
	coordinator := newTestCoordinator(t, []Descriptor{descriptor("a")}, backend, nil, nil)
	cases := []struct {
		name        string
		participant Participant
		want        error
	}{
		{name: "empty input", participant: &participantFake{id: "a", input: []byte{}}, want: ErrInput},
		{name: "participant input overflow", participant: &participantFake{id: "a", input: make([]byte, ParticipantInputByteLimit+1)}, want: ErrInput},
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
		Backend: backend, Catalog: catalog, Timeout: time.Minute, Clock: clock, FatalSink: fatal,
	})
	if err != nil {
		t.Fatal(err)
	}
	return coordinator
}

type participantFake struct {
	id            string
	input         []byte
	keys          []SerializationKey
	validation    ValidationResult
	events        *[]string
	cancelOnWrite context.CancelFunc
}

func (p *participantFake) ID() string { return p.id }

func (p *participantFake) BuildInput(context.Context, OperationContext) (Input, error) {
	p.event("input:" + p.id)
	input := p.input
	if input == nil {
		input = []byte(`{"ok":true}`)
	}
	return Input{SchemaID: p.id + ".input", CanonicalBytes: input}, nil
}

func (p *participantFake) Prepare(context.Context, Invocation) (PrepareResult, error) {
	p.event("prepare:" + p.id)
	return PrepareResult{SerializationKeys: p.keys}, nil
}

func (p *participantFake) Validate(context.Context, Invocation) (ValidationResult, error) {
	p.event("validate:" + p.id)
	if p.validation.Status == "" {
		return Valid(), nil
	}
	return p.validation, nil
}

func (p *participantFake) Write(context.Context, Invocation) (WriteResult, error) {
	p.event("write:" + p.id)
	if p.cancelOnWrite != nil {
		p.cancelOnWrite()
	}
	return Written("value:" + p.id), nil
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
	events        *[]string
	commitOutcome CommitOutcome
	commitErr     error
	commitCalls   int
	rolledBack    bool
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
func (t *transactionFake) Commit(context.Context) (CommitOutcome, error) {
	t.commitCalls++
	t.event("commit")
	return t.commitOutcome, t.commitErr
}
func (t *transactionFake) Rollback(context.Context) (CommitOutcome, error) {
	t.rolledBack = true
	t.event("rollback")
	return CommitAbsent, nil
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

type finalizerFake struct{ events *[]string }

func (f finalizerFake) Publish(context.Context, FinalizationCapability, map[string]any) error {
	*f.events = append(*f.events, "finalize")
	return nil
}

type sqlStateError string

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
