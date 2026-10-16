package jobassembly

import (
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/app/extensionassembly"
	"github.com/JochiRaider/cartulary/internal/gen/contractjobs"
	"github.com/JochiRaider/cartulary/internal/modules/extensions"
	"github.com/JochiRaider/cartulary/internal/platform/jobs"
)

func TestBaseJobsRemainSelectedWithoutOptionalClaims(t *testing.T) {
	coordinator, err := extensions.NewGeneratedCoordinator()
	if err != nil {
		t.Fatal(err)
	}
	resolution, err := coordinator.ResolveClaims(nil)
	if err != nil {
		t.Fatal(err)
	}
	plan, err := coordinator.BuildPublicationPlan(resolution)
	if err != nil {
		t.Fatal(err)
	}
	published, err := extensionassembly.NewPublicationCatalog(plan, coordinator.ParticipantContracts())
	if err != nil {
		t.Fatal(err)
	}
	selected, err := SelectedDefinitions(published)
	if err != nil {
		t.Fatal(err)
	}
	if len(selected) != 1 || selected[0].JobKind != "evidence.verify_blob_v1" || selected[0].Extension != nil {
		t.Fatalf("unexpected base selection: %#v", selected)
	}
	workers, err := WorkerContracts(published)
	if err != nil {
		t.Fatal(err)
	}
	if len(workers) != 1 || workers[0].ProfileID != "base" || workers[0].WorkerKind != "evidence.blob_verification_worker_v1" || workers[0].MaxActiveAttemptsPerProcess != 8 {
		t.Fatalf("unexpected base worker: %#v", workers)
	}
	recognized, err := RecognizedDefinitions(coordinator.JobKindContracts(), coordinator.WorkerRuntimeContracts())
	if err != nil {
		t.Fatal(err)
	}
	if len(recognized) <= len(selected) {
		t.Fatal("inactive optional definitions disappeared from recognition")
	}
	catalog, err := jobs.NewCatalog(recognized)
	if err != nil {
		t.Fatal(err)
	}
	if _, err = jobs.NewRuntimeSelection(catalog, extensionassembly.JobKinds(selected), workers); err != nil {
		t.Fatal(err)
	}
}

func TestBaseJobContractRejectsUnknownAndInconsistentInputs(t *testing.T) {
	raw := contractjobs.Index["contracts/jobs/base.v1.json"].JSON
	if _, _, err := decodeBaseDefinitions(raw); err != nil {
		t.Fatal(err)
	}
	for name, invalid := range map[string]string{
		"unknown field":         strings.Replace(raw, "{", "{\"compatibility\":true,", 1),
		"duplicate field":       strings.Replace(raw, "{", "{\"schema_id\":\"cartulary.base_job_catalog.v1\",", 1),
		"retired identity":      strings.ReplaceAll(raw, "cartulary.base_job_catalog.v1", "cartulary.base_job_catalog.v0"),
		"extension base worker": strings.ReplaceAll(raw, "\"profile_id\":\"base\"", "\"profile_id\":\"import\""),
		"over capacity":         strings.ReplaceAll(raw, "\"max_active_attempts_per_process\":8", "\"max_active_attempts_per_process\":9"),
		"unassigned handler":    strings.Replace(raw, "evidence.blob_verification_worker_v1", "evidence.unassigned_worker_v1", 1),
		"missing jobs":          "{\"schema_id\":\"cartulary.base_job_catalog.v1\",\"workers\":[]}",
		"trailing value":        raw + " {}",
	} {
		t.Run(name, func(t *testing.T) {
			if _, _, err := decodeBaseDefinitions(invalid); err == nil {
				t.Fatal("invalid base catalog was accepted")
			}
		})
	}
}
