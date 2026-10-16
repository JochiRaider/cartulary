package reporting

import (
	"context"
	"fmt"
	"sort"
)

// Selection uses the admitted addresses. Rebuild only after that selection is
// complete: renumbering earlier would make a path rule select a different value.
// This projection never mutates the admitted source or copies a raw value over
// its selected retained value.
func projectRetainedSections(ctx context.Context, source []ReportingSection, fields map[string]ReportingField) ([]ReportingSection, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	sections := make([]ReportingSection, 0, len(source))
	ordered := append([]ReportingSection{}, source...)
	sort.Slice(ordered, func(i, j int) bool {
		if ordered[i].OrderingKey != ordered[j].OrderingKey {
			return ordered[i].OrderingKey < ordered[j].OrderingKey
		}
		return ordered[i].SectionID < ordered[j].SectionID
	})
	var projectBlocks func([]ReportingBlock, string, bool) ([]ReportingBlock, error)
	projectBlocks = func(source []ReportingBlock, parent string, nested bool) ([]ReportingBlock, error) {
		blocks := make([]ReportingBlock, 0, len(source))
		for _, original := range source {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			block := original
			block.Fields = []ReportingField{}
			for _, originalField := range original.Fields {
				if len(originalField.SourceRefs) == 0 {
					return nil, newRenderValidationError("export_model_invalid", "dangling_source_ref")
				}
				field, retained := fields[originalField.SourceRefs[0]]
				if !retained {
					continue
				}
				field.FieldOrdinal = len(block.Fields) + 1
				block.Fields = append(block.Fields, field)
			}
			block.BlockOrdinal = len(blocks) + 1
			block.BlockID = fmt.Sprintf("%s.b%04d", parent, block.BlockOrdinal)
			block.ParentBlockID = nil
			if nested {
				id := parent
				block.ParentBlockID = &id
			}
			children, err := projectBlocks(original.Children, block.BlockID, true)
			if err != nil {
				return nil, err
			}
			block.Children = children
			if len(block.Fields) == 0 && len(block.Children) == 0 {
				continue
			}
			refs := emptyContainerRefs()
			for _, field := range block.Fields {
				refs.include(field.SourceRefs, field.SupportRefs, field.DisclosurePartitionRefs)
			}
			for _, child := range block.Children {
				refs.include(child.SourceRefs, child.SupportRefs, child.DisclosurePartitionRefs)
			}
			refs = refs.normalized()
			block.SourceRefs, block.SupportRefs, block.DisclosurePartitionRefs = refs.source, refs.support, refs.partitions
			blocks = append(blocks, block)
		}
		return blocks, nil
	}
	for _, original := range ordered {
		if err := ctx.Err(); err != nil {
			return nil, err
		}
		section := original
		section.SectionID = "sec-" + section.OrderingKey
		blocks, err := projectBlocks(original.Blocks, section.SectionID, false)
		if err != nil {
			return nil, err
		}
		if len(blocks) == 0 && original.SectionKind != "validation" {
			continue
		}
		section.Blocks = blocks
		refs := emptyContainerRefs()
		for _, block := range blocks {
			refs.include(block.SourceRefs, block.SupportRefs, block.DisclosurePartitionRefs)
		}
		refs = refs.normalized()
		section.SourceRefs, section.SupportRefs, section.DisclosurePartitionRefs = refs.source, refs.support, refs.partitions
		section.ContentClassSummary, err = summarizeReportingBlocks(blocks)
		if err != nil {
			return nil, err
		}
		section.SectionValidation = ReportingSectionValidation{Result: "passed"}
		sections = append(sections, section)
	}
	return sections, nil
}
