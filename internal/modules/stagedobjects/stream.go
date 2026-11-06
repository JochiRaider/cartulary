package stagedobjects

import (
	"context"
	"encoding/hex"
	"hash"
	"io"
)

// measuredReader never reads past the admitted ceiling except for one byte
// needed to prove excess. Memory use is bounded by the downstream read buffer.
type measuredReader struct {
	ctx              context.Context
	source           io.Reader
	remaining, count int64
	hash             hash.Hash
	complete         bool
	failure          error
}

func (r *measuredReader) Read(p []byte) (int, error) {
	if len(p) == 0 {
		return 0, nil
	}
	if r.failure != nil {
		return 0, r.failure
	}
	if err := r.ctx.Err(); err != nil {
		r.failure = err
		return 0, err
	}
	if r.complete {
		return 0, io.EOF
	}
	if r.remaining == 0 {
		var probe [1]byte
		n, err := r.source.Read(probe[:])
		if n > 0 {
			r.failure = NewFailure(FailureIntegrity, "member_bytes_exceeded", nil)
			return 0, r.failure
		}
		if err == io.EOF {
			r.complete = true
		} else if err != nil {
			r.failure = err
		}
		return 0, err
	}
	if int64(len(p)) > r.remaining {
		p = p[:r.remaining]
	}
	n, err := r.source.Read(p)
	if n < 0 || n > len(p) {
		r.failure = ErrIntegrity
		return 0, r.failure
	}
	r.count += int64(n)
	r.remaining -= int64(n)
	if n > 0 {
		_, _ = r.hash.Write(p[:n])
	}
	if err == io.EOF {
		r.complete = true
	} else if err != nil {
		r.failure = err
	}
	return n, err
}

type verifiedReadCloser struct {
	source   io.ReadCloser
	reader   measuredReader
	expected ContentBinding
}

func (r *verifiedReadCloser) Read(p []byte) (int, error) {
	n, err := r.reader.Read(p)
	if err == io.EOF && (r.reader.count != r.expected.ByteSize || hex.EncodeToString(r.reader.hash.Sum(nil)) != r.expected.SHA256) {
		r.reader.failure = ErrIntegrity
		return n, ErrIntegrity
	}
	return n, err
}
func (r *verifiedReadCloser) Close() error { return r.source.Close() }

type StreamAllocator interface {
	Allocator
	AllocateStream(context.Context, string, string, io.Reader, int64) (SealedReference, error)
	OpenSealed(context.Context, string, string, SealedReference) (io.ReadCloser, error)
}

func (s *Scope) AllocateStream(ctx context.Context, operationID string, source io.Reader, maximumBytes int64) (SealedReference, error) {
	if s == nil || s.closed || operationID != s.operationID {
		return SealedReference{}, ErrScopeDenied
	}
	allocator, ok := s.allocator.(StreamAllocator)
	if !ok {
		return SealedReference{}, ErrDependency
	}
	sealed, err := allocator.AllocateStream(ctx, s.operationID, s.profileID, source, maximumBytes)
	if err != nil {
		return SealedReference{}, err
	}
	if sealed.StagingID == "" {
		return SealedReference{}, ErrInvalidLifecycle
	}
	s.references = append(s.references, sealed.Reference)
	return sealed, nil
}

func (s *Scope) OpenSealed(ctx context.Context, operationID string, reference SealedReference) (io.ReadCloser, error) {
	if s == nil || s.closed || operationID != s.operationID {
		return nil, ErrScopeDenied
	}
	owned := false
	for _, ref := range s.references {
		if ref == reference.Reference {
			owned = true
			break
		}
	}
	if !owned {
		return nil, ErrScopeDenied
	}
	allocator, ok := s.allocator.(StreamAllocator)
	if !ok {
		return nil, ErrDependency
	}
	return allocator.OpenSealed(ctx, s.operationID, s.profileID, reference)
}

// ResourceAllocator is restricted to source owners with an admitted physical
// identity resolver. Participant code receives narrower logical output scopes.
type ResourceAllocator interface {
	Allocator
	AllocateResourceStream(context.Context, string, string, ResourceBinding, io.Reader, int64) (SealedReference, string, error)
}

func (s *Scope) AllocateResourceStream(ctx context.Context, operationID string, resource ResourceBinding, source io.Reader, maximumBytes int64) (SealedReference, string, error) {
	if s == nil || s.closed || operationID != s.operationID {
		return SealedReference{}, "", ErrScopeDenied
	}
	allocator, ok := s.allocator.(ResourceAllocator)
	if !ok {
		return SealedReference{}, "", ErrDependency
	}
	sealed, identity, err := allocator.AllocateResourceStream(ctx, s.operationID, s.profileID, resource, source, maximumBytes)
	if err != nil {
		return SealedReference{}, "", err
	}
	if sealed.StagingID == "" || identity == "" {
		return SealedReference{}, "", ErrInvalidLifecycle
	}
	s.references = append(s.references, sealed.Reference)
	return sealed, identity, nil
}
