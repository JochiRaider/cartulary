package testcontainersx

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"
	"time"
)

func TestStartWithRetryFailsFastOnPreflightError(t *testing.T) {
	t.Parallel()

	preflightCalls := 0
	startCalls := 0

	_, err := StartWithRetry(context.Background(), StartConfig{
		Service:      "postgres testcontainer",
		Image:        "postgres:16-alpine",
		MaxAttempts:  DefaultMaxAttempts,
		RetryBackoff: DefaultRetryBackoff,
		Preflight: func(context.Context) (string, error) {
			preflightCalls++
			return "unix:///var/run/docker.sock", errors.New("docker ping refused")
		},
	}, func(context.Context) (int, error) {
		startCalls++
		return 0, nil
	})
	if err == nil {
		t.Fatal("expected preflight failure")
	}
	if preflightCalls != 1 {
		t.Fatalf("expected one preflight call, got %d", preflightCalls)
	}
	if startCalls != 0 {
		t.Fatalf("expected no startup attempts after preflight failure, got %d", startCalls)
	}
	if !strings.Contains(err.Error(), "docker preflight postgres testcontainer") {
		t.Fatalf("expected preflight failure message, got %q", err)
	}

	var startFailure *StartFailure
	if !errors.As(err, &startFailure) {
		t.Fatalf("expected StartFailure, got %T", err)
	}
	if startFailure.Operation != "docker preflight" {
		t.Fatalf("unexpected operation: got %q", startFailure.Operation)
	}
	if startFailure.AttemptsStarted != 0 {
		t.Fatalf("unexpected attempt count: got %d", startFailure.AttemptsStarted)
	}
	if startFailure.MaxAttempts != DefaultMaxAttempts {
		t.Fatalf("unexpected max attempts: got %d want %d", startFailure.MaxAttempts, DefaultMaxAttempts)
	}
	if startFailure.Retryable {
		t.Fatal("preflight failure must not be retryable")
	}
}

func TestStartWithRetryRetriesTransientDockerStartupError(t *testing.T) {
	t.Parallel()

	startCalls := 0
	sleepCalls := 0

	value, err := StartWithRetry(context.Background(), StartConfig{
		Service:      "postgres testcontainer",
		Image:        "postgres:16-alpine",
		MaxAttempts:  DefaultMaxAttempts,
		RetryBackoff: DefaultRetryBackoff,
		Preflight: func(context.Context) (string, error) {
			return "unix:///var/run/docker.sock", nil
		},
		Sleep: func(context.Context, time.Duration) error {
			sleepCalls++
			return nil
		},
	}, func(context.Context) (string, error) {
		startCalls++
		if startCalls == 1 {
			return "", errors.New(`wait until ready: get state: Get "http://%2Fvar%2Frun%2Fdocker.sock/v1.51/containers/id/json": context deadline exceeded`)
		}
		return "started", nil
	})
	if err != nil {
		t.Fatalf("expected retry success, got %v", err)
	}
	if value != "started" {
		t.Fatalf("unexpected startup result %q", value)
	}
	if startCalls != 2 {
		t.Fatalf("expected two startup attempts, got %d", startCalls)
	}
	if sleepCalls != 1 {
		t.Fatalf("expected one retry backoff, got %d", sleepCalls)
	}
}

func TestStartWithRetryDoesNotRetryWithoutDeclaredAttempts(t *testing.T) {
	t.Parallel()

	startCalls := 0
	sleepCalls := 0
	retryableErr := errors.New(`wait until ready: get state: Get "http://%2Fvar%2Frun%2Fdocker.sock/v1.51/containers/id/json": context deadline exceeded`)

	_, err := StartWithRetry(context.Background(), StartConfig{
		Service: "postgres testcontainer",
		Image:   "postgres:16-alpine",
		Preflight: func(context.Context) (string, error) {
			return "unix:///var/run/docker.sock", nil
		},
		Sleep: func(context.Context, time.Duration) error {
			sleepCalls++
			return nil
		},
	}, func(context.Context) (string, error) {
		startCalls++
		return "", retryableErr
	})
	if err == nil {
		t.Fatal("expected startup failure")
	}
	if startCalls != 1 {
		t.Fatalf("retry must not be implicit, got %d startup attempts", startCalls)
	}
	if sleepCalls != 0 {
		t.Fatalf("retry backoff must not run without declared attempts, got %d", sleepCalls)
	}

	var startFailure *StartFailure
	if !errors.As(err, &startFailure) {
		t.Fatalf("expected StartFailure, got %T", err)
	}
	if !startFailure.Retryable {
		t.Fatal("failure should still be classified retryable")
	}
	if startFailure.AttemptsStarted != 1 || startFailure.MaxAttempts != 1 {
		t.Fatalf("unexpected attempts: got %d/%d", startFailure.AttemptsStarted, startFailure.MaxAttempts)
	}
}

