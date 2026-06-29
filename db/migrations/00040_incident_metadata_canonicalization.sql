-- +goose Up
-- +goose StatementBegin
DO $$
DECLARE
    report jsonb;
BEGIN
    WITH normalized AS (
        SELECT id::text AS incident_id,
               tlp,
               CASE WHEN tlp IS NULL THEN NULL ELSE upper(btrim(tlp)) END AS tlp_key,
               description,
               CASE
                   WHEN description IS NULL THEN NULL
                   ELSE regexp_replace(replace(replace(description, E'\r\n', E'\n'), E'\r', E'\n'), '^[[:space:]]+|[[:space:]]+$', '', 'g')
               END AS normalized_description,
               severity,
               CASE WHEN severity IS NULL THEN NULL ELSE regexp_replace(severity, '^[[:space:]]+|[[:space:]]+$', '', 'g') END AS normalized_severity,
               current_phase,
               CASE WHEN current_phase IS NULL THEN NULL ELSE regexp_replace(current_phase, '^[[:space:]]+|[[:space:]]+$', '', 'g') END AS normalized_current_phase,
               primary_external_case_ref,
               CASE
                   WHEN primary_external_case_ref IS NULL THEN NULL
                   ELSE regexp_replace(primary_external_case_ref, '^[[:space:]]+|[[:space:]]+$', '', 'g')
               END AS normalized_primary_external_case_ref
          FROM incidents
    ),
    findings AS (
        SELECT incident_id,
               'tlp' AS field,
               tlp AS raw_value,
               'unknown_tlp' AS reason_code,
               'Set tlp to one canonical cartulary.tlp.v1 token or a supported legacy alias before rerunning migration.' AS remediation_hint
          FROM normalized
         WHERE tlp IS NOT NULL
           AND tlp_key NOT IN (
               'TLP:CLEAR', 'TLP:GREEN', 'TLP:AMBER', 'TLP:AMBER+STRICT', 'TLP:RED',
               'CLEAR', 'WHITE', 'TLP:WHITE', 'GREEN', 'AMBER', 'AMBER+STRICT', 'RED'
           )
        UNION ALL
        SELECT incident_id,
               'description',
               description,
               'invalid_description',
               'Remove unsupported controls and keep description within 16384 Unicode scalar values after line-ending normalization and trimming.'
          FROM normalized
         WHERE normalized_description IS NOT NULL
           AND normalized_description <> ''
           AND (
               char_length(normalized_description) > 16384
               OR EXISTS (
                   SELECT 1
                     FROM generate_series(1, char_length(normalized_description)) AS pos(i)
                    WHERE ascii(substr(normalized_description, pos.i, 1)) BETWEEN 1 AND 8
                       OR ascii(substr(normalized_description, pos.i, 1)) IN (11, 12)
                       OR ascii(substr(normalized_description, pos.i, 1)) BETWEEN 14 AND 31
                       OR ascii(substr(normalized_description, pos.i, 1)) BETWEEN 127 AND 159
               )
           )
        UNION ALL
        SELECT incident_id,
               'severity',
               severity,
               'invalid_severity',
               'Remove unsupported controls and keep severity within 128 Unicode scalar values after trimming.'
          FROM normalized
         WHERE normalized_severity IS NOT NULL
           AND normalized_severity <> ''
           AND (
               char_length(normalized_severity) > 128
               OR EXISTS (
                   SELECT 1
                     FROM generate_series(1, char_length(normalized_severity)) AS pos(i)
                    WHERE ascii(substr(normalized_severity, pos.i, 1)) BETWEEN 1 AND 31
                       OR ascii(substr(normalized_severity, pos.i, 1)) BETWEEN 127 AND 159
               )
           )
        UNION ALL
        SELECT incident_id,
               'current_phase',
               current_phase,
               'invalid_current_phase',
               'Remove unsupported controls and keep current_phase within 128 Unicode scalar values after trimming.'
          FROM normalized
         WHERE normalized_current_phase IS NOT NULL
           AND normalized_current_phase <> ''
           AND (
               char_length(normalized_current_phase) > 128
               OR EXISTS (
                   SELECT 1
                     FROM generate_series(1, char_length(normalized_current_phase)) AS pos(i)
                    WHERE ascii(substr(normalized_current_phase, pos.i, 1)) BETWEEN 1 AND 31
                       OR ascii(substr(normalized_current_phase, pos.i, 1)) BETWEEN 127 AND 159
               )
           )
        UNION ALL
        SELECT incident_id,
               'primary_external_case_ref',
               primary_external_case_ref,
               'invalid_primary_external_case_ref',
               'Remove unsupported controls and keep primary_external_case_ref within 128 Unicode scalar values after trimming.'
          FROM normalized
         WHERE normalized_primary_external_case_ref IS NOT NULL
           AND normalized_primary_external_case_ref <> ''
           AND (
               char_length(normalized_primary_external_case_ref) > 128
               OR EXISTS (
                   SELECT 1
                     FROM generate_series(1, char_length(normalized_primary_external_case_ref)) AS pos(i)
                    WHERE ascii(substr(normalized_primary_external_case_ref, pos.i, 1)) BETWEEN 1 AND 31
                       OR ascii(substr(normalized_primary_external_case_ref, pos.i, 1)) BETWEEN 127 AND 159
               )
           )
    ),
    ordered_findings AS (
        SELECT jsonb_build_object(
                   'incident_id', incident_id,
                   'field', field,
                   'raw_value', raw_value,
                   'reason_code', reason_code,
                   'remediation_hint', remediation_hint
               ) AS finding
          FROM findings
         ORDER BY incident_id ASC, field ASC, reason_code ASC
    )
    SELECT jsonb_build_object(
               'schema_id', 'cartulary.migration_remediation_report.v1',
               'boundary', 'incident_metadata_canonicalization_v40',
               'from_version', 39,
               'to_version', 40,
               'findings', COALESCE(jsonb_agg(finding), '[]'::jsonb)
           )
      INTO report
      FROM ordered_findings;

    IF jsonb_array_length(report->'findings') > 0 THEN
        RAISE EXCEPTION 'cartulary.migration_remediation_report.v1:%', report::text
            USING ERRCODE = '23514';
    END IF;
