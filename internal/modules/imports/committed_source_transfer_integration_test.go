package imports_test

import (
	"archive/zip"
	"bytes"
	"encoding/json"
	"io"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"reflect"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestCommittedSourcesTransferThroughPublicBundleOperations_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	source := runtime.StartServer(t, appsupport.ServerOptions{Prefix: "committed-source-a", TestRouteMode: httptestx.TestRouteModeDisabled})
	target := runtime.StartServer(t, appsupport.ServerOptions{Prefix: "committed-source-b", TestRouteMode: httptestx.TestRouteModeDisabled})
	sourceLogin, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, source.Server.HTTP.URL)
	targetLogin, targetActor := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, target.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, source.Server, sourceLogin, map[string]any{"client_txn_id": "committed-source-incident", "incident_key": "COMMITTED-SOURCES", "title": "Immutable original sources"})
	id := incident["incident_id"].(string)
	original := []byte("Title,Unmapped\r\n\"Literal <script>界</script>\",\"original\r\nbytes\"\r\n")
	var originSession string
	// Two successful applications share the blob, but preserve separate origins.
	for _, txn := range []string{"first", "second"} {
		session, unit := startCSVImportSession(t, source.Server.HTTP.URL, sourceLogin, id, "upload-"+txn, string(original), "original.csv")
		originSession = session
		prefix := "/api/v1/import-sessions/" + session
		columns := []map[string]any{
			{"source_column_ordinal": 1, "source_header_text": "Title", "field_key": "timeline.activity_synopsis_text", "entity_binding_mode": nil, "transform_id": nil, "transform_options": map[string]any{}, "empty_value_policy": "omit_field"},
			{"source_column_ordinal": 2, "source_header_text": "Unmapped", "field_key": nil, "entity_binding_mode": nil, "transform_id": nil, "transform_options": map[string]any{}, "empty_value_policy": "omit_field"},
		}
		httptestx.RequireSuccessEnvelope(t, doImportJSON(t, source.Server.HTTP.URL, sourceLogin, http.MethodPut, prefix+"/units/"+unit+"/mapping", map[string]any{"client_txn_id": "map-" + txn, "target_view_schema_id": "cartulary.view.timeline.v2", "header_row_ref": 1, "data_start_row_ref": 2, "unknown_column_policy": "preserve_raw_capture", "source_columns": columns}), http.StatusOK)
		httptestx.RequireSuccessEnvelope(t, doImportJSON(t, source.Server.HTTP.URL, sourceLogin, http.MethodPost, prefix+"/units/"+unit+"/select", map[string]any{"client_txn_id": "select-" + txn}), http.StatusOK)
		job := httptestx.RequireSuccessEnvelope(t, doImportJSON(t, source.Server.HTTP.URL, sourceLogin, http.MethodPost, prefix+"/apply", map[string]any{"client_txn_id": "apply-" + txn}), http.StatusAccepted)["data"].(map[string]any)
		if terminal := waitIRApply(t, source.Server.HTTP.URL, sourceLogin, job["job_id"].(string)); terminal["status"] != "succeeded" {
			t.Fatalf("source import failed: %#v", terminal)
		}
	}
	before := irCommittedSources(t, source.Server.HTTP.URL, sourceLogin, id)
	if len(before) != 2 {
		t.Fatal("successful application identities collapsed")
	}
	bundle := readPublicBundle(t, source.Server.HTTP.URL, sourceLogin, id, "export")
	zipReader, err := zip.NewReader(bytes.NewReader(bundle), int64(len(bundle)))
	if err != nil {
		t.Fatal(err)
	}
	sourceMembers := 0
	for _, member := range zipReader.File {
		if len(member.Name) >= len("sources/sha256/") && member.Name[:len("sources/sha256/")] == "sources/sha256/" {
			sourceMembers++
		}
	}
	if sourceMembers != 1 {
		t.Fatalf("source byte deduplication produced %d members", sourceMembers)
	}
	job := uploadPublicSourceBundle(t, target.Server.HTTP.URL, targetLogin, bundle, "transfer")
	if terminal := waitIRApply(t, target.Server.HTTP.URL, targetLogin, job["job_id"].(string)); terminal["status"] != "succeeded" {
		t.Fatalf("transfer failed: %#v", terminal)
	}
	replay := uploadPublicSourceBundle(t, target.Server.HTTP.URL, targetLogin, bundle, "transfer")
	if replay["job_id"] != job["job_id"] {
		t.Fatal("exact transfer replay duplicated the operation")
	}
	after := irCommittedSources(t, target.Server.HTTP.URL, targetLogin, id)
	if !reflect.DeepEqual(before, after) {
		t.Fatal("committed source summaries changed across transfer")
	}
	for _, summary := range before {
		path := "/api/v1/incidents/" + id + "/committed-sources/" + summary["committed_source_id"].(string)
		originalMetadata := httptestx.RequireSuccessEnvelope(t, doImportJSON(t, source.Server.HTTP.URL, sourceLogin, http.MethodGet, path, nil), http.StatusOK)["data"]
		importedMetadata := httptestx.RequireSuccessEnvelope(t, doImportJSON(t, target.Server.HTTP.URL, targetLogin, http.MethodGet, path, nil), http.StatusOK)["data"]
		if !reflect.DeepEqual(originalMetadata, importedMetadata) {
			t.Fatalf("portable source metadata changed: original=%#v imported=%#v", originalMetadata, importedMetadata)
		}
		response := doImportJSON(t, target.Server.HTTP.URL, targetLogin, http.MethodGet, path+"/content", nil)
		if response.StatusCode != http.StatusOK {
			t.Fatalf("source content status=%d", response.StatusCode)
		}
		actual, err := io.ReadAll(response.Body)
		_ = response.Body.Close()
		if err != nil || !bytes.Equal(actual, original) {
			t.Fatal("exact source bytes changed after transfer", err)
		}
	}
	response := doImportJSON(t, target.Server.HTTP.URL, targetLogin, http.MethodGet, "/api/v1/import-sessions/"+originSession, nil)
	if response.StatusCode != http.StatusNotFound {
		t.Fatalf("source workflow copied or accessible at destination: %d", response.StatusCode)
	}
	_ = response.Body.Close()
	response = doImportJSON(t, target.Server.HTTP.URL, sourceLogin, http.MethodGet, "/api/v1/incidents/"+id+"/committed-sources", nil)
	if response.StatusCode != http.StatusUnauthorized {
		t.Fatalf("source credentials authenticated to destination: %d", response.StatusCode)
	}
	_ = response.Body.Close()
	again := readPublicBundle(t, target.Server.HTTP.URL, targetLogin, id, "reexport")
	if !bytes.Equal(sourceBundleMember(t, bundle, "data/committed_sources.ndjson"), sourceBundleMember(t, again, "data/committed_sources.ndjson")) {
		t.Fatal("re-export changed historical provenance")
	}
	// Subsequent destination work must supersede only its mutable attribution.
	// The importer is the physical FK actor before and after this edit, so
	// comparing only user IDs would incorrectly retain the original author.
	rows := irQueryAll(t, target.Server.HTTP.URL, targetLogin, id, "cartulary.view.timeline.v2")
	client := irPublicClient{t: t, origin: target.Server.HTTP.URL, login: targetLogin, incident: id}
	edited := client.patch("cartulary.view.timeline.v2", "destination-edit", rows[0], irValue("timeline.activity_synopsis_text", "Destination investigation continues"))
	changed := readPublicBundle(t, client.origin, targetLogin, id, "export-after-destination-edit")
	var originalRow, changedRow map[string]any
	for _, pair := range []struct {
		archive []byte
		target  *map[string]any
	}{{bundle, &originalRow}, {changed, &changedRow}} {
		decoder := json.NewDecoder(bytes.NewReader(sourceBundleMember(t, pair.archive, "data/records.ndjson")))
		for decoder.More() {
			var row map[string]any
			if err := decoder.Decode(&row); err != nil {
				t.Fatal(err)
			}
			if row["record_id"] == edited["record_id"] {
				*pair.target = row
			}
		}
	}
	if originalRow == nil || changedRow == nil || originalRow["created_by_user_id"] != changedRow["created_by_user_id"] || changedRow["updated_by_user_id"] != targetActor.String() || originalRow["updated_by_user_id"] == changedRow["updated_by_user_id"] {
		t.Fatal("destination edit lost creation attribution or retained the source updater")
	}
	if !bytes.Equal(sourceBundleMember(t, bundle, "data/committed_sources.ndjson"), sourceBundleMember(t, changed, "data/committed_sources.ndjson")) {
		t.Fatal("destination edit changed immutable source attribution")
	}
}