func TestStartWithRetryEmitsObserverEvents(t *testing.T) {
	t.Parallel()

	startCalls := 0
	var events []StartEvent

	value, err := StartWithRetry(context.Background(), StartConfig{
		Service:      "postgres testcontainer",
		Image:        "postgres:16-alpine",
		MaxAttempts:  DefaultMaxAttempts,
		RetryBackoff: DefaultRetryBackoff,
		Preflight: func(context.Context) (string, error) {
			return "unix:///var/run/docker.sock", nil
		},
		Sleep: func(context.Context, time.Duration) error {
			return nil
		},
		Observer: func(event StartEvent) {
			events = append(events, event)
		},
	}, func(context.Context) (string, error) {
		startCalls++
		if startCalls == 1 {
			return "", errors.New(`wait until ready: get state: Get "http://%2Fvar%2Frun%2Fdocker.sock/v1.51/containers/id/json": context deadline exceeded`)
		}
		return "started", nil
	})
	if err != nil {
		t.Fatalf("expected retry success, got %v", err)
	}
	if value != "started" {
		t.Fatalf("unexpected value: %q", value)
	}

	got := eventTypes(events)
	want := strings.Join([]string{
		StartEventPreflightStart,
		StartEventPreflightEnd,
		StartEventAttemptStart,
		StartEventAttemptEnd,
		StartEventRetryScheduled,
		StartEventAttemptStart,
		StartEventAttemptEnd,
	}, ",")
	if got != want {
		t.Fatalf("unexpected observer events: got %s want %s", got, want)
	}
	if events[3].Attempt != 1 || !events[3].Retryable || events[3].Status != "fail" || events[3].Err == nil {
		t.Fatalf("unexpected failed attempt event: %#v", events[3])
	}
	if events[6].Attempt != 2 || events[6].Status != "pass" || events[6].Duration < 0 {
		t.Fatalf("unexpected successful attempt event: %#v", events[6])
	}
}

func TestStartWithRetryRetriesCustomClassifiedStartupError(t *testing.T) {
	t.Parallel()

	readinessErr := errors.New("wait for minio readiness: minio did not become ready via authenticated api: context deadline exceeded")
	startCalls := 0
	sleepCalls := 0

	value, err := StartWithRetry(context.Background(), StartConfig{
		Service:      "minio testcontainer",
		Image:        "minio/minio:RELEASE.2025-09-07T16-13-09Z",
		MaxAttempts:  DefaultMaxAttempts,
		RetryBackoff: DefaultRetryBackoff,
		Preflight: func(context.Context) (string, error) {
			return "unix:///var/run/docker.sock", nil
		},
		Retryable: func(err error) bool {
			return errors.Is(err, readinessErr)
		},
		Sleep: func(context.Context, time.Duration) error {
			sleepCalls++
			return nil
		},
	}, func(context.Context) (string, error) {
		startCalls++
		if startCalls == 1 {
			return "", readinessErr
		}
		return "started", nil
	})
	if err != nil {
		t.Fatalf("expected retry success, got %v", err)
	}
	if value != "started" {
		t.Fatalf("unexpected startup result %q", value)
	}
	if startCalls != 2 {
		t.Fatalf("expected two startup attempts, got %d", startCalls)
	}
	if sleepCalls != 1 {
		t.Fatalf("expected one retry backoff, got %d", sleepCalls)
	}
}

