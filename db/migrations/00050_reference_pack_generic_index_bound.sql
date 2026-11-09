-- +goose Up
-- Admission owns canonical JSON and profile-specific row limits. Storage owns
-- only the maximum admitted row bytes, excluding the mandatory NDJSON LF.
-- Inert object data must never select a different database constraint.
ALTER TABLE public.reference_pack_indexes DROP CONSTRAINT rp_indexes_ck3;
ALTER TABLE public.reference_pack_indexes ADD CONSTRAINT rp_indexes_ck3 CHECK
    (octet_length(canonical_item) >= 1 AND octet_length(canonical_item) <= 4194303);

-- +goose Down
-- A predecessor application must use its preserved compatible environment.
-- +goose StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'reference pack generic contract downgrade is unsupported'; END $$;
-- +goose StatementEnd
