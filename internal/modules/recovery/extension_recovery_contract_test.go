package recovery_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"

	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/testutil/pgtest"
)

func TestExtensionBackupManifestRecordsCanonicalBindingProofs_Integration(t *testing.T) {
	db := pgtest.Start(t).BeginRollbackDBT(t, "backup_restore-i-10-04-extension-manifest")
	store := recovery.NewStore(db)
	storage := newEncryptedBackupStorage(t, t.TempDir())
	catalog := testExtensionBackupCatalog(t)
	now := time.Date(2026, 7, 24, 15, 0, 0, 0, time.UTC)
	backupSet, err := recovery.NewCaptureService(store, storage, catalog).CaptureBackupSet(context.Background(), captureParams(recovery.CaptureBackupSetParams{
		BackupSetID:        uuid.MustParse("00000000-0000-0000-0000-000000104001"),
		ConsistencyPointAt: now.Add(-time.Minute),
		CreatedAt:          now,
		RetainedUntil:      now.Add(31 * 24 * time.Hour),
	}))
	if err != nil {
		t.Fatalf("capture extension-aware backup: %v", err)
	}
	body, err := recovery.VerifyArtifactProof(context.Background(), storage, recovery.BackupArtifactProof{
		Key: backupSet.IntegrityManifestKey, SHA256: backupSet.IntegrityManifestSHA256,
		SizeBytes: backupSet.IntegrityManifestSizeBytes,
	})
	if err != nil {
		t.Fatalf("read integrity manifest: %v", err)
	}
	manifest, err := recovery.DecodeIntegrityManifest(body)
	if err != nil {
		t.Fatalf("decode integrity manifest: %v", err)
	}
	if manifest.SchemaID != recovery.BackupIntegrityManifestSchemaID {
		t.Fatalf("manifest schema got %q want %q", manifest.SchemaID, recovery.BackupIntegrityManifestSchemaID)
	}
	bindings := catalog.Bindings()
	if len(manifest.ExtensionBindings) != len(bindings) || len(bindings) != 5 {
		t.Fatalf("extension binding proof count got %d catalog=%d want 5", len(manifest.ExtensionBindings), len(bindings))
	}
	for index, proof := range manifest.ExtensionBindings {
		binding := bindings[index]
		if proof.ProfileID != binding.ProfileID ||
			proof.ImplementationBindingSHA256 != binding.ImplementationBindingSHA256 ||
			proof.PhysicalBindingSHA256 != binding.PhysicalStateBindingSHA256 ||
			proof.BindingID != binding.BindingID ||
			proof.CodecID != binding.CurrentCodec.CodecID ||
			proof.CodecSHA256 != binding.CurrentCodec.CodecSHA256 ||
			proof.ItemCount != 0 ||
			proof.ContentByteLength != 0 ||
			len(proof.ContentSHA256) != 64 {
			t.Fatalf("binding proof %d does not match immutable catalog: proof=%#v binding=%#v", index, proof, binding)
		}
	}
}

