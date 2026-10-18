// Package harnessfixtures owns shared fixture lease and cleanup policy. Physical
// service/browser adapters retain acquisition and termination of their resources.
package harnessfixtures

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"reflect"
	"slices"
	"strings"
	"sync"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessschema"
)

type Failure struct {
	Class  string   `json:"failure_class"`
	Reason string   `json:"failure_reason"`
	Refs   []string `json:"artifact_refs,omitempty"`
	Cause  error    `json:"-"`
}

func (f *Failure) Error() string                    { return f.Class + "/" + f.Reason }
func (f *Failure) Unwrap() error                    { return f.Cause }
func (f *Failure) HarnessFailure() (string, string) { return f.Class, f.Reason }
func (f *Failure) ArtifactReferences() []string     { return slices.Clone(f.Refs) }
func fixtureFailure(err error) error {
	if err == nil {
		return nil
	}
	var f *Failure
	if errors.As(err, &f) {
		return err
	}
	return &Failure{Class: "harness", Reason: "fixture_error", Cause: err}
}
func artifactFailure(err error) error {
	return &Failure{Class: "artifact", Reason: "artifact_error", Cause: err}
}
func CleanupFailure(err error) error {
	if err == nil {
		return nil
	}
	var f *Failure
	if errors.As(err, &f) && (f.Reason == "cleanup_error" || f.Reason == "artifact_error" || f.Reason == "timeout_failure") {
		return err
	}
	if errors.Is(err, context.DeadlineExceeded) {
		return &Failure{Class: "timing", Reason: "timeout_failure", Cause: err}
	}
	return &Failure{Class: "harness", Reason: "cleanup_error", Cause: err}
}

type Request struct {
	Capability     string `json:"capability"`
	UnitID         string `json:"unit_id"`
	Affinity       string `json:"affinity_key,omitempty"`
	Digest         string `json:"digest,omitempty"`
	BrowserStage   string `json:"browser_stage,omitempty"`
	RuntimeProfile string `json:"runtime_profile_id,omitempty"`
	FixtureProfile string `json:"fixture_profile_id,omitempty"`
	SnapshotKey    string `json:"snapshot_key,omitempty"`
	BuilderUnit    string `json:"builder_unit_id,omitempty"`
	RowID          string `json:"row_id,omitempty"`
	PredicateID    string `json:"predicate_id,omitempty"`
	LeaseID        string `json:"-"`
	AllocationRef  string `json:"-"`
}
type Allocation struct {
	Ownership                                string
	ResourceIDs                              []string
	Environment                              map[string]string
	FixtureProfile, SnapshotKey, BuilderUnit string
	CloneOrdinal                             int
	Release                                  func(context.Context, bool) error
	Detach                                   func(context.Context) error
}
type Provider interface {
	Acquire(context.Context, Request) (*Allocation, error)
	Close(context.Context) error
}
type ProviderFuncs struct {
	AcquireFunc func(context.Context, Request) (*Allocation, error)
	CloseFunc   func(context.Context) error
}

func (p *ProviderFuncs) Acquire(c context.Context, r Request) (*Allocation, error) {
	return p.AcquireFunc(c, r)
}
func (p *ProviderFuncs) Close(c context.Context) error {
	if p.CloseFunc != nil {
		return p.CloseFunc(c)
	}
	return nil
}

type LeaseRecord struct {
	Schema         string   `json:"schema_id"`
	ID             string   `json:"lease_id"`
	Capability     string   `json:"capability"`
	Ownership      string   `json:"ownership"`
	State          string   `json:"state"`
	Cleanup        string   `json:"cleanup_outcome"`
	CleanupReason  *string  `json:"cleanup_failure_reason"`
	Resources      []string `json:"resource_ids"`
	Affinity       string   `json:"affinity_key,omitempty"`
	FixtureProfile string   `json:"fixture_profile_id,omitempty"`
	SnapshotKey    string   `json:"snapshot_key,omitempty"`
	BuilderUnit    string   `json:"builder_unit_id,omitempty"`
	CloneOrdinal   int      `json:"clone_ordinal,omitempty"`
	Created        string   `json:"created_at"`
}

