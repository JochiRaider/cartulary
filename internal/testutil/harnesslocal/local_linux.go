//go:build linux

package harnesslocal

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

// Traverse every component through an open descriptor. No ancestor symlinks,
// mutations, or reliance on an earlier path-based check are permitted.
func OpenDirectory(name string, private bool) (*os.File, error) {
	absolute, err := filepath.Abs(name)
	if err != nil {
		return nil, err
	}
	fd, err := syscall.Open("/", syscall.O_RDONLY|syscall.O_DIRECTORY|syscall.O_NOFOLLOW|syscall.O_CLOEXEC, 0)
	if err != nil {
		return nil, err
	}
	for _, part := range strings.Split(strings.TrimPrefix(absolute, "/"), "/") {
		if part == "" {
			continue
		}
		next, e := syscall.Openat(fd, part, syscall.O_RDONLY|syscall.O_DIRECTORY|syscall.O_NOFOLLOW|syscall.O_CLOEXEC, 0)
		syscall.Close(fd)
		if e != nil {
			return nil, e
		}
		fd = next
	}
	f := os.NewFile(uintptr(fd), absolute)
	if private {
		var info syscall.Stat_t
		if err := syscall.Fstat(fd, &info); err != nil {
			f.Close()
			return nil, err
		}
		if info.Uid != uint32(os.Getuid()) || info.Mode&0777 != 0700 {
			f.Close()
			return nil, ErrUnsafe
		}
	}
	return f, nil
}
func ReadFile(name string, maximum int64, mutable bool) ([]byte, error) {
	parent, err := OpenDirectory(filepath.Dir(name), false)
	if err != nil {
		return nil, rejectRead("opening", "directory_unavailable", err)
	}
	defer parent.Close()
	return ReadFileAt(parent, filepath.Base(name), maximum, mutable)
}

// ReadFileAt retains the caller's directory identity across a transaction.
func ReadFileAt(parent *os.File, name string, maximum int64, mutable bool) ([]byte, error) {
	return readFileAt(parent, name, maximum, mutable, nil)
}

// Per-call barriers let package tests coordinate real filesystem operations;
// production callers cannot install global hooks or select a different reader.
type readCheckpoint uint8

const (
	readBeforeOpen readCheckpoint = iota
	readAfterOpen
	readAfterMetadata
	readDuringBytes
	readBeforeFinalMetadata
)

func readFileAt(parent *os.File, name string, maximum int64, mutable bool, checkpoint func(readCheckpoint)) ([]byte, error) {
	at := func(stage readCheckpoint) {
		if checkpoint != nil {
			checkpoint(stage)
		}
	}
	if name == "" || name == "." || name == ".." || filepath.Base(name) != name {
		return nil, rejectRead("opening", "invalid_name", nil)
	}
	at(readBeforeOpen)
	fd, err := syscall.Openat(int(parent.Fd()), name, syscall.O_RDONLY|syscall.O_NOFOLLOW|syscall.O_NONBLOCK|syscall.O_CLOEXEC, 0)
	if err != nil {
		return nil, rejectRead("opening", "file_unavailable", err)
	}
	f := os.NewFile(uintptr(fd), name)
	defer f.Close()
	at(readAfterOpen)
	var before, after syscall.Stat_t
	if err := syscall.Fstat(fd, &before); err != nil {
		return nil, rejectRead("initial_metadata", "stat_unavailable", err)
	}
	switch {
	case before.Mode&syscall.S_IFMT != syscall.S_IFREG:
		return nil, rejectRead("initial_metadata", "not_regular", nil)
	case before.Uid != uint32(os.Getuid()):
		return nil, rejectRead("initial_metadata", "foreign_owner", nil)
	case before.Mode&0777 != 0600:
		return nil, rejectRead("initial_metadata", "unsafe_mode", nil)
	case before.Size < 0 || before.Size > maximum:
		return nil, rejectRead("initial_metadata", "size_limit", nil)
	case before.Nlink != 1 && !(mutable && before.Nlink == 0):
		return nil, rejectRead("initial_metadata", "unsafe_links", nil)
	}
	at(readAfterMetadata)
	// Keep the same single-descriptor bounded read. The optional test reader
	// barrier fires after its first read, including when that read consumes EOF.
	reader := io.Reader(f)
	if checkpoint != nil {
		reader = &checkpointReader{Reader: f, after: func() { at(readDuringBytes) }}
	}
	data, err := io.ReadAll(io.LimitReader(reader, before.Size+1))
	if err != nil {
		return nil, rejectRead("reading", "read_failed", err)
	}
	at(readBeforeFinalMetadata)
	if err := syscall.Fstat(fd, &after); err != nil {
		return nil, rejectRead("final_metadata", "stat_unavailable", err)
	}
	if err := validateReadTransition(before, after, len(data), mutable); err != nil {
		return nil, err
	}
	return data, nil
}

