package reporting_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"github.com/JochiRaider/cartulary/internal/app/extensionassembly"
	"github.com/JochiRaider/cartulary/internal/app/referenceassembly"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	reportingjson "github.com/JochiRaider/cartulary/internal/modules/reporting/canonicaljson"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"os"
	"path/filepath"
	"reflect"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	referencetest "github.com/JochiRaider/cartulary/internal/modules/reference_data/testsupport"
	"github.com/JochiRaider/cartulary/internal/modules/reporting"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestSnapshotAdmissionAndRerenderRetainSetAcrossActivation_Integration(t *testing.T) {
	expected := referencetest.OwnerFixture(t, "snapshot_binding")
	runtime := appsupport.StartRuntime(t)
	input, err := os.ReadFile("../../../contracts/reference-pack-fixtures/fixtures/portable-input.v1.json")
	if err != nil {
		t.Fatal(err)
	}
	refs, err := os.ReadFile("../../../contracts/reference-pack-fixtures/fixtures/portable-references.v1.json")
	if err != nil {
		t.Fatal(err)
	}
	fixture := referencetest.PortableFixture(t, time.Now().UTC(), input, refs)
	bootstrap := filepath.Join(t.TempDir(), "bootstrap.json")
	if err := os.WriteFile(bootstrap, fixture.Bootstrap, 0600); err != nil {
		t.Fatal(err)
	}
	harness := runtime.StartServer(t, appsupport.ServerOptions{Prefix: "snapshot-reference-binding", TestRouteMode: httptestx.TestRouteModeDisabled, Env: map[string]string{"CARTULARY__REFERENCE_PACKS__TRUST_BOOTSTRAP_PATH": bootstrap, "CARTULARY__REFERENCE_PACKS__CLOCK_TRUSTED": "true"}})
	admin, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, harness.Server.HTTP.URL)
	post := func(path string, body map[string]any) *http.Response {
		return httptestx.DoJSON(t, http.DefaultClient, http.MethodPost, harness.Server.HTTP.URL+path, body, httptestx.WithCookies(admin.Session, admin.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, admin.CSRF.Value))
	}
	importJob := httptestx.RequireSuccessEnvelope(t, uploadReportingReferencePack(t, harness, admin, fixture.Container), http.StatusAccepted)["data"].(map[string]any)
	requireJobStatus(t, harness, admin, importJob["job_id"].(string), "succeeded")
	incident := scenariotest.CreateIncident(t, harness.Server, admin, map[string]any{"client_txn_id": "binding-incident", "incident_key": "RP-SNAPSHOT-BINDING", "title": "Exact reference binding"})
	// Block publication, not admission. This makes activation occur after the
	// durable snapshot Job pin but before the snapshot worker can finish.
	barrier, err := harness.DB.BeginTx(context.Background(), nil)
	if err != nil {
		t.Fatal(err)
	}
	defer barrier.Rollback()
	if _, err := barrier.Exec(`LOCK TABLE reporting_snapshots IN SHARE MODE`); err != nil {
		t.Fatal(err)
	}
	snapshotJob := httptestx.RequireSuccessEnvelope(t, post("/api/v1/snapshots", map[string]any{"client_txn_id": "binding-snapshot", "incident_id": incident["incident_id"]}), http.StatusAccepted)["data"].(map[string]any)
	var admitted, digest string
	if err := harness.DB.QueryRow(`SELECT p.pack_set_id,s.pack_set_sha256 FROM reference_pack_pins p JOIN reference_pack_sets s USING(pack_set_id) WHERE p.owner_kind='reporting_snapshot_job' AND p.owner_id=$1`, snapshotJob["job_id"]).Scan(&admitted, &digest); err != nil {
		t.Fatal("missing admission pin", err)
	}
	if digest != *expected.Set || admitted != "rpset_"+digest {
		t.Fatal("independent Base identity changed", admitted)
	}
	httptestx.RequireSuccessEnvelope(t, post("/api/v1/reference-packs/"+fixture.Key+"/"+fixture.Version+"/activate", map[string]any{"client_txn_id": "binding-activate"}), http.StatusOK)
	var current string
	if err := harness.DB.QueryRow(`SELECT pack_set_id FROM reference_pack_current_set WHERE singleton`).Scan(&current); err != nil || current == admitted {
		t.Fatal("fixture did not change the current selection", current, err)
	}
	if err := barrier.Commit(); err != nil {
		t.Fatal(err)
	}
	snapshot := requireSucceededJobResourceID(t, harness, admin, snapshotJob, "snapshot")
	model := requireSnapshotExportModel(t, harness.DB, snapshot)
	binding, ok := model["reference_packs"].(map[string]any)
	if !ok || binding["pack_set_id"] != admitted || binding["pack_set_sha256"] != digest {
		t.Fatal("snapshot worker rebased admission", binding)
	}
	release := func(transaction string) *http.Response {
		return post("/api/v1/releases", map[string]any{"snapshot_id": snapshot, "client_txn_id": transaction, "template_id": reporting.DefaultTemplateID, "template_version": reporting.DefaultTemplateVersion, "redaction_profile_id": reporting.InternalRedactionProfileID, "redaction_profile_version": "5", "release_scope": reporting.ReleaseScopeInternalDraft, "output_kind": reporting.OutputKindSlidev})
	}
	for _, transaction := range []string{"binding-render", "binding-rerender"} {
		job := httptestx.RequireSuccessEnvelope(t, release(transaction), http.StatusAccepted)["data"].(map[string]any)
		id := requireSucceededJobResourceID(t, harness, admin, job, "release")
		_, manifest := requireReleaseBundle(t, harness.DB, id)
		if !reflect.DeepEqual(manifest["reference_packs"], binding) {
			t.Fatal("render or rerender changed historical provenance")
		}
	}
	retention := reportingReadinessReferences(t, harness)
	requireRetainedSnapshotReadiness(t, harness, snapshot, retention)
	// Lose only the pinned Base host manifest. The active imported replacement
	// remains intact and cannot be silently substituted into the old snapshot.
	if _, err := harness.DB.Exec(`UPDATE reference_pack_objects SET available=false WHERE sha256 IN (SELECT v.manifest_sha256 FROM reference_pack_versions v JOIN reference_pack_candidates c USING(pack_key,pack_version) WHERE c.pack_key='type_registry.host' AND c.distribution_kind='packaged_builtin')`); err != nil {
		t.Fatal(err)
	}
	httptestx.RequireErrorEnvelope(t, release("binding-loss"), http.StatusConflict, expected.Error.Code)
	if err := reporting.ValidateRetainedSnapshotModels(context.Background(), harness.Pool, retention); err == nil {
		t.Fatal("readiness replaced a missing pinned binding with the active set")
	}
}

