package extensions

import (
	"bytes"
	"encoding/json"
	"strings"
	"testing"

	"github.com/JochiRaider/cartulary/internal/modules/extensions/registryfacts"
	"github.com/JochiRaider/cartulary/internal/platform/canonicaljson"
	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"
)

func checkOwnerIdentityContract(t *testing.T) {
	// These byte and digest literals are independent of the production encoder.
	vectors := []struct{ kind, parent, key, value, bytes, digest string }{
		{"recognized_profile", "", "", "", "{\"fact_kind\":\"recognized_profile\",\"identity\":{},\"profile_id\":\"example\",\"schema_id\":\"cartulary.extension_owner_fact_identity.v2\"}\n", "23e89be9f509f4e09b7c47c2fcc847141483563e18e416ad9745f01254c67b6e"},
		{"claim_configuration", "", "", "", "{\"fact_kind\":\"claim_configuration\",\"identity\":{},\"profile_id\":\"example\",\"schema_id\":\"cartulary.extension_owner_fact_identity.v2\"}\n", "a3721d51db4631476a3eaeaf3342a8a3fc701ba29c55d3a284fd98003110e7f5"},
		{"state_ownership", "", "", "", "{\"fact_kind\":\"state_ownership\",\"identity\":{},\"profile_id\":\"example\",\"schema_id\":\"cartulary.extension_owner_fact_identity.v2\"}\n", "dd7b6ade10e27414c4b9bcdca2954c6979fe515702b59644e21ab83ded283c27"},
		{"admission_validation", "", "", "", "{\"fact_kind\":\"admission_validation\",\"identity\":{},\"profile_id\":\"example\",\"schema_id\":\"cartulary.extension_owner_fact_identity.v2\"}\n", "176289ec597003f19c5382efab6512153904c542b180fbae95854a3b33e4ab09"},
		{"egress", "", "", "", "{\"fact_kind\":\"egress\",\"identity\":{},\"profile_id\":\"example\",\"schema_id\":\"cartulary.extension_owner_fact_identity.v2\"}\n", "5567d0a9a924265f5ae2bbd1995383f1cf4d39eee011d570f77189c9527a8e3d"},
		{"portability", "", "", "", "{\"fact_kind\":\"portability\",\"identity\":{},\"profile_id\":\"example\",\"schema_id\":\"cartulary.extension_owner_fact_identity.v2\"}\n", "f14ea975c74c0451e70b5e0b87ac7dc0d6a01492f8a3df2d9f57771fea03ab89"},
		{"snapshot_reporting", "", "", "", "{\"fact_kind\":\"snapshot_reporting\",\"identity\":{},\"profile_id\":\"example\",\"schema_id\":\"cartulary.extension_owner_fact_identity.v2\"}\n", "7afadd6be0c698b1f7ef19a12a47f306a7a1f4a31908e059898376376d3a07b0"},
		{"runtime_dependency", "", "dependency_profile_id", "dependency", "{\"fact_kind\":\"runtime_dependency\",\"identity\":{\"dependency_profile_id\":\"dependency\"},\"profile_id\":\"example\",\"schema_id\":\"cartulary.extension_owner_fact_identity.v2\"}\n", "d3b64aa963caa89448982db99ecbda312bb72a46e5c1b07a626c602de031faf8"},
		{"route_family", "", "route_family", "/api/v1/example/{id}", "{\"fact_kind\":\"route_family\",\"identity\":{\"route_family\":\"/api/v1/example/{id}\"},\"profile_id\":\"example\",\"schema_id\":\"cartulary.extension_owner_fact_identity.v2\"}\n", "0c7c23d6e69cefb9f69ed8f18680b6f18179d15d4e916ce6bf023b1ef5afba60"},
		{"workspace", "", "workspace_key", "example_workspace", "{\"fact_kind\":\"workspace\",\"identity\":{\"workspace_key\":\"example_workspace\"},\"profile_id\":\"example\",\"schema_id\":\"cartulary.extension_owner_fact_identity.v2\"}\n", "2d0053e770581a27f79f0dfde47b4577e222e96999d532fd7b7da88a96a81757"},
		{"public_schema", "", "public_schema_id", "cartulary.example.v1", "{\"fact_kind\":\"public_schema\",\"identity\":{\"public_schema_id\":\"cartulary.example.v1\"},\"profile_id\":\"example\",\"schema_id\":\"cartulary.extension_owner_fact_identity.v2\"}\n", "a115cb8efeecbfb5b0c352683618e8840dfabd17852f4f6abbbea955c76b8a20"},
		{"contribution", "contribution", "contribution_id", "example.route", "{\"fact_kind\":\"contribution\",\"identity\":{\"contribution_id\":\"example.route\"},\"profile_id\":\"example\",\"schema_id\":\"cartulary.extension_owner_fact_identity.v2\"}\n", "cf567c72f9c3680ca8365c309875120556f0941ec3e994a267812871f332a85a"},
		{"migration_definition", "migration_definition", "migration_id", "example.migrate", "{\"fact_kind\":\"migration_definition\",\"identity\":{\"migration_id\":\"example.migrate\"},\"profile_id\":\"example\",\"schema_id\":\"cartulary.extension_owner_fact_identity.v2\"}\n", "fa6bc7b4a208b00b36594522d862cf8db7cd5ce1ac528e55a6a95f17468e6f7a"},
		{"migration_ledger_definition", "migration_ledger_definition", "migration_id", "example.migrate", "{\"fact_kind\":\"migration_ledger_definition\",\"identity\":{\"migration_id\":\"example.migrate\"},\"profile_id\":\"example\",\"schema_id\":\"cartulary.extension_owner_fact_identity.v2\"}\n", "d5e76e94ec83d21678eaa5a3f08b317e314a83bc28087da154effda6fcfaabc5"},
		{"worker_kind", "", "worker_kind", "example.worker", "{\"fact_kind\":\"worker_kind\",\"identity\":{\"worker_kind\":\"example.worker\"},\"profile_id\":\"example\",\"schema_id\":\"cartulary.extension_owner_fact_identity.v2\"}\n", "f283616a27b270a1566a552f6b816d9579c30193bb1b090cde0ae91b22799348"},
		{"job_kind", "job_kind_contract", "job_kind", "example.work", "{\"fact_kind\":\"job_kind\",\"identity\":{\"job_kind\":\"example.work\"},\"profile_id\":\"example\",\"schema_id\":\"cartulary.extension_owner_fact_identity.v2\"}\n", "bfdce97f95f8ab96adf2be04d707330a2ba3510a37d3d06a59ffab148d49a13d"},
	}
	source := newMutableGeneratedSource()
	schemaObject := decodeMutableObject(t, source[generatedExtensionsRoot+"schemas/"+registryfacts.IdentitySchemaID+".schema.json"].JSON)
	compiler := jsonschema.NewCompiler()
	compiler.UseLoader(jsonschema.SchemeURLLoader{})
	compiler.DefaultDraft(jsonschema.Draft2020)
	if err := compiler.AddResource("https://extensions.invalid/identity", schemaObject); err != nil {
		t.Fatal(err)
	}
	schema, err := compiler.Compile("https://extensions.invalid/identity")
	if err != nil {
		t.Fatal(err)
	}
	for _, vector := range vectors {
		t.Run(vector.kind, func(t *testing.T) {
			fact := map[string]any{"profile_id": "example", "fact_kind": vector.kind}
			if vector.key != "" {
				body := fact
				if vector.parent != "" {
					body = map[string]any{}
					fact[vector.parent] = body
				}
				body[vector.key] = vector.value
			}
			identity, err := registryfacts.Identify(fact)
			if err != nil || string(identity.CanonicalBytes()) != vector.bytes || identity.Token() != vector.kind+":sha256:"+vector.digest {
				t.Fatalf("identity vector mismatch: %v, %q, %s", err, identity.CanonicalBytes(), identity.Token())
			}
			value, err := canonicaljson.DecodeExtension(identity.CanonicalBytes())
			if err != nil || schema.Validate(value) != nil {
				t.Fatal("identity does not satisfy its packaged schema")
			}
			copy := identity.CanonicalBytes()
			copy[0] = '!'
			if string(identity.CanonicalBytes()) != vector.bytes {
				t.Fatal("identity bytes escaped ownership")
			}
			for _, mutate := range []func(map[string]any){
				func(v map[string]any) { v["schema_id"] = "cartulary.extension_owner_fact_identity.v1" },
				func(v map[string]any) { delete(v, "identity") },
				func(v map[string]any) { v["identity"].(map[string]any)["extra"] = true },
			} {
				invalid := decodeMutableObject(t, identity.CanonicalBytes())
				mutate(invalid)
				if schema.Validate(invalid) == nil {
					t.Fatal("identity schema admitted an incompatible object")
				}
			}
		})
	}
	for _, fact := range []map[string]any{
		{"profile_id": "example", "fact_kind": "unsupported"},
		{"profile_id": "example", "fact_kind": "capability"},
		{"profile_id": "example", "fact_kind": "workspace", "workspace_key": nil},
		{"profile_id": "example", "fact_kind": "job_kind", "job_kind_contract": true},
		{"profile_id": "Example", "fact_kind": "recognized_profile"},
		{"profile_id": "example", "fact_kind": "public_schema", "public_schema_id": strings.Repeat("x", 161)},
		{"profile_id": "example", "fact_kind": "contribution", "contribution": map[string]any{"contribution_id": "invalid"}},
		{"profile_id": "example", "fact_kind": "route_family", "route_family": "/api/v1/\xff"},
	} {
		if _, err := registryfacts.Identify(fact); err == nil {
			t.Fatal("malformed identity was accepted")
		}
	}
	for _, route := range []string{"/api/v1/é", "/api/v1/a b", "/api/v1/a?x", "/api/v1/a%20b", "/api/v1/a/", "/api/v1/a//b", "/api/v1/./b", "/api/v1/../b", "/api/v1/{Bad}", "/api/v1/{id", "/api/v1/{}", "/api/v1/" + strings.Repeat("x/", 30) + "x", "/api/v1/" + strings.Repeat("x", 505)} {
		if _, err := registryfacts.Identify(map[string]any{"profile_id": "example", "fact_kind": "route_family", "route_family": route}); err == nil {
			t.Fatalf("invalid route identity accepted: %q", route)
		}
		invalid := map[string]any{"schema_id": registryfacts.IdentitySchemaID, "profile_id": "example", "fact_kind": "route_family", "identity": map[string]any{"route_family": route}}
		if schema.Validate(invalid) == nil {
			t.Fatalf("identity schema admitted invalid route: %q", route)
		}
	}
	for _, route := range []string{"/api/v1/{id}", "/api/v1/a._~-1", "/api/v1/" + strings.Repeat("x", 504), "/api/v1/" + strings.Repeat("x/", 29) + "x"} {
		if _, err := registryfacts.Identify(map[string]any{"profile_id": "example", "fact_kind": "route_family", "route_family": route}); err != nil {
			t.Fatalf("valid route identity rejected: %q", route)
		}
	}
	t.Run("normalization isolation and duplicate preservation", checkOwnerInputNormalization)
	t.Run("current definition and canonical package admission", checkOwnerIdentityPackage)
}

