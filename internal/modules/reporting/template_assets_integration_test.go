package reporting_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"github.com/jackc/pgx/v5/pgconn"
	"net/http"
	"reflect"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/reporting"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/reportingfixture"
)

func declaredTemplateImageRegistry(t *testing.T) (reporting.TemplatePack, *reporting.TemplatePackRegistry) {
	t.Helper()
	pack := reportingfixture.TemplatePresentationPack()
	registry, err := reporting.NewTemplatePackRegistry(pack)
	if err != nil {
		t.Fatal(err)
	}
	return pack, registry
}

// This exercises ordinary authenticated admission and retained-state recovery
// with the declared application protocol fixture. Native browser rendering is
// qualified separately by TestReportingInstalledSandboxRenders.
func assertDeclaredTemplateImageAdmissions(t *testing.T, harness *appsupport.ServerHarness, actor flowtest.LoginResult, snapshotID, incidentID string, pack reporting.TemplatePack) {
	t.Helper()
	post := func(path string, body map[string]any, status int) map[string]any {
		return httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+path, body, httptestx.WithCookies(actor.SessionCookie, actor.CSRFCookie), httptestx.WithHeader(authn.CSRFHeaderName, actor.CSRFCookie.Value)), status)["data"].(map[string]any)
	}
	request := map[string]any{"snapshot_id": snapshotID, "client_txn_id": "local-assets-release", "template_id": pack.Manifest.TemplateID, "template_version": pack.Manifest.TemplateVersion, "output_kind": reporting.OutputKindSlidev, "release_scope": reporting.ReleaseScopeInternalReview}
	job := post("/api/v1/releases", request, http.StatusAccepted)
	releaseID := requireSucceededJobResourceID(t, harness, actor, job, "release")
	if replay := post("/api/v1/releases", request, http.StatusAccepted); replay["job_id"] != job["job_id"] {
		t.Fatal("asset release replay created another admission")
	}
	composition := post("/api/v1/incidents/"+incidentID+"/report-compositions", map[string]any{"client_txn_id": "local-assets-composition", "template_id": pack.Manifest.TemplateID, "template_version": pack.Manifest.TemplateVersion, "deck_ops": []any{}, "diagram_decls": []any{}, "authored_texts": []any{}}, http.StatusCreated)
	raw, err := json.Marshal(requireSnapshotExportModel(t, harness.DB, snapshotID))
	if err != nil {
		t.Fatal(err)
	}
	var snapshot reporting.SnapshotModel
	if err := json.Unmarshal(raw, &snapshot); err != nil {
		t.Fatal(err)
	}
	_, profileSHA, err := reporting.ResolveRedactionProfile(reporting.InternalRedactionProfileID, reporting.DefaultRedactionProfileVersion, nil, snapshot.SnapshotContent)
	if err != nil {
		t.Fatal(err)
	}
	previewRequest := map[string]any{"client_txn_id": "local-assets-preview", "source_kind": "draft", "snapshot_id": snapshotID, "derivation_version": reporting.DerivationVersion, "template_id": pack.Manifest.TemplateID, "template_version": pack.Manifest.TemplateVersion, "redaction_profile_id": reporting.InternalRedactionProfileID, "redaction_profile_version": reporting.DefaultRedactionProfileVersion, "redaction_profile_sha256": profileSHA, "render_environment_profile_id": "cartulary.reporting.local.v1", "output_kind": reporting.OutputKindSlidev, "output_options": map[string]any{}, "recipient_partition_refs": []any{}, "graph_projection_refs": []any{}}
	previewURL := "/api/v1/incidents/" + incidentID + "/report-compositions/" + composition["composition_id"].(string) + "/preview"
	preview := post(previewURL, previewRequest, http.StatusAccepted)
	previewJob := preview["render_attempt_id"].(string)
	requireJobStatus(t, harness, actor, previewJob, "succeeded")
	if replay := post(previewURL, previewRequest, http.StatusAccepted); !reflect.DeepEqual(replay, preview) {
		t.Fatal("asset preview replay changed the committed preview view")
	}
	var previewID string
	if err := harness.DB.QueryRow(`SELECT preview_attempt_id::text FROM reporting_composition_preview_outputs WHERE render_attempt_id=$1`, previewJob).Scan(&previewID); err != nil {
		t.Fatal(err)
	}
	references := reportingReadinessReferences(t, harness)
	for _, item := range []struct {
		id, job, query string
		ready          func() error
	}{
		{releaseID, job["job_id"].(string), `SELECT inline_bytes FROM reporting_render_bundle_files WHERE release_id=$1 AND bundle_path=$2`, func() error {
			return reporting.ValidateRetainedReleases(context.Background(), harness.Pool, references)
		}},
		{previewID, previewJob, `SELECT inline_bytes FROM reporting_composition_preview_output_files WHERE preview_attempt_id=$1 AND bundle_path=$2`, func() error {
			return reporting.ValidateRetainedPreviews(context.Background(), harness.Pool, references)
		}},
	} {
		var data, input []byte
		var admitted string
		for _, asset := range pack.Manifest.Assets {
			if err := harness.DB.QueryRow(item.query, item.id, asset.BundlePath).Scan(&data); err != nil {
				t.Fatal(err)
			}
			for _, file := range pack.Files {
				if file.SourcePath == asset.SourcePath && !bytes.Equal(data, file.Bytes) {
					t.Fatal("ordinary admission changed template presentation bytes")
				}
			}
		}
		if err := harness.DB.QueryRow(`SELECT request_json,admitted_template_manifest_sha256 FROM reporting_job_payloads WHERE job_id=$1`, item.job).Scan(&input, &admitted); err != nil {
			t.Fatal(err)
		}
		if len(admitted) != 64 || item.ready() != nil {
			t.Fatal("admitted asset history failed readiness")
		}
		for _, mutation := range []struct {
			query string
			value any
		}{
			{`UPDATE reporting_job_payloads SET admitted_template_manifest_sha256=$2 WHERE job_id=$1`, strings.Repeat("0", 64)},
			{`UPDATE reporting_job_payloads SET request_json=jsonb_set(request_json,'{template_contract,deck_title}',$2::jsonb) WHERE job_id=$1`, `"substituted"`},
			{`UPDATE reporting_job_payloads SET request_json=jsonb_set(request_json,'{template_files,0,bytes}',$2::jsonb) WHERE job_id=$1`, `"YWx0ZXJlZA=="`},
		} {
			if _, err := harness.DB.Exec(mutation.query, item.job, mutation.value); err != nil {
				var constraint *pgconn.PgError
				if errors.As(err, &constraint) && constraint.Code == "23514" {
					continue
				}
				t.Fatal(err)
			}
			rejected := item.ready() != nil
			if _, err := harness.DB.Exec(`UPDATE reporting_job_payloads SET request_json=$2,admitted_template_manifest_sha256=$3 WHERE job_id=$1`, item.job, input, admitted); err != nil {
				t.Fatal(err)
			}
			if !rejected {
				t.Fatal("recovery accepted replaced template admission")
			}
		}
		if err := item.ready(); err != nil {
			t.Fatal("original frozen asset bytes failed replay", err)
		}
	}
}
