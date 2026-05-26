package incidentbundles

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"os"
	"slices"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"gopkg.in/yaml.v3"
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
		"data/records.ndjson":           []byte("{}\n"),
		"data/incident.json":            []byte(`{"id":"inc"}` + "\n"),
		"data/reference_pack_refs.json": []byte("[]\n"),
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
		"data/reference_pack_refs.json": []byte("[]\n"),
		"data/incident.json":            []byte(`{"id":"inc"}` + "\n"),
		"data/records.ndjson":           []byte("{}\n"),
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
	if first.Manifest.BundleVersion != 1 {
		t.Fatalf("manifest bundle_version must be numeric 1, got %#v", first.Manifest.BundleVersion)
	}
	if first.Manifest.SourceChangeSetHighWatermark == "" {
		t.Fatalf("manifest must expose source_change_set_high_watermark: %#v", first.Manifest)
	}
	for _, file := range first.Manifest.Files {
		if file.Path == "data/records.ndjson" && file.SizeBytes != int64(len(files["data/records.ndjson"])) {
			t.Fatalf("manifest file must use exact size_bytes: %#v", file)
		}
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
		"data/incident.json":            []byte(`{"id":"11111111-1111-1111-1111-111111111111"}` + "\n"),
		"data/records.ndjson":           []byte(""),
		"data/reference_pack_refs.json": []byte("[]\n"),
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

	validBundle, err := BuildBundleArchive(ManifestInput{
		BundleID:          "33333333-3333-3333-3333-333333333333",
		IncidentID:        "11111111-1111-1111-1111-111111111111",
		IncidentKey:       "INC-1",
		ExportedAt:        "2026-05-25T00:00:00Z",
		ReferencePackMode: ReferencePackModeRefsOnly,
	}, files)
	if err != nil {
		t.Fatalf("BuildBundleArchive valid bundle: %v", err)
	}
	missingRequired := removeZipMember(t, validBundle.Bytes, "data/records.ndjson")
	_, err = VerifyBundle(VerificationInput{
		Bundle: missingRequired,
		Limits: config.LimitConfig{Archives: config.ArchiveLimits{MaxMembers: 100, MaxCompressionRatio: 100}, IncidentBundles: config.IncidentBundleLimits{MaxExtractedBytes: 1024 * 1024}},
	})
	if !isVerificationReason(err, "missing_required_file") {
		t.Fatalf("missing required file reason mismatch: %v", err)
	}

	withSignature := appendZipMember(t, validBundle.Bytes, "integrity/signature.ed25519", []byte("not-a-supported-signature"))
	_, err = VerifyBundle(VerificationInput{
		Bundle: withSignature,
		Limits: config.LimitConfig{Archives: config.ArchiveLimits{MaxMembers: 100}, IncidentBundles: config.IncidentBundleLimits{MaxExtractedBytes: 1024 * 1024}},
	})
	if !isVerificationReason(err, "signature_mismatch") {
		t.Fatalf("signature reason mismatch: %v", err)
	}

	malformedMode := replaceManifestFields(t, validBundle.Bytes, func(manifest map[string]any) {
		manifest["reference_pack_mode"] = "floating"
	})
	_, err = VerifyBundle(VerificationInput{
		Bundle: malformedMode,
		Limits: config.LimitConfig{Archives: config.ArchiveLimits{MaxMembers: 100, MaxCompressionRatio: 100}, IncidentBundles: config.IncidentBundleLimits{MaxExtractedBytes: 1024 * 1024}},
	})
	if !isVerificationReason(err, "malformed_manifest") {
		t.Fatalf("invalid reference_pack_mode reason mismatch: %v", err)
	}

	malformedOptionalSection := replaceManifestFields(t, validBundle.Bytes, func(manifest map[string]any) {
		manifest["optional_sections"] = []any{"unknown_section"}
	})
	_, err = VerifyBundle(VerificationInput{
		Bundle: malformedOptionalSection,
		Limits: config.LimitConfig{Archives: config.ArchiveLimits{MaxMembers: 100, MaxCompressionRatio: 100}, IncidentBundles: config.IncidentBundleLimits{MaxExtractedBytes: 1024 * 1024}},
	})
	if !isVerificationReason(err, "malformed_manifest") {
		t.Fatalf("unknown optional_sections token reason mismatch: %v", err)
	}

	unsupportedRequired := replaceManifestFields(t, validBundle.Bytes, func(manifest map[string]any) {
		manifest["required_capabilities"] = []any{"snapshots"}
	})
	_, err = VerifyBundle(VerificationInput{
		Bundle: unsupportedRequired,
		Limits: config.LimitConfig{Archives: config.ArchiveLimits{MaxMembers: 100, MaxCompressionRatio: 100}, IncidentBundles: config.IncidentBundleLimits{MaxExtractedBytes: 1024 * 1024}},
	})
	if !isVerificationReason(err, "unsupported_required_capability") {
		t.Fatalf("unsupported required capability reason mismatch: %v", err)
	}

	withKnownOptionalSection, err := BuildBundleArchive(ManifestInput{
		BundleID:             "44444444-4444-4444-4444-444444444444",
		IncidentID:           "11111111-1111-1111-1111-111111111111",
		IncidentKey:          "INC-1",
		ExportedAt:           "2026-05-25T00:00:00Z",
		ReferencePackMode:    ReferencePackModeRefsOnly,
		OptionalSections:     []string{"snapshots"},
		RequiredCapabilities: []string{},
	}, map[string][]byte{
		"data/incident.json":            []byte(`{"id":"11111111-1111-1111-1111-111111111111"}` + "\n"),
		"data/records.ndjson":           []byte(""),
		"data/reference_pack_refs.json": []byte("[]\n"),
		"ext/snapshots/snapshot.json":   []byte(`{"snapshot_id":"snap-1"}` + "\n"),
	})
	if err != nil {
		t.Fatalf("BuildBundleArchive optional section: %v", err)
	}
	if _, err = VerifyBundle(VerificationInput{
		Bundle: withKnownOptionalSection.Bytes,
		Limits: config.LimitConfig{Archives: config.ArchiveLimits{MaxMembers: 100, MaxCompressionRatio: 100}, IncidentBundles: config.IncidentBundleLimits{MaxExtractedBytes: 1024 * 1024}},
	}); err != nil {
		t.Fatalf("known unsupported optional embedded section must not block core verification: %v", err)
	}
}

