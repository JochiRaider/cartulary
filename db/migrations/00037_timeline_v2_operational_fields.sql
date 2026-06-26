-- +goose Up
ALTER TABLE timeline_events
    ADD COLUMN IF NOT EXISTS date_entered_text text,
    ADD COLUMN IF NOT EXISTS analyst_text text,
    ADD COLUMN IF NOT EXISTS mitre_stage_text text,
    ADD COLUMN IF NOT EXISTS device_object_text text,
    ADD COLUMN IF NOT EXISTS ip_address_text text,
    ADD COLUMN IF NOT EXISTS activity_utc_text text,
    ADD COLUMN IF NOT EXISTS activity_local_text text,
    ADD COLUMN IF NOT EXISTS raw_activity_text text,
    ADD COLUMN IF NOT EXISTS activity_synopsis_text text,
    ADD COLUMN IF NOT EXISTS data_source_text text,
    ADD COLUMN IF NOT EXISTS activity_utc_generated boolean NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS activity_local_generated boolean NOT NULL DEFAULT false,
    ADD COLUMN IF NOT EXISTS activity_time_pair_state text NOT NULL DEFAULT 'disabled'
        CHECK (activity_time_pair_state IN (
            'disabled',
            'empty',
            'paired_generated',
            'paired_user_preserved',
            'paired_mismatch',
            'conversion_unavailable'
        ));

