package incidentbundles_test

import (
	"github.com/JochiRaider/cartulary/internal/modules/reference_data/producer"
	referencetest "github.com/JochiRaider/cartulary/internal/modules/reference_data/testsupport"
	"testing"
	"time"
)

func TestWindowsProducedPackSurvivesBundleRoundTrip_Integration(t *testing.T) {
	keys, err := producer.NewWindowsFixtureSigningContext()
	if err != nil {
		t.Fatal(err)
	}
	f := referencetest.WindowsFixture(t, "fixture-a", time.Now().UTC(), keys, 3)
	runNativeSnapshotRoundTrip(t, roundTripPackFixture{f.PortablePackFixture, referencetest.WindowsSetDigest(t, "../../..", f)}, true)
}
