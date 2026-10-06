package reference_data_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/app/referenceassembly"
	"github.com/JochiRaider/cartulary/internal/modules/auth/testsupport/flowtest"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/JochiRaider/cartulary/internal/platform/pagination"
	"github.com/JochiRaider/cartulary/internal/testutil/appsupport"
)

func TestAllCanonicalProfilesThroughProductionConsumers_Integration(t *testing.T) {
	ctx := context.Background()
	runtime := appsupport.StartRuntime(t)
	harness := startReferencePackServer(t, runtime, "reference-pack-all-profiles")
	admin, _ := flowtest.ProvisionBootstrapAdmin(t, http.DefaultClient, harness.Server.HTTP.URL)
	storage, err := referenceassembly.NewRootStorage(harness.Server.Config.Roots.TemporaryWork.Path, harness.Server.Config.Roots.ReferencePackStorage.Path)
	if err != nil {
		t.Fatal(err)
	}
	defer storage.Close()
	consumer, err := reference_data.NewConsumer(harness.Pool, storage, pagination.NewCodec([32]byte([]byte(strings.Repeat("v", 32)))), func() time.Time { return time.Now().UTC() }, consumerIntegrityOptions(t))
	if err != nil {
		t.Fatal(err)
	}
	var fixtures struct {
		Cases []struct {
			Key     string            `json:"pack_key"`
			Members map[string]string `json:"members"`
		} `json:"cases"`
	}
	raw, err := os.ReadFile("../../../contracts/reference-pack-fixtures/fixtures/profiles.v1.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(raw, &fixtures); err != nil {
		t.Fatal(err)
	}
	if len(fixtures.Cases) != 16 {
		t.Fatal("profile fixture catalog incomplete")
	}
	for _, fixture := range fixtures.Cases {
		t.Run(fixture.Key, func(t *testing.T) {
			options := bundleOptions{PackKey: fixture.Key, PackVersion: "1"}
			if strings.HasPrefix(fixture.Key, "framework.") {
				options.ObjectsTransform = func(raw []byte) []byte {
					lines := strings.Split(strings.TrimSuffix(string(raw), "\n"), "\n")
					for i, line := range lines {
						var row map[string]any
						if err := json.Unmarshal([]byte(line), &row); err != nil {
							t.Fatal(err)
						}
						row["external_refs"] = []any{map[string]any{"source_name": "a", "external_id": "shared", "url": nil}, map[string]any{"source_name": "a", "external_id": "shared", "url": "https://example.com/"}, map[string]any{"source_name": "b", "external_id": "shared", "url": nil}}
						lines[i] = string(integrationCanonical(t, row))
					}
					return []byte(strings.Join(lines, "\n") + "\n")
				}
				fixture.Members["payload/objects.ndjson"] = string(options.ObjectsTransform([]byte(fixture.Members["payload/objects.ndjson"])))
			}
			uploaded := postReferencePackUpload(t, harness.Server.HTTP.URL, admin, fmt.Sprintf(`{"client_txn_id":"all-import-%s"}`, fixture.Key), referencePackBundle(t, options), "fixture.zip", reference_data.MediaTypeZip)
			job := requireSuccessEnvelope(t, uploaded, http.StatusAccepted)["data"].(map[string]any)
			if terminal := requireJob(t, harness, admin, job["job_id"].(string)); terminal["status"] != "succeeded" {
				t.Fatalf("profile import: %#v", terminal)
			}
			requireSuccessEnvelope(t, postAction(t, harness, admin, "/api/v1/reference-packs/"+fixture.Key+"/1/activate", "all-activate-"+fixture.Key, ""), http.StatusOK)
			set := consumer.ResolveCurrentPackSet(ctx)
			assertConsumerContract(t, "ResolveCurrentPackSet", set)
			if set.Error != nil {
				t.Fatal(set.Error)
			}
			provenance := consumer.GetPackProvenance(ctx, reference_data.GetPackProvenanceRequest{PackSetID: set.Value.ID, PackKey: fixture.Key})
			assertConsumerContract(t, "GetPackProvenance", provenance)
			path := "payload/entries.ndjson"
			idField := "entry_id"
			if strings.HasPrefix(fixture.Key, "framework.") {
				path = "payload/objects.ndjson"
				idField = "object_id"
			}
			for _, line := range strings.Split(strings.TrimSuffix(fixture.Members[path], "\n"), "\n") {
				var row map[string]any
				decoder := json.NewDecoder(strings.NewReader(line))
				decoder.UseNumber()
				if err := decoder.Decode(&row); err != nil {
					t.Fatal(err)
				}
				id := row[idField].(string)
				got := consumer.GetPackEntry(ctx, reference_data.GetPackEntryRequest{PackSetID: set.Value.ID, PackKey: fixture.Key, EntryID: id})
				assertConsumerContract(t, "GetPackEntry", got)
				if got.Error != nil || got.Value == nil {
					t.Fatalf("entry %s: %#v", id, got.Error)
				}
				var returned map[string]any
				if err := json.Unmarshal(got.Value.Item, &returned); err != nil {
					t.Fatal(err)
				}
				if returned[idField] != id || got.Value.Provenance.PackSetID != set.Value.ID || got.Value.Provenance.PackVersion != "1" {
					t.Fatal("entry identity/provenance mismatch")
				}
				if strings.HasPrefix(fixture.Key, "framework.") && string(got.Value.Item) != line {
					t.Fatal("framework reference provenance changed")
				}
				kind := idField
				var value any = id
				switch fixture.Key {
				case "enrichment.tor":
					kind = "network"
					value = row["network"]
				case "enrichment.cisa_kev":
					kind = "cve_id"
				case "enrichment.ms_portals":
					kind = "service_id"
				case "enrichment.windows_event_ids":
					kind = "event_id"
					value = row["event_id"]
				case "enrichment.entra_app_ids":
					kind = "app_id"
				case "enrichment.loldrivers":
					kind = "driver_id"
				case "enrichment.hijacklibs":
					kind = "library_name"
					value = row["library_name"]
				case "enrichment.windows_sids":
					kind = "sid"
				}
				page := consumer.LookupPackEntries(ctx, reference_data.LookupPackEntriesRequest{PackSetID: set.Value.ID, PackKey: fixture.Key, LookupKind: kind, LookupValue: value})
				assertConsumerContract(t, "LookupPackEntries", page)
				expectedCount := 1
				if fixture.Key == "enrichment.windows_event_ids" {
					expectedCount = 2
				}
				if page.Error != nil || page.Value == nil || len(page.Value.Items) != expectedCount {
					t.Fatalf("lookup %s %s: %#v", fixture.Key, kind, page)
				}
			}
			if strings.HasPrefix(fixture.Key, "framework.") {
				limit := 1
				request := reference_data.LookupPackEntriesRequest{PackSetID: set.Value.ID, PackKey: fixture.Key, LookupKind: "external_id", LookupValue: "shared", Limit: &limit}
				for _, want := range []string{"fixture:1", "fixture:2"} {
					page := consumer.LookupPackEntries(ctx, request)
					if page.Error != nil || page.Value == nil || len(page.Value.Items) != 1 {
						t.Fatalf("external ID page: %#v", page)
					}
					var item map[string]any
					if err := json.Unmarshal(page.Value.Items[0], &item); err != nil {
						t.Fatal(err)
					}
					if item["object_id"] != want || len(item["external_refs"].([]any)) != 3 {
						t.Fatal("lookup duplicated object or lost references", item)
					}
					request.Cursor = page.Value.NextCursor
					if (want == "fixture:2") != (request.Cursor == nil) {
						t.Fatal("incorrect final page")
					}
				}
				if got := queryCount(t, harness.DB, `SELECT count(*) FROM reference_pack_lookup_keys k JOIN reference_pack_candidates c ON c.current_index_id=k.index_id WHERE c.pack_key=$1 AND k.lookup_kind='external_id'`, fixture.Key); got != 2 {
					t.Fatalf("persisted external keys=%d", got)
				}
				invalid := bundleOptions{PackKey: fixture.Key, PackVersion: "2", ObjectsTransform: func(raw []byte) []byte {
					lines := strings.Split(strings.TrimSuffix(string(options.ObjectsTransform(raw)), "\n"), "\n")
					var row map[string]any
					if err := json.Unmarshal([]byte(lines[0]), &row); err != nil {
						t.Fatal(err)
					}
					refs := row["external_refs"].([]any)
					row["external_refs"] = append([]any{refs[0]}, refs...)
					lines[0] = string(integrationCanonical(t, row))
					return []byte(strings.Join(lines, "\n") + "\n")
				}}
				rejected := postReferencePackUpload(t, harness.Server.HTTP.URL, admin, fmt.Sprintf(`{"client_txn_id":"duplicate-ref-%s"}`, fixture.Key), referencePackBundle(t, invalid), "duplicate.zip", reference_data.MediaTypeZip)
				rejectedJob := requireSuccessEnvelope(t, rejected, http.StatusAccepted)["data"].(map[string]any)
				terminal := requireJob(t, harness, admin, rejectedJob["job_id"].(string))
				if terminal["status"] != "failed" || terminal["error_summary"].(map[string]any)["details"].(map[string]any)["reason_code"] != "content_semantic_invalid" {
					t.Fatalf("exact duplicate reference accepted: %#v", terminal)
				}
				if got := queryCount(t, harness.DB, `SELECT count(*) FROM reference_pack_index_generations WHERE pack_key=$1 AND pack_version='2'`, fixture.Key); got != 0 {
					t.Fatal("failed import published a partial index")
				}
			}
		})
	}
	set := consumer.ResolveCurrentPackSet(ctx)
	assertConsumerContract(t, "ResolveCurrentPackSet", set)
	if set.Error != nil || len(set.Value.Members) != 16 {
		t.Fatal("complete reachable set not published")
	}
	// Independently stated canonical values cross the actual index/registry
	// boundary, including transformations absent from the retired canonicalizer.
	for _, v := range []struct{ typ, raw, want string }{
		{"ipv4_addr", " 192.0.2.7\u00a0", "192.0.2.7"},
		{"ipv6_addr", "2001:0DB8:0:0:1:0:0:1", "2001:db8::1:0:0:1"},
		{"domain_name", "EXAMPLE.COM.", "example.com"},
		{"url", "HTTPS://Example.COM:0443/a/%2e%2E/b/%7e?q=%2f&x=%7e#", "https://example.com/b/~?q=%2F&x=~#"},
		{"sha256", strings.Repeat("AB", 32), strings.Repeat("ab", 32)},
		{"email_addr", "User+Case@EXAMPLE.COM.", "User+Case@example.com"},
		{"registry_key", `hklm//Software\Vendor\`, `HKEY_LOCAL_MACHINE\SOFTWARE\VENDOR`},
		{"process_name", " Cafe\u0301.exe ", "Café.exe"},
		{"text", " A\r\nB\tC\rD ", "A\nB\tC\nD"},
	} {
		t.Run(v.typ, func(t *testing.T) {
			request := reference_data.EvaluateIndicatorRequest{PackSetID: set.Value.ID, IndicatorTypeID: v.typ, ValueKind: "atomic", RawValue: v.raw}
			got := consumer.EvaluateIndicatorValue(ctx, request)
			assertConsumerContract(t, "EvaluateIndicatorValue", got)
			if got.Error != nil || got.Value == nil || !got.Value.Valid || got.Value.Normalized == nil || *got.Value.Normalized != v.want {
				t.Fatalf("evaluation=%#v", got)
			}
			request.RawValue = v.want
			again := consumer.EvaluateIndicatorValue(ctx, request)
			assertConsumerContract(t, "EvaluateIndicatorValue", again)
			if again.Error != nil || again.Value == nil || !again.Value.Valid || *again.Value.DedupeKey != *got.Value.DedupeKey || *again.Value.Normalized != v.want {
				t.Fatal("normalization is not idempotent through consumer")
			}
		})
	}
	// Addressability is checked after the real importer builds the index. The
	// canonical inputs independently state the longest profile identities.
	for _, name := range []string{"enrichment_windows_event_ids", "enrichment_cisa_kev", "enrichment_windows_sids"} {
		t.Run(name+" maximum identity", func(t *testing.T) {
			data, err := os.ReadFile("../../../contracts/reference-pack-fixtures/fixtures/profiles/" + name + "_boundary.v1.json")
			if err != nil {
				t.Fatal(err)
			}
			var boundary struct {
				Key     string            `json:"pack_key"`
				Members map[string]string `json:"members"`
			}
			if err := json.Unmarshal(data, &boundary); err != nil {
				t.Fatal(err)
			}
			payload := []byte(boundary.Members["payload/entries.ndjson"])
			container := referencePackBundle(t, bundleOptions{PackKey: boundary.Key, PackVersion: "2", PayloadTransform: func([]byte) []byte { return payload }})
			response := postReferencePackUpload(t, harness.Server.HTTP.URL, admin, fmt.Sprintf(`{"client_txn_id":"boundary-import-%s"}`, name), container, "boundary.zip", reference_data.MediaTypeZip)
			job := requireSuccessEnvelope(t, response, http.StatusAccepted)["data"].(map[string]any)
			if terminal := requireJob(t, harness, admin, job["job_id"].(string)); terminal["status"] != "succeeded" {
				t.Fatalf("boundary import: %#v", terminal)
			}
			requireSuccessEnvelope(t, postAction(t, harness, admin, "/api/v1/reference-packs/"+boundary.Key+"/2/activate", "boundary-activate-"+name, ""), http.StatusOK)
			current := consumer.ResolveCurrentPackSet(ctx)
			assertConsumerContract(t, "ResolveCurrentPackSet", current)
			if current.Error != nil {
				t.Fatal(current.Error)
			}
			for _, line := range strings.Split(strings.TrimSuffix(string(payload), "\n"), "\n") {
				var row map[string]any
				decoder := json.NewDecoder(strings.NewReader(line))
				decoder.UseNumber()
				if err := decoder.Decode(&row); err != nil {
					t.Fatal(err)
				}
				id := row["entry_id"].(string)
				entry := consumer.GetPackEntry(ctx, reference_data.GetPackEntryRequest{PackSetID: current.Value.ID, PackKey: boundary.Key, EntryID: id})
				assertConsumerContract(t, "GetPackEntry", entry)
				if entry.Error != nil || entry.Value == nil || string(entry.Value.Item) != line {
					t.Fatal("maximum identity was not addressable", entry.Error)
				}
				kind, value := "sid", any(id)
				if boundary.Key == "enrichment.cisa_kev" {
					kind = "cve_id"
				}
				if boundary.Key == "enrichment.windows_event_ids" {
					kind = "provider_event_id"
					value = map[string]any{"provider_name": row["provider_name"], "event_id": row["event_id"]}
					if row["event_version"] != nil {
						value.(map[string]any)["event_version"] = row["event_version"]
					}
				}
				page := consumer.LookupPackEntries(ctx, reference_data.LookupPackEntriesRequest{PackSetID: current.Value.ID, PackKey: boundary.Key, LookupKind: kind, LookupValue: value})
				assertConsumerContract(t, "LookupPackEntries", page)
				want := 1
				if boundary.Key == "enrichment.windows_event_ids" && row["event_version"] == nil {
					want = 2
				}
				if page.Error != nil || page.Value == nil || len(page.Value.Items) != want {
					t.Fatal("maximum lookup identity failed", page.Error)
				}
			}
		})
	}
}

func assertConsumerContract(t *testing.T, operation string, result any) {
	t.Helper()
	raw, err := json.Marshal(result)
	if err != nil {
		t.Fatal(err)
	}
	if err := packformat.ValidateConsumerResult(operation, raw); err != nil {
		t.Fatal(operation, err)
	}
}
