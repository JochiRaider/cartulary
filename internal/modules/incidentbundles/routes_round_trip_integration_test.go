package incidentbundles_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	referencetest "github.com/JochiRaider/cartulary/internal/modules/reference_data/testsupport"
	referencehttp "github.com/JochiRaider/cartulary/internal/modules/reference_data/testsupport/fixturehttp"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/evidence/blobref"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport/asserttest"
	timelineroutetest "github.com/JochiRaider/cartulary/internal/modules/timeline/testsupport/routetest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestIncidentReferenceCatalogSurvivesUnavailablePackRoundTrip_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	t.Run("historical reporting artifacts", func(t *testing.T) { testReportingArtifactRoundTrip(t, runtime) })
	source := runtime.StartServer(t, appsupport.ServerOptions{Prefix: "reference-catalog-source", TestRouteMode: httptestx.TestRouteModeDisabled})
	t.Run("embedded destination verification", func(t *testing.T) {
		testEmbeddedReferenceDestinationVerification(t, runtime, startIsolatedIncidentBundleServer(t, runtime, "embedded-reference-source"))
	})
	target := startIsolatedIncidentBundleServer(t, runtime, "reference-catalog-target")
	sourceAdmin, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, source.Server.HTTP.URL)
	targetAdmin, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, target.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, source.Server, sourceAdmin, map[string]any{"client_txn_id": "reference-catalog-incident", "incident_key": "REFERENCE-CATALOG", "title": "Portable reference history"})
	id := incident["incident_id"].(string)
	container := exportBundleBytes(t, source, sourceAdmin, id, "reference-catalog-export")
	fixture, err := os.ReadFile("../../../contracts/reference-pack-fixtures/fixtures/portable-references.v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var fixtureValue map[string]any
	if err := json.Unmarshal(fixture, &fixtureValue); err != nil {
		t.Fatal(err)
	}
	wantReferences, err := json.Marshal(fixtureValue)
	if err != nil {
		t.Fatal(err)
	}
	wantReferences = append(wantReferences, '\n')
	container = replaceStructuredBundleMember(t, container, "data/reference_pack_refs.json", wantReferences)
	importBundleAndWait(t, target.Server, targetAdmin, container, "reference-catalog-import")
	reexported := exportBundleBytes(t, target, targetAdmin, id, "reference-catalog-reexport")
	if actual := zipMemberBytes(t, reexported, "data/reference_pack_refs.json"); !bytes.Equal(actual, wantReferences) {
		t.Fatalf("unavailable references differ from independent fixture: got=%s want=%s", actual, wantReferences)
	}
}

type roundTripPackFixture struct {
	referencetest.PortablePackFixture
	SetSHA256 string
}

