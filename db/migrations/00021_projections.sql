-- +goose Up
-- PostgreSQL 18.6 consolidated baseline: projections.
-- Administrator provisioning and final admission are separate lifecycle steps.

CREATE TABLE public.artifact_grid_projection (
    record_id uuid NOT NULL,
    incident_id uuid NOT NULL,
    row_version bigint NOT NULL,
    artifact_type text NOT NULL,
    title text,
    body text,
    timestamp_utc timestamp with time zone,
    updated_at timestamp with time zone NOT NULL,
    created_at timestamp with time zone NOT NULL,
    created_by_user_id uuid,
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
    finding_statement text,
    finding_kind text,
    finding_state text,
    finding_owner_user_id uuid,
    finding_confidence_score integer,
    finding_closed_at timestamp with time zone,
    finding_updated_at timestamp with time zone,
    finding_confidence_band text,
    investigative_query_query_id text,
    investigative_query_platform text,
    investigative_query_purpose text,
    investigative_query_query_text text,
    investigative_query_created_by_user_id uuid,
    investigative_query_created_at timestamp with time zone,
    investigative_query_created_day date,
    forensic_keyword_keyword_id text,
    forensic_keyword_pattern text,
    forensic_keyword_reason text,
    forensic_keyword_match_mode text,
    forensic_keyword_case_sensitive boolean,
    forensic_keyword_created_at timestamp with time zone,
    forensic_keyword_created_day date,
    timestamp_day date,
    next_report_day date,
    ack_state text NOT NULL,
    linked_record_count integer DEFAULT 0 NOT NULL,
    CONSTRAINT artifact_grid_projection_pkey PRIMARY KEY (record_id)
);

REVOKE ALL ON TABLE public.artifact_grid_projection FROM PUBLIC;

REVOKE ALL ON TYPE public.artifact_grid_projection FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.artifact_grid_projection TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.artifact_grid_projection TO cartulary_recovery;

CREATE TABLE public.assessment_grid_projection (
    record_id uuid NOT NULL,
    incident_id uuid NOT NULL,
    row_version bigint NOT NULL,
    subject_ref uuid NOT NULL,
    subject_type text NOT NULL,
    assessment_state text NOT NULL,
    confidence_score integer,
    confidence_band text NOT NULL,
    rationale text NOT NULL,
    assessor uuid NOT NULL,
    assessed_at timestamp with time zone NOT NULL,
    supporting_link_count integer DEFAULT 0 NOT NULL,
    CONSTRAINT assessment_grid_projection_assessment_state_check CHECK (assessment_state = ANY (ARRAY['unknown'::text, 'suspected'::text, 'confirmed'::text, 'disproven'::text, 'cleared'::text])),
    CONSTRAINT assessment_grid_projection_confidence_band_check CHECK (confidence_band = ANY (ARRAY['unset'::text, 'low'::text, 'medium'::text, 'high'::text])),
    CONSTRAINT assessment_grid_projection_confidence_score_check CHECK (confidence_score IS NULL OR confidence_score >= 0 AND confidence_score <= 100),
    CONSTRAINT assessment_grid_projection_pkey PRIMARY KEY (record_id),
    CONSTRAINT assessment_grid_projection_subject_type_check CHECK (subject_type = ANY (ARRAY['host'::text, 'identity'::text]))
);

REVOKE ALL ON TABLE public.assessment_grid_projection FROM PUBLIC;

REVOKE ALL ON TYPE public.assessment_grid_projection FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.assessment_grid_projection TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.assessment_grid_projection TO cartulary_recovery;

CREATE TABLE public.decision_grid_projection (
    record_id uuid NOT NULL,
    incident_id uuid NOT NULL,
    row_version bigint NOT NULL,
    summary text,
    status text NOT NULL,
    owner_user_id uuid,
    decision_type text,
    decided_at timestamp with time zone,
    rationale text,
    affected_record_count integer DEFAULT 0 NOT NULL,
    supersedes_record_id uuid,
    updated_at timestamp with time zone NOT NULL,
    is_superseded boolean DEFAULT false NOT NULL,
    CONSTRAINT decision_grid_projection_pkey PRIMARY KEY (record_id)
);

