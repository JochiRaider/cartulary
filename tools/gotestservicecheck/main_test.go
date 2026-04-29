package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const testModulePath = "github.com/JochiRaider/cartulary"

func TestScanBlocksFuturePhaseBackendUnitHelper(t *testing.T) {
	root := newFixtureRepo(t)
	writeFixtureFile(t, root, "internal/modules/future/phase10_store_test.go", `package future

import (
	"testing"

	store "github.com/JochiRaider/cartulary/internal/testutil/phase10storetest"
)

func TestPhase10_NewStore_U_10_01(t *testing.T) {
	store.StartStore(t, "phase10")
}
`)
	writeFixtureManifest(t, root, "phase10", `[
	{
		"coverage": "authoritative",
		"runner": "go_test",
		"execution_dependency": "backend_unit",
		"symbol": "TestPhase10_NewStore_U_10_01"
	}
]`)

	findings := scanFixture(t, root)
	requireFinding(t, findings, "TestPhase10_NewStore_U_10_01", "store.StartStore", testModulePath+"/internal/testutil/phase10storetest")
}

func TestScanAllowsBackendStoreManifestSymbol(t *testing.T) {
	root := newFixtureRepo(t)
	writeFixtureFile(t, root, "internal/modules/future/phase10_store_test.go", `package future

import (
	"testing"

	store "github.com/JochiRaider/cartulary/internal/testutil/phase10storetest"
)

func TestPhase10_NewStore_U_10_01(t *testing.T) {
	store.StartStore(t, "phase10")
}
`)
	writeFixtureManifest(t, root, "phase10", `[
	{
		"coverage": "authoritative",
		"runner": "go_test",
		"execution_dependency": "backend_store",
		"symbol": "TestPhase10_NewStore_U_10_01"
	}
]`)

	if findings := scanFixture(t, root); len(findings) != 0 {
		t.Fatalf("expected backend_store manifest symbol to be allowed, got %#v", findings)
	}
}

func TestScanAllowsBackendStoreManifestSymbolsArray(t *testing.T) {
	root := newFixtureRepo(t)
	writeFixtureFile(t, root, "internal/modules/future/phase10_store_test.go", `package future

import (
	"testing"

	store "github.com/JochiRaider/cartulary/internal/testutil/phase10storetest"
)

func TestPhase10_ArrayAllowed_U_10_02(t *testing.T) {
	store.StartStore(t, "phase10")
}
`)
	writeFixtureManifest(t, root, "phase10", `[
	{
		"coverage": "authoritative",
		"runner": "go_test",
		"execution_dependency": "backend_store",
		"symbols": ["TestPhase10_ArrayAllowed_U_10_02"]
	}
]`)

	if findings := scanFixture(t, root); len(findings) != 0 {
		t.Fatalf("expected backend_store manifest symbols[] entry to be allowed, got %#v", findings)
	}
}

func TestScanBlocksAliasedFutureRuntimeHelper(t *testing.T) {
	root := newFixtureRepo(t)
	writeFixtureFile(t, root, "internal/modules/future/phase12_runtime_test.go", `package future

import (
	"testing"

	rt "github.com/JochiRaider/cartulary/internal/testutil/phase12test"
)

func TestPhase12_Runtime_U_12_01(t *testing.T) {
	rt.StartRuntime(t)
}
`)

	findings := scanFixture(t, root)
	requireFinding(t, findings, "TestPhase12_Runtime_U_12_01", "rt.StartRuntime", testModulePath+"/internal/testutil/phase12test")
}

func TestScanBlocksServiceHarnessesByImportPath(t *testing.T) {
	root := newFixtureRepo(t)
	writeFixtureFile(t, root, "internal/modules/future/phase10_services_test.go", `package future

import (
	"testing"

	database "github.com/JochiRaider/cartulary/internal/testutil/pgtest"
	objectstore "github.com/JochiRaider/cartulary/internal/testutil/s3test"
)

func TestPhase10_ServiceHarnesses_U_10_03(t *testing.T) {
	database.Start(t)
	objectstore.StartOwned(t)
}
`)

	findings := scanFixture(t, root)
	requireFinding(t, findings, "TestPhase10_ServiceHarnesses_U_10_03", "database.Start", testModulePath+"/internal/testutil/pgtest")
	requireFinding(t, findings, "TestPhase10_ServiceHarnesses_U_10_03", "objectstore.StartOwned", testModulePath+"/internal/testutil/s3test")
}

func TestScanIgnoresSupportAndIntegrationNames(t *testing.T) {
	root := newFixtureRepo(t)
	writeFixtureFile(t, root, "internal/modules/future/phase10_support_test.go", `package future

import (
	"testing"

	store "github.com/JochiRaider/cartulary/internal/testutil/phase10storetest"
)

func TestSupportPhase10_StoreHelper(t *testing.T) {
	store.StartStore(t, "phase10")
}

func TestPhase10_Runtime_I_10_01(t *testing.T) {
	store.StartStore(t, "phase10")
}
`)

	if findings := scanFixture(t, root); len(findings) != 0 {
		t.Fatalf("expected support and integration-looking tests to be ignored, got %#v", findings)
	}
}

func newFixtureRepo(t *testing.T) string {
	t.Helper()

	root := t.TempDir()
	writeFixtureFile(t, root, "go.mod", "module "+testModulePath+"\n")
	if err := os.MkdirAll(filepath.Join(root, "tools"), 0o755); err != nil {
		t.Fatalf("create tools dir: %v", err)
	}
	return root
}

func writeFixtureManifest(t *testing.T, root, phase, unitEntries string) {
	t.Helper()

	writeFixtureFile(t, root, filepath.Join("tools", phase+"_test_map.json"), fmt.Sprintf(`{
	"unit": %s
}
`, unitEntries))
}

func writeFixtureFile(t *testing.T, root, relativePath, content string) {
	t.Helper()

	fullPath := filepath.Join(root, relativePath)
	if err := os.MkdirAll(filepath.Dir(fullPath), 0o755); err != nil {
		t.Fatalf("create fixture dir for %s: %v", relativePath, err)
	}
	if err := os.WriteFile(fullPath, []byte(content), 0o644); err != nil {
		t.Fatalf("write fixture file %s: %v", relativePath, err)
	}
}

func scanFixture(t *testing.T, root string) []finding {
	t.Helper()

	t.Setenv("CARTULARY_PHASE_MANIFEST_ROOT", root)
	findings, err := scanRoot(root)
	if err != nil {
		t.Fatalf("scan fixture: %v", err)
	}
	return findings
}

func requireFinding(t *testing.T, findings []finding, testName, selector, importPath string) {
	t.Helper()

	for _, finding := range findings {
		if finding.Test == testName && finding.Selector == selector && finding.ImportPath == importPath {
			if !strings.HasPrefix(finding.File, "internal/") {
				t.Fatalf("expected repo-relative finding file, got %#v", finding)
			}
			return
		}
	}
	t.Fatalf("missing finding test=%s selector=%s import=%s in %#v", testName, selector, importPath, findings)
}
