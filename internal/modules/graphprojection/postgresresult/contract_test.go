package postgresresult

import (
	"encoding/json"
	"testing"

	contractgraphprojection "github.com/JochiRaider/cartulary/internal/gen/contractgraphprojection"
)

func TestStorageMaintenanceProjectionMatchesRuntime_Unit(t *testing.T) {
	t.Parallel()

	artifact := contractgraphprojection.Index["contracts/graph-projection/storage-maintenance.v1.json"]
	var contract map[string]any
	if err := json.Unmarshal([]byte(artifact.JSON), &contract); err != nil {
		t.Fatalf("decode Graph storage-maintenance contract: %v", err)
	}
	operations := contract["operations"].([]any)
	found := map[string]bool{}
	for _, raw := range operations {
		operation := raw.(map[string]any)
		id := operation["operation_id"].(string)
		found[id] = true
		if id == "release_scoped_leases_tx_v1" && (operation["maximum_rows"] != nil || operation["requires_transaction"] != true || operation["includes_expired"] != true || len(operation["exact_scope"].([]any)) != 4) {
			t.Fatalf("scoped release projection drift: %#v", operation)
		}
		if id == "lookup_unexpired_lease_v1" && (operation["identity"] != "stored_lease_id" || operation["expiry_predicate"] != "leased_until > observed_at" || len(operation["exact_key"].([]any)) != 4) {
			t.Fatalf("exact lookup projection drift: %#v", operation)
		}
	}
	for _, id := range []string{"release_scoped_leases_tx_v1", "lookup_unexpired_lease_v1"} {
		if !found[id] {
			t.Fatalf("missing lease capability %s", id)
		}
	}
	for _, raw := range operations {
		operation := raw.(map[string]any)
		if operation["operation_id"] == "delete_expired_leases_tx_v1" {
			if operation["maximum_rows"] != float64(maximumExpiredLeaseBatch) || operation["requires_transaction"] != true {
				t.Fatalf("expired-lease runtime/contract boundary drifted: %#v", operation)
			}
			return
		}
	}
	t.Fatal("Graph storage-maintenance contract omits expired-lease operation")
}
