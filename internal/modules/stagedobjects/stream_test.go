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
}
