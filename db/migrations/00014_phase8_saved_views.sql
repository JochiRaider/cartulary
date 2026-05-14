-- +goose Up
CREATE TABLE IF NOT EXISTS saved_views (
    saved_view_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id uuid NOT NULL REFERENCES incidents (id) ON DELETE CASCADE,
    view_schema_id text NOT NULL,
    scope text NOT NULL CHECK (scope IN ('private', 'shared', 'system')),
    display_name text NOT NULL,
    query_json jsonb NOT NULL,
    layout_json jsonb NOT NULL,
    owner_user_id uuid REFERENCES users (id),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    saved_view_version bigint NOT NULL DEFAULT 1,
    CONSTRAINT saved_views_owner_scope_ck CHECK (owner_user_id IS NOT NULL OR scope = 'system')
);

CREATE INDEX IF NOT EXISTS saved_views_incident_order_idx
    ON saved_views (incident_id, updated_at DESC, saved_view_id ASC);

CREATE INDEX IF NOT EXISTS saved_views_owner_lookup_idx
    ON saved_views (incident_id, owner_user_id)
    WHERE scope = 'private';

-- +goose Down
DROP INDEX IF EXISTS saved_views_owner_lookup_idx;
DROP INDEX IF EXISTS saved_views_incident_order_idx;
DROP TABLE IF EXISTS saved_views;
