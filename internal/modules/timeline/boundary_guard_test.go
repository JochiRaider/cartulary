package timeline

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

const repoImportPrefix = "github.com/JochiRaider/cartulary/"

func TestTimelineProductionImportBoundaries(t *testing.T) {
	entries, err := os.ReadDir(".")
	if err != nil {
		t.Fatalf("read timeline package directory: %v", err)
	}

	allowedImports := map[string]map[string]bool{
		"github.com/JochiRaider/cartulary/internal/modules/auth": {
			"api_errors.go":      true,
			"api.go":             true,
			"clipboard_paste.go": true,
			"routes.go":          true,
		},
		"github.com/JochiRaider/cartulary/internal/modules/entities/mentions": {
			"mentions_collections_store.go": true,
			"ports.go":                      true,
		},
		"github.com/JochiRaider/cartulary/internal/modules/collaboration": {
			"routes.go": true,
		},
		"github.com/JochiRaider/cartulary/internal/modules/incidents": {
			"api_errors.go":            true,
			"clipboard_paste_store.go": true,
			"lifecycle_store.go":       true,
			"routes.go":                true,
			"store.go":                 true,
			"time_conversion_store.go": true,
		},
		"github.com/JochiRaider/cartulary/internal/modules/incidentportability": {
			"incident_bundle_portability.go": true,
		},
		"github.com/JochiRaider/cartulary/internal/modules/imports/tabularingest": {
			"clipboard_paste.go": true,
		},
		"github.com/JochiRaider/cartulary/internal/modules/links": {
			"ports.go": true,
		},
		"github.com/JochiRaider/cartulary/internal/modules/projections": {
			"ports.go": true,
		},
		"github.com/JochiRaider/cartulary/internal/modules/projections/adapters": {
			"ports.go": true,
		},
		"github.com/JochiRaider/cartulary/internal/modules/timeline/rowpresenter": {
			"api.go": true,
		},
		"github.com/JochiRaider/cartulary/internal/modules/timeline/timecontract": {
			"store.go":                 true,
			"time_conversion_store.go": true,
		},
		"github.com/JochiRaider/cartulary/internal/modules/records": {
			"ports.go": true,
		},
		"github.com/JochiRaider/cartulary/internal/modules/revisions": {
			"api_errors.go": true,
			"facade.go":     true,
			"ports.go":      true,
			"routes.go":     true,
			"store.go":      true,
		},
		"github.com/JochiRaider/cartulary/internal/modules/revisions/conflicttokens": {
			"facade.go": true,
			"routes.go": true,
			"store.go":  true,
		},
		"github.com/JochiRaider/cartulary/internal/platform/httpapi": {
			"api_errors.go":      true,
			"api.go":             true,
			"bulk_mutation.go":   true,
			"clipboard_paste.go": true,
			"hooks.go":           true,
			"routes.go":          true,
		},
		"github.com/JochiRaider/cartulary/internal/platform/ws": {
			"routes.go": true,
		},
	}
	disallowedParserImports := []string{
		"archive/zip",
		"encoding/xml",
		"github.com/xuri/excelize",
		"github.com/tealeg/xlsx",
		"openxml",
	}

	for _, entry := range entries {
		fileName := entry.Name()
		if entry.IsDir() || !strings.HasSuffix(fileName, ".go") || strings.HasSuffix(fileName, "_test.go") {
			continue
		}
		fileImports := productionImports(t, fileName)
		for _, importPath := range fileImports {
			if allowedFiles, ok := allowedImports[importPath]; ok {
				if !allowedFiles[fileName] {
					t.Fatalf("%s imports %s; allowed files are %v", fileName, importPath, allowedFileNames(allowedFiles))
				}
			} else if strings.HasPrefix(importPath, repoImportPrefix+"internal/modules/") {
				t.Fatalf("%s imports unapproved sibling module %s", fileName, importPath)
			}
			for _, disallowed := range disallowedParserImports {
				if strings.Contains(importPath, disallowed) {
					t.Fatalf("%s imports parser dependency %s; Timeline paste must stay on shared tabularingest", fileName, importPath)
				}
			}
		}
	}
}

