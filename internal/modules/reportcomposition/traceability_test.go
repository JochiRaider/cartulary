package reportcomposition

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"
)

func TestReportCompositionTraceabilityAndFixtureCorpus(t *testing.T) {
	root := reportCompositionRepoRoot(t)
	requireReportCompositionVerificationOwner(t, root)
	corpus := readReportCompositionFixtureCorpus(t, root)
	if _, ok := corpus["RC-FIX-023"]; !ok {
		t.Fatal("fixture corpus omits RC-FIX-023 traceability fixture")
	}
	for id, row := range corpus {
		if !strings.HasPrefix(id, "RC-FIX-") {
			t.Fatalf("fixture corpus contains invalid fixture id %q", id)
		}
		if row.Status != "accepted" && row.Status != "future_only" {
			t.Fatalf("fixture %s has unsupported status %q", id, row.Status)
		}
		if row.Status == "future_only" && strings.TrimSpace(row.OwnerApproval) == "" {
			t.Fatalf("future-only fixture %s must record owner approval", id)
		}
		if len(row.Evidence) == 0 {
			t.Fatalf("fixture %s must list evidence selectors or inert traceability references", id)
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
		StatusVocab []string                      `json:"status_vocab"`
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
	if !reflect.DeepEqual(payload.StatusVocab, []string{"accepted", "future_only"}) {
		t.Fatalf("unexpected fixture status vocabulary: %v", payload.StatusVocab)
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

func requireReportCompositionVerificationOwner(t testing.TB, root string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, "contracts", "verification", "owners", "module.reportcomposition.json"))
	if err != nil {
		t.Fatalf("read report-composition verification contract: %v", err)
	}
	var contract struct {
		SchemaID      string `json:"schema_id"`
		OwnerID       string `json:"owner_id"`
		Verifications []struct {
			VerificationID string   `json:"verification_id"`
			EvidenceKinds  []string `json:"evidence_kinds"`
			Status         string   `json:"status"`
		} `json:"verifications"`
	}
	if err := json.Unmarshal(data, &contract); err != nil {
		t.Fatalf("decode report-composition verification contract: %v", err)
	}
	if contract.SchemaID != "cartulary.verification_contract.v1" || contract.OwnerID != "module.reportcomposition" {
		t.Fatalf("unexpected report-composition verification identity: %s/%s", contract.SchemaID, contract.OwnerID)
	}
	for _, verification := range contract.Verifications {
		if verification.VerificationID == "module.reportcomposition.verification.fixture_corpus" &&
			verification.Status == "active" &&
			reflect.DeepEqual(verification.EvidenceKinds, []string{"go_test", "static_check"}) {
			return
		}
	}
	t.Fatal("report-composition fixture corpus verification is not active")
}

func reportCompositionRepoRoot(t testing.TB) string {
	t.Helper()
	dir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, "go.mod")); err == nil {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			t.Fatal("could not locate repository root")
		}
		dir = parent
	}
}
