package recovery

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"slices"
	"testing"
	"time"

	contractrecovery "github.com/JochiRaider/cartulary/internal/gen/contractrecovery"
	"github.com/JochiRaider/cartulary/internal/platform/recoverystate"
)

func TestVNextGraphRestoreV4ProjectionContract_Unit(t *testing.T) {
	t.Parallel()
	if len(contractrecovery.RecoveryGenerations) != 1 {
		t.Fatalf("Recovery generations = %d, want one current generation", len(contractrecovery.RecoveryGenerations))
	}
	if got, want := contractrecovery.CurrentGraphProjectionRestoreImplementationBindingSHA256, "2a04b36c624970358a52fda65efd9b0f1ab1398cbb0a049b06e77ac9b7f84ac7"; got != want {
		t.Fatalf("current Graph v4 binding digest = %s, want %s", got, want)
	}
	if got, want := contractrecovery.CurrentGraphProjectionRestoreSourceRegistrySHA256, "a18774fbb30712823a95c90f43517ca19484f37f3e7f685cfe75401eaec6b634"; got != want {
		t.Fatalf("current Graph v4 source registry digest = %s, want %s", got, want)
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
		t.Fatalf("decode generated Graph restore v4 registry: %v", err)
	}
	if registry.SchemaID != "cartulary.graph_projection_restore_source_registry.v4" || len(registry.Entries) != 1 {
		t.Fatalf("current Graph restore registry is not exact v4: %#v", registry)
	}
	entry := registry.Entries[0]
	if entry.SourceOwnerID != "network_flow_activity" ||
		entry.AuthoritativeFamilyID != "network_flow_activity.graph_views" ||
		entry.ProjectionInputContractID != "graph_projection.v2" ||
		entry.ProjectionResultContractID != "graph_projection_result.v2" ||
		!slices.Equal(entry.SemanticQuerySchemaIDs, []string{"cartulary.network_flow.graph_semantic_query.v2"}) {
		t.Fatalf("current Graph restore registry entry drifted: %#v", entry)
	}

	var binding map[string]any
	if err := json.Unmarshal([]byte(contractrecovery.CurrentGraphProjectionRestoreImplementationBindingJSON), &binding); err != nil {
		t.Fatalf("decode generated Graph restore v4 binding: %v", err)
	}
	if binding["schema_id"] != "cartulary.graph_projection_restore_implementation_binding.v4" ||
		binding["algorithm_id"] != "graphprojection.restore_rebuild.v4" ||
		binding["historical_dispatch_algorithm_ids"] != nil {
		t.Fatalf("current Graph restore implementation binding drifted: %#v", binding)
	}
}

func TestVNextGraphRestoreArtifactsResolveOnlyCurrentV4_Unit(t *testing.T) {
	registry, err := loadVNextRecoveryGenerationRegistry()
	if err != nil {
		t.Fatalf("load Recovery generation registry: %v", err)
	}
	projected := contractrecovery.RecoveryGenerations[0]
	storage := &graphRestoreArtifactStorage{bodies: map[string][]byte{
		"graph-registry-v4": []byte(projected.GraphSourceRegistryJSON),
		"graph-binding-v4":  []byte(projected.GraphImplementationBindingJSON),
	}}
	service := &VNextRestoreService{storage: storage, generations: registry}
	proofs := currentGraphRestoreProofs(projected)
	resolved, err := service.resolveGraphProjectionRestoreArtifacts(context.Background(), registry.current, proofs)
	if err != nil || resolved.AlgorithmID != "graphprojection.restore_rebuild.v4" ||
		resolved.RecoveryStateCatalogSHA256 != projected.CatalogDigestSHA256 ||
		resolved.SourceRegistrySHA256 != projected.GraphSourceRegistrySHA256 ||
		resolved.ImplementationBindingSHA256 != projected.GraphImplementationBindingSHA256 {
		t.Fatalf("resolve exact current Graph restore artifacts: result=%#v err=%v", resolved, err)
	}
	if storage.readCalls != 2 {
		t.Fatalf("current Graph artifact reads = %d, want 2", storage.readCalls)
	}
}

