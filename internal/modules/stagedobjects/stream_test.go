package stagedobjects

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"testing"
	"time"
)

type zeroStream struct {
	remaining   int64
	largestRead int
}

func (s *zeroStream) Read(p []byte) (int, error) {
	if len(p) > s.largestRead {
		s.largestRead = len(p)
	}
	if s.remaining == 0 {
		return 0, io.EOF
	}
	n := min(int64(len(p)), s.remaining)
	clear(p[:n])
	s.remaining -= n
	return int(n), nil
}

type discardStore struct {
	byteStoreFake
	readError error
	stopEarly bool
}

func (s *discardStore) Put(_ context.Context, _ string, source io.Reader) (ByteWriteOutcome, error) {
	if s.stopEarly {
		return ByteWriteSuccess, nil
	}
	_, err := io.CopyBuffer(io.Discard, struct{ io.Reader }{source}, make([]byte, 32*1024))
	if err != nil {
		return ByteWriteIntegrity, err
	}
	if s.readError != nil {
		return ByteWriteIndeterminate, s.readError
	}
	return ByteWriteSuccess, nil
}

func TestStream_Unit_BoundedBulkSealAndFailures(t *testing.T) {
	const size = 65 * 1024 * 1024
	now := time.Date(2026, 11, 5, 12, 0, 0, 0, time.UTC)
	t.Run("bulk above control ceiling remains incremental", func(t *testing.T) {
		repo := &serviceRepositoryFake{}
		service := newTestService(t, repo, &discardStore{}, now, func() (string, error) { return "bulk", nil }, nil)
		source := &zeroStream{remaining: size}
		sealed, err := service.AllocateStream(t.Context(), "operation", "incident_portability", source, size)
		if err != nil {
			t.Fatal(err)
		}
		expected := sha256.New()
		_, _ = io.Copy(expected, io.LimitReader(&zeroStream{remaining: size}, size))
		if sealed.ByteSize != size || sealed.SHA256 != hex.EncodeToString(expected.Sum(nil)) || source.largestRead > 32*1024 || len(repo.ready) != 1 {
			t.Fatalf("invalid streamed seal: %#v largest read %d", sealed, source.largestRead)
		}
	})
	for _, tc := range []struct {
		name   string
		source io.Reader
		limit  int64
		store  *discardStore
	}{
		{"excess", bytes.NewReader([]byte("excess")), 5, &discardStore{}},
		{"source error", io.MultiReader(bytes.NewReader([]byte("part")), errorReader{}), 10, &discardStore{}},
		{"short acknowledgement", bytes.NewReader([]byte("whole")), 10, &discardStore{stopEarly: true}},
		{"lost acknowledgement", bytes.NewReader([]byte("whole")), 10, &discardStore{readError: errors.New("lost acknowledgement")}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			repo := &serviceRepositoryFake{}
			service := newTestService(t, repo, tc.store, now, func() (string, error) { return "failed", nil }, nil)
			if _, err := service.AllocateStream(t.Context(), "operation", "incident_portability", tc.source, tc.limit); err == nil {
				t.Fatal("invalid stream sealed")
			}
			if len(repo.ready) != 0 || len(repo.abandoned) != 1 {
				t.Fatalf("incomplete staging lifecycle: %#v", repo)
			}
		})
	}
}

type errorReader struct{}

func (errorReader) Read([]byte) (int, error) { return 0, io.ErrUnexpectedEOF }

func TestStream_Unit_PrivateReadsVerifyOperationAndContent(t *testing.T) {
	repo := &serviceRepositoryFake{}
	store := &byteStoreFake{putOutcome: ByteWriteSuccess}
	service := newTestService(t, repo, store, time.Now(), func() (string, error) { return "read", nil }, nil)
	sealed, err := service.AllocateStream(t.Context(), "operation", "incident_portability", bytes.NewReader([]byte("original")), 64)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := service.OpenSealed(t.Context(), "other", "incident_portability", sealed); !errors.Is(err, ErrScopeDenied) {
		t.Fatalf("cross-operation read: %v", err)
	}
	for _, payload := range [][]byte{[]byte("modified"), []byte("short"), []byte("original-long")} {
		store.putPayload = payload
		stream, err := service.OpenSealed(t.Context(), "operation", "incident_portability", sealed)
		if err != nil {
			t.Fatal(err)
		}
		_, err = io.Copy(io.Discard, stream)
		_ = stream.Close()
		if !errors.Is(err, ErrIntegrity) {
			t.Fatalf("altered content admitted: %v", err)
		}
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := service.AllocateStream(ctx, "operation", "incident_portability", bytes.NewReader(nil), 1); err == nil {
		t.Fatal("cancelled allocation sealed")
	}
	t.Run("a scope validates each seal once per operation", checkScopeSealValidation)
}

// checkScopeSealValidation: operation, owner and seal are validated once per
// scope; every read still rechecks the immutable lease and verifies bytes.
func checkScopeSealValidation(t *testing.T) {
	now := time.Now()
	repo := &serviceRepositoryFake{expiresAt: now.Add(StagingLifetime)}
	store := &byteStoreFake{putOutcome: ByteWriteSuccess}
	service := newTestService(t, repo, store, now, func() (string, error) { return "cached", nil }, nil)
	scope, err := NewScope("operation", "incident_portability", service)
	if err != nil {
		t.Fatal(err)
	}
	if !scope.VerifiesSealedContent() {
		t.Fatal("service-backed scope does not report seal verification")
	}
	sealed, err := scope.AllocateStream(t.Context(), "operation", bytes.NewReader([]byte("original")), 64)
	if err != nil {
		t.Fatal(err)
	}
	read := func() error {
		stream, err := scope.OpenSealed(t.Context(), "operation", sealed)
		if err != nil {
			return err
		}
		_, err = io.Copy(io.Discard, stream)
		return errors.Join(err, stream.Close())
	}
	repo.readSealedCalls = 0
	for range 3 {
		if err := read(); err != nil {
			t.Fatal(err)
		}
	}
	if repo.readSealedCalls != 1 {
		t.Fatalf("seal validated %d times for one operation", repo.readSealedCalls)
	}
	store.putPayload = []byte("modified")
	if err := read(); !errors.Is(err, ErrIntegrity) {
		t.Fatalf("cached seal skipped byte verification: %v", err)
	}
	expired := newTestService(t, repo, store, now.Add(StagingLifetime), func() (string, error) { return "late", nil }, nil)
	if _, err := expired.OpenValidated(t.Context(), SealedObject{StorageIdentity: "x", ContentBinding: sealed.ContentBinding, ExpiresAt: repo.expiresAt}); !errors.Is(err, ErrScopeDenied) {
		t.Fatalf("expired lease opened: %v", err)
	}
}
