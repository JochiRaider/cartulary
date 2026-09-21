package historytest

const RecordID = "20000000-0000-4000-8000-000000000001"

// Authored canonical retained facts, including nullable columns. These fixtures
// never infer or repair snapshots supplied to the running application.
func Source(table string) map[string]any {
	tables := map[string]map[string]any{
		"artifact_findings":              {"closed_at": nil, "confidence_score": nil, "created_at": "2026-09-21T00:00:00Z", "incident_id": RecordID, "kind": "public", "owner_user_id": RecordID, "record_id": RecordID, "state": "public", "statement": "public", "updated_at": "2026-09-21T00:00:00Z"},
		"artifact_forensic_keywords":     {"case_sensitive": false, "created_at": "2026-09-21T00:00:00Z", "incident_id": RecordID, "keyword_id": "public", "match_mode": "public", "pattern": "public", "reason": "public", "record_id": RecordID, "updated_at": "2026-09-21T00:00:00Z"},
		"artifact_investigative_queries": {"created_at": "2026-09-21T00:00:00Z", "created_by_user_id": RecordID, "incident_id": RecordID, "platform": "public", "purpose": "public", "query_id": "public", "query_text": "public", "record_id": RecordID, "updated_at": "2026-09-21T00:00:00Z"},
		"artifacts":                      {"acknowledged_at": nil, "active_risks_summary": nil, "artifact_type": "public", "audience": nil, "body": nil, "channel_or_meeting": nil, "closure_state": nil, "comm_id": nil, "comm_type": nil, "created_at": "2026-09-21T00:00:00Z", "created_by_user_id": nil, "current_state_summary": nil, "handoff_id": nil, "incident_id": RecordID, "incoming_owner_user_id": nil, "lesson_id": nil, "next_checks": nil, "next_report_at": nil, "outgoing_owner_user_id": nil, "owner_user_id": nil, "privilege_tag": nil, "record_id": RecordID, "review_owner_user_id": nil, "status_review_id": nil, "summary": nil, "timestamp_utc": nil, "title": nil, "updated_at": "2026-09-21T00:00:00Z"},
		"assessments":                    {"assessed_at": "2026-09-21T00:00:00Z", "assessment_state": "public", "assessor_user_id": RecordID, "confidence_score": nil, "created_at": "2026-09-21T00:00:00Z", "deleted_at": nil, "deleted_by_user_id": nil, "incident_id": RecordID, "rationale": "public", "record_id": RecordID, "subject_record_id": RecordID, "subject_type": "public", "updated_at": "2026-09-21T00:00:00Z"},
		"decisions":                      {"created_at": "2026-09-21T00:00:00Z", "decided_at": nil, "decision_type": nil, "incident_id": RecordID, "owner_user_id": nil, "rationale": nil, "record_id": RecordID, "status": "public", "summary": nil, "updated_at": "2026-09-21T00:00:00Z"},
		"entity_aliases":                 {"classification": "public", "created_at": "2026-09-21T00:00:00Z", "created_by_user_id": RecordID, "deleted_at": nil, "entity_alias_id": RecordID, "entity_type": "public", "incident_id": RecordID, "normalized_text": "public", "raw_text": "public", "record_id": RecordID},
		"entity_mentions":                {"created_at": "2026-09-21T00:00:00Z", "created_by_user_id": RecordID, "entity_mention_id": RecordID, "entity_type": "public", "normalized_text": "public", "ordinal": 0, "origin_kind": "public", "origin_locator": "public", "raw_text": "public", "resolution_method": nil, "resolution_status": "public", "resolved_at": nil, "resolved_by_user_id": nil, "resolved_record_id": nil, "row_version": 1, "source_field_key": "public", "source_record_id": RecordID},
		"entity_preserved_identifiers":   {"classification": "public", "created_at": "2026-09-21T00:00:00Z", "created_by_user_id": RecordID, "deleted_at": nil, "entity_preserved_identifier_id": RecordID, "entity_type": "public", "identifier_type": "public", "incident_id": RecordID, "normalized_value": "public", "raw_value": "public", "record_id": RecordID},
		"evidence":                       {"blob_hash": nil, "collector_party_id": nil, "collector_party_text": nil, "created_at": "2026-09-21T00:00:00Z", "incident_id": RecordID, "lifecycle_state": "public", "object_blob_id": nil, "received_at": nil, "record_id": RecordID, "requested_at": nil, "source_party_id": nil, "source_party_text": nil, "storage_ref": nil, "title": nil, "updated_at": "2026-09-21T00:00:00Z", "upload_state": "public"},
		"hosts":                          {"aad_device_id": nil, "business_owner": nil, "containment_status": nil, "created_at": "2026-09-21T00:00:00Z", "created_by_user_id": RecordID, "criticality": nil, "display_name": "public", "entity_origin": "public", "fqdn": nil, "host_state": "public", "hostname": nil, "incident_id": RecordID, "location": nil, "merged_into_record_id": nil, "os_platform": nil, "record_id": RecordID, "row_version": 1, "seed_entity_mention_id": nil, "updated_at": "2026-09-21T00:00:00Z", "updated_by_user_id": RecordID},
		"identities":                     {"aad_object_id": nil, "created_at": "2026-09-21T00:00:00Z", "created_by_user_id": RecordID, "display_name": "public", "email": nil, "entity_origin": "public", "identity_state": "public", "incident_id": RecordID, "merged_into_record_id": nil, "mfa_state": nil, "privilege_level": nil, "record_id": RecordID, "reset_status": nil, "row_version": 1, "sam_account_name": nil, "seed_entity_mention_id": nil, "sid": nil, "updated_at": "2026-09-21T00:00:00Z", "updated_by_user_id": RecordID, "upn": nil},
		"indicator_observations":         {"created_at": "2026-09-21T00:00:00Z", "created_by_user_id": RecordID, "deleted_at": nil, "deleted_by_user_id": nil, "incident_id": RecordID, "indicator_observation_id": RecordID, "normalized_candidate": nil, "observed_text": "public", "origin_kind": "public", "origin_locator": "public", "parsed_indicator_type": nil, "resolution_method": nil, "resolution_status": "public", "resolved_at": nil, "resolved_by_user_id": nil, "resolved_indicator_record_id": nil, "row_version": 1, "source_field_key": "public", "source_record_id": RecordID},
		"indicator_state_intervals":      {"assessed_at": "2026-09-21T00:00:00Z", "assessor": nil, "confidence": nil, "created_at": "2026-09-21T00:00:00Z", "created_by_user_id": RecordID, "deleted_at": nil, "deleted_by_user_id": nil, "incident_id": RecordID, "indicator_record_id": RecordID, "indicator_state_interval_id": RecordID, "lifecycle_state": "public", "rationale": nil, "row_version": 1, "support_refs": []string{}, "valid_from": "2026-09-21T00:00:00Z", "valid_to": nil},
		"indicators":                     {"dedupe_key": "public", "defanged_value": nil, "display_value": "public", "hash_algorithm": nil, "hash_value": nil, "incident_id": RecordID, "indicator_type": "public", "normalized_value": nil, "record_id": RecordID, "stix_pattern": nil, "value_kind": "public"},
		"parties":                        {"created_at": "2026-09-21T00:00:00Z", "display_name": nil, "external_ref": nil, "incident_id": RecordID, "notes": nil, "organization_name": nil, "party_kind": nil, "primary_email": nil, "record_id": RecordID, "role_title": nil, "timezone_name": nil, "updated_at": "2026-09-21T00:00:00Z"},
		"task_requests":                  {"blocked_reason": nil, "closure_summary": nil, "completed_at": nil, "created_at": "2026-09-21T00:00:00Z", "decision_record_id": nil, "due_at": nil, "external_ticket_ref": nil, "incident_id": RecordID, "owner_user_id": nil, "priority": nil, "record_id": RecordID, "requester_party_id": nil, "requester_party_text": nil, "status": "public", "task_kind": nil, "title": nil, "updated_at": "2026-09-21T00:00:00Z", "workstream": nil},
		"timeline_events":                {"activity_local_generated": false, "activity_local_text": nil, "activity_synopsis_text": nil, "activity_time_pair_state": "public", "activity_utc_generated": false, "activity_utc_text": nil, "analyst_text": nil, "capture_state": "public", "created_by_user_id": RecordID, "data_source_text": nil, "date_entered_text": nil, "device_object_text": nil, "edited_at": "2026-09-21T00:00:00Z", "incident_id": RecordID, "ip_address_text": nil, "mitre_stage_text": nil, "raw_activity_text": nil, "record_id": RecordID, "recorded_at": "2026-09-21T00:00:00Z", "reviewed_at": nil, "reviewed_by_user_id": nil, "row_version": 1, "superseded_at": nil, "superseded_by_user_id": nil, "updated_by_user_id": RecordID},
	}
	value := tables[table]
	value["upload_token"] = "private-upload-token"
	value["credentials"] = "private-credential"
	return value
}

