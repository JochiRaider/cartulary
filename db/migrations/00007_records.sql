-- +goose Up
-- PostgreSQL 18.6 consolidated baseline: records.
-- Administrator provisioning and final admission are separate lifecycle steps.

CREATE TABLE public.records (
    record_id uuid DEFAULT gen_random_uuid() NOT NULL,
    incident_id uuid NOT NULL,
    record_type text NOT NULL,
    created_by_user_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_by_user_id uuid NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    row_version bigint DEFAULT 1 NOT NULL,
    deleted_at timestamp with time zone,
    deleted_by_user_id uuid,
    CONSTRAINT records_delete_state_ck CHECK (deleted_at IS NULL AND deleted_by_user_id IS NULL OR deleted_at IS NOT NULL AND deleted_by_user_id IS NOT NULL),
    CONSTRAINT records_incident_id_record_id_key UNIQUE (incident_id, record_id),
    CONSTRAINT records_pkey PRIMARY KEY (record_id),
    CONSTRAINT records_record_type_check CHECK (record_type = ANY (ARRAY['timeline_event'::text, 'host'::text, 'identity'::text, 'party'::text, 'indicator'::text, 'artifact'::text, 'task_request'::text, 'decision'::text, 'evidence'::text, 'assessment'::text])),
    CONSTRAINT records_row_version_check CHECK (row_version >= 1),
    CONSTRAINT records_incident_record_type_key UNIQUE (incident_id, record_id, record_type)
);

REVOKE ALL ON TABLE public.records FROM PUBLIC;

REVOKE ALL ON TYPE public.records FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.records TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.records TO cartulary_recovery;

CREATE INDEX records_active_incident_type_idx ON public.records USING btree (incident_id, record_type, updated_at DESC, record_id) WHERE (deleted_at IS NULL);

CREATE INDEX records_created_by_user_id_fk_idx ON public.records USING btree (created_by_user_id);

CREATE INDEX records_deleted_by_user_id_fk_idx ON public.records USING btree (deleted_by_user_id);

CREATE INDEX records_updated_by_user_id_fk_idx ON public.records USING btree (updated_by_user_id);

ALTER TABLE public.records ADD CONSTRAINT records_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.records ADD CONSTRAINT records_deleted_by_user_id_fkey FOREIGN KEY (deleted_by_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.records ADD CONSTRAINT records_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.records ADD CONSTRAINT records_updated_by_user_id_fkey FOREIGN KEY (updated_by_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

-- +goose Down
-- Disposable harness cleanup only; no production downgrade or restore guarantee.
ALTER TABLE public.records DROP CONSTRAINT records_updated_by_user_id_fkey;
ALTER TABLE public.records DROP CONSTRAINT records_incident_id_fkey;
ALTER TABLE public.records DROP CONSTRAINT records_deleted_by_user_id_fkey;
ALTER TABLE public.records DROP CONSTRAINT records_created_by_user_id_fkey;
DROP TABLE public.records;
