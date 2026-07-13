package graphprojection

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestConformanceMatrixDoesNotOverclaimUnexecutableFixtures(t *testing.T) {
	root := repoRoot(t)
	matrixBody, err := os.ReadFile(filepath.Join(root, "contracts", "graph-projection", "conformance_matrix.v1.json"))
	if err != nil {
		t.Fatalf("read graph projection conformance matrix: %v", err)
	}
	var matrix struct {
		AcceptanceCriteria []struct {
			ID                string   `json:"id"`
			CoverageStatus    string   `json:"coverage_status"`
			EvidenceSelectors []string `json:"evidence_selectors"`
		} `json:"acceptance_criteria"`
	}
	if err := json.Unmarshal(matrixBody, &matrix); err != nil {
		t.Fatalf("decode graph projection conformance matrix: %v", err)
	}
	if len(matrix.AcceptanceCriteria) != 69 {
		t.Fatalf("acceptance criteria = %d want 69", len(matrix.AcceptanceCriteria))
	}
	for _, criterion := range matrix.AcceptanceCriteria {
		for _, selector := range criterion.EvidenceSelectors {
			if criterion.CoverageStatus == "implemented" && strings.Contains(selector, "conformance_matrix.v1.json") {
				t.Fatalf("%s uses the conformance matrix as self-referential implementation evidence", criterion.ID)
			}
		}
	}

	corpusBody, err := os.ReadFile(filepath.Join(root, "contracts", "graph-projection", "fixtures", "corpus.v1.json"))
	if err != nil {
		t.Fatalf("read graph projection fixture corpus: %v", err)
	}
	var corpus struct {
		Fixtures []struct {
			FixtureID string `json:"fixture_id"`
		} `json:"fixtures"`
	}
	if err := json.Unmarshal(corpusBody, &corpus); err != nil {
		t.Fatalf("decode graph projection fixture corpus: %v", err)
	}
	if len(corpus.Fixtures) != 36 {
		t.Fatalf("fixture registry = %d want 36", len(corpus.Fixtures))
	}
}
