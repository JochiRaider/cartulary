package recovery_test

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/app/extensionassembly"
	"github.com/JochiRaider/cartulary/internal/app/recoveryassembly"
	"github.com/JochiRaider/cartulary/internal/modules/recovery"
	"github.com/JochiRaider/cartulary/internal/platform/recoverystate"
)

func exerciseTransferClosure(t *testing.T, storage recovery.BackupStorage, key recovery.RecoveryEncryptionKey, state *recoverystate.Catalog, captured recovery.VNextCapturedBackup) {
	t.Helper()
	ctx := context.Background()
	extensions, err := extensionassembly.GeneratedRecoveryCatalog()
	if err != nil {
		t.Fatal(err)
	}
	catalog := recovery.NewBackupCatalog(nil, storage, extensions, state)
	at := captured.IntegrityManifest.ConsistencyPointAt
	backup := recovery.BackupSet{BackupSetID: captured.BackupSetID, ConsistencyPointAt: at, CreatedAt: captured.IntegrityManifest.CreatedAt, RetainedUntil: captured.IntegrityManifest.RetainedUntil,
		PostgresArtifactKey: recovery.VNextMetadataArtifactKey(captured.PostgresProof.LogicalRef), PostgresArtifactSHA256: captured.PostgresProof.PlaintextSHA256, PostgresArtifactSizeBytes: captured.PostgresProof.PlaintextBytes,
		ObjectStoreArtifactKey: recovery.VNextMetadataArtifactKey(captured.ObjectManifestProof.LogicalRef), ObjectStoreArtifactSHA256: captured.ObjectManifestProof.PlaintextSHA256, ObjectStoreArtifactSizeBytes: captured.ObjectManifestProof.PlaintextBytes,
		IntegrityManifestKey: recovery.VNextMetadataArtifactKey(captured.IntegrityProof.LogicalRef), IntegrityManifestSHA256: captured.IntegrityProof.PlaintextSHA256, IntegrityManifestSizeBytes: captured.IntegrityProof.PlaintextBytes}
	output := filepath.Join(t.TempDir(), "export")
	destination, err := recoveryassembly.NewTransferDirectory(output)
	if err != nil {
		t.Fatal(err)
	}
	defer destination.Close()
	release := strings.Repeat("a", 64)
	manifest, err := catalog.ExportTransfer(ctx, backup, release, at.Add(25*time.Hour), destination)
	if err != nil {
		t.Fatal("export", err)
	}
	if len(manifest.Artifacts) != len(captured.IntegrityManifest.Artifacts)+1+len(state.Document().ObjectFamilies) {
		t.Fatal("incomplete closure")
	}
	if _, err := os.Stat(output); !errors.Is(err, os.ErrNotExist) {
		t.Fatal("published before commit", err)
	}
	if err := destination.Publish(ctx); err != nil {
		t.Fatal(err)
	}
	if _, err := recoveryassembly.NewTransferDirectory(output); err == nil {
		t.Fatal("occupied output admitted")
	}
	reader, err := recoveryassembly.OpenTransferDirectory(output)
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()
	admitted, err := catalog.ValidateTransfer(ctx, reader, key, release)
	if err != nil || admitted.Selection.BackupSetID != backup.BackupSetID {
		t.Fatal("readback", err)
	}
	if _, err := catalog.ValidateTransfer(ctx, reader, key, strings.Repeat("b", 64)); !errors.Is(err, recovery.ErrTransferReleaseMismatch) {
		t.Fatal("wrong release admitted", err)
	}
	wrongKey, err := recovery.ParseRecoveryEncryptionKey("YWJjZGVmZ2hpamtsbW5vcHFyc3R1dnd4eXoxMjM0NTY=")
	if err != nil {
		t.Fatal(err)
	}
	if _, err := catalog.ValidateTransfer(ctx, reader, wrongKey, release); err == nil {
		t.Fatal("wrong key admitted")
	}
	cancelled, cancel := context.WithCancel(ctx)
	cancel()
	if _, err := catalog.ValidateTransfer(cancelled, reader, key, release); err == nil {
		t.Fatal("cancelled read admitted")
	}
	capturedDirectory, err := recoveryassembly.NewTransferDirectory(filepath.Join(t.TempDir(), "captured"))
	if err != nil {
		t.Fatal(err)
	}
	defer capturedDirectory.Close()
	frozen, transferDigest, err := catalog.CaptureTransfer(ctx, struct{ recovery.TransferSource }{reader}, capturedDirectory, key, release)
	if err != nil || frozen.Selection.BackupSetID != backup.BackupSetID || len(transferDigest) != 64 {
		t.Fatal("capture transfer", err)
	}
	var objectReference string
	for _, artifact := range manifest.Artifacts {
		if strings.Contains(artifact.Reference, "/vnext/objects/") {
			objectReference = artifact.Reference
			break
		}
	}
	if objectReference == "" {
		t.Fatal("object payload missing from transfer closure")
	}
	member := filepath.Join(output, filepath.FromSlash(objectReference))
	original, err := os.ReadFile(member)
	if err != nil {
		t.Fatal(err)
	}
	for _, body := range [][]byte{original[:len(original)-1], append([]byte("x"), original[1:]...)} {
		if err := os.WriteFile(member, body, 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := catalog.ValidateTransfer(ctx, reader, key, release); err == nil {
			t.Fatal("changed artifact admitted")
		}
	}
	if _, err := catalog.ValidateTransfer(ctx, capturedDirectory, key, release); err != nil {
		t.Fatal("external input mutation changed captured bytes", err)
	}
	// Execute the existing restore algorithm against captured bytes while the
	// original exported payload is corrupt; no live source capability is supplied.
	capturedEncrypted, err := recovery.NewEncryptedBackupStorage(capturedDirectory, key)
	if err != nil {
		t.Fatal(err)
	}
	capturedStreaming, err := recovery.RequireStreamingBackupStorage(capturedEncrypted)
	if err != nil {
		t.Fatal(err)
	}
	algorithms, err := recovery.NewVNextRestoreAlgorithmCatalog(state, recovery.RequiredVNextRestoreAlgorithmIDs(state)...)
	if err != nil {
		t.Fatal(err)
	}
	restorer, err := recovery.NewVNextRestoreService(capturedStreaming, state, algorithms)
	if err != nil {
		t.Fatal(err)
	}
	target := &vNextRestoreTargetFake{}
	if err := restorer.Restore(ctx, target, captured.IntegrityProof); err != nil {
		t.Fatal("captured closure restore", err)
	}
	if len(target.objects) != len(state.Document().ObjectFamilies) {
		t.Fatal("captured restore omitted owner object")
	}

	if err := os.WriteFile(member, original, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.Remove(member); err != nil {
		t.Fatal(err)
	}
	if _, err := catalog.ValidateTransfer(ctx, reader, key, release); err == nil {
		t.Fatal("missing artifact admitted")
	}
	if err := os.WriteFile(member, original, 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(output, "extra"), []byte("extra"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := catalog.ValidateTransfer(ctx, reader, key, release); err == nil {
		t.Fatal("extra artifact admitted")
	}
	if err := os.Remove(filepath.Join(output, "extra")); err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(filepath.Join(output, "extra-directory"), 0700); err != nil {
		t.Fatal(err)
	}
	if _, err := catalog.ValidateTransfer(ctx, reader, key, release); err == nil {
		t.Fatal("extra empty directory admitted")
	}
	if err := os.Remove(filepath.Join(output, "extra-directory")); err != nil {
		t.Fatal(err)
	}
	manifestPath := filepath.Join(output, recovery.TransferManifestName)
	sealed, err := os.ReadFile(manifestPath)
	if err != nil {
		t.Fatal(err)
	}
	for _, body := range [][]byte{sealed[:len(sealed)-1], []byte(strings.Repeat("x", recovery.TransferManifestMaximumBytes+65537))} {
		if err := os.WriteFile(manifestPath, body, 0600); err != nil {
			t.Fatal(err)
		}
		if _, err := catalog.ValidateTransfer(ctx, reader, key, release); err == nil {
			t.Fatal("invalid or oversized manifest admitted")
		}
	}
	if err := os.WriteFile(manifestPath, sealed, 0600); err != nil {
		t.Fatal(err)
	}
	// Authentication and release admission do not consult source retention time.
	if _, err := catalog.ValidateTransfer(ctx, reader, key, release); err != nil {
		t.Fatal(err)
	}
	abandoned := filepath.Join(t.TempDir(), "abandoned")
	stage, err := recoveryassembly.NewTransferDirectory(abandoned)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := stage.WriteArtifact(ctx, "partial", []byte("encrypted"), "application/octet-stream"); err != nil {
		t.Fatal(err)
	}
	if err := stage.Close(); err != nil {
		t.Fatal(err)
	}
	entries, err := os.ReadDir(filepath.Dir(abandoned))
	if err != nil || len(entries) != 0 {
		t.Fatal("unpublished cleanup", err)
	}
	occupied := filepath.Join(t.TempDir(), "occupied-after-stage")
	race, err := recoveryassembly.NewTransferDirectory(occupied)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Mkdir(occupied, 0700); err != nil {
		t.Fatal(err)
	}
	if err := race.Publish(ctx); err == nil {
		t.Fatal("publication overwrote competing destination")
	}
	if err := race.Close(); err != nil {
		t.Fatal(err)
	}
	if _, err := os.Stat(occupied); err != nil {
		t.Fatal("cleanup removed competing destination", err)
	}

}
