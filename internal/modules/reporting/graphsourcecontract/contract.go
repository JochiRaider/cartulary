// Package graphsourcecontract defines the narrow typed handoff from a graph
// source owner to Reporting. It deliberately contains no source-owner parsing,
// redaction policy, rendering behavior, or persistence mechanics.
package graphsourcecontract

import "github.com/JochiRaider/cartulary/internal/modules/graphprojection"

const (
	SchemaID = "cartulary.network_flow.reporting_graph_source.v2"

	ComponentStatePresent = "present"
	ComponentStateAbsent  = "absent"
	ComponentStateRemoved = "removed"
	ComponentStateMissing = "missing"

	ValueKindString  = "string"
	ValueKindInteger = "integer"

	ClassificationDerivedAnalytic = "derived_analytic"
	DisclosurePartitionInternal   = "internal_only"
	RedactionInternalOnly         = "allow_internal_only"
)

type Result struct {
	Projection      graphprojection.CompletedResultV2
	LabelCandidates LabelCandidates
}

type LabelCandidates struct {
	SchemaID              string
	SourceProjectionRef   graphprojection.ResultBindingV2
	VertexLabelCandidates []VertexLabelCandidate
	EdgeLabelCandidates   []EdgeLabelCandidate
}

type VertexLabelCandidate struct {
	ProjectedVertexID string
	Endpoint          LabelComponent
}

type EdgeLabelCandidate struct {
	Kind            string
	ProjectedEdgeID string
	Protocol        LabelComponent
	DestinationPort LabelComponent
	BucketStartUTC  *LabelComponent
	BucketEndUTC    *LabelComponent
}

type LabelComponent struct {
	ComponentKind        string
	FieldPath            string
	SourceObjectRef      string
	Classification       string
	DisclosurePartitions []string
	RedactionBehavior    string
	State                string
	ValueKind            string
	StringValue          string
	IntegerValue         int64
}
