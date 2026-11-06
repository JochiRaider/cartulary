-- +goose Up
-- State 7 is a fresh-state cutover. Never infer provenance or historical
-- selected inputs for previously retained Network Flow state.
LOCK TABLE public.network_flow_tables, public.network_flow_rows,
    public.network_flow_rejected_row_diagnostics, public.network_flow_indicator_bindings,
    public.network_flow_graph_views IN ACCESS EXCLUSIVE MODE;
LOCK TABLE public.extension_state_metadata IN SHARE ROW EXCLUSIVE MODE;
-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM public.network_flow_tables)
       OR EXISTS (SELECT 1 FROM public.network_flow_rows)
       OR EXISTS (SELECT 1 FROM public.network_flow_rejected_row_diagnostics)
       OR EXISTS (SELECT 1 FROM public.network_flow_indicator_bindings)
       OR EXISTS (SELECT 1 FROM public.network_flow_graph_views)
       OR EXISTS (SELECT 1 FROM public.extension_state_metadata WHERE profile_id = 'network_flow_activity') THEN
        RAISE EXCEPTION 'Network Flow state 7 requires fresh initialization; retain historical data with its matching executable';
    END IF;
END
$$;
-- +goose StatementEnd

ALTER TABLE public.network_flow_tables
    DROP CONSTRAINT network_flow_tables_source_import_session_id_fkey,
    DROP CONSTRAINT network_flow_tables_source_import_unit_id_fkey,
    DROP CONSTRAINT network_flow_tables_source_import_unit_id_key;
DROP INDEX public.network_flow_tables_source_import_session_id_fk_idx;
DROP INDEX public.network_flow_tables_source_import_session_id_scope_idx;
ALTER TABLE public.network_flow_tables
    DROP COLUMN source_import_session_id,
    DROP COLUMN source_import_unit_id,
    ADD COLUMN committed_source_id uuid NOT NULL,
    ADD COLUMN source_interpretation_sha256 text NOT NULL,
    ADD COLUMN source_raw_cell_scalar_limit bigint NOT NULL,
    ADD CONSTRAINT network_flow_tables_source_raw_cell_scalar_limit_ck CHECK (source_raw_cell_scalar_limit BETWEEN 1 AND 16384),
    ADD CONSTRAINT network_flow_tables_committed_source_key UNIQUE (committed_source_id),
    ADD CONSTRAINT network_flow_tables_committed_source_fk FOREIGN KEY (incident_id, committed_source_id)
        REFERENCES public.import_committed_sources (incident_id, committed_source_id)
        ON UPDATE NO ACTION ON DELETE NO ACTION DEFERRABLE INITIALLY DEFERRED,
    ADD CONSTRAINT network_flow_tables_interpretation_ck CHECK (source_interpretation_sha256 ~ '^[0-9a-f]{64}$');
CREATE INDEX network_flow_tables_committed_source_scope_idx
    ON public.network_flow_tables (incident_id, committed_source_id);

ALTER TABLE public.network_flow_graph_views
    ADD COLUMN refresh_intent_json jsonb NOT NULL,
    ADD COLUMN selected_source_binding_json jsonb,
    ADD CONSTRAINT network_flow_graph_views_refresh_intent_ck CHECK (
        jsonb_typeof(refresh_intent_json) = 'object'),
    ADD CONSTRAINT network_flow_graph_views_selected_source_ck CHECK (
        (selected_projection_result_id IS NULL AND selected_source_binding_json IS NULL)
        OR (selected_projection_result_id IS NOT NULL AND selected_source_binding_json IS NOT NULL
            AND jsonb_typeof(selected_source_binding_json) = 'object'));

-- +goose Down
-- +goose StatementBegin
DO $$
BEGIN
    RAISE EXCEPTION 'Network Flow state 7 cannot be downgraded; use the previous executable with its own retained data';
END
$$;
-- +goose StatementEnd
