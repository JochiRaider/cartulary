//go:build linux

package harnesssession

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessadmission"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
	"github.com/JochiRaider/cartulary/internal/testutil/objectstoreproxy"
	"golang.org/x/sys/unix"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
	"syscall"
	"time"
)

// StopResource invokes a specialized physical adapter through private pipes.
// The caller owns ordering and records release only after this operation succeeds.
func (runtime *Runtime) StopResource(ctx context.Context, root, runRoot string, environment map[string]string, resource Resource) error {
	if _, err := ReadBinding(root, runRoot, runtime.Root, runtime.RunID, runtime.LeaseID); err != nil {
		return err
	}
	return runtime.AttemptCleanup(ctx, resource.Kind, json.RawMessage(resource.Target), func(step context.Context) error {
		return runtime.stopResource(step, root, runRoot, environment, resource)
	})
}
func (runtime *Runtime) stopResource(ctx context.Context, root, runRoot string, environment map[string]string, resource Resource) error {
	if err := runtime.checkResource(resource); err != nil {
		return err
	}
	if resource.Kind == "diagnostic_scope" {
		return runtime.SettleDiagnosticScope(ctx, resource.Target)
	}
	if resource.Kind == "helper_process" || resource.Kind == "browser_process" {
		var proof ProcessProof
		if json.Unmarshal(resource.Target, &proof) != nil {
			return harnesslocal.ErrUnsafe
		}
		return StopProcess(ctx, proof, time.Second)
	}
	if resource.Kind == "object_store_proxy" {
		file, err := resource.Path()
		if err != nil {
			return err
		}
		data, err := harnesslocal.ReadFile(file, 1048576, false)
		if err != nil {
			return err
		}
		snapshot := file + ".cleanup-" + identifier()
		if err := harnesslocal.PublishNew(snapshot, data); err != nil {
			return err
		}
		stopErr := objectstoreproxy.Stop(snapshot)
		parent, err := harnesslocal.OpenDirectory(filepath.Dir(snapshot), true)
		if err != nil {
			return errors.Join(stopErr, err)
		}
		defer parent.Close()
		err = unix.Unlinkat(int(parent.Fd()), filepath.Base(snapshot), 0)
		if errors.Is(err, os.ErrNotExist) {
			err = nil
		}
		return errors.Join(stopErr, err)
	}
	data, err := json.Marshal(resource)
	if err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, environment["NODE_BIN"], filepath.Join(root, "tools/harness/runtime/resource-stop-private.mjs"))
	cmd.Dir = root
	env := maps.Clone(environment)
	maps.Copy(env, runtime.Environment())
	env["CARTULARY_TEST_RESULTS_DIR"] = filepath.Dir(runRoot)
	env["CARTULARY_HARNESS_PHYSICAL_OUTPUT_ROOT"] = runRoot
	env["CARTULARY_TEST_RUN_ID"] = runtime.RunID
	for _, k := range slices.Sorted(maps.Keys(env)) {
		cmd.Env = append(cmd.Env, k+"="+env[k])
	}
	cmd.Stdin = bytes.NewReader(data)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true, Pdeathsig: syscall.SIGKILL}
	cmd.WaitDelay = time.Second
	release, err := harnessadmission.InheritWorkspace(cmd)
	if err != nil {
		return err
	}
	defer release()
	cmd.Cancel = func() error {
		if cmd.Process == nil {
			return nil
		}
		return syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL)
	}
	return waitPhysicalCommand(ctx, cmd, runtime.RetireDetailProducers)
}

func waitPhysicalCommand(ctx context.Context, cmd *exec.Cmd, retire func(ProcessProof) error) error {
	if err := cmd.Start(); err != nil {
		return err
	}
	proof, proofErr := CaptureProcess(cmd.Process.Pid)
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
		// Cancellation is not permission to wait forever for a child whose
		// identity cannot be settled. The native owner still has to prove it.
		settleCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), time.Second)
		settleErr := proofErr
		if settleErr == nil {
			settleErr = StopProcess(settleCtx, proof, 0)
		}
		select {
		case terminalErr = <-terminal:
			cancel()
			if settleErr == nil && retire != nil {
				settleErr = retire(proof)
			}
			return errors.Join(ctx.Err(), terminalErr, settleErr, cmd.Wait())
		case <-settleCtx.Done():
			cancel()
			// Keep the direct-child anchor until it terminates; do not let an
			// unresolved child make the owning cleanup deadline unbounded.
			go func() { <-terminal; _ = cmd.Wait() }()
			return errors.Join(ctx.Err(), settleErr, ErrUnresolved)
		}
	}
	// Keep the direct leader pinned until the shared physical owner has proven
	// the complete group settled. CommandContext cancellation is only a wakeup.
	settleCtx, cancel := context.WithTimeout(context.WithoutCancel(ctx), time.Second)
	settleErr := proofErr
	if settleErr == nil {
		settleErr = StopProcess(settleCtx, proof, 0)
	}
	cancel()
	if settleErr == nil && retire != nil {
		settleErr = retire(proof)
	}
	return errors.Join(terminalErr, settleErr, cmd.Wait())
}
