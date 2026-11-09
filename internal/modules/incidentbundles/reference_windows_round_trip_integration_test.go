package incidentbundles_test

import (
	referencetest "github.com/JochiRaider/cartulary/internal/modules/reference_data/testsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/referencepackfixture"
	"testing"
	"time"
)

func TestWindowsProducedPackSurvivesBundleRoundTrip_Integration(t *testing.T) {
	keys, err := referencepackfixture.NewSigningContext("fixture.windows.dataset.v1", "enrichment.windows_event_ids")
	if err != nil {
		t.Fatal(err)
	}
	f := referencetest.WindowsFixture(t, "fixture-a", time.Now().UTC(), keys, 3)
	runNativeSnapshotRoundTrip(t, roundTripPackFixture{f.PortablePackFixture, referencetest.WindowsSetDigest(t, "../../..", f)}, true)
}
