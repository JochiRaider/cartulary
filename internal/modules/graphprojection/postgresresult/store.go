package postgresresult

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/JochiRaider/cartulary/internal/modules/graphprojection"
)

var (
	resultIDPattern   = regexp.MustCompile(`^gpres_[a-f0-9]{64}$`)
	vertexIDPattern   = regexp.MustCompile(`^vx_[a-f0-9]{64}$`)
	edgeIDPattern     = regexp.MustCompile(`^ed_[a-f0-9]{64}$`)
	sha256Pattern     = regexp.MustCompile(`^[a-f0-9]{64}$`)
	leaseTokenPattern = regexp.MustCompile(`^[a-z][a-z0-9_.:-]{0,254}$`)
)

type queryer interface {
	Exec(context.Context, string, ...any) (pgconn.CommandTag, error)
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

type Publisher struct {
	tx pgx.Tx
}

func NewPublisher(tx pgx.Tx) (*Publisher, error) {
	if tx == nil {
		return nil, fmt.Errorf("graph projection result publisher transaction is required")
	}
	return &Publisher{tx: tx}, nil
}

func (publisher *Publisher) PublishResult(ctx context.Context, result graphprojection.CompletedResultV2) error {
	if publisher == nil || publisher.tx == nil {
		return fmt.Errorf("publish Graph Projection result: %w", graphprojection.ErrResultV2Invalid)
	}
	if err := validateCompletedResult(result); err != nil {
		return err
	}

	binding := result.Binding
	tag, err := publisher.tx.Exec(ctx, `
INSERT INTO graph_projection_results (
    projection_result_id, graph_view_id, source_owner_id, source_snapshot_id,
    projection_schema_id, projection_version,
    normalized_configuration_sha256, normalized_source_sha256,
    canonical_output_sha256, vertex_count, edge_count, result_json, published_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
ON CONFLICT (projection_result_id) DO NOTHING
`, binding.ProjectionResultID, binding.GraphViewID, binding.SourceOwnerID, binding.SourceSnapshotID,
		binding.ProjectionSchemaID, binding.ProjectionVersion,
		binding.NormalizedConfigurationSHA256, binding.NormalizedSourceSHA256,
		binding.CanonicalOutputSHA256, len(result.Vertices), len(result.Edges), []byte(result.ResultJSON), result.PublishedAt.UTC())
	if err != nil {
		return fmt.Errorf("insert Graph Projection result envelope: %w", err)
	}
	if tag.RowsAffected() == 0 {
		return publisher.requireExactExistingResult(ctx, result)
	}

	for ordinal, vertex := range result.Vertices {
		if _, err := publisher.tx.Exec(ctx, `
INSERT INTO graph_projection_result_vertices (
    projection_result_id, vertex_id, vertex_kind, sort_ordinal, sort_key, vertex_json
) VALUES ($1, $2, $3, $4, $5, $6)
`, binding.ProjectionResultID, vertex.VertexID, vertex.VertexKind, ordinal, vertex.SortKey, []byte(vertex.JSON)); err != nil {
			return fmt.Errorf("insert Graph Projection result vertex: %w", err)
		}
	}
	for ordinal, edge := range result.Edges {
		if _, err := publisher.tx.Exec(ctx, `
INSERT INTO graph_projection_result_edges (
    projection_result_id, edge_id, edge_kind, src_vertex_id, dst_vertex_id,
    direction, sort_ordinal, sort_key, edge_json
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9)
`, binding.ProjectionResultID, edge.EdgeID, edge.EdgeKind, edge.SrcVertexID, edge.DstVertexID,
			edge.Direction, ordinal, edge.SortKey, []byte(edge.JSON)); err != nil {
			return fmt.Errorf("insert Graph Projection result edge: %w", err)
		}
	}
	return nil
}

func (publisher *Publisher) requireExactExistingResult(ctx context.Context, want graphprojection.CompletedResultV2) error {
	reader, err := NewReader(publisher.tx)
	if err != nil {
		return err
	}
	got, err := reader.ReadExactResult(ctx, want.Binding)
	if err != nil {
		if errors.Is(err, graphprojection.ErrResultV2BindingMismatch) || errors.Is(err, graphprojection.ErrResultV2NotFound) {
			return graphprojection.ErrResultV2IdentityConflict
		}
		return err
	}
	if !bytes.Equal(got.ResultJSON, want.ResultJSON) ||
		!equalVertices(got.Vertices, want.Vertices) || !equalEdges(got.Edges, want.Edges) {
		return graphprojection.ErrResultV2IdentityConflict
	}
	return nil
}

type Reader struct {
	db queryer
}

func NewReader(db queryer) (*Reader, error) {
	if db == nil {
		return nil, fmt.Errorf("graph projection result reader database is required")
	}
	return &Reader{db: db}, nil
}

func (reader *Reader) ReadExactResult(ctx context.Context, binding graphprojection.ResultBindingV2) (graphprojection.CompletedResultV2, error) {
	if err := validateBinding(binding); err != nil {
		return graphprojection.CompletedResultV2{}, err
	}
	var result graphprojection.CompletedResultV2
	var vertexCount, edgeCount int
	result.Binding.ProjectionResultID = binding.ProjectionResultID
	err := reader.db.QueryRow(ctx, `
SELECT graph_view_id, source_owner_id, source_snapshot_id, projection_schema_id,
       projection_version, normalized_configuration_sha256, normalized_source_sha256,
       canonical_output_sha256, vertex_count, edge_count, result_json, published_at
  FROM graph_projection_results
 WHERE projection_result_id = $1
`, binding.ProjectionResultID).Scan(
		&result.Binding.GraphViewID, &result.Binding.SourceOwnerID, &result.Binding.SourceSnapshotID,
		&result.Binding.ProjectionSchemaID, &result.Binding.ProjectionVersion,
		&result.Binding.NormalizedConfigurationSHA256, &result.Binding.NormalizedSourceSHA256,
		&result.Binding.CanonicalOutputSHA256, &vertexCount, &edgeCount, &result.ResultJSON, &result.PublishedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return graphprojection.CompletedResultV2{}, graphprojection.ErrResultV2NotFound
	}
	if err != nil {
		return graphprojection.CompletedResultV2{}, fmt.Errorf("read Graph Projection result envelope: %w", err)
	}
	if result.Binding != binding {
		return graphprojection.CompletedResultV2{}, graphprojection.ErrResultV2BindingMismatch
	}
	result.Vertices, err = reader.ReadVertices(ctx, binding.ProjectionResultID, graphprojection.MaximumResultVerticesV2)
	if err != nil {
		return graphprojection.CompletedResultV2{}, err
	}
	result.Edges, err = reader.ReadEdges(ctx, binding.ProjectionResultID, graphprojection.MaximumResultEdgesV2)
	if err != nil {
		return graphprojection.CompletedResultV2{}, err
	}
	if len(result.Vertices) != vertexCount || len(result.Edges) != edgeCount {
		return graphprojection.CompletedResultV2{}, graphprojection.ErrResultV2IdentityConflict
	}
	return result, nil
}

func (reader *Reader) ReadVertices(ctx context.Context, projectionResultID string, maximum int) ([]graphprojection.ResultVertexV2, error) {
	if !resultIDPattern.MatchString(projectionResultID) || maximum < 1 || maximum > graphprojection.MaximumResultVerticesV2 {
		return nil, graphprojection.ErrResultV2Invalid
	}
	rows, err := reader.db.Query(ctx, `
SELECT vertex_id, vertex_kind, sort_key, vertex_json
  FROM graph_projection_result_vertices
 WHERE projection_result_id = $1
 ORDER BY sort_ordinal ASC
 LIMIT $2
`, projectionResultID, maximum+1)
	if err != nil {
		return nil, fmt.Errorf("read Graph Projection result vertices: %w", err)
	}
	defer rows.Close()
	vertices := make([]graphprojection.ResultVertexV2, 0)
	for rows.Next() {
		var vertex graphprojection.ResultVertexV2
		if err := rows.Scan(&vertex.VertexID, &vertex.VertexKind, &vertex.SortKey, &vertex.JSON); err != nil {
			return nil, fmt.Errorf("scan Graph Projection result vertex: %w", err)
		}
		vertices = append(vertices, vertex)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read Graph Projection result vertices: %w", err)
	}
	if len(vertices) > maximum {
		return nil, graphprojection.ErrResultV2Invalid
	}
	return vertices, nil
}

func (reader *Reader) ReadEdges(ctx context.Context, projectionResultID string, maximum int) ([]graphprojection.ResultEdgeV2, error) {
	if !resultIDPattern.MatchString(projectionResultID) || maximum < 1 || maximum > graphprojection.MaximumResultEdgesV2 {
		return nil, graphprojection.ErrResultV2Invalid
	}
	rows, err := reader.db.Query(ctx, `
SELECT edge_id, edge_kind, src_vertex_id, dst_vertex_id, direction, sort_key, edge_json
  FROM graph_projection_result_edges
 WHERE projection_result_id = $1
 ORDER BY sort_ordinal ASC
 LIMIT $2
`, projectionResultID, maximum+1)
	if err != nil {
		return nil, fmt.Errorf("read Graph Projection result edges: %w", err)
	}
	defer rows.Close()
	edges := make([]graphprojection.ResultEdgeV2, 0)
	for rows.Next() {
		var edge graphprojection.ResultEdgeV2
		if err := rows.Scan(&edge.EdgeID, &edge.EdgeKind, &edge.SrcVertexID, &edge.DstVertexID, &edge.Direction, &edge.SortKey, &edge.JSON); err != nil {
			return nil, fmt.Errorf("scan Graph Projection result edge: %w", err)
		}
		edges = append(edges, edge)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("read Graph Projection result edges: %w", err)
	}
	if len(edges) > maximum {
		return nil, graphprojection.ErrResultV2Invalid
	}
	return edges, nil
}

func (reader *Reader) Traverse(ctx context.Context, request graphprojection.TraversalRequestV2) (graphprojection.TraversalResultV2, error) {
	if err := validateTraversalRequest(request); err != nil {
		return graphprojection.TraversalResultV2{}, err
	}
	vertices, err := reader.ReadVertices(ctx, request.ProjectionResultID, graphprojection.MaximumResultVerticesV2)
	if err != nil {
		return graphprojection.TraversalResultV2{}, err
	}
	edges, err := reader.ReadEdges(ctx, request.ProjectionResultID, graphprojection.MaximumResultEdgesV2)
	if err != nil {
		return graphprojection.TraversalResultV2{}, err
	}

	vertexByID := make(map[string]graphprojection.ResultVertexV2, len(vertices))
	vertexKinds := makeStringMembership(request.VertexKinds)
	for _, vertex := range vertices {
		vertexByID[vertex.VertexID] = vertex
	}
	edgeKinds := makeStringMembership(request.EdgeKinds)
	selectedVertices := make(map[string]struct{})
	frontier := make(map[string]struct{})
	for _, seed := range request.SeedVertexIDs {
		vertex, ok := vertexByID[seed]
		if !ok || !kindAllowed(vertex.VertexKind, vertexKinds) {
			return graphprojection.TraversalResultV2{}, graphprojection.ErrResultV2Invalid
		}
		selectedVertices[seed] = struct{}{}
		frontier[seed] = struct{}{}
	}
	selectedEdges := make(map[string]struct{})
	if len(selectedVertices) > request.MaximumVertices {
		return graphprojection.TraversalResultV2{}, graphprojection.ErrResultV2Invalid
	}
	for depth := 0; depth < request.MaximumDepth && len(frontier) > 0; depth++ {
		next := make(map[string]struct{})
		for _, edge := range edges {
			if !kindAllowed(edge.EdgeKind, edgeKinds) {
				continue
			}
			neighbor := ""
			switch request.Direction {
			case graphprojection.TraversalOutgoingV2:
				if _, ok := frontier[edge.SrcVertexID]; ok {
					neighbor = edge.DstVertexID
				}
			case graphprojection.TraversalIncomingV2:
				if _, ok := frontier[edge.DstVertexID]; ok {
					neighbor = edge.SrcVertexID
				}
			case graphprojection.TraversalBothV2:
				if _, ok := frontier[edge.SrcVertexID]; ok {
					neighbor = edge.DstVertexID
				} else if _, ok := frontier[edge.DstVertexID]; ok {
					neighbor = edge.SrcVertexID
				}
			}
			if neighbor == "" {
				continue
			}
			neighborVertex, ok := vertexByID[neighbor]
			if !ok || !kindAllowed(neighborVertex.VertexKind, vertexKinds) {
				continue
			}
			selectedEdges[edge.EdgeID] = struct{}{}
			if _, seen := selectedVertices[neighbor]; !seen {
				selectedVertices[neighbor] = struct{}{}
				next[neighbor] = struct{}{}
			}
			if len(selectedVertices) > request.MaximumVertices || len(selectedEdges) > request.MaximumEdges {
				return graphprojection.TraversalResultV2{}, graphprojection.ErrResultV2Invalid
			}
		}
		frontier = next
	}

	result := graphprojection.TraversalResultV2{
		Vertices: make([]graphprojection.ResultVertexV2, 0, len(selectedVertices)),
		Edges:    make([]graphprojection.ResultEdgeV2, 0, len(selectedEdges)),
	}
	for _, vertex := range vertices {
		if _, ok := selectedVertices[vertex.VertexID]; ok {
			result.Vertices = append(result.Vertices, vertex)
		}
	}
	for _, edge := range edges {
		if _, ok := selectedEdges[edge.EdgeID]; ok {
			result.Edges = append(result.Edges, edge)
		}
	}
	return result, nil
}

type LeaseWriter struct {
	db queryer
}

func NewLeaseWriter(db queryer) (*LeaseWriter, error) {
	if db == nil {
		return nil, fmt.Errorf("graph projection lease writer database is required")
	}
	return &LeaseWriter{db: db}, nil
}

func (writer *LeaseWriter) AcquireLease(ctx context.Context, lease graphprojection.ResultLeaseV2) (graphprojection.ResultLeaseV2, error) {
	if err := validateLease(lease); err != nil {
		return graphprojection.ResultLeaseV2{}, err
	}
	var acquired graphprojection.ResultLeaseV2
	err := writer.db.QueryRow(ctx, `
INSERT INTO graph_projection_result_leases (
    lease_id, projection_result_id, lease_owner_id, lease_owner_resource_id,
    lease_purpose, leased_until, created_at, renewed_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (projection_result_id, lease_owner_id, lease_owner_resource_id, lease_purpose)
DO UPDATE SET
    leased_until = GREATEST(graph_projection_result_leases.leased_until, EXCLUDED.leased_until),
    renewed_at = GREATEST(graph_projection_result_leases.renewed_at, EXCLUDED.renewed_at)
RETURNING lease_id, projection_result_id, lease_owner_id, lease_owner_resource_id,
          lease_purpose, leased_until, created_at, renewed_at
`, lease.LeaseID, lease.ProjectionResultID, lease.LeaseOwnerID, lease.LeaseOwnerResourceID,
		lease.LeasePurpose, lease.LeasedUntil.UTC(), lease.CreatedAt.UTC(), lease.RenewedAt.UTC()).Scan(
		&acquired.LeaseID, &acquired.ProjectionResultID, &acquired.LeaseOwnerID,
		&acquired.LeaseOwnerResourceID, &acquired.LeasePurpose, &acquired.LeasedUntil,
		&acquired.CreatedAt, &acquired.RenewedAt,
	)
	if err != nil {
		return graphprojection.ResultLeaseV2{}, fmt.Errorf("acquire Graph Projection result lease: %w", err)
	}
	return acquired, nil
}

func (writer *LeaseWriter) RenewLease(ctx context.Context, leaseID string, observedAt, leasedUntil time.Time) (graphprojection.ResultLeaseV2, error) {
	if _, err := uuid.Parse(leaseID); err != nil || observedAt.IsZero() || !leasedUntil.After(observedAt) {
		return graphprojection.ResultLeaseV2{}, graphprojection.ErrResultV2Invalid
	}
	var renewed graphprojection.ResultLeaseV2
	err := writer.db.QueryRow(ctx, `
UPDATE graph_projection_result_leases
   SET leased_until = $3,
       renewed_at = $2
 WHERE lease_id = $1
   AND leased_until > $2
RETURNING lease_id, projection_result_id, lease_owner_id, lease_owner_resource_id,
          lease_purpose, leased_until, created_at, renewed_at
`, leaseID, observedAt.UTC(), leasedUntil.UTC()).Scan(
		&renewed.LeaseID, &renewed.ProjectionResultID, &renewed.LeaseOwnerID,
		&renewed.LeaseOwnerResourceID, &renewed.LeasePurpose, &renewed.LeasedUntil,
		&renewed.CreatedAt, &renewed.RenewedAt,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return graphprojection.ResultLeaseV2{}, graphprojection.ErrResultV2LeaseNotFound
	}
	if err != nil {
		return graphprojection.ResultLeaseV2{}, fmt.Errorf("renew Graph Projection result lease: %w", err)
	}
	return renewed, nil
}

func (writer *LeaseWriter) ReleaseLease(ctx context.Context, leaseID string) error {
	if _, err := uuid.Parse(leaseID); err != nil {
		return graphprojection.ErrResultV2Invalid
	}
	if _, err := writer.db.Exec(ctx, `DELETE FROM graph_projection_result_leases WHERE lease_id = $1`, leaseID); err != nil {
		return fmt.Errorf("release Graph Projection result lease: %w", err)
	}
	return nil
}

type Cleaner struct {
	tx pgx.Tx
}

func NewCleaner(tx pgx.Tx) (*Cleaner, error) {
	if tx == nil {
		return nil, fmt.Errorf("graph projection reachability cleaner transaction is required")
	}
	return &Cleaner{tx: tx}, nil
}

func (cleaner *Cleaner) DeleteUnreachableResults(ctx context.Context, reachableResultIDs []string, observedAt time.Time, maximum int) ([]string, error) {
	if observedAt.IsZero() || maximum < 1 || maximum > 10000 || !slices.IsSorted(reachableResultIDs) {
		return nil, graphprojection.ErrResultV2Invalid
	}
	for index, resultID := range reachableResultIDs {
		if !resultIDPattern.MatchString(resultID) || (index > 0 && resultID == reachableResultIDs[index-1]) {
			return nil, graphprojection.ErrResultV2Invalid
		}
	}
	if _, err := cleaner.tx.Exec(ctx, `
DELETE FROM graph_projection_result_leases
 WHERE leased_until <= $1
`, observedAt.UTC()); err != nil {
		return nil, fmt.Errorf("delete expired Graph Projection result leases: %w", err)
	}
	rows, err := cleaner.tx.Query(ctx, `
WITH candidates AS (
    SELECT result.projection_result_id
      FROM graph_projection_results result
     WHERE NOT (result.projection_result_id = ANY($1::text[]))
       AND NOT EXISTS (
            SELECT 1
              FROM graph_projection_result_leases lease
             WHERE lease.projection_result_id = result.projection_result_id
               AND lease.leased_until > $2
       )
     ORDER BY result.projection_result_id ASC
     LIMIT $3
     FOR UPDATE
), deleted AS (
    DELETE FROM graph_projection_results result
     USING candidates
     WHERE result.projection_result_id = candidates.projection_result_id
     RETURNING result.projection_result_id
)
SELECT projection_result_id FROM deleted ORDER BY projection_result_id ASC
`, reachableResultIDs, observedAt.UTC(), maximum)
	if err != nil {
		return nil, fmt.Errorf("delete unreachable Graph Projection results: %w", err)
	}
	defer rows.Close()
	deleted := make([]string, 0)
	for rows.Next() {
		var resultID string
		if err := rows.Scan(&resultID); err != nil {
			return nil, fmt.Errorf("scan deleted Graph Projection result: %w", err)
		}
		deleted = append(deleted, resultID)
	}
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("delete unreachable Graph Projection results: %w", err)
	}
	return deleted, nil
}

