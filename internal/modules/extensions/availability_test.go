package extensions

import (
	"errors"
	"reflect"
	"testing"
)

func checkUnavailableProfiles(t *testing.T) {
	baseline := requireGeneratedCoordinator(t)
	t.Run("unrecognized implementation owner", func(t *testing.T) {
		source := newMutableGeneratedSource()
		binding := decodeMutableObject(t, source[generatedExtensionsRoot+"implementation-bindings/enterprise_authentication.json"].JSON)
		binding["profile_id"] = "unrecognized"
		path := generatedExtensionsRoot + "implementation-bindings/unrecognized.json"
		source.replace(t, path, binding)
		rewriteCoherentPackage(t, source, func(object map[string]any) {
			if object["schema_id"] == "cartulary.extension_registry_integrity.v2" {
				if rows, ok := object["implementation_binding_digests"].([]any); ok {
					object["implementation_binding_digests"] = append(rows, map[string]any{"profile_id": "unrecognized", "binding_sha256": source[path].SHA256})
				}
			}
		})
		_, err := NewCoordinator(source)
		requireFinding(t, err, "extension_registry_invalid", "registry_generation")
	})
	t.Run("unclaimable reserved profile", func(t *testing.T) {
		source := registryWithoutBinding(t, "enterprise_authentication")
		rewriteCoherentPackage(t, source, func(object map[string]any) {
			if object["profile_id"] == "enterprise_authentication" {
				if _, recognition := object["claimable"]; recognition {
					object["claimable"], object["contract_major"] = false, nil
				}
			}
		})
		coordinator, err := NewCoordinator(source)
		if err != nil {
			t.Fatalf("admit unclaimable reservation: %#v", err)
		}
		retainedPolicy := false
		for _, policy := range coordinator.InactiveConfigurationPolicies() {
			if policy.ProfileID == "enterprise_authentication" {
				retainedPolicy = true
				if policy.ClaimKey != "enterprise_authentication.claimed" || policy.Namespace != "enterprise_authentication" {
					t.Fatal("unclaimable retained policy lost canonical identity")
				}
			}
		}
		if !retainedPolicy {
			t.Fatal("unclaimable configuration contract was discarded")
		}
		retainedContract := false
		for _, contract := range coordinator.ConfigurationSchemaContracts() {
			if contract.ProfileID == "enterprise_authentication" {
				retainedContract = contract.Namespace == "enterprise_authentication" && len(contract.Values) == 1
			}
		}
		if !retainedContract {
			t.Fatal("unclaimable configuration namespace/reference inventory was discarded")
		}
		resolution, err := coordinator.ResolveClaims(nil)
		if err != nil {
			t.Fatal(err)
		}
		plan, err := coordinator.BuildPublicationPlan(resolution)
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, profile := range plan.Discovery() {
			if profile.ProfileID == "enterprise_authentication" {
				found = true
				if profile.Claimable || profile.Claimed || profile.ContractMajor != nil || len(profile.RouteFamilies) != 4 {
					t.Fatal("unclaimable discovery lost its null-major reservation variant")
				}
			}
		}
		if !found {
			t.Fatal("unclaimable profile disappeared from discovery")
		}
		_, err = coordinator.ResolveClaims([]string{"enterprise_authentication"})
		requireFinding(t, err, "extension_profile_not_claimable", "claim_configuration")
	})
	t.Run("unclaimable profile with reserved binding", func(t *testing.T) {
		source := newMutableGeneratedSource()
		rewriteCoherentPackage(t, source, func(object map[string]any) {
			if object["profile_id"] != "enterprise_authentication" {
				return
			}
			if _, recognized := object["claimable"]; recognized {
				object["claimable"], object["contract_major"] = false, nil
			} else if object["schema_id"] == "cartulary.extension_implementation_binding.v2" {
				object["contract_major"] = nil
			}
		})
		coordinator, err := NewCoordinator(source)
		if err != nil {
			t.Fatalf("reserved binding changed claimability: %#v", err)
		}
		_, err = coordinator.ResolveClaims([]string{"enterprise_authentication"})
		requireFinding(t, err, "extension_profile_not_claimable", "claim_configuration")
	})
	for _, profile := range []string{"enterprise_authentication", "import", "network_flow_activity"} {
		t.Run(profile, func(t *testing.T) {
			source := registryWithoutBinding(t, profile)
			coordinator, err := NewCoordinator(source)
			if err != nil {
				t.Fatalf("recognized profile without an implementation prevented package admission: %#v", err)
			}
			expectedBackups := baseline.BackupPlans()
			for i := range expectedBackups {
				if expectedBackups[i].ProfileID == profile {
					expectedBackups[i].ImplementationBindingSHA256 = ""
				}
			}
			if !reflect.DeepEqual(baseline.Descriptors(), coordinator.Descriptors()) ||
				!reflect.DeepEqual(baseline.JobKindContracts(), coordinator.JobKindContracts()) ||
				!reflect.DeepEqual(baseline.WorkerRuntimeContracts(), coordinator.WorkerRuntimeContracts()) ||
				!reflect.DeepEqual(expectedBackups, coordinator.BackupPlans()) {
				t.Fatal("implementation availability changed the declared identity, job or storage catalog")
			}
			resolution, err := coordinator.ResolveClaims(nil)
			if err != nil {
				t.Fatal(err)
			}
			plan, err := coordinator.BuildPublicationPlan(resolution)
			if err != nil {
				t.Fatal(err)
			}
			if len(plan.Discovery()) != len(baseline.Descriptors()) || len(plan.Workers()) != 0 || len(plan.Workspaces()) != 0 || len(plan.Contributions()) != 0 {
				t.Fatal("unclaimed package lost discovery or published executable contributions")
			}
			for _, route := range plan.Routes() {
				if route.DispatchState != "inactive" {
					t.Fatal("an unclaimed route became active")
				}
			}
			if _, err := coordinator.ResolveClaims([]string{profile}); err == nil {
				t.Fatal("profile without implementation was claimed")
			} else {
				var validation *ValidationError
				if !errors.As(err, &validation) || len(validation.Findings()) != 1 || validation.Findings()[0].Code != "extension_implementation_unavailable" {
					t.Fatalf("unavailable claim classification: %v", err)
				}
			}
			if profile == "network_flow_activity" {
				if _, ok := coordinator.StatePlan(profile); ok {
					t.Fatal("unimplemented profile exposed an executable state plan")
				}
			}
		})
	}
}

func registryWithoutBinding(t *testing.T, profile string) mutableArtifactSource {
	t.Helper()
	source := newMutableGeneratedSource()
	delete(source, generatedExtensionsRoot+"implementation-bindings/"+profile+".json")
	rewriteCoherentPackage(t, source, func(object map[string]any) {
		if object["schema_id"] == "cartulary.extension_registry_integrity.v2" {
			if rows, ok := object["implementation_binding_digests"].([]any); ok {
				kept := []any{}
				for _, raw := range rows {
					if raw.(map[string]any)["profile_id"] != profile {
						kept = append(kept, raw)
					}
				}
				object["implementation_binding_digests"] = kept
			}
		}
		if object["profile_id"] == profile {
			if _, stateRow := object["implementation_binding_sha256"]; stateRow {
				object["implementation_binding_sha256"] = nil
			}
		}
	})
	return source
}
