-- +goose Up
ALTER TABLE evidence_access_handles
    ADD COLUMN record_row_version bigint;

UPDATE evidence_access_handles h
   SET record_row_version = r.row_version
  FROM records r
 WHERE r.incident_id = h.incident_id
   AND r.record_id = h.record_id;

ALTER TABLE evidence_access_handles
    ALTER COLUMN record_row_version SET NOT NULL;

-- +goose Down
ALTER TABLE evidence_access_handles
    DROP COLUMN IF EXISTS record_row_version;
