package imports_test

import (
	"archive/zip"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"math/rand/v2"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	evidenceflow "github.com/JochiRaider/cartulary/internal/modules/evidence/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/networkflowsupport"
)

// This exercises a valid bulk member beyond the control-payload ceiling through
// the public lifecycle. Instrumented member/store tests separately prove bounded
// reads and cleanup; this is not a throughput or total-memory benchmark.
func TestLargeIncidentMemberTransfersThroughPublicOperations_Integration(t *testing.T) {
	const size = int64(65 << 20)
	runtime := appsupport.StartRuntime(t)
	start := func(prefix string) *appsupport.ServerHarness {
		return runtime.StartServer(t, appsupport.ServerOptions{Prefix: prefix, TestRouteMode: httptestx.TestRouteModeDisabled, Env: map[string]string{
			"CARTULARY__NETWORK_FLOW_ACTIVITY__CLAIMED":                "true",
			"CARTULARY__NETWORK_FLOW_ACTIVITY__KEY_RING_MANIFEST_PATH": fixtures.Path("network-flow", "key-rings.json"),
			"CARTULARY_SECRET_TEST_NETWORK_FLOW_CURSOR":                "AQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQEBAQE",
			"CARTULARY_SECRET_TEST_NETWORK_FLOW_SAFE_DIGEST":           "AgICAgICAgICAgICAgICAgICAgICAgICAgICAgICAgI",
		}})
	}
	a, b := start("bulk-member-a"), start("bulk-member-b")
	loginA, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, a.Server.HTTP.URL)
	loginB, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, b.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, a.Server, loginA, map[string]any{"client_txn_id": "bulk-incident", "incident_key": "BULK-MEMBER", "title": "Synthetic bounded bulk transport"})["incident_id"].(string)
	left := irPublicClient{t: t, origin: a.Server.HTTP.URL, login: loginA, incident: incident}
	right := irPublicClient{t: t, origin: b.Server.HTTP.URL, login: loginB, incident: incident}
	table := networkflowsupport.SeedGraphSource(t, http.DefaultClient, left.origin, loginA, incident)

	// Deterministic harmless bytes, generated independently of product digests.
	// Incompressible input also stays within the unchanged compression ceiling.
	source, err := os.CreateTemp(t.TempDir(), "bulk-source-*")
	if err != nil {
		t.Fatal(err)
	}
	defer source.Close()
	hash := sha256.New()
	random := rand.NewChaCha8([32]byte{65, 6, 10, 7})
	if n, err := io.CopyBuffer(io.MultiWriter(source, hash), io.LimitReader(random, size), make([]byte, 64<<10)); err != nil || n != size {
		t.Fatalf("create bulk source: n=%d err=%v", n, err)
	}
	digest := hex.EncodeToString(hash.Sum(nil))
	if _, err := source.Seek(0, io.SeekStart); err != nil {
		t.Fatal(err)
	}
	slot := left.request(http.MethodPost, "/api/v1/object-blobs", map[string]any{"incident_id": incident, "client_txn_id": "bulk-blob", "byte_size": size, "filename_hint": "synthetic-bulk.bin", "content_type_hint": "application/octet-stream", "sha256_hex": digest}, http.StatusCreated)
	href := slot["upload_target"].(map[string]any)["href"].(string)
	if strings.HasPrefix(href, "/") {
		href = left.origin + href
	}
	request, err := http.NewRequestWithContext(t.Context(), http.MethodPut, href, source)
	if err != nil {
		t.Fatal(err)
	}
	request.ContentLength = size
	request.Header.Set("Content-Type", "application/octet-stream")
	request.Header.Set(authn.CSRFHeaderName, loginA.CSRF.Value)
	request.AddCookie(loginA.Session)
	request.AddCookie(loginA.CSRF)
	response := httptestx.Do(t, http.DefaultClient, request)
	response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode >= 300 {
		t.Fatalf("bulk upload status=%d", response.StatusCode)
	}
	evidenceflow.VerifyUploadedBlob(t, http.DefaultClient, left.origin, loginA.Session, loginA.CSRF, slot["object_blob_id"].(string), "bulk-verify")
	evidence := left.create("cartulary.view.evidence.v1", "bulk-evidence", map[string]any{"evidence.title": "SYNTHETIC 65 MiB transfer witness", "evidence.initial_object_blob_id": slot["object_blob_id"]})
	evidence = left.patch("cartulary.view.evidence.v1", "bulk-evidence-available", evidence, irValue("evidence.lifecycle_state", "available"))

	job := left.request(http.MethodPost, "/api/v1/incident-bundles/export", map[string]any{"incident_id": incident, "client_txn_id": "bulk-export", "optional_sections": []any{}}, http.StatusAccepted)
	terminal := waitIRApply(t, left.origin, loginA, job["job_id"].(string))
	if terminal["status"] != "succeeded" {
		t.Fatalf("bulk export failed: %#v", terminal)
	}
	bundleID := terminal["result_summary"].(map[string]any)["resource_refs"].([]any)[0].(map[string]any)["id"].(string)
	response = doImportJSON(t, left.origin, loginA, http.MethodGet, "/api/v1/incident-bundles/"+bundleID+"/content", nil)
	if response.StatusCode != http.StatusOK {
		response.Body.Close()
		t.Fatalf("bulk download status=%d", response.StatusCode)
	}
	archive, err := os.CreateTemp(t.TempDir(), "bulk-bundle-*.zip")
	if err != nil {
		response.Body.Close()
		t.Fatal(err)
	}
	defer archive.Close()
	n, err := io.CopyBuffer(archive, response.Body, make([]byte, 64<<10))
	response.Body.Close()
	if err != nil || n < size {
		t.Fatalf("bulk archive truncated: n=%d err=%v", n, err)
	}
	reader, err := zip.NewReader(archive, n)
	if err != nil {
		t.Fatal(err)
	}
	large := 0
	for _, member := range reader.File {
		if member.UncompressedSize64 <= 64<<20 {
			continue
		}
		large++
		if member.UncompressedSize64 != uint64(size) {
			t.Fatal("unexpected large member size")
		}
		content, err := member.Open()
		if err != nil {
			t.Fatal(err)
		}
		defer content.Close()
		assertBulkHash(t, content, size, digest)
		content.Close()
	}
	if large != 1 {
		t.Fatalf("large members=%d, want exactly one", large)
	}
	if _, err := archive.Seek(0, io.SeekStart); err != nil {
		t.Fatal(err)
	}
	imported := uploadPublicBundleFile(t, right, archive, "bulk-import")
	if terminal := waitIRApply(t, right.origin, loginB, imported["job_id"].(string)); terminal["status"] != "succeeded" {
		t.Fatalf("bulk import failed: %#v", terminal)
	}
	right.request(http.MethodGet, "/api/v1/incidents/"+incident+"/network-flow/tables/"+table, nil, http.StatusOK)
	handle := right.request(http.MethodPost, "/api/v1/evidence-records/"+evidence["record_id"].(string)+"/download-handle", map[string]any{}, http.StatusOK)
	href = handle["href"].(string)
	if strings.HasPrefix(href, "/") {
		href = right.origin + href
	}
	request, err = http.NewRequestWithContext(t.Context(), http.MethodGet, href, nil)
	if err != nil {
		t.Fatal(err)
	}
	request.AddCookie(loginB.Session)
	response = httptestx.Do(t, http.DefaultClient, request)
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK {
		t.Fatalf("retained bulk content status=%d", response.StatusCode)
	}
	assertBulkHash(t, response.Body, size, digest)
}

