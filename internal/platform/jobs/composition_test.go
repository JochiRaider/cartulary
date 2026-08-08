package jobs

import (
	"errors"
	"strings"
	"testing"
)

func TestImmutableCompositionContract_Unit(t *testing.T) {
	definitions := []Definition{{
		JobKind:        "test_profile.run_v1",
		ProgressUnitID: "test_profile.run.attempt.v1",
		HandlerName:    "test_profile.worker_v1",
		Extension: &ExtensionPolicy{
			OwnerProfileID: "test_profile",
			OperationKind:  "test_profile.run",
			ContractSHA256: strings.Repeat("a", 64),
			ProofRequired:  true,
			MaxProofBytes:  4096,
			ResourceRefs: []ExtensionResourceRefContract{{
				Kind:    "test_profile.resource",
				MaxRefs: 1,
			}},
		},
	}}
	catalog, err := NewCatalog(definitions)
	if err != nil {
		t.Fatalf("NewCatalog: %v", err)
	}
	definitions[0].JobKind = "mutated.run_v1"
	definitions[0].Extension.ResourceRefs[0].Kind = "mutated.resource"
	definition, present := catalog.definition("test_profile.run_v1")
	if !present || definition.JobKind != "test_profile.run_v1" ||
		definition.Extension == nil || definition.Extension.ResourceRefs[0].Kind != "test_profile.resource" {
		t.Fatalf("catalog changed through caller mutation: %#v", definition)
	}
	if _, err := NewManager(ManagerOptions{}); err == nil {
		t.Fatal("NewManager accepted missing dependencies")
	}
	if _, err := NewRunner(RunnerOptions{}); err == nil {
		t.Fatal("NewRunner accepted missing dependencies")
	}
	emptySelection, err := NewRuntimeSelection(catalog, nil)
	if err != nil || len(emptySelection.jobKinds()) != 0 || len(emptySelection.handlerNames()) != 0 {
		t.Fatalf("empty runtime selection = %#v/%v", emptySelection, err)
	}
	if _, err := NewRuntimeSelection(catalog, []string{"unknown.run_v1"}); !errors.Is(err, ErrInvalidJobDefinition) {
		t.Fatalf("unknown runtime selection error = %v", err)
	}
	if _, err := NewRuntimeSelection(catalog, []string{"test_profile.run_v1", "test_profile.run_v1"}); !errors.Is(err, ErrInvalidJobDefinition) {
		t.Fatalf("duplicate runtime selection error = %v", err)
	}
	if _, err := NewCatalog(nil); !errors.Is(err, ErrInvalidJobDefinition) {
		t.Fatalf("empty catalog error = %v", err)
	}
}
