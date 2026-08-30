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
