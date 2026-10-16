//go:build linux && !amd64

package renderprocess

import (
	"context"
	"golang.org/x/sys/unix"
)

// No unconfined fallback on an architecture without a reviewed syscall ABI.
func ExecNetworkChild(int, []string, []string) error { return ErrStart }
func superviseNetwork(_ context.Context, channel int, _ uint16) ([]NetworkObservation, error) {
	_ = unix.Close(channel)
	return nil, ErrNetworkObservation
}
