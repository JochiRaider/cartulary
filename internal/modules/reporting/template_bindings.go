package reporting

// A template may select the same admitted field through a source section,
// its canonical Core summary, or its Timeline view. These are addresses of one
// value, never three independent values or three redaction identities.
func templateFieldBindings(model SnapshotContent, canonical map[string]string) (map[string][]string, error) {
	addresses := make(map[string][]string, len(canonical))
	for ref, path := range canonical {
		addresses[ref] = []string{path}
	}
	include := func(member, id string, fields []ReportingField) error {
		parent, err := keyedModelPath("$", member, id)
		if err != nil {
			return err
		}
		for _, field := range fields {
			if len(field.SourceRefs) == 0 || canonical[field.SourceRefs[0]] == "" {
				return newRenderValidationError("export_model_invalid", "dangling_source_ref")
			}
			path, err := keyedModelPath(parent, "fields", field.FieldKey)
			if err != nil {
				return err
			}
			addresses[field.SourceRefs[0]] = append(addresses[field.SourceRefs[0]], path)
		}
		return nil
	}
	for _, record := range model.Records {
		if err := include("records", record.RecordID, record.Fields); err != nil {
			return nil, err
		}
	}
	for _, event := range model.TimelineEvents {
		if err := include("timeline_events", event.RecordID, event.Fields); err != nil {
			return nil, err
		}
	}
	return addresses, nil
}

func templateBindingsSelect(bindings, addresses []string) bool {
	for _, binding := range bindings {
		for _, address := range addresses {
			if templateBindingContains(binding, address) {
				return true
			}
		}
	}
	return false
}

// This is a typed field-selection grammar, not a general JSONPath evaluator.
// Metadata scalars, field-value traversal and unrelated collections cannot
// accidentally become empty optional selections.
func supportedTemplateFieldBinding(binding string) bool {
	steps, ok := parseModelPath(binding, true)
	if !ok || len(steps) == 0 || steps[0].Key {
		return false
	}
	switch steps[0].Value {
	case "assets":
		return len(steps) == 2 && steps[1].Key
	case "records", "timeline_events":
		return len(steps) == 1 || len(steps) >= 2 && steps[1].Key && (len(steps) == 2 || !steps[2].Key && steps[2].Value == "fields" && (len(steps) == 3 || len(steps) == 4 && steps[3].Key))
	case "sections":
		if len(steps) == 1 {
			return true
		}
		if !steps[1].Key {
			return false
		}
		for i := 2; i < len(steps); i += 2 {
			member := steps[i]
			if member.Key {
				return false
			}
			if i == 2 && member.Value != "blocks" || i > 2 && member.Value != "children" && member.Value != "fields" {
				return false
			}
			if i == len(steps)-1 {
				return true
			}
			if !steps[i+1].Key {
				return false
			}
			if member.Value == "fields" {
				return i+2 == len(steps)
			}
		}
		return true
	default:
		return false
	}
}
