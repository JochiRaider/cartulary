-- +goose Up
-- Revisions owns immutable provenance of the actor namespace in each capture.
-- Existing pre-production imported history requires recreation; no inferred
-- context or actor rewrite can recover a missing historical import boundary.
-- +goose StatementBegin
DO $$ BEGIN
    IF EXISTS (SELECT 1 FROM public.change_sets) THEN
        RAISE EXCEPTION 'revision import provenance requires fresh history state';
    END IF;
END $$;
-- +goose StatementEnd

ALTER TABLE public.change_sets
    ADD COLUMN actor_import_contexts jsonb NOT NULL DEFAULT '[]'::jsonb,
    ADD CONSTRAINT change_sets_actor_import_contexts_array_ck
        CHECK (jsonb_typeof(actor_import_contexts) = 'array');

-- +goose Down
-- +goose StatementBegin
DO $$ BEGIN
    RAISE EXCEPTION 'revision import provenance requires fresh state and has no in-place downgrade';
END $$;
-- +goose StatementEnd
