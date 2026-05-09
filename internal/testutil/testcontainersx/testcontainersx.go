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

const (
	StartEventPreflightStart = "preflight-start"
	StartEventPreflightEnd   = "preflight-end"
	StartEventAttemptStart   = "attempt-start"
	StartEventAttemptEnd     = "attempt-end"
	StartEventRetryScheduled = "retry-scheduled"
	StartEventRetryBlocked   = "retry-blocked"
)

type StartConfig struct {
	Service          string
	Image            string
	PreflightTimeout time.Duration
	AttemptTimeout   time.Duration
	RetryBackoff     time.Duration
	MaxAttempts      int
	Preflight        func(context.Context) (string, error)
	Retryable        func(error) bool
	Sleep            func(context.Context, time.Duration) error
	Observer         StartObserver
}

type StartObserver func(StartEvent)

type StartEvent struct {
	Type                  string
	Operation             string
	Service               string
	Image                 string
	DockerEndpoint        string
	Attempt               int
	MaxAttempts           int
	StartTime             time.Time
	EndTime               time.Time
	Duration              time.Duration
	Status                string
	Retryable             bool
	RetryBlockedByContext bool
	Err                   error
}

type StartFailure struct {
	Operation             string
	Service               string
	Image                 string
	DockerEndpoint        string
	AttemptsStarted       int
	MaxAttempts           int
	Retryable             bool
	Cause                 error
	RetryBlockedByContext bool

	retryBlockedErr error
}

func (f *StartFailure) Error() string {
	if f == nil {
		return ""
	}

	endpointLabel := strings.TrimSpace(f.DockerEndpoint)
	if endpointLabel == "" {
		endpointLabel = "unresolved"
	}

	maxAttemptsLabel := f.MaxAttempts
	if maxAttemptsLabel <= 0 {
		maxAttemptsLabel = f.AttemptsStarted
	}

	prefix := fmt.Sprintf("%s %s (image %s, docker endpoint %s) failed", f.Operation, f.Service, f.Image, endpointLabel)
	if f.AttemptsStarted > 0 {
		prefix = fmt.Sprintf("%s after %d/%d attempt(s)", prefix, f.AttemptsStarted, maxAttemptsLabel)
	}

	message := f.DiagnosticMessage()
	if message == "" {
		return prefix
	}
	return fmt.Sprintf("%s: %s", prefix, message)
}

func (f *StartFailure) Unwrap() error {
	if f == nil {
		return nil
	}
	return f.Cause
}

