// Package renderprocess supervises an external renderer in a private Linux
// namespace. Reporting owns input admission, artifact validation and publication.
// Namespace confinement is not a substitute for renderer/network observations.
package renderprocess

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"golang.org/x/sys/unix"
)

var (
	ErrConfiguration = errors.New("render process configuration invalid")
	ErrStart         = errors.New("render process sandbox unavailable")
	ErrExecution     = errors.New("render process execution failed")
	ErrOutput        = errors.New("render process output invalid")
	ErrOutputLimit   = fmt.Errorf("%w: size limit exceeded", ErrOutput)
	ErrCleanup       = errors.New("render process cleanup failed")
)

const terminationGrace = 2 * time.Second
const protocolLimit = 64 << 10

// ExecutionError retains only the process exit status, never stderr or argv.
type ExecutionError struct {
	ExitCode int
	Reason   string
}

func (e *ExecutionError) Error() string {
	return ErrExecution.Error() + ": " + e.Reason + ", exit " + strconv.Itoa(e.ExitCode)
}
func (e *ExecutionError) Unwrap() error { return ErrExecution }

// Runtime is assembled from the installed, verified runtime distribution. Root
// must contain only the declared renderer, libraries, packages and fonts. It is
// never the application root, a checkout, or the host filesystem. No caller
// supplied template value is permitted to select these capabilities.
type Runtime struct {
	SandboxCommand []string
	Root           string
	TemporaryRoot  string
	Command        []string
	Launcher       []string
	LoopbackPort   uint16
	Environment    map[string]string
}

type Output struct {
	Path     string
	MaxBytes int64
}

type Result struct {
	Protocol            []byte
	Files               map[string][]byte
	NetworkObservations []NetworkObservation
}

