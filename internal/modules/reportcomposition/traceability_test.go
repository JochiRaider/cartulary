package reportcomposition

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

func TestReportCompositionTraceabilityAndFixtureCorpus(t *testing.T) {
	root := reportCompositionRepoRoot(t)
	specBytes, err := os.ReadFile(filepath.Join(root, "docs", "report-composition-nlspec.md"))
	if err != nil {
		t.Fatalf("read report composition NLSpec: %v", err)
	}
	spec := string(specBytes)

	requirements := collectReportCompositionRequirements(spec)
	fixtureRows := collectReportCompositionFixtureRows(spec)
	acceptanceRows := collectReportCompositionAcceptanceRows(spec)
	coverage := collectReportCompositionCoverageRows(t, spec, requirements)

	for id := range requirements {
		if len(coverage[id]) == 0 {
			t.Fatalf("requirement %s has no Table 14-C coverage", id)
		}
	}
	for id := range coverage {
		if _, ok := requirements[id]; !ok {
			t.Fatalf("Table 14-C maps unknown requirement %s", id)
		}
	}
	for id, targets := range coverage {
		for _, target := range targets {
			switch {
			case strings.HasPrefix(target, "RC-FIX-"):
				if _, ok := fixtureRows[target]; !ok {
					t.Fatalf("%s maps to undefined fixture %s", id, target)
				}
			case strings.HasPrefix(target, "RC-AC-"):
				if _, ok := acceptanceRows[target]; !ok {
					t.Fatalf("%s maps to undefined acceptance criterion %s", id, target)
				}
			default:
				t.Fatalf("%s maps to unsupported target %s", id, target)
			}
		}
	}

	corpus := readReportCompositionFixtureCorpus(t, root)
	if _, ok := corpus["RC-FIX-023"]; !ok {
		t.Fatal("fixture corpus omits RC-FIX-023 traceability fixture")
	}
	for id := range fixtureRows {
		row, ok := corpus[id]
		if !ok {
			t.Fatalf("fixture corpus omits NLSpec fixture %s", id)
		}
		if row.Status != "accepted" && row.Status != "future_only" {
			t.Fatalf("fixture %s has unsupported status %q", id, row.Status)
		}
		if row.Status == "future_only" && strings.TrimSpace(row.OwnerApproval) == "" {
			t.Fatalf("future-only fixture %s must record owner approval", id)
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
}

type reportCompositionFixtureRow struct {
	ID            string   `json:"id"`
	Status        string   `json:"status"`
	OwnerApproval string   `json:"owner_approval"`
	Evidence      []string `json:"evidence"`
}

func readReportCompositionFixtureCorpus(t testing.TB, root string) map[string]reportCompositionFixtureRow {
	t.Helper()
	corpusBytes, err := os.ReadFile(filepath.Join(root, "contracts", "report-composition", "fixtures", "corpus.v1.json"))
	if err != nil {
		t.Fatalf("read report composition fixture corpus: %v", err)
	}
	var payload struct {
		SchemaID    string                        `json:"schema_id"`
		Owner       string                        `json:"owner"`
		FixtureRows []reportCompositionFixtureRow `json:"fixture_rows"`
	}
	if err := json.Unmarshal(corpusBytes, &payload); err != nil {
		t.Fatalf("decode report composition fixture corpus: %v", err)
	}
	if payload.SchemaID != "cartulary.report_composition_fixture_corpus.v1" {
		t.Fatalf("unexpected fixture corpus schema_id %q", payload.SchemaID)
	}
	if payload.Owner != "reportcomposition" {
		t.Fatalf("unexpected fixture corpus owner %q", payload.Owner)
	}
	rows := make(map[string]reportCompositionFixtureRow, len(payload.FixtureRows))
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

func collectReportCompositionRequirements(spec string) map[string]int {
	re := regexp.MustCompile(`\*\*(REQ-RC-([0-9]{3})([a-z]?))\*\*`)
	requirements := map[string]int{}
	for _, match := range re.FindAllStringSubmatch(spec, -1) {
		base, _ := strconv.Atoi(match[2])
		requirements[match[1]] = base
	}
	return requirements
}

func collectReportCompositionFixtureRows(spec string) map[string]struct{} {
	re := regexp.MustCompile("(?m)^\\| `(RC-FIX-[0-9]{3})` \\|")
	return collectReportCompositionIDs(spec, re, 1)
}

func collectReportCompositionAcceptanceRows(spec string) map[string]struct{} {
	re := regexp.MustCompile("(?m)^\\| `(RC-AC-[A-Z0-9-]+)` \\|")
	return collectReportCompositionIDs(spec, re, 1)
}

func collectReportCompositionIDs(spec string, re *regexp.Regexp, group int) map[string]struct{} {
	ids := map[string]struct{}{}
	for _, match := range re.FindAllStringSubmatch(spec, -1) {
		ids[match[group]] = struct{}{}
	}
	return ids
}

func collectReportCompositionCoverageRows(t testing.TB, spec string, requirements map[string]int) map[string][]string {
	t.Helper()
	coverage := map[string][]string{}
	reqRef := regexp.MustCompile(`REQ-RC-[0-9]{3}[a-z]?`)
	targetRef := regexp.MustCompile(`RC-(?:AC|FIX)-[A-Z0-9-]+`)
	for _, line := range strings.Split(spec, "\n") {
		if !strings.HasPrefix(line, "| `REQ-RC-") {
			continue
		}
		cells := strings.Split(line, "|")
		if len(cells) < 4 {
			t.Fatalf("malformed Table 14-C row: %s", line)
		}
		rangeCell := cells[1]
		coverageCell := cells[2]
		targets := uniqueSorted(targetRef.FindAllString(coverageCell, -1))
		if len(targets) == 0 {
			t.Fatalf("Table 14-C row has no RC-AC or RC-FIX target: %s", line)
		}
		refs := reqRef.FindAllString(rangeCell, -1)
		if len(refs) == 0 {
			t.Fatalf("Table 14-C row has no requirement reference: %s", line)
		}
		if strings.Contains(rangeCell, "..") {
			if len(refs) != 2 {
				t.Fatalf("Table 14-C range row must contain exactly two requirement refs: %s", line)
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
		coverage[id] = uniqueSorted(targets)
	}
	return coverage
}

func uniqueSorted(values []string) []string {
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

func reportCompositionRepoRoot(t testing.TB) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "docs", "report-composition-nlspec.md")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not locate repository root")
		}
		dir = parent
	}
}
