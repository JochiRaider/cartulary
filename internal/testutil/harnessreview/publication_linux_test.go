//go:build linux

package harnessreview

import (
	"context"
	"encoding/json"
	"errors"
	"net"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessadmission"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
	"github.com/JochiRaider/cartulary/internal/testutil/harnesssession"
	"golang.org/x/sys/unix"
)

func publicationFixture(t *testing.T) (*diskPublisher, Locator) {
	t.Helper()
	repo, err := filepath.Abs("../../..")
	if err != nil {
		t.Fatal(err)
	}
	run, scratch := filepath.Join(t.TempDir(), "fixture"), t.TempDir()
	if err := os.Mkdir(run, 0700); err != nil {
		t.Fatal(err)
	}
	runtime, err := harnesssession.NewRuntime(repo, run, "fixture", scratch)
	if err != nil {
		t.Fatal(err)
	}
	id := "uireview-" + nonce(16)
	socket := filepath.Join(registryRoot(), strings.TrimPrefix(id, "uireview-")+".sock")
	if err = os.MkdirAll(registryRoot(), 0700); err != nil {
		t.Fatal(err)
	}
	listener, err := net.ListenUnix("unix", &net.UnixAddr{Name: socket, Net: "unix"})
	if err != nil {
		t.Fatal(err)
	}
	listener.SetUnlinkOnClose(false)
	if err = os.Chmod(socket, 0600); err != nil {
		t.Fatal(err)
	}
	var stat unix.Stat_t
	if err = unix.Lstat(socket, &stat); err != nil {
		t.Fatal(err)
	}
	proof, err := harnessadmission.Identity(os.Getpid())
	if err != nil {
		t.Fatal(err)
	}
	record := &registryRecord{Schema: registrySchema, Session: id, Locator: filepath.Join(run, "ui-review/session.json"), Repository: repo, Process: proof, StartedTick: "1", Token: nonce(32), Socket: socket, Device: stat.Dev, Inode: stat.Ino, Runtime: runtimeRef{runtime.Root, runtime.LeaseID, runtime.RunID}}
	record.Presentation = "ui"
	if err = writeRegistry(record, false); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		listener.Close()
		os.Remove(socket)
		os.Remove(registryPath(id))
		_ = runtime.Close()
		runtime.ReleaseHandles()
	})
	p := &diskPublisher{runtime: runtime, record: record, runRoot: run}
	l := Locator{Schema: "cartulary.ui_review_session.v1", Session: id, RunID: "fixture", Mode: "artifacts", State: "ready", Created: timestamp(time.Now()), Updated: timestamp(time.Now())}
	if err = p.Locator(l); err != nil {
		t.Fatal(err)
	}
	return p, l
}
func TestReviewPublication(t *testing.T) {
	t.Run("existing_terminal_is_never_overwritten", func(t *testing.T) {
		p, l := publicationFixture(t)
		l.State = "closed"
		file := filepath.Join(p.runRoot, "ui-review/terminal.json")
		original := []byte("existing immutable artifact\n")
		if err := os.WriteFile(file, original, 0600); err != nil {
			t.Fatal(err)
		}
		receipt := Receipt{Schema: "cartulary.ui_review_receipt.v2", Command: commandID("ui-review"), Session: l.Session, Mode: l.Mode, State: "closed", Status: "ok", Started: l.Created, Finished: l.Updated, Failures: []Failure{}, Cleanup: "complete"}
		if _, err := p.Terminal(l, receipt, false); err == nil {
			t.Fatal("existing terminal replaced")
		}
		data, err := os.ReadFile(file)
		if err != nil || string(data) != string(original) {
			t.Fatal("immutable artifact changed", err)
		}
		if _, err = os.Stat(registryPath(l.Session)); err != nil {
			t.Fatal("publication recovery lost", err)
		}
	})
	t.Run("failed_publication_preserves_primary_through_exact_recovery", func(t *testing.T) {
		p, l := publicationFixture(t)
		original := contextual("build_failed", "build", "frontend", "child_failed", "build_web")
		l.State = "failed"
		if err := p.runtime.RegisterSecret("private-publication-sentinel"); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(p.runRoot, "unsafe.log"), []byte("private-publication-sentinel"), 0600); err != nil {
			t.Fatal(err)
		}
		receipt := Receipt{Schema: "cartulary.ui_review_receipt.v2", Command: commandID("ui-review"), Session: l.Session, Mode: l.Mode, State: "failed", Status: "error", Exit: 1, Started: l.Created, Finished: l.Updated, Failures: []Failure{original.Record}, Cleanup: "complete"}
		if _, err := p.Terminal(l, receipt, false); err == nil {
			t.Fatal("unsafe publication accepted")
		}
		if err := p.Locator(l); err != nil {
			t.Fatal(err)
		}
		// Exact ownership from an earlier boot cannot name a live process.
		dead := harnessadmission.Proof{Boot: "dead-fixture-boot", PID: 12345, Start: "1"}
		data, _ := json.Marshal(dead)
		if err := harnesslocal.ReplaceFile(filepath.Join(p.runtime.Root, "runtime-process.json"), data); err != nil {
			t.Fatal(err)
		}
		p.record.Process = dead
		if err := writeRegistry(p.record, true); err != nil {
			t.Fatal(err)
		}
		q := commandRequest{Command: "ui-review-stop", Locator: p.record.Locator}
		first, err := control(context.Background(), p.record.Repository, q, map[string]string{})
		if err != nil || first.Receipt == nil || first.Exit != 1 || len(first.Failures) != 2 || first.Failures[0].Code != "build_failed" || first.Failures[1].Code != "unsafe_artifact" {
			t.Fatal(first, err)
		}
		second, err := control(context.Background(), p.record.Repository, q, map[string]string{})
		if err != nil || second.Receipt == nil || *second.Receipt != *first.Receipt {
			t.Fatal(second, err)
		}
		if _, err := os.Stat(p.runtime.Root); !errors.Is(err, os.ErrNotExist) {
			t.Fatal("private state survived recovery", err)
		}
	})
	t.Run("terminal_binding_and_immutable_repetition", func(t *testing.T) {
		p, l := publicationFixture(t)
		l.State = "closed"
		receipt := Receipt{Schema: "cartulary.ui_review_receipt.v2", Command: commandID("ui-review"), Session: l.Session, Mode: l.Mode, State: "closed", Status: "ok", Started: l.Created, Finished: l.Updated, Failures: []Failure{}, Cleanup: "complete"}
		ref, err := p.Terminal(l, receipt, false)
		if err != nil {
			t.Fatal(err)
		}
		l.Receipt = ref
		v, err := terminalResult(p.record.Locator, l, "ui-review-stop")
		if err != nil || v.Exit != 0 || v.Receipt.SHA256 != ref.SHA256 {
			t.Fatal(v, err)
		}
		if _, err = os.Stat(p.runtime.Root); !errors.Is(err, os.ErrNotExist) {
			t.Fatal("root survived", err)
		}
		if _, err = os.Stat(registryPath(l.Session)); !errors.Is(err, os.ErrNotExist) {
			t.Fatal("registry survived", err)
		}
		forged := *ref
		forged.Bytes++
		l.Receipt = &forged
		if _, err = terminalResult(p.record.Locator, l, "ui-review-status"); err == nil {
			t.Fatal("forged receipt accepted")
		}
	})
	t.Run("scan_failure_keeps_recovery_and_publishes_no_terminal", func(t *testing.T) {
		p, l := publicationFixture(t)
		l.State = "closed"
		if err := p.runtime.RegisterSecret("private-sentinel-123"); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(filepath.Join(p.runRoot, "unsafe.log"), []byte("private-sentinel-123"), 0600); err != nil {
			t.Fatal(err)
		}
		receipt := Receipt{Schema: "cartulary.ui_review_receipt.v2", Command: commandID("ui-review"), Session: l.Session, Mode: l.Mode, State: "closed", Status: "ok", Started: l.Created, Finished: l.Updated, Failures: []Failure{}, Cleanup: "complete"}
		_, err := p.Terminal(l, receipt, false)
		if err == nil {
			t.Fatal("scan failure accepted")
		}
		if _, err = os.Stat(filepath.Join(p.runRoot, "unsafe.log")); !errors.Is(err, os.ErrNotExist) {
			t.Fatal("unsafe evidence survived scan", err)
		}
		if _, err = os.Stat(registryPath(l.Session)); err != nil {
			t.Fatal("recovery lost", err)
		}
		if _, err = os.Stat(filepath.Join(p.runRoot, "ui-review/terminal.json")); !errors.Is(err, os.ErrNotExist) {
			t.Fatal("terminal created before scan", err)
		}
	})
	t.Run("concurrent_recovery_lock_is_exclusive", func(t *testing.T) {
		p, _ := publicationFixture(t)
		release, err := recoveryLock(context.Background(), p.record.Session)
		if err != nil {
			t.Fatal(err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 20*time.Millisecond)
		defer cancel()
		if second, err := recoveryLock(ctx, p.record.Session); err == nil {
			second()
			t.Fatal("parallel recovery admitted")
		}
		release()
		os.Remove(registryPath(p.record.Session) + ".lock")
	})
}
