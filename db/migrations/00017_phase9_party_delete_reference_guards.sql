-- +goose Up
CREATE INDEX IF NOT EXISTS evidence_collector_party_ref_lookup_idx
    ON evidence (incident_id, collector_party_id)
    WHERE collector_party_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS evidence_source_party_ref_lookup_idx
    ON evidence (incident_id, source_party_id)
    WHERE source_party_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS task_requests_requester_party_ref_lookup_idx
    ON task_requests (incident_id, requester_party_id)
    WHERE requester_party_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS record_links_active_party_ref_dst_lookup_idx
    ON record_links (incident_id, dst_record_id, link_type, field_key)
    WHERE deleted_at IS NULL
      AND field_key IN ('comm_log.audience_party_ids', 'comm_log.attendee_party_ids');

-- +goose Down
DROP INDEX IF EXISTS record_links_active_party_ref_dst_lookup_idx;
DROP INDEX IF EXISTS task_requests_requester_party_ref_lookup_idx;
DROP INDEX IF EXISTS evidence_source_party_ref_lookup_idx;
DROP INDEX IF EXISTS evidence_collector_party_ref_lookup_idx;
