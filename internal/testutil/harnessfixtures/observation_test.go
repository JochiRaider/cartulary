package harnessfixtures

import (
	"context"
	"errors"
	"testing"
	"time"
)

func testAllocationLifecycleObservations(t *testing.T) {
	t.Run("retirement_trigger_is_frozen_before_settlement", func(t *testing.T) {
		cancelled := false
		var observed *AllocationLifecycle
		b := New(Options{Cancelled: func() bool { return cancelled }, Observe: func(v Observation) {
			if v.Lifecycle != nil {
				observed = v.Lifecycle
			}
		}, Providers: map[string]Provider{"browser_stack": &ProviderFuncs{AcquireFunc: func(context.Context, Request) (*Allocation, error) {
			return ownedAllocation(func(context.Context, bool) error { cancelled = true; return nil }), nil
		}}}})
		lease, err := b.Acquire(t.Context(), testRequest())
		if err != nil {
			t.Fatal(err)
		}
		_, _ = lease.Release(t.Context(), true, true)
		_, err = b.retireIdle(t.Context(), 1, nil, func(ctx context.Context, e *entry) error {
			return b.clean(ctx, e, Retirement{"capacity_pressure", "waiting"})
		})
		if err != nil || observed == nil || observed.Reason != "capacity_pressure" {
			t.Fatal("late cancellation rewrote actual retirement cause", observed, err)
		}
		_ = b.Close(t.Context())
	})
	for _, reason := range []string{"acquisition_failure", "unhealthy_borrower", "terminal_borrower_release", "capacity_pressure", "exclusive_admission", "invocation_cancellation", "run_finalization"} {
		t.Run("lifecycle/"+reason, func(t *testing.T) {
			for _, cleanupFails := range []bool{false, true} {
				var records []AllocationLifecycle
				cancelled := false
				tick := 0
				b := New(Options{Clock: func() time.Time { tick++; return time.Unix(0, int64(tick)*1000000) }, Cancelled: func() bool { return cancelled }, Observe: func(v Observation) {
					if v.Lifecycle != nil {
						records = append(records, *v.Lifecycle)
					}
				}, Providers: map[string]Provider{"browser_stack": &ProviderFuncs{AcquireFunc: func(context.Context, Request) (*Allocation, error) {
					allocation := &Allocation{Ownership: "owned", Environment: map[string]string{}, Release: func(context.Context, bool) error {
						if cleanupFails {
							return errors.New("cleanup")
						}
						return nil
					}}
					if reason == "acquisition_failure" {
						return allocation, errors.New("acquisition")
					}
					return allocation, nil
				}}}})
				req := Request{UnitID: "u", Capability: "browser_stack", PostgresIsolation: "exclusive", Services: []string{"postgres"}, Affinity: "a"}
				lease, err := b.Acquire(t.Context(), req)
				if reason == "acquisition_failure" {
					if err == nil {
						t.Fatal("expected acquisition failure")
					}
				} else {
					if err != nil {
						t.Fatal(err)
					}
					switch reason {
					case "unhealthy_borrower":
						_, _ = lease.Release(t.Context(), false, false)
					case "terminal_borrower_release":
						_, _ = lease.Release(t.Context(), true, false)
					case "invocation_cancellation":
						cancelled = true
						_, _ = lease.Release(t.Context(), false, false)
					case "run_finalization":
						_, _ = lease.Release(t.Context(), true, true)
						_ = b.Close(t.Context())
					default:
						_, _ = lease.Release(t.Context(), true, true)
						_, _ = b.retireIdle(t.Context(), 0, nil, func(ctx context.Context, e *entry) error { return b.clean(ctx, e, Retirement{reason, "trigger"}) })
					}
				}
				_ = b.Close(t.Context())
				if len(records) != 1 {
					t.Fatalf("records=%d", len(records))
				}
				v := records[0]
				if v.Reason != reason {
					t.Fatalf("reason=%s", v.Reason)
				}
				if (v.Settlement == "unresolved") != cleanupFails {
					t.Fatalf("settlement=%s", v.Settlement)
				}
				if v.Acquisition.End < v.Acquisition.Start || v.Retirement.Start < v.Acquisition.End || v.Retirement.End < v.Retirement.Start {
					t.Fatal("interval order")
				}
				if reason == "run_finalization" && v.Trigger != nil {
					t.Fatal("invented finalization trigger")
				}
				if reason != "acquisition_failure" && (len(v.Borrows) != 1 || v.Borrows[0].UnitID != "u") {
					t.Fatal("borrow identity")
				}
			}
		})
	}
}
