package imports_test

import (
	"archive/zip"
	"bytes"
	"crypto/sha256"
	"fmt"
	"io"
	"net/http"
	"path"
	"reflect"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/reporting"
)

type irReport struct {
	snapshot, composition, release    string
	snapshotResource, releaseResource map[string]any
	retained                          map[string][]byte
}

const irReportTemplateID = "cartulary.template.ir_summary.v1"

// A report deliberately selects summary fields and the exact declared graph.
// The complete incident remains in its immutable source capture and workbook.
func irReportTemplates(t *testing.T) *reporting.TemplatePackRegistry {
	t.Helper()
	template, ok := reporting.ResolveTemplateContract(reporting.DefaultTemplateID, reporting.DefaultTemplateVersion)
	if !ok {
		t.Fatal("missing built-in template")
	}
	template.TemplateID, template.TemplateVersion = irReportTemplateID, "1"
	template.DeckTitle = "Synthetic IR investigation summary"
	template.AllowedExportModelBindings = []string{"$.records[*].fields['display_name']"}
	template.Sections[0].Bindings = append([]string{}, template.AllowedExportModelBindings...)
	template.Sections[0].Title = "Investigation roster"
	registry, err := reporting.NewTemplatePackRegistry(reporting.TemplatePack{Manifest: template, Files: []reporting.TemplatePackFile{}})
	if err != nil {
		t.Fatal(err)
	}
	return registry
}

func irJobResource(c irPublicClient, path string, body map[string]any, kind string) string {
	c.t.Helper()
	job := c.request(http.MethodPost, path, body, http.StatusAccepted)
	terminal := waitIRApply(c.t, c.origin, c.login, job["job_id"].(string))
	if terminal["status"] != "succeeded" {
		c.t.Fatalf("%s failed: %#v", kind, terminal)
	}
	for _, raw := range terminal["result_summary"].(map[string]any)["resource_refs"].([]any) {
		ref := raw.(map[string]any)
		if ref["kind"] == kind {
			return ref["id"].(string)
		}
	}
	c.t.Fatalf("completed %s has no resource", kind)
	return ""
}

func irCreateReport(c irPublicClient, selected map[string]any) irReport {
	c.t.Helper()
	graph := selected["graph_view"].(map[string]any)
	projection := selected["result"].(map[string]any)["graph_projection_result"].(map[string]any)
	edges := projection["edges"].([]any)
	if len(edges) != 651 {
		c.t.Fatal("report must select the independently expected day-1 graph")
	}
	edge := edges[0].(map[string]any)
	ref := map[string]any{"source_owner_id": "network_flow_activity", "graph_view_id": graph["graph_view_id"]}
	for key, value := range graph["selected_result_binding"].(map[string]any) {
		ref[key] = value
	}
	state := irReport{}
	state.snapshot = irJobResource(c, "/api/v1/snapshots", map[string]any{"incident_id": c.incident, "client_txn_id": "ir-report-snapshot"}, "snapshot")
	state.snapshotResource = c.request(http.MethodGet, "/api/v1/snapshots/"+state.snapshot, nil, http.StatusOK)
	root := "/api/v1/incidents/" + c.incident + "/report-compositions"
	composition := c.request(http.MethodPost, root, map[string]any{
		"client_txn_id": "ir-report-composition", "template_id": irReportTemplateID, "template_version": "1", "deck_ops": []any{},
		"authored_texts": []any{map[string]any{"authored_text_id": "synthetic-handoff", "text_role": "authored_text", "body": "SYNTHETIC IR FIXTURE: retained day-1 analysis. Graph adjacency does not establish compromise.", "disclosure_partition_ref": "internal_only"}},
		"diagram_decls": []any{map[string]any{
			"decl_id": "ir-day1-flow", "replaces_decl_id": nil, "diagram_kind": "flowchart", "diagram_source_kind": "graph", "source_graph_view_id": graph["graph_view_id"], "layout_mode": "auto", "layout": nil, "label_overrides": []any{},
			"selection_rule": map[string]any{"schema_id": "diagram_selection_rule.v1", "rule_id": "explicit_refs", "vertex_refs": []any{edge["src_vertex_id"], edge["dst_vertex_id"]}, "edge_refs": []any{edge["edge_id"]}, "seed_vertex_refs": []any{}, "depth": nil, "edge_kind_filter": []any{}, "timeline_record_ids": []any{}, "overflow_policy": "fail"},
		}},
	}, http.StatusCreated)
	state.composition = composition["composition_id"].(string)
	version := c.request(http.MethodPost, root+"/"+state.composition+"/versions", map[string]any{"client_txn_id": "ir-report-freeze", "base_draft_version": composition["draft_version"]}, http.StatusCreated)
	state.release = irJobResource(c, "/api/v1/releases", map[string]any{
		"snapshot_id": state.snapshot, "client_txn_id": "ir-report-release", "template_id": irReportTemplateID, "template_version": "1",
		"output_kind": reporting.OutputKindMermaid, "output_options": map[string]any{"source_only": true}, "release_scope": reporting.ReleaseScopeInternalDraft,
		"redaction_profile_id": reporting.InternalRedactionProfileID, "redaction_profile_version": reporting.DefaultRedactionProfileVersion,
		"composition_id": state.composition, "composition_version": version["composition_version"], "composition_sha256": version["composition_sha256"], "graph_projection_refs": []any{ref},
	}, "release")
	state.releaseResource = c.request(http.MethodGet, "/api/v1/releases/"+state.release, nil, http.StatusOK)
	if state.releaseResource["release_state"] != "approved" || !reflect.DeepEqual(state.releaseResource["graph_projection_refs"], []any{ref}) {
		c.t.Fatal("approved report did not retain the selected graph")
	}
	bundle := readPublicBundle(c.t, c.origin, c.login, c.incident, "ir-report-before-live-edit", "snapshots")
	state.retained = irReportMembers(c.t, bundle)
	if len(state.retained) < 4 {
		c.t.Fatal("report export omitted retained inputs or output")
	}
	sourceFound := false
	for path, data := range state.retained {
		if strings.HasSuffix(path, ".mmd") && bytes.Contains(data, []byte("flowchart")) {
			sourceFound = true
		}
	}
	if !sourceFound {
		c.t.Fatal("report has no rendered Mermaid source")
	}
	return state
}

