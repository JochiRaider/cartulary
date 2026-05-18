-- +goose Up
ALTER TABLE timeline_events
    ADD COLUMN IF NOT EXISTS raw_capture jsonb NOT NULL DEFAULT '{}'::jsonb;

-- +goose Down
ALTER TABLE timeline_events
    DROP COLUMN IF EXISTS raw_capture;
