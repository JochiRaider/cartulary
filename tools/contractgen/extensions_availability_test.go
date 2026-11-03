package main

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func extensionInputFixture(t *testing.T, root string, mutate func(string, map[string]any)) string {
	t.Helper()
	inputs, err := readExtensionInputCatalog(root)
	if err != nil {
		t.Fatal(err)
	}
	refs := map[string]bool{"contracts/extensions/index.json": true}
	for path := range inputs {
		refs["contracts/extensions/"+path] = true
	}
	for _, path := range extensionProducerSources {
		refs[path] = true
	}
	directory := t.TempDir()
	for ref := range refs {
		data, err := readExtensionSourceFile(root, ref)
		if err != nil {
			t.Fatal(err)
		}
		if strings.HasPrefix(ref, "contracts/extensions/") {
			decoded, err := decodeContract(data)
			if err != nil {
				t.Fatal(err)
			}
			object := decoded.(map[string]any)
			mutate(ref, object)
			data, err = json.Marshal(object)
			if err != nil {
				t.Fatal(err)
			}
		}
		path := filepath.Join(directory, filepath.FromSlash(ref))
		if err := os.MkdirAll(filepath.Dir(path), 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, data, 0600); err != nil {
			t.Fatal(err)
		}
	}
	return directory
}

func checkUnavailableExtensionGeneration(t *testing.T, root string) {
	t.Run("unclaimable profile with reserved binding", func(t *testing.T) {
		directory := extensionInputFixture(t, root, func(path string, object map[string]any) {
			switch path {
			case "contracts/extensions/fragments/core00.recognition.json":
				for _, raw := range object["facts"].([]any) {
					fact := raw.(map[string]any)
					if fact["profile_id"] == "enterprise_authentication" {
						fact["claimable"], fact["contract_major"] = false, nil
					}
				}
			case "contracts/extensions/build/implementation-bindings.json":
				for _, raw := range object["bindings"].([]any) {
					binding := raw.(map[string]any)
					if binding["profile_id"] == "enterprise_authentication" {
						binding["contract_major"] = nil
					}
				}
			}
		})
		generated, err := deriveExtensionArtifacts(directory)
		if err != nil {
			t.Fatalf("generate reserved binding: %v", err)
		}
		found := false
		for _, item := range generated {
			if item.Path != generatedExtensionPrefix+"implementation-bindings/enterprise_authentication.json" {
				continue
			}
			decoded, err := decodeContract([]byte(item.JSON))
			if err != nil {
				t.Fatal(err)
			}
			if decoded.(map[string]any)["contract_major"] != nil {
				t.Fatal("binding invented an adopted major")
			}
			found = true
		}
		if !found {
			t.Fatal("declared reservation binding disappeared")
		}
	})
	t.Run("unrecognized implementation owner", func(t *testing.T) {
		directory := extensionInputFixture(t, root, func(path string, object map[string]any) {
			if path != "contracts/extensions/build/implementation-bindings.json" {
				return
			}
			rows := object["bindings"].([]any)
			body, err := json.Marshal(rows[0])
			if err != nil {
				t.Fatal(err)
			}
			unknown, err := decodeContract([]byte(strings.ReplaceAll(string(body), "enterprise_authentication", "unrecognized")))
			if err != nil {
				t.Fatal(err)
			}
			object["bindings"] = append(rows, unknown)
		})
		if _, err := deriveExtensionArtifacts(directory); err == nil {
			t.Fatal("generator accepted an implementation for an unrecognized profile")
		}
	})
	baseline, err := deriveExtensionArtifacts(root)
	if err != nil {
		t.Fatal(err)
	}
	baselineByPath := map[string]string{}
	for _, item := range baseline {
		baselineByPath[item.Path] = item.JSON
	}
	t.Run("unclaimable reserved profile", func(t *testing.T) {
		directory := extensionInputFixture(t, root, func(path string, object map[string]any) {
			if path == "contracts/extensions/build/implementation-bindings.json" {
				kept := []any{}
				for _, raw := range object["bindings"].([]any) {
					if raw.(map[string]any)["profile_id"] != "enterprise_authentication" {
						kept = append(kept, raw)
					}
				}
				object["bindings"] = kept
			}
			if path == "contracts/extensions/fragments/core00.recognition.json" {
				for _, raw := range object["facts"].([]any) {
					fact := raw.(map[string]any)
					if fact["profile_id"] == "enterprise_authentication" {
						fact["claimable"], fact["contract_major"] = false, nil
					}
				}
			}
		})
		generated, err := deriveExtensionArtifacts(directory)
		if err != nil {
			t.Fatalf("generate unclaimable reserved profile: %v", err)
		}
		found := false
		for _, item := range generated {
			if item.Path != generatedExtensionPrefix+"descriptors/enterprise_authentication.json" {
				continue
			}
			decoded, err := decodeContract([]byte(item.JSON))
			if err != nil {
				t.Fatal(err)
			}
			descriptor := decoded.(map[string]any)
			if descriptor["claimable"] != false || descriptor["contract_major"] != nil || len(descriptor["route_families"].([]any)) != 4 {
				t.Fatal("generator lost the unclaimable reservation variant")
			}
			found = true
		}
		if !found {
			t.Fatal("generator omitted reserved descriptor")
		}
	})
	for _, profile := range []string{"enterprise_authentication", "import", "network_flow_activity"} {
		t.Run(profile, func(t *testing.T) {
			directory := extensionInputFixture(t, root, func(path string, object map[string]any) {
				if path != "contracts/extensions/build/implementation-bindings.json" {
					return
				}
				kept := []any{}
				for _, raw := range object["bindings"].([]any) {
					if raw.(map[string]any)["profile_id"] != profile {
						kept = append(kept, raw)
					}
				}
				object["bindings"] = kept
			})
			generated, err := deriveExtensionArtifacts(directory)
			if err != nil {
				t.Fatalf("generate declared profile without implementation: %v", err)
			}
			seen := map[string]bool{}
			for _, item := range generated {
				seen[item.Path] = true
				if item.Path == generatedExtensionPrefix+"implementation-bindings/"+profile+".json" {
					t.Fatal("generator invented an implementation")
				}
				if item.Path == generatedExtensionPrefix+"profile-registry.json" || item.Path == generatedExtensionPrefix+"backup-registry.json" || strings.HasPrefix(item.Path, generatedExtensionPrefix+"job-contracts/") || strings.HasPrefix(item.Path, generatedExtensionPrefix+"worker-runtime-contracts/") {
					if item.JSON != baselineByPath[item.Path] {
						t.Fatalf("availability changed declared catalog %s", item.Path)
					}
				}
				if item.Path == generatedExtensionPrefix+"state-registry.json" {
					decoded, err := decodeContract([]byte(item.JSON))
					if err != nil {
						t.Fatal(err)
					}
					rows, _ := objectArrayAllowEmpty(decoded.(map[string]any)["profiles"], "state profiles")
					for _, row := range rows {
						if row["profile_id"] == profile && row["implementation_binding_sha256"] != nil {
							t.Fatal("state catalog invented an implementation digest")
						}
					}
				}
			}
			for path := range baselineByPath {
				if (strings.HasPrefix(path, generatedExtensionPrefix+"job-contracts/") || strings.HasPrefix(path, generatedExtensionPrefix+"worker-runtime-contracts/")) && !seen[path] {
					t.Fatalf("lost durable catalog artifact %s", path)
				}
			}
		})
	}
}

