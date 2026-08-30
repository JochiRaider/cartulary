package networkflow

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/httpapi"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

func TestStreamingDefaultGraphV1Golden_Unit(t *testing.T) {
	incidentID := IncidentID()
	tableA := TableRecord{IncidentID: incidentID, TableID: "nft_" + strings.Repeat("a", 64), MappingFingerprint: strings.Repeat("1", 64)}
	tableB := TableRecord{IncidentID: incidentID, TableID: "nft_" + strings.Repeat("b", 64), MappingFingerprint: strings.Repeat("2", 64)}
	destinationPort := int32(443)
	start := time.Date(2026, 7, 13, 12, 0, 0, 0, time.UTC)
	rows := []FlowRow{
		{
			NetworkFlowTableID: tableA.TableID, RowID: "nfr_" + strings.Repeat("a", 64), SourceRowNumber: 8,
			FlowStartUTC: start.Add(2 * time.Minute), FlowEndUTC: start.Add(3 * time.Minute),
			SrcIP: "192.0.2.10", DstIP: "2001:db8::1", DstPort: &destinationPort, IPProtocol: 6,
			BytesCount: "40", PacketsCount: "4", MappingFingerprint: tableA.MappingFingerprint,
		},
		{
			NetworkFlowTableID: tableB.TableID, RowID: "nfr_" + strings.Repeat("b", 64), SourceRowNumber: 3,
			FlowStartUTC: start, FlowEndUTC: start.Add(time.Minute),
			SrcIP: "192.0.2.10", DstIP: "2001:db8::1", DstPort: &destinationPort, IPProtocol: 6,
			BytesCount: "2", PacketsCount: "1", MappingFingerprint: tableB.MappingFingerprint,
		},
	}
	tableRanks := map[string]int{tableB.TableID: 0, tableA.TableID: 1}
	sortContributorRows(rows, tableRanks)
	limits := DefaultEffectiveLimits()
	composition := graphComposition{
		SourceTables:     []TableRecord{tableB, tableA},
		TableRanks:       tableRanks,
		Vertices:         map[string]*graphVertex{},
		Edges:            map[string]*graphEdge{},
		SelectedTableIDs: []string{tableB.TableID, tableA.TableID},
		ResultLimits: graphResultLimits{
			MaxVertices:               int(limits.MaxGraphVertices),
			MaxEdges:                  int(limits.MaxGraphEdges),
			MaxExampleRowRefsPerEdge:  1,
			MaxAggregateCounterDigits: int(limits.MaxAggregateCounterDigits),
			MaxContributingRows:       int(limits.MaxContributingRowsPerGraph),
		},
		IncludeExamples: true,
	}
	if apiErr := composeGraphObjectsForTest(incidentID, rows, map[string]TableRecord{tableA.TableID: tableA, tableB.TableID: tableB}, &composition); apiErr != nil {
		t.Fatalf("compose streaming graph: %#v", apiErr)
	}
	if composition.ContributingRows != 2 || len(composition.Vertices) != 2 || len(composition.Edges) != 1 {
		t.Fatalf("streaming aggregate cardinality = rows:%d vertices:%d edges:%d", composition.ContributingRows, len(composition.Vertices), len(composition.Edges))
	}
	for _, vertex := range composition.Vertices {
		if vertex.FlowRowCount != 2 {
			t.Fatalf("vertex flow-row count = %d", vertex.FlowRowCount)
		}
	}
	edge := composition.Edges[FlowEdgeID(incidentID, EndpointID(incidentID, "ip", rows[0].SrcIP), EndpointID(incidentID, "ip", rows[0].DstIP), rows[0].IPProtocol, rows[0].DstPort)]
	if edge == nil || edge.FlowRowCount != 2 || edge.BytesSum.String() != "42" || edge.PacketsSum.String() != "5" || len(edge.ExampleRows) != 1 || edge.ExampleRows[0].RowID != rows[0].RowID {
		t.Fatalf("streaming edge aggregate = %#v", edge)
	}
	if _, found := reflect.TypeOf(graphVertex{}).FieldByName("Rows"); found {
		t.Fatal("graph vertex retained contributor rows")
	}
	if _, found := reflect.TypeOf(graphEdge{}).FieldByName("Rows"); found {
		t.Fatal("graph edge retained contributor rows")
	}

	golden := map[string]any{
		"entities":      graphProjectionEntities(composition),
		"relationships": graphProjectionRelationships(composition),
		"annotations":   graphEdgeAnnotations(composition),
	}
	if got := sha256Hex(canonicalJSON(golden)); got != "a650c1c8a12e5adca802874fb8ace5a5a373ba838a2afc05cf5472f742991be7" {
		t.Fatalf("persisted-v1 streaming golden = %s", got)
	}
}

