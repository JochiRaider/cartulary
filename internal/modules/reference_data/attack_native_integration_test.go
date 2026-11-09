package reference_data_test

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/app/referenceassembly"
	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	referencetest "github.com/JochiRaider/cartulary/internal/modules/reference_data/testsupport"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/referencepackfixture"
)

func TestAuthenticATTACKProductionLifecycle_Integration(t *testing.T) {
	runAuthenticATTACKLifecycle(t, false)
}

// Qualify the exact production archive, without fixture signing or renewal.
func TestRetainedATTACKArtifactImport_Integration(t *testing.T) { runAuthenticATTACKLifecycle(t, true) }

func runAuthenticATTACKLifecycle(t *testing.T, retained bool) {
	ctx := context.Background()
	now := time.Now().UTC()
	var err error
	var keys *referencepackfixture.SigningContext
	var fixture referencetest.AuthenticATTACKPack
	if retained {
		fixture = referencetest.ATTACKRelease(t, "../../..")
	} else {
		keys, err = referencepackfixture.NewSigningContext("fixture.attack.dataset.v1", "framework.attack")
		if err != nil {
			t.Fatal(err)
		}
		fixture = referencetest.AuthenticATTACKFixture(t, "../../..", "20.1", now, keys, 3)
	}
	bootstrap := filepath.Join(t.TempDir(), "bootstrap.json")
	if err := os.WriteFile(bootstrap, fixture.Bootstrap, 0600); err != nil {
		t.Fatal(err)
	}
	runtime := appsupport.StartRuntime(t)
	harness := runtime.StartServer(t, appsupport.ServerOptions{Prefix: "authentic-attack", TestRouteMode: httptestx.TestRouteModeDisabled, Env: map[string]string{"CARTULARY__REFERENCE_PACKS__TRUST_BOOTSTRAP_PATH": bootstrap, "CARTULARY__REFERENCE_PACKS__CLOCK_TRUSTED": "true"}})
	admin, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, harness.Server.HTTP.URL)
	storage, err := referenceassembly.NewRootStorage(harness.Server.Config.Roots.TemporaryWork.Path, harness.Server.Config.Roots.ReferencePackStorage.Path)
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()
	consumer, err := reference_data.NewConsumer(harness.Pool, storage, pagination.NewCodec([32]byte([]byte(strings.Repeat("a", 32)))), func() time.Time { return time.Now().UTC() }, consumerIntegrityOptions(t))
	if err != nil {
		t.Fatal(err)
	}
	before := consumer.ResolveCurrentPackSet(ctx)
	if before.Error != nil {
		t.Fatal(before.Error)
	}
	upload := func(container []byte, transaction string) {
		response := postReferencePackUpload(t, harness.Server.HTTP.URL, admin, `{"client_txn_id":"`+transaction+`"}`, container, "attack.tar", "application/x-tar")
		job := httptestx.RequireSuccessEnvelope(t, response, http.StatusAccepted)["data"].(map[string]any)
		if terminal := requireJobWithin(t, harness, admin, job["job_id"].(string), 120*time.Second); terminal["status"] != "succeeded" {
			t.Fatalf("authentic import: %#v", terminal)
		}
	}
	upload(fixture.Container, "authentic-import")
	for kind, want := range map[string]int{"object": int(fixture.Objects), "relationship": int(fixture.Relationships)} {
		if got := queryCount(t, harness.DB, `SELECT count(*) FROM reference_pack_indexes i
            JOIN reference_pack_candidates c ON c.current_index_id=i.index_id
            WHERE c.pack_key=$1 AND c.pack_version=$2 AND i.entry_kind=$3`, fixture.Key, fixture.Version, kind); got != want {
			t.Fatalf("retained %s count=%d, want %d", kind, got, want)
		}
	}
	staged := consumer.ResolveCurrentPackSet(ctx)
	if staged.Error != nil || staged.Value.ID != before.Value.ID {
		t.Fatal("import changed active set", staged.Error)
	}
	prefix := "/api/v1/reference-packs/" + fixture.Key + "/" + fixture.Version
	httptestx.RequireSuccessEnvelope(t, postAction(t, harness, admin, prefix+"/activate", "authentic-activate", ""), http.StatusOK)
	active := consumer.ResolveCurrentPackSet(ctx)
	assertConsumerContract(t, "ResolveCurrentPackSet", active)
	if active.Error != nil || active.Value.ID != "rpset_"+fixture.SetSHA256 {
		t.Fatal("independent authentic set", active.Error)
	}
	for _, q := range []struct{ kind, value, id string }{{"object_id", "attack-pattern--7385dfaf-6886-4229-9ecd-6fd678040830", "attack-pattern--7385dfaf-6886-4229-9ecd-6fd678040830"}, {"external_id", "T1059", "attack-pattern--7385dfaf-6886-4229-9ecd-6fd678040830"}, {"alias", "mimikatz", "tool--afc079f3-c0ea-4096-b75d-3f05338b7f60"}} {
		page := consumer.LookupPackEntries(ctx, reference_data.LookupPackEntriesRequest{PackSetID: active.Value.ID, PackKey: fixture.Key, LookupKind: q.kind, LookupValue: q.value})
		assertConsumerContract(t, "LookupPackEntries", page)
		if page.Error != nil || len(page.Value.Items) != 1 {
			t.Fatalf("authentic %s lookup: %#v", q.kind, page)
		}
		var item map[string]any
		if err := json.Unmarshal(page.Value.Items[0], &item); err != nil || item["object_id"] != q.id {
			t.Fatal("authentic lookup identity", err)
		}
		got := consumer.GetPackEntry(ctx, reference_data.GetPackEntryRequest{PackSetID: active.Value.ID, PackKey: fixture.Key, EntryID: q.id})
		assertConsumerContract(t, "GetPackEntry", got)
		if got.Error != nil || got.Value.Provenance.ManifestSHA256 != fixture.ManifestSHA256 || got.Value.Provenance.SourceVersion != "20.1" {
			t.Fatal("authentic provenance", got.Error)
		}
	}
	upload(fixture.Container, "authentic-replay")
	if !retained {
		renewed := referencetest.AuthenticATTACKFixture(t, "../../..", "20.1", now, keys, 4)
		upload(renewed.Container, "authentic-renew")
		if after := consumer.ResolveCurrentPackSet(ctx); after.Error != nil || after.Value.ID != active.Value.ID {
			t.Fatal("renewal changed immutable selection", after.Error)
		}
	}

	httptestx.RequireSuccessEnvelope(t, postAction(t, harness, admin, prefix+"/disable", "authentic-disable", "fixture lifecycle"), http.StatusOK)
	if fallback := consumer.ResolveCurrentPackSet(ctx); fallback.Error != nil || fallback.Value.ID != before.Value.ID {
		t.Fatal("disablement fallback", fallback.Error)
	}
	if old := consumer.GetPackEntry(ctx, reference_data.GetPackEntryRequest{PackSetID: active.Value.ID, PackKey: fixture.Key, EntryID: "attack-pattern--7385dfaf-6886-4229-9ecd-6fd678040830"}); old.Error != nil || old.Value.Provenance.ManifestSHA256 != fixture.ManifestSHA256 {
		t.Fatal("disablement changed historical content", old.Error)
	}
	httptestx.RequireSuccessEnvelope(t, postAction(t, harness, admin, prefix+"/remove", "authentic-remove", "fixture retention complete"), http.StatusOK)
	if old := consumer.GetPackEntry(ctx, reference_data.GetPackEntryRequest{PackSetID: active.Value.ID, PackKey: fixture.Key, EntryID: "attack-pattern--7385dfaf-6886-4229-9ecd-6fd678040830"}); old.Error == nil {
		t.Fatal("removed content remained consumable")
	}
}
