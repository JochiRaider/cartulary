-- +goose Up
-- PostgreSQL 18.6 consolidated baseline: artifacts.
-- Administrator provisioning and final admission are separate lifecycle steps.

-- +goose StatementBegin
CREATE FUNCTION public.cartulary_confidence_band(confidence_score integer)
 RETURNS text
 LANGUAGE sql
 IMMUTABLE
 SET search_path = pg_catalog, public
AS $$
    SELECT CASE
        WHEN confidence_score IS NULL THEN 'unset'
        WHEN confidence_score BETWEEN 0 AND 39 THEN 'low'
        WHEN confidence_score BETWEEN 40 AND 69 THEN 'medium'
        WHEN confidence_score BETWEEN 70 AND 100 THEN 'high'
        ELSE NULL
    END
$$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION public.cartulary_confidence_band(confidence_score integer) FROM PUBLIC;

GRANT EXECUTE ON FUNCTION public.cartulary_confidence_band(confidence_score integer) TO cartulary_runtime;

GRANT EXECUTE ON FUNCTION public.cartulary_confidence_band(confidence_score integer) TO cartulary_recovery;

CREATE TABLE public.artifact_findings (
    record_id uuid NOT NULL,
    incident_id uuid NOT NULL,
    kind text NOT NULL,
    statement text NOT NULL,
    state text NOT NULL,
    confidence_score integer,
    owner_user_id uuid NOT NULL,
    closed_at timestamp with time zone,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT artifact_findings_confidence_score_check CHECK (confidence_score IS NULL OR confidence_score >= 0 AND confidence_score <= 100),
    CONSTRAINT artifact_findings_kind_check CHECK (kind = ANY (ARRAY['finding'::text, 'hypothesis'::text])),
    CONSTRAINT artifact_findings_pkey PRIMARY KEY (record_id),
    CONSTRAINT artifact_findings_state_check CHECK (state = ANY (ARRAY['open'::text, 'closed'::text]))
);

REVOKE ALL ON TABLE public.artifact_findings FROM PUBLIC;

REVOKE ALL ON TYPE public.artifact_findings FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.artifact_findings TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.artifact_findings TO cartulary_recovery;

CREATE TABLE public.artifact_forensic_keywords (
    record_id uuid NOT NULL,
    incident_id uuid NOT NULL,
    keyword_id text NOT NULL,
    pattern text NOT NULL,
    reason text NOT NULL,
    match_mode text NOT NULL,
    case_sensitive boolean DEFAULT false NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT artifact_forensic_keywords_keyword_id_unique UNIQUE (incident_id, keyword_id),
    CONSTRAINT artifact_forensic_keywords_match_mode_check CHECK (match_mode = ANY (ARRAY['literal'::text, 'regex'::text])),
    CONSTRAINT artifact_forensic_keywords_pkey PRIMARY KEY (record_id)
);

REVOKE ALL ON TABLE public.artifact_forensic_keywords FROM PUBLIC;

REVOKE ALL ON TYPE public.artifact_forensic_keywords FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.artifact_forensic_keywords TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.artifact_forensic_keywords TO cartulary_recovery;

CREATE TABLE public.artifact_investigative_queries (
    record_id uuid NOT NULL,
    incident_id uuid NOT NULL,
    query_id text NOT NULL,
    platform text NOT NULL,
    purpose text NOT NULL,
    query_text text NOT NULL,
    created_by_user_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT artifact_investigative_queries_pkey PRIMARY KEY (record_id),
    CONSTRAINT artifact_investigative_queries_query_id_unique UNIQUE (incident_id, query_id)
);

REVOKE ALL ON TABLE public.artifact_investigative_queries FROM PUBLIC;

REVOKE ALL ON TYPE public.artifact_investigative_queries FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.artifact_investigative_queries TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.artifact_investigative_queries TO cartulary_recovery;