func (r LeaseRecord) Copy() LeaseRecord {
	r.Resources = slices.Clone(r.Resources)
	if r.CleanupReason != nil {
		v := *r.CleanupReason
		r.CleanupReason = &v
	}
	return r
}

type CleanupRecord struct {
	Operation, UnitID, LeaseID string
	Err                        error
	Published                  bool
}
type Observation struct {
	Kind          string `json:"kind"`
	Ref           string `json:"ref"`
	AllocationRef string `json:"allocation_ref"`
	LeaseRef      string `json:"lease_ref"`
	UnitID        string `json:"unit_id"`
	Capability    string `json:"capability"`
	Ownership     string `json:"ownership"`
	Outcome       string `json:"outcome"`
}
type Options struct {
	Providers    map[string]Provider
	Clock        func() time.Time
	ID           func() string
	AllocationID func() string
	Publish      func(LeaseRecord) error
	Observe      func(Observation)
	Cleanup      func(CleanupRecord)
}
type entry struct {
	allocation *Allocation
	request    Request
	key        string
	refs       int
	tainted    bool
	last       *Lease
	settled    bool
	cleanOnce  sync.Once
	cleanErr   error
}
type Lease struct {
	broker          *Broker
	allocation      *entry
	request         Request
	record          LeaseRecord
	published       bool
	releaseOnce     sync.Once
	dispositionSet  bool
	healthy, retain bool
	retained        bool
	releaseErr      error
}

func (l *Lease) Record() LeaseRecord {
	l.broker.mu.Lock()
	defer l.broker.mu.Unlock()
	return l.record.Copy()
}
func (l *Lease) AllocationRef() string { return l.allocation.request.AllocationRef }
func (l *Lease) Environment() map[string]string {
	return maps.Clone(l.allocation.allocation.Environment)
}
func (l *Lease) Release(ctx context.Context, healthy, retain bool) (bool, error) {
	b := l.broker
	b.mu.Lock()
	if l.dispositionSet && (l.healthy != healthy || l.retain != retain) {
		b.mu.Unlock()
		return false, fixtureFailure(errors.New("contradictory release disposition"))
	}
	l.dispositionSet = true
	l.healthy = healthy
	l.retain = retain
	b.mu.Unlock()
	l.releaseOnce.Do(func() { l.retained, l.releaseErr = b.release(ctx, l, healthy, retain) })
	return l.retained, l.releaseErr
}

type Broker struct {
	mu             sync.Mutex
	options        Options
	shared         map[string]*entry
	pending        map[string]chan struct{}
	reserved       map[string]int
	allocations    []*entry
	active         map[*Lease]bool
	acquiring      sync.WaitGroup
	life           context.Context
	cancel         context.CancelFunc
	closed         bool
	closeOnce      sync.Once
	closeErr       error
	providerFailed bool
	sequence       uint64
}

func New(options Options) *Broker {
	if options.Clock == nil {
		options.Clock = time.Now
	}
	if options.Publish == nil {
		options.Publish = func(LeaseRecord) error { return nil }
	}
	life, cancel := context.WithCancel(context.Background())
	return &Broker{options: options, shared: map[string]*entry{}, pending: map[string]chan struct{}{}, reserved: map[string]int{}, active: map[*Lease]bool{}, life: life, cancel: cancel}
}

func sharedKey(request Request) string {
	if request.Capability == "browser_stack" {
		return request.Capability + ":" + request.Affinity + ":" + request.FixtureProfile + ":" + request.SnapshotKey
	}
	return ""
}