func requireRetainedSnapshotReadiness(t *testing.T, harness *appsupport.ServerHarness, snapshot string, references reference_data.RetainedBindingValidator) {
	t.Helper()
	ctx := context.Background()
	if err := reporting.ValidateRetainedSnapshotModels(ctx, harness.Pool, references); err != nil {
		t.Fatal("valid historical snapshot rejected after current-set activation", err)
	}
	var original []byte
	var originalDigest string
	if err := harness.DB.QueryRow(`SELECT p.request_json->'source_capture',s.source_capture_sha256 FROM reporting_snapshots s JOIN reporting_job_payloads p ON p.job_id=s.create_job_id WHERE s.snapshot_id=$1`, snapshot).Scan(&original, &originalDigest); err != nil {
		t.Fatal(err)
	}
	restore := func() {
		t.Helper()
		if _, err := harness.DB.Exec(`WITH changed AS (UPDATE reporting_job_payloads SET request_json=jsonb_set(request_json,'{source_capture}',$2::jsonb) WHERE job_id=(SELECT create_job_id FROM reporting_snapshots WHERE snapshot_id=$1)) UPDATE reporting_snapshots SET source_capture_sha256=$3 WHERE snapshot_id=$1`, snapshot, original, originalDigest); err != nil {
			t.Fatal(err)
		}
	}
	defer restore()
	for _, mutation := range []string{"changed frozen content", "missing source fields", "retired capture"} {
		var candidate map[string]any
		if err := json.Unmarshal(original, &candidate); err != nil {
			t.Fatal(err)
		}
		switch mutation {
		case "changed frozen content":
			candidate["incident"].(map[string]any)["title"] = "Substituted after admission"
		case "missing source fields":
			delete(candidate, "fields")
		case "retired capture":
			candidate["schema_id"] = "cartulary.reporting_snapshot_model.v6"
		}
		raw, err := reportingjson.Marshal(candidate)
		if err != nil {
			t.Fatal(err)
		}
		sum := sha256.Sum256(append([]byte("cartulary.reporting_snapshot_capture.v1\n"), raw...))
		digest := hex.EncodeToString(sum[:])
		if _, err := harness.DB.Exec(`WITH changed AS (UPDATE reporting_job_payloads SET request_json=jsonb_set(request_json,'{source_capture}',$2::jsonb) WHERE job_id=(SELECT create_job_id FROM reporting_snapshots WHERE snapshot_id=$1)) UPDATE reporting_snapshots SET source_capture_sha256=$3 WHERE snapshot_id=$1`, snapshot, raw, digest); err != nil {
			t.Fatal(err)
		}
		if err := reporting.ValidateRetainedSnapshotModels(ctx, harness.Pool, references); err == nil {
			t.Fatal("hostile retained snapshot passed readiness", mutation)
		}
		var retained string
		if err := harness.DB.QueryRow(`SELECT source_capture_sha256 FROM reporting_snapshots WHERE snapshot_id=$1`, snapshot).Scan(&retained); err != nil || retained != digest {
			t.Fatal("readiness repaired retained evidence", err)
		}
		restore()
	}
	if err := reporting.ValidateRetainedSnapshotModels(ctx, harness.Pool, references); err != nil {
		t.Fatal("restored original snapshot rejected", err)
	}
	var jobID string
	var admitted []byte
	if err := harness.DB.QueryRow(`SELECT s.create_job_id,p.request_json FROM reporting_snapshots s JOIN reporting_job_payloads p ON p.job_id=s.create_job_id WHERE s.snapshot_id=$1`, snapshot).Scan(&jobID, &admitted); err != nil {
		t.Fatal(err)
	}
	restoreAdmission := func() {
		t.Helper()
		if _, err := harness.DB.Exec(`UPDATE reporting_job_payloads SET request_json=$2 WHERE job_id=$1`, jobID, admitted); err != nil {
			t.Fatal(err)
		}
	}
	defer restoreAdmission()
	for _, mutation := range []string{"unknown top-level member", "missing model member", "changed source boundary", "changed transaction"} {
		var candidate map[string]any
		if err := json.Unmarshal(admitted, &candidate); err != nil {
			t.Fatal(err)
		}
		switch mutation {
		case "unknown top-level member":
			candidate["future"] = true
		case "missing model member":
			delete(candidate["source_capture"].(map[string]any), "fields")
		case "changed source boundary":
			candidate["source_boundary_json"].(map[string]any)["incident_version"] = 999
		case "changed transaction":
			candidate["client_txn_id"] = "substituted-admission"
		}
		raw, err := json.Marshal(candidate)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := harness.DB.Exec(`UPDATE reporting_job_payloads SET request_json=$2 WHERE job_id=$1`, jobID, raw); err != nil {
			t.Fatal(err)
		}
		if err := reporting.ValidateRetainedSnapshotModels(ctx, harness.Pool, references); err == nil {
			t.Fatal("altered original admission passed readiness", mutation)
		}
		restoreAdmission()
	}
	// Model the Jobs owner's compacted successful tombstone. Its public response
	// and idempotency fields are erased; Reporting's required private history is
	// retained. No expiry runner or clock policy is being tested by this fixture.
	if _, err := harness.DB.Exec(`UPDATE jobs SET retained_until=finished_at,expired_at=now(),handler_payload_json=NULL,handler_attempt_id=NULL,handler_lease_expires_at=NULL,handler_failure_count=0,handler_next_attempt_at=NULL,handler_last_attempted_at=NULL,handler_last_error=NULL,result_summary_json=NULL,error_summary_json=NULL,message=NULL,extension_idempotency_identity=NULL,extension_idempotency_route_key=NULL,extension_idempotency_scope_key=NULL,extension_normalized_request_sha256=NULL WHERE job_id=$1`, jobID); err != nil {
		t.Fatal(err)
	}
	if err := reporting.ValidateRetainedSnapshotModels(ctx, harness.Pool, references); err != nil {
		t.Fatal("compacted public Job response invalidated retained snapshot", err)
	}
}

