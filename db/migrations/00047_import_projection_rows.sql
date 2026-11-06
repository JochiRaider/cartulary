-- +goose Up
-- Fresh lineage v4 only. No backfill or historical projection reader.
ALTER TABLE public.import_units ADD COLUMN source_rows_sha256 text;
ALTER TABLE public.import_units ADD CONSTRAINT import_units_source_rows_digest_ck
    CHECK (source_rows_sha256 IS NULL OR source_rows_sha256 ~ '^[0-9a-f]{64}$');
ALTER TABLE public.import_units DROP CONSTRAINT import_units_projection_shape_ck;
ALTER TABLE public.import_units ADD CONSTRAINT import_units_projection_shape_ck CHECK (
    (locator_kind='csv_file' AND source_rows_json IS NULL AND (
        num_nonnulls(source_rect_a1,header_row_ref,data_start_row_ref,inferred_row_count,
            inferred_column_count,columns_json,source_rows_sha256,preview_rows_json)=0
        OR num_nonnulls(source_rect_a1,header_row_ref,data_start_row_ref,inferred_row_count,
            inferred_column_count,columns_json,source_rows_sha256,preview_rows_json)=8))
    OR (locator_kind<>'csv_file' AND source_rows_sha256 IS NULL AND
        num_nonnulls(source_rect_a1,header_row_ref,data_start_row_ref,inferred_row_count,
            inferred_column_count,columns_json,source_rows_json,preview_rows_json)=8)
);
CREATE TABLE public.import_unit_projection_rows (
    import_unit_id uuid NOT NULL,
    source_row_ref integer NOT NULL,
    row_json bytea NOT NULL,
    CONSTRAINT import_unit_projection_rows_pkey PRIMARY KEY (import_unit_id, source_row_ref),
    CONSTRAINT import_unit_projection_rows_ref_ck CHECK (source_row_ref >= 2),
    CONSTRAINT import_unit_projection_rows_bytes_ck CHECK (octet_length(row_json) > 0),
    CONSTRAINT import_unit_projection_rows_unit_fk FOREIGN KEY (import_unit_id)
        REFERENCES public.import_units(import_unit_id) ON UPDATE NO ACTION ON DELETE CASCADE
);
ALTER TABLE public.import_unit_projection_rows ALTER COLUMN row_json SET STORAGE EXTERNAL;
REVOKE ALL ON TABLE public.import_unit_projection_rows FROM PUBLIC;
REVOKE ALL ON TYPE public.import_unit_projection_rows FROM PUBLIC;
GRANT SELECT, INSERT, UPDATE, DELETE ON TABLE public.import_unit_projection_rows TO cartulary_runtime;
GRANT SELECT, INSERT, UPDATE, DELETE, TRUNCATE ON TABLE public.import_unit_projection_rows TO cartulary_recovery;

-- Successful owner results outlive their workflow journals. These immutable
-- associations bind a destination record to one source application and row.
CREATE TABLE public.import_committed_source_records (
    incident_id uuid NOT NULL,
    committed_source_id uuid NOT NULL,
    source_row_ref integer NOT NULL,
    record_id uuid NOT NULL,
    CONSTRAINT import_committed_source_records_pkey PRIMARY KEY (committed_source_id, source_row_ref),
    CONSTRAINT import_committed_source_records_row_ck CHECK (source_row_ref >= 2),
    CONSTRAINT import_committed_source_records_source_fk FOREIGN KEY (incident_id, committed_source_id)
        REFERENCES public.import_committed_sources(incident_id, committed_source_id) ON UPDATE NO ACTION ON DELETE CASCADE,
    CONSTRAINT import_committed_source_records_record_fk FOREIGN KEY (incident_id, record_id)
        REFERENCES public.records(incident_id, record_id) ON UPDATE NO ACTION ON DELETE CASCADE
);
CREATE INDEX import_committed_source_records_record_idx ON public.import_committed_source_records (incident_id, record_id, committed_source_id);
REVOKE ALL ON TABLE public.import_committed_source_records FROM PUBLIC;
REVOKE ALL ON TYPE public.import_committed_source_records FROM PUBLIC;
GRANT SELECT, INSERT ON TABLE public.import_committed_source_records TO cartulary_runtime;
GRANT SELECT, INSERT, UPDATE, DELETE, TRUNCATE ON TABLE public.import_committed_source_records TO cartulary_recovery;

-- +goose Down
-- Downgrade would reinterpret current provenance and is unsupported.
-- +goose StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'import projection downgrade is unsupported'; END $$;
-- +goose StatementEnd
