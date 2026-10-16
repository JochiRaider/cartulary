package reporting

import "github.com/JochiRaider/cartulary/internal/modules/reporting/exportprovider"

type lifecycleRestriction uint8

const (
	lifecycleActive lifecycleRestriction = iota
	lifecycleSuperseded
	lifecycleDeleted
)

func validateDeletedRecordBindings(facts exportprovider.SourceFacts, input renderSelectionInput) error {
	deleted := map[string]bool{}
	for _, record := range facts.DeletedRecords {
		deleted[record.RecordID] = true
	}
	forced := func(path string) bool {
		steps, ok := parseModelPath(path, true)
		return ok && len(steps) >= 2 && !steps[0].Key && (steps[0].Value == "records" || steps[0].Value == "timeline_events") &&
			steps[1].Key && !steps[1].Wildcard && deleted[steps[1].Value]
	}
	if input.template != nil {
		for _, slot := range input.template.NarrativeSlots {
			if forced(slot.Binding) {
				return newRenderValidationError("export_model_invalid", "deleted_record_not_releasable")
			}
		}
		for _, section := range input.template.Sections {
			for _, binding := range section.Bindings {
				if forced(binding) {
					return newRenderValidationError("export_model_invalid", "deleted_record_not_releasable")
				}
			}
			for _, diagram := range section.DiagramDecls {
				for _, id := range diagram.SelectionRule.TimelineRecordIDs {
					if deleted[id] {
						return newRenderValidationError("export_model_invalid", "deleted_record_not_releasable")
					}
				}
			}
		}
	}
	if input.composition != nil {
		for _, diagram := range input.composition.DiagramDecls {
			for _, id := range diagram.SelectionRule.TimelineRecordIDs {
				if deleted[id] {
					return newRenderValidationError("export_model_invalid", "deleted_record_not_releasable")
				}
			}
		}
		for _, op := range input.composition.DeckOps {
			anchor := op.Payload.BlockAnchor
			if op.OpKind != "exclude_block" && anchor != nil && anchor.RecordAnchor != nil && deleted[anchor.RecordAnchor.RecordID] {
				return newRenderValidationError("export_model_invalid", "deleted_record_not_releasable")
			}
		}
	}
	return nil
}

// A standalone composition diagram has no template opt-in. Template-owned
// diagrams inherit their declaration; inserted diagrams must also satisfy every
// explicit placement. An opt-in elsewhere in the deck cannot grant permission.
func diagramAllowsSuperseded(contract TemplateContract, composition *renderComposition, id string) bool {
	allowed, bound := true, false
	sections := map[string]bool{}
	for _, section := range contract.Sections {
		sections[section.DeclID] = section.IncludeSuperseded
		for _, diagram := range section.DiagramDecls {
			if diagram.DeclID == id {
				bound, allowed = true, allowed && section.IncludeSuperseded
			}
		}
	}
	if composition != nil {
		for _, op := range composition.DeckOps {
			if op.OpKind != "insert_diagram_slide" || op.Payload.DiagramAnchor == nil || op.Payload.DiagramAnchor.DeclarationID != id {
				continue
			}
			bound = true
			allowed = allowed && op.Payload.SectionAnchor != nil && sections[op.Payload.SectionAnchor.DeclarationID]
		}
	}
	return bound && allowed
}

func eligibleDiagramEvents(model RedactedExportModel, includeSuperseded bool) []ReportingTimelineEvent {
	events := []ReportingTimelineEvent{}
	for _, event := range model.TimelineEvents {
		allowed := true
		for _, ref := range event.SourceRefs {
			state := model.lifecycleRestrictions[ref]
			allowed = allowed && state != lifecycleDeleted && (model.ReleaseScope != ReleaseScopeExternal || state != lifecycleSuperseded || includeSuperseded)
		}
		if allowed {
			events = append(events, event)
		}
	}
	return events
}

func validateSlideLifecycle(slide reportingDeckSlide, model RedactedExportModel) error {
	check := func(refs []string) error {
		for _, ref := range refs {
			state := model.lifecycleRestrictions[ref]
			if state == lifecycleDeleted || model.ReleaseScope == ReleaseScopeExternal && state == lifecycleSuperseded && !slide.includeSuperseded {
				return newRenderValidationError("redaction_manifest_invalid", "redaction_action_unresolved")
			}
		}
		return nil
	}
	if err := check(slide.titleSourceRefs); err != nil {
		return err
	}
	if err := check(slide.noteSourceRefs); err != nil {
		return err
	}
	var blocks func([]ReportingBlock) error
	blocks = func(items []ReportingBlock) error {
		for _, block := range items {
			for _, field := range block.Fields {
				if err := check(field.SourceRefs); err != nil {
					return err
				}
			}
			if err := blocks(block.Children); err != nil {
				return err
			}
		}
		return nil
	}
	return blocks(slide.Blocks)
}

// Only this private tree is used for structural-removal accounting. Preserve
// admitted addresses, but exclude ineligible content before identifying Core
// actions and their consequences. Public retained trees are rebuilt separately.
func eligibleSourceTree(source SnapshotContent, fields []ExportField) SnapshotContent {
	allowed := map[string]bool{}
	for _, field := range fields {
		allowed[field.Path] = true
	}
	var prune func([]ReportingBlock) []ReportingBlock
	prune = func(blocks []ReportingBlock) []ReportingBlock {
		out := []ReportingBlock{}
		for _, block := range blocks {
			retained := []ReportingField{}
			for _, field := range block.Fields {
				if len(field.SourceRefs) > 0 && allowed[field.SourceRefs[0]] {
					retained = append(retained, field)
				}
			}
			block.Fields, block.Children = retained, prune(block.Children)
			if len(block.Fields) > 0 || len(block.Children) > 0 {
				out = append(out, block)
			}
		}
		return out
	}
	sections := []ReportingSection{}
	for _, section := range source.Sections {
		section.Blocks = prune(section.Blocks)
		if len(section.Blocks) > 0 || section.SectionKind == "validation" {
			sections = append(sections, section)
		}
	}
	source.Sections = sections
	return source
}