func checkOwnerInputNormalization(t *testing.T) {
	fragment := func(owner, id string, flag bool) map[string]any {
		return map[string]any{"schema_id": "cartulary.extension_owner_fragment.v3", "owner_id": owner, "owner_fragment_id": owner + "." + id,
			"facts": []any{map[string]any{"fact_kind": "recognized_profile", "profile_id": "example", "primary_owner_id": "example", "claimable": flag, "contract_major": json.Number("1")}}}
	}
	input := []map[string]any{fragment("z", "facts", true), fragment("a", "facts", true), fragment("a", "facts", false), fragment("a", "facts", false)}
	input[0]["facts"] = append(input[0]["facts"].([]any), input[1]["facts"].([]any)[0])
	before, err := canonicaljson.MarshalExtension(input)
	if err != nil {
		t.Fatal(err)
	}
	got, err := registryfacts.NormalizeFragments(input)
	if err != nil {
		t.Fatal(err)
	}
	after, _ := canonicaljson.MarshalExtension(input)
	if !bytes.Equal(before, after) || len(got) != 4 || len(got[3]["facts"].([]any)) != 2 {
		t.Fatal("normalizer changed input or lost duplicates")
	}
	normalized, _ := canonicaljson.MarshalExtension(got)
	input[0], input[3] = input[3], input[0]
	input[1], input[2] = input[2], input[1]
	reversed, err := registryfacts.NormalizeFragments(input)
	if err != nil {
		t.Fatal(err)
	}
	reversedBytes, _ := canonicaljson.MarshalExtension(reversed)
	if !bytes.Equal(normalized, reversedBytes) {
		t.Fatal("fragment order affects normalization")
	}
	got[0]["facts"].([]any)[0].(map[string]any)["profile_id"] = "changed"
	unchanged, _ := canonicaljson.MarshalExtension(input)
	for _, value := range input {
		if value["facts"].([]any)[0].(map[string]any)["profile_id"] != "example" {
			t.Fatal("returned nested fact aliases input")
		}
	}
	if bytes.Contains(unchanged, []byte("changed")) {
		t.Fatal("returned objects mutate callers")
	}
	empty, err := registryfacts.NormalizeFragments(nil)
	encoded, _ := canonicaljson.MarshalExtension(empty)
	if err != nil || string(encoded) != "[]\n" {
		t.Fatal("empty registry is not canonical")
	}
	bad := fragment("a", "facts", true)
	bad["facts"].([]any)[0].(map[string]any)["unsafe"] = "\xff"
	if _, err := registryfacts.NormalizeFragments([]map[string]any{bad}); err == nil {
		t.Fatal("normalization repaired invalid Unicode")
	}
}

