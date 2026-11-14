-- +goose Up
-- Non-secret imported attribution hints are authoritative portable actor data.
-- They do not create credentials, account bindings, or incident memberships.
ALTER TABLE public.incident_bundle_imported_actors
    ADD COLUMN provider_subject_hint text;

-- +goose Down
-- +goose StatementBegin
DO $$ BEGIN
    RAISE EXCEPTION 'incident portability requires fresh state and has no in-place downgrade; retain the prior executable with its own data';
END $$;
-- +goose StatementEnd
