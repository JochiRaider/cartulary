package graphprojection

import (
	"context"
	"maps"
)

type resultRegistryIndexV2 struct {
	vertices, edges                    map[string]resultKindProvenanceV2
	entityOutputs, relationshipOutputs map[string]int
	rules                              map[string]bool
	properties, metadata               resultDeclarationsV2
}

type resultDeclarationsV2 map[definitionTarget]*resultFieldSetV2

type resultFieldSetV2 struct {
	fields    map[string]resultFieldV2
	mandatory int
}

type resultFieldV2 struct {
	projectedType          string
	canEmitNull, mandatory bool
}

func (fields resultDeclarationsV2) admits(scope, kind string, raw any) bool {
	values, ok := raw.(map[string]any)
	if !ok || len(values) > graphProjectionLimits.MaxPropertyKeys {
		return false
	}
	set := fields[definitionTarget{scope, kind}]
	if set == nil {
		return len(values) == 0
	}
	present := 0
	for key, value := range values {
		field, declared := set.fields[key]
		if !declared {
			return false
		}
		if field.mandatory {
			present++
		}
		if value == nil {
			if !field.canEmitNull {
				return false
			}
		} else if !valueMatchesType(field.projectedType, value) {
			return false
		}
	}
	return present == set.mandatory
}

// The nested property declarations and the expanded top-level declarations
// are two representations of the same references. Consume each match once.
func indexResultRegistryV2(registry map[string]any) (resultRegistryIndexV2, bool) {
	index := resultRegistryIndexV2{rules: make(map[string]bool), properties: make(resultDeclarationsV2), metadata: make(resultDeclarationsV2)}
	if !registryDeclarationsFitV2(registry) {
		return index, false
	}
	references := make(map[declarationKeyV2]PropertySchemaReference)
	var ok bool
	index.vertices, ok = index.indexKinds(registry["vertex_kinds"].([]any), "vertex", "source_entity_kinds", graphProjectionLimits.MaxEntityMappings, references)
	if !ok {
		return index, false
	}
	index.edges, ok = index.indexKinds(registry["edge_kinds"].([]any), "edge", "source_relationship_kinds", graphProjectionLimits.MaxRelationshipMappings, references)
	if !ok {
		return index, false
	}
	return index, index.validDeclarations(registry["property_keys"].([]any), "projected_key", index.properties, references) && len(references) == 0 && index.validDeclarations(registry["metadata_keys"].([]any), "projected_metadata_key", index.metadata, nil)
}

func registryDeclarationsFitV2(registry map[string]any) bool {
	properties := len(registry["property_keys"].([]any))
	if properties > graphProjectionLimits.MaxExpandedRegistryDeclarations-len(registry["metadata_keys"].([]any)) {
		return false
	}
	// Bound nested references before retaining any of them, even for malformed registries.
	remaining := properties
	for _, scope := range []string{"vertex_kinds", "edge_kinds"} {
		for _, raw := range registry[scope].([]any) {
			row, _ := raw.(map[string]any)
			refs, _ := row["properties"].([]any)
			remaining -= len(refs)
			if remaining < 0 {
				return false
			}
		}
	}
	return true
}

func registryPropertyRefV2(row map[string]any, keyField string) (PropertySchemaReference, bool) {
	key, _ := row[keyField].(string)
	typ, _ := row["projected_type"].(string)
	required, hasRequired := row["required"].(bool)
	nullable, hasNullable := row["nullable_output"].(bool)
	ref := PropertySchemaReference{ProjectedKey: key, ProjectedType: typ, Required: required, NullableOutput: nullable}
	return ref, validPropertyKey(key) && validProjectedType(typ) && hasRequired && hasNullable
}

func validMissingPolicyV2(policy string) bool {
	return policy == "error" || policy == "omit" || policy == "default"
}

