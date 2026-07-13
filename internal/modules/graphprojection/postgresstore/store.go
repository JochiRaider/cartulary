package postgresstore

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"time"
	"unicode/utf8"

	"github.com/jackc/pgx/v5"

	graphprojection "github.com/JochiRaider/cartulary/internal/modules/graphprojection"
	"github.com/JochiRaider/cartulary/internal/platform/postgres"
)

type Store struct {
	pool        postgres.DB
	cursorCodec *graphCursorCodec
	now         func() time.Time
	hooks       Hooks
}

type Hooks struct {
	BeforePublication func(context.Context, graphprojection.ProjectionRun) error
}

func New(pool postgres.DB) *Store {
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

func NewWithCursorKey(pool postgres.DB, key []byte) (*Store, error) {
	return newStoreWithCursorKeyAndClock(pool, key, func() time.Time { return time.Now().UTC() })
}

func NewWithClock(pool postgres.DB, now func() time.Time) *Store {
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

func NewWithClockAndHooks(pool postgres.DB, now func() time.Time, hooks Hooks) *Store {
	store := NewWithClock(pool, now)
	store.hooks = hooks
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

func (s *Store) CreateProjection(ctx context.Context, data []byte, options graphprojection.RetainedProjectionOptions) (graphprojection.ProjectionRun, error) {
	return s.projectRetained(ctx, "create_projection", data, options)
}

func (s *Store) RefreshProjection(ctx context.Context, data []byte, options graphprojection.RetainedProjectionOptions) (graphprojection.ProjectionRun, error) {
	return s.projectRetained(ctx, "refresh_projection", data, options)
}

func (s *Store) projectRetained(ctx context.Context, operation string, data []byte, options graphprojection.RetainedProjectionOptions) (graphprojection.ProjectionRun, error) {
	run, err := graphprojection.AdmitRetainedProjection(data, options.ProjectionRunNonce, options.AcceptedAt)
	if err != nil {
		return graphprojection.ProjectionRun{}, err
	}
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return graphprojection.ProjectionRun{}, fmt.Errorf("begin graph projection admission: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()

	if options.IdempotencyKey != "" {
		fingerprint, err := retainedReplayFingerprint(operation, run)
		if err != nil {
			return graphprojection.ProjectionRun{}, err
		}
		replayed, replay, err := s.checkIdempotencyTx(ctx, tx, operation, run.GraphViewID, options.IdempotencyKey, fingerprint, run.AcceptedAt)
		if err != nil {
			return graphprojection.ProjectionRun{}, err
		}
		if replay {
			return replayed, nil
		}
	}

	var currentViewState string
	var currentLatestRunID string
	var currentSelectedRunID string
	var currentUpdatedAt time.Time
	viewErr := tx.QueryRow(ctx, `SELECT state, COALESCE(latest_projection_run_id, ''), COALESCE(selected_projection_run_id, latest_projection_run_id, ''), updated_at FROM graph_projection_views WHERE graph_view_id = $1 FOR UPDATE`, run.GraphViewID).Scan(&currentViewState, &currentLatestRunID, &currentSelectedRunID, &currentUpdatedAt)
	viewExists := viewErr == nil
	if viewErr != nil && !errors.Is(viewErr, pgx.ErrNoRows) {
		return graphprojection.ProjectionRun{}, fmt.Errorf("load graph projection lifecycle state: %w", viewErr)
	}
	if viewExists && !run.AcceptedAt.After(currentUpdatedAt) {
		run.AcceptedAt = currentUpdatedAt.Add(time.Microsecond)
	}
	if operation == "create_projection" && viewExists {
		switch graphprojection.GraphViewState(currentViewState) {
		case graphprojection.GraphViewStateCreating, graphprojection.GraphViewStateRefreshing:
			return graphprojection.ProjectionRun{}, &graphprojection.OperationError{Code: "operation_conflict", ReasonCode: "run_already_active", Details: map[string]any{"operation": operation, "reason_code": "run_already_active", "active_projection_run_id": currentLatestRunID}}
		case graphprojection.GraphViewStateAvailable, graphprojection.GraphViewStateInvalidated:
			return graphprojection.ProjectionRun{}, &graphprojection.OperationError{Code: "invalid_operation", ReasonCode: "graph_view_already_exists", Details: map[string]any{"operation": operation, "reason_code": "graph_view_already_exists"}}
		}
	}
	if operation == "refresh_projection" {
		if !viewExists {
			return graphprojection.ProjectionRun{}, graphprojection.ErrGraphViewNotFound
		}
		switch graphprojection.GraphViewState(currentViewState) {
		case graphprojection.GraphViewStateCreating, graphprojection.GraphViewStateRefreshing:
			return graphprojection.ProjectionRun{}, &graphprojection.OperationError{Code: "operation_conflict", ReasonCode: "run_already_active", Details: map[string]any{"operation": operation, "reason_code": "run_already_active", "active_projection_run_id": currentLatestRunID}}
		case graphprojection.GraphViewStateFailed:
			return graphprojection.ProjectionRun{}, &graphprojection.OperationError{Code: "invalid_operation", ReasonCode: "no_consumable_prior_run", Details: map[string]any{"operation": operation, "reason_code": "no_consumable_prior_run"}}
		}
	}

	var previousRunID *string
	if operation == "refresh_projection" && currentSelectedRunID != "" {
		previousRunID = &currentSelectedRunID
		run.PreviousProjectionRunID = previousRunID
	}
	admissionViewState := graphprojection.GraphViewStateCreating
	if operation == "refresh_projection" {
		admissionViewState = graphprojection.GraphViewStateRefreshing
	}
	if err := s.persistAcceptedRunTx(ctx, tx, run, admissionViewState, currentSelectedRunID); err != nil {
		return graphprojection.ProjectionRun{}, err
	}
	if options.IdempotencyKey != "" {
		fingerprint, err := retainedReplayFingerprint(operation, run)
		if err != nil {
			return graphprojection.ProjectionRun{}, err
		}
		if err := s.recordIdempotencyTx(ctx, tx, operation, run.GraphViewID, options.IdempotencyKey, fingerprint, run); err != nil {
			return graphprojection.ProjectionRun{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return graphprojection.ProjectionRun{}, fmt.Errorf("commit graph projection admission: %w", err)
	}

	startedAt := options.GeneratedAt.UTC()
	if !startedAt.After(run.AcceptedAt) {
		startedAt = run.AcceptedAt.Add(time.Microsecond)
	}
	run.StartedAt = &startedAt
	run.State = graphprojection.RunStateComputing
	if _, err := s.pool.Exec(ctx, `UPDATE graph_projection_runs SET state = 'computing', started_at = $2 WHERE projection_run_id = $1 AND state = 'accepted'`, run.ProjectionRunID, startedAt); err != nil {
		return graphprojection.ProjectionRun{}, fmt.Errorf("start graph projection run: %w", err)
	}

	generatedAt := options.GeneratedAt.UTC()
	if !generatedAt.After(startedAt) {
		generatedAt = startedAt.Add(time.Microsecond)
	}
	result := graphprojection.ProjectAdmittedRetainedProjection(run, generatedAt, previousRunID)
	completedAt := generatedAt.Add(time.Microsecond)
	result.CompletedAt = &completedAt
	if s.hooks.BeforePublication != nil {
		if err := s.hooks.BeforePublication(ctx, result); err != nil {
			return graphprojection.ProjectionRun{}, err
		}
	}
	terminalTx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return graphprojection.ProjectionRun{}, fmt.Errorf("begin graph projection publication: %w", err)
	}
	defer func() { _ = terminalTx.Rollback(ctx) }()
	var terminalViewState string
	var terminalSelectedRunID string
	var invalidationJSON []byte
	if err := terminalTx.QueryRow(ctx, `SELECT state, COALESCE(selected_projection_run_id, ''), COALESCE(invalidation_json, 'null'::jsonb) FROM graph_projection_views WHERE graph_view_id = $1 FOR UPDATE`, result.GraphViewID).Scan(&terminalViewState, &terminalSelectedRunID, &invalidationJSON); err != nil {
		return graphprojection.ProjectionRun{}, err
	}
	wasInvalidated := graphprojection.GraphViewState(terminalViewState) == graphprojection.GraphViewStateInvalidated || string(invalidationJSON) != "null"
	if wasInvalidated && result.State == graphprojection.RunStateAvailable {
		result.State = graphprojection.RunStateInvalidated
		result.PreviousProjectionRunID = nil
		var invalidation graphprojection.Invalidation
		if string(invalidationJSON) != "null" && json.Unmarshal(invalidationJSON, &invalidation) == nil {
			result.Invalidation = &invalidation
			if timestamp, parseErr := time.Parse("2006-01-02T15:04:05.999999Z", invalidation.InvalidatedAt); parseErr == nil {
				result.InvalidatedAt = &timestamp
			}
		}
		result.GraphView = nil
		result.ProjectionOutputDigest = ""
	}
	publicationPriorState := graphprojection.GraphViewState(terminalViewState)
	if wasInvalidated {
		publicationPriorState = graphprojection.GraphViewStateInvalidated
	}
	if err := s.publishProjectionRunTx(ctx, terminalTx, result, operation, publicationPriorState, terminalSelectedRunID); err != nil {
		return graphprojection.ProjectionRun{}, err
	}
	queryReceivedAt := result.AcceptedAt
	if result.CompletedAt != nil {
		queryReceivedAt = *result.CompletedAt
	}
	if err := s.pruneRetentionTx(ctx, terminalTx, result.GraphViewID, result.Request.ProjectionConfig.RetentionPolicy, queryReceivedAt); err != nil {
		return graphprojection.ProjectionRun{}, err
	}
	if err := terminalTx.Commit(ctx); err != nil {
		return graphprojection.ProjectionRun{}, fmt.Errorf("commit graph projection publication: %w", err)
	}
	return result, nil
}

func (s *Store) persistAcceptedRunTx(ctx context.Context, tx pgx.Tx, run graphprojection.ProjectionRun, viewState graphprojection.GraphViewState, selectedRunID string) error {
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

func (s *Store) publishProjectionRunTx(ctx context.Context, tx pgx.Tx, run graphprojection.ProjectionRun, operation string, priorViewState graphprojection.GraphViewState, selectedRunID string) error {
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
	var runInvalidationJSON []byte
	if run.Invalidation != nil {
		runInvalidationJSON, err = json.Marshal(run.Invalidation)
		if err != nil {
			return fmt.Errorf("marshal graph projection invalidation: %w", err)
		}
	}
	var retentionExpiresAt any
	if run.State == graphprojection.RunStateFailed && operation == "refresh_projection" && run.CompletedAt != nil {
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
	   retention_expires_at = $10,
	   invalidation_json = $11::jsonb
 WHERE projection_run_id = $1
   AND state = 'computing'
`, run.ProjectionRunID, string(run.State), nullString(run.ProjectionOutputDigest), run.GeneratedAt, run.CompletedAt, run.InvalidatedAt, string(summaryJSON), nullString(run.FailureReason), nullJSON(graphJSON), retentionExpiresAt, nullJSON(runInvalidationJSON)); err != nil {
		return fmt.Errorf("publish graph projection run: %w", err)
	}

	updatedAt := run.AcceptedAt
	if run.CompletedAt != nil {
		updatedAt = *run.CompletedAt
	}
	viewState := graphprojection.GraphViewStateFailed
	viewLatestRunID := run.ProjectionRunID
	viewSelectedRunID := ""
	validationStatus := run.ValidationSummary.Status
	if validationStatus == "" {
		validationStatus = "failed"
	}
	preserveInvalidation := false
	switch run.State {
	case graphprojection.RunStateAvailable:
		viewState = graphprojection.GraphViewStateAvailable
		viewSelectedRunID = run.ProjectionRunID
		if selectedRunID != "" && selectedRunID != run.ProjectionRunID {
			retentionExpiresAt := updatedAt.Add(time.Duration(run.Request.ProjectionConfig.RetentionPolicy.RetentionDurationSeconds) * time.Second)
			if _, err := tx.Exec(ctx, `UPDATE graph_projection_runs SET state = 'replaced', replaced_at = $2, retention_expires_at = $3 WHERE projection_run_id = $1 AND state = 'available'`, selectedRunID, updatedAt, retentionExpiresAt); err != nil {
				return fmt.Errorf("replace prior graph projection run: %w", err)
			}
		}
	case graphprojection.RunStateInvalidated:
		viewState = graphprojection.GraphViewStateInvalidated
		viewSelectedRunID = run.ProjectionRunID
		validationStatus = "invalidated"
		preserveInvalidation = true
	case graphprojection.RunStateFailed:
		if operation == "refresh_projection" && selectedRunID != "" {
			viewSelectedRunID = selectedRunID
			if priorViewState == graphprojection.GraphViewStateInvalidated {
				viewState = graphprojection.GraphViewStateInvalidated
				validationStatus = "invalidated"
				preserveInvalidation = true
			} else {
				viewState = graphprojection.GraphViewStateAvailable
				validationStatus = "passed"
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

func (s *Store) GetProjectionRun(ctx context.Context, projectionRunID string) (graphprojection.ProjectionRun, error) {
	queryReceivedAt := s.now().UTC()
	if _, err := s.pool.Exec(ctx, `DELETE FROM graph_projection_runs WHERE projection_run_id = $1 AND retention_expires_at IS NOT NULL AND $2::timestamptz >= retention_expires_at`, projectionRunID, queryReceivedAt); err != nil {
		return graphprojection.ProjectionRun{}, fmt.Errorf("expire graph projection run: %w", err)
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
	   invalidation_json,
	       retention_expires_at
  FROM graph_projection_runs
 WHERE projection_run_id = $1
`, projectionRunID)
	var run graphprojection.ProjectionRun
	var state string
	var summaryJSON []byte
	var graphJSON []byte
	var invalidationJSON []byte
	var failureReason string
	var projectionOutputDigest *string
	var completedAt *time.Time
	var retentionExpiresAt *time.Time
	if err := row.Scan(&run.GraphViewID, &run.Request.SourceSnapshotID, &run.Request.ProjectionConfig.ProjectionVersion, &state, &run.ProjectionRunNonce, &run.ProjectionConfigDigest, &run.ProjectionSourceDigest, &projectionOutputDigest, &run.AcceptedAt, &run.StartedAt, &run.GeneratedAt, &completedAt, &run.ReplacedAt, &run.InvalidatedAt, &summaryJSON, &failureReason, &graphJSON, &invalidationJSON, &retentionExpiresAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return graphprojection.ProjectionRun{}, graphprojection.ErrProjectionRunNotFound
		}
		return graphprojection.ProjectionRun{}, fmt.Errorf("get graph projection run: %w", err)
	}
	run.ProjectionRunID = projectionRunID
	run.State = graphprojection.RunState(state)
	run.CompletedAt = completedAt
	if projectionOutputDigest != nil {
		run.ProjectionOutputDigest = *projectionOutputDigest
	}
	run.FailureReason = failureReason
	run.RetentionExpiresAt = retentionExpiresAt
	if len(summaryJSON) > 0 {
		if err := json.Unmarshal(summaryJSON, &run.ValidationSummary); err != nil {
			return graphprojection.ProjectionRun{}, fmt.Errorf("decode validation summary: %w", err)
		}
	}
	if len(graphJSON) > 0 {
		var graphView graphprojection.GraphView
		if err := json.Unmarshal(graphJSON, &graphView); err != nil {
			return graphprojection.ProjectionRun{}, fmt.Errorf("decode graph view: %w", err)
		}
		run.GraphView = &graphView
	}
	if len(invalidationJSON) > 0 {
		var invalidation graphprojection.Invalidation
		if err := json.Unmarshal(invalidationJSON, &invalidation); err != nil {
			return graphprojection.ProjectionRun{}, fmt.Errorf("decode graph projection invalidation: %w", err)
		}
		run.Invalidation = &invalidation
	}
	return run, nil
}

func (s *Store) GetGraphView(ctx context.Context, graphViewID string, projectionRunID string) (graphprojection.GraphView, error) {
	resolvedRunID, err := s.resolveReadableRunID(ctx, graphViewID, projectionRunID)
	if err != nil {
		return graphprojection.GraphView{}, err
	}
	run, err := s.GetProjectionRun(ctx, resolvedRunID)
	if err != nil {
		return graphprojection.GraphView{}, err
	}
	if run.GraphView == nil || (run.State != graphprojection.RunStateAvailable && run.State != graphprojection.RunStateReplaced) {
		return graphprojection.GraphView{}, graphprojection.ErrGraphViewUnavailable
	}
	if run.GraphViewID != graphViewID {
		return graphprojection.GraphView{}, graphprojection.ErrProjectionRunNotFound
	}
	return *run.GraphView, nil
}

func (s *Store) GetVertex(ctx context.Context, graphViewID, projectionRunID, vertexID string) (graphprojection.Vertex, error) {
	projectionRunID, err := s.resolveReadableRunID(ctx, graphViewID, projectionRunID)
	if err != nil {
		return graphprojection.Vertex{}, err
	}
	var payload []byte
	if err := s.pool.QueryRow(ctx, `SELECT vertex_json FROM graph_projection_vertices WHERE graph_view_id = $1 AND projection_run_id = $2 AND vertex_id = $3`, graphViewID, projectionRunID, vertexID).Scan(&payload); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return graphprojection.Vertex{}, graphprojection.ErrVertexNotFound
		}
		return graphprojection.Vertex{}, fmt.Errorf("get graph projection vertex: %w", err)
	}
	var vertex graphprojection.Vertex
	if err := json.Unmarshal(payload, &vertex); err != nil {
		return graphprojection.Vertex{}, fmt.Errorf("decode graph projection vertex: %w", err)
	}
	return vertex, nil
}

func (s *Store) GetEdge(ctx context.Context, graphViewID, projectionRunID, edgeID string) (graphprojection.Edge, error) {
	projectionRunID, err := s.resolveReadableRunID(ctx, graphViewID, projectionRunID)
	if err != nil {
		return graphprojection.Edge{}, err
	}
	var payload []byte
	if err := s.pool.QueryRow(ctx, `SELECT edge_json FROM graph_projection_edges WHERE graph_view_id = $1 AND projection_run_id = $2 AND edge_id = $3`, graphViewID, projectionRunID, edgeID).Scan(&payload); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return graphprojection.Edge{}, graphprojection.ErrEdgeNotFound
		}
		return graphprojection.Edge{}, fmt.Errorf("get graph projection edge: %w", err)
	}
	var edge graphprojection.Edge
	if err := json.Unmarshal(payload, &edge); err != nil {
		return graphprojection.Edge{}, fmt.Errorf("decode graph projection edge: %w", err)
	}
	return edge, nil
}

func (s *Store) resolveReadableRunID(ctx context.Context, graphViewID, projectionRunID string) (string, error) {
	suppliedRun := projectionRunID != ""
	if projectionRunID == "" {
		var viewState string
		if err := s.pool.QueryRow(ctx, `SELECT state, COALESCE(selected_projection_run_id, '') FROM graph_projection_views WHERE graph_view_id = $1`, graphViewID).Scan(&viewState, &projectionRunID); err != nil {
			if errors.Is(err, pgx.ErrNoRows) {
				return "", graphprojection.NewOperationError("graph_view_not_found", "", map[string]any{"graph_view_id": graphViewID}, graphprojection.ErrGraphViewNotFound)
			}
			return "", fmt.Errorf("resolve selected graph projection run: %w", err)
		}
		if projectionRunID == "" || graphprojection.GraphViewState(viewState) == graphprojection.GraphViewStateCreating || graphprojection.GraphViewState(viewState) == graphprojection.GraphViewStateRefreshing || graphprojection.GraphViewState(viewState) == graphprojection.GraphViewStateFailed || graphprojection.GraphViewState(viewState) == graphprojection.GraphViewStateInvalidated {
			return "", graphprojection.NewOperationError("projection_not_available", viewState, map[string]any{"graph_view_id": graphViewID, "state": viewState}, graphprojection.ErrGraphViewUnavailable)
		}
	}
	var state string
	var invalidationJSON []byte
	if err := s.pool.QueryRow(ctx, `SELECT state, COALESCE(invalidation_json, 'null'::jsonb) FROM graph_projection_runs WHERE graph_view_id = $1 AND projection_run_id = $2`, graphViewID, projectionRunID).Scan(&state, &invalidationJSON); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return "", graphprojection.NewOperationError("projection_run_not_found", "", map[string]any{"graph_view_id": graphViewID, "projection_run_id": projectionRunID}, graphprojection.ErrProjectionRunNotFound)
		}
		return "", fmt.Errorf("resolve graph projection run: %w", err)
	}
	switch graphprojection.RunState(state) {
	case graphprojection.RunStateAvailable, graphprojection.RunStateReplaced:
		return projectionRunID, nil
	case graphprojection.RunStateAccepted, graphprojection.RunStateComputing:
		return "", graphprojection.NewOperationError("projection_not_available", state, map[string]any{"graph_view_id": graphViewID, "state": state}, graphprojection.ErrGraphViewUnavailable)
	case graphprojection.RunStateFailed:
		return "", graphprojection.NewOperationError("projection_run_failed", "", map[string]any{"graph_view_id": graphViewID, "projection_run_id": projectionRunID}, graphprojection.ErrGraphViewUnavailable)
	case graphprojection.RunStateInvalidated:
		if !suppliedRun {
			return "", graphprojection.NewOperationError("projection_not_available", string(graphprojection.GraphViewStateInvalidated), map[string]any{"graph_view_id": graphViewID, "state": graphprojection.GraphViewStateInvalidated}, graphprojection.ErrGraphViewUnavailable)
		}
		invalidationDetails := map[string]any{"reason_code": nil, "invalidated_at": nil}
		details := map[string]any{"graph_view_id": graphViewID, "projection_run_id": projectionRunID, "invalidation": invalidationDetails}
		var invalidation graphprojection.Invalidation
		if string(invalidationJSON) != "null" && json.Unmarshal(invalidationJSON, &invalidation) == nil {
			invalidationDetails["reason_code"] = invalidation.ReasonCode
			invalidationDetails["invalidated_at"] = invalidation.InvalidatedAt
		}
		return "", graphprojection.NewOperationError("projection_run_invalidated", "", details, graphprojection.ErrGraphViewUnavailable)
	default:
		return "", graphprojection.NewOperationError("projection_run_not_found", "", map[string]any{"graph_view_id": graphViewID, "projection_run_id": projectionRunID}, graphprojection.ErrProjectionRunNotFound)
	}
}

