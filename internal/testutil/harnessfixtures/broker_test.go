package harnessfixtures

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func ownedAllocation(release func(context.Context, bool) error) *Allocation {
	return &Allocation{Ownership: "owned", ResourceIDs: []string{"fixture:one"}, Environment: map[string]string{"FIXTURE": "value"}, Release: release}
}
func testRequest() Request {
	return Request{Capability: "browser_stack", UnitID: "unit:a", Affinity: "browser"}
}
func TestFixtureBrokerOwnership(t *testing.T) {
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
		l, err := b.Acquire(context.Background(), Request{Capability: "none", UnitID: "unit"})
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
