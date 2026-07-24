package stagedobjects

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"time"
)

var semanticIDPattern = regexp.MustCompile(`^[a-z][a-z0-9_]{0,127}$`)

type IDGenerator func() (string, error)

type ServiceOptions struct {
	Repository   Repository
	Bytes        ByteStore
	Now          func() time.Time
	NewID        IDGenerator
	MaximumBytes int
	FatalSink    func(error)
}

type Service struct {
	repository   Repository
	bytes        ByteStore
	now          func() time.Time
	newID        IDGenerator
	maximumBytes int
	fatalSink    func(error)
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
		repository:   options.Repository,
		bytes:        options.Bytes,
		now:          options.Now,
		newID:        options.NewID,
		maximumBytes: options.MaximumBytes,
		fatalSink:    options.FatalSink,
	}, nil
}

func (s *Service) Allocate(ctx context.Context, operationID, profileID string, payload []byte) (Reference, error) {
	if s == nil ||
		operationID == "" ||
		!semanticIDPattern.MatchString(profileID) ||
		payload == nil ||
		len(payload) > s.maximumBytes {
		return Reference{}, ErrScopeDenied
	}
	stagingID, err := s.newID()
	if err != nil || stagingID == "" {
		return Reference{}, ErrDependency
	}
	sum := sha256.Sum256(payload)
	now := s.now().UTC().Truncate(time.Microsecond)
	allocation := Allocation{
		StagingID:       stagingID,
		OperationID:     operationID,
		OwnerProfileID:  profileID,
		StorageIdentity: fmt.Sprintf(".cartulary/staged/%s/%s", profileID, stagingID),
		ByteSize:        int64(len(payload)),
		SHA256:          hex.EncodeToString(sum[:]),
		StagedAt:        now,
		ExpiresAt:       now.Add(StagingLifetime),
	}
	if err := s.repository.Allocate(ctx, allocation); err != nil {
		return Reference{}, err
	}
	outcome, uploadErr := s.bytes.Put(ctx, allocation.StorageIdentity, append([]byte(nil), payload...))
	if outcome != ByteWriteSuccess || uploadErr != nil {
		return Reference{}, s.abandonAfterFailure(ctx, stagingID, classifyWriteFailure(outcome, uploadErr))
	}
	if err := s.repository.MarkReady(ctx, stagingID, s.now().UTC().Truncate(time.Microsecond)); err != nil {
		return Reference{}, s.abandonAfterFailure(ctx, stagingID, err)
	}
	return Reference{StagingID: stagingID}, nil
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
