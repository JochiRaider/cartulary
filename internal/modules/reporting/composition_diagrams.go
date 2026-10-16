package reporting

import (
	"errors"
	"fmt"
	"sort"
)

func resolveDiagramDeclarations(contract TemplateContract, composition *renderComposition) ([]compositionDiagramDecl, map[string]bool, error) {
	byID := map[string]compositionDiagramDecl{}
	owned := map[string]bool{}
	for _, s := range contract.Sections {
		for _, d := range s.DiagramDecls {
			if _, exists := byID[d.DeclID]; exists {
				return nil, nil, compositionFailure("duplicate_stable_id")
			}
			byID[d.DeclID] = compositionDiagramDecl{DeclID: d.DeclID, DiagramKind: d.DiagramKind, DiagramSourceKind: d.DiagramSourceKind, SourceGraphViewID: d.SourceGraphViewID, SelectionRule: d.SelectionRule, LayoutMode: "auto", LabelOverrides: []diagramLabelOverride{}}
		}
	}
	if composition != nil {
		for _, d := range composition.DiagramDecls {
			_, exists := byID[d.DeclID]
			if d.ReplacesDeclID != nil {
				if *d.ReplacesDeclID != d.DeclID || owned[d.DeclID] {
					return nil, nil, compositionFailure("diagram_layout_invalid")
				}
				if !exists {
					return nil, nil, compositionFailure("composition_replacement_target_missing")
				}
			} else if exists {
				return nil, nil, compositionFailure("duplicate_stable_id")
			}
			byID[d.DeclID], owned[d.DeclID] = d, true
		}
	}
	decls := make([]compositionDiagramDecl, 0, len(byID))
	for _, d := range byID {
		decls = append(decls, d)
	}
	sort.Slice(decls, func(i, j int) bool { return decls[i].DeclID < decls[j].DeclID })
	return decls, owned, nil
}

func insertCompositionDiagramSlide(slides *[]reportingDeckSlide, diagrams []reportingDiagramModel, op compositionDeckOp, index int, after map[string]string, inserted map[string]bool) error {
	if op.Payload.Position != "before" && op.Payload.Position != "after" {
		return compositionFailure("invalid_scalar_value")
	}
	d, err := resolveCompositionDiagram(diagrams, op.Payload.DiagramAnchor)
	if err != nil {
		return err
	}
	diagram := diagrams[d]
	if !diagram.compositionOwned {
		return compositionFailure("invalid_scalar_value")
	}
	if inserted[diagram.DiagramID] {
		return compositionFailure("composition_duplicate_diagram_insert")
	}
	s, err := resolveCompositionSection(*slides, op.Payload.SectionAnchor)
	if err != nil {
		return err
	}
	anchorID := (*slides)[s].SlideID
	includeSuperseded := (*slides)[s].includeSuperseded
	if op.Payload.Position == "after" {
		if previous := after[anchorID]; previous != "" {
			for i, slide := range *slides {
				if slide.SlideID == previous {
					s = i
					break
				}
			}
		}
		s++
	}
	blocks := []ReportingBlock{diagramReferenceBlock(diagram)}
	slide := reportingDeckSlide{SlideID: fmt.Sprintf("composition-slide-%04d", index+1), Layout: "default", Title: diagram.DiagramID, Blocks: blocks, ClickSteps: []reportingClickStep{}, clickProfile: "none", derivedBlockIDs: map[string]bool{}, includeSuperseded: includeSuperseded}
	*slides = append(*slides, reportingDeckSlide{})
	copy((*slides)[s+1:], (*slides)[s:])
	(*slides)[s] = slide
	inserted[diagram.DiagramID] = true
	if op.Payload.Position == "after" {
		after[anchorID] = slide.SlideID
	}
	return nil
}

func resolveCompositionDiagram(diagrams []reportingDiagramModel, anchor *compositionDiagramAnchor) (int, error) {
	if anchor == nil || anchor.AnchorKind != "diagram_anchor" {
		return -1, compositionFailure("invalid_scalar_value")
	}
	match := -1
	for i, d := range diagrams {
		if d.DiagramID != anchor.DeclarationID {
			continue
		}
		if match >= 0 {
			return -1, compositionFailure("composition_anchor_ambiguous")
		}
		match = i
	}
	if match < 0 {
		return -1, compositionFailure("composition_anchor_unresolved")
	}
	return match, nil
}

func applyCompositionDiagramOps(diagrams *[]reportingDiagramModel, composition *renderComposition, scope string) error {
	if composition == nil {
		return nil
	}
	for _, op := range composition.DeckOps {
		if op.OpKind != "exclude_diagram" && op.OpKind != "override_diagram_labels" {
			continue
		}
		if op.OnUnresolved == "drop" && scope == ReleaseScopeExternal {
			return compositionFailure("composition_drop_invalid_for_external_release")
		}
		i, err := resolveCompositionDiagram(*diagrams, op.Payload.DiagramAnchor)
		var invalid *renderValidationError
		if op.OnUnresolved == "drop" && errors.As(err, &invalid) && invalid.ReasonCode == "composition_anchor_unresolved" {
			continue
		}
		if err != nil {
			return err
		}
		if op.OpKind == "exclude_diagram" {
			if (*diagrams)[i].compositionOwned {
				return compositionFailure("invalid_scalar_value")
			}
			*diagrams = append((*diagrams)[:i], (*diagrams)[i+1:]...)
		} else {
			if err := applyDiagramLabelOverrides(&(*diagrams)[i], op.Payload.LabelOverrides); err != nil {
				return err
			}
		}
	}
	return nil
}

func diagramReferenceBlock(d reportingDiagramModel) ReportingBlock {
	return ReportingBlock{SchemaID: ReportingBlockSchemaID, BlockKind: "diagram_ref", ContentClass: ContentClassDerivedAnalytic,
		Fields:   []ReportingField{{SchemaID: "cartulary.reporting_field.v1", FieldKey: "diagram_id", FieldOrdinal: 1, SourceValueState: "derived", RedactedValueState: "unchanged", Value: d.DiagramID, SourceRefs: append([]string{}, d.SourceRefs...), SupportRefs: append([]string{}, d.SupportRefs...), DisclosurePartitionRefs: append([]string{}, d.DisclosurePartitionRefs...)}},
		Children: []ReportingBlock{}, SourceRefs: append([]string{}, d.SourceRefs...), SupportRefs: append([]string{}, d.SupportRefs...), DisclosurePartitionRefs: append([]string{}, d.DisclosurePartitionRefs...)}
}
