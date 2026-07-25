package objectstore_test

import (
	"context"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/testutil/s3test"
)

func TestManagedServiceObjectStoreBinding(t *testing.T) {
	s3Harness := s3test.Start(t)

	t.Run("derives managed-service settings from roots.object_storage.service_ref", func(t *testing.T) {
		bucket := fmt.Sprintf("bootstrap-support-managed-object-%d", time.Now().UnixNano())
		defer func() {
			if err := s3Harness.CleanupBucket(context.Background(), bucket); err != nil {
				t.Logf("cleanup bucket: %v", err)
			}
		}()
		if err := s3Harness.CreateBucket(context.Background(), bucket); err != nil {
			t.Fatalf("create managed-service test bucket: %v", err)
		}

		store, err := setupObjectStore(context.Background(), managedObjectStoreConfig(t), s3Harness.EnvForServiceRef("object_primary", bucket))
		if err != nil {
			t.Fatalf("setup object store from managed-service binding: %v", err)
		}

		if err := store.PutObject(context.Background(), "support/proof.txt", strings.NewReader("proof"), int64(len("proof")), "text/plain"); err != nil {
			t.Fatalf("put managed-service proof object: %v", err)
		}
		if _, err := store.StatObject(context.Background(), "support/proof.txt"); err != nil {
			t.Fatalf("stat managed-service proof object: %v", err)
		}
	})

	t.Run("fails closed when the configured managed-service binding cannot resolve service-ref settings", func(t *testing.T) {
		bucket := fmt.Sprintf("bootstrap-support-managed-object-fail-%d", time.Now().UnixNano())

		_, err := setupObjectStore(context.Background(), managedObjectStoreConfig(t), s3Harness.Env(bucket))
		if err == nil {
			t.Fatal("expected managed-service object-store binding to reject unrelated generic env settings")
		}
		if !strings.Contains(err.Error(), "missing object-store endpoint") {
			t.Fatalf("unexpected managed-service binding failure: %v", err)
		}
	})

	t.Run("fails closed when the configured managed-service bucket is missing", func(t *testing.T) {
		bucket := fmt.Sprintf("bootstrap-support-managed-object-missing-%d", time.Now().UnixNano())

		_, err := setupObjectStore(context.Background(), managedObjectStoreConfig(t), s3Harness.EnvForServiceRef("object_primary", bucket))
		if err == nil {
			t.Fatal("expected managed-service object-store setup to reject a missing bucket")
		}
		adapterErr, ok := objectstore.AsAdapterError(err)
		if !ok || adapterErr.Code != objectstore.ErrorCodeUnavailable || adapterErr.Reason != objectstore.ReasonBucketMissing {
			t.Fatalf("unexpected missing bucket error: %#v %v", adapterErr, err)
		}
	})

	t.Run("deployment-local init creates the configured managed-service bucket", func(t *testing.T) {
		bucket := fmt.Sprintf("bootstrap-support-managed-object-init-%d", time.Now().UnixNano())
		defer func() {
			if err := s3Harness.CleanupBucket(context.Background(), bucket); err != nil {
				t.Logf("cleanup bucket: %v", err)
			}
		}()

		env := s3Harness.EnvForServiceRef("object_primary", bucket)
		if _, err := setupObjectStore(context.Background(), managedObjectStoreConfig(t), env); err == nil {
			t.Fatal("expected managed-service object-store setup to fail before bucket init")
		}

		result, err := ensureObjectStoreBucket(context.Background(), managedObjectStoreConfig(t), env)
		if err != nil {
			t.Fatalf("ensure configured bucket: %v", err)
		}
		if !result.Created || result.AlreadyExists {
			t.Fatalf("unexpected first ensure result: %#v", result)
		}

		store, err := setupObjectStore(context.Background(), managedObjectStoreConfig(t), env)
		if err != nil {
			t.Fatalf("setup object store after bucket init: %v", err)
		}
		defer store.Close()

		second, err := ensureObjectStoreBucket(context.Background(), managedObjectStoreConfig(t), env)
		if err != nil {
			t.Fatalf("ensure existing configured bucket: %v", err)
		}
		if second.Created || !second.AlreadyExists {
			t.Fatalf("unexpected second ensure result: %#v", second)
		}
	})

	t.Run("managed service env key normalization is ASCII only", func(t *testing.T) {
		keys, err := objectstore.EnvKeysForServiceRef("object.primary-1")
		if err != nil {
			t.Fatalf("resolve punctuation service-ref keys: %v", err)
		}
		if keys.Endpoint != "CARTULARY_S3_OBJECT_PRIMARY_1_ENDPOINT" ||
			keys.AccessKey != "CARTULARY_S3_OBJECT_PRIMARY_1_ACCESS_KEY_ID" ||
			keys.SecretKey != "CARTULARY_S3_OBJECT_PRIMARY_1_SECRET_ACCESS_KEY" ||
			keys.Secure != "CARTULARY_S3_OBJECT_PRIMARY_1_SECURE" ||
			keys.Bucket != "CARTULARY_S3_OBJECT_PRIMARY_1_BUCKET" {
			t.Fatalf("unexpected normalized object-store keys: %#v", keys)
		}

		if _, err := objectstore.EnvKeysForServiceRef("é"); err == nil {
			t.Fatal("expected non-ASCII-only service_ref to be rejected")
		}
	})

	t.Run("object_store_adapter_contract_hardening", requireObjectStoreAdapterContractHardening)
}

func managedObjectStoreConfig(t testing.TB) objectStoreTestSettings {
	t.Helper()
	return objectStoreTestSettings{binding: objectstore.Binding{
		BindingKind: "managed_service",
		ServiceRef:  "object_primary",
	}}
}
