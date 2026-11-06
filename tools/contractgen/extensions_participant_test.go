package main

import (
	"encoding/json"
	jsonschema "github.com/santhosh-tekuri/jsonschema/v6"
	"strings"
	"testing"
)

// Structural vectors use the actual emitted schemas. Invocation identity and
// owner-mapped finding semantics belong to the runtime result classifier.
func checkBindingRestoreResultSchema(t *testing.T, artifacts []artifact) {
	compiler := jsonschema.NewCompiler()
	compiler.UseLoader(jsonschema.SchemeURLLoader{})
	compiler.DefaultDraft(jsonschema.Draft2020)
	compiler.AssertFormat()
	for _, artifact := range artifacts {
		if !strings.HasPrefix(artifact.Path, generatedExtensionPrefix+"schemas/") {
			continue
		}
		object, err := decodeContract([]byte(artifact.JSON))
		if err != nil {
			t.Fatal(err)
		}
		id := stringValue(object.(map[string]any)["$id"])
		if err := compiler.AddResource("https://extensions.invalid/schemas/"+id, object); err != nil {
			t.Fatal(err)
		}
	}
	t.Run("Portability specialization schema", func(t *testing.T) { checkPortabilitySpecializationSchema(t, compiler) })
	t.Run("typed Reporting source schema", func(t *testing.T) { checkReportingSourceSchema(t, compiler) })
	t.Run("current transaction result schemas", func(t *testing.T) { checkTransactionResultSchemas(t, compiler) })
	t.Run("durable metadata projections", func(t *testing.T) { checkDurableMetadataSchemas(t, compiler) })
	const id = "cartulary.extension_binding_restore_validation_result.v1"
	schema, err := compiler.Compile("https://extensions.invalid/schemas/" + id)
	if err != nil {
		t.Fatal(err)
	}
	t.Run("retired backup specialization", func(t *testing.T) {
		operations := []any{}
		for _, kind := range []string{"backup_enumerate", "restore_rebuild", "restore_validate"} {
			operations = append(operations, map[string]any{
				"operation_kind": kind, "result_schema_id": "cartulary.extension_backup_restore_participant_result.v1",
				"algorithm_id": "network_flow_activity." + kind + "_v1", "output_schema_id": "cartulary.retired_backup_output.v1",
				"ordering_algorithm_id": "network_flow_activity.retired_order_v1", "state_family_ids": []any{},
				"max_input_bytes": 67108864, "max_output_bytes": 0, "max_items": 0,
			})
		}
		retired := map[string]any{
			"schema_id": "cartulary.extension_participant_specialization.v5", "profile_id": "network_flow_activity",
			"participant_id": "network_flow_activity.backup_restore_v1", "participant_kind": "backup_restore",
			"shared_context_schema_id": "cartulary.extension_backup_restore_participant_context.v1", "operations": operations,
		}
		specialization, err := compiler.Compile("https://extensions.invalid/schemas/cartulary.extension_participant_specialization.v5")
		if err != nil {
			t.Fatal(err)
		}
		if err := specialization.Validate(retired); err == nil {
			t.Fatal("packaged specialization schema admits retired backup kind")
		}
		requireErrorContains(t, validateExtensionArtifactShape(retired, "retired.json"), "invalid participant kind")
	})
	for _, tc := range []struct {
		name, status string
		count        int
		valid        bool
	}{{"valid empty", "valid", 0, true}, {"invalid one", "invalid", 1, true}, {"invalid ceiling", "invalid", 256, true}, {"invalid overflow", "invalid", 257, false}, {"invalid empty", "invalid", 0, false}, {"valid findings", "valid", 1, false}, {"unknown status", "error", 0, false}} {
		t.Run(tc.name, func(t *testing.T) {
			findings := make([]any, tc.count)
			for i := range findings {
				findings[i] = map[string]any{"path": "binding", "reason_code": "fixture", "message": "fixture", "details": map[string]any{}}
			}
			value := map[string]any{"schema_id": id, "binding_id": "network_flow_activity.tables", "status": tc.status, "findings": findings}
			if err := schema.Validate(value); (err == nil) != tc.valid {
				t.Fatalf("structural result admission = %v", err)
			}
		})
	}
	for _, name := range []string{"unknown member", "missing identity", "invalid identity", "null findings", "unknown finding member"} {
		t.Run(name, func(t *testing.T) {
			value := map[string]any{"schema_id": id, "binding_id": "network_flow_activity.tables", "status": "valid", "findings": []any{}}
			switch name {
			case "unknown member":
				value["profile_id"] = "network_flow_activity"
			case "missing identity":
				delete(value, "binding_id")
			case "invalid identity":
				value["binding_id"] = "../physical"
			case "null findings":
				value["findings"] = nil
			case "unknown finding member":
				value["status"] = "invalid"
				value["findings"] = []any{map[string]any{"path": "binding", "reason_code": "fixture", "message": "fixture", "details": map[string]any{}, "extra": true}}
			}
			if err := schema.Validate(value); err == nil {
				t.Fatal("malformed structural result admitted")
			}
		})
	}
}

