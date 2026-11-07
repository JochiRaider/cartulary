package referenceadmissionfixture

import (
	"crypto/sha256"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

// seedEntra changes only the per-key profile generation in otherwise current
// shared contracts. Reporting cases retain a compatible set and reject only
// the stale provenance inside the snapshot/release/composition binding.
func seedEntra(t testing.TB, db *sql.DB, kind string) {
	t.Helper()
	encode := func(v any) []byte {
		t.Helper()
		b, e := canonicaljson.Marshal(v)
		if e != nil {
			t.Fatal(e)
		}
		return b
	}
	read := func(path string) []byte {
		t.Helper()
		b, e := os.ReadFile("../../../" + path)
		if e != nil {
			t.Fatal(e)
		}
		return b
	}
	decode := func(b []byte) map[string]any {
		t.Helper()
		var v map[string]any
		if e := json.Unmarshal(b, &v); e != nil {
			t.Fatal(e)
		}
		return v
	}
	exec := func(query string, args ...any) {
		t.Helper()
		if _, e := db.ExecContext(t.Context(), query, args...); e != nil {
			t.Fatal(kind, e)
		}
	}
	release := decode(read("fixtures/entra-app-ids/release.v2.json"))
	manifest := decode(read(release["logical_root"].(string) + "/manifest.json"))
	vector := decode(read("contracts/reference-pack-fixtures/fixtures/complete-set.v1.json"))
	members := []any{}
	for _, v := range vector["members"].([]any) {
		m := v.(map[string]any)
		if m["pack_key"] == "enrichment.entra_app_ids" || strings.HasPrefix(m["pack_key"].(string), "type_registry.") {
			members = append(members, m)
		}
	}
	slices.SortFunc(members, func(a, b any) int {
		return strings.Compare(a.(map[string]any)["pack_key"].(string), b.(map[string]any)["pack_key"].(string))
	})
	reporting := kind == "snapshot" || kind == "render" || kind == "composition"
	if !reporting {
		members[0].(map[string]any)["content_profile_id"] = "cartulary.reference_pack.enrichment.entra_app_ids.v1"
		members[0].(map[string]any)["content_profile_version"] = "1"
	}
	set := map[string]any{"schema_id": "cartulary.reference_pack_set.v2", "members": members}
	sum := fmt.Sprintf("%x", sha256.Sum256(encode(set)))
	set["pack_set_id"] = "rpset_" + sum
	set["pack_set_sha256"] = sum
	versions := []any{}
	anchors := []any{}
	for _, v := range members {
		m := v.(map[string]any)
		p := map[string]any{}
		pv := map[string]any{}
		for k, v := range m {
			p[k] = v
			pv[k] = v
		}
		pv["distribution_kind"] = "operator_imported"
		pv["verification_method"] = "tuf_1_0_35_offline_bundle_v1"
		pv["source_profile_id"] = "fixture.entra.admission"
		pv["source_profile_sha256"] = digest
		versions = append(versions, pv)
		p["pack_set_id"] = "rpset_" + sum
		p["authority_class"] = "registry"
		if m["pack_key"] == "enrichment.entra_app_ids" {
			p["authority_class"] = "advisory_enrichment"
		}
		p["source_profile_id"] = "fixture.entra.admission"
		p["source_profile_sha256"] = digest
		p["source_identifier"] = "Synthetic retained contract fixture"
		p["source_version"] = "1"
		p["source_as_of"] = nil
		p["source_artifacts"] = manifest["source_artifacts"]
		p["license"] = manifest["license"]
		p["verification_method"] = "packaged_release_manifest_v1"
		p["last_verified_at"] = "2026-11-07T00:00:00Z"
		p["trust_valid_until"] = nil
		p["verified_signer_key_ids"] = []any{}
		anchors = append(anchors, p)
	}
	refs := encode(map[string]any{"schema_id": "reference_pack_refs.v2", "sets": []any{set}, "versions": versions})
	manifest["content_profile_id"] = "cartulary.reference_pack.enrichment.entra_app_ids.v1"
	manifest["content_profile_version"] = "1"
	overrides := map[string][]byte{"manifest": encode(manifest), "set": encode(set), "references": refs}
	if reporting {
		anchors[0].(map[string]any)["content_profile_id"] = "cartulary.reference_pack.enrichment.entra_app_ids.v1"
		anchors[0].(map[string]any)["content_profile_version"] = "1"
		overrides["binding"] = encode(map[string]any{"pack_set_id": "rpset_" + sum, "pack_set_sha256": sum, "provenance": anchors})
	}
	seedWith(t, db, kind, refs, overrides)

	if reporting {
		exec(`INSERT INTO reference_pack_sets(pack_set_id,pack_set_sha256,canonical_set,canonical_provenance,first_operation_id) VALUES($1,$2,$3,$4,$5)`, "rpset_"+sum, sum, encode(set), encode(anchors), identity)
	}
}
