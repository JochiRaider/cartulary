package postgresresult

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"regexp"
	"slices"
	"strings"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/modules/graphprojection"
	"github.com/JochiRaider/cartulary/internal/modules/graphprojection/internal/semanticlimits"
)

var graphProjectionLimits = semanticlimits.CurrentV2()

var (
	resultIDPattern         = regexp.MustCompile(`^gpres_[a-f0-9]{64}$`)
	vertexIDPattern         = regexp.MustCompile(`^vx_[a-f0-9]{64}$`)
	sourceOwnerTokenPattern = regexp.MustCompile(`^[a-z][a-z0-9_.:-]{0,254}$`)
)

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
	if err := validateCompletedResult(ctx, result); err != nil {
		return err
	}

	inserted, err := publisher.insertResultEnvelope(ctx, result)
	if err != nil {
		return err
	}
	if !inserted {
		err = publisher.requireExactExistingResult(ctx, result)
		if !errors.Is(err, graphprojection.ErrResultV2NotFound) {
			return err
		}
		// A cleanup transaction can delete the conflicting row after this
		// transaction's INSERT began but before ON CONFLICT resolves. Retry the
		// insert exactly once after the vanished conflict; a new insert remains
		// locked until the caller publishes its source declaration.
		inserted, err = publisher.insertResultEnvelope(ctx, result)
		if err != nil {
			return err
		}
		if !inserted {
			return publisher.requireExactExistingResult(ctx, result)
		}
	}
	return publisher.insertResultObjects(ctx, result)
}

func (publisher *Publisher) insertResultEnvelope(ctx context.Context, result graphprojection.CompletedResultV2) (bool, error) {
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
		return false, fmt.Errorf("insert Graph Projection result envelope: %w", err)
	}
	return tag.RowsAffected() == 1, nil
}

