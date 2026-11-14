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
		owner, startErr := harnesssession.StartOwnedCommand(cmd, runtime.RetireDetailProducers)
		result.Started = cmd.Process != nil
		if startErr != nil {
			return result, startErr
		}
		target := owner.Proof()
		registrationErr := runtime.RecordDependentResource("helper_process", target, "acquired", dependencies)
		registered := registrationErr == nil
		if startupWrite != nil {
			startupRead.Close()
			if registrationErr == nil {
				_, registrationErr = startupWrite.Write([]byte{1})
			}
		}
		if registrationErr == nil {
			select {
			case <-owner.Exited():
			case <-ctx.Done():
				if startupWrite != nil {
					// Startup owns partially acquired services. Closing its private
					// control first permits the existing bounded rollback handshake.
					startupWrite.Close()
					timer := time.NewTimer(2 * time.Minute)
					select {
					case <-owner.Exited():
					case <-timer.C:
					}
					timer.Stop()
				}
			}
		}
		if startupWrite != nil {
			startupWrite.Close()
		}
		settleCtx, settleCancel := context.WithTimeout(context.WithoutCancel(ctx), 5*time.Second)
		_, settleErr := owner.Settle(settleCtx, 0)
		settleCancel()
		if settleErr != nil {
			// The native owner retains the unreaped anchor. Output writers and
			// journal dependencies remain owned until later recovery succeeds.
			return result, errors.Join(registrationErr, ctx.Err(), CleanupFailure(settleErr))
		}
		result.Exit = cmd.ProcessState.ExitCode()
		result.Stdout, result.Stderr = output.data, diagnostic.data
		var retireErr error
		if registered {
			retireErr = runtime.RecordResource("helper_process", target, "released")
		}
		if registrationErr != nil || retireErr != nil {
			return result, errors.Join(registrationErr, CleanupFailure(retireErr))
		}
		if ctx.Err() != nil {
			return result, ctx.Err()
		}
		if output.exceeded || diagnostic.exceeded {
			return result, artifactFailure(errors.New("lifecycle output exceeds bound"))
		}
		if waitErr := owner.WaitError(); waitErr != nil {
			return result, waitErr
		}
		return result, nil
	}
}
