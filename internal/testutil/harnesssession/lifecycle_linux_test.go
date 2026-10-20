//go:build linux

package harnesssession

import (
	"bufio"
	"context"
	"encoding/json"
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

	"github.com/JochiRaider/cartulary/internal/testutil/harnessadmission"
	"golang.org/x/sys/unix"
)

// The producer acknowledges child readiness, never its PID. Setsid happens in
// the fork/exec boundary before the child can publish readiness, so recovery
// observes a fully established session even when the producer has already died.
func diagnosticOrphanFixture() {
	switch os.Getenv("CARTULARY_SCOPE_TEST_ORPHAN") {
	case "child":
		fmt.Fprintln(os.Stdout, "ready")
		os.Stdout.Close()
		time.Sleep(time.Minute)
		os.Exit(0)
	case "producer":
		child := exec.Command(os.Args[0], "-test.run=^TestSessionRuntimeProof$")
		child.Env = append(os.Environ(), "CARTULARY_SCOPE_TEST_ORPHAN=child")
		child.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
		output, err := child.StdoutPipe()
		if err != nil || child.Start() != nil {
			os.Exit(2)
		}
		line, err := bufio.NewReader(output).ReadString('\n')
		output.Close()
		if err != nil || line != "ready\n" {
			_ = child.Process.Kill()
			_ = child.Wait()
			os.Exit(3)
		}
		if child.Process.Release() != nil {
			_ = child.Process.Kill()
			_ = child.Wait()
			os.Exit(4)
		}
		os.Exit(0)
	}
}

func testCorruptJournalDetail(t *testing.T) {
	t.Run("persisted_corruption_cannot_retain_disposable_detail", func(t *testing.T) {
		r, _, _ := runtimeFixture(t)
		proof := filepath.Join(r.Root, "proof.json")
		detail, err := r.PrivatePath("artifacts", "sensitive.log")
		if err != nil {
			t.Fatal(err)
		}
		if err = os.MkdirAll(filepath.Dir(detail), 0700); err != nil {
			t.Fatal(err)
		}
		for _, file := range []string{proof, detail} {
			if err = os.WriteFile(file, []byte("private fixture"), 0600); err != nil {
				t.Fatal(err)
			}
		}
		if err = r.RecordDependentResource("managed_suite", proof, "acquired", []string{}); err != nil {
			t.Fatal(err)
		}
		raw, _ := json.Marshal(proof)
		name, _ := resourceName("managed_suite", raw)
		journal := filepath.Join(r.Root, "recovery", name)
		original, err := os.ReadFile(journal)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			if err := os.WriteFile(journal, original, 0600); err != nil {
				t.Error(err)
			}
			if err := r.RecordResource("managed_suite", proof, "released"); err != nil {
				t.Error(err)
			}
		})
		var forged map[string]any
		if err = json.Unmarshal(original, &forged); err != nil {
			t.Fatal(err)
		}
		forged["recovery_files"] = []string{detail}
		forgedBytes, _ := json.Marshal(forged)
		for _, corrupt := range [][]byte{[]byte(`{"kind":`), forgedBytes} {
			if err = os.MkdirAll(filepath.Dir(detail), 0700); err != nil {
				t.Fatal(err)
			}
			if err = os.WriteFile(detail, []byte("new private detail"), 0600); err != nil {
				t.Fatal(err)
			}
			if err = os.WriteFile(journal, corrupt, 0600); err != nil {
				t.Fatal(err)
			}
			if r.PreserveRecovery() == nil {
				t.Error("corrupt authority accepted")
			}
			if _, err = os.Stat(detail); !os.IsNotExist(err) {
				t.Error("safe disposable detail survived corrupt journal", err)
			}
			if _, err = os.Stat(proof); err != nil {
				t.Error("unresolved authority removed", err)
			}
		}
	})
}