CREATE TABLE public.artifacts (
    record_id uuid DEFAULT gen_random_uuid() NOT NULL,
    incident_id uuid NOT NULL,
    artifact_type text NOT NULL,
    title text,
    body text,
    timestamp_utc timestamp with time zone,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    comm_id text,
    comm_type text,
    audience text,
    channel_or_meeting text,
    summary text,
    next_report_at timestamp with time zone,
    privilege_tag text,
    handoff_id text,
    outgoing_owner_user_id uuid,
    incoming_owner_user_id uuid,
    current_state_summary text,
    next_checks text,
    acknowledged_at timestamp with time zone,
    status_review_id text,
    review_owner_user_id uuid,
    active_risks_summary text,
    lesson_id text,
    owner_user_id uuid,
    closure_state text,
    created_by_user_id uuid,
    CONSTRAINT artifacts_artifact_type_check CHECK (artifact_type = ANY (ARRAY['note'::text, 'comm_log'::text, 'handoff'::text, 'status_review'::text, 'lesson'::text, 'finding'::text, 'investigative_query'::text, 'forensic_keyword'::text])),
    CONSTRAINT artifacts_pkey PRIMARY KEY (record_id),
    CONSTRAINT artifacts_incident_identity_key UNIQUE (incident_id, record_id)
);

REVOKE ALL ON TABLE public.artifacts FROM PUBLIC;

REVOKE ALL ON TYPE public.artifacts FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.artifacts TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.artifacts TO cartulary_recovery;

CREATE TABLE public.handoff_risk_refs (
    risk_ref_id uuid DEFAULT gen_random_uuid() NOT NULL,
    incident_id uuid NOT NULL,
    handoff_record_id uuid NOT NULL,
    risk_ref_text text NOT NULL,
    normalized_risk_ref_text text NOT NULL,
    created_by_user_id uuid NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    deleted_at timestamp with time zone,
    deleted_by_user_id uuid,
    CONSTRAINT handoff_risk_refs_delete_state_ck CHECK (deleted_at IS NULL AND deleted_by_user_id IS NULL OR deleted_at IS NOT NULL AND deleted_by_user_id IS NOT NULL),
    CONSTRAINT handoff_risk_refs_pkey PRIMARY KEY (risk_ref_id)
);

REVOKE ALL ON TABLE public.handoff_risk_refs FROM PUBLIC;

REVOKE ALL ON TYPE public.handoff_risk_refs FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.handoff_risk_refs TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.handoff_risk_refs TO cartulary_recovery;

CREATE INDEX artifact_findings_incident_closed_idx ON public.artifact_findings USING btree (incident_id, closed_at DESC, record_id);

CREATE INDEX artifact_findings_incident_confidence_idx ON public.artifact_findings USING btree (incident_id, confidence_score, record_id);

CREATE INDEX artifact_findings_incident_id_record_id_fk_idx ON public.artifact_findings USING btree (incident_id, record_id);

CREATE INDEX artifact_findings_incident_state_idx ON public.artifact_findings USING btree (incident_id, state, kind, owner_user_id, updated_at DESC, record_id);

CREATE INDEX artifact_findings_owner_user_id_fk_idx ON public.artifact_findings USING btree (owner_user_id);

CREATE INDEX artifact_forensic_keywords_incident_created_idx ON public.artifact_forensic_keywords USING btree (incident_id, created_at DESC, record_id);

CREATE INDEX artifact_forensic_keywords_incident_id_record_id_fk_idx ON public.artifact_forensic_keywords USING btree (incident_id, record_id);

CREATE INDEX artifact_forensic_keywords_incident_mode_idx ON public.artifact_forensic_keywords USING btree (incident_id, match_mode, case_sensitive, created_at DESC, record_id);

CREATE INDEX artifact_investigative_queries_created_by_user_id_fk_idx ON public.artifact_investigative_queries USING btree (created_by_user_id);

CREATE INDEX artifact_investigative_queries_incident_created_idx ON public.artifact_investigative_queries USING btree (incident_id, created_at DESC, record_id);

CREATE INDEX artifact_investigative_queries_incident_id_record_id_fk_idx ON public.artifact_investigative_queries USING btree (incident_id, record_id);

