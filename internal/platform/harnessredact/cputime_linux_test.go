package harnessredact

import (
	"runtime"
	"testing"
	"time"

	"golang.org/x/sys/unix"
)

// threadCPU runs f on one locked OS thread and returns the CPU time that
// thread spent. Unlike wall time, it does not grow when other work on the host
// competes for cores.
func threadCPU(t *testing.T, f func()) time.Duration {
	t.Helper()
	runtime.LockOSThread()
	defer runtime.UnlockOSThread()
	before := threadUsage(t)
	f()
	return threadUsage(t) - before
}

func threadUsage(t *testing.T) time.Duration {
	t.Helper()
	var usage unix.Rusage
	if err := unix.Getrusage(unix.RUSAGE_THREAD, &usage); err != nil {
		t.Fatalf("read thread CPU time: %v", err)
	}
	return time.Duration(usage.Utime.Nano() + usage.Stime.Nano())
}