func validateCompletedResult(result graphprojection.CompletedResultV2) error {
	if err := validateBinding(result.Binding); err != nil || result.PublishedAt.IsZero() || !validJSONObject(result.ResultJSON) ||
		len(result.Vertices) > graphprojection.MaximumResultVerticesV2 || len(result.Edges) > graphprojection.MaximumResultEdgesV2 {
		return graphprojection.ErrResultV2Invalid
	}
	vertices := make(map[string]struct{}, len(result.Vertices))
	for index, vertex := range result.Vertices {
		if !vertexIDPattern.MatchString(vertex.VertexID) || strings.TrimSpace(vertex.VertexKind) == "" || !validJSONObject(vertex.JSON) {
			return graphprojection.ErrResultV2Invalid
		}
		if _, duplicate := vertices[vertex.VertexID]; duplicate {
			return graphprojection.ErrResultV2Invalid
		}
		vertices[vertex.VertexID] = struct{}{}
		if index > 0 && compareOrderedObject(result.Vertices[index-1].SortKey, result.Vertices[index-1].VertexID, vertex.SortKey, vertex.VertexID) >= 0 {
			return graphprojection.ErrResultV2Invalid
		}
	}
	edges := make(map[string]struct{}, len(result.Edges))
	for index, edge := range result.Edges {
		if !edgeIDPattern.MatchString(edge.EdgeID) || strings.TrimSpace(edge.EdgeKind) == "" ||
			(edge.Direction != "directed" && edge.Direction != "undirected" && edge.Direction != "bidirectional") || !validJSONObject(edge.JSON) {
			return graphprojection.ErrResultV2Invalid
		}
		if _, duplicate := edges[edge.EdgeID]; duplicate {
			return graphprojection.ErrResultV2Invalid
		}
		edges[edge.EdgeID] = struct{}{}
		if _, ok := vertices[edge.SrcVertexID]; !ok {
			return graphprojection.ErrResultV2Invalid
		}
		if _, ok := vertices[edge.DstVertexID]; !ok {
			return graphprojection.ErrResultV2Invalid
		}
		if index > 0 && compareOrderedObject(result.Edges[index-1].SortKey, result.Edges[index-1].EdgeID, edge.SortKey, edge.EdgeID) >= 0 {
			return graphprojection.ErrResultV2Invalid
		}
	}
	return nil
}

