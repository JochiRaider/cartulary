package reporting

import (
	"fmt"
	"strings"
)

func serializeSlidevAssetRef(block ReportingBlock, contract TemplateContract) (string, error) {
	invalid := newRenderValidationError("undeclared_asset", "undeclared_asset_ref")
	if len(block.Children) != 0 || len(block.Fields) != 1 || block.Fields[0].FieldKey != "asset_id" {
		return "", invalid
	}
	id, ok := block.Fields[0].Value.(string)
	if !ok {
		return "", invalid
	}
	var target *TemplateAsset
	for i := range contract.Assets {
		if contract.Assets[i].AssetID == id {
			if target != nil {
				return "", invalid
			}
			target = &contract.Assets[i]
		}
	}
	if target == nil || target.AssetRole != renderBundleRoleLocalAsset || !templateAssetMember(*target) {
		return "", invalid
	}
	label := id
	if block.Fields[0].DisplayLabel != nil {
		label = *block.Fields[0].DisplayLabel
	}
	if !validReportingText(label, false) {
		return "", invalid
	}
	return "![" + escapeMarkdownText(label) + "](" + templateAssetURL(target.BundlePath) + ")", nil
}

// Encode all but RFC 3986 unreserved octets and the owned segment separators.
// This is a URL spelling only; exact NFC filesystem paths stay unchanged.
func templateAssetURL(path string) string {
	var out strings.Builder
	for _, b := range []byte(path) {
		if b >= 'a' && b <= 'z' || b >= 'A' && b <= 'Z' || b >= '0' && b <= '9' || strings.ContainsRune("-._~/", rune(b)) {
			out.WriteByte(b)
		} else {
			fmt.Fprintf(&out, "%%%02X", b)
		}
	}
	return out.String()
}

func serializeSlidevDiagramRef(block ReportingBlock, diagrams []preparedMermaidDiagram) (string, error) {
	if len(block.Children) != 0 || len(block.Fields) != 1 || block.Fields[0].FieldKey != "diagram_id" {
		return "", invalidSlidevBlock()
	}
	id, ok := block.Fields[0].Value.(string)
	if !ok || !validReportingIdentifier(id) {
		return "", invalidSlidevBlock()
	}
	var target *preparedMermaidDiagram
	for i := range diagrams {
		if diagrams[i].ID == id {
			if target != nil {
				return "", invalidSlidevBlock()
			}
			target = &diagrams[i]
		}
	}
	if target == nil {
		return "", invalidSlidevBlock()
	}
	if target.Model.LayoutMode == "manual" {
		if !renderDiagramPath.MatchString("diagrams/"+id+".svg") || validateRenderedSVG(target.SVG) != nil {
			return "", invalidSlidevBlock()
		}
		return "![" + escapeMarkdownText(id) + "](diagrams/" + id + ".svg)", nil
	}
	if len(target.Source) == 0 || strings.Contains(string(target.Source), "```") {
		return "", invalidSlidevBlock()
	}
	return "```mermaid\n" + string(target.Source) + "```", nil
}

func invalidSlidevBlock() error {
	return newRenderValidationError("slidev_source_invalid", "invalid_slidev_construct")
}

// Each serializer emits structure itself. Field text can never contribute
// Markdown, Vue components, file paths or executable renderer instructions.
func serializeSlidevBlock(block ReportingBlock, listDepth int) (string, error) {
	switch block.BlockKind {
	case "paragraph":
		text, err := slidevFieldText(block, true)
		return escapeMarkdownText(text), err
	case "speaker_note":
		_, err := slidevFieldText(block, true)
		return "", err
	case "metric":
		if len(block.Children) != 0 {
			return "", invalidSlidevBlock()
		}
		lines := make([]string, 0, len(block.Fields))
		for _, field := range block.Fields {
			label, err := slidevFieldLabel(field)
			if err != nil {
				return "", err
			}
			value, err := fieldValueToText(field.Value, false)
			if err != nil {
				return "", err
			}
			lines = append(lines, "**"+escapeMarkdownText(label)+":** "+escapeMarkdownText(value))
		}
		return strings.Join(lines, "\n"), nil
	case "table":
		return serializeSlidevTable(block)
	case "bullet_list":
		return serializeSlidevList(block, listDepth+1)
	default:
		// References need a separately admitted typed target. Unknown kinds
		// cannot fall back to a paragraph or a dump of their field values.
		return "", invalidSlidevBlock()
	}
}

