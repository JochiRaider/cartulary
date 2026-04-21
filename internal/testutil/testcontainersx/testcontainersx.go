package testcontainersx

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"syscall"
	"time"

	"github.com/moby/moby/client"
	testcontainers "github.com/testcontainers/testcontainers-go"
)

const (
	DefaultPreflightTimeout = 3 * time.Second
	DefaultRetryBackoff     = 250 * time.Millisecond
	DefaultMaxAttempts      = 2
)

type StartConfig struct {
	Service          string
	Image            string
	PreflightTimeout time.Duration
	RetryBackoff     time.Duration
	MaxAttempts      int
	Preflight        func(context.Context) (string, error)
	Sleep            func(context.Context, time.Duration) error
}

func StartWithRetry[T any](ctx context.Context, config StartConfig, startup func(context.Context) (T, error)) (T, error) {
	var zero T

	service := strings.TrimSpace(config.Service)
	if service == "" {
		return zero, errors.New("testcontainersx: service is required")
	}
	image := strings.TrimSpace(config.Image)
	if image == "" {
		return zero, errors.New("testcontainersx: image is required")
	}

	preflight := config.Preflight
	if preflight == nil {
		preflight = defaultDockerPreflight
	}
	sleep := config.Sleep
	if sleep == nil {
		sleep = sleepWithContext
	}

	preflightTimeout := config.PreflightTimeout
	if preflightTimeout <= 0 {
		preflightTimeout = DefaultPreflightTimeout
	}

	retryBackoff := config.RetryBackoff
	if retryBackoff <= 0 {
		retryBackoff = DefaultRetryBackoff
	}

	maxAttempts := config.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = DefaultMaxAttempts
	}

	preflightCtx, cancel := context.WithTimeout(ctx, preflightTimeout)
	endpoint, err := preflight(preflightCtx)
	cancel()
	if err != nil {
		return zero, formatFailure("docker preflight", service, image, endpoint, 0, err)
	}

	for attempt := 1; attempt <= maxAttempts; attempt++ {
		value, startErr := startup(ctx)
		if startErr == nil {
			return value, nil
		}

		if attempt >= maxAttempts || !isTransientDockerStartupError(startErr) {
			return zero, formatFailure("start", service, image, endpoint, attempt, startErr)
		}

		if err := sleep(ctx, retryBackoff); err != nil {
			return zero, formatFailure("start", service, image, endpoint, attempt, fmt.Errorf("wait for retry backoff: %w", err))
		}
	}

	return zero, formatFailure("start", service, image, endpoint, maxAttempts, errors.New("startup failed without an attributed cause"))
}

func defaultDockerPreflight(ctx context.Context) (string, error) {
	cli, err := testcontainers.NewDockerClientWithOpts(ctx)
	if err != nil {
		return "", err
	}
	defer cli.Close()

	endpoint := cli.DaemonHost()
	if _, err := cli.Ping(ctx, client.PingOptions{NegotiateAPIVersion: true}); err != nil {
		return endpoint, err
	}

	return endpoint, nil
}

func sleepWithContext(ctx context.Context, delay time.Duration) error {
	if delay <= 0 {
		return nil
	}

	timer := time.NewTimer(delay)
	defer timer.Stop()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func isTransientDockerStartupError(err error) bool {
	if err == nil {
		return false
	}

	lower := strings.ToLower(err.Error())
	if errors.Is(err, io.EOF) || errors.Is(err, syscall.ECONNRESET) || errors.Is(err, syscall.ECONNREFUSED) || errors.Is(err, syscall.EPIPE) {
		return true
	}

	if strings.Contains(lower, "unexpected eof") {
		return true
	}

	if errors.Is(err, context.DeadlineExceeded) {
		return hasDockerTransportContext(lower)
	}

	return hasDockerTransportContext(lower)
}

func hasDockerTransportContext(lower string) bool {
	return strings.Contains(lower, "docker.sock") ||
		strings.Contains(lower, "/containers/") ||
		strings.Contains(lower, "connection reset by peer") ||
		strings.Contains(lower, "connection refused") ||
		strings.Contains(lower, "broken pipe") ||
		strings.Contains(lower, "unexpected eof") ||
		strings.Contains(lower, "eof") ||
		strings.Contains(lower, "client.ping") ||
		strings.Contains(lower, "daemon") ||
		strings.Contains(lower, "http://%2fvar%2frun%2fdocker.sock") ||
		strings.Contains(lower, "http://docker.sock")
}

func formatFailure(operation string, service string, image string, endpoint string, attempts int, err error) error {
	endpointLabel := strings.TrimSpace(endpoint)
	if endpointLabel == "" {
		endpointLabel = "unresolved"
	}

	if attempts > 0 {
		return fmt.Errorf("%s %s (image %s, docker endpoint %s) failed after %d attempt(s): %w", operation, service, image, endpointLabel, attempts, err)
	}

	return fmt.Errorf("%s %s (image %s, docker endpoint %s) failed: %w", operation, service, image, endpointLabel, err)
}
