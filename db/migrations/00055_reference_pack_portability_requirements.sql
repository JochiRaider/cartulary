-- +goose Up
-- Previously retained source catalogs did not preserve embedded-content or
-- required-member declarations. Never fabricate that historical classification.
-- +goose StatementBegin
DO $$ DECLARE retained bigint; BEGIN
    LOCK TABLE public.reference_pack_operations IN ACCESS EXCLUSIVE MODE;
    SELECT count(*) INTO retained FROM public.reference_pack_operations WHERE kind='portable_retention';
    IF retained > 0 THEN
        RAISE EXCEPTION 'reference_pack_cutover_required: unclassified_portable_content=%; preserve a matching backup and explicitly reset disposable development data', retained USING ERRCODE='55000';
    END IF;
END $$;
-- +goose StatementEnd

ALTER TABLE public.reference_pack_operations ADD CONSTRAINT rp_portable_content_input_ck CHECK (
    kind <> 'portable_retention' OR coalesce(
        convert_from(frozen_input,'UTF8')::jsonb->'content_manifest'->>'schema_id'='reference_pack_content.v1'
        AND jsonb_typeof(convert_from(frozen_input,'UTF8')::jsonb->'content_manifest'->'containers')='array'
        AND jsonb_typeof(convert_from(frozen_input,'UTF8')::jsonb->'content_manifest'->'required_members')='array', false)
);

-- +goose Down
-- +goose StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'reference pack required-member evidence cannot be discarded; restore a complete matching backup' USING ERRCODE='55000'; END $$;
-- +goose StatementEnd
