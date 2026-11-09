package testsupport

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/JochiRaider/cartulary/internal/testutil/referencepackfixture"
	"github.com/JochiRaider/cartulary/referencepacks/format/dataset"
)

type ProducedWindowsPack struct {
	PortablePackFixture
	ManifestSHA256, PayloadSHA256 string
	Members                       map[string][]byte
}

func WindowsSetDigest(t testing.TB, repositoryRoot string, f ProducedWindowsPack) string {
	t.Helper()
	return producedSetDigest(t, repositoryRoot, f.PortablePackFixture, f.ManifestSHA256, f.PayloadSHA256, dataset.ProfileID)
}

func producedSetDigest(t testing.TB, repositoryRoot string, f PortablePackFixture, manifestSHA, payloadSHA, profile string) string {
	t.Helper()
	b, e := os.ReadFile(filepath.Join(repositoryRoot, "contracts/reference-packs/builtins/release.v3.json"))
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
		m := map[string]any{"pack_contract_version": "cartulary.reference_pack_contract.v3"}
		for _, k := range []string{"pack_key", "pack_version", "manifest_sha256", "payload_sha256", "content_profile_id", "content_profile_version"} {
			m[k] = p.Binding[k]
		}
		members = append(members, m)
	}
	if len(members) != 3 {
		t.Fatal("expected three Base registries")
	}
	members = append(members, map[string]any{"pack_key": f.Key, "pack_version": f.Version, "manifest_sha256": manifestSHA, "payload_sha256": payloadSHA, "pack_contract_version": "cartulary.reference_pack_contract.v3", "content_profile_id": profile, "content_profile_version": "1"})
	slices.SortFunc(members, func(a, b map[string]any) int { return strings.Compare(a["pack_key"].(string), b["pack_key"].(string)) })
	b, e = json.Marshal(map[string]any{"schema_id": "cartulary.reference_pack_set.v3", "members": members})
	if e != nil {
		t.Fatal(e)
	}
	sum := sha256.Sum256(b)
	return hex.EncodeToString(sum[:])
}

// WindowsFixture reads a small, authored canonical generic artifact. It covers
// provider ambiguity and absent/numeric versions without running a producer or
// parsing source documents. Exact real-release qualification is separate.
func WindowsFixture(t testing.TB, version string, now time.Time, keys *referencepackfixture.SigningContext, metadataVersion int64) ProducedWindowsPack {
	t.Helper()
	members := map[string][]byte{}
	for _, name := range []string{dataset.DescriptorPath, dataset.ObjectsPath, dataset.RelationshipsPath, "notices/fixture.txt"} {
		raw, err := os.ReadFile(filepath.Join("../../..", "fixtures/reference-packs/dataset-v1/windows-synthetic", name))
		if err != nil {
			t.Fatal(err)
		}
		members[name] = raw
	}
	f := signSyntheticDataset(t, "windows", version, now, keys, metadataVersion, members, 4, 0)
	return ProducedWindowsPack{f.PortablePackFixture, f.ManifestSHA256, f.PayloadSHA256, f.Members}
}
