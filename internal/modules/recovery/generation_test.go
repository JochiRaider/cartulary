package recovery

import (
	"errors"
	"slices"
	"strings"
	"testing"
	"time"

	contractrecovery "github.com/JochiRaider/cartulary/internal/gen/contractrecovery"
)

func TestRecoveryGenerationRegistryOwnsThreeExactClosedGenerations_Unit(t *testing.T) {
	registry, err := loadVNextRecoveryGenerationRegistry()
	if err != nil {
		t.Fatalf("load Recovery generation registry: %v", err)
	}
	wantIDs := []string{
		"recovery.current.workbook_owned.graph_v3",
		"recovery.historical.incidents_owned.graph_v3",
		"recovery.historical.graph_v2",
	}
	gotIDs := make([]string, 0, len(contractrecovery.RecoveryGenerations))
	for _, generation := range contractrecovery.RecoveryGenerations {
		gotIDs = append(gotIDs, generation.GenerationID)
		selected, admitted := registry.lookup(generation.CatalogDigestSHA256, generation.CodecRegistrySHA256)
		if !admitted || selected.id != generation.GenerationID ||
			selected.stateCatalog.DigestSHA256() != generation.CatalogDigestSHA256 {
			t.Fatalf("generation %s did not round trip through its exact lookup pair", generation.GenerationID)
		}
	}
	if !slices.Equal(gotIDs, wantIDs) || registry.current.id != wantIDs[0] {
		t.Fatalf("Recovery generation order/current selection = %#v current=%s", gotIDs, registry.current.id)
	}
	if _, admitted := registry.lookup(
		contractrecovery.RecoveryGenerations[0].CatalogDigestSHA256,
		contractrecovery.RecoveryGenerations[2].CodecRegistrySHA256,
	); admitted {
		t.Fatal("cross-generation catalog/codec pair was admitted")
	}
	if _, admitted := registry.lookup("aaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaaa", "bbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbbb"); admitted {
		t.Fatal("unknown Recovery generation pair was admitted")
	}
}

