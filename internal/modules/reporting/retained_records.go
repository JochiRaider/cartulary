package reporting

import "context"

// Core identities survive only with selected content. Display labels and
// reference aggregates are rebuilt from retained fields, never copied from the
// unredacted summary or recovered from an auxiliary source with the same ID.
func projectRetainedRecords(ctx context.Context, source []ReportingRecordSummary, fields map[string]ReportingField) ([]ReportingRecordSummary, error) {
	records := make([]ReportingRecordSummary, 0, len(source))
	for _, original := range source {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		record := original
		record.Fields = []ReportingField{}
		refs := emptyContainerRefs()
		for _, originalField := range original.Fields {
			if len(originalField.SourceRefs) == 0 {
				return nil, newRenderValidationError("export_model_invalid", "dangling_source_ref")
			}
			field, retained := fields[originalField.SourceRefs[0]]
			if !retained {
				continue
			}
			field.FieldOrdinal = len(record.Fields) + 1
			record.Fields = append(record.Fields, field)
			refs.include(field.SourceRefs, field.SupportRefs, field.DisclosurePartitionRefs)
		}
		if len(record.Fields) == 0 {
			continue
		}
		refs = refs.normalized()
		record.SourceRefs, record.SupportRefs, record.DisclosurePartitionRefs = refs.source, refs.support, refs.partitions
		record.DisplayName = sourceGroupDisplayName(record.SourceRecordRef.SourceFamily, record.Fields)
		records = append(records, record)
	}
	return records, nil
}
