-- +goose Up
LOCK TABLE public.administrative_audit_projections IN ACCESS EXCLUSIVE MODE;

DROP TRIGGER administrative_audit_projections_immutable
    ON public.administrative_audit_projections;

DELETE FROM public.administrative_audit_projections
WHERE
    action_code = 'legacy_administrative_event' OR
    target_kind = 'legacy_administrative_event';

ALTER TABLE public.administrative_audit_projections
    DROP CONSTRAINT administrative_audit_projections_changes_check,
    DROP CONSTRAINT administrative_audit_projections_action_scope_check,
    DROP CONSTRAINT administrative_audit_projections_target_check;

ALTER TABLE public.administrative_audit_projections
    ADD CONSTRAINT administrative_audit_projections_changes_check CHECK (
        jsonb_typeof(changes) = 'array' AND
        jsonb_array_length(changes) > 0
    ),
    ADD CONSTRAINT administrative_audit_projections_action_scope_check CHECK (
        (scope_kind = 'incident' AND action_code IN (
            'membership_created',
            'membership_role_changed',
            'membership_deleted'
        )) OR
        (scope_kind = 'deployment' AND action_code IN (
            'bootstrap_admin_created',
            'user_created',
            'user_profile_updated',
            'user_status_changed',
            'deployment_admin_granted',
            'deployment_admin_revoked',
            'password_changed',
            'password_reset',
            'totp_enrollment_begun',
            'totp_enrollment_completed',
            'totp_reset',
            'sessions_revoked',
            'auth_binding_created',
            'auth_binding_rotated',
            'auth_binding_retired',
            'account_preferences_updated',
            'backup_created',
            'restore_started',
            'restore_completed',
            'restore_failed',
            'restore_verification_completed'
        ))
    ),
    ADD CONSTRAINT administrative_audit_projections_target_check CHECK (
        (action_code IN (
            'bootstrap_admin_created',
            'user_created',
            'user_profile_updated',
            'user_status_changed',
            'deployment_admin_granted',
            'deployment_admin_revoked',
            'password_changed',
            'password_reset',
            'totp_enrollment_begun',
            'totp_enrollment_completed',
            'totp_reset',
            'sessions_revoked'
        ) AND target_kind = 'user' AND target_id IS NOT NULL) OR
        (action_code = 'account_preferences_updated' AND target_kind = 'account_preferences' AND target_id IS NOT NULL) OR
        (action_code IN ('auth_binding_created', 'auth_binding_rotated', 'auth_binding_retired') AND target_kind = 'auth_binding' AND target_id IS NOT NULL) OR
        (action_code = 'backup_created' AND target_kind = 'backup_set' AND target_id IS NOT NULL) OR
        (action_code IN ('restore_started', 'restore_completed', 'restore_failed', 'restore_verification_completed') AND target_kind = 'restore_operation' AND target_id IS NOT NULL) OR
        (action_code IN ('membership_created', 'membership_role_changed', 'membership_deleted') AND target_kind = 'incident_membership' AND target_id IS NOT NULL)
    );

-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
        FROM public.administrative_audit_projections
        WHERE
            action_code = 'legacy_administrative_event' OR
            target_kind = 'legacy_administrative_event'
    ) THEN
        RAISE EXCEPTION 'legacy administrative audit projections remain after cleanup';
    END IF;
END;
$$;
-- +goose StatementEnd

CREATE TRIGGER administrative_audit_projections_immutable
BEFORE UPDATE OR DELETE ON public.administrative_audit_projections
FOR EACH ROW EXECUTE FUNCTION public.reject_administrative_audit_mutation();

-- +goose Down
LOCK TABLE public.administrative_audit_projections IN ACCESS EXCLUSIVE MODE;

DROP TRIGGER administrative_audit_projections_immutable
    ON public.administrative_audit_projections;

ALTER TABLE public.administrative_audit_projections
    DROP CONSTRAINT administrative_audit_projections_changes_check,
    DROP CONSTRAINT administrative_audit_projections_action_scope_check,
    DROP CONSTRAINT administrative_audit_projections_target_check;

ALTER TABLE public.administrative_audit_projections
    ADD CONSTRAINT administrative_audit_projections_changes_check CHECK (
        jsonb_typeof(changes) = 'array' AND
        (action_code = 'legacy_administrative_event' OR jsonb_array_length(changes) > 0)
    ),
    ADD CONSTRAINT administrative_audit_projections_action_scope_check CHECK (
        (scope_kind = 'incident' AND action_code IN (
            'membership_created',
            'membership_role_changed',
            'membership_deleted',
            'legacy_administrative_event'
        )) OR
        (scope_kind = 'deployment' AND action_code IN (
            'bootstrap_admin_created',
            'user_created',
            'user_profile_updated',
            'user_status_changed',
            'deployment_admin_granted',
            'deployment_admin_revoked',
            'password_changed',
            'password_reset',
            'totp_enrollment_begun',
            'totp_enrollment_completed',
            'totp_reset',
            'sessions_revoked',
            'auth_binding_created',
            'auth_binding_rotated',
            'auth_binding_retired',
            'account_preferences_updated',
            'backup_created',
            'restore_started',
            'restore_completed',
            'restore_failed',
            'restore_verification_completed',
            'legacy_administrative_event'
        ))
    ),
    ADD CONSTRAINT administrative_audit_projections_target_check CHECK (
        (action_code = 'legacy_administrative_event' AND target_kind = 'legacy_administrative_event' AND target_id IS NULL) OR
        (action_code IN (
            'bootstrap_admin_created',
            'user_created',
            'user_profile_updated',
            'user_status_changed',
            'deployment_admin_granted',
            'deployment_admin_revoked',
            'password_changed',
            'password_reset',
            'totp_enrollment_begun',
            'totp_enrollment_completed',
            'totp_reset',
            'sessions_revoked'
        ) AND target_kind = 'user' AND target_id IS NOT NULL) OR
        (action_code = 'account_preferences_updated' AND target_kind = 'account_preferences' AND target_id IS NOT NULL) OR
        (action_code IN ('auth_binding_created', 'auth_binding_rotated', 'auth_binding_retired') AND target_kind = 'auth_binding' AND target_id IS NOT NULL) OR
        (action_code = 'backup_created' AND target_kind = 'backup_set' AND target_id IS NOT NULL) OR
        (action_code IN ('restore_started', 'restore_completed', 'restore_failed', 'restore_verification_completed') AND target_kind = 'restore_operation' AND target_id IS NOT NULL) OR
        (action_code IN ('membership_created', 'membership_role_changed', 'membership_deleted') AND target_kind = 'incident_membership' AND target_id IS NOT NULL)
    );

CREATE TRIGGER administrative_audit_projections_immutable
BEFORE UPDATE OR DELETE ON public.administrative_audit_projections
FOR EACH ROW EXECUTE FUNCTION public.reject_administrative_audit_mutation();
