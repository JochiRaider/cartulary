-- +goose Up
CREATE TABLE IF NOT EXISTS report_compositions (
    composition_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id uuid NOT NULL REFERENCES incidents (id) ON DELETE CASCADE,
    created_by_user_id uuid NOT NULL REFERENCES users (id),
    client_txn_id text NOT NULL,
    template_id text NOT NULL,
    template_version text NOT NULL,
    draft_version bigint NOT NULL DEFAULT 1 CHECK (draft_version > 0),
    authored_against_snapshot_id text,
    deck_ops jsonb NOT NULL DEFAULT '[]'::jsonb,
    diagram_decls jsonb NOT NULL DEFAULT '[]'::jsonb,
    authored_texts jsonb NOT NULL DEFAULT '[]'::jsonb,
    latest_composition_version bigint CHECK (latest_composition_version IS NULL OR latest_composition_version > 0),
    retired_at timestamptz,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT report_compositions_deck_ops_array_ck CHECK (jsonb_typeof(deck_ops) = 'array'),
    CONSTRAINT report_compositions_diagram_decls_array_ck CHECK (jsonb_typeof(diagram_decls) = 'array'),
    CONSTRAINT report_compositions_authored_texts_array_ck CHECK (jsonb_typeof(authored_texts) = 'array')
);

CREATE INDEX IF NOT EXISTS report_compositions_incident_lookup_idx
    ON report_compositions (incident_id, retired_at NULLS FIRST, template_id, template_version, composition_id);

CREATE INDEX IF NOT EXISTS report_compositions_created_by_lookup_idx
    ON report_compositions (created_by_user_id, created_at DESC, composition_id);

CREATE TABLE IF NOT EXISTS report_composition_versions (
    composition_id uuid NOT NULL REFERENCES report_compositions (composition_id) ON DELETE CASCADE,
    composition_version bigint NOT NULL CHECK (composition_version > 0),
    composition_sha256 text NOT NULL CHECK (composition_sha256 ~ '^[a-f0-9]{64}$'),
    canonical_composition jsonb NOT NULL CHECK (jsonb_typeof(canonical_composition) = 'object'),
    created_by_user_id uuid NOT NULL REFERENCES users (id),
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (composition_id, composition_version)
);

CREATE UNIQUE INDEX IF NOT EXISTS report_composition_versions_digest_idx
    ON report_composition_versions (composition_id, composition_version, composition_sha256);

CREATE TABLE IF NOT EXISTS report_composition_release_bindings (
    composition_id uuid NOT NULL,
    composition_version bigint NOT NULL,
    composition_sha256 text NOT NULL CHECK (composition_sha256 ~ '^[a-f0-9]{64}$'),
    release_id uuid NOT NULL,
    release_scope text NOT NULL CHECK (release_scope IN ('internal_draft', 'internal_review', 'external_release')),
    created_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (composition_id, composition_version, release_id),
    FOREIGN KEY (composition_id, composition_version)
        REFERENCES report_composition_versions (composition_id, composition_version)
        ON DELETE RESTRICT
);

CREATE INDEX IF NOT EXISTS report_composition_release_bindings_release_lookup_idx
    ON report_composition_release_bindings (release_id, composition_id, composition_version);

CREATE TABLE IF NOT EXISTS report_composition_preview_attempts (
    preview_attempt_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id uuid NOT NULL REFERENCES incidents (id) ON DELETE CASCADE,
    composition_id uuid NOT NULL REFERENCES report_compositions (composition_id) ON DELETE CASCADE,
    source_kind text NOT NULL CHECK (source_kind IN ('draft', 'version')),
    draft_version bigint CHECK (draft_version IS NULL OR draft_version > 0),
    composition_version bigint CHECK (composition_version IS NULL OR composition_version > 0),
    preview_source_sha256 text NOT NULL CHECK (preview_source_sha256 ~ '^[a-f0-9]{64}$'),
    composition_sha256 text CHECK (composition_sha256 IS NULL OR composition_sha256 ~ '^[a-f0-9]{64}$'),
    preview_source_json jsonb NOT NULL CHECK (jsonb_typeof(preview_source_json) = 'object'),
    snapshot_id text NOT NULL,
    derivation_version text NOT NULL,
    template_id text NOT NULL,
    template_version text NOT NULL,
    redaction_profile_id text NOT NULL,
    redaction_profile_version text NOT NULL,
    redaction_profile_sha256 text NOT NULL CHECK (redaction_profile_sha256 ~ '^[a-f0-9]{64}$'),
    render_environment_profile_id text NOT NULL,
    output_kind text NOT NULL CHECK (output_kind IN ('slidev', 'mermaid')),
    output_options jsonb NOT NULL DEFAULT '{}'::jsonb CHECK (jsonb_typeof(output_options) = 'object'),
    recipient_partition_refs jsonb NOT NULL DEFAULT '[]'::jsonb CHECK (jsonb_typeof(recipient_partition_refs) = 'array'),
    graph_projection_refs jsonb NOT NULL DEFAULT '[]'::jsonb CHECK (jsonb_typeof(graph_projection_refs) = 'array'),
    render_attempt_id uuid,
    created_by_user_id uuid NOT NULL REFERENCES users (id),
    created_at timestamptz NOT NULL DEFAULT now(),
    CONSTRAINT report_composition_preview_source_shape_ck CHECK (
        (source_kind = 'draft' AND draft_version IS NOT NULL AND composition_version IS NULL AND composition_sha256 IS NULL)
        OR
        (source_kind = 'version' AND draft_version IS NULL AND composition_version IS NOT NULL AND composition_sha256 IS NOT NULL)
    )
);

CREATE INDEX IF NOT EXISTS report_composition_preview_attempts_composition_lookup_idx
    ON report_composition_preview_attempts (composition_id, created_at DESC, preview_attempt_id);

-- +goose Down
DROP INDEX IF EXISTS report_composition_preview_attempts_composition_lookup_idx;
DROP TABLE IF EXISTS report_composition_preview_attempts;

DROP INDEX IF EXISTS report_composition_release_bindings_release_lookup_idx;
DROP TABLE IF EXISTS report_composition_release_bindings;

DROP INDEX IF EXISTS report_composition_versions_digest_idx;
DROP TABLE IF EXISTS report_composition_versions;

DROP INDEX IF EXISTS report_compositions_created_by_lookup_idx;
DROP INDEX IF EXISTS report_compositions_incident_lookup_idx;
DROP TABLE IF EXISTS report_compositions;
