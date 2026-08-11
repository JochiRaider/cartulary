-- +goose Up
CREATE TABLE public.deployment_admin_audit_events (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    actor_user_id uuid,
    target_user_id uuid,
    event_source text NOT NULL,
    event_kind text NOT NULL,
    before_json jsonb,
    after_json jsonb,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    reason_code text,
    client_txn_id text,
    request_id text,
    incident_id uuid,
    CONSTRAINT deployment_admin_audit_events_pkey PRIMARY KEY (id),
    CONSTRAINT deployment_admin_audit_events_actor_user_id_fkey
        FOREIGN KEY (actor_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION,
    CONSTRAINT deployment_admin_audit_events_target_user_id_fkey
        FOREIGN KEY (target_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION,
    CONSTRAINT deployment_admin_audit_events_incident_id_fkey
        FOREIGN KEY (incident_id) REFERENCES public.incidents(id) ON UPDATE NO ACTION ON DELETE NO ACTION
);

CREATE INDEX deployment_admin_audit_events_incident_lookup_idx
    ON public.deployment_admin_audit_events USING btree (incident_id, created_at DESC);

CREATE TABLE public.deployment_bootstrap_state (
    slot text NOT NULL,
    bootstrap_schema_id text NOT NULL,
    bootstrap_artifact_id uuid NOT NULL,
    artifact_sha256 bytea NOT NULL,
    created_user_id uuid NOT NULL,
    consumed_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT deployment_bootstrap_state_pkey PRIMARY KEY (slot),
    CONSTRAINT deployment_bootstrap_state_bootstrap_artifact_id_key UNIQUE (bootstrap_artifact_id),
    CONSTRAINT deployment_bootstrap_state_bootstrap_schema_id_check
        CHECK (bootstrap_schema_id = 'cartulary.bootstrap_admin.v1'::text),
    CONSTRAINT deployment_bootstrap_state_slot_check
        CHECK (slot = 'first_deployment_admin'::text),
    CONSTRAINT deployment_bootstrap_state_created_user_id_fkey
        FOREIGN KEY (created_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION
);

CREATE INDEX deployment_admin_audit_events_actor_user_id_fk_idx ON public.deployment_admin_audit_events (actor_user_id);
CREATE INDEX deployment_admin_audit_events_target_user_id_fk_idx ON public.deployment_admin_audit_events (target_user_id);
CREATE INDEX deployment_bootstrap_state_created_user_id_fk_idx ON public.deployment_bootstrap_state (created_user_id);

-- +goose Down
DROP TABLE public.deployment_bootstrap_state;
DROP TABLE public.deployment_admin_audit_events;