func TestRestoreRejectsNonemptyTargetBeforeArtifactRead_Integration(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name    string
		wantErr error
		mutate  func(t *testing.T, target *recovery.RestoreTarget)
	}{
		{
			name:    "authoritative row",
			wantErr: recovery.ErrRestoreTargetNotEmpty,
			mutate: func(t *testing.T, target *recovery.RestoreTarget) {
				t.Helper()
				if _, err := target.Postgres.Exec(ctx, `
INSERT INTO users (email, display_name, password_hash, mfa_required, is_active, is_deployment_admin)
VALUES ('restore-target@example.test', 'Restore Target', 'not-a-real-hash', false, true, false)
`); err != nil {
					t.Fatalf("seed nonempty restore target: %v", err)
				}
			},
		},
		{
			name:    "altered extension metadata",
			wantErr: recovery.ErrRestoreTargetNotEmpty,
			mutate: func(t *testing.T, target *recovery.RestoreTarget) {
				t.Helper()
				if _, err := target.Postgres.Exec(ctx, `
INSERT INTO extension_state_metadata (
    profile_id, migration_lineage_id, state_version, last_migration_id,
    metadata_version, created_at, updated_at
) VALUES (
    'network_flow_activity', 'network_flow_activity.state_v1', 2, NULL,
    1, statement_timestamp(), statement_timestamp()
)
`); err != nil {
					t.Fatalf("alter extension metadata: %v", err)
				}
			},
		},
		{
			name:    "object",
			wantErr: recovery.ErrRestoreTargetNotEmpty,
			mutate: func(t *testing.T, target *recovery.RestoreTarget) {
				t.Helper()
				body := []byte("nonempty")
				if err := target.ObjectStore.PutObject(ctx, "restore-target/nonempty", bytes.NewReader(body), int64(len(body)), "text/plain"); err != nil {
					t.Fatalf("seed nonempty target object store: %v", err)
				}
			},
		},
	}
	for index, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			fixture := newRestoreProjectionContractFixture(t, ctx,
				"backup_restore-i-10-05-"+strings.ReplaceAll(tc.name, " ", "-"),
				uuid.MustParse("00000000-0000-0000-0000-00000010410"+string(rune('1'+index))),
			)
			fixture.Target.Projections = &recordingProjectionRebuilder{}
			failure := &recordingRestoreFailureGate{}
			fixture.Target.Failure = failure
			tc.mutate(t, &fixture.Target)
			storage := &countingBackupStorage{Inner: fixture.BackupStorage}
			_, err := recovery.NewVersionedRestoreRunner(fixture.Store, storage, testExtensionBackupCatalog(t), currentStateCatalog(t)).
				RestoreBackupSet(ctx, fixture.Target, fixture.BackupSet)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("restore error got %v want %v", err, tc.wantErr)
			}
			if storage.Reads != 0 {
				t.Fatalf("preflight failure read %d backup artifacts", storage.Reads)
			}
			if len(failure.Causes) != 1 || !errors.Is(failure.Causes[0], tc.wantErr) {
				t.Fatalf("failed target gate got %#v want one %v", failure.Causes, tc.wantErr)
			}
		})
	}
}

