package incidentbundles_test

import (
	referencetest "github.com/JochiRaider/cartulary/internal/modules/reference_data/testsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/referencepackfixture"
	"testing"
	"time"
)

func TestVERISProducedPackSurvivesBundleRoundTrip_Integration(t *testing.T) {
	expected := referencetest.OwnerFixture(t, "veris_portability")
	if expected.Condition == nil || *expected.Condition != "verified_available" {
		t.Fatal("VERIS owner expectation")
	}
	keys, err := referencepackfixture.NewSigningContext("fixture.veris.dataset.v1", "framework.veris")
	if err != nil {
		t.Fatal(err)
	}
	for _, tc := range []struct {
		name, version string
		embed         bool
	}{{"allowed", "fixture-a", true}, {"restricted-fixture", "fixture-restricted", false}} {
		t.Run(tc.name, func(t *testing.T) {
			f := referencetest.VERISFixture(t, tc.version, time.Now().UTC(), keys, 3)
			runNativeSnapshotRoundTrip(t, roundTripPackFixture{f.PortablePackFixture, referencetest.VERISSetDigest(t, "../../..", f)}, tc.embed)
		})
	}
}
