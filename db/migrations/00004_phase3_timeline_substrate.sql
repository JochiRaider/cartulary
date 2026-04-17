-- +goose Up
CREATE TABLE IF NOT EXISTS timeline_events (
    record_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id uuid NOT NULL REFERENCES incidents (id) ON DELETE CASCADE,
    occurred_at timestamptz,
    summary text,
    details text,
    source_text text,
    capture_state text NOT NULL CHECK (capture_state IN ('rough', 'enriched', 'reviewed', 'superseded')),
    row_version bigint NOT NULL DEFAULT 1,
    recorded_at timestamptz NOT NULL DEFAULT now(),
    edited_at timestamptz NOT NULL DEFAULT now(),
    created_by_user_id uuid NOT NULL REFERENCES users (id),
    updated_by_user_id uuid NOT NULL REFERENCES users (id),
    reviewed_by_user_id uuid REFERENCES users (id),
    reviewed_at timestamptz,
    superseded_by_user_id uuid REFERENCES users (id),
    superseded_at timestamptz
);

CREATE INDEX IF NOT EXISTS timeline_events_incident_lookup_idx
    ON timeline_events (incident_id, edited_at DESC, record_id ASC);

CREATE TABLE IF NOT EXISTS change_sets (
    change_set_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id uuid NOT NULL REFERENCES incidents (id) ON DELETE CASCADE,
    actor_user_id uuid NOT NULL REFERENCES users (id),
    source text NOT NULL,
    reason text,
    client_txn_id text,
    request_id text,
    created_at timestamptz NOT NULL DEFAULT now()
);

CREATE INDEX IF NOT EXISTS change_sets_incident_lookup_idx
    ON change_sets (incident_id, created_at DESC, change_set_id DESC);

CREATE TABLE IF NOT EXISTS change_set_mutations (
    change_set_id uuid NOT NULL REFERENCES change_sets (change_set_id) ON DELETE CASCADE,
    sequence_no integer NOT NULL,
    target_kind text NOT NULL,
    target_id text NOT NULL,
    operation_kind text NOT NULL,
    before_version_id text,
    after_version_id text,
    before_value jsonb,
    after_value jsonb,
    PRIMARY KEY (change_set_id, sequence_no)
);

CREATE INDEX IF NOT EXISTS change_set_mutations_target_lookup_idx
    ON change_set_mutations (target_kind, target_id);

CREATE TABLE IF NOT EXISTS record_revisions (
    revision_id bigserial PRIMARY KEY,
    change_set_id uuid NOT NULL REFERENCES change_sets (change_set_id) ON DELETE CASCADE,
    record_id uuid NOT NULL REFERENCES timeline_events (record_id) ON DELETE CASCADE,
    row_version bigint NOT NULL,
    before_json jsonb,
    after_json jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (record_id, row_version)
);

CREATE INDEX IF NOT EXISTS record_revisions_record_lookup_idx
    ON record_revisions (record_id, row_version DESC);

CREATE TABLE IF NOT EXISTS record_links (
    record_link_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id uuid NOT NULL REFERENCES incidents (id) ON DELETE CASCADE,
    src_record_id uuid NOT NULL REFERENCES timeline_events (record_id) ON DELETE CASCADE,
    dst_record_id uuid NOT NULL REFERENCES timeline_events (record_id) ON DELETE CASCADE,
    link_type text NOT NULL CHECK (link_type IN ('supersedes')),
    provenance text NOT NULL CHECK (provenance IN ('manual', 'auto_match', 'import', 'rollback', 'system')),
    confidence integer CHECK (confidence IS NULL OR (confidence >= 0 AND confidence <= 100)),
    owner_user_id uuid NOT NULL REFERENCES users (id),
    decided_at timestamptz NOT NULL DEFAULT now(),
    created_at timestamptz NOT NULL DEFAULT now(),
    deleted_at timestamptz,
    deleted_by_user_id uuid REFERENCES users (id),
    CONSTRAINT record_links_distinct_endpoints_ck CHECK (src_record_id <> dst_record_id)
);

CREATE UNIQUE INDEX IF NOT EXISTS record_links_active_unique_idx
    ON record_links (incident_id, src_record_id, dst_record_id, link_type)
    WHERE deleted_at IS NULL;

CREATE UNIQUE INDEX IF NOT EXISTS record_links_active_timeline_supersedes_dst_idx
    ON record_links (incident_id, dst_record_id, link_type)
    WHERE deleted_at IS NULL AND link_type = 'supersedes';

CREATE TABLE IF NOT EXISTS timeline_grid_projection (
    record_id uuid PRIMARY KEY REFERENCES timeline_events (record_id) ON DELETE CASCADE,
    incident_id uuid NOT NULL REFERENCES incidents (id) ON DELETE CASCADE,
    row_version bigint NOT NULL,
    occurred_at timestamptz,
    summary text,
    details text,
    source_text text,
    recorded_at timestamptz NOT NULL,
    edited_at timestamptz NOT NULL,
    sort_ts timestamptz NOT NULL,
    capture_state text NOT NULL CHECK (capture_state IN ('rough', 'enriched', 'reviewed', 'superseded')),
    replacement_record_id uuid REFERENCES timeline_events (record_id),
    occurred_day date,
    recorded_day date NOT NULL,
    evidence_count integer NOT NULL DEFAULT 0,
    has_evidence boolean NOT NULL DEFAULT false,
    has_unresolved_mentions boolean NOT NULL DEFAULT false
);

CREATE INDEX IF NOT EXISTS timeline_grid_projection_incident_sort_idx
    ON timeline_grid_projection (incident_id, sort_ts ASC, record_id ASC);

CREATE INDEX IF NOT EXISTS timeline_grid_projection_incident_capture_state_idx
    ON timeline_grid_projection (incident_id, capture_state, sort_ts ASC, record_id ASC);

-- +goose Down
DROP INDEX IF EXISTS timeline_grid_projection_incident_capture_state_idx;
DROP INDEX IF EXISTS timeline_grid_projection_incident_sort_idx;
DROP TABLE IF EXISTS timeline_grid_projection;

DROP INDEX IF EXISTS record_links_active_timeline_supersedes_dst_idx;
DROP INDEX IF EXISTS record_links_active_unique_idx;
DROP TABLE IF EXISTS record_links;

DROP INDEX IF EXISTS record_revisions_record_lookup_idx;
DROP TABLE IF EXISTS record_revisions;

DROP INDEX IF EXISTS change_set_mutations_target_lookup_idx;
DROP TABLE IF EXISTS change_set_mutations;

DROP INDEX IF EXISTS change_sets_incident_lookup_idx;
DROP TABLE IF EXISTS change_sets;

DROP INDEX IF EXISTS timeline_events_incident_lookup_idx;
DROP TABLE IF EXISTS timeline_events;
