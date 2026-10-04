package reference_data

import (
	"context"
	"errors"
	"fmt"
	"io"
	pathpkg "path"
	"strings"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

var ErrInvalidReferencePackStorageReference = errors.New("reference pack: invalid storage reference")

// ErrArtifactUnavailable means the addressed immutable file is definitively
// absent or has been replaced by a prohibited filesystem object. Permission,
// device, transport and root-capability failures remain operational errors.
var ErrArtifactUnavailable = errors.New("reference pack: authoritative artifact unavailable")

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

// VerificationStorage is the streaming boundary for canonical verification.
// The application supplies a fresh workspace per execution attempt. Publication
// is a separate operation; workspace bytes are never consumer-visible.
type VerificationStorage interface {
	OpenStaged(context.Context, StagingRef) (ContainerReader, int64, error)
	OpenPublished(context.Context, StorageRef) (ContainerReader, int64, error)
	NewWorkspace(context.Context) (VerificationWorkspace, error)
}

// ArtifactStorage admits and retains bytes without whole-container allocation.
// Published objects are immutable; their references become usable only when
// the owner commits them. PublishStream returns a namespace lease which must
// remain open until registration commits or aborts. A caller must not delete an
// uncertain publication; its transaction must also hold the retention guard.
type ArtifactStorage interface {
	VerificationStorage
	StageStream(context.Context, io.Reader, int64) (StagingRef, string, int64, error)
	PublishStream(context.Context, string, int64, io.Reader) (StorageRef, io.Closer, error)
	RemoveStaged(StagingRef) error
	RemovePublished(StorageRef) error
}

type ContainerReader interface {
	io.ReaderAt
	io.Closer
}

// Archive decoders can collapse a failed read into a malformed-stream error.
// Preserve an underlying operational failure independently of decoder errors.
// Only one verification goroutine owns the reader.
type checkedContainerReader struct {
	ContainerReader
	fault error
}

func (r *checkedContainerReader) ReadAt(p []byte, offset int64) (int, error) {
	n, err := r.ContainerReader.ReadAt(p, offset)
	if err != nil && !errors.Is(err, io.EOF) && !errors.Is(err, io.ErrUnexpectedEOF) && r.fault == nil {
		r.fault = err
	}
	return n, err
}

type VerificationWorkspace interface {
	Write(context.Context, string, func(io.Writer) error) error
	Open(context.Context, string) (io.ReadCloser, error)
	Remove(context.Context, string) error
	Close() error
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

// CollectionStorage holds the publication namespace exclusively for the entire
// callback, including its database retention guard and all physical deletions.
// References are opaque and visited in bounded storage batches.
type CollectionStorage interface {
	ArtifactStorage
	WithPublishedCollection(context.Context, func(PublishedObjects) error) error
}
type PublishedObjects interface {
	Visit(context.Context, func(StorageRef) error) error
	Remove(StorageRef) error
}
