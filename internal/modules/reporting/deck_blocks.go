package reporting

// A deck carries the retained block schema. Copying the tree separates mutable
// chunk/overlay state from the immutable source, including array field values.
func copyDeckBlocks(source []ReportingBlock) ([]ReportingBlock, error) {
	blocks := make([]ReportingBlock, 0, len(source))
	for i, original := range source {
		block := original
		block.BlockOrdinal = i + 1
		block.ParentBlockID = cloneStringPtr(original.ParentBlockID)
		block.SourceRefs = append([]string{}, original.SourceRefs...)
		block.SupportRefs = append([]string{}, original.SupportRefs...)
		block.DisclosurePartitionRefs = append([]string{}, original.DisclosurePartitionRefs...)
		block.Fields = make([]ReportingField, 0, len(original.Fields))
		for _, field := range original.Fields {
			value, valid := admitReportingFieldValue(field.Value, true)
			if !valid {
				return nil, newRenderValidationError("slidev_source_invalid", "invalid_field_value_for_text")
			}
			field.Value = value
			field.DisplayLabel = cloneStringPtr(field.DisplayLabel)
			field.RawValueSHA256 = cloneStringPtr(field.RawValueSHA256)
			field.SourceRefs = append([]string{}, field.SourceRefs...)
			field.SupportRefs = append([]string{}, field.SupportRefs...)
			field.DisclosurePartitionRefs = append([]string{}, field.DisclosurePartitionRefs...)
			block.Fields = append(block.Fields, field)
		}
		children, err := copyDeckBlocks(original.Children)
		if err != nil {
			return nil, err
		}
		block.Children = children
		blocks = append(blocks, block)
	}
	return blocks, nil
}
