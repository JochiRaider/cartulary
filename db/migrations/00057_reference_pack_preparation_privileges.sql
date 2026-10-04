-- +goose Up
-- Private preparation follows the same roles as the immutable owner inventory.
-- Runtime can append and read, but cannot rewrite a captured admission.
GRANT SELECT, INSERT ON TABLE public.reference_pack_portable_preparations, public.reference_pack_portable_selections TO cartulary_runtime;
GRANT SELECT, INSERT, UPDATE, DELETE, TRUNCATE ON TABLE public.reference_pack_portable_preparations, public.reference_pack_portable_selections TO cartulary_recovery;
REVOKE USAGE ON TYPE public.reference_pack_portable_preparations, public.reference_pack_portable_selections FROM PUBLIC;
GRANT USAGE ON TYPE public.reference_pack_portable_preparations, public.reference_pack_portable_selections TO cartulary_runtime,cartulary_recovery;

-- +goose Down
-- +goose StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'reference pack preparation permissions cannot be retired independently of retained state' USING ERRCODE='55000'; END $$;
-- +goose StatementEnd
