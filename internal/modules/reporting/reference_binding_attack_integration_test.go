package reporting_test

import (
	referencetest "github.com/JochiRaider/cartulary/internal/modules/reference_data/testsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/referencepackfixture"
	"testing"
	"time"
)

func TestATTACKSnapshotAdmissionRetainsProducedSet_Integration(t *testing.T) {
	now := time.Now().UTC()
	keys, err := referencepackfixture.NewSigningContext("fixture.attack.dataset.v1", "framework.attack")
	if err != nil {
		t.Fatal(err)
	}
	a := referencetest.AuthenticATTACKFixture(t, "../../..", "20.0", now, keys, 2)
	b := referencetest.AuthenticATTACKFixture(t, "../../..", "20.1", now, keys, 3)
	previous := bindingFixture{a.PortablePackFixture, a.ManifestSHA256, a.SetSHA256}
	fixture := bindingFixture{b.PortablePackFixture, b.ManifestSHA256, b.SetSHA256}
	runSnapshotReferenceBinding(t, referencetest.OwnerFixture(t, "snapshot_binding"), previous, fixture)
}
