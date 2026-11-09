package incidentbundles_test

import (
	referencetest "github.com/JochiRaider/cartulary/internal/modules/reference_data/testsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/referencepackfixture"
	"testing"
	"time"
)

func TestATTACKSnapshotReferencesSurviveBundleRoundTrip_Integration(t *testing.T) {
	t.Run("production MITRE license permits exact embedding", func(t *testing.T) {
		f := referencetest.ATTACKRelease(t, "../../..")
		runNativeSnapshotRoundTrip(t, roundTripPackFixture{f.PortablePackFixture, f.SetSHA256}, true)
	})
	t.Run("restricted fixture stays references-only", func(t *testing.T) {
		keys, err := referencepackfixture.NewSigningContext("fixture.attack.dataset.v1", "framework.attack")
		if err != nil {
			t.Fatal(err)
		}
		f := referencetest.AuthenticATTACKFixture(t, "../../..", "20.1", time.Now().UTC(), keys, 3)
		runNativeSnapshotRoundTrip(t, roundTripPackFixture{f.PortablePackFixture, f.SetSHA256}, false)
	})
}
