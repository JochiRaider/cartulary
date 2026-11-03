package postgresresult

import (
	"encoding/json"
	"testing"

	contractgraphprojection "github.com/JochiRaider/cartulary/internal/gen/contractgraphprojection"
)

func TestStorageMaintenanceProjectionMatchesRuntime_Unit(t *testing.T) {
	t.Parallel()

	artifact := contractgraphprojection.Index["contracts/graph-projection/storage-maintenance.v2.json"]
	var contract map[string]any
	if err := json.Unmarshal([]byte(artifact.JSON), &contract); err != nil {
		t.Fatalf("decode Graph storage-maintenance contract: %v", err)
	}
	if contract["schema_id"] != "cartulary.graph_projection_storage_maintenance.v2" || contract["transaction_ownership"] != "borrowed" {
		t.Fatal("current borrowed storage contract missing")
	}
	operations := contract["operations"].([]any)
	if len(operations) != 2 {
		t.Fatal("unexpected storage maintenance capability")
	}
	found := map[string]bool{}
	for _, raw := range operations {
		op := raw.(map[string]any)
		id := op["operation_id"].(string)
		if (id != "lock_cleanup_candidate_tx_v1" && id != "delete_locked_result_tx_v1") || found[id] || op["maximum_rows"] != float64(1) || op["requires_transaction"] != true {
			t.Fatalf("invalid borrowed operation: %#v", op)
		}
		found[id] = true
	}
}
