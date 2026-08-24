package database_migrations_test

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"
	"testing"
)

var (
	ddlObjectPattern    = regexp.MustCompile(`(?i)\b(?:CREATE|ALTER|DROP)\s+(?:MATERIALIZED\s+)?(?:TABLE|VIEW|SEQUENCE|FUNCTION)\s+(?:ONLY\s+)?(?:IF\s+EXISTS\s+)?([A-Za-z_][A-Za-z0-9_.]*)`)
	ddlReferencePattern = regexp.MustCompile(`(?is)\bREFERENCES\s+([A-Za-z_][A-Za-z0-9_.]*)\s*\([^)]*\)([\s\S]{0,160})`)
	ddlRoutinePattern   = regexp.MustCompile(`(?is)CREATE\s+FUNCTION\s+(public\.([A-Za-z_][A-Za-z0-9_]*))\s*\([\s\S]*?\$\$;`)
)

func TestProductionDDLCatalogStructuralPolicy(t *testing.T) {
	root := repositoryRoot(t)
	files, err := filepath.Glob(filepath.Join(root, "db", "migrations", "*.sql"))
	if err != nil {
		t.Fatal(err)
	}
	sort.Strings(files)
	if len(files) < 29 {
		t.Fatalf("migration file count = %d, want at least immutable baseline 29", len(files))
	}
	for index, filename := range files {
		wantPrefix := fmt.Sprintf("%05d_", index+1)
		if !strings.HasPrefix(filepath.Base(filename), wantPrefix) {
			t.Fatalf("migration %d = %q, want prefix %q", index, filepath.Base(filename), wantPrefix)
		}
		body, err := os.ReadFile(filename)
		if err != nil {
			t.Fatal(err)
		}
		if violations := productionDDLPolicyViolations(string(body)); len(violations) != 0 {
			t.Fatalf("%s violates structural policy: %v", filepath.Base(filename), violations)
		}
	}
}

