-- +goose Up
-- Earlier verification did not authenticate dependency availability/graphs.
-- Never retrofit that evidence or invent frozen reads for queued operations.
-- +goose StatementBegin
DO $$ DECLARE pending bigint; dependent bigint; BEGIN
    LOCK TABLE public.reference_pack_operations,public.reference_pack_versions IN ACCESS EXCLUSIVE MODE;
    SELECT count(*) INTO pending FROM public.reference_pack_operations
        WHERE kind IN ('import','reverify','refresh') AND terminal_at IS NULL;
    SELECT count(*) INTO dependent FROM public.reference_pack_versions
        WHERE jsonb_array_length(convert_from(manifest_bytes,'UTF8')::jsonb->'dependencies') > 0;
    IF pending > 0 OR dependent > 0 THEN
        RAISE EXCEPTION 'reference_pack_cutover_required: pending_dependency_captures=%, unproven_dependency_history=%; preserve a matching backup and explicitly reset disposable development data', pending,dependent USING ERRCODE='55000';
    END IF;
END $$;
-- +goose StatementEnd

CREATE TABLE public.reference_pack_operation_dependency_keys (
    operation_id uuid NOT NULL CONSTRAINT rp_operation_dependency_keys_operation_fk REFERENCES public.reference_pack_operations(operation_id) ON UPDATE RESTRICT ON DELETE RESTRICT,
    pack_key text NOT NULL CONSTRAINT rp_operation_dependency_keys_key_fk REFERENCES public.reference_pack_key_state(pack_key) ON UPDATE RESTRICT ON DELETE RESTRICT,
    admitted_revision bigint NOT NULL CONSTRAINT rp_operation_dependency_keys_revision_ck CHECK (admitted_revision > 0),
    CONSTRAINT rp_operation_dependency_keys_pk PRIMARY KEY (operation_id,pack_key)
);
CREATE INDEX reference_pack_operation_dependency_keys_key_idx ON public.reference_pack_operation_dependency_keys(pack_key);
CREATE TRIGGER reference_pack_operation_dependency_keys_immutable BEFORE UPDATE OR DELETE ON public.reference_pack_operation_dependency_keys FOR EACH ROW EXECUTE FUNCTION public.reference_pack_immutable_row();
GRANT SELECT, INSERT ON TABLE public.reference_pack_operation_dependency_keys TO cartulary_runtime;
GRANT SELECT, INSERT, UPDATE, DELETE, TRUNCATE ON TABLE public.reference_pack_operation_dependency_keys TO cartulary_recovery;
REVOKE USAGE ON TYPE public.reference_pack_operation_dependency_keys FROM PUBLIC;
GRANT USAGE ON TYPE public.reference_pack_operation_dependency_keys TO cartulary_runtime,cartulary_recovery;

-- +goose Down
-- +goose StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'reference pack dependency evidence cannot be discarded; restore a complete matching backup' USING ERRCODE='55000'; END $$;
-- +goose StatementEnd
