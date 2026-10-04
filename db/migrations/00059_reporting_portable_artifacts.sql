-- +goose Up
-- Imported artifacts are immutable historical source evidence. They neither
-- create local snapshot/render Jobs nor authorize local release approvals.
CREATE TABLE public.reporting_imported_artifact_files (
    incident_id uuid NOT NULL CONSTRAINT reporting_imported_artifact_incident_fk REFERENCES public.incidents(id) ON UPDATE RESTRICT ON DELETE CASCADE,
    bundle_path text NOT NULL,
    operation_id uuid NOT NULL CONSTRAINT reporting_imported_artifact_operation_fk REFERENCES public.jobs(job_id) ON UPDATE RESTRICT ON DELETE RESTRICT,
    size_bytes bigint NOT NULL CONSTRAINT reporting_imported_artifact_size_ck CHECK (size_bytes >= 0),
    sha256 text NOT NULL CONSTRAINT reporting_imported_artifact_sha256_ck CHECK (sha256 ~ '^[0-9a-f]{64}$'),
    content bytea NOT NULL,
    CONSTRAINT reporting_imported_artifact_files_pk PRIMARY KEY (incident_id, bundle_path),
    CONSTRAINT reporting_imported_artifact_path_ck CHECK (
        bundle_path LIKE 'ext/snapshots/%' AND octet_length(bundle_path) <= 640
        AND bundle_path !~ '(^/|(^|/)\.\.?(/|$)|//)' AND position(chr(92) in bundle_path) = 0
    ),
    CONSTRAINT reporting_imported_artifact_bytes_ck CHECK (
        octet_length(content) = size_bytes AND encode(sha256(content), 'hex') = sha256
    )
);
CREATE INDEX reporting_imported_artifact_files_operation_id_fk_idx ON public.reporting_imported_artifact_files(operation_id);
-- +goose StatementBegin
CREATE FUNCTION public.reporting_reject_imported_artifact_update() RETURNS trigger
LANGUAGE plpgsql SET search_path = pg_catalog, public AS $$ BEGIN
    RAISE EXCEPTION 'reporting imported artifacts are immutable' USING ERRCODE='55000';
END $$;
-- +goose StatementEnd
CREATE TRIGGER reporting_imported_artifacts_immutable BEFORE UPDATE ON public.reporting_imported_artifact_files
FOR EACH ROW EXECUTE FUNCTION public.reporting_reject_imported_artifact_update();
GRANT SELECT, INSERT ON TABLE public.reporting_imported_artifact_files TO cartulary_runtime;
GRANT SELECT, INSERT, UPDATE, DELETE, TRUNCATE ON TABLE public.reporting_imported_artifact_files TO cartulary_recovery;
REVOKE USAGE ON TYPE public.reporting_imported_artifact_files FROM PUBLIC;
GRANT USAGE ON TYPE public.reporting_imported_artifact_files TO cartulary_runtime,cartulary_recovery;
REVOKE ALL ON FUNCTION public.reporting_reject_imported_artifact_update() FROM PUBLIC;
GRANT EXECUTE ON FUNCTION public.reporting_reject_imported_artifact_update() TO cartulary_runtime,cartulary_recovery;

-- +goose Down
-- +goose StatementBegin
DO $$ BEGIN RAISE EXCEPTION 'retained reporting artifacts cannot be discarded by downgrade; restore a complete matching backup' USING ERRCODE='55000'; END $$;
-- +goose StatementEnd
