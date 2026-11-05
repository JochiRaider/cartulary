package networkflow

import (
	"github.com/JochiRaider/cartulary/internal/modules/graphprojection"
	"github.com/JochiRaider/cartulary/internal/modules/reporting/graphsourcecontract"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

// Accounting projects the logical source schema, without changing the native
// structs serialized into Reporting's independently owned frozen input.
type reportingLabelBudget struct{ used, maximum int64 }

func newReportingLabelBudget(binding graphprojection.ResultBindingV2, maximum int64) (reportingLabelBudget, error) {
	b := reportingLabelBudget{maximum: maximum}
	ref := map[string]any{
		"source_owner_id": binding.SourceOwnerID, "graph_view_id": binding.GraphViewID,
		"projection_result_id": binding.ProjectionResultID, "source_snapshot_id": binding.SourceSnapshotID,
		"projection_schema_id": binding.ProjectionSchemaID, "projection_version": binding.ProjectionVersion,
		"normalized_configuration_sha256": binding.NormalizedConfigurationSHA256,
		"normalized_source_sha256":        binding.NormalizedSourceSHA256, "canonical_output_sha256": binding.CanonicalOutputSHA256,
	}
	encoded, err := canonicaljson.MarshalExtension(map[string]any{"schema_id": graphsourcecontract.SchemaID, "source_projection_ref": ref, "vertex_label_candidates": []any{}, "edge_label_candidates": []any{}})
	if err != nil || int64(len(encoded)) > maximum {
		return b, graphprojection.ErrResultV2Invalid
	}
	b.used = int64(len(encoded)) // Includes the sole final LF and both array brackets.
	return b, nil
}

func (b *reportingLabelBudget) add(value any, comma bool) error {
	encoded, err := canonicaljson.MarshalExtension(value)
	if err != nil {
		return graphprojection.ErrResultV2Invalid
	}
	size := int64(len(encoded) - 1) // Elements have no LF inside the enclosing arrays.
	if comma {
		size++
	}
	if size > b.maximum-b.used {
		return graphprojection.ErrResultV2Invalid
	}
	b.used += size
	return nil
}

func (b *reportingLabelBudget) addVertex(candidate graphsourcecontract.VertexLabelCandidate, comma bool) error {
	return b.add(map[string]any{"projected_vertex_id": candidate.ProjectedVertexID, "endpoint": reportingLabelComponent(candidate.Endpoint)}, comma)
}
func (b *reportingLabelBudget) addEdge(candidate graphsourcecontract.EdgeLabelCandidate, comma bool) error {
	var start, end any
	if candidate.BucketStartUTC != nil {
		start = reportingLabelComponent(*candidate.BucketStartUTC)
	}
	if candidate.BucketEndUTC != nil {
		end = reportingLabelComponent(*candidate.BucketEndUTC)
	}
	return b.add(map[string]any{"kind": candidate.Kind, "projected_edge_id": candidate.ProjectedEdgeID, "protocol": reportingLabelComponent(candidate.Protocol), "direction_kind": candidate.DirectionKind, "to_port": optionalReportingLabelComponent(candidate.ToPort), "endpoint_a_port": optionalReportingLabelComponent(candidate.EndpointAPort), "endpoint_b_port": optionalReportingLabelComponent(candidate.EndpointBPort), "bucket_start_utc": start, "bucket_end_utc": end}, comma)
}
func reportingLabelComponent(component graphsourcecontract.LabelComponent) map[string]any {
	var text, number any
	if component.State == graphsourcecontract.ComponentStatePresent {
		if component.ValueKind == graphsourcecontract.ValueKindString {
			text = component.StringValue
		}
		if component.ValueKind == graphsourcecontract.ValueKindInteger {
			number = component.IntegerValue
		}
	}
	return map[string]any{"component_kind": component.ComponentKind, "field_path": component.FieldPath, "source_object_ref": component.SourceObjectRef, "classification": component.Classification, "disclosure_partition_refs": component.DisclosurePartitions, "redaction_behavior": component.RedactionBehavior, "state": component.State, "value_kind": component.ValueKind, "string_value": text, "integer_value": number}
}

func optionalReportingLabelComponent(c *graphsourcecontract.LabelComponent) any {
	if c == nil {
		return nil
	}
	return reportingLabelComponent(*c)
}
