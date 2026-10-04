//go:build linux

package rootedfs

import (
	"context"
	"errors"
	"io"
	"os"
	"sync"
	"time"

	"golang.org/x/sys/unix"
)

type directoryLease struct {
	fd   int
	once sync.Once
	err  error
}

func (l *directoryLease) Close() error {
	l.once.Do(func() { l.err = unix.Close(l.fd) })
	return l.err
}

// TryLockDirectory acquires an exclusive advisory lease on the admitted
// directory inode. Independent processes contend on the same inode; process
// exit releases the lease. The caller must serialize directory replacement
// with its parent lease, rather than removing/recreating a live lock target.
func (root *Root) TryLockDirectory(reference Reference) (io.Closer, bool, error) {
	return root.tryLockDirectory(reference, unix.LOCK_EX)
}

func (root *Root) tryLockDirectory(reference Reference, mode int) (io.Closer, bool, error) {
	if err := validateReference(reference.value); err != nil {
		return nil, false, err
	}
	root.mu.RLock()
	defer root.mu.RUnlock()
	if err := root.checkReady("lock-directory"); err != nil {
		return nil, false, err
	}
	chain, err := root.openDirectoryChain(reference, false)
	if err != nil {
		return nil, false, operationError("lock-directory", reference, "directory traversal failed", err)
	}
	defer chain.close()
	fd, err := duplicateDescriptor(chain.lastFD())
	if err != nil {
		return nil, false, err
	}
	if err := unix.Flock(fd, mode|unix.LOCK_NB); err != nil {
		_ = unix.Close(fd)
		if errors.Is(err, unix.EWOULDBLOCK) {
			return nil, false, nil
		}
		return nil, false, operationError("lock-directory", reference, "directory lease unavailable", err)
	}
	if err := root.validateChain(chain); err != nil {
		_ = unix.Close(fd)
		return nil, false, err
	}
	if err := root.checkRootIdentity(); err != nil {
		_ = unix.Close(fd)
		return nil, false, err
	}
	return &directoryLease{fd: fd}, true, nil
}

// LockDirectory waits with cancellation. It does not hold a Root mutex while
// waiting, and never infers ownership from a PID or a wall-clock age.
func (root *Root) LockDirectory(ctx context.Context, reference Reference) (io.Closer, error) {
	return root.lockDirectory(ctx, reference, unix.LOCK_EX)
}

// LockSharedDirectory protects a stable namespace from exclusive collection
// while allowing concurrent publishers. The caller must close the lease.
func (root *Root) LockSharedDirectory(ctx context.Context, reference Reference) (io.Closer, error) {
	return root.lockDirectory(ctx, reference, unix.LOCK_SH)
}

func (root *Root) lockDirectory(ctx context.Context, reference Reference, mode int) (io.Closer, error) {
	for {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		lease, acquired, err := root.tryLockDirectory(reference, mode)
		if err != nil || acquired {
			return lease, err
		}
		timer := time.NewTimer(25 * time.Millisecond)
		select {
		case <-ctx.Done():
			timer.Stop()
			return nil, ctx.Err()
		case <-timer.C:
		}
	}
}

// VisitDirectory streams immediate children in bounded batches. Enumeration
// order is unspecified. The visitor may use other Root operations; no Root
// mutex is held during its callback. Callers requiring a stable namespace
// must hold the relevant directory lease across enumeration and mutation.
func (root *Root) VisitDirectory(ctx context.Context, reference Reference, visit func(DirectoryEntry) error) (resultErr error) {
	if err := validateReference(reference.value); err != nil {
		return err
	}
	if visit == nil {
		return operationError("visit-directory", reference, "visitor required", nil)
	}
	root.mu.RLock()
	if err := root.checkReady("visit-directory"); err != nil {
		root.mu.RUnlock()
		return err
	}
	chain, err := root.openDirectoryChain(reference, false)
	root.mu.RUnlock()
	if err != nil {
		return operationError("visit-directory", reference, "directory traversal failed", err)
	}
	defer chain.close()
	fd, err := unix.Openat(chain.lastFD(), ".", directoryOpenFlags, 0)
	if err != nil {
		return err
	}
	directory := os.NewFile(uintptr(fd), "rooted-directory")
	defer func() { resultErr = errors.Join(resultErr, directory.Close()) }()
	check := func() error {
		root.mu.RLock()
		defer root.mu.RUnlock()
		if err := root.checkReady("visit-directory"); err != nil {
			return err
		}
		return root.validateChain(chain)
	}
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		if err := check(); err != nil {
			return err
		}
		entries, readErr := directory.ReadDir(128)
		if readErr != nil && !errors.Is(readErr, io.EOF) {
			return readErr
		}
		for _, entry := range entries {
			if err := ctx.Err(); err != nil {
				return err
			}
			child, err := ParseReference(reference.String() + "/" + entry.Name())
			if err != nil {
				return err
			}
			var stat unix.Stat_t
			if err := unix.Fstatat(chain.lastFD(), entry.Name(), &stat, unix.AT_SYMLINK_NOFOLLOW); err != nil {
				return err
			}
			isDirectory := stat.Mode&unix.S_IFMT == unix.S_IFDIR
			if !isDirectory {
				if err := requireAllowedRegularStat(stat, true); err != nil {
					return err
				}
			}
			if err := check(); err != nil {
				return err
			}
			if err := visit(DirectoryEntry{Reference: child, Directory: isDirectory}); err != nil {
				return err
			}
		}
		if errors.Is(readErr, io.EOF) {
			return check()
		}
	}
}
