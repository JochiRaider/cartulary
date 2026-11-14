//go:build linux

package harnesssession

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

type pipeCommandFixture struct {
	control *os.File
	events  *os.File
	decoder *json.Decoder
	done    <-chan int
}

func newPipeCommandFixture(t *testing.T, command *exec.Cmd, budget time.Duration, captures ...*commandCapture) *pipeCommandFixture {
	var capture *commandCapture
	if len(captures) != 0 {
		capture = captures[0]
	}
	t.Helper()
	control, write, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	read, events, err := os.Pipe()
	if err != nil {
		control.Close()
		write.Close()
		t.Fatal(err)
	}
	parent, err := CaptureProcess(os.Getpid())
	if err != nil {
		t.Fatal(err)
	}
	done := make(chan int, 1)
	finished := make(chan struct{})
	fixture := &pipeCommandFixture{write, read, json.NewDecoder(read), done}
	t.Cleanup(func() {
		write.Close()
		select {
		case <-finished:
		case <-time.After(commandProtocolBudget):
			t.Error("pipe command cleanup did not finish")
		}
		read.Close()
	})
	go func() {
		done <- servePipeCommand(command, control, events, parent, budget, capture)
		events.Close()
		close(finished)
	}()
	return fixture
}

func (fixture *pipeCommandFixture) event(t *testing.T) commandPipeEvent {
	t.Helper()
	if err := fixture.events.SetReadDeadline(time.Now().Add(commandProtocolBudget)); err != nil {
		t.Fatal(err)
	}
	var event commandPipeEvent
	if err := fixture.decoder.Decode(&event); err != nil {
		t.Fatal("pipe event", err)
	}
	return event
}

func (fixture *pipeCommandFixture) ready(t *testing.T) ProcessProof {
	t.Helper()
	event := fixture.event(t)
	if event.Event != "ready" || event.Proof == nil || event.Exit != nil || event.Cause != "" || event.Signal != nil || event.Diagnostic != nil {
		t.Fatal("invalid ready event", event)
	}
	actual, err := CaptureProcess(event.Proof.PID)
	if err != nil || actual != *event.Proof || !actual.Group {
		t.Fatal("ready proof does not identify retained child", err)
	}
	return actual
}

func (fixture *pipeCommandFixture) settled(t *testing.T, proof ProcessProof, cause string, exit, status int) commandPipeEvent {
	t.Helper()
	event := fixture.event(t)
	if event.Event != "settled" || event.Proof != nil || event.Exit == nil || *event.Exit != exit || event.Cause != cause || event.Diagnostic != nil {
		t.Fatal("invalid settled event", event)
	}
	awaitCommandControl(t, fixture.done, status)
	if _, err := readProcess(proof.PID); !processGone(err) {
		t.Fatal("settled command leader was not reaped", err)
	}
	var extra any
	if err := fixture.decoder.Decode(&extra); err != io.EOF {
		t.Fatal("event stream did not terminate exactly once", err)
	}
	return event
}

