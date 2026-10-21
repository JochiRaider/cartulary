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
	"github.com/JochiRaider/cartulary/internal/testutil/harnesssession"
	"golang.org/x/sys/unix"
)

func TestRuntimeProcessLifecycle(t *testing.T) {
	if mode := os.Getenv("CARTULARY_NESTED_TEST_HELPER"); mode != "" {
		ctx, finish, err := joinNested(context.Background(), os.Getenv(nestedContextKey))
		if err != nil {
			os.Exit(21)
		}
		defer finish(mode != "unresolved")
		if err := os.WriteFile("ready", []byte("ready"), 0600); err != nil {
			os.Exit(22)
		}
		<-ctx.Done()
		if mode == "stalled" {
			time.Sleep(time.Minute)
			return
		}
		if mode == "abrupt" {
			os.Exit(23)
		}
		// A finalizer longer than ordinary grace must survive parent cancellation.
		time.Sleep(200 * time.Millisecond)
		if err := os.WriteFile("settled", []byte("settled"), 0600); err != nil {
			os.Exit(24)
		}
		return
	}
	t.Run("nested lifecycle settles before wrapper termination", func(t *testing.T) {
		for _, mode := range []string{"normal", "watchdog", "unresolved", "abrupt", "stalled"} {
			t.Run(mode, func(t *testing.T) {
				binary, err := os.Executable()
				if err != nil {
					t.Fatal(err)
				}
				root := t.TempDir()
				p := processRunner{Root: root, Environment: map[string]string{"PATH": "/usr/bin:/bin", "CARTULARY_NESTED_TEST_HELPER": mode}, KillGrace: 20 * time.Millisecond, NestedSettlement: time.Second}
				u := engineUnit("test:nested", 1)
				u.TimeoutMS = 10000
				if mode == "watchdog" {
					u.TimeoutMS = 1000
				}
				u.Command = command{Executable: binary, Args: []string{"-test.run=^TestRuntimeProcessLifecycle$"}, Environment: map[string]string{}}
				ctx, cancel := context.WithCancel(context.Background())
				defer cancel()
				result := make(chan unitResult, 1)
				go func() { result <- p.Execute(ctx, executionRequest{Unit: u}) }()
				deadline := time.After(5 * time.Second)
				for {
					if _, err := os.Stat(filepath.Join(root, "ready")); err == nil {
						break
					}
					select {
					case r := <-result:
						t.Fatalf("child failed before registration: %+v stderr=%s", r, r.Stderr)
					case <-deadline:
						t.Fatal("child registration stalled")
					case <-time.After(10 * time.Millisecond):
					}
				}
				if mode != "watchdog" {
					cancel()
				}
				select {
				case r := <-result:
					wantStatus, wantReason := "cancelled", "cancelled_or_interrupted"
					if mode == "watchdog" {
						wantStatus, wantReason = "failed", "timeout_failure"
					}
					if r.Status != wantStatus || r.Failure.Reason != wantReason {
						t.Fatalf("primary lost: %+v", r)
					}
					if mode == "normal" || mode == "watchdog" {
						if _, err := os.Stat(filepath.Join(root, "settled")); err != nil || len(r.Cleanup) != 0 {
							t.Fatalf("finalizer interrupted: %+v, %v", r, err)
						}
					} else if len(r.Cleanup) == 0 {
						t.Fatal("unacknowledged child settlement passed")
					}
				case <-time.After(5 * time.Second):
					t.Fatal("nested cancellation exceeded bound")
				}
			})
		}
	})
	t.Run("nested registration ownership fencing and multiple graphs", func(t *testing.T) {
		s, err := newNestedScope()
		if err != nil {
			t.Fatal(err)
		}
		defer s.close()
		proof, err := harnesssession.CaptureProcess(os.Getpid())
		if err != nil {
			t.Fatal(err)
		}
		s.bind(proof)
		first, finishFirst, err := joinNested(context.Background(), s.environment())
		if err != nil {
			t.Fatal(err)
		}
		second, finishSecond, err := joinNested(context.Background(), s.environment())
		if err != nil {
			finishFirst(true)
			t.Fatal(err)
		}
		drained := s.cancel()
		for _, child := range []context.Context{first, second} {
			select {
			case <-child.Done():
			case <-time.After(time.Second):
				t.Fatal("cancellation not delivered")
			}
		}
		if _, finish, err := joinNested(context.Background(), s.environment()); err == nil {
			finish(true)
			t.Fatal("late registration admitted")
		}
		finishFirst(true)
		select {
		case <-drained:
			t.Fatal("settled before all children returned")
		default:
		}
		finishSecond(true)
		select {
		case <-drained:
		case <-time.After(time.Second):
			t.Fatal("settlement not acknowledged")
		}
		if err := s.close(); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("nested foreign identity and malformed authority fail closed", func(t *testing.T) {
		for _, mode := range []string{"foreign_launch", "foreign_owner", "duplicate", "unknown"} {
			t.Run(mode, func(t *testing.T) {
				s, err := newNestedScope()
				if err != nil {
					t.Fatal(err)
				}
				defer s.close()
				proof, _ := harnesssession.CaptureProcess(os.Getpid())
				if mode == "foreign_launch" {
					proof.Start = "0"
				}
				s.bind(proof)
				binding := s.context
				if mode == "foreign_owner" {
					binding.Owner.Start = "0"
				}
				data, _ := json.Marshal(binding)
				value := string(data)
				if mode == "duplicate" {
					value = strings.Replace(value, `"version":1`, `"version":1,"version":1`, 1)
				}
				if mode == "unknown" {
					value = strings.Replace(value, `"version":1`, `"version":1,"unknown":1`, 1)
				}
				if _, finish, err := joinNested(context.Background(), value); err == nil {
					finish(true)
					t.Fatal("invalid registration admitted")
				}
			})
		}
	})
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
	admit := nativeAdmission(capacities, nil, root, nil)
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
	queued, done := context.WithTimeout(t.Context(), 2*time.Second)
	defer done()
	type outcome struct {
		lease *hostLease
		err   error
	}
	result := make(chan outcome, 1)
	go func() {
		lease, err := admit(queued, short)
		result <- outcome{lease, err}
	}()
	select {
	case outcome := <-result:
		if outcome.lease != nil {
			_ = outcome.lease.Release(context.Background())
		}
		t.Fatalf("capacity waiting consumed the execution watchdog: %v", outcome.err)
	case <-time.After(3 * time.Duration(short.TimeoutMS) * time.Millisecond):
	}
	if err := lease.Release(context.Background()); err != nil {
		t.Fatal(err)
	}
	acquired := <-result
	if acquired.err != nil || acquired.lease == nil {
		t.Fatal("healthy queued work did not acquire released capacity", acquired.err)
	}
	if err := acquired.lease.Release(context.Background()); err != nil {
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
