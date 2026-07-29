package imports

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestImportsTargetsUseOnlyGeneratedRegistry(t *testing.T) {
	body, err := os.ReadFile(filepath.Clean("targets.go"))
	if err != nil {
		t.Fatalf("read targets.go: %v", err)
	}
	content := string(body)
	if !strings.Contains(
		content,
		"internal/gen/importtargetregistry",
	) || !strings.Contains(content, "importtargetregistry.Targets") {
		t.Fatal("targets.go must project the generated import target registry")
	}
	for _, forbidden := range []string{
		"internal/modules/artifacts",
		"internal/modules/assessments",
		"internal/modules/entities",
		"internal/modules/evidence",
		"internal/modules/indicators",
		"internal/modules/parties",
		"internal/modules/tasksdecisions",
		"internal/modules/timeline",
		`"cartulary.view.hosts.v1"`,
		`"cartulary.view.indicators.v1"`,
		`"network_flow_import_facade_v1"`,
	} {
		if strings.Contains(content, forbidden) {
			t.Fatalf("targets.go contains duplicated owner fact %q", forbidden)
		}
	}
}

func TestImportsOwnerApplyUsesInjectedFacadeRegistry(t *testing.T) {
	body, err := os.ReadFile(filepath.Clean("owner_apply.go"))
	if err != nil {
		t.Fatalf("read owner_apply.go: %v", err)
	}
	content := string(body)
	if !strings.Contains(content, "s.ownerCreateRegistry.Resolve(") ||
		!strings.Contains(content, "owner.CreateImportRowTx(") {
		t.Fatal("owner_apply.go must resolve and invoke the injected owner facade")
	}
	for _, forbidden := range []string{
		"internal/modules/artifacts",
		"internal/modules/assessments",
		"internal/modules/entities",
		"internal/modules/evidence",
		"internal/modules/indicators",
		"internal/modules/parties",
		"internal/modules/tasksdecisions",
		"artifacts.NewStore(",
		"assessments.NewStore(",
		"hostidentity.NewStore(",
		"evidence.NewStore(",
		"indicators.NewStore(",
		"parties.NewStore(",
		"tasksdecisions.NewStore(",
		"FROM users",
		"FROM records",
	} {
		if strings.Contains(content, forbidden) {
			t.Fatalf("owner_apply.go retains cross-owner dependency %q", forbidden)
		}
	}
}

func TestImportsProductionPackageHasNoConcretePeerStoresOrPeerTableSQL(t *testing.T) {
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatalf("glob imports package files: %v", err)
	}
	forbiddenImports := []string{
		"internal/modules/artifacts",
		"internal/modules/assessments",
		"internal/modules/entities/hostidentity",
		"internal/modules/evidence",
		"internal/modules/indicators",
		"internal/modules/parties",
		"internal/modules/tasksdecisions",
		"internal/modules/timeline",
	}
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		data, err := os.ReadFile(file)
		if err != nil {
			t.Fatalf("read %s: %v", file, err)
		}
		content := string(data)
		for _, forbidden := range forbiddenImports {
			if strings.Contains(content, forbidden) {
				t.Fatalf("%s imports peer owner %q", file, forbidden)
			}
		}
	}
}
