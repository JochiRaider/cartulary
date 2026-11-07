package testsupport

import (
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	windowsinputs "github.com/JochiRaider/cartulary/contracts/reference-pack-producers/windows-event-ids"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data/producer"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
)

type ProducedWindowsPack struct {
	PortablePackFixture
	ManifestSHA256, PayloadSHA256 string
	Members                       map[string][]byte
}

// WindowsSetDigest independently assembles the expected four-member set from
// authored Base bindings and the actual produced Windows logical identity.
func WindowsSetDigest(t testing.TB, repositoryRoot string, f ProducedWindowsPack) string {
	t.Helper()
	return producedSetDigest(t, repositoryRoot, f.PortablePackFixture, f.ManifestSHA256, f.PayloadSHA256, "cartulary.reference_pack.enrichment.windows_event_ids.v2")
}

func producedSetDigest(t testing.TB, repositoryRoot string, f PortablePackFixture, manifestSHA, payloadSHA, profile string) string {
	t.Helper()
	b, e := os.ReadFile(filepath.Join(repositoryRoot, "contracts/reference-packs/builtins/release.v2.json"))
	if e != nil {
		t.Fatal(e)
	}
	var release struct {
		Packs []struct {
			Binding map[string]any `json:"binding"`
		} `json:"packs"`
	}
	if e = json.Unmarshal(b, &release); e != nil {
		t.Fatal(e)
	}
	members := []map[string]any{}
	for _, p := range release.Packs {
		m := map[string]any{"pack_contract_version": "cartulary.reference_pack_contract.v2"}
		for _, k := range []string{"pack_key", "pack_version", "manifest_sha256", "payload_sha256", "content_profile_id", "content_profile_version"} {
			m[k] = p.Binding[k]
		}
		members = append(members, m)
	}
	if len(members) != 3 {
		t.Fatal("expected three Base registries")
	}
	members = append(members, map[string]any{"pack_key": f.Key, "pack_version": f.Version, "manifest_sha256": manifestSHA, "payload_sha256": payloadSHA, "pack_contract_version": "cartulary.reference_pack_contract.v2", "content_profile_id": profile, "content_profile_version": "2"})
	slices.SortFunc(members, func(a, b map[string]any) int { return strings.Compare(a["pack_key"].(string), b["pack_key"].(string)) })
	b, e = json.Marshal(map[string]any{"schema_id": "cartulary.reference_pack_set.v2", "members": members})
	if e != nil {
		t.Fatal(e)
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

type windowsSourceFixture map[string][]byte

func (f windowsSourceFixture) FetchWindows(_ context.Context, url string) (producer.WindowsDownload, error) {
	b, ok := f[url]
	if !ok {
		return producer.WindowsDownload{}, fmt.Errorf("unlisted synthetic source")
	}
	if strings.HasPrefix(url, "urn:") {
		return producer.WindowsDownload{Bytes: b, URL: url, ContentType: "application/octet-stream", ReceivedAt: "2026-01-01T00:00:00Z", Method: "supplied"}, nil
	}
	return producer.WindowsDownload{Bytes: b, URL: url, ContentType: "text/plain", RetrievedAt: "2026-01-01T00:00:00Z"}, nil
}

// WindowsFixture exercises the real offline producer with explicitly synthetic
// sources. Its licensing decision covers only project-authored fixture text.
// The current test executable is the truthful builder; public artifacts are
// retained through the existing suite facility, while private keys stay in RAM.
func WindowsFixture(t testing.TB, version string, now time.Time, keys producer.SigningContext, metadataVersion int64) ProducedWindowsPack {
	t.Helper()
	encode := func(v any) []byte {
		t.Helper()
		b, e := canonicaljson.Marshal(v)
		if e != nil {
			t.Fatal(e)
		}
		return b
	}
	artifacts, bound, e := suiteservices.ResolveSuiteArtifactDir(nil)
	if e != nil || !bound {
		t.Fatal("Windows fixture requires Harness artifacts", e)
	}
	work := t.TempDir()
	root, e := os.OpenRoot(work)
	if e != nil {
		t.Fatal(e)
	}
	defer root.Close()
	raw, e := windowsinputs.Files.ReadFile("fixture-sources.v2.json")
	if e != nil {
		t.Fatal(e)
	}
	var fixture struct {
		Documents   []struct{ ID, Kind, Body string }
		ExpectedIDs []string `json:"expected_ids"`
	}
	if e = json.Unmarshal(raw, &fixture); e != nil {
		t.Fatal(e)
	}
	inventory := producer.WindowsInventory{SchemaID: "cartulary.windows_event_inventory.v2", Purpose: "synthetic", Sources: []producer.WindowsSource{}, Discovery: []producer.WindowsDiscovery{}}
	fetch := windowsSourceFixture{}
	for _, doc := range fixture.Documents {
		url := "https://learn.microsoft.com/en-us/cartulary-synthetic/" + doc.ID
		inventory.Sources = append(inventory.Sources, producer.WindowsSource{ID: doc.ID, Kind: doc.Kind, URL: url})
		fetch[url] = []byte(doc.Body)
	}
	localRaw, e := windowsinputs.Files.ReadFile("fixture-local.v2.json")
	if e != nil {
		t.Fatal(e)
	}
	var local struct {
		Sources []struct {
			ID    string `json:"id"`
			Kind  string `json:"kind"`
			Bytes string `json:"bytes_base64"`
		}
	}
	if e = json.Unmarshal(localRaw, &local); e != nil {
		t.Fatal(e)
	}
	for _, s := range local.Sources {
		b, e := base64.StdEncoding.DecodeString(s.Bytes)
		if e != nil {
			t.Fatal(e)
		}
		digest := producer.Digest(b)
		url := "urn:cartulary:supplied:" + digest
		inventory.Sources = append(inventory.Sources, producer.WindowsSource{ID: s.ID, Kind: s.Kind, URL: url, ExpectedSHA256: &digest})
		fetch[url] = b
	}
	acquired, e := producer.WindowsAcquire(context.Background(), root, encode(inventory), fetch)
	if e != nil {
		t.Fatal(e)
	}
	builder, e := producer.BuilderIdentity()
	if e != nil {
		t.Fatal(e)
	}
	editorial, e := windowsinputs.Files.ReadFile("editorial.v2.json")
	if e != nil {
		t.Fatal(e)
	}
	decision := producer.WindowsLicenseDecision{SchemaID: "cartulary.windows_event_license_decision.v2", Purpose: "synthetic", SourceLockSHA256: *acquired.LockSHA256, DecisionRef: "project-authored-synthetic-fixture", ApprovedBy: "Cartulary fixture author", License: packformat.License{Expression: "LicenseRef-Cartulary-Synthetic", ListVersion: "3.28.0", Redistribution: "allowed", Notices: []string{"notices/synthetic.txt"}, Bindings: []packformat.LicenseBinding{{Ref: "LicenseRef-Cartulary-Synthetic", Path: "notices/synthetic.txt"}}}, Notices: []producer.WindowsNotice{{Path: "notices/synthetic.txt", SourceID: "notice"}}}
	for name, b := range map[string][]byte{"inventory.json": encode(inventory), "editorial.json": editorial, "license.json": encode(decision)} {
		if e = producer.Retain(root, name, b, 0600); e != nil {
			t.Fatal(e)
		}
	}
	output := filepath.Join(artifacts, "windows-producer", version)
	input := producer.WindowsBuildInput{SchemaID: "cartulary.windows_event_build_input.v2", SourceRoot: work, SourceInventory: filepath.Join(work, "inventory.json"), Editorial: filepath.Join(work, "editorial.json"), LicenseDecision: filepath.Join(work, "license.json"), OutputRoot: output, LedgerRoot: filepath.Join(artifacts, "windows-producer", "ledger"), BuiltAt: "2026-01-02T00:00:00Z", PackVersion: version, BuilderSHA256: builder.SHA256, Repository: producer.WindowsFixtureRepository}
	report, e := producer.AssembleWindows(context.Background(), encode(input), builder)
	if e != nil {
		t.Fatal(e)
	}
	logical := filepath.Join(output, report.LogicalPath)
	manifestBytes, e := os.ReadFile(filepath.Join(logical, "manifest.json"))
	if e != nil {
		t.Fatal(e)
	}
	manifest, e := packformat.DecodeManifest(manifestBytes, true)
	if e != nil {
		t.Fatal(e)
	}
	members := map[string][]byte{"manifest.json": manifestBytes}
	for _, f := range manifest.Files {
		b, e := os.ReadFile(filepath.Join(logical, f.Path))
		if e != nil {
			t.Fatal(e)
		}
		members[f.Path] = b
	}
	actualIDs := []string{}
	for _, line := range strings.Split(strings.TrimSuffix(string(members["payload/entries.ndjson"]), "\n"), "\n") {
		var row struct {
			ID string `json:"entry_id"`
		}
		if e = json.Unmarshal([]byte(line), &row); e != nil {
			t.Fatal(e)
		}
		actualIDs = append(actualIDs, row.ID)
	}
	if !slices.Equal(actualIDs, fixture.ExpectedIDs) {
		t.Fatal("independently authored Windows identities differ", actualIDs)
	}
	container, bootstrap, e := producer.FixtureEnvelope(context.Background(), members, keys, now.Add(365*24*time.Hour).UTC().Format(time.RFC3339), now.Add(7*24*time.Hour).UTC().Format(time.RFC3339), metadataVersion)
	if e != nil {
		t.Fatal(e)
	}
	retained := filepath.Join(artifacts, "windows-envelopes", producer.Digest(container))
	if e = os.MkdirAll(retained, 0700); e != nil {
		t.Fatal(e)
	}
	out, e := os.OpenRoot(retained)
	if e != nil {
		t.Fatal(e)
	}
	defer out.Close()
	for name, b := range map[string][]byte{"container.pack": container, "bootstrap.json": bootstrap} {
		if e = producer.Retain(out, name, b, 0600); e != nil {
			t.Fatal(e)
		}
	}
	t.Logf("synthetic Windows %s manifest=%s payload=%s container=%s bootstrap=%s retained=%s", version, report.ManifestSHA256, report.PayloadSHA256, producer.Digest(container), producer.Digest(bootstrap), retained)
	return ProducedWindowsPack{PortablePackFixture: PortablePackFixture{Bootstrap: bootstrap, Container: container, Key: producer.WindowsKey, Version: version, Path: "ext/reference_packs/containers/" + report.ManifestSHA256 + ".pack"}, ManifestSHA256: report.ManifestSHA256, PayloadSHA256: report.PayloadSHA256, Members: members}
}
