-- +goose Up
CREATE TABLE IF NOT EXISTS backup_sets (
    backup_set_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    consistency_point_at timestamptz NOT NULL,
    postgres_restore_anchor text NOT NULL,
    object_store_restore_anchor text NOT NULL,
    postgres_artifact_key text NOT NULL,
    postgres_artifact_sha256 text NOT NULL,
    postgres_artifact_size_bytes bigint NOT NULL,
    object_store_artifact_key text NOT NULL,
    object_store_artifact_sha256 text NOT NULL,
    object_store_artifact_size_bytes bigint NOT NULL,
    integrity_manifest_key text NOT NULL,
    integrity_manifest_sha256 text NOT NULL,
    integrity_manifest_size_bytes bigint NOT NULL,
    created_at timestamptz NOT NULL DEFAULT now(),
    retained_until timestamptz NOT NULL,
    postgres_restore_anchor_retained_until timestamptz NOT NULL,
    object_store_restore_anchor_retained_until timestamptz NOT NULL,
    verification_state text NOT NULL DEFAULT 'unverified',
    last_verified_restore_at timestamptz,
    CONSTRAINT backup_sets_postgres_restore_anchor_non_empty
        CHECK (btrim(postgres_restore_anchor) <> ''),
    CONSTRAINT backup_sets_object_store_restore_anchor_non_empty
        CHECK (btrim(object_store_restore_anchor) <> ''),
    CONSTRAINT backup_sets_postgres_artifact_key_non_empty
        CHECK (btrim(postgres_artifact_key) <> ''),
    CONSTRAINT backup_sets_object_store_artifact_key_non_empty
        CHECK (btrim(object_store_artifact_key) <> ''),
    CONSTRAINT backup_sets_integrity_manifest_key_non_empty
        CHECK (btrim(integrity_manifest_key) <> ''),
    CONSTRAINT backup_sets_postgres_artifact_sha256_check
        CHECK (postgres_artifact_sha256 ~ '^[0-9a-f]{64}$'),
    CONSTRAINT backup_sets_object_store_artifact_sha256_check
        CHECK (object_store_artifact_sha256 ~ '^[0-9a-f]{64}$'),
    CONSTRAINT backup_sets_integrity_manifest_sha256_check
        CHECK (integrity_manifest_sha256 ~ '^[0-9a-f]{64}$'),
    CONSTRAINT backup_sets_postgres_artifact_size_check
        CHECK (postgres_artifact_size_bytes > 0),
    CONSTRAINT backup_sets_object_store_artifact_size_check
        CHECK (object_store_artifact_size_bytes > 0),
    CONSTRAINT backup_sets_integrity_manifest_size_check
        CHECK (integrity_manifest_size_bytes > 0),
    CONSTRAINT backup_sets_verification_state_check
        CHECK (verification_state IN ('unverified', 'verified', 'failed')),
    CONSTRAINT backup_sets_verification_timestamp_check
        CHECK (
            (verification_state = 'unverified' AND last_verified_restore_at IS NULL)
            OR (verification_state IN ('verified', 'failed') AND last_verified_restore_at IS NOT NULL)
        ),
    CONSTRAINT backup_sets_retained_until_floor_check
        CHECK (retained_until >= created_at + interval '30 days'),
    CONSTRAINT backup_sets_postgres_anchor_retained_until_floor_check
        CHECK (postgres_restore_anchor_retained_until >= created_at + interval '30 days'),
    CONSTRAINT backup_sets_object_store_anchor_retained_until_floor_check
        CHECK (object_store_restore_anchor_retained_until >= created_at + interval '30 days')
);

CREATE INDEX IF NOT EXISTS backup_sets_latest_retained_idx
    ON backup_sets (consistency_point_at DESC, backup_set_id ASC, retained_until);

-- +goose Down
DROP TABLE IF EXISTS backup_sets;
