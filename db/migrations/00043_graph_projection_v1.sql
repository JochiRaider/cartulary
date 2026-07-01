-- +goose Up
CREATE TABLE IF NOT EXISTS graph_projection_views (
    graph_view_id text PRIMARY KEY,
    graph_view_key text NOT NULL UNIQUE,
    state text NOT NULL,
    latest_projection_run_id text,
    latest_source_snapshot_id text,
    projection_version text,
    updated_at timestamptz NOT NULL,
    validation_status text NOT NULL
);

CREATE TABLE IF NOT EXISTS graph_projection_runs (
    projection_run_id text PRIMARY KEY,
    graph_view_id text NOT NULL REFERENCES graph_projection_views(graph_view_id) ON DELETE CASCADE,
    source_snapshot_id text NOT NULL,
    projection_version text NOT NULL,
    state text NOT NULL,
    projection_run_nonce text NOT NULL,
    projection_config_digest text NOT NULL,
    projection_source_digest text NOT NULL,
    accepted_at timestamptz NOT NULL,
    completed_at timestamptz,
    validation_summary_json jsonb NOT NULL,
    failure_reason text,
    graph_view_json jsonb,
    invalidation_json jsonb,
    retention_expires_at timestamptz
);

CREATE INDEX IF NOT EXISTS graph_projection_runs_view_state_idx
    ON graph_projection_runs (graph_view_id, state, completed_at DESC, projection_run_id ASC);

CREATE TABLE IF NOT EXISTS graph_projection_vertices (
    projection_run_id text NOT NULL REFERENCES graph_projection_runs(projection_run_id) ON DELETE CASCADE,
    graph_view_id text NOT NULL,
    vertex_id text NOT NULL,
    vertex_kind text NOT NULL,
    sort_key text NOT NULL,
    vertex_json jsonb NOT NULL,
    PRIMARY KEY (projection_run_id, vertex_id)
);

CREATE INDEX IF NOT EXISTS graph_projection_vertices_view_sort_idx
    ON graph_projection_vertices (graph_view_id, sort_key, vertex_id);

CREATE TABLE IF NOT EXISTS graph_projection_edges (
    projection_run_id text NOT NULL REFERENCES graph_projection_runs(projection_run_id) ON DELETE CASCADE,
    graph_view_id text NOT NULL,
    edge_id text NOT NULL,
    edge_kind text NOT NULL,
    src_vertex_id text NOT NULL,
    dst_vertex_id text NOT NULL,
    direction text NOT NULL,
    sort_key text NOT NULL,
    edge_json jsonb NOT NULL,
    PRIMARY KEY (projection_run_id, edge_id)
);

CREATE INDEX IF NOT EXISTS graph_projection_edges_view_sort_idx
    ON graph_projection_edges (graph_view_id, sort_key, edge_id);

CREATE INDEX IF NOT EXISTS graph_projection_edges_endpoint_idx
    ON graph_projection_edges (projection_run_id, src_vertex_id, dst_vertex_id, direction);

CREATE TABLE IF NOT EXISTS graph_projection_idempotency (
    operation text NOT NULL,
    idempotency_key text NOT NULL,
    request_fingerprint text NOT NULL,
    graph_view_id text,
    projection_run_id text,
    response_json jsonb NOT NULL,
    created_at timestamptz NOT NULL,
    expires_at timestamptz NOT NULL,
    PRIMARY KEY (operation, idempotency_key)
);

CREATE INDEX IF NOT EXISTS graph_projection_idempotency_expiry_idx
    ON graph_projection_idempotency (expires_at);

-- +goose Down
DROP INDEX IF EXISTS graph_projection_idempotency_expiry_idx;
DROP TABLE IF EXISTS graph_projection_idempotency;

DROP INDEX IF EXISTS graph_projection_edges_endpoint_idx;
DROP INDEX IF EXISTS graph_projection_edges_view_sort_idx;
DROP TABLE IF EXISTS graph_projection_edges;

DROP INDEX IF EXISTS graph_projection_vertices_view_sort_idx;
DROP TABLE IF EXISTS graph_projection_vertices;

DROP INDEX IF EXISTS graph_projection_runs_view_state_idx;
DROP TABLE IF EXISTS graph_projection_runs;

DROP TABLE IF EXISTS graph_projection_views;
