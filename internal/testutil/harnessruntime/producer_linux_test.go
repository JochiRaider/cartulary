//go:build linux

package harnessruntime

import (
	"context"
	"encoding/json"
	"maps"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessfixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesssession"
)

// These tests exercise the retained Node producers through their actual Go
// executor. No test-only runner is installed or exposed as a runtime operation.
func TestRuntimeProducerInterop(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../.."))
	node := filepath.Join(root, ".cache/cartulary/node-runtime/bin/node")
	t.Run("selected_node_cannot_be_rebound_by_unit_or_fixture", func(t *testing.T) {
		scratch := t.TempDir()
		competing := filepath.Join(scratch, "node")
		if err := os.WriteFile(competing, []byte("#!/bin/sh\nprintf competing"), 0700); err != nil {
			t.Fatal(err)
		}
		p := processRunner{Root: root, Environment: map[string]string{"NODE_BIN": node, "PATH": scratch}}
		script := `const {execFileSync}=require('node:child_process');process.stdout.write(JSON.stringify([process.execPath,execFileSync('node',['-p','process.execPath'],{encoding:'utf8'}).trim(),process.env.NODE_BIN,process.env.PATH]))`
		for _, executable := range []string{"node", node} {
			u := engineUnit("test:node", 1)
			u.Command = command{Executable: executable, Args: []string{"-e", script}, Environment: map[string]string{"NODE_BIN": competing, "PATH": scratch}}
			u.TimeoutMS = 10000
			outcome := p.Execute(context.Background(), executionRequest{Unit: u, Environment: map[string]string{"NODE_BIN": competing}})
			if outcome.Status != "passed" {
				t.Fatalf("%+v %s", outcome, outcome.Stderr)
			}
			var values []string
			if err := json.Unmarshal(outcome.Stdout, &values); err != nil {
				t.Fatal(err)
			}
			actual, _ := filepath.EvalSymlinks(node)
			direct, _ := filepath.EvalSymlinks(values[0])
			nested, _ := filepath.EvalSymlinks(values[1])
			if direct != actual || nested != actual || values[2] != node || values[3] != filepath.Dir(node)+":"+scratch {
				t.Fatal(values)
			}
		}
		u := engineUnit("test:missing-node", 1)
		u.Command.Executable = "node"
		for _, selected := range []string{"", filepath.Join(scratch, "absent"), scratch} {
			p.Environment["NODE_BIN"] = selected
			r := p.Execute(context.Background(), executionRequest{Unit: u})
			if r.Failure == nil || r.Failure.Reason != "configuration_error" || r.ExitCode != 2 || len(r.Stdout) != 0 {
				t.Fatal(r)
			}
		}
	})
	t.Run("private_diagnostic_survives_make_and_contradiction", func(t *testing.T) {
		scratch := t.TempDir()
		run := filepath.Join(scratch, "run")
		if err := os.Mkdir(run, 0700); err != nil {
			t.Fatal(err)
		}
		rt, err := harnesssession.NewRuntime(root, run, "run", filepath.Join(scratch, "private"))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if err := rt.Close(); err != nil {
				t.Errorf("close producer fixture runtime: %v", err)
			}
		})
		env := map[string]string{}
		for _, entry := range os.Environ() {
			k, v, ok := strings.Cut(entry, "=")
			if ok {
				env[k] = v
			}
		}
		for _, k := range []string{"MAKEFLAGS", "MAKEOVERRIDES", "MFLAGS", launchContextKey, failureContextKey} {
			delete(env, k)
		}
		maps.Copy(env, rt.Environment())
		env["NODE_BIN"] = node
		env["CARTULARY_TEST_RESULTS_DIR"] = scratch
		env["CARTULARY_TEST_RUN_ID"] = "run"
		p := processRunner{Root: root, Environment: env, FailureScope: func(identity map[string]any) (failureScope, error) { return createFailureScope(rt.Directory, identity) }}
		fixture := filepath.Join(root, "tools/harness/tests/command-failure-fixture.mjs")
		makefile := filepath.Join(scratch, "Makefile")
		if err := os.WriteFile(makefile, []byte("diagnostic contradictory assertion conflict:\n\t@\""+node+"\" \""+fixture+"\" $@\n"), 0600); err != nil {
			t.Fatal(err)
		}
		for _, c := range []struct {
			mode, reason string
			code         int
		}{{"diagnostic", "artifact_error", 11}, {"contradictory", "scheduler_accounting_error", 11}, {"conflict", "scheduler_accounting_error", 11}, {"assertion", "test_assertion_failure", 10}} {
			u := engineUnit("target:build", 1)
			u.TimeoutMS = 10000
			u.Command = command{Executable: "make", Args: []string{"--silent", "-f", makefile, c.mode}, Environment: map[string]string{"CARTULARY_TEST_TARGET": "build"}}
			if c.mode == "assertion" {
				u.Command.Executable = node
				u.Command.Args = []string{fixture, c.mode}
			}
			r := p.Execute(context.Background(), executionRequest{Unit: u})
			if r.Failure == nil || r.Failure.Reason != c.reason || r.ExitCode != c.code {
				t.Fatalf("%s: %+v %s", c.mode, r, r.Stderr)
			}
			entries, _ := os.ReadDir(filepath.Join(rt.Root, "detail"))
			for _, entry := range entries {
				if strings.HasPrefix(entry.Name(), "command-failure-") {
					t.Fatal("private diagnostic survived completion")
				}
			}
		}
	})
	t.Run("cancel_running_cooperative_child", func(t *testing.T) {
		scratch := t.TempDir()
		ready := filepath.Join(scratch, "ready")
		p := processRunner{Root: root, Environment: map[string]string{"NODE_BIN": node, "PATH": "/usr/bin:/bin"}, KillGrace: 100 * time.Millisecond}
		u := engineUnit("test:cancel", 1)
		u.TimeoutMS = 10000
		u.Command = command{Executable: node, Args: []string{"tools/harness/tests/command-failure-fixture.mjs", "cooperative", ready}, Environment: map[string]string{}}
		cancel, done := startRuntimeProcessFixture(t, &p, executionRequest{Unit: u})
		deadline := time.After(5 * time.Second)
		tick := time.NewTicker(5 * time.Millisecond)
		defer tick.Stop()
		for {
			if _, err := os.Stat(ready); err == nil {
				break
			}
			select {
			case r := <-done:
				t.Fatalf("child exited before installing handlers: %+v", r)
			case <-deadline:
				t.Fatal("child did not install handlers")
			case <-tick.C:
			}
		}
		cancel()
		var r unitResult
		select {
		case r = <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("producer fixture did not settle after cancellation")
		}
		if r.Failure == nil || r.Failure.Reason != "cancelled_or_interrupted" || r.Status != "cancelled" {
			t.Fatal(r)
		}
	})
	t.Run("cancel_private_capture_drains_without_publishing_private_output", func(t *testing.T) {
		scratch := t.TempDir()
		run := filepath.Join(scratch, "run")
		if err := os.Mkdir(run, 0700); err != nil {
			t.Fatal(err)
		}
		rt, err := harnesssession.NewRuntime(root, run, "run", filepath.Join(scratch, "private"))
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if _, err := os.Stat(rt.Root); os.IsNotExist(err) {
				return
			}
			// Assertion failures still settle exact journal ownership after the
			// process fixture's later-registered cancellation/join cleanup.
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()
			owner := harnessfixtures.ProductionOptions{Root: root, RunRoot: run, Runtime: rt}
			if err := owner.Recover(ctx); err != nil {
				t.Errorf("recover producer fixture: %v", err)
			}
			if err := rt.Close(); err != nil {
				t.Errorf("close producer fixture runtime: %v", err)
			}
		})
		ready := filepath.Join(scratch, "ready")
		payloadReady := filepath.Join(scratch, "payload-ready")
		env := rt.Environment()
		env["NODE_BIN"], env["PATH"] = node, "/usr/bin:/bin"
		env["CAPTURE_READY"], env["CAPTURE_PAYLOAD_READY"], env["CAPTURE_RUN"] = ready, payloadReady, run
		p := processRunner{Root: root, Environment: env, KillGrace: time.Second}
		u := engineUnit("test:capture-cancel", 1)
		u.TimeoutMS = 10000
		u.Command = command{Executable: node, Args: []string{"--input-type=module", "--eval", `
import {writeFileSync} from 'node:fs';
import {runPrivateCapturedProcess} from './tools/harness/runtime/private-child-process.mjs';
import {createLaunchContext} from './tools/harness/runtime/launch-context.mjs';
process.on('SIGTERM', () => {});
const launch = createLaunchContext({repoRoot: process.cwd(), environment: process.env});
const observedLaunch = {...launch, spawned(pid) {
  launch.spawned(pid);
  // The current onReady callback completes synchronous journal registration
  // before this microtask can publish readiness to the cancelling parent.
  queueMicrotask(() => writeFileSync(process.env.CAPTURE_READY, String(pid)));
}};
await runPrivateCapturedProcess(process.execPath, ['--input-type=module', '--eval',
  "import {writeFileSync} from 'node:fs'; process.stdout.write('private cancellation output'); writeFileSync(process.env.CAPTURE_PAYLOAD_READY, String(process.pid)); setInterval(() => {}, 1000);"],
  {repoRoot: process.cwd(), runRoot: process.env.CAPTURE_RUN, cwd: process.cwd(), env: process.env, dependencies: [], launch: observedLaunch});
`}, Environment: map[string]string{}}
		cancel, done := startRuntimeProcessFixture(t, &p, executionRequest{Unit: u})
		deadline := time.After(5 * time.Second)
		tick := time.NewTicker(5 * time.Millisecond)
		defer tick.Stop()
		var pid int
		for {
			if data, err := os.ReadFile(ready); err == nil {
				pid, _ = strconv.Atoi(string(data))
			}
			payload, err := os.ReadFile(payloadReady)
			payloadPID, _ := strconv.Atoi(string(payload))
			if pid > 0 && err == nil && payloadPID == pid {
				resources, err := rt.Resources(16)
				if err != nil || len(resources) != 1 || resources[0].Kind != "helper_process" {
					t.Fatal("capture was not registered", resources, err)
				}
				var proof harnesssession.ProcessProof
				if json.Unmarshal(resources[0].Target, &proof) != nil || proof.PID != pid {
					t.Fatal("registered identity differs from ready payload")
				}
				break
			}
			select {
			case r := <-done:
				t.Fatalf("capture exited before cancellation: %+v %s", r, r.Stderr)
			case <-deadline:
				t.Fatal("registered captured child did not become ready")
			case <-tick.C:
			}
		}
		cancel()
		var r unitResult
		select {
		case r = <-done:
		case <-time.After(5 * time.Second):
			t.Fatal("producer fixture did not settle after cancellation")
		}
		if r.Failure == nil || r.Failure.Reason != "cancelled_or_interrupted" || len(r.Stdout) != 0 || len(r.Stderr) != 0 {
			t.Fatalf("private output escaped or cancellation lost: %+v", r)
		}
		if err := syscall.Kill(pid, 0); err != syscall.ESRCH {
			t.Fatalf("captured child was not reaped: %v", err)
		}
		captures, err := os.ReadDir(filepath.Join(rt.Root, "detail", "child-captures"))
		if err != nil || len(captures) != 1 {
			t.Fatalf("suite must own the interrupted private capture: %v %v", captures, err)
		}
		if err := rt.Close(); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(rt.Root); !os.IsNotExist(err) {
			t.Fatalf("suite did not remove private residue: %v", err)
		}
	})
}