// Run owns the entire attempt directory and returns bytes only after the child
// exited, all declared outputs were read, and cleanup succeeded. It never emits
// raw renderer stderr, host paths, input bytes or subprocess error strings.
func (r Runtime) Run(ctx context.Context, input []byte, outputs []Output) (result Result, err error) {
	if err := r.validate(outputs); err != nil {
		return Result{}, err
	}
	if err := ctx.Err(); err != nil {
		return Result{}, err
	}
	attempt, err := os.MkdirTemp(r.TemporaryRoot, "report-render-")
	if err != nil {
		return Result{}, ErrStart
	}
	defer func() {
		if cleanupErr := os.RemoveAll(attempt); cleanupErr != nil {
			result = Result{}
			err = errors.Join(err, ErrCleanup)
		}
		if contextErr := ctx.Err(); contextErr != nil {
			result = Result{}
			err = errors.Join(err, contextErr)
		}
	}()
	work := filepath.Join(attempt, "work")
	for _, directory := range []string{work, filepath.Join(work, "tmp"), filepath.Join(work, "home"), filepath.Join(attempt, "diagnostics")} {
		if err := os.Mkdir(directory, 0o700); err != nil {
			return Result{}, ErrStart
		}
	}
	args := []string{
		"--unshare-user", "--unshare-pid", "--unshare-net", "--unshare-ipc", "--unshare-uts", "--unshare-cgroup",
		"--die-with-parent", "--cap-drop", "ALL",
		"--clearenv", "--ro-bind", r.Root, "/", "--proc", "/proc",
		// Headless rendering needs no terminal/PTY namespace. A fixed minimal
		// device directory avoids devpts' additional user namespace and exposes
		// no host device tree. Chromium shared memory uses the private /tmp.
		"--tmpfs", "/dev", "--dev-bind", "/dev/null", "/dev/null",
		"--dev-bind", "/dev/zero", "/dev/zero", "--dev-bind", "/dev/random", "/dev/random",
		"--dev-bind", "/dev/urandom", "/dev/urandom", "--symlink", "/proc/self/fd", "/dev/fd",
		"--symlink", "/proc/self/fd/0", "/dev/stdin", "--symlink", "/proc/self/fd/1", "/dev/stdout",
		"--symlink", "/proc/self/fd/2", "/dev/stderr", "--remount-ro", "/dev",
		"--bind", work, "/work", "--bind", filepath.Join(attempt, "diagnostics"), "/diagnostics",
		"--bind", filepath.Join(work, "tmp"), "/tmp", "--chdir", "/work",
	}
	// Sorting is part of constructing an auditable, repeatable argv. Values are
	// explicit installation capabilities, never inherited from the server.
	for _, key := range sortedEnvironmentKeys(r.Environment) {
		args = append(args, "--setenv", key, r.Environment[key])
	}
	args = append(args, "--setenv", "CARTULARY_RENDER_LOOPBACK_PORT", strconv.Itoa(int(r.LoopbackPort)))
	args = append(args, "--")
	args = append(args, r.Launcher...)
	args = append(args, r.Command...)
	channel, err := unix.Socketpair(unix.AF_UNIX, unix.SOCK_SEQPACKET|unix.SOCK_CLOEXEC, 0)
	if err != nil {
		return Result{}, ErrStart
	}
	childChannel := os.NewFile(uintptr(channel[1]), "render-network-child")
	defer childChannel.Close()
	command := exec.Command(r.SandboxCommand[0], append(append([]string{}, r.SandboxCommand[1:]...), args...)...)
	command.ExtraFiles = []*os.File{childChannel}
	command.Env = []string{}
	command.Stdin = bytes.NewReader(input)
	var protocol, diagnostic boundedProtocol
	command.Stdout = &protocol
	command.Stderr = &diagnostic
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true, Pdeathsig: syscall.SIGKILL}
	command.WaitDelay = terminationGrace
	// Linux parent-death signals are tied to the creating thread. Keep it alive
	// for the full owned process lifetime, including cancellation and reaping.
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	if err := command.Start(); err != nil {
		_ = unix.Close(channel[0])
		return Result{}, ErrStart
	}
	_ = childChannel.Close()
	networkCtx, stopNetwork := context.WithCancel(context.Background())
	defer stopNetwork()
	type networkResult struct {
		observations []NetworkObservation
		err          error
	}
	networkDone := make(chan networkResult, 1)
	go func() {
		observations, err := superviseNetwork(networkCtx, channel[0], r.LoopbackPort)
		networkDone <- networkResult{observations, err}
	}()
	done := make(chan error, 1)
	go func() { done <- command.Wait() }()
	var network networkResult
	select {
	case waitErr := <-done:
		// All filtered descendants have exited and the listener will report HUP.
		// Drain it before accepting any bytes, including on an ordinary exit.
		network = <-networkDone
		if ctx.Err() != nil {
			return Result{}, ctx.Err()
		}
		if waitErr != nil {
			return Result{}, errors.Join(&ExecutionError{ExitCode: command.ProcessState.ExitCode(), Reason: safeExitReason(diagnostic.Bytes())}, network.err)
		}
	case network = <-networkDone:
		if network.err != nil {
			terminateProcessGroup(command.Process.Pid, done)
			return Result{}, errors.Join(network.err, &ExecutionError{ExitCode: command.ProcessState.ExitCode(), Reason: safeExitReason(diagnostic.Bytes())})
		}
		select {
		case waitErr := <-done:
			if waitErr != nil {
				return Result{}, &ExecutionError{ExitCode: command.ProcessState.ExitCode(), Reason: safeExitReason(diagnostic.Bytes())}
			}
		case <-ctx.Done():
			terminateProcessGroup(command.Process.Pid, done)
			return Result{}, ctx.Err()
		}
	case <-ctx.Done():
		// The private PID namespace also prevents a descendant from surviving by
		// creating its own process group. Killing its init tears down the tree.
		terminateProcessGroup(command.Process.Pid, done)
		stopNetwork()
		<-networkDone
		return Result{}, ctx.Err()
	}
	if network.err != nil {
		return Result{}, network.err
	}
	if protocol.overflow {
		return Result{}, ErrOutput
	}
	files := make(map[string][]byte, len(outputs))
	for _, output := range outputs {
		content, err := readOutput(work, output)
		if err != nil {
			return Result{}, err
		}
		files[output.Path] = content
	}
	return Result{Protocol: bytes.Clone(protocol.Bytes()), Files: files, NetworkObservations: network.observations}, nil
}

func terminateProcessGroup(pid int, done <-chan error) {
	_ = syscall.Kill(-pid, syscall.SIGTERM)
	timer := time.NewTimer(terminationGrace)
	defer timer.Stop()
	select {
	case <-done:
	case <-timer.C:
		_ = syscall.Kill(-pid, syscall.SIGKILL)
		<-done
	}
}

