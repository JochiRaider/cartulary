//go:build linux

package harnessruntime

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"syscall"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
	"github.com/JochiRaider/cartulary/internal/testutil/harnessschema"
	"golang.org/x/sys/unix"
)

const launchContextKey = "CARTULARY_HARNESS_LAUNCH_CONTEXT"
const failureContextKey = "CARTULARY_HARNESS_COMMAND_FAILURE_CONTEXT"

type processObservation interface {
	Spawned(int)
	Settled(string)
}
type failureScope interface {
	Environment() map[string]string
	Read() *failure
	Close() error
}
type processRunner struct {
	Root         string
	Environment  map[string]string
	Commands     map[string]string
	FailureScope func(map[string]any) (failureScope, error)
	Observe      func(map[string]any) processObservation
	KillGrace    time.Duration
}
type boundedTail struct {
	bytes []byte
	limit int
}

func (t *boundedTail) Write(p []byte) (int, error) {
	n := len(p)
	if n >= t.limit {
		t.bytes = append(t.bytes[:0], p[n-t.limit:]...)
	} else {
		if len(t.bytes)+n > t.limit {
			t.bytes = slices.Delete(t.bytes, 0, len(t.bytes)+n-t.limit)
		}
		t.bytes = append(t.bytes, p...)
	}
	return n, nil
}
func uuid() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		panic(err)
	}
	b[6] = (b[6] & 15) | 64
	b[8] = (b[8] & 63) | 128
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[:4], b[4:6], b[6:8], b[8:10], b[10:])
}
func documentBytes(schema string, v any) ([]byte, error) {
	b, err := json.Marshal(v)
	if err != nil {
		return nil, err
	}
	normalized, err := harnessschema.Decode(b)
	if err == nil {
		err = harnessschema.Validate(schema, normalized)
	}
	return b, err
}
func freshLaunch(u unit, environment map[string]string, commands map[string]string) (map[string]any, error) {
	run := environment["CARTULARY_TEST_RUN_ID"]
	if run == "" {
		run = "standalone"
	}
	var parent, commandID any
	if text := environment[launchContextKey]; text != "" {
		if inherited, err := harnessschema.Decode([]byte(text)); err == nil && harnessschema.Validate("cartulary.harness_launch_context.v1", inherited) == nil {
			context := inherited.(map[string]any)
			if context["run_id"] == run {
				parent = context["invocation_id"]
			}
		}
	}
	target := u.Command.Environment["CARTULARY_TEST_TARGET"]
	id := runtimeCommandID(target)
	if commands != nil {
		id = commands[target]
	}
	if id != "" {
		commandID = id
	}
	identity := map[string]any{"schema_id": "cartulary.harness_launch_context.v1", "run_id": run, "invocation_id": uuid(), "parent_invocation_id": parent, "unit_id": u.ID, "command_id": commandID, "row_ids": rowIDs(u), "producer": "unit"}
	_, err := documentBytes("cartulary.harness_launch_context.v1", identity)
	return identity, err
}
func (p *processRunner) Execute(ctx context.Context, request executionRequest) (result unitResult) {
	u := request.Unit
	environment := maps.Clone(p.Environment)
	if environment == nil {
		environment = map[string]string{}
	}
	maps.Copy(environment, request.Environment)
	maps.Copy(environment, u.Command.Environment)
	result = unitResult{Status: "failed"}
	fail := func(f *failure) {
		result.Failure = f
		result.ExitCode = failureExitCode(f, result.Signal)
		if f.Class == "interrupted" {
			result.Status = "cancelled"
		} else {
			result.Status = "failed"
		}
	}
	identity, err := freshLaunch(u, environment, p.Commands)
	if err != nil {
		fail(failed("harness", "scheduler_accounting_error"))
		return
	}
	launchBytes, _ := json.Marshal(identity)
	environment[launchContextKey] = string(launchBytes)
	delete(environment, failureContextKey)
	var observation processObservation
	if p.Observe != nil {
		observation = p.Observe(identity)
	}
	outcome := "spawn_failed"
	defer func() {
		if observation != nil {
			observation.Settled(outcome)
		}
	}()
	var scope failureScope
	if p.FailureScope != nil && identity["command_id"] != nil {
		scope, err = p.FailureScope(identity)
		if err != nil {
			fail(failed("harness", "scheduler_accounting_error"))
			return
		}
		maps.Copy(environment, scope.Environment())
		defer func() {
			if err := scope.Close(); err != nil {
				secondary := failed("harness", "cleanup_error")
				secondary.Lifecycle = "cleanup_finalizers"
				result.Cleanup = append(result.Cleanup, secondary)
				if result.Failure == nil {
					fail(secondary)
				}
			}
		}()
	}
	executable := u.Command.Executable
	if node := environment["NODE_BIN"]; node != "" || executable == "node" {
		if node == "" {
			fail(failed("config", "configuration_error"))
			return
		}
		if !filepath.IsAbs(node) {
			node = filepath.Join(p.Root, node)
		}
		info, err := os.Stat(node)
		if err != nil || !info.Mode().IsRegular() || info.Mode().Perm()&0111 == 0 {
			fail(failed("config", "configuration_error"))
			return
		}
		environment["NODE_BIN"] = node
		directory := filepath.Dir(node)
		paths := strings.Split(environment["PATH"], string(os.PathListSeparator))
		for len(paths) > 0 && paths[0] == directory {
			paths = paths[1:]
		}
		environment["PATH"] = strings.Join(append([]string{directory}, paths...), string(os.PathListSeparator))
		if executable == "node" {
			executable = node
		}
	}
	if !strings.ContainsRune(executable, '/') {
		found := ""
		for _, directory := range filepath.SplitList(environment["PATH"]) {
			if directory == "" || !filepath.IsAbs(directory) {
				continue
			}
			candidate := filepath.Join(directory, executable)
			if info, err := os.Stat(candidate); err == nil && info.Mode().IsRegular() && info.Mode().Perm()&0111 != 0 {
				found = candidate
				break
			}
		}
		if found == "" {
			fail(failed("infra", "service_start_error"))
			return
		}
		executable = found
	}
	cmd := exec.Command(executable, u.Command.Args...)
	cmd.Dir = p.Root
	keys := slices.Sorted(maps.Keys(environment))
	for _, key := range keys {
		cmd.Env = append(cmd.Env, key+"="+environment[key])
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true, Pdeathsig: syscall.SIGKILL}
	stdout, stderr := &boundedTail{limit: 1048576}, &boundedTail{limit: 1048576}
	cmd.Stdout = stdout
	cmd.Stderr = stderr
	closeDescriptors, err := inheritWorkspaceAdmission(cmd)
	if err != nil {
		fail(failed("infra", "resource_conflict"))
		return
	}
	defer closeDescriptors()
	if ctx.Err() != nil {
		outcome = "cancelled"
		fail(failed("interrupted", "cancelled_or_interrupted"))
		return
	}
	if err = cmd.Start(); err != nil {
		if u.Command.Executable == "node" {
			fail(failed("config", "configuration_error"))
		} else {
			fail(failed("infra", "service_start_error"))
		}
		return
	}
	outcome = "failed"
	if observation != nil {
		observation.Spawned(cmd.Process.Pid)
	}
	signalGroup := func(signal syscall.Signal) { _ = syscall.Kill(-cmd.Process.Pid, signal) }
	if request.Host != nil && request.Host.Bind != nil {
		if err := request.Host.Bind(cmd.Process.Pid); err != nil {
			fail(failed("infra", "resource_conflict"))
			signalGroup(syscall.SIGKILL)
		}
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
	grace := p.KillGrace
	if grace <= 0 {
		grace = 2 * time.Second
	}
	var kill *time.Timer
	var killC <-chan time.Time
	deadline := time.NewTimer(time.Duration(u.TimeoutMS) * time.Millisecond)
	defer deadline.Stop()
	cancelled, timedOut := false, false
	cancellation := ctx.Done()
	timeout := deadline.C
	terminate := func() {
		signalGroup(syscall.SIGTERM)
		if kill == nil {
			kill = time.NewTimer(grace)
			killC = kill.C
		}
	}
	var terminalErr error
	waiting := true
	for waiting {
		select {
		case terminalErr = <-terminal:
			waiting = false
		case <-cancellation:
			cancelled = true
			cancellation = nil
			terminate()
		case <-timeout:
			timedOut = true
			timeout = nil
			terminate()
		case <-killC:
			signalGroup(syscall.SIGKILL)
			killC = nil
		}
	}
	if kill != nil {
		kill.Stop()
	}
	// WNOWAIT pins the leader until every signal has been issued. An ordinary
	// exit also retires lingering group members before pipe collection and reap.
	signalGroup(syscall.SIGKILL)
	cmd.WaitDelay = grace
	waitErr := cmd.Wait()
	result.Stdout = stdout.bytes
	result.Stderr = stderr.bytes
	if cmd.ProcessState != nil {
		result.ExitCode = cmd.ProcessState.ExitCode()
		if status, ok := cmd.ProcessState.Sys().(syscall.WaitStatus); ok && status.Signaled() {
			result.Signal = unix.SignalName(status.Signal())
		}
	}
	if result.Failure == nil {
		if cancelled {
			fail(failed("interrupted", "cancelled_or_interrupted"))
			outcome = "cancelled"
		} else if timedOut {
			fail(failed("timing", "timeout_failure"))
			outcome = "timeout"
		} else if terminalErr != nil {
			fail(failed("harness", "scheduler_accounting_error"))
			outcome = "failed"
		} else if waitErr == nil {
			result.Status = "passed"
			outcome = "passed"
		} else {
			childExit := result.ExitCode
			fail(classifyProcessFailure(u, result.ExitCode, result.Stdout, result.Stderr))
			result.ExitCode = childExit
			outcome = "failed"
		}
	}
	if u.Kind == "lifecycle" && result.Status != "passed" && !cancelled && !timedOut {
		childExit := result.ExitCode
		fail(retainedLifecycleFailure(u, p.Root, environment))
		result.ExitCode = childExit
		outcome = "failed"
	}
	if scope != nil && !cancelled && !timedOut {
		if diagnosed := scope.Read(); diagnosed != nil {
			if result.Status == "passed" {
				diagnosed = failed("harness", "scheduler_accounting_error")
			}
			fail(diagnosed)
			outcome = "failed"
		}
	}
	return
}

var failureClassMarker = regexp.MustCompile(`(?:^|\s)failure_class=([a-z][a-z0-9_]*)`)
var failureReasonMarker = regexp.MustCompile(`(?:^|\s)(?:failure_reason|reason)=([a-z][a-z0-9_]*)`)

func classifyProcessFailure(u unit, code int, stdout, stderr []byte) *failure {
	text := string(stdout) + "\n" + string(stderr)
	classes := failureClassMarker.FindAllStringSubmatch(text, -1)
	reasons := failureReasonMarker.FindAllStringSubmatch(text, -1)
	for i := len(classes) - 1; i >= 0; i-- {
		if slices.Contains(failureClasses, classes[i][1]) {
			reason := "unknown_failure"
			if len(reasons) > 0 {
				reason = reasons[len(reasons)-1][1]
			}
			return normalizedFailure(failed(classes[i][1], reason))
		}
	}
	if code == 2 && filepath.Base(u.Command.Executable) != "make" {
		return failed("config", "configuration_error")
	}
	switch code {
	case 3:
		return failed("infra", "preflight_error")
	case 10:
		return failed("product", "test_assertion_failure")
	case 11:
		return failed("artifact", "artifact_error")
	case 13:
		return failed("timing", "timeout_failure")
	}
	if u.Kind == "lifecycle" {
		return failed("harness", "fixture_error")
	}
	if filepath.Base(u.Command.Executable) == "make" {
		return failed("harness", "child_target_failure")
	}
	return failed("product", "test_assertion_failure")
}
func retainedLifecycleFailure(u unit, root string, environment map[string]string) *failure {
	for _, name := range u.Evidence {
		if !strings.HasSuffix(name, ".attempt.json") {
			continue
		}
		if !safeRelative(name) || environment["CARTULARY_TEST_RESULTS_DIR"] == "" || environment["CARTULARY_TEST_RUN_ID"] == "" {
			break
		}
		run := filepath.Join(root, environment["CARTULARY_TEST_RESULTS_DIR"], environment["CARTULARY_TEST_RUN_ID"])
		if filepath.IsAbs(environment["CARTULARY_TEST_RESULTS_DIR"]) {
			run = filepath.Join(environment["CARTULARY_TEST_RESULTS_DIR"], environment["CARTULARY_TEST_RUN_ID"])
		}
		bytes, err := harnesslocal.ReadFile(filepath.Join(run, name), 1024*1024, false)
		if err != nil {
			break
		}
		v, err := harnessschema.Decode(bytes)
		if err != nil || harnessschema.Validate("cartulary.browser_reset_attempt.v2", v) != nil {
			break
		}
		record := v.(map[string]any)
		if record["status"] != "fail" {
			break
		}
		class, _ := record["failure_class"].(string)
		reason, _ := record["failure_reason"].(string)
		return normalizedFailure(failed(class, reason))
	}
	return failed("artifact", "artifact_error")
}
func inheritWorkspaceAdmission(cmd *exec.Cmd) (func(), error) {
	close := func() {}
	bytes, err := os.ReadFile("/proc/self/fdinfo/9")
	if err != nil || !strings.Contains(string(bytes), "FLOCK") || !strings.Contains(string(bytes), "READ") {
		return close, nil
	}
	fd, err := unix.FcntlInt(9, unix.F_DUPFD_CLOEXEC, 0)
	if err != nil {
		return close, err
	}
	admission := os.NewFile(uintptr(fd), "workspace-admission")
	null, err := os.OpenFile(os.DevNull, os.O_RDWR, 0)
	if err != nil {
		admission.Close()
		return close, err
	}
	for len(cmd.ExtraFiles) < 6 {
		cmd.ExtraFiles = append(cmd.ExtraFiles, null)
	}
	if len(cmd.ExtraFiles) != 6 {
		admission.Close()
		null.Close()
		return close, errors.New("reserved workspace descriptor is occupied")
	}
	cmd.ExtraFiles = append(cmd.ExtraFiles, admission)
	return func() { admission.Close(); null.Close() }, nil
}
