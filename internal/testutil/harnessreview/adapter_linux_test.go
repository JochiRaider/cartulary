//go:build linux

package harnessreview

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessfixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesssession"
)

func TestReviewIPC(t *testing.T) {
	if stage := os.Getenv("CARTULARY_REVIEW_STALL_HELPER"); stage != "" {
		preparationStallHelper(t, stage)
		return
	}
	testPreparationDeadlines(t)
	t.Run("diagnostic_failure_does_not_starve_fixture_close", func(t *testing.T) {
		root, run := t.TempDir(), t.TempDir()
		r, err := harnesssession.NewRuntime(root, run, "fixture", t.TempDir())
		if err != nil {
			t.Fatal(err)
		}
		defer r.Close()
		proof := map[string]any{"boot": "fixture", "start": "1", "uid": os.Getuid(), "token": strings.Repeat("a", 64)}
		if err := r.RecordDependentResource("diagnostic_scope", proof, "acquired", []string{}); err != nil {
			t.Fatal(err)
		}
		defer r.RecordResource("diagnostic_scope", proof, "released")
		closed := 0
		fixtures := harnessfixtures.NewProduction(harnessfixtures.ProductionOptions{Root: root, RunRoot: run, Runtime: r}, harnessfixtures.Options{})
		fixtures.Suite.Acquire = func(context.Context) (*harnessfixtures.Suite, error) {
			return &harnessfixtures.Suite{LeaseFile: filepath.Join(r.Root, "suite.json"), Close: func(context.Context) error { closed++; return nil }}, nil
		}
		if _, err := fixtures.Suite.Ensure(context.Background()); err != nil {
			t.Fatal(err)
		}
		d := &productionDriver{root: root, runRoot: run, runtime: r, fixtures: fixtures, environment: map[string]string{"NODE_BIN": filepath.Join(root, "missing-node")}}
		if failures := d.Close(context.Background()); len(failures) == 0 || closed != 1 {
			t.Fatal("lost failure or starved independent close", failures, closed)
		}
	})
	t.Run("independent_response_rejections", func(t *testing.T) {
		for _, variant := range []string{"valid", "duplicate", "truncated", "oversized", "unknown_field", "foreign_run", "foreign_lease", "unknown_id", "wrong_operation", "unclassified_exit", "wrong_exit", "too_many_failures"} {
			t.Run(variant, func(t *testing.T) {
				r, w, err := os.Pipe()
				if err != nil {
					t.Fatal(err)
				}
				defer r.Close()
				done := make(chan struct{})
				close(done)
				response := make(chan adapterReply, 1)
				a := &adapter{runtime: &harnesssession.Runtime{RunID: "fixture", LeaseID: "00000000-0000-4000-8000-000000000000"}, session: "uireview-" + strings.Repeat("a", 32), channel: r, done: done, readDone: make(chan struct{}), idle: make(chan struct{}), pending: map[int]pendingRequest{1: {operation: "build-web", response: response}}}
				outcome := adapterOutcome{Counts: Counts{}, Private: []PrivateReference{}, Failures: []Failure{}}
				message := map[string]any{"schema_id": adapterSchema, "session_id": a.session, "run_id": "fixture", "runtime_lease_id": a.runtime.LeaseID, "request_id": 1, "kind": "response", "operation": "build-web", "facts": nil, "outcome": outcome}
				switch variant {
				case "unknown_field":
					message["private_detail"] = "must never escape"
				case "foreign_run":
					message["run_id"] = "foreign"
				case "foreign_lease":
					message["runtime_lease_id"] = "11111111-1111-4111-8111-111111111111"
				case "unknown_id":
					message["request_id"] = 2
				case "wrong_operation":
					message["operation"] = "seed"
				case "unclassified_exit":
					outcome.Exit = 1
				case "wrong_exit":
					outcome.Failures = []Failure{failure("startup_failed").Record}
					outcome.Exit = 1
				case "too_many_failures":
					for range 65 {
						outcome.Failures = append(outcome.Failures, failure("cleanup_failed").Record)
					}
					outcome.Exit = 12
				}
				message["outcome"] = outcome
				data, err := json.Marshal(message)
				if err != nil {
					t.Fatal(err)
				}
				if variant == "oversized" {
					data = []byte(strings.Repeat(" ", harnesssession.MaximumFrame+1))
				}
				if variant != "truncated" {
					data = append(data, '\n')
				}
				if variant == "duplicate" {
					data = append(data, data...)
				}
				go func() { defer w.Close(); _, _ = w.Write(data) }()
				a.read()
				if (a.failure == nil) != (variant == "valid") {
					t.Fatal("unexpected response acceptance", a.failure)
				}
			})
		}
	})

	root, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	environment, err := installedEnvironment(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, forged := range []bool{false, true} {
		t.Run(map[bool]string{false: "actual_private_adapter", true: "foreign_session_rejected"}[forged], func(t *testing.T) {
			run, scratch := t.TempDir(), t.TempDir()
			if err := os.Chmod(run, 0700); err != nil {
				t.Fatal(err)
			}
			runtime, err := harnesssession.NewRuntime(root, run, "fixture", scratch)
			if err != nil {
				t.Fatal(err)
			}
			t.Cleanup(func() {
				if err := runtime.Close(); err != nil {
					t.Error(err)
				}
			})
			session := "uireview-" + strings.Repeat("a", 32)
			worker, err := startAdapter(root, session, runtime, environment, nil)
			if err != nil {
				t.Fatal(err)
			}
			ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer cancel()
			t.Cleanup(func() {
				worker.kill()
				<-worker.done
				if failures := (harnessfixtures.ProductionOptions{Root: root, RunRoot: run, Runtime: runtime, Environment: environment}).Recover(context.Background()); failures != nil {
					t.Error(failures)
				}
			})
			if forged {
				raw := `{"schema_id":"cartulary.harness_review_adapter.v2","session_id":"uireview-bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb","run_id":"fixture","runtime_lease_id":"` + runtime.LeaseID + `","request_id":1,"kind":"request","operation":"close","payload":null}` + "\n"
				if _, err := worker.channel.Write([]byte(raw)); err != nil {
					t.Fatal(err)
				}
				select {
				case <-worker.done:
				case <-ctx.Done():
					t.Fatal("forged request did not close private adapter")
				}
				return
			}
			profile := map[string]any{"pins_sha256": strings.Repeat("a", 64), "lock_sha256": strings.Repeat("b", 64), "node_version": "fixture", "playwright_version": "fixture", "sharp_version": "fixture", "axe_version": "fixture"}
			reply, err := worker.call(ctx, "initialize", map[string]any{"input": reviewInput{Mode: "artifacts", Policy: "installed_only", Presentation: "ui"}, "profile": profile, "repository": root, "run_root": run, "runtime_root": runtime.Root})
			if err != nil || reply.Facts == nil || !strings.HasPrefix(reply.Facts.Source, "sha256:") {
				t.Fatal("initialize", reply.Outcome, err)
			}
			if _, err = worker.call(ctx, "validate", map[string]any{"command": "ui-review-report", "request": nil}); err != nil {
				t.Fatal(err)
			}
			var calls sync.WaitGroup
			for range 16 {
				calls.Go(func() {
					if _, err := worker.call(ctx, "validate", map[string]any{"command": "ui-review-report", "request": nil}); err != nil {
						t.Error("concurrent private request", err)
					}
				})
			}
			calls.Wait()
			if err = worker.close(ctx); err != nil {
				t.Fatal("close", err)
			}
			if failures := (harnessfixtures.ProductionOptions{Root: root, RunRoot: run, Runtime: runtime, Environment: environment}).Recover(ctx); failures != nil {
				t.Fatal(failures)
			}
			resources, err := runtime.Resources(4096)
			if err != nil || len(resources) > 0 {
				t.Fatal("producer not reaped", resources, err)
			}
		})
	}
}
