package graphprojection

import (
	"context"
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
	ErrVertexNotFound        = errors.New("graphprojection: vertex not found")
	ErrEdgeNotFound          = errors.New("graphprojection: edge not found")
)

type Store struct {
	pool        postgres.DB
	cursorCodec *graphCursorCodec
	now         func() time.Time
}

type StoreProjectionOptions struct {
	ProjectionRunNonce string
	AcceptedAt         time.Time
	GeneratedAt        time.Time
	IdempotencyKey     string
}

type ListGraphViewsOptions struct {
	Limit                 *int
	CursorToken           string
	Now                   time.Time
	QueryShapeDigest      string
	VisibilityScopeDigest string
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
	MaxDepth        *int
	Direction       string
	VertexKinds     []string
	EdgeKinds       []string
}

type TraverseResult struct {
	GraphViewID          string
	ProjectionRunID      string
	SeedVertexIDs        []string
	OmittedSeedVertexIDs []string
	Vertices             []Vertex
	Edges                []Edge
	Metadata             map[string]any
}

func NewStore(pool postgres.DB) *Store {
	key, err := randomCursorKey()
	if err != nil {
		panic(fmt.Sprintf("create graph projection cursor key: %v", err))
	}
	store, err := newStoreWithCursorKeyAndClock(pool, key, func() time.Time { return time.Now().UTC() })
	if err != nil {
		panic(err)
	}
	return store
}

func NewStoreWithCursorKey(pool postgres.DB, key []byte) (*Store, error) {
	return newStoreWithCursorKeyAndClock(pool, key, func() time.Time { return time.Now().UTC() })
}

func NewStoreWithClock(pool postgres.DB, now func() time.Time) *Store {
	key, err := randomCursorKey()
	if err != nil {
		panic(fmt.Sprintf("create graph projection cursor key: %v", err))
	}
	store, err := newStoreWithCursorKeyAndClock(pool, key, now)
	if err != nil {
		panic(err)
	}
	return store
}

func newStoreWithCursorKeyAndClock(pool postgres.DB, key []byte, now func() time.Time) (*Store, error) {
	codec, err := newGraphCursorCodec(key)
	if err != nil {
		return nil, err
	}
	if now == nil {
		now = func() time.Time { return time.Now().UTC() }
	}
	return &Store{pool: pool, cursorCodec: codec, now: now}, nil
}

func (s *Store) CreateProjection(ctx context.Context, data []byte, options StoreProjectionOptions) (ProjectionRun, error) {
	return s.projectRetained(ctx, "create_projection", data, options)
}

func (s *Store) RefreshProjection(ctx context.Context, data []byte, options StoreProjectionOptions) (ProjectionRun, error) {
	return s.projectRetained(ctx, "refresh_projection", data, options)
}