func TestStreamingGraphRetainedStateFollowsAggregateCardinality_Unit(t *testing.T) {
	incidentID := IncidentID()
	port := int32(443)
	row := FlowRow{
		NetworkFlowTableID: "nft_" + strings.Repeat("a", 64), RowID: "nfr_" + strings.Repeat("a", 64), SourceRowNumber: 1,
		FlowStartUTC: time.Date(2026, 7, 13, 12, 0, 0, 0, time.UTC), FlowEndUTC: time.Date(2026, 7, 13, 12, 1, 0, 0, time.UTC),
		SrcIP: "192.0.2.10", DstIP: "198.51.100.20", DstPort: &port, IPProtocol: 6, BytesCount: "1", PacketsCount: "1",
	}
	const contributingRows = 4096
	composition := graphComposition{
		Vertices: map[string]*graphVertex{}, Edges: map[string]*graphEdge{},
		ResultLimits: graphResultLimits{
			MaxVertices: 2, MaxEdges: 1, MaxExampleRowRefsPerEdge: 2,
			MaxAggregateCounterDigits: 39, MaxContributingRows: contributingRows,
		},
		IncludeExamples: true,
	}
	for index := 0; index < contributingRows; index++ {
		row.SourceRowNumber = int64(index + 1)
		if apiErr := composeGraphRow(incidentID, row, nil, &composition); apiErr != nil {
			t.Fatalf("compose repeated contributor %d: %#v", index, apiErr)
		}
	}
	edge := composition.Edges[FlowEdgeID(incidentID, EndpointID(incidentID, "ip", row.SrcIP), EndpointID(incidentID, "ip", row.DstIP), row.IPProtocol, row.DstPort)]
	if composition.ContributingRows != contributingRows || len(composition.Vertices) != 2 || len(composition.Edges) != 1 || edge == nil || edge.FlowRowCount != contributingRows || len(edge.ExampleRows) != 2 {
		t.Fatalf("retained aggregate state = rows:%d vertices:%d edges:%d edge:%#v", composition.ContributingRows, len(composition.Vertices), len(composition.Edges), edge)
	}
}

