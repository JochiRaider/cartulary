package extensions

import (
	"errors"
	"github.com/JochiRaider/cartulary/internal/modules/extensions/registryfacts"
	"sort"
	"strings"
)

func (p *admittedPackage) validateOwnerInputs() error {
	dependency := p.objects[generatedExtensionsRoot+"dependency-snapshot.json"]
	if err := validateDependencySnapshot(dependency); err != nil {
		return err
	}
	input := p.objects[generatedExtensionsRoot+"owner-input-registry.json"]
	if requireExactKeys(input, "schema_id", "dependency_snapshot", "owner_fragments") != nil || input["schema_id"] != "cartulary.extension_owner_input_registry.v2" {
		return registryInvalid("owner_input_shape_invalid")
	}
	embedded, ok := input["dependency_snapshot"].(map[string]any)
	if !ok || !equalCanonicalObjects(embedded, dependency) {
		return registryInvalid("owner_dependency_mismatch")
	}
	fragments, ok := objectSlice(input["owner_fragments"])
	if !ok {
		return registryInvalid("owner_fragments_not_array")
	}
	normalized, err := registryfacts.NormalizeFragments(fragments)
	if err != nil || !equalCanonicalObjects(map[string]any{"fragments": fragments}, map[string]any{"fragments": normalized}) {
		return registryInvalid("owner_input_order_invalid")
	}
	packaged := map[string]map[string]any{}
	for _, object := range p.objects {
		if object["schema_id"] == "cartulary.extension_owner_fragment.v3" {
			packaged[stringValue(object["owner_fragment_id"])] = object
		}
	}
	if len(fragments) != len(packaged) {
		return registryInvalid("owner_fragment_set_mismatch")
	}
	previous := ""
	for _, fragment := range fragments {
		id, owner := stringValue(fragment["owner_fragment_id"]), stringValue(fragment["owner_id"])
		if requireExactKeys(fragment, "schema_id", "owner_fragment_id", "owner_id", "facts") != nil || fragment["schema_id"] != "cartulary.extension_owner_fragment.v3" || !extensionProfileIDPattern.MatchString(owner) || !strings.HasPrefix(id, owner+".") || len(id) > 129 || id <= previous || !equalCanonicalObjects(fragment, packaged[id]) {
			return registryInvalid("owner_fragment_invalid")
		}
		previous = id
	}
	if err := p.registryDiagnostics(fragments); err != nil {
		return err
	}
	type ownedFact struct {
		owner string
		body  map[string]any
	}
	facts := map[string]map[string][]ownedFact{}
	seen := map[string]bool{}
	for _, fragment := range fragments {
		owner := stringValue(fragment["owner_id"])
		rows, ok := objectSlice(fragment["facts"])
		if !ok {
			return registryInvalid("owner_facts_missing")
		}
		previousFact := ""
		for _, fact := range rows {
			profile, kind := stringValue(fact["profile_id"]), stringValue(fact["fact_kind"])
			if !extensionProfileIDPattern.MatchString(profile) || validateOwnerFactShape(fact) != nil {
				return registryInvalid("owner_fact_shape_invalid")
			}
			identity, err := registryfacts.Identify(fact)
			if err != nil {
				return registryInvalid("owner_fact_identity_invalid")
			}
			key := identity.Key()
			if key <= previousFact || seen[key] {
				return registryInvalid("owner_fact_duplicate_or_unordered")
			}
			previousFact = key
			seen[key] = true
			if facts[profile] == nil {
				facts[profile] = map[string][]ownedFact{}
			}
			facts[profile][kind] = append(facts[profile][kind], ownedFact{owner: owner, body: fact})
		}
	}
	registry := p.objects[generatedExtensionsRoot+"profile-registry.json"]
	if requireExactKeys(registry, "schema_id", "profiles") != nil || registry["schema_id"] != "cartulary.extension_profile_registry.v1" {
		return registryInvalid("profile_registry_shape_invalid")
	}
	profiles, ok := objectSlice(registry["profiles"])
	if !ok {
		return registryInvalid("profiles_not_array")
	}
	if len(profiles) > 256 {
		return registryLimit("profiles", 256, len(profiles))
	}
	if len(profiles) != len(facts) {
		return registryInvalid("recognized_profile_set_mismatch")
	}
	previous = ""
	for _, descriptor := range profiles {
		profile := stringValue(descriptor["profile_id"])
		if profile <= previous || !extensionProfileIDPattern.MatchString(profile) {
			return registryInvalid("profile_order_invalid")
		}
		previous = profile
		byKind := facts[profile]
		for _, kind := range registryfacts.RequiredScalarKinds() {
			if len(byKind[kind]) != 1 {
				return registryInvalid("owner_scalar_fact_missing_or_duplicate")
			}
		}
		recognition := byKind["recognized_profile"][0].body
		primary := stringValue(recognition["primary_owner_id"])
		if descriptor["owner_id"] != primary || descriptor["claimable"] != recognition["claimable"] || descriptor["contract_major"] != recognition["contract_major"] {
			return registryInvalid("recognition_projection_mismatch")
		}
		for kind, rows := range byKind {
			owner := registryfacts.FactOwner(kind, primary)
			for _, row := range rows {
				if row.owner != owner {
					return registryInvalid("owner_allocation_invalid")
				}
			}
		}
		for _, pair := range [][2]string{{"claim_configuration", "claim_config_key"}, {"state_ownership", "state_ownership"}, {"admission_validation", "admission_validation"}, {"egress", "egress_mode"}, {"portability", "incident_portability_mode"}, {"snapshot_reporting", "snapshot_reporting_mode"}} {
			if canonicalDigest(descriptor[pair[1]]) != canonicalDigest(byKind[pair[0]][0].body[pair[1]]) {
				return registryInvalid("owner_scalar_projection_mismatch")
			}
		}
		configuration := p.objects["contracts/extensions/profiles/"+profile+"/configuration.json"]
		if configuration == nil || byKind["claim_configuration"][0].body["configuration_contract_sha256"] != canonicalDigest(configuration) {
			return registryInvalid("configuration_owner_digest_mismatch")
		}
		for _, pair := range [][3]string{{"route_family", "route_family", "route_families"}, {"workspace", "workspace_key", "workspace_keys"}, {"public_schema", "public_schema_id", "public_schema_ids"}} {
			values := make([]string, 0, len(byKind[pair[0]]))
			for _, row := range byKind[pair[0]] {
				values = append(values, stringValue(row.body[pair[1]]))
			}
			sort.Strings(values)
			declared, ok := stringSlice(descriptor[pair[2]])
			if !ok || !strictlySortedUnique(values) || len(values) != len(declared) {
				return registryInvalid("owner_collection_projection_mismatch")
			}
			for i, value := range values {
				if declared[i] != value {
					return registryInvalid("owner_collection_projection_mismatch")
				}
			}
		}
		dependencies := make([]map[string]any, 0, len(byKind["runtime_dependency"]))
		for _, row := range byKind["runtime_dependency"] {
			dependencies = append(dependencies, map[string]any{"profile_id": row.body["dependency_profile_id"], "required_contract_major": row.body["required_contract_major"]})
		}
		sort.Slice(dependencies, func(i, j int) bool {
			return stringValue(dependencies[i]["profile_id"]) < stringValue(dependencies[j]["profile_id"])
		})
		if canonicalDigest(dependencies) != canonicalDigest(descriptor["runtime_dependencies"]) {
			return registryInvalid("owner_dependency_projection_mismatch")
		}
		contributions := make([]map[string]any, 0, len(byKind["contribution"]))
		for _, row := range byKind["contribution"] {
			contribution, ok := row.body["contribution"].(map[string]any)
			if !ok {
				return registryInvalid("contribution_not_object")
			}
			if contribution["kind"] == "backup_restore_participant" {
				return registryInvalid("retired_profile_backup_participant")
			}
			contributions = append(contributions, contribution)
		}
		sort.Slice(contributions, func(i, j int) bool {
			return stringValue(contributions[i]["contribution_id"]) < stringValue(contributions[j]["contribution_id"])
		})
		if canonicalDigest(contributions) != canonicalDigest(descriptor["contributions"]) {
			return registryInvalid("owner_contribution_projection_mismatch")
		}
		binding := p.objects[generatedExtensionsRoot+"implementation-bindings/"+profile+".json"]
		jobs := make([]map[string]any, 0, len(byKind["job_kind"]))
		for _, row := range byKind["job_kind"] {
			job, ok := row.body["job_kind_contract"].(map[string]any)
			if !ok {
				return registryInvalid("job_fact_not_object")
			}
			jobs = append(jobs, map[string]any{"job_kind": job["job_kind"], "job_kind_contract_sha256": canonicalDigest(job)})
		}
		sort.Slice(jobs, func(i, j int) bool { return stringValue(jobs[i]["job_kind"]) < stringValue(jobs[j]["job_kind"]) })
		workers := make([]map[string]any, 0, len(byKind["worker_kind"]))
		for _, row := range byKind["worker_kind"] {
			workers = append(workers, map[string]any{"schema_id": "cartulary.extension_worker_runtime_contract.v1", "profile_id": profile, "worker_kind": row.body["worker_kind"], "job_kinds": row.body["job_kinds"], "max_active_attempts_per_process": row.body["max_active_attempts_per_process"]})
		}
		sort.Slice(workers, func(i, j int) bool {
			return stringValue(workers[i]["worker_kind"]) < stringValue(workers[j]["worker_kind"])
		})
		if binding != nil && (canonicalDigest(jobs) != canonicalDigest(binding["job_kind_contracts"]) || canonicalDigest(workers) != canonicalDigest(binding["worker_runtime_contracts"])) {
			return registryInvalid("owner_job_worker_projection_mismatch")
		}
	}
	return nil
}