func validateBinding(binding graphprojection.ResultBindingV2) error {
	if !resultIDPattern.MatchString(binding.ProjectionResultID) ||
		strings.TrimSpace(binding.GraphViewID) == "" || strings.TrimSpace(binding.SourceOwnerID) == "" ||
		strings.TrimSpace(binding.SourceSnapshotID) == "" || binding.ProjectionSchemaID != graphprojection.ProjectionSchemaIDV2 ||
		strings.TrimSpace(binding.ProjectionVersion) == "" ||
		!sha256Pattern.MatchString(binding.NormalizedConfigurationSHA256) ||
		!sha256Pattern.MatchString(binding.NormalizedSourceSHA256) ||
		!sha256Pattern.MatchString(binding.CanonicalOutputSHA256) {
		return graphprojection.ErrResultV2Invalid
	}
	return nil
}

func validateLease(lease graphprojection.ResultLeaseV2) error {
	if _, err := uuid.Parse(lease.LeaseID); err != nil || !resultIDPattern.MatchString(lease.ProjectionResultID) ||
		!leaseTokenPattern.MatchString(lease.LeaseOwnerID) || strings.TrimSpace(lease.LeaseOwnerResourceID) == "" ||
		!leaseTokenPattern.MatchString(lease.LeasePurpose) || lease.CreatedAt.IsZero() || lease.RenewedAt.IsZero() ||
		lease.RenewedAt.Before(lease.CreatedAt) || !lease.LeasedUntil.After(lease.RenewedAt) {
		return graphprojection.ErrResultV2Invalid
	}
	return nil
}

