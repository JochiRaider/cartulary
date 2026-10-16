package reporting

import (
	"slices"
	"sort"
	"unicode/utf8"
)

func validateNarrativeSourceBindings(contract *TemplateContract, model SnapshotContent) error {
	if contract == nil {
		return nil
	}
	for _, slot := range contract.NarrativeSlots {
		for _, record := range model.Records {
			path, err := keyedModelPath("$", "records", record.RecordID)
			if err != nil {
				return err
			}
			for _, field := range record.Fields {
				address, err := keyedModelPath(path, "fields", field.FieldKey)
				if err != nil {
					return err
				}
				if templateBindingContains(slot.Binding, address) && record.RecordType != "artifact" {
					return newRenderValidationError("export_model_invalid", "invalid_scalar_value")
				}
			}
		}
	}
	return nil
}

// Narrative slots are typed projections of retained Artifact fields. No text
// arrives from the release request, and no raw snapshot value is reintroduced.
func deriveNarrativeBlocks(contract TemplateContract, decl TemplateSection, dimension templateSectionDimension, model RedactedExportModel) ([]ReportingBlock, error) {
	blocks := []ReportingBlock{}
	records := slices.Clone(model.Records)
	sort.Slice(records, func(i, j int) bool { return records[i].RecordID < records[j].RecordID })
	sources := reportingFieldSources(model.SnapshotContent)
	for _, slot := range contract.NarrativeSlots {
		if slot.SectionDeclID != decl.DeclID {
			continue
		}
		count, chars := 0, 0
		for _, record := range records {
			path, err := keyedModelPath("$", "records", record.RecordID)
			if err != nil {
				return nil, err
			}
			fields := slices.Clone(record.Fields)
			sort.Slice(fields, func(i, j int) bool { return fields[i].FieldKey < fields[j].FieldKey })
			for _, field := range fields {
				fieldPath, err := keyedModelPath(path, "fields", field.FieldKey)
				if err != nil {
					return nil, err
				}
				if !templateBindingContains(slot.Binding, fieldPath) {
					continue
				}
				if record.RecordType != "artifact" {
					return nil, newRenderValidationError("export_model_invalid", "invalid_scalar_value")
				}
				if !dimension.includes(field) || field.Value == nil {
					continue
				}
				text, ok := field.Value.(string)
				if !ok {
					return nil, newRenderValidationError("export_model_invalid", "invalid_type")
				}
				if text == "" {
					continue
				}
				if len(field.SourceRefs) == 0 {
					return nil, newRenderValidationError("export_model_invalid", "dangling_source_ref")
				}
				source, exists := sources[field.SourceRefs[0]]
				if !exists {
					return nil, newRenderValidationError("export_model_invalid", "dangling_source_ref")
				}
				length := utf8.RuneCountInString(text)
				if length > slot.MaxChars-chars {
					return nil, newRenderValidationError("export_model_invalid", "invalid_scalar_value")
				}
				chars += length
				count++
				blocks = append(blocks, ReportingBlock{SchemaID: ReportingBlockSchemaID, BlockKind: "paragraph", ContentClass: source.ContentClass,
					Fields: []ReportingField{field}, Children: []ReportingBlock{}})
			}
		}
		if count == 0 && slot.Required {
			return nil, newRenderValidationError("export_model_invalid", "narrative_slot_unbound")
		}
	}
	return blocks, nil
}