REVOKE ALL ON TABLE public.decision_grid_projection FROM PUBLIC;

REVOKE ALL ON TYPE public.decision_grid_projection FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.decision_grid_projection TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.decision_grid_projection TO cartulary_recovery;

CREATE TABLE public.evidence_grid_projection (
    record_id uuid NOT NULL,
    incident_id uuid NOT NULL,
    row_version bigint NOT NULL,
    title text,
    lifecycle_state text NOT NULL,
    requested_at timestamp with time zone,
    received_at timestamp with time zone,
    storage_ref text,
    blob_hash text,
    collector_party_text text,
    collector_party_id uuid,
    source_party_text text,
    source_party_id uuid,
    upload_state text NOT NULL,
    linked_record_count integer DEFAULT 0 NOT NULL,
    edited_at timestamp with time zone NOT NULL,
    CONSTRAINT evidence_grid_projection_pkey PRIMARY KEY (record_id)
);

REVOKE ALL ON TABLE public.evidence_grid_projection FROM PUBLIC;

REVOKE ALL ON TYPE public.evidence_grid_projection FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.evidence_grid_projection TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.evidence_grid_projection TO cartulary_recovery;

CREATE TABLE public.host_grid_projection (
    record_id uuid NOT NULL,
    incident_id uuid NOT NULL,
    row_version bigint NOT NULL,
    display_name text NOT NULL,
    hostname text,
    host_state text NOT NULL,
    linked_event_count integer DEFAULT 0 NOT NULL,
    evidence_count integer DEFAULT 0 NOT NULL,
    location text,
    os_platform text,
    business_owner text,
    criticality text,
    containment_status text,
    edited_at timestamp with time zone NOT NULL,
    CONSTRAINT host_grid_projection_host_state_check CHECK (host_state = ANY (ARRAY['stub'::text, 'canonical'::text, 'merged'::text])),
    CONSTRAINT host_grid_projection_pkey PRIMARY KEY (record_id)
);

REVOKE ALL ON TABLE public.host_grid_projection FROM PUBLIC;

REVOKE ALL ON TYPE public.host_grid_projection FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.host_grid_projection TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.host_grid_projection TO cartulary_recovery;

CREATE TABLE public.identity_grid_projection (
    record_id uuid NOT NULL,
    incident_id uuid NOT NULL,
    row_version bigint NOT NULL,
    display_name text NOT NULL,
    upn text,
    email public.citext,
    sam_account_name text,
    identity_state text NOT NULL,
    linked_event_count integer DEFAULT 0 NOT NULL,
    evidence_count integer DEFAULT 0 NOT NULL,
    privilege_level text,
    mfa_state text,
    reset_status text,
    edited_at timestamp with time zone NOT NULL,
    CONSTRAINT identity_grid_projection_identity_state_check CHECK (identity_state = ANY (ARRAY['stub'::text, 'canonical'::text, 'merged'::text])),
    CONSTRAINT identity_grid_projection_pkey PRIMARY KEY (record_id)
);

REVOKE ALL ON TABLE public.identity_grid_projection FROM PUBLIC;

REVOKE ALL ON TYPE public.identity_grid_projection FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.identity_grid_projection TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.identity_grid_projection TO cartulary_recovery;

CREATE TABLE public.indicator_grid_projection (
    record_id uuid NOT NULL,
    incident_id uuid NOT NULL,
    row_version bigint NOT NULL,
    indicator_type text NOT NULL,
    value_kind text NOT NULL,
    display_value text NOT NULL,
    normalized_value text,
    dedupe_key text NOT NULL,
    defanged_value text,
    hash_algorithm text,
    hash_value text,
    stix_pattern text,
    first_observed_at timestamp with time zone,
    last_observed_at timestamp with time zone,
    observation_count integer DEFAULT 0 NOT NULL,
    lifecycle_summary text,
    supporting_link_count integer DEFAULT 0 NOT NULL,
    edited_at timestamp with time zone NOT NULL,
    CONSTRAINT indicator_grid_projection_pkey PRIMARY KEY (record_id)
);

