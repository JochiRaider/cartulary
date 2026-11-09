package testsupport

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
)

// A retained qualification never constructs, renews or signs its subject. It
// checks exact public identities before sending these bytes to the importer,
// whose clock and complete verification lifecycle supply freshness admission.
func retainedDataset(t testing.TB, repositoryRoot, name, key, repository, class string, objects, relationships int64) (datasetFixture, packformat.Manifest) {
	t.Helper()
	read := func(name string) []byte {
		t.Helper()
		b, err := os.ReadFile(filepath.Join(repositoryRoot, name))
		if err != nil {
			t.Fatal(err)
		}
		return b
	}
	var release struct {
		Schema        string `json:"schema_id"`
		Key           string `json:"pack_key"`
		Version       string `json:"pack_version"`
		Class         string `json:"authority_class"`
		LogicalRoot   string `json:"logical_root"`
		EnvelopeRoot  string `json:"envelope_root"`
		Manifest      string `json:"manifest_sha256"`
		Payload       string `json:"payload_sha256"`
		Container     string `json:"container_sha256"`
		Bootstrap     string `json:"bootstrap_sha256"`
		Report        string `json:"envelope_report_sha256"`
		Objects       int64  `json:"objects"`
		Relationships int64  `json:"relationships"`
	}
	selected := read("fixtures/reference-packs/dataset-v1/" + name + "/release.v3.json")
	if err := json.Unmarshal(selected, &release); err != nil {
		t.Fatal(err)
	}
	if release.Schema != "cartulary.reference_pack_release_descriptor.v3" || release.Key != key || release.Class != class || release.Objects != objects || release.Relationships != relationships {
		t.Fatal("retained release descriptor identity")
	}
	if !bytes.Equal(selected, read("fixtures/reference-packs/dataset-v1/"+name+"/descriptors/"+release.Container+".json")) {
		t.Fatal("selected release differs from its immutable descriptor")
	}
	m, members := readDatasetMembers(t, filepath.Join(repositoryRoot, release.LogicalRoot))
	payload, err := packformat.PayloadDigest(m.Files)
	if err != nil || packformat.Digest(members["manifest.json"]) != release.Manifest || payload != release.Payload || m.Key != key || m.Version != release.Version || m.Repository == nil || *m.Repository != repository || m.Summary.Objects == nil || *m.Summary.Objects != objects || m.Summary.Relationships == nil || *m.Summary.Relationships != relationships {
		t.Fatal("retained logical identity", err)
	}
	container, bootstrap := read(release.EnvelopeRoot+"/container.pack"), read(release.EnvelopeRoot+"/bootstrap.json")
	if packformat.Digest(container) != release.Container || packformat.Digest(bootstrap) != release.Bootstrap || packformat.Digest(read(release.EnvelopeRoot+"/envelope-report.json")) != release.Report {
		t.Fatal("retained signed envelope identity")
	}
	return datasetFixture{PortablePackFixture{Bootstrap: bootstrap, Container: container, Key: key, Version: m.Version, Path: "ext/reference_packs/containers/" + release.Manifest + ".pack"}, release.Manifest, payload, members}, m
}

func WindowsRelease(t testing.TB, root string) ProducedWindowsPack {
	t.Helper()
	f, _ := retainedDataset(t, root, "windows", "enrichment.windows_event_ids", "cartulary.windows_event_ids", "release", 503, 0)
	return ProducedWindowsPack{f.PortablePackFixture, f.ManifestSHA256, f.PayloadSHA256, f.Members}
}

func ATTACKRelease(t testing.TB, root string) AuthenticATTACKPack {
	t.Helper()
	f, m := retainedDataset(t, root, "attack/release", "framework.attack", "cartulary.attack", "release", 6221, 34632)
	if m.License.Expression != "LicenseRef-MITRE-ATTACK" || m.License.Redistribution != "allowed" ||
		m.SourceProfileID != "cartulary.attack.all_domains.dataset.v2" ||
		packformat.Digest(f.Members["notices/MITRE-ATTACK.txt"]) != "738144f7fb054722a4ef9d3367c51710341dc12fc574c6ac3a41daaaecd8bf5e" {
		t.Fatal("production release must retain the exact MITRE license and approved distribution binding")
	}
	set := producedSetDigest(t, root, f.PortablePackFixture, f.ManifestSHA256, f.PayloadSHA256, "cartulary.reference_pack.dataset.v1")
	return AuthenticATTACKPack{f.PortablePackFixture, f.ManifestSHA256, f.PayloadSHA256, set, 6221, 34632}
}