func testLifecycleGaps(t *testing.T) {
	t.Run("diagnostic_scope_recovers_orphan_before_pid_publication", func(t *testing.T) {
		root, err := filepath.Abs("../../..")
		if err != nil {
			t.Fatal(err)
		}
		run := t.TempDir()
		r, err := NewRuntime(root, run, "orphan-scope", t.TempDir())
		if err != nil {
			t.Fatal(err)
		}
		defer r.Close()
		identity, err := CaptureProcess(os.Getpid())
		if err != nil {
			t.Fatal(err)
		}
		token := strings.ReplaceAll(identifier()+identifier(), "-", "")
		scope := map[string]any{"boot": identity.Boot, "start": identity.Start, "uid": os.Getuid(), "token": token}
		if err := r.RecordDependentResource("diagnostic_scope", scope, "acquired", []string{}); err != nil {
			t.Fatal(err)
		}
		defer r.RecordResource("diagnostic_scope", scope, "released")
		// Wait for session establishment, then lose the producer before any PID
		// publication. A nonce alone cannot establish fixture readiness while
		// the old asynchronous setsid command is still changing PGID/SID.
		launch, stopLaunch := context.WithTimeout(context.Background(), 5*time.Second)
		defer stopLaunch()
		producer := exec.CommandContext(launch, os.Args[0], "-test.run=^TestSessionRuntimeProof$")
		producer.Env = append(os.Environ(), "CARTULARY_DIAGNOSTIC_TOKEN="+token, "CARTULARY_SCOPE_TEST_ORPHAN=producer")
		if err := producer.Run(); err != nil {
			t.Fatal(err)
		}
		// Independent fixture census establishes the real orphan, not the
		// implementation's authorization or discovery result.
		var orphan ProcessProof
		deadline := time.Now().Add(5 * time.Second)
		for orphan.PID == 0 && time.Now().Before(deadline) {
			entries, err := os.ReadDir("/proc")
			if err != nil {
				t.Fatal(err)
			}
			for _, entry := range entries {
				pid, err := strconv.Atoi(entry.Name())
				if err != nil {
					continue
				}
				data, err := os.ReadFile(filepath.Join("/proc", entry.Name(), "environ"))
				if err == nil && strings.Contains("\x00"+string(data), "\x00CARTULARY_DIAGNOSTIC_TOKEN="+token+"\x00") {
					orphan, _ = CaptureProcess(pid)
					break
				}
			}
			if orphan.PID == 0 {
				time.Sleep(10 * time.Millisecond)
			}
		}
		if orphan.PID == 0 {
			t.Fatal("producer did not create marked orphan")
		}
		defer syscall.Kill(orphan.PID, syscall.SIGKILL)
		if orphan.PGID != orphan.PID || orphan.SID != orphan.PID {
			t.Fatal("orphan readiness preceded session establishment")
		}
		resources, err := r.Resources(8)
		if err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()
		if err := r.StopResource(ctx, root, run, map[string]string{"NODE_BIN": os.Getenv("NODE_BIN")}, resources[0]); err != nil {
			t.Fatal("journal-authorized orphan recovery failed", err)
		}
		state, err := readProcess(orphan.PID)
		if err == nil && state.state != "Z" && state.state != "X" {
			t.Fatal("orphan survived scope cleanup")
		}
		if err != nil && !processGone(err) {
			t.Fatal(err)
		}
	})
	t.Run("physical_worker_closed_pipes_do_not_prove_descendant_settlement", func(t *testing.T) {
		file := filepath.Join(t.TempDir(), "child")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, "/bin/sh", "-c", `sleep 60 </dev/null >/dev/null 2>&1 & echo $! > "$1"; exit 0`, "fixture", file)
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		if err := waitPhysicalCommand(ctx, cmd, nil); err != nil {
			t.Fatal(err)
		}
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
		if err != nil {
			t.Fatal(err)
		}
		defer syscall.Kill(pid, syscall.SIGKILL)
		state, err := readProcess(pid)
		if err == nil && state.state != "Z" && state.state != "X" {
			t.Fatal("cleanup acknowledged live descendant")
		}
		if err != nil && !processGone(err) {
			t.Fatal(err)
		}
	})
	t.Run("settlement_deadline_keeps_live_identity_unresolved", func(t *testing.T) {
		p := processState{ProcessProof: ProcessProof{PID: 101, Start: "11", PGID: 101, SID: 50, Group: true}, state: "S"}
		ctx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
		defer cancel()
		err := settleProcess(ctx, p.ProcessProof, 0, func(ProcessProof) ([]processState, error) { return []processState{p}, nil }, func(processState, unix.Signal) error { return nil })
		if !errors.Is(err, context.DeadlineExceeded) {
			t.Fatal("signal delivery substituted for absence", err)
		}
	})
	t.Run("signal_observes_reaped_leader_while_group_survives", func(t *testing.T) {
		leader := exec.Command("/bin/sh", "-c", "read -r release")
		leader.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		release, err := leader.StdinPipe()
		if err != nil {
			t.Fatal(err)
		}
		defer release.Close()
		if err := leader.Start(); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = leader.Process.Kill(); _ = leader.Wait() }()
		observed, err := readProcess(leader.Process.Pid)
		if err != nil {
			t.Fatal(err)
		}
		member := exec.Command("sleep", "60")
		member.SysProcAttr = &syscall.SysProcAttr{Setpgid: true, Pgid: leader.Process.Pid}
		if err := member.Start(); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = member.Process.Kill(); _ = member.Wait() }()
		before, err := readProcess(member.Process.Pid)
		if err != nil || before.PGID != observed.PGID || before.SID != observed.SID {
			t.Fatal("fixture did not establish the shared group", err)
		}
		forged := observed
		forged.Start = "0"
		if err := signalProcess(forged, unix.SIGTERM); !errors.Is(err, ErrUnresolved) || DescribeCleanupFailure(err).Stage != "process_identity" {
			t.Fatal("acquired handle bypassed process identity validation", err)
		}
		if _, err := fmt.Fprintln(release, "release"); err != nil {
			t.Fatal(err)
		}
		if err := leader.Wait(); err != nil {
			t.Fatal(err)
		}
		if _, err := readProcess(observed.PID); !processGone(err) {
			t.Fatal("fixture leader was not reaped", err)
		}
		// The process group keeps the numeric PID allocated on Linux even though
		// its task has been reaped. Signaling this stale observation must defer
		// settlement to another complete discovery sweep, not fail or signal the group.
		if err := signalProcess(observed, unix.SIGTERM); !processGone(err) {
			t.Fatal("reaped member did not become an absence observation", err)
		}
		after, err := readProcess(member.Process.Pid)
		if err != nil || !sameProcess(before.ProcessProof, after.ProcessProof) || after.state == "Z" || after.state == "X" {
			t.Fatal("stale leader signaling affected its surviving group member", err)
		}
	})
	t.Run("settlement_requires_observed_absence", func(t *testing.T) {
		leader := processState{ProcessProof: ProcessProof{PID: 100, Start: "10", PGID: 100, SID: 50, Group: true}, state: "Z"}
		child := processState{ProcessProof: ProcessProof{PID: 101, Start: "11", PGID: 100, SID: 50}, state: "S"}
		for _, signalErr := range []error{nil, unix.EPERM, unix.EINVAL} {
			sweeps, signals := 0, 0
			err := settleProcess(context.Background(), leader.ProcessProof, 0, func(ProcessProof) ([]processState, error) {
				sweeps++
				if sweeps == 1 {
					return []processState{leader, child}, nil
				}
				if sweeps < 4 {
					return []processState{child}, nil
				}
				return nil, nil
			}, func(m processState, s unix.Signal) error {
				signals++
				if m.PID != 101 || s != unix.SIGKILL {
					t.Fatal("wrong signaling target")
				}
				return signalErr
			})
			if signalErr != nil {
				if !errors.Is(err, signalErr) || sweeps != 1 {
					t.Fatal("unresolved signaling failure lost", err, sweeps)
				}
			} else if err != nil || sweeps != 4 || signals != 3 {
				t.Fatal("premature settlement", err, sweeps, signals)
			}
		}
		for _, altered := range []processState{child, {ProcessProof: ProcessProof{PID: 100, Start: "12", PGID: 100, SID: 50, Group: true}, state: "S"}} {
			ctx, cancel := context.WithTimeout(context.Background(), 25*time.Millisecond)
			defer cancel()
			if err := settleProcess(ctx, leader.ProcessProof, 0, func(ProcessProof) ([]processState, error) { return []processState{altered}, nil }, func(processState, unix.Signal) error { t.Fatal("ambiguous identity signaled"); return nil }); !errors.Is(err, ErrUnresolved) {
				t.Fatal("ambiguous identity accepted", err)
			}
		}
	})
	t.Run("unanchored_membership_requires_absence_within_existing_attempt", func(t *testing.T) {
		leader := ProcessProof{PID: 100, Start: "10", PGID: 100, SID: 50, Group: true}
		unknown := processState{ProcessProof: ProcessProof{PID: 101, Start: "11", PGID: 100, SID: 50}, state: "S"}
		for _, disappears := range []bool{true, false} {
			limit := 25 * time.Millisecond
			if disappears {
				limit = time.Second
			}
			ctx, cancel := context.WithTimeout(context.Background(), limit)
			defer cancel()
			sweeps := 0
			err := settleProcess(ctx, leader, 0, func(ProcessProof) ([]processState, error) {
				sweeps++
				if disappears && sweeps == 3 {
					return nil, nil
				}
				return []processState{unknown}, nil
			}, func(processState, unix.Signal) error {
				t.Fatal("unanchored observations became signaling authority")
				return nil
			})
			if disappears {
				if err != nil || sweeps != 3 {
					t.Fatal("verified absence within bounded settlement was rejected", sweeps, err)
				}
			} else if !errors.Is(err, ErrUnresolved) || !errors.Is(err, context.DeadlineExceeded) || DescribeCleanupFailure(err).Stage != "process_group_anchor" {
				t.Fatal("persistent ambiguity lost its existing deadline", sweeps, err)
			}
		}
		if err := settleProcess(context.Background(), leader, 0,
			func(ProcessProof) ([]processState, error) { return []processState{unknown}, nil },
			func(processState, unix.Signal) error { t.Fatal("unbounded ambiguity signaled"); return nil }); !errors.Is(err, ErrUnresolved) {
			t.Fatal("unbounded ambiguity must fail closed", err)
		}
	})
	t.Run("native_group_with_inherited_session_settles", func(t *testing.T) {
		root, err := filepath.Abs("../../..")
		if err != nil {
			t.Fatal(err)
		}
		run := t.TempDir()
		r, err := NewRuntime(root, run, "native-group", t.TempDir())
		if err != nil {
			t.Fatal(err)
		}
		defer r.Close()
		cmd := exec.Command("sleep", "60")
		cmd.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
		if err = cmd.Start(); err != nil {
			t.Fatal(err)
		}
		defer func() { _ = syscall.Kill(-cmd.Process.Pid, syscall.SIGKILL); _ = cmd.Wait() }()
		proof, err := harnessadmission.Identity(cmd.Process.Pid)
		if err != nil {
			t.Fatal(err)
		}
		target, err := CaptureProcess(cmd.Process.Pid)
		if err != nil {
			t.Fatal(err)
		}
		if err = r.RecordDependentResource("helper_process", target, "acquired", []string{}); err != nil {
			t.Fatal(err)
		}
		defer r.RecordResource("helper_process", target, "released")
		reused := target
		reused.Start = "0"
		if err := StopProcess(context.Background(), reused, 0); err == nil {
			t.Error("reused identity accepted")
		}
		if alive, err := harnessadmission.Alive(proof); err != nil || !alive {
			t.Fatal("unowned identity signaled", err)
		}
		resources, err := r.Resources(8)
		if err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err = r.StopResource(ctx, root, run, map[string]string{"NODE_BIN": os.Getenv("NODE_BIN")}, resources[0]); err != nil {
			t.Fatal(err)
		}
		if alive, err := harnessadmission.Alive(proof); err != nil || alive {
			t.Error("physical owner acknowledged a live native group", err)
		}
	})
}
