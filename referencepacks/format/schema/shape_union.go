package schema

import (
	"encoding/json"
	"slices"

	"github.com/JochiRaider/cartulary/referencepacks/format/canonicaljson"
)

// unionDispatch is derived only from trusted, resolved schema constraints.
// It chooses a branch without evaluating that branch's input validation.
type unionDispatch struct {
	property string
	byKind   bool
	branches map[string]*Shape
}

func literalKey(value any) string {
	b, err := canonicaljson.Marshal(value)
	if err != nil {
		panic("invalid trusted union literal")
	}
	return string(b)
}

func finiteValues(s *Shape) ([]any, bool) {
	if s == nil {
		return nil, false
	}
	if s.hasConstant {
		return []any{s.constant}, true
	}
	if len(s.enum) > 0 {
		return s.enum, true
	}
	if s.kind == "null" {
		return []any{nil}, true
	}
	return nil, false
}

func compileUnion(branches []*Shape) *unionDispatch {
	if len(branches) == 0 {
		return nil
	}
	kinds := &unionDispatch{byKind: true, branches: map[string]*Shape{}}
	disjoint := true
	for _, branch := range branches {
		if branch.kind == "" || kinds.branches[branch.kind] != nil {
			disjoint = false
			break
		}
		kinds.branches[branch.kind] = branch
		if branch.nullable {
			if kinds.branches["null"] != nil {
				disjoint = false
				break
			}
			kinds.branches["null"] = branch
		}
	}
	if disjoint {
		return kinds
	}
	names := slices.Clone(branches[0].required)
	slices.Sort(names)
	for _, name := range names {
		dispatch := &unionDispatch{property: name, branches: map[string]*Shape{}}
		disjoint = true
		for _, branch := range branches {
			if branch.kind != "object" || branch.nullable || !slices.Contains(branch.required, name) {
				disjoint = false
				break
			}
			values, ok := finiteValues(branch.properties[name])
			if !ok {
				disjoint = false
				break
			}
			for _, value := range values {
				key := literalKey(value)
				if previous := dispatch.branches[key]; previous != nil && previous != branch {
					disjoint = false
					break
				}
				dispatch.branches[key] = branch
			}
			if !disjoint {
				break
			}
		}
		if disjoint {
			return dispatch
		}
	}
	return nil
}

func (d *unionDispatch) selectBranch(value any) *Shape {
	if d.byKind {
		return d.branches[diagnosticKind(value)]
	}
	object, ok := value.(map[string]any)
	if !ok {
		return nil
	}
	tag, present := object[d.property]
	if !present {
		return nil
	}
	// Tags are scalar. Never serialize arbitrary hostile arrays or objects merely
	// to choose a branch; ordinary validation will report the bounded union error.
	switch tag.(type) {
	case nil, string, bool, float64, json.Number:
	default:
		return nil
	}
	return d.branches[literalKey(tag)]
}
