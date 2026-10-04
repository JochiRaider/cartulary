-- +goose Up
-- A portable verification cohort is private preparation owned by the parent
-- incident Job. Reuse-only historical catalogs need no new preparation record.
-- This additive migration never invents trust or rewrites retained catalogs.
CREATE TABLE public.reference_pack_portable_preparations (
    operation_id uuid CONSTRAINT rp_portable_preparations_pk PRIMARY KEY CONSTRAINT rp_portable_preparations_operation_fk REFERENCES public.reference_pack_operations(operation_id) ON UPDATE RESTRICT ON DELETE RESTRICT,
    canonical_references bytea NOT NULL CONSTRAINT rp_portable_preparations_references_ck CHECK (octet_length(canonical_references) BETWEEN 1 AND 67108864),
    canonical_context bytea NOT NULL CONSTRAINT rp_portable_preparations_context_ck CHECK (octet_length(canonical_context) BETWEEN 1 AND 4096)
);
ALTER TABLE public.reference_pack_operation_members ADD CONSTRAINT rp_operation_members_portable_binding_uq UNIQUE (operation_id,ordinal,pack_key,pack_version);
CREATE TABLE public.reference_pack_portable_selections (
    operation_id uuid NOT NULL CONSTRAINT rp_portable_selections_operation_fk REFERENCES public.reference_pack_portable_preparations(operation_id) ON UPDATE RESTRICT ON DELETE RESTRICT,
    ordinal integer NOT NULL CONSTRAINT rp_portable_selections_ordinal_ck CHECK (ordinal BETWEEN 1 AND 16384),
    pack_key text NOT NULL,
    pack_version text NOT NULL,
    envelope_id text,
    available boolean NOT NULL,
    reason_code text CONSTRAINT rp_portable_selections_reason_ck CHECK (reason_code IN ('not_retained','removed','not_usable','content_unavailable')),
    verification_ordinal bigint,
    input_object_id uuid CONSTRAINT rp_portable_selections_object_fk REFERENCES public.reference_pack_objects(object_id) ON UPDATE RESTRICT ON DELETE RESTRICT,
    expected_container_sha256 text CONSTRAINT rp_portable_selections_digest_ck CHECK (expected_container_sha256 ~ '^[0-9a-f]{64}$'),
    expected_container_bytes bigint CONSTRAINT rp_portable_selections_size_ck CHECK (expected_container_bytes BETWEEN 1 AND 9007199254740991),
    CONSTRAINT rp_portable_selections_pk PRIMARY KEY (operation_id,ordinal),
    CONSTRAINT rp_portable_selections_version_uq UNIQUE (operation_id,pack_key,pack_version),
    CONSTRAINT rp_portable_selections_verification_uq UNIQUE (operation_id,verification_ordinal),
    CONSTRAINT rp_portable_selections_envelope_fk FOREIGN KEY (envelope_id,pack_key,pack_version) REFERENCES public.reference_pack_envelopes(envelope_id,pack_key,pack_version) ON UPDATE RESTRICT ON DELETE RESTRICT,
    CONSTRAINT rp_portable_selections_member_fk FOREIGN KEY (operation_id,verification_ordinal,pack_key,pack_version) REFERENCES public.reference_pack_operation_members(operation_id,ordinal,pack_key,pack_version) ON UPDATE RESTRICT ON DELETE RESTRICT,
    CONSTRAINT rp_portable_selections_availability_ck CHECK (available = (reason_code IS NULL) AND (NOT available OR envelope_id IS NOT NULL)),
    CONSTRAINT rp_portable_selections_input_ck CHECK (
        (verification_ordinal IS NULL AND input_object_id IS NULL AND expected_container_sha256 IS NULL AND expected_container_bytes IS NULL)
        OR (verification_ordinal IS NOT NULL AND input_object_id IS NOT NULL AND expected_container_sha256 IS NOT NULL AND expected_container_bytes IS NOT NULL AND NOT available)
    )
);
CREATE INDEX rp_portable_selections_member_idx ON public.reference_pack_portable_selections(operation_id,verification_ordinal,pack_key,pack_version);
CREATE INDEX rp_portable_selections_envelope_idx ON public.reference_pack_portable_selections(envelope_id,pack_key,pack_version);
CREATE INDEX rp_portable_selections_object_idx ON public.reference_pack_portable_selections(input_object_id);
CREATE TRIGGER reference_pack_portable_preparations_immutable BEFORE UPDATE OR DELETE ON public.reference_pack_portable_preparations FOR EACH ROW EXECUTE FUNCTION public.reference_pack_immutable_row();
CREATE TRIGGER reference_pack_portable_selections_immutable BEFORE UPDATE OR DELETE ON public.reference_pack_portable_selections FOR EACH ROW EXECUTE FUNCTION public.reference_pack_immutable_row();

-- +goose Down
-- +goose StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'reference pack portable preparation cannot be discarded; restore a complete matching backup' USING ERRCODE='55000'; END $$;
-- +goose StatementEnd
