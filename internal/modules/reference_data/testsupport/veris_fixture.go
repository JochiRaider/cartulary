package testsupport

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data/producer"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	"github.com/JochiRaider/cartulary/internal/testutil/suiteservices"
)

type ProducedVERISPack struct {
	PortablePackFixture
	ManifestSHA256, PayloadSHA256 string
	Members                       map[string][]byte
}

func VERISSetDigest(t testing.TB, repositoryRoot string, f ProducedVERISPack) string {
	t.Helper()
	return producedSetDigest(t, repositoryRoot, f.PortablePackFixture, f.ManifestSHA256, f.PayloadSHA256, packformat.VERISProfile)
}

// VERISRelease reads only canonical public artifacts and their JSON descriptor. Upstream source documents and research retention are never test inputs.
func VERISRelease(t testing.TB, repositoryRoot string) ProducedVERISPack {
	t.Helper()
	read := func(path string) []byte {
		t.Helper()
		b, err := os.ReadFile(filepath.Join(repositoryRoot, path))
		if err != nil {
			t.Fatal(err)
		}
		return b
	}
	var release struct {
		LogicalRoot     string `json:"logical_root"`
		EnvelopeRoot    string `json:"envelope_root"`
		ManifestSHA256  string `json:"manifest_sha256"`
		PayloadSHA256   string `json:"payload_sha256"`
		ContainerSHA256 string `json:"container_sha256"`
		BootstrapSHA256 string `json:"bootstrap_sha256"`
	}
	if err := json.Unmarshal(read("fixtures/veris/release.v1.json"), &release); err != nil {
		t.Fatal(err)
	}
	members := map[string][]byte{"manifest.json": read(release.LogicalRoot + "/manifest.json")}
	m, err := packformat.DecodeManifest(members["manifest.json"], true)
	if err != nil {
		t.Fatal(err)
	}
	for _, f := range m.Files {
		members[f.Path] = read(release.LogicalRoot + "/" + f.Path)
		if producer.Digest(members[f.Path]) != f.SHA256 || int64(len(members[f.Path])) != f.Size {
			t.Fatal("canonical member drift", f.Path)
		}
	}
	payload, err := packformat.PayloadDigest(m.Files)
	if err != nil || payload != release.PayloadSHA256 || producer.Digest(members["manifest.json"]) != release.ManifestSHA256 {
		t.Fatal("logical release identity", err)
	}
	container, bootstrap := read(release.EnvelopeRoot+"/container.pack"), read(release.EnvelopeRoot+"/bootstrap.json")
	if producer.Digest(container) != release.ContainerSHA256 || producer.Digest(bootstrap) != release.BootstrapSHA256 {
		t.Fatal("public envelope drift")
	}
	if err := packformat.ValidateContent(context.Background(), m, verisMemberSource(members), nil); err != nil {
		t.Fatal(err)
	}
	return ProducedVERISPack{PortablePackFixture: PortablePackFixture{Bootstrap: bootstrap, Container: container, Key: m.Key, Version: m.Version, Path: "ext/reference_packs/containers/" + release.ManifestSHA256 + ".pack"}, ManifestSHA256: release.ManifestSHA256, PayloadSHA256: payload, Members: members}
}

