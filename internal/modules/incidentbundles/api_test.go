package incidentbundles

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"os"
	"slices"
	"strings"
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
	requireClosedRequiredSourceFileRegistry(t)

	files := minimalRequiredBundleFiles()
	files["data/records.ndjson"] = []byte("{}\n")
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
	secondFiles := minimalRequiredBundleFiles()
	secondFiles["data/records.ndjson"] = []byte("{}\n")
	second, err := BuildBundleArchive(ManifestInput{
		BundleID:             "22222222-2222-2222-2222-222222222222",
		IncidentID:           "11111111-1111-1111-1111-111111111111",
		IncidentKey:          "INC-1",
		ExportedAt:           "2026-05-25T00:00:00Z",
		ReferencePackMode:    ReferencePackModeRefsOnly,
		OptionalSections:     []string{},
		RequiredCapabilities: []string{},
	}, secondFiles)
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

func requireClosedRequiredSourceFileRegistry(t testing.TB) {
	t.Helper()

	want := []string{
		"data/incident.json",
		"data/actors.ndjson",
		"data/records.ndjson",
		"data/timeline_time_conversion_profiles.ndjson",
		"data/timeline_events.ndjson",
		"data/parties.ndjson",
		"data/entity_mentions.ndjson",
		"data/hosts.ndjson",
		"data/identities.ndjson",
		"data/entity_preserved_identifiers.ndjson",
		"data/entity_aliases.ndjson",
		"data/indicators.ndjson",
		"data/indicator_observations.ndjson",
		"data/indicator_state_intervals.ndjson",
		"data/artifacts.ndjson",
		"data/artifact_findings.ndjson",
		"data/artifact_investigative_queries.ndjson",
		"data/artifact_forensic_keywords.ndjson",
		"data/handoff_risk_refs.ndjson",
		"data/task_requests.ndjson",
		"data/decisions.ndjson",
		"data/evidence_records.ndjson",
		"data/evidence_custody_events.ndjson",
		"data/object_blobs.ndjson",
		"data/compromise_assessments.ndjson",
		"data/record_links.ndjson",
		"data/tags.ndjson",
		"data/record_tags.ndjson",
		"data/change_sets.ndjson",
		"data/change_set_mutations.ndjson",
		"data/record_revisions.ndjson",
		"data/saved_views.ndjson",
		"data/reference_pack_refs.json",
	}
	if !slices.Equal(requiredStructuredFiles, want) {
		t.Fatalf("required source-file registry drifted:\n got %#v\nwant %#v", requiredStructuredFiles, want)
	}

	files := minimalRequiredBundleFiles()
	delete(files, "data/parties.ndjson")
	_, err := BuildBundleArchive(ManifestInput{
		BundleID:          "55555555-5555-5555-5555-555555555555",
		IncidentID:        "11111111-1111-1111-1111-111111111111",
		IncidentKey:       "INC-1",
		ExportedAt:        "2026-05-25T00:00:00Z",
		ReferencePackMode: ReferencePackModeRefsOnly,
	}, files)
	if err == nil || !strings.Contains(err.Error(), "data/parties.ndjson is required") {
		t.Fatalf("BuildBundleArchive must reject missing required source files, got %v", err)
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

	files := minimalRequiredBundleFiles()
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

	withZipDirectories := appendZipMember(t, appendZipMember(t, validBundle.Bytes, "data/", nil), "integrity/", nil)
	if _, err = VerifyBundle(VerificationInput{
		Bundle: withZipDirectories,
		Limits: config.LimitConfig{Archives: config.ArchiveLimits{MaxMembers: 100, MaxCompressionRatio: 100}, IncidentBundles: config.IncidentBundleLimits{MaxExtractedBytes: 1024 * 1024}},
	}); err != nil {
		t.Fatalf("safe ZIP directory members must be ignored as structural entries: %v", err)
	}
	_, err = VerifyBundle(VerificationInput{
		Bundle: withZipDirectories,
		Limits: config.LimitConfig{Archives: config.ArchiveLimits{MaxMembers: int64(len(zipFilesMap(t, validBundle.Bytes)) + 1), MaxCompressionRatio: 100}, IncidentBundles: config.IncidentBundleLimits{MaxExtractedBytes: 1024 * 1024}},
	})
	if !isVerificationReason(err, "archive_member_count_exceeded") {
		t.Fatalf("ZIP directory entries must count against member limits: %v", err)
	}
	unsafeDirectory := appendZipMember(t, validBundle.Bytes, "../data/", nil)
	_, err = VerifyBundle(VerificationInput{
		Bundle: unsafeDirectory,
		Limits: config.LimitConfig{Archives: config.ArchiveLimits{MaxMembers: 100, MaxCompressionRatio: 100}, IncidentBundles: config.IncidentBundleLimits{MaxExtractedBytes: 1024 * 1024}},
	})
	if !isVerificationReason(err, "invalid_member_path") {
		t.Fatalf("unsafe ZIP directory reason mismatch: %v", err)
	}
	unsupportedZipMember := appendZipSymlink(t, validBundle.Bytes, "data/member-link", "manifest.json")
	_, err = VerifyBundle(VerificationInput{
		Bundle: unsupportedZipMember,
		Limits: config.LimitConfig{Archives: config.ArchiveLimits{MaxMembers: 100, MaxCompressionRatio: 100}, IncidentBundles: config.IncidentBundleLimits{MaxExtractedBytes: 1024 * 1024}},
	})
	if !isVerificationReason(err, "unsupported_member_type") {
		t.Fatalf("unsupported ZIP member reason mismatch: %v", err)
	}
	withTarDirectories := newTarGzip(t, []string{"data", "integrity/"}, zipFilesMap(t, validBundle.Bytes))
	if _, err = VerifyBundle(VerificationInput{
		Bundle: withTarDirectories,
		Limits: config.LimitConfig{Archives: config.ArchiveLimits{MaxMembers: 100, MaxCompressionRatio: 100}, IncidentBundles: config.IncidentBundleLimits{MaxExtractedBytes: 1024 * 1024}},
	}); err != nil {
		t.Fatalf("safe TAR directory members must be ignored as structural entries: %v", err)
	}
	_, err = VerifyBundle(VerificationInput{
		Bundle: withTarDirectories,
		Limits: config.LimitConfig{Archives: config.ArchiveLimits{MaxMembers: int64(len(zipFilesMap(t, validBundle.Bytes)) + 1), MaxCompressionRatio: 100}, IncidentBundles: config.IncidentBundleLimits{MaxExtractedBytes: 1024 * 1024}},
	})
	if !isVerificationReason(err, "archive_member_count_exceeded") {
		t.Fatalf("TAR directory entries must count against member limits: %v", err)
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
	}, withAdditionalBundleFiles(minimalRequiredBundleFiles(), map[string][]byte{
		"ext/snapshots/snapshot.json": []byte(`{"snapshot_id":"snap-1"}` + "\n"),
	}))
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

func minimalRequiredBundleFiles() map[string][]byte {
	files := make(map[string][]byte, len(requiredStructuredFiles))
	for _, path := range requiredStructuredFiles {
		switch path {
		case "data/incident.json":
			files[path] = []byte(`{"id":"11111111-1111-1111-1111-111111111111"}` + "\n")
		case "data/reference_pack_refs.json":
			files[path] = []byte("[]\n")
		default:
			files[path] = []byte{}
		}
	}
	return files
}

func withAdditionalBundleFiles(files map[string][]byte, additional map[string][]byte) map[string][]byte {
	for path, payload := range additional {
		files[path] = payload
	}
	return files
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

func TestPhase11_U_11_INCIDENT_BUNDLES_05_ErrorRegistryUsesExactClosedIncidentBundleSets(t *testing.T) {
	errorsDoc, err := os.ReadFile("../../../contracts/errors/index.json")
	if err != nil {
		t.Fatalf("read errors: %v", err)
	}
	var registry struct {
		Errors []struct {
			Code string `json:"code"`
		} `json:"errors"`
		ReasonRegistries []struct {
			ErrorCode   string `json:"error_code"`
			ReasonCodes []struct {
				Code string `json:"code"`
			} `json:"reason_codes"`
		} `json:"reason_registries"`
	}
	if err := json.Unmarshal(errorsDoc, &registry); err != nil {
		t.Fatalf("decode errors registry: %v", err)
	}

	var incidentBundleCodes []string
	for _, entry := range registry.Errors {
		if strings.Contains(entry.Code, "incident_bundle") {
			incidentBundleCodes = append(incidentBundleCodes, entry.Code)
		}
	}
	slices.Sort(incidentBundleCodes)
	wantTopLevel := []string{
		"incident_bundle_export_rejected",
		"incident_bundle_import_rejected",
		"incident_bundle_not_found",
		"invalid_incident_bundle_request",
	}
	if !slices.Equal(incidentBundleCodes, wantTopLevel) {
		t.Fatalf("incident-bundle top-level errors changed: got %#v want %#v", incidentBundleCodes, wantTopLevel)
	}

	wantReasons := map[string][]string{
		"invalid_incident_bundle_request": {
			"blob_mode_not_supported",
			"duplicate_part",
			"field_not_nullable",
			"history_mode_not_supported",
			"invalid_metadata_encoding",
			"invalid_optional_sections",
			"invalid_part_content_type",
			"invalid_reference_pack_mode",
			"invalid_required_capabilities",
			"invalid_value",
			"malformed_metadata_json",
			"missing_required_field",
			"missing_required_part",
			"request_not_object",
			"unexpected_part",
			"unknown_field",
			"unsupported_upload_envelope",
		},
		"incident_bundle_export_rejected": {
			"missing_required_blob",
			"missing_required_file",
		},
		"incident_bundle_import_rejected": {
			"archive_compression_ratio_exceeded",
			"archive_extracted_bytes_exceeded",
			"archive_member_count_exceeded",
			"blob_hash_mismatch",
			"checksum_mismatch",
			"duplicate_incident_id",
			"invalid_member_path",
			"malformed_manifest",
			"missing_required_blob",
			"missing_required_file",
			"remote_fetch_required",
			"signature_mismatch",
			"unsupported_member_type",
			"unsupported_required_capability",
		},
	}
	gotReasons := map[string][]string{}
	for _, entry := range registry.ReasonRegistries {
		if _, ok := wantReasons[entry.ErrorCode]; !ok {
			continue
		}
		for _, reason := range entry.ReasonCodes {
			gotReasons[entry.ErrorCode] = append(gotReasons[entry.ErrorCode], reason.Code)
		}
		slices.Sort(gotReasons[entry.ErrorCode])
	}
	for errorCode, want := range wantReasons {
		if !slices.Equal(gotReasons[errorCode], want) {
			t.Fatalf("%s reason registry changed: got %#v want %#v", errorCode, gotReasons[errorCode], want)
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

func appendZipSymlink(t testing.TB, bundle []byte, memberPath string, target string) []byte {
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
	header := &zip.FileHeader{Name: memberPath}
	header.SetMode(os.ModeSymlink | 0o777)
	w, err := zw.CreateHeader(header)
	if err != nil {
		t.Fatalf("create symlink member %s: %v", memberPath, err)
	}
	if _, err := w.Write([]byte(target)); err != nil {
		t.Fatalf("write symlink member %s: %v", memberPath, err)
	}
	if err := zw.Close(); err != nil {
		t.Fatalf("close zip: %v", err)
	}
	return buf.Bytes()
}

func newTarGzip(t testing.TB, directories []string, files map[string][]byte) []byte {
	t.Helper()
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	tw := tar.NewWriter(gz)
	for _, dir := range directories {
		if err := tw.WriteHeader(&tar.Header{Name: dir, Typeflag: tar.TypeDir, Mode: 0o700}); err != nil {
			t.Fatalf("write tar directory %s: %v", dir, err)
		}
	}
	paths := make([]string, 0, len(files))
	for path := range files {
		paths = append(paths, path)
	}
	slices.Sort(paths)
	for _, path := range paths {
		data := files[path]
		if err := tw.WriteHeader(&tar.Header{Name: path, Typeflag: tar.TypeReg, Mode: 0o600, Size: int64(len(data))}); err != nil {
			t.Fatalf("write tar header %s: %v", path, err)
		}
		if _, err := tw.Write(data); err != nil {
			t.Fatalf("write tar member %s: %v", path, err)
		}
	}
	if err := tw.Close(); err != nil {
		t.Fatalf("close tar: %v", err)
	}
	if err := gz.Close(); err != nil {
		t.Fatalf("close gzip: %v", err)
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
