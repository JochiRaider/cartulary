package extensions

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"reflect"
	"sort"
	"strings"
	"testing"

	contractsgen "github.com/JochiRaider/cartulary/internal/gen/contractextensions"
)

func TestCoordinatorGeneratedRegistry_Unit(t *testing.T) {
	t.Run("owner input limit admission", checkOwnerInputLimitAdmission)
	t.Run("owner input findings", checkOwnerInputDiagnostics)
	t.Run("owner fact identity and normalization", checkOwnerIdentityContract)
	t.Run("startup diagnostic projection", verifyStartupDiagnosticProjection)
	t.Run("package closure", checkRegistryPackageClosure)
	t.Run("empty registry", checkEmptyRegistry)
	t.Run("recognized profiles without implementations", checkUnavailableProfiles)
	t.Run("retired backup participant", checkRetiredBackupParticipant)
	t.Run("typed binding contract references", checkJobBindingReferences)
	t.Run("job policy variants", checkJobPolicyVariants)
	t.Run("binding diagnostics", checkMalformedBindingContracts)
	t.Run("client support joins", checkClientSupportJoins)
	t.Run("final browser package", checkFinalBrowserPackage)
	t.Run("typed Reporting source contracts", checkReportingSourceContracts)
	t.Run("participant contribution references", checkParticipantContributionReferences)
	t.Run("nested metadata contracts", checkNestedMetadataContracts)
	t.Run("participant schema and algorithm bindings", checkParticipantSchemaBindings)
	t.Run("schema resource confinement", checkSchemaResourceConfinement)
	t.Run("current Network Flow state contract", func(t *testing.T) {
		coordinator := requireGeneratedCoordinator(t)
		plan, ok := coordinator.StatePlan("network_flow_activity")
		if !ok || plan.CurrentStateVersion != 7 || plan.MinimumMigratableStateVersion != 7 || len(plan.MigrationDefinitions) != 0 || len(plan.MigrationLedgerDefinitions) != 0 {
			t.Fatal("current state-6-only catalog retains unsupported migration declarations")
		}
	})

	t.Run("complete integrity contract", func(t *testing.T) {
		for _, name := range []string{"unknown field", "missing algorithm", "wrong algorithm", "wrong registry schema", "missing dependency digest", "wrong owner input digest", "missing fragment", "missing schema", "missing supporting contract", "source traversal", "source digest", "source order", "descriptor order"} {
			t.Run(name, func(t *testing.T) {
				source := newMutableGeneratedSource()
				path := generatedExtensionsRoot + "registry-integrity.json"
				root := decodeMutableObject(t, source[path].JSON)
				switch name {
				case "unknown field":
					root["unknown"] = true
				case "missing algorithm":
					delete(root, "canonicalization_algorithm_id")
				case "wrong algorithm":
					root["canonicalization_algorithm_id"] = "unsupported"
				case "wrong registry schema":
					root["registry_schema_id"] = "retired"
				case "missing dependency digest":
					delete(root, "dependency_snapshot_sha256")
				case "wrong owner input digest":
					root["owner_input_registry_sha256"] = "incorrect"
				case "missing fragment":
					root["owner_fragment_digests"] = root["owner_fragment_digests"].([]any)[1:]
				case "missing schema":
					root["generated_schema_digests"] = root["generated_schema_digests"].([]any)[1:]
				case "missing supporting contract":
					rows := root["supporting_contract_artifact_digests"].([]any)
					root["supporting_contract_artifact_digests"] = rows[:len(rows)-1]
				case "source traversal":
					root["generator_sources"].([]any)[0].(map[string]any)["source_ref"] = "../outside.go"
				case "source digest":
					root["generator_sources"].([]any)[0].(map[string]any)["source_sha256"] = "incorrect"
				case "source order":
					rows := root["generator_sources"].([]any)
					rows[0], rows[1] = rows[1], rows[0]
				case "descriptor order":
					rows := root["descriptor_digests"].([]any)
					rows[0], rows[1] = rows[1], rows[0]
				}
				source.replace(t, path, root)
				_, err := NewCoordinator(source)
				requireFinding(t, err, "extension_registry_invalid", "registry_generation")
			})
		}
	})
	t.Run("packaged logical codec definitions", func(t *testing.T) {
		source := newMutableGeneratedSource()
		artifact, present := source.Artifact("contracts/extensions/profiles/network_flow_activity/postgres-rows.json")
		if !present {
			t.Fatal("logical PostgreSQL row contract is not packaged")
		}
		definition := decodeMutableObject(t, artifact.JSON)
		families, ok := definition["families"].([]any)
		if !ok || len(families) != 5 || definition["schema_id"] != "cartulary.network_flow_activity.postgres_rows.v3" {
			t.Fatal("logical PostgreSQL row contract is incomplete")
		}
	})
	t.Run("unresolved logical codec reference", func(t *testing.T) {
		source := newMutableGeneratedSource()
		registry := decodeMutableObject(t, source[generatedExtensionsRoot+"backup-registry.json"].JSON)
		row := registry["profiles"].([]any)[0].(map[string]any)
		encoded := row["codecs"].([]any)[0].(map[string]any)
		body := encoded["codec"].(map[string]any)
		body["codec_contract_ref"] = "cartulary.unresolved.logical_rows.v1"
		encoded["backup_codec_sha256"] = canonicalDigest(body)
		physical := row["physical_state_binding"].(map[string]any)
		for _, raw := range physical["bindings"].([]any) {
			binding := raw.(map[string]any)
			if binding["backup_codec_id"] == encoded["backup_codec_id"] {
				binding["backup_codec_sha256"] = encoded["backup_codec_sha256"]
			}
		}
		if _, err := parseBackupPlan(row); err == nil {
			t.Fatal("unresolved logical codec reference admitted")
		}
	})
	t.Run("logical definition and mapping rejection", func(t *testing.T) {
		for _, name := range []string{"missing definition", "missing mapping", "extra mapping", "mapping digest", "mapping selector", "mapping unknown field", "definition missing nullable", "duplicate identity field"} {
			t.Run(name, func(t *testing.T) {
				source := newMutableGeneratedSource()
				registry := decodeMutableObject(t, source[generatedExtensionsRoot+"backup-registry.json"].JSON)
				row := registry["profiles"].([]any)[0].(map[string]any)
				definition := row["logical_contracts"].([]any)[0].(map[string]any)
				mapping := row["postgres_mappings"].([]any)[0].(map[string]any)
				switch name {
				case "missing definition":
					delete(row, "logical_contracts")
				case "missing mapping":
					row["postgres_mappings"] = row["postgres_mappings"].([]any)[1:]
				case "extra mapping":
					row["postgres_mappings"] = append(row["postgres_mappings"].([]any), mapping)
				case "mapping digest":
					mapping["mapping_sha256"] = "incorrect"
				case "mapping selector":
					body := mapping["mapping"].(map[string]any)
					body["table_name"] = "other_table"
					mapping["mapping_sha256"] = canonicalDigest(body)
				case "mapping unknown field":
					body := mapping["mapping"].(map[string]any)
					body["sql"] = "untrusted expression"
					mapping["mapping_sha256"] = canonicalDigest(body)
				case "definition missing nullable", "duplicate identity field":
					body := definition["contract"].(map[string]any)
					family := body["families"].([]any)[0].(map[string]any)
					if name == "definition missing nullable" {
						delete(family["fields"].([]any)[0].(map[string]any), "nullable")
					} else {
						identity := family["identity_field_ids"].([]any)[0]
						family["identity_field_ids"] = []any{identity, identity}
					}
					definition["codec_contract_sha256"] = canonicalDigest(body)
				}
				if _, err := parseBackupPlan(row); err == nil {
					t.Fatal("invalid logical codec admitted")
				}
			})
		}
	})
	t.Run("canonical bytes", func(t *testing.T) {
		value := map[string]any{"text": "<>&/\u2028\u2029\x01", "\ue000": json.Number("9223372036854775807"), "😀": json.Number("-9223372036854775808")}
		want := "{\"text\":\"<>&/\u2028\u2029\\u0001\",\"\ue000\":9223372036854775807,\"😀\":-9223372036854775808}\n"
		got, err := canonicalJSON(value, true)
		if err != nil || string(got) != want {
			t.Errorf("canonical bytes = %q, error %v; want %q", got, err, want)
		}
		for _, invalid := range []any{json.Number("1.5"), json.Number("1e0"), json.Number("-0"), "\xff"} {
			if _, err := canonicalJSON(invalid, true); err == nil {
				t.Errorf("invalid canonical input %T accepted", invalid)
			}
		}
	})
	t.Run("strict original artifact bytes", func(t *testing.T) {
		for name, mutate := range map[string]func([]byte) []byte{
			"missing LF":             func(v []byte) []byte { return bytes.TrimSuffix(v, []byte("\n")) },
			"extra LF":               func(v []byte) []byte { return append(v, '\n') },
			"BOM":                    func(v []byte) []byte { return append([]byte{0xef, 0xbb, 0xbf}, v...) },
			"trailing invalid bytes": func(v []byte) []byte { return append(v, '!') },
			"leading whitespace":     func(v []byte) []byte { return append([]byte(" "), v...) },
			"duplicate member": func(v []byte) []byte {
				return bytes.Replace(v, []byte(`"schema_id":`), []byte(`"schema_id":"ignored","schema_id":`), 1)
			},
			"noncanonical escape": func(v []byte) []byte { return bytes.ReplaceAll(v, []byte("cartulary"), []byte(`\u0063artulary`)) },
		} {
			t.Run(name, func(t *testing.T) {
				source := newMutableGeneratedSource()
				path := generatedExtensionsRoot + "registry-integrity.json"
				artifact := source[path]
				artifact.JSON = mutate(artifact.JSON)
				digest := sha256.Sum256(artifact.JSON)
				artifact.SHA256 = hex.EncodeToString(digest[:])
				source[path] = artifact
				_, err := NewCoordinator(source)
				requireFinding(t, err, "extension_registry_invalid", "registry_generation")
			})
		}
	})
	coordinator, err := NewGeneratedCoordinator()
	if err != nil {
		var validation *ValidationError
		if errors.As(err, &validation) {
			t.Fatalf("admit generated registry: %#v", validation.Findings())
		}
		t.Fatalf("admit generated registry: %v", err)
	}
	graphAdmission := coordinator.profiles["network_flow_activity"].bindingObject
	if graphAdmission["preflight_algorithm_id"] != "network_flow_activity.incident_portability_cutover_v1" || graphAdmission["post_migration_algorithm_id"] != "network_flow_activity.validate_state_v7" {
		t.Fatal("packaged admission algorithms were silently omitted")
	}
	descriptors := coordinator.Descriptors()
	if len(descriptors) != 6 || len(coordinator.RegistrySHA256()) != 64 {
		t.Fatalf("generated registry identity = %d/%q", len(descriptors), coordinator.RegistrySHA256())
	}
	if descriptors[0].ProfileID != "enterprise_authentication" || descriptors[5].ProfileID != "snapshot_reporting" {
		t.Fatalf("descriptors are not canonical: %#v", descriptors)
	}
	inactivePolicies := coordinator.InactiveConfigurationPolicies()
	if len(inactivePolicies) != 6 ||
		inactivePolicies[0].Key != "enterprise_authentication.provider_manifest_path" ||
		inactivePolicies[0].Kind != "forbidden" ||
		inactivePolicies[1].Key != "network_flow_activity.key_ring_manifest_path" ||
		inactivePolicies[1].Kind != "forbidden" ||
		inactivePolicies[2].Key != "network_flow_activity.resource_limits" ||
		inactivePolicies[2].Kind != "forbidden" ||
		inactivePolicies[3].Key != "reference_packs.clock_trusted" ||
		inactivePolicies[3].Kind != "forbidden" ||
		inactivePolicies[4].Key != "reference_packs.trust_bootstrap_path" ||
		inactivePolicies[4].Kind != "forbidden" ||
		inactivePolicies[5].ProfileID != "snapshot_reporting" ||
		inactivePolicies[5].Key != "snapshot_reporting.runtime_receipt_path" ||
		inactivePolicies[5].Kind != "forbidden" {
		t.Fatalf("inactive configuration policies = %#v", inactivePolicies)
	}
	descriptors[0].RouteFamilies[0] = "mutated"
	again, _ := coordinator.Descriptor("enterprise_authentication")
	if again.RouteFamilies[0] == "mutated" {
		t.Fatal("descriptor query leaked mutable coordinator state")
	}
	t.Run("current-only backup codec contracts", func(t *testing.T) {
		source := newMutableGeneratedSource()
		registry := decodeMutableObject(t, source[generatedExtensionsRoot+"backup-registry.json"].JSON)
		count := 0
		for _, rawProfile := range registry["profiles"].([]any) {
			for _, rawCodec := range rawProfile.(map[string]any)["codecs"].([]any) {
				body := rawCodec.(map[string]any)["codec"].(map[string]any)
				if body["schema_id"] != "cartulary.extension_backup_binding_codec.v4" {
					t.Errorf("admitted codec %v uses %v", body["backup_codec_id"], body["schema_id"])
				}
				if _, present := body["historical_restore_codecs"]; present {
					t.Errorf("admitted codec %v retains historical support", body["backup_codec_id"])
				}
				if body["codec_contract_ref"] != "cartulary.network_flow_activity.postgres_rows.v3" {
					t.Errorf("admitted codec %v omits its logical contract", body["backup_codec_id"])
				}
				count++
			}
		}
		if count != 5 {
			t.Fatalf("current backup codec set has %d entries; want five", count)
		}
	})
}

