//go:build linux

package harnesssession

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/harnessadmission"
	"golang.org/x/sys/unix"
)

// The asynchronous transport has no PID-addressed operations or filesystem
// capability. Its parent registers the ready proof before adopting the launch;
// even an already-exited leader remains unreaped until that handshake finishes.
type commandPipeEvent struct {
	Event          string             `json:"event"`
	Proof          *ProcessProof      `json:"proof,omitempty"`
	Exit           *int               `json:"exit,omitempty"`
	Cause          string             `json:"cause,omitempty"`
	Signal         *int               `json:"signal,omitempty"`
	Diagnostic     *CleanupDiagnostic `json:"diagnostic,omitempty"`
	CaptureFailure string             `json:"capture_failure,omitempty"`
}

func commandParent(parent ProcessProof) (int, error) {
	fd, err := unix.PidfdOpen(parent.PID, 0)
	if err != nil {
		return -1, err
	}
	current, err := CaptureProcess(parent.PID)
	if err == nil && current != parent {
		err = ErrUnresolved
	}
	if err == nil {
		var gone bool
		gone, err = commandParentGone(fd)
		if gone {
			err = ErrUnresolved
		}
	}
	if err != nil {
		unix.Close(fd)
		return -1, err
	}
	return fd, nil
}

func commandParentGone(fd int) (bool, error) {
	for {
		poll := []unix.PollFd{{Fd: int32(fd), Events: unix.POLLIN}}
		_, err := unix.Poll(poll, 0)
		if errors.Is(err, unix.EINTR) {
			continue
		}
		return poll[0].Revents != 0, err
	}
}

func commandPipe(file *os.File, write bool) bool {
	if file == nil {
		return false
	}
	var stat unix.Stat_t
	if unix.Fstat(int(file.Fd()), &stat) != nil {
		return false
	}
	mode := stat.Mode & unix.S_IFMT
	flags, err := unix.FcntlInt(file.Fd(), unix.F_GETFL, 0)
	if err != nil || mode != unix.S_IFIFO && mode != unix.S_IFSOCK {
		return false
	}
	access := flags & unix.O_ACCMODE
	return access == unix.O_RDWR || write && access == unix.O_WRONLY || !write && access == unix.O_RDONLY
}

