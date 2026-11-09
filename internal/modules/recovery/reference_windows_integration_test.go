package recovery_test

import (
	referencetest "github.com/JochiRaider/cartulary/internal/modules/reference_data/testsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/referencepackfixture"
	"testing"
	"time"
)

func TestWindowsProducedPackRestoresExactProvenance_Integration(t *testing.T) {
	keys, err := referencepackfixture.NewSigningContext("fixture.windows.dataset.v1", "enrichment.windows_event_ids")
	if err != nil {
		t.Fatal(err)
	}
	f := referencetest.WindowsFixture(t, "fixture-a", time.Now().UTC(), keys, 1)
	runBackingStorageMetadataRestore(t, &recoveryPackFixture{f.PortablePackFixture, f.ManifestSHA256, f.PayloadSHA256, referencetest.WindowsSetDigest(t, "../../..", f)})
}