func TestExtensionProfileAdoptionMatrix_Static(t *testing.T) {
	coordinator := requireGeneratedCoordinator(t)
	descriptors := coordinator.Descriptors()
	byProfile := make(map[string]Descriptor, len(descriptors))
	for _, descriptor := range descriptors {
		byProfile[descriptor.ProfileID] = descriptor
	}

	type jobIdentity struct {
		jobKind       string
		operationKind string
		workerKind    string
	}
	contractMajors := map[string]int{
		"enterprise_authentication": 1,
		"import":                    1,
		"incident_portability":      1,
		"network_flow_activity":     10,
		"reference_pack":            4,
		"snapshot_reporting":        1,
	}
	adoption := map[string][]jobIdentity{
		"enterprise_authentication": nil,
		"import": {
			{jobKind: "import.discovery_v1", operationKind: "import.discovery", workerKind: "import.discovery_worker_v1"},
			{jobKind: "import.apply_v1", operationKind: "import.apply", workerKind: "import.apply_worker_v1"},
		},
		"incident_portability": {
			{jobKind: "incident_portability.export_v2", operationKind: "incident_portability.export", workerKind: "incident_portability.bundle_worker_v2"},
			{jobKind: "incident_portability.import_v2", operationKind: "incident_portability.import", workerKind: "incident_portability.bundle_worker_v2"},
		},
		"network_flow_activity": {
			{jobKind: "network_flow_activity.graph_view_materialize_v1", operationKind: "network_flow_activity.graph_view_materialize", workerKind: "network_flow_activity.graph_view_worker_v1"},
		},
		"reference_pack": {
			{jobKind: "reference_pack.import_v2", operationKind: "reference_pack.import", workerKind: "reference_pack.lifecycle_worker_v2"},
			{jobKind: "reference_pack.reverify_v2", operationKind: "reference_pack.reverify", workerKind: "reference_pack.lifecycle_worker_v2"},
			{jobKind: "reference_pack.refresh_v2", operationKind: "reference_pack.refresh", workerKind: "reference_pack.lifecycle_worker_v2"},
		},
		"snapshot_reporting": {
			{jobKind: "snapshot_reporting.snapshot_create_v4", operationKind: "snapshot_reporting.snapshot_create", workerKind: "snapshot_reporting.job_worker_v2"},
			{jobKind: "snapshot_reporting.release_create_v4", operationKind: "snapshot_reporting.release_create", workerKind: "snapshot_reporting.job_worker_v2"},
			{jobKind: "snapshot_reporting.composition_preview_v4", operationKind: "snapshot_reporting.composition_preview", workerKind: "snapshot_reporting.job_worker_v2"},
		},
	}
	if len(adoption) != 6 {
		t.Fatalf("profile adoption matrix has %d profiles; want 6", len(adoption))
	}
	jobKinds := map[string]string{}
	workerKinds := map[string]bool{}
	for profileID, jobs := range adoption {
		descriptor, ok := byProfile[profileID]
		if !ok || !descriptor.Claimable || descriptor.ContractMajor != contractMajors[profileID] {
			t.Fatalf("canonical profile %q = %#v/%t", profileID, descriptor, ok)
		}
		for _, job := range jobs {
			if previous, duplicate := jobKinds[job.jobKind]; duplicate {
				t.Fatalf("job kind %q is shared by %q and %q", job.jobKind, previous, profileID)
			}
			jobKinds[job.jobKind] = profileID
			if job.operationKind == "" || job.workerKind == "" {
				t.Fatalf("incomplete adoption identity for %q: %#v", profileID, job)
			}
			workerKinds[job.workerKind] = true
		}
	}
	if len(jobKinds) != 11 || len(workerKinds) != 6 {
		t.Fatalf("adoption identity totals = %d jobs/%d workers; want 11/6", len(jobKinds), len(workerKinds))
	}
	progressUnits := map[string]string{
		"import.discovery_v1":                             "import.discovery.session.v1",
		"import.apply_v1":                                 "import.apply.import_unit.v1",
		"incident_portability.export_v2":                  "incident_portability.export.request.v2",
		"incident_portability.import_v2":                  "incident_portability.import.request.v2",
		"network_flow_activity.graph_view_materialize_v1": "network_flow_activity.graph_view_materialize.projection_result.v1",
		"reference_pack.import_v2":                        "reference_pack.import.request.v2",
		"reference_pack.refresh_v2":                       "reference_pack.refresh.cohort_member.v2",
		"reference_pack.reverify_v2":                      "reference_pack.reverify.pack_version.v2",
		"snapshot_reporting.composition_preview_v4":       "snapshot_reporting.composition_preview.render_attempt.v1",
		"snapshot_reporting.release_create_v4":            "snapshot_reporting.release_create.render_attempt.v1",
		"snapshot_reporting.snapshot_create_v4":           "snapshot_reporting.snapshot_create.materialization.v1",
	}
	liveJobs := coordinator.JobKindContracts()
	if len(liveJobs) != len(jobKinds) {
		t.Fatalf("generated live job catalog = %d; want %d", len(liveJobs), len(jobKinds))
	}
	for _, liveJob := range liveJobs {
		expectedProfile, ok := jobKinds[liveJob.JobKind]
		if !ok || expectedProfile != liveJob.ProfileID {
			t.Fatalf("unexpected live job contract %#v", liveJob)
		}
		foundOperation := false
		for _, expectedJob := range adoption[liveJob.ProfileID] {
			if expectedJob.jobKind == liveJob.JobKind && expectedJob.operationKind == liveJob.OperationKind {
				foundOperation = true
				break
			}
		}
		if !foundOperation ||
			liveJob.ProgressUnitID != progressUnits[liveJob.JobKind] ||
			liveJob.ProofPolicy != "required_on_terminal_success" ||
			liveJob.IdempotencyPolicy != "required" ||
			liveJob.CancellationPolicy != "precommit_observable" {
			t.Fatalf("live job contract does not match adoption facts: %#v", liveJob)
		}
	}
	allClaims, err := coordinator.ResolveClaims([]string{
		"enterprise_authentication",
		"import",
		"incident_portability",
		"network_flow_activity",
		"reference_pack",
		"snapshot_reporting",
	})
	if err != nil {
		t.Fatalf("resolve complete claim set: %v", err)
	}
	plan, err := coordinator.BuildPublicationPlan(allClaims)
	if err != nil {
		t.Fatalf("build complete publication plan: %v", err)
	}
	liveWorkers := plan.Workers()
	if len(liveWorkers) != len(workerKinds) {
		t.Fatalf("generated live worker catalog = %d; want %d", len(liveWorkers), len(workerKinds))
	}
	for _, liveWorker := range liveWorkers {
		if !workerKinds[liveWorker.WorkerKind] {
			t.Fatalf("unexpected live worker %#v", liveWorker)
		}
		wantMaximum := 8
		if liveWorker.WorkerKind == "network_flow_activity.graph_view_worker_v1" {
			wantMaximum = 1
		}
		if len(liveWorker.JobKinds) == 0 || liveWorker.MaxActiveAttemptsPerProcess != wantMaximum {
			t.Fatalf("invalid live worker runtime contract %#v", liveWorker)
		}
	}
	if networkFlow, ok := byProfile["network_flow_activity"]; !ok || !networkFlow.Claimable || networkFlow.ContractMajor != 10 {
		t.Fatalf("Network Flow v5 adopted profile = %#v/%t", networkFlow, ok)
	}

	var participant *ParticipantContract
	for _, contract := range coordinator.ParticipantContracts() {
		if contract.ParticipantID == "snapshot_reporting.render_export_v1" {
			t.Fatal("retired self-participant remains")
		}
		if contract.ParticipantID == "network_flow_activity.reporting_graph_source_v2" {
			copy := contract
			participant = &copy
		}
	}
	if participant == nil || participant.OwnerProfileID != "network_flow_activity" || participant.ContractKind != "cartulary.extension_reporting_source_contract.v1" || participant.InputSchemaID != "" || len(participant.Operations) != 0 || participant.ReportingSource.MaxLabelBytes != 67108864 {
		t.Fatalf("current typed source contract = %#v", participant)
	}

}

