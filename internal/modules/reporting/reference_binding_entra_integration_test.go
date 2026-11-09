package reporting_test

import (
	referencetest "github.com/JochiRaider/cartulary/internal/modules/reference_data/testsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/referencepackfixture"
	"testing"
	"time"
)

func TestEntraSnapshotAdmissionRetainsProducedSet_Integration(t *testing.T) {
	now := time.Now().UTC()
	keys, err := referencepackfixture.NewSigningContext("fixture.entra.dataset.v1", "enrichment.entra_app_ids")
	if err != nil {
		t.Fatal(err)
	}
	a := referencetest.EntraFixture(t, "fixture-a", now, keys, 2)
	b := referencetest.EntraFixture(t, "fixture-b", now, keys, 3)
	previous := bindingFixture{a.PortablePackFixture, a.ManifestSHA256, referencetest.EntraSetDigest(t, "../../..", a)}
	fixture := bindingFixture{b.PortablePackFixture, b.ManifestSHA256, referencetest.EntraSetDigest(t, "../../..", b)}
	runSnapshotReferenceBinding(t, referencetest.OwnerFixture(t, "entra_snapshot_binding"), previous, fixture)
}
