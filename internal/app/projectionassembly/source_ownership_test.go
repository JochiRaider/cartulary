package projectionassembly

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/projections/providercontract"
)

type providerSQLSource struct {
	path     string
	section  string
	provider string
}

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

type providerSQLReference struct {
	operation string
	table     string
}

type projectionBoundaryManifest struct {
	SQLTableAccess []struct {
		ID                    string   `json:"id"`
		Table                 string   `json:"table"`
		TestReadAllowedPaths  []string `json:"test_read_allowed_paths"`
		TestWriteAllowedPaths []string `json:"test_write_allowed_paths"`
	} `json:"sql_table_access"`
}

var providerSQLReferencePattern = regexp.MustCompile(`(?i)\b(FROM|JOIN|UPDATE|INSERT\s+INTO|DELETE\s+FROM)\s+([a-z_][a-z0-9_]*)`)
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

func TestProjectionProviderSQLSourceOwnership(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	ownerPatterns := loadSchemaOwnerPatterns(t, root)
	bundle, err := Build(&projectionManifestDB{})
	if err != nil {
		t.Fatalf("assemble projection adapter: %v", err)
	}
	descriptors := map[string]providercontract.ProviderDescriptor{}
	for _, descriptor := range bundle.DescriptorSet().All() {
		descriptors[descriptor.ProviderID] = descriptor
	}

	sources := []providerSQLSource{
		{path: "internal/modules/timeline/sourcerepository/repository.go", provider: "timeline"},
		{path: "internal/modules/entities/timelinefacts/reader.go", provider: "timeline"},
		{path: "internal/modules/links/active_facts.go", provider: "timeline"},
		{path: "internal/modules/evidence/timeline_facts.go", provider: "timeline"},
		{path: "internal/modules/projections/internal/storage/timeline.go", provider: "timeline"},
		{path: "internal/modules/projections/internal/queryengine/timeline.go", provider: "timeline"},
		{path: "internal/modules/entities/hostidentity/projectionprovider/host_source.go", provider: "host"},
		{path: "internal/modules/projections/internal/storage/host.go", provider: "host"},
		{path: "internal/modules/projections/internal/queryengine/host.go", provider: "host"},
		{path: "internal/modules/entities/hostidentity/projectionprovider/identity_source.go", provider: "identity"},
		{path: "internal/modules/projections/internal/storage/identity.go", provider: "identity"},
		{path: "internal/modules/projections/internal/queryengine/identity.go", provider: "identity"},
		{path: "internal/modules/indicators/internal/providers/projection/source.go", provider: "indicator"},
		{path: "internal/modules/projections/internal/storage/indicator.go", provider: "indicator"},
		{path: "internal/modules/projections/internal/queryengine/indicator.go", provider: "indicator"},
		{path: "internal/modules/assessments/internal/providers/projection/provider.go", provider: "assessment"},
		{path: "internal/modules/projections/internal/storage/assessment.go", provider: "assessment"},
		{path: "internal/modules/projections/internal/queryengine/assessment.go", provider: "assessment"},
		{path: "internal/modules/artifacts/internal/providers/projection/source.go", provider: "artifact"},
		{path: "internal/modules/projections/internal/storage/artifact.go", provider: "artifact"},
		{path: "internal/modules/projections/internal/queryengine/artifact.go", provider: "artifact"},
		{path: "internal/modules/evidence/internal/providers/projection/source.go", provider: "evidence"},
		{path: "internal/modules/projections/internal/storage/evidence.go", provider: "evidence"},
		{path: "internal/modules/projections/internal/queryengine/evidence.go", provider: "evidence"},
		{path: "internal/modules/parties/internal/providers/projection/source.go", provider: "party"},
		{path: "internal/modules/projections/internal/storage/party.go", provider: "party"},
		{path: "internal/modules/projections/internal/queryengine/party.go", provider: "party"},
		{path: "internal/modules/tasksdecisions/internal/providers/projection/task_source.go", provider: "task_request"},
		{path: "internal/modules/projections/internal/storage/task_request.go", provider: "task_request"},
		{path: "internal/modules/projections/internal/queryengine/task_request.go", provider: "task_request"},
		{path: "internal/modules/tasksdecisions/internal/providers/projection/decision_source.go", provider: "decision"},
		{path: "internal/modules/projections/internal/storage/decision.go", provider: "decision"},
		{path: "internal/modules/projections/internal/queryengine/decision.go", provider: "decision"},
	}
	covered := map[string]bool{}
	for _, source := range sources {
		descriptor, ok := descriptors[source.provider]
		if !ok {
			t.Fatalf("SQL source %s names unknown provider %q", source.path, source.provider)
		}
		covered[source.provider] = true
		body, err := os.ReadFile(filepath.Join(root, source.path))
		if err != nil {
			t.Fatalf("read provider SQL source %s: %v", source.path, err)
		}
		sqlText := string(body)
		if source.section != "" {
			sqlText, err = namedSQLSection(sqlText, source.section)
			if err != nil {
				t.Fatalf("read %s section %s: %v", source.path, source.section, err)
			}
		}
		for _, ref := range providerSQLReferences(sqlText) {
			if err := validateProviderSQLReference(descriptor, ref, ownerPatterns); err != nil {
				t.Errorf("%s provider=%s: %v", source.path, source.provider, err)
			}
		}
	}
	for providerKey := range descriptors {
		if !covered[providerKey] {
			t.Errorf("provider %q has no SQL ownership source coverage", providerKey)
		}
	}
}

