-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM report_composition_versions LIMIT 1) THEN
        RAISE EXCEPTION 'report composition canonical-byte migration cannot run with existing composition versions; export or reset draft-extension composition data before applying this migration';
    END IF;
    IF EXISTS (SELECT 1 FROM report_composition_preview_attempts LIMIT 1) THEN
        RAISE EXCEPTION 'report composition preview-attempt migration cannot run with existing preview attempts; reset draft-extension preview data before applying this migration';
    END IF;
    IF EXISTS (
        SELECT 1
          FROM graph_projection_runs
         WHERE state IN ('available', 'replaced')
           AND graph_view_json IS NOT NULL
         LIMIT 1
    ) THEN
        RAISE EXCEPTION 'graph projection output-digest migration cannot backfill existing completed runs safely; reset or export graph projection data before applying this migration';
    END IF;
END $$;
-- +goose StatementEnd

ALTER TABLE report_composition_versions
    ADD COLUMN IF NOT EXISTS canonical_composition_bytes bytea;

ALTER TABLE report_composition_versions
    ALTER COLUMN canonical_composition_bytes SET NOT NULL;

ALTER TABLE report_composition_versions
    DROP CONSTRAINT IF EXISTS report_composition_versions_canonical_bytes_ck;

ALTER TABLE report_composition_versions
    ADD CONSTRAINT report_composition_versions_canonical_bytes_ck CHECK (octet_length(canonical_composition_bytes) > 0);

ALTER TABLE report_composition_preview_attempts
    ADD COLUMN IF NOT EXISTS preview_source_bytes bytea;

ALTER TABLE report_composition_preview_attempts
    ALTER COLUMN preview_source_bytes SET NOT NULL;

ALTER TABLE report_composition_preview_attempts
    ALTER COLUMN render_attempt_id SET NOT NULL;

ALTER TABLE report_composition_preview_attempts
    DROP CONSTRAINT IF EXISTS report_composition_preview_attempts_render_attempt_fk;

ALTER TABLE report_composition_preview_attempts
    ADD CONSTRAINT report_composition_preview_attempts_render_attempt_fk
    FOREIGN KEY (render_attempt_id) REFERENCES jobs (job_id) ON DELETE RESTRICT;

ALTER TABLE report_composition_release_bindings
    DROP CONSTRAINT IF EXISTS report_composition_release_bindings_release_fk;

ALTER TABLE report_composition_release_bindings
    ADD CONSTRAINT report_composition_release_bindings_release_fk
    FOREIGN KEY (release_id) REFERENCES reporting_releases (release_id) ON DELETE RESTRICT;

ALTER TABLE graph_projection_runs
    ADD COLUMN IF NOT EXISTS projection_output_digest text;

ALTER TABLE graph_projection_runs
    DROP CONSTRAINT IF EXISTS graph_projection_runs_projection_output_digest_ck;

ALTER TABLE graph_projection_runs
    ADD CONSTRAINT graph_projection_runs_projection_output_digest_ck CHECK (
        (
            state IN ('available', 'replaced')
            AND projection_output_digest ~ '^[a-f0-9]{64}$'
        )
        OR
        (
            state NOT IN ('available', 'replaced')
            AND (projection_output_digest IS NULL OR projection_output_digest ~ '^[a-f0-9]{64}$')
        )
    );

ALTER TABLE reporting_releases
    ADD COLUMN IF NOT EXISTS output_options jsonb NOT NULL DEFAULT '{}'::jsonb,
    ADD COLUMN IF NOT EXISTS graph_projection_refs jsonb NOT NULL DEFAULT '[]'::jsonb,
    ADD COLUMN IF NOT EXISTS composition_id uuid,
    ADD COLUMN IF NOT EXISTS composition_version text,
    ADD COLUMN IF NOT EXISTS composition_sha256 text,
    ADD COLUMN IF NOT EXISTS render_admitted_at timestamptz;

UPDATE reporting_releases
   SET render_admitted_at = created_at
 WHERE render_admitted_at IS NULL;

UPDATE reporting_releases
   SET output_options = jsonb_build_object(
       'schema_id', 'cartulary.reporting_render_request_options.v1',
       'source_only', false,
       'pdf', output_kind = 'slidev',
       'svg', output_kind = 'mermaid',
       'png', false,
       'pptx', false,
       'rendered_diagrams', true
   )
 WHERE output_options = '{}'::jsonb
    OR output_options IS NULL;

ALTER TABLE reporting_releases
    ALTER COLUMN render_admitted_at SET NOT NULL;

ALTER TABLE reporting_releases
    DROP CONSTRAINT IF EXISTS reporting_releases_output_options_ck,
    DROP CONSTRAINT IF EXISTS reporting_releases_graph_projection_refs_ck,
    DROP CONSTRAINT IF EXISTS reporting_releases_composition_tuple_ck;

ALTER TABLE reporting_releases
    ADD CONSTRAINT reporting_releases_output_options_ck CHECK (jsonb_typeof(output_options) = 'object'),
    ADD CONSTRAINT reporting_releases_graph_projection_refs_ck CHECK (jsonb_typeof(graph_projection_refs) = 'array'),
    ADD CONSTRAINT reporting_releases_composition_tuple_ck CHECK (
        (
            composition_id IS NULL
            AND composition_version IS NULL
            AND composition_sha256 IS NULL
        )
        OR
        (
            composition_id IS NOT NULL
            AND composition_version ~ '^v[1-9][0-9]*$'
            AND composition_sha256 ~ '^[a-f0-9]{64}$'
        )
    );

-- +goose Down
ALTER TABLE reporting_releases
    DROP CONSTRAINT IF EXISTS reporting_releases_composition_tuple_ck,
    DROP CONSTRAINT IF EXISTS reporting_releases_graph_projection_refs_ck,
    DROP CONSTRAINT IF EXISTS reporting_releases_output_options_ck;

ALTER TABLE reporting_releases
    DROP COLUMN IF EXISTS render_admitted_at,
    DROP COLUMN IF EXISTS composition_sha256,
    DROP COLUMN IF EXISTS composition_version,
    DROP COLUMN IF EXISTS composition_id,
    DROP COLUMN IF EXISTS graph_projection_refs,
    DROP COLUMN IF EXISTS output_options;

ALTER TABLE graph_projection_runs
    DROP CONSTRAINT IF EXISTS graph_projection_runs_projection_output_digest_ck;

ALTER TABLE graph_projection_runs
    DROP COLUMN IF EXISTS projection_output_digest;

ALTER TABLE report_composition_release_bindings
    DROP CONSTRAINT IF EXISTS report_composition_release_bindings_release_fk;

ALTER TABLE report_composition_preview_attempts
    DROP CONSTRAINT IF EXISTS report_composition_preview_attempts_render_attempt_fk;

ALTER TABLE report_composition_preview_attempts
    ALTER COLUMN render_attempt_id DROP NOT NULL;

ALTER TABLE report_composition_preview_attempts
    DROP COLUMN IF EXISTS preview_source_bytes;

ALTER TABLE report_composition_versions
    DROP CONSTRAINT IF EXISTS report_composition_versions_canonical_bytes_ck;

ALTER TABLE report_composition_versions
    DROP COLUMN IF EXISTS canonical_composition_bytes;
