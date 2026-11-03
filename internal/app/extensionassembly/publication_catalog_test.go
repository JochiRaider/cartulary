package extensionassembly

import (
	"reflect"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/extensions"
	"github.com/JochiRaider/cartulary/internal/modules/networkflow"
)

func TestPublicationCatalog_ExactGeneratedSets_Unit(t *testing.T) {
	t.Run("Reporting source contract admission", checkReportingSourceAdmission)
	coordinator, err := extensions.NewGeneratedCoordinator()
	if err != nil {
		t.Fatal(err)
	}
	profileIDs := []string{
		"enterprise_authentication",
		"import",
		"incident_portability",
		"network_flow_activity",
		"reference_pack",
		"snapshot_reporting",
	}
	resolution, err := coordinator.ResolveClaims(profileIDs)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := coordinator.BuildPublicationPlan(resolution)
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := NewPublicationCatalog(plan, coordinator.ParticipantContracts())
	if err != nil {
		t.Fatal(err)
	}
	if got := catalog.ContributionIDs(""); len(got) != 20 {
		t.Fatalf("contribution catalog = %d; want 20: %v", len(got), got)
	}
	if got, want := catalog.WorkerKinds(), []string{
		"import.apply_worker_v1",
		"import.discovery_worker_v1",
		"incident_portability.bundle_worker_v2",
		"network_flow_activity.graph_view_worker_v1",
		"reference_pack.lifecycle_worker_v2",
		"snapshot_reporting.job_worker_v2",
	}; !reflect.DeepEqual(got, want) {
		t.Fatalf("worker catalog = %v; want %v", got, want)
	}
	if got := catalog.JobKinds(); len(got) != 11 {
		t.Fatalf("job catalog = %d; want 11: %v", len(got), got)
	}
	jobContracts, err := JobDefinitions(catalog)
	if err != nil {
		t.Fatal(err)
	}
	gotJobWorkers := make(map[string]string, len(jobContracts))
	for _, contract := range jobContracts {
		if contract.Extension == nil || !contract.Extension.ProofRequired || contract.Extension.ContractSHA256 == "" {
			t.Fatalf("job contract is not proof-bound: %#v", contract)
		}
		gotJobWorkers[contract.JobKind] = contract.HandlerName
	}
	wantJobWorkers := map[string]string{
		"import.discovery_v1":                             "import.discovery_worker_v1",
		"import.apply_v1":                                 "import.apply_worker_v1",
		"incident_portability.export_v2":                  "incident_portability.bundle_worker_v2",
		"incident_portability.import_v2":                  "incident_portability.bundle_worker_v2",
		"reference_pack.import_v2":                        "reference_pack.lifecycle_worker_v2",
		"reference_pack.reverify_v2":                      "reference_pack.lifecycle_worker_v2",
		"reference_pack.refresh_v2":                       "reference_pack.lifecycle_worker_v2",
		"snapshot_reporting.snapshot_create_v4":           "snapshot_reporting.job_worker_v2",
		"snapshot_reporting.release_create_v4":            "snapshot_reporting.job_worker_v2",
		"snapshot_reporting.composition_preview_v4":       "snapshot_reporting.job_worker_v2",
		"network_flow_activity.graph_view_materialize_v1": "network_flow_activity.graph_view_worker_v1",
	}
	if !reflect.DeepEqual(gotJobWorkers, wantJobWorkers) {
		t.Fatalf("job/worker catalog = %v; want %v", gotJobWorkers, wantJobWorkers)
	}
	workerContracts, err := WorkerRuntimeContracts(catalog)
	if err != nil {
		t.Fatal(err)
	}
	for _, worker := range workerContracts {
		wantMaximum := 8
		if worker.WorkerKind == "network_flow_activity.graph_view_worker_v1" {
			wantMaximum = 1
		}
		if worker.MaxActiveAttemptsPerProcess != wantMaximum {
			t.Fatalf("worker capacity = %#v; want %d", worker, wantMaximum)
		}
	}
	if got, want := catalog.ParticipantIDs(), []string{
		"network_flow_activity.import_apply_v1",
		"network_flow_activity.indicator_link_v1",
		"network_flow_activity.reporting_graph_source_v1",
	}; !reflect.DeepEqual(got, want) {
		t.Fatalf("participant catalog = %v; want %v", got, want)
	}
	if got := catalog.BindingProfileIDs(); !reflect.DeepEqual(got, profileIDs) {
		t.Fatalf("binding profile catalog = %v; want %v", got, profileIDs)
	}
	if _, present := catalog.Contribution("snapshot_reporting.render_export"); present {
		t.Fatal("retired self-contribution remains")
	}
	if _, present := catalog.Participant("snapshot_reporting.render_export_v1"); present {
		t.Fatal("retired self-participant remains")
	}

}

