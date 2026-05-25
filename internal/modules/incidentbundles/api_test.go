package incidentbundles

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"os"
	"slices"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
)

func TestPhase11_U_11_INCIDENT_BUNDLES_01_DecodeExportRequestCanonicalizesAndRejectsModes(t *testing.T) {
	request, apiErr := DecodeExportRequest(bytes.NewBufferString(`{
		"incident_id":"11111111-1111-1111-1111-111111111111",
		"client_txn_id":"txn-export",
		"optional_sections":["reference_packs","snapshots","reference_packs"],
		"required_capabilities":["snapshots"],
		"reference_pack_mode":"embedded"
	}`))
	if apiErr != nil {
		t.Fatalf("DecodeExportRequest returned error: %#v", apiErr)
	}
	if request.HistoryMode != HistoryModeFull || request.BlobMode != BlobModeFull {
		t.Fatalf("export request must force full modes, got history=%q blob=%q", request.HistoryMode, request.BlobMode)
	}
	if !slices.Equal(request.OptionalSections, []string{"reference_packs", "snapshots"}) {
		t.Fatalf("optional sections not canonicalized: %#v", request.OptionalSections)
	}
	if !slices.Equal(request.RequiredCapabilities, []string{"snapshots"}) {
		t.Fatalf("required capabilities not canonicalized: %#v", request.RequiredCapabilities)
	}
	var normalized map[string]any
	if err := json.Unmarshal(request.Normalized, &normalized); err != nil {
		t.Fatalf("decode normalized request: %v", err)
	}
	if normalized["history_mode"] != "full" || normalized["blob_mode"] != "full" {
		t.Fatalf("normalized request must include fixed modes: %#v", normalized)
	}

	for _, body := range []string{
		`{"incident_id":"11111111-1111-1111-1111-111111111111","client_txn_id":"txn","history_mode":"partial"}`,
		`{"incident_id":"11111111-1111-1111-1111-111111111111","client_txn_id":"txn","blob_mode":"metadata_only"}`,
	} {
		_, apiErr := DecodeExportRequest(bytes.NewBufferString(body))
		if apiErr == nil || apiErr.Code != "invalid_incident_bundle_request" {
			t.Fatalf("forbidden modes must fail with invalid_incident_bundle_request, got %#v", apiErr)
		}
	}
}

func TestPhase11_U_11_INCIDENT_BUNDLES_02_BundleManifestChecksumDeterministic(t *testing.T) {
	files := map[string][]byte{
		"records.ndjson":           []byte("{}\n"),
		"data/incident.json":       []byte(`{"id":"inc"}` + "\n"),
		"reference_pack_refs.json": []byte("[]\n"),
	}
	first, err := BuildBundleArchive(ManifestInput{
		BundleID:             "22222222-2222-2222-2222-222222222222",
		IncidentID:           "11111111-1111-1111-1111-111111111111",
		IncidentKey:          "INC-1",
		ExportedAt:           "2026-05-25T00:00:00Z",
		ReferencePackMode:    ReferencePackModeRefsOnly,
		OptionalSections:     []string{},
		RequiredCapabilities: []string{},
	}, files)
	if err != nil {
		t.Fatalf("BuildBundleArchive first: %v", err)
	}
	second, err := BuildBundleArchive(ManifestInput{
		BundleID:             "22222222-2222-2222-2222-222222222222",
		IncidentID:           "11111111-1111-1111-1111-111111111111",
		IncidentKey:          "INC-1",
		ExportedAt:           "2026-05-25T00:00:00Z",
		ReferencePackMode:    ReferencePackModeRefsOnly,
		OptionalSections:     []string{},
		RequiredCapabilities: []string{},
	}, map[string][]byte{
		"reference_pack_refs.json": []byte("[]\n"),
		"data/incident.json":       []byte(`{"id":"inc"}` + "\n"),
		"records.ndjson":           []byte("{}\n"),
	})
	if err != nil {
		t.Fatalf("BuildBundleArchive second: %v", err)
	}
	if !bytes.Equal(first.Bytes, second.Bytes) {
		t.Fatal("deterministic bundle archive changed when input map order changed")
	}
	if first.ManifestSHA256 == "" || len(first.ChecksumLines) == 0 {
		t.Fatalf("bundle result must expose manifest hash and checksums: %#v", first)
	}
}

