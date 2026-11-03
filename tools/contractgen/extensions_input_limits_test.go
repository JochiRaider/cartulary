package main

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/extensions/registryfacts"
)

func checkGeneratorInputLimits(t *testing.T, root string) {
	t.Run("actual capacities", func(t *testing.T) { checkGeneratorInputCapacities(t, root) })
	inputs, err := readExtensionInputCatalog(root)
	if err != nil {
		t.Fatal(err)
	}
	t.Run("empty fragment", func(t *testing.T) {
		jobs := inputs["fragments/core01.profile-jobs.json"]["facts"].([]any)
		fixture := extensionInputFixture(t, root, func(path string, object map[string]any) {
			switch path {
			case "contracts/extensions/fragments/core01.profile-jobs.json":
				object["facts"] = []any{}
			case "contracts/extensions/fragments/core01.profile-classifications.json":
				object["facts"] = append(object["facts"].([]any), jobs...)
			}
		})
		if _, err := collectArtifacts(fixture, "extensions"); err != nil {
			t.Fatalf("producer rejects empty fragment: %v", err)
		}
	})
	for _, tc := range []struct {
		name     string
		facts    int
		overflow bool
	}{{"4096 facts", 4096, false}, {"4097 facts", 4097, false}, {"4097 findings", 4097, true}} {
		t.Run(tc.name, func(t *testing.T) {
			want := 0
			fixture := extensionInputFixture(t, root, func(path string, object map[string]any) {
				if path != "contracts/extensions/fragments/core01.profile-classifications.json" {
					return
				}
				rows := object["facts"].([]any)
				want = tc.facts - len(rows)
				if tc.overflow {
					want = tc.facts
				}
				for i := 0; i < want; i++ {
					rows = append(rows, map[string]any{"fact_kind": "egress", "profile_id": fmt.Sprintf("private_fact_%04d", i), "egress_mode": "none"})
				}
				object["facts"] = rows
				body, err := canonicalizeExtension(object)
				if err != nil || len(body) > 1048576 {
					t.Fatal("fixture exceeds canonical fragment byte bound")
				}
			})
			_, err := collectArtifacts(fixture, "extensions")
			var failure *extensionRegistryError
			if !errors.As(err, &failure) {
				t.Fatalf("producer applied unsupported input bound: %v", err)
			}
			if tc.overflow {
				if len(failure.findings) != 1 || failure.findings[0]["reason_code"] != "extension_diagnostic_overflow" {
					t.Fatal("incorrect one-fragment diagnostic overflow")
				}
			} else if len(failure.findings) != want {
				t.Fatalf("producer findings truncated: %d, want %d", len(failure.findings), want)
			}
		})
	}
}

