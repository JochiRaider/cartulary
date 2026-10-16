package reporting

// The source state is frozen independently of the selected action. In
// particular a mask cannot turn an absent source into a present source merely
// because the renderer can display its replacement text.
func retainedReportingField(source ReportingField, value any, outcome, precedence string, multiline bool) (ReportingField, error) {
	value, valid := admitReportingFieldValue(value, multiline)
	if !valid {
		return ReportingField{}, newRenderValidationError("redaction_manifest_invalid", "redaction_action_unresolved")
	}
	retained := source
	retained.Value = value
	retained.SourceRefs = append([]string{}, source.SourceRefs...)
	retained.SupportRefs = append([]string{}, source.SupportRefs...)
	retained.DisclosurePartitionRefs = append([]string{}, source.DisclosurePartitionRefs...)
	if source.DisplayLabel != nil {
		label := *source.DisplayLabel
		retained.DisplayLabel = &label
	}
	retained.RedactedValueState = outcome
	if outcome == "unchanged" && precedence != "profile_default" {
		retained.RedactedValueState = "allowed"
	}
	if !validReportingField(retained, retained.FieldOrdinal, multiline) {
		return ReportingField{}, newRenderValidationError("redaction_manifest_invalid", "redaction_action_unresolved")
	}
	return retained, nil
}

type reportingFieldSource struct {
	Field        ReportingField
	Multiline    bool
	ContentClass string
}

// Call only after snapshot binding validation. The lookup uses the explicit
// source identity shared by the section, record and event projections.
func reportingFieldSources(content SnapshotContent) map[string]reportingFieldSource {
	fields := map[string]reportingFieldSource{}
	var visit func([]ReportingBlock)
	visit = func(blocks []ReportingBlock) {
		for _, block := range blocks {
			for _, field := range block.Fields {
				for _, source := range field.SourceRefs {
					fields[source] = reportingFieldSource{Field: field, Multiline: block.BlockKind == "paragraph", ContentClass: block.ContentClass}
				}
			}
			visit(block.Children)
		}
	}
	for _, section := range content.Sections {
		visit(section.Blocks)
	}
	return fields
}
