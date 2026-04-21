package testcontainersx

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestStartWithRetryFailsFastOnPreflightError(t *testing.T) {
	t.Parallel()

	preflightCalls := 0
	startCalls := 0

	_, err := StartWithRetry(context.Background(), StartConfig{
		Service: "postgres testcontainer",
		Image:   "postgres:16-alpine",
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
}

func TestStartWithRetryRetriesTransientDockerStartupError(t *testing.T) {
	t.Parallel()

	startCalls := 0
	sleepCalls := 0

	value, err := StartWithRetry(context.Background(), StartConfig{
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

func TestStartWithRetryDoesNotRetryLogicalReadinessFailure(t *testing.T) {
	t.Parallel()

	startCalls := 0

	_, err := StartWithRetry(context.Background(), StartConfig{
		Service: "postgres testcontainer",
		Image:   "postgres:16-alpine",
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
}

func TestStartWithRetryFormatsFinalFailureWithContext(t *testing.T) {
	t.Parallel()

	_, err := StartWithRetry(context.Background(), StartConfig{
		Service: "minio testcontainer",
		Image:   "minio/minio:RELEASE.2025-09-07T16-13-09Z",
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
		"after 2 attempt(s)",
		"connection refused",
	} {
		if !strings.Contains(message, want) {
			t.Fatalf("expected error message to contain %q, got %q", want, message)
		}
	}
}
