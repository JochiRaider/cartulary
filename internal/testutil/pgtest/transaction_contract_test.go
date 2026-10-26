package pgtest

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"
)

func TestRollbackFixtureCleanupRejectsUnprovenIsolation(t *testing.T) {
	for _, failed := range []string{"", "rollback", "close"} {
		t.Run("failure_"+failed, func(t *testing.T) {
			fixture := &transactionDatabase{db: &TestDatabase{Name: "retained-owned-database"}}
			if err := fixture.lockForBorrow(); err != nil {
				t.Fatal(err)
			}
			cancelled, cancel := context.WithCancel(t.Context())
			cancel()
			var contexts []context.Context
			calls := []string{}
			operation := func(name string) func(context.Context) error {
				return func(ctx context.Context) error {
					if ctx.Err() != nil {
						t.Fatal("cleanup inherited test cancellation")
					}
					deadline, ok := ctx.Deadline()
					if !ok || time.Until(deadline) <= 0 || time.Until(deadline) > 5*time.Second {
						t.Fatal("missing bounded cleanup budget")
					}
					if len(contexts) > 0 && contexts[0].Err() == nil {
						t.Fatal("previous budget was not settled")
					}
					contexts = append(contexts, ctx)
					calls = append(calls, name)
					if failed == name {
						return context.DeadlineExceeded
					}
					return nil
				}
			}
			err := fixture.settle(cancelled, operation("rollback"), operation("close"))
			fixture.mu.Unlock()
			if !reflect.DeepEqual(calls, []string{"rollback", "close"}) {
				t.Fatal(calls)
			}
			if (err != nil) != (failed != "") {
				t.Fatal("incorrect cleanup result", err)
			}
			if fixture.db.Name != "retained-owned-database" {
				t.Fatal("cleanup discarded ownership")
			}
			borrowErr := fixture.lockForBorrow()
			if failed == "" {
				if borrowErr != nil {
					t.Fatal(borrowErr)
				}
				fixture.mu.Unlock()
			} else {
				if !errors.Is(borrowErr, context.DeadlineExceeded) {
					t.Fatal("poisoned fixture admitted another borrower", borrowErr)
				}
				if !fixture.mu.TryLock() {
					t.Fatal("rejected borrower retained fixture mutex")
				}
				fixture.mu.Unlock()
			}
		})
	}
}