func runNativeSnapshotRoundTrip(t *testing.T, fixture roundTripPackFixture, embedAllowed bool) {
	runtime := appsupport.StartRuntime(t)
	bootstrap := filepath.Join(t.TempDir(), "bootstrap.json")
	if err := os.WriteFile(bootstrap, fixture.Bootstrap, 0600); err != nil {
		t.Fatal(err)
	}
	source := runtime.StartServer(t, appsupport.ServerOptions{Prefix: "native-reference-source", TestRouteMode: httptestx.TestRouteModeDisabled, Env: map[string]string{"CARTULARY__REFERENCE_PACKS__TRUST_BOOTSTRAP_PATH": bootstrap, "CARTULARY__REFERENCE_PACKS__CLOCK_TRUSTED": "true"}})
	target := startIsolatedIncidentBundleServerWithEnv(t, runtime, "native-reference-target", map[string]string{"CARTULARY__REFERENCE_PACKS__TRUST_BOOTSTRAP_PATH": bootstrap, "CARTULARY__REFERENCE_PACKS__CLOCK_TRUSTED": "true"})
	sourceAdmin, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, source.Server.HTTP.URL)
	targetAdmin, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, target.Server.HTTP.URL)
	referencehttp.ImportPackFixture(t, source.Server, sourceAdmin, fixture.PortablePackFixture, "native-pack-import")
	httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, source.Server.HTTP.URL+"/api/v1/reference-packs/"+fixture.Key+"/"+fixture.Version+"/activate", map[string]any{"client_txn_id": "native-pack-activate"}, httptestx.WithCookies(sourceAdmin.Session, sourceAdmin.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, sourceAdmin.CSRF.Value)), http.StatusOK)
	incident := scenariotest.CreateIncident(t, source.Server, sourceAdmin, map[string]any{"client_txn_id": "native-reference-incident", "incident_key": "NATIVE-REFERENCES", "title": "  Retained cafe\u0301 ", "description": " Start\r\n\t\u200d\u2028\u2029" + strings.Repeat("😀", 16370) + " end ", "severity": " high "})
	id := incident["incident_id"].(string)
	removedActor := flowtest.SeedActor(t, source.DB, flowtest.ActorAttributes{Email: "portable-removed@example.test", DisplayName: "Removed member"})
	scenariotest.CreateMembership(t, source.Server, sourceAdmin, id, map[string]any{"client_txn_id": "portable-removed", "user_id": removedActor.ID.String(), "role": "viewer"})
	removed := httptestx.DoJSON(t, http.DefaultClient, http.MethodDelete, source.Server.HTTP.URL+"/api/v1/incidents/"+id+"/memberships/"+removedActor.ID.String(), map[string]any{"base_membership_version": 1}, httptestx.WithCookies(sourceAdmin.Session, sourceAdmin.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, sourceAdmin.CSRF.Value))
	if removed.StatusCode != 204 {
		t.Fatalf("remove source membership: %d", removed.StatusCode)
	}
	_ = removed.Body.Close()

	for _, txn := range []string{"snapshot-one", "snapshot-two"} {
		response := httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, source.Server.HTTP.URL+"/api/v1/snapshots", map[string]any{"incident_id": id, "client_txn_id": txn}, httptestx.WithCookies(sourceAdmin.Session, sourceAdmin.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, sourceAdmin.CSRF.Value))
		job := httptestx.RequireSuccessEnvelope(t, response, http.StatusAccepted)["data"].(map[string]any)
		waitJob(t, source.Server, sourceAdmin, job["job_id"].(string))
	}
	embedded := exportBundleBytes(t, source, sourceAdmin, id, "native-pack-embedding-refusal", "embedded")
	if included, ok := zipMemberMap(t, embedded)[fixture.Path]; embedAllowed {
		if !ok || !bytes.Equal(included, fixture.Container) {
			t.Fatal("allowed pack container not embedded exactly")
		}
	} else if ok {
		t.Fatal("restricted pack bytes embedded")
	}

	container := exportBundleBytes(t, source, sourceAdmin, id, "native-reference-export")
	if embedAllowed {
		container = embedded
	}
	exported := zipMemberBytes(t, container, "data/reference_pack_refs.json")
	refs, err := reference_data.DecodeIncidentBundleReferences(exported)
	if err != nil || len(refs.Sets) != 1 || len(refs.Versions) != 4 {
		t.Fatal("native snapshot bindings absent or duplicated", string(exported), err)
	}
	var pinned string
	if err := source.DB.QueryRow(`SELECT p.snapshot_admission_json->'reference_packs'->>'pack_set_id' FROM reporting_snapshots s JOIN reporting_job_payloads p ON p.job_id=s.create_job_id WHERE s.incident_id=$1 ORDER BY s.snapshot_id LIMIT 1`, id).Scan(&pinned); err != nil {
		t.Fatal(err)
	}
	if refs.Sets[0].ID != pinned || pinned != "rpset_"+fixture.SetSHA256 {
		t.Fatal("native binding substituted")
	}

	var sourceRow map[string]any
	if err := json.Unmarshal(zipMemberBytes(t, container, "data/incident.json"), &sourceRow); err != nil {
		t.Fatal(err)
	}
	if len(sourceRow) != 16 {
		t.Fatalf("portable incident shape=%d", len(sourceRow))
	}
	sourceRow["title"] = " padded title "
	malformed, _ := json.Marshal(sourceRow)
	assertImportFailureLeavesState(t, target, targetAdmin, id, "noncanonical-incident-source", replaceStructuredBundleMember(t, container, "data/incident.json", malformed), "source_family_invalid")
	importBundleAndWait(t, target.Server, targetAdmin, container, "native-reference-import")
	roundTrip := exportBundleBytes(t, target, targetAdmin, id, "native-reference-reexport")
	if embedAllowed {
		roundTrip = exportBundleBytes(t, target, targetAdmin, id, "pack-embedded-reexport", "embedded")
		if !bytes.Equal(zipMemberBytes(t, roundTrip, fixture.Path), fixture.Container) {
			t.Fatal("destination changed retained pack bytes")
		}
		var health, current string
		if err := target.DB.QueryRow(`SELECT c.health,s.pack_set_id FROM reference_pack_candidates c CROSS JOIN reference_pack_current_set s WHERE c.pack_key=$1 AND c.pack_version=$2`, fixture.Key, fixture.Version).Scan(&health, &current); err != nil || health != "verified_available" || current == "rpset_"+fixture.SetSHA256 {
			t.Fatal("destination verification or explicit activation", health, current, err)
		}
	}
	var restoredRow map[string]any
	if err := json.Unmarshal(zipMemberBytes(t, roundTrip, "data/incident.json"), &restoredRow); err != nil {
		t.Fatal(err)
	}
	if len(restoredRow) != 16 || restoredRow["title"] != incident["title"] || restoredRow["description"] != incident["description"] || restoredRow["severity"] != incident["severity"] {
		t.Fatal("normalized source fields did not round trip")
	}
	var histories int
	if err := target.DB.QueryRow(`SELECT count(*) FROM incident_membership_versions WHERE incident_id=$1`, id).Scan(&histories); err != nil || histories != 1 {
		t.Fatalf("import bootstrap history=%d err=%v", histories, err)
	}

	if !bytes.Equal(zipMemberBytes(t, roundTrip, "data/reference_pack_refs.json"), exported) {
		t.Fatal("native references changed after import")
	}
}