func assertBulkHash(t *testing.T, input io.Reader, size int64, digest string) {
	t.Helper()
	hash := sha256.New()
	n, err := io.CopyBuffer(hash, io.LimitReader(input, size+1), make([]byte, 64<<10))
	if err != nil || n != size || hex.EncodeToString(hash.Sum(nil)) != digest {
		t.Fatalf("bulk bytes/hash changed: n=%d err=%v", n, err)
	}
}

func uploadPublicBundleFile(t *testing.T, c irPublicClient, archive *os.File, txn string) map[string]any {
	t.Helper()
	input, output := io.Pipe()
	writer := multipart.NewWriter(output)
	finished := make(chan error, 1)
	go func() {
		part, err := writer.CreatePart(textproto.MIMEHeader{"Content-Disposition": {`form-data; name="metadata"`}, "Content-Type": {"application/json"}})
		if err == nil {
			err = json.NewEncoder(part).Encode(map[string]any{"client_txn_id": txn})
		}
		if err == nil {
			part, err = writer.CreateFormFile("file", filepath.Base(archive.Name()))
		}
		if err == nil {
			_, err = io.CopyBuffer(part, archive, make([]byte, 64<<10))
		}
		if err == nil {
			err = writer.Close()
		}
		_ = output.CloseWithError(err)
		finished <- err
	}()
	request, err := http.NewRequestWithContext(t.Context(), http.MethodPost, c.origin+"/api/v1/incident-bundles/import", input)
	if err != nil {
		_ = input.CloseWithError(err)
		<-finished
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.Header.Set(authn.CSRFHeaderName, c.login.CSRF.Value)
	request.AddCookie(c.login.Session)
	request.AddCookie(c.login.CSRF)
	response, err := http.DefaultClient.Do(request)
	_ = input.CloseWithError(err)
	writeErr := <-finished
	if err != nil {
		t.Fatal(err)
	}
	if writeErr != nil {
		response.Body.Close()
		t.Fatal(writeErr)
	}
	return httptestx.RequireSuccessEnvelope(t, response, http.StatusAccepted)["data"].(map[string]any)
}
