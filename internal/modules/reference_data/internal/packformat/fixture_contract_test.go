package packformat

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/gen/contractreferencepackfixtures"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

// This is a machine-projection check, not a specification/adoption check.
// Only closed executable inputs are opened; prose is never inspected.
func TestExecutableFixtureContracts_Unit(t *testing.T) {
	t.Run("runtime isolation", testFixtureProjectionIsolation)
	root := "../../../../../"
	raw, err := os.ReadFile(root + "tools/schemas/cartulary.reference_pack_fixture_manifest.v1.schema.json")
	if err != nil {
		t.Fatal(err)
	}
	var schema map[string]any
	if err := json.Unmarshal(raw, &schema); err != nil {
		t.Fatal(err)
	}
	shape := compileShape(schema, map[string]bool{})
	seen := map[string]bool{}
	families := map[string]int{}
	profileCases := map[string]map[string]int{}
	for _, artifact := range contractreferencepackfixtures.Artifacts {
		if !strings.HasPrefix(artifact.Path, "contracts/reference-pack-fixtures/fixture-manifests/") {
			continue
		}
		value, err := canonicaljson.DecodeStrict([]byte(artifact.JSON))
		if err != nil || !shape.matches(value) {
			t.Fatalf("fixture manifest shape: %s %v", artifact.Path, err)
		}
		m := value.(map[string]any)
		id := m["fixture_id"].(string)
		if seen[id] {
			t.Fatal("duplicate fixture identity", id)
		}
		seen[id] = true
		if err := validateFixtureRelations(root, m); err != nil {
			t.Fatal(id, err)
		}
		// Omission never means a default, including nullable members.
		for key := range m {
			original := m[key]
			delete(m, key)
			if shape.matches(m) {
				t.Fatal("omitted fixture field admitted", key)
			}
			m[key] = true
			if shape.matches(m) {
				t.Fatal("wrong fixture field type admitted", key)
			}
			m[key] = original
		}
		m["unknown"] = true
		if shape.matches(m) {
			t.Fatal("unknown fixture field admitted")
		}
		delete(m, "unknown")
		input := fixtureProjection(strings.TrimPrefix(m["input_refs"].([]any)[0].(string), "contracts/reference-pack-fixtures/"))
		var vector map[string]any
		if err := json.Unmarshal(input, &vector); err != nil {
			t.Fatal(err)
		}
		families[m["fixture_family"].(string)]++
		if m["fixture_family"] == "content_profile" {
			key := vector["pack_key"].(string)
			if _, ok := profiles[key]; !ok {
				t.Fatal("fixture names an unsupported profile", key)
			}
			outcome := "valid"
			if expected, ok := m["expected_public_error"].(map[string]any); ok {
				outcome = expected["reason_code"].(string)
			}
			if profileCases[key] == nil {
				profileCases[key] = map[string]int{}
			}
			profileCases[key][outcome]++
		}
		switch id {
		case "rpfx_set_cardinality_64", "rpfx_set_cardinality_65":
			// A complete 64-key set is unreachable in the closed 16-key
			// catalog. Exercise only the independently declared count guard;
			// the complete-set fixture separately admits all reachable keys.
			if len(vector) != 2 || vector["schema_id"] != "cartulary.reference_pack_cardinality_guard_fixture.v1" {
				t.Fatal("invalid cardinality guard fixture")
			}
			count, ok := vector["count"].(float64)
			if !ok || count != 64 && count != 65 {
				t.Fatal("invalid isolated guard count")
			}
			if admittedSetCardinality(int(count)) != (m["expected_public_error"] == nil) {
				t.Fatal("set cardinality guard changed")
			}
			if m["expected_public_error"] != nil {
				if public := m["expected_public_error"].(map[string]any); public["code"] != "reference_pack_verification_failed" || public["reason_code"] != "contract_incompatible" {
					t.Fatal("invalid guard error mapping")
				}
			}
		case "rpfx_manifest_identity":
			manifest, err := DecodeManifest([]byte(vector["canonical_manifest"].(string)), true)
			if err != nil {
				t.Fatal(err)
			}
			payload, err := PayloadDigest(manifest.Files)
			if err != nil || payload != m["expected_payload_sha256"] || Digest([]byte(vector["canonical_manifest"].(string))) != m["expected_manifest_sha256"] {
				t.Fatal("manifest expectation differs", err)
			}
		case "rpfx_complete_set":
			var vector struct {
				Members []SetMember `json:"members"`
			}
			if err := json.Unmarshal(input, &vector); err != nil {
				t.Fatal(err)
			}
			set, err := BuildSet(vector.Members)
			if err != nil || set.SHA256 != m["expected_pack_set_sha256"] {
				t.Fatal("set expectation differs", err)
			}
		case "rpfx_signed_container":
			data, err := base64.StdEncoding.DecodeString(vector["container_base64"].(string))
			if err != nil || Digest(data) != m["expected_container_sha256"] {
				t.Fatal("container expectation differs", err)
			}
			for _, key := range []string{"manifest_sha256", "payload_sha256"} {
				if vector[key] != m["expected_"+key] {
					t.Fatal("independent engine-vector expectation differs", key)
				}
			}
		default:
			if strings.HasPrefix(id, "rpfx_profile_") || strings.HasPrefix(id, "rpfx_type_compatibility_") {
				runProfileManifest(t, input, m)
			} else if strings.HasPrefix(id, "rpfx_dependency_") {
				t.Run(id, func(t *testing.T) { runDependencyManifest(t, input, m) })
			} else if strings.HasPrefix(id, "rpfx_admission_") {
				t.Run(id, func(t *testing.T) { runAdmissionManifest(t, input, m) })
			} else if strings.HasPrefix(id, "rpfx_trust_") {
				t.Run(id, func(t *testing.T) { runTrustManifest(t, input, m) })
			} else if strings.HasPrefix(id, "rpfx_lifecycle_") {
				// The application owner executes these fixtures with PostgreSQL
				// and its real Job finalizer. This pure-format row validates their
				// closed inputs and explicit execution routing, not their effects.
				validateLifecycleFixtureBinding(t, root, id)
			} else if strings.HasPrefix(id, "rpfx_owner_") {
				validateOwnerFixtureBinding(t, root, id)
			} else {
				t.Fatal("fixture has no expectation runner", id)
			}
		}
	}
	if err := fixtureCoverage(shape, families, profileCases); err != nil {
		t.Fatal(err)
	}
	// Removing the last case of any required class must fail even when the
	// remaining corpus is otherwise well formed and every row is routed.
	for family, count := range families {
		families[family] = 0
		if err := fixtureCoverage(shape, families, profileCases); err == nil {
			t.Fatal("missing behavioral family accepted", family)
		}
		families[family] = count
	}
	for key, cases := range profileCases {
		for _, outcome := range []string{"valid", "content_schema_invalid", "content_semantic_invalid"} {
			count := cases[outcome]
			cases[outcome] = 0
			if err := fixtureCoverage(shape, families, profileCases); err == nil {
				t.Fatal("missing profile case accepted", key, outcome)
			}
			cases[outcome] = count
		}
	}
	for _, ref := range []string{"", "/tmp/a", "../a", "contracts//a", "contracts/./a", "contracts/a\\b", strings.Join([]string{"docs", "spec.md"}, "/"), "README", "README.txt", "contracts/notes.MD", "contracts/no\x00"} {
		if fixtureInputPathAllowed(ref) {
			t.Fatal("unsafe/prose fixture input admitted", ref)
		}
	}
	base := fixtureRelationScratch(t)
	for _, test := range []struct {
		refs, expected, forbidden []any
		good                      bool
	}{
		{[]any{"fixtures/input.json"}, []any{}, []any{"network_request_attempted"}, true},
		{[]any{"linked/input.json"}, []any{}, []any{}, false},
		{[]any{"fixtures/input.json", "fixtures/input.json"}, []any{}, []any{}, false},
		{[]any{"fixtures/input.json"}, []any{"job_admitted"}, []any{"job_admitted"}, false},
		{[]any{"fixtures/input.json"}, []any{"trust_state_changed", "job_admitted"}, []any{}, false},
	} {
		err := validateFixtureRelations(base, map[string]any{"input_refs": test.refs, "expected_side_effects": test.expected, "forbidden_side_effects": test.forbidden})
		if (err == nil) != test.good {
			t.Fatal("fixture relation validation", err)
		}
	}
}

