package httpapi

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"reflect"
	"strings"
	"testing"
)

var testUploadEnvelopeFileTypes = []string{
	"text/csv",
	"application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
	"application/octet-stream",
}

func TestUploadEnvelopeAcceptsExactMetadataAndFileParts(t *testing.T) {
	t.Parallel()

	fileBody := []byte("col\nvalue\n")
	request := uploadEnvelopeRequest(t, []uploadEnvelopeTestPart{
		{
			Name:        "file",
			ContentType: "Text/CSV; charset=utf-8",
			Filename:    "source.csv",
			Body:        fileBody,
		},
		{
			Name:        "metadata",
			ContentType: "application/json; charset=UTF-8",
			Filename:    "ignored.json",
			Body:        []byte(`{"client_txn_id":"txn-1","nested":{"ok":true}}`),
		},
	}, "")

	envelope, apiErr := ParseUploadEnvelope(request, UploadEnvelopePolicy{FileContentTypes: testUploadEnvelopeFileTypes})
	if apiErr != nil {
		t.Fatalf("ParseUploadEnvelope() error: %v details=%#v", apiErr, apiErr.Details())
	}
	if string(envelope.File) != string(fileBody) {
		t.Fatalf("unexpected file body: %q", string(envelope.File))
	}
	hash := sha256.Sum256(fileBody)
	if envelope.FileSHA256Hex != hex.EncodeToString(hash[:]) {
		t.Fatalf("unexpected file sha256: got %s want %s", envelope.FileSHA256Hex, hex.EncodeToString(hash[:]))
	}
	if envelope.FileContentType != "text/csv" {
		t.Fatalf("unexpected normalized file content type: %q", envelope.FileContentType)
	}
	if _, ok := envelope.Metadata["client_txn_id"]; !ok {
		t.Fatalf("metadata object missing client_txn_id: %#v", envelope.Metadata)
	}
	if string(envelope.MetadataRaw) != `{"client_txn_id":"txn-1","nested":{"ok":true}}` {
		t.Fatalf("metadata raw changed: %s", envelope.MetadataRaw)
	}
}

