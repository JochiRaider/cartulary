-- +goose Up
-- Fresh-state replacement. Never convert retained inline artifacts.
-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM public.reporting_imported_artifact_files) THEN
        RAISE EXCEPTION 'incident portability requires fresh retained reporting state';
    END IF;
END;
$$;
-- +goose StatementEnd

ALTER TABLE public.reporting_imported_artifact_files DROP CONSTRAINT reporting_imported_artifact_bytes_ck;
ALTER TABLE public.reporting_imported_artifact_files DROP COLUMN content;

CREATE TABLE public.reporting_imported_artifact_chunks (
    incident_id uuid NOT NULL,
    bundle_path text NOT NULL,
    chunk_ordinal bigint NOT NULL,
    content bytea NOT NULL,
    CONSTRAINT reporting_imported_artifact_chunks_pk PRIMARY KEY (incident_id, bundle_path, chunk_ordinal),
    CONSTRAINT reporting_imported_artifact_chunks_ordinal_ck CHECK (chunk_ordinal >= 0),
    CONSTRAINT reporting_imported_artifact_chunks_bytes_ck CHECK (octet_length(content) BETWEEN 1 AND 65536),
    CONSTRAINT reporting_imported_artifact_chunks_file_fk FOREIGN KEY (incident_id, bundle_path) REFERENCES public.reporting_imported_artifact_files(incident_id, bundle_path) ON UPDATE RESTRICT ON DELETE CASCADE
);
ALTER TABLE public.reporting_imported_artifact_chunks ALTER COLUMN content SET STORAGE EXTERNAL;
REVOKE ALL ON TABLE public.reporting_imported_artifact_chunks FROM PUBLIC;
REVOKE ALL ON TYPE public.reporting_imported_artifact_chunks FROM PUBLIC;
GRANT SELECT, INSERT ON TABLE public.reporting_imported_artifact_chunks TO cartulary_runtime;
GRANT SELECT, INSERT, UPDATE, DELETE, TRUNCATE ON TABLE public.reporting_imported_artifact_chunks TO cartulary_recovery;
CREATE TRIGGER reporting_imported_artifact_chunks_immutable BEFORE UPDATE ON public.reporting_imported_artifact_chunks
FOR EACH ROW EXECUTE FUNCTION public.reporting_reject_imported_artifact_update();

-- +goose Down
-- +goose StatementBegin
DO $$ BEGIN
    RAISE EXCEPTION 'incident portability requires fresh state and has no in-place downgrade; retain the prior executable with its own data';
END $$;
-- +goose StatementEnd