CREATE INDEX artifact_investigative_queries_incident_platform_idx ON public.artifact_investigative_queries USING btree (incident_id, platform, created_by_user_id, created_at DESC, record_id);

CREATE INDEX artifacts_created_by_user_id_fk_idx ON public.artifacts USING btree (created_by_user_id);

CREATE INDEX artifacts_incident_id_record_id_fk_idx ON public.artifacts USING btree (incident_id, record_id);

CREATE INDEX artifacts_incident_type_updated_idx ON public.artifacts USING btree (incident_id, artifact_type, updated_at DESC, record_id);

CREATE INDEX artifacts_incoming_owner_user_id_fk_idx ON public.artifacts USING btree (incoming_owner_user_id);

CREATE INDEX artifacts_outgoing_owner_user_id_fk_idx ON public.artifacts USING btree (outgoing_owner_user_id);

CREATE INDEX artifacts_owner_user_id_fk_idx ON public.artifacts USING btree (owner_user_id);

CREATE INDEX artifacts_review_owner_user_id_fk_idx ON public.artifacts USING btree (review_owner_user_id);

CREATE UNIQUE INDEX handoff_risk_refs_active_text_idx ON public.handoff_risk_refs USING btree (handoff_record_id, normalized_risk_ref_text) WHERE (deleted_at IS NULL);

CREATE INDEX handoff_risk_refs_created_by_user_id_fk_idx ON public.handoff_risk_refs USING btree (created_by_user_id);

CREATE INDEX handoff_risk_refs_deleted_by_user_id_fk_idx ON public.handoff_risk_refs USING btree (deleted_by_user_id);

CREATE INDEX handoff_risk_refs_incident_id_fk_idx ON public.handoff_risk_refs USING btree (incident_id);

CREATE INDEX handoff_risk_refs_incident_id_handoff_record_id_fk_idx ON public.handoff_risk_refs USING btree (incident_id, handoff_record_id);

