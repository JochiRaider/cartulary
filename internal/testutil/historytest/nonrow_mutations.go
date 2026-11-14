package historytest

import (
	"encoding/json"

	"github.com/JochiRaider/cartulary/internal/modules/revisions"
)

// NonRowMutation authors one valid source-owner mutation for projection tests.
// It does not seed current source rows or claim import/rollback eligibility.
func NonRowMutation(kind, recordID, incidentID, actorID, physicalID string) revisions.StoredMutation {
	const clock = "2026-09-21T00:00:00Z"
	m := revisions.StoredMutation{TargetKind: kind, TargetID: physicalID, OperationKind: "create"}
	switch kind {
	case "entity_mention":
		m.OperationKind = "patch"
		m.BeforeValue = map[string]any{
			"entity_mention_id": physicalID, "source_record_id": recordID, "entity_type": "host", "source_field_key": "timeline.device_object_text", "origin_kind": "manual_entry", "origin_locator": "span:0:4", "raw_text": "host", "normalized_text": "host", "resolution_status": "unresolved", "row_version": json.Number("1"), "ordinal": json.Number("1"), "created_by_user_id": actorID, "created_at": clock, "resolved_record_id": nil, "resolved_by_user_id": nil, "resolved_at": nil, "resolution_method": nil,
		}
		m.AfterValue = map[string]any{
			"entity_mention_id": physicalID, "source_record_id": recordID, "entity_type": "host", "source_field_key": "timeline.device_object_text", "origin_kind": "manual_entry", "origin_locator": "span:0:4", "raw_text": "host", "normalized_text": "host", "resolution_status": "dismissed", "row_version": json.Number("2"), "ordinal": json.Number("1"), "created_by_user_id": actorID, "created_at": clock, "resolved_record_id": nil, "resolved_by_user_id": nil, "resolved_at": nil, "resolution_method": nil,
		}
	case "entity_alias":
		m.TargetID = "entity_alias:" + physicalID
		m.AfterValue = map[string]any{"entity_alias_id": physicalID, "incident_id": incidentID, "record_id": recordID, "entity_type": "host", "raw_text": "Public alias", "normalized_text": "Public alias", "classification": "suggestion_only", "created_by_user_id": actorID, "created_at": clock, "deleted_at": nil}
	case "entity_preserved_identifier":
		m.TargetID = "entity_preserved_identifier:" + physicalID
		m.AfterValue = map[string]any{"entity_preserved_identifier_id": physicalID, "incident_id": incidentID, "record_id": recordID, "entity_type": "host", "identifier_type": "hostname", "raw_value": "public-host", "normalized_value": "public-host", "classification": "provenance_only", "created_by_user_id": actorID, "created_at": clock, "deleted_at": nil}
	case "indicator_observation":
		m.AfterValue = map[string]any{"indicator_observation_id": physicalID, "incident_id": incidentID, "source_record_id": recordID, "source_field_key": "timeline.raw_activity_text", "origin_kind": "manual_entry", "origin_locator": "span:0:6", "observed_text": "public", "parsed_indicator_type": nil, "normalized_candidate": nil, "resolution_status": "unresolved", "resolved_indicator_record_id": nil, "row_version": json.Number("1"), "created_by_user_id": actorID, "created_at": clock, "resolved_by_user_id": nil, "resolved_at": nil, "resolution_method": nil, "deleted_at": nil, "deleted_by_user_id": nil}
	case "indicator_state_interval":
		m.AfterValue = map[string]any{"indicator_state_interval_id": physicalID, "incident_id": incidentID, "indicator_record_id": recordID, "lifecycle_state": "active", "valid_from": clock, "valid_to": nil, "confidence": nil, "rationale": nil, "support_refs": []string{}, "assessor": nil, "assessed_at": clock, "row_version": json.Number("1"), "created_by_user_id": actorID, "created_at": clock, "deleted_at": nil, "deleted_by_user_id": nil}
	case "record_link":
		m.AfterValue = map[string]any{"record_link_id": physicalID, "src_record_id": recordID, "dst_record_id": "20000000-0000-4000-8000-000000000002", "link_type": "attached_evidence", "provenance": "manual", "field_key": nil, "confidence": nil, "incident_id": incidentID, "owner_user_id": actorID, "created_by_user_id": actorID, "created_at": clock, "decided_at": clock, "deleted_at": nil, "deleted_by_user_id": nil}
	case "record_tag":
		m.TargetID = "record_tag:" + recordID + ":" + physicalID
		m.AfterValue = map[string]any{"record_tag_id": physicalID, "record_id": recordID, "tag_name": "Important", "normalized_tag_name": "important", "incident_id": incidentID, "created_by_user_id": actorID, "created_at": clock, "updated_at": clock, "deleted_at": nil, "deleted_by_user_id": nil}
	default:
		panic("unknown non-row history fixture")
	}
	return m
}
