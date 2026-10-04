-- +goose Up
-- Retired stores are not trusted history. Refuse retained rows under locks;
-- never infer canonical envelopes, discard audit evidence, or convert jobs.
-- +goose StatementBegin
DO $$
DECLARE
    packs bigint;
    payloads bigint;
    attestations bigint;
    activations bigint;
BEGIN
    LOCK TABLE public.reference_packs, public.reference_pack_job_payloads,
        public.reference_pack_attestations, public.reference_pack_activation_state
        IN ACCESS EXCLUSIVE MODE;
    SELECT count(*) INTO packs FROM public.reference_packs;
    SELECT count(*) INTO payloads FROM public.reference_pack_job_payloads;
    SELECT count(*) INTO attestations FROM public.reference_pack_attestations;
    SELECT count(*) INTO activations FROM public.reference_pack_activation_state;
    IF packs + payloads + attestations + activations > 0 THEN
        RAISE EXCEPTION 'reference_pack_cutover_required'
            USING ERRCODE = '55000',
            DETAIL = format('legacy_packs=%s legacy_pack_jobs=%s legacy_pack_attestations=%s legacy_pack_activations=%s', packs, payloads, attestations, activations),
            HINT = 'Preserve a full backup and exports with the matching historical application. Reset only disposable development data through make db-reset CARTULARY_DESTRUCTIVE_CONFIRM=db-reset; reset does not convert historical backups.';
    END IF;
END $$;
-- +goose StatementEnd
DROP TABLE public.reference_pack_job_payloads;
DROP TABLE public.reference_pack_attestations;
DROP TABLE public.reference_pack_activation_state;
DROP TABLE public.reference_packs;

-- +goose Down
-- A downgrade cannot manufacture the retired trust and lifecycle representation.
-- +goose StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'reference pack legacy tables cannot be restored by downgrade; restore a complete matching backup' USING ERRCODE = '55000'; END $$;
-- +goose StatementEnd
