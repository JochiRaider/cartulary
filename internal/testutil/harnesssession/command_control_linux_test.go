//go:build linux

package harnesssession

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
)

type controlParentFixture struct {
	Directory  string       `json:"directory"`
	Payload    ProcessProof `json:"payload"`
	Supervisor ProcessProof `json:"supervisor"`
}

func commandControlFixtureMode() {
	mode := os.Getenv("CARTULARY_OWNED_COMMAND_TEST")
	if mode == "serve" {
		for i, arg := range os.Args {
			if arg == "--" {
				os.Exit(OwnedCommandMain(os.Args[i+1:], os.Stdout, os.Stderr))
			}
		}
		os.Exit(2)
	}
	if mode == "pipe-parent" {
		commandPipeParentFixture()
		os.Exit(12)
	}
	if mode != "parent" {
		return
	}
	var buffer bytes.Buffer
	if OwnedCommandMain([]string{"allocate"}, &buffer, os.Stderr) != 0 {
		os.Exit(12)
	}
	name := strings.TrimSpace(buffer.String())
	child := exec.Command(os.Args[0], "-test.run=^TestSessionRuntimeProof$", "--", "serve", name, "sleep", "60")
	child.Env = append(os.Environ(), "CARTULARY_OWNED_COMMAND_TEST=serve")
	child.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	if child.Start() != nil {
		os.Exit(12)
	}
	supervisor, err := CaptureProcess(child.Process.Pid)
	if err != nil {
		child.Process.Kill()
		child.Wait()
		os.Exit(12)
	}
	buffer.Reset()
	if OwnedCommandMain([]string{"ready", name}, &buffer, os.Stderr) != 0 {
		ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
		StopProcess(ctx, supervisor, time.Second)
		cancel()
		child.Wait()
		os.Exit(12)
	}
	pid, _ := strconv.Atoi(strings.TrimSpace(buffer.String()))
	payload, err := CaptureProcess(pid)
	if err != nil {
		os.Exit(12)
	}
	json.NewEncoder(os.Stdout).Encode(controlParentFixture{Directory: name, Payload: payload, Supervisor: supervisor})
	for {
		time.Sleep(time.Hour)
	}
}

func commandControlFixture(t *testing.T, command *exec.Cmd) (*os.File, <-chan int) {
	t.Helper()
	name := filepath.Join(t.TempDir(), "command")
	directory, err := harnesslocal.PrivateDirectory(name)
	if err != nil {
		t.Fatal(err)
	}
	parent, err := CaptureProcess(os.Getpid())
	if err != nil {
		directory.Close()
		t.Fatal(err)
	}
	done := make(chan int, 1)
	finished := make(chan struct{})
	t.Cleanup(func() {
		defer directory.Close()
		select {
		case <-finished:
			return
		default:
		}
		if code := commandClient(directory, "stop", io.Discard); code != 0 {
			t.Error("control fixture settlement", code)
		}
		select {
		case <-finished:
		case <-time.After(8 * time.Second):
			t.Error("control owner did not exit")
		}
	})
	go func() {
		done <- serveOwnedCommand(directory, command, parent, io.Discard)
		close(finished)
	}()
	return directory, done
}

func awaitCommandControl(t *testing.T, done <-chan int, expected int) {
	t.Helper()
	select {
	case code := <-done:
		if code != expected {
			t.Fatal("control native result", code, expected)
		}
	case <-time.After(8 * time.Second):
		t.Fatal("control owner did not settle")
	}
}

