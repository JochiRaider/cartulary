//go:build linux

// Package harnesssession owns shared private invocation state and retained scans.
// It has no dependency on graph scheduling, fixture policy or browser execution.
package harnesssession

import (
	"errors"
	"os"

	"github.com/JochiRaider/cartulary/internal/testutil/harnesslocal"
	"golang.org/x/sys/unix"
)

var ErrUnresolved = errors.New("runtime retains unresolved ownership")

func (r *Runtime) mutable() error {
	var state unix.Stat_t
	err := unix.Fstatat(int(r.Directory.Fd()), "finalizing", &state, unix.AT_SYMLINK_NOFOLLOW)
	if errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return harnesslocal.ErrUnsafe
}

func (r *Runtime) lock() (func(), error) {
	r.mu.Lock()
	if r.closed {
		r.mu.Unlock()
		return nil, harnesslocal.ErrUnsafe
	}
	if err := unix.Flock(int(r.Directory.Fd()), unix.LOCK_EX); err != nil {
		r.mu.Unlock()
		return nil, err
	}
	unlock := func() {
		_ = unix.Flock(int(r.Directory.Fd()), unix.LOCK_UN)
		r.mu.Unlock()
	}
	if err := harnesslocal.CheckDirectory(r.Directory, r.Root); err != nil {
		unlock()
		return nil, err
	}
	if _, err := ReadBinding(r.Repository, r.RunRoot, r.Root, r.RunID, r.LeaseID); err != nil {
		unlock()
		return nil, err
	}
	return unlock, nil
}
