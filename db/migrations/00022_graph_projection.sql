-- +goose Up
-- PostgreSQL 18.6 consolidated baseline: graph projection.
-- Administrator provisioning and final admission are separate lifecycle steps.

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
    CONSTRAINT graph_projection_result_edges_direction_ck CHECK (direction = ANY (ARRAY['directed'::text, 'undirected'::text, 'bidirectional'::text])),
    CONSTRAINT graph_projection_result_edges_id_ck CHECK (edge_id ~ '^ed_[a-f0-9]{64}$'::text),
    CONSTRAINT graph_projection_result_edges_json_ck CHECK (octet_length(edge_json) >= 2),
    CONSTRAINT graph_projection_result_edges_kind_ck CHECK (char_length(edge_kind) >= 1 AND char_length(edge_kind) <= 255),
    CONSTRAINT graph_projection_result_edges_ordinal_ck CHECK (sort_ordinal >= 0),
    CONSTRAINT graph_projection_result_edges_pkey PRIMARY KEY (projection_result_id, edge_id),
    CONSTRAINT graph_projection_result_edges_result_ordinal_key UNIQUE (projection_result_id, sort_ordinal)
);

REVOKE ALL ON TABLE public.graph_projection_result_edges FROM PUBLIC;

REVOKE ALL ON TYPE public.graph_projection_result_edges FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.graph_projection_result_edges TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.graph_projection_result_edges TO cartulary_recovery;

GRANT USAGE ON TYPE public.graph_projection_result_edges TO cartulary_runtime;

GRANT USAGE ON TYPE public.graph_projection_result_edges TO cartulary_recovery;

CREATE TABLE public.graph_projection_result_leases (
    lease_id uuid NOT NULL,
    projection_result_id text NOT NULL,
    lease_owner_id text NOT NULL,
    lease_owner_resource_id text NOT NULL,
    lease_purpose text NOT NULL,
    leased_until timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    renewed_at timestamp with time zone NOT NULL,
    CONSTRAINT graph_projection_result_leases_identity_key UNIQUE (projection_result_id, lease_owner_id, lease_owner_resource_id, lease_purpose),
    CONSTRAINT graph_projection_result_leases_owner_ck CHECK (char_length(lease_owner_id) >= 1 AND char_length(lease_owner_id) <= 255 AND char_length(lease_owner_resource_id) >= 1 AND char_length(lease_owner_resource_id) <= 255 AND lease_purpose ~ '^[a-z][a-z0-9_.:-]{0,254}$'::text),
    CONSTRAINT graph_projection_result_leases_pkey PRIMARY KEY (lease_id),
    CONSTRAINT graph_projection_result_leases_time_ck CHECK (renewed_at >= created_at AND leased_until > renewed_at)
);

REVOKE ALL ON TABLE public.graph_projection_result_leases FROM PUBLIC;

REVOKE ALL ON TYPE public.graph_projection_result_leases FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.graph_projection_result_leases TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.graph_projection_result_leases TO cartulary_recovery;

GRANT USAGE ON TYPE public.graph_projection_result_leases TO cartulary_runtime;

GRANT USAGE ON TYPE public.graph_projection_result_leases TO cartulary_recovery;

CREATE TABLE public.graph_projection_result_vertices (
    projection_result_id text NOT NULL,
    vertex_id text NOT NULL,
    vertex_kind text NOT NULL,
    sort_ordinal bigint NOT NULL,
    sort_key text NOT NULL,
    vertex_json bytea NOT NULL,
    CONSTRAINT graph_projection_result_vertices_id_ck CHECK (vertex_id ~ '^vx_[a-f0-9]{64}$'::text),
    CONSTRAINT graph_projection_result_vertices_json_ck CHECK (octet_length(vertex_json) >= 2),
    CONSTRAINT graph_projection_result_vertices_kind_ck CHECK (char_length(vertex_kind) >= 1 AND char_length(vertex_kind) <= 255),
    CONSTRAINT graph_projection_result_vertices_ordinal_ck CHECK (sort_ordinal >= 0),
    CONSTRAINT graph_projection_result_vertices_pkey PRIMARY KEY (projection_result_id, vertex_id),
    CONSTRAINT graph_projection_result_vertices_result_ordinal_key UNIQUE (projection_result_id, sort_ordinal)
);

REVOKE ALL ON TABLE public.graph_projection_result_vertices FROM PUBLIC;

REVOKE ALL ON TYPE public.graph_projection_result_vertices FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.graph_projection_result_vertices TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.graph_projection_result_vertices TO cartulary_recovery;

GRANT USAGE ON TYPE public.graph_projection_result_vertices TO cartulary_runtime;

GRANT USAGE ON TYPE public.graph_projection_result_vertices TO cartulary_recovery;

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
    CONSTRAINT graph_projection_results_binding_key UNIQUE (projection_result_id, graph_view_id, source_owner_id, source_snapshot_id, projection_schema_id, projection_version, normalized_configuration_sha256, normalized_source_sha256, canonical_output_sha256),
    CONSTRAINT graph_projection_results_counts_ck CHECK (vertex_count >= 0 AND vertex_count <= 100000 AND edge_count >= 0 AND edge_count <= 250000),
    CONSTRAINT graph_projection_results_id_ck CHECK (projection_result_id ~ '^gpres_[a-f0-9]{64}$'::text),
    CONSTRAINT graph_projection_results_identity_ck CHECK (char_length(graph_view_id) >= 1 AND char_length(graph_view_id) <= 255 AND char_length(source_owner_id) >= 1 AND char_length(source_owner_id) <= 255 AND char_length(source_snapshot_id) >= 1 AND char_length(source_snapshot_id) <= 255 AND projection_schema_id = 'graph_projection.v2'::text AND char_length(projection_version) >= 1 AND char_length(projection_version) <= 255 AND normalized_configuration_sha256 ~ '^[a-f0-9]{64}$'::text AND normalized_source_sha256 ~ '^[a-f0-9]{64}$'::text AND canonical_output_sha256 ~ '^[a-f0-9]{64}$'::text),
    CONSTRAINT graph_projection_results_json_ck CHECK (octet_length(result_json) >= 2),
    CONSTRAINT graph_projection_results_pkey PRIMARY KEY (projection_result_id)
);

