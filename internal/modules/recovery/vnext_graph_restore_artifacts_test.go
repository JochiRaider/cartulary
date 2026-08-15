package recovery

import (
	"context"
	"errors"
	"testing"

	contractrecovery "github.com/JochiRaider/cartulary/internal/gen/contractrecovery"
)

func TestVNextGraphRestoreArtifactsFailClosedAndSelectExactLegacyBinding_Unit(t *testing.T) {
	service := &VNextRestoreService{}
	current := VNextBackupIntegrityManifest{CodecRegistrySHA256: VNextCodecRegistrySHA256()}
	_, err := service.resolveGraphProjectionRestoreArtifacts(context.Background(), current, map[string]VNextArtifactProof{
		"registry": {
			Kind: "graph_projection_restore_source_registry", SchemaID: GraphProjectionRestoreSourceRegistryV1SchemaID,
			LogicalRef: "registry", PlaintextSHA256: contractrecovery.CurrentGraphProjectionRestoreSourceRegistrySHA256,
		},
	})
	if err == nil || !errors.Is(err, ErrVNextBackup) {
		t.Fatalf("current backup without implementation binding did not fail closed: %v", err)
	}

	legacy := VNextBackupIntegrityManifest{CodecRegistrySHA256: LegacyEmptyRegistryVNextCodecRegistrySHA256()}
	artifacts, err := service.resolveGraphProjectionRestoreArtifacts(context.Background(), legacy, map[string]VNextArtifactProof{})
	if err != nil {
		t.Fatalf("select exact historical empty-registry binding: %v", err)
	}
	if !artifacts.LegacyEmptyRegistryBinding ||
		artifacts.SourceRegistrySHA256 != contractrecovery.CurrentGraphProjectionRestoreSourceRegistrySHA256 ||
		artifacts.ImplementationBindingSHA256 != contractrecovery.LegacyGraphProjectionRestoreImplementationBindingSHA256 ||
		artifacts.ImplementationBindingSHA256 == contractrecovery.CurrentGraphProjectionRestoreImplementationBindingSHA256 ||
		VNextCodecRegistrySHA256() == LegacyEmptyRegistryVNextCodecRegistrySHA256() {
		t.Fatalf("legacy Graph Projection binding selection is not exact and distinct: %#v", artifacts)
	}
	_, err = service.resolveGraphProjectionRestoreArtifacts(context.Background(), legacy, map[string]VNextArtifactProof{
		"forged": {Kind: "graph_projection_restore_implementation_binding", SchemaID: GraphProjectionRestoreImplementationBindingV1SchemaID, LogicalRef: "forged"},
	})
	if err == nil {
		t.Fatal("historical codec digest admitted a supplied or mismatched Graph binding")
	}
}
