package networkflow

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/graphprojection"
	"github.com/JochiRaider/cartulary/internal/modules/graphprojection/postgresresult"
	"github.com/JochiRaider/cartulary/internal/modules/reporting/graphsourcecontract"
)

const (
	ReportingGraphSourceContributionID            = "network_flow_activity.reporting_graph_source"
	ReportingGraphSourceParticipantID             = "network_flow_activity.reporting_graph_source_v2"
	ReportingGraphSourceReadAlgorithmID           = "network_flow_activity.reporting_graph_source_v2"
	ReportingGraphSourceOrderingAlgorithmID       = "network_flow_activity.reporting_graph_order_v1"
	ReportingGraphSourceStateFamilyID             = "network_flow_activity.graph_views"
	ReportingGraphSourceMaxLabelBytes       int64 = 67_108_864
)

// ReportingGraphSource is Network Flow's typed immutable-result contribution.
// It owns declaration interpretation while Graph Projection owns exact result
// admission locking.
type ReportingGraphSource struct {
	store           reportingGraphDeclarationReader
	graphStateReady func() bool
}

type reportingGraphDeclarationReader interface {
	GetGraphViewDeclarationTx(context.Context, pgx.Tx, uuid.UUID, string, bool) (graphViewDeclaration, error)
}

func newReportingGraphSource(declarations reportingGraphDeclarationReader, ready func() bool) (*ReportingGraphSource, error) {
	if declarations == nil || ready == nil {
		return nil, errors.New("network flow Reporting graph source requires persistence")
	}
	return &ReportingGraphSource{store: declarations, graphStateReady: ready}, nil
}

func (m *Module) ReportingGraphSource() *ReportingGraphSource {
	if m == nil {
		return nil
	}
	return m.reportingSource
}

func (*ReportingGraphSource) SourceOwnerID() string { return ProfileID }

func (source *ReportingGraphSource) GraphStateReady() bool {
	return source != nil && source.graphStateReady != nil && source.graphStateReady()
}

func (source *ReportingGraphSource) ReadSelectedResultTx(ctx context.Context, tx pgx.Tx, incidentID uuid.UUID, binding graphprojection.ResultBindingV2) (graphsourcecontract.Result, error) {
	if source == nil || source.store == nil || !source.GraphStateReady() || tx == nil || binding.SourceOwnerID != ProfileID || binding.ProjectionSchemaID != graphprojection.ProjectionSchemaIDV2 || incidentID == uuid.Nil || !graphViewIDPattern.MatchString(binding.GraphViewID) || !sourceSnapshotIDPattern.MatchString(binding.SourceSnapshotID) || !graphViewSHA256Pattern.MatchString(binding.NormalizedConfigurationSHA256) || !graphViewSHA256Pattern.MatchString(binding.NormalizedSourceSHA256) || !graphViewSHA256Pattern.MatchString(binding.CanonicalOutputSHA256) || (binding.ProjectionVersion != "network_flow_activity.v2" && binding.ProjectionVersion != "network_flow_activity.time_bucket.v2") || binding.ValidateIdentity() != nil {
		return graphsourcecontract.Result{}, graphprojection.ErrResultV2BindingMismatch
	}
	if err := ctx.Err(); err != nil {
		return graphsourcecontract.Result{}, err
	}
	reader, err := postgresresult.NewReader(tx)
	if err != nil {
		return graphsourcecontract.Result{}, err
	}
	storedBinding, err := reader.LockResultEnvelope(ctx, binding.ProjectionResultID)
	if err != nil {
		return graphsourcecontract.Result{}, err
	}
	declaration, err := source.store.GetGraphViewDeclarationTx(ctx, tx, incidentID, binding.GraphViewID, true)
	if errors.Is(err, errGraphViewDeclarationNotFound) {
		return graphsourcecontract.Result{}, graphprojection.ErrResultV2NotFound
	}
	if err != nil {
		return graphsourcecontract.Result{}, err
	}
	if declaration.DeclarationState != graphViewDeclarationStateActive || declaration.SelectedResult == nil {
		return graphsourcecontract.Result{}, graphsourcecontract.ErrNotSelected
	}
	selected := graphViewResultBinding(declaration)
	if selected.SourceSnapshotID != binding.SourceSnapshotID {
		return graphsourcecontract.Result{}, graphsourcecontract.ErrSourceStale
	}
	if selected != binding {
		return graphsourcecontract.Result{}, graphprojection.ErrResultV2BindingMismatch
	}
	if storedBinding != binding {
		return graphsourcecontract.Result{}, graphprojection.ErrResultV2BindingMismatch
	}
	result, err := reader.ReadExactResult(ctx, binding)
	if err != nil {
		return graphsourcecontract.Result{}, err
	}
	return reportingGraphResult(ctx, result)
}

