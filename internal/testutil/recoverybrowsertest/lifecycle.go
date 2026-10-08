package recoverybrowsertest

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"
)

// Lifecycle records stay on the private diagnostic pipe. The browser process
// owner redacts them before publishing an artifact; stdout carries credentials.
type Lifecycle struct {
	attempt string
	output  io.Writer
	start   time.Time
	mu      sync.Mutex
}

func NewLifecycle(attempt string, output io.Writer) *Lifecycle {
	return &Lifecycle{attempt: attempt, output: output, start: time.Now()}
}

func (l *Lifecycle) event(stage, phase, outcome string, deadline time.Duration, err error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	message := ""
	if err != nil {
		message = err.Error()
	}
	data := struct {
		SchemaID   string `json:"schema_id"`
		AttemptID  string `json:"attempt_id"`
		Stage      string `json:"stage"`
		Phase      string `json:"phase"`
		Outcome    string `json:"outcome"`
		ElapsedMS  int64  `json:"elapsed_ms"`
		DeadlineMS int64  `json:"deadline_ms"`
		Message    string `json:"message"`
	}{"cartulary.browser_fixture_event.v1", l.attempt, stage, phase, outcome, time.Since(l.start).Milliseconds(), deadline.Milliseconds(), message}
	_ = json.NewEncoder(l.output).Encode(data)
}

// cleanup wraps each acquired resource at its acquisition site. Go's defer
// order remains the ownership order, and one error never skips later releases.
func (l *Lifecycle) cleanup(stage string, deadline time.Duration, close func() error, result *error) {
	l.event(stage, "cleanup", "started", deadline, nil)
	err := close()
	outcome := "succeeded"
	if err != nil {
		outcome = "failed"
		*result = errors.Join(*result, fmt.Errorf("%s: %w", stage, err))
	}
	l.event(stage, "cleanup", outcome, deadline, err)
}

func closeWithoutError(close func()) func() error {
	return func() error { close(); return nil }
}