func checkOwnerIdentityPackage(t *testing.T) {
	t.Run("current definition and normalization parity", func(t *testing.T) {
		source := newMutableGeneratedSource()
		input := decodeMutableObject(t, source[generatedExtensionsRoot+"owner-input-registry.json"].JSON)
		fragments, ok := objectSlice(input["owner_fragments"])
		if !ok {
			t.Fatal("invalid fragment array")
		}
		normalized, err := registryfacts.NormalizeFragments(fragments)
		if err != nil {
			t.Fatal(err)
		}
		if !equalCanonicalObjects(map[string]any{"rows": fragments}, map[string]any{"rows": normalized}) {
			t.Fatal("packaged normalized order diverges")
		}
		if _, err := NewCoordinator(source); err != nil {
			t.Fatal("current packaged identity admission: ", err)
		}
	})
	for _, mode := range []string{"missing schema", "retired schema", "future schema", "fact order", "fragment order"} {
		t.Run(mode, func(t *testing.T) {
			source := newMutableGeneratedSource()
			path := generatedExtensionsRoot + "schemas/" + registryfacts.IdentitySchemaID + ".schema.json"
			alternate := ""
			if mode == "retired schema" {
				alternate = "cartulary.extension_owner_fact_identity.v1"
			}
			if mode == "future schema" {
				alternate = "cartulary.extension_owner_fact_identity.v99"
			}
			if alternate != "" {
				source[generatedExtensionsRoot+"schemas/"+alternate+".schema.json"] = source[path]
				delete(source, path)
			} else if mode == "missing schema" {
				delete(source, path)
			}
			rewriteCoherentPackage(t, source, func(object map[string]any) {
				if _, present := object["generated_schema_digests"]; present && object["schema_id"] == "cartulary.extension_registry_integrity.v2" {
					rows := object["generated_schema_digests"].([]any)
					for i, raw := range rows {
						row := raw.(map[string]any)
						if row["schema_id"] != registryfacts.IdentitySchemaID {
							continue
						}
						if alternate != "" {
							row["schema_id"] = alternate
						}
						if mode == "missing schema" {
							object["generated_schema_digests"] = append(rows[:i], rows[i+1:]...)
						}
						break
					}
				}
				if object["$id"] == registryfacts.IdentitySchemaID && alternate != "" {
					object["$id"] = alternate
					object["properties"].(map[string]any)["schema_id"].(map[string]any)["const"] = alternate
				}
				if _, present := object["facts"]; present && mode == "fact order" && object["schema_id"] == "cartulary.extension_owner_fragment.v3" {
					rows := object["facts"].([]any)
					if len(rows) > 1 {
						rows[0], rows[1] = rows[1], rows[0]
					}
				}
				if _, present := object["owner_fragments"]; present && mode == "fragment order" && object["schema_id"] == "cartulary.extension_owner_input_registry.v2" {
					rows := object["owner_fragments"].([]any)
					rows[0], rows[1] = rows[1], rows[0]
				}
			})
			if _, err := NewCoordinator(source); err == nil {
				t.Fatal("coherent incompatible identity package was accepted")
			}
		})
	}
}