func reportingGraphResult(ctx context.Context, result graphprojection.CompletedResultV2) (graphsourcecontract.Result, error) {
	if err := ctx.Err(); err != nil {
		return graphsourcecontract.Result{}, err
	}
	if len(result.Vertices) > graphprojection.MaximumResultVerticesV2 || len(result.Edges) > graphprojection.MaximumResultEdgesV2 {
		return graphsourcecontract.Result{}, graphprojection.ErrResultV2Invalid
	}
	budget, err := newReportingLabelBudget(result.Binding, ReportingGraphSourceMaxLabelBytes)
	if err != nil {
		return graphsourcecontract.Result{}, err
	}
	candidates := graphsourcecontract.LabelCandidates{
		SchemaID:              graphsourcecontract.SchemaID,
		SourceProjectionRef:   result.Binding,
		VertexLabelCandidates: make([]graphsourcecontract.VertexLabelCandidate, 0, len(result.Vertices)),
		EdgeLabelCandidates:   make([]graphsourcecontract.EdgeLabelCandidate, 0, len(result.Edges)),
	}
	for index, vertex := range result.Vertices {
		if err := ctx.Err(); err != nil {
			return graphsourcecontract.Result{}, err
		}
		object, err := decodeReportingGraphObject(vertex.JSON)
		if err != nil || objectString(object, "vertex_id") != vertex.VertexID {
			return graphsourcecontract.Result{}, graphprojection.ErrResultV2BindingMismatch
		}
		sourceRef := objectMapValue(object, "source_entity_ref")
		sourceObjectRef := objectString(sourceRef, "source_entity_id")
		if sourceObjectRef == "" || len(sourceObjectRef) > 255 {
			return graphsourcecontract.Result{}, graphprojection.ErrResultV2BindingMismatch
		}
		properties := objectMapValue(object, "properties")
		endpoint := reportingStringComponent(
			"endpoint_value", fmt.Sprintf("/graph/vertices/%04d/endpoint_value", index+1), sourceObjectRef,
			properties["endpoint_value"],
		)
		if endpoint.State == graphsourcecontract.ComponentStatePresent {
			canonical, parseErr := parseIPLiteral(endpoint.StringValue)
			if parseErr != nil || canonical != endpoint.StringValue {
				return graphsourcecontract.Result{}, graphprojection.ErrResultV2BindingMismatch
			}
		}
		candidate := graphsourcecontract.VertexLabelCandidate{ProjectedVertexID: vertex.VertexID, Endpoint: endpoint}
		if err := budget.addVertex(candidate, index > 0); err != nil {
			return graphsourcecontract.Result{}, err
		}
		candidates.VertexLabelCandidates = append(candidates.VertexLabelCandidates, candidate)
	}
	for index, edge := range result.Edges {
		if err := ctx.Err(); err != nil {
			return graphsourcecontract.Result{}, err
		}
		object, err := decodeReportingGraphObject(edge.JSON)
		if err != nil || objectString(object, "edge_id") != edge.EdgeID {
			return graphsourcecontract.Result{}, graphprojection.ErrResultV2BindingMismatch
		}
		sourceRef := objectMapValue(object, "source_relationship_ref")
		sourceObjectRef := objectString(sourceRef, "source_relationship_id")
		if sourceObjectRef == "" || len(sourceObjectRef) > 255 {
			return graphsourcecontract.Result{}, graphprojection.ErrResultV2BindingMismatch
		}
		properties := objectMapValue(object, "properties")
		pathPrefix := fmt.Sprintf("/graph/edges/%04d", index+1)
		protocol, ok := reportingIntegerComponent("protocol", pathPrefix+"/protocol", sourceObjectRef, properties["ip_protocol"], 0, 255)
		if !ok {
			return graphsourcecontract.Result{}, graphprojection.ErrResultV2BindingMismatch
		}
		directionKind := objectString(properties, "direction_kind")
		candidate := graphsourcecontract.EdgeLabelCandidate{Kind: "default_flow_edge_v1", ProjectedEdgeID: edge.EdgeID, Protocol: protocol, DirectionKind: directionKind}
		if directionKind == directionUnknown {
			if edge.Direction != "undirected" {
				return graphsourcecontract.Result{}, graphprojection.ErrResultV2BindingMismatch
			}
			a, okA := reportingIntegerComponent("endpoint_a_port", pathPrefix+"/endpoint_a_port", sourceObjectRef, properties["src_port"], 0, 65535)
			b, okB := reportingIntegerComponent("endpoint_b_port", pathPrefix+"/endpoint_b_port", sourceObjectRef, properties["dst_port"], 0, 65535)
			if !okA || !okB {
				return graphsourcecontract.Result{}, graphprojection.ErrResultV2BindingMismatch
			}
			candidate.EndpointAPort, candidate.EndpointBPort = &a, &b
		} else {
			if directionKind != directionClientServer && directionKind != directionSourceDestination || edge.Direction != "directed" && edge.Direction != "forward" {
				return graphsourcecontract.Result{}, graphprojection.ErrResultV2BindingMismatch
			}
			port, ok := reportingIntegerComponent("to_port", pathPrefix+"/to_port", sourceObjectRef, properties["dst_port"], 0, 65535)
			if !ok {
				return graphsourcecontract.Result{}, graphprojection.ErrResultV2BindingMismatch
			}
			candidate.ToPort = &port
		}
		switch edge.EdgeKind {
		case "network_flow.flow_edge.v2":
		case "network_flow.bucketed_flow_edge.v2":
			candidate.Kind = "time_bucket_v1"
			bucketStart := reportingStringComponent("bucket_start_utc", pathPrefix+"/bucket_start_utc", sourceObjectRef, properties["bucket_start_utc"])
			bucketEnd := reportingStringComponent("bucket_end_utc", pathPrefix+"/bucket_end_utc", sourceObjectRef, properties["bucket_end_utc"])
			if !validReportingBucketComponent(bucketStart) || !validReportingBucketComponent(bucketEnd) || bucketStart.StringValue >= bucketEnd.StringValue {
				return graphsourcecontract.Result{}, graphprojection.ErrResultV2BindingMismatch
			}
			candidate.BucketStartUTC = &bucketStart
			candidate.BucketEndUTC = &bucketEnd
		default:
			return graphsourcecontract.Result{}, graphprojection.ErrResultV2BindingMismatch
		}
		if err := budget.addEdge(candidate, index > 0); err != nil {
			return graphsourcecontract.Result{}, err
		}
		candidates.EdgeLabelCandidates = append(candidates.EdgeLabelCandidates, candidate)
	}
	if err := ctx.Err(); err != nil {
		return graphsourcecontract.Result{}, err
	}
	return graphsourcecontract.Result{Projection: result, LabelCandidates: candidates}, nil
}

