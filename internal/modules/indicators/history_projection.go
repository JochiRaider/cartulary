package indicators

import "github.com/JochiRaider/cartulary/internal/modules/revisions/historycontract"

func projectObservationHistory(facts historycontract.Facts) ([]historycontract.Unit, error) {
	fields := historycontract.Fields("indicator_observation", "text", "indicator_observation_id source_record_id source_field_key origin_kind origin_locator observed_text parsed_indicator_type normalized_candidate resolution_status resolved_indicator_record_id resolved_by_user_id resolved_at resolution_method deleted_at")
	for i := range fields {
		switch fields[i].Member {
		case "indicator_observation_id", "source_record_id", "source_field_key", "observed_text", "resolution_status":
			fields[i].Required = true
		}
	}
	for _, value := range []map[string]any{facts.Before, facts.After} {
		if value == nil {
			continue
		}
		switch value["resolution_status"] {
		case "unresolved", "dismissed":
			if value["resolved_indicator_record_id"] != nil {
				return nil, historycontract.ErrInvalidFacts
			}
		case "resolved":
			if value["resolved_indicator_record_id"] == nil {
				return nil, historycontract.ErrInvalidFacts
			}
		default:
			return nil, historycontract.ErrInvalidFacts
		}
	}
	return historycontract.Collection(facts, "indicator_observation", fields, []string{"indicator_observation_id"}, []string{"source_record_id", "resolved_indicator_record_id"})
}

func projectIntervalHistory(facts historycontract.Facts) ([]historycontract.Unit, error) {
	fields := historycontract.Fields("indicator_interval", "text", "indicator_state_interval_id indicator_record_id lifecycle_state valid_from valid_to rationale assessor assessed_at deleted_at")
	fields = append(fields, historycontract.Fields("indicator_interval", "number", "confidence")...)
	fields = append(fields, historycontract.Fields("indicator_interval", "strings", "support_refs")...)
	for i := range fields {
		switch fields[i].Member {
		case "indicator_state_interval_id", "indicator_record_id", "lifecycle_state", "valid_from":
			fields[i].Required = true
		}
	}
	return historycontract.Collection(facts, "indicator_interval", fields, []string{"indicator_state_interval_id"}, []string{"indicator_record_id"})
}

func projectRecordHistory(facts historycontract.Facts) ([]historycontract.Unit, error) {
	fields := historycontract.Fields("indicator", "text", "indicator_type value_kind display_value normalized_value defanged_value hash_algorithm hash_value stix_pattern")
	for i := range fields {
		if fields[i].Member == "indicator_type" || fields[i].Member == "value_kind" || fields[i].Member == "display_value" {
			fields[i].Required = true
		}
	}
	return historycontract.Row(facts, fields)
}
