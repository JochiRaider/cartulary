package reportcomposition

import (
	"encoding/json"
	"github.com/google/uuid"
	"reflect"
	"strings"
	"testing"
)

func TestCompositionLocalValidationScopeAndOrdering(t *testing.T) {
	for _, raw := range []string{`{"validation_context":null}`, `{"validation_context":{}}`, `{"release_scope":"external_release"}`, `{"snapshot_id":"ignored"}`} {
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
	unsafe := issue("composition_schema_invalid", map[string]any{"composition_op_id": "secret/value", "target_ref": "https://private.invalid", "unexpected": "raw payload", "field": "deck_ops"}).public()
	if details := unsafe["safe_details"].(map[string]any); len(details) != 1 || details["field"] != "deck_ops" {
		t.Fatal("unsafe diagnostic values escaped", unsafe)
	}
}