func registryInputLimitError(err error) error {
	var limit *registryfacts.LimitError
	if errors.As(err, &limit) {
		return registryLimit(limit.Quantity, limit.Limit, limit.Actual)
	}
	return registryInvalid("owner_input_limits_invalid")
}

// Resource ceilings precede element/schema diagnostics once the owning array
// exists. This preserves the required limit classification for maxItems too.
func checkRegistryInputLimits(object map[string]any) error {
	switch object["schema_id"] {
	case "cartulary.extension_profile_registry.v1":
		if rows, ok := object["profiles"].([]any); ok && len(rows) > registryfacts.ProfileLimit {
			return registryLimit("profiles", registryfacts.ProfileLimit, len(rows))
		}
	case "cartulary.extension_owner_input_registry.v2":
		if rows, ok := object["owner_fragments"].([]any); ok && len(rows) > registryfacts.FragmentLimit {
			return registryLimit("owner_fragments", registryfacts.FragmentLimit, len(rows))
		}
		fragments, ok := objectSlice(object["owner_fragments"])
		if !ok {
			return registryInvalid("owner_fragments_not_array")
		}
		if err := registryfacts.CheckOwnerInputLimits(fragments); err != nil {
			return registryInputLimitError(err)
		}
	}
	return nil
}

func validateDependencySnapshot(object map[string]any) error {
	if requireExactKeys(object, "schema_id", "dependencies") != nil || object["schema_id"] != "cartulary.extension_dependency_snapshot.v3" {
		return registryInvalid("dependency_snapshot_shape_invalid")
	}
	expected := []string{"core00", "core01", "core02", "core03", "core04", "network_flow_activity", "opentelemetry", "report_composition", "reporting", "testing_harness"}
	rows, ok := objectSlice(object["dependencies"])
	if !ok || len(rows) != len(expected) {
		return registryInvalid("dependency_set_invalid")
	}
	for i, row := range rows {
		if requireExactKeys(row, "dependency_id", "imported_schema_ids", "imported_algorithm_ids", "imported_artifacts") != nil || row["dependency_id"] != expected[i] {
			return registryInvalid("dependency_identity_invalid")
		}
		schemas := map[string]bool{}
		for _, key := range []string{"imported_schema_ids", "imported_algorithm_ids"} {
			values, ok := stringSlice(row[key])
			if !ok || !strictlySortedUnique(values) {
				return registryInvalid("dependency_imports_invalid")
			}
			for _, value := range values {
				if !asciiIdentity(value, 160) {
					return registryInvalid("dependency_import_identity_invalid")
				}
				if key == "imported_schema_ids" {
					schemas[value] = true
				}
			}
		}
		artifacts, ok := objectSlice(row["imported_artifacts"])
		if !ok {
			return registryInvalid("dependency_artifacts_not_array")
		}
		previous := ""
		for _, artifact := range artifacts {
			id := stringValue(artifact["artifact_id"])
			if requireExactKeys(artifact, "artifact_id", "schema_id", "artifact_sha256", "safe_ref") != nil || !asciiIdentity(id, 160) || id <= previous || !schemas[stringValue(artifact["schema_id"])] || artifact["safe_ref"] != "artifact:"+id || !lowerDigest(stringValue(artifact["artifact_sha256"])) {
				return registryInvalid("dependency_artifact_invalid")
			}
			previous = id
		}
	}
	return nil
}

