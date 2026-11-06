package imports_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/incidents/testsupport/scenariotest"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/authcookietest"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/irfixtures"
	"github.com/JochiRaider/cartulary/internal/testutil/workbookscenariotest"
)

func TestIRCorpusCoreImportLifecycle_Integration(t *testing.T) {
	root, err := filepath.Abs("../../../fixtures")
	if err != nil {
		t.Fatal(err)
	}
	corpus := irfixtures.ForTest(t, root)

	// Admission completes before allocating any application services.
	runtime := appsupport.StartRuntime(t)
	h := runtime.StartServer(t, appsupport.ServerOptions{Prefix: "ir-corpus-core", TestRouteMode: httptestx.TestRouteModeDisabled})
	login, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, h.Server.HTTP.URL)
	incident := scenariotest.CreateIncident(t, h.Server, login, map[string]any{"client_txn_id": "ir-corpus-incident", "incident_key": "IR-CORPUS", "title": "Supplied IR corpus"})
	id := incident["incident_id"].(string)
	irImportCoreCorpus(t, corpus, h, login, id)
	if t.Failed() {
		t.FailNow()
	}
	// Simulate the owner retention boundary only after terminal application.
	// Committed documentary sources have no workflow foreign-key dependency.
	if _, err := h.DB.ExecContext(t.Context(), `DELETE FROM import_sessions WHERE incident_id=$1`, id); err != nil {
		t.Fatal(err)
	}
	retained := irCommittedSources(t, h.Server.HTTP.URL, login, id)
	if len(retained) != 3 {
		t.Fatalf("workflow cleanup removed committed provenance: %d", len(retained))
	}
	for _, name := range []string{"hosts.csv", "identities.csv", "timeline.csv"} {
		original, err := corpus.Read("ir-incident/v1/" + name)
		if err != nil {
			t.Fatal(err)
		}
		irVerifyCommittedSource(t, h.Server.HTTP.URL, login, id, name, "", original, retained)
	}
}

