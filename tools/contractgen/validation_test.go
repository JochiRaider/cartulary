package main

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
)

var fixtureExtensionProfiles = []string{
	"enterprise_authentication",
	"import",
	"incident_portability",
	"network_flow_activity",
	"reference_pack",
	"snapshot_reporting",
}

func TestDecodeContractRejectsDuplicateJSONMembers(t *testing.T) {
	for _, tc := range []struct {
		name string
		body []byte
		want error
	}{
		{"nested duplicate", []byte(`{"nested":{"value":1,"value":2}}`), canonicaljson.ErrDuplicateMember},
		{"escaped duplicate", []byte(`{"value":1,"\u0076alue":2}`), canonicaljson.ErrDuplicateMember},
		{"unsafe duplicate key", []byte(`{"private-contract-value":1,"private-contract-value":2}`), canonicaljson.ErrDuplicateMember},
		{"invalid UTF8", []byte("{\"value\":\"\xff\"}"), canonicaljson.ErrInvalidUnicode},
		{"isolated high surrogate", []byte(`{"value":"\ud800"}`), canonicaljson.ErrInvalidUnicode},
		{"isolated low surrogate", []byte(`{"value":"\udc00"}`), canonicaljson.ErrInvalidUnicode},
		{"trailing value", []byte(`{"value":"ok"} {}`), canonicaljson.ErrInvalidJSON},
		{"trailing invalid content", []byte(`{"value":"ok"} private-contract-value`), canonicaljson.ErrInvalidJSON},
	} {
		t.Run(tc.name, func(t *testing.T) {
			_, err := decodeContract(tc.body)
			if !errors.Is(err, tc.want) || strings.Contains(err.Error(), "private-contract-value") {
				t.Fatalf("common reader did not safely classify malformed bytes: %v", err)
			}
			path := filepath.Join(t.TempDir(), "contract.json")
			if err := os.WriteFile(path, tc.body, 0o600); err != nil {
				t.Fatal(err)
			}
			var document struct {
				Value string `json:"value"`
			}
			err = decodeStrictContractFile(path, &document)
			if !errors.Is(err, tc.want) || strings.Contains(err.Error(), "private-contract-value") {
				t.Fatalf("typed reader did not safely classify malformed bytes: %v", err)
			}
		})
	}
	t.Run("valid Unicode and exact number spelling", func(t *testing.T) {
		value, err := decodeContract([]byte(`{"value":"\ud83d\ude00","number":1e+02,"large":9007199254740993,"literal":"�"}`))
		if err != nil {
			t.Fatal(err)
		}
		encoded, err := canonicalizeDecoded(value)
		if err != nil || encoded != `{"large":9007199254740993,"literal":"�","number":1e+02,"value":"😀"}` {
			t.Fatalf("valid contract bytes changed: %s, %v", encoded, err)
		}
	})
	t.Run("typed shape error does not expose unknown keys", func(t *testing.T) {
		path := filepath.Join(t.TempDir(), "contract.json")
		if err := os.WriteFile(path, []byte(`{"private-contract-value":"ignored"}`), 0o600); err != nil {
			t.Fatal(err)
		}
		var document struct {
			Value string `json:"value"`
		}
		err := decodeStrictContractFile(path, &document)
		if err == nil || strings.Contains(err.Error(), "private-contract-value") {
			t.Fatalf("unsafe typed shape error: %v", err)
		}
	})
	t.Run("partial Party view reader rejects malformed ignored values", func(t *testing.T) {
		root := t.TempDir()
		path := filepath.Join(root, "contracts", "view-schemas", "cartulary.view.parties.v1.json")
		if err := os.MkdirAll(filepath.Dir(path), 0o700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(`{"view_schema_id":"cartulary.view.parties.v1","fields":[],"description":"\ud800"}`), 0o600); err != nil {
			t.Fatal(err)
		}
		err := validatePartyViewClearability(root, partyFieldRegistry{ViewSchemaID: partyViewSchemaID})
		if !errors.Is(err, canonicaljson.ErrInvalidUnicode) {
			t.Fatalf("partial reader repaired malformed bytes: %v", err)
		}
	})
}

func TestValidateRecoveryRegistryRejectsUnsortedSchemaIDs(t *testing.T) {
	registry := map[string]any{
		"$schema":            contractDraft202012Schema,
		"schema_id":          recoveryRegistrySchemaID,
		"canonicalization":   "cartulary.recovery_canonical_json.v1",
		"current_schema_ids": []any{"cartulary.z.v1", "cartulary.a.v1"},
		"retired_schema_ids": []any{"cartulary.historical.v1"},
		"limits":             map[string]any{"maximum_tables": json.Number("4096")},
		"schemas":            []any{"schema.v1.schema.json"},
		"canonical_fixtures": []any{"fixtures/schema.v1.json"},
	}
	requireErrorContains(
		t,
		validateRecoveryRegistry(registry),
		"current_schema_ids must be unique and sorted",
	)
}

func TestValidateExtensionDependencyDeclarationsRejectsOmittedAndNullArrays(t *testing.T) {
	valid := validExtensionDependencyDeclarationSet()
	dependencies := valid["dependencies"].([]any)

	for _, key := range []string{"imported_schema_ids", "imported_algorithm_ids", "imported_artifacts"} {
		t.Run("omitted "+key, func(t *testing.T) {
			row := cloneJSONMap(t, dependencies[0].(map[string]any))
			delete(row, key)
			copySet := cloneJSONMap(t, valid)
			copyRows := append([]any(nil), dependencies...)
			copyRows[0] = row
			copySet["dependencies"] = copyRows
			requireErrorContains(t, validateExtensionDependencyDeclarations(copySet), key)
		})
		t.Run("null "+key, func(t *testing.T) {
			row := cloneJSONMap(t, dependencies[0].(map[string]any))
			row[key] = nil
			copySet := cloneJSONMap(t, valid)
			copyRows := append([]any(nil), dependencies...)
			copyRows[0] = row
			copySet["dependencies"] = copyRows
			requireErrorContains(t, validateExtensionDependencyDeclarations(copySet), key)
		})
	}
}

func TestValidateExtensionOwnerFragmentRejectsCapabilities(t *testing.T) {
	fragment := map[string]any{
		"schema_id":         "cartulary.extension_owner_fragment.v3",
		"owner_fragment_id": "test.fragment.v1",
		"owner_id":          "test",
		"facts": []any{map[string]any{
			"fact_kind":     "capability",
			"profile_id":    "test",
			"capability_id": "test.capability",
		}},
	}
	requireErrorContains(t, validateExtensionOwnerFragment(fragment, "fragment.json"), "capability facts are prohibited")
}

func TestValidateExtensionConfigurationRequiresInertSchemaReference(t *testing.T) {
	contract := validExtensionConfigurationContract()
	contract["keys"].([]any)[0].(map[string]any)["inactive_policy"] = "syntax_only"
	requireErrorContains(t, validateExtensionConfigurationContract(contract, "configuration.json"), "syntax_only requires a non-null inactive_value_schema_id")
}