func TestVNextGraphRestoreArtifactsRejectV2AndV3BeforeArtifactRead_Unit(t *testing.T) {
	registry, err := loadVNextRecoveryGenerationRegistry()
	if err != nil {
		t.Fatalf("load Recovery generation registry: %v", err)
	}
	projected := contractrecovery.RecoveryGenerations[0]
	for _, version := range []string{"v2", "v3"} {
		t.Run(version, func(t *testing.T) {
			storage := &graphRestoreArtifactStorage{bodies: map[string][]byte{}}
			service := &VNextRestoreService{storage: storage, generations: registry}
			proofs := currentGraphRestoreProofs(projected)
			registryProof := proofs["graph-registry-v4"]
			registryProof.SchemaID = "cartulary.graph_projection_restore_source_registry." + version
			proofs["graph-registry-v4"] = registryProof
			bindingProof := proofs["graph-binding-v4"]
			bindingProof.SchemaID = "cartulary.graph_projection_restore_implementation_binding." + version
			proofs["graph-binding-v4"] = bindingProof
			if _, err := service.resolveGraphProjectionRestoreArtifacts(context.Background(), registry.current, proofs); !errors.Is(err, ErrVNextBackup) {
				t.Fatalf("old Graph %s artifacts were admitted: %v", version, err)
			}
			if storage.readCalls != 0 {
				t.Fatalf("old Graph %s artifacts caused %d reads before rejection", version, storage.readCalls)
			}
		})
	}
}

func TestVNextRestoreRejectsV2AndV3GraphBindingsBeforeTargetMutation_Unit(t *testing.T) {
	registry, err := loadVNextRecoveryGenerationRegistry()
	if err != nil {
		t.Fatalf("load Recovery generation registry: %v", err)
	}
	projected := contractrecovery.RecoveryGenerations[0]
	for _, version := range []string{"v2", "v3"} {
		t.Run(version, func(t *testing.T) {
			proofs := currentGraphRestoreProofs(projected)
			registryProof := proofs["graph-registry-v4"]
			registryProof.SchemaID = "cartulary.graph_projection_restore_source_registry." + version
			proofs["graph-registry-v4"] = registryProof
			bindingProof := proofs["graph-binding-v4"]
			bindingProof.SchemaID = "cartulary.graph_projection_restore_implementation_binding." + version
			proofs["graph-binding-v4"] = bindingProof
			otherProof := VNextArtifactProof{
				Kind: "postgres_snapshot", SchemaID: PostgresSnapshotArtifactV2SchemaID,
				LogicalRef: "postgres", ContentType: "application/json", PlaintextBytes: 2,
			}
			at := time.Date(2026, 8, 29, 20, 0, 0, 0, time.UTC)
			manifest := VNextBackupIntegrityManifest{
				SchemaID: BackupIntegrityManifestV3SchemaID, BackupSetID: "00000000-0000-0000-0000-000000009401",
				ConsistencyPointAt: at, CreatedAt: at.Add(time.Minute), RetainedUntil: at.Add(30 * 24 * time.Hour),
				RecoveryStateCatalogSHA256: projected.CatalogDigestSHA256,
				CodecRegistrySHA256:        projected.CodecRegistrySHA256,
				PostgresSnapshotRef:        "postgres", ObjectStoreManifestRef: "postgres",
				Artifacts: []VNextArtifactProof{proofs["graph-registry-v4"], proofs["graph-binding-v4"], otherProof},
			}
			if err := assignSelfDigest(vNextIntegrityManifestDigestDomain, &manifest.ManifestSHA256, manifest); err != nil {
				t.Fatalf("digest old-Graph manifest: %v", err)
			}
			manifestBody, err := json.Marshal(manifest)
			if err != nil {
				t.Fatalf("encode old-Graph manifest: %v", err)
			}
			storage := &graphRestoreArtifactStorage{bodies: map[string][]byte{
				"integrity":         manifestBody,
				"graph-registry-v4": []byte(projected.GraphSourceRegistryJSON),
				"graph-binding-v4":  []byte(projected.GraphImplementationBindingJSON),
				"postgres":          []byte(`{}`),
			}}
			service := &VNextRestoreService{storage: storage, generations: registry}
			target := &recordingAtomicRestoreTarget{}
			if err := service.Restore(context.Background(), target, BackupArtifactStreamProof{
				LogicalRef: "integrity", ContentType: "application/json", PlaintextBytes: int64(len(manifestBody)),
			}); !errors.Is(err, ErrVNextBackup) {
				t.Fatalf("old Graph %s restore error = %v, want ErrVNextBackup", version, err)
			}
			if target.atomicCalls != 0 {
				t.Fatalf("old Graph %s restore entered target mutation %d times", version, target.atomicCalls)
			}
		})
	}
}

