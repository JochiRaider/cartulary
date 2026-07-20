package extensions

import (
	"context"
	"errors"
	"fmt"
	"reflect"
	"sync"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/extensionstore"
)

func assertBC001EmptyStatePolicy(t *testing.T) {
	t.Helper()
	tests := []struct {
		policy        string
		metadata      bool
		authoritative bool
		decision      StatePresenceDecision
		err           error
	}{
		{"allowed", false, false, StateInitialize, nil},
		{"allowed", false, true, "", ErrStateMetadataMissing},
		{"allowed", true, false, StateValidate, nil},
		{"allowed", true, true, StateValidate, nil},
		{"forbidden", false, false, "", ErrStateIncomplete},
		{"forbidden", false, true, "", ErrStateMetadataMissing},
		{"forbidden", true, false, "", ErrStateIncomplete},
		{"forbidden", true, true, StateValidate, nil},
		{"", false, false, "", ErrStateIncomplete},
	}
	for _, test := range tests {
		name := fmt.Sprintf("%s_metadata_%t_state_%t", test.policy, test.metadata, test.authoritative)
		t.Run(name, func(t *testing.T) {
			decision, err := DecideStatePresence(test.policy, test.metadata, test.authoritative)
			if decision != test.decision || !errors.Is(err, test.err) {
				t.Fatalf("decision = %q/%v; want %q/%v", decision, err, test.decision, test.err)
			}
		})
	}
	coordinator, err := NewGeneratedCoordinator()
	if err != nil {
		t.Fatal(err)
	}
	plan, ok := coordinator.StatePlan("network_flow_activity")
	if !ok || plan.EmptyStatePolicy != "allowed" || plan.InitializationKind != "empty" {
		t.Fatalf("Network Flow state plan = %#v/%t", plan, ok)
	}
	if !reflect.DeepEqual(plan.DatabaseFamilyIDs, []string{
		"network_flow_activity.indicator_bindings",
		"network_flow_activity.rejected_row_diagnostics",
		"network_flow_activity.rows",
		"network_flow_activity.tables",
	}) {
		t.Fatalf("authoritative presence families = %#v", plan.DatabaseFamilyIDs)
	}
}

type portabilityAllocatorFake struct {
	allocated []string
	abandoned []string
}

func (a *portabilityAllocatorFake) Allocate(_ context.Context, operationID, profileID string, _ []byte) (string, error) {
	ref := "staged:" + operationID + ":" + profileID
	a.allocated = append(a.allocated, ref)
	return ref, nil
}

func (a *portabilityAllocatorFake) Abandon(_ context.Context, ref string) error {
	a.abandoned = append(a.abandoned, ref)
	return nil
}

type portabilityParticipantFake struct {
	id        string
	called    int
	malformed bool
	fail      bool
	allocate  bool
}

func (p *portabilityParticipantFake) ID() string { return p.id }

func (p *portabilityParticipantFake) PrepareImport(ctx context.Context, payload []byte, scope *StagedOutputScope) (PortabilityImportPreparation, error) {
	p.called++
	if p.allocate {
		if _, err := scope.Allocate(ctx, "operation", []byte("staged")); err != nil {
			return PortabilityImportPreparation{}, err
		}
	}
	if p.fail {
		return PortabilityImportPreparation{}, errors.New("profile detail must not escape")
	}
	result := PortabilityImportPreparation{SchemaID: PortabilityImportResultSchema, Status: "prepared", ParticipantInput: append([]byte(nil), payload...), ParticipantInputSHA256: "digest", StagedOutputRefs: scope.Refs()}
	if p.malformed {
		result.SchemaID = PortabilityExportResultSchema
	}
	return result, nil
}

