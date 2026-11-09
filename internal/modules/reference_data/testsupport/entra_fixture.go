package testsupport

import (
	"github.com/JochiRaider/cartulary/internal/testutil/referencepackfixture"
	"testing"
	"time"
)

type ProducedEntraPack struct {
	PortablePackFixture
	ManifestSHA256, PayloadSHA256 string
	Members                       map[string][]byte
}

func EntraSetDigest(t testing.TB, repositoryRoot string, f ProducedEntraPack) string {
	t.Helper()
	return producedSetDigest(t, repositoryRoot, f.PortablePackFixture, f.ManifestSHA256, f.PayloadSHA256, "cartulary.reference_pack.dataset.v1")
}

// EntraRelease reads only canonical public artifacts and the independent JSON
// baseline. Upstream source documents and research retention are never test inputs.
func EntraRelease(t testing.TB, repositoryRoot string) ProducedEntraPack {
	t.Helper()
	f, m := retainedDataset(t, repositoryRoot, "entra", "enrichment.entra_app_ids", "cartulary.entra_app_ids", "release", 139, 0)
	if err := checkEntraBaseline(repositoryRoot, m, f.Members["payload/objects.ndjson"]); err != nil {
		t.Fatal(err)
	}
	return ProducedEntraPack{f.PortablePackFixture, f.ManifestSHA256, f.PayloadSHA256, f.Members}
}

// EntraFixture signs a canonical generic corpus derivative with isolated test
// authority. Exact retained release qualification never calls this helper.
func EntraFixture(t testing.TB, version string, now time.Time, keys *referencepackfixture.SigningContext, metadataVersion int64) ProducedEntraPack {
	t.Helper()
	f := corpusDatasetFixture(t, "entra", version, now, keys, metadataVersion)
	return ProducedEntraPack{f.PortablePackFixture, f.ManifestSHA256, f.PayloadSHA256, f.Members}
}
