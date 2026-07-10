-- +goose Up
--
-- Name: Network Flow Activity indicator bindings; Type: TABLES; Schema: public; Owner: -
--

CREATE TABLE public.network_flow_indicator_bindings (
    network_flow_indicator_binding_id text NOT NULL,
    incident_id uuid NOT NULL,
    target_indicator_record_id uuid NOT NULL,
    target_indicator_type text NOT NULL,
    target_indicator_value_kind text NOT NULL,
    target_indicator_normalized_value text NOT NULL,
    selector_kind text NOT NULL,
    candidate_value text NOT NULL,
    source_row_refs jsonb NOT NULL,
    source_row_ref_row_ids text[] NOT NULL,
    source_row_refs_truncated boolean DEFAULT false NOT NULL,
    source_row_refs_total_count bigint NOT NULL,
    created_by_user_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT network_flow_indicator_bindings_pkey PRIMARY KEY (network_flow_indicator_binding_id),
    CONSTRAINT network_flow_indicator_bindings_id_ck CHECK (network_flow_indicator_binding_id ~ '^nfb_[a-f0-9]{32}$'::text),
    CONSTRAINT network_flow_indicator_bindings_selector_kind_ck CHECK (selector_kind = ANY (ARRAY['row_field_value'::text, 'row_refs'::text, 'graph_vertex'::text, 'graph_edge'::text])),
    CONSTRAINT network_flow_indicator_bindings_target_ck CHECK ((target_indicator_type = ANY (ARRAY['ipv4_addr'::text, 'ipv6_addr'::text])) AND (target_indicator_value_kind = 'atomic'::text)),
    CONSTRAINT network_flow_indicator_bindings_candidate_ck CHECK ((char_length(candidate_value) >= 1) AND (char_length(candidate_value) <= 45) AND (candidate_value = target_indicator_normalized_value)),
    CONSTRAINT network_flow_indicator_bindings_source_refs_ck CHECK ((jsonb_typeof(source_row_refs) = 'array'::text) AND (jsonb_array_length(source_row_refs) >= 1) AND (cardinality(source_row_ref_row_ids) >= 1)),
    CONSTRAINT network_flow_indicator_bindings_source_count_ck CHECK ((source_row_refs_total_count >= cardinality(source_row_ref_row_ids)) AND (source_row_refs_total_count > 0))
);

ALTER TABLE ONLY public.network_flow_indicator_bindings
    ADD CONSTRAINT network_flow_indicator_bindings_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE,
    ADD CONSTRAINT network_flow_indicator_bindings_target_indicator_record_id_fkey FOREIGN KEY (target_indicator_record_id) REFERENCES public.indicators(record_id) ON DELETE CASCADE,
    ADD CONSTRAINT network_flow_indicator_bindings_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(id);

CREATE UNIQUE INDEX network_flow_indicator_bindings_identity_uidx
    ON public.network_flow_indicator_bindings USING btree (incident_id, target_indicator_record_id, candidate_value, source_row_ref_row_ids);

CREATE INDEX network_flow_indicator_bindings_incident_created_idx
    ON public.network_flow_indicator_bindings USING btree (incident_id, created_at DESC, network_flow_indicator_binding_id);

-- +goose Down
DROP INDEX IF EXISTS public.network_flow_indicator_bindings_incident_created_idx;
DROP INDEX IF EXISTS public.network_flow_indicator_bindings_identity_uidx;
DROP TABLE IF EXISTS public.network_flow_indicator_bindings;
