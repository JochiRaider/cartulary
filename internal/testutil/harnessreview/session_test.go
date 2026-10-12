package harnessreview

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"testing/synctest"
	"time"
)

type testDriver struct {
	prepare   func(context.Context) (*int, error)
	validate  func(context.Context, string, jsonRequest) error
	execute   func(context.Context, string, jsonRequest, *string, int) (Operation, error)
	close     func(context.Context) []error
	interrupt func()
}

func (d *testDriver) Prepare(c context.Context) (*int, error) {
	if d.prepare != nil {
		return d.prepare(c)
	}
	return pointer(0), nil
}
func (d *testDriver) Validate(c context.Context, s string, r jsonRequest) error {
	if d.validate != nil {
		return d.validate(c, s, r)
	}
	return nil
}
func (d *testDriver) Execute(c context.Context, s string, r jsonRequest, b *string, id int) (Operation, error) {
	return d.execute(c, s, r, b, id)
}
func (d *testDriver) Close(c context.Context) []error {
	if d.close != nil {
		return d.close(c)
	}
	return nil
}
func (d *testDriver) Interrupt() {
	if d.interrupt != nil {
		d.interrupt()
	}
}

type testPublisher struct {
	mu             sync.Mutex
	locators       []Locator
	receipts       []Receipt
	terminals      []Receipt
	terminalError  error
	operationError error
}

