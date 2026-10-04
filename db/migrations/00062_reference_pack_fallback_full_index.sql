-- +goose Up
-- The adopted DDL profile requires a full leading-column index for every FK.
DROP INDEX public.rp_candidate_fallback_version_idx;
CREATE INDEX rp_candidate_fallback_version_idx ON public.reference_pack_candidates(pack_key,fallback_from_version);

-- +goose Down
DROP INDEX public.rp_candidate_fallback_version_idx;
CREATE INDEX rp_candidate_fallback_version_idx ON public.reference_pack_candidates(pack_key,fallback_from_version) WHERE fallback_from_version IS NOT NULL;
