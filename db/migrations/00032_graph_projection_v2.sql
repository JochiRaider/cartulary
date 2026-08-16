-- +goose Up
--
-- Network Flow owns saved-graph declarations. Graph Projection owns only the
-- immutable derived result tables below. The declaration deliberately has no
-- foreign key to graph_projection_results: Recovery restores authority before
-- rebuilding derived state.
--

CREATE TABLE public.network_flow_graph_views (
    graph_view_id text NOT NULL,
    incident_id uuid NOT NULL,
    display_name text NOT NULL,
    normalized_display_name text NOT NULL,
    declaration_state text NOT NULL,
    semantic_query_json jsonb NOT NULL,
    semantic_query_sha256 text NOT NULL,
    desired_source_snapshot_id text NOT NULL,
    selected_projection_result_id text,
    selected_source_snapshot_id text,
    selected_projection_schema_id text,
    selected_projection_version text,
    selected_normalized_configuration_sha256 text,
    selected_normalized_source_sha256 text,
    selected_canonical_output_sha256 text,
    graph_view_version bigint NOT NULL,
    materialization_generation bigint NOT NULL,
    created_by_user_id uuid NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    retired_at timestamp with time zone,
    latest_job_id uuid,
    last_failure_code text,
    last_failed_at timestamp with time zone,
    CONSTRAINT network_flow_graph_views_pkey PRIMARY KEY (graph_view_id),
    CONSTRAINT network_flow_graph_views_id_ck CHECK (graph_view_id ~ '^nfgv_[a-f0-9]{32}$'::text),
    CONSTRAINT network_flow_graph_views_display_name_ck CHECK (
        char_length(display_name) BETWEEN 1 AND 64
        AND display_name !~ '[[:cntrl:]]'::text
        AND char_length(normalized_display_name) BETWEEN 1 AND 64
        AND normalized_display_name !~ '[[:cntrl:]]'::text
    ),
    CONSTRAINT network_flow_graph_views_state_ck CHECK (declaration_state IN ('active', 'retired')),
    CONSTRAINT network_flow_graph_views_state_time_ck CHECK (
        (declaration_state = 'active' AND retired_at IS NULL)
        OR (declaration_state = 'retired' AND retired_at IS NOT NULL)
    ),
    CONSTRAINT network_flow_graph_views_semantic_query_ck CHECK (
        jsonb_typeof(semantic_query_json) = 'object'
        AND semantic_query_sha256 ~ '^[a-f0-9]{64}$'::text
    ),
    CONSTRAINT network_flow_graph_views_source_snapshot_ck CHECK (
        char_length(desired_source_snapshot_id) BETWEEN 1 AND 255
        AND desired_source_snapshot_id !~ '[[:cntrl:]]'::text
    ),
    CONSTRAINT network_flow_graph_views_selected_binding_ck CHECK (
        (selected_projection_result_id IS NULL
            AND selected_source_snapshot_id IS NULL
            AND selected_projection_schema_id IS NULL
            AND selected_projection_version IS NULL
            AND selected_normalized_configuration_sha256 IS NULL
            AND selected_normalized_source_sha256 IS NULL
            AND selected_canonical_output_sha256 IS NULL)
        OR
        (selected_projection_result_id ~ '^gpres_[a-f0-9]{64}$'::text
            AND selected_source_snapshot_id IS NOT NULL
            AND selected_projection_schema_id = 'graph_projection.v2'
            AND selected_projection_version IS NOT NULL
            AND selected_normalized_configuration_sha256 ~ '^[a-f0-9]{64}$'::text
            AND selected_normalized_source_sha256 ~ '^[a-f0-9]{64}$'::text
            AND selected_canonical_output_sha256 ~ '^[a-f0-9]{64}$'::text)
    ),
    CONSTRAINT network_flow_graph_views_versions_ck CHECK (
        graph_view_version >= 1 AND materialization_generation >= 1
    ),
    CONSTRAINT network_flow_graph_views_failure_ck CHECK (
        (last_failure_code IS NULL AND last_failed_at IS NULL)
        OR (last_failure_code ~ '^network_flow_[a-z0-9_]+$'::text AND last_failed_at IS NOT NULL)
    ),
    CONSTRAINT network_flow_graph_views_times_ck CHECK (
        updated_at >= created_at
        AND (retired_at IS NULL OR retired_at >= created_at)
        AND (last_failed_at IS NULL OR last_failed_at >= created_at)
    )
);

