package rootedfs

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	pathpkg "path"
	"sort"
	"strings"
	"time"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

var (
	ErrClosed              = errors.New("rooted filesystem is closed")
	ErrInvalidReference    = errors.New("invalid rooted filesystem reference")
	ErrReferenceCollision  = errors.New("rooted filesystem reference collision")
	ErrRootIdentityChanged = errors.New("rooted filesystem identity changed")
	ErrUnsupportedPlatform = errors.New("rooted filesystem is unsupported on this platform")
)

// Reference is a canonical, relative POSIX storage reference. Its zero value is
// invalid. It deliberately contains no host filesystem root.
type Reference struct {
	value string
}

func ParseReference(raw string) (Reference, error) {
	if err := validateReference(raw); err != nil {
		return Reference{}, err
	}
	return Reference{value: raw}, nil
}

func MustParseReference(raw string) Reference {
	reference, err := ParseReference(raw)
	if err != nil {
		panic(err)
	}
	return reference
}

func (reference Reference) String() string {
	return reference.value
}

// ValidateReferenceSet validates a set of file-like references and rejects
// duplicate, Unicode-normalization, and file/directory-prefix collisions.
func ValidateReferenceSet(rawReferences []string) ([]Reference, error) {
	normalizedOwners := make(map[string]string, len(rawReferences))
	for _, raw := range rawReferences {
		if !utf8.ValidString(raw) {
			continue
		}
		normalized := norm.NFC.String(raw)
		if owner, ok := normalizedOwners[normalized]; ok && owner != raw {
			return nil, fmt.Errorf("%w: Unicode normalization", ErrReferenceCollision)
		}
		normalizedOwners[normalized] = raw
	}

	references := make([]Reference, 0, len(rawReferences))
	for _, raw := range rawReferences {
		reference, err := ParseReference(raw)
		if err != nil {
			return nil, err
		}
		references = append(references, reference)
	}
	sort.Slice(references, func(i, j int) bool {
		return references[i].value < references[j].value
	})
	for index, reference := range references {
		if index > 0 && references[index-1].value == reference.value {
			return nil, fmt.Errorf("%w: duplicate", ErrReferenceCollision)
		}
		if index > 0 && strings.HasPrefix(reference.value, references[index-1].value+"/") {
			return nil, fmt.Errorf("%w: file/directory prefix", ErrReferenceCollision)
		}
		for prior := index - 2; prior >= 0; prior-- {
			if !strings.HasPrefix(reference.value, references[prior].value) {
				break
			}
			if strings.HasPrefix(reference.value, references[prior].value+"/") {
				return nil, fmt.Errorf("%w: file/directory prefix", ErrReferenceCollision)
			}
		}
	}
	return references, nil
}

func validateReference(raw string) error {
	switch {
	case raw == "":
		return fmt.Errorf("%w: empty", ErrInvalidReference)
	case !utf8.ValidString(raw):
		return fmt.Errorf("%w: invalid UTF-8", ErrInvalidReference)
	case strings.IndexByte(raw, 0) >= 0:
		return fmt.Errorf("%w: NUL", ErrInvalidReference)
	case strings.Contains(raw, `\`):
		return fmt.Errorf("%w: backslash", ErrInvalidReference)
	case strings.HasPrefix(raw, "/") || pathpkg.IsAbs(raw):
		return fmt.Errorf("%w: absolute", ErrInvalidReference)
	case norm.NFC.String(raw) != raw:
		return fmt.Errorf("%w: non-canonical Unicode", ErrInvalidReference)
	case pathpkg.Clean(raw) != raw:
		return fmt.Errorf("%w: non-canonical path", ErrInvalidReference)
	}
	for _, component := range strings.Split(raw, "/") {
		if component == "" || component == "." || component == ".." {
			return fmt.Errorf("%w: invalid component", ErrInvalidReference)
		}
	}
	return nil
}

type Metadata struct {
	Size    int64
	Mode    fs.FileMode
	ModTime time.Time
}

type ReadSeekCloser interface {
	io.Reader
	io.Seeker
	io.Closer
}

type RegularEntry struct {
	Reference Reference
	Metadata  Metadata
}

type WriteFunc func(io.Writer) error

type OperationError struct {
	Operation string
	Reference string
	Reason    string
	cause     error
}

func (err *OperationError) Error() string {
	if err.Reference == "" {
		return fmt.Sprintf("rootedfs %s: %s", err.Operation, err.Reason)
	}
	return fmt.Sprintf("rootedfs %s %q: %s", err.Operation, err.Reference, err.Reason)
}

func (err *OperationError) Unwrap() error {
	return err.cause
}

func operationError(operation string, reference Reference, reason string, cause error) error {
	return &OperationError{
		Operation: operation,
		Reference: reference.value,
		Reason:    reason,
		cause:     cause,
	}
}