func servePipeCommand(command *exec.Cmd, control, events *os.File, parent ProcessProof, adoptionBudget time.Duration, capture *commandCapture) int {
	if !commandPipe(control, false) || !commandPipe(events, true) || control.Fd() == events.Fd() || adoptionBudget <= 0 {
		return 12
	}
	// Neither private channel is a workload descriptor. ExtraFiles is populated
	// only by the existing workspace-admission projection below.
	unix.CloseOnExec(int(control.Fd()))
	unix.CloseOnExec(int(events.Fd()))
	publish := func(event commandPipeEvent) error {
		if err := events.SetWriteDeadline(time.Now().Add(time.Second)); err != nil {
			return err
		}
		return json.NewEncoder(events).Encode(event)
	}
	fail := func(cause string, failure error) int {
		_ = publish(commandPipeEvent{Event: "failed", Cause: cause, Diagnostic: DescribeCleanupFailure(failure), CaptureFailure: capture.Failure()})
		return 12
	}
	parentFD, err := commandParent(parent)
	if err != nil {
		return fail("parent", err)
	}
	defer unix.Close(parentFD)
	ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer cancel()
	command.Stdin = nil
	command.SysProcAttr = &syscall.SysProcAttr{Setpgid: true}
	release, err := harnessadmission.InheritWorkspace(command)
	if err != nil {
		return fail("admission", err)
	}
	defer release()
	owner, err := StartOwnedCommand(command, nil)
	if err != nil {
		if command.Process != nil {
			// A post-launch proof/cleanup failure cannot claim that acquisition
			// never happened, even when no ready frame could be published.
			return fail("cleanup", err)
		}
		cause := "launch"
		if errors.Is(err, os.ErrNotExist) || errors.Is(err, exec.ErrNotFound) {
			cause = "launch_missing"
		}
		return fail(cause, nil)
	}
	settle := func(cause string, status int) int {
		settlement, stop := context.WithTimeout(context.Background(), commandSettlementBudget)
		code, failure := owner.Settle(settlement, commandGrace)
		stop()
		if failure != nil {
			return fail("cleanup", failure)
		}
		captureFailure := capture.Failure()
		if cause == "exited" && captureFailure != "" {
			cause, status = "capture", 12
		}
		event := commandPipeEvent{Event: "settled", Exit: &code, Cause: cause, CaptureFailure: captureFailure}
		if status, ok := owner.command.ProcessState.Sys().(syscall.WaitStatus); ok && status.Signaled() {
			signal := int(status.Signal())
			event.Signal = &signal
		}
		if publish(event) != nil {
			return 12
		}
		return status
	}
	if publish(commandPipeEvent{Event: "ready", Proof: &owner.proof}) != nil {
		return settle("protocol", 12)
	}
	operations := make(chan string)
	finished := make(chan struct{})
	defer close(finished)
	defer control.Close()
	go func() {
		scanner := bufio.NewScanner(control)
		scanner.Buffer(make([]byte, 16), 16)
		scanner.Split(splitCommandControl)
		for scanner.Scan() {
			select {
			case operations <- scanner.Text():
			case <-finished:
				return
			}
		}
		operation := "control_lost"
		if scanner.Err() != nil {
			operation = "protocol"
		}
		select {
		case operations <- operation:
		case <-finished:
		}
	}()
	adoption := time.NewTimer(adoptionBudget)
	defer adoption.Stop()
	poll := time.NewTicker(25 * time.Millisecond)
	defer poll.Stop()
	var exited <-chan struct{}
	var captureFailed <-chan struct{}
	for {
		select {
		case operation := <-operations:
			switch operation {
			case "adopt":
				if exited != nil {
					return settle("protocol", 12)
				}
				exited = owner.exited
				if capture != nil {
					captureFailed = capture.failed
				}
				adoption.Stop()
			case "stop":
				return settle("stopped", 0)
			case "control_lost":
				return settle("control_lost", 12)
			default:
				return settle("protocol", 12)
			}
		case <-adoption.C:
			return settle("adoption_timeout", 12)
		case <-exited:
			return settle("exited", 0)
		case <-captureFailed:
			return settle("capture", 12)
		case <-ctx.Done():
			return settle("signal", 12)
		case <-poll.C:
			gone, err := commandParentGone(parentFD)
			if gone || err != nil {
				return settle("parent_lost", 12)
			}
		}
	}
}

func pipeCommandMain(args []string, capturing bool) int {
	var capture *commandCapture
	if capturing {
		if len(args) < 2 {
			return 2
		}
		limit, err := strconv.ParseInt(args[0], 10, 64)
		if err != nil || limit < 1 || limit > maximumCommandCaptureBytes || strconv.FormatInt(limit, 10) != args[0] {
			return 2
		}
		capture = newCommandCapture(limit)
		args = args[1:]
	}
	if len(args) == 0 {
		return 2
	}
	parent, err := CaptureProcess(os.Getppid())
	if err != nil || parent.PID < 2 {
		return 12
	}
	// The inherited channel may be blocking. Register it with Go's poller so
	// a vanished/non-reading parent cannot block settlement acknowledgement.
	var stat unix.Stat_t
	if unix.Fstat(3, &stat) != nil || (stat.Mode&unix.S_IFMT != unix.S_IFIFO && stat.Mode&unix.S_IFMT != unix.S_IFSOCK) || unix.SetNonblock(3, true) != nil {
		return 12
	}
	events := os.NewFile(3, "owned-command-events")
	if events == nil {
		return 12
	}
	defer events.Close()
	command := exec.Command(args[0], args[1:]...)
	command.Stdout, command.Stderr = os.Stdout, os.Stderr
	if capture != nil {
		command.Stdout = commandCaptureWriter{capture, os.Stdout}
		command.Stderr = commandCaptureWriter{capture, os.Stderr}
	}
	return servePipeCommand(command, os.Stdin, events, parent, 8*time.Second, capture)
}

// A control frame must end in LF. EOF is always loss of the control channel.
func splitCommandControl(data []byte, eof bool) (int, []byte, error) {
	for i, value := range data {
		if value == '\n' {
			return i + 1, data[:i], nil
		}
	}
	if eof && len(data) != 0 {
		return 0, nil, io.ErrUnexpectedEOF
	}
	return 0, nil, nil
}