UPDATE timeline_events
   SET activity_utc_text = COALESCE(activity_utc_text, CASE WHEN occurred_at IS NULL THEN NULL ELSE to_char(occurred_at AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"') END),
       activity_synopsis_text = COALESCE(activity_synopsis_text, summary),
       raw_activity_text = COALESCE(raw_activity_text, source_text),
       date_entered_text = COALESCE(date_entered_text, to_char(recorded_at AT TIME ZONE 'UTC', 'YYYY-MM-DD"T"HH24:MI:SS"Z"'))
 WHERE activity_utc_text IS NULL
    OR activity_synopsis_text IS NULL
    OR raw_activity_text IS NULL
    OR date_entered_text IS NULL;

ALTER TABLE timeline_grid_projection
    ADD COLUMN IF NOT EXISTS date_entered_text text,
    ADD COLUMN IF NOT EXISTS analyst_text text,
    ADD COLUMN IF NOT EXISTS mitre_stage_text text,
    ADD COLUMN IF NOT EXISTS device_object_text text,
    ADD COLUMN IF NOT EXISTS ip_address_text text,
    ADD COLUMN IF NOT EXISTS activity_utc_text text,
    ADD COLUMN IF NOT EXISTS activity_local_text text,
    ADD COLUMN IF NOT EXISTS raw_activity_text text,
    ADD COLUMN IF NOT EXISTS activity_synopsis_text text,
    ADD COLUMN IF NOT EXISTS data_source_text text,
    ADD COLUMN IF NOT EXISTS activity_sort_ts timestamptz,
    ADD COLUMN IF NOT EXISTS date_entered_sort_day date,
    ADD COLUMN IF NOT EXISTS activity_time_pair_state text NOT NULL DEFAULT 'disabled'
        CHECK (activity_time_pair_state IN (
            'disabled',
            'empty',
            'paired_generated',
            'paired_user_preserved',
            'paired_mismatch',
            'conversion_unavailable'
        ));

UPDATE timeline_grid_projection p
   SET date_entered_text = e.date_entered_text,
       analyst_text = e.analyst_text,
       mitre_stage_text = e.mitre_stage_text,
       device_object_text = e.device_object_text,
       ip_address_text = e.ip_address_text,
       activity_utc_text = e.activity_utc_text,
       activity_local_text = e.activity_local_text,
       raw_activity_text = e.raw_activity_text,
       activity_synopsis_text = e.activity_synopsis_text,
       data_source_text = e.data_source_text,
       activity_sort_ts = e.occurred_at,
       date_entered_sort_day = e.recorded_at::date,
       activity_time_pair_state = e.activity_time_pair_state
  FROM timeline_events e
 WHERE e.record_id = p.record_id;

UPDATE entity_mentions
   SET source_field_key = CASE source_field_key
       WHEN 'timeline.summary' THEN 'timeline.activity_synopsis_text'
       WHEN 'timeline.details' THEN 'timeline.raw_activity_text'
       WHEN 'timeline.source_text' THEN 'timeline.raw_activity_text'
       WHEN 'timeline.occurred_at' THEN 'timeline.activity_utc_text'
       ELSE source_field_key
   END
 WHERE source_field_key IN (
       'timeline.summary',
       'timeline.details',
       'timeline.source_text',
       'timeline.occurred_at'
   );

CREATE TABLE IF NOT EXISTS timeline_time_conversion_profiles (
    incident_id uuid PRIMARY KEY REFERENCES incidents (id) ON DELETE CASCADE,
    enabled boolean NOT NULL DEFAULT false,
    local_offset_minutes integer CHECK (local_offset_minutes IS NULL OR (local_offset_minutes >= -840 AND local_offset_minutes <= 840)),
    local_label text,
    profile_version bigint NOT NULL DEFAULT 1,
    updated_at timestamptz NOT NULL DEFAULT now(),
    updated_by_user_id uuid REFERENCES users (id),
    CONSTRAINT timeline_time_conversion_profiles_enabled_offset_ck
        CHECK (enabled = false OR local_offset_minutes IS NOT NULL)
);

CREATE INDEX IF NOT EXISTS timeline_grid_projection_incident_activity_sort_idx
    ON timeline_grid_projection (incident_id, activity_sort_ts ASC, record_id ASC);

DROP INDEX IF EXISTS timeline_grid_projection_incident_capture_state_idx;
DROP INDEX IF EXISTS timeline_grid_projection_incident_sort_idx;

ALTER TABLE timeline_grid_projection
    DROP COLUMN IF EXISTS recorded_day,
    DROP COLUMN IF EXISTS occurred_day,
    DROP COLUMN IF EXISTS sort_ts,
    DROP COLUMN IF EXISTS source_text,
    DROP COLUMN IF EXISTS details,
    DROP COLUMN IF EXISTS summary,
    DROP COLUMN IF EXISTS occurred_at;

ALTER TABLE timeline_events
    DROP COLUMN IF EXISTS source_text,
    DROP COLUMN IF EXISTS details,
    DROP COLUMN IF EXISTS summary,
    DROP COLUMN IF EXISTS occurred_at;

-- +goose Down
ALTER TABLE timeline_events
    ADD COLUMN IF NOT EXISTS occurred_at timestamptz,
    ADD COLUMN IF NOT EXISTS summary text,
    ADD COLUMN IF NOT EXISTS details text,
    ADD COLUMN IF NOT EXISTS source_text text;

UPDATE timeline_events
   SET occurred_at = COALESCE(
           occurred_at,
           CASE
               WHEN activity_utc_text ~ '^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}Z$'
                   THEN activity_utc_text::timestamptz
               WHEN activity_local_text ~ '^[0-9]{4}-[0-9]{2}-[0-9]{2}T[0-9]{2}:[0-9]{2}:[0-9]{2}[+-][0-9]{2}:[0-9]{2}$'
                   THEN activity_local_text::timestamptz
               ELSE NULL
           END
       ),
       summary = COALESCE(summary, activity_synopsis_text),
       source_text = COALESCE(source_text, raw_activity_text)
 WHERE occurred_at IS NULL
    OR summary IS NULL
    OR source_text IS NULL;

ALTER TABLE timeline_grid_projection
    ADD COLUMN IF NOT EXISTS occurred_at timestamptz,
    ADD COLUMN IF NOT EXISTS summary text,
    ADD COLUMN IF NOT EXISTS details text,
    ADD COLUMN IF NOT EXISTS source_text text,
    ADD COLUMN IF NOT EXISTS sort_ts timestamptz,
    ADD COLUMN IF NOT EXISTS occurred_day date,
    ADD COLUMN IF NOT EXISTS recorded_day date;

UPDATE timeline_grid_projection
   SET occurred_at = activity_sort_ts,
       summary = activity_synopsis_text,
       source_text = raw_activity_text,
       sort_ts = COALESCE(activity_sort_ts, edited_at),
       occurred_day = activity_sort_ts::date,
       recorded_day = COALESCE(date_entered_sort_day, recorded_at::date);

ALTER TABLE timeline_grid_projection
    ALTER COLUMN sort_ts SET NOT NULL,
    ALTER COLUMN recorded_day SET NOT NULL;

CREATE INDEX IF NOT EXISTS timeline_grid_projection_incident_sort_idx
    ON timeline_grid_projection (incident_id, sort_ts ASC, record_id ASC);

CREATE INDEX IF NOT EXISTS timeline_grid_projection_incident_capture_state_idx
    ON timeline_grid_projection (incident_id, capture_state, sort_ts ASC, record_id ASC);

UPDATE entity_mentions
   SET source_field_key = CASE source_field_key
       WHEN 'timeline.activity_synopsis_text' THEN 'timeline.summary'
       WHEN 'timeline.raw_activity_text' THEN 'timeline.source_text'
       WHEN 'timeline.activity_utc_text' THEN 'timeline.occurred_at'
       ELSE source_field_key
   END
 WHERE source_field_key IN (
       'timeline.activity_synopsis_text',
       'timeline.raw_activity_text',
       'timeline.activity_utc_text'
   );

DROP INDEX IF EXISTS timeline_grid_projection_incident_activity_sort_idx;

DROP TABLE IF EXISTS timeline_time_conversion_profiles;

ALTER TABLE timeline_grid_projection
    DROP COLUMN IF EXISTS activity_time_pair_state,
    DROP COLUMN IF EXISTS date_entered_sort_day,
    DROP COLUMN IF EXISTS activity_sort_ts,
    DROP COLUMN IF EXISTS data_source_text,
    DROP COLUMN IF EXISTS activity_synopsis_text,
    DROP COLUMN IF EXISTS raw_activity_text,
    DROP COLUMN IF EXISTS activity_local_text,
    DROP COLUMN IF EXISTS activity_utc_text,
    DROP COLUMN IF EXISTS ip_address_text,
    DROP COLUMN IF EXISTS device_object_text,
    DROP COLUMN IF EXISTS mitre_stage_text,
    DROP COLUMN IF EXISTS analyst_text,
    DROP COLUMN IF EXISTS date_entered_text;

ALTER TABLE timeline_events
    DROP COLUMN IF EXISTS activity_time_pair_state,
    DROP COLUMN IF EXISTS activity_local_generated,
    DROP COLUMN IF EXISTS activity_utc_generated,
    DROP COLUMN IF EXISTS data_source_text,
    DROP COLUMN IF EXISTS activity_synopsis_text,
    DROP COLUMN IF EXISTS raw_activity_text,
    DROP COLUMN IF EXISTS activity_local_text,
    DROP COLUMN IF EXISTS activity_utc_text,
    DROP COLUMN IF EXISTS ip_address_text,
    DROP COLUMN IF EXISTS device_object_text,
    DROP COLUMN IF EXISTS mitre_stage_text,
    DROP COLUMN IF EXISTS analyst_text,
    DROP COLUMN IF EXISTS date_entered_text;