func TestValidateExtensionConfigurationRequiresClosedCanonicalRows(t *testing.T) {
	tests := []struct {
		name    string
		mutate  func(map[string]any)
		expects string
	}{
		{
			name: "missing value schema",
			mutate: func(contract map[string]any) {
				delete(contract["keys"].([]any)[0].(map[string]any), "value_schema_id")
			},
			expects: "missing required key value_schema_id",
		},
		{
			name: "wrong namespace schema",
			mutate: func(contract map[string]any) {
				contract["namespace_schema_id"] = "cartulary.other.configuration_namespace.v1"
			},
			expects: "namespace_schema_id must match",
		},
		{
			name: "unknown omission member",
			mutate: func(contract map[string]any) {
				contract["keys"].([]any)[0].(map[string]any)["omission_policy"].(map[string]any)["value"] = "unexpected"
			},
			expects: "unknown key value",
		},
		{
			name: "unsorted keys",
			mutate: func(contract map[string]any) {
				second := cloneJSONMap(t, contract["keys"].([]any)[0].(map[string]any))
				second["key"] = "test.alpha"
				contract["keys"] = append(contract["keys"].([]any), second)
			},
			expects: "keys must be sorted and unique",
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			contract := validExtensionConfigurationContract()
			test.mutate(contract)
			requireErrorContains(t, validateExtensionConfigurationContract(contract, "configuration.json"), test.expects)
		})
	}
}

func TestValidateExtensionConfigurationReferencesRejectsUnresolvedSchema(t *testing.T) {
	contract := validExtensionConfigurationContract()
	indexed := map[string]map[string]any{
		"profiles/test/configuration.json": contract,
		"specification/contract-definitions.json": {
			"definitions": []any{},
		},
		"specification/inactive-value-schemas.json": {
			"schemas": []any{},
		},
	}
	requireErrorContains(t, validateExtensionConfigurationReferences(indexed), "value_schema_id cartulary.test_value.v1 is unresolved")
}

