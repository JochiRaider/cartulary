package reporting

import (
	"errors"
	"fmt"
	"slices"
)

type compositionSectionAnchor struct {
	AnchorKind    string  `json:"anchor_kind"`
	DeclarationID string  `json:"template_section_decl_id"`
	ExpansionKey  *string `json:"expansion_key"`
}
type compositionRecordAnchor struct {
	AnchorKind string `json:"anchor_kind"`
	RecordID   string `json:"record_id"`
}
type compositionBlockAnchor struct {
	AnchorKind    string                   `json:"anchor_kind"`
	SectionAnchor compositionSectionAnchor `json:"section_anchor"`
	BlockKind     string                   `json:"block_kind"`
	RecordAnchor  *compositionRecordAnchor `json:"record_anchor"`
}
type compositionDiagramAnchor struct {
	AnchorKind    string `json:"anchor_kind"`
	DeclarationID string `json:"decl_id"`
}
type deckSectionOrigin struct {
	DeclarationID string
	ExpansionKey  *string
	Required      bool
}

func compositionFailure(reason string) error {
	return newRenderValidationError("composition_invalid", reason)
}

func resolveCompositionSection(slides []reportingDeckSlide, anchor *compositionSectionAnchor) (int, error) {
	if anchor == nil || anchor.AnchorKind != "section_anchor" {
		return -1, compositionFailure("invalid_scalar_value")
	}
	found := -1
	for i, slide := range slides {
		origin := slide.sectionOrigin
		if origin == nil || origin.DeclarationID != anchor.DeclarationID || !sameOptionalText(origin.ExpansionKey, anchor.ExpansionKey) {
			continue
		}
		if found >= 0 {
			return -1, compositionFailure("composition_anchor_ambiguous")
		}
		found = i
	}
	if found < 0 {
		return -1, compositionFailure("composition_anchor_unresolved")
	}
	return found, nil
}
func sameOptionalText(a, b *string) bool {
	return a == nil && b == nil || a != nil && b != nil && *a == *b
}

func derivedDeckBlockIDs(blocks []ReportingBlock) map[string]bool {
	result := map[string]bool{}
	var visit func([]ReportingBlock)
	visit = func(items []ReportingBlock) {
		for _, b := range items {
			result[b.BlockID] = true
			visit(b.Children)
		}
	}
	visit(blocks)
	return result
}

type resolvedCompositionBlock struct {
	section  int
	siblings *[]ReportingBlock
	index    int
}

func resolveCompositionBlock(slides []reportingDeckSlide, anchor *compositionBlockAnchor) (resolvedCompositionBlock, error) {
	if anchor == nil || anchor.AnchorKind != "block_anchor" {
		return resolvedCompositionBlock{}, compositionFailure("invalid_scalar_value")
	}
	section, err := resolveCompositionSection(slides, &anchor.SectionAnchor)
	if err != nil {
		return resolvedCompositionBlock{}, err
	}
	records := slides[section].recordAnchors
	if anchor.RecordAnchor != nil {
		if err := records.resolve(anchor.RecordAnchor); err != nil {
			return resolvedCompositionBlock{}, err
		}
	}
	count := 0
	var result resolvedCompositionBlock
	var visit func(*[]ReportingBlock)
	visit = func(items *[]ReportingBlock) {
		for i := range *items {
			b := &(*items)[i]
			if slides[section].derivedBlockIDs[b.BlockID] && b.BlockKind == anchor.BlockKind &&
				(anchor.RecordAnchor == nil || records.blockContains(*b, anchor.RecordAnchor.RecordID, slides[section].derivedBlockIDs)) {
				count++
				result = resolvedCompositionBlock{section, items, i}
			}
			visit(&b.Children)
		}
	}
	visit(&slides[section].Blocks)
	if count == 0 {
		return result, compositionFailure("composition_anchor_unresolved")
	}
	if count > 1 {
		return result, compositionFailure("composition_anchor_ambiguous")
	}
	return result, nil
}

func resolveDeckLayout(contract TemplateContract, slide reportingDeckSlide) (TemplateLayout, error) {
	for _, layout := range contract.Layouts {
		if layout.LayoutID != slide.Layout {
			continue
		}
		var allowed func([]ReportingBlock) bool
		allowed = func(blocks []ReportingBlock) bool {
			for _, b := range blocks {
				if !slices.Contains(layout.AllowedBlockKinds, b.BlockKind) || !allowed(b.Children) {
					return false
				}
			}
			return true
		}
		if !allowed(slide.Blocks) {
			return TemplateLayout{}, newRenderValidationError("deck_model_invalid", "block_kind_not_allowed")
		}
		return layout, nil
	}
	return TemplateLayout{}, newRenderValidationError("deck_model_invalid", "invalid_scalar_value")
}