func TestProductionDDLCatalogStructuralPolicyRejectsRepresentativeViolations(t *testing.T) {
	const valid = `-- +goose Up
CREATE TABLE public.sample (
    id bigint CONSTRAINT sample_pkey PRIMARY KEY
);
-- +goose Down
DROP TABLE public.sample;
`
	tests := []struct {
		name string
		body string
		want string
	}{
		{name: "missing down", body: strings.ReplaceAll(valid, "-- +goose Down\n", ""), want: "marker_pair"},
		{name: "nontransactional", body: strings.Replace(valid, "-- +goose Up", "-- +goose Up\n-- +goose NO TRANSACTION", 1), want: "no_transaction"},
		{name: "permissive create", body: strings.Replace(valid, "CREATE TABLE", "CREATE TABLE IF NOT EXISTS", 1), want: "permissive_create"},
		{name: "extension install", body: strings.Replace(valid, "CREATE TABLE public.sample (", "CREATE EXTENSION citext;\nCREATE TABLE public.sample (", 1), want: "extension_install"},
		{name: "unqualified object", body: strings.ReplaceAll(valid, "public.sample", "sample"), want: "unqualified_object"},
		{name: "unqualified reference", body: strings.Replace(valid, "id bigint CONSTRAINT sample_pkey PRIMARY KEY", "id bigint CONSTRAINT sample_pkey PRIMARY KEY, parent_id bigint CONSTRAINT sample_parent_fkey REFERENCES parent(id) ON UPDATE NO ACTION ON DELETE NO ACTION", 1), want: "unqualified_reference"},
		{name: "implicit fk actions", body: strings.Replace(valid, "id bigint CONSTRAINT sample_pkey PRIMARY KEY", "id bigint CONSTRAINT sample_pkey PRIMARY KEY, parent_id bigint CONSTRAINT sample_parent_fkey REFERENCES public.parent(id)", 1), want: "foreign_key_actions"},
		{name: "unnamed constraint", body: strings.Replace(valid, "CONSTRAINT sample_pkey ", "", 1), want: "unnamed_constraint"},
		{name: "unvalidated constraint", body: strings.Replace(valid, ");", ") NOT VALID;", 1), want: "unvalidated_constraint"},
		{name: "concurrent index", body: strings.Replace(valid, "-- +goose Down", "CREATE INDEX CONCURRENTLY sample_idx ON public.sample (id);\n-- +goose Down", 1), want: "concurrent_index"},
		{name: "cascade cleanup", body: strings.Replace(valid, "DROP TABLE public.sample;", "DROP TABLE public.sample CASCADE;", 1), want: "cascade_drop"},
		{name: "replacement ddl", body: strings.Replace(valid, "CREATE TABLE public.sample (", "CREATE OR REPLACE VIEW public.compatibility_view AS SELECT 1;\nCREATE TABLE public.sample (", 1), want: "replacement_ddl"},
		{name: "historical lineage", body: strings.Replace(valid, "CREATE TABLE", "SELECT 'cartulary.prod_ddl_rebaseline.v1';\nCREATE TABLE", 1), want: "historical_lineage"},
		{name: "routine path", body: routineFixture("", "REVOKE ALL ON FUNCTION public.sample_routine() FROM PUBLIC;"), want: "routine_search_path"},
		{name: "routine public execute", body: routineFixture("SET search_path = pg_catalog, public", ""), want: "routine_public_execute"},
		{name: "unsafe definer", body: strings.Replace(routineFixture("SET search_path = pg_catalog, public", "REVOKE ALL ON FUNCTION public.sample_routine() FROM PUBLIC;"), "LANGUAGE plpgsql", "LANGUAGE plpgsql SECURITY DEFINER", 1), want: "routine_security_class"},
		{name: "unsafe dynamic sql", body: strings.Replace(routineFixture("SET search_path = pg_catalog, public", "REVOKE ALL ON FUNCTION public.sample_routine() FROM PUBLIC;"), "RETURN 1;", "EXECUTE 'SELECT 1'; RETURN 1;", 1), want: "routine_dynamic_sql"},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			violations := productionDDLPolicyViolations(test.body)
			if !containsString(violations, test.want) {
				t.Fatalf("violations = %v, want %q", violations, test.want)
			}
		})
	}
}

