//go:build linux

package harnesslocal

import (
	"bufio"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"syscall"
)

// Stream bounds pending memory while keeping unpublished records private.
// Close publishes once; Abort removes the owned staging inode. Neither follows
// a path replacement or overwrites a previously published artifact.
type Stream struct {
	parent, file               *os.File
	buffer                     *bufio.Writer
	directory, temporary, leaf string
	closed                     bool
}

// StagingPath is the exact current-epoch event stream consumed by graph units.
// It is private until Close publishes the immutable terminal path.
func (s *Stream) StagingPath() string { return filepath.Join(s.directory, s.temporary) }

func NewStream(name string, bufferBytes int) (*Stream, error) {
	parent, err := PrivateDirectory(filepath.Dir(name))
	if err != nil {
		return nil, err
	}
	var nonce [16]byte
	if _, err = rand.Read(nonce[:]); err != nil {
		parent.Close()
		return nil, err
	}
	temporary := ".publishing-" + hex.EncodeToString(nonce[:])
	fd, err := syscall.Openat(int(parent.Fd()), temporary, syscall.O_WRONLY|syscall.O_CREAT|syscall.O_EXCL|syscall.O_NOFOLLOW|syscall.O_CLOEXEC, 0600)
	if err != nil {
		parent.Close()
		return nil, err
	}
	file := os.NewFile(uintptr(fd), temporary)
	return &Stream{parent: parent, file: file, buffer: bufio.NewWriterSize(file, bufferBytes), directory: filepath.Dir(name), temporary: temporary, leaf: filepath.Base(name)}, nil
}
func (s *Stream) Write(data []byte) error {
	if s.closed {
		return ErrUnsafe
	}
	_, err := s.buffer.Write(data)
	return err
}
func (s *Stream) Flush() error {
	if s.closed {
		return ErrUnsafe
	}
	return s.buffer.Flush()
}
func (s *Stream) Abort() error {
	if s.closed {
		return nil
	}
	s.closed = true
	prefix := fmt.Sprintf("/proc/self/fd/%d/", s.parent.Fd())
	return errors.Join(s.file.Close(), os.Remove(prefix+s.temporary), s.parent.Close())
}
func (s *Stream) Close() (err error) {
	if s.closed {
		return ErrUnsafe
	}
	defer func() {
		if err != nil {
			err = errors.Join(err, s.Abort())
		}
	}()
	if err = s.buffer.Flush(); err != nil {
		return err
	}
	if err = s.file.Sync(); err != nil {
		return err
	}
	if err = CheckDirectory(s.parent, s.directory); err != nil {
		return err
	}
	prefix := fmt.Sprintf("/proc/self/fd/%d/", s.parent.Fd())
	if err = os.Link(prefix+s.temporary, prefix+s.leaf); err != nil {
		return err
	}
	if err = os.Remove(prefix + s.temporary); err != nil {
		return err
	}
	// Once linked, a later failure is explicit incomplete publication. The index
	// must not claim a sample digest unless all completion steps succeeded.
	s.closed = true
	return errors.Join(s.file.Close(), s.parent.Sync(), CheckDirectory(s.parent, s.directory), s.parent.Close())
}