// Apply against current retained objects in array order, before chunking and
// layout validation. Removed targets never survive as resolvable tombstones.
func applyCompositionToDeck(slides *[]reportingDeckSlide, composition *renderComposition, contract TemplateContract, scope string, diagrams ...reportingDiagramModel) error {
	texts := map[string]compositionAuthoredText{}
	for _, text := range composition.AuthoredTexts {
		texts[text.AuthoredTextID] = text
	}
	insertedAfter := map[string]string{}
	insertedDiagrams := map[string]bool{}
	for index, op := range composition.DeckOps {
		if op.OnUnresolved == "drop" && scope == ReleaseScopeExternal {
			return compositionFailure("composition_drop_invalid_for_external_release")
		}
		textRole := map[string]string{"override_title": "title_override", "set_speaker_notes": "speaker_notes", "insert_authored_block": "authored_text"}[op.OpKind]
		text := texts[op.Payload.AuthoredTextRef]
		if textRole != "" && text.TextRole != textRole {
			return compositionFailure("authored_subject_ref_unresolved")
		}
		var err error
		switch op.OpKind {
		case "exclude_diagram", "override_diagram_labels":
			// The diagram materializer has already applied these in operation
			// order, before source and deck serialization.
			continue
		case "insert_diagram_slide":
			err = insertCompositionDiagramSlide(slides, diagrams, op, index, insertedAfter, insertedDiagrams)
		default:
			err = applyCompositionDeckOp(slides, op, text, contract, index, insertedAfter)
		}
		var invalid *renderValidationError
		if op.OnUnresolved == "drop" && errors.As(err, &invalid) && invalid.ReasonCode == "composition_anchor_unresolved" {
			continue
		}
		if err != nil {
			return err
		}
	}
	for i := range *slides {
		rebuildCompositionBlocks((*slides)[i].Blocks, fmt.Sprintf("sec-%04d", i+1), false)
	}
	return nil
}

func rebuildCompositionBlocks(blocks []ReportingBlock, parent string, nested bool) {
	for i := range blocks {
		b := &blocks[i]
		b.BlockOrdinal = i + 1
		b.BlockID = fmt.Sprintf("%s.b%04d", parent, i+1)
		b.ParentBlockID = nil
		if nested {
			b.ParentBlockID = cloneStringPtr(&parent)
		}
		rebuildCompositionBlocks(b.Children, b.BlockID, true)
		refs := emptyContainerRefs()
		for j := range b.Fields {
			b.Fields[j].FieldOrdinal = j + 1
			f := b.Fields[j]
			refs.include(f.SourceRefs, f.SupportRefs, f.DisclosurePartitionRefs)
		}
		for _, child := range b.Children {
			refs.include(child.SourceRefs, child.SupportRefs, child.DisclosurePartitionRefs)
		}
		refs = refs.normalized()
		b.SourceRefs, b.SupportRefs, b.DisclosurePartitionRefs = refs.source, refs.support, refs.partitions
	}
}

