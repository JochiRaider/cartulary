package objectstore_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore/s3transport"
	"github.com/JochiRaider/cartulary/internal/testutil/s3test"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/aws/aws-sdk-go-v2/service/s3/types"
)

func TestS3QualifiedStreamingAndPresigning(t *testing.T) {
	harness := s3test.Start(t)
	bucket := harness.BootstrapBucketT(t, "qualified-s3")
	settings, err := objectstore.ResolveSettings(objectstore.Binding{BindingKind: "managed_service", ServiceRef: "object_primary"}, harness.EnvForServiceRef("object_primary", bucket))
	if err != nil {
		t.Fatal(err)
	}
	store, err := objectstore.Setup(t.Context(), settings, objectstore.Instrumentation{})
	if err != nil {
		t.Fatal(err)
	}
	defer store.Close()
	typed := store.(objectstore.TypedStore)
	client, err := harness.Client(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name  string
		count int
		known bool
	}{
		{"empty", 0, true}, {"known", 1024, true}, {"streaming", 8193, false}, {"multipart", (16 << 20) + 7, true}, {"multipart-streaming", (16 << 20) + 11, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			payload := bytes.Repeat([]byte{0x71}, tc.count)
			size := int64(-1)
			if tc.known {
				size = int64(len(payload))
			}
			result, err := typed.Put(t.Context(), objectstore.PutObjectRequest{Key: tc.name, Body: readerOnly{bytes.NewReader(payload)}, Size: size, ContentType: "application/octet-stream", Purpose: objectstore.PurposeProductUpload})
			if err != nil {
				t.Fatal(err)
			}
			if result.SizeBytes != int64(len(payload)) {
				t.Fatalf("size=%d want=%d", result.SizeBytes, len(payload))
			}
			body, _, err := typed.Get(t.Context(), objectstore.GetObjectRequest{Key: tc.name, Purpose: objectstore.PurposeProductRead})
			if err != nil {
				t.Fatal(err)
			}
			got, readErr := io.ReadAll(body)
			closeErr := body.Close()
			if readErr != nil || closeErr != nil || !bytes.Equal(got, payload) {
				t.Fatalf("round trip: %v %v equal=%t", readErr, closeErr, bytes.Equal(got, payload))
			}
			metadata, err := client.HeadObject(t.Context(), &s3.HeadObjectInput{Bucket: aws.String(bucket), Key: aws.String(tc.name), ChecksumMode: "ENABLED"})
			if err != nil {
				t.Fatal(err)
			}
			if aws.ToString(metadata.ChecksumSHA256) == "" {
				t.Fatal("backend retained no SHA-256 checksum")
			}
		})
	}
	for _, tc := range []struct {
		name     string
		actual   int
		declared int64
	}{
		{"short", 5, 6}, {"extra", 6, 5}, {"empty-extra", 1, 0},
		{"multipart-short", (16 << 20) + 1, (16 << 20) + 2},
		{"multipart-extra", (16 << 20) + 2, (16 << 20) + 1},
		{"boundary-extra", (8 << 20) + 1, 8 << 20},
		{"oversize-hint", 0, 1 << 60},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := typed.Put(t.Context(), objectstore.PutObjectRequest{Key: tc.name, Body: bytes.NewReader(bytes.Repeat([]byte{1}, tc.actual)), Size: tc.declared, Purpose: objectstore.PurposeProductUpload})
			if err == nil {
				t.Fatal("mismatched upload accepted")
			}
			if _, err := typed.Head(t.Context(), objectstore.HeadObjectRequest{Key: tc.name, Purpose: objectstore.PurposeProductRead}); !objectstore.IsObjectNotFound(err) {
				t.Fatalf("mismatched upload published: %v", err)
			}
			pending, err := client.ListMultipartUploads(t.Context(), &s3.ListMultipartUploadsInput{Bucket: aws.String(bucket), Prefix: aws.String(tc.name)})
			if err != nil || len(pending.Uploads) != 0 {
				t.Fatalf("multipart state retained: %v", err)
			}
		})
	}
	t.Run("incorrect upload checksum is rejected before publication", func(t *testing.T) {
		wrong := base64.StdEncoding.EncodeToString(make([]byte, sha256.Size))
		_, err := client.PutObject(t.Context(), &s3.PutObjectInput{Bucket: aws.String(bucket), Key: aws.String("bad-checksum"), Body: bytes.NewReader([]byte("payload")), ChecksumAlgorithm: types.ChecksumAlgorithmSha256, ChecksumSHA256: aws.String(wrong)})
		if s3transport.ErrorCode(err) != "BadDigest" {
			t.Fatalf("wrong checksum: %v", err)
		}
		if _, err := typed.Head(t.Context(), objectstore.HeadObjectRequest{Key: "bad-checksum", Purpose: objectstore.PurposeProductRead}); !objectstore.IsObjectNotFound(err) {
			t.Fatalf("bad checksum published: %v", err)
		}
		created, err := client.CreateMultipartUpload(t.Context(), &s3.CreateMultipartUploadInput{Bucket: aws.String(bucket), Key: aws.String("bad-part"), ChecksumAlgorithm: types.ChecksumAlgorithmSha256})
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() {
			ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
			defer cancel()
			_, err := client.AbortMultipartUpload(ctx, &s3.AbortMultipartUploadInput{Bucket: aws.String(bucket), Key: aws.String("bad-part"), UploadId: created.UploadId})
			if err != nil {
				t.Errorf("abort rejected-part upload: %v", err)
			}
		})
		_, err = client.UploadPart(t.Context(), &s3.UploadPartInput{Bucket: aws.String(bucket), Key: aws.String("bad-part"), UploadId: created.UploadId, PartNumber: aws.Int32(1), Body: bytes.NewReader([]byte("payload")), ChecksumAlgorithm: types.ChecksumAlgorithmSha256, ChecksumSHA256: aws.String(wrong)})
		if s3transport.ErrorCode(err) != "BadDigest" {
			t.Fatalf("wrong part checksum: %v", err)
		}
		parts, err := client.ListParts(t.Context(), &s3.ListPartsInput{Bucket: aws.String(bucket), Key: aws.String("bad-part"), UploadId: created.UploadId})
		if err != nil || len(parts.Parts) != 0 {
			t.Fatalf("rejected part retained: %v", err)
		}
	})
	t.Run("failed multipart is aborted and never published", func(t *testing.T) {
		input := io.MultiReader(bytes.NewReader(bytes.Repeat([]byte{0x62}, 8<<20)), failedUploadReader{})
		_, err := typed.Put(t.Context(), objectstore.PutObjectRequest{Key: "failed", Body: input, Size: -1, Purpose: objectstore.PurposeProductUpload})
		if err == nil {
			t.Fatal("broken stream accepted")
		}
		if _, err := typed.Head(t.Context(), objectstore.HeadObjectRequest{Key: "failed", Purpose: objectstore.PurposeProductRead}); !objectstore.IsObjectNotFound(err) {
			t.Fatalf("failed upload was published: %v", err)
		}
		uploads, err := client.ListMultipartUploads(t.Context(), &s3.ListMultipartUploadsInput{Bucket: aws.String(bucket)})
		if err != nil || len(uploads.Uploads) != 0 {
			t.Fatalf("multipart cleanup: %v remaining=%v", err, uploads)
		}
	})
	t.Run("presigned body and expiration binding", func(t *testing.T) {
		payload := []byte("signed upload payload")
		digest := sha256.Sum256(payload)
		input := objectstore.UploadTargetRequest{Key: "presigned", ByteSize: int64(len(payload)), SHA256Hex: hex.EncodeToString(digest[:]), ExpiresAt: time.Now().Add(time.Minute), Purpose: objectstore.PurposeProductUpload, ContentType: "application/octet-stream"}
		target, err := typed.CreateUploadTarget(t.Context(), input)
		if err != nil {
			t.Fatal(err)
		}
		upload := func(target objectstore.UploadTarget, payload []byte) int {
			t.Helper()
			request, err := http.NewRequestWithContext(t.Context(), target.Method, target.Href, bytes.NewReader(payload))
			if err != nil {
				t.Fatal(err)
			}
			for name, value := range target.Headers {
				request.Header.Set(name, value)
			}
			response, err := client.Options().HTTPClient.Do(request)
			if err != nil {
				t.Fatal(err)
			}
			_, _ = io.Copy(io.Discard, response.Body)
			if err := response.Body.Close(); err != nil {
				t.Fatal(err)
			}
			return response.StatusCode
		}
		if status := upload(target, payload); status < 200 || status >= 300 {
			t.Fatalf("valid presign status=%d", status)
		}
		changed := append([]byte(nil), payload...)
		changed[0] ^= 1
		if status := upload(target, changed); status >= 200 && status < 300 {
			t.Fatal("wrong body accepted")
		}
		body, _, err := typed.Get(t.Context(), objectstore.GetObjectRequest{Key: "presigned", Purpose: objectstore.PurposeProductRead})
		if err != nil {
			t.Fatal(err)
		}
		retained, readErr := io.ReadAll(body)
		closeErr := body.Close()
		if readErr != nil || closeErr != nil || !bytes.Equal(retained, payload) {
			t.Fatal("rejected presign overwrote original bytes")
		}

		substituted := target
		substituted.Method = http.MethodPost
		if status := upload(substituted, payload); status >= 200 && status < 300 {
			t.Fatal("wrong method accepted")
		}
		input.Key = "expired"
		input.ExpiresAt = time.Now().Add(2 * time.Second)
		expired, err := typed.CreateUploadTarget(t.Context(), input)
		if err != nil {
			t.Fatal(err)
		}
		timer := time.NewTimer(3 * time.Second)
		defer timer.Stop()
		select {
		case <-timer.C:
		case <-t.Context().Done():
			t.Fatal(t.Context().Err())
		}
		if status := upload(expired, payload); status >= 200 && status < 300 {
			t.Fatal("expired presign accepted")
		}
	})
	t.Run("cancelled upload has no publication", func(t *testing.T) {
		ctx, cancel := context.WithCancel(t.Context())
		cancel()
		_, err := typed.Put(ctx, objectstore.PutObjectRequest{Key: "cancelled", Body: bytes.NewReader([]byte("payload")), Size: 7, Purpose: objectstore.PurposeProductUpload})
		if err == nil {
			t.Fatal("cancelled upload accepted")
		}
		if _, err := typed.Head(t.Context(), objectstore.HeadObjectRequest{Key: "cancelled", Purpose: objectstore.PurposeProductRead}); !objectstore.IsObjectNotFound(err) {
			t.Fatalf("cancelled upload was published: %v", err)
		}
	})
}

type readerOnly struct{ io.Reader }
type failedUploadReader struct{}

func (failedUploadReader) Read([]byte) (int, error) {
	return 0, errors.New("fixture input interrupted")
}