func TestRecoveryGenerationSelectionDrivesCatalogCodecAndGraphValidation_Unit(t *testing.T) {
	registry, err := loadVNextRecoveryGenerationRegistry()
	if err != nil {
		t.Fatalf("load Recovery generation registry: %v", err)
	}
	service := &VNextRestoreService{generations: registry}
	for _, projected := range contractrecovery.RecoveryGenerations {
		t.Run(projected.GenerationID, func(t *testing.T) {
			generation, admitted := registry.lookup(projected.CatalogDigestSHA256, projected.CodecRegistrySHA256)
			if !admitted {
				t.Fatal("exact generated pair was not admitted")
			}
			integrity := VNextBackupIntegrityManifest{
				SchemaID: BackupIntegrityManifestV3SchemaID, BackupSetID: "00000000-0000-0000-0000-000000000901",
				ConsistencyPointAt:         time.Date(2026, 8, 19, 6, 0, 0, 0, time.UTC),
				CreatedAt:                  time.Date(2026, 8, 19, 6, 1, 0, 0, time.UTC),
				RetainedUntil:              time.Date(2026, 9, 18, 6, 1, 0, 0, time.UTC),
				RecoveryStateCatalogSHA256: projected.CatalogDigestSHA256,
				CodecRegistrySHA256:        projected.CodecRegistrySHA256,
				PostgresSnapshotRef:        "postgres", ObjectStoreManifestRef: "objects",
				Artifacts: []VNextArtifactProof{{LogicalRef: "postgres"}, {LogicalRef: "objects"}, {LogicalRef: "graph"}},
			}
			if err := assignSelfDigest(vNextIntegrityManifestDigestDomain, &integrity.ManifestSHA256, integrity); err != nil {
				t.Fatalf("digest integrity manifest: %v", err)
			}
			selected, err := service.validateIntegrityManifest(integrity)
			if err != nil || selected.id != generation.id {
				t.Fatalf("select integrity generation: selected=%v err=%v", selected, err)
			}

			tables := generation.stateCatalog.RequiredTableNames()
			proofs := make(map[string]VNextArtifactProof, len(tables))
			units := make([]VNextPostgresUnitDescriptor, 0, len(tables))
			for index, tableName := range tables {
				logicalRef := "table-" + tableName
				units = append(units, VNextPostgresUnitDescriptor{
					TableName: tableName, CodecID: PostgresSnapshotUnitV1SchemaID,
					LogicalRef: logicalRef, RowCount: int64(index), PlaintextSHA256: strings.Repeat("a", 64),
				})
				proofs[logicalRef] = VNextArtifactProof{
					SchemaID: PostgresSnapshotUnitV1SchemaID, LogicalRef: logicalRef,
					PlaintextSHA256: strings.Repeat("a", 64),
				}
			}
			postgres := VNextPostgresSnapshotArtifact{
				SchemaID: PostgresSnapshotArtifactV2SchemaID, BackupSetID: integrity.BackupSetID,
				ConsistencyPointAt:         integrity.ConsistencyPointAt,
				RecoveryStateCatalogSHA256: generation.stateCatalog.DigestSHA256(),
				TransactionIsolation:       VNextTransactionIsolation, Units: units,
			}
			if err := assignSelfDigest(vNextPostgresSnapshotDigestDomain, &postgres.SnapshotSHA256, postgres); err != nil {
				t.Fatalf("digest Postgres artifact: %v", err)
			}
			if err := service.validatePostgresArtifact(integrity, generation, postgres, proofs); err != nil {
				t.Fatalf("validate generation Postgres order: %v", err)
			}

			objects := VNextObjectStoreBackupManifest{
				SchemaID: ObjectStoreBackupManifestV2SchemaID, BackupSetID: integrity.BackupSetID,
				ConsistencyPointAt:         integrity.ConsistencyPointAt,
				RecoveryStateCatalogSHA256: generation.stateCatalog.DigestSHA256(),
				Objects:                    []VNextObjectManifestEntry{},
			}
			if err := assignSelfDigest(vNextObjectManifestDigestDomain, &objects.ManifestSHA256, objects); err != nil {
				t.Fatalf("digest object manifest: %v", err)
			}
			if err := service.validateObjectManifest(integrity, generation, objects); err != nil {
				t.Fatalf("validate generation object families: %v", err)
			}
			if !slices.Contains(RequiredVNextRestoreAlgorithmIDs(generation.stateCatalog), generation.graph.AlgorithmID) {
				t.Fatalf("selected generation catalog omits Graph algorithm %s", generation.graph.AlgorithmID)
			}
			if !generation.identity().AdmitsGraphCompletion(
				projected.CatalogDigestSHA256,
				projected.GraphSourceRegistrySHA256,
				projected.GraphImplementationBindingSHA256,
			) {
				t.Fatal("selected generation replay identity was rejected")
			}
		})
	}
	if registry.current.identity().AdmitsGraphCompletion(
		contractrecovery.RecoveryGenerations[1].CatalogDigestSHA256,
		contractrecovery.RecoveryGenerations[1].GraphSourceRegistrySHA256,
		contractrecovery.RecoveryGenerations[0].GraphImplementationBindingSHA256,
	) {
		t.Fatal("cross-generation replay identity was admitted")
	}
}

func TestRecoveryGenerationSelectionDrivesVerificationBasisAndCadence_Unit(t *testing.T) {
	registry, err := loadVNextRecoveryGenerationRegistry()
	if err != nil {
		t.Fatalf("load Recovery generation registry: %v", err)
	}
	base := RestoreVerificationBasis{
		MechanismID:                VNextBackupMechanismID,
		DatabaseBindingSHA256:      strings.Repeat("a", 64),
		ObjectStoreBindingSHA256:   strings.Repeat("b", 64),
		BackupStorageBindingSHA256: strings.Repeat("c", 64),
		RecoveryStateCatalogSHA256: registry.current.stateCatalog.DigestSHA256(),
		CodecRegistrySHA256:        registry.current.codecRegistrySHA256,
	}
	asOf := time.Date(2026, 8, 19, 7, 0, 0, 0, time.UTC)
	recentlyVerifiedAt := asOf.Add(-24 * time.Hour)
	for _, projected := range contractrecovery.RecoveryGenerations {
		t.Run(projected.GenerationID, func(t *testing.T) {
			generation, admitted := registry.lookup(projected.CatalogDigestSHA256, projected.CodecRegistrySHA256)
			if !admitted {
				t.Fatal("exact generated pair was not admitted")
			}
			basis := restoreVerificationBasisForGeneration(base, generation.identity())
			basisSHA256, err := basis.SHA256()
			if err != nil {
				t.Fatalf("generation verification basis: %v", err)
			}
			backupSet := BackupSet{
				LastVerifiedRestoreAt:       &recentlyVerifiedAt,
				LastVerificationBasisSHA256: basisSHA256,
			}
			if backupDueForRestoreVerification(backupSet, asOf, basisSHA256) {
				t.Fatal("recent exact-generation verification was incorrectly due")
			}
			currentBasisSHA256, err := base.SHA256()
			if err != nil {
				t.Fatalf("current verification basis: %v", err)
			}
			if projected.GenerationID != registry.current.id &&
				!backupDueForRestoreVerification(backupSet, asOf, currentBasisSHA256) {
				t.Fatal("historical generation did not distinguish the current basis")
			}
			staleVerifiedAt := asOf.Add(-restoreVerificationMaximumAge)
			backupSet.LastVerifiedRestoreAt = &staleVerifiedAt
			if !backupDueForRestoreVerification(backupSet, asOf, basisSHA256) {
				t.Fatal("seven-day-old exact-generation verification was not due")
			}
		})
	}
}

