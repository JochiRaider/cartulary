//go:build linux

package rootedfs

import (
	"context"
	"errors"
	"io"
	"io/fs"
	"os"
	"sync"

	"golang.org/x/sys/unix"
)

// Stage creates bounded private scratch space on the admitted filesystem. The
// unnamed inode cannot be linked or published, and is reclaimed on close or
// process exit. Only a successful callback can expose its read capability.
func (root *Root) Stage(ctx context.Context, maxBytes int64, write WriteFunc) (result ReadSeekCloser, resultErr error) {
	root.mu.RLock()
	defer root.mu.RUnlock()
	if maxBytes <= 0 || write == nil {
		return nil, operationError("stage", Reference{}, "invalid staging request", fs.ErrInvalid)
	}
	if err := root.checkReady("stage"); err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	fd, err := unix.Openat(root.fd, ".", unix.O_TMPFILE|unix.O_EXCL|unix.O_RDWR|unix.O_CLOEXEC, 0o600)
	if err != nil {
		return nil, operationError("stage", Reference{}, "private staging is unavailable", err)
	}
	file := os.NewFile(uintptr(fd), "private-stage")
	defer func() {
		if result == nil {
			resultErr = errors.Join(resultErr, file.Close())
		}
	}()
	writer := &stageWriter{file: file, ctx: ctx, remaining: maxBytes}
	writeErr := write(writer)
	writer.mu.Lock()
	writer.sealed = true
	writeErr = errors.Join(writeErr, writer.failure)
	writer.mu.Unlock()
	if writeErr != nil {
		return nil, operationError("stage", Reference{}, "staging callback failed", writeErr)
	}
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	if err := root.checkRootIdentity(); err != nil {
		return nil, err
	}
	if _, err := file.Seek(0, io.SeekStart); err != nil {
		return nil, operationError("stage", Reference{}, "staging rewind failed", err)
	}
	return &stageReader{file: file}, nil
}

// The callback cannot retain a usable writer after Stage returns, even when it
// ignores a failed write. Neither capability exposes the underlying descriptor.
type stageWriter struct {
	mu        sync.Mutex
	file      *os.File
	ctx       context.Context
	remaining int64
	sealed    bool
	failure   error
}

func (writer *stageWriter) Write(body []byte) (int, error) {
	writer.mu.Lock()
	defer writer.mu.Unlock()
	if writer.sealed {
		return 0, fs.ErrClosed
	}
	if writer.failure != nil {
		return 0, writer.failure
	}
	if err := writer.ctx.Err(); err != nil {
		writer.failure = err
		return 0, err
	}
	if int64(len(body)) > writer.remaining {
		writer.failure = fs.ErrInvalid
		return 0, writer.failure
	}
	n, err := writer.file.Write(body)
	writer.remaining -= int64(n)
	if err == nil && n != len(body) {
		err = io.ErrShortWrite
	}
	writer.failure = err
	return n, err
}

type stageReader struct {
	file *os.File
	once sync.Once
	err  error
}

func (reader *stageReader) Read(body []byte) (int, error) { return reader.file.Read(body) }
func (reader *stageReader) ReadAt(body []byte, offset int64) (int, error) {
	return reader.file.ReadAt(body, offset)
}
func (reader *stageReader) Seek(offset int64, whence int) (int64, error) {
	return reader.file.Seek(offset, whence)
}
func (reader *stageReader) Close() error {
	reader.once.Do(func() { reader.err = reader.file.Close() })
	return reader.err
}
