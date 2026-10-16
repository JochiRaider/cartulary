package reporting

import (
	"fmt"
	"regexp"
	"slices"
	"sort"
)

var sectionOrderingKey = regexp.MustCompile(`^(?:[0-9]{4})(?:\.[0-9]{4})?$`)

func supportedReportingBlockKinds() []string {
	return []string{"paragraph", "bullet_list", "table", "metric", "timeline_rows", "diagram_ref", "asset_ref", "speaker_note", "overflow_summary"}
}

type reportingContainerRefs struct{ source, support, partitions []string }

func emptyContainerRefs() reportingContainerRefs {
	return reportingContainerRefs{[]string{}, []string{}, []string{}}
}

func (refs *reportingContainerRefs) include(source, support, partitions []string) {
	refs.source = append(refs.source, source...)
	refs.support = append(refs.support, support...)
	refs.partitions = append(refs.partitions, partitions...)
}

func (refs reportingContainerRefs) normalized() reportingContainerRefs {
	sort.Strings(refs.source)
	sort.Strings(refs.support)
	refs.source = slices.Compact(refs.source)
	refs.support = slices.Compact(refs.support)
	refs.partitions = orderedDisclosurePartitions(refs.partitions)
	return refs
}

func (refs reportingContainerRefs) matches(source, support, partitions []string) bool {
	return orderedReportingRefs(source) && orderedReportingRefs(support) && partitions != nil &&
		slices.Equal(refs.source, source) && slices.Equal(refs.support, support) && slices.Equal(refs.partitions, partitions)
}

// Snapshot containers are source projections. Their references are exactly the
// union of their children, and their hierarchy is already explicit. Admission
// rejects malformed history; post-redaction pruning does not repair it.
func validateSnapshotSectionStructure(model SnapshotContent) error {
	return validateSectionStructure(model, nil)
}

func validateSectionStructure(model SnapshotContent, titles map[string]reportingContainerRefs) error {
	issues := renderIssueCollector{stage: "export_model_materialization"}
	invalid := func() { issues.add(newRenderValidationError("export_model_invalid", "invalid_scalar_value")) }
	seen := map[string]bool{}
	identity := func(id string) {
		if seen[id] {
			issues.add(newRenderValidationError("export_model_invalid", "duplicate_stable_id"))
		}
		seen[id] = true
	}
	var blocks func([]ReportingBlock, string, bool) reportingContainerRefs
	blocks = func(items []ReportingBlock, parent string, nested bool) reportingContainerRefs {
		refs := emptyContainerRefs()
		if items == nil {
			invalid()
		}
		for i, block := range items {
			identity(block.BlockID)
			if block.BlockOrdinal != i+1 || block.BlockID != fmt.Sprintf("%s.b%04d", parent, i+1) ||
				!nested && block.ParentBlockID != nil || nested && (block.ParentBlockID == nil || *block.ParentBlockID != parent) {
				invalid()
			}
			if !slices.Contains(supportedReportingBlockKinds(), block.BlockKind) {
				invalid()
			}
			if block.Fields == nil {
				invalid()
			}
			owned := blocks(block.Children, block.BlockID, true)
			for _, field := range block.Fields {
				owned.include(field.SourceRefs, field.SupportRefs, field.DisclosurePartitionRefs)
			}
			owned = owned.normalized()
			if !owned.matches(block.SourceRefs, block.SupportRefs, block.DisclosurePartitionRefs) {
				invalid()
			}
			refs.include(owned.source, owned.support, owned.partitions)
		}
		return refs.normalized()
	}
	if model.Sections == nil {
		invalid()
	}
	previous := ""
	for _, section := range model.Sections {
		identity(section.SectionID)
		if !sectionOrderingKey.MatchString(section.OrderingKey) || section.OrderingKey <= previous || section.OrderingKey[:4] == "0000" || len(section.OrderingKey) == 9 && section.OrderingKey[5:] == "0000" || section.SectionID != "sec-"+section.OrderingKey || !validReportingText(section.Title, false) {
			invalid()
		}
		previous = section.OrderingKey
		switch section.SectionKind {
		case "timeline", "entity_summary", "evidence_summary", "relationship_summary", "diagram", "narrative", "appendix", "validation":
		default:
			invalid()
		}
		refs := blocks(section.Blocks, section.SectionID, false)
		title := titles[section.SectionID]
		refs.include(title.source, title.support, title.partitions)
		refs = refs.normalized()
		if !refs.matches(section.SourceRefs, section.SupportRefs, section.DisclosurePartitionRefs) || section.SectionValidation != (ReportingSectionValidation{Result: "passed"}) {
			invalid()
		}
	}
	return issues.finish()
}