func TestCoordinatorPortabilityPolicyProjection_Unit(t *testing.T) {
	coordinator := requireGeneratedCoordinator(t)
	policies := coordinator.PortabilityPolicies()
	if len(policies) != 6 {
		t.Fatalf("portability policies = %d; want 6", len(policies))
	}
	byProfile := make(map[string]PortabilityPolicy, len(policies))
	for _, policy := range policies {
		byProfile[policy.ProfileID] = policy
	}
	networkFlow := byProfile["network_flow_activity"]
	if networkFlow.Mode != PortabilityParticipant || networkFlow.ContractMajor != 10 ||
		networkFlow.ParticipantID != "network_flow_activity.incident_portability_v1" ||
		networkFlow.ParticipantSchemaID != "cartulary.extension_participant_specialization.v5" ||
		len(networkFlow.ParticipantSHA256) != 64 ||
		networkFlow.MaximumInputBytes != 64*1024*1024 || networkFlow.MaximumOutputBytes != 64*1024*1024 ||
		len(networkFlow.BlockingFamilyIDs) != 0 ||
		!reflect.DeepEqual(networkFlow.AuthoritativeFamilyIDs, []string{
			"network_flow_activity.graph_views",
			"network_flow_activity.indicator_bindings",
			"network_flow_activity.rejected_row_diagnostics",
			"network_flow_activity.rows",
			"network_flow_activity.tables",
		}) {
		t.Fatalf("Network Flow portability policy = %#v", networkFlow)
	}
	reporting := byProfile["snapshot_reporting"]
	if reporting.Mode != PortabilityNoAuthoritativeState || reporting.ParticipantID != "" {
		t.Fatalf("Snapshot/Reporting portability policy = %#v", reporting)
	}
	policies[3].AuthoritativeFamilyIDs[0] = "mutated"
	if coordinator.PortabilityPolicies()[3].AuthoritativeFamilyIDs[0] == "mutated" {
		t.Fatal("portability policy projection leaked mutable state")
	}
}