func (f *StartFailure) DiagnosticMessage() string {
	if f == nil {
		return ""
	}

	message := ""
	if f.Cause != nil {
		message = f.Cause.Error()
	}
	if !f.RetryBlockedByContext {
		return message
	}

	contextMessage := "context ended before retry"
	if f.retryBlockedErr != nil {
		contextMessage = f.retryBlockedErr.Error()
	}
	if message == "" {
		return fmt.Sprintf("retry blocked by context: %s", contextMessage)
	}
	return fmt.Sprintf("%s (retry blocked by context: %s)", message, contextMessage)
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

	maxAttempts := config.MaxAttempts
	if maxAttempts <= 0 {
		maxAttempts = 1
	}

	retryBackoff := config.RetryBackoff
	if maxAttempts > 1 && retryBackoff <= 0 {
		return zero, errors.New("testcontainersx: retry backoff is required when max attempts is greater than 1")
	}

	preflightStart := time.Now().UTC()
	observeStart(config.Observer, StartEvent{
		Type:        StartEventPreflightStart,
		Operation:   "docker preflight",
		Service:     service,
		Image:       image,
		MaxAttempts: maxAttempts,
		StartTime:   preflightStart,
	})
	preflightCtx, cancel := context.WithTimeout(ctx, preflightTimeout)
	endpoint, err := preflight(preflightCtx)
	cancel()
	preflightEnd := time.Now().UTC()
	observeStart(config.Observer, StartEvent{
		Type:           StartEventPreflightEnd,
		Operation:      "docker preflight",
		Service:        service,
		Image:          image,
		DockerEndpoint: endpoint,
		MaxAttempts:    maxAttempts,
		StartTime:      preflightStart,
		EndTime:        preflightEnd,
		Duration:       preflightEnd.Sub(preflightStart),
		Status:         statusForErr(err),
		Err:            err,
	})
	if err != nil {
		return zero, newStartFailure("docker preflight", service, image, endpoint, 0, maxAttempts, false, err)
	}

	var lastRetryableErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if attempt > 1 {
			if err := ctx.Err(); err != nil && lastRetryableErr != nil {
				observeStart(config.Observer, StartEvent{
					Type:                  StartEventRetryBlocked,
					Operation:             "start",
					Service:               service,
					Image:                 image,
					DockerEndpoint:        endpoint,
					Attempt:               attempt - 1,
					MaxAttempts:           maxAttempts,
					Status:                "fail",
					Retryable:             true,
					RetryBlockedByContext: true,
					Err:                   err,
				})
				return zero, newRetryBlockedFailure("start", service, image, endpoint, attempt-1, maxAttempts, lastRetryableErr, err)
			}
		}

		attemptStart := time.Now().UTC()
		observeStart(config.Observer, StartEvent{
			Type:           StartEventAttemptStart,
			Operation:      "start",
			Service:        service,
			Image:          image,
			DockerEndpoint: endpoint,
			Attempt:        attempt,
			MaxAttempts:    maxAttempts,
			StartTime:      attemptStart,
		})
		attemptCtx := ctx
		cancelAttempt := func() {}
		if config.AttemptTimeout > 0 {
			attemptCtx, cancelAttempt = context.WithTimeout(ctx, config.AttemptTimeout)
		}
		value, startErr := startup(attemptCtx)
		cancelAttempt()
		attemptEnd := time.Now().UTC()
		if startErr == nil {
			observeStart(config.Observer, StartEvent{
				Type:           StartEventAttemptEnd,
				Operation:      "start",
				Service:        service,
				Image:          image,
				DockerEndpoint: endpoint,
				Attempt:        attempt,
				MaxAttempts:    maxAttempts,
				StartTime:      attemptStart,
				EndTime:        attemptEnd,
				Duration:       attemptEnd.Sub(attemptStart),
				Status:         "pass",
			})
			return value, nil
		}

		retryable := isTransientDockerStartupError(startErr)
		if config.Retryable != nil && config.Retryable(startErr) {
			retryable = true
		}
		observeStart(config.Observer, StartEvent{
			Type:           StartEventAttemptEnd,
			Operation:      "start",
			Service:        service,
			Image:          image,
			DockerEndpoint: endpoint,
			Attempt:        attempt,
			MaxAttempts:    maxAttempts,
			StartTime:      attemptStart,
			EndTime:        attemptEnd,
			Duration:       attemptEnd.Sub(attemptStart),
			Status:         "fail",
			Retryable:      retryable,
			Err:            startErr,
		})
		if attempt >= maxAttempts || !retryable {
			return zero, newStartFailure("start", service, image, endpoint, attempt, maxAttempts, retryable, startErr)
		}
		lastRetryableErr = startErr

		observeStart(config.Observer, StartEvent{
			Type:           StartEventRetryScheduled,
			Operation:      "start",
			Service:        service,
			Image:          image,
			DockerEndpoint: endpoint,
			Attempt:        attempt,
			MaxAttempts:    maxAttempts,
			Status:         "pass",
			Retryable:      true,
			Err:            startErr,
		})
		if err := sleep(ctx, retryBackoff); err != nil {
			if isContextCompletion(err) {
				observeStart(config.Observer, StartEvent{
					Type:                  StartEventRetryBlocked,
					Operation:             "start",
					Service:               service,
					Image:                 image,
					DockerEndpoint:        endpoint,
					Attempt:               attempt,
					MaxAttempts:           maxAttempts,
					Status:                "fail",
					Retryable:             true,
					RetryBlockedByContext: true,
					Err:                   err,
				})
				return zero, newRetryBlockedFailure("start", service, image, endpoint, attempt, maxAttempts, startErr, err)
			}
			return zero, newStartFailure("start", service, image, endpoint, attempt, maxAttempts, retryable, fmt.Errorf("wait for retry backoff: %w", err))
		}
	}

	if lastRetryableErr != nil {
		return zero, newStartFailure("start", service, image, endpoint, maxAttempts, maxAttempts, true, lastRetryableErr)
	}
	return zero, newStartFailure("start", service, image, endpoint, maxAttempts, maxAttempts, false, errors.New("startup failed without an attributed cause"))
}

func observeStart(observer StartObserver, event StartEvent) {
	if observer == nil {
		return
	}
	if event.EndTime.Before(event.StartTime) {
		event.EndTime = event.StartTime
	}
	if event.Duration < 0 {
		event.Duration = 0
	}
	observer(event)
}

func statusForErr(err error) string {
	if err == nil {
		return "pass"
	}
	return "fail"
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

func isContextCompletion(err error) bool {
	return errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded)
}

func newStartFailure(operation string, service string, image string, endpoint string, attemptsStarted int, maxAttempts int, retryable bool, err error) error {
	return &StartFailure{
		Operation:       operation,
		Service:         service,
		Image:           image,
		DockerEndpoint:  endpoint,
		AttemptsStarted: attemptsStarted,
		MaxAttempts:     maxAttempts,
		Retryable:       retryable,
		Cause:           err,
	}
}

func newRetryBlockedFailure(operation string, service string, image string, endpoint string, attemptsStarted int, maxAttempts int, cause error, retryBlockedErr error) error {
	return &StartFailure{
		Operation:             operation,
		Service:               service,
		Image:                 image,
		DockerEndpoint:        endpoint,
		AttemptsStarted:       attemptsStarted,
		MaxAttempts:           maxAttempts,
		Retryable:             true,
		Cause:                 cause,
		RetryBlockedByContext: true,
		retryBlockedErr:       retryBlockedErr,
	}
}
