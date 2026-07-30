package projectionassembly

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/projections"
)

type providerSQLSource struct {
	path     string
	section  string
	provider string
}

type schemaOwnershipManifest struct {
	Entries []struct {
		Owner          string `json:"owner"`
		ObjectPatterns []struct {
			NamePattern string `json:"name_pattern"`
		} `json:"object_patterns"`
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

var providerSQLReferencePattern = regexp.MustCompile(`(?i)\b(FROM|JOIN|UPDATE|INSERT\s+INTO|DELETE\s+FROM)\s+([a-z_][a-z0-9_]*)`)

func TestProjectionProviderSQLSourceOwnership(t *testing.T) {
	root := filepath.Join("..", "..", "..")
	ownerPatterns := loadSchemaOwnerPatterns(t, root)
	catalog, err := NewCatalog(nil)
	if err != nil {
		t.Fatalf("assemble projection catalog: %v", err)
	}
	descriptors := map[string]projections.ProviderDescriptor{}
	for _, descriptor := range catalog.Descriptors() {
		descriptors[descriptor.ProviderKey] = descriptor
	}

	sources := []providerSQLSource{
		{path: "internal/modules/timeline/sourcerepository/repository.go", provider: "timeline"},
		{path: "internal/modules/entities/timelinefacts/reader.go", provider: "timeline"},
		{path: "internal/modules/links/timeline_facts.go", provider: "timeline"},
		{path: "internal/modules/evidence/timeline_facts.go", provider: "timeline"},
		{path: "internal/modules/entities/hostidentity/projectionprovider/provider.go", provider: "host"},
		{path: "internal/modules/entities/hostidentity/projectionprovider/provider.go", provider: "identity"},
		{path: "internal/modules/indicators/projectionprovider/provider.go", provider: "indicator"},
		{path: "internal/modules/indicators/projectionprovider/query_surfaces.go", provider: "indicator"},
		{path: "internal/modules/assessments/projectionprovider/provider.go", provider: "assessment"},
		{path: "internal/modules/assessments/projectionprovider/query_surfaces.go", provider: "assessment"},
		{path: "internal/modules/artifacts/projectionprovider/provider.go", provider: "artifact"},
		{path: "internal/modules/artifacts/projectionprovider/query_surfaces.go", provider: "artifact"},
		{path: "internal/modules/evidence/projectionprovider/provider.go", provider: "evidence"},
		{path: "internal/modules/evidence/projectionprovider/query_surfaces.go", provider: "evidence"},
		{path: "internal/modules/parties/projectionprovider/provider.go", provider: "party"},
		{path: "internal/modules/parties/projectionprovider/query_surfaces.go", provider: "party"},
		{path: "internal/modules/tasksdecisions/projectionprovider/provider.go", provider: "task_request"},
		{path: "internal/modules/tasksdecisions/projectionprovider/query_surfaces.go", provider: "task_request"},
		{path: "internal/modules/tasksdecisions/projectionprovider/provider.go", provider: "decision"},
		{path: "internal/modules/tasksdecisions/projectionprovider/query_surfaces.go", provider: "decision"},
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
	descriptor := projections.ProviderDescriptor{
		ProviderKey:            "host",
		SourceAuthorityModules: []string{"entities"},
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
		for _, objectPattern := range entry.ObjectPatterns {
			if objectPattern.NamePattern == "" {
				continue
			}
			compiled, err := regexp.Compile(objectPattern.NamePattern)
			if err != nil {
				t.Fatalf("compile schema owner %s pattern %q: %v", entry.Owner, objectPattern.NamePattern, err)
			}
			patterns = append(patterns, schemaOwnerPattern{owner: entry.Owner, pattern: compiled})
		}
	}
	return patterns
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

func validateProviderSQLReference(descriptor projections.ProviderDescriptor, ref providerSQLReference, patterns []schemaOwnerPattern) error {
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
		return nil
	}
	if owner == "projections" {
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
