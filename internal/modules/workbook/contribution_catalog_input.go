package workbook

func cloneContributionCatalogInput(input ContributionCatalogInput) ContributionCatalogInput {
	cloned := input
	cloned.Queries = append([]QueryContribution(nil), input.Queries...)
	for index := range cloned.Queries {
		cloned.Queries[index].SourceRecordTypes = append(
			[]string(nil),
			input.Queries[index].SourceRecordTypes...,
		)
	}
	cloned.Creates = append([]CreateContribution(nil), input.Creates...)
	for index := range cloned.Creates {
		cloned.Creates[index].SourceRecordTypes = append(
			[]string(nil),
			input.Creates[index].SourceRecordTypes...,
		)
	}
	cloned.Patches = append([]PatchContribution(nil), input.Patches...)
	for index := range cloned.Patches {
		cloned.Patches[index].ViewSchemaIDs = append(
			[]string(nil),
			input.Patches[index].ViewSchemaIDs...,
		)
	}
	cloned.Conflicts = append([]ConflictContribution(nil), input.Conflicts...)
	for index := range cloned.Conflicts {
		cloned.Conflicts[index].ViewSchemaIDs = append(
			[]string(nil),
			input.Conflicts[index].ViewSchemaIDs...,
		)
	}
	cloned.ActionRequirements = cloneActionCapabilityRequirements(input.ActionRequirements)
	cloned.Actions = MutationActionContributions{
		Clipboard:  append([]ClipboardContribution(nil), input.Actions.Clipboard...),
		Bulk:       append([]BulkContribution(nil), input.Actions.Bulk...),
		LinkedNote: append([]LinkedNoteContribution(nil), input.Actions.LinkedNote...),
		Supersede:  append([]SupersedeContribution(nil), input.Actions.Supersede...),
	}
	return cloned
}

func cloneActionCapabilityRequirements(input ActionCapabilityRequirements) ActionCapabilityRequirements {
	return ActionCapabilityRequirements{
		ClipboardViewSchemaIDs: append([]string(nil), input.ClipboardViewSchemaIDs...),
		BulkViewSchemaIDs:      append([]string(nil), input.BulkViewSchemaIDs...),
		LinkedNoteRecordTypes:  append([]string(nil), input.LinkedNoteRecordTypes...),
		SupersedeRecordTypes:   append([]string(nil), input.SupersedeRecordTypes...),
	}
}
