package contractadmission

import (
	"bytes"
	"encoding/json"
	"os"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/reference_data/internal/packformat"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

func TestEntraCurrentContractAdmission_Unit(t *testing.T) {
	testCurrentContentProfileAdmission(t, "enrichment.entra_app_ids", "fixtures/entra-app-ids/release.v2.json")
}
func TestVERISCurrentContractAdmission_Unit(t *testing.T) {
	testCurrentContentProfileAdmission(t, "framework.veris", "fixtures/veris/release.v1.json")
}
func testCurrentContentProfileAdmission(t *testing.T, key, releasePath string) {
	read := func(path string) []byte {
		t.Helper()
		b, e := os.ReadFile("../../../../" + path)
		if e != nil {
			t.Fatal(e)
		}
		return b
	}
	encode := func(v any) []byte {
		t.Helper()
		b, e := canonicaljson.Marshal(v)
		if e != nil {
			t.Fatal(e)
		}
		return b
	}
	var release struct {
		LogicalRoot string `json:"logical_root"`
	}
	if err := json.Unmarshal(read(releasePath), &release); err != nil {
		t.Fatal(err)
	}
	manifest := read(release.LogicalRoot + "/manifest.json")
	var vector struct {
		Members []packformat.SetMember `json:"members"`
	}
	if err := json.Unmarshal(read("contracts/reference-pack-fixtures/fixtures/complete-set.v1.json"), &vector); err != nil {
		t.Fatal(err)
	}
	set, err := packformat.BuildSet(vector.Members)
	if err != nil {
		t.Fatal(err)
	}
	versions := []packformat.PortableVersion{}
	for _, m := range set.Members {
		versions = append(versions, packformat.PortableVersion{SetMember: m, DistributionKind: "operator_imported", VerificationMethod: "tuf_1_0_35_offline_bundle_v1", SourceProfileID: "fixture.current-profile.admission", SourceProfileSHA256: m.ManifestSHA256})
	}
	refs, err := packformat.EncodePortableReferences([]packformat.Set{set}, versions)
	if err != nil {
		t.Fatal(err)
	}
	// Transport producers append LF; durable catalogs retain the canonical object.
	refs = bytes.TrimSuffix(refs, []byte{'\n'})
	for _, tc := range []struct {
		kind    string
		current []byte
	}{{"manifest:operator_imported", manifest}, {"set", encode(set)}, {"references", refs}} {
		t.Run(tc.kind, func(t *testing.T) {
			if err := validate(tc.kind, tc.current); err != nil {
				t.Fatal("current retained state rejected", err)
			}
			old := bytes.ReplaceAll(tc.current, []byte("cartulary.reference_pack."+key+".v2"), []byte("cartulary.reference_pack."+key+".v1"))
			if bytes.Equal(old, tc.current) {
				t.Fatal("missing profile binding")
			}
			if err := validate(tc.kind, old); !IsIncompatible(err) {
				t.Fatal("unsupported predecessor profile admitted", err)
			}
		})
	}
	for i := range vector.Members {
		if vector.Members[i].Key == key {
			vector.Members[i].ProfileVersion = "1"
		}
	}
	if _, err := packformat.BuildSet(vector.Members); err == nil {
		t.Fatal("mixed profile version admitted")
	}
}
