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
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

const reportingGraphLeaseOwnerID = "snapshot_reporting"
const reportingGraphLeasePurpose = "render"

// ReportingGraphSource is Network Flow's typed immutable-result contribution.
// It owns declaration interpretation while Graph Projection owns exact result
// and lease mechanics.
type ReportingGraphSource struct {
	db    postgres.DB
	store reportingGraphDeclarationReader
}

type reportingGraphDeclarationReader interface {
	GetGraphViewDeclarationTx(context.Context, pgx.Tx, uuid.UUID, string, bool) (GraphViewDeclaration, error)
}

func NewReportingGraphSource(db postgres.DB, declarations reportingGraphDeclarationReader) (*ReportingGraphSource, error) {
	if db == nil || declarations == nil {
		return nil, errors.New("network flow Reporting graph source requires persistence")
	}
	return &ReportingGraphSource{db: db, store: declarations}, nil
}

func (m *Module) ReportingGraphSource() *ReportingGraphSource {
	if m == nil || m.store == nil {
		return nil
	}
	source, err := NewReportingGraphSource(m.store.pool, m.store)
	if err != nil {
		return nil
	}
	return source
}

func (*ReportingGraphSource) SourceOwnerID() string { return ProfileID }

func (source *ReportingGraphSource) ValidateAndLeaseResultTx(ctx context.Context, tx pgx.Tx, incidentID, jobID uuid.UUID, binding graphprojection.ResultBindingV2, observedAt, leasedUntil time.Time) (graphprojection.ResultLeaseV2, error) {
	if source == nil || source.store == nil || tx == nil || binding.SourceOwnerID != ProfileID || binding.ProjectionSchemaID != graphprojection.ProjectionSchemaIDV2 {
		return graphprojection.ResultLeaseV2{}, graphprojection.ErrResultV2BindingMismatch
	}
	reader, err := postgresresult.NewReader(tx)
	if err != nil {
		return graphprojection.ResultLeaseV2{}, err
	}
	storedBinding, err := reader.LockResultEnvelope(ctx, binding.ProjectionResultID)
	if err != nil {
		return graphprojection.ResultLeaseV2{}, err
	}
	declaration, err := source.store.GetGraphViewDeclarationTx(ctx, tx, incidentID, binding.GraphViewID, true)
	if errors.Is(err, ErrGraphViewDeclarationNotFound) {
		return graphprojection.ResultLeaseV2{}, graphprojection.ErrResultV2NotFound
	}
	if err != nil {
		return graphprojection.ResultLeaseV2{}, err
	}
	if declaration.DeclarationState != GraphViewDeclarationStateActive || declaration.SelectedResult == nil {
		return graphprojection.ResultLeaseV2{}, graphprojection.ErrResultV2NotSelected
	}
	selected := graphViewResultBinding(declaration)
	if selected.SourceSnapshotID != binding.SourceSnapshotID {
		return graphprojection.ResultLeaseV2{}, graphprojection.ErrResultV2SourceStale
	}
	if selected != binding {
		return graphprojection.ResultLeaseV2{}, graphprojection.ErrResultV2BindingMismatch
	}
	if storedBinding != binding {
		return graphprojection.ResultLeaseV2{}, graphprojection.ErrResultV2BindingMismatch
	}
	if _, err := reader.ReadExactResult(ctx, binding); err != nil {
		return graphprojection.ResultLeaseV2{}, err
	}
	writer, err := postgresresult.NewLeaseWriter(tx)
	if err != nil {
		return graphprojection.ResultLeaseV2{}, err
	}
	return writer.AcquireLease(ctx, graphprojection.ResultLeaseV2{
		LeaseID:              uuid.NewSHA1(jobID, []byte(binding.ProjectionResultID+":"+reportingGraphLeasePurpose)).String(),
		ProjectionResultID:   binding.ProjectionResultID,
		LeaseOwnerID:         reportingGraphLeaseOwnerID,
		LeaseOwnerResourceID: jobID.String(),
		LeasePurpose:         reportingGraphLeasePurpose,
		LeasedUntil:          leasedUntil.UTC(),
		CreatedAt:            observedAt.UTC(),
		RenewedAt:            observedAt.UTC(),
	})
}

