package incidentbundles_test

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/modules/reporting"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func testReportingArtifactRoundTrip(t *testing.T, runtime *appsupport.Runtime) {
	source := startIsolatedIncidentBundleServer(t, runtime, "reporting-artifact-source")
	target := startIsolatedIncidentBundleServer(t, runtime, "reporting-artifact-target")
	admin, actorID := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, source.Server.HTTP.URL)
	destination, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, target.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, source.Server, admin, map[string]any{"client_txn_id": "artifact-incident", "incident_key": "ARTIFACT-ROUNDTRIP", "title": "Original immutable title", "current_phase": "analysis"})
	incidentID := incident["incident_id"].(string)
	party := uuid.New()
	if _, err := source.DB.Exec(`INSERT INTO records(record_id,incident_id,record_type,created_by_user_id,updated_by_user_id) VALUES($1,$2,'party',$3,$3)`, party, incidentID, actorID); err != nil {
		t.Fatal(err)
	}
	if _, err := source.DB.Exec(`INSERT INTO parties(record_id,incident_id,display_name,party_kind,organization_name,primary_email) VALUES($1,$2,'Historical subject','person','Example','source@example.test')`, party, incidentID); err != nil {
		t.Fatal(err)
	}
	jobResource := func(route string, body map[string]any) string {
		t.Helper()
		response := httptestx.DoJSON(t, http.MethodPost, source.Server.HTTP.URL+route, body, httptestx.WithCookies(admin.SessionCookie, admin.CSRFCookie), httptestx.WithHeader(authn.CSRFHeaderName, admin.CSRFCookie.Value))
		job := httptestx.RequireSuccessEnvelope(t, response, http.StatusAccepted)["data"].(map[string]any)
		terminal := waitJob(t, source.Server, admin, job["job_id"].(string))
		return terminal["result_summary"].(map[string]any)["resource_refs"].([]any)[0].(map[string]any)["id"].(string)
	}
	snapshot := jobResource("/api/v1/snapshots", map[string]any{"incident_id": incidentID, "client_txn_id": "artifact-snapshot"})
	release := jobResource("/api/v1/releases", map[string]any{"snapshot_id": snapshot, "client_txn_id": "artifact-release", "template_id": reporting.DefaultTemplateID, "template_version": reporting.DefaultTemplateVersion, "redaction_profile_id": reporting.TokenizedRedactionProfileID, "redaction_profile_version": "1", "release_scope": reporting.ReleaseScopeInternalReview, "output_kind": reporting.OutputKindSlidev})
	scenariotest.PatchIncident(t, source.Server, admin, incidentID, map[string]any{"base_incident_version": 1, "current_phase": "containment"})
	export := func(h *appsupport.ServerHarness, actor flowtest.LoginResult, txn string, include bool) []byte {
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
	const catalogPath = "ext/snapshots/catalog.json"
	var catalog struct {
		Snapshots []struct {
			ID   string `json:"snapshot_id"`
			Path string `json:"export_model_path"`
		} `json:"snapshots"`
		Releases []struct {
			ID string `json:"release_id"`
		} `json:"releases"`
		Omitted []struct {
			Release string `json:"release_id"`
			Path    string `json:"path"`
			Reason  string `json:"reason"`
		} `json:"omitted_files"`
	}
	if err := json.Unmarshal(files[catalogPath], &catalog); err != nil {
		t.Fatal(err)
	}
	if len(catalog.Snapshots) != 1 || catalog.Snapshots[0].ID != snapshot || len(catalog.Releases) != 1 || catalog.Releases[0].ID != release || len(catalog.Omitted) != 1 || catalog.Omitted[0].Reason != "sensitive_reveal_map" {
		t.Fatal("incomplete artifact catalog", catalog)
	}
	modelPath := catalog.Snapshots[0].Path
	if !bytes.Contains(files[modelPath], []byte("analysis")) || bytes.Contains(files[modelPath], []byte("containment")) {
		t.Fatal("snapshot was rematerialized from live state")
	}
	sensitivePath := "ext/snapshots/releases/" + release + "/files/" + catalog.Omitted[0].Path
	if _, present := files[sensitivePath]; present {
		t.Fatal("reveal map leaked")
	}
	invalid := map[string][]byte{
		"altered-model":   replaceStructuredBundleMember(t, original, modelPath, append(append([]byte{}, files[modelPath]...), byte(' '))),
		"undeclared-file": replaceStructuredBundleMember(t, original, "ext/snapshots/undeclared.json", []byte(`{}`)),
		"sensitive-file":  replaceStructuredBundleMember(t, original, sensitivePath, []byte(`{"secret":"must-not-enter-import"}`)),
	}
	var altered map[string]any
	if err := json.Unmarshal(files[catalogPath], &altered); err != nil {
		t.Fatal(err)
	}
	altered["snapshots"].([]any)[0].(map[string]any)["export_model_sha256"] = strings.Repeat("0", 64)
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
	for name, data := range files {
		if strings.HasPrefix(name, "ext/snapshots/") || name == "data/reference_pack_refs.json" {
			if !bytes.Equal(data, exported[name]) {
				t.Fatal("historical source bytes changed", name)
			}
		}
	}
	referenceCatalog := reportingArtifactCatalogFixture{files["data/reference_pack_refs.json"]}
	if err := reporting.ValidateRetainedArtifacts(t.Context(), target.Pool, referenceCatalog); err != nil {
		t.Fatal("valid historical artifacts rejected during recovery", err)
	}
	// A checksummed but undeclared retained file is corruption, even when it
	// could not be reached through normal import. Use only the disposable DB.
	extraPath := "ext/snapshots/corrupt-retained-file.json"
	if _, err := target.DB.Exec(`INSERT INTO reporting_imported_artifact_files(incident_id,bundle_path,operation_id,size_bytes,sha256,content) VALUES($1,$2,$3,2,$4,$5)`, incidentID, extraPath, terminalImport["job_id"], hashHexBytes([]byte(`{}`)), []byte(`{}`)); err != nil {
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
	tx, err := target.Pool.Begin(t.Context())
	if err != nil {
		t.Fatal(err)
	}
	_, err = tx.Exec(t.Context(), `UPDATE reporting_imported_artifact_files SET content=content WHERE incident_id=$1`, incidentID)
	_ = tx.Rollback(t.Context())
	if err == nil {
		t.Fatal("retained artifact update was allowed")
	}
}

type reportingArtifactCatalogFixture struct{ references []byte }

func (f reportingArtifactCatalogFixture) ExportTx(context.Context, pgx.Tx, uuid.UUID, []reference_data.SetBinding) ([]byte, error) {
	return f.references, nil
}
