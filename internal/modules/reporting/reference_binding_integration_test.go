package reporting_test

import (
	"bytes"
	"context"
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
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestSnapshotAdmissionAndRerenderRetainSetAcrossActivation_Integration(t *testing.T) {
	expected := referencetest.OwnerFixture(t, "snapshot_binding")
	runtime := appsupport.StartRuntime(t)
	input, err := os.ReadFile("../../../contracts/reference-packs/fixtures/portable-input.v1.json")
	if err != nil {
		t.Fatal(err)
	}
	refs, err := os.ReadFile("../../../contracts/reference-packs/fixtures/portable-references.v1.json")
	if err != nil {
		t.Fatal(err)
	}
	fixture := referencetest.PortableFixture(t, time.Now().UTC(), input, refs)
	bootstrap := filepath.Join(t.TempDir(), "bootstrap.json")
	if err := os.WriteFile(bootstrap, fixture.Bootstrap, 0600); err != nil {
		t.Fatal(err)
	}
	harness := runtime.StartServer(t, appsupport.ServerOptions{Prefix: "snapshot-reference-binding", TestRouteMode: httptestx.TestRouteModeDisabled, Env: map[string]string{"CARTULARY__REFERENCE_PACKS__TRUST_BOOTSTRAP_PATH": bootstrap, "CARTULARY__REFERENCE_PACKS__CLOCK_TRUSTED": "true"}})
	admin, _ := flowtest.ProvisionBootstrapAdmin(t, harness.Server.HTTP.URL)
	post := func(path string, body map[string]any) *http.Response {
		return httptestx.DoJSON(t, http.MethodPost, harness.Server.HTTP.URL+path, body, httptestx.WithCookies(admin.SessionCookie, admin.CSRFCookie), httptestx.WithHeader(authn.CSRFHeaderName, admin.CSRFCookie.Value))
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
		return post("/api/v1/releases", map[string]any{"snapshot_id": snapshot, "client_txn_id": transaction, "template_id": reporting.DefaultTemplateID, "template_version": reporting.DefaultTemplateVersion, "redaction_profile_id": reporting.InternalRedactionProfileID, "redaction_profile_version": "1", "release_scope": reporting.ReleaseScopeInternalDraft, "output_kind": reporting.OutputKindSlidev})
	}
	for _, transaction := range []string{"binding-render", "binding-rerender"} {
		job := httptestx.RequireSuccessEnvelope(t, release(transaction), http.StatusAccepted)["data"].(map[string]any)
		id := requireSucceededJobResourceID(t, harness, admin, job, "release")
		_, manifest := requireReleaseBundle(t, harness.DB, id)
		if !reflect.DeepEqual(manifest["reference_packs"], binding) {
			t.Fatal("render or rerender changed historical provenance")
		}
	}
	// Lose only the pinned Base host manifest. The active imported replacement
	// remains intact and cannot be silently substituted into the old snapshot.
	if _, err := harness.DB.Exec(`UPDATE reference_pack_objects SET available=false WHERE sha256 IN (SELECT v.manifest_sha256 FROM reference_pack_versions v JOIN reference_pack_candidates c USING(pack_key,pack_version) WHERE c.pack_key='type_registry.host' AND c.distribution_kind='packaged_builtin')`); err != nil {
		t.Fatal(err)
	}
	httptestx.RequireErrorEnvelope(t, release("binding-loss"), http.StatusConflict, expected.Error.Code)
}

func uploadReportingReferencePack(t testing.TB, harness *appsupport.ServerHarness, admin flowtest.LoginResult, container []byte) *http.Response {
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
	request.Header.Set(authn.CSRFHeaderName, admin.CSRFCookie.Value)
	request.AddCookie(admin.SessionCookie)
	request.AddCookie(admin.CSRFCookie)
	return httptestx.Do(t, http.DefaultClient, request)
}
