package codegenboundary_test

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"testing"
)

const testutilImportPathPrefix = repoImportPrefix + "internal/testutil"
const revisionsSupportImportPath = testutilImportPathPrefix + "/revisionsupport"
const generatedContractsImportPath = repoImportPrefix + "internal/gen/contract"

func TestRuntimeGoFilesDoNotImportInternalTestutil(t *testing.T) {
	offenders := runtimeGoImportOffenders(t, testutilImportPathPrefix)
	if len(offenders) > 0 {
		t.Fatalf("runtime Go files must not import internal/testutil:\n%s", strings.Join(offenders, "\n"))
	}
}

func TestRuntimeGoFilesDoNotImportRevisionsSupport(t *testing.T) {
	offenders := runtimeGoImportOffenders(t, revisionsSupportImportPath)
	if len(offenders) > 0 {
		t.Fatalf("runtime Go files must not import Revisions test support:\n%s", strings.Join(offenders, "\n"))
	}
}

func runtimeGoImportOffenders(t *testing.T, importPathPrefix string) []string {
	t.Helper()
	repoRoot := filepath.Clean(filepath.Join("..", "..", ".."))

	var offenders []string
	err := filepath.WalkDir(repoRoot, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relPath, relErr := filepath.Rel(repoRoot, path)
		if relErr != nil {
			return relErr
		}
		relPath = filepath.ToSlash(relPath)

		if entry.IsDir() {
			switch {
			case relPath == ".git",
				relPath == ".cartulary",
				relPath == "internal/gen",
				relPath == "tmp":
				return filepath.SkipDir
			case relPath != "." &&
				!strings.HasPrefix(relPath, "cmd/") &&
				!strings.HasPrefix(relPath, "internal/"):
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(relPath, ".go") || strings.HasSuffix(relPath, "_test.go") {
			return nil
		}
		if strings.HasPrefix(relPath, "internal/testutil/") ||
			strings.HasPrefix(relPath, "internal/testharness/") ||
			isOwnerTestSupportPath(relPath) {
			return nil
		}
		if importsPathPrefix(t, path, importPathPrefix) {
			offenders = append(offenders, relPath)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("scan repository Go files: %v", err)
	}
	sort.Strings(offenders)
	return offenders
}

func TestOnlyApprovedGeneratedContractBoundariesImportGeneratedContracts(t *testing.T) {
	repoRoot := filepath.Clean(filepath.Join("..", "..", ".."))
	allowed := map[string]bool{
		"internal/modules/extensions/coordinator.go": true,
		"internal/platform/viewschema/registry.go":   true,
	}

	var offenders []string
	err := filepath.WalkDir(repoRoot, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relPath, relErr := filepath.Rel(repoRoot, path)
		if relErr != nil {
			return relErr
		}
		relPath = filepath.ToSlash(relPath)

		if entry.IsDir() {
			switch {
			case relPath == ".git",
				relPath == ".cartulary",
				relPath == "internal/gen",
				relPath == "tmp":
				return filepath.SkipDir
			case relPath != "." &&
				!strings.HasPrefix(relPath, "cmd/") &&
				!strings.HasPrefix(relPath, "internal/"):
				return filepath.SkipDir
			}
			return nil
		}
		if !strings.HasSuffix(relPath, ".go") || strings.HasSuffix(relPath, "_test.go") ||
			strings.HasPrefix(relPath, "internal/testutil/") ||
			strings.HasPrefix(relPath, "internal/platform/contracttest/") {
			return nil
		}
		if !allowed[relPath] && importsPathPrefix(t, path, generatedContractsImportPath) {
			offenders = append(offenders, relPath)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("scan repository Go files: %v", err)
	}
	if len(offenders) > 0 {
		sort.Strings(offenders)
		t.Fatalf("generated contract artifacts must be accessed through approved platform boundaries:\n%s", strings.Join(offenders, "\n"))
	}
}

func isOwnerTestSupportPath(relPath string) bool {
	return strings.HasPrefix(relPath, "internal/modules/") &&
		strings.Contains(relPath, "/testsupport/")
}

func importsPathPrefix(t *testing.T, filePath string, importPathPrefix string) bool {
	t.Helper()
	parsed, err := parser.ParseFile(token.NewFileSet(), filePath, nil, parser.ImportsOnly)
	if err != nil {
		t.Fatalf("parse imports for %s: %v", filePath, err)
	}
	for _, importSpec := range parsed.Imports {
		importPath, err := strconv.Unquote(importSpec.Path.Value)
		if err != nil {
			t.Fatalf("unquote import in %s: %v", filePath, err)
		}
		if importPath == importPathPrefix || strings.HasPrefix(importPath, importPathPrefix+"/") {
			return true
		}
	}
	return false
}
