-- +goose Up
CREATE TABLE IF NOT EXISTS evidence_custody_events (
    custody_event_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id uuid NOT NULL REFERENCES incidents (id) ON DELETE CASCADE,
    evidence_record_id uuid NOT NULL REFERENCES evidence (record_id) ON DELETE CASCADE,
    custody_event_type text NOT NULL CHECK (
        custody_event_type IN (
            'requested',
            'received',
            'made_available',
            'transferred',
            'quarantined',
            'released'
        )
    ),
    actor_user_id uuid REFERENCES users (id),
    occurred_at timestamptz NOT NULL DEFAULT now(),
    location_text text,
    note text,
    metadata jsonb NOT NULL DEFAULT '{}'::jsonb
);

CREATE INDEX IF NOT EXISTS evidence_custody_events_record_time_idx
    ON evidence_custody_events (evidence_record_id, occurred_at DESC, custody_event_id DESC);

CREATE TABLE IF NOT EXISTS incident_bundle_imported_attributions (
    imported_attribution_id uuid PRIMARY KEY DEFAULT gen_random_uuid(),
    incident_id uuid NOT NULL REFERENCES incidents (id) ON DELETE CASCADE,
    source_table text NOT NULL,
    source_row_id text NOT NULL,
    source_column text NOT NULL,
    source_actor_id text NOT NULL,
    local_user_id uuid NOT NULL REFERENCES users (id),
    imported_at timestamptz NOT NULL DEFAULT now(),
    UNIQUE (incident_id, source_table, source_row_id, source_column)
);

CREATE INDEX IF NOT EXISTS incident_bundle_imported_attributions_actor_idx
    ON incident_bundle_imported_attributions (incident_id, source_actor_id);

-- +goose Down
DROP INDEX IF EXISTS incident_bundle_imported_attributions_actor_idx;
DROP TABLE IF EXISTS incident_bundle_imported_attributions;
DROP INDEX IF EXISTS evidence_custody_events_record_time_idx;
DROP TABLE IF EXISTS evidence_custody_events;