func TestPhase11_U_11_INCIDENT_BUNDLES_03_CompressionRatioBoundaryUsesThresholdComparison(t *testing.T) {
	limits := config.LimitConfig{Archives: config.ArchiveLimits{MaxCompressionRatio: 3}}
	if err := checkCompressionRatio(300, 100, limits); err != nil {
		t.Fatalf("exact compression ratio limit must pass: %v", err)
	}
	err := checkCompressionRatio(301, 100, limits)
	if !isVerificationReason(err, "archive_compression_ratio_exceeded") {
		t.Fatalf("one byte over compression ratio limit reason mismatch: %v", err)
	}
}

func TestPhase11_U_11_INCIDENT_BUNDLES_04_OpenAPIAndErrorRegistryContainIncidentBundleContracts(t *testing.T) {
	openAPI, err := os.ReadFile("../../../contracts/openapi/cartulary.openapi.yaml")
	if err != nil {
		t.Fatalf("read openapi: %v", err)
	}
	var openAPIDoc map[string]any
	if err := yaml.Unmarshal(openAPI, &openAPIDoc); err != nil {
		t.Fatalf("decode openapi yaml: %v", err)
	}
	for _, field := range []string{"optional_sections", "required_capabilities"} {
		schema := openAPIObjectAt(t, openAPIDoc, "components", "schemas", "IncidentBundleResource", "properties", field)
		if schema["type"] != "array" || schema["uniqueItems"] != true {
			t.Fatalf("%s must be a unique array schema: %#v", field, schema)
		}
		items, ok := schema["items"].(map[string]any)
		if !ok {
			t.Fatalf("%s items schema missing: %#v", field, schema)
		}
		enumValues, ok := items["enum"].([]any)
		if !ok || !slices.Equal(enumStrings(enumValues), []string{"reference_packs", "snapshots"}) {
			t.Fatalf("%s enum tokens mismatch: %#v", field, items["enum"])
		}
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
		`"code": "invalid_value"`,
		`"code": "checksum_mismatch"`,
		`"code": "missing_required_blob"`,
		`"code": "malformed_manifest"`,
		`"code": "unsupported_required_capability"`,
	} {
		if !bytes.Contains(errorsDoc, []byte(needle)) {
			t.Fatalf("errors registry missing %q", needle)
		}
	}
}

