-- +goose Up
ALTER TABLE jobs
    DROP CONSTRAINT IF EXISTS jobs_auth_policy_ck;

UPDATE jobs j
   SET scope_kind = 'incident',
       incident_id = p.incident_id,
       auth_policy = 'deployment_admin_incident_membership'
  FROM incident_bundle_job_payloads p
 WHERE p.job_id = j.job_id
   AND p.job_kind = 'export'
   AND p.incident_id IS NOT NULL
   AND j.scope_kind = 'deployment'
   AND j.auth_policy = 'deployment_admin';

ALTER TABLE jobs
    ADD CONSTRAINT jobs_auth_policy_ck CHECK (
        (scope_kind = 'incident' AND auth_policy IN ('incident_membership', 'deployment_admin_incident_membership')) OR
        (scope_kind = 'deployment' AND auth_policy IN ('submitter_or_deployment_admin', 'deployment_admin'))
    );

-- +goose Down
ALTER TABLE jobs
    DROP CONSTRAINT IF EXISTS jobs_auth_policy_ck;

UPDATE jobs j
   SET scope_kind = 'deployment',
       incident_id = NULL,
       auth_policy = 'deployment_admin'
  FROM incident_bundle_job_payloads p
 WHERE p.job_id = j.job_id
   AND p.job_kind = 'export'
   AND j.scope_kind = 'incident'
   AND j.auth_policy = 'deployment_admin_incident_membership';

ALTER TABLE jobs
    ADD CONSTRAINT jobs_auth_policy_ck CHECK (
        (scope_kind = 'incident' AND auth_policy = 'incident_membership') OR
        (scope_kind = 'deployment' AND auth_policy IN ('submitter_or_deployment_admin', 'deployment_admin'))
    );