// Pin the stack identity while admission refers to its resident generation.
// This is logical ownership only, never permission to start physical work.
func (b *Broker) reserveStack(ctx context.Context, request Request) (func(), error) {
	key := sharedKey(request)
	if key == "" {
		return func() {}, nil
	}
	for {
		b.mu.Lock()
		if b.closed || ctx.Err() != nil {
			b.mu.Unlock()
			return nil, context.Canceled
		}
		if pending := b.pending[key]; pending != nil {
			b.mu.Unlock()
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-pending:
			}
			continue
		}
		b.reserved[key]++
		b.mu.Unlock()
		var once sync.Once
		return func() {
			once.Do(func() {
				b.mu.Lock()
				defer b.mu.Unlock()
				b.reserved[key]--
				if b.reserved[key] == 0 {
					delete(b.reserved, key)
				}
			})
		}, nil
	}
}
func validRecord(record LeaseRecord) error {
	bytes, err := json.Marshal(record)
	if err != nil {
		return err
	}
	v, err := harnessschema.Decode(bytes)
	if err != nil {
		return err
	}
	return harnessschema.Validate("cartulary.harness_fixture_lease.v4", v)
}
func (b *Broker) publish(l *Lease) error {
	b.mu.Lock()
	record := l.record.Copy()
	b.mu.Unlock()
	if err := validRecord(record); err != nil {
		return artifactFailure(err)
	}
	if (record.State == "released" || record.State == "destroyed") && record.Cleanup == "failed" {
		return artifactFailure(errors.New("settled lease has failed cleanup"))
	}
	if err := b.options.Publish(record); err != nil {
		return artifactFailure(err)
	}
	b.mu.Lock()
	l.published = true
	b.mu.Unlock()
	return nil
}
func (b *Broker) observe(o Observation) {
	if b.options.Observe == nil {
		return
	}
	defer func() { _ = recover() }()
	b.options.Observe(o)
}
func (b *Broker) cleanupRecord(r CleanupRecord) {
	if b.options.Cleanup != nil {
		b.options.Cleanup(r)
	}
}
func (b *Broker) Acquire(ctx context.Context, request Request) (lease *Lease, result error) {
	b.mu.Lock()
	if b.closed {
		b.mu.Unlock()
		return nil, fixtureFailure(errors.New("fixture owner is closed"))
	}
	b.acquiring.Add(1)
	b.sequence++
	sequence := b.sequence
	b.mu.Unlock()
	defer b.acquiring.Done()
	ctx, cancel := context.WithCancel(ctx)
	stop := context.AfterFunc(b.life, cancel)
	defer func() { stop(); cancel() }()
	request.LeaseID = fmt.Sprintf("lease-%06d", sequence)
	if b.options.ID != nil {
		request.LeaseID = b.options.ID()
	}
	probe := LeaseRecord{Schema: "cartulary.harness_fixture_lease.v4", ID: request.LeaseID, Capability: request.Capability, Ownership: "borrowed", State: "leased", Cleanup: "not_required", Resources: []string{}, Affinity: request.Affinity, Created: b.options.Clock().UTC().Format(time.RFC3339Nano)}
	if err := validRecord(probe); err != nil {
		return nil, fixtureFailure(err)
	}
	key := sharedKey(request)
	var e *entry
	creator := false
	for {
		if ctx.Err() != nil {
			return nil, ctx.Err()
		}
		b.mu.Lock()
		if b.closed {
			b.mu.Unlock()
			return nil, context.Canceled
		}
		if key != "" && b.pending[key] != nil {
			pending := b.pending[key]
			b.mu.Unlock()
			select {
			case <-ctx.Done():
				return nil, ctx.Err()
			case <-pending:
			}
			continue
		}
		e = b.shared[key]
		if key != "" {
			for _, prior := range b.allocations {
				if prior.key == key && prior.tainted && !prior.settled && prior.refs == 0 {
					b.mu.Unlock()
					return nil, CleanupFailure(ErrCleanupBlocked)
				}
			}
		}
		if e == nil {
			creator = true
			if key != "" {
				b.pending[key] = make(chan struct{})
			}
		} else {
			e.refs++
		}
		b.mu.Unlock()
		break
	}
	if creator && key != "" {
		defer func() { b.mu.Lock(); close(b.pending[key]); delete(b.pending, key); b.mu.Unlock() }()
	}
	if creator {
		request.AllocationRef = fmt.Sprintf("allocation:%d", sequence)
		if b.options.AllocationID != nil {
			request.AllocationRef = b.options.AllocationID()
		}
		notify := func(outcome, ownership string) {
			b.observe(Observation{Kind: "allocation", Ref: request.AllocationRef, AllocationRef: request.AllocationRef, UnitID: request.UnitID, Capability: request.Capability, Ownership: ownership, Outcome: outcome})
		}
		notify("incomplete", "")
		provider := b.options.Providers[request.Capability]
		var allocation *Allocation
		var err error
		if request.Capability == "none" {
			allocation = &Allocation{Ownership: "borrowed", ResourceIDs: []string{}}
		} else if provider == nil {
			err = fixtureFailure(errors.New("unknown fixture capability"))
		} else {
			allocation, err = provider.Acquire(ctx, request)
		}
		// Register returned ownership before any validation or retained publication.
		if allocation != nil {
			e = &entry{allocation: allocation, request: request, key: key}
			allocation.Environment = maps.Clone(allocation.Environment)
			allocation.ResourceIDs = append([]string{}, allocation.ResourceIDs...)
			slices.Sort(allocation.ResourceIDs)
			b.mu.Lock()
			b.allocations = append(b.allocations, e)
			b.mu.Unlock()
		}
		if err == nil && (allocation == nil || allocation.Ownership != "owned" && allocation.Ownership != "borrowed") {
			err = fixtureFailure(errors.New("invalid fixture ownership"))
		}
		if err == nil && request.Capability != "none" && (allocation.Ownership == "owned" && allocation.Release == nil || allocation.Ownership == "borrowed" && allocation.Detach == nil) {
			err = fixtureFailure(errors.New("missing ownership-specific cleanup"))
		}
		if err == nil && len(slices.Compact(slices.Clone(allocation.ResourceIDs))) != len(allocation.ResourceIDs) {
			err = fixtureFailure(errors.New("duplicate resource identity"))
		}
		if err != nil {
			notify("failed", "")
			if e != nil {
				b.mu.Lock()
				e.tainted = true
				b.mu.Unlock()
				cleanupCtx, done := context.WithTimeout(context.Background(), 2*time.Minute)
				cleanup := b.clean(cleanupCtx, e)
				done()
				err = errors.Join(err, cleanup)
			}
			return nil, err
		}
		notify("passed", allocation.Ownership)
	}
	a := e.allocation
	record := LeaseRecord{Schema: "cartulary.harness_fixture_lease.v4", ID: request.LeaseID, Capability: request.Capability, Ownership: a.Ownership, State: "leased", Cleanup: "not_required", Resources: slices.Clone(a.ResourceIDs), Affinity: request.Affinity, FixtureProfile: a.FixtureProfile, SnapshotKey: a.SnapshotKey, BuilderUnit: a.BuilderUnit, CloneOrdinal: a.CloneOrdinal, Created: b.options.Clock().UTC().Format("2006-01-02T15:04:05.000Z")}
	lease = &Lease{broker: b, allocation: e, request: request, record: record}
	b.mu.Lock()
	if creator {
		e.refs++
	}
	e.last = lease
	b.active[lease] = true
	if key != "" {
		b.shared[key] = e
	}
	b.mu.Unlock()
	err := b.publish(lease)
	if err == nil {
		b.observe(Observation{Kind: "lease", AllocationRef: e.request.AllocationRef, LeaseRef: record.ID, UnitID: request.UnitID, Capability: request.Capability, Ownership: a.Ownership, Outcome: "passed"})
		b.mu.Lock()
		closed := b.closed
		b.mu.Unlock()
		if closed || ctx.Err() != nil {
			err = context.Canceled
		}
	}
	if err != nil {
		cleanupCtx, done := context.WithTimeout(context.Background(), 2*time.Minute)
		_, cleanup := lease.Release(cleanupCtx, false, false)
		done()
		return nil, errors.Join(err, cleanup)
	}
	return lease, nil
}
func (b *Broker) unshare(e *entry) {
	if b.shared[e.key] == e {
		delete(b.shared, e.key)
	}
}
func (b *Broker) clean(ctx context.Context, e *entry) error {
	e.cleanOnce.Do(func() {
		b.mu.Lock()
		lease := e.last
		tainted := e.tainted
		if lease != nil && e.request.Capability != "none" {
			lease.record.State = "leased"
			if tainted {
				lease.record.State = "quarantined"
			}
			lease.record.Cleanup = "pending"
		}
		b.mu.Unlock()
		var errs []error
		if lease != nil && e.request.Capability != "none" {
			errs = append(errs, b.publish(lease))
		}
		var physical error
		if e.request.Capability != "none" {
			if e.allocation.Ownership == "borrowed" && e.allocation.Detach != nil {
				physical = e.allocation.Detach(ctx)
			} else if e.allocation.Ownership == "owned" && e.allocation.Release != nil {
				physical = e.allocation.Release(ctx, !tainted)
			} else {
				physical = errors.New("unresolved allocation cleanup hook")
			}
		}
		b.mu.Lock()
		e.settled = physical == nil
		e.tainted = e.tainted || physical != nil
		if lease != nil {
			lease.record.State = "released"
			if tainted {
				lease.record.State = "destroyed"
			}
			lease.record.Cleanup = "completed"
			if e.request.Capability == "none" {
				lease.record.Cleanup = "not_required"
			}
			if physical != nil {
				lease.record.State = "failed"
				if tainted {
					lease.record.State = "quarantined"
				}
				lease.record.Cleanup = "failed"
				reason := "cleanup_error"
				lease.record.CleanupReason = &reason
			}
		}
		b.mu.Unlock()
		errs = append(errs, CleanupFailure(physical))
		if lease != nil {
			errs = append(errs, b.publish(lease))
		}
		e.cleanErr = errors.Join(errs...)
		if e.request.Capability != "none" {
			operation := "fixture_release"
			if e.allocation.Ownership == "borrowed" {
				operation = "fixture_detach"
			}
			id, unit, published := e.request.LeaseID, e.request.UnitID, false
			if lease != nil {
				b.mu.Lock()
				id = lease.record.ID
				unit = lease.request.UnitID
				published = lease.published
				b.mu.Unlock()
			}
			b.cleanupRecord(CleanupRecord{Operation: operation, UnitID: unit, LeaseID: id, Err: e.cleanErr, Published: published})
		}
	})
	return e.cleanErr
}
func (b *Broker) release(ctx context.Context, l *Lease, healthy, retain bool) (retained bool, result error) {
	e := l.allocation
	b.mu.Lock()
	e.refs--
	e.tainted = e.tainted || !healthy
	if e.tainted {
		b.unshare(e)
	}
	retained = e.refs == 0 && !e.tainted && retain && e.request.Capability == "browser_stack" && !b.closed
	clean := e.refs == 0 && !retained
	if clean {
		b.unshare(e)
	} else {
		l.record.State = "released"
		if e.tainted {
			l.record.State = "quarantined"
		}
		l.record.Cleanup = "not_required"
	}
	b.mu.Unlock()
	if clean {
		result = b.clean(ctx, e)
	} else {
		result = b.publish(l)
	}
	if result != nil {
		b.mu.Lock()
		e.tainted = true
		b.unshare(e)
		unreferenced := e.refs == 0
		b.mu.Unlock()
		if unreferenced {
			result = errors.Join(result, b.clean(ctx, e))
		}
		retained = false
	}
	b.mu.Lock()
	delete(b.active, l)
	b.mu.Unlock()
	return
}
func (b *Broker) Unresolved() bool {
	b.mu.Lock()
	defer b.mu.Unlock()
	if b.providerFailed {
		return true
	}
	for _, e := range b.allocations {
		if e.request.Capability != "none" && !e.settled {
			return true
		}
	}
	return false
}

