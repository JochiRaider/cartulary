-- +goose Up
--
-- Graph Projection v1 is removed only after a fail-closed cutover preflight.
-- The exception reports counts, never stored values. A deployment that fails
-- this gate must remain on the pre-cutover binary/schema pair until the data is
-- explicitly dispositioned and a replacement-target rollback backup exists.
--

-- +goose StatementBegin
DO $$
DECLARE
    legacy_edge_count bigint;
    legacy_idempotency_count bigint;
    legacy_run_count bigint;
    legacy_vertex_count bigint;
    legacy_view_count bigint;
    legacy_release_ref_count bigint;
    malformed_job_ref_count bigint;
    legacy_job_ref_count bigint;
BEGIN
    SELECT count(*) INTO legacy_edge_count FROM public.graph_projection_edges;
    SELECT count(*) INTO legacy_idempotency_count FROM public.graph_projection_idempotency;
    SELECT count(*) INTO legacy_run_count FROM public.graph_projection_runs;
    SELECT count(*) INTO legacy_vertex_count FROM public.graph_projection_vertices;
    SELECT count(*) INTO legacy_view_count FROM public.graph_projection_views;

    SELECT count(*)
      INTO legacy_release_ref_count
      FROM public.reporting_releases release_row
      CROSS JOIN LATERAL jsonb_array_elements(release_row.graph_projection_refs) AS projection_ref
     WHERE projection_ref ? 'projection_run_id'
        OR projection_ref ? 'projection_config_digest'
        OR projection_ref ? 'projection_source_digest'
        OR projection_ref ? 'projection_output_digest'
        OR projection_ref ->> 'projection_schema_id' = 'graph_projection.v1';

    SELECT count(*)
      INTO malformed_job_ref_count
      FROM public.reporting_job_payloads payload
     WHERE payload.request_json ? 'graph_projection_refs'
       AND jsonb_typeof(payload.request_json -> 'graph_projection_refs') <> 'array';

    SELECT count(*)
      INTO legacy_job_ref_count
      FROM public.reporting_job_payloads payload
      CROSS JOIN LATERAL jsonb_array_elements(
          CASE
              WHEN jsonb_typeof(payload.request_json -> 'graph_projection_refs') = 'array'
                  THEN payload.request_json -> 'graph_projection_refs'
              ELSE '[]'::jsonb
          END
      ) AS projection_ref
     WHERE projection_ref ? 'projection_run_id'
        OR projection_ref ? 'projection_config_digest'
        OR projection_ref ? 'projection_source_digest'
        OR projection_ref ? 'projection_output_digest'
        OR projection_ref ->> 'projection_schema_id' = 'graph_projection.v1';

    IF legacy_edge_count <> 0
        OR legacy_idempotency_count <> 0
        OR legacy_run_count <> 0
        OR legacy_vertex_count <> 0
        OR legacy_view_count <> 0
        OR legacy_release_ref_count <> 0
        OR malformed_job_ref_count <> 0
        OR legacy_job_ref_count <> 0
    THEN
        RAISE EXCEPTION USING
            ERRCODE = '23514',
            MESSAGE = 'graph_projection_v1_cutover_preflight_failed',
            DETAIL = format(
                'edges=%s idempotency=%s runs=%s vertices=%s views=%s reporting_release_refs=%s malformed_reporting_job_refs=%s reporting_job_refs=%s',
                legacy_edge_count,
                legacy_idempotency_count,
                legacy_run_count,
                legacy_vertex_count,
                legacy_view_count,
                legacy_release_ref_count,
                malformed_job_ref_count,
                legacy_job_ref_count
            ),
            HINT = 'Keep the pre-cutover binary/schema pair active, disposition unsupported v1 state, and preserve the exact replacement-target rollback backup before retrying.';
    END IF;
END;
$$;
-- +goose StatementEnd

ALTER TABLE public.graph_projection_views
    DROP CONSTRAINT graph_projection_views_selected_run_fkey;

DROP TABLE public.graph_projection_edges;
DROP TABLE public.graph_projection_vertices;
DROP TABLE public.graph_projection_runs;
DROP TABLE public.graph_projection_views;
DROP TABLE public.graph_projection_idempotency;

