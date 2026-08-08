package jobs

import (
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
}
