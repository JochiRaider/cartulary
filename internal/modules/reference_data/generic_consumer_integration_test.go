package reference_data_test

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/app/referenceassembly"
	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data/contractadmission"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/httptestx"
)

func TestUnregisteredGenericDatasetLifecycle_Integration(t *testing.T) {
	for _, kind := range []string{"framework", "enrichment"} {
		t.Run(kind, func(t *testing.T) {
			runGenericProfileConsumer(t, "advisory.unregistered_fifth_dataset", kind)
			t.Run("recovery", func(t *testing.T) { runUnregisteredGenericRecovery(t, kind) })
		})
	}
}

// This exercises the same importer and consumer for standardized and previously
// unknown keys. Expected queries are authored here, not obtained by normalizing
// the indexed values or invoking a source producer.
func runGenericProfileConsumer(t *testing.T, key string, kinds ...string) {
	kind := ""
	if len(kinds) != 0 {
		kind = kinds[0]
	}
	t.Helper()
	var fixture canonicalProfileFixture
	unknown := key == "advisory.unregistered_fifth_dataset"
	if unknown {
		fixture = unregisteredGenericFixture(t, key)
	} else {
		fixture = canonicalProfileFixtureFor(t, key)
	}
	ctx := context.Background()
	runtime := appsupport.StartRuntime(t)
	harness := startReferencePackServer(t, runtime, "generic-profile-consumer")
	admin, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, harness.Server.HTTP.URL)
	storage, err := referenceassembly.NewRootStorage(harness.Server.Config.Roots.TemporaryWork.Path, harness.Server.Config.Roots.ReferencePackStorage.Path)
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()
	now := time.Now().UTC()
	consumer, err := reference_data.NewConsumer(harness.Pool, storage, pagination.NewCodec([32]byte{1, 2, 3}), func() time.Time { return now }, consumerIntegrityOptions(t))
	if err != nil {
		t.Fatal(err)
	}
	before := consumer.ResolveCurrentPackSet(ctx)
	if before.Error != nil || len(before.Value.Members) != 3 {
		t.Fatal("required Base set", before.Error)
	}
	base := consumer.GetPackDescriptor(ctx, reference_data.GetPackDescriptorRequest{PackSetID: before.Value.ID, PackKey: "type_registry.host"})
	assertConsumerContract(t, "GetPackDescriptor", base)
	if base.Error != nil || base.Value.Descriptor != nil {
		t.Fatal("non-dataset descriptor must be null", base.Error)
	}
	checkError := func(got *reference_data.ConsumerError, code string) {
		t.Helper()
		if got == nil || got.Code != code {
			t.Fatalf("want %s, got %#v", code, got)
		}
	}
	checkError(consumer.GetPackDescriptor(ctx, reference_data.GetPackDescriptorRequest{PackSetID: before.Value.ID, PackKey: key}).Error, "pack_unavailable")
	container := referencePackBundle(t, bundleOptions{PackKey: key, PackKind: kind, PackVersion: "1", Members: fixture.Members})
	response := postReferencePackUpload(t, harness.Server.HTTP.URL, admin, `{"client_txn_id":"generic-import"}`, container, "generic.zip", reference_data.MediaTypeZip)
	job := httptestx.RequireSuccessEnvelope(t, response, http.StatusAccepted)["data"].(map[string]any)
	if terminal := requireJob(t, harness, admin, job["job_id"].(string)); terminal["status"] != "succeeded" {
		t.Fatalf("generic import: %#v", terminal)
	}
	if set := consumer.ResolveCurrentPackSet(ctx); set.Error != nil || set.Value.ID != before.Value.ID {
		t.Fatal("import activated pack", set.Error)
	}
	prefix := "/api/v1/reference-packs/" + key + "/1"
	httptestx.RequireSuccessEnvelope(t, postAction(t, harness, admin, prefix+"/activate", "generic-activate", ""), http.StatusOK)
	selected := consumer.ResolveCurrentPackSet(ctx)
	if selected.Error != nil || len(selected.Value.Members) != 4 {
		t.Fatal("generic plus Base set", selected.Error)
	}
	descriptorRequest := reference_data.GetPackDescriptorRequest{PackSetID: selected.Value.ID, PackKey: key}
	descriptor := consumer.GetPackDescriptor(ctx, descriptorRequest)
	assertConsumerContract(t, "GetPackDescriptor", descriptor)
	if descriptor.Error != nil || string(descriptor.Value.Descriptor) != fixture.Members["payload/descriptor.json"] || descriptor.Value.Provenance.PackSetID != selected.Value.ID || descriptor.Value.Provenance.PackKey != key || descriptor.Value.Provenance.PackVersion != "1" {
		t.Fatal("descriptor bytes/provenance", descriptor.Error)
	}
	request := reference_data.LookupPackEntriesRequest{PackSetID: selected.Value.ID, PackKey: key, LookupKind: "label", LookupValue: "FIXTURE"}
	page := consumer.LookupPackEntries(ctx, request)
	assertConsumerContract(t, "LookupPackEntries", page)
	expected := strings.Split(strings.TrimSuffix(fixture.Members["payload/objects.ndjson"], "\n"), "\n")
	if page.Error != nil || len(page.Value.Items) != len(expected) {
		t.Fatal("generic normalized lookup", page.Error)
	}
	for i, row := range expected {
		if string(page.Value.Items[i]) != row {
			t.Fatal("inert data or deterministic order changed")
		}
		var object struct {
			ID string `json:"object_id"`
		}
		if err := json.Unmarshal([]byte(row), &object); err != nil {
			t.Fatal(err)
		}
		entry := consumer.GetPackEntry(ctx, reference_data.GetPackEntryRequest{PackSetID: selected.Value.ID, PackKey: key, EntryID: object.ID})
		assertConsumerContract(t, "GetPackEntry", entry)
		if entry.Error != nil || string(entry.Value.Item) != row {
			t.Fatal("generic identity lookup", entry.Error)
		}
	}
	for _, tc := range []struct {
		kind  string
		value any
		code  string
	}{
		{"unknown", "x", "lookup_kind_unsupported"}, {"label", 42, "invalid_pack_request"},
	} {
		changed := request
		changed.LookupKind, changed.LookupValue = tc.kind, tc.value
		checkError(consumer.LookupPackEntries(ctx, changed).Error, tc.code)
	}
	missing := request
	missing.LookupValue = "absent"
	if got := consumer.LookupPackEntries(ctx, missing); got.Error != nil || len(got.Value.Items) != 0 {
		t.Fatal("no-hit is not empty success", got.Error)
	}
	badCursor := "invalid"
	bad := request
	bad.Cursor = &badCursor
	checkError(consumer.LookupPackEntries(ctx, bad).Error, "cursor_invalid")
	if unknown {
		expectedAuthority := "advisory_enrichment"
		if kind == "framework" {
			expectedAuthority = "framework_reference"
		}
		if descriptor.Value.Provenance.AuthorityClass != expectedAuthority {
			t.Fatal("manifest authority", descriptor.Value.Provenance.AuthorityClass)
		}
		limit := 1
		paginated := request
		paginated.Limit = &limit
		first := consumer.LookupPackEntries(ctx, paginated)
		if first.Error != nil || len(first.Value.Items) != 1 || first.Value.NextCursor == nil {
			t.Fatal("first page", first.Error)
		}
		paginated.Cursor = first.Value.NextCursor
		second := consumer.LookupPackEntries(ctx, paginated)
		if second.Error != nil || len(second.Value.Items) != 1 || second.Value.NextCursor != nil || string(second.Value.Items[0]) != expected[1] {
			t.Fatal("second page", second.Error)
		}
		mismatch := paginated
		mismatch.LookupValue = "absent"
		checkError(consumer.LookupPackEntries(ctx, mismatch).Error, "cursor_query_mismatch")
		now = now.Add(15 * time.Minute)
		checkError(consumer.LookupPackEntries(ctx, paginated).Error, "cursor_invalid")
		now = now.Add(-15 * time.Minute)
		composite := request
		composite.LookupKind, composite.LookupValue = "provider_number", map[string]any{"provider": "PROVIDER", "number": json.Number("42")}
		if got := consumer.LookupPackEntries(ctx, composite); got.Error != nil || len(got.Value.Items) != 2 {
			t.Fatal("composite lookup", got.Error)
		}
		composite.LookupValue = map[string]any{"provider": "PROVIDER", "number": 42, "version": 0}
		checkError(consumer.LookupPackEntries(ctx, composite).Error, "invalid_pack_request")
		composite.LookupKind, composite.LookupValue = "unused", "absent"
		if got := consumer.LookupPackEntries(ctx, composite); got.Error != nil || len(got.Value.Items) != 0 {
			t.Fatal("zero coverage lookup", got.Error)
		}
	}
	httptestx.RequireSuccessEnvelope(t, postAction(t, harness, admin, prefix+"/disable", "generic-disable", ""), http.StatusOK)
	current := consumer.ResolveCurrentPackSet(ctx)
	if current.Error != nil || current.Value.ID != before.Value.ID {
		t.Fatal("disable did not restore Base set", current.Error)
	}
	currentRequest := request
	currentRequest.PackSetID = current.Value.ID
	checkError(consumer.LookupPackEntries(ctx, currentRequest).Error, "pack_unavailable")
	historical := consumer.GetPackDescriptor(ctx, descriptorRequest)
	if historical.Error != nil || string(historical.Value.Descriptor) != string(descriptor.Value.Descriptor) || historical.Value.Provenance.ManifestSHA256 != descriptor.Value.Provenance.ManifestSHA256 {
		t.Fatal("historical descriptor changed", historical.Error)
	}
	if got := consumer.LookupPackEntries(ctx, request); got.Error != nil || len(got.Value.Items) != len(expected) {
		t.Fatal("historical lookup lost", got.Error)
	}
	if unknown {
		var original []byte
		injectProvenance := func(raw []byte) {
			t.Helper()
			tx, err := harness.DB.BeginTx(ctx, nil)
			if err != nil {
				t.Fatal(err)
			}
			defer tx.Rollback()
			if _, err := tx.ExecContext(ctx, `ALTER TABLE reference_pack_sets DISABLE TRIGGER USER`); err != nil {
				t.Fatal(err)
			}
			if _, err := tx.ExecContext(ctx, `UPDATE reference_pack_sets SET canonical_provenance=$2 WHERE pack_set_id=$1`, selected.Value.ID, raw); err != nil {
				t.Fatal(err)
			}
			if _, err := tx.ExecContext(ctx, `ALTER TABLE reference_pack_sets ENABLE TRIGGER USER`); err != nil {
				t.Fatal(err)
			}
			if err := tx.Commit(); err != nil {
				t.Fatal(err)
			}
		}
		if err := harness.Pool.QueryRow(ctx, `SELECT canonical_provenance FROM reference_pack_sets WHERE pack_set_id=$1`, selected.Value.ID).Scan(&original); err != nil {
			t.Fatal(err)
		}
		var anchors []map[string]any
		if err := json.Unmarshal(original, &anchors); err != nil {
			t.Fatal(err)
		}
		for _, anchor := range anchors {
			if anchor["pack_key"] == key {
				anchor["authority_class"] = "framework_reference"
				if kind == "framework" {
					anchor["authority_class"] = "advisory_enrichment"
				}
			}
		}
		injectProvenance(integrationCanonical(t, anchors))
		checkError(consumer.GetPackDescriptor(ctx, descriptorRequest).Error, "pack_unavailable")
		if err := contractadmission.Validate(ctx, harness.Pool); !contractadmission.IsIncompatible(err) {
			t.Fatal("incorrect local authority admitted", err)
		}
		injectProvenance(original)
		httptestx.RequireSuccessEnvelope(t, postAction(t, harness, admin, prefix+"/activate", "generic-reactivate", ""), http.StatusOK)
		name := storedPackMemberPath(t, harness, key, "1", "payload/descriptor.json")
		damaged := []byte(fixture.Members["payload/descriptor.json"])
		damaged[len(damaged)-1] = '!'
		if err := os.WriteFile(name, damaged, 0600); err != nil {
			t.Fatal(err)
		}
		checkError(consumer.GetPackDescriptor(ctx, descriptorRequest).Error, "pack_unavailable")
		if after := consumer.ResolveCurrentPackSet(ctx); after.Error != nil || after.Value.ID != before.Value.ID {
			t.Fatal("descriptor corruption left active content", after.Error)
		}
		if count := queryCount(t, harness.DB, `SELECT count(*) FROM reference_pack_events WHERE pack_key=$1 AND pack_version='1' AND event_kind='payload_invalidation'`, key); count != 1 {
			t.Fatal("descriptor invalidation not published exactly once", count)
		}
	}

}

