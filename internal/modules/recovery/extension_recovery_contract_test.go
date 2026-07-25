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
	if len(manifest.ExtensionBindings) != len(bindings) || len(bindings) != 4 {
		t.Fatalf("extension binding proof count got %d catalog=%d want 4", len(manifest.ExtensionBindings), len(bindings))
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

func TestRestoreRejectsRunningOrNonemptyTargetBeforeArtifactRead_Integration(t *testing.T) {
	ctx := context.Background()
	tests := []struct {
		name    string
		wantErr error
		mutate  func(t *testing.T, target *recovery.RestoreTarget)
	}{
		{
			name:    "running",
			wantErr: recovery.ErrRestoreTargetNotStopped,
			mutate: func(_ *testing.T, target *recovery.RestoreTarget) {
				target.Stopped = false
			},
		},
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
UPDATE extension_state_metadata
   SET state_version = 2, updated_at = updated_at + interval '1 second'
 WHERE profile_id = 'network_flow_activity'
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
			_, err := recovery.NewRestoreRunner(fixture.Store, storage, testExtensionBackupCatalog(t)).
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
	original, err := fixture.BackupStorage.ReadArtifact(ctx, fixture.BackupSet.IntegrityManifestKey, fixture.BackupSet.IntegrityManifestSizeBytes)
	if err != nil {
		t.Fatalf("read fixture manifest: %v", err)
	}
	base, err := recovery.DecodeIntegrityManifest(original)
	if err != nil {
		t.Fatalf("decode fixture manifest: %v", err)
	}
	tests := []struct {
		name    string
		mutate  func(*recovery.BackupIntegrityManifest)
		wantErr error
	}{
		{
			name: "legacy v1 manifest",
			mutate: func(manifest *recovery.BackupIntegrityManifest) {
				manifest.SchemaID = "cartulary.backup_integrity_manifest.v1"
			},
			wantErr: recovery.ErrInvalidBackupArtifact,
		},
		{
			name: "missing extension proofs",
			mutate: func(manifest *recovery.BackupIntegrityManifest) {
				manifest.ExtensionBindings = nil
			},
			wantErr: recovery.ErrExtensionBindingInvalid,
		},
		{
			name: "unpackaged codec digest",
			mutate: func(manifest *recovery.BackupIntegrityManifest) {
				manifest.ExtensionBindings[0].CodecSHA256 = strings.Repeat("f", 64)
			},
			wantErr: recovery.ErrExtensionCodecUnsupported,
		},
		{
			name: "implementation binding mismatch",
			mutate: func(manifest *recovery.BackupIntegrityManifest) {
				manifest.ExtensionBindings[0].ImplementationBindingSHA256 = strings.Repeat("e", 64)
			},
			wantErr: recovery.ErrExtensionBindingInvalid,
		},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			manifest := base
			manifest.ExtensionBindings = append([]recovery.ExtensionBindingProof(nil), base.ExtensionBindings...)
			tc.mutate(&manifest)
			body, err := json.Marshal(manifest)
			if err != nil {
				t.Fatalf("encode rewritten manifest: %v", err)
			}
			backupSet := fixture.BackupSet
			backupSet.IntegrityManifestSHA256 = digestHex(body)
			backupSet.IntegrityManifestSizeBytes = int64(len(body))
			storage := &replacementBackupStorage{
				Inner: fixture.BackupStorage,
				Key:   backupSet.IntegrityManifestKey,
				Body:  body,
			}
			readiness := &recordingRestoreReadinessGate{}
			failure := &recordingRestoreFailureGate{}
			observer := &restoreStepRecorder{}
			target := fixture.Target
			target.Readiness = readiness
			target.Failure = failure
			target.Observer = observer
			_, err = recovery.NewRestoreRunner(fixture.Store, storage, testExtensionBackupCatalog(t)).
				RestoreBackupSet(ctx, target, backupSet)
			if !errors.Is(err, tc.wantErr) {
				t.Fatalf("restore error got %v want %v", err, tc.wantErr)
			}
			if len(observer.Steps) != 0 || readiness.Calls != 0 {
				t.Fatalf("invalid extension evidence mutated target: steps=%v readiness=%d", observer.Steps, readiness.Calls)
			}
			if len(failure.Causes) != 1 || !errors.Is(failure.Causes[0], tc.wantErr) {
				t.Fatalf("failed target gate got %#v want one %v", failure.Causes, tc.wantErr)
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

type replacementBackupStorage struct {
	Inner recovery.BackupStorage
	Key   string
	Body  []byte
}

func (storage *replacementBackupStorage) WriteArtifact(ctx context.Context, key string, body []byte, contentType string) (recovery.BackupArtifactProof, error) {
	return storage.Inner.WriteArtifact(ctx, key, body, contentType)
}

func (storage *replacementBackupStorage) ReadArtifact(ctx context.Context, key string, maxBytes int64) ([]byte, error) {
	if key == storage.Key {
		return append([]byte(nil), storage.Body...), nil
	}
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
