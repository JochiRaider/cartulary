//go:build linux

package harnesssession

import (
	"io"
	"sync"
)

const maximumCommandCaptureBytes int64 = 256 * 1024 * 1024

// commandCapture accounts both streams before writing. The workload receives
// pipes, never the backing file descriptors. After failure, draining discards
// bytes until the existing process owner settles every writer and joins Wait.
type commandCapture struct {
	mu        sync.Mutex
	remaining int64
	failure   string
	failed    chan struct{}
}

type commandCaptureWriter struct {
	capture *commandCapture
	output  io.Writer
}

func newCommandCapture(limit int64) *commandCapture {
	return &commandCapture{remaining: limit, failed: make(chan struct{})}
}

func (capture *commandCapture) Failure() string {
	if capture == nil {
		return ""
	}
	capture.mu.Lock()
	defer capture.mu.Unlock()
	return capture.failure
}

func (writer commandCaptureWriter) Write(data []byte) (int, error) {
	capture := writer.capture
	capture.mu.Lock()
	defer capture.mu.Unlock()
	if capture.failure != "" {
		return len(data), nil
	}
	count := min(int64(len(data)), capture.remaining)
	written, err := writer.output.Write(data[:count])
	capture.remaining -= int64(written)
	if err != nil || int64(written) != count {
		capture.failure = "io"
	} else if count < int64(len(data)) {
		capture.failure = "limit"
	}
	if capture.failure != "" {
		close(capture.failed)
	}
	// Failure is an explicit capture fact, separate from the command's Wait
	// status and physical settlement. An io.Copy error must not conflate them.
	return len(data), nil
}