func TestCoordinatorClaimResolution_Unit(t *testing.T) {
	t.Run("bounded dependency phase", verifyDependencyFindingBound)
	verifyDependencyGraphs(t)
	coordinator := requireGeneratedCoordinator(t)
	resolution, err := coordinator.ResolveClaims([]string{"network_flow_activity", "reference_pack", "import"})
	if err != nil {
		t.Fatalf("resolve explicit dependency set: %v", err)
	}
	if got, want := resolution.Claims().ProfileIDs(), []string{"import", "network_flow_activity", "reference_pack"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("claim set = %v; want %v", got, want)
	}
	if got, want := resolution.AdmissionOrder(), []string{"import", "network_flow_activity", "reference_pack"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("admission order = %v; want %v", got, want)
	}
	_, err = coordinator.ResolveClaims([]string{"network_flow_activity"})
	requireFinding(t, err, "extension_dependency_not_claimed", "dependency_validation")
	_, err = coordinator.ResolveClaims([]string{"future_profile"})
	requireFinding(t, err, "extension_profile_unrecognized", "claim_configuration")
}

func TestCoordinatorCollisionAdmission_Unit(t *testing.T) {
	t.Run("unadmitted identities", checkUnadmittedCollisionValues)
	t.Run("complete stage collection", checkRegistryCollisionCollection)
	t.Run("declared namespace collision", checkDeclaredNamespaceCollision)
	source := newMutableGeneratedSource()
	basePath := "contracts/extensions/build/base-route-reservations.json"
	base := decodeMutableObject(t, source[basePath].JSON)
	reservations := base["reservations"].([]any)
	reservations = append(reservations, map[string]any{
		"reservation_id": "base.test.capture", "path_template": "/api/v1/import-sessions", "match_scope": "descendants",
	})
	base["reservations"] = reservations
	source.replace(t, basePath, base)
	integrityPath := generatedExtensionsRoot + "registry-integrity.json"
	integrity := decodeMutableObject(t, source[integrityPath].JSON)
	for _, raw := range integrity["supporting_contract_artifact_digests"].([]any) {
		row := raw.(map[string]any)
		if row["artifact_id"] == "build.base-route-reservations" {
			row["artifact_sha256"] = source[basePath].SHA256
		}
	}
	source.replace(t, integrityPath, integrity)
	_, err := NewCoordinator(source)
	requireCollision(t, err, "base_route_capture")
}

