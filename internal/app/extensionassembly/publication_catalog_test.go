package extensionassembly

import (
	"reflect"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/extensions"
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
		participant.ContractKind != "cartulary.extension_participant_specialization.v1" {
		t.Fatalf("Snapshot/Reporting participant = %#v/%t", participant, present)
	}
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
