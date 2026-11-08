package incidentbundles_test

import (
	"github.com/JochiRaider/cartulary/internal/modules/reference_data/producer"
	referencetest "github.com/JochiRaider/cartulary/internal/modules/reference_data/testsupport"
	"testing"
	"time"
)

func TestVERISProducedPackSurvivesBundleRoundTrip_Integration(t *testing.T) {
	expected := referencetest.OwnerFixture(t, "veris_portability")
	if expected.Condition == nil || *expected.Condition != "verified_available" {
		t.Fatal("VERIS owner expectation")
	}
	keys, err := producer.NewVERISFixtureSigningContext()
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
