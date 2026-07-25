package server

import (
	"context"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/rootedfs"
)

func TestReferencePackRootStorageEnforcesReferencesAndLifecycle_Unit(t *testing.T) {
	temporaryRoot := filepath.Join(t.TempDir(), "temporary")
	publishedRoot := filepath.Join(t.TempDir(), "reference-packs")
	storage, err := newReferencePackRootStorage(temporaryRoot, publishedRoot)
	if err != nil {
		t.Fatalf("newReferencePackRootStorage: %v", err)
	}
	t.Cleanup(storage.Close)

	staged, err := storage.Stage(context.Background(), strings.Repeat("a", 64), []byte("staged pack"))
	if err != nil {
		t.Fatalf("Stage: %v", err)
	}
	if filepath.IsAbs(staged.String()) || strings.Contains(staged.String(), temporaryRoot) {
		t.Fatalf("Stage returned host path %q", staged.String())
	}
	stagedPath := filepath.Join(temporaryRoot, filepath.FromSlash(staged.String()))
	assertPrivateRegularFile(t, stagedPath, []byte("staged pack"))
	data, err := storage.ReadStaged(staged, 1024)
	if err != nil || string(data) != "staged pack" {
		t.Fatalf("ReadStaged = %q, %v", data, err)
	}
	if err := storage.RemoveStaged(staged); err != nil {
		t.Fatalf("RemoveStaged: %v", err)
	}
	if err := storage.RemoveStaged(staged); err != nil {
		t.Fatalf("RemoveStaged idempotent retry: %v", err)
	}

	published, err := storage.Publish(context.Background(), strings.Repeat("b", 64), []byte("published pack"))
	if err != nil {
		t.Fatalf("Publish: %v", err)
	}
	if filepath.IsAbs(published.String()) || strings.Contains(published.String(), publishedRoot) {
		t.Fatalf("Publish returned host path %q", published.String())
	}
	publishedPath := filepath.Join(publishedRoot, filepath.FromSlash(published.String()))
	assertPrivateRegularFile(t, publishedPath, []byte("published pack"))
	data, err = storage.ReadPublished(published, 1024)
	if err != nil || string(data) != "published pack" {
		t.Fatalf("ReadPublished = %q, %v", data, err)
	}
	second, err := storage.Publish(context.Background(), strings.Repeat("b", 64), []byte("second pack"))
	if err != nil {
		t.Fatalf("second Publish: %v", err)
	}
	if second.String() == published.String() {
		t.Fatal("separate publications reused a mutable destination")
	}
	assertPrivateRegularFile(t, publishedPath, []byte("published pack"))
	if err := storage.RemovePublished(second); err != nil {
		t.Fatalf("RemovePublished: %v", err)
	}
}

func TestReferencePackRootStorageFailsClosedOnCancellationSymlinkAndRootReplacement_Unit(t *testing.T) {
	t.Run("cancellation leaves no partial publication", func(t *testing.T) {
		storage, err := newReferencePackRootStorage(
			filepath.Join(t.TempDir(), "temporary"),
			filepath.Join(t.TempDir(), "published"),
		)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(storage.Close)
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		if _, err := storage.Stage(ctx, strings.Repeat("c", 64), []byte("must not publish")); err == nil {
			t.Fatal("canceled Stage succeeded")
		}
	})

	t.Run("symlinked component", func(t *testing.T) {
		temporaryRoot := filepath.Join(t.TempDir(), "temporary")
		publishedRoot := filepath.Join(t.TempDir(), "published")
		outside := t.TempDir()
		for _, root := range []string{temporaryRoot, publishedRoot} {
			if err := os.Mkdir(root, 0o700); err != nil {
				t.Fatal(err)
			}
		}
		if err := os.Mkdir(filepath.Join(temporaryRoot, "reference-packs"), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink(outside, filepath.Join(temporaryRoot, "reference-packs", "imports")); err != nil {
			t.Fatal(err)
		}
		storage, err := newReferencePackRootStorage(temporaryRoot, publishedRoot)
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(storage.Close)
		if _, err := storage.Stage(context.Background(), strings.Repeat("d", 64), []byte("escape")); err == nil {
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
		publishedRoot := filepath.Join(base, "published")
		for _, root := range []string{temporaryRoot, publishedRoot} {
			if err := os.Mkdir(root, 0o700); err != nil {
				t.Fatal(err)
			}
		}
		storage, err := newReferencePackRootStorage(temporaryRoot, publishedRoot)
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
		if _, err := storage.Stage(context.Background(), strings.Repeat("e", 64), []byte("replacement")); !errors.Is(err, rootedfs.ErrRootIdentityChanged) {
			t.Fatalf("Stage after root replacement error = %v; want ErrRootIdentityChanged", err)
		}
	})
}
