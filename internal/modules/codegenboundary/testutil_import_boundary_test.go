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

func TestRuntimeGoFilesDoNotImportInternalTestutil(t *testing.T) {
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
		if importsTestutil(t, path) {
			offenders = append(offenders, relPath)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("scan repository Go files: %v", err)
	}
	if len(offenders) > 0 {
		sort.Strings(offenders)
		t.Fatalf("runtime Go files must not import internal/testutil:\n%s", strings.Join(offenders, "\n"))
	}
}

func isOwnerTestSupportPath(relPath string) bool {
	return strings.HasPrefix(relPath, "internal/modules/") &&
		strings.Contains(relPath, "/testsupport/")
}

func importsTestutil(t *testing.T, filePath string) bool {
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
		if importPath == testutilImportPathPrefix || strings.HasPrefix(importPath, testutilImportPathPrefix+"/") {
			return true
		}
	}
	return false
}