func TestUploadEnvelopeRejectsClosedEnvelopeFailures(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name       string
		request    func(t testing.TB) *http.Request
		reasonCode string
		partName   string
	}{
		{
			name: "unsupported request media type",
			request: func(t testing.TB) *http.Request {
				t.Helper()
				request := httptest.NewRequest(http.MethodPost, "/upload", strings.NewReader(`{"metadata":{}}`))
				request.Header.Set("Content-Type", "application/json")
				return request
			},
			reasonCode: UploadEnvelopeReasonUnsupportedUploadEnvelope,
		},
		{
			name: "missing multipart boundary",
			request: func(t testing.TB) *http.Request {
				t.Helper()
				request := httptest.NewRequest(http.MethodPost, "/upload", strings.NewReader(""))
				request.Header.Set("Content-Type", "multipart/form-data")
				return request
			},
			reasonCode: UploadEnvelopeReasonUnsupportedUploadEnvelope,
		},
		{
			name: "missing metadata",
			request: func(t testing.TB) *http.Request {
				t.Helper()
				return uploadEnvelopeRequest(t, []uploadEnvelopeTestPart{validFilePart()}, "")
			},
			reasonCode: UploadEnvelopeReasonMissingRequiredPart,
			partName:   "metadata",
		},
		{
			name: "missing file",
			request: func(t testing.TB) *http.Request {
				t.Helper()
				return uploadEnvelopeRequest(t, []uploadEnvelopeTestPart{validMetadataPart(`{"client_txn_id":"txn-1"}`)}, "")
			},
			reasonCode: UploadEnvelopeReasonMissingRequiredPart,
			partName:   "file",
		},
		{
			name: "duplicate metadata",
			request: func(t testing.TB) *http.Request {
				t.Helper()
				return uploadEnvelopeRequest(t, []uploadEnvelopeTestPart{
					validMetadataPart(`{"client_txn_id":"txn-1"}`),
					validMetadataPart(`{"client_txn_id":"txn-2"}`),
					validFilePart(),
				}, "")
			},
			reasonCode: UploadEnvelopeReasonDuplicatePart,
			partName:   "metadata",
		},
		{
			name: "duplicate file",
			request: func(t testing.TB) *http.Request {
				t.Helper()
				return uploadEnvelopeRequest(t, []uploadEnvelopeTestPart{
					validMetadataPart(`{"client_txn_id":"txn-1"}`),
					validFilePart(),
					validFilePart(),
				}, "")
			},
			reasonCode: UploadEnvelopeReasonDuplicatePart,
			partName:   "file",
		},
		{
			name: "unexpected part",
			request: func(t testing.TB) *http.Request {
				t.Helper()
				return uploadEnvelopeRequest(t, []uploadEnvelopeTestPart{
					validMetadataPart(`{"client_txn_id":"txn-1"}`),
					validFilePart(),
					{Name: "extra", ContentType: "text/plain", Body: []byte("extra")},
				}, "")
			},
			reasonCode: UploadEnvelopeReasonUnexpectedPart,
		},
		{
			name: "nested metadata multipart",
			request: func(t testing.TB) *http.Request {
				t.Helper()
				return uploadEnvelopeRequest(t, []uploadEnvelopeTestPart{
					{Name: "metadata", ContentType: "multipart/mixed; boundary=inner", Body: []byte("--inner--")},
					validFilePart(),
				}, "")
			},
			reasonCode: UploadEnvelopeReasonUnsupportedUploadEnvelope,
			partName:   "metadata",
		},
		{
			name: "nested file multipart",
			request: func(t testing.TB) *http.Request {
				t.Helper()
				return uploadEnvelopeRequest(t, []uploadEnvelopeTestPart{
					validMetadataPart(`{"client_txn_id":"txn-1"}`),
					{Name: "file", ContentType: "multipart/mixed; boundary=inner", Body: []byte("--inner--")},
				}, "")
			},
			reasonCode: UploadEnvelopeReasonUnsupportedUploadEnvelope,
			partName:   "file",
		},
		{
			name: "malformed metadata json",
			request: func(t testing.TB) *http.Request {
				t.Helper()
				return uploadEnvelopeRequest(t, []uploadEnvelopeTestPart{
					validMetadataPart(`{"client_txn_id":`),
					validFilePart(),
				}, "")
			},
			reasonCode: UploadEnvelopeReasonMalformedMetadataJSON,
			partName:   "metadata",
		},
		{
			name: "metadata json not object",
			request: func(t testing.TB) *http.Request {
				t.Helper()
				return uploadEnvelopeRequest(t, []uploadEnvelopeTestPart{
					validMetadataPart(`[]`),
					validFilePart(),
				}, "")
			},
			reasonCode: UploadEnvelopeReasonRequestNotObject,
			partName:   "metadata",
		},
		{
			name: "non utf-8 metadata",
			request: func(t testing.TB) *http.Request {
				t.Helper()
				return uploadEnvelopeRequest(t, []uploadEnvelopeTestPart{
					{Name: "metadata", ContentType: "application/json", Body: []byte{0xff, '{', '}'}},
					validFilePart(),
				}, "")
			},
			reasonCode: UploadEnvelopeReasonInvalidMetadataEncoding,
			partName:   "metadata",
		},
		{
			name: "bom metadata",
			request: func(t testing.TB) *http.Request {
				t.Helper()
				return uploadEnvelopeRequest(t, []uploadEnvelopeTestPart{
					{Name: "metadata", ContentType: "application/json", Body: append([]byte{0xef, 0xbb, 0xbf}, []byte(`{}`)...)},
					validFilePart(),
				}, "")
			},
			reasonCode: UploadEnvelopeReasonInvalidMetadataEncoding,
			partName:   "metadata",
		},
		{
			name: "duplicate top level json key",
			request: func(t testing.TB) *http.Request {
				t.Helper()
				return uploadEnvelopeRequest(t, []uploadEnvelopeTestPart{
					validMetadataPart(`{"client_txn_id":"txn-1","client_txn_id":"txn-2"}`),
					validFilePart(),
				}, "")
			},
			reasonCode: UploadEnvelopeReasonMalformedMetadataJSON,
			partName:   "metadata",
		},
		{
			name: "duplicate nested json key",
			request: func(t testing.TB) *http.Request {
				t.Helper()
				return uploadEnvelopeRequest(t, []uploadEnvelopeTestPart{
					validMetadataPart(`{"outer":{"same":1,"same":2}}`),
					validFilePart(),
				}, "")
			},
			reasonCode: UploadEnvelopeReasonMalformedMetadataJSON,
			partName:   "metadata",
		},
		{
			name: "disallowed metadata content type",
			request: func(t testing.TB) *http.Request {
				t.Helper()
				return uploadEnvelopeRequest(t, []uploadEnvelopeTestPart{
					{Name: "metadata", ContentType: "application/json; charset=utf-16", Body: []byte(`{}`)},
					validFilePart(),
				}, "")
			},
			reasonCode: UploadEnvelopeReasonInvalidPartContentType,
			partName:   "metadata",
		},
		{
			name: "disallowed file content type",
			request: func(t testing.TB) *http.Request {
				t.Helper()
				return uploadEnvelopeRequest(t, []uploadEnvelopeTestPart{
					validMetadataPart(`{"client_txn_id":"txn-1"}`),
					{Name: "file", ContentType: "text/plain", Body: []byte("plain")},
				}, "")
			},
			reasonCode: UploadEnvelopeReasonInvalidPartContentType,
			partName:   "file",
		},
		{
			name: "missing file content type",
			request: func(t testing.TB) *http.Request {
				t.Helper()
				return uploadEnvelopeRequest(t, []uploadEnvelopeTestPart{
					validMetadataPart(`{"client_txn_id":"txn-1"}`),
					{Name: "file", Body: []byte("plain")},
				}, "")
			},
			reasonCode: UploadEnvelopeReasonInvalidPartContentType,
			partName:   "file",
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			_, apiErr := ParseUploadEnvelope(tc.request(t), UploadEnvelopePolicy{FileContentTypes: testUploadEnvelopeFileTypes})
			if apiErr == nil {
				t.Fatal("ParseUploadEnvelope() succeeded, want error")
			}
			if apiErr.ReasonCode != tc.reasonCode {
				t.Fatalf("unexpected reason: got %q want %q details=%#v", apiErr.ReasonCode, tc.reasonCode, apiErr.Details())
			}
			if apiErr.PartName != tc.partName {
				t.Fatalf("unexpected part name: got %q want %q details=%#v", apiErr.PartName, tc.partName, apiErr.Details())
			}
		})
	}
}