func uploadReportingReferencePack(t testing.TB, harness *appsupport.ServerHarness, admin authcookietest.AuthCookies, container []byte) *http.Response {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	for _, part := range []struct {
		disposition, media string
		data               []byte
	}{
		{`form-data; name="metadata"`, "application/json", []byte(`{"client_txn_id":"binding-import"}`)},
		{`form-data; name="file"; filename="reference.zip"`, "application/zip", container},
	} {
		header := textproto.MIMEHeader{}
		header.Set("Content-Disposition", part.disposition)
		header.Set("Content-Type", part.media)
		out, err := writer.CreatePart(header)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := io.Copy(out, bytes.NewReader(part.data)); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	request, err := http.NewRequest(http.MethodPost, harness.Server.HTTP.URL+"/api/v1/reference-packs/import", &body)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.Header.Set(authn.CSRFHeaderName, admin.CSRF.Value)
	request.AddCookie(admin.Session)
	request.AddCookie(admin.CSRF)
	return httptestx.Do(t, http.DefaultClient, request)
}

func reportingReadinessReferences(t *testing.T, harness *appsupport.ServerHarness) reference_data.RetainedBindingValidator {
	t.Helper()
	storage, err := referenceassembly.NewRootStorage(harness.Server.Config.Roots.TemporaryWork.Path, harness.Server.Config.Roots.ReferencePackStorage.Path)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(storage.Close)
	finalizer, err := extensionassembly.NewReferencePackMutationFinalizer(func(err error) { t.Errorf("unexpected integrity mutation: %v", err) })
	if err != nil {
		t.Fatal(err)
	}
	retention, err := reference_data.NewRetention(harness.Pool, storage, time.Now, reference_data.IntegrityOptions{Finalizer: finalizer, Limits: reference_data.DefaultLimits()})
	if err != nil {
		t.Fatal(err)
	}
	return retention
}
