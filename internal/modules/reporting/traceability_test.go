package reporting

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"
)

func TestReportingTraceabilityAndFixtureCorpus(t *testing.T) {
	root := reportingRepoRoot(t)
	specBytes, err := os.ReadFile(filepath.Join(root, "docs", "reporting-subsystem-nlspec.md"))
	if err != nil {
		t.Fatalf("read reporting NLSpec: %v", err)
	}
	spec := string(specBytes)

	requirements := collectReportingRequirements(spec)
	fixtureRows := collectReportingFixtureRows(spec)
	acceptanceRows := collectReportingAcceptanceRows(spec)
	coverage := collectReportingCoverageRows(t, spec, requirements)

	for id := range requirements {
		if len(coverage[id]) == 0 {
			t.Fatalf("requirement %s has no Table 27-B coverage", id)
		}
	}
	for id := range coverage {
		if _, ok := requirements[id]; !ok {
			t.Fatalf("Table 27-B maps unknown requirement %s", id)
		}
	}
	for id, targets := range coverage {
		for _, target := range targets {
			switch {
			case strings.HasPrefix(target, "RPT-FIX-"):
				if _, ok := fixtureRows[target]; !ok {
					t.Fatalf("%s maps to undefined fixture %s", id, target)
				}
			case strings.HasPrefix(target, "RPT-AC-"):
				if _, ok := acceptanceRows[target]; !ok {
					t.Fatalf("%s maps to undefined acceptance criterion %s", id, target)
				}
			default:
				t.Fatalf("%s maps to unsupported target %s", id, target)
			}
		}
	}

	corpus := readReportingFixtureCorpus(t, root)
	for id := range fixtureRows {
		row, ok := corpus[id]
		if !ok {
			t.Fatalf("fixture corpus omits NLSpec fixture %s", id)
		}
		if !reportingFixtureStatusAllowed(row.Status) {
			t.Fatalf("fixture %s has unsupported status %q", id, row.Status)
		}
		if len(row.Evidence) == 0 {
			t.Fatalf("fixture %s must list evidence selectors", id)
		}
	}
	for id := range corpus {
		if _, ok := fixtureRows[id]; !ok {
			t.Fatalf("fixture corpus includes unknown fixture %s", id)
		}
	}

	requireCurrentOutputKindTable(t, spec)
	requireCoreReportingOutputKindAlignment(t, root)
}

type reportingFixtureRow struct {
	ID       string   `json:"id"`
	Status   string   `json:"status"`
	Evidence []string `json:"evidence"`
}

func readReportingFixtureCorpus(t testing.TB, root string) map[string]reportingFixtureRow {
	t.Helper()
	corpusBytes, err := os.ReadFile(filepath.Join(root, "contracts", "reporting", "fixtures", "corpus.v1.json"))
	if err != nil {
		t.Fatalf("read reporting fixture corpus: %v", err)
	}
	var payload struct {
		SchemaID    string                `json:"schema_id"`
		Owner       string                `json:"owner"`
		StatusVocab []string              `json:"status_vocab"`
		FixtureRows []reportingFixtureRow `json:"fixture_rows"`
	}
	if err := json.Unmarshal(corpusBytes, &payload); err != nil {
		t.Fatalf("decode reporting fixture corpus: %v", err)
	}
	if payload.SchemaID != "cartulary.reporting_fixture_corpus.v1" {
		t.Fatalf("unexpected fixture corpus schema_id %q", payload.SchemaID)
	}
	if payload.Owner != "reporting" {
		t.Fatalf("unexpected fixture corpus owner %q", payload.Owner)
	}
	declared := uniqueSortedStrings(payload.StatusVocab)
	want := []string{"blocked_core_dependency", "future_only", "implemented", "specified"}
	if strings.Join(declared, "\x00") != strings.Join(want, "\x00") {
		t.Fatalf("unexpected fixture status vocabulary: got %v want %v", declared, want)
	}
	rows := make(map[string]reportingFixtureRow, len(payload.FixtureRows))
	for _, row := range payload.FixtureRows {
		if row.ID == "" {
			t.Fatal("fixture corpus row omits id")
		}
		if _, exists := rows[row.ID]; exists {
			t.Fatalf("fixture corpus duplicates %s", row.ID)
		}
		rows[row.ID] = row
	}
	return rows
}

func reportingFixtureStatusAllowed(status string) bool {
	switch status {
	case "specified", "implemented", "blocked_core_dependency", "future_only":
		return true
	default:
		return false
	}
}

func collectReportingRequirements(spec string) map[string]int {
	re := regexp.MustCompile(`\*\*(REQ-RPT-([0-9]{3})([a-z]?))\*\*`)
	requirements := map[string]int{}
	for _, match := range re.FindAllStringSubmatch(spec, -1) {
		base, _ := strconv.Atoi(match[2])
		requirements[match[1]] = base
	}
	return requirements
}

func collectReportingFixtureRows(spec string) map[string]struct{} {
	re := regexp.MustCompile("(?m)^\\| `(RPT-FIX-[0-9]{3})` \\|")
	return collectReportingIDs(spec, re, 1)
}

func collectReportingAcceptanceRows(spec string) map[string]struct{} {
	re := regexp.MustCompile("(?m)^\\| `(RPT-AC-[A-Z0-9-]+)` \\|")
	return collectReportingIDs(spec, re, 1)
}

