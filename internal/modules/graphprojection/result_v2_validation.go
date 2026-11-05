package graphprojection

import (
	"bufio"
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"io"
	"maps"
	"slices"
	"strings"
	"unicode/utf8"
)

// Validate proves the current semantic identity and canonical envelope/object
// correspondence. Publication time and transaction ownership belong to storage.
// It does not rerun projection or invoke a source owner.
func (result CompletedResultV2) Validate(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if len(result.ResultJSON) > graphProjectionLimits.MaxResultBytes || result.Binding.ValidateIdentity() != nil || len(result.Vertices) > MaximumResultVerticesV2 || len(result.Edges) > MaximumResultEdgesV2 || !utf8.Valid(result.ResultJSON) {
		return ErrResultV2Invalid
	}
	decoder := json.NewDecoder(bytes.NewReader(result.ResultJSON))
	decoder.UseNumber()
	var root map[string]any
	if decoder.Decode(&root) != nil {
		return ErrResultV2Invalid
	}
	if _, err := decoder.Token(); err != io.EOF {
		return ErrResultV2Invalid
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	identity := map[string]string{
		"projection_schema_id": result.Binding.ProjectionSchemaID, "projection_result_id": result.Binding.ProjectionResultID,
		"graph_view_id": result.Binding.GraphViewID, "source_owner_id": result.Binding.SourceOwnerID,
		"source_snapshot_id": result.Binding.SourceSnapshotID, "projection_version": result.Binding.ProjectionVersion,
		"normalized_configuration_sha256": result.Binding.NormalizedConfigurationSHA256,
		"normalized_source_sha256":        result.Binding.NormalizedSourceSHA256, "canonical_output_sha256": result.Binding.CanonicalOutputSHA256,
	}
	members := map[string]memberSpec{}
	for key, expected := range identity {
		members[key] = memberSpec{kind: kindString, required: true}
		if root[key] != expected {
			return ErrResultV2Invalid
		}
	}
	output := map[string]any{}
	for _, key := range []string{"mapped_metadata", "schema_registry", "vertices", "edges", "validation_summary", "consumer_capabilities"} {
		kind := kindObject
		if key == "vertices" || key == "edges" {
			kind = kindArray
		}
		members[key] = memberSpec{kind: kind, required: true}
	}
	if validateClosedObjectV2(root, "$", members) != nil {
		return ErrResultV2Invalid
	}
	registry := root["schema_registry"].(map[string]any)
	if validateClosedObjectV2(registry, "$.schema_registry", map[string]memberSpec{
		"vertex_kinds": {kind: kindArray, required: true}, "edge_kinds": {kind: kindArray, required: true},
		"property_keys": {kind: kindArray, required: true}, "metadata_keys": {kind: kindArray, required: true},
	}) != nil {
		return ErrResultV2Invalid
	}
	indexRegistry, ok := indexResultRegistryV2(registry)
	if !ok || !indexRegistry.metadata.admits("graph_view", "*", root["mapped_metadata"]) {
		return ErrResultV2Invalid
	}
	// Re-encoding with the producer's actual canonical algorithm also rejects
	// duplicate members, replacement-decoded Unicode, and noncanonical numbers.
	normalized := normalizeCanonicalValueV2(root).(map[string]any)
	canonical, err := canonicalJSON(normalized)
	if err != nil || !bytes.Equal(canonical, result.ResultJSON) {
		return ErrResultV2Invalid
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	for key := range members {
		if _, isIdentity := identity[key]; !isIdentity {
			output[key] = normalized[key]
		}
	}
	digest, err := canonicalOutputDigestV2(output)
	if err != nil || digest != result.Binding.CanonicalOutputSHA256 {
		return ErrResultV2Invalid
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if !sameResultValueV2(root["validation_summary"], validationSummaryResource(ValidationSummary{Status: "passed"})) ||
		!sameResultValueV2(root["consumer_capabilities"], consumerCapabilitiesResource(defaultConsumerCapabilities())) {
		return ErrResultV2Invalid
	}
	vertices, edges := root["vertices"].([]any), root["edges"].([]any)
	if len(vertices) != len(result.Vertices) || len(edges) != len(result.Edges) {
		return ErrResultV2Invalid
	}
	vertexObjects := make(map[string]map[string]any, len(vertices))
	// Registry admission has already fixed one output kind for each source kind.
	sources := map[string]map[string]string{"source_entity": {}, "source_relationship": {}}
	entityKinds := sources["source_entity"]
	mappingByKind := make(map[string]string)
	kindByMapping := make(map[string]string)
	for index, row := range result.Vertices {
		if err := ctx.Err(); err != nil {
			return err
		}
		object, ok := vertices[index].(map[string]any)
		if !ok || !validResultObjectV2(object, false, indexRegistry, indexRegistry.vertices[row.VertexKind]) || vertexObjects[row.VertexID] != nil ||
			object["vertex_id"] != row.VertexID || object["vertex_kind"] != row.VertexKind || object["sort_key"] != row.SortKey || !sameResultBytesV2(object, row.JSON) {
			return ErrResultV2Invalid
		}
		if object["vertex_family"] == "direct" {
			ref := object["source_entity_ref"].(map[string]any)
			kind := ref["source_entity_kind"].(string)
			rule := ref["mapping_rule_id"].(string)
			source := ref["source_entity_id"].(string)
			if len(entityKinds) >= graphProjectionLimits.MaxSourceEntities || entityKinds[source] != "" || (mappingByKind[kind] != "" && mappingByKind[kind] != rule) || (kindByMapping[rule] != "" && kindByMapping[rule] != kind) {
				return ErrResultV2Invalid
			}
			entityKinds[source] = kind
			mappingByKind[kind] = rule
			kindByMapping[rule] = kind
			mapping := entityMappingDigestV2(rule, kind, row.VertexKind)
			id, key := directVertexIdentityV2(result.Binding.GraphViewID, row.VertexKind, kind, source, mapping)
			if id != row.VertexID || key != row.SortKey {
				return ErrResultV2Invalid
			}
		}
		if index > 0 && !resultOrderLessV2(result.Vertices[index-1].SortKey, result.Vertices[index-1].VertexID, row.SortKey, row.VertexID) {
			return ErrResultV2Invalid
		}
		vertexObjects[row.VertexID] = object
	}
	edgeObjects := make(map[string]map[string]any, len(edges))
	for index, row := range result.Edges {
		if err := ctx.Err(); err != nil {
			return err
		}
		object, ok := edges[index].(map[string]any)
		if !ok || !validResultObjectV2(object, true, indexRegistry, indexRegistry.edges[row.EdgeKind]) || edgeObjects[row.EdgeID] != nil ||
			object["edge_id"] != row.EdgeID || object["edge_kind"] != row.EdgeKind || object["sort_key"] != row.SortKey ||
			object["src_vertex_id"] != row.SrcVertexID || object["dst_vertex_id"] != row.DstVertexID || object["direction"] != row.Direction ||
			!sameResultBytesV2(object, row.JSON) {
			return ErrResultV2Invalid
		}
		endpointFamily := "direct"
		if object["edge_family"] == "aggregated" {
			endpointFamily = "aggregated"
		}
		if vertexObjects[row.SrcVertexID]["vertex_family"] != endpointFamily || vertexObjects[row.DstVertexID]["vertex_family"] != endpointFamily {
			return ErrResultV2Invalid
		}
		if index > 0 && !resultOrderLessV2(result.Edges[index-1].SortKey, result.Edges[index-1].EdgeID, row.SortKey, row.EdgeID) {
			return ErrResultV2Invalid
		}
		edgeObjects[row.EdgeID] = object
	}
	if err := validateScopeLabelsV2(ctx, indexRegistry.vertices, vertices, "vertex_family"); err != nil {
		return err
	}
	if err := validateScopeLabelsV2(ctx, indexRegistry.edges, edges, "edge_family"); err != nil {
		return err
	}
	if err := validateRelationshipClaimsV2(ctx, result.Binding.GraphViewID, edgeObjects, kindByMapping, sources["source_relationship"], indexRegistry); err != nil {
		return err
	}
	// Index groups before checking any producer so later rows are available.
	groups := make(map[string][]map[string]any)
	for _, rows := range [][]any{vertices, edges} {
		for _, raw := range rows {
			if err := ctx.Err(); err != nil {
				return err
			}
			object := raw.(map[string]any)
			if rule, ok := object["metadata"].(map[string]any)["aggregation_rule_id"].(string); ok {
				groups[rule] = append(groups[rule], object)
			}
		}
	}
	dependencies := make(ruleDependenciesV2, len(indexRegistry.rules))
	for rule := range indexRegistry.rules {
		dependencies[rule] = nil
	}
	for producer, rows := range groups {
		observed := ruleGroupProofV2{contributors: make(map[string]struct{})}
		dependencies[producer] = make(map[string]struct{})
		for _, object := range rows {
			if err := observed.validateContributors(ctx, object, vertexObjects, edgeObjects, sources, indexRegistry, dependencies[producer]); err != nil {
				return err
			}
			if object["edge_family"] == "aggregated" {
				var pair [2]string
				for i, field := range []string{"src_vertex_id", "dst_vertex_id"} {
					endpoint := vertexObjects[object[field].(string)]
					rule := endpoint["metadata"].(map[string]any)["aggregation_rule_id"].(string)
					pair[i] = rule
					dependencies[producer][rule] = struct{}{}
				}
				if observed.endpoints[0] != "" && observed.endpoints != pair {
					return ErrResultV2Invalid
				}
				observed.endpoints = pair
			}
		}
		var declarations map[string]bool
		switch observed.scope {
		case "projected_vertex":
			declarations = indexRegistry.vertices[observed.kind].rules
		case "projected_edge":
			declarations = indexRegistry.edges[observed.kind].rules
		}
		addOtherProducersV2(dependencies[producer], producer, maps.Keys(declarations))
	}

	if len(dependencies.orderedIDs()) != len(dependencies) {
		return ErrResultV2Invalid
	}
	return ctx.Err()
}

func canonicalOutputDigestV2(output map[string]any) (string, error) {
	digest := sha256.New()
	out := bufio.NewWriter(digest)
	if err := writeCanonicalJSON(out, output); err != nil {
		return "", err
	}
	if err := out.Flush(); err != nil {
		return "", err
	}
	return hex.EncodeToString(digest.Sum(nil)), nil
}

func sameResultBytesV2(value any, raw []byte) bool {
	encoded, err := canonicalJSON(normalizeCanonicalValueV2(value))
	return err == nil && bytes.Equal(encoded, raw)
}

func sameResultValueV2(left, right any) bool {
	encoded, err := canonicalJSON(right)
	return err == nil && sameResultBytesV2(left, encoded)
}

func resultOrderLessV2(leftKey, leftID, rightKey, rightID string) bool {
	return leftKey < rightKey || leftKey == rightKey && leftID < rightID
}

func validResultObjectV2(object map[string]any, edge bool, registry resultRegistryIndexV2, declaration resultKindProvenanceV2) bool {
	kind, sourceRef := "vertex", "source_entity_ref"
	if edge {
		kind, sourceRef = "edge", "source_relationship_ref"
	}
	members := map[string]memberSpec{
		kind + "_id": {kind: kindString, required: true}, kind + "_kind": {kind: kindString, required: true},
		kind + "_family": {kind: kindString, required: true}, "sort_key": {kind: kindString, required: true},
		"labels": {kind: kindArray, required: true}, "properties": {kind: kindObject, required: true},
		"metadata": {kind: kindObject, required: true}, sourceRef: {kind: kindObject, required: true, nullable: true},
	}
	if edge {
		for _, key := range []string{"src_vertex_id", "dst_vertex_id", "direction"} {
			members[key] = memberSpec{kind: kindString, required: true}
		}
	}
	if validateClosedObjectV2(object, "$", members) != nil {
		return false
	}
	sortDigest, hasPrefix := strings.CutPrefix(object["sort_key"].(string), "sk_")
	if !hasPrefix || !validResultSHA256V2(sortDigest) {
		return false
	}
	outputKind := object[kind+"_kind"].(string)
	if !registry.properties.admits(kind, outputKind, object["properties"]) || !registry.metadata.admits(kind, outputKind, object["metadata"].(map[string]any)["mapped_metadata"]) {
		return false
	}
	id := object[kind+"_id"].(string)
	prefix := "vx_"
	if edge {
		prefix = "ed_"
	}
	if !strings.HasPrefix(id, prefix) || !validResultSHA256V2(strings.TrimPrefix(id, prefix)) || !validIdentifierV2(object[kind+"_kind"].(string)) ||
		(object[kind+"_family"] != "direct" && object[kind+"_family"] != "aggregated" && !(edge && object[kind+"_family"] == "reverse")) {
		return false
	}
	if edge && !slices.Contains(declaration.directions, object["direction"]) {
		return false
	}
	labels := object["labels"].([]any)
	if len(labels) > graphProjectionLimits.MaxLabelsPerObject {
		return false
	}
	allowSourceLabels := declaration.preservesLabels && object[kind+"_family"] != "aggregated"
	previous := ""
	for _, value := range labels {
		label, ok := value.(string)
		if !ok || !validLabelV2(label) || label <= previous || (!allowSourceLabels && !declaration.labels[label]) {
			return false
		}
		previous = label
	}
	meta := object["metadata"].(map[string]any)
	mapping, hasMapping := meta["mapping_rule_id"]
	aggregation, hasAggregation := meta["aggregation_rule_id"]
	refs, hasRefs := meta["aggregation_source_refs"].([]any)
	metadataFields := 4
	if edge {
		metadataFields += 2 // Reverse flag and forward edge ID are proved after indexing.
	}
	if !hasMapping || !hasAggregation || !hasRefs || len(meta) != metadataFields {
		return false
	}
	if object[kind+"_family"] == "aggregated" {
		rule, ok := aggregation.(string)
		return object[sourceRef] == nil && mapping == nil && ok && declaration.rules[rule] && len(refs) > 0
	}
	ref, ok := object[sourceRef].(map[string]any)
	rule, okRule := mapping.(string)
	if !ok || len(ref) != 3 || !okRule || !validIdentifierV2(rule) || ref["mapping_rule_id"] != rule || aggregation != nil || len(refs) != 0 {
		return false
	}
	sourcePrefix := strings.TrimSuffix(sourceRef, "_ref")
	sourceID, _ := ref[sourcePrefix+"_id"].(string)
	sourceKind, _ := ref[sourcePrefix+"_kind"].(string)
	return validIdentifierV2(sourceID) && declaration.sourceKinds[sourceKind]
}

type ruleGroupProofV2 struct {
	scope, kind  string
	endpoints    [2]string
	contributors map[string]struct{}
}

func (proof *ruleGroupProofV2) validateContributors(ctx context.Context, object map[string]any, vertices, edges map[string]map[string]any, sources map[string]map[string]string, registry resultRegistryIndexV2, prerequisites map[string]struct{}) error {
	refs := object["metadata"].(map[string]any)["aggregation_source_refs"].([]any)
	previous := ""
	for _, raw := range refs {
		if err := ctx.Err(); err != nil {
			return err
		}
		ref, ok := raw.(map[string]any)
		if !ok || len(ref) != 4 {
			return ErrResultV2Invalid
		}
		scope, _ := ref["ref_kind"].(string)
		id, _ := ref["ref_id"].(string)
		kind, _ := ref["ref_kind_name"].(string)
		if !validIdentifierV2(kind) || !validIdentifierV2(id) || (proof.scope != "" && (scope != proof.scope || kind != proof.kind)) {
			return ErrResultV2Invalid
		}
		if _, seen := proof.contributors[id]; seen {
			return ErrResultV2Invalid
		}
		var order contributorOrder
		var limit int
		switch scope {
		case "source_entity", "source_relationship":
			kinds := registry.entityOutputs
			limit = graphProjectionLimits.MaxSourceEntities
			if scope == "source_relationship" {
				kinds = registry.relationshipOutputs
				limit = graphProjectionLimits.MaxSourceRelationships
			}
			if kinds[kind] == 0 {
				return ErrResultV2Invalid
			}
			claims := sources[scope]
			if prior, seen := claims[id]; seen {
				if prior != kind {
					return ErrResultV2Invalid
				}
			} else {
				if len(claims) >= limit {
					return ErrResultV2Invalid
				}
				claims[id] = kind
			}
			order = contributorOrder{kind, id, "", ""}
		case "projected_vertex", "projected_edge":
			limit = MaximumResultVerticesV2
			related, kindField, source := vertices[id], "vertex_kind", "source_entity"
			if scope == "projected_edge" {
				limit = MaximumResultEdgesV2
				related, kindField, source = edges[id], "edge_kind", "source_relationship"
			}
			if related == nil || related[kindField] != kind {
				return ErrResultV2Invalid
			}
			if direct, ok := related[source+"_ref"].(map[string]any); ok {
				order = contributorOrder{direct[source+"_kind"].(string), direct[source+"_id"].(string), direct["mapping_rule_id"].(string), id}
			} else {
				rule := related["metadata"].(map[string]any)["aggregation_rule_id"].(string)
				prerequisites[rule] = struct{}{}
				order = contributorOrder{kind, id, rule, id}
			}
		default:
			return ErrResultV2Invalid
		}
		key := order.key()
		if ref["contributor_sort_key"] != key || key <= previous {
			return ErrResultV2Invalid
		}
		if len(proof.contributors) >= limit {
			return ErrResultV2Invalid
		}
		proof.contributors[id] = struct{}{}
		previous, proof.scope, proof.kind = key, scope, kind
	}
	return nil
}

func validResultSHA256V2(value string) bool {
	if len(value) != 64 {
		return false
	}
	for _, character := range value {
		if !(character >= '0' && character <= '9' || character >= 'a' && character <= 'f') {
			return false
		}
	}
	return true
}
