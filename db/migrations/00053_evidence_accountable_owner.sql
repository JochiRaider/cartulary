-- +goose Up
ALTER TABLE public.evidence ADD COLUMN owner_user_id uuid;
ALTER TABLE public.evidence ADD CONSTRAINT evidence_owner_user_id_fkey
    FOREIGN KEY (owner_user_id) REFERENCES public.users(id) ON UPDATE NO ACTION ON DELETE NO ACTION;
CREATE INDEX evidence_owner_user_id_fk_idx ON public.evidence (owner_user_id);

-- +goose Down
-- Retiring authoritative assignments requires an explicit migration policy.
-- +goose StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'evidence owner downgrade is unsupported'; END $$;
-- +goose StatementEnd
