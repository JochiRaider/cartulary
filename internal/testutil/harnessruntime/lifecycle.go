package harnessruntime

import (
	"context"
	"fmt"
	"maps"
	"slices"
	"time"
)

func (e *graphEngine) send(m engineMessage) error {
	m.At = e.ports.Now()
	m.Ack = make(chan error, 1)
	e.mailbox <- m
	return <-m.Ack
}
func (e *graphEngine) workerPhase(u unit, name string, operation func() error) error {
	id := fmt.Sprintf("phase:%d", e.phaseSequence.Add(1))
	err := e.send(engineMessage{ID: u.ID, Event: "phase_started", Extra: map[string]any{"interval_id": id, "phase": name}})
	// Lifecycle cleanup must still execute after a required publication failed.
	operationError := operation()
	outcome := "completed"
	if operationError != nil {
		outcome = "failed"
		if e.ctx.Err() != nil {
			outcome = "cancelled"
		}
	}
	publicationError := e.send(engineMessage{ID: u.ID, Event: "phase_finished", Extra: map[string]any{"interval_id": id, "phase": name, "phase_outcome": outcome}})
	if operationError != nil {
		return operationError
	}
	if err != nil {
		return err
	}
	return publicationError
}
func (e *graphEngine) work(u unit, c cacheResult) {
	r := unitResult{Status: "failed", Cache: c}
	defer func() { e.mailbox <- engineMessage{ID: u.ID, At: e.ports.Now(), Result: &r} }()
	activity := func(name, reason string) { _ = e.send(engineMessage{ID: u.ID, Activity: name, WaitReason: reason}) }
	setFailure := func(err error) {
		if r.Failure == nil {
			r.Failure = failureOf(err)
			r.Status = "failed"
			if r.Failure.Class == "interrupted" {
				r.Status = "cancelled"
			}
			r.ExitCode = failureExitCode(r.Failure, r.Signal)
		}
	}
	cleanup := func(operation, lease string, err error) {
		record := cleanupRecord{Operation: operation, UnitID: u.ID, LeaseID: lease, Outcome: "completed", Cause: err}
		if err != nil {
			record.Outcome = "failed"
			record.Failure = cleanupFailureOf(err)
			record.Failure.Lifecycle = "cleanup_finalizers"
			r.Cleanup = append(r.Cleanup, record.Failure)
			setFailure(record.Failure)
		}
		_ = e.send(engineMessage{ID: u.ID, Cleanup: &record})
	}
	quiet := slices.Contains(u.ExclusiveLocks, "host_activity")
	if quiet && e.ports.Pause != nil {
		if err := e.ports.Pause(e.ctx); err != nil {
			setFailure(err)
			return
		}
		if e.ports.Resume != nil {
			defer e.ports.Resume()
		}
	}
	var host *hostLease
	activity("waiting", "host_admission")
	err := e.workerPhase(u, "host_admission", func() error {
		if e.ports.Admit == nil {
			return nil
		}
		var err error
		host, err = e.ports.Admit(e.ctx, u)
		return err
	})
	// Finalizers receive independent deadlines even when the work was cancelled.
	finishHost := func() {
		if host == nil {
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		cleanup("host_release", "", host.Release(ctx))
		host = nil
	}
	defer finishHost()
	if err != nil {
		setFailure(err)
		return
	}
	var lease *fixtureLease
	finishFixture := func() {
		if lease == nil || lease.Release == nil {
			return
		}
		owned := lease
		lease = nil
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
		defer cancel()
		err := e.workerPhase(u, "fixture_release", func() error {
			var err error
			r.RetainedFixture, err = owned.Release(ctx, e.ctx.Err() == nil && (u.Fixture != "browser_stack" || r.Status == "passed"), u.Command.Environment["CARTULARY_BROWSER_RELEASE_AFFINITY"] != "1")
			return err
		})
		cleanup("fixture_release", owned.ID, err)
		extra := map[string]any{"fixture_lease_id": owned.ID}
		if err != nil {
			maps.Copy(extra, failureFields(failureOf(err)))
		}
		_ = e.send(engineMessage{ID: u.ID, Event: "fixture_released", Extra: extra})
	}
	defer finishFixture()
	if u.Fixture != "none" || u.Renderer != "none" && u.Renderer != "" {
		activity("preparing", "fixture_acquisition")
		err = e.workerPhase(u, "fixture_acquire", func() error {
			if e.ports.Fixtures == nil {
				return failed("harness", "fixture_error")
			}
			var err error
			lease, err = e.ports.Fixtures.Acquire(e.ctx, fixtureRequest{PostgresIsolation: u.PostgresIsolation, Services: u.Services, UnitID: u.ID, Owner: u.Owner, Capability: u.Fixture, Affinity: fixtureAffinity(u), Environment: maps.Clone(u.Command.Environment)})
			return err
		})
		if err != nil {
			setFailure(err)
			return
		}
		if lease == nil || lease.Release == nil {
			setFailure(failed("harness", "fixture_error"))
			return
		}
		_ = e.send(engineMessage{ID: u.ID, Event: "fixture_acquired", Extra: map[string]any{"fixture_lease_id": lease.ID}})
	}
	environment := map[string]string{}
	if host != nil {
		environment["CARTULARY_HOST_ADMISSION_LEASE"] = host.Token
	}
	if lease != nil {
		maps.Copy(environment, lease.Environment)
	}
	activity("executing", "")
	err = e.workerPhase(u, "runner", func() error {
		if err := e.ctx.Err(); err != nil {
			return err
		}
		r = e.ports.Execute(e.ctx, executionRequest{Unit: u, Environment: environment, Host: host})
		r.Cache = c
		for _, secondary := range r.Cleanup {
			_ = e.send(engineMessage{ID: u.ID, Cleanup: &cleanupRecord{Operation: "runtime_close", UnitID: u.ID, Outcome: "failed", Cause: secondary, Failure: cleanupFailureOf(secondary)}})
		}
		if r.Status != "passed" && r.Status != "failed" && r.Status != "cancelled" {
			return failed("harness", "scheduler_accounting_error")
		}
		if r.Status != "passed" && r.Failure == nil {
			r.Failure = failed("unknown", "unknown_failure")
		}
		return nil
	})
	if err != nil {
		setFailure(err)
	}
	activity("finalizing", "")
	finishFixture()
	finishHost()
	if e.ports.Terminal != nil {
		err = e.workerPhase(u, "evidence_write", func() error { return e.ports.Terminal(u, &r) })
		if err != nil {
			setFailure(failed("artifact", "artifact_error"))
		}
	}
	if r.Status == "passed" && !r.RetainedFixture && e.ports.Cache != nil {
		err = e.workerPhase(u, "cache_publish", func() error { return e.ports.Cache.Store(e.ctx, u) })
		if err != nil {
			setFailure(failed("artifact", "artifact_error"))
		}
	}
}
