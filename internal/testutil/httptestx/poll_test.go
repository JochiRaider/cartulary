package httptestx

import (
	"context"
	"database/sql"
	"io"
	"net/http"
	stdhttptest "net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

type pollStep struct {
	status int
	body   string
}

// pollServer answers each request with the next step and repeats the last.
func pollServer(t *testing.T, steps ...pollStep) (*stdhttptest.Server, *atomic.Int32) {
	t.Helper()
	var calls atomic.Int32
	server := stdhttptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		step := steps[min(int(calls.Add(1))-1, len(steps)-1)]
		w.Header().Set(httpapi.RequestIDHeader, "request-1")
		w.WriteHeader(step.status)
		_, _ = io.WriteString(w, step.body)
	}))
	t.Cleanup(server.Close)
	return server, &calls
}

func jobStep(status string) pollStep {
	return pollStep{http.StatusOK, `{"data":{"job_id":"job-1","status":"` + status + `"},"meta":{"request_id":"request-1"}}`}
}

const (
	retryableUnavailable = `{"error":{"code":"service_unavailable","message":"service unavailable","status":503,"request_id":"request-1","retryable":true,"details":{"reason_code":"extension_publication_pending"}}}`
	nonRetryableConflict = `{"error":{"code":"invalid_import_state","message":"conflict","status":409,"request_id":"request-1","retryable":false,"details":{"csrf_token":"sentinel-csrf-value"}}}`
)

// TestPollKeepsEvidence covers TH-HARNESS-REQ-864 and AC-159: a poll waits
// through a retryable 503, including a paused one, stops on a non-retryable
// error and on a closing server's permanent refusal, fatal or shutting down,
// fails a job that reaches another terminal status, and reports the last
// response, redacted and bounded, at its deadline.
func TestPollKeepsEvidence(t *testing.T) {
	fatalUnavailable := strings.Replace(retryableUnavailable, "extension_publication_pending", "extension_integrity_failure", 1)
	pausedUnavailable := strings.Replace(retryableUnavailable, "extension_publication_pending", "admission_paused", 1)
	stoppingUnavailable := strings.Replace(retryableUnavailable, "extension_publication_pending", "shutdown_in_progress", 1)
	for _, tc := range []struct {
		name     string
		steps    []pollStep
		want     []string
		timeout  time.Duration
		calls    int32
		status   string
		failure  []string
		redacted []string
	}{
		{name: "waits through a retryable 503", steps: []pollStep{{503, retryableUnavailable}, {503, retryableUnavailable}, jobStep("running"), jobStep("succeeded")}, want: []string{"succeeded"}, calls: 4, status: "succeeded"},
		{name: "returns a wanted terminal status", steps: []pollStep{jobStep("queued"), jobStep("canceled")}, want: []string{"canceled"}, calls: 2, status: "canceled"},
		{name: "returns any wanted status", steps: []pollStep{jobStep("failed")}, want: jobTerminalStatuses, calls: 1, status: "failed"},
		{name: "stops on a non-retryable error", steps: []pollStep{jobStep("running"), {409, nonRetryableConflict}, jobStep("succeeded")}, want: []string{"succeeded"}, calls: 2,
			failure: []string{"stopped on a non-retryable response", "status 409", "invalid_import_state", "csrf_token"}, redacted: []string{"sentinel-csrf-value"}},
		{name: "stops on a fatal server's refusal", steps: []pollStep{{503, fatalUnavailable}, jobStep("succeeded")}, want: []string{"succeeded"}, calls: 1,
			failure: []string{"stopped on a non-retryable response", "status 503", "extension_integrity_failure"}},
		{name: "waits through paused admission", steps: []pollStep{{503, pausedUnavailable}, jobStep("succeeded")}, want: []string{"succeeded"}, calls: 2, status: "succeeded"},
		{name: "stops on a shutting-down server's refusal", steps: []pollStep{{503, stoppingUnavailable}, jobStep("succeeded")}, want: []string{"succeeded"}, calls: 1,
			failure: []string{"stopped on a non-retryable response", "status 503", "shutdown_in_progress"}},
		{name: "fails a job that reaches another terminal status", steps: []pollStep{jobStep("running"), jobStep("failed")}, want: []string{"succeeded"}, calls: 2,
			failure: []string{`terminal status "failed"`, "last response status 200", "job-1"}},
		{name: "fails an invalid success envelope", steps: []pollStep{{200, `{"data":{"status":"running"}}`}}, want: []string{"succeeded"}, calls: 1,
			failure: []string{"meta", "last response status 200"}},
		{name: "reports the last response at its deadline", steps: []pollStep{jobStep("queued"), jobStep("running")}, want: []string{"succeeded"}, timeout: 150 * time.Millisecond,
			failure: []string{"did not finish within", "last response status 200", `\"status\":\"running\"`}},
		{name: "reports a retryable response at its deadline", steps: []pollStep{jobStep("running"), {503, retryableUnavailable}}, want: []string{"succeeded"}, timeout: 150 * time.Millisecond,
			failure: []string{"did not finish within", "last response status 503", "extension_publication_pending"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			server, calls := pollServer(t, tc.steps...)
			timeout := tc.timeout
			if timeout == 0 {
				timeout = 5 * time.Second
			}
			job, err := pollJob(t.Context(), server.URL, "job-1", PollOptions{Timeout: timeout, Interval: time.Millisecond}, tc.want)
			if tc.failure == nil {
				if err != nil || job["status"] != tc.status {
					t.Fatalf("poll = %v, %v; want %s", job, err, tc.status)
				}
			} else {
				if err == nil {
					t.Fatalf("poll returned %v; want failure", job)
				}
				for _, text := range tc.failure {
					if !strings.Contains(err.Error(), text) {
						t.Fatalf("failure %q does not report %q", err, text)
					}
				}
				for _, secret := range tc.redacted {
					if strings.Contains(err.Error(), secret) {
						t.Fatalf("failure %q leaks %q", err, secret)
					}
				}
			}
			if tc.calls != 0 && calls.Load() != tc.calls {
				t.Fatalf("poll made %d requests, want %d", calls.Load(), tc.calls)
			}
		})
	}
	t.Run("the public helper returns the job", func(t *testing.T) {
		server, _ := pollServer(t, pollStep{503, retryableUnavailable}, jobStep("succeeded"))
		if job := PollJob(t, server.URL, "job-1", PollOptions{Timeout: 5 * time.Second, Interval: time.Millisecond}, "succeeded"); job["job_id"] != "job-1" {
			t.Fatalf("job = %v", job)
		}
	})
	t.Run("a generic poll sees each successful response", func(t *testing.T) {
		server, _ := pollServer(t, pollStep{503, retryableUnavailable}, pollStep{200, "pending"}, pollStep{204, ""})
		seen := []int{}
		last := Poll(t, server.URL, PollOptions{Timeout: 5 * time.Second, Interval: time.Millisecond}, func(o Observation) (bool, error) {
			seen = append(seen, o.Status)
			return o.Status == http.StatusNoContent, nil
		})
		if last.Status != http.StatusNoContent || len(seen) != 2 {
			t.Fatalf("poll observed %v, returned %d", seen, last.Status)
		}
	})
	t.Run("a deadline with a database names its lock-wait snapshot", func(t *testing.T) {
		server, _ := pollServer(t, jobStep("running"))
		db, err := sql.Open("pgx", "postgres://cartulary@127.0.0.1:1/unreachable?connect_timeout=1")
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = db.Close() })
		_, err = pollJob(t.Context(), server.URL, "job-1", PollOptions{Timeout: 50 * time.Millisecond, Interval: time.Millisecond, LockWaits: db}, []string{"succeeded"})
		if err == nil || !strings.Contains(err.Error(), "did not finish within") || !strings.Contains(err.Error(), "lock-wait snapshot unavailable") {
			t.Fatalf("deadline failure = %v", err)
		}
	})
	t.Run("an unreachable resource fails without waiting", func(t *testing.T) {
		server, _ := pollServer(t, jobStep("running"))
		server.Close()
		ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
		defer cancel()
		if _, err := pollJob(ctx, server.URL, "job-1", PollOptions{Timeout: 5 * time.Second}, []string{"succeeded"}); err == nil || !strings.Contains(err.Error(), "request failed") || ctx.Err() != nil {
			t.Fatalf("transport failure = %v", err)
		}
	})
}

