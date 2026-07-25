package objectstore_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"golang.org/x/sys/unix"

	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/platform/rootedfs"
)

type objectStoreTestSettings struct {
	binding         objectstore.Binding
	instrumentation objectstore.Instrumentation
}

func TestObjectStoreInitialization_Integration(t *testing.T) {
	t.Run("derives disconnected filesystem-root storage only from roots.object_storage.path", func(t *testing.T) {
		cfg := disconnectedObjectStoreConfig(t)
		store, err := setupObjectStore(context.Background(), cfg, map[string]string{
			objectstore.EndpointEnv:  "127.0.0.1:1",
			objectstore.AccessKeyEnv: "wrong",
			objectstore.SecretKeyEnv: "wrong",
			objectstore.BucketEnv:    "wrong-bucket",
		})
		if err != nil {
			t.Fatalf("setup filesystem-root object store: %v", err)
		}
		defer store.Close()

		payload := []byte("bootstrap object-store bootstrap")
		if err := store.PutObject(context.Background(), "proof/proof.txt", bytes.NewReader(payload), int64(len(payload)), "text/plain"); err != nil {
			t.Fatalf("put filesystem object: %v", err)
		}

		stat, err := store.StatObject(context.Background(), "proof/proof.txt")
		if err != nil {
			t.Fatalf("stat filesystem object: %v", err)
		}
		if stat.Size != int64(len(payload)) || stat.ContentType != "text/plain" {
			t.Fatalf("unexpected filesystem object stat: %#v", stat)
		}

		object, _, err := store.ReadObject(context.Background(), "proof/proof.txt", objectstore.ReadOptions{})
		if err != nil {
			t.Fatalf("read filesystem object: %v", err)
		}
		defer object.Close()

		got, err := io.ReadAll(object)
		if err != nil {
			t.Fatalf("read object payload: %v", err)
		}
		if !bytes.Equal(got, payload) {
			t.Fatalf("unexpected object payload: got %q want %q", got, payload)
		}

		objects, err := store.ListObjects(context.Background(), "")
		if err != nil {
			t.Fatalf("list filesystem objects: %v", err)
		}
		if len(objects) != 1 || objects[0].Key != "proof/proof.txt" {
			t.Fatalf("unexpected filesystem object listing: %#v", objects)
		}

		if err := store.DeleteObject(context.Background(), "proof/proof.txt"); err != nil {
			t.Fatalf("delete filesystem object: %v", err)
		}
		objects, err = store.ListObjects(context.Background(), "")
		if err != nil {
			t.Fatalf("list filesystem objects after delete: %v", err)
		}
		if len(objects) != 0 {
			t.Fatalf("expected no filesystem objects after delete, got %#v", objects)
		}
	})

	t.Run("issues same-origin upload targets for filesystem-root storage", func(t *testing.T) {
		store, err := setupObjectStore(context.Background(), disconnectedObjectStoreConfig(t), nil)
		if err != nil {
			t.Fatalf("setup filesystem-root object store: %v", err)
		}
		defer store.Close()
		target, err := store.UploadTarget(context.Background(), "uploads/proof.txt", time.Now().Add(time.Minute))
		if err != nil {
			t.Fatalf("create filesystem upload target: %v", err)
		}
		if !strings.HasPrefix(target.Href, "/api/v1/object-uploads/") || target.Method != "PUT" {
			t.Fatalf("unexpected filesystem upload target: %#v", target)
		}
		token := strings.TrimPrefix(target.Href, "/api/v1/object-uploads/")
		if err := store.CompleteUploadTarget(context.Background(), token, strings.NewReader("uploaded"), "text/plain"); err != nil {
			t.Fatalf("complete filesystem upload target: %v", err)
		}
		stat, err := store.StatObject(context.Background(), "uploads/proof.txt")
		if err != nil {
			t.Fatalf("stat uploaded filesystem object: %v", err)
		}
		if stat.Size != int64(len("uploaded")) {
			t.Fatalf("unexpected uploaded filesystem object size: %#v", stat)
		}
	})

	t.Run("rejects object keys that escape the configured root", func(t *testing.T) {
		store, err := setupObjectStore(context.Background(), disconnectedObjectStoreConfig(t), nil)
		if err != nil {
			t.Fatalf("setup filesystem-root object store: %v", err)
		}
		defer store.Close()
		if err := store.PutObject(context.Background(), "../escape.txt", strings.NewReader("escape"), int64(len("escape")), "text/plain"); err == nil {
			t.Fatal("expected filesystem object store to reject escaping keys")
		}
	})

	t.Run("rejects every noncanonical logical object reference", func(t *testing.T) {
		store, err := setupObjectStore(context.Background(), disconnectedObjectStoreConfig(t), nil)
		if err != nil {
			t.Fatalf("setup filesystem-root object store: %v", err)
		}
		defer store.Close()
		for name, key := range map[string]string{
			"absolute":        "/escape.txt",
			"backslash":       `escape\file.txt`,
			"dot":             "escape/./file.txt",
			"empty component": "escape//file.txt",
			"nul":             "escape/\x00file.txt",
			"parent":          "escape/../file.txt",
			"unicode":         "cafe\u0301/file.txt",
		} {
			t.Run(name, func(t *testing.T) {
				if err := store.PutObject(context.Background(), key, strings.NewReader("escape"), 6, "text/plain"); err == nil {
					t.Fatalf("noncanonical key %q was accepted", key)
				}
			})
		}
	})

	t.Run("rejects symlinks that escape the configured root", func(t *testing.T) {
		cfg := disconnectedObjectStoreConfig(t)
		store, err := setupObjectStore(context.Background(), cfg, nil)
		if err != nil {
			t.Fatalf("setup filesystem-root object store: %v", err)
		}
		defer store.Close()

		outsideDir := t.TempDir()
		if err := os.WriteFile(filepath.Join(outsideDir, "payload.txt"), []byte("escaped"), 0o600); err != nil {
			t.Fatalf("write escaped object payload: %v", err)
		}
		if err := os.Symlink(outsideDir, filepath.Join(cfg.binding.RootPath, "linked")); err != nil {
			t.Fatalf("create escaping object-store symlink: %v", err)
		}

		if err := store.PutObject(context.Background(), "linked/write.txt", strings.NewReader("escape"), int64(len("escape")), "text/plain"); err == nil {
			t.Fatal("expected filesystem object store to reject writes through an escaping symlink")
		}
		object, _, err := store.ReadObject(context.Background(), "linked/payload.txt", objectstore.ReadOptions{})
		if err == nil {
			_ = object.Close()
			t.Fatal("expected filesystem object store to reject reads through an escaping symlink")
		}
	})

	t.Run("rejects hard links and special objects at actual reads and listings", func(t *testing.T) {
		cfg := disconnectedObjectStoreConfig(t)
		store, err := setupObjectStore(context.Background(), cfg, nil)
		if err != nil {
			t.Fatalf("setup filesystem-root object store: %v", err)
		}
		defer store.Close()
		if err := store.PutObject(context.Background(), "safe/payload.txt", strings.NewReader("safe"), 4, "text/plain"); err != nil {
			t.Fatalf("write safe object: %v", err)
		}
		if err := os.Link(
			filepath.Join(cfg.binding.RootPath, "safe", "payload.txt"),
			filepath.Join(cfg.binding.RootPath, "safe", "hard-link.txt"),
		); err != nil {
			t.Fatalf("create hard link: %v", err)
		}
		if object, _, err := store.ReadObject(context.Background(), "safe/hard-link.txt", objectstore.ReadOptions{}); err == nil {
			_ = object.Close()
			t.Fatal("hard-linked object read succeeded")
		}
		if _, err := store.ListObjects(context.Background(), ""); err == nil {
			t.Fatal("listing with a hard-linked object succeeded")
		}
		if err := os.Remove(filepath.Join(cfg.binding.RootPath, "safe", "hard-link.txt")); err != nil {
			t.Fatalf("remove hard link fixture: %v", err)
		}
		if err := unix.Mkfifo(filepath.Join(cfg.binding.RootPath, "safe", "fifo"), 0o600); err != nil {
			t.Fatalf("create FIFO: %v", err)
		}
		if object, _, err := store.ReadObject(context.Background(), "safe/fifo", objectstore.ReadOptions{}); err == nil {
			_ = object.Close()
			t.Fatal("FIFO object read succeeded")
		}
		if _, err := store.ListObjects(context.Background(), ""); err == nil {
			t.Fatal("listing with a FIFO succeeded")
		}
	})

	t.Run("cancellation and root replacement fail without partial publication or root disclosure", func(t *testing.T) {
		cfg := disconnectedObjectStoreConfig(t)
		store, err := setupObjectStore(context.Background(), cfg, nil)
		if err != nil {
			t.Fatalf("setup filesystem-root object store: %v", err)
		}
		defer store.Close()
		canceled, cancel := context.WithCancel(context.Background())
		cancel()
		if err := store.PutObject(canceled, "canceled/payload.txt", strings.NewReader("partial"), 7, "text/plain"); !errors.Is(err, context.Canceled) {
			t.Fatalf("canceled put error = %v", err)
		}
		if _, err := os.Stat(filepath.Join(cfg.binding.RootPath, "canceled", "payload.txt")); !os.IsNotExist(err) {
			t.Fatalf("canceled put left a partial destination: %v", err)
		}

		original := cfg.binding.RootPath + "-original"
		if err := os.Rename(cfg.binding.RootPath, original); err != nil {
			t.Fatalf("replace object-store root: %v", err)
		}
		if err := os.Mkdir(cfg.binding.RootPath, 0o700); err != nil {
			t.Fatalf("create replacement object-store root: %v", err)
		}
		err = store.PutObject(context.Background(), "replacement/payload.txt", strings.NewReader("unsafe"), 6, "text/plain")
		if !errors.Is(err, rootedfs.ErrRootIdentityChanged) {
			t.Fatalf("root replacement error = %v", err)
		}
		if strings.Contains(err.Error(), cfg.binding.RootPath) {
			t.Fatalf("root replacement error disclosed host root: %v", err)
		}
		if _, err := os.Stat(filepath.Join(cfg.binding.RootPath, "replacement", "payload.txt")); !os.IsNotExist(err) {
			t.Fatalf("root replacement published into replacement root: %v", err)
		}
	})
}

func setupObjectStore(ctx context.Context, cfg objectStoreTestSettings, env map[string]string) (objectstore.Store, error) {
	settings, err := objectstore.ResolveSettings(cfg.binding, env)
	if err != nil {
		return nil, err
	}
	return objectstore.Setup(ctx, settings, cfg.instrumentation)
}

func ensureObjectStoreBucket(ctx context.Context, cfg objectStoreTestSettings, env map[string]string) (objectstore.EnsureBucketResult, error) {
	settings, err := objectstore.ResolveSettings(cfg.binding, env)
	if err != nil {
		return objectstore.EnsureBucketResult{}, err
	}
	return objectstore.EnsureBucket(ctx, settings)
}

func disconnectedObjectStoreConfig(t testing.TB) objectStoreTestSettings {
	t.Helper()

	base := t.TempDir()
	return objectStoreTestSettings{
		binding: objectstore.Binding{
			BindingKind: "filesystem_root",
			RootPath:    filepath.Join(base, "object-store"),
		},
	}
}