func TestTimelineProductionFacadeCallersUseCommandBoundary(t *testing.T) {
	disallowed := map[string][]string{
		"routes.go": {
			".MarkReviewed(",
		},
		filepath.Join("..", "workbook", "routes.go"): {
			".CreateTimelineRow(",
			".PatchTimelineRow(",
			".ResolveTimelineConflict(",
			".ClipboardPaste(",
			".Supersede(",
		},
		filepath.Join("..", "imports", "routes.go"): {
			".CreateImportedTimelineRow(",
		},
	}

	for path, snippets := range disallowed {
		body, err := os.ReadFile(filepath.Clean(path))
		if err != nil {
			t.Fatalf("read %s: %v", path, err)
		}
		content := string(body)
		for _, snippet := range snippets {
			if strings.Contains(content, snippet) {
				t.Fatalf("%s still uses legacy Timeline facade call %s; use command boundary methods instead", path, snippet)
			}
		}
	}
}

func TestWorkbookRoutesDoNotClassifyTimelineStoreErrors(t *testing.T) {
	disallowed := []string{
		"timeline.ErrRecordNotFound",
		"timeline.ErrRowVersionConflict",
		"timeline.ErrIllegalTransition",
		"timeline.ErrNoEffectiveChange",
		"timeline.RowVersionConflictError",
		"timeline.SameFieldConflictError",
		"timeline.IllegalTransitionError",
	}

	path := filepath.Join("..", "workbook", "routes.go")
	body, err := os.ReadFile(filepath.Clean(path))
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	content := string(body)
	for _, snippet := range disallowed {
		if strings.Contains(content, snippet) {
			t.Fatalf("%s still classifies Timeline store error %s directly; use ClassifyMutationAPIError instead", path, snippet)
		}
	}
}

func TestTimelineStoreInternalsStayPrivate(t *testing.T) {
	disallowed := []*regexp.Regexp{
		regexp.MustCompile(`\btimeline\.NewStore\b`),
		regexp.MustCompile(`\*timeline\.Store\b`),
		regexp.MustCompile(`\btimeline\.Store\b`),
	}
	roots := []string{
		filepath.Join(".."),
		filepath.Join("..", "..", "testutil"),
	}

	for _, root := range roots {
		err := filepath.WalkDir(filepath.Clean(root), func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() {
				return nil
			}
			if !strings.HasSuffix(path, ".go") || filepath.Base(path) == "boundary_guard_test.go" {
				return nil
			}
			body, err := os.ReadFile(path)
			if err != nil {
				return err
			}
			content := string(body)
			for _, pattern := range disallowed {
				if pattern.MatchString(content) {
					t.Fatalf("%s still references private Timeline store surface %s", path, pattern.String())
				}
			}
			return nil
		})
		if err != nil {
			t.Fatalf("scan %s: %v", root, err)
		}
	}
}

func TestTimelineFacadeDoesNotExposeLegacyMutationShims(t *testing.T) {
	body, err := os.ReadFile(filepath.Clean("facade.go"))
	if err != nil {
		t.Fatalf("read facade.go: %v", err)
	}
	content := string(body)
	disallowed := []string{
		"CreateTimelineRow(",
		"CreateImportedTimelineRow(",
		"PatchTimelineRow(",
		"ResolveTimelineConflict(",
		"ClipboardPaste(ctx context.Context, actor",
		"MarkReviewed(ctx context.Context, actor",
		"Supersede(ctx context.Context, actor",
	}
	for _, snippet := range disallowed {
		if strings.Contains(content, snippet) {
			t.Fatalf("facade.go still exposes legacy Timeline mutation shim %s", snippet)
		}
	}
}

func productionImports(t testing.TB, fileName string) []string {
	t.Helper()
	parsed, err := parser.ParseFile(token.NewFileSet(), filepath.Clean(fileName), nil, parser.ImportsOnly)
	if err != nil {
		t.Fatalf("parse imports for %s: %v", fileName, err)
	}
	imports := make([]string, 0, len(parsed.Imports))
	for _, spec := range parsed.Imports {
		importPath := strings.Trim(spec.Path.Value, "\"")
		if !strings.HasPrefix(importPath, repoImportPrefix) {
			imports = append(imports, importPath)
			continue
		}
		imports = append(imports, importPath)
	}
	return imports
}

func allowedFileNames(files map[string]bool) []string {
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	return names
}