// RetireIdle closes ordinary, unborrowed stacks before an exclusive allocation.
// Future consumers acquire a fresh database or their sealed fixture snapshot;
// existing reset operations and group boundaries still execute unchanged.
func (b *Broker) RetireIdle(ctx context.Context) error {
	_, err := b.retireIdle(ctx, 0, nil, b.clean)
	return err
}

// Pressure retirement uses the same settlement path as exclusive retirement,
// selecting at most one eligible allocation before admission is evaluated again.
func (b *Broker) retireIdle(ctx context.Context, maximum int, needed func(Request) (bool, error), cleanup func(context.Context, *entry) error) (int, error) {
	b.mu.Lock()
	entries := []*entry{}
	for _, e := range b.allocations {
		if e.request.Capability == "browser_stack" && e.refs == 0 && b.reserved[e.key] == 0 && !e.settled && b.shared[e.key] == e {
			entries = append(entries, e)
		}
	}
	b.mu.Unlock()
	slices.SortFunc(entries, func(a, b *entry) int { return strings.Compare(a.request.AllocationRef, b.request.AllocationRef) })
	var result error
	retired := 0
	for _, e := range entries {
		if needed != nil {
			match, err := needed(e.request)
			if err != nil {
				return retired, errors.Join(result, err)
			}
			if !match {
				continue
			}
		}
		b.mu.Lock()
		// Selection runs outside the broker lock. Recheck the borrower boundary
		// before revoking this allocation and blocking a concurrent acquisition.
		if e.refs != 0 || b.reserved[e.key] != 0 || e.settled || b.shared[e.key] != e {
			b.mu.Unlock()
			continue
		}
		b.unshare(e)
		b.pending[e.key] = make(chan struct{})
		b.mu.Unlock()
		result = errors.Join(result, cleanup(ctx, e))
		b.mu.Lock()
		close(b.pending[e.key])
		delete(b.pending, e.key)
		b.mu.Unlock()
		retired++
		if maximum > 0 && retired >= maximum {
			break
		}
	}
	return retired, result
}

