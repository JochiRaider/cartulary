package httptestx

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"testing"
	"time"
)

const (
	defaultPollInterval = 25 * time.Millisecond
	pollBodyLimit       = 1 << 20
	// A fatal test server answers 503 service_unavailable with retryable true
	// and this reason. It is permanent for the process, so a poll stops and
	// reports it instead of waiting to its deadline; the server_fatal evidence
	// record names the cause (TH-HARNESS-REQ-863).
	fatalAdmissionReason = "extension_integrity_failure"
	// A server in ordinary shutdown never reopens admission either
	// (TH-HARNESS-REQ-864, Core 04 REQ-04-146a).
	shutdownAdmissionReason = "shutdown_in_progress"
)

// jobTerminalStatuses are the job states that never change again.
var jobTerminalStatuses = []string{"succeeded", "failed", "canceled"}

// PollOptions bound one poll of a job or other eventually-consistent resource
// (TH-HARNESS-REQ-864).
type PollOptions struct {
	// Timeout is the poll's own deadline, for a wait the test asserts. Zero
	// bounds the poll by the test's deadline less a settlement margin, which
	// is the right bound for a hang guard: load can slow a job without the
	// test failing, and a real hang still reports its last response before
	// the test binary's own timeout.
	Timeout time.Duration
	// Interval separates observations; zero selects 25 ms.
	Interval time.Duration
	// Client sends each request; nil selects http.DefaultClient.
	Client *http.Client
	// Request applies cookies, headers and other per-request options.
	Request []func(*http.Request)
	// LockWaits, when set, is the test database a lock-wait snapshot is taken
	// from when the deadline expires (RecordLockWaits).
	LockWaits *sql.DB
}

// Observation is one response a poll received.
type Observation struct {
	Status int
	Header http.Header
	Body   []byte
}

func (o Observation) String() string {
	return fmt.Sprintf("status %d body: %q", o.Status, reportedBody(o.Body))
}

// Poll sends GET target until done reports true for a 2xx response, then
// returns that response. It waits through an error response whose envelope
// says retryable: true, stops on any other error response or on done's error,
// and at its deadline fails with the last response it received. Waiting
// through a retryable response is part of this test's execution, not a retry
// of a product test.
func Poll(t testing.TB, target string, options PollOptions, done func(Observation) (bool, error)) Observation {
	t.Helper()
	ctx, cancel := testPollContext(t)
	defer cancel()
	observation, err := poll(ctx, target, options, done)
	if err != nil {
		t.Fatal(err)
	}
	return observation
}

// pollSettlement is left between a test-bounded poll's deadline and the
// test's, so the failure, its evidence and the test's cleanup all fit.
const pollSettlement = 30 * time.Second

// testPollContext bounds a poll by the test's deadline less pollSettlement
// (TH-HARNESS-REQ-864), never below one second. A test without a deadline
// keeps only its context.
func testPollContext(t testing.TB) (context.Context, context.CancelFunc) {
	if test, ok := t.(interface{ Deadline() (time.Time, bool) }); ok {
		if deadline, bounded := test.Deadline(); bounded {
			return context.WithTimeout(t.Context(), max(time.Until(deadline)-pollSettlement, time.Second))
		}
	}
	return context.WithCancel(t.Context())
}

// PollJob polls the job origin/api/v1/jobs/{jobID} until its status is one of
// want, and returns the job resource. Each response must be a 200 success
// envelope. A terminal status outside want (succeeded, failed or canceled)
// fails at once and reports the job.
func PollJob(t testing.TB, origin, jobID string, options PollOptions, want ...string) map[string]any {
	t.Helper()
	ctx, cancel := testPollContext(t)
	defer cancel()
	job, err := pollJob(ctx, origin, jobID, options, want)
	if err != nil {
		t.Fatal(err)
	}
	return job
}