func validateReadTransition(before, after syscall.Stat_t, length int, mutable bool) error {
	switch {
	case int64(length) != before.Size:
		return rejectRead("final_metadata", "length_changed", nil)
	case before.Size != after.Size:
		return rejectRead("final_metadata", "size_changed", nil)
	case before.Mtim != after.Mtim:
		return rejectRead("final_metadata", "mtime_changed", nil)
	case before.Ino != after.Ino || before.Dev != after.Dev:
		return rejectRead("final_metadata", "identity_changed", nil)
	case before.Mode != after.Mode:
		return rejectRead("final_metadata", "mode_changed", nil)
	case before.Uid != after.Uid:
		return rejectRead("final_metadata", "owner_changed", nil)
	}
	if mutable {
		// Atomic publication changes namespace metadata independently of content.
		// In particular, ctime may change while fstat still reports one link.
		// An already-opened snapshot may lose its name, but cannot gain links.
		if after.Nlink > 1 || after.Nlink > before.Nlink {
			return rejectRead("final_metadata", "links_changed", nil)
		}
		return nil
	}
	// Immutable evidence has no publication transition after it is opened.
	if after.Nlink != 1 {
		return rejectRead("final_metadata", "links_changed", nil)
	}
	if before.Ctim != after.Ctim {
		return rejectRead("final_metadata", "ctime_changed", nil)
	}
	return nil
}

type checkpointReader struct {
	io.Reader
	after func()
}

func (r *checkpointReader) Read(p []byte) (int, error) {
	n, err := r.Reader.Read(p)
	if r.after != nil {
		after := r.after
		r.after = nil
		after()
	}
	return n, err
}

// PublishNew installs one owner-only immutable file, without following an
// existing destination or replacing any scanned bytes. The open directory is
// kept until publication and its path identity is checked before returning.
func PublishNew(name string, data []byte) (result error) {
	parent, err := OpenDirectory(filepath.Dir(name), true)
	if err != nil {
		return err
	}
	defer parent.Close()
	var nonce [16]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return err
	}
	temporary := ".publishing-" + hex.EncodeToString(nonce[:])
	prefix := fmt.Sprintf("/proc/self/fd/%d/", parent.Fd())
	fd, err := syscall.Openat(int(parent.Fd()), temporary, syscall.O_WRONLY|syscall.O_CREAT|syscall.O_EXCL|syscall.O_NOFOLLOW|syscall.O_CLOEXEC, 0600)
	if err != nil {
		return err
	}
	f := os.NewFile(uintptr(fd), temporary)
	defer func() {
		if cleanupErr := os.Remove(prefix + temporary); cleanupErr != nil && !errors.Is(cleanupErr, os.ErrNotExist) {
			result = errors.Join(result, cleanupErr)
		}
	}()
	if _, err = f.Write(data); err != nil {
		f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	if err = os.Link(prefix+temporary, prefix+filepath.Base(name)); err != nil {
		return err
	}
	if err = os.Remove(prefix + temporary); err != nil {
		return err
	}
	if err = parent.Sync(); err != nil {
		return err
	}
	current, err := OpenDirectory(filepath.Dir(name), true)
	if err != nil {
		return err
	}
	defer current.Close()
	before, err := parent.Stat()
	if err != nil {
		return err
	}
	after, err := current.Stat()
	if err != nil {
		return err
	}
	if !os.SameFile(before, after) {
		return ErrUnsafe
	}
	return nil
}

