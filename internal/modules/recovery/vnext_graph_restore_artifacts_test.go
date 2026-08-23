package recovery

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"slices"
	"sort"
	"strings"
	"testing"

	contractrecovery "github.com/JochiRaider/cartulary/internal/gen/contractrecovery"
)

func TestVNextGraphRestoreV3ProjectionContract_Unit(t *testing.T) {
	t.Parallel()
	if got, want := contractrecovery.CurrentGraphProjectionRestoreImplementationBindingSHA256, "18f7c517c18b0ed25d1950ccb950bc06fad63eb28d3a68176f0d795eaf65bbd9"; got != want {
		t.Fatalf("current Workbook-owned Graph v3 binding digest = %s, want %s", got, want)
	}
	if got, want := contractrecovery.RecoveryGenerations[1].GraphImplementationBindingSHA256, "6ec244d0b82466a18adbdb82554be29f5e4baac384175538acbc92e56f14b8d5"; got != want {
		t.Fatalf("pre-Workbook-ownership Graph v3 binding digest = %s, want %s", got, want)
	}
	if got, want := contractrecovery.CurrentGraphProjectionRestoreSourceRegistrySHA256, "61c3f7348c4df2bee3e969c905c91c9857082cf2839a0b57104e40339e3e16d3"; got != want {
		t.Fatalf("Graph v3 source registry digest = %s, want %s", got, want)
	}

	var registry struct {
		SchemaID string `json:"schema_id"`
		Entries  []struct {
			SourceOwnerID              string   `json:"source_owner_id"`
			AuthoritativeFamilyID      string   `json:"authoritative_family_id"`
			ProjectionInputContractID  string   `json:"projection_input_contract_id"`
			ProjectionResultContractID string   `json:"projection_result_contract_id"`
			SemanticQuerySchemaIDs     []string `json:"semantic_query_schema_ids"`
		} `json:"entries"`
	}
	if err := json.Unmarshal([]byte(contractrecovery.CurrentGraphProjectionRestoreSourceRegistryJSON), &registry); err != nil {
		t.Fatalf("decode generated Graph restore v3 registry: %v", err)
	}
	if registry.SchemaID != "cartulary.graph_projection_restore_source_registry.v3" || len(registry.Entries) != 1 {
		t.Fatalf("current Graph restore registry is not the exact v3 registry: %#v", registry)
	}
	entry := registry.Entries[0]
	if entry.SourceOwnerID != "network_flow_activity" ||
		entry.AuthoritativeFamilyID != "network_flow_activity.graph_views" ||
		entry.ProjectionInputContractID != "graph_projection.v2" ||
		entry.ProjectionResultContractID != "graph_projection_result.v2" ||
		!slices.Equal(entry.SemanticQuerySchemaIDs, []string{
			"cartulary.network_flow.graph_semantic_query.v1",
			"cartulary.network_flow.graph_semantic_query.v2",
		}) {
		t.Fatalf("current Graph restore registry entry drifted: %#v", entry)
	}

	var binding struct {
		SchemaID                       string   `json:"schema_id"`
		AlgorithmID                    string   `json:"algorithm_id"`
		GraphTableIDs                  []string `json:"graph_table_ids"`
		HistoricalDispatchAlgorithmIDs []string `json:"historical_dispatch_algorithm_ids"`
	}
	if err := json.Unmarshal([]byte(contractrecovery.CurrentGraphProjectionRestoreImplementationBindingJSON), &binding); err != nil {
		t.Fatalf("decode generated Graph restore v3 binding: %v", err)
	}
	wantTables := []string{
		"graph_projection_result_edges",
		"graph_projection_result_leases",
		"graph_projection_result_vertices",
		"graph_projection_results",
	}
	if binding.SchemaID != "cartulary.graph_projection_restore_implementation_binding.v3" ||
		binding.AlgorithmID != "graphprojection.restore_rebuild.v3" ||
		!slices.Equal(binding.HistoricalDispatchAlgorithmIDs, []string{"graphprojection.restore_rebuild.v2"}) ||
		!slices.Equal(binding.GraphTableIDs, wantTables) {
		t.Fatalf("current Graph restore implementation binding drifted: %#v", binding)
	}
}

