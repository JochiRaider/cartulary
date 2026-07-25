package reference_data

import (
	"context"
	"errors"
	"fmt"
	pathpkg "path"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

var ErrInvalidReferencePackStorageReference = errors.New("reference pack: invalid storage reference")

// StorageRef is an opaque logical reference to a published Reference Pack.
// It deliberately carries no host filesystem root.
type StorageRef struct {
	value string
}

// StagingRef is an opaque logical reference to a staged Reference Pack.
// It deliberately carries no host filesystem root.
type StagingRef struct {
	value string
}

func ParseStorageRef(raw string) (StorageRef, error) {
	if err := validateStorageReference(raw); err != nil {
		return StorageRef{}, err
	}
	return StorageRef{value: raw}, nil
}

func ParseStagingRef(raw string) (StagingRef, error) {
	if err := validateStorageReference(raw); err != nil {
		return StagingRef{}, err
	}
	return StagingRef{value: raw}, nil
}

func (reference StorageRef) String() string {
	return reference.value
}

func (reference StagingRef) String() string {
	return reference.value
}

// Storage is the Reference Pack owner's root-free persistence boundary.
// Application assembly supplies a rooted implementation.
type Storage interface {
	Stage(context.Context, string, []byte) (StagingRef, error)
	Publish(context.Context, string, []byte) (StorageRef, error)
	ReadStaged(StagingRef, int64) ([]byte, error)
	ReadPublished(StorageRef, int64) ([]byte, error)
	RemoveStaged(StagingRef) error
	RemovePublished(StorageRef) error
}

func validateStorageReference(raw string) error {
	switch {
	case raw == "":
		return fmt.Errorf("%w: empty", ErrInvalidReferencePackStorageReference)
	case !utf8.ValidString(raw):
		return fmt.Errorf("%w: invalid UTF-8", ErrInvalidReferencePackStorageReference)
	case strings.IndexByte(raw, 0) >= 0:
		return fmt.Errorf("%w: NUL", ErrInvalidReferencePackStorageReference)
	case strings.Contains(raw, `\`):
		return fmt.Errorf("%w: backslash", ErrInvalidReferencePackStorageReference)
	case strings.HasPrefix(raw, "/") || pathpkg.IsAbs(raw):
		return fmt.Errorf("%w: absolute", ErrInvalidReferencePackStorageReference)
	case norm.NFC.String(raw) != raw:
		return fmt.Errorf("%w: non-canonical Unicode", ErrInvalidReferencePackStorageReference)
	case pathpkg.Clean(raw) != raw:
		return fmt.Errorf("%w: non-canonical path", ErrInvalidReferencePackStorageReference)
	}
	for _, component := range strings.Split(raw, "/") {
		if component == "" || component == "." || component == ".." {
			return fmt.Errorf("%w: invalid component", ErrInvalidReferencePackStorageReference)
		}
	}
	return nil
}
