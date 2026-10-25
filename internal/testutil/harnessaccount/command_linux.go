//go:build linux

package harnessaccount

import (
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessdiag"
)

func childExit(err error) int {
	if err == nil {
		return 0
	}
	if e, ok := err.(*exec.ExitError); ok {
		if status, ok := e.Sys().(syscall.WaitStatus); ok && status.Signaled() {
			return 128 + int(status.Signal())
		}
		if e.ExitCode() >= 0 {
			return e.ExitCode()
		}
	}
	return 2
}

func inheritAdmission(cmd *exec.Cmd) {
	if _, err := os.Stat("/proc/self/fd/9"); err == nil {
		cmd.ExtraFiles = make([]*os.File, 7)
		cmd.ExtraFiles[6] = os.NewFile(9, "workspace-admission")
	}
}

func runChild(cmd *exec.Cmd) (int, int, time.Duration, error) {
	interrupts := make(chan os.Signal, 2)
	signal.Notify(interrupts, syscall.SIGINT, syscall.SIGTERM)
	defer signal.Stop(interrupts)
	start := time.Now()
	if err := cmd.Start(); err != nil {
		return 2, 0, 0, err
	}
	done := make(chan error, 1)
	go func() { done <- cmd.Wait() }()
	observed := 0
	for {
		select {
		case received := <-interrupts:
			sig := received.(syscall.Signal)
			observed = int(sig)
			_ = syscall.Kill(-cmd.Process.Pid, sig)
		case err := <-done:
			return childExit(err), observed, time.Since(start), nil
		}
	}
}

func scoped(args []string, stdout, stderr io.Writer) int {
	setup := time.Now()
	if len(args) < 3 || args[1] != "--" {
		return 2
	}
	g, err := createGroup(args[0])
	if err != nil {
		fmt.Fprintln(stderr, "invocation accounting: delegated hierarchy unavailable")
		return 2
	}
	defer g.close()
	root, err := createPrivateRoot(args[0])
	if err != nil {
		return 2
	}
	// These files only fence accounting. Physical recovery authority remains
	// with the existing fixture owner, including after this controller exits.
	settled := false
	defer os.RemoveAll(root)
	invocation := filepath.Join(root, "invocation")
	if os.Mkdir(invocation, 0700) != nil || harnessdiag.OpenInvocation(invocation) != nil || writeBinding(root, g) != nil {
		return 2
	}
	cmd := exec.Command(args[2], args[3:]...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, stdout, stderr
	cmd.Env = append(os.Environ(), "CARTULARY_HARNESS_INVOCATION_ROOT="+invocation)
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true, UseCgroupFD: true, CgroupFD: int(g.commands.Fd())}
	inheritAdmission(cmd)
	setupDuration := time.Since(setup)
	code, interrupted, elapsed, launchErr := runChild(cmd)
	if launchErr != nil {
		fmt.Fprintln(stderr, "invocation accounting: measured command could not start")
		return code
	}
	publication := time.Now()
	final, counterErr := g.snapshot()
	settled = counterErr == nil && final.Peak >= g.initialPeak && !final.Populated && final.OOMKills == 0 && g.release() == nil
	binary, executableErr := os.Executable()
	currentExecutable := executableErr == nil && sameExecutable("/proc/self/exe", binary)
	terminalStage := "completion"
	completeErr := harnessdiag.CompleteInvocation(invocation, code, interrupted)
	if completeErr == nil {
		terminalStage = "receipts"
		completeErr = harnessdiag.PublishTerminalReceipts(invocation)
	}
	if settled && currentExecutable && completeErr == nil {
		terminalStage = "accounting"
		completeErr = harnessdiag.PublishInvocationAccounting(invocation, map[string]any{
			"duration_ms":  float64(elapsed.Nanoseconds()) / 1e6,
			"cpu_usage_us": float64(final.CPU), "memory_peak_bytes": float64(final.Peak),
			"initial_memory_peak_bytes":        float64(g.initialPeak),
			"controller_setup_ms":              float64(setupDuration.Nanoseconds()) / 1e6,
			"controller_publication_prefix_ms": float64(time.Since(publication).Nanoseconds()) / 1e6,
			"child_exit_code":                  float64(code), "signal": float64(interrupted),
			"scope": "invocation_cgroup", "memory_semantics": "peak_cgroup_charge", "settlement": "empty",
		})
	}
	if !settled || !currentExecutable || completeErr != nil {
		// Report only closed stage/boolean facts; dependency errors may contain
		// private paths or values and must not enter retained diagnostics.
		fmt.Fprintf(stderr, "invocation accounting: incomplete; scope_settled=%t controller_current=%t terminal_stage=%s terminal_failed=%t; no complete resource claim\n", settled, currentExecutable, terminalStage, completeErr != nil)
		if code == 0 {
			code = 1
		}
	}
	return code
}

func Main(args []string, stdout, stderr io.Writer) int {
	if len(args) == 1 && args[0] == "verify" {
		parent, err := ContainerParent()
		if err != nil || parent == "" {
			return 2
		}
		return 0
	}
	if len(args) == 1 && args[0] == "container-parent" {
		parent, err := ContainerParent()
		if err != nil {
			return 2
		}
		fmt.Fprintln(stdout, parent)
		return 0
	}
	if len(args) == 2 && args[0] == "verify-container" {
		parent, err := ContainerParent()
		if err != nil || parent == "" || VerifyContainer(parent, args[1]) != nil {
			return 2
		}
		return 0
	}
	if len(args) > 0 && args[0] == "scope" {
		return scoped(args[1:], stdout, stderr)
	}
	if len(args) < 2 || args[0] != "--" || os.Getenv("HARNESS_DIAGNOSTICS") != "basic" || os.Getenv("CARTULARY_HARNESS_CACHE_MODE") != "off" {
		return 2
	}
	if parent, err := ContainerParent(); err != nil {
		return 2
	} else if parent != "" {
		invocation, err := InvocationRoot()
		if err != nil {
			return 2
		}
		command, err := exec.LookPath(args[1])
		if err != nil {
			return 2
		}
		if syscall.Exec(command, args[1:], append(os.Environ(), "CARTULARY_HARNESS_INVOCATION_ROOT="+invocation)) != nil {
			return 2
		}
	}
	id, err := os.ReadFile("/proc/sys/kernel/random/uuid")
	if err != nil {
		return 2
	}
	scope := "cartulary-account-" + strings.TrimSpace(string(id)) + ".scope"
	binary, err := os.Executable()
	if err != nil {
		return 2
	}
	command := []string{"--user", "--quiet", "--scope", "--property=Delegate=yes", "--unit=" + scope, "--same-dir", binary, "account", "scope", scope}
	command = append(command, args...)
	cmd := exec.Command("systemd-run", command...)
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, stdout, stderr
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	inheritAdmission(cmd)
	code, _, _, err := runChild(cmd)
	if err != nil {
		fmt.Fprintln(stderr, "invocation accounting: delegated scope unavailable")
	}
	return code
}
