package extensions

import "strings"

// validateCompanionIdentities checks the joins between independent packaged
// projections. Each contract's parser owns its own behavioral validation.
func (p *admittedPackage) validateCompanionIdentities() error {
	profiles, _ := objectSlice(p.objects[generatedExtensionsRoot+"profile-registry.json"]["profiles"])
	if err := p.validateClientIdentities(profiles); err != nil {
		return err
	}
	versioned := map[string]map[string]any{}
	participants := map[string]map[string]any{}
	participantReferences := map[string]map[string]any{}
	boundParticipants := map[string]map[string]any{}
	jobs, workers, err := p.declaredJobAndWorkerContracts(profiles)
	if err != nil {
		return err
	}
	boundCatalogPaths := map[string]bool{}
	for _, descriptor := range profiles {
		profile := stringValue(descriptor["profile_id"])
		state, _ := descriptor["state_ownership"].(map[string]any)
		if state["kind"] == "extension_versioned" {
			versioned[profile] = descriptor
		}
		binding := p.objects[generatedExtensionsRoot+"implementation-bindings/"+profile+".json"]
		participantBindings := []map[string]any{}
		if binding != nil {
			var ok bool
			participantBindings, ok = objectSlice(binding["participant_contracts"])
			if !ok {
				return registryInvalid("participant_bindings_not_array")
			}
		}
		previous := ""
		for _, row := range participantBindings {
			id := stringValue(row["participant_id"])
			if id == "" || id <= previous || boundParticipants[id] != nil {
				return registryInvalid("participant_binding_identity_invalid")
			}
			previous = id
			boundParticipants[id] = map[string]any{"profile_id": profile, "digest": row["participant_contract_sha256"]}
		}
		contributions, ok := objectSlice(descriptor["contributions"])
		if !ok {
			return registryInvalid("contributions_not_array")
		}
		for _, contribution := range contributions {
			id := stringValue(contribution["participant_id"])
			if id == "" {
				continue
			}
			if participants[id] != nil {
				return registryInvalid("participant_identity_duplicate")
			}
			participants[id] = map[string]any{"profile_id": profile, "digest": contribution["participant_contract_sha256"]}
			participantReferences[id] = contribution
		}
		for _, set := range []struct {
			field, identity, directory string
			expected                   map[string]map[string]any
		}{
			{"job_kind_contracts", "job_kind", "job-contracts", jobs}, {"worker_runtime_contracts", "worker_kind", "worker-runtime-contracts", workers},
		} {
			if binding == nil {
				continue
			}
			rows, ok := objectSlice(binding[set.field])
			if !ok {
				return registryInvalid("binding_contract_set_invalid")
			}
			for _, row := range rows {
				id := stringValue(row[set.identity])
				path := generatedExtensionsRoot + set.directory + "/" + profile + "/" + id + ".json"
				matches := equalCanonicalObjects(row, p.objects[path])
				if set.field == "job_kind_contracts" {
					matches = requireExactKeys(row, "job_kind", "job_kind_contract_sha256") == nil && row["job_kind_contract_sha256"] == p.artifacts[path].SHA256
				}
				if set.expected[path] == nil || boundCatalogPaths[path] || !matches {
					return registryInvalid("binding_contract_projection_mismatch")
				}
				boundCatalogPaths[path] = true
			}
		}
	}
	for id, expected := range participants {
		profile := stringValue(expected["profile_id"])
		if p.objects[generatedExtensionsRoot+"implementation-bindings/"+profile+".json"] != nil && !equalCanonicalObjects(expected, boundParticipants[id]) {
			return registryInvalid("participant_binding_projection_mismatch")
		}
	}
	for id := range boundParticipants {
		if participants[id] == nil {
			return registryInvalid("participant_binding_set_mismatch")
		}
	}
	for _, path := range p.paths {
		job, worker := jobs[path], workers[path]
		isJob := strings.HasPrefix(path, generatedExtensionsRoot+"job-contracts/")
		isWorker := strings.HasPrefix(path, generatedExtensionsRoot+"worker-runtime-contracts/")
		if isJob && job == nil || isWorker && worker == nil {
			return registryInvalid("unbound_job_or_worker_contract")
		}
		if isJob || isWorker {
			profile := stringValue(p.objects[path]["profile_id"])
			if p.objects[generatedExtensionsRoot+"implementation-bindings/"+profile+".json"] != nil && !boundCatalogPaths[path] {
				return registryInvalid("binding_contract_set_incomplete")
			}
		}
	}
	stateRows, err := p.companionRows("state-registry", "cartulary.extension_state_registry.v1", "profiles", "profile_id")
	if err != nil {
		return err
	}
	backupRows, err := p.companionRows("backup-registry", "cartulary.extension_backup_registry.v1", "profiles", "profile_id")
	if err != nil {
		return err
	}
	if len(stateRows) != len(versioned) || len(backupRows) != len(versioned) {
		return registryInvalid("versioned_profile_set_mismatch")
	}
	for _, row := range stateRows {
		profile := stringValue(row["profile_id"])
		descriptor := versioned[profile]
		if descriptor == nil {
			return registryInvalid("state_profile_unrecognized")
		}
		binding := p.objects[generatedExtensionsRoot+"implementation-bindings/"+profile+".json"]
		state, _ := descriptor["state_ownership"].(map[string]any)
		var bindingDigest any
		if binding != nil {
			bindingDigest = canonicalDigest(binding)
		}
		for _, key := range []string{"migration_lineage_id", "current_state_version", "minimum_migratable_state_version", "empty_state_policy", "initialization_definition_sha256", "final_state_validation_algorithm_id"} {
			if canonicalDigest(row[key]) != canonicalDigest(state[key]) {
				return registryInvalid("state_owner_projection_mismatch")
			}
		}
		if canonicalDigest(row["contract_major"]) != canonicalDigest(descriptor["contract_major"]) || row["implementation_binding_sha256"] != bindingDigest || binding != nil && canonicalDigest(row["physical_state_binding_sha256"]) != canonicalDigest(binding["physical_state_binding_sha256"]) {
			return registryInvalid("state_binding_projection_mismatch")
		}
		presence := p.objects["contracts/extensions/profiles/"+profile+"/state-presence.json"]
		initialization := p.objects["contracts/extensions/profiles/"+profile+"/state-initialization.json"]
		if presence == nil || initialization == nil || row["state_presence_manifest_sha256"] != canonicalDigest(presence) || row["initialization_definition_sha256"] != canonicalDigest(initialization) {
			return registryInvalid("state_definition_digest_mismatch")
		}
		if !strings.HasPrefix(stringValue(presence["state_presence_manifest_id"]), profile+".") {
			return registryInvalid("state_presence_identity_mismatch")
		}
		for _, key := range []string{"database_family_ids", "object_reference_family_ids", "empty_state_policy", "migration_lineage_id", "profile_id"} {
			if canonicalDigest(row[key]) != canonicalDigest(presence[key]) {
				return registryInvalid("state_presence_projection_mismatch")
			}
		}
		initial, _ := initialization["initialization"].(map[string]any)
		if canonicalDigest(row["initialization_kind"]) != canonicalDigest(initial["kind"]) || canonicalDigest(row["initialization_algorithm_id"]) != canonicalDigest(initial["algorithm_id"]) || canonicalDigest(row["initialization_algorithm_definition_sha256"]) != canonicalDigest(initial["algorithm_definition_sha256"]) {
			return registryInvalid("state_initialization_projection_mismatch")
		}
		migrations, ok := objectSlice(row["migration_definitions"])
		declared := p.declaredMigrationBindings(profile)
		if !ok || len(migrations) != len(declared) {
			return registryInvalid("migration_binding_set_mismatch")
		}
		if binding != nil {
			bound, bindingsOK := objectSlice(binding["migration_definitions"])
			if !bindingsOK || canonicalDigest(bound) != canonicalDigest(declared) {
				return registryInvalid("migration_owner_projection_mismatch")
			}
		}
		for i, migration := range migrations {
			expected := cloneObject(declared[i])
			expected["migration_lineage_id"] = state["migration_lineage_id"]
			expected["implementation_binding_profile_id"] = profile
			expected["implementation_binding_sha256"] = bindingDigest
			if !equalCanonicalObjects(migration, expected) {
				return registryInvalid("migration_binding_projection_mismatch")
			}
		}
	}
	// Authored contracts are independently packaged; an embedded copy must have
	// exactly the same bytes, not only a self-consistent wrapper digest.
	authored := map[string]map[string]any{}
	for _, path := range p.paths {
		if !strings.HasPrefix(path, generatedExtensionsRoot) {
			authored[p.artifacts[path].SHA256] = p.objects[path]
		}
	}
	for _, row := range backupRows {
		profile := stringValue(row["profile_id"])
		if versioned[profile] == nil {
			return registryInvalid("backup_profile_unrecognized")
		}
		physical, _ := row["physical_state_binding"].(map[string]any)
		digest := stringValue(row["physical_state_binding_sha256"])
		binding := p.objects[generatedExtensionsRoot+"implementation-bindings/"+profile+".json"]
		if physical == nil || physical["profile_id"] != profile || canonicalDigest(physical) != digest || !equalCanonicalObjects(physical, authored[digest]) || binding != nil && binding["physical_state_binding_sha256"] != digest {
			return registryInvalid("physical_binding_projection_mismatch")
		}
		for _, set := range []struct{ rows, body, digest string }{{"codecs", "codec", "backup_codec_sha256"}, {"logical_contracts", "contract", "codec_contract_sha256"}, {"postgres_mappings", "mapping", "mapping_sha256"}} {
			entries, ok := objectSlice(row[set.rows])
			if !ok {
				return registryInvalid("backup_contract_set_invalid")
			}
			for _, entry := range entries {
				body, ok := entry[set.body].(map[string]any)
				sha := stringValue(entry[set.digest])
				if !ok || canonicalDigest(body) != sha || !equalCanonicalObjects(body, authored[sha]) {
					return registryInvalid("backup_contract_projection_mismatch")
				}
			}
		}
	}
	participantRows, err := p.companionRows("participant-registry", "cartulary.extension_participant_registry.v1", "participants", "participant_id")
	if err != nil {
		return err
	}
	if len(participantRows) != len(participants) {
		return registryInvalid("participant_set_mismatch")
	}
	for _, row := range participantRows {
		id := stringValue(row["participant_id"])
		expected := participants[id]
		contract, ok := row["contract"].(map[string]any)
		digest := stringValue(row["participant_contract_sha256"])
		ownerKey := "profile_id"
		if contract["schema_id"] == "cartulary.extension_transaction_participant_contract.v3" {
			ownerKey = "owner_profile_id"
		}
		if !ok || expected == nil || row["profile_id"] != expected["profile_id"] || digest != expected["digest"] || canonicalDigest(contract) != digest || !equalCanonicalObjects(contract, authored[digest]) || contract["participant_id"] != id || canonicalDigest(contract[ownerKey]) != canonicalDigest(row["profile_id"]) || canonicalDigest(contract["schema_id"]) != canonicalDigest(row["contract_schema_id"]) {
			return registryInvalid("participant_projection_mismatch")
		}
		reference := participantReferences[id]
		expectedSchema := "cartulary.extension_participant_specialization.v5"
		kind := stringValue(reference["kind"])
		switch kind {
		case "cross_owner_transaction_participant":
			expectedSchema = "cartulary.extension_transaction_participant_contract.v3"
		case "snapshot_reporting_participant", "incident_portability_participant":
			if kind == "snapshot_reporting_participant" {
				expectedSchema = "cartulary.extension_reporting_source_contract.v1"
			}
			if contract["participant_kind"] != strings.TrimSuffix(kind, "_participant") {
				return registryInvalid("participant_specialization_mismatch")
			}
		default:
			return registryInvalid("participant_contribution_kind_invalid")
		}
		if reference["participant_contract_ref"] != expectedSchema || contract["schema_id"] != expectedSchema {
			return registryInvalid("participant_contract_reference_invalid")
		}
	}
	return nil
}