func decodeReportingGraphObject(raw []byte) (map[string]any, error) {
	decoder := json.NewDecoder(strings.NewReader(string(raw)))
	decoder.UseNumber()
	var object map[string]any
	if err := decoder.Decode(&object); err != nil || object == nil {
		return nil, errors.New("invalid Reporting graph object")
	}
	return object, nil
}

func objectMapValue(object map[string]any, key string) map[string]any {
	value, _ := object[key].(map[string]any)
	return value
}

func objectString(object map[string]any, key string) string {
	value, _ := object[key].(string)
	return value
}

func reportingStringComponent(kind, path, sourceObjectRef string, raw any) graphsourcecontract.LabelComponent {
	component := reportingComponent(kind, path, sourceObjectRef, graphsourcecontract.ValueKindString)
	value, ok := raw.(string)
	if !ok || value == "" {
		component.State = graphsourcecontract.ComponentStateMissing
		return component
	}
	component.State = graphsourcecontract.ComponentStatePresent
	component.StringValue = value
	return component
}

func reportingIntegerComponent(kind, path, sourceObjectRef string, raw any, minimum, maximum int64) (graphsourcecontract.LabelComponent, bool) {
	component := reportingComponent(kind, path, sourceObjectRef, graphsourcecontract.ValueKindInteger)
	if raw == nil {
		component.State = graphsourcecontract.ComponentStateAbsent
		return component, true
	}
	number, ok := raw.(json.Number)
	if !ok {
		component.State = graphsourcecontract.ComponentStateMissing
		return component, true
	}
	value, err := number.Int64()
	if err != nil || value < minimum || value > maximum {
		return graphsourcecontract.LabelComponent{}, false
	}
	component.State = graphsourcecontract.ComponentStatePresent
	component.IntegerValue = value
	return component, true
}

func reportingComponent(kind, path, sourceObjectRef, valueKind string) graphsourcecontract.LabelComponent {
	return graphsourcecontract.LabelComponent{
		ComponentKind: kind, FieldPath: path, SourceObjectRef: sourceObjectRef,
		Classification:       graphsourcecontract.ClassificationDerivedAnalytic,
		DisclosurePartitions: []string{graphsourcecontract.DisclosurePartitionInternal},
		RedactionBehavior:    graphsourcecontract.RedactionInternalOnly,
		ValueKind:            valueKind,
	}
}

func validReportingBucketComponent(component graphsourcecontract.LabelComponent) bool {
	if component.State != graphsourcecontract.ComponentStatePresent {
		return false
	}
	parsed, err := time.Parse(time.RFC3339Nano, component.StringValue)
	return err == nil && canonicalTimestampText(component.StringValue, parsed)
}