func TestValidateExtensionProfileFactClosureRejectsStaleConfigurationDigest(t *testing.T) {
	root, err := repoRoot()
	if err != nil {
		t.Fatal(err)
	}
	indexed, err := readExtensionInputCatalog(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, object := range indexed {
		if object["schema_id"] != "cartulary.extension_owner_fragment.v3" {
			continue
		}
		for _, raw := range object["facts"].([]any) {
			fact := raw.(map[string]any)
			if fact["fact_kind"] == "claim_configuration" && fact["profile_id"] == "import" {
				fact["configuration_contract_sha256"] = strings.Repeat("0", 64)
			}
		}
	}
	requireErrorContains(t, validateExtensionProfileFactClosure(indexed), "stale contract digest")
}

func validExtensionConfigurationContract() map[string]any {
	return map[string]any{
		"schema_id":                    "cartulary.extension_profile_configuration_contract.v4",
		"configuration_contract_id":    "test.configuration.v1",
		"profile_id":                   "test",
		"configuration_contract_major": json.Number("1"),
		"configuration_namespace":      "test",
		"namespace_schema_id":          "cartulary.test.configuration_namespace.v1",
		"keys": []any{map[string]any{
			"key":                      "test.value",
			"value_schema_id":          "cartulary.test_value.v1",
			"omission_policy":          map[string]any{"kind": "required"},
			"inactive_policy":          "forbidden",
			"resolution_kind":          "plain",
			"diagnostic_policy":        "name_only",
			"inactive_value_schema_id": nil,
		}},
	}
}

func TestExtensionCanonicalDigestIncludesRequiredFinalLF(t *testing.T) {
	t.Run("strict authored bytes", func(t *testing.T) {
		for _, raw := range []string{"{\"v\":\"\xff\"}", `{"v":"\ud800"}`, `{"v":1,"v":2}`, "{}!"} {
			if _, err := decodeContract([]byte(raw)); err == nil {
				t.Errorf("malformed authored bytes %q accepted", raw)
			}
		}
	})
	t.Run("exact Unicode bytes", func(t *testing.T) {
		value := map[string]any{"text": "<>&/\u2028\u2029\x01", "\ue000": json.Number("9223372036854775807"), "😀": json.Number("-9223372036854775808")}
		want := "{\"text\":\"<>&/\u2028\u2029\\u0001\",\"\ue000\":9223372036854775807,\"😀\":-9223372036854775808}\n"
		digest, err := extensionCanonicalDigest(value)
		expected := sha256.Sum256([]byte(want))
		if err != nil || digest != hex.EncodeToString(expected[:]) {
			t.Fatalf("canonical digest = %s, error %v", digest, err)
		}
		artifact, err := extensionGeneratedArtifact("canonical-vector.json", value)
		if err != nil || artifact.JSON != want || artifact.SHA256 != digest {
			t.Fatalf("generated canonical artifact disagrees with the fixed vector: %v", err)
		}
	})
	value := map[string]any{"schema_id": "test.v1"}
	digest, err := extensionCanonicalDigest(value)
	if err != nil {
		t.Fatalf("digest extension artifact: %v", err)
	}
	if digest != "a03b53978fe4447e8fd7539e7778e4764686c694690529ce76b9fe2e3b51f9ae" {
		t.Fatalf("unexpected extension canonical digest %s", digest)
	}
}

func TestValidateExtensionBindingSourcesRejectsUnsortedContributions(t *testing.T) {
	bindings := make([]any, 0, len(fixtureExtensionProfiles))
	for _, profileID := range fixtureExtensionProfiles {
		bindings = append(bindings, map[string]any{
			"profile_id":                   profileID,
			"contract_major":               json.Number("1"),
			"implementation_id":            "cartulary." + profileID + ".v1",
			"state_ownership_kind":         "core_managed",
			"implemented_contribution_ids": []any{},
			"implemented_job_kinds":        []any{},
			"implemented_worker_kinds":     []any{},
			"algorithm_ids":                []any{},
			"participant_implementations":  []any{},
		})
	}
	source := map[string]any{
		"schema_id": "cartulary.extension_implementation_binding_source_set.v1",
		"bindings":  bindings,
	}
	bindings[0].(map[string]any)["implemented_contribution_ids"] = []any{"z", "a"}
	requireErrorContains(t, validateExtensionBindingSources(source), "must be sorted and unique")
}

func TestValidateExtensionParticipantContractEnforcesAggregateCeiling(t *testing.T) {
	contract := portabilitySpecializationFixture()
	contract["operations"].([]any)[0].(map[string]any)["max_input_bytes"] = json.Number("67108865")
	requireErrorContains(t, validateExtensionParticipantSpecialization(contract, "participant.json"), "max_input_bytes")
	for _, id := range []string{"cartulary.extension_participant_specialization.v3", "cartulary.extension_snapshot_reporting_participant_context.v1", "cartulary.extension_snapshot_reporting_participant_result.v1", "cartulary.extension_backup_participant_context.v1", "cartulary.extension_backup_participant_result.v1", "cartulary.extension_backup_restore_participant_context.v1", "cartulary.extension_backup_restore_participant_result.v1", "cartulary.extension_participant_contract.v1"} {
		requireErrorContains(t, validateExtensionArtifactShape(map[string]any{"schema_id": id}, "retired.json"), "retired participant contract")
		_, err := materializeExtensionSchemas(map[string]any{"schema_id": "cartulary.extension_generated_schema_source_set.v4", "schemas": []any{map[string]any{"$id": id, "$schema": "https://json-schema.org/draft/2020-12/schema", "type": "object"}}})
		requireErrorContains(t, err, "retired participant schema")
	}
}

func TestValidateExtensionParticipantSpecializationRejectsWrongSharedResult(t *testing.T) {
	contract := portabilitySpecializationFixture()
	contract["operations"].([]any)[0].(map[string]any)["result_schema_id"] = "cartulary.obsolete_result.v1"
	requireErrorContains(t, validateExtensionParticipantSpecialization(contract, "participant.json"), "result_schema_id does not match")
}

func TestDeriveExtensionArtifactsIsDeterministicAndPhaseFree(t *testing.T) {
	t.Run("owner fact duplicate normalization", checkOwnerFactNormalization)
	t.Run("current backup codec admission", func(t *testing.T) {
		for name, mutate := range map[string]func(map[string]any){
			"current":       func(map[string]any) {},
			"retired v3":    func(v map[string]any) { v["schema_id"] = "cartulary.extension_backup_binding_codec.v3" },
			"future schema": func(v map[string]any) { v["schema_id"] = "cartulary.extension_backup_binding_codec.v99" },
			"empty history": func(v map[string]any) { v["historical_restore_codecs"] = []any{} },
			"null history":  func(v map[string]any) { v["historical_restore_codecs"] = nil },
			"missing ref":   func(v map[string]any) { delete(v, "codec_contract_ref") },
			"null ref":      func(v map[string]any) { v["codec_contract_ref"] = nil },
		} {
			t.Run(name, func(t *testing.T) {
				codec := map[string]any{
					"schema_id": "cartulary.extension_backup_binding_codec.v4", "backup_codec_id": "test.codec.v1",
					"binding_id": "test.binding", "storage_kind": "postgres", "codec_contract_ref": "test.rows.v1",
					"logical_identity_algorithm_id": "test.identity.v1", "content_encoding_algorithm_id": "extension_registry_canonical_json_v1",
					"max_items": json.Number("10"), "max_entry_bytes": json.Number("128"), "max_binding_bytes": json.Number("1024"),
				}
				mutate(codec)
				err := validateExtensionArtifactShape(codec, "codec.json")
				if (err == nil) != (name == "current") {
					t.Fatalf("codec admission error = %v", err)
				}
			})
		}
	})
	root, err := repoRoot()
	if err != nil {
		t.Fatalf("resolve repository root: %v", err)
	}
	t.Run("actual producer collision admission", func(t *testing.T) { checkGeneratorCollisionEntry(t, root) })
	t.Run("complete collision matrix", func(t *testing.T) { checkGeneratorCollisionMatrix(t, root) })
	t.Run("owner input limits", func(t *testing.T) { checkGeneratorInputLimits(t, root) })
	t.Run("owner input findings", func(t *testing.T) { checkOwnerInputGenerationDiagnostics(t, root) })
	t.Run("current owner identity schemas", func(t *testing.T) { checkOwnerIdentitySchemaGeneration(t, root) })
	t.Run("single transaction limit source", func(t *testing.T) { checkTransactionLimitProjection(t, root) })
	t.Run("empty recognized catalog", func(t *testing.T) { checkEmptyExtensionGeneration(t, root) })
	t.Run("recognized profiles without implementations", func(t *testing.T) { checkUnavailableExtensionGeneration(t, root) })
	t.Run("job policy variants", func(t *testing.T) { checkExtensionJobPolicyGeneration(t, root) })
	t.Run("operational dependency references", func(t *testing.T) { checkExtensionImportReferences(t, root) })
	first, err := deriveExtensionArtifacts(root)
	if err != nil {
		t.Fatalf("derive extension artifacts: %v", err)
	}
	t.Run("shared binding result schema", func(t *testing.T) { checkBindingRestoreResultSchema(t, first) })
	t.Run("confined generator provenance", func(t *testing.T) {
		directory := t.TempDir()
		if err := os.MkdirAll(directory+"/tools/real", 0700); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(directory+"/tools/real/source.go", []byte("source bytes"), 0600); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink("real/source.go", directory+"/tools/link.go"); err != nil {
			t.Fatal(err)
		}
		if err := os.Symlink("real", directory+"/tools/alias"); err != nil {
			t.Fatal(err)
		}
		body, err := readExtensionSourceFile(directory, "tools/real/source.go")
		if err != nil || string(body) != "source bytes" {
			t.Fatalf("regular source read: %v", err)
		}
		for _, ref := range []string{"tools/../outside.go", "/absolute.go", "tools/real//source.go", "tools/real/./source.go", "tools/link.go", "tools/alias/source.go", "tools/real", "unrelated/source.go", "tools/source.txt"} {
			if _, err := readExtensionSourceFile(directory, ref); err == nil {
				t.Fatalf("invalid source path %q admitted", ref)
			}
		}
		inputs, err := readExtensionInputCatalog(root)
		if err != nil {
			t.Fatal(err)
		}
		rows, err := extensionGeneratorSources(root, inputs)
		if err != nil {
			t.Fatal(err)
		}
		seen := map[string]bool{}
		for _, row := range rows {
			ref := stringValue(row["source_ref"])
			seen[ref] = true
			raw, err := os.ReadFile(filepath.Join(root, filepath.FromSlash(ref)))
			if err != nil {
				t.Fatal(err)
			}
			sum := sha256.Sum256(raw)
			if row["source_sha256"] != hex.EncodeToString(sum[:]) {
				t.Fatal("provenance digest does not match source bytes")
			}
		}
		for _, ref := range []string{"contracts/extensions/index.json", "contracts/index.json", "go.mod", "go.sum", "tools/toolchain_pins.json", "tools/contractgen/families.go", "tools/contractgen/extensions_transaction.go"} {
			if !seen[ref] {
				t.Fatalf("unbound generator control input %s", ref)
			}
		}
	})
	t.Run("runtime protocol integrity", func(t *testing.T) {
		inputs, err := readExtensionInputCatalog(root)
		if err != nil {
			t.Fatal(err)
		}
		digest, err := extensionCanonicalDigest(inputs["specification/shared-protocols.json"])
		if err != nil {
			t.Fatal(err)
		}
		found := false
		for _, artifact := range first {
			if artifact.Path != generatedExtensionPrefix+"registry-integrity.json" {
				continue
			}
			value, err := decodeContract([]byte(artifact.JSON))
			if err != nil {
				t.Fatal(err)
			}
			rows, err := objectArray(value.(map[string]any)["supporting_contract_artifact_digests"], "supporting")
			if err != nil {
				t.Fatal(err)
			}
			for _, row := range rows {
				if row["artifact_id"] == "specification.shared-protocols" {
					found = row["artifact_sha256"] == digest && row["schema_id"] == "cartulary.extension_shared_protocol_contract_set.v5"
				}
			}
		}
		if !found || !extensionRuntimeArtifactPath("contracts/extensions/specification/shared-protocols.json") {
			t.Fatal("runtime protocol is not digest-bound")
		}
	})

	t.Run("packaged owner fragment parity", func(t *testing.T) {
		artifacts, err := collectArtifacts(root, "extensions")
		if err != nil {
			t.Fatal(err)
		}
		fragments := map[string]artifact{}
		var ownerInput map[string]any
		for _, artifact := range artifacts {
			value, err := decodeContract([]byte(artifact.JSON))
			if err != nil {
				t.Fatal(err)
			}
			object := value.(map[string]any)
			if object["schema_id"] == "cartulary.extension_owner_fragment.v3" {
				fragments[stringValue(object["owner_fragment_id"])] = artifact
			}
			if object["schema_id"] == "cartulary.extension_owner_input_registry.v2" {
				ownerInput = object
			}
		}
		rows, err := objectArray(ownerInput["owner_fragments"], "fragments")
		if err != nil {
			t.Fatal(err)
		}
		if len(rows) != len(fragments) {
			t.Fatal("packaged fragment set mismatch")
		}
		for _, row := range rows {
			digest, err := extensionCanonicalDigest(row)
			if err != nil {
				t.Fatal(err)
			}
			if fragments[stringValue(row["owner_fragment_id"])].SHA256 != digest {
				t.Fatal("embedded and packaged owner fragment bytes differ")
			}
			if !strings.HasPrefix(stringValue(row["owner_fragment_id"]), stringValue(row["owner_id"])+".") {
				t.Fatal("fragment ID does not use its owner prefix")
			}
		}
	})
	t.Run("logical codec resolution", func(t *testing.T) {
		for name, mutate := range map[string]func(map[string]map[string]any){
			"missing definition": func(inputs map[string]map[string]any) {
				delete(inputs, "profiles/network_flow_activity/postgres-rows.json")
			},
			"unresolved reference": func(inputs map[string]map[string]any) {
				inputs["profiles/network_flow_activity/tables-codec.json"]["codec_contract_ref"] = "cartulary.unknown.rows.v1"
			},
			"unsupported encoding": func(inputs map[string]map[string]any) {
				inputs["profiles/network_flow_activity/tables-codec.json"]["content_encoding_algorithm_id"] = "database_native_json"
			},
			"missing mapping": func(inputs map[string]map[string]any) { delete(inputs, "build/network-flow-tables-mapping.json") },
			"ambiguous definition": func(inputs map[string]map[string]any) {
				inputs["duplicate.json"] = cloneObject(inputs["profiles/network_flow_activity/postgres-rows.json"])
			},
			"unbound mapping": func(inputs map[string]map[string]any) {
				extra := cloneObject(inputs["build/network-flow-tables-mapping.json"])
				extra["binding_id"] = "network_flow_activity.unused"
				inputs["extra.json"] = extra
			},
			"nondeclarative mapping": func(inputs map[string]map[string]any) {
				inputs["build/network-flow-tables-mapping.json"]["table_name"] = "table; SELECT secret"
			},
		} {
			t.Run(name, func(t *testing.T) {
				inputs, err := readExtensionInputCatalog(root)
				if err != nil {
					t.Fatal(err)
				}
				mutate(inputs)
				bindings, err := objectArray(inputs["build/network-flow-state-binding.json"]["bindings"], "bindings")
				if err != nil {
					t.Fatal(err)
				}
				if _, _, err := extensionLogicalBackupArtifacts("network_flow_activity", bindings, inputs); err == nil {
					t.Fatal("invalid logical codec resolution accepted")
				}
			})
		}
	})
	second, err := deriveExtensionArtifacts(root)
	if err != nil {
		t.Fatalf("derive extension artifacts again: %v", err)
	}
	if len(first) == 0 || len(first) != len(second) {
		t.Fatalf("unexpected derived artifact counts %d and %d", len(first), len(second))
	}
	seen := map[string]artifact{}
	for index := range first {
		if first[index] != second[index] {
			t.Fatalf("derived artifact %d is not byte-stable", index)
		}
		if strings.Contains(first[index].Path, "phase2") || strings.Contains(first[index].JSON, "cartulary.extensions.phase2.v1") {
			t.Fatalf("derived artifact retains phase-shaped identity: %s", first[index].Path)
		}
		seen[first[index].Path] = first[index]
	}
	for _, requiredPath := range []string{
		generatedExtensionPrefix + "profile-registry.json",
		generatedExtensionPrefix + "registry-integrity.json",
		generatedExtensionPrefix + "validation-condition-registry.json",
		generatedExtensionPrefix + "implementation-bindings/network_flow_activity.json",
		generatedExtensionPrefix + "job-contracts/import/import.apply_v1.json",
		generatedExtensionPrefix + "job-contracts/import/import.discovery_v1.json",
		generatedExtensionPrefix + "job-contracts/incident_portability/incident_portability.export_v2.json",
		generatedExtensionPrefix + "job-contracts/incident_portability/incident_portability.import_v2.json",
		generatedExtensionPrefix + "job-contracts/reference_pack/reference_pack.import_v2.json",
		generatedExtensionPrefix + "job-contracts/reference_pack/reference_pack.refresh_v2.json",
		generatedExtensionPrefix + "job-contracts/reference_pack/reference_pack.reverify_v2.json",
		generatedExtensionPrefix + "job-contracts/snapshot_reporting/snapshot_reporting.composition_preview_v4.json",
		generatedExtensionPrefix + "job-contracts/snapshot_reporting/snapshot_reporting.release_create_v4.json",
		generatedExtensionPrefix + "job-contracts/snapshot_reporting/snapshot_reporting.snapshot_create_v4.json",
	} {
		if _, exists := seen[requiredPath]; !exists {
			t.Fatalf("missing generated artifact %s", requiredPath)
		}
	}
	registry := seen[generatedExtensionPrefix+"profile-registry.json"]
	if !strings.HasSuffix(registry.JSON, "\n") || strings.HasSuffix(registry.JSON, "\n\n") {
		t.Fatalf("registry does not use exactly one final LF")
	}

	var stateRegistry map[string]any
	if err := json.Unmarshal([]byte(seen[generatedExtensionPrefix+"state-registry.json"].JSON), &stateRegistry); err != nil {
		t.Fatalf("decode generated state registry: %v", err)
	}
	profiles, ok := stateRegistry["profiles"].([]any)
	if !ok || len(profiles) != 1 {
		t.Fatalf("generated state profiles = %#v", stateRegistry["profiles"])
	}
	networkFlow, ok := profiles[0].(map[string]any)
	if !ok ||
		networkFlow["profile_id"] != "network_flow_activity" ||
		networkFlow["initialization_kind"] != "empty" ||
		networkFlow["initialization_algorithm_id"] != nil ||
		networkFlow["initialization_algorithm_definition_sha256"] != nil ||
		networkFlow["initialization_definition_sha256"] == "" ||
		networkFlow["implementation_binding_sha256"] == "" {
		t.Fatalf("generated Network Flow state projection = %#v", profiles[0])
	}
}

func validExtensionDependencyDeclarationSet() map[string]any {
	rows := make([]any, 0, len(requiredExtensionDependencies))
	for _, dependencyID := range requiredExtensionDependencies {
		rows = append(rows, map[string]any{
			"dependency_id":          dependencyID,
			"imported_schema_ids":    []any{},
			"imported_algorithm_ids": []any{},
			"imported_artifacts":     []any{},
		})
	}
	return map[string]any{
		"schema_id":    "cartulary.extension_dependency_declaration_set.v3",
		"dependencies": rows,
	}
}

func cloneJSONMap(t *testing.T, value map[string]any) map[string]any {
	t.Helper()
	encoded, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal clone: %v", err)
	}
	var cloned map[string]any
	decoder := json.NewDecoder(strings.NewReader(string(encoded)))
	decoder.UseNumber()
	if err := decoder.Decode(&cloned); err != nil {
		t.Fatalf("decode clone: %v", err)
	}
	return cloned
}