func TestCoordinatorBindingAdmission_Unit(t *testing.T) {
	for name, mutate := range map[string]func(map[string]any){
		"retired codec schema":      func(codec map[string]any) { codec["schema_id"] = "cartulary.extension_backup_binding_codec.v3" },
		"removed empty history":     func(codec map[string]any) { codec["historical_restore_codecs"] = []any{} },
		"removed null history":      func(codec map[string]any) { codec["historical_restore_codecs"] = nil },
		"unknown codec member":      func(codec map[string]any) { codec["fallback"] = false },
		"missing logical contract":  func(codec map[string]any) { delete(codec, "codec_contract_ref") },
		"corrupted packaged digest": func(map[string]any) {},
	} {
		t.Run(name, func(t *testing.T) {
			source := newMutableGeneratedSource()
			path := generatedExtensionsRoot + "backup-registry.json"
			registry := decodeMutableObject(t, source[path].JSON)
			profile := registry["profiles"].([]any)[0].(map[string]any)
			encoded := profile["codecs"].([]any)[0].(map[string]any)
			body := encoded["codec"].(map[string]any)
			mutate(body)
			canonical, err := canonicalJSON(body, true)
			if err != nil {
				t.Fatal(err)
			}
			digest := sha256.Sum256(canonical)
			encoded["backup_codec_sha256"] = hex.EncodeToString(digest[:])
			if name == "corrupted packaged digest" {
				encoded["backup_codec_sha256"] = hex.EncodeToString(make([]byte, sha256.Size))
			}
			source.replace(t, path, registry)
			integrityPath := generatedExtensionsRoot + "registry-integrity.json"
			integrity := decodeMutableObject(t, source[integrityPath].JSON)
			for _, raw := range integrity["supporting_contract_artifact_digests"].([]any) {
				row := raw.(map[string]any)
				if row["artifact_id"] == "backup-registry" {
					row["artifact_sha256"] = source[path].SHA256
				}
			}
			source.replace(t, integrityPath, integrity)
			_, err = NewCoordinator(source)
			requireFinding(t, err, "extension_registry_invalid", "registry_generation")
		})
	}
	source := newMutableGeneratedSource()
	bindingPath := generatedExtensionsRoot + "implementation-bindings/import.json"
	binding := decodeMutableObject(t, source[bindingPath].JSON)
	binding["contract_major"] = json.Number("99")
	source.replace(t, bindingPath, binding)

	integrityPath := generatedExtensionsRoot + "registry-integrity.json"
	integrity := decodeMutableObject(t, source[integrityPath].JSON)
	for _, raw := range integrity["implementation_binding_digests"].([]any) {
		row := raw.(map[string]any)
		if row["profile_id"] == "import" {
			row["binding_sha256"] = source[bindingPath].SHA256
		}
	}
	source.replace(t, integrityPath, integrity)
	_, err := NewCoordinator(source)
	requireFinding(t, err, "extension_implementation_unavailable", "implementation_binding")
}

