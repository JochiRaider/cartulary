package recovery_test

import (
	"testing"

	"github.com/JochiRaider/cartulary/internal/app/extensionassembly"
	"github.com/JochiRaider/cartulary/internal/modules/recovery"
)

func testExtensionBackupCatalog(t testing.TB) *recovery.ExtensionBackupCatalog {
	t.Helper()
	catalog, err := extensionassembly.GeneratedRecoveryCatalog()
	if err != nil {
		t.Fatalf("construct generated extension recovery catalog: %v", err)
	}
	return catalog
}

func emptyExtensionPostgresArtifact() []byte {
	return []byte(`{"schema_id":"cartulary.postgres_snapshot_artifact.v1","tables":[{"table_name":"network_flow_graph_views","row_count":0,"rows":[]},{"table_name":"network_flow_indicator_bindings","row_count":0,"rows":[]},{"table_name":"network_flow_rejected_row_diagnostics","row_count":0,"rows":[]},{"table_name":"network_flow_rows","row_count":0,"rows":[]},{"table_name":"network_flow_tables","row_count":0,"rows":[]}]}`)
}