func TestVNextGraphRestoreArtifactsFailClosedAndRejectRetiredRegistry_Unit(t *testing.T) {
	registry, err := loadVNextRecoveryGenerationRegistry()
	if err != nil {
		t.Fatalf("load Recovery generation registry: %v", err)
	}
	service := &VNextRestoreService{generations: registry}
	_, err = service.resolveGraphProjectionRestoreArtifacts(context.Background(), registry.current, map[string]VNextArtifactProof{
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
	if _, err = service.resolveGraphProjectionRestoreArtifacts(context.Background(), nil, map[string]VNextArtifactProof{}); err == nil || !errors.Is(err, ErrVNextBackup) {
		t.Fatalf("retired empty-registry backup remained admitted: %v", err)
	}
}

func TestVNextGraphRestoreArtifactsRetainExactHistoricalV2Dispatch_Unit(t *testing.T) {
	if got, want := contractrecovery.RecoveryGenerations[2].CatalogDigestSHA256, "ce0a1f4053a9ce156273e4adf40c8b4185fa616170eadd6a860500d0b24fd22f"; got != want {
		t.Fatalf("historical Graph v2 catalog digest = %s, want %s", got, want)
	}
	if got, want := contractrecovery.HistoricalGraphProjectionRestoreImplementationBindingV2SHA256, "235c69bbc0e5d4f25f3fab7b1f2b8c30ba6370bfc65abcba75822007802621b9"; got != want {
		t.Fatalf("historical Graph v2 binding digest = %s, want %s", got, want)
	}
	if got, want := contractrecovery.HistoricalGraphProjectionRestoreSourceRegistryV2SHA256, "e75697ef1f6b5a197d299746fd42d2bf07afcd2e1c9d187a6fe695bca3096730"; got != want {
		t.Fatalf("historical Graph v2 source registry digest = %s, want %s", got, want)
	}
	registryBody := []byte(contractrecovery.HistoricalGraphProjectionRestoreSourceRegistryV2JSON)
	bindingBody := []byte(contractrecovery.HistoricalGraphProjectionRestoreImplementationBindingV2JSON)
	storage := &graphRestoreArtifactStorage{bodies: map[string][]byte{
		"graph-registry-v2": registryBody,
		"graph-binding-v2":  bindingBody,
	}}
	registry, err := loadVNextRecoveryGenerationRegistry()
	if err != nil {
		t.Fatalf("load Recovery generation registry: %v", err)
	}
	generation, admitted := registry.lookup(
		contractrecovery.RecoveryGenerations[2].CatalogDigestSHA256,
		contractrecovery.RecoveryGenerations[2].CodecRegistrySHA256,
	)
	if !admitted {
		t.Fatal("historical Graph v2 generation is not admitted")
	}
	service := &VNextRestoreService{storage: storage, generations: registry}
	proofs := map[string]VNextArtifactProof{
		"graph-registry-v2": {
			Kind: "graph_projection_restore_source_registry", SchemaID: GraphProjectionRestoreSourceRegistryV2SchemaID,
			LogicalRef: "graph-registry-v2", PlaintextBytes: int64(len(registryBody)),
			PlaintextSHA256: contractrecovery.HistoricalGraphProjectionRestoreSourceRegistryV2SHA256,
		},
		"graph-binding-v2": {
			Kind: "graph_projection_restore_implementation_binding", SchemaID: GraphProjectionRestoreImplementationBindingV2SchemaID,
			LogicalRef: "graph-binding-v2", PlaintextBytes: int64(len(bindingBody)),
			PlaintextSHA256: contractrecovery.HistoricalGraphProjectionRestoreImplementationBindingV2SHA256,
		},
	}
	resolved, err := service.resolveGraphProjectionRestoreArtifacts(context.Background(), generation, proofs)
	if err != nil || resolved.AlgorithmID != "graphprojection.restore_rebuild.v2" ||
		!bytes.Equal(resolved.SourceRegistryJSON, registryBody) || !bytes.Equal(resolved.ImplementationBindingJSON, bindingBody) {
		t.Fatalf("resolve exact historical Graph restore artifacts: result=%#v err=%v", resolved, err)
	}

	mixed := mapsCloneGraphRestoreProofs(proofs)
	bindingProof := mixed["graph-binding-v2"]
	bindingProof.SchemaID = GraphProjectionRestoreImplementationBindingV3SchemaID
	bindingProof.PlaintextSHA256 = contractrecovery.CurrentGraphProjectionRestoreImplementationBindingSHA256
	mixed["graph-binding-v2"] = bindingProof
	if _, err := service.resolveGraphProjectionRestoreArtifacts(context.Background(), generation, mixed); !errors.Is(err, ErrVNextBackup) {
		t.Fatalf("mixed historical/current Graph restore artifacts were admitted: %v", err)
	}
}

func TestVNextGraphRestoreArtifactsRetainExactPreWorkbookV3Dispatch_Unit(t *testing.T) {
	registry, err := loadVNextRecoveryGenerationRegistry()
	if err != nil {
		t.Fatalf("load Recovery generation registry: %v", err)
	}
	projected := contractrecovery.RecoveryGenerations[1]
	generation, admitted := registry.lookup(projected.CatalogDigestSHA256, projected.CodecRegistrySHA256)
	if !admitted {
		t.Fatal("pre-Workbook-ownership Graph v3 generation is not admitted")
	}
	storage := &graphRestoreArtifactStorage{bodies: map[string][]byte{
		"graph-registry-v3": []byte(projected.GraphSourceRegistryJSON),
		"graph-binding-v3":  []byte(projected.GraphImplementationBindingJSON),
	}}
	service := &VNextRestoreService{storage: storage, generations: registry}
	proofs := map[string]VNextArtifactProof{
		"graph-registry-v3": {
			Kind: "graph_projection_restore_source_registry", SchemaID: projected.GraphSourceRegistrySchemaID,
			LogicalRef: "graph-registry-v3", PlaintextBytes: int64(len(projected.GraphSourceRegistryJSON)),
			PlaintextSHA256: projected.GraphSourceRegistrySHA256,
		},
		"graph-binding-v3": {
			Kind: "graph_projection_restore_implementation_binding", SchemaID: projected.GraphImplementationBindingSchemaID,
			LogicalRef: "graph-binding-v3", PlaintextBytes: int64(len(projected.GraphImplementationBindingJSON)),
			PlaintextSHA256: projected.GraphImplementationBindingSHA256,
		},
	}
	resolved, err := service.resolveGraphProjectionRestoreArtifacts(context.Background(), generation, proofs)
	if err != nil || resolved.AlgorithmID != "graphprojection.restore_rebuild.v3" ||
		resolved.RecoveryStateCatalogSHA256 != projected.CatalogDigestSHA256 ||
		resolved.ImplementationBindingSHA256 != "6ec244d0b82466a18adbdb82554be29f5e4baac384175538acbc92e56f14b8d5" {
		t.Fatalf("resolve exact pre-Workbook Graph v3 artifacts: result=%#v err=%v", resolved, err)
	}
	mixed := mapsCloneGraphRestoreProofs(proofs)
	bindingProof := mixed["graph-binding-v3"]
	bindingProof.PlaintextSHA256 = contractrecovery.RecoveryGenerations[0].GraphImplementationBindingSHA256
	mixed["graph-binding-v3"] = bindingProof
	if _, err := service.resolveGraphProjectionRestoreArtifacts(context.Background(), generation, mixed); !errors.Is(err, ErrVNextBackup) {
		t.Fatalf("mixed pre-Workbook/current Graph v3 artifacts were admitted: %v", err)
	}
}

type graphRestoreArtifactStorage struct{ bodies map[string][]byte }

func (*graphRestoreArtifactStorage) WriteArtifact(context.Context, string, []byte, string) (BackupArtifactProof, error) {
	return BackupArtifactProof{}, errors.New("unexpected artifact write")
}

func (storage *graphRestoreArtifactStorage) ReadArtifact(_ context.Context, key string, _ int64) ([]byte, error) {
	body, ok := storage.bodies[key]
	if !ok {
		return nil, errors.New("artifact unavailable")
	}
	return append([]byte(nil), body...), nil
}

func (*graphRestoreArtifactStorage) WriteArtifactStream(context.Context, BackupArtifactStreamWriteRequest) (BackupArtifactStreamProof, error) {
	return BackupArtifactStreamProof{}, errors.New("unexpected artifact stream write")
}

func (storage *graphRestoreArtifactStorage) ReadArtifactStream(_ context.Context, proof BackupArtifactStreamProof, destination io.Writer) error {
	body, ok := storage.bodies[proof.LogicalRef]
	if !ok {
		return errors.New("artifact unavailable")
	}
	_, err := destination.Write(body)
	return err
}

func mapsCloneGraphRestoreProofs(source map[string]VNextArtifactProof) map[string]VNextArtifactProof {
	cloned := make(map[string]VNextArtifactProof, len(source))
	for key, value := range source {
		cloned[key] = value
	}
	return cloned
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
