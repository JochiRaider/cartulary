-- +goose Up
UPDATE public.timeline_grid_projection
   SET host_refs = COALESCE(host_refs, '[]'::jsonb),
       identity_refs = COALESCE(identity_refs, '[]'::jsonb),
       tags = COALESCE(tags, '[]'::jsonb),
       attached_evidence_refs = COALESCE(attached_evidence_refs, '[]'::jsonb);

ALTER TABLE public.timeline_grid_projection
    ALTER COLUMN host_refs SET DEFAULT '[]'::jsonb,
    ALTER COLUMN host_refs SET NOT NULL,
    ALTER COLUMN identity_refs SET DEFAULT '[]'::jsonb,
    ALTER COLUMN identity_refs SET NOT NULL,
    ALTER COLUMN tags SET DEFAULT '[]'::jsonb,
    ALTER COLUMN tags SET NOT NULL,
    ALTER COLUMN attached_evidence_refs SET DEFAULT '[]'::jsonb,
    ALTER COLUMN attached_evidence_refs SET NOT NULL,
    ADD CONSTRAINT timeline_grid_projection_host_refs_array_ck
        CHECK (jsonb_typeof(host_refs) = 'array'),
    ADD CONSTRAINT timeline_grid_projection_identity_refs_array_ck
        CHECK (jsonb_typeof(identity_refs) = 'array'),
    ADD CONSTRAINT timeline_grid_projection_tags_array_ck
        CHECK (jsonb_typeof(tags) = 'array'),
    ADD CONSTRAINT timeline_grid_projection_attached_evidence_refs_array_ck
        CHECK (jsonb_typeof(attached_evidence_refs) = 'array');

-- +goose Down
ALTER TABLE public.timeline_grid_projection
    DROP CONSTRAINT timeline_grid_projection_attached_evidence_refs_array_ck,
    DROP CONSTRAINT timeline_grid_projection_tags_array_ck,
    DROP CONSTRAINT timeline_grid_projection_identity_refs_array_ck,
    DROP CONSTRAINT timeline_grid_projection_host_refs_array_ck,
    ALTER COLUMN attached_evidence_refs DROP NOT NULL,
    ALTER COLUMN attached_evidence_refs DROP DEFAULT,
    ALTER COLUMN tags DROP NOT NULL,
    ALTER COLUMN tags DROP DEFAULT,
    ALTER COLUMN identity_refs DROP NOT NULL,
    ALTER COLUMN identity_refs DROP DEFAULT,
    ALTER COLUMN host_refs DROP NOT NULL,
    ALTER COLUMN host_refs DROP DEFAULT;
