package extensions

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"testing"

	contractsgen "github.com/JochiRaider/cartulary/internal/gen/contractextensions"
)

func TestCoordinatorGeneratedRegistry_Unit(t *testing.T) {
	coordinator, err := NewGeneratedCoordinator()
	if err != nil {
		if validation, ok := err.(*ValidationError); ok {
			t.Fatalf("admit generated registry: %#v", validation.Findings())
		}
		t.Fatalf("admit generated registry: %v", err)
	}
	descriptors := coordinator.Descriptors()
	if len(descriptors) != 6 || len(coordinator.RegistrySHA256()) != 64 {
		t.Fatalf("generated registry identity = %d/%q", len(descriptors), coordinator.RegistrySHA256())
	}
	if descriptors[0].ProfileID != "enterprise_authentication" || descriptors[5].ProfileID != "snapshot_reporting" {
		t.Fatalf("descriptors are not canonical: %#v", descriptors)
	}
	inactivePolicies := coordinator.InactiveConfigurationPolicies()
	if len(inactivePolicies) != 3 ||
		inactivePolicies[0].Key != "enterprise_authentication.provider_manifest_path" ||
		inactivePolicies[0].Kind != "forbidden" ||
		inactivePolicies[1].Key != "network_flow_activity.key_ring_manifest_path" ||
		inactivePolicies[1].Kind != "forbidden" ||
		inactivePolicies[2].Key != "network_flow_activity.resource_limits" ||
		inactivePolicies[2].Kind != "forbidden" {
		t.Fatalf("inactive configuration policies = %#v", inactivePolicies)
	}
	descriptors[0].RouteFamilies[0] = "mutated"
	again, _ := coordinator.Descriptor("enterprise_authentication")
	if again.RouteFamilies[0] == "mutated" {
		t.Fatal("descriptor query leaked mutable coordinator state")
	}
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
		"network_flow_activity":     4,
		"reference_pack":            1,
		"snapshot_reporting":        1,
	}
	adoption := map[string][]jobIdentity{
		"enterprise_authentication": nil,
		"import": {
			{jobKind: "import.discovery_v1", operationKind: "import.discovery", workerKind: "import.discovery_worker_v1"},
			{jobKind: "import.apply_v1", operationKind: "import.apply", workerKind: "import.apply_worker_v1"},
		},
		"incident_portability": {
			{jobKind: "incident_portability.export_v1", operationKind: "incident_portability.export", workerKind: "incident_portability.bundle_worker_v1"},
			{jobKind: "incident_portability.import_v1", operationKind: "incident_portability.import", workerKind: "incident_portability.bundle_worker_v1"},
		},
		"network_flow_activity": {
			{jobKind: "network_flow_activity.graph_view_materialize_v1", operationKind: "network_flow_activity.graph_view_materialize", workerKind: "network_flow_activity.graph_view_worker_v1"},
		},
		"reference_pack": {
			{jobKind: "reference_pack.import_v1", operationKind: "reference_pack.import", workerKind: "reference_pack.lifecycle_worker_v1"},
			{jobKind: "reference_pack.reverify_v1", operationKind: "reference_pack.reverify", workerKind: "reference_pack.lifecycle_worker_v1"},
			{jobKind: "reference_pack.refresh_v1", operationKind: "reference_pack.refresh", workerKind: "reference_pack.lifecycle_worker_v1"},
		},
		"snapshot_reporting": {
			{jobKind: "snapshot_reporting.snapshot_create_v1", operationKind: "snapshot_reporting.snapshot_create", workerKind: "snapshot_reporting.job_worker_v1"},
			{jobKind: "snapshot_reporting.release_create_v1", operationKind: "snapshot_reporting.release_create", workerKind: "snapshot_reporting.job_worker_v1"},
			{jobKind: "snapshot_reporting.composition_preview_v1", operationKind: "snapshot_reporting.composition_preview", workerKind: "snapshot_reporting.job_worker_v1"},
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
		"incident_portability.export_v1":                  "incident_portability.export.request.v1",
		"incident_portability.import_v1":                  "incident_portability.import.request.v1",
		"network_flow_activity.graph_view_materialize_v1": "network_flow_activity.graph_view_materialize.projection_result.v1",
		"reference_pack.import_v1":                        "reference_pack.import.request.v1",
		"reference_pack.refresh_v1":                       "reference_pack.refresh.pack_key.v1",
		"reference_pack.reverify_v1":                      "reference_pack.reverify.pack_version.v1",
		"snapshot_reporting.composition_preview_v1":       "snapshot_reporting.composition_preview.render_attempt.v1",
		"snapshot_reporting.release_create_v1":            "snapshot_reporting.release_create.render_attempt.v1",
		"snapshot_reporting.snapshot_create_v1":           "snapshot_reporting.snapshot_create.materialization.v1",
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
	if networkFlow, ok := byProfile["network_flow_activity"]; !ok || !networkFlow.Claimable || networkFlow.ContractMajor != 4 {
		t.Fatalf("Network Flow v4 adopted profile = %#v/%t", networkFlow, ok)
	}

	var participant *ParticipantContract
	for _, contract := range coordinator.ParticipantContracts() {
		if contract.ParticipantID == "snapshot_reporting.render_export_v1" {
			copy := contract
			participant = &copy
			break
		}
	}
	if participant == nil || participant.OwnerProfileID != "snapshot_reporting" {
		t.Fatalf("Snapshot/Reporting participant identity = %#v", participant)
	}
	if participant.ContractKind != "cartulary.extension_participant_specialization.v3" ||
		participant.InputSchemaID != "cartulary.extension_snapshot_reporting_participant_context.v1" ||
		!reflect.DeepEqual(participant.AlgorithmIDs, []string{
			"materialize_reporting_export_model_v1",
			"snapshot_reporting.render_export_v1",
		}) ||
		len(participant.Operations) != 1 ||
		participant.Operations[0].OperationKind != "emit" ||
		participant.Operations[0].ResultSchemaID != "cartulary.extension_snapshot_reporting_participant_result.v1" ||
		participant.Operations[0].OutputSchemaID != "cartulary.reporting_export_model.v1" {
		t.Fatalf("Snapshot/Reporting participant specialization = %#v", participant)
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
	if networkFlow.Mode != PortabilityBlockedWhenPresent ||
		!reflect.DeepEqual(networkFlow.BlockingFamilyIDs, []string{
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
	policies[3].BlockingFamilyIDs[0] = "mutated"
	if coordinator.PortabilityPolicies()[3].BlockingFamilyIDs[0] == "mutated" {
		t.Fatal("portability policy projection leaked mutable state")
	}
}

func TestCoordinatorClaimResolution_Unit(t *testing.T) {
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
	if claimed != 2 || inactive != 9 || len(plan.Workspaces()) != 1 {
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
		if validation, ok := err.(*ValidationError); ok {
			t.Fatalf("admit generated coordinator: %#v", validation.Findings())
		}
		t.Fatalf("admit generated coordinator: %v", err)
	}
	return coordinator
}

func requireFinding(t testing.TB, err error, code, phase string) {
	t.Helper()
	validation, ok := err.(*ValidationError)
	if !ok {
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
	validation, ok := err.(*ValidationError)
	if !ok {
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
