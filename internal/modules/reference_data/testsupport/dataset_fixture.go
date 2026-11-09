package testsupport

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/referencepackfixture"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
	"github.com/JochiRaider/cartulary/referencepacks/format/canonicaljson"
	"github.com/JochiRaider/cartulary/referencepacks/format/dataset"
	"github.com/JochiRaider/cartulary/referencepacks/format/envelope"
	"github.com/JochiRaider/cartulary/referencepacks/format/publication"
)

type datasetFixture struct {
	PortablePackFixture
	ManifestSHA256, PayloadSHA256 string
	Members                       map[string][]byte
}

func signSyntheticDataset(t testing.TB, name, version string, now time.Time, signer *referencepackfixture.SigningContext, metadataVersion int64, members map[string][]byte, objects, relationships int64) datasetFixture {
	t.Helper()
	repository := signer.Repository()
	m := envelope.Manifest{SchemaID: "cartulary.reference_pack_manifest.v2", Contract: envelope.ContractID, Key: signer.Scope().PackKey(), Kind: "enrichment", Version: version, Sequence: 1, ProfileID: dataset.ProfileID, ProfileVersion: "1", SourceProfileID: "cartulary.fixture." + name + ".dataset.v1", SourceProfileSHA256: envelope.Digest(members[dataset.DescriptorPath]), Repository: &repository, SourceIdentifier: "Project-authored synthetic lifecycle fixture", SourceVersion: "1", BuiltAt: "2026-11-08T00:00:00Z", Artifacts: []envelope.SourceArtifact{{Ref: "urn:cartulary:synthetic:" + name, Version: "1", SHA256: envelope.Digest(members[dataset.ObjectsPath])}}, Dependencies: []envelope.Dependency{}, Conflicts: []envelope.Conflict{}, Files: publication.LogicalMemberInventory(members), License: envelope.License{Expression: "CC0-1.0", ListVersion: "3.28.0", Redistribution: "allowed", Notices: []string{"notices/fixture.txt"}, Bindings: []envelope.LicenseBinding{}}, Extensions: map[string]any{}}
	if version == "fixture-b" {
		m.Sequence = 2
	}
	m.Builder.ID = "cartulary.fixture.artifact"
	sourceAsOf := "2026-11-08T00:00:00Z"
	m.SourceAsOf = &sourceAsOf
	m.Builder.Version = "1"
	m.Builder.SHA256 = envelope.Digest(members[dataset.ObjectsPath])
	m.Compatibility.Majors = []int{1}
	m.Compatibility.Capabilities = []string{}
	m.Summary.Kind = "objects_relationships"
	m.Summary.Objects, m.Summary.Relationships = &objects, &relationships
	return signDataset(t, name, now, signer, metadataVersion, m, members)
}

func signDataset(t testing.TB, name string, now time.Time, signer *referencepackfixture.SigningContext, metadataVersion int64, m envelope.Manifest, members map[string][]byte) datasetFixture {
	t.Helper()
	var err error
	members["manifest.json"], err = canonicaljson.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	manifestSHA := envelope.Digest(members["manifest.json"])
	payloadSHA, err := envelope.PayloadDigest(m.Files)
	if err != nil {
		t.Fatal(err)
	}
	container, bootstrap, err := referencepackfixture.Envelope(context.Background(), members, signer, now.Add(365*24*time.Hour).UTC().Format(time.RFC3339), now.Add(7*24*time.Hour).UTC().Format(time.RFC3339), metadataVersion)
	if err != nil {
		t.Fatal(err)
	}
	artifacts, bound, err := suiteservices.ResolveSuiteArtifactDir(nil)
	if err != nil || !bound {
		t.Fatal("fixture envelope requires Harness artifacts", err)
	}
	retained := filepath.Join(artifacts, name+"-envelopes", envelope.Digest(container))
	if err := publication.PublishTree(retained, "public", map[string][]byte{"container.pack": container, "bootstrap.json": bootstrap}); err != nil {
		t.Fatal(err)
	}
	t.Logf("%s fixture %s manifest=%s payload=%s container=%s retained=%s", name, m.Version, manifestSHA, payloadSHA, envelope.Digest(container), retained)
	return datasetFixture{PortablePackFixture{Bootstrap: bootstrap, Container: container, Key: m.Key, Version: m.Version, Path: "ext/reference_packs/containers/" + manifestSHA + ".pack"}, manifestSHA, payloadSHA, members}
}

