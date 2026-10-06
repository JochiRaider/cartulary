-- +goose Up
-- Current journal payloads require this release; retained older deployments use
-- their matching release instead of an in-place compatibility conversion.
-- +goose StatementBegin
DO $$ BEGIN
    IF EXISTS (SELECT 1 FROM public.operator_recovery_journal) THEN
        RAISE EXCEPTION 'incompatible retained recovery journal';
    END IF;
END $$;
-- +goose StatementEnd
ALTER TABLE public.operator_recovery_journal
    DROP CONSTRAINT operator_recovery_journal_operation_check,
    ADD CONSTRAINT operator_recovery_journal_operation_check CHECK (operation IN (
        'backup_create', 'backup_export_latest', 'restore_latest', 'restore_bundle',
        'restore_verify_latest', 'restore_verify_due'
    ));

-- +goose Down
-- Retained transfer evidence deliberately prevents downgrade.
ALTER TABLE public.operator_recovery_journal
    DROP CONSTRAINT operator_recovery_journal_operation_check,
    ADD CONSTRAINT operator_recovery_journal_operation_check CHECK (operation IN (
        'backup_create', 'restore_latest', 'restore_verify_latest', 'restore_verify_due'
    ));