func applyCompositionDeckOp(slides *[]reportingDeckSlide, op compositionDeckOp, text compositionAuthoredText, contract TemplateContract, opIndex int, insertedAfter map[string]string) error {
	switch op.OpKind {
	case "reorder_sections":
		if len(op.Payload.SectionAnchors) == 0 {
			return compositionFailure("invalid_scalar_value")
		}
		ordered := make([]reportingDeckSlide, 0, len(*slides))
		selected := map[int]bool{}
		// Resolve the entire operation before mutation, so a dropped operation
		// cannot apply only a prefix of its requested reorder.
		for _, anchor := range op.Payload.SectionAnchors {
			i, err := resolveCompositionSection(*slides, &anchor)
			if err != nil {
				return err
			}
			if selected[i] {
				return compositionFailure("invalid_scalar_value")
			}
			selected[i] = true
			ordered = append(ordered, (*slides)[i])
		}
		for i, slide := range *slides {
			if !selected[i] {
				ordered = append(ordered, slide)
			}
		}
		*slides = ordered
	case "exclude_section", "override_slide_layout", "override_title", "set_speaker_notes", "override_click_profile":
		// Payload validation is independent of anchor existence and cannot be
		// suppressed by on_unresolved=drop.
		if op.OpKind == "override_slide_layout" {
			found := false
			for _, l := range contract.Layouts {
				found = found || l.LayoutID == op.Payload.LayoutID
			}
			if !found {
				return compositionFailure("invalid_scalar_value")
			}
		}
		if op.OpKind == "override_click_profile" && !slices.Contains([]string{"none", "reveal_blocks", "reveal_list_items"}, op.Payload.ClickProfile) {
			return compositionFailure("invalid_scalar_value")
		}
		i, err := resolveCompositionSection(*slides, op.Payload.SectionAnchor)
		if err != nil {
			return err
		}
		if text.omitted {
			return nil
		}
		slide := &(*slides)[i]
		switch op.OpKind {
		case "exclude_section":
			if slide.sectionOrigin.Required {
				return newRenderValidationError("export_model_invalid", "required_section_empty")
			}
			*slides = append((*slides)[:i], (*slides)[i+1:]...)
		case "override_slide_layout":
			slide.Layout = op.Payload.LayoutID
		case "override_title":
			slide.Title = text.Body
			slide.titlePartitions = []string{text.DisclosurePartitionRef}
			if text.retainedBody != nil {
				slide.titleSourceRefs, slide.titleSupportRefs = text.retainedBody.SourceRefs, text.retainedBody.SupportRefs
			}
		case "set_speaker_notes":
			slide.SpeakerNotes = cloneStringPtr(&text.Body)
			slide.notePartitions = []string{text.DisclosurePartitionRef}
			if text.retainedBody != nil {
				slide.noteSourceRefs, slide.noteSupportRefs = text.retainedBody.SourceRefs, text.retainedBody.SupportRefs
			}
		case "override_click_profile":
			slide.clickProfile = op.Payload.ClickProfile
		}
	case "insert_authored_block", "exclude_block":
		if op.OpKind == "insert_authored_block" && op.Payload.Position != "before" && op.Payload.Position != "after" {
			return compositionFailure("invalid_scalar_value")
		}
		target, err := resolveCompositionBlock(*slides, op.Payload.BlockAnchor)
		if err != nil {
			return err
		}
		if text.omitted {
			return nil
		}
		items := target.siblings
		if op.OpKind == "exclude_block" {
			*items = append((*items)[:target.index], (*items)[target.index+1:]...)
			pruneEmptyDeckContainers(&(*slides)[target.section].Blocks)
			if len((*slides)[target.section].Blocks) == 0 {
				if (*slides)[target.section].sectionOrigin.Required {
					return newRenderValidationError("export_model_invalid", "required_section_empty")
				}
				*slides = append((*slides)[:target.section], (*slides)[target.section+1:]...)
			}
			break
		}
		block := authoredDeckBlock(text, fmt.Sprintf("composition-block-%04d", opIndex+1))
		at := target.index
		if op.Payload.Position == "after" {
			key := (*items)[target.index].BlockID
			at++
			for at < len(*items) && insertedAfter[(*items)[at].BlockID] == key {
				at++
			}
			insertedAfter[block.BlockID] = key
		}
		*items = slices.Insert(*items, at, block)
	default:
		// Diagram effects are admitted only by the separate diagram materializer.
		// They must not become successful no-ops on a Slidev request.
		return compositionFailure("composition_anchor_unresolved")
	}
	return nil
}

func authoredDeckBlock(text compositionAuthoredText, id string) ReportingBlock {
	field := ReportingField{SchemaID: "cartulary.reporting_field.v1", FieldKey: "body", FieldOrdinal: 1, SourceValueState: "derived", RedactedValueState: "unchanged", Value: text.Body, SourceRefs: []string{}, SupportRefs: []string{}, DisclosurePartitionRefs: []string{text.DisclosurePartitionRef}}
	if text.retainedBody != nil {
		field = *text.retainedBody
	}
	return ReportingBlock{SchemaID: ReportingBlockSchemaID, BlockID: id, BlockKind: "paragraph", ContentClass: ContentClassPresentationText,
		Fields:   []ReportingField{field},
		Children: []ReportingBlock{}, SourceRefs: append([]string{}, field.SourceRefs...), SupportRefs: append([]string{}, field.SupportRefs...), DisclosurePartitionRefs: append([]string{}, field.DisclosurePartitionRefs...)}
}

func pruneEmptyDeckContainers(blocks *[]ReportingBlock) {
	result := (*blocks)[:0]
	for _, b := range *blocks {
		pruneEmptyDeckContainers(&b.Children)
		if (b.BlockKind == "bullet_list" || b.BlockKind == "table") && len(b.Children) == 0 && len(b.Fields) == 0 {
			continue
		}
		result = append(result, b)
	}
	*blocks = result
}
