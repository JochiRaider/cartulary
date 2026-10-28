-- +goose Up
-- PostgreSQL 18.6 consolidated baseline: tasks and decisions.
-- Administrator provisioning and final admission are separate lifecycle steps.

CREATE TABLE public.decisions (
    record_id uuid DEFAULT gen_random_uuid() NOT NULL,
    incident_id uuid NOT NULL,
    summary text,
    status text DEFAULT 'proposed'::text NOT NULL,
    owner_user_id uuid,
    decision_type text,
    decided_at timestamp with time zone,
    rationale text,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT decisions_pkey PRIMARY KEY (record_id),
    CONSTRAINT decisions_incident_identity_key UNIQUE (incident_id, record_id)
);

REVOKE ALL ON TABLE public.decisions FROM PUBLIC;

REVOKE ALL ON TYPE public.decisions FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.decisions TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.decisions TO cartulary_recovery;

CREATE TABLE public.task_requests (
    record_id uuid DEFAULT gen_random_uuid() NOT NULL,
    incident_id uuid NOT NULL,
    title text,
    status text DEFAULT 'open'::text NOT NULL,
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
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    updated_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT task_requests_pkey PRIMARY KEY (record_id),
    CONSTRAINT task_requests_incident_identity_key UNIQUE (incident_id, record_id)
);

REVOKE ALL ON TABLE public.task_requests FROM PUBLIC;

REVOKE ALL ON TYPE public.task_requests FROM PUBLIC;

GRANT INSERT, SELECT, UPDATE, DELETE ON TABLE public.task_requests TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.task_requests TO cartulary_recovery;

CREATE INDEX decisions_incident_decided_idx ON public.decisions USING btree (incident_id, decided_at DESC, record_id);

CREATE INDEX decisions_incident_id_record_id_fk_idx ON public.decisions USING btree (incident_id, record_id);

CREATE INDEX decisions_owner_user_id_fk_idx ON public.decisions USING btree (owner_user_id);

CREATE INDEX task_requests_incident_id_record_id_fk_idx ON public.task_requests USING btree (incident_id, record_id);

CREATE INDEX task_requests_incident_updated_idx ON public.task_requests USING btree (incident_id, updated_at DESC, record_id);

CREATE INDEX task_requests_owner_user_id_fk_idx ON public.task_requests USING btree (owner_user_id);

CREATE INDEX task_requests_requester_party_ref_lookup_idx ON public.task_requests USING btree (incident_id, requester_party_id) WHERE (requester_party_id IS NOT NULL);

ALTER TABLE public.decisions ADD CONSTRAINT decisions_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.decisions ADD CONSTRAINT decisions_owner_user_id_fkey FOREIGN KEY (owner_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.decisions ADD CONSTRAINT decisions_record_envelope_fkey FOREIGN KEY (incident_id, record_id) REFERENCES public.records(incident_id, record_id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.task_requests ADD CONSTRAINT task_requests_incident_id_fkey FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON DELETE CASCADE ON UPDATE NO ACTION;

ALTER TABLE public.task_requests ADD CONSTRAINT task_requests_owner_user_id_fkey FOREIGN KEY (owner_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

ALTER TABLE public.task_requests ADD CONSTRAINT task_requests_record_envelope_fkey FOREIGN KEY (incident_id, record_id) REFERENCES public.records(incident_id, record_id) ON DELETE CASCADE ON UPDATE NO ACTION;

-- +goose Down
-- Disposable harness cleanup only; no production downgrade or restore guarantee.
ALTER TABLE public.task_requests DROP CONSTRAINT task_requests_record_envelope_fkey;
ALTER TABLE public.task_requests DROP CONSTRAINT task_requests_owner_user_id_fkey;
ALTER TABLE public.task_requests DROP CONSTRAINT task_requests_incident_id_fkey;
ALTER TABLE public.decisions DROP CONSTRAINT decisions_record_envelope_fkey;
ALTER TABLE public.decisions DROP CONSTRAINT decisions_owner_user_id_fkey;
ALTER TABLE public.decisions DROP CONSTRAINT decisions_incident_id_fkey;
DROP TABLE public.task_requests;
DROP TABLE public.decisions;