ALTER TABLE public.artifact_findings ADD CONSTRAINT artifact_findings_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.artifact_findings ADD CONSTRAINT artifact_findings_owner_user_id_fkey FOREIGN KEY (owner_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.artifact_findings ADD CONSTRAINT artifact_findings_record_envelope_fkey FOREIGN KEY (incident_id, record_id) REFERENCES public.records(incident_id, record_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.artifact_findings ADD CONSTRAINT artifact_findings_record_id_fkey FOREIGN KEY (incident_id, record_id) REFERENCES public.artifacts(incident_id, record_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.artifact_forensic_keywords ADD CONSTRAINT artifact_forensic_keywords_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.artifact_forensic_keywords ADD CONSTRAINT artifact_forensic_keywords_record_envelope_fkey FOREIGN KEY (incident_id, record_id) REFERENCES public.records(incident_id, record_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.artifact_forensic_keywords ADD CONSTRAINT artifact_forensic_keywords_record_id_fkey FOREIGN KEY (incident_id, record_id) REFERENCES public.artifacts(incident_id, record_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.artifact_investigative_queries ADD CONSTRAINT artifact_investigative_queries_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.artifact_investigative_queries ADD CONSTRAINT artifact_investigative_queries_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.artifact_investigative_queries ADD CONSTRAINT artifact_investigative_queries_record_envelope_fkey FOREIGN KEY (incident_id, record_id) REFERENCES public.records(incident_id, record_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.artifact_investigative_queries ADD CONSTRAINT artifact_investigative_queries_record_id_fkey FOREIGN KEY (incident_id, record_id) REFERENCES public.artifacts(incident_id, record_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.artifacts ADD CONSTRAINT artifacts_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.artifacts ADD CONSTRAINT artifacts_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.artifacts ADD CONSTRAINT artifacts_incoming_owner_user_id_fkey FOREIGN KEY (incoming_owner_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.artifacts ADD CONSTRAINT artifacts_outgoing_owner_user_id_fkey FOREIGN KEY (outgoing_owner_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.artifacts ADD CONSTRAINT artifacts_owner_user_id_fkey FOREIGN KEY (owner_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.artifacts ADD CONSTRAINT artifacts_record_envelope_fkey FOREIGN KEY (incident_id, record_id) REFERENCES public.records(incident_id, record_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.artifacts ADD CONSTRAINT artifacts_review_owner_user_id_fkey FOREIGN KEY (review_owner_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.handoff_risk_refs ADD CONSTRAINT handoff_risk_refs_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.handoff_risk_refs ADD CONSTRAINT handoff_risk_refs_deleted_by_user_id_fkey FOREIGN KEY (deleted_by_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.handoff_risk_refs ADD CONSTRAINT handoff_risk_refs_handoff_fkey FOREIGN KEY (incident_id, handoff_record_id) REFERENCES public.records(incident_id, record_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.handoff_risk_refs ADD CONSTRAINT handoff_risk_refs_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE ON UPDATE NO ACTION;

-- +goose Down
-- Disposable harness cleanup only; no production downgrade or restore guarantee.
ALTER TABLE public.handoff_risk_refs DROP CONSTRAINT handoff_risk_refs_incident_id_fkey;
ALTER TABLE public.handoff_risk_refs DROP CONSTRAINT handoff_risk_refs_handoff_fkey;
ALTER TABLE public.handoff_risk_refs DROP CONSTRAINT handoff_risk_refs_deleted_by_user_id_fkey;
ALTER TABLE public.handoff_risk_refs DROP CONSTRAINT handoff_risk_refs_created_by_user_id_fkey;
ALTER TABLE public.artifacts DROP CONSTRAINT artifacts_review_owner_user_id_fkey;
ALTER TABLE public.artifacts DROP CONSTRAINT artifacts_record_envelope_fkey;
ALTER TABLE public.artifacts DROP CONSTRAINT artifacts_owner_user_id_fkey;
ALTER TABLE public.artifacts DROP CONSTRAINT artifacts_outgoing_owner_user_id_fkey;
ALTER TABLE public.artifacts DROP CONSTRAINT artifacts_incoming_owner_user_id_fkey;
ALTER TABLE public.artifacts DROP CONSTRAINT artifacts_incident_id_fkey;
ALTER TABLE public.artifacts DROP CONSTRAINT artifacts_created_by_user_id_fkey;
ALTER TABLE public.artifact_investigative_queries DROP CONSTRAINT artifact_investigative_queries_record_id_fkey;
ALTER TABLE public.artifact_investigative_queries DROP CONSTRAINT artifact_investigative_queries_record_envelope_fkey;
ALTER TABLE public.artifact_investigative_queries DROP CONSTRAINT artifact_investigative_queries_incident_id_fkey;
ALTER TABLE public.artifact_investigative_queries DROP CONSTRAINT artifact_investigative_queries_created_by_user_id_fkey;
ALTER TABLE public.artifact_forensic_keywords DROP CONSTRAINT artifact_forensic_keywords_record_id_fkey;
ALTER TABLE public.artifact_forensic_keywords DROP CONSTRAINT artifact_forensic_keywords_record_envelope_fkey;
ALTER TABLE public.artifact_forensic_keywords DROP CONSTRAINT artifact_forensic_keywords_incident_id_fkey;
ALTER TABLE public.artifact_findings DROP CONSTRAINT artifact_findings_record_id_fkey;
ALTER TABLE public.artifact_findings DROP CONSTRAINT artifact_findings_record_envelope_fkey;
ALTER TABLE public.artifact_findings DROP CONSTRAINT artifact_findings_owner_user_id_fkey;
ALTER TABLE public.artifact_findings DROP CONSTRAINT artifact_findings_incident_id_fkey;
DROP TABLE public.handoff_risk_refs;
DROP TABLE public.artifacts;
DROP TABLE public.artifact_investigative_queries;
DROP TABLE public.artifact_forensic_keywords;
DROP TABLE public.artifact_findings;
DROP FUNCTION public.cartulary_confidence_band(confidence_score integer);
