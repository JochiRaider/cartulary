package graphprojection

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/jackc/pgx/v5"

	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

var (
	ErrProjectionRunNotFound = errors.New("graphprojection: projection run not found")
	ErrGraphViewNotFound     = errors.New("graphprojection: graph view not found")
	ErrCursorInvalid         = errors.New("graphprojection: cursor invalid")
	ErrGraphViewUnavailable  = errors.New("graphprojection: graph view unavailable")
)

type Store struct {
	pool postgres.DB
}

type StoreProjectionOptions struct {
	ProjectionRunNonce string
	AcceptedAt         time.Time
	GeneratedAt        time.Time
	IdempotencyKey     string
}

type ListGraphViewsOptions struct {
	Limit       int
	CursorToken string
	Now         time.Time
}

type GraphViewSummary struct {
	GraphViewID            string
	GraphViewKey           string
	State                  GraphViewState
	LatestProjectionRunID  string
	LatestSourceSnapshotID string
	ProjectionVersion      string
	UpdatedAt              time.Time
	ValidationStatus       string
}

type TraverseRequest struct {
	GraphViewID     string
	ProjectionRunID string
	SeedVertexIDs   []string
	MaxDepth        int
	Direction       string
	VertexKinds     []string
	EdgeKinds       []string
}

type TraverseResult struct {
	GraphViewID     string
	ProjectionRunID string
	Vertices        []Vertex
	Edges           []Edge
	Metadata        map[string]any
}

func NewStore(pool postgres.DB) *Store {
	return &Store{pool: pool}
}

