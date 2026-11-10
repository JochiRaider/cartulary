package codegenboundary_test

import (
	"encoding/json"
	"fmt"
	"os"
	"path"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/testutil/sourcecheck"
)

// Exact package capabilities project Core 01's internal ownership rules. Empty
// entries intentionally declare families with no production Go consumers.
var generatedContractConsumers = map[string][]string{
	"openapi": {}, "ws": {}, "errors": {}, "graph-projection": {},
	"audit": {}, "imports": {}, "collaboration": {}, "view-inspector": {},
	"database-migrations": {}, "performance": {}, "parties": {},
	"string-contracts": {}, "view-references": {}, "reporting": {},
	"reference-pack-fixtures": {}, "cryptography": {},
	"view-schemas":    {"internal/platform/viewschema"},
	"extensions":      {"internal/modules/extensions"},
	"network-flow":    {"internal/modules/networkflow"},
	"recovery":        {"internal/modules/recovery", "internal/modules/graphprojection/restore", "internal/platform/recoverystate"},
	"revisions":       {"internal/modules/revisions"},
	"artifacts":       {"internal/modules/artifacts", "internal/modules/artifacts/internal/sourcecatalog"},
	"tasksdecisions":  {"internal/modules/tasksdecisions", "internal/modules/tasksdecisions/internal/sourcecatalog"},
	"entities":        {"internal/platform/fieldnorm"},
	"reference-packs": {"internal/modules/reference_data", "internal/modules/reference_data/internal/packformat"},
	"jobs":            {"internal/app/jobassembly"},
}

type contractFamily struct {
	ID      string   `json:"family_id"`
	Status  string   `json:"generation_status"`
	Outputs []string `json:"generated_outputs"`
}

func generatedContractPolicy(families []contractFamily, consumers map[string][]string) (map[string][]string, error) {
	policy := map[string][]string{}
	seen := map[string]bool{}
	for _, family := range families {
		if family.Status != "active" {
			continue
		}
		allowed, ok := consumers[family.ID]
		if !ok {
			return nil, fmt.Errorf("unassigned generated family %s", family.ID)
		}
		if seen[family.ID] {
			return nil, fmt.Errorf("duplicate family %s", family.ID)
		}
		seen[family.ID] = true
		for _, output := range family.Outputs {
			if !strings.HasPrefix(output, "internal/gen/contract") || !strings.HasSuffix(output, ".go") {
				continue
			}
			key := repoImportPrefix + path.Dir(output)
			if prior, ok := policy[key]; ok && !slices.Equal(prior, allowed) {
				return nil, fmt.Errorf("conflicting consumers for %s", key)
			}
			policy[key] = allowed
		}
	}
	for id := range consumers {
		if !seen[id] {
			return nil, fmt.Errorf("stale generated family assignment %s", id)
		}
	}
	return policy, nil
}

func generatedContractOffenders(files []sourcecheck.FileImports, policy map[string][]string) []string {
	var offenders []string
	for _, file := range files {
		for _, imported := range file.Imports {
			if !strings.HasPrefix(imported, repoImportPrefix+"internal/gen/contract") {
				continue
			}
			allowed, known := policy[imported]
			if !known || !slices.Contains(allowed, path.Dir(file.Path)) {
				offenders = append(offenders, file.Path+" -> "+imported)
			}
		}
	}
	sort.Strings(offenders)
	return offenders
}

func TestOnlyApprovedGeneratedContractBoundariesImportGeneratedContracts(t *testing.T) {
	t.Run("policy fixtures", func(t *testing.T) {
		families := []contractFamily{{ID: "sample", Status: "active", Outputs: []string{"internal/gen/contractsample/types_gen.go"}}}
		consumers := map[string][]string{"sample": {"internal/modules/owner"}}
		policy, err := generatedContractPolicy(families, consumers)
		if err != nil {
			t.Fatal(err)
		}
		imported := repoImportPrefix + "internal/gen/contractsample"
		for _, file := range []string{"internal/modules/owner/renamed.go", "internal/modules/owner/split.go"} {
			if got := generatedContractOffenders([]sourcecheck.FileImports{{Path: file, Imports: []string{imported}}}, policy); len(got) != 0 {
				t.Fatal(got)
			}
		}
		for _, file := range []string{"cmd/server/main.go", "internal/modules/peer/use.go", "internal/modules/owner/child/use.go"} {
			if got := generatedContractOffenders([]sourcecheck.FileImports{{Path: file, Imports: []string{imported}}}, policy); len(got) != 1 {
				t.Fatalf("missing rejection: %s", file)
			}
		}
		for _, denied := range []string{repoImportPrefix + "internal/gen/contract", imported + "/private", imported + "future"} {
			if got := generatedContractOffenders([]sourcecheck.FileImports{{Path: "internal/modules/owner/use.go", Imports: []string{denied}}}, policy); len(got) != 1 {
				t.Fatalf("unknown package accepted: %s", denied)
			}
		}
		consumers["sample"] = nil
		policy, err = generatedContractPolicy(families, consumers)
		if err != nil {
			t.Fatal(err)
		}
		if len(generatedContractOffenders([]sourcecheck.FileImports{{Path: "internal/modules/owner/use.go", Imports: []string{imported}}}, policy)) != 1 {
			t.Fatal("empty consumer disposition allowed access")
		}
		if _, err := generatedContractPolicy(families, map[string][]string{}); err == nil {
			t.Fatal("unassigned family accepted")
		}
		if _, err := generatedContractPolicy(nil, consumers); err == nil {
			t.Fatal("stale family accepted")
		}
	})
	root := filepath.Join("..", "..", "..")
	data, err := os.ReadFile(filepath.Join(root, "contracts", "index.json"))
	if err != nil {
		t.Fatal(err)
	}
	var registry struct {
		Families []contractFamily `json:"families"`
	}
	if err := json.Unmarshal(data, &registry); err != nil {
		t.Fatal(err)
	}
	policy, err := generatedContractPolicy(registry.Families, generatedContractConsumers)
	if err != nil {
		t.Fatal(err)
	}
	files, err := runtimeImports(os.DirFS(root))
	if err != nil {
		t.Fatal(err)
	}
	if offenders := generatedContractOffenders(files, policy); len(offenders) > 0 {
		t.Fatalf("unapproved generated-contract consumers:\n%s", strings.Join(offenders, "\n"))
	}
}