func (s *Store) projectRetained(ctx context.Context, operation string, data []byte, options StoreProjectionOptions) (ProjectionRun, error) {
	run, err := admitProjectionInput(data, admitOptions{
		ProjectionRunNonce: options.ProjectionRunNonce,
		AcceptedAt:         options.AcceptedAt,
	})
	if err != nil {
		return ProjectionRun{}, err
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return ProjectionRun{}, fmt.Errorf("begin graph projection admission: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if options.IdempotencyKey != "" {
		fingerprint, err := retainedReplayFingerprint(operation, run)
		if err != nil {
			return ProjectionRun{}, err
		}
		replayed, replay, err := s.checkIdempotencyTx(ctx, tx, operation, run.GraphViewID, options.IdempotencyKey, fingerprint, run.AcceptedAt)
		if err != nil {
			return ProjectionRun{}, err
		}
		if replay {
			return replayed, nil
		}
	}

	var currentViewState string
	var currentLatestRunID string
	var currentSelectedRunID string
	viewErr := tx.QueryRow(ctx, `SELECT state, COALESCE(latest_projection_run_id, ''), COALESCE(selected_projection_run_id, latest_projection_run_id, '') FROM graph_projection_views WHERE graph_view_id = $1 FOR UPDATE`, run.GraphViewID).Scan(&currentViewState, &currentLatestRunID, &currentSelectedRunID)
	viewExists := viewErr == nil
	if viewErr != nil && !errors.Is(viewErr, pgx.ErrNoRows) {
		return ProjectionRun{}, fmt.Errorf("load graph projection lifecycle state: %w", viewErr)
	}
	if operation == "create_projection" && viewExists {
		switch GraphViewState(currentViewState) {
		case GraphViewStateCreating, GraphViewStateRefreshing:
			return ProjectionRun{}, &OperationError{Code: "operation_conflict", ReasonCode: "run_already_active", Details: map[string]any{"operation": operation, "reason_code": "run_already_active", "active_projection_run_id": currentLatestRunID}}
		case GraphViewStateAvailable, GraphViewStateInvalidated:
			return ProjectionRun{}, &OperationError{Code: "invalid_operation", ReasonCode: "graph_view_already_exists", Details: map[string]any{"operation": operation, "reason_code": "graph_view_already_exists"}}
		}
	}
	if operation == "refresh_projection" {
		if !viewExists {
			return ProjectionRun{}, ErrGraphViewNotFound
		}
		switch GraphViewState(currentViewState) {
		case GraphViewStateCreating, GraphViewStateRefreshing:
			return ProjectionRun{}, &OperationError{Code: "operation_conflict", ReasonCode: "run_already_active", Details: map[string]any{"operation": operation, "reason_code": "run_already_active", "active_projection_run_id": currentLatestRunID}}
		case GraphViewStateFailed:
			return ProjectionRun{}, &OperationError{Code: "invalid_operation", ReasonCode: "no_consumable_prior_run", Details: map[string]any{"operation": operation, "reason_code": "no_consumable_prior_run"}}
		}
	}

	var previousRunID *string
	if operation == "refresh_projection" && currentSelectedRunID != "" {
		previousRunID = &currentSelectedRunID
		run.PreviousProjectionRunID = previousRunID
	}
	admissionViewState := GraphViewStateCreating
	if operation == "refresh_projection" {
		admissionViewState = GraphViewStateRefreshing
	}
	if err := s.persistAcceptedRunTx(ctx, tx, run, admissionViewState, currentSelectedRunID); err != nil {
		return ProjectionRun{}, err
	}
	if options.IdempotencyKey != "" {
		fingerprint, err := retainedReplayFingerprint(operation, run)
		if err != nil {
			return ProjectionRun{}, err
		}
		if err := s.recordIdempotencyTx(ctx, tx, operation, run.GraphViewID, options.IdempotencyKey, fingerprint, run); err != nil {
			return ProjectionRun{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return ProjectionRun{}, fmt.Errorf("commit graph projection admission: %w", err)
	}

	startedAt := run.AcceptedAt.UTC()
	run.StartedAt = &startedAt
	run.State = RunStateComputing
	if _, err := s.pool.Exec(ctx, `UPDATE graph_projection_runs SET state = 'computing', started_at = $2 WHERE projection_run_id = $1 AND state = 'accepted'`, run.ProjectionRunID, startedAt); err != nil {
		return ProjectionRun{}, fmt.Errorf("start graph projection run: %w", err)
	}

	result := projectAdmittedRun(run, projectOptions{GeneratedAt: options.GeneratedAt, PreviousProjectionRunID: previousRunID})
	terminalTx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return ProjectionRun{}, fmt.Errorf("begin graph projection publication: %w", err)
	}
	defer func() { _ = terminalTx.Rollback(ctx) }()
	var terminalViewState string
	var terminalSelectedRunID string
	var invalidationJSON []byte
	if err := terminalTx.QueryRow(ctx, `SELECT state, COALESCE(selected_projection_run_id, ''), COALESCE(invalidation_json, 'null'::jsonb) FROM graph_projection_views WHERE graph_view_id = $1 FOR UPDATE`, result.GraphViewID).Scan(&terminalViewState, &terminalSelectedRunID, &invalidationJSON); err != nil {
		return ProjectionRun{}, err
	}
	wasInvalidated := GraphViewState(terminalViewState) == GraphViewStateInvalidated || string(invalidationJSON) != "null"
	if wasInvalidated && result.State == RunStateAvailable {
		result.State = RunStateInvalidated
		result.InvalidatedAt = result.CompletedAt
		result.PreviousProjectionRunID = nil
		if result.GraphView != nil {
			result.GraphView.State = RunStateInvalidated
			result.GraphView.Metadata.PreviousProjectionRunID = nil
			var invalidation InvalidationSummary
			if string(invalidationJSON) != "null" && json.Unmarshal(invalidationJSON, &invalidation) == nil {
				result.GraphView.Metadata.Invalidation = &invalidation
			}
		}
	}
	publicationPriorState := GraphViewState(terminalViewState)
	if wasInvalidated {
		publicationPriorState = GraphViewStateInvalidated
	}
	if err := s.publishProjectionRunTx(ctx, terminalTx, result, operation, publicationPriorState, terminalSelectedRunID); err != nil {
		return ProjectionRun{}, err
	}
	queryReceivedAt := result.AcceptedAt
	if result.CompletedAt != nil {
		queryReceivedAt = *result.CompletedAt
	}
	if err := s.pruneRetentionTx(ctx, terminalTx, result.GraphViewID, result.Request.ProjectionConfig.RetentionPolicy, queryReceivedAt); err != nil {
		return ProjectionRun{}, err
	}
	if err := terminalTx.Commit(ctx); err != nil {
		return ProjectionRun{}, fmt.Errorf("commit graph projection publication: %w", err)
	}
	return result, nil
}

func (s *Store) persistAcceptedRunTx(ctx context.Context, tx pgx.Tx, run ProjectionRun, viewState GraphViewState, selectedRunID string) error {
	retentionPolicyJSON, err := json.Marshal(run.Request.ProjectionConfig.RetentionPolicy)
	if err != nil {
		return fmt.Errorf("marshal graph projection retention policy: %w", err)
	}
	if _, err := tx.Exec(ctx, `
INSERT INTO graph_projection_views (
    graph_view_id,
    graph_view_key,
    state,
    latest_projection_run_id,
    selected_projection_run_id,
    latest_source_snapshot_id,
    projection_version,
    updated_at,
    validation_status
) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, 'pending')
ON CONFLICT (graph_view_id) DO UPDATE
SET graph_view_key = EXCLUDED.graph_view_key,
    state = EXCLUDED.state,
    latest_projection_run_id = EXCLUDED.latest_projection_run_id,
    selected_projection_run_id = EXCLUDED.selected_projection_run_id,
    latest_source_snapshot_id = EXCLUDED.latest_source_snapshot_id,
    projection_version = EXCLUDED.projection_version,
    updated_at = EXCLUDED.updated_at,
    validation_status = EXCLUDED.validation_status
`, run.GraphViewID, run.Request.ProjectionConfig.GraphViewKey, string(viewState), run.ProjectionRunID, nullString(selectedRunID), run.Request.SourceSnapshotID, run.Request.ProjectionConfig.ProjectionVersion, run.AcceptedAt); err != nil {
		return fmt.Errorf("upsert accepted graph projection view: %w", err)
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
    validation_summary_json,
    retention_policy_json
) VALUES ($1, $2, $3, $4, 'accepted', $5, $6, $7, $8, 'null'::jsonb, $9::jsonb)
`, run.ProjectionRunID, run.GraphViewID, run.Request.SourceSnapshotID, run.Request.ProjectionConfig.ProjectionVersion, run.ProjectionRunNonce, run.ProjectionConfigDigest, run.ProjectionSourceDigest, run.AcceptedAt, string(retentionPolicyJSON)); err != nil {
		return fmt.Errorf("insert accepted graph projection run: %w", err)
	}
	return nil
}

func (s *Store) publishProjectionRunTx(ctx context.Context, tx pgx.Tx, run ProjectionRun, operation string, priorViewState GraphViewState, selectedRunID string) error {
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
	var retentionExpiresAt any
	if run.State == RunStateFailed && operation == "refresh_projection" && run.CompletedAt != nil {
		retentionExpiresAt = run.CompletedAt.Add(time.Duration(run.Request.ProjectionConfig.RetentionPolicy.FailedRetentionDurationSecs) * time.Second)
	}
	if _, err := tx.Exec(ctx, `
UPDATE graph_projection_runs
   SET state = $2,
       projection_output_digest = $3,
       generated_at = $4,
       completed_at = $5,
       invalidated_at = $6,
       validation_summary_json = $7::jsonb,
       failure_reason = $8,
       graph_view_json = $9::jsonb,
	   retention_expires_at = $10
 WHERE projection_run_id = $1
   AND state = 'computing'
`, run.ProjectionRunID, string(run.State), nullString(run.ProjectionOutputDigest), run.GeneratedAt, run.CompletedAt, run.InvalidatedAt, string(summaryJSON), nullString(run.FailureReason), nullJSON(graphJSON), retentionExpiresAt); err != nil {
		return fmt.Errorf("publish graph projection run: %w", err)
	}

	updatedAt := run.AcceptedAt
	if run.CompletedAt != nil {
		updatedAt = *run.CompletedAt
	}
	viewState := GraphViewStateFailed
	viewLatestRunID := run.ProjectionRunID
	viewSelectedRunID := ""
	validationStatus := run.ValidationSummary.Status
	if validationStatus == "" {
		validationStatus = "failed"
	}
	preserveInvalidation := false
	switch run.State {
	case RunStateAvailable:
		viewState = GraphViewStateAvailable
		viewSelectedRunID = run.ProjectionRunID
		if selectedRunID != "" && selectedRunID != run.ProjectionRunID {
			retentionExpiresAt := updatedAt.Add(time.Duration(run.Request.ProjectionConfig.RetentionPolicy.RetentionDurationSeconds) * time.Second)
			if _, err := tx.Exec(ctx, `UPDATE graph_projection_runs SET state = 'replaced', replaced_at = $2, retention_expires_at = $3 WHERE projection_run_id = $1 AND state = 'available'`, selectedRunID, updatedAt, retentionExpiresAt); err != nil {
				return fmt.Errorf("replace prior graph projection run: %w", err)
			}
		}
	case RunStateInvalidated:
		viewState = GraphViewStateInvalidated
		viewSelectedRunID = run.ProjectionRunID
		validationStatus = "invalidated"
		preserveInvalidation = true
	case RunStateFailed:
		if operation == "refresh_projection" && selectedRunID != "" {
			viewLatestRunID = selectedRunID
			viewSelectedRunID = selectedRunID
			if priorViewState == GraphViewStateInvalidated {
				viewState = GraphViewStateInvalidated
				validationStatus = "invalidated"
				preserveInvalidation = true
			} else {
				viewState = GraphViewStateAvailable
				validationStatus = "valid"
			}
		}
	}
	if _, err := tx.Exec(ctx, `
UPDATE graph_projection_views
   SET state = $2,
       latest_projection_run_id = $3,
       selected_projection_run_id = $4,
       latest_source_snapshot_id = CASE WHEN $4::text IS NULL THEN $5 ELSE (SELECT source_snapshot_id FROM graph_projection_runs WHERE projection_run_id = $4) END,
       projection_version = CASE WHEN $4::text IS NULL THEN $6 ELSE (SELECT projection_version FROM graph_projection_runs WHERE projection_run_id = $4) END,
       updated_at = $7,
       validation_status = $8,
       invalidation_json = CASE WHEN $9 THEN invalidation_json ELSE NULL END
 WHERE graph_view_id = $1
`, run.GraphViewID, string(viewState), viewLatestRunID, nullString(viewSelectedRunID), run.Request.SourceSnapshotID, run.Request.ProjectionConfig.ProjectionVersion, updatedAt, validationStatus, preserveInvalidation); err != nil {
		return fmt.Errorf("publish graph projection view: %w", err)
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
	queryReceivedAt := s.now().UTC()
	if _, err := s.pool.Exec(ctx, `DELETE FROM graph_projection_runs WHERE projection_run_id = $1 AND retention_expires_at IS NOT NULL AND $2::timestamptz >= retention_expires_at`, projectionRunID, queryReceivedAt); err != nil {
		return ProjectionRun{}, fmt.Errorf("expire graph projection run: %w", err)
	}
	row := s.pool.QueryRow(ctx, `
SELECT graph_view_id,
       source_snapshot_id,
       projection_version,
       state,
       projection_run_nonce,
       projection_config_digest,
       projection_source_digest,
       projection_output_digest,
       accepted_at,
	   started_at,
	   generated_at,
       completed_at,
	   replaced_at,
	   invalidated_at,
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
	var projectionOutputDigest *string
	var completedAt *time.Time
	var retentionExpiresAt *time.Time
	if err := row.Scan(&run.GraphViewID, &run.Request.SourceSnapshotID, &run.Request.ProjectionConfig.ProjectionVersion, &state, &run.ProjectionRunNonce, &run.ProjectionConfigDigest, &run.ProjectionSourceDigest, &projectionOutputDigest, &run.AcceptedAt, &run.StartedAt, &run.GeneratedAt, &completedAt, &run.ReplacedAt, &run.InvalidatedAt, &summaryJSON, &failureReason, &graphJSON, &retentionExpiresAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ProjectionRun{}, ErrProjectionRunNotFound
		}
		return ProjectionRun{}, fmt.Errorf("get graph projection run: %w", err)
	}
	run.ProjectionRunID = projectionRunID
	run.State = RunState(state)
	run.CompletedAt = completedAt
	if projectionOutputDigest != nil {
		run.ProjectionOutputDigest = *projectionOutputDigest
	}
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
		row := s.pool.QueryRow(ctx, `SELECT COALESCE(selected_projection_run_id, latest_projection_run_id) FROM graph_projection_views WHERE graph_view_id = $1`, graphViewID)
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
	if run.GraphViewID != graphViewID {
		return GraphView{}, ErrProjectionRunNotFound
	}
	return *run.GraphView, nil
}

func (s *Store) GetVertex(ctx context.Context, graphViewID, projectionRunID, vertexID string) (Vertex, error) {
	projectionRunID, err := s.resolveReadableRunID(ctx, graphViewID, projectionRunID)
	if err != nil {
		return Vertex{}, err
	}
	var payload []byte
	if err := s.pool.QueryRow(ctx, `SELECT vertex_json FROM graph_projection_vertices WHERE graph_view_id = $1 AND projection_run_id = $2 AND vertex_id = $3`, graphViewID, projectionRunID, vertexID).Scan(&payload); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Vertex{}, ErrVertexNotFound
		}
		return Vertex{}, fmt.Errorf("get graph projection vertex: %w", err)
	}
	var vertex Vertex
	if err := json.Unmarshal(payload, &vertex); err != nil {
		return Vertex{}, fmt.Errorf("decode graph projection vertex: %w", err)
	}
	return vertex, nil
}