END $$;
-- +goose StatementEnd

UPDATE incidents
   SET tlp = CASE upper(btrim(tlp))
                 WHEN 'TLP:CLEAR' THEN 'TLP:CLEAR'
                 WHEN 'CLEAR' THEN 'TLP:CLEAR'
                 WHEN 'WHITE' THEN 'TLP:CLEAR'
                 WHEN 'TLP:WHITE' THEN 'TLP:CLEAR'
                 WHEN 'TLP:GREEN' THEN 'TLP:GREEN'
                 WHEN 'GREEN' THEN 'TLP:GREEN'
                 WHEN 'TLP:AMBER' THEN 'TLP:AMBER'
                 WHEN 'AMBER' THEN 'TLP:AMBER'
                 WHEN 'TLP:AMBER+STRICT' THEN 'TLP:AMBER+STRICT'
                 WHEN 'AMBER+STRICT' THEN 'TLP:AMBER+STRICT'
                 WHEN 'TLP:RED' THEN 'TLP:RED'
                 WHEN 'RED' THEN 'TLP:RED'
             END
 WHERE tlp IS NOT NULL;

UPDATE incidents
   SET description = NULLIF(regexp_replace(replace(replace(description, E'\r\n', E'\n'), E'\r', E'\n'), '^[[:space:]]+|[[:space:]]+$', '', 'g'), '')
 WHERE description IS NOT NULL;

UPDATE incidents
   SET severity = NULLIF(regexp_replace(severity, '^[[:space:]]+|[[:space:]]+$', '', 'g'), '')
 WHERE severity IS NOT NULL;

UPDATE incidents
   SET current_phase = NULLIF(regexp_replace(current_phase, '^[[:space:]]+|[[:space:]]+$', '', 'g'), '')
 WHERE current_phase IS NOT NULL;

UPDATE incidents
   SET primary_external_case_ref = NULLIF(regexp_replace(primary_external_case_ref, '^[[:space:]]+|[[:space:]]+$', '', 'g'), '')
 WHERE primary_external_case_ref IS NOT NULL;

ALTER TABLE incidents
    ADD CONSTRAINT incidents_tlp_ck CHECK (
        tlp IS NULL OR tlp IN ('TLP:CLEAR', 'TLP:GREEN', 'TLP:AMBER', 'TLP:AMBER+STRICT', 'TLP:RED')
    ),
    ADD CONSTRAINT incidents_description_contract_ck CHECK (
        description IS NULL OR (
            char_length(description) BETWEEN 1 AND 16384
            AND regexp_replace(description, E'[\\n\\t]', '', 'g') !~ '[[:cntrl:]]'
        )
    ),
    ADD CONSTRAINT incidents_severity_contract_ck CHECK (
        severity IS NULL OR (
            char_length(severity) BETWEEN 1 AND 128
            AND severity !~ '[[:cntrl:]]'
        )
    ),
    ADD CONSTRAINT incidents_current_phase_contract_ck CHECK (
        current_phase IS NULL OR (
            char_length(current_phase) BETWEEN 1 AND 128
            AND current_phase !~ '[[:cntrl:]]'
        )
    ),
    ADD CONSTRAINT incidents_primary_external_case_ref_contract_ck CHECK (
        primary_external_case_ref IS NULL OR (
            char_length(primary_external_case_ref) BETWEEN 1 AND 128
            AND primary_external_case_ref !~ '[[:cntrl:]]'
        )
    );

-- +goose Down
ALTER TABLE incidents
    DROP CONSTRAINT IF EXISTS incidents_primary_external_case_ref_contract_ck,
    DROP CONSTRAINT IF EXISTS incidents_current_phase_contract_ck,
    DROP CONSTRAINT IF EXISTS incidents_severity_contract_ck,
    DROP CONSTRAINT IF EXISTS incidents_description_contract_ck,
    DROP CONSTRAINT IF EXISTS incidents_tlp_ck;