func (s *Store) CreateProjection(ctx context.Context, data []byte, options StoreProjectionOptions) (ProjectionRun, error) {
	run, err := Project(data, ProjectOptions{
		ProjectionRunNonce: options.ProjectionRunNonce,
		AcceptedAt:         options.AcceptedAt,
		GeneratedAt:        options.GeneratedAt,
	})
	if err != nil {
		return ProjectionRun{}, err
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return ProjectionRun{}, fmt.Errorf("begin graph projection create: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if options.IdempotencyKey != "" {
		replayed, replay, err := s.checkIdempotencyTx(ctx, tx, "create_projection", options.IdempotencyKey, run.ProjectionConfigDigest+"."+run.ProjectionSourceDigest)
		if err != nil {
			return ProjectionRun{}, err
		}
		if replay {
			return replayed, nil
		}
	}

	var previousRunID *string
	row := tx.QueryRow(ctx, `SELECT latest_projection_run_id FROM graph_projection_views WHERE graph_view_id = $1 AND state = 'available'`, run.GraphViewID)
	var previous string
	if err := row.Scan(&previous); err == nil && previous != "" {
		previousRunID = &previous
		run.PreviousProjectionRunID = previousRunID
		if run.GraphView != nil {
			run.GraphView.Metadata.PreviousProjectionRunID = previousRunID
		}
	} else if err != nil && !errors.Is(err, pgx.ErrNoRows) {
		return ProjectionRun{}, fmt.Errorf("load previous graph projection run: %w", err)
	}

	if previousRunID != nil && run.State == RunStateAvailable {
		if _, err := tx.Exec(ctx, `
UPDATE graph_projection_runs
   SET state = 'replaced',
       retention_expires_at = $2
 WHERE projection_run_id = $1
`, *previousRunID, run.CompletedAt.Add(time.Duration(run.Request.ProjectionConfig.RetentionPolicy.RetentionDurationSeconds)*time.Second)); err != nil {
			return ProjectionRun{}, fmt.Errorf("replace previous graph projection run: %w", err)
		}
	}

	if err := s.persistProjectionRunTx(ctx, tx, run); err != nil {
		return ProjectionRun{}, err
	}
	if options.IdempotencyKey != "" {
		if err := s.recordIdempotencyTx(ctx, tx, "create_projection", options.IdempotencyKey, run.ProjectionConfigDigest+"."+run.ProjectionSourceDigest, run); err != nil {
			return ProjectionRun{}, err
		}
	}
	if err := s.pruneRetentionTx(ctx, tx, run.GraphViewID, run.Request.ProjectionConfig.RetentionPolicy); err != nil {
		return ProjectionRun{}, err
	}
	if err := tx.Commit(ctx); err != nil {
		return ProjectionRun{}, fmt.Errorf("commit graph projection create: %w", err)
	}
	return run, nil
}

func (s *Store) RefreshProjection(ctx context.Context, data []byte, options StoreProjectionOptions) (ProjectionRun, error) {
	return s.CreateProjection(ctx, data, options)
}

func (s *Store) persistProjectionRunTx(ctx context.Context, tx pgx.Tx, run ProjectionRun) error {
	updatedAt := run.AcceptedAt
	if run.CompletedAt != nil {
		updatedAt = *run.CompletedAt
	}
	viewState := GraphViewStateCreating
	validationStatus := run.ValidationSummary.Status
	if validationStatus == "" {
		validationStatus = "pending"
	}
	switch run.State {
	case RunStateAvailable:
		viewState = GraphViewStateAvailable
	case RunStateFailed:
		viewState = GraphViewStateFailed
	}
	if _, err := tx.Exec(ctx, `
INSERT INTO graph_projection_views (
    graph_view_id,
    graph_view_key,
    state,
    latest_projection_run_id,
    latest_source_snapshot_id,
    projection_version,
    updated_at,
    validation_status
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
ON CONFLICT (graph_view_id) DO UPDATE
SET graph_view_key = EXCLUDED.graph_view_key,
    state = EXCLUDED.state,
    latest_projection_run_id = EXCLUDED.latest_projection_run_id,
    latest_source_snapshot_id = EXCLUDED.latest_source_snapshot_id,
    projection_version = EXCLUDED.projection_version,
    updated_at = EXCLUDED.updated_at,
    validation_status = EXCLUDED.validation_status
`, run.GraphViewID, run.Request.ProjectionConfig.GraphViewKey, string(viewState), run.ProjectionRunID, run.Request.SourceSnapshotID, run.Request.ProjectionConfig.ProjectionVersion, updatedAt, validationStatus); err != nil {
		return fmt.Errorf("upsert graph projection view: %w", err)
	}

	summaryJSON, err := json.Marshal(run.ValidationSummary)
	if err != nil {
		return fmt.Errorf("marshal validation summary: %w", err)
	}
	var graphJSON []byte
	if run.GraphView != nil {
		graphJSON, err = json.Marshal(run.GraphView)
		if err != nil {
			return fmt.Errorf("marshal graph view: %w", err)
		}
	}
	var completedAt any
	if run.CompletedAt != nil {
		completedAt = *run.CompletedAt
	}
	if _, err := tx.Exec(ctx, `
INSERT INTO graph_projection_runs (
    projection_run_id,
    graph_view_id,
    source_snapshot_id,
    projection_version,
    state,
    projection_run_nonce,
    projection_config_digest,
    projection_source_digest,
    accepted_at,
    completed_at,
    validation_summary_json,
    failure_reason,
    graph_view_json,
    retention_expires_at
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11::jsonb, $12, $13::jsonb, $14)
`, run.ProjectionRunID, run.GraphViewID, run.Request.SourceSnapshotID, run.Request.ProjectionConfig.ProjectionVersion, string(run.State), run.ProjectionRunNonce, run.ProjectionConfigDigest, run.ProjectionSourceDigest, run.AcceptedAt, completedAt, string(summaryJSON), nullString(run.FailureReason), nullJSON(graphJSON), run.RetentionExpiresAt); err != nil {
		return fmt.Errorf("insert graph projection run: %w", err)
	}
	if run.GraphView != nil {
		for _, vertex := range run.GraphView.Vertices {
			vertexJSON, err := json.Marshal(vertex)
			if err != nil {
				return fmt.Errorf("marshal graph vertex: %w", err)
			}
			if _, err := tx.Exec(ctx, `
INSERT INTO graph_projection_vertices (
    projection_run_id,
    graph_view_id,
    vertex_id,
    vertex_kind,
    sort_key,
    vertex_json
) VALUES ($1, $2, $3, $4, $5, $6::jsonb)
`, run.ProjectionRunID, run.GraphViewID, vertex.VertexID, vertex.VertexKind, vertex.SortKey, string(vertexJSON)); err != nil {
				return fmt.Errorf("insert graph projection vertex: %w", err)
			}
		}
		for _, edge := range run.GraphView.Edges {
			edgeJSON, err := json.Marshal(edge)
			if err != nil {
				return fmt.Errorf("marshal graph edge: %w", err)
			}
			if _, err := tx.Exec(ctx, `
INSERT INTO graph_projection_edges (
    projection_run_id,
    graph_view_id,
    edge_id,
    edge_kind,
    src_vertex_id,
    dst_vertex_id,
    direction,
    sort_key,
    edge_json
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9::jsonb)
`, run.ProjectionRunID, run.GraphViewID, edge.EdgeID, edge.EdgeKind, edge.SrcVertexID, edge.DstVertexID, edge.Direction, edge.SortKey, string(edgeJSON)); err != nil {
				return fmt.Errorf("insert graph projection edge: %w", err)
			}
		}
	}
	return nil
}

func (s *Store) GetProjectionRun(ctx context.Context, projectionRunID string) (ProjectionRun, error) {
	row := s.pool.QueryRow(ctx, `
SELECT graph_view_id,
       source_snapshot_id,
       projection_version,
       state,
       projection_run_nonce,
       projection_config_digest,
       projection_source_digest,
       accepted_at,
       completed_at,
       validation_summary_json,
       COALESCE(failure_reason, ''),
       graph_view_json,
       retention_expires_at
  FROM graph_projection_runs
 WHERE projection_run_id = $1
`, projectionRunID)
	var run ProjectionRun
	var state string
	var summaryJSON []byte
	var graphJSON []byte
	var failureReason string
	var completedAt *time.Time
	var retentionExpiresAt *time.Time
	if err := row.Scan(&run.GraphViewID, &run.Request.SourceSnapshotID, &run.Request.ProjectionConfig.ProjectionVersion, &state, &run.ProjectionRunNonce, &run.ProjectionConfigDigest, &run.ProjectionSourceDigest, &run.AcceptedAt, &completedAt, &summaryJSON, &failureReason, &graphJSON, &retentionExpiresAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ProjectionRun{}, ErrProjectionRunNotFound
		}
		return ProjectionRun{}, fmt.Errorf("get graph projection run: %w", err)
	}
	run.ProjectionRunID = projectionRunID
	run.State = RunState(state)
	run.CompletedAt = completedAt
	run.FailureReason = failureReason
	run.RetentionExpiresAt = retentionExpiresAt
	if len(summaryJSON) > 0 {
		if err := json.Unmarshal(summaryJSON, &run.ValidationSummary); err != nil {
			return ProjectionRun{}, fmt.Errorf("decode validation summary: %w", err)
		}
	}
	if len(graphJSON) > 0 {
		var graphView GraphView
		if err := json.Unmarshal(graphJSON, &graphView); err != nil {
			return ProjectionRun{}, fmt.Errorf("decode graph view: %w", err)
		}
		run.GraphView = &graphView
	}
	return run, nil
}

