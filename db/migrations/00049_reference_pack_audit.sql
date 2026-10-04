-- +goose Up
-- Add the Reference Pack operation audit projection without changing historical events.
ALTER TABLE public.administrative_audit_projections
    DROP CONSTRAINT administrative_audit_projections_action_scope_check,
    DROP CONSTRAINT administrative_audit_projections_target_check;
ALTER TABLE public.administrative_audit_projections
    ADD CONSTRAINT administrative_audit_projections_action_scope_check CHECK (
        (scope_kind='deployment' AND action_code IN ('account_preferences_updated', 'auth_binding_created', 'auth_binding_retired', 'auth_binding_rotated', 'backup_created', 'bootstrap_admin_created', 'deployment_admin_granted', 'deployment_admin_revoked', 'password_changed', 'password_reset', 'reference_pack_activation', 'reference_pack_dependency_invalidation', 'reference_pack_disablement', 'reference_pack_exact_reimport', 'reference_pack_import_admitted', 'reference_pack_import_verification', 'reference_pack_payload_invalidation', 'reference_pack_profile_reconciliation', 'reference_pack_refresh_admitted', 'reference_pack_refresh_verification', 'reference_pack_removal', 'reference_pack_reverification', 'reference_pack_reverify_admitted', 'reference_pack_rollback_activation', 'reference_pack_root_import', 'reference_pack_safety_fallback', 'reference_pack_sequence_rejection', 'reference_pack_trust_root_rejection', 'reference_pack_trust_root_update', 'reference_pack_verification_completed', 'restore_completed', 'restore_failed', 'restore_started', 'restore_verification_completed', 'sessions_revoked', 'totp_enrollment_begun', 'totp_enrollment_completed', 'totp_reset', 'user_created', 'user_profile_updated', 'user_status_changed'))
        OR
        (scope_kind='incident' AND action_code IN ('membership_created', 'membership_deleted', 'membership_role_changed'))
    ),
    ADD CONSTRAINT administrative_audit_projections_target_check CHECK (
        (target_kind='account_preferences' AND target_id IS NOT NULL AND action_code IN ('account_preferences_updated'))
        OR
        (target_kind='auth_binding' AND target_id IS NOT NULL AND action_code IN ('auth_binding_created', 'auth_binding_retired', 'auth_binding_rotated'))
        OR
        (target_kind='backup_set' AND target_id IS NOT NULL AND action_code IN ('backup_created'))
        OR
        (target_kind='user' AND target_id IS NOT NULL AND action_code IN ('bootstrap_admin_created', 'deployment_admin_granted', 'deployment_admin_revoked', 'password_changed', 'password_reset', 'sessions_revoked', 'totp_enrollment_begun', 'totp_enrollment_completed', 'totp_reset', 'user_created', 'user_profile_updated', 'user_status_changed'))
        OR
        (target_kind='incident_membership' AND target_id IS NOT NULL AND action_code IN ('membership_created', 'membership_deleted', 'membership_role_changed'))
        OR
        (target_kind='reference_pack_operation' AND target_id IS NOT NULL AND action_code IN ('reference_pack_activation', 'reference_pack_dependency_invalidation', 'reference_pack_disablement', 'reference_pack_exact_reimport', 'reference_pack_import_admitted', 'reference_pack_import_verification', 'reference_pack_payload_invalidation', 'reference_pack_profile_reconciliation', 'reference_pack_refresh_admitted', 'reference_pack_refresh_verification', 'reference_pack_removal', 'reference_pack_reverification', 'reference_pack_reverify_admitted', 'reference_pack_rollback_activation', 'reference_pack_root_import', 'reference_pack_safety_fallback', 'reference_pack_sequence_rejection', 'reference_pack_trust_root_rejection', 'reference_pack_trust_root_update', 'reference_pack_verification_completed'))
        OR
        (target_kind='restore_operation' AND target_id IS NOT NULL AND action_code IN ('restore_completed', 'restore_failed', 'restore_started', 'restore_verification_completed'))
    );

-- +goose Down
-- Retained Reference Pack audit events deliberately prevent downgrade.
ALTER TABLE public.administrative_audit_projections
    DROP CONSTRAINT administrative_audit_projections_action_scope_check,
    DROP CONSTRAINT administrative_audit_projections_target_check;
ALTER TABLE public.administrative_audit_projections
    ADD CONSTRAINT administrative_audit_projections_action_scope_check CHECK (
        (scope_kind='deployment' AND action_code IN ('account_preferences_updated', 'auth_binding_created', 'auth_binding_retired', 'auth_binding_rotated', 'backup_created', 'bootstrap_admin_created', 'deployment_admin_granted', 'deployment_admin_revoked', 'password_changed', 'password_reset', 'restore_completed', 'restore_failed', 'restore_started', 'restore_verification_completed', 'sessions_revoked', 'totp_enrollment_begun', 'totp_enrollment_completed', 'totp_reset', 'user_created', 'user_profile_updated', 'user_status_changed'))
        OR
        (scope_kind='incident' AND action_code IN ('membership_created', 'membership_deleted', 'membership_role_changed'))
    ),
    ADD CONSTRAINT administrative_audit_projections_target_check CHECK (
        (target_kind='account_preferences' AND target_id IS NOT NULL AND action_code IN ('account_preferences_updated'))
        OR
        (target_kind='auth_binding' AND target_id IS NOT NULL AND action_code IN ('auth_binding_created', 'auth_binding_retired', 'auth_binding_rotated'))
        OR
        (target_kind='backup_set' AND target_id IS NOT NULL AND action_code IN ('backup_created'))
        OR
        (target_kind='user' AND target_id IS NOT NULL AND action_code IN ('bootstrap_admin_created', 'deployment_admin_granted', 'deployment_admin_revoked', 'password_changed', 'password_reset', 'sessions_revoked', 'totp_enrollment_begun', 'totp_enrollment_completed', 'totp_reset', 'user_created', 'user_profile_updated', 'user_status_changed'))
        OR
        (target_kind='incident_membership' AND target_id IS NOT NULL AND action_code IN ('membership_created', 'membership_deleted', 'membership_role_changed'))
        OR
        (target_kind='restore_operation' AND target_id IS NOT NULL AND action_code IN ('restore_completed', 'restore_failed', 'restore_started', 'restore_verification_completed'))
    );
