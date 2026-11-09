package reference_data_test

import (
	"bytes"
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
)

// The release is imported unchanged, with its actual publisher and expiry.
// This test never builds a replacement corpus or signs a fresh test envelope.
func TestWindowsRealReleaseImport_Integration(t *testing.T) {
	release := referencetest.WindowsRelease(t, "../../..")
	container, bootstrap, members := release.Container, release.Bootstrap, release.Members
	var manifest struct {
		Key        string `json:"pack_key"`
		Version    string `json:"pack_version"`
		Repository string `json:"trust_repository_id"`
	}
	if err := json.Unmarshal(members["manifest.json"], &manifest); err != nil {
		t.Fatal(err)
	}
	if manifest.Key != "enrichment.windows_event_ids" || manifest.Repository != "cartulary.windows_event_ids" {
		t.Fatal("real publisher required")
	}
	rows := bytes.Split(bytes.TrimSuffix(members["payload/objects.ndjson"], []byte{'\n'}), []byte{'\n'})
	if len(rows) != 503 {
		t.Fatal("retained corpus count", len(rows))
	}
	bootstrapPath := filepath.Join(t.TempDir(), "bootstrap.json")
	if err := os.WriteFile(bootstrapPath, bootstrap, 0600); err != nil {
		t.Fatal(err)
	}
	runtime := appsupport.StartRuntime(t)
	harness := runtime.StartServer(t, appsupport.ServerOptions{Prefix: "windows-release", TestRouteMode: httptestx.TestRouteModeDisabled, Env: map[string]string{"CARTULARY__REFERENCE_PACKS__TRUST_BOOTSTRAP_PATH": bootstrapPath, "CARTULARY__REFERENCE_PACKS__CLOCK_TRUSTED": "true"}})
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
	ctx := context.Background()
	before := consumer.ResolveCurrentPackSet(ctx)
	if before.Error != nil || len(before.Value.Members) != 3 {
		t.Fatal("Base-only setup", before.Error)
	}
	response := postReferencePackUpload(t, harness.Server.HTTP.URL, admin, `{"client_txn_id":"windows-real-import"}`, container, "windows.pack", "application/octet-stream")
	job := httptestx.RequireSuccessEnvelope(t, response, http.StatusAccepted)["data"].(map[string]any)
	if terminal := requireJob(t, harness, admin, job["job_id"].(string)); terminal["status"] != "succeeded" {
		t.Fatalf("retained release import: %#v", terminal)
	}
	if staged := consumer.ResolveCurrentPackSet(ctx); staged.Error != nil || staged.Value.ID != before.Value.ID {
		t.Fatal("import activated release")
	}
	prefix := "/api/v1/reference-packs/" + manifest.Key + "/" + manifest.Version
	httptestx.RequireSuccessEnvelope(t, postAction(t, harness, admin, prefix+"/activate", "windows-real-activate", ""), http.StatusOK)
	active := consumer.ResolveCurrentPackSet(ctx)
	if active.Error != nil || len(active.Value.Members) != 4 {
		t.Fatal("independent Windows activation", active.Error)
	}
	for _, row := range rows {
		var item struct {
			ID string `json:"object_id"`
		}
		if err := json.Unmarshal(row, &item); err != nil {
			t.Fatal(err)
		}
		got := consumer.GetPackEntry(ctx, reference_data.GetPackEntryRequest{PackSetID: active.Value.ID, PackKey: manifest.Key, EntryID: item.ID})
		if got.Error != nil || !bytes.Equal(got.Value.Item, row) || got.Value.Provenance.ManifestSHA256 != release.ManifestSHA256 {
			t.Fatal("complete retained definition", item.ID, got.Error)
		}
	}
	lookup := func(kind string, value any, expected []string) {
		t.Helper()
		page := consumer.LookupPackEntries(ctx, reference_data.LookupPackEntriesRequest{PackSetID: active.Value.ID, PackKey: manifest.Key, LookupKind: kind, LookupValue: value})
		if page.Error != nil {
			t.Fatal(kind, page.Error)
		}
		ids := []string{}
		for _, raw := range page.Value.Items {
			var item struct {
				ID string `json:"object_id"`
			}
			if err := json.Unmarshal(raw, &item); err != nil {
				t.Fatal(err)
			}
			ids = append(ids, item.ID)
		}
		if !slices.Equal(ids, expected) || page.Value.NextCursor != nil {
			t.Fatal(kind, ids, expected)
		}
	}
	lookup("provider_event_id", map[string]any{"provider_name": "MICROSOFT-WINDOWS-SECURITY-AUDITING", "event_id": json.Number("4624")}, []string{"microsoft-windows-security-auditing:4624:0", "microsoft-windows-security-auditing:4624:1", "microsoft-windows-security-auditing:4624:2"})
	lookup("provider_event_version", map[string]any{"provider_name": "Microsoft-Windows-Security-Auditing", "event_id": json.Number("4624"), "event_version": json.Number("2")}, []string{"microsoft-windows-security-auditing:4624:2"})
	lookup("event_id", json.Number("4104"), []string{"microsoft-windows-powershell:4104:*", "powershellcore:4104:*", "powershellcore:4104:1"})
	lookup("event_id", json.Number("65535"), []string{})
	httptestx.RequireSuccessEnvelope(t, postAction(t, harness, admin, prefix+"/disable", "windows-real-disable", "retained release qualification"), http.StatusOK)
	current := consumer.ResolveCurrentPackSet(ctx)
	if current.Error != nil || current.Value.ID != before.Value.ID {
		t.Fatal("disable did not restore Base-only set", current.Error)
	}
	if got := consumer.GetPackEntry(ctx, reference_data.GetPackEntryRequest{PackSetID: active.Value.ID, PackKey: manifest.Key, EntryID: "microsoft-windows-security-auditing:4624:2"}); got.Error != nil {
		t.Fatal("disable lost historical content", got.Error)
	}
}