func testCommandPipe(t *testing.T) {
	t.Run("bounded_capture", testCommandCapture)
	t.Run("pipe_original_exit_distinguishes_signal_from_equal_numeric_status", func(t *testing.T) {
		for _, fixtureCase := range []struct {
			script       string
			exit, signal int
		}{{"exit 143", 143, 0}, {"kill -TERM $$", 143, 15}, {"kill -34 $$", 162, 34}} {
			fixture := newPipeCommandFixture(t, exec.Command("sh", "-c", fixtureCase.script), 8*time.Second)
			proof := fixture.ready(t)
			fmt.Fprintln(fixture.control, "adopt")
			event := fixture.settled(t, proof, "exited", fixtureCase.exit, 0)
			if fixtureCase.signal == 0 && event.Signal != nil || fixtureCase.signal != 0 && (event.Signal == nil || *event.Signal != fixtureCase.signal) {
				t.Fatal("native wait signal fact lost", event.Signal, fixtureCase.signal)
			}
		}
	})
	t.Run("pipe_registration_retains_exited_anchor_then_settles_descendants", func(t *testing.T) {
		peer := ownedCommandFixture(t, exec.Command("sleep", "60"))
		file := filepath.Join(t.TempDir(), "descendant")
		fixture := newPipeCommandFixture(t, exec.Command("sh", "-c", `sleep 60 </dev/null >/dev/null 2>&1 & echo $! > "$1"; exit 17`, "fixture", file), 8*time.Second)
		proof := fixture.ready(t)
		deadline := time.Now().Add(5 * time.Second)
		for {
			state, err := readProcess(proof.PID)
			if err != nil || !sameProcess(state.ProcessProof, proof) {
				t.Fatal("registration lost child anchor", err)
			}
			if state.state == "Z" {
				break
			}
			if time.Now().After(deadline) {
				t.Fatal("fixture did not exit")
			}
			time.Sleep(time.Millisecond)
		}
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatal(err)
		}
		pid, err := strconv.Atoi(strings.TrimSpace(string(data)))
		if err != nil {
			t.Fatal(err)
		}
		child, err := readProcess(pid)
		if err != nil || child.PGID != proof.PGID || child.state == "Z" {
			t.Fatal("fixture descendant not live before adoption", err)
		}
		fmt.Fprintln(fixture.control, "adopt")
		fixture.settled(t, proof, "exited", 17, 0)
		child, err = readProcess(pid)
		if err != nil && !processGone(err) || err == nil && child.state != "Z" && child.state != "X" {
			t.Fatal("settled acknowledgement left live descendant", err)
		}
		state, err := readProcess(peer.proof.PID)
		if err != nil || !sameProcess(state.ProcessProof, peer.proof) || state.state == "Z" {
			t.Fatal("unrelated process affected", err)
		}
	})
	for _, fixtureCase := range []struct {
		name, frames, cause string
		closeControl        bool
		status              int
	}{
		{"cancel_before_adoption", "stop\n", "stopped", false, 0},
		{"cancel_after_adoption", "adopt\nstop\n", "stopped", false, 0},
		{"unknown", "detach\n", "protocol", false, 12},
		{"oversized", strings.Repeat("x", 32), "protocol", false, 12},
		{"repeated_adoption", "adopt\nadopt\n", "protocol", false, 12},
		{"truncated_adoption", "adopt", "protocol", true, 12},
		{"control_loss", "", "control_lost", true, 12},
		{"adopted_control_loss", "adopt\n", "control_lost", true, 12},
	} {
		t.Run("pipe_"+fixtureCase.name, func(t *testing.T) {
			fixture := newPipeCommandFixture(t, exec.Command("sleep", "60"), 8*time.Second)
			proof := fixture.ready(t)
			if _, err := io.WriteString(fixture.control, fixtureCase.frames); err != nil {
				t.Fatal(err)
			}
			if fixtureCase.closeControl {
				fixture.control.Close()
			}
			fixture.settled(t, proof, fixtureCase.cause, 143, fixtureCase.status)
		})
	}
	t.Run("pipe_missing_adoption_is_failure_after_settlement", func(t *testing.T) {
		fixture := newPipeCommandFixture(t, exec.Command("sleep", "60"), 100*time.Millisecond)
		proof := fixture.ready(t)
		fixture.settled(t, proof, "adoption_timeout", 143, 12)
	})
	t.Run("pipe_missing_executable_has_no_ready_or_settled_claim", func(t *testing.T) {
		fixture := newPipeCommandFixture(t, exec.Command(filepath.Join(t.TempDir(), "absent")), time.Second)
		event := fixture.event(t)
		if event.Event != "failed" || event.Cause != "launch_missing" || event.Proof != nil || event.Exit != nil || event.Diagnostic != nil {
			t.Fatal("failed launch misclassified", event)
		}
		awaitCommandControl(t, fixture.done, 12)
	})
	t.Run("pipe_regular_file_cannot_be_control_or_evidence", func(t *testing.T) {
		file, err := os.CreateTemp(t.TempDir(), "invalid")
		if err != nil {
			t.Fatal(err)
		}
		defer file.Close()
		command := exec.Command("sleep", "60")
		if servePipeCommand(command, file, file, ProcessProof{}, time.Second, nil) != 12 || command.Process != nil {
			t.Fatal("invalid descriptors launched command")
		}
	})
	t.Run("pipe_invalid_direction_or_parent_rejects_before_launch", func(t *testing.T) {
		control, writer, err := os.Pipe()
		if err != nil {
			t.Fatal(err)
		}
		defer control.Close()
		defer writer.Close()
		reader, events, err := os.Pipe()
		if err != nil {
			t.Fatal(err)
		}
		defer reader.Close()
		defer events.Close()
		parent, err := CaptureProcess(os.Getpid())
		if err != nil {
			t.Fatal(err)
		}
		for _, pair := range [][2]*os.File{{writer, events}, {control, reader}, {control, control}} {
			command := exec.Command("sleep", "60")
			if servePipeCommand(command, pair[0], pair[1], parent, time.Second, nil) != 12 || command.Process != nil {
				t.Fatal("invalid channel direction acquired child")
			}
		}
		parent.Start = "0"
		command := exec.Command("sleep", "60")
		if servePipeCommand(command, control, events, parent, time.Second, nil) != 12 || command.Process != nil {
			t.Fatal("mismatched parent acquired child")
		}
		var event commandPipeEvent
		reader.SetReadDeadline(time.Now().Add(time.Second))
		if json.NewDecoder(reader).Decode(&event) != nil || event.Event != "failed" || event.Cause != "parent" || event.Exit != nil || event.Proof != nil {
			t.Fatal("mismatched parent claimed acquisition or settlement", event)
		}
	})
	t.Run("pipe_parent_loss_settles_even_with_control_writer_retained", testPipeParentLoss)
	t.Run("pipe_native_dispatch_preserves_workspace_and_closes_private_channels", testPipeWorkspace)
}

