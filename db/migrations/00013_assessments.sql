-- +goose Up
-- PostgreSQL 18.6 consolidated baseline: assessments.
-- Administrator provisioning and final admission are separate lifecycle steps.

CREATE TABLE public.assessments (
    record_id uuid DEFAULT gen_random_uuid() NOT NULL,
    incident_id uuid NOT NULL,
    subject_record_id uuid NOT NULL,
    subject_type text NOT NULL,
    assessment_state text NOT NULL,
    confidence_score integer,
    assessor_user_id uuid NOT NULL,
    assessed_at timestamp with time zone DEFAULT now() NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    deleted_by_user_id uuid,
    rationale text NOT NULL,
    CONSTRAINT assessments_assessment_state_ck CHECK (assessment_state = ANY (ARRAY['unknown'::text, 'suspected'::text, 'confirmed'::text, 'disproven'::text, 'cleared'::text])),
    CONSTRAINT assessments_confidence_score_ck CHECK (confidence_score IS NULL OR confidence_score >= 0 AND confidence_score <= 100),
    CONSTRAINT assessments_subject_type_ck CHECK (subject_type = ANY (ARRAY['host'::text, 'identity'::text])),
    CONSTRAINT compromise_assessments_pkey PRIMARY KEY (record_id),
    CONSTRAINT assessments_incident_identity_key UNIQUE (incident_id, record_id)
);

REVOKE ALL ON TABLE public.assessments FROM PUBLIC;

REVOKE ALL ON TYPE public.assessments FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.assessments TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.assessments TO cartulary_recovery;

CREATE INDEX assessments_active_subject_lookup_idx ON public.assessments USING btree (incident_id, subject_type, subject_record_id, assessed_at DESC, record_id) WHERE (deleted_at IS NULL);

CREATE INDEX assessments_assessor_user_id_fk_idx ON public.assessments USING btree (assessor_user_id);

CREATE INDEX assessments_deleted_by_user_id_fk_idx ON public.assessments USING btree (deleted_by_user_id);

CREATE INDEX assessments_incident_id_record_id_fk_idx ON public.assessments USING btree (incident_id, record_id);

ALTER TABLE public.assessments ADD CONSTRAINT assessments_record_envelope_fkey FOREIGN KEY (incident_id, record_id) REFERENCES public.records(incident_id, record_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.assessments ADD CONSTRAINT compromise_assessments_assessed_by_user_id_fkey FOREIGN KEY (assessor_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.assessments ADD CONSTRAINT compromise_assessments_deleted_by_user_id_fkey FOREIGN KEY (deleted_by_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.assessments ADD CONSTRAINT compromise_assessments_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE ON UPDATE NO ACTION;

-- +goose Down
-- Disposable harness cleanup only; no production downgrade or restore guarantee.
ALTER TABLE public.assessments DROP CONSTRAINT compromise_assessments_incident_id_fkey;
ALTER TABLE public.assessments DROP CONSTRAINT compromise_assessments_deleted_by_user_id_fkey;
ALTER TABLE public.assessments DROP CONSTRAINT compromise_assessments_assessed_by_user_id_fkey;
ALTER TABLE public.assessments DROP CONSTRAINT assessments_record_envelope_fkey;
DROP TABLE public.assessments;
