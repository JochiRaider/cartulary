-- +goose Up
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
    CONSTRAINT administrative_audit_projections_pkey PRIMARY KEY (audit_event_id),
    CONSTRAINT administrative_audit_projections_raw_event_fkey FOREIGN KEY (audit_event_id)
        REFERENCES public.deployment_admin_audit_events(id) ON UPDATE NO ACTION ON DELETE NO ACTION,
    CONSTRAINT administrative_audit_projections_scope_check CHECK (
        (scope_kind = 'deployment' AND scope_id IS NULL)
        OR (scope_kind = 'incident' AND scope_id IS NOT NULL)
    ),
    CONSTRAINT administrative_audit_projections_actor_check CHECK (
        (actor_kind = 'user' AND actor_user_id IS NOT NULL)
        OR (actor_kind IN ('system', 'operator') AND actor_user_id IS NULL)
    ),
    CONSTRAINT administrative_audit_projections_source_check CHECK (
        source IN ('ui', 'api', 'startup', 'operator', 'system')
    ),
    CONSTRAINT administrative_audit_projections_changes_check CHECK (
        jsonb_typeof(changes) = 'array' AND jsonb_array_length(changes) > 0
    ),
    CONSTRAINT administrative_audit_projections_action_scope_check CHECK (
        (scope_kind = 'incident' AND action_code IN (
            'membership_created', 'membership_role_changed', 'membership_deleted'
        ))
        OR (scope_kind = 'deployment' AND action_code IN (
            'bootstrap_admin_created', 'user_created', 'user_profile_updated',
            'user_status_changed', 'deployment_admin_granted', 'deployment_admin_revoked',
            'password_changed', 'password_reset', 'totp_enrollment_begun',
            'totp_enrollment_completed', 'totp_reset', 'sessions_revoked',
            'auth_binding_created', 'auth_binding_rotated', 'auth_binding_retired',
            'account_preferences_updated', 'backup_created', 'restore_started',
            'restore_completed', 'restore_failed', 'restore_verification_completed'
        ))
    ),
    CONSTRAINT administrative_audit_projections_target_check CHECK (
        (action_code IN (
            'bootstrap_admin_created', 'user_created', 'user_profile_updated',
            'user_status_changed', 'deployment_admin_granted', 'deployment_admin_revoked',
            'password_changed', 'password_reset', 'totp_enrollment_begun',
            'totp_enrollment_completed', 'totp_reset', 'sessions_revoked'
        ) AND target_kind = 'user' AND target_id IS NOT NULL)
        OR (action_code = 'account_preferences_updated'
            AND target_kind = 'account_preferences' AND target_id IS NOT NULL)
        OR (action_code IN ('auth_binding_created', 'auth_binding_rotated', 'auth_binding_retired')
            AND target_kind = 'auth_binding' AND target_id IS NOT NULL)
        OR (action_code = 'backup_created'
            AND target_kind = 'backup_set' AND target_id IS NOT NULL)
        OR (action_code IN (
            'restore_started', 'restore_completed', 'restore_failed',
            'restore_verification_completed'
        ) AND target_kind = 'restore_operation' AND target_id IS NOT NULL)
        OR (action_code IN ('membership_created', 'membership_role_changed', 'membership_deleted')
            AND target_kind = 'incident_membership' AND target_id IS NOT NULL)
    )
);

CREATE INDEX administrative_audit_projections_deployment_page_idx
    ON public.administrative_audit_projections (occurred_at DESC, audit_event_id DESC)
    WHERE scope_kind = 'deployment';
CREATE INDEX administrative_audit_projections_incident_page_idx
    ON public.administrative_audit_projections (scope_id, occurred_at DESC, audit_event_id DESC)
    WHERE scope_kind = 'incident';
CREATE INDEX administrative_audit_projections_filter_idx
    ON public.administrative_audit_projections (
        scope_kind, scope_id, actor_user_id, action_code, target_kind, target_id,
        occurred_at DESC, audit_event_id DESC
    );

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

CREATE TRIGGER deployment_admin_audit_events_immutable
BEFORE UPDATE OR DELETE ON public.deployment_admin_audit_events
FOR EACH ROW EXECUTE FUNCTION public.reject_administrative_audit_mutation();

CREATE TRIGGER administrative_audit_projections_immutable
BEFORE UPDATE OR DELETE ON public.administrative_audit_projections
FOR EACH ROW EXECUTE FUNCTION public.reject_administrative_audit_mutation();

REVOKE ALL ON FUNCTION public.reject_administrative_audit_mutation() FROM PUBLIC;

-- +goose Down
DROP TRIGGER administrative_audit_projections_immutable ON public.administrative_audit_projections;
DROP TRIGGER deployment_admin_audit_events_immutable ON public.deployment_admin_audit_events;
DROP FUNCTION public.reject_administrative_audit_mutation();
DROP TABLE public.administrative_audit_projections;