// Other owners execute these scenarios through their actual application
// boundary. This check establishes input closure and execution routing only.
func validateOwnerFixtureBinding(t *testing.T, root, id string) {
	t.Helper()
	value, err := canonicaljson.DecodeStrict(fixtureProjection("fixtures/" + strings.TrimPrefix(id, "rpfx_") + ".v1.json"))
	if err != nil || !compileFixtureProjection("owner_fixture.v1.schema.json").matches(value) {
		t.Fatal("invalid owner fixture", err)
	}
	m := value.(map[string]any)
	if m["fixture_id"] != id {
		t.Fatal("owner fixture identity mismatch")
	}
	raw, err := os.ReadFile(root + "tools/test_families/" + m["owner_id"].(string) + ".json")
	if err != nil {
		t.Fatal(err)
	}
	var routing struct {
		Rows []struct {
			ID       string `json:"row_id"`
			Selector struct {
				Tests []string `json:"tests"`
			} `json:"selector"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(raw, &routing); err != nil {
		t.Fatal(err)
	}
	for _, row := range routing.Rows {
		if row.ID == m["row_id"] && slices.Contains(row.Selector.Tests, m["test_name"].(string)) {
			return
		}
	}
	t.Fatal("owner fixture scenario is not routed", id)
}

func validateLifecycleFixtureBinding(t *testing.T, root, id string) {
	t.Helper()
	value, err := canonicaljson.DecodeStrict(fixtureProjection("fixtures/lifecycle.v1.json"))
	if err != nil || !compileFixtureProjection("lifecycle_fixture.v1.schema.json").matches(value) {
		t.Fatal("invalid lifecycle fixture", err)
	}
	found := 0
	for _, raw := range value.(map[string]any)["steps"].([]any) {
		if raw.(map[string]any)["fixture_id"] == id {
			found++
		}
	}
	if found != 1 {
		t.Fatal("missing or duplicate lifecycle case", id)
	}
	raw, err := os.ReadFile(root + "tools/test_families/module.reference_data.json")
	if err != nil {
		t.Fatal(err)
	}
	var routing struct {
		Rows []struct {
			ID       string `json:"row_id"`
			Selector struct {
				Tests []string `json:"tests"`
			} `json:"selector"`
		} `json:"rows"`
	}
	if err := json.Unmarshal(raw, &routing); err != nil {
		t.Fatal(err)
	}
	found = 0
	for _, row := range routing.Rows {
		if row.ID == "module.reference_data.integration.canonical_persistence" && slices.Contains(row.Selector.Tests, "TestCanonicalLifecycleFixtureManifests_Integration") {
			found++
		}
	}
	if found != 1 {
		t.Fatal("lifecycle fixture execution is not routed")
	}
}

func fixtureInputPathAllowed(ref string) bool {
	if ref == "" || strings.ContainsAny(ref, "\\\x00\r\n\t") || strings.HasPrefix(ref, "/") {
		return false
	}
	for _, part := range strings.Split(ref, "/") {
		if part == "" || part == "." || part == ".." || strings.EqualFold(part, "docs") || strings.EqualFold(part, "README") || strings.HasPrefix(strings.ToLower(part), "readme.") || strings.HasSuffix(strings.ToLower(part), ".md") || strings.HasSuffix(strings.ToLower(part), ".markdown") {
			return false
		}
	}
	return true
}
func validateFixtureRelations(root string, m map[string]any) error {
	sets := map[string]map[string]bool{}
	for _, key := range []string{"input_refs", "expected_side_effects", "forbidden_side_effects"} {
		sets[key] = map[string]bool{}
		prior := ""
		for _, item := range m[key].([]any) {
			v := item.(string)
			if v <= prior {
				return fmt.Errorf("unordered or duplicate %s", key)
			}
			prior = v
			sets[key][v] = true
			if key == "input_refs" {
				// Reject prose and malformed paths BEFORE any filesystem operation.
				if !fixtureInputPathAllowed(v) {
					return fmt.Errorf("invalid fixture input")
				}
				path := root
				segments := strings.Split(v, "/")
				for i, part := range segments {
					path = filepath.Join(path, part)
					info, err := os.Lstat(path)
					if err != nil || info.Mode()&os.ModeSymlink != 0 {
						return fmt.Errorf("missing or linked fixture input")
					}
					if i == len(segments)-1 && !info.Mode().IsRegular() {
						return fmt.Errorf("nonregular fixture input")
					}
				}
			}
		}
	}
	for key := range sets["expected_side_effects"] {
		if sets["forbidden_side_effects"][key] {
			return fmt.Errorf("contradictory fixture effects")
		}
	}
	return nil
}

// Coverage follows executable family and profile declarations, not a fixed
// inventory total or a second catalog of prose requirement identifiers.
func fixtureCoverage(manifestShape *shape, families map[string]int, cases map[string]map[string]int) error {
	for _, value := range manifestShape.properties["fixture_family"].enum {
		family := value.(string)
		if families[family] == 0 {
			return fmt.Errorf("missing fixture family %s", family)
		}
	}
	for key := range profiles {
		for _, outcome := range []string{"valid", "content_schema_invalid", "content_semantic_invalid"} {
			if cases[key][outcome] == 0 {
				return fmt.Errorf("missing %s fixture for profile %s", outcome, key)
			}
		}
	}
	return nil
}

// Expectations are checked-in independent bytes. This runner never derives
// expected verdicts, issue identities or diagnostic fields from its own result.
func runProfileManifest(t *testing.T, input []byte, manifest map[string]any) {
	t.Helper()
	value, err := canonicaljson.DecodeStrict(input)
	if err != nil || !compileFixtureProjection("content_fixture.v1.schema.json").matches(value) {
		t.Fatal("invalid content fixture", err)
	}
	var fixture contentFixture
	if err := json.Unmarshal(input, &fixture); err != nil {
		t.Fatal(err)
	}
	var rows contentRows
	err = ValidateContent(context.Background(), fixtureManifest(fixture), contentMemory(fixture.Members), &rows)
	expected := manifest["expected_public_error"]
	if expected == nil {
		if err != nil || len(rows) == 0 || len(manifest["expected_issues"].([]any)) != 0 {
			t.Fatal("canonical content fixture rejected", fixture.Key, err)
		}
		return
	}
	public := expected.(map[string]any)
	var failure *Failure
	if !errors.As(err, &failure) || public["code"] != "reference_pack_verification_failed" || failure.Code != public["reason_code"] || len(rows) != 0 {
		t.Fatal("content fixture winner or index side effect changed", fixture.Key, err)
	}
	check := "content_schema"
	if failure.Code == "content_semantic_invalid" {
		check = "content_semantics"
	} else if failure.Code == "type_registry_incompatible" {
		check = "type_registry"
	}
	summary, err := CheckSummary(context.Background(), check, newDiagnosticTestScratch(), func(ctx context.Context, emit FindingSink) error {
		if check == "content_schema" {
			return ContentSchemaFindings(ctx, fixtureManifest(fixture), contentMemory(fixture.Members), emit)
		}
		if check == "type_registry" {
			return RegistryCompatibilityFindings(ctx, fixtureManifest(fixture), contentMemory(fixture.Members), emit)
		}
		return ContentSemanticsFindings(ctx, fixtureManifest(fixture), contentMemory(fixture.Members), emit)
	})
	if err != nil {
		t.Fatal(err)
	}
	got, err := canonicaljson.Marshal(summary.Issues)
	if err != nil {
		t.Fatal(err)
	}
	want, err := canonicaljson.Marshal(manifest["expected_issues"])
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, want) {
		t.Fatalf("profile diagnostics differ: %s\ngot %s\nwant %s", fixture.Key, got, want)
	}
}
