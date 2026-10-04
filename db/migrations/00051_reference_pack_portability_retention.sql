-- +goose Up
-- A prior format-5 import may have admitted references without retaining
-- them. Neither empty references nor historical pack bindings can be inferred.
-- +goose StatementBegin
DO $$
DECLARE affected bigint;
BEGIN
    LOCK TABLE public.incident_bundle_imported_actors,public.incident_bundle_imported_attributions,public.incident_bundle_job_payloads IN SHARE ROW EXCLUSIVE MODE;
    SELECT count(*) INTO affected FROM (
        SELECT incident_id FROM public.incident_bundle_imported_actors
        UNION SELECT incident_id FROM public.incident_bundle_imported_attributions
        UNION SELECT imported_incident_id FROM public.incident_bundle_job_payloads WHERE imported_incident_id IS NOT NULL
    ) imported;
    IF affected>0 THEN
        RAISE EXCEPTION 'reference_pack_cutover_required: historical_incident_reference_bindings=%; preserve a backup/export with the previous application, then use the explicit development reset procedure',affected USING ERRCODE='55000';
    END IF;
END $$;
-- +goose StatementEnd
ALTER TABLE public.reference_pack_operations DROP CONSTRAINT rp_operations_ck1;
ALTER TABLE public.reference_pack_operations ADD CONSTRAINT rp_operations_ck1 CHECK (kind IN ('import','renewal','reverify','refresh','activate','disable','remove','integrity','reconcile','portable_retention'));

-- Source catalogs are inert historical evidence. They can name unavailable
-- optional versions and therefore must not reference the trusted-version table.
CREATE TABLE public.reference_pack_portable_catalogs (
    incident_id uuid CONSTRAINT rp_portable_catalogs_pk PRIMARY KEY CONSTRAINT rp_portable_catalogs_incident_fk REFERENCES public.incidents(id) ON UPDATE RESTRICT ON DELETE RESTRICT,
    operation_id uuid NOT NULL CONSTRAINT rp_portable_catalogs_operation_uq UNIQUE CONSTRAINT rp_portable_catalogs_operation_fk REFERENCES public.reference_pack_operations(operation_id) ON UPDATE RESTRICT ON DELETE RESTRICT,
    catalog_sha256 text NOT NULL CONSTRAINT rp_portable_catalogs_digest_ck CHECK (catalog_sha256 ~ '^[0-9a-f]{64}$'),
    canonical_references bytea NOT NULL,
    canonical_resolution bytea NOT NULL,
    CONSTRAINT rp_portable_catalogs_bytes_ck CHECK (encode(public.digest(canonical_references,'sha256'),'hex')=catalog_sha256)
);
CREATE TRIGGER reference_pack_portable_catalogs_immutable BEFORE UPDATE OR DELETE ON public.reference_pack_portable_catalogs FOR EACH ROW EXECUTE FUNCTION public.reference_pack_immutable_row();

-- A partially available set still retains every reusable member. This pin
-- records destination evidence and never imports a source trust envelope.
CREATE TABLE public.reference_pack_version_pins (
    owner_kind text NOT NULL,
    owner_id text NOT NULL,
    pack_key text NOT NULL,
    pack_version text NOT NULL,
    envelope_id text NOT NULL,
    operation_id uuid NOT NULL CONSTRAINT rp_version_pins_operation_fk REFERENCES public.reference_pack_operations(operation_id) ON UPDATE RESTRICT ON DELETE RESTRICT,
    CONSTRAINT rp_version_pins_pk PRIMARY KEY(owner_kind,owner_id,pack_key,pack_version),
    CONSTRAINT rp_version_pins_envelope_fk FOREIGN KEY(envelope_id,pack_key,pack_version) REFERENCES public.reference_pack_envelopes(envelope_id,pack_key,pack_version) ON UPDATE RESTRICT ON DELETE RESTRICT
);
CREATE INDEX rp_version_pins_version_idx ON public.reference_pack_version_pins(pack_key,pack_version);
CREATE INDEX rp_version_pins_envelope_idx ON public.reference_pack_version_pins(envelope_id,pack_key,pack_version);
CREATE INDEX rp_version_pins_operation_idx ON public.reference_pack_version_pins(operation_id);

GRANT SELECT,INSERT ON TABLE public.reference_pack_portable_catalogs TO cartulary_runtime;
GRANT SELECT,INSERT,UPDATE,DELETE ON TABLE public.reference_pack_version_pins TO cartulary_runtime;
GRANT SELECT,INSERT,UPDATE,DELETE,TRUNCATE ON TABLE public.reference_pack_portable_catalogs,public.reference_pack_version_pins TO cartulary_recovery;
REVOKE ALL ON TYPE public.reference_pack_portable_catalogs,public.reference_pack_version_pins FROM PUBLIC;
GRANT USAGE ON TYPE public.reference_pack_portable_catalogs,public.reference_pack_version_pins TO cartulary_runtime,cartulary_recovery;

-- +goose Down
-- +goose StatementBegin
DO $$ BEGIN
    IF EXISTS(SELECT 1 FROM public.reference_pack_portable_catalogs) OR EXISTS(SELECT 1 FROM public.reference_pack_version_pins) OR EXISTS(SELECT 1 FROM public.reference_pack_operations WHERE kind='portable_retention') THEN
        RAISE EXCEPTION 'reference pack portability history cannot be discarded; restore a complete matching backup' USING ERRCODE='55000';
    END IF;
END $$;
-- +goose StatementEnd
DROP TABLE public.reference_pack_version_pins,public.reference_pack_portable_catalogs;
ALTER TABLE public.reference_pack_operations DROP CONSTRAINT rp_operations_ck1;
ALTER TABLE public.reference_pack_operations ADD CONSTRAINT rp_operations_ck1 CHECK (kind IN ('import','renewal','reverify','refresh','activate','disable','remove','integrity','reconcile'));