func TestValidateProjectionProviderSQLReference(t *testing.T) {
	patterns := []schemaOwnerPattern{
		{owner: "records", pattern: regexp.MustCompile(`^records$`)},
		{owner: "projections", pattern: regexp.MustCompile(`^host_grid_projection$`)},
	}
	descriptor := providercontract.ProviderDescriptor{
		ProviderID:             "host",
		SourceAuthorityModules: []string{"entities"},
		ProjectionTableIDs:     []string{"host_grid_projection"},
	}
	tests := []struct {
		name string
		ref  providerSQLReference
		want string
	}{
		{name: "undeclared authoritative read", ref: providerSQLReference{operation: "FROM", table: "records"}, want: "undeclared source authority"},
		{name: "authoritative write", ref: providerSQLReference{operation: "UPDATE", table: "records"}, want: "writes authoritative"},
		{name: "projection write", ref: providerSQLReference{operation: "INSERT INTO", table: "host_grid_projection"}},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			err := validateProviderSQLReference(descriptor, tc.ref, patterns)
			if tc.want == "" && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if tc.want != "" && (err == nil || !strings.Contains(err.Error(), tc.want)) {
				t.Fatalf("error = %v, want containing %q", err, tc.want)
			}
		})
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

func namedSQLSection(body string, name string) (string, error) {
	marker := "-- name: " + name + " "
	start := strings.Index(body, marker)
	if start < 0 {
		return "", fmt.Errorf("missing named query")
	}
	remainder := body[start+len(marker):]
	if next := strings.Index(remainder, "\n-- name: "); next >= 0 {
		remainder = remainder[:next]
	}
	return remainder, nil
}

func providerSQLReferences(body string) []providerSQLReference {
	refs := make([]providerSQLReference, 0)
	seen := map[providerSQLReference]struct{}{}
	for _, match := range providerSQLReferencePattern.FindAllStringSubmatch(body, -1) {
		operation := strings.ToUpper(strings.Join(strings.Fields(match[1]), " "))
		table := strings.ToLower(match[2])
		// ON CONFLICT DO UPDATE SET is not a table mutation. The deliberately
		// small scanner sees "UPDATE SET"; discard that clause rather than
		// pretending SET is an owner-controlled table.
		if table == "lateral" || (operation == "UPDATE" && table == "set") {
			continue
		}
		ref := providerSQLReference{operation: operation, table: table}
		if _, exists := seen[ref]; exists {
			continue
		}
		seen[ref] = struct{}{}
		refs = append(refs, ref)
	}
	sort.Slice(refs, func(left, right int) bool {
		if refs[left].table == refs[right].table {
			return refs[left].operation < refs[right].operation
		}
		return refs[left].table < refs[right].table
	})
	return refs
}

func validateProviderSQLReference(descriptor providercontract.ProviderDescriptor, ref providerSQLReference, patterns []schemaOwnerPattern) error {
	owners := map[string]struct{}{}
	for _, pattern := range patterns {
		if pattern.pattern.MatchString(ref.table) {
			owners[pattern.owner] = struct{}{}
		}
	}
	if len(owners) != 1 {
		return fmt.Errorf("SQL %s %s maps to %d schema owners", ref.operation, ref.table, len(owners))
	}
	owner := ""
	for value := range owners {
		owner = value
	}
	if isSQLWriteOperation(ref.operation) {
		if owner != "projections" {
			return fmt.Errorf("SQL %s %s writes authoritative owner %q", ref.operation, ref.table, owner)
		}
		if !slices.Contains(descriptor.ProjectionTableIDs, ref.table) {
			return fmt.Errorf("SQL %s %s writes undeclared projection table", ref.operation, ref.table)
		}
		return nil
	}
	if owner == "projections" {
		if !slices.Contains(descriptor.ProjectionTableIDs, ref.table) {
			return fmt.Errorf("SQL %s %s reads undeclared projection table", ref.operation, ref.table)
		}
		return nil
	}
	for _, declared := range descriptor.SourceAuthorityModules {
		if owner == declared {
			return nil
		}
	}
	return fmt.Errorf("SQL %s %s reads undeclared source authority %q", ref.operation, ref.table, owner)
}

func isSQLWriteOperation(operation string) bool {
	switch operation {
	case "UPDATE", "INSERT INTO", "DELETE FROM":
		return true
	default:
		return false
	}
}
