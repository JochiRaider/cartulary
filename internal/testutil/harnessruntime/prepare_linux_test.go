//go:build linux

package harnessruntime

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"
)

func TestRuntimeCompilerProcessHelper(t *testing.T) {
	mode := os.Getenv("CARTULARY_RUNTIME_COMPILER_FIXTURE")
	if mode == "" {
		return
	}
	if mode == "descendant-child" {
		time.Sleep(10 * time.Second)
		os.Exit(0)
	}
	f := os.NewFile(3, "plan")
	if f == nil {
		os.Exit(2)
	}
	if mode == "oversized" {
		for i := 0; i < 33; i++ {
			_, _ = f.Write([]byte(strings.Repeat(" ", 1024*1024)))
		}
	} else {
		_, _ = f.Write(planBytes(t, planFixture(t)))
	}
	if mode == "descendant" {
		child := exec.Command(os.Args[0], "-test.run=^TestRuntimeCompilerProcessHelper$")
		child.Env = append(os.Environ(), "CARTULARY_RUNTIME_COMPILER_FIXTURE=descendant-child")
		child.ExtraFiles = []*os.File{f}
		if child.Start() != nil {
			os.Exit(2)
		}
		if err := os.WriteFile(os.Getenv("CARTULARY_COMPILER_CHILD_PROOF"), []byte(strconv.Itoa(child.Process.Pid)), 0600); err != nil {
			os.Exit(2)
		}
	}
	_ = f.Close()
	switch mode {
	case "failure":
		os.Exit(2)
	case "linger":
		time.Sleep(10 * time.Second)
	case "wait":
		time.Sleep(80 * time.Millisecond)
	}
	os.Exit(0)
}
func TestRuntimeCompilerLifetime(t *testing.T) {
	for _, fixture := range []struct {
		stage string
		env   []string
	}{
		{"nested_lifecycle", []string{nestedContextKey + "=invalid private context"}},
		{"diagnostic_mode", []string{"HARNESS_DIAGNOSTICS=invalid private mode"}},
	} {
		t.Run(fixture.stage, func(t *testing.T) {
			root := t.TempDir()
			report, code, err := Run(context.Background(), root, "", nil, fixture.env)
			class, reason := PublicFailure(err)
			entries, readErr := os.ReadDir(root)
			if report != nil || code != 2 || class != "config" || reason != "configuration_error" || PublicFailureStage(err) != fixture.stage || readErr != nil || len(entries) != 0 {
				t.Fatal("startup rejection lost its stage, classification, or no-publication boundary")
			}
		})
	}
	for _, mode := range []string{"wait", "failure", "linger", "oversized", "descendant"} {
		t.Run(mode, func(t *testing.T) {
			timeout := 3 * time.Second
			if mode == "linger" {
				timeout = 100 * time.Millisecond
			}
			ctx, cancel := context.WithTimeout(context.Background(), timeout)
			defer cancel()
			cmd := exec.CommandContext(ctx, os.Args[0], "-test.run=^TestRuntimeCompilerProcessHelper$")
			proofFile := filepath.Join(t.TempDir(), "child")
			cmd.Env = append(os.Environ(), "CARTULARY_RUNTIME_COMPILER_FIXTURE="+mode, "CARTULARY_COMPILER_CHILD_PROOF="+proofFile)
			start := time.Now()
			p, err := receiveCompiler(ctx, cmd)
			if mode == "wait" || mode == "descendant" {
				if err != nil || p == nil || (mode == "wait" && time.Since(start) < 80*time.Millisecond) {
					t.Fatalf("compiler exposed before exit: %v", err)
				}
			} else if err == nil || p != nil {
				t.Fatal("failed compiler exposed plan")
			}
			if mode == "descendant" {
				data, e := os.ReadFile(proofFile)
				if e != nil {
					t.Fatal(e)
				}
				pid, e := strconv.Atoi(string(data))
				if e != nil {
					t.Fatal(e)
				}
				defer syscall.Kill(pid, syscall.SIGKILL)
				state, e := os.ReadFile(fmt.Sprintf("/proc/%d/stat", pid))
				if e == nil {
					fields := strings.Fields(string(state)[strings.LastIndex(string(state), ") ")+2:])
					if fields[0] != "Z" && fields[0] != "X" {
						t.Fatal("compiler admitted live descendant")
					}
				} else if !os.IsNotExist(e) {
					t.Fatal(e)
				}
			}
			if cmd.ProcessState == nil || !cmd.ProcessState.Exited() && ctx.Err() == nil && mode != "oversized" {
				t.Fatalf("compiler not waited: %v", cmd.ProcessState)
			}
		})
	}
}
func TestRuntimePreparationParity(t *testing.T) {
	_, file, _, _ := runtime.Caller(0)
	root := filepath.Clean(filepath.Join(filepath.Dir(file), "../../.."))
	node := os.Getenv("NODE_BIN")
	if node == "" {
		node = filepath.Join(root, ".cache/cartulary/node-runtime/bin/node")
	}
	if _, err := os.Stat(node); err != nil {
		t.Fatalf("Make must supply prepared NODE_BIN: %v", err)
	}
	id := fmt.Sprintf("private-preparation-%d", os.Getpid())
	results := t.TempDir()
	env := append(os.Environ(), "NODE_BIN="+node, "CARTULARY_TEST_RUN_ID="+id, "CARTULARY_TEST_RESULTS_DIR="+results, "CARTULARY_HARNESS_CACHE_MODE=off", "HARNESS_DIAGNOSTICS=basic")
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()
	p, err := Prepare(ctx, root, []string{"--selection", "owner", "--target", "test-slice", "--owner", "harness.evidence_accounting", "--rows", "harness.evidence_accounting.behavior.local_migration_activities"}, env)
	if err != nil {
		t.Fatal(err)
	}
	if p.UnitCount() < 1 || !strings.HasPrefix(p.GraphDigest(), "sha256:") || len(p.facts.SourceInputs) == 0 {
		t.Fatal("missing validated compiler identity")
	}
	selected := false
	for _, u := range p.facts.Graph.Units {
		if strings.Contains(u.Command.Environment["CARTULARY_TEST_ROWS"], "local_migration_activities") {
			selected = true
		}
	}
	if !selected {
		t.Fatal("selected row not preserved")
	}
	entries, err := os.ReadDir(results)
	if err != nil || len(entries) != 0 {
		t.Fatalf("preparation published retained artifacts: %v", err)
	}
	for _, mode := range []string{"normal", "cold"} {
		current, err := Prepare(ctx, root, []string{"--selection", "owner", "--target", "test-slice", "--owner", "harness.evidence_accounting", "--rows", "harness.evidence_accounting.behavior.local_migration_activities"}, append(env, "CARTULARY_HARNESS_CACHE_MODE="+mode))
		if err != nil {
			t.Fatal(mode, err)
		}
		if current.GraphDigest() != p.GraphDigest() || len(current.facts.SourceInputs) == 0 {
			t.Fatal("cache preparation changed graph or lost current source facts")
		}
	}
	if len(p.facts.Preparation["intervals"].([]any)) != 7 {
		t.Fatal("preparation phase observations missing")
	}
}
