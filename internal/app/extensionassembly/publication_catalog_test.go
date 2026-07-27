package extensionassembly

import (
	"context"
	"errors"
	"reflect"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/extensions"
	"github.com/JochiRaider/cartulary/internal/modules/reporting"
)

func TestPublicationCatalog_ExactGeneratedSets_Unit(t *testing.T) {
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
	if got := catalog.ContributionIDs(""); len(got) != 18 {
		t.Fatalf("contribution catalog = %d; want 18: %v", len(got), got)
	}
	if got, want := catalog.WorkerKinds(), []string{
		"import.apply_worker_v1",
		"import.discovery_worker_v1",
		"incident_portability.bundle_worker_v1",
		"reference_pack.lifecycle_worker_v1",
		"snapshot_reporting.job_worker_v1",
	}; !reflect.DeepEqual(got, want) {
		t.Fatalf("worker catalog = %v; want %v", got, want)
	}
	if got := catalog.JobKinds(); len(got) != 10 {
		t.Fatalf("job catalog = %d; want 10: %v", len(got), got)
	}
	jobContracts, err := JobContracts(catalog)
	if err != nil {
		t.Fatal(err)
	}
	gotJobWorkers := make(map[string]string, len(jobContracts))
	for _, contract := range jobContracts {
		if !contract.ProofRequired || contract.ContractSHA256 == "" {
			t.Fatalf("job contract is not proof-bound: %#v", contract)
		}
		gotJobWorkers[contract.JobKind] = contract.WorkerKind
	}
	if !reflect.DeepEqual(gotJobWorkers, canonicalWorkerByJobKind) {
		t.Fatalf("job/worker catalog = %v; want %v", gotJobWorkers, canonicalWorkerByJobKind)
	}
	if got, want := catalog.ParticipantIDs(), []string{
		"network_flow_activity.backup_restore_v1",
		"network_flow_activity.import_apply_v1",
		"network_flow_activity.indicator_link_v1",
		"snapshot_reporting.render_export_v1",
	}; !reflect.DeepEqual(got, want) {
		t.Fatalf("participant catalog = %v; want %v", got, want)
	}
	if got := catalog.BindingProfileIDs(); !reflect.DeepEqual(got, profileIDs) {
		t.Fatalf("binding profile catalog = %v; want %v", got, profileIDs)
	}
	if _, present := catalog.Contribution("snapshot_reporting.render_export"); !present {
		t.Fatal("Snapshot/Reporting contribution is absent")
	}
	if participant, present := catalog.Participant("snapshot_reporting.render_export_v1"); !present ||
		participant.ContractKind != "cartulary.extension_participant_specialization.v3" {
		t.Fatalf("Snapshot/Reporting participant = %#v/%t", participant, present)
	}
}

func TestRenderExportInvokerRequiresExactAdmittedContract_Unit(t *testing.T) {
	coordinator, err := extensions.NewGeneratedCoordinator()
	if err != nil {
		t.Fatal(err)
	}
	buildCatalog := func(claimed bool) PublicationCatalog {
		t.Helper()
		claims := []string{}
		if claimed {
			claims = []string{reporting.ProfileID}
		}
		resolution, err := coordinator.ResolveClaims(claims)
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
		return catalog
	}
	if _, err := NewAdmittedRenderExportInvoker(
		buildCatalog(false),
		reporting.BuiltInRenderExportParticipant{},
		time.Second,
	); err == nil {
		t.Fatal("unclaimed Snapshot/Reporting participant was admitted")
	}
	claimed := buildCatalog(true)
	invoker, err := NewAdmittedRenderExportInvoker(
		claimed,
		blockingRenderExportParticipant{},
		time.Millisecond,
	)
	if err != nil {
		t.Fatalf("admit exact participant: %v", err)
	}
	if _, err := invoker.Invoke(context.Background(), reporting.RenderExportInvocation{}); !errors.Is(err, context.DeadlineExceeded) {
		t.Fatalf("participant timeout err = %v", err)
	}
	contract := claimed.participants[reporting.RenderExportParticipantID]
	contract.Operations[0].MaxOutputBytes--
	claimed.participants[reporting.RenderExportParticipantID] = contract
	if _, err := NewAdmittedRenderExportInvoker(
		claimed,
		reporting.BuiltInRenderExportParticipant{},
		time.Second,
	); err == nil {
		t.Fatal("participant with mismatched output bounds was admitted")
	}
}

type blockingRenderExportParticipant struct{}

func (blockingRenderExportParticipant) Emit(
	ctx context.Context,
	_ reporting.RenderExportInvocation,
) (reporting.RenderExportResult, error) {
	<-ctx.Done()
	return reporting.RenderExportResult{}, ctx.Err()
}

func TestPublicationCatalog_RejectsParticipantMismatch_Unit(t *testing.T) {
	coordinator, err := extensions.NewGeneratedCoordinator()
	if err != nil {
		t.Fatal(err)
	}
	resolution, err := coordinator.ResolveClaims([]string{"snapshot_reporting"})
	if err != nil {
		t.Fatal(err)
	}
	plan, err := coordinator.BuildPublicationPlan(resolution)
	if err != nil {
		t.Fatal(err)
	}
	participants := coordinator.ParticipantContracts()
	for index := range participants {
		if participants[index].ParticipantID == "snapshot_reporting.render_export_v1" {
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
