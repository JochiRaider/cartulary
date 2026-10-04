package testsupport

import (
	"encoding/json"
	"testing"

	"github.com/JochiRaider/cartulary/internal/gen/contractreferencepacks"
)

// OwnerFixtureExpectation carries independently authored canonical identities
// into live cross-owner tests. Format tests separately validate closed manifest
// and scenario shapes; loading expectations does not itself execute evidence.
type OwnerFixtureExpectation struct {
	Manifest  *string `json:"expected_manifest_sha256"`
	Payload   *string `json:"expected_payload_sha256"`
	Set       *string `json:"expected_pack_set_sha256"`
	Condition *string `json:"expected_condition"`
	Error     *struct {
		Code   string  `json:"code"`
		Reason *string `json:"reason_code"`
	} `json:"expected_public_error"`
}

func OwnerFixture(t testing.TB, scenario string) OwnerFixtureExpectation {
	t.Helper()
	for _, artifact := range contractreferencepacks.Artifacts {
		if artifact.Path != "contracts/reference-packs/fixture-manifests/owner_"+scenario+".v1.json" {
			continue
		}
		var expected OwnerFixtureExpectation
		if err := json.Unmarshal([]byte(artifact.JSON), &expected); err != nil || expected.Set == nil {
			t.Fatal("invalid canonical owner fixture", err)
		}
		return expected
	}
	t.Fatal("canonical owner fixture is missing", scenario)
	return OwnerFixtureExpectation{}
}