// Stderr can contain incident data. Only these fixed classifications leave the
// supervisor; neither matched lines nor arbitrary tokens are exposed.
func safeExitReason(data []byte) string {
	for _, reason := range []string{"stage_browser_library_missing", "stage_browser_crashpad_failed", "stage_browser_operation_denied", "stage_browser_launch_failed", "renderer_server_start_failed", "renderer_page_load_failed", "renderer_print_surface_missing", "renderer_fonts_or_paint_failed", "renderer_pdf_export_failed", "browser_render_error", "browser_module_export_missing", "browser_slide_import_failed", "renderer_resource_missing", "renderer_resource_transform_failed", "renderer_resource_load_failed", "renderer_svg_export_failed", "invalid_mermaid_construct", "click_export_page_count_mismatch", "outbound_request_observed", "loopback_not_declared", "external_reference_in_rendered_output"} {
		if bytes.Contains(data, []byte("report-render-stage: "+reason+"\n")) {
			return reason
		}
	}
	for _, item := range []struct{ pattern, reason string }{
		{"ResetStdio", "runtime_stdio_setup_failed"},
		{"uv_thread_create", "runtime_thread_setup_failed"},
		{"error while loading shared libraries", "runtime_library_unavailable"},
		{"Cannot find package", "runtime_package_unavailable"},
		{"Cannot find module", "runtime_module_unavailable"},
	} {
		if bytes.Contains(data, []byte(item.pattern)) {
			return item.reason
		}
	}
	if bytes.Contains(data, []byte("bwrap: ")) {
		for _, item := range []struct{ pattern, reason string }{
			{"cannot open /proc/sys/user/max_user_namespaces", "sandbox_namespace_lock_unavailable"},
			{"sysctl user.max_user_namespaces = 1", "sandbox_namespace_lock_failed"},
			{"creation of new user namespaces was not disabled", "sandbox_namespace_lock_ineffective"},
			{"No permissions to create a new namespace", "sandbox_namespace_create_denied"},
			{"nesting depth or /proc/sys/user/max_", "sandbox_namespace_capacity_exhausted"},
			{"kernel does not support user namespaces", "sandbox_namespace_kernel_unsupported"},
			{"unshare user ns", "sandbox_nested_namespace_failed"},
			{"Creating new namespace failed", "sandbox_namespace_create_failed"},
			{"userns", "sandbox_namespace_unavailable"},
			{"namespace", "sandbox_namespace_unavailable"},
			{"mount", "sandbox_mount_failed"},
			{"Creating", "sandbox_filesystem_setup_failed"},
			{"Can't create", "sandbox_filesystem_setup_failed"},
			{"exec", "sandbox_executable_unavailable"},
		} {
			if bytes.Contains(data, []byte(item.pattern)) {
				return item.reason
			}
		}
		return "sandbox_setup_failed"
	}
	return "stage_failed"
}

type boundedProtocol struct {
	bytes.Buffer
	overflow bool
}

func (b *boundedProtocol) Write(value []byte) (int, error) {
	length := len(value)
	remaining := protocolLimit - b.Len()
	if len(value) > remaining {
		b.overflow = true
		value = value[:remaining]
	}
	_, _ = b.Buffer.Write(value)
	return length, nil
}

func readOutput(work string, output Output) ([]byte, error) {
	root, err := os.OpenRoot(work)
	if err != nil {
		return nil, ErrOutput
	}
	defer root.Close()
	current := ""
	for _, segment := range strings.Split(output.Path, "/") {
		current = filepath.Join(current, segment)
		info, err := root.Lstat(current)
		if err != nil || info.Mode()&os.ModeSymlink != 0 {
			return nil, ErrOutput
		}
	}
	file, err := root.OpenFile(current, os.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_NONBLOCK, 0)
	if err != nil {
		return nil, ErrOutput
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() {
		return nil, ErrOutput
	}
	stat, ok := info.Sys().(*syscall.Stat_t)
	if !ok || stat.Nlink != 1 {
		return nil, ErrOutput
	}
	if info.Size() > output.MaxBytes {
		return nil, ErrOutputLimit
	}
	content, err := io.ReadAll(io.LimitReader(file, output.MaxBytes+1))
	if err != nil {
		return nil, ErrOutput
	}
	if int64(len(content)) > output.MaxBytes {
		return nil, ErrOutputLimit
	}
	return content, nil
}
