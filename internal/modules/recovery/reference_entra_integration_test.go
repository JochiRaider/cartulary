package recovery_test

import (
	referencetest "github.com/JochiRaider/cartulary/internal/modules/reference_data/testsupport"
	"github.com/JochiRaider/cartulary/internal/testutil/referencepackfixture"
	"testing"
	"time"
)

func TestEntraProducedPackRestoresExactProvenance_Integration(t *testing.T) {
	expected := referencetest.OwnerFixture(t, "entra_recovery")
	if expected.Condition == nil || *expected.Condition != "verified_available" {
		t.Fatal("Entra owner expectation")
	}
	keys, err := referencepackfixture.NewSigningContext("fixture.entra.dataset.v1", "enrichment.entra_app_ids")
	if err != nil {
		t.Fatal(err)
	}
	now := time.Now().UTC()
	a := referencetest.EntraFixture(t, "fixture-a", now, keys, 1)
	f := referencetest.EntraFixture(t, "fixture-b", now, keys, 2)
	runBackingStorageMetadataRestore(t, &recoveryPackFixture{f.PortablePackFixture, f.ManifestSHA256, f.PayloadSHA256, referencetest.EntraSetDigest(t, "../../..", f)}, &recoveryPackFixture{a.PortablePackFixture, a.ManifestSHA256, a.PayloadSHA256, referencetest.EntraSetDigest(t, "../../..", a)})
}