func TestPublicationCatalog_RejectsParticipantMismatch_Unit(t *testing.T) {
	coordinator, err := extensions.NewGeneratedCoordinator()
	if err != nil {
		t.Fatal(err)
	}
	resolution, err := coordinator.ResolveClaims([]string{"import", "network_flow_activity"})
	if err != nil {
		t.Fatal(err)
	}
	plan, err := coordinator.BuildPublicationPlan(resolution)
	if err != nil {
		t.Fatal(err)
	}
	participants := coordinator.ParticipantContracts()
	for index := range participants {
		if participants[index].ParticipantID == "network_flow_activity.reporting_graph_source_v1" {
			participants[index].OwnerProfileID = "import"
		}
	}
	if _, err := NewPublicationCatalog(plan, participants); err == nil {
		t.Fatal("participant owner mismatch was accepted")
	}
}

func TestPublicationCatalog_ExactProfileContributionSet_Unit(t *testing.T) {
	t.Parallel()
	coordinator, err := extensions.NewGeneratedCoordinator()
	if err != nil {
		t.Fatal(err)
	}
	expected := []string{
		"enterprise_authentication.auth_oidc_route",
		"enterprise_authentication.auth_providers_route",
		"enterprise_authentication.auth_saml_route",
		"enterprise_authentication.user_auth_bindings_route",
	}

	t.Run("unadmitted profile has no executable projection", func(t *testing.T) {
		resolution, resolveErr := coordinator.ResolveClaims(nil)
		if resolveErr != nil {
			t.Fatal(resolveErr)
		}
		plan, planErr := coordinator.BuildPublicationPlan(resolution)
		if planErr != nil {
			t.Fatal(planErr)
		}
		catalog, catalogErr := NewPublicationCatalog(plan, coordinator.ParticipantContracts())
		if catalogErr != nil {
			t.Fatal(catalogErr)
		}
		admitted, projectionErr := catalog.ExactProfileContributionSet("enterprise_authentication", "http_route_family", expected)
		if projectionErr != nil {
			t.Fatal(projectionErr)
		}
		if admitted {
			t.Fatal("unclaimed Enterprise Authentication profile was admitted")
		}
	})

	t.Run("admitted profile requires the exact application set", func(t *testing.T) {
		resolution, resolveErr := coordinator.ResolveClaims([]string{"enterprise_authentication"})
		if resolveErr != nil {
			t.Fatal(resolveErr)
		}
		plan, planErr := coordinator.BuildPublicationPlan(resolution)
		if planErr != nil {
			t.Fatal(planErr)
		}
		catalog, catalogErr := NewPublicationCatalog(plan, coordinator.ParticipantContracts())
		if catalogErr != nil {
			t.Fatal(catalogErr)
		}
		admitted, projectionErr := catalog.ExactProfileContributionSet("enterprise_authentication", "http_route_family", expected)
		if projectionErr != nil {
			t.Fatal(projectionErr)
		}
		if !admitted {
			t.Fatal("claimed Enterprise Authentication profile was not admitted")
		}

		delete(catalog.contributions, expected[0])
		if _, projectionErr := catalog.ExactProfileContributionSet("enterprise_authentication", "http_route_family", expected); projectionErr == nil {
			t.Fatal("partial Enterprise Authentication application set was accepted")
		}
	})
}