func validateTraversalRequest(request graphprojection.TraversalRequestV2) error {
	if !resultIDPattern.MatchString(request.ProjectionResultID) || len(request.SeedVertexIDs) < 1 || len(request.SeedVertexIDs) > 1024 ||
		request.MaximumDepth < 0 || request.MaximumDepth > 16 || request.MaximumVertices < 1 ||
		request.MaximumVertices > graphprojection.MaximumResultVerticesV2 || request.MaximumEdges < 0 ||
		request.MaximumEdges > graphprojection.MaximumResultEdgesV2 || len(request.VertexKinds) > 1024 || len(request.EdgeKinds) > 1024 ||
		(request.Direction != graphprojection.TraversalOutgoingV2 && request.Direction != graphprojection.TraversalIncomingV2 && request.Direction != graphprojection.TraversalBothV2) {
		return graphprojection.ErrResultV2Invalid
	}
	for _, values := range [][]string{request.SeedVertexIDs, request.VertexKinds, request.EdgeKinds} {
		if !slices.IsSorted(values) {
			return graphprojection.ErrResultV2Invalid
		}
		for index, value := range values {
			if strings.TrimSpace(value) == "" || (index > 0 && value == values[index-1]) {
				return graphprojection.ErrResultV2Invalid
			}
		}
	}
	for _, seed := range request.SeedVertexIDs {
		if !vertexIDPattern.MatchString(seed) {
			return graphprojection.ErrResultV2Invalid
		}
	}
	return nil
}

