package reporting_test

import (
	"net/http"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/modules/reporting"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/dbassert"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestCompositionPreviewDelegatesToReportingAndRemainsInternalDraft_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	harness := runtime.StartDefaultServer(t, "extension-profile-reporting-composition-preview")
	adminLogin, _ := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-preview-incident",
		"incident_key":  "IR-PREVIEW-01",
		"title":         "Preview Incident",
		"severity":      "medium",
		"current_phase": "analysis",
	})
	incidentID := incident["incident_id"].(string)

	snapshotJob := httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/snapshots",
		map[string]any{"incident_id": incidentID, "client_txn_id": "txn-preview-snapshot"},
		httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	), http.StatusAccepted)["data"].(map[string]any)
	snapshotID := requireSucceededJobResourceID(t, harness, adminLogin, snapshotJob, "snapshot")

	composition := httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/report-compositions",
		map[string]any{
			"client_txn_id":    "txn-preview-composition",
			"template_id":      reporting.DefaultTemplateID,
			"template_version": reporting.DefaultTemplateVersion,
			"deck_ops":         []any{},
			"diagram_decls":    []any{},
			"authored_texts":   []any{},
		},
		httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	), http.StatusCreated)["data"].(map[string]any)
	compositionID := composition["composition_id"].(string)
	_, profileSHA, err := reporting.ResolveRedactionProfile(reporting.InternalRedactionProfileID, "1", nil)
	if err != nil {
		t.Fatalf("resolve preview redaction profile: %v", err)
	}
	previewURL := harness.Server.HTTP.URL + "/api/v1/incidents/" + incidentID + "/report-compositions/" + compositionID + "/preview"
	createPreview := func(clientTxnID string, sourceKind string, compositionVersion any) map[string]any {
		t.Helper()
		body := map[string]any{
			"client_txn_id":                 clientTxnID,
			"source_kind":                   sourceKind,
			"snapshot_id":                   snapshotID,
			"derivation_version":            reporting.DerivationVersion,
			"template_id":                   reporting.DefaultTemplateID,
			"template_version":              reporting.DefaultTemplateVersion,
			"redaction_profile_id":          reporting.InternalRedactionProfileID,
			"redaction_profile_version":     "1",
			"redaction_profile_sha256":      profileSHA,
			"render_environment_profile_id": "cartulary.reporting.local.v1",
			"output_kind":                   reporting.OutputKindSlidev,
			"output_options":                map[string]any{},
			"recipient_partition_refs":      []any{},
			"graph_projection_refs":         []any{},
		}
		if compositionVersion != nil {
			body["composition_version"] = compositionVersion
		}
		return httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(
			t,
			http.MethodPost,
			previewURL,
			body,
			httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
			httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
		), http.StatusAccepted)["data"].(map[string]any)
	}

	draftPreview := createPreview("txn-preview-draft", "draft", nil)
	draftJobID := draftPreview["render_attempt_id"].(string)
	draftJob := requireJobStatus(t, harness, adminLogin, draftJobID, "succeeded")
	if draftJob["result_summary"].(map[string]any)["code"] != "composition_preview_rendered" {
		t.Fatalf("draft preview terminal result = %#v", draftJob["result_summary"])
	}
	if refs, present := draftJob["result_summary"].(map[string]any)["resource_refs"]; present && refs != nil {
		if values, ok := refs.([]any); !ok || len(values) != 0 {
			t.Fatalf("preview must not publish a resource identity: %#v", refs)
		}
	}
	if got := dbassert.CountSQL(t, harness.DB, `
SELECT COUNT(*)
  FROM reporting_composition_preview_outputs
 WHERE render_attempt_id::text = $1
   AND release_scope = 'internal_draft'
`, draftJobID); got != 1 {
		t.Fatalf("draft preview internal output rows = %d want 1", got)
	}
	if got := dbassert.CountSQL(t, harness.DB, `SELECT COUNT(*) FROM reporting_releases WHERE create_job_id::text = $1`, draftJobID); got != 0 {
		t.Fatalf("preview must not create a release, got %d rows", got)
	}
	replayedDraft := createPreview("txn-preview-draft", "draft", nil)
	replayedJobID := replayedDraft["render_attempt_id"]
	if replayedJobID == nil {
		replayedJobID = replayedDraft["job_id"]
	}
	if replayedJobID != draftJobID {
		t.Fatalf("preview replay changed job identity: first=%#v replay=%#v", draftPreview, replayedDraft)
	}

	version := httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(
		t,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/report-compositions/"+compositionID+"/versions",
		map[string]any{"client_txn_id": "txn-preview-freeze", "base_draft_version": 1},
		httptestx.WithCookies(adminLogin.SessionCookie, adminLogin.CSRFCookie),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRFCookie.Value),
	), http.StatusCreated)["data"].(map[string]any)
	versionPreview := createPreview("txn-preview-version", "version", version["composition_version"])
	versionJobID := versionPreview["render_attempt_id"].(string)
	requireJobStatus(t, harness, adminLogin, versionJobID, "succeeded")
	if got := dbassert.CountSQL(t, harness.DB, `
SELECT COUNT(*)
  FROM reporting_composition_preview_outputs
 WHERE render_attempt_id::text IN ($1, $2)
   AND release_scope = 'internal_draft'
`, draftJobID, versionJobID); got != 2 {
		t.Fatalf("draft and immutable-version preview outputs = %d want 2", got)
	}
	for _, jobID := range []string{draftJobID, versionJobID} {
		if got := dbassert.CountSQL(t, harness.DB, `SELECT COUNT(*) FROM extension_job_commit_proofs WHERE job_id::text = $1`, jobID); got != 1 {
			t.Fatalf("preview job %s proof rows = %d want 1", jobID, got)
		}
	}
}
