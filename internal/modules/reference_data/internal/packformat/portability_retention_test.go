package packformat

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

func TestPortableRetentionClosedShapeAndCatalogBinding_Unit(t *testing.T) {
	testAttemptResultBoundary(t)
	t.Run("frozen portable verification context", testPortableVerificationContext)
	t.Run("embedded content and exact requirements", testPortableContentManifest)
	fixture, err := os.ReadFile("../../../../../contracts/reference-pack-fixtures/fixtures/portable-references.v1.json")
	if err != nil {
		t.Fatal(err)
	}
	refs, err := DecodePortableReferences(fixture)
	if err != nil {
		t.Fatal(err)
	}
	canonical, err := canonicaljson.Marshal(refs)
	if err != nil {
		t.Fatal(err)
	}
	input := map[string]any{"schema_id": "cartulary.reference_pack_portability_input.v1", "incident_id": "58000000-0000-4000-8000-000000000001", "source_operation_id": "58000000-0000-4000-8000-000000000002", "catalog_sha256": Digest(canonical), "content_manifest": EmptyPortableContent()}
	versions := []any{}
	for _, v := range refs.Versions {
		versions = append(versions, map[string]any{"pack_key": v.Key, "pack_version": v.Version, "available": false, "reason_code": "not_retained"})
	}
	result := map[string]any{"schema_id": "cartulary.reference_pack_portability_resolution.v1", "versions": versions}
	in, _ := canonicaljson.Marshal(input)
	out, _ := canonicaljson.Marshal(result)
	if err := ValidatePortabilityRetention(in, out, refs); err != nil {
		t.Fatal(err)
	}
	for _, boundary := range []string{"input", "result", "version"} {
		original := input
		if boundary == "result" {
			original = result
		}
		if boundary == "version" {
			original = versions[0].(map[string]any)
		}
		for field := range original {
			for _, kind := range []string{"omitted", "null", "wrong_type"} {
				t.Run(boundary+"/"+field+"/"+kind, func(t *testing.T) {
					var i, o map[string]any
					_ = json.Unmarshal(in, &i)
					_ = json.Unmarshal(out, &o)
					target := i
					if boundary == "result" {
						target = o
					}
					if boundary == "version" {
						target = o["versions"].([]any)[0].(map[string]any)
					}
					switch kind {
					case "omitted":
						delete(target, field)
					case "null":
						target[field] = nil
					default:
						target[field] = []any{}
					}
					a, _ := canonicaljson.Marshal(i)
					b, _ := canonicaljson.Marshal(o)
					if err := ValidatePortabilityRetention(a, b, refs); err == nil {
						t.Fatal("malformed retained fact admitted")
					}
				})
			}
		}
		original["hostile_unknown"] = true
		a, _ := canonicaljson.Marshal(input)
		b, _ := canonicaljson.Marshal(result)
		if err := ValidatePortabilityRetention(a, b, refs); err == nil {
			t.Fatal("unknown retained member admitted")
		}
		delete(original, "hostile_unknown")
	}
	input["catalog_sha256"] = strings.Repeat("0", 64)
	changed, _ := canonicaljson.Marshal(input)
	if ValidatePortabilityRetention(changed, out, refs) == nil {
		t.Fatal("catalog substitution admitted")
	}
	versions[0].(map[string]any)["available"] = true
	changed, _ = canonicaljson.Marshal(result)
	if ValidatePortabilityRetention(in, changed, refs) == nil {
		t.Fatal("contradictory availability admitted")
	}
	versions[0].(map[string]any)["reason_code"] = nil
	changed, _ = canonicaljson.Marshal(result)
	if err := ValidatePortabilityRetention(in, changed, refs); err != nil {
		t.Fatal("explicit null success reason rejected", err)
	}
	if ValidatePortabilityRetention(append(in, '\n'), out, refs) == nil {
		t.Fatal("noncanonical retained evidence admitted")
	}
	if !admittedPortableReferenceSize(67108864) || admittedPortableReferenceSize(67108865) || admittedPortableReferenceSize(0) {
		t.Fatal("catalog byte guard boundary")
	}
	for _, limit := range []struct {
		field   string
		maximum int
	}{{"sets", 1024}, {"versions", 16384}} {
		rule := portableReferencesShape.properties[limit.field]
		value, err := canonicaljson.DecodeStrict(fixture)
		if err != nil {
			t.Fatal(err)
		}
		item := value.(map[string]any)[limit.field].([]any)[0]
		items := make([]any, limit.maximum+1)
		for index := range items {
			items[index] = item
		}
		if *rule.maxItems != limit.maximum || !rule.matches(items[:limit.maximum]) || rule.matches(items) {
			t.Fatal("catalog cardinality guard", limit.field)
		}
	}
}

