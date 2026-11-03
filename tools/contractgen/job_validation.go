package main

import (
	"fmt"

	"github.com/JochiRaider/cartulary/internal/platform/jobs/jobcontract"
)

// This adapter supplies the same pure Core policy used by actual publication
// and retained-proof admission; owner fragments remain the authored inputs.
func jobTerminalContract(object map[string]any, label string) (jobcontract.TerminalContract, error) {
	var contract jobcontract.TerminalContract
	schema, err := requiredString(object, "terminal_result_schema_id", label)
	if err != nil {
		return contract, err
	}
	rawCodes, ok := object["terminal_success_codes"].([]any)
	if !ok || len(rawCodes) < 1 || len(rawCodes) > 64 {
		return contract, fmt.Errorf("%s.terminal_success_codes must contain 1..64 values", label)
	}
	codes, err := sortedUniqueStringArray(rawCodes, label+".terminal_success_codes", false)
	if err != nil {
		return contract, err
	}
	contract.SchemaID, contract.SuccessCodes = schema, codes
	refs, ok := object["resource_ref_contracts"].([]any)
	if !ok || len(refs) > 64 {
		return contract, fmt.Errorf("%s.resource_ref_contracts must contain 0..64 rows", label)
	}
	for index, raw := range refs {
		row, ok := raw.(map[string]any)
		rowLabel := fmt.Sprintf("%s.resource_ref_contracts[%d]", label, index)
		if !ok {
			return contract, fmt.Errorf("%s must be an object", rowLabel)
		}
		if err := requireExtensionExactKeys(row, stringSet("resource_ref_kind", "resource_id_schema_id", "min_refs", "max_refs"), rowLabel); err != nil {
			return contract, err
		}
		kind, err := requiredString(row, "resource_ref_kind", rowLabel)
		if err != nil {
			return contract, err
		}
		identity, err := requiredString(row, "resource_id_schema_id", rowLabel)
		if err != nil {
			return contract, err
		}
		maximum, err := positiveJSONInt(row["max_refs"], rowLabel+".max_refs")
		if err != nil || maximum > 1024 {
			return contract, fmt.Errorf("%s.max_refs must be in 1..1024", rowLabel)
		}
		minimum, err := nonnegativeJSONInt(row["min_refs"], rowLabel+".min_refs")
		if err != nil || minimum > maximum {
			return contract, fmt.Errorf("%s.min_refs must be in 0..max_refs", rowLabel)
		}
		contract.Resources = append(contract.Resources, jobcontract.ResourceContract{Kind: kind, SchemaID: identity, MinRefs: int(minimum), MaxRefs: int(maximum)})
	}
	if err := jobcontract.ValidateTerminalContract(contract); err != nil {
		return contract, fmt.Errorf("%s: %w", label, err)
	}
	return contract, nil
}

func validateBaseJobCatalog(value any) error {
	object, ok := value.(map[string]any)
	if !ok {
		return fmt.Errorf("base Jobs catalog must be an object")
	}
	if err := requireExtensionExactKeys(object, stringSet("schema_id", "jobs", "workers"), "base Jobs catalog"); err != nil {
		return err
	}
	if object["schema_id"] != "cartulary.base_job_catalog.v2" {
		return fmt.Errorf("unsupported base Jobs catalog")
	}
	definitions, ok := object["jobs"].([]any)
	if !ok || len(definitions) == 0 {
		return fmt.Errorf("base Jobs catalog requires explicit definitions")
	}
	handlers := make(map[string]string, len(definitions))
	for _, raw := range definitions {
		row, ok := raw.(map[string]any)
		if !ok {
			return fmt.Errorf("base Job definition must be an object")
		}
		if err := requireExtensionExactKeys(row, stringSet("job_kind", "progress_unit_id", "handler_name", "terminal_result_schema_id", "terminal_success_codes", "resource_ref_contracts"), "base Job definition"); err != nil {
			return err
		}
		kind, err := requiredString(row, "job_kind", "base Job definition")
		if err != nil {
			return err
		}
		handler, err := requiredString(row, "handler_name", "base Job definition")
		if err != nil {
			return err
		}
		if _, duplicate := handlers[kind]; duplicate || !validExtensionProgressUnitID(stringValue(row["progress_unit_id"])) {
			return fmt.Errorf("base Job definition has invalid identity")
		}
		handlers[kind] = handler
		if _, err := jobTerminalContract(row, "base Job definition"); err != nil {
			return err
		}
	}
	workers, ok := object["workers"].([]any)
	if !ok {
		return fmt.Errorf("base workers must be an array")
	}
	assigned := make(map[string]bool, len(handlers))
	seenWorkers := make(map[string]bool, len(workers))
	for _, raw := range workers {
		row, ok := raw.(map[string]any)
		if !ok {
			return fmt.Errorf("base worker must be an object")
		}
		if err := requireExtensionExactKeys(row, stringSet("profile_id", "worker_kind", "job_kinds", "max_active_attempts_per_process"), "base worker"); err != nil {
			return err
		}
		worker, err := requiredString(row, "worker_kind", "base worker")
		if err != nil {
			return err
		}
		maximum, err := positiveJSONInt(row["max_active_attempts_per_process"], "base worker capacity")
		if err != nil || maximum > 8 || row["profile_id"] != "base" || seenWorkers[worker] {
			return fmt.Errorf("base worker has invalid identity or capacity")
		}
		seenWorkers[worker] = true
		kinds, err := sortedUniqueStringArray(row["job_kinds"], "base worker kinds", false)
		if err != nil {
			return err
		}
		for _, kind := range kinds {
			if handlers[kind] != worker || assigned[kind] {
				return fmt.Errorf("base Job assignment is invalid")
			}
			assigned[kind] = true
		}
	}
	if len(assigned) != len(handlers) {
		return fmt.Errorf("base Job assignment is incomplete")
	}
	return nil
}