func (index resultRegistryIndexV2) validDeclarations(rows []any, keyField string, fields resultDeclarationsV2, references map[declarationKeyV2]PropertySchemaReference) bool {
	var previous declarationKeyV2
	for i, raw := range rows {
		row, _ := raw.(map[string]any)
		ref, ok := registryPropertyRefV2(row, keyField)
		if !ok || len(row) != 8 {
			return false
		}
		scope, _ := row["target_scope"].(string)
		kind, _ := row["target_kind"].(string)
		if !validIdentifierV2(kind) {
			return false
		}
		// An explicit kind without a configured producer retains its top-level
		// declaration. Only configured kinds can have nested references.
		var configured bool
		switch scope {
		case "vertex":
			_, configured = index.vertices[kind]
		case "edge":
			_, configured = index.edges[kind]
		case "graph_view":
			if keyField == "projected_key" {
				return false
			}
		default:
			return false
		}
		if scope != "graph_view" && !configured && kind == "*" {
			return false
		}
		key := declarationKeyV2{scope, kind, ref.ProjectedKey}
		missing, _ := row["missing_behavior"].(string)
		null, _ := row["source_null_behavior"].(string)
		if (i > 0 && !previous.less(key)) || !validMissingPolicyV2(missing) || !(validMissingPolicyV2(null) || (null == "emit_null" && ref.NullableOutput)) {
			return false
		}
		if keyField == "projected_metadata_key" && reservedSystemMetadataTerminal(key.key) {
			return false
		}
		if references != nil && configured {
			if references[key] != ref {
				return false
			}
			delete(references, key)
		}
		target := definitionTarget{scope, kind}
		set := fields[target]
		if set == nil {
			set = &resultFieldSetV2{fields: make(map[string]resultFieldV2)}
			fields[target] = set
		}
		mandatory := missing != "omit" && null != "omit"
		set.fields[key.key] = resultFieldV2{projectedType: ref.ProjectedType, canEmitNull: ref.NullableOutput && (missing == "default" || null == "default" || null == "emit_null"), mandatory: mandatory}
		if mandatory {
			set.mandatory++
		}
		previous = key
	}
	return true
}

func validRegistryDirectionsV2(values []any) bool {
	if len(values) == 0 {
		return false
	}
	position := 0
	for _, raw := range values {
		direction, _ := raw.(string)
		for position < len(registryDirectionsV2) && registryDirectionsV2[position] != direction {
			position++
		}
		if position == len(registryDirectionsV2) {
			return false
		}
		position++
	}
	return true
}

type resultKindProvenanceV2 struct {
	labels          map[string]bool
	preservesLabels bool
	directions      []any
	sourceKinds     map[string]bool
	rules           map[string]bool
}

