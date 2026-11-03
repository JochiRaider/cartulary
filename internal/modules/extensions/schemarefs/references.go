// Package schemarefs projects the explicit schemas supplied by an Extensions
// profile's operational contracts. It performs no I/O or semantic execution;
// each actual owner remains responsible for admitting its implementation.
package schemarefs

import (
	"fmt"
	"github.com/JochiRaider/cartulary/internal/platform/jobs/jobcontract"
	"sort"
	"strings"
)

// Inventory contains exact sets from already shape-admitted machine objects.
// Inputs may contain authored fragments or their generated projections. The
// implementation binding itself is excluded: it cannot authorize its own refs.
// Only contract members are visited; defaults, data and JSON Schema definitions
// are never interpreted as another operational contract.
type Inventory struct {
	Required   map[string][]string
	schemas    map[string]struct{}
	algorithms map[string]struct{}
}

// Collect indexes declared operational references. Reference closure does not
// replace matching a configuration, participant or algorithm to its typed owner.
func Collect(objects []map[string]any) Inventory {
	inventory := Inventory{schemas: map[string]struct{}{}, algorithms: map[string]struct{}{}}

	sets := map[string]map[string]struct{}{}
	var collect func(string, map[string]any)
	collect = func(profile string, object map[string]any) {
		if profile == "" {
			return
		}
		set := sets[profile]
		if set == nil {
			set = map[string]struct{}{}
			sets[profile] = set
		}
		for key, value := range object {
			if key == "schema_id" || strings.HasSuffix(key, "_schema_id") || key == "codec_contract_ref" || key == "participant_contract_ref" {
				if id, ok := value.(string); ok && id != "" {
					set[id] = struct{}{}
					inventory.schemas[id] = struct{}{}
				}
			}
			if strings.HasSuffix(key, "_schema_ids") {
				if values, ok := value.([]any); ok {
					for _, raw := range values {
						if id, ok := raw.(string); ok && id != "" {
							set[id] = struct{}{}
							inventory.schemas[id] = struct{}{}
						}
					}
				}
			}
			if key == "algorithm_id" || strings.HasSuffix(key, "_algorithm_id") || key == "preflight_algorithmid" || key == "post_migration_algorithmid" {
				if id, ok := value.(string); ok && id != "" {
					inventory.algorithms[id] = struct{}{}
				}
			}
		}

		if object["fact_kind"] == "worker_kind" {
			set["cartulary.extension_worker_runtime_contract.v1"] = struct{}{}
		}
		for _, key := range []string{"state_ownership", "job_kind_contract", "contribution", "migration_definition", "initialization", "admission_validation"} {
			if child, ok := object[key].(map[string]any); ok {
				collect(profile, child)
			}
		}
		for _, key := range []string{"keys", "operations", "resource_ref_contracts", "bindings", "contributions"} {
			if rows, ok := object[key].([]any); ok {
				for _, raw := range rows {
					if child, ok := raw.(map[string]any); ok {
						collect(profile, child)
					}
				}
			}
		}
	}
	for _, object := range objects {
		if id, ok := object["$id"].(string); ok {
			inventory.schemas[id] = struct{}{}
		}
		if id, ok := object["schema_id"].(string); ok {
			inventory.schemas[id] = struct{}{}
		}
		if object["schema_id"] == "cartulary.extension_contract_definition_set.v3" {
			if rows, ok := object["definitions"].([]any); ok {
				for _, raw := range rows {
					if row, ok := raw.(map[string]any); ok {
						if id, ok := row["schema_id"].(string); ok {
							inventory.schemas[id] = struct{}{}
						}
					}
				}
			}
		}

		if object["schema_id"] == "cartulary.extension_implementation_binding.v2" {
			continue
		}
		profile, _ := object["profile_id"].(string)
		if profile == "" {
			profile, _ = object["owner_profile_id"].(string)
		}
		if profile == "" {
			if binding, ok := object["binding_id"].(string); ok {
				profile, _, _ = strings.Cut(binding, ".")
			}
		}
		if profile != "" {
			collect(profile, object)
		}
		if object["schema_id"] == "cartulary.extension_owner_fragment.v3" {
			if rows, ok := object["facts"].([]any); ok {
				for _, raw := range rows {
					if fact, ok := raw.(map[string]any); ok {
						profile, _ := fact["profile_id"].(string)
						collect(profile, fact)
					}
				}
			}
		}
	}
	result := make(map[string][]string, len(sets))
	for profile, set := range sets {
		refs := make([]string, 0, len(set))
		for id := range set {
			refs = append(refs, id)
		}
		sort.Strings(refs)
		result[profile] = refs
	}
	inventory.Required = result
	return inventory
}

// ValidateImports rejects imported labels with no operational reference or
// packaged schema definition. Actual executable support is checked at the
// typed owner boundary, never conferred by an import declaration.
func (inventory Inventory) ValidateImports(dependency map[string]any) error {
	rows, ok := dependency["dependencies"].([]any)
	if !ok {
		return fmt.Errorf("dependency rows are missing")
	}
	for _, raw := range rows {
		row, ok := raw.(map[string]any)
		if !ok {
			return fmt.Errorf("invalid dependency row")
		}
		for _, field := range []string{"imported_schema_ids", "imported_algorithm_ids"} {
			values, ok := row[field].([]any)
			if !ok {
				return fmt.Errorf("invalid dependency import list")
			}
			for _, raw := range values {
				id, ok := raw.(string)
				if !ok {
					return fmt.Errorf("invalid imported identity")
				}
				known := inventory.schemas
				if field == "imported_algorithm_ids" {
					known = inventory.algorithms
				}
				if field == "imported_schema_ids" {
					_, terminal := jobcontract.ResolveTerminal(id)
					if terminal || jobcontract.HasIdentitySchema(id) || id == jobcontract.ResourceIDSchema || id == jobcontract.ResourceRefSchema {
						continue
					}
				}
				if _, ok := known[id]; !ok {
					return fmt.Errorf("unresolved dependency import")
				}
			}
		}
	}
	return nil
}
