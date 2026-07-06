-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
          FROM reporting_releases
         WHERE output_kind NOT IN ('slidev', 'mermaid')
         LIMIT 1
    ) THEN
        RAISE EXCEPTION 'reporting current output-kind migration cannot run with legacy reporting output_kind rows; reset or export reporting data before applying this migration';
    END IF;
    IF EXISTS (
        SELECT 1
          FROM reporting_job_payloads
         WHERE job_kind = 'release_create'
           AND request_json->>'output_kind' NOT IN ('slidev', 'mermaid')
         LIMIT 1
    ) THEN
        RAISE EXCEPTION 'reporting current output-kind migration cannot run with legacy reporting release job payloads; reset or export reporting data before applying this migration';
    END IF;
END $$;
-- +goose StatementEnd

ALTER TABLE reporting_releases
    DROP CONSTRAINT IF EXISTS reporting_releases_output_kind_check;

ALTER TABLE reporting_releases
    ADD CONSTRAINT reporting_releases_output_kind_check CHECK (output_kind IN ('slidev', 'mermaid'));

-- +goose Down
ALTER TABLE reporting_releases
    DROP CONSTRAINT IF EXISTS reporting_releases_output_kind_check;

ALTER TABLE reporting_releases
    ADD CONSTRAINT reporting_releases_output_kind_check CHECK (output_kind IN ('html', 'markdown', 'slidev', 'mermaid', 'reenactment'));
