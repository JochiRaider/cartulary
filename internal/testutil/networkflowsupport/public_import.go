package networkflowsupport

import (
	"bytes"
	"encoding/json"
	"mime"
	"mime/multipart"
	"net/http"
	"net/textproto"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/authn"
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
	"github.com/JochiRaider/cartulary/internal/testutil/fixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/networkflowfixtures"
)

// SeedGraphSource imports the maintained minimal source through ordinary
// discovery, explicit mapping and apply. Process fixtures therefore retain the
// same committed provenance as production, without manufacturing workflow rows.
func SeedGraphSource(t testing.TB, client *http.Client, origin string, login authcookietest.AuthCookies, incident string) string {
	t.Helper()
	runRoot, err := networkflowfixtures.SelectedRunRoot()
	if err != nil {
		t.Fatal(err)
	}
	var table string
	summary, err := networkflowfixtures.Run(networkflowfixtures.Config{CorpusRoot: fixtures.Path("../../..", "fixtures", "network-flow"), RunRoot: runRoot, FixtureID: "NF-FIX-001-cisco-sna-minimal", Revision: 11}, func(bundle *networkflowfixtures.Bundle) error {
		input, err := bundle.Read("source/cases.json")
		if err != nil {
			return err
		}
		var cases []struct {
			Mapping json.RawMessage `json:"mapping"`
		}
		if err := json.Unmarshal(input, &cases); err != nil {
			return err
		}
		if len(cases) != 1 {
			t.Fatal("minimal source requires exactly one mapping")
		}
		source, err := bundle.Read("source/cisco-sna-minimal.csv")
		if err != nil {
			return err
		}
		table = ImportCSV(t, client, origin, login, incident, "process-source", "process.csv", source, cases[0].Mapping)
		return nil
	})
	t.Logf("public graph source comparison: %s", summary)
	if err != nil {
		t.Fatal(err)
	}
	return table
}

// ImportCSV is a fixed public workflow helper, not a fixture interpreter. The
// caller supplies admitted source bytes and an explicitly selected mapping.
func ImportCSV(t testing.TB, client *http.Client, origin string, login authcookietest.AuthCookies, incident, txn, filename string, source []byte, mapping json.RawMessage) string {
	t.Helper()
	options := []func(*http.Request){httptestx.WithCookies(login.Session, login.CSRF), httptestx.WithHeader(authn.CSRFHeaderName, login.CSRF.Value)}
	call := func(method, path string, body any, status int) map[string]any {
		t.Helper()
		return httptestx.RequireSuccessEnvelope(t, httptestx.DoJSON(t, client, method, origin+path, body, options...), status)["data"].(map[string]any)
	}
	await := func(id string) map[string]any {
		t.Helper()
		for {
			job := call(http.MethodGet, "/api/v1/jobs/"+id, nil, http.StatusOK)
			switch job["status"] {
			case "succeeded":
				return job
			case "failed", "canceled":
				t.Fatalf("Network Flow fixture workflow ended as %v", job["status"])
			}
			select {
			case <-t.Context().Done():
				t.Fatal("Network Flow fixture workflow did not complete")
			case <-time.After(50 * time.Millisecond):
			}
		}
	}
	var upload bytes.Buffer
	writer := multipart.NewWriter(&upload)
	part, err := writer.CreatePart(textproto.MIMEHeader{"Content-Disposition": {`form-data; name="metadata"`}, "Content-Type": {"application/json"}})
	if err != nil {
		t.Fatal(err)
	}
	if err := json.NewEncoder(part).Encode(map[string]string{"incident_id": incident, "client_txn_id": txn + "-upload"}); err != nil {
		t.Fatal(err)
	}
	part, err = writer.CreatePart(textproto.MIMEHeader{"Content-Disposition": {mime.FormatMediaType("form-data", map[string]string{"name": "file", "filename": filename})}, "Content-Type": {"text/csv"}})
	if err != nil {
		t.Fatal(err)
	}
	if _, err := part.Write(source); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	request, err := http.NewRequestWithContext(t.Context(), http.MethodPost, origin+"/api/v1/import-sessions", &upload)
	if err != nil {
		t.Fatal(err)
	}
	request.Header.Set("Content-Type", writer.FormDataContentType())
	for _, option := range options {
		option(request)
	}
	response, err := client.Do(request)
	if err != nil {
		t.Fatal(err)
	}
	job := httptestx.RequireSuccessEnvelope(t, response, http.StatusAccepted)["data"].(map[string]any)
	discovered := await(job["job_id"].(string))
	session := ""
	for _, raw := range discovered["result_summary"].(map[string]any)["resource_refs"].([]any) {
		ref := raw.(map[string]any)
		if ref["kind"] == "import_session" {
			session = ref["id"].(string)
		}
	}
	if session == "" {
		t.Fatal("source discovery omitted its session")
	}
	units := call(http.MethodGet, "/api/v1/import-sessions/"+session+"/units", nil, http.StatusOK)["import_units"].([]any)
	if len(units) != 1 {
		t.Fatal("CSV discovery did not produce exactly one unit")
	}
	unit := "/api/v1/import-sessions/" + session + "/units/" + units[0].(map[string]any)["import_unit_id"].(string)
	inspection := call(http.MethodPost, unit+"/source-inspection", map[string]any{"target_kind": "network_flow_table", "extension_profile_id": "network_flow_activity"}, http.StatusOK)
	columns := []map[string]any{}
	for _, raw := range inspection["owner_result"].(map[string]any)["source_columns"].([]any) {
		column := raw.(map[string]any)
		columns = append(columns, map[string]any{"source_column_ordinal": column["source_column_ordinal"], "source_header_text": column["raw_header_text"], "field_key": nil, "entity_binding_mode": nil, "transform_id": nil, "transform_options": map[string]any{}, "empty_value_policy": "omit_field"})
	}
	approval := map[string]any{"target_kind": "network_flow_table", "extension_profile_id": "network_flow_activity", "owner_mapping_schema_id": "cartulary.network_flow.mapping_candidate.v2", "owner_mapping": mapping}
	call(http.MethodPost, unit+"/mapping-preview", approval, http.StatusOK)
	approval["client_txn_id"], approval["header_row_ref"], approval["data_start_row_ref"], approval["source_columns"] = txn+"-approve", 1, 2, columns
	call(http.MethodPut, unit+"/mapping", approval, http.StatusOK)
	call(http.MethodPost, unit+"/select", map[string]string{"client_txn_id": txn + "-select"}, http.StatusOK)
	job = call(http.MethodPost, "/api/v1/import-sessions/"+session+"/apply", map[string]string{"client_txn_id": txn + "-apply"}, http.StatusAccepted)
	completed := await(job["job_id"].(string))
	for _, raw := range completed["result_summary"].(map[string]any)["resource_refs"].([]any) {
		ref := raw.(map[string]any)
		if ref["kind"] == "network_flow_table" {
			return ref["id"].(string)
		}
	}
	t.Fatal("committed flow import omitted its table")
	return ""
}