REVOKE ALL ON TABLE public.indicator_grid_projection FROM PUBLIC;

REVOKE ALL ON TYPE public.indicator_grid_projection FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.indicator_grid_projection TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.indicator_grid_projection TO cartulary_recovery;

CREATE TABLE public.party_grid_projection (
    record_id uuid NOT NULL,
    incident_id uuid NOT NULL,
    row_version bigint NOT NULL,
    display_name text,
    party_kind text,
    organization_name text,
    role_title text,
    primary_email text,
    timezone_name text,
    external_ref text,
    notes text,
    updated_at timestamp with time zone NOT NULL,
    CONSTRAINT party_grid_projection_pkey PRIMARY KEY (record_id)
);

REVOKE ALL ON TABLE public.party_grid_projection FROM PUBLIC;

REVOKE ALL ON TYPE public.party_grid_projection FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.party_grid_projection TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.party_grid_projection TO cartulary_recovery;

CREATE TABLE public.task_request_grid_projection (
    record_id uuid NOT NULL,
    incident_id uuid NOT NULL,
    row_version bigint NOT NULL,
    title text,
    status text NOT NULL,
    owner_user_id uuid,
    priority text,
    task_kind text,
    workstream text,
    due_at timestamp with time zone,
    requester_party_text text,
    requester_party_id uuid,
    blocked_reason text,
    completed_at timestamp with time zone,
    external_ticket_ref text,
    closure_summary text,
    decision_record_id uuid,
    linked_record_count integer DEFAULT 0 NOT NULL,
    updated_at timestamp with time zone NOT NULL,
    no_owner boolean DEFAULT false NOT NULL,
    CONSTRAINT task_request_grid_projection_pkey PRIMARY KEY (record_id)
);

REVOKE ALL ON TABLE public.task_request_grid_projection FROM PUBLIC;

REVOKE ALL ON TYPE public.task_request_grid_projection FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.task_request_grid_projection TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.task_request_grid_projection TO cartulary_recovery;

CREATE TABLE public.timeline_grid_projection (
    record_id uuid NOT NULL,
    incident_id uuid NOT NULL,
    row_version bigint NOT NULL,
    recorded_at timestamp with time zone NOT NULL,
    edited_at timestamp with time zone NOT NULL,
    capture_state text NOT NULL,
    replacement_record_id uuid,
    evidence_count integer DEFAULT 0 NOT NULL,
    has_evidence boolean DEFAULT false NOT NULL,
    has_unresolved_mentions boolean DEFAULT false NOT NULL,
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
    activity_sort_ts timestamp with time zone,
    date_entered_sort_day date,
    activity_time_pair_state text DEFAULT 'disabled'::text NOT NULL,
    host_refs jsonb DEFAULT '[]'::jsonb NOT NULL,
    identity_refs jsonb DEFAULT '[]'::jsonb NOT NULL,
    tags jsonb DEFAULT '[]'::jsonb NOT NULL,
    attached_evidence_refs jsonb DEFAULT '[]'::jsonb NOT NULL,
    CONSTRAINT timeline_grid_projection_activity_time_pair_state_check CHECK (activity_time_pair_state = ANY (ARRAY['disabled'::text, 'empty'::text, 'paired_generated'::text, 'paired_user_preserved'::text, 'paired_mismatch'::text, 'conversion_unavailable'::text])),
    CONSTRAINT timeline_grid_projection_attached_evidence_refs_array_ck CHECK (jsonb_typeof(attached_evidence_refs) = 'array'::text),
    CONSTRAINT timeline_grid_projection_capture_state_check CHECK (capture_state = ANY (ARRAY['rough'::text, 'enriched'::text, 'reviewed'::text, 'superseded'::text])),
    CONSTRAINT timeline_grid_projection_host_refs_array_ck CHECK (jsonb_typeof(host_refs) = 'array'::text),
    CONSTRAINT timeline_grid_projection_identity_refs_array_ck CHECK (jsonb_typeof(identity_refs) = 'array'::text),
    CONSTRAINT timeline_grid_projection_pkey PRIMARY KEY (record_id),
    CONSTRAINT timeline_grid_projection_tags_array_ck CHECK (jsonb_typeof(tags) = 'array'::text)
);

