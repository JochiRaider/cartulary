//go:build linux

package securefile

import (
	"errors"
	"io"
	"io/fs"
	"os"
	pathpkg "path"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/sys/unix"
	"golang.org/x/text/unicode/norm"
)

var errTooLarge = errors.New("secure file exceeds maximum size")

const (
	secureDirectoryFlags = unix.O_RDONLY | unix.O_DIRECTORY | unix.O_CLOEXEC | unix.O_NOFOLLOW
	secureRegularFlags   = unix.O_RDONLY | unix.O_CLOEXEC | unix.O_NOFOLLOW | unix.O_NONBLOCK
)

func Read(absolutePath string, maxBytes int64) (Document, error) {
	if maxBytes <= 0 {
		return Document{}, secureFileError("read", FailureInvalidPath, "maximum size must be positive", fs.ErrInvalid)
	}
	if err := validateAbsolutePath(absolutePath); err != nil {
		return Document{}, secureFileError("read", FailureInvalidPath, "path is not a canonical absolute POSIX path", err)
	}
	fd, err := openAbsoluteNoFollow(absolutePath)
	if err != nil {
		kind := FailureUnavailable
		if errors.Is(err, unix.ELOOP) || errors.Is(err, unix.ENOTDIR) {
			kind = FailureUnsafeObject
		}
		return Document{}, secureFileError("read", kind, "file is unavailable or unsafe", err)
	}
	file := os.NewFile(uintptr(fd), "secure-manifest")
	defer file.Close()

	before, err := boundedRegularStat(fd, maxBytes)
	if err != nil {
		kind := FailureUnsafeObject
		if errors.Is(err, errTooLarge) {
			kind = FailureTooLarge
		}
		return Document{}, secureFileError("read", kind, "object is not a bounded regular file", err)
	}
	payload, err := io.ReadAll(io.LimitReader(file, maxBytes+1))
	if err != nil {
		return Document{}, secureFileError("read", FailureRead, "file read failed", err)
	}
	if int64(len(payload)) > maxBytes {
		return Document{}, secureFileError("read", FailureTooLarge, "file exceeds maximum size", errTooLarge)
	}
	after, err := boundedRegularStat(fd, maxBytes)
	if err != nil || before.Dev != after.Dev || before.Ino != after.Ino ||
		before.Size != after.Size || before.Mtim != after.Mtim || before.Ctim != after.Ctim {
		return Document{}, secureFileError("read", FailureChanged, "file identity changed during read", err)
	}
	return Document{
		bytes:    append([]byte(nil), payload...),
		size:     after.Size,
		mode:     fs.FileMode(after.Mode & 0o777),
		modified: time.Unix(after.Mtim.Sec, after.Mtim.Nsec),
	}, nil
}

func openAbsoluteNoFollow(absolutePath string) (int, error) {
	current, err := unix.Open("/", secureDirectoryFlags, 0)
	if err != nil {
		return -1, err
	}
	components := strings.Split(strings.TrimPrefix(absolutePath, "/"), "/")
	for index, component := range components {
		flags := secureDirectoryFlags
		if index == len(components)-1 {
			flags = secureRegularFlags
		}
		next, openErr := unix.Openat(current, component, flags, 0)
		unix.Close(current)
		if openErr != nil {
			return -1, openErr
		}
		current = next
	}
	return current, nil
}

func validateAbsolutePath(absolutePath string) error {
	switch {
	case absolutePath == "" || absolutePath == "/" || !strings.HasPrefix(absolutePath, "/"):
		return fs.ErrInvalid
	case !utf8.ValidString(absolutePath) || strings.IndexByte(absolutePath, 0) >= 0:
		return fs.ErrInvalid
	case strings.Contains(absolutePath, `\`):
		return fs.ErrInvalid
	case norm.NFC.String(absolutePath) != absolutePath:
		return fs.ErrInvalid
	case pathpkg.Clean(absolutePath) != absolutePath:
		return fs.ErrInvalid
	default:
		return nil
	}
}

func boundedRegularStat(fd int, maxBytes int64) (unix.Stat_t, error) {
	var stat unix.Stat_t
	if err := unix.Fstat(fd, &stat); err != nil {
		return unix.Stat_t{}, err
	}
	if stat.Mode&unix.S_IFMT != unix.S_IFREG || stat.Size < 0 {
		return unix.Stat_t{}, fs.ErrInvalid
	}
	if stat.Size > maxBytes {
		return unix.Stat_t{}, errTooLarge
	}
	return stat, nil
}