func TestCoordinatorPublicationPlan_Unit(t *testing.T) {
	coordinator := requireGeneratedCoordinator(t)
	resolution, err := coordinator.ResolveClaims([]string{"import", "network_flow_activity"})
	if err != nil {
		t.Fatal(err)
	}
	plan, err := coordinator.BuildPublicationPlan(resolution)
	if err != nil {
		t.Fatal(err)
	}
	summary := plan.Summary()
	if summary.SchemaID != "cartulary.extension_publication_plan.v1" || summary.RegistrySHA256 != coordinator.RegistrySHA256() {
		t.Fatalf("unexpected publication summary: %#v", summary)
	}
	for _, digest := range []string{summary.ResolvedClaimSetSHA256, summary.ContributionRegistrySHA256, summary.RouteDispatchPlanSHA256, summary.WorkspaceRegistrySHA256, summary.WorkerPlanSHA256, summary.ListenerPlanSHA256, summary.ClientSupportRegistrySHA256, summary.ImplementationBindingSetSHA256} {
		if len(digest) != 64 {
			t.Fatalf("publication component digest %q is invalid", digest)
		}
	}
	routes := plan.Routes()
	claimed, inactive := 0, 0
	for _, route := range routes {
		if route.DispatchState == "claimed" {
			claimed++
			if route.ContributionID == nil {
				t.Fatal("claimed route has no contribution")
			}
		} else {
			inactive++
			if route.ContributionID != nil {
				t.Fatal("inactive route exposes a contribution")
			}
		}
	}
	if claimed != 3 || inactive != 10 || len(plan.Workspaces()) != 1 {
		t.Fatalf("publication route/workspace counts = %d/%d/%d", claimed, inactive, len(plan.Workspaces()))
	}
	discovery := plan.Discovery()
	discovery[0].RouteFamilies[0] = "mutated"
	if plan.Discovery()[0].RouteFamilies[0] == "mutated" {
		t.Fatal("publication plan leaked mutable discovery rows")
	}
	workers := plan.Workers()
	if len(workers) == 0 || len(workers[0].JobKinds) == 0 {
		t.Fatalf("publication plan omitted worker runtime contracts: %#v", workers)
	}
	workers[0].JobKinds[0] = "mutated.run_v1"
	if plan.Workers()[0].JobKinds[0] == "mutated.run_v1" {
		t.Fatal("publication plan leaked mutable worker job assignments")
	}
	if len(plan.Listeners()) != 3 {
		t.Fatalf("publication listener projection = %#v", plan.Listeners())
	}
	for _, binding := range plan.ImplementationBindings() {
		if len(binding.BindingSHA256) != 64 {
			t.Fatalf("publication binding projection = %#v", binding)
		}
	}
}

