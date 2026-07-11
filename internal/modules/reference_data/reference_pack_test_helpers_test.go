package reference_data_test

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

type bundleOptions struct {
	PackKey         string
	PackKind        string
	PackVersion     string
	ContractVersion string
	PayloadPath     string
	BadPayloadSHA   bool
	Signed          bool
	BadSignature    bool
	ExtraPath       string
	OmitPayload     bool
}

func referencePackBundle(t testing.TB, options bundleOptions) []byte {
	t.Helper()
	if options.ContractVersion == "" {
		options.ContractVersion = reference_data.PackContractVersionV1
	}
	if options.PayloadPath == "" {
		options.PayloadPath = "payload/data.json"
	}
	payload := []byte(`{"items":[{"key":"host","label":"Host"}]}`)
	payloadSHABytes := sha256.Sum256(payload)
	payloadSHA := hex.EncodeToString(payloadSHABytes[:])
	if options.BadPayloadSHA {
		payloadSHA = "0000000000000000000000000000000000000000000000000000000000000000"
	}
	payloads := []map[string]string{{"path": options.PayloadPath, "sha256": payloadSHA}}
	payloadJSON, err := json.Marshal(payloads)
	if err != nil {
		t.Fatalf("canonical payload json: %v", err)
	}
	payloadCanonicalSHABytes := sha256.Sum256(payloadJSON)
	canonicalPayloadSHA := hex.EncodeToString(payloadCanonicalSHABytes[:])
	manifest := map[string]any{
		"pack_key":              options.PackKey,
		"pack_kind":             options.PackKind,
		"pack_version":          options.PackVersion,
		"pack_contract_version": options.ContractVersion,
		"source_identifier":     "https://offline.invalid/reference-packs/" + options.PackKey,
		"verification_method":   "manifest_sha256_v1",
		"payloads": []map[string]any{
			{"path": options.PayloadPath, "sha256": payloadSHA},
		},
	}
	if options.Signed {
		signatureSHA := canonicalPayloadSHA
		if options.BadSignature {
			signatureSHA = "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
		}
		manifest["verification_method"] = "signed_manifest_v1"
		manifest["signer_key_id"] = "fixture-key"
		manifest["signature"] = map[string]any{"payload_sha256": signatureSHA}
	}
	manifestBytes, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("marshal manifest: %v", err)
	}
	var buffer bytes.Buffer
	writer := zip.NewWriter(&buffer)
	addZipFile(t, writer, "manifest.json", manifestBytes)
	if !options.OmitPayload {
		addZipFile(t, writer, options.PayloadPath, payload)
	}
	if options.ExtraPath != "" {
		addZipFile(t, writer, options.ExtraPath, []byte("{}"))
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
	return buffer.Bytes()
}

func addZipFile(t testing.TB, writer *zip.Writer, name string, data []byte) {
	t.Helper()
	file, err := writer.Create(name)
	if err != nil {
		t.Fatalf("create zip member %s: %v", name, err)
	}
	if _, err := file.Write(data); err != nil {
		t.Fatalf("write zip member %s: %v", name, err)
	}
}

func postReferencePackUpload(t testing.TB, baseURL string, login flowtest.LoginResult, metadata string, bundle []byte, filename string, contentType string) *http.Response {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	metadataHeader := textproto.MIMEHeader{}
	metadataHeader.Set("Content-Disposition", `form-data; name="metadata"`)
	metadataHeader.Set("Content-Type", "application/json")
	metadataPart, err := writer.CreatePart(metadataHeader)
	if err != nil {
		t.Fatalf("create metadata part: %v", err)
	}
	if _, err := io.WriteString(metadataPart, metadata); err != nil {
		t.Fatalf("write metadata part: %v", err)
	}
	fileHeader := textproto.MIMEHeader{}
	fileHeader.Set("Content-Disposition", `form-data; name="file"; filename="`+filename+`"`)
	fileHeader.Set("Content-Type", contentType)
	filePart, err := writer.CreatePart(fileHeader)
	if err != nil {
		t.Fatalf("create file part: %v", err)
	}
	if _, err := filePart.Write(bundle); err != nil {
		t.Fatalf("write file part: %v", err)
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close multipart writer: %v", err)
	}
	req, err := http.NewRequest(http.MethodPost, baseURL+"/api/v1/reference-packs/import", &body)
	if err != nil {
		t.Fatalf("new upload request: %v", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set(authn.CSRFHeaderName, login.CSRFCookie.Value)
	req.AddCookie(login.SessionCookie)
	req.AddCookie(login.CSRFCookie)
	return httptestx.Do(t, http.DefaultClient, req)
}

func requireJob(t testing.TB, harness *scenariotest.ServerHarness, login flowtest.LoginResult, jobID string) map[string]any {
	t.Helper()
	deadline := time.Now().Add(5 * time.Second)
	var job map[string]any
	for {
		job = requireJobNow(t, harness, login, jobID)
		switch job["status"] {
		case "succeeded", "failed", "canceled":
			return job
		}
		if time.Now().After(deadline) {
			t.Fatalf("job %s did not reach terminal state: %#v", jobID, job)
		}
		time.Sleep(25 * time.Millisecond)
	}
}

func requireJobNow(t testing.TB, harness *scenariotest.ServerHarness, login flowtest.LoginResult, jobID string) map[string]any {
	t.Helper()
	resp := httptestx.DoJSON(t, http.MethodGet, harness.Server.HTTP.URL+"/api/v1/jobs/"+jobID, nil, httptestx.WithCookies(login.SessionCookie))
	return httptestx.RequireSuccessEnvelope(t, resp, http.StatusOK)["data"].(map[string]any)
}