REVOKE ALL ON TABLE public.timeline_grid_projection FROM PUBLIC;

REVOKE ALL ON TYPE public.timeline_grid_projection FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.timeline_grid_projection TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.timeline_grid_projection TO cartulary_recovery;

CREATE INDEX artifact_grid_projection_created_by_user_id_fk_idx ON public.artifact_grid_projection USING btree (created_by_user_id);

CREATE INDEX artifact_grid_projection_finding_owner_user_id_fk_idx ON public.artifact_grid_projection USING btree (finding_owner_user_id);

CREATE INDEX artifact_grid_projection_finding_state_idx ON public.artifact_grid_projection USING btree (incident_id, artifact_type, finding_state, finding_kind, finding_owner_user_id, finding_updated_at DESC, record_id);

CREATE INDEX artifact_grid_projection_incident_type_updated_idx ON public.artifact_grid_projection USING btree (incident_id, artifact_type, updated_at DESC, record_id);

CREATE INDEX artifact_grid_projection_incoming_owner_user_id_fk_idx ON public.artifact_grid_projection USING btree (incoming_owner_user_id);

CREATE INDEX artifact_grid_projection_investigative_query_created_b_47384917 ON public.artifact_grid_projection USING btree (investigative_query_created_by_user_id);

CREATE INDEX artifact_grid_projection_keyword_mode_idx ON public.artifact_grid_projection USING btree (incident_id, artifact_type, forensic_keyword_match_mode, forensic_keyword_case_sensitive, forensic_keyword_created_at DESC, record_id);

CREATE INDEX artifact_grid_projection_note_linked_idx ON public.artifact_grid_projection USING btree (incident_id, artifact_type, linked_record_count DESC, updated_at DESC, record_id);

CREATE INDEX artifact_grid_projection_outgoing_owner_user_id_fk_idx ON public.artifact_grid_projection USING btree (outgoing_owner_user_id);

CREATE INDEX artifact_grid_projection_owner_user_id_fk_idx ON public.artifact_grid_projection USING btree (owner_user_id);

CREATE INDEX artifact_grid_projection_query_platform_idx ON public.artifact_grid_projection USING btree (incident_id, artifact_type, investigative_query_platform, investigative_query_created_by_user_id, investigative_query_created_at DESC, record_id);

CREATE INDEX artifact_grid_projection_record_id_scope_idx ON public.artifact_grid_projection (incident_id, record_id);

CREATE INDEX artifact_grid_projection_review_owner_user_id_fk_idx ON public.artifact_grid_projection USING btree (review_owner_user_id);

CREATE INDEX assessment_grid_projection_assessor_fk_idx ON public.assessment_grid_projection USING btree (assessor);

CREATE INDEX assessment_grid_projection_incident_confidence_band_idx ON public.assessment_grid_projection USING btree (incident_id, confidence_band, assessed_at DESC, record_id);

CREATE INDEX assessment_grid_projection_incident_sort_idx ON public.assessment_grid_projection USING btree (incident_id, assessed_at DESC, record_id);

CREATE INDEX assessment_grid_projection_incident_state_idx ON public.assessment_grid_projection USING btree (incident_id, assessment_state, assessed_at DESC, record_id);

CREATE INDEX assessment_grid_projection_record_id_scope_idx ON public.assessment_grid_projection (incident_id, record_id);

CREATE INDEX decision_grid_projection_incident_decided_idx ON public.decision_grid_projection USING btree (incident_id, decided_at DESC, record_id);

CREATE INDEX decision_grid_projection_owner_user_id_fk_idx ON public.decision_grid_projection USING btree (owner_user_id);

CREATE INDEX decision_grid_projection_record_id_scope_idx ON public.decision_grid_projection (incident_id, record_id);

CREATE INDEX decision_grid_projection_review_queue_idx ON public.decision_grid_projection USING btree (incident_id, status, owner_user_id, decision_type, decided_at, record_id);

CREATE INDEX decision_grid_projection_superseded_idx ON public.decision_grid_projection USING btree (incident_id, is_superseded, decided_at DESC, record_id);

