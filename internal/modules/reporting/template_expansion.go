package reporting

import (
	"fmt"
	"slices"
	"sort"
)

type templateSectionDimension struct {
	key            *string
	label          string
	partyPartition string
	internal       bool
	sourceRefs     map[string]bool
	labelRefs      reportingContainerRefs
}

func (dimension templateSectionDimension) includes(field ReportingField) bool {
	if dimension.partyPartition != "" {
		if len(field.DisclosurePartitionRefs) == 0 {
			return false
		}
		for _, partition := range field.DisclosurePartitionRefs {
			if partition != "public" && partition != dimension.partyPartition && !(dimension.internal && partition == "internal_only") {
				return false
			}
		}
	}
	if dimension.sourceRefs != nil {
		for _, ref := range field.SourceRefs {
			if dimension.sourceRefs[ref] {
				return true
			}
		}
		return false
	}
	return true
}

// Dimension membership is independent of display text. Labels come only from
// selected tokens or retained Party fields, never the private source facts.
func templateSectionDimensions(decl TemplateSection, model RedactedExportModel) ([]templateSectionDimension, error) {
	result := []templateSectionDimension{}
	if decl.Expansion == "none" {
		return []templateSectionDimension{{}}, nil
	}
	fields := reportingFieldSources(model.SnapshotContent)
	subjectsByRef := map[string]TokenizableSubject{}
	for _, subject := range model.Subjects {
		subjectsByRef[subject.StableSubjectRef] = subject
	}
	partiesByID := map[string]ReportingRecordSummary{}
	for _, record := range model.Records {
		if record.RecordType == "party" {
			partiesByID[record.RecordID] = record
		}
	}
	recipients := map[string]bool{}
	for _, partition := range model.RecipientPartitionRefs {
		recipients[partition] = true
	}
	switch decl.Expansion {
	case "per_party":
		ids := slices.Clone(model.admittedPartyIDs)
		sort.Strings(ids)
		for _, id := range ids {
			partition := "party:" + id
			if model.ReleaseScope == ReleaseScopeExternal && !recipients[partition] {
				continue
			}
			dimension := templateSectionDimension{key: &id, partyPartition: partition, internal: model.ReleaseScope != ReleaseScopeExternal, label: fmt.Sprintf("Party %04d", len(result)+1)}
			subjectRef := canonicalSubjectRef("party", id)
			if !dimension.useToken(model.subjectDisplayTokens[subjectRef], subjectsByRef[subjectRef], fields) {
				if record, exists := partiesByID[id]; exists {
					fields := []ReportingField{}
					for _, field := range record.Fields {
						if dimension.includes(field) {
							fields = append(fields, field)
						}
					}
					if label := sourceGroupDisplayName("party", fields); label != nil && *label != "" {
						dimension.label = *label
						for _, field := range fields {
							if field.FieldKey == "display_name" {
								dimension.labelRefs.include(field.SourceRefs, field.SupportRefs, field.DisclosurePartitionRefs)
							}
						}
					}
				}
			}
			if !validReportingText(dimension.label, false) {
				dimension.label, dimension.labelRefs = fmt.Sprintf("Party %04d", len(result)+1), emptyContainerRefs()
			}
			result = append(result, dimension)
		}
	case "per_subject":
		subjects := slices.Clone(model.Subjects)
		sort.Slice(subjects, func(i, j int) bool { return subjects[i].StableSubjectRef < subjects[j].StableSubjectRef })
		for _, subject := range subjects {
			ref := subject.StableSubjectRef
			dimension := templateSectionDimension{key: &ref, label: fmt.Sprintf("Subject %04d", len(result)+1), sourceRefs: map[string]bool{}}
			for _, source := range subject.SourceRefs {
				dimension.sourceRefs[source] = true
			}
			dimension.useToken(model.subjectDisplayTokens[ref], subject, fields)
			result = append(result, dimension)
		}
	default:
		return nil, newRenderValidationError("export_model_invalid", "unsupported_derivation_algorithm")
	}
	return result, nil
}

func (dimension *templateSectionDimension) useToken(token string, subject TokenizableSubject, fields map[string]reportingFieldSource) bool {
	if token == "" {
		return false
	}
	found := false
	for _, ref := range subject.SourceRefs {
		field, exists := fields[ref]
		if !exists || field.Field.RedactedValueState != "tokenized" || field.Field.Value != token || !dimension.includes(field.Field) {
			continue
		}
		dimension.labelRefs.include(field.Field.SourceRefs, field.Field.SupportRefs, field.Field.DisclosurePartitionRefs)
		found = true
	}
	if found {
		dimension.label = token
	}
	return found
}

// Expansion multiplies retained content. Charge emitted instances, including
// repeated fields, before materializing the final tree or launching a renderer.
type templateExpansionBudget struct {
	limits                   renderLimits
	sections, blocks, fields int64
}

func (budget *templateExpansionBudget) include(section ReportingSection) error {
	budget.sections++
	var count func([]ReportingBlock)
	count = func(blocks []ReportingBlock) {
		budget.blocks += int64(len(blocks))
		for _, block := range blocks {
			budget.fields += int64(len(block.Fields))
			count(block.Children)
		}
	}
	count(section.Blocks)
	for _, item := range []struct {
		key   string
		value int64
	}{{"sections.count", budget.sections}, {"blocks.count", budget.blocks}, {"fields.count", budget.fields}} {
		if err := budget.limits.check(item.key, item.value); err != nil {
			return err
		}
	}
	return nil
}
