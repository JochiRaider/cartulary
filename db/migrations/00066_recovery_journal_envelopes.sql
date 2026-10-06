-- +goose Up
-- The current release admits fresh deployments, never relabels historical journals.
-- +goose StatementBegin
DO $$ BEGIN
    IF EXISTS (SELECT 1 FROM public.operator_recovery_journal) THEN
        RAISE EXCEPTION 'incompatible retained recovery journal';
    END IF;
END $$;
-- +goose StatementEnd
ALTER TABLE public.operator_recovery_journal
    DROP COLUMN nonce, DROP COLUMN ciphertext,
    DROP CONSTRAINT operator_recovery_journal_envelope_schema_id_check,
    ADD COLUMN sealed_payload bytea NOT NULL,
    ADD CONSTRAINT operator_recovery_journal_envelope_schema_id_check CHECK (
        envelope_schema_id = 'cartulary.operator_recovery_journal_envelope.v2'),
    ADD CONSTRAINT operator_recovery_journal_sealed_payload_check CHECK (
        octet_length(sealed_payload) BETWEEN 62 AND 16777277 AND get_byte(sealed_payload, 0) = 1);

-- +goose Down
-- +goose StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'recovery journal envelope migration is irreversible'; END $$;
-- +goose StatementEnd