func TestStartWithRetryReportsCustomRetryMetadataAfterExhaustion(t *testing.T) {
	t.Parallel()

	readinessErr := errors.New("wait for minio readiness: context deadline exceeded")
	startCalls := 0

	_, err := StartWithRetry(context.Background(), StartConfig{
		Service:      "minio testcontainer",
		Image:        "minio/minio:RELEASE.2025-09-07T16-13-09Z",
		MaxAttempts:  3,
		RetryBackoff: DefaultRetryBackoff,
		Preflight: func(context.Context) (string, error) {
			return "unix:///var/run/docker.sock", nil
		},
		Retryable: func(err error) bool {
			return errors.Is(err, readinessErr)
		},
		Sleep: func(context.Context, time.Duration) error {
			return nil
		},
	}, func(context.Context) (int, error) {
		startCalls++
		return 0, readinessErr
	})
	if err == nil {
		t.Fatal("expected exhausted startup failure")
	}
	if startCalls != 3 {
		t.Fatalf("expected three startup attempts, got %d", startCalls)
	}

	var startFailure *StartFailure
	if !errors.As(err, &startFailure) {
		t.Fatalf("expected StartFailure, got %T", err)
	}
	if !startFailure.Retryable {
		t.Fatal("expected custom readiness failure to be marked retryable")
	}
	if startFailure.AttemptsStarted != 3 || startFailure.MaxAttempts != 3 {
		t.Fatalf("unexpected attempts: got %d/%d", startFailure.AttemptsStarted, startFailure.MaxAttempts)
	}
	if startFailure.Cause != readinessErr {
		t.Fatalf("expected original readiness cause to be preserved, got %#v", startFailure.Cause)
	}
}

func TestStartWithRetryReturnsOriginalCauseWhenContextExpiresDuringBackoff(t *testing.T) {
	t.Parallel()

	startCalls := 0
	sleepCalls := 0

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
	defer cancel()

	retryableErr := errors.New(`wait until ready: get state: Get "http://%2Fvar%2Frun%2Fdocker.sock/v1.51/containers/id/json": context deadline exceeded`)

	_, err := StartWithRetry(ctx, StartConfig{
		Service:      "minio testcontainer",
		Image:        "minio/minio:RELEASE.2025-09-07T16-13-09Z",
		MaxAttempts:  DefaultMaxAttempts,
		RetryBackoff: DefaultRetryBackoff,
		Preflight: func(context.Context) (string, error) {
			return "unix:///var/run/docker.sock", nil
		},
		Sleep: func(ctx context.Context, _ time.Duration) error {
			sleepCalls++
			<-ctx.Done()
			return ctx.Err()
		},
		Observer: func(event StartEvent) {
			if event.Type == StartEventRetryBlocked && !event.RetryBlockedByContext {
				t.Fatalf("retry-blocked event must mark context blocking: %#v", event)
			}
		},
	}, func(context.Context) (int, error) {
		startCalls++
		return 0, retryableErr
	})
	if err == nil {
		t.Fatal("expected retry-blocked failure")
	}
	if startCalls != 1 {
		t.Fatalf("expected one startup attempt before context expiry, got %d", startCalls)
	}
	if sleepCalls != 1 {
		t.Fatalf("expected one retry backoff wait, got %d", sleepCalls)
	}

	var startFailure *StartFailure
	if !errors.As(err, &startFailure) {
		t.Fatalf("expected StartFailure, got %T", err)
	}
	if !startFailure.Retryable {
		t.Fatal("expected retryable startup classification")
	}
	if !startFailure.RetryBlockedByContext {
		t.Fatal("expected retry to be blocked by context")
	}
	if startFailure.Cause != retryableErr {
		t.Fatalf("expected original startup cause to be preserved, got %#v", startFailure.Cause)
	}
	if startFailure.AttemptsStarted != 1 || startFailure.MaxAttempts != DefaultMaxAttempts {
		t.Fatalf("unexpected attempts: got %d/%d", startFailure.AttemptsStarted, startFailure.MaxAttempts)
	}

	message := err.Error()
	for _, want := range []string{
		"start minio testcontainer",
		"image minio/minio:RELEASE.2025-09-07T16-13-09Z",
		"docker endpoint unix:///var/run/docker.sock",
		"after 1/2 attempt(s)",
		"retry blocked by context: context deadline exceeded",
		`Get "http://%2Fvar%2Frun%2Fdocker.sock`,
	} {
		if !strings.Contains(message, want) {
			t.Fatalf("expected error message to contain %q, got %q", want, message)
		}
	}
}