ALTER TABLE ONLY public.network_flow_graph_views
    ADD CONSTRAINT network_flow_graph_views_incident_id_fkey
        FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON UPDATE NO ACTION ON DELETE CASCADE,
    ADD CONSTRAINT network_flow_graph_views_created_by_user_id_fkey
        FOREIGN KEY (created_by_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION,
    ADD CONSTRAINT network_flow_graph_views_latest_job_id_fkey
        FOREIGN KEY (latest_job_id) REFERENCES public.jobs(job_id) ON UPDATE NO ACTION ON DELETE NO ACTION;

CREATE INDEX network_flow_graph_views_active_order_idx
    ON public.network_flow_graph_views (incident_id, normalized_display_name, graph_view_id)
    WHERE declaration_state = 'active';

CREATE INDEX network_flow_graph_views_retained_count_idx
    ON public.network_flow_graph_views (incident_id, declaration_state, graph_view_id);

CREATE INDEX network_flow_graph_views_selected_result_idx
    ON public.network_flow_graph_views (selected_projection_result_id)
    WHERE selected_projection_result_id IS NOT NULL;

CREATE INDEX network_flow_graph_views_created_by_user_id_fk_idx
    ON public.network_flow_graph_views (created_by_user_id);

CREATE INDEX network_flow_graph_views_latest_job_id_fk_idx
    ON public.network_flow_graph_views (latest_job_id);

CREATE TABLE public.graph_projection_results (
    projection_result_id text NOT NULL,
    graph_view_id text NOT NULL,
    source_owner_id text NOT NULL,
    source_snapshot_id text NOT NULL,
    projection_schema_id text NOT NULL,
    projection_version text NOT NULL,
    normalized_configuration_sha256 text NOT NULL,
    normalized_source_sha256 text NOT NULL,
    canonical_output_sha256 text NOT NULL,
    vertex_count bigint NOT NULL,
    edge_count bigint NOT NULL,
    result_json bytea NOT NULL,
    published_at timestamp with time zone NOT NULL,
    CONSTRAINT graph_projection_results_pkey PRIMARY KEY (projection_result_id),
    CONSTRAINT graph_projection_results_id_ck CHECK (projection_result_id ~ '^gpres_[a-f0-9]{64}$'::text),
    CONSTRAINT graph_projection_results_identity_ck CHECK (
        char_length(graph_view_id) BETWEEN 1 AND 255
        AND char_length(source_owner_id) BETWEEN 1 AND 255
        AND char_length(source_snapshot_id) BETWEEN 1 AND 255
        AND projection_schema_id = 'graph_projection.v2'
        AND char_length(projection_version) BETWEEN 1 AND 255
        AND normalized_configuration_sha256 ~ '^[a-f0-9]{64}$'::text
        AND normalized_source_sha256 ~ '^[a-f0-9]{64}$'::text
        AND canonical_output_sha256 ~ '^[a-f0-9]{64}$'::text
    ),
    CONSTRAINT graph_projection_results_counts_ck CHECK (
        vertex_count BETWEEN 0 AND 100000
        AND edge_count BETWEEN 0 AND 250000
    ),
    CONSTRAINT graph_projection_results_json_ck CHECK (octet_length(result_json) >= 2),
    CONSTRAINT graph_projection_results_binding_key UNIQUE (
        projection_result_id,
        graph_view_id,
        source_owner_id,
        source_snapshot_id,
        projection_schema_id,
        projection_version,
        normalized_configuration_sha256,
        normalized_source_sha256,
        canonical_output_sha256
    )
);

CREATE TABLE public.graph_projection_result_vertices (
    projection_result_id text NOT NULL,
    vertex_id text NOT NULL,
    vertex_kind text NOT NULL,
    sort_ordinal bigint NOT NULL,
    sort_key text NOT NULL,
    vertex_json bytea NOT NULL,
    CONSTRAINT graph_projection_result_vertices_pkey PRIMARY KEY (projection_result_id, vertex_id),
    CONSTRAINT graph_projection_result_vertices_id_ck CHECK (vertex_id ~ '^vx_[a-f0-9]{64}$'::text),
    CONSTRAINT graph_projection_result_vertices_kind_ck CHECK (char_length(vertex_kind) BETWEEN 1 AND 255),
    CONSTRAINT graph_projection_result_vertices_ordinal_ck CHECK (sort_ordinal >= 0),
    CONSTRAINT graph_projection_result_vertices_json_ck CHECK (octet_length(vertex_json) >= 2),
    CONSTRAINT graph_projection_result_vertices_result_ordinal_key UNIQUE (projection_result_id, sort_ordinal)
);

ALTER TABLE ONLY public.graph_projection_result_vertices
    ADD CONSTRAINT graph_projection_result_vertices_result_id_fkey
        FOREIGN KEY (projection_result_id) REFERENCES public.graph_projection_results(projection_result_id)
        ON UPDATE NO ACTION ON DELETE CASCADE;

CREATE INDEX graph_projection_result_vertices_order_idx
    ON public.graph_projection_result_vertices (projection_result_id, sort_key, vertex_id);

CREATE TABLE public.graph_projection_result_edges (
    projection_result_id text NOT NULL,
    edge_id text NOT NULL,
    edge_kind text NOT NULL,
    src_vertex_id text NOT NULL,
    dst_vertex_id text NOT NULL,
    direction text NOT NULL,
    sort_ordinal bigint NOT NULL,
    sort_key text NOT NULL,
    edge_json bytea NOT NULL,
    CONSTRAINT graph_projection_result_edges_pkey PRIMARY KEY (projection_result_id, edge_id),
    CONSTRAINT graph_projection_result_edges_id_ck CHECK (edge_id ~ '^ed_[a-f0-9]{64}$'::text),
    CONSTRAINT graph_projection_result_edges_kind_ck CHECK (char_length(edge_kind) BETWEEN 1 AND 255),
    CONSTRAINT graph_projection_result_edges_direction_ck CHECK (direction IN ('directed', 'undirected', 'bidirectional')),
    CONSTRAINT graph_projection_result_edges_ordinal_ck CHECK (sort_ordinal >= 0),
    CONSTRAINT graph_projection_result_edges_json_ck CHECK (octet_length(edge_json) >= 2),
    CONSTRAINT graph_projection_result_edges_result_ordinal_key UNIQUE (projection_result_id, sort_ordinal)
);

ALTER TABLE ONLY public.graph_projection_result_edges
    ADD CONSTRAINT graph_projection_result_edges_result_id_fkey
        FOREIGN KEY (projection_result_id) REFERENCES public.graph_projection_results(projection_result_id)
        ON UPDATE NO ACTION ON DELETE CASCADE,
    ADD CONSTRAINT graph_projection_result_edges_src_vertex_fkey
        FOREIGN KEY (projection_result_id, src_vertex_id)
        REFERENCES public.graph_projection_result_vertices(projection_result_id, vertex_id)
        ON UPDATE NO ACTION ON DELETE CASCADE,
    ADD CONSTRAINT graph_projection_result_edges_dst_vertex_fkey
        FOREIGN KEY (projection_result_id, dst_vertex_id)
        REFERENCES public.graph_projection_result_vertices(projection_result_id, vertex_id)
        ON UPDATE NO ACTION ON DELETE CASCADE;

CREATE INDEX graph_projection_result_edges_order_idx
    ON public.graph_projection_result_edges (projection_result_id, sort_key, edge_id);

CREATE INDEX graph_projection_result_edges_src_traversal_idx
    ON public.graph_projection_result_edges (projection_result_id, src_vertex_id, sort_ordinal, edge_id);

CREATE INDEX graph_projection_result_edges_dst_traversal_idx
    ON public.graph_projection_result_edges (projection_result_id, dst_vertex_id, sort_ordinal, edge_id);

CREATE TABLE public.graph_projection_result_leases (
    lease_id uuid NOT NULL,
    projection_result_id text NOT NULL,
    lease_owner_id text NOT NULL,
    lease_owner_resource_id text NOT NULL,
    lease_purpose text NOT NULL,
    leased_until timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    renewed_at timestamp with time zone NOT NULL,
    CONSTRAINT graph_projection_result_leases_pkey PRIMARY KEY (lease_id),
    CONSTRAINT graph_projection_result_leases_owner_ck CHECK (
        char_length(lease_owner_id) BETWEEN 1 AND 255
        AND char_length(lease_owner_resource_id) BETWEEN 1 AND 255
        AND lease_purpose ~ '^[a-z][a-z0-9_.:-]{0,254}$'::text
    ),
    CONSTRAINT graph_projection_result_leases_time_ck CHECK (
        renewed_at >= created_at AND leased_until > renewed_at
    ),
    CONSTRAINT graph_projection_result_leases_identity_key UNIQUE (
        projection_result_id,
        lease_owner_id,
        lease_owner_resource_id,
        lease_purpose
    )
);

ALTER TABLE ONLY public.graph_projection_result_leases
    ADD CONSTRAINT graph_projection_result_leases_result_id_fkey
        FOREIGN KEY (projection_result_id) REFERENCES public.graph_projection_results(projection_result_id)
        ON UPDATE NO ACTION ON DELETE RESTRICT;

CREATE INDEX graph_projection_result_leases_expiry_idx
    ON public.graph_projection_result_leases (leased_until, projection_result_id);

-- +goose StatementBegin
CREATE FUNCTION public.graph_projection_reject_immutable_result_update()
RETURNS trigger
LANGUAGE plpgsql
SET search_path = pg_catalog, public
AS $$
BEGIN
    RAISE EXCEPTION 'Graph Projection v2 result rows are immutable'
        USING ERRCODE = '23514';
END;
$$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION public.graph_projection_reject_immutable_result_update() FROM PUBLIC;
REVOKE USAGE ON TYPE
    public.network_flow_graph_views,
    public.graph_projection_results,
    public.graph_projection_result_vertices,
    public.graph_projection_result_edges,
    public.graph_projection_result_leases
FROM PUBLIC;
GRANT USAGE ON TYPE
    public.network_flow_graph_views,
    public.graph_projection_results,
    public.graph_projection_result_vertices,
    public.graph_projection_result_edges,
    public.graph_projection_result_leases
TO cartulary_schema_owner, cartulary_runtime, cartulary_recovery;

CREATE TRIGGER graph_projection_results_immutable_update
    BEFORE UPDATE ON public.graph_projection_results
    FOR EACH ROW EXECUTE FUNCTION public.graph_projection_reject_immutable_result_update();

CREATE TRIGGER graph_projection_result_vertices_immutable_update
    BEFORE UPDATE ON public.graph_projection_result_vertices
    FOR EACH ROW EXECUTE FUNCTION public.graph_projection_reject_immutable_result_update();

CREATE TRIGGER graph_projection_result_edges_immutable_update
    BEFORE UPDATE ON public.graph_projection_result_edges
    FOR EACH ROW EXECUTE FUNCTION public.graph_projection_reject_immutable_result_update();

GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE public.network_flow_graph_views TO cartulary_runtime;
GRANT SELECT, INSERT, UPDATE, DELETE, TRUNCATE ON TABLE public.network_flow_graph_views TO cartulary_recovery;

GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE public.graph_projection_results TO cartulary_runtime;
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE public.graph_projection_result_vertices TO cartulary_runtime;
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE public.graph_projection_result_edges TO cartulary_runtime;
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE public.graph_projection_result_leases TO cartulary_runtime;

GRANT SELECT, INSERT, UPDATE, DELETE, TRUNCATE ON TABLE public.graph_projection_results TO cartulary_recovery;
GRANT SELECT, INSERT, UPDATE, DELETE, TRUNCATE ON TABLE public.graph_projection_result_vertices TO cartulary_recovery;
GRANT SELECT, INSERT, UPDATE, DELETE, TRUNCATE ON TABLE public.graph_projection_result_edges TO cartulary_recovery;
GRANT SELECT, INSERT, UPDATE, DELETE, TRUNCATE ON TABLE public.graph_projection_result_leases TO cartulary_recovery;

-- +goose Down
DROP TRIGGER graph_projection_result_edges_immutable_update ON public.graph_projection_result_edges;
DROP TRIGGER graph_projection_result_vertices_immutable_update ON public.graph_projection_result_vertices;
DROP TRIGGER graph_projection_results_immutable_update ON public.graph_projection_results;
DROP FUNCTION public.graph_projection_reject_immutable_result_update();

DROP TABLE public.graph_projection_result_leases;
DROP TABLE public.graph_projection_result_edges;
DROP TABLE public.graph_projection_result_vertices;
DROP TABLE public.graph_projection_results;
DROP TABLE public.network_flow_graph_views;
