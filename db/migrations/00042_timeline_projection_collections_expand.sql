-- +goose Up
ALTER TABLE public.timeline_grid_projection
    ADD COLUMN host_refs jsonb,
    ADD COLUMN identity_refs jsonb,
    ADD COLUMN tags jsonb,
    ADD COLUMN attached_evidence_refs jsonb;

-- +goose Down
ALTER TABLE public.timeline_grid_projection
    DROP COLUMN attached_evidence_refs,
    DROP COLUMN tags,
    DROP COLUMN identity_refs,
    DROP COLUMN host_refs;
