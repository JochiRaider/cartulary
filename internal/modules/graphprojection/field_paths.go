package graphprojection

import "strings"

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
	parts := strings.Split(path, ".")
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

type definitionReadContexts map[definitionTarget]fieldScope

func configuredDefinitionReadContexts(config projectionConfig) definitionReadContexts {
	contexts := definitionReadContexts{}
	add := func(scope, kind string, input fieldScope) { contexts[definitionTarget{scope, kind}] |= input }
	for _, mapping := range config.EntityMappings {
		add("vertex", mapping.ProjectedVertexKind, fieldSourceEntity)
	}
	for _, mapping := range config.RelationshipMappings {
		add("edge", mapping.ProjectedEdgeKind, fieldSourceRelationship)
		if mapping.EmitReverseEdge {
			add("edge", mapping.ReverseEdgeKind, fieldSourceRelationship)
		}
	}
	for _, rule := range config.AggregationRules {
		var input fieldScope
		switch rule.InputScope {
		case "source_entity":
			input = fieldSourceEntity
		case "source_relationship":
			input = fieldSourceRelationship
		case "projected_vertex":
			input = fieldProjectedVertex
		case "projected_edge":
			input = fieldProjectedEdge
		}
		add(rule.TargetScope, rule.ProjectedKind, input)
	}
	return contexts
}

func (contexts definitionReadContexts) allows(path, scope, kind string) bool {
	field, ok := classifyFieldPath(path)
	if !ok {
		return false
	}
	if scope == "graph_view" {
		return field.scopes&fieldGraphMetadata != 0
	}
	if scope != "vertex" && scope != "edge" || field.scopes&fieldGraphMetadata != 0 {
		return false
	}
	required := contexts[definitionTarget{scope, kind}]
	if kind == "*" {
		for target, inputs := range contexts {
			if target.scope == scope {
				required |= inputs
			}
		}
	}
	// Empty producer sets are inert. Otherwise every configured input must
	// support the path, regardless of filters or the current source cardinality.
	return field.scopes&required == required
}
