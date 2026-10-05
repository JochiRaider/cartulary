package packformat

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

func testPortableContentManifest(t *testing.T) {
	var fixture struct {
		References json.RawMessage `json:"references"`
		Content    json.RawMessage `json:"content"`
		Path       string          `json:"expected_container_path"`
		Required   struct {
			Key     string `json:"pack_key"`
			Version string `json:"pack_version"`
		} `json:"expected_required_version"`
	}
	data, err := os.ReadFile("../../../../../contracts/reference-pack-fixtures/fixtures/portable-content.v1.json")
	if err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(data, &fixture); err != nil {
		t.Fatal(err)
	}
	refs, err := DecodePortableReferences(fixture.References)
	if err != nil {
		t.Fatal(err)
	}
	content, err := DecodePortableContent(fixture.Content, refs)
	if err != nil {
		t.Fatal(err)
	}
	path, err := PortableContainerPath(content.Containers[0].ManifestSHA256)
	if err != nil || path != fixture.Path {
		t.Fatalf("path: %q %v", path, err)
	}
	if err := ValidatePortableContainerInventory(content, []string{path}); err != nil {
		t.Fatal(err)
	}
	if err := ValidatePortableContainerInventory(content, nil); err != nil {
		t.Fatal("optional missing bytes rejected before resolution", err)
	}
	for _, paths := range [][]string{{path, path}, {"ext/reference_packs/alias.zip"}, {"ext/reference_packs/containers/" + strings.Repeat("e", 64) + ".pack"}} {
		if err := ValidatePortableContainerInventory(content, paths); err == nil {
			t.Fatal("unbound inventory admitted")
		}
	}
	required := RequiredPortableVersions(content, refs)
	if len(required) != 1 || !required[fixture.Required.Key+"\x00"+fixture.Required.Version] {
		t.Fatalf("required selection: %v", required)
	}
	encoded, err := EncodePortableContent(content, refs)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := DecodePortableContent(encoded, refs); err != nil {
		t.Fatal(err)
	}
	empty, err := EncodePortableContent(EmptyPortableContent(), refs)
	if err != nil || string(empty) != "{\"containers\":[],\"required_members\":[],\"schema_id\":\"reference_pack_content.v1\"}\n" {
		t.Fatalf("empty content: %s %v", empty, err)
	}
	for _, invalid := range []string{"", "null", "[]", `{"schema_id":"reference_pack_content.v1","containers":[],"containers":[],"required_members":[]}`, `{"schema_id":"\ud800","containers":[],"required_members":[]}`} {
		if _, err := DecodePortableContent([]byte(invalid), refs); err == nil {
			t.Fatal("malformed content admitted", invalid)
		}
	}
	// Exercise every closed nested object boundary without repairing input.
	var root map[string]any
	if err := json.Unmarshal(fixture.Content, &root); err != nil {
		t.Fatal(err)
	}
	assertInvalid := func(want string) {
		t.Helper()
		b, err := canonicaljson.Marshal(root)
		if err != nil {
			t.Fatal(err)
		}
		_, err = DecodePortableContent(b, refs)
		var failure *Failure
		if !errors.As(err, &failure) || failure.Code != want {
			t.Fatalf("expected %s: %v", want, err)
		}
	}
	var visit func(any)
	visit = func(v any) {
		switch node := v.(type) {
		case map[string]any:
			for key, original := range node {
				delete(node, key)
				assertInvalid("reference_pack_refs.exact_shape")
				node[key] = nil
				assertInvalid("reference_pack_refs.exact_shape")
				node[key] = true
				assertInvalid("reference_pack_refs.exact_shape")
				node[key] = original
			}
			node["private_unknown"] = "secret"
			assertInvalid("reference_pack_refs.exact_shape")
			delete(node, "private_unknown")
			for _, child := range node {
				visit(child)
			}
		case []any:
			for _, child := range node {
				visit(child)
			}
		}
	}
	visit(root)
	for _, field := range []string{"containers", "required_members"} {
		original := root[field]
		items := original.([]any)
		root[field] = []any{items[0], items[0]}
		assertInvalid("reference_pack_refs.identity_exact")
		root[field] = original
	}
	container := root["containers"].([]any)[0].(map[string]any)
	original := container["manifest_sha256"]
	for _, digest := range []string{strings.Repeat("e", 64), refs.Versions[1].ManifestSHA256} {
		container["manifest_sha256"] = digest
		assertInvalid("reference_pack_refs.identity_exact")
	}
	container["manifest_sha256"] = original
	member := root["required_members"].([]any)[0].(map[string]any)
	original = member["pack_key"]
	member["pack_key"] = "vendor.absent"
	assertInvalid("reference_pack_refs.identity_exact")
	member["pack_key"] = original
	for _, bad := range []string{"../escape", strings.Repeat("A", 64), strings.Repeat("0", 63)} {
		if _, err := PortableContainerPath(bad); err == nil {
			t.Fatal("unsafe path token accepted")
		}
	}
	for _, field := range []struct {
		name  string
		limit int
	}{{"containers", 16384}, {"required_members", 65536}} {
		rule := portableContentShape.properties[field.name]
		items := make([]any, field.limit+1)
		value, err := canonicaljson.DecodeStrict(fixture.Content)
		if err != nil {
			t.Fatal(err)
		}
		item := value.(map[string]any)[field.name].([]any)[0]
		for i := range items {
			items[i] = item
		}
		if *rule.maxItems != field.limit || !rule.matches(items[:field.limit]) || rule.matches(items) {
			t.Fatal("content cardinality guard", field.name)
		}
	}
	original = container["size_bytes"]
	for _, invalid := range []any{0, -1, 1.5, 9007199254740992.0, "1"} {
		container["size_bytes"] = invalid
		assertInvalid("reference_pack_refs.exact_shape")
	}
	for _, valid := range []any{1, 9007199254740991.0} {
		container["size_bytes"] = valid
		b, err := canonicaljson.Marshal(root)
		if err != nil {
			t.Fatal(err)
		}
		if _, err = DecodePortableContent(b, refs); err != nil {
			t.Fatal("size guard equality", err)
		}
	}
	container["size_bytes"] = original
	// The document byte equality is reachable through admitted JSON whitespace.
	atLimit := append(bytes.Repeat([]byte{' '}, 16777216-len(empty)), empty...)
	if _, err := DecodePortableContent(atLimit, refs); err != nil {
		t.Fatal("at-limit content rejected", err)
	}
	if _, err := DecodePortableContent(append(atLimit, ' '), refs); err == nil {
		t.Fatal("over-limit content admitted")
	}
}

