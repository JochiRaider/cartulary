package reference_data_test

import (
	"context"
	"github.com/JochiRaider/cartulary/internal/app/server"
	"github.com/JochiRaider/cartulary/internal/platform/objectstore"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"net/http"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
)

func TestWorkerDatabaseBarriersAreInstanceScoped_Integration(t *testing.T) {
	runtime := appsupport.StartRuntime(t)
	firstBarrier, secondBarrier := &appsupport.ReferencePackVerificationBarrier{}, &appsupport.ReferencePackVerificationBarrier{}
	first := startReferencePackServerWithEnv(t, runtime, "reference-worker-first", nil, firstBarrier)
	// The package bucket is leased until test cleanup; a simultaneous server
	// needs an independent fixture instead of requesting that lease again.
	bucket := runtime.S3.BootstrapBucketT(t, "reference-worker-second")
	store, err := objectstore.Setup(context.Background(), objectstore.Settings{BindingKind: "managed_service", Endpoint: runtime.S3.Endpoint, AccessKey: runtime.S3.AccessKey, SecretKey: runtime.S3.SecretKey, Secure: runtime.S3.Secure, Bucket: bucket}, objectstore.Instrumentation{})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = store.Close() })
	second := startReferencePackServerConfigured(t, runtime, appsupport.ServerOptions{Prefix: "reference-worker-second", ObjectStore: store, Env: runtime.S3.Env(bucket), TestRouteMode: httptestx.TestRouteModeDisabled, ConfigureRuntime: func(o *server.Options) { o.ReferenceDataComposition = secondBarrier }})
	firstAdmin, _ := flowtest.ProvisionBootstrapAdmin(t, first.Server.HTTP.URL)
	secondAdmin, _ := flowtest.ProvisionBootstrapAdmin(t, second.Server.HTTP.URL)
	entered, release := make(chan struct{}), make(chan struct{})
	released := false
	defer func() {
		if !released {
			close(release)
		}
	}()
	firstBarrier.BlockNext(entered, release)
	body := referencePackBundle(t, bundleOptions{PackKey: "type_registry.host", PackKind: "type_registry", PackVersion: "1"})
	response := postReferencePackUpload(t, first.Server.HTTP.URL, firstAdmin, `{"client_txn_id":"isolated-first"}`, body, "pack.zip", reference_data.MediaTypeZip)
	firstJob := requireSuccessEnvelope(t, response, http.StatusAccepted)["data"].(map[string]any)["job_id"].(string)
	select {
	case <-entered:
	case <-time.After(10 * time.Second):
		t.Fatal("first worker did not reach barrier")
	}
	// The second application must finish while the first application is blocked.
	importReferencePack(t, second, secondAdmin, "type_registry.host", "1", "isolated-second")
	if job := requireJobNow(t, first, firstAdmin, firstJob); job["status"] != "running" {
		t.Fatal("first worker escaped its barrier", job)
	}
	if queryCount(t, first.DB, `SELECT count(*) FROM reference_pack_versions WHERE pack_key='type_registry.host' AND pack_version='1'`) != 0 {
		t.Fatal("blocked worker published")
	}
	close(release)
	released = true
	if job := requireJob(t, first, firstAdmin, firstJob); job["status"] != "succeeded" {
		t.Fatal("released worker failed", job)
	}
}