func TestRecoveryGenerationRegistryRejectsMalformedDuplicateAndMixedEntries_Unit(t *testing.T) {
	tests := []struct {
		name   string
		mutate func([]contractrecovery.RecoveryGeneration)
	}{
		{
			name: "duplicate generation ID",
			mutate: func(values []contractrecovery.RecoveryGeneration) {
				values[1].GenerationID = values[0].GenerationID
			},
		},
		{
			name: "duplicate lookup pair",
			mutate: func(values []contractrecovery.RecoveryGeneration) {
				values[1] = cloneRecoveryGeneration(values[0])
				values[1].GenerationID = "recovery.historical.duplicate_pair"
				values[1].CaptureCurrent = false
			},
		},
		{
			name: "multiple current entries",
			mutate: func(values []contractrecovery.RecoveryGeneration) {
				values[1].CaptureCurrent = true
			},
		},
		{
			name: "no current entry",
			mutate: func(values []contractrecovery.RecoveryGeneration) {
				values[0].CaptureCurrent = false
			},
		},
		{
			name: "malformed frozen catalog",
			mutate: func(values []contractrecovery.RecoveryGeneration) {
				values[2].CatalogJSON = `{"schema_id":"cartulary.recovery_state_catalog.v1"}`
				values[2].CatalogCanonicalSHA256 = digestBytes([]byte(values[2].CatalogJSON))
			},
		},
		{
			name: "unsorted codec IDs",
			mutate: func(values []contractrecovery.RecoveryGeneration) {
				values[0].CodecSchemaIDs[0], values[0].CodecSchemaIDs[1] = values[0].CodecSchemaIDs[1], values[0].CodecSchemaIDs[0]
			},
		},
		{
			name: "mixed Graph binding",
			mutate: func(values []contractrecovery.RecoveryGeneration) {
				values[1].GraphImplementationBindingJSON = values[0].GraphImplementationBindingJSON
				values[1].GraphImplementationBindingSHA256 = values[0].GraphImplementationBindingSHA256
				values[1].GraphImplementationBindingSchemaID = values[0].GraphImplementationBindingSchemaID
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			values := cloneRecoveryGenerations(contractrecovery.RecoveryGenerations)
			test.mutate(values)
			if _, err := loadVNextRecoveryGenerationRegistryFrom(values); !errors.Is(err, ErrVNextBackup) {
				t.Fatalf("load malformed Recovery generation registry error = %v, want ErrVNextBackup", err)
			}
		})
	}
}

func cloneRecoveryGenerations(values []contractrecovery.RecoveryGeneration) []contractrecovery.RecoveryGeneration {
	cloned := make([]contractrecovery.RecoveryGeneration, len(values))
	for index, value := range values {
		cloned[index] = cloneRecoveryGeneration(value)
	}
	return cloned
}

func cloneRecoveryGeneration(value contractrecovery.RecoveryGeneration) contractrecovery.RecoveryGeneration {
	value.CodecSchemaIDs = append([]string(nil), value.CodecSchemaIDs...)
	return value
}