func assertBC003PortabilitySeparation(t *testing.T) {
	t.Helper()
	if err := (PortabilityExportResult{SchemaID: PortabilityExportResultSchema, Kind: "omit"}).Validate(); err != nil {
		t.Fatal(err)
	}
	if err := (PortabilityExportResult{SchemaID: PortabilityImportResultSchema, Kind: "omit"}).Validate(); !errors.Is(err, ErrPortabilityResultInvalid) {
		t.Fatalf("import result accepted as export: %v", err)
	}
	allocator := &portabilityAllocatorFake{}
	participant := &portabilityParticipantFake{id: "participant"}
	base := PortabilityImportRequest{OperationID: "operation", ProfileID: "profile", Participant: participant, Allocator: allocator}
	if _, _, err := PreparePortabilityImports(context.Background(), []PortabilityImportRequest{base}); !errors.Is(err, ErrPortabilityPayloadMissing) || participant.called != 0 {
		t.Fatalf("absent payload = %v/calls=%d", err, participant.called)
	}
	large := make([]byte, TransactionByteLimit+1)
	over := base
	over.Payload = large
	if _, _, err := PreparePortabilityImports(context.Background(), []PortabilityImportRequest{over}); !errors.Is(err, ErrPortabilityInputLimit) || participant.called != 0 {
		t.Fatalf("64 MiB plus one = %v/calls=%d", err, participant.called)
	}
	exact := base
	exact.Payload = large[:TransactionByteLimit]
	results, scopes, err := PreparePortabilityImports(context.Background(), []PortabilityImportRequest{exact})
	if err != nil || len(results) != 1 || len(scopes) != 1 || participant.called != 1 {
		t.Fatalf("64 MiB preparation = %v/results=%d/scopes=%d/calls=%d", err, len(results), len(scopes), participant.called)
	}
	malformed := &portabilityParticipantFake{id: "participant", malformed: true}
	bad := base
	bad.Payload = []byte("payload")
	bad.Participant = malformed
	if _, _, err := PreparePortabilityImports(context.Background(), []PortabilityImportRequest{bad}); !errors.Is(err, ErrPortabilityResultInvalid) {
		t.Fatalf("malformed result = %v", err)
	}
	failing := &portabilityParticipantFake{id: "participant", allocate: true, fail: true}
	rollback := base
	rollback.Payload = []byte("payload")
	rollback.Participant = failing
	if _, _, err := PreparePortabilityImports(context.Background(), []PortabilityImportRequest{rollback}); !errors.Is(err, ErrPortabilityResultInvalid) || len(allocator.abandoned) != 1 {
		t.Fatalf("preparation rollback = %v/abandoned=%v", err, allocator.abandoned)
	}
	scope, err := NewStagedOutputScope("operation", "profile", allocator)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := scope.Allocate(context.Background(), "other-operation", nil); !errors.Is(err, ErrStagedOutputScope) {
		t.Fatalf("cross-operation staged access = %v", err)
	}
}

type transactionParticipantFake struct {
	id            string
	inputSize     int64
	result        []byte
	prepareErr    error
	validateErr   error
	prepareCalls  *[]string
	validateCalls *[]string
}

func (p transactionParticipantFake) ID() string       { return p.id }
func (p transactionParticipantFake) InputSize() int64 { return p.inputSize }
func (p transactionParticipantFake) Prepare(context.Context) (PreparedTransactionResult, error) {
	if p.prepareCalls != nil {
		*p.prepareCalls = append(*p.prepareCalls, p.id)
	}
	return PreparedTransactionResult{CanonicalBytes: p.result}, p.prepareErr
}
func (p transactionParticipantFake) Validate(context.Context, PreparedTransactionResult) error {
	if p.validateCalls != nil {
		*p.validateCalls = append(*p.validateCalls, p.id)
	}
	return p.validateErr
}

type sharedTransactionFake struct {
	writes     []string
	rolledBack bool
	outcome    TransactionCommitOutcome
	commitErr  error
}

func (t *sharedTransactionFake) Write(_ context.Context, id string, _ PreparedTransactionResult) error {
	t.writes = append(t.writes, id)
	return nil
}
func (t *sharedTransactionFake) Commit(context.Context) (TransactionCommitOutcome, error) {
	return t.outcome, t.commitErr
}
func (t *sharedTransactionFake) Rollback(context.Context) error {
	t.rolledBack = true
	return nil
}

type transactionBackendFake struct{ tx *sharedTransactionFake }

func (b transactionBackendFake) Begin(context.Context) (SharedTransaction, error) { return b.tx, nil }

