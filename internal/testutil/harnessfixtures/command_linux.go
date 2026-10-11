//go:build linux

package harnessfixtures

import (
	"context"
	"errors"
	"maps"
	"os/exec"
	"slices"
	"syscall"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessadmission"
	"golang.org/x/sys/unix"
)

type CommandResult struct {
	Exit           int
	Stdout, Stderr []byte
	Started        bool
}
type Command func(context.Context, string, []string, map[string]string) (CommandResult, error)
type boundedOutput struct {
	data     []byte
	maximum  int
	exceeded bool
}

func (b *boundedOutput) Write(p []byte) (int, error) {
	n := len(p)
	remaining := b.maximum - len(b.data)
	b.data = append(b.data, p[:min(n, remaining)]...)
	b.exceeded = b.exceeded || n > remaining
	return n, nil
}

type processTarget struct {
	Boot  string `json:"boot"`
	PID   int    `json:"pid"`
	Start string `json:"start"`
	Group bool   `json:"group"`
}

// commandRunner is private to fixed lifecycle adapters; neither the fixture
// session nor public Make inputs can supply an arbitrary command or argv.
func commandRunner(root string, runtime *Runtime) Command {
	return func(ctx context.Context, executable string, args []string, environment map[string]string) (result CommandResult, err error) {
		result.Exit = -1
		cmd := exec.Command(executable, args...)
		cmd.Dir = root
		for _, key := range slices.Sorted(maps.Keys(environment)) {
			cmd.Env = append(cmd.Env, key+"="+environment[key])
		}
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true, Pdeathsig: syscall.SIGKILL}
		cmd.WaitDelay = 2 * time.Second
		output, diagnostic := &boundedOutput{maximum: 1048576}, &boundedOutput{maximum: 256 * 1024}
		cmd.Stdout = output
		cmd.Stderr = diagnostic
		close, err := harnessadmission.InheritWorkspace(cmd)
		if err != nil {
			return result, err
		}
		defer close()
		if ctx.Err() != nil {
			return result, ctx.Err()
		}
		if err := cmd.Start(); err != nil {
			return result, err
		}
		result.Started = true
		signal := func() { _ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL) }
		proof, proofErr := harnessadmission.Identity(cmd.Process.Pid)
		target := processTarget{proof.Boot, proof.PID, proof.Start, true}
		if proofErr == nil {
			proofErr = runtime.RecordResource("helper_process", target, "acquired")
		}
		if proofErr != nil {
			signal()
		}
		terminal := make(chan error, 1)
		go func() {
			for {
				var info unix.Siginfo
				err := unix.Waitid(unix.P_PID, cmd.Process.Pid, &info, unix.WEXITED|unix.WNOWAIT, nil)
				if err == unix.EINTR {
					continue
				}
				terminal <- err
				return
			}
		}()
		var terminalErr error
		select {
		case terminalErr = <-terminal:
		case <-ctx.Done():
			signal()
			terminalErr = <-terminal
		}
		// Keep the zombie leader reserved until group signaling has finished.
		signal()
		waitErr := cmd.Wait()
		if cmd.ProcessState != nil {
			result.Exit = cmd.ProcessState.ExitCode()
		}
		result.Stdout = output.data
		result.Stderr = diagnostic.data
		releaseErr := error(nil)
		if proofErr == nil {
			releaseErr = runtime.RecordResource("helper_process", target, "released")
		}
		if proofErr != nil || terminalErr != nil || releaseErr != nil {
			return result, errors.Join(proofErr, terminalErr, CleanupFailure(releaseErr))
		}
		if ctx.Err() != nil {
			return result, ctx.Err()
		}
		if output.exceeded || diagnostic.exceeded {
			return result, artifactFailure(errors.New("lifecycle output exceeds bound"))
		}
		if waitErr != nil {
			return result, waitErr
		}
		return result, nil
	}
}