// PrivateDirectory creates missing components, retaining no-follow traversal.
func PrivateDirectory(name string) (*os.File, error) {
	absolute, err := filepath.Abs(name)
	if err != nil {
		return nil, err
	}
	fd, err := syscall.Open("/", syscall.O_RDONLY|syscall.O_DIRECTORY|syscall.O_NOFOLLOW|syscall.O_CLOEXEC, 0)
	if err != nil {
		return nil, err
	}
	for _, part := range strings.Split(strings.TrimPrefix(absolute, "/"), "/") {
		if part == "" {
			continue
		}
		err := syscall.Mkdirat(fd, part, 0700)
		if err != nil && err != syscall.EEXIST {
			syscall.Close(fd)
			return nil, err
		}
		next, e := syscall.Openat(fd, part, syscall.O_RDONLY|syscall.O_DIRECTORY|syscall.O_NOFOLLOW|syscall.O_CLOEXEC, 0)
		syscall.Close(fd)
		if e != nil {
			return nil, e
		}
		fd = next
	}
	file := os.NewFile(uintptr(fd), absolute)
	var stat syscall.Stat_t
	if err := syscall.Fstat(fd, &stat); err != nil {
		file.Close()
		return nil, err
	}
	if stat.Uid != uint32(os.Getuid()) || stat.Mode&0777 != 0700 {
		file.Close()
		return nil, ErrUnsafe
	}
	return file, nil
}

// ReplaceFile changes a private mutable state file only after validating the
// old inode. Writers serialize through their owning protocol's mutex.
func ReplaceFile(name string, data []byte) error {
	parent, err := OpenDirectory(filepath.Dir(name), true)
	if err != nil {
		return err
	}
	defer parent.Close()
	if err := ReplaceFileAt(parent, filepath.Base(name), data); err != nil {
		return err
	}
	return CheckDirectory(parent, filepath.Dir(name))
}
func CheckDirectory(parent *os.File, name string) error {
	current, err := OpenDirectory(name, true)
	if err != nil {
		return err
	}
	defer current.Close()
	before, err := parent.Stat()
	if err != nil {
		return err
	}
	after, err := current.Stat()
	if err != nil {
		return err
	}
	if !os.SameFile(before, after) {
		return ErrUnsafe
	}
	return nil
}
func ReplaceFileAt(parent *os.File, name string, data []byte) (result error) {
	if name == "" || name == "." || name == ".." || filepath.Base(name) != name {
		return ErrUnsafe
	}
	if _, err := ReadFileAt(parent, name, 4*1024*1024, false); err != nil && !errors.Is(err, os.ErrNotExist) {
		return err
	}
	var nonce [16]byte
	if _, err := rand.Read(nonce[:]); err != nil {
		return err
	}
	temporary := ".publishing-" + hex.EncodeToString(nonce[:])
	prefix := fmt.Sprintf("/proc/self/fd/%d/", parent.Fd())
	fd, err := syscall.Openat(int(parent.Fd()), temporary, syscall.O_WRONLY|syscall.O_CREAT|syscall.O_EXCL|syscall.O_NOFOLLOW|syscall.O_CLOEXEC, 0600)
	if err != nil {
		return err
	}
	f := os.NewFile(uintptr(fd), temporary)
	defer func() {
		if cleanupErr := os.Remove(prefix + temporary); cleanupErr != nil && !errors.Is(cleanupErr, os.ErrNotExist) {
			result = errors.Join(result, cleanupErr)
		}
	}()
	if _, err = f.Write(data); err != nil {
		f.Close()
		return err
	}
	if err = f.Sync(); err != nil {
		f.Close()
		return err
	}
	if err = f.Close(); err != nil {
		return err
	}
	if err = os.Rename(prefix+temporary, prefix+name); err != nil {
		return err
	}
	return parent.Sync()
}
