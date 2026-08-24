package incidentbundles_test

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport/asserttest"
	timelineroutetest "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport/routetest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestSupersededTimelineReplacementSurvivesImport_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	sourceHarness := runtime.StartDefaultServer(t, "extension_profile-incident-bundle-supersede-source")
	targetHarness := startIsolatedIncidentBundleServer(t, runtime, "extension_profile-incident-bundle-supersede-target")
	sourceAdmin, _ := flowtest.ProvisionBootstrapAdmin(t, sourceHarness.Server.HTTP.URL)
	targetAdmin, _ := flowtest.ProvisionBootstrapAdmin(t, targetHarness.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, sourceHarness.Server, sourceAdmin, map[string]any{
		"client_txn_id": "txn-incident-bundle-supersede-source",
		"incident_key":  "BUNDLE-SUPERSEDE",
		"title":         "Incident bundle supersede",
	})
	incidentID := incident["incident_id"].(string)
	replacement := timelineroutetest.CreateRow(t, sourceHarness.Server, sourceAdmin, incidentID, map[string]any{
		"client_txn_id":                   "txn-incident-bundle-supersede-replacement",
		"timeline.activity_synopsis_text": "Replacement event",
	})
	replacementID := replacement["row"].(map[string]any)["record_id"].(string)
	superseded := timelineroutetest.CreateRow(t, sourceHarness.Server, sourceAdmin, incidentID, map[string]any{
		"client_txn_id":                   "txn-incident-bundle-superseded-event",
		"timeline.activity_synopsis_text": "Superseded event",
	})
	supersededID := superseded["row"].(map[string]any)["record_id"].(string)

	supersedeResp := httptestx.DoJSON(t, http.MethodPost, sourceHarness.Server.HTTP.URL+"/api/v1/records/"+supersededID+"/supersede", map[string]any{
		"base_row_version":      1,
		"client_txn_id":         "txn-incident-bundle-supersede-action",
		"reason":                "Replacement preserves portability lineage",
		"replacement_record_id": replacementID,
	}, httptestx.WithCookies(sourceAdmin.SessionCookie, sourceAdmin.CSRFCookie), httptestx.WithHeader(authn.CSRFHeaderName, sourceAdmin.CSRFCookie.Value))
	supersedeData := httptestx.RequireSuccessEnvelope(t, supersedeResp, http.StatusOK)["data"].(map[string]any)
	if supersedeData["replacement_record_id"] != replacementID {
		t.Fatalf("source supersede response missing replacement id: %#v", supersedeData)
	}
	if got := asserttest.CountActiveSupersedesLinks(t, asserttest.SQLDatabase(sourceHarness.DB), incidentID, replacementID, supersededID); got != 1 {
		t.Fatalf("source supersedes link count: got %d want 1", got)
	}
	sourceRow := requireTimelineQueryRow(t, sourceHarness.Server, sourceAdmin, incidentID, supersededID)
	if got := timelineCellValue(t, sourceRow, "timeline.replacement_record_id"); got != replacementID {
		t.Fatalf("source query did not surface replacement id: got %#v want %s", got, replacementID)
	}

	bundleBytes := exportBundleBytes(t, sourceHarness, sourceAdmin, incidentID, "txn-export-superseded-timeline")
	importBundleAndWait(t, targetHarness.Server, targetAdmin, bundleBytes, "txn-import-superseded-timeline")

	if got := asserttest.CountActiveSupersedesLinks(t, asserttest.SQLDatabase(targetHarness.DB), incidentID, replacementID, supersededID); got != 1 {
		t.Fatalf("target supersedes link count: got %d want 1", got)
	}
	targetRow := requireTimelineQueryRow(t, targetHarness.Server, targetAdmin, incidentID, supersededID)
	if got := timelineCellValue(t, targetRow, "timeline.replacement_record_id"); got != replacementID {
		t.Fatalf("target query did not surface imported replacement id: got %#v want %s", got, replacementID)
	}
}