func TestValidateViewSchemaRejectsUnknownKeys(t *testing.T) {
	schema := validViewSchema("cartulary.view.test.v1")
	schema["legacy_key"] = true

	err := validateViewSchemaShape(schema, "cartulary.view.test.v1.json")
	requireErrorContains(t, err, "unknown key legacy_key")
}

func TestValidateViewSchemaRejectsMissingRequiredFields(t *testing.T) {
	schema := validViewSchema("cartulary.view.test.v1")
	delete(schema, "inline_create")

	err := validateViewSchemaShape(schema, "cartulary.view.test.v1.json")
	requireErrorContains(t, err, "inline_create must be an object")
}

func TestValidateViewSchemaRejectsInvalidCreateOnlyOverride(t *testing.T) {
	t.Run("false override", func(t *testing.T) {
		schema := validViewSchema("cartulary.view.test.v1")
		field := schema["fields"].([]any)[0].(map[string]any)
		field["create_writable"] = false

		err := validateViewSchemaShape(schema, "cartulary.view.test.v1.json")
		requireErrorContains(t, err, "create-only override and may only be declared as true")
	})

	t.Run("ordinary writable field", func(t *testing.T) {
		schema := validViewSchema("cartulary.view.test.v1")
		field := schema["fields"].([]any)[0].(map[string]any)
		field["create_writable"] = true

		err := validateViewSchemaShape(schema, "cartulary.view.test.v1.json")
		requireErrorContains(t, err, "requires a create-only non-read-only field")
	})

	t.Run("read-only field", func(t *testing.T) {
		schema := validViewSchema("cartulary.view.test.v1")
		field := schema["fields"].([]any)[0].(map[string]any)
		field["writable"] = false
		field["write_kind"] = "read_only"
		field["grid_editable"] = false
		field["create_writable"] = true
		delete(field, "write_target")

		err := validateViewSchemaShape(schema, "cartulary.view.test.v1.json")
		requireErrorContains(t, err, "requires a create-only non-read-only field")
	})
}

