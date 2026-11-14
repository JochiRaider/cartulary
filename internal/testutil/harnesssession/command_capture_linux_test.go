//go:build linux

package harnesssession

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"testing"
	"time"
)

type shortCaptureWriter struct{}

func (shortCaptureWriter) Write(data []byte) (int, error) { return len(data) - 1, nil }

func testCommandCapture(t *testing.T) {
	for _, c := range []struct {
		name, script, failure string
		bytes                 int
	}{
		{"exact", "head -c 1024 /dev/zero", "", 1024},
		{"overflow", "head -c 1025 /dev/zero", "limit", 1024},
		{"simultaneous", "head -c 1024 /dev/zero & head -c 1024 /dev/zero >&2; wait", "limit", 1024},
		{"both_exact", "head -c 512 /dev/zero & head -c 512 /dev/zero >&2; wait", "", 1024},
	} {
		t.Run(c.name, func(t *testing.T) {
			var stdout, stderr bytes.Buffer
			capture := newCommandCapture(1024)
			command := exec.Command("sh", "-c", c.script)
			command.Stdout, command.Stderr = commandCaptureWriter{capture, &stdout}, commandCaptureWriter{capture, &stderr}
			fixture := newPipeCommandFixture(t, command, 8*time.Second, capture)
			proof := fixture.ready(t)
			fmt.Fprintln(fixture.control, "adopt")
			event := fixture.event(t)
			status := 0
			if c.failure != "" {
				status = 12
			}
			awaitCommandControl(t, fixture.done, status)
			if event.Event != "settled" || event.CaptureFailure != c.failure || event.Exit == nil || capture.Failure() != c.failure {
				t.Fatal("capture settlement facts", event)
			}
			if c.failure == "" && (event.Cause != "exited" || *event.Exit != 0) || c.failure != "" && event.Cause != "capture" {
				t.Fatal("capture failure lost or successful command failed", event)
			}
			if stdout.Len()+stderr.Len() != c.bytes {
				t.Fatal("aggregate capture size", stdout.Len(), stderr.Len())
			}
			if _, err := readProcess(proof.PID); !processGone(err) {
				t.Fatal("capture leader not reaped", err)
			}
		})
	}
	t.Run("write_errors_are_capture_failures_not_settlement_failures", func(t *testing.T) {
		full, err := os.OpenFile("/dev/full", os.O_WRONLY, 0)
		if err != nil {
			t.Fatal(err)
		}
		defer full.Close()
		for _, output := range []io.Writer{full, shortCaptureWriter{}} {
			capture := newCommandCapture(1024)
			command := exec.Command("sh", "-c", "printf output")
			command.Stdout = commandCaptureWriter{capture, output}
			fixture := newPipeCommandFixture(t, command, 8*time.Second, capture)
			fixture.ready(t)
			fmt.Fprintln(fixture.control, "adopt")
			event := fixture.event(t)
			awaitCommandControl(t, fixture.done, 12)
			if event.Event != "settled" || event.Cause != "capture" || event.CaptureFailure != "io" || event.Diagnostic != nil {
				t.Fatal("capture write error misclassified", event)
			}
		}
	})
	t.Run("overflow_retains_registration_anchor_and_settles_descendant_writer", func(t *testing.T) {
		file := filepath.Join(t.TempDir(), "descendant")
		var stdout, stderr bytes.Buffer
		capture := newCommandCapture(1024)
		command := exec.Command("sh", "-c", `sh -c 'while :; do printf output; done' & echo $! > "$1"; wait`, "fixture", file)
		command.Stdout, command.Stderr = commandCaptureWriter{capture, &stdout}, commandCaptureWriter{capture, &stderr}
		fixture := newPipeCommandFixture(t, command, 8*time.Second, capture)
		proof := fixture.ready(t)
		select {
		case <-capture.failed:
		case <-time.After(5 * time.Second):
			t.Fatal("writer did not overflow")
		}
		actual, err := CaptureProcess(proof.PID)
		if err != nil || actual != proof {
			t.Fatal("overflow reaped registration anchor", err)
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
		if err != nil || child.PGID != proof.PGID {
			t.Fatal("descendant ownership", err)
		}
		fmt.Fprintln(fixture.control, "adopt")
		event := fixture.event(t)
		awaitCommandControl(t, fixture.done, 12)
		if event.Event != "settled" || event.CaptureFailure != "limit" {
			t.Fatal("overflow result", event)
		}
		if stdout.Len()+stderr.Len() != 1024 {
			t.Fatal("overflow exceeded storage bound")
		}
		child, err = readProcess(pid)
		if !processGone(err) && (err != nil || child.state != "Z" && child.state != "X") {
			t.Fatal("descendant writer survived", err)
		}
		if _, err := readProcess(proof.PID); !processGone(err) {
			t.Fatal("leader not reaped", err)
		}
	})
	t.Run("private_dispatch_rejects_invalid_limits_before_launch", func(t *testing.T) {
		for _, limit := range []string{"0", "-1", "01", "+1", "268435457", "9223372036854775808", "bad"} {
			marker := filepath.Join(t.TempDir(), "must-not-launch")
			if pipeCommandMain([]string{limit, "touch", marker}, true) != 2 {
				t.Fatal("invalid limit admitted", limit)
			}
			if _, err := os.Stat(marker); !errors.Is(err, os.ErrNotExist) {
				t.Fatal("invalid limit launched", err)
			}
		}
	})
}