func validJSONObject(value json.RawMessage) bool {
	trimmed := bytes.TrimSpace(value)
	return len(trimmed) >= 2 && trimmed[0] == '{' && trimmed[len(trimmed)-1] == '}' && json.Valid(trimmed)
}

func compareOrderedObject(leftKey, leftID, rightKey, rightID string) int {
	if comparison := strings.Compare(leftKey, rightKey); comparison != 0 {
		return comparison
	}
	return strings.Compare(leftID, rightID)
}

func equalVertices(left, right []graphprojection.ResultVertexV2) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index].VertexID != right[index].VertexID || left[index].VertexKind != right[index].VertexKind ||
			left[index].SortKey != right[index].SortKey || !bytes.Equal(left[index].JSON, right[index].JSON) {
			return false
		}
	}
	return true
}

func equalEdges(left, right []graphprojection.ResultEdgeV2) bool {
	if len(left) != len(right) {
		return false
	}
	for index := range left {
		if left[index].EdgeID != right[index].EdgeID || left[index].EdgeKind != right[index].EdgeKind ||
			left[index].SrcVertexID != right[index].SrcVertexID || left[index].DstVertexID != right[index].DstVertexID ||
			left[index].Direction != right[index].Direction || left[index].SortKey != right[index].SortKey ||
			!bytes.Equal(left[index].JSON, right[index].JSON) {
			return false
		}
	}
	return true
}

func makeStringMembership(values []string) map[string]struct{} {
	if len(values) == 0 {
		return nil
	}
	result := make(map[string]struct{}, len(values))
	for _, value := range values {
		result[value] = struct{}{}
	}
	return result
}

func kindAllowed(value string, allowed map[string]struct{}) bool {
	if len(allowed) == 0 {
		return true
	}
	_, ok := allowed[value]
	return ok
}

var _ graphprojection.ResultPublisherV2 = (*Publisher)(nil)
var _ graphprojection.ExactResultReaderV2 = (*Reader)(nil)
var _ graphprojection.ResultLeaseWriterV2 = (*LeaseWriter)(nil)
var _ graphprojection.ReachabilityCleanerV2 = (*Cleaner)(nil)