func Variants(recordType string) []string {
	if recordType == "artifact" {
		return []string{"comm_log", "finding", "forensic_keyword", "handoff", "investigative_query", "lesson", "note", "status_review"}
	}
	return []string{""}
}
func Snapshot(recordType, schemaID, variant string) map[string]any {
	tables := map[string]string{"host": "hosts", "identity": "identities", "timeline_event": "timeline_events", "artifact": "artifacts", "assessment": "assessments", "evidence": "evidence", "party": "parties", "task_request": "task_requests", "decision": "decisions", "indicator": "indicators"}
	source := Source(tables[recordType])
	switch recordType {
	case "artifact":
		source["artifact_type"] = variant
		for _, table := range []string{"artifact_findings", "artifact_forensic_keywords", "artifact_investigative_queries"} {
			for key, value := range Source(table) {
				if _, exists := source[key]; !exists {
					source[key] = value
				}
			}
		}
	case "timeline_event":
		source["capture_state"] = "rough"
	case "host":
		source["host_state"] = "active"
	case "identity":
		source["identity_state"] = "active"
	case "evidence":
		source["lifecycle_state"] = "requested"
		source["upload_state"] = "pending"
	}
	return map[string]any{"snapshot_schema_id": schemaID, "record": map[string]any{"record_id": RecordID, "record_type": recordType}, "source": source}
}

