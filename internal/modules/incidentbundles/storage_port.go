package incidentbundles

import (
	"context"
	"errors"
	"fmt"
	pathpkg "path"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

var ErrInvalidStorageReference = errors.New("incident bundle: invalid storage reference")

// BundleStorageRef is an opaque logical reference to a published incident
// bundle. It never contains a host filesystem root.
type BundleStorageRef struct {
	value string
}

// BundleStagingRef is an opaque logical reference to a staged incident bundle.
// It never contains a host filesystem root.
type BundleStagingRef struct {
	value string
}

func ParseBundleStorageRef(raw string) (BundleStorageRef, error) {
	if err := validateStorageReference(raw); err != nil {
		return BundleStorageRef{}, err
	}
	return BundleStorageRef{value: raw}, nil
}

func ParseBundleStagingRef(raw string) (BundleStagingRef, error) {
	if err := validateStorageReference(raw); err != nil {
		return BundleStagingRef{}, err
	}
	return BundleStagingRef{value: raw}, nil
}

func (reference BundleStorageRef) String() string {
	return reference.value
}

func (reference BundleStagingRef) String() string {
	return reference.value
}

// BundleStorage is the incident-bundle owner's root-free persistence boundary.
// Application assembly provides an implementation backed by an admitted
// deployment capability.
type BundleStorage interface {
	Stage(context.Context, string, []byte) (BundleStagingRef, error)
	Publish(context.Context, string, []byte) (BundleStorageRef, error)
	ReadStaged(BundleStagingRef, int64) ([]byte, error)
	RemoveStaged(BundleStagingRef) error
	RemovePublished(BundleStorageRef) error
}

func validateStorageReference(raw string) error {
	switch {
	case raw == "":
		return fmt.Errorf("%w: empty", ErrInvalidStorageReference)
	case !utf8.ValidString(raw):
		return fmt.Errorf("%w: invalid UTF-8", ErrInvalidStorageReference)
	case strings.IndexByte(raw, 0) >= 0:
		return fmt.Errorf("%w: NUL", ErrInvalidStorageReference)
	case strings.Contains(raw, `\`):
		return fmt.Errorf("%w: backslash", ErrInvalidStorageReference)
	case strings.HasPrefix(raw, "/") || pathpkg.IsAbs(raw):
		return fmt.Errorf("%w: absolute", ErrInvalidStorageReference)
	case norm.NFC.String(raw) != raw:
		return fmt.Errorf("%w: non-canonical Unicode", ErrInvalidStorageReference)
	case pathpkg.Clean(raw) != raw:
		return fmt.Errorf("%w: non-canonical path", ErrInvalidStorageReference)
	}
	for _, component := range strings.Split(raw, "/") {
		if component == "" || component == "." || component == ".." {
			return fmt.Errorf("%w: invalid component", ErrInvalidStorageReference)
		}
	}
	return nil
}