func (p *testPublisher) Locator(l Locator) error {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.locators = append(p.locators, l)
	return nil
}
func (p *testPublisher) Operation(r Receipt) (*Reference, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	if len(p.terminals) > 0 {
		panic("late operation publication")
	}
	p.receipts = append(p.receipts, r)
	return &Reference{Path: "operation"}, p.operationError
}
func (p *testPublisher) Terminal(_ Locator, r Receipt, _ bool) (*Reference, error) {
	p.mu.Lock()
	defer p.mu.Unlock()
	p.terminals = append(p.terminals, r)
	return &Reference{Path: "terminal"}, p.terminalError
}
func readySession(t *testing.T, d *testDriver, p *testPublisher) *Session {
	t.Helper()
	s := NewSession("uireview-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "run", "artifacts", d, p)
	s.lockTimeout = 20 * time.Millisecond
	s.operationTimeout = time.Second
	s.cleanupTimeout = time.Second
	if err := s.Prepare(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Stop(nil); awaitSession(t, s) })
	return s
}
func awaitSession(t *testing.T, s *Session) {
	t.Helper()
	select {
	case <-s.Done():
	case <-time.After(3 * time.Second):
		t.Fatal("native session did not settle")
	}
}
func TestReviewLifecycle(t *testing.T) {
	t.Run("borrowed_origin_deadline_has_no_destructive_effect", func(t *testing.T) {
		server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(503) }))
		defer server.Close()
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Millisecond)
		defer cancel()
		if normalized(waitOrigin(ctx, server.URL), "unsafe_artifact").Record.Code != "readiness_expired" {
			t.Fatal("wrong deadline failure")
		}
		response, err := http.Get(server.URL)
		if err != nil {
			t.Fatal("borrowed origin stopped", err)
		}
		response.Body.Close()
	})
	t.Run("bounded_secondary_diagnostics_preserve_primary", func(t *testing.T) {
		values := []Failure{failure("startup_failed").Record}
		for _, phase := range []string{"prerequisites", "build", "service_acquisition", "service_readiness", "seeding", "browser_start", "cleanup"} {
			for code := range failurePolicy {
				f := contextual(code, phase, "server_harness", "child_failed", "inspect_failure").Record
				if err := f.validate(); err != nil {
					t.Fatal(err)
				}
				values = appendFailure(values, f)
			}
		}
		if len(values) != 64 || values[0].Code != "startup_failed" || values[63].Code != "diagnostic_invalid" {
			t.Fatal("unbounded failure projection")
		}
	})

	t.Run("exact_eight_hour_lifetime", func(t *testing.T) {
		synctest.Test(t, func(t *testing.T) {
			p := &testPublisher{}
			s := readySession(t, &testDriver{}, p)
			synctest.Wait()
			time.Sleep(8*time.Hour - time.Nanosecond)
			if !s.Ready() {
				t.Fatal("expired before boundary")
			}
			time.Sleep(time.Nanosecond)
			synctest.Wait()
			select {
			case <-s.Done():
			default:
				t.Fatal("did not close at boundary")
			}
			if len(p.terminals) != 1 || s.Value("ui-review-stop").Exit != 0 {
				t.Fatal(p.terminals)
			}
		})
	})
	t.Run("every_primary_survives_cleanup_and_repetition", func(t *testing.T) {
		expected := map[string]int{"invalid_request": 2, "tool_configuration": 2, "session_mismatch": 2,
			"target_unavailable": 1, "navigation_boundary": 1, "environment_unavailable": 3, "startup_failed": 3,
			"readiness_expired": 3, "capacity_exceeded": 4, "invalid_artifact": 11, "unstable_capture": 11,
			"observation_limit": 11, "analysis_failed": 1, "operation_expired": 13, "session_lost": 3,
			"unsafe_artifact": 11, "input_boundary": 11, "cleanup_failed": 12, "interrupted": 15,
			"build_failed": 1, "fixture_failed": 3, "diagnostic_invalid": 11, "preparation_failed": 1}
		if len(expected) != len(failurePolicy) {
			t.Fatal("update independent primary expectations")
		}
		for code, exit := range expected {
			t.Run(code, func(t *testing.T) {
				p := &testPublisher{}
				s := readySession(t, &testDriver{close: func(context.Context) []error { return []error{errors.New("private cleanup detail")} }}, p)
				primary := failure(code)
				if code == "build_failed" || code == "fixture_failed" || code == "diagnostic_invalid" || code == "preparation_failed" {
					primary = contextual(code, "build", "server_harness", "child_failed", "build_server_harness")
				}
				s.Stop(primary)
				awaitSession(t, s)
				first := s.Value("ui-review-stop")
				if first.Exit != exit || first.Failures[0].Code != code || p.terminals[0].Cleanup != "failed" {
					t.Fatal(first)
				}
				for range 3 {
					s.Stop(failure("session_lost"))
					if s.Value("ui-review-stop").Exit != exit || len(p.terminals) != 1 {
						t.Fatal("terminal changed")
					}
				}
			})
		}
	})
	t.Run("late_acquisition_is_closed_before_terminal", func(t *testing.T) {
		entered, release := make(chan struct{}), make(chan struct{})
		var acquired, closed bool
		p := &testPublisher{}
		s := NewSession("session", "run", "seeded", &testDriver{prepare: func(ctx context.Context) (*int, error) {
			close(entered)
			<-release
			acquired = true
			return nil, ctx.Err()
		}, close: func(context.Context) []error {
			if !acquired {
				t.Error("cleanup before acquisition settled")
			}
			closed = true
			return nil
		}}, p)
		prepared := make(chan error, 1)
		go func() { prepared <- s.Prepare() }()
		<-entered
		s.Stop(nil)
		close(release)
		<-prepared
		awaitSession(t, s)
		if !closed || len(p.terminals) != 1 || *s.Value("ui-review-stop").State != "closed" {
			t.Fatal("late acquisition leaked")
		}
	})

	t.Run("undrained_preparation_blocks_cleanup_and_terminal", func(t *testing.T) {
		entered, release := make(chan struct{}), make(chan struct{})
		p := &testPublisher{}
		s := NewSession("session", "run", "seeded", &testDriver{prepare: func(context.Context) (*int, error) { close(entered); <-release; return nil, context.Canceled }, close: func(context.Context) []error { t.Error("cleanup ran under a live preparation producer"); return nil }}, p)
		s.cleanupTimeout = 10 * time.Millisecond
		prepared := make(chan error, 1)
		go func() { prepared <- s.Prepare() }()
		<-entered
		s.Stop(nil)
		awaitSession(t, s)
		if len(p.terminals) != 0 || s.Value("ui-review-stop").Exit != 12 {
			t.Fatal("undrained owner finalized", p.terminals)
		}
		close(release)
		<-prepared
	})
	t.Run("compiled_action_crosses_private_command_boundaries", func(t *testing.T) {
		request := jsonRequest{"schema_id": "cartulary.ui_review_action.v3", "expected_epoch": float64(0), "action": "navigate", "parameters": map[string]any{"path": "/slow"}}
		if err := validateCommand("ui-browser", request, nil); err != nil {
			t.Fatal("action", err)
		}
		if _, err := jsonDocument(commandSchema, map[string]any{"schema_id": commandSchema, "command": "ui-browser", "request": request, "locator": "/tmp/fixture/ui-review/session.json", "bundle_id": nil}); err != nil {
			t.Fatal("command", err)
		}
		if _, err := jsonDocument(channelSchema, controlRequest{Schema: channelSchema, Session: "uireview-aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Token: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", Command: "ui-browser", Request: request}); err != nil {
			t.Fatal("channel", err)
		}
	})
	t.Run("success_counts_and_immutable_repeated_stop", func(t *testing.T) {
		calls := 0
		p := &testPublisher{}
		d := &testDriver{execute: func(context.Context, string, jsonRequest, *string, int) (Operation, error) {
			return Operation{Counts: Counts{Images: 2, Elements: 7}, Epoch: pointer(4), Bundle: pointer("bundle-1")}, nil
		}, close: func(context.Context) []error { calls++; return nil }}
		s := readySession(t, d, p)
		v := s.Handle(context.Background(), "ui-capture", nil, nil)
		if v.Exit != 0 || v.Operation == nil || *v.Operation != 1 || v.Receipt == nil {
			t.Fatal(v)
		}
		var wait sync.WaitGroup
		for range 12 {
			wait.Go(func() { s.Stop(nil) })
		}
		wait.Wait()
		awaitSession(t, s)
		if calls != 1 || len(p.terminals) != 1 || p.terminals[0].Counts.Images != 2 || p.terminals[0].Counts.Elements != 7 {
			t.Fatal(calls, p.terminals)
		}
		first := s.Value("ui-review-stop")
		for range 3 {
			next := s.Handle(context.Background(), "ui-review-stop", nil, nil)
			if next.Exit != first.Exit || *next.State != "closed" || next.Receipt.Path != first.Receipt.Path {
				t.Fatal(next)
			}
		}
	})
	t.Run("rejected_action_does_not_allocate_operation", func(t *testing.T) {
		p := &testPublisher{}
		s := readySession(t, &testDriver{validate: func(context.Context, string, jsonRequest) error { return failure("session_mismatch") }, execute: func(context.Context, string, jsonRequest, *string, int) (Operation, error) {
			t.Error("rejected action executed")
			return Operation{}, nil
		}}, p)
		v := s.Handle(context.Background(), "ui-browser", nil, nil)
		if v.Exit != 2 || v.Operation != nil || len(p.receipts) != 0 {
			t.Fatal(v)
		}
	})
	t.Run("single_admission_and_stop_drains_before_terminal", func(t *testing.T) {
		entered, release := make(chan struct{}), make(chan struct{})
		p := &testPublisher{}
		d := &testDriver{execute: func(ctx context.Context, _ string, _ jsonRequest, _ *string, _ int) (Operation, error) {
			close(entered)
			<-ctx.Done()
			<-release
			return Operation{}, failure("interrupted")
		}, close: func(context.Context) []error { close(release); return nil }}
		s := readySession(t, d, p)
		done := make(chan Result, 1)
		go func() { done <- s.Handle(context.Background(), "ui-capture", nil, nil) }()
		<-entered
		waiting := s.Handle(context.Background(), "ui-analyze", nil, nil)
		if waiting.Exit != 4 || waiting.Operation != nil {
			t.Fatal(waiting)
		}
		s.Stop(nil)
		v := <-done
		awaitSession(t, s)
		if v.Exit != 15 || len(p.receipts) != 1 || len(p.terminals) != 1 || len(v.Private) != 0 {
			t.Fatal(v, p.receipts, p.terminals)
		}
	})
	t.Run("primary_precedes_independent_cleanup_failure", func(t *testing.T) {
		p := &testPublisher{}
		s := readySession(t, &testDriver{close: func(context.Context) []error {
			return []error{errors.New("failed close"), errors.New("another failed close")}
		}}, p)
		s.Stop(failure("startup_failed"))
		awaitSession(t, s)
		v := s.Value("ui-review-stop")
		if v.Exit != 3 || len(v.Failures) != 2 || v.Failures[0].Code != "startup_failed" || v.Failures[1].Code != "cleanup_failed" || p.terminals[0].Cleanup != "failed" {
			t.Fatal(v, p.terminals)
		}
	})
	t.Run("cleanup_only_failure_and_failed_publication", func(t *testing.T) {
		for _, publication := range []bool{false, true} {
			p := &testPublisher{}
			d := &testDriver{}
			if publication {
				p.terminalError = errors.New("publication")
			} else {
				d.close = func(context.Context) []error { return []error{errors.New("close")} }
			}
			s := readySession(t, d, p)
			s.Stop(nil)
			awaitSession(t, s)
			v := s.Value("ui-review-stop")
			expected := 12
			if publication {
				expected = 11
				if v.Receipt != nil {
					t.Fatal("failed publication claimed receipt")
				}
			}
			if v.Exit != expected || *v.State != "failed" {
				t.Fatal(v)
			}
		}
	})
	t.Run("late_preparation_failure_survives_stop", func(t *testing.T) {
		entered, release := make(chan struct{}), make(chan struct{})
		p := &testPublisher{}
		s := NewSession("session", "run", "seeded", &testDriver{prepare: func(context.Context) (*int, error) {
			close(entered)
			<-release
			return nil, contextual("build_failed", "build", "frontend", "child_failed", "build_web")
		}}, p)
		prepared := make(chan error, 1)
		go func() { prepared <- s.Prepare() }()
		<-entered
		s.Stop(failure("interrupted"))
		close(release)
		<-prepared
		awaitSession(t, s)
		v := s.Value("ui-review-stop")
		if v.Exit != 1 || v.Failures[0].Code != "build_failed" {
			t.Fatal(v)
		}
	})
	t.Run("operation_deadline_stops_owner", func(t *testing.T) {
		p := &testPublisher{}
		s := readySession(t, &testDriver{execute: func(ctx context.Context, _ string, _ jsonRequest, _ *string, _ int) (Operation, error) {
			<-ctx.Done()
			return Operation{}, ctx.Err()
		}}, p)
		s.operationTimeout = 10 * time.Millisecond
		v := s.Handle(context.Background(), "ui-capture", nil, nil)
		awaitSession(t, s)
		if v.Exit != 13 || s.Value("ui-review-stop").Exit != 13 {
			t.Fatal(v)
		}
	})
}