func (p *admittedPackage) companionRows(name, schema, field, identity string) ([]map[string]any, error) {
	object := p.objects[generatedExtensionsRoot+name+".json"]
	if requireExactKeys(object, "schema_id", field) != nil || object["schema_id"] != schema {
		return nil, registryInvalid("companion_registry_shape_invalid")
	}
	rows, ok := objectSlice(object[field])
	if !ok {
		return nil, registryInvalid("companion_rows_not_array")
	}
	previous := ""
	for _, row := range rows {
		id := stringValue(row[identity])
		if id == "" || id <= previous {
			return nil, registryInvalid("companion_identity_duplicate_or_unsorted")
		}
		previous = id
	}
	return rows, nil
}

func (p *admittedPackage) validateClientIdentities(profiles []map[string]any) error {
	descriptors := map[string]map[string]any{}
	for _, descriptor := range profiles {
		descriptors[stringValue(descriptor["profile_id"])] = descriptor
	}
	rows, ok := objectSlice(p.objects[generatedExtensionsRoot+"client-support-registry.json"]["profiles"])
	if !ok || len(rows) > 256 {
		return registryInvalid("client_profiles_invalid")
	}
	seen := map[string]bool{}
	previous := ""
	for _, row := range rows {
		id := stringValue(row["profile_id"])
		descriptor := descriptors[id]
		if requireExactKeys(row, "profile_id", "supported_contract_majors", "workspace_keys", "capability_ids", "public_schema_ids") != nil || id <= previous || descriptor == nil || descriptor["claimable"] != true {
			return registryInvalid("client_profile_unrecognized_or_invalid")
		}
		previous = id
		seen[id] = true
		if canonicalDigest(row["supported_contract_majors"]) != canonicalDigest([]any{descriptor["contract_major"]}) {
			return registryInvalid("client_profile_major_mismatch")
		}
		for _, set := range []struct {
			field   string
			maximum int
		}{{"workspace_keys", 64}, {"capability_ids", 0}, {"public_schema_ids", 256}} {
			values, ok := stringSlice(row[set.field])
			if !ok || len(values) > set.maximum || !strictlySortedUnique(values) {
				return registryInvalid("client_profile_members_invalid")
			}
			allowed, _ := stringSlice(descriptor[set.field])
			membership := map[string]bool{}
			for _, value := range allowed {
				membership[value] = true
			}
			for _, value := range values {
				if !membership[value] {
					return registryInvalid("client_profile_member_unrecognized")
				}
			}
			if set.field == "workspace_keys" && len(values) != len(allowed) {
				return registryInvalid("client_profile_workspace_mismatch")
			}
		}
	}
	for id, descriptor := range descriptors {
		workspaces, _ := stringSlice(descriptor["workspace_keys"])
		if descriptor["claimable"] == true && len(workspaces) > 0 && !seen[id] {
			return registryInvalid("client_required_profile_missing")
		}
	}
	return nil
}
