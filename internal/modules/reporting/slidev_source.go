package reporting

import (
	"encoding/json"
	"strconv"
	"strings"
)

// Slidev's headmatter belongs to its first slide. A second frontmatter block
// before any content creates an extra blank slide; a synthetic cover also
// disagrees with the admitted model's page count. Emit exactly the model slides.
func serializeSlidevMarkdown(deck reportingDeckModel, template TemplateContract, diagrams ...preparedMermaidDiagram) ([]byte, error) {
	issues := renderIssueCollector{stage: "slidev_source"}
	invalid := func() {
		issues.add(newRenderValidationError("slidev_source_invalid", "invalid_slidev_construct"))
	}
	for _, valid := range []bool{len(deck.Slides) != 0,
		validReportingText(deck.Title, false) && len([]rune(deck.Title)) <= 120,
		validReportingText(template.Author, false) && len([]rune(template.Author)) <= 120,
		template.Theme == "default"} {
		if !valid {
			invalid()
		}
	}
	var out strings.Builder
	limits, err := resolveRenderLimits(template.DeclaredLimits)
	if err != nil {
		return nil, err
	}
	out.WriteString("---\ntitle: " + slidevYAMLString(deck.Title))
	out.WriteString("\nauthor: " + slidevYAMLString(template.Author))
	out.WriteString("\ntheme: " + slidevYAMLString(template.Theme))
	out.WriteString("\ndownload: false\nexport:\n  format: \"pdf\"\n  withClicks: true\nclickAnimation: false\nlineNumbers: false\nmonaco: false\ntwoslash: false\n")
	for index, slide := range deck.Slides {
		if !validReportingText(slide.Title, false) {
			invalid()
		}
		layout, err := resolveDeckLayout(template, slide)
		if err != nil {
			invalid()
		}
		if index > 0 {
			out.WriteString("\n---\n")
		}
		out.WriteString("layout: " + slidevYAMLString(layout.SlidevLayout) + "\n---\n\n# " + escapeMarkdownText(slide.Title) + "\n")
		wrappers, err := validatedClickWrappers(slide, limits)
		issues.add(err)
		for _, block := range slide.Blocks {
			var body string
			var err error
			if block.BlockKind == "diagram_ref" {
				body, err = serializeSlidevDiagramRef(block, diagrams)
			} else if block.BlockKind == "asset_ref" {
				body, err = serializeSlidevAssetRef(block, template)
			} else {
				body, err = serializeSlidevBlock(block, 0)
			}
			issues.add(err)
			if body != "" {
				if step, ok := wrappers[block.BlockID]; ok {
					body = "<" + step.Component + " at=\"" + strconv.Itoa(step.At) + "\">\n\n" + body + "\n\n</" + step.Component + ">"
				}
				out.WriteString("\n" + body + "\n")
			}
		}
		noteLines, err := slideSpeakerNotes(slide)
		issues.add(err)
		if len(noteLines) > 0 {
			notes, err := serializeSpeakerNoteLines(noteLines)
			if err != nil {
				issues.add(err)
				continue
			}
			out.WriteString("\n<!--\n" + notes + "\n-->\n")
		}
	}
	if err := issues.finish(); err != nil {
		return nil, err
	}
	return []byte(out.String()), nil
}

// JSON quoted strings are YAML double-quoted scalars. Text admission excludes
// controls and line breaks; the encoder protects quotes, backslashes and HTML.
func slidevYAMLString(value string) string {
	encoded, _ := json.Marshal(value)
	return string(encoded)
}