type deadlineTB struct {
	testing.TB
	deadline time.Time
	bounded  bool
}

func (d deadlineTB) Deadline() (time.Time, bool) { return d.deadline, d.bounded }

// A zero-timeout poll is a hang guard: it takes the test's deadline less the
// settlement margin, never below one second, and keeps only the test's
// context when the test has no deadline (TH-HARNESS-REQ-864).
func TestPollDefaultsToTheTestDeadline(t *testing.T) {
	for _, tc := range []struct {
		name      string
		remaining time.Duration
		bounded   bool
		want      time.Duration
	}{
		{name: "long", remaining: 10 * time.Minute, bounded: true, want: 10*time.Minute - pollSettlement},
		{name: "inside_settlement", remaining: 10 * time.Second, bounded: true, want: time.Second},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx, cancel := testPollContext(deadlineTB{TB: t, deadline: time.Now().Add(tc.remaining), bounded: tc.bounded})
			defer cancel()
			deadline, ok := ctx.Deadline()
			if !ok {
				t.Fatal("test-bounded poll has no deadline")
			}
			if got := time.Until(deadline); got > tc.want || got < tc.want-5*time.Second {
				t.Fatalf("poll deadline in %s, want about %s", got, tc.want)
			}
		})
	}
	ctx, cancel := testPollContext(deadlineTB{TB: t})
	defer cancel()
	if _, ok := ctx.Deadline(); ok {
		t.Fatal("a test without a deadline gained one")
	}
	cancel()
	if ctx.Err() == nil {
		t.Fatal("poll context ignores its cancellation")
	}
}
