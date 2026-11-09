package testsupport

import (
	"github.com/JochiRaider/cartulary/internal/testutil/referencepackfixture"
	"github.com/JochiRaider/cartulary/referencepacks/format/dataset"
	"testing"
	"time"
)

type AuthenticATTACKPack struct {
	PortablePackFixture
	ManifestSHA256, PayloadSHA256, SetSHA256 string
	Objects, Relationships                   int64
}

// AuthenticATTACKFixture uses immutable canonical generic artifacts derived
// from the selected source release. Only its fixture envelope is renewed.
func AuthenticATTACKFixture(t testing.TB, repositoryRoot, sourceVersion string, now time.Time, keys *referencepackfixture.SigningContext, metadataVersion int64) AuthenticATTACKPack {
	t.Helper()
	m, members := corpusDatasetMembers(t, repositoryRoot, "attack/"+sourceVersion)
	if m.SourceVersion != sourceVersion {
		t.Fatal("authentic source version mismatch")
	}
	repository := keys.Repository()
	m.Repository = &repository
	f := signDataset(t, "attack", now, keys, metadataVersion, m, members)
	set := producedSetDigest(t, repositoryRoot, f.PortablePackFixture, f.ManifestSHA256, f.PayloadSHA256, dataset.ProfileID)
	return AuthenticATTACKPack{f.PortablePackFixture, f.ManifestSHA256, f.PayloadSHA256, set, *m.Summary.Objects, *m.Summary.Relationships}
}