func NonRowFacts(kind string) map[string]any {
	var source map[string]any
	switch kind {
	case "record_link":
		source = map[string]any{"record_link_id": RecordID, "src_record_id": RecordID, "dst_record_id": "20000000-0000-4000-8000-000000000002", "link_type": "attached_evidence", "provenance": "manual", "field_key": nil, "confidence": nil, "incident_id": RecordID, "owner_user_id": RecordID, "created_by_user_id": RecordID, "created_at": "2026-09-21T00:00:00Z", "decided_at": "2026-09-21T00:00:00Z", "deleted_at": nil, "deleted_by_user_id": nil}
	case "record_tag":
		source = map[string]any{"record_tag_id": RecordID, "record_id": RecordID, "tag_name": "Important", "normalized_tag_name": "important", "incident_id": RecordID, "created_by_user_id": RecordID, "created_at": "2026-09-21T00:00:00Z", "updated_at": "2026-09-21T00:00:00Z", "deleted_at": nil, "deleted_by_user_id": nil}
	default:
		source = Source(map[string]string{"entity_mention": "entity_mentions", "entity_alias": "entity_aliases", "entity_preserved_identifier": "entity_preserved_identifiers", "indicator_observation": "indicator_observations", "indicator_state_interval": "indicator_state_intervals"}[kind])
	}
	if kind == "entity_mention" {
		source["entity_type"] = "host"
		source["resolution_status"] = "unresolved"
	}
	if kind == "indicator_observation" {
		source["resolution_status"] = "unresolved"
	}
	return source
}
