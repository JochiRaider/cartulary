package extensions

import (
	"sort"
	"strings"
	"testing"
)

func checkJobBindingReferences(t *testing.T) {
	t.Run("exact supporting schemas", checkExactSupportingSchemas)
	t.Run("operational dependency imports", checkOperationalDependencyImports)
	source := newMutableGeneratedSource()
	count := 0
	for path, artifact := range source {
		if !strings.HasPrefix(path, generatedExtensionsRoot+"implementation-bindings/") {
			continue
		}
		binding := decodeMutableObject(t, artifact.JSON)
		rows, ok := objectSlice(binding["job_kind_contracts"])
		if !ok {
			t.Fatal("missing job references")
		}
		for _, row := range rows {
			if err := requireExactKeys(row, "job_kind", "job_kind_contract_sha256"); err != nil {
				t.Fatalf("binding embeds a job body instead of its exact reference: %v", err)
			}
			jobPath := generatedExtensionsRoot + "job-contracts/" + stringValue(binding["profile_id"]) + "/" + stringValue(row["job_kind"]) + ".json"
			artifact, ok := source[jobPath]
			if !ok || row["job_kind_contract_sha256"] != artifact.SHA256 {
				t.Fatal("unresolved job contract reference")
			}
			count++
		}
	}
	if count == 0 {
		t.Fatal("fixture has no job references")
	}
	coordinator, err := NewCoordinator(source)
	if err != nil {
		t.Fatalf("current job references rejected: %#v", err)
	}
	for _, contract := range coordinator.JobKindContracts() {
		path := generatedExtensionsRoot + "job-contracts/" + contract.ProfileID + "/" + contract.JobKind + ".json"
		if contract.SHA256() != source[path].SHA256 {
			t.Fatal("typed catalog changed the admitted job contract identity")
		}
	}
	for _, name := range []string{"embedded body", "missing digest", "extra member", "stale digest", "duplicate reference"} {
		t.Run(name, func(t *testing.T) {
			malformed := newMutableGeneratedSource()
			rewriteCoherentPackage(t, malformed, func(object map[string]any) {
				if object["schema_id"] != "cartulary.extension_implementation_binding.v2" {
					return
				}
				rows, ok := object["job_kind_contracts"].([]any)
				if !ok || len(rows) == 0 {
					return
				}
				reference := rows[0].(map[string]any)
				switch name {
				case "embedded body":
					path := generatedExtensionsRoot + "job-contracts/" + stringValue(object["profile_id"]) + "/" + stringValue(reference["job_kind"]) + ".json"
					rows[0] = decodeMutableObject(t, malformed[path].JSON)
				case "missing digest":
					delete(reference, "job_kind_contract_sha256")
				case "extra member":
					reference["extra"] = true
				case "stale digest":
					reference["job_kind_contract_sha256"] = strings.Repeat("0", 64)
				case "duplicate reference":
					object["job_kind_contracts"] = append(rows, reference)
				}
			})
			if _, err := NewCoordinator(malformed); err == nil {
				t.Fatal("invalid job reference admitted")
			}
		})
	}

}

func checkMalformedBindingContracts(t *testing.T) {
	for _, name := range []string{"job proof budget", "job terminal schema", "retired job schema", "missing success codes", "empty success codes", "duplicate success codes", "unsorted success codes", "cancellation success code", "invalid success token", "missing minimum", "negative minimum", "inverted bounds", "worker budget", "worker duplicate assignment"} {
		t.Run(name, func(t *testing.T) {
			source := newMutableGeneratedSource()
			rewriteCoherentPackage(t, source, func(object map[string]any) {
				if object["schema_id"] == "cartulary.extension_job_kind_contract.v3" && object["job_kind"] != nil {
					switch name {
					case "job proof budget":
						object["max_proof_bytes"] = 0
					case "job terminal schema":
						object["terminal_result_schema_id"] = "cartulary.unrecognized.v1"
					case "retired job schema":
						object["schema_id"] = "cartulary.extension_job_kind_contract.v2"
					case "missing success codes":
						delete(object, "terminal_success_codes")
					case "empty success codes":
						object["terminal_success_codes"] = []any{}
					case "duplicate success codes":
						object["terminal_success_codes"] = []any{"done", "done"}
					case "unsorted success codes":
						object["terminal_success_codes"] = []any{"z", "a"}
					case "cancellation success code":
						object["terminal_success_codes"] = []any{"job_canceled"}
					case "invalid success token":
						object["terminal_success_codes"] = []any{"unsupported.token"}
					case "missing minimum", "negative minimum", "inverted bounds":
						rows, _ := object["resource_ref_contracts"].([]any)
						if len(rows) != 0 {
							row := rows[0].(map[string]any)
							if name == "missing minimum" {
								delete(row, "min_refs")
							}
							if name == "negative minimum" {
								row["min_refs"] = -1
							}
							if name == "inverted bounds" {
								row["min_refs"] = 1025
							}
						}
					}
				}
				if (object["schema_id"] == "cartulary.extension_worker_runtime_contract.v1" || object["fact_kind"] == "worker_kind") && object["worker_kind"] != nil {
					switch name {
					case "worker budget":
						object["max_active_attempts_per_process"] = 9
					case "worker duplicate assignment":
						if jobs, ok := object["job_kinds"].([]any); ok && len(jobs) > 0 {
							object["job_kinds"] = append(jobs, jobs[0])
						}
					}
				}
			})
			_, err := NewCoordinator(source)
			if err == nil {
				t.Fatal("malformed binding admitted")
			}
			expected := "extension_registry_invalid"
			if name == "job terminal schema" || name == "unsorted success codes" {
				// Schema-valid contracts rejected by typed binding admission use
				// the implementation-boundary diagnostic. Package schema failures
				// are rejected earlier by registry admission.
				expected = "extension_implementation_unavailable"
			}
			if err.Error() != expected {
				t.Fatalf("binding diagnostic = %v, want %s", err, expected)
			}
		})
	}
}

