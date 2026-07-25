//go:build !linux

package rootedfs

import (
	"context"
)

type Root struct{}

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

func (*Root) MakePrivateDir(Reference) error {
	return ErrUnsupportedPlatform
}

func (*Root) ReadRegular(Reference, int64) ([]byte, Metadata, error) {
	return nil, Metadata{}, ErrUnsupportedPlatform
}

func (*Root) OpenRegular(Reference) (ReadSeekCloser, Metadata, error) {
	return nil, Metadata{}, ErrUnsupportedPlatform
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
