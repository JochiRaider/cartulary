package networkflow

import (
	"context"
	"errors"
	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/graphprojection"
	"github.com/JochiRaider/cartulary/internal/modules/reporting/graphsourcecontract"
)

func checkReportingSourceBudget(t *testing.T) {
	// Fixed schema-shaped vector: UTF-8 byte accounting, no HTML escaping, null
	// unused scalar, the array comma, and exactly one document LF are observable.
	binding := graphprojection.ResultBindingV2{SourceOwnerID: "owner", GraphViewID: "view", ProjectionResultID: "result", SourceSnapshotID: "snapshot", ProjectionSchemaID: "graph_projection.v2", ProjectionVersion: "v1", NormalizedConfigurationSHA256: "config", NormalizedSourceSHA256: "source", CanonicalOutputSHA256: "output"}
	candidate := graphsourcecontract.VertexLabelCandidate{ProjectedVertexID: "vertex", Endpoint: graphsourcecontract.LabelComponent{ComponentKind: "endpoint_value", FieldPath: "/graph/vertices/10000/endpoint_value", SourceObjectRef: "é<&>", Classification: "derived_analytic", DisclosurePartitions: []string{"internal_only"}, RedactionBehavior: "allow_internal_only", State: "present", ValueKind: "string", StringValue: "192.0.2.1"}}
	const envelope = `{"edge_label_candidates":[],"schema_id":"cartulary.network_flow.reporting_graph_source.v3","source_projection_ref":{"canonical_output_sha256":"output","graph_view_id":"view","normalized_configuration_sha256":"config","normalized_source_sha256":"source","projection_result_id":"result","projection_schema_id":"graph_projection.v2","projection_version":"v1","source_owner_id":"owner","source_snapshot_id":"snapshot"},"vertex_label_candidates":[]}` + "\n"
	const element = `{"endpoint":{"classification":"derived_analytic","component_kind":"endpoint_value","disclosure_partition_refs":["internal_only"],"field_path":"/graph/vertices/10000/endpoint_value","integer_value":null,"redaction_behavior":"allow_internal_only","source_object_ref":"é<&>","state":"present","string_value":"192.0.2.1","value_kind":"string"},"projected_vertex_id":"vertex"}`
	resource := decodeNetworkFlowContractArtifact(t, "contracts/network-flow/reporting-graph-source.v3.schema.json")
	compiler := jsonschema.NewCompiler()
	compiler.UseLoader(jsonschema.SchemeURLLoader{})
	const sourceURL = "https://network-flow.invalid/source"
	if err := compiler.AddResource(sourceURL, resource); err != nil {
		t.Fatal(err)
	}
	componentSchema, err := compiler.Compile(sourceURL + "#/$defs/LabelComponentV1")
	if err != nil {
		t.Fatal(err)
	}
	component := reportingLabelComponent(candidate.Endpoint)
	// JSON-schema validation consumes JSON values, rather than native []string.
	component["disclosure_partition_refs"] = []any{"internal_only"}
	if err := componentSchema.Validate(component); err != nil {
		t.Fatal("admitted five-digit path rejected", err)
	}
	component["field_path"] = "/graph/edges/250000/bucket_end_utc"
	if err := componentSchema.Validate(component); err != nil {
		t.Fatal("admitted six-digit path rejected", err)
	}
	expected := int64(len(envelope) + 2*len(element) + 1)
	b, err := newReportingLabelBudget(binding, expected)
	if err != nil || b.used != int64(len(envelope)) {
		t.Fatalf("empty vector size: %d %v", b.used, err)
	}
	if err = b.addVertex(candidate, false); err != nil {
		t.Fatal(err)
	}
	if err = b.addVertex(candidate, true); err != nil || b.used != expected {
		t.Fatalf("exact byte ceiling: %d/%d %v", b.used, expected, err)
	}
	before := b.used
	if err = b.addVertex(candidate, true); !errors.Is(err, graphprojection.ErrResultV2Invalid) || b.used != before {
		t.Fatal("overflow accepted or advanced budget")
	}
	b, err = newReportingLabelBudget(binding, expected-1)
	if err != nil {
		t.Fatal(err)
	}
	if err = b.addVertex(candidate, false); err != nil {
		t.Fatal(err)
	}
	if err = b.addVertex(candidate, true); !errors.Is(err, graphprojection.ErrResultV2Invalid) {
		t.Fatal("one-byte overflow accepted")
	}
	if _, err = newReportingLabelBudget(binding, int64(len(envelope)-1)); err == nil {
		t.Fatal("empty envelope omitted final LF")
	}
	missing := reportingLabelComponent(graphsourcecontract.LabelComponent{State: "missing", ValueKind: "integer", IntegerValue: 42, StringValue: "must not encode"})
	if missing["string_value"] != nil || missing["integer_value"] != nil {
		t.Fatal("missing component exposed a scalar")
	}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := reportingGraphResult(ctx, graphprojection.CompletedResultV2{Vertices: []graphprojection.ResultVertexV2{{}}}); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancellation lost: %v", err)
	}
	if _, err := reportingGraphResult(context.Background(), graphprojection.CompletedResultV2{Vertices: make([]graphprojection.ResultVertexV2, graphprojection.MaximumResultVerticesV2+1)}); !errors.Is(err, graphprojection.ErrResultV2Invalid) {
		t.Fatal("vertex bound not enforced before elements")
	}
	if _, err := reportingGraphResult(context.Background(), graphprojection.CompletedResultV2{Edges: make([]graphprojection.ResultEdgeV2, graphprojection.MaximumResultEdgesV2+1)}); !errors.Is(err, graphprojection.ErrResultV2Invalid) {
		t.Fatal("edge bound not enforced before elements")
	}
}
