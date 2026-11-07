package testsupport

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data/producer"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
)

type AuthenticATTACKPack struct {
	PortablePackFixture
	ManifestSHA256, PayloadSHA256, SetSHA256 string
	Objects, Relationships                   int64
}

// AuthenticATTACKFixture reads immutable canonical bytes. Only its transport
// envelope is renewed, using caller-owned isolated keys and an explicit time.
// Nothing writes to the fixture tree or changes production verification clocks.
func AuthenticATTACKFixture(t testing.TB, repositoryRoot, version string, now time.Time, keys producer.SigningContext, metadataVersion int64) AuthenticATTACKPack {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(repositoryRoot, "fixtures/reference-packs/attack-native-v2/catalog.json"))
	if err != nil {
		t.Fatal(err)
	}
	var catalog map[string]struct {
		Set           string `json:"pack_set_sha256"`
		Root          string `json:"logical_root"`
		Manifest      string `json:"manifest_sha256"`
		Payload       string `json:"payload_sha256"`
		Objects       int64  `json:"object_count"`
		Relationships int64  `json:"relationship_count"`
	}
	if err := json.Unmarshal(b, &catalog); err != nil {
		t.Fatal(err)
	}
	entry, ok := catalog[version]
	if !ok {
		t.Fatal("missing authentic fixture", version)
	}
	root := filepath.Join(repositoryRoot, entry.Root)
	b, err = os.ReadFile(filepath.Join(root, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	m, err := packformat.DecodeManifest(b, true)
	if err != nil || packformat.Digest(b) != entry.Manifest {
		t.Fatal("authentic manifest integrity", err)
	}
	payload, err := packformat.PayloadDigest(m.Files)
	if err != nil || payload != entry.Payload || *m.Summary.Objects != entry.Objects || *m.Summary.Relationships != entry.Relationships {
		t.Fatal("authentic fixture witness", err)
	}
	members := map[string][]byte{"manifest.json": b}
	for _, f := range m.Files {
		b, err := os.ReadFile(filepath.Join(root, f.Path))
		if err != nil {
			t.Fatal(err)
		}
		members[f.Path] = b
	}
	container, bootstrap, err := producer.FixtureEnvelope(context.Background(), members, keys, now.Add(365*24*time.Hour).UTC().Format(time.RFC3339), now.Add(7*24*time.Hour).UTC().Format(time.RFC3339), metadataVersion)
	if err != nil {
		t.Fatal(err)
	}
	// Retain only public transport bytes in the existing suite artifact tree.
	// Fixture keys stay in caller memory; canonical fixtures remain immutable.
	artifacts, bound, err := suiteservices.ResolveSuiteArtifactDir(nil)
	if err != nil || !bound {
		t.Fatal("authentic envelope requires a harness artifact binding", err)
	}
	retained := filepath.Join(artifacts, "reference-pack-envelopes", packformat.Digest(container))
	if err := os.MkdirAll(retained, 0700); err != nil {
		t.Fatal(err)
	}
	output, err := os.OpenRoot(retained)
	if err != nil {
		t.Fatal(err)
	}
	defer output.Close()
	for name, bytes := range map[string][]byte{"container.pack": container, "bootstrap.json": bootstrap} {
		if err := producer.Retain(output, name, bytes, 0600); err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("authentic ATT&CK %s public envelope retained: %s", version, retained)
	return AuthenticATTACKPack{PortablePackFixture: PortablePackFixture{Bootstrap: bootstrap, Container: container, Key: m.Key, Version: m.Version, Path: "ext/reference_packs/containers/" + entry.Manifest + ".pack"}, SetSHA256: entry.Set, ManifestSHA256: entry.Manifest, PayloadSHA256: entry.Payload, Objects: entry.Objects, Relationships: entry.Relationships}
}