func productionDDLPolicyViolations(body string) []string {
	violations := make(map[string]struct{})
	add := func(code string) { violations[code] = struct{}{} }
	if strings.Count(body, "-- +goose Up") != 1 || strings.Count(body, "-- +goose Down") != 1 || strings.Index(body, "-- +goose Up") > strings.Index(body, "-- +goose Down") {
		add("marker_pair")
	}
	upper := strings.ToUpper(body)
	for _, check := range []struct {
		code    string
		pattern *regexp.Regexp
	}{
		{"no_transaction", regexp.MustCompile(`--\s*\+GOOSE\s+NO\s+TRANSACTION`)},
		{"permissive_create", regexp.MustCompile(`\bIF\s+NOT\s+EXISTS\b`)},
		{"extension_install", regexp.MustCompile(`\bCREATE\s+EXTENSION\b`)},
		{"replacement_ddl", regexp.MustCompile(`\bCREATE\s+OR\s+REPLACE\b`)},
		{"unvalidated_constraint", regexp.MustCompile(`\bNOT\s+VALID\b`)},
		{"concurrent_index", regexp.MustCompile(`\bCREATE\s+(?:UNIQUE\s+)?INDEX\s+CONCURRENTLY\b`)},
		{"cascade_drop", regexp.MustCompile(`(?s)\bDROP\b[^;]*\bCASCADE\b`)},
		{"historical_lineage", regexp.MustCompile(`CARTULARY\.PROD_DDL_REBASELINE\.V1|00062_|MIGRATION\s+62`)},
	} {
		if check.pattern.MatchString(upper) {
			add(check.code)
		}
	}
	up := strings.SplitN(body, "-- +goose Down", 2)[0]
	for _, match := range ddlObjectPattern.FindAllStringSubmatch(up, -1) {
		if !strings.Contains(match[1], ".") {
			add("unqualified_object")
		}
	}
	for _, match := range ddlReferencePattern.FindAllStringSubmatch(up, -1) {
		if !strings.Contains(match[1], ".") {
			add("unqualified_reference")
		}
		actions := strings.ToUpper(strings.SplitN(match[2], ",", 2)[0])
		if !strings.Contains(actions, "ON UPDATE ") || !strings.Contains(actions, "ON DELETE ") {
			add("foreign_key_actions")
		}
	}
	lines := strings.Split(up, "\n")
	for index, line := range lines {
		upperLine := strings.ToUpper(line)
		if strings.HasPrefix(strings.TrimSpace(line), "--") {
			continue
		}
		if (!strings.Contains(upperLine, "PRIMARY KEY") && !strings.Contains(upperLine, "UNIQUE")) || strings.Contains(upperLine, "CREATE UNIQUE INDEX") {
			continue
		}
		previous := ""
		if index > 0 {
			previous = lines[index-1]
		}
		if !regexp.MustCompile(`(?i)\bCONSTRAINT\s+[A-Za-z_][A-Za-z0-9_]*\b`).MatchString(previous + "\n" + line) {
			add("unnamed_constraint")
		}
	}
	allowedDefiners := map[string]bool{
		"revisions_incident_bundle_sequence_begin_v1":    true,
		"revisions_incident_bundle_sequence_finish_v1":   true,
		"entities_refresh_active_identifier_claims_v1":   true,
		"entities_release_active_identifier_claims_v1":   true,
		"entities_sync_active_identifier_claims_v1":      true,
		"entities_rebuild_active_identifier_claims_v1":   true,
		"entities_active_identifier_claims_are_valid_v1": true,
		"parties_refresh_active_key_claims_v1":           true,
		"parties_release_active_key_claims_v1":           true,
		"parties_sync_active_key_claims_v1":              true,
		"parties_rebuild_active_key_claims_v1":           true,
		"parties_active_key_claims_are_valid_v1":         true,
	}
	for _, match := range ddlRoutinePattern.FindAllStringSubmatch(up, -1) {
		definition := match[0]
		name := match[2]
		if !regexp.MustCompile(`(?i)SET\s+search_path\s*=\s*pg_catalog\s*,\s*public`).MatchString(definition) {
			add("routine_search_path")
		}
		revoke := regexp.MustCompile(`(?i)REVOKE\s+(?:ALL|EXECUTE)\s+ON\s+FUNCTION\s+public\.` + regexp.QuoteMeta(name) + `\s*\(`)
		if !revoke.MatchString(up) {
			add("routine_public_execute")
		}
		definer := regexp.MustCompile(`(?i)\bSECURITY\s+DEFINER\b`).MatchString(definition)
		if definer && !allowedDefiners[name] {
			add("routine_security_class")
		}
		if regexp.MustCompile(`(?i)\bEXECUTE\b`).MatchString(definition) && (!definer || !regexp.MustCompile(`(?i)EXECUTE\s+pg_catalog\.format\s*\(`).MatchString(definition)) {
			add("routine_dynamic_sql")
		}
	}
	result := make([]string, 0, len(violations))
	for violation := range violations {
		result = append(result, violation)
	}
	sort.Strings(result)
	return result
}

func routineFixture(searchPath string, revoke string) string {
	return fmt.Sprintf(`-- +goose Up
CREATE FUNCTION public.sample_routine() RETURNS bigint
LANGUAGE plpgsql
%s
AS $$
BEGIN
    RETURN 1;
END;
$$;
%s
-- +goose Down
DROP FUNCTION public.sample_routine();
`, searchPath, revoke)
}

func repositoryRoot(t testing.TB) string {
	t.Helper()
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		t.Fatal("resolve repository root")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(filename), "..", "..", ".."))
}

func containsString(values []string, want string) bool {
	for _, value := range values {
		if value == want {
			return true
		}
	}
	return false
}
