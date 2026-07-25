package reporting

import (
	"encoding/json"
	"os"
	"path/filepath"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func TestReportingTraceabilityAndFixtureCorpus(t *testing.T) {
	root := reportingRepoRoot(t)
	requireReportingVerificationOwner(t, root)

	corpus := readReportingFixtureCorpus(t, root)
	if len(corpus) == 0 {
		t.Fatal("reporting fixture corpus must not be empty")
	}
	for id, row := range corpus {
		if !strings.HasPrefix(id, "RPT-FIX-") {
			t.Fatalf("fixture corpus contains invalid fixture id %q", id)
		}
		if !reportingFixtureStatusAllowed(row.Status) {
			t.Fatalf("fixture %s has unsupported status %q", id, row.Status)
		}
		if !reflect.DeepEqual(row.RequirementIDs, []string{id}) {
			t.Fatalf("fixture %s requirement_ids = %v, want [%s]", id, row.RequirementIDs, id)
		}
	}

	gotOutputKinds := supportedOutputKinds()
	sort.Strings(gotOutputKinds)
	wantOutputKinds := []string{OutputKindMermaid, OutputKindSlidev}
	if !reflect.DeepEqual(gotOutputKinds, wantOutputKinds) {
		t.Fatalf("current output kinds = %v, want %v", gotOutputKinds, wantOutputKinds)
	}
}

type reportingFixtureRow struct {
	ID             string   `json:"id"`
	Status         string   `json:"status"`
	RequirementIDs []string `json:"requirement_ids"`
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
		FixtureRows []reportingFixtureRow `json:"fixture_rows"`
	}
	if err := json.Unmarshal(corpusBytes, &payload); err != nil {
		t.Fatalf("decode reporting fixture corpus: %v", err)
	}
	if payload.SchemaID != "cartulary.reporting_fixture_corpus.v2" {
		t.Fatalf("unexpected fixture corpus schema_id %q", payload.SchemaID)
	}
	if payload.Owner != "reporting" {
		t.Fatalf("unexpected fixture corpus owner %q", payload.Owner)
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

func requireReportingVerificationOwner(t testing.TB, root string) {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, "contracts", "verification", "owners", "module.reporting.json"))
	if err != nil {
		t.Fatalf("read reporting verification contract: %v", err)
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
		t.Fatalf("decode reporting verification contract: %v", err)
	}
	if contract.SchemaID != "cartulary.verification_contract.v2" || contract.OwnerID != "module.reporting" {
		t.Fatalf("unexpected reporting verification identity: %s/%s", contract.SchemaID, contract.OwnerID)
	}
	for _, verification := range contract.Verifications {
		if verification.VerificationID == "module.reporting.verification.fixture_corpus" &&
			verification.Status == "active" &&
			reflect.DeepEqual(verification.EvidenceKinds, []string{"go_test", "static_check"}) {
			return
		}
	}
	t.Fatal("reporting fixture corpus verification is not active")
}

func reportingFixtureStatusAllowed(status string) bool {
	switch status {
	case "implemented", "planned":
		return true
	default:
		return false
	}
}

func reportingRepoRoot(t testing.TB) string {
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
