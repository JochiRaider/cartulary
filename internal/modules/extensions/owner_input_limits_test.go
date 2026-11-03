package extensions

import (
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/extensions/registryfacts"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

func checkOwnerInputLimitAdmission(t *testing.T) {
	t.Run("actual capacities", checkOwnerInputCapacities)
	t.Run("empty fragment preserves declarations", func(t *testing.T) {
		source := newMutableGeneratedSource()
		jobs := decodeMutableObject(t, source["contracts/extensions/fragments/core01.profile-jobs.json"].JSON)["facts"].([]any)
		rewriteCoherentPackage(t, source, func(object map[string]any) {
			if object["schema_id"] != "cartulary.extension_owner_fragment.v3" || object["owner_fragment_id"] == nil {
				return
			}
			switch object["owner_fragment_id"] {
			case "core01.extension_profile_jobs_v1":
				object["facts"] = []any{}
			case "core01.extension_profile_classifications_v1":
				object["facts"] = append(object["facts"].([]any), jobs...)
			}
			normalizeLimitFixture(t, object)
		})
		if _, err := NewCoordinator(source); err != nil {
			t.Fatalf("empty fragment rejected after preserving all declarations: %v", err)
		}
	})
	for _, tc := range []struct {
		name     string
		facts    int
		overflow bool
	}{{"4096 facts", 4096, false}, {"4097 facts", 4097, false}, {"4097 findings", 4097, true}} {
		t.Run(tc.name, func(t *testing.T) {
			source := newMutableGeneratedSource()
			want := 0
			rewriteCoherentPackage(t, source, func(object map[string]any) {
				if object["schema_id"] != "cartulary.extension_owner_fragment.v3" || object["owner_fragment_id"] != "core01.extension_profile_classifications_v1" {
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
				normalizeLimitFixture(t, object)
				body, err := canonicaljson.MarshalExtension(object)
				if err != nil || len(body) > 1048576 {
					t.Fatal("fixture exceeds actual canonical fragment byte bound")
				}
			})
			_, err := NewCoordinator(source)
			var failure *StartupValidationError
			if !errors.As(err, &failure) {
				t.Fatalf("missing diagnostics: %v", err)
			}
			if tc.overflow {
				if len(failure.Findings) != 1 || failure.Findings[0].ReasonCode != "extension_diagnostic_overflow" {
					t.Fatalf("fact count confused with diagnostic limit: %#v", failure.Findings)
				}
			} else {
				if len(failure.Findings) != want {
					t.Fatalf("unsupported fact limit: got %d findings, want %d", len(failure.Findings), want)
				}
				for _, finding := range failure.Findings {
					if finding.ReasonCode != "extension_registry_invalid" || finding.Details["artifact_kind"] != "owner_fragment" {
						t.Fatalf("incorrect owner finding: %#v", finding)
					}
				}
			}
			body, _ := json.Marshal(failure.Findings)
			if strings.Contains(string(body), "private_fact_") {
				t.Fatal("unknown value leaked")
			}
		})
	}
}

func normalizeLimitFixture(t *testing.T, object map[string]any) {
	t.Helper()
	fragments, err := registryfacts.NormalizeFragments([]map[string]any{object})
	if err != nil {
		t.Fatal(err)
	}
	object["facts"] = fragments[0]["facts"]
}

func checkOwnerInputCapacities(t *testing.T) {
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
			source := newMutableGeneratedSource()
			baseProfiles := len(decodeMutableObject(t, source[generatedExtensionsRoot+"profile-registry.json"].JSON)["profiles"].([]any))
			ids := []string{}
			for path, artifact := range source {
				if strings.HasPrefix(path, "contracts/extensions/fragments/") {
					ids = append(ids, decodeMutableObject(t, artifact.JSON)["owner_fragment_id"].(string))
				}
			}
			sort.Strings(ids)
			rewriteCoherentPackage(t, source, func(object map[string]any) {
				if object["schema_id"] != "cartulary.extension_owner_fragment.v3" || object["owner_fragment_id"] == nil {
					return
				}
				id := object["owner_fragment_id"].(string)
				switch {
				case strings.HasPrefix(tc.name, "global"):
					target := tc.n / len(ids)
					if sort.SearchStrings(ids, id) < tc.n%len(ids) {
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
				case strings.HasPrefix(tc.name, "fragment byte") && id == "core01.extension_profile_classifications_v1":
					fillLimitFragmentBytes(t, object, tc.n)
				}
				normalizeLimitFixture(t, object)
			})
			_, err := NewCoordinator(source)
			var failure *StartupValidationError
			if !errors.As(err, &failure) || len(failure.Findings) == 0 {
				t.Fatalf("missing boundary result: %v", err)
			}
			for _, finding := range failure.Findings {
				if finding.ReasonCode != tc.reason {
					t.Fatalf("incorrect capacity classification: %#v", finding)
				}
			}
			if tc.limit > 0 {
				if len(failure.Findings) != 1 || fmt.Sprint(failure.Findings[0].Details["limit"]) != fmt.Sprint(tc.limit) || fmt.Sprint(failure.Findings[0].Details["actual"]) != fmt.Sprint(tc.n) {
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
					want = len(failure.Findings)
					if want > 2 {
						t.Fatal("unexpected byte-boundary unknown identities")
					}
				}
				if len(failure.Findings) != want {
					t.Fatalf("capacity lost owner findings: got %d, want %d", len(failure.Findings), want)
				}
				for _, finding := range failure.Findings {
					if finding.Details["artifact_kind"] != kind || finding.Details["safe_ref"] == nil || !strings.HasPrefix(fmt.Sprint(finding.Details["expected"]), "exact:") {
						t.Fatalf("capacity prevented owner diagnostic collection: %#v", finding)
					}
				}
			}
		})
	}
	for _, count := range []int{512, 513} {
		t.Run(fmt.Sprintf("%d fragments", count), func(t *testing.T) {
			source := newMutableGeneratedSource()
			extra := []any{}
			digests := []any{}
			existing := 0
			for path := range source {
				if strings.HasPrefix(path, "contracts/extensions/fragments/") {
					existing++
				}
			}
			for i := existing; i < count; i++ {
				id := fmt.Sprintf("core01.limit_%03d", i)
				object := map[string]any{"schema_id": "cartulary.extension_owner_fragment.v3", "owner_fragment_id": id, "owner_id": "core01", "facts": []any{}}
				source.replace(t, "contracts/extensions/fragments/"+id+".json", object)
				extra = append(extra, object)
				digests = append(digests, map[string]any{"owner_fragment_id": id, "owner_fragment_sha256": canonicalDigest(object)})
			}
			rewriteCoherentPackage(t, source, func(object map[string]any) {
				switch object["schema_id"] {
				case "cartulary.extension_owner_input_registry.v2":
					if object["owner_fragments"] == nil {
						return
					}
					rows := append(object["owner_fragments"].([]any), extra...)
					sort.Slice(rows, func(i, j int) bool {
						return rows[i].(map[string]any)["owner_fragment_id"].(string) < rows[j].(map[string]any)["owner_fragment_id"].(string)
					})
					object["owner_fragments"] = rows
				case "cartulary.extension_registry_integrity.v2":
					if object["owner_fragment_digests"] == nil {
						return
					}
					rows := append(object["owner_fragment_digests"].([]any), digests...)
					sort.Slice(rows, func(i, j int) bool {
						return rows[i].(map[string]any)["owner_fragment_id"].(string) < rows[j].(map[string]any)["owner_fragment_id"].(string)
					})
					object["owner_fragment_digests"] = rows
				}
			})
			_, err := NewCoordinator(source)
			if count == 512 {
				if err != nil {
					t.Fatalf("maximum empty fragments rejected: %v", err)
				}
			} else {
				requireFinding(t, err, "extension_registry_limit_exceeded", "registry_generation")
			}
		})
	}
	for _, id := range []string{"cartulary.extension_owner_fragment.v3", "cartulary.extension_owner_input_registry.v2", "cartulary.extension_profile_registry.v1"} {
		for _, bound := range []string{"minItems", "maxItems"} {
			t.Run(id+bound, func(t *testing.T) {
				source := newMutableGeneratedSource()
				rewriteCoherentPackage(t, source, func(object map[string]any) {
					if object["$id"] != id {
						return
					}
					field := "facts"
					if strings.Contains(id, "input_registry") {
						field = "owner_fragments"
					}
					if strings.Contains(id, "profile_registry") {
						field = "profiles"
					}
					properties := object["properties"].(map[string]any)[field].(map[string]any)
					if bound == "minItems" {
						properties[bound] = 1
					} else {
						properties[bound] = 65537
					}
				})
				_, err := NewCoordinator(source)
				requireFinding(t, err, "extension_registry_invalid", "registry_generation")
			})
		}
	}
}

func fillLimitFragmentBytes(t *testing.T, object map[string]any, target int) {
	t.Helper()
	body, err := canonicaljson.MarshalExtension(object)
	if err != nil {
		t.Fatal(err)
	}
	fact := map[string]any{"fact_kind": "egress", "profile_id": "p", "egress_mode": "none"}
	unit, err := canonicaljson.MarshalExtension(fact)
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
	body, err = canonicaljson.MarshalExtension(object)
	if err != nil || len(body) != target || body[len(body)-1] != '\n' {
		t.Fatal("incorrect LF-inclusive byte fixture")
	}
}
