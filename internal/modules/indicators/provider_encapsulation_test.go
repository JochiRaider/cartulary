package indicators

import (
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	indicatorprojection "github.com/JochiRaider/cartulary/internal/modules/indicators/projectionprovider"
)

const indicatorProviderImportPrefix = "github.com/JochiRaider/cartulary/internal/modules/indicators/internal/providers/"

func TestIndicatorProviderImplementationsAreInternal(t *testing.T) {
	t.Parallel()
	repoRoot := filepath.Clean(filepath.Join("..", "..", ".."))
	for _, relativeRoot := range []string{
		"internal/app",
		"internal/modules/incidentbundles",
		"internal/modules/projections",
		"internal/modules/revisions",
	} {
		root := filepath.Join(repoRoot, relativeRoot)
		if err := filepath.WalkDir(root, func(path string, entry os.DirEntry, err error) error {
			if err != nil {
				return err
			}
			if entry.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
				return nil
			}
			parsed, err := parser.ParseFile(token.NewFileSet(), path, nil, parser.ImportsOnly)
			if err != nil {
				return err
			}
			for _, spec := range parsed.Imports {
				importPath, err := strconv.Unquote(spec.Path.Value)
				if err != nil {
					return err
				}
				if strings.HasPrefix(importPath, indicatorProviderImportPrefix) {
					relative, _ := filepath.Rel(repoRoot, path)
					t.Errorf("generic coordinator %s imports Indicator implementation %s", filepath.ToSlash(relative), importPath)
				}
			}
			return nil
		}); err != nil {
			t.Fatalf("scan %s: %v", relativeRoot, err)
		}
	}

	projection, err := indicatorprojection.NewContribution()
	if err != nil {
		t.Fatalf("construct Indicator projection contribution: %v", err)
	}
	if projection.Source() == nil || len(projection.ProjectionContribution().SurfaceIntents()) != 1 {
		t.Fatalf("incomplete Indicator projection contribution: %#v", projection.ProjectionContribution().SurfaceIntents())
	}
	bundle, err := NewIncidentBundleContribution()
	if err != nil {
		t.Fatalf("construct Indicator incident-bundle contribution: %v", err)
	}
	if bundle.SourcePort == nil || bundle.SubtypePresence.Source == nil {
		t.Fatal("incomplete Indicator incident-bundle contribution")
	}
	revision := NewRevisionContribution()
	if len(revision.Records) != 1 || len(revision.NonRowTargets) != 2 {
		t.Fatalf("incomplete Indicator revision contribution: records=%d non-row=%d", len(revision.Records), len(revision.NonRowTargets))
	}
}
