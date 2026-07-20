package main

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExtensionNormativeSourceAndClauseExtraction(t *testing.T) {
	root, err := repoRoot()
	if err != nil {
		t.Fatal(err)
	}
	document, err := os.ReadFile(filepath.Join(root, "docs", "extension-subsystem-nlspec.md"))
	if err != nil {
		t.Fatal(err)
	}
	if findings := lintExtensionNormativeSource(document); len(findings) != 0 {
		t.Fatalf("current adopted source findings: %v", findings)
	}
	clauses, err := extractExtensionClauses(document)
	if err != nil {
		t.Fatal(err)
	}
	if len(clauses) < 394 || len(clauses) > 65536 {
		t.Fatalf("extracted clauses = %d; want complete bounded inventory", len(clauses))
	}
	requirements := map[string]bool{}
	criteria := map[string]bool{}
	for index, clause := range clauses {
		if clause.start < 0 || clause.start >= clause.end || clause.end > len(document) {
			t.Fatalf("clause %d has invalid half-open range [%d,%d)", index, clause.start, clause.end)
		}
		if index > 0 && clause.start < clauses[index-1].end {
			t.Fatalf("clause %d overlaps clause %d", index, index-1)
		}
		if len(clause.requirementIDs) == 0 || len(clause.acceptanceIDs) == 0 || len(clause.verificationIDs) == 0 {
			t.Fatalf("clause %d has incomplete trace associations: %#v", index, clause)
		}
		for _, requirementID := range clause.requirementIDs {
			requirements[requirementID] = true
		}
		for _, acceptanceID := range clause.acceptanceIDs {
			criteria[acceptanceID] = true
		}
	}
	for value := 1; value <= 236; value++ {
		if id := fmt.Sprintf("EXT-REQ-%03d", value); !requirements[id] {
			t.Fatalf("clause extraction omits %s", id)
		}
	}
	for value := 1; value <= 158; value++ {
		if id := fmt.Sprintf("EXT-AC-%03d", value); !criteria[id] {
			t.Fatalf("clause extraction omits %s", id)
		}
	}
}

func TestExtensionNormativeSourceLintVectors(t *testing.T) {
	root, err := repoRoot()
	if err != nil {
		t.Fatal(err)
	}
	valid, err := os.ReadFile(filepath.Join(root, "docs", "extension-subsystem-nlspec.md"))
	if err != nil {
		t.Fatal(err)
	}
	replaceOnce := func(old, replacement string) []byte {
		if !bytes.Contains(valid, []byte(old)) {
			t.Fatalf("vector source does not contain %q", old)
		}
		return bytes.Replace(valid, []byte(old), []byte(replacement), 1)
	}
	vectors := map[string][]byte{
		"bom":                  append([]byte{0xef, 0xbb, 0xbf}, valid...),
		"crlf":                 bytes.Replace(valid, []byte("\n"), []byte("\r\n"), 1),
		"nul":                  append(append([]byte(nil), valid...), 0),
		"tab":                  replaceOnce("status: adopted/current", "status:\tadopted/current"),
		"missing_final_lf":     bytes.TrimSuffix(valid, []byte("\n")),
		"extra_final_lf":       append(append([]byte(nil), valid...), '\n'),
		"heading_skip":         replaceOnce("## 4.1", "#### 4.1"),
		"setext":               replaceOnce("# Appendix A.", "Setext heading\n---\n\n# Appendix A."),
		"four_backtick":        replaceOnce("```text", "````text"),
		"unterminated_fence":   bytes.Replace(valid, []byte("```"), nil, 1),
		"raw_html":             replaceOnce("# Appendix A.", "<div>\n\n# Appendix A."),
		"nested_list":          replaceOnce("# Appendix A.", "    - nested\n\n# Appendix A."),
		"table_column_count":   replaceOnce("| --- | --- |", "| --- |"),
		"duplicate_acceptance": replaceOnce("| `EXT-AC-158` |", "| `EXT-AC-157` |"),
	}
	for name, source := range vectors {
		t.Run(name, func(t *testing.T) {
			if findings := lintExtensionNormativeSource(source); len(findings) == 0 {
				t.Fatal("invalid normative source was accepted")
			}
		})
	}

	// Pipes inside a fence are data, not a table, and requirement-looking tokens
	// inside that same fence do not change the source inventory.
	fenced := replaceOnce("```text\nunacquired", "```text\n| not | a | table |\n**EXT-REQ-999**\nunacquired")
	if findings := lintExtensionNormativeSource(fenced); len(findings) != 0 {
		t.Fatalf("fenced literals produced findings: %s", strings.Join(findings, "; "))
	}
}