func TestValidateErrorRegistryRejectsDuplicateCodes(t *testing.T) {
	registry := validErrorRegistry()
	registry["errors"] = []any{
		validErrorEntry("cartulary.test.duplicate", "409"),
		validErrorEntry("cartulary.test.duplicate", "400"),
	}

	err := validateErrorRegistry(registry)
	requireErrorContains(t, err, "duplicate error code cartulary.test.duplicate")
}

func TestValidateErrorRegistryRejectsInvalidReasonRegistries(t *testing.T) {
	t.Run("unknown error code", func(t *testing.T) {
		registry := validErrorRegistry()
		registry["reason_registries"] = []any{
			validReasonRegistry("cartulary.test.missing", "reason"),
		}

		err := validateErrorRegistry(registry)
		requireErrorContains(t, err, "references unknown error code cartulary.test.missing")
	})

	t.Run("duplicate reason code", func(t *testing.T) {
		registry := validErrorRegistry()
		registry["reason_registries"] = []any{
			map[string]any{
				"error_code": "cartulary.test.conflict",
				"reason_codes": []any{
					validReasonCode("duplicate"),
					validReasonCode("duplicate"),
				},
			},
		}

		err := validateErrorRegistry(registry)
		requireErrorContains(t, err, "contains duplicate reason code duplicate")
	})
}

func TestValidateViewSchemaRejectsDuplicateFieldKeys(t *testing.T) {
	schema := validViewSchema("cartulary.view.test.v1")
	schema["fields"] = []any{
		validViewField("name"),
		validViewField("name"),
	}

	err := validateViewSchemaShape(schema, "cartulary.view.test.v1.json")
	requireErrorContains(t, err, "duplicate field_key name")
}

func TestValidateViewSchemaIndexRejectsStaleArtifactPaths(t *testing.T) {
	root := t.TempDir()
	ownerRoot, err := repoRoot()
	if err != nil {
		t.Fatal(err)
	}
	registryBytes, err := os.ReadFile(filepath.Join(ownerRoot, "contracts", "view-inspector", "index.json"))
	if err != nil {
		t.Fatal(err)
	}
	registry, err := decodeContract(registryBytes)
	if err != nil {
		t.Fatal(err)
	}
	writeJSONFile(t, filepath.Join(root, "contracts", "view-inspector", "index.json"), registry)

	writeJSONFile(t, filepath.Join(root, "contracts", "view-schemas", "index.json"), map[string]any{
		"$schema":     contractDraft202012Schema,
		"registry_id": "cartulary.view_schemas.v1",
		"note":        "test index",
		"view_schemas": []any{
			map[string]any{
				"view_schema_id":               "cartulary.view.test.v1",
				"title":                        "Test View",
				"surface_kind":                 "built_in_sheet",
				"surface_status":               "required_built_in_sheet",
				"source_record_types":          []any{"note"},
				"required_reference_pack_keys": []any{},
				"artifact_path":                "contracts/view-schemas/missing.json",
			},
		},
	})
	writeJSONFile(
		t,
		filepath.Join(root, "contracts", "view-schemas", "cartulary.view.test.v1.json"),
		validViewSchema("cartulary.view.test.v1"),
	)

	err = validateContractFamily(root, "view-schemas")
	requireErrorContains(t, err, "references missing artifact contracts/view-schemas/missing.json")
}

func TestLoadFamiliesUsesActiveRegistryEntries(t *testing.T) {
	root := t.TempDir()
	for _, dir := range []string{"contracts/openapi", "contracts/network-flow"} {
		if err := os.MkdirAll(filepath.Join(root, dir), 0o755); err != nil {
			t.Fatalf("create contract dir: %v", err)
		}
	}
	writeJSONFile(t, filepath.Join(root, "contracts", "index.json"), map[string]any{
		"$schema":     contractDraft202012Schema,
		"schema_id":   contractFamilyRegistrySchemaID,
		"registry_id": contractFamilyRegistryID,
		"note":        "test registry",
		"families": []any{
			validContractFamilyEntry("openapi", "contracts/openapi", "active", "OpenAPIArtifacts", 0, nil),
			validContractFamilyEntry("network-flow", "contracts/network-flow", "planned", "NetworkFlowArtifacts", 1, []any{"NFA-GEN-002"}),
		},
	})

	families, err := loadFamilies(root)
	if err != nil {
		t.Fatalf("load families: %v", err)
	}
	if len(families) != 1 {
		t.Fatalf("expected one active family, got %d", len(families))
	}
	if families[0].Dir != "openapi" || families[0].GoName != "OpenAPIArtifacts" {
		t.Fatalf("unexpected active family: %#v", families[0])
	}
}

func TestLoadFamiliesRejectsUnsafeOrAmbiguousRegistry(t *testing.T) {
	t.Run("planned family requires activation dependency", func(t *testing.T) {
		root := t.TempDir()
		for _, dir := range []string{"contracts/openapi", "contracts/network-flow"} {
			if err := os.MkdirAll(filepath.Join(root, dir), 0o755); err != nil {
				t.Fatalf("create contract dir: %v", err)
			}
		}
		writeJSONFile(t, filepath.Join(root, "contracts", "index.json"), map[string]any{
			"$schema":     contractDraft202012Schema,
			"schema_id":   contractFamilyRegistrySchemaID,
			"registry_id": contractFamilyRegistryID,
			"note":        "test registry",
			"families": []any{
				validContractFamilyEntry("openapi", "contracts/openapi", "active", "OpenAPIArtifacts", 0, nil),
				validContractFamilyEntry("network-flow", "contracts/network-flow", "planned", "NetworkFlowArtifacts", 1, nil),
			},
		})

		_, err := loadFamilies(root)
		requireErrorContains(t, err, "activation_dependency_ids must not be empty for planned families")
	})

	t.Run("duplicate generated name is rejected", func(t *testing.T) {
		root := t.TempDir()
		for _, dir := range []string{"contracts/openapi", "contracts/ws"} {
			if err := os.MkdirAll(filepath.Join(root, dir), 0o755); err != nil {
				t.Fatalf("create contract dir: %v", err)
			}
		}
		writeJSONFile(t, filepath.Join(root, "contracts", "index.json"), map[string]any{
			"$schema":     contractDraft202012Schema,
			"schema_id":   contractFamilyRegistrySchemaID,
			"registry_id": contractFamilyRegistryID,
			"note":        "test registry",
			"families": []any{
				validContractFamilyEntry("openapi", "contracts/openapi", "active", "OpenAPIArtifacts", 0, nil),
				validContractFamilyEntry("ws", "contracts/ws", "active", "OpenAPIArtifacts", 1, nil),
			},
		})

		_, err := loadFamilies(root)
		requireErrorContains(t, err, "go_name duplicates family openapi")
	})

	t.Run("escaped contract root is rejected", func(t *testing.T) {
		root := t.TempDir()
		if err := os.MkdirAll(filepath.Join(root, "contracts", "openapi"), 0o755); err != nil {
			t.Fatalf("create contract dir: %v", err)
		}
		writeJSONFile(t, filepath.Join(root, "contracts", "index.json"), map[string]any{
			"$schema":     contractDraft202012Schema,
			"schema_id":   contractFamilyRegistrySchemaID,
			"registry_id": contractFamilyRegistryID,
			"note":        "test registry",
			"families": []any{
				validContractFamilyEntry("openapi", "contracts/../openapi", "active", "OpenAPIArtifacts", 0, nil),
			},
		})

		_, err := loadFamilies(root)
		requireErrorContains(t, err, "contract_root escapes contracts/")
	})
}