func irReportMembers(t *testing.T, bundle []byte) map[string][]byte {
	t.Helper()
	archive, err := zip.NewReader(bytes.NewReader(bundle), int64(len(bundle)))
	if err != nil {
		t.Fatal(err)
	}
	members := map[string][]byte{}
	for _, file := range archive.File {
		if !strings.HasPrefix(file.Name, "ext/snapshots/") {
			continue
		}
		if strings.Contains(file.Name, "/internal/") || strings.Contains(file.Name, "reveal-map") {
			t.Fatal("sensitive reveal map entered incident bundle")
		}
		members[file.Name] = sourceBundleMember(t, bundle, file.Name)
	}
	return members
}

func irVerifyReport(c irPublicClient, state irReport, bundle []byte, native bool) {
	c.t.Helper()
	if native {
		if !reflect.DeepEqual(state.snapshotResource, c.request(http.MethodGet, "/api/v1/snapshots/"+state.snapshot, nil, http.StatusOK)) || !reflect.DeepEqual(state.releaseResource, c.request(http.MethodGet, "/api/v1/releases/"+state.release, nil, http.StatusOK)) {
			c.t.Fatal("live edits changed frozen report inputs or approved output")
		}
	}

	if !native {
		resource := c.request(http.MethodGet, "/api/v1/incidents/"+c.incident+"/retained-report-artifacts", nil, http.StatusOK)
		artifacts := resource["artifacts"].([]any)
		if len(artifacts) != len(state.retained) {
			c.t.Fatal("public historical report inventory changed")
		}
		expected := map[string]string{}
		for name := range state.retained {
			expected[fmt.Sprintf("%x", sha256.Sum256([]byte("cartulary.retained_report_artifact.v1\x00"+c.incident+"\x00"+name)))] = name
		}
		for _, raw := range artifacts {
			artifact := raw.(map[string]any)
			name, ok := expected[artifact["artifact_id"].(string)]
			if !ok {
				c.t.Fatal("unexpected historical report file")
			}
			content := state.retained[name]
			if artifact["filename"] != path.Base(name) || artifact["byte_size"] != float64(len(content)) || artifact["sha256"] != fmt.Sprintf("%x", sha256.Sum256(content)) {
				c.t.Fatal("historical report metadata changed")
			}
			response := doImportJSON(c.t, c.origin, c.login, http.MethodGet, artifact["content_href"].(string), nil)
			actual, err := io.ReadAll(io.LimitReader(response.Body, int64(len(content)+1)))
			response.Body.Close()
			if err != nil || response.StatusCode != http.StatusOK || !bytes.Equal(content, actual) || response.Header.Get("Cache-Control") != "no-store" || response.Header.Get("X-Content-Type-Options") != "nosniff" {
				c.t.Fatal("historical report download changed", name, err)
			}
			delete(expected, artifact["artifact_id"].(string))
		}
	}
	if !reflect.DeepEqual(state.retained, irReportMembers(c.t, bundle)) {
		c.t.Fatal("retained report inputs or approved output changed")
	}
}
