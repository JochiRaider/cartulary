package incidentportabilityassembly_test

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/app/incidentportabilityassembly"
	"github.com/JochiRaider/cartulary/internal/platform/rootedfs"
)

func TestIncidentBundleRootStorageEnforcesReferencesAndLifecycle_Unit(t *testing.T) {
	temporaryRoot := filepath.Join(t.TempDir(), "temporary")
	exportRoot := filepath.Join(t.TempDir(), "exports")
	for _, root := range []string{temporaryRoot, exportRoot} {
		if err := os.Mkdir(root, 0o700); err != nil {
			t.Fatal(err)
		}
	}
	storage, err := incidentportabilityassembly.NewRootStorage(temporaryRoot, exportRoot)
	if err != nil {
		t.Fatalf("NewRootStorage: %v", err)
	}
	t.Cleanup(storage.Close)

	staged, err := storage.Stage(context.Background(), strings.Repeat("a", 64), []byte("staged bundle"))
	if err != nil {
		t.Fatalf("Stage: %v", err)
	}
	if filepath.IsAbs(staged.String()) || strings.Contains(staged.String(), temporaryRoot) {
		t.Fatalf("Stage returned host path %q", staged.String())
	}
	stagedPath := filepath.Join(temporaryRoot, filepath.FromSlash(staged.String()))
	assertPrivateRegularFile(t, stagedPath, []byte("staged bundle"))
	data, err := storage.ReadStaged(staged, 1024)
	if err != nil {
		t.Fatalf("ReadStaged: %v", err)
	}
	if string(data) != "staged bundle" {
		t.Fatalf("ReadStaged = %q", data)
	}
	if err := storage.RemoveStaged(staged); err != nil {
		t.Fatalf("RemoveStaged: %v", err)
	}
	if err := storage.RemoveStaged(staged); err != nil {
		t.Fatalf("RemoveStaged idempotent retry: %v", err)
	}

	bundleID := "22222222-2222-4222-8222-222222222222"
	published, err := storage.Publish(context.Background(), bundleID, []byte("published bundle"))
	if err != nil {
		t.Fatalf("Publish: %v", err)
	}
	if filepath.IsAbs(published.String()) || strings.Contains(published.String(), exportRoot) {
		t.Fatalf("Publish returned host path %q", published.String())
	}
	publishedPath := filepath.Join(exportRoot, filepath.FromSlash(published.String()))
	assertPrivateRegularFile(t, publishedPath, []byte("published bundle"))
	if _, err := storage.Publish(context.Background(), bundleID, []byte("replacement")); !errors.Is(err, os.ErrExist) {
		t.Fatalf("duplicate Publish error = %v; want exclusive destination failure", err)
	}
	assertPrivateRegularFile(t, publishedPath, []byte("published bundle"))
	t.Run("recovery preserves confined published bytes", func(t *testing.T) {
		source, err := incidentportabilityassembly.NewRecoveryStorage(exportRoot)
		if err != nil {
			t.Fatal(err)
		}
		defer source.Close()
		targetRoot := t.TempDir()
		target, err := incidentportabilityassembly.NewRecoveryStorage(targetRoot)
		if err != nil {
			t.Fatal(err)
		}
		defer target.Close()
		ctx := context.Background()
		if err := target.RequireEmpty(ctx); err != nil {
			t.Fatal(err)
		}
		info, err := source.StatRecoveryObject(ctx, published.String())
		if err != nil || info.PlaintextBytes != int64(len("published bundle")) {
			t.Fatal(info, err)
		}
		hash := sha256.Sum256([]byte("published bundle"))
		digest := hex.EncodeToString(hash[:])
		for _, invalid := range []string{"../escape", "/outside", "evidence/blobs/test", "incident-bundles/not-a-uuid.zip"} {
			if target.RestoreMember(ctx, invalid, digest, info.PlaintextBytes, strings.NewReader("published bundle")) == nil {
				t.Fatal("unsafe reference admitted", invalid)
			}
		}
		if target.RestoreMember(ctx, published.String(), strings.Repeat("0", 64), info.PlaintextBytes, strings.NewReader("published bundle")) == nil {
			t.Fatal("corrupt bytes admitted")
		}
		if err := target.RequireEmpty(ctx); err != nil {
			t.Fatal("failed restore left partial bytes", err)
		}
		reader, err := source.OpenRecoveryObject(ctx, published.String())
		if err != nil {
			t.Fatal(err)
		}
		err = target.RestoreMember(ctx, published.String(), digest, info.PlaintextBytes, reader)
		reader.Close()
		if err != nil {
			t.Fatal(err)
		}
		if target.RequireEmpty(ctx) == nil {
			t.Fatal("populated root admitted as empty")
		}
		restored, err := target.OpenRecoveryObject(ctx, published.String())
		if err != nil {
			t.Fatal(err)
		}
		body, err := io.ReadAll(restored)
		restored.Close()
		if err != nil || string(body) != "published bundle" {
			t.Fatal("restore changed bytes", err)
		}
		if err := target.ResetVerificationTarget(ctx); err != nil {
			t.Fatal(err)
		}
		if err := target.RequireEmpty(ctx); err != nil {
			t.Fatal(err)
		}
		assertPrivateRegularFile(t, publishedPath, []byte("published bundle"))
	})
}