func requireGeneratedCoordinator(t testing.TB) *Coordinator {
	t.Helper()
	coordinator, err := NewGeneratedCoordinator()
	if err != nil {
		var validation *ValidationError
		if errors.As(err, &validation) {
			t.Fatalf("admit generated coordinator: %#v", validation.Findings())
		}
		t.Fatalf("admit generated coordinator: %v", err)
	}
	return coordinator
}

func requireFinding(t testing.TB, err error, code, phase string) {
	t.Helper()
	var validation *ValidationError
	if !errors.As(err, &validation) {
		t.Fatalf("error = %T %v; want ValidationError", err, err)
	}
	for _, finding := range validation.Findings() {
		if finding.Code == code && finding.Phase == phase {
			return
		}
	}
	t.Fatalf("findings = %#v; want %s/%s", validation.Findings(), code, phase)
}

func requireCollision(t testing.TB, err error, class string) {
	t.Helper()
	var validation *ValidationError
	if !errors.As(err, &validation) {
		t.Fatalf("error = %T %v; want ValidationError", err, err)
	}
	for _, finding := range validation.Findings() {
		if finding.Code == "extension_registry_conflict" && finding.CollisionClass == class {
			return
		}
	}
	t.Fatalf("findings = %#v; want collision %s", validation.Findings(), class)
}

