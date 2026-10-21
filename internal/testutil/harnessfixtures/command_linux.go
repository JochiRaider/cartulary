//go:build linux

package harnessfixtures

import (
	"context"
	"errors"
	"maps"
	"os"
	"os/exec"
	"slices"
	"strconv"
	"syscall"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessadmission"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesssession"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices/startupproof"
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

type processTarget = harnesssession.ProcessProof

// commandRunner is private to fixed lifecycle adapters; neither an adapter
// request nor public Make inputs can supply an arbitrary command or argv.
func commandRunner(root string, runtime *harnesssession.Runtime, dependencies []string) Command {
	return func(ctx context.Context, executable string, args []string, environment map[string]string) (result CommandResult, err error) {
		result.Exit = -1
		cmd := exec.Command(executable, args...)
		cmd.Dir = root
		environment = maps.Clone(environment)
		maps.Copy(environment, runtime.Environment())
		var startupRead, startupWrite *os.File
		if len(args) > 0 && args[0] == "start-suite" {
			startupRead, startupWrite, err = os.Pipe()
			if err != nil {
				return result, err
			}
			defer startupRead.Close()
			defer startupWrite.Close()
			cmd.ExtraFiles = append(cmd.ExtraFiles, startupRead)
			environment[startupproof.ControlEnv] = strconv.Itoa(3 + len(cmd.ExtraFiles) - 1)
		}
		for _, key := range slices.Sorted(maps.Keys(environment)) {
			cmd.Env = append(cmd.Env, key+"="+environment[key])
		}
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true, Pdeathsig: syscall.SIGKILL}
		if startupWrite != nil {
			cmd.SysProcAttr.Pdeathsig = 0
		}
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
		target, proofErr := harnesssession.CaptureProcess(cmd.Process.Pid)
		if proofErr == nil {
			proofErr = runtime.RecordDependentResource("helper_process", target, "acquired", dependencies)
		}
		if startupWrite != nil {
			startupRead.Close()
			if proofErr == nil {
				_, proofErr = startupWrite.Write([]byte{1})
			}
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
			if startupWrite == nil {
				signal()
				terminalErr = <-terminal
			} else {
				startupWrite.Close()
				timer := time.NewTimer(2 * time.Minute)
				select {
				case terminalErr = <-terminal:
				case <-timer.C:
					signal()
					terminalErr = <-terminal
				}
				timer.Stop()
			}
		}
		// Keep the zombie leader reserved until group signaling has finished.
		settleCtx, settleCancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		settleErr := harnesssession.StopProcess(settleCtx, target, 0)
		if settleErr == nil {
			settleErr = runtime.RetireDetailProducers(target)
		}
		settleCancel()
		waitErr := cmd.Wait()
		if cmd.ProcessState != nil {
			result.Exit = cmd.ProcessState.ExitCode()
		}
		result.Stdout = output.data
		result.Stderr = diagnostic.data
		releaseErr := error(nil)
		if proofErr == nil && settleErr == nil {
			releaseErr = runtime.RecordResource("helper_process", target, "released")
		}
		if proofErr != nil || terminalErr != nil || releaseErr != nil || settleErr != nil {
			return result, errors.Join(proofErr, terminalErr, CleanupFailure(errors.Join(settleErr, releaseErr)))
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
