-- +goose Up
-- PostgreSQL 18.6 consolidated baseline: administrative audit.
-- Administrator provisioning and final admission are separate lifecycle steps.

CREATE TABLE public.administrative_audit_projections (
    audit_event_id uuid NOT NULL,
    scope_kind text NOT NULL,
    scope_id uuid,
    occurred_at timestamp with time zone NOT NULL,
    actor_kind text NOT NULL,
    actor_user_id uuid,
    source text NOT NULL,
    action_code text NOT NULL,
    target_kind text NOT NULL,
    target_id text,
    changes jsonb NOT NULL,
    reason_code text,
    CONSTRAINT administrative_audit_projections_action_scope_check CHECK (scope_kind = 'deployment'::text AND (action_code = ANY (ARRAY['account_preferences_updated'::text, 'auth_binding_created'::text, 'auth_binding_retired'::text, 'auth_binding_rotated'::text, 'backup_created'::text, 'bootstrap_admin_created'::text, 'deployment_admin_granted'::text, 'deployment_admin_revoked'::text, 'password_changed'::text, 'password_reset'::text, 'reference_pack_activation'::text, 'reference_pack_dependency_invalidation'::text, 'reference_pack_disablement'::text, 'reference_pack_exact_reimport'::text, 'reference_pack_import_admitted'::text, 'reference_pack_import_verification'::text, 'reference_pack_payload_invalidation'::text, 'reference_pack_profile_reconciliation'::text, 'reference_pack_refresh_admitted'::text, 'reference_pack_refresh_verification'::text, 'reference_pack_removal'::text, 'reference_pack_reverification'::text, 'reference_pack_reverify_admitted'::text, 'reference_pack_rollback_activation'::text, 'reference_pack_root_import'::text, 'reference_pack_safety_fallback'::text, 'reference_pack_sequence_rejection'::text, 'reference_pack_trust_root_rejection'::text, 'reference_pack_trust_root_update'::text, 'reference_pack_verification_completed'::text, 'restore_completed'::text, 'restore_failed'::text, 'restore_started'::text, 'restore_verification_completed'::text, 'sessions_revoked'::text, 'totp_enrollment_begun'::text, 'totp_enrollment_completed'::text, 'totp_reset'::text, 'user_created'::text, 'user_profile_updated'::text, 'user_status_changed'::text])) OR scope_kind = 'incident'::text AND (action_code = ANY (ARRAY['membership_created'::text, 'membership_deleted'::text, 'membership_role_changed'::text]))),
    CONSTRAINT administrative_audit_projections_actor_check CHECK (actor_kind = 'user'::text AND actor_user_id IS NOT NULL OR (actor_kind = ANY (ARRAY['system'::text, 'operator'::text])) AND actor_user_id IS NULL),
    CONSTRAINT administrative_audit_projections_changes_check CHECK (jsonb_typeof(changes) = 'array'::text AND jsonb_array_length(changes) > 0),
    CONSTRAINT administrative_audit_projections_pkey PRIMARY KEY (audit_event_id),
    CONSTRAINT administrative_audit_projections_scope_check CHECK (scope_kind = 'deployment'::text AND scope_id IS NULL OR scope_kind = 'incident'::text AND scope_id IS NOT NULL),
    CONSTRAINT administrative_audit_projections_source_check CHECK (source = ANY (ARRAY['ui'::text, 'api'::text, 'startup'::text, 'operator'::text, 'system'::text])),
    CONSTRAINT administrative_audit_projections_target_check CHECK (target_kind = 'account_preferences'::text AND target_id IS NOT NULL AND action_code = 'account_preferences_updated'::text OR target_kind = 'auth_binding'::text AND target_id IS NOT NULL AND (action_code = ANY (ARRAY['auth_binding_created'::text, 'auth_binding_retired'::text, 'auth_binding_rotated'::text])) OR target_kind = 'backup_set'::text AND target_id IS NOT NULL AND action_code = 'backup_created'::text OR target_kind = 'user'::text AND target_id IS NOT NULL AND (action_code = ANY (ARRAY['bootstrap_admin_created'::text, 'deployment_admin_granted'::text, 'deployment_admin_revoked'::text, 'password_changed'::text, 'password_reset'::text, 'sessions_revoked'::text, 'totp_enrollment_begun'::text, 'totp_enrollment_completed'::text, 'totp_reset'::text, 'user_created'::text, 'user_profile_updated'::text, 'user_status_changed'::text])) OR target_kind = 'incident_membership'::text AND target_id IS NOT NULL AND (action_code = ANY (ARRAY['membership_created'::text, 'membership_deleted'::text, 'membership_role_changed'::text])) OR target_kind = 'reference_pack_operation'::text AND target_id IS NOT NULL AND (action_code = ANY (ARRAY['reference_pack_activation'::text, 'reference_pack_dependency_invalidation'::text, 'reference_pack_disablement'::text, 'reference_pack_exact_reimport'::text, 'reference_pack_import_admitted'::text, 'reference_pack_import_verification'::text, 'reference_pack_payload_invalidation'::text, 'reference_pack_profile_reconciliation'::text, 'reference_pack_refresh_admitted'::text, 'reference_pack_refresh_verification'::text, 'reference_pack_removal'::text, 'reference_pack_reverification'::text, 'reference_pack_reverify_admitted'::text, 'reference_pack_rollback_activation'::text, 'reference_pack_root_import'::text, 'reference_pack_safety_fallback'::text, 'reference_pack_sequence_rejection'::text, 'reference_pack_trust_root_rejection'::text, 'reference_pack_trust_root_update'::text, 'reference_pack_verification_completed'::text])) OR target_kind = 'restore_operation'::text AND target_id IS NOT NULL AND (action_code = ANY (ARRAY['restore_completed'::text, 'restore_failed'::text, 'restore_started'::text, 'restore_verification_completed'::text])))
);

