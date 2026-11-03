package recovery_test

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
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
	capture := newCurrentMetadataCapture(t, store)
	catalog := testExtensionBackupCatalog(t)
	now := time.Date(2026, 7, 24, 15, 0, 0, 0, time.UTC)
	backupSet, err := capture.CaptureAndPublish(context.Background(), recovery.VNextCaptureParams{
		BackupSetID:        uuid.MustParse("00000000-0000-0000-0000-000000104001"),
		ConsistencyPointAt: now.Add(-time.Minute), CreatedAt: now,
		RetainedUntil: now.Add(31 * 24 * time.Hour),
	})
	if err != nil {
		t.Fatal(err)
	}
	proof, err := recovery.VNextProofFromMetadata(context.Background(), capture.Storage, backupSet.PostgresArtifactKey, "application/json", backupSet.PostgresArtifactSizeBytes, backupSet.PostgresArtifactSHA256)
	if err != nil {
		t.Fatal(err)
	}
	var body bytes.Buffer
	if err := capture.Storage.ReadArtifactStream(context.Background(), proof, &body); err != nil {
		t.Fatal(err)
	}
	var snapshot recovery.VNextPostgresSnapshotArtifact
	if err := json.Unmarshal(body.Bytes(), &snapshot); err != nil {
		t.Fatal(err)
	}
	if snapshot.SchemaID != recovery.PostgresSnapshotArtifactV3SchemaID {
		t.Fatalf("snapshot schema = %q", snapshot.SchemaID)
	}
	bindings := catalog.Bindings()
	if len(snapshot.ExtensionBindingProofs) != len(bindings) || len(bindings) != 5 {
		t.Fatalf("binding proof count = %d, catalog = %d", len(snapshot.ExtensionBindingProofs), len(bindings))
	}
	for index, proof := range snapshot.ExtensionBindingProofs {
		binding := bindings[index]
		if proof.SchemaID != recovery.PostgresBindingProofV1SchemaID ||
			proof.ProfileID != binding.ProfileID ||
			proof.ImplementationBindingSHA256 != binding.ImplementationBindingSHA256 ||
			proof.PhysicalStateBindingSHA256 != binding.PhysicalStateBindingSHA256 ||
			proof.BindingID != binding.BindingID ||
			proof.BackupCodecID != binding.CurrentCodec.CodecID ||
			proof.BackupCodecSHA256 != binding.CurrentCodec.CodecSHA256 ||
			proof.ItemCount != 0 || proof.ContentBytes != 0 || len(proof.BindingDigest) != 64 {
			t.Fatalf("binding proof %d does not match immutable current catalog: %#v", index, proof)
		}
	}
}

func TestRestoreRejectsNonemptyTargetBeforeMutation_Integration(t *testing.T) {
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
			tc.mutate(t, &fixture.Target)
			storage := &countingBackupStorage{Inner: fixture.BackupStorage}
			runner := recovery.NewVersionedRestoreRunner(fixture.Store, storage, testExtensionBackupCatalog(t), currentStateCatalog(t))
			observer := &restoreStepRecorder{}
			fixture.Target.Observer = observer
			before, err := recovery.ReadTargetControl(ctx, fixture.Target.Postgres)
			if err != nil {
				t.Fatal(err)
			}
			admitted, err := tryAdmitCurrentRestore(t, fixture.capturedBackupContractFixture, runner, fixture.Target, fixture.BackupSet)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("restore error got %v want %v", err, tc.wantErr)
			}
			if storage.Reads == 0 {
				t.Fatal("current selected input was not authenticated")
			}
			if admitted.Fence != nil || len(observer.Steps) != 0 {
				t.Fatal("nonempty target reached mutation")
			}
			after, err := recovery.ReadTargetControl(ctx, fixture.Target.Postgres)
			if err != nil || before != after {
				t.Fatal("rejected target control changed", err)
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
	retiredRunner := recovery.NewVersionedRestoreRunner(fixture.Store, unread, testExtensionBackupCatalog(t), currentStateCatalog(t))
	retiredAdmission, retiredErr := tryAdmitCurrentRestore(t, fixture.capturedBackupContractFixture, retiredRunner, fixture.Target, retired)
	if err := retiredAdmission.Admission.Release(ctx); err != nil {
		t.Fatal(err)
	}
	if err := retiredErr; !errors.Is(err, recovery.ErrInvalidBackupArtifact) {
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
			manifest.ManifestSHA256 = digestHex(append([]byte("CARTULARY-BACKUP-INTEGRITY-MANIFEST-V4\n"), preimage...))
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
			observer := &restoreStepRecorder{}
			target := fixture.Target
			target.Observer = observer
			runner := recovery.NewVersionedRestoreRunner(fixture.Store, fixture.BackupStorage, testExtensionBackupCatalog(t), currentStateCatalog(t))
			admitted, err := tryAdmitCurrentRestore(t, fixture.capturedBackupContractFixture, runner, target, backupSet)
			if !errors.Is(err, recovery.ErrVNextBackup) {
				t.Fatalf("restore error got %v want invalid catalog-driven backup", err)
			}
			if len(observer.Steps) != 0 || admitted.Fence != nil {
				t.Fatalf("invalid evidence reached mutation: %v", observer.Steps)
			}
			control, err := recovery.ReadTargetControl(ctx, target.Postgres)
			if err != nil || control.State != recovery.TargetInitialized {
				t.Fatal("rejected input changed target admission", err)
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

func digestHex(body []byte) string {
	sum := sha256.Sum256(body)
	return hex.EncodeToString(sum[:])
}

func (storage *countingBackupStorage) WriteArtifactStream(ctx context.Context, request recovery.BackupArtifactStreamWriteRequest) (recovery.BackupArtifactStreamProof, error) {
	inner, err := recovery.RequireStreamingBackupStorage(storage.Inner)
	if err != nil {
		return recovery.BackupArtifactStreamProof{}, err
	}
	return inner.WriteArtifactStream(ctx, request)
}

func (storage *countingBackupStorage) ReadArtifactStream(ctx context.Context, proof recovery.BackupArtifactStreamProof, destination io.Writer) error {
	storage.Reads++
	inner, err := recovery.RequireStreamingBackupStorage(storage.Inner)
	if err != nil {
		return err
	}
	return inner.ReadArtifactStream(ctx, proof, destination)
}

func (storage *countingBackupStorage) ResolveObjectProof(ctx context.Context, entry recovery.VNextObjectManifestEntry) (recovery.BackupArtifactStreamProof, error) {
	storage.Reads++
	resolver, ok := storage.Inner.(recovery.VNextObjectProofResolver)
	if !ok {
		return recovery.BackupArtifactStreamProof{}, recovery.ErrInvalidBackupArtifact
	}
	return resolver.ResolveObjectProof(ctx, entry)
}
