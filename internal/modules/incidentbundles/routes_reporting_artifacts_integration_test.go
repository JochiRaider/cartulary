package incidentbundles_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/app/server"
	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/modules/reporting"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/reportingfixture"
)

func testReportingArtifactRoundTrip(t *testing.T, runtime *appsupport.Runtime) {
	pack := reportingfixture.TemplatePresentationPack()
	templates, err := reporting.NewTemplatePackRegistry(pack)
	if err != nil {
		t.Fatal(err)
	}
	source := startIsolatedIncidentBundleServerConfigured(t, runtime, "reporting-artifact-source", nil, func(options *server.Options) { options.ReportingTemplatePacks = templates })
	target := startIsolatedIncidentBundleServer(t, runtime, "reporting-artifact-target")
	admin, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, source.Server.HTTP.URL)
	destination, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, target.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, source.Server, admin, map[string]any{"client_txn_id": "artifact-incident", "incident_key": "ARTIFACT-ROUNDTRIP", "title": "Original immutable title", "current_phase": "analysis"})
	incidentID := incident["incident_id"].(string)
	partyResponse := httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, source.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/views/cartulary.view.parties.v1/rows", map[string]any{
		"client_txn_id": "artifact-party", "party.display_name": "Historical subject", "party.party_kind": "person", "party.organization_name": "Example", "party.primary_email": "source@example.test",
	}, httptestx.WithCookies(admin.Session, admin.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, admin.CSRF.Value))
	partyData := httptestx.RequireSuccessEnvelope(t, partyResponse, http.StatusCreated)["data"].(map[string]any)
	partyID := partyData["row"].(map[string]any)["record_id"].(string)
	var retainedSourceJSON []byte
	if err := source.DB.QueryRow(`SELECT jsonb_agg(rev.after_json->'source' ORDER BY rev.revision_id) FROM record_revisions rev JOIN records r ON r.record_id=rev.record_id WHERE r.incident_id=$1 AND r.record_type='party'`, incidentID).Scan(&retainedSourceJSON); err != nil {
		t.Fatal(err)
	}
	var retainedSources []map[string]any
	if err := json.Unmarshal(retainedSourceJSON, &retainedSources); err != nil {
		t.Fatal(err)
	}
	wantSources := []map[string]any{{
		"record_id": partyID, "incident_id": incidentID,
		"display_name": "Historical subject", "party_kind": "person",
		"organization_name": "Example", "primary_email": "source@example.test",
		"role_title": nil, "timezone_name": nil, "external_ref": nil, "notes": nil,
	}}
	if !reflect.DeepEqual(retainedSources, wantSources) {
		t.Fatalf("retained Party source facts = %#v, want %#v", retainedSources, wantSources)
	}
	jobResource := func(route string, body map[string]any) string {
		t.Helper()
		response := httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, source.Server.HTTP.URL+route, body, httptestx.WithCookies(admin.Session, admin.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, admin.CSRF.Value))
		job := httptestx.RequireSuccessEnvelope(t, response, http.StatusAccepted)["data"].(map[string]any)
		terminal := waitJob(t, source.Server, admin, job["job_id"].(string))
		return terminal["result_summary"].(map[string]any)["resource_refs"].([]any)[0].(map[string]any)["id"].(string)
	}
	snapshot := jobResource("/api/v1/snapshots", map[string]any{"incident_id": incidentID, "client_txn_id": "artifact-snapshot"})
	release := jobResource("/api/v1/releases", map[string]any{"snapshot_id": snapshot, "client_txn_id": "artifact-release", "template_id": pack.Manifest.TemplateID, "template_version": pack.Manifest.TemplateVersion, "redaction_profile_id": reporting.TokenizedRedactionProfileID, "redaction_profile_version": reporting.DefaultRedactionProfileVersion, "release_scope": reporting.ReleaseScopeInternalReview, "output_kind": reporting.OutputKindSlidev})
	scenariotest.PatchIncident(t, source.Server, admin, incidentID, map[string]any{"base_incident_version": 1, "current_phase": "containment"})
	export := func(h *appsupport.ServerHarness, actor authcookietest.AuthCookies, txn string, include bool) []byte {
		t.Helper()
		body := map[string]any{"incident_id": incidentID, "client_txn_id": txn}
		if include {
			body["optional_sections"] = []string{"snapshots"}
		}
		job := httptestx.RequireSuccessEnvelope(t, postExport(t, h.Server, actor, body), http.StatusAccepted)["data"].(map[string]any)
		terminal := waitJob(t, h.Server, actor, job["job_id"].(string))
		id := terminal["result_summary"].(map[string]any)["resource_refs"].([]any)[0].(map[string]any)["id"].(string)
		ref := stringScalar(t, h.DB, `SELECT bundle_storage_ref FROM incident_bundle_exports WHERE bundle_id=$1`, id)
		data, err := os.ReadFile(exportedBundleTestPath(t, h.Server, ref))
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	without := zipMemberMap(t, export(source, admin, "artifact-refs-only", false))
	for name := range without {
		if strings.HasPrefix(name, "ext/snapshots/") {
			t.Fatal("unselected artifacts were exported")
		}
	}
	original := export(source, admin, "artifact-export", true)
	files := zipMemberMap(t, original)
	imagePath := "ext/snapshots/releases/" + release + "/files/" + pack.Manifest.Assets[0].BundlePath
	if !bytes.Equal(files[imagePath], pack.Files[0].Bytes) {
		t.Fatal("export did not preserve frozen image bytes")
	}
	for _, asset := range pack.Manifest.Assets {
		for _, file := range pack.Files {
			if file.SourcePath == asset.SourcePath && !bytes.Equal(files["ext/snapshots/releases/"+release+"/files/"+asset.BundlePath], file.Bytes) {
				t.Fatal("export did not preserve frozen theme and font bytes")
			}
		}
	}
	const catalogPath = "ext/snapshots/catalog.json"
	capturePath := "ext/snapshots/" + snapshot + "/source-capture.json"
	manifestPath := "ext/snapshots/releases/" + release + "/manifest.json"
	objectDigest := func(schema, name string) string {
		t.Helper()
		data, exists := files[name]
		if !exists || len(data) == 0 {
			t.Fatalf("missing transported control object %s", name)
		}
		digest := sha256.Sum256(append([]byte(schema+"\n"), data...))
		return hex.EncodeToString(digest[:])
	}
	wantCatalog := map[string]any{
		"schema_id": "cartulary.reporting_portable_artifacts.v3", "incident_id": incidentID,
		"snapshots": []map[string]any{{
			"snapshot_id": snapshot, "source_capture_path": capturePath,
			"source_capture_sha256": objectDigest("cartulary.reporting_snapshot_capture.v1", capturePath),
		}},
		"releases": []map[string]any{{
			"release_id": release, "snapshot_id": snapshot, "manifest_path": manifestPath,
			"manifest_sha256": objectDigest("cartulary.render_bundle_manifest.v3", manifestPath),
		}},
	}
	wantCatalogBytes, err := json.Marshal(wantCatalog)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(files[catalogPath], wantCatalogBytes) {
		t.Fatalf("artifact catalog = %s, want %s", files[catalogPath], wantCatalogBytes)
	}
	for name, data := range files {
		if strings.HasPrefix(name, "ext/snapshots/releases/") && (strings.Contains(name, "internal/") || bytes.Contains(data, []byte("sensitive_reveal_map"))) {
			t.Fatal("sensitive adjunct entered portable release evidence")
		}
	}
	modelPath := capturePath
	var capture map[string]any
	if err := json.Unmarshal(files[modelPath], &capture); err != nil {
		t.Fatal(err)
	}
	if capture["schema_id"] != "cartulary.reporting_snapshot_capture.v1" || capture["snapshot_id"] != snapshot || capture["incident_id"] != incidentID {
		t.Fatalf("snapshot capture identity mismatch: %#v", capture)
	}
	wantCapturedIncident := map[string]any{
		"id": incidentID, "title": "Original immutable title", "description": nil,
		"status": "active", "severity": nil, "tlp": nil,
		"current_phase": "analysis", "version": float64(1),
	}
	if !reflect.DeepEqual(capture["incident"], wantCapturedIncident) {
		t.Fatalf("captured incident = %#v, want %#v", capture["incident"], wantCapturedIncident)
	}
	sensitivePath := "ext/snapshots/releases/" + release + "/files/" + "internal/reveal-map.json"
	if _, present := files[sensitivePath]; present {
		t.Fatal("reveal map leaked")
	}
	invalid := map[string][]byte{
		"altered-image":   replaceStructuredBundleMember(t, original, imagePath, []byte(`<svg xmlns="http://www.w3.org/2000/svg"><script>invalid</script></svg>`)),
		"altered-model":   replaceStructuredBundleMember(t, original, modelPath, append(append([]byte{}, files[modelPath]...), byte(' '))),
		"undeclared-file": replaceStructuredBundleMember(t, original, "ext/snapshots/undeclared.json", []byte(`{}`)),
		"sensitive-file":  replaceStructuredBundleMember(t, original, sensitivePath, []byte(`{"secret":"must-not-enter-import"}`)),
	}
	var altered map[string]any
	if err := json.Unmarshal(files[catalogPath], &altered); err != nil {
		t.Fatal(err)
	}
	altered["snapshots"].([]any)[0].(map[string]any)["source_capture_sha256"] = strings.Repeat("0", 64)
	badCatalog, err := canonicaljson.Marshal(altered)
	if err != nil {
		t.Fatal(err)
	}
	invalid["substituted-digest"] = replaceStructuredBundleMember(t, original, catalogPath, badCatalog)
	for name, bundle := range invalid {
		t.Run(name, func(t *testing.T) {
			response := postImport(t, target.Server, destination, `{"client_txn_id":"artifact-`+name+`"}`, bundle, "bundle.zip")
			job := httptestx.RequireSuccessEnvelope(t, response, http.StatusAccepted)["data"].(map[string]any)
			terminal := waitFailedJob(t, target.Server, destination, job["job_id"].(string))
			requireFailedJobReason(t, terminal, "incident_bundle_import_rejected", "malformed_manifest")
			if countRows(t, target.DB, `SELECT count(*) FROM incidents WHERE id=$1`, incidentID) != 0 || countRows(t, target.DB, `SELECT count(*) FROM reporting_imported_artifact_files`) != 0 || countRows(t, target.DB, `SELECT count(*) FROM reference_pack_portable_catalogs WHERE incident_id=$1`, incidentID) != 0 {
				t.Fatal("artifact failure published partial incident state")
			}
		})
	}
	active := stringScalar(t, target.DB, `SELECT pack_set_id FROM reference_pack_current_set WHERE singleton`)
	terminalImport := importBundleAndWait(t, target.Server, destination, original, "artifact-valid-import")
	if got := stringScalar(t, target.DB, `SELECT pack_set_id FROM reference_pack_current_set WHERE singleton`); got != active {
		t.Fatal("artifact import changed activation")
	}
	if countRows(t, target.DB, `SELECT count(*) FROM reporting_snapshots`)+countRows(t, target.DB, `SELECT count(*) FROM reporting_releases`)+countRows(t, target.DB, `SELECT count(*) FROM reporting_release_approvals`) != 0 {
		t.Fatal("source artifacts fabricated native reporting state")
	}
	exported := zipMemberMap(t, export(target, destination, "artifact-reexport", true))
	historicalMembers := func(members map[string][]byte) map[string][]byte {
		selected := map[string][]byte{}
		for name, data := range members {
			if strings.HasPrefix(name, "ext/snapshots/") || name == "data/reference_pack_refs.json" {
				selected[name] = data
			}
		}
		return selected
	}
	if !reflect.DeepEqual(historicalMembers(exported), historicalMembers(files)) {
		t.Fatal("historical re-export member inventory or bytes changed")
	}
	referenceCatalog := reportingArtifactCatalogFixture{files["data/reference_pack_refs.json"]}
	if err := reporting.ValidateRetainedArtifacts(t.Context(), target.Pool, referenceCatalog); err != nil {
		t.Fatal("valid historical artifacts rejected during recovery", err)
	}
	// A checksummed but undeclared retained file is corruption, even when it
	// could not be reached through normal import. Use only the disposable DB.
	extraPath := "ext/snapshots/corrupt-retained-file.json"
	if _, err := target.DB.Exec(`INSERT INTO reporting_imported_artifact_files(incident_id,bundle_path,operation_id,size_bytes,sha256) VALUES($1,$2,$3,2,$4)`, incidentID, extraPath, terminalImport["job_id"], hashHexBytes([]byte(`{}`))); err != nil {
		t.Fatal(err)
	}
	recoveryErr := reporting.ValidateRetainedArtifacts(t.Context(), target.Pool, referenceCatalog)
	if _, err := target.DB.Exec(`DELETE FROM reporting_imported_artifact_files WHERE incident_id=$1 AND bundle_path=$2`, incidentID, extraPath); err != nil {
		t.Fatal(err)
	}
	if recoveryErr == nil {
		t.Fatal("recovery admitted an undeclared historical artifact")
	}
	if err := reporting.ValidateRetainedArtifacts(t.Context(), target.Pool, referenceCatalog); err != nil {
		t.Fatal("repeated historical validation is not idempotent", err)
	}
	// Stored bytes remain immutable even to an application role with UPDATE
	// privileges accidentally added later. This failed statement is isolated.
	tx, err := target.DB.BeginTx(t.Context(), nil)
	if err != nil {
		t.Fatal(err)
	}
	_, err = tx.ExecContext(t.Context(), `UPDATE reporting_imported_artifact_files SET size_bytes=size_bytes WHERE incident_id=$1`, incidentID)
	if rollbackErr := tx.Rollback(); rollbackErr != nil {
		t.Fatal("rollback immutable artifact probe", rollbackErr)
	}
	var pgErr *pgconn.PgError
	if !errors.As(err, &pgErr) || pgErr.Code != "55000" {
		t.Fatalf("retained artifact update error = %v, want immutability rejection (55000)", err)
	}
}

type reportingArtifactCatalogFixture struct{ references []byte }

func (f reportingArtifactCatalogFixture) ExportTx(context.Context, pgx.Tx, uuid.UUID, []reference_data.SetBinding) ([]byte, error) {
	return f.references, nil
}
