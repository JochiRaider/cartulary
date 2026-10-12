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
		defer rt.Close()
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
			u := engineUnit("target:build-web", 1)
			u.TimeoutMS = 10000
			u.Command = command{Executable: "make", Args: []string{"--silent", "-f", makefile, c.mode}, Environment: map[string]string{"CARTULARY_TEST_TARGET": "build-web"}}
			if c.mode == "assertion" {
				u.Command.Executable = node
				u.Command.Args = []string{fixture, c.mode}
			}
			r := p.Execute(context.Background(), executionRequest{Unit: u})
			if r.Failure == nil || r.Failure.Reason != c.reason || r.ExitCode != c.code {
				t.Fatalf("%s: %+v %s", c.mode, r, r.Stderr)
			}
			entries, _ := os.ReadDir(rt.Root)
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
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		done := make(chan unitResult, 1)
		go func() { done <- p.Execute(ctx, executionRequest{Unit: u}) }()
		deadline := time.After(5 * time.Second)
		tick := time.NewTicker(5 * time.Millisecond)
		defer tick.Stop()
		for {
			if _, err := os.Stat(ready); err == nil {
				break
			}
			select {
			case <-deadline:
				t.Fatal("child did not install handlers")
			case <-tick.C:
			}
		}
		cancel()
		r := <-done
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
		defer rt.Close()
		ready := filepath.Join(scratch, "ready")
		env := rt.Environment()
		env["NODE_BIN"], env["PATH"] = node, "/usr/bin:/bin"
		env["CAPTURE_READY"], env["CAPTURE_RUN"] = ready, run
		p := processRunner{Root: root, Environment: env, KillGrace: time.Second}
		u := engineUnit("test:capture-cancel", 1)
		u.TimeoutMS = 10000
		u.Command = command{Executable: node, Args: []string{"--input-type=module", "--eval", `
import {runPrivateCapturedProcess} from './tools/harness/runtime/private-child-process.mjs';
process.on('SIGTERM', () => {});
await runPrivateCapturedProcess(process.execPath, ['--input-type=module', '--eval',
  "import {writeFileSync} from 'node:fs'; process.stdout.write('private cancellation output'); writeFileSync(process.env.CAPTURE_READY, String(process.pid)); setInterval(() => {}, 1000);"],
  {repoRoot: process.cwd(), runRoot: process.env.CAPTURE_RUN, cwd: process.cwd(), env: process.env});
`}, Environment: map[string]string{}}
		ctx, cancel := context.WithCancel(context.Background())
		defer cancel()
		done := make(chan unitResult, 1)
		go func() { done <- p.Execute(ctx, executionRequest{Unit: u}) }()
		deadline := time.After(5 * time.Second)
		tick := time.NewTicker(5 * time.Millisecond)
		defer tick.Stop()
		var pid int
		for pid == 0 {
			if data, err := os.ReadFile(ready); err == nil {
				pid, _ = strconv.Atoi(string(data))
			}
			select {
			case r := <-done:
				t.Fatalf("capture exited before cancellation: %+v %s", r, r.Stderr)
			case <-deadline:
				t.Fatal("captured child did not become ready")
			case <-tick.C:
			}
		}
		cancel()
		r := <-done
		if r.Failure == nil || r.Failure.Reason != "cancelled_or_interrupted" || len(r.Stdout) != 0 || len(r.Stderr) != 0 {
			t.Fatalf("private output escaped or cancellation lost: %+v", r)
		}
		if err := syscall.Kill(pid, 0); err != syscall.ESRCH {
			t.Fatalf("captured child was not reaped: %v", err)
		}
		captures, err := os.ReadDir(filepath.Join(rt.Root, "child-captures"))
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
