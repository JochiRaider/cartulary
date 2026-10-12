//go:build linux

package harnesssession

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"sync"
	"syscall"
	"testing"
	"time"
)

func testDetailLifecycle(t *testing.T) {
	for _, transition := range []string{"finalize", "preserve", "close"} {
		t.Run("ordinary_detail_fenced_after_"+transition, func(t *testing.T) {
			r, repo, run := runtimeFixture(t)
			borrowed, err := BorrowRuntime(repo, run, r.Root, r.RunID, r.LeaseID)
			if err != nil {
				t.Fatal(err)
			}
			defer borrowed.ReleaseHandles()
			if _, err := borrowed.PrivatePath("before"); err != nil {
				t.Fatal(err)
			}
			switch transition {
			case "finalize":
				if _, err := r.BeginFinalization(); err != nil {
					t.Fatal(err)
				}
			case "preserve":
				proof := filepath.Join(r.Root, "unresolved-suite.json")
				if err := os.WriteFile(proof, []byte("private proof"), 0600); err != nil {
					t.Fatal(err)
				}
				if err := r.RecordDependentResource("managed_suite", proof, "acquired", []string{}); err != nil {
					t.Fatal(err)
				}
				defer r.RecordResource("managed_suite", proof, "released")
				if err := r.PreserveRecovery(); err != nil {
					t.Fatal(err)
				}
			case "close":
				if err := r.Close(); err != nil {
					t.Fatal(err)
				}
			}
			for _, handle := range []*Runtime{r, borrowed} {
				if _, err := handle.PrivatePath("late"); err == nil {
					t.Error("ordinary allocation survived lifecycle boundary")
				}
			}
		})
	}
}

// Executed by the race-enabled native route in an independent process. The
// writer deliberately retains both a pathname and an open descriptor.
func detailProcessFixture() bool {
	mode := os.Getenv("CARTULARY_DETAIL_TEST_MODE")
	if mode == "" {
		return false
	}
	r, err := BorrowRuntime(os.Getenv("DETAIL_REPO"), os.Getenv("DETAIL_RUN"), os.Getenv("DETAIL_ROOT"), "fixture", os.Getenv("DETAIL_LEASE"))
	if err != nil {
		os.Exit(2)
	}
	defer r.ReleaseHandles()
	if mode == "borrow" {
		fmt.Fprintln(os.Stdout, "ready")
		if _, err := bufio.NewReader(os.Stdin).ReadString('\n'); err != nil {
			os.Exit(3)
		}
		if _, err := r.PrivatePath("late"); err == nil {
			os.Exit(4)
		}
		if _, err := r.BeginRecoveryAttempt(); err == nil {
			os.Exit(5)
		}
		return true
	}
	file, err := r.PrivatePath("writer", "private.log")
	if err != nil {
		os.Exit(6)
	}
	if err := os.MkdirAll(filepath.Dir(file), 0700); err != nil {
		os.Exit(7)
	}
	fd, err := os.OpenFile(file, os.O_CREATE|os.O_WRONLY, 0600)
	if err != nil {
		os.Exit(8)
	}
	defer fd.Close()
	fmt.Fprintln(os.Stdout, "ready")
	for {
		_, _ = fd.WriteString("private detail\n")
		_ = os.MkdirAll(filepath.Dir(file), 0700)
		_ = os.WriteFile(file, []byte("retained path"), 0600)
		time.Sleep(time.Millisecond)
	}
}

func detailChild(t *testing.T, r *Runtime, repo, run, mode string) (*exec.Cmd, *bufio.Reader, io.WriteCloser) {
	t.Helper()
	cmd := exec.Command(os.Args[0], "-test.run=^TestSessionRuntimeProof$")
	cmd.Env = append(os.Environ(), "CARTULARY_DETAIL_TEST_MODE="+mode, "DETAIL_REPO="+repo, "DETAIL_RUN="+run, "DETAIL_ROOT="+r.Root, "DETAIL_LEASE="+r.LeaseID)
	if mode == "write" {
		for key, value := range r.Environment() {
			cmd.Env = append(cmd.Env, key+"="+value)
		}
	}
	cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	output, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	input, err := cmd.StdinPipe()
	if err != nil {
		t.Fatal(err)
	}
	if err := cmd.Start(); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = cmd.Process.Kill(); _ = cmd.Wait() })
	reader := bufio.NewReader(output)
	if line, err := reader.ReadString('\n'); err != nil || line != "ready\n" {
		t.Fatal("child admission", line, err)
	}
	return cmd, reader, input
}