func checkDurableMetadataSchemas(t *testing.T, compiler *jsonschema.Compiler) {
	metadata := map[string]any{
		"schema_id": "cartulary.extension_state_metadata.v1", "profile_id": "test",
		"migration_lineage_id": "test.state_v1", "state_version": 1, "metadata_version": 1, "last_migration_id": nil,
	}
	cancellation := map[string]any{
		"schema_id": "cartulary.extension_job_cancellation_observation.v1", "job_id": "00000000-0000-0000-0000-000000000001",
		"cancellation_request_id": "00000000-0000-0000-0000-000000000002", "observed_at": "2026-01-01T00:00:00Z", "observed_before_final_commit": true,
	}
	staged := map[string]any{
		"schema_id": "cartulary.extension_staged_object.v1", "staging_id": "stage:1", "owner_profile_id": "test", "storage_identity": "opaque-1",
		"expected_byte_size": 0, "expected_sha256": strings.Repeat("a", 64), "staged_at": "2026-01-01T00:00:00Z", "staging_expires_at": "2026-01-02T00:00:00Z",
		"ready_at": nil, "published_at": nil, "abandoned_at": nil, "state": "allocated", "delete_state": "not_applicable", "delete_attempt_count": 0,
		"next_delete_attempt_at": nil, "last_delete_error_code": nil,
	}
	for _, vector := range []struct {
		name    string
		value   map[string]any
		field   string
		invalid any
	}{
		{"zero metadata version", metadata, "metadata_version", 0},
		{"oversized metadata version", metadata, "metadata_version", int64(2147483648)},
		{"zero state version", metadata, "state_version", 0},
		{"malformed lineage", metadata, "migration_lineage_id", "unknown"},
		{"postcommit cancellation observation", cancellation, "observed_before_final_commit", false},
		{"unknown staged state", staged, "state", "complete"},
		{"negative object size", staged, "expected_byte_size", -1},
		{"oversized delete attempts", staged, "delete_attempt_count", int64(2147483648)},
		{"premature ready time", staged, "ready_at", "2026-01-01T00:00:00Z"},
		{"delete live object", staged, "delete_state", "pending"},
		{"unexpected deletion error", staged, "last_delete_error_code", "delete_failed"},
	} {
		t.Run(vector.name, func(t *testing.T) {
			schema, err := compiler.Compile("https://extensions.invalid/schemas/" + stringValue(vector.value["schema_id"]))
			if err != nil {
				t.Fatal(err)
			}
			if err := schema.Validate(vector.value); err != nil {
				t.Fatalf("valid durable metadata rejected: %v", err)
			}
			invalid := cloneObject(vector.value)
			invalid[vector.field] = vector.invalid
			if schema.Validate(invalid) == nil {
				t.Fatal("invalid durable metadata accepted")
			}
		})
	}
	definition := map[string]any{
		"schema_id": "cartulary.extension_migration_ledger_definition.v1", "profile_id": "test", "migration_lineage_id": "test.state_v1",
		"migration_id": "test.step_v1", "from_state_version": 1, "to_state_version": 2, "migration_definition_sha256": strings.Repeat("a", 64),
	}
	// An owner-authored ledger fact and the emitted schema must describe the same
	// seven-field non-executable value, independently of today's empty NF ledger.
	for _, kind := range []string{"current", "missing owner", "wrong owner", "executable field"} {
		t.Run("ledger "+kind, func(t *testing.T) {
			value := cloneObject(definition)
			switch kind {
			case "missing owner":
				delete(value, "profile_id")
			case "wrong owner":
				value["profile_id"] = "other"
			case "executable field":
				value["apply_algorithm_id"] = "test.apply_v1"
			}
			fragment := map[string]any{"schema_id": "cartulary.extension_owner_fragment.v3", "owner_fragment_id": "test.ledger_v1", "owner_id": "test", "facts": []any{
				map[string]any{"fact_kind": "migration_ledger_definition", "profile_id": "test", "migration_ledger_definition": value},
			}}
			err := validateExtensionOwnerFragment(fragment, "test.json")
			if (err == nil) != (kind == "current") {
				t.Fatalf("ledger fact validation: %v", err)
			}
			schema, err := compiler.Compile("https://extensions.invalid/schemas/cartulary.extension_migration_ledger_definition.v1")
			if err != nil {
				t.Fatal(err)
			}
			err = schema.Validate(value)
			if (err == nil) != (kind == "current" || kind == "wrong owner") {
				t.Fatalf("ledger structural schema: %v", err)
			}
		})
	}
}

