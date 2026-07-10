-- +goose Up
--
-- Name: import source stream capabilities and target metadata; Type: TABLE EXTENSION; Schema: public; Owner: -
--

ALTER TABLE public.import_units
    ADD COLUMN source_stream_ref text,
    ADD COLUMN approved_target_kind text,
    ADD COLUMN approved_extension_profile_id text,
    ADD COLUMN approved_target_view_schema_id text;

UPDATE public.import_units
   SET approved_target_kind = 'view_schema',
       approved_target_view_schema_id = approved_mapping_json->>'target_view_schema_id'
 WHERE approved_mapping_json IS NOT NULL
   AND approved_target_kind IS NULL;

ALTER TABLE public.import_units
    ADD CONSTRAINT import_units_source_stream_ref_ck CHECK ((source_stream_ref IS NULL) OR (source_stream_ref ~ '^impsrc_[0-9a-f]{32}$'::text)),
    ADD CONSTRAINT import_units_approved_target_kind_ck CHECK ((approved_target_kind IS NULL) OR (approved_target_kind = ANY (ARRAY['view_schema'::text, 'network_flow_table'::text]))),
    ADD CONSTRAINT import_units_approved_target_shape_ck CHECK (
        (approved_target_kind IS NULL AND approved_extension_profile_id IS NULL AND approved_target_view_schema_id IS NULL)
        OR (approved_target_kind = 'view_schema' AND approved_target_view_schema_id IS NOT NULL AND approved_extension_profile_id IS NULL)
        OR (approved_target_kind = 'network_flow_table' AND approved_target_view_schema_id IS NULL AND approved_extension_profile_id = 'network_flow_activity')
    );

CREATE TABLE public.import_source_streams (
    source_stream_ref text NOT NULL,
    import_session_id uuid NOT NULL,
    import_unit_id uuid NOT NULL,
    source_content_sha256 text NOT NULL,
    source_media_type text NOT NULL,
    source_byte_size bigint NOT NULL,
    source_bytes bytea NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT import_source_streams_pkey PRIMARY KEY (source_stream_ref),
    CONSTRAINT import_source_streams_ref_ck CHECK (source_stream_ref ~ '^impsrc_[0-9a-f]{32}$'::text),
    CONSTRAINT import_source_streams_sha_ck CHECK (source_content_sha256 ~ '^[0-9a-f]{64}$'::text),
    CONSTRAINT import_source_streams_source_byte_size_ck CHECK ((source_byte_size >= 0) AND (octet_length(source_bytes) = source_byte_size))
);

ALTER TABLE ONLY public.import_source_streams
    ADD CONSTRAINT import_source_streams_import_session_id_fkey FOREIGN KEY (import_session_id) REFERENCES public.import_sessions(import_session_id) ON DELETE CASCADE,
    ADD CONSTRAINT import_source_streams_import_unit_id_fkey FOREIGN KEY (import_unit_id) REFERENCES public.import_units(import_unit_id) ON DELETE CASCADE,
    ADD CONSTRAINT import_source_streams_import_unit_id_key UNIQUE (import_unit_id);

CREATE UNIQUE INDEX import_units_source_stream_ref_idx ON public.import_units USING btree (source_stream_ref) WHERE (source_stream_ref IS NOT NULL);
CREATE INDEX import_source_streams_session_idx ON public.import_source_streams USING btree (import_session_id, import_unit_id);

-- +goose Down
DROP INDEX IF EXISTS public.import_source_streams_session_idx;
DROP INDEX IF EXISTS public.import_units_source_stream_ref_idx;

DROP TABLE IF EXISTS public.import_source_streams;

ALTER TABLE public.import_units
    DROP CONSTRAINT IF EXISTS import_units_approved_target_shape_ck,
    DROP CONSTRAINT IF EXISTS import_units_approved_target_kind_ck,
    DROP CONSTRAINT IF EXISTS import_units_source_stream_ref_ck,
    DROP COLUMN IF EXISTS approved_target_view_schema_id,
    DROP COLUMN IF EXISTS approved_extension_profile_id,
    DROP COLUMN IF EXISTS approved_target_kind,
    DROP COLUMN IF EXISTS source_stream_ref;
