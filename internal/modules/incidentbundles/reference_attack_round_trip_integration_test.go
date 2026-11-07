package incidentbundles_test

import (
	"github.com/JochiRaider/cartulary/internal/modules/reference_data/producer"
	referencetest "github.com/JochiRaider/cartulary/internal/modules/reference_data/testsupport"
	"testing"
	"time"
)

func TestATTACKSnapshotReferencesSurviveBundleRoundTrip_Integration(t *testing.T) {
	keys, err := producer.NewFixtureSigningContext()
	if err != nil {
		t.Fatal(err)
	}
	f := referencetest.AuthenticATTACKFixture(t, "../../..", "20.1", time.Now().UTC(), keys, 3)
	runNativeSnapshotRoundTrip(t, roundTripPackFixture{f.PortablePackFixture, f.SetSHA256}, false)
}
