-- +goose Up
-- Current selection attribution cannot be reconstructed from wall-clock event
-- order. Refuse intermediate development state rather than guess its cause.
-- +goose StatementBegin
DO $$ DECLARE affected bigint; BEGIN
    LOCK TABLE public.reference_pack_events IN SHARE MODE;
    SELECT count(*) INTO affected FROM public.reference_pack_events WHERE event_kind='safety_fallback';
    IF affected > 0 THEN
        RAISE EXCEPTION 'reference_pack_cutover_required: unclassified_safety_fallbacks=%; preserve a matching backup, then use the explicit development reset procedure', affected USING ERRCODE='55000';
    END IF;
END $$;
-- +goose StatementEnd
ALTER TABLE public.reference_pack_candidates ADD COLUMN fallback_from_version text;
ALTER TABLE public.reference_pack_candidates ADD CONSTRAINT rp_candidate_fallback_kind_ck
    CHECK (fallback_from_version IS NULL OR (distribution_kind='packaged_builtin' AND fallback_from_version<>pack_version));
ALTER TABLE public.reference_pack_candidates ADD CONSTRAINT rp_candidate_fallback_version_fk
    FOREIGN KEY (pack_key,fallback_from_version) REFERENCES public.reference_pack_candidates(pack_key,pack_version) ON UPDATE RESTRICT ON DELETE RESTRICT;
CREATE INDEX rp_candidate_fallback_version_idx ON public.reference_pack_candidates(pack_key,fallback_from_version) WHERE fallback_from_version IS NOT NULL;

-- +goose Down
-- +goose StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'current fallback attribution cannot be discarded by downgrade; restore a complete matching backup' USING ERRCODE='55000'; END $$;
-- +goose StatementEnd