func checkReportingSourceAdmission(t *testing.T) {
	coordinator, err := extensions.NewGeneratedCoordinator()
	if err != nil {
		t.Fatal(err)
	}
	resolution, err := coordinator.ResolveClaims([]string{"import", networkflow.ProfileID})
	if err != nil {
		t.Fatal(err)
	}
	plan, err := coordinator.BuildPublicationPlan(resolution)
	if err != nil {
		t.Fatal(err)
	}
	catalog, err := NewPublicationCatalog(plan, coordinator.ParticipantContracts())
	if err != nil {
		t.Fatal(err)
	}
	contract, ok := catalog.Participant(networkflow.ReportingGraphSourceParticipantID)
	if !ok {
		t.Fatal("missing actual source contract")
	}
	if err := validateNetworkFlowReportingSourceContract(contract); err != nil {
		t.Fatal(err)
	}
	for name, mutate := range map[string]func(*extensions.ParticipantContract){
		"schema": func(c *extensions.ParticipantContract) {
			c.ContractKind = "cartulary.extension_participant_specialization.v3"
		},
		"profile":     func(c *extensions.ParticipantContract) { c.OwnerProfileID = "snapshot_reporting" },
		"participant": func(c *extensions.ParticipantContract) { c.ParticipantID += "_unknown" },
		"kind":        func(c *extensions.ParticipantContract) { c.ParticipantKind = "incident_portability" },
		"context": func(c *extensions.ParticipantContract) {
			c.InputSchemaID = "cartulary.extension_snapshot_reporting_participant_context.v1"
		},
		"operation": func(c *extensions.ParticipantContract) {
			c.Operations = []extensions.ParticipantOperation{{OperationKind: "emit"}}
		},
		"family": func(c *extensions.ParticipantContract) {
			c.OwnedStateFamilyIDs = []string{"network_flow_activity.rows"}
		},
		"extra family": func(c *extensions.ParticipantContract) {
			c.OwnedStateFamilyIDs = append(c.OwnedStateFamilyIDs, "network_flow_activity.rows")
		},
		"algorithm set": func(c *extensions.ParticipantContract) { c.AlgorithmIDs = []string{"unimplemented"} },
		"projection":    func(c *extensions.ParticipantContract) { c.ReportingSource.ProjectionSchemaID = "graph_projection.v1" },
		"labels":        func(c *extensions.ParticipantContract) { c.ReportingSource.LabelSchemaID = "retired" },
		"read":          func(c *extensions.ParticipantContract) { c.ReportingSource.ReadAlgorithmID = "unimplemented" },
		"order":         func(c *extensions.ParticipantContract) { c.ReportingSource.OrderingAlgorithmID = "unimplemented" },
		"vertices":      func(c *extensions.ParticipantContract) { c.ReportingSource.MaxVertices-- },
		"edges":         func(c *extensions.ParticipantContract) { c.ReportingSource.MaxEdges-- },
		"bytes":         func(c *extensions.ParticipantContract) { c.ReportingSource.MaxLabelBytes-- },
	} {
		t.Run(name, func(t *testing.T) {
			changed, _ := catalog.Participant(networkflow.ReportingGraphSourceParticipantID)
			mutate(&changed)
			if validateNetworkFlowReportingSourceContract(changed) == nil {
				t.Fatal("unsupported source contract admitted")
			}
		})
	}
	if _, err := NewAdmittedNetworkFlowReportingSource(catalog, nil); err == nil {
		t.Fatal("missing source admitted")
	}
	if _, err := NewAdmittedNetworkFlowReportingSource(catalog, &networkflow.Module{}); err == nil {
		t.Fatal("uninitialized owner admitted")
	}
	if _, err := NewAdmittedNetworkFlowReportingSource(PublicationCatalog{}, nil); err == nil {
		t.Fatal("unclaimed source admitted")
	}
	if err := validateNetworkFlowReportingSourceContract(contract); err != nil {
		t.Fatal("catalog mutation leaked", err)
	}
}
