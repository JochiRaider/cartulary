-- +goose Up
-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (
        SELECT 1
          FROM compromise_assessments
         WHERE state NOT IN ('unknown', 'suspected', 'confirmed', 'disproven', 'cleared', 'compromised')
    ) THEN
        RAISE EXCEPTION 'cannot migrate compromise_assessments: unexpected assessment state outside Core 02 vocabulary';
    END IF;
END $$;
-- +goose StatementEnd

ALTER TABLE IF EXISTS compromise_assessments
    DROP CONSTRAINT IF EXISTS compromise_assessments_record_envelope_fkey;

ALTER TABLE IF EXISTS compromise_assessments
    RENAME TO assessments;

ALTER TABLE assessments
    RENAME COLUMN compromise_assessment_id TO record_id;

ALTER TABLE assessments
    RENAME COLUMN subject_id TO subject_record_id;

ALTER TABLE assessments
    RENAME COLUMN state TO assessment_state;

ALTER TABLE assessments
    RENAME COLUMN confidence TO confidence_score;

ALTER TABLE assessments
    RENAME COLUMN assessed_by_user_id TO assessor_user_id;

DROP INDEX IF EXISTS compromise_assessments_active_subject_lookup_idx;

UPDATE assessments
   SET assessment_state = 'confirmed'
 WHERE assessment_state = 'compromised';

ALTER TABLE assessments
    ADD COLUMN IF NOT EXISTS rationale text;

UPDATE assessments
   SET rationale = 'Migrated legacy assessment without captured rationale.'
 WHERE rationale IS NULL;

UPDATE assessments a
   SET assessor_user_id = r.created_by_user_id
  FROM records r
 WHERE r.record_id = a.record_id
   AND a.assessor_user_id IS NULL
   AND r.created_by_user_id IS NOT NULL;

-- +goose StatementBegin
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM assessments WHERE assessor_user_id IS NULL) THEN
        RAISE EXCEPTION 'cannot migrate assessments: assessor_user_id is required';
    END IF;
END $$;
-- +goose StatementEnd

ALTER TABLE assessments
    DROP CONSTRAINT IF EXISTS compromise_assessments_subject_type_check,
    DROP CONSTRAINT IF EXISTS compromise_assessments_confidence_check,
    DROP CONSTRAINT IF EXISTS assessments_subject_type_ck,
    DROP CONSTRAINT IF EXISTS assessments_assessment_state_ck,
    DROP CONSTRAINT IF EXISTS assessments_confidence_score_ck,
    DROP CONSTRAINT IF EXISTS assessments_record_envelope_fkey;

