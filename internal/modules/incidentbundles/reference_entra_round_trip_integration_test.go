package incidentbundles_test

import (
	"github.com/JochiRaider/cartulary/internal/modules/reference_data/producer"
	referencetest "github.com/JochiRaider/cartulary/internal/modules/reference_data/testsupport"
	"testing"
	"time"
)

func TestEntraProducedPackSurvivesBundleRoundTrip_Integration(t *testing.T) {
	expected := referencetest.OwnerFixture(t, "entra_portability")
	if expected.Condition == nil || *expected.Condition != "verified_available" {
		t.Fatal("Entra owner expectation")
	}
	keys, err := producer.NewEntraFixtureSigningContext()
	if err != nil {
		t.Fatal(err)
	}
	f := referencetest.EntraFixture(t, "fixture-a", time.Now().UTC(), keys, 3)
	runNativeSnapshotRoundTrip(t, roundTripPackFixture{f.PortablePackFixture, referencetest.EntraSetDigest(t, "../../..", f)}, true)
}
