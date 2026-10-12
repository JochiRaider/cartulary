//go:build linux

package harnesssession

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"syscall"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessadmission"
	"golang.org/x/sys/unix"
)

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
		if err = r.RecordResource("managed_suite", proof, "acquired"); err != nil {
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
	t.Run("settlement_requires_observed_absence", func(t *testing.T) {
		leader := processState{ProcessProof: ProcessProof{PID: 100, Start: "10", PGID: 100, SID: 50, Group: true}, state: "Z"}
		child := processState{ProcessProof: ProcessProof{PID: 101, Start: "11", PGID: 100, SID: 50}, state: "S"}
		for _, denied := range []bool{false, true} {
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
				if denied {
					return unix.EPERM
				}
				return nil
			})
			if denied {
				if !errors.Is(err, unix.EPERM) || sweeps != 1 {
					t.Fatal("permission failure lost", err, sweeps)
				}
			} else if err != nil || sweeps != 4 || signals != 3 {
				t.Fatal("premature settlement", err, sweeps, signals)
			}
		}
		for _, altered := range []processState{child, {ProcessProof: ProcessProof{PID: 100, Start: "12", PGID: 100, SID: 50, Group: true}, state: "S"}} {
			if err := settleProcess(context.Background(), leader.ProcessProof, 0, func(ProcessProof) ([]processState, error) { return []processState{altered}, nil }, func(processState, unix.Signal) error { t.Fatal("ambiguous identity signaled"); return nil }); !errors.Is(err, ErrUnresolved) {
				t.Fatal("ambiguous identity accepted", err)
			}
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
		if err = r.RecordResource("helper_process", target, "acquired"); err != nil {
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
