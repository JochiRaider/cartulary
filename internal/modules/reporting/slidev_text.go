package reporting

import (
	"encoding/json"
	"strings"
)

// fieldValueToText is the Reporting source boundary, not a diagnostic formatter.
// Canonical admission owns number and Unicode validity; this conversion owns the
// closed scalar/flat-array shape and the text position's newline policy.
func fieldValueToText(value any, multiline bool) (string, error) {
	invalid := func() (string, error) {
		return "", newRenderValidationError("slidev_source_invalid", "invalid_field_value_for_text")
	}
	value, valid := admitReportingFieldValue(value, multiline)
	if !valid {
		return invalid()
	}
	if values, ok := value.([]any); ok {
		parts := make([]string, len(values))
		for i, item := range values {
			part, ok := fieldScalarToText(item, multiline)
			if !ok {
				return invalid()
			}
			parts[i] = part
		}
		return strings.Join(parts, ", "), nil
	}
	if text, ok := fieldScalarToText(value, multiline); ok {
		return text, nil
	}
	return invalid()
}

func fieldScalarToText(value any, multiline bool) (string, bool) {
	switch value := value.(type) {
	case nil:
		return "", true
	case string:
		return value, validReportingText(value, multiline)
	case json.Number:
		// The canonical reader has already required a finite integer token.
		return value.String(), true
	case bool:
		if value {
			return "true", true
		}
		return "false", true
	default:
		return "", false
	}
}

func serializeSpeakerNoteLines(notes []string) (string, error) {
	// Reject before escaping: escaping must never repair a forbidden comment
	// delimiter into an admissible note or obscure the original failure.
	joined := strings.Join(notes, "\n")
	issues := renderIssueCollector{stage: "slidev_source"}
	if !validReportingText(joined, true) {
		issues.add(newRenderValidationError("slidev_source_invalid", "invalid_field_value_for_text"))
	}
	if strings.Contains(joined, "-->") {
		issues.add(newRenderValidationError("slidev_source_invalid", "invalid_speaker_note_sequence"))
	}
	if err := issues.finish(); err != nil {
		return "", err
	}
	return escapeMarkdownText(joined), nil
}

// escapeMarkdownText implements slidev_markdown_escape_v2. Markdown parsing
// preserves literal character references and structure. The renderer separately
// protects text tokens from Vue interpolation after Markdown tokenization.
func escapeMarkdownText(value string) string {
	var output strings.Builder
	for lineIndex, line := range strings.Split(value, "\n") {
		if lineIndex > 0 {
			output.WriteByte('\n')
		}
		first := len(line) - len(strings.TrimLeft(line, " "))
		orderedSeparator := first
		for orderedSeparator < len(line) && line[orderedSeparator] >= '0' && line[orderedSeparator] <= '9' {
			orderedSeparator++
		}
		for index, character := range line {
			escape := strings.ContainsRune("\\`|*_[]<~$&", character)
			if index == first && strings.ContainsRune("#>-+=", character) {
				escape = true
			}
			if index == orderedSeparator && orderedSeparator > first && (character == '.' || character == ')') {
				escape = true
			}
			if escape {
				output.WriteByte('\\')
			}
			output.WriteRune(character)
		}
	}
	return output.String()
}
