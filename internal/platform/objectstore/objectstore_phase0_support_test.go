package objectstore_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/config"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/testutil/s3test"
)

func TestSupportPhase0_ManagedServiceObjectStoreBinding(t *testing.T) {
	s3Harness := s3test.Start(t)

	t.Run("derives managed-service settings from roots.object_storage.service_ref", func(t *testing.T) {
		bucket := fmt.Sprintf("phase0-support-managed-object-%d", time.Now().UnixNano())
		defer func() {
			if err := s3Harness.CleanupBucket(context.Background(), bucket); err != nil {
				t.Logf("cleanup bucket: %v", err)
			}
		}()

		client, err := objectstore.SetupWithEnv(context.Background(), managedObjectStoreConfig(t), s3Harness.EnvForServiceRef("object_primary", bucket))
		if err != nil {
			t.Fatalf("setup object store from managed-service binding: %v", err)
		}

		exists, err := client.BucketExists(context.Background(), bucket)
		if err != nil {
			t.Fatalf("check bucket existence: %v", err)
		}
		if !exists {
			t.Fatalf("expected setup to create configured bucket %q", bucket)
		}
	})

	t.Run("fails closed when the configured managed-service binding cannot resolve service-ref settings", func(t *testing.T) {
		bucket := fmt.Sprintf("phase0-support-managed-object-fail-%d", time.Now().UnixNano())

		_, err := objectstore.SetupWithEnv(context.Background(), managedObjectStoreConfig(t), s3Harness.Env(bucket))
		if err == nil {
			t.Fatal("expected managed-service object-store binding to reject unrelated generic env settings")
		}
		if !strings.Contains(err.Error(), "missing object-store endpoint") {
			t.Fatalf("unexpected managed-service binding failure: %v", err)
		}
	})
}

func managedObjectStoreConfig(t testing.TB) config.Config {
	t.Helper()

	cfg, err := config.Validate(config.Config{
		ConfigSchemaID:    "cartulary.deployment_config.v1",
		DeploymentProfile: "on_prem",
		Roots: config.RootBindings{
			DatabaseStorage: config.RootBinding{
				BindingKind: "filesystem_root",
				Path:        "/var/lib/cartulary/postgres",
			},
			ObjectStorage: config.RootBinding{
				BindingKind: "managed_service",
				ServiceRef:  "object_primary",
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
		t.Fatalf("validate managed object-store config: %v", err)
	}

	return cfg
}
