-- +goose Up
ALTER TABLE incidents
    DROP CONSTRAINT IF EXISTS incidents_status_check;

ALTER TABLE incidents
    ADD CONSTRAINT incidents_status_ck CHECK (status IN ('active', 'closed')),
    ADD CONSTRAINT incidents_closed_at_status_ck CHECK (
        (status = 'active' AND closed_at IS NULL) OR
        (status = 'closed' AND closed_at IS NOT NULL)
    );

CREATE INDEX IF NOT EXISTS incidents_status_updated_lookup_idx
    ON incidents (status, updated_at DESC, id ASC);

CREATE TABLE IF NOT EXISTS account_preferences (
    user_id uuid PRIMARY KEY REFERENCES users (id) ON DELETE CASCADE,
    density_mode text CHECK (density_mode IN ('compact', 'default', 'comfortable')),
    preferences_version bigint NOT NULL DEFAULT 1,
    created_at timestamptz NOT NULL DEFAULT now(),
    updated_at timestamptz NOT NULL DEFAULT now()
);

-- +goose Down
DROP TABLE IF EXISTS account_preferences;

DROP INDEX IF EXISTS incidents_status_updated_lookup_idx;

ALTER TABLE incidents
    DROP CONSTRAINT IF EXISTS incidents_closed_at_status_ck,
    DROP CONSTRAINT IF EXISTS incidents_status_ck;

UPDATE incidents
   SET status = 'active',
       closed_at = NULL
 WHERE status <> 'active'
    OR closed_at IS NOT NULL;

ALTER TABLE incidents
    ADD CONSTRAINT incidents_status_check CHECK (status IN ('active'));