func validateOwnerFactShape(fact map[string]any) error {
	fields := map[string][]string{
		"recognized_profile": {"claimable", "contract_major", "primary_owner_id"}, "runtime_dependency": {"dependency_profile_id", "required_contract_major"},
		"route_family": {"route_family"}, "workspace": {"workspace_key"}, "public_schema": {"public_schema_id"},
		"claim_configuration": {"claim_config_key", "configuration_contract_sha256"}, "state_ownership": {"state_ownership"}, "admission_validation": {"admission_validation"},
		"egress": {"egress_mode"}, "portability": {"incident_portability_mode", "participant_id", "blocking_predicate"}, "snapshot_reporting": {"snapshot_reporting_mode", "participant_id"},
		"contribution": {"contribution"}, "migration_definition": {"migration_definition"}, "migration_ledger_definition": {"migration_ledger_definition"},
		"worker_kind": {"worker_kind", "job_kinds", "max_active_attempts_per_process"}, "job_kind": {"job_kind_contract"},
	}
	keys, ok := fields[stringValue(fact["fact_kind"])]
	if !ok {
		return registryInvalid("owner_fact_kind_invalid")
	}
	if err := requireExactKeys(fact, append([]string{"fact_kind", "profile_id"}, keys...)...); err != nil {
		return err
	}
	switch fact["fact_kind"] {
	case "recognized_profile":
		claimable, ok := fact["claimable"].(bool)
		major, majorOK := integerValue(fact["contract_major"])
		if !ok || !extensionProfileIDPattern.MatchString(stringValue(fact["primary_owner_id"])) || (fact["contract_major"] == nil && claimable) || (fact["contract_major"] != nil && (!majorOK || major < 1 || major > 2147483647)) {
			return registryInvalid("recognition_scalar_invalid")
		}
	case "runtime_dependency":
		major, ok := integerValue(fact["required_contract_major"])
		if !ok || major < 1 || major > 2147483647 || !extensionProfileIDPattern.MatchString(stringValue(fact["dependency_profile_id"])) {
			return registryInvalid("dependency_scalar_invalid")
		}
	case "claim_configuration":
		if fact["claim_config_key"] != stringValue(fact["profile_id"])+".claimed" || !lowerDigest(stringValue(fact["configuration_contract_sha256"])) {
			return registryInvalid("claim_fact_invalid")
		}
	}
	return nil
}
