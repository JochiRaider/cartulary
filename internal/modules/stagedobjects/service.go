package stagedobjects

import (
	"bytes"
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"regexp"
	"sort"
	"time"
)

var semanticIDPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,127}$`)

type IDGenerator func() (string, error)

type ServiceOptions struct {
	Repository       Repository
	Bytes            ByteStore
	Now              func() time.Time
	NewID            IDGenerator
	MaximumBytes     int
	FatalSink        func(error)
	ResourceIdentity ResourceIdentity
}

type Service struct {
	repository       Repository
	bytes            ByteStore
	now              func() time.Time
	newID            IDGenerator
	maximumBytes     int
	fatalSink        func(error)
	resourceIdentity ResourceIdentity
}

func NewService(options ServiceOptions) (*Service, error) {
	if options.Repository == nil || options.Bytes == nil || options.FatalSink == nil {
		return nil, errors.New("staged-object service dependencies are incomplete")
	}
	if options.Now == nil {
		options.Now = func() time.Time { return time.Now().UTC() }
	}
	if options.NewID == nil {
		options.NewID = randomStagingID
	}
	if options.MaximumBytes <= 0 {
		options.MaximumBytes = DefaultMaximumBytes
	}
	return &Service{
		repository:       options.Repository,
		bytes:            options.Bytes,
		now:              options.Now,
		newID:            options.NewID,
		maximumBytes:     options.MaximumBytes,
		fatalSink:        options.FatalSink,
		resourceIdentity: options.ResourceIdentity,
	}, nil
}

// Allocate is the bounded control-payload convenience. Bulk owners use
// AllocateStream with their admitted member ceiling; both share one lifecycle.
func (s *Service) Allocate(ctx context.Context, operationID, profileID string, payload []byte) (Reference, error) {
	if s == nil || payload == nil || len(payload) > s.maximumBytes {
		return Reference{}, ErrScopeDenied
	}
	sealed, err := s.AllocateStream(ctx, operationID, profileID, bytes.NewReader(payload), int64(s.maximumBytes))
	return sealed.Reference, err
}

func (s *Service) AllocateStream(ctx context.Context, operationID, profileID string, source io.Reader, maximumBytes int64) (SealedReference, error) {
	sealed, _, err := s.allocateStream(ctx, operationID, profileID, nil, source, maximumBytes)
	return sealed, err
}

// AllocateResourceStream reserves an immutable owner identity before writing any
// bytes. The trusted owner resolver derives a different key for every attempt.
func (s *Service) AllocateResourceStream(ctx context.Context, operationID, profileID string, resource ResourceBinding, source io.Reader, maximumBytes int64) (SealedReference, string, error) {
	if !resource.Valid() {
		return SealedReference{}, "", ErrScopeDenied
	}
	return s.allocateStream(ctx, operationID, profileID, &resource, source, maximumBytes)
}

func (s *Service) allocateStream(ctx context.Context, operationID, profileID string, resource *ResourceBinding, source io.Reader, maximumBytes int64) (SealedReference, string, error) {
	if s == nil || operationID == "" || !semanticIDPattern.MatchString(profileID) || source == nil || maximumBytes < 0 {
		return SealedReference{}, "", ErrScopeDenied
	}
	stagingID, err := s.newID()
	if err != nil || stagingID == "" {
		return SealedReference{}, "", ErrDependency
	}
	identity := fmt.Sprintf(".cartulary/staged/%s/%s", profileID, stagingID)
	if resource != nil {
		if s.resourceIdentity == nil {
			return SealedReference{}, "", ErrScopeDenied
		}
		identity, err = s.resourceIdentity(profileID, *resource, stagingID)
		if err != nil || identity == "" {
			return SealedReference{}, "", ErrScopeDenied
		}
	}
	now := s.now().UTC().Truncate(time.Microsecond)
	allocation := Allocation{StagingID: stagingID, OperationID: operationID, OwnerProfileID: profileID,
		StorageIdentity: identity, Resource: resource, StagedAt: now, ExpiresAt: now.Add(StagingLifetime)}
	if err := s.repository.Allocate(ctx, allocation); err != nil {
		return SealedReference{}, "", err
	}
	measured := &measuredReader{ctx: ctx, source: source, remaining: maximumBytes, hash: sha256.New()}
	outcome, err := s.bytes.Put(ctx, allocation.StorageIdentity, measured)
	if measured.failure != nil {
		return SealedReference{}, "", s.abandonAfterFailure(ctx, stagingID, measured.failure)
	}
	if outcome != ByteWriteSuccess || err != nil {
		return SealedReference{}, "", s.abandonAfterFailure(ctx, stagingID, classifyWriteFailure(outcome, err))
	}
	// A backend must consume the complete reader. Never seal an acknowledged
	// short write or a source whose limit/error was ignored by the adapter.
	if !measured.complete || measured.failure != nil {
		return SealedReference{}, "", s.abandonAfterFailure(ctx, stagingID, NewFailure(FailureIntegrity, "incomplete_stream", measured.failure))
	}
	sealed := SealedReference{Reference: Reference{StagingID: stagingID}, ContentBinding: ContentBinding{ByteSize: measured.count, SHA256: hex.EncodeToString(measured.hash.Sum(nil))}}
	if err := s.repository.Seal(ctx, operationID, profileID, sealed, s.now().UTC().Truncate(time.Microsecond)); err != nil {
		return SealedReference{}, "", s.abandonAfterFailure(ctx, stagingID, err)
	}
	return sealed, identity, nil
}

// OpenSealed is private preparation access. It requires the allocating operation
// and verifies the complete bytes against the immutable seal as they are read.
func (s *Service) OpenSealed(ctx context.Context, operationID, profileID string, reference SealedReference) (io.ReadCloser, error) {
	if s == nil || operationID == "" || !semanticIDPattern.MatchString(profileID) || reference.StagingID == "" {
		return nil, ErrScopeDenied
	}
	object, err := s.repository.ReadSealed(ctx, operationID, profileID, reference.Reference, s.now().UTC().Truncate(time.Microsecond))
	if err != nil {
		return nil, err
	}
	if object.ContentBinding != reference.ContentBinding || !validSHA256(object.SHA256) || object.ByteSize < 0 {
		return nil, ErrIntegrity
	}
	source, err := s.bytes.Open(ctx, object.StorageIdentity)
	if err != nil {
		return nil, err
	}
	return &verifiedReadCloser{source: source, reader: measuredReader{ctx: ctx, source: source, remaining: object.ByteSize, hash: sha256.New()}, expected: object.ContentBinding}, nil
}

func (s *Service) Abandon(ctx context.Context, reference Reference) error {
	if s == nil || reference.StagingID == "" {
		return ErrScopeDenied
	}
	return s.repository.Abandon(ctx, reference.StagingID, s.now().UTC().Truncate(time.Microsecond))
}

func (s *Service) abandonAfterFailure(ctx context.Context, stagingID string, cause error) error {
	abandonErr := s.repository.Abandon(context.WithoutCancel(ctx), stagingID, s.now().UTC().Truncate(time.Microsecond))
	if abandonErr != nil {
		fatalErr := fmt.Errorf("%w: abandon %s after %v: %v", ErrIntegrity, stagingID, cause, abandonErr)
		s.fatalSink(fatalErr)
		return &FatalIntegrityError{Cause: fatalErr}
	}
	return cause
}

func classifyWriteFailure(outcome ByteWriteOutcome, cause error) error {
	switch outcome {
	case ByteWriteDependency:
		return NewFailure(FailureDependency, "object_store_unavailable", cause)
	case ByteWriteIntegrity:
		return NewFailure(FailureIntegrity, "object_write_integrity", cause)
	case ByteWriteIndeterminate:
		return NewFailure(FailureRetryable, "object_write_indeterminate", cause)
	default:
		if cause != nil {
			return NewFailure(FailureRetryable, "object_write_failed", cause)
		}
		return ErrInvalidLifecycle
	}
}

type Allocator interface {
	Allocate(context.Context, string, string, []byte) (Reference, error)
	Abandon(context.Context, Reference) error
}

type Scope struct {
	operationID string
	profileID   string
	allocator   Allocator
	references  []Reference
	closed      bool
}

func NewScope(operationID, profileID string, allocator Allocator) (*Scope, error) {
	if operationID == "" || !semanticIDPattern.MatchString(profileID) || allocator == nil {
		return nil, ErrScopeDenied
	}
	return &Scope{operationID: operationID, profileID: profileID, allocator: allocator}, nil
}

func (s *Scope) Allocate(ctx context.Context, operationID string, payload []byte) (string, error) {
	if s == nil || s.closed || operationID != s.operationID {
		return "", ErrScopeDenied
	}
	reference, err := s.allocator.Allocate(ctx, s.operationID, s.profileID, payload)
	if err != nil {
		return "", err
	}
	if reference.StagingID == "" {
		return "", ErrInvalidLifecycle
	}
	s.references = append(s.references, reference)
	return reference.String(), nil
}

func (s *Scope) Refs() []string {
	if s == nil {
		return nil
	}
	references := make([]string, len(s.references))
	for index, reference := range s.references {
		references[index] = reference.String()
	}
	sort.Strings(references)
	return references
}

func (s *Scope) Abandon(ctx context.Context) error {
	if s == nil || s.closed {
		return nil
	}
	s.closed = true
	var joined error
	for _, reference := range s.references {
		joined = errors.Join(joined, s.allocator.Abandon(ctx, reference))
	}
	return joined
}

func (s *Scope) Transfer(operationID string) (Transfer, error) {
	if s == nil || s.closed || operationID != s.operationID {
		return Transfer{}, ErrScopeDenied
	}
	s.closed = true
	return Transfer{
		operationID: s.operationID,
		profileID:   s.profileID,
		references:  s.Refs(),
	}, nil
}

func randomStagingID() (string, error) {
	var value [16]byte
	if _, err := rand.Read(value[:]); err != nil {
		return "", err
	}
	return hex.EncodeToString(value[:]), nil
}

func validSHA256(value string) bool {
	if len(value) != sha256.Size*2 {
		return false
	}
	decoded, err := hex.DecodeString(value)
	return err == nil && len(decoded) == sha256.Size
}
