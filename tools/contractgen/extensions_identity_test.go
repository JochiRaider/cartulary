package main

import (
	"github.com/JochiRaider/cartulary/internal/modules/extensions/registryfacts"
	"testing"
)

func checkOwnerFactNormalization(t *testing.T) {
	fragment := map[string]any{
		"schema_id": "cartulary.extension_owner_fragment.v3", "owner_id": "core00", "owner_fragment_id": "core00.recognition",
		"facts": []any{
			map[string]any{"fact_kind": "recognized_profile", "profile_id": "example", "primary_owner_id": "example", "claimable": true, "contract_major": 1},
			map[string]any{"fact_kind": "recognized_profile", "profile_id": "example", "primary_owner_id": "example", "claimable": false, "contract_major": 1},
		},
	}
	other := cloneObject(fragment)
	facts := other["facts"].([]any)
	facts[0], facts[1] = facts[1], facts[0]
	if err := normalizeExtensionArtifact(fragment); err != nil {
		t.Fatal(err)
	}
	if err := normalizeExtensionArtifact(other); err != nil {
		t.Fatal(err)
	}
	left, err := extensionCanonicalDigest(fragment)
	if err != nil {
		t.Fatal(err)
	}
	right, err := extensionCanonicalDigest(other)
	if err != nil {
		t.Fatal(err)
	}
	if left != right {
		t.Fatal("normalized duplicate fact values depend on authored order")
	}
	if len(fragment["facts"].([]any)) != 2 || fragment["facts"].([]any)[0].(map[string]any)["claimable"] != false {
		t.Fatal("normalization lost a duplicate or ignored canonical fact tie-break")
	}
}

func checkOwnerIdentitySchemaGeneration(t *testing.T, root string) {
	inputs, err := readExtensionInputCatalog(root)
	if err != nil {
		t.Fatal(err)
	}
	for _, mode := range []string{"missing", "retired", "duplicate", "wrong owner"} {
		t.Run("definition "+mode, func(t *testing.T) {
			set := cloneObject(inputs["specification/contract-definitions.json"])
			rows := set["definitions"].([]any)
			for i, raw := range rows {
				row := raw.(map[string]any)
				if row["schema_id"] != registryfacts.IdentitySchemaID {
					continue
				}
				switch mode {
				case "missing":
					set["definitions"] = append(rows[:i], rows[i+1:]...)
				case "retired":
					row["schema_id"] = "cartulary.extension_owner_fact_identity.v1"
				case "duplicate":
					set["definitions"] = append(rows, row)
				case "wrong owner":
					row["owner_id"] = "core00"
				}
				break
			}
			if err := registryfacts.ValidateIdentityDefinition(set); err == nil {
				t.Fatal("invalid current definition accepted")
			}
		})
	}
	for _, mode := range []string{"missing", "retired", "future"} {
		t.Run(mode, func(t *testing.T) {
			source := cloneObject(inputs["specification/generated-schema-sources.json"])
			rows := source["schemas"].([]any)
			for i, raw := range rows {
				schema := raw.(map[string]any)
				if schema["$id"] != "cartulary.extension_owner_fact_identity.v2" {
					continue
				}
				switch mode {
				case "missing":
					source["schemas"] = append(rows[:i], rows[i+1:]...)
				case "retired":
					schema["$id"] = "cartulary.extension_owner_fact_identity.v1"
				case "future":
					schema["$id"] = "cartulary.extension_owner_fact_identity.v99"
				}
				break
			}
			if _, err := materializeExtensionSchemas(source); err == nil {
				t.Fatal("generator accepted incompatible identity schema")
			}
		})
	}
}