func checkReportingSourceSchema(t *testing.T, compiler *jsonschema.Compiler) {
	schema, err := compiler.Compile("https://extensions.invalid/schemas/cartulary.extension_reporting_source_contract.v1")
	if err != nil {
		t.Fatal(err)
	}
	current := map[string]any{"schema_id": "cartulary.extension_reporting_source_contract.v1", "profile_id": "network_flow_activity", "participant_id": "network_flow_activity.reporting_graph_source_v1", "participant_kind": "snapshot_reporting", "projection_schema_id": "graph_projection.v2", "label_schema_id": "cartulary.network_flow.reporting_graph_source.v3", "read_algorithm_id": "network_flow_activity.reporting_graph_source_v1", "ordering_algorithm_id": "network_flow_activity.reporting_graph_order_v1", "state_family_ids": []any{"network_flow_activity.graph_views"}, "max_vertices": 100000, "max_edges": 250000, "max_label_bytes": 67108864}
	if err := schema.Validate(current); err != nil {
		t.Fatal(err)
	}
	if err := validateExtensionReportingSource(current, "source.json"); err != nil {
		t.Fatal(err)
	}
	for _, key := range []string{"max_vertices", "max_edges", "max_label_bytes", "projection_schema_id", "label_schema_id", "read_algorithm_id", "ordering_algorithm_id", "state_family_ids"} {
		t.Run("missing "+key, func(t *testing.T) {
			v := cloneObject(current)
			delete(v, key)
			if schema.Validate(v) == nil {
				t.Fatal("missing required source field accepted")
			}
		})
	}
	for key, value := range map[string]any{"max_vertices": 100001, "max_edges": 250001, "max_label_bytes": 67108865, "operations": []any{}, "shared_context_schema_id": "retired", "participant_kind": "incident_portability", "state_family_ids": []any{"duplicate", "duplicate"}} {
		t.Run("invalid "+key, func(t *testing.T) {
			v := cloneObject(current)
			v[key] = value
			if schema.Validate(v) == nil {
				t.Fatal("invalid source field accepted")
			}
		})
	}
}

func portabilitySpecializationFixture() map[string]any {
	operations := []any{}
	for _, pair := range [][2]string{{"export", "cartulary.extension_portability_export_result.v2"}, {"import", "cartulary.extension_portability_import_preparation_result.v2"}} {
		operations = append(operations, map[string]any{
			"operation_kind": pair[0], "result_schema_id": pair[1], "algorithm_id": "fixture." + pair[0] + "_v1",
			"output_schema_id": "cartulary.fixture_output.v1", "ordering_algorithm_id": "fixture.order_v1", "state_family_ids": []any{},
			"max_input_bytes": json.Number("67108864"), "max_output_bytes": json.Number("67108864"), "max_items": json.Number("1048576"),
		})
	}
	return map[string]any{"schema_id": "cartulary.extension_participant_specialization.v5", "profile_id": "fixture", "participant_id": "fixture.portability_v1", "participant_kind": "incident_portability", "shared_context_schema_id": "cartulary.extension_portability_participant_context.v2", "operations": operations}
}

func checkPortabilitySpecializationSchema(t *testing.T, compiler *jsonschema.Compiler) {
	schema, err := compiler.Compile("https://extensions.invalid/schemas/cartulary.extension_participant_specialization.v5")
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{"current", "unknown contract member", "unknown operation member", "null output budget", "oversized input budget", "unsupported operation", "retired emitter", "retired schema"} {
		t.Run(name, func(t *testing.T) {
			v := portabilitySpecializationFixture()
			operation := v["operations"].([]any)[0].(map[string]any)
			switch name {
			case "unknown contract member":
				v["extra"] = true
			case "unknown operation member":
				operation["extra"] = true
			case "null output budget":
				operation["max_output_bytes"] = nil
			case "oversized input budget":
				operation["max_input_bytes"] = json.Number("67108865")
			case "unsupported operation":
				operation["operation_kind"] = "unsupported"
			case "retired emitter":
				v["participant_kind"] = "snapshot_reporting"
			case "retired schema":
				v["schema_id"] = "cartulary.extension_participant_specialization.v3"
			}
			if err := schema.Validate(v); (err == nil) != (name == "current") {
				t.Fatal("packaged specialization classification", err)
			}
			if err := validateExtensionArtifactShape(v, "fixture.json"); (err == nil) != (name == "current") {
				t.Fatal("authored specialization classification", err)
			}
		})
	}
}
