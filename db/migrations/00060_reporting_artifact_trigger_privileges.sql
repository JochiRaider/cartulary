-- +goose Up
-- PostgreSQL checks trigger EXECUTE at trigger creation, not at invocation.
-- The migration role owns this private routine. Application and recovery
-- writes still run the immutable trigger without a direct invocation grant.
REVOKE ALL ON FUNCTION public.reporting_reject_imported_artifact_update() FROM cartulary_runtime, cartulary_recovery;

-- +goose Down
-- +goose StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'private artifact trigger privileges cannot be widened by downgrade; restore a complete matching backup' USING ERRCODE='55000'; END $$;
-- +goose StatementEnd
