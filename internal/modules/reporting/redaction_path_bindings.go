package reporting

// Profile shape admission is independent of model binding. An exact selector
// must resolve to an executable candidate in this admitted model; a syntactically
// valid path is not permission to ignore a rule. Duplicate views are derived
// from canonical content and never create a second selectable source value.
func validateRedactionPathBindings(paths map[string]string, view RedactionProfileView, source SnapshotContent) error {
	candidates := make(map[string]bool, len(paths))
	for _, path := range paths {
		if candidates[path] {
			return newRenderValidationError("export_model_invalid", "duplicate_export_model_path")
		}
		candidates[path] = true
	}
	tree, err := redactionTreeCandidates(source)
	if err != nil {
		return err
	}
	for _, candidate := range tree {
		if candidate.kind != "field" {
			candidates[candidate.path] = true
		}
	}
	for _, rule := range view.PathRules {
		if rule.PathSelector == nil || !candidates[*rule.PathSelector] {
			return newRenderValidationError("redaction_manifest_invalid", "redaction_action_unresolved")
		}
	}
	return nil
}
