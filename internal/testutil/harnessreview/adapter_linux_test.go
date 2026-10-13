//go:build linux

package harnessreview

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"testing/synctest"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessfixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesssession"
)

func TestReviewIPC(t *testing.T) {
	if stage := os.Getenv("CARTULARY_REVIEW_STALL_HELPER"); stage != "" {
		preparationStallHelper(t, stage)
		return
	}
	t.Run("stop_drains_accepted_reply", testStopDrainsAcceptedReply)
	t.Run("cleanup_receipts_keep_closed_context_without_private_errors", func(t *testing.T) {
		for _, test := range []struct {
			operation, subject, condition string
			cause                         error
		}{
			{"provider_close", "preparation_child", "timeout", context.DeadlineExceeded},
			{"managed_suite", "test_services", "cancelled", context.Canceled},
			{"browser_process", "browser", "invalid_artifact", harnesslocal.ErrUnsafe},
			{"diagnostic_scope", "browser", "resource_conflict", harnesssession.ErrUnresolved},
			{"browser_stack", "browser_stack", "resource_conflict", harnessfixtures.ErrCleanupBlocked},
			{"provider_close", "preparation_child", "child_failed", failure("session_lost")},
			{"host_release", "host_lock", "unknown", errors.New("private-token-and-path")},
		} {
			cause := errors.Join(errors.New("private-token-and-path"), test.cause)
			value := cleanupFailure(test.operation, cause)
			if value.Record.validate() != nil || value.Record.Code != "cleanup_failed" || *value.Record.Phase != "cleanup" || *value.Record.Subject != test.subject || *value.Record.Condition != test.condition || *value.Record.Recovery != "exact_stop" || !errors.Is(value, test.cause) {
				t.Fatal("lost cleanup boundary, classification or internal cause", value.Record)
			}
			data, err := json.Marshal(value.Record)
			if err != nil || strings.Contains(string(data), "private-token-and-path") {
				t.Fatal("private error escaped structural receipt")
			}
		}
	})
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
		failures := d.Close(context.Background())
		if len(failures) == 0 || closed != 1 {
			t.Fatal("lost failure or starved independent close", failures, closed)
		}
		found := false
		for _, err := range failures {
			f := normalized(err, "cleanup_failed").Record
			found = found || f.Phase != nil && *f.Phase == "cleanup" && *f.Subject == "browser" && *f.Condition == "resource_conflict"
		}
		if !found {
			t.Fatal("resource cleanup cause lost before receipt")
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

// A pipe blocks the final reply until its independent client reads it. Session
// completion must not let the process owner return while that reply is pending.
func testStopDrainsAcceptedReply(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		session := NewSession("uireview-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "run", "artifacts", &testDriver{}, &controlTestPublisher{})
		if err := session.Prepare(); err != nil {
			t.Fatal(err)
		}
		server, client := net.Pipe()
		defer client.Close()
		listener := &controlTestListener{connection: make(chan net.Conn, 1), done: make(chan struct{})}
		listener.connection <- server
		record := &registryRecord{Session: session.id, Token: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa"}
		drain := serveControls(listener, record, session)
		message := controlRequest{Schema: channelSchema, Session: session.id, Token: record.Token, Command: "ui-review-stop"}
		data, err := jsonDocument(channelSchema, message)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := client.Write(data); err != nil {
			t.Fatal(err)
		}
		<-session.Done()
		drained := make(chan struct{})
		go func() { drain(); close(drained) }()
		synctest.Wait()
		select {
		case <-drained:
			t.Error("owner returned before accepted stop reply was delivered")
		default:
		}
		data, err = bufio.NewReader(client).ReadBytes('\n')
		if err != nil {
			t.Fatal(err)
		}
		var result Result
		if err := json.Unmarshal(data, &result); err != nil || result.Status != "ok" || result.State == nil || *result.State != "closed" {
			t.Fatalf("stop reply: %s; %v", data, err)
		}
		<-drained
	})
}

type controlTestListener struct {
	connection chan net.Conn
	done       chan struct{}
	once       sync.Once
}

func (l *controlTestListener) Accept() (net.Conn, error) {
	select {
	case connection := <-l.connection:
		return connection, nil
	case <-l.done:
		return nil, net.ErrClosed
	}
}
func (l *controlTestListener) Close() error { l.once.Do(func() { close(l.done) }); return nil }
func (l *controlTestListener) Addr() net.Addr {
	return &net.UnixAddr{Name: "control-test", Net: "unix"}
}

type controlTestPublisher struct{ testPublisher }

func (p *controlTestPublisher) Terminal(l Locator, r Receipt, failed bool) (*Reference, error) {
	return &Reference{Path: "ui-review/terminal.json", SHA256: strings.Repeat("a", 64), Bytes: 1, MediaType: "application/json"}, nil
}
