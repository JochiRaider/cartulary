package reportcomposition

import (
	"bytes"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/google/uuid"
)

func TestCompositionRenderSourceAdmission_Unit(t *testing.T) {
	for _, ref := range []string{"public", "internal_only", "party:é", "party:" + strings.Repeat("p", 128)} {
		if !validAuthoredDisclosurePartition(ref) {
			t.Fatal("valid authored partition rejected", ref)
		}
	}
	for _, ref := range []string{"", "internal", "blocked", "public ", "party:", "party:a/b", "party:a:b", "party:a\\b", "party:a#b", "party:a\u2003b", "party:a\u0085b", "party:" + strings.Repeat("p", 129)} {
		raw, _ := json.Marshal([]any{map[string]any{"authored_text_id": "text", "text_role": "authored_text", "body": "Text", "disclosure_partition_ref": ref}})
		if len(validateAuthoredTexts(raw, map[string]string{})) == 0 {
			t.Fatal("invalid authored partition admitted", ref)
		}
	}
	record := ResourceRecord{CompositionID: uuid.New(), IncidentID: uuid.New(), TemplateID: "template", TemplateVersion: "1", DraftVersion: 3, DeckOps: json.RawMessage(`[]`), DiagramDecls: json.RawMessage(`[]`), AuthoredTexts: json.RawMessage(`[]`)}
	raw, digest, err := canonicalComposition(record, 2)
	if err != nil {
		t.Fatal(err)
	}
	decoded, err := DecodeRenderSource(raw)
	if err != nil || decoded.SchemaID != CompositionSchemaID || decoded.CompositionSHA256 == nil || *decoded.CompositionSHA256 != digest {
		t.Fatalf("immutable source: %#v %v", decoded, err)
	}
	version := &VersionRecord{CanonicalComposition: raw, CompositionSHA256: digest, CompositionVersion: 2}
	for _, kind := range []string{"draft", "version"} {
		var selected *VersionRecord
		if kind == "version" {
			selected = version
		}
		preview, _, _, _, err := previewSource(record, selected, PreviewRequest{SourceKind: kind})
		if err != nil {
			t.Fatal(err)
		}
		got, err := DecodeRenderSource(preview)
		if err != nil || got.SourceKind != kind || (got.DraftVersion != nil) != (kind == "draft") {
			t.Fatalf("preview source: %#v %v", got, err)
		}
		object, _ := decodeCompositionObject(preview)
		if kind == "draft" {
			object["composition_sha256"] = digest
		} else {
			object["composition_sha256"] = strings.Repeat("f", 64)
		}
		assertRenderSourceRejects(t, resealRenderSource(t, object, "preview_source_sha256"))
	}
	for name, change := range map[string]func(map[string]any){
		"unknown member":            func(o map[string]any) { o["future"] = true },
		"missing schema":            func(o map[string]any) { delete(o, "schema_id") },
		"retired schema":            func(o map[string]any) { o["schema_id"] = "cartulary.report_composition.v0" },
		"missing required nullable": func(o map[string]any) { delete(o, "authored_against_snapshot_id") },
		"null arrays":               func(o map[string]any) { o["deck_ops"] = nil },
		"null array item":           func(o map[string]any) { o["deck_ops"] = []any{nil} },
		"invalid identifier":        func(o map[string]any) { o["template_id"] = "outside/path" },
		"version alias":             func(o map[string]any) { o["composition_version"] = "latest" },
		"version null":              func(o map[string]any) { o["composition_version"] = nil },
		"unknown operation field": func(o map[string]any) {
			o["deck_ops"] = []any{map[string]any{"op_id": "op", "op_kind": "exclude_section", "payload": map[string]any{}, "ignored": "secret"}}
		},
		"unknown authored field": func(o map[string]any) {
			o["authored_texts"] = []any{map[string]any{"authored_text_id": "text", "text_role": "authored_text", "body": "safe", "disclosure_partition_ref": "internal_only", "ignored": "secret"}}
		},
	} {
		t.Run(name, func(t *testing.T) {
			o, _ := decodeCompositionObject(raw)
			change(o)
			assertRenderSourceRejects(t, resealRenderSource(t, o, "composition_sha256"))
		})
	}
	for _, altered := range [][]byte{
		append([]byte(" "), raw...),
		bytes.Replace(raw, []byte(digest), []byte(strings.Repeat("0", 64)), 1),
		bytes.Replace(raw, []byte(`"deck_ops":[]`), []byte(`"deck_ops":[],"deck_ops":[]`), 1),
		bytes.Replace(raw, []byte(`"deck_ops":[]`), []byte(`"deck_ops":[{"op_id":"x","op_id":"y"}]`), 1),
		[]byte(`null`), []byte(`[]`),
	} {
		assertRenderSourceRejects(t, altered)
	}
	t.Run("defaults NFC and opaque identifiers", func(t *testing.T) {
		draft := record
		draft.DeckOps = json.RawMessage(`[{"op_id":"op","op_kind":"override_title","payload":{"section_anchor":{"anchor_kind":"section_anchor","template_section_decl_id":"mermaid-section","expansion_key":null},"authored_text_ref":"text"}}]`)
		draft.AuthoredTexts = json.RawMessage(`[{"authored_text_id":"text","text_role":"title_override","body":"Cafe\u0301","disclosure_partition_ref":"internal_only"}]`)
		draft.DiagramDecls = json.RawMessage(`[{"decl_id":"diagram","diagram_kind":"flowchart","diagram_source_kind":"graph","source_graph_view_id":"mermaid-view","selection_rule":{"schema_id":"diagram_selection_rule.v1","rule_id":"explicit_refs","vertex_refs":["vx_opaque-mermaid"],"edge_refs":[],"seed_vertex_refs":[],"depth":null,"edge_kind_filter":[],"timeline_record_ids":[],"overflow_policy":"fail"}}]`)
		first, _, err := canonicalComposition(draft, 2)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := DecodeRenderSource(first); err != nil {
			t.Fatal(err)
		}
		explicit := draft
		explicit.DeckOps = bytes.Replace(draft.DeckOps, []byte(`"op_kind":"override_title"`), []byte(`"op_kind":"override_title","on_unresolved":"fail"`), 1)
		explicit.AuthoredTexts = bytes.Replace(draft.AuthoredTexts, []byte(`Cafe\u0301`), []byte("Café"), 1)
		explicit.DiagramDecls = bytes.Replace(draft.DiagramDecls, []byte(`"decl_id":"diagram"`), []byte(`"decl_id":"diagram","replaces_decl_id":null,"label_overrides":[],"layout_mode":"auto","layout":null`), 1)
		second, _, err := canonicalComposition(explicit, 2)
		if err != nil || !bytes.Equal(first, second) {
			t.Fatalf("default/normalization mismatch: %v", err)
		}
		for name, mutate := range map[string]func(map[string]any){
			"selection unknown":            func(o map[string]any) { o["selection_rule"].(map[string]any)["future"] = "secret" },
			"selection null array":         func(o map[string]any) { o["selection_rule"].(map[string]any)["edge_refs"] = nil },
			"selection source mismatch":    func(o map[string]any) { o["diagram_source_kind"] = "timeline"; o["source_graph_view_id"] = nil },
			"null layout mode":             func(o map[string]any) { o["layout_mode"] = nil },
			"missing materialized default": func(o map[string]any) { delete(o, "label_overrides") },
			"unknown diagram kind":         func(o map[string]any) { o["diagram_kind"] = "future" },
		} {
			t.Run(name, func(t *testing.T) {
				object, _ := decodeCompositionObject(first)
				mutate(object["diagram_decls"].([]any)[0].(map[string]any))
				assertRenderSourceRejects(t, resealRenderSource(t, object, "composition_sha256"))
			})
		}
		object, _ := decodeCompositionObject(first)
		object["authored_texts"].([]any)[0].(map[string]any)["body"] = "Cafe\u0301"
		assertRenderSourceRejects(t, resealRenderSource(t, object, "composition_sha256"))
	})
}

func resealRenderSource(t *testing.T, object map[string]any, digestMember string) []byte {
	t.Helper()
	delete(object, digestMember)
	base, err := canonicalJSON(object)
	if err != nil {
		t.Fatal(err)
	}
	object[digestMember] = hashHex(base)
	sealed, err := canonicalJSON(object)
	if err != nil {
		t.Fatal(err)
	}
	return sealed
}

func assertRenderSourceRejects(t *testing.T, raw []byte) {
	t.Helper()
	_, err := DecodeRenderSource(raw)
	var sourceErr *RenderSourceError
	if !errors.As(err, &sourceErr) || sourceErr.ReasonCode == "" || strings.Contains(err.Error(), "secret") {
		t.Fatalf("expected bounded source rejection, got %v", err)
	}
}
