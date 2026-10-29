-- +goose Up
-- Diagnostic identity is independent of table allocation.
ALTER TABLE public.network_flow_rejected_row_diagnostics
    DROP CONSTRAINT network_flow_rejected_row_diagnostics_pkey;
ALTER TABLE public.network_flow_rejected_row_diagnostics
    ADD CONSTRAINT network_flow_rejected_row_diagnostics_pkey
    PRIMARY KEY (network_flow_table_id, diagnostic_id);

-- +goose Down
-- A duplicate across tables makes downgrade fail atomically; never delete data.
ALTER TABLE public.network_flow_rejected_row_diagnostics
    DROP CONSTRAINT network_flow_rejected_row_diagnostics_pkey;
ALTER TABLE public.network_flow_rejected_row_diagnostics
    ADD CONSTRAINT network_flow_rejected_row_diagnostics_pkey PRIMARY KEY (diagnostic_id);