func (s *Store) GetEdge(ctx context.Context, graphViewID, projectionRunID, edgeID string) (Edge, error) {
	projectionRunID, err := s.resolveReadableRunID(ctx, graphViewID, projectionRunID)
	if err != nil {
		return Edge{}, err
	}
	var payload []byte
	if err := s.pool.QueryRow(ctx, `SELECT edge_json FROM graph_projection_edges WHERE graph_view_id = $1 AND projection_run_id = $2 AND edge_id = $3`, graphViewID, projectionRunID, edgeID).Scan(&payload); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Edge{}, ErrEdgeNotFound
		}
		return Edge{}, fmt.Errorf("get graph projection edge: %w", err)
	}
	var edge Edge
	if err := json.Unmarshal(payload, &edge); err != nil {
		return Edge{}, fmt.Errorf("decode graph projection edge: %w", err)
	}
	return edge, nil
}

func (s *Store) resolveReadableRunID(ctx context.Context, graphViewID, projectionRunID string) (string, error) {
	if projectionRunID == "" {
		if err := s.pool.QueryRow(ctx, `SELECT COALESCE(selected_projection_run_id, latest_projection_run_id) FROM graph_projection_views WHERE graph_view_id = $1`, graphViewID).Scan(&projectionRunID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return "", ErrGraphViewNotFound
			}
			return "", fmt.Errorf("resolve selected graph projection run: %w", err)
		}
	}
	var state string
	if err := s.pool.QueryRow(ctx, `SELECT state FROM graph_projection_runs WHERE graph_view_id = $1 AND projection_run_id = $2`, graphViewID, projectionRunID).Scan(&state); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", ErrProjectionRunNotFound
		}
		return "", fmt.Errorf("resolve graph projection run: %w", err)
	}
	if RunState(state) != RunStateAvailable && RunState(state) != RunStateReplaced {
		return "", ErrGraphViewUnavailable
	}
	return projectionRunID, nil
}