func pollJob(ctx context.Context, origin, jobID string, options PollOptions, want []string) (map[string]any, error) {
	if len(want) == 0 {
		return nil, errors.New("job poll requires at least one wanted status")
	}
	var job map[string]any
	_, err := poll(ctx, origin+"/api/v1/jobs/"+url.PathEscape(jobID), options, func(o Observation) (bool, error) {
		data, err := readSuccessData(&http.Response{StatusCode: o.Status, Header: o.Header, Body: io.NopCloser(bytes.NewReader(o.Body))}, http.StatusOK)
		if err != nil {
			return false, err
		}
		status, _ := data["status"].(string)
		if slices.Contains(want, status) {
			job = data
			return true, nil
		}
		if slices.Contains(jobTerminalStatuses, status) {
			return false, fmt.Errorf("job %s reached terminal status %q while waiting for %q", jobID, status, want)
		}
		return false, nil
	})
	return job, err
}

func poll(ctx context.Context, target string, options PollOptions, done func(Observation) (bool, error)) (Observation, error) {
	if options.Timeout < 0 || done == nil {
		return Observation{}, errors.New("invalid poll options")
	}
	interval := options.Interval
	if interval <= 0 {
		interval = defaultPollInterval
	}
	client := options.Client
	if client == nil {
		client = http.DefaultClient
	}
	if options.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, options.Timeout)
		defer cancel()
	}
	// Failure messages name only the resource path, not the test origin.
	path := target
	if parsed, err := url.Parse(target); err == nil && parsed.Path != "" {
		path = parsed.Path
	}
	started := time.Now()
	var last *Observation
	lastResponse := func() string {
		if last == nil {
			return "no response received"
		}
		return "last response " + last.String()
	}
	for {
		observation, err := observe(ctx, client, target, options.Request)
		if err != nil {
			if ctx.Err() != nil {
				break
			}
			return Observation{}, fmt.Errorf("poll %s: request failed: %v; %s", path, err, lastResponse())
		}
		last = &observation
		if observation.Status < 200 || observation.Status > 299 {
			if !retryableEnvelope(observation.Body) {
				return observation, fmt.Errorf("poll %s stopped on a non-retryable response: %s", path, observation)
			}
		} else if finished, err := done(observation); err != nil {
			return observation, fmt.Errorf("poll %s: %v; %s", path, err, lastResponse())
		} else if finished {
			return observation, nil
		}
		timer := time.NewTimer(interval)
		select {
		case <-ctx.Done():
			timer.Stop()
		case <-timer.C:
		}
		if ctx.Err() != nil {
			break
		}
	}
	message := fmt.Sprintf("poll %s did not finish within %s; %s", path, time.Since(started).Round(time.Millisecond), lastResponse())
	if options.LockWaits != nil {
		message += "; " + describeLockWaits(options.LockWaits)
	}
	var result Observation
	if last != nil {
		result = *last
	}
	return result, errors.New(message)
}

func observe(ctx context.Context, client *http.Client, target string, options []func(*http.Request)) (Observation, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, target, nil)
	if err != nil {
		return Observation{}, err
	}
	request.Header.Set("Content-Type", "application/json")
	for _, option := range options {
		option(request)
	}
	response, err := client.Do(request)
	if err != nil {
		return Observation{}, err
	}
	defer response.Body.Close()
	body, err := io.ReadAll(io.LimitReader(response.Body, pollBodyLimit+1))
	if err != nil {
		return Observation{}, err
	}
	if len(body) > pollBodyLimit {
		return Observation{}, fmt.Errorf("response body exceeds %d bytes", pollBodyLimit)
	}
	return Observation{Status: response.StatusCode, Header: response.Header.Clone(), Body: body}, nil
}

// retryableEnvelope reports whether an error response says retryable: true,
// excluding the permanent admission refusals of a fatal or shutting-down
// test server.
func retryableEnvelope(body []byte) bool {
	var envelope struct {
		Error struct {
			Retryable bool `json:"retryable"`
			Details   struct {
				ReasonCode string `json:"reason_code"`
			} `json:"details"`
		} `json:"error"`
	}
	if json.Unmarshal(body, &envelope) != nil {
		return false
	}
	reason := envelope.Error.Details.ReasonCode
	return envelope.Error.Retryable && reason != fatalAdmissionReason && reason != shutdownAdmissionReason
}