func (b *Broker) Close(ctx context.Context) error {
	b.closeOnce.Do(func() {
		b.mu.Lock()
		b.closed = true
		b.cancel()
		b.mu.Unlock()
		b.acquiring.Wait()
		b.mu.Lock()
		entries := slices.Clone(b.allocations)
		active := maps.Clone(b.active)
		b.mu.Unlock()
		errs := []error{}
		for i := len(entries) - 1; i >= 0; i-- {
			e := entries[i]
			for l := range active {
				if l.allocation != e {
					continue
				}
				b.mu.Lock()
				set, healthy, retain := l.dispositionSet, l.healthy, l.retain
				b.mu.Unlock()
				if !set {
					healthy, retain = true, false
				}
				_, err := l.Release(ctx, healthy, retain)
				errs = append(errs, err)
			}
			b.mu.Lock()
			b.unshare(e)
			b.mu.Unlock()
			errs = append(errs, b.clean(ctx, e))
		}
		capabilities := slices.Sorted(maps.Keys(b.options.Providers))
		seen := map[Provider]bool{}
		for _, capability := range capabilities {
			provider := b.options.Providers[capability]
			if provider == nil {
				continue
			}
			if reflect.ValueOf(provider).Comparable() {
				if seen[provider] {
					continue
				}
				seen[provider] = true
			}
			err := CleanupFailure(provider.Close(ctx))
			if err != nil {
				b.mu.Lock()
				b.providerFailed = true
				b.mu.Unlock()
				errs = append(errs, err)
			}
			b.cleanupRecord(CleanupRecord{Operation: "provider_close", Err: err})
		}
		b.closeErr = errors.Join(errs...)
	})
	return b.closeErr
}
