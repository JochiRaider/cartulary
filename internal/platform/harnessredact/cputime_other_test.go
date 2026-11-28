//go:build !linux

package harnessredact

import (
	"testing"
	"time"
)

// threadCPU is measured on Linux, the reference profile, only.
func threadCPU(t *testing.T, _ func()) time.Duration {
	t.Helper()
	t.Skip("thread CPU time is measured on Linux only")
	return 0
}