func openAPIObjectAt(t testing.TB, root map[string]any, path ...string) map[string]any {
	t.Helper()
	current := any(root)
	for _, key := range path {
		object, ok := current.(map[string]any)
		if !ok {
			t.Fatalf("openapi path %v reached non-object %#v", path, current)
		}
		current, ok = object[key]
		if !ok {
			t.Fatalf("openapi path %v missing key %q", path, key)
		}
	}
	object, ok := current.(map[string]any)
	if !ok {
		t.Fatalf("openapi path %v reached non-object %#v", path, current)
	}
	return object
}

func enumStrings(values []any) []string {
	result := make([]string, 0, len(values))
	for _, value := range values {
		text, ok := value.(string)
		if !ok {
			return nil
		}
		result = append(result, text)
	}
	return result
}

func replaceManifestFields(t testing.TB, bundle []byte, update func(map[string]any)) []byte {
	t.Helper()
	files := zipFilesMap(t, bundle)
	var manifest map[string]any
	if err := json.Unmarshal(files["manifest.json"], &manifest); err != nil {
		t.Fatalf("decode manifest: %v", err)
	}
	update(manifest)
	manifestBytes, err := json.Marshal(manifest)
	if err != nil {
		t.Fatalf("encode manifest: %v", err)
	}
	files["manifest.json"] = manifestBytes
	return zipFromFiles(t, files)
}

func zipFilesMap(t testing.TB, bundle []byte) map[string][]byte {
	t.Helper()
	reader, err := zip.NewReader(bytes.NewReader(bundle), int64(len(bundle)))
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}
	files := map[string][]byte{}
	for _, member := range reader.File {
		rc, err := member.Open()
		if err != nil {
			t.Fatalf("open member %s: %v", member.Name, err)
		}
		data, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			t.Fatalf("read member %s: %v", member.Name, err)
		}
		files[member.Name] = data
	}
	return files
}

func zipFromFiles(t testing.TB, files map[string][]byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	paths := make([]string, 0, len(files))
	for path := range files {
		paths = append(paths, path)
	}
	slices.Sort(paths)
	for _, path := range paths {
		w, err := zw.Create(path)
		if err != nil {
			t.Fatalf("create member %s: %v", path, err)
		}
		if _, err := w.Write(files[path]); err != nil {
			t.Fatalf("write member %s: %v", path, err)
		}
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
	return buf.Bytes()
}

func removeZipMember(t testing.TB, bundle []byte, memberPath string) []byte {
	t.Helper()
	reader, err := zip.NewReader(bytes.NewReader(bundle), int64(len(bundle)))
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	found := false
	for _, member := range reader.File {
		if member.Name == memberPath {
			found = true
			continue
		}
		rc, err := member.Open()
		if err != nil {
			t.Fatalf("open member %s: %v", member.Name, err)
		}
		data, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			t.Fatalf("read member %s: %v", member.Name, err)
		}
		w, err := zw.Create(member.Name)
		if err != nil {
			t.Fatalf("create member %s: %v", member.Name, err)
		}
		if _, err := w.Write(data); err != nil {
			t.Fatalf("write member %s: %v", member.Name, err)
		}
	}
	if !found {
		t.Fatalf("zip member %s not found", memberPath)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
	return buf.Bytes()
}

func appendZipMember(t testing.TB, bundle []byte, memberPath string, payload []byte) []byte {
	t.Helper()
	reader, err := zip.NewReader(bytes.NewReader(bundle), int64(len(bundle)))
	if err != nil {
		t.Fatalf("open zip: %v", err)
	}
	var buf bytes.Buffer
	zw := zip.NewWriter(&buf)
	for _, member := range reader.File {
		rc, err := member.Open()
		if err != nil {
			t.Fatalf("open member %s: %v", member.Name, err)
		}
		data, err := io.ReadAll(rc)
		_ = rc.Close()
		if err != nil {
			t.Fatalf("read member %s: %v", member.Name, err)
		}
		w, err := zw.Create(member.Name)
		if err != nil {
			t.Fatalf("create member %s: %v", member.Name, err)
		}
		if _, err := w.Write(data); err != nil {
			t.Fatalf("write member %s: %v", member.Name, err)
		}
	}
	w, err := zw.Create(memberPath)
	if err != nil {
		t.Fatalf("create appended member %s: %v", memberPath, err)
	}
	if _, err := w.Write(payload); err != nil {
		t.Fatalf("write appended member %s: %v", memberPath, err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
	return buf.Bytes()
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