func checkGeneratorInputCapacities(t *testing.T, root string) {
	inputs, err := readExtensionInputCatalog(root)
	if err != nil {
		t.Fatal(err)
	}
	baseProfiles := 0
	for _, object := range inputs {
		if object["schema_id"] == "cartulary.extension_owner_fragment.v3" && object["owner_id"] == "core00" {
			for _, raw := range object["facts"].([]any) {
				if raw.(map[string]any)["fact_kind"] == "recognized_profile" {
					baseProfiles++
				}
			}
		}
	}
	paths := []string{}
	for path, object := range inputs {
		if object["schema_id"] == "cartulary.extension_owner_fragment.v3" {
			paths = append(paths, "contracts/extensions/"+path)
		}
	}
	sort.Strings(paths)
	for _, tc := range []struct {
		name   string
		n      int
		reason string
		limit  int
	}{
		{"global facts", 65536, "extension_registry_invalid", 0}, {"global excess", 65537, "extension_registry_limit_exceeded", 65536},
		{"recognized profiles", 256, "extension_registry_invalid", 0}, {"recognized excess", 257, "extension_registry_limit_exceeded", 256},
		{"fragment bytes", 1048576, "extension_registry_invalid", 0}, {"fragment byte excess", 1048577, "extension_registry_limit_exceeded", 1048576},
	} {
		t.Run(tc.name, func(t *testing.T) {
			fixture := extensionInputFixture(t, root, func(path string, object map[string]any) {
				if object["schema_id"] != "cartulary.extension_owner_fragment.v3" {
					return
				}
				switch {
				case strings.HasPrefix(tc.name, "global"):
					target := tc.n / len(paths)
					if sort.SearchStrings(paths, path) < tc.n%len(paths) {
						target++
					}
					rows := object["facts"].([]any)
					for len(rows) < target {
						rows = append(rows, map[string]any{"fact_kind": "egress", "profile_id": "p", "egress_mode": "none"})
					}
					object["facts"] = rows
				case strings.HasPrefix(tc.name, "recognized") && object["owner_id"] == "core00":
					rows := object["facts"].([]any)
					recognized := 0
					for _, raw := range rows {
						if raw.(map[string]any)["fact_kind"] == "recognized_profile" {
							recognized++
						}
					}
					for i := recognized; i < tc.n; i++ {
						rows = append(rows, map[string]any{"fact_kind": "recognized_profile", "profile_id": fmt.Sprintf("limit_profile_%03d", i), "primary_owner_id": "core01", "claimable": false, "contract_major": nil})
					}
					object["facts"] = rows
				case strings.HasPrefix(tc.name, "fragment byte") && path == "contracts/extensions/fragments/core01.profile-classifications.json":
					fillProducerFragmentBytes(t, object, tc.n)
				}
			})
			_, err := collectArtifacts(fixture, "extensions")
			var failure *extensionRegistryError
			if !errors.As(err, &failure) || len(failure.findings) == 0 {
				t.Fatalf("missing producer boundary result: %v", err)
			}
			for _, finding := range failure.findings {
				if finding["reason_code"] != tc.reason {
					t.Fatalf("incorrect capacity classification: %#v", finding)
				}
			}
			details := failure.findings[0]["details"].(map[string]any)
			if tc.limit > 0 {
				if len(failure.findings) != 1 || fmt.Sprint(details["limit"]) != fmt.Sprint(tc.limit) || fmt.Sprint(details["actual"]) != fmt.Sprint(tc.n) {
					t.Fatal("incorrect numeric capacity finding")
				}
			}
			if tc.limit == 0 {
				kind := "owner_fragment"
				want := 1
				if strings.HasPrefix(tc.name, "recognized") {
					kind = "owner_input_registry"
					want = (tc.n - baseProfiles) * (len(registryfacts.RequiredScalarKinds()) - 1)
				}
				if strings.HasPrefix(tc.name, "fragment byte") {
					want = len(failure.findings)
					if want > 2 {
						t.Fatal("unexpected byte-boundary unknown identities")
					}
				}
				if len(failure.findings) != want {
					t.Fatalf("capacity lost owner findings: got %d, want %d", len(failure.findings), want)
				}
				for _, finding := range failure.findings {
					details := finding["details"].(map[string]any)
					if details["artifact_kind"] != kind || details["safe_ref"] == nil || !strings.HasPrefix(fmt.Sprint(details["expected"]), "exact:") {
						t.Fatalf("capacity prevented owner diagnostic collection: %#v", finding)
					}
				}
			}
		})
	}
	for _, count := range []int{512, 513} {
		t.Run(fmt.Sprintf("%d fragments", count), func(t *testing.T) {
			extra := map[string]map[string]any{}
			for i := len(paths); i < count; i++ {
				id := fmt.Sprintf("core01.limit_%03d", i)
				extra["fragments/"+id+".json"] = map[string]any{"schema_id": "cartulary.extension_owner_fragment.v3", "owner_fragment_id": id, "owner_id": "core01", "facts": []any{}}
			}
			fixture := extensionInputFixture(t, root, func(path string, object map[string]any) {
				if path != "contracts/extensions/index.json" {
					return
				}
				rows := object["artifacts"].([]any)
				for path := range extra {
					rows = append(rows, map[string]any{"path": path, "schema_id": "cartulary.extension_owner_fragment.v3", "owner_id": "core01", "artifact_class": "owner_fragment"})
				}
				sort.Slice(rows, func(i, j int) bool {
					return rows[i].(map[string]any)["path"].(string) < rows[j].(map[string]any)["path"].(string)
				})
				object["artifacts"] = rows
			})
			for path, object := range extra {
				body, err := canonicalizeExtension(object)
				if err != nil {
					t.Fatal(err)
				}
				if err := os.WriteFile(filepath.Join(fixture, "contracts/extensions", path), []byte(body), 0600); err != nil {
					t.Fatal(err)
				}
			}
			_, err := collectArtifacts(fixture, "extensions")
			if count == 512 {
				if err != nil {
					t.Fatalf("maximum empty fragments rejected: %v", err)
				}
			} else {
				var failure *extensionRegistryError
				if !errors.As(err, &failure) || len(failure.findings) != 1 || failure.findings[0]["reason_code"] != "extension_registry_limit_exceeded" {
					t.Fatalf("incorrect fragment limit finding: %v", err)
				}
			}
		})
	}
	for _, id := range []string{"cartulary.extension_owner_fragment.v3", "cartulary.extension_owner_input_registry.v2", "cartulary.extension_profile_registry.v1"} {
		for _, bound := range []string{"minItems", "maxItems"} {
			t.Run(id+bound, func(t *testing.T) {
				fixture := extensionInputFixture(t, root, func(path string, object map[string]any) {
					if path != "contracts/extensions/specification/generated-schema-sources.json" {
						return
					}
					for _, raw := range object["schemas"].([]any) {
						schema := raw.(map[string]any)
						if schema["$id"] != id {
							continue
						}
						field := "facts"
						if strings.Contains(id, "input_registry") {
							field = "owner_fragments"
						}
						if strings.Contains(id, "profile_registry") {
							field = "profiles"
						}
						properties := schema["properties"].(map[string]any)[field].(map[string]any)
						if bound == "minItems" {
							properties[bound] = 1
						} else {
							properties[bound] = 65537
						}
					}
				})
				if _, err := collectArtifacts(fixture, "extensions"); err == nil {
					t.Fatal("inconsistent schema capacity admitted")
				}
			})
		}
	}
}

func fillProducerFragmentBytes(t *testing.T, object map[string]any, target int) {
	t.Helper()
	body, err := canonicalizeExtension(object)
	if err != nil {
		t.Fatal(err)
	}
	fact := map[string]any{"fact_kind": "egress", "profile_id": "p", "egress_mode": "none"}
	unit, err := canonicalizeExtension(fact)
	if err != nil {
		t.Fatal(err)
	}
	remaining := target - len(body)
	count, padding := remaining/len(unit), remaining%len(unit)
	if count < 1 || padding > 63 {
		t.Fatal("invalid byte fixture")
	}
	rows := object["facts"].([]any)
	for i := 0; i < count; i++ {
		row := map[string]any{"fact_kind": "egress", "profile_id": "p", "egress_mode": "none"}
		if i == count-1 {
			row["profile_id"] = strings.Repeat("p", padding+1)
		}
		rows = append(rows, row)
	}
	object["facts"] = rows
	body, err = canonicalizeExtension(object)
	if err != nil || len(body) != target || body[len(body)-1] != '\n' {
		t.Fatal("incorrect LF-inclusive byte fixture")
	}
}
