-- +goose Up
ALTER TABLE jobs
    ADD COLUMN auth_policy text;

UPDATE jobs
   SET auth_policy = CASE
       WHEN scope_kind = 'incident' THEN 'incident_membership'
       ELSE 'submitter_or_deployment_admin'
   END;

UPDATE jobs j
   SET auth_policy = 'deployment_admin'
  FROM reference_pack_job_payloads rpjp
 WHERE rpjp.job_id = j.job_id;

ALTER TABLE jobs
    ALTER COLUMN auth_policy SET NOT NULL,
    ALTER COLUMN auth_policy SET DEFAULT 'submitter_or_deployment_admin',
    ADD CONSTRAINT jobs_auth_policy_ck CHECK (
        (scope_kind = 'incident' AND auth_policy = 'incident_membership') OR
        (scope_kind = 'deployment' AND auth_policy IN ('submitter_or_deployment_admin', 'deployment_admin'))
    );

-- +goose Down
ALTER TABLE jobs
    DROP CONSTRAINT IF EXISTS jobs_auth_policy_ck,
    ALTER COLUMN auth_policy DROP DEFAULT,
    DROP COLUMN IF EXISTS auth_policy;