// VERISFixture freshly signs canonical corpus derivatives with independent test
// authority. Version B changes original text explicitly, never historical facts.
func VERISFixture(t testing.TB, version string, now time.Time, keys producer.SigningContext, metadataVersion int64) ProducedVERISPack {
	t.Helper()
	f := VERISRelease(t, "../../..")
	m, err := packformat.DecodeManifest(f.Members["manifest.json"], true)
	if err != nil {
		t.Fatal(err)
	}
	m.Repository = &keys.Repository
	m.Version = version
	m.Sequence = 1
	m.SourceIdentifier = "Cartulary synthetic VERIS successor fixture; not a new upstream release"
	if version == "fixture-restricted" {
		m.License.Redistribution = "restricted"
	}
	if version == "fixture-b" {
		m.Sequence = 2
		lines := strings.Split(strings.TrimSuffix(string(f.Members["payload/objects.ndjson"]), "\n"), "\n")
		var row map[string]any
		if err = json.Unmarshal([]byte(lines[3]), &row); err != nil {
			t.Fatal(err)
		}
		row["source"].(map[string]any)["definition"] = "Synthetic lifecycle B description; not an upstream assertion."
		row["description"] = "Synthetic lifecycle B description; not an upstream assertion."
		b, e := canonicaljson.Marshal(row)
		if e != nil {
			t.Fatal(e)
		}
		lines[3] = string(b)
		f.Members["payload/objects.ndjson"] = []byte(strings.Join(lines, "\n") + "\n")
	}
	// Negative derivatives retain a v2 manifest and valid hashes/signatures, so
	// admission reaches the intended content check rather than a trust failure.
	if strings.HasPrefix(version, "negative-") {
		m.Sequence = 3
		lines := strings.Split(strings.TrimSuffix(string(f.Members["payload/objects.ndjson"]), "\n"), "\n")
		var row map[string]any
		if err = json.Unmarshal([]byte(lines[0]), &row); err != nil {
			t.Fatal(err)
		}
		switch version {
		case "negative-v1-entry":
			row["schema_id"] = "cartulary.reference_pack.framework.veris.object.v1"
		case "negative-identity":
			row["source"].(map[string]any)["value"] = "Synthetic wrong value"
		case "negative-provenance":
			row["source_refs"] = []any{map[string]any{"artifact_index": len(m.Artifacts), "locator": "synthetic"}}
		default:
			t.Fatal("unknown VERIS negative")
		}
		b, e := canonicaljson.Marshal(row)
		if e != nil {
			t.Fatal(e)
		}
		lines[0] = string(b)
		f.Members["payload/objects.ndjson"] = []byte(strings.Join(lines, "\n") + "\n")
	}
	for i := range m.Files {
		b := f.Members[m.Files[i].Path]
		m.Files[i].Size = int64(len(b))
		m.Files[i].SHA256 = producer.Digest(b)
	}
	f.Members["manifest.json"], err = canonicaljson.Marshal(m)
	if err != nil {
		t.Fatal(err)
	}
	f.ManifestSHA256 = producer.Digest(f.Members["manifest.json"])
	f.PayloadSHA256, err = packformat.PayloadDigest(m.Files)
	if err != nil {
		t.Fatal(err)
	}
	f.Container, f.Bootstrap, err = producer.FixtureEnvelope(context.Background(), f.Members, keys, now.Add(365*24*time.Hour).Format(time.RFC3339), now.Add(7*24*time.Hour).Format(time.RFC3339), metadataVersion)
	if err != nil {
		t.Fatal(err)
	}
	f.Version = version
	f.Path = "ext/reference_packs/containers/" + f.ManifestSHA256 + ".pack"
	artifacts, bound, err := suiteservices.ResolveSuiteArtifactDir(nil)
	if err != nil || !bound {
		t.Fatal("VERIS fixture requires Harness artifacts", err)
	}
	retained := filepath.Join(artifacts, "veris-envelopes", producer.Digest(f.Container))
	if err = os.MkdirAll(retained, 0700); err != nil {
		t.Fatal(err)
	}
	for name, b := range map[string][]byte{"container.pack": f.Container, "bootstrap.json": f.Bootstrap} {
		if err = os.WriteFile(filepath.Join(retained, name), b, 0600); err != nil {
			t.Fatal(err)
		}
	}
	t.Logf("VERIS fixture %s manifest=%s payload=%s container=%s retained=%s", version, f.ManifestSHA256, f.PayloadSHA256, producer.Digest(f.Container), retained)
	return f
}

type verisMemberSource map[string][]byte

func (s verisMemberSource) Open(_ context.Context, name string) (io.ReadCloser, error) {
	b, ok := s[name]
	if !ok {
		return nil, errors.New("missing VERIS member")
	}
	return io.NopCloser(bytes.NewReader(b)), nil
}
