-- +goose Up
CREATE TABLE IF NOT EXISTS incidents (
    id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_key text NOT NULL,
    incident_key_canonical text NOT NULL UNIQUE,
    title text NOT NULL,
    description text,
    status text NOT NULL CHECK (status IN ('active')),
    severity text,
    tlp text,
    current_phase text,
    primary_external_case_ref text,
    created_by_user_id uuid NOT NULL REFERENCES users (id),
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    updated_by_user_id uuid NOT NULL REFERENCES users (id),
    incident_version bigint NOT NULL DEFAULT 1,
    closed_at timestamptz
);

CREATE INDEX IF NOT EXISTS incidents_updated_lookup_idx
    ON incidents (updated_at DESC, id ASC);

CREATE TABLE IF NOT EXISTS incident_memberships (
    incident_id uuid NOT NULL REFERENCES incidents (id) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    role text NOT NULL CHECK (role IN ('viewer', 'editor', 'reviewer', 'admin')),
    joined_at timestamptz NOT NULL DEFAULT now(),
    added_by_user_id uuid NOT NULL REFERENCES users (id),
    updated_at timestamptz NOT NULL DEFAULT now(),
    updated_by_user_id uuid NOT NULL REFERENCES users (id),
    membership_version bigint NOT NULL DEFAULT 1,
    PRIMARY KEY (incident_id, user_id)
);

CREATE INDEX IF NOT EXISTS incident_memberships_user_lookup_idx
    ON incident_memberships (user_id, incident_id);

CREATE INDEX IF NOT EXISTS incident_memberships_incident_lookup_idx
    ON incident_memberships (incident_id, joined_at ASC, user_id ASC);

CREATE TABLE IF NOT EXISTS incident_workbook_preferences (
    incident_id uuid PRIMARY KEY REFERENCES incidents (id) ON DELETE CASCADE,
    default_sheet_ref jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    updated_by_user_id uuid NOT NULL REFERENCES users (id)
);

CREATE TABLE IF NOT EXISTS user_workbook_preferences (
    incident_id uuid NOT NULL REFERENCES incidents (id) ON DELETE CASCADE,
    user_id uuid NOT NULL REFERENCES users (id) ON DELETE CASCADE,
    home_sheet_ref jsonb,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now(),
    PRIMARY KEY (incident_id, user_id)
);

ALTER TABLE deployment_admin_audit_events
    ADD COLUMN IF NOT EXISTS incident_id uuid REFERENCES incidents (id);

CREATE INDEX IF NOT EXISTS deployment_admin_audit_events_incident_lookup_idx
    ON deployment_admin_audit_events (incident_id, created_at DESC);

-- +goose Down
DROP INDEX IF EXISTS deployment_admin_audit_events_incident_lookup_idx;

ALTER TABLE deployment_admin_audit_events
    DROP COLUMN IF EXISTS incident_id;

DROP TABLE IF EXISTS user_workbook_preferences;
DROP TABLE IF EXISTS incident_workbook_preferences;

DROP INDEX IF EXISTS incident_memberships_incident_lookup_idx;
DROP INDEX IF EXISTS incident_memberships_user_lookup_idx;
DROP TABLE IF EXISTS incident_memberships;

DROP INDEX IF EXISTS incidents_updated_lookup_idx;
DROP TABLE IF EXISTS incidents;