func testCommandControl(t *testing.T) {
	t.Run("command_control_original_exit_and_descendant_settlement", func(t *testing.T) {
		file := filepath.Join(t.TempDir(), "descendant")
		directory, done := commandControlFixture(t, exec.Command("sh", "-c", `sleep 60 </dev/null >/dev/null 2>&1 & echo $! > "$1"; exit 17`, "fixture", file))
		var ready bytes.Buffer
		if code := commandClient(directory, "ready", &ready); code != 0 {
			t.Fatal("ready", code)
		}
		awaitCommandControl(t, done, 0)
		if commandClient(directory, "running", io.Discard) != 1 || commandClient(directory, "result", io.Discard) != 17 {
			t.Fatal("settled status lost native exit")
		}
		for range 2 {
			if commandClient(directory, "stop", io.Discard) != 0 {
				t.Fatal("repeated stop")
			}
		}
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		pid, _ := strconv.Atoi(strings.TrimSpace(string(data)))
		state, err := readProcess(pid)
		if err != nil && !processGone(err) || err == nil && state.state != "Z" && state.state != "X" {
			t.Fatal("live descendant after acknowledgement", state, err)
		}
		root, _ := strconv.Atoi(strings.TrimSpace(ready.String()))
		if _, err := readProcess(root); !processGone(err) {
			t.Fatal("leader not reaped", err)
		}
	})
	t.Run("command_control_stop_and_closed_protocol", func(t *testing.T) {
		directory, done := commandControlFixture(t, exec.Command("sleep", "60"))
		if commandClient(directory, "ready", io.Discard) != 0 || commandClient(directory, "running", io.Discard) != 0 {
			t.Fatal("live command unavailable")
		}
		if commandClient(directory, "detach", io.Discard) != 12 || commandClient(directory, "running", io.Discard) != 0 {
			t.Fatal("invalid protocol changed command lifetime")
		}
		if commandClient(directory, "stop", io.Discard) != 0 {
			t.Fatal("stop failed")
		}
		awaitCommandControl(t, done, 0)
		if commandClient(directory, "result", io.Discard) != 143 {
			t.Fatal("signal exit not preserved")
		}
		data, _ := harnesslocal.ReadFileAt(directory, "settled", 4096, false)
		var receipt commandReceipt
		json.Unmarshal(data, &receipt)
		receipt.Parent.Start = "0"
		data, _ = json.Marshal(receipt)
		if harnesslocal.ReplaceFileAt(directory, "settled", data) != nil || commandClient(directory, "stop", io.Discard) != 12 {
			t.Fatal("foreign receipt admitted")
		}
	})
	t.Run("command_control_failed_exec_and_invalid_handle", func(t *testing.T) {
		directory, done := commandControlFixture(t, exec.Command(filepath.Join(t.TempDir(), "missing")))
		if commandClient(directory, "ready", io.Discard) != 12 {
			t.Fatal("failed exec admitted")
		}
		awaitCommandControl(t, done, 12)
		if commandClient(directory, "stop", io.Discard) != 0 {
			t.Fatal("failed acquisition not acknowledged")
		}
		if OwnedCommandMain([]string{"release", t.TempDir()}, io.Discard, io.Discard) != 12 {
			t.Fatal("foreign directory admitted")
		}
		old, err := harnesslocal.OpenDirectory(directory.Name(), true)
		if err != nil {
			t.Fatal(err)
		}
		defer old.Close()
		parent, err := CaptureProcess(os.Getpid())
		if err != nil {
			t.Fatal(err)
		}
		second := exec.Command("sleep", "60")
		if serveOwnedCommand(old, second, parent, io.Discard) != 12 || second.Process != nil {
			t.Fatal("used capability acquired a second workload")
		}
	})
	t.Run("command_control_parent_loss_and_foreign_client", func(t *testing.T) {
		command := exec.Command(os.Args[0], "-test.run=^TestSessionRuntimeProof$")
		command.Env = append(os.Environ(), "CARTULARY_OWNED_COMMAND_TEST=parent")
		stdout, err := command.StdoutPipe()
		if err != nil {
			t.Fatal(err)
		}
		parent := ownedCommandFixture(t, command)
		var fixture controlParentFixture
		read := make(chan error, 1)
		go func() { read <- json.NewDecoder(bufio.NewReader(stdout)).Decode(&fixture) }()
		select {
		case err := <-read:
			if err != nil {
				t.Fatal(err)
			}
		case <-time.After(8 * time.Second):
			t.Fatal("parent fixture did not publish ownership")
		}
		t.Cleanup(func() {
			ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
			defer cancel()
			if err := StopProcess(ctx, fixture.Supervisor, time.Second); err != nil {
				t.Error("supervisor fixture cleanup", err)
			}
		})
		if OwnedCommandMain([]string{"stop", fixture.Directory}, io.Discard, io.Discard) != 12 {
			t.Fatal("unrelated ancestry admitted")
		}
		if state, err := readProcess(fixture.Payload.PID); err != nil || state.state == "Z" || !sameProcess(state.ProcessProof, fixture.Payload) {
			t.Fatal("rejected client changed payload", state, err)
		}
		ctx, cancel := context.WithTimeout(context.Background(), 6*time.Second)
		defer cancel()
		if _, err := parent.Settle(ctx, 0); err != nil {
			t.Fatal(err)
		}
		deadline := time.Now().Add(8 * time.Second)
		for {
			state, err := readProcess(fixture.Payload.PID)
			_, directoryErr := os.Lstat(fixture.Directory)
			if (processGone(err) || err == nil && (state.state == "Z" || state.state == "X")) && os.IsNotExist(directoryErr) {
				break
			}
			if time.Now().After(deadline) {
				t.Fatal("parent loss did not settle payload and private handle", state, err, directoryErr)
			}
			time.Sleep(10 * time.Millisecond)
		}
	})
}
