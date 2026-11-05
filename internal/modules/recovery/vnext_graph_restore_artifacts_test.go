package recovery

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"slices"
	"testing"
	"time"

	contractrecovery "github.com/JochiRaider/cartulary/internal/gen/contractrecovery"
	"github.com/JochiRaider/cartulary/internal/platform/recoverystate"
)

func TestVNextGraphRestoreCurrentProjectionContract_Unit(t *testing.T) {
	t.Parallel()
	if len(contractrecovery.RecoveryGenerations) != 1 {
		t.Fatalf("Recovery generations = %d, want one current generation", len(contractrecovery.RecoveryGenerations))
	}
	projected := contractrecovery.RecoveryGenerations[0]
	for _, artifact := range []struct {
		name, body, digest, selectedBody, selectedDigest string
	}{
		{"registry", contractrecovery.CurrentGraphProjectionRestoreSourceRegistryJSON, contractrecovery.CurrentGraphProjectionRestoreSourceRegistrySHA256, projected.GraphSourceRegistryJSON, projected.GraphSourceRegistrySHA256},
		{"binding", contractrecovery.CurrentGraphProjectionRestoreImplementationBindingJSON, contractrecovery.CurrentGraphProjectionRestoreImplementationBindingSHA256, projected.GraphImplementationBindingJSON, projected.GraphImplementationBindingSHA256},
	} {
		actual := fmt.Sprintf("%x", sha256.Sum256([]byte(artifact.body)))
		if actual != artifact.digest || actual != artifact.selectedDigest || artifact.body != artifact.selectedBody {
			t.Errorf("current Graph %s bytes, published digest and selected generation disagree", artifact.name)
		}
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
		t.Fatalf("decode generated Graph restore current registry: %v", err)
	}
	if registry.SchemaID != "cartulary.graph_projection_restore_source_registry.v6" || len(registry.Entries) != 1 {
		t.Fatalf("current Graph restore registry is not current v5: %#v", registry)
	}
	entry := registry.Entries[0]
	if entry.SourceOwnerID != "network_flow_activity" ||
		entry.AuthoritativeFamilyID != "network_flow_activity.graph_views" ||
		entry.ProjectionInputContractID != "graph_projection.v2" ||
		entry.ProjectionResultContractID != "graph_projection_result.v2" ||
		!slices.Equal(entry.SemanticQuerySchemaIDs, []string{"cartulary.network_flow.graph_semantic_query.v3"}) {
		t.Fatalf("current Graph restore registry entry drifted: %#v", entry)
	}

	var binding map[string]any
	if err := json.Unmarshal([]byte(contractrecovery.CurrentGraphProjectionRestoreImplementationBindingJSON), &binding); err != nil {
		t.Fatalf("decode generated Graph restore current binding: %v", err)
	}
	if binding["schema_id"] != "cartulary.graph_projection_restore_implementation_binding.v8" ||
		binding["algorithm_id"] != "graphprojection.restore_rebuild.v5" ||
		binding["binding_id"] != "graphprojection.restore_rebuild.network_flow_graph_views.v6" ||
		binding["projection_schema_id"] != "graph_projection.v2" ||
		binding["recovery_state_catalog_sha256"] != projected.CatalogDigestSHA256 ||
		binding["source_registry_sha256"] != projected.GraphSourceRegistrySHA256 ||
		projected.GraphSourceRegistrySchemaID != registry.SchemaID ||
		projected.GraphImplementationBindingSchemaID != binding["schema_id"] ||
		binding["historical_dispatch_algorithm_ids"] != nil {
		t.Fatalf("current Graph restore implementation binding drifted: %#v", binding)
	}
}

