//go:build linux

package harnesssession

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"sync"
	"testing"
	"testing/synctest"
	"time"
)

func testCleanupAttempts(t *testing.T) {
	t.Run("equivalent_process_proofs_share_attempt", func(t *testing.T) {
		runtime, _, _ := runtimeFixture(t)
		calls := 0
		for _, target := range []json.RawMessage{json.RawMessage(`{"boot":"fixture","pid":12345,"start":"1","group":true}`), json.RawMessage(`{"group":true,"start":"1","pid":12345,"boot":"fixture"}`)} {
			if err := runtime.AttemptCleanup(context.Background(), "helper_process", target, func(context.Context) error { calls++; return nil }); err != nil {
				t.Fatal(err)
			}
		}
		if calls != 1 {
			t.Fatal("proof field ordering duplicated cleanup budget", calls)
		}
	})
	t.Run("cleanup_attempts_have_independent_budgets", func(t *testing.T) {
		runtime, _, _ := runtimeFixture(t)
		synctest.Test(t, func(t *testing.T) {
			calls := 0
			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			before := time.Now()
			file := filepath.Join(runtime.Root, "browser", "owner.json")
			err := runtime.AttemptCleanup(ctx, "browser_stack", file, func(step context.Context) error { calls++; <-step.Done(); return step.Err() })
			if !errors.Is(err, context.DeadlineExceeded) || time.Since(before) != 120*time.Second {
				t.Fatal("wrong owner deadline", err)
			}
			var repeats sync.WaitGroup
			for range 8 {
				repeats.Go(func() {
					if !errors.Is(runtime.AttemptCleanup(ctx, "browser_stack", file, func(context.Context) error { t.Error("second cleanup attempt"); return nil }), context.DeadlineExceeded) {
						t.Error("lost original outcome")
					}
				})
			}
			repeats.Wait()
			before = time.Now()
			sibling := filepath.Join(runtime.Root, "suite", "lease.json")
			err = runtime.AttemptCleanup(ctx, "managed_suite", sibling, func(step context.Context) error {
				calls++
				end, _ := step.Deadline()
				if time.Until(end) != 120*time.Second {
					t.Error("sibling budget starved")
				}
				return nil
			})
			if err != nil || calls != 2 || time.Since(before) != 0 {
				t.Fatal("independent cleanup skipped", err, calls)
			}
		})
	})
}
