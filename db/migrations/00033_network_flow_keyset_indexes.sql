-- +goose NO TRANSACTION
-- +goose Up
--
-- Name: Network Flow keyset pagination indexes; Type: INDEXES; Schema: public; Owner: -
--

CREATE INDEX CONCURRENTLY network_flow_rows_table_default_keyset_idx
    ON public.network_flow_rows USING btree
    (network_flow_table_id, flow_start_utc, flow_end_utc, source_row_number, network_flow_row_id);

CREATE INDEX CONCURRENTLY network_flow_rows_incident_default_keyset_idx
    ON public.network_flow_rows USING btree
    (incident_id, flow_start_utc, flow_end_utc, source_row_number, network_flow_row_id, network_flow_table_id);

CREATE INDEX CONCURRENTLY network_flow_rejected_diagnostics_keyset_idx
    ON public.network_flow_rejected_row_diagnostics USING btree
    (network_flow_table_id, source_row_number, source_column_ordinal ASC NULLS LAST,
     field_key ASC NULLS LAST, error_code, reason_code, diagnostic_id);

DROP INDEX CONCURRENTLY IF EXISTS public.network_flow_rows_table_order_idx;
DROP INDEX CONCURRENTLY IF EXISTS public.network_flow_rejected_row_diagnostics_table_order_idx;

-- +goose Down
CREATE INDEX CONCURRENTLY network_flow_rows_table_order_idx
    ON public.network_flow_rows USING btree
    (network_flow_table_id, source_row_number, network_flow_row_id);

CREATE INDEX CONCURRENTLY network_flow_rejected_row_diagnostics_table_order_idx
    ON public.network_flow_rejected_row_diagnostics USING btree
    (network_flow_table_id, source_row_number, source_column_ordinal, field_key, error_code, diagnostic_id);

DROP INDEX CONCURRENTLY IF EXISTS public.network_flow_rejected_diagnostics_keyset_idx;
DROP INDEX CONCURRENTLY IF EXISTS public.network_flow_rows_incident_default_keyset_idx;
DROP INDEX CONCURRENTLY IF EXISTS public.network_flow_rows_table_default_keyset_idx;
