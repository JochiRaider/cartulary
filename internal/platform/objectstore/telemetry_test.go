package objectstore

import (
	"context"
	"errors"
	"io"
	"strings"
	"testing"
	"time"
)

func TestTelemetryStorePreservesStoreBehaviorNoSDK(t *testing.T) {
	fake := &telemetryFakeStore{}
	store := InstrumentStore(fake, "0.0.0+unknown")

	target, err := store.UploadTarget(t.Context(), "incident/10000000/blob", time.Now().Add(time.Minute))
	if err != nil {
		t.Fatalf("upload target through telemetry wrapper: %v", err)
	}
	if target.Method != "PUT" {
		t.Fatalf("unexpected target: %#v", target)
	}

	if err := store.PutObject(t.Context(), "incident/10000000/blob", strings.NewReader("payload"), 7, "text/plain"); err != nil {
		t.Fatalf("put through telemetry wrapper: %v", err)
	}
	reader, info, err := store.ReadObject(t.Context(), "incident/10000000/blob", ReadOptions{})
	if err != nil {
		t.Fatalf("read through telemetry wrapper: %v", err)
	}
	_ = reader.Close()
	if info.Size != 7 {
		t.Fatalf("unexpected read info: %#v", info)
	}
	if fake.putKey != "incident/10000000/blob" || fake.readKey != "incident/10000000/blob" {
		t.Fatalf("wrapper changed object keys passed to store: %#v", fake)
	}
}

func TestObjectStoreErrorClass(t *testing.T) {
	if got := objectStoreErrorClass(adapterError(OperationGetObject, ErrorCodeUnavailable, ReasonEndpointUnreachable, true, "unavailable", nil)); got != "dependency_unavailable" {
		t.Fatalf("unexpected unavailable class: %q", got)
	}
	if got := objectStoreErrorClass(adapterError(OperationGetObject, ErrorCodeDeadlineExceeded, ReasonDeadlineExceeded, true, "timeout", nil)); got != "timeout" {
		t.Fatalf("unexpected timeout class: %q", got)
	}
	if got := objectStoreErrorClass(errors.New("plain failure")); got != "internal_error" {
		t.Fatalf("unexpected generic class: %q", got)
	}
}

type telemetryFakeStore struct {
	putKey  string
	readKey string
}

func (s *telemetryFakeStore) UploadTarget(context.Context, string, time.Time) (UploadTarget, error) {
	return UploadTarget{Href: "/upload", Method: "PUT", Headers: map[string]string{}}, nil
}

func (s *telemetryFakeStore) CompleteUploadTarget(context.Context, string, io.Reader, string) error {
	return nil
}

func (s *telemetryFakeStore) PutObject(_ context.Context, key string, body io.Reader, _ int64, _ string) error {
	s.putKey = key
	_, _ = io.Copy(io.Discard, body)
	return nil
}

func (s *telemetryFakeStore) ReadObject(_ context.Context, key string, _ ReadOptions) (io.ReadCloser, ObjectInfo, error) {
	s.readKey = key
	return io.NopCloser(strings.NewReader("payload")), ObjectInfo{Key: key, Size: 7, ContentType: "text/plain"}, nil
}

func (s *telemetryFakeStore) StatObject(_ context.Context, key string) (ObjectInfo, error) {
	return ObjectInfo{Key: key, Size: 7, ContentType: "text/plain"}, nil
}

func (s *telemetryFakeStore) ListObjects(context.Context, string) ([]ObjectInfo, error) {
	return []ObjectInfo{{Key: "object", Size: 1}}, nil
}

func (s *telemetryFakeStore) DeleteObject(context.Context, string) error {
	return nil
}

func (s *telemetryFakeStore) Close() error {
	return nil
}
