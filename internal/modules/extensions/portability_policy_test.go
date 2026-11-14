package extensions

import (
	"reflect"
	"testing"
)

func checkPortabilityOperationPolicies(t *testing.T) {
	source := newMutableGeneratedSource()
	rewriteCoherentPackage(t, source, func(object map[string]any) {
		if object["schema_id"] != "cartulary.extension_participant_specialization.v5" || object["participant_id"] != "network_flow_activity.incident_portability_v1" {
			return
		}
		for _, value := range object["operations"].([]any) {
			operation := value.(map[string]any)
			if operation["operation_kind"] == "export" {
				operation["max_input_bytes"], operation["max_output_bytes"], operation["max_items"] = 1024, 2048, 3
			} else {
				operation["max_input_bytes"], operation["max_output_bytes"], operation["max_items"] = 4096, 8192, 7
			}
		}
	})
	coordinator, err := NewCoordinator(source)
	if err != nil {
		t.Fatal(err)
	}
	families := []string{"network_flow_activity.graph_views", "network_flow_activity.indicator_bindings", "network_flow_activity.rejected_row_diagnostics", "network_flow_activity.rows", "network_flow_activity.tables"}
	wantExport := PortabilityOperationPolicy{OutputSchemaID: "cartulary.network_flow.incident_members.v1", StateFamilyIDs: families, MaximumInputBytes: 1024, MaximumOutputBytes: 2048, MaximumItems: 3}
	wantImport := PortabilityOperationPolicy{OutputSchemaID: "cartulary.network_flow.incident_import_input.v1", StateFamilyIDs: families, MaximumInputBytes: 4096, MaximumOutputBytes: 8192, MaximumItems: 7}
	var selected *PortabilityPolicy
	for _, policy := range coordinator.PortabilityPolicies() {
		if policy.ProfileID == "network_flow_activity" {
			selected = &policy
		}
	}
	if selected == nil || !reflect.DeepEqual(selected.Export, wantExport) || !reflect.DeepEqual(selected.Import, wantImport) {
		t.Fatalf("admitted independent operation policies = %#v", selected)
	}
	selected.Export.StateFamilyIDs[0] = "mutated export"
	selected.Import.StateFamilyIDs[0] = "mutated import"
	for _, policy := range coordinator.PortabilityPolicies() {
		if policy.ProfileID == "network_flow_activity" && (!reflect.DeepEqual(policy.Export, wantExport) || !reflect.DeepEqual(policy.Import, wantImport)) {
			t.Fatalf("operation policy leaked caller mutation: %#v", policy)
		}
	}
}
