package recoveryassembly

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"golang.org/x/sys/unix"

	"github.com/JochiRaider/cartulary/internal/platform/rootedfs"
)

func TestRecoveryFilesystemStorageContainment_Unit(t *testing.T) {
	t.Run("publishes private exclusive root-free artifacts and bounded reads", func(t *testing.T) {
		rootPath := filepath.Join(t.TempDir(), "backups")
		storage, err := NewFilesystemStorage(rootPath)
		if err != nil {
			t.Fatalf("NewFilesystemStorage: %v", err)
		}
		defer storage.Close()
		proof, err := storage.WriteArtifact(
			context.Background(),
			"backup_sets/00000000-0000-0000-0000-000000000001/proof.json",
			[]byte(`{"proof":true}`),
			"application/json",
		)
		if err != nil {
			t.Fatalf("WriteArtifact: %v", err)
		}
		if filepath.IsAbs(proof.Key) || strings.Contains(proof.Key, rootPath) {
			t.Fatalf("artifact proof disclosed host root: %#v", proof)
		}
		if _, err := storage.WriteArtifact(context.Background(), proof.Key, []byte("replacement"), "text/plain"); err == nil {
			t.Fatal("duplicate recovery artifact publication succeeded")
		}
		body, err := storage.ReadArtifact(context.Background(), proof.Key, proof.SizeBytes)
		if err != nil || string(body) != `{"proof":true}` {
			t.Fatalf("ReadArtifact = %q, %v", body, err)
		}
		if _, err := storage.ReadArtifact(context.Background(), proof.Key, proof.SizeBytes-1); err == nil {
			t.Fatal("undersized recovery read bound succeeded")
		}
		requireMode(t, filepath.Join(rootPath, filepath.FromSlash(proof.Key)), 0o600)
		requireMode(t, filepath.Dir(filepath.Join(rootPath, filepath.FromSlash(proof.Key))), 0o700)
	})

	t.Run("rejects hostile references symlinks hard links and special objects", func(t *testing.T) {
		rootPath := filepath.Join(t.TempDir(), "backups")
		storage, err := NewFilesystemStorage(rootPath)
		if err != nil {
			t.Fatalf("NewFilesystemStorage: %v", err)
		}
		defer storage.Close()
		for _, key := range []string{"../escape", "/absolute", `bad\path`, "bad//path", "bad/./path", "cafe\u0301/path", "bad/\x00path"} {
			if _, err := storage.WriteArtifact(context.Background(), key, []byte("unsafe"), "text/plain"); err == nil {
				t.Fatalf("hostile recovery key %q was accepted", key)
			}
		}
		if err := os.Mkdir(filepath.Join(rootPath, "linked-parent-target"), 0o700); err != nil {
			t.Fatalf("create target directory: %v", err)
		}
		if err := os.Symlink(filepath.Join(rootPath, "linked-parent-target"), filepath.Join(rootPath, "linked")); err != nil {
			t.Fatalf("create parent symlink: %v", err)
		}
		if _, err := storage.WriteArtifact(context.Background(), "linked/escape", []byte("unsafe"), "text/plain"); err == nil {
			t.Fatal("recovery write through a parent symlink succeeded")
		}
		markerTarget := filepath.Join(t.TempDir(), "marker.json")
		if err := os.WriteFile(markerTarget, []byte(`{"purpose":"unsafe"}`), 0o600); err != nil {
			t.Fatalf("write marker target: %v", err)
		}
		if err := os.Symlink(markerTarget, filepath.Join(rootPath, "restore-verification-target.json")); err != nil {
			t.Fatalf("create marker symlink: %v", err)
		}
		if _, err := storage.ReadMarker(65536); err == nil {
			t.Fatal("symlinked restore-verification marker read succeeded")
		}
		if err := os.Remove(filepath.Join(rootPath, "restore-verification-target.json")); err != nil {
			t.Fatalf("remove marker symlink: %v", err)
		}

		proof, err := storage.WriteArtifact(context.Background(), "safe/proof", []byte("safe"), "text/plain")
		if err != nil {
			t.Fatalf("write safe fixture: %v", err)
		}
		path := filepath.Join(rootPath, filepath.FromSlash(proof.Key))
		hardLink := filepath.Join(rootPath, "safe", "hard-link")
		if err := os.Link(path, hardLink); err != nil {
			t.Fatalf("create hard link: %v", err)
		}
		if _, err := storage.ReadArtifact(context.Background(), proof.Key, proof.SizeBytes); err == nil {
			t.Fatal("hard-linked recovery artifact read succeeded")
		}
		if err := os.Remove(hardLink); err != nil {
			t.Fatalf("remove hard-link fixture: %v", err)
		}
		fifo := filepath.Join(rootPath, "safe", "fifo")
		if err := unix.Mkfifo(fifo, 0o600); err != nil {
			t.Fatalf("create FIFO: %v", err)
		}
		if _, err := storage.ReadArtifact(context.Background(), "safe/fifo", 1024); err == nil {
			t.Fatal("FIFO recovery artifact read succeeded")
		}
	})

	t.Run("cancellation and root replacement leave no partial publication or disclosure", func(t *testing.T) {
		rootPath := filepath.Join(t.TempDir(), "backups")
		storage, err := NewFilesystemStorage(rootPath)
		if err != nil {
			t.Fatalf("NewFilesystemStorage: %v", err)
		}
		defer storage.Close()
		canceled, cancel := context.WithCancel(context.Background())
		cancel()
		if _, err := storage.WriteArtifact(canceled, "canceled/proof", []byte("partial"), "text/plain"); !errors.Is(err, context.Canceled) {
			t.Fatalf("canceled recovery write error = %v", err)
		}
		if _, err := os.Stat(filepath.Join(rootPath, "canceled", "proof")); !os.IsNotExist(err) {
			t.Fatalf("canceled recovery write left partial output: %v", err)
		}
		moved := rootPath + "-original"
		if err := os.Rename(rootPath, moved); err != nil {
			t.Fatalf("move recovery root: %v", err)
		}
		if err := os.Mkdir(rootPath, 0o700); err != nil {
			t.Fatalf("create replacement recovery root: %v", err)
		}
		_, err = storage.WriteArtifact(context.Background(), "replacement/proof", []byte("unsafe"), "text/plain")
		if !errors.Is(err, rootedfs.ErrRootIdentityChanged) {
			t.Fatalf("root replacement error = %v", err)
		}
		if strings.Contains(err.Error(), rootPath) {
			t.Fatalf("root replacement error disclosed host root: %v", err)
		}
		if _, err := os.Stat(filepath.Join(rootPath, "replacement", "proof")); !os.IsNotExist(err) {
			t.Fatalf("replacement root contains a published artifact: %v", err)
		}
	})

	t.Run("managed service never creates a filesystem fallback", func(t *testing.T) {
		rootPath := filepath.Join(t.TempDir(), "must-not-exist")
		if _, err := NewBackupStorage("managed_service", rootPath, nil); !errors.Is(err, ErrUnsupportedBackupBinding) {
			t.Fatalf("managed-service backup binding error = %v", err)
		}
		if _, err := os.Stat(rootPath); !os.IsNotExist(err) {
			t.Fatalf("managed-service binding created a filesystem fallback: %v", err)
		}
	})
}

func requireMode(t testing.TB, path string, want os.FileMode) {
	t.Helper()
	info, err := os.Lstat(path)
	if err != nil {
		t.Fatalf("stat %s: %v", filepath.Base(path), err)
	}
	if info.Mode().Perm() != want {
		t.Fatalf("%s mode = %o want %o", filepath.Base(path), info.Mode().Perm(), want)
	}
}