-- +goose Down
-- This recreates only an empty v1 schema for disposable migration mechanics.
-- It does not restore deleted v1 data and is not an operational rollback.

CREATE TABLE public.graph_projection_edges (
    projection_run_id text NOT NULL,
    graph_view_id text NOT NULL,
    edge_id text NOT NULL,
    edge_kind text NOT NULL,
    src_vertex_id text NOT NULL,
    dst_vertex_id text NOT NULL,
    direction text NOT NULL,
    sort_key text NOT NULL,
    edge_json jsonb NOT NULL
);

CREATE TABLE public.graph_projection_idempotency (
    operation text NOT NULL,
    scope_key text NOT NULL,
    idempotency_key text NOT NULL,
    request_fingerprint text NOT NULL,
    graph_view_id text,
    projection_run_id text,
    response_json jsonb NOT NULL,
    created_at timestamp with time zone NOT NULL,
    expires_at timestamp with time zone NOT NULL
);

CREATE TABLE public.graph_projection_runs (
    projection_run_id text NOT NULL,
    graph_view_id text NOT NULL,
    source_snapshot_id text NOT NULL,
    projection_version text NOT NULL,
    state text NOT NULL,
    projection_run_nonce text NOT NULL,
    projection_config_digest text NOT NULL,
    projection_source_digest text NOT NULL,
    accepted_at timestamp with time zone NOT NULL,
    started_at timestamp with time zone,
    generated_at timestamp with time zone,
    completed_at timestamp with time zone,
    replaced_at timestamp with time zone,
    invalidated_at timestamp with time zone,
    validation_summary_json jsonb NOT NULL,
    failure_reason text,
    graph_view_json jsonb,
    invalidation_json jsonb,
    retention_expires_at timestamp with time zone,
    retention_policy_json jsonb NOT NULL DEFAULT '{
      "retain_replaced_results": true,
      "retention_count": 5,
      "retention_duration_seconds": 2592000,
      "retain_failed_results": true,
      "failed_retention_count": 20,
      "failed_retention_duration_seconds": 2592000
    }'::jsonb,
    projection_output_digest text,
    CONSTRAINT graph_projection_runs_projection_output_digest_ck CHECK ((((state = ANY (ARRAY['available'::text, 'replaced'::text])) AND (projection_output_digest ~ '^[a-f0-9]{64}$'::text)) OR ((state <> ALL (ARRAY['available'::text, 'replaced'::text])) AND ((projection_output_digest IS NULL) OR (projection_output_digest ~ '^[a-f0-9]{64}$'::text))))),
    CONSTRAINT graph_projection_runs_state_ck CHECK (state = ANY (ARRAY['accepted', 'computing', 'available', 'failed', 'replaced', 'invalidated'])),
    CONSTRAINT graph_projection_runs_retention_policy_ck CHECK (jsonb_typeof(retention_policy_json) = 'object')
);

CREATE TABLE public.graph_projection_vertices (
    projection_run_id text NOT NULL,
    graph_view_id text NOT NULL,
    vertex_id text NOT NULL,
    vertex_kind text NOT NULL,
    sort_key text NOT NULL,
    vertex_json jsonb NOT NULL
);

CREATE TABLE public.graph_projection_views (
    graph_view_id text NOT NULL,
    graph_view_key text NOT NULL,
    state text NOT NULL,
    latest_projection_run_id text,
    latest_source_snapshot_id text,
    projection_version text,
    selected_projection_run_id text,
    updated_at timestamp with time zone NOT NULL,
    validation_status text NOT NULL,
    invalidation_json jsonb,
    CONSTRAINT graph_projection_views_state_ck CHECK (state = ANY (ARRAY['creating', 'available', 'refreshing', 'failed', 'invalidated'])),
    CONSTRAINT graph_projection_views_invalidation_ck CHECK (invalidation_json IS NULL OR jsonb_typeof(invalidation_json) = 'object')
);

ALTER TABLE ONLY public.graph_projection_edges
    ADD CONSTRAINT graph_projection_edges_pkey PRIMARY KEY (projection_run_id, edge_id);