type mutableArtifactSource map[string]PackagedArtifact

func newMutableGeneratedSource() mutableArtifactSource {
	source := mutableArtifactSource{}
	for path, artifact := range contractsgen.Index {
		if len(path) >= len("contracts/extensions/") && path[:len("contracts/extensions/")] == "contracts/extensions/" {
			source[path] = PackagedArtifact{JSON: []byte(artifact.JSON), SHA256: artifact.SHA256}
		}
	}
	return source
}

func (source mutableArtifactSource) Artifact(path string) (PackagedArtifact, bool) {
	artifact, ok := source[path]
	artifact.JSON = append([]byte(nil), artifact.JSON...)
	return artifact, ok
}

func (source mutableArtifactSource) ArtifactPaths() []string {
	paths := make([]string, 0, len(source))
	for path := range source {
		relative := strings.TrimPrefix(path, "contracts/extensions/")
		if (relative == "specification/inactive-value-schemas.json" || relative == "specification/shared-protocols.json") || (relative != "index.json" && relative != "dependencies.json" && relative != "build/implementation-bindings.json" && !strings.HasPrefix(relative, "specification/") && !strings.HasPrefix(relative, "validation/")) {
			paths = append(paths, path)
		}
	}
	sort.Strings(paths)
	return paths
}
func (source mutableArtifactSource) IntegritySHA256() string {
	return source[generatedExtensionsRoot+"registry-integrity.json"].SHA256
}

func (source mutableArtifactSource) replace(t testing.TB, path string, object map[string]any) {
	t.Helper()
	encoded, err := canonicalJSON(object, true)
	if err != nil {
		t.Fatal(err)
	}
	digest := sha256.Sum256(encoded)
	source[path] = PackagedArtifact{JSON: encoded, SHA256: hex.EncodeToString(digest[:])}
}

func decodeMutableObject(t testing.TB, data []byte) map[string]any {
	t.Helper()
	var object map[string]any
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.UseNumber()
	if err := decoder.Decode(&object); err != nil {
		t.Fatal(err)
	}
	return object
}