func testDetailAdversaries(t *testing.T) {
	t.Run("separate_handles_allocation_serializes_with_fence", func(t *testing.T) {
		r, repo, run := runtimeFixture(t)
		borrowed, err := BorrowRuntime(repo, run, r.Root, r.RunID, r.LeaseID)
		if err != nil {
			t.Fatal(err)
		}
		defer borrowed.ReleaseHandles()
		start := make(chan struct{})
		var workers sync.WaitGroup
		workers.Go(func() {
			<-start
			for range 100 {
				if _, err := borrowed.PrivatePath("racing"); err != nil {
					return
				}
			}
		})
		close(start)
		if _, err := r.BeginFinalization(); err != nil {
			t.Fatal(err)
		}
		workers.Wait()
		for range 10 {
			if _, err := borrowed.PrivatePath("late"); err == nil {
				t.Fatal("fence lost across handles")
			}
		}
		if _, err := os.Stat(filepath.Join(r.Root, "detail")); !os.IsNotExist(err) {
			t.Fatal("detail recreated", err)
		}
	})

	for _, transition := range []string{"finalize", "preserve", "close"} {
		t.Run("separate_borrower_rejects_"+transition, func(t *testing.T) {
			r, repo, run := runtimeFixture(t)
			cmd, _, input := detailChild(t, r, repo, run, "borrow")
			switch transition {
			case "finalize":
				_, err := r.BeginFinalization()
				if err != nil {
					t.Fatal(err)
				}
			case "preserve":
				if err := r.PreserveRecovery(); err != nil {
					t.Fatal(err)
				}
			case "close":
				if err := r.Close(); err != nil {
					t.Fatal(err)
				}
			}
			if _, err := io.WriteString(input, "allocate\n"); err != nil {
				t.Fatal(err)
			}
			input.Close()
			if err := cmd.Wait(); err != nil {
				t.Fatal("stale borrower admitted", err)
			}
		})
	}
	t.Run("retained_path_and_fd_producer_settles_before_purge", func(t *testing.T) {
		r, repo, run := runtimeFixture(t)
		cmd, _, _ := detailChild(t, r, repo, run, "write")
		if _, err := r.BeginFinalization(); err != nil {
			t.Fatal(err)
		}
		_ = cmd.Wait()
		if state, err := readProcess(cmd.Process.Pid); err == nil && state.state != "Z" && state.state != "X" {
			t.Fatal("writer survives purge")
		}
		if _, err := os.Stat(filepath.Join(r.Root, "detail")); !os.IsNotExist(err) {
			t.Fatal("detail recreated", err)
		}
	})
	t.Run("unresolved_producer_prevents_successful_purge", func(t *testing.T) {
		r, _, _ := runtimeFixture(t)
		file, err := r.PrivatePath("sensitive.log")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(file, []byte("private"), 0600); err != nil {
			t.Fatal(err)
		}
		authority, err := r.readDetail("ordinary")
		if err != nil {
			t.Fatal(err)
		}
		original := authority
		proof, err := CaptureProcess(os.Getpid())
		if err != nil {
			t.Fatal(err)
		}
		proof.Start = "0"
		authority.Producers = []ProcessProof{proof}
		if err := r.writeDetail(authority); err != nil {
			t.Fatal(err)
		}
		if _, err := r.BeginFinalization(); err == nil {
			t.Fatal("conflicting live producer accepted")
		}
		if _, err := os.Stat(file); err != nil {
			t.Fatal("unresolved detail removed", err)
		}
		if _, err := r.PrivatePath("late"); err == nil {
			t.Fatal("cleanup failure reopened allocation")
		}
		original.State = "fenced"
		if err := r.writeDetail(original); err != nil {
			t.Fatal(err)
		}
		if _, err := r.BeginFinalization(); err != nil {
			t.Fatal(err)
		}
	})
	t.Run("recovery_capabilities_are_fresh_bounded_and_do_not_reopen_ordinary_detail", func(t *testing.T) {
		r, repo, run := runtimeFixture(t)
		borrowed, err := BorrowRuntime(repo, run, r.Root, r.RunID, r.LeaseID)
		if err != nil {
			t.Fatal(err)
		}
		defer borrowed.ReleaseHandles()
		if _, err := borrowed.BeginRecoveryAttempt(); err == nil {
			t.Fatal("borrow minted authority")
		}
		first, err := r.BeginRecoveryAttempt()
		if err != nil {
			t.Fatal(err)
		}
		file, err := r.RecoveryDetailPath("attempt.log")
		if err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(file, []byte("private"), 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := r.PrivatePath("ordinary"); err == nil {
			t.Fatal("recovery reopened ordinary detail")
		}
		if err := r.FinishRecoveryAttempt(); err != nil {
			t.Fatal(err)
		}
		if _, err := os.Stat(file); !os.IsNotExist(err) {
			t.Fatal("attempt detail retained", err)
		}
		if err := borrowed.bindRecoveryAttempt(first); err == nil {
			t.Fatal("stale attempt admitted")
		}
		second, err := r.BeginRecoveryAttempt()
		if err != nil || first == second {
			t.Fatal("attempt reused", err)
		}
		if err := r.FinishRecoveryAttempt(); err != nil {
			t.Fatal(err)
		}
	})
}