func (index *resultRegistryIndexV2) indexKinds(rows []any, scope, sourcesField string, sourceLimit int, references map[declarationKeyV2]PropertySchemaReference) (map[string]resultKindProvenanceV2, bool) {
	result := make(map[string]resultKindProvenanceV2)
	sourceOutputs := make(map[string]int)
	maxKinds := 1
	if scope == "vertex" {
		index.entityOutputs = sourceOutputs
	}
	if scope == "edge" {
		// One relationship mapping can contribute a forward and a reverse kind.
		index.relationshipOutputs = sourceOutputs
		maxKinds = 2
	}
	previous := ""
	for _, raw := range rows {
		row, _ := raw.(map[string]any)
		kind, _ := row[scope+"_kind"].(string)
		if !validIdentifierV2(kind) || kind <= previous {
			return nil, false
		}
		var directions []any
		fieldCount := 6
		if scope == "edge" {
			fieldCount++
			var ok bool
			directions, ok = row["directions"].([]any)
			if !ok || !validRegistryDirectionsV2(directions) {
				return nil, false
			}
		}
		preserved, hasPreserved := row["source_labels_preserved"].(bool)
		labels, hasLabels := row["labels"].([]any)
		props, hasProps := row["properties"].([]any)
		if len(row) != fieldCount || !hasPreserved || !hasLabels || !hasProps {
			return nil, false
		}
		labelSet := make(map[string]bool, len(labels))
		last := ""
		for _, rawLabel := range labels {
			label, _ := rawLabel.(string)
			if !validLabelV2(label) || label <= last {
				return nil, false
			}
			labelSet[label] = true
			last = label
		}
		last = ""
		for _, rawProp := range props {
			prop, _ := rawProp.(map[string]any)
			ref, ok := registryPropertyRefV2(prop, "projected_key")
			if !ok || len(prop) != 4 || ref.ProjectedKey <= last {
				return nil, false
			}
			references[declarationKeyV2{scope, kind, ref.ProjectedKey}] = ref
			last = ref.ProjectedKey
		}
		sources, ok := orderedIdentifierSetV2(row[sourcesField], sourceLimit)
		if !ok {
			return nil, false
		}
		for source := range sources {
			sourceOutputs[source]++
			if sourceOutputs[source] > maxKinds || len(sourceOutputs) > sourceLimit {
				return nil, false
			}
		}
		rules, ok := orderedIdentifierSetV2(row["aggregation_rule_ids"], graphProjectionLimits.MaxAggregationRules-len(index.rules))
		if !ok || (len(sources) == 0 && (preserved || len(rules) == 0)) {
			return nil, false
		}
		for id := range rules {
			if index.rules[id] {
				return nil, false
			}
			index.rules[id] = true
		}
		result[kind] = resultKindProvenanceV2{sourceKinds: sources, rules: rules, directions: directions, labels: labelSet, preservesLabels: preserved}
		previous = kind
	}
	return result, true
}

func orderedIdentifierSetV2(value any, limit int) (map[string]bool, bool) {
	items, ok := value.([]any)
	if !ok || len(items) > limit {
		return nil, false
	}
	set := make(map[string]bool, len(items))
	previous := ""
	for _, raw := range items {
		id, _ := raw.(string)
		if !validIdentifierV2(id) || id <= previous {
			return nil, false
		}
		set[id] = true
		previous = id
	}
	return set, true
}

// All aggregation-only declarations and emitted aggregates expose the same
// defaults. Mixed declarations can also contain configured mapping labels.
func validateScopeLabelsV2(ctx context.Context, kinds map[string]resultKindProvenanceV2, rows []any, family string) error {
	var defaults map[string]bool
	for _, kind := range kinds {
		if len(kind.sourceKinds) != 0 {
			continue
		}
		if len(kind.labels) > graphProjectionLimits.MaxLabelsPerObject || (defaults != nil && !maps.Equal(defaults, kind.labels)) {
			return ErrResultV2Invalid
		}
		defaults = kind.labels
	}
	for _, raw := range rows {
		if err := ctx.Err(); err != nil {
			return err
		}
		row := raw.(map[string]any)
		if row[family] != "aggregated" {
			continue
		}
		labels := row["labels"].([]any)
		if defaults == nil {
			defaults = make(map[string]bool, len(labels))
			for _, label := range labels {
				defaults[label.(string)] = true
			}
		} else if len(labels) != len(defaults) || !containsScopeLabelsV2(labels, defaults) {
			return ErrResultV2Invalid
		}
	}
	if len(defaults) == 0 {
		return nil
	}
	for _, kind := range kinds {
		for label := range defaults {
			if !kind.labels[label] {
				return ErrResultV2Invalid
			}
		}
	}
	for _, raw := range rows {
		if err := ctx.Err(); err != nil {
			return err
		}
		if !containsScopeLabelsV2(raw.(map[string]any)["labels"].([]any), defaults) {
			return ErrResultV2Invalid
		}
	}
	return nil
}

// Row admission has already established label uniqueness.
func containsScopeLabelsV2(labels []any, defaults map[string]bool) bool {
	matched := 0
	for _, label := range labels {
		if defaults[label.(string)] {
			matched++
		}
	}
	return matched == len(defaults)
}