func TestPhase11_U_11_INCIDENT_BUNDLES_03_VerifyBundleRejectsUnsafeAndCapabilityFailures(t *testing.T) {
	unsafe := newZip(t, map[string][]byte{"../manifest.json": []byte("{}")})
	_, err := VerifyBundle(VerificationInput{
		Bundle: unsafe,
		Limits: config.LimitConfig{Archives: config.ArchiveLimits{MaxMembers: 100, MaxCompressionRatio: 100}, IncidentBundles: config.IncidentBundleLimits{MaxExtractedBytes: 1024}},
	})
	if !isVerificationReason(err, "invalid_member_path") {
		t.Fatalf("unsafe path reason mismatch: %v", err)
	}

	files := map[string][]byte{
		"data/incident.json":       []byte(`{"id":"11111111-1111-1111-1111-111111111111"}` + "\n"),
		"records.ndjson":           []byte(""),
		"reference_pack_refs.json": []byte("[]\n"),
	}
	bundle, err := BuildBundleArchive(ManifestInput{
		BundleID:             "22222222-2222-2222-2222-222222222222",
		IncidentID:           "11111111-1111-1111-1111-111111111111",
		IncidentKey:          "INC-1",
		ExportedAt:           "2026-05-25T00:00:00Z",
		ReferencePackMode:    ReferencePackModeRefsOnly,
		RequiredCapabilities: []string{"snapshots"},
	}, files)
	if err != nil {
		t.Fatalf("BuildBundleArchive: %v", err)
	}
	_, err = VerifyBundle(VerificationInput{
		Bundle: bundle.Bytes,
		Limits: config.LimitConfig{Archives: config.ArchiveLimits{MaxMembers: 100, MaxCompressionRatio: 100}, IncidentBundles: config.IncidentBundleLimits{MaxExtractedBytes: 1024 * 1024}},
	})
	if !isVerificationReason(err, "unsupported_required_capability") {
		t.Fatalf("required capability reason mismatch: %v", err)
	}
}

func TestPhase11_U_11_INCIDENT_BUNDLES_04_OpenAPIAndErrorRegistryContainIncidentBundleContracts(t *testing.T) {
	openAPI, err := os.ReadFile("../../../contracts/openapi/cartulary.openapi.yaml")
	if err != nil {
		t.Fatalf("read openapi: %v", err)
	}
	for _, needle := range []string{
		"/api/v1/incident-bundles/export",
		"/api/v1/incident-bundles/{bundle_id}",
		"/api/v1/incident-bundles/import",
		"IncidentBundleResource",
	} {
		if !bytes.Contains(openAPI, []byte(needle)) {
			t.Fatalf("openapi missing %q", needle)
		}
	}
	errorsDoc, err := os.ReadFile("../../../contracts/errors/index.json")
	if err != nil {
		t.Fatalf("read errors: %v", err)
	}
	for _, needle := range []string{
		`"code": "invalid_incident_bundle_request"`,
		`"code": "incident_bundle_not_found"`,
		`"code": "incident_bundle_export_rejected"`,
		`"code": "incident_bundle_import_rejected"`,
		`"error_code": "incident_bundle_import_rejected"`,
		`"code": "checksum_mismatch"`,
		`"code": "unsupported_required_capability"`,
	} {
		if !bytes.Contains(errorsDoc, []byte(needle)) {
			t.Fatalf("errors registry missing %q", needle)
		}
	}
}

func newZip(t testing.TB, files map[string][]byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for path, data := range files {
		w, err := zw.Create(path)
		if err != nil {
			t.Fatalf("create zip member: %v", err)
		}
		if _, err := w.Write(data); err != nil {
			t.Fatalf("write zip member: %v", err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
	return buf.Bytes()
}

func isVerificationReason(err error, reason string) bool {
	if err == nil {
		return false
	}
	verificationErr, ok := err.(*VerificationError)
	return ok && verificationErr.ReasonCode == reason
}

func TestPhase11_U_11_INCIDENT_BUNDLES_01_DecodeImportMetadataUsesUploadHash(t *testing.T) {
	request, apiErr := DecodeImportMetadata(httpapi.UploadEnvelope{
		Metadata: map[string]json.RawMessage{
			"client_txn_id": json.RawMessage(`"txn-import"`),
		},
		FileSHA256Hex: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	})
	if apiErr != nil {
		t.Fatalf("DecodeImportMetadata returned error: %#v", apiErr)
	}
	var normalized map[string]any
	if err := json.Unmarshal(request.Normalized, &normalized); err != nil {
		t.Fatalf("decode normalized import request: %v", err)
	}
	if normalized["file_sha256"] != "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" {
		t.Fatalf("normalized import request must include file hash: %#v", normalized)
	}
	for _, field := range []string{"clone_mode", "merge_mode", "identifier_remap", "remote_fetch"} {
		_, apiErr := DecodeImportMetadata(httpapi.UploadEnvelope{
			Metadata: map[string]json.RawMessage{
				"client_txn_id": json.RawMessage(`"txn-import"`),
				field:           json.RawMessage(`true`),
			},
			FileSHA256Hex: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
		})
		if apiErr == nil || apiErr.Code != "invalid_incident_bundle_request" {
			t.Fatalf("%s must be rejected with invalid_incident_bundle_request, got %#v", field, apiErr)
		}
	}
}
