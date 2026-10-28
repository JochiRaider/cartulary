package reporting_test

import (
	"encoding/json"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/app/server"
	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/modules/reporting"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/dbassert"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/jackc/pgx/v5/pgxpool"
)

func TestCompositionPreviewDelegatesToReportingAndRemainsInternalDraft_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	assetPack, templates := declaredTemplateImageRegistry(t)
	var manager *jobs.Manager
	var transactions *jobs.TransactionService
	var runner *jobs.Runner
	harness := runtime.StartServer(t, appsupport.ServerOptions{Prefix: "extension-profile-reporting-composition-preview", TestRouteMode: httptestx.TestRouteModeDisabled, ConfigureRuntime: func(options *server.Options) {
		options.ReportingTemplatePacks = templates
		options.ObserveJobs = func(m *jobs.Manager, txs *jobs.TransactionService, r *jobs.Runner, _ *pgxpool.Pool) {
			manager, transactions, runner = m, txs, r
		}
	}})
	adminLogin, adminUserID := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, harness.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, harness.Server, adminLogin, map[string]any{
		"client_txn_id": "txn-preview-incident",
		"incident_key":  "IR-PREVIEW-01",
		"title":         "Preview Incident",
		"severity":      "medium",
		"current_phase": "analysis",
	})
	incidentID := incident["incident_id"].(string)
	partyID := querySeedID(t, harness.DB, `
WITH rec AS (
    INSERT INTO records (incident_id, record_type, created_by_user_id, updated_by_user_id)
    VALUES ($1, 'party', $2, $2)
    RETURNING record_id
), party AS (
    INSERT INTO parties (record_id, incident_id, display_name, party_kind)
    SELECT record_id, $1, 'External recipient', 'organization' FROM rec
    RETURNING record_id
)
SELECT record_id::text FROM party
`, incidentID, adminUserID.String())

	snapshotJob := httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/snapshots",
		map[string]any{"incident_id": incidentID, "client_txn_id": "txn-preview-snapshot"},
		httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value),
	), http.StatusAccepted)["data"].(map[string]any)
	snapshotID := requireSucceededJobResourceID(t, harness, adminLogin, snapshotJob, "snapshot")

	composition := httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/report-compositions",
		map[string]any{
			"client_txn_id":    "txn-preview-composition",
			"template_id":      reporting.DefaultTemplateID,
			"template_version": reporting.DefaultTemplateVersion,
			"deck_ops":         []any{},
			"diagram_decls":    []any{},
			"authored_texts":   []any{map[string]any{"authored_text_id": "ordinary-text", "text_role": "authored_text", "body": "R&D <review> \u2028 separator \u2029", "disclosure_partition_ref": "internal_only"}},
		},
		httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value),
	), http.StatusCreated)["data"].(map[string]any)
	compositionID := composition["composition_id"].(string)
	profileID := reporting.InternalRedactionProfileID
	snapshotBytes, err := json.Marshal(requireSnapshotExportModel(t, harness.DB, snapshotID))
	if err != nil {
		t.Fatal(err)
	}
	var retained reporting.SnapshotModel
	if err := json.Unmarshal(snapshotBytes, &retained); err != nil {
		t.Fatal(err)
	}
	_, profileSHA, err := reporting.ResolveRedactionProfile(profileID, reporting.DefaultRedactionProfileVersion, nil, retained.SnapshotContent)
	if err != nil {
		t.Fatalf("resolve preview redaction profile: %v", err)
	}
	previewURL := harness.Server.HTTP.URL + "/api/v1/incidents/" + incidentID + "/report-compositions/" + compositionID + "/preview"
	previewRequest := func(clientTxnID string, sourceKind string, compositionVersion any) map[string]any {
		t.Helper()
		body := map[string]any{
			"client_txn_id":                 clientTxnID,
			"source_kind":                   sourceKind,
			"snapshot_id":                   snapshotID,
			"derivation_version":            reporting.DerivationVersion,
			"template_id":                   reporting.DefaultTemplateID,
			"template_version":              reporting.DefaultTemplateVersion,
			"redaction_profile_id":          profileID,
			"redaction_profile_version":     "5",
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
		return body
	}
	createPreview := func(clientTxnID string, sourceKind string, compositionVersion any) map[string]any {
		t.Helper()
		body := previewRequest(clientTxnID, sourceKind, compositionVersion)
		return httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(
			t, http.DefaultClient,
			http.MethodPost,
			previewURL,
			body,
			httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF),
			httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value),
		), http.StatusAccepted)["data"].(map[string]any)
	}

	for _, field := range []string{"snapshot_id", "redaction_profile_id", "redaction_profile_sha256", "render_environment_profile_id"} {
		t.Run("unsupported preview "+field, func(t *testing.T) {
			body := previewRequest("txn-preview-invalid-"+field, "draft", nil)
			body[field] = "unsupported"
			if field == "snapshot_id" {
				body[field] = "71000000-0000-4000-8000-000000000999"
			}
			if field == "redaction_profile_sha256" {
				body[field] = strings.Repeat("0", 64)
			}
			beforeJobs := dbassert.CountSQL(t, harness.DB, `SELECT count(*) FROM jobs`)
			beforeSources := dbassert.CountSQL(t, harness.DB, `SELECT count(*) FROM report_composition_preview_attempts`)
			response := httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, previewURL, body, httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value))
			if field == "redaction_profile_sha256" {
				admitted := httptestx.RequireSuccessEnvelope(t, response, http.StatusAccepted)["data"].(map[string]any)
				id := admitted["render_attempt_id"].(string)
				failed := requireJobStatus(t, harness, adminLogin, id, "failed")
				if failed["error_summary"].(map[string]any)["code"] != "release_render_failed" {
					t.Fatal("profile mismatch was not a render preparation failure")
				}
				if dbassert.CountSQL(t, harness.DB, `SELECT count(*) FROM reporting_job_payloads WHERE job_id=$1 AND render_preparation_json IS NOT NULL`, id) != 0 || dbassert.CountSQL(t, harness.DB, `SELECT count(*) FROM reporting_composition_preview_outputs WHERE render_attempt_id=$1`, id) != 0 {
					t.Fatal("failed preparation published derived state")
				}
				if dbassert.CountSQL(t, harness.DB, `SELECT count(*) FROM jobs`) != beforeJobs+1 || dbassert.CountSQL(t, harness.DB, `SELECT count(*) FROM report_composition_preview_attempts`) != beforeSources+1 {
					t.Fatal("preparation failure lost its admission")
				}
				return
			}
			denied := httptestx.RequireErrorEnvelope(t, response, http.StatusBadRequest, "invalid_request")
			if denied["error"].(map[string]any)["details"].(map[string]any)["validation_code"] != "composition_preview_input_invalid" {
				t.Fatal("unsafe preview denial", denied)
			}
			if dbassert.CountSQL(t, harness.DB, `SELECT count(*) FROM jobs`) != beforeJobs || dbassert.CountSQL(t, harness.DB, `SELECT count(*) FROM report_composition_preview_attempts`) != beforeSources {
				t.Fatal("rejected preview committed admission")
			}
		})
	}

	draftPreview := createPreview("txn-preview-draft", "draft", nil)
	draftJobID := draftPreview["render_attempt_id"].(string)
	var retainedOptions []byte
	if err := harness.DB.QueryRow(`SELECT output_options FROM report_composition_preview_attempts WHERE render_attempt_id::text = $1`, draftJobID).Scan(&retainedOptions); err != nil {
		t.Fatal("read admitted preview options", err)
	}
	var options map[string]any
	if err := json.Unmarshal(retainedOptions, &options); err != nil || len(options) != 7 || options["schema_id"] != reporting.OutputOptionsSchemaID || options["pdf"] != true || options["source_only"] != false {
		t.Fatal("preview persisted incomplete output defaults", err)
	}
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
	retainedBinding := func(jobID string) string {
		t.Helper()
		var binding string
		if err := harness.DB.QueryRow(`
SELECT jsonb_build_object('attempt', to_jsonb(a), 'bundle', o.bundle_manifest_json)::text
  FROM report_composition_preview_attempts a
  JOIN reporting_composition_preview_outputs o USING (preview_attempt_id)
 WHERE a.render_attempt_id::text = $1
`, jobID).Scan(&binding); err != nil {
			t.Fatal(err)
		}
		return binding
	}
	draftBinding := retainedBinding(draftJobID)
	replayedDraft := createPreview("txn-preview-draft", "draft", nil)
	if !reflect.DeepEqual(replayedDraft, draftPreview) {
		t.Fatal("preview replay changed the original committed preview view")
	}
	assertPreviewReceiptIntegrity(t, harness, draftJobID, "txn-preview-draft")
	if retainedBinding(draftJobID) != draftBinding {
		t.Fatal("preview replay changed admitted identity, time or snapshot binding")
	}

	version := httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(
		t, http.DefaultClient,
		http.MethodPost,
		harness.Server.HTTP.URL+"/api/v1/incidents/"+incidentID+"/report-compositions/"+compositionID+"/versions",
		map[string]any{"client_txn_id": "txn-preview-freeze", "base_draft_version": 1},
		httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value),
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
		if got := dbassert.CountSQL(t, harness.DB, `
SELECT count(*)
  FROM report_composition_preview_attempts a
  JOIN reporting_composition_preview_outputs o USING (preview_attempt_id)
  JOIN reporting_snapshots s ON s.snapshot_id::text = a.snapshot_id
 WHERE a.render_attempt_id::text = $1
   AND o.bundle_manifest_json->>'preview_attempt_id' = a.preview_attempt_id::text
   AND o.bundle_manifest_json->>'release_id' IS NULL
   AND (o.bundle_manifest_json->>'render_admitted_at')::timestamptz = a.created_at
   AND o.bundle_manifest_json->>'snapshot_id' = a.snapshot_id
   AND o.bundle_manifest_json->'reference_packs' = (SELECT p.request_json->'source_capture'->'reference_packs' FROM reporting_job_payloads p WHERE p.job_id=s.create_job_id)
`, jobID); got != 1 {
			t.Fatal("preview lost admitted render or pack-set binding", jobID)
		}
		if got := dbassert.CountSQL(t, harness.DB, `SELECT COUNT(*) FROM reporting_releases WHERE create_job_id::text = $1`, jobID); got != 0 {
			t.Fatal("preview created a release or approval", jobID)
		}
	}
	// A release consumes the same immutable version through the retained job
	// codec, while keeping its release identity separate from both previews.
	releaseJob := httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(
		t, http.DefaultClient, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/releases",
		map[string]any{
			"snapshot_id": snapshotID, "client_txn_id": "txn-preview-version-release",
			"template_id": reporting.DefaultTemplateID, "template_version": reporting.DefaultTemplateVersion,
			"redaction_profile_id": reporting.InternalRedactionProfileID, "redaction_profile_version": "5",
			"output_kind": reporting.OutputKindSlidev, "release_scope": reporting.ReleaseScopeInternalDraft,
			"composition_id": compositionID, "composition_version": version["composition_version"], "composition_sha256": version["composition_sha256"],
		},
		httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value),
	), http.StatusAccepted)["data"].(map[string]any)
	releaseID := requireSucceededJobResourceID(t, harness, adminLogin, releaseJob, "release")
	release := requireRelease(t, harness, adminLogin, releaseID)
	if release["composition_id"] != compositionID || release["composition_version"] != version["composition_version"] || release["composition_sha256"] != version["composition_sha256"] {
		t.Fatal("release lost its admitted immutable composition tuple")
	}
	_, bundle := requireReleaseBundle(t, harness.DB, releaseID)
	if bundle["release_id"] != releaseID || bundle["preview_attempt_id"] != nil {
		t.Fatal("composition release inherited preview identity")
	}
	// Internal preview/release success cannot authorize the same authored text
	// for external publication. The current external profile denies it.
	externalJob := httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(
		t, http.DefaultClient, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/releases",
		map[string]any{
			"snapshot_id": snapshotID, "client_txn_id": "txn-preview-external-text-denied",
			"template_id": reporting.DefaultTemplateID, "template_version": reporting.DefaultTemplateVersion,
			"redaction_profile_id": reporting.ExternalRedactionProfileID, "redaction_profile_version": "5",
			"output_kind": reporting.OutputKindSlidev, "release_scope": reporting.ReleaseScopeExternal,
			"recipient_partition_refs": []string{"party:" + partyID},
			"composition_id":           compositionID, "composition_version": version["composition_version"], "composition_sha256": version["composition_sha256"],
		},
		httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF),
		httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value),
	), http.StatusAccepted)["data"].(map[string]any)
	deniedID := requireFailedReleaseJob(t, harness, adminLogin, externalJob, "authored_text_not_permitted")
	denied := requireRelease(t, harness, adminLogin, deniedID)
	if denied["release_state"] != reporting.ReleaseStateRenderFailed || denied["output_sha256"] != nil || denied["redaction_manifest_sha256"] != nil || denied["output_media_type"] != nil {
		t.Fatal("denied composition published render artifacts", denied)
	}
	profileID = reporting.TokenizedRedactionProfileID
	_, profileSHA, err = reporting.ResolveRedactionProfile(profileID, reporting.DefaultRedactionProfileVersion, nil, retained.SnapshotContent)
	if err != nil {
		t.Fatal(err)
	}
	tokenized := createPreview("txn-preview-tokenized", "version", version["composition_version"])
	tokenizedJob := tokenized["render_attempt_id"].(string)
	requireJobStatus(t, harness, adminLogin, tokenizedJob, "succeeded")
	var tokenizedID string
	if err := harness.DB.QueryRow(`SELECT preview_attempt_id::text FROM reporting_composition_preview_outputs WHERE render_attempt_id::text=$1`, tokenizedJob).Scan(&tokenizedID); err != nil {
		t.Fatal(err)
	}
	assertPrivateRevealReadiness(t, harness, tokenizedID, true)
	assertNativePreviewReadiness(t, harness, tokenizedID)
	if got := dbassert.CountSQL(t, harness.DB, `SELECT count(*) FROM reporting_composition_preview_output_files WHERE preview_attempt_id::text=$1 AND (role='sensitive_reveal_map' OR bundle_path LIKE 'internal/%')`, tokenizedID); got != 0 {
		t.Fatal("preview leaked private adjunct")
	}

	tokenReleaseJob := httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(
		t, http.DefaultClient, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/releases",
		map[string]any{"snapshot_id": snapshotID, "client_txn_id": "txn-preview-tokenized-release", "template_id": reporting.DefaultTemplateID, "template_version": reporting.DefaultTemplateVersion, "redaction_profile_id": profileID, "redaction_profile_version": "5", "output_kind": reporting.OutputKindSlidev, "release_scope": reporting.ReleaseScopeInternalReview},
		httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value),
	), http.StatusAccepted)["data"].(map[string]any)
	tokenReleaseID := requireSucceededJobResourceID(t, harness, adminLogin, tokenReleaseJob, "release")
	assertConcurrentReportingAdmissions(t, harness, adminLogin, incidentID)
	assertDeclaredTemplateImageAdmissions(t, harness, adminLogin, snapshotID, incidentID, assetPack)
	assertReportingPrivateBackupRestore(t, runtime, harness, tokenReleaseID, tokenizedID)
	assertUnfinishedPreviewReadiness(t, harness, manager, transactions, runner, adminUserID.String(), func(txn string) map[string]any { return createPreview(txn, "draft", nil) })
	assertUnfinishedReleaseReadiness(t, harness, manager, transactions, adminUserID.String(), func(txn string) map[string]any {
		return httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(
			t, http.DefaultClient, http.MethodPost, harness.Server.HTTP.URL+"/api/v1/releases",
			map[string]any{"snapshot_id": snapshotID, "client_txn_id": txn, "template_id": reporting.DefaultTemplateID, "template_version": reporting.DefaultTemplateVersion, "redaction_profile_id": reporting.InternalRedactionProfileID, "redaction_profile_version": "5", "output_kind": reporting.OutputKindSlidev, "release_scope": reporting.ReleaseScopeInternalReview},
			httptestx.WithCookies(adminLogin.Session, adminLogin.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, adminLogin.CSRF.Value),
		), http.StatusAccepted)["data"].(map[string]any)
	})
}
