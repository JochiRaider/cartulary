package recovery_test

import (
	"github.com/JochiRaider/cartulary/internal/modules/reference_data/producer"
	referencetest "github.com/JochiRaider/cartulary/internal/modules/reference_data/testsupport"
	"testing"
	"time"
)

func TestWindowsProducedPackRestoresExactProvenance_Integration(t *testing.T) {
	keys, err := producer.NewWindowsFixtureSigningContext()
	if err != nil {
		t.Fatal(err)
	}
	f := referencetest.WindowsFixture(t, "fixture-a", time.Now().UTC(), keys, 1)
	runBackingStorageMetadataRestore(t, &recoveryPackFixture{f.PortablePackFixture, f.ManifestSHA256, f.PayloadSHA256, referencetest.WindowsSetDigest(t, "../../..", f)})
}