CREATE INDEX evidence_grid_projection_incident_requested_idx ON public.evidence_grid_projection USING btree (incident_id, requested_at DESC, record_id);

CREATE INDEX evidence_grid_projection_lifecycle_idx ON public.evidence_grid_projection USING btree (incident_id, lifecycle_state, upload_state, edited_at DESC, record_id);

CREATE INDEX evidence_grid_projection_record_id_scope_idx ON public.evidence_grid_projection (incident_id, record_id);

CREATE INDEX host_grid_projection_incident_sort_idx ON public.host_grid_projection USING btree (incident_id, display_name, record_id);

CREATE INDEX host_grid_projection_record_id_scope_idx ON public.host_grid_projection (incident_id, record_id);

CREATE INDEX identity_grid_projection_incident_sort_idx ON public.identity_grid_projection USING btree (incident_id, display_name, record_id);

CREATE INDEX identity_grid_projection_record_id_scope_idx ON public.identity_grid_projection (incident_id, record_id);

CREATE INDEX indicator_grid_projection_incident_lifecycle_idx ON public.indicator_grid_projection USING btree (incident_id, lifecycle_summary, record_id);

CREATE INDEX indicator_grid_projection_incident_sort_idx ON public.indicator_grid_projection USING btree (incident_id, indicator_type, display_value, record_id);

CREATE INDEX indicator_grid_projection_record_id_scope_idx ON public.indicator_grid_projection (incident_id, record_id);

CREATE INDEX party_grid_projection_incident_display_name_idx ON public.party_grid_projection USING btree (incident_id, display_name, record_id);

CREATE INDEX party_grid_projection_kind_idx ON public.party_grid_projection USING btree (incident_id, party_kind, updated_at DESC, record_id);

CREATE INDEX party_grid_projection_record_id_scope_idx ON public.party_grid_projection (incident_id, record_id);

CREATE INDEX task_request_grid_projection_due_idx ON public.task_request_grid_projection USING btree (incident_id, due_at, record_id);

CREATE INDEX task_request_grid_projection_incident_updated_idx ON public.task_request_grid_projection USING btree (incident_id, updated_at DESC, record_id);

CREATE INDEX task_request_grid_projection_no_owner_idx ON public.task_request_grid_projection USING btree (incident_id, no_owner, updated_at DESC, record_id);

CREATE INDEX task_request_grid_projection_owner_user_id_fk_idx ON public.task_request_grid_projection USING btree (owner_user_id);

CREATE INDEX task_request_grid_projection_queue_idx ON public.task_request_grid_projection USING btree (incident_id, status, owner_user_id, priority, due_at, record_id);

CREATE INDEX task_request_grid_projection_record_id_scope_idx ON public.task_request_grid_projection (incident_id, record_id);

CREATE INDEX timeline_grid_projection_incident_activity_sort_idx ON public.timeline_grid_projection USING btree (incident_id, activity_sort_ts, record_id);

CREATE INDEX timeline_grid_projection_record_id_scope_idx ON public.timeline_grid_projection (incident_id, record_id);

CREATE INDEX timeline_grid_projection_replacement_record_id_fk_idx ON public.timeline_grid_projection USING btree (replacement_record_id);

CREATE INDEX timeline_grid_projection_replacement_record_id_scope_idx ON public.timeline_grid_projection (incident_id, replacement_record_id);