// irImportCoreCorpus drives the authored corpus through production operations.
// Cleanup is separately qualified by the Imports owner, never a lifecycle setup shortcut.
func irImportCoreCorpus(t *testing.T, corpus *irfixtures.Corpus, h *appsupport.ServerHarness, login authcookietest.AuthCookies, id string) {
	t.Helper()
	usersBefore := appsupport.QueryCount(t, h.DB, `SELECT count(*) FROM users`)
	response := doImportJSON(t, h.Server.HTTP.URL, login, http.MethodPut, "/api/v1/incidents/"+id+"/timeline-time-conversion-profile", map[string]any{"base_profile_version": 1, "enabled": true, "local_offset_minutes": -240, "local_label": "UTC-04"})
	httptestx.RequireSuccessEnvelope(t, response, http.StatusOK)
	views := map[string]string{"hosts.csv": "cartulary.view.hosts.v1", "identities.csv": "cartulary.view.identities.v1", "timeline.csv": "cartulary.view.timeline.v2"}
	rosters := map[string]map[string]map[string]any{}
	for sourceIndex, name := range []string{"hosts.csv", "identities.csv", "timeline.csv"} {
		t.Run(name, func(t *testing.T) {
			var source irfixtures.Source
			for _, candidate := range corpus.Manifest.Sources {
				if candidate.Path == name {
					source = candidate
				}
			}
			data, err := corpus.Read("ir-incident/v1/" + name)
			if err != nil {
				t.Fatal(err)
			}
			session, unit := startCSVImportSession(t, h.Server.HTTP.URL, login, id, "ir-upload-"+name, string(data), name)
			columns := []map[string]any{}
			policy := "reject_if_unmapped"
			var binding any = "entity_origin"
			if name == "timeline.csv" {
				policy = "preserve_raw_capture"
				binding = nil
			}
			for i, header := range source.Columns {
				columns = append(columns, map[string]any{"source_column_ordinal": i + 1, "source_header_text": header, "field_key": source.FieldKeys[i], "entity_binding_mode": binding, "transform_id": nil, "transform_options": map[string]any{}, "empty_value_policy": "omit_field"})
			}
			prefix := "/api/v1/import-sessions/" + session
			response := doImportJSON(t, h.Server.HTTP.URL, login, http.MethodPut, prefix+"/units/"+unit+"/mapping", map[string]any{"client_txn_id": "ir-map-" + name, "target_view_schema_id": views[name], "header_row_ref": 1, "data_start_row_ref": 2, "unknown_column_policy": policy, "source_columns": columns})
			httptestx.RequireSuccessEnvelope(t, response, http.StatusOK)
			response = doImportJSON(t, h.Server.HTTP.URL, login, http.MethodPost, prefix+"/units/"+unit+"/select", map[string]any{"client_txn_id": "ir-select-" + name})
			httptestx.RequireSuccessEnvelope(t, response, http.StatusOK)
			apply := map[string]any{"client_txn_id": "ir-apply-" + name}
			response = doImportJSON(t, h.Server.HTTP.URL, login, http.MethodPost, prefix+"/apply", apply)
			job := httptestx.RequireSuccessEnvelope(t, response, http.StatusAccepted)["data"].(map[string]any)
			terminal := waitIRApply(t, h.Server.HTTP.URL, login, job["job_id"].(string))
			if terminal["status"] != "succeeded" {
				t.Fatalf("corpus apply failed: %#v", terminal)
			}
			response = doImportJSON(t, h.Server.HTTP.URL, login, http.MethodPost, prefix+"/apply", apply)
			replay := httptestx.RequireSuccessEnvelope(t, response, http.StatusAccepted)["data"].(map[string]any)
			if replay["job_id"] != job["job_id"] {
				t.Fatal("exact apply replay changed job identity")
			}
			retained := irCommittedSources(t, h.Server.HTTP.URL, login, id)
			if len(retained) != sourceIndex+1 {
				t.Fatalf("replay changed committed provenance count: %d", len(retained))
			}
			irVerifyCommittedSource(t, h.Server.HTTP.URL, login, id, name, unit, data, retained)
			rows := irQueryAll(t, h.Server.HTTP.URL, login, id, views[name])
			if len(rows) != source.Records {
				t.Fatalf("imported %d records, want %d", len(rows), source.Records)
			}
			sourceRows, err := corpus.Records("ir-incident/v1/"+name, ',')
			if err != nil {
				t.Fatal(err)
			}
			keyColumn := 1
			if name == "timeline.csv" {
				keyColumn = 7
			}
			bySource := map[string]map[string]any{}
			timeStates := map[string]int{}
			for _, row := range rows {
				key := irText(t, row, source.FieldKeys[keyColumn])
				if bySource[key] != nil {
					t.Fatal("duplicate source record")
				}
				bySource[key] = row
				if name == "timeline.csv" {
					timeStates[irText(t, row, "timeline.activity_time_pair_state")]++
					for _, field := range []string{"timeline.host_refs", "timeline.identity_refs"} {
						if len(workbookscenariotest.CollectionItems(t, row, field)) != 0 {
							t.Fatalf("source text inferred %s", field)
						}
					}
				}
			}
			for ordinal, cells := range sourceRows[1:] {
				row := bySource[cells[keyColumn]]
				if row == nil {
					t.Fatalf("missing logical source record %d", ordinal+1)
				}
				for i, field := range source.FieldKeys {
					if got := irText(t, row, field); got != cells[i] {
						t.Fatalf("source record %d field %s changed", ordinal+1, field)
					}
				}
			}
			if name == "timeline.csv" && !reflect.DeepEqual(timeStates, map[string]int{"paired_user_preserved": 1450, "empty": 50}) {
				t.Fatalf("time pair states = %#v", timeStates)
			}
			rosters[name] = bySource
			// A new public query must reconstruct the same source inventory.
			reopened := irQueryAll(t, h.Server.HTTP.URL, login, id, views[name])
			if len(reopened) != len(rows) {
				t.Fatal("reopening changed imported inventory")
			}
		})
	}
	if t.Failed() {
		return
	}
	aliases, err := corpus.Records("ir-incident/v1/aliases.csv", ',')
	if err != nil {
		t.Fatal(err)
	}
	type aliasGroup struct {
		kind, canonical string
		aliases         []string
	}
	groups := map[string]*aliasGroup{}
	for _, row := range aliases[1:] {
		key := row[0] + ":" + row[1]
		if groups[key] == nil {
			groups[key] = &aliasGroup{kind: row[0], canonical: row[1]}
		}
		groups[key].aliases = append(groups[key].aliases, row[3])
	}
	keys := make([]string, 0, len(groups))
	for key := range groups {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for index, key := range keys {
		group := groups[key]
		name := "hosts.csv"
		if group.kind == "identity" {
			name = "identities.csv"
		}
		row := rosters[name][group.canonical]
		if row == nil {
			t.Fatalf("alias target missing: %s", key)
		}
		actions := []map[string]any{}
		for _, alias := range group.aliases {
			actions = append(actions, map[string]any{"op": "add_alias", "alias_text": alias})
		}
		body := map[string]any{"view_schema_id": views[name], "base_row_version": row["row_version"], "client_txn_id": fmt.Sprintf("ir-alias-%d", index), "changes": []map[string]any{{"field_key": group.kind + ".aliases", "action_payload": map[string]any{"kind": "collection_actions_v1", "actions": actions}}}}
		response := doImportJSON(t, h.Server.HTTP.URL, login, http.MethodPatch, "/api/v1/records/"+row["record_id"].(string), body)
		patched := httptestx.RequireSuccessEnvelope(t, response, http.StatusOK)["data"].(map[string]any)["row"].(map[string]any)
		response = doImportJSON(t, h.Server.HTTP.URL, login, http.MethodPatch, "/api/v1/records/"+row["record_id"].(string), body)
		replayed := httptestx.RequireSuccessEnvelope(t, response, http.StatusOK)["data"].(map[string]any)["row"].(map[string]any)
		if !reflect.DeepEqual(patched, replayed) {
			t.Fatal("alias replay changed record")
		}
		items := workbookscenariotest.CollectionItems(t, patched, group.kind+".aliases")
		for _, alias := range group.aliases {
			found := false
			for _, item := range items {
				if item["alias_text"] == alias {
					found = true
				}
			}
			if !found {
				encoded, _ := json.Marshal(items)
				t.Fatalf("alias %q missing from %s", alias, encoded)
			}
		}
	}
	for _, ambiguity := range []struct{ kind, alias, source string }{
		{"host", "research-files", "hosts.csv"},
		{"identity", "nsr-admin", "identities.csv"},
	} {
		want := map[string]bool{}
		for _, row := range aliases[1:] {
			if row[0] == ambiguity.kind && row[3] == ambiguity.alias {
				want[rosters[ambiguity.source][row[1]]["record_id"].(string)] = true
			}
		}
		if len(want) != 2 {
			t.Fatalf("ambiguous source association changed: %s", ambiguity.alias)
		}
		response := doImportJSON(t, h.Server.HTTP.URL, login, http.MethodGet,
			"/api/v1/incidents/"+id+"/entity-candidates?entity_type="+ambiguity.kind+"&search="+url.QueryEscape(ambiguity.alias), nil)
		candidates := httptestx.RequireSuccessEnvelope(t, response, http.StatusOK)["data"].(map[string]any)["candidates"].([]any)
		got := map[string]bool{}
		for _, candidate := range candidates {
			got[candidate.(map[string]any)["record_id"].(string)] = true
		}
		if !reflect.DeepEqual(got, want) {
			t.Fatalf("shared alias %s no longer exposes both intended candidates", ambiguity.alias)
		}
	}
	for _, view := range []string{"cartulary.view.evidence.v1", "cartulary.view.indicators.v1"} {
		if rows := irQueryAll(t, h.Server.HTTP.URL, login, id, view); len(rows) != 0 {
			t.Fatalf("text created implicit records in %s", view)
		}
	}
	if usersAfter := appsupport.QueryCount(t, h.DB, `SELECT count(*) FROM users`); usersAfter != usersBefore {
		t.Fatal("source documentation created authenticated users")
	}
}

func irCommittedSources(t *testing.T, baseURL string, login authcookietest.AuthCookies, incidentID string) []map[string]any {
	t.Helper()
	var result []map[string]any
	path := "/api/v1/incidents/" + incidentID + "/committed-sources?limit=1"
	for {
		envelope := httptestx.RequireSuccessEnvelope(t, doImportJSON(t, baseURL, login, http.MethodGet, path, nil), http.StatusOK)
		for _, row := range envelope["data"].(map[string]any)["committed_sources"].([]any) {
			result = append(result, row.(map[string]any))
		}
		meta := envelope["meta"].(map[string]any)
		token, _ := meta["paging"].(map[string]any)["next_cursor"].(string)
		if token == "" {
			break
		}
		path = "/api/v1/incidents/" + incidentID + "/committed-sources?cursor_token=" + url.QueryEscape(token)
	}
	return result
}

func irVerifyCommittedSource(t *testing.T, baseURL string, login authcookietest.AuthCookies, incidentID, name, unit string, original []byte, retained []map[string]any) {
	t.Helper()
	var selected map[string]any
	for _, source := range retained {
		if source["original_filename"] == name {
			if selected != nil {
				t.Fatal("duplicate committed source after replay")
			}
			selected = source
		}
	}
	if selected == nil {
		t.Fatalf("missing committed source %s", name)
	}
	path := "/api/v1/incidents/" + incidentID + "/committed-sources/" + selected["committed_source_id"].(string)
	metadata := httptestx.RequireSuccessEnvelope(t, doImportJSON(t, baseURL, login, http.MethodGet, path, nil), http.StatusOK)["data"].(map[string]any)
	if metadata["content_sha256"] != fmt.Sprintf("%x", sha256.Sum256(original)) || metadata["original_filename"] != name || metadata["byte_size"] != float64(len(original)) {
		t.Fatal("committed source metadata changed")
	}
	if unit != "" && metadata["origin_import_unit_id"] != unit {
		t.Fatal("committed source lost its inert origin")
	}
	if metadata["header_row_ref"] != float64(1) || metadata["data_start_row_ref"] != float64(2) {
		t.Fatal("committed source lost its approved source row selection")
	}
	if metadata["approved_mapping"].(map[string]any)["source_columns"] == nil {
		t.Fatal("committed source lost approved mapping")
	}
	response := doImportJSON(t, baseURL, login, http.MethodGet, path+"/content", nil)
	if got := httptestx.ReadResponseBody(t, response, http.StatusOK); !bytes.Equal(got, original) {
		t.Fatal("committed source bytes changed")
	}
	if response.Header.Get("Cache-Control") != "no-store" || response.Header.Get("X-Content-Type-Options") != "nosniff" {
		t.Fatal("source download protection headers missing")
	}
}

// Full-corpus apply is governed by the selected harness test's deadline. The
// five-second helper for tiny import examples is not a corpus performance SLA.
func waitIRApply(t *testing.T, url string, login authcookietest.AuthCookies, jobID string) map[string]any {
	t.Helper()
	deadline, ok := t.Deadline()
	if !ok {
		t.Fatal("full-corpus apply requires a harness deadline")
	}
	ctx, cancel := context.WithDeadline(t.Context(), deadline)
	defer cancel()
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
	for {
		response := doImportJSON(t, url, login, http.MethodGet, "/api/v1/jobs/"+jobID, nil)
		job := httptestx.RequireSuccessEnvelope(t, response, http.StatusOK)["data"].(map[string]any)
		switch job["status"] {
		case "succeeded", "failed", "canceled":
			return job
		}
		select {
		case <-ctx.Done():
			t.Fatalf("corpus apply did not terminate: %v", ctx.Err())
		case <-ticker.C:
		}
	}
}

func irText(t testing.TB, row map[string]any, key string) string {
	t.Helper()
	cell, ok := row["cells"].(map[string]any)[key].(map[string]any)
	if !ok {
		t.Fatalf("missing field %s", key)
	}
	if cell["value"] == nil {
		return ""
	}
	value, ok := cell["value"].(string)
	if !ok {
		t.Fatalf("nontext field %s", key)
	}
	return value
}

func irQueryAll(t testing.TB, url string, login authcookietest.AuthCookies, incident, view string) []map[string]any {
	t.Helper()
	rows := []map[string]any{}
	seen := map[string]bool{}
	body := map[string]any{"limit": 200}
	for page := 0; page < 20; page++ {
		response := doImportJSON(t, url, login, http.MethodPost, "/api/v1/incidents/"+incident+"/views/"+view+"/query", body)
		envelope := httptestx.RequireSuccessEnvelope(t, response, http.StatusOK)
		for _, value := range envelope["data"].(map[string]any)["rows"].([]any) {
			row := value.(map[string]any)
			id := row["record_id"].(string)
			if seen[id] {
				t.Fatal("duplicate paginated row")
			}
			seen[id] = true
			rows = append(rows, row)
		}
		paging := envelope["meta"].(map[string]any)["paging"].(map[string]any)
		if paging["has_more"] == false {
			return rows
		}
		cursor, ok := paging["next_cursor"].(string)
		if !ok || strings.TrimSpace(cursor) == "" {
			t.Fatal("missing continuation")
		}
		body = map[string]any{"cursor_token": cursor}
	}
	t.Fatal("unexpectedly unbounded corpus query")
	return nil
}