func checkExactSupportingSchemas(t *testing.T) {
	for _, mutation := range []string{"empty", "missing job identity", "missing configuration", "missing worker", "missing Reporting source schema", "extra supported", "extra unknown", "extra participant algorithm"} {
		t.Run(mutation, func(t *testing.T) {
			source := newMutableGeneratedSource()
			rewriteCoherentPackage(t, source, func(object map[string]any) {
				profile := "import"
				if mutation == "missing configuration" {
					profile = "enterprise_authentication"
				}
				if mutation == "missing Reporting source schema" || mutation == "extra participant algorithm" {
					profile = "network_flow_activity"
				}
				if object["schema_id"] != "cartulary.extension_implementation_binding.v2" || object["profile_id"] != profile {
					return
				}
				refs, _ := stringSlice(object["supporting_schema_ids"])
				switch mutation {
				case "empty":
					refs = []string{}
				case "missing job identity", "missing configuration", "missing worker", "missing Reporting source schema":
					removed := "cartulary.route_scoped_idempotency_identity.v1"
					if mutation == "missing configuration" {
						removed = "cartulary.enterprise_authentication_provider_manifest_path.v1"
					}
					if mutation == "missing worker" {
						removed = "cartulary.extension_worker_runtime_contract.v1"
					}
					if mutation == "missing Reporting source schema" {
						removed = "cartulary.network_flow.reporting_graph_source.v2"
					}
					next := []string{}
					for _, id := range refs {
						if id != removed {
							next = append(next, id)
						}
					}
					refs = next
				case "extra participant algorithm":
					rows, _ := objectSlice(object["participant_contracts"])
					ids, _ := stringSlice(rows[0]["algorithm_ids"])
					ids = append(ids, "network_flow_activity.unbound_v1")
					sort.Strings(ids)
					values := make([]any, len(ids))
					for i, id := range ids {
						values[i] = id
					}
					rows[0]["algorithm_ids"] = values
				case "extra supported":
					refs = append(refs, "cartulary.extension_state_metadata.v1")
				case "extra unknown":
					refs = append(refs, "cartulary.unimplemented_schema.v1")
				}
				sort.Strings(refs)
				values := make([]any, len(refs))
				for i, id := range refs {
					values[i] = id
				}
				object["supporting_schema_ids"] = values
			})
			_, err := NewCoordinator(source)
			requireFinding(t, err, "extension_implementation_unavailable", "implementation_binding")
		})
	}
}

func checkOperationalDependencyImports(t *testing.T) {
	for _, mutation := range []string{"schema", "algorithm", "retired harness"} {
		t.Run(mutation, func(t *testing.T) {
			source := newMutableGeneratedSource()
			rewriteCoherentPackage(t, source, func(object map[string]any) {
				if object["schema_id"] != "cartulary.extension_dependency_snapshot.v3" {
					return
				}
				rows, ok := object["dependencies"].([]any)
				if !ok {
					return
				}
				for _, raw := range rows {
					row := raw.(map[string]any)
					if row["dependency_id"] != "testing_harness" {
						continue
					}
					field, id := "imported_schema_ids", "cartulary.unimplemented.v1"
					if mutation == "algorithm" {
						field, id = "imported_algorithm_ids", "unimplemented_v1"
					}
					if mutation == "retired harness" {
						id = "cartulary.harness_row_result.v5"
					}
					row[field] = []any{id}
				}
			})
			_, err := NewCoordinator(source)
			requireFinding(t, err, "extension_registry_invalid", "registry_generation")
		})
	}
}
