-- +goose Up
-- PostgreSQL 18.6 consolidated baseline: timeline.
-- Administrator provisioning and final admission are separate lifecycle steps.

CREATE TABLE public.timeline_events (
    record_id uuid DEFAULT gen_random_uuid() NOT NULL,
    incident_id uuid NOT NULL,
    capture_state text NOT NULL,
    row_version bigint DEFAULT 1 NOT NULL,
    recorded_at timestamp with time zone DEFAULT now() NOT NULL,
    edited_at timestamp with time zone DEFAULT now() NOT NULL,
    created_by_user_id uuid NOT NULL,
    updated_by_user_id uuid NOT NULL,
    reviewed_by_user_id uuid,
    reviewed_at timestamp with time zone,
    superseded_by_user_id uuid,
    superseded_at timestamp with time zone,
    date_entered_text text,
    analyst_text text,
    mitre_stage_text text,
    device_object_text text,
    ip_address_text text,
    activity_utc_text text,
    activity_local_text text,
    raw_activity_text text,
    activity_synopsis_text text,
    data_source_text text,
    activity_utc_generated boolean DEFAULT false NOT NULL,
    activity_local_generated boolean DEFAULT false NOT NULL,
    activity_time_pair_state text DEFAULT 'disabled'::text NOT NULL,
    CONSTRAINT timeline_events_activity_time_pair_state_check CHECK (activity_time_pair_state = ANY (ARRAY['disabled'::text, 'empty'::text, 'paired_generated'::text, 'paired_user_preserved'::text, 'paired_mismatch'::text, 'conversion_unavailable'::text])),
    CONSTRAINT timeline_events_capture_state_check CHECK (capture_state = ANY (ARRAY['rough'::text, 'enriched'::text, 'reviewed'::text, 'superseded'::text])),
    CONSTRAINT timeline_events_pkey PRIMARY KEY (record_id),
    CONSTRAINT timeline_events_incident_identity_key UNIQUE (incident_id, record_id)
);

REVOKE ALL ON TABLE public.timeline_events FROM PUBLIC;

REVOKE ALL ON TYPE public.timeline_events FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.timeline_events TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.timeline_events TO cartulary_recovery;

CREATE TABLE public.timeline_source_provenance (
    record_id uuid NOT NULL,
    source_identity_hash bytea NOT NULL,
    source_row_ordinal integer NOT NULL,
    source_column_ordinal integer NOT NULL,
    source_kind text NOT NULL,
    source_metadata jsonb NOT NULL,
    source_header_json jsonb NOT NULL,
    raw_value text NOT NULL,
    cell_kind text,
    created_at timestamp with time zone NOT NULL,
    CONSTRAINT timeline_source_provenance_cell_kind_ck CHECK (cell_kind IS NULL OR octet_length(cell_kind) >= 1 AND octet_length(cell_kind) <= 64 AND cell_kind !~ '[[:cntrl:]]'::text),
    CONSTRAINT timeline_source_provenance_header_ck CHECK (octet_length(source_header_json::text) <= 65536),
    CONSTRAINT timeline_source_provenance_identity_hash_ck CHECK (octet_length(source_identity_hash) = 32),
    CONSTRAINT timeline_source_provenance_metadata_ck CHECK (jsonb_typeof(source_metadata) = 'object'::text AND octet_length(source_metadata::text) <= 65536),
    CONSTRAINT timeline_source_provenance_ordinals_ck CHECK (source_row_ordinal >= 0 AND source_column_ordinal >= 0),
    CONSTRAINT timeline_source_provenance_pkey PRIMARY KEY (record_id, source_identity_hash, source_row_ordinal, source_column_ordinal),
    CONSTRAINT timeline_source_provenance_raw_value_ck CHECK (octet_length(raw_value) <= 65536),
    CONSTRAINT timeline_source_provenance_source_kind_ck CHECK (octet_length(source_kind) >= 1 AND octet_length(source_kind) <= 64 AND source_kind !~ '[[:cntrl:]]'::text)
);

REVOKE ALL ON TABLE public.timeline_source_provenance FROM PUBLIC;

REVOKE ALL ON TYPE public.timeline_source_provenance FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.timeline_source_provenance TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.timeline_source_provenance TO cartulary_recovery;

