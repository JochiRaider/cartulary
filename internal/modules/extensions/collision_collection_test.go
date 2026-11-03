package extensions

import (
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/extensions/registryfacts"
)

func checkRegistryCollisionCollection(t *testing.T) {
	source := newMutableGeneratedSource()
	rewriteCoherentPackage(t, source, func(object map[string]any) {
		if reservations, ok := object["reservations"].([]any); ok && object["schema_id"] == "cartulary.base_route_reservation_registry.v3" {
			object["reservations"] = append(reservations,
				map[string]any{"reservation_id": "base.test.import", "path_template": "/api/v1/import-sessions/{session_id}", "match_scope": "exact"},
				map[string]any{"reservation_id": "base.test.reference", "path_template": "/api/v1/reference-packs/{pack_id}", "match_scope": "exact"})
		}
		rows, ok := object["facts"].([]any)
		if !ok || object["schema_id"] != "cartulary.extension_owner_fragment.v3" {
			return
		}
		amended := []any{}
		for _, raw := range rows {
			fact := raw.(map[string]any)
			if fact["profile_id"] == "import" && fact["fact_kind"] == "egress" {
				unknown := cloneObject(fact)
				unknown["profile_id"] = "private_profile_collision"
				amended = append(amended, unknown)
				continue
			}
			amended = append(amended, fact)
			if fact["profile_id"] == "import" && fact["fact_kind"] == "recognized_profile" || fact["profile_id"] == "reference_pack" && fact["fact_kind"] == "portability" {
				amended = append(amended, cloneObject(fact), cloneObject(fact))
			}
		}
		object["facts"] = amended
		normalized, err := registryfacts.NormalizeFragments([]map[string]any{object})
		if err != nil {
			t.Fatal(err)
		}
		object["facts"] = normalized[0]["facts"]
	})
	_, err := NewCoordinator(source)
	var failure *StartupValidationError
	if !errors.As(err, &failure) || len(failure.Findings) != 6 {
		t.Fatalf("same-stage findings incomplete: %#v", failure)
	}
	classes := map[string]int{}
	for _, finding := range failure.Findings {
		if finding.ReasonCode == "extension_registry_invalid" {
			classes["owner_input"]++
			continue
		}
		if finding.ReasonCode != "extension_registry_conflict" {
			t.Fatalf("incorrect collision classification: %#v", finding)
		}
		class := stringValue(finding.Details["collision_class"])
		classes[class]++
		tokens, _ := stringSlice(finding.Details["conflicting_tokens"])
		profiles, _ := stringSlice(finding.Details["conflicting_profile_ids"])
		if len(profiles) != 1 {
			t.Fatalf("missing or spurious participating profiles: %#v", finding)
		}
		switch class {
		case "profile_id":
			if profiles[0] != "import" || len(tokens) != 1 || tokens[0] != "import" || fmt.Sprint(finding.Details["duplicate_count"]) != "3" {
				t.Fatalf("incorrect grouped profile: %#v", finding)
			}
		case "owner_fact_identity":
			if profiles[0] != "reference_pack" || len(tokens) != 1 || !strings.HasPrefix(tokens[0], "portability:sha256:") || fmt.Sprint(finding.Details["duplicate_count"]) != "3" {
				t.Fatalf("incorrect fallback identity: %#v", finding)
			}
		case "base_route_capture":
			if len(tokens) != 2 || tokens[0] > tokens[1] || finding.Details["duplicate_count"] != nil || (profiles[0] != "import" && profiles[0] != "reference_pack") {
				t.Fatalf("incorrect Base capture: %#v", finding)
			}
		default:
			t.Fatalf("unrequested or duplicate fallback collision: %s", class)
		}
	}
	if classes["owner_input"] != 2 || classes["profile_id"] != 1 || classes["owner_fact_identity"] != 1 || classes["base_route_capture"] != 2 {
		t.Fatalf("incorrect grouping: %#v", classes)
	}
	body, _ := json.Marshal(failure.Findings)
	if strings.Contains(string(body), "private_profile_collision") {
		t.Fatal("unknown profile leaked during collision collection")
	}
}

func checkUnadmittedCollisionValues(t *testing.T) {
	for _, mode := range []string{"owner allocation", "namespace identity"} {
		t.Run(mode, func(t *testing.T) {
			source := newMutableGeneratedSource()
			rewriteCoherentPackage(t, source, func(object map[string]any) {
				if mode == "namespace identity" && object["schema_id"] == "cartulary.extension_profile_configuration_contract.v4" && object["profile_id"] != nil {
					object["configuration_namespace"] = "private_marker"
				}
				rows, ok := object["facts"].([]any)
				if mode != "owner allocation" || !ok || object["schema_id"] != "cartulary.extension_owner_fragment.v3" || object["owner_id"] != "core04" {
					return
				}
				fact := map[string]any{"fact_kind": "contribution", "profile_id": "import", "contribution": map[string]any{"kind": "extension_resource_kind", "contribution_id": "import.private_marker", "resource_kind": "private_marker"}}
				object["facts"] = append(rows, fact, cloneObject(fact))
				normalized, err := registryfacts.NormalizeFragments([]map[string]any{object})
				if err != nil {
					t.Fatal(err)
				}
				object["facts"] = normalized[0]["facts"]
			})
			_, err := NewCoordinator(source)
			var failure *StartupValidationError
			if !errors.As(err, &failure) {
				t.Fatalf("invalid declaration admitted: %v", err)
			}
			body, _ := json.Marshal(failure.Findings)
			if strings.Contains(string(body), "private_marker") {
				t.Fatal("unadmitted declaration value escaped through collision diagnostics")
			}
			for _, finding := range failure.Findings {
				if finding.ReasonCode != "extension_registry_invalid" {
					t.Fatal("invalid declarations became authoritative collision identities")
				}
			}
		})
	}
}

func checkDeclaredNamespaceCollision(t *testing.T) {
	source := newMutableGeneratedSource()
	rewriteCoherentPackage(t, source, func(object map[string]any) {
		if object["schema_id"] == "cartulary.extension_profile_configuration_contract.v4" && object["profile_id"] == "reference_pack" {
			object["configuration_namespace"] = "import"
			for _, raw := range object["keys"].([]any) {
				row := raw.(map[string]any)
				row["key"] = "import." + strings.TrimPrefix(row["key"].(string), "reference_packs.")
			}
		}
	})
	_, err := NewCoordinator(source)
	var failure *StartupValidationError
	if !errors.As(err, &failure) || len(failure.Findings) != 1 {
		t.Fatalf("namespace collision not collected: %v", err)
	}
	finding := failure.Findings[0]
	details := finding.Details
	profiles, _ := stringSlice(details["conflicting_profile_ids"])
	tokens, _ := stringSlice(details["conflicting_tokens"])
	if finding.ReasonCode != "extension_registry_conflict" || details["collision_class"] != "configuration_namespace" || fmt.Sprint(details["duplicate_count"]) != "2" || strings.Join(profiles, ",") != "import,reference_pack" || strings.Join(tokens, ",") != "import" {
		t.Fatalf("namespace/profile identities were conflated: %#v", finding)
	}
}