func (s *Store) ListGraphViews(ctx context.Context, options graphprojection.ListGraphViewsOptions) ([]graphprojection.GraphViewSummary, string, error) {
	limit := 100
	if options.Limit != nil {
		limit = *options.Limit
		if limit < 1 || limit > graphprojection.ResourceLimits().MaxListGraphViewsLimit {
			return nil, "", &graphprojection.OperationError{Code: "invalid_argument", ReasonCode: "invalid_limit", Field: "limit", Details: map[string]any{"field": "limit"}}
		}
	}
	after := ""
	now := s.now().UTC()
	if options.CursorToken != "" {
		if utf8.RuneCountInString(options.CursorToken) > graphprojection.ResourceLimits().MaxCursorTokenLength {
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
	summaries := []graphprojection.GraphViewSummary{}
	for rows.Next() {
		var summary graphprojection.GraphViewSummary
		var state string
		if err := rows.Scan(&summary.GraphViewID, &summary.GraphViewKey, &state, &summary.LatestProjectionRunID, &summary.LatestSourceSnapshotID, &summary.ProjectionVersion, &summary.UpdatedAt, &summary.ValidationStatus); err != nil {
			return nil, "", fmt.Errorf("scan graph view summary: %w", err)
		}
		summary.State = graphprojection.GraphViewState(state)
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

func (s *Store) Traverse(ctx context.Context, request graphprojection.TraverseRequest) (graphprojection.TraverseResult, error) {
	graphView, err := s.GetGraphView(ctx, request.GraphViewID, request.ProjectionRunID)
	if err != nil {
		return graphprojection.TraverseResult{}, err
	}
	maxDepth := 1
	if request.MaxDepth != nil {
		maxDepth = *request.MaxDepth
	}
	if maxDepth < 0 || maxDepth > graphprojection.ResourceLimits().MaxTraversalDepth {
		return graphprojection.TraverseResult{}, &graphprojection.OperationError{Code: "invalid_argument", ReasonCode: "invalid_max_depth", Details: map[string]any{"field": "max_depth"}}
	}
	direction := request.Direction
	if direction == "" {
		direction = "outbound"
	}
	if direction != "outbound" && direction != "inbound" && direction != "any" {
		return graphprojection.TraverseResult{}, &graphprojection.OperationError{Code: "invalid_argument", ReasonCode: "invalid_direction", Details: map[string]any{"field": "direction"}}
	}
	if len(request.SeedVertexIDs) > graphprojection.ResourceLimits().MaxTraversalSeedVertices || len(request.VertexKinds) > graphprojection.ResourceLimits().MaxTraversalKindFilters || len(request.EdgeKinds) > graphprojection.ResourceLimits().MaxTraversalKindFilters {
		return graphprojection.TraverseResult{}, &graphprojection.OperationError{Code: "invalid_argument", ReasonCode: "collection_limit_exceeded", Details: map[string]any{"field": "traverse"}}
	}
	if hasDuplicateStrings(request.VertexKinds) || hasDuplicateStrings(request.EdgeKinds) {
		return graphprojection.TraverseResult{}, &graphprojection.OperationError{Code: "invalid_argument", ReasonCode: "duplicate_filter_value", Details: map[string]any{"field": "traverse"}}
	}
	vertexKinds := stringSet(request.VertexKinds)
	edgeKinds := stringSet(request.EdgeKinds)
	filterVertexKinds := request.VertexKinds != nil
	filterEdgeKinds := request.EdgeKinds != nil
	vertexByID := map[string]graphprojection.Vertex{}
	for _, vertex := range graphView.Vertices {
		vertexByID[vertex.VertexID] = vertex
	}
	visited := map[string]bool{}
	frontier := []string{}
	omittedSeeds := []string{}
	seenSeeds := map[string]bool{}
	for _, seed := range request.SeedVertexIDs {
		if seenSeeds[seed] {
			return graphprojection.TraverseResult{}, &graphprojection.OperationError{Code: "invalid_argument", ReasonCode: "duplicate_seed_vertex_id", Details: map[string]any{"field": "seed_vertex_ids"}}
		}
		seenSeeds[seed] = true
		if _, ok := vertexByID[seed]; ok {
			frontier = append(frontier, seed)
			visited[seed] = true
		} else {
			omittedSeeds = append(omittedSeeds, seed)
		}
	}
	selectedEdges := map[string]graphprojection.Edge{}
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
	vertices := []graphprojection.Vertex{}
	for vertexID := range visited {
		vertices = append(vertices, vertexByID[vertexID])
	}
	edges := []graphprojection.Edge{}
	for _, edge := range selectedEdges {
		edges = append(edges, edge)
	}
	graphprojection.SortVertices(vertices)
	graphprojection.SortEdges(edges)
	seeds := make([]string, 0, len(request.SeedVertexIDs)-len(omittedSeeds))
	requestedSeeds := stringSet(request.SeedVertexIDs)
	for _, vertex := range graphView.Vertices {
		if requestedSeeds[vertex.VertexID] {
			seeds = append(seeds, vertex.VertexID)
		}
	}
	return graphprojection.TraverseResult{GraphViewID: graphView.GraphViewID, ProjectionRunID: graphView.ProjectionRunID, SeedVertexIDs: seeds, OmittedSeedVertexIDs: omittedSeeds, Vertices: vertices, Edges: edges, Metadata: map[string]any{}}, nil
}

func traversalNeighbor(edge graphprojection.Edge, current, requestedDirection string) string {
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

func stringSet(values []string) map[string]bool {
	set := make(map[string]bool, len(values))
	for _, value := range values {
		set[value] = true
	}
	return set
}

func (s *Store) InvalidateGraphView(ctx context.Context, request graphprojection.RetainedInvalidation) (graphprojection.InvalidationSummary, error) {
	return s.invalidateGraphViewAt(ctx, request.GraphViewID, request.ReasonCode, request.RequestedAt, request.RequestedBy, request.IdempotencyKey, request.InvalidatedAt)
}

func (s *Store) invalidateGraphViewAt(ctx context.Context, graphViewID, reasonCode, requestedAt, requestedBy, idempotencyKey string, invalidatedAt time.Time) (graphprojection.InvalidationSummary, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return graphprojection.InvalidationSummary{}, fmt.Errorf("begin graph projection invalidation: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	fingerprint, err := invalidationReplayFingerprint("invalidate_graph_view", graphViewID, "", reasonCode, requestedBy)
	if err != nil {
		return graphprojection.InvalidationSummary{}, err
	}
	if idempotencyKey != "" {
		if replayed, replay, err := s.checkInvalidationIdempotencyTx(ctx, tx, "invalidate_graph_view", graphViewID, idempotencyKey, fingerprint, invalidatedAt); err != nil {
			return graphprojection.InvalidationSummary{}, err
		} else if replay {
			return replayed, nil
		}
	}
	var selectedRunID string
	var retentionPolicyJSON []byte
	var currentUpdatedAt time.Time
	if err := tx.QueryRow(ctx, `
SELECT COALESCE(graph_view.selected_projection_run_id, graph_view.latest_projection_run_id), run.retention_policy_json, graph_view.updated_at
  FROM graph_projection_views AS graph_view
  JOIN graph_projection_runs AS run ON run.projection_run_id = COALESCE(graph_view.selected_projection_run_id, graph_view.latest_projection_run_id)
 WHERE graph_view.graph_view_id = $1
 FOR UPDATE OF graph_view, run
`, graphViewID).Scan(&selectedRunID, &retentionPolicyJSON, &currentUpdatedAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return graphprojection.InvalidationSummary{}, graphprojection.ErrGraphViewNotFound
		}
		return graphprojection.InvalidationSummary{}, err
	}
	var retentionPolicy graphprojection.RetentionPolicy
	if err := json.Unmarshal(retentionPolicyJSON, &retentionPolicy); err != nil {
		return graphprojection.InvalidationSummary{}, fmt.Errorf("decode graph projection retention policy: %w", err)
	}
	if !invalidatedAt.After(currentUpdatedAt) {
		invalidatedAt = currentUpdatedAt.Add(time.Microsecond)
	}
	rows, err := tx.Query(ctx, `SELECT projection_run_id FROM graph_projection_runs WHERE graph_view_id = $1 AND state IN ('available', 'replaced') ORDER BY projection_run_id ASC`, graphViewID)
	if err != nil {
		return graphprojection.InvalidationSummary{}, fmt.Errorf("list invalidation targets: %w", err)
	}
	runIDs := []string{}
	for rows.Next() {
		var runID string
		if err := rows.Scan(&runID); err != nil {
			rows.Close()
			return graphprojection.InvalidationSummary{}, err
		}
		runIDs = append(runIDs, runID)
	}
	rows.Close()
	if len(runIDs) == 0 {
		return graphprojection.InvalidationSummary{}, graphprojection.ErrGraphViewNotFound
	}
	summary := graphprojection.InvalidationSummary{GraphViewID: graphViewID, TargetScope: "graph_view", InvalidatedRunIDs: runIDs, GraphViewStateAfter: graphprojection.GraphViewStateInvalidated, InvalidatedAt: graphprojection.FormatLifecycleTimestamp(invalidatedAt), ReasonCode: reasonCode, RequestedAt: requestedAt, RequestedBy: requestedBy}
	if idempotencyKey != "" {
		expiresAt := graphprojection.FormatLifecycleTimestamp(invalidatedAt.Add(24 * time.Hour))
		summary.IdempotencyExpiresAt = &expiresAt
	}
	invalidationJSON, _ := json.Marshal(graphprojection.Invalidation{InvalidatedAt: summary.InvalidatedAt, ReasonCode: reasonCode, RequestedBy: requestedBy, TargetScope: "graph_view"})
	if _, err := tx.Exec(ctx, `UPDATE graph_projection_runs SET state = 'invalidated', invalidated_at = $2::timestamptz, retention_expires_at = CASE WHEN projection_run_id = $4 THEN NULL ELSE $2::timestamptz + make_interval(secs => $5::integer) END, invalidation_json = $3::jsonb WHERE graph_view_id = $1 AND state IN ('available', 'replaced')`, graphViewID, invalidatedAt, string(invalidationJSON), selectedRunID, retentionPolicy.RetentionDurationSeconds); err != nil {
		return graphprojection.InvalidationSummary{}, fmt.Errorf("invalidate graph projection runs: %w", err)
	}
	if _, err := tx.Exec(ctx, `UPDATE graph_projection_views SET state = 'invalidated', validation_status = 'invalidated', updated_at = $2, invalidation_json = $3::jsonb WHERE graph_view_id = $1`, graphViewID, invalidatedAt, string(invalidationJSON)); err != nil {
		return graphprojection.InvalidationSummary{}, fmt.Errorf("invalidate graph projection view: %w", err)
	}
	if err := s.pruneInvalidatedRetentionTx(ctx, tx, graphViewID, selectedRunID, retentionPolicy, invalidatedAt); err != nil {
		return graphprojection.InvalidationSummary{}, err
	}
	if idempotencyKey != "" {
		if err := s.recordInvalidationIdempotencyTx(ctx, tx, "invalidate_graph_view", graphViewID, idempotencyKey, fingerprint, summary, invalidatedAt); err != nil {
			return graphprojection.InvalidationSummary{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return graphprojection.InvalidationSummary{}, fmt.Errorf("commit graph projection invalidation: %w", err)
	}
	return summary, nil
}

func (s *Store) pruneInvalidatedRetentionTx(ctx context.Context, tx pgx.Tx, graphViewID, selectedRunID string, policy graphprojection.RetentionPolicy, queryReceivedAt time.Time) error {
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

func (s *Store) InvalidateProjectionRun(ctx context.Context, request graphprojection.RetainedInvalidation) (graphprojection.InvalidationSummary, error) {
	return s.invalidateProjectionRunAt(ctx, request.GraphViewID, request.ProjectionRunID, request.ReasonCode, request.RequestedAt, request.RequestedBy, request.IdempotencyKey, request.InvalidatedAt)
}

func (s *Store) invalidateProjectionRunAt(ctx context.Context, graphViewID, projectionRunID, reasonCode, requestedAt, requestedBy, idempotencyKey string, invalidatedAt time.Time) (graphprojection.InvalidationSummary, error) {
	tx, err := s.pool.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return graphprojection.InvalidationSummary{}, fmt.Errorf("begin graph projection run invalidation: %w", err)
	}
	defer func() { _ = tx.Rollback(ctx) }()
	fingerprint, err := invalidationReplayFingerprint("invalidate_projection_run", graphViewID, projectionRunID, reasonCode, requestedBy)
	if err != nil {
		return graphprojection.InvalidationSummary{}, err
	}
	if idempotencyKey != "" {
		scopeKey := graphViewID + "\n" + projectionRunID
		if replayed, replay, err := s.checkInvalidationIdempotencyTx(ctx, tx, "invalidate_projection_run", scopeKey, idempotencyKey, fingerprint, invalidatedAt); err != nil {
			return graphprojection.InvalidationSummary{}, err
		} else if replay {
			return replayed, nil
		}
	}
	var state string
	if err := tx.QueryRow(ctx, `SELECT state FROM graph_projection_runs WHERE graph_view_id = $1 AND projection_run_id = $2 FOR UPDATE`, graphViewID, projectionRunID).Scan(&state); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return graphprojection.InvalidationSummary{}, graphprojection.ErrProjectionRunNotFound
		}
		return graphprojection.InvalidationSummary{}, err
	}
	if state != string(graphprojection.RunStateAvailable) && state != string(graphprojection.RunStateReplaced) {
		return graphprojection.InvalidationSummary{}, &graphprojection.OperationError{Code: "invalid_operation", ReasonCode: "invalid_invalidation_target", Details: map[string]any{"operation": "invalidate_projection_run", "reason_code": "invalid_invalidation_target"}}
	}
	graphViewStateAfter := graphprojection.GraphViewStateAvailable
	var selectedRunID string
	var currentViewState string
	var retentionPolicyJSON []byte
	var currentUpdatedAt time.Time
	if err := tx.QueryRow(ctx, `
SELECT COALESCE(graph_view.selected_projection_run_id, graph_view.latest_projection_run_id, ''),
       graph_view.state,
       selected_run.retention_policy_json,
       graph_view.updated_at
  FROM graph_projection_views AS graph_view
  JOIN graph_projection_runs AS selected_run
    ON selected_run.projection_run_id = COALESCE(graph_view.selected_projection_run_id, graph_view.latest_projection_run_id)
 WHERE graph_view.graph_view_id = $1
 FOR UPDATE OF graph_view, selected_run
`, graphViewID).Scan(&selectedRunID, &currentViewState, &retentionPolicyJSON, &currentUpdatedAt); err != nil {
		return graphprojection.InvalidationSummary{}, err
	}
	var retentionPolicy graphprojection.RetentionPolicy
	if err := json.Unmarshal(retentionPolicyJSON, &retentionPolicy); err != nil {
		return graphprojection.InvalidationSummary{}, fmt.Errorf("decode graph projection retention policy: %w", err)
	}
	if !invalidatedAt.After(currentUpdatedAt) {
		invalidatedAt = currentUpdatedAt.Add(time.Microsecond)
	}
	if currentViewState == string(graphprojection.GraphViewStateInvalidated) || selectedRunID == projectionRunID {
		graphViewStateAfter = graphprojection.GraphViewStateInvalidated
	}
	summary := graphprojection.InvalidationSummary{GraphViewID: graphViewID, TargetScope: "projection_run", TargetProjectionRunID: &projectionRunID, InvalidatedRunIDs: []string{projectionRunID}, GraphViewStateAfter: graphViewStateAfter, InvalidatedAt: graphprojection.FormatLifecycleTimestamp(invalidatedAt), ReasonCode: reasonCode, RequestedAt: requestedAt, RequestedBy: requestedBy}
	if idempotencyKey != "" {
		expiresAt := graphprojection.FormatLifecycleTimestamp(invalidatedAt.Add(24 * time.Hour))
		summary.IdempotencyExpiresAt = &expiresAt
	}
	invalidationJSON, err := json.Marshal(graphprojection.Invalidation{InvalidatedAt: summary.InvalidatedAt, ReasonCode: reasonCode, RequestedBy: requestedBy, TargetScope: "projection_run", TargetProjectionRunID: &projectionRunID})
	if err != nil {
		return graphprojection.InvalidationSummary{}, err
	}
	if _, err := tx.Exec(ctx, `UPDATE graph_projection_runs SET state = 'invalidated', completed_at = COALESCE(completed_at, $3), invalidated_at = $3, retention_expires_at = CASE WHEN $2 = $5 THEN NULL ELSE $3::timestamptz + make_interval(secs => $6::integer) END, invalidation_json = $4::jsonb WHERE graph_view_id = $1 AND projection_run_id = $2`, graphViewID, projectionRunID, invalidatedAt, string(invalidationJSON), selectedRunID, retentionPolicy.RetentionDurationSeconds); err != nil {
		return graphprojection.InvalidationSummary{}, err
	}
	if graphViewStateAfter == graphprojection.GraphViewStateInvalidated {
		if _, err := tx.Exec(ctx, `UPDATE graph_projection_views SET state = 'invalidated', validation_status = 'invalidated', updated_at = $2, invalidation_json = $3::jsonb WHERE graph_view_id = $1`, graphViewID, invalidatedAt, string(invalidationJSON)); err != nil {
			return graphprojection.InvalidationSummary{}, err
		}
	}
	if err := s.pruneInvalidatedRetentionTx(ctx, tx, graphViewID, selectedRunID, retentionPolicy, invalidatedAt); err != nil {
		return graphprojection.InvalidationSummary{}, err
	}
	if idempotencyKey != "" {
		scopeKey := graphViewID + "\n" + projectionRunID
		if err := s.recordInvalidationIdempotencyTx(ctx, tx, "invalidate_projection_run", scopeKey, idempotencyKey, fingerprint, summary, invalidatedAt); err != nil {
			return graphprojection.InvalidationSummary{}, err
		}
	}
	if err := tx.Commit(ctx); err != nil {
		return graphprojection.InvalidationSummary{}, err
	}
	return summary, nil
}

func (s *Store) pruneRetentionTx(ctx context.Context, tx pgx.Tx, graphViewID string, policy graphprojection.RetentionPolicy, queryReceivedAt time.Time) error {
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

func invalidationReplayFingerprint(operation, graphViewID, projectionRunID, reasonCode, requestedBy string) (string, error) {
	comparison := map[string]any{"operation": operation, "graph_view_id": graphViewID, "reason_code": reasonCode, "requested_by": requestedBy}
	if projectionRunID != "" {
		comparison["projection_run_id"] = projectionRunID
	}
	bytes, err := graphprojection.CanonicalJSON(comparison)
	if err != nil {
		return "", err
	}
	return string(bytes), nil
}

func (s *Store) checkInvalidationIdempotencyTx(ctx context.Context, tx pgx.Tx, operation, scopeKey, key, fingerprint string, queryReceivedAt time.Time) (graphprojection.InvalidationSummary, bool, error) {
	row := tx.QueryRow(ctx, `SELECT request_fingerprint, response_json, expires_at FROM graph_projection_idempotency WHERE operation = $1 AND scope_key = $2 AND idempotency_key = $3 FOR UPDATE`, operation, scopeKey, key)
	var existingFingerprint string
	var responseJSON []byte
	var expiresAt time.Time
	if err := row.Scan(&existingFingerprint, &responseJSON, &expiresAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return graphprojection.InvalidationSummary{}, false, nil
		}
		return graphprojection.InvalidationSummary{}, false, fmt.Errorf("load graph projection invalidation idempotency: %w", err)
	}
	if !queryReceivedAt.Before(expiresAt) {
		if _, err := tx.Exec(ctx, `DELETE FROM graph_projection_idempotency WHERE operation = $1 AND scope_key = $2 AND idempotency_key = $3`, operation, scopeKey, key); err != nil {
			return graphprojection.InvalidationSummary{}, false, err
		}
		return graphprojection.InvalidationSummary{}, false, nil
	}
	if existingFingerprint != fingerprint {
		return graphprojection.InvalidationSummary{}, false, &graphprojection.OperationError{Code: "operation_conflict", ReasonCode: "idempotency_key_conflict", Details: map[string]any{"operation": operation, "reason_code": "idempotency_key_conflict"}}
	}
	var summary graphprojection.InvalidationSummary
	if err := json.Unmarshal(responseJSON, &summary); err != nil {
		return graphprojection.InvalidationSummary{}, false, fmt.Errorf("decode graph projection invalidation replay: %w", err)
	}
	return summary, true, nil
}

func (s *Store) recordInvalidationIdempotencyTx(ctx context.Context, tx pgx.Tx, operation, scopeKey, key, fingerprint string, summary graphprojection.InvalidationSummary, createdAt time.Time) error {
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

func (s *Store) checkIdempotencyTx(ctx context.Context, tx pgx.Tx, operation, scopeKey, key, fingerprint string, queryReceivedAt time.Time) (graphprojection.ProjectionRun, bool, error) {
	row := tx.QueryRow(ctx, `SELECT request_fingerprint, response_json, expires_at FROM graph_projection_idempotency WHERE operation = $1 AND scope_key = $2 AND idempotency_key = $3 FOR UPDATE`, operation, scopeKey, key)
	var existingFingerprint string
	var responseJSON []byte
	var expiresAt time.Time
	if err := row.Scan(&existingFingerprint, &responseJSON, &expiresAt); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return graphprojection.ProjectionRun{}, false, nil
		}
		return graphprojection.ProjectionRun{}, false, fmt.Errorf("load graph projection idempotency: %w", err)
	}
	if !queryReceivedAt.Before(expiresAt) {
		if _, err := tx.Exec(ctx, `DELETE FROM graph_projection_idempotency WHERE operation = $1 AND scope_key = $2 AND idempotency_key = $3`, operation, scopeKey, key); err != nil {
			return graphprojection.ProjectionRun{}, false, err
		}
		return graphprojection.ProjectionRun{}, false, nil
	}
	if existingFingerprint != fingerprint {
		return graphprojection.ProjectionRun{}, false, &graphprojection.OperationError{Code: "operation_conflict", ReasonCode: "idempotency_key_conflict", Details: map[string]any{"operation": operation, "reason_code": "idempotency_key_conflict"}}
	}
	var response graphprojection.AcceptedRunSummary
	if err := json.Unmarshal(responseJSON, &response); err != nil {
		return graphprojection.ProjectionRun{}, false, fmt.Errorf("decode graph projection accepted replay: %w", err)
	}
	acceptedAt, err := time.Parse("2006-01-02T15:04:05.999999Z", response.AcceptedAt)
	if err != nil {
		return graphprojection.ProjectionRun{}, false, fmt.Errorf("decode graph projection accepted replay timestamp: %w", err)
	}
	run := graphprojection.ProjectionRun{GraphViewID: response.GraphViewID, ProjectionRunID: response.ProjectionRunID, State: graphprojection.RunStateAccepted, AcceptedAt: acceptedAt, AcceptedReplay: &response}
	run.Request.SourceSnapshotID = response.SourceSnapshotID
	run.Request.ProjectionConfig.ProjectionVersion = response.ProjectionVersion
	return run, true, nil
}

func retainedReplayFingerprint(operation string, run graphprojection.ProjectionRun) (string, error) {
	bytes, err := graphprojection.CanonicalJSON(map[string]any{
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

func (s *Store) recordIdempotencyTx(ctx context.Context, tx pgx.Tx, operation, scopeKey, key, fingerprint string, run graphprojection.ProjectionRun) error {
	expiresAt := graphprojection.FormatLifecycleTimestamp(run.AcceptedAt.Add(24 * time.Hour))
	responseJSON, err := json.Marshal(graphprojection.AcceptedRunSummary{
		GraphViewID:          run.GraphViewID,
		ProjectionRunID:      run.ProjectionRunID,
		State:                graphprojection.RunStateAccepted,
		SourceSnapshotID:     run.Request.SourceSnapshotID,
		ProjectionVersion:    run.Request.ProjectionConfig.ProjectionVersion,
		AcceptedAt:           graphprojection.FormatLifecycleTimestamp(run.AcceptedAt),
		IdempotencyExpiresAt: &expiresAt,
	})
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
	return graphprojection.NewOperationError("cursor_invalid", reason, map[string]any{"reason_code": reason}, graphprojection.ErrCursorInvalid)
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
