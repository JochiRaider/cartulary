package testsupport

import (
	"github.com/JochiRaider/cartulary/internal/testutil/referencepackfixture"
	"testing"
	"time"
)

type ProducedVERISPack struct {
	PortablePackFixture
	ManifestSHA256, PayloadSHA256 string
	Members                       map[string][]byte
}

func VERISSetDigest(t testing.TB, repositoryRoot string, f ProducedVERISPack) string {
	t.Helper()
	return producedSetDigest(t, repositoryRoot, f.PortablePackFixture, f.ManifestSHA256, f.PayloadSHA256, "cartulary.reference_pack.dataset.v1")
}

// VERISRelease reads only canonical public artifacts and their JSON descriptor. Upstream source documents and research retention are never test inputs.
func VERISRelease(t testing.TB, repositoryRoot string) ProducedVERISPack {
	t.Helper()
	f, _ := retainedDataset(t, repositoryRoot, "veris", "framework.veris", "cartulary.veris.community", "release", 2491, 2533)
	return ProducedVERISPack{f.PortablePackFixture, f.ManifestSHA256, f.PayloadSHA256, f.Members}
}

// VERISFixture signs a canonical generic corpus derivative with isolated test
// authority. Exact retained release qualification never calls this helper.
func VERISFixture(t testing.TB, version string, now time.Time, keys *referencepackfixture.SigningContext, metadataVersion int64) ProducedVERISPack {
	t.Helper()
	f := corpusDatasetFixture(t, "veris", version, now, keys, metadataVersion)
	return ProducedVERISPack{f.PortablePackFixture, f.ManifestSHA256, f.PayloadSHA256, f.Members}
}