func TestPortableCatalogReachableSetAndVersionLimits_Unit(t *testing.T) {
	t.Run("embedded content limit", testPortableContentReachableLimits)
	// Independently specified catalog, with an expected digest calculated by a
	// separate Python standard-library encoder/hash. No BuildSet/producer helper
	// supplies the expected admission result or fixture identities.
	keys := []string{"type_registry.host", "type_registry.evidence", "type_registry.indicator", "framework.attack", "framework.d3fend", "framework.veris", "enrichment.tor", "enrichment.cisa_kev", "enrichment.ms_portals", "enrichment.windows_event_ids", "enrichment.entra_app_ids", "enrichment.lolbas", "enrichment.loldrivers", "enrichment.lolesxi", "enrichment.hijacklibs", "enrichment.windows_sids"}
	slices.Sort(keys)
	refs := PortableReferences{SchemaID: "reference_pack_refs.v1", Sets: []Set{}, Versions: []PortableVersion{}}
	for i := range 1024 {
		members := []SetMember{}
		for _, key := range keys {
			member := SetMember{Key: key, Version: fmt.Sprintf("v%04d", i), ManifestSHA256: strings.Repeat("1", 64), PayloadSHA256: strings.Repeat("2", 64), Contract: "cartulary.reference_pack_contract.v1", ProfileID: "cartulary.reference_pack." + key + ".v1", ProfileVersion: "1"}
			members = append(members, member)
			refs.Versions = append(refs.Versions, PortableVersion{SetMember: member, DistributionKind: "operator_imported", VerificationMethod: "tuf_1_0_35_offline_bundle_v1", SourceProfileID: "fixture.portable.v1", SourceProfileSHA256: strings.Repeat("3", 64)})
		}
		preimage, err := canonicaljson.Marshal(map[string]any{"schema_id": "cartulary.reference_pack_set.v1", "members": members})
		if err != nil {
			t.Fatal(err)
		}
		digest := Digest(preimage)
		refs.Sets = append(refs.Sets, Set{SchemaID: "cartulary.reference_pack_set.v1", ID: "rpset_" + digest, SHA256: digest, Members: members})
	}
	slices.SortFunc(refs.Sets, func(a, b Set) int { return strings.Compare(a.ID, b.ID) })
	slices.SortFunc(refs.Versions, func(a, b PortableVersion) int { return strings.Compare(a.Key+"\x00"+a.Version, b.Key+"\x00"+b.Version) })
	encoded, err := canonicaljson.Marshal(refs)
	if err != nil {
		t.Fatal(err)
	}
	if len(encoded) != 16801852 || Digest(encoded) != "77c739df8eb8f288d36f3d926984fe373f879b4878236ff38e63db3d2a59c24e" {
		t.Fatal("independent at-limit catalog changed")
	}
	admitted, err := DecodePortableReferences(encoded)
	if err != nil || len(admitted.Sets) != 1024 || len(admitted.Versions) != 16384 {
		t.Fatal("reachable catalog equality rejected", err)
	}
	// Whitespace is admitted at this unsigned transport boundary. It makes the
	// byte ceiling reachable without changing semantic cardinality constraints.
	empty := []byte(`{"schema_id":"reference_pack_refs.v1","sets":[],"versions":[]}`)
	atLimit := append(empty, bytes.Repeat([]byte{' '}, 67108864-len(empty))...)
	if _, err := DecodePortableReferences(atLimit); err != nil {
		t.Fatal("complete byte-limit catalog rejected", err)
	}
	if _, err := DecodePortableReferences(append(atLimit, ' ')); err == nil {
		t.Fatal("over-byte-limit catalog admitted")
	}
}
