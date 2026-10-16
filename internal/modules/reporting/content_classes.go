package reporting

// Source owners supply only the four Core classifications. The remaining
// Reporting categories belong to generated presentation; they cannot relabel
// source facts or bypass source eligibility, partitioning or redaction.
func validReportingContentClass(class string) bool {
	if validContentClass(class) {
		return true
	}
	switch class {
	case ContentClassPresentationText, ContentClassSupportReference, ContentClassValidation, ContentClassTemplateBoilerplate:
		return true
	default:
		return false
	}
}

func summarizeReportingBlocks(blocks []ReportingBlock) (ReportingContentClassSummary, error) {
	var summary ReportingContentClassSummary
	var visit func([]ReportingBlock) error
	visit = func(items []ReportingBlock) error {
		for _, block := range items {
			switch block.ContentClass {
			case ContentClassSourceEvidence:
				summary.SourceEvidence++
			case ContentClassCuratedNarrative:
				summary.CuratedNarrative++
			case ContentClassDerivedAnalytic:
				summary.DerivedAnalytic++
			case ContentClassWorkingMaterial:
				summary.WorkingMaterial++
			case ContentClassPresentationText:
				summary.PresentationText++
			case ContentClassSupportReference:
				summary.SupportReference++
			case ContentClassValidation:
				summary.Validation++
			case ContentClassTemplateBoilerplate:
				summary.TemplateBoilerplate++
			default:
				return newRenderValidationError("export_model_invalid", "invalid_scalar_value")
			}
			if err := visit(block.Children); err != nil {
				return err
			}
		}
		return nil
	}
	if err := visit(blocks); err != nil {
		return ReportingContentClassSummary{}, err
	}
	return summary, nil
}

func validateSnapshotBlockClasses(model SnapshotContent) error {
	issues := renderIssueCollector{stage: "export_model_materialization"}
	var visit func([]ReportingBlock)
	visit = func(blocks []ReportingBlock) {
		for _, block := range blocks {
			if block.SchemaID != ReportingBlockSchemaID {
				issues.add(newRenderValidationError("export_model_invalid", "invalid_schema_id"))
			}
			issues.add(validateSourceContentClass(block.ContentClass))
			visit(block.Children)
		}
	}
	for _, section := range model.Sections {
		if section.SchemaID != ReportingSectionSchemaID {
			issues.add(newRenderValidationError("export_model_invalid", "invalid_schema_id"))
		}
		visit(section.Blocks)
		summary, err := summarizeReportingBlocks(section.Blocks)
		// An invalid class already has a diagnostic; an incomplete summary
		// cannot establish an additional mismatch.
		if err == nil && section.ContentClassSummary != summary {
			issues.add(newRenderValidationError("export_model_invalid", "invalid_scalar_value"))
		}
	}
	return issues.finish()
}