func unregisteredGenericFixture(t testing.TB, key string) canonicalProfileFixture {
	t.Helper()
	fixture := canonicalProfileFixture{Key: key, Members: map[string]string{
		"payload/descriptor.json":      `{"lookups":[{"components":[{"name":"value","normalization":"ascii_case_insensitive","type":"string"}],"coverage":{"distinct_keys":1,"objects":2,"postings":2},"input_type":"scalar","name":"label","unique":false},{"components":[{"name":"number","normalization":"safe_integer","type":"integer"},{"name":"provider","normalization":"ascii_case_insensitive","type":"string"}],"coverage":{"distinct_keys":1,"objects":2,"postings":2},"input_type":"composite","name":"provider_number","unique":false},{"components":[{"name":"value","normalization":"exact_nfc","type":"string"}],"coverage":{"distinct_keys":0,"objects":0,"postings":0},"input_type":"scalar","name":"unused","unique":true}],"schema_id":"cartulary.reference_pack.dataset.descriptor.v1"}`,
		"payload/relationships.ndjson": "",
	}}
	for _, id := range []string{"example:1", strings.Repeat("z", 512)} {
		row := map[string]any{"schema_id": "cartulary.reference_pack.dataset.object.v1", "object_id": id, "object_type": "example", "display_label": "Fixture", "description": nil, "deprecated": false, "replacement_object_id": nil, "source_refs": []any{map[string]any{"artifact_index": 0, "locator": "fixture"}}, "data": map[string]any{"remote_schema": "https://never-fetch.invalid/schema", "script": "inert text"}, "lookup_values": map[string]any{"label": []any{"Fixture"}, "provider_number": []any{map[string]any{"provider": "Provider", "number": 42}}}}
		fixture.Members["payload/objects.ndjson"] += string(integrationCanonical(t, row)) + "\n"
	}
	return fixture
}
