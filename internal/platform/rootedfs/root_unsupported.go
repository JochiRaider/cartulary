//go:build !linux

package rootedfs

import (
	"context"
	"io"
)

func (*Root) TryLockDirectory(Reference) (io.Closer, bool, error) {
	return nil, false, ErrUnsupportedPlatform
}
func (*Root) LockDirectory(context.Context, Reference) (io.Closer, error) {
	return nil, ErrUnsupportedPlatform
}
func (*Root) VisitDirectory(context.Context, Reference, func(DirectoryEntry) error) error {
	return ErrUnsupportedPlatform
}

type Root struct{}

func (*Root) isEmpty() (bool, error) { return false, ErrUnsupportedPlatform }

func Open(string) (*Root, error) {
	return nil, operationError("open-root", Reference{}, "platform is unsupported", ErrUnsupportedPlatform)
}

func OpenOrCreate(string) (*Root, error) {
	return nil, operationError("open-root", Reference{}, "platform is unsupported", ErrUnsupportedPlatform)
}

func (*Root) Close() error {
	return nil
}

func (*Root) Check() error {
	return ErrUnsupportedPlatform
}

func (*Root) Exists(Reference) (bool, error) {
	return false, ErrUnsupportedPlatform
}

func (*Root) MakePrivateDir(Reference) error {
	return ErrUnsupportedPlatform
}

func (*Root) ReadRegular(Reference, int64) ([]byte, Metadata, error) {
	return nil, Metadata{}, ErrUnsupportedPlatform
}

func (*Root) OpenRegular(Reference) (ReadSeekCloser, Metadata, error) {
	return nil, Metadata{}, ErrUnsupportedPlatform
}

func (*Root) Stage(context.Context, int64, WriteFunc) (ReadSeekCloser, error) {
	return nil, ErrUnsupportedPlatform
}

func (*Root) ListRegular() ([]RegularEntry, error) {
	return nil, ErrUnsupportedPlatform
}

func (*Root) CreateExclusive(context.Context, Reference, WriteFunc) error {
	return ErrUnsupportedPlatform
}

func (*Root) AtomicReplace(context.Context, Reference, WriteFunc) error {
	return ErrUnsupportedPlatform
}

func (*Root) RenameExclusive(Reference, Reference) error {
	return ErrUnsupportedPlatform
}

func (*Root) RemoveRegular(Reference) error {
	return ErrUnsupportedPlatform
}

func (*Root) RemoveEmptyDir(Reference) error {
	return ErrUnsupportedPlatform
}

func (*Root) PublishSiblingExclusive(Reference) error { return ErrUnsupportedPlatform }