func TestIncidentBundlesEvidenceBlobStagingCleanup_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	sourceHarness := runtime.StartDefaultServer(t, "extension_profile-incident-bundle-failure-source")
	targetHarness := startIsolatedIncidentBundleServer(t, runtime, "extension_profile-incident-bundle-failure-target")
	sourceAdmin, sourceAdminID := flowtest.ProvisionBootstrapAdmin(t, sourceHarness.Server.HTTP.URL)
	targetAdmin, _ := flowtest.ProvisionBootstrapAdmin(t, targetHarness.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, sourceHarness.Server, sourceAdmin, map[string]any{
		"client_txn_id": "txn-incident-bundle-failure-source",
		"incident_key":  "BUNDLE-FAILURES",
		"title":         "Incident bundle failures",
	})
	incidentID := incident["incident_id"].(string)
	row := timelineroutetest.CreateRow(t, sourceHarness.Server, sourceAdmin, incidentID, map[string]any{
		"client_txn_id":                   "txn-incident-bundle-failure-row",
		"timeline.activity_synopsis_text": "Portable failure event",
	})
	recordID := row["row"].(map[string]any)["record_id"].(string)
	seedIncidentBundlePortableState(t, sourceHarness, incidentID, recordID, sourceAdminID)
	bundleBytes := exportBundleBytes(t, sourceHarness, sourceAdmin, incidentID, "txn-export-failure-fixture")
	blobPath := firstZipMemberWithPrefix(t, bundleBytes, "blobs/sha256/")

	t.Run("export missing blob", func(t *testing.T) {
		brokenHarness := startIsolatedIncidentBundleServer(t, runtime, "extension_profile-incident-bundle-export-missing-blob")
		brokenAdmin, brokenAdminID := flowtest.ProvisionBootstrapAdmin(t, brokenHarness.Server.HTTP.URL)
		brokenIncident := scenariotest.CreateIncident(t, brokenHarness.Server, brokenAdmin, map[string]any{
			"client_txn_id": "txn-incident-bundle-export-missing-blob",
			"incident_key":  "BUNDLE-MISSING-BLOB",
			"title":         "Incident bundle missing blob",
		})
		brokenIncidentID := brokenIncident["incident_id"].(string)
		seedMissingIncidentBundleBlob(t, brokenHarness, brokenIncidentID, brokenAdminID)
		job := httptestx.RequireSuccessEnvelope(t, postExport(t, brokenHarness.Server, brokenAdmin, map[string]any{
			"incident_id":   brokenIncidentID,
			"client_txn_id": "txn-export-missing-blob",
		}), http.StatusAccepted)["data"].(map[string]any)
		terminal := waitFailedJob(t, brokenHarness.Server, brokenAdmin, job["job_id"].(string))
		requireFailedJobReason(t, terminal, "incident_bundle_export_rejected", "missing_required_blob")
		if countRows(t, brokenHarness.DB, `SELECT count(*) FROM incident_bundle_exports WHERE export_job_id = $1`, job["job_id"].(string)) != 0 {
			t.Fatalf("failed missing-blob export must not publish a bundle descriptor")
		}
	})

	importCases := []struct {
		name       string
		txn        string
		bundle     []byte
		wantReason string
	}{
		{
			name:       "checksum mismatch",
			txn:        "txn-import-checksum-mismatch",
			bundle:     corruptZipMember(t, bundleBytes, "data/records.ndjson"),
			wantReason: "checksum_mismatch",
		},
		{
			name:       "missing blob",
			txn:        "txn-import-missing-blob",
			bundle:     removeZipMember(t, bundleBytes, blobPath),
			wantReason: "missing_required_blob",
		},
		{
			name:       "blob hash mismatch",
			txn:        "txn-import-blob-hash-mismatch",
			bundle:     replaceZipMemberAndChecksum(t, bundleBytes, blobPath, []byte("extension_profile wrong blob bytes\n")),
			wantReason: "blob_hash_mismatch",
		},
		{
			name: "malformed manifest",
			txn:  "txn-import-malformed-manifest",
			bundle: replaceZipMember(t, bundleBytes, "manifest.json", func(original []byte) []byte {
				var manifest map[string]any
				if err := json.Unmarshal(original, &manifest); err != nil {
					t.Fatalf("decode manifest: %v", err)
				}
				manifest["bundle_version"] = nil
				payload, err := json.Marshal(manifest)
				if err != nil {
					t.Fatalf("encode manifest: %v", err)
				}
				return append(payload, '\n')
			}),
			wantReason: "malformed_manifest",
		},
		{
			name: "unsupported bundle version",
			txn:  "txn-import-unsupported-bundle-version",
			bundle: replaceZipMember(t, bundleBytes, "manifest.json", func(original []byte) []byte {
				var manifest map[string]any
				if err := json.Unmarshal(original, &manifest); err != nil {
					t.Fatalf("decode manifest: %v", err)
				}
				manifest["bundle_version"] = float64(2)
				payload, err := json.Marshal(manifest)
				if err != nil {
					t.Fatalf("encode manifest: %v", err)
				}
				return append(payload, '\n')
			}),
			wantReason: "unsupported_bundle_version",
		},
		{
			name: "unsupported required capability",
			txn:  "txn-import-unsupported-required-capability",
			bundle: replaceZipMember(t, bundleBytes, "manifest.json", func(original []byte) []byte {
				var manifest map[string]any
				if err := json.Unmarshal(original, &manifest); err != nil {
					t.Fatalf("decode manifest: %v", err)
				}
				manifest["required_capabilities"] = []any{"snapshots"}
				payload, err := json.Marshal(manifest)
				if err != nil {
					t.Fatalf("encode manifest: %v", err)
				}
				return append(payload, '\n')
			}),
			wantReason: "extension_capability_not_supported",
		},
		{
			name:       "signature mismatch",
			txn:        "txn-import-signature-mismatch",
			bundle:     appendZipMember(t, bundleBytes, "integrity/signature.ed25519", []byte("not-a-supported-signature")),
			wantReason: "signature_mismatch",
		},
		{
			name: "unknown optional section",
			txn:  "txn-import-unknown-optional-section",
			bundle: replaceZipMember(t, bundleBytes, "manifest.json", func(original []byte) []byte {
				var manifest map[string]any
				if err := json.Unmarshal(original, &manifest); err != nil {
					t.Fatalf("decode manifest: %v", err)
				}
				manifest["optional_sections"] = []any{"unknown_section"}
				payload, err := json.Marshal(manifest)
				if err != nil {
					t.Fatalf("encode manifest: %v", err)
				}
				return append(payload, '\n')
			}),
			wantReason: "malformed_manifest",
		},
	}
	for _, tc := range importCases {
		t.Run(tc.name, func(t *testing.T) {
			assertImportFailureLeavesState(t, targetHarness, targetAdmin, incidentID, tc.txn, tc.bundle, tc.wantReason)
		})
	}

	t.Run("records invariant failures are closed and atomic", func(t *testing.T) {
		recordRows := decodeNDJSONRows(t, zipMemberMap(t, bundleBytes)["data/records.ndjson"])
		if len(recordRows) == 0 {
			t.Fatal("records invariant fixture requires at least one envelope")
		}
		incidentMismatchRows := append([]map[string]any(nil), recordRows...)
		incidentMismatchRows[0] = mapsClone(incidentMismatchRows[0])
		incidentMismatchRows[0]["incident_id"] = uuid.NewString()
		envelopeRows := append([]map[string]any(nil), recordRows...)
		envelopeRows[0] = mapsClone(envelopeRows[0])
		envelopeRows[0]["hostile_member_name"] = "SELECT secret FROM records"
		cases := []struct {
			name        string
			txn         string
			bundle      []byte
			invariantID string
			hostile     string
		}{
			{
				name: "incident scope",
				txn:  "txn-import-records-incident-scope",
				bundle: replaceStructuredBundleMember(
					t,
					bundleBytes,
					"data/records.ndjson",
					encodeNDJSONRows(t, incidentMismatchRows),
				),
				invariantID: "records.incident_scope",
			},
			{
				name: "envelope legal",
				txn:  "txn-import-records-envelope-legal",
				bundle: replaceStructuredBundleMember(
					t,
					bundleBytes,
					"data/records.ndjson",
					encodeNDJSONRows(t, envelopeRows),
				),
				invariantID: "records.envelope_legal",
				hostile:     "SELECT secret FROM records",
			},
			{
				name: "subtype complete",
				txn:  "txn-import-records-subtype-complete",
				bundle: replaceStructuredBundleMember(
					t,
					bundleBytes,
					"data/timeline_records.ndjson",
					nil,
				),
				invariantID: "records.subtype_complete",
			},
		}
		for _, testCase := range cases {
			t.Run(testCase.name, func(t *testing.T) {
				terminal := assertImportFailureLeavesState(
					t,
					targetHarness,
					targetAdmin,
					incidentID,
					testCase.txn,
					testCase.bundle,
					"source_family_invalid",
				)
				errorSummary := terminal["error_summary"].(map[string]any)
				details := errorSummary["details"].(map[string]any)
				if len(details) != 3 ||
					details["reason_code"] != "source_family_invalid" ||
					details["source_family_id"] != "records" ||
					details["invariant_id"] != testCase.invariantID {
					t.Fatalf("records failure details are not closed: %#v", details)
				}
				encoded, err := json.Marshal(terminal)
				if err != nil {
					t.Fatalf("encode records failure result: %v", err)
				}
				if testCase.hostile != "" &&
					strings.Contains(string(encoded), testCase.hostile) {
					t.Fatalf("records failure exposed hostile source content: %s", encoded)
				}
			})
		}
	})

	t.Run("safe directory entries import", func(t *testing.T) {
		directoryHarness := startIsolatedIncidentBundleServer(t, runtime, "extension_profile-incident-bundle-directory-import")
		directoryAdmin, _ := flowtest.ProvisionBootstrapAdmin(t, directoryHarness.Server.HTTP.URL)
		terminal := importBundleAndWait(t, directoryHarness.Server, directoryAdmin, appendZipDirectoryMembers(t, bundleBytes, "data/", "integrity/", "ext/"), "txn-import-safe-directories")
		if terminal["status"] != "succeeded" {
			encoded, _ := json.MarshalIndent(terminal, "", "  ")
			t.Fatalf("directory-bearing import did not succeed: %s", encoded)
		}
		summary := terminal["result_summary"].(map[string]any)
		if summary["code"] != "incident_bundle_imported" {
			t.Fatalf("directory-bearing import summary mismatch: %#v", summary)
		}
		refs := summary["resource_refs"].([]any)
		if len(refs) != 1 || refs[0].(map[string]any)["id"] != incidentID {
			t.Fatalf("directory-bearing import refs mismatch: %#v", refs)
		}
		assertNoIncidentBundleStaging(t, directoryHarness.Server)
	})

	t.Run("archive extracted byte limit", func(t *testing.T) {
		limitHarness := startIsolatedIncidentBundleServerWithEnv(t, runtime, "extension_profile-incident-bundle-extracted-limit", map[string]string{
			"CARTULARY__LIMITS__INCIDENT_BUNDLES__MAX_EXTRACTED_BYTES": "1",
		})
		limitAdmin, _ := flowtest.ProvisionBootstrapAdmin(t, limitHarness.Server.HTTP.URL)
		assertImportFailureLeavesState(t, limitHarness, limitAdmin, incidentID, "txn-import-extracted-limit", bundleBytes, "archive_extracted_bytes_exceeded")
	})
	t.Run("archive member count limit", func(t *testing.T) {
		limitHarness := startIsolatedIncidentBundleServerWithEnv(t, runtime, "extension_profile-incident-bundle-member-limit", map[string]string{
			"CARTULARY__LIMITS__ARCHIVES__MAX_MEMBERS": "20",
		})
		limitAdmin, _ := flowtest.ProvisionBootstrapAdmin(t, limitHarness.Server.HTTP.URL)
		assertImportFailureLeavesState(t, limitHarness, limitAdmin, incidentID, "txn-import-member-limit", bundleBytes, "archive_member_count_exceeded")
	})

	importBundleAndWait(t, targetHarness.Server, targetAdmin, bundleBytes, "txn-import-duplicate-baseline")
	assertImportFailureLeavesState(t, targetHarness, targetAdmin, incidentID, "txn-import-duplicate-incident", bundleBytes, "duplicate_incident_id")
}
