package testsupport

import (
	"archive/zip"
	"bytes"
	"crypto/ed25519"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"io"
	"slices"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data"
	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

type PortablePackFixture struct {
	Bootstrap, Container, References, Content []byte
	Path, Key, Version                        string
}

// PortableFixture renews only the test transport's TUF envelope using an
// explicitly test-only key and relative expiries. The input contract fixtures
// independently specify payload and identities; this is application-integration
// setup, not an independent expected cryptographic vector.
func PortableFixture(t testing.TB, now time.Time, inputVector, referenceVector []byte) PortablePackFixture {
	return portableFixture(t, now, inputVector, referenceVector, "", 0, 0)
}

// PortableCohortFixture supplies two fresh historical versions whose lexical
// order opposes their release sequence. A conflicting cohort contains two
// independently valid root envelopes with different complete signer sets.
func PortableCohortFixture(t testing.TB, now time.Time, inputVector, referenceVector []byte, conflicting bool) ([]byte, []byte, []byte, map[string][]byte) {
	t.Helper()
	first := portableFixture(t, now, inputVector, referenceVector, "z-older", 3, 3)
	signatures := 3
	if conflicting {
		signatures = 2
	}
	second := portableFixture(t, now, inputVector, referenceVector, "a-newer", 4, signatures)
	if !bytes.Equal(first.Bootstrap, second.Bootstrap) {
		t.Fatal("cohort bootstrap mismatch")
	}
	refs, err := reference_data.DecodeIncidentBundleReferences(first.References)
	if err != nil {
		t.Fatal(err)
	}
	other, err := reference_data.DecodeIncidentBundleReferences(second.References)
	if err != nil {
		t.Fatal(err)
	}
	refs.Sets = append(refs.Sets, other.Sets...)
	for _, v := range other.Versions {
		if v.Key == second.Key {
			refs.Versions = append(refs.Versions, v)
		}
	}
	slices.SortFunc(refs.Sets, func(a, b packformat.Set) int { return strings.Compare(a.ID, b.ID) })
	slices.SortFunc(refs.Versions, func(a, b reference_data.IncidentBundleVersionReference) int {
		if c := strings.Compare(a.Key, b.Key); c != 0 {
			return c
		}
		return strings.Compare(a.Version, b.Version)
	})
	references, err := reference_data.EncodeIncidentBundleReferences(refs.Sets, refs.Versions)
	if err != nil {
		t.Fatal(err)
	}
	content := packformat.EmptyPortableContent()
	for _, fixture := range []PortablePackFixture{first, second} {
		var partial packformat.PortableContent
		if err := json.Unmarshal(fixture.Content, &partial); err != nil {
			t.Fatal(err)
		}
		content.Containers = append(content.Containers, partial.Containers...)
		content.RequiredMembers = append(content.RequiredMembers, partial.RequiredMembers...)
	}
	slices.SortFunc(content.Containers, func(a, b packformat.PortableContainer) int {
		return strings.Compare(a.ManifestSHA256, b.ManifestSHA256)
	})
	slices.SortFunc(content.RequiredMembers, func(a, b packformat.PortableRequiredMember) int {
		if c := strings.Compare(a.SetID, b.SetID); c != 0 {
			return c
		}
		return strings.Compare(a.Key, b.Key)
	})
	encoded, err := packformat.EncodePortableContent(content, refs)
	if err != nil {
		t.Fatal(err)
	}
	return first.Bootstrap, references, encoded, map[string][]byte{first.Path: first.Container, second.Path: second.Container}
}

func portableFixture(t testing.TB, now time.Time, inputVector, referenceVector []byte, version string, sequence int64, rootSignatures int) PortablePackFixture {
	t.Helper()
	canonical := func(v any) []byte {
		t.Helper()
		data, err := canonicaljson.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		return data
	}
	digest := func(data []byte) string { sum := sha256.Sum256(data); return hex.EncodeToString(sum[:]) }
	var vector struct {
		Allowed struct {
			Container string                                        `json:"container_base64"`
			Reference reference_data.IncidentBundleVersionReference `json:"reference"`
		}
	}
	if err := json.Unmarshal(inputVector, &vector); err != nil {
		t.Fatal(err)
	}
	body, err := base64.StdEncoding.DecodeString(vector.Allowed.Container)
	if err != nil {
		t.Fatal(err)
	}
	archive, err := zip.NewReader(bytes.NewReader(body), int64(len(body)))
	if err != nil {
		t.Fatal(err)
	}
	files := map[string][]byte{}
	for _, file := range archive.File {
		reader, err := file.Open()
		if err != nil {
			t.Fatal(err)
		}
		data, readErr := io.ReadAll(reader)
		closeErr := reader.Close()
		if readErr != nil || closeErr != nil {
			t.Fatal(readErr, closeErr)
		}
		files[file.Name] = data
	}
	reference := vector.Allowed.Reference
	if version != "" {
		var manifest map[string]any
		if err := json.Unmarshal(files["manifest.json"], &manifest); err != nil {
			t.Fatal(err)
		}
		manifest["pack_version"] = version
		manifest["pack_release_sequence"] = sequence
		files["manifest.json"] = canonical(manifest)
		reference.Version = version
		reference.ManifestSHA256 = digest(files["manifest.json"])
	}
	keys := map[string]any{}
	privateKeys := map[string]ed25519.PrivateKey{}
	ids := []string{}
	online := ""
	for ordinal := 0; ordinal < 3; ordinal++ {
		seed := sha256.Sum256([]byte("cartulary portable integration test key only " + strconv.Itoa(ordinal)))
		private := ed25519.NewKeyFromSeed(seed[:])
		public := map[string]any{"keytype": "ed25519", "scheme": "ed25519", "keyval": map[string]any{"public": hex.EncodeToString(private.Public().(ed25519.PublicKey))}}
		id := digest(canonical(public))
		ids = append(ids, id)
		keys[id] = public
		privateKeys[id] = private
		if ordinal == 0 {
			online = id
		}
	}
	slices.Sort(ids)
	sign := func(signed map[string]any) []byte {
		signatures := []any{}
		for _, id := range ids {
			if signed["_type"] != "root" && id != online {
				continue
			}
			if signed["_type"] == "root" && signed["version"] == 2 && rootSignatures > 0 && len(signatures) == rootSignatures {
				break
			}
			signatures = append(signatures, map[string]any{"keyid": id, "sig": hex.EncodeToString(ed25519.Sign(privateKeys[id], canonical(signed)))})
		}
		return canonical(map[string]any{"signed": signed, "signatures": signatures})
	}
	roles := map[string]any{"root": map[string]any{"keyids": ids, "threshold": 2}}
	for _, role := range []string{"targets", "snapshot", "timestamp"} {
		roles[role] = map[string]any{"keyids": []string{online}, "threshold": 1}
	}
	root := sign(map[string]any{"_type": "root", "spec_version": "1.0.35", "version": 1, "expires": now.Add(365 * 24 * time.Hour).UTC().Format(time.RFC3339), "consistent_snapshot": false, "keys": keys, "roles": roles, "cartulary": map[string]any{"schema_id": "cartulary.reference_pack_tuf_root_binding.v1", "trust_repository_id": "fixture.repo"}})
	if rootSignatures > 0 {
		var rootEnvelope struct {
			Signed map[string]any `json:"signed"`
		}
		if err := json.Unmarshal(root, &rootEnvelope); err != nil {
			t.Fatal(err)
		}
		rootEnvelope.Signed["version"] = 2
		files["metadata/2.root.json"] = sign(rootEnvelope.Signed)
	}
	for _, role := range []string{"targets", "snapshot", "timestamp"} {
		path := "metadata/" + role + ".json"
		var envelope struct {
			Signed map[string]any `json:"signed"`
		}
		if err := json.Unmarshal(files[path], &envelope); err != nil {
			t.Fatal(err)
		}
		envelope.Signed["expires"] = now.Add(7 * 24 * time.Hour).UTC().Format(time.RFC3339)
		if role == "targets" && version != "" {
			binding := envelope.Signed["cartulary"].(map[string]any)
			binding["manifest_sha256"] = reference.ManifestSHA256
			binding["pack_version"] = version
			binding["pack_release_sequence"] = sequence
			envelope.Signed["targets"].(map[string]any)["manifest.json"] = map[string]any{"length": len(files["manifest.json"]), "hashes": map[string]any{"sha256": reference.ManifestSHA256}}
		}
		if role != "targets" {
			child := "targets.json"
			if role == "timestamp" {
				child = "snapshot.json"
			}
			data := files["metadata/"+child]
			envelope.Signed["meta"] = map[string]any{child: map[string]any{"version": 1, "length": len(data), "hashes": map[string]any{"sha256": digest(data)}}}
		}
		files[path] = sign(envelope.Signed)
	}
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	slices.Sort(names)
	var out bytes.Buffer
	writer := zip.NewWriter(&out)
	for _, name := range names {
		member, err := writer.CreateHeader(&zip.FileHeader{Name: name, Method: zip.Store})
		if err != nil {
			t.Fatal(err)
		}
		if _, err := member.Write(files[name]); err != nil {
			t.Fatal(err)
		}
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	refs, err := reference_data.DecodeIncidentBundleReferences(referenceVector)
	if err != nil {
		t.Fatal(err)
	}
	for index := range refs.Versions {
		if refs.Versions[index].Key == reference.Key {
			refs.Versions[index] = reference
		}
	}
	for index := range refs.Sets[0].Members {
		if refs.Sets[0].Members[index].Key == reference.Key {
			refs.Sets[0].Members[index] = reference.SetMember
		}
	}
	refs.Sets[0], err = packformat.BuildSet(refs.Sets[0].Members)
	if err != nil {
		t.Fatal(err)
	}
	references, err := reference_data.EncodeIncidentBundleReferences(refs.Sets, refs.Versions)
	if err != nil {
		t.Fatal(err)
	}
	content := packformat.EmptyPortableContent()
	content.Containers = []packformat.PortableContainer{{ManifestSHA256: reference.ManifestSHA256, ContainerSHA256: digest(out.Bytes()), SizeBytes: int64(out.Len())}}
	content.RequiredMembers = []packformat.PortableRequiredMember{{SetID: refs.Sets[0].ID, Key: reference.Key}}
	encoded, err := packformat.EncodePortableContent(content, refs)
	if err != nil {
		t.Fatal(err)
	}
	path, err := packformat.PortableContainerPath(reference.ManifestSHA256)
	if err != nil {
		t.Fatal(err)
	}
	bootstrap := canonical(map[string]any{"schema_id": "cartulary.reference_pack_trust_bootstrap.v1", "repositories": []any{map[string]any{"repository_id": "fixture.repo", "trusted_root": json.RawMessage(root), "trusted_root_sha256": digest(root)}}})
	return PortablePackFixture{Bootstrap: bootstrap, Container: out.Bytes(), References: references, Content: encoded, Path: path, Key: reference.Key, Version: reference.Version}
}
