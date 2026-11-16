package reportcomposition

import (
	"encoding/json"
	"github.com/google/uuid"
	"reflect"
	"strings"
	"testing"
)

func TestCompositionLocalValidationScopeAndOrdering(t *testing.T) {
	for _, raw := range []string{`{"validation_context":null}`, `{"validation_context":{}}`, `{"release_scope":"external_release"}`, `{"release_scope":null}`, `{"snapshot_id":"ignored"}`} {
		if _, apiErr := DecodeValidateRequest(strings.NewReader(raw)); apiErr == nil || apiErr.Status != 400 {
			t.Fatal("retired contextual validation input was accepted", raw, apiErr)
		}
	}
	// Source array order wins over validator call order and local ID spelling.
	ops := json.RawMessage(`[{"op_id":"z","op_kind":"unknown","payload":{}},{"op_id":"a","op_kind":"unknown","payload":{}}]`)
	diagrams := json.RawMessage(`[{"decl_id":"diagram","unknown":true}]`)
	texts := json.RawMessage(`[{"authored_text_id":"text","unknown":true}]`)
	summary := validateDraft(ops, diagrams, texts)
	if summary["validation_scope"] != "local_authoring" || summary["valid"] != false || summary["stage"] != "schema_validation" {
		t.Fatal("ambiguous validation scope", summary)
	}
	issues := summary["issues"].([]map[string]any)
	if len(issues) != 4 || issues[0]["safe_details"].(map[string]any)["composition_op_id"] != "z" || issues[1]["safe_details"].(map[string]any)["composition_op_id"] != "a" || issues[2]["safe_details"].(map[string]any)["field"] != "diagram_decls" || issues[3]["safe_details"].(map[string]any)["field"] != "authored_texts" {
		t.Fatal("issue order depended on traversal or ID spelling", issues)
	}
	// A later stage in an earlier source array cannot precede a schema defect.
	badTitle := json.RawMessage(`[{"authored_text_id":"title","text_role":"title_override","body":"bad\nline","disclosure_partition_ref":"internal_only"}]`)
	summary = validateDraft(ops, json.RawMessage(`[]`), badTitle)
	issues = summary["issues"].([]map[string]any)
	if issues[len(issues)-1]["code"] != "authored_title_limit_exceeded" {
		t.Fatal("authored validation ran before schema errors", issues)
	}
	if only := validateDraft(json.RawMessage(`[]`), json.RawMessage(`[]`), badTitle); only["stage"] != "authored_text_validation" {
		t.Fatal("first failing stage was mislabeled", only)
	}
	// Schema beats a bad digest; resource binding beats digest and later semantics.
	record := ResourceRecord{CompositionID: uuid.New(), IncidentID: uuid.New(), TemplateID: "template", TemplateVersion: "1", DeckOps: json.RawMessage(`[]`), DiagramDecls: json.RawMessage(`[]`), AuthoredTexts: json.RawMessage(`[]`)}
	raw, _, err := canonicalComposition(record, 1)
	if err != nil {
		t.Fatal(err)
	}
	object, _ := decodeCompositionObject(raw)
	object["composition_sha256"] = strings.Repeat("0", 64)
	object["deck_ops"] = []any{map[string]any{"op_id": "bad", "op_kind": "unknown", "on_unresolved": "fail", "payload": map[string]any{}}}
	raw, _ = canonicalJSON(object)
	if result := validateInlineComposition(record, raw); result["stage"] != "schema_validation" {
		t.Fatal("digest masked a schema defect", result)
	}
	object["deck_ops"] = []any{}
	object["incident_id"] = uuid.NewString()
	raw, _ = canonicalJSON(object)
	if result := validateInlineComposition(record, raw); result["stage"] != "resource_binding_validation" {
		t.Fatal("digest masked a binding defect", result)
	}
	// Nested source positions and field details stay independent of append order.
	findings := []compositionFinding{
		compositionIssueAt(4, 0)("diagram_layout_invalid", map[string]any{"diagram_id": "d", "target_ref": "edge"}),
		compositionIssueAt(3, 1)("diagram_layout_invalid", map[string]any{"diagram_id": "d", "target_ref": "a"}),
		compositionIssueAt(3, 0)("diagram_layout_unknown_target", map[string]any{"diagram_id": "d", "target_ref": "z"}),
	}
	first, stage := orderedCompositionFindings(findings)
	if stage != "diagram_validation" || first[0]["safe_details"].(map[string]any)["target_ref"] != "z" || first[2]["safe_details"].(map[string]any)["target_ref"] != "edge" {
		t.Fatal("nested source order changed", first)
	}
	reverse := []compositionFinding{findings[2], findings[1], findings[0]}
	second, _ := orderedCompositionFindings(reverse)
	if !reflect.DeepEqual(first, second) {
		t.Fatal("diagnostics depend on append order")
	}
	// Each locally emitted diagram code has a single-defect witness.
	const validDiagram = `{"decl_id":"diagram","diagram_kind":"flowchart","diagram_source_kind":"graph","source_graph_view_id":"mermaid-view","selection_rule":{"schema_id":"diagram_selection_rule.v1","rule_id":"explicit_refs","vertex_refs":["vx_opaque-mermaid"],"edge_refs":[],"seed_vertex_refs":[],"depth":null,"edge_kind_filter":[],"timeline_record_ids":[],"overflow_policy":"fail"}}`
	if clean := validateDraft(json.RawMessage(`[]`), json.RawMessage(`[`+validDiagram+`]`), json.RawMessage(`[]`)); clean["valid"] != true {
		t.Fatal("valid diagram declaration rejected", clean)
	}
	withMember := func(member string) string {
		return strings.Replace(validDiagram, `"decl_id":"diagram",`, `"decl_id":"diagram",`+member+`,`, 1)
	}
	for code, decls := range map[string]string{
		"composition_duplicate_id":               `[` + validDiagram + `,` + validDiagram + `]`,
		"composition_replacement_target_missing": `[` + withMember(`"replaces_decl_id":"other"`) + `]`,
		"raw_generated_source_invalid":           `[` + strings.Replace(validDiagram, `"rule_id":"explicit_refs",`, `"rule_id":"explicit_refs","mermaid_source":"graph TD",`, 1) + `]`,
		"diagram_label_override_invalid":         `[` + withMember(`"label_overrides":[{"target":{"target_kind":"section","ref":"vx_opaque-mermaid"},"label":"Renamed"}]`) + `]`,
	} {
		summary := validateDraft(json.RawMessage(`[]`), json.RawMessage(decls), json.RawMessage(`[]`))
		found := false
		for _, issue := range summary["issues"].([]map[string]any) {
			found = found || issue["code"] == code
		}
		if summary["valid"] != false || !found {
			t.Fatalf("%s was not reported: %#v", code, summary)
		}
	}
	unsafe := issue("composition_schema_invalid", map[string]any{"composition_op_id": "secret/value", "target_ref": "https://private.invalid", "unexpected": "raw payload", "field": "deck_ops"}).public()
	if details := unsafe["safe_details"].(map[string]any); len(details) != 1 || details["field"] != "deck_ops" {
		t.Fatal("unsafe diagnostic values escaped", unsafe)
	}
}