func (publisher *Publisher) insertResultObjects(ctx context.Context, result graphprojection.CompletedResultV2) error {
	binding := result.Binding
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
	got, err := reader.ReadExactResultForUpdate(ctx, want.Binding)
	if err != nil {
		if errors.Is(err, graphprojection.ErrResultV2NotFound) {
			return graphprojection.ErrResultV2NotFound
		}
		if errors.Is(err, graphprojection.ErrResultV2BindingMismatch) {
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

type resultReader interface {
	Query(context.Context, string, ...any) (pgx.Rows, error)
	QueryRow(context.Context, string, ...any) pgx.Row
}

type Reader struct {
	db resultReader
}

func NewReader(db resultReader) (*Reader, error) {
	if db == nil {
		return nil, fmt.Errorf("graph projection result reader database is required")
	}
	return &Reader{db: db}, nil
}

// ReadResultEnvelope reads only immutable identity metadata, without mounting
// graph objects or acquiring a publication lock.
func (reader *Reader) ReadResultEnvelope(ctx context.Context, projectionResultID string) (graphprojection.ResultBindingV2, error) {
	return reader.readResultEnvelope(ctx, projectionResultID, false)
}

// LockResultEnvelope locks one immutable result and returns its stored binding
// without interpreting a source owner's declaration. Source-owner adapters use
// it before taking their declaration lock, then apply their own error ordering.
func (reader *Reader) LockResultEnvelope(ctx context.Context, projectionResultID string) (graphprojection.ResultBindingV2, error) {
	return reader.readResultEnvelope(ctx, projectionResultID, true)
}

func (reader *Reader) readResultEnvelope(ctx context.Context, projectionResultID string, lock bool) (graphprojection.ResultBindingV2, error) {
	if reader == nil || reader.db == nil || !resultIDPattern.MatchString(projectionResultID) {
		return graphprojection.ResultBindingV2{}, graphprojection.ErrResultV2Invalid
	}
	binding := graphprojection.ResultBindingV2{ProjectionResultID: projectionResultID}
	query := `
SELECT graph_view_id, source_owner_id, source_snapshot_id, projection_schema_id,
       projection_version, normalized_configuration_sha256, normalized_source_sha256,
       canonical_output_sha256
  FROM graph_projection_results
 WHERE projection_result_id = $1
`
	if lock {
		query += ` FOR UPDATE`
	}
	err := reader.db.QueryRow(ctx, query, projectionResultID).Scan(
		&binding.GraphViewID, &binding.SourceOwnerID, &binding.SourceSnapshotID,
		&binding.ProjectionSchemaID, &binding.ProjectionVersion,
		&binding.NormalizedConfigurationSHA256, &binding.NormalizedSourceSHA256,
		&binding.CanonicalOutputSHA256,
	)
	if errors.Is(err, pgx.ErrNoRows) {
		return graphprojection.ResultBindingV2{}, graphprojection.ErrResultV2NotFound
	}
	if err != nil {
		return graphprojection.ResultBindingV2{}, fmt.Errorf("read Graph Projection result envelope: %w", err)
	}
	if err := binding.ValidateIdentity(); err != nil {
		return graphprojection.ResultBindingV2{}, err
	}
	return binding, nil
}

func (reader *Reader) ReadExactResult(ctx context.Context, binding graphprojection.ResultBindingV2) (graphprojection.CompletedResultV2, error) {
	return reader.readExactResult(ctx, binding, false)
}

// ReadExactResultForUpdate locks the immutable result envelope before reading
// its owned objects. Callers that subsequently lock source declarations use
// this path to preserve the cross-owner result-before-declaration order.
func (reader *Reader) ReadExactResultForUpdate(ctx context.Context, binding graphprojection.ResultBindingV2) (graphprojection.CompletedResultV2, error) {
	return reader.readExactResult(ctx, binding, true)
}

func (reader *Reader) readExactResult(ctx context.Context, binding graphprojection.ResultBindingV2, lock bool) (graphprojection.CompletedResultV2, error) {
	if err := binding.ValidateIdentity(); err != nil {
		return graphprojection.CompletedResultV2{}, err
	}
	var result graphprojection.CompletedResultV2
	var vertexCount, edgeCount int
	result.Binding.ProjectionResultID = binding.ProjectionResultID
	query := `
SELECT graph_view_id, source_owner_id, source_snapshot_id, projection_schema_id,
       projection_version, normalized_configuration_sha256, normalized_source_sha256,
       canonical_output_sha256, vertex_count, edge_count, result_json, published_at
  FROM graph_projection_results
 WHERE projection_result_id = $1
`
	if lock {
		query += ` FOR UPDATE`
	}
	err := reader.db.QueryRow(ctx, query, binding.ProjectionResultID).Scan(
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
	result.Vertices, err = reader.readVertices(ctx, binding.ProjectionResultID, graphprojection.MaximumResultVerticesV2)
	if err != nil {
		return graphprojection.CompletedResultV2{}, err
	}
	result.Edges, err = reader.readEdges(ctx, binding.ProjectionResultID, graphprojection.MaximumResultEdgesV2)
	if err != nil {
		return graphprojection.CompletedResultV2{}, err
	}
	if len(result.Vertices) != vertexCount || len(result.Edges) != edgeCount {
		return graphprojection.CompletedResultV2{}, graphprojection.ErrResultV2Invalid
	}
	if err := validateCompletedResult(ctx, result); err != nil {
		return graphprojection.CompletedResultV2{}, err
	}
	return result, nil
}

// Content-returning queries share exact result admission; raw row queries stay private.
func (reader *Reader) readVerifiedResult(ctx context.Context, id string) (graphprojection.CompletedResultV2, error) {
	binding, err := reader.ReadResultEnvelope(ctx, id)
	if err != nil {
		return graphprojection.CompletedResultV2{}, err
	}
	return reader.ReadExactResult(ctx, binding)
}

func (reader *Reader) ReadVertices(ctx context.Context, id string, maximum int) ([]graphprojection.ResultVertexV2, error) {
	if maximum < 1 || maximum > graphprojection.MaximumResultVerticesV2 {
		return nil, graphprojection.ErrResultV2Invalid
	}
	result, err := reader.readVerifiedResult(ctx, id)
	if err != nil {
		return nil, err
	}
	if len(result.Vertices) > maximum {
		return nil, graphprojection.ErrResultV2Invalid
	}
	return result.Vertices, nil
}

func (reader *Reader) ReadEdges(ctx context.Context, id string, maximum int) ([]graphprojection.ResultEdgeV2, error) {
	if maximum < 1 || maximum > graphprojection.MaximumResultEdgesV2 {
		return nil, graphprojection.ErrResultV2Invalid
	}
	result, err := reader.readVerifiedResult(ctx, id)
	if err != nil {
		return nil, err
	}
	if len(result.Edges) > maximum {
		return nil, graphprojection.ErrResultV2Invalid
	}
	return result.Edges, nil
}

func (reader *Reader) readVertices(ctx context.Context, projectionResultID string, maximum int) ([]graphprojection.ResultVertexV2, error) {
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

func (reader *Reader) readEdges(ctx context.Context, projectionResultID string, maximum int) ([]graphprojection.ResultEdgeV2, error) {
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
	verified, err := reader.readVerifiedResult(ctx, request.ProjectionResultID)
	if err != nil {
		return graphprojection.TraversalResultV2{}, err
	}
	vertices, edges := verified.Vertices, verified.Edges

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

type Cleaner struct {
	tx             pgx.Tx
	lockedResultID string
}

func NewCleaner(tx pgx.Tx) (*Cleaner, error) {
	if tx == nil {
		return nil, fmt.Errorf("graph projection reachability cleaner transaction is required")
	}
	return &Cleaner{tx: tx}, nil
}

func (cleaner *Cleaner) LockCleanupCandidate(
	ctx context.Context,
	sourceOwnerID string,
	after *graphprojection.ResultCleanupCandidateV2,
) (*graphprojection.ResultCleanupCandidateV2, error) {
	if cleaner == nil || cleaner.tx == nil || !sourceOwnerTokenPattern.MatchString(sourceOwnerID) || cleaner.lockedResultID != "" {
		return nil, graphprojection.ErrResultV2Invalid
	}
	query := `
SELECT result.projection_result_id, result.published_at
  FROM graph_projection_results result
 WHERE result.source_owner_id = $1`
	args := []any{sourceOwnerID}
	if after != nil {
		if !resultIDPattern.MatchString(after.ProjectionResultID) || after.PublishedAt.IsZero() {
			return nil, graphprojection.ErrResultV2Invalid
		}
		query += `
   AND (result.published_at, result.projection_result_id) > ($2, $3)`
		args = append(args, after.PublishedAt.UTC(), after.ProjectionResultID)
	}
	query += `
 ORDER BY result.published_at, result.projection_result_id
 FOR UPDATE SKIP LOCKED
 LIMIT 1`
	var candidate graphprojection.ResultCleanupCandidateV2
	err := cleaner.tx.QueryRow(ctx, query, args...).Scan(&candidate.ProjectionResultID, &candidate.PublishedAt)
	if errors.Is(err, pgx.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("lock Graph Projection cleanup candidate: %w", err)
	}
	candidate.PublishedAt = candidate.PublishedAt.UTC()
	cleaner.lockedResultID = candidate.ProjectionResultID
	return &candidate, nil
}

func (cleaner *Cleaner) DeleteLockedResult(ctx context.Context, projectionResultID string) (bool, error) {
	if cleaner == nil || cleaner.tx == nil || cleaner.lockedResultID != projectionResultID {
		return false, graphprojection.ErrResultV2Invalid
	}
	tag, err := cleaner.tx.Exec(ctx, `
DELETE FROM graph_projection_results
 WHERE projection_result_id = $1
`, projectionResultID)
	if err != nil {
		return false, fmt.Errorf("delete locked Graph Projection result: %w", err)
	}
	cleaner.lockedResultID = ""
	if tag.RowsAffected() > 1 {
		return false, graphprojection.ErrResultV2IdentityConflict
	}
	return tag.RowsAffected() == 1, nil
}

func validateCompletedResult(ctx context.Context, result graphprojection.CompletedResultV2) error {
	if result.PublishedAt.IsZero() {
		return graphprojection.ErrResultV2Invalid
	}
	return result.Validate(ctx)
}

func validateTraversalRequest(request graphprojection.TraversalRequestV2) error {
	if !resultIDPattern.MatchString(request.ProjectionResultID) || len(request.SeedVertexIDs) < 1 || len(request.SeedVertexIDs) > graphProjectionLimits.MaxTraversalSeedVertices ||
		request.MaximumDepth < 0 || request.MaximumDepth > graphProjectionLimits.MaxTraversalDepth || request.MaximumVertices < 1 ||
		request.MaximumVertices > graphprojection.MaximumResultVerticesV2 || request.MaximumEdges < 0 ||
		request.MaximumEdges > graphprojection.MaximumResultEdgesV2 || len(request.VertexKinds) > graphProjectionLimits.MaxTraversalKindFilters || len(request.EdgeKinds) > graphProjectionLimits.MaxTraversalKindFilters ||
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