func (s *Store) ListGraphViews(ctx context.Context, options ListGraphViewsOptions) ([]GraphViewSummary, string, error) {
	limit := 100
	if options.Limit != nil {
		limit = *options.Limit
		if limit < 1 || limit > 1000 {
			return nil, "", &OperationError{Code: "invalid_argument", ReasonCode: "invalid_limit", Field: "limit", Details: map[string]any{"field": "limit"}}
		}
	}
	after := ""
	now := options.Now.UTC()
	if now.IsZero() {
		now = time.Now().UTC()
	}
	if options.CursorToken != "" {
		if len(options.CursorToken) > 4096 {
			return nil, "", cursorInvalid("cursor_token_too_long")
		}
		cursor, err := s.cursorCodec.decode(options.CursorToken)
		if err != nil {
			return nil, "", cursorInvalid("malformed")
		}
		if cursor.Operation != "list_graph_views" || cursor.QueryShapeDigest != options.QueryShapeDigest || cursor.VisibilityScopeDigest != options.VisibilityScopeDigest {
			return nil, "", cursorInvalid("wrong_query_shape")
		}
		if !now.Before(cursor.IssuedAt.Add(15 * time.Minute)) {
			return nil, "", cursorInvalid("expired")
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
		encoded, err := s.cursorCodec.encode(listCursor{Operation: "list_graph_views", AfterGraphViewID: summaries[len(summaries)-1].GraphViewID, IssuedAt: now, QueryShapeDigest: options.QueryShapeDigest, VisibilityScopeDigest: options.VisibilityScopeDigest})
		if err != nil {
			return nil, "", err
		}
		nextCursor = encoded
	}
	return summaries, nextCursor, nil
}

func (s *Store) Traverse(ctx context.Context, request TraverseRequest) (TraverseResult, error) {
	graphView, err := s.GetGraphView(ctx, request.GraphViewID, request.ProjectionRunID)
	if err != nil {
		return TraverseResult{}, err
	}
	maxDepth := 1
	if request.MaxDepth != nil {
		maxDepth = *request.MaxDepth
	}
	if maxDepth < 0 || maxDepth > 16 {
		return TraverseResult{}, &OperationError{Code: "invalid_argument", ReasonCode: "invalid_max_depth", Details: map[string]any{"field": "max_depth"}}
	}
	direction := request.Direction
	if direction == "" {
		direction = "outbound"
	}
	if direction != "outbound" && direction != "inbound" && direction != "any" {
		return TraverseResult{}, &OperationError{Code: "invalid_argument", ReasonCode: "invalid_direction", Details: map[string]any{"field": "direction"}}
	}
	if len(request.SeedVertexIDs) > 1024 || len(request.VertexKinds) > 1024 || len(request.EdgeKinds) > 1024 {
		return TraverseResult{}, &OperationError{Code: "invalid_argument", ReasonCode: "collection_limit_exceeded", Details: map[string]any{"field": "traverse"}}
	}
	if hasDuplicateStrings(request.VertexKinds) || hasDuplicateStrings(request.EdgeKinds) {
		return TraverseResult{}, &OperationError{Code: "invalid_argument", ReasonCode: "duplicate_filter_value", Details: map[string]any{"field": "traverse"}}
	}
	vertexKinds := stringSet(request.VertexKinds)
	edgeKinds := stringSet(request.EdgeKinds)
	filterVertexKinds := request.VertexKinds != nil
	filterEdgeKinds := request.EdgeKinds != nil
	vertexByID := map[string]Vertex{}
	for _, vertex := range graphView.Vertices {
		vertexByID[vertex.VertexID] = vertex
	}
	visited := map[string]bool{}
	frontier := []string{}
	omittedSeeds := []string{}
	seenSeeds := map[string]bool{}
	for _, seed := range request.SeedVertexIDs {
		if seenSeeds[seed] {
			return TraverseResult{}, &OperationError{Code: "invalid_argument", ReasonCode: "duplicate_seed_vertex_id", Details: map[string]any{"field": "seed_vertex_ids"}}
		}
		seenSeeds[seed] = true
		if _, ok := vertexByID[seed]; ok {
			frontier = append(frontier, seed)
			visited[seed] = true
		} else {
			omittedSeeds = append(omittedSeeds, seed)
		}
	}
	selectedEdges := map[string]Edge{}
	for depth := 0; depth < maxDepth && len(frontier) > 0; depth++ {
		next := []string{}
		for _, current := range frontier {
			for _, edge := range graphView.Edges {
				if filterEdgeKinds && !edgeKinds[edge.EdgeKind] {
					continue
				}
				candidate := traversalNeighbor(edge, current, direction)
				if candidate == "" {
					continue
				}
				vertex := vertexByID[candidate]
				if filterVertexKinds && !vertexKinds[vertex.VertexKind] {
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
	seeds := make([]string, 0, len(request.SeedVertexIDs)-len(omittedSeeds))
	requestedSeeds := stringSet(request.SeedVertexIDs)
	for _, vertex := range graphView.Vertices {
		if requestedSeeds[vertex.VertexID] {
			seeds = append(seeds, vertex.VertexID)
		}
	}
	return TraverseResult{GraphViewID: graphView.GraphViewID, ProjectionRunID: graphView.ProjectionRunID, SeedVertexIDs: seeds, OmittedSeedVertexIDs: omittedSeeds, Vertices: vertices, Edges: edges, Metadata: map[string]any{}}, nil
}

func traversalNeighbor(edge Edge, current, requestedDirection string) string {
	if edge.Direction == "undirected" || edge.Direction == "bidirectional" || requestedDirection == "any" {
		if edge.SrcVertexID == current {
			return edge.DstVertexID
		}
		if edge.DstVertexID == current {
			return edge.SrcVertexID
		}
		return ""
	}
	if requestedDirection == "outbound" && edge.SrcVertexID == current {
		return edge.DstVertexID
	}
	if requestedDirection == "inbound" && edge.DstVertexID == current {
		return edge.SrcVertexID
	}
	return ""
}

func hasDuplicateStrings(values []string) bool {
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		if _, exists := seen[value]; exists {
			return true
		}
		seen[value] = struct{}{}
	}
	return false
}

func (s *Store) InvalidateGraphView(ctx context.Context, graphViewID, reasonCode, requestedAt, requestedBy string) (InvalidationSummary, error) {
	return s.invalidateGraphViewAt(ctx, graphViewID, reasonCode, requestedAt, requestedBy, "", time.Now().UTC())
}

func (s *Store) invalidateGraphViewAt(ctx context.Context, graphViewID, reasonCode, requestedAt, requestedBy, idempotencyKey string, invalidatedAt time.Time) (InvalidationSummary, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return InvalidationSummary{}, fmt.Errorf("begin graph projection invalidation: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	fingerprint, err := invalidationReplayFingerprint("invalidate_graph_view", graphViewID, "", reasonCode, requestedAt, requestedBy)
	if err != nil {
		return InvalidationSummary{}, err
	}
	if idempotencyKey != "" {
		if replayed, replay, err := s.checkInvalidationIdempotencyTx(ctx, tx, "invalidate_graph_view", graphViewID, idempotencyKey, fingerprint, invalidatedAt); err != nil {
			return InvalidationSummary{}, err
		} else if replay {
			return replayed, nil
		}
	}
	var selectedRunID string
	var retentionPolicyJSON []byte
	if err := tx.QueryRow(ctx, `
SELECT COALESCE(graph_view.selected_projection_run_id, graph_view.latest_projection_run_id), run.retention_policy_json
  FROM graph_projection_views AS graph_view
  JOIN graph_projection_runs AS run ON run.projection_run_id = COALESCE(graph_view.selected_projection_run_id, graph_view.latest_projection_run_id)
 WHERE graph_view.graph_view_id = $1
 FOR UPDATE OF graph_view, run
`, graphViewID).Scan(&selectedRunID, &retentionPolicyJSON); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return InvalidationSummary{}, ErrGraphViewNotFound
		}
		return InvalidationSummary{}, err
	}
	var retentionPolicy RetentionPolicy
	if err := json.Unmarshal(retentionPolicyJSON, &retentionPolicy); err != nil {
		return InvalidationSummary{}, fmt.Errorf("decode graph projection retention policy: %w", err)
	}
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
	summary := InvalidationSummary{GraphViewID: graphViewID, TargetScope: "graph_view", InvalidatedRunIDs: runIDs, GraphViewStateAfter: GraphViewStateInvalidated, InvalidatedAt: formatLifecycleTimestamp(invalidatedAt), ReasonCode: reasonCode, RequestedAt: requestedAt, RequestedBy: requestedBy}
	if idempotencyKey != "" {
		expiresAt := formatLifecycleTimestamp(invalidatedAt.Add(24 * time.Hour))
		summary.IdempotencyExpiresAt = &expiresAt
	}
	summaryJSON, _ := json.Marshal(summary)
	if _, err := tx.Exec(ctx, `UPDATE graph_projection_runs SET state = 'invalidated', invalidated_at = $2::timestamptz, retention_expires_at = CASE WHEN projection_run_id = $4 THEN NULL ELSE $2::timestamptz + make_interval(secs => $5::integer) END, invalidation_json = $3::jsonb WHERE graph_view_id = $1 AND state IN ('available', 'replaced')`, graphViewID, invalidatedAt, string(summaryJSON), selectedRunID, retentionPolicy.RetentionDurationSeconds); err != nil {
		return InvalidationSummary{}, fmt.Errorf("invalidate graph projection runs: %w", err)
	}
	if _, err := tx.Exec(ctx, `UPDATE graph_projection_views SET state = 'invalidated', validation_status = 'invalidated', updated_at = $2, invalidation_json = $3::jsonb WHERE graph_view_id = $1`, graphViewID, invalidatedAt, string(summaryJSON)); err != nil {
		return InvalidationSummary{}, fmt.Errorf("invalidate graph projection view: %w", err)
	}
	if err := s.pruneInvalidatedRetentionTx(ctx, tx, graphViewID, selectedRunID, retentionPolicy, invalidatedAt); err != nil {
		return InvalidationSummary{}, err
	}
	if idempotencyKey != "" {
		if err := s.recordInvalidationIdempotencyTx(ctx, tx, "invalidate_graph_view", graphViewID, idempotencyKey, fingerprint, summary, invalidatedAt); err != nil {
			return InvalidationSummary{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return InvalidationSummary{}, fmt.Errorf("commit graph projection invalidation: %w", err)
	}
	return summary, nil
}

func (s *Store) pruneInvalidatedRetentionTx(ctx context.Context, tx pgx.Tx, graphViewID, selectedRunID string, policy RetentionPolicy, queryReceivedAt time.Time) error {
	if _, err := tx.Exec(ctx, `
WITH ranked AS (
    SELECT projection_run_id,
           row_number() OVER (ORDER BY invalidated_at DESC, projection_run_id ASC) AS retention_rank
      FROM graph_projection_runs
     WHERE graph_view_id = $1 AND state = 'invalidated' AND projection_run_id <> $2
)
DELETE FROM graph_projection_runs AS run
 USING ranked
 WHERE run.projection_run_id = ranked.projection_run_id
   AND (
       $3::integer = 0
       OR $4::integer = 0
       OR ranked.retention_rank > $3
       OR $5::timestamptz >= run.invalidated_at + make_interval(secs => $4)
   )
`, graphViewID, selectedRunID, policy.RetentionCount, policy.RetentionDurationSeconds, queryReceivedAt); err != nil {
		return fmt.Errorf("prune invalidated graph projection runs: %w", err)
	}
	return nil
}

func (s *Store) invalidateProjectionRunAt(ctx context.Context, graphViewID, projectionRunID, reasonCode, requestedAt, requestedBy, idempotencyKey string, invalidatedAt time.Time) (InvalidationSummary, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return InvalidationSummary{}, fmt.Errorf("begin graph projection run invalidation: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	fingerprint, err := invalidationReplayFingerprint("invalidate_projection_run", graphViewID, projectionRunID, reasonCode, requestedAt, requestedBy)
	if err != nil {
		return InvalidationSummary{}, err
	}
	if idempotencyKey != "" {
		if replayed, replay, err := s.checkInvalidationIdempotencyTx(ctx, tx, "invalidate_projection_run", graphViewID, idempotencyKey, fingerprint, invalidatedAt); err != nil {
			return InvalidationSummary{}, err
		} else if replay {
			return replayed, nil
		}
	}
	var state string
	if err := tx.QueryRow(ctx, `SELECT state FROM graph_projection_runs WHERE graph_view_id = $1 AND projection_run_id = $2 FOR UPDATE`, graphViewID, projectionRunID).Scan(&state); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return InvalidationSummary{}, ErrProjectionRunNotFound
		}
		return InvalidationSummary{}, err
	}
	if state != string(RunStateAvailable) && state != string(RunStateReplaced) {
		return InvalidationSummary{}, &OperationError{Code: "invalid_operation", ReasonCode: "invalid_invalidation_target", Details: map[string]any{"operation": "invalidate_projection_run", "reason_code": "invalid_invalidation_target"}}
	}
	graphViewStateAfter := GraphViewStateAvailable
	var selectedRunID string
	var currentViewState string
	var retentionPolicyJSON []byte
	if err := tx.QueryRow(ctx, `
SELECT COALESCE(graph_view.selected_projection_run_id, graph_view.latest_projection_run_id, ''),
       graph_view.state,
       selected_run.retention_policy_json
  FROM graph_projection_views AS graph_view
  JOIN graph_projection_runs AS selected_run
    ON selected_run.projection_run_id = COALESCE(graph_view.selected_projection_run_id, graph_view.latest_projection_run_id)
 WHERE graph_view.graph_view_id = $1
 FOR UPDATE OF graph_view, selected_run
`, graphViewID).Scan(&selectedRunID, &currentViewState, &retentionPolicyJSON); err != nil {
		return InvalidationSummary{}, err
	}
	var retentionPolicy RetentionPolicy
	if err := json.Unmarshal(retentionPolicyJSON, &retentionPolicy); err != nil {
		return InvalidationSummary{}, fmt.Errorf("decode graph projection retention policy: %w", err)
	}
	if currentViewState == string(GraphViewStateInvalidated) || selectedRunID == projectionRunID {
		graphViewStateAfter = GraphViewStateInvalidated
	}
	summary := InvalidationSummary{GraphViewID: graphViewID, TargetScope: "projection_run", TargetProjectionRunID: &projectionRunID, InvalidatedRunIDs: []string{projectionRunID}, GraphViewStateAfter: graphViewStateAfter, InvalidatedAt: formatLifecycleTimestamp(invalidatedAt), ReasonCode: reasonCode, RequestedAt: requestedAt, RequestedBy: requestedBy}
	if idempotencyKey != "" {
		expiresAt := formatLifecycleTimestamp(invalidatedAt.Add(24 * time.Hour))
		summary.IdempotencyExpiresAt = &expiresAt
	}
	summaryJSON, err := json.Marshal(summary)
	if err != nil {
		return InvalidationSummary{}, err
	}
	if _, err := tx.Exec(ctx, `UPDATE graph_projection_runs SET state = 'invalidated', completed_at = COALESCE(completed_at, $3), invalidated_at = $3, retention_expires_at = CASE WHEN $2 = $5 THEN NULL ELSE $3::timestamptz + make_interval(secs => $6::integer) END, invalidation_json = $4::jsonb WHERE graph_view_id = $1 AND projection_run_id = $2`, graphViewID, projectionRunID, invalidatedAt, string(summaryJSON), selectedRunID, retentionPolicy.RetentionDurationSeconds); err != nil {
		return InvalidationSummary{}, err
	}
	if graphViewStateAfter == GraphViewStateInvalidated {
		if _, err := tx.Exec(ctx, `UPDATE graph_projection_views SET state = 'invalidated', validation_status = 'invalidated', updated_at = $2, invalidation_json = $3::jsonb WHERE graph_view_id = $1`, graphViewID, invalidatedAt, string(summaryJSON)); err != nil {
			return InvalidationSummary{}, err
		}
	}
	if err := s.pruneInvalidatedRetentionTx(ctx, tx, graphViewID, selectedRunID, retentionPolicy, invalidatedAt); err != nil {
		return InvalidationSummary{}, err
	}
	if idempotencyKey != "" {
		if err := s.recordInvalidationIdempotencyTx(ctx, tx, "invalidate_projection_run", graphViewID, idempotencyKey, fingerprint, summary, invalidatedAt); err != nil {
			return InvalidationSummary{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return InvalidationSummary{}, err
	}
	return summary, nil
}

func (s *Store) pruneRetentionTx(ctx context.Context, tx pgx.Tx, graphViewID string, policy RetentionPolicy, queryReceivedAt time.Time) error {
	if _, err := tx.Exec(ctx, `
WITH ranked AS (
    SELECT projection_run_id,
           row_number() OVER (ORDER BY replaced_at DESC NULLS LAST, projection_run_id ASC) AS retention_rank
      FROM graph_projection_runs
     WHERE graph_view_id = $1 AND state = 'replaced'
)
DELETE FROM graph_projection_runs AS run
 USING ranked
 WHERE run.projection_run_id = ranked.projection_run_id
   AND (
       $2::boolean = false
       OR $3::integer = 0
       OR $4::integer = 0
       OR ranked.retention_rank > $3
       OR $5::timestamptz >= run.replaced_at + make_interval(secs => $4)
   )
`, graphViewID, policy.RetainReplacedResults, policy.RetentionCount, policy.RetentionDurationSeconds, queryReceivedAt); err != nil {
		return fmt.Errorf("prune replaced graph projection runs: %w", err)
	}
	if _, err := tx.Exec(ctx, `
WITH ranked AS (
    SELECT projection_run_id,
           row_number() OVER (ORDER BY completed_at DESC NULLS LAST, projection_run_id ASC) AS retention_rank
      FROM graph_projection_runs
     WHERE graph_view_id = $1 AND state = 'failed'
)
DELETE FROM graph_projection_runs AS run
 USING ranked
 WHERE run.projection_run_id = ranked.projection_run_id
   AND run.projection_run_id <> COALESCE((SELECT latest_projection_run_id FROM graph_projection_views WHERE graph_view_id = $1 AND state = 'failed'), '')
   AND (
       $2::boolean = false
       OR $3::integer = 0
       OR $4::integer = 0
       OR ranked.retention_rank > $3
       OR $5::timestamptz >= run.completed_at + make_interval(secs => $4)
   )
`, graphViewID, policy.RetainFailedResults, policy.FailedRetentionCount, policy.FailedRetentionDurationSecs, queryReceivedAt); err != nil {
		return fmt.Errorf("prune failed graph projection runs: %w", err)
	}
	return nil
}

func invalidationReplayFingerprint(operation, graphViewID, projectionRunID, reasonCode, requestedAt, requestedBy string) (string, error) {
	bytes, err := canonicalJSON(map[string]any{
		"operation":         operation,
		"graph_view_id":     graphViewID,
		"projection_run_id": projectionRunID,
		"reason_code":       reasonCode,
		"requested_at":      requestedAt,
		"requested_by":      requestedBy,
	})
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (s *Store) checkInvalidationIdempotencyTx(ctx context.Context, tx pgx.Tx, operation, scopeKey, key, fingerprint string, queryReceivedAt time.Time) (InvalidationSummary, bool, error) {
	row := tx.QueryRow(ctx, `SELECT request_fingerprint, response_json, expires_at FROM graph_projection_idempotency WHERE operation = $1 AND scope_key = $2 AND idempotency_key = $3 FOR UPDATE`, operation, scopeKey, key)
	var existingFingerprint string
	var responseJSON []byte
	var expiresAt time.Time
	if err := row.Scan(&existingFingerprint, &responseJSON, &expiresAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return InvalidationSummary{}, false, nil
		}
		return InvalidationSummary{}, false, fmt.Errorf("load graph projection invalidation idempotency: %w", err)
	}
	if !queryReceivedAt.Before(expiresAt) {
		if _, err := tx.Exec(ctx, `DELETE FROM graph_projection_idempotency WHERE operation = $1 AND scope_key = $2 AND idempotency_key = $3`, operation, scopeKey, key); err != nil {
			return InvalidationSummary{}, false, err
		}
		return InvalidationSummary{}, false, nil
	}
	if existingFingerprint != fingerprint {
		return InvalidationSummary{}, false, &OperationError{Code: "operation_conflict", ReasonCode: "idempotency_key_conflict", Details: map[string]any{"operation": operation, "reason_code": "idempotency_key_conflict"}}
	}
	var summary InvalidationSummary
	if err := json.Unmarshal(responseJSON, &summary); err != nil {
		return InvalidationSummary{}, false, fmt.Errorf("decode graph projection invalidation replay: %w", err)
	}
	return summary, true, nil
}

func (s *Store) recordInvalidationIdempotencyTx(ctx context.Context, tx pgx.Tx, operation, scopeKey, key, fingerprint string, summary InvalidationSummary, createdAt time.Time) error {
	responseJSON, err := json.Marshal(summary)
	if err != nil {
		return err
	}
	var targetProjectionRunID any
	if summary.TargetProjectionRunID != nil {
		targetProjectionRunID = *summary.TargetProjectionRunID
	}
	_, err = tx.Exec(ctx, `
INSERT INTO graph_projection_idempotency (
    operation,
    scope_key,
    idempotency_key,
    request_fingerprint,
    graph_view_id,
    projection_run_id,
    response_json,
    created_at,
    expires_at
) VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb, $8, $9)
`, operation, scopeKey, key, fingerprint, summary.GraphViewID, targetProjectionRunID, string(responseJSON), createdAt, createdAt.Add(24*time.Hour))
	if err != nil {
		return fmt.Errorf("record graph projection invalidation idempotency: %w", err)
	}
	return nil
}

func (s *Store) checkIdempotencyTx(ctx context.Context, tx pgx.Tx, operation, scopeKey, key, fingerprint string, queryReceivedAt time.Time) (ProjectionRun, bool, error) {
	row := tx.QueryRow(ctx, `SELECT request_fingerprint, projection_run_id, expires_at FROM graph_projection_idempotency WHERE operation = $1 AND scope_key = $2 AND idempotency_key = $3 FOR UPDATE`, operation, scopeKey, key)
	var existingFingerprint string
	var projectionRunID string
	var expiresAt time.Time
	if err := row.Scan(&existingFingerprint, &projectionRunID, &expiresAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ProjectionRun{}, false, nil
		}
		return ProjectionRun{}, false, fmt.Errorf("load graph projection idempotency: %w", err)
	}
	if !queryReceivedAt.Before(expiresAt) {
		if _, err := tx.Exec(ctx, `DELETE FROM graph_projection_idempotency WHERE operation = $1 AND scope_key = $2 AND idempotency_key = $3`, operation, scopeKey, key); err != nil {
			return ProjectionRun{}, false, err
		}
		return ProjectionRun{}, false, nil
	}
	if existingFingerprint != fingerprint {
		return ProjectionRun{}, false, &OperationError{Code: "operation_conflict", ReasonCode: "idempotency_key_conflict", Details: map[string]any{"operation": operation, "reason_code": "idempotency_key_conflict"}}
	}
	run, err := s.GetProjectionRun(ctx, projectionRunID)
	return run, true, err
}

func retainedReplayFingerprint(operation string, run ProjectionRun) (string, error) {
	bytes, err := canonicalJSON(map[string]any{
		"operation":                operation,
		"graph_view_id":            run.GraphViewID,
		"projection_config_digest": run.ProjectionConfigDigest,
		"projection_source_digest": run.ProjectionSourceDigest,
	})
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (s *Store) recordIdempotencyTx(ctx context.Context, tx pgx.Tx, operation, scopeKey, key, fingerprint string, run ProjectionRun) error {
	responseJSON, err := json.Marshal(map[string]any{"projection_run_id": run.ProjectionRunID, "graph_view_id": run.GraphViewID, "status": run.State})
	if err != nil {
		return err
	}
	_, err = tx.Exec(ctx, `
INSERT INTO graph_projection_idempotency (
    operation,
	scope_key,
    idempotency_key,
    request_fingerprint,
    graph_view_id,
    projection_run_id,
    response_json,
    created_at,
    expires_at
) VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb, $8, $9)
`, operation, scopeKey, key, fingerprint, run.GraphViewID, run.ProjectionRunID, string(responseJSON), run.AcceptedAt, run.AcceptedAt.Add(24*time.Hour))
	if err != nil {
		return fmt.Errorf("record graph projection idempotency: %w", err)
	}
	return nil
}

type listCursor struct {
	Operation             string    `json:"operation"`
	AfterGraphViewID      string    `json:"after_graph_view_id"`
	IssuedAt              time.Time `json:"issued_at"`
	QueryShapeDigest      string    `json:"query_shape_digest"`
	VisibilityScopeDigest string    `json:"visibility_scope_digest"`
}

func cursorInvalid(reason string) error {
	return &OperationError{Code: "cursor_invalid", ReasonCode: reason, Details: map[string]any{"reason_code": reason}, cause: ErrCursorInvalid}
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