func commandPipeParentFixture() {
	control, write, err := os.Pipe()
	if err != nil {
		return
	}
	defer control.Close()
	defer write.Close()
	read, events, err := os.Pipe()
	if err != nil {
		return
	}
	defer read.Close()
	defer events.Close()
	command := exec.Command(os.Args[0], "-test.run=^TestSessionRuntimeProof$", "--", "pipe", "sleep", "60")
	command.Env = append(os.Environ(), "CARTULARY_OWNED_COMMAND_TEST=serve")
	command.SysProcAttr = &syscall.SysProcAttr{Setsid: true}
	command.Stdin = control
	// Keep a writer in the supervisor so parent death cannot be detected merely
	// by EOF. The workload receives neither this writer nor its event channel.
	command.ExtraFiles = []*os.File{events, write}
	if command.Start() != nil {
		return
	}
	var ready commandPipeEvent
	if json.NewDecoder(read).Decode(&ready) != nil || ready.Event != "ready" || ready.Proof == nil {
		return
	}
	supervisor, err := CaptureProcess(command.Process.Pid)
	if err != nil {
		return
	}
	if json.NewEncoder(os.Stdout).Encode(controlParentFixture{Payload: *ready.Proof, Supervisor: supervisor}) != nil {
		return
	}
	for {
		time.Sleep(time.Hour)
	}
}

func testPipeWorkspace(t *testing.T) {
	lease, err := os.CreateTemp(t.TempDir(), "workspace")
	if err != nil {
		t.Fatal(err)
	}
	defer lease.Close()
	if err := unix.Flock(int(lease.Fd()), unix.LOCK_SH|unix.LOCK_NB); err != nil {
		t.Fatal(err)
	}
	var stat unix.Stat_t
	if err := unix.Fstat(int(lease.Fd()), &stat); err != nil {
		t.Fatal(err)
	}
	control, write, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer control.Close()
	defer write.Close()
	read, events, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	defer read.Close()
	defer events.Close()
	command := exec.Command(os.Args[0], "-test.run=^TestSessionRuntimeProof$", "--", "pipe", "sh", "-c", `
test "$(stat -Lc '%d:%i' /proc/self/fd/9)" = "$1" || exit 40
grep -Eq 'FLOCK[[:space:]]+ADVISORY[[:space:]]+READ' /proc/self/fdinfo/9 || exit 41
if read -r value; then exit 42; fi
test "$(readlink /proc/self/fd/3)" = /dev/null || exit 43
exit 17`, "fixture", fmt.Sprintf("%d:%d", stat.Dev, stat.Ino))
	command.Env = append(os.Environ(), "CARTULARY_OWNED_COMMAND_TEST=serve")
	command.Stdin = control
	command.ExtraFiles = []*os.File{events, nil, nil, nil, nil, nil, lease}
	supervisor := ownedCommandFixture(t, command)
	control.Close()
	events.Close()
	done := make(chan int, 1)
	fixture := &pipeCommandFixture{write, read, json.NewDecoder(read), done}
	proof := fixture.ready(t)
	fmt.Fprintln(write, "adopt")
	go func() {
		<-supervisor.exited
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		code, err := supervisor.Settle(ctx, 0)
		if err != nil {
			code = -1
		}
		done <- code
	}()
	fixture.settled(t, proof, "exited", 17, 0)
}

func testPipeParentLoss(t *testing.T) {
	command := exec.Command(os.Args[0], "-test.run=^TestSessionRuntimeProof$")
	command.Env = append(os.Environ(), "CARTULARY_OWNED_COMMAND_TEST=pipe-parent")
	read, err := command.StdoutPipe()
	if err != nil {
		t.Fatal(err)
	}
	parent := ownedCommandFixture(t, command)
	var fixture controlParentFixture
	decoded := make(chan error, 1)
	go func() { decoded <- json.NewDecoder(read).Decode(&fixture) }()
	select {
	case err := <-decoded:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(8 * time.Second):
		t.Fatal("pipe parent fixture did not become ready")
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := StopProcess(ctx, fixture.Supervisor, 0); err != nil {
			t.Error("pipe supervisor cleanup", err)
		}
	})
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if _, err := parent.Settle(ctx, 0); err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(commandProtocolBudget)
	for {
		payload, payloadErr := readProcess(fixture.Payload.PID)
		supervisor, supervisorErr := readProcess(fixture.Supervisor.PID)
		payloadGone := processGone(payloadErr) || payloadErr == nil && payload.state == "Z"
		supervisorGone := processGone(supervisorErr) || supervisorErr == nil && supervisor.state == "Z"
		if payloadGone && supervisorGone {
			return
		}
		if time.Now().After(deadline) {
			t.Fatal("parent loss did not settle pipe-owned processes", payloadErr, supervisorErr)
		}
		time.Sleep(20 * time.Millisecond)
	}
}
