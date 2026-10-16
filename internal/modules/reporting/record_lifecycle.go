package reporting

// Lifecycle comes from the owning domain's frozen state, never from display
// text, a selected redaction action or a later query of mutable records.
func sourceGroupLifecycle(family string, fields []ReportingField) (string, error) {
	if family != "timeline_event" && family != "decision" {
		return "active", nil
	}
	for _, field := range fields {
		if family == "timeline_event" && field.FieldKey == "capture_state" {
			switch field.Value {
			case "rough", "enriched", "reviewed":
				return "active", nil
			case "superseded":
				return "superseded", nil
			}
		}
		if family == "decision" && field.FieldKey == "is_superseded" {
			if value, ok := field.Value.(bool); ok {
				if value {
					return "superseded", nil
				}
				return "active", nil
			}
		}
	}
	return "", newRenderValidationError("export_model_invalid", "invalid_scalar_value")
}

// Materialization marks all non-null source copies before freezing them.
// Admission validates those copies without repairing or renumbering anything.
func snapshotLifecycleFields(model *SnapshotContent, materialize bool) error {
	states := map[string]string{}
	for _, record := range model.Records {
		states[record.RecordID] = record.DeletedState
	}
	byRef := map[string]bool{}
	for _, ref := range model.SourceIndex {
		if coreTypeForSourceFamily(ref.SourceFamily) != "" || ref.SourceFamily == "record_envelope" {
			byRef[ref.SourceRefID] = states[ref.SourceRecordID] == "superseded"
		}
	}
	check := func(fields []ReportingField) error {
		for i := range fields {
			field := &fields[i]
			superseded := false
			for _, ref := range field.SourceRefs {
				superseded = superseded || byRef[ref]
			}
			if field.Value == nil {
				continue
			}
			if materialize && superseded && field.SourceValueState == "present" {
				field.SourceValueState = "superseded"
			}
			if superseded != (field.SourceValueState == "superseded") {
				return newRenderValidationError("export_model_invalid", "invalid_scalar_value")
			}
		}
		return nil
	}
	var blocks func([]ReportingBlock) error
	blocks = func(items []ReportingBlock) error {
		for _, block := range items {
			if err := check(block.Fields); err != nil {
				return err
			}
			if err := blocks(block.Children); err != nil {
				return err
			}
		}
		return nil
	}
	for _, section := range model.Sections {
		if err := blocks(section.Blocks); err != nil {
			return err
		}
	}
	for _, record := range model.Records {
		if err := check(record.Fields); err != nil {
			return err
		}
	}
	for _, event := range model.TimelineEvents {
		if err := check(event.Fields); err != nil {
			return err
		}
	}
	return nil
}
