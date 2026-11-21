-- +goose Up
-- PostgreSQL 18.6 consolidated baseline: incident bundles.
-- Administrator provisioning and final admission are separate lifecycle steps.

CREATE TABLE public.incident_bundle_exports (
    bundle_id uuid DEFAULT gen_random_uuid() NOT NULL,
    incident_id uuid NOT NULL,
    export_job_id uuid NOT NULL,
    exported_by_user_id uuid NOT NULL,
    exported_at timestamp with time zone NOT NULL,
    manifest_sha256 text NOT NULL,
    reference_pack_mode text NOT NULL,
    history_mode text DEFAULT 'full'::text NOT NULL,
    blob_mode text DEFAULT 'full'::text NOT NULL,
    optional_sections text[] DEFAULT '{}'::text[] NOT NULL,
    required_capabilities text[] DEFAULT '{}'::text[] NOT NULL,
    bundle_sha256 text NOT NULL,
    bundle_byte_size bigint NOT NULL,
    bundle_storage_ref text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT incident_bundle_exports_blob_mode_check CHECK (blob_mode = 'full'::text),
    CONSTRAINT incident_bundle_exports_bundle_byte_size_check CHECK (bundle_byte_size >= 0),
    CONSTRAINT incident_bundle_exports_bundle_sha256_check CHECK (bundle_sha256 ~ '^[0-9a-f]{64}$'::text),
    CONSTRAINT incident_bundle_exports_export_job_id_key UNIQUE (export_job_id),
    CONSTRAINT incident_bundle_exports_history_mode_check CHECK (history_mode = 'full'::text),
    CONSTRAINT incident_bundle_exports_manifest_sha256_check CHECK (manifest_sha256 ~ '^[0-9a-f]{64}$'::text),
    CONSTRAINT incident_bundle_exports_pkey PRIMARY KEY (bundle_id),
    CONSTRAINT incident_bundle_exports_reference_pack_mode_check CHECK (reference_pack_mode = ANY (ARRAY['refs_only'::text, 'embedded'::text])),
    CONSTRAINT incident_bundle_exports_storage_ref_ck CHECK (bundle_storage_ref <> ''::text AND bundle_storage_ref !~ '^/'::text AND strpos(bundle_storage_ref, '\'::text) = 0 AND bundle_storage_ref !~ '(^|/)\.\.?(/|$)'::text AND bundle_storage_ref !~ '//'::text AND bundle_storage_ref !~ '/$'::text)
);

REVOKE ALL ON TABLE public.incident_bundle_exports FROM PUBLIC;

REVOKE ALL ON TYPE public.incident_bundle_exports FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.incident_bundle_exports TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.incident_bundle_exports TO cartulary_recovery;

CREATE TABLE public.incident_bundle_imported_actors (
    imported_actor_descriptor_id uuid DEFAULT gen_random_uuid() NOT NULL,
    incident_id uuid NOT NULL,
    source_actor_id text NOT NULL,
    display_name text,
    email_hint text,
    local_user_id uuid,
    imported_at timestamp with time zone DEFAULT now() NOT NULL,
    provider_subject_hint text,
    CONSTRAINT incident_bundle_imported_actors_incident_id_source_actor_id_key UNIQUE (incident_id, source_actor_id),
    CONSTRAINT incident_bundle_imported_actors_pkey PRIMARY KEY (imported_actor_descriptor_id)
);

REVOKE ALL ON TABLE public.incident_bundle_imported_actors FROM PUBLIC;

REVOKE ALL ON TYPE public.incident_bundle_imported_actors FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.incident_bundle_imported_actors TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.incident_bundle_imported_actors TO cartulary_recovery;

CREATE TABLE public.incident_bundle_imported_attributions (
    imported_attribution_id uuid DEFAULT gen_random_uuid() NOT NULL,
    incident_id uuid NOT NULL,
    source_table text NOT NULL,
    source_row_id text NOT NULL,
    source_column text NOT NULL,
    source_actor_id text NOT NULL,
    local_user_id uuid NOT NULL,
    imported_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT incident_bundle_imported_attr_incident_id_source_table_sour_key UNIQUE (incident_id, source_table, source_row_id, source_column),
    CONSTRAINT incident_bundle_imported_attributions_pkey PRIMARY KEY (imported_attribution_id)
);

REVOKE ALL ON TABLE public.incident_bundle_imported_attributions FROM PUBLIC;

REVOKE ALL ON TYPE public.incident_bundle_imported_attributions FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.incident_bundle_imported_attributions TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.incident_bundle_imported_attributions TO cartulary_recovery;

CREATE TABLE public.incident_bundle_job_payloads (
    job_id uuid NOT NULL,
    job_kind text NOT NULL,
    actor_user_id uuid NOT NULL,
    incident_id uuid,
    bundle_id uuid,
    uploaded_sha256 text,
    bundle_staging_ref text,
    imported_incident_id uuid,
    manifest_sha256 text,
    failure_reason text,
    request_json jsonb NOT NULL,
    created_at timestamp with time zone NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT incident_bundle_job_payloads_job_kind_check CHECK (job_kind = ANY (ARRAY['export'::text, 'import'::text])),
    CONSTRAINT incident_bundle_job_payloads_manifest_sha256_check CHECK (manifest_sha256 IS NULL OR manifest_sha256 ~ '^[0-9a-f]{64}$'::text),
    CONSTRAINT incident_bundle_job_payloads_pkey PRIMARY KEY (job_id),
    CONSTRAINT incident_bundle_job_payloads_staging_ref_ck CHECK (bundle_staging_ref IS NULL OR bundle_staging_ref <> ''::text AND bundle_staging_ref !~ '^/'::text AND strpos(bundle_staging_ref, '\'::text) = 0 AND bundle_staging_ref !~ '(^|/)\.\.?(/|$)'::text AND bundle_staging_ref !~ '//'::text AND bundle_staging_ref !~ '/$'::text),
    CONSTRAINT incident_bundle_job_payloads_uploaded_sha256_check CHECK (uploaded_sha256 IS NULL OR uploaded_sha256 ~ '^[0-9a-f]{64}$'::text),
    CONSTRAINT incident_bundle_payload_kind_ck CHECK (job_kind = 'export'::text AND incident_id IS NOT NULL AND uploaded_sha256 IS NULL OR job_kind = 'import'::text AND incident_id IS NULL AND uploaded_sha256 IS NOT NULL)
);

REVOKE ALL ON TABLE public.incident_bundle_job_payloads FROM PUBLIC;

REVOKE ALL ON TYPE public.incident_bundle_job_payloads FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.incident_bundle_job_payloads TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.incident_bundle_job_payloads TO cartulary_recovery;

CREATE TABLE public.incident_bundle_manifest_files (
    bundle_id uuid NOT NULL,
    path text NOT NULL,
    sha256 text NOT NULL,
    size_bytes bigint NOT NULL,
    required boolean DEFAULT true NOT NULL,
    CONSTRAINT incident_bundle_manifest_files_pkey PRIMARY KEY (bundle_id, path),
    CONSTRAINT incident_bundle_manifest_files_sha256_check CHECK (sha256 ~ '^[0-9a-f]{64}$'::text),
    CONSTRAINT incident_bundle_manifest_files_size_bytes_check CHECK (size_bytes >= 0)
);

REVOKE ALL ON TABLE public.incident_bundle_manifest_files FROM PUBLIC;

REVOKE ALL ON TYPE public.incident_bundle_manifest_files FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.incident_bundle_manifest_files TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.incident_bundle_manifest_files TO cartulary_recovery;

CREATE INDEX incident_bundle_exports_export_job_id_scope_idx ON public.incident_bundle_exports (incident_id, export_job_id);

CREATE INDEX incident_bundle_exports_exported_by_user_id_fk_idx ON public.incident_bundle_exports USING btree (exported_by_user_id);

CREATE INDEX incident_bundle_exports_incident_lookup_idx ON public.incident_bundle_exports USING btree (incident_id, exported_at DESC, bundle_id);

CREATE INDEX incident_bundle_imported_actors_local_user_id_fk_idx ON public.incident_bundle_imported_actors USING btree (local_user_id);

CREATE INDEX incident_bundle_imported_attributions_actor_idx ON public.incident_bundle_imported_attributions USING btree (incident_id, source_actor_id);

CREATE INDEX incident_bundle_imported_attributions_local_user_id_fk_idx ON public.incident_bundle_imported_attributions USING btree (local_user_id);

CREATE INDEX incident_bundle_job_payloads_actor_idx ON public.incident_bundle_job_payloads USING btree (actor_user_id, created_at DESC);

CREATE INDEX incident_bundle_job_payloads_incident_id_fk_idx ON public.incident_bundle_job_payloads USING btree (incident_id);

CREATE INDEX incident_bundle_job_payloads_job_id_scope_idx ON public.incident_bundle_job_payloads (incident_id, job_id);

ALTER TABLE public.incident_bundle_exports ADD CONSTRAINT incident_bundle_exports_export_job_id_fkey FOREIGN KEY (incident_id, export_job_id) REFERENCES public.jobs(incident_id, job_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.incident_bundle_exports ADD CONSTRAINT incident_bundle_exports_exported_by_user_id_fkey FOREIGN KEY (exported_by_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.incident_bundle_exports ADD CONSTRAINT incident_bundle_exports_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.incident_bundle_imported_actors ADD CONSTRAINT incident_bundle_imported_actors_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.incident_bundle_imported_actors ADD CONSTRAINT incident_bundle_imported_actors_local_user_id_fkey FOREIGN KEY (local_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.incident_bundle_imported_attributions ADD CONSTRAINT incident_bundle_imported_attributions_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.incident_bundle_imported_attributions ADD CONSTRAINT incident_bundle_imported_attributions_local_user_id_fkey FOREIGN KEY (local_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.incident_bundle_job_payloads ADD CONSTRAINT incident_bundle_job_payloads_actor_user_id_fkey FOREIGN KEY (actor_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.incident_bundle_job_payloads ADD CONSTRAINT incident_bundle_job_payloads_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.incident_bundle_job_payloads ADD CONSTRAINT incident_bundle_job_payloads_job_id_fkey FOREIGN KEY (job_id) REFERENCES public.jobs(job_id) ON DELETE CASCADE ON UPDATE NO ACTION;

-- Import jobs are deployment-scoped. Keep the global FK for nullable import
-- payloads; the additional composite FK binds incident-scoped exports.
ALTER TABLE public.incident_bundle_job_payloads ADD CONSTRAINT incident_bundle_job_payloads_incident_job_fkey FOREIGN KEY (incident_id, job_id) REFERENCES public.jobs(incident_id, job_id) ON DELETE CASCADE ON UPDATE NO ACTION;


ALTER TABLE public.incident_bundle_manifest_files ADD CONSTRAINT incident_bundle_manifest_files_bundle_id_fkey FOREIGN KEY (bundle_id) REFERENCES public.incident_bundle_exports(bundle_id) ON DELETE CASCADE ON UPDATE NO ACTION;

-- Incident Bundles owns historical attribution. Source owners declare the exact
-- mutable attribution tuples below. Effective native mutations retire only the
-- affected override in their transaction; immutable creation/history is retained.
-- No source-table query, native-account mutation, or historical backfill occurs.

-- +goose StatementBegin
CREATE FUNCTION public.incident_bundle_retire_mutated_attribution() RETURNS trigger
LANGUAGE plpgsql SET search_path = pg_catalog, public AS $$
DECLARE
    before_row jsonb := to_jsonb(OLD);
    after_row jsonb := to_jsonb(NEW);
    changed boolean := false;
    position integer;
    scoped_incident uuid;
BEGIN
    IF TG_OP <> 'UPDATE' OR TG_NARGS < 3 THEN
        RAISE EXCEPTION 'invalid historical attribution mutation binding';
    END IF;
    FOR position IN 2..TG_NARGS-1 LOOP
        IF NOT (before_row ? TG_ARGV[position]) OR NOT (after_row ? TG_ARGV[position]) THEN
            RAISE EXCEPTION 'historical attribution mutation field absent';
        END IF;
        changed := changed OR (before_row -> TG_ARGV[position]) IS DISTINCT FROM (after_row -> TG_ARGV[position]);
    END LOOP;
    IF changed THEN
        IF before_row ->> TG_ARGV[1] IS NULL OR (TG_ARGV[0] <> '' AND before_row ->> TG_ARGV[0] IS NULL) THEN
            RAISE EXCEPTION 'historical attribution mutation identity absent';
        END IF;
        scoped_incident := CASE WHEN TG_ARGV[0] <> '' THEN (before_row ->> TG_ARGV[0])::uuid ELSE NULL END;
        DELETE FROM public.incident_bundle_imported_attributions
        WHERE (scoped_incident IS NULL OR incident_id = scoped_incident)
          AND source_table = TG_TABLE_NAME
          AND source_row_id = before_row ->> TG_ARGV[1]
          AND source_column = TG_ARGV[2];
    END IF;
    RETURN NEW;
END;
$$;
-- +goose StatementEnd
REVOKE ALL ON FUNCTION public.incident_bundle_retire_mutated_attribution() FROM PUBLIC;

CREATE TRIGGER incident_bundle_attribution_mutation_01
AFTER UPDATE OF updated_by_user_id, incident_version, updated_at ON public.incidents
FOR EACH ROW EXECUTE FUNCTION public.incident_bundle_retire_mutated_attribution('id', 'id', 'updated_by_user_id', 'incident_version', 'updated_at');

CREATE TRIGGER incident_bundle_attribution_mutation_02
AFTER UPDATE OF updated_by_user_id, row_version, updated_at ON public.records
FOR EACH ROW EXECUTE FUNCTION public.incident_bundle_retire_mutated_attribution('incident_id', 'record_id', 'updated_by_user_id', 'row_version', 'updated_at');

CREATE TRIGGER incident_bundle_attribution_mutation_03
AFTER UPDATE OF deleted_by_user_id, deleted_at ON public.records
FOR EACH ROW EXECUTE FUNCTION public.incident_bundle_retire_mutated_attribution('incident_id', 'record_id', 'deleted_by_user_id', 'deleted_at');

CREATE TRIGGER incident_bundle_attribution_mutation_04
AFTER UPDATE OF updated_by_user_id, row_version, updated_at ON public.hosts
FOR EACH ROW EXECUTE FUNCTION public.incident_bundle_retire_mutated_attribution('incident_id', 'record_id', 'updated_by_user_id', 'row_version', 'updated_at');

CREATE TRIGGER incident_bundle_attribution_mutation_05
AFTER UPDATE OF updated_by_user_id, row_version, updated_at ON public.identities
FOR EACH ROW EXECUTE FUNCTION public.incident_bundle_retire_mutated_attribution('incident_id', 'record_id', 'updated_by_user_id', 'row_version', 'updated_at');

CREATE TRIGGER incident_bundle_attribution_mutation_06
AFTER UPDATE OF updated_by_user_id, profile_version, updated_at ON public.timeline_time_conversion_profiles
FOR EACH ROW EXECUTE FUNCTION public.incident_bundle_retire_mutated_attribution('incident_id', 'incident_id', 'updated_by_user_id', 'profile_version', 'updated_at');

CREATE TRIGGER incident_bundle_attribution_mutation_07
AFTER UPDATE OF reviewed_by_user_id, reviewed_at ON public.timeline_events
FOR EACH ROW EXECUTE FUNCTION public.incident_bundle_retire_mutated_attribution('incident_id', 'record_id', 'reviewed_by_user_id', 'reviewed_at');

CREATE TRIGGER incident_bundle_attribution_mutation_08
AFTER UPDATE OF superseded_by_user_id, superseded_at ON public.timeline_events
FOR EACH ROW EXECUTE FUNCTION public.incident_bundle_retire_mutated_attribution('incident_id', 'record_id', 'superseded_by_user_id', 'superseded_at');

-- Entity mention identity is a global primary key; the row has no incident column.
-- Its declared relation/identity still names one source row across all incidents.
CREATE TRIGGER incident_bundle_attribution_mutation_09
AFTER UPDATE OF resolved_by_user_id, resolved_at ON public.entity_mentions
FOR EACH ROW EXECUTE FUNCTION public.incident_bundle_retire_mutated_attribution('', 'entity_mention_id', 'resolved_by_user_id', 'resolved_at');

CREATE TRIGGER incident_bundle_attribution_mutation_10
AFTER UPDATE OF resolved_by_user_id, resolved_at ON public.indicator_observations
FOR EACH ROW EXECUTE FUNCTION public.incident_bundle_retire_mutated_attribution('incident_id', 'indicator_observation_id', 'resolved_by_user_id', 'resolved_at');

CREATE TRIGGER incident_bundle_attribution_mutation_11
AFTER UPDATE OF deleted_by_user_id, deleted_at ON public.indicator_observations
FOR EACH ROW EXECUTE FUNCTION public.incident_bundle_retire_mutated_attribution('incident_id', 'indicator_observation_id', 'deleted_by_user_id', 'deleted_at');

CREATE TRIGGER incident_bundle_attribution_mutation_12
AFTER UPDATE OF deleted_by_user_id, deleted_at ON public.indicator_state_intervals
FOR EACH ROW EXECUTE FUNCTION public.incident_bundle_retire_mutated_attribution('incident_id', 'indicator_state_interval_id', 'deleted_by_user_id', 'deleted_at');

CREATE TRIGGER incident_bundle_attribution_mutation_13
AFTER UPDATE OF assessor_user_id, assessed_at ON public.assessments
FOR EACH ROW EXECUTE FUNCTION public.incident_bundle_retire_mutated_attribution('incident_id', 'record_id', 'assessor_user_id', 'assessed_at');

CREATE TRIGGER incident_bundle_attribution_mutation_14
AFTER UPDATE OF deleted_by_user_id, deleted_at ON public.assessments
FOR EACH ROW EXECUTE FUNCTION public.incident_bundle_retire_mutated_attribution('incident_id', 'record_id', 'deleted_by_user_id', 'deleted_at');

CREATE TRIGGER incident_bundle_attribution_mutation_15
AFTER UPDATE OF owner_user_id, decided_at ON public.record_links
FOR EACH ROW EXECUTE FUNCTION public.incident_bundle_retire_mutated_attribution('incident_id', 'record_link_id', 'owner_user_id', 'decided_at');

CREATE TRIGGER incident_bundle_attribution_mutation_16
AFTER UPDATE OF deleted_by_user_id, deleted_at ON public.record_links
FOR EACH ROW EXECUTE FUNCTION public.incident_bundle_retire_mutated_attribution('incident_id', 'record_link_id', 'deleted_by_user_id', 'deleted_at');

CREATE TRIGGER incident_bundle_attribution_mutation_17
AFTER UPDATE OF deleted_by_user_id, deleted_at ON public.record_tags
FOR EACH ROW EXECUTE FUNCTION public.incident_bundle_retire_mutated_attribution('incident_id', 'record_tag_id', 'deleted_by_user_id', 'deleted_at');

CREATE TRIGGER incident_bundle_attribution_mutation_18
AFTER UPDATE OF owner_user_id ON public.task_requests
FOR EACH ROW EXECUTE FUNCTION public.incident_bundle_retire_mutated_attribution('incident_id', 'record_id', 'owner_user_id');

CREATE TRIGGER incident_bundle_attribution_mutation_19
AFTER UPDATE OF owner_user_id ON public.decisions
FOR EACH ROW EXECUTE FUNCTION public.incident_bundle_retire_mutated_attribution('incident_id', 'record_id', 'owner_user_id');

CREATE TRIGGER incident_bundle_attribution_mutation_20
AFTER UPDATE OF owner_user_id ON public.saved_views
FOR EACH ROW EXECUTE FUNCTION public.incident_bundle_retire_mutated_attribution('incident_id', 'saved_view_id', 'owner_user_id');

-- +goose Down
-- Disposable harness cleanup only; no production downgrade or restore guarantee.
DROP TRIGGER incident_bundle_attribution_mutation_20 ON public.saved_views;
DROP TRIGGER incident_bundle_attribution_mutation_19 ON public.decisions;
DROP TRIGGER incident_bundle_attribution_mutation_18 ON public.task_requests;
DROP TRIGGER incident_bundle_attribution_mutation_17 ON public.record_tags;
DROP TRIGGER incident_bundle_attribution_mutation_16 ON public.record_links;
DROP TRIGGER incident_bundle_attribution_mutation_15 ON public.record_links;
DROP TRIGGER incident_bundle_attribution_mutation_14 ON public.assessments;
DROP TRIGGER incident_bundle_attribution_mutation_13 ON public.assessments;
DROP TRIGGER incident_bundle_attribution_mutation_12 ON public.indicator_state_intervals;
DROP TRIGGER incident_bundle_attribution_mutation_11 ON public.indicator_observations;
DROP TRIGGER incident_bundle_attribution_mutation_10 ON public.indicator_observations;
DROP TRIGGER incident_bundle_attribution_mutation_09 ON public.entity_mentions;
DROP TRIGGER incident_bundle_attribution_mutation_08 ON public.timeline_events;
DROP TRIGGER incident_bundle_attribution_mutation_07 ON public.timeline_events;
DROP TRIGGER incident_bundle_attribution_mutation_06 ON public.timeline_time_conversion_profiles;
DROP TRIGGER incident_bundle_attribution_mutation_05 ON public.identities;
DROP TRIGGER incident_bundle_attribution_mutation_04 ON public.hosts;
DROP TRIGGER incident_bundle_attribution_mutation_03 ON public.records;
DROP TRIGGER incident_bundle_attribution_mutation_02 ON public.records;
DROP TRIGGER incident_bundle_attribution_mutation_01 ON public.incidents;
DROP FUNCTION public.incident_bundle_retire_mutated_attribution();
ALTER TABLE public.incident_bundle_manifest_files DROP CONSTRAINT incident_bundle_manifest_files_bundle_id_fkey;
ALTER TABLE public.incident_bundle_job_payloads DROP CONSTRAINT incident_bundle_job_payloads_incident_job_fkey;
ALTER TABLE public.incident_bundle_job_payloads DROP CONSTRAINT incident_bundle_job_payloads_job_id_fkey;
ALTER TABLE public.incident_bundle_job_payloads DROP CONSTRAINT incident_bundle_job_payloads_incident_id_fkey;
ALTER TABLE public.incident_bundle_job_payloads DROP CONSTRAINT incident_bundle_job_payloads_actor_user_id_fkey;
ALTER TABLE public.incident_bundle_imported_attributions DROP CONSTRAINT incident_bundle_imported_attributions_local_user_id_fkey;
ALTER TABLE public.incident_bundle_imported_attributions DROP CONSTRAINT incident_bundle_imported_attributions_incident_id_fkey;
ALTER TABLE public.incident_bundle_imported_actors DROP CONSTRAINT incident_bundle_imported_actors_local_user_id_fkey;
ALTER TABLE public.incident_bundle_imported_actors DROP CONSTRAINT incident_bundle_imported_actors_incident_id_fkey;
ALTER TABLE public.incident_bundle_exports DROP CONSTRAINT incident_bundle_exports_incident_id_fkey;
ALTER TABLE public.incident_bundle_exports DROP CONSTRAINT incident_bundle_exports_exported_by_user_id_fkey;
ALTER TABLE public.incident_bundle_exports DROP CONSTRAINT incident_bundle_exports_export_job_id_fkey;
DROP TABLE public.incident_bundle_manifest_files;
DROP TABLE public.incident_bundle_job_payloads;
DROP TABLE public.incident_bundle_imported_attributions;
DROP TABLE public.incident_bundle_imported_actors;
DROP TABLE public.incident_bundle_exports;
