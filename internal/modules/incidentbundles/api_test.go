package incidentbundles

import (
	"archive/tar"
	"archive/zip"
	"bytes"
	"compress/gzip"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/JochiRaider/cartulary/internal/modules/crossownertransaction"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/importfinalizerport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentbundles/sourceport"
	"github.com/JochiRaider/cartulary/internal/modules/incidentportability"
	"github.com/JochiRaider/cartulary/internal/platform/contracttest"
	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/JochiRaider/cartulary/internal/testutil/httpapiextensions"
)

func TestDecodeExportRequestCanonicalizesAndRejectsModes_Unit(t *testing.T) {
	request, apiErr := decodeExportRequest(bytes.NewBufferString(`{
		"incident_id":"11111111-1111-1111-1111-111111111111",
		"client_txn_id":"txn-export",
		"optional_sections":["reference_packs","snapshots","reference_packs"],
		"required_capabilities":[],
		"reference_pack_mode":"embedded"
	}`))
	if apiErr != nil {
		t.Fatalf("decodeExportRequest returned error: %#v", apiErr)
	}
	if request.HistoryMode != historyModeFull || request.BlobMode != blobModeFull {
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

	omitted, omittedErr := decodeExportRequest(bytes.NewBufferString(`{
		"incident_id":"11111111-1111-1111-1111-111111111111",
		"client_txn_id":"txn-export-equivalent"
	}`))
	explicitEmpty, emptyErr := decodeExportRequest(bytes.NewBufferString(`{
		"incident_id":"11111111-1111-1111-1111-111111111111",
		"client_txn_id":"txn-export-equivalent",
		"optional_sections":[],
		"required_capabilities":[],
		"reference_pack_mode":"refs_only"
	}`))
	if omittedErr != nil || emptyErr != nil {
		t.Fatalf("omitted/empty equivalent requests failed: omitted=%#v empty=%#v", omittedErr, emptyErr)
	}
	if !bytes.Equal(omitted.Normalized, explicitEmpty.Normalized) {
		t.Fatalf("omitted and explicit-empty defaults normalize differently:\n omitted=%s\nexplicit=%s", omitted.Normalized, explicitEmpty.Normalized)
	}

	_, apiErr = decodeExportRequest(bytes.NewBufferString(`{
		"incident_id":"11111111-1111-1111-1111-111111111111",
		"client_txn_id":"txn-export-malformed-capability",
		"required_capabilities":[1]
	}`))
	if apiErr == nil || apiErr.Status != http.StatusBadRequest || apiErr.Code != "invalid_incident_bundle_request" || apiErr.Details["reason_code"] != "invalid_required_capabilities" {
		t.Fatalf("non-string capability member must remain a structural request failure: %#v", apiErr)
	}

	activation, apiErr := decodeExportRequest(bytes.NewBufferString(`{
		"incident_id":"11111111-1111-1111-1111-111111111111",
		"client_txn_id":"txn-export-required-capability",
		"required_capabilities":["future-capability","future-capability"]
	}`))
	if apiErr != nil || !activation.CapabilityActivationRequested || len(activation.RequiredCapabilities) != 0 || strings.Contains(string(activation.Normalized), "future-capability") {
		t.Fatalf("valid capability activation was classified or retained: request=%#v error=%#v", activation, apiErr)
	}

	for _, body := range []string{
		`{"incident_id":"11111111-1111-1111-1111-111111111111","client_txn_id":"txn","history_mode":"partial"}`,
		`{"incident_id":"11111111-1111-1111-1111-111111111111","client_txn_id":"txn","blob_mode":"metadata_only"}`,
	} {
		_, apiErr := decodeExportRequest(bytes.NewBufferString(body))
		if apiErr == nil || apiErr.Code != "invalid_incident_bundle_request" {
			t.Fatalf("forbidden modes must fail with invalid_incident_bundle_request, got %#v", apiErr)
		}
	}
}

func TestBundleManifestChecksumDeterministic_Unit(t *testing.T) {
	requireClosedRequiredSourceFileRegistry(t)

	files := minimalRequiredBundleFiles()
	files["data/records.ndjson"] = []byte("{}\n")
	files["data/artifacts.ndjson"] = []byte("{\"record_id\":\"33333333-3333-4333-8333-333333333333\"}\n")
	first, err := buildBundleArchive(manifestInput{
		BundleID:             "22222222-2222-2222-2222-222222222222",
		IncidentID:           "11111111-1111-1111-1111-111111111111",
		IncidentKey:          "INC-1",
		ExportedAt:           "2026-05-25T00:00:00Z",
		ReferencePackMode:    referencePackModeRefsOnly,
		OptionalSections:     []string{},
		RequiredCapabilities: []string{},
	}, files)
	if err != nil {
		t.Fatalf("buildBundleArchive first: %v", err)
	}
	secondFiles := minimalRequiredBundleFiles()
	secondFiles["data/records.ndjson"] = []byte("{}\n")
	secondFiles["data/artifacts.ndjson"] = []byte("{\"record_id\":\"33333333-3333-4333-8333-333333333333\"}\n")
	second, err := buildBundleArchive(manifestInput{
		BundleID:             "22222222-2222-2222-2222-222222222222",
		IncidentID:           "11111111-1111-1111-1111-111111111111",
		IncidentKey:          "INC-1",
		ExportedAt:           "2026-05-25T00:00:00Z",
		ReferencePackMode:    referencePackModeRefsOnly,
		OptionalSections:     []string{},
		RequiredCapabilities: []string{},
	}, secondFiles)
	if err != nil {
		t.Fatalf("buildBundleArchive second: %v", err)
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
		if file.Path == "data/artifacts.ndjson" && file.SizeBytes != int64(len(files["data/artifacts.ndjson"])) {
			t.Fatalf("artifact manifest file must use exact size_bytes: %#v", file)
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
	_, err := buildBundleArchive(manifestInput{
		BundleID:          "55555555-5555-5555-5555-555555555555",
		IncidentID:        "11111111-1111-1111-1111-111111111111",
		IncidentKey:       "INC-1",
		ExportedAt:        "2026-05-25T00:00:00Z",
		ReferencePackMode: referencePackModeRefsOnly,
	}, files)
	if err == nil || !strings.Contains(err.Error(), "data/parties.ndjson is required") {
		t.Fatalf("buildBundleArchive must reject missing required source files, got %v", err)
	}
	files = minimalRequiredBundleFiles()
	delete(files, "data/saved_views.ndjson")
	_, err = buildBundleArchive(manifestInput{
		BundleID:          "55555555-5555-5555-5555-555555555556",
		IncidentID:        "11111111-1111-1111-1111-111111111111",
		IncidentKey:       "INC-1",
		ExportedAt:        "2026-05-25T00:00:00Z",
		ReferencePackMode: referencePackModeRefsOnly,
	}, files)
	if err == nil || !strings.Contains(err.Error(), "data/saved_views.ndjson is required") {
		t.Fatalf("buildBundleArchive must reject missing saved views source file, got %v", err)
	}
}

func TestVerifyBundleRejectsUnsafeAndCapabilityFailures_Unit(t *testing.T) {
	unsafe := newZip(t, map[string][]byte{"../manifest.json": []byte("{}")})
	_, err := verifyBundle(verificationInput{
		Bundle: unsafe,
		Limits: Limits{Archives: ArchiveLimits{MaxMembers: 100, MaxCompressionRatio: 100}, IncidentBundles: IncidentBundleLimits{MaxExtractedBytes: 1024}},
	})
	if !isVerificationReason(err, "invalid_member_path") {
		t.Fatalf("unsafe path reason mismatch: %v", err)
	}

	files := minimalRequiredBundleFiles()
	bundle, err := buildBundleArchive(manifestInput{
		BundleID:             "22222222-2222-2222-2222-222222222222",
		IncidentID:           "11111111-1111-1111-1111-111111111111",
		IncidentKey:          "INC-1",
		ExportedAt:           "2026-05-25T00:00:00Z",
		ReferencePackMode:    referencePackModeRefsOnly,
		RequiredCapabilities: []string{"snapshots"},
	}, files)
	if err != nil {
		t.Fatalf("buildBundleArchive: %v", err)
	}
	_, err = verifyBundle(verificationInput{
		Bundle: bundle.Bytes,
		Limits: Limits{Archives: ArchiveLimits{MaxMembers: 100, MaxCompressionRatio: 100}, IncidentBundles: IncidentBundleLimits{MaxExtractedBytes: 1024 * 1024}},
	})
	if !errors.Is(err, errExtensionCapabilityNotSupported) {
		t.Fatalf("required capability reason mismatch: %v", err)
	}

	validBundle, err := buildBundleArchive(manifestInput{
		BundleID:          "33333333-3333-3333-3333-333333333333",
		IncidentID:        "11111111-1111-1111-1111-111111111111",
		IncidentKey:       "INC-1",
		ExportedAt:        "2026-05-25T00:00:00Z",
		ReferencePackMode: referencePackModeRefsOnly,
	}, files)
	if err != nil {
		t.Fatalf("buildBundleArchive valid bundle: %v", err)
	}
	missingRequired := removeZipMember(t, validBundle.Bytes, "data/records.ndjson")
	_, err = verifyBundle(verificationInput{
		Bundle: missingRequired,
		Limits: Limits{Archives: ArchiveLimits{MaxMembers: 100, MaxCompressionRatio: 100}, IncidentBundles: IncidentBundleLimits{MaxExtractedBytes: 1024 * 1024}},
	})
	if !isVerificationReason(err, "missing_required_file") {
		t.Fatalf("missing required file reason mismatch: %v", err)
	}

	withZipDirectories := appendZipMember(t, appendZipMember(t, validBundle.Bytes, "data/", nil), "integrity/", nil)
	if _, err = verifyBundle(verificationInput{
		Bundle: withZipDirectories,
		Limits: Limits{Archives: ArchiveLimits{MaxMembers: 100, MaxCompressionRatio: 100}, IncidentBundles: IncidentBundleLimits{MaxExtractedBytes: 1024 * 1024}},
	}); err != nil {
		t.Fatalf("safe ZIP directory members must be ignored as structural entries: %v", err)
	}
	_, err = verifyBundle(verificationInput{
		Bundle: withZipDirectories,
		Limits: Limits{Archives: ArchiveLimits{MaxMembers: int64(len(zipFilesMap(t, validBundle.Bytes)) + 1), MaxCompressionRatio: 100}, IncidentBundles: IncidentBundleLimits{MaxExtractedBytes: 1024 * 1024}},
	})
	if !isVerificationReason(err, "archive_member_count_exceeded") {
		t.Fatalf("ZIP directory entries must count against member limits: %v", err)
	}
	unsafeDirectory := appendZipMember(t, validBundle.Bytes, "../data/", nil)
	_, err = verifyBundle(verificationInput{
		Bundle: unsafeDirectory,
		Limits: Limits{Archives: ArchiveLimits{MaxMembers: 100, MaxCompressionRatio: 100}, IncidentBundles: IncidentBundleLimits{MaxExtractedBytes: 1024 * 1024}},
	})
	if !isVerificationReason(err, "invalid_member_path") {
		t.Fatalf("unsafe ZIP directory reason mismatch: %v", err)
	}
	unsupportedZipMember := appendZipSymlink(t, validBundle.Bytes, "data/member-link", "manifest.json")
	_, err = verifyBundle(verificationInput{
		Bundle: unsupportedZipMember,
		Limits: Limits{Archives: ArchiveLimits{MaxMembers: 100, MaxCompressionRatio: 100}, IncidentBundles: IncidentBundleLimits{MaxExtractedBytes: 1024 * 1024}},
	})
	if !isVerificationReason(err, "unsupported_member_type") {
		t.Fatalf("unsupported ZIP member reason mismatch: %v", err)
	}
	withTarDirectories := newTarGzip(t, []string{"data", "integrity/"}, zipFilesMap(t, validBundle.Bytes))
	if _, err = verifyBundle(verificationInput{
		Bundle: withTarDirectories,
		Limits: Limits{Archives: ArchiveLimits{MaxMembers: 100, MaxCompressionRatio: 100}, IncidentBundles: IncidentBundleLimits{MaxExtractedBytes: 1024 * 1024}},
	}); err != nil {
		t.Fatalf("safe TAR directory members must be ignored as structural entries: %v", err)
	}
	_, err = verifyBundle(verificationInput{
		Bundle: withTarDirectories,
		Limits: Limits{Archives: ArchiveLimits{MaxMembers: int64(len(zipFilesMap(t, validBundle.Bytes)) + 1), MaxCompressionRatio: 100}, IncidentBundles: IncidentBundleLimits{MaxExtractedBytes: 1024 * 1024}},
	})
	if !isVerificationReason(err, "archive_member_count_exceeded") {
		t.Fatalf("TAR directory entries must count against member limits: %v", err)
	}

	withSignature := appendZipMember(t, validBundle.Bytes, "integrity/signature.ed25519", []byte("not-a-supported-signature"))
	_, err = verifyBundle(verificationInput{
		Bundle: withSignature,
		Limits: Limits{Archives: ArchiveLimits{MaxMembers: 100}, IncidentBundles: IncidentBundleLimits{MaxExtractedBytes: 1024 * 1024}},
	})
	if !isVerificationReason(err, "signature_mismatch") {
		t.Fatalf("signature reason mismatch: %v", err)
	}

	malformedMode := replaceManifestFields(t, validBundle.Bytes, func(manifest map[string]any) {
		manifest["reference_pack_mode"] = "floating"
	})
	_, err = verifyBundle(verificationInput{
		Bundle: malformedMode,
		Limits: Limits{Archives: ArchiveLimits{MaxMembers: 100, MaxCompressionRatio: 100}, IncidentBundles: IncidentBundleLimits{MaxExtractedBytes: 1024 * 1024}},
	})
	if !isVerificationReason(err, "malformed_manifest") {
		t.Fatalf("invalid reference_pack_mode reason mismatch: %v", err)
	}

	malformedOptionalSection := replaceManifestFields(t, validBundle.Bytes, func(manifest map[string]any) {
		manifest["optional_sections"] = []any{"unknown_section"}
	})
	_, err = verifyBundle(verificationInput{
		Bundle: malformedOptionalSection,
		Limits: Limits{Archives: ArchiveLimits{MaxMembers: 100, MaxCompressionRatio: 100}, IncidentBundles: IncidentBundleLimits{MaxExtractedBytes: 1024 * 1024}},
	})
	if !isVerificationReason(err, "malformed_manifest") {
		t.Fatalf("unknown optional_sections token reason mismatch: %v", err)
	}

	unsupportedRequired := replaceManifestFields(t, validBundle.Bytes, func(manifest map[string]any) {
		manifest["required_capabilities"] = []any{"snapshots"}
	})
	_, err = verifyBundle(verificationInput{
		Bundle: unsupportedRequired,
		Limits: Limits{Archives: ArchiveLimits{MaxMembers: 100, MaxCompressionRatio: 100}, IncidentBundles: IncidentBundleLimits{MaxExtractedBytes: 1024 * 1024}},
	})
	if !errors.Is(err, errExtensionCapabilityNotSupported) {
		t.Fatalf("unsupported required capability reason mismatch: %v", err)
	}
	malformedRequired := replaceManifestFields(t, validBundle.Bytes, func(manifest map[string]any) {
		manifest["required_capabilities"] = []any{1}
	})
	_, err = verifyBundle(verificationInput{
		Bundle: malformedRequired,
		Limits: Limits{Archives: ArchiveLimits{MaxMembers: 100, MaxCompressionRatio: 100}, IncidentBundles: IncidentBundleLimits{MaxExtractedBytes: 1024 * 1024}},
	})
	if !isVerificationReason(err, "malformed_manifest") {
		t.Fatalf("structurally invalid required capability reason mismatch: %v", err)
	}
	capabilityFiles := zipFilesMap(t, unsupportedRequired)
	capabilityFiles["data/records.ndjson"] = append(capabilityFiles["data/records.ndjson"], []byte("{}\n")...)
	capabilityWithBadChecksum := newZip(t, capabilityFiles)
	_, err = verifyBundle(verificationInput{
		Bundle: capabilityWithBadChecksum,
		Limits: Limits{Archives: ArchiveLimits{MaxMembers: 100, MaxCompressionRatio: 100}, IncidentBundles: IncidentBundleLimits{MaxExtractedBytes: 1024 * 1024}},
	})
	if !isVerificationReason(err, "checksum_mismatch") {
		t.Fatalf("archive integrity must precede capability rejection: %v", err)
	}

	withKnownOptionalSection, err := buildBundleArchive(manifestInput{
		BundleID:             "44444444-4444-4444-4444-444444444444",
		IncidentID:           "11111111-1111-1111-1111-111111111111",
		IncidentKey:          "INC-1",
		ExportedAt:           "2026-05-25T00:00:00Z",
		ReferencePackMode:    referencePackModeRefsOnly,
		OptionalSections:     []string{"snapshots"},
		RequiredCapabilities: []string{},
	}, withAdditionalBundleFiles(minimalRequiredBundleFiles(), map[string][]byte{
		"ext/snapshots/snapshot.json": []byte(`{"snapshot_id":"snap-1"}` + "\n"),
	}))
	if err != nil {
		t.Fatalf("buildBundleArchive optional section: %v", err)
	}
	if _, err = verifyBundle(verificationInput{
		Bundle: withKnownOptionalSection.Bytes,
		Limits: Limits{Archives: ArchiveLimits{MaxMembers: 100, MaxCompressionRatio: 100}, IncidentBundles: IncidentBundleLimits{MaxExtractedBytes: 1024 * 1024}},
	}); err != nil {
		t.Fatalf("known unsupported optional embedded section must not block core verification: %v", err)
	}
}

func TestVerifyBundleRejectsMalformedManifestVersion_Unit(t *testing.T) {
	files := minimalRequiredBundleFiles()
	bundle, err := buildBundleArchive(manifestInput{
		BundleID:          "22222222-2222-4222-8222-222222222222",
		IncidentID:        "11111111-1111-4111-8111-111111111111",
		IncidentKey:       "INC-VERSION-MALFORMED",
		ExportedAt:        "2026-07-27T00:00:00Z",
		ReferencePackMode: referencePackModeRefsOnly,
	}, files)
	if err != nil {
		t.Fatalf("buildBundleArchive: %v", err)
	}
	cases := map[string]func(map[string]any){
		"omitted":     func(manifest map[string]any) { delete(manifest, "bundle_version") },
		"null":        func(manifest map[string]any) { manifest["bundle_version"] = nil },
		"string":      func(manifest map[string]any) { manifest["bundle_version"] = "2" },
		"non_integer": func(manifest map[string]any) { manifest["bundle_version"] = 2.5 },
	}
	for name, mutate := range cases {
		t.Run(name, func(t *testing.T) {
			_, verifyErr := verifyBundle(verificationInput{
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
	bundle, err := buildBundleArchive(manifestInput{
		BundleID:          "22222222-2222-4222-8222-222222222223",
		IncidentID:        "11111111-1111-4111-8111-111111111111",
		IncidentKey:       "INC-VERSION-CLOSED",
		ExportedAt:        "2026-07-27T00:00:00Z",
		ReferencePackMode: referencePackModeRefsOnly,
	}, minimalRequiredBundleFiles())
	if err != nil {
		t.Fatalf("buildBundleArchive: %v", err)
	}
	for _, version := range []int{1, 3} {
		t.Run(fmt.Sprintf("unsupported_version_%d", version), func(t *testing.T) {
			unsupported := replaceManifestFields(t, bundle.Bytes, func(manifest map[string]any) {
				manifest["bundle_version"] = version
			})
			if _, verifyErr := verifyBundle(verificationInput{
				Bundle: unsupported,
				Limits: Limits{Archives: ArchiveLimits{MaxMembers: 100, MaxCompressionRatio: 100}, IncidentBundles: IncidentBundleLimits{MaxExtractedBytes: 1024 * 1024}},
			}); !isVerificationReason(verifyErr, "unsupported_bundle_version") {
				t.Fatalf("unsupported integer version %d reason mismatch: %v", version, verifyErr)
			}
		})
	}

	for name, additionalPaths := range map[string][]string{
		"retired timeline profile": {"data/timeline_time_conversion_profiles.ndjson"},
		"retired timeline records": {"data/timeline_events.ndjson"},
		"both retired paths":       {"data/timeline_time_conversion_profiles.ndjson", "data/timeline_events.ndjson"},
		"unknown core member":      {"data/unknown.ndjson"},
	} {
		t.Run(name, func(t *testing.T) {
			mixedFiles := zipFilesMap(t, bundle.Bytes)
			for _, path := range additionalPaths {
				mixedFiles[path] = []byte{}
			}
			if _, verifyErr := verifyBundle(verificationInput{
				Bundle: zipFromFiles(t, mixedFiles),
				Limits: Limits{Archives: ArchiveLimits{MaxMembers: 100, MaxCompressionRatio: 100}, IncidentBundles: IncidentBundleLimits{MaxExtractedBytes: 1024 * 1024}},
			}); !isVerificationReason(verifyErr, "malformed_manifest") {
				t.Fatalf("closed path classification mismatch: %v", verifyErr)
			}
		})
	}
}

func TestWorkerRejectsRetiredVersionBeforePreparationAndTransaction_Unit(t *testing.T) {
	bundle, err := buildBundleArchive(manifestInput{
		BundleID:          "22222222-2222-4222-8222-222222222224",
		IncidentID:        "11111111-1111-4111-8111-111111111111",
		IncidentKey:       "INC-RETIRED-VERSION-BOUNDARY",
		ExportedAt:        "2026-07-27T00:00:00Z",
		ReferencePackMode: referencePackModeRefsOnly,
	}, minimalRequiredBundleFiles())
	if err != nil {
		t.Fatalf("buildBundleArchive: %v", err)
	}
	retired := replaceManifestFields(t, bundle.Bytes, func(manifest map[string]any) {
		manifest["bundle_version"] = 1
	})
	stagingRef, err := ParseBundleStagingRef("incident-bundles/imports/retired-version.bundle")
	if err != nil {
		t.Fatalf("parse staging reference: %v", err)
	}
	manager := &recordingJobOperations{}
	storage := &recordingBundleStorage{staged: retired}
	boundaries := &importBoundaryTripwires{}
	worker := &incidentBundleWorker{
		jobManager: manager,
		results:    incidentBundleJobResultSink{manager: manager},
		storage:    storage,
		limits: Limits{
			Archives:        ArchiveLimits{MaxMembers: 100, MaxCompressionRatio: 100},
			IncidentBundles: IncidentBundleLimits{MaxExtractedBytes: 1024 * 1024},
		},
		// Each collaborator panics at its first preparation or mutation method.
		// A retired version must complete without reaching any of them.
		sourceCatalog: sourcePreparationTripwire{boundaries: boundaries},
		portability:   extensionPreparationTripwire{boundaries: boundaries},
		transactions:  transactionExecutionTripwire{boundaries: boundaries},
	}
	worker.executeImportJob(context.Background(), jobs.Execution{}, jobPayload{
		JobKind:          "import",
		JobID:            uuid.MustParse("44444444-4444-4444-8444-444444444445"),
		ActorUserID:      uuid.MustParse("55555555-5555-4555-8555-555555555555"),
		BundleStagingRef: &stagingRef,
	})
	if manager.failed == nil || manager.failed.ErrorSummary.Code != "incident_bundle_import_rejected" ||
		manager.failed.ErrorSummary.Retryable ||
		manager.failed.ErrorSummary.Details["reason_code"] != "unsupported_bundle_version" {
		t.Fatalf("retired version completion = %#v", manager.failed)
	}
	if storage.stagedReads != 1 || storage.stagedRemovals != 1 {
		t.Fatalf("retired version staging lifecycle = reads %d removals %d", storage.stagedReads, storage.stagedRemovals)
	}
	if boundaries.sourcePreparation || boundaries.extensionPreparation || boundaries.publicationValidation || boundaries.transactionExecution {
		t.Fatalf("retired version reached downstream boundaries: %#v", boundaries)
	}
}

type importBoundaryTripwires struct {
	sourcePreparation     bool
	extensionPreparation  bool
	publicationValidation bool
	transactionExecution  bool
}

type sourcePreparationTripwire struct {
	boundaries *importBoundaryTripwires
}

func (t sourcePreparationTripwire) Ports() []sourceport.Port {
	t.boundaries.sourcePreparation = true
	panic("retired bundle reached source preparation")
}

type extensionPreparationTripwire struct {
	boundaries *importBoundaryTripwires
}

func (t extensionPreparationTripwire) Export(context.Context, StatePresenceQuery, uuid.UUID) ([]ExtensionPayload, error) {
	panic("retired bundle reached the export-only extension boundary")
}

func (t extensionPreparationTripwire) PrepareImport(context.Context, string, uuid.UUID, map[string][]byte) (PreparedPortability, error) {
	t.boundaries.extensionPreparation = true
	panic("retired bundle reached extension preparation")
}

func (t extensionPreparationTripwire) ValidatePublication(context.Context, StatePresenceQuery, uuid.UUID) error {
	t.boundaries.publicationValidation = true
	panic("retired bundle reached publication validation")
}

type transactionExecutionTripwire struct {
	boundaries *importBoundaryTripwires
}

func (t transactionExecutionTripwire) Execute(context.Context, crossownertransaction.Operation) (crossownertransaction.Result, error) {
	t.boundaries.transactionExecution = true
	panic("retired bundle reached transaction execution")
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
	t.Run("production export surface is closed", func(t *testing.T) {
		actual := incidentBundleExportedDeclarations(t)
		unapproved := incidentBundleExportDifference(actual, incidentBundleExportAllowlist)
		stale := incidentBundleExportDifference(incidentBundleExportAllowlist, actual)
		if len(unapproved) != 0 || len(stale) != 0 {
			t.Fatalf("Incident Bundles exported surface mismatch: unapproved=%v stale=%v", unapproved, stale)
		}
	})

	t.Run("internal errors use one value-free response tuple", func(t *testing.T) {
		sentinels := []string{
			"SELECT secret_value FROM deployment_credentials",
			"/var/lib/cartulary/private/archive.zip",
			"s3://private-bucket/object?token=storage-secret",
			"postgres://operator:credential@database/cartulary",
			"upstream unauthorized: api_key=credential-secret",
		}
		for _, sentinel := range sentinels {
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(http.MethodGet, "/api/v1/incident-bundles/00000000-0000-4000-8000-000000000001", nil)
			writeAPIError(recorder, request, &httpapi.APIError{
				Status: http.StatusTeapot, Code: "internal_error", Message: sentinel,
				Details: map[string]any{"raw": sentinel}, Conflict: sentinel, Retryable: true,
			})
			if recorder.Code != http.StatusInternalServerError {
				t.Fatalf("internal error status = %d", recorder.Code)
			}
			var envelope map[string]any
			if err := json.Unmarshal(recorder.Body.Bytes(), &envelope); err != nil {
				t.Fatalf("decode internal error response: %v", err)
			}
			payload, ok := envelope["error"].(map[string]any)
			if !ok || payload["code"] != "internal_error" || payload["message"] != "internal_error" || payload["retryable"] != false {
				t.Fatalf("internal error payload = %#v", envelope)
			}
			details, ok := payload["details"].(map[string]any)
			if !ok || len(details) != 0 || payload["conflict"] != nil || strings.Contains(recorder.Body.String(), sentinel) {
				t.Fatalf("internal error disclosed sentinel: %s", recorder.Body.String())
			}
		}
		apiErr := internalAPIError()
		if apiErr.Status != http.StatusInternalServerError || apiErr.Code != "internal_error" ||
			apiErr.Message != "internal_error" || len(apiErr.Details) != 0 {
			t.Fatalf("internal error constructor = %#v", apiErr)
		}
	})

	openAPIDoc := contracttest.OpenAPIDocument(t)
	optionalSections := openAPIObjectAt(t, openAPIDoc, "components", "schemas", "IncidentBundleResource", "properties", "optional_sections")
	if optionalSections["type"] != "array" || optionalSections["uniqueItems"] != true {
		t.Fatalf("optional_sections must remain a unique array schema: %#v", optionalSections)
	}
	optionalItems, ok := optionalSections["items"].(map[string]any)
	if !ok {
		t.Fatalf("optional_sections items schema missing: %#v", optionalSections)
	}
	enumValues, ok := optionalItems["enum"].([]any)
	if !ok || !slices.Equal(enumStrings(enumValues), []string{"reference_packs", "snapshots"}) {
		t.Fatalf("optional_sections enum tokens mismatch: %#v", optionalItems["enum"])
	}

	resourceCapabilities := openAPIObjectAt(t, openAPIDoc, "components", "schemas", "IncidentBundleResource", "properties", "required_capabilities")
	if resourceCapabilities["type"] != "array" || resourceCapabilities["maxItems"] != float64(0) {
		t.Fatalf("successful descriptor capabilities must be an empty-only array: %#v", resourceCapabilities)
	}
	resourceItems, ok := resourceCapabilities["items"].(map[string]any)
	if !ok || resourceItems["type"] != "string" || resourceItems["enum"] != nil {
		t.Fatalf("descriptor capabilities must not publish obsolete capability tokens: %#v", resourceItems)
	}
	exportRequiredCapabilities := openAPIObjectAt(t, openAPIDoc, "components", "schemas", "IncidentBundleExportRequest", "properties", "required_capabilities")
	exportItems, ok := exportRequiredCapabilities["items"].(map[string]any)
	if !ok || exportRequiredCapabilities["maxItems"] != nil || exportRequiredCapabilities["uniqueItems"] != nil || exportItems["type"] != "string" || exportItems["enum"] != nil {
		t.Fatalf("export required_capabilities must admit any structurally valid string array for semantic rejection: %#v", exportRequiredCapabilities)
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
	exportConflict := openAPIObjectAt(t, openAPIDoc, "paths", "/api/v1/incident-bundles/export", "post", "responses", "409")
	exportCodes, ok := exportConflict["x-cartulary-error-codes"].([]any)
	if !ok || !slices.Contains(enumStrings(exportCodes), "extension_capability_not_supported") {
		t.Fatalf("export 409 contract omits shared capability error: %#v", exportConflict)
	}
	descriptorInvalid := openAPIObjectAt(t, openAPIDoc, "paths", "/api/v1/incident-bundles/{bundle_id}", "get", "responses", "400")
	descriptorCodes, ok := descriptorInvalid["x-cartulary-error-codes"].([]any)
	if !ok || !slices.Equal(enumStrings(descriptorCodes), []string{"invalid_pagination_request"}) {
		t.Fatalf("descriptor 400 contract must use only invalid_pagination_request: %#v", descriptorInvalid)
	}
	_ = openAPIObjectAt(t, openAPIDoc, "components", "schemas", "IncidentBundleResource")
	for code, status := range map[string]int{
		"extension_capability_not_supported": http.StatusConflict,
		"incident_bundle_export_rejected":    http.StatusConflict,
		"incident_bundle_import_rejected":    http.StatusConflict,
		"incident_bundle_not_found":          http.StatusNotFound,
		"invalid_incident_bundle_request":    http.StatusBadRequest,
	} {
		contracttest.RequireErrorContract(t, code, status)
	}
}

var incidentBundleExportAllowlist = map[string]struct{}{
	"ArchiveLimits":                               {},
	"BlobPortability":                             {},
	"BundleStagingRef":                            {},
	"BundleStagingRef.String":                     {},
	"BundleStorage":                               {},
	"BundleStorageRef":                            {},
	"BundleStorageRef.String":                     {},
	"BundlesRouteContributionID":                  {},
	"ClaimStateClaimed":                           {},
	"ClaimStateRecognizedUnclaimable":             {},
	"ClaimStateUnclaimed":                         {},
	"EncodeExtensionPayload":                      {},
	"ErrInvalidStorageReference":                  {},
	"ErrJobFinalizationIndeterminate":             {},
	"ErrPortabilityBlocked":                       {},
	"ErrPortabilityLimit":                         {},
	"ErrPortabilityPayload":                       {},
	"ErrPortabilityResult":                        {},
	"ErrPortabilityUnavailable":                   {},
	"ExportInvocation":                            {},
	"ExportResult":                                {},
	"ExtensionExportResultSchema":                 {},
	"ExtensionImportResultSchema":                 {},
	"ExtensionParticipant":                        {},
	"ExtensionPayload":                            {},
	"ExtensionPayloadSchema":                      {},
	"ExtensionPolicy":                             {},
	"HistoricalIntentPolicy":                      {},
	"ImportInvocation":                            {},
	"ImportPreparation":                           {},
	"ImportProjectionRebuilder":                   {},
	"ImportTransactionDescriptor":                 {},
	"ImportTransactionParticipantID":              {},
	"ImportedAttributionResolver":                 {},
	"IncidentBundleLimits":                        {},
	"IncidentPublicationLock":                     {},
	"JobFailureFinalization":                      {},
	"JobOperations":                               {},
	"JobRunner":                                   {},
	"JobSuccessFinalization":                      {},
	"JobSuccessFinalizer":                         {},
	"JobSuccessMutation":                          {},
	"JobTransactions":                             {},
	"Limits":                                      {},
	"Module":                                      {},
	"Module.InstallCrossOwnerCoordinator":         {},
	"Module.RegisterBundleWorker":                 {},
	"Module.RegisterRoutes":                       {},
	"Module.TransactionCapabilities":              {},
	"ModuleDependencies":                          {},
	"NewModule":                                   {},
	"NewPortabilityOrchestrator":                  {},
	"ParseBundleStagingRef":                       {},
	"ParseBundleStorageRef":                       {},
	"PortabilityFailure":                          {},
	"PortabilityFailure.Error":                    {},
	"PortabilityFailure.Unwrap":                   {},
	"PortabilityModeBlockedWhenPresent":           {},
	"PortabilityModeNoAuthoritativeState":         {},
	"PortabilityModeParticipant":                  {},
	"PortabilityOrchestrator":                     {},
	"PortabilityOrchestrator.Export":              {},
	"PortabilityOrchestrator.PrepareImport":       {},
	"PortabilityOrchestrator.ValidatePublication": {},
	"PortabilityParticipantByteLimit":             {},
	"PortabilityStagedOutputScope":                {},
	"PortabilityStagedOutputScope.Allocate":       {},
	"PortabilityStagedOutputScope.Refs":           {},
	"PreparedPortability":                         {},
	"PreparedPortability.Abandon":                 {},
	"PreparedPortability.Committed":               {},
	"ProfileID":                                   {},
	"RecoveryStateContribution":                   {},
	"StatePresence":                               {},
	"StatePresenceQuery":                          {},
	"VNextRecoveryObjectInventory":                {},
}

func incidentBundleExportedDeclarations(t testing.TB) map[string]struct{} {
	t.Helper()
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read Incident Bundles package: %v", err)
	}
	fileSet := token.NewFileSet()
	result := map[string]struct{}{}
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".go") || strings.HasSuffix(entry.Name(), "_test.go") {
			continue
		}
		file, err := parser.ParseFile(fileSet, filepath.Clean(entry.Name()), nil, 0)
		if err != nil {
			t.Fatalf("parse %s: %v", entry.Name(), err)
		}
		for _, declaration := range file.Decls {
			switch typed := declaration.(type) {
			case *ast.GenDecl:
				for _, specification := range typed.Specs {
					switch spec := specification.(type) {
					case *ast.TypeSpec:
						if ast.IsExported(spec.Name.Name) {
							result[spec.Name.Name] = struct{}{}
						}
					case *ast.ValueSpec:
						for _, name := range spec.Names {
							if ast.IsExported(name.Name) {
								result[name.Name] = struct{}{}
							}
						}
					}
				}
			case *ast.FuncDecl:
				if !ast.IsExported(typed.Name.Name) {
					continue
				}
				name := typed.Name.Name
				if typed.Recv != nil {
					receiver := incidentBundleExportedReceiver(typed.Recv.List[0].Type)
					if receiver == "" {
						continue
					}
					name = receiver + "." + name
				}
				result[name] = struct{}{}
			}
		}
	}
	return result
}

func incidentBundleExportedReceiver(expression ast.Expr) string {
	switch typed := expression.(type) {
	case *ast.Ident:
		if ast.IsExported(typed.Name) {
			return typed.Name
		}
	case *ast.StarExpr:
		return incidentBundleExportedReceiver(typed.X)
	}
	return ""
}

func incidentBundleExportDifference(left map[string]struct{}, right map[string]struct{}) []string {
	var result []string
	for name := range left {
		if _, ok := right[name]; !ok {
			result = append(result, name)
		}
	}
	sort.Strings(result)
	return result
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
			"extension_state_not_portable",
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
	module, err := NewModule(moduleTestDependencies())
	if err != nil {
		t.Fatalf("construct complete module: %v", err)
	}
	mux := http.NewServeMux()
	registrar := module.RegisterRoutes()
	if err := registrar(mux, deps); err == nil || !strings.Contains(err.Error(), "coordinator is not installed") {
		t.Fatalf("routes before coordinator error = %v", err)
	}
	request := httptest.NewRequest(http.MethodGet, "/api/v1/incident-bundles/11111111-1111-1111-1111-111111111111", nil)
	response := httptest.NewRecorder()
	mux.ServeHTTP(response, request)
	if response.Code != http.StatusNotFound {
		t.Fatalf("failed lifecycle published a partial route: status=%d", response.Code)
	}

	coordinator := &crossownertransaction.Coordinator{}
	if err := module.InstallCrossOwnerCoordinator(nil); err == nil || !strings.Contains(err.Error(), "coordinator is required") {
		t.Fatalf("nil coordinator error = %v", err)
	}
	if err := module.InstallCrossOwnerCoordinator(coordinator); err != nil {
		t.Fatalf("install coordinator: %v", err)
	}
	if err := module.InstallCrossOwnerCoordinator(coordinator); err == nil || !strings.Contains(err.Error(), "already installed") {
		t.Fatalf("duplicate coordinator error = %v", err)
	}
	if err := registrar(mux, deps); err == nil || !strings.Contains(err.Error(), "worker is not registered") {
		t.Fatalf("routes before worker error = %v", err)
	}
	if err := module.RegisterBundleWorker(); err != nil {
		t.Fatalf("register worker: %v", err)
	}
	if err := module.RegisterBundleWorker(); err == nil || !strings.Contains(err.Error(), "already registered") {
		t.Fatalf("duplicate worker registration error = %v", err)
	}
	if err := registrar(mux, deps); err != nil {
		t.Fatalf("bind routes after lifecycle completion: %v", err)
	}

	unordered, err := NewModule(moduleTestDependencies())
	if err != nil {
		t.Fatalf("construct unordered module: %v", err)
	}
	if err := unordered.RegisterBundleWorker(); err == nil || !strings.Contains(err.Error(), "coordinator is not installed") {
		t.Fatalf("worker before coordinator error = %v", err)
	}
}

func TestImportBundleRequestIDIsDeterministic_Unit(t *testing.T) {
	jobID := uuid.MustParse("00000000-0000-0000-0000-000000000905")
	want := "incident_bundle_import:00000000-0000-0000-0000-000000000905"
	if first, second := importBundleRequestID(jobID), importBundleRequestID(jobID); first != want || second != want {
		t.Fatalf("request IDs = %q and %q, want %q", first, second, want)
	}
}

func TestClaimedIncidentPortabilityRejectsMissingJobsBeforePublication_Unit(t *testing.T) {
	baseline := moduleTestDependencies()
	tests := []struct {
		name   string
		want   string
		mutate func(*ModuleDependencies)
	}{
		{name: "PostgreSQL", want: "PostgreSQL", mutate: func(dependencies *ModuleDependencies) { dependencies.Postgres = nil }},
		{name: "Jobs transactions", want: "Jobs transaction", mutate: func(dependencies *ModuleDependencies) { dependencies.JobTransactions = nil }},
		{name: "Jobs operations", want: "Jobs operations", mutate: func(dependencies *ModuleDependencies) { dependencies.JobOperations = nil }},
		{name: "Jobs runner", want: "Jobs runner", mutate: func(dependencies *ModuleDependencies) { dependencies.JobRunner = nil }},
		{name: "storage", want: "storage", mutate: func(dependencies *ModuleDependencies) { dependencies.Storage = nil }},
		{name: "import finalizer", want: "import finalizer", mutate: func(dependencies *ModuleDependencies) { dependencies.ImportFinalizer = nil }},
		{name: "job finalizer", want: "job finalizer", mutate: func(dependencies *ModuleDependencies) { dependencies.JobFinalizer = nil }},
		{name: "portability", want: "portability", mutate: func(dependencies *ModuleDependencies) { dependencies.Portability = nil }},
		{name: "publication lock", want: "publication lock", mutate: func(dependencies *ModuleDependencies) { dependencies.IncidentPublicationLock = nil }},
		{name: "projection rebuild", want: "projection rebuilder", mutate: func(dependencies *ModuleDependencies) { dependencies.ProjectionRebuilder = nil }},
		{name: "source catalog", want: "source catalog", mutate: func(dependencies *ModuleDependencies) { dependencies.SourceCatalog = nil }},
		{name: "historical intents", want: "historical intent policy", mutate: func(dependencies *ModuleDependencies) { dependencies.HistoricalIntentPolicy = nil }},
		{name: "blob portability", want: "blob portability", mutate: func(dependencies *ModuleDependencies) { dependencies.BlobPortability = nil }},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			dependencies := baseline
			test.mutate(&dependencies)
			_, err := NewModule(dependencies)
			if err == nil || !strings.Contains(err.Error(), test.want) {
				t.Fatalf("missing %s error = %v; want %q", test.name, err, test.want)
			}
		})
	}
}

func moduleTestDependencies() ModuleDependencies {
	return ModuleDependencies{
		Postgres:                &pgxpool.Pool{},
		JobTransactions:         jobAdmissionStub{},
		JobOperations:           &recordingJobOperations{},
		JobRunner:               &registrationRunnerStub{},
		Storage:                 bundleStorageStub{},
		ImportFinalizer:         importFinalizerStub{},
		JobFinalizer:            jobFinalizerStub{},
		Portability:             &PortabilityOrchestrator{},
		IncidentPublicationLock: publicationLockStub{},
		ProjectionRebuilder:     projectionRebuilderStub{},
		SourceCatalog:           &sourceport.Catalog{},
		HistoricalIntentPolicy:  historicalIntentPolicyStub{},
		BlobPortability:         &recordingBlobPortability{},
	}
}

type importFinalizerStub struct{}

func (importFinalizerStub) FinalizeIncidentBundleImportTx(
	context.Context,
	pgx.Tx,
	importfinalizerport.Params,
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

type publicationLockStub struct{}

func (publicationLockStub) LockIncidentTx(context.Context, pgx.Tx, uuid.UUID) (bool, error) {
	return true, nil
}

type jobAdmissionStub struct{}

func (jobAdmissionStub) CreateQueuedTx(context.Context, pgx.Tx, jobs.EnqueueParams, time.Time) (jobs.Resource, error) {
	return jobs.Resource{}, nil
}

func (jobAdmissionStub) ValidateExecutionTx(context.Context, pgx.Tx, jobs.Execution) error {
	return nil
}

type historicalIntentPolicyStub struct{}

func (historicalIntentPolicyStub) SuppressTx(context.Context, pgx.Tx) error { return nil }

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
	if exportTransition.ResultSummary.Code != resultIncidentBundleExported || len(exportTransition.ResultSummary.ResourceRefs) != 1 {
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
	if importTransition.ResultSummary.Code != resultIncidentBundleImported || len(importTransition.ResultSummary.ResourceRefs) != 1 {
		t.Fatalf("import result summary mismatch: %#v", importTransition.ResultSummary)
	}
	importRef := importTransition.ResultSummary.ResourceRefs[0]
	if importRef.Kind != "incident" || importRef.ID != incidentID.String() || importRef.Route != "/api/v1/incidents/"+incidentID.String() {
		t.Fatalf("import resource ref mismatch: %#v", importRef)
	}

	t.Run("internal failures discard dependency values and retain safe diagnostics", func(t *testing.T) {
		sentinels := map[string]string{
			"persistence":  "SELECT secret_value FROM incident_bundle_jobs",
			"storage":      "s3://private-bucket/object?token=storage-secret",
			"finalization": "/var/lib/cartulary/private/finalization.sql",
			"upstream":     "upstream unauthorized: api_key=credential-secret",
		}
		for source, sentinel := range sentinels {
			t.Run(source, func(t *testing.T) {
				manager := &recordingJobOperations{}
				sink := incidentBundleJobResultSink{manager: manager}
				sink.completeFailedFromError(context.Background(), jobs.Execution{}, "internal_error", errors.New(sentinel))
				requireGenericInternalFailure(t, manager.failed, sentinel)
			})
		}

		hostile := failedCompletion("internal_error", map[string]any{
			"job_kind": "SELECT raw_job_kind FROM secrets",
		})
		requireGenericInternalFailure(t, &hostile, "raw_job_kind")

		manager := &recordingJobOperations{}
		worker := &incidentBundleWorker{results: incidentBundleJobResultSink{manager: manager}}
		worker.executePayload(context.Background(), jobs.Execution{}, jobPayload{JobKind: "credential-shaped-job-kind"})
		requireGenericInternalFailure(t, manager.failed, "credential-shaped-job-kind")

		manager = &recordingJobOperations{}
		storage := &recordingBundleStorage{}
		worker = &incidentBundleWorker{storage: storage, results: incidentBundleJobResultSink{manager: manager}}
		reference, err := ParseBundleStorageRef("incident-bundles/22222222-2222-2222-2222-222222222222.zip")
		if err != nil {
			t.Fatal(err)
		}
		finalizationSentinel := "constraint incident_bundle_secret_path /private/finalizer"
		worker.handleExportFinalizationError(context.Background(), jobs.Execution{}, reference, errors.New(finalizationSentinel))
		requireGenericInternalFailure(t, manager.failed, finalizationSentinel)

		jobID := uuid.MustParse("44444444-4444-4444-8444-444444444444")
		diagnosticErr := newWorkerDiagnosticError("payload_load_failed", jobID)
		var diagnostic *workerDiagnosticError
		if !errors.As(diagnosticErr, &diagnostic) || diagnostic.classification != "payload_load_failed" || diagnostic.jobID != jobID {
			t.Fatalf("worker diagnostic = %#v, %v", diagnostic, diagnosticErr)
		}
		hostileDiagnostic := newWorkerDiagnosticError("SELECT credential FROM secrets", jobID)
		if strings.Contains(hostileDiagnostic.Error(), "SELECT") || !strings.Contains(hostileDiagnostic.Error(), "class=unclassified_failure") || !strings.Contains(hostileDiagnostic.Error(), jobID.String()) {
			t.Fatalf("unsafe worker diagnostic = %q", hostileDiagnostic)
		}

		observeSentinel := "postgres://operator:password@database/cartulary"
		observeManager := &recordingJobOperations{observeErr: errors.New(observeSentinel)}
		worker = &incidentBundleWorker{jobManager: observeManager}
		if err := worker.executeJobID(context.Background(), jobs.Execution{}); err == nil ||
			strings.Contains(err.Error(), observeSentinel) || !strings.Contains(err.Error(), "class=execution_observation_failed") {
			t.Fatalf("worker observation diagnostic = %v", err)
		}
	})

	t.Run("blocked publication removes physical object and uses the closed failure", func(t *testing.T) {
		storage := &recordingBundleStorage{}
		manager := &recordingJobOperations{}
		worker := &incidentBundleWorker{
			storage: storage,
			results: incidentBundleJobResultSink{manager: manager},
		}
		reference, err := ParseBundleStorageRef("incident-bundles/22222222-2222-2222-2222-222222222222.zip")
		if err != nil {
			t.Fatal(err)
		}
		worker.handleExportFinalizationError(context.Background(), jobs.Execution{}, reference, newPortabilityFailure(ErrPortabilityBlocked, "a-profile"))
		if len(storage.removed) != 1 || storage.removed[0].String() != reference.String() {
			t.Fatalf("published object cleanup = %#v", storage.removed)
		}
		if manager.failed == nil || manager.failed.ErrorSummary.Code != "incident_bundle_export_rejected" || manager.failed.ErrorSummary.Retryable {
			t.Fatalf("blocked publication completion = %#v", manager.failed)
		}
		details := manager.failed.ErrorSummary.Details
		if len(details) != 2 || details["reason_code"] != "extension_state_not_portable" || details["profile_id"] != "a-profile" {
			t.Fatalf("blocked publication details = %#v", details)
		}
	})

	t.Run("prepared import cleanup uses the private blob consumer port", func(t *testing.T) {
		blob := &recordingBlobPortability{}
		prepared := &preparedImport{blobPort: blob, stagedObjectKeys: []string{"staged/one", "staged/two"}}
		prepared.cleanup(context.Background())
		if !slices.Equal(blob.cleaned, []string{"staged/one", "staged/two"}) || len(prepared.stagedObjectKeys) != 0 {
			t.Fatalf("private blob cleanup seam = cleaned %#v remaining %#v", blob.cleaned, prepared.stagedObjectKeys)
		}
	})
}

type recordingBundleStorage struct {
	removed        []BundleStorageRef
	staged         []byte
	stagedReads    int
	stagedRemovals int
}

func (*recordingBundleStorage) Stage(context.Context, string, []byte) (BundleStagingRef, error) {
	return BundleStagingRef{}, nil
}

func (*recordingBundleStorage) Publish(context.Context, string, []byte) (BundleStorageRef, error) {
	return BundleStorageRef{}, nil
}

func (s *recordingBundleStorage) ReadStaged(BundleStagingRef, int64) ([]byte, error) {
	s.stagedReads++
	return append([]byte(nil), s.staged...), nil
}

func (s *recordingBundleStorage) RemoveStaged(BundleStagingRef) error {
	s.stagedRemovals++
	return nil
}

func (s *recordingBundleStorage) RemovePublished(reference BundleStorageRef) error {
	s.removed = append(s.removed, reference)
	return nil
}

type recordingJobOperations struct {
	failed     *jobs.FailureCompletion
	observeErr error
}

type recordingBlobPortability struct {
	cleaned []string
}

var _ BlobPortability = (*recordingBlobPortability)(nil)

func (*recordingBlobPortability) ExportBlobFiles(context.Context, incidentportability.Queryer, uuid.UUID, map[string][]byte) error {
	return nil
}

func (*recordingBlobPortability) RewriteAndStageObjectBlobs(context.Context, map[string][]byte, uuid.UUID, uuid.UUID, incidentportability.AttributionRecorder) ([]byte, []string, error) {
	return []byte("[]\n"), []string{"staged/object"}, nil
}

func (p *recordingBlobPortability) CleanupStagedObjects(_ context.Context, keys []string) {
	p.cleaned = append(p.cleaned, keys...)
}

func (*recordingJobOperations) Get(context.Context, uuid.UUID) (jobs.Resource, error) {
	return jobs.Resource{}, nil
}

func (o *recordingJobOperations) ObserveExecution(context.Context, jobs.Execution) (jobs.Resource, error) {
	if o.observeErr != nil {
		return jobs.Resource{}, o.observeErr
	}
	return jobs.Resource{Status: jobs.StatusRunning}, nil
}

func (*recordingJobOperations) UpdateProgress(context.Context, jobs.Execution, jobs.Progress, *string) (jobs.Resource, error) {
	return jobs.Resource{}, nil
}

func (o *recordingJobOperations) CompleteFailed(_ context.Context, _ jobs.Execution, completion jobs.FailureCompletion) (jobs.Resource, error) {
	o.failed = &completion
	return jobs.Resource{}, nil
}

func (*recordingJobOperations) CompleteCanceled(context.Context, jobs.Execution, jobs.CancellationCompletion) (jobs.Resource, error) {
	return jobs.Resource{}, nil
}

func requireGenericInternalFailure(t testing.TB, completion *jobs.FailureCompletion, forbidden string) {
	t.Helper()
	if completion == nil || completion.ErrorSummary.Code != "internal_error" ||
		completion.ErrorSummary.Message != "internal_error" || completion.ErrorSummary.Retryable ||
		len(completion.ErrorSummary.Details) != 0 {
		t.Fatalf("internal failure completion = %#v", completion)
	}
	encoded, err := json.Marshal(completion)
	if err != nil {
		t.Fatalf("encode internal failure completion: %v", err)
	}
	if forbidden != "" && strings.Contains(string(encoded), forbidden) {
		t.Fatalf("internal failure disclosed %q: %s", forbidden, encoded)
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

	runner := &registrationRunnerStub{}
	worker.jobRunner = runner
	if err := worker.registerJobHandler(); err != nil {
		t.Fatalf("register named handler: %v", err)
	}
	if runner.name != bundleWorkerKind {
		t.Fatalf("registered handler name = %q; want %q", runner.name, bundleWorkerKind)
	}
	if err := worker.registerJobHandler(); !errors.Is(err, jobs.ErrHandlerAlreadyRegistered) {
		t.Fatalf("duplicate named handler error = %v; want ErrHandlerAlreadyRegistered", err)
	}
}

type registrationRunnerStub struct {
	name       string
	registered bool
}

func (runner *registrationRunnerStub) RegisterHandler(name string, _ jobs.HandlerFunc) error {
	if runner.registered {
		return jobs.ErrHandlerAlreadyRegistered
	}
	runner.name = name
	runner.registered = true
	return nil
}

func (*registrationRunnerStub) Notify(uuid.UUID) {}

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
	verificationErr, ok := err.(*verificationError)
	return ok && verificationErr.ReasonCode == reason
}

func TestDecodeImportMetadataUsesUploadHash_Unit(t *testing.T) {
	wantMediaTypes := []string{
		"application/zip",
		"application/x-tar",
		"application/gzip",
		"application/x-gzip",
		"application/octet-stream",
	}
	firstMediaTypes := acceptedIncidentBundleFileContentTypes()
	if !slices.Equal(firstMediaTypes, wantMediaTypes) {
		t.Fatalf("incident bundle upload media types = %#v; want %#v", firstMediaTypes, wantMediaTypes)
	}
	firstMediaTypes[0] = "application/x-mutated-by-caller"
	if laterMediaTypes := acceptedIncidentBundleFileContentTypes(); !slices.Equal(laterMediaTypes, wantMediaTypes) {
		t.Fatalf("incident bundle upload media types retained caller mutation: %#v", laterMediaTypes)
	}
	for _, contentType := range wantMediaTypes {
		t.Run(contentType, func(t *testing.T) {
			envelope, envelopeErr := httpapi.ParseUploadEnvelope(
				incidentBundleUploadEnvelopeRequest(t, contentType),
				httpapi.UploadEnvelopePolicy{FileContentTypes: acceptedIncidentBundleFileContentTypes()},
			)
			if envelopeErr != nil {
				t.Fatalf("exact media type %q was rejected: %v", contentType, envelopeErr)
			}
			if envelope.FileContentType != contentType {
				t.Fatalf("normalized media type = %q; want %q", envelope.FileContentType, contentType)
			}
		})
	}

	request, apiErr := decodeImportMetadata(httpapi.UploadEnvelope{
		Metadata: map[string]json.RawMessage{
			"client_txn_id": json.RawMessage(`"txn-import"`),
		},
		FileSHA256Hex: "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa",
	})
	if apiErr != nil {
		t.Fatalf("decodeImportMetadata returned error: %#v", apiErr)
	}
	var normalized map[string]any
	if err := json.Unmarshal(request.Normalized, &normalized); err != nil {
		t.Fatalf("decode normalized import request: %v", err)
	}
	if normalized["file_sha256"] != "aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa" {
		t.Fatalf("normalized import request must include file hash: %#v", normalized)
	}
	for _, field := range []string{"clone_mode", "merge_mode", "identifier_remap", "remote_fetch"} {
		_, apiErr := decodeImportMetadata(httpapi.UploadEnvelope{
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

func incidentBundleUploadEnvelopeRequest(t testing.TB, contentType string) *http.Request {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for _, part := range []struct {
		name        string
		filename    string
		contentType string
		body        string
	}{
		{name: "metadata", contentType: "application/json", body: `{"client_txn_id":"txn-import"}`},
		{name: "file", filename: "bundle.bin", contentType: contentType, body: "bundle"},
	} {
		header := textproto.MIMEHeader{}
		parameters := map[string]string{"name": part.name}
		if part.filename != "" {
			parameters["filename"] = part.filename
		}
		header.Set("Content-Disposition", mime.FormatMediaType("form-data", parameters))
		header.Set("Content-Type", part.contentType)
		partWriter, err := writer.CreatePart(header)
		if err != nil {
			t.Fatalf("create %s upload part: %v", part.name, err)
		}
		if _, err := io.WriteString(partWriter, part.body); err != nil {
			t.Fatalf("write %s upload part: %v", part.name, err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatalf("close incident bundle upload envelope: %v", err)
	}
	request := httptest.NewRequest(http.MethodPost, "/api/v1/incident-bundles/import", &body)
	request.Header.Set("Content-Type", writer.FormDataContentType())
	return request
}