REVOKE ALL ON TABLE public.graph_projection_results FROM PUBLIC;

REVOKE ALL ON TYPE public.graph_projection_results FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.graph_projection_results TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.graph_projection_results TO cartulary_recovery;

GRANT USAGE ON TYPE public.graph_projection_results TO cartulary_runtime;

GRANT USAGE ON TYPE public.graph_projection_results TO cartulary_recovery;

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

CREATE INDEX graph_projection_result_edges_dst_traversal_idx ON public.graph_projection_result_edges USING btree (projection_result_id, dst_vertex_id, sort_ordinal, edge_id);

CREATE INDEX graph_projection_result_edges_order_idx ON public.graph_projection_result_edges USING btree (projection_result_id, sort_key, edge_id);

CREATE INDEX graph_projection_result_edges_src_traversal_idx ON public.graph_projection_result_edges USING btree (projection_result_id, src_vertex_id, sort_ordinal, edge_id);

CREATE INDEX graph_projection_result_leases_expiry_idx ON public.graph_projection_result_leases USING btree (leased_until, projection_result_id);

CREATE INDEX graph_projection_result_vertices_order_idx ON public.graph_projection_result_vertices USING btree (projection_result_id, sort_key, vertex_id);

CREATE INDEX graph_projection_results_cleanup_candidate_idx ON public.graph_projection_results USING btree (source_owner_id, published_at, projection_result_id);

ALTER TABLE public.graph_projection_result_edges ADD CONSTRAINT graph_projection_result_edges_dst_vertex_fkey FOREIGN KEY (projection_result_id, dst_vertex_id) REFERENCES public.graph_projection_result_vertices(projection_result_id, vertex_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.graph_projection_result_edges ADD CONSTRAINT graph_projection_result_edges_result_id_fkey FOREIGN KEY (projection_result_id) REFERENCES public.graph_projection_results(projection_result_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.graph_projection_result_edges ADD CONSTRAINT graph_projection_result_edges_src_vertex_fkey FOREIGN KEY (projection_result_id, src_vertex_id) REFERENCES public.graph_projection_result_vertices(projection_result_id, vertex_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.graph_projection_result_leases ADD CONSTRAINT graph_projection_result_leases_result_id_fkey FOREIGN KEY (projection_result_id) REFERENCES public.graph_projection_results(projection_result_id) ON DELETE RESTRICT ON UPDATE NO ACTION;

ALTER TABLE public.graph_projection_result_vertices ADD CONSTRAINT graph_projection_result_vertices_result_id_fkey FOREIGN KEY (projection_result_id) REFERENCES public.graph_projection_results(projection_result_id) ON DELETE CASCADE ON UPDATE NO ACTION;

CREATE TRIGGER graph_projection_result_edges_immutable_update BEFORE UPDATE ON public.graph_projection_result_edges FOR EACH ROW EXECUTE FUNCTION public.graph_projection_reject_immutable_result_update();

CREATE TRIGGER graph_projection_result_vertices_immutable_update BEFORE UPDATE ON public.graph_projection_result_vertices FOR EACH ROW EXECUTE FUNCTION public.graph_projection_reject_immutable_result_update();

CREATE TRIGGER graph_projection_results_immutable_update BEFORE UPDATE ON public.graph_projection_results FOR EACH ROW EXECUTE FUNCTION public.graph_projection_reject_immutable_result_update();

-- +goose Down
-- Disposable harness cleanup only; no production downgrade or restore guarantee.
DROP TRIGGER graph_projection_results_immutable_update ON public.graph_projection_results;
DROP TRIGGER graph_projection_result_vertices_immutable_update ON public.graph_projection_result_vertices;
DROP TRIGGER graph_projection_result_edges_immutable_update ON public.graph_projection_result_edges;
ALTER TABLE public.graph_projection_result_vertices DROP CONSTRAINT graph_projection_result_vertices_result_id_fkey;
ALTER TABLE public.graph_projection_result_leases DROP CONSTRAINT graph_projection_result_leases_result_id_fkey;
ALTER TABLE public.graph_projection_result_edges DROP CONSTRAINT graph_projection_result_edges_src_vertex_fkey;
ALTER TABLE public.graph_projection_result_edges DROP CONSTRAINT graph_projection_result_edges_result_id_fkey;
ALTER TABLE public.graph_projection_result_edges DROP CONSTRAINT graph_projection_result_edges_dst_vertex_fkey;
DROP FUNCTION public.graph_projection_reject_immutable_result_update();
DROP TABLE public.graph_projection_results;
DROP TABLE public.graph_projection_result_vertices;
DROP TABLE public.graph_projection_result_leases;
DROP TABLE public.graph_projection_result_edges;