func collectReportingIDs(spec string, re *regexp.Regexp, group int) map[string]struct{} {
	ids := map[string]struct{}{}
	for _, match := range re.FindAllStringSubmatch(spec, -1) {
		ids[match[group]] = struct{}{}
	}
	return ids
}

func collectReportingCoverageRows(t testing.TB, spec string, requirements map[string]int) map[string][]string {
	t.Helper()
	coverage := map[string][]string{}
	reqRef := regexp.MustCompile(`REQ-RPT-[0-9]{3}[a-z]?`)
	targetRef := regexp.MustCompile(`RPT-(?:AC|FIX)-[A-Z0-9-]+`)
	for _, line := range strings.Split(spec, "\n") {
		if !strings.HasPrefix(line, "| `REQ-RPT-") {
			continue
		}
		cells := strings.Split(line, "|")
		if len(cells) < 4 {
			t.Fatalf("malformed Table 27-B row: %s", line)
		}
		rangeCell := cells[1]
		coverageCell := cells[2]
		targets := uniqueSortedStrings(targetRef.FindAllString(coverageCell, -1))
		if len(targets) == 0 {
			t.Fatalf("Table 27-B row has no RPT-AC or RPT-FIX target: %s", line)
		}
		refs := reqRef.FindAllString(rangeCell, -1)
		if len(refs) == 0 {
			t.Fatalf("Table 27-B row has no requirement reference: %s", line)
		}
		if strings.Contains(rangeCell, "..") {
			if len(refs) != 2 {
				t.Fatalf("Table 27-B range row must contain exactly two requirement refs: %s", line)
			}
			start := requirements[refs[0]]
			end := requirements[refs[1]]
			for id, base := range requirements {
				if base >= start && base <= end {
					coverage[id] = append(coverage[id], targets...)
				}
			}
			continue
		}
		for _, id := range refs {
			coverage[id] = append(coverage[id], targets...)
		}
	}
	for id, targets := range coverage {
		coverage[id] = uniqueSortedStrings(targets)
	}
	return coverage
}

func requireCurrentOutputKindTable(t testing.TB, spec string) {
	t.Helper()
	start := strings.Index(spec, "**Table 7-C. Current v1 output kinds**")
	if start < 0 {
		t.Fatal("missing Table 7-C current output-kind table")
	}
	rest := spec[start:]
	end := strings.Index(rest, "**REQ-RPT-030**")
	if end < 0 {
		t.Fatal("could not locate end of Table 7-C")
	}
	table := rest[:end]
	rowRe := regexp.MustCompile("(?m)^\\| `([^`]+)` \\|")
	var rows []string
	for _, match := range rowRe.FindAllStringSubmatch(table, -1) {
		rows = append(rows, match[1])
	}
	got := uniqueSortedStrings(rows)
	want := []string{"mermaid", "slidev"}
	if strings.Join(got, "\x00") != strings.Join(want, "\x00") {
		t.Fatalf("current output kinds = %v, want %v", got, want)
	}
	for _, legacy := range []string{"markdown", "html", "reenactment"} {
		if strings.Contains(table, "`"+legacy+"`") {
			t.Fatalf("future-only output kind %q appears in current output table", legacy)
		}
	}
}

func requireCoreReportingOutputKindAlignment(t testing.TB, root string) {
	t.Helper()
	core01Bytes, err := os.ReadFile(filepath.Join(root, "docs", "spec", "01_architecture_storage_and_view_contracts.md"))
	if err != nil {
		t.Fatalf("read Core 01 spec: %v", err)
	}
	core04Bytes, err := os.ReadFile(filepath.Join(root, "docs", "spec", "04_security_deployment_and_conformance.md"))
	if err != nil {
		t.Fatalf("read Core 04 spec: %v", err)
	}
	core01 := string(core01Bytes)
	core04 := string(core04Bytes)
	if !strings.Contains(core01, "The current Reporting v1 release output vocabulary is exactly:") {
		t.Fatal("Core 01 does not explicitly define current Reporting v1 output vocabulary")
	}
	if strings.Contains(core01, "`output_kind` MUST use a stable closed vocabulary equivalent to:\n\n- `html`") {
		t.Fatal("Core 01 still treats legacy reporting output kinds as current")
	}
	if strings.Contains(core04, "the only allowed `output_kind` values are `html`, `markdown`, `slidev`, `mermaid`, and `reenactment`") {
		t.Fatal("Core 04 AC-267 still treats legacy reporting output kinds as current")
	}
	if !strings.Contains(core04, "the only allowed `output_kind` values are `slidev` and `mermaid`") {
		t.Fatal("Core 04 AC-267 does not define the current Reporting output-kind vocabulary")
	}
}

func uniqueSortedStrings(values []string) []string {
	seen := map[string]struct{}{}
	for _, value := range values {
		seen[value] = struct{}{}
	}
	result := make([]string, 0, len(seen))
	for value := range seen {
		result = append(result, value)
	}
	sort.Strings(result)
	return result
}

func reportingRepoRoot(t testing.TB) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "docs", "reporting-subsystem-nlspec.md")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not locate repository root")
		}
		dir = parent
	}
}
