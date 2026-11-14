-- +goose Up
ALTER TABLE public.indicators DROP CONSTRAINT indicators_required_text_ck;
ALTER TABLE public.indicators ADD CONSTRAINT indicators_required_text_ck CHECK (
    octet_length(display_value) > 0 AND octet_length(dedupe_key) > 0
    AND octet_length(normalized_value) > 0
    AND (hash_algorithm IS NULL OR octet_length(hash_algorithm) > 0)
    AND (hash_value IS NULL OR octet_length(hash_value) > 0)
);

-- +goose Down
-- Fail transactionally if admitted empty presentation values prevent downgrade.
-- Retained source facts must never be rewritten to satisfy an older constraint.
ALTER TABLE public.indicators DROP CONSTRAINT indicators_required_text_ck;
ALTER TABLE public.indicators ADD CONSTRAINT indicators_required_text_ck CHECK (
    octet_length(display_value) > 0 AND octet_length(dedupe_key) > 0
    AND (normalized_value IS NULL OR octet_length(normalized_value) > 0)
    AND (defanged_value IS NULL OR octet_length(defanged_value) > 0)
    AND (hash_algorithm IS NULL OR octet_length(hash_algorithm) > 0)
    AND (hash_value IS NULL OR octet_length(hash_value) > 0)
    AND (stix_pattern IS NULL OR octet_length(stix_pattern) > 0)
);
