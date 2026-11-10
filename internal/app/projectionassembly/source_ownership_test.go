package projectionassembly

import (
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/projections/providercontract"
)

type schemaOwnershipManifest struct {
	Entries []struct {
		ObjectKind      string `json:"object_kind"`
		QualifiedName   string `json:"qualified_name"`
		ManagementClass string `json:"management_class"`
		SourceOwner     string `json:"source_owner"`
	} `json:"entries"`
}

type schemaOwnerPattern struct {
	owner   string
	pattern *regexp.Regexp
}

type projectionBoundaryManifest struct {
	SQLTableAccess []struct {
		ID                    string   `json:"id"`
		Table                 string   `json:"table"`
		TestReadAllowedPaths  []string `json:"test_read_allowed_paths"`
		TestWriteAllowedPaths []string `json:"test_write_allowed_paths"`
	} `json:"sql_table_access"`
}

var migrationCreateTablePattern = regexp.MustCompile(`(?im)\bCREATE\s+TABLE(?:\s+IF\s+NOT\s+EXISTS)?\s+(?:public\.)?([a-z_][a-z0-9_]*)\b`)

func TestProjectionTableOwnershipSetsAreExactlyEqual(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	bundle, err := Build(&projectionManifestDB{})
	if err != nil {
		t.Fatalf("assemble projection adapter: %v", err)
	}

	descriptorTables := make([]string, 0, bundle.DescriptorSet().Len())
	for _, descriptor := range bundle.DescriptorSet().All() {
		if descriptor.Status == providercontract.ProviderStatusActive {
			descriptorTables = append(descriptorTables, descriptor.ProjectionTableIDs...)
		}
	}
	descriptorTables = sortedUniqueStrings(t, "active descriptor tables", descriptorTables)

	boundaryData, err := os.ReadFile(filepath.Join(root, "tools", "backend_module_boundaries.json"))
	if err != nil {
		t.Fatalf("read backend boundary manifest: %v", err)
	}
	var boundary projectionBoundaryManifest
	if err := json.Unmarshal(boundaryData, &boundary); err != nil {
		t.Fatalf("decode backend boundary manifest: %v", err)
	}
	boundaryTables := make([]string, 0, len(descriptorTables))
	fixturePermissions := map[string][]string{}
	for _, rule := range boundary.SQLTableAccess {
		if !strings.HasSuffix(rule.ID, "-projection-storage-access") {
			continue
		}
		boundaryTables = append(boundaryTables, rule.Table)
		fixturePermissions[rule.Table] = append(
			append([]string(nil), rule.TestReadAllowedPaths...),
			rule.TestWriteAllowedPaths...,
		)
	}
	boundaryTables = sortedUniqueStrings(t, "projection table policies", boundaryTables)

	migrationTables := migrationOwnedProjectionTables(t, root, loadSchemaOwnerPatterns(t, root))
	recoveryTables := sortedUniqueStrings(t, "recovery projection tables", providercontract.RecoveryProjectionTableIDs())
	for label, actual := range map[string][]string{
		"boundary policy":  boundaryTables,
		"recovery state":   recoveryTables,
		"migration schema": migrationTables,
	} {
		if !slices.Equal(descriptorTables, actual) {
			t.Fatalf("active descriptor tables differ from %s:\ndescriptors=%q\n%s=%q", label, descriptorTables, label, actual)
		}
	}

	wantFixturePermissions := map[string][]string{
		"assessment_grid_projection": {"internal/modules/projections/testsupport/capability.go"},
		"timeline_grid_projection":   {"internal/modules/timeline/testsupport/asserttest/assertions.go"},
	}
	for table, paths := range fixturePermissions {
		sort.Strings(paths)
		want := wantFixturePermissions[table]
		sort.Strings(want)
		if !slices.Equal(paths, want) {
			t.Fatalf("projection table %q test-fixture permissions = %q, want %q", table, paths, want)
		}
	}
}

func loadSchemaOwnerPatterns(t testing.TB, root string) []schemaOwnerPattern {
	t.Helper()
	data, err := os.ReadFile(filepath.Join(root, "tools", "schema_object_ownership_manifest.json"))
	if err != nil {
		t.Fatalf("read schema ownership manifest: %v", err)
	}
	var manifest schemaOwnershipManifest
	if err := json.Unmarshal(data, &manifest); err != nil {
		t.Fatalf("decode schema ownership manifest: %v", err)
	}
	patterns := make([]schemaOwnerPattern, 0)
	for _, entry := range manifest.Entries {
		if entry.ManagementClass != "cartulary_authored" || (entry.ObjectKind != "table" && entry.ObjectKind != "view") {
			continue
		}
		name := strings.TrimPrefix(entry.QualifiedName, "public.")
		if name == "" || name == entry.QualifiedName {
			t.Fatalf("schema owner %s has invalid qualified relation %q", entry.SourceOwner, entry.QualifiedName)
		}
		patterns = append(patterns, schemaOwnerPattern{
			owner:   entry.SourceOwner,
			pattern: regexp.MustCompile("^" + regexp.QuoteMeta(name) + "$"),
		})
	}
	return patterns
}

func migrationOwnedProjectionTables(t testing.TB, root string, patterns []schemaOwnerPattern) []string {
	t.Helper()
	migrationPaths, err := filepath.Glob(filepath.Join(root, "db", "migrations", "*.sql"))
	if err != nil {
		t.Fatalf("list migrations: %v", err)
	}
	tables := make([]string, 0)
	for _, migrationPath := range migrationPaths {
		body, err := os.ReadFile(migrationPath)
		if err != nil {
			t.Fatalf("read migration %s: %v", migrationPath, err)
		}
		for _, match := range migrationCreateTablePattern.FindAllStringSubmatch(string(body), -1) {
			table := strings.ToLower(match[1])
			owners := map[string]struct{}{}
			for _, pattern := range patterns {
				if pattern.pattern.MatchString(table) {
					owners[pattern.owner] = struct{}{}
				}
			}
			if _, owned := owners["projections"]; !owned {
				continue
			}
			if len(owners) != 1 {
				t.Fatalf("migration-created projection table %q maps to %d schema owners", table, len(owners))
			}
			tables = append(tables, table)
		}
	}
	return sortedUniqueStrings(t, "migration-owned projection tables", tables)
}

func sortedUniqueStrings(t testing.TB, label string, values []string) []string {
	t.Helper()
	result := append([]string(nil), values...)
	sort.Strings(result)
	for index := 1; index < len(result); index++ {
		if result[index] == result[index-1] {
			t.Fatalf("%s contain duplicate %q", label, result[index])
		}
	}
	return result
}
