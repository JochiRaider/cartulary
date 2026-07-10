package workbook

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

type backendBoundaryManifestForTest struct {
	ForbiddenGoImports []forbiddenGoImportBoundaryRuleForTest `json:"forbidden_go_imports"`
	GoImportAllowlists []goImportAllowlistBoundaryRuleForTest `json:"go_import_allowlists"`
}

type forbiddenGoImportBoundaryRuleForTest struct {
	ID             string   `json:"id"`
	Imports        []string `json:"imports"`
	ScanPaths      []string `json:"scan_paths"`
	AllowedPaths   []string `json:"allowed_paths"`
	ProductionOnly bool     `json:"production_only"`
}

type goImportAllowlistBoundaryRuleForTest struct {
	ID              string   `json:"id"`
	ImportPrefix    string   `json:"import_prefix"`
	ScanPaths       []string `json:"scan_paths"`
	ProductionOnly  bool     `json:"production_only"`
	AllowedPrefixes []string `json:"allowed_prefixes"`
	AllowedImports  []string `json:"allowed_imports"`
}

func TestWorkbookBoundaryRulesAreManifestBacked(t *testing.T) {
	manifestPath := filepath.Clean(filepath.Join("..", "..", "..", "tools", "backend_module_boundaries.json"))
	body, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatalf("read backend module boundary manifest: %v", err)
	}
	var manifest backendBoundaryManifestForTest
	if err := json.Unmarshal(body, &manifest); err != nil {
		t.Fatalf("decode backend module boundary manifest: %v", err)
	}

	forbiddenRule, ok := findForbiddenGoImportRule(manifest, "source-owners-no-workbook-mutation-conflict-imports")
	if !ok {
		t.Fatalf("missing source-owner workbook mutation/conflict import rule")
	}
	if !forbiddenRule.ProductionOnly ||
		!stringSliceContains(forbiddenRule.ScanPaths, "internal/modules/**") ||
		!stringSliceContains(forbiddenRule.AllowedPaths, "internal/modules/workbook/**") ||
		!stringSliceContains(forbiddenRule.Imports, "github.com/JochiRaider/cartulary/internal/modules/workbook/collectionpolicy") ||
		!stringSliceContains(forbiddenRule.Imports, "github.com/JochiRaider/cartulary/internal/modules/workbook/conflicts") {
		t.Fatalf("source-owner forbidden import rule is incomplete: %#v", forbiddenRule)
	}

	allowlistRule, ok := findGoImportAllowlistRule(manifest, "workbook-approved-owner-imports")
	if !ok {
		t.Fatalf("missing workbook owner import allowlist rule")
	}
	requiredImports := []string{
		"github.com/JochiRaider/cartulary/internal/modules/artifacts",
		"github.com/JochiRaider/cartulary/internal/modules/collaboration",
		"github.com/JochiRaider/cartulary/internal/modules/entities/hostidentity",
		"github.com/JochiRaider/cartulary/internal/modules/evidence",
		"github.com/JochiRaider/cartulary/internal/modules/indicators",
		"github.com/JochiRaider/cartulary/internal/modules/records",
		"github.com/JochiRaider/cartulary/internal/modules/revisions/conflicttokens",
		"github.com/JochiRaider/cartulary/internal/modules/savedviews",
		"github.com/JochiRaider/cartulary/internal/modules/tasksdecisions",
		"github.com/JochiRaider/cartulary/internal/modules/timeline",
	}
	if !allowlistRule.ProductionOnly ||
		allowlistRule.ImportPrefix != "github.com/JochiRaider/cartulary/internal/modules/" ||
		!stringSliceContains(allowlistRule.ScanPaths, "internal/modules/workbook/**") ||
		!stringSliceContains(allowlistRule.AllowedPrefixes, "github.com/JochiRaider/cartulary/internal/modules/workbook") {
		t.Fatalf("workbook import allowlist rule is incomplete: %#v", allowlistRule)
	}
	for _, requiredImport := range requiredImports {
		if !stringSliceContains(allowlistRule.AllowedImports, requiredImport) {
			t.Fatalf("workbook import allowlist must include %s", requiredImport)
		}
	}
}

func findForbiddenGoImportRule(manifest backendBoundaryManifestForTest, id string) (forbiddenGoImportBoundaryRuleForTest, bool) {
	for _, rule := range manifest.ForbiddenGoImports {
		if rule.ID == id {
			return rule, true
		}
	}
	return forbiddenGoImportBoundaryRuleForTest{}, false
}

func findGoImportAllowlistRule(manifest backendBoundaryManifestForTest, id string) (goImportAllowlistBoundaryRuleForTest, bool) {
	for _, rule := range manifest.GoImportAllowlists {
		if rule.ID == id {
			return rule, true
		}
	}
	return goImportAllowlistBoundaryRuleForTest{}, false
}

func stringSliceContains(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
