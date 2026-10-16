package reporting

// Mention identity, resolution and origin are source-owned frozen facts. A
// resolved mention names the canonical Host/Identity; it is not a second
// unresolved subject with an unrelated token or an inferred Party assignment.
type reportingMentionBinding struct {
	Family, OriginRecordID, OriginFieldKey string
	ResolvedRecordID                       *string
}

func reportingMentionSource(record reportingSourceGroup) (reportingMentionBinding, error) {
	invalid := newRenderValidationError("token_manifest_invalid", "token_subject_not_stable")
	values := map[string]any{}
	for _, field := range record.Fields {
		switch field.FieldKey {
		case "entity_type", "source_record_id", "source_field_key", "resolution_status", "resolved_record_id":
			if previous, exists := values[field.FieldKey]; exists && !sameReportingValue(previous, field.Value) {
				return reportingMentionBinding{}, invalid
			}
			values[field.FieldKey] = field.Value
		}
	}
	family, _ := values["entity_type"].(string)
	origin, _ := values["source_record_id"].(string)
	key, _ := values["source_field_key"].(string)
	status, _ := values["resolution_status"].(string)
	resolved, present := values["resolved_record_id"]
	if (family != "host" && family != "identity") || !validReportingIdentifier(origin) || !validReportingIdentifier(key) || !present {
		return reportingMentionBinding{}, invalid
	}
	binding := reportingMentionBinding{Family: family, OriginRecordID: origin, OriginFieldKey: key}
	switch status {
	case "unresolved":
		if resolved != nil {
			return reportingMentionBinding{}, invalid
		}
	case "resolved":
		id, ok := resolved.(string)
		if !ok || !validReportingIdentifier(id) {
			return reportingMentionBinding{}, invalid
		}
		binding.ResolvedRecordID = &id
	default:
		return reportingMentionBinding{}, invalid
	}
	return binding, nil
}