ALTER TABLE ONLY public.graph_projection_idempotency
    ADD CONSTRAINT graph_projection_idempotency_pkey PRIMARY KEY (operation, scope_key, idempotency_key);
ALTER TABLE ONLY public.graph_projection_runs
    ADD CONSTRAINT graph_projection_runs_pkey PRIMARY KEY (projection_run_id);
ALTER TABLE ONLY public.graph_projection_vertices
    ADD CONSTRAINT graph_projection_vertices_pkey PRIMARY KEY (projection_run_id, vertex_id);
ALTER TABLE ONLY public.graph_projection_views
    ADD CONSTRAINT graph_projection_views_graph_view_key_key UNIQUE (graph_view_key);
ALTER TABLE ONLY public.graph_projection_views
    ADD CONSTRAINT graph_projection_views_pkey PRIMARY KEY (graph_view_id);

CREATE INDEX graph_projection_edges_endpoint_idx ON public.graph_projection_edges USING btree (projection_run_id, src_vertex_id, dst_vertex_id, direction);
CREATE INDEX graph_projection_edges_view_sort_idx ON public.graph_projection_edges USING btree (graph_view_id, sort_key, edge_id);
CREATE INDEX graph_projection_idempotency_expiry_idx ON public.graph_projection_idempotency USING btree (expires_at);
CREATE INDEX graph_projection_runs_view_state_idx ON public.graph_projection_runs USING btree (graph_view_id, state, completed_at DESC, projection_run_id);
CREATE INDEX graph_projection_vertices_view_sort_idx ON public.graph_projection_vertices USING btree (graph_view_id, sort_key, vertex_id);

ALTER TABLE ONLY public.graph_projection_edges
    ADD CONSTRAINT graph_projection_edges_projection_run_id_fkey FOREIGN KEY (projection_run_id) REFERENCES public.graph_projection_runs(projection_run_id) ON UPDATE NO ACTION ON DELETE CASCADE;
ALTER TABLE ONLY public.graph_projection_runs
    ADD CONSTRAINT graph_projection_runs_graph_view_id_fkey FOREIGN KEY (graph_view_id) REFERENCES public.graph_projection_views(graph_view_id) ON UPDATE NO ACTION ON DELETE CASCADE;
ALTER TABLE ONLY public.graph_projection_vertices
    ADD CONSTRAINT graph_projection_vertices_projection_run_id_fkey FOREIGN KEY (projection_run_id) REFERENCES public.graph_projection_runs(projection_run_id) ON UPDATE NO ACTION ON DELETE CASCADE;
ALTER TABLE public.graph_projection_views
    ADD CONSTRAINT graph_projection_views_selected_run_fkey
        FOREIGN KEY (selected_projection_run_id)
        REFERENCES public.graph_projection_runs(projection_run_id) ON UPDATE NO ACTION ON DELETE NO ACTION
        DEFERRABLE INITIALLY DEFERRED;

CREATE UNIQUE INDEX graph_projection_runs_one_active_per_view_idx
    ON public.graph_projection_runs (graph_view_id)
    WHERE state IN ('accepted', 'computing');
CREATE INDEX graph_projection_runs_replaced_retention_idx
    ON public.graph_projection_runs (graph_view_id, replaced_at DESC, projection_run_id ASC)
    WHERE state = 'replaced';
CREATE INDEX graph_projection_runs_failed_retention_idx
    ON public.graph_projection_runs (graph_view_id, completed_at DESC, projection_run_id ASC)
    WHERE state = 'failed';
CREATE INDEX graph_projection_runs_invalidated_retention_idx
    ON public.graph_projection_runs (graph_view_id, invalidated_at DESC, projection_run_id ASC)
    WHERE state = 'invalidated';
CREATE INDEX graph_projection_vertices_run_id_lookup_idx
    ON public.graph_projection_vertices (graph_view_id, projection_run_id, vertex_id);
CREATE INDEX graph_projection_edges_run_id_lookup_idx
    ON public.graph_projection_edges (graph_view_id, projection_run_id, edge_id);
CREATE INDEX graph_projection_views_selected_projection_run_id_fk_idx ON public.graph_projection_views (selected_projection_run_id);
