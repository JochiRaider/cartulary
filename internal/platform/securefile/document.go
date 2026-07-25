package securefile

import (
	"errors"
	"fmt"
	"io/fs"
	"time"
)

var ErrUnsupportedPlatform = errors.New("secure file reads are unsupported on this platform")

type FailureKind string

const (
	FailureInvalidPath  FailureKind = "invalid_path"
	FailureUnavailable  FailureKind = "unavailable"
	FailureUnsafeObject FailureKind = "unsafe_object"
	FailureTooLarge     FailureKind = "too_large"
	FailureRead         FailureKind = "read_failed"
	FailureChanged      FailureKind = "identity_changed"
	FailureUnsupported  FailureKind = "unsupported_platform"
)

type Document struct {
	bytes    []byte
	size     int64
	mode     fs.FileMode
	modified time.Time
}

func (document Document) Bytes() []byte {
	return append([]byte(nil), document.bytes...)
}

func (document Document) Size() int64 {
	return document.size
}

func (document Document) Mode() fs.FileMode {
	return document.mode
}

func (document Document) ModTime() time.Time {
	return document.modified
}

type Error struct {
	Operation string
	Kind      FailureKind
	Reason    string
	cause     error
}

func (err *Error) Error() string {
	return fmt.Sprintf("securefile %s: %s", err.Operation, err.Reason)
}

func (err *Error) Unwrap() error {
	return err.cause
}

func secureFileError(operation string, kind FailureKind, reason string, cause error) error {
	return &Error{Operation: operation, Kind: kind, Reason: reason, cause: cause}
}
