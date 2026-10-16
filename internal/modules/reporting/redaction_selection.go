package reporting

// Selection and partition qualification finish before token assignment or
// output serialization. The selected action and trace remain immutable.
type selectedRedactionField struct {
	Field   ExportField
	RuleID  string
	Action  RedactionActionSpec
	Trace   SelectedRuleTrace
	Value   any
	Include bool
	Outcome string
}

func selectRedactionFields(fields []ExportField, modelPaths map[string]string, profile RedactionProfileView, scope string, effective []string) ([]selectedRedactionField, error) {
	if !profileSupportsScope(profile, scope) {
		return nil, newRenderValidationError("redaction_manifest_invalid", "redaction_profile_invalid")
	}
	selected := make([]selectedRedactionField, 0, len(fields))
	seen := map[string]bool{}
	allowed := map[string]bool{}
	declared := map[string]bool{}
	for _, ref := range profile.AllowedDisclosurePartitionRefs {
		declared[ref] = true
	}
	for _, ref := range effective {
		allowed[ref] = declared[ref]
	}
	for _, field := range fields {
		if field.Path == "" || seen[field.Path] {
			return nil, newRenderValidationError("export_model_invalid", "duplicate_export_model_path")
		}
		seen[field.Path] = true
		modelPath := modelPaths[field.Path]
		if modelPath == "" {
			return nil, newRenderValidationError("export_model_invalid", "dangling_source_ref")
		}
		ruleID, action, trace := selectRedactionRule(field, modelPath, profile)
		value, include, outcome, err := applyRedactionAction(field.Value, action)
		if err != nil {
			return nil, err
		}
		if err := validateDisclosurePartitionSet(field.DisclosurePartitionRefs); err != nil {
			return nil, newRenderValidationError("export_model_invalid", "invalid_scalar_value")
		}
		if include && !fieldDisclosureAllowed(allowed, field) {
			return nil, newRenderValidationError("redaction_manifest_invalid", "disclosure_partition_unresolved")
		}
		selected = append(selected, selectedRedactionField{Field: field, RuleID: ruleID, Action: action, Trace: trace, Value: value, Include: include, Outcome: outcome})
	}
	return selected, nil
}