func TestGenerationPlanRejectsUnsafeOrAmbiguousTypeScriptProjections(t *testing.T) {
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "contracts", "errors"), 0o755); err != nil {
		t.Fatalf("create contract dir: %v", err)
	}

	tests := []struct {
		name       string
		mutate     func(*contractFamilyRegistry)
		wantDetail string
	}{
		{
			name: "artifact collections are unsupported",
			mutate: func(registry *contractFamilyRegistry) {
				registry.Families[0].TypeScriptProjections[0].ProjectionKind = "artifact_collection"
			},
			wantDetail: "projection_kind must be json_value",
		},
		{
			name: "projection output must be declared",
			mutate: func(registry *contractFamilyRegistry) {
				registry.Families[0].TypeScriptProjections[0].OutputPath = "packages/protocol-ts/src/generated/undeclared.ts"
			},
			wantDetail: "must be declared in generated_outputs",
		},
		{
			name: "selection cannot escape family root",
			mutate: func(registry *contractFamilyRegistry) {
				registry.Families[0].TypeScriptProjections[0].ArtifactPath = "contracts/ws/index.json"
			},
			wantDetail: "must stay within contracts/errors",
		},
		{
			name: "json value selection is exact",
			mutate: func(registry *contractFamilyRegistry) {
				registry.Families[0].TypeScriptProjections[0].ArtifactPath = "contracts/errors/"
			},
			wantDetail: "must select one exact artifact",
		},
		{
			name: "projection identifiers are globally unique",
			mutate: func(registry *contractFamilyRegistry) {
				registry.Families[0].GeneratedOutputs = append(registry.Families[0].GeneratedOutputs, "packages/protocol-ts/src/generated/second.ts")
				registry.Families[0].TypeScriptProjections = append(registry.Families[0].TypeScriptProjections, typeScriptProjectionEntry{
					ProjectionKind: jsonValueProjection,
					OutputPath:     "packages/protocol-ts/src/generated/second.ts",
					Identifier:     "errorRegistry",
					ArtifactPath:   "contracts/errors/index.json",
				})
			},
			wantDetail: "duplicates projection identifier",
		},
		{
			name: "projection outputs are globally unique",
			mutate: func(registry *contractFamilyRegistry) {
				registry.Families[0].TypeScriptProjections = append(registry.Families[0].TypeScriptProjections, typeScriptProjectionEntry{
					ProjectionKind: jsonValueProjection,
					OutputPath:     "packages/protocol-ts/src/generated/error-registry.ts",
					Identifier:     "secondRegistry",
					ArtifactPath:   "contracts/errors/index.json",
				})
			},
			wantDetail: "duplicates projection owned by family errors",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			registry := validProjectionRegistry()
			tc.mutate(&registry)
			_, err := generationPlanFromRegistry(root, registry)
			requireErrorContains(t, err, tc.wantDetail)
		})
	}
}

func TestTypeScriptProjectionSourcesAreExactAndDeeplyReadonly(t *testing.T) {
	current := family{
		Dir:          "errors",
		ContractRoot: "contracts/errors",
		Artifacts: []artifact{
			{Path: "contracts/errors/index.json", JSON: `{"schema_id":"test.errors.v1","values":["one"]}`, SHA256: "abc"},
			{Path: "contracts/errors/index.json.backup", JSON: `{"schema_id":"wrong.v1"}`, SHA256: "def"},
		},
	}
	projection := typeScriptProjection{
		Kind:         jsonValueProjection,
		OutputPath:   "packages/protocol-ts/src/generated/error-registry.ts",
		Identifier:   "errorRegistry",
		ArtifactPath: "contracts/errors/index.json",
	}
	source, err := jsonValueSource(current, projection)
	if err != nil {
		t.Fatalf("generate json_value source: %v", err)
	}
	text := string(source)
	for _, required := range []string{"function deepFreeze", `"test.errors.v1"`, "as const", "export const errorRegistry = deepFreeze(value)"} {
		if !strings.Contains(text, required) {
			t.Fatalf("json_value source missing %q:\n%s", required, text)
		}
	}
	if strings.Contains(text, "wrong.v1") {
		t.Fatalf("json_value source widened exact selection:\n%s", text)
	}
}

func validProjectionRegistry() contractFamilyRegistry {
	return contractFamilyRegistry{
		Schema:     contractDraft202012Schema,
		SchemaID:   contractFamilyRegistrySchemaID,
		RegistryID: contractFamilyRegistryID,
		Note:       "test registry",
		Families: []contractFamilyEntry{
			{
				FamilyID:         "errors",
				ContractRoot:     "contracts/errors",
				GenerationStatus: "active",
				GoName:           "ErrorArtifacts",
				OutputOrder:      0,
				GeneratedOutputs: []string{
					"internal/gen/contracterrors/artifacts_gen.go",
					"packages/protocol-ts/src/generated/error-registry.ts",
				},
				TypeScriptProjections: []typeScriptProjectionEntry{
					{
						ProjectionKind: jsonValueProjection,
						OutputPath:     "packages/protocol-ts/src/generated/error-registry.ts",
						Identifier:     "errorRegistry",
						ArtifactPath:   "contracts/errors/index.json",
					},
				},
				Description: "test family",
			},
		},
	}
}

func TestValidateViewSchemaRejectsInvalidFieldReferences(t *testing.T) {
	schema := validViewSchema("cartulary.view.test.v1")
	schema["sort_fields"] = []any{"missing_field"}

	err := validateViewSchemaShape(schema, "cartulary.view.test.v1.json")
	requireErrorContains(t, err, "sort_fields references unknown field missing_field")
}

func TestValidateViewSchemaRejectsMismatchedEnumMetadata(t *testing.T) {
	t.Run("enum values require enum read kind", func(t *testing.T) {
		schema := validViewSchema("cartulary.view.test.v1")
		field := validViewField("name")
		field["read_kind"] = "text"
		field["enum_values"] = []any{"open", "closed"}
		schema["fields"] = []any{field}

		err := validateViewSchemaShape(schema, "cartulary.view.test.v1.json")
		requireErrorContains(t, err, "enum_values requires read_kind=enum")
	})

	t.Run("enum read kind requires values", func(t *testing.T) {
		schema := validViewSchema("cartulary.view.test.v1")
		field := validViewField("name")
		field["read_kind"] = "enum"
		field["enum_values"] = nil
		schema["fields"] = []any{field}

		err := validateViewSchemaShape(schema, "cartulary.view.test.v1.json")
		requireErrorContains(t, err, "read_kind=enum requires enum_values")
	})
}

func TestValidateViewSchemaAcceptsCanonicalArtifactSourceFilter(t *testing.T) {
	schema := validViewSchema("cartulary.view.test.v1")
	schema["base_projection"] = "artifact_grid_projection"
	schema["canonical_source_filter"] = map[string]any{
		"kind":  "artifact_type",
		"field": "artifact_type",
		"value": "test_artifact",
	}

	if err := validateViewSchemaShape(schema, "cartulary.view.test.v1.json"); err != nil {
		t.Fatalf("expected canonical artifact source filter to validate: %v", err)
	}
}

