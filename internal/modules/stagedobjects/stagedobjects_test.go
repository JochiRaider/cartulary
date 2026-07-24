package stagedobjects

import (
	"context"
	"errors"
	"reflect"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestService_Unit_AllocationReadyAbandonAndFatalOutcomes(t *testing.T) {
	now := time.Date(2026, 7, 24, 15, 0, 0, 0, time.UTC)
	t.Run("allocates immutable bytes and marks the record ready", func(t *testing.T) {
		repository := &serviceRepositoryFake{}
		bytes := &byteStoreFake{putOutcome: ByteWriteSuccess}
		service := newTestService(t, repository, bytes, now, func() (string, error) { return "staging-1", nil }, nil)
		reference, err := service.Allocate(context.Background(), "operation-1", "network_flow_activity", []byte("payload"))
		if err != nil {
			t.Fatal(err)
		}
		if reference.StagingID != "staging-1" ||
			len(repository.allocations) != 1 ||
			!reflect.DeepEqual(repository.ready, []string{"staging-1"}) ||
			len(repository.abandoned) != 0 ||
			bytes.putIdentity != ".cartulary/staged/network_flow_activity/staging-1" ||
			string(bytes.putPayload) != "payload" {
			t.Fatalf("allocation = %#v repo=%#v bytes=%#v", reference, repository, bytes)
		}
		allocation := repository.allocations[0]
		if allocation.OperationID != "operation-1" ||
			allocation.ByteSize != 7 ||
			allocation.SHA256 != "239f59ed55e737c77147cf55ad0c1b030b6d7ee748a7426952f9b852d5a935e5" ||
			!allocation.ExpiresAt.Equal(now.Add(StagingLifetime)) {
			t.Fatalf("allocation contract = %#v", allocation)
		}
	})

	t.Run("upload uncertainty abandons before returning", func(t *testing.T) {
		repository := &serviceRepositoryFake{}
		bytes := &byteStoreFake{putOutcome: ByteWriteIndeterminate, putErr: errors.New("private backend detail")}
		service := newTestService(t, repository, bytes, now, func() (string, error) { return "staging-2", nil }, nil)
		if _, err := service.Allocate(context.Background(), "operation-2", "network_flow_activity", []byte("payload")); FailureKind(err) != FailureRetryable {
			t.Fatalf("indeterminate upload error = %v", err)
		}
		if !reflect.DeepEqual(repository.abandoned, []string{"staging-2"}) {
			t.Fatalf("abandoned = %v", repository.abandoned)
		}
	})

	t.Run("failed durable abandon is fatal", func(t *testing.T) {
		repository := &serviceRepositoryFake{abandonErr: NewFailure(FailureDependency, "postgres_unavailable", nil)}
		bytes := &byteStoreFake{putOutcome: ByteWriteDependency, putErr: errors.New("private backend detail")}
		var fatalCalls atomic.Int32
		service := newTestService(t, repository, bytes, now, func() (string, error) { return "staging-3", nil }, func(error) {
			fatalCalls.Add(1)
		})
		_, err := service.Allocate(context.Background(), "operation-3", "network_flow_activity", []byte("payload"))
		if !IsFatalIntegrity(err) || fatalCalls.Load() != 1 {
			t.Fatalf("fatal abandon = %v calls=%d", err, fatalCalls.Load())
		}
	})
}

func TestScope_Unit_OperationTransferAndFinalPublication(t *testing.T) {
	allocator := &allocatorFake{}
	scope, err := NewScope("operation", "network_flow_activity", allocator)
	if err != nil {
		t.Fatal(err)
	}
	first, err := scope.Allocate(context.Background(), "operation", []byte("one"))
	if err != nil {
		t.Fatal(err)
	}
	second, err := scope.Allocate(context.Background(), "operation", []byte("two"))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := scope.Allocate(context.Background(), "other", nil); !errors.Is(err, ErrScopeDenied) {
		t.Fatalf("cross-operation allocation = %v", err)
	}
	transfer, err := scope.Transfer("operation")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := scope.Allocate(context.Background(), "operation", nil); !errors.Is(err, ErrScopeDenied) {
		t.Fatalf("closed scope allocation = %v", err)
	}
	capability := &publicationCapabilityFake{operationID: "operation"}
	publications := []Publication{
		{StagingID: first, ResourceKind: "incident_bundle", ResourceID: "bundle-1", ByteSize: 3, SHA256: testStagedDigest("a")},
		{StagingID: second, ResourceKind: "incident_bundle", ResourceID: "bundle-2", ByteSize: 3, SHA256: testStagedDigest("b")},
	}
	if err := PublishTransferred(context.Background(), transfer, publications, capability); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(capability.published, publications) {
		t.Fatalf("published = %#v", capability.published)
	}
	if err := PublishTransferred(context.Background(), transfer, publications, &publicationCapabilityFake{operationID: "other"}); !errors.Is(err, ErrScopeDenied) {
		t.Fatalf("wrong operation publication = %v", err)
	}
}

func TestJanitor_Unit_CapturedCutoffDrainsBoundedBatchesWithoutDeleteLock(t *testing.T) {
	now := time.Date(2026, 7, 24, 16, 0, 0, 0, time.UTC)
	repository := &janitorRepositoryFake{
		pages: [][]CleanupCandidate{
			{{StagingID: "a", StorageIdentity: "key-a"}, {StagingID: "b", StorageIdentity: "key-b"}},
			{{StagingID: "c", StorageIdentity: "key-c"}},
			{},
		},
	}
	bytes := &byteStoreFake{
		deleteOutcomes: map[string]DeleteOutcome{
			"key-a": DeleteSuccess,
			"key-b": DeleteAbsent,
			"key-c": DeleteSuccess,
		},
		beforeDelete: func() {
			if repository.preparing.Load() {
				t.Fatal("physical deletion ran while cleanup preparation transaction was open")
			}
		},
	}
	janitor := newTestJanitor(t, repository, bytes, func() time.Time { return now }, nil, nil, nil)
	if err := janitor.Sweep(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !reflect.DeepEqual(repository.successes, []string{"a", "b", "c"}) ||
		!reflect.DeepEqual(bytes.deleted, []string{"key-a", "key-b", "key-c"}) ||
		repository.prepareCalls != 3 {
		t.Fatalf("drain success=%v deleted=%v prepares=%d", repository.successes, bytes.deleted, repository.prepareCalls)
	}
	for _, cutoff := range repository.cutoffs {
		if !cutoff.Equal(now) {
			t.Fatalf("cutoff changed within drain: %v", repository.cutoffs)
		}
	}
}

func TestJanitor_Unit_SerializesAndCoalescesExactlyOneFollowUp(t *testing.T) {
	entered := make(chan struct{})
	release := make(chan struct{})
	repository := &janitorRepositoryFake{
		pages:        [][]CleanupCandidate{{}, {}},
		firstEntered: entered,
		releaseFirst: release,
	}
	janitor := newTestJanitor(t, repository, &byteStoreFake{}, time.Now, nil, nil, nil)
	done := make(chan error, 1)
	go func() { done <- janitor.Sweep(context.Background()) }()
	<-entered
	for range 3 {
		if err := janitor.Sweep(context.Background()); err != nil {
			t.Fatal(err)
		}
	}
	close(release)
	if err := <-done; err != nil {
		t.Fatal(err)
	}
	if repository.prepareCalls != 2 {
		t.Fatalf("prepare calls = %d, want exactly initial plus one follow-up", repository.prepareCalls)
	}
}

func TestJanitor_Unit_RetryScheduleSaturatesAndDependencyRecovers(t *testing.T) {
	now := time.Date(2026, 7, 24, 17, 0, 0, 0, time.UTC)
	repository := &janitorRepositoryFake{
		pages: [][]CleanupCandidate{
			{{StagingID: "retry", StorageIdentity: "key-retry", DeleteAttemptCount: 0}},
			{},
			{{StagingID: "retry", StorageIdentity: "key-retry", DeleteAttemptCount: 1}},
			{},
		},
	}
	bytes := &byteStoreFake{
		deleteSequence: []DeleteOutcome{DeleteDependency, DeleteSuccess},
		deleteErr:      errors.New("private dependency detail"),
	}
	health := NewHealth()
	janitor := newTestJanitor(t, repository, bytes, func() time.Time { return now }, health, nil, nil)
	if err := janitor.Sweep(context.Background()); !errors.Is(err, ErrDependency) {
		t.Fatalf("dependency sweep = %v", err)
	}
	state := health.State()
	if state.Available || state.ReasonCode != "object_store_unavailable" {
		t.Fatalf("degraded health = %#v", state)
	}
	if len(repository.failures) != 1 ||
		repository.failures[0].AttemptCount != 1 ||
		!repository.failures[0].NextAttemptAt.Equal(now.Add(time.Minute)) {
		t.Fatalf("retry = %#v", repository.failures)
	}
	bytes.deleteErr = nil
	if err := janitor.Sweep(context.Background()); err != nil {
		t.Fatalf("recovery sweep = %v", err)
	}
	if state := health.State(); !state.Available {
		t.Fatalf("health did not recover: %#v", state)
	}
	if RetryDelay(1) != time.Minute ||
		RetryDelay(11) != 1024*time.Minute ||
		RetryDelay(12) != 24*time.Hour ||
		RetryDelay(100) != 24*time.Hour ||
		SaturatingAttempt(1<<31-1) != 1<<31-1 {
		t.Fatal("retry policy is not exact and saturating")
	}
}

func TestJanitor_Unit_FatalContradictionAndCleanupTimeout(t *testing.T) {
	t.Run("integrity outcome is fatal", func(t *testing.T) {
		repository := &janitorRepositoryFake{
			pages: [][]CleanupCandidate{{{StagingID: "fatal", StorageIdentity: "key-fatal"}}},
		}
		bytes := &byteStoreFake{deleteOutcomes: map[string]DeleteOutcome{"key-fatal": DeleteIntegrity}}
		var fatalCalls atomic.Int32
		janitor := newTestJanitor(t, repository, bytes, time.Now, nil, func(error) { fatalCalls.Add(1) }, nil)
		err := janitor.Sweep(context.Background())
		if !IsFatalIntegrity(err) || fatalCalls.Load() != 1 {
			t.Fatalf("fatal contradiction = %v calls=%d", err, fatalCalls.Load())
		}
	})

	t.Run("monotonic timeout prevents late deletion", func(t *testing.T) {
		var monotonic atomic.Int64
		repository := &janitorRepositoryFake{
			pages: [][]CleanupCandidate{{{StagingID: "late", StorageIdentity: "key-late"}}},
			afterPrepare: func() {
				monotonic.Store(int64(2 * time.Second))
			},
		}
		bytes := &byteStoreFake{}
		janitor := newTestJanitor(t, repository, bytes, time.Now, nil, nil, monotonic.Load)
		if err := janitor.Sweep(context.Background()); !errors.Is(err, ErrCleanupTimeout) {
			t.Fatalf("cleanup timeout = %v", err)
		}
		if len(bytes.deleted) != 0 {
			t.Fatalf("late deletion ran: %v", bytes.deleted)
		}
	})
}

func TestJanitor_Unit_PeriodicErrorsAreSurfacedWithoutStoppingRecovery(t *testing.T) {
	repository := &janitorRepositoryFake{
		pages: [][]CleanupCandidate{
			{{StagingID: "dependency", StorageIdentity: "key-dependency"}},
			{},
		},
	}
	bytes := &byteStoreFake{
		deleteOutcomes: map[string]DeleteOutcome{"key-dependency": DeleteDependency},
		deleteErr:      errors.New("private dependency detail"),
	}
	var surfaced atomic.Int32
	janitor, err := NewJanitor(JanitorOptions{
		Repository:       repository,
		Bytes:            bytes,
		BatchLimit:       1,
		OperationTimeout: time.Second,
		FatalSink:        func(err error) { t.Fatalf("unexpected fatal: %v", err) },
		ErrorSink:        func(error) { surfaced.Add(1) },
	})
	if err != nil {
		t.Fatal(err)
	}
	if err := janitor.runScheduledSweep(context.Background()); err != nil {
		t.Fatalf("nonfatal scheduled dependency stopped the janitor: %v", err)
	}
	if surfaced.Load() != 1 {
		t.Fatalf("surfaced errors = %d, want 1", surfaced.Load())
	}
}

type serviceRepositoryFake struct {
	allocations []Allocation
	ready       []string
	abandoned   []string
	allocateErr error
	readyErr    error
	abandonErr  error
}

func (r *serviceRepositoryFake) Allocate(_ context.Context, allocation Allocation) error {
	r.allocations = append(r.allocations, allocation)
	return r.allocateErr
}

func (r *serviceRepositoryFake) MarkReady(_ context.Context, stagingID string, _ time.Time) error {
	r.ready = append(r.ready, stagingID)
	return r.readyErr
}

func (r *serviceRepositoryFake) Abandon(_ context.Context, stagingID string, _ time.Time) error {
	r.abandoned = append(r.abandoned, stagingID)
	return r.abandonErr
}

func (*serviceRepositoryFake) PrepareCleanupBatch(context.Context, time.Time, time.Time, int) ([]CleanupCandidate, error) {
	return nil, nil
}

func (*serviceRepositoryFake) RecordDeletionSuccess(context.Context, string) error {
	return nil
}

func (*serviceRepositoryFake) RecordDeletionFailure(context.Context, DeletionFailure) error {
	return nil
}

type byteStoreFake struct {
	putOutcome     ByteWriteOutcome
	putErr         error
	putIdentity    string
	putPayload     []byte
	deleteOutcomes map[string]DeleteOutcome
	deleteSequence []DeleteOutcome
	deleteErr      error
	deleted        []string
	beforeDelete   func()
}

func (s *byteStoreFake) Put(_ context.Context, identity string, payload []byte) (ByteWriteOutcome, error) {
	s.putIdentity = identity
	s.putPayload = append([]byte(nil), payload...)
	return s.putOutcome, s.putErr
}

func (s *byteStoreFake) Delete(_ context.Context, identity string) (DeleteOutcome, error) {
	if s.beforeDelete != nil {
		s.beforeDelete()
	}
	s.deleted = append(s.deleted, identity)
	if len(s.deleteSequence) > 0 {
		outcome := s.deleteSequence[0]
		s.deleteSequence = append([]DeleteOutcome(nil), s.deleteSequence[1:]...)
		return outcome, s.deleteErr
	}
	if outcome := s.deleteOutcomes[identity]; outcome != "" {
		return outcome, s.deleteErr
	}
	return DeleteSuccess, s.deleteErr
}

type allocatorFake struct {
	next      int
	abandoned []Reference
}

func (a *allocatorFake) Allocate(context.Context, string, string, []byte) (Reference, error) {
	a.next++
	return Reference{StagingID: "staging-" + string(rune('0'+a.next))}, nil
}

func (a *allocatorFake) Abandon(_ context.Context, reference Reference) error {
	a.abandoned = append(a.abandoned, reference)
	return nil
}

type publicationCapabilityFake struct {
	operationID string
	published   []Publication
}

func (c *publicationCapabilityFake) OperationID() string {
	return c.operationID
}

func (c *publicationCapabilityFake) PublishStagedObject(_ context.Context, publication Publication) error {
	c.published = append(c.published, publication)
	return nil
}

type janitorRepositoryFake struct {
	mu           sync.Mutex
	pages        [][]CleanupCandidate
	prepareCalls int
	cutoffs      []time.Time
	successes    []string
	failures     []DeletionFailure
	preparing    atomic.Bool
	firstEntered chan struct{}
	releaseFirst chan struct{}
	afterPrepare func()
}

func (r *janitorRepositoryFake) Allocate(context.Context, Allocation) error {
	return nil
}

func (r *janitorRepositoryFake) MarkReady(context.Context, string, time.Time) error {
	return nil
}

func (r *janitorRepositoryFake) Abandon(context.Context, string, time.Time) error {
	return nil
}

func (r *janitorRepositoryFake) PrepareCleanupBatch(_ context.Context, cutoff, _ time.Time, _ int) ([]CleanupCandidate, error) {
	r.preparing.Store(true)
	r.mu.Lock()
	r.prepareCalls++
	call := r.prepareCalls
	r.cutoffs = append(r.cutoffs, cutoff)
	var page []CleanupCandidate
	if len(r.pages) > 0 {
		page = append([]CleanupCandidate(nil), r.pages[0]...)
		r.pages = append([][]CleanupCandidate(nil), r.pages[1:]...)
	}
	entered := r.firstEntered
	release := r.releaseFirst
	after := r.afterPrepare
	r.mu.Unlock()
	if call == 1 && entered != nil {
		close(entered)
		<-release
	}
	r.preparing.Store(false)
	if after != nil {
		after()
	}
	return page, nil
}

func (r *janitorRepositoryFake) RecordDeletionSuccess(_ context.Context, stagingID string) error {
	r.mu.Lock()
	r.successes = append(r.successes, stagingID)
	r.mu.Unlock()
	return nil
}

func (r *janitorRepositoryFake) RecordDeletionFailure(_ context.Context, failure DeletionFailure) error {
	r.mu.Lock()
	r.failures = append(r.failures, failure)
	r.mu.Unlock()
	return nil
}

func newTestService(t testing.TB, repository Repository, bytes ByteStore, now time.Time, newID IDGenerator, fatalSink func(error)) *Service {
	t.Helper()
	if fatalSink == nil {
		fatalSink = func(err error) { t.Fatalf("unexpected fatal: %v", err) }
	}
	service, err := NewService(ServiceOptions{
		Repository: repository,
		Bytes:      bytes,
		Now:        func() time.Time { return now },
		NewID:      newID,
		FatalSink:  fatalSink,
	})
	if err != nil {
		t.Fatal(err)
	}
	return service
}

func newTestJanitor(t testing.TB, repository Repository, bytes ByteStore, now func() time.Time, health *Health, fatalSink func(error), monotonic func() int64) *Janitor {
	t.Helper()
	if fatalSink == nil {
		fatalSink = func(err error) { t.Fatalf("unexpected fatal: %v", err) }
	}
	janitor, err := NewJanitor(JanitorOptions{
		Repository:       repository,
		Bytes:            bytes,
		Health:           health,
		Now:              now,
		MonotonicNowNS:   monotonic,
		BatchLimit:       2,
		OperationTimeout: time.Second,
		FatalSink:        fatalSink,
	})
	if err != nil {
		t.Fatal(err)
	}
	return janitor
}

func testStagedDigest(character string) string {
	value := make([]byte, 64)
	for index := range value {
		value[index] = character[0]
	}
	return string(value)
}