CREATE TABLE public.timeline_time_conversion_profiles (
    incident_id uuid NOT NULL,
    enabled boolean DEFAULT false NOT NULL,
    local_offset_minutes integer,
    local_label text,
    profile_version bigint DEFAULT 1 NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_by_user_id uuid,
    CONSTRAINT timeline_time_conversion_profiles_enabled_offset_ck CHECK (enabled = false OR local_offset_minutes IS NOT NULL),
    CONSTRAINT timeline_time_conversion_profiles_local_offset_minutes_check CHECK (local_offset_minutes IS NULL OR local_offset_minutes >= '-840'::integer AND local_offset_minutes <= 840),
    CONSTRAINT timeline_time_conversion_profiles_pkey PRIMARY KEY (incident_id)
);

REVOKE ALL ON TABLE public.timeline_time_conversion_profiles FROM PUBLIC;

REVOKE ALL ON TYPE public.timeline_time_conversion_profiles FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.timeline_time_conversion_profiles TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.timeline_time_conversion_profiles TO cartulary_recovery;

CREATE INDEX idx_timeline_source_provenance_record ON public.timeline_source_provenance USING btree (record_id, source_row_ordinal, source_column_ordinal, source_identity_hash);

CREATE INDEX timeline_events_created_by_user_id_fk_idx ON public.timeline_events USING btree (created_by_user_id);

CREATE INDEX timeline_events_incident_id_record_id_fk_idx ON public.timeline_events USING btree (incident_id, record_id);

CREATE INDEX timeline_events_incident_lookup_idx ON public.timeline_events USING btree (incident_id, edited_at DESC, record_id);

CREATE INDEX timeline_events_reviewed_by_user_id_fk_idx ON public.timeline_events USING btree (reviewed_by_user_id);

CREATE INDEX timeline_events_superseded_by_user_id_fk_idx ON public.timeline_events USING btree (superseded_by_user_id);

CREATE INDEX timeline_events_updated_by_user_id_fk_idx ON public.timeline_events USING btree (updated_by_user_id);

CREATE INDEX timeline_time_conversion_profiles_updated_by_user_id_fk_idx ON public.timeline_time_conversion_profiles USING btree (updated_by_user_id);

ALTER TABLE public.timeline_events ADD CONSTRAINT timeline_events_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.timeline_events ADD CONSTRAINT timeline_events_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.timeline_events ADD CONSTRAINT timeline_events_record_envelope_fkey FOREIGN KEY (incident_id, record_id) REFERENCES public.records(incident_id, record_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.timeline_events ADD CONSTRAINT timeline_events_reviewed_by_user_id_fkey FOREIGN KEY (reviewed_by_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.timeline_events ADD CONSTRAINT timeline_events_superseded_by_user_id_fkey FOREIGN KEY (superseded_by_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.timeline_events ADD CONSTRAINT timeline_events_updated_by_user_id_fkey FOREIGN KEY (updated_by_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.timeline_source_provenance ADD CONSTRAINT timeline_source_provenance_record_id_fkey FOREIGN KEY (record_id) REFERENCES public.timeline_events(record_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.timeline_time_conversion_profiles ADD CONSTRAINT timeline_time_conversion_profiles_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.timeline_time_conversion_profiles ADD CONSTRAINT timeline_time_conversion_profiles_updated_by_user_id_fkey FOREIGN KEY (updated_by_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

-- +goose Down
-- Disposable harness cleanup only; no production downgrade or restore guarantee.
ALTER TABLE public.timeline_time_conversion_profiles DROP CONSTRAINT timeline_time_conversion_profiles_updated_by_user_id_fkey;
ALTER TABLE public.timeline_time_conversion_profiles DROP CONSTRAINT timeline_time_conversion_profiles_incident_id_fkey;
ALTER TABLE public.timeline_source_provenance DROP CONSTRAINT timeline_source_provenance_record_id_fkey;
ALTER TABLE public.timeline_events DROP CONSTRAINT timeline_events_updated_by_user_id_fkey;
ALTER TABLE public.timeline_events DROP CONSTRAINT timeline_events_superseded_by_user_id_fkey;
ALTER TABLE public.timeline_events DROP CONSTRAINT timeline_events_reviewed_by_user_id_fkey;
ALTER TABLE public.timeline_events DROP CONSTRAINT timeline_events_record_envelope_fkey;
ALTER TABLE public.timeline_events DROP CONSTRAINT timeline_events_incident_id_fkey;
ALTER TABLE public.timeline_events DROP CONSTRAINT timeline_events_created_by_user_id_fkey;
DROP TABLE public.timeline_time_conversion_profiles;
DROP TABLE public.timeline_source_provenance;
DROP TABLE public.timeline_events;
