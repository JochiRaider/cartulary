//go:build linux

package harnessruntime

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessadmission"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
	"golang.org/x/sys/unix"
)

func TestRuntimeProcessLifecycle(t *testing.T) {
	p := processRunner{Root: t.TempDir(), Environment: map[string]string{"PATH": "/usr/bin:/bin"}, KillGrace: 50 * time.Millisecond}
	t.Run("host binding failure retains stage and successful child exit", func(t *testing.T) {
		u := engineUnit("test:process", 1)
		u.Command = command{Executable: "/bin/sh", Args: []string{"-c", "printf child-output >&2; exit 0"}, Environment: map[string]string{}}
		host := &hostLease{Token: "private-token", Bind: func(pid int) error {
			for {
				var info unix.Siginfo
				err := unix.Waitid(unix.P_PID, pid, &info, unix.WEXITED|unix.WNOWAIT, nil)
				if err == unix.EINTR {
					continue
				}
				if err != nil {
					t.Fatal(err)
				}
				return errors.Join(errors.New("private-path-and-token"), harnessadmission.ErrConflict)
			}
		}}
		r := p.Execute(context.Background(), executionRequest{Unit: u, Host: host})
		if r.Status != "failed" || r.Failure == nil || r.Failure.Class != "infra" || r.Failure.Reason != "resource_conflict" || r.ExitCode != 0 || len(r.Cleanup) != 0 {
			t.Fatalf("failure or settlement changed: %+v", r)
		}
		if string(r.Stderr) != "child-output\n[RUNTIME-EXECUTION] stage=host_bind cause=admission_conflict\n" {
			t.Fatalf("missing or unsafe diagnostic: %q", r.Stderr)
		}
	})
	t.Run("execution diagnostics retain closed values only", func(t *testing.T) {
		for _, c := range []struct {
			err   error
			cause string
		}{
			{context.Canceled, "cancelled"}, {context.DeadlineExceeded, "deadline_exceeded"},
			{os.ErrPermission, "permission_denied"}, {os.ErrNotExist, "identity_unavailable"},
			{unix.EBADF, "descriptor_closed"}, {errors.New("private-path-and-token"), "unknown"},
		} {
			want := "[RUNTIME-EXECUTION] stage=workspace_inherit cause=" + c.cause + "\n"
			if got := string(executionDiagnostic("workspace_inherit", c.err)); got != want {
				t.Fatalf("diagnostic: %q", got)
			}
		}
		if got := string(executionDiagnostic("private-stage", errors.New("private-error"))); got != "[RUNTIME-EXECUTION] stage=unknown cause=unknown\n" {
			t.Fatalf("unsafe diagnostic: %q", got)
		}
	})
	for _, c := range []struct {
		name, script, status, reason string
		timeout                      int64
	}{{"pass", "printf done", "passed", "", 1000}, {"assertion", "exit 10", "failed", "test_assertion_failure", 1000}, {"cleanup", "exit 12", "failed", "cleanup_error", 1000}, {"marker", "printf 'failure_class=infra failure_reason=resource_conflict'; exit 2", "failed", "resource_conflict", 1000}, {"deadline", "trap '' TERM; sleep 30", "failed", "timeout_failure", 20}} {
		t.Run(c.name, func(t *testing.T) {
			u := engineUnit("test:process", 1)
			u.Command = command{Executable: "/bin/sh", Args: []string{"-c", c.script}, Environment: map[string]string{}}
			u.TimeoutMS = c.timeout
			start := time.Now()
			r := p.Execute(context.Background(), executionRequest{Unit: u})
			if r.Status != c.status || c.reason != "" && (r.Failure == nil || r.Failure.Reason != c.reason) {
				t.Fatalf("%+v", r)
			}
			if time.Since(start) > time.Second {
				t.Fatal("process group did not drain within fixture bound")
			}
		})
	}
	t.Run("spawn_failure", func(t *testing.T) {
		u := engineUnit("test:process", 1)
		u.Command.Executable = "/does/not/exist"
		r := p.Execute(context.Background(), executionRequest{Unit: u})
		if r.Failure.Reason != "service_start_error" {
			t.Fatal(r)
		}
	})
	t.Run("lingering_group", func(t *testing.T) {
		u := engineUnit("test:process", 1)
		u.Command = command{Executable: "/bin/sh", Args: []string{"-c", "sleep 30 & printf '%s' $!"}, Environment: map[string]string{}}
		r := p.Execute(context.Background(), executionRequest{Unit: u})
		if r.Status != "passed" {
			t.Fatal(r)
		}
		pid, err := strconv.Atoi(string(r.Stdout))
		if err != nil {
			t.Fatal(err)
		}
		if proof, err := harnessadmission.Identity(pid); err == nil {
			if alive, err := harnessadmission.Alive(proof); err != nil || alive {
				t.Fatalf("lingering child %d alive=%v error=%v", pid, alive, err)
			}
		}
	})
	t.Run("cancel_before_spawn", func(t *testing.T) {
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		u := engineUnit("test:process", 1)
		r := p.Execute(ctx, executionRequest{Unit: u})
		if r.Status != "cancelled" {
			t.Fatal(r)
		}
	})
	t.Run("bounded_tail", func(t *testing.T) {
		tail := boundedTail{limit: 8}
		tail.Write([]byte("012345"))
		tail.Write([]byte("6789"))
		if string(tail.bytes) != "23456789" {
			t.Fatal(string(tail.bytes))
		}
		tail.Write([]byte("abcdefghijkl"))
		if string(tail.bytes) != "efghijkl" {
			t.Fatal(string(tail.bytes))
		}
	})
}
func TestRuntimePrivateFailureScope(t *testing.T) {
	root := t.TempDir()
	if err := os.Chmod(root, 0700); err != nil {
		t.Fatal(err)
	}
	directory, err := harnesslocal.OpenDirectory(root, true)
	if err != nil {
		t.Fatal(err)
	}
	defer directory.Close()
	u := engineUnit("test:scope", 1)
	u.Command.Environment["CARTULARY_TEST_TARGET"] = "test-slice"
	commands := map[string]string{"test-slice": "cartulary.harness.command.test_slice.v1"}
	environment := map[string]string{"CARTULARY_TEST_RUN_ID": "fixture"}
	for _, kind := range []string{"valid", "foreign", "duplicate", "incomplete", "conflict", "inconsistent", "symlink"} {
		t.Run(kind, func(t *testing.T) {
			identity, err := freshLaunch(u, environment, commands)
			if err != nil {
				t.Fatal(err)
			}
			scope, err := createFailureScope(directory, identity)
			if err != nil {
				t.Fatal(err)
			}
			defer func() {
				if err := scope.Close(); err != nil {
					t.Error(err)
				}
			}()
			s := scope.(*privateFailureScope)
			if s.Read() != nil {
				t.Fatal("empty channel failed")
			}
			record := map[string]any{}
			for k, v := range s.expected {
				record[k] = v
			}
			record["failure_class"] = "infra"
			record["failure_reason"] = "resource_conflict"
			if kind == "foreign" {
				record["run_id"] = "other"
			}
			if kind == "inconsistent" {
				record["failure_class"] = "product"
			}
			data, _ := json.Marshal(record)
			if kind == "duplicate" {
				data = []byte(strings.Replace(string(data), `"failure_class":"infra"`, `"failure_class":"infra","failure_class":"infra"`, 1))
			}
			file := filepath.Join(root, "detail", s.name, "failure.json")
			switch kind {
			case "incomplete":
				err = os.Mkdir(filepath.Join(root, "detail", s.name, "writer"), 0700)
			case "symlink":
				err = os.Symlink("/dev/null", file)
			default:
				err = harnesslocal.PublishNew(file, data)
			}
			if err != nil {
				t.Fatal(err)
			}
			if kind == "conflict" {
				if err := os.WriteFile(filepath.Join(root, "detail", s.name, "conflict"), nil, 0600); err != nil {
					t.Fatal(err)
				}
			}
			f := s.Read()
			if f == nil {
				t.Fatal("missing failure")
			}
			expected := "scheduler_accounting_error"
			if kind == "valid" {
				expected = "resource_conflict"
			}
			if f.Reason != expected {
				t.Fatal(f)
			}
		})
	}
	first, _ := freshLaunch(u, environment, commands)
	bytes, _ := json.Marshal(first)
	environment[launchContextKey] = string(bytes)
	second, _ := freshLaunch(u, environment, commands)
	if first["invocation_id"] == second["invocation_id"] || second["parent_invocation_id"] != first["invocation_id"] {
		t.Fatal("launch identity reused")
	}
	environment["CARTULARY_TEST_RUN_ID"] = "other"
	third, _ := freshLaunch(u, environment, commands)
	if third["parent_invocation_id"] != nil {
		t.Fatal("foreign launch inherited")
	}
}
func TestRuntimeHostAdmission(t *testing.T) {
	root := filepath.Join(t.TempDir(), "host")
	capacities := map[string]int64{"cpu": 1, "process": 2, "io": 1, "memory_mb": 100, "browser_stack": 1}
	admit := nativeAdmission(capacities, nil, root)
	u := engineUnit("a", 1)
	lease, err := admit(context.Background(), u)
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Millisecond)
	defer cancel()
	if _, err := admit(ctx, u); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatal("host wait lost cancellation or overbooked", err)
	}
	short := u
	short.TimeoutMS = 30
	if _, err := admit(context.Background(), short); err == nil {
		t.Fatal("unit admission deadline ignored")
	} else if class, reason := PublicFailure(err); class != "infra" || reason != "resource_conflict" {
		t.Fatalf("admission timeout classification: %s/%s: %v", class, reason, err)
	}
	if err := lease.Release(context.Background()); err != nil {
		t.Fatal(err)
	}
	if err := lease.Release(context.Background()); err != nil {
		t.Fatal(err)
	}
	lease, err = admit(context.Background(), u)
	if err != nil {
		t.Fatal("cancelled request retained capacity", err)
	}
	// A nested compiler receives the admitted unit budget even if ambient,
	// fixture and command inputs all try to supply the whole-host budget.
	p := processRunner{Root: t.TempDir(), Environment: map[string]string{"PATH": "/usr/bin:/bin", "CARTULARY_UNIT_CPU_TOKENS": "19"}}
	u.Command = command{Executable: "/bin/sh", Args: []string{"-c", `printf '%s:%s' "$CARTULARY_HOST_ADMISSION_LEASE" "$CARTULARY_UNIT_CPU_TOKENS"`}, Environment: map[string]string{"CARTULARY_HOST_ADMISSION_LEASE": "command-override", "CARTULARY_UNIT_CPU_TOKENS": "37"}}
	r := p.Execute(context.Background(), executionRequest{Unit: u, Host: lease, Environment: map[string]string{"CARTULARY_HOST_ADMISSION_LEASE": "fixture-override", "CARTULARY_UNIT_CPU_TOKENS": "23"}})
	if r.Status != "passed" || string(r.Stdout) != lease.Token+":1" {
		t.Fatal("admitted unit lease and CPU budget were not authoritative")
	}
	if err := lease.Release(context.Background()); err != nil {
		t.Fatal(err)
	}
}
