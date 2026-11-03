package extensions

import (
	"strings"
	"testing"
)

func checkEmptyRegistry(t *testing.T) {
	source := emptyRegistrySource(t)
	coordinator, err := NewCoordinator(source)
	if err != nil {
		t.Fatalf("admit fully bound empty registry: %#v", err)
	}
	if len(coordinator.Descriptors()) != 0 {
		t.Fatal("empty registry invented descriptors")
	}
	resolution, err := coordinator.ResolveClaims(nil)
	if err != nil {
		t.Fatal(err)
	}
	if len(resolution.Claims().ProfileIDs()) != 0 || len(resolution.AdmissionOrder()) != 0 {
		t.Fatal("empty resolution invented a claim")
	}
	plan, err := coordinator.BuildPublicationPlan(resolution)
	if err != nil {
		t.Fatal(err)
	}
	if len(plan.Discovery()) != 0 || len(plan.Routes()) != 0 || len(plan.Workspaces()) != 0 || len(plan.Workers()) != 0 || len(plan.JobKindContracts()) != 0 || len(plan.Contributions()) != 0 {
		t.Fatal("empty publication invented components")
	}
	if _, err := coordinator.ResolveClaims([]string{"network_flow_activity"}); err == nil {
		t.Fatal("empty registry admitted a claim")
	}
	if len(plan.Listeners()) != 3 {
		t.Fatal("empty catalog lost Core activation gates")
	}
	for _, schema := range []string{"cartulary.extension_state_registry.v1", "cartulary.extension_backup_registry.v1", "cartulary.client_extension_support_registry.v1"} {
		t.Run("nonempty companion "+schema, func(t *testing.T) {
			inconsistent := emptyRegistrySource(t)
			rewriteCoherentPackage(t, inconsistent, func(object map[string]any) {
				if object["schema_id"] == schema {
					if _, data := object["profiles"]; data {
						object["profiles"] = []any{map[string]any{"profile_id": "unrecognized"}}
					}
				}
			})
			if _, err := NewCoordinator(inconsistent); err == nil {
				t.Fatal("empty registry admitted a nonempty companion")
			}
		})
	}

}

func emptyRegistrySource(t *testing.T) mutableArtifactSource {
	t.Helper()
	source := newMutableGeneratedSource()
	for path := range source {
		relative := strings.TrimPrefix(path, "contracts/extensions/")
		remove := strings.HasPrefix(relative, "profiles/") || strings.HasPrefix(relative, "fragments/")
		if strings.HasPrefix(relative, "build/") && relative != "build/base-route-reservations.json" && relative != "build/client-support.json" && relative != "build/implementation-bindings.json" {
			remove = true
		}
		for _, dir := range []string{"descriptors/", "implementation-bindings/", "job-contracts/", "worker-runtime-contracts/"} {
			if strings.HasPrefix(path, generatedExtensionsRoot+dir) {
				remove = true
			}
		}
		if remove {
			delete(source, path)
		}
	}
	runtimePaths := map[string]bool{}
	for _, path := range source.ArtifactPaths() {
		runtimePaths[path] = true
	}
	for path := range source {
		if !runtimePaths[path] {
			delete(source, path)
		}
	}
	retained := map[string]bool{}
	for _, path := range source.ArtifactPaths() {
		id := strings.TrimSuffix(strings.ReplaceAll(strings.TrimPrefix(path, "contracts/extensions/"), "/", "."), ".json")
		if strings.HasPrefix(path, generatedExtensionsRoot) {
			id = strings.TrimSuffix(strings.TrimPrefix(path, generatedExtensionsRoot), ".json")
		}
		retained[id] = true
	}
	rewriteCoherentPackage(t, source, func(object map[string]any) {
		if _, definition := object["artifact_class"]; definition {
			return
		}
		if _, schema := object["$schema"]; schema {
			return
		}
		switch object["schema_id"] {
		case "cartulary.extension_dependency_snapshot.v3":
			rows, ok := object["dependencies"].([]any)
			if !ok {
				return
			}
			for _, raw := range rows {
				row := raw.(map[string]any)
				for _, key := range []string{"imported_schema_ids", "imported_algorithm_ids", "imported_artifacts"} {
					row[key] = []any{}
				}
			}

		case "cartulary.extension_profile_registry.v1", "cartulary.extension_state_registry.v1", "cartulary.extension_backup_registry.v1", "cartulary.client_extension_support_registry.v1":
			if _, exists := object["profiles"]; exists {
				object["profiles"] = []any{}
			}
		case "cartulary.extension_client_support_registry_source.v1":
			if _, exists := object["rows"]; exists {
				object["rows"] = []any{}
			}
		case "cartulary.extension_owner_input_registry.v2":
			if _, exists := object["owner_fragments"]; exists {
				object["owner_fragments"] = []any{}
			}
		case "cartulary.extension_participant_registry.v1":
			if _, exists := object["participants"]; exists {
				object["participants"] = []any{}
			}
		case "cartulary.extension_registry_integrity.v2":
			if _, root := object["generator_id"]; !root {
				return
			}
			for _, field := range []string{"descriptor_digests", "owner_fragment_digests", "implementation_binding_digests"} {
				object[field] = []any{}
			}
			rows := []any{}
			for _, raw := range object["supporting_contract_artifact_digests"].([]any) {
				row := raw.(map[string]any)
				if retained[stringValue(row["artifact_id"])] {
					rows = append(rows, row)
				}
			}
			object["supporting_contract_artifact_digests"] = rows
		}
	})
	return source
}
