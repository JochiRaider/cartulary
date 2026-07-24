package extensions

import (
	"errors"
	"fmt"
	"reflect"
	"testing"
)

func assertBC001EmptyStatePolicy(t *testing.T) {
	t.Helper()
	tests := []struct {
		policy        string
		metadata      bool
		authoritative bool
		decision      StatePresenceDecision
		err           error
	}{
		{"allowed", false, false, StateInitialize, nil},
		{"allowed", false, true, "", ErrStateMetadataMissing},
		{"allowed", true, false, StateValidate, nil},
		{"allowed", true, true, StateValidate, nil},
		{"forbidden", false, false, StateInitialize, nil},
		{"forbidden", false, true, "", ErrStateMetadataMissing},
		{"forbidden", true, false, "", ErrStateIncomplete},
		{"forbidden", true, true, StateValidate, nil},
		{"", false, false, "", ErrStateIncomplete},
	}
	for _, test := range tests {
		name := fmt.Sprintf("%s_metadata_%t_state_%t", test.policy, test.metadata, test.authoritative)
		t.Run(name, func(t *testing.T) {
			decision, err := DecideStatePresence(test.policy, test.metadata, test.authoritative)
			if decision != test.decision || !errors.Is(err, test.err) {
				t.Fatalf("decision = %q/%v; want %q/%v", decision, err, test.decision, test.err)
			}
		})
	}
	coordinator, err := NewGeneratedCoordinator()
	if err != nil {
		t.Fatal(err)
	}
	plan, ok := coordinator.StatePlan("network_flow_activity")
	if !ok ||
		plan.EmptyStatePolicy != "allowed" ||
		plan.InitializationKind != "empty" ||
		plan.InitializationAlgorithmID != "" ||
		plan.InitializationAlgorithmDefinitionSHA256 != "" ||
		plan.InitializationDefinitionSHA256 == "" ||
		plan.ImplementationBindingSHA256 == "" {
		t.Fatalf("Network Flow state plan = %#v/%t", plan, ok)
	}
	if !reflect.DeepEqual(plan.DatabaseFamilyIDs, []string{
		"network_flow_activity.indicator_bindings",
		"network_flow_activity.rejected_row_diagnostics",
		"network_flow_activity.rows",
		"network_flow_activity.tables",
	}) {
		t.Fatalf("authoritative presence families = %#v", plan.DatabaseFamilyIDs)
	}
}