func TestSupersededTimelineReplacementSurvivesImport_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	sourceHarness := runtime.StartServer(t, appsupport.ServerOptions{Prefix: "extension_profile-incident-bundle-supersede-source", TestRouteMode: httptestx.TestRouteModeDisabled})
	targetHarness := startIsolatedIncidentBundleServer(t, runtime, "extension_profile-incident-bundle-supersede-target")
	sourceAdmin, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, sourceHarness.Server.HTTP.URL)
	targetAdmin, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, targetHarness.Server.HTTP.URL)
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

	supersedeResp := httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, sourceHarness.Server.HTTP.URL+"/api/v1/records/"+supersededID+"/supersede", map[string]any{
		"base_row_version":      1,
		"client_txn_id":         "txn-incident-bundle-supersede-action",
		"reason":                "Replacement preserves portability lineage",
		"replacement_record_id": replacementID,
	}, httptestx.WithCookies(sourceAdmin.Session, sourceAdmin.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, sourceAdmin.CSRF.Value))
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
	sourceHarness := runtime.StartServer(t, appsupport.ServerOptions{Prefix: "extension_profile-incident-bundle-failure-source", TestRouteMode: httptestx.TestRouteModeDisabled})
	targetHarness := startIsolatedIncidentBundleServer(t, runtime, "extension_profile-incident-bundle-failure-target")
	sourceAdmin, sourceAdminID := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, sourceHarness.Server.HTTP.URL)
	targetAdmin, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, targetHarness.Server.HTTP.URL)
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
	seededState := seedIncidentBundlePortableState(t, sourceHarness, incidentID, recordID, sourceAdminID.String())
	bundleBytes := exportBundleBytes(t, sourceHarness, sourceAdmin, incidentID, "txn-export-failure-fixture")
	blobPath := firstZipMemberWithPrefix(t, bundleBytes, "blobs/sha256/")
	assertPortableBlobShape := func(t *testing.T, archive []byte, count int) []map[string]any {
		t.Helper()
		rows := decodeNDJSONRows(t, zipMemberBytes(t, archive, "data/object_blobs.ndjson"))
		if len(rows) != count {
			t.Fatalf("portable blob rows=%d want=%d", len(rows), count)
		}
		want := strings.Fields("object_blob_id incident_id created_by_user_id upload_state byte_size filename_hint content_type_hint expected_sha256_hex observed_size observed_content_type observed_sha256_hex target_expires_at pending_expires_at finalized_at terminal_reason failed_at cleanup_due_at cleaned_up_at created_at updated_at")
		for _, row := range rows {
			if len(row) != len(want) {
				t.Fatalf("portable blob source field count=%d want=%d", len(row), len(want))
			}
			for _, field := range want {
				if _, exists := row[field]; !exists {
					t.Fatalf("portable blob source field %s missing", field)
				}
			}
			if _, exists := row["storage_key"]; exists {
				t.Fatal("physical blob key escaped into portable source")
			}
		}
		return rows
	}
	assertPortableBlobShape(t, bundleBytes, 1)

	for _, row := range decodeNDJSONRows(t, zipMemberBytes(t, bundleBytes, "data/evidence_custody_events.ndjson")) {
		want := strings.Fields("custody_event_id incident_id evidence_record_id custody_event_type actor_user_id occurred_at location_text note metadata")
		if len(row) != len(want) {
			t.Fatalf("portable custody fields=%d want=%d", len(row), len(want))
		}
		for _, field := range want {
			if _, ok := row[field]; !ok {
				t.Fatalf("portable custody field %s missing", field)
			}
		}
	}
	t.Run("typed portable custody history", func(t *testing.T) {
		for index, tc := range []struct {
			name, field, invariant, family string
			value                          any
			missing                        bool
		}{
			{name: "missing nullable note", field: "note", invariant: "row_shape_exact", missing: true},
			{name: "unknown member", field: "unrecognized", invariant: "row_shape_exact", value: true},
			{name: "metadata array", field: "metadata", invariant: "row_shape_exact", value: []any{}},
			{name: "metadata NUL", field: "metadata", invariant: "row_shape_exact", value: map[string]any{"context": "bad\x00text"}},
			{name: "lossy occurrence", field: "occurred_at", invariant: "custody_ordered", value: "2024-01-01T00:00:00.1234567Z"},
			{name: "unknown event", field: "custody_event_type", invariant: "row_shape_exact", value: "unknown"},
			{name: "other incident", field: "incident_id", invariant: "custody_same_incident", value: uuid.NewString()},
			{name: "missing Evidence", field: "evidence_record_id", invariant: "custody_same_incident", value: uuid.NewString()},
			{name: "missing historical actor", field: "actor_user_id", invariant: "reference_complete", family: "actors", value: uuid.NewString()},
		} {
			t.Run(tc.name, func(t *testing.T) {
				rows := decodeNDJSONRows(t, zipMemberBytes(t, bundleBytes, "data/evidence_custody_events.ndjson"))
				if len(rows) != 1 {
					t.Fatalf("custody fixture rows=%d", len(rows))
				}
				if tc.missing {
					delete(rows[0], tc.field)
				} else {
					rows[0][tc.field] = tc.value
				}
				candidate := replaceStructuredBundleMember(t, bundleBytes, "data/evidence_custody_events.ndjson", encodeNDJSONRows(t, rows))
				failed := assertImportFailureLeavesState(t, targetHarness, targetAdmin, incidentID, fmt.Sprintf("typed-custody-%d", index), candidate, "source_family_invalid")
				details := failed["error_summary"].(map[string]any)["details"].(map[string]any)
				family := tc.family
				if family == "" {
					family = "evidence"
				}
				if len(details) != 3 || details["reason_code"] != "source_family_invalid" || details["source_family_id"] != family || details["invariant_id"] != family+"."+tc.invariant {
					t.Fatalf("typed custody failure=%#v", details)
				}
			})
		}
	})

	t.Run("typed portable blob metadata", func(t *testing.T) {
		for index, tc := range []struct {
			name, field, invariant string
			value                  any
			missing                bool
		}{
			{"missing nullable hint", "filename_hint", "row_shape_exact", nil, true},
			{"unknown member", "unrecognized", "row_shape_exact", true, false},
			{"noncanonical hint", "filename_hint", "row_shape_exact", " padded", false},
			{"mistyped observed content type", "observed_content_type", "row_shape_exact", true, false},
			{"lossy expiry", "target_expires_at", "row_shape_exact", "2026-01-01T00:00:00.1234567Z", false},
			{"null finalization", "finalized_at", "lifecycle_legal", nil, false},
			{"unknown state", "upload_state", "lifecycle_legal", "unknown", false},
			{"finalized failure reason", "terminal_reason", "lifecycle_legal", "pending_timeout", false},
		} {
			t.Run(tc.name, func(t *testing.T) {
				rows := decodeNDJSONRows(t, zipMemberBytes(t, bundleBytes, "data/object_blobs.ndjson"))
				if tc.missing {
					delete(rows[0], tc.field)
				} else {
					rows[0][tc.field] = tc.value
				}
				candidate := replaceStructuredBundleMember(t, bundleBytes, "data/object_blobs.ndjson", encodeNDJSONRows(t, rows))
				failed := assertImportFailureLeavesState(t, targetHarness, targetAdmin, incidentID, fmt.Sprintf("typed-blob-%d", index), candidate, "source_family_invalid")
				details := failed["error_summary"].(map[string]any)["details"].(map[string]any)
				if len(details) != 3 || details["reason_code"] != "source_family_invalid" || details["source_family_id"] != "evidence" || details["invariant_id"] != "evidence."+tc.invariant {
					t.Fatalf("typed Blob metadata failure=%#v", details)
				}
			})
		}
	})

	t.Run("archive cannot supply physical blob keys", func(t *testing.T) {
		for _, state := range []string{"pending", "failed", "available", "quarantined"} {
			for index, key := range []any{nil, "", "incidents/source/object-blobs/source/objects/source", map[string]any{"key": "untrusted"}} {
				t.Run(fmt.Sprintf("%s/%d", state, index), func(t *testing.T) {
					rows := decodeNDJSONRows(t, zipMemberBytes(t, bundleBytes, "data/object_blobs.ndjson"))
					rows[0]["upload_state"], rows[0]["storage_key"] = state, key
					candidate := replaceStructuredBundleMember(t, bundleBytes, "data/object_blobs.ndjson", encodeNDJSONRows(t, rows))
					failed := assertImportFailureLeavesState(t, targetHarness, targetAdmin, incidentID, fmt.Sprintf("physical-key-%s-%d", state, index), candidate, "source_family_invalid")
					details := failed["error_summary"].(map[string]any)["details"].(map[string]any)
					if len(details) != 3 || details["reason_code"] != "source_family_invalid" || details["source_family_id"] != "evidence" || details["invariant_id"] != "evidence.storage_reference_legal" {
						t.Fatalf("physical key admission lost owner: %#v", details)
					}
				})
			}
		}
	})

	t.Run("Evidence record typed admission", func(t *testing.T) {
		cases := []struct {
			name, field, invariant, family string
			value                          any
			missing                        bool
		}{
			{name: "missing nullable owner", field: "owner_user_id", invariant: "row_shape_exact", missing: true},
			{name: "missing title", field: "title", invariant: "row_shape_exact", missing: true},
			{name: "unknown field", field: "unrecognized", value: true, invariant: "row_shape_exact"},
			{name: "unknown actor column", field: "extra_user_id", value: sourceAdminID.String(), invariant: "row_shape_exact"},
			{name: "numeric text", field: "title", value: 42, invariant: "row_shape_exact"},
			{name: "NUL attachment hash", field: "blob_hash", value: "bad\x00digest", invariant: "row_shape_exact"},
			{name: "missing collector Party", field: "collector_party_id", value: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", invariant: "references_same_incident"},
			{name: "missing source Party", field: "source_party_id", value: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", invariant: "references_same_incident"},
			{name: "noncanonical text", field: "title", value: " padded", invariant: "row_shape_exact"},
			{name: "lossy timestamp", field: "updated_at", value: "2026-01-01T00:00:00.1234567Z", invariant: "row_shape_exact"},
			{name: "nil identity", field: "record_id", value: uuid.Nil.String(), invariant: "source_identity_admitted"},
			{name: "other incident", field: "incident_id", value: uuid.NewString(), invariant: "envelope_type_scope"},
			{name: "unknown lifecycle", field: "lifecycle_state", value: "unrecognized", invariant: "lifecycle_legal"},
			{name: "logical object mismatch", field: "storage_ref", value: "object://10000000-0000-4000-8000-000000000001", invariant: "storage_reference_legal"},
			{name: "missing historical owner", field: "owner_user_id", value: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", invariant: "reference_complete", family: "actors"},
		}
		for index, tc := range cases {
			t.Run(tc.name, func(t *testing.T) {
				rows := decodeNDJSONRows(t, zipMemberBytes(t, bundleBytes, "data/evidence_records.ndjson"))
				if tc.missing {
					delete(rows[0], tc.field)
				} else {
					rows[0][tc.field] = tc.value
				}
				candidate := replaceStructuredBundleMember(t, bundleBytes, "data/evidence_records.ndjson", encodeNDJSONRows(t, rows))
				failed := assertImportFailureLeavesState(t, targetHarness, targetAdmin, incidentID, fmt.Sprintf("typed-evidence-%d", index), candidate, "source_family_invalid")
				details := failed["error_summary"].(map[string]any)["details"].(map[string]any)
				family := tc.family
				if family == "" {
					family = "evidence"
				}
				if len(details) != 3 || details["reason_code"] != "source_family_invalid" || details["source_family_id"] != family || details["invariant_id"] != family+"."+tc.invariant {
					t.Fatalf("Evidence typed failure=%#v", details)
				}
			})
		}
	})

	t.Run("Evidence associations use authoritative Blob state", func(t *testing.T) {
		for _, variant := range []string{"missing", "quarantined bridge"} {
			t.Run(variant, func(t *testing.T) {
				rows := decodeNDJSONRows(t, zipMemberBytes(t, bundleBytes, "data/evidence_records.ndjson"))
				rows[0]["lifecycle_state"], rows[0]["upload_state"] = "available", "available"
				candidate := bundleBytes
				invariant := "lifecycle_legal"
				if variant == "missing" {
					rows[0]["object_blob_id"], rows[0]["storage_ref"] = "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", nil
					invariant = "object_metadata_agree"
				} else {
					blobs := decodeNDJSONRows(t, zipMemberBytes(t, candidate, "data/object_blobs.ndjson"))
					blobs[0]["upload_state"] = "quarantined"
					candidate = replaceStructuredBundleMember(t, candidate, "data/object_blobs.ndjson", encodeNDJSONRows(t, blobs))
				}
				candidate = replaceStructuredBundleMember(t, candidate, "data/evidence_records.ndjson", encodeNDJSONRows(t, rows))
				failed := assertImportFailureLeavesState(t, targetHarness, targetAdmin, incidentID, "blob-association-"+variant, candidate, "source_family_invalid")
				details := failed["error_summary"].(map[string]any)["details"].(map[string]any)
				if len(details) != 3 || details["reason_code"] != "source_family_invalid" || details["source_family_id"] != "evidence" || details["invariant_id"] != "evidence."+invariant {
					t.Fatalf("Blob association failure=%#v", details)
				}
			})
		}
	})

	t.Run("export finalized observation disagreement", func(t *testing.T) {
		// Arrange invalid retained metadata; public commands do not create these states.
		blobBytes := []byte("extension_profile incident bundle blob\n")
		digest := hashHexBytes(blobBytes)
		reset := func(t *testing.T) {
			t.Helper()
			if _, err := sourceHarness.DB.Exec(`UPDATE object_blobs
SET upload_state='available', byte_size=$2, observed_size=$2,
    observed_sha256_hex=$3, expected_sha256_hex=$3 WHERE incident_id=$1`, incidentID, len(blobBytes), digest); err != nil {
				t.Fatal(err)
			}
		}
		var originalID, originalKey string
		if err := sourceHarness.DB.QueryRow(`SELECT object_blob_id, storage_key FROM object_blobs WHERE incident_id=$1`, incidentID).Scan(&originalID, &originalKey); err != nil {
			t.Fatal(err)
		}
		cases := []struct {
			name, update string
			duplicate    bool
		}{
			{"absent observed digest", "observed_sha256_hex=NULL", false},
			{"absent observed size", "observed_size=NULL", false},
			{"negative observed size", "observed_size=-1", false},
			{"declared observed size disagreement", "observed_size=byte_size+1", false},
			{"expected observed digest disagreement", "expected_sha256_hex=repeat('0',64)", false},
			{"actual member size disagreement", "byte_size=byte_size+1, observed_size=observed_size+1", false},
			{"duplicate digest size disagreement", "byte_size=byte_size+1, observed_size=observed_size+1", true},
		}
		for _, state := range []string{"available", "quarantined"} {
			for index, tc := range cases {
				t.Run(state+"/"+tc.name, func(t *testing.T) {
					reset(t)
					t.Cleanup(func() { reset(t) })
					targetID := originalID
					if tc.duplicate {
						targetID = uuid.NewString()
						duplicateKey := originalKey + "/export-duplicate-" + targetID
						if err := sourceHarness.ObjectStore.PutObject(context.Background(), duplicateKey, bytes.NewReader(blobBytes), int64(len(blobBytes)), "text/plain"); err != nil {
							t.Fatal(err)
						}
						if _, err := sourceHarness.DB.Exec(`INSERT INTO object_blobs
(object_blob_id,incident_id,created_by_user_id,storage_key,upload_state,byte_size,observed_size,observed_sha256_hex,expected_sha256_hex,target_expires_at,pending_expires_at,finalized_at,created_at,updated_at)
SELECT $2,incident_id,created_by_user_id,$3,$4,byte_size,observed_size,observed_sha256_hex,expected_sha256_hex,target_expires_at,pending_expires_at,finalized_at,created_at,updated_at FROM object_blobs WHERE object_blob_id=$1`, originalID, targetID, duplicateKey, state); err != nil {
							t.Fatal(err)
						}
						t.Cleanup(func() {
							if _, err := sourceHarness.DB.Exec(`DELETE FROM object_blobs WHERE object_blob_id=$1`, targetID); err != nil {
								t.Fatal(err)
							}
						})
						valid := exportBundleBytes(t, sourceHarness, sourceAdmin, incidentID, "export-valid-duplicate-"+state)
						assertPortableBlobShape(t, valid, 2)
						blobCount := 0
						for path, payload := range zipMemberMap(t, valid) {
							if strings.HasPrefix(path, "blobs/sha256/") {
								blobCount++
								if path != "blobs/sha256/"+digest || !bytes.Equal(payload, blobBytes) {
									t.Fatal("deduplicated finalized bytes changed")
								}
							}
						}
						if blobCount != 1 {
							t.Fatalf("deduplicated blob members=%d", blobCount)
						}
					}
					if _, err := sourceHarness.DB.Exec(`UPDATE object_blobs SET upload_state=$2, `+tc.update+` WHERE incident_id=$1 AND object_blob_id=$3`, incidentID, state, targetID); err != nil {
						t.Fatal(err)
					}
					job := httptestx.RequireSuccessEnvelope(t, postExport(t, sourceHarness.Server, sourceAdmin, map[string]any{
						"incident_id":   incidentID,
						"client_txn_id": fmt.Sprintf("export-observation-%s-%d", state, index),
					}), http.StatusAccepted)["data"].(map[string]any)
					jobID := job["job_id"].(string)
					terminal := waitFailedJob(t, sourceHarness.Server, sourceAdmin, jobID)
					requireFailedJobReason(t, terminal, "incident_bundle_export_rejected", "source_family_invalid")
					details := terminal["error_summary"].(map[string]any)["details"].(map[string]any)
					if len(details) != 3 || details["reason_code"] != "source_family_invalid" || details["source_family_id"] != "evidence" || details["invariant_id"] != "evidence.byte_size_digest_agree" {
						t.Fatalf("export observation disagreement lost owner: %#v", details)
					}
					if countRows(t, sourceHarness.DB, `SELECT count(*) FROM incident_bundle_exports WHERE export_job_id=$1`, jobID) != 0 {
						t.Fatal("invalid finalized metadata published a bundle descriptor")
					}
					assertNoIncidentBundleStaging(t, sourceHarness.Server)
				})
			}
		}
	})

	t.Run("export missing blob", func(t *testing.T) {
		brokenHarness := startIsolatedIncidentBundleServer(t, runtime, "extension_profile-incident-bundle-export-missing-blob")
		brokenAdmin, brokenAdminID := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, brokenHarness.Server.HTTP.URL)
		brokenIncident := scenariotest.CreateIncident(t, brokenHarness.Server, brokenAdmin, map[string]any{
			"client_txn_id": "txn-incident-bundle-export-missing-blob",
			"incident_key":  "BUNDLE-MISSING-BLOB",
			"title":         "Incident bundle missing blob",
		})
		brokenIncidentID := brokenIncident["incident_id"].(string)
		seedMissingIncidentBundleBlob(t, brokenHarness, brokenIncidentID, brokenAdminID.String())
		for _, state := range []string{"available", "quarantined"} {
			if _, err := brokenHarness.DB.Exec(`UPDATE object_blobs SET upload_state=$2 WHERE incident_id=$1`, brokenIncidentID, state); err != nil {
				t.Fatal(err)
			}
			job := httptestx.RequireSuccessEnvelope(t, postExport(t, brokenHarness.Server, brokenAdmin, map[string]any{
				"incident_id":   brokenIncidentID,
				"client_txn_id": "txn-export-missing-blob-" + state,
			}), http.StatusAccepted)["data"].(map[string]any)
			terminal := waitFailedJob(t, brokenHarness.Server, brokenAdmin, job["job_id"].(string))
			requireFailedJobReason(t, terminal, "incident_bundle_export_rejected", "missing_required_blob")
			if countRows(t, brokenHarness.DB, `SELECT count(*) FROM incident_bundle_exports WHERE export_job_id = $1`, job["job_id"].(string)) != 0 {
				t.Fatalf("failed %s missing-blob export published a bundle descriptor", state)
			}
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

	t.Run("blob observations agree before staging", func(t *testing.T) {
		rows := decodeNDJSONRows(t, zipMemberBytes(t, bundleBytes, "data/object_blobs.ndjson"))
		if len(rows) != 1 || rows[0]["upload_state"] != "available" ||
			rows[0]["byte_size"] != float64(len("extension_profile incident bundle blob\n")) ||
			rows[0]["observed_size"] != float64(len("extension_profile incident bundle blob\n")) {
			t.Fatalf("expected one complete available blob observation: %#v", rows)
		}
		cases := []struct {
			name   string
			mutate func(map[string]any)
		}{
			{"both sizes disagree with bytes", func(row map[string]any) { row["byte_size"], row["observed_size"] = 123, 123 }},
			{"missing observed digest with expected digest", func(row map[string]any) { delete(row, "observed_sha256_hex") }},
			{"null observed digest with expected digest", func(row map[string]any) { row["observed_sha256_hex"] = nil }},
			{"empty observed digest", func(row map[string]any) { row["observed_sha256_hex"] = "" }},
			{"noncanonical observed digest", func(row map[string]any) { row["observed_sha256_hex"] = strings.Repeat("A", 64) }},
			{"nonhex observed digest", func(row map[string]any) { row["observed_sha256_hex"] = strings.Repeat("g", 64) }},
			{"mistyped observed digest", func(row map[string]any) { row["observed_sha256_hex"] = true }},
			{"expected digest disagrees", func(row map[string]any) { row["expected_sha256_hex"] = strings.Repeat("0", 64) }},
			{"missing expected digest member", func(row map[string]any) { delete(row, "expected_sha256_hex") }},
			{"mistyped expected digest", func(row map[string]any) { row["expected_sha256_hex"] = false }},
		}
		for _, field := range []string{"byte_size", "observed_size"} {
			cases = append(cases, struct {
				name   string
				mutate func(map[string]any)
			}{"missing " + field, func(row map[string]any) { delete(row, field) }})
			for _, value := range []any{nil, "38", -1, json.Number("-0"), json.Number("38.0"), json.Number("38e0"), json.Number("9223372036854775808"), 123} {
				cases = append(cases, struct {
					name   string
					mutate func(map[string]any)
				}{field + "=" + fmt.Sprint(value), func(row map[string]any) { row[field] = value }})
			}
		}
		for index, test := range cases {
			t.Run(test.name, func(t *testing.T) {
				row := mapsClone(rows[0])
				test.mutate(row)
				candidate := replaceStructuredBundleMember(t, bundleBytes, "data/object_blobs.ndjson", encodeNDJSONRows(t, []map[string]any{row}))
				terminal := assertImportFailureLeavesState(t, targetHarness, targetAdmin, incidentID,
					fmt.Sprintf("blob-observation-%d", index), candidate, "source_family_invalid")
				details := terminal["error_summary"].(map[string]any)["details"].(map[string]any)
				if len(details) != 3 || details["reason_code"] != "source_family_invalid" ||
					details["source_family_id"] != "evidence" || details["invariant_id"] != "evidence.byte_size_digest_agree" {
					t.Fatalf("blob observation failure is not exact and redacted: %#v", details)
				}
			})
		}
	})

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
		// Isolate Records validation from earlier Revisions Apply reference checks.
		subtypeIncomplete := replaceStructuredBundleMember(t, bundleBytes, "data/timeline_records.ndjson", nil)
		for _, path := range []string{"data/change_sets.ndjson", "data/change_set_mutations.ndjson", "data/record_revisions.ndjson"} {
			subtypeIncomplete = replaceStructuredBundleMember(t, subtypeIncomplete, path, nil)
		}
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
				name:        "subtype complete",
				txn:         "txn-import-records-subtype-complete",
				bundle:      subtypeIncomplete,
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
		directoryAdmin, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, directoryHarness.Server.HTTP.URL)
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
		limitAdmin, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, limitHarness.Server.HTTP.URL)
		assertImportFailureLeavesState(t, limitHarness, limitAdmin, incidentID, "txn-import-extracted-limit", bundleBytes, "archive_extracted_bytes_exceeded")
	})
	t.Run("archive member count limit", func(t *testing.T) {
		limitHarness := startIsolatedIncidentBundleServerWithEnv(t, runtime, "extension_profile-incident-bundle-member-limit", map[string]string{
			"CARTULARY__LIMITS__ARCHIVES__MAX_MEMBERS": "20",
		})
		limitAdmin, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, limitHarness.Server.HTTP.URL)
		assertImportFailureLeavesState(t, limitHarness, limitAdmin, incidentID, "txn-import-member-limit", bundleBytes, "archive_member_count_exceeded")
	})

	importBundleAndWait(t, targetHarness.Server, targetAdmin, bundleBytes, "txn-import-duplicate-baseline")
	assertImportFailureLeavesState(t, targetHarness, targetAdmin, incidentID, "txn-import-duplicate-incident", bundleBytes, "duplicate_incident_id")

	t.Run("zero byte blob with no expected digest", func(t *testing.T) {
		const emptySHA = "e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855"
		files := zipMemberMap(t, bundleBytes)
		blobRows := decodeNDJSONRows(t, files["data/object_blobs.ndjson"])
		blobRows[0]["byte_size"], blobRows[0]["observed_size"] = 0, 0
		blobRows[0]["observed_sha256_hex"], blobRows[0]["expected_sha256_hex"] = emptySHA, nil
		evidenceRows := decodeNDJSONRows(t, files["data/evidence_records.ndjson"])
		for _, row := range evidenceRows {
			row["blob_hash"] = emptySHA
			replaceEvidenceCreationFixtureHash(t, files, row["record_id"].(string), emptySHA)
		}
		files["data/evidence_records.ndjson"] = encodeNDJSONRows(t, evidenceRows)
		delete(files, blobPath)
		files["blobs/sha256/"+emptySHA] = []byte{}
		candidate := replaceStructuredBundleMember(t, writeZipMemberMap(t, files), "data/object_blobs.ndjson", encodeNDJSONRows(t, blobRows))
		emptyTarget := startIsolatedIncidentBundleServer(t, runtime, "incident-bundle-empty-blob")
		emptyAdmin, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, emptyTarget.Server.HTTP.URL)
		importBundleAndWait(t, emptyTarget.Server, emptyAdmin, candidate, "empty-observed-blob")
		if countRows(t, emptyTarget.DB, `SELECT count(*) FROM object_blobs WHERE incident_id=$1 AND upload_state='available' AND byte_size=0 AND observed_size=0 AND expected_sha256_hex IS NULL AND observed_sha256_hex='e3b0c44298fc1c149afbf4c8996fb92427ae41e4649b934ca495991b7852b855'`, incidentID) != 1 {
			t.Fatal("empty observed blob did not preserve independent metadata")
		}
		reexported := exportBundleBytes(t, emptyTarget, emptyAdmin, incidentID, "empty-observed-reexport")
		if payload, exists := zipMemberMap(t, reexported)["blobs/sha256/"+emptySHA]; !exists || len(payload) != 0 {
			t.Fatal("empty observed blob bytes did not survive staging and re-export")
		}
		assertNoIncidentBundleStaging(t, emptyTarget.Server)
	})

	t.Run("quarantined bytes survive without state promotion", func(t *testing.T) {
		// Arrange retained source state; this is not a test of the quarantine command.
		seedPortableEvidenceQuarantine(t, sourceHarness, uuid.MustParse(incidentID), uuid.MustParse(seededState.EvidenceRecordID), sourceAdminID)
		quarantined := exportBundleBytes(t, sourceHarness, sourceAdmin, incidentID, "quarantined-blob-export")
		if payload, exists := zipMemberMap(t, quarantined)[blobPath]; !exists || string(payload) != "extension_profile incident bundle blob\n" {
			t.Fatal("quarantined source bytes were omitted or changed during export")
		}
		blobRows := assertPortableBlobShape(t, quarantined, 1)
		if len(blobRows) != 1 || blobRows[0]["upload_state"] != "quarantined" {
			t.Fatalf("quarantined source metadata changed: %#v", blobRows)
		}
		blobID := blobRows[0]["object_blob_id"].(string)
		var sourceKey string
		if err := sourceHarness.DB.QueryRow(`SELECT storage_key FROM object_blobs WHERE object_blob_id=$1`, blobID).Scan(&sourceKey); err != nil {
			t.Fatal(err)
		}
		quarantineTarget := startIsolatedIncidentBundleServer(t, runtime, "incident-bundle-quarantined-blob")
		quarantineAdmin, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, quarantineTarget.Server.HTTP.URL)
		assertImportFailureLeavesState(t, quarantineTarget, quarantineAdmin, incidentID, "quarantine-missing-blob",
			removeZipMember(t, quarantined, blobPath), "missing_required_blob")
		badRow := mapsClone(blobRows[0])
		badRow["observed_size"] = 123
		badMetadata := replaceStructuredBundleMember(t, quarantined, "data/object_blobs.ndjson", encodeNDJSONRows(t, []map[string]any{badRow}))
		failed := assertImportFailureLeavesState(t, quarantineTarget, quarantineAdmin, incidentID, "quarantine-bad-observation", badMetadata, "source_family_invalid")
		details := failed["error_summary"].(map[string]any)["details"].(map[string]any)
		if len(details) != 3 || details["source_family_id"] != "evidence" || details["invariant_id"] != "evidence.byte_size_digest_agree" {
			t.Fatalf("quarantined observation error lost its owner: %#v", details)
		}
		importBundleAndWait(t, quarantineTarget.Server, quarantineAdmin, quarantined, "quarantined-blob-import")
		if countRows(t, quarantineTarget.DB, `SELECT count(*) FROM evidence e JOIN object_blobs b USING(object_blob_id) WHERE e.incident_id=$1 AND e.lifecycle_state='quarantined' AND e.upload_state='quarantined' AND b.upload_state='quarantined'`, incidentID) != 1 {
			t.Fatal("quarantined Evidence or blob was promoted during import")
		}
		var targetKey string
		if err := quarantineTarget.DB.QueryRow(`SELECT storage_key FROM object_blobs WHERE object_blob_id=$1`, blobID).Scan(&targetKey); err != nil {
			t.Fatal(err)
		}
		parts, err := blobref.ParseObjectBlobStorageKey(targetKey)
		if err != nil || targetKey == sourceKey || parts.IncidentID.String() != incidentID || parts.ObjectBlobID.String() != blobID {
			t.Fatalf("quarantined blob did not bind a fresh destination identity: %v", err)
		}
		if countRows(t, quarantineTarget.DB, `SELECT count(*) FROM extension_staged_objects s JOIN extension_staged_object_references r USING(staging_id) WHERE s.storage_identity=$1 AND s.state='published' AND s.owner_profile_id='evidence' AND r.owner_resource_kind='object_blob' AND r.owner_resource_id=$2`, targetKey, incidentID+":"+blobID) != 1 {
			t.Fatal("quarantined blob publication absent or duplicated")
		}
		reader, _, err := quarantineTarget.ObjectStore.ReadObject(t.Context(), targetKey, objectstore.ReadOptions{})
		if err != nil {
			t.Fatal(err)
		}
		payload, readErr := io.ReadAll(reader)
		closeErr := reader.Close()
		if readErr != nil || closeErr != nil || string(payload) != "extension_profile incident bundle blob\n" {
			t.Fatalf("quarantined bytes did not survive staging: read=%v close=%v", readErr, closeErr)
		}
		reexported := exportBundleBytes(t, quarantineTarget, quarantineAdmin, incidentID, "quarantined-blob-reexport")
		assertPortableBlobShape(t, reexported, 1)
		if payload, exists := zipMemberMap(t, reexported)[blobPath]; !exists || string(payload) != "extension_profile incident bundle blob\n" {
			t.Fatal("quarantined bytes did not survive re-export")
		}
		assertNoIncidentBundleStaging(t, quarantineTarget.Server)
	})
}
