package extensions

import (
	"sort"
	"strings"
)

// Durable job definitions remain available for inactive reconciliation even
// when this build has no implementation of their owner. The sealed companion
// joins prove these artifacts equal the owner facts before they reach here.
func admitDeclaredProfileContracts(profile string, source *admittedPackage) (bindingContracts, error) {
	jobReferences := []any{}
	workerContracts := []any{}
	workerKinds := []string{}
	for _, path := range source.paths {
		object := source.objects[path]
		if strings.HasPrefix(path, generatedExtensionsRoot+"job-contracts/"+profile+"/") {
			jobReferences = append(jobReferences, map[string]any{"job_kind": object["job_kind"], "job_kind_contract_sha256": source.artifacts[path].SHA256})
		}
		if strings.HasPrefix(path, generatedExtensionsRoot+"worker-runtime-contracts/"+profile+"/") {
			workerContracts = append(workerContracts, object)
			workerKinds = append(workerKinds, stringValue(object["worker_kind"]))
		}
	}
	jobs, err := admitJobKindContracts(profile, jobReferences, source)
	if err != nil {
		return bindingContracts{}, err
	}
	workers, err := parseWorkerRuntimeContracts(profile, workerContracts, workerKinds, jobs)
	if err != nil {
		return bindingContracts{}, err
	}
	return bindingContracts{jobs: jobs, workers: workers}, nil
}

func (p *admittedPackage) declaredJobAndWorkerContracts(profiles []map[string]any) (map[string]map[string]any, map[string]map[string]any, error) {
	known := map[string]bool{}
	for _, profile := range profiles {
		known[stringValue(profile["profile_id"])] = true
	}
	jobs, workers := map[string]map[string]any{}, map[string]map[string]any{}
	fragments, ok := objectSlice(p.objects[generatedExtensionsRoot+"owner-input-registry.json"]["owner_fragments"])
	if !ok {
		return nil, nil, registryInvalid("owner_fragments_invalid")
	}
	for _, fragment := range fragments {
		facts, ok := objectSlice(fragment["facts"])
		if !ok {
			return nil, nil, registryInvalid("owner_facts_invalid")
		}
		for _, fact := range facts {
			profile := stringValue(fact["profile_id"])
			var object map[string]any
			var identity, directory string
			var set map[string]map[string]any
			switch fact["fact_kind"] {
			case "job_kind":
				object, _ = fact["job_kind_contract"].(map[string]any)
				identity, directory, set = "job_kind", "job-contracts", jobs
			case "worker_kind":
				object = cloneObject(fact)
				delete(object, "fact_kind")
				object["schema_id"] = "cartulary.extension_worker_runtime_contract.v1"
				identity, directory, set = "worker_kind", "worker-runtime-contracts", workers
			default:
				continue
			}
			path := generatedExtensionsRoot + directory + "/" + profile + "/" + stringValue(object[identity]) + ".json"
			if !known[profile] || object["profile_id"] != profile || set[path] != nil || !equalCanonicalObjects(object, p.objects[path]) {
				return nil, nil, registryInvalid("durable_catalog_owner_projection_mismatch")
			}
			set[path] = object
		}
	}
	return jobs, workers, nil
}

func (p *admittedPackage) declaredMigrationBindings(profile string) []map[string]any {
	result := []map[string]any{}
	fragments, _ := objectSlice(p.objects[generatedExtensionsRoot+"owner-input-registry.json"]["owner_fragments"])
	for _, fragment := range fragments {
		facts, _ := objectSlice(fragment["facts"])
		for _, fact := range facts {
			if fact["profile_id"] != profile || fact["fact_kind"] != "migration_definition" {
				continue
			}
			definition, _ := fact["migration_definition"].(map[string]any)
			result = append(result, map[string]any{
				"migration_id": definition["migration_id"], "from_state_version": definition["from_state_version"], "to_state_version": definition["to_state_version"],
				"migration_definition_sha256": canonicalDigest(definition), "apply_algorithm_id": definition["apply_algorithm_id"], "validation_algorithm_id": definition["validation_algorithm_id"],
			})
		}
	}
	sort.Slice(result, func(i, j int) bool {
		for _, key := range []string{"from_state_version", "to_state_version"} {
			left, right := intValue(result[i][key]), intValue(result[j][key])
			if left != right {
				return left < right
			}
		}
		return stringValue(result[i]["migration_id"]) < stringValue(result[j]["migration_id"])
	})
	return result
}
