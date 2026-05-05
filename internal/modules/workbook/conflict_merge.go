package workbook

import "strings"

type textMergeHunk struct {
	start       int
	end         int
	replacement []string
}

func suggestedTextMergeValue(baseValue any, serverValue any, clientValue any) (string, bool) {
	base, ok := conflictTextForMerge(baseValue)
	if !ok {
		return "", false
	}
	server, ok := conflictTextForMerge(serverValue)
	if !ok {
		return "", false
	}
	client, ok := conflictTextForMerge(clientValue)
	if !ok {
		return "", false
	}
	baseLines := strings.Split(base, "\n")
	serverHunk := changedTextMergeHunk(baseLines, strings.Split(server, "\n"))
	clientHunk := changedTextMergeHunk(baseLines, strings.Split(client, "\n"))
	if textMergeHunksEqual(serverHunk, clientHunk) {
		return strings.Join(applyTextMergeHunks(baseLines, serverHunk), "\n"), true
	}
	if serverHunk.start == serverHunk.end && clientHunk.start == clientHunk.end && serverHunk.start == clientHunk.start {
		return "", false
	}
	if serverHunk.end <= clientHunk.start {
		return strings.Join(applyTextMergeHunks(baseLines, clientHunk, serverHunk), "\n"), true
	}
	if clientHunk.end <= serverHunk.start {
		return strings.Join(applyTextMergeHunks(baseLines, serverHunk, clientHunk), "\n"), true
	}
	return "", false
}

func conflictTextForMerge(value any) (string, bool) {
	if value == nil {
		return "", true
	}
	text, ok := value.(string)
	if !ok {
		return "", false
	}
	text = strings.ReplaceAll(text, "\r\n", "\n")
	text = strings.ReplaceAll(text, "\r", "\n")
	return text, true
}

func changedTextMergeHunk(baseLines []string, variantLines []string) textMergeHunk {
	prefix := 0
	for prefix < len(baseLines) && prefix < len(variantLines) && baseLines[prefix] == variantLines[prefix] {
		prefix++
	}
	baseSuffix := len(baseLines)
	variantSuffix := len(variantLines)
	for baseSuffix > prefix && variantSuffix > prefix && baseLines[baseSuffix-1] == variantLines[variantSuffix-1] {
		baseSuffix--
		variantSuffix--
	}
	replacement := append([]string(nil), variantLines[prefix:variantSuffix]...)
	return textMergeHunk{start: prefix, end: baseSuffix, replacement: replacement}
}

func textMergeHunksEqual(left textMergeHunk, right textMergeHunk) bool {
	if left.start != right.start || left.end != right.end || len(left.replacement) != len(right.replacement) {
		return false
	}
	for index := range left.replacement {
		if left.replacement[index] != right.replacement[index] {
			return false
		}
	}
	return true
}

func applyTextMergeHunks(baseLines []string, hunks ...textMergeHunk) []string {
	result := append([]string(nil), baseLines...)
	for _, hunk := range hunks {
		next := make([]string, 0, len(result)-hunk.end+hunk.start+len(hunk.replacement))
		next = append(next, result[:hunk.start]...)
		next = append(next, hunk.replacement...)
		next = append(next, result[hunk.end:]...)
		result = next
	}
	return result
}
