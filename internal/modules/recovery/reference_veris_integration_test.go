package recovery_test

import (
	"github.com/JochiRaider/cartulary/internal/modules/reference_data/producer"
	referencetest "github.com/JochiRaider/cartulary/internal/modules/reference_data/testsupport"
	"testing"
	"time"
)

func TestVERISProducedPackRestoresExactProvenance_Integration(t *testing.T) {
	expected := referencetest.OwnerFixture(t, "veris_recovery")
	if expected.Condition == nil || *expected.Condition != "verified_available" {
		t.Fatal("VERIS owner expectation")
	}
	keys, err := producer.NewVERISFixtureSigningContext()
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	a := referencetest.VERISFixture(t, "fixture-a", now, keys, 1)
	f := referencetest.VERISFixture(t, "fixture-b", now, keys, 2)
	runBackingStorageMetadataRestore(t, &recoveryPackFixture{f.PortablePackFixture, f.ManifestSHA256, f.PayloadSHA256, referencetest.VERISSetDigest(t, "../../..", f)}, &recoveryPackFixture{a.PortablePackFixture, a.ManifestSHA256, a.PayloadSHA256, referencetest.VERISSetDigest(t, "../../..", a)})
}
