package harnessfixtures

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func ownedAllocation(release func(context.Context, bool) error) *Allocation {
	return &Allocation{Ownership: "owned", ResourceIDs: []string{"fixture:one"}, Environment: map[string]string{"FIXTURE": "value"}, Release: release}
}
func testRequest() Request {
	return Request{PostgresIsolation: "exclusive", Services: []string{"object_store", "postgres"}, Capability: "browser_stack", UnitID: "unit:a", Affinity: "browser"}
}
func TestFixtureBrokerOwnership(t *testing.T) {
	testConsumingBrowserPreparation(t)
	testAllocationLifecycleObservations(t)
	t.Run("pressure_retires_one_eligible_idle_stack_and_reacquires_fresh_state", func(t *testing.T) {
		released := []string{}
		b := New(Options{Providers: map[string]Provider{"browser_stack": &ProviderFuncs{AcquireFunc: func(_ context.Context, request Request) (*Allocation, error) {
			return ownedAllocation(func(context.Context, bool) error { released = append(released, request.Affinity); return nil }), nil
		}}}})
		leases := map[string]*Lease{}
		for _, name := range []string{"active", "idle-a", "idle-b", "unrelated"} {
			request := testRequest()
			request.Affinity, request.UnitID = name, name
			lease, err := b.Acquire(t.Context(), request)
			if err != nil {
				t.Fatal(err)
			}
			leases[name] = lease
			if name != "active" {
				if _, err := lease.Release(t.Context(), true, true); err != nil {
					t.Fatal(err)
				}
			}
		}
		reservation := testRequest()
		reservation.Affinity = "idle-a"
		unpin, err := b.reserveStack(t.Context(), reservation)
		if err != nil {
			t.Fatal(err)
		}
		defer unpin()
		needed := func(request Request) (bool, error) { return request.Affinity != "unrelated", nil }
		count, err := b.retireIdle(t.Context(), 1, needed, func(ctx context.Context, e *entry) error { return b.clean(ctx, e, Retirement{"capacity_pressure", ""}) })
		if err != nil || count != 1 || !slices.Equal(released, []string{"idle-b"}) {
			t.Fatal("pressure retirement crossed its borrower or selection boundary", count, released, err)
		}
		request := testRequest()
		request.Affinity, request.UnitID = released[0], "replacement"
		replacement, err := b.Acquire(t.Context(), request)
		if err != nil || replacement.AllocationRef() == leases[released[0]].AllocationRef() {
			t.Fatal("retired state reused", err)
		}
		if _, err := b.retireIdle(t.Context(), 1, func(Request) (bool, error) { return false, nil }, func(ctx context.Context, e *entry) error { return b.clean(ctx, e, Retirement{"capacity_pressure", ""}) }); err != nil || len(released) != 1 {
			t.Fatal("unrelated pressure retired a stack", err)
		}
		if err := b.Close(t.Context()); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("idle_retirement_preserves_borrowers_and_reacquires_fresh_state", func(t *testing.T) {
		acquired, released := 0, 0
		b := New(Options{Providers: map[string]Provider{"browser_stack": &ProviderFuncs{AcquireFunc: func(context.Context, Request) (*Allocation, error) {
			acquired++
			return ownedAllocation(func(context.Context, bool) error { released++; return nil }), nil
		}}}})
		first, err := b.Acquire(t.Context(), testRequest())
		if err != nil {
			t.Fatal(err)
		}
		if err := b.RetireIdle(t.Context()); err != nil || released != 0 {
			t.Fatal("retired active borrower", err, released)
		}
		if _, err := first.Release(t.Context(), true, true); err != nil {
			t.Fatal(err)
		}
		if err := b.RetireIdle(t.Context()); err != nil || released != 1 {
			t.Fatal("idle retirement", err, released)
		}
		second, err := b.Acquire(t.Context(), testRequest())
		if err != nil || acquired != 2 || first.AllocationRef() == second.AllocationRef() {
			t.Fatal("fresh allocation required", err, acquired)
		}
		if _, err := second.Release(t.Context(), true, false); err != nil {
			t.Fatal(err)
		}
		if err := b.Close(t.Context()); err != nil || released != 2 {
			t.Fatal(err, released)
		}
	})
	for _, borrowed := range []bool{false, true} {
		name := "owned"
		if borrowed {
			name = "borrowed"
		}
		t.Run(name, func(t *testing.T) {
			release, detach, close := 0, 0, 0
			provider := &ProviderFuncs{AcquireFunc: func(context.Context, Request) (*Allocation, error) {
				a := ownedAllocation(func(context.Context, bool) error { release++; return nil })
				a.Detach = func(context.Context) error { detach++; return nil }
				if borrowed {
					a.Ownership = "borrowed"
				}
				return a, nil
			}, CloseFunc: func(context.Context) error { close++; return nil }}
			records := []LeaseRecord{}
			b := New(Options{Providers: map[string]Provider{"browser_stack": provider}, Publish: func(r LeaseRecord) error { records = append(records, r); return nil }})
			l, err := b.Acquire(context.Background(), testRequest())
			if err != nil {
				t.Fatal(err)
			}
			environment := l.Environment()
			environment["FIXTURE"] = "changed"
			if l.Environment()["FIXTURE"] != "value" {
				t.Fatal("mutable resource escaped")
			}
			if _, err := l.Release(context.Background(), true, false); err != nil {
				t.Fatal(err)
			}
			if _, err := l.Release(context.Background(), true, false); err != nil {
				t.Fatal(err)
			}
			if _, err := l.Release(context.Background(), false, false); err == nil {
				t.Fatal("contradictory release accepted")
			}
			if err := b.Close(context.Background()); err != nil {
				t.Fatal(err)
			}
			if err := b.Close(context.Background()); err != nil {
				t.Fatal(err)
			}
			if borrowed && (release != 0 || detach != 1) || !borrowed && (release != 1 || detach != 0) || close != 1 || b.Unresolved() {
				t.Fatal(release, detach, close, b.Unresolved())
			}
			if len(records) != 3 || records[0].State != "leased" || records[1].Cleanup != "pending" || records[2].Cleanup != "completed" || records[2].State != "released" {
				t.Fatal(records)
			}
		})
	}
	t.Run("warm_reuse_and_unhealthy_replacement", func(t *testing.T) {
		acquired, released := 0, 0
		b := New(Options{Providers: map[string]Provider{"browser_stack": &ProviderFuncs{AcquireFunc: func(context.Context, Request) (*Allocation, error) {
			acquired++
			return ownedAllocation(func(context.Context, bool) error { released++; return nil }), nil
		}}}})
		a, err := b.Acquire(context.Background(), testRequest())
		if err != nil {
			t.Fatal(err)
		}
		if retained, err := a.Release(context.Background(), true, true); err != nil || !retained {
			t.Fatal(retained, err)
		}
		next := testRequest()
		next.UnitID = "unit:b"
		l, err := b.Acquire(context.Background(), next)
		if err != nil {
			t.Fatal(err)
		}
		if a.Record().ID == l.Record().ID || acquired != 1 {
			t.Fatal("reuse lost lease identity", acquired)
		}
		if retained, err := l.Release(context.Background(), false, true); err != nil || retained || released != 1 {
			t.Fatal(retained, err, released)
		}
		if l.Record().State != "destroyed" {
			t.Fatal(l.Record())
		}
		_, err = b.Acquire(context.Background(), testRequest())
		if err != nil || acquired != 2 {
			t.Fatal(acquired, err)
		}
		if err := b.Close(context.Background()); err != nil || released != 2 {
			t.Fatal(released, err)
		}
	})
}
func TestFixtureBrokerConcurrency(t *testing.T) {
	t.Run("one_shared_allocation_during_overlapping_acquisition", func(t *testing.T) {
		var acquired, released atomic.Int64
		start, unblock := make(chan struct{}), make(chan struct{})
		b := New(Options{Providers: map[string]Provider{"browser_stack": &ProviderFuncs{AcquireFunc: func(context.Context, Request) (*Allocation, error) {
			if acquired.Add(1) == 1 {
				close(start)
			}
			<-unblock
			return ownedAllocation(func(context.Context, bool) error { released.Add(1); return nil }), nil
		}}}})
		leases := make(chan *Lease, 2)
		errs := make(chan error, 2)
		for range 2 {
			go func() { l, err := b.Acquire(context.Background(), testRequest()); leases <- l; errs <- err }()
		}
		<-start
		close(unblock)
		a, c := <-leases, <-leases
		for range 2 {
			if err := <-errs; err != nil {
				t.Fatal(err)
			}
		}
		if acquired.Load() != 1 {
			t.Fatal(acquired.Load())
		}
		if _, err := a.Release(context.Background(), true, false); err != nil {
			t.Fatal(err)
		}
		if released.Load() != 0 {
			t.Fatal("released shared active fixture")
		}
		if _, err := c.Release(context.Background(), true, false); err != nil {
			t.Fatal(err)
		}
		if released.Load() != 1 {
			t.Fatal(released.Load())
		}
		if err := b.Close(context.Background()); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("close_cancels_and_drains_pending_acquisition", func(t *testing.T) {
		start := make(chan struct{})
		var drained atomic.Bool
		b := New(Options{Providers: map[string]Provider{"browser_stack": &ProviderFuncs{AcquireFunc: func(ctx context.Context, _ Request) (*Allocation, error) {
			close(start)
			<-ctx.Done()
			drained.Store(true)
			return nil, ctx.Err()
		}}}})
		done := make(chan error, 1)
		go func() { _, err := b.Acquire(context.Background(), testRequest()); done <- err }()
		<-start
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		if err := b.Close(ctx); err != nil {
			t.Fatal(err)
		}
		if !drained.Load() {
			t.Fatal("close acknowledged before acquisition stopped")
		}
		if err := <-done; !errors.Is(err, context.Canceled) {
			t.Fatal(err)
		}
	})
	t.Run("concurrent_release_is_exactly_once", func(t *testing.T) {
		var released atomic.Int64
		b := New(Options{Providers: map[string]Provider{"browser_stack": &ProviderFuncs{AcquireFunc: func(context.Context, Request) (*Allocation, error) {
			return ownedAllocation(func(context.Context, bool) error { released.Add(1); return nil }), nil
		}}}})
		l, err := b.Acquire(context.Background(), testRequest())
		if err != nil {
			t.Fatal(err)
		}
		var wg sync.WaitGroup
		for range 8 {
			wg.Go(func() {
				if _, err := l.Release(context.Background(), true, false); err != nil {
					t.Error(err)
				}
			})
		}
		wg.Wait()
		if err := b.Close(context.Background()); err != nil {
			t.Fatal(err)
		}
		if released.Load() != 1 {
			t.Fatal(released.Load())
		}
	})
}
func TestFixtureBrokerFaults(t *testing.T) {
	t.Run("quarantine_excludes_replacement_while_another_borrower_is_active", func(t *testing.T) {
		for _, cleanupFails := range []bool{false, true} {
			t.Run(fmt.Sprint(cleanupFails), func(t *testing.T) {
				starts, closes := 0, 0
				b := New(Options{Providers: map[string]Provider{"browser_stack": &ProviderFuncs{AcquireFunc: func(_ context.Context, request Request) (*Allocation, error) {
					starts++
					return ownedAllocation(func(context.Context, bool) error {
						closes++
						if cleanupFails && request.Affinity == testRequest().Affinity {
							return errors.New("uncertain cleanup")
						}
						return nil
					}), nil
				}}}})
				defer b.Close(t.Context())
				first, err := b.Acquire(t.Context(), testRequest())
				if err != nil {
					t.Fatal(err)
				}
				second, err := b.Acquire(t.Context(), testRequest())
				if err != nil || first.AllocationRef() != second.AllocationRef() {
					t.Fatal("initial borrowers must share the same allocation", err)
				}
				if _, err := first.Release(t.Context(), false, false); err != nil || closes != 0 {
					t.Fatal("active sibling must retain physical ownership", err, closes)
				}
				if _, err := b.reserveStack(t.Context(), testRequest()); !errors.Is(err, ErrCleanupBlocked) {
					t.Fatal("quarantined generation accepted a provisional reservation", err)
				}
				if _, err := b.Acquire(t.Context(), testRequest()); !errors.Is(err, ErrCleanupBlocked) || starts != 1 {
					t.Fatal("quarantined generation permitted same-affinity replacement", err, starts)
				}
				independent := testRequest()
				independent.Affinity = "independent"
				other, err := b.Acquire(t.Context(), independent)
				if err != nil || other.AllocationRef() == second.AllocationRef() {
					t.Fatal("independent allocation blocked or reused quarantined identity", err)
				}
				if _, err := other.Release(t.Context(), true, false); err != nil {
					t.Fatal(err)
				}
				_, err = second.Release(t.Context(), true, false)
				if (err != nil) != cleanupFails {
					t.Fatal("physical cleanup outcome lost", err)
				}
				replacement, err := b.Acquire(t.Context(), testRequest())
				if cleanupFails {
					if !errors.Is(err, ErrCleanupBlocked) || starts != 2 {
						t.Fatal("uncertain settlement cleared quarantine", err, starts)
					}
				} else {
					if err != nil || replacement.AllocationRef() == second.AllocationRef() {
						t.Fatal("settled generation was not replaced", err)
					}
					if _, err := replacement.Release(t.Context(), true, false); err != nil {
						t.Fatal(err)
					}
				}
			})
		}
	})
	t.Run("invalid_capability_has_no_side_effect", func(t *testing.T) {
		b := New(Options{Providers: map[string]Provider{"invalid": &ProviderFuncs{AcquireFunc: func(context.Context, Request) (*Allocation, error) {
			t.Fatal("invalid capability reached provider")
			return nil, nil
		}}}})
		if _, err := b.Acquire(context.Background(), Request{Capability: "invalid", UnitID: "unit"}); err == nil {
			t.Fatal("accepted invalid capability")
		}
	})
	for _, fault := range []string{"publish", "duplicate_resource", "partial_acquisition", "physical_cleanup"} {
		t.Run(fault, func(t *testing.T) {
			released := 0
			provider := &ProviderFuncs{AcquireFunc: func(context.Context, Request) (*Allocation, error) {
				a := ownedAllocation(func(context.Context, bool) error {
					released++
					if fault == "physical_cleanup" {
						return errors.New("physical failure")
					}
					return nil
				})
				if fault == "duplicate_resource" {
					a.ResourceIDs = []string{"same", "same"}
				}
				if fault == "partial_acquisition" {
					return a, errors.New("failure after allocation")
				}
				return a, nil
			}}
			b := New(Options{Providers: map[string]Provider{"browser_stack": provider}, Publish: func(LeaseRecord) error {
				if fault == "publish" {
					return errors.New("disk full")
				}
				return nil
			}})
			l, err := b.Acquire(context.Background(), testRequest())
			if fault == "physical_cleanup" {
				if err != nil {
					t.Fatal(err)
				}
				_, err = l.Release(context.Background(), true, false)
			}
			if err == nil || released != 1 {
				t.Fatal(fault, err, released)
			}
			_ = b.Close(context.Background())
			if released != 1 || b.Unresolved() != (fault == "physical_cleanup") {
				t.Fatal(released, b.Unresolved())
			}
		})
	}
	t.Run("optional_observation_failure_preserves_lifecycle", func(t *testing.T) {
		b := New(Options{Observe: func(Observation) { panic("optional") }})
		l, err := b.Acquire(context.Background(), Request{PostgresIsolation: "none", Capability: "none", UnitID: "unit"})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := l.Release(context.Background(), true, false); err != nil {
			t.Fatal(err)
		}
		if err := b.Close(context.Background()); err != nil {
			t.Fatal(err)
		}
	})
}

func testConsumingBrowserPreparation(t *testing.T) {
	for _, mode := range []string{"reuse", "pressure", "failure", "cancelled", "failed_cleanup", "concurrent"} {
		t.Run("consuming_preparation_"+mode, func(t *testing.T) {
			allocations, preparations, releases := 0, 0, 0
			var cleanup []CleanupRecord
			var allocationRef string
			var b *Broker
			b = New(Options{Cleanup: func(record CleanupRecord) { cleanup = append(cleanup, record) }, Providers: map[string]Provider{
				"browser_stack": &ProviderFuncs{AcquireFunc: func(_ context.Context, request Request) (*Allocation, error) {
					allocations++
					allocationRef = request.AllocationRef
					a := ownedAllocation(func(_ context.Context, healthy bool) error {
						releases++
						if mode == "failed_cleanup" {
							return errors.New("cleanup witness")
						}
						return nil
					})
					a.Prepare = func(ctx context.Context, consumer Request) error {
						preparations++
						if consumer.UnitID != "consumer:second" || consumer.LeaseID == "" {
							t.Error("preparation lost consuming identity")
						}
						if mode == "cancelled" {
							return context.Canceled
						}
						if mode == "failure" || mode == "failed_cleanup" {
							return fixtureFailure(errors.New("preparation witness"))
						}
						if mode == "concurrent" {
							blocked, cancel := context.WithTimeout(ctx, 50*time.Millisecond)
							defer cancel()
							if sibling, err := b.Acquire(blocked, testRequest()); sibling != nil || !errors.Is(err, context.DeadlineExceeded) {
								t.Error("borrower escaped preparation fence", sibling, err)
							}
						}
						return ctx.Err()
					}
					return a, nil
				}},
			}})
			request := testRequest()
			request.Preparation = &BrowserPreparation{Label: "first"}
			first, err := b.Acquire(t.Context(), request)
			if err != nil || preparations != 0 {
				t.Fatal("fresh allocation was reset", err, preparations)
			}
			firstRef := first.AllocationRef()
			if _, err := first.Release(t.Context(), true, true); err != nil {
				t.Fatal(err)
			}
			if mode == "pressure" {
				count, err := b.retireIdle(t.Context(), 1, func(Request) (bool, error) { return true, nil }, func(ctx context.Context, e *entry) error {
					return b.clean(ctx, e, Retirement{"capacity_pressure", "consumer:second"})
				})
				if err != nil || count != 1 {
					t.Fatal(count, err)
				}
			}
			request.UnitID = "consumer:second"
			request.Preparation.Label = "second"
			second, err := b.Acquire(t.Context(), request)
			failing := mode == "failure" || mode == "cancelled" || mode == "failed_cleanup"
			if failing {
				if err == nil || second != nil || preparations != 1 || releases != 1 {
					t.Fatal("failed preparation exposed or leaked allocation", err, preparations, releases)
				}
				if mode == "cancelled" && !errors.Is(err, context.Canceled) {
					t.Fatal(err)
				}
				if mode == "failed_cleanup" {
					if !b.Unresolved() || len(cleanup) == 0 || cleanup[len(cleanup)-1].Err == nil {
						t.Fatal("cleanup failure was hidden", cleanup)
					}
					if later, err := b.Acquire(t.Context(), request); err == nil || later != nil {
						t.Fatal("quarantined allocation admitted")
					}
					return
				}
				request.UnitID = "consumer:later"
				later, err := b.Acquire(t.Context(), request)
				if err != nil || later.AllocationRef() == firstRef || preparations != 1 {
					t.Fatal("later work did not initialize fresh", err)
				}
				if _, err := later.Release(t.Context(), true, false); err != nil {
					t.Fatal(err)
				}
			} else {
				if err != nil {
					t.Fatal(err)
				}
				if mode == "pressure" {
					if preparations != 0 || allocationRef == firstRef || allocations != 2 {
						t.Fatal("retired generation was treated as reusable")
					}
				} else if preparations != 1 || second.AllocationRef() != firstRef || allocations != 1 {
					t.Fatal("actual reuse did not prepare exactly once")
				}
				if mode == "concurrent" {
					if sibling, err := b.Acquire(t.Context(), request); sibling != nil || err == nil || preparations != 1 || releases != 0 {
						t.Fatal("reset overlapped an active consumer", sibling, err)
					}
				}
				if _, err := second.Release(t.Context(), true, false); err != nil {
					t.Fatal(err)
				}
			}
			if err := b.Close(t.Context()); err != nil || b.Unresolved() {
				t.Fatal(err)
			}
			if releases != allocations {
				t.Fatal("unsettled ownership", releases, allocations)
			}
		})
	}
}