func (s *Store) GetGraphView(ctx context.Context, graphViewID string, projectionRunID string) (GraphView, error) {
	if projectionRunID == "" {
		row := s.pool.QueryRow(ctx, `SELECT latest_projection_run_id FROM graph_projection_views WHERE graph_view_id = $1`, graphViewID)
		if err := row.Scan(&projectionRunID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return GraphView{}, ErrGraphViewNotFound
			}
			return GraphView{}, fmt.Errorf("load latest graph projection run: %w", err)
		}
	}
	run, err := s.GetProjectionRun(ctx, projectionRunID)
	if err != nil {
		return GraphView{}, err
	}
	if run.GraphView == nil || (run.State != RunStateAvailable && run.State != RunStateReplaced) {
		return GraphView{}, ErrGraphViewUnavailable
	}
	return *run.GraphView, nil
}

func (s *Store) ListGraphViews(ctx context.Context, options ListGraphViewsOptions) ([]GraphViewSummary, string, error) {
	limit := options.Limit
	if limit <= 0 {
		limit = 100
	}
	if limit > 1000 {
		limit = 1000
	}
	after := ""
	now := options.Now.UTC()
	if now.IsZero() {
		now = time.Now().UTC()
	}
	if options.CursorToken != "" {
		cursor, err := decodeCursor(options.CursorToken)
		if err != nil || cursor.Operation != "list_graph_views" || now.After(cursor.IssuedAt.Add(15*time.Minute)) {
			return nil, "", ErrCursorInvalid
		}
		after = cursor.AfterGraphViewID
	}
	rows, err := s.pool.Query(ctx, `
SELECT graph_view_id,
       graph_view_key,
       state,
       COALESCE(latest_projection_run_id, ''),
       COALESCE(latest_source_snapshot_id, ''),
       COALESCE(projection_version, ''),
       updated_at,
       validation_status
  FROM graph_projection_views
 WHERE graph_view_id > $1
 ORDER BY graph_view_id ASC
 LIMIT $2
`, after, limit+1)
	if err != nil {
		return nil, "", fmt.Errorf("list graph views: %w", err)
	}
	defer rows.Close()
	summaries := []GraphViewSummary{}
	for rows.Next() {
		var summary GraphViewSummary
		var state string
		if err := rows.Scan(&summary.GraphViewID, &summary.GraphViewKey, &state, &summary.LatestProjectionRunID, &summary.LatestSourceSnapshotID, &summary.ProjectionVersion, &summary.UpdatedAt, &summary.ValidationStatus); err != nil {
			return nil, "", fmt.Errorf("scan graph view summary: %w", err)
		}
		summary.State = GraphViewState(state)
		summaries = append(summaries, summary)
	}
	if rows.Err() != nil {
		return nil, "", rows.Err()
	}
	nextCursor := ""
	if len(summaries) > limit {
		summaries = summaries[:limit]
		nextCursor = encodeCursor(listCursor{Operation: "list_graph_views", AfterGraphViewID: summaries[len(summaries)-1].GraphViewID, IssuedAt: now})
	}
	return summaries, nextCursor, nil
}

