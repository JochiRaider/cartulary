//go:build linux

package startupproof

import (
	"errors"
	"os"
	"syscall"
)

// OpenControl makes an inherited descriptor close-interruptible by Go's poller.
// Without nonblocking mode an inherited pipe read can outlive File.Close.
func OpenControl(fd int) (*os.File, error) {
	if fd < 3 {
		return nil, errors.New("invalid startup control capability")
	}
	if err := syscall.SetNonblock(fd, true); err != nil {
		return nil, err
	}
	syscall.CloseOnExec(fd)
	return os.NewFile(uintptr(fd), "startup-control"), nil
}