ALTER TABLE assessments
    ALTER COLUMN rationale SET NOT NULL,
    ALTER COLUMN assessor_user_id SET NOT NULL,
    ADD CONSTRAINT assessments_subject_type_ck CHECK (subject_type IN ('host', 'identity')),
    ADD CONSTRAINT assessments_assessment_state_ck CHECK (assessment_state IN ('unknown', 'suspected', 'confirmed', 'disproven', 'cleared')),
    ADD CONSTRAINT assessments_confidence_score_ck CHECK (confidence_score IS NULL OR (confidence_score >= 0 AND confidence_score <= 100)),
    ADD CONSTRAINT assessments_record_envelope_fkey
        FOREIGN KEY (incident_id, record_id) REFERENCES records (incident_id, record_id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS assessments_active_subject_lookup_idx
    ON assessments (incident_id, subject_type, subject_record_id, assessed_at DESC, record_id ASC)
    WHERE deleted_at IS NULL;

ALTER TABLE record_links
    DROP CONSTRAINT IF EXISTS record_links_link_type_check;

ALTER TABLE record_links
    ADD CONSTRAINT record_links_link_type_check CHECK (link_type IN ('supersedes', 'observed_on_host', 'observed_as_identity', 'supported_by'));

CREATE TABLE IF NOT EXISTS assessment_grid_projection (
    record_id uuid PRIMARY KEY REFERENCES assessments (record_id) ON DELETE CASCADE,
    incident_id uuid NOT NULL REFERENCES incidents (id) ON DELETE CASCADE,
    row_version bigint NOT NULL,
    subject_ref uuid NOT NULL,
    subject_type text NOT NULL CHECK (subject_type IN ('host', 'identity')),
    assessment_state text NOT NULL CHECK (assessment_state IN ('unknown', 'suspected', 'confirmed', 'disproven', 'cleared')),
    confidence_score integer CHECK (confidence_score IS NULL OR (confidence_score >= 0 AND confidence_score <= 100)),
    confidence_band text NOT NULL CHECK (confidence_band IN ('unset', 'low', 'medium', 'high')),
    rationale text NOT NULL,
    assessor uuid NOT NULL REFERENCES users (id),
    assessed_at timestamptz NOT NULL,
    supporting_link_count integer NOT NULL DEFAULT 0
);

CREATE INDEX IF NOT EXISTS assessment_grid_projection_incident_sort_idx
    ON assessment_grid_projection (incident_id, assessed_at DESC, record_id ASC);

CREATE INDEX IF NOT EXISTS assessment_grid_projection_incident_state_idx
    ON assessment_grid_projection (incident_id, assessment_state, assessed_at DESC, record_id ASC);

CREATE INDEX IF NOT EXISTS assessment_grid_projection_incident_confidence_band_idx
    ON assessment_grid_projection (incident_id, confidence_band, assessed_at DESC, record_id ASC);

INSERT INTO assessment_grid_projection (
    record_id,
    incident_id,
    row_version,
    subject_ref,
    subject_type,
    assessment_state,
    confidence_score,
    confidence_band,
    rationale,
    assessor,
    assessed_at,
    supporting_link_count
)
SELECT
    a.record_id,
    a.incident_id,
    r.row_version,
    a.subject_record_id,
    a.subject_type,
    a.assessment_state,
    a.confidence_score,
    CASE
        WHEN a.confidence_score IS NULL THEN 'unset'
        WHEN a.confidence_score BETWEEN 0 AND 39 THEN 'low'
        WHEN a.confidence_score BETWEEN 40 AND 69 THEN 'medium'
        ELSE 'high'
    END,
    a.rationale,
    a.assessor_user_id,
    a.assessed_at,
    COALESCE(links.supporting_link_count, 0)::integer
  FROM assessments a
  JOIN records r
    ON r.record_id = a.record_id
  LEFT JOIN (
        SELECT
            rl.incident_id,
            rl.src_record_id,
            COUNT(*) AS supporting_link_count
          FROM record_links rl
          JOIN records src
            ON src.incident_id = rl.incident_id
           AND src.record_id = rl.src_record_id
           AND src.deleted_at IS NULL
          JOIN records dst
            ON dst.incident_id = rl.incident_id
           AND dst.record_id = rl.dst_record_id
           AND dst.deleted_at IS NULL
         WHERE rl.deleted_at IS NULL
           AND rl.link_type = 'supported_by'
         GROUP BY rl.incident_id, rl.src_record_id
  ) links
    ON links.incident_id = a.incident_id
   AND links.src_record_id = a.record_id
 WHERE a.deleted_at IS NULL
   AND r.deleted_at IS NULL
ON CONFLICT (record_id) DO NOTHING;

-- +goose Down
DROP INDEX IF EXISTS assessment_grid_projection_incident_confidence_band_idx;
DROP INDEX IF EXISTS assessment_grid_projection_incident_state_idx;
DROP INDEX IF EXISTS assessment_grid_projection_incident_sort_idx;
DROP TABLE IF EXISTS assessment_grid_projection;

ALTER TABLE record_links
    DROP CONSTRAINT IF EXISTS record_links_link_type_check;

ALTER TABLE record_links
    ADD CONSTRAINT record_links_link_type_check CHECK (link_type IN ('supersedes', 'observed_on_host', 'observed_as_identity'));

DROP INDEX IF EXISTS assessments_active_subject_lookup_idx;

ALTER TABLE assessments
    DROP CONSTRAINT IF EXISTS assessments_record_envelope_fkey,
    DROP CONSTRAINT IF EXISTS assessments_confidence_score_ck,
    DROP CONSTRAINT IF EXISTS assessments_assessment_state_ck,
    DROP CONSTRAINT IF EXISTS assessments_subject_type_ck;

ALTER TABLE assessments
    ALTER COLUMN assessor_user_id DROP NOT NULL,
    ALTER COLUMN rationale DROP NOT NULL;

ALTER TABLE assessments
    DROP COLUMN IF EXISTS rationale;

ALTER TABLE assessments
    RENAME COLUMN assessor_user_id TO assessed_by_user_id;

ALTER TABLE assessments
    RENAME COLUMN confidence_score TO confidence;

ALTER TABLE assessments
    RENAME COLUMN assessment_state TO state;

ALTER TABLE assessments
    RENAME COLUMN subject_record_id TO subject_id;

ALTER TABLE assessments
    RENAME COLUMN record_id TO compromise_assessment_id;

ALTER TABLE assessments
    RENAME TO compromise_assessments;

ALTER TABLE compromise_assessments
    ADD CONSTRAINT compromise_assessments_subject_type_check CHECK (subject_type IN ('host', 'identity')),
    ADD CONSTRAINT compromise_assessments_confidence_check CHECK (confidence IS NULL OR (confidence >= 0 AND confidence <= 100)),
    ADD CONSTRAINT compromise_assessments_record_envelope_fkey
        FOREIGN KEY (incident_id, compromise_assessment_id) REFERENCES records (incident_id, record_id) ON DELETE CASCADE;

CREATE INDEX IF NOT EXISTS compromise_assessments_active_subject_lookup_idx
    ON compromise_assessments (incident_id, subject_type, subject_id, compromise_assessment_id)
    WHERE deleted_at IS NULL;
