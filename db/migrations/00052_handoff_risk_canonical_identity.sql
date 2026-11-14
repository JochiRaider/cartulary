-- +goose Up
-- The title contract preserves case. Its canonical text is the identity within
-- one parent; a second case-folded value is neither source state nor identity.
DROP INDEX public.handoff_risk_refs_active_text_idx;
ALTER TABLE public.handoff_risk_refs DROP COLUMN normalized_risk_ref_text;
CREATE UNIQUE INDEX handoff_risk_refs_active_text_idx
    ON public.handoff_risk_refs (handoff_record_id, risk_ref_text)
    WHERE deleted_at IS NULL;

-- +goose Down
-- Case-distinct references cannot be collapsed without discarding source state.
-- +goose StatementBegin
DO $$ BEGIN
    RAISE EXCEPTION 'canonical risk identity has no in-place downgrade; retain the prior executable with its own data';
END $$;
-- +goose StatementEnd