func checkExtensionJobPolicyGeneration(t *testing.T, root string) {
	for _, policy := range []string{"forbidden", "none", "not_cancelable", "aggregate overflow", "retired job schema", "missing codes", "empty codes", "duplicate codes", "unsorted codes", "cancellation code", "invalid token", "missing minimum", "negative minimum", "excess minimum"} {
		t.Run(policy, func(t *testing.T) {
			changed := 0
			directory := extensionInputFixture(t, root, func(_ string, object map[string]any) {
				facts, _ := object["facts"].([]any)
				for _, raw := range facts {
					row, _ := raw.(map[string]any)["job_kind_contract"].(map[string]any)
					if row["job_kind"] != "import.apply_v1" || row["proof_policy"] == nil {
						continue
					}
					changed++
					switch policy {
					case "forbidden":
						row["proof_policy"], row["resource_ref_contracts"] = "forbidden", []any{}
					case "none":
						row["idempotency_policy"], row["idempotency_identity_schema_id"] = "none", nil
					case "not_cancelable":
						row["cancellation_policy"] = "not_cancelable"
					case "aggregate overflow":
						row["resource_ref_contracts"].([]any)[0].(map[string]any)["max_refs"] = 2
					case "retired job schema":
						row["schema_id"] = "cartulary.extension_job_kind_contract.v2"
					case "missing codes":
						delete(row, "terminal_success_codes")
					case "empty codes":
						row["terminal_success_codes"] = []any{}
					case "duplicate codes":
						row["terminal_success_codes"] = []any{"done", "done"}
					case "unsorted codes":
						row["terminal_success_codes"] = []any{"z", "a"}
					case "cancellation code":
						row["terminal_success_codes"] = []any{"job_canceled"}
					case "invalid token":
						row["terminal_success_codes"] = []any{"owner.code"}
					case "missing minimum":
						delete(row["resource_ref_contracts"].([]any)[0].(map[string]any), "min_refs")
					case "negative minimum":
						row["resource_ref_contracts"].([]any)[0].(map[string]any)["min_refs"] = -1
					case "excess minimum":
						row["resource_ref_contracts"].([]any)[0].(map[string]any)["min_refs"] = 2
					}
				}
			})
			if changed != 1 {
				t.Fatalf("mutated %d job facts", changed)
			}
			_, err := deriveExtensionArtifacts(directory)
			if policy != "forbidden" && policy != "none" && policy != "not_cancelable" {
				if err == nil {
					t.Fatal("generated invalid selected terminal contract")
				}
				return
			}
			if err != nil {
				t.Fatalf("adopted policy rejected: %v", err)
			}
		})
	}
	t.Run("base selected terminal contract", func(t *testing.T) {
		raw, err := os.ReadFile(filepath.Join(root, "contracts/jobs/base.v2.json"))
		if err != nil {
			t.Fatal(err)
		}
		for _, name := range []string{"current", "retired", "missing codes", "empty codes", "cancellation code", "missing minimum"} {
			t.Run(name, func(t *testing.T) {
				value, err := decodeContract(raw)
				if err != nil {
					t.Fatal(err)
				}
				object := value.(map[string]any)
				row := object["jobs"].([]any)[0].(map[string]any)
				switch name {
				case "retired":
					object["schema_id"] = "cartulary.base_job_catalog.v1"
				case "missing codes":
					delete(row, "terminal_success_codes")
				case "empty codes":
					row["terminal_success_codes"] = []any{}
				case "cancellation code":
					row["terminal_success_codes"] = []any{"job_canceled"}
				case "missing minimum":
					row["resource_ref_contracts"] = []any{map[string]any{"resource_ref_kind": "incident", "resource_id_schema_id": "cartulary.common_job_resource_ref_id.v1", "max_refs": json.Number("1")}}
				}
				if err := validateBaseJobCatalog(object); (err == nil) != (name == "current") {
					t.Fatalf("admission error: %v", err)
				}
			})
		}
	})
}