func TestRestoreRejectsLegacyOrInvalidExtensionBindingEvidenceBeforeMutation_Integration(t *testing.T) {
	ctx := context.Background()
	fixture := newRestoreProjectionContractFixture(t, ctx, "backup_restore-i-10-06-extension-proof", uuid.MustParse("00000000-0000-0000-0000-000000104201"))
	fixture.Target.Projections = &recordingProjectionRebuilder{}
	retired := fixture.BackupSet
	retired.IntegrityManifestKey = "backup_sets/retired/integrity-manifest.json"
	unread := &countingBackupStorage{Inner: fixture.BackupStorage}
	if _, err := recovery.NewVersionedRestoreRunner(fixture.Store, unread, testExtensionBackupCatalog(t), currentStateCatalog(t)).RestoreBackupSet(ctx, fixture.Target, retired); !errors.Is(err, recovery.ErrInvalidBackupArtifact) {
		t.Fatalf("retired restore representation: %v", err)
	}
	if err := recovery.NewBackupCatalog(fixture.Store, unread, testExtensionBackupCatalog(t), currentStateCatalog(t)).VerifyBackupSetDurability(ctx, retired); !errors.Is(err, recovery.ErrInvalidBackupArtifact) {
		t.Fatalf("retired catalog representation: %v", err)
	}
	if unread.Reads != 0 {
		t.Fatal("retired representation read storage")
	}
	streaming, err := recovery.RequireStreamingBackupStorage(fixture.BackupStorage)
	if err != nil {
		t.Fatal(err)
	}
	proof, err := recovery.VNextProofFromMetadata(ctx, streaming, fixture.BackupSet.IntegrityManifestKey, "application/json", fixture.BackupSet.IntegrityManifestSizeBytes, fixture.BackupSet.IntegrityManifestSHA256)
	if err != nil {
		t.Fatal(err)
	}
	var original bytes.Buffer
	if err := streaming.ReadArtifactStream(ctx, proof, &original); err != nil {
		t.Fatal(err)
	}
	var base recovery.VNextBackupIntegrityManifest
	if err := json.Unmarshal(original.Bytes(), &base); err != nil {
		t.Fatal(err)
	}
	tests := []struct {
		name   string
		mutate func(*recovery.VNextBackupIntegrityManifest)
	}{
		{
			name: "legacy v1 manifest",
			mutate: func(manifest *recovery.VNextBackupIntegrityManifest) {
				manifest.SchemaID = "cartulary.backup_integrity_manifest.v1"
			},
		},
		{
			name: "missing owner catalog binding",
			mutate: func(manifest *recovery.VNextBackupIntegrityManifest) {
				manifest.RecoveryStateCatalogSHA256 = ""
			},
		},
		{
			name: "unpackaged codec digest",
			mutate: func(manifest *recovery.VNextBackupIntegrityManifest) {
				manifest.CodecRegistrySHA256 = strings.Repeat("f", 64)
			},
		},
		{
			name: "implementation binding mismatch",
			mutate: func(manifest *recovery.VNextBackupIntegrityManifest) {
				for index := range manifest.Artifacts {
					if manifest.Artifacts[index].Kind == "graph_projection_restore_implementation_binding" {
						manifest.Artifacts[index].PlaintextSHA256 = strings.Repeat("e", 64)
						return
					}
				}
				t.Fatal("missing implementation binding in producer fixture")
			},
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			manifest := base
			manifest.Artifacts = append([]recovery.VNextArtifactProof(nil), base.Artifacts...)
			tc.mutate(&manifest)
			manifest.ManifestSHA256 = ""
			preimage, err := json.Marshal(manifest)
			if err != nil {
				t.Fatal(err)
			}
			manifest.ManifestSHA256 = digestHex(append([]byte("CARTULARY-BACKUP-INTEGRITY-MANIFEST-V3\n"), preimage...))
			body, err := json.Marshal(manifest)
			if err != nil {
				t.Fatalf("encode rewritten manifest: %v", err)
			}
			backupSet := fixture.BackupSet
			logical := "backup_sets/" + backupSet.BackupSetID.String() + "/tampered/" + uuid.NewString() + ".json"
			// Give the modified manifest a valid encrypted storage envelope and
			// self digest, so rejection proves semantic binding validation.
			changed, err := streaming.WriteArtifactStream(ctx, recovery.BackupArtifactStreamWriteRequest{LogicalRef: logical, EnvelopeRef: logical + ".envelope.json", ContentType: "application/json", Plaintext: bytes.NewReader(body)})
			if err != nil {
				t.Fatal(err)
			}
			backupSet.IntegrityManifestKey = recovery.VNextMetadataArtifactKey(logical)
			backupSet.IntegrityManifestSHA256 = changed.PlaintextSHA256
			backupSet.IntegrityManifestSizeBytes = changed.PlaintextBytes
			readiness := &recordingRestoreReadinessGate{}
			failure := &recordingRestoreFailureGate{}
			observer := &restoreStepRecorder{}
			target := fixture.Target
			target.Readiness = readiness
			target.Failure = failure
			target.Observer = observer
			_, err = recovery.NewVersionedRestoreRunner(fixture.Store, fixture.BackupStorage, testExtensionBackupCatalog(t), currentStateCatalog(t)).
				RestoreBackupSet(ctx, target, backupSet)
			if !errors.Is(err, recovery.ErrVNextBackup) {
				t.Fatalf("restore error got %v want invalid catalog-driven backup", err)
			}
			if len(observer.Steps) != 0 || readiness.Calls != 0 {
				t.Fatalf("invalid extension evidence mutated target: steps=%v readiness=%d", observer.Steps, readiness.Calls)
			}
			if len(failure.Causes) != 1 || !errors.Is(failure.Causes[0], recovery.ErrVNextBackup) {
				t.Fatalf("failed target gate got %#v want one invalid-backup failure", failure.Causes)
			}
		})
	}
}

type countingBackupStorage struct {
	Inner recovery.BackupStorage
	Reads int
}

func (storage *countingBackupStorage) WriteArtifact(ctx context.Context, key string, body []byte, contentType string) (recovery.BackupArtifactProof, error) {
	return storage.Inner.WriteArtifact(ctx, key, body, contentType)
}

func (storage *countingBackupStorage) ReadArtifact(ctx context.Context, key string, maxBytes int64) ([]byte, error) {
	storage.Reads++
	return storage.Inner.ReadArtifact(ctx, key, maxBytes)
}

type recordingRestoreFailureGate struct {
	Causes []error
}

func (gate *recordingRestoreFailureGate) MarkRestoreFailed(_ context.Context, cause error) {
	gate.Causes = append(gate.Causes, cause)
}

func digestHex(body []byte) string {
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}
