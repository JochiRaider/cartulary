package recoverybrowsertest

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"sync"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/postgrescleanup"
)

// Lifecycle records stay on the private diagnostic pipe. The browser process
// owner redacts them before publishing an artifact; stdout carries credentials.
type Lifecycle struct {
	attempt       string
	output        io.Writer
	start         time.Time
	cleanupBudget time.Duration
	mu            sync.Mutex
}

func NewLifecycle(attempt string, output io.Writer, cleanupBudget time.Duration) *Lifecycle {
	return &Lifecycle{attempt: attempt, output: output, start: time.Now(), cleanupBudget: cleanupBudget}
}

func (l *Lifecycle) event(stage, phase, outcome string, deadline time.Duration, err error) {
	l.mu.Lock()
	defer l.mu.Unlock()
	message := ""
	postgres := postgrescleanup.ObservationFromError(err)
	if postgres != nil {
		message = "PostgreSQL cleanup details attached"
	} else if err != nil {
		message = err.Error()
	}
	data := struct {
		SchemaID   string                       `json:"schema_id"`
		AttemptID  string                       `json:"attempt_id"`
		Stage      string                       `json:"stage"`
		Phase      string                       `json:"phase"`
		Outcome    string                       `json:"outcome"`
		ElapsedMS  int64                        `json:"elapsed_ms"`
		DeadlineMS int64                        `json:"deadline_ms"`
		Message    string                       `json:"message"`
		Postgres   *postgrescleanup.DropAttempt `json:"postgres,omitempty"`
	}{"cartulary.browser_fixture_event.v3", l.attempt, stage, phase, outcome, time.Since(l.start).Milliseconds(), deadline.Milliseconds(), message, postgres}
	_ = json.NewEncoder(l.output).Encode(data)
}

// Each phase inherits the one fixture cleanup deadline. Synchronous closers
// are never abandoned in a goroutine and represented as settled.
func (l *Lifecycle) cleanup(ctx context.Context, stage string, limit time.Duration, close func(context.Context) error) error {
	if close == nil {
		return nil
	}
	if limit > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, limit)
		defer cancel()
	}
	deadline := time.Duration(0)
	if end, ok := ctx.Deadline(); ok {
		deadline = max(0, time.Until(end))
	}
	l.event(stage, "cleanup", "started", deadline, nil)
	err := errors.Join(close(ctx), ctx.Err())
	outcome := "succeeded"
	if err != nil {
		outcome = "failed"
		err = fmt.Errorf("%s: %w", stage, err)
	}
	l.event(stage, "cleanup", outcome, deadline, err)
	return err
}

func closeWithoutError(close func()) func(context.Context) error {
	return func(context.Context) error { close(); return nil }
}
func closeResult(close func() error) func(context.Context) error {
	return func(context.Context) error { return close() }
}

// This private fixture has exactly two deployments and one backup store. Its
// phase order is explicit; it is not a general resource registry or scheduler.
type deploymentCleanup struct {
	http, listener, runtime, runtimePool, recovery, database, objects, referencePacks func(context.Context) error
}
type fixtureCleanup struct {
	lifecycle      *Lifecycle
	source, target deploymentCleanup
	backup         func(context.Context) error
}

func (c *fixtureCleanup) close(ctx context.Context) error {
	var result error
	deployments := []struct {
		name  string
		owned *deploymentCleanup
	}{{"source", &c.source}, {"target", &c.target}}
	run := func(stage string, limit time.Duration, close func(context.Context) error) {
		result = errors.Join(result, c.lifecycle.cleanup(ctx, stage, limit, close))
	}
	for _, d := range deployments {
		if d.owned.http != nil {
			run(d.name+".http", 5*time.Second, d.owned.http)
		} else {
			run(d.name+".listener", 0, d.owned.listener)
		}
	}
	for _, d := range deployments {
		run(d.name+".runtime", 0, d.owned.runtime)
	}
	for _, d := range deployments {
		run(d.name+".runtime_pool", 0, d.owned.runtimePool)
		run(d.name+".recovery", 0, d.owned.recovery)
	}
	// At most two independent targets; join both results before closing stores or
	// emitting terminal evidence. The existing browser profile claims postgres=2.
	failures := make([]error, len(deployments))
	var workers sync.WaitGroup
	for i, d := range deployments {
		if d.owned.database != nil {
			workers.Go(func() { failures[i] = c.lifecycle.cleanup(ctx, d.name+".database", 0, d.owned.database) })
		}
	}
	workers.Wait()
	result = errors.Join(result, errors.Join(failures...))
	for _, d := range deployments {
		run(d.name+".reference_packs", 0, d.owned.referencePacks)
		run(d.name+".objects", 0, d.owned.objects)
	}
	run("backup.storage", 0, c.backup)
	return errors.Join(result, ctx.Err())
}
