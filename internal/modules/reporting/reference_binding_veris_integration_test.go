package reporting_test

import (
	referencetest "github.com/JochiRaider/cartulary/internal/modules/reference_data/testsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/referencepackfixture"
	"testing"
	"time"
)

func TestVERISSnapshotAdmissionRetainsProducedSet_Integration(t *testing.T) {
	now := time.Now().UTC()
	keys, err := referencepackfixture.NewSigningContext("fixture.veris.dataset.v1", "framework.veris")
	if err != nil {
		t.Fatal(err)
	}
	a := referencetest.VERISFixture(t, "fixture-a", now, keys, 2)
	b := referencetest.VERISFixture(t, "fixture-b", now, keys, 3)
	previous := bindingFixture{a.PortablePackFixture, a.ManifestSHA256, referencetest.VERISSetDigest(t, "../../..", a)}
	fixture := bindingFixture{b.PortablePackFixture, b.ManifestSHA256, referencetest.VERISSetDigest(t, "../../..", b)}
	runSnapshotReferenceBinding(t, referencetest.OwnerFixture(t, "veris_snapshot_binding"), previous, fixture)
}
