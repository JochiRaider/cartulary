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
	// ReleaseIssuedAt is the retained release envelope's issuance instant;
	// zero for a test-signed fixture. See RetainedReleaseClock.
	ReleaseIssuedAt time.Time
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
	return AuthenticATTACKPack{PortablePackFixture: f.PortablePackFixture, ManifestSHA256: f.ManifestSHA256, PayloadSHA256: f.PayloadSHA256, SetSHA256: set, Objects: *m.Summary.Objects, Relationships: *m.Summary.Relationships}
}
