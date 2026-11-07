-- +goose Up
-- Cover the complete incident-scoped source reference without changing row identity.
CREATE INDEX import_committed_source_records_source_fk_idx
ON public.import_committed_source_records USING btree (incident_id, committed_source_id);

-- +goose Down
DROP INDEX public.import_committed_source_records_source_fk_idx;