func (s *Store) Traverse(ctx context.Context, request TraverseRequest) (TraverseResult, error) {
	graphView, err := s.GetGraphView(ctx, request.GraphViewID, request.ProjectionRunID)
	if err != nil {
		return TraverseResult{}, err
	}
	maxDepth := request.MaxDepth
	if maxDepth <= 0 || maxDepth > 16 {
		maxDepth = 16
	}
	direction := request.Direction
	if direction == "" {
		direction = "out"
	}
	vertexKinds := stringSet(request.VertexKinds)
	edgeKinds := stringSet(request.EdgeKinds)
	vertexByID := map[string]Vertex{}
	for _, vertex := range graphView.Vertices {
		vertexByID[vertex.VertexID] = vertex
	}
	visited := map[string]bool{}
	frontier := []string{}
	for _, seed := range request.SeedVertexIDs {
		if _, ok := vertexByID[seed]; ok {
			frontier = append(frontier, seed)
			visited[seed] = true
		}
	}
	selectedEdges := map[string]Edge{}
	for depth := 0; depth < maxDepth && len(frontier) > 0; depth++ {
		next := []string{}
		for _, current := range frontier {
			for _, edge := range graphView.Edges {
				if len(edgeKinds) > 0 && !edgeKinds[edge.EdgeKind] {
					continue
				}
				var candidate string
				if (direction == "out" || direction == "both") && edge.SrcVertexID == current {
					candidate = edge.DstVertexID
				}
				if candidate == "" && (direction == "in" || direction == "both") && edge.DstVertexID == current {
					candidate = edge.SrcVertexID
				}
				if candidate == "" {
					continue
				}
				vertex := vertexByID[candidate]
				if len(vertexKinds) > 0 && !vertexKinds[vertex.VertexKind] {
					continue
				}
				selectedEdges[edge.EdgeID] = edge
				if !visited[candidate] {
					visited[candidate] = true
					next = append(next, candidate)
				}
			}
		}
		sort.Strings(next)
		frontier = next
	}
	vertices := []Vertex{}
	for vertexID := range visited {
		vertices = append(vertices, vertexByID[vertexID])
	}
	edges := []Edge{}
	for _, edge := range selectedEdges {
		edges = append(edges, edge)
	}
	sortVertices(vertices)
	sortEdges(edges)
	return TraverseResult{GraphViewID: graphView.GraphViewID, ProjectionRunID: graphView.ProjectionRunID, Vertices: vertices, Edges: edges, Metadata: map[string]any{}}, nil
}

