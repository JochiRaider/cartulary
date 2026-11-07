package reporting_test

import (
	"github.com/JochiRaider/cartulary/internal/modules/reference_data/producer"
	referencetest "github.com/JochiRaider/cartulary/internal/modules/reference_data/testsupport"
	"testing"
	"time"
)

func TestWindowsSnapshotAdmissionRetainsProducedSet_Integration(t *testing.T) {
	now := time.Now().UTC()
	keys, err := producer.NewWindowsFixtureSigningContext()
	if err != nil {
		t.Fatal(err)
	}
	a := referencetest.WindowsFixture(t, "fixture-a", now, keys, 2)
	b := referencetest.WindowsFixture(t, "fixture-b", now, keys, 3)
	previous := bindingFixture{a.PortablePackFixture, a.ManifestSHA256, referencetest.WindowsSetDigest(t, "../../..", a)}
	fixture := bindingFixture{b.PortablePackFixture, b.ManifestSHA256, referencetest.WindowsSetDigest(t, "../../..", b)}
	runSnapshotReferenceBinding(t, referencetest.OwnerFixture(t, "windows_snapshot_binding"), previous, fixture)
}