func readDatasetMembers(t testing.TB, root string) (envelope.Manifest, map[string][]byte) {
	t.Helper()
	b, err := os.ReadFile(filepath.Join(root, "manifest.json"))
	if err != nil {
		t.Fatal(err)
	}
	m, err := envelope.DecodeManifest(b, true)
	if err != nil {
		t.Fatal(err)
	}
	members := map[string][]byte{"manifest.json": b}
	for _, f := range m.Files {
		b, err := os.ReadFile(filepath.Join(root, f.Path))
		if err != nil || int64(len(b)) != f.Size || envelope.Digest(b) != f.SHA256 {
			t.Fatal("canonical dataset member integrity", f.Path, err)
		}
		members[f.Path] = b
	}
	return m, members
}

func corpusDatasetMembers(t testing.TB, repositoryRoot, name string) (envelope.Manifest, map[string][]byte) {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(repositoryRoot, "fixtures/reference-packs/dataset-v1/catalog.json"))
	if err != nil {
		t.Fatal(err)
	}
	var catalog map[string]struct {
		Root          string `json:"logical_root"`
		Manifest      string `json:"manifest_sha256"`
		Payload       string `json:"payload_sha256"`
		Objects       int64  `json:"objects"`
		Relationships int64  `json:"relationships"`
	}
	if err := json.Unmarshal(raw, &catalog); err != nil {
		t.Fatal(err)
	}
	entry, ok := catalog[name]
	if !ok {
		t.Fatal("unregistered canonical dataset fixture", name)
	}
	m, members := readDatasetMembers(t, filepath.Join(repositoryRoot, entry.Root))
	payload, err := envelope.PayloadDigest(m.Files)
	if err != nil || envelope.Digest(members["manifest.json"]) != entry.Manifest || payload != entry.Payload || m.Summary.Objects == nil || m.Summary.Relationships == nil || *m.Summary.Objects != entry.Objects || *m.Summary.Relationships != entry.Relationships {
		t.Fatal("canonical dataset catalog mismatch", name, err)
	}
	return m, members
}

func corpusDatasetFixture(t testing.TB, name, version string, now time.Time, keys *referencepackfixture.SigningContext, metadataVersion int64) datasetFixture {
	t.Helper()
	m, members := corpusDatasetMembers(t, "../../..", name)
	repository := keys.Repository()
	m.Repository, m.Version, m.Sequence = &repository, version, 1
	m.SourceIdentifier = "Synthetic lifecycle derivative of the retained factual " + name + " catalog"
	if version == "fixture-restricted" {
		m.License.Redistribution = "restricted"
	}
	if version == "fixture-b" || strings.HasPrefix(version, "negative-") {
		m.Sequence = 2
		lines := strings.Split(strings.TrimSuffix(string(members[dataset.ObjectsPath]), "\n"), "\n")
		v, err := canonicaljson.DecodeStrict([]byte(lines[0]))
		if err != nil {
			t.Fatal(err)
		}
		row := v.(map[string]any)
		switch version {
		case "fixture-b":
			row["description"] = "Synthetic lifecycle B description; not an upstream assertion."
		case "negative-v1-entry":
			m.Sequence = 3
			row["schema_id"] = "cartulary.reference_pack.predecessor.v1"
		case "negative-identity":
			m.Sequence = 4
			next, err := canonicaljson.DecodeStrict([]byte(lines[1]))
			if err != nil {
				t.Fatal(err)
			}
			row["object_id"] = next.(map[string]any)["object_id"]
		case "negative-provenance":
			m.Sequence = 5
			row["source_refs"] = []any{map[string]any{"artifact_index": len(m.Artifacts), "locator": "synthetic"}}
		default:
			t.Fatal("unregistered generic negative fixture", version)
		}
		encoded, err := canonicaljson.Marshal(row)
		if err != nil {
			t.Fatal(err)
		}
		lines[0] = string(encoded)
		members[dataset.ObjectsPath] = []byte(strings.Join(lines, "\n") + "\n")
	}
	for i := range m.Files {
		b := members[m.Files[i].Path]
		m.Files[i].Size, m.Files[i].SHA256 = int64(len(b)), envelope.Digest(b)
	}
	return signDataset(t, name, now, keys, metadataVersion, m, members)
}