func slidevFieldText(block ReportingBlock, multiline bool) (string, error) {
	if len(block.Children) != 0 {
		return "", invalidSlidevBlock()
	}
	parts := make([]string, 0, len(block.Fields))
	for _, field := range block.Fields {
		value, err := fieldValueToText(field.Value, multiline)
		if err != nil {
			return "", err
		}
		parts = append(parts, value)
	}
	return strings.Join(parts, "\n"), nil
}

func slidevFieldLabel(field ReportingField) (string, error) {
	label := field.FieldKey
	if field.DisplayLabel != nil {
		label = *field.DisplayLabel
	}
	if !validReportingText(label, false) {
		return "", newRenderValidationError("slidev_source_invalid", "invalid_field_value_for_text")
	}
	return label, nil
}

func serializeSlidevTable(block ReportingBlock) (string, error) {
	if len(block.Fields) != 0 || len(block.Children) == 0 || len(block.Children[0].Fields) == 0 {
		return "", invalidSlidevBlock()
	}
	columns := block.Children[0].Fields
	header, delimiter := make([]string, len(columns)), make([]string, len(columns))
	for i, field := range columns {
		label, err := slidevFieldLabel(field)
		if err != nil {
			return "", err
		}
		header[i], delimiter[i] = escapeMarkdownText(label), "---"
	}
	row := func(cells []string) string { return "| " + strings.Join(cells, " | ") + " |" }
	lines := []string{row(header), row(delimiter)}
	for _, child := range block.Children {
		if child.BlockKind != "paragraph" || len(child.Children) != 0 || len(child.Fields) != len(columns) {
			return "", invalidSlidevBlock()
		}
		cells := make([]string, len(columns))
		for i, field := range child.Fields {
			if field.FieldKey != columns[i].FieldKey {
				return "", invalidSlidevBlock()
			}
			value, err := fieldValueToText(field.Value, false)
			if err != nil {
				return "", err
			}
			cells[i] = escapeMarkdownText(value)
		}
		lines = append(lines, row(cells))
	}
	return strings.Join(lines, "\n"), nil
}

func serializeSlidevList(block ReportingBlock, depth int) (string, error) {
	if depth > 3 {
		return "", newRenderValidationError("slidev_source_invalid", "list_nesting_exceeded")
	}
	if len(block.Fields) != 0 || len(block.Children) == 0 {
		return "", invalidSlidevBlock()
	}
	lines := make([]string, 0, len(block.Children))
	indent := strings.Repeat("  ", depth-1)
	for _, child := range block.Children {
		if child.BlockKind != "paragraph" {
			return "", invalidSlidevBlock()
		}
		item := child
		item.Children = nil
		text, err := slidevFieldText(item, true)
		if err != nil {
			return "", err
		}
		lines = append(lines, indent+"- "+strings.ReplaceAll(escapeMarkdownText(text), "\n", "\n"+indent+"  "))
		for _, nested := range child.Children {
			if nested.BlockKind != "bullet_list" {
				return "", invalidSlidevBlock()
			}
			text, err := serializeSlidevList(nested, depth+1)
			if err != nil {
				return "", err
			}
			lines = append(lines, text)
		}
	}
	return strings.Join(lines, "\n"), nil
}

func slideSpeakerNotes(slide reportingDeckSlide) ([]string, error) {
	notes := []string{}
	if slide.SpeakerNotes != nil {
		notes = append(notes, *slide.SpeakerNotes)
	}
	for _, block := range slide.Blocks {
		if block.BlockKind != "speaker_note" {
			continue
		}
		text, err := slidevFieldText(block, true)
		if err != nil {
			return nil, err
		}
		notes = append(notes, text)
	}
	return notes, nil
}
