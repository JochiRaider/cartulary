//go:build linux

package harnessruntime

import (
	"context"
	"errors"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessadmission"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesssession"
	"golang.org/x/sys/unix"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"sync/atomic"
	"syscall"
	"time"
)

// Prepare starts only the repository's source compiler, receives its bounded
// private pipe, and waits for successful compiler exit before exposing any plan.
// It performs no unit, fixture, retained-publication or cache transaction.
func Prepare(ctx context.Context, root string, args, environment []string) (*Plan, error) {
	node := ""
	for _, entry := range environment {
		if strings.HasPrefix(entry, "NODE_BIN=") {
			node = strings.TrimPrefix(entry, "NODE_BIN=")
		}
	}
	if node == "" {
		return nil, errors.New("selected Node runtime is unavailable")
	}
	if !filepath.IsAbs(node) {
		node = filepath.Join(root, node)
	}
	info, err := os.Stat(node)
	if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0111 == 0 {
		return nil, errors.New("selected Node runtime is unavailable")
	}
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()
	cmd := exec.CommandContext(ctx, node, append([]string{filepath.Join(root, "tools/harness/scheduler/work-graph/prepare-cli.mjs")}, args...)...)
	cmd.Dir = root
	cmd.Env = environment
	return receiveCompiler(ctx, cmd)
}
func receiveCompiler(ctx context.Context, cmd *exec.Cmd) (*Plan, error) {
	read, write, err := os.Pipe()
	if err != nil {
		return nil, err
	}
	defer read.Close()
	defer write.Close()
	cmd.ExtraFiles = []*os.File{write}
	closeDescriptors, err := harnessadmission.InheritWorkspace(cmd)
	if err != nil {
		return nil, err
	}
	defer closeDescriptors()
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true, Pdeathsig: syscall.SIGKILL}
	cmd.Stdout = io.Discard
	cmd.Stderr = io.Discard
	var settled atomic.Bool
	cmd.Cancel = func() error {
		if settled.Load() {
			return os.ErrProcessDone
		}
		err := syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
		if errors.Is(err, syscall.ESRCH) {
			return os.ErrProcessDone
		}
		return err
	}
	cmd.WaitDelay = 2 * time.Second
	if err := cmd.Start(); err != nil {
		return nil, errors.New("graph compiler failed to start")
	}
	proof, proofErr := harnesssession.CaptureProcess(cmd.Process.Pid)
	write.Close()
	type publication struct {
		data []byte
		err  error
	}
	output := make(chan publication, 1)
	go func() { b, e := io.ReadAll(io.LimitReader(read, maximumPlanBytes+1)); output <- publication{b, e} }()
	// Observe terminal state without reaping: the zombie pins the process-group
	// identity until descendants are stopped, avoiding a signal to a recycled PID.
	terminal := make(chan error, 1)
	go func() {
		for {
			var info unix.Siginfo
			e := unix.Waitid(unix.P_PID, cmd.Process.Pid, &info, unix.WEXITED|unix.WNOWAIT, nil)
			if e == unix.EINTR {
				continue
			}
			terminal <- e
			return
		}
	}()
	var received publication
	var terminalErr error
	readDone, rootDone := false, false
	cancellation := ctx.Done()
	var drain <-chan time.Time
receive:
	for !readDone || !rootDone {
		select {
		case received = <-output:
			readDone = true
			output = nil
			if received.err != nil || len(received.data) > maximumPlanBytes {
				_ = cmd.Cancel()
			}
		case terminalErr = <-terminal:
			rootDone = true
			terminal = nil
			_ = cmd.Cancel()
			if !readDone {
				drain = time.After(2 * time.Second)
			}
		case <-cancellation:
			_ = cmd.Cancel()
			cancellation = nil
			drain = time.After(2 * time.Second)
			if !readDone {
				_ = read.Close()
			}
		case <-drain:
			_ = read.Close()
			drain = nil
			if !rootDone {
				break receive
			}
		}
	}
	settleCtx, settleCancel := context.WithTimeout(context.WithoutCancel(ctx), 2*time.Second)
	settleErr := proofErr
	if settleErr == nil {
		settleErr = harnesssession.StopProcess(settleCtx, proof, 0)
	}
	settleCancel()
	if !rootDone {
		// A denied signal or an uninterruptible child must not bypass the
		// compiler deadline. Keep its anchor until it can be reaped, and never
		// admit a plan from this unresolved compiler.
		go func() { <-terminal; settled.Store(true); _ = cmd.Wait() }()
		return nil, errors.Join(failed("harness", "cleanup_error"), ctx.Err(), settleErr, harnesssession.ErrUnresolved)
	}
	settled.Store(true)
	err = cmd.Wait()
	if settleErr != nil {
		return nil, errors.Join(failed("harness", "cleanup_error"), ctx.Err(), err, settleErr)
	}
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if err != nil || terminalErr != nil || ctx.Err() != nil || received.err != nil || len(received.data) > maximumPlanBytes {
		return nil, errors.New("graph preparation failed")
	}
	return decodePlan(received.data)
}
