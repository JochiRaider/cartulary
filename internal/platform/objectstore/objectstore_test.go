package objectstore_test

import (
	"bytes"
	"context"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
)

func TestObjectStoreInitialization_Integration(t *testing.T) {
	t.Run("derives disconnected filesystem-root storage only from roots.object_storage.path", func(t *testing.T) {
		cfg := disconnectedObjectStoreConfig(t)
		store, err := objectstore.SetupWithEnv(context.Background(), cfg, map[string]string{
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
		store, err := objectstore.SetupWithEnv(context.Background(), disconnectedObjectStoreConfig(t), nil)
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
		store, err := objectstore.SetupWithEnv(context.Background(), disconnectedObjectStoreConfig(t), nil)
		if err != nil {
			t.Fatalf("setup filesystem-root object store: %v", err)
		}
		defer store.Close()
		if err := store.PutObject(context.Background(), "../escape.txt", strings.NewReader("escape"), int64(len("escape")), "text/plain"); err == nil {
			t.Fatal("expected filesystem object store to reject escaping keys")
		}
	})

	t.Run("rejects symlinks that escape the configured root", func(t *testing.T) {
		cfg := disconnectedObjectStoreConfig(t)
		store, err := objectstore.SetupWithEnv(context.Background(), cfg, nil)
		if err != nil {
			t.Fatalf("setup filesystem-root object store: %v", err)
		}
		defer store.Close()

		outsideDir := t.TempDir()
		if err := os.WriteFile(filepath.Join(outsideDir, "payload.txt"), []byte("escaped"), 0o600); err != nil {
			t.Fatalf("write escaped object payload: %v", err)
		}
		if err := os.Symlink(outsideDir, filepath.Join(cfg.Roots.ObjectStorage.Path, "linked")); err != nil {
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
}

func disconnectedObjectStoreConfig(t testing.TB) config.Config {
	t.Helper()

	base := t.TempDir()
	cfg, err := config.Validate(config.Config{
		ConfigSchemaID:    "cartulary.deployment_config.v1",
		DeploymentProfile: "disconnected",
		Application: config.ApplicationConfig{
			PublicOrigin: "http://localhost:5173",
		},
		Roots: config.RootBindings{
			DatabaseStorage: config.RootBinding{
				BindingKind: "filesystem_root",
				Path:        filepath.Join(base, "postgres"),
			},
			ObjectStorage: config.RootBinding{
				BindingKind: "filesystem_root",
				Path:        filepath.Join(base, "object-store"),
			},
			BackupStorage: config.RootBinding{
				BindingKind: "filesystem_root",
				Path:        filepath.Join(base, "backups"),
			},
			ReferencePackStorage: config.RootBinding{
				BindingKind: "filesystem_root",
				Path:        filepath.Join(base, "reference-packs"),
			},
			TemporaryWork: config.RootBinding{
				BindingKind: "filesystem_root",
				Path:        filepath.Join(base, "tmp"),
			},
			ExportOutputs: config.RootBinding{
				BindingKind: "filesystem_root",
				Path:        filepath.Join(base, "exports"),
			},
		},
	})
	if err != nil {
		t.Fatalf("validate disconnected object-store config: %v", err)
	}

	return cfg
}
