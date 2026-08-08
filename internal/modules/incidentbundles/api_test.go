package incidentbundles

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/crossownertransaction"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidents"
	"github.com/JochiRaider/cartulary/internal/platform/contracttest"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/JochiRaider/cartulary/internal/testutil/httpapiextensions"
)

func TestDecodeExportRequestCanonicalizesAndRejectsModes_Unit(t *testing.T) {
	request, apiErr := DecodeExportRequest(bytes.NewBufferString(`{
		"incident_id":"11111111-1111-1111-1111-111111111111",
		"client_txn_id":"txn-export",
		"optional_sections":["reference_packs","snapshots","reference_packs"],
		"required_capabilities":[],
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
	if len(request.RequiredCapabilities) != 0 {
		t.Fatalf("required capabilities not canonicalized: %#v", request.RequiredCapabilities)
	}
	var normalized map[string]any
	if err := json.Unmarshal(request.Normalized, &normalized); err != nil {
		t.Fatalf("decode normalized request: %v", err)
	}
	if normalized["history_mode"] != "full" || normalized["blob_mode"] != "full" {
		t.Fatalf("normalized request must include fixed modes: %#v", normalized)
	}

	_, apiErr = DecodeExportRequest(bytes.NewBufferString(`{
		"incident_id":"11111111-1111-1111-1111-111111111111",
		"client_txn_id":"txn-export-required-capability",
		"required_capabilities":["snapshots"]
	}`))
	if apiErr == nil {
		t.Fatal("required optional-section capabilities must be rejected until implemented")
	}
	if apiErr.Status != http.StatusBadRequest || apiErr.Code != "invalid_incident_bundle_request" || apiErr.Details["reason_code"] != "invalid_required_capabilities" {
		t.Fatalf("required capability rejection mismatch: %#v", apiErr)
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

func TestBundleManifestChecksumDeterministic_Unit(t *testing.T) {
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
	if first.Manifest.BundleVersion != 2 {
		t.Fatalf("manifest bundle_version must be numeric 2, got %#v", first.Manifest.BundleVersion)
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
		"data/timeline_time_profiles.ndjson",
		"data/timeline_records.ndjson",
		"data/timeline_source_provenance.ndjson",
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
	files = minimalRequiredBundleFiles()
	delete(files, "data/saved_views.ndjson")
	_, err = BuildBundleArchive(ManifestInput{
		BundleID:          "55555555-5555-5555-5555-555555555556",
		IncidentID:        "11111111-1111-1111-1111-111111111111",
		IncidentKey:       "INC-1",
		ExportedAt:        "2026-05-25T00:00:00Z",
		ReferencePackMode: ReferencePackModeRefsOnly,
	}, files)
	if err == nil || !strings.Contains(err.Error(), "data/saved_views.ndjson is required") {
		t.Fatalf("BuildBundleArchive must reject missing saved views source file, got %v", err)
	}
}

func TestVerifyBundleRejectsUnsafeAndCapabilityFailures_Unit(t *testing.T) {
	unsafe := newZip(t, map[string][]byte{"../manifest.json": []byte("{}")})
	_, err := VerifyBundle(VerificationInput{
		Bundle: unsafe,
		Limits: Limits{Archives: ArchiveLimits{MaxMembers: 100, MaxCompressionRatio: 100}, IncidentBundles: IncidentBundleLimits{MaxExtractedBytes: 1024}},
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
		Limits: Limits{Archives: ArchiveLimits{MaxMembers: 100, MaxCompressionRatio: 100}, IncidentBundles: IncidentBundleLimits{MaxExtractedBytes: 1024 * 1024}},
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
		Limits: Limits{Archives: ArchiveLimits{MaxMembers: 100, MaxCompressionRatio: 100}, IncidentBundles: IncidentBundleLimits{MaxExtractedBytes: 1024 * 1024}},
	})
	if !isVerificationReason(err, "missing_required_file") {
		t.Fatalf("missing required file reason mismatch: %v", err)
	}

	withZipDirectories := appendZipMember(t, appendZipMember(t, validBundle.Bytes, "data/", nil), "integrity/", nil)
	if _, err = VerifyBundle(VerificationInput{
		Bundle: withZipDirectories,
		Limits: Limits{Archives: ArchiveLimits{MaxMembers: 100, MaxCompressionRatio: 100}, IncidentBundles: IncidentBundleLimits{MaxExtractedBytes: 1024 * 1024}},
	}); err != nil {
		t.Fatalf("safe ZIP directory members must be ignored as structural entries: %v", err)
	}
	_, err = VerifyBundle(VerificationInput{
		Bundle: withZipDirectories,
		Limits: Limits{Archives: ArchiveLimits{MaxMembers: int64(len(zipFilesMap(t, validBundle.Bytes)) + 1), MaxCompressionRatio: 100}, IncidentBundles: IncidentBundleLimits{MaxExtractedBytes: 1024 * 1024}},
	})
	if !isVerificationReason(err, "archive_member_count_exceeded") {
		t.Fatalf("ZIP directory entries must count against member limits: %v", err)
	}
	unsafeDirectory := appendZipMember(t, validBundle.Bytes, "../data/", nil)
	_, err = VerifyBundle(VerificationInput{
		Bundle: unsafeDirectory,
		Limits: Limits{Archives: ArchiveLimits{MaxMembers: 100, MaxCompressionRatio: 100}, IncidentBundles: IncidentBundleLimits{MaxExtractedBytes: 1024 * 1024}},
	})
	if !isVerificationReason(err, "invalid_member_path") {
		t.Fatalf("unsafe ZIP directory reason mismatch: %v", err)
	}
	unsupportedZipMember := appendZipSymlink(t, validBundle.Bytes, "data/member-link", "manifest.json")
	_, err = VerifyBundle(VerificationInput{
		Bundle: unsupportedZipMember,
		Limits: Limits{Archives: ArchiveLimits{MaxMembers: 100, MaxCompressionRatio: 100}, IncidentBundles: IncidentBundleLimits{MaxExtractedBytes: 1024 * 1024}},
	})
	if !isVerificationReason(err, "unsupported_member_type") {
		t.Fatalf("unsupported ZIP member reason mismatch: %v", err)
	}
	withTarDirectories := newTarGzip(t, []string{"data", "integrity/"}, zipFilesMap(t, validBundle.Bytes))
	if _, err = VerifyBundle(VerificationInput{
		Bundle: withTarDirectories,
		Limits: Limits{Archives: ArchiveLimits{MaxMembers: 100, MaxCompressionRatio: 100}, IncidentBundles: IncidentBundleLimits{MaxExtractedBytes: 1024 * 1024}},
	}); err != nil {
		t.Fatalf("safe TAR directory members must be ignored as structural entries: %v", err)
	}
	_, err = VerifyBundle(VerificationInput{
		Bundle: withTarDirectories,
		Limits: Limits{Archives: ArchiveLimits{MaxMembers: int64(len(zipFilesMap(t, validBundle.Bytes)) + 1), MaxCompressionRatio: 100}, IncidentBundles: IncidentBundleLimits{MaxExtractedBytes: 1024 * 1024}},
	})
	if !isVerificationReason(err, "archive_member_count_exceeded") {
		t.Fatalf("TAR directory entries must count against member limits: %v", err)
	}

	withSignature := appendZipMember(t, validBundle.Bytes, "integrity/signature.ed25519", []byte("not-a-supported-signature"))
	_, err = VerifyBundle(VerificationInput{
		Bundle: withSignature,
		Limits: Limits{Archives: ArchiveLimits{MaxMembers: 100}, IncidentBundles: IncidentBundleLimits{MaxExtractedBytes: 1024 * 1024}},
	})
	if !isVerificationReason(err, "signature_mismatch") {
		t.Fatalf("signature reason mismatch: %v", err)
	}

	malformedMode := replaceManifestFields(t, validBundle.Bytes, func(manifest map[string]any) {
		manifest["reference_pack_mode"] = "floating"
	})
	_, err = VerifyBundle(VerificationInput{
		Bundle: malformedMode,
		Limits: Limits{Archives: ArchiveLimits{MaxMembers: 100, MaxCompressionRatio: 100}, IncidentBundles: IncidentBundleLimits{MaxExtractedBytes: 1024 * 1024}},
	})
	if !isVerificationReason(err, "malformed_manifest") {
		t.Fatalf("invalid reference_pack_mode reason mismatch: %v", err)
	}

	malformedOptionalSection := replaceManifestFields(t, validBundle.Bytes, func(manifest map[string]any) {
		manifest["optional_sections"] = []any{"unknown_section"}
	})
	_, err = VerifyBundle(VerificationInput{
		Bundle: malformedOptionalSection,
		Limits: Limits{Archives: ArchiveLimits{MaxMembers: 100, MaxCompressionRatio: 100}, IncidentBundles: IncidentBundleLimits{MaxExtractedBytes: 1024 * 1024}},
	})
	if !isVerificationReason(err, "malformed_manifest") {
		t.Fatalf("unknown optional_sections token reason mismatch: %v", err)
	}

	unsupportedRequired := replaceManifestFields(t, validBundle.Bytes, func(manifest map[string]any) {
		manifest["required_capabilities"] = []any{"snapshots"}
	})
	_, err = VerifyBundle(VerificationInput{
		Bundle: unsupportedRequired,
		Limits: Limits{Archives: ArchiveLimits{MaxMembers: 100, MaxCompressionRatio: 100}, IncidentBundles: IncidentBundleLimits{MaxExtractedBytes: 1024 * 1024}},
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
		Limits: Limits{Archives: ArchiveLimits{MaxMembers: 100, MaxCompressionRatio: 100}, IncidentBundles: IncidentBundleLimits{MaxExtractedBytes: 1024 * 1024}},
	}); err != nil {
		t.Fatalf("known unsupported optional embedded section must not block core verification: %v", err)
	}
}

func TestVerifyBundleRejectsMalformedManifestVersion_Unit(t *testing.T) {
	files := minimalRequiredBundleFiles()
	bundle, err := BuildBundleArchive(ManifestInput{
		BundleID:          "22222222-2222-4222-8222-222222222222",
		IncidentID:        "11111111-1111-4111-8111-111111111111",
		IncidentKey:       "INC-VERSION-MALFORMED",
		ExportedAt:        "2026-07-27T00:00:00Z",
		ReferencePackMode: ReferencePackModeRefsOnly,
	}, files)
	if err != nil {
		t.Fatalf("BuildBundleArchive: %v", err)
	}
	cases := map[string]func(map[string]any){
		"omitted":     func(manifest map[string]any) { delete(manifest, "bundle_version") },
		"null":        func(manifest map[string]any) { manifest["bundle_version"] = nil },
		"string":      func(manifest map[string]any) { manifest["bundle_version"] = "2" },
		"non_integer": func(manifest map[string]any) { manifest["bundle_version"] = 2.5 },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			_, verifyErr := VerifyBundle(VerificationInput{
				Bundle: replaceManifestFields(t, bundle.Bytes, mutate),
				Limits: Limits{Archives: ArchiveLimits{MaxMembers: 100, MaxCompressionRatio: 100}, IncidentBundles: IncidentBundleLimits{MaxExtractedBytes: 1024 * 1024}},
			})
			if !isVerificationReason(verifyErr, "malformed_manifest") {
				t.Fatalf("version form must fail as malformed_manifest, got %v", verifyErr)
			}
		})
	}
}

func TestVerifyBundleRejectsUnsupportedAndMixedTimelineVersions_Unit(t *testing.T) {
	bundle, err := BuildBundleArchive(ManifestInput{
		BundleID:          "22222222-2222-4222-8222-222222222223",
		IncidentID:        "11111111-1111-4111-8111-111111111111",
		IncidentKey:       "INC-VERSION-CLOSED",
		ExportedAt:        "2026-07-27T00:00:00Z",
		ReferencePackMode: ReferencePackModeRefsOnly,
	}, minimalRequiredBundleFiles())
	if err != nil {
		t.Fatalf("BuildBundleArchive: %v", err)
	}
	unsupported := replaceManifestFields(t, bundle.Bytes, func(manifest map[string]any) {
		manifest["bundle_version"] = 3
	})
	if _, verifyErr := VerifyBundle(VerificationInput{
		Bundle: unsupported,
		Limits: Limits{Archives: ArchiveLimits{MaxMembers: 100, MaxCompressionRatio: 100}, IncidentBundles: IncidentBundleLimits{MaxExtractedBytes: 1024 * 1024}},
	}); !isVerificationReason(verifyErr, "unsupported_bundle_version") {
		t.Fatalf("unsupported integer version reason mismatch: %v", verifyErr)
	}

	mixedFiles := zipFilesMap(t, bundle.Bytes)
	mixedFiles["data/timeline_events.ndjson"] = []byte{}
	if _, verifyErr := VerifyBundle(VerificationInput{
		Bundle: zipFromFiles(t, mixedFiles),
		Limits: Limits{Archives: ArchiveLimits{MaxMembers: 100, MaxCompressionRatio: 100}, IncidentBundles: IncidentBundleLimits{MaxExtractedBytes: 1024 * 1024}},
	}); !isVerificationReason(verifyErr, "malformed_manifest") {
		t.Fatalf("mixed Timeline paths reason mismatch: %v", verifyErr)
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

func TestCompressionRatioBoundaryUsesThresholdComparison_Unit(t *testing.T) {
	limits := Limits{Archives: ArchiveLimits{MaxCompressionRatio: 3}}
	if err := checkCompressionRatio(300, 100, limits); err != nil {
		t.Fatalf("exact compression ratio limit must pass: %v", err)
	}
	err := checkCompressionRatio(301, 100, limits)
	if !isVerificationReason(err, "archive_compression_ratio_exceeded") {
		t.Fatalf("one byte over compression ratio limit reason mismatch: %v", err)
	}
}

func TestOpenAPIAndErrorRegistryContainIncidentBundleContracts_Unit(t *testing.T) {
	openAPIDoc := contracttest.OpenAPIDocument(t)
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
	exportRequiredCapabilities := openAPIObjectAt(t, openAPIDoc, "components", "schemas", "IncidentBundleExportRequest", "properties", "required_capabilities")
	maxItems, ok := exportRequiredCapabilities["maxItems"].(float64)
	if !ok || maxItems != 0 {
		t.Fatalf("export request required_capabilities must advertise empty current support: %#v", exportRequiredCapabilities)
	}
	paths := openAPIObjectAt(t, openAPIDoc, "paths")
	for _, path := range []string{
		"/api/v1/incident-bundles/export",
		"/api/v1/incident-bundles/{bundle_id}",
		"/api/v1/incident-bundles/import",
	} {
		if _, ok := paths[path]; !ok {
			t.Fatalf("openapi missing path %q", path)
		}
	}
	_ = openAPIObjectAt(t, openAPIDoc, "components", "schemas", "IncidentBundleResource")
	for code, status := range map[string]int{
		"incident_bundle_export_rejected": http.StatusConflict,
		"incident_bundle_import_rejected": http.StatusConflict,
		"incident_bundle_not_found":       http.StatusNotFound,
		"invalid_incident_bundle_request": http.StatusBadRequest,
	} {
		contracttest.RequireErrorContract(t, code, status)
	}
}

func TestErrorRegistryUsesExactClosedIncidentBundleSets_Unit(t *testing.T) {
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
	if err := json.Unmarshal([]byte(contracttest.ErrorRegistryArtifactJSON(t)), &registry); err != nil {
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
			"initial_admin_unavailable",
			"invalid_member_path",
			"malformed_manifest",
			"missing_required_blob",
			"missing_required_file",
			"remote_fetch_required",
			"signature_mismatch",
			"source_family_invalid",
			"unsupported_bundle_version",
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

func TestIncidentBundlePortabilityFailuresUseClosedRedactedDetails_Unit(t *testing.T) {
	tests := []struct {
		name          string
		err           error
		wantInvariant string
	}{
		{name: "participant admission", err: newPortabilityFailure(ErrPortabilityUnavailable, "secret-profile"), wantInvariant: "extension_payload.participant_admitted"},
		{name: "blocked participant", err: newPortabilityFailure(ErrPortabilityBlocked, "secret-profile"), wantInvariant: "extension_payload.participant_admitted"},
		{name: "resource bound", err: newPortabilityFailure(ErrPortabilityLimit, "secret-profile"), wantInvariant: "extension_payload.resource_bounded"},
		{name: "schema digest", err: newPortabilityFailure(ErrPortabilityPayload, "secret-profile"), wantInvariant: "extension_payload.schema_digest_valid"},
		{name: "contract", err: newPortabilityFailure(ErrPortabilityResult, "secret-profile"), wantInvariant: "extension_payload.contract_compatible"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			reason, details := incidentBundleFailureDetails("incident_bundle_import_rejected", test.err)
			if reason != "source_family_invalid" ||
				details["reason_code"] != "source_family_invalid" ||
				details["source_family_id"] != "extension_payload" ||
				details["invariant_id"] != test.wantInvariant {
				t.Fatalf("failure details = %#v, reason=%q", details, reason)
			}
			if len(details) != 3 {
				t.Fatalf("failure details exposed an unregistered field: %#v", details)
			}
		})
	}
}

func TestAdmittedRouteSetupRequiresImportFinalizer_Unit(t *testing.T) {
	projections := httpapiextensions.New([]httpapi.ExtensionProfile{{ProfileID: ProfileID, Claimed: true}})
	deps := projections.Dependencies(httpapi.DependencySet{})
	err := RegisterRoutes()(http.NewServeMux(), deps)
	if err == nil || !strings.Contains(err.Error(), "import finalizer") {
		t.Fatalf("admitted incident portability setup must fail without import finalizer, got %v", err)
	}
	err = RegisterRoutes(WithImportFinalizer(importFinalizerStub{}))(
		http.NewServeMux(),
		deps,
	)
	if err == nil || !strings.Contains(err.Error(), "job success finalizer") {
		t.Fatalf("admitted incident portability setup must fail without job success finalizer, got %v", err)
	}
}

func TestClaimedIncidentPortabilityRejectsMissingJobsBeforePublication_Unit(t *testing.T) {
	_, err := newService(httpapi.DependencySet{}, routeOptions{
		importFinalizer:   importFinalizerStub{},
		jobFinalizer:      jobFinalizerStub{},
		portability:       &PortabilityOrchestrator{},
		transactions:      &crossownertransaction.Coordinator{},
		storage:           bundleStorageStub{},
		projectionRebuild: projectionRebuilderStub{},
		sourceCatalog:     &sourceport.Catalog{},
	})
	if err == nil || !strings.Contains(err.Error(), "jobs composition is required") {
		t.Fatalf("claimed Incident Portability without Jobs composition error = %v", err)
	}
}

type importFinalizerStub struct{}

func (importFinalizerStub) FinalizeIncidentBundleImportTx(
	context.Context,
	pgx.Tx,
	incidents.IncidentBundleImportFinalizationParams,
) error {
	return nil
}

type jobFinalizerStub struct{}

func (jobFinalizerStub) FinalizeIncidentBundleJobSuccess(context.Context, JobSuccessFinalization) (jobs.Resource, error) {
	return jobs.Resource{}, nil
}

func (jobFinalizerStub) FinalizeIncidentBundleJobSuccessTx(context.Context, crossownertransaction.FinalizationCapability, JobSuccessFinalization) (jobs.Resource, error) {
	return jobs.Resource{}, nil
}

func (jobFinalizerStub) FinalizeIncidentBundleJobFailure(context.Context, JobFailureFinalization) (jobs.Resource, error) {
	return jobs.Resource{}, nil
}

type bundleStorageStub struct{}

func (bundleStorageStub) Stage(context.Context, string, []byte) (BundleStagingRef, error) {
	return BundleStagingRef{}, nil
}

func (bundleStorageStub) Publish(context.Context, string, []byte) (BundleStorageRef, error) {
	return BundleStorageRef{}, nil
}

func (bundleStorageStub) ReadStaged(BundleStagingRef, int64) ([]byte, error) {
	return nil, nil
}

func (bundleStorageStub) RemoveStaged(BundleStagingRef) error {
	return nil
}

func (bundleStorageStub) RemovePublished(BundleStorageRef) error {
	return nil
}

type projectionRebuilderStub struct{}

func (projectionRebuilderStub) RebuildImportedIncidentTx(context.Context, pgx.Tx, uuid.UUID) error {
	return nil
}

func TestWorkerResultTransitionsPreservePublicSummaries_Unit(t *testing.T) {
	bundleID := uuid.MustParse("22222222-2222-2222-2222-222222222222")
	incidentID := uuid.MustParse("33333333-3333-3333-3333-333333333333")

	exportTransition := exportSuccessCompletion(bundleID)
	if exportTransition.ResultSummary.Code == "" {
		t.Fatalf("export success transition missing summary: %#v", exportTransition)
	}
	if exportTransition.ResultSummary.Code != ResultIncidentBundleExported || len(exportTransition.ResultSummary.ResourceRefs) != 1 {
		t.Fatalf("export result summary mismatch: %#v", exportTransition.ResultSummary)
	}
	exportRef := exportTransition.ResultSummary.ResourceRefs[0]
	if exportRef.Kind != "incident_bundle" || exportRef.ID != bundleID.String() || exportRef.Route != "/api/v1/incident-bundles/"+bundleID.String() {
		t.Fatalf("export resource ref mismatch: %#v", exportRef)
	}

	importTransition := importSuccessCompletion(incidentID)
	if importTransition.ResultSummary.Code == "" {
		t.Fatalf("import success transition missing summary: %#v", importTransition)
	}
	if importTransition.ResultSummary.Code != ResultIncidentBundleImported || len(importTransition.ResultSummary.ResourceRefs) != 1 {
		t.Fatalf("import result summary mismatch: %#v", importTransition.ResultSummary)
	}
	importRef := importTransition.ResultSummary.ResourceRefs[0]
	if importRef.Kind != "incident" || importRef.ID != incidentID.String() || importRef.Route != "/api/v1/incidents/"+incidentID.String() {
		t.Fatalf("import resource ref mismatch: %#v", importRef)
	}
}

func TestIncidentBundleStorageReferencesAreStrictAndRootFree_Unit(t *testing.T) {
	valid := []string{
		"incident-bundles/22222222-2222-2222-2222-222222222222.zip",
		"incident-bundles/imports/abc.bundle",
	}
	for _, raw := range valid {
		if reference, err := ParseBundleStorageRef(raw); err != nil || reference.String() != raw {
			t.Fatalf("ParseBundleStorageRef(%q) = %q, %v", raw, reference.String(), err)
		}
		if reference, err := ParseBundleStagingRef(raw); err != nil || reference.String() != raw {
			t.Fatalf("ParseBundleStagingRef(%q) = %q, %v", raw, reference.String(), err)
		}
	}

	invalid := []string{
		"",
		".",
		"..",
		"/var/lib/cartulary/export.zip",
		"incident-bundles/../export.zip",
		"incident-bundles//export.zip",
		"incident-bundles/./export.zip",
		`incident-bundles\export.zip`,
		"incident-bundles/export.zip/",
		"incident-bundles/\x00export.zip",
		"incident-bundles/cafe\u0301.zip",
	}
	for _, raw := range invalid {
		if _, err := ParseBundleStorageRef(raw); !errors.Is(err, ErrInvalidStorageReference) {
			t.Fatalf("ParseBundleStorageRef(%q) error = %v; want ErrInvalidStorageReference", raw, err)
		}
		if _, err := ParseBundleStagingRef(raw); !errors.Is(err, ErrInvalidStorageReference) {
			t.Fatalf("ParseBundleStagingRef(%q) error = %v; want ErrInvalidStorageReference", raw, err)
		}
	}
}

func TestIncidentBundleWorkerRequiresNamedRunner_Unit(t *testing.T) {
	worker := &incidentBundleWorker{}
	if err := worker.registerJobHandler(); !errors.Is(err, jobs.ErrNotConfigured) {
		t.Fatalf("register without named runner error = %v; want ErrNotConfigured", err)
	}
	if err := worker.dispatch(uuid.NewString()); !errors.Is(err, jobs.ErrNotConfigured) {
		t.Fatalf("dispatch without named runner error = %v; want ErrNotConfigured", err)
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

func TestDecodeImportMetadataUsesUploadHash_Unit(t *testing.T) {
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
