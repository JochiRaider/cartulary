-- +goose Up

CREATE TABLE public.reporting_composition_preview_outputs (
    preview_attempt_id uuid NOT NULL,
    render_attempt_id uuid NOT NULL,
    release_scope text DEFAULT 'internal_draft' NOT NULL,
    output_kind text NOT NULL,
    output_media_type text NOT NULL,
    output_sha256 text NOT NULL,
    redaction_profile_id text NOT NULL,
    redaction_profile_version text NOT NULL,
    redaction_profile_sha256 text NOT NULL,
    redaction_manifest_sha256 text NOT NULL,
    redaction_manifest_json jsonb NOT NULL,
    bundle_manifest_sha256 text NOT NULL,
    bundle_manifest_json jsonb NOT NULL,
    primary_bundle_path text NOT NULL,
    primary_media_type text NOT NULL,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT reporting_composition_preview_outputs_pkey PRIMARY KEY (preview_attempt_id),
    CONSTRAINT reporting_composition_preview_outputs_render_attempt_key UNIQUE (render_attempt_id),
    CONSTRAINT reporting_composition_preview_outputs_scope_ck CHECK (release_scope = 'internal_draft'),
    CONSTRAINT reporting_composition_preview_outputs_kind_ck CHECK (output_kind = ANY (ARRAY['slidev'::text, 'mermaid'::text])),
    CONSTRAINT reporting_composition_preview_outputs_sha_ck CHECK (
        output_sha256 ~ '^[a-f0-9]{64}$'
        AND redaction_profile_sha256 ~ '^[a-f0-9]{64}$'
        AND redaction_manifest_sha256 ~ '^[a-f0-9]{64}$'
        AND bundle_manifest_sha256 ~ '^[a-f0-9]{64}$'
    ),
    CONSTRAINT reporting_composition_preview_outputs_manifest_ck CHECK (
        jsonb_typeof(redaction_manifest_json) = 'object'
        AND jsonb_typeof(bundle_manifest_json) = 'object'
    ),
    CONSTRAINT reporting_composition_preview_outputs_attempt_fk
        FOREIGN KEY (preview_attempt_id)
        REFERENCES public.report_composition_preview_attempts(preview_attempt_id)
        ON DELETE CASCADE,
    CONSTRAINT reporting_composition_preview_outputs_job_fk
        FOREIGN KEY (render_attempt_id)
        REFERENCES public.jobs(job_id)
        ON DELETE RESTRICT
);

CREATE TABLE public.reporting_composition_preview_output_files (
    preview_attempt_id uuid NOT NULL,
    bundle_path text NOT NULL,
    role text NOT NULL,
    media_type text NOT NULL,
    file_sha256 text NOT NULL,
    size_bytes bigint NOT NULL,
    storage_kind text NOT NULL,
    object_ref text,
    inline_bytes bytea,
    created_at timestamp with time zone DEFAULT now() NOT NULL,
    CONSTRAINT reporting_composition_preview_output_files_pkey PRIMARY KEY (preview_attempt_id, bundle_path),
    CONSTRAINT reporting_composition_preview_output_files_path_ck CHECK (
        bundle_path <> '' AND bundle_path !~ '(^/|(^|/)\.\.?(/|$))'
    ),
    CONSTRAINT reporting_composition_preview_output_files_sha_ck CHECK (file_sha256 ~ '^[a-f0-9]{64}$'),
    CONSTRAINT reporting_composition_preview_output_files_size_ck CHECK (size_bytes >= 0),
    CONSTRAINT reporting_composition_preview_output_files_storage_ck CHECK (
        (storage_kind = 'database_inline' AND inline_bytes IS NOT NULL AND object_ref IS NULL)
        OR
        (storage_kind = 'object_store' AND inline_bytes IS NULL AND object_ref IS NOT NULL AND object_ref <> '')
    ),
    CONSTRAINT reporting_composition_preview_output_files_attempt_fk
        FOREIGN KEY (preview_attempt_id)
        REFERENCES public.reporting_composition_preview_outputs(preview_attempt_id)
        ON DELETE CASCADE
);

-- +goose Down

DROP TABLE public.reporting_composition_preview_output_files;
DROP TABLE public.reporting_composition_preview_outputs;
