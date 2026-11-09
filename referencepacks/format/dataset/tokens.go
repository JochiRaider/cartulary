package dataset

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"

	"github.com/JochiRaider/cartulary/referencepacks/format/canonicaljson"
	"github.com/JochiRaider/cartulary/referencepacks/format/schema"
)

type rowProjection struct {
	fields   map[string]any
	data     json.RawMessage
	lookups  map[string][]json.RawMessage
	postings int
}

// projectRow validates bounded views, never a tree of inert data or a complete
// lookup array. Structural traversal still visits every excess posting, one
// bounded input at a time; only the semantic phase rejects the posting budget.
func projectRow(ctx context.Context, raw []byte, kind, path string, terminated, retain bool, emit schema.FindingSink) (rowProjection, error) {
	p := rowProjection{fields: map[string]any{}, lookups: map[string][]json.RawMessage{}}
	doc, err := canonicaljson.ScanTokens(ctx, raw, MaxDepth)
	if err != nil {
		if errors.Is(err, canonicaljson.ErrDepth) {
			return p, reportFailure(path, excess("max_dataset_json_depth", MaxDepth, MaxDepth+1, true), emit)
		}
		if ctx.Err() != nil {
			return p, ctx.Err()
		}
		return p, reportFailure(path, invalid(), emit)
	}
	if doc.Value.Kind() != '{' {
		return p, reportFailure(path, invalid(), emit)
	}
	if !doc.SafeIntegers {
		if err := emit(schema.Finding{Path: path}); err != nil {
			return p, err
		}
	}
	err = doc.Value.Children(ctx, func(name string, _ int, value canonicaljson.TokenValue) error {
		declared := datasetShapes[kind].Property(name)
		if declared == nil {
			p.fields["_unknown"] = nil
			return nil
		}
		if name == "data" && value.Kind() == '{' {
			p.data = value.Raw()
			p.fields[name] = map[string]any{}
			return nil
		}
		if name == "lookup_values" && value.Kind() == '{' {
			projected := map[string]any{}
			p.fields[name] = projected
			return value.Children(ctx, func(lookup string, i int, inputs canonicaljson.TokenValue) error {
				if i >= MaxLookups {
					return reportFailure(path+".lookup_values", excess("max_dataset_lookups", MaxLookups, uint64(i+1), true), emit)
				}
				if inputs.Kind() != '[' {
					projected[lookup] = nil
					return nil
				}
				projected[lookup] = []any{}
				if retain {
					p.lookups[lookup] = []json.RawMessage{}
				}
				return inputs.Children(ctx, func(_ string, j int, input canonicaljson.TokenValue) error {
					p.postings++
					where := path + ".lookup_values[" + strconv.Itoa(i) + "][" + strconv.Itoa(j) + "]"
					decoded, err := lookupInput(ctx, input)
					if err != nil {
						return reportFailure(where, invalid(), emit)
					}
					if err := declared.Additional().Items().Walk(decoded, where, emit); err != nil {
						return err
					}
					if retain && p.postings <= MaxObjectPostings {
						p.lookups[lookup] = append(p.lookups[lookup], input.Raw())
					}
					return nil
				})
			})
		}
		if name == "source_refs" && value.Kind() == '[' {
			count := 0
			if err := value.Children(ctx, func(_ string, _ int, _ canonicaljson.TokenValue) error {
				count++
				if count > 64 {
					return reportFailure(path+".source_refs", excess("max_source_refs", 64, uint64(count), true), emit)
				}
				return nil
			}); err != nil {
				return err
			}
		}
		// All ordinary envelope fields are scalar or schema-bounded source refs.
		// Wrong container types are represented without decoding their descendants.
		if value.Kind() == '{' || value.Kind() == '[' && name != "source_refs" {
			p.fields[name] = containerPlaceholder(value)
			return nil
		}
		decoded, err := canonicaljson.DecodeStrictBounded(value.Raw(), MaxDepth)
		if err != nil {
			return reportFailure(path, invalid(), emit)
		}
		p.fields[name] = decoded
		return nil
	})
	if err != nil {
		return p, err
	}
	var id *string
	if value, ok := p.fields[kind+"_id"].(string); ok {
		id = attributable(value)
	}
	if err := datasetShapes[kind].Walk(p.fields, path, func(f schema.Finding) error { f.EntryID = id; return emit(f) }); err != nil {
		return p, err
	}
	if !doc.Canonical || !terminated {
		want, actual := "canonical", "noncanonical"
		if err := emit(schema.Finding{Path: path, EntryID: id, Details: schema.SafeDetails{ExpectedToken: &want, ActualToken: &actual}}); err != nil {
			return p, err
		}
	}
	return p, ctx.Err()
}

func containerPlaceholder(value canonicaljson.TokenValue) any {
	if value.Kind() == '{' {
		return map[string]any{}
	}
	return []any{}
}
func lookupInput(ctx context.Context, value canonicaljson.TokenValue) (any, error) {
	if value.Kind() == '[' {
		return []any{}, nil
	}
	if value.Kind() != '{' {
		return canonicaljson.DecodeStrictBounded(value.Raw(), 2)
	}
	result := map[string]any{}
	err := value.Children(ctx, func(name string, i int, component canonicaljson.TokenValue) error {
		if i >= MaxComponents {
			return invalid()
		}
		if component.Kind() == '[' || component.Kind() == '{' {
			result[name] = containerPlaceholder(component)
			return nil
		}
		v, err := canonicaljson.DecodeStrictBounded(component.Raw(), 1)
		result[name] = v
		return err
	})
	return result, err
}

func materializeRow(ctx context.Context, raw []byte, kind string, target any) (rowProjection, error) {
	p, err := projectRow(ctx, raw, kind, "$", true, true, firstFailure("content_schema_invalid"))
	if err != nil {
		return p, err
	}
	// Only the small admitted envelope is decoded into a typed record. Data and
	// lookup values are borrowed immutable byte views, never decoded JSON trees.
	encoded, err := canonicaljson.Marshal(p.fields)
	if err != nil {
		return p, invalid()
	}
	if json.Unmarshal(encoded, target) != nil {
		return p, invalid()
	}
	return p, nil
}