REVOKE ALL ON TABLE public.administrative_audit_projections FROM PUBLIC;

REVOKE ALL ON TYPE public.administrative_audit_projections FROM PUBLIC;

GRANT INSERT, SELECT ON TABLE public.administrative_audit_projections TO cartulary_runtime;

GRANT INSERT, SELECT, UPDATE, DELETE, TRUNCATE ON TABLE public.administrative_audit_projections TO cartulary_recovery;

-- +goose StatementBegin
CREATE FUNCTION public.reject_administrative_audit_mutation()
 RETURNS trigger
 LANGUAGE plpgsql
 SET search_path = pg_catalog, public
AS $$
BEGIN
    RAISE EXCEPTION 'administrative audit rows are immutable' USING ERRCODE = '55000';
END
$$;
-- +goose StatementEnd

REVOKE ALL ON FUNCTION public.reject_administrative_audit_mutation() FROM PUBLIC;

GRANT EXECUTE ON FUNCTION public.reject_administrative_audit_mutation() TO cartulary_runtime;

GRANT EXECUTE ON FUNCTION public.reject_administrative_audit_mutation() TO cartulary_recovery;

CREATE INDEX administrative_audit_projections_deployment_page_idx ON public.administrative_audit_projections USING btree (occurred_at DESC, audit_event_id DESC) WHERE (scope_kind = 'deployment'::text);

CREATE INDEX administrative_audit_projections_filter_idx ON public.administrative_audit_projections USING btree (scope_kind, scope_id, actor_user_id, action_code, target_kind, target_id, occurred_at DESC, audit_event_id DESC);

CREATE INDEX administrative_audit_projections_incident_page_idx ON public.administrative_audit_projections USING btree (scope_id, occurred_at DESC, audit_event_id DESC) WHERE (scope_kind = 'incident'::text);

ALTER TABLE public.administrative_audit_projections ADD CONSTRAINT administrative_audit_projections_raw_event_fkey FOREIGN KEY (audit_event_id) REFERENCES public.deployment_admin_audit_events(id) ON UPDATE NO ACTION ON DELETE NO ACTION;

CREATE TRIGGER administrative_audit_projections_immutable BEFORE DELETE OR UPDATE ON public.administrative_audit_projections FOR EACH ROW EXECUTE FUNCTION public.reject_administrative_audit_mutation();

CREATE TRIGGER deployment_admin_audit_events_immutable BEFORE DELETE OR UPDATE ON public.deployment_admin_audit_events FOR EACH ROW EXECUTE FUNCTION public.reject_administrative_audit_mutation();

-- +goose Down
-- Disposable harness cleanup only; no production downgrade or restore guarantee.
DROP TRIGGER deployment_admin_audit_events_immutable ON public.deployment_admin_audit_events;
DROP TRIGGER administrative_audit_projections_immutable ON public.administrative_audit_projections;
ALTER TABLE public.administrative_audit_projections DROP CONSTRAINT administrative_audit_projections_raw_event_fkey;
DROP FUNCTION public.reject_administrative_audit_mutation();
DROP TABLE public.administrative_audit_projections;