func (source *ReportingGraphSource) ReadAndRenewLeasedResult(ctx context.Context, jobID uuid.UUID, binding graphprojection.ResultBindingV2, observedAt, leasedUntil time.Time) (graphsourcecontract.Result, error) {
	if source == nil || source.db == nil || jobID == uuid.Nil || binding.SourceOwnerID != ProfileID {
		return graphsourcecontract.Result{}, graphprojection.ErrResultV2BindingMismatch
	}
	var leaseID string
	err := source.db.QueryRow(ctx, `
SELECT lease_id
  FROM graph_projection_result_leases
 WHERE projection_result_id = $1
   AND lease_owner_id = $2
   AND lease_owner_resource_id = $3
   AND lease_purpose = $4
   AND leased_until > $5
`, binding.ProjectionResultID, reportingGraphLeaseOwnerID, jobID.String(), reportingGraphLeasePurpose, observedAt.UTC()).Scan(&leaseID)
	if errors.Is(err, pgx.ErrNoRows) {
		return graphsourcecontract.Result{}, graphprojection.ErrResultV2LeaseNotFound
	}
	if err != nil {
		return graphsourcecontract.Result{}, fmt.Errorf("read Reporting graph result lease: %w", err)
	}
	writer, err := postgresresult.NewLeaseWriter(source.db)
	if err != nil {
		return graphsourcecontract.Result{}, err
	}
	if _, err := writer.RenewLease(ctx, leaseID, observedAt.UTC(), leasedUntil.UTC()); err != nil {
		return graphsourcecontract.Result{}, err
	}
	reader, err := postgresresult.NewReader(source.db)
	if err != nil {
		return graphsourcecontract.Result{}, err
	}
	result, err := reader.ReadExactResult(ctx, binding)
	if err != nil {
		return graphsourcecontract.Result{}, err
	}
	return reportingGraphResult(result)
}

func reportingGraphResult(result graphprojection.CompletedResultV2) (graphsourcecontract.Result, error) {
	candidates := graphsourcecontract.LabelCandidates{
		SchemaID:              graphsourcecontract.SchemaID,
		SourceProjectionRef:   result.Binding,
		VertexLabelCandidates: make([]graphsourcecontract.VertexLabelCandidate, 0, len(result.Vertices)),
		EdgeLabelCandidates:   make([]graphsourcecontract.EdgeLabelCandidate, 0, len(result.Edges)),
	}
	for index, vertex := range result.Vertices {
		object, err := decodeReportingGraphObject(vertex.JSON)
		if err != nil || objectString(object, "vertex_id") != vertex.VertexID {
			return graphsourcecontract.Result{}, graphprojection.ErrResultV2BindingMismatch
		}
		sourceRef := objectMapValue(object, "source_entity_ref")
		sourceObjectRef := objectString(sourceRef, "source_entity_id")
		if sourceObjectRef == "" {
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
		candidates.VertexLabelCandidates = append(candidates.VertexLabelCandidates, graphsourcecontract.VertexLabelCandidate{
			ProjectedVertexID: vertex.VertexID,
			Endpoint:          endpoint,
		})
	}
	for index, edge := range result.Edges {
		object, err := decodeReportingGraphObject(edge.JSON)
		if err != nil || objectString(object, "edge_id") != edge.EdgeID {
			return graphsourcecontract.Result{}, graphprojection.ErrResultV2BindingMismatch
		}
		sourceRef := objectMapValue(object, "source_relationship_ref")
		sourceObjectRef := objectString(sourceRef, "source_relationship_id")
		if sourceObjectRef == "" {
			return graphsourcecontract.Result{}, graphprojection.ErrResultV2BindingMismatch
		}
		properties := objectMapValue(object, "properties")
		pathPrefix := fmt.Sprintf("/graph/edges/%04d", index+1)
		protocol, ok := reportingIntegerComponent("protocol", pathPrefix+"/protocol", sourceObjectRef, properties["ip_protocol"], 0, 255)
		if !ok {
			return graphsourcecontract.Result{}, graphprojection.ErrResultV2BindingMismatch
		}
		destinationPort, ok := reportingIntegerComponent("destination_port", pathPrefix+"/destination_port", sourceObjectRef, properties["dst_port"], 0, 65535)
		if !ok {
			return graphsourcecontract.Result{}, graphprojection.ErrResultV2BindingMismatch
		}
		candidate := graphsourcecontract.EdgeLabelCandidate{
			Kind: "default_flow_edge_v1", ProjectedEdgeID: edge.EdgeID,
			Protocol: protocol, DestinationPort: destinationPort,
		}
		switch edge.EdgeKind {
		case "network_flow.flow_edge.v1":
		case "network_flow.bucketed_flow_edge.v1":
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
		candidates.EdgeLabelCandidates = append(candidates.EdgeLabelCandidates, candidate)
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

func (source *ReportingGraphSource) ReleaseJobLeasesTx(ctx context.Context, tx pgx.Tx, jobID uuid.UUID) error {
	if source == nil || tx == nil || jobID == uuid.Nil {
		return nil
	}
	_, err := tx.Exec(ctx, `
DELETE FROM graph_projection_result_leases lease
 USING graph_projection_results result
 WHERE lease.projection_result_id = result.projection_result_id
   AND result.source_owner_id = $1
   AND lease.lease_owner_id = $2
   AND lease.lease_owner_resource_id = $3
   AND lease.lease_purpose = $4
`, ProfileID, reportingGraphLeaseOwnerID, jobID.String(), reportingGraphLeasePurpose)
	return err
}

func (source *ReportingGraphSource) ReleaseJobLeases(ctx context.Context, jobID uuid.UUID) error {
	if source == nil || source.db == nil || jobID == uuid.Nil {
		return nil
	}
	tx, err := source.db.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer func() { _ = tx.Rollback(ctx) }()
	if err := source.ReleaseJobLeasesTx(ctx, tx, jobID); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
