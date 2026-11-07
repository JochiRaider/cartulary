package reference_data_test

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/app/referenceassembly"
	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data/producer"
	referencetest "github.com/JochiRaider/cartulary/internal/modules/reference_data/testsupport"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestWindowsOnlineProductionLifecycle_Integration(t *testing.T) {
	expected := referencetest.OwnerFixture(t, "windows_online")
	if expected.Condition == nil || *expected.Condition != "verified_available" {
		t.Fatal("Windows fixture expectation")
	}
	ctx := context.Background()
	now := time.Now().UTC()
	keys, err := producer.NewWindowsFixtureSigningContext()
	if err != nil {
		t.Fatal(err)
	}
	a := referencetest.WindowsFixture(t, "fixture-a", now, keys, 1)
	b := referencetest.WindowsFixture(t, "fixture-b", now, keys, 2)
	bootstrap := filepath.Join(t.TempDir(), "bootstrap.json")
	if err = os.WriteFile(bootstrap, a.Bootstrap, 0600); err != nil {
		t.Fatal(err)
	}
	runtime := appsupport.StartRuntime(t)
	harness := runtime.StartServer(t, appsupport.ServerOptions{Prefix: "windows-online", TestRouteMode: httptestx.TestRouteModeDisabled, Env: map[string]string{"CARTULARY__REFERENCE_PACKS__TRUST_BOOTSTRAP_PATH": bootstrap, "CARTULARY__REFERENCE_PACKS__CLOCK_TRUSTED": "true"}})
	admin, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, harness.Server.HTTP.URL)
	storage, err := referenceassembly.NewRootStorage(harness.Server.Config.Roots.TemporaryWork.Path, harness.Server.Config.Roots.ReferencePackStorage.Path)
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()
	consumer, err := reference_data.NewConsumer(harness.Pool, storage, pagination.NewCodec([32]byte([]byte(strings.Repeat("w", 32)))), func() time.Time { return time.Now().UTC() }, consumerIntegrityOptions(t))
	if err != nil {
		t.Fatal(err)
	}
	before := consumer.ResolveCurrentPackSet(ctx)
	if before.Error != nil || len(before.Value.Members) != 3 {
		t.Fatal("three required Base registries", before.Error)
	}
	for _, m := range before.Value.Members {
		if !slices.Contains([]string{"type_registry.host", "type_registry.evidence", "type_registry.indicator"}, m.Key) {
			t.Fatal("unexpected Base member", m.Key)
		}
	}
	upload := func(f referencetest.ProducedWindowsPack, txn string) {
		response := postReferencePackUpload(t, harness.Server.HTTP.URL, admin, `{"client_txn_id":"`+txn+`"}`, f.Container, "windows.tar", "application/x-tar")
		job := httptestx.RequireSuccessEnvelope(t, response, http.StatusAccepted)["data"].(map[string]any)
		if terminal := requireJob(t, harness, admin, job["job_id"].(string)); terminal["status"] != "succeeded" {
			t.Fatalf("Windows producer import: %#v", terminal)
		}
	}
	activate := func(f referencetest.ProducedWindowsPack, txn string) {
		httptestx.RequireSuccessEnvelope(t, postAction(t, harness, admin, "/api/v1/reference-packs/"+f.Key+"/"+f.Version+"/activate", txn, ""), http.StatusOK)
	}
	upload(a, "windows-import-a")
	if staged := consumer.ResolveCurrentPackSet(ctx); staged.Error != nil || staged.Value.ID != before.Value.ID {
		t.Fatal("staging changed activation")
	}
	activate(a, "windows-activate-a")
	selected := consumer.ResolveCurrentPackSet(ctx)
	if selected.Error != nil || len(selected.Value.Members) != 4 {
		t.Fatal("Windows plus Base set", selected.Error)
	}
	if selected.Value.ID != "rpset_"+referencetest.WindowsSetDigest(t, "../../..", a) {
		t.Fatal("independent exact-set identity")
	}
	t.Logf("Windows A exact set=%s", selected.Value.ID)
	limit := 1
	request := reference_data.LookupPackEntriesRequest{PackSetID: selected.Value.ID, PackKey: a.Key, LookupKind: "provider_event_id", LookupValue: map[string]any{"provider_name": "PoWeRsHeLlCoRe", "event_id": json.Number("4104")}, Limit: &limit}
	for i, want := range []string{"powershellcore:4104:*", "powershellcore:4104:0", "powershellcore:4104:1"} {
		page := consumer.LookupPackEntries(ctx, request)
		assertConsumerContract(t, "LookupPackEntries", page)
		if page.Error != nil || len(page.Value.Items) != 1 {
			t.Fatal("version page", page.Error)
		}
		var row map[string]any
		if err = json.Unmarshal(page.Value.Items[0], &row); err != nil || row["entry_id"] != want {
			t.Fatal("ordered nullable versions", err, row)
		}

		request.Cursor = page.Value.NextCursor
		if (i == 2) != (request.Cursor == nil) {
			t.Fatal("pagination boundary")
		}
	}
	request.Limit = nil
	request.Cursor = nil
	request.LookupValue = map[string]any{"provider_name": "PowerShellCore", "event_id": json.Number("4104"), "event_version": json.Number("0")}
	if page := consumer.LookupPackEntries(ctx, request); page.Error != nil || len(page.Value.Items) != 1 {
		t.Fatal("numeric version excludes null", page.Error)
	}
	request.LookupValue = map[string]any{"provider_name": "PowerShellCore", "event_id": json.Number("4104"), "event_version": nil}
	if page := consumer.LookupPackEntries(ctx, request); page.Error == nil {
		t.Fatal("explicit null version accepted")
	}
	request.LookupKind = "event_id"
	request.LookupValue = json.Number("4104")
	if page := consumer.LookupPackEntries(ctx, request); page.Error != nil || len(page.Value.Items) != 4 {
		t.Fatal("ID-only provider ambiguity", page.Error)
	}
	upload(b, "windows-import-b")
	activate(b, "windows-activate-b")
	current := consumer.ResolveCurrentPackSet(ctx)
	if current.Error != nil || current.Value.ID != "rpset_"+referencetest.WindowsSetDigest(t, "../../..", b) || current.Value.ID == selected.Value.ID {
		t.Fatal("derivative B set", current.Error)
	}
	t.Logf("Windows B exact set=%s", current.Value.ID)
	retained := consumer.GetPackEntry(ctx, reference_data.GetPackEntryRequest{PackSetID: selected.Value.ID, PackKey: a.Key, EntryID: "powershellcore:4104:*"})
	assertConsumerContract(t, "GetPackEntry", retained)
	if retained.Error != nil || retained.Value.Provenance.ManifestSHA256 != a.ManifestSHA256 || retained.Value.Provenance.PackVersion != a.Version {
		t.Fatal("retained A provenance", retained.Error)
	}
	renewed := referencetest.WindowsFixture(t, "fixture-b", now, keys, 3)
	if renewed.ManifestSHA256 != b.ManifestSHA256 || renewed.PayloadSHA256 != b.PayloadSHA256 {
		t.Fatal("renewal changed logical content")
	}
	upload(renewed, "windows-renew-b")
	if after := consumer.ResolveCurrentPackSet(ctx); after.Error != nil || after.Value.ID != current.Value.ID {
		t.Fatal("renewal rebased current set", after.Error)
	}
}
