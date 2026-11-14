-- +goose Up
ALTER TABLE public.evidence_grid_projection ADD COLUMN owner_user_id uuid;

-- +goose Down
ALTER TABLE public.evidence_grid_projection DROP COLUMN owner_user_id;
