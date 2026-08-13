package objectstore_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
)

func TestObjectStoreAdapterContractHardening(t *testing.T) {
	requireObjectStoreAdapterContractHardening(t)
}

func requireObjectStoreAdapterContractHardening(t *testing.T) {
	t.Helper()
	t.Run("input_contracts", requireObjectStoreAdapterInputContracts)
	t.Run("retry_algorithm", requireObjectStoreAdapterRetryAlgorithm)
	t.Run("read_streams_are_close_observable", requireObjectStoreAdapterReadStreamsAreCloseObservable)
}

func requireObjectStoreAdapterInputContracts(t *testing.T) {
	store, err := objectstore.NewFilesystemStore(t.TempDir())
	if err != nil {
		t.Fatalf("create filesystem store: %v", err)
	}
	defer store.Close()

	_, err = store.CreateUploadTarget(context.Background(), objectstore.UploadTargetRequest{
		Key:       "../escape.txt",
		ByteSize:  1,
		ExpiresAt: time.Now().Add(time.Minute),
		Purpose:   objectstore.PurposeProductUpload,
	})
	requireAdapterError(t, err, objectstore.ErrorCodeInvalidRequest, objectstore.ReasonInvalidRequest)

	_, err = store.Put(context.Background(), objectstore.PutObjectRequest{
		Key:         "proof.txt",
		Body:        strings.NewReader("proof"),
		Size:        int64(len("proof")),
		ContentType: "text/plain\r\nx-bad: yes",
		Purpose:     objectstore.PurposeProductUpload,
	})
	requireAdapterError(t, err, objectstore.ErrorCodeInvalidRequest, objectstore.ReasonInvalidRequest)

	start := int64(10)
	end := int64(4)
	_, _, err = store.Get(context.Background(), objectstore.GetObjectRequest{
		Key:        "proof.txt",
		RangeStart: &start,
		RangeEnd:   &end,
		Purpose:    objectstore.PurposeProductRead,
	})
	requireAdapterError(t, err, objectstore.ErrorCodeInvalidRequest, objectstore.ReasonInvalidRequest)

	typed := objectstore.TypedStore(store)
	if err := store.PutObject(context.Background(), "evidence/cleanup-purpose.txt", strings.NewReader("cleanup"), int64(len("cleanup")), "text/plain"); err != nil {
		t.Fatalf("seed Evidence cleanup purpose object: %v", err)
	}
	if err := typed.Delete(context.Background(), objectstore.DeleteObjectRequest{
		Key: "evidence/cleanup-purpose.txt", Purpose: objectstore.PurposeEvidenceCleanup,
	}); err != nil {
		t.Fatalf("typed Evidence cleanup purpose was rejected: %v", err)
	}
	if err := typed.Delete(context.Background(), objectstore.DeleteObjectRequest{
		Key: "evidence/disallowed-delete.txt", Purpose: objectstore.PurposeProductRead,
	}); err == nil {
		t.Fatal("non-delete object-store purpose was accepted for delete")
	}
}

func requireObjectStoreAdapterRetryAlgorithm(t *testing.T) {
	restore := objectstore.SetRetryBackoffForTest(0)
	defer restore()

	const bucket = "object-store-retry"
	var mu sync.Mutex
	headAttempts := map[string]int{}
	putAttempts := map[string]int{}
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		bucketMatched, key := adapterContractS3Key(r, bucket)
		switch {
		case r.Method == http.MethodHead && bucketMatched && key == "":
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodHead && bucketMatched && key == ".cartulary/startup/capability-check":
			writeS3Error(w, http.StatusNotFound, "NoSuchKey")
		case r.Method == http.MethodHead && bucketMatched && key == "retry-once.txt":
			mu.Lock()
			headAttempts[key]++
			attempt := headAttempts[key]
			mu.Unlock()
			if attempt == 1 {
				writeS3Error(w, http.StatusServiceUnavailable, "ServiceUnavailable")
				return
			}
			w.Header().Set("Content-Length", "5")
			w.Header().Set("Content-Type", "text/plain")
			w.Header().Set("ETag", `"retry-once"`)
			w.Header().Set("Last-Modified", time.Now().UTC().Format(http.TimeFormat))
			w.WriteHeader(http.StatusOK)
		case r.Method == http.MethodHead && bucketMatched && key == "retry-exhausted.txt":
			mu.Lock()
			headAttempts[key]++
			mu.Unlock()
			writeS3Error(w, http.StatusServiceUnavailable, "ServiceUnavailable")
		case r.Method == http.MethodPut && bucketMatched && key == "put-not-retried.txt":
			mu.Lock()
			putAttempts[key]++
			mu.Unlock()
			writeS3Error(w, http.StatusServiceUnavailable, "ServiceUnavailable")
		case r.Method == http.MethodGet && strings.Contains(r.URL.RawQuery, "location"):
			w.Header().Set("Content-Type", "application/xml")
			w.WriteHeader(http.StatusOK)
			_, _ = io.WriteString(w, "<LocationConstraint></LocationConstraint>")
		default:
			writeS3Error(w, http.StatusNotFound, "NoSuchKey")
		}
	}))
	defer server.Close()

	store, err := setupObjectStore(context.Background(), managedObjectStoreConfig(t), map[string]string{
		"CARTULARY_S3_OBJECT_PRIMARY_ENDPOINT":          strings.TrimPrefix(server.URL, "http://"),
		"CARTULARY_S3_OBJECT_PRIMARY_ACCESS_KEY_ID":     "object-store-access",
		"CARTULARY_S3_OBJECT_PRIMARY_SECRET_ACCESS_KEY": "object-store-secret",
		"CARTULARY_S3_OBJECT_PRIMARY_SECURE":            "false",
		"CARTULARY_S3_OBJECT_PRIMARY_BUCKET":            bucket,
	})
	if err != nil {
		t.Fatalf("setup fake S3 store: %v", err)
	}
	defer store.Close()

	typed, ok := store.(objectstore.TypedStore)
	if !ok {
		t.Fatalf("configured store does not expose typed adapter boundary: %T", store)
	}

	if _, err := typed.Head(context.Background(), objectstore.HeadObjectRequest{Key: "retry-once.txt", Purpose: objectstore.PurposeProductRead}); err != nil {
		t.Fatalf("head retry-once should succeed: %v", err)
	}
	if got := countFor(headAttempts, "retry-once.txt", &mu); got != 2 {
		t.Fatalf("head retry-once attempt count got %d want 2", got)
	}

	_, err = typed.Head(context.Background(), objectstore.HeadObjectRequest{Key: "retry-exhausted.txt", Purpose: objectstore.PurposeProductRead})
	requireAdapterError(t, err, objectstore.ErrorCodeRetryExhausted, objectstore.ReasonRetryExhausted)
	if got := countFor(headAttempts, "retry-exhausted.txt", &mu); got != 2 {
		t.Fatalf("head retry-exhausted attempt count got %d want 2", got)
	}

	_, err = typed.Put(context.Background(), objectstore.PutObjectRequest{
		Key:         "put-not-retried.txt",
		Body:        strings.NewReader("proof"),
		Size:        int64(len("proof")),
		ContentType: "text/plain",
		Purpose:     objectstore.PurposeProductUpload,
	})
	requireAdapterError(t, err, objectstore.ErrorCodeUnavailable, objectstore.ReasonEndpointUnreachable)
	if got := countFor(putAttempts, "put-not-retried.txt", &mu); got != 1 {
		t.Fatalf("put attempt count got %d want 1", got)
	}
}