func TestStreamingGraphLimitsAndCanonicalSelectors_Unit(t *testing.T) {
	incidentID := IncidentID()
	port := int32(443)
	row := FlowRow{
		NetworkFlowTableID: "nft_" + strings.Repeat("a", 64), RowID: "nfr_" + strings.Repeat("a", 64), SourceRowNumber: 1,
		FlowStartUTC: time.Date(2026, 7, 13, 12, 0, 0, 0, time.UTC), FlowEndUTC: time.Date(2026, 7, 13, 12, 1, 0, 0, time.UTC),
		SrcIP: "192.0.2.10", DstIP: "198.51.100.20", DstPort: &port, IPProtocol: 6, BytesCount: "1", PacketsCount: "1",
	}
	composition := graphComposition{
		Vertices: map[string]*graphVertex{}, Edges: map[string]*graphEdge{},
		ResultLimits: graphResultLimits{MaxVertices: 2, MaxEdges: 1, MaxExampleRowRefsPerEdge: 0, MaxAggregateCounterDigits: 39, MaxContributingRows: 1},
	}
	if apiErr := composeGraphObjectsForTest(incidentID, []FlowRow{row, row}, map[string]TableRecord{}, &composition); apiErr == nil || apiErr.Code != "network_flow_graph_limit_exceeded" || apiErr.Details["reason_code"] != "contributing_row_limit_exceeded" || apiErr.Details["actual"] != 2 {
		t.Fatalf("contributing-row limit+1 = %#v", apiErr)
	}

	vertexLimited := graphComposition{Vertices: map[string]*graphVertex{}, Edges: map[string]*graphEdge{}, ResultLimits: graphResultLimits{MaxVertices: 1, MaxEdges: 1, MaxAggregateCounterDigits: 39, MaxContributingRows: 2}}
	if apiErr := composeGraphObjectsForTest(incidentID, []FlowRow{row}, map[string]TableRecord{}, &vertexLimited); apiErr == nil || apiErr.Details["reason_code"] != "vertex_limit_exceeded" || len(vertexLimited.Vertices) != 2 {
		t.Fatalf("vertex limit+1 = %#v / %d", apiErr, len(vertexLimited.Vertices))
	}

	srcID := EndpointID(incidentID, "ip", row.SrcIP)
	dstID := EndpointID(incidentID, "ip", row.DstIP)
	edgeID := FlowEdgeID(incidentID, srcID, dstID, row.IPProtocol, row.DstPort)
	raw, _ := json.Marshal(map[string]any{
		"kind": "default_edge", "source_edge_id": edgeID,
		"source_endpoint_value": row.SrcIP, "destination_endpoint_value": row.DstIP,
		"protocol": row.IPProtocol, "destination_port_present": true, "destination_port": port,
	})
	selector, apiErr := decodeGraphSelector(raw)
	if apiErr != nil {
		t.Fatalf("decode canonical edge selector: %#v", apiErr)
	}
	predicate, apiErr := canonicalGraphContributorPredicate(incidentID, selector)
	if apiErr != nil || predicate.Kind != "default_edge" || predicate.DestinationPort == nil || *predicate.DestinationPort != port {
		t.Fatalf("canonical edge predicate = %#v err=%#v", predicate, apiErr)
	}
	selector.SourceEdgeID = FlowEdgeID(incidentID, dstID, srcID, row.IPProtocol, row.DstPort)
	if _, apiErr := canonicalGraphContributorPredicate(incidentID, selector); apiErr == nil || apiErr.Details["reason_code"] != "id_key_mismatch" {
		t.Fatalf("mismatched selector ID = %#v", apiErr)
	}
	if _, apiErr := decodeGraphSelector(json.RawMessage(`{"kind":"default_edge","source_edge_id":"nff_invalid","source_endpoint_value":"192.0.2.10","destination_endpoint_value":"198.51.100.20","protocol":6,"destination_port_present":false,"destination_port":443}`)); apiErr == nil || apiErr.Details["reason_code"] != "variant_member_conflict" {
		t.Fatalf("conflicting destination-port selector = %#v", apiErr)
	}
}

func TestDefaultGraphV2IdentityFixture_Unit(t *testing.T) {
	incidentID := IncidentID()
	tables := []string{"nft_" + strings.Repeat("b", 64), "nft_" + strings.Repeat("a", 64)}
	aggregation := graphAggregation{Mode: "default_flow_edge_v1", IncludeExampleRowRefs: true}
	timeRange := graphTimeRange{Omitted: true}
	digest := graphQueryDigestV2(incidentID, tables, nil, timeRange, aggregation)
	if digest != "0da3d5731bfa0c924c3baddc41d44f178a2dd62f41f27d2824b2d3eb111ba5b6" {
		t.Fatalf("default semantic-query-v2 digest fixture = %s", digest)
	}
	if reversed := graphQueryDigestV2(incidentID, []string{tables[1], tables[0]}, []Filter{}, timeRange, aggregation); reversed != digest {
		t.Fatalf("v2 table order changed digest: %s != %s", reversed, digest)
	}
}

func TestStreamingGraphDatabaseErrorsRemainVisible_Unit(t *testing.T) {
	sentinel := errors.New("ordered graph query failed")
	store := NewStore(&graphQueryErrorDB{queryErr: sentinel}, DefaultEffectiveLimits())
	err := store.IterateRowsForTables(context.Background(), IncidentID(), []string{"nft_" + strings.Repeat("a", 64)}, func(FlowRow) error {
		t.Fatal("visitor ran after query failure")
		return nil
	})
	if !errors.Is(err, sentinel) {
		t.Fatalf("ordered graph query error = %v", err)
	}
}

type graphQueryErrorDB struct {
	postgres.DB
	queryErr error
}

func (db *graphQueryErrorDB) Query(context.Context, string, ...any) (pgx.Rows, error) {
	return nil, db.queryErr
}

var _ postgres.DB = (*graphQueryErrorDB)(nil)

func composeGraphObjectsForTest(incidentID uuid.UUID, rows []FlowRow, tableByID map[string]TableRecord, composition *graphComposition) *httpapi.APIError {
	for _, row := range rows {
		if apiErr := composeGraphRow(incidentID, row, tableByID, composition); apiErr != nil {
			return apiErr
		}
	}
	return nil
}
