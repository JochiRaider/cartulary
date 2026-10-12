//go:build linux

package harnessreview

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessadmission"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessfixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesssession"
	"golang.org/x/sys/unix"
)

func recoveryLock(ctx context.Context, session string) (func(), error) {
	name := registryPath(session) + ".lock"
	fd, err := unix.Open(name, unix.O_RDWR|unix.O_CREAT|unix.O_NOFOLLOW|unix.O_CLOEXEC, 0600)
	if err != nil {
		return nil, err
	}
	file := os.NewFile(uintptr(fd), name)
	if _, err = harnesslocal.ReadFile(name, 0, false); err != nil {
		file.Close()
		return nil, err
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	for {
		err = unix.Flock(fd, unix.LOCK_EX|unix.LOCK_NB)
		if err == nil {
			break
		}
		if err != unix.EWOULDBLOCK {
			file.Close()
			return nil, err
		}
		timer := time.NewTimer(20 * time.Millisecond)
		select {
		case <-timer.C:
		case <-ctx.Done():
			timer.Stop()
			file.Close()
			return nil, failure("capacity_exceeded")
		}
	}
	return func() {
		_ = unix.Flock(fd, unix.LOCK_UN)
		file.Close()
		if _, err := os.Lstat(registryPath(session)); errors.Is(err, os.ErrNotExist) {
			_ = os.Remove(name)
		}
	}, nil
}
func recoverSession(ctx context.Context, repository, file string, locator Locator, command string, environment map[string]string) (Result, error) {
	release, err := recoveryLock(ctx, locator.Session)
	if err != nil {
		return result(command), err
	}
	defer release()
	// Re-read after locking; a concurrent stop may have completed immutably.
	_, locator, err = readLocator(repository, file)
	if err != nil {
		return result(command), err
	}
	terminal := locator.Receipt != nil
	var prior Result
	if terminal {
		prior, err = terminalResult(file, locator, command)
		if err != nil {
			return prior, err
		}
	}
	record, err := resolveRecord(repository, file, locator)
	if errors.Is(err, os.ErrNotExist) && terminal {
		return prior, nil
	}
	if err != nil {
		return result(command), err
	}
	if record.Outcome != nil {
		pending := record.Outcome
		if pending.Session != locator.Session || pending.Mode != locator.Mode || pending.Operation != nil || pending.Command != commandID("ui-review") || pending.Cleanup == "not_terminal" {
			return result(command), failure("invalid_artifact")
		}
		for _, f := range pending.Failures {
			if f.validate() != nil {
				return result(command), failure("invalid_artifact")
			}
		}
	}
	alive, err := harnessadmission.Alive(record.Process)
	if err != nil || alive {
		return result(command), failure("session_mismatch")
	}
	ctx, cancel := context.WithTimeout(ctx, 330*time.Second)
	defer cancel()
	runRoot := filepath.Dir(filepath.Dir(file))
	proof, err := harnessadmission.Identity(os.Getpid())
	if err != nil {
		return result(command), err
	}
	var runtime *harnesssession.Runtime
	runtime, err = harnesssession.OwnRuntime(repository, runRoot, record.Runtime.Root, record.Runtime.RunID, record.Runtime.Lease, proof, true)
	if err != nil && !(errors.Is(err, os.ErrNotExist) && record.Released) {
		return result(command), failure("unsafe_artifact")
	}
	failures := []Failure{failure("session_lost").Record}
	cleanupFailed := false
	secrets := []string{}
	if runtime != nil {
		defer runtime.ReleaseHandles()
		secrets, err = runtime.ForbiddenValues()
		if err != nil {
			return result(command), err
		}
		cleanup := recoverIndependent(ctx, repository, runRoot, runtime, environment)
		if len(cleanup) == 0 {
			physicalRoot := runRoot
			if record.Presentation == "ui" {
				physicalRoot, err = runtime.PrivatePath("lifecycle", runtime.RunID)
				if err != nil {
					return result(command), err
				}
			}
			owner := harnessfixtures.ProductionOptions{Root: repository, RunRoot: physicalRoot, Target: "ui-review", Runtime: runtime, Environment: environment}
			if err = owner.Recover(ctx); err != nil {
				cleanup = append(cleanup, err)
			}
		}
		cleanupFailed = len(cleanup) > 0
		if cleanupFailed {
			failures = append(failures, failure("cleanup_failed").Record)
		}
	}
	if terminal {
		if !cleanupFailed {
			if runtime != nil {
				if _, err = runtime.BeginFinalization(); err != nil {
					return prior, err
				}
				record.Released = true
				if err = writeRegistry(record, true); err != nil {
					return prior, err
				}
				if err = runtime.Close(); err != nil {
					return prior, err
				}
			}
			if _, err = harnesssession.ScanRetained(runRoot, secrets); err != nil {
				return prior, err
			}
			if err = removeRegistry(record); err != nil {
				return prior, err
			}
		}
		return prior, nil
	}
	counts, err := recoverCounts(runRoot, locator.Session)
	if err != nil {
		failures = append(failures, failure("invalid_artifact").Record)
	}
	tick, e := strconv.ParseInt(record.StartedTick, 10, 64)
	if e != nil || tick > monotonic() {
		return result(command), failure("invalid_artifact")
	}
	locator.State = "failed"
	publisher := &diskPublisher{runtime: runtime, record: record, runRoot: runRoot, secrets: secrets}
	receipt := Receipt{Schema: "cartulary.ui_review_receipt.v2", Command: commandID("ui-review"), Session: locator.Session, Mode: locator.Mode, State: "failed", Status: "error", Exit: 3, Started: locator.Created, Finished: timestamp(time.Now()), Duration: (monotonic() - tick) / int64(time.Millisecond), Failures: failures, Counts: counts, Cleanup: "complete"}
	if record.Outcome != nil {
		receipt.Failures = append([]Failure{}, record.Outcome.Failures...)
		receipt.Exit = record.Outcome.Exit
		receipt.Failures = appendFailure(receipt.Failures, failure("unsafe_artifact").Record)
		if receipt.Exit == 0 {
			receipt.Exit = 11
		}
	}
	if cleanupFailed {
		receipt.Cleanup = "failed"
		receipt.Failures = appendFailure(receipt.Failures, failure("cleanup_failed").Record)
	}
	var owner Publisher = publisher
	if record.Presentation == "design" {
		owner = &designPublisher{diskPublisher: publisher, profile: record.Profile, ready: record.Review, driver: &productionDriver{cleanup: []cleanupStep{{"provider_close", func() error {
			if cleanupFailed {
				return failure("cleanup_failed")
			}
			return nil
		}()}}}}
	}
	ref, err := owner.Terminal(locator, receipt, cleanupFailed)
	if err != nil {
		locator.Receipt = nil
		_ = owner.Locator(locator)
		v := result(command)
		v.Session = &locator.Session
		v.State = &locator.State
		v.Status = "error"
		v.Exit = receipt.Exit
		v.Failures = appendFailure(receipt.Failures, failure("unsafe_artifact").Record)
		return v, nil
	}
	value := result(command)
	value.Session = &locator.Session
	value.State = &locator.State
	value.Status = "error"
	value.Exit = receipt.Exit
	value.Failures = receipt.Failures
	value.Receipt = ref
	return value, nil
}
func recoverCounts(root, session string) (Counts, error) {
	var counts Counts
	directory := filepath.Join(root, "ui-review/operations")
	d, err := harnesslocal.OpenDirectory(directory, true)
	if errors.Is(err, os.ErrNotExist) {
		return counts, nil
	}
	if err != nil {
		return counts, err
	}
	defer d.Close()
	names, err := d.Readdirnames(-1)
	if err != nil {
		return counts, err
	}
	for _, name := range names {
		id, err := strconv.Atoi(name)
		if err != nil || id < 1 || strconv.Itoa(id) != name {
			return counts, failure("invalid_artifact")
		}
		var receipt Receipt
		if err = readDocument(filepath.Join(directory, name, "receipt.json"), "cartulary.ui_review_receipt.v2", LimitRequest, &receipt); err != nil {
			return counts, err
		}
		if receipt.Session != session || receipt.Operation == nil || *receipt.Operation != id || receipt.State != "ready" || receipt.Cleanup != "not_terminal" || (receipt.Status == "ok") != (len(receipt.Failures) == 0) || (receipt.Exit == 0) != (receipt.Status == "ok") {
			return counts, failure("invalid_artifact")
		}
		for _, failure := range receipt.Failures {
			if err = failure.validate(); err != nil {
				return counts, err
			}
		}
		if receipt.Status == "ok" {
			counts.Add(receipt.Counts)
		}
	}
	return counts, nil
}
