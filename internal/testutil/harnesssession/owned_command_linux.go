//go:build linux

package harnesssession

import (
	"context"
	"errors"
	"os/exec"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

// OwnedCommand keeps the direct child unreaped until the physical owner has
// verified group settlement. The caller chooses when to stop; child exit is
// only a wakeup, never permission to release its dependencies.
type OwnedCommand struct {
	command         *exec.Cmd
	retire          func(ProcessProof) error
	proof           ProcessProof
	exited          chan struct{}
	exitObservation error
	settlement      chan struct{}
	settled         bool
	code            int
	waitError       error
	settlementError error
}

// StartOwnedCommand acquires exclusive launch/wait authority. Callers must not
// invoke Wait, Kill or Signal on command after a successful acquisition. The
// optional retirement callback runs after group settlement, before reaping. Its
// failure retains the anchor and must permit a later idempotent retry.
func StartOwnedCommand(command *exec.Cmd, retire func(ProcessProof) error) (*OwnedCommand, error) {
	// CommandContext's independent cancellation and Wait watcher cannot share
	// this command's wait/signal authority.
	if command == nil || command.Cancel != nil || command.Process != nil ||
		command.SysProcAttr == nil || (!command.SysProcAttr.Setpgid && !command.SysProcAttr.Setsid) ||
		command.SysProcAttr.Pgid != 0 {
		return nil, ErrUnresolved
	}
	if err := command.Start(); err != nil {
		return nil, err
	}
	proof, err := CaptureProcess(command.Process.Pid)
	if err != nil {
		// No other waiter can have reaped this direct child. Its os.Process
		// handle remains the only safe fallback when proof acquisition fails.
		return nil, errors.Join(err, command.Process.Kill(), command.Wait())
	}
	owner := &OwnedCommand{command: command, retire: retire, proof: proof, exited: make(chan struct{}), settlement: make(chan struct{}, 1)}
	go func() {
		defer close(owner.exited)
		for {
			var info unix.Siginfo
			owner.exitObservation = unix.Waitid(unix.P_PID, command.Process.Pid, &info, unix.WEXITED|unix.WNOWAIT, nil)
			if owner.exitObservation != unix.EINTR {
				return
			}
		}
	}()
	return owner, nil
}

// Settle returns the original native exit separately from cleanup failure.
// An unresolved attempt retains the child anchor and can be retried. Success
// is idempotent and does not signal or wait on a numeric PID again.
func (owner *OwnedCommand) Settle(ctx context.Context, grace time.Duration) (int, error) {
	select {
	case owner.settlement <- struct{}{}:
		defer func() { <-owner.settlement }()
	case <-ctx.Done():
		return 0, ctx.Err()
	}
	if owner.settled {
		return owner.code, owner.settlementError
	}
	if err := ctx.Err(); err != nil {
		return 0, err
	}
	if err := StopProcess(ctx, owner.proof, grace); err != nil {
		return 0, AtCleanupStage("process_settlement", err)
	}
	select {
	case <-owner.exited:
	case <-ctx.Done():
		return 0, ctx.Err()
	}
	if owner.exitObservation != nil {
		return 0, owner.exitObservation
	}
	if owner.retire != nil {
		if err := owner.retire(owner.proof); err != nil {
			return 0, err
		}
	}
	owner.waitError = owner.command.Wait()
	if owner.command.ProcessState == nil {
		return 0, errors.Join(ErrUnresolved, owner.waitError)
	}
	owner.code = owner.command.ProcessState.ExitCode()
	if status, ok := owner.command.ProcessState.Sys().(syscall.WaitStatus); ok && status.Signaled() {
		owner.code = 128 + int(status.Signal())
	}
	owner.settled = true
	var exit *exec.ExitError
	if owner.waitError != nil && !errors.As(owner.waitError, &exit) {
		owner.settlementError = owner.waitError
	}
	return owner.code, owner.settlementError
}

// Proof identifies the acquired process group. It is a registration fact, not
// authority for a second waiter or signal policy.
func (owner *OwnedCommand) Proof() ProcessProof { return owner.proof }

// Exited wakes the caller when the direct child exits without reaping it.
// Settlement is still required before output or dependencies may be released.
func (owner *OwnedCommand) Exited() <-chan struct{} { return owner.exited }

// WaitError preserves the original exec outcome after successful settlement.
func (owner *OwnedCommand) WaitError() error {
	if !owner.settled {
		return ErrUnresolved
	}
	return owner.waitError
}