func TestValidateViewSchemaRejectsInvalidCanonicalSourceFilter(t *testing.T) {
	tests := []struct {
		name       string
		filter     any
		wantDetail string
	}{
		{
			name:       "not object",
			filter:     "artifact_type=test",
			wantDetail: "canonical_source_filter must be an object",
		},
		{
			name: "unknown key",
			filter: map[string]any{
				"kind":   "artifact_type",
				"field":  "artifact_type",
				"value":  "test_artifact",
				"legacy": true,
			},
			wantDetail: "unknown key legacy",
		},
		{
			name: "unsupported kind",
			filter: map[string]any{
				"kind":  "record_type",
				"field": "artifact_type",
				"value": "test_artifact",
			},
			wantDetail: "kind must be one of artifact_type",
		},
		{
			name: "unsupported field",
			filter: map[string]any{
				"kind":  "artifact_type",
				"field": "kind",
				"value": "test_artifact",
			},
			wantDetail: "field must be one of artifact_type",
		},
		{
			name: "missing value",
			filter: map[string]any{
				"kind":  "artifact_type",
				"field": "artifact_type",
			},
			wantDetail: "value is required",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			schema := validViewSchema("cartulary.view.test.v1")
			schema["canonical_source_filter"] = tc.filter

			err := validateViewSchemaShape(schema, "cartulary.view.test.v1.json")
			requireErrorContains(t, err, tc.wantDetail)
		})
	}
}

func TestValidateViewSchemaRejectsInvalidInspectorConfig(t *testing.T) {
	t.Run("missing config", func(t *testing.T) {
		schema := validViewSchema("cartulary.view.test.v1")
		delete(schema, "inspector_config")

		err := validateViewSchemaShape(schema, "cartulary.view.test.v1.json")
		requireErrorContains(t, err, "inspector_config must be an object")
	})

	t.Run("mismatched view schema", func(t *testing.T) {
		schema := validViewSchema("cartulary.view.test.v1")
		config := schema["inspector_config"].(map[string]any)
		config["view_schema_id"] = "cartulary.view.other.v1"

		err := validateViewSchemaShape(schema, "cartulary.view.test.v1.json")
		requireErrorContains(t, err, "view_schema_id must match containing view_schema_id")
	})

	t.Run("unknown panel", func(t *testing.T) {
		schema := validViewSchema("cartulary.view.test.v1")
		config := schema["inspector_config"].(map[string]any)
		config["panels"] = []any{map[string]any{"panel_id": "legacy", "label": "Legacy"}}
		config["feature_groups"].([]any)[0].(map[string]any)["panel_id"] = "legacy"

		err := validateViewSchemaWithInspectorRegistry(t, schema, []string{"details.read"})
		requireErrorContains(t, err, "panel_id references unknown panel legacy")
	})

	t.Run("duplicate panel", func(t *testing.T) {
		schema := validViewSchema("cartulary.view.test.v1")
		config := schema["inspector_config"].(map[string]any)
		config["panels"] = []any{
			map[string]any{"panel_id": "details", "label": "Details"},
			map[string]any{"panel_id": "details", "label": "Details Again"},
		}

		err := validateViewSchemaShape(schema, "cartulary.view.test.v1.json")
		requireErrorContains(t, err, "duplicate panel_id details")
	})

	t.Run("duplicate feature key", func(t *testing.T) {
		schema := validViewSchema("cartulary.view.test.v1")
		config := schema["inspector_config"].(map[string]any)
		group := validInspectorFeatureGroup("details.read")
		config["feature_groups"] = []any{group, validInspectorFeatureGroup("details.read")}

		err := validateViewSchemaShape(schema, "cartulary.view.test.v1.json")
		requireErrorContains(t, err, "duplicate feature_group_key details.read")
	})

	t.Run("feature group bound", func(t *testing.T) {
		schema := validViewSchema("cartulary.view.test.v1")
		config := schema["inspector_config"].(map[string]any)
		groups := make([]any, 0, 65)
		for index := 0; index < 65; index++ {
			groups = append(groups, validInspectorFeatureGroup("details.read_"+string(rune('a'+index%26))+"_"+string(rune('a'+index/26))))
		}
		config["feature_groups"] = groups

		err := validateViewSchemaShape(schema, "cartulary.view.test.v1.json")
		requireErrorContains(t, err, "feature_groups must contain at most 64 entries")
	})

	t.Run("unknown route kind", func(t *testing.T) {
		schema := validViewSchema("cartulary.view.test.v1")
		group := schema["inspector_config"].(map[string]any)["feature_groups"].([]any)[0].(map[string]any)
		group["route_binding"] = map[string]any{"kind": "legacy_route", "owner": "current_row_projection"}

		err := validateViewSchemaWithInspectorRegistry(t, schema, []string{"details.read"})
		requireErrorContains(t, err, "route_binding.kind references unknown kind legacy_route")
	})

	t.Run("unknown route owner", func(t *testing.T) {
		schema := validViewSchema("cartulary.view.test.v1")
		group := schema["inspector_config"].(map[string]any)["feature_groups"].([]any)[0].(map[string]any)
		group["route_binding"] = map[string]any{"kind": "panel_read", "owner": "legacy_owner"}

		err := validateViewSchemaWithInspectorRegistry(t, schema, []string{"details.read"})
		requireErrorContains(t, err, "route_binding.owner references unknown owner legacy_owner")
	})

	t.Run("unknown disabled condition", func(t *testing.T) {
		schema := validViewSchema("cartulary.view.test.v1")
		group := schema["inspector_config"].(map[string]any)["feature_groups"].([]any)[0].(map[string]any)
		group["disabled_when"] = []any{"stale_legacy_state"}

		err := validateViewSchemaWithInspectorRegistry(t, schema, []string{"details.read"})
		requireErrorContains(t, err, "disabled_when references unknown field stale_legacy_state")
	})

	t.Run("unknown result behaviors", func(t *testing.T) {
		for field, want := range map[string]string{
			"success_result_behavior": "success_result_behavior references unknown behavior legacy_result",
			"failure_result_behavior": "failure_result_behavior references unknown behavior legacy_result",
		} {
			schema := validViewSchema("cartulary.view.test.v1")
			group := schema["inspector_config"].(map[string]any)["feature_groups"].([]any)[0].(map[string]any)
			group[field] = "legacy_result"

			err := validateViewSchemaWithInspectorRegistry(t, schema, []string{"details.read"})
			requireErrorContains(t, err, want)
		}
	})

	t.Run("invalid feature key grammar", func(t *testing.T) {
		schema := validViewSchema("cartulary.view.test.v1")
		group := schema["inspector_config"].(map[string]any)["feature_groups"].([]any)[0].(map[string]any)
		group["feature_group_key"] = "Details Read"

		err := validateViewSchemaShape(schema, "cartulary.view.test.v1.json")
		requireErrorContains(t, err, "feature_group_key must be ASCII lower snake or dotted key")
	})

	t.Run("current profile registry completeness", func(t *testing.T) {
		schema := validViewSchema("cartulary.view.timeline.v2")
		config := schema["inspector_config"].(map[string]any)
		config["view_schema_id"] = "cartulary.view.timeline.v2"

		registry := loadTestViewInspectorRegistry(t)
		expected := registry.ViewFeatureKeys["cartulary.view.timeline.v2"]
		err := validateInspectorConfigAgainstRegistry(config, "cartulary.view.timeline.v2", expected, registry)
		requireErrorContains(t, err, "feature_groups must contain exactly 27 ordered feature groups, got 1")

		groups := make([]any, 0, len(expected))
		for _, key := range expected {
			groups = append(groups, validInspectorFeatureGroup(key))
		}
		groups[0].(map[string]any)["feature_group_key"] = "details.future"
		config["feature_groups"] = groups

		err = validateInspectorConfigAgainstRegistry(config, "cartulary.view.timeline.v2", expected, registry)
		requireErrorContains(t, err, "feature_group_key must be details.read, got details.future")
	})
}