func TestVNextGraphRestoreArtifactsFailClosedWhenMissingOrUnselected_Unit(t *testing.T) {
	registry, err := loadVNextRecoveryGenerationRegistry()
	if err != nil {
		t.Fatalf("load Recovery generation registry: %v", err)
	}
	service := &VNextRestoreService{generations: registry}
	if _, err := service.resolveGraphProjectionRestoreArtifacts(context.Background(), registry.current, map[string]VNextArtifactProof{}); !errors.Is(err, ErrVNextBackup) {
		t.Fatalf("missing Graph restore artifacts were admitted: %v", err)
	}
	if _, err := service.resolveGraphProjectionRestoreArtifacts(context.Background(), nil, currentGraphRestoreProofs(contractrecovery.RecoveryGenerations[0])); !errors.Is(err, ErrVNextBackup) {
		t.Fatalf("unselected Graph restore generation was admitted: %v", err)
	}
}

func currentGraphRestoreProofs(projected contractrecovery.RecoveryGeneration) map[string]VNextArtifactProof {
	return map[string]VNextArtifactProof{
		"graph-registry-v4": {
			Kind: "graph_projection_restore_source_registry", SchemaID: projected.GraphSourceRegistrySchemaID,
			LogicalRef: "graph-registry-v4", PlaintextBytes: int64(len(projected.GraphSourceRegistryJSON)),
			PlaintextSHA256: projected.GraphSourceRegistrySHA256,
		},
		"graph-binding-v4": {
			Kind: "graph_projection_restore_implementation_binding", SchemaID: projected.GraphImplementationBindingSchemaID,
			LogicalRef: "graph-binding-v4", PlaintextBytes: int64(len(projected.GraphImplementationBindingJSON)),
			PlaintextSHA256: projected.GraphImplementationBindingSHA256,
		},
	}
}

type graphRestoreArtifactStorage struct {
	bodies    map[string][]byte
	readCalls int
}

type recordingAtomicRestoreTarget struct{ atomicCalls int }

func (target *recordingAtomicRestoreTarget) WithAtomicRestore(
	context.Context,
	*recoverystate.Catalog,
	func(VNextRestoreMutation) error,
) error {
	target.atomicCalls++
	return errors.New("unexpected target mutation")
}

func (*graphRestoreArtifactStorage) WriteArtifact(context.Context, string, []byte, string) (BackupArtifactProof, error) {
	return BackupArtifactProof{}, errors.New("unexpected artifact write")
}

func (storage *graphRestoreArtifactStorage) ReadArtifact(_ context.Context, key string, _ int64) ([]byte, error) {
	storage.readCalls++
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
	storage.readCalls++
	body, ok := storage.bodies[proof.LogicalRef]
	if !ok {
		return errors.New("artifact unavailable")
	}
	_, err := destination.Write(body)
	return err
}
