package graphprojection

import (
	"encoding/json"
	"strings"
)

type fieldScope uint8

const (
	fieldSourceEntity fieldScope = 1 << iota
	fieldSourceRelationship
	fieldProjectedVertex
	fieldProjectedEdge
	fieldGraphMetadata
)

type classifiedFieldPath struct {
	scopes fieldScope
	member string
	key    string
}

// classifyFieldPath is the lexical owner shared by admission and resolution.
// The returned scopes describe objects that can supply the exact path; callers
// still enforce their own read context. It never translates legacy aliases.
func classifyFieldPath(path string) (classifiedFieldPath, bool) {
	if path == "" || len(path) > graphProjectionLimits.MaxFieldPathBytes {
		return classifiedFieldPath{}, false
	}
	// Bracket notation denotes one literal dotted terminal. Convert it to
	// the same parsed representation used by ordinary qualified keys.
	for _, prefix := range []string{"properties", "metadata", "source_metadata", "projected.properties", "projected.metadata"} {
		if quoted, bracket := strings.CutPrefix(path, prefix+"["); bracket {
			quoted, closed := strings.CutSuffix(quoted, "]")
			var key string
			if !closed || json.Unmarshal([]byte(quoted), &key) != nil || !validPropertyKey(key) || !strings.Contains(key, ".") || quoted != canonicalJSONString(key) {
				return classifiedFieldPath{}, false
			}
			return classifyFieldPartsV2(append(strings.Split(prefix, "."), key))
		}
	}
	return classifyFieldPartsV2(strings.Split(path, "."))
}

func classifyFieldPartsV2(parts []string) (classifiedFieldPath, bool) {
	field := classifiedFieldPath{member: parts[0]}
	if len(parts) == 1 {
		switch parts[0] {
		case "source_entity_id", "source_entity_kind":
			field.scopes = fieldSourceEntity
		case "source_relationship_id", "source_relationship_kind", "src_source_entity_id", "dst_source_entity_id", "direction":
			field.scopes = fieldSourceRelationship
		}
	} else if len(parts) == 2 && parts[0] == "projected" {
		field.member = parts[1]
		switch parts[1] {
		case "vertex_id":
			field.scopes = fieldProjectedVertex
		case "edge_id", "src_vertex_id", "dst_vertex_id", "direction":
			field.scopes = fieldProjectedEdge
		}
	} else if len(parts) == 2 && validPropertyKey(parts[1]) {
		field.key = parts[1]
		switch parts[0] {
		case "properties", "metadata":
			field.scopes = fieldSourceEntity | fieldSourceRelationship
		case "source_metadata":
			field.scopes = fieldGraphMetadata
		}
	} else if len(parts) == 3 && parts[0] == "projected" && validPropertyKey(parts[2]) {
		field.member, field.key = parts[1], parts[2]
		if parts[1] == "properties" || parts[1] == "metadata" && !reservedSystemMetadataTerminal(parts[2]) {
			field.scopes = fieldProjectedVertex | fieldProjectedEdge
		}
	}
	return field, field.scopes != 0
}

func validFieldPath(path string) bool {
	_, ok := classifyFieldPath(path)
	return ok
}

func fieldPathAllows(path string, scopes fieldScope) bool {
	field, ok := classifyFieldPath(path)
	return ok && field.scopes&scopes != 0
}

func validFilterFieldPath(path, scope string) bool {
	switch scope {
	case "source_entity":
		return fieldPathAllows(path, fieldSourceEntity)
	case "source_relationship":
		return fieldPathAllows(path, fieldSourceRelationship)
	default:
		return false
	}
}

func validAggregationFieldPath(path, scope string) bool {
	switch scope {
	case "projected_vertex":
		return fieldPathAllows(path, fieldProjectedVertex)
	case "projected_edge":
		return fieldPathAllows(path, fieldProjectedEdge)
	default:
		return validFilterFieldPath(path, scope)
	}
}

type definitionTarget struct{ scope, kind string }

// One bounded index serves preflight, path applicability, collision checks and emission.
type configuredTargetsV2 struct {
	contexts         map[definitionTarget]fieldScope
	byScope          map[string][]string
	wildcardContexts map[string]fieldScope
}

func (targets configuredTargetsV2) expand(scope, kind string) []string {
	if kind == "*" && scope != "graph_view" {
		return targets.byScope[scope]
	}
	return []string{kind}
}

func (targets configuredTargetsV2) allows(field classifiedFieldPath, scope, kind string) bool {
	if field.scopes == 0 {
		return false
	}
	if scope == "graph_view" {
		return field.scopes&fieldGraphMetadata != 0
	}
	if scope != "vertex" && scope != "edge" || field.scopes&fieldGraphMetadata != 0 {
		return false
	}
	required := targets.contexts[definitionTarget{scope, kind}]
	if kind == "*" {
		required = targets.wildcardContexts[scope]
	}
	// Empty producer sets are inert. Otherwise every configured input must
	// support the path, regardless of filters or the current source cardinality.
	return field.scopes&required == required
}