ALTER TABLE public.artifact_grid_projection ADD CONSTRAINT artifact_grid_projection_created_by_user_id_fkey FOREIGN KEY (created_by_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.artifact_grid_projection ADD CONSTRAINT artifact_grid_projection_finding_owner_user_id_fkey FOREIGN KEY (finding_owner_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.artifact_grid_projection ADD CONSTRAINT artifact_grid_projection_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.artifact_grid_projection ADD CONSTRAINT artifact_grid_projection_incoming_owner_user_id_fkey FOREIGN KEY (incoming_owner_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.artifact_grid_projection ADD CONSTRAINT artifact_grid_projection_investigative_query_created_by_us_fkey FOREIGN KEY (investigative_query_created_by_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.artifact_grid_projection ADD CONSTRAINT artifact_grid_projection_outgoing_owner_user_id_fkey FOREIGN KEY (outgoing_owner_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.artifact_grid_projection ADD CONSTRAINT artifact_grid_projection_owner_user_id_fkey FOREIGN KEY (owner_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.artifact_grid_projection ADD CONSTRAINT artifact_grid_projection_record_id_fkey FOREIGN KEY (incident_id, record_id) REFERENCES public.artifacts(incident_id, record_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.artifact_grid_projection ADD CONSTRAINT artifact_grid_projection_review_owner_user_id_fkey FOREIGN KEY (review_owner_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.assessment_grid_projection ADD CONSTRAINT assessment_grid_projection_assessor_fkey FOREIGN KEY (assessor) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.assessment_grid_projection ADD CONSTRAINT assessment_grid_projection_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.assessment_grid_projection ADD CONSTRAINT assessment_grid_projection_record_id_fkey FOREIGN KEY (incident_id, record_id) REFERENCES public.assessments(incident_id, record_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.decision_grid_projection ADD CONSTRAINT decision_grid_projection_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.decision_grid_projection ADD CONSTRAINT decision_grid_projection_owner_user_id_fkey FOREIGN KEY (owner_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.decision_grid_projection ADD CONSTRAINT decision_grid_projection_record_id_fkey FOREIGN KEY (incident_id, record_id) REFERENCES public.decisions(incident_id, record_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.evidence_grid_projection ADD CONSTRAINT evidence_grid_projection_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.evidence_grid_projection ADD CONSTRAINT evidence_grid_projection_record_id_fkey FOREIGN KEY (incident_id, record_id) REFERENCES public.evidence(incident_id, record_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.host_grid_projection ADD CONSTRAINT host_grid_projection_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.host_grid_projection ADD CONSTRAINT host_grid_projection_record_id_fkey FOREIGN KEY (incident_id, record_id) REFERENCES public.hosts(incident_id, record_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.identity_grid_projection ADD CONSTRAINT identity_grid_projection_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.identity_grid_projection ADD CONSTRAINT identity_grid_projection_record_id_fkey FOREIGN KEY (incident_id, record_id) REFERENCES public.identities(incident_id, record_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.indicator_grid_projection ADD CONSTRAINT indicator_grid_projection_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.indicator_grid_projection ADD CONSTRAINT indicator_grid_projection_record_id_fkey FOREIGN KEY (incident_id, record_id) REFERENCES public.indicators(incident_id, record_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.party_grid_projection ADD CONSTRAINT party_grid_projection_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.party_grid_projection ADD CONSTRAINT party_grid_projection_record_id_fkey FOREIGN KEY (incident_id, record_id) REFERENCES public.parties(incident_id, record_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.task_request_grid_projection ADD CONSTRAINT task_request_grid_projection_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.task_request_grid_projection ADD CONSTRAINT task_request_grid_projection_owner_user_id_fkey FOREIGN KEY (owner_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.task_request_grid_projection ADD CONSTRAINT task_request_grid_projection_record_id_fkey FOREIGN KEY (incident_id, record_id) REFERENCES public.task_requests(incident_id, record_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.timeline_grid_projection ADD CONSTRAINT timeline_grid_projection_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.timeline_grid_projection ADD CONSTRAINT timeline_grid_projection_record_id_fkey FOREIGN KEY (incident_id, record_id) REFERENCES public.timeline_events(incident_id, record_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.timeline_grid_projection ADD CONSTRAINT timeline_grid_projection_replacement_record_id_fkey FOREIGN KEY (incident_id, replacement_record_id) REFERENCES public.timeline_events(incident_id, record_id) ON UPDATE NO ACTION ON DELETE NO ACTION;

-- +goose Down
-- Disposable harness cleanup only; no production downgrade or restore guarantee.
ALTER TABLE public.timeline_grid_projection DROP CONSTRAINT timeline_grid_projection_replacement_record_id_fkey;
ALTER TABLE public.timeline_grid_projection DROP CONSTRAINT timeline_grid_projection_record_id_fkey;
ALTER TABLE public.timeline_grid_projection DROP CONSTRAINT timeline_grid_projection_incident_id_fkey;
ALTER TABLE public.task_request_grid_projection DROP CONSTRAINT task_request_grid_projection_record_id_fkey;
ALTER TABLE public.task_request_grid_projection DROP CONSTRAINT task_request_grid_projection_owner_user_id_fkey;
ALTER TABLE public.task_request_grid_projection DROP CONSTRAINT task_request_grid_projection_incident_id_fkey;
ALTER TABLE public.party_grid_projection DROP CONSTRAINT party_grid_projection_record_id_fkey;
ALTER TABLE public.party_grid_projection DROP CONSTRAINT party_grid_projection_incident_id_fkey;
ALTER TABLE public.indicator_grid_projection DROP CONSTRAINT indicator_grid_projection_record_id_fkey;
ALTER TABLE public.indicator_grid_projection DROP CONSTRAINT indicator_grid_projection_incident_id_fkey;
ALTER TABLE public.identity_grid_projection DROP CONSTRAINT identity_grid_projection_record_id_fkey;
ALTER TABLE public.identity_grid_projection DROP CONSTRAINT identity_grid_projection_incident_id_fkey;
ALTER TABLE public.host_grid_projection DROP CONSTRAINT host_grid_projection_record_id_fkey;
ALTER TABLE public.host_grid_projection DROP CONSTRAINT host_grid_projection_incident_id_fkey;
ALTER TABLE public.evidence_grid_projection DROP CONSTRAINT evidence_grid_projection_record_id_fkey;
ALTER TABLE public.evidence_grid_projection DROP CONSTRAINT evidence_grid_projection_incident_id_fkey;
ALTER TABLE public.decision_grid_projection DROP CONSTRAINT decision_grid_projection_record_id_fkey;
ALTER TABLE public.decision_grid_projection DROP CONSTRAINT decision_grid_projection_owner_user_id_fkey;
ALTER TABLE public.decision_grid_projection DROP CONSTRAINT decision_grid_projection_incident_id_fkey;
ALTER TABLE public.assessment_grid_projection DROP CONSTRAINT assessment_grid_projection_record_id_fkey;
ALTER TABLE public.assessment_grid_projection DROP CONSTRAINT assessment_grid_projection_incident_id_fkey;
ALTER TABLE public.assessment_grid_projection DROP CONSTRAINT assessment_grid_projection_assessor_fkey;
ALTER TABLE public.artifact_grid_projection DROP CONSTRAINT artifact_grid_projection_review_owner_user_id_fkey;
ALTER TABLE public.artifact_grid_projection DROP CONSTRAINT artifact_grid_projection_record_id_fkey;
ALTER TABLE public.artifact_grid_projection DROP CONSTRAINT artifact_grid_projection_owner_user_id_fkey;
ALTER TABLE public.artifact_grid_projection DROP CONSTRAINT artifact_grid_projection_outgoing_owner_user_id_fkey;
ALTER TABLE public.artifact_grid_projection DROP CONSTRAINT artifact_grid_projection_investigative_query_created_by_us_fkey;
ALTER TABLE public.artifact_grid_projection DROP CONSTRAINT artifact_grid_projection_incoming_owner_user_id_fkey;
ALTER TABLE public.artifact_grid_projection DROP CONSTRAINT artifact_grid_projection_incident_id_fkey;
ALTER TABLE public.artifact_grid_projection DROP CONSTRAINT artifact_grid_projection_finding_owner_user_id_fkey;
ALTER TABLE public.artifact_grid_projection DROP CONSTRAINT artifact_grid_projection_created_by_user_id_fkey;
DROP TABLE public.timeline_grid_projection;
DROP TABLE public.task_request_grid_projection;
DROP TABLE public.party_grid_projection;
DROP TABLE public.indicator_grid_projection;
DROP TABLE public.identity_grid_projection;
DROP TABLE public.host_grid_projection;
DROP TABLE public.evidence_grid_projection;
DROP TABLE public.decision_grid_projection;
DROP TABLE public.assessment_grid_projection;
DROP TABLE public.artifact_grid_projection;
