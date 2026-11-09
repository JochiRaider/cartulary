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
	referencetest "github.com/JochiRaider/cartulary/internal/modules/reference_data/testsupport"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
	"github.com/JochiRaider/cartulary/internal/testutil/referencepackfixture"
)

func TestEntraCanonicalProductionLifecycle_Integration(t *testing.T) { runEntraLifecycle(t, false) }

func TestEntraRealReleaseImport_Integration(t *testing.T) { runEntraLifecycle(t, true) }

func runEntraLifecycle(t *testing.T, realRelease bool) {
	expected := referencetest.OwnerFixture(t, "entra_lifecycle")
	if expected.Condition == nil || *expected.Condition != "verified_available" {
		t.Fatal("Entra fixture expectation")
	}
	ctx := context.Background()
	now := time.Now().UTC()
	var err error
	var keys *referencepackfixture.SigningContext
	var a referencetest.ProducedEntraPack
	if realRelease {
		a = referencetest.EntraRelease(t, "../../..")
	} else {
		keys, err = referencepackfixture.NewSigningContext("fixture.entra.dataset.v1", "enrichment.entra_app_ids")
		if err != nil {
			t.Fatal(err)
		}
		a = referencetest.EntraFixture(t, "fixture-a", now, keys, 1)
	}
	bootstrap := filepath.Join(t.TempDir(), "bootstrap.json")
	if err = os.WriteFile(bootstrap, a.Bootstrap, 0600); err != nil {
		t.Fatal(err)
	}
	runtime := appsupport.StartRuntime(t)
	harness := runtime.StartServer(t, appsupport.ServerOptions{Prefix: "entra-online", TestRouteMode: httptestx.TestRouteModeDisabled, Env: map[string]string{"CARTULARY__REFERENCE_PACKS__TRUST_BOOTSTRAP_PATH": bootstrap, "CARTULARY__REFERENCE_PACKS__CLOCK_TRUSTED": "true"}})
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
	upload := func(f referencetest.ProducedEntraPack, txn string) {
		response := postReferencePackUpload(t, harness.Server.HTTP.URL, admin, `{"client_txn_id":"`+txn+`"}`, f.Container, "entra.tar", "application/x-tar")
		job := httptestx.RequireSuccessEnvelope(t, response, http.StatusAccepted)["data"].(map[string]any)
		if terminal := requireJob(t, harness, admin, job["job_id"].(string)); terminal["status"] != "succeeded" {
			t.Fatalf("Entra producer import: %#v", terminal)
		}
	}
	activate := func(f referencetest.ProducedEntraPack, txn string) {
		httptestx.RequireSuccessEnvelope(t, postAction(t, harness, admin, "/api/v1/reference-packs/"+f.Key+"/"+f.Version+"/activate", txn, ""), http.StatusOK)
	}
	upload(a, "entra-import-a")
	if realRelease {
		for kind, want := range map[string]int{"object": 139, "relationship": 0} {
			if got := queryCount(t, harness.DB, `SELECT count(*) FROM reference_pack_indexes i
                JOIN reference_pack_candidates c ON c.current_index_id=i.index_id
                WHERE c.pack_key=$1 AND c.pack_version=$2 AND i.entry_kind=$3`, a.Key, a.Version, kind); got != want {
				t.Fatalf("retained %s count=%d, want %d", kind, got, want)
			}
		}
	}
	if staged := consumer.ResolveCurrentPackSet(ctx); staged.Error != nil || staged.Value.ID != before.Value.ID {
		t.Fatal("staging changed activation")
	}
	activate(a, "entra-activate-a")
	selected := consumer.ResolveCurrentPackSet(ctx)
	if selected.Error != nil || len(selected.Value.Members) != 4 {
		t.Fatal("Entra plus Base set", selected.Error)
	}
	if selected.Value.ID != "rpset_"+referencetest.EntraSetDigest(t, "../../..", a) {
		t.Fatal("independent exact-set identity")
	}
	t.Logf("Entra A exact set=%s", selected.Value.ID)
	provenance := consumer.GetPackProvenance(ctx, reference_data.GetPackProvenanceRequest{PackSetID: selected.Value.ID, PackKey: a.Key})
	assertConsumerContract(t, "GetPackProvenance", provenance)
	if provenance.Error != nil || provenance.Value.ManifestSHA256 != a.ManifestSHA256 || provenance.Value.PayloadSHA256 != a.PayloadSHA256 {
		t.Fatal("exact pack provenance", provenance.Error)
	}
	lookup := func(kind, value string, want []string) {
		t.Helper()
		page := consumer.LookupPackEntries(ctx, reference_data.LookupPackEntriesRequest{PackSetID: selected.Value.ID, PackKey: a.Key, LookupKind: kind, LookupValue: value})
		assertConsumerContract(t, "LookupPackEntries", page)
		if page.Error != nil {
			t.Fatal(kind, value, page.Error)
		}
		ids := []string{}
		for _, raw := range page.Value.Items {
			var row struct {
				ID string `json:"object_id"`
			}
			if err = json.Unmarshal(raw, &row); err != nil {
				t.Fatal(err)
			}
			ids = append(ids, row.ID)
		}
		if !slices.Equal(ids, want) || page.Value.NextCursor != nil {
			t.Fatal("lookup identity", kind, value, ids, want)
		}
	}
	// Every canonical byte and field-level source locator survives the real owner interfaces.
	for _, line := range strings.Split(strings.TrimSuffix(string(a.Members["payload/objects.ndjson"]), "\n"), "\n") {
		var row struct {
			ID      string `json:"object_id"`
			Label   string `json:"display_label"`
			Lookups struct {
				Aliases []string `json:"alias"`
			} `json:"lookup_values"`
		}
		if err = json.Unmarshal([]byte(line), &row); err != nil {
			t.Fatal(err)
		}
		entry := consumer.GetPackEntry(ctx, reference_data.GetPackEntryRequest{PackSetID: selected.Value.ID, PackKey: a.Key, EntryID: row.ID})
		assertConsumerContract(t, "GetPackEntry", entry)
		if entry.Error != nil || string(entry.Value.Item) != line || entry.Value.Provenance.ManifestSHA256 != a.ManifestSHA256 {
			t.Fatal("canonical entry/provenance", row.ID, entry.Error)
		}
		lookup("app_id", strings.ToUpper(row.ID), []string{row.ID})
		for _, alias := range row.Lookups.Aliases {
			lookup("alias", strings.ToUpper(alias), []string{row.ID})
		}
		if len(row.Lookups.Aliases) == 0 {
			lookup("alias", row.Label, []string{})
		}
	}
	lookup("app_id", "ffffffff-ffff-ffff-ffff-ffffffffffff", []string{})
	lookup("service_principal_name", "https://example.test/Absent/", []string{})
	errorCode := func(got *reference_data.ConsumerError, want string) {
		t.Helper()
		if got == nil || got.Code != want {
			t.Fatalf("expected %s, got %#v", want, got)
		}
	}
	for _, tc := range []struct{ kind, value, want string }{{"app_id", "not-a-uuid", "invalid_pack_request"}, {"unknown", "x", "lookup_kind_unsupported"}} {
		got := consumer.LookupPackEntries(ctx, reference_data.LookupPackEntriesRequest{PackSetID: selected.Value.ID, PackKey: a.Key, LookupKind: tc.kind, LookupValue: tc.value})
		errorCode(got.Error, tc.want)
	}
	errorCode(consumer.GetPackEntry(ctx, reference_data.GetPackEntryRequest{PackSetID: selected.Value.ID, PackKey: a.Key, EntryID: "ffffffff-ffff-ffff-ffff-ffffffffffff"}).Error, "entry_not_found")
	errorCode(consumer.GetPackProvenance(ctx, reference_data.GetPackProvenanceRequest{PackSetID: "rpset_" + strings.Repeat("f", 64), PackKey: a.Key}).Error, "pack_set_not_found")
	errorCode(consumer.GetPackProvenance(ctx, reference_data.GetPackProvenanceRequest{PackSetID: selected.Value.ID, PackKey: "enrichment.tor"}).Error, "pack_unavailable")
	cursor := "invalid"
	errorCode(consumer.LookupPackEntries(ctx, reference_data.LookupPackEntriesRequest{PackSetID: selected.Value.ID, PackKey: a.Key, LookupKind: "alias", LookupValue: "Microsoft Graph", Cursor: &cursor}).Error, "cursor_invalid")
	httptestx.RequireSuccessEnvelope(t, postAction(t, harness, admin, "/api/v1/reference-packs/"+a.Key+"/"+a.Version+"/disable", "entra-disable", ""), http.StatusOK)
	if disabled := consumer.ResolveCurrentPackSet(ctx); disabled.Error != nil || disabled.Value.ID != before.Value.ID {
		t.Fatal("disable did not return exact Base set", disabled.Error)
	}
	lookup("alias", "Microsoft Graph", []string{"00000003-0000-0000-c000-000000000000"})
	if realRelease {
		return
	}
	b := referencetest.EntraFixture(t, "fixture-b", now, keys, 2)
	upload(b, "entra-import-b")
	activate(b, "entra-activate-b")
	current := consumer.ResolveCurrentPackSet(ctx)
	if current.Error != nil || current.Value.ID != "rpset_"+referencetest.EntraSetDigest(t, "../../..", b) || current.Value.ID == selected.Value.ID {
		t.Fatal("synthetic B identity", current.Error)
	}
	renewed := referencetest.EntraFixture(t, "fixture-b", now, keys, 3)
	if renewed.ManifestSHA256 != b.ManifestSHA256 || renewed.PayloadSHA256 != b.PayloadSHA256 {
		t.Fatal("renewal changed logical identity")
	}
	upload(renewed, "entra-renew-b")
	if after := consumer.ResolveCurrentPackSet(ctx); after.Error != nil || after.Value.ID != current.Value.ID {
		t.Fatal("renewal changed current set")
	}
	checkBytes, err := os.ReadFile("../../../contracts/reference-packs/checks.v1.json")
	if err != nil {
		t.Fatal(err)
	}
	var checkRegistry struct {
		Checks []struct {
			ID   string `json:"check_id"`
			Code string `json:"code"`
		} `json:"checks"`
	}
	if err = json.Unmarshal(checkBytes, &checkRegistry); err != nil {
		t.Fatal(err)
	}
	codes := map[string]string{}
	for _, check := range checkRegistry.Checks {
		codes[check.ID] = check.Code
	}
	for i, tc := range []struct{ version, check string }{{"negative-v1-entry", "content_schema"}, {"negative-identity", "content_semantics"}, {"negative-provenance", "content_semantics"}} {
		negative := referencetest.EntraFixture(t, tc.version, now, keys, int64(4+i))
		response := postReferencePackUpload(t, harness.Server.HTTP.URL, admin, `{"client_txn_id":"`+tc.version+`"}`, negative.Container, "negative.tar", "application/x-tar")
		job := httptestx.RequireSuccessEnvelope(t, response, http.StatusAccepted)["data"].(map[string]any)
		terminal := requireJob(t, harness, admin, job["job_id"].(string))
		if terminal["status"] != "failed" || terminal["error_summary"].(map[string]any)["details"].(map[string]any)["reason_code"] != codes[tc.check] || terminal["error_summary"].(map[string]any)["details"].(map[string]any)["check_id"] != tc.check {
			t.Fatalf("signed negative %s: %#v", tc.version, terminal)
		}
		if got := queryCount(t, harness.DB, `SELECT count(*) FROM reference_pack_index_generations WHERE pack_key=$1 AND pack_version=$2`, a.Key, tc.version); got != 0 {
			t.Fatal("failed import published index")
		}
		if after := consumer.ResolveCurrentPackSet(ctx); after.Error != nil || after.Value.ID != current.Value.ID {
			t.Fatal("failed import changed current set")
		}
	}

}
