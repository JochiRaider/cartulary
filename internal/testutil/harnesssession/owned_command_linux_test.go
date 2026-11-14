//go:build linux

package harnesssession

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

func ownedCommandFixture(t *testing.T, command *exec.Cmd) *OwnedCommand {
	t.Helper()
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	owner, err := StartOwnedCommand(command, nil)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if _, err := owner.Settle(ctx, 0); err != nil {
			t.Error("owned fixture cleanup", err)
		}
	})
	return owner
}

func testOwnedCommand(t *testing.T) {
	t.Run("owned_command_retirement_precedes_reap_and_retains_failed_anchor", func(t *testing.T) {
		command := exec.Command("/bin/sh", "-c", "exit 17")
		command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		refused := errors.New("retirement refused")
		calls := 0
		owner, err := StartOwnedCommand(command, func(proof ProcessProof) error {
			calls++
			leader, err := readProcess(proof.PID)
			if err != nil || leader.state != "Z" || !sameProcess(leader.ProcessProof, proof) {
				return errors.New("retirement lost unreaped anchor")
			}
			if calls == 1 {
				return refused
			}
			return nil
		})
		if err != nil {
			t.Fatal(err)
		}
		defer func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if _, err := owner.Settle(ctx, 0); err != nil {
				t.Error(err)
			}
		}()
		select {
		case <-owner.Exited():
		case <-time.After(5 * time.Second):
			t.Fatal("leader did not exit")
		}
		ctx, cancel := context.WithTimeout(t.Context(), 5*time.Second)
		defer cancel()
		if _, err := owner.Settle(ctx, 0); !errors.Is(err, refused) || command.ProcessState != nil {
			t.Fatal("failed retirement reaped the child", err)
		}
		if !errors.Is(owner.WaitError(), ErrUnresolved) {
			t.Fatal("unsettled command exposed wait outcome")
		}
		for range 2 {
			code, err := owner.Settle(ctx, 0)
			var exit *exec.ExitError
			if err != nil || code != 17 || !errors.As(owner.WaitError(), &exit) || exit.ExitCode() != 17 {
				t.Fatal("retirement retry lost original outcome", code, err)
			}
		}
		if calls != 2 {
			t.Fatal("successful retirement was repeated", calls)
		}
		if _, err := readProcess(owner.Proof().PID); !processGone(err) {
			t.Fatal("settled child remains unreaped", err)
		}
	})
	t.Run("owned_command_retains_anchor_and_original_exit", func(t *testing.T) {
		peer := ownedCommandFixture(t, exec.Command("sleep", "60"))
		file := filepath.Join(t.TempDir(), "descendant")
		owner := ownedCommandFixture(t, exec.Command("/bin/sh", "-c", `sleep 60 </dev/null >/dev/null 2>&1 & echo $! > "$1"; exit 17`, "fixture", file))
		select {
		case <-owner.exited:
		case <-time.After(5 * time.Second):
			t.Fatal("leader did not exit")
		}
		leader, err := readProcess(owner.proof.PID)
		if err != nil || leader.state != "Z" || !sameProcess(leader.ProcessProof, owner.proof) {
			t.Fatal("exit observation lost unreaped identity anchor", leader, err)
		}
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
		if err != nil {
			t.Fatal(err)
		}
		before, err := readProcess(pid)
		if err != nil || before.state == "Z" || before.PGID != owner.proof.PGID {
			t.Fatal("fixture has no live owned descendant", before, err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		for range 2 {
			code, err := owner.Settle(ctx, 0)
			if err != nil || code != 17 {
				t.Fatal("original result/idempotence", code, err)
			}
		}
		if _, err := readProcess(owner.proof.PID); !processGone(err) {
			t.Fatal("settled direct child was not reaped", err)
		}
		after, err := readProcess(pid)
		if err != nil && !processGone(err) || err == nil && after.state != "Z" && after.state != "X" {
			t.Fatal("live descendant survived settlement", after, err)
		}
		live, err := readProcess(peer.proof.PID)
		if err != nil || !sameProcess(live.ProcessProof, peer.proof) || live.state == "Z" || live.state == "X" {
			t.Fatal("unrelated process changed", err)
		}
	})
	t.Run("owned_command_expired_attempt_retains_retry_authority", func(t *testing.T) {
		owner := ownedCommandFixture(t, exec.Command("sleep", "60"))
		cancelled, cancel := context.WithCancel(context.Background())
		cancel()
		if _, err := owner.Settle(cancelled, 0); !errors.Is(err, context.Canceled) {
			t.Fatal("expired settlement admitted", err)
		}
		state, err := readProcess(owner.proof.PID)
		if err != nil || state.state == "Z" || state.state == "X" || owner.settled {
			t.Fatal("expired attempt changed ownership", state, err)
		}
		ctx, stop := context.WithTimeout(context.Background(), 5*time.Second)
		defer stop()
		code, err := owner.Settle(ctx, 0)
		if err != nil || code != 137 {
			t.Fatal("retry/native signal status", code, err)
		}
	})
	t.Run("owned_command_rejects_competing_wait_authority", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		for _, command := range []*exec.Cmd{
			exec.CommandContext(ctx, "sleep", "60"),
			exec.Command("sleep", "60"),
		} {
			if command.Cancel != nil {
				command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
			}
			if _, err := StartOwnedCommand(command, nil); err == nil || command.Process != nil {
				t.Fatal("invalid launch acquired a child", err)
			}
		}
		missing := exec.Command(filepath.Join(t.TempDir(), "missing"))
		missing.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		if _, err := StartOwnedCommand(missing, nil); err == nil || missing.Process != nil {
			t.Fatal("missing executable admission", err)
		}
	})
	t.Run("cancelled_physical_worker_does_not_launch", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		command := exec.Command("sleep", "60")
		command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		if err := waitPhysicalCommand(ctx, command, nil); !errors.Is(err, context.Canceled) || command.Process != nil {
			t.Fatal("cancelled worker acquired a process", err)
		}
	})
}
