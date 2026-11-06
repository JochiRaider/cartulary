package imports_test

import (
	"bytes"
	"encoding/csv"
	"fmt"
	"net/http"
	"path/filepath"
	"reflect"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/irfixtures"
)

func TestIRCorpusPublicClipboardTransactions_Integration(t *testing.T) {
	root, err := filepath.Abs("../../../fixtures")
	if err != nil {
		t.Fatal(err)
	}
	corpus := irfixtures.ForTest(t, root)
	original, err := corpus.Records("ir-incident/v1/timeline.tsv", '\t')
	if err != nil {
		t.Fatal(err)
	}
	whole, err := corpus.Read("ir-incident/v1/timeline.tsv")
	if err != nil {
		t.Fatal(err)
	}
	var fields []string
	for _, source := range corpus.Manifest.Sources {
		if source.Path == "timeline.tsv" {
			fields = source.FieldKeys
		}
	}
	h := appsupport.StartRuntime(t).StartServer(t, appsupport.ServerOptions{Prefix: "ir-clipboard", TestRouteMode: httptestx.TestRouteModeDisabled})
	login, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, h.Server.HTTP.URL)
	const view = "cartulary.view.timeline.v2"
	for _, count := range []int{499, 500, 501, 1500} {
		t.Run(fmt.Sprint(count), func(t *testing.T) {
			incident := scenariotest.CreateIncident(t, h.Server, login, map[string]any{"client_txn_id": fmt.Sprintf("clipboard-incident-%d", count), "incident_key": fmt.Sprintf("IR-PASTE-%d", count), "title": "IR clipboard transaction"})
			id := incident["incident_id"].(string)
			var text bytes.Buffer
			writer := csv.NewWriter(&text)
			writer.Comma = '\t'
			if err := writer.WriteAll(original[1 : count+1]); err != nil {
				t.Fatal(err)
			}
			content := text.String()
			if count == 1500 {
				content = string(whole)
			}
			targets := make([]map[string]any, count)
			for i := range targets {
				targets[i] = map[string]any{"kind": "create"}
			}
			body := map[string]any{"view_schema_id": view, "client_txn_id": fmt.Sprintf("clipboard-%d", count), "clipboard_text": content, "format": "tsv", "header_mode": "none", "columns": fields, "start_field_key": fields[0], "targets": targets}
			path := "/api/v1/incidents/" + id + "/views/" + view + "/clipboard-paste"
			response := doImportJSON(t, h.Server.HTTP.URL, login, http.MethodPost, path, body)
			if count > 500 {
				httptestx.RequireErrorEnvelope(t, response, http.StatusBadRequest, "invalid_mutation_payload")
				if rows := irQueryAll(t, h.Server.HTTP.URL, login, id, view); len(rows) != 0 {
					t.Fatalf("rejected paste left %d records", len(rows))
				}
				return
			}
			result := httptestx.RequireSuccessEnvelope(t, response, http.StatusOK)["data"].(map[string]any)
			pasted := result["rows"].([]any)
			if len(pasted) != count {
				t.Fatal("paste response lost rows")
			}
			for index, item := range pasted {
				row := item.(map[string]any)
				for column, field := range fields {
					if got := irText(t, row, field); got != original[index+1][column] {
						t.Fatalf("logical record %d field %s changed", index+1, field)
					}
				}
			}
			replay := httptestx.RequireSuccessEnvelope(t, doImportJSON(t, h.Server.HTTP.URL, login, http.MethodPost, path, body), http.StatusOK)["data"].(map[string]any)
			if !reflect.DeepEqual(result["rows"], replay["rows"]) || result["change_set_id"] != replay["change_set_id"] {
				t.Fatal("paste replay changed committed result")
			}
			if rows := irQueryAll(t, h.Server.HTTP.URL, login, id, view); len(rows) != count {
				t.Fatalf("paste/replay stored %d records, want %d", len(rows), count)
			}
		})
	}
}
