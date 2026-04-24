package objectstore_test

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/minio/minio-go/v7"

	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/testutil/s3test"
)

func TestPhase0_ObjectStoreInitialization_I_0_02(t *testing.T) {
	s3Harness := s3test.Start(t)

	t.Run("derives disconnected filesystem-root settings from the generic CARTULARY_S3 environment contract", func(t *testing.T) {
		bucket := fmt.Sprintf("phase0-i-0-02-%d", time.Now().UnixNano())
		defer func() {
			if err := s3Harness.CleanupBucket(context.Background(), bucket); err != nil {
				t.Logf("cleanup bucket: %v", err)
			}
		}()

		client, err := objectstore.SetupWithEnv(context.Background(), disconnectedObjectStoreConfig(t), s3Harness.Env(bucket))
		if err != nil {
			t.Fatalf("setup object store from disconnected filesystem-root binding: %v", err)
		}

		exists, err := client.BucketExists(context.Background(), bucket)
		if err != nil {
			t.Fatalf("check bucket existence: %v", err)
		}
		if !exists {
			t.Fatalf("expected setup to create configured bucket %q", bucket)
		}

		payload := []byte("phase0 object-store bootstrap")
		_, err = client.PutObject(context.Background(), bucket, "proof.txt", bytes.NewReader(payload), int64(len(payload)), minio.PutObjectOptions{})
		if err != nil {
			t.Fatalf("put object: %v", err)
		}

		object, err := client.GetObject(context.Background(), bucket, "proof.txt", minio.GetObjectOptions{})
		if err != nil {
			t.Fatalf("get object: %v", err)
		}
		defer object.Close()

		got, err := io.ReadAll(object)
		if err != nil {
			t.Fatalf("read object payload: %v", err)
		}
		if !bytes.Equal(got, payload) {
			t.Fatalf("unexpected object payload: got %q want %q", got, payload)
		}

		if err := client.RemoveObject(context.Background(), bucket, "proof.txt", minio.RemoveObjectOptions{}); err != nil {
			t.Fatalf("remove object: %v", err)
		}
	})

	t.Run("fails closed when only managed-service variables are present for a disconnected filesystem-root binding", func(t *testing.T) {
		bucket := fmt.Sprintf("phase0-i-0-02-fail-%d", time.Now().UnixNano())
		env := s3Harness.EnvForServiceRef("object_primary", bucket)
		env[objectstore.EndpointEnv] = "127.0.0.1:1"
		env[objectstore.AccessKeyEnv] = "wrong"
		env[objectstore.SecretKeyEnv] = "wrong"
		env[objectstore.BucketEnv] = "wrong-bucket"

		_, err := objectstore.SetupWithEnv(context.Background(), disconnectedObjectStoreConfig(t), env)
		if err == nil {
			t.Fatal("expected disconnected filesystem-root binding to reject an unusable generic object-store environment")
		}
		if !strings.Contains(err.Error(), `check object store bucket "wrong-bucket"`) {
			t.Fatalf("unexpected disconnected object-store binding failure: %v", err)
		}
	})
}

func disconnectedObjectStoreConfig(t testing.TB) config.Config {
	t.Helper()

	cfg, err := config.Validate(config.Config{
		ConfigSchemaID:    "cartulary.deployment_config.v1",
		DeploymentProfile: "disconnected",
		Application: config.ApplicationConfig{
			PublicOrigin: "http://localhost:5173",
		},
		Roots: config.RootBindings{
			DatabaseStorage: config.RootBinding{
				BindingKind: "filesystem_root",
				Path:        "/var/lib/cartulary/postgres",
			},
			ObjectStorage: config.RootBinding{
				BindingKind: "filesystem_root",
				Path:        "/var/lib/cartulary/object-store",
			},
			BackupStorage: config.RootBinding{
				BindingKind: "filesystem_root",
				Path:        "/var/lib/cartulary/backups",
			},
			ReferencePackStorage: config.RootBinding{
				BindingKind: "filesystem_root",
				Path:        "/var/lib/cartulary/reference-packs",
			},
			TemporaryWork: config.RootBinding{
				BindingKind: "filesystem_root",
				Path:        "/var/lib/cartulary/tmp",
			},
			ExportOutputs: config.RootBinding{
				BindingKind: "filesystem_root",
				Path:        "/var/lib/cartulary/exports",
			},
		},
	})
	if err != nil {
		t.Fatalf("validate disconnected object-store config: %v", err)
	}

	return cfg
}
