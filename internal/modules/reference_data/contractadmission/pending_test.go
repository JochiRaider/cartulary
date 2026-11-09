package contractadmission

import (
	"maps"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

func TestPendingCurrentContractAdmission_Unit(t *testing.T) {
	encode := func(v any) []byte {
		t.Helper()
		b, err := canonicaljson.Marshal(v)
		if err != nil {
			t.Fatal(err)
		}
		return b
	}
	digest := strings.Repeat("a", 64)
	operation := func() map[string]any {
		return map[string]any{
			"schema_id":             "cartulary.reference_pack_operation_input.v4",
			"pack_contract_version": "cartulary.reference_pack_contract.v3",
			"kind":                  "import", "admitted_at": "2026-11-07T00:00:00Z",
			"pack_set_id": "rpset_" + digest, "set_revision": 1,
			"configuration_sha256": digest, "clock_trusted": true, "timeout_seconds": 60,
			"container_reference": "objects/v1", "container_sha256": digest, "container_bytes": 1,
		}
	}
	portability := func() map[string]any {
		return map[string]any{
			"schema_id":             "cartulary.reference_pack_portability_input.v3",
			"pack_contract_version": "cartulary.reference_pack_contract.v3",
			"incident_id":           "00000000-0000-4000-8000-000000000001",
			"source_operation_id":   "00000000-0000-4000-8000-000000000002",
			"catalog_sha256":        digest,
			"content_manifest":      map[string]any{"schema_id": "reference_pack_content.v1", "containers": []any{}, "required_members": []any{}},
		}
	}
	bootstrap := map[string]any{"schema_id": "cartulary.reference_pack_root_bootstrap.v2", "pack_contract_version": "cartulary.reference_pack_contract.v3", "repository_id": "fixture.example", "root_version": 1, "root_sha256": digest}
	reconciliation := map[string]any{"schema_id": "cartulary.reference_pack_reconciliation.v2", "pack_contract_version": "cartulary.reference_pack_contract.v3", "application_release_id": "fixture.release.v3", "profile_claimed": true, "clock_trusted": true, "max_verification_seconds": 60, "max_container_bytes": 536870912, "max_extracted_bytes": 536870912, "max_compression_ratio": 100, "max_members": 1024}
	for name, input := range map[string]map[string]any{"operation": operation(), "portability": portability(), "completed bootstrap": bootstrap, "completed reconciliation": reconciliation} {
		t.Run(name, func(t *testing.T) {
			if err := validate("pending", encode(input)); err != nil {
				t.Fatal("current input rejected", err)
			}
			for _, schema := range []string{"unsupported", "cartulary.reference_pack_operation_input.v1", "cartulary.reference_pack_operation_input.v3", "cartulary.reference_pack_operation_input.v999", "cartulary.reference_pack_root_bootstrap.v1", "cartulary.reference_pack_reconciliation.v1", "cartulary.reference_pack_portability_input.v2"} {
				original := input["schema_id"]
				input["schema_id"] = schema
				if !IsIncompatible(validate("pending", encode(input))) {
					t.Fatal("unknown input contract admitted", schema)
				}
				input["schema_id"] = original
			}
			if !IsIncompatible(validate("pending", append(encode(input), '\n'))) {
				t.Fatal("noncanonical persisted input admitted")
			}
			for field := range input {
				missing := maps.Clone(input)
				delete(missing, field)
				if !IsIncompatible(validate("pending", encode(missing))) {
					t.Fatal("omitted field admitted", field)
				}
			}
		})
	}
	// Closed current inputs cannot carry invented nested bindings, regardless of
	// whether the embedded profile is current, obsolete, or a future identifier.
	for _, version := range []string{"1", "2", "999"} {
		input := operation()
		input["retained_binding"] = map[string]any{"pack_key": "enrichment.entra_app_ids", "content_profile_id": "cartulary.reference_pack.enrichment.entra_app_ids.v" + version, "content_profile_version": version}
		if !IsIncompatible(validate("pending", encode(input))) {
			t.Fatal("undeclared binding admitted", version)
		}
	}
	for _, change := range []func(map[string]any){
		func(m map[string]any) { m["pack_contract_version"] = "cartulary.reference_pack_contract.v1" },
		func(m map[string]any) { m["timeout_seconds"] = 0 },
		func(m map[string]any) { m["clock_trusted"] = nil },
		func(m map[string]any) { m["container_reference"] = "" },
		func(m map[string]any) { m["kind"] = "refresh" },
	} {
		input := operation()
		change(input)
		if !IsIncompatible(validate("pending", encode(input))) {
			t.Fatal("malformed current operation admitted")
		}
	}
	input := portability()
	input["content_manifest"].(map[string]any)["extra"] = true
	if !IsIncompatible(validate("pending", encode(input))) {
		t.Fatal("unknown content manifest member admitted")
	}
	if !IsIncompatible(validate("unknown", encode(operation()))) {
		t.Fatal("unknown retained artifact kind admitted")
	}
}