func requireObjectStoreAdapterReadStreamsAreCloseObservable(t *testing.T) {
	store, err := objectstore.NewFilesystemStore(t.TempDir())
	if err != nil {
		t.Fatalf("create filesystem store: %v", err)
	}
	defer store.Close()

	payload := []byte("object store close tracking")
	if err := store.PutObject(context.Background(), "proof/close.txt", bytes.NewReader(payload), int64(len(payload)), "text/plain"); err != nil {
		t.Fatalf("put object: %v", err)
	}

	events := make([]objectstore.StreamEvent, 0)
	restore := objectstore.SetStreamObserverForTest(func(event objectstore.StreamEvent) {
		events = append(events, event)
	})
	defer restore()

	reader, _, err := store.Get(context.Background(), objectstore.GetObjectRequest{Key: "proof/close.txt", Purpose: objectstore.PurposeProductRead})
	if err != nil {
		t.Fatalf("get object: %v", err)
	}
	if _, err := io.ReadAll(reader); err != nil {
		t.Fatalf("read object: %v", err)
	}
	if err := reader.Close(); err != nil {
		t.Fatalf("close object: %v", err)
	}

	if len(events) != 2 {
		t.Fatalf("stream events got %#v want open and close", events)
	}
	if events[0].Event != "open" || events[1].Event != "close" || events[1].Key != "proof/close.txt" {
		t.Fatalf("unexpected stream events: %#v", events)
	}
}

func requireAdapterError(t testing.TB, err error, code objectstore.ErrorCode, reason objectstore.ReasonCode) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected adapter error %s/%s", code, reason)
	}
	var adapterErr *objectstore.AdapterError
	if !errors.As(err, &adapterErr) {
		t.Fatalf("expected adapter error, got %T %v", err, err)
	}
	if adapterErr.Code != code || adapterErr.Reason != reason {
		t.Fatalf("adapter error got %s/%s want %s/%s", adapterErr.Code, adapterErr.Reason, code, reason)
	}
}

func writeS3Error(w http.ResponseWriter, status int, code string) {
	w.Header().Set("Content-Type", "application/xml")
	w.WriteHeader(status)
	_, _ = fmt.Fprintf(w, "<Error><Code>%s</Code><Message>%s</Message></Error>", code, code)
}

func adapterContractS3Key(r *http.Request, bucket string) (bool, string) {
	host := r.Host
	if colon := strings.LastIndex(host, ":"); colon >= 0 {
		host = host[:colon]
	}
	path := strings.TrimPrefix(r.URL.Path, "/")
	if host == bucket || strings.HasPrefix(host, bucket+".") {
		return true, path
	}
	if path == bucket {
		return true, ""
	}
	prefix := bucket + "/"
	if strings.HasPrefix(path, prefix) {
		return true, strings.TrimPrefix(path, prefix)
	}
	return false, path
}

func countFor(counts map[string]int, key string, mu *sync.Mutex) int {
	mu.Lock()
	defer mu.Unlock()
	return counts[key]
}
