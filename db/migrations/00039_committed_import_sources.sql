-- +goose Up
-- Fresh-state cutover only. Historical source workflows are never backfilled.
LOCK TABLE public.import_sessions IN ACCESS EXCLUSIVE MODE;
LOCK TABLE public.network_flow_tables IN ACCESS EXCLUSIVE MODE;
-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM public.import_sessions)
       OR EXISTS (SELECT 1 FROM public.network_flow_tables)
       OR EXISTS (SELECT 1 FROM public.extension_state_metadata WHERE profile_id = 'network_flow_activity') THEN
        RAISE EXCEPTION 'committed provenance requires fresh state; retain the previous deployment for historical access';
    END IF;
END
$$;
-- +goose StatementEnd

-- Approved source coordinates are independent of discovery's optional tabular
-- projection. Analytical owners validate them without materializing Core rows.
ALTER TABLE public.import_units
    ADD COLUMN approved_header_row_ref integer,
    ADD COLUMN approved_data_start_row_ref integer,
    ADD CONSTRAINT import_units_approved_source_selection_ck CHECK (
        (approved_header_row_ref IS NULL AND approved_data_start_row_ref IS NULL)
        OR (approved_header_row_ref IS NOT NULL AND approved_data_start_row_ref IS NOT NULL
            AND approved_header_row_ref >= 1 AND approved_data_start_row_ref > approved_header_row_ref)
    );

CREATE TABLE public.import_committed_source_blobs (
    content_sha256 text NOT NULL,
    byte_size bigint NOT NULL,
    CONSTRAINT import_committed_source_blobs_pkey PRIMARY KEY (content_sha256),
    CONSTRAINT import_committed_source_blobs_digest_ck CHECK (content_sha256 ~ '^[0-9a-f]{64}$'),
    CONSTRAINT import_committed_source_blobs_nonnegative_size_ck CHECK (byte_size >= 0)
);

CREATE TABLE public.import_committed_source_chunks (
    content_sha256 text NOT NULL,
    chunk_ordinal bigint NOT NULL,
    source_bytes bytea NOT NULL,
    CONSTRAINT import_committed_source_chunks_pkey PRIMARY KEY (content_sha256, chunk_ordinal),
    CONSTRAINT import_committed_source_chunks_ordinal_ck CHECK (chunk_ordinal >= 0),
    CONSTRAINT import_committed_source_chunks_size_ck CHECK (octet_length(source_bytes) BETWEEN 1 AND 65536),
    CONSTRAINT import_committed_source_chunks_blob_fk FOREIGN KEY (content_sha256) REFERENCES public.import_committed_source_blobs(content_sha256) ON UPDATE NO ACTION ON DELETE CASCADE
);
ALTER TABLE public.import_committed_source_chunks ALTER COLUMN source_bytes SET STORAGE EXTERNAL;

CREATE TABLE public.import_committed_sources (
    committed_source_id uuid NOT NULL,
    incident_id uuid NOT NULL,
    content_sha256 text NOT NULL,
    source_file_kind text NOT NULL,
    source_media_type text NOT NULL,
    original_filename text NOT NULL,
    parser_profile_id text NOT NULL,
    parser_version text NOT NULL,
    locator_kind text NOT NULL,
    locator text NOT NULL,
    source_rect_a1 text NOT NULL,
    header_row_ref integer NOT NULL,
    data_start_row_ref integer NOT NULL,
    approved_mapping_json jsonb NOT NULL,
    mapping_fingerprint text NOT NULL,
    target_kind text NOT NULL,
    target_view_schema_id text NOT NULL,
    extension_profile_id text NOT NULL,
    origin_import_session_id uuid NOT NULL,
    origin_import_unit_id uuid NOT NULL,
    created_by_user_id uuid NOT NULL,
    committed_at timestamptz NOT NULL,
    CONSTRAINT import_committed_sources_pkey PRIMARY KEY (committed_source_id),
    CONSTRAINT import_committed_sources_scope_key UNIQUE (incident_id, committed_source_id),
    CONSTRAINT import_committed_sources_origin_key UNIQUE (incident_id, origin_import_unit_id),
    CONSTRAINT import_committed_sources_file_kind_ck CHECK (source_file_kind IN ('csv', 'xlsx')),
    CONSTRAINT import_committed_sources_row_selection_ck CHECK (header_row_ref >= 1 AND data_start_row_ref > header_row_ref),
    CONSTRAINT import_committed_sources_mapping_ck CHECK (jsonb_typeof(approved_mapping_json) = 'object'),
    CONSTRAINT import_committed_sources_incident_fk FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON UPDATE NO ACTION ON DELETE CASCADE,
    CONSTRAINT import_committed_sources_blob_fk FOREIGN KEY (content_sha256) REFERENCES public.import_committed_source_blobs(content_sha256) ON UPDATE NO ACTION ON DELETE NO ACTION,
    CONSTRAINT import_committed_sources_actor_fk FOREIGN KEY (created_by_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION
);
CREATE INDEX import_committed_sources_content_idx ON public.import_committed_sources(content_sha256);
CREATE INDEX import_committed_sources_actor_idx ON public.import_committed_sources(created_by_user_id);
REVOKE ALL ON TABLE public.import_committed_source_blobs, public.import_committed_source_chunks, public.import_committed_sources FROM PUBLIC;
REVOKE ALL ON TYPE public.import_committed_source_blobs, public.import_committed_source_chunks, public.import_committed_sources FROM PUBLIC;
GRANT SELECT, INSERT, DELETE ON TABLE public.import_committed_source_blobs, public.import_committed_source_chunks, public.import_committed_sources TO cartulary_runtime;
GRANT SELECT, INSERT, UPDATE, DELETE, TRUNCATE ON TABLE public.import_committed_source_blobs, public.import_committed_source_chunks, public.import_committed_sources TO cartulary_recovery;

-- +goose StatementBegin
CREATE FUNCTION public.import_reject_committed_source_update() RETURNS trigger
LANGUAGE plpgsql SET search_path = pg_catalog, public AS $$
BEGIN
    RAISE EXCEPTION 'committed import sources are immutable';
END;
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION public.import_reject_committed_source_update() FROM PUBLIC;
CREATE TRIGGER import_committed_sources_immutable_update BEFORE UPDATE ON public.import_committed_sources
FOR EACH ROW EXECUTE FUNCTION public.import_reject_committed_source_update();
CREATE TRIGGER import_committed_source_blobs_immutable_update BEFORE UPDATE ON public.import_committed_source_blobs
FOR EACH ROW EXECUTE FUNCTION public.import_reject_committed_source_update();
CREATE TRIGGER import_committed_source_chunks_immutable_update BEFORE UPDATE ON public.import_committed_source_chunks
FOR EACH ROW EXECUTE FUNCTION public.import_reject_committed_source_update();

-- +goose Down
-- +goose StatementBegin
DO $$ BEGIN
    RAISE EXCEPTION 'incident portability requires fresh state and has no in-place downgrade; retain the prior executable with its own data';
END $$;
-- +goose StatementEnd
