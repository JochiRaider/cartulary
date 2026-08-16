package recovery

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"slices"
	"sort"
	"strings"
	"testing"

	contractrecovery "github.com/JochiRaider/cartulary/internal/gen/contractrecovery"
)

func TestVNextGraphRestoreV2ProjectionContract_Unit(t *testing.T) {
	t.Parallel()

	var registry struct {
		SchemaID string `json:"schema_id"`
		Entries  []struct {
			SourceOwnerID              string `json:"source_owner_id"`
			AuthoritativeFamilyID      string `json:"authoritative_family_id"`
			ProjectionInputContractID  string `json:"projection_input_contract_id"`
			ProjectionResultContractID string `json:"projection_result_contract_id"`
		} `json:"entries"`
	}
	if err := json.Unmarshal([]byte(contractrecovery.CurrentGraphProjectionRestoreSourceRegistryJSON), &registry); err != nil {
		t.Fatalf("decode generated Graph restore v2 registry: %v", err)
	}
	if registry.SchemaID != "cartulary.graph_projection_restore_source_registry.v2" || len(registry.Entries) != 1 {
		t.Fatalf("current Graph restore registry is not the exact v2 registry: %#v", registry)
	}
	entry := registry.Entries[0]
	if entry.SourceOwnerID != "network_flow_activity" ||
		entry.AuthoritativeFamilyID != "network_flow_activity.graph_views" ||
		entry.ProjectionInputContractID != "graph_projection.v2" ||
		entry.ProjectionResultContractID != "graph_projection_result.v2" {
		t.Fatalf("current Graph restore registry entry drifted: %#v", entry)
	}

	var binding struct {
		SchemaID      string   `json:"schema_id"`
		AlgorithmID   string   `json:"algorithm_id"`
		GraphTableIDs []string `json:"graph_table_ids"`
	}
	if err := json.Unmarshal([]byte(contractrecovery.CurrentGraphProjectionRestoreImplementationBindingJSON), &binding); err != nil {
		t.Fatalf("decode generated Graph restore v2 binding: %v", err)
	}
	wantTables := []string{
		"graph_projection_result_edges",
		"graph_projection_result_leases",
		"graph_projection_result_vertices",
		"graph_projection_results",
	}
	if binding.SchemaID != "cartulary.graph_projection_restore_implementation_binding.v2" ||
		binding.AlgorithmID != "graphprojection.restore_rebuild.v2" ||
		!slices.Equal(binding.GraphTableIDs, wantTables) {
		t.Fatalf("current Graph restore implementation binding drifted: %#v", binding)
	}
}

func TestVNextGraphRestoreArtifactsFailClosedAndRejectRetiredRegistry_Unit(t *testing.T) {
	service := &VNextRestoreService{}
	current := VNextBackupIntegrityManifest{CodecRegistrySHA256: VNextCodecRegistrySHA256()}
	_, err := service.resolveGraphProjectionRestoreArtifacts(context.Background(), current, map[string]VNextArtifactProof{
		"registry": {
			Kind: "graph_projection_restore_source_registry", SchemaID: GraphProjectionRestoreSourceRegistryV2SchemaID,
			LogicalRef: "registry", PlaintextSHA256: contractrecovery.CurrentGraphProjectionRestoreSourceRegistrySHA256,
		},
	})
	if err == nil || !errors.Is(err, ErrVNextBackup) {
		t.Fatalf("current backup without implementation binding did not fail closed: %v", err)
	}

	retired := VNextBackupIntegrityManifest{CodecRegistrySHA256: retiredEmptyRegistryCodecSHA256ForTest()}
	if retired.CodecRegistrySHA256 == VNextCodecRegistrySHA256() {
		t.Fatal("retired empty-registry codec digest unexpectedly equals the current inventory")
	}
	if _, err = service.resolveGraphProjectionRestoreArtifacts(context.Background(), retired, map[string]VNextArtifactProof{}); err == nil || !errors.Is(err, ErrVNextBackup) {
		t.Fatalf("retired empty-registry backup remained admitted: %v", err)
	}
}

func retiredEmptyRegistryCodecSHA256ForTest() string {
	codecs := []string{
		BackupArtifactEnvelopeV2SchemaID,
		BackupIntegrityManifestV3SchemaID,
		ObjectStoreBackupManifestV2SchemaID,
		ObjectStoreBackupSummaryV2SchemaID,
		PostgresSnapshotArtifactV2SchemaID,
		PostgresSnapshotUnitV1SchemaID,
	}
	sort.Strings(codecs)
	sum := sha256.Sum256([]byte(vNextCodecRegistryDomain + strings.Join(codecs, "\n") + "\n"))
	return hex.EncodeToString(sum[:])
}
