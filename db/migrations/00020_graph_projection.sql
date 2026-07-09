-- +goose Up
--
-- Name: graph_projection_edges; Type: TABLE; Schema: public; Owner: -
--

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

--
-- Name: graph_projection_idempotency; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.graph_projection_idempotency (
    operation text NOT NULL,
    idempotency_key text NOT NULL,
    request_fingerprint text NOT NULL,
    graph_view_id text,
    projection_run_id text,
    response_json jsonb NOT NULL,
    created_at timestamp with time zone NOT NULL,
    expires_at timestamp with time zone NOT NULL
);

--
-- Name: graph_projection_runs; Type: TABLE; Schema: public; Owner: -
--

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
    completed_at timestamp with time zone,
    validation_summary_json jsonb NOT NULL,
    failure_reason text,
    graph_view_json jsonb,
    invalidation_json jsonb,
    retention_expires_at timestamp with time zone,
    projection_output_digest text,
    CONSTRAINT graph_projection_runs_projection_output_digest_ck CHECK ((((state = ANY (ARRAY['available'::text, 'replaced'::text])) AND (projection_output_digest ~ '^[a-f0-9]{64}$'::text)) OR ((state <> ALL (ARRAY['available'::text, 'replaced'::text])) AND ((projection_output_digest IS NULL) OR (projection_output_digest ~ '^[a-f0-9]{64}$'::text)))))
);

--
-- Name: graph_projection_vertices; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.graph_projection_vertices (
    projection_run_id text NOT NULL,
    graph_view_id text NOT NULL,
    vertex_id text NOT NULL,
    vertex_kind text NOT NULL,
    sort_key text NOT NULL,
    vertex_json jsonb NOT NULL
);

--
-- Name: graph_projection_views; Type: TABLE; Schema: public; Owner: -
--

CREATE TABLE public.graph_projection_views (
    graph_view_id text NOT NULL,
    graph_view_key text NOT NULL,
    state text NOT NULL,
    latest_projection_run_id text,
    latest_source_snapshot_id text,
    projection_version text,
    updated_at timestamp with time zone NOT NULL,
    validation_status text NOT NULL
);

--
-- Name: graph_projection_edges graph_projection_edges_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.graph_projection_edges
    ADD CONSTRAINT graph_projection_edges_pkey PRIMARY KEY (projection_run_id, edge_id);

--
-- Name: graph_projection_idempotency graph_projection_idempotency_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.graph_projection_idempotency
    ADD CONSTRAINT graph_projection_idempotency_pkey PRIMARY KEY (operation, idempotency_key);

--
-- Name: graph_projection_runs graph_projection_runs_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.graph_projection_runs
    ADD CONSTRAINT graph_projection_runs_pkey PRIMARY KEY (projection_run_id);

--
-- Name: graph_projection_vertices graph_projection_vertices_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.graph_projection_vertices
    ADD CONSTRAINT graph_projection_vertices_pkey PRIMARY KEY (projection_run_id, vertex_id);

--
-- Name: graph_projection_views graph_projection_views_graph_view_key_key; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.graph_projection_views
    ADD CONSTRAINT graph_projection_views_graph_view_key_key UNIQUE (graph_view_key);

--
-- Name: graph_projection_views graph_projection_views_pkey; Type: CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.graph_projection_views
    ADD CONSTRAINT graph_projection_views_pkey PRIMARY KEY (graph_view_id);

--
-- Name: graph_projection_edges_endpoint_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX graph_projection_edges_endpoint_idx ON public.graph_projection_edges USING btree (projection_run_id, src_vertex_id, dst_vertex_id, direction);

--
-- Name: graph_projection_edges_view_sort_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX graph_projection_edges_view_sort_idx ON public.graph_projection_edges USING btree (graph_view_id, sort_key, edge_id);

--
-- Name: graph_projection_idempotency_expiry_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX graph_projection_idempotency_expiry_idx ON public.graph_projection_idempotency USING btree (expires_at);

--
-- Name: graph_projection_runs_view_state_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX graph_projection_runs_view_state_idx ON public.graph_projection_runs USING btree (graph_view_id, state, completed_at DESC, projection_run_id);

--
-- Name: graph_projection_vertices_view_sort_idx; Type: INDEX; Schema: public; Owner: -
--

CREATE INDEX graph_projection_vertices_view_sort_idx ON public.graph_projection_vertices USING btree (graph_view_id, sort_key, vertex_id);

--
-- Name: graph_projection_edges graph_projection_edges_projection_run_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.graph_projection_edges
    ADD CONSTRAINT graph_projection_edges_projection_run_id_fkey FOREIGN KEY (projection_run_id) REFERENCES public.graph_projection_runs(projection_run_id) ON DELETE CASCADE;

--
-- Name: graph_projection_runs graph_projection_runs_graph_view_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.graph_projection_runs
    ADD CONSTRAINT graph_projection_runs_graph_view_id_fkey FOREIGN KEY (graph_view_id) REFERENCES public.graph_projection_views(graph_view_id) ON DELETE CASCADE;

--
-- Name: graph_projection_vertices graph_projection_vertices_projection_run_id_fkey; Type: FK CONSTRAINT; Schema: public; Owner: -
--

ALTER TABLE ONLY public.graph_projection_vertices
    ADD CONSTRAINT graph_projection_vertices_projection_run_id_fkey FOREIGN KEY (projection_run_id) REFERENCES public.graph_projection_runs(projection_run_id) ON DELETE CASCADE;

-- +goose Down
DROP TABLE IF EXISTS public.graph_projection_views CASCADE;
DROP TABLE IF EXISTS public.graph_projection_runs CASCADE;
DROP TABLE IF EXISTS public.graph_projection_vertices CASCADE;
DROP TABLE IF EXISTS public.graph_projection_edges CASCADE;
DROP TABLE IF EXISTS public.graph_projection_idempotency CASCADE;