func loadTestViewInspectorRegistry(t *testing.T) viewInspectorRegistry {
	t.Helper()
	root, err := repoRoot()
	if err != nil {
		t.Fatalf("find repository root: %v", err)
	}
	registry, err := loadViewInspectorRegistry(root)
	if err != nil {
		t.Fatalf("load view Inspector registry: %v", err)
	}
	return registry
}

func validateViewSchemaWithInspectorRegistry(
	t *testing.T,
	schema map[string]any,
	expected []string,
) error {
	t.Helper()
	if err := validateViewSchemaShape(schema, "cartulary.view.test.v1.json"); err != nil {
		return err
	}
	config := schema["inspector_config"].(map[string]any)
	return validateInspectorConfigAgainstRegistry(
		config,
		"cartulary.view.test.v1",
		expected,
		loadTestViewInspectorRegistry(t),
	)
}

func validViewSchema(id string) map[string]any {
	return map[string]any{
		"$schema":                      contractDraft202012Schema,
		"schema_id":                    "cartulary.view_schema_source.v1",
		"view_schema_id":               id,
		"title":                        "Test View",
		"surface_kind":                 "built_in_sheet",
		"source_record_types":          []any{"note"},
		"technical_fields":             []any{"record_id"},
		"required_reference_pack_keys": []any{},
		"default_visible_fields":       []any{"name"},
		"default_hidden_fields":        []any{"record_id"},
		"default_sort": []any{
			map[string]any{
				"field_key": "record_id",
				"direction": "asc",
			},
		},
		"sort_fields":                 []any{"name"},
		"sort_null_order":             "last",
		"filter_fields":               []any{"name"},
		"synthetic_filter_predicates": []any{},
		"grouping_fields":             []any{},
		"create_capable":              true,
		"create_inputs":               []any{},
		"inline_create":               map[string]any{"permits_zero_field_create": false},
		"inspector_config":            validInspectorConfig(id),
		"fields":                      []any{validViewField("name")},
	}
}

func validInspectorConfig(id string) map[string]any {
	return map[string]any{
		"inspector_config_schema_id":   "cartulary.inspector_config.v1",
		"view_schema_id":               id,
		"default_open":                 false,
		"subject_binding":              map[string]any{"kind": "selected_record"},
		"no_row_state":                 "no_row_selected",
		"unsupported_feature_behavior": "omit_feature",
		"panels": []any{
			map[string]any{"panel_id": "details", "label": "Details"},
		},
		"feature_groups": []any{validInspectorFeatureGroup("details.read")},
	}
}

func validInspectorFeatureGroup(key string) map[string]any {
	return map[string]any{
		"feature_group_key":       key,
		"panel_id":                "details",
		"label":                   "Details panel",
		"minimum_incident_role":   nil,
		"mutates":                 false,
		"requires_confirmation":   false,
		"route_binding":           map[string]any{"kind": "panel_read", "owner": "current_row_projection"},
		"seed_bindings":           []any{},
		"disabled_when":           []any{"no_row_selected"},
		"success_result_behavior": "preserve_selected_row",
		"failure_result_behavior": "show_same_shell_error_preserve_selection",
	}
}

func validViewField(key string) map[string]any {
	return map[string]any{
		"field_key":                    key,
		"label":                        "Name",
		"default_hidden":               false,
		"sortable":                     true,
		"header_sort_field_key":        nil,
		"filter_ops":                   []any{"eq"},
		"groupable":                    true,
		"read_kind":                    "scalar",
		"write_kind":                   "direct_value",
		"grid_editable":                true,
		"conflict_resolution_class":    nil,
		"entity_binding_mode":          nil,
		"string_contract_id":           nil,
		"direct_scalar_contract_id":    nil,
		"direct_reference_contract_id": nil,
		"clearable":                    true,
		"writable":                     true,
		"read_model":                   "record",
		"write_target":                 key,
	}
}

func validErrorRegistry() map[string]any {
	return map[string]any{
		"$schema":     contractDraft202012Schema,
		"registry_id": "cartulary.errors.v1",
		"note":        "test registry",
		"errors": []any{
			validErrorEntry("cartulary.test.conflict", "409"),
		},
	}
}

func validReasonRegistry(errorCode string, reasonCodes ...string) map[string]any {
	entries := make([]any, 0, len(reasonCodes))
	for _, reasonCode := range reasonCodes {
		entries = append(entries, validReasonCode(reasonCode))
	}
	return map[string]any{
		"error_code":   errorCode,
		"reason_codes": entries,
	}
}

func validReasonCode(code string) map[string]any {
	return map[string]any{
		"code":    code,
		"summary": "test reason",
	}
}

func validContractFamilyEntry(familyID, root, status, goName string, order int, activationDependencyIDs []any) map[string]any {
	if activationDependencyIDs == nil {
		activationDependencyIDs = []any{}
	}
	return map[string]any{
		"family_id":                 familyID,
		"contract_root":             root,
		"generation_status":         status,
		"go_name":                   goName,
		"output_order":              order,
		"generated_outputs":         []any{"internal/gen/contracts/" + familyID + "_gen.go"},
		"typescript_projections":    []any{},
		"activation_dependency_ids": activationDependencyIDs,
		"description":               "test family",
	}
}

func validErrorEntry(code string, status string) map[string]any {
	return map[string]any{
		"code":        code,
		"http_status": json.Number(status),
		"summary":     "test error",
	}
}

func writeJSONFile(t *testing.T, file string, value any) {
	t.Helper()
	if err := os.MkdirAll(filepath.Dir(file), 0o755); err != nil {
		t.Fatalf("create fixture dir: %v", err)
	}
	data, err := json.MarshalIndent(value, "", "  ")
	if err != nil {
		t.Fatalf("marshal fixture: %v", err)
	}
	if err := os.WriteFile(file, append(data, '\n'), 0o644); err != nil {
		t.Fatalf("write fixture: %v", err)
	}
}

func requireErrorContains(t *testing.T, err error, substring string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error containing %q", substring)
	}
	if !strings.Contains(err.Error(), substring) {
		t.Fatalf("expected error containing %q, got %q", substring, err.Error())
	}
}

func checkExtensionImportReferences(t *testing.T, root string) {
	for _, kind := range []string{"schema", "algorithm", "retired harness"} {
		t.Run(kind, func(t *testing.T) {
			directory := extensionInputFixture(t, root, func(_ string, object map[string]any) {
				if object["schema_id"] != "cartulary.extension_dependency_declaration_set.v3" {
					return
				}
				for _, raw := range object["dependencies"].([]any) {
					row := raw.(map[string]any)
					if row["dependency_id"] != "testing_harness" {
						continue
					}
					field, id := "imported_schema_ids", "cartulary.unimplemented.v1"
					if kind == "algorithm" {
						field, id = "imported_algorithm_ids", "unimplemented_v1"
					}
					if kind == "retired harness" {
						id = "cartulary.harness_row_result.v5"
					}
					row[field] = []any{id}
				}
			})
			if _, err := deriveExtensionArtifacts(directory); err == nil {
				t.Fatal("generated unresolved dependency import")
			}
		})
	}
}