func TestVNextGraphRestoreArtifactsResolveOnlyCurrent_Unit(t *testing.T) {
	registry, err := loadVNextRecoveryGenerationRegistry()
	if err != nil {
		t.Fatalf("load Recovery generation registry: %v", err)
	}
	projected := contractrecovery.RecoveryGenerations[0]
	storage := &graphRestoreArtifactStorage{bodies: map[string][]byte{
		"graph-registry-current": []byte(projected.GraphSourceRegistryJSON),
		"graph-binding-current":  []byte(projected.GraphImplementationBindingJSON),
	}}
	service := &VNextRestoreService{storage: storage, generations: registry}
	proofs := currentGraphRestoreProofs(projected)
	resolved, err := service.resolveGraphProjectionRestoreArtifacts(context.Background(), registry.current, proofs)
	if err != nil || resolved.AlgorithmID != "graphprojection.restore_rebuild.v5" ||
		resolved.RecoveryStateCatalogSHA256 != projected.CatalogDigestSHA256 ||
		resolved.SourceRegistrySHA256 != projected.GraphSourceRegistrySHA256 ||
		resolved.ImplementationBindingSHA256 != projected.GraphImplementationBindingSHA256 {
		t.Fatalf("resolve exact current Graph restore artifacts: result=%#v err=%v", resolved, err)
	}
	if storage.readCalls != 2 {
		t.Fatalf("current Graph artifact reads = %d, want 2", storage.readCalls)
	}
}

func TestVNextGraphRestoreArtifactsRejectRetiredBeforeArtifactRead_Unit(t *testing.T) {
	registry, err := loadVNextRecoveryGenerationRegistry()
	if err != nil {
		t.Fatalf("load Recovery generation registry: %v", err)
	}
	projected := contractrecovery.RecoveryGenerations[0]
	for _, schemas := range retiredGraphRestoreSchemaPairs() {
		t.Run(schemas[0]+"-"+schemas[1], func(t *testing.T) {
			storage := &graphRestoreArtifactStorage{bodies: map[string][]byte{}}
			service := &VNextRestoreService{storage: storage, generations: registry}
			proofs := currentGraphRestoreProofs(projected)
			registryProof := proofs["graph-registry-current"]
			registryProof.SchemaID = "cartulary.graph_projection_restore_source_registry." + schemas[0]
			proofs["graph-registry-current"] = registryProof
			bindingProof := proofs["graph-binding-current"]
			bindingProof.SchemaID = "cartulary.graph_projection_restore_implementation_binding." + schemas[1]
			proofs["graph-binding-current"] = bindingProof
			if _, err := service.resolveGraphProjectionRestoreArtifacts(context.Background(), registry.current, proofs); !errors.Is(err, ErrVNextBackup) {
				t.Fatalf("retired Graph %v artifacts were admitted: %v", schemas, err)
			}
			if storage.readCalls != 0 {
				t.Fatalf("retired Graph %v artifacts caused %d reads before rejection", schemas, storage.readCalls)
			}
		})
	}
}

