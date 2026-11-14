//go:build linux

package harnessfixtures

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnesssession"
)

func testFixtureCommandOwnership(t *testing.T) {
	t.Run("unacquired_commands_leave_no_resources", func(t *testing.T) {
		runtime, root, _ := runtimeFixture(t)
		run := commandRunner(root, runtime, []string{})
		ctx, cancel := context.WithCancel(t.Context())
		cancel()
		result, err := run(ctx, "/bin/sh", []string{"-c", "exit 0"}, map[string]string{})
		if !errors.Is(err, context.Canceled) || result.Started || result.Exit != -1 {
			t.Fatal("cancelled launch acquired command", result, err)
		}
		result, err = run(t.Context(), filepath.Join(root, "missing"), nil, map[string]string{})
		if !errors.Is(err, os.ErrNotExist) || result.Started || result.Exit != -1 {
			t.Fatal("missing executable acquired command", result, err)
		}
		resources, err := runtime.Resources(64)
		if err != nil || len(resources) != 0 {
			t.Fatal("failed acquisition published resources", resources, err)
		}
	})
	t.Run("exited_leader_settles_descendant_and_preserves_peer", func(t *testing.T) {
		runtime, root, _ := runtimeFixture(t)
		peer := exec.Command("sleep", "60")
		peer.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		owner, err := harnesssession.StartOwnedCommand(peer, nil)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			if _, err := owner.Settle(ctx, 0); err != nil {
				t.Error(err)
			}
		})
		result, err := commandRunner(root, runtime, []string{})(t.Context(), "/bin/sh", []string{"-c", "sleep 60 </dev/null >/dev/null 2>&1 & echo $$ $!; echo diagnostic >&2; exit 17"}, map[string]string{"PATH": "/usr/bin:/bin"})
		var exit *exec.ExitError
		if !errors.As(err, &exit) || exit.ExitCode() != 17 || result.Exit != 17 || !result.Started || string(result.Stderr) != "diagnostic\n" {
			t.Fatal("native outcome lost", result, err)
		}
		fields := strings.Fields(string(result.Stdout))
		if len(fields) != 2 {
			t.Fatal("missing exact process identities", string(result.Stdout))
		}
		for i, field := range fields {
			pid, err := strconv.Atoi(field)
			if err != nil || pid < 2 {
				t.Fatal("invalid process identity", field)
			}
			stat, err := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
			if errors.Is(err, os.ErrNotExist) {
				continue
			}
			if err != nil {
				t.Fatal(err)
			}
			if i == 0 {
				t.Fatal("direct child returned without reaping", pid)
			}
			boundary := strings.LastIndex(string(stat), ") ")
			if boundary < 0 {
				t.Fatal("invalid descendant observation")
			}
			state := strings.Fields(string(stat)[boundary+2:])
			if len(state) == 0 || state[0] != "Z" && state[0] != "X" {
				t.Fatal("live descendant survived return", pid)
			}
		}
		proof, err := harnesssession.CaptureProcess(peer.Process.Pid)
		if err != nil || proof != owner.Proof() {
			t.Fatal("unrelated peer changed", err)
		}
		select {
		case <-owner.Exited():
			t.Fatal("unrelated peer terminated")
		default:
		}
		resources, err := runtime.Resources(64)
		if err != nil || len(resources) != 0 {
			t.Fatal("settled helper retained journal entry", resources, err)
		}
	})
	t.Run("publication_failure_does_not_admit_startup", func(t *testing.T) {
		runtime, root, _ := runtimeFixture(t)
		marker := filepath.Join(root, "admitted")
		script := filepath.Join(root, "startup")
		body := "#!/bin/sh\nif test \"$(dd bs=1 count=1 <&3 2>/dev/null | od -An -tu1 | tr -d ' ')\" = 1; then touch \"$2\"; fi\nsleep 60\n"
		if err := os.WriteFile(script, []byte(body), 0700); err != nil {
			t.Fatal(err)
		}
		result, err := commandRunner(root, runtime, []string{"invalid-dependency"})(t.Context(), script, []string{"start-suite", marker}, map[string]string{"PATH": "/usr/bin:/bin"})
		if err == nil || !result.Started {
			t.Fatal("publication rejection lost", result, err)
		}
		if _, err := os.Stat(marker); !errors.Is(err, os.ErrNotExist) {
			t.Fatal("unregistered helper was admitted", err)
		}
		resources, err := runtime.Resources(64)
		if err != nil || len(resources) != 0 {
			t.Fatal("failed registration leaked resources", resources, err)
		}
	})
	t.Run("diagnostics_remain_bounded", func(t *testing.T) {
		runtime, root, _ := runtimeFixture(t)
		result, err := commandRunner(root, runtime, []string{})(t.Context(), "/bin/sh", []string{"-c", "head -c 262145 /dev/zero >&2"}, map[string]string{"PATH": "/usr/bin:/bin"})
		var failure *Failure
		if !errors.As(err, &failure) || failure.Class != "artifact" || len(result.Stderr) != 262144 || len(result.Stdout) != 0 || result.Exit != 0 {
			t.Fatal("diagnostic bound/classification lost", len(result.Stderr), result.Exit, err)
		}
	})
}
