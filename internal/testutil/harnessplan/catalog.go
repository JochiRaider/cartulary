package harnessplan

import (
	"fmt"
	"path"
	"strings"
)

// Catalog is the owner catalog: every registered owner and its rows.
type Catalog struct {
	Owners []CatalogOwner
	Rows   []Row
}

// CatalogOwner registers one owner and its family manifest.
type CatalogOwner struct {
	OwnerID      string `json:"owner_id"`
	ManifestPath string `json:"manifest_path"`
	Status       string `json:"status"`
}

// Row is one owner-catalog verification row.
type Row struct {
	RowID                 string   `json:"row_id"`
	OwnerID               string   `json:"owner_id"`
	CollaboratorIDs       []string `json:"collaborator_ids,omitempty"`
	FamilyID              string   `json:"family_id"`
	VerificationIDs       []string `json:"verification_ids"`
	Runner                string   `json:"runner"`
	Selector              Selector `json:"selector"`
	EvidenceClass         string   `json:"evidence_class"`
	RuntimeProfileID      string   `json:"runtime_profile_id"`
	ResourceProfileID     string   `json:"resource_profile_id"`
	FixtureCapability     string   `json:"fixture_capability"`
	AggregatePlacement    string   `json:"aggregate_placement"`
	PostgresIsolation     string   `json:"postgres_isolation"`
	ProcessIsolation      string   `json:"process_isolation,omitempty"`
	ServiceDependencies   []string `json:"service_dependencies"`
	RuntimeBinaryIDs      []string `json:"runtime_binary_ids,omitempty"`
	PreparationOperations []string `json:"preparation_operations,omitempty"`
	FixtureProfileID      string   `json:"fixture_profile_id,omitempty"`
	FixtureInputRoots     []string `json:"fixture_input_roots,omitempty"`
	BrowserStateScope     string   `json:"browser_state_scope,omitempty"`
	EstimatedWorkMS       int64    `json:"estimated_work_ms,omitempty"`
	Status                string   `json:"status"`
}

// Selector names exactly what a row runs. Fields depend on the runner:
// Go rows name a package and tests, Vitest rows a project file and exact full
// titles, Playwright rows also a stage and scenario IDs, and shell rows an
// operation.
type Selector struct {
	Package     string   `json:"package,omitempty"`
	Tests       []string `json:"tests,omitempty"`
	File        string   `json:"file,omitempty"`
	ProjectID   string   `json:"project_id,omitempty"`
	Stage       string   `json:"stage,omitempty"`
	Titles      []string `json:"titles,omitempty"`
	ScenarioIDs []string `json:"scenario_ids,omitempty"`
	OperationID string   `json:"operation_id,omitempty"`
}

// Runner names.
const (
	RunnerGo         = "go"
	RunnerVitest     = "vitest"
	RunnerPlaywright = "playwright"
	RunnerShell      = "shell"
)

type catalogRegistry struct {
	Owners []CatalogOwner `json:"owners"`
}

type familyManifest struct {
	OwnerID string `json:"owner_id"`
	Rows    []Row  `json:"rows"`
}

func loadCatalog(root string) (Catalog, error) {
	var registry catalogRegistry
	if err := readDocument(root, catalogRegistryPath, "cartulary.test_owner_registry.v1", &registry); err != nil {
		return Catalog{}, err
	}
	catalog := Catalog{Owners: registry.Owners}
	for _, owner := range registry.Owners {
		manifest := path.Clean(owner.ManifestPath)
		if manifest != owner.ManifestPath || !strings.HasPrefix(manifest, "tools/test_families/") {
			return Catalog{}, fmt.Errorf("%w: %s: owner %s manifest path %q is outside tools/test_families", ErrInput, catalogRegistryPath, owner.OwnerID, owner.ManifestPath)
		}
		var family familyManifest
		if err := readDocument(root, manifest, "cartulary.test_family_manifest.v15", &family); err != nil {
			return Catalog{}, err
		}
		if family.OwnerID != owner.OwnerID {
			return Catalog{}, fmt.Errorf("%w: %s: owner_id %q does not match registered owner %q", ErrInput, manifest, family.OwnerID, owner.OwnerID)
		}
		catalog.Rows = append(catalog.Rows, family.Rows...)
	}
	return catalog, nil
}