func (s *Store) InvalidateGraphView(ctx context.Context, graphViewID, reasonCode, requestedAt, requestedBy string) (InvalidationSummary, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return InvalidationSummary{}, fmt.Errorf("begin graph projection invalidation: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	rows, err := tx.Query(ctx, `SELECT projection_run_id FROM graph_projection_runs WHERE graph_view_id = $1 AND state IN ('available', 'replaced') ORDER BY projection_run_id ASC`, graphViewID)
	if err != nil {
		return InvalidationSummary{}, fmt.Errorf("list invalidation targets: %w", err)
	}
	runIDs := []string{}
	for rows.Next() {
		var runID string
		if err := rows.Scan(&runID); err != nil {
			rows.Close()
			return InvalidationSummary{}, err
		}
		runIDs = append(runIDs, runID)
	}
	rows.Close()
	if len(runIDs) == 0 {
		return InvalidationSummary{}, ErrGraphViewNotFound
	}
	summary := InvalidationSummary{GraphViewID: graphViewID, InvalidatedRunIDs: runIDs, ReasonCode: reasonCode, RequestedAt: requestedAt, RequestedBy: requestedBy}
	summaryJSON, _ := json.Marshal(summary)
	if _, err := tx.Exec(ctx, `UPDATE graph_projection_runs SET state = 'invalidated', invalidation_json = $2::jsonb WHERE graph_view_id = $1 AND state IN ('available', 'replaced')`, graphViewID, string(summaryJSON)); err != nil {
		return InvalidationSummary{}, fmt.Errorf("invalidate graph projection runs: %w", err)
	}
	if _, err := tx.Exec(ctx, `UPDATE graph_projection_views SET state = 'invalidated', validation_status = 'invalidated', updated_at = now() WHERE graph_view_id = $1`, graphViewID); err != nil {
		return InvalidationSummary{}, fmt.Errorf("invalidate graph projection view: %w", err)
	}
	if err := tx.Commit(ctx); err != nil {
		return InvalidationSummary{}, fmt.Errorf("commit graph projection invalidation: %w", err)
	}
	return summary, nil
}

func (s *Store) pruneRetentionTx(ctx context.Context, tx pgx.Tx, graphViewID string, policy RetentionPolicy) error {
	if policy.RetentionCount > 0 {
		if _, err := tx.Exec(ctx, `
DELETE FROM graph_projection_runs
 WHERE projection_run_id IN (
       SELECT projection_run_id
         FROM graph_projection_runs
        WHERE graph_view_id = $1
          AND state = 'replaced'
        ORDER BY completed_at DESC NULLS LAST, projection_run_id DESC
        OFFSET $2
 )
`, graphViewID, policy.RetentionCount); err != nil {
			return fmt.Errorf("prune replaced graph projection runs: %w", err)
		}
	}
	if policy.FailedRetentionCount > 0 {
		if _, err := tx.Exec(ctx, `
DELETE FROM graph_projection_runs
 WHERE projection_run_id IN (
       SELECT projection_run_id
         FROM graph_projection_runs
        WHERE graph_view_id = $1
          AND state = 'failed'
        ORDER BY completed_at DESC NULLS LAST, projection_run_id DESC
        OFFSET $2
 )
`, graphViewID, policy.FailedRetentionCount); err != nil {
			return fmt.Errorf("prune failed graph projection runs: %w", err)
		}
	}
	return nil
}

func (s *Store) checkIdempotencyTx(ctx context.Context, tx pgx.Tx, operation, key, fingerprint string) (ProjectionRun, bool, error) {
	row := tx.QueryRow(ctx, `SELECT request_fingerprint, projection_run_id FROM graph_projection_idempotency WHERE operation = $1 AND idempotency_key = $2`, operation, key)
	var existingFingerprint string
	var projectionRunID string
	if err := row.Scan(&existingFingerprint, &projectionRunID); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ProjectionRun{}, false, nil
		}
		return ProjectionRun{}, false, fmt.Errorf("load graph projection idempotency: %w", err)
	}
	if existingFingerprint != fingerprint {
		return ProjectionRun{}, false, &OperationError{Code: "invalid_projection_request", ReasonCode: "idempotency_conflict", Details: map[string]any{"reason_code": "idempotency_conflict"}}
	}
	run, err := s.GetProjectionRun(ctx, projectionRunID)
	return run, true, err
}

func (s *Store) recordIdempotencyTx(ctx context.Context, tx pgx.Tx, operation, key, fingerprint string, run ProjectionRun) error {
	responseJSON, err := json.Marshal(map[string]any{"projection_run_id": run.ProjectionRunID, "graph_view_id": run.GraphViewID, "status": run.State})
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
INSERT INTO graph_projection_idempotency (
    operation,
    idempotency_key,
    request_fingerprint,
    graph_view_id,
    projection_run_id,
    response_json,
    created_at,
    expires_at
) VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7, $8)
`, operation, key, fingerprint, run.GraphViewID, run.ProjectionRunID, string(responseJSON), run.AcceptedAt, run.AcceptedAt.Add(24*time.Hour))
	if err != nil {
		return fmt.Errorf("record graph projection idempotency: %w", err)
	}
	return nil
}

type listCursor struct {
	Operation        string    `json:"operation"`
	AfterGraphViewID string    `json:"after_graph_view_id"`
	IssuedAt         time.Time `json:"issued_at"`
}

func encodeCursor(cursor listCursor) string {
	data, _ := json.Marshal(cursor)
	return base64.RawURLEncoding.EncodeToString(data)
}

func decodeCursor(token string) (listCursor, error) {
	data, err := base64.RawURLEncoding.DecodeString(token)
	if err != nil {
		return listCursor{}, err
	}
	var cursor listCursor
	if err := json.Unmarshal(data, &cursor); err != nil {
		return listCursor{}, err
	}
	return cursor, nil
}

func nullString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func nullJSON(value []byte) any {
	if len(value) == 0 {
		return nil
	}
	return string(value)
}
