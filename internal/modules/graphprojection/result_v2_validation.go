package graphprojection

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
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
	if result.Binding.ValidateIdentity() != nil || len(result.Vertices) > MaximumResultVerticesV2 || len(result.Edges) > MaximumResultEdgesV2 || !utf8.Valid(result.ResultJSON) {
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
	for _, key := range []string{"properties", "mapped_metadata", "schema_registry", "vertices", "edges", "validation_summary", "consumer_capabilities"} {
		kind := kindObject
		if key == "vertices" || key == "edges" {
			kind = kindArray
		}
		members[key] = memberSpec{kind: kind, required: true}
	}
	if validateClosedObjectV2(root, "$", members) != nil || !validResultPropertyMapV2(root["properties"]) || !validResultPropertyMapV2(root["mapped_metadata"]) {
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
	registry := root["schema_registry"].(map[string]any)
	if validateClosedObjectV2(registry, "$.schema_registry", map[string]memberSpec{
		"vertex_kinds": {kind: kindArray, required: true}, "edge_kinds": {kind: kindArray, required: true},
		"property_keys": {kind: kindArray, required: true}, "metadata_keys": {kind: kindArray, required: true},
	}) != nil {
		return ErrResultV2Invalid
	}
	vertexKinds, ok := resultKindSetV2(registry["vertex_kinds"].([]any), "vertex_kind")
	if !ok {
		return ErrResultV2Invalid
	}
	edgeKinds, ok := resultKindSetV2(registry["edge_kinds"].([]any), "edge_kind")
	if !ok {
		return ErrResultV2Invalid
	}
	vertices, edges := root["vertices"].([]any), root["edges"].([]any)
	if len(vertices) != len(result.Vertices) || len(edges) != len(result.Edges) {
		return ErrResultV2Invalid
	}
	vertexIDs := make(map[string]bool, len(vertices))
	for index, row := range result.Vertices {
		if err := ctx.Err(); err != nil {
			return err
		}
		object, ok := vertices[index].(map[string]any)
		if !ok || !validResultObjectV2(object, false) || !vertexKinds[row.VertexKind] || vertexIDs[row.VertexID] ||
			object["vertex_id"] != row.VertexID || object["vertex_kind"] != row.VertexKind || object["sort_key"] != row.SortKey || !sameResultBytesV2(object, row.JSON) {
			return ErrResultV2Invalid
		}
		if index > 0 && !resultOrderLessV2(result.Vertices[index-1].SortKey, result.Vertices[index-1].VertexID, row.SortKey, row.VertexID) {
			return ErrResultV2Invalid
		}
		vertexIDs[row.VertexID] = true
	}
	edgeObjects := make(map[string]map[string]any, len(edges))
	for index, row := range result.Edges {
		if err := ctx.Err(); err != nil {
			return err
		}
		object, ok := edges[index].(map[string]any)
		if !ok || !validResultObjectV2(object, true) || !edgeKinds[row.EdgeKind] || edgeObjects[row.EdgeID] != nil ||
			object["edge_id"] != row.EdgeID || object["edge_kind"] != row.EdgeKind || object["sort_key"] != row.SortKey ||
			object["src_vertex_id"] != row.SrcVertexID || object["dst_vertex_id"] != row.DstVertexID || object["direction"] != row.Direction ||
			!vertexIDs[row.SrcVertexID] || !vertexIDs[row.DstVertexID] || !sameResultBytesV2(object, row.JSON) {
			return ErrResultV2Invalid
		}
		if index > 0 && !resultOrderLessV2(result.Edges[index-1].SortKey, result.Edges[index-1].EdgeID, row.SortKey, row.EdgeID) {
			return ErrResultV2Invalid
		}
		edgeObjects[row.EdgeID] = object
	}
	for _, raw := range edges {
		if err := ctx.Err(); err != nil {
			return err
		}
		edge := raw.(map[string]any)
		meta := edge["metadata"].(map[string]any)
		reverse, ok := meta["is_reverse_edge"].(bool)
		link, hasLink := meta["reverse_of_edge_id"]
		if !ok || !hasLink || reverse != (edge["edge_family"] == "reverse") {
			return ErrResultV2Invalid
		}
		if !reverse {
			if link != nil {
				return ErrResultV2Invalid
			}
			continue
		}
		forwardID, _ := link.(string)
		if !validReverseLinkV2(edge, edgeObjects[forwardID]) {
			return ErrResultV2Invalid
		}
	}
	return ctx.Err()
}

func validReverseLinkV2(reverse, forward map[string]any) bool {
	if forward["edge_family"] != "direct" || forward["direction"] != "directed" || reverse["direction"] != "directed" ||
		reverse["src_vertex_id"] != forward["dst_vertex_id"] || reverse["dst_vertex_id"] != forward["src_vertex_id"] {
		return false
	}
	left, _ := reverse["source_relationship_ref"].(map[string]any)
	right, _ := forward["source_relationship_ref"].(map[string]any)
	reverseMapping, _ := reverse["metadata"].(map[string]any)["mapping_rule_id"].(string)
	forwardMapping, _ := forward["metadata"].(map[string]any)["mapping_rule_id"].(string)
	return left != nil && right != nil && reverseMapping != "" && reverseMapping == forwardMapping &&
		left["mapping_rule_id"] == reverseMapping && right["mapping_rule_id"] == forwardMapping &&
		left["source_relationship_kind"] == right["source_relationship_kind"] && left["source_relationship_id"] == right["source_relationship_id"]
}

func canonicalOutputDigestV2(output map[string]any) (string, error) {
	encoded, err := canonicalJSON(output)
	if err != nil {
		return "", err
	}
	return sha256Hex(encoded), nil
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

func resultKindSetV2(rows []any, key string) (map[string]bool, bool) {
	result := make(map[string]bool, len(rows))
	previous := ""
	for _, raw := range rows {
		row, ok := raw.(map[string]any)
		if !ok {
			return nil, false
		}
		id, ok := row[key].(string)
		if !ok || !validIdentifierV2(id) || id <= previous {
			return nil, false
		}
		result[id], previous = true, id
	}
	return result, true
}

func validResultPropertyMapV2(value any) bool {
	object, ok := value.(map[string]any)
	if !ok || len(object) > graphProjectionLimits.MaxPropertyKeys {
		return false
	}
	for _, value := range object {
		if array, ok := value.([]any); ok && len(array) > graphProjectionLimits.MaxArrayItems {
			return false
		}
	}
	return validatePropertyMapV2(object, "$") == nil
}

func validResultObjectV2(object map[string]any, edge bool) bool {
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
	if !validResultPropertyMapV2(object["metadata"].(map[string]any)["mapped_metadata"]) {
		return false
	}
	id := object[kind+"_id"].(string)
	prefix := "vx_"
	if edge {
		prefix = "ed_"
	}
	if !strings.HasPrefix(id, prefix) || !validResultSHA256V2(strings.TrimPrefix(id, prefix)) || !validIdentifierV2(object[kind+"_kind"].(string)) ||
		(object[kind+"_family"] != "direct" && object[kind+"_family"] != "aggregated" && !(edge && object[kind+"_family"] == "reverse")) || !validResultPropertyMapV2(object["properties"]) {
		return false
	}
	if edge && object["direction"] != "directed" && object["direction"] != "undirected" && object["direction"] != "bidirectional" {
		return false
	}
	labels := object["labels"].([]any)
	if len(labels) > graphProjectionLimits.MaxLabelsPerObject {
		return false
	}
	seen := make(map[string]bool, len(labels))
	for _, value := range labels {
		label, ok := value.(string)
		if !ok || len(label) > graphProjectionLimits.MaxLabelBytes || seen[label] {
			return false
		}
		seen[label] = true
	}
	if raw := object[sourceRef]; raw != nil {
		ref := raw.(map[string]any)
		prefix := "source_entity"
		if edge {
			prefix = "source_relationship"
		}
		if len(ref) != 3 {
			return false
		}
		for _, key := range []string{prefix + "_id", prefix + "_kind", "mapping_rule_id"} {
			value, ok := ref[key].(string)
			if !ok || !validIdentifierV2(value) {
				return false
			}
		}
	}
	return true
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
