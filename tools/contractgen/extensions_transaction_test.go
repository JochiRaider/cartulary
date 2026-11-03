package main

import (
	"encoding/json"
	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"
	"testing"
)

func checkTransactionLimitProjection(t *testing.T, root string) {
	t.Helper()
	indexed, err := readExtensionInputCatalog(root)
	if err != nil {
		t.Fatal(err)
	}
	limits, err := extensionTransactionProtocol(indexed)
	if err != nil {
		t.Fatal(err)
	}
	for _, field := range []string{"canonical_encoding_id", "input_byte_admission", "prepare_result_schema_id", "validation_result_schema_id", "write_result_schema_id"} {
		current := limits[field]
		limits[field] = "unsupported"
		if _, err := extensionTransactionProtocol(indexed); err == nil {
			t.Fatalf("unsupported %s admitted", field)
		}
		limits[field] = current
	}
	protocol := indexed["specification/shared-protocols.json"]
	currentSchema := protocol["schema_id"]
	protocol["schema_id"] = "cartulary.extension_shared_protocol_contract_set.v4"
	if _, err := extensionTransactionProtocol(indexed); err == nil {
		t.Fatal("retired protocol admitted")
	}
	protocol["schema_id"] = currentSchema
	// A deliberately distinct fixture value proves dependency on the source,
	// rather than two projections coincidentally sharing a copied literal.
	limits["per_participant_input_max_bytes"] = json.Number("16777217")
	bindings := extensionBindingTransactionLimits(limits)
	if bindings["participant_input_bytes"] != json.Number("16777217") {
		t.Fatal("binding ignored protocol input")
	}
	schemas := cloneObject(indexed["specification/generated-schema-sources.json"])
	if err := projectExtensionTransactionProtocol(schemas, limits); err != nil {
		t.Fatal(err)
	}
	rows, err := materializeExtensionSchemas(schemas)
	if err != nil {
		t.Fatal(err)
	}
	checked := 0
	for _, row := range rows {
		props, _ := row["properties"].(map[string]any)
		switch row["$id"] {
		case "cartulary.extension_implementation_binding.v2":
			limit := props["transaction_participant_limits"].(map[string]any)["properties"].(map[string]any)["participant_input_bytes"].(map[string]any)
			if limit["const"] != json.Number("16777217") || limit["x-cartulary-transaction-value"] != nil {
				t.Fatal("binding schema ignored protocol input")
			}
			checked++
		case "cartulary.extension_shared_protocol_contract_set.v5":
			fields := props["transaction"].(map[string]any)["properties"].(map[string]any)
			for _, field := range []string{"canonical_encoding_id", "input_byte_admission", "prepare_result_schema_id", "validation_result_schema_id", "write_result_schema_id"} {
				if fields[field].(map[string]any)["const"] != limits[field] {
					t.Fatalf("unbound %s", field)
				}
			}

			limit := props["transaction"].(map[string]any)["properties"].(map[string]any)["per_participant_input_max_bytes"].(map[string]any)
			if limit["const"] != json.Number("16777217") {
				t.Fatal("protocol schema ignored source input")
			}
			checked++
		}
	}
	if checked != 2 {
		t.Fatal("missing projected schema")
	}
	if err := projectExtensionTransactionProtocol(map[string]any{"x-cartulary-transaction-value": "unknown"}, limits); err == nil {
		t.Fatal("unknown projection admitted")
	}
	delete(limits, "result_max_bytes")
	if _, err := extensionTransactionProtocol(indexed); err == nil {
		t.Fatal("incomplete protocol admitted")
	}
}

func checkTransactionResultSchemas(t *testing.T, compiler *jsonschema.Compiler) {
	for _, kind := range []string{"prepare", "validation", "write"} {
		id := "cartulary.extension_transaction_participant_" + kind + "_result.v2"
		schema, err := compiler.Compile("https://extensions.invalid/schemas/" + id)
		if err != nil {
			t.Fatal(err)
		}
		value := map[string]any{"schema_id": id, "participant_id": "network_flow_activity.indicator_link_v1"}
		switch kind {
		case "prepare":
			value["serialization_keys"] = []any{}
		case "validation":
			value["status"] = "valid"
			value["findings"] = []any{}
		case "write":
			value["status"] = "written"
		}
		if err := schema.Validate(value); err != nil {
			t.Fatalf("current %s: %v", kind, err)
		}
		value["receipt"] = "private"
		if schema.Validate(value) == nil {
			t.Fatalf("%s admitted an owner receipt", kind)
		}
		delete(value, "receipt")
		value["schema_id"] = "cartulary.extension_transaction_participant_" + kind + "_result.v1"
		if schema.Validate(value) == nil {
			t.Fatalf("retired %s result admitted", kind)
		}
		if _, err := compiler.Compile("https://extensions.invalid/schemas/" + value["schema_id"].(string)); err == nil {
			t.Fatalf("retired %s schema packaged", kind)
		}
		value["schema_id"] = id
		if kind != "write" {
			field := "serialization_keys"
			if kind == "validation" {
				field = "findings"
			}
			value[field] = nil
			if schema.Validate(value) == nil {
				t.Fatalf("%s admitted null array", kind)
			}
		}
	}
}
