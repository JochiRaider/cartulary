package reportcomposition

import (
	"encoding/json"
	"errors"

	reportingjson "github.com/JochiRaider/cartulary/internal/modules/reporting/canonicaljson"
	strictjson "github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"golang.org/x/text/unicode/norm"
)

// Materialize only owner-declared defaults and authored-body NFC. Identifiers,
// opaque references and user strings without a normalization rule stay exact.
func materializeCompositionArray(raw json.RawMessage, field string) (json.RawMessage, error) {
	canonical, err := reportingjson.Canonicalize(raw)
	if err != nil {
		return nil, err
	}
	value, err := strictjson.DecodeStrict(canonical)
	if err != nil {
		return nil, err
	}
	items, ok := value.([]any)
	if !ok {
		return nil, errors.New("composition array required")
	}
	for _, value := range items {
		item, ok := value.(map[string]any)
		if !ok {
			return nil, errors.New("composition item object required")
		}
		switch field {
		case "deck_ops":
			if _, present := item["on_unresolved"]; !present {
				item["on_unresolved"] = "fail"
			}
		case "diagram_decls":
			for key, fallback := range map[string]any{"replaces_decl_id": nil, "label_overrides": []any{}, "layout_mode": "auto", "layout": nil} {
				if _, present := item[key]; !present {
					item[key] = fallback
				}
			}
		case "authored_texts":
			if body, ok := item["body"].(string); ok {
				item["body"] = norm.NFC.String(body)
			}
		}
	}
	return canonicalJSON(items)
}

func materializeCompositionBody(object map[string]any) error {
	for _, field := range []string{"deck_ops", "diagram_decls", "authored_texts"} {
		raw, err := canonicalJSON(object[field])
		if err != nil {
			return err
		}
		materialized, err := materializeCompositionArray(raw, field)
		if err != nil {
			return err
		}
		object[field] = materialized
	}
	return nil
}