func assertBC012ParticipantLimits(t *testing.T) {
	t.Helper()
	if !errors.Is(ValidateTransactionInputs(nil), ErrTransactionParticipants) {
		t.Fatal("zero participants accepted")
	}
	one := []TransactionParticipant{transactionParticipantFake{id: "00001", inputSize: TransactionByteLimit}}
	if err := ValidateTransactionInputs(one); err != nil {
		t.Fatalf("one exact-limit participant = %v", err)
	}
	one[0] = transactionParticipantFake{id: "00001", inputSize: TransactionByteLimit + 1}
	if !errors.Is(ValidateTransactionInputs(one), ErrTransactionInput) {
		t.Fatal("participant byte 67108865 accepted")
	}
	many := make([]TransactionParticipant, TransactionParticipantLimit)
	for index := range many {
		many[index] = transactionParticipantFake{id: fmt.Sprintf("%05d", index), inputSize: 0}
	}
	if err := ValidateTransactionInputs(many); err != nil {
		t.Fatalf("16384 participants = %v", err)
	}
	many = append(many, transactionParticipantFake{id: "99999"})
	if !errors.Is(ValidateTransactionInputs(many), ErrTransactionParticipants) {
		t.Fatal("16385 participants accepted")
	}
	prepareOrder, validateOrder := []string{}, []string{}
	participants := []TransactionParticipant{
		transactionParticipantFake{id: "a", result: []byte("a"), prepareCalls: &prepareOrder, validateCalls: &validateOrder},
		transactionParticipantFake{id: "b", result: []byte("b"), validateErr: errors.New("invalid"), prepareCalls: &prepareOrder, validateCalls: &validateOrder},
		transactionParticipantFake{id: "c", result: []byte("c"), prepareCalls: &prepareOrder, validateCalls: &validateOrder},
	}
	tx := &sharedTransactionFake{outcome: TransactionCommitProven}
	coordinator, _ := NewTransactionCoordinator(transactionBackendFake{tx: tx})
	if err := coordinator.Execute(context.Background(), participants); !errors.Is(err, ErrTransactionPrepare) || !reflect.DeepEqual(validateOrder, []string{"a", "b"}) || len(tx.writes) != 0 {
		t.Fatalf("first invalid = %v/order=%v/writes=%v", err, validateOrder, tx.writes)
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if err := coordinator.Execute(ctx, []TransactionParticipant{transactionParticipantFake{id: "a"}}); !errors.Is(err, ErrTransactionCancelled) {
		t.Fatalf("pre-step cancellation = %v", err)
	}
	tx = &sharedTransactionFake{outcome: TransactionCommitProven}
	coordinator, _ = NewTransactionCoordinator(transactionBackendFake{tx: tx})
	if err := coordinator.Execute(context.Background(), []TransactionParticipant{transactionParticipantFake{id: "a", result: []byte("ok")}}); err != nil || !reflect.DeepEqual(tx.writes, []string{"a"}) || tx.rolledBack {
		t.Fatalf("proven commit = %v/writes=%v/rollback=%t", err, tx.writes, tx.rolledBack)
	}
	tx = &sharedTransactionFake{outcome: TransactionCommitUnknown, commitErr: errors.New("connection lost")}
	coordinator, _ = NewTransactionCoordinator(transactionBackendFake{tx: tx})
	if err := coordinator.Execute(context.Background(), []TransactionParticipant{transactionParticipantFake{id: "a"}}); !errors.Is(err, ErrTransactionIndeterminate) || tx.rolledBack {
		t.Fatalf("indeterminate commit = %v/rollback=%t", err, tx.rolledBack)
	}
}

type stagedCleanupStoreFake struct {
	mu       sync.Mutex
	objects  []extensionstore.StagedObject
	success  []string
	failures []string
	prepare  int
	block    chan struct{}
}

func (s *stagedCleanupStoreFake) PrepareCleanupBatch(context.Context, time.Time, time.Time, int) ([]extensionstore.StagedObject, error) {
	s.mu.Lock()
	s.prepare++
	block := s.block
	objects := append([]extensionstore.StagedObject(nil), s.objects...)
	s.mu.Unlock()
	if block != nil {
		<-block
	}
	return objects, nil
}
func (s *stagedCleanupStoreFake) RecordDeletionSuccess(_ context.Context, id string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.success = append(s.success, id)
	return nil
}
func (s *stagedCleanupStoreFake) RecordDeletionFailure(_ context.Context, id, _ string, _ time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.failures = append(s.failures, id)
	return nil
}

type stagedDeleterFake struct {
	fail       map[string]bool
	identities []string
}

func (d *stagedDeleterFake) DeleteStagedObject(_ context.Context, identity string) error {
	d.identities = append(d.identities, identity)
	if d.fail[identity] {
		return errors.New("delete failed")
	}
	return nil
}

func assertBC013StagedObjectLifecycle(t *testing.T) {
	t.Helper()
	now := time.Date(2026, 7, 20, 0, 0, 0, 0, time.UTC)
	object := extensionstore.NewStagedObject("staging", "profile", "object/key", 1, "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", now)
	if object.State != extensionstore.StagedAllocated || object.DeleteState != extensionstore.DeleteNotApplicable || !object.StagingExpiresAt.Equal(now.Add(24*time.Hour)) || object.ReadyAt != nil || object.AbandonedAt != nil || object.DeleteAttemptCount != 0 {
		t.Fatalf("staged defaults = %#v", object)
	}
	if extensionstore.RetryDelay(1) != time.Minute || extensionstore.RetryDelay(12) != 24*time.Hour || extensionstore.RetryDelay(100) != 24*time.Hour {
		t.Fatal("retry delay is not ordered and saturated")
	}
	store := &stagedCleanupStoreFake{objects: []extensionstore.StagedObject{{StagingID: "a", StorageIdentity: "a-key"}, {StagingID: "b", StorageIdentity: "b-key"}}}
	deleter := &stagedDeleterFake{fail: map[string]bool{"b-key": true}}
	janitor, err := NewStagedObjectJanitor(store, deleter, nil, func() time.Time { return now }, 100)
	if err != nil || janitor.Sweep(context.Background()) != nil || !reflect.DeepEqual(store.success, []string{"a"}) || !reflect.DeepEqual(store.failures, []string{"b"}) || !reflect.DeepEqual(deleter.identities, []string{"a-key", "b-key"}) {
		t.Fatalf("janitor = %v/success=%v/fail=%v/delete=%v", err, store.success, store.failures, deleter.identities)
	}
	block := make(chan struct{})
	coalescedStore := &stagedCleanupStoreFake{block: block}
	coalesced, _ := NewStagedObjectJanitor(coalescedStore, &stagedDeleterFake{fail: map[string]bool{}}, nil, func() time.Time { return now }, 1)
	done := make(chan error, 1)
	go func() { done <- coalesced.Sweep(context.Background()) }()
	for {
		coalescedStore.mu.Lock()
		started := coalescedStore.prepare > 0
		coalescedStore.mu.Unlock()
		if started {
			break
		}
		time.Sleep(time.Millisecond)
	}
	_ = coalesced.Sweep(context.Background())
	_ = coalesced.Sweep(context.Background())
	close(block)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	coalescedStore.mu.Lock()
	passes := coalescedStore.prepare
	coalescedStore.mu.Unlock()
	if passes != 2 {
		t.Fatalf("overlapping sweeps produced %d passes; want 2", passes)
	}
}

type backupPortFake struct {
	events      []string
	failBinding string
}

func (p *backupPortFake) RestoreBinding(_ context.Context, binding BackupBinding, _ BackupCodec, _ RestoreBindingInput) error {
	p.events = append(p.events, "restore:"+binding.BindingID)
	if binding.BindingID == p.failBinding {
		return errors.New("restore failed")
	}
	return nil
}
func (p *backupPortFake) ValidateBinding(_ context.Context, binding BackupBinding) error {
	p.events = append(p.events, "validate:"+binding.BindingID)
	return nil
}

func assertBC014RestoreOrdering(t *testing.T) {
	t.Helper()
	coordinator, err := NewGeneratedCoordinator()
	if err != nil {
		t.Fatal(err)
	}
	plan, ok := coordinator.BackupPlan("network_flow_activity")
	if !ok || len(plan.Bindings) != 4 {
		t.Fatalf("backup plan = %#v/%t", plan, ok)
	}
	inputs := make([]RestoreBindingInput, 0, len(plan.Bindings))
	wantEvents := make([]string, 0, len(plan.Bindings)*2)
	groups := []int{100, 200, 300, 400}
	for index, binding := range plan.Bindings {
		if binding.RestoreOrderGroup != groups[index] {
			t.Fatalf("restore group %d = %d", index, binding.RestoreOrderGroup)
		}
		inputs = append(inputs, RestoreBindingInput{BindingID: binding.BindingID, CodecID: binding.BackupCodecID, CodecSHA256: binding.BackupCodecSHA256})
		wantEvents = append(wantEvents, "restore:"+binding.BindingID, "validate:"+binding.BindingID)
	}
	target := &RestoreTarget{Stopped: true, Empty: true}
	port := &backupPortFake{}
	if err := (BackupCoordinator{}).Restore(context.Background(), target, plan, inputs, port); err != nil || !reflect.DeepEqual(port.events, wantEvents) || !target.MayServe() {
		t.Fatalf("restore = %v/events=%v/serve=%t", err, port.events, target.MayServe())
	}
	for name, badTarget := range map[string]*RestoreTarget{"running": {Stopped: false, Empty: true}, "nonempty": {Stopped: true, Empty: false}} {
		t.Run(name, func(t *testing.T) {
			if err := (BackupCoordinator{}).Restore(context.Background(), badTarget, plan, inputs, &backupPortFake{}); err == nil {
				t.Fatal("invalid restore target accepted")
			}
		})
	}
	badInputs := append([]RestoreBindingInput(nil), inputs...)
	badInputs[0].CodecSHA256 = "unsupported"
	failed := &RestoreTarget{Stopped: true, Empty: true}
	if err := (BackupCoordinator{}).Restore(context.Background(), failed, plan, badInputs, &backupPortFake{}); !errors.Is(err, ErrBackupCodecUnsupported) || failed.MayServe() {
		t.Fatalf("unsupported codec = %v/serve=%t", err, failed.MayServe())
	}
}
