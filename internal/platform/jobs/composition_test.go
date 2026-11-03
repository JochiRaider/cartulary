package jobs

import (
	"errors"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/jobs/jobcontract"
)

func TestImmutableCompositionContract_Unit(t *testing.T) {
	t.Run("attributed admission", testAttributedAdmission)
	t.Run("job policy matrix", testJobPolicyMatrix)
	definitions := []Definition{{
		JobKind:        "test_profile.run_v1",
		ProgressUnitID: "test_profile.run.attempt.v1",
		HandlerName:    "test_profile.worker_v1",
		Terminal:       jobcontract.TerminalContract{SchemaID: jobcontract.TerminalSuccessSchema, SuccessCodes: []string{"done"}, Resources: []jobcontract.ResourceContract{{Kind: "incident", SchemaID: jobcontract.ResourceIDSchema, MinRefs: 1, MaxRefs: 1}}},
		Extension: &ExtensionPolicy{IdempotencyPolicy: "required", CancellationPolicy: "precommit_observable", IdentitySchemaID: HumanRouteIdentitySchema,
			OwnerProfileID: "test_profile",
			OperationKind:  "test_profile.run",
			ContractSHA256: strings.Repeat("a", 64),
			ProofRequired:  true,
			MaxProofBytes:  4096,
		},
	}}
	catalog, err := NewCatalog(definitions)
	if err != nil {
		t.Fatalf("NewCatalog: %v", err)
	}
	definitions[0].JobKind = "mutated.run_v1"
	definitions[0].Terminal.Resources[0].Kind = "mutated.resource"
	definitions[0].Terminal.SuccessCodes[0] = "mutated"
	definition, present := catalog.definition("test_profile.run_v1")
	if !present || definition.JobKind != "test_profile.run_v1" ||
		definition.Extension == nil || definition.Terminal.Resources[0].Kind != "incident" || definition.Terminal.SuccessCodes[0] != "done" {
		t.Fatalf("catalog changed through caller mutation: %#v", definition)
	}
	if _, err := NewManager(ManagerOptions{}); err == nil {
		t.Fatal("NewManager accepted missing dependencies")
	}
	if _, err := NewRunner(RunnerOptions{}); err == nil {
		t.Fatal("NewRunner accepted missing dependencies")
	}
	emptySelection, err := NewRuntimeSelection(catalog, nil, nil)
	if err != nil || len(emptySelection.jobKinds()) != 0 || len(emptySelection.handlerNames()) != 0 {
		t.Fatalf("empty runtime selection = %#v/%v", emptySelection, err)
	}
	if _, err := NewRuntimeSelection(catalog, []string{"unknown.run_v1"}, nil); !errors.Is(err, ErrInvalidJobDefinition) {
		t.Fatalf("unknown runtime selection error = %v", err)
	}
	if _, err := NewRuntimeSelection(catalog, []string{"test_profile.run_v1", "test_profile.run_v1"}, nil); !errors.Is(err, ErrInvalidJobDefinition) {
		t.Fatalf("duplicate runtime selection error = %v", err)
	}
	worker := WorkerRuntimeContract{ProfileID: "test_profile", WorkerKind: "test_profile.worker_v1", JobKinds: []string{"test_profile.run_v1"}, MaxActiveAttemptsPerProcess: 1}
	selection, err := NewRuntimeSelection(catalog, []string{"test_profile.run_v1"}, []WorkerRuntimeContract{worker})
	if err != nil || selection.workerByJob["test_profile.run_v1"] != worker.WorkerKind {
		t.Fatalf("worker runtime selection = %#v/%v", selection, err)
	}
	worker.JobKinds[0] = "mutated.run_v1"
	if selected, _ := selection.workerContract("test_profile.worker_v1"); selected.JobKinds[0] != "test_profile.run_v1" {
		t.Fatalf("worker runtime selection changed through caller mutation: %#v", selected)
	}
	for name, contracts := range map[string][]WorkerRuntimeContract{
		"missing":       nil,
		"unknown":       {{ProfileID: "test_profile", WorkerKind: "test_profile.worker_v1", JobKinds: []string{"unknown.run_v1"}, MaxActiveAttemptsPerProcess: 1}},
		"wrong_handler": {{ProfileID: "test_profile", WorkerKind: "test_profile.other_worker_v1", JobKinds: []string{"test_profile.run_v1"}, MaxActiveAttemptsPerProcess: 1}},
		"wrong_profile": {{ProfileID: "base", WorkerKind: "test_profile.worker_v1", JobKinds: []string{"test_profile.run_v1"}, MaxActiveAttemptsPerProcess: 1}},
		"zero_capacity": {{ProfileID: "test_profile", WorkerKind: "test_profile.worker_v1", JobKinds: []string{"test_profile.run_v1"}, MaxActiveAttemptsPerProcess: 0}},
	} {
		t.Run(name, func(t *testing.T) {
			if _, err := NewRuntimeSelection(catalog, []string{"test_profile.run_v1"}, contracts); !errors.Is(err, ErrInvalidJobDefinition) {
				t.Fatalf("invalid worker selection error = %v", err)
			}
		})
	}
	if _, err := NewCatalog(nil); !errors.Is(err, ErrInvalidJobDefinition) {
		t.Fatalf("empty catalog error = %v", err)
	}
}
