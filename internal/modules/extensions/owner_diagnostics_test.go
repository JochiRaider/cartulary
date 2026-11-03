package extensions

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/extensions/registryfacts"
)

func checkOwnerInputDiagnostics(t *testing.T) {
	for _, conditionID := range []string{"extension_registry_collision", "extension_registry_limit"} {
		for _, mode := range []string{"missing condition", "wrong formatter", "wrong class"} {
			t.Run(conditionID+"/"+mode, func(t *testing.T) {
				source := newMutableGeneratedSource()
				rewriteCoherentPackage(t, source, func(object map[string]any) {
					rows, ok := object["conditions"].([]any)
					if !ok || object["schema_id"] != "cartulary.extension_validation_condition_registry.v2" {
						return
					}
					kept := []any{}
					for _, raw := range rows {
						condition := raw.(map[string]any)
						if condition["condition_id"] == conditionID {
							if mode == "missing condition" {
								continue
							}
							if mode == "wrong formatter" {
								condition["expected_formatter_id"] = "diagnostic_redacted_v1"
							}
							if mode == "wrong class" {
								condition["condition_class"] = "invalid_scalar"
							}
						}
						kept = append(kept, condition)
					}
					object["conditions"] = kept
				})
				_, err := NewCoordinator(source)
				if err == nil {
					t.Fatal("invalid registry diagnostic definition admitted on a valid catalog")
				}
				var failure *StartupValidationError
				if !errors.As(err, &failure) || len(failure.Findings) != 1 || failure.Findings[0].ReasonCode != "extension_registry_invalid" {
					t.Fatalf("incorrect definition admission failure: %v", err)
				}
			})
		}
	}

	t.Run("collect every missing profile fact", func(t *testing.T) {
		source := newMutableGeneratedSource()
		rewriteCoherentPackage(t, source, func(object map[string]any) {
			rows, ok := object["facts"].([]any)
			if !ok || object["schema_id"] != "cartulary.extension_owner_fragment.v3" {
				return
			}
			kept := make([]any, 0, len(rows))
			for _, raw := range rows {
				fact := raw.(map[string]any)
				profile := stringValue(fact["profile_id"])
				if fact["fact_kind"] == "egress" && (profile == "import" || profile == "reference_pack") {
					continue
				}
				kept = append(kept, fact)
			}
			object["facts"] = kept
		})
		_, err := NewCoordinator(source)
		var failure *StartupValidationError
		if !errors.As(err, &failure) || len(failure.Findings) != 2 {
			t.Fatalf("missing facts were not collected: %#v", failure)
		}
		seen := map[string]bool{}
		for _, finding := range failure.Findings {
			if finding.Path != "$.owner_fragments" || finding.ReasonCode != "extension_registry_invalid" || finding.Message != "The extension contract artifact set is invalid." || finding.Details["artifact_kind"] != "owner_input_registry" || finding.Details["expected"] != "exact:egress" || finding.Details["actual"] != "missing" {
				t.Fatalf("incorrect missing-fact diagnostic: %#v", finding)
			}
			seen[stringValue(finding.Details["safe_ref"])] = true
		}
		if !seen["cartulary:profile:import"] || !seen["cartulary:profile:reference_pack"] {
			t.Fatal("missing profile identity was replaced by an owner identity")
		}
	})
	t.Run("unknown identities use normalized paths and safe refs", func(t *testing.T) {
		source := newMutableGeneratedSource()
		rewriteCoherentPackage(t, source, func(object map[string]any) {
			rows, ok := object["facts"].([]any)
			if !ok || object["schema_id"] != "cartulary.extension_owner_fragment.v3" || object["owner_fragment_id"] != "core01.extension_profile_classifications_v1" {
				return
			}
			for _, raw := range rows {
				fact := raw.(map[string]any)
				if fact["fact_kind"] != "egress" {
					continue
				}
				for _, profile := range []string{"private_profile_alpha", "private_profile_beta", "private_profile_alpha"} {
					copy := cloneObject(fact)
					copy["profile_id"] = profile
					rows = append(rows, copy)
				}
				break
			}
			object["facts"] = rows
			normalized, err := registryfacts.NormalizeFragments([]map[string]any{object})
			if err != nil {
				t.Fatal(err)
			}
			object["facts"] = normalized[0]["facts"]
		})
		input := decodeMutableObject(t, source[generatedExtensionsRoot+"owner-input-registry.json"].JSON)
		expected := map[string]bool{}
		identities := map[string]bool{}
		for fi, fragment := range input["owner_fragments"].([]any) {
			for i, raw := range fragment.(map[string]any)["facts"].([]any) {
				profile := stringValue(raw.(map[string]any)["profile_id"])
				if strings.HasPrefix(profile, "private_profile_") && !identities[profile] {
					expected[fmt.Sprintf("$.owner_fragments[%d].facts[%d].profile_id", fi, i)] = true
					identities[profile] = true
				}
			}
		}
		if len(expected) != 2 {
			t.Fatal("fixture did not create two unrecognized identities")
		}
		_, err := NewCoordinator(source)
		var failure *StartupValidationError
		if !errors.As(err, &failure) || len(failure.Findings) != 2 {
			t.Fatalf("unknown facts were not collected by identity: %#v", failure)
		}
		for _, finding := range failure.Findings {
			if !expected[finding.Path] || finding.ReasonCode != "extension_registry_invalid" || finding.Details["safe_ref"] != "cartulary:owner_fragment:core01.extension_profile_classifications_v1" || finding.Details["artifact_kind"] != "owner_fragment" || finding.Details["expected"] != "exact:recognized_profile_id" || finding.Details["actual"] != "redacted" {
				t.Fatalf("unsafe or misindexed unknown fact: %#v", finding)
			}
			delete(expected, finding.Path)
		}
		encoded, _ := json.Marshal(failure.Findings)
		if len(expected) != 0 || strings.Contains(string(encoded), "private_profile_") || strings.Contains(string(encoded), "sha256") {
			t.Fatal("unknown fact value or identity hash escaped")
		}
	})
	for _, tc := range []struct {
		count     int
		collision bool
	}{{4096, false}, {4097, false}, {4096, true}} {
		count := tc.count
		t.Run(fmt.Sprintf("bounded findings %d collision %t", count, tc.collision), func(t *testing.T) {
			source := newMutableGeneratedSource()
			var template map[string]any
			for _, artifact := range source {
				object := decodeMutableObject(t, artifact.JSON)
				if object["owner_fragment_id"] != "core01.extension_profile_classifications_v1" {
					continue
				}
				for _, raw := range object["facts"].([]any) {
					fact := raw.(map[string]any)
					if fact["fact_kind"] == "egress" {
						template = fact
						break
					}
				}
			}
			if template == nil {
				t.Fatal("missing egress fixture")
			}
			rewriteCoherentPackage(t, source, func(object map[string]any) {
				rows, ok := object["facts"].([]any)
				if !ok || object["schema_id"] != "cartulary.extension_owner_fragment.v3" {
					return
				}
				start, end := 0, 0
				switch object["owner_fragment_id"] {
				case "core00.extension_recognition_v1":
					start, end = 0, count/2
					if tc.collision {
						for _, raw := range rows {
							if raw.(map[string]any)["fact_kind"] == "recognized_profile" {
								rows = append(rows, cloneObject(raw.(map[string]any)))
								break
							}
						}
					}
				case "core01.extension_profile_classifications_v1":
					start, end = count/2, count
				default:
					return
				}
				for i := start; i < end; i++ {
					copy := cloneObject(template)
					copy["profile_id"] = fmt.Sprintf("unknown_%04d", i)
					rows = append(rows, copy)
				}
				object["facts"] = rows
				normalized, err := registryfacts.NormalizeFragments([]map[string]any{object})
				if err != nil {
					t.Fatal(err)
				}
				object["facts"] = normalized[0]["facts"]
			})
			_, err := NewCoordinator(source)
			var failure *StartupValidationError
			if !errors.As(err, &failure) {
				t.Fatalf("missing bounded findings: %v", err)
			}
			if count == 4096 && !tc.collision {
				if len(failure.Findings) != 4096 {
					t.Fatalf("findings truncated at %d", len(failure.Findings))
				}
				prior := ""
				for _, finding := range failure.Findings {
					if finding.ReasonCode != "extension_registry_invalid" || finding.Path < prior {
						t.Fatal("incorrect finding ordering")
					}
					prior = finding.Path
				}
			} else if len(failure.Findings) != 1 || failure.Findings[0].ReasonCode != "extension_diagnostic_overflow" || failure.Findings[0].Details["phase"] != "registry_generation" || fmt.Sprint(failure.Findings[0].Details["limit"]) != "4096" || fmt.Sprint(failure.Findings[0].Details["actual"]) != "4097" {
				t.Fatalf("incorrect overflow: %#v", failure.Findings)
			}
		})
	}

}
