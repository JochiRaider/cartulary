package reporting

// Retained structure and field copies use the same contracts as admitted
// content. Source addresses deliberately do not: they name immutable snapshot
// paths, even after the retained tree has been renumbered.
func validateRetainedProjection(model ExportModel) error {
	issues := renderIssueCollector{stage: "export_model_materialization"}
	issues.add(validateSnapshotSectionStructure(model.SnapshotContent))
	issues.add(validateSnapshotBlockClasses(model.SnapshotContent))
	issues.add(validateSnapshotFieldValues(model.SnapshotContent))
	issues.add(validateSnapshotFieldBindings(model.SnapshotContent))
	invalid := func(reason string) { issues.add(newRenderValidationError("export_model_invalid", reason)) }
	if model.Sections == nil || model.Records == nil || model.Relationships == nil || model.TimelineEvents == nil || model.Subjects == nil || model.Diagrams == nil || model.Assets == nil || model.PresentationTexts == nil || model.SourceIndex == nil || model.SupportIndex == nil || model.RecipientPartitionRefs == nil {
		invalid("invalid_scalar_value")
	}
	sources := map[string]ReportingSourceRef{}
	anchors := map[sourceOwner]bool{}
	previous := ""
	for _, ref := range model.SourceIndex {
		id, err := sourceReferenceID(model.SnapshotID, sourceOwner{ref.SourceFamily, ref.SourceRecordID}, ref.SourcePath)
		if err != nil || id != ref.SourceRefID || ref.SourceRefID <= previous || ref.SourceSnapshotID != model.SnapshotID || ref.SourceSummary != nil {
			invalid("dangling_source_ref")
		}
		previous = ref.SourceRefID
		sources[ref.SourceRefID] = ref
		if ref.SourcePath == nil {
			anchors[sourceOwner{ref.SourceFamily, ref.SourceRecordID}] = true
		}
	}
	for _, ref := range model.SourceIndex {
		if ref.SourcePath != nil && !anchors[sourceOwner{ref.SourceFamily, ref.SourceRecordID}] {
			invalid("dangling_source_ref")
		}
	}
	supports := map[string]bool{}
	previous = ""
	for _, ref := range model.SupportIndex {
		id, err := supportReferenceID(model.SnapshotID, supportTarget{ref.SupportKind, ref.SupportTargetRef})
		if err != nil || id != ref.SupportRefID || ref.SupportRefID <= previous || ref.SourceSnapshotID != model.SnapshotID || ref.SourceSummary != nil {
			issues.add(invalidSupportRef())
		}
		if owner, exists := supportSourceOwner(ref); exists {
			if ref.SourceRefID == nil {
				issues.add(invalidSupportRef())
			} else if source, found := sources[*ref.SourceRefID]; !found || source.SourcePath != nil || source.SourceFamily != owner.family || source.SourceRecordID != owner.id {
				issues.add(invalidSupportRef())
			}
		}
		previous = ref.SupportRefID
		supports[ref.SupportRefID] = false
	}
	check := func(refs, support, partitions []string) {
		if !orderedReportingRefs(refs) || !orderedReportingRefs(support) || partitions == nil || validateDisclosurePartitionSet(partitions) != nil {
			invalid("invalid_scalar_value")
		}
		for _, ref := range refs {
			if _, exists := sources[ref]; !exists {
				invalid("dangling_source_ref")
			}
		}
		for _, ref := range support {
			if _, exists := supports[ref]; !exists {
				issues.add(invalidSupportRef())
			} else {
				supports[ref] = true
			}
		}
	}
	checkFields := func(fields []ReportingField) {
		for _, field := range fields {
			check(field.SourceRefs, field.SupportRefs, field.DisclosurePartitionRefs)
		}
	}
	var blocks func([]ReportingBlock)
	blocks = func(items []ReportingBlock) {
		for _, block := range items {
			check(block.SourceRefs, block.SupportRefs, block.DisclosurePartitionRefs)
			checkFields(block.Fields)
			blocks(block.Children)
		}
	}
	for _, section := range model.Sections {
		check(section.SourceRefs, section.SupportRefs, section.DisclosurePartitionRefs)
		blocks(section.Blocks)
	}
	anchor := func(record SourceRecordRef) {
		if record.SourceRefID == nil {
			invalid("dangling_source_ref")
			return
		}
		ref, exists := sources[*record.SourceRefID]
		if !exists || ref.SourcePath != nil || ref.SourceRecordID != record.SourceRecordID || ref.SourceFamily != record.SourceFamily || record.SourceSnapshotID != model.SnapshotID {
			invalid("dangling_source_ref")
		}
	}
	for _, record := range model.Records {
		anchor(record.SourceRecordRef)
		check(record.SourceRefs, record.SupportRefs, record.DisclosurePartitionRefs)
		checkFields(record.Fields)
	}
	for _, event := range model.TimelineEvents {
		check(event.SourceRefs, event.SupportRefs, event.DisclosurePartitionRefs)
		checkFields(event.Fields)
	}
	for _, subject := range model.Subjects {
		check(subject.SourceRefs, subject.SupportRefs, subject.DisclosurePartitionRefs)
		_, _, err := subjectSourceIdentity(subject)
		issues.add(err)
	}
	for _, relation := range model.Relationships {
		anchor(relation.SrcRecordRef.SourceRecordRef)
		anchor(relation.DstRecordRef.SourceRecordRef)
		check(relation.SourceRefs, relation.SupportRefs, relation.DisclosurePartitionRefs)
	}
	for _, text := range model.PresentationTexts {
		checkFields([]ReportingField{text.Body})
	}
	for _, used := range supports {
		if !used {
			issues.add(invalidSupportRef())
		}
	}
	return issues.finish()
}
