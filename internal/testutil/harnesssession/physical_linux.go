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
	return cmd.Run()
}
