package recoverybrowsertest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/postgrescleanup"
)

func TestLifecyclePreservesEveryCleanupFailure(t *testing.T) {
	t.Run("drop_wait_diagnostics_exclude_private_values", func(t *testing.T) {
		if got := postgrescleanup.DropWaitDescription("IPC", "ProcSignalBarrier", 0); got != "type=IPC event=ProcSignalBarrier blockers=0" {
			t.Fatal(got)
		}
		for _, input := range [][2]string{{"private_type", "ProcSignalBarrier"}, {"Extension", "PrivateCredentialAlpha123"}, {"Lock", "postgres://private:secret@host"}} {
			got := postgrescleanup.DropWaitDescription(input[0], input[1], 1)
			if strings.Contains(got, "private") || strings.Contains(got, "PrivateCredential") || strings.Contains(got, "secret") {
				t.Fatalf("private wait leaked: %s", got)
			}
		}
		if got := postgrescleanup.DropWaitDescription("IO", "DataFileSync", 0); got != "type=IO event=DataFileSync blockers=0" {
			t.Fatal(got)
		}
		if got := postgrescleanup.CheckpointProgressDescription(-1, 0, 0); got != "progress_unavailable" {
			t.Fatal(got)
		}
		if got := postgrescleanup.CheckpointProgressDescription(7, 6, 200); got != "requested=7 done=6 buffers=200" {
			t.Fatal(got)
		}
	})

	t.Run("ordered_owners_parallel_drops_and_all_failures", func(t *testing.T) {
		var output bytes.Buffer
		lifecycle := NewLifecycle("attempt", &output, time.Second)
		cleanup := fixtureCleanup{lifecycle: lifecycle}
		var httpClosed, runtimeClosed, handlesClosed, storesClosed atomic.Int32
		started, release := make(chan struct{}, 2), make(chan struct{})
		ownerFailure, sourceFailure, targetFailure := errors.New("owner close"), errors.New("source drop"), errors.New("target drop")
		for i, d := range []*deploymentCleanup{&cleanup.source, &cleanup.target} {
			d.http = func(context.Context) error {
				httpClosed.Add(1)
				if i == 0 {
					return ownerFailure
				}
				return nil
			}
			d.runtime = func(context.Context) error {
				if httpClosed.Load() != 2 {
					t.Error("runtime before both HTTP lifetimes stopped")
				}
				runtimeClosed.Add(1)
				return nil
			}
			handle := func(context.Context) error {
				if runtimeClosed.Load() != 2 {
					t.Error("pool closed before runtime")
				}
				handlesClosed.Add(1)
				return nil
			}
			d.runtimePool, d.recovery = handle, handle
			d.database = func(ctx context.Context) error {
				if handlesClosed.Load() != 4 {
					t.Error("drop before all handles closed")
				}
				if deadline, ok := ctx.Deadline(); !ok || time.Until(deadline) > time.Second {
					t.Error("drop lost owner deadline")
				}
				started <- struct{}{}
				select {
				case <-release:
				case <-ctx.Done():
					return ctx.Err()
				}
				if i == 0 {
					return sourceFailure
				}
				return targetFailure
			}
			d.objects = func(context.Context) error { storesClosed.Add(1); return nil }
		}
		ctx, cancel := context.WithTimeout(t.Context(), time.Second)
		defer cancel()
		done := make(chan error, 1)
		go func() { done <- cleanup.close(ctx) }()
		for range 2 {
			select {
			case <-started:
			case <-ctx.Done():
				t.Fatal("independent drops did not overlap")
			}
		}
		if storesClosed.Load() != 0 {
			t.Fatal("storage closed before database settlement")
		}
		close(release)
		original := errors.New("body failure")
		result := errors.Join(original, <-done)
		for _, cause := range []error{original, ownerFailure, sourceFailure, targetFailure} {
			if !errors.Is(result, cause) {
				t.Fatalf("lost cause %v: %v", cause, result)
			}
		}
		if storesClosed.Load() != 2 {
			t.Fatal("failure skipped independent storage cleanup")
		}
		lines := strings.Split(strings.TrimSpace(output.String()), "\n")
		if len(lines) != 24 {
			t.Fatal("incomplete cleanup events", len(lines))
		}
		for _, line := range lines {
			var event map[string]any
			if err := json.Unmarshal([]byte(line), &event); err != nil {
				t.Fatal(err)
			}
			if event["attempt_id"] != "attempt" || event["deadline_ms"].(float64) <= 0 || event["deadline_ms"].(float64) > 1000 {
				t.Fatal("lost effective deadline", event)
			}
		}
	})
	t.Run("partial_acquisition_and_expired_owner", func(t *testing.T) {
		var output bytes.Buffer
		cleanup := fixtureCleanup{lifecycle: NewLifecycle("partial", &output, time.Second)}
		listenerClosed := false
		cleanup.source.listener = func(context.Context) error { listenerClosed = true; return nil }
		cleanup.source.database = func(ctx context.Context) error { <-ctx.Done(); return ctx.Err() }
		ctx, cancel := context.WithCancel(t.Context())
		cancel()
		if err := cleanup.close(ctx); !errors.Is(err, context.Canceled) || !listenerClosed {
			t.Fatal("partial cleanup lost owner", err)
		}
		if strings.Contains(output.String(), "target.database") {
			t.Fatal("unowned database was retired")
		}
		if !strings.Contains(output.String(), `"deadline_ms":0`) {
			t.Fatal("expired budget was not explicit")
		}
	})
}