// Independent ASCII catalog recipe. Expected bytes/digests were calculated
// separately with Python json.dumps(sort_keys=True,separators=(",",":")) and
// hashlib.sha256, not this producer or decoder. No fixture reads Markdown.
func testPortableContentReachableLimits(t *testing.T) {
	keys := []string{"type_registry.host", "type_registry.evidence", "type_registry.indicator", "framework.attack", "framework.d3fend", "framework.veris", "enrichment.tor", "enrichment.cisa_kev", "enrichment.ms_portals", "enrichment.windows_event_ids", "enrichment.entra_app_ids", "enrichment.lolbas", "enrichment.loldrivers", "enrichment.lolesxi", "enrichment.hijacklibs", "enrichment.windows_sids"}
	slices.Sort(keys)
	refs := PortableReferences{SchemaID: "reference_pack_refs.v1", Sets: []Set{}, Versions: []PortableVersion{}}
	content := EmptyPortableContent()
	for i := range 1024 {
		members := []SetMember{}
		for _, key := range keys {
			version := fmt.Sprintf("v%04d", i)
			manifest := Digest([]byte("manifest:" + key + "/" + version))
			member := SetMember{Key: key, Version: version, ManifestSHA256: manifest, PayloadSHA256: strings.Repeat("2", 64), Contract: "cartulary.reference_pack_contract.v1", ProfileID: "cartulary.reference_pack." + key + ".v1", ProfileVersion: "1"}
			members = append(members, member)
			refs.Versions = append(refs.Versions, PortableVersion{SetMember: member, DistributionKind: "operator_imported", VerificationMethod: "tuf_1_0_35_offline_bundle_v1", SourceProfileID: "fixture.portable.v1", SourceProfileSHA256: strings.Repeat("3", 64)})
			content.Containers = append(content.Containers, PortableContainer{ManifestSHA256: manifest, ContainerSHA256: strings.Repeat("4", 64), SizeBytes: 1})
		}
		preimage, err := canonicaljson.Marshal(map[string]any{"schema_id": "cartulary.reference_pack_set.v1", "members": members})
		if err != nil {
			t.Fatal(err)
		}
		digest := Digest(preimage)
		id := "rpset_" + digest
		refs.Sets = append(refs.Sets, Set{SchemaID: "cartulary.reference_pack_set.v1", ID: id, SHA256: digest, Members: members})
		for _, key := range keys {
			content.RequiredMembers = append(content.RequiredMembers, PortableRequiredMember{SetID: id, Key: key})
		}
	}
	slices.SortFunc(refs.Sets, func(a, b Set) int { return strings.Compare(a.ID, b.ID) })
	slices.SortFunc(refs.Versions, func(a, b PortableVersion) int { return strings.Compare(a.Key+"\x00"+a.Version, b.Key+"\x00"+b.Version) })
	encoded, err := canonicaljson.Marshal(refs)
	if err != nil {
		t.Fatal(err)
	}
	if len(encoded) != 16801852 || Digest(encoded) != "3d9bc8208997cd162253dc1dcc2edffdd2316c6fe5d7d5a5fbb70a4a979466c1" {
		t.Fatal("independent catalog changed")
	}
	admitted, err := DecodePortableReferences(encoded)
	if err != nil {
		t.Fatal(err)
	}
	slices.SortFunc(content.Containers, func(a, b PortableContainer) int { return strings.Compare(a.ManifestSHA256, b.ManifestSHA256) })
	slices.SortFunc(content.RequiredMembers, func(a, b PortableRequiredMember) int {
		return strings.Compare(a.SetID+"\x00"+a.Key, b.SetID+"\x00"+b.Key)
	})
	encoded, err = canonicaljson.Marshal(content)
	if err != nil {
		t.Fatal(err)
	}
	if len(encoded) != 5091405 || Digest(encoded) != "78d7f14cc0a3914823e3bec2fe21b2f4385f524004292fe3176be22a8ccaed89" {
		t.Fatal("independent content changed")
	}
	got, err := DecodePortableContent(encoded, admitted)
	if err != nil || len(got.Containers) != 16384 || len(got.RequiredMembers) != 16384 {
		t.Fatal("complete content limit rejected", err)
	}
	content.Containers[0], content.Containers[1] = content.Containers[1], content.Containers[0]
	encoded, err = canonicaljson.Marshal(content)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := DecodePortableContent(encoded, admitted); err == nil {
		t.Fatal("unordered containers admitted")
	}
	content.Containers[0], content.Containers[1] = content.Containers[1], content.Containers[0]
	content.RequiredMembers[0], content.RequiredMembers[1] = content.RequiredMembers[1], content.RequiredMembers[0]
	encoded, err = canonicaljson.Marshal(content)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := DecodePortableContent(encoded, admitted); err == nil {
		t.Fatal("unordered required members admitted")
	}
}
