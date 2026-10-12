//go:build linux

package harnesssession

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessadmission"
	"maps"
	"os/exec"
	"path/filepath"
	"slices"
	"syscall"
	"time"
)

// StopResource invokes a specialized physical adapter through private pipes.
// The caller owns ordering and records release only after this operation succeeds.
func (runtime *Runtime) StopResource(ctx context.Context, root, runRoot string, environment map[string]string, resource Resource) error {
	data, err := json.Marshal(resource)
	if err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx, environment["NODE_BIN"], filepath.Join(root, "tools/harness/runtime/resource-stop-private.mjs"))
	cmd.Dir = root
	env := maps.Clone(environment)
	maps.Copy(env, runtime.Environment())
	env["CARTULARY_TEST_RESULTS_DIR"] = filepath.Dir(runRoot)
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
