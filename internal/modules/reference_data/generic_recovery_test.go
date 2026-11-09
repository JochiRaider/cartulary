package reference_data_test

import (
	"net/http"
	"reflect"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/app/referenceassembly"
	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/google/uuid"
)

func runUnregisteredGenericRecovery(t *testing.T, kind string) {
	t.Helper()
	ctx := t.Context()
	key := "advisory.unregistered_fifth_dataset"
	runtime := appsupport.StartRuntime(t)
	source := startReferencePackServer(t, runtime, "generic-recovery-source")
	admin, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, source.Server.HTTP.URL)
	fixture := unregisteredGenericFixture(t, key)
	container := referencePackBundle(t, bundleOptions{PackKey: key, PackKind: kind, PackVersion: "1", Members: fixture.Members})
	response := postReferencePackUpload(t, source.Server.HTTP.URL, admin, `{"client_txn_id":"generic-recovery-import"}`, container, "generic.zip", reference_data.MediaTypeZip)
	job := httptestx.RequireSuccessEnvelope(t, response, http.StatusAccepted)["data"].(map[string]any)
	if state := requireJob(t, source, admin, job["job_id"].(string)); state["status"] != "succeeded" {
		t.Fatal(state)
	}
	prefix := "/api/v1/reference-packs/" + key + "/1"
	httptestx.RequireSuccessEnvelope(t, postAction(t, source, admin, prefix+"/activate", "generic-recovery-activate", ""), http.StatusOK)
	storage, err := referenceassembly.NewRootStorage(source.Server.Config.Roots.TemporaryWork.Path, source.Server.Config.Roots.ReferencePackStorage.Path)
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()
	retention, err := reference_data.NewRetention(source.Pool, storage, time.Now, consumerIntegrityOptions(t))
	if err != nil {
		t.Fatal(err)
	}
	tx, err := source.Pool.Begin(ctx)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback(ctx)
	binding, err := retention.CaptureCurrentTx(ctx, tx, "test_report", uuid.NewString(), uuid.New())
	if err != nil {
		t.Fatal(err)
	}
	if err = tx.Commit(ctx); err != nil {
		t.Fatal(err)
	}
	httptestx.RequireSuccessEnvelope(t, postAction(t, source, admin, prefix+"/disable", "generic-recovery-disable", ""), http.StatusOK)
	source.Server.Close()
	restored := runtime.CaptureRestoreServer(t, source, appsupport.ServerOptions{Prefix: "generic-recovery-target", TestRouteMode: httptestx.TestRouteModeDisabled})
	restoredStorage, err := referenceassembly.NewRootStorage(restored.Server.Config.Roots.TemporaryWork.Path, restored.Server.Config.Roots.ReferencePackStorage.Path)
	if err != nil {
		t.Fatal(err)
	}
	defer restoredStorage.Close()
	retained, err := reference_data.NewRetention(restored.Pool, restoredStorage, time.Now, consumerIntegrityOptions(t))
	if err != nil || retained.ValidateBinding(ctx, binding) != nil {
		t.Fatal("restored immutable binding", err)
	}
	consumer, err := reference_data.NewConsumer(restored.Pool, restoredStorage, pagination.NewCodec([32]byte{7}), time.Now, consumerIntegrityOptions(t))
	if err != nil {
		t.Fatal(err)
	}
	got := consumer.GetPackDescriptor(ctx, reference_data.GetPackDescriptorRequest{PackSetID: binding.SetID, PackKey: key})
	if got.Error != nil || string(got.Value.Descriptor) != fixture.Members["payload/descriptor.json"] {
		t.Fatal("historical descriptor", got.Error)
	}
	var expected reference_data.PackProvenance
	for _, anchor := range binding.Provenance {
		if anchor.PackKey == key {
			expected = anchor
		}
	}
	if !reflect.DeepEqual(expected, got.Value.Provenance) {
		t.Fatal("restore changed authority/provenance")
	}
	current := consumer.ResolveCurrentPackSet(ctx)
	if current.Error != nil {
		t.Fatal(current.Error)
	}
	unavailable := consumer.GetPackDescriptor(ctx, reference_data.GetPackDescriptorRequest{PackSetID: current.Value.ID, PackKey: key})
	if unavailable.Error == nil || unavailable.Error.Code != "pack_unavailable" {
		t.Fatal("restore activated optional pack")
	}
	lookup := consumer.LookupPackEntries(ctx, reference_data.LookupPackEntriesRequest{PackSetID: binding.SetID, PackKey: key, LookupKind: "label", LookupValue: "FIXTURE"})
	if lookup.Error != nil || len(lookup.Value.Items) != 2 {
		t.Fatal("restored historical lookup", lookup.Error)
	}
}
