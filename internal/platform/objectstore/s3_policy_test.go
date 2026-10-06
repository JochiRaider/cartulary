package objectstore_test

import (
	"context"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"encoding/hex"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/cryptography"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/testutil/tlstest"
)

func policyS3Server(t *testing.T, handler http.Handler) (*httptest.Server, objectstore.Settings) {
	t.Helper()
	ca, err := tlstest.NewAuthority()
	if err != nil {
		t.Fatal(err)
	}
	identity, err := ca.Issue("s3", []string{"127.0.0.1"}, x509.ExtKeyUsageServerAuth)
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	root, err := tlstest.WriteFile(directory, "root.pem", ca.CertificatePEM)
	if err != nil {
		t.Fatal(err)
	}
	cert, err := tlstest.WriteFile(directory, "server.pem", identity.CertificatePEM)
	if err != nil {
		t.Fatal(err)
	}
	key, err := tlstest.WriteFile(directory, "server.key", identity.PrivateKeyPEM)
	if err != nil {
		t.Fatal(err)
	}
	server := httptest.NewUnstartedServer(handler)
	server.TLS, err = cryptography.TLSServer("127.0.0.1", cert, key)
	if err != nil {
		t.Fatal(err)
	}
	server.StartTLS()
	t.Cleanup(server.Close)
	return server, objectstore.Settings{BindingKind: "managed_service", Endpoint: strings.TrimPrefix(server.URL, "https://"), AccessKey: "test-access", SecretKey: "test-secret-key-at-least-32-bytes", Secure: true, Bucket: "policy-test", RootCertificatePath: root}
}

func TestS3CryptographicPolicy(t *testing.T) {
	t.Run("HEAD denial retains credential classification", func(t *testing.T) {
		_, settings := policyS3Server(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusForbidden) }))
		_, err := objectstore.EnsureBucket(t.Context(), settings)
		adapter, ok := objectstore.AsAdapterError(err)
		if !ok || adapter.Reason != objectstore.ReasonCredentialDenied || adapter.Retryable {
			t.Fatalf("denied HEAD classification: %v", err)
		}
	})

	t.Run("region rejection never negotiates another crypto path", func(t *testing.T) {
		var creates atomic.Int32
		_, settings := policyS3Server(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.TLS == nil || r.TLS.Version != tls.VersionTLS13 {
				t.Error("request did not use TLS 1.3")
			}
			if !strings.HasPrefix(r.Header.Get("Authorization"), "AWS4-HMAC-SHA256 ") {
				t.Error("request did not use SigV4")
			}
			if r.Header.Get("Content-Md5") != "" {
				t.Error("MD5 checksum used")
			}
			if r.Method == http.MethodHead {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			creates.Add(1)
			w.Header().Set("Content-Type", "application/xml")
			w.WriteHeader(http.StatusBadRequest)
			_, _ = io.WriteString(w, "<Error><Code>AuthorizationHeaderMalformed</Code><Region>eu-west-2</Region></Error>")
		}))
		_, err := objectstore.EnsureBucket(t.Context(), settings)
		if err == nil {
			t.Fatal("region correction accepted")
		}
		if creates.Load() != 1 {
			t.Fatalf("unexpected create retries: %d", creates.Load())
		}
	})
	t.Run("regional initialization has no MD5 checksum", func(t *testing.T) {
		var creates atomic.Int32
		_, settings := policyS3Server(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method == http.MethodHead {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			if r.Method != http.MethodPut {
				t.Errorf("unexpected method %s", r.Method)
				return
			}
			creates.Add(1)
			payload, _ := io.ReadAll(r.Body)
			if !strings.Contains(string(payload), "eu-west-2") {
				t.Error("explicit creation region missing")
			}
			if r.Header.Get("Content-Md5") != "" {
				t.Error("MD5 checksum used")
			}
			if !strings.Contains(r.Header.Get("Authorization"), "/eu-west-2/s3/aws4_request") {
				t.Error("wrong signing region")
			}
			w.WriteHeader(http.StatusOK)
		}))
		settings.Region = "eu-west-2"
		result, err := objectstore.EnsureBucket(t.Context(), settings)
		if err != nil || !result.Created || creates.Load() != 1 {
			t.Fatalf("regional init: %+v %v", result, err)
		}
	})
	t.Run("configuration and trust reject before application requests", func(t *testing.T) {
		var requests atomic.Int32
		_, settings := policyS3Server(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { requests.Add(1); w.WriteHeader(http.StatusOK) }))
		for _, mutate := range []func(*objectstore.Settings){
			func(s *objectstore.Settings) { s.Secure = false }, func(s *objectstore.Settings) { s.Endpoint = "http://" + s.Endpoint },
			func(s *objectstore.Settings) { s.Endpoint = "user:secret@" + s.Endpoint }, func(s *objectstore.Settings) { s.Endpoint += "/path" },
			func(s *objectstore.Settings) { s.SecretKey = "short" }, func(s *objectstore.Settings) { s.Region = "../invalid" },
			func(s *objectstore.Settings) { s.RootCertificatePath = t.TempDir() + "/missing" },
		} {
			candidate := settings
			mutate(&candidate)
			if store, err := objectstore.Setup(t.Context(), candidate, objectstore.Instrumentation{}); err == nil {
				store.Close()
				t.Fatal("unsupported connection accepted")
			}
		}
		if requests.Load() != 0 {
			t.Fatal("rejected config reached S3")
		}
	})
	t.Run("presigned PUT binds digest size and method", func(t *testing.T) {
		_, settings := policyS3Server(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if strings.Contains(r.URL.Path, "capability-check") {
				w.WriteHeader(http.StatusNotFound)
			} else {
				w.WriteHeader(http.StatusOK)
			}
		}))
		store, err := objectstore.Setup(t.Context(), settings, objectstore.Instrumentation{})
		if err != nil {
			t.Fatal(err)
		}
		defer store.Close()
		if _, err := store.UploadTarget(t.Context(), "legacy", time.Now().Add(time.Minute)); err == nil {
			t.Fatal("checksum-less legacy target accepted")
		}
		typed := store.(objectstore.TypedStore)
		input := objectstore.UploadTargetRequest{Key: "current", ByteSize: 7, ExpiresAt: time.Now().Add(time.Minute), Purpose: objectstore.PurposeProductUpload}
		if _, err := typed.CreateUploadTarget(t.Context(), input); err == nil {
			t.Fatal("missing digest accepted")
		}
		digest := sha256.Sum256([]byte("payload"))
		input.SHA256Hex = hex.EncodeToString(digest[:])
		target, err := typed.CreateUploadTarget(context.Background(), input)
		if err != nil {
			t.Fatal(err)
		}
		request, err := http.NewRequest(target.Method, target.Href, nil)
		if err != nil {
			t.Fatal(err)
		}
		checksum := target.Headers["X-Amz-Checksum-Sha256"]
		if checksum == "" {
			checksum = request.URL.Query().Get("X-Amz-Checksum-Sha256")
		}
		if target.Method != "PUT" || request.URL.Scheme != "https" || checksum != base64.StdEncoding.EncodeToString(digest[:]) {
			t.Fatal("presign checksum binding missing")
		}
	})
}