func TestUploadEnvelopeInvalidPartContentTypeDetails(t *testing.T) {
	t.Parallel()

	request := uploadEnvelopeRequest(t, []uploadEnvelopeTestPart{
		validMetadataPart(`{"client_txn_id":"txn-1"}`),
		{Name: "file", ContentType: "text/plain", Body: []byte("plain")},
	}, "")
	_, apiErr := ParseUploadEnvelope(request, UploadEnvelopePolicy{FileContentTypes: []string{
		"application/octet-stream",
		"text/csv",
	}})
	if apiErr == nil {
		t.Fatal("ParseUploadEnvelope() succeeded, want error")
	}

	details := apiErr.Details()
	if details["reason_code"] != UploadEnvelopeReasonInvalidPartContentType || details["part_name"] != "file" {
		t.Fatalf("unexpected details: %#v", details)
	}
	if details["received_content_type"] != "text/plain" {
		t.Fatalf("unexpected received content type: %#v", details)
	}
	wantAllowed := []string{"application/octet-stream", "text/csv"}
	if !reflect.DeepEqual(details["allowed_content_types"], wantAllowed) {
		t.Fatalf("unexpected allowed content types: got %#v want %#v", details["allowed_content_types"], wantAllowed)
	}
}

func TestUploadEnvelopeClosedReasonRegistry(t *testing.T) {
	t.Parallel()

	wantReasons := []string{
		UploadEnvelopeReasonUnsupportedUploadEnvelope,
		UploadEnvelopeReasonMissingRequiredPart,
		UploadEnvelopeReasonDuplicatePart,
		UploadEnvelopeReasonUnexpectedPart,
		UploadEnvelopeReasonInvalidPartContentType,
		UploadEnvelopeReasonInvalidMetadataEncoding,
		UploadEnvelopeReasonMalformedMetadataJSON,
	}
	if got := SharedUploadEnvelopeReasonCodes(); !reflect.DeepEqual(got, wantReasons) {
		t.Fatalf("unexpected reason registry: got %#v want %#v", got, wantReasons)
	}
	wantMetadataTypes := []string{"application/json", "application/json; charset=utf-8"}
	if got := MetadataUploadEnvelopeContentTypes(); !reflect.DeepEqual(got, wantMetadataTypes) {
		t.Fatalf("unexpected metadata content types: got %#v want %#v", got, wantMetadataTypes)
	}
}

type uploadEnvelopeTestPart struct {
	Name        string
	ContentType string
	Filename    string
	Body        []byte
}

func validMetadataPart(body string) uploadEnvelopeTestPart {
	return uploadEnvelopeTestPart{
		Name:        "metadata",
		ContentType: "application/json",
		Body:        []byte(body),
	}
}

func validFilePart() uploadEnvelopeTestPart {
	return uploadEnvelopeTestPart{
		Name:        "file",
		ContentType: "text/csv",
		Filename:    "source.csv",
		Body:        []byte("column\nvalue\n"),
	}
}

func uploadEnvelopeRequest(t testing.TB, parts []uploadEnvelopeTestPart, contentTypeOverride string) *http.Request {
	t.Helper()

	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for _, part := range parts {
		header := textproto.MIMEHeader{}
		dispositionParams := map[string]string{"name": part.Name}
		if part.Filename != "" {
			dispositionParams["filename"] = part.Filename
		}
		header.Set("Content-Disposition", mime.FormatMediaType("form-data", dispositionParams))
		if part.ContentType != "" {
			header.Set("Content-Type", part.ContentType)
		}
		partWriter, err := writer.CreatePart(header)
		if err != nil {
			t.Fatalf("create multipart part %q: %v", part.Name, err)
		}
		if _, err := io.Copy(partWriter, bytes.NewReader(part.Body)); err != nil {
			t.Fatalf("write multipart part %q: %v", part.Name, err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}

	request := httptest.NewRequest(http.MethodPost, "/upload", &body)
	contentType := writer.FormDataContentType()
	if contentTypeOverride != "" {
		contentType = fmt.Sprintf(contentTypeOverride, writer.Boundary())
	}
	request.Header.Set("Content-Type", contentType)
	return request
}