func readPublicBundle(t *testing.T, base string, login authcookietest.AuthCookies, id, txn string, optionalSections ...string) []byte {
	t.Helper()
	if optionalSections == nil {
		optionalSections = []string{}
	}
	job := httptestx.RequireSuccessEnvelope(t, doImportJSON(t, base, login, http.MethodPost, "/api/v1/incident-bundles/export", map[string]any{"incident_id": id, "client_txn_id": txn, "optional_sections": optionalSections}), http.StatusAccepted)["data"].(map[string]any)
	terminal := waitIRApply(t, base, login, job["job_id"].(string))
	if terminal["status"] != "succeeded" {
		t.Fatalf("bundle export failed: %#v", terminal)
	}
	ref := terminal["result_summary"].(map[string]any)["resource_refs"].([]any)[0].(map[string]any)
	response := doImportJSON(t, base, login, http.MethodGet, "/api/v1/incident-bundles/"+ref["id"].(string)+"/content", nil)
	defer response.Body.Close()
	if response.StatusCode != http.StatusOK || response.Header.Get("Content-Type") != "application/zip" || response.Header.Get("Cache-Control") != "no-store" || response.Header.Get("X-Content-Type-Options") != "nosniff" {
		t.Fatalf("bundle content response status=%d headers=%v", response.StatusCode, response.Header)
	}
	content, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	return content
}
func uploadPublicSourceBundle(t *testing.T, base string, login authcookietest.AuthCookies, content []byte, txn string) map[string]any {
	t.Helper()
	var body bytes.Buffer
	writer := multipart.NewWriter(&body)
	part, err := writer.CreatePart(textproto.MIMEHeader{"Content-Disposition": {`form-data; name="metadata"`}, "Content-Type": {"application/json"}})
	if err != nil {
		t.Fatal(err)
	}
	if err := json.NewEncoder(part).Encode(map[string]any{"client_txn_id": txn}); err != nil {
		t.Fatal(err)
	}
	part, err = writer.CreatePart(textproto.MIMEHeader{"Content-Disposition": {`form-data; name="file"; filename="incident.zip"`}, "Content-Type": {"application/zip"}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(content); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	request, err := http.NewRequest(http.MethodPost, base+"/api/v1/incident-bundles/import", &body)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", writer.FormDataContentType())
	request.Header.Set(authn.CSRFHeaderName, login.CSRF.Value)
	request.AddCookie(login.Session)
	request.AddCookie(login.CSRF)
	return httptestx.RequireSuccessEnvelope(t, httptestx.Do(t, http.DefaultClient, request), http.StatusAccepted)["data"].(map[string]any)
}
func sourceBundleMember(t *testing.T, bundle []byte, path string) []byte {
	t.Helper()
	reader, err := zip.NewReader(bytes.NewReader(bundle), int64(len(bundle)))
	if err != nil {
		t.Fatal(err)
	}
	for _, member := range reader.File {
		if member.Name == path {
			source, err := member.Open()
			if err != nil {
				t.Fatal(err)
			}
			defer source.Close()
			content, err := io.ReadAll(source)
			if err != nil {
				t.Fatal(err)
			}
			return content
		}
	}
	t.Fatalf("missing bundle member %s", path)
	return nil
}