func TestIncidentBundleRootStorageFailsClosedOnCancellationSymlinkAndRootReplacement_Unit(t *testing.T) {
	t.Run("cancellation leaves no partial publication", func(t *testing.T) {
		temporaryRoot := filepath.Join(t.TempDir(), "temporary")
		exportRoot := filepath.Join(t.TempDir(), "exports")
		for _, root := range []string{temporaryRoot, exportRoot} {
			if err := os.Mkdir(root, 0o700); err != nil {
				t.Fatal(err)
			}
		}
		storage, err := incidentportabilityassembly.NewRootStorage(temporaryRoot, exportRoot)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(storage.Close)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if _, err := storage.Stage(ctx, strings.Repeat("b", 64), []byte("must not publish")); err == nil {
			t.Fatal("canceled Stage succeeded")
		}
		entries, err := os.ReadDir(filepath.Join(temporaryRoot, "incident-bundles", "imports"))
		if err != nil {
			t.Fatal(err)
		}
		if len(entries) != 0 {
			t.Fatalf("canceled Stage retained files: %#v", entries)
		}
	})

	t.Run("symlinked component", func(t *testing.T) {
		temporaryRoot := filepath.Join(t.TempDir(), "temporary")
		exportRoot := filepath.Join(t.TempDir(), "exports")
		outside := t.TempDir()
		for _, root := range []string{temporaryRoot, exportRoot} {
			if err := os.Mkdir(root, 0o700); err != nil {
				t.Fatal(err)
			}
		}
		if err := os.Mkdir(filepath.Join(temporaryRoot, "incident-bundles"), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(outside, filepath.Join(temporaryRoot, "incident-bundles", "imports")); err != nil {
			t.Fatal(err)
		}
		storage, err := incidentportabilityassembly.NewRootStorage(temporaryRoot, exportRoot)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(storage.Close)
		if _, err := storage.Stage(context.Background(), strings.Repeat("c", 64), []byte("escape")); err == nil {
			t.Fatal("Stage followed a symlinked component")
		} else if strings.Contains(err.Error(), temporaryRoot) || strings.Contains(err.Error(), outside) {
			t.Fatalf("Stage error disclosed host root: %v", err)
		}
		entries, err := os.ReadDir(outside)
		if err != nil {
			t.Fatal(err)
		}
		if len(entries) != 0 {
			t.Fatalf("symlink escape wrote outside root: %#v", entries)
		}
	})

	t.Run("root replacement", func(t *testing.T) {
		base := t.TempDir()
		temporaryRoot := filepath.Join(base, "temporary")
		originalRoot := filepath.Join(base, "temporary-original")
		exportRoot := filepath.Join(base, "exports")
		for _, root := range []string{temporaryRoot, exportRoot} {
			if err := os.Mkdir(root, 0o700); err != nil {
				t.Fatal(err)
			}
		}
		storage, err := incidentportabilityassembly.NewRootStorage(temporaryRoot, exportRoot)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(storage.Close)
		if err := os.Rename(temporaryRoot, originalRoot); err != nil {
			t.Fatal(err)
		}
		if err := os.Mkdir(temporaryRoot, 0o700); err != nil {
			t.Fatal(err)
		}
		if _, err := storage.Stage(context.Background(), strings.Repeat("d", 64), []byte("replacement")); !errors.Is(err, rootedfs.ErrRootIdentityChanged) {
			t.Fatalf("Stage after root replacement error = %v; want ErrRootIdentityChanged", err)
		}
		entries, err := os.ReadDir(temporaryRoot)
		if err != nil {
			t.Fatal(err)
		}
		if len(entries) != 0 {
			t.Fatalf("root replacement received writes: %#v", entries)
		}
	})
}

func assertPrivateRegularFile(t testing.TB, path string, want []byte) {
	t.Helper()
	info, err := os.Lstat(path)
	if err != nil {
		t.Fatal(err)
	}
	if !info.Mode().IsRegular() || info.Mode().Perm() != 0o600 {
		t.Fatalf("%s mode = %v; want private regular file", path, info.Mode())
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != string(want) {
		t.Fatalf("%s bytes = %q want %q", path, data, want)
	}
}