func TestStartWithRetryBoundsRetryableStartupAttempt(t *testing.T) {
	t.Parallel()

	startCalls := 0
	sleepCalls := 0

	_, err := StartWithRetry(context.Background(), StartConfig{
		Service:        "postgres testcontainer",
		Image:          "postgres:16-alpine",
		AttemptTimeout: 20 * time.Millisecond,
		MaxAttempts:    DefaultMaxAttempts,
		RetryBackoff:   DefaultRetryBackoff,
		Preflight: func(context.Context) (string, error) {
			return "unix:///var/run/docker.sock", nil
		},
		Sleep: func(context.Context, time.Duration) error {
			sleepCalls++
			return nil
		},
	}, func(ctx context.Context) (int, error) {
		startCalls++
		<-ctx.Done()
		return 0, fmt.Errorf(`wait until ready: get state: Get "http://%s/containers/id/json": %w`, "%2Fvar%2Frun%2Fdocker.sock", ctx.Err())
	})
	if err == nil {
		t.Fatal("expected bounded startup failure")
	}
	if startCalls != DefaultMaxAttempts {
		t.Fatalf("expected bounded retries for each attempt, got %d", startCalls)
	}
	if sleepCalls != DefaultMaxAttempts-1 {
		t.Fatalf("expected one retry backoff between attempts, got %d", sleepCalls)
	}

	var startFailure *StartFailure
	if !errors.As(err, &startFailure) {
		t.Fatalf("expected StartFailure, got %T", err)
	}
	if !startFailure.Retryable {
		t.Fatal("bounded docker-context startup failure should be retryable")
	}
	if startFailure.AttemptsStarted != DefaultMaxAttempts {
		t.Fatalf("unexpected attempts: got %d", startFailure.AttemptsStarted)
	}
}

func eventTypes(events []StartEvent) string {
	types := make([]string, 0, len(events))
	for _, event := range events {
		types = append(types, event.Type)
	}
	return strings.Join(types, ",")
}

func TestStartWithRetryDoesNotRetryLogicalReadinessFailure(t *testing.T) {
	t.Parallel()

	startCalls := 0

	_, err := StartWithRetry(context.Background(), StartConfig{
		Service:      "postgres testcontainer",
		Image:        "postgres:16-alpine",
		MaxAttempts:  DefaultMaxAttempts,
		RetryBackoff: DefaultRetryBackoff,
		Preflight: func(context.Context) (string, error) {
			return "unix:///var/run/docker.sock", nil
		},
	}, func(context.Context) (int, error) {
		startCalls++
		return 0, errors.New("wait for postgres readiness: postgres did not become ready: context deadline exceeded")
	})
	if err == nil {
		t.Fatal("expected readiness failure")
	}
	if startCalls != 1 {
		t.Fatalf("expected one startup attempt, got %d", startCalls)
	}

	var startFailure *StartFailure
	if !errors.As(err, &startFailure) {
		t.Fatalf("expected StartFailure, got %T", err)
	}
	if startFailure.Retryable {
		t.Fatal("logical readiness failure must not be retryable")
	}
	if startFailure.AttemptsStarted != 1 || startFailure.MaxAttempts != DefaultMaxAttempts {
		t.Fatalf("unexpected attempts: got %d/%d", startFailure.AttemptsStarted, startFailure.MaxAttempts)
	}
}

func TestStartWithRetryFormatsFinalFailureWithContext(t *testing.T) {
	t.Parallel()

	_, err := StartWithRetry(context.Background(), StartConfig{
		Service:      "minio testcontainer",
		Image:        "minio/minio:RELEASE.2025-09-07T16-13-09Z",
		MaxAttempts:  DefaultMaxAttempts,
		RetryBackoff: DefaultRetryBackoff,
		Preflight: func(context.Context) (string, error) {
			return "unix:///var/run/docker.sock", nil
		},
		Sleep: func(context.Context, time.Duration) error {
			return nil
		},
	}, func(context.Context) (string, error) {
		return "", errors.New("docker.sock connection refused")
	})
	if err == nil {
		t.Fatal("expected final startup failure")
	}

	message := err.Error()
	for _, want := range []string{
		"start minio testcontainer",
		"image minio/minio:RELEASE.2025-09-07T16-13-09Z",
		"docker endpoint unix:///var/run/docker.sock",
		"after 2/2 attempt(s)",
		"connection refused",
	} {
		if !strings.Contains(message, want) {
			t.Fatalf("expected error message to contain %q, got %q", want, message)
		}
	}
}