func TestVNextRestoreRejectsRetiredGraphBindingsBeforeTargetMutation_Unit(t *testing.T) {
	registry, err := loadVNextRecoveryGenerationRegistry()
	if err != nil {
		t.Fatalf("load Recovery generation registry: %v", err)
	}
	projected := contractrecovery.RecoveryGenerations[0]
	for _, schemas := range retiredGraphRestoreSchemaPairs() {
		t.Run(schemas[0]+"-"+schemas[1], func(t *testing.T) {
			proofs := currentGraphRestoreProofs(projected)
			registryProof := proofs["graph-registry-current"]
			registryProof.SchemaID = "cartulary.graph_projection_restore_source_registry." + schemas[0]
			proofs["graph-registry-current"] = registryProof
			bindingProof := proofs["graph-binding-current"]
			bindingProof.SchemaID = "cartulary.graph_projection_restore_implementation_binding." + schemas[1]
			proofs["graph-binding-current"] = bindingProof
			otherProof := VNextArtifactProof{
				Kind: "postgres_snapshot", SchemaID: PostgresSnapshotArtifactV3SchemaID,
				LogicalRef: "postgres", ContentType: "application/json", PlaintextBytes: 2,
			}
			at := time.Date(2026, 8, 29, 20, 0, 0, 0, time.UTC)
			manifest := VNextBackupIntegrityManifest{
				ApplicationCryptoFormat: ApplicationCryptoFormatID,
				SchemaID:                BackupIntegrityManifestV4SchemaID, BackupSetID: "00000000-0000-0000-0000-000000009401",
				ConsistencyPointAt: at, CreatedAt: at.Add(time.Minute), RetainedUntil: at.Add(30 * 24 * time.Hour),
				RecoveryStateCatalogSHA256: projected.CatalogDigestSHA256,
				CodecRegistrySHA256:        projected.CodecRegistrySHA256,
				PostgresSnapshotRef:        "postgres", ObjectStoreManifestRef: "postgres",
				Artifacts: []VNextArtifactProof{proofs["graph-registry-current"], proofs["graph-binding-current"], otherProof},
			}
			if err := assignSelfDigest(vNextIntegrityManifestDigestDomain, &manifest.ManifestSHA256, manifest); err != nil {
				t.Fatalf("digest old-Graph manifest: %v", err)
			}
			manifestBody, err := json.Marshal(manifest)
			if err != nil {
				t.Fatalf("encode old-Graph manifest: %v", err)
			}
			storage := &graphRestoreArtifactStorage{bodies: map[string][]byte{
				"integrity":              manifestBody,
				"graph-registry-current": []byte(projected.GraphSourceRegistryJSON),
				"graph-binding-current":  []byte(projected.GraphImplementationBindingJSON),
				"postgres":               []byte(`{}`),
			}}
			service := &VNextRestoreService{storage: storage, generations: registry}
			if _, err := service.validateIntegrityManifest(manifest); err != nil {
				t.Fatalf("negative fixture fails unrelated manifest admission: %v", err)
			}
			target := &recordingAtomicRestoreTarget{}
			if err := service.Restore(context.Background(), target, BackupArtifactStreamProof{
				LogicalRef: "integrity", ContentType: "application/json", PlaintextBytes: int64(len(manifestBody)),
			}); !errors.Is(err, ErrVNextBackup) {
				t.Fatalf("retired Graph %v restore error = %v, want ErrVNextBackup", schemas, err)
			}
			if storage.readCalls != 4 {
				t.Fatalf("retired Graph fixture did not reach artifact preflight: %d reads, want 4", storage.readCalls)
			}
			if target.atomicCalls != 0 {
				t.Fatalf("retired Graph %v restore entered target mutation %d times", schemas, target.atomicCalls)
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
	projected := contractrecovery.RecoveryGenerations[0]
	for _, ref := range []string{"graph-registry-current", "graph-binding-current"} {
		for _, corruptProof := range []bool{false, true} {
			t.Run(fmt.Sprintf("%s/proof=%t", ref, corruptProof), func(t *testing.T) {
				storage := &graphRestoreArtifactStorage{bodies: map[string][]byte{
					"graph-registry-current": []byte(projected.GraphSourceRegistryJSON),
					"graph-binding-current":  []byte(projected.GraphImplementationBindingJSON),
				}}
				proofs := currentGraphRestoreProofs(projected)
				if corruptProof {
					proof := proofs[ref]
					proof.PlaintextSHA256 = "unsupported"
					proofs[ref] = proof
				} else {
					// Still valid JSON, but not the generation's exact canonical bytes.
					storage.bodies[ref] = append(storage.bodies[ref], ' ')
				}
				service := &VNextRestoreService{storage: storage, generations: registry}
				if _, err := service.resolveGraphProjectionRestoreArtifacts(context.Background(), registry.current, proofs); !errors.Is(err, ErrVNextBackup) {
					t.Fatalf("corrupted current Graph artifact admitted: %v", err)
				}
				if corruptProof && storage.readCalls != 0 {
					t.Fatal("unsupported digest caused artifact reads")
				}
			})
		}
	}
}

// Vary each retired identity independently while keeping its counterpart current.
func retiredGraphRestoreSchemaPairs() [][2]string {
	return [][2]string{
		{"v2", "v7"}, {"v3", "v7"}, {"v4", "v7"},
		{"v5", "v2"}, {"v5", "v3"}, {"v5", "v4"}, {"v5", "v5"}, {"v5", "v6"},
	}
}

func currentGraphRestoreProofs(projected contractrecovery.RecoveryGeneration) map[string]VNextArtifactProof {
	return map[string]VNextArtifactProof{
		"graph-registry-current": {
			Kind: "graph_projection_restore_source_registry", SchemaID: projected.GraphSourceRegistrySchemaID,
			LogicalRef: "graph-registry-current", PlaintextBytes: int64(len(projected.GraphSourceRegistryJSON)),
			PlaintextSHA256: projected.GraphSourceRegistrySHA256,
		},
		"graph-binding-current": {
			Kind: "graph_projection_restore_implementation_binding", SchemaID: projected.GraphImplementationBindingSchemaID,
			LogicalRef: "graph-binding-current", PlaintextBytes: int64(len(projected.GraphImplementationBindingJSON)),
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
