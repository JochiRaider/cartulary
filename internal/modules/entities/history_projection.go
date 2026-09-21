package entities

import "github.com/JochiRaider/cartulary/internal/modules/revisions/historycontract"

func projectMentionHistory(facts historycontract.Facts) ([]historycontract.Unit, error) {
	fields := historycontract.Fields("mention", "text", "entity_mention_id source_record_id entity_type source_field_key origin_kind origin_locator raw_text normalized_text resolution_status resolved_record_id resolved_by_user_id resolved_at resolution_method")
	fields = append(fields, historycontract.Fields("mention", "number", "ordinal")...)
	for i := range fields {
		switch fields[i].Member {
		case "entity_mention_id", "source_record_id", "entity_type", "source_field_key", "raw_text", "resolution_status":
			fields[i].Required = true
		}
	}
	for _, value := range []map[string]any{facts.Before, facts.After} {
		if value == nil {
			continue
		}
		if value["entity_type"] != "host" && value["entity_type"] != "identity" {
			return nil, historycontract.ErrInvalidFacts
		}
		switch value["resolution_status"] {
		case "unresolved", "dismissed":
			if value["resolved_record_id"] != nil {
				return nil, historycontract.ErrInvalidFacts
			}
		case "resolved":
			if value["resolved_record_id"] == nil {
				return nil, historycontract.ErrInvalidFacts
			}
		default:
			return nil, historycontract.ErrInvalidFacts
		}
	}
	return historycontract.Collection(facts, "mention", fields, []string{"entity_mention_id"}, []string{"source_record_id", "resolved_record_id"})
}

func projectAliasHistory(facts historycontract.Facts) ([]historycontract.Unit, error) {
	fields := historycontract.Fields("entity_alias", "text", "entity_alias_id record_id entity_type raw_text normalized_text classification deleted_at")
	for i := range fields {
		if fields[i].Member != "deleted_at" && fields[i].Member != "entity_alias_id" {
			fields[i].Required = true
		}
	}
	return historycontract.Collection(facts, "entity_identifier", fields, []string{"record_id", "entity_type", "normalized_text"}, []string{"record_id"})
}

func projectPreservedIdentifierHistory(facts historycontract.Facts) ([]historycontract.Unit, error) {
	fields := historycontract.Fields("entity_identifier", "text", "record_id entity_type identifier_type raw_value normalized_value classification deleted_at")
	for i := range fields {
		if fields[i].Member != "deleted_at" {
			fields[i].Required = true
		}
	}
	return historycontract.Collection(facts, "entity_identifier", fields, []string{"record_id", "entity_type", "identifier_type", "normalized_value", "classification"}, []string{"record_id"})
}

func projectHostHistory(facts historycontract.Facts) ([]historycontract.Unit, error) {
	fields := historycontract.Fields("host", "text", "display_name hostname aad_device_id fqdn entity_origin host_state merged_into_record_id location os_platform business_owner criticality containment_status")
	for i := range fields {
		if fields[i].Member == "merged_into_record_id" {
			fields[i].Kind = "record"
		}
		if fields[i].Member == "display_name" {
			fields[i].Required = true
		}
	}
	return historycontract.Row(facts, fields)
}
func projectIdentityHistory(facts historycontract.Facts) ([]historycontract.Unit, error) {
	fields := historycontract.Fields("identity", "text", "display_name upn email sam_account_name aad_object_id sid entity_origin identity_state merged_into_record_id privilege_level mfa_state reset_status")
	for i := range fields {
		if fields[i].Member == "merged_into_record_id" {
			fields[i].Kind = "record"
		}
		if fields[i].Member == "display_name" {
			fields[i].Required = true
		}
	}
	return historycontract.Row(facts, fields)
}
